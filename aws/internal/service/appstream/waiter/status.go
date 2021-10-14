package waiter

import (
	"context"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/appstream"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/aws/internal/service/appstream/finder"
)

//StackState fetches the fleet and its state
func StackState(ctx context.Context, conn *appstream.AppStream, name string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		stack, err := finder.StackByName(ctx, conn, name)
		if err != nil {
			return nil, "Unknown", err
		}

		if stack == nil {
			return stack, "NotFound", nil
		}

		return stack, "AVAILABLE", nil
	}
}

//FleetState fetches the fleet and its state
func FleetState(ctx context.Context, conn *appstream.AppStream, name string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		fleet, err := finder.FleetByName(ctx, conn, name)

		if err != nil {
			return nil, "Unknown", err
		}

		if fleet == nil {
			return fleet, "NotFound", nil
		}

		return fleet, aws.StringValue(fleet.State), nil
	}
}

//ImageBuilderState fetches the ImageBuilder and its state
func ImageBuilderState(ctx context.Context, conn *appstream.AppStream, name string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		imageBuilder, err := finder.ImageBuilderByName(ctx, conn, name)

		if err != nil {
			return nil, "", err
		}

		if imageBuilder == nil {
			return nil, "", nil
		}

		return imageBuilder, aws.StringValue(imageBuilder.State), nil
	}
}
