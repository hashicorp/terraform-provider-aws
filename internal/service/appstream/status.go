package appstream

import (
	"context"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/appstream"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

// statusStackState fetches the fleet and its state
func statusStackState(ctx context.Context, conn *appstream.AppStream, name string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		stack, err := FindStackByName(ctx, conn, name)
		if err != nil {
			return nil, "Unknown", err
		}

		if stack == nil {
			return stack, "NotFound", nil
		}

		return stack, "AVAILABLE", nil
	}
}

// statusFleetState fetches the fleet and its state
func statusFleetState(ctx context.Context, conn *appstream.AppStream, name string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		fleet, err := FindFleetByName(ctx, conn, name)

		if err != nil {
			return nil, "Unknown", err
		}

		if fleet == nil {
			return fleet, "NotFound", nil
		}

		return fleet, aws.StringValue(fleet.State), nil
	}
}

func statusImageBuilderState(ctx context.Context, conn *appstream.AppStream, name string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		output, err := FindImageBuilderByName(ctx, conn, name)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, aws.StringValue(output.State), nil
	}
}

// statusUserAvailable fetches the user available
func statusUserAvailable(ctx context.Context, conn *appstream.AppStream, username, authType string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		user, err := FindUserByUserNameAndAuthType(ctx, conn, username, authType)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return user, userAvailable, nil
	}
}
