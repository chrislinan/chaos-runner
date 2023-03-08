package scrub

import (
	"context"

	"github.com/litmuschaos/chaos-runner/pkg/config"
	"github.com/litmuschaos/chaos-runner/pkg/internal/cache"
	"github.com/litmuschaos/chaos-runner/pkg/internal/issues"
	"github.com/litmuschaos/chaos-runner/pkg/internal/sanitize"
)

// Node represents a Node scruber.
type Node struct {
	*issues.Collector
	*cache.Node
	*cache.Pod
	*cache.NodesMetrics
	*config.Config
}

// NewNode return a new Node scruber.
func NewNode(ctx context.Context, c *Cache, codes *issues.Codes) Sanitizer {
	n := Node{
		Collector: issues.NewCollector(codes, c.config),
		Config:    c.config,
	}

	var err error
	n.Node, err = c.nodes()
	if err != nil {
		n.AddErr(ctx, err)
	}

	n.Pod, err = c.pods()
	if err != nil {
		n.AddErr(ctx, err)
	}

	n.NodesMetrics, _ = c.nodesMx()

	return &n
}

// Sanitize all available Nodes.
func (n *Node) Sanitize(ctx context.Context) error {
	return sanitize.NewNode(n.Collector, n).Sanitize(ctx)
}
