// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package finspace

import (
	"context"
	"errors"
	"fmt"
	"github.com/aws/aws-sdk-go-v2/aws/arn"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/finspace"
	awstypes "github.com/aws/aws-sdk-go-v2/service/finspace/types"
	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/boolplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/id"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	fwflex "github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkResource(name="Kx Dataview")
// @Tags(identifierAttribute="arn")
func newResourceKxDataview(context.Context) (resource.ResourceWithConfigure, error) {
	r := &resourceKxDataview{}
	r.SetDefaultCreateTimeout(30 * time.Minute)
	r.SetDefaultReadTimeout(30 * time.Minute)
	r.SetDefaultDeleteTimeout(30 * time.Minute)
	r.SetMigratedFromPluginSDK(true)

	return r, nil
}

type resourceKxDataview struct {
	framework.ResourceWithConfigure
	framework.WithTimeouts
}

func (r *resourceKxDataview) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = "aws_finspace_kx_dataview"
}

func (r *resourceKxDataview) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func (r *resourceKxDataview) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"id":  framework.IDAttribute(),
			"arn": framework.ARNAttributeComputedOnly(),
			"name": schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
				Validators: []validator.String{
					stringvalidator.LengthBetween(3, 63),
				},
			},
			"environment_id": schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
				Validators: []validator.String{
					stringvalidator.LengthBetween(3, 63),
				},
			},
			"database_name": schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
				Validators: []validator.String{
					stringvalidator.LengthBetween(3, 63),
				},
			},
			"description": schema.StringAttribute{
				Optional: true,
				Validators: []validator.String{
					stringvalidator.LengthBetween(1, 1000),
				},
			},
			"auto_update": schema.BoolAttribute{
				Required: true,
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.RequiresReplace(),
				},
			},
			"changeset_id": schema.StringAttribute{
				Optional: true,
			},
			"az_mode": schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
				Validators: []validator.String{
					stringvalidator.OneOf(string(awstypes.KxAzModeSingle), string(awstypes.KxAzModeMulti)),
				},
			},
			"availability_zone_id": schema.StringAttribute{
				Optional: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
				Validators: []validator.String{
					stringvalidator.LengthBetween(1, 32),
				},
			},
			"status": schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			names.AttrTags:    tftags.TagsAttribute(),
			names.AttrTagsAll: tftags.TagsAttributeComputedOnly(),
			"created_timestamp": schema.StringAttribute{
				Computed: true,
			},
			"last_modified_timestamp": schema.StringAttribute{
				Computed: true,
			},
		},
		Blocks: map[string]schema.Block{
			"segment_configurations": schema.ListNestedBlock{
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"db_paths": schema.ListAttribute{
							ElementType: types.StringType,
							Required:    true,
						},
						"volume_name": schema.StringAttribute{
							Required: true,
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

func (r *resourceKxDataview) ModifyPlan(ctx context.Context, req resource.ModifyPlanRequest, resp *resource.ModifyPlanResponse) {
	r.SetTagsAll(ctx, req, resp)
}

const (
	ResNameKxDataview     = "Kx Dataview"
	kxDataviewIdPartCount = 3
)

func (r *resourceKxDataview) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan resourceKxDataviewData
	conn := r.Meta().FinSpaceClient(ctx)
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	idParts := []string{
		plan.EnvironmentId.ValueString(),
		plan.DatabaseName.ValueString(),
		plan.Name.ValueString(),
	}
	rId, err := flex.FlattenResourceId(idParts, kxDataviewIdPartCount, false)
	plan.ID = fwflex.StringValueToFramework(ctx, rId)

	var configs []segmentConfigData
	resp.Diagnostics.Append(plan.SegmentConfigurations.ElementsAs(ctx, &configs, false)...)

	createReq := &finspace.CreateKxDataviewInput{
		DatabaseName:       aws.String(plan.DatabaseName.ValueString()),
		DataviewName:       aws.String(plan.Name.ValueString()),
		EnvironmentId:      aws.String(plan.EnvironmentId.ValueString()),
		AutoUpdate:         plan.AutoUpdate.ValueBool(),
		AzMode:             awstypes.KxAzMode(plan.AzMode.ValueString()),
		AvailabilityZoneId: aws.String(plan.AvailabilityZoneId.ValueString()),
		Tags:               getTagsIn(ctx),
		ClientToken:        aws.String(id.UniqueId()),
	}

	if !(plan.Description.IsNull() || plan.Description.IsUnknown()) {
		createReq.Description = aws.String(plan.Description.ValueString())
	}
	if !plan.SegmentConfigurations.IsNull() && len(plan.SegmentConfigurations.Elements()) > 0 {
		createReq.SegmentConfigurations = expandSegmentConfigurations(ctx, configs)
	}

	dataview, err := conn.CreateKxDataview(ctx, createReq)
	if err != nil {
		resp.Diagnostics.AddError("Error creating dataview", err.Error())
		return
	}
	if dataview == nil || dataview.DataviewName == nil {
		resp.Diagnostics.AddError("Error creating dataview", "empty output")
		return
	}

	state := plan
	state.EnvironmentId = fwflex.StringToFramework(ctx, dataview.EnvironmentId)
	state.DatabaseName = fwflex.StringToFramework(ctx, dataview.DatabaseName)
	state.Name = fwflex.StringToFramework(ctx, dataview.DataviewName)

	state.AutoUpdate = types.BoolValue(dataview.AutoUpdate)
	state.ChangesetId = fwflex.StringToFramework(ctx, dataview.ChangesetId)
	state.AvailabilityZoneId = fwflex.StringToFramework(ctx, dataview.AvailabilityZoneId)
	state.AzMode = fwflex.StringValueToFramework(ctx, dataview.AzMode)
	state.CreatedTimestamp = fwflex.StringValueToFramework(ctx, dataview.CreatedTimestamp.String())
	if dataview.Description != nil {
		state.Description = fwflex.StringToFramework(ctx, dataview.Description)
	}
	if dataview.SegmentConfigurations != nil {
		state.SegmentConfigurations = flattenSegmentConfigurations(ctx, dataview.SegmentConfigurations, &resp.Diagnostics)
	}
	//resp.Diagnostics.Append(state.refreshFromOutput(ctx, dataview)...)
	if resp.Diagnostics.HasError() {
		return
	}
	createTimeout := r.CreateTimeout(ctx, plan.Timeouts)
	createdDv, waitErr := waitKxDataviewCreated(ctx, conn, plan.ID.ValueString(), createTimeout)
	state.LastModifiedTimestamp = fwflex.StringValueToFramework(ctx, createdDv.LastModifiedTimestamp.String())
	if waitErr != nil {
		resp.Diagnostics.AddError("Error waiting for dataview creation", err.Error())
		return
	}
	state.Status = fwflex.StringValueToFramework(ctx, createdDv.Status)
	arnParts := []interface{}{
		aws.ToString(createdDv.EnvironmentId),
		aws.ToString(createdDv.DatabaseName),
		aws.ToString(createdDv.DataviewName),
	}
	dvArn := arn.ARN{
		Partition: r.Meta().Partition,
		Service:   "finspace",
		Region:    r.Meta().Region,
		AccountID: r.Meta().AccountID,
		Resource:  fmt.Sprintf("kxEnvironment/%s/kxDatabase/%s/kxDataview/%s", arnParts...),
	}.String()
	state.ARN = fwflex.StringValueToFramework(ctx, dvArn)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}
func (r *resourceKxDataview) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state resourceKxDataviewData
	conn := r.Meta().FinSpaceClient(ctx)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	dataview, err := FindKxDataviewById(ctx, conn, state.ID.ValueString())
	if tfresource.NotFound(err) {
		create.LogNotFoundRemoveState(names.FinSpace, create.ErrActionReading, ResNameKxDataview, state.ID.ValueString())
		resp.State.RemoveResource(ctx)
		return
	}
	if err != nil {
		resp.Diagnostics.Append(create.DiagErrorFramework(names.FinSpace, create.ErrActionReading, ResNameKxDataview, state.ID.ValueString(), err))
		return
	}
	arnParts := []interface{}{
		aws.ToString(dataview.EnvironmentId),
		aws.ToString(dataview.DatabaseName),
		aws.ToString(dataview.DataviewName),
	}
	dvArn := arn.ARN{
		Partition: r.Meta().Partition,
		Service:   "finspace",
		Region:    r.Meta().Region,
		AccountID: r.Meta().AccountID,
		Resource:  fmt.Sprintf("kxEnvironment/%s/kxDatabase/%s/kxDataview/%s", arnParts...),
	}.String()
	state.ARN = fwflex.StringValueToFramework(ctx, dvArn)
	state.Name = fwflex.StringToFramework(ctx, dataview.DataviewName)
	state.DatabaseName = fwflex.StringToFramework(ctx, dataview.DatabaseName)
	state.EnvironmentId = fwflex.StringToFramework(ctx, dataview.EnvironmentId)
	state.AutoUpdate = types.BoolValue(dataview.AutoUpdate)
	state.AvailabilityZoneId = fwflex.StringToFramework(ctx, dataview.AvailabilityZoneId)
	state.AzMode = fwflex.StringValueToFramework(ctx, dataview.AzMode)
	state.Status = fwflex.StringValueToFramework(ctx, dataview.Status)
	state.CreatedTimestamp = fwflex.StringValueToFramework(ctx, dataview.CreatedTimestamp.String())
	state.ChangesetId = fwflex.StringToFramework(ctx, dataview.ChangesetId)
	state.LastModifiedTimestamp = fwflex.StringValueToFramework(ctx, dataview.LastModifiedTimestamp.String())
	if dataview.Description != nil {
		state.Description = fwflex.StringToFramework(ctx, dataview.Description)
	}
	if dataview.SegmentConfigurations != nil {
		state.SegmentConfigurations = flattenSegmentConfigurations(ctx, dataview.SegmentConfigurations, &resp.Diagnostics)
	}

	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}
func (r *resourceKxDataview) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan, state resourceKxDataviewData
	conn := r.Meta().FinSpaceClient(ctx)
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)

	if resp.Diagnostics.HasError() {
		return
	}

	updateReq := &finspace.UpdateKxDataviewInput{
		EnvironmentId: aws.String(plan.EnvironmentId.ValueString()),
		DatabaseName:  aws.String(plan.DatabaseName.ValueString()),
		DataviewName:  aws.String(plan.Name.ValueString()),
		ClientToken:   aws.String(id.UniqueId()),
	}

	if !(plan.ChangesetId.IsNull() || plan.ChangesetId.IsUnknown()) && !plan.ChangesetId.Equal(plan.ChangesetId) && !state.AutoUpdate.ValueBool() {
		updateReq.ChangesetId = aws.String(plan.ChangesetId.ValueString())
	}
	if !plan.SegmentConfigurations.IsNull() && len(plan.SegmentConfigurations.Elements()) > 0 && !plan.SegmentConfigurations.Equal(state.SegmentConfigurations) {
		var configs []segmentConfigData
		resp.Diagnostics.Append(plan.SegmentConfigurations.ElementsAs(ctx, &configs, false)...)
		updateReq.SegmentConfigurations = expandSegmentConfigurations(ctx, configs)
	}

	if _, err := conn.UpdateKxDataview(ctx, updateReq); err != nil {
		resp.Diagnostics.AddError("Error updating dataview", err.Error())
		return
	}

	updateTimeout := r.UpdateTimeout(ctx, state.Timeouts)
	dataview, err := waitKxDataviewUpdated(ctx, conn, state.ID.ValueString(), updateTimeout)
	if err != nil {
		resp.Diagnostics.AddError("Error waiting for dataview update", err.Error())
		return
	}

	plan.ChangesetId = fwflex.StringToFramework(ctx, dataview.ChangesetId)
	plan.SegmentConfigurations = flattenSegmentConfigurations(ctx, dataview.SegmentConfigurations, &resp.Diagnostics)
	plan.Status = fwflex.StringValueToFramework(ctx, dataview.Status)
	plan.LastModifiedTimestamp = fwflex.StringValueToFramework(ctx, dataview.LastModifiedTimestamp.String())
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}
func (r *resourceKxDataview) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state resourceKxDataviewData
	conn := r.Meta().FinSpaceClient(ctx)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	_, err := conn.DeleteKxDataview(ctx, &finspace.DeleteKxDataviewInput{
		EnvironmentId: aws.String(state.EnvironmentId.ValueString()),
		DatabaseName:  aws.String(state.DatabaseName.ValueString()),
		DataviewName:  aws.String(state.Name.ValueString()),
		ClientToken:   aws.String(id.UniqueId()),
	})

	if err != nil {
		var nfe *awstypes.ResourceNotFoundException
		if errors.As(err, &nfe) {
			return
		}
		resp.Diagnostics.AddError("Error deleting dataview", err.Error())
		return
	}

	deleteTimeout := r.DeleteTimeout(ctx, state.Timeouts)
	if _, err := waitKxDataviewDeleted(ctx, conn, state.ID.ValueString(), deleteTimeout); err != nil && !tfresource.NotFound(err) {
		resp.Diagnostics.AddError("Error waiting for dataview deletion", err.Error())
		return
	}
}
func FindKxDataviewById(ctx context.Context, conn *finspace.Client, id string) (*finspace.GetKxDataviewOutput, error) {
	idParts, err := flex.ExpandResourceId(id, kxDataviewIdPartCount, false)
	if err != nil {
		return nil, err
	}

	in := &finspace.GetKxDataviewInput{
		EnvironmentId: aws.String(idParts[0]),
		DatabaseName:  aws.String(idParts[1]),
		DataviewName:  aws.String(idParts[2]),
	}

	out, err := conn.GetKxDataview(ctx, in)
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

	if out == nil || out.DataviewName == nil {
		return nil, tfresource.NewEmptyResultError(in)
	}
	return out, nil
}

