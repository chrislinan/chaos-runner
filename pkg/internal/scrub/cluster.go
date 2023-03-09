package scrub

import (
	"context"

	"github.com/litmuschaos/chaos-runner/pkg/config"
	"github.com/litmuschaos/chaos-runner/pkg/types"
	"github.com/litmuschaos/chaos-runner/pkg/internal/cache"
	"github.com/litmuschaos/chaos-runner/pkg/internal/issues"
	"github.com/litmuschaos/chaos-runner/pkg/internal/sanitize"
)

// Cluster represents a Cluster scruber.
type Cluster struct {
	*issues.Collector
	*cache.Cluster
	*config.Config

	client types.Connection
}

// NewCluster return a new Cluster scruber.
func NewCluster(ctx context.Context, c *Cache, codes *issues.Codes) Sanitizer {
	cl := Cluster{
		client:    c.factory.Client(),
		Config:    c.config,
		Collector: issues.NewCollector(codes, c.config),
	}

	var err error
	cl.Cluster, err = c.cluster()
	if err != nil {
		cl.AddErr(ctx, err)
	}

	return &cl
}

// Sanitize all available Clusters.
func (d *Cluster) Sanitize(ctx context.Context) error {
	return sanitize.NewCluster(d.Collector, d).Sanitize(ctx)
}
