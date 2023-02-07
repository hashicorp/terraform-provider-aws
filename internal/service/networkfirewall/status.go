package networkfirewall

import (
	"context"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/networkfirewall"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

const (
	resourceStatusFailed  = "Failed"
	resourceStatusUnknown = "Unknown"
	resourceStatusDeleted = "Deleted"
)

// statusRuleGroup fetches the Rule Group and its Status
func statusRuleGroup(ctx context.Context, conn *networkfirewall.NetworkFirewall, arn string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		input := &networkfirewall.DescribeRuleGroupInput{
			RuleGroupArn: aws.String(arn),
		}

		output, err := conn.DescribeRuleGroupWithContext(ctx, input)

		if tfawserr.ErrCodeEquals(err, networkfirewall.ErrCodeResourceNotFoundException) {
			return output, resourceStatusDeleted, nil
		}

		if err != nil {
			return nil, resourceStatusUnknown, err
		}

		if output == nil || output.RuleGroupResponse == nil {
			return nil, resourceStatusUnknown, nil
		}

		return output.RuleGroup, aws.StringValue(output.RuleGroupResponse.RuleGroupStatus), nil
	}
}
