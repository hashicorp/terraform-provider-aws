// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

// DONOTCOPY: Copying old resources spreads bad habits. Use skaff instead.

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
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/fwdiag"
	intflex "github.com/hashicorp/terraform-provider-aws/internal/flex"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	"github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/smerr"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	inttypes "github.com/hashicorp/terraform-provider-aws/internal/types"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkResource("aws_datazone_glossary", name="Glossary")
// @IdentityAttribute("domain_identifier")
// @IdentityAttribute("id")
// @ImportIDHandler("glossaryImportID")
// @Testing(existsType="github.com/aws/aws-sdk-go-v2/service/datazone;datazone.GetGlossaryOutput")
// @Testing(importStateIdAttributes="domain_identifier;id", importStateIdAttributesSep="flex.ResourceIdSeparator")
// @Testing(preIdentityVersion="v6.47.0")
func newGlossaryResource(_ context.Context) (resource.ResourceWithConfigure, error) {
	r := &glossaryResource{}
	return r, nil
}

const (
	ResNameGlossary = "Glossary"
)

type glossaryResource struct {
	framework.ResourceWithModel[glossaryResourceModel]
	framework.WithImportByIdentity
}

func (r *glossaryResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			names.AttrDescription: schema.StringAttribute{
				Optional: true,
				Validators: []validator.String{
					stringvalidator.RegexMatches(regexache.MustCompile(`^[\x21-\x7E]+$`), "must conform to: ^[\x21-\x7E]+$ "),
					stringvalidator.LengthBetween(1, 128),
				},
			},
			"domain_identifier": schema.StringAttribute{
				Required: true,
			},
			names.AttrID: framework.IDAttribute(),
			"owning_project_identifier": schema.StringAttribute{
				Required: true,
				Validators: []validator.String{
					stringvalidator.RegexMatches(regexache.MustCompile(`^[a-zA-Z0-9_-]{1,36}$`), "must conform to: ^[a-zA-Z0-9_-]{1,36}$ "),
				},
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			names.AttrName: schema.StringAttribute{
				Required: true,
				Validators: []validator.String{
					stringvalidator.LengthBetween(1, 256),
				},
			},
			names.AttrStatus: schema.StringAttribute{
				CustomType: fwtypes.StringEnumType[awstypes.GlossaryStatus](),
				Optional:   true,
			},
		},
	}
}

func (r *glossaryResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	conn := r.Meta().DataZoneClient(ctx)

	var plan glossaryResourceModel
	smerr.AddEnrich(ctx, &resp.Diagnostics, req.Plan.Get(ctx, &plan))
	if resp.Diagnostics.HasError() {
		return
	}

	in := &datazone.CreateGlossaryInput{}

	smerr.AddEnrich(ctx, &resp.Diagnostics, flex.Expand(ctx, &plan, in))
	if resp.Diagnostics.HasError() {
		return
	}
	out, err := conn.CreateGlossary(ctx, in)
	if err != nil {
		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, plan.Name.ValueString())
		return
	}
	if out == nil {
		smerr.AddError(ctx, &resp.Diagnostics, errors.New("empty output"), smerr.ID, plan.Name.ValueString())
		return
	}

	smerr.AddEnrich(ctx, &resp.Diagnostics, flex.Flatten(ctx, out, &plan))
	if resp.Diagnostics.HasError() {
		return
	}
	smerr.AddEnrich(ctx, &resp.Diagnostics, resp.State.Set(ctx, plan))
}

func (r *glossaryResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	conn := r.Meta().DataZoneClient(ctx)

	var state glossaryResourceModel
	smerr.AddEnrich(ctx, &resp.Diagnostics, req.State.Get(ctx, &state))
	if resp.Diagnostics.HasError() {
		return
	}

	out, err := findGlossaryByID(ctx, conn, state.Id.ValueString(), state.DomainIdentifier.ValueString())
	if retry.NotFound(err) {
		resp.Diagnostics.Append(fwdiag.NewResourceNotFoundWarningDiagnostic(err))
		resp.State.RemoveResource(ctx)
		return
	}
	if err != nil {
		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, state.Id.ValueString())
		return
	}
	smerr.AddEnrich(ctx, &resp.Diagnostics, flex.Flatten(ctx, out, &state))
	if resp.Diagnostics.HasError() {
		return
	}

	state.OwningProjectIdentifier = flex.StringToFramework(ctx, out.OwningProjectId)

	smerr.AddEnrich(ctx, &resp.Diagnostics, resp.State.Set(ctx, &state))
}

