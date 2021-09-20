package budgets

import (
	"time"

	"github.com/aws/aws-sdk-go/service/budgets"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

const (
	actionAvailableTimeout = 2 * time.Minute
)

func waitActionAvailable(conn *budgets.Budgets, accountID, actionID, budgetName string) (*budgets.Action, error) {
	stateConf := &resource.StateChangeConf{
		Pending: []string{
			budgets.ActionStatusExecutionInProgress,
			budgets.ActionStatusStandby,
		},
		Target: []string{
			budgets.ActionStatusExecutionSuccess,
			budgets.ActionStatusExecutionFailure,
			budgets.ActionStatusPending,
		},
		Refresh: statusAction(conn, accountID, actionID, budgetName),
		Timeout: actionAvailableTimeout,
	}

	outputRaw, err := stateConf.WaitForState()

	if v, ok := outputRaw.(*budgets.Action); ok {
		return v, err
	}

	return nil, err
}
