// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package datazone

// **PLEASE DELETE THIS AND ALL TIP COMMENTS BEFORE SUBMITTING A PR FOR REVIEW!**
//
// TIP: ==== INTRODUCTION ====
// Thank you for trying the skaff tool!
//
// You have opted to include these helpful comments. They all include "TIP:"
// to help you find and remove them when you're done with them.
//
// While some aspects of this file are customized to your input, the
// scaffold tool does *not* look at the AWS API and ensure it has correct
// function, structure, and variable names. It makes guesses based on
// commonalities. You will need to make significant adjustments.
//
// In other words, as generated, this is a rough outline of the work you will
// need to do. If something doesn't make sense for your situation, get rid of
// it.

import (
	// TIP: ==== IMPORTS ====
	// This is a common set of imports but not customized to your code since
	// your code hasn't been written yet. Make sure you, your IDE, or
	// goimports -w <file> fixes these imports.
	//
	// The provider linter wants your imports to be in two groups: first,
	// standard library (i.e., "fmt" or "strings"), second, everything else.
	//
	// Also, AWS Go SDK v2 may handle nested structures differently than v1,
	// using the services/datazone/types package. If so, you'll
	// need to import types and reference the nested types, e.g., as
	// awstypes.<Type Name>.
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/datazone"
	awstypes "github.com/aws/aws-sdk-go-v2/service/datazone/types"
	"github.com/hashicorp/terraform-plugin-framework-timetypes/timetypes"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	"github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkResource("aws_datazone_asset_type", name="Asset Type")
func newResourceAssetType(_ context.Context) (resource.ResourceWithConfigure, error) {
	return &resourceAssetType{}, nil
}

const (
	ResNameAssetType = "Asset Type"
)

type resourceAssetType struct {
	framework.ResourceWithConfigure
	framework.WithTimeouts
	framework.WithNoUpdate
}

func (r *resourceAssetType) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = "aws_datazone_asset_type"
}
func (r *resourceAssetType) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"created_at": schema.StringAttribute{
				CustomType: timetypes.RFC3339Type{},
				Computed:   true,
			},
			"created_by": schema.StringAttribute{
				Computed: true,
			},
			"domain_identifier": schema.StringAttribute{
				Required: true,
			},
			"forms_output": schema.MapAttribute{
				Computed:    true,
				ElementType: types.ObjectType{AttrTypes: resourceFormEntryOutputData},
			},
			"forms_input": schema.MapAttribute{
				CustomType: fwtypes.NewMapTypeOf[fwtypes.ObjectValueOf[resourceFormEntryInputData]](ctx),
				Required:   true,
			},
			"origin_domain_id": schema.StringAttribute{
				Computed: true,
			},
			"origin_project_id": schema.StringAttribute{
				Computed: true,
			},
			"revision": schema.StringAttribute{
				Computed: true,
			},
			"updated_at": schema.StringAttribute{
				CustomType: timetypes.RFC3339Type{},
				Computed:   true,
			},
			"updated_by": schema.StringAttribute{
				Computed: true,
			},
			names.AttrDescription: schema.StringAttribute{
				Optional: true,
			},
			names.AttrName: schema.StringAttribute{
				Required: true,
			},
			"owning_project_identifier": schema.StringAttribute{
				Required: true,
			},
		},
	}
}

func (r *resourceAssetType) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	conn := r.Meta().DataZoneClient(ctx)

	var plan resourceAssetTypeData
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	in := &datazone.CreateAssetTypeInput{}
	resp.Diagnostics.Append(flex.Expand(ctx, &plan, in)...)
	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.Append(formInputExpander(ctx, plan, in)...)
	if resp.Diagnostics.HasError() {
		return
	}
	in.DomainIdentifier = plan.DomainId.ValueStringPointer()
	in.OwningProjectIdentifier = plan.OwningProjectId.ValueStringPointer()
	out, err := conn.CreateAssetType(ctx, in)
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.DataZone, create.ErrActionCreating, ResNameAssetType, plan.Name.String(), err),
			err.Error(),
		)
		return
	}
	if out == nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.DataZone, create.ErrActionCreating, ResNameAssetType, plan.Name.String(), nil),
			errors.New("empty output").Error(),
		)
		return
	}

	resp.Diagnostics.Append(flex.Flatten(ctx, out, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	plan.OriginDomainId = flex.StringToFrameworkLegacy(ctx, out.OriginDomainId)
	plan.OriginProjectId = flex.StringToFrameworkLegacy(ctx, out.OriginProjectId)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *resourceAssetType) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	conn := r.Meta().DataZoneClient(ctx)

	var state resourceAssetTypeData
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	out, err := findAssetTypeByID(ctx, conn, state.DomainId.ValueString(), state.Name.ValueString(), state.Revision.ValueString())
	if tfresource.NotFound(err) {
		resp.State.RemoveResource(ctx)
		return
	}
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.DataZone, create.ErrActionSetting, ResNameAssetType, state.Name.ValueString(), err),
			err.Error(),
		)
		return
	}

	resp.Diagnostics.Append(flex.Flatten(ctx, out, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	state.OwningProjectId = flex.StringToFramework(ctx, out.OwningProjectId)
	state.OriginDomainId = flex.StringToFrameworkLegacy(ctx, out.OriginDomainId)
	state.OriginProjectId = flex.StringToFrameworkLegacy(ctx, out.OriginProjectId)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}
func (r *resourceAssetType) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	conn := r.Meta().DataZoneClient(ctx)

	var state resourceAssetTypeData
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	in := &datazone.DeleteAssetTypeInput{
		DomainIdentifier: state.DomainId.ValueStringPointer(),
		Identifier:       state.Name.ValueStringPointer(),
	}

	_, err := conn.DeleteAssetType(ctx, in)
	if err != nil {
		if errs.IsA[*awstypes.ResourceNotFoundException](err) {
			return
		}
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.DataZone, create.ErrActionDeleting, ResNameAssetType, state.Name.String(), err),
			err.Error(),
		)
		return
	}
}
func (r *resourceAssetType) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	parts := strings.Split(req.ID, ":")

	if len(parts) != 2 {
		resp.Diagnostics.AddError("Resource Import Invalid ID", fmt.Sprintf(`Unexpected format for import ID (%s), use: "DomainIdentifier:Id"`, req.ID))
	}
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("domain_identifier"), parts[0])...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root(names.AttrNames), parts[1])...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("revision"), parts[1])...)

}

