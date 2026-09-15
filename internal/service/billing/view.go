// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

// DONOTCOPY: Copying old resources spreads bad habits. Use skaff instead.

package billing

import (
	"context"
	"errors"
	"time"

	"github.com/YakDriver/smarterr"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/billing"
	awstypes "github.com/aws/aws-sdk-go-v2/service/billing/types"
	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework-timetypes/timetypes"
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/setvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/listplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/fwdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	"github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/smerr"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkResource("aws_billing_view", name="View")
// @Tags(identifierAttribute="arn")
func newResourceView(_ context.Context) (resource.ResourceWithConfigure, error) {
	r := &resourceView{}

	r.SetDefaultCreateTimeout(5 * time.Minute)
	r.SetDefaultUpdateTimeout(5 * time.Minute)
	r.SetDefaultDeleteTimeout(5 * time.Minute)

	return r, nil
}

const (
	ResNameView = "View"
)

type resourceView struct {
	framework.ResourceWithModel[resourceViewModel]
	framework.WithTimeouts
}

func (r *resourceView) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			names.AttrARN: framework.ARNAttributeComputedOnly(),
			"billing_view_type": schema.StringAttribute{
				CustomType: fwtypes.StringEnumType[awstypes.BillingViewType](),
				Computed:   true,
			},
			names.AttrCreatedAt: schema.StringAttribute{
				CustomType: timetypes.RFC3339Type{},
				Computed:   true,
			},
			"derived_view_count": schema.Int32Attribute{
				Computed: true,
			},
			names.AttrDescription: schema.StringAttribute{
				Optional:   true,
				Validators: []validator.String{stringvalidator.LengthBetween(0, 1024)},
			},
			names.AttrName: schema.StringAttribute{
				Required:   true,
				Validators: []validator.String{stringvalidator.LengthBetween(1, 128)},
			},
			names.AttrOwnerAccountID: schema.StringAttribute{
				Computed: true,
			},
			"source_account_id": schema.StringAttribute{
				Computed: true,
			},
			"source_view_count": schema.Int32Attribute{
				Computed: true,
			},
			"source_views": schema.ListAttribute{
				CustomType:  fwtypes.ListOfStringType,
				ElementType: types.StringType,
				Optional:    true,
				Validators: []validator.List{
					listvalidator.SizeBetween(1, 10),
				},
				PlanModifiers: []planmodifier.List{
					listplanmodifier.RequiresReplace(),
				},
			},
			names.AttrTags:    tftags.TagsAttribute(),
			names.AttrTagsAll: tftags.TagsAttributeComputedOnly(),
			"updated_at": schema.StringAttribute{
				CustomType: timetypes.RFC3339Type{},
				Computed:   true,
			},
			"view_definition_last_updated_at": schema.StringAttribute{
				CustomType: timetypes.RFC3339Type{},
				Computed:   true,
			},
		},
		Blocks: map[string]schema.Block{
			"data_filter_expression": schema.ListNestedBlock{
				CustomType: fwtypes.NewListNestedObjectTypeOf[dataFilterExpressionModel](ctx),
				Validators: []validator.List{
					listvalidator.SizeAtMost(1),
				},
				NestedObject: schema.NestedBlockObject{
					Blocks: map[string]schema.Block{
						"dimensions": schema.ListNestedBlock{
							CustomType: fwtypes.NewListNestedObjectTypeOf[dimensionsModel](ctx),
							Validators: []validator.List{
								listvalidator.SizeAtMost(1),
							},
							NestedObject: schema.NestedBlockObject{
								Attributes: map[string]schema.Attribute{
									names.AttrKey: schema.StringAttribute{
										CustomType: fwtypes.StringEnumType[awstypes.Dimension](),
										Required:   true,
									},
									names.AttrValues: schema.SetAttribute{
										CustomType:  fwtypes.SetOfStringType,
										ElementType: types.StringType,
										Required:    true,
										Validators: []validator.Set{
											setvalidator.SizeAtLeast(1),
										},
									},
								},
							},
						},
						names.AttrTags: schema.ListNestedBlock{
							CustomType: fwtypes.NewListNestedObjectTypeOf[tagValuesModel](ctx),
							NestedObject: schema.NestedBlockObject{
								Attributes: map[string]schema.Attribute{
									names.AttrKey: schema.StringAttribute{
										Required: true,
									},
									names.AttrValues: schema.ListAttribute{
										CustomType:  fwtypes.ListOfStringType,
										ElementType: types.StringType,
										Required:    true,
										Validators: []validator.List{
											listvalidator.SizeAtLeast(1),
										},
									},
								},
							},
						},
						"time_range": schema.ListNestedBlock{
							CustomType: fwtypes.NewListNestedObjectTypeOf[timeRangeModel](ctx),
							Validators: []validator.List{
								listvalidator.SizeAtMost(1),
							},
							NestedObject: schema.NestedBlockObject{
								Attributes: map[string]schema.Attribute{
									"begin_date_inclusive": schema.StringAttribute{
										CustomType: timetypes.RFC3339Type{},
										Required:   true,
									},
									"end_date_inclusive": schema.StringAttribute{
										CustomType: timetypes.RFC3339Type{},
										Required:   true,
									},
								},
							},
						},
					},
				},
			},
			names.AttrTimeouts: timeouts.Block(ctx, timeouts.Opts{
				Create: true,
				Update: true,
				Delete: true,
			}),
		},
	}
}

