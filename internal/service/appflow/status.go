package appflow

import (
	"context"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/appflow"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

func FlowStatus(ctx context.Context, conn *appflow.Appflow, id string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		out, err := FindFlowByArn(ctx, conn, id)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return out, aws.StringValue(out.FlowStatus), nil
	}
}