func findAssetTypeByID(ctx context.Context, conn *datazone.Client, domain_id, id, revision string) (*datazone.GetAssetTypeOutput, error) {
	in := &datazone.GetAssetTypeInput{
		DomainIdentifier: aws.String(domain_id),
		Identifier:       aws.String(id),
		Revision:         aws.String(revision),
	}

	out, err := conn.GetAssetType(ctx, in)
	if err != nil {
		if errs.IsA[*awstypes.ResourceNotFoundException](err) {
			return nil, &retry.NotFoundError{
				LastError:   err,
				LastRequest: in,
			}
		}

		return nil, err
	}

	if out == nil {
		return nil, tfresource.NewEmptyResultError(in)
	}

	return out, nil
}

type resourceAssetTypeData struct {
	CreatedAt       timetypes.RFC3339                                                     `tfsdk:"created_at"`
	CreatedBy       types.String                                                          `tfsdk:"created_by"`
	Description     types.String                                                          `tfsdk:"description"`
	DomainId        types.String                                                          `tfsdk:"domain_identifier"`
	FormsOutput     types.Map                                                             `tfsdk:"forms_output"`
	FormsInput      fwtypes.MapValueOf[fwtypes.ObjectValueOf[resourceFormEntryInputData]] `tfsdk:"forms_input"`
	Name            types.String                                                          `tfsdk:"name"`
	OriginDomainId  types.String                                                          `tfsdk:"origin_domain_id"`
	OriginProjectId types.String                                                          `tfsdk:"origin_project_id"`
	OwningProjectId types.String                                                          `tfsdk:"owning_project_identifier"`
	Revision        types.String                                                          `tfsdk:"revision"`
	UpdatedAt       timetypes.RFC3339                                                     `tfsdk:"updated_at"`
	UpdatedBy       types.String                                                          `tfsdk:"updated_by"`
}

type resourceFormEntryOutputStruct struct {
	TypeName     types.String `tfsdk:"type_name"`
	TypeRevision types.String `tfsdk:"type_revision"`
	Required     types.Bool   `tfsdk:"required"`
}

type resourceFormEntryInputData struct {
	TypeIdentifier types.String `tfsdk:"type_identifier"`
	TypeRevision   types.String `tfsdk:"type_revision"`
	Required       types.Bool   `tfsdk:"required"`
}

var resourceFormEntryOutputData = map[string]attr.Type{
	"type_name":     types.StringType,
	"type_revision": types.StringType,
	"Required":      types.BoolType,
}

func formEx2pander(ctx context.Context, c resourceAssetTypeData, r *datazone.CreateAssetTypeInput) diag.Diagnostics {
	var diags diag.Diagnostics

	curr := c.FormsInput
	r.FormsInput = make(map[string]awstypes.FormEntryInput)
	vv, err := curr.ToMapValue(ctx)
	if err != nil {
		return err
	}

	for val, ind := range flex.ExpandFrameworkStringMap(ctx, vv) {
		var v resourceFormEntryInputData
		var t awstypes.FormEntryInput
		fmt.Printf("inddddd: %v\n", ind)
		diags.Append(flex.Expand(ctx, &ind, &v)...)
		flex.Expand(ctx, &ind, v)
		if diags.HasError() {
			fmt.Printf("\"frrr\": %v\n", "firrr")
			return diags
		}
		diags.Append(flex.Expand(ctx, &v, &t)...)
		if diags.HasError() {
			fmt.Printf("\"sccc\": %v\n", "second")
			return diags
		}
		r.FormsInput[val] = t
	}
	return nil
}

func formInputExpander(ctx context.Context, c resourceAssetTypeData, r *datazone.CreateAssetTypeInput) diag.Diagnostics {
	var to map[string]resourceFormEntryInputData
	var diags diag.Diagnostics
	r.FormsInput = make(map[string]awstypes.FormEntryInput)
	diags.Append(c.FormsInput.ElementsAs(ctx, &to, false)...)
	if diags.HasError() {
		return diags
	}
	for key, val := range to {
		fmt.Printf("\"hi\": %v\n", "hi")
		var t awstypes.FormEntryInput
		diags.Append(flex.Expand(ctx, &val, &t)...)
		if diags.HasError() {
			return diags
		}
		r.FormsInput[key] = t
	}
	return diags
}

func formOutputFlattener(ctx context.Context, c *resourceAssetTypeData, r *datazone.CreateAssetTypeOutput) {
	if len(r.FormsOutput) == 0 {
		return
	}

	tfObj := make(map[string]resourceFormEntryOutputStruct)

	for key, val := range r.FormsOutput {
		var t resourceFormEntryOutputStruct
		t.Required = flex.BoolToFramework(ctx, val.Required)
		t.TypeName = flex.StringToFramework(ctx, val.TypeName)
		t.TypeRevision = flex.StringToFramework(ctx, val.TypeRevision)
		tfObj[key] = t
	}

}
