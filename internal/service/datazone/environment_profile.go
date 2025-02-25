// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package datazone

import (
	"context"
	"errors"
	"fmt"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/datazone"
	awstypes "github.com/aws/aws-sdk-go-v2/service/datazone/types"
	"github.com/hashicorp/terraform-plugin-framework-timetypes/timetypes"
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
	intflex "github.com/hashicorp/terraform-provider-aws/internal/flex"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	"github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkResource("aws_datazone_environment_profile", name="Environment Profile")
func newResourceEnvironmentProfile(_ context.Context) (resource.ResourceWithConfigure, error) {
	return &resourceEnvironmentProfile{}, nil
}

const (
	ResNameEnvironmentProfile = "Environment Profile"

	environmentProfileIDParts = 2
)

type resourceEnvironmentProfile struct {
	framework.ResourceWithConfigure
}

func (r *resourceEnvironmentProfile) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
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
					stringvalidator.RegexMatches(regexache.MustCompile("^[a-z]{2}-[a-z]{4,10}-\\d$"), "must match ^[a-z]{2}-[a-z]{4,10}-\\d$"),
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

func (r *resourceEnvironmentProfile) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	conn := r.Meta().DataZoneClient(ctx)

	var plan environmentProfileData
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	in := &datazone.CreateEnvironmentProfileInput{}
	resp.Diagnostics.Append(flex.Expand(ctx, &plan, in)...)
	if resp.Diagnostics.HasError() {
		return
	}
	in.EnvironmentBlueprintIdentifier = plan.EnvironmentBlueprintId.ValueStringPointer()

	out, err := conn.CreateEnvironmentProfile(ctx, in)
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.DataZone, create.ErrActionCreating, ResNameEnvironmentProfile, plan.Name.String(), err),
			err.Error(),
		)
		return
	}
	if out == nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.DataZone, create.ErrActionCreating, ResNameEnvironmentProfile, plan.Name.String(), nil),
			errors.New("empty output").Error(),
		)
		return
	}

	option := flex.WithIgnoredFieldNamesAppend("UserParameters")
	resp.Diagnostics.Append(flex.Flatten(ctx, out, &plan, option)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *resourceEnvironmentProfile) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	conn := r.Meta().DataZoneClient(ctx)

	var state environmentProfileData
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	out, err := findEnvironmentProfileByID(ctx, conn, state.Id.ValueString(), state.DomainIdentifier.ValueString())
	if tfresource.NotFound(err) {
		resp.State.RemoveResource(ctx)
		return
	}
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.DataZone, create.ErrActionReading, ResNameEnvironmentProfile, state.Id.String(), err),
			err.Error(),
		)
		return
	}

	option := flex.WithIgnoredFieldNamesAppend("UserParameters")
	resp.Diagnostics.Append(flex.Flatten(ctx, out, &state, option)...)
	if resp.Diagnostics.HasError() {
		return
	}
	state.EnvironmentBlueprintId = flex.StringToFramework(ctx, out.EnvironmentBlueprintId)
	state.ProjectIdentifier = flex.StringToFramework(ctx, out.ProjectId)

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *resourceEnvironmentProfile) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	conn := r.Meta().DataZoneClient(ctx)

	var plan, state environmentProfileData
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if !plan.Description.Equal(state.Description) ||
		!plan.Name.Equal(state.Name) ||
		!plan.UserParameters.Equal(state.UserParameters) ||
		!plan.AwsAccountId.Equal(state.AwsAccountId) ||
		!plan.AwsAccountRegion.Equal(state.AwsAccountRegion) {
		in := &datazone.UpdateEnvironmentProfileInput{}

		resp.Diagnostics.Append(flex.Expand(ctx, plan, in)...)
		if resp.Diagnostics.HasError() {
			return
		}
		in.Identifier = state.Id.ValueStringPointer()

		out, err := conn.UpdateEnvironmentProfile(ctx, in)
		if err != nil {
			resp.Diagnostics.AddError(
				create.ProblemStandardMessage(names.DataZone, create.ErrActionUpdating, ResNameEnvironmentProfile, state.Id.ValueString(), err),
				err.Error(),
			)
			return
		}

		option := flex.WithIgnoredFieldNamesAppend("UserParameters")
		resp.Diagnostics.Append(flex.Flatten(ctx, out, &state, option)...)
		if resp.Diagnostics.HasError() {
			return
		}
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *resourceEnvironmentProfile) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	conn := r.Meta().DataZoneClient(ctx)

	var state environmentProfileData
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	input := datazone.DeleteEnvironmentProfileInput{
		Identifier:       state.Id.ValueStringPointer(),
		DomainIdentifier: state.DomainIdentifier.ValueStringPointer(),
	}
	_, err := conn.DeleteEnvironmentProfile(ctx, &input)

	if err != nil && !errs.IsA[*awstypes.ResourceNotFoundException](err) {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.DataZone, create.ErrActionDeleting, ResNameEnvironmentProfile, state.Id.ValueString(), err),
			err.Error(),
		)
		return
	}
}

func (r *resourceEnvironmentProfile) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	parts, err := intflex.ExpandResourceId(req.ID, environmentProfileIDParts, false)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unexpected Import Identifier",
			fmt.Sprintf("Expected import identifier with format: id,domain_identifier. Got: %q", req.ID),
		)
		return
	}

	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root(names.AttrID), parts[0])...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("domain_identifier"), parts[1])...)
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

type environmentProfileData struct {
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
