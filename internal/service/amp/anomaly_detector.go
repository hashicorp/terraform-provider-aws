// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

// DONOTCOPY: Copying old resources spreads bad habits. Use skaff instead.

package amp

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/YakDriver/smarterr"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/amp"
	awstypes "github.com/aws/aws-sdk-go-v2/service/amp/types"
	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework-timetypes/timetypes"
	"github.com/hashicorp/terraform-plugin-framework-validators/boolvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/float64validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/int32validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int32planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/fwdiag"
	intflex "github.com/hashicorp/terraform-provider-aws/internal/flex"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	fwflex "github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/smerr"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	inttypes "github.com/hashicorp/terraform-provider-aws/internal/types"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkResource("aws_prometheus_anomaly_detector", name="Anomaly Detector")
// @Tags(identifierAttribute="arn")
// @IdentityAttribute("id")
// @IdentityAttribute("workspace_id")
// @ImportIDHandler("anomalyDetectorImportID")
// @Testing(hasNoPreExistingResource=true)
// @Testing(importStateIdFunc="testAccAnomalyDetectorImportState")
func newAnomalyDetectorResource(_ context.Context) (resource.ResourceWithConfigure, error) {
	r := &anomalyDetectorResource{}

	r.SetDefaultCreateTimeout(10 * time.Minute)
	r.SetDefaultUpdateTimeout(10 * time.Minute)
	r.SetDefaultDeleteTimeout(10 * time.Minute)

	return r, nil
}

const (
	ResNameAnomalyDetector = "Anomaly Detector"
)

type anomalyDetectorResource struct {
	framework.ResourceWithModel[anomalyDetectorResourceModel]
	framework.WithTimeouts
	framework.WithImportByIdentity
}

