package appstream

import (
	"context"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/appstream"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/aws/internal/service/appstream/finder"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	tfappstream "github.com/hashicorp/terraform-provider-aws/internal/service/appstream"
	tfappstream "github.com/hashicorp/terraform-provider-aws/internal/service/appstream"
	tfappstream "github.com/hashicorp/terraform-provider-aws/internal/service/appstream"
)

//statusStackState fetches the fleet and its state
func statusStackState(ctx context.Context, conn *appstream.AppStream, name string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		stack, err := tfappstream.FindStackByName(ctx, conn, name)
		if err != nil {
			return nil, "Unknown", err
		}

		if stack == nil {
			return stack, "NotFound", nil
		}

		return stack, "AVAILABLE", nil
	}
}

//statusFleetState fetches the fleet and its state
func statusFleetState(ctx context.Context, conn *appstream.AppStream, name string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		fleet, err := tfappstream.FindFleetByName(ctx, conn, name)

		if err != nil {
			return nil, "Unknown", err
		}

		if fleet == nil {
			return fleet, "NotFound", nil
		}

		return fleet, aws.StringValue(fleet.State), nil
	}
}

//statusImageBuilderState fetches the ImageBuilder and its state
func statusImageBuilderState(ctx context.Context, conn *appstream.AppStream, name string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		imageBuilder, err := tfappstream.FindImageBuilderByName(ctx, conn, name)

		if err != nil {
			return nil, "", err
		}

		if imageBuilder == nil {
			return nil, "", nil
		}

		return imageBuilder, aws.StringValue(imageBuilder.State), nil
	}
}