func (r *resourceView) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	conn := r.Meta().BillingClient(ctx)

	var plan resourceViewModel
	smerr.AddEnrich(ctx, &resp.Diagnostics, req.Plan.Get(ctx, &plan))
	if resp.Diagnostics.HasError() {
		return
	}

	var input billing.CreateBillingViewInput
	// Using WithNoIgnoredFieldNames() because DataFilterExpression.Tags is a filter field,
	// not resource tags. The SDK uses "ResourceTags" for resource tags, not "Tags".
	smerr.AddEnrich(ctx, &resp.Diagnostics, flex.Expand(ctx, plan, &input, flex.WithNoIgnoredFieldNames()))
	if resp.Diagnostics.HasError() {
		return
	}

	input.ResourceTags = getTagsIn(ctx)

	out, err := conn.CreateBillingView(ctx, &input)
	if err != nil {
		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, plan.Name.String())
		return
	}
	if out == nil || out.Arn == nil {
		smerr.AddError(ctx, &resp.Diagnostics, errors.New("empty output"), smerr.ID, plan.Name.String())
		return
	}

	plan.ARN = flex.StringToFramework(ctx, out.Arn)

	createTimeout := r.CreateTimeout(ctx, plan.Timeouts)
	view, err := waitViewCreated(ctx, conn, plan.ARN.ValueString(), createTimeout)
	if err != nil {
		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, plan.Name.String())
		return
	}
	smerr.AddEnrich(ctx, &resp.Diagnostics, flex.Flatten(ctx, view, &plan, flex.WithNoIgnoredFieldNames()))
	if resp.Diagnostics.HasError() {
		return
	}

	smerr.AddEnrich(ctx, &resp.Diagnostics, resp.State.Set(ctx, plan))
}

func (r *resourceView) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	conn := r.Meta().BillingClient(ctx)

	var state resourceViewModel
	smerr.AddEnrich(ctx, &resp.Diagnostics, req.State.Get(ctx, &state))
	if resp.Diagnostics.HasError() {
		return
	}

	out, err := findViewByARN(ctx, conn, state.ARN.ValueString())

	if retry.NotFound(err) {
		resp.Diagnostics.Append(fwdiag.NewResourceNotFoundWarningDiagnostic(err))
		resp.State.RemoveResource(ctx)
		return
	}
	if err != nil {
		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, state.Name.String())
		return
	}

	smerr.AddEnrich(ctx, &resp.Diagnostics, flex.Flatten(ctx, out, &state, flex.WithNoIgnoredFieldNames()))
	if resp.Diagnostics.HasError() {
		return
	}

	sourceViews, err := findSourceViewsByARN(ctx, conn, state.ARN.ValueString())
	if err != nil {
		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, state.Name.String())
		return
	}
	state.SourceViews = flex.FlattenFrameworkStringValueListOfString(ctx, sourceViews)

	smerr.AddEnrich(ctx, &resp.Diagnostics, resp.State.Set(ctx, &state))
}

