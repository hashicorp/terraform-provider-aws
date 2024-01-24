// // Copyright (c) HashiCorp, Inc.
// // SPDX-License-Identifier: MPL-2.0
package finspace

import (
	"context"
	"errors"
	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/listplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/id"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	fwflex "github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/finspace"
	awstypes "github.com/aws/aws-sdk-go-v2/service/finspace/types"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	//"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	//"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	//"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkResource(name="Kx Volume")
// @Tags(identifierAttribute="arn")
func newResourceKxVolume(context.Context) (resource.ResourceWithConfigure, error) {
	r := &resourceKxVolume{}
	r.SetDefaultCreateTimeout(30 * time.Minute)
	r.SetDefaultReadTimeout(30 * time.Minute)
	r.SetDefaultDeleteTimeout(45 * time.Minute)
	r.SetMigratedFromPluginSDK(true)

	return r, nil
}

type resourceKxVolume struct {
	framework.ResourceWithConfigure
	framework.WithTimeouts
}

func (r *resourceKxVolume) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = "aws_finspace_kx_volume"
}

func (r *resourceKxVolume) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func (r *resourceKxVolume) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"id":  framework.IDAttribute(),
			"arn": framework.ARNAttributeComputedOnly(),
			"availability_zones": schema.ListAttribute{
				ElementType: types.StringType,
				Required:    true,
				PlanModifiers: []planmodifier.List{
					listplanmodifier.RequiresReplace(),
				},
			},
			"az_mode": schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
				Validators: []validator.String{
					enum.FrameworkValidate[awstypes.KxAzMode](),
				},
			},
			"description": schema.StringAttribute{
				Optional: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
				Validators: []validator.String{
					stringvalidator.LengthBetween(1, 1000),
				},
			},
			"environment_id": schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
				Validators: []validator.String{
					stringvalidator.LengthBetween(1, 63),
				},
			},
			"name": schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
				Validators: []validator.String{
					stringvalidator.LengthBetween(3, 63),
				},
			},
			"status": schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"status_reason": schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"type": schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
				Validators: []validator.String{
					enum.FrameworkValidate[awstypes.KxVolumeType](),
				},
			},
			"created_timestamp": schema.StringAttribute{
				Computed: true,
			},
			"last_modified_timestamp": schema.StringAttribute{
				Computed: true,
			},
			names.AttrTags:    tftags.TagsAttribute(),
			names.AttrTagsAll: tftags.TagsAttributeComputedOnly(),
		},
		Blocks: map[string]schema.Block{
			names.AttrTimeouts: timeouts.Block(ctx, timeouts.Opts{
				Create: true,
				Update: true,
				Delete: true,
			}),
			"attached_clusters": schema.ListNestedBlock{
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"cluster_name": schema.StringAttribute{
							Required: true,
							PlanModifiers: []planmodifier.String{
								stringplanmodifier.RequiresReplace(),
							},
							Validators: []validator.String{
								stringvalidator.LengthBetween(3, 63),
							},
						},
						"cluster_status": schema.StringAttribute{
							Required: true,
							PlanModifiers: []planmodifier.String{
								stringplanmodifier.RequiresReplace(),
							},
							Validators: []validator.String{
								enum.FrameworkValidate[awstypes.KxClusterStatus](),
							},
						},
						"cluster_type": schema.StringAttribute{
							Required: true,
							PlanModifiers: []planmodifier.String{
								stringplanmodifier.RequiresReplace(),
							},
							Validators: []validator.String{
								enum.FrameworkValidate[awstypes.KxClusterType](),
							},
						},
					},
				},
			},
			"nas1_configuration": schema.SingleNestedBlock{
				Attributes: map[string]schema.Attribute{
					"size": schema.Int64Attribute{
						Required: true,
						PlanModifiers: []planmodifier.Int64{
							int64planmodifier.RequiresReplace(),
						},
						Validators: []validator.Int64{
							int64validator.Between(1200, 33600),
						},
					},
					"type": schema.StringAttribute{
						Required: true,
						PlanModifiers: []planmodifier.String{
							stringplanmodifier.RequiresReplace(),
						},
						Validators: []validator.String{
							enum.FrameworkValidate[awstypes.KxNAS1Type](),
						},
					},
				},
			},
		},
	}
}

