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
	"github.com/hashicorp/terraform-plugin-framework-timetypes/timetypes"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/setplanmodifier"
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

// @FrameworkResource("aws_datazone_environment_profile", name="Environment Profile")
func newResourceEnvironmentProfile(_ context.Context) (resource.ResourceWithConfigure, error) {
	return &resourceEnvironmentProfile{}, nil
}

const (
	ResNameEnvironmentProfile = "Environment Profile"
)

type resourceEnvironmentProfile struct {
	framework.ResourceWithConfigure
}

func (r *resourceEnvironmentProfile) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = "aws_datazone_environment_profile"
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
			"environment_blueprint_identifier": schema.StringAttribute{
				Required: true,
				Validators: []validator.String{
					stringvalidator.RegexMatches(regexache.MustCompile("^[a-zA-Z0-9_-]{1,36}$"), "must match ^[a-zA-Z0-9_-]{1,36}$"),
				},
			},
			names.AttrDescription: schema.StringAttribute{
				Optional: true,
				Computed: true,
				Validators: []validator.String{
					stringvalidator.LengthBetween(0, 2048),
				},
			},
			names.AttrID: schema.StringAttribute{
				Computed: true,
			},
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
			names.AttrCreatedAt: schema.StringAttribute{
				CustomType: timetypes.RFC3339Type{},
				Computed:   true,
			},
			"created_by": schema.StringAttribute{
				Computed: true,
			},
			"domain_identifier": schema.StringAttribute{
				Required: true,
			},
			"updated_at": schema.StringAttribute{
				CustomType: timetypes.RFC3339Type{},
				Computed:   true,
			},
		},
		Blocks: map[string]schema.Block{
			"user_parameters": schema.SetNestedBlock{
				PlanModifiers: []planmodifier.Set{
					setplanmodifier.RequiresReplace(),
					setplanmodifier.RequiresReplaceIfConfigured(),
				},
				CustomType: fwtypes.NewSetNestedObjectTypeOf[dUserParameters](ctx),
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						names.AttrName: schema.StringAttribute{
							Optional: true,
						},
						names.AttrValue: schema.StringAttribute{
							Optional: true,
							Computed: false,
						},
					},
				},
			},
		},
	}
}