func (r *anomalyDetectorResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			names.AttrAlias: schema.StringAttribute{
				Required: true,
			},
			names.AttrARN: framework.ARNAttributeComputedOnly(),
			names.AttrCreatedAt: schema.StringAttribute{
				CustomType: timetypes.RFC3339Type{},
				Computed:   true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"evaluation_interval_in_seconds": schema.Int32Attribute{
				Optional: true,
				Computed: true,
				PlanModifiers: []planmodifier.Int32{
					int32planmodifier.UseStateForUnknown(),
				},
			},
			names.AttrID: framework.IDAttribute(),
			"labels": schema.MapAttribute{
				CustomType:  fwtypes.MapOfStringType,
				ElementType: types.StringType,
				Optional:    true,
				Computed:    true,
			},
			names.AttrTags:    tftags.TagsAttribute(),
			names.AttrTagsAll: tftags.TagsAttributeComputedOnly(),
			"workspace_id": schema.StringAttribute{
				Required: true,
			},
		},
		Blocks: map[string]schema.Block{
			names.AttrConfiguration: schema.ListNestedBlock{
				CustomType: fwtypes.NewListNestedObjectTypeOf[anomalyDetectorConfigurationModel](ctx),
				Validators: []validator.List{
					listvalidator.IsRequired(),
					listvalidator.SizeAtLeast(1),
					listvalidator.SizeAtMost(1),
				},
				NestedObject: schema.NestedBlockObject{
					Blocks: map[string]schema.Block{
						"random_cut_forest": schema.ListNestedBlock{
							CustomType: fwtypes.NewListNestedObjectTypeOf[randomCutForestConfigurationModel](ctx),
							Validators: []validator.List{
								listvalidator.IsRequired(),
								listvalidator.SizeAtLeast(1),
								listvalidator.SizeAtMost(1),
							},
							NestedObject: schema.NestedBlockObject{
								Attributes: map[string]schema.Attribute{
									"query": schema.StringAttribute{
										Required: true,
									},
									"sample_size": schema.Int32Attribute{
										Optional: true,
										Computed: true,
										Validators: []validator.Int32{
											int32validator.AtLeast(256),
										},
									},
									"shingle_size": schema.Int32Attribute{
										Optional: true,
										Computed: true,
										Validators: []validator.Int32{
											int32validator.AtLeast(2),
										},
									},
								},
								Blocks: map[string]schema.Block{
									"ignore_near_expected_from_above": schema.ListNestedBlock{
										CustomType: fwtypes.NewListNestedObjectTypeOf[ignoreNearExpectedModel](ctx),
										Validators: []validator.List{
											listvalidator.SizeAtMost(1),
										},
										NestedObject: schema.NestedBlockObject{
											Attributes: map[string]schema.Attribute{
												"amount": schema.Float64Attribute{
													Optional: true,
													Validators: []validator.Float64{
														float64validator.ExactlyOneOf(
															path.MatchRelative().AtParent().AtName("ratio"),
														),
														float64validator.AtLeast(0),
													},
												},
												"ratio": schema.Float64Attribute{
													Optional: true,
													Validators: []validator.Float64{
														float64validator.ExactlyOneOf(
															path.MatchRelative().AtParent().AtName("amount"),
														),
														float64validator.AtLeast(0),
													},
												},
											},
										},
									},
									"ignore_near_expected_from_below": schema.ListNestedBlock{
										CustomType: fwtypes.NewListNestedObjectTypeOf[ignoreNearExpectedModel](ctx),
										Validators: []validator.List{
											listvalidator.SizeAtMost(1),
										},
										NestedObject: schema.NestedBlockObject{
											Attributes: map[string]schema.Attribute{
												"amount": schema.Float64Attribute{
													Optional: true,
													Validators: []validator.Float64{
														float64validator.ExactlyOneOf(
															path.MatchRelative().AtParent().AtName("ratio"),
														),
														float64validator.AtLeast(0),
													},
												},
												"ratio": schema.Float64Attribute{
													Optional: true,
													Validators: []validator.Float64{
														float64validator.ExactlyOneOf(
															path.MatchRelative().AtParent().AtName("amount"),
														),
														float64validator.AtLeast(0),
													},
												},
											},
										},
									},
								},
							},
						},
					},
				},
			},
			"missing_data_action": schema.ListNestedBlock{
				CustomType: fwtypes.NewListNestedObjectTypeOf[anomalyDetectorMissingDataActionModel](ctx),
				Validators: []validator.List{
					listvalidator.IsRequired(),
					listvalidator.SizeAtMost(1),
				},
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"mark_as_anomaly": schema.BoolAttribute{
							Optional: true,
							Validators: []validator.Bool{
								boolvalidator.ExactlyOneOf(
									path.MatchRelative().AtParent().AtName("skip"),
								),
								boolvalidator.Equals(true),
							},
						},
						"skip": schema.BoolAttribute{
							Optional: true,
							Validators: []validator.Bool{
								boolvalidator.ExactlyOneOf(
									path.MatchRelative().AtParent().AtName("mark_as_anomaly"),
								),
								boolvalidator.Equals(true),
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

func (r *anomalyDetectorResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	conn := r.Meta().AMPClient(ctx)

	var plan anomalyDetectorResourceModel
	smerr.AddEnrich(ctx, &resp.Diagnostics, req.Plan.Get(ctx, &plan))
	if resp.Diagnostics.HasError() {
		return
	}

	var input amp.CreateAnomalyDetectorInput
	smerr.AddEnrich(ctx, &resp.Diagnostics, fwflex.Expand(ctx, plan, &input, fwflex.WithFieldNamePrefix("AnomalyDetector")))
	if resp.Diagnostics.HasError() {
		return
	}

	// Additional fields not covered by AutoFlex
	input.ClientToken = aws.String(create.UniqueId(ctx))
	input.Tags = getTagsIn(ctx)

	out, err := conn.CreateAnomalyDetector(ctx, &input)
	if err != nil {
		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, plan.Alias.String())
		return
	}
	if out == nil {
		smerr.AddError(ctx, &resp.Diagnostics, errors.New("empty output"), smerr.ID, plan.Alias.String())
		return
	}

	// Set computed values
	plan.ID = fwflex.StringToFramework(ctx, out.AnomalyDetectorId)
	plan.ARN = fwflex.StringToFramework(ctx, out.Arn)

	detector, err := waitAnomalyDetectorCreated(ctx, conn, plan.ID.ValueString(), plan.WorkspaceID.ValueString(), r.CreateTimeout(ctx, plan.Timeouts))
	if err != nil {
		resp.State.SetAttribute(ctx, path.Root(names.AttrID), plan.ID)
		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, plan.ID.ValueString())
		return
	}

	smerr.AddEnrich(ctx, &resp.Diagnostics, fwflex.Flatten(ctx, detector, &plan, fwflex.WithFieldNamePrefix("AnomalyDetector")))
	if resp.Diagnostics.HasError() {
		return
	}

	smerr.AddEnrich(ctx, &resp.Diagnostics, resp.State.Set(ctx, plan))
}

func (r *anomalyDetectorResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	conn := r.Meta().AMPClient(ctx)

	var state anomalyDetectorResourceModel
	smerr.AddEnrich(ctx, &resp.Diagnostics, req.State.Get(ctx, &state))
	if resp.Diagnostics.HasError() {
		return
	}

	out, err := findAnomalyDetectorByID(ctx, conn, state.ID.ValueString(), state.WorkspaceID.ValueString())
	if retry.NotFound(err) {
		resp.Diagnostics.Append(fwdiag.NewResourceNotFoundWarningDiagnostic(err))
		resp.State.RemoveResource(ctx)
		return
	}
	if err != nil {
		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, state.ID.String())
		return
	}

	smerr.AddEnrich(ctx, &resp.Diagnostics, fwflex.Flatten(ctx, out, &state, fwflex.WithFieldNamePrefix("AnomalyDetector")))
	if resp.Diagnostics.HasError() {
		return
	}

	setTagsOut(ctx, out.Tags)
	smerr.AddEnrich(ctx, &resp.Diagnostics, resp.State.Set(ctx, &state))
}

func (r *anomalyDetectorResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	conn := r.Meta().AMPClient(ctx)

	// plan = new, state = old
	var plan, state anomalyDetectorResourceModel
	smerr.AddEnrich(ctx, &resp.Diagnostics, req.Plan.Get(ctx, &plan))
	smerr.AddEnrich(ctx, &resp.Diagnostics, req.State.Get(ctx, &state))
	if resp.Diagnostics.HasError() {
		return
	}

	diff, d := fwflex.Diff(ctx, plan, state)
	smerr.AddEnrich(ctx, &resp.Diagnostics, d)
	if resp.Diagnostics.HasError() {
		return
	}

	if diff.HasChanges() {
		var input amp.PutAnomalyDetectorInput
		smerr.AddEnrich(ctx, &resp.Diagnostics, fwflex.Expand(ctx, plan, &input, fwflex.WithFieldNamePrefix("AnomalyDetector")))
		if resp.Diagnostics.HasError() {
			return
		}

		input.ClientToken = aws.String(create.UniqueId(ctx))

		_, err := conn.PutAnomalyDetector(ctx, &input)
		if err != nil {
			smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, plan.ID.ValueString())
			return
		}

		updated, err := waitAnomalyDetectorUpdated(ctx, conn, plan.ID.ValueString(), plan.WorkspaceID.ValueString(), r.UpdateTimeout(ctx, plan.Timeouts))
		if err != nil {
			smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, plan.ID.ValueString())
			return
		}

		// Re-setting computed values
		smerr.AddEnrich(ctx, &resp.Diagnostics, fwflex.Flatten(ctx, updated, &plan, fwflex.WithFieldNamePrefix("AnomalyDetector")))
		if resp.Diagnostics.HasError() {
			return
		}
	}

	smerr.AddEnrich(ctx, &resp.Diagnostics, resp.State.Set(ctx, &plan))
}