func (r *resourceKxVolume) ModifyPlan(ctx context.Context, req resource.ModifyPlanRequest, resp *resource.ModifyPlanResponse) {
	r.SetTagsAll(ctx, req, resp)
}

const (
	ResNameKxVolume     = "Kx Volume"
	kxVolumeIDPartCount = 2
)

type resourceKxVolumeModel struct {
	ARN                   types.String   `tfsdk:"arn"`
	ID                    types.String   `tfsdk:"id"`
	EnvironmentId         types.String   `tfsdk:"environment_id"`
	Name                  types.String   `tfsdk:"name"`
	Description           types.String   `tfsdk:"description"`
	AvailabilityZones     types.List     `tfsdk:"availability_zones"`
	Type                  types.String   `tfsdk:"type"`
	AzMode                types.String   `tfsdk:"az_mode"`
	AttachedClusters      types.List     `tfsdk:"attached_clusters"`
	Nas1Configuration     types.Object   `tfsdk:"nas1_configuration"`
	Status                types.String   `tfsdk:"status"`
	StatusReason          types.String   `tfsdk:"status_reason"`
	CreatedTimestamp      types.String   `tfsdk:"created_timestamp"`
	LastModifiedTimestamp types.String   `tfsdk:"last_modified_timestamp"`
	Tags                  types.Map      `tfsdk:"tags"`
	TagsAll               types.Map      `tfsdk:"tags_all"`
	Timeouts              timeouts.Value `tfsdk:"timeouts"`
}

type attachedClustersModel struct {
	ClusterName   types.String `tfsdk:"cluster_name"`
	ClusterType   types.String `tfsdk:"cluster_type"`
	ClusterStatus types.String `tfsdk:"cluster_status"`
}

type nas1ConfigurationModel struct {
	Size types.Int64  `tfsdk:"size"`
	Type types.String `tfsdk:"type"`
}

