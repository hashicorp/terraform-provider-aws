// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package backup

import (
	"context"
	"fmt"
	"strings"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/backup"
	awstypes "github.com/aws/aws-sdk-go-v2/service/backup/types"
	"github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/resourcevalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/setvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/mapplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/setplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	sdkid "github.com/hashicorp/terraform-plugin-sdk/v2/helper/id"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/fwdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	fwflex "github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/framework/validators"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkResource("aws_backup_restore_testing_selection", name="Restore Testing Plan Selection")
func newRestoreTestingSelectionResource(_ context.Context) (resource.ResourceWithConfigure, error) {
	r := &restoreTestingSelectionResource{}

	return r, nil
}

type restoreTestingSelectionResource struct {
	framework.ResourceWithConfigure
}

func (r *restoreTestingSelectionResource) Schema(ctx context.Context, request resource.SchemaRequest, response *resource.SchemaResponse) {
	response.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			names.AttrIAMRoleARN: schema.StringAttribute{
				CustomType: fwtypes.ARNType,
				Required:   true,
			},
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
			"protected_resource_arns": schema.SetAttribute{
				CustomType:  fwtypes.SetOfStringType,
				ElementType: types.StringType,
				Optional:    true,
				Computed:    true,
				Validators: []validator.Set{
					setvalidator.ValueStringsAre(
						stringvalidator.Any(
							validators.ARN(),
							stringvalidator.OneOf("*"),
						),
					),
				},
				PlanModifiers: []planmodifier.Set{
					setplanmodifier.UseStateForUnknown(),
				},
			},
			"protected_resource_type": schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"restore_metadata_overrides": schema.MapAttribute{
				CustomType: fwtypes.MapOfStringType,
				Optional:   true,
				Computed:   true,
				PlanModifiers: []planmodifier.Map{
					mapplanmodifier.UseStateForUnknown(),
				},
			},
			"restore_testing_plan_name": schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
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
			"protected_resource_conditions": schema.ListNestedBlock{
				CustomType: fwtypes.NewListNestedObjectTypeOf[protectedResourceConditionsModel](ctx),
				Validators: []validator.List{
					listvalidator.SizeAtMost(1),
				},
				NestedObject: schema.NestedBlockObject{
					Blocks: map[string]schema.Block{
						"string_equals": schema.ListNestedBlock{
							CustomType: fwtypes.NewListNestedObjectTypeOf[keyValueModel](ctx),
							NestedObject: schema.NestedBlockObject{
								Attributes: map[string]schema.Attribute{
									names.AttrKey: schema.StringAttribute{
										Required: true,
									},
									names.AttrValue: schema.StringAttribute{
										Required: true,
									},
								},
							},
						},
						"string_not_equals": schema.ListNestedBlock{
							CustomType: fwtypes.NewListNestedObjectTypeOf[keyValueModel](ctx),
							NestedObject: schema.NestedBlockObject{
								Attributes: map[string]schema.Attribute{
									names.AttrKey: schema.StringAttribute{
										Required: true,
									},
									names.AttrValue: schema.StringAttribute{
										Required: true,
									},
								},
							},
						},
					},
				},
			},
		},
	}
}

