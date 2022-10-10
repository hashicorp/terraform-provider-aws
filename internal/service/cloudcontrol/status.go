package cloudcontrol

import (
	"context"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/cloudcontrolapi"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

func statusProgressEventOperation(ctx context.Context, conn *cloudcontrolapi.CloudControlApi, requestToken string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		output, err := FindProgressEventByRequestToken(ctx, conn, requestToken)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, aws.StringValue(output.OperationStatus), nil
	}
}
