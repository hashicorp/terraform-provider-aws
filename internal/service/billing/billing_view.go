// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package billing

import (
	"context"
	"errors"
	"time"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/billing"
	awstypes "github.com/aws/aws-sdk-go-v2/service/billing/types"
	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework-timetypes/timetypes"
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	"github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/framework/validators"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// Function annotations are used for resource registration to the Provider. DO NOT EDIT.
// @FrameworkResource("aws_billing_view", name="View")
func newResourceView(_ context.Context) (resource.ResourceWithConfigure, error) {
	r := &resourceView{}

	r.SetDefaultCreateTimeout(30 * time.Minute)
	r.SetDefaultUpdateTimeout(30 * time.Minute)
	r.SetDefaultDeleteTimeout(30 * time.Minute)

	return r, nil
}

const (
	ResNameView = "View"
)

type resourceView struct {
	framework.ResourceWithConfigure
	framework.WithTimeouts
}

func (r *resourceView) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = "aws_billing_view"
}

func (r *resourceView) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			names.AttrARN: framework.ARNAttributeComputedOnly(),
			"client_token": schema.StringAttribute{
				Optional: true,
			},
			names.AttrCreatedAt: schema.StringAttribute{
				CustomType: timetypes.RFC3339Type{},
				Computed:   true,
			},
			names.AttrDescription: schema.StringAttribute{
				Optional: true,
				Validators: []validator.String{
					stringvalidator.LengthBetween(0, 1024),
				},
			},
			names.AttrName: schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
				Validators: []validator.String{
					stringvalidator.LengthBetween(1, 128),
					stringvalidator.RegexMatches(regexache.MustCompile(`[ a-zA-Z0-9_\+=\.\-@]+`), "must contain only alphanumeric characters, spaces, and the following special characters: _+=.-@"),
				},
			},
			"source_views": schema.ListAttribute{
				Required:    true,
				CustomType:  fwtypes.ListOfStringType,
				ElementType: types.StringType,
				Validators: []validator.List{
					listvalidator.ValueStringsAre(validators.ARN()),
				},
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
							CustomType: fwtypes.NewListNestedObjectTypeOf[dimensionModel](ctx),
							Validators: []validator.List{
								listvalidator.SizeAtMost(1),
							},
							NestedObject: schema.NestedBlockObject{
								Attributes: map[string]schema.Attribute{
									names.AttrKey: schema.StringAttribute{
										Required:   true,
										CustomType: fwtypes.StringEnumType[awstypes.Dimension](),
									},
									names.AttrValues: schema.ListAttribute{
										Required:    true,
										CustomType:  fwtypes.ListOfStringType,
										ElementType: types.StringType,
										Validators: []validator.List{
											listvalidator.SizeBetween(1, 200),
											listvalidator.ValueStringsAre(
												stringvalidator.RegexMatches(regexache.MustCompile(`[\S\s]*`), "must contain any character"),
												stringvalidator.LengthBetween(0, 1024),
											),
										},
									},
								},
							},
						},
						names.AttrTags: schema.ListNestedBlock{
							CustomType: fwtypes.NewListNestedObjectTypeOf[tagsModel](ctx),
							Validators: []validator.List{
								listvalidator.SizeAtMost(1),
							},
							NestedObject: schema.NestedBlockObject{
								Attributes: map[string]schema.Attribute{
									names.AttrKey: schema.StringAttribute{
										Required: true,
										Validators: []validator.String{
											stringvalidator.LengthBetween(0, 1024),
											stringvalidator.RegexMatches(regexache.MustCompile(`[\S\s]*`), "must contain any character"),
										},
									},
									names.AttrValues: schema.ListAttribute{
										Required:    true,
										CustomType:  fwtypes.ListOfStringType,
										ElementType: types.StringType,
										Validators: []validator.List{
											listvalidator.SizeBetween(1, 200),
											listvalidator.ValueStringsAre(
												stringvalidator.RegexMatches(regexache.MustCompile(`[\S\s]*`), "must contain any character"),
												stringvalidator.LengthBetween(0, 1024),
											),
										},
									},
								},
							},
						},
					},
				},
			},
			names.AttrResourceTags: schema.ListNestedBlock{
				CustomType: fwtypes.NewListNestedObjectTypeOf[tagsModel](ctx),
				Validators: []validator.List{
					listvalidator.SizeBetween(0, 200),
				},
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						names.AttrKey: schema.StringAttribute{
							Required: true,
							Validators: []validator.String{
								stringvalidator.LengthBetween(1, 128),
							},
						},
						names.AttrValue: schema.StringAttribute{
							Required: true,
							Validators: []validator.String{
								stringvalidator.LengthBetween(1, 256),
							},
						},
					},
				},
			},
			names.AttrTimeouts: timeouts.Block(ctx, timeouts.Opts{
				Create: true,
				Delete: true,
			}),
		},
	}
}

