// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package backup

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/service/backup"
	awstypes "github.com/aws/aws-sdk-go-v2/service/backup/types"
	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/objectvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/listplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/objectplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	"github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkResource(name="Restore Testing Plan Selection")
func newResourceRestoreTestingSelection(_ context.Context) (resource.ResourceWithConfigure, error) {
	r := &resourceRestoreTestingSelection{}

	r.SetDefaultCreateTimeout(5 * time.Minute)
	r.SetDefaultUpdateTimeout(5 * time.Minute)
	r.SetDefaultDeleteTimeout(5 * time.Minute)

	return r, nil
}

const (
	ResNameRestoreTestingSelection = "Restore Testing Selection"
)

type resourceRestoreTestingSelection struct {
	framework.ResourceWithConfigure
	framework.WithTimeouts
}

func (r *resourceRestoreTestingSelection) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = "aws_backup_restore_testing_selection"
}

func (r *resourceRestoreTestingSelection) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
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
			"restore_testing_plan_name": schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"protected_resource_type": schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"iam_role_arn": schema.StringAttribute{
				Required: true,
			},
			"restore_metadata_overrides": schema.MapAttribute{
				Optional:   true,
				Computed:   true,
				CustomType: fwtypes.NewMapTypeOf[types.String](ctx),
			},
			"validation_window_hours": schema.Int64Attribute{
				Optional: true,
				Computed: true,
				Validators: []validator.Int64{
					int64validator.Between(1, 168),
				},
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
			},
		},
		Blocks: map[string]schema.Block{
			"protected_resource_conditions": schema.SingleNestedBlock{
				CustomType: fwtypes.NewObjectTypeOf[protectedResourceConditionsModel](ctx),
				PlanModifiers: []planmodifier.Object{
					objectplanmodifier.UseStateForUnknown(),
				},
				Validators: []validator.Object{
					objectvalidator.IsRequired(),
				},
				Attributes: map[string]schema.Attribute{
					"string_equals": schema.ListAttribute{
						CustomType: fwtypes.NewListNestedObjectTypeOf[keyValueMap](ctx),
						Optional:   true,
						Computed:   true,
						PlanModifiers: []planmodifier.List{
							listplanmodifier.UseStateForUnknown(),
						},
					},
					"string_not_equals": schema.ListAttribute{
						CustomType: fwtypes.NewListNestedObjectTypeOf[keyValueMap](ctx),
						Optional:   true,
						Computed:   true,
						PlanModifiers: []planmodifier.List{
							listplanmodifier.UseStateForUnknown(),
						},
					},
				},
			},
			"timeouts": timeouts.Block(ctx, timeouts.Opts{
				Create: true,
				Update: true,
				Delete: true,
			}),
		},
	}
}

