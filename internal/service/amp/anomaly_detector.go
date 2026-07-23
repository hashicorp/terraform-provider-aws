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
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/fwdiag"
	intflex "github.com/hashicorp/terraform-provider-aws/internal/flex"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	fwflex "github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/smerr"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep"
	sweepfw "github.com/hashicorp/terraform-provider-aws/internal/sweep/framework"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	inttypes "github.com/hashicorp/terraform-provider-aws/internal/types"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// TIP: ==== FILE STRUCTURE ====
// All resources should follow this basic outline. Improve this resource's
// maintainability by sticking to it.
//
// 1. Package declaration
// 2. Imports
// 3. Main resource struct with schema method
// 4. Create, read, update, delete methods (in that order)
// 5. Other functions (flatteners, expanders, waiters, finders, etc.)

// Function annotations are used for resource registration to the Provider. DO NOT EDIT.
// @FrameworkResource("aws_prometheus_anomaly_detector", name="Anomaly Detector")
// @Tags(identifierAttribute="arn")
// @IdentityAttribute("id")
// @IdentityAttribute("workspace_id")
// @ImportIDHandler("anomalyDetectorImportID")
//
// TIP: ==== GENERATED ACCEPTANCE TESTS ====
// Resource Identity and tagging make use of automatically generated acceptance tests.
// For more information about automatically generated acceptance tests, see
// https://hashicorp.github.io/terraform-provider-aws/acc-test-generation/
//
// Some common annotations are included below:
// // @Testing(existsType="github.com/aws/aws-sdk-go-v2/service/amp;amp.DescribeAnomalyDetectorResponse")
// @Testing(preCheck="testAccPreCheckAnomalyDetector")
// // @Testing(importIgnore="...;...")
// @Testing(hasNoPreExistingResource=true)
// @Testing(importStateIdFunc="testAccAnomalyDetectorImportState")
func newAnomalyDetectorResource(_ context.Context) (resource.ResourceWithConfigure, error) {
	r := &anomalyDetectorResource{}

	// TIP: ==== CONFIGURABLE TIMEOUTS ====
	// Users can configure timeout lengths but you need to use the times they
	// provide. Access the timeout they configure (or the defaults) using,
	// e.g., r.CreateTimeout(ctx, plan.Timeouts) (see below). The times here are
	// the defaults if they don't configure timeouts.
	r.SetDefaultCreateTimeout(30 * time.Minute)
	r.SetDefaultUpdateTimeout(30 * time.Minute)
	r.SetDefaultDeleteTimeout(30 * time.Minute)

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
			},
			"evaluation_interval_in_seconds": schema.Int32Attribute{
				Optional: true,
				Computed: true,
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
			"configuration": schema.ListNestedBlock{
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
									},
									"shingle_size": schema.Int32Attribute{
										Optional: true,
										Computed: true,
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
												},
												"ratio": schema.Float64Attribute{
													Optional: true,
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
												},
												"ratio": schema.Float64Attribute{
													Optional: true,
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
			// "status": schema.ListNestedBlock{
			// 	CustomType: fwtypes.NewListNestedObjectTypeOf[anomalyDetectorStatusModel](ctx),
			// 	Validators: []validator.List{
			// 		listvalidator.SizeAtMost(1),
			// 	},
			// 	NestedObject: schema.NestedBlockObject{
			// 		Attributes: map[string]schema.Attribute{
			// 			"status_code": schema.StringAttribute{
			// 				Computed: true,
			// 			},
			// 			"status_reason": schema.StringAttribute{
			// 				Computed: true,
			// 			},
			// 		},
			// 	},
			// },
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
	smerr.AddEnrich(ctx, &resp.Diagnostics, fwflex.Expand(ctx, plan, &input))
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

	smerr.AddEnrich(ctx, &resp.Diagnostics, fwflex.Flatten(ctx, out, &state))
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

// TIP: ==== TERRAFORM IMPORTING ====
// The built-in import function, and Import ID Handler, if any, should handle populating the required
// attributes from the Import ID or Resource Identity.
// In some cases, additional attributes must be set when importing.
// Adding a custom ImportState function can handle those.
//
// See more:
// https://hashicorp.github.io/terraform-provider-aws/add-resource-identity-support/
// func (r *anomalyDetectorResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
// 	r.WithImportByIdentity.ImportState(ctx, req, resp)
//
// 	// Set needed attribute values here
// }

// TIP: ==== WAITERS ====
// Some resources of some services have waiters provided by the AWS API.
// Unless they do not work properly, use them rather than defining new ones
// here.
//
// Sometimes we define the wait, status, and find functions in separate
// files, wait.go, status.go, and find.go. Follow the pattern set out in the
// service and define these where it makes the most sense.
//
// If these functions are used in the _test.go file, they will need to be
// exported (i.e., capitalized).
//
// You will need to adjust the parameters and names to fit the service.
func waitAnomalyDetectorCreated(ctx context.Context, conn *amp.Client, id, workspaceID string, timeout time.Duration) (*awstypes.AnomalyDetectorDescription, error) {
	stateConf := &retry.StateChangeConf{
		Pending:                   []string{string(awstypes.AnomalyDetectorStatusCodeCreating)},
		Target:                    []string{string(awstypes.AnomalyDetectorStatusCodeActive)},
		Refresh:                   statusAnomalyDetector(conn, id, workspaceID),
		Timeout:                   timeout,
		NotFoundChecks:            20,
		ContinuousTargetOccurence: 2,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)
	if out, ok := outputRaw.(*awstypes.AnomalyDetectorDescription); ok {
		return out, smarterr.NewError(err)
	}

	return nil, smarterr.NewError(err)
}

// TIP: It is easier to determine whether a resource is updated for some
// resources than others. The best case is a status flag that tells you when
// the update has been fully realized. Other times, you can check to see if a
// key resource argument is updated to a new value or not.
func waitAnomalyDetectorUpdated(ctx context.Context, conn *amp.Client, id, workspaceID string, timeout time.Duration) (*awstypes.AnomalyDetectorDescription, error) {
	stateConf := &retry.StateChangeConf{
		Pending:                   []string{string(awstypes.AnomalyDetectorStatusCodeUpdating)},
		Target:                    []string{string(awstypes.AnomalyDetectorStatusCodeActive)},
		Refresh:                   statusAnomalyDetector(conn, id, workspaceID),
		Timeout:                   timeout,
		NotFoundChecks:            20,
		ContinuousTargetOccurence: 2,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)
	if out, ok := outputRaw.(*awstypes.AnomalyDetectorDescription); ok {
		return out, smarterr.NewError(err)
	}

	return nil, smarterr.NewError(err)
}

// TIP: A deleted waiter is almost like a backwards created waiter. There may
// be additional pending states, however.
func waitAnomalyDetectorDeleted(ctx context.Context, conn *amp.Client, id, workspaceID string, timeout time.Duration) (*awstypes.AnomalyDetectorDescription, error) {
	stateConf := &retry.StateChangeConf{
		Pending: []string{string(awstypes.AnomalyDetectorStatusCodeDeleting), string(awstypes.AnomalyDetectorStatusCodeActive)},
		Target:  []string{},
		Refresh: statusAnomalyDetector(conn, id, workspaceID),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)
	if out, ok := outputRaw.(*awstypes.AnomalyDetectorDescription); ok {
		return out, smarterr.NewError(err)
	}

	return nil, smarterr.NewError(err)
}

// TIP: ==== STATUS ====
// The status function can return an actual status when that field is
// available from the API (e.g., out.Status). Otherwise, you can use custom
// statuses to communicate the states of the resource.
//
// Waiters consume the values returned by status functions. Design status so
// that it can be reused by a create, update, and delete waiter, if possible.
func statusAnomalyDetector(conn *amp.Client, id string, workspaceID string) retry.StateRefreshFunc {
	return func(ctx context.Context) (any, string, error) {
		out, err := findAnomalyDetectorByID(ctx, conn, id, workspaceID)
		if retry.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", smarterr.NewError(err)
		}

		return out, string(out.Status.StatusCode), nil
	}
}

// TIP: ==== FINDERS ====
// The find function is not strictly necessary. You could do the API
// request from the status function. However, we have found that find often
// comes in handy in other places besides the status function. As a result, it
// is good practice to define it separately.
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

// TIP: ==== DATA STRUCTURES ====
// With Terraform Plugin-Framework configurations are deserialized into
// Go types, providing type safety without the need for type assertions.
// These structs should match the schema definition exactly, and the `tfsdk`
// tag value should match the attribute name.
//
// Nested objects are represented in their own data struct. These will
// also have a corresponding attribute type mapping for use inside flex
// functions.
//
// See more:
// https://developer.hashicorp.com/terraform/plugin/framework/handling-data/accessing-values

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
	// Status                      fwtypes.ListNestedObjectValueOf[anomalyDetectorStatusModel]            `tfsdk:"status"`
	Tags        tftags.Map     `tfsdk:"tags"`
	TagsAll     tftags.Map     `tfsdk:"tags_all"`
	Timeouts    timeouts.Value `tfsdk:"timeouts"`
	WorkspaceID types.String   `tfsdk:"workspace_id"`
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
	case *awstypes.IgnoreNearExpectedMemberAmount:
		m.Amount = types.Float64Value(t.Value)
	case *awstypes.IgnoreNearExpectedMemberRatio:
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

type anomalyDetectorStatusModel struct {
	StatusCode   fwtypes.StringEnum[awstypes.AnomalyDetectorStatusCode] `tfsdk:"status_code"`
	StatusReason types.String                                           `tfsdk:"status_reason"`
}

// TIP: ==== IMPORT ID HANDLER ====
// When a resource type has a Resource Identity with multiple attributes, it needs a handler to
// parse the Import ID used for the `terraform import` command or an `import` block with the `id` parameter.
//
// The parser takes the string value of the Import ID and returns:
// * A string value that is typically ignored. See documentation for more details.
// * A map of the resource attributes derived from the Import ID.
// * An error value if there are parsing errors.
//
// For more information, see https://hashicorp.github.io/terraform-provider-aws/resource-identity/#plugin-framework
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
		"id":           anomalyDetectorID,
		"workspace_id": workspaceID,
	}

	return id, result, nil
}

