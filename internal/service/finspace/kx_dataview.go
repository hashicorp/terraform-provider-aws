// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package finspace

import (
	"context"
	"errors"
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
	"time"
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
			"id": framework.IDAttribute(),
			"arn": schema.StringAttribute{
				Computed: true,
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
			"status": schema.StringAttribute{
				Computed: true,
			},
			names.AttrTags: tftags.TagsAttributeComputedOnly(),
			"created_timestamp": schema.StringAttribute{
				Computed: true,
			},
			"last_modified_timestamp": schema.StringAttribute{
				Computed: true,
			},
		},
		Blocks: map[string]schema.Block{
			"segment_configurations": schema.SetNestedBlock{
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"volume_name": schema.StringAttribute{
							Required: true,
						},
						"db_paths": schema.ListAttribute{
							ElementType: types.StringType,
							Required:    true,
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

//	func ResourceKxDataview() *schema.Resource {
//		return &schema.Resource{
//			CreateWithoutTimeout: resourceKxDataviewCreate,
//			ReadWithoutTimeout:   resourceKxDataviewRead,
//			UpdateWithoutTimeout: resourceKxDataviewUpdate,
//			DeleteWithoutTimeout: resourceKxDataviewDelete,
//
//			Importer: &schema.ResourceImporter{
//				StateContext: schema.ImportStatePassthroughContext,
//			},
//
//			Timeouts: &schema.ResourceTimeout{
//				Create: schema.DefaultTimeout(30 * time.Minute),
//				Update: schema.DefaultTimeout(30 * time.Minute),
//				Delete: schema.DefaultTimeout(30 * time.Minute),
//			},
//			Schema: map[string]*schema.Schema{
//				"arn": {
//					Type:     schema.TypeString,
//					Computed: true,
//				},
//				"auto_update": {
//					Type:     schema.TypeBool,
//					ForceNew: true,
//					Required: true,
//				},
//				"availability_zone_id": {
//					Type:     schema.TypeString,
//					ForceNew: true,
//					Optional: true,
//				},
//				"az_mode": {
//					Type:             schema.TypeString,
//					Required:         true,
//					ForceNew:         true,
//					ValidateDiagFunc: enum.Validate[types.KxAzMode](),
//				},
//				"changeset_id": {
//					Type:     schema.TypeString,
//					Optional: true,
//				},
//				"created_timestamp": {
//					Type:     schema.TypeString,
//					Computed: true,
//				},
//				"database_name": {
//					Type:         schema.TypeString,
//					Required:     true,
//					ForceNew:     true,
//					ValidateFunc: validation.StringLenBetween(3, 63),
//				},
//				"description": {
//					Type:         schema.TypeString,
//					Optional:     true,
//					ValidateFunc: validation.StringLenBetween(1, 1000),
//				},
//				"environment_id": {
//					Type:         schema.TypeString,
//					Required:     true,
//					ForceNew:     true,
//					ValidateFunc: validation.StringLenBetween(3, 63),
//				},
//				"last_modified_timestamp": {
//					Type:     schema.TypeString,
//					Computed: true,
//				},
//				"name": {
//					Type:         schema.TypeString,
//					Required:     true,
//					ForceNew:     true,
//					ValidateFunc: validation.StringLenBetween(3, 63),
//				},
//				"segment_configurations": {
//					Type: schema.TypeList,
//					Elem: &schema.Resource{
//						Schema: map[string]*schema.Schema{
//							"volume_name": {
//								Type:     schema.TypeString,
//								Required: true,
//							},
//							"db_paths": {
//								Type: schema.TypeList,
//								Elem: &schema.Schema{
//									Type: schema.TypeString,
//								},
//								Required: true,
//							},
//						},
//					},
//					Optional: true,
//				},
//				"status": {
//					Type:     schema.TypeString,
//					Computed: true,
//				},
//				names.AttrTags:    tftags.TagsSchema(),
//				names.AttrTagsAll: tftags.TagsSchemaComputed(),
//			},
//			CustomizeDiff: verify.SetTagsDiff,
//		}
//	}
const (
	ResNameKxDataview     = "Kx Dataview"
	kxDataviewIdPartCount = 3
)

//func resourceKxDataviewCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
//	var diags diag.Diagnostics
//	conn := TempFinspaceClient()
//
//	idParts := []string{
//		d.Get("environment_id").(string),
//		d.Get("database_name").(string),
//		d.Get("name").(string),
//	}
//
//	rId, err := flex.FlattenResourceId(idParts, kxDataviewIdPartCount, false)
//	if err != nil {
//		return create.AppendDiagError(diags, names.FinSpace, create.ErrActionFlatteningResourceId, ResNameKxDataview, d.Get("name").(string), err)
//	}
//	d.SetId(rId)
//
//	in := &finspace.CreateKxDataviewInput{
//		DatabaseName:  aws.String(d.Get("database_name").(string)),
//		DataviewName:  aws.String(d.Get("name").(string)),
//		EnvironmentId: aws.String(d.Get("environment_id").(string)),
//		AutoUpdate:    d.Get("auto_update").(bool),
//		AzMode:        types.KxAzMode(d.Get("az_mode").(string)),
//		ClientToken:   aws.String(id.UniqueId()),
//		Tags:          getTagsIn(ctx),
//	}
//
//	if v, ok := d.GetOk("description"); ok {
//		in.Description = aws.String(v.(string))
//	}
//
//	if v, ok := d.GetOk("changeset_id"); ok {
//		in.ChangesetId = aws.String(v.(string))
//	}
//
//	if v, ok := d.GetOk("availability_zone_id"); ok {
//		in.AvailabilityZoneId = aws.String(v.(string))
//	}
//
//	if v, ok := d.GetOk("segment_configurations"); ok && len(v.([]interface{})) > 0 {
//		in.SegmentConfigurations = expandSegmentConfigurations(v.([]interface{}))
//	}
//
//	out, err := conn.CreateKxDataview(ctx, in)
//	if err != nil {
//		return create.AppendDiagError(diags, names.FinSpace, create.ErrActionCreating, ResNameKxDataview, d.Get("name").(string), err)
//	}
//	if out == nil || out.DataviewName == nil {
//		return create.AppendDiagError(diags, names.FinSpace, create.ErrActionCreating, ResNameKxDataview, d.Get("name").(string), errors.New("empty output"))
//	}
//	if _, err := waitKxDataviewCreated(ctx, conn, d.Id(), d.Timeout(schema.TimeoutCreate)); err != nil {
//		return create.AppendDiagError(diags, names.FinSpace, create.ErrActionWaitingForCreation, ResNameKxDataview, d.Get("name").(string), err)
//	}
//
//	return append(diags, resourceKxDataviewRead(ctx, d, meta)...)
//}

func (r *resourceKxDataview) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var state resourceKxDataviewData
	conn := TempFinspaceClient()
	//conn := r.Meta().FinSpaceClient(ctx)
	resp.Diagnostics.Append(req.Plan.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	idParts := []string{
		state.EnvironmentId.ValueString(),
		state.DatabaseName.ValueString(),
		state.Name.ValueString(),
	}
	rId, err := flex.FlattenResourceId(idParts, kxDataviewIdPartCount, false)
	state.ID = fwflex.StringValueToFramework(ctx, rId)

	var configs []segmentConfigData
	resp.Diagnostics.Append(state.SegmentConfigurations.ElementsAs(ctx, &configs, false)...)

	createReq := &finspace.CreateKxDataviewInput{
		DatabaseName:  aws.String(state.DatabaseName.ValueString()),
		DataviewName:  aws.String(state.Name.ValueString()),
		EnvironmentId: aws.String(state.EnvironmentId.ValueString()),
		AutoUpdate:    state.AutoUpdate.ValueBool(),
		AzMode:        awstypes.KxAzMode(state.AzMode.ValueString()),
		Tags:          getTagsIn(ctx),
		ClientToken:   aws.String(id.UniqueId()),
	}

	if !(state.Description.IsNull() || state.Description.IsUnknown()) {
		createReq.Description = aws.String(state.Description.ValueString())
	}
	if !state.SegmentConfigurations.IsNull() && len(state.SegmentConfigurations.Elements()) > 0 {
		createReq.SegmentConfigurations = expandSegmentConfigurations(configs)
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

	createTimeout := r.CreateTimeout(ctx, state.Timeouts)
	if _, err := waitKxDataviewCreated(ctx, conn, state.ID.ValueString(), createTimeout); err != nil {
		resp.Diagnostics.AddError("Error waiting for dataview creation", err.Error())
		return
	}

	state.EnvironmentId = fwflex.StringToFramework(ctx, dataview.EnvironmentId)
	state.DatabaseName = fwflex.StringToFramework(ctx, dataview.DatabaseName)
	state.Name = fwflex.StringToFramework(ctx, dataview.DataviewName)

	state.AutoUpdate = types.BoolValue(dataview.AutoUpdate)
	state.ChangesetId = fwflex.StringToFramework(ctx, dataview.ChangesetId)
	state.AvailabilityZoneId = fwflex.StringToFramework(ctx, dataview.AvailabilityZoneId)
	state.AzMode = fwflex.StringValueToFramework(ctx, dataview.AzMode)
	state.Status = fwflex.StringValueToFramework(ctx, dataview.Status)
	state.CreatedTimestamp = fwflex.StringValueToFramework(ctx, dataview.CreatedTimestamp.String())
	state.LastModifiedTimestamp = fwflex.StringValueToFramework(ctx, dataview.LastModifiedTimestamp.String())
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
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

//func resourceKxDataviewRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
//	var diags diag.Diagnostics
//	conn := TempFinspaceClient()
//
//	out, err := FindKxDataviewById(ctx, conn, d.Id())
//	if !d.IsNewResource() && tfresource.NotFound(err) {
//		log.Printf("[WARN] FinSpace KxDataview (%s) not found, removing from state", d.Id())
//		d.SetId("")
//		return diags
//	}
//
//	if err != nil {
//		return create.AppendDiagError(diags, names.FinSpace, create.ErrActionReading, ResNameKxDataview, d.Id(), err)
//	}
//	d.Set("name", out.DataviewName)
//	d.Set("description", out.Description)
//	d.Set("auto_update", out.AutoUpdate)
//	d.Set("changeset_id", out.ChangesetId)
//	d.Set("availability_zone_id", out.AvailabilityZoneId)
//	d.Set("status", out.Status)
//	d.Set("created_timestamp", out.CreatedTimestamp.String())
//	d.Set("last_modified_timestamp", out.LastModifiedTimestamp.String())
//	d.Set("database_name", out.DatabaseName)
//	d.Set("environment_id", out.EnvironmentId)
//	d.Set("az_mode", out.AzMode)
//	if err := d.Set("segment_configurations", flattenSegmentConfigurations(out.SegmentConfigurations)); err != nil {
//		return create.AppendDiagError(diags, names.FinSpace, create.ErrActionReading, ResNameKxDataview, d.Id(), err)
//	}
//
//	return diags
//}

func (r *resourceKxDataview) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state resourceKxDataviewData
	conn := TempFinspaceClient()
	//conn := r.Meta().FinSpaceClient(ctx)
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

	state.Name = fwflex.StringToFramework(ctx, dataview.DataviewName)
	state.DatabaseName = fwflex.StringToFramework(ctx, dataview.DatabaseName)
	state.EnvironmentId = fwflex.StringToFramework(ctx, dataview.EnvironmentId)
	state.AutoUpdate = types.BoolValue(dataview.AutoUpdate)
	state.AvailabilityZoneId = fwflex.StringToFramework(ctx, dataview.AvailabilityZoneId)
	state.AzMode = fwflex.StringValueToFramework(ctx, dataview.AzMode)
	state.Status = fwflex.StringValueToFramework(ctx, dataview.Status)
	state.CreatedTimestamp = fwflex.StringValueToFramework(ctx, dataview.CreatedTimestamp.String())
	state.LastModifiedTimestamp = fwflex.StringValueToFramework(ctx, dataview.LastModifiedTimestamp.String())
	state.ChangesetId = fwflex.StringToFramework(ctx, dataview.ChangesetId)
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

//func resourceKxDataviewUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
//	var diags diag.Diagnostics
//	conn := TempFinspaceClient()
//	in := &finspace.UpdateKxDataviewInput{
//		EnvironmentId: aws.String(d.Get("environment_id").(string)),
//		DatabaseName:  aws.String(d.Get("database_name").(string)),
//		DataviewName:  aws.String(d.Get("name").(string)),
//		ClientToken:   aws.String(id.UniqueId()),
//	}
//
//	if v, ok := d.GetOk("changeset_id"); ok && d.HasChange("changeset_id") && !d.Get("auto_update").(bool) {
//		in.ChangesetId = aws.String(v.(string))
//	}
//
//	if v, ok := d.GetOk("segment_configurations"); ok && len(v.([]interface{})) > 0 && d.HasChange("segment_configurations") {
//		in.SegmentConfigurations = expandSegmentConfigurations(v.([]interface{}))
//	}
//
//	if _, err := conn.UpdateKxDataview(ctx, in); err != nil {
//		return create.AppendDiagError(diags, names.FinSpace, create.ErrActionUpdating, ResNameKxDataview, d.Get("name").(string), err)
//	}
//
//	if _, err := waitKxDataviewUpdated(ctx, conn, d.Id(), d.Timeout(schema.TimeoutUpdate)); err != nil {
//		return create.AppendDiagError(diags, names.FinSpace, create.ErrActionWaitingForUpdate, ResNameKxDataview, d.Get("name").(string), err)
//	}
//
//	return append(diags, resourceKxDataviewRead(ctx, d, meta)...)
//}

func (r *resourceKxDataview) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan, state resourceKxDataviewData
	conn := TempFinspaceClient()
	//conn := r.Meta().FinSpaceClient(ctx)
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)

	if resp.Diagnostics.HasError() {
		return
	}

	updateReq := &finspace.UpdateKxDataviewInput{
		EnvironmentId: aws.String(state.EnvironmentId.ValueString()),
		DatabaseName:  aws.String(state.DatabaseName.ValueString()),
		DataviewName:  aws.String(state.Name.ValueString()),
		ClientToken:   aws.String(id.UniqueId()),
	}

	if !plan.ChangesetId.IsNull() && !plan.ChangesetId.IsUnknown() && !plan.ChangesetId.Equal(state.ChangesetId) && !state.AutoUpdate.ValueBool() {
		updateReq.ChangesetId = aws.String(state.ChangesetId.ValueString())
	}
	if !plan.SegmentConfigurations.IsNull() && len(plan.SegmentConfigurations.Elements()) > 0 && !plan.SegmentConfigurations.Equal(state.SegmentConfigurations) {
		var configs []segmentConfigData
		resp.Diagnostics.Append(state.SegmentConfigurations.ElementsAs(ctx, &configs, false)...)
		updateReq.SegmentConfigurations = expandSegmentConfigurations(configs)
	}

	if _, err := conn.UpdateKxDataview(ctx, updateReq); err != nil {
		resp.Diagnostics.AddError("Error updating dataview", err.Error())
		return
	}

	updateTimeout := r.UpdateTimeout(ctx, state.Timeouts)
	if _, err := waitKxDataviewUpdated(ctx, conn, state.ID.ValueString(), updateTimeout); err != nil {
		resp.Diagnostics.AddError("Error waiting for dataview update", err.Error())
		return
	}

	dataview, err := FindKxDataviewById(ctx, conn, state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Error reading dataview", err.Error())
		return
	}

	state.ChangesetId = fwflex.StringToFramework(ctx, dataview.ChangesetId)
	state.SegmentConfigurations = flattenSegmentConfigurations(ctx, dataview.SegmentConfigurations, &resp.Diagnostics)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

//func resourceKxDataviewDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
//	var diags diag.Diagnostics
//	conn := TempFinspaceClient()
//
//	_, err := conn.DeleteKxDataview(ctx, &finspace.DeleteKxDataviewInput{
//		EnvironmentId: aws.String(d.Get("environment_id").(string)),
//		DatabaseName:  aws.String(d.Get("database_name").(string)),
//		DataviewName:  aws.String(d.Get("name").(string)),
//		ClientToken:   aws.String(id.UniqueId()),
//	})
//
//	if err != nil {
//		var nfe *types.ResourceNotFoundException
//		if errors.As(err, &nfe) {
//			return diags
//		}
//		return create.AppendDiagError(diags, names.FinSpace, create.ErrActionDeleting, ResNameKxDataview, d.Get("name").(string), err)
//	}
//
//	if _, err := waitKxDataviewDeleted(ctx, conn, d.Id(), d.Timeout(schema.TimeoutDelete)); err != nil && !tfresource.NotFound(err) {
//		return create.AppendDiagError(diags, names.FinSpace, create.ErrActionWaitingForDeletion, ResNameKxDataview, d.Id(), err)
//	}
//	return diags
//}

func (r *resourceKxDataview) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state resourceKxDataviewData
	conn := TempFinspaceClient()
	//conn := r.Meta().FinSpaceClient(ctx)
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
		Pending:                   enum.Slice(awstypes.KxDataviewStatusCreating),
		Target:                    enum.Slice(awstypes.KxDataviewStatusActive),
		Refresh:                   statusKxDataview(ctx, conn, id),
		Timeout:                   timeout,
		NotFoundChecks:            20,
		ContinuousTargetOccurence: 2,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)
	if out, ok := outputRaw.(*finspace.GetKxDataviewOutput); ok {
		return out, err
	}
	return nil, err
}

func waitKxDataviewUpdated(ctx context.Context, conn *finspace.Client, id string, timeout time.Duration) (*finspace.GetKxDataviewOutput, error) {
	stateConf := &retry.StateChangeConf{
		Pending:                   enum.Slice(awstypes.KxDataviewStatusUpdating),
		Target:                    enum.Slice(awstypes.KxDataviewStatusActive),
		Refresh:                   statusKxDataview(ctx, conn, id),
		Timeout:                   timeout,
		NotFoundChecks:            20,
		ContinuousTargetOccurence: 2,
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

func expandSegmentConfigurations(tfList []segmentConfigData) []awstypes.KxDataviewSegmentConfiguration {
	if tfList == nil {
		return nil
	}
	configs := make([]awstypes.KxDataviewSegmentConfiguration, len(tfList))

	for i, v := range tfList {
		configs[i] = v.expand()
	}

	return configs
}

type segmentConfigData struct {
	VolumeName types.String `tfsdk:"volume_name"`
	DbPaths    types.List   `tfsdk:"db_paths"`
}

func (s *segmentConfigData) expand() awstypes.KxDataviewSegmentConfiguration {
	return awstypes.KxDataviewSegmentConfiguration{
		VolumeName: aws.String(s.VolumeName.ValueString()),
		DbPaths:    expandDBPath(s.DbPaths.Elements()),
	}
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
}

func (m *resourceKxDataviewData) refreshFromOutput(ctx context.Context, out *finspace.CreateKxDataviewOutput) diag.Diagnostics {
	var diags diag.Diagnostics
	if out == nil {
		return diags
	}
	m.EnvironmentId = fwflex.StringToFramework(ctx, out.EnvironmentId)
	m.DatabaseName = fwflex.StringToFramework(ctx, out.DatabaseName)
	m.Name = fwflex.StringToFramework(ctx, out.DataviewName)

	m.AutoUpdate = types.BoolValue(out.AutoUpdate)
	m.ChangesetId = fwflex.StringToFramework(ctx, out.ChangesetId)
	m.AvailabilityZoneId = fwflex.StringToFramework(ctx, out.AvailabilityZoneId)
	m.AzMode = fwflex.StringValueToFramework(ctx, out.AzMode)
	m.Status = fwflex.StringValueToFramework(ctx, out.Status)
	m.CreatedTimestamp = fwflex.StringValueToFramework(ctx, out.CreatedTimestamp.String())
	m.LastModifiedTimestamp = fwflex.StringValueToFramework(ctx, out.LastModifiedTimestamp.String())
	if out.Description != nil {
		m.Description = fwflex.StringToFramework(ctx, out.Description)
	}
	if out.SegmentConfigurations != nil {
		m.SegmentConfigurations = flattenSegmentConfigurations(ctx, out.SegmentConfigurations, &diags)
	}

	return diags
}
