package scrub

import (
	"context"

	"github.com/litmuschaos/chaos-runner/pkg/internal/cache"
	"github.com/litmuschaos/chaos-runner/pkg/internal/issues"
	"github.com/litmuschaos/chaos-runner/pkg/internal/sanitize"
)

// PodDisruptionBudget represents a pdb scruber.
type PodDisruptionBudget struct {
	*issues.Collector
	*cache.Pod
	*cache.PodDisruptionBudget
}

// NewPodDisruptionBudget return a new PodDisruptionBudget scruber.
func NewPodDisruptionBudget(ctx context.Context, c *Cache, codes *issues.Codes) Sanitizer {
	s := PodDisruptionBudget{Collector: issues.NewCollector(codes, c.config)}

	var err error
	s.PodDisruptionBudget, err = c.podDisruptionBudgets()
	if err != nil {
		s.AddErr(ctx, err)
	}

	s.Pod, err = c.pods()
	if err != nil {
		s.AddErr(ctx, err)
	}

	return &s
}

// Sanitize all available PodDisruptionBudgets.
func (c *PodDisruptionBudget) Sanitize(ctx context.Context) error {
	return sanitize.NewPodDisruptionBudget(c.Collector, c).Sanitize(ctx)
}
