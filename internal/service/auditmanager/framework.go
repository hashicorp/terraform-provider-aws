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

// @FrameworkResource(name="Framework")
// @Tags(identifierAttribute="arn")
func newResourceFramework(_ context.Context) (resource.ResourceWithConfigure, error) {
	return &resourceFramework{}, nil
}

const (
	ResNameFramework = "Framework"
)

type resourceFramework struct {
	framework.ResourceWithConfigure
}

func (r *resourceFramework) Metadata(_ context.Context, request resource.MetadataRequest, response *resource.MetadataResponse) {
	response.TypeName = "aws_auditmanager_framework"
}

func (r *resourceFramework) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			names.AttrARN: framework.ARNAttributeComputedOnly(),
			"compliance_type": schema.StringAttribute{
				Optional: true,
			},
			names.AttrDescription: schema.StringAttribute{
				Optional: true,
			},
			"framework_type": schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			names.AttrID: framework.IDAttribute(),
			names.AttrName: schema.StringAttribute{
				Required: true,
			},
			names.AttrTags:    tftags.TagsAttribute(),
			names.AttrTagsAll: tftags.TagsAttributeComputedOnly(),
		},
		Blocks: map[string]schema.Block{
			"control_sets": schema.SetNestedBlock{
				Validators: []validator.Set{
					setvalidator.SizeAtLeast(1),
				},
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						names.AttrID: framework.IDAttribute(),
						names.AttrName: schema.StringAttribute{
							Required: true,
						},
					},
					Blocks: map[string]schema.Block{
						"controls": schema.SetNestedBlock{
							Validators: []validator.Set{
								setvalidator.SizeAtLeast(1),
							},
							NestedObject: schema.NestedBlockObject{
								Attributes: map[string]schema.Attribute{
									names.AttrID: schema.StringAttribute{
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

func (r *resourceFramework) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	conn := r.Meta().AuditManagerClient(ctx)

	var plan resourceFrameworkData
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var cs []frameworkControlSetsData
	resp.Diagnostics.Append(plan.ControlSets.ElementsAs(ctx, &cs, false)...)
	if resp.Diagnostics.HasError() {
		return
	}

	csInput, d := expandFrameworkControlSetsCreate(ctx, cs)
	resp.Diagnostics.Append(d...)
	if resp.Diagnostics.HasError() {
		return
	}

	in := auditmanager.CreateAssessmentFrameworkInput{
		Name:        aws.String(plan.Name.ValueString()),
		ControlSets: csInput,
		Tags:        getTagsIn(ctx),
	}

	if !plan.ComplianceType.IsNull() {
		in.ComplianceType = aws.String(plan.ComplianceType.ValueString())
	}
	if !plan.Description.IsNull() {
		in.Description = aws.String(plan.Description.ValueString())
	}

	out, err := conn.CreateAssessmentFramework(ctx, &in)
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.AuditManager, create.ErrActionCreating, ResNameFramework, plan.Name.String(), nil),
			err.Error(),
		)
		return
	}
	if out == nil || out.Framework == nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.AuditManager, create.ErrActionCreating, ResNameFramework, plan.Name.String(), nil),
			errors.New("empty output").Error(),
		)
		return
	}

	state := plan
	resp.Diagnostics.Append(state.refreshFromOutput(ctx, out.Framework)...)
	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}