func (r *anomalyDetectorResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	conn := r.Meta().AMPClient(ctx)

	var state anomalyDetectorResourceModel
	smerr.AddEnrich(ctx, &resp.Diagnostics, req.State.Get(ctx, &state))
	if resp.Diagnostics.HasError() {
		return
	}

	input := amp.DeleteAnomalyDetectorInput{
		AnomalyDetectorId: state.ID.ValueStringPointer(),
		WorkspaceId:       state.WorkspaceID.ValueStringPointer(),
		ClientToken:       aws.String(create.UniqueId(ctx)),
	}

	_, err := conn.DeleteAnomalyDetector(ctx, &input)
	if err != nil {
		if errs.IsA[*awstypes.ResourceNotFoundException](err) {
			return
		}

		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, state.ID.ValueString())
		return
	}

	_, err = waitAnomalyDetectorDeleted(ctx, conn, state.ID.ValueString(), state.WorkspaceID.ValueString(), r.DeleteTimeout(ctx, state.Timeouts))
	if err != nil {
		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, state.ID.ValueString())
		return
	}
}

func waitAnomalyDetectorCreated(ctx context.Context, conn *amp.Client, id, workspaceID string, timeout time.Duration) (*awstypes.AnomalyDetectorDescription, error) {
	stateConf := &retry.StateChangeConf{
		Pending:                   enum.Slice(awstypes.AnomalyDetectorStatusCodeCreating),
		Target:                    enum.Slice(awstypes.AnomalyDetectorStatusCodeActive),
		Refresh:                   statusAnomalyDetector(conn, id, workspaceID),
		Timeout:                   timeout,
		ContinuousTargetOccurence: 2,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)
	if out, ok := outputRaw.(*awstypes.AnomalyDetectorDescription); ok {
		if status := out.Status; status != nil && status.StatusCode == awstypes.AnomalyDetectorStatusCodeCreationFailed {
			retry.SetLastError(err, errors.New(aws.ToString(status.StatusReason)))
		}
		return out, smarterr.NewError(err)
	}

	return nil, smarterr.NewError(err)
}