func (r *resourceRestoreTestingSelection) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	conn := r.Meta().BackupClient(ctx)

	var plan restoreTestingSelectionResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	requestBody := &awstypes.RestoreTestingSelectionForCreate{}
	resp.Diagnostics.Append(flex.Expand(ctx, plan, requestBody)...)
	if resp.Diagnostics.HasError() {
		return
	}

	input := &backup.CreateRestoreTestingSelectionInput{
		RestoreTestingSelection: requestBody,
		RestoreTestingPlanName:  plan.RestoreTestingPlanName.ValueStringPointer(),
	}

	out, err := conn.CreateRestoreTestingSelection(ctx, input)
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.Backup, create.ErrActionCreating, ResNameRestoreTestingSelection, plan.RestoreTestingSelectionName.ValueString(), err),
			err.Error(),
		)
		return
	}
	if out == nil || out.CreationTime == nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.Backup, create.ErrActionCreating, ResNameRestoreTestingSelection, plan.RestoreTestingSelectionName.ValueString(), err),
			errors.New("empty output").Error(),
		)
		return
	}

	// "wait" for creation.. this is to get the _optional_ values
	created, err := waitRestoreTestingSelectionLatest(ctx, conn, plan.RestoreTestingSelectionName.ValueString(), plan.RestoreTestingPlanName.ValueString(), r.CreateTimeout(ctx, plan.Timeouts))
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.Backup, create.ErrActionWaitingForCreation, ResNameRestoreTestingPlan, "", err),
			err.Error(),
		)
		return
	}

	// var state restoreTestingSelectionResourceModel
	resp.Diagnostics.Append(flex.Flatten(ctx, created.RestoreTestingSelection, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *resourceRestoreTestingSelection) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	conn := r.Meta().BackupClient(ctx)

	var state restoreTestingSelectionResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	out, err := findRestoreTestingSelectionByName(ctx, conn, state.RestoreTestingSelectionName.ValueString(), state.RestoreTestingPlanName.ValueString())

	if tfresource.NotFound(err) {
		resp.State.RemoveResource(ctx)
		return
	}

	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.Backup, create.ErrActionSetting, ResNameRestoreTestingSelection, state.RestoreTestingSelectionName.ValueString(), err),
			err.Error(),
		)
		return
	}

	resp.Diagnostics.Append(flex.Flatten(ctx, out.RestoreTestingSelection, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *resourceRestoreTestingSelection) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	conn := r.Meta().BackupClient(ctx)

	var state, plan restoreTestingSelectionResourceModel
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

	if !state.RestoreMetadataOverrides.Equal(plan.RestoreMetadataOverrides) ||
		!state.ValidationWindowHours.Equal(plan.ValidationWindowHours) ||
		!state.IamRoleArn.Equal(plan.IamRoleArn) ||
		!state.ProtectedResourceConditions.Equal(plan.ProtectedResourceConditions) {

		in := &awstypes.RestoreTestingSelectionForUpdate{}
		resp.Diagnostics.Append(flex.Expand(ctx, plan, in)...)
		if resp.Diagnostics.HasError() {
			return
		}

		out, err := conn.UpdateRestoreTestingSelection(ctx, &backup.UpdateRestoreTestingSelectionInput{
			RestoreTestingSelectionName: plan.RestoreTestingSelectionName.ValueStringPointer(),
			RestoreTestingPlanName:      plan.RestoreTestingPlanName.ValueStringPointer(),
			RestoreTestingSelection:     in,
		})

		if err != nil {
			resp.Diagnostics.AddError(
				create.ProblemStandardMessage(names.Backup, create.ErrActionUpdating, ResNameRestoreTestingSelection, plan.RestoreTestingPlanName.ValueString(), err),
				err.Error(),
			)
			return
		}
		if out == nil || out.RestoreTestingPlanName == nil {
			resp.Diagnostics.AddError(
				create.ProblemStandardMessage(names.Backup, create.ErrActionUpdating, ResNameRestoreTestingSelection, plan.RestoreTestingPlanName.ValueString(), nil),
				errors.New("empty output").Error(),
			)
			return
		}

		// "wait" for update.. this is to get the _optional_ values
		created, err := waitRestoreTestingSelectionLatest(ctx, conn, plan.RestoreTestingSelectionName.ValueString(), plan.RestoreTestingPlanName.ValueString(), r.UpdateTimeout(ctx, plan.Timeouts))
		if err != nil {
			resp.Diagnostics.AddError(
				create.ProblemStandardMessage(names.Backup, create.ErrActionWaitingForUpdate, ResNameRestoreTestingSelection, "", err),
				err.Error(),
			)
			return
		}

		resp.Diagnostics.Append(flex.Flatten(ctx, created.RestoreTestingSelection, &plan)...)
		if resp.Diagnostics.HasError() {
			return
		}

		resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *resourceRestoreTestingSelection) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	conn := r.Meta().BackupClient(ctx)

	var state restoreTestingSelectionResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	in := &backup.DeleteRestoreTestingSelectionInput{
		RestoreTestingSelectionName: state.RestoreTestingSelectionName.ValueStringPointer(),
		RestoreTestingPlanName:      state.RestoreTestingPlanName.ValueStringPointer(),
	}

	_, err := conn.DeleteRestoreTestingSelection(ctx, in)

	if err != nil {
		if errs.IsA[*awstypes.ResourceNotFoundException](err) {
			return
		}
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.Backup, create.ErrActionDeleting, ResNameRestoreTestingPlan, state.RestoreTestingSelectionName.String(), err),
			err.Error(),
		)
		return
	}

	if _, err := waitRestoreTestingSelectionDeleted(ctx, conn, state.RestoreTestingSelectionName.ValueString(), state.RestoreTestingPlanName.ValueString(), r.DeleteTimeout(ctx, state.Timeouts)); err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.Backup, create.ErrActionWaitingForDeletion, ResNameRestoreTestingPlan, state.RestoreTestingSelectionName.String(), err),
			err.Error(),
		)
		return
	}
}