func (r *resourceFramework) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	conn := r.Meta().AuditManagerClient(ctx)

	var state resourceFrameworkData
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	out, err := FindFrameworkByID(ctx, conn, state.ID.ValueString())
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
			create.ProblemStandardMessage(names.AuditManager, create.ErrActionReading, ResNameFramework, state.ID.String(), nil),
			err.Error(),
		)
		return
	}

	resp.Diagnostics.Append(state.refreshFromOutput(ctx, out)...)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *resourceFramework) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	conn := r.Meta().AuditManagerClient(ctx)

	var plan, state resourceFrameworkData
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if !plan.Name.Equal(state.Name) ||
		!plan.ControlSets.Equal(state.ControlSets) ||
		!plan.ComplianceType.Equal(state.ComplianceType) ||
		!plan.Description.Equal(state.Description) {
		var cs []frameworkControlSetsData
		resp.Diagnostics.Append(plan.ControlSets.ElementsAs(ctx, &cs, false)...)
		if resp.Diagnostics.HasError() {
			return
		}

		csInput, d := expandFrameworkControlSetsUpdate(ctx, cs)
		resp.Diagnostics.Append(d...)
		if resp.Diagnostics.HasError() {
			return
		}

		in := &auditmanager.UpdateAssessmentFrameworkInput{
			ControlSets: csInput,
			FrameworkId: aws.String(plan.ID.ValueString()),
			Name:        aws.String(plan.Name.ValueString()),
		}

		if !plan.ComplianceType.IsNull() {
			in.ComplianceType = aws.String(plan.ComplianceType.ValueString())
		}
		if !plan.Description.IsNull() {
			in.Description = aws.String(plan.Description.ValueString())
		}

		out, err := conn.UpdateAssessmentFramework(ctx, in)
		if err != nil {
			resp.Diagnostics.AddError(
				create.ProblemStandardMessage(names.AuditManager, create.ErrActionUpdating, ResNameFramework, plan.ID.String(), nil),
				err.Error(),
			)
			return
		}
		if out == nil || out.Framework == nil {
			resp.Diagnostics.AddError(
				create.ProblemStandardMessage(names.AuditManager, create.ErrActionUpdating, ResNameFramework, plan.ID.String(), nil),
				errors.New("empty output").Error(),
			)
			return
		}
		state.refreshFromOutput(ctx, out.Framework)
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *resourceFramework) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	conn := r.Meta().AuditManagerClient(ctx)

	var state resourceFrameworkData
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	_, err := conn.DeleteAssessmentFramework(ctx, &auditmanager.DeleteAssessmentFrameworkInput{
		FrameworkId: aws.String(state.ID.ValueString()),
	})
	if err != nil {
		var nfe *awstypes.ResourceNotFoundException
		if errors.As(err, &nfe) {
			return
		}
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.AuditManager, create.ErrActionDeleting, ResNameFramework, state.ID.String(), nil),
			err.Error(),
		)
	}
}

func (r *resourceFramework) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root(names.AttrID), req, resp)
}

func (r *resourceFramework) ModifyPlan(ctx context.Context, req resource.ModifyPlanRequest, resp *resource.ModifyPlanResponse) {
	if !req.State.Raw.IsNull() && !req.Plan.Raw.IsNull() {
		var plan resourceFrameworkData
		resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
		if resp.Diagnostics.HasError() {
			return
		}

		var planCs []frameworkControlSetsData
		resp.Diagnostics.Append(plan.ControlSets.ElementsAs(ctx, &planCs, false)...)
		if resp.Diagnostics.HasError() {
			return
		}

		// A net-new control set will be missing id, so force a replacement
		//
		// Attribute level plan modifiers are applied before resource modifiers, so ID's
		// previously in state should never be unknown.
		for _, item := range planCs {
			if item.ID.IsUnknown() {
				resp.RequiresReplace = []path.Path{path.Root("control_sets")}
			}
		}
	}

	r.SetTagsAll(ctx, req, resp)
}

func FindFrameworkByID(ctx context.Context, conn *auditmanager.Client, id string) (*awstypes.Framework, error) {
	in := &auditmanager.GetAssessmentFrameworkInput{
		FrameworkId: aws.String(id),
	}
	out, err := conn.GetAssessmentFramework(ctx, in)
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

	if out == nil || out.Framework == nil {
		return nil, tfresource.NewEmptyResultError(in)
	}

	return out.Framework, nil
}

var (
	frameworkControlSetsAttrTypes = map[string]attr.Type{
		"controls":     types.SetType{ElemType: types.ObjectType{AttrTypes: frameworkControlSetsControlsAttrTypes}},
		names.AttrID:   types.StringType,
		names.AttrName: types.StringType,
	}

	frameworkControlSetsControlsAttrTypes = map[string]attr.Type{
		names.AttrID: types.StringType,
	}
)

type resourceFrameworkData struct {
	ARN            types.String `tfsdk:"arn"`
	ComplianceType types.String `tfsdk:"compliance_type"`
	ControlSets    types.Set    `tfsdk:"control_sets"`
	Description    types.String `tfsdk:"description"`
	FrameworkType  types.String `tfsdk:"framework_type"`
	ID             types.String `tfsdk:"id"`
	Name           types.String `tfsdk:"name"`
	Tags           types.Map    `tfsdk:"tags"`
	TagsAll        types.Map    `tfsdk:"tags_all"`
}

type frameworkControlSetsData struct {
	Controls types.Set    `tfsdk:"controls"`
	ID       types.String `tfsdk:"id"`
	Name     types.String `tfsdk:"name"`
}

type frameworkControlSetsControlsData struct {
	ID types.String `tfsdk:"id"`
}

