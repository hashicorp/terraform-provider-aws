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

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/datazone"
	awstypes "github.com/aws/aws-sdk-go-v2/service/datazone/types"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
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

// TIP: ==== FILE STRUCTURE ====
// All resources should follow this basic outline. Improve this resource's
// maintainability by sticking to it.
//
// 1. Package declaration
// 2. Imports
// 3. Main resource struct with schema method
// 4. Create, read, update, delete methods (in that order)
// 5. Other functions (flatteners, expanders, waiters, finders, etc.)

// Function annotations are used for resource registration to the Provider. DO NOT EDIT.
// @FrameworkResource("aws_datazone_glossary", name="Glossary")
func newResourceGlossary(_ context.Context) (resource.ResourceWithConfigure, error) {
	r := &resourceGlossary{}

	return r, nil
}

const (
	ResNameGlossary = "Glossary"
)

type resourceGlossary struct {
	framework.ResourceWithConfigure
}

func (r *resourceGlossary) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = "aws_datazone_glossary"
}

func (r *resourceGlossary) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			names.AttrDescription: schema.StringAttribute{
				Optional: true,
				Validators: []validator.String{
					stringvalidator.RegexMatches(regexache.MustCompile(`^[\x21-\x7E]+$`), "must conform to: ^[\x21-\x7E]+$ "),
					stringvalidator.LengthBetween(1, 128),
				},
			},
			names.AttrName: schema.StringAttribute{
				Required: true,
				Validators: []validator.String{
					stringvalidator.LengthBetween(1, 256),
				},
			},
			"owning_project_identifier": schema.StringAttribute{
				Required: true,
				Validators: []validator.String{
					stringvalidator.RegexMatches(regexache.MustCompile(`^[a-zA-Z0-9_-]{1,36}$`), "must conform to: ^[a-zA-Z0-9_-]{1,36}$ "),
				},
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			names.AttrStatus: schema.StringAttribute{
				CustomType: fwtypes.StringEnumType[awstypes.GlossaryStatus](),
				Optional:   true,
			},
			"domain_id": schema.StringAttribute{
				Required: true,
			},
			names.AttrID: schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
		},
	}
}

func (r *resourceGlossary) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	conn := r.Meta().DataZoneClient(ctx)

	var plan glossaryData
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	in := &datazone.CreateGlossaryInput{}

	resp.Diagnostics.Append(flex.Expand(ctx, &plan, in)...)
	if resp.Diagnostics.HasError() {
		return
	}
	in.DomainIdentifier = plan.DomainId.ValueStringPointer()
	out, err := conn.CreateGlossary(ctx, in)
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
	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *resourceGlossary) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	conn := r.Meta().DataZoneClient(ctx)

	var state glossaryData
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	out, err := findGlossaryByID(ctx, conn, state.Id.ValueString(), state.DomainId.ValueString())
	if tfresource.NotFound(err) {
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

func (r *resourceGlossary) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	conn := r.Meta().DataZoneClient(ctx)

	var plan, state glossaryData
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if !plan.Description.Equal(state.Description) || !plan.Name.Equal(state.Name) || !plan.Description.Equal(state.Description) {
		in := &datazone.UpdateGlossaryInput{}
		resp.Diagnostics.Append(flex.Expand(ctx, &plan, in)...)
		if resp.Diagnostics.HasError() {
			return
		}
		in.DomainIdentifier = plan.DomainId.ValueStringPointer()
		in.Identifier = plan.Id.ValueStringPointer()
		out, err := conn.UpdateGlossary(ctx, in)

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
				errors.New("empty output from glossary update").Error(),
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

func (r *resourceGlossary) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	conn := r.Meta().DataZoneClient(ctx)
	var state glossaryData
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	in := &datazone.UpdateGlossaryInput{}
	resp.Diagnostics.Append(flex.Expand(ctx, &state, in)...)
	if resp.Diagnostics.HasError() {
		return
	}
	in.DomainIdentifier = state.DomainId.ValueStringPointer()
	in.Identifier = state.Id.ValueStringPointer()
	in.Status = "DISABLED"

	_, err := conn.UpdateGlossary(ctx, in)
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

	in2 := &datazone.DeleteGlossaryInput{
		DomainIdentifier: state.DomainId.ValueStringPointer(),
		Identifier:       state.Id.ValueStringPointer(),
	}

	_, err2 := conn.DeleteGlossary(ctx, in2)
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

func (r *resourceGlossary) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	parts := strings.Split(req.ID, ",")

	if len(parts) != 3 {
		resp.Diagnostics.AddError("Resource Import Invalid ID", fmt.Sprintf(`Unexpected format for import ID (%s), use: "DomainIdentifier,Id,OwningProjectIdentifier"`, req.ID))
	}
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("domain_id"), parts[0])...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root(names.AttrID), parts[1])...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("owning_project_identifier"), parts[2])...)
}

func findGlossaryByID(ctx context.Context, conn *datazone.Client, id string, domain_id string) (*datazone.GetGlossaryOutput, error) {
	in := &datazone.GetGlossaryInput{
		Identifier:       aws.String(id),
		DomainIdentifier: aws.String(domain_id),
	}

	out, err := conn.GetGlossary(ctx, in)
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

type glossaryData struct {
	Description             types.String                                `tfsdk:"description"`
	Name                    types.String                                `tfsdk:"name"`
	OwningProjectIdentifier types.String                                `tfsdk:"owning_project_identifier"`
	Status                  fwtypes.StringEnum[awstypes.GlossaryStatus] `tfsdk:"status"`
	DomainId                types.String                                `tfsdk:"domain_id"`
	Id                      types.String                                `tfsdk:"id"`
}
