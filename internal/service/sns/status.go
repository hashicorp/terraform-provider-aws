package sns

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/sns"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

func statusSubscriptionPendingConfirmation(conn *sns.SNS, id string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		output, err := FindSubscriptionByARN(conn, id)
		if err != nil {
			return nil, "", err
		}

		if output == nil {
			return nil, "", nil
		}

		return output, aws.StringValue(output.Attributes["PendingConfirmation"]), nil
	}
}
