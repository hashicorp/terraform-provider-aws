package detective

import (
	"context"
	"time"

	"github.com/aws/aws-sdk-go/service/detective"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

const (
	// DetectiveOperationTimeout Maximum amount of time to wait for a detective graph to be created, deleted
	DetectiveOperationTimeout = 4 * time.Minute
)

// MemberInvited waits for an AdminAccount and graph arn to return Invited
func MemberInvited(ctx context.Context, conn *detective.Detective, graphARN, adminAccountID string) (*detective.MemberDetail, error) {
	stateConf := &resource.StateChangeConf{
		Pending: []string{detective.MemberStatusVerificationInProgress},
		Target:  []string{detective.MemberStatusInvited},
		Refresh: StatusMember(ctx, conn, graphARN, adminAccountID),
		Timeout: DetectiveOperationTimeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*detective.MemberDetail); ok {
		return output, err
	}

	return nil, err
}
