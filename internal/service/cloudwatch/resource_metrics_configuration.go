// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package cloudwatch

import (
	"context"

	"github.com/YakDriver/smarterr"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/cloudwatch"
	awstypes "github.com/aws/aws-sdk-go-v2/service/cloudwatch/types"
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/setvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/fwdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	fwflex "github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/smerr"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkResource("aws_cloudwatch_resource_metrics_configuration", name="Resource Metrics Configuration")
// @ArnIdentity("resource_arn")
// @Testing(hasNoPreExistingResource=true)
// @Testing(preCheck="testAccPreCheckResourceMetricsConfiguration")
func newResourceMetricsConfigurationResource(_ context.Context) (resource.ResourceWithConfigure, error) {
	return &resourceMetricsConfigurationResource{}, nil
}

type resourceMetricsConfigurationResource struct {
	framework.ResourceWithModel[resourceMetricsConfigurationResourceModel]
	framework.WithImportByIdentity
}

func (r *resourceMetricsConfigurationResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			names.AttrResourceARN: schema.StringAttribute{
				CustomType: fwtypes.ARNType,
				Required:   true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
		},
		Blocks: map[string]schema.Block{
			"metric_selection": schema.ListNestedBlock{
				CustomType: fwtypes.NewListNestedObjectTypeOf[resourceMetricSelectionModel](ctx),
				Validators: []validator.List{
					listvalidator.SizeAtMost(1),
				},
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"include_metrics": schema.SetAttribute{
							CustomType: fwtypes.SetOfStringType,
							Required:   true,
							Validators: []validator.Set{
								setvalidator.SizeBetween(1, 500),
								setvalidator.ValueStringsAre(
									stringvalidator.LengthBetween(1, 255),
								),
							},
						},
					},
				},
			},
		},
	}
}

func (r *resourceMetricsConfigurationResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan resourceMetricsConfigurationResourceModel
	smerr.AddEnrich(ctx, &resp.Diagnostics, req.Plan.Get(ctx, &plan))
	if resp.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().CloudWatchClient(ctx)

	resourceARN := fwflex.StringValueFromFramework(ctx, plan.ResourceARN)
	var input cloudwatch.CreateResourceMetricsConfigurationInput
	smerr.AddEnrich(ctx, &resp.Diagnostics, fwflex.Expand(ctx, plan, &input))
	if resp.Diagnostics.HasError() {
		return
	}

	_, err := conn.CreateResourceMetricsConfiguration(ctx, &input)
	if err != nil {
		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, resourceARN)
		return
	}

	smerr.AddEnrich(ctx, &resp.Diagnostics, resp.State.Set(ctx, &plan))
}

func (r *resourceMetricsConfigurationResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state resourceMetricsConfigurationResourceModel
	smerr.AddEnrich(ctx, &resp.Diagnostics, req.State.Get(ctx, &state))
	if resp.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().CloudWatchClient(ctx)

	resourceARN := fwflex.StringValueFromFramework(ctx, state.ResourceARN)
	out, err := findResourceMetricsConfigurationByARN(ctx, conn, resourceARN)
	if retry.NotFound(err) {
		smerr.AddOne(ctx, &resp.Diagnostics, fwdiag.NewResourceNotFoundWarningDiagnostic(err))
		resp.State.RemoveResource(ctx)
		return
	}
	if err != nil {
		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, resourceARN)
		return
	}

	smerr.AddEnrich(ctx, &resp.Diagnostics, r.flatten(ctx, out, &state))
	if resp.Diagnostics.HasError() {
		return
	}

	smerr.AddEnrich(ctx, &resp.Diagnostics, resp.State.Set(ctx, &state))
}

func (r *resourceMetricsConfigurationResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan, state resourceMetricsConfigurationResourceModel
	smerr.AddEnrich(ctx, &resp.Diagnostics, req.Plan.Get(ctx, &plan))
	smerr.AddEnrich(ctx, &resp.Diagnostics, req.State.Get(ctx, &state))
	if resp.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().CloudWatchClient(ctx)

	if !plan.MetricSelections.Equal(state.MetricSelections) {
		resourceARN := fwflex.StringValueFromFramework(ctx, plan.ResourceARN)
		var input cloudwatch.UpdateResourceMetricsConfigurationInput
		smerr.AddEnrich(ctx, &resp.Diagnostics, fwflex.Expand(ctx, plan, &input))
		if resp.Diagnostics.HasError() {
			return
		}

		// Omitting MetricSelections clears any existing selection filter, so an
		// empty plan value is expanded to a nil slice intentionally.
		_, err := conn.UpdateResourceMetricsConfiguration(ctx, &input)
		if err != nil {
			smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, resourceARN)
			return
		}
	}

	smerr.AddEnrich(ctx, &resp.Diagnostics, resp.State.Set(ctx, &plan))
}

func (r *resourceMetricsConfigurationResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state resourceMetricsConfigurationResourceModel
	smerr.AddEnrich(ctx, &resp.Diagnostics, req.State.Get(ctx, &state))
	if resp.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().CloudWatchClient(ctx)

	resourceARN := fwflex.StringValueFromFramework(ctx, state.ResourceARN)
	input := cloudwatch.DeleteResourceMetricsConfigurationInput{
		ResourceArn: aws.String(resourceARN),
	}
	_, err := conn.DeleteResourceMetricsConfiguration(ctx, &input)
	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return
	}
	if err != nil {
		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, resourceARN)
		return
	}
}

func (r *resourceMetricsConfigurationResource) flatten(ctx context.Context, out *awstypes.ResourceMetricsConfiguration, data *resourceMetricsConfigurationResourceModel) diag.Diagnostics {
	var diags diag.Diagnostics
	// Terraform represents an absent block as an empty list, but the API omits
	// MetricSelections entirely when no selection filter is set, which flattens
	// to a null list. Normalize so that omitting the block does not produce a
	// perpetual diff.
	if len(out.MetricSelections) == 0 {
		out.MetricSelections = nil
	}
	diags.Append(fwflex.Flatten(ctx, out, data)...)
	return diags
}

func findResourceMetricsConfiguration(ctx context.Context, conn *cloudwatch.Client, input *cloudwatch.GetResourceMetricsConfigurationInput) (*awstypes.ResourceMetricsConfiguration, error) {
	out, err := conn.GetResourceMetricsConfiguration(ctx, input)
	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return nil, smarterr.NewError(&retry.NotFoundError{
			LastError: err,
		})
	}
	if err != nil {
		return nil, smarterr.NewError(err)
	}

	if out == nil || out.ResourceMetricsConfiguration == nil {
		return nil, smarterr.NewError(tfresource.NewEmptyResultError())
	}

	return out.ResourceMetricsConfiguration, nil
}

func findResourceMetricsConfigurationByARN(ctx context.Context, conn *cloudwatch.Client, resourceARN string) (*awstypes.ResourceMetricsConfiguration, error) {
	input := cloudwatch.GetResourceMetricsConfigurationInput{
		ResourceArn: aws.String(resourceARN),
	}

	out, err := findResourceMetricsConfiguration(ctx, conn, &input)
	if err != nil {
		return nil, smarterr.NewError(err)
	}

	return out, nil
}

type resourceMetricsConfigurationResourceModel struct {
	framework.WithRegionModel
	MetricSelections fwtypes.ListNestedObjectValueOf[resourceMetricSelectionModel] `tfsdk:"metric_selection"`
	ResourceARN      fwtypes.ARN                                                   `tfsdk:"resource_arn"`
}

type resourceMetricSelectionModel struct {
	IncludeMetrics fwtypes.SetOfString `tfsdk:"include_metrics"`
}
