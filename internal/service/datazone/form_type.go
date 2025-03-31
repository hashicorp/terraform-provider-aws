// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package datazone

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/datazone"
	awstypes "github.com/aws/aws-sdk-go-v2/service/datazone/types"
	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework-timetypes/timetypes"
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/listplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
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

// @FrameworkResource("aws_datazone_form_type", name="Form Type")
func newResourceFormType(_ context.Context) (resource.ResourceWithConfigure, error) {
	r := &resourceFormType{}
	r.SetDefaultCreateTimeout(30 * time.Second)

	return r, nil
}

const (
	ResNameFormType = "Form Type"
)

type resourceFormType struct {
	framework.ResourceWithConfigure
	framework.WithTimeouts
	framework.WithNoUpdate
}

func (r *resourceFormType) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			names.AttrCreatedAt: schema.StringAttribute{
				CustomType: timetypes.RFC3339Type{},
				Computed:   true,
			},
			"created_by": schema.StringAttribute{
				Computed: true,
			},
			names.AttrDescription: schema.StringAttribute{
				Optional: true,
				Validators: []validator.String{
					stringvalidator.LengthBetween(1, 2048),
				},
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"domain_identifier": schema.StringAttribute{
				Required: true,
				Validators: []validator.String{
					stringvalidator.RegexMatches(regexache.MustCompile(`^dzd[-_][a-zA-Z0-9_-]{1,36}$`), "^dzd[-_][a-zA-Z0-9_-]{1,36}$"),
				},
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"imports": schema.ListAttribute{
				CustomType: fwtypes.NewListNestedObjectTypeOf[importData](ctx),
				Computed:   true,
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
			"owning_project_identifier": schema.StringAttribute{
				Required: true,
				Validators: []validator.String{
					stringvalidator.RegexMatches(regexache.MustCompile(`^[a-zA-Z0-9_-]{1,36}$`), "^[a-zA-Z0-9_-]{1,36}$"),
				},
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"revision": schema.StringAttribute{
				Computed: true,
			},
			names.AttrStatus: schema.StringAttribute{
				CustomType: fwtypes.StringEnumType[awstypes.FormTypeStatus](),
				Optional:   true,
				Computed:   true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
					stringplanmodifier.UseStateForUnknown(),
				},
			},
		},
		Blocks: map[string]schema.Block{
			"model": schema.ListNestedBlock{
				CustomType: fwtypes.NewListNestedObjectTypeOf[modelData](ctx),
				Validators: []validator.List{
					listvalidator.SizeAtMost(1),
					listvalidator.IsRequired(),
				},
				PlanModifiers: []planmodifier.List{
					listplanmodifier.RequiresReplace(),
				},
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"smithy": schema.StringAttribute{
							Required: true,
							PlanModifiers: []planmodifier.String{
								stringplanmodifier.RequiresReplace(),
							},
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

func (r *resourceFormType) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	conn := r.Meta().DataZoneClient(ctx)

	var plan resourceFormTypeData
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	in := datazone.CreateFormTypeInput{}
	resp.Diagnostics.Append(flex.Expand(ctx, plan, &in)...)
	if resp.Diagnostics.HasError() {
		return
	}

	out, err := conn.CreateFormType(ctx, &in)
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.DataZone, create.ErrActionCreating, ResNameFormType, plan.Name.String(), err),
			err.Error(),
		)
		return
	}
	if out == nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.DataZone, create.ErrActionCreating, ResNameFormType, plan.Name.String(), nil),
			errors.New("empty output").Error(),
		)
		return
	}

	createTimeout := r.CreateTimeout(ctx, plan.Timeouts)
	outputRaws, err := tfresource.RetryWhenNotFound(ctx, createTimeout, func() (any, error) {
		return findFormTypeByID(ctx, conn, *out.DomainId, *out.Name, *out.Revision)
	})
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.DataZone, create.ErrActionCreating, ResNameFormType, plan.Name.String(), err),
			err.Error(),
		)
		return
	}

	output := outputRaws.(*datazone.GetFormTypeOutput)
	option := flex.WithIgnoredFieldNames([]string{"Model"})
	resp.Diagnostics.Append(flex.Flatten(ctx, output, &plan, option)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *resourceFormType) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	conn := r.Meta().DataZoneClient(ctx)

	var state resourceFormTypeData
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	out, err := findFormTypeByID(ctx, conn, state.DomainIdentifier.ValueString(), state.Name.ValueString(), state.Revision.ValueString())
	if tfresource.NotFound(err) {
		resp.State.RemoveResource(ctx)
		return
	}
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.DataZone, create.ErrActionSetting, ResNameFormType, state.Name.String(), err),
			err.Error(),
		)
		return
	}
	option := flex.WithIgnoredFieldNames([]string{"Model"})

	resp.Diagnostics.Append(flex.Flatten(ctx, out, &state, option)...)
	if resp.Diagnostics.HasError() {
		return
	}

	state.OwningProjectIdentifier = flex.StringToFramework(ctx, out.OwningProjectId)

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *resourceFormType) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	conn := r.Meta().DataZoneClient(ctx)

	var state resourceFormTypeData
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	in := &datazone.DeleteFormTypeInput{
		DomainIdentifier:   state.DomainIdentifier.ValueStringPointer(),
		FormTypeIdentifier: state.Name.ValueStringPointer(),
	}

	_, err := conn.DeleteFormType(ctx, in)
	if err != nil {
		if errs.IsA[*awstypes.ResourceNotFoundException](err) || errs.IsA[*awstypes.AccessDeniedException](err) {
			return
		}
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.DataZone, create.ErrActionDeleting, ResNameFormType, state.Name.String(), err),
			err.Error(),
		)
		return
	}
}
func (r *resourceFormType) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	parts := strings.Split(req.ID, ",")

	if len(parts) != 3 {
		resp.Diagnostics.AddError("Resource Import Invalid ID", `Unexpected format for import ID, use: "DomainIdentifier:Name,Revision"`)
		return
	}

	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("domain_identifier"), parts[0])...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root(names.AttrName), parts[1])...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("revision"), parts[2])...)
}

