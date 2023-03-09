package scrub

import (
	"context"

	"github.com/litmuschaos/chaos-runner/pkg/config"
	"github.com/litmuschaos/chaos-runner/pkg/types"
	"github.com/litmuschaos/chaos-runner/pkg/internal/cache"
	"github.com/litmuschaos/chaos-runner/pkg/internal/issues"
	"github.com/litmuschaos/chaos-runner/pkg/internal/sanitize"
)

// Ingress represents a Ingress scruber.
type Ingress struct {
	*issues.Collector
	*cache.Ingress
	*config.Config

	client types.Connection
}

// NewIngress return a new Ingress scruber.
func NewIngress(ctx context.Context, c *Cache, codes *issues.Codes) Sanitizer {
	d := Ingress{
		client:    c.factory.Client(),
		Config:    c.config,
		Collector: issues.NewCollector(codes, c.config),
	}

	var err error
	d.Ingress, err = c.ingresses()
	if err != nil {
		d.AddErr(ctx, err)
	}

	return &d
}

// Sanitize all available Ingresss.
func (i *Ingress) Sanitize(ctx context.Context) error {
	return sanitize.NewIngress(i.Collector, i).Sanitize(ctx)
}
