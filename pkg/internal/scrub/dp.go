package scrub

import (
	"context"

	"github.com/litmuschaos/chaos-runner/pkg/config"
	"github.com/litmuschaos/chaos-runner/types"
	"github.com/litmuschaos/chaos-runner/pkg/internal/cache"
	"github.com/litmuschaos/chaos-runner/pkg/internal/issues"
	"github.com/litmuschaos/chaos-runner/pkg/internal/sanitize"
)

// Deployment represents a Deployment scruber.
type Deployment struct {
	*issues.Collector
	*cache.Deployment
	*cache.PodsMetrics
	*cache.Pod
	*cache.ServiceAccount
	*config.Config

	client types.Connection
}

// NewDeployment return a new Deployment scruber.
func NewDeployment(ctx context.Context, c *Cache, codes *issues.Codes) Sanitizer {
	d := Deployment{
		client:    c.factory.Client(),
		Config:    c.config,
		Collector: issues.NewCollector(codes, c.config),
	}

	var err error
	d.Deployment, err = c.deployments()
	if err != nil {
		d.AddErr(ctx, err)
	}

	d.PodsMetrics, _ = c.podsMx()

	d.Pod, err = c.pods()
	if err != nil {
		d.AddErr(ctx, err)
	}

	d.ServiceAccount, err = c.serviceaccounts()
	if err != nil {
		d.AddErr(ctx, err)
	}

	return &d
}

// Sanitize all available Deployments.
func (d *Deployment) Sanitize(ctx context.Context) error {
	return sanitize.NewDeployment(d.Collector, d).Sanitize(ctx)
}
