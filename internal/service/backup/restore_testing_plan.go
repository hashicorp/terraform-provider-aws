// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package backup

import (
	"context"
	"errors"
	"time"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/backup"
	awstypes "github.com/aws/aws-sdk-go-v2/service/backup/types"
	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/setvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/setplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	"github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkResource(name="Restore Testing Plan")
// @Tags(identifierAttribute="arn")
func newResourceRestoreTestingPlan(_ context.Context) (resource.ResourceWithConfigure, error) {
	r := &restoreTestingPlanResource{}
	r.SetDefaultCreateTimeout(5 * time.Minute)
	r.SetDefaultDeleteTimeout(5 * time.Minute)
	r.SetDefaultUpdateTimeout(5 * time.Minute)
	return r, nil
}

const (
	ResNameRestoreTestingPlan = "Restore Testing Plan"
)

type restoreTestingPlanResource struct {
	framework.ResourceWithConfigure
	framework.WithImportByID
	framework.WithTimeouts
}

func (r *restoreTestingPlanResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = "aws_backup_restore_testing_plan"
}

func (r *restoreTestingPlanResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			names.AttrARN: framework.ARNAttributeComputedOnly(),
			names.AttrName: schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
				Validators: []validator.String{
					stringvalidator.LengthBetween(1, 50),
					stringvalidator.RegexMatches(regexache.MustCompile(`^[0-9A-Za-z_]+$`), "must contain only alphanumeric characters, and underscores"),
				},
			},
			"schedule_expression": schema.StringAttribute{
				Required: true,
			},
			"schedule_expression_timezone": schema.StringAttribute{
				Computed: true,
				Optional: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"start_window_hours": schema.Int64Attribute{
				Optional: true,
				Computed: true,
				Validators: []validator.Int64{
					int64validator.Between(1, 168),
				},
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
			},
			names.AttrTags:    tftags.TagsAttribute(),
			names.AttrTagsAll: tftags.TagsAttributeComputedOnly(),
		},
		Blocks: map[string]schema.Block{
			"recovery_point_selection": schema.ListNestedBlock{
				CustomType: fwtypes.NewListNestedObjectTypeOf[restoreRecoveryPointSelectionModel](ctx),
				Validators: []validator.List{
					listvalidator.SizeAtLeast(1),
					listvalidator.SizeAtMost(1),
				},
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"algorithm": schema.StringAttribute{
							Required: true,
							Validators: []validator.String{
								stringvalidator.OneOf("RANDOM_WITHIN_WINDOW", "LATEST_WITHIN_WINDOW"),
							},
						},
						"include_vaults": schema.SetAttribute{
							CustomType:  fwtypes.SetOfStringType,
							ElementType: types.StringType,
							Required:    true,
							Validators: []validator.Set{
								setvalidator.SizeAtLeast(1),
								setvalidator.ValueStringsAre(
									stringvalidator.RegexMatches(regexache.MustCompile(`^arn:aws:backup:\w+(?:-\w+)+:\d{12}:backup-vault:[A-Za-z0-9_\-\*]+|\*$`), "must be either an AWS ARN for a backup vault or a *"),
								),
							},
						},
						"recovery_point_types": schema.SetAttribute{
							Required:    true,
							CustomType:  fwtypes.SetOfStringType,
							ElementType: types.StringType,
							Validators: []validator.Set{
								setvalidator.SizeAtLeast(1),
								setvalidator.ValueStringsAre(
									stringvalidator.OneOf("CONTINUOUS", "SNAPSHOT"),
								),
							},
						},
						"exclude_vaults": schema.SetAttribute{
							CustomType:  fwtypes.SetOfStringType,
							ElementType: types.StringType,
							Optional:    true,
							Computed:    true,
							Validators: []validator.Set{
								setvalidator.ValueStringsAre(
									stringvalidator.RegexMatches(regexache.MustCompile(`^arn:aws:backup:\w+(?:-\w+)+:\d{12}:backup-vault:[A-Za-z0-9_\-\*]+|\*$`), "must be either an AWS ARN for a backup vault or a *"),
								),
							},
							PlanModifiers: []planmodifier.Set{
								setplanmodifier.UseStateForUnknown(),
							},
						},
						"selection_window_days": schema.Int64Attribute{
							Optional: true,
							Computed: true,
							Validators: []validator.Int64{
								int64validator.Between(1, 365),
							},
							PlanModifiers: []planmodifier.Int64{
								int64planmodifier.UseStateForUnknown(),
							},
						},
					},
				},
			},
			"timeouts": timeouts.Block(ctx, timeouts.Opts{
				Create: true,
				Delete: true,
			}),
		},
	}
}

