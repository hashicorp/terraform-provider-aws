package gamelift

import (
	"context"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/gamelift"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

func statusBuild(ctx context.Context, conn *gamelift.GameLift, id string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		output, err := FindBuildByID(ctx, conn, id)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, aws.StringValue(output.Status), nil
	}
}

func statusFleet(ctx context.Context, conn *gamelift.GameLift, id string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		output, err := FindFleetByID(ctx, conn, id)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, aws.StringValue(output.Status), nil
	}
}

func statusGameServerGroup(ctx context.Context, conn *gamelift.GameLift, name string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		output, err := FindGameServerGroupByName(ctx, conn, name)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, aws.StringValue(output.Status), nil
	}
}
