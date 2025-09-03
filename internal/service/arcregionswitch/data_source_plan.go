// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package arcregionswitch

import (
	"context"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/arcregionswitch"
	awstypes "github.com/aws/aws-sdk-go-v2/service/arcregionswitch/types"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	fwschema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	fwdiag "github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	fwvalidators "github.com/hashicorp/terraform-provider-aws/internal/framework/validators"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func findRoute53HealthChecks(ctx context.Context, conn *arcregionswitch.Client, planArn string) ([]awstypes.Route53HealthCheck, error) {
	input := &arcregionswitch.ListRoute53HealthChecksInput{
		Arn: aws.String(planArn),
	}

	output, err := conn.ListRoute53HealthChecks(ctx, input)

	if err != nil {
		return nil, err
	}

	return output.HealthChecks, nil
}

// @FrameworkDataSource("aws_arcregionswitch_plan", name="Plan")
func newDataSourcePlan(context.Context) (datasource.DataSourceWithConfigure, error) {
	d := &dataSourcePlan{}
	return d, nil
}

type dataSourcePlan struct {
	framework.DataSourceWithConfigure
}

func (d *dataSourcePlan) Metadata(_ context.Context, request datasource.MetadataRequest, response *datasource.MetadataResponse) {
	response.TypeName = request.ProviderTypeName + "_arcregionswitch_plan"
}

func (d *dataSourcePlan) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = fwschema.Schema{
		Attributes: map[string]fwschema.Attribute{
			names.AttrARN: fwschema.StringAttribute{
				Required: true,
				Validators: []validator.String{
					fwvalidators.ARN(),
				},
			},
			names.AttrName: fwschema.StringAttribute{
				Computed: true,
			},
			"execution_role": fwschema.StringAttribute{
				Computed: true,
			},
			"recovery_approach": fwschema.StringAttribute{
				Computed: true,
			},
			"regions": fwschema.ListAttribute{
				Computed:    true,
				ElementType: types.StringType,
			},
			names.AttrDescription: fwschema.StringAttribute{
				Computed: true,
			},
			"primary_region": fwschema.StringAttribute{
				Computed: true,
			},
			"recovery_time_objective_minutes": fwschema.Int64Attribute{
				Computed: true,
			},
			"wait_for_health_checks": fwschema.BoolAttribute{
				Optional:    true,
				Description: "Wait for Route53 health check IDs to be populated (takes ~4 minutes)",
			},
			"route53_health_checks": fwschema.ListAttribute{
				ElementType: types.ObjectType{
					AttrTypes: map[string]attr.Type{
						"health_check_id":      types.StringType,
						names.AttrHostedZoneID: types.StringType,
						"record_name":          types.StringType,
						names.AttrRegion:       types.StringType,
					},
				},
				Computed:    true,
				Description: "Route53 health checks associated with the plan",
			},
		},
		Blocks: map[string]fwschema.Block{},
	}
}

type dataSourcePlanModel struct {
	ARN                          types.String `tfsdk:"arn"`
	Region                       types.String `tfsdk:"region"`
	Name                         types.String `tfsdk:"name"`
	ExecutionRole                types.String `tfsdk:"execution_role"`
	RecoveryApproach             types.String `tfsdk:"recovery_approach"`
	Regions                      types.List   `tfsdk:"regions"`
	Description                  types.String `tfsdk:"description"`
	PrimaryRegion                types.String `tfsdk:"primary_region"`
	RecoveryTimeObjectiveMinutes types.Int64  `tfsdk:"recovery_time_objective_minutes"`
	WaitForHealthChecks          types.Bool   `tfsdk:"wait_for_health_checks"`
	Route53HealthChecks          types.List   `tfsdk:"route53_health_checks"`
}