func (r *restoreTestingSelectionResource) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	var data restoreTestingSelectionResourceModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().BackupClient(ctx)

	restoreTestingPlanName := data.RestoreTestingPlanName.ValueString()
	name := data.RestoreTestingSelectionName.ValueString()
	input := &backup.CreateRestoreTestingSelectionInput{
		CreatorRequestId:        aws.String(sdkid.UniqueId()),
		RestoreTestingPlanName:  aws.String(restoreTestingPlanName),
		RestoreTestingSelection: &awstypes.RestoreTestingSelectionForCreate{},
	}
	response.Diagnostics.Append(fwflex.Expand(ctx, data, input.RestoreTestingSelection)...)
	if response.Diagnostics.HasError() {
		return
	}

	_, err := conn.CreateRestoreTestingSelection(ctx, input)

	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("creating Backup Restore Testing Selection (%s)", name), err.Error())

		return
	}

	// Set values for unknowns.
	restoreTestingSelection, err := findRestoreTestingSelectionByTwoPartKey(ctx, conn, restoreTestingPlanName, name)

	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("reading Backup Restore Testing Selection (%s)", name), err.Error())

		return
	}

	if v := restoreTestingSelection.ProtectedResourceConditions; v != nil {
		// The default is
		//
		// "ProtectedResourceConditions": {
		// 	"StringEquals": [],
		// 	"StringNotEquals": []
		// },
		if len(v.StringEquals) == 0 {
			v.StringEquals = nil
		}
		if len(v.StringNotEquals) == 0 {
			v.StringNotEquals = nil
		}
		if v.StringEquals == nil && v.StringNotEquals == nil {
			restoreTestingSelection.ProtectedResourceConditions = nil
		}
	}

	response.Diagnostics.Append(fwflex.Flatten(ctx, restoreTestingSelection, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	response.Diagnostics.Append(response.State.Set(ctx, &data)...)
}

func (r *restoreTestingSelectionResource) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	var data restoreTestingSelectionResourceModel
	response.Diagnostics.Append(request.State.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().BackupClient(ctx)

	restoreTestingPlanName := data.RestoreTestingPlanName.ValueString()
	name := data.RestoreTestingSelectionName.ValueString()
	restoreTestingSelection, err := findRestoreTestingSelectionByTwoPartKey(ctx, conn, restoreTestingPlanName, name)

	if tfresource.NotFound(err) {
		response.Diagnostics.Append(fwdiag.NewResourceNotFoundWarningDiagnostic(err))
		response.State.RemoveResource(ctx)

		return
	}

	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("reading Backup Restore Testing Selection (%s)", name), err.Error())

		return
	}

	if v := restoreTestingSelection.ProtectedResourceConditions; v != nil {
		// The default is
		//
		// "ProtectedResourceConditions": {
		// 	"StringEquals": [],
		// 	"StringNotEquals": []
		// },
		if len(v.StringEquals) == 0 {
			v.StringEquals = nil
		}
		if len(v.StringNotEquals) == 0 {
			v.StringNotEquals = nil
		}
		if v.StringEquals == nil && v.StringNotEquals == nil {
			restoreTestingSelection.ProtectedResourceConditions = nil
		}
	}

	// Set attributes for import.
	response.Diagnostics.Append(fwflex.Flatten(ctx, restoreTestingSelection, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	response.Diagnostics.Append(response.State.Set(ctx, &data)...)
}

func (r *restoreTestingSelectionResource) Update(ctx context.Context, request resource.UpdateRequest, response *resource.UpdateResponse) {
	var old, new restoreTestingSelectionResourceModel
	response.Diagnostics.Append(request.State.Get(ctx, &old)...)
	if response.Diagnostics.HasError() {
		return
	}
	response.Diagnostics.Append(request.Plan.Get(ctx, &new)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().BackupClient(ctx)

	if !old.IAMRoleARN.Equal(new.IAMRoleARN) ||
		!old.ProtectedResourceConditions.Equal(new.ProtectedResourceConditions) ||
		!old.RestoreMetadataOverrides.Equal(new.RestoreMetadataOverrides) ||
		!old.ValidationWindowHours.Equal(new.ValidationWindowHours) {
		restoreTestingPlanName := new.RestoreTestingPlanName.ValueString()
		name := new.RestoreTestingSelectionName.ValueString()
		input := &backup.UpdateRestoreTestingSelectionInput{
			RestoreTestingPlanName:      aws.String(restoreTestingPlanName),
			RestoreTestingSelection:     &awstypes.RestoreTestingSelectionForUpdate{},
			RestoreTestingSelectionName: aws.String(name),
		}
		response.Diagnostics.Append(fwflex.Expand(ctx, new, input.RestoreTestingSelection)...)
		if response.Diagnostics.HasError() {
			return
		}

		_, err := conn.UpdateRestoreTestingSelection(ctx, input)

		if err != nil {
			response.Diagnostics.AddError(fmt.Sprintf("updating Backup Restore Testing Selection (%s)", name), err.Error())

			return
		}
	}

	response.Diagnostics.Append(response.State.Set(ctx, &new)...)
}

func (r *restoreTestingSelectionResource) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	var data restoreTestingSelectionResourceModel
	response.Diagnostics.Append(request.State.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().BackupClient(ctx)

	restoreTestingPlanName := data.RestoreTestingPlanName.ValueString()
	name := data.RestoreTestingSelectionName.ValueString()
	input := backup.DeleteRestoreTestingSelectionInput{
		RestoreTestingPlanName:      aws.String(restoreTestingPlanName),
		RestoreTestingSelectionName: aws.String(name),
	}
	_, err := conn.DeleteRestoreTestingSelection(ctx, &input)

	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return
	}

	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("deleting Backup Restore Testing Selection (%s)", name), err.Error())

		return
	}
}

