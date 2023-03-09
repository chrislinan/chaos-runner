package scrub

import (
	"context"

	"github.com/litmuschaos/chaos-runner/pkg/config"
	"github.com/litmuschaos/chaos-runner/pkg/types"
	"github.com/litmuschaos/chaos-runner/pkg/internal/cache"
	"github.com/litmuschaos/chaos-runner/pkg/internal/issues"
	"github.com/litmuschaos/chaos-runner/pkg/internal/sanitize"
)

// Role represents a Role scruber.
type Role struct {
	client types.Connection
	*config.Config
	*issues.Collector

	*cache.Role
	*cache.ClusterRoleBinding
	*cache.RoleBinding
}

// NewRole return a new Role scruber.
func NewRole(ctx context.Context, c *Cache, codes *issues.Codes) Sanitizer {
	crb := Role{
		client:    c.factory.Client(),
		Config:    c.config,
		Collector: issues.NewCollector(codes, c.config),
	}

	var err error
	crb.Role, err = c.roles()
	if err != nil {
		crb.AddErr(ctx, err)
	}

	crb.ClusterRoleBinding, err = c.clusterrolebindings()
	if err != nil {
		crb.AddCode(ctx, 402, err)
	}

	crb.RoleBinding, err = c.rolebindings()
	if err != nil {
		crb.AddErr(ctx, err)
	}

	return &crb
}

// Sanitize all available Roles.
func (c *Role) Sanitize(ctx context.Context) error {
	return sanitize.NewRole(c.Collector, c).Sanitize(ctx)
}