func (r *resourceKxVolume) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan resourceKxVolumeModel
	conn := r.Meta().FinSpaceClient(ctx)
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	idParts := []string{
		plan.EnvironmentId.ValueString(),
		plan.Name.ValueString(),
	}
	rId, _ := flex.FlattenResourceId(idParts, kxVolumeIDPartCount, false)
	plan.ID = fwflex.StringValueToFramework(ctx, rId)

	input := &finspace.CreateKxVolumeInput{
		EnvironmentId:       aws.String(plan.EnvironmentId.ValueString()),
		VolumeName:          aws.String(plan.Name.ValueString()),
		AvailabilityZoneIds: fwflex.ExpandFrameworkStringValueList(ctx, plan.AvailabilityZones),
		VolumeType:          awstypes.KxVolumeType(plan.Type.ValueString()),
		AzMode:              awstypes.KxAzMode(plan.AzMode.ValueString()),
		Tags:                getTagsIn(ctx),
		ClientToken:         aws.String(id.UniqueId()),
	}
	if !(plan.Description.IsNull() || plan.Description.IsUnknown()) {
		input.Description = aws.String(plan.Description.ValueString())
	}

	if !(plan.Nas1Configuration.IsNull() || plan.Nas1Configuration.IsUnknown()) {
		input.Nas1Configuration = expandNas1Configuration(ctx, plan.Nas1Configuration, resp.Diagnostics)
	}
	output, err := conn.CreateKxVolume(ctx, input)
	if err != nil {
		resp.Diagnostics.AddError("Error creating volume", err.Error())
		return
	}

	if output == nil || output.VolumeName == nil {
		resp.Diagnostics.AddError("Error creating volume", "Empty output")
		return
	}

	state := plan
	state.EnvironmentId = fwflex.StringToFramework(ctx, output.EnvironmentId)
	state.Name = fwflex.StringToFramework(ctx, output.VolumeName)
	state.AvailabilityZones = fwflex.FlattenFrameworkStringValueList(ctx, output.AvailabilityZoneIds)
	state.Type = fwflex.StringValueToFramework(ctx, output.VolumeType)
	state.AzMode = fwflex.StringValueToFramework(ctx, output.AzMode)
	state.CreatedTimestamp = fwflex.StringValueToFramework(ctx, output.CreatedTimestamp.String())
	state.ARN = fwflex.StringToFramework(ctx, output.VolumeArn)
	if output.Description != nil {
		state.Description = fwflex.StringToFramework(ctx, output.Description)
	}
	if output.Nas1Configuration != nil {
		state.Nas1Configuration = flattenNas1Configuration(ctx, output.Nas1Configuration)
	}
	resp.Diagnostics.Append(fwflex.Flatten(ctx, output, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	createTimeout := r.CreateTimeout(ctx, plan.Timeouts)
	createdVolume, err := waitKxVolumeCreated(ctx, conn, plan.ID.ValueString(), createTimeout)
	if err != nil {
		resp.Diagnostics.AddError("Error waiting for volume creation", err.Error())
		return
	}
	state.Status = fwflex.StringValueToFramework(ctx, createdVolume.Status)
	state.StatusReason = fwflex.StringToFramework(ctx, createdVolume.StatusReason)
	state.LastModifiedTimestamp = fwflex.StringValueToFramework(ctx, createdVolume.LastModifiedTimestamp.String())
	// currently tags are not included in createKxVolume Output
	// setTagsOut(ctx, output.Tags)

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *resourceKxVolume) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state resourceKxVolumeModel
	conn := r.Meta().FinSpaceClient(ctx)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	output, err := FindKxVolumeByID(ctx, conn, state.ID.ValueString())
	if tfresource.NotFound(err) {
		create.LogNotFoundRemoveState(names.FinSpace, create.ErrActionReading, ResNameKxVolume, state.ID.ValueString())
		resp.State.RemoveResource(ctx)
		return
	}
	if err != nil {
		resp.Diagnostics.AddError("Error reading volume", err.Error())
		return
	}

	if output == nil || output.VolumeName == nil {
		resp.Diagnostics.AddError("Error reading volume", "Empty output")
		return
	}

	state.EnvironmentId = fwflex.StringToFramework(ctx, output.EnvironmentId)
	state.Name = fwflex.StringToFramework(ctx, output.VolumeName)
	state.AvailabilityZones = fwflex.FlattenFrameworkStringValueList(ctx, output.AvailabilityZoneIds)
	state.Type = fwflex.StringValueToFramework(ctx, output.VolumeType)
	state.AzMode = fwflex.StringValueToFramework(ctx, output.AzMode)
	state.CreatedTimestamp = fwflex.StringValueToFramework(ctx, output.CreatedTimestamp.String())
	state.Status = fwflex.StringValueToFramework(ctx, output.Status)
	state.StatusReason = fwflex.StringToFramework(ctx, output.StatusReason)
	state.LastModifiedTimestamp = fwflex.StringValueToFramework(ctx, output.LastModifiedTimestamp.String())
	state.ARN = fwflex.StringToFramework(ctx, output.VolumeArn)
	if output.Nas1Configuration != nil {
		state.Nas1Configuration = flattenNas1Configuration(ctx, output.Nas1Configuration)
	}
	if output.AttachedClusters != nil {
		state.AttachedClusters = flattenAttachedClusters(ctx, output.AttachedClusters)
	}
	if output.Description != nil {
		state.Description = fwflex.StringToFramework(ctx, output.Description)
	}
	resp.Diagnostics.Append(fwflex.Flatten(ctx, output, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// currently tags are not included in getKxVolume Output
	// setTagsOut(ctx, output.Tags)

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *resourceKxVolume) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan, state resourceKxVolumeModel
	conn := r.Meta().FinSpaceClient(ctx)
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	updateReq := &finspace.UpdateKxVolumeInput{
		EnvironmentId: aws.String(plan.EnvironmentId.ValueString()),
		VolumeName:    aws.String(plan.Name.ValueString()),
		ClientToken:   aws.String(id.UniqueId()),
	}

	if !(plan.Description.IsNull() || plan.Description.IsUnknown()) {
		updateReq.Description = aws.String(plan.Description.ValueString())
	}
	if !(plan.Nas1Configuration.IsNull() || plan.Nas1Configuration.IsUnknown()) && !plan.Nas1Configuration.Equal(state.Nas1Configuration) {
		updateReq.Nas1Configuration = expandNas1ConfigurationForUpdate(ctx, plan.Nas1Configuration, resp.Diagnostics)
	}
	if resp.Diagnostics.HasError() {
		return
	}

	out, err := conn.UpdateKxVolume(ctx, updateReq)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error updating volume", err.Error(),
		)
		return
	}
	plan.Nas1Configuration = flattenNas1Configuration(ctx, out.Nas1Configuration)
	plan.Status = fwflex.StringValueToFramework(ctx, out.Status)
	plan.StatusReason = fwflex.StringToFramework(ctx, out.StatusReason)
	plan.LastModifiedTimestamp = fwflex.StringValueToFramework(ctx, out.LastModifiedTimestamp.String())
	plan.CreatedTimestamp = fwflex.StringValueToFramework(ctx, out.CreatedTimestamp.String())
	if out.Status == awstypes.KxVolumeStatusUpdating {
		updateTimeout := r.UpdateTimeout(ctx, plan.Timeouts)
		volume, err := waitKxVolumeUpdated(ctx, conn, plan.ID.ValueString(), updateTimeout)
		if err != nil {
			resp.Diagnostics.AddError("Error waiting for volume update", err.Error())
			return
		}
		plan.Status = fwflex.StringValueToFramework(ctx, volume.Status)
		plan.StatusReason = fwflex.StringToFramework(ctx, volume.StatusReason)
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *resourceKxVolume) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state resourceKxVolumeModel
	conn := r.Meta().FinSpaceClient(ctx)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if _, err := conn.DeleteKxVolume(ctx, &finspace.DeleteKxVolumeInput{
		VolumeName:    aws.String(state.Name.ValueString()),
		EnvironmentId: aws.String(state.EnvironmentId.ValueString()),
		ClientToken:   aws.String(id.UniqueId()),
	}); err != nil {
		var nfe *awstypes.ResourceNotFoundException
		if errors.As(err, &nfe) {
			return
		}
		resp.Diagnostics.AddError("Error deleting volume", err.Error())
		return
	}
}