func (r *resourceView) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	conn := r.Meta().BillingClient(ctx)

	var plan resourceViewModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	input := new(billing.CreateBillingViewInput)

	resp.Diagnostics.Append(flex.Expand(ctx, plan, input)...)
	if resp.Diagnostics.HasError() {
		return
	}

	input.ResourceTags = getTagsIn(ctx)

	out, err := conn.CreateBillingView(ctx, input)
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.Billing, create.ErrActionCreating, ResNameView, plan.Name.String(), err),
			err.Error(),
		)
		return
	}

	if out == nil || out.Arn == nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.Billing, create.ErrActionCreating, ResNameView, plan.Name.String(), nil),
			errors.New("empty output").Error(),
		)
		return
	}

	resp.Diagnostics.Append(flex.Flatten(ctx, out, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	createTimeout := r.CreateTimeout(ctx, plan.Timeouts)
	_, err = waitViewCreated(ctx, conn, plan.ARN.ValueString(), createTimeout)
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.Billing, create.ErrActionWaitingForCreation, ResNameView, plan.Name.String(), err),
			err.Error(),
		)
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *resourceView) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	conn := r.Meta().BillingClient(ctx)

	var state resourceViewModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	out, err := findViewByARN(ctx, conn, state.ARN.ValueString())
	if tfresource.NotFound(err) {
		resp.State.RemoveResource(ctx)
		return
	}
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.Billing, create.ErrActionSetting, ResNameView, state.ARN.String(), err),
			err.Error(),
		)
		return
	}

	resp.Diagnostics.Append(flex.Flatten(ctx, out, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	sourceViews, err := findSourceViewsByARN(ctx, conn, state.ARN.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.Billing, create.ErrActionSetting, ResNameView, state.ARN.String(), err),
			err.Error(),
		)
		return
	}

	state.SourceViews = flex.FlattenFrameworkStringValueListOfString(ctx, sourceViews)

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *resourceView) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	conn := r.Meta().BillingClient(ctx)

	var plan, state resourceViewModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if !plan.Name.Equal(state.Name) ||
		!plan.Description.Equal(state.Description) ||
		!plan.DataFilterExpression.Equal(state.DataFilterExpression) {
		input := new(billing.UpdateBillingViewInput)
		resp.Diagnostics.Append(flex.Expand(ctx, plan, &input)...)
		if resp.Diagnostics.HasError() {
			return
		}

		out, err := conn.UpdateBillingView(ctx, input)
		if err != nil {
			resp.Diagnostics.AddError(
				create.ProblemStandardMessage(names.Billing, create.ErrActionUpdating, ResNameView, state.ARN.String(), err),
				err.Error(),
			)
			return
		}
		if out == nil || out.Arn == nil {
			resp.Diagnostics.AddError(
				create.ProblemStandardMessage(names.Billing, create.ErrActionUpdating, ResNameView, state.ARN.String(), nil),
				errors.New("empty output").Error(),
			)
			return
		}

		resp.Diagnostics.Append(flex.Flatten(ctx, out, &plan)...)
		if resp.Diagnostics.HasError() {
			return
		}
	}

	updateTimeout := r.UpdateTimeout(ctx, plan.Timeouts)
	_, err := waitViewUpdated(ctx, conn, plan.ARN.ValueString(), updateTimeout)
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.Billing, create.ErrActionWaitingForUpdate, ResNameView, plan.ARN.String(), err),
			err.Error(),
		)
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *resourceView) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	conn := r.Meta().BillingClient(ctx)

	var state resourceViewModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
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

		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.Billing, create.ErrActionDeleting, ResNameView, state.ARN.String(), err),
			err.Error(),
		)
		return
	}

	deleteTimeout := r.DeleteTimeout(ctx, state.Timeouts)
	_, err = waitViewDeleted(ctx, conn, state.ARN.ValueString(), deleteTimeout)
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.Billing, create.ErrActionWaitingForDeletion, ResNameView, state.ARN.String(), err),
			err.Error(),
		)
		return
	}
}

func (r *resourceView) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root(names.AttrARN), req, resp)
}

const (
	statusChangePending = "Pending"
	statusDeleting      = "Deleting"
	statusNormal        = "Normal"
	statusUpdated       = "Updated"
)