// refreshFromOutput writes state data from an AWS response object
func (rd *resourceFrameworkData) refreshFromOutput(ctx context.Context, out *awstypes.Framework) diag.Diagnostics {
	var diags diag.Diagnostics

	if out == nil {
		return diags
	}

	rd.ID = types.StringValue(aws.ToString(out.Id))
	rd.Name = types.StringValue(aws.ToString(out.Name))
	cs, d := flattenFrameworkControlSets(ctx, out.ControlSets)
	diags.Append(d...)
	rd.ControlSets = cs

	rd.ComplianceType = flex.StringToFramework(ctx, out.ComplianceType)
	rd.Description = flex.StringToFramework(ctx, out.Description)
	rd.FrameworkType = flex.StringValueToFramework(ctx, out.Type)
	rd.ARN = flex.StringToFramework(ctx, out.Arn)

	setTagsOut(ctx, out.Tags)

	return diags
}

func expandFrameworkControlSetsCreate(ctx context.Context, tfList []frameworkControlSetsData) ([]awstypes.CreateAssessmentFrameworkControlSet, diag.Diagnostics) {
	var ccs []awstypes.CreateAssessmentFrameworkControlSet
	var diags diag.Diagnostics

	for _, item := range tfList {
		var controls []frameworkControlSetsControlsData
		diags.Append(item.Controls.ElementsAs(ctx, &controls, false)...)

		new := awstypes.CreateAssessmentFrameworkControlSet{
			Name:     aws.String(item.Name.ValueString()),
			Controls: expandFrameworkControlSetsControls(controls),
		}

		ccs = append(ccs, new)
	}
	return ccs, diags
}

func expandFrameworkControlSetsUpdate(ctx context.Context, tfList []frameworkControlSetsData) ([]awstypes.UpdateAssessmentFrameworkControlSet, diag.Diagnostics) {
	var ucs []awstypes.UpdateAssessmentFrameworkControlSet
	var diags diag.Diagnostics

	for _, item := range tfList {
		var controls []frameworkControlSetsControlsData
		diags.Append(item.Controls.ElementsAs(ctx, &controls, false)...)

		new := awstypes.UpdateAssessmentFrameworkControlSet{
			Controls: expandFrameworkControlSetsControls(controls),
			Id:       aws.String(item.ID.ValueString()),
			Name:     aws.String(item.Name.ValueString()),
		}

		ucs = append(ucs, new)
	}
	return ucs, diags
}

func expandFrameworkControlSetsControls(tfList []frameworkControlSetsControlsData) []awstypes.CreateAssessmentFrameworkControl {
	if len(tfList) == 0 {
		return nil
	}

	var controlsInput []awstypes.CreateAssessmentFrameworkControl
	for _, item := range tfList {
		new := awstypes.CreateAssessmentFrameworkControl{
			Id: aws.String(item.ID.ValueString()),
		}
		controlsInput = append(controlsInput, new)
	}

	return controlsInput
}

func flattenFrameworkControlSets(ctx context.Context, apiObject []awstypes.ControlSet) (types.Set, diag.Diagnostics) {
	var diags diag.Diagnostics
	elemType := types.ObjectType{AttrTypes: frameworkControlSetsAttrTypes}

	elems := []attr.Value{}
	for _, item := range apiObject {
		controls, d := flattenFrameworkControlSetsControls(ctx, item.Controls)
		diags.Append(d...)

		obj := map[string]attr.Value{
			"controls":     controls,
			names.AttrID:   flex.StringToFramework(ctx, item.Id),
			names.AttrName: flex.StringToFramework(ctx, item.Name),
		}
		objVal, d := types.ObjectValue(frameworkControlSetsAttrTypes, obj)
		diags.Append(d...)

		elems = append(elems, objVal)
	}
	setVal, d := types.SetValue(elemType, elems)
	diags.Append(d...)

	return setVal, diags
}

func flattenFrameworkControlSetsControls(ctx context.Context, apiObject []awstypes.Control) (types.Set, diag.Diagnostics) {
	var diags diag.Diagnostics
	elemType := types.ObjectType{AttrTypes: frameworkControlSetsControlsAttrTypes}

	if apiObject == nil {
		return types.SetValueMust(elemType, []attr.Value{}), diags
	}

	elems := []attr.Value{}
	for _, item := range apiObject {
		obj := map[string]attr.Value{
			names.AttrID: flex.StringToFramework(ctx, item.Id),
		}
		objVal, d := types.ObjectValue(frameworkControlSetsControlsAttrTypes, obj)
		diags.Append(d...)

		elems = append(elems, objVal)
	}
	setVal, d := types.SetValue(elemType, elems)
	diags.Append(d...)

	return setVal, diags
}