func waitKxVolumeCreated(ctx context.Context, conn *finspace.Client, id string, timeout time.Duration) (*finspace.GetKxVolumeOutput, error) {
	stateConf := &retry.StateChangeConf{
		Pending:                   enum.Slice(awstypes.KxVolumeStatusCreating),
		Target:                    enum.Slice(awstypes.KxVolumeStatusActive),
		Refresh:                   statusKxVolume(ctx, conn, id),
		Timeout:                   timeout,
		NotFoundChecks:            20,
		ContinuousTargetOccurence: 2,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)
	if out, ok := outputRaw.(*finspace.GetKxVolumeOutput); ok {
		return out, err
	}

	return nil, err
}

func waitKxVolumeUpdated(ctx context.Context, conn *finspace.Client, id string, timeout time.Duration) (*finspace.GetKxVolumeOutput, error) {
	stateConf := &retry.StateChangeConf{
		Pending:                   enum.Slice(awstypes.KxVolumeStatusUpdating),
		Target:                    enum.Slice(awstypes.KxVolumeStatusActive),
		Refresh:                   statusKxVolume(ctx, conn, id),
		Timeout:                   timeout,
		NotFoundChecks:            20,
		ContinuousTargetOccurence: 2,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)
	if out, ok := outputRaw.(*finspace.GetKxVolumeOutput); ok {
		return out, err
	}

	return nil, err
}

func statusKxVolume(ctx context.Context, conn *finspace.Client, id string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		out, err := FindKxVolumeByID(ctx, conn, id)
		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return out, string(out.Status), nil
	}
}

func FindKxVolumeByID(ctx context.Context, conn *finspace.Client, id string) (*finspace.GetKxVolumeOutput, error) {
	parts, err := flex.ExpandResourceId(id, kxVolumeIDPartCount, false)
	if err != nil {
		return nil, err
	}

	in := &finspace.GetKxVolumeInput{
		EnvironmentId: aws.String(parts[0]),
		VolumeName:    aws.String(parts[1]),
	}

	out, err := conn.GetKxVolume(ctx, in)
	if err != nil {
		var nfe *awstypes.ResourceNotFoundException
		if errors.As(err, &nfe) {
			return nil, &retry.NotFoundError{
				LastError:   err,
				LastRequest: in,
			}
		}

		return nil, err
	}

	if out == nil || out.VolumeArn == nil {
		return nil, tfresource.NewEmptyResultError(in)
	}

	return out, nil
}

