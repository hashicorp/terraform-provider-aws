// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package auditmanager

import (
	"context"
	"errors"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/auditmanager"
	awstypes "github.com/aws/aws-sdk-go-v2/service/auditmanager/types"
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/setvalidator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	"github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkResource(name="Control")
// @Tags(identifierAttribute="arn")
func newResourceControl(_ context.Context) (resource.ResourceWithConfigure, error) {
	return &resourceControl{}, nil
}

const (
	ResNameControl = "Control"
)

type resourceControl struct {
	framework.ResourceWithConfigure
}

func (r *resourceControl) Metadata(_ context.Context, request resource.MetadataRequest, response *resource.MetadataResponse) {
	response.TypeName = "aws_auditmanager_control"
}

func (r *resourceControl) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"action_plan_instructions": schema.StringAttribute{
				Optional: true,
			},
			"action_plan_title": schema.StringAttribute{
				Optional: true,
			},
			names.AttrARN: framework.ARNAttributeComputedOnly(),
			names.AttrDescription: schema.StringAttribute{
				Optional: true,
			},
			names.AttrID: framework.IDAttribute(),
			names.AttrName: schema.StringAttribute{
				Required: true,
			},
			names.AttrTags:    tftags.TagsAttribute(),
			names.AttrTagsAll: tftags.TagsAttributeComputedOnly(),
			"testing_information": schema.StringAttribute{
				Optional: true,
			},
			names.AttrType: schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
		},
		Blocks: map[string]schema.Block{
			"control_mapping_sources": schema.SetNestedBlock{
				Validators: []validator.Set{
					setvalidator.SizeAtLeast(1),
				},
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"source_description": schema.StringAttribute{
							Optional: true,
						},
						"source_frequency": schema.StringAttribute{
							Optional: true,
						},
						"source_id": framework.IDAttribute(),
						"source_name": schema.StringAttribute{
							Required: true,
						},
						"source_set_up_option": schema.StringAttribute{
							Required: true,
						},
						names.AttrSourceType: schema.StringAttribute{
							Required: true,
						},
						"troubleshooting_text": schema.StringAttribute{
							Optional: true,
						},
					},
					Blocks: map[string]schema.Block{
						"source_keyword": schema.ListNestedBlock{
							Validators: []validator.List{
								listvalidator.SizeAtMost(1),
							},
							NestedObject: schema.NestedBlockObject{
								Attributes: map[string]schema.Attribute{
									"keyword_input_type": schema.StringAttribute{
										Required: true,
									},
									"keyword_value": schema.StringAttribute{
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

func (r *resourceControl) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	conn := r.Meta().AuditManagerClient(ctx)

	var plan resourceControlData
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var cms []controlMappingSourcesData
	resp.Diagnostics.Append(plan.ControlMappingSources.ElementsAs(ctx, &cms, false)...)
	if resp.Diagnostics.HasError() {
		return
	}

	cmsInput, d := expandControlMappingSourcesCreate(ctx, cms)
	resp.Diagnostics.Append(d...)
	if resp.Diagnostics.HasError() {
		return
	}

	in := auditmanager.CreateControlInput{
		Name:                  aws.String(plan.Name.ValueString()),
		ControlMappingSources: cmsInput,
		Tags:                  getTagsIn(ctx),
	}
	if !plan.ActionPlanInstructions.IsNull() {
		in.ActionPlanInstructions = aws.String(plan.ActionPlanInstructions.ValueString())
	}
	if !plan.ActionPlanTitle.IsNull() {
		in.ActionPlanTitle = aws.String(plan.ActionPlanTitle.ValueString())
	}
	if !plan.Description.IsNull() {
		in.Description = aws.String(plan.Description.ValueString())
	}
	if !plan.TestingInformation.IsNull() {
		in.TestingInformation = aws.String(plan.TestingInformation.ValueString())
	}

	out, err := conn.CreateControl(ctx, &in)
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.AuditManager, create.ErrActionCreating, ResNameControl, plan.Name.String(), nil),
			err.Error(),
		)
		return
	}
	if out == nil || out.Control == nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.AuditManager, create.ErrActionCreating, ResNameControl, plan.Name.String(), nil),
			errors.New("empty output").Error(),
		)
		return
	}

	state := plan
	resp.Diagnostics.Append(state.refreshFromOutput(ctx, out.Control)...)
	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}