func waitAnomalyDetectorUpdated(ctx context.Context, conn *amp.Client, id, workspaceID string, timeout time.Duration) (*awstypes.AnomalyDetectorDescription, error) {
	stateConf := &retry.StateChangeConf{
		Pending:                   enum.Slice(awstypes.AnomalyDetectorStatusCodeUpdating),
		Target:                    enum.Slice(awstypes.AnomalyDetectorStatusCodeActive),
		Refresh:                   statusAnomalyDetector(conn, id, workspaceID),
		Timeout:                   timeout,
		ContinuousTargetOccurence: 2,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)
	if out, ok := outputRaw.(*awstypes.AnomalyDetectorDescription); ok {
		if status := out.Status; status != nil && status.StatusCode == awstypes.AnomalyDetectorStatusCodeUpdateFailed {
			retry.SetLastError(err, errors.New(aws.ToString(status.StatusReason)))
		}
		return out, smarterr.NewError(err)
	}

	return nil, smarterr.NewError(err)
}

func waitAnomalyDetectorDeleted(ctx context.Context, conn *amp.Client, id, workspaceID string, timeout time.Duration) (*awstypes.AnomalyDetectorDescription, error) {
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(awstypes.AnomalyDetectorStatusCodeDeleting, awstypes.AnomalyDetectorStatusCodeActive),
		Target:  []string{},
		Refresh: statusAnomalyDetector(conn, id, workspaceID),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)
	if out, ok := outputRaw.(*awstypes.AnomalyDetectorDescription); ok {
		if status := out.Status; status != nil && status.StatusCode == awstypes.AnomalyDetectorStatusCodeDeletionFailed {
			retry.SetLastError(err, errors.New(aws.ToString(status.StatusReason)))
		}
		return out, smarterr.NewError(err)
	}

	return nil, smarterr.NewError(err)
}

