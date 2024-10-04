// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package backup

import (
	"context"
	"fmt"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/backup"
	awstypes "github.com/aws/aws-sdk-go-v2/service/backup/types"
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
	sdkid "github.com/hashicorp/terraform-plugin-sdk/v2/helper/id"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/fwdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	fwflex "github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/framework/validators"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkResource(name="Restore Testing Plan")
// @Tags(identifierAttribute="arn")
func newRestoreTestingPlanResource(_ context.Context) (resource.ResourceWithConfigure, error) {
	r := &restoreTestingPlanResource{}

	return r, nil
}

type restoreTestingPlanResource struct {
	framework.ResourceWithConfigure
}

func (*restoreTestingPlanResource) Metadata(_ context.Context, request resource.MetadataRequest, response *resource.MetadataResponse) {
	response.TypeName = "aws_backup_restore_testing_plan"
}

func (r *restoreTestingPlanResource) Schema(ctx context.Context, request resource.SchemaRequest, response *resource.SchemaResponse) {
	response.Schema = schema.Schema{
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
			names.AttrScheduleExpression: schema.StringAttribute{
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
					int64validator.Between(0, 168),
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
					listvalidator.IsRequired(),
					listvalidator.SizeAtLeast(1),
					listvalidator.SizeAtMost(1),
				},
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"algorithm": schema.StringAttribute{
							CustomType: fwtypes.StringEnumType[awstypes.RestoreTestingRecoveryPointSelectionAlgorithm](),
							Required:   true,
						},
						"exclude_vaults": schema.SetAttribute{
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
						"include_vaults": schema.SetAttribute{
							CustomType:  fwtypes.SetOfStringType,
							ElementType: types.StringType,
							Required:    true,
							Validators: []validator.Set{
								setvalidator.ValueStringsAre(
									stringvalidator.Any(
										validators.ARN(),
										stringvalidator.OneOf("*"),
									),
								),
							},
						},
						"recovery_point_types": schema.SetAttribute{
							CustomType:  fwtypes.NewSetTypeOf[fwtypes.StringEnum[awstypes.RestoreTestingRecoveryPointType]](ctx),
							Required:    true,
							ElementType: fwtypes.StringEnumType[awstypes.RestoreTestingRecoveryPointType](),
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
		},
	}
}

func (r *restoreTestingPlanResource) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	var data restoreTestingPlanResourceModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().BackupClient(ctx)

	name := data.RestoreTestingPlanName.ValueString()
	input := &backup.CreateRestoreTestingPlanInput{
		CreatorRequestId:   aws.String(sdkid.UniqueId()),
		RestoreTestingPlan: &awstypes.RestoreTestingPlanForCreate{},
		Tags:               getTagsIn(ctx),
	}
	response.Diagnostics.Append(fwflex.Expand(ctx, data, input.RestoreTestingPlan)...)
	if response.Diagnostics.HasError() {
		return
	}

	_, err := conn.CreateRestoreTestingPlan(ctx, input)

	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("creating Backup Restore Testing Plan (%s)", name), err.Error())

		return
	}

	// Set values for unknowns.
	restoreTestingPlan, err := findRestoreTestingPlanByName(ctx, conn, name)

	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("reading Backup Restore Testing Plan (%s)", name), err.Error())

		return
	}

	response.Diagnostics.Append(fwflex.Flatten(ctx, restoreTestingPlan, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	response.Diagnostics.Append(response.State.Set(ctx, &data)...)
}

func (r *restoreTestingPlanResource) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	var data restoreTestingPlanResourceModel
	response.Diagnostics.Append(request.State.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().BackupClient(ctx)

	name := data.RestoreTestingPlanName.ValueString()
	restoreTestingPlan, err := findRestoreTestingPlanByName(ctx, conn, name)

	if tfresource.NotFound(err) {
		response.Diagnostics.Append(fwdiag.NewResourceNotFoundWarningDiagnostic(err))
		response.State.RemoveResource(ctx)

		return
	}

	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("reading Backup Restore Testing Plan (%s)", name), err.Error())

		return
	}

	// Set attributes for import.
	response.Diagnostics.Append(fwflex.Flatten(ctx, restoreTestingPlan, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	response.Diagnostics.Append(response.State.Set(ctx, &data)...)
}

func (r *restoreTestingPlanResource) Update(ctx context.Context, request resource.UpdateRequest, response *resource.UpdateResponse) {
	var old, new restoreTestingPlanResourceModel
	response.Diagnostics.Append(request.State.Get(ctx, &old)...)
	if response.Diagnostics.HasError() {
		return
	}
	response.Diagnostics.Append(request.Plan.Get(ctx, &new)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().BackupClient(ctx)

	if !old.RecoveryPointSelection.Equal(new.RecoveryPointSelection) ||
		!old.ScheduleExpression.Equal(new.ScheduleExpression) ||
		!old.ScheduleExpressionTimezone.Equal(new.ScheduleExpressionTimezone) ||
		!old.StartWindowHours.Equal(new.StartWindowHours) {
		name := new.RestoreTestingPlanName.ValueString()
		input := &backup.UpdateRestoreTestingPlanInput{
			RestoreTestingPlan:     &awstypes.RestoreTestingPlanForUpdate{},
			RestoreTestingPlanName: aws.String(name),
		}
		response.Diagnostics.Append(fwflex.Expand(ctx, new, input.RestoreTestingPlan)...)
		if response.Diagnostics.HasError() {
			return
		}

		_, err := conn.UpdateRestoreTestingPlan(ctx, input)

		if err != nil {
			response.Diagnostics.AddError(fmt.Sprintf("updating Backup Restore Testing Plan (%s)", name), err.Error())

			return
		}
	}

	response.Diagnostics.Append(response.State.Set(ctx, &new)...)
}

func (r *restoreTestingPlanResource) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	var data restoreTestingPlanResourceModel
	response.Diagnostics.Append(request.State.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().BackupClient(ctx)

	name := data.RestoreTestingPlanName.ValueString()
	_, err := conn.DeleteRestoreTestingPlan(ctx, &backup.DeleteRestoreTestingPlanInput{
		RestoreTestingPlanName: aws.String(name),
	})

	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return
	}

	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("deleting Backup Restore Testing Plan (%s)", name), err.Error())

		return
	}
}

