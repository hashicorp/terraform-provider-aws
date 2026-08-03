// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package arcregionswitch

import (
	"context"

	awstypes "github.com/aws/aws-sdk-go-v2/service/arcregionswitch/types"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	fwschema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	fwflex "github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	fwvalidators "github.com/hashicorp/terraform-provider-aws/internal/framework/validators"
	"github.com/hashicorp/terraform-provider-aws/internal/smerr"
)

// @FrameworkDataSource("aws_arcregionswitch_route53_health_checks", name="Route53 Health Checks")
// @Region(overrideDeprecated=true)
func newRoute53HealthChecksDataSource(context.Context) (datasource.DataSourceWithConfigure, error) {
	return &route53HealthChecksDataSource{}, nil
}

type route53HealthChecksDataSource struct {
	framework.DataSourceWithModel[route53HealthChecksDataSourceModel]
}

func (d *route53HealthChecksDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = fwschema.Schema{
		Attributes: map[string]fwschema.Attribute{
			"plan_arn": fwschema.StringAttribute{
				CustomType: fwtypes.ARNType,
				Required:   true,
				Validators: []validator.String{
					fwvalidators.ARN(),
				},
			},
			"health_checks": framework.DataSourceComputedListOfObjectAttribute[route53HealthCheckModel](ctx),
		},
	}
}

func (d *route53HealthChecksDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data route53HealthChecksDataSourceModel
	smerr.AddEnrich(ctx, &resp.Diagnostics, req.Config.Get(ctx, &data))
	if resp.Diagnostics.HasError() {
		return
	}

	conn := d.Meta().ARCRegionSwitchClient(ctx)

	healthChecks, err := findRoute53HealthChecksByARN(ctx, conn, data.PlanARN.ValueString())
	if err != nil {
		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, data.PlanARN.ValueString())
		return
	}

	smerr.AddEnrich(ctx, &resp.Diagnostics, fwflex.Flatten(ctx, healthChecks, &data.HealthChecks), smerr.ID, data.PlanARN.ValueString())
	if resp.Diagnostics.HasError() {
		return
	}

	smerr.AddEnrich(ctx, &resp.Diagnostics, resp.State.Set(ctx, &data))
}

type route53HealthChecksDataSourceModel struct {
	framework.WithRegionModel
	route53HealthChecksModel
}

type route53HealthChecksModel struct {
	PlanARN      fwtypes.ARN                                              `tfsdk:"plan_arn"`
	HealthChecks fwtypes.ListNestedObjectValueOf[route53HealthCheckModel] `tfsdk:"health_checks"`
}

type route53HealthCheckModel struct {
	HealthCheckID types.String                                          `tfsdk:"health_check_id"`
	HostedZoneID  types.String                                          `tfsdk:"hosted_zone_id"`
	RecordName    types.String                                          `tfsdk:"record_name"`
	Region        types.String                                          `tfsdk:"region"`
	Status        fwtypes.StringEnum[awstypes.Route53HealthCheckStatus] `tfsdk:"status"`
}
