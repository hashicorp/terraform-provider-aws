package sfn

import (
	"time"

	"github.com/aws/aws-sdk-go/service/sfn"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

const (
	stateMachineCreatedTimeout = 5 * time.Minute
	stateMachineDeletedTimeout = 5 * time.Minute
	stateMachineUpdatedTimeout = 1 * time.Minute
)

func waitStateMachineDeleted(conn *sfn.SFN, stateMachineArn string) (*sfn.DescribeStateMachineOutput, error) {
	stateConf := &resource.StateChangeConf{
		Pending: []string{sfn.StateMachineStatusActive, sfn.StateMachineStatusDeleting},
		Target:  []string{},
		Refresh: statusStateMachine(conn, stateMachineArn),
		Timeout: stateMachineDeletedTimeout,
	}

	outputRaw, err := stateConf.WaitForState()

	if output, ok := outputRaw.(*sfn.DescribeStateMachineOutput); ok {
		return output, err
	}

	return nil, err
}