func waitKxDataviewCreated(ctx context.Context, conn *finspace.Client, id string, timeout time.Duration) (*finspace.GetKxDataviewOutput, error) {
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(awstypes.KxDataviewStatusCreating),
		Target:  enum.Slice(awstypes.KxDataviewStatusActive, awstypes.KxDataviewStatusFailed),
		Refresh: statusKxDataview(ctx, conn, id),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)
	if out, ok := outputRaw.(*finspace.GetKxDataviewOutput); ok {
		return out, err
	}
	return nil, err
}

func waitKxDataviewUpdated(ctx context.Context, conn *finspace.Client, id string, timeout time.Duration) (*finspace.GetKxDataviewOutput, error) {
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(awstypes.KxDataviewStatusUpdating),
		Target:  enum.Slice(awstypes.KxDataviewStatusActive),
		Refresh: statusKxDataview(ctx, conn, id),
		Timeout: timeout,
	}
	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if out, ok := outputRaw.(*finspace.GetKxDataviewOutput); ok {
		return out, err
	}
	return nil, err
}

func waitKxDataviewDeleted(ctx context.Context, conn *finspace.Client, id string, timeout time.Duration) (*finspace.GetKxDataviewOutput, error) {
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(awstypes.KxDataviewStatusDeleting),
		Target:  []string{},
		Refresh: statusKxDataview(ctx, conn, id),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)
	if out, ok := outputRaw.(*finspace.GetKxDataviewOutput); ok {
		return out, err
	}

	return nil, err
}