func (d *dataSourcePlan) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data dataSourcePlanModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	conn := d.Meta().ARCRegionSwitchClient(ctx)

	plan, err := FindPlanByARN(ctx, conn, data.ARN.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("reading ARC Region Switch Plan", err.Error())
		return
	}

	data.ARN = types.StringValue(aws.ToString(plan.Arn))
	data.Name = types.StringValue(aws.ToString(plan.Name))
	data.ExecutionRole = types.StringValue(aws.ToString(plan.ExecutionRole))
	data.RecoveryApproach = types.StringValue(string(plan.RecoveryApproach))

	regions, diags := types.ListValueFrom(ctx, types.StringType, plan.Regions)
	resp.Diagnostics.Append(diags...)
	data.Regions = regions

	if plan.Description != nil {
		data.Description = types.StringValue(aws.ToString(plan.Description))
	} else {
		data.Description = types.StringNull()
	}

	if plan.PrimaryRegion != nil {
		data.PrimaryRegion = types.StringValue(aws.ToString(plan.PrimaryRegion))
	} else {
		data.PrimaryRegion = types.StringNull()
	}

	if plan.RecoveryTimeObjectiveMinutes != nil {
		data.RecoveryTimeObjectiveMinutes = types.Int64Value(int64(aws.ToInt32(plan.RecoveryTimeObjectiveMinutes)))
	} else {
		data.RecoveryTimeObjectiveMinutes = types.Int64Null()
	}

	// Always fetch Route53 health checks
	var healthChecks []awstypes.Route53HealthCheck
	var healthCheckErr error

	if data.WaitForHealthChecks.ValueBool() {
		// Wait for health check IDs to be populated (takes ~4 minutes)
		timeout := 5 * time.Minute
		healthCheckErr = retry.RetryContext(ctx, timeout, func() *retry.RetryError {
			healthChecks, healthCheckErr = findRoute53HealthChecks(ctx, conn, data.ARN.ValueString())
			if healthCheckErr != nil {
				return retry.NonRetryableError(healthCheckErr)
			}

			// Check if all health check IDs are populated
			for _, hc := range healthChecks {
				if aws.ToString(hc.HealthCheckId) == "" {
					return retry.RetryableError(fmt.Errorf("waiting for Route53 health check IDs to be populated"))
				}
			}

			return nil
		})
		if tfresource.TimedOut(healthCheckErr) {
			healthChecks, healthCheckErr = findRoute53HealthChecks(ctx, conn, data.ARN.ValueString())
			if healthCheckErr != nil {
				resp.Diagnostics.AddError("reading Route53 health checks", healthCheckErr.Error())
				return
			}
		}
		if healthCheckErr != nil {
			resp.Diagnostics.AddError("waiting for Route53 health checks", healthCheckErr.Error())
			return
		}
	} else {
		// Fetch health checks without waiting
		healthChecks, healthCheckErr = findRoute53HealthChecks(ctx, conn, data.ARN.ValueString())
		if healthCheckErr != nil {
			resp.Diagnostics.AddError("listing Route53 health checks", healthCheckErr.Error())
			return
		}
	}

	// Convert health checks to Framework types
	if len(healthChecks) == 0 {
		// Return known empty list (not unknown) when plan exists but has no health checks
		data.Route53HealthChecks = types.ListValueMust(types.ObjectType{
			AttrTypes: map[string]attr.Type{
				"health_check_id":      types.StringType,
				names.AttrHostedZoneID: types.StringType,
				"record_name":          types.StringType,
				names.AttrRegion:       types.StringType,
			},
		}, []attr.Value{})
		resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
		return
	}

	healthCheckElements := make([]attr.Value, len(healthChecks))
	for i, hc := range healthChecks {
		healthCheckAttrs := map[string]attr.Value{
			"health_check_id":      types.StringValue(aws.ToString(hc.HealthCheckId)),
			names.AttrHostedZoneID: types.StringValue(aws.ToString(hc.HostedZoneId)),
			"record_name":          types.StringValue(aws.ToString(hc.RecordName)),
			names.AttrRegion:       types.StringValue(aws.ToString(hc.Region)),
		}
		healthCheckObj, objDiags := types.ObjectValue(map[string]attr.Type{
			"health_check_id":      types.StringType,
			names.AttrHostedZoneID: types.StringType,
			"record_name":          types.StringType,
			names.AttrRegion:       types.StringType,
		}, healthCheckAttrs)
		resp.Diagnostics.Append(objDiags...)
		healthCheckElements[i] = healthCheckObj
	}

	healthChecksList, listDiags := types.ListValue(types.ObjectType{
		AttrTypes: map[string]attr.Type{
			"health_check_id":      types.StringType,
			names.AttrHostedZoneID: types.StringType,
			"record_name":          types.StringType,
			names.AttrRegion:       types.StringType,
		},
	}, healthCheckElements)
	resp.Diagnostics.Append(listDiags...)
	data.Route53HealthChecks = healthChecksList

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (d *dataSourcePlan) ValidateModel(ctx context.Context, schema *fwschema.Schema) fwdiag.Diagnostics {
	var diags fwdiag.Diagnostics
	// Basic validation is handled by the schema validators
	return diags
}
