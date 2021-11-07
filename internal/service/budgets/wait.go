package budgets

import (
	"time"

	"github.com/aws/aws-sdk-go/service/budgets"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

const (
	actionAvailableTimeout = 2 * time.Minute
)

func waitActionAvailable(conn *budgets.Budgets, accountID, actionID, budgetName string) (*budgets.Action, error) { //nolint:unparam
	stateConf := &resource.StateChangeConf{
		Target:  budgets.ActionStatus_Values(),
		Refresh: statusAction(conn, accountID, actionID, budgetName),
		Timeout: actionAvailableTimeout,
	}

	outputRaw, err := stateConf.WaitForState()

	if v, ok := outputRaw.(*budgets.Action); ok {
		return v, err
	}

	return nil, err
}
