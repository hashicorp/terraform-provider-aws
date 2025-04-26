// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package datazone

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/datazone"
	awstypes "github.com/aws/aws-sdk-go-v2/service/datazone/types"
	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework-timetypes/timetypes"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/setplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
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
	r := &resourceAssetType{}
	r.SetDefaultCreateTimeout(30 * time.Second)
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

func (r *resourceAssetType) Schema(ctx context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			names.AttrCreatedAt: schema.StringAttribute{
				CustomType: timetypes.RFC3339Type{},
				Computed:   true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
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
				PlanModifiers: []planmodifier.Set{
					setplanmodifier.RequiresReplace(),
				},
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{ // nosemgrep:ci.semgrep.framework.map_block_key-meaningful-names
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
			names.AttrTimeouts: timeouts.Block(ctx, timeouts.Opts{
				Create: true,
			}),
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

	if in.FormsInput == nil {
		in.FormsInput = map[string]awstypes.FormEntryInput{}
	}

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

	createTimeout := r.CreateTimeout(ctx, plan.Timeouts)
	outputRaw, err := tfresource.RetryWhenNotFound(ctx, createTimeout, func() (any, error) {
		return findAssetTypeByID(ctx, conn, plan.DomainIdentifier.ValueString(), plan.Name.ValueString())
	})
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.DataZone, create.ErrActionCreating, ResNameAssetType, plan.Name.String(), err),
			err.Error(),
		)
		return
	}

	output := outputRaw.(*datazone.GetAssetTypeOutput)
	resp.Diagnostics.Append(flex.Flatten(ctx, output, &plan, flex.WithIgnoredFieldNamesAppend("OwningProjectId"))...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *resourceAssetType) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	conn := r.Meta().DataZoneClient(ctx)

	var state resourceAssetTypeData
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	out, err := findAssetTypeByID(ctx, conn, state.DomainIdentifier.ValueString(), state.Name.ValueString())
	if tfresource.NotFound(err) {
		resp.Diagnostics.Append(fwdiag.NewResourceNotFoundWarningDiagnostic(err))
		resp.State.RemoveResource(ctx)
		return
	}

	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.DataZone, create.ErrActionSetting, ResNameAssetType, state.Name.String(), err),
			err.Error(),
		)
		return
	}

	resp.Diagnostics.Append(flex.Flatten(ctx, out, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

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
		resp.Diagnostics.AddError("Resource Import Invalid ID", fmt.Sprintf(`Unexpected format for import ID (%s), use: "domain_identifier,name"`, req.ID))
		return
	}

	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("domain_identifier"), parts[0])...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root(names.AttrName), parts[1])...)
}

func findAssetTypeByID(ctx context.Context, conn *datazone.Client, domainId, id string) (*datazone.GetAssetTypeOutput, error) {
	in := &datazone.GetAssetTypeInput{
		DomainIdentifier: aws.String(domainId),
		Identifier:       aws.String(id),
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
	OwningProjectId  types.String                                               `tfsdk:"owning_project_identifier"`
	Revision         types.String                                               `tfsdk:"revision"`
	Timeouts         timeouts.Value                                             `tfsdk:"timeouts"`
}

type resourceFormEntryInputData struct {
	MapBlockKey    types.String `tfsdk:"map_block_key"`
	TypeIdentifier types.String `tfsdk:"type_identifier"`
	TypeRevision   types.String `tfsdk:"type_revision"`
	Required       types.Bool   `tfsdk:"required"`
}