func (r *resourceControl) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	conn := r.Meta().AuditManagerClient(ctx)

	var state resourceControlData
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	out, err := FindControlByID(ctx, conn, state.ID.ValueString())
	if tfresource.NotFound(err) {
		resp.Diagnostics.AddWarning(
			"AWS Resource Not Found During Refresh",
			fmt.Sprintf("Automatically removing from Terraform State instead of returning the error, which may trigger resource recreation. Original Error: %s", err.Error()),
		)
		resp.State.RemoveResource(ctx)
		return
	}
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.AuditManager, create.ErrActionReading, ResNameControl, state.Name.String(), nil),
			err.Error(),
		)
		return
	}

	resp.Diagnostics.Append(state.refreshFromOutput(ctx, out)...)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *resourceControl) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	conn := r.Meta().AuditManagerClient(ctx)

	var plan, state resourceControlData
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if !plan.Name.Equal(state.Name) ||
		!plan.ControlMappingSources.Equal(state.ControlMappingSources) ||
		!plan.ActionPlanInstructions.Equal(state.ActionPlanInstructions) ||
		!plan.ActionPlanTitle.Equal(state.ActionPlanTitle) ||
		!plan.Description.Equal(state.Description) ||
		!plan.TestingInformation.Equal(state.TestingInformation) {
		var cms []controlMappingSourcesData
		resp.Diagnostics.Append(plan.ControlMappingSources.ElementsAs(ctx, &cms, false)...)
		if resp.Diagnostics.HasError() {
			return
		}

		cmsInput, d := expandControlMappingSourcesUpdate(ctx, cms)
		resp.Diagnostics.Append(d...)
		if resp.Diagnostics.HasError() {
			return
		}

		in := &auditmanager.UpdateControlInput{
			ControlId:             aws.String(plan.ID.ValueString()),
			Name:                  aws.String(plan.Name.ValueString()),
			ControlMappingSources: cmsInput,
		}
		if !plan.ActionPlanInstructions.IsNull() {
			in.ActionPlanInstructions = aws.String(plan.ActionPlanInstructions.ValueString())
		}
		if !plan.ActionPlanTitle.IsNull() {
			in.ActionPlanTitle = aws.String(plan.ActionPlanTitle.ValueString())
		}
		if !plan.Description.IsNull() {
			in.Description = aws.String(plan.Description.ValueString())
		}
		if !plan.TestingInformation.IsNull() {
			in.TestingInformation = aws.String(plan.TestingInformation.ValueString())
		}

		out, err := conn.UpdateControl(ctx, in)
		if err != nil {
			resp.Diagnostics.AddError(
				create.ProblemStandardMessage(names.AuditManager, create.ErrActionUpdating, ResNameControl, plan.ID.String(), nil),
				err.Error(),
			)
			return
		}
		if out == nil || out.Control == nil {
			resp.Diagnostics.AddError(
				create.ProblemStandardMessage(names.AuditManager, create.ErrActionUpdating, ResNameControl, plan.ID.String(), nil),
				errors.New("empty output").Error(),
			)
			return
		}
		state.refreshFromOutput(ctx, out.Control)
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *resourceControl) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	conn := r.Meta().AuditManagerClient(ctx)

	var state resourceControlData
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	_, err := conn.DeleteControl(ctx, &auditmanager.DeleteControlInput{
		ControlId: aws.String(state.ID.ValueString()),
	})
	if err != nil {
		var nfe *awstypes.ResourceNotFoundException
		if errors.As(err, &nfe) {
			return
		}
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.AuditManager, create.ErrActionDeleting, ResNameControl, state.ID.String(), nil),
			err.Error(),
		)
	}
}

func (r *resourceControl) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root(names.AttrID), req, resp)
}

func (r *resourceControl) ModifyPlan(ctx context.Context, req resource.ModifyPlanRequest, resp *resource.ModifyPlanResponse) {
	if !req.State.Raw.IsNull() && !req.Plan.Raw.IsNull() {
		var plan resourceControlData
		resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
		if resp.Diagnostics.HasError() {
			return
		}

		var planCms []controlMappingSourcesData
		resp.Diagnostics.Append(plan.ControlMappingSources.ElementsAs(ctx, &planCms, false)...)
		if resp.Diagnostics.HasError() {
			return
		}

		// A net-new control mapping will be missing source_id, so force a replacement
		//
		// Attribute level plan modifiers are applied before resource modifiers, so ID's
		// previously in state should never be unknown.
		for _, item := range planCms {
			if item.SourceID.IsUnknown() {
				resp.RequiresReplace = []path.Path{path.Root("control_mapping_sources")}
			}
		}
	}

	r.SetTagsAll(ctx, req, resp)
}

