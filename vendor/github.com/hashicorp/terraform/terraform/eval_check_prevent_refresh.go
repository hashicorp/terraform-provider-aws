package terraform

import (
	"fmt"

	"github.com/hashicorp/terraform/config"
)

// EvalCheckDataDependsOn is an EvalNode implementation that returns an
// error if a data source has an explicit dependency that contains a diff. If
// the dependency has a diff, the data source refresh can't be completed until
// apply.
type EvalCheckDataDependsOn struct {
	Refresh  bool
	Info     *InstanceInfo
	Config   *config.Resource
	Provider *ResourceProvider
	State    **InstanceState
}

func (n *EvalCheckDataDependsOn) Eval(ctx EvalContext) (interface{}, error) {
	if len(n.Config.DependsOn) == 0 {
		return nil, nil
	}

	state := *n.State
	provider := *n.Provider

	// The state for the diff must never be nil
	diffState := state
	if diffState == nil {
		diffState = new(InstanceState)
	}
	diffState.init()

	resourceCfg := new(ResourceConfig)

	diff, err := provider.Diff(n.Info, diffState, resourceCfg)
	if err != nil {
		return nil, err
	}

	fmt.Printf("Refresh:%t Name:%s Modes:%s DIFF: %#v\n", n.Refresh, n.Config.Name, n.Config.Mode, diff)
	if len(n.Config.DependsOn) > 0 {
		return nil, EvalEarlyExitError{}
	}

	return nil, nil
}
