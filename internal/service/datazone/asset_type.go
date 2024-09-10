// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package datazone

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/datazone"
	awstypes "github.com/aws/aws-sdk-go-v2/service/datazone/types"
	"github.com/hashicorp/terraform-plugin-framework-timetypes/timetypes"
	"github.com/hashicorp/terraform-plugin-framework-validators/setvalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/setplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/fwdiag"
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
			names.AttrCreatedAt: schema.StringAttribute{
				CustomType: timetypes.RFC3339Type{},
				Computed:   true,
			},
			"created_by": schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			names.AttrDescription: schema.StringAttribute{
				Optional: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"domain_identifier": schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			names.AttrName: schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"origin_domain_id": schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"origin_project_id": schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"revision": schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"owning_project_identifier": schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
		},
		Blocks: map[string]schema.Block{
			"forms_input": schema.SetNestedBlock{
				CustomType: fwtypes.NewSetNestedObjectTypeOf[resourceFormEntryInputData](ctx),
				Validators: []validator.Set{
					setvalidator.IsRequired(),
				},
				PlanModifiers: []planmodifier.Set{
					setplanmodifier.RequiresReplace(),
				},
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"map_block_key": schema.StringAttribute{
							Required: true,
						},
						"type_identifier": schema.StringAttribute{
							Required: true,
						},
						"type_revision": schema.StringAttribute{
							Required: true,
						},
						"required": schema.BoolAttribute{
							Optional: true,
						},
					},
				},
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
	in := datazone.CreateAssetTypeInput{}
	resp.Diagnostics.Append(flex.Expand(ctx, &plan, &in)...)
	if resp.Diagnostics.HasError() {
		return
	}

	in.OwningProjectIdentifier = plan.OwningProjectId.ValueStringPointer()
	out, err := conn.CreateAssetType(ctx, &in)
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

	//plan.OriginDomainId = flex.StringToFrameworkLegacy(ctx, out.OriginDomainId)
	//plan.OriginProjectId = flex.StringToFrameworkLegacy(ctx, out.OriginProjectId)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *resourceAssetType) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	conn := r.Meta().DataZoneClient(ctx)

	var state resourceAssetTypeData
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	out, err := findAssetTypeByID(ctx, conn, state.DomainIdentifier.ValueString(), state.Name.ValueString(), state.Revision.ValueString())
	if tfresource.NotFound(err) {
		resp.Diagnostics.Append(fwdiag.NewResourceNotFoundWarningDiagnostic(err))
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
		DomainIdentifier: state.DomainIdentifier.ValueStringPointer(),
		Identifier:       state.Name.ValueStringPointer(),
	}

	_, err := conn.DeleteAssetType(ctx, in)
	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return
	}

	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.DataZone, create.ErrActionDeleting, ResNameAssetType, state.Name.String(), err),
			err.Error(),
		)
		return
	}
}
func (r *resourceAssetType) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	parts := strings.Split(req.ID, ",")

	if len(parts) != 2 {
		resp.Diagnostics.AddError("Resource Import Invalid ID", fmt.Sprintf(`Unexpected format for import ID (%s), use: "domain_identifier,id,revision"`, req.ID))
	}
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("domain_identifier"), parts[0])...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root(names.AttrNames), parts[1])...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("revision"), parts[1])...)

}

func findAssetTypeByID(ctx context.Context, conn *datazone.Client, domainId, id, revision string) (*datazone.GetAssetTypeOutput, error) {
	in := &datazone.GetAssetTypeInput{
		DomainIdentifier: aws.String(domainId),
		Identifier:       aws.String(id),
		Revision:         aws.String(revision),
	}

	out, err := conn.GetAssetType(ctx, in)
	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: in,
		}
	}

	if err != nil {
		return nil, err
	}

	if out == nil {
		return nil, tfresource.NewEmptyResultError(in)
	}

	return out, nil
}

type resourceAssetTypeData struct {
	CreatedAt        timetypes.RFC3339                                          `tfsdk:"created_at"`
	CreatedBy        types.String                                               `tfsdk:"created_by"`
	Description      types.String                                               `tfsdk:"description"`
	DomainIdentifier types.String                                               `tfsdk:"domain_identifier"`
	FormsInput       fwtypes.SetNestedObjectValueOf[resourceFormEntryInputData] `tfsdk:"forms_input"`
	Name             types.String                                               `tfsdk:"name"`
	OriginDomainId   types.String                                               `tfsdk:"origin_domain_id"`
	OriginProjectId  types.String                                               `tfsdk:"origin_project_id"`
	OwningProjectId  types.String                                               `tfsdk:"owning_project_identifier"`
	Revision         types.String                                               `tfsdk:"revision"`
}

type resourceFormEntryInputData struct {
	TypeIdentifier types.String `tfsdk:"type_identifier"`
	TypeRevision   types.String `tfsdk:"type_revision"`
	Required       types.Bool   `tfsdk:"required"`
}