func FindControlByID(ctx context.Context, conn *auditmanager.Client, id string) (*awstypes.Control, error) {
	in := &auditmanager.GetControlInput{
		ControlId: aws.String(id),
	}
	out, err := conn.GetControl(ctx, in)
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

	if out == nil || out.Control == nil {
		return nil, tfresource.NewEmptyResultError(in)
	}

	return out.Control, nil
}

var (
	controlMappingSourceAttrTypes = map[string]attr.Type{
		"source_description":   types.StringType,
		"source_frequency":     types.StringType,
		"source_id":            types.StringType,
		"source_keyword":       types.ListType{ElemType: types.ObjectType{AttrTypes: sourceKeywordAttrTypes}},
		"source_name":          types.StringType,
		"source_set_up_option": types.StringType,
		names.AttrSourceType:   types.StringType,
		"troubleshooting_text": types.StringType,
	}

	sourceKeywordAttrTypes = map[string]attr.Type{
		"keyword_input_type": types.StringType,
		"keyword_value":      types.StringType,
	}
)

type resourceControlData struct {
	ActionPlanInstructions types.String `tfsdk:"action_plan_instructions"`
	ActionPlanTitle        types.String `tfsdk:"action_plan_title"`
	ARN                    types.String `tfsdk:"arn"`
	ControlMappingSources  types.Set    `tfsdk:"control_mapping_sources"`
	Description            types.String `tfsdk:"description"`
	ID                     types.String `tfsdk:"id"`
	Name                   types.String `tfsdk:"name"`
	Tags                   types.Map    `tfsdk:"tags"`
	TagsAll                types.Map    `tfsdk:"tags_all"`
	TestingInformation     types.String `tfsdk:"testing_information"`
	Type                   types.String `tfsdk:"type"`
}

type controlMappingSourcesData struct {
	SourceDescription   types.String `tfsdk:"source_description"`
	SourceFrequency     types.String `tfsdk:"source_frequency"`
	SourceID            types.String `tfsdk:"source_id"`
	SourceKeyword       types.List   `tfsdk:"source_keyword"`
	SourceName          types.String `tfsdk:"source_name"`
	SourceSetUpOption   types.String `tfsdk:"source_set_up_option"`
	SourceType          types.String `tfsdk:"source_type"`
	TroubleshootingText types.String `tfsdk:"troubleshooting_text"`
}

type sourceKeywordData struct {
	KeywordInputType types.String `tfsdk:"keyword_input_type"`
	KeywordValue     types.String `tfsdk:"keyword_value"`
}

// refreshFromOutput writes state data from an AWS response object
func (rd *resourceControlData) refreshFromOutput(ctx context.Context, out *awstypes.Control) diag.Diagnostics {
	var diags diag.Diagnostics

	if out == nil {
		return diags
	}

	rd.ID = types.StringValue(aws.ToString(out.Id))
	rd.Name = types.StringValue(aws.ToString(out.Name))
	cms, d := flattenControlMappingSources(ctx, out.ControlMappingSources)
	diags.Append(d...)
	rd.ControlMappingSources = cms

	rd.ActionPlanInstructions = flex.StringToFramework(ctx, out.ActionPlanInstructions)
	rd.ActionPlanTitle = flex.StringToFramework(ctx, out.ActionPlanTitle)
	rd.Description = flex.StringToFramework(ctx, out.Description)
	rd.TestingInformation = flex.StringToFramework(ctx, out.TestingInformation)
	rd.ARN = flex.StringToFramework(ctx, out.Arn)
	rd.Type = types.StringValue(string(out.Type))

	setTagsOut(ctx, out.Tags)

	return diags
}

func expandControlMappingSourcesCreate(ctx context.Context, tfList []controlMappingSourcesData) ([]awstypes.CreateControlMappingSource, diag.Diagnostics) {
	var ccms []awstypes.CreateControlMappingSource
	var diags diag.Diagnostics

	for _, item := range tfList {
		new := awstypes.CreateControlMappingSource{
			SourceName:        aws.String(item.SourceName.ValueString()),
			SourceSetUpOption: awstypes.SourceSetUpOption(item.SourceSetUpOption.ValueString()),
			SourceType:        awstypes.SourceType(item.SourceType.ValueString()),
		}

		if !item.SourceDescription.IsNull() {
			new.SourceDescription = aws.String(item.SourceDescription.ValueString())
		}
		if !item.SourceFrequency.IsNull() {
			new.SourceFrequency = awstypes.SourceFrequency(item.SourceFrequency.ValueString())
		}
		if !item.SourceKeyword.IsNull() {
			var sk []sourceKeywordData
			diags.Append(item.SourceKeyword.ElementsAs(ctx, &sk, false)...)
			new.SourceKeyword = expandSourceKeyword(sk)
		}
		if !item.TroubleshootingText.IsNull() {
			new.TroubleshootingText = aws.String(item.TroubleshootingText.ValueString())
		}
		ccms = append(ccms, new)
	}
	return ccms, diags
}

