package scrub

import (
	"context"

	"github.com/litmuschaos/chaos-runner/pkg/config"
	"github.com/litmuschaos/chaos-runner/types"
	"github.com/litmuschaos/chaos-runner/pkg/internal/cache"
	"github.com/litmuschaos/chaos-runner/pkg/internal/issues"
	"github.com/litmuschaos/chaos-runner/pkg/internal/sanitize"
)

// RoleBinding represents a RoleBinding scruber.
type RoleBinding struct {
	client types.Connection
	*config.Config
	*issues.Collector

	*cache.RoleBinding
	*cache.ClusterRole
	*cache.Role
}

// NewRoleBinding return a new RoleBinding scruber.
func NewRoleBinding(ctx context.Context, c *Cache, codes *issues.Codes) Sanitizer {
	crb := RoleBinding{
		client:    c.factory.Client(),
		Config:    c.config,
		Collector: issues.NewCollector(codes, c.config),
	}

	var err error
	crb.RoleBinding, err = c.rolebindings()
	if err != nil {
		crb.AddErr(ctx, err)
	}

	crb.ClusterRole, err = c.clusterroles()
	if err != nil {
		crb.AddCode(ctx, 402, err)
	}

	crb.Role, err = c.roles()
	if err != nil {
		crb.AddErr(ctx, err)
	}

	return &crb
}

// Sanitize all available RoleBindings.
func (c *RoleBinding) Sanitize(ctx context.Context) error {
	return sanitize.NewRoleBinding(c.Collector, c).Sanitize(ctx)
}