func (r *restoreTestingPlanResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	conn := r.Meta().BackupClient(ctx)

	var plan restoreTestingPlanResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	in := &awstypes.RestoreTestingPlanForCreate{}
	resp.Diagnostics.Append(flex.Expand(ctx, plan, in)...)
	if resp.Diagnostics.HasError() {
		return
	}

	tags := make(map[string]string)
	for k, v := range getTagsIn(ctx) {
		tags[k] = v
	}

	out, err := conn.CreateRestoreTestingPlan(ctx, &backup.CreateRestoreTestingPlanInput{
		RestoreTestingPlan: in,
		Tags:               tags,
	})

	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.Backup, create.ErrActionCreating, ResNameRestoreTestingPlan, plan.RestoreTestingPlanName.ValueString(), err),
			err.Error(),
		)
		return
	}
	if out == nil || out.RestoreTestingPlanName == nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.Backup, create.ErrActionCreating, ResNameRestoreTestingPlan, plan.RestoreTestingPlanName.ValueString(), nil),
			errors.New("empty output").Error(),
		)
		return
	}

	plan.RestoreTestingPlanArn = flex.StringToFramework(ctx, out.RestoreTestingPlanArn)

	// "wait" for creation.. this is to get the _optional_ values
	created, err := waitRestoreTestingPlanLatest(ctx, conn, plan.RestoreTestingPlanName.ValueString(), r.CreateTimeout(ctx, plan.Timeouts))
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.Backup, create.ErrActionWaitingForCreation, ResNameRestoreTestingPlan, "", err),
			err.Error(),
		)
		return
	}

	// var state restoreTestingSelectionResourceModel
	resp.Diagnostics.Append(flex.Flatten(ctx, created.RestoreTestingPlan, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *restoreTestingPlanResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	conn := r.Meta().BackupClient(ctx)

	var state restoreTestingPlanResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	out, err := findRestoreTestingPlanByName(ctx, conn, state.RestoreTestingPlanName.ValueString())

	if tfresource.NotFound(err) {
		resp.State.RemoveResource(ctx)
		return
	}

	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.Backup, create.ErrActionSetting, ResNameRestoreTestingPlan, state.RestoreTestingPlanName.ValueString(), err),
			err.Error(),
		)
		return
	}

	resp.Diagnostics.Append(flex.Flatten(ctx, out.RestoreTestingPlan, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if ok := out.ResultMetadata.Has("Tags"); ok {
		v := out.ResultMetadata.Get("Tags")
		setTagsOut(ctx, v.(map[string]string))
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *restoreTestingPlanResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	conn := r.Meta().BackupClient(ctx)

	var state, plan restoreTestingPlanResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if !state.RestoreTestingPlanName.Equal(plan.RestoreTestingPlanName) {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.Backup, create.ErrActionUpdating, ResNameRestoreTestingPlan, plan.RestoreTestingPlanName.ValueString(), errors.New("name changes are not supported")),
			"changing the name of a restore testing plan is not supported",
		)
		return
	}

	if !state.RecoveryPointSelection.Equal(plan.RecoveryPointSelection) ||
		!state.ScheduleExpression.Equal(plan.ScheduleExpression) ||
		!state.ScheduleExpressionTimezone.Equal(plan.ScheduleExpressionTimezone) ||
		!state.StartWindowHours.Equal(plan.StartWindowHours) {

		in := &awstypes.RestoreTestingPlanForUpdate{}
		resp.Diagnostics.Append(flex.Expand(ctx, &plan, in)...)
		if resp.Diagnostics.HasError() {
			return
		}

		out, err := conn.UpdateRestoreTestingPlan(ctx, &backup.UpdateRestoreTestingPlanInput{
			RestoreTestingPlanName: aws.String(plan.RestoreTestingPlanName.ValueString()),
			RestoreTestingPlan:     in,
		})

		if err != nil {
			resp.Diagnostics.AddError(
				create.ProblemStandardMessage(names.Backup, create.ErrActionUpdating, ResNameRestoreTestingPlan, plan.RestoreTestingPlanName.ValueString(), err),
				err.Error(),
			)
			return
		}
		if out == nil || out.RestoreTestingPlanName == nil {
			resp.Diagnostics.AddError(
				create.ProblemStandardMessage(names.Backup, create.ErrActionUpdating, ResNameRestoreTestingPlan, plan.RestoreTestingPlanName.ValueString(), nil),
				errors.New("empty output").Error(),
			)
			return
		}

		plan.RestoreTestingPlanArn = flex.StringToFramework(ctx, out.RestoreTestingPlanArn)

		// "wait" for update.. this is to get the _optional_ values
		created, err := waitRestoreTestingPlanLatest(ctx, conn, plan.RestoreTestingPlanName.ValueString(), r.UpdateTimeout(ctx, plan.Timeouts))
		if err != nil {
			resp.Diagnostics.AddError(
				create.ProblemStandardMessage(names.Backup, create.ErrActionWaitingForUpdate, ResNameRestoreTestingPlan, "", err),
				err.Error(),
			)
			return
		}

		resp.Diagnostics.Append(flex.Flatten(ctx, created.RestoreTestingPlan, &plan)...)
		if resp.Diagnostics.HasError() {
			return
		}

		resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *restoreTestingPlanResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	conn := r.Meta().BackupClient(ctx)

	var state restoreTestingPlanResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	in := &backup.DeleteRestoreTestingPlanInput{
		RestoreTestingPlanName: state.RestoreTestingPlanName.ValueStringPointer(),
	}

	_, err := conn.DeleteRestoreTestingPlan(ctx, in)

	if err != nil {
		if errs.IsA[*awstypes.ResourceNotFoundException](err) {
			return
		}
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.Backup, create.ErrActionDeleting, ResNameRestoreTestingPlan, state.RestoreTestingPlanName.String(), err),
			err.Error(),
		)
		return
	}

	if _, err := waitRestoreTestingPlanDeleted(ctx, conn, state.RestoreTestingPlanName.ValueString(), r.DeleteTimeout(ctx, state.Timeouts)); err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.Backup, create.ErrActionWaitingForDeletion, ResNameRestoreTestingPlan, state.RestoreTestingPlanName.String(), err),
			err.Error(),
		)
		return
	}
}

