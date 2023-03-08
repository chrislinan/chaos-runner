package scrub

import (
	"context"

	"github.com/litmuschaos/chaos-runner/pkg/internal/cache"
	"github.com/litmuschaos/chaos-runner/pkg/internal/issues"
	"github.com/litmuschaos/chaos-runner/pkg/internal/sanitize"
)

// Service represents a Service scruber.
type Service struct {
	*issues.Collector
	*cache.Service
	*cache.Pod
	*cache.Endpoints
}

// NewService return a new Service scruber.
func NewService(ctx context.Context, c *Cache, codes *issues.Codes) Sanitizer {
	s := Service{Collector: issues.NewCollector(codes, c.config)}

	var err error
	s.Service, err = c.services()
	if err != nil {
		s.AddErr(ctx, err)
	}

	s.Pod, err = c.pods()
	if err != nil {
		s.AddErr(ctx, err)
	}

	s.Endpoints, err = c.endpoints()
	if err != nil {
		s.AddErr(ctx, err)
	}

	return &s
}

// Sanitize all available Services.
func (s *Service) Sanitize(ctx context.Context) error {
	return sanitize.NewService(s.Collector, s).Sanitize(ctx)
}