func waitViewCreated(ctx context.Context, conn *billing.Client, id string, timeout time.Duration) (*awstypes.BillingViewElement, error) {
	stateConf := &retry.StateChangeConf{
		Pending:                   []string{},
		Target:                    []string{statusNormal},
		Refresh:                   statusView(ctx, conn, id),
		Timeout:                   timeout,
		NotFoundChecks:            20,
		ContinuousTargetOccurence: 2,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)
	if out, ok := outputRaw.(*awstypes.BillingViewElement); ok {
		return out, err
	}

	return nil, err
}

func waitViewUpdated(ctx context.Context, conn *billing.Client, id string, timeout time.Duration) (*awstypes.BillingViewElement, error) {
	stateConf := &retry.StateChangeConf{
		Pending:                   []string{statusChangePending},
		Target:                    []string{statusUpdated},
		Refresh:                   statusView(ctx, conn, id),
		Timeout:                   timeout,
		NotFoundChecks:            20,
		ContinuousTargetOccurence: 2,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)
	if out, ok := outputRaw.(*awstypes.BillingViewElement); ok {
		return out, err
	}

	return nil, err
}

func waitViewDeleted(ctx context.Context, conn *billing.Client, id string, timeout time.Duration) (*awstypes.BillingViewElement, error) {
	stateConf := &retry.StateChangeConf{
		Pending: []string{statusDeleting, statusNormal},
		Target:  []string{},
		Refresh: statusView(ctx, conn, id),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)
	if out, ok := outputRaw.(*awstypes.BillingViewElement); ok {
		return out, err
	}

	return nil, err
}

func statusView(ctx context.Context, conn *billing.Client, id string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		out, err := findViewByARN(ctx, conn, id)
		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return out, statusNormal, nil
	}
}

func findViewByARN(ctx context.Context, conn *billing.Client, arn string) (*awstypes.BillingViewElement, error) {
	in := new(billing.GetBillingViewInput)
	in.Arn = aws.String(arn)

	out, err := conn.GetBillingView(ctx, in)
	if err != nil {
		if errs.IsA[*awstypes.ResourceNotFoundException](err) {
			return nil, &retry.NotFoundError{
				LastError:   err,
				LastRequest: in,
			}
		}

		return nil, err
	}

	if out == nil || out.BillingView == nil {
		return nil, tfresource.NewEmptyResultError(in)
	}

	return out.BillingView, nil
}

func findSourceViewsByARN(ctx context.Context, conn *billing.Client, arn string) ([]string, error) {
	sourceViews := make([]string, 0)

	sourceViewsPag := billing.NewListSourceViewsForBillingViewPaginator(conn, &billing.ListSourceViewsForBillingViewInput{
		Arn: aws.String(arn),
	})

	for sourceViewsPag.HasMorePages() {
		sourceView, err := sourceViewsPag.NextPage(ctx)
		if err != nil {
			tflog.Error(ctx, "Error listing source views for billing view", map[string]interface{}{
				names.AttrARN: arn,
				"error":       err.Error(),
			})
			return nil, err
		}

		sourceViews = append(sourceViews, sourceView.SourceViews...)
	}

	return sourceViews, nil
}

type resourceViewModel struct {
	ARN                  types.String                                               `tfsdk:"arn"`
	DataFilterExpression fwtypes.ListNestedObjectValueOf[dataFilterExpressionModel] `tfsdk:"data_filter_expression"`
	Description          types.String                                               `tfsdk:"description"`
	Name                 types.String                                               `tfsdk:"name"`
	ResourceTags         fwtypes.ListNestedObjectValueOf[tagsModel]                 `tfsdk:"resource_tags"`
	ClientToken          types.String                                               `tfsdk:"client_token"`
	SourceViews          fwtypes.ListValueOf[types.String]                          `tfsdk:"source_views"`
	CreatedAt            timetypes.RFC3339                                          `tfsdk:"created_at"`
	Timeouts             timeouts.Value                                             `tfsdk:"timeouts"`
}

type dataFilterExpressionModel struct {
	Dimensions fwtypes.ListNestedObjectValueOf[dimensionModel] `tfsdk:"dimensions"`
	Tags       fwtypes.ListNestedObjectValueOf[tagsModel]      `tfsdk:"tags"`
}

type dimensionModel struct {
	Key    fwtypes.StringEnum[awstypes.Dimension] `tfsdk:"key"`
	Values fwtypes.ListValueOf[types.String]      `tfsdk:"values"`
}

type tagsModel struct {
	Key    types.String                      `tfsdk:"key"`
	Values fwtypes.ListValueOf[types.String] `tfsdk:"values"`
}
