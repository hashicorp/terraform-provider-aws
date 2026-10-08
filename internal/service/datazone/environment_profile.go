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
	"github.com/hashicorp/terraform-plugin-framework-timetypes/timetypes"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	intflex "github.com/hashicorp/terraform-provider-aws/internal/flex"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	fwflex "github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	fwvalidators "github.com/hashicorp/terraform-provider-aws/internal/framework/validators"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/smerr"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	inttypes "github.com/hashicorp/terraform-provider-aws/internal/types"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkResource("aws_datazone_environment_profile", name="Environment Profile")
// @IdentityAttribute("domain_identifier")
// @IdentityAttribute("id")
// @ImportIDHandler("environmentProfileImportID")
// @Testing(existsType="github.com/aws/aws-sdk-go-v2/service/datazone;datazone.GetEnvironmentProfileOutput")
// @Testing(importStateIdAttributes="id;domain_identifier", importStateIdAttributesSep="flex.ResourceIdSeparator")
// @Testing(preIdentityVersion="v6.47.0")
func newEnvironmentProfileResource(_ context.Context) (resource.ResourceWithConfigure, error) {
	return &environmentProfileResource{}, nil
}

const (
	ResNameEnvironmentProfile = "Environment Profile"
)

type environmentProfileResource struct {
	framework.ResourceWithModel[environmentProfileResourceModel]
	framework.WithImportByIdentity
}

func (r *environmentProfileResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			names.AttrAWSAccountID: schema.StringAttribute{
				Optional: true,
				Computed: true,
				Validators: []validator.String{
					stringvalidator.RegexMatches(regexache.MustCompile("^\\d{12}$"), "must match ^\\d{12}$"),
				},
			},
			"aws_account_region": schema.StringAttribute{
				Required: true,
				Validators: []validator.String{
					fwvalidators.AWSRegion(),
				},
			},
			names.AttrCreatedAt: schema.StringAttribute{
				CustomType: timetypes.RFC3339Type{},
				Computed:   true,
			},
			"created_by": schema.StringAttribute{
				Computed: true,
			},
			names.AttrDescription: schema.StringAttribute{
				Optional: true,
				Computed: true,
				Validators: []validator.String{
					stringvalidator.LengthBetween(0, 2048),
				},
			},
			"domain_identifier": schema.StringAttribute{
				Required: true,
			},
			"environment_blueprint_identifier": schema.StringAttribute{
				Required: true,
				Validators: []validator.String{
					stringvalidator.RegexMatches(regexache.MustCompile("^[a-zA-Z0-9_-]{1,36}$"), "must match ^[a-zA-Z0-9_-]{1,36}$"),
				},
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			names.AttrID: framework.IDAttribute(),
			names.AttrName: schema.StringAttribute{
				Required: true,
				Validators: []validator.String{
					stringvalidator.RegexMatches(regexache.MustCompile("^[\\w -]+$"), "must match ^[\\w -]+$"),
					stringvalidator.LengthBetween(1, 64),
				},
			},
			"project_identifier": schema.StringAttribute{
				Required: true,
				Validators: []validator.String{
					stringvalidator.RegexMatches(regexache.MustCompile("^[a-zA-Z0-9_-]{1,36}$"), "must match ^[a-zA-Z0-9_-]{1,36}$"),
				},
			},
			"updated_at": schema.StringAttribute{
				CustomType: timetypes.RFC3339Type{},
				Computed:   true,
			},
		},
		Blocks: map[string]schema.Block{
			"user_parameters": schema.ListNestedBlock{
				CustomType: fwtypes.NewListNestedObjectTypeOf[userParametersData](ctx),
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						names.AttrName: schema.StringAttribute{
							Optional: true,
						},
						names.AttrValue: schema.StringAttribute{
							Optional: true,
						},
					},
				},
			},
		},
	}
}

func (r *environmentProfileResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	conn := r.Meta().DataZoneClient(ctx)

	var plan environmentProfileResourceModel
	smerr.AddEnrich(ctx, &resp.Diagnostics, req.Plan.Get(ctx, &plan))
	if resp.Diagnostics.HasError() {
		return
	}

	in := &datazone.CreateEnvironmentProfileInput{}
	smerr.AddEnrich(ctx, &resp.Diagnostics, fwflex.Expand(ctx, &plan, in))
	if resp.Diagnostics.HasError() {
		return
	}
	in.EnvironmentBlueprintIdentifier = plan.EnvironmentBlueprintId.ValueStringPointer()

	out, err := conn.CreateEnvironmentProfile(ctx, in)
	if err != nil {
		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, plan.Name.ValueString())
		return
	}
	if out == nil {
		smerr.AddError(ctx, &resp.Diagnostics, errors.New("empty output"), smerr.ID, plan.Name.ValueString())
		return
	}

	option := fwflex.WithIgnoredFieldNamesAppend("UserParameters")
	smerr.AddEnrich(ctx, &resp.Diagnostics, fwflex.Flatten(ctx, out, &plan, option))
	if resp.Diagnostics.HasError() {
		return
	}

	smerr.AddEnrich(ctx, &resp.Diagnostics, resp.State.Set(ctx, &plan))
}