func (r *restoreTestingPlanResource) ModifyPlan(ctx context.Context, request resource.ModifyPlanRequest, response *resource.ModifyPlanResponse) {
	r.SetTagsAll(ctx, request, response)
}

func (r *restoreTestingPlanResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root(names.AttrName), req.ID)...)
}

const (
	stateNormal   = "NORMAL"
	stateNotFound = "NOT_FOUND"
)

func waitRestoreTestingPlanDeleted(ctx context.Context, conn *backup.Client, name string, timeout time.Duration) (*backup.GetRestoreTestingPlanOutput, error) {
	stateConf := &retry.StateChangeConf{
		Pending: []string{stateNormal},
		Target:  []string{},
		Refresh: statusRestorePlan(ctx, conn, name),
		Timeout: timeout,
	}
	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if out, ok := outputRaw.(*backup.GetRestoreTestingPlanOutput); ok {
		return out, err
	}

	return nil, err
}

func waitRestoreTestingPlanLatest(ctx context.Context, conn *backup.Client, name string, timeout time.Duration) (*backup.GetRestoreTestingPlanOutput, error) {
	stateConf := &retry.StateChangeConf{
		Pending: []string{},
		Target:  []string{stateNormal},
		Refresh: statusRestorePlan(ctx, conn, name),
		Timeout: timeout,
	}
	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if out, ok := outputRaw.(*backup.GetRestoreTestingPlanOutput); ok {
		return out, err
	}

	return nil, err
}

func statusRestorePlan(ctx context.Context, conn *backup.Client, name string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		output, err := findRestoreTestingPlanByName(ctx, conn, name)
		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, stateNormal, nil
	}
}

type restoreTestingPlanResourceModel struct {
	RestoreTestingPlanArn      types.String                                                        `tfsdk:"arn"`
	RestoreTestingPlanName     types.String                                                        `tfsdk:"name"`
	ScheduleExpression         types.String                                                        `tfsdk:"schedule_expression"`
	ScheduleExpressionTimezone types.String                                                        `tfsdk:"schedule_expression_timezone"`
	StartWindowHours           types.Int64                                                         `tfsdk:"start_window_hours"`
	RecoveryPointSelection     fwtypes.ListNestedObjectValueOf[restoreRecoveryPointSelectionModel] `tfsdk:"recovery_point_selection"`
	Tags                       types.Map                                                           `tfsdk:"tags"`
	TagsAll                    types.Map                                                           `tfsdk:"tags_all"`
	Timeouts                   timeouts.Value                                                      `tfsdk:"timeouts"`
}

type restoreRecoveryPointSelectionModel struct {
	Algorithm           types.String                     `tfsdk:"algorithm"`
	IncludeVaults       fwtypes.SetValueOf[types.String] `tfsdk:"include_vaults"`
	RecoveryPointTypes  fwtypes.SetValueOf[types.String] `tfsdk:"recovery_point_types"`
	ExcludeVaults       fwtypes.SetValueOf[types.String] `tfsdk:"exclude_vaults"`
	SelectionWindowDays types.Int64                      `tfsdk:"selection_window_days"`
}