func (r *resourceRestoreTestingSelection) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	parts := strings.Split(req.ID, ":")
	if len(parts) != 2 {
		resp.Diagnostics.AddError("Resource Import Invalid ID", fmt.Sprintf(`Unexpected format for import ID (%s), use: "RestoreTestingSelectionName:RestoreTestingPlanName"`, req.ID))
		return
	}
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root(names.AttrName), parts[0])...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("restore_testing_plan_name"), parts[1])...)
}

// ==== WAITERS ==== //
func waitRestoreTestingSelectionLatest(ctx context.Context, conn *backup.Client, name string, restoreTestingPlanName string, timeout time.Duration) (*backup.GetRestoreTestingSelectionOutput, error) {
	stateConf := &retry.StateChangeConf{
		Pending: []string{},
		Target:  []string{stateNormal},
		Refresh: statusRestoreSelection(ctx, conn, name, restoreTestingPlanName),
		Timeout: timeout,
	}
	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if out, ok := outputRaw.(*backup.GetRestoreTestingSelectionOutput); ok {
		return out, err
	}

	return nil, err

}

func waitRestoreTestingSelectionDeleted(ctx context.Context, conn *backup.Client, name string, restoreTestingPlanName string, timeout time.Duration) (*backup.GetRestoreTestingSelectionOutput, error) {
	stateConf := &retry.StateChangeConf{
		Pending: []string{stateNormal},
		Target:  []string{},
		Refresh: statusRestoreSelection(ctx, conn, name, restoreTestingPlanName),
		Timeout: timeout,
	}
	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if out, ok := outputRaw.(*backup.GetRestoreTestingSelectionOutput); ok {
		return out, err
	}

	return nil, err

}

func statusRestoreSelection(ctx context.Context, conn *backup.Client, name, restoreTestingPlanName string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		output, err := findRestoreTestingSelectionByName(ctx, conn, name, restoreTestingPlanName)
		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, stateNormal, nil
	}
}

type restoreTestingSelectionResourceModel struct {
	RestoreTestingSelectionName types.String                                            `tfsdk:"name"`
	RestoreTestingPlanName      types.String                                            `tfsdk:"restore_testing_plan_name"`
	ProtectedResourceConditions fwtypes.ObjectValueOf[protectedResourceConditionsModel] `tfsdk:"protected_resource_conditions"`
	IamRoleArn                  types.String                                            `tfsdk:"iam_role_arn"`
	ValidationWindowHours       types.Int64                                             `tfsdk:"validation_window_hours"`
	ProtectedResourceType       types.String                                            `tfsdk:"protected_resource_type"`
	RestoreMetadataOverrides    fwtypes.MapValueOf[basetypes.StringValue]               `tfsdk:"restore_metadata_overrides"`
	Timeouts                    timeouts.Value                                          `tfsdk:"timeouts"`
}

type protectedResourceConditionsModel struct {
	StringEquals    fwtypes.ListNestedObjectValueOf[keyValueMap] `tfsdk:"string_equals"`
	StringNotEquals fwtypes.ListNestedObjectValueOf[keyValueMap] `tfsdk:"string_not_equals"`
}

type keyValueMap struct {
	Key   types.String `tfsdk:"key"`
	Value types.String `tfsdk:"value"`
}
