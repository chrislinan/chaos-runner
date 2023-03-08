package scrub

import (
	"context"

	"github.com/litmuschaos/chaos-runner/pkg/config"
	"github.com/litmuschaos/chaos-runner/types"
	"github.com/litmuschaos/chaos-runner/pkg/internal/cache"
	"github.com/litmuschaos/chaos-runner/pkg/internal/issues"
	"github.com/litmuschaos/chaos-runner/pkg/internal/sanitize"
)

// PodSecurityPolicy represents a PodSecurityPolicy scruber.
type PodSecurityPolicy struct {
	*issues.Collector
	*cache.PodSecurityPolicy
	*config.Config

	client types.Connection
}

// NewPodSecurityPolicy return a new PodSecurityPolicy scruber.
func NewPodSecurityPolicy(ctx context.Context, c *Cache, codes *issues.Codes) Sanitizer {
	p := PodSecurityPolicy{
		client:    c.factory.Client(),
		Config:    c.config,
		Collector: issues.NewCollector(codes, c.config),
	}

	var err error
	p.PodSecurityPolicy, err = c.podsecuritypolicies()
	if err != nil {
		p.AddErr(ctx, err)
	}

	return &p
}

// Sanitize all available PodSecurityPolicys.
func (p *PodSecurityPolicy) Sanitize(ctx context.Context) error {
	return sanitize.NewPodSecurityPolicy(p.Collector, p).Sanitize(ctx)
}