func (r *glossaryResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	conn := r.Meta().DataZoneClient(ctx)

	var plan, state glossaryResourceModel
	smerr.AddEnrich(ctx, &resp.Diagnostics, req.Plan.Get(ctx, &plan))
	smerr.AddEnrich(ctx, &resp.Diagnostics, req.State.Get(ctx, &state))
	if resp.Diagnostics.HasError() {
		return
	}

	if !plan.Description.Equal(state.Description) || !plan.Name.Equal(state.Name) {
		in := &datazone.UpdateGlossaryInput{}
		smerr.AddEnrich(ctx, &resp.Diagnostics, flex.Expand(ctx, &plan, in))
		if resp.Diagnostics.HasError() {
			return
		}
		in.Identifier = plan.Id.ValueStringPointer()
		out, err := conn.UpdateGlossary(ctx, in)

		if err != nil {
			smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, plan.Id.ValueString())
			return
		}
		if out == nil {
			smerr.AddError(ctx, &resp.Diagnostics, errors.New("empty output from glossary update"), smerr.ID, plan.Id.ValueString())
			return
		}
		smerr.AddEnrich(ctx, &resp.Diagnostics, flex.Flatten(ctx, out, &state))
		if resp.Diagnostics.HasError() {
			return
		}
	}

	smerr.AddEnrich(ctx, &resp.Diagnostics, resp.State.Set(ctx, &state))
}

func (r *glossaryResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	conn := r.Meta().DataZoneClient(ctx)
	var state glossaryResourceModel
	smerr.AddEnrich(ctx, &resp.Diagnostics, req.State.Get(ctx, &state))
	if resp.Diagnostics.HasError() {
		return
	}

	in := &datazone.UpdateGlossaryInput{}
	smerr.AddEnrich(ctx, &resp.Diagnostics, flex.Expand(ctx, &state, in))
	if resp.Diagnostics.HasError() {
		return
	}
	in.Identifier = state.Id.ValueStringPointer()
	in.Status = "DISABLED"

	_, err := conn.UpdateGlossary(ctx, in)
	if isResourceMissing(err) {
		return
	}
	if err != nil {
		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, state.Id.ValueString())
		return
	}

	in2 := &datazone.DeleteGlossaryInput{
		DomainIdentifier: state.DomainIdentifier.ValueStringPointer(),
		Identifier:       state.Id.ValueStringPointer(),
	}

	_, err2 := conn.DeleteGlossary(ctx, in2)
	if err2 != nil {
		if isResourceMissing(err2) {
			return
		}
		smerr.AddError(ctx, &resp.Diagnostics, err2, smerr.ID, state.Id.ValueString())
		return
	}
}

func findGlossaryByID(ctx context.Context, conn *datazone.Client, id string, domain_id string) (*datazone.GetGlossaryOutput, error) {
	in := &datazone.GetGlossaryInput{
		Identifier:       aws.String(id),
		DomainIdentifier: aws.String(domain_id),
	}

	out, err := conn.GetGlossary(ctx, in)
	if err != nil {
		if isResourceMissing(err) {
			return nil, &retry.NotFoundError{
				LastError: err,
			}
		}
		return nil, err
	}

	if out == nil {
		return nil, tfresource.NewEmptyResultError()
	}

	return out, nil
}

var (
	_ inttypes.ImportIDParser = glossaryImportID{}
)

type glossaryImportID struct{}

func (glossaryImportID) Parse(id string) (string, map[string]any, error) {
	domainID, glossaryID, found := strings.Cut(id, intflex.ResourceIdSeparator)
	if !found {
		return "", nil, fmt.Errorf("id %q should be in the format <domain-identifier>%s<id>", id, intflex.ResourceIdSeparator)
	}

	result := map[string]any{
		"domain_identifier": domainID,
		names.AttrID:        glossaryID,
	}

	return id, result, nil
}

type glossaryResourceModel struct {
	framework.WithRegionModel
	Description             types.String                                `tfsdk:"description"`
	Name                    types.String                                `tfsdk:"name"`
	OwningProjectIdentifier types.String                                `tfsdk:"owning_project_identifier"`
	Status                  fwtypes.StringEnum[awstypes.GlossaryStatus] `tfsdk:"status"`
	DomainIdentifier        types.String                                `tfsdk:"domain_identifier"`
	Id                      types.String                                `tfsdk:"id"`
}
