package applicationinsights

import (
	"context"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/applicationinsights"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

func statusApplication(ctx context.Context, conn *applicationinsights.ApplicationInsights, name string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		output, err := FindApplicationByName(ctx, conn, name)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, aws.StringValue(output.LifeCycle), nil
	}
}