func statusAnomalyDetector(conn *amp.Client, id string, workspaceID string) retry.StateRefreshFunc {
	return func(ctx context.Context) (any, string, error) {
		out, err := findAnomalyDetectorByID(ctx, conn, id, workspaceID)
		if retry.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", smarterr.NewError(err)
		}

		var statusCode string
		if out != nil && out.Status != nil {
			statusCode = string(out.Status.StatusCode)
		}

		return out, statusCode, nil
	}
}

func findAnomalyDetectorByID(ctx context.Context, conn *amp.Client, id, workspaceID string) (*awstypes.AnomalyDetectorDescription, error) {
	input := amp.DescribeAnomalyDetectorInput{
		AnomalyDetectorId: aws.String(id),
		WorkspaceId:       aws.String(workspaceID),
	}

	out, err := conn.DescribeAnomalyDetector(ctx, &input)
	if err != nil {
		if errs.IsA[*awstypes.ResourceNotFoundException](err) {
			return nil, smarterr.NewError(&retry.NotFoundError{
				LastError: err,
			})
		}

		return nil, smarterr.NewError(err)
	}

	if out == nil || out.AnomalyDetector == nil {
		return nil, smarterr.NewError(tfresource.NewEmptyResultError())
	}

	return out.AnomalyDetector, nil
}

type anomalyDetectorResourceModel struct {
	framework.WithRegionModel
	Alias                       types.String                                                           `tfsdk:"alias"`
	ARN                         types.String                                                           `tfsdk:"arn"`
	Configuration               fwtypes.ListNestedObjectValueOf[anomalyDetectorConfigurationModel]     `tfsdk:"configuration"`
	CreatedAt                   timetypes.RFC3339                                                      `tfsdk:"created_at"`
	EvaluationIntervalInSeconds types.Int32                                                            `tfsdk:"evaluation_interval_in_seconds"`
	ID                          types.String                                                           `tfsdk:"id"`
	Labels                      fwtypes.MapOfString                                                    `tfsdk:"labels"`
	MissingDataAction           fwtypes.ListNestedObjectValueOf[anomalyDetectorMissingDataActionModel] `tfsdk:"missing_data_action"`
	Tags                        tftags.Map                                                             `tfsdk:"tags"`
	TagsAll                     tftags.Map                                                             `tfsdk:"tags_all"`
	Timeouts                    timeouts.Value                                                         `tfsdk:"timeouts"`
	WorkspaceID                 types.String                                                           `tfsdk:"workspace_id"`
}

var (
	_ fwflex.Expander  = anomalyDetectorConfigurationModel{}
	_ fwflex.Flattener = &anomalyDetectorConfigurationModel{}
)

type anomalyDetectorConfigurationModel struct {
	RandomCutForest fwtypes.ListNestedObjectValueOf[randomCutForestConfigurationModel] `tfsdk:"random_cut_forest"`
}

func (m anomalyDetectorConfigurationModel) Expand(ctx context.Context) (result any, diags diag.Diagnostics) {
	switch {
	case !m.RandomCutForest.IsNull():
		data, d := m.RandomCutForest.ToPtr(ctx)
		diags.Append(d...)
		if diags.HasError() {
			return nil, diags
		}

		var r awstypes.AnomalyDetectorConfigurationMemberRandomCutForest
		diags.Append(fwflex.Expand(ctx, data, &r.Value)...)
		if diags.HasError() {
			return nil, diags
		}
		return &r, diags
	}
	return nil, diags
}

func (m *anomalyDetectorConfigurationModel) Flatten(ctx context.Context, v any) (diags diag.Diagnostics) {
	switch t := v.(type) {
	case awstypes.AnomalyDetectorConfigurationMemberRandomCutForest:
		var data randomCutForestConfigurationModel
		diags.Append(fwflex.Flatten(ctx, t.Value, &data)...)
		if diags.HasError() {
			return diags
		}

		m.RandomCutForest = fwtypes.NewListNestedObjectValueOfPtrMust(ctx, &data)
	}
	return diags
}