func (r *resourceView) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	conn := r.Meta().BillingClient(ctx)

	var plan, state resourceViewModel
	smerr.AddEnrich(ctx, &resp.Diagnostics, req.Plan.Get(ctx, &plan))
	smerr.AddEnrich(ctx, &resp.Diagnostics, req.State.Get(ctx, &state))
	if resp.Diagnostics.HasError() {
		return
	}

	diff, d := flex.Diff(ctx, plan, state)
	smerr.AddEnrich(ctx, &resp.Diagnostics, d)
	if resp.Diagnostics.HasError() {
		return
	}

	if diff.HasChanges() {
		var input billing.UpdateBillingViewInput
		smerr.AddEnrich(ctx, &resp.Diagnostics, flex.Expand(ctx, plan, &input, flex.WithNoIgnoredFieldNames()))
		if resp.Diagnostics.HasError() {
			return
		}

		out, err := conn.UpdateBillingView(ctx, &input)
		if err != nil {
			smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, plan.Name.String())
			return
		}
		if out == nil || out.Arn == nil {
			smerr.AddError(ctx, &resp.Diagnostics, errors.New("empty output"), smerr.ID, plan.Name.String())
			return
		}

		plan.ARN = flex.StringToFramework(ctx, out.Arn)
	}

	updateTimeout := r.UpdateTimeout(ctx, plan.Timeouts)
	view, err := waitViewUpdated(ctx, conn, plan.ARN.ValueString(), updateTimeout)
	if err != nil {
		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, plan.Name.String())
		return
	}
	smerr.AddEnrich(ctx, &resp.Diagnostics, flex.Flatten(ctx, view, &plan, flex.WithNoIgnoredFieldNames()))

	smerr.AddEnrich(ctx, &resp.Diagnostics, resp.State.Set(ctx, &plan))
}

func (r *resourceView) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	conn := r.Meta().BillingClient(ctx)

	var state resourceViewModel
	smerr.AddEnrich(ctx, &resp.Diagnostics, req.State.Get(ctx, &state))
	if resp.Diagnostics.HasError() {
		return
	}

	input := billing.DeleteBillingViewInput{
		Arn: state.ARN.ValueStringPointer(),
	}

	_, err := conn.DeleteBillingView(ctx, &input)
	if err != nil {
		if errs.IsA[*awstypes.ResourceNotFoundException](err) {
			return
		}

		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, state.Name.String())
		return
	}

	deleteTimeout := r.DeleteTimeout(ctx, state.Timeouts)
	_, err = waitViewDeleted(ctx, conn, state.ARN.ValueString(), deleteTimeout)
	if err != nil {
		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, state.Name.String())
		return
	}
}

func (r *resourceView) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root(names.AttrARN), req, resp)
}

func findViewByARN(ctx context.Context, conn *billing.Client, arn string) (*awstypes.BillingViewElement, error) {
	input := billing.GetBillingViewInput{
		Arn: aws.String(arn),
	}

	out, err := conn.GetBillingView(ctx, &input)
	if err != nil {
		if errs.IsA[*awstypes.ResourceNotFoundException](err) {
			return nil, smarterr.NewError(&retry.NotFoundError{
				LastError: err,
			})
		}

		return nil, smarterr.NewError(err)
	}

	if out == nil || out.BillingView.Arn == nil {
		return nil, smarterr.NewError(tfresource.NewEmptyResultError())
	}

	return out.BillingView, nil
}

func findSourceViewsByARN(ctx context.Context, conn *billing.Client, arn string) ([]string, error) {
	sourceViews := make([]string, 0)

	paginator := billing.NewListSourceViewsForBillingViewPaginator(conn, &billing.ListSourceViewsForBillingViewInput{
		Arn: aws.String(arn),
	})

	for paginator.HasMorePages() {
		sourceView, err := paginator.NextPage(ctx)
		if err != nil {
			tflog.Error(ctx, "Listing source views for billing view", map[string]any{
				names.AttrARN: arn,
				"error":       err.Error(),
			})
			return nil, err
		}

		sourceViews = append(sourceViews, sourceView.SourceViews...)
	}

	return sourceViews, nil
}