func statusKxDataview(ctx context.Context, conn *finspace.Client, id string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		out, err := FindKxDataviewById(ctx, conn, id)
		if tfresource.NotFound(err) {
			return nil, "", nil
		}
		if err != nil {
			return nil, "", err
		}
		return out, string(out.Status), nil
	}
}

func expandDBPath(tfList []attr.Value) []string {
	if tfList == nil {
		return nil
	}
	var s []string

	for _, v := range tfList {
		s = append(s, v.String())
	}
	return s
}

func expandSegmentConfigurations(ctx context.Context, tfList []segmentConfigData) []awstypes.KxDataviewSegmentConfiguration {
	if tfList == nil {
		return nil
	}
	configs := make([]awstypes.KxDataviewSegmentConfiguration, len(tfList))

	for i, v := range tfList {
		configs[i] = awstypes.KxDataviewSegmentConfiguration{
			VolumeName: aws.String(v.VolumeName.ValueString()),
			DbPaths:    fwflex.ExpandFrameworkStringValueList(ctx, v.DbPaths),
		}
	}

	return configs
}

type segmentConfigData struct {
	VolumeName types.String `tfsdk:"volume_name"`
	DbPaths    types.List   `tfsdk:"db_paths"`
}

func flattenSegmentConfiguration(ctx context.Context, apiObject *awstypes.KxDataviewSegmentConfiguration) *segmentConfigData {
	if apiObject == nil {
		return nil
	}
	//var m segmentConfigData
	//if v := apiObject.VolumeName; aws.ToString(v) != "" {
	//	m.VolumeName = aws.ToString(v)
	//}
	//if v := apiObject.DbPaths; v != nil {
	//	m.DbPaths = v
	//}
	return &segmentConfigData{
		VolumeName: fwflex.StringToFramework(ctx, apiObject.VolumeName),
		DbPaths:    fwflex.FlattenFrameworkStringValueList(ctx, apiObject.DbPaths),
	}
}