func (r *restoreTestingSelectionResource) ImportState(ctx context.Context, request resource.ImportStateRequest, response *resource.ImportStateResponse) {
	parts := strings.Split(request.ID, ":")
	if len(parts) != 2 {
		response.Diagnostics.AddError("Resource Import Invalid ID", fmt.Sprintf(`Unexpected format for import ID (%s), use: "RestoreTestingSelectionName:RestoreTestingPlanName"`, request.ID))
		return
	}
	response.Diagnostics.Append(response.State.SetAttribute(ctx, path.Root(names.AttrName), parts[0])...)
	response.Diagnostics.Append(response.State.SetAttribute(ctx, path.Root("restore_testing_plan_name"), parts[1])...)
}

func (r *restoreTestingSelectionResource) ConfigValidators(_ context.Context) []resource.ConfigValidator {
	return []resource.ConfigValidator{
		resourcevalidator.ExactlyOneOf(
			path.MatchRoot("protected_resource_arns"),
			path.MatchRoot("protected_resource_conditions"),
		),
	}
}

func findRestoreTestingSelectionByTwoPartKey(ctx context.Context, conn *backup.Client, restoreTestingPlanName, restoreTestingSelectionName string) (*awstypes.RestoreTestingSelectionForGet, error) {
	input := &backup.GetRestoreTestingSelectionInput{
		RestoreTestingPlanName:      aws.String(restoreTestingPlanName),
		RestoreTestingSelectionName: aws.String(restoreTestingSelectionName),
	}

	return findRestoreTestingSelection(ctx, conn, input)
}

func findRestoreTestingSelection(ctx context.Context, conn *backup.Client, input *backup.GetRestoreTestingSelectionInput) (*awstypes.RestoreTestingSelectionForGet, error) {
	output, err := conn.GetRestoreTestingSelection(ctx, input)

	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil || output.RestoreTestingSelection == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	return output.RestoreTestingSelection, nil
}

type restoreTestingSelectionResourceModel struct {
	IAMRoleARN                  fwtypes.ARN                                                       `tfsdk:"iam_role_arn"`
	ProtectedResourceARNs       fwtypes.SetOfString                                               `tfsdk:"protected_resource_arns"`
	ProtectedResourceConditions fwtypes.ListNestedObjectValueOf[protectedResourceConditionsModel] `tfsdk:"protected_resource_conditions"`
	ProtectedResourceType       types.String                                                      `tfsdk:"protected_resource_type"`
	RestoreMetadataOverrides    fwtypes.MapOfString                                               `tfsdk:"restore_metadata_overrides"`
	RestoreTestingSelectionName types.String                                                      `tfsdk:"name"`
	RestoreTestingPlanName      types.String                                                      `tfsdk:"restore_testing_plan_name"`
	ValidationWindowHours       types.Int64                                                       `tfsdk:"validation_window_hours"`
}

type protectedResourceConditionsModel struct {
	StringEquals    fwtypes.ListNestedObjectValueOf[keyValueModel] `tfsdk:"string_equals"`
	StringNotEquals fwtypes.ListNestedObjectValueOf[keyValueModel] `tfsdk:"string_not_equals"`
}

type keyValueModel struct {
	Key   types.String `tfsdk:"key"`
	Value types.String `tfsdk:"value"`
}
