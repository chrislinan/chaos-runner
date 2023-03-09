package main

import (
	"fmt"
	"github.com/litmuschaos/chaos-runner/pkg/config"
	"github.com/litmuschaos/chaos-runner/pkg/log"
	"github.com/litmuschaos/chaos-runner/pkg/popeye"
	"github.com/litmuschaos/chaos-runner/pkg/utils"
	"github.com/litmuschaos/chaos-runner/pkg/utils/analytics"
	zerolog "github.com/rs/zerolog/log"
	"github.com/sirupsen/logrus"
	"github.com/wI2L/jsondiff"
	"regexp"
	"strings"
)

type ResourceIssues struct {
	resource string
	issue    string
}

var resultMap = make(map[string][]ResourceIssues)

func init() {
	// Log as JSON instead of the default ASCII formatter.
	logrus.SetFormatter(&logrus.TextFormatter{
		FullTimestamp:          true,
		DisableSorting:         true,
		DisableLevelTruncation: true,
	})
}

func collectResult(gvr, issue, path string) {
	if _, ok := resultMap[gvr]; ok {
		issueList := resultMap[gvr]
		issueList = append(issueList, ResourceIssues{issue: issue, resource: path})
		resultMap[gvr] = issueList
	} else {
		issuelist := []ResourceIssues{{issue: issue, resource: path}}
		resultMap[gvr] = issuelist
	}
}