func flattenSegmentConfigurations(ctx context.Context, apiObjects []awstypes.KxDataviewSegmentConfiguration, diags *diag.Diagnostics) types.List {
	attributeTypes := fwtypes.AttributeTypesMust[segmentConfigData](ctx)
	attributeTypes["db_paths"] = types.ListType{ElemType: types.StringType}
	elemType := types.ObjectType{AttrTypes: attributeTypes}
	if apiObjects == nil {
		return types.ListNull(elemType)
	}

	//var l = make([]segmentConfigData, len(apiObjects))
	//
	//for _, apiObject := range apiObjects {
	//	l = append(l, *flattenSegmentConfiguration(ctx, &apiObject))
	//}
	attrs := make([]attr.Value, 0, len(apiObjects))
	for _, apiObject := range apiObjects {
		attr := map[string]attr.Value{}
		attr["volume_name"] = fwflex.StringToFramework(ctx, apiObject.VolumeName)
		attr["db_paths"] = fwflex.FlattenFrameworkStringValueList(ctx, apiObject.DbPaths)
		val := types.ObjectValueMust(attributeTypes, attr)
		attrs = append(attrs, val)
	}
	result, d := types.ListValueFrom(ctx, elemType, attrs)
	diags.Append(d...)
	return result
}

type resourceKxDataviewData struct {
	ID                    types.String   `tfsdk:"id"`
	ARN                   types.String   `tfsdk:"arn"`
	EnvironmentId         types.String   `tfsdk:"environment_id"`
	DatabaseName          types.String   `tfsdk:"database_name"`
	Name                  types.String   `tfsdk:"name"`
	Description           types.String   `tfsdk:"description"`
	AutoUpdate            types.Bool     `tfsdk:"auto_update"`
	ChangesetId           types.String   `tfsdk:"changeset_id"`
	AvailabilityZoneId    types.String   `tfsdk:"availability_zone_id"`
	AzMode                types.String   `tfsdk:"az_mode"`
	Status                types.String   `tfsdk:"status"`
	CreatedTimestamp      types.String   `tfsdk:"created_timestamp"`
	LastModifiedTimestamp types.String   `tfsdk:"last_modified_timestamp"`
	SegmentConfigurations types.List     `tfsdk:"segment_configurations"`
	Timeouts              timeouts.Value `tfsdk:"timeouts"`
	Tags                  types.Map      `tfsdk:"tags"`
	TagsAll               types.Map      `tfsdk:"tags_all"`
}
