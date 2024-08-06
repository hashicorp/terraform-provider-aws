// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package datazone

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/datazone"
	awstypes "github.com/aws/aws-sdk-go-v2/service/datazone/types"
	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework-timetypes/timetypes"
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
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

// @FrameworkResource("aws_datazone_glossary_term", name="Glossary Term")
func newResourceGlossaryTerm(_ context.Context) (resource.ResourceWithConfigure, error) {
	r := &resourceGlossaryTerm{}
	return r, nil
}

const (
	ResNameGlossaryTerm = "Glossary Term"
)

type resourceGlossaryTerm struct {
	framework.ResourceWithConfigure
	framework.WithTimeouts
}

func (r *resourceGlossaryTerm) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = "aws_datazone_glossary_term"
}

func (r *resourceGlossaryTerm) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"domain_identifier": schema.StringAttribute{
				Optional: true,
				Validators: []validator.String{
					stringvalidator.RegexMatches(regexache.MustCompile(`^dzd[-_][a-zA-Z0-9_-]{1,36}$`), "must conform to: ^dzd[-_][a-zA-Z0-9_-]{1,36}$ "),
				},
			},
			"glossary_identifier": schema.StringAttribute{
				Required: true,
				Validators: []validator.String{
					stringvalidator.RegexMatches(regexache.MustCompile(`^[a-zA-Z0-9_-]{1,36}$`), "must conform to: ^[a-zA-Z0-9_-]{1,36}$"),
				},
			},
			names.AttrID: framework.IDAttribute(),
			"long_description": schema.StringAttribute{
				Optional: true,
			},
			names.AttrName: schema.StringAttribute{
				Required: true,
				Validators: []validator.String{
					stringvalidator.LengthBetween(1, 256),
				},
			},
			"short_description": schema.StringAttribute{
				Optional: true,
			},
			names.AttrStatus: schema.StringAttribute{
				CustomType: fwtypes.StringEnumType[awstypes.GlossaryTermStatus](),
				Optional:   true,
			},
			names.AttrCreatedAt: schema.StringAttribute{
				CustomType: timetypes.RFC3339Type{},
				Computed:   true,
			},
			"updated_at": schema.StringAttribute{
				CustomType: timetypes.RFC3339Type{},
				Computed:   true,
			},
			"created_by": schema.StringAttribute{
				Computed: true,
			},
			"updated_by": schema.StringAttribute{
				Computed: true,
			},
		},
		Blocks: map[string]schema.Block{
			"term_relations": schema.ListNestedBlock{
				CustomType: fwtypes.NewListNestedObjectTypeOf[resourceTermRelationsData](ctx),
				Validators: []validator.List{
					listvalidator.SizeAtMost(1),
				},
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"classifies": schema.ListAttribute{
							CustomType:  fwtypes.ListOfStringType,
							ElementType: types.StringType,
							Optional:    true,
							Validators: []validator.List{
								listvalidator.SizeBetween(1, 10),
							},
						},
						"is_a": schema.ListAttribute{
							CustomType:  fwtypes.ListOfStringType,
							ElementType: types.StringType,
							Optional:    true,
							Validators: []validator.List{
								listvalidator.SizeBetween(1, 10),
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

func (r *resourceGlossaryTerm) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	conn := r.Meta().DataZoneClient(ctx)

	var plan resourceGlossaryTermData
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	in := &datazone.CreateGlossaryTermInput{}
	resp.Diagnostics.Append(flex.Expand(ctx, &plan, in)...)
	if resp.Diagnostics.HasError() {
		return
	}
	out, err := conn.CreateGlossaryTerm(ctx, in)
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.DataZone, create.ErrActionCreating, ResNameGlossary, plan.Name.String(), err),
			err.Error(),
		)
		return
	}
	if out == nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.DataZone, create.ErrActionCreating, ResNameGlossary, plan.Name.String(), nil),
			errors.New("empty output").Error(),
		)
		return
	}

	resp.Diagnostics.Append(flex.Flatten(ctx, out, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	createTimeout := r.CreateTimeout(ctx, plan.Timeouts)
	outputRaws, err := tfresource.RetryWhenNotFound(ctx, createTimeout, func() (interface{}, error) {
		return findGlossaryTermByID(ctx, conn, *out.Id, *out.DomainId)
	})
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.DataZone, create.ErrActionCreating, ResNameGlossaryTerm, plan.Name.String(), err),
			err.Error(),
		)
		return
	}
	output := outputRaws.(*datazone.GetGlossaryTermOutput)
	resp.Diagnostics.Append(flex.Flatten(ctx, output, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *resourceGlossaryTerm) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	conn := r.Meta().DataZoneClient(ctx)
	var state resourceGlossaryTermData
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	out, err := findGlossaryTermByID(ctx, conn, state.Id.ValueString(), state.DomainIdentifier.ValueString())
	if tfresource.NotFound(err) {
		resp.Diagnostics.Append(fwdiag.NewResourceNotFoundWarningDiagnostic(err))
		resp.State.RemoveResource(ctx)
		return
	}
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.DataZone, create.ErrActionSetting, ResNameProject, state.Id.String(), err),
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

