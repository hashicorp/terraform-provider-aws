package modifiers

import (
	"context"
	awstypes "github.com/aws/aws-sdk-go-v2/service/s3/types"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/service/s3/models"
)

type dataRedundancyPlanModifier struct{}

func (d dataRedundancyPlanModifier) Description(ctx context.Context) string {
	return "Sets default value for data_redundancy based on location type."
}

func (d dataRedundancyPlanModifier) MarkdownDescription(ctx context.Context) string {
	return "Sets default value for `data_redundancy` based on the `location` type."
}

func (d dataRedundancyPlanModifier) PlanModifyString(ctx context.Context, req planmodifier.StringRequest, resp *planmodifier.StringResponse) {
	if !req.ConfigValue.IsNull() {
		return
	}

	// Fetch the location list attribute.
	var locationList fwtypes.ListNestedObjectValueOf[models.LocationInfoModel]
	diags := req.Config.GetAttribute(ctx, path.Root("location"), &locationList)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	if locationList.IsNull() || len(locationList.Elements()) == 0 {
		return
	}

	// Extract location information.
	locationInfoModel, Diags := locationList.ToPtr(ctx)
	resp.Diagnostics.Append(Diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Determine the location type (defaulting to AvailabilityZone if unspecified).
	var locationType awstypes.LocationType
	if locationInfoModel.Type.IsNull() || locationInfoModel.Type.ValueEnum() == "" {
		locationType = awstypes.LocationTypeAvailabilityZone
	} else {
		locationType = locationInfoModel.Type.ValueEnum()
	}

	// Set the default value for data_redundancy based on the location type.
	switch locationType {
	case awstypes.LocationTypeLocalZone:
		resp.PlanValue = types.StringValue(string(awstypes.DataRedundancySingleLocalZone))
	default:
		resp.PlanValue = types.StringValue(string(awstypes.DataRedundancySingleAvailabilityZone))
	}
}

func ApplyDataRedundancyPlanModifier() []planmodifier.String {
	return []planmodifier.String{
		dataRedundancyPlanModifier{},
	}
}