type randomCutForestConfigurationModel struct {
	Query                       types.String                                             `tfsdk:"query"`
	IgnoreNearExpectedFromAbove fwtypes.ListNestedObjectValueOf[ignoreNearExpectedModel] `tfsdk:"ignore_near_expected_from_above"`
	IgnoreNearExpectedFromBelow fwtypes.ListNestedObjectValueOf[ignoreNearExpectedModel] `tfsdk:"ignore_near_expected_from_below"`
	SampleSize                  types.Int32                                              `tfsdk:"sample_size"`
	ShingleSize                 types.Int32                                              `tfsdk:"shingle_size"`
}

var (
	_ fwflex.Expander  = ignoreNearExpectedModel{}
	_ fwflex.Flattener = &ignoreNearExpectedModel{}
)

type ignoreNearExpectedModel struct {
	Amount types.Float64 `tfsdk:"amount"`
	Ratio  types.Float64 `tfsdk:"ratio"`
}

func (m ignoreNearExpectedModel) Expand(ctx context.Context) (result any, diags diag.Diagnostics) {
	switch {
	case !m.Amount.IsNull():
		return &awstypes.IgnoreNearExpectedMemberAmount{
			Value: m.Amount.ValueFloat64(),
		}, diags
	case !m.Ratio.IsNull():
		return &awstypes.IgnoreNearExpectedMemberRatio{
			Value: m.Ratio.ValueFloat64(),
		}, diags
	}
	return nil, diags
}

func (m *ignoreNearExpectedModel) Flatten(ctx context.Context, v any) (diags diag.Diagnostics) {
	switch t := v.(type) {
	case awstypes.IgnoreNearExpectedMemberAmount:
		m.Amount = types.Float64Value(t.Value)
	case awstypes.IgnoreNearExpectedMemberRatio:
		m.Ratio = types.Float64Value(t.Value)
	}
	return diags
}

var (
	_ fwflex.Expander  = anomalyDetectorMissingDataActionModel{}
	_ fwflex.Flattener = &anomalyDetectorMissingDataActionModel{}
)

type anomalyDetectorMissingDataActionModel struct {
	MarkAsAnomaly types.Bool `tfsdk:"mark_as_anomaly"`
	Skip          types.Bool `tfsdk:"skip"`
}

func (m anomalyDetectorMissingDataActionModel) Expand(ctx context.Context) (result any, diags diag.Diagnostics) {
	switch {
	case !m.MarkAsAnomaly.IsNull():
		return &awstypes.AnomalyDetectorMissingDataActionMemberMarkAsAnomaly{
			Value: m.MarkAsAnomaly.ValueBool(),
		}, diags
	case !m.Skip.IsNull():
		return &awstypes.AnomalyDetectorMissingDataActionMemberSkip{
			Value: m.Skip.ValueBool(),
		}, diags
	}
	return nil, diags
}

func (m *anomalyDetectorMissingDataActionModel) Flatten(ctx context.Context, v any) (diags diag.Diagnostics) {
	switch t := v.(type) {
	case awstypes.AnomalyDetectorMissingDataActionMemberMarkAsAnomaly:
		m.MarkAsAnomaly = fwflex.BoolValueToFramework(ctx, t.Value)
	case awstypes.AnomalyDetectorMissingDataActionMemberSkip:
		m.Skip = fwflex.BoolValueToFramework(ctx, t.Value)
	}
	return diags
}

var (
	_ inttypes.ImportIDParser = anomalyDetectorImportID{}
)

type anomalyDetectorImportID struct{}

func (anomalyDetectorImportID) Parse(id string) (string, map[string]any, error) {
	anomalyDetectorID, workspaceID, found := strings.Cut(id, intflex.ResourceIdSeparator)
	if !found {
		return "", nil, fmt.Errorf("id \"%s\" should be in the format <anomaly-detector-id>"+intflex.ResourceIdSeparator+"<workspace-id>", id)
	}

	result := map[string]any{
		names.AttrID:   anomalyDetectorID,
		"workspace_id": workspaceID,
	}

	return id, result, nil
}