// TIP: ==== SWEEPERS ====
// When acceptance testing resources, interrupted or failed tests may
// leave behind orphaned resources in an account. To facilitate cleaning
// up lingering resources, each resource implementation should include
// a corresponding "sweeper" function.
//
// The sweeper function lists all resources of a given type and sets the
// appropriate identifers required to delete the resource via the Delete
// method implemented above.
//
// Once the sweeper function is implemented, register it in sweep.go
// as follows:
//
//	awsv2.Register("aws_prometheus_anomaly_detector", sweepAnomalyDetectors)
//
// See more:
// https://hashicorp.github.io/terraform-provider-aws/running-and-writing-acceptance-tests/#acceptance-test-sweepers
func sweepAnomalyDetectors(ctx context.Context, client *conns.AWSClient) ([]sweep.Sweepable, error) {
	input := amp.ListAnomalyDetectorsInput{}
	conn := client.AMPClient(ctx)
	var sweepResources []sweep.Sweepable

	pages := amp.NewListAnomalyDetectorsPaginator(conn, &input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)
		if err != nil {
			return nil, smarterr.NewError(err)
		}

		for _, v := range page.AnomalyDetectors {
			sweepResources = append(sweepResources, sweepfw.NewSweepResource(newAnomalyDetectorResource, client,
				sweepfw.NewAttribute(names.AttrID, aws.ToString(v.AnomalyDetectorId))),
			)
		}
	}

	return sweepResources, nil
}
