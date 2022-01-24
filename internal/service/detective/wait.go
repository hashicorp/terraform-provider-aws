package detective

import (
	"context"
	"time"

	"github.com/aws/aws-sdk-go/service/detective"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

const (
	// GraphOperationTimeout Maximum amount of time to wait for a detective graph to be created, deleted
	GraphOperationTimeout = 4 * time.Minute
	// MemberStatusPropagationTimeout Maximum amount of time to wait for a detective member status to return Invited
	MemberStatusPropagationTimeout = 4 * time.Minute
)

// MemberStatusUpdated waits for an AdminAccount and graph arn to return Invited
func MemberStatusUpdated(ctx context.Context, conn *detective.Detective, graphARN, adminAccountID, expectedValue string) (*detective.MemberDetail, error) {
	stateConf := &resource.StateChangeConf{
		Pending: []string{detective.MemberStatusVerificationInProgress},
		Target:  []string{detective.MemberStatusInvited},
		Refresh: MemberStatus(ctx, conn, graphARN, adminAccountID),
		Timeout: MemberStatusPropagationTimeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*detective.MemberDetail); ok {
		return output, err
	}

	return nil, err
}