func findFormTypeByID(ctx context.Context, conn *datazone.Client, domainId string, name string, revision string) (*datazone.GetFormTypeOutput, error) {
	in := &datazone.GetFormTypeInput{
		DomainIdentifier:   aws.String(domainId),
		FormTypeIdentifier: aws.String(name),
		Revision:           aws.String(revision),
	}

	out, err := conn.GetFormType(ctx, in)
	if errs.IsA[*awstypes.ResourceNotFoundException](err) || errs.IsA[*awstypes.AccessDeniedException](err) {
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

func (m modelData) Expand(ctx context.Context) (result any, diags diag.Diagnostics) {
	switch {
	case !m.Smithy.IsNull():

		var r awstypes.ModelMemberSmithy

		r.Value = m.Smithy.ValueString()
		return &r, diags
	}
	return
}

type resourceFormTypeData struct {
	CreatedAt               timetypes.RFC3339                           `tfsdk:"created_at"`
	CreatedBy               types.String                                `tfsdk:"created_by"`
	Description             types.String                                `tfsdk:"description"`
	DomainIdentifier        types.String                                `tfsdk:"domain_identifier"`
	Imports                 fwtypes.ListNestedObjectValueOf[importData] `tfsdk:"imports"`
	Model                   fwtypes.ListNestedObjectValueOf[modelData]  `tfsdk:"model"`
	Name                    types.String                                `tfsdk:"name"`
	OriginDomainId          types.String                                `tfsdk:"origin_domain_id"`
	OriginProjectId         types.String                                `tfsdk:"origin_project_id"`
	OwningProjectIdentifier types.String                                `tfsdk:"owning_project_identifier"`
	Revision                types.String                                `tfsdk:"revision"`
	Status                  fwtypes.StringEnum[awstypes.FormTypeStatus] `tfsdk:"status"`
	Timeouts                timeouts.Value                              `tfsdk:"timeouts"`
}

type modelData struct {
	Smithy types.String `tfsdk:"smithy"`
}

type importData struct {
	Name     types.String `tfsdk:"name"`
	Revision types.String `tfsdk:"revision"`
}