func (r *resourceEnvironmentProfile) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	conn := r.Meta().DataZoneClient(ctx)
	var plan dEnvProfile
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	in := &datazone.CreateEnvironmentProfileInput{}
	resp.Diagnostics.Append(flex.Expand(ctx, &plan, in)...)
	if resp.Diagnostics.HasError() {
		return
	}
	in.Name = plan.Name.ValueStringPointer()
	in.EnvironmentBlueprintIdentifier = plan.EnvironmentBlueprintIdentifier.ValueStringPointer()
	in.AwsAccountId = plan.AwsAccountId.ValueStringPointer()

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

	plan.AwsAccountId = flex.StringToFramework(ctx, out.AwsAccountId)
	plan.AwsAccountRegion = flex.StringToFramework(ctx, out.AwsAccountRegion)
	plan.CreatedAt = flex.TimeToFramework(ctx, out.CreatedAt)
	plan.CreatedBy = flex.StringToFramework(ctx, out.CreatedBy)
	plan.Description = flex.StringToFramework(ctx, out.Description)
	plan.DomainIdentifier = flex.StringToFramework(ctx, out.DomainId)
	plan.EnvironmentBlueprintIdentifier = flex.StringToFramework(ctx, out.EnvironmentBlueprintId)
	plan.Id = flex.StringToFramework(ctx, out.Id)
	plan.Name = flex.StringToFramework(ctx, out.Name)
	plan.ProjectIdentifier = flex.StringToFramework(ctx, out.ProjectId)
	plan.UpdatedAt = flex.TimeToFramework(ctx, out.UpdatedAt)

	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *resourceEnvironmentProfile) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	conn := r.Meta().DataZoneClient(ctx)
	o := flex.AutoFlexOptions{}
	o.AddIgnoredField("user_parameters")

	// TIP: -- 2. Fetch the state
	var state dEnvProfile
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
			create.ProblemStandardMessage(names.DataZone, create.ErrActionSetting, ResNameEnvironmentProfile, state.Id.String(), err),
			err.Error(),
		)
		return
	}

	state.AwsAccountId = flex.StringToFramework(ctx, out.AwsAccountId)
	state.AwsAccountRegion = flex.StringToFramework(ctx, out.AwsAccountRegion)
	state.CreatedAt = flex.TimeToFramework(ctx, out.CreatedAt)
	state.CreatedBy = flex.StringToFramework(ctx, out.CreatedBy)
	state.Description = flex.StringToFramework(ctx, out.Description)
	state.DomainIdentifier = flex.StringToFramework(ctx, out.DomainId)
	state.EnvironmentBlueprintIdentifier = flex.StringToFramework(ctx, out.EnvironmentBlueprintId)
	state.Id = flex.StringToFramework(ctx, out.Id)
	state.Name = flex.StringToFramework(ctx, out.Name)
	state.ProjectIdentifier = flex.StringToFramework(ctx, out.ProjectId)
	state.UpdatedAt = flex.TimeToFramework(ctx, out.UpdatedAt)

	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *resourceEnvironmentProfile) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	conn := r.Meta().DataZoneClient(ctx)

	var plan, state dEnvProfile
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if plan.EnvironmentBlueprintIdentifier.Equal(state.EnvironmentBlueprintIdentifier) &&
		plan.ProjectIdentifier.Equal(state.ProjectIdentifier) && plan.DomainIdentifier.Equal(state.DomainIdentifier) {
		in := &datazone.UpdateEnvironmentProfileInput{}

		resp.Diagnostics.Append(flex.Expand(ctx, plan, in)...)
		if resp.Diagnostics.HasError() {
			return
		}
		in.Identifier = state.Id.ValueStringPointer()
		out, err := conn.UpdateEnvironmentProfile(ctx, in)
		if err != nil {
			resp.Diagnostics.AddError(
				create.ProblemStandardMessage(names.DataZone, create.ErrActionSetting, ResNameEnvironmentProfile, state.Id.ValueString(), err),
				err.Error(),
			)
			return
		}
		state.AwsAccountId = flex.StringToFramework(ctx, out.AwsAccountId)
		state.AwsAccountRegion = flex.StringToFramework(ctx, out.AwsAccountRegion)
		state.CreatedAt = flex.TimeToFramework(ctx, out.CreatedAt)
		state.CreatedBy = flex.StringToFramework(ctx, out.CreatedBy)
		state.Description = flex.StringToFramework(ctx, out.Description)
		state.DomainIdentifier = flex.StringToFramework(ctx, out.DomainId)
		state.EnvironmentBlueprintIdentifier = flex.StringToFramework(ctx, out.EnvironmentBlueprintId)
		state.Id = flex.StringToFramework(ctx, out.Id)
		state.Name = flex.StringToFramework(ctx, out.Name)
		state.ProjectIdentifier = flex.StringToFramework(ctx, out.ProjectId)
		state.UpdatedAt = flex.TimeToFramework(ctx, out.UpdatedAt)

		if resp.Diagnostics.HasError() {
			return
		}
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *resourceEnvironmentProfile) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	conn := r.Meta().DataZoneClient(ctx)
	var state dEnvProfile
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)

	_, err := conn.DeleteEnvironmentProfile(ctx, &datazone.DeleteEnvironmentProfileInput{
		DomainIdentifier: state.DomainIdentifier.ValueStringPointer(),
		Identifier:       state.Id.ValueStringPointer(),
	})

	if err != nil && !errs.IsA[*awstypes.ResourceNotFoundException](err) {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.DataZone, create.ErrActionSetting, ResNameEnvironmentProfile, state.Id.ValueString(), err),
			err.Error(),
		)
		return
	}
}

func (r *resourceEnvironmentProfile) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	parts := strings.Split(req.ID, ":")

	if len(parts) != 4 {
		resp.Diagnostics.AddError("Resource Import Invalid ID", fmt.Sprintf(`Unexpected format for import ID (%s), use: "DomainIdentifier:Id"`, req.ID))
	}
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("domain_identifier"), parts[0])...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root(names.AttrID), parts[1])...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("environment_blueprint_identifier"), parts[2])...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("project_identifier"), parts[3])...)
}

func findEnvironmentProfileByID(ctx context.Context, conn *datazone.Client, id string, domain_id string) (*datazone.GetEnvironmentProfileOutput, error) {
	in := &datazone.GetEnvironmentProfileInput{
		Identifier:       aws.String(id),
		DomainIdentifier: aws.String(domain_id),
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

type dEnvProfile struct {
	// entered and or computed
	AwsAccountId                   types.String                                    `tfsdk:"aws_account_id"`
	AwsAccountRegion               types.String                                    `tfsdk:"aws_account_region"`
	Description                    types.String                                    `tfsdk:"description"`
	Id                             types.String                                    `tfsdk:"id"`
	EnvironmentBlueprintIdentifier types.String                                    `tfsdk:"environment_blueprint_identifier"`
	UserParameters                 fwtypes.SetNestedObjectValueOf[dUserParameters] `tfsdk:"user_parameters"`

	Name              types.String `tfsdk:"name"`
	ProjectIdentifier types.String `tfsdk:"project_identifier"`

	// computed
	CreatedAt        timetypes.RFC3339 `tfsdk:"created_at"`
	CreatedBy        types.String      `tfsdk:"created_by"`
	DomainIdentifier types.String      `tfsdk:"domain_identifier"`
	UpdatedAt        timetypes.RFC3339 `tfsdk:"updated_at"`
}

type dUserParameters struct {
	// entered
	Name  types.String `tfsdk:"name"`
	Value types.String `tfsdk:"value"`
}