func flattenNas1Configuration(ctx context.Context, apiObject *awstypes.KxNAS1Configuration) types.Object {
	attributeTypes := fwtypes.AttributeTypesMust[nas1ConfigurationModel](ctx)
	attrs := map[string]attr.Value{
		"size": fwflex.Int32ToFramework(ctx, apiObject.Size),
		"type": fwflex.StringValueToFramework(ctx, apiObject.Type),
	}
	return types.ObjectValueMust(attributeTypes, attrs)
}

func expandAttachedClusters(ctx context.Context, tfList []attachedClustersModel) []awstypes.KxAttachedCluster {
	if len(tfList) == 0 {
		return nil
	}

	apiObjects := make([]awstypes.KxAttachedCluster, len(tfList))
	for i, item := range tfList {
		apiObjects[i] = item.expand(ctx)
	}
	return apiObjects
}
func flattenAttachedCluster(ctx context.Context, apiObject awstypes.KxAttachedCluster) *attachedClustersModel {
	return &attachedClustersModel{
		ClusterName:   fwflex.StringToFramework(ctx, apiObject.ClusterName),
		ClusterType:   fwflex.StringValueToFramework(ctx, apiObject.ClusterType),
		ClusterStatus: fwflex.StringValueToFramework(ctx, apiObject.ClusterStatus),
	}
}

func flattenAttachedClusters(ctx context.Context, apiObjects []awstypes.KxAttachedCluster) types.List {
	elemType := fwtypes.NewObjectTypeOf[attachedClustersModel](ctx).ObjectType

	if len(apiObjects) == 0 {
		return types.ListValueMust(elemType, []attr.Value{})
	}

	values := make([]attr.Value, len(apiObjects))
	for i, o := range apiObjects {
		values[i] = flattenAttachedCluster(ctx, o).value(ctx)
	}

	result, _ := types.ListValueFrom(ctx, elemType, values)

	return result
}

func (m *nas1ConfigurationModel) expand(ctx context.Context) *awstypes.KxNAS1Configuration {
	return &awstypes.KxNAS1Configuration{
		Size: aws.Int32(int32(m.Size.ValueInt64())),
		Type: awstypes.KxNAS1Type(m.Type.String()),
	}
}
func (m *nas1ConfigurationModel) value(ctx context.Context) types.Object {
	return fwtypes.NewObjectValueOf[nas1ConfigurationModel](ctx, m).ObjectValue
}

func (m *attachedClustersModel) expand(ctx context.Context) awstypes.KxAttachedCluster {
	return awstypes.KxAttachedCluster{
		ClusterName:   aws.String(m.ClusterName.String()),
		ClusterType:   awstypes.KxClusterType(m.ClusterType.String()),
		ClusterStatus: awstypes.KxClusterStatus(m.ClusterStatus.String()),
	}
}

func expandNas1Configuration(ctx context.Context, object types.Object, diags diag.Diagnostics) *awstypes.KxNAS1Configuration {
	if object.IsNull() {
		return nil
	}

	var config nas1ConfigurationModel
	diags.Append(object.As(ctx, &config, basetypes.ObjectAsOptions{})...)
	if diags.HasError() {
		return nil
	}

	return &awstypes.KxNAS1Configuration{
		Type: awstypes.KxNAS1Type(config.Type.ValueString()),
		Size: aws.Int32(int32(config.Size.ValueInt64())),
	}
}

func expandNas1ConfigurationForUpdate(ctx context.Context, object types.Object, diags diag.Diagnostics) *awstypes.KxNAS1Configuration {
	if object.IsNull() {
		return nil
	}

	var config nas1ConfigurationModel
	diags.Append(object.As(ctx, &config, basetypes.ObjectAsOptions{})...)
	if diags.HasError() {
		return nil
	}

	return &awstypes.KxNAS1Configuration{
		Size: aws.Int32(int32(config.Size.ValueInt64())),
	}
}

func (m *attachedClustersModel) value(ctx context.Context) types.Object {
	return fwtypes.NewObjectValueOf[attachedClustersModel](ctx, m).ObjectValue
}