func expandControlMappingSourcesUpdate(ctx context.Context, tfList []controlMappingSourcesData) ([]awstypes.ControlMappingSource, diag.Diagnostics) {
	var cms []awstypes.ControlMappingSource
	var diags diag.Diagnostics

	for _, item := range tfList {
		new := awstypes.ControlMappingSource{
			SourceId:          aws.String(item.SourceID.ValueString()),
			SourceName:        aws.String(item.SourceName.ValueString()),
			SourceSetUpOption: awstypes.SourceSetUpOption(item.SourceSetUpOption.ValueString()),
			SourceType:        awstypes.SourceType(item.SourceType.ValueString()),
		}

		if !item.SourceDescription.IsNull() {
			new.SourceDescription = aws.String(item.SourceDescription.ValueString())
		}
		if !item.SourceFrequency.IsNull() {
			new.SourceFrequency = awstypes.SourceFrequency(item.SourceFrequency.ValueString())
		}
		if !item.SourceKeyword.IsNull() {
			var sk []sourceKeywordData
			diags.Append(item.SourceKeyword.ElementsAs(ctx, &sk, false)...)
			new.SourceKeyword = expandSourceKeyword(sk)
		}
		if !item.TroubleshootingText.IsNull() {
			new.TroubleshootingText = aws.String(item.TroubleshootingText.ValueString())
		}
		cms = append(cms, new)
	}
	return cms, diags
}

func expandSourceKeyword(tfList []sourceKeywordData) *awstypes.SourceKeyword {
	if len(tfList) == 0 {
		return nil
	}
	sk := tfList[0]
	return &awstypes.SourceKeyword{
		KeywordInputType: awstypes.KeywordInputType(sk.KeywordInputType.ValueString()),
		KeywordValue:     aws.String(sk.KeywordValue.ValueString()),
	}
}

func flattenControlMappingSources(ctx context.Context, apiObject []awstypes.ControlMappingSource) (types.Set, diag.Diagnostics) {
	var diags diag.Diagnostics
	elemType := types.ObjectType{AttrTypes: controlMappingSourceAttrTypes}

	elems := []attr.Value{}
	for _, source := range apiObject {
		sk, d := flattenSourceKeyword(ctx, source.SourceKeyword)
		diags.Append(d...)

		obj := map[string]attr.Value{
			"source_description":   flex.StringToFramework(ctx, source.SourceDescription),
			"source_frequency":     flex.StringValueToFramework(ctx, source.SourceFrequency),
			"source_id":            types.StringValue(aws.ToString(source.SourceId)),
			"source_keyword":       sk,
			"source_name":          types.StringValue(aws.ToString(source.SourceName)),
			"source_set_up_option": types.StringValue(string(source.SourceSetUpOption)),
			names.AttrSourceType:   types.StringValue(string(source.SourceType)),
			"troubleshooting_text": flex.StringToFramework(ctx, source.TroubleshootingText),
		}
		objVal, d := types.ObjectValue(controlMappingSourceAttrTypes, obj)
		diags.Append(d...)

		elems = append(elems, objVal)
	}
	setVal, d := types.SetValue(elemType, elems)
	diags.Append(d...)

	return setVal, diags
}

func flattenSourceKeyword(ctx context.Context, apiObject *awstypes.SourceKeyword) (types.List, diag.Diagnostics) {
	var diags diag.Diagnostics
	elemType := types.ObjectType{AttrTypes: sourceKeywordAttrTypes}

	if apiObject == nil {
		return types.ListValueMust(elemType, []attr.Value{}), diags
	}

	obj := map[string]attr.Value{
		"keyword_input_type": flex.StringValueToFramework(ctx, apiObject.KeywordInputType),
		"keyword_value":      types.StringValue(aws.ToString(apiObject.KeywordValue)),
	}
	objVal, d := types.ObjectValue(sourceKeywordAttrTypes, obj)
	diags.Append(d...)

	listVal, d := types.ListValue(elemType, []attr.Value{objVal})
	diags.Append(d...)

	return listVal, diags
}