func statusView(conn *billing.Client, arn string) retry.StateRefreshFunc {
	return func(ctx context.Context) (any, string, error) {
		output, err := findViewByARN(ctx, conn, arn)

		if retry.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, string(output.HealthStatus.StatusCode), nil
	}
}

func waitViewCreated(ctx context.Context, conn *billing.Client, arn string, timeout time.Duration) (*awstypes.BillingViewElement, error) {
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(awstypes.BillingViewStatusCreating),
		Target:  enum.Slice(awstypes.BillingViewStatusHealthy),
		Refresh: statusView(conn, arn),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*awstypes.BillingViewElement); ok {
		return output, err
	}

	return nil, err
}

func waitViewUpdated(ctx context.Context, conn *billing.Client, arn string, timeout time.Duration) (*awstypes.BillingViewElement, error) {
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(awstypes.BillingViewStatusUpdating),
		Target:  enum.Slice(awstypes.BillingViewStatusHealthy),
		Refresh: statusView(conn, arn),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*awstypes.BillingViewElement); ok {
		return output, err
	}

	return nil, err
}

func waitViewDeleted(ctx context.Context, conn *billing.Client, arn string, timeout time.Duration) (*awstypes.BillingViewElement, error) {
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(awstypes.BillingViewStatusHealthy),
		Target:  []string{},
		Refresh: statusView(conn, arn),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*awstypes.BillingViewElement); ok {
		return output, err
	}

	return nil, err
}

type resourceViewModel struct {
	ARN                         types.String                                               `tfsdk:"arn"`
	BillingViewType             fwtypes.StringEnum[awstypes.BillingViewType]               `tfsdk:"billing_view_type"`
	CreatedAt                   timetypes.RFC3339                                          `tfsdk:"created_at"`
	DataFilterExpression        fwtypes.ListNestedObjectValueOf[dataFilterExpressionModel] `tfsdk:"data_filter_expression"`
	DerivedViewCount            types.Int32                                                `tfsdk:"derived_view_count"`
	Description                 types.String                                               `tfsdk:"description"`
	Name                        types.String                                               `tfsdk:"name"`
	OwnerAccountId              types.String                                               `tfsdk:"owner_account_id"`
	Tags                        tftags.Map                                                 `tfsdk:"tags"`
	TagsAll                     tftags.Map                                                 `tfsdk:"tags_all"`
	SourceAccountId             types.String                                               `tfsdk:"source_account_id"`
	SourceViewCount             types.Int32                                                `tfsdk:"source_view_count"`
	SourceViews                 fwtypes.ListOfString                                       `tfsdk:"source_views"`
	Timeouts                    timeouts.Value                                             `tfsdk:"timeouts"`
	UpdatedAt                   timetypes.RFC3339                                          `tfsdk:"updated_at"`
	ViewDefinitionLastUpdatedAt timetypes.RFC3339                                          `tfsdk:"view_definition_last_updated_at"`
}

type dataFilterExpressionModel struct {
	Dimensions fwtypes.ListNestedObjectValueOf[dimensionsModel] `tfsdk:"dimensions"`
	Tags       fwtypes.ListNestedObjectValueOf[tagValuesModel]  `tfsdk:"tags"`
	TimeRange  fwtypes.ListNestedObjectValueOf[timeRangeModel]  `tfsdk:"time_range"`
}

type dimensionsModel struct {
	Key    fwtypes.StringEnum[awstypes.Dimension] `tfsdk:"key"`
	Values fwtypes.SetOfString                    `tfsdk:"values"`
}

type tagValuesModel struct {
	Key    types.String         `tfsdk:"key"`
	Values fwtypes.ListOfString `tfsdk:"values"`
}

type timeRangeModel struct {
	BeginDateInclusive timetypes.RFC3339 `tfsdk:"begin_date_inclusive"`
	EndDateInclusive   timetypes.RFC3339 `tfsdk:"end_date_inclusive"`
}