func main() {
	flags := config.NewFlags()
	flags.StandAlone = true
	level := "error"
	flags.LintLevel = &level

	engineDetails := utils.EngineDetails{}
	clients := utils.ClientSets{}
	// Getting kubeConfig and Generate ClientSets
	if err := clients.GenerateClientSetFromKubeConfig(); err != nil {
		log.Errorf("unable to create ClientSets, error: %v", err)
		return
	}
	// Fetching all the ENVs passed from the chaos-operator
	// create and initialize the experimentList
	if err := engineDetails.SetEngineDetails().SetEngineUID(clients); err != nil {
		log.Errorf("unable to get ChaosEngineUID, error: %v", err)
		return
	}
	experimentList := engineDetails.CreateExperimentList()
	log.InfoWithValues("Experiments details are as follows", logrus.Fields{
		"Experiments List":     engineDetails.Experiments,
		"Engine Name":          engineDetails.Name,
		"Targets":              engineDetails.Targets,
		"Service Account Name": engineDetails.SvcAccount,
		"Engine Namespace":     engineDetails.EngineNamespace,
	})

	if err := utils.InitialPatchEngine(engineDetails, clients, experimentList); err != nil {
		log.Errorf("unable to patch Initial ExperimentStatus in ChaosEngine, error: %v", err)
		return
	}

	// Steps for each Experiment
	for _, experiment := range experimentList {

		// Sending event to GA instance
		if engineDetails.ClientUUID != "" {
			analytics.TriggerAnalytics(experiment.Name, engineDetails.ClientUUID)
		}
		// check the existence of chaosexperiment inside the cluster
		if err := experiment.HandleChaosExperimentExistence(engineDetails, clients); err != nil {
			log.Errorf("unable to get ChaosExperiment name: %v, in namespace: %v, error: %v", experiment.Name, experiment.Namespace, err)
			experiment.ExperimentSkipped(utils.ExperimentNotFoundErrorReason, engineDetails, clients)
			continue
		}
		// derive the required field from the experiment & engine and set into experimentDetails struct
		if err := experiment.SetValueFromChaosResources(&engineDetails, clients); err != nil {
			log.Errorf("unable to set values from Chaos Resources, error: %v", err)
			experiment.ExperimentSkipped(utils.ExperimentNotFoundErrorReason, engineDetails, clients)
			engineDetails.ExperimentSkippedPatchEngine(&experiment, clients)
			continue
		}
		// derive the envs from the chaos experiment and override their values from chaosengine if any
		if err := experiment.SetENV(engineDetails, clients); err != nil {
			log.Errorf("unable to patch ENV, error: %v", err)
			experiment.ExperimentSkipped(utils.ExperimentEnvParseErrorReason, engineDetails, clients)
			engineDetails.ExperimentSkippedPatchEngine(&experiment, clients)
			continue
		}
		// derive the sidecar details from chaosengine
		if err := experiment.SetSideCarDetails(engineDetails.Name, clients); err != nil {
			log.Errorf("unable to get sidecar details, error: %v", err)
			experiment.ExperimentSkipped(utils.ExperimentSideCarPatchErrorReason, engineDetails, clients)
			engineDetails.ExperimentSkippedPatchEngine(&experiment, clients)
			continue
		}

		log.Infof("Scan hole cluster before running Chaos Experiment: %v", experiment.Name)

		p, err := popeye.NewPopeye(flags, &zerolog.Logger)
		if err != nil {
			log.Errorf("Popeye configuration load failed %v", err)
		}
		if e := p.Init(); e != nil {
			log.Errorf(e.Error())
		}
		_, _, before, err := p.Sanitize()
		if err != nil {
			log.Errorf(err.Error())
		}

		log.Infof("Preparing to run Chaos Experiment: %v", experiment.Name)

		if err := experiment.PatchResources(engineDetails, clients); err != nil {
			log.Errorf("unable to patch Chaos Resources required for Chaos Experiment: %v, error: %v", experiment.Name, err)
			experiment.ExperimentSkipped(utils.ExperimentDependencyCheckReason, engineDetails, clients)
			engineDetails.ExperimentSkippedPatchEngine(&experiment, clients)
			continue
		}
		// generating experiment dependency check event inside chaosengine
		experiment.ExperimentDependencyCheck(engineDetails, clients)

		// Creation of PodTemplateSpec, and Final Job
		if err := utils.BuildingAndLaunchJob(&experiment, clients); err != nil {
			log.Errorf("unable to construct chaos experiment job, error: %v", err)
			experiment.ExperimentSkipped(utils.ExperimentDependencyCheckReason, engineDetails, clients)
			engineDetails.ExperimentSkippedPatchEngine(&experiment, clients)
			continue
		}

		experiment.ExperimentJobCreate(engineDetails, clients)

		log.Infof("Started Chaos Experiment Name: %v, with Job Name: %v", experiment.Name, experiment.JobName)
		// Watching the chaos container till Completion
		if err := engineDetails.WatchChaosContainerForCompletion(&experiment, clients); err != nil {
			log.Errorf("unable to Watch the chaos container, error: %v", err)
			experiment.ExperimentSkipped(utils.ExperimentChaosContainerWatchErrorReason, engineDetails, clients)
			engineDetails.ExperimentSkippedPatchEngine(&experiment, clients)
			continue
		}

		log.Infof("Chaos Pod Completed, Experiment Name: %v, with Job Name: %v", experiment.Name, experiment.JobName)

		log.Infof("Scan hole cluster after running Chaos Experiment: %v", experiment.Name)

		_, _, after, err := p.Sanitize()
		if err != nil {
			log.Errorf(err.Error())
		}

		diffs, err := jsondiff.CompareJSON([]byte(before), []byte(after))
		if err != nil {
			// handle error
		}

		reg := regexp.MustCompile(`.*issues/`)
		if reg == nil { //解释失败，返回nil
			fmt.Println("regexp err")
			return
		}
		var gvr, issue string
		for _, diff := range diffs {
			path := reg.ReplaceAllString(diff.Path.String(), "")
			path = strings.Replace(path, "~1", "/", -1)
			if strings.Contains(path, "tally") || diff.Type != "add" {
				continue
			}
			switch diff.Value.(type) {
			case map[string]interface{}:
				if diff.Value.(map[string]interface{})["gvr"] == nil || diff.Value.(map[string]interface{})["message"] == nil {
					continue
				}
				gvr = diff.Value.(map[string]interface{})["gvr"].(string)
				issue = diff.Value.(map[string]interface{})["message"].(string)
			case []interface{}:
				if len(diff.Value.([]interface{})) == 0 {
					continue
				}
				for _, value := range diff.Value.([]interface{}) {
					if value.(map[string]interface{})["gvr"] == nil || value.(map[string]interface{})["message"] == nil {
						continue
					}
					gvr = value.(map[string]interface{})["gvr"].(string)
					issue = value.(map[string]interface{})["message"].(string)
				}
			}
			collectResult(gvr, issue, path)
		}

		for key, value := range resultMap {
			for _, v := range value {
				log.Infof("New issues detected after Chaos Experiment: %s, gvr: %s, resource: %s, issue: %s", experiment.Name, key, v.resource, v.issue)
			}
		}

		// Will Update the chaosEngine Status
		if err := engineDetails.UpdateEngineWithResult(&experiment, clients); err != nil {
			log.Errorf("unable to Update ChaosEngine Status, error: %v", err)
		}

		log.Infof("Chaos Engine has been updated with result, Experiment Name: %v", experiment.Name)

		// Delete/Retain the Job, based on the jobCleanUpPolicy
		jobCleanUpPolicy, err := engineDetails.DeleteJobAccordingToJobCleanUpPolicy(&experiment, clients)
		if err != nil {
			log.Errorf("unable to Delete ChaosExperiment Job, error: %v", err)
		}
		experiment.ExperimentJobCleanUp(string(jobCleanUpPolicy), engineDetails, clients)
	}
}
