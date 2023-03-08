package scrub

import (
	"context"

	"github.com/litmuschaos/chaos-runner/pkg/config"
	"github.com/litmuschaos/chaos-runner/pkg/internal/issues"
)

// Sanitizer represents a resource sanitizer.
type Sanitizer interface {
	Collector
	Sanitize(context.Context) error
}

// Collector collects sanitization issues.
type Collector interface {
	MaxSeverity(res string) config.Level
	Outcome() issues.Outcome
}
