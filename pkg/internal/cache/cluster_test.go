package cache_test

import (
	"testing"

	"github.com/litmuschaos/chaos-runner/pkg/internal/cache"
	"github.com/stretchr/testify/assert"
)

func TestCluster(t *testing.T) {
	c := cache.NewCluster("1", "9")

	ma, mi := c.ListVersion()
	assert.Equal(t, "1", ma)
	assert.Equal(t, "9", mi)
}
