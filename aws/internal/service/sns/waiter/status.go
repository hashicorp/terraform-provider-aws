package waiter

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/sns"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/aws/internal/service/sns/finder"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

func SubscriptionPendingConfirmation(conn *sns.SNS, id string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		output, err := finder.SubscriptionByARN(conn, id)
		if err != nil {
			return nil, "", err
		}

		if output == nil {
			return nil, "", nil
		}

		return output, aws.StringValue(output.Attributes["PendingConfirmation"]), nil
	}
}