func (r *restoreTestingPlanResource) ModifyPlan(ctx context.Context, request resource.ModifyPlanRequest, response *resource.ModifyPlanResponse) {
	r.SetTagsAll(ctx, request, response)
}

func (r *restoreTestingPlanResource) ImportState(ctx context.Context, request resource.ImportStateRequest, response *resource.ImportStateResponse) {
	response.Diagnostics.Append(response.State.SetAttribute(ctx, path.Root(names.AttrName), request.ID)...)
}

func findRestoreTestingPlanByName(ctx context.Context, conn *backup.Client, name string) (*awstypes.RestoreTestingPlanForGet, error) {
	input := &backup.GetRestoreTestingPlanInput{
		RestoreTestingPlanName: aws.String(name),
	}

	return findRestoreTestingPlan(ctx, conn, input)
}

func findRestoreTestingPlan(ctx context.Context, conn *backup.Client, input *backup.GetRestoreTestingPlanInput) (*awstypes.RestoreTestingPlanForGet, error) {
	output, err := conn.GetRestoreTestingPlan(ctx, input)

	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil || output.RestoreTestingPlan == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	return output.RestoreTestingPlan, nil
}

type restoreTestingPlanResourceModel struct {
	RecoveryPointSelection     fwtypes.ListNestedObjectValueOf[restoreRecoveryPointSelectionModel] `tfsdk:"recovery_point_selection"`
	RestoreTestingPlanARN      types.String                                                        `tfsdk:"arn"`
	RestoreTestingPlanName     types.String                                                        `tfsdk:"name"`
	ScheduleExpression         types.String                                                        `tfsdk:"schedule_expression"`
	ScheduleExpressionTimezone types.String                                                        `tfsdk:"schedule_expression_timezone"`
	StartWindowHours           types.Int64                                                         `tfsdk:"start_window_hours"`
	Tags                       tftags.Map                                                          `tfsdk:"tags"`
	TagsAll                    tftags.Map                                                          `tfsdk:"tags_all"`
}

type restoreRecoveryPointSelectionModel struct {
	Algorithm           fwtypes.StringEnum[awstypes.RestoreTestingRecoveryPointSelectionAlgorithm]       `tfsdk:"algorithm"`
	ExcludeVaults       fwtypes.SetOfString                                                              `tfsdk:"exclude_vaults"`
	IncludeVaults       fwtypes.SetOfString                                                              `tfsdk:"include_vaults"`
	RecoveryPointTypes  fwtypes.SetValueOf[fwtypes.StringEnum[awstypes.RestoreTestingRecoveryPointType]] `tfsdk:"recovery_point_types"`
	SelectionWindowDays types.Int64                                                                      `tfsdk:"selection_window_days"`
}