func (r *environmentProfileResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	conn := r.Meta().DataZoneClient(ctx)

	var state environmentProfileResourceModel
	smerr.AddEnrich(ctx, &resp.Diagnostics, req.State.Get(ctx, &state))
	if resp.Diagnostics.HasError() {
		return
	}

	out, err := findEnvironmentProfileByID(ctx, conn, state.Id.ValueString(), state.DomainIdentifier.ValueString())
	if retry.NotFound(err) {
		resp.State.RemoveResource(ctx)
		return
	}
	if err != nil {
		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, state.Id.ValueString())
		return
	}

	option := fwflex.WithIgnoredFieldNamesAppend("UserParameters")
	smerr.AddEnrich(ctx, &resp.Diagnostics, fwflex.Flatten(ctx, out, &state, option))
	if resp.Diagnostics.HasError() {
		return
	}
	state.EnvironmentBlueprintId = fwflex.StringToFramework(ctx, out.EnvironmentBlueprintId)
	state.ProjectIdentifier = fwflex.StringToFramework(ctx, out.ProjectId)

	smerr.AddEnrich(ctx, &resp.Diagnostics, resp.State.Set(ctx, &state))
}

func (r *environmentProfileResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	conn := r.Meta().DataZoneClient(ctx)

	var plan, state environmentProfileResourceModel
	smerr.AddEnrich(ctx, &resp.Diagnostics, req.Plan.Get(ctx, &plan))
	smerr.AddEnrich(ctx, &resp.Diagnostics, req.State.Get(ctx, &state))
	if resp.Diagnostics.HasError() {
		return
	}

	diff, d := fwflex.Diff(ctx, plan, state)
	smerr.AddEnrich(ctx, &resp.Diagnostics, d)
	if resp.Diagnostics.HasError() {
		return
	}

	if diff.HasChanges() {
		in := &datazone.UpdateEnvironmentProfileInput{}
		smerr.AddEnrich(ctx, &resp.Diagnostics, fwflex.Expand(ctx, plan, in))
		if resp.Diagnostics.HasError() {
			return
		}
		in.Identifier = state.Id.ValueStringPointer()

		out, err := conn.UpdateEnvironmentProfile(ctx, in)
		if err != nil {
			smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, state.Id.ValueString())
			return
		}

		option := fwflex.WithIgnoredFieldNamesAppend("UserParameters")
		smerr.AddEnrich(ctx, &resp.Diagnostics, fwflex.Flatten(ctx, out, &state, option))
		if resp.Diagnostics.HasError() {
			return
		}
	}

	smerr.AddEnrich(ctx, &resp.Diagnostics, resp.State.Set(ctx, &state))
}

func (r *environmentProfileResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	conn := r.Meta().DataZoneClient(ctx)

	var state environmentProfileResourceModel
	smerr.AddEnrich(ctx, &resp.Diagnostics, req.State.Get(ctx, &state))
	if resp.Diagnostics.HasError() {
		return
	}

	input := datazone.DeleteEnvironmentProfileInput{
		Identifier:       state.Id.ValueStringPointer(),
		DomainIdentifier: state.DomainIdentifier.ValueStringPointer(),
	}
	_, err := conn.DeleteEnvironmentProfile(ctx, &input)

	if err != nil && !errs.IsA[*awstypes.ResourceNotFoundException](err) {
		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, state.Id.ValueString())
		return
	}
}

func findEnvironmentProfileByID(ctx context.Context, conn *datazone.Client, id string, domainID string) (*datazone.GetEnvironmentProfileOutput, error) {
	in := &datazone.GetEnvironmentProfileInput{
		Identifier:       aws.String(id),
		DomainIdentifier: aws.String(domainID),
	}

	out, err := conn.GetEnvironmentProfile(ctx, in)
	if err != nil {
		if errs.IsA[*awstypes.ResourceNotFoundException](err) {
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
	_ inttypes.ImportIDParser = environmentProfileImportID{}
)

type environmentProfileImportID struct{}

func (environmentProfileImportID) Parse(id string) (string, map[string]any, error) {
	profileID, domainID, found := strings.Cut(id, intflex.ResourceIdSeparator)
	if !found {
		return "", nil, fmt.Errorf("id %q should be in the format <id>%s<domain-identifier>", id, intflex.ResourceIdSeparator)
	}

	result := map[string]any{
		names.AttrID:        profileID,
		"domain_identifier": domainID,
	}

	return id, result, nil
}

type environmentProfileResourceModel struct {
	framework.WithRegionModel
	AwsAccountId           types.String                                        `tfsdk:"aws_account_id"`
	AwsAccountRegion       types.String                                        `tfsdk:"aws_account_region"`
	CreatedAt              timetypes.RFC3339                                   `tfsdk:"created_at"`
	CreatedBy              types.String                                        `tfsdk:"created_by"`
	Description            types.String                                        `tfsdk:"description"`
	DomainIdentifier       types.String                                        `tfsdk:"domain_identifier"`
	EnvironmentBlueprintId types.String                                        `tfsdk:"environment_blueprint_identifier"`
	Id                     types.String                                        `tfsdk:"id"`
	Name                   types.String                                        `tfsdk:"name"`
	ProjectIdentifier      types.String                                        `tfsdk:"project_identifier"`
	UpdatedAt              timetypes.RFC3339                                   `tfsdk:"updated_at"`
	UserParameters         fwtypes.ListNestedObjectValueOf[userParametersData] `tfsdk:"user_parameters"`
}

type userParametersData struct {
	Name  types.String `tfsdk:"name"`
	Value types.String `tfsdk:"value"`
}
