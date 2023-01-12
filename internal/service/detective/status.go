package detective

import (
	"context"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/detective"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

// MemberStatus fetches the Member and its status
func MemberStatus(ctx context.Context, conn *detective.Detective, graphARN, adminAccountID string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		output, err := FindMemberByGraphARNAndAccountID(ctx, conn, graphARN, adminAccountID)

		if err != nil {
			return nil, "Unknown", err
		}

		if output == nil {
			return output, "NotFound", nil
		}

		return output, aws.StringValue(output.Status), nil
	}
}