func (r *resourceGlossaryTerm) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	conn := r.Meta().DataZoneClient(ctx)

	var plan, state resourceGlossaryTermData
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if !plan.ShortDescription.Equal(state.ShortDescription) || !plan.LongDescription.Equal(state.LongDescription) || !plan.Name.Equal(state.Name) || !plan.Status.Equal(state.Status) ||
		!plan.TermRelations.Equal(state.TermRelations) {
		in := &datazone.UpdateGlossaryTermInput{}
		resp.Diagnostics.Append(flex.Expand(ctx, &plan, in)...)
		if resp.Diagnostics.HasError() {
			return
		}
		in.Identifier = plan.Id.ValueStringPointer()
		out, err := conn.UpdateGlossaryTerm(ctx, in)

		if err != nil {
			resp.Diagnostics.AddError(
				create.ProblemStandardMessage(names.DataZone, create.ErrActionUpdating, ResNameProject, plan.Id.String(), err),
				err.Error(),
			)
			return
		}
		if out == nil {
			resp.Diagnostics.AddError(
				create.ProblemStandardMessage(names.DataZone, create.ErrActionUpdating, ResNameProject, plan.Id.String(), nil),
				errors.New("empty output from glossary term update").Error(),
			)
			return
		}
		resp.Diagnostics.Append(flex.Flatten(ctx, out, &state)...)
		if resp.Diagnostics.HasError() {
			return
		}
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *resourceGlossaryTerm) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	conn := r.Meta().DataZoneClient(ctx)
	var state resourceGlossaryTermData
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	option := flex.WithIgnoredFieldNames([]string{"TermRelations"})
	in := &datazone.UpdateGlossaryTermInput{}
	resp.Diagnostics.Append(flex.Expand(ctx, &state, in, option)...)
	if resp.Diagnostics.HasError() {
		return
	}
	in.Status = "DISABLED"
	in.Identifier = state.Id.ValueStringPointer()

	_, err := conn.UpdateGlossaryTerm(ctx, in)
	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return
	}
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.DataZone, create.ErrActionDeleting, ResNameGlossary, state.Id.String(), err),
			err.Error(),
		)
		return
	}

	in2 := &datazone.DeleteGlossaryTermInput{
		DomainIdentifier: state.DomainIdentifier.ValueStringPointer(),
		Identifier:       state.Id.ValueStringPointer(),
	}

	_, err2 := conn.DeleteGlossaryTerm(ctx, in2)
	if err2 != nil {
		if errs.IsA[*awstypes.ResourceNotFoundException](err2) {
			return
		}
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.DataZone, create.ErrActionDeleting, ResNameGlossary, state.Id.String(), err),
			err2.Error(),
		)
		return
	}
}

func (r *resourceGlossaryTerm) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	parts := strings.Split(req.ID, ",")

	if len(parts) != 3 {
		resp.Diagnostics.AddError("Resource Import Invalid ID", fmt.Sprintf(`Unexpected format for import ID (%s), use: "DomainIdentifier,Id,OwningProjectIdentifier"`, req.ID))
		return
	}
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("domain_identifier"), parts[0])...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root(names.AttrID), parts[1])...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("glossary_identifier"), parts[2])...)
}

func findGlossaryTermByID(ctx context.Context, conn *datazone.Client, id string, domain_id string) (*datazone.GetGlossaryTermOutput, error) {
	in := &datazone.GetGlossaryTermInput{
		Identifier:       aws.String(id),
		DomainIdentifier: aws.String(domain_id),
	}

	out, err := conn.GetGlossaryTerm(ctx, in)
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

type resourceGlossaryTermData struct {
	GlossaryIdentifier types.String                                               `tfsdk:"glossary_identifier"`
	DomainIdentifier   types.String                                               `tfsdk:"domain_identifier"`
	LongDescription    types.String                                               `tfsdk:"long_description"`
	Name               types.String                                               `tfsdk:"name"`
	ShortDescription   types.String                                               `tfsdk:"short_description"`
	Status             fwtypes.StringEnum[awstypes.GlossaryTermStatus]            `tfsdk:"status"`
	TermRelations      fwtypes.ListNestedObjectValueOf[resourceTermRelationsData] `tfsdk:"term_relations"`
	Id                 types.String                                               `tfsdk:"id"`

	// from get
	CreatedAt timetypes.RFC3339 `tfsdk:"created_at"`
	CreatedBy types.String      `tfsdk:"created_by"`
	UpdatedAt timetypes.RFC3339 `tfsdk:"updated_at"`
	UpdatedBy types.String      `tfsdk:"updated_by"`

	Timeouts timeouts.Value `tfsdk:"timeouts"`
}

type resourceTermRelationsData struct {
	Classifies fwtypes.ListValueOf[types.String] `tfsdk:"classifies"`
	IsA        fwtypes.ListValueOf[types.String] `tfsdk:"is_a"`
}
