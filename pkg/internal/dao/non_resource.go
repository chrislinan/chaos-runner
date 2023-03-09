package dao

import (
	"context"
	"fmt"

	"github.com/litmuschaos/chaos-runner/pkg/types"
	"github.com/litmuschaos/chaos-runner/pkg/internal/client"
	"k8s.io/apimachinery/pkg/runtime"
)

// NonResource represents a non k8s resource.
type NonResource struct {
	types.Factory

	gvr client.GVR
}

// Init initializes the resource.
func (n *NonResource) Init(f types.Factory, gvr client.GVR) {
	n.Factory, n.gvr = f, gvr
}

// GVR returns a gvr.
func (n *NonResource) GVR() string {
	return n.gvr.String()
}

// Get returns the given resource.
func (n *NonResource) Get(context.Context, string) (runtime.Object, error) {
	return nil, fmt.Errorf("NYI!")
}
