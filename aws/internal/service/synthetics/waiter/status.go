package waiter

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/synthetics"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/aws/internal/service/synthetics/finder"
	"github.com/hashicorp/terraform-provider-aws/aws/internal/tfresource"
)

func CanaryState(conn *synthetics.Synthetics, name string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		output, err := finder.CanaryByName(conn, name)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, aws.StringValue(output.Status.State), nil
	}
}
