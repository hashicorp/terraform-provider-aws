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
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/listplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	"github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkResource("aws_datazone_environment", name="Environment")
func newResourceEnvironment(_ context.Context) (resource.ResourceWithConfigure, error) {
	r := &resourceEnvironment{}

	r.SetDefaultCreateTimeout(30 * time.Minute)
	r.SetDefaultUpdateTimeout(30 * time.Minute)
	r.SetDefaultDeleteTimeout(30 * time.Minute)

	return r, nil
}

const (
	ResNameEnvironment = "Environment"
)

type resourceEnvironment struct {
	framework.ResourceWithConfigure
	framework.WithTimeouts
}

func (r *resourceEnvironment) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = "aws_datazone_environment"
}

func (r *resourceEnvironment) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"description": schema.StringAttribute{
				Optional: true,
				Computed: true,
			},
			"environment_account_identifier": schema.StringAttribute{
				Optional: true,
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"environment_account_region": schema.StringAttribute{
				Optional: true,
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			names.AttrCreatedAt: schema.StringAttribute{
				CustomType: timetypes.RFC3339Type{},
				Computed:   true,
			},
			"created_by": schema.StringAttribute{
				Computed: true,
			},
			"custom_parameters": schema.ListAttribute{
				CustomType: fwtypes.NewListNestedObjectTypeOf[resourceCustomParameterData](ctx),
				Computed:   true,
			},
			"deployment_parameters": schema.ListAttribute{
				CustomType: fwtypes.NewListNestedObjectTypeOf[resourceDeploymentPropertiesData](ctx), // double cehck this
				Computed:   true,
			},
			"domain_identifier": schema.StringAttribute{
				Required: true,
			},
			"environment_actions": schema.ListAttribute{
				CustomType: fwtypes.NewListNestedObjectTypeOf[resourceEnvironmentActionData](ctx),
				Computed:   true,
			},
			"environment_blueprint_identifier": schema.StringAttribute{
				Optional: true,
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"environment_profile_identifier": schema.StringAttribute{
				Required: true,
			},
			"glossary_terms": schema.ListAttribute{
				CustomType: fwtypes.ListOfStringType,
				Optional:   true,
			},
			"id": schema.StringAttribute{ // fix this i forgot
				Computed: true,
			},
			"last_deployment": schema.ListAttribute{
				CustomType: fwtypes.NewListNestedObjectTypeOf[resourceLastDeployment](ctx),
				Computed:   true,
			},
			"name": schema.StringAttribute{
				Required: true,
			},
			"project_identifier": schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"provider_environment": schema.StringAttribute{
				Computed: true,
			},
			"provisioned_resources": schema.ListAttribute{
				CustomType: fwtypes.NewListNestedObjectTypeOf[resourceProvisionedResourcesData](ctx),
				Computed:   true,
			},
			"provisioning_properties": schema.ListAttribute{
				CustomType: fwtypes.NewListNestedObjectTypeOf[resourceProvisionPropertiesData](ctx),
				Computed:   true,
			},
			"status": schema.StringAttribute{
				CustomType: fwtypes.StringEnumType[awstypes.EnvironmentStatus](),
				Computed:   true,
			},
			"updated_at": schema.StringAttribute{
				CustomType: timetypes.RFC3339Type{},
				Computed:   true,
			},
		},
		Blocks: map[string]schema.Block{
			"user_parameters": schema.ListNestedBlock{
				CustomType: fwtypes.NewListNestedObjectTypeOf[resourceUserParametersData](ctx),
				PlanModifiers: []planmodifier.List{
					listplanmodifier.UseStateForUnknown(),
				},
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
			"timeouts": timeouts.Block(ctx, timeouts.Opts{
				Create: true,
				Update: true,
				Delete: true,
			}),
		},
	}
}

func (r *resourceEnvironment) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	conn := r.Meta().DataZoneClient(ctx)
	var plan resourceEnvironmentData
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	fmt.Printf("plan: %v\n", plan)
	fmt.Printf("plan.DomainIdentifier: %v\n", plan.DomainIdentifier)
	fmt.Printf("plan.ProjectIdentifier: %v\n", plan.ProjectIdentifier)
	fmt.Printf("plan.EnvironmentProfileIdentifier: %v\n", plan.EnvironmentProfileIdentifier)
	fmt.Printf("plan.EnvironmentBlueprintId: %v\n", plan.EnvironmentBlueprintId)
	fmt.Printf("plan.AwsAccountId: %v\n", plan.AwsAccountId)
	fmt.Printf("plan.AwsAccountRegion: %v\n", plan.AwsAccountRegion)
	fmt.Printf("plan.Name: %v\n", plan.Name)

	option := flex.WithIgnoredFieldNames([]string{"UserParameters", "CustomParameters"})
	in := &datazone.CreateEnvironmentInput{}
	resp.Diagnostics.Append(flex.Expand(ctx, &plan, in)...)
	if resp.Diagnostics.HasError() {
		return
	}

	out, err := conn.CreateEnvironment(ctx, in)
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.DataZone, create.ErrActionCreating, ResNameEnvironment, plan.Name.String(), err),
			err.Error(),
		)
		return
	}
	if out == nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.DataZone, create.ErrActionCreating, ResNameEnvironment, plan.Name.String(), nil),
			errors.New("empty output").Error(),
		)
		return
	}

	resp.Diagnostics.Append(flex.Flatten(ctx, out, &plan, option)...)
	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.Append(flex.Flatten(ctx, out.UserParameters, &plan.CustomParameters)...)
	if resp.Diagnostics.HasError() {
		return
	}
	createTimeout := r.CreateTimeout(ctx, plan.Timeouts)
	_, err = waitEnvironmentCreated(ctx, conn, plan.DomainIdentifier.ValueString(), plan.Id.ValueString(), createTimeout)
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.DataZone, create.ErrActionWaitingForCreation, ResNameEnvironment, plan.Name.String(), err),
			err.Error(),
		)
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}
func (r *resourceEnvironment) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	conn := r.Meta().DataZoneClient(ctx)

	var state resourceEnvironmentData
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	out, err := findEnvironmentByID(ctx, conn, state.DomainIdentifier.ValueString(), state.Id.ValueString())
	if tfresource.NotFound(err) {
		resp.State.RemoveResource(ctx)
		return
	}
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.DataZone, create.ErrActionSetting, ResNameEnvironment, state.Id.String(), err),
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
func (r *resourceEnvironment) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	conn := r.Meta().DataZoneClient(ctx)

	// TIP: -- 2. Fetch the plan
	var plan, state resourceEnvironmentData
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if !plan.Name.Equal(state.Name) ||
		!plan.Description.Equal(state.Description) ||
		!plan.GlossaryTerms.Equal(state.GlossaryTerms) {

		in := &datazone.UpdateEnvironmentInput{}
		resp.Diagnostics.Append(flex.Expand(ctx, plan, in)...)
		if resp.Diagnostics.HasError() {
			return
		}
		out, err := conn.UpdateEnvironment(ctx, in)
		if err != nil {
			resp.Diagnostics.AddError(
				create.ProblemStandardMessage(names.DataZone, create.ErrActionUpdating, ResNameEnvironment, plan.Id.String(), err),
				err.Error(),
			)
			return
		}
		if out == nil {
			resp.Diagnostics.AddError(
				create.ProblemStandardMessage(names.DataZone, create.ErrActionUpdating, ResNameEnvironment, plan.Id.String(), nil),
				errors.New("empty output").Error(),
			)
			return
		}

	}

	updateTimeout := r.UpdateTimeout(ctx, plan.Timeouts)
	_, err := waitEnvironmentUpdated(ctx, conn, plan.DomainIdentifier.ValueString(), plan.Id.ValueString(), updateTimeout)
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.DataZone, create.ErrActionWaitingForUpdate, ResNameEnvironment, plan.Id.String(), err),
			err.Error(),
		)
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *resourceEnvironment) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	conn := r.Meta().DataZoneClient(ctx)

	var state resourceEnvironmentData
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	in := &datazone.DeleteEnvironmentInput{
		DomainIdentifier: state.DomainIdentifier.ValueStringPointer(),
		Identifier:       state.Id.ValueStringPointer(),
	}

	_, err := conn.DeleteEnvironment(ctx, in)
	if err != nil {
		if errs.IsA[*awstypes.ResourceNotFoundException](err) {
			return
		}
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.DataZone, create.ErrActionDeleting, ResNameEnvironment, state.Id.String(), err),
			err.Error(),
		)
		return
	}

	// TIP: -- 5. Use a waiter to wait for delete to complete
	deleteTimeout := r.DeleteTimeout(ctx, state.Timeouts)
	_, err = waitEnvironmentDeleted(ctx, conn, state.DomainIdentifier.ValueString(), state.Id.ValueString(), deleteTimeout)
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.DataZone, create.ErrActionWaitingForDeletion, ResNameEnvironment, state.Id.String(), err),
			err.Error(),
		)
		return
	}
}

func (r *resourceEnvironment) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	parts := strings.Split(req.ID, ":")

	if len(parts) != 2 {
		resp.Diagnostics.AddError("Resource Import Invalid ID", fmt.Sprintf(`Unexpected format for import ID (%s), use: "DomainIdentifier,Id"`, req.ID))
	}
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("domain_identifier"), parts[0])...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root(names.AttrIdentifier), parts[1])...)
}

func waitEnvironmentCreated(ctx context.Context, conn *datazone.Client, domainId string, id string, timeout time.Duration) (*datazone.GetEnvironmentOutput, error) {
	stateConf := &retry.StateChangeConf{
		Pending:                   enum.Slice[awstypes.EnvironmentStatus](awstypes.EnvironmentStatusCreating),
		Target:                    enum.Slice[awstypes.EnvironmentStatus](awstypes.EnvironmentStatusActive),
		Refresh:                   statusEnvironment(ctx, conn, domainId, id),
		Timeout:                   timeout,
		NotFoundChecks:            20,
		ContinuousTargetOccurence: 2,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)
	if out, ok := outputRaw.(*datazone.GetEnvironmentOutput); ok {
		return out, err
	}

	return nil, err
}

func waitEnvironmentUpdated(ctx context.Context, conn *datazone.Client, domainId string, id string, timeout time.Duration) (*datazone.GetEnvironmentOutput, error) {
	stateConf := &retry.StateChangeConf{
		Pending:                   enum.Slice[awstypes.EnvironmentStatus](awstypes.EnvironmentStatusUpdating),
		Target:                    enum.Slice[awstypes.EnvironmentStatus](awstypes.EnvironmentStatusActive),
		Refresh:                   statusEnvironment(ctx, conn, domainId, id),
		Timeout:                   timeout,
		NotFoundChecks:            20,
		ContinuousTargetOccurence: 2,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)
	if out, ok := outputRaw.(*datazone.GetEnvironmentOutput); ok {
		return out, err
	}

	return nil, err
}

func waitEnvironmentDeleted(ctx context.Context, conn *datazone.Client, domainId string, id string, timeout time.Duration) (*datazone.GetEnvironmentOutput, error) {
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice[awstypes.EnvironmentStatus](awstypes.EnvironmentStatusDeleting, awstypes.EnvironmentStatusActive),
		Target:  enum.Slice[awstypes.EnvironmentStatus](awstypes.EnvironmentStatusDeleted),
		Refresh: statusEnvironment(ctx, conn, domainId, id),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)
	if out, ok := outputRaw.(*datazone.GetEnvironmentOutput); ok {
		return out, err
	}

	return nil, err
}

func statusEnvironment(ctx context.Context, conn *datazone.Client, domainId string, id string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		out, err := findEnvironmentByID(ctx, conn, domainId, id)
		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return out, aws.ToString((*string)(&out.Status)), nil
	}
}

func findEnvironmentByID(ctx context.Context, conn *datazone.Client, domainId string, id string) (*datazone.GetEnvironmentOutput, error) {
	in := &datazone.GetEnvironmentInput{
		DomainIdentifier: aws.String(domainId),
		Identifier:       aws.String(id),
	}

	out, err := conn.GetEnvironment(ctx, in)
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

type resourceEnvironmentData struct {
	AwsAccountId                 types.String                                                      `tfsdk:"environment_account_identifier"`
	AwsAccountRegion             types.String                                                      `tfsdk:"environment_account_region"`
	CreatedAt                    timetypes.RFC3339                                                 `tfsdk:"created_at"`
	CreatedBy                    types.String                                                      `tfsdk:"created_by"`
	CustomParameters             fwtypes.ListNestedObjectValueOf[resourceCustomParameterData]      `tfsdk:"custom_parameters"`
	DeploymentParameters         fwtypes.ListNestedObjectValueOf[resourceDeploymentPropertiesData] `tfsdk:"deployment_parameters"`
	Description                  types.String                                                      `tfsdk:"description"`
	DomainIdentifier             types.String                                                      `tfsdk:"domain_identifier"`
	EnvironmentActions           fwtypes.ListNestedObjectValueOf[resourceEnvironmentActionData]    `tfsdk:"environment_actions"`
	EnvironmentBlueprintId       types.String                                                      `tfsdk:"environment_blueprint_identifier"`
	EnvironmentProfileIdentifier types.String                                                      `tfsdk:"environment_profile_identifier"`
	GlossaryTerms                fwtypes.ListValueOf[types.String]                                 `tfsdk:"glossary_terms"`
	Id                           types.String                                                      `tfsdk:"id"`
	LastDeployment               fwtypes.ListNestedObjectValueOf[resourceLastDeployment]           `tfsdk:"last_deployment"`
	Name                         types.String                                                      `tfsdk:"name"`
	ProjectIdentifier            types.String                                                      `tfsdk:"project_identifier"`
	Provider                     types.String                                                      `tfsdk:"provider_environment"`
	ProvisionedResources         fwtypes.ListNestedObjectValueOf[resourceProvisionedResourcesData] `tfsdk:"provisioned_resources"`
	ProvisioningProperties       fwtypes.ListNestedObjectValueOf[resourceProvisionPropertiesData]  `tfsdk:"provisioning_properties"`
	Status                       fwtypes.StringEnum[awstypes.EnvironmentStatus]                    `tfsdk:"status"`
	Timeouts                     timeouts.Value                                                    `tfsdk:"timeouts"`
	UpdatedAt                    timetypes.RFC3339                                                 `tfsdk:"updated_at"`
	UserParameters               fwtypes.ListNestedObjectValueOf[resourceUserParametersData]       `tfsdk:"user_parameters"`
}

type resourceCustomParameterData struct {
	DefaultValue types.String `tfsdk:"default_value"`
	Description  types.String `tfsdk:"description"`
	FieldType    types.String `tfsdk:"field_type"`
	IsEditable   types.Bool   `tfsdk:"is_editable"`
	IsOptional   types.Bool   `tfsdk:"is_optional"`
	KeyName      types.String `tfsdk:"key_name"`
}

type resourceDeploymentPropertiesData struct {
	EndTimeoutMinutes   types.Int64 `tfsdk:"is_optional"`
	StartTimeoutMinutes types.Int64 `tfsdk:"key_name"`
}

type resourceEnvironmentActionData struct {
	Auth       types.String                                            `tfsdk:"auth"`
	Parameters fwtypes.ListNestedObjectValueOf[resourceParametersData] `tfsdk:"parameters"`
	Type       types.String                                            `tfsdk:"type"`
	// awstypes.ConfigurableActionTypeAuthorization???
}

type resourceParametersData struct {
	Key   types.String `tfsdk:"key"`
	Value types.String `tfsdk:"value"`
}

type resourceLastDeployment struct {
	DeploymentId         types.String                                                `tfsdk:"deployment_id"`
	DeploymentStatus     types.String                                                `tfsdk:"deployment_status"`
	DeploymentType       fwtypes.StringEnum[awstypes.DeploymentType]                 `tfsdk:"deployment_type"`
	FailureReasons       fwtypes.ListNestedObjectValueOf[resourceFailureReasonsData] `tfsdk:"failure_reasons"`
	IsDeploymentComplete types.Bool                                                  `tfsdk:"is_deployment_complete"`
	Messages             fwtypes.ListValueOf[types.String]                           `tfsdk:"messages"`
}

type resourceFailureReasonsData struct {
	Code    types.String `tfsdk:"code"`
	Message types.String `tfsdk:"message"`
}

type resourceProvisionedResourcesData struct {
	Name     types.String `tfsdk:"name"`
	Provider types.String `tfsdk:"provider"`
	Type     types.String `tfsdk:"type"`
	Value    types.String `tfsdk:"value"`
}

type resourceProvisionPropertiesData struct {
	CloudFormation types.String `tfsdk:"cloud_formation"`
}

type resourceCloudFormationData struct {
	template_url types.String `tfsdk:"template_url"`
}

type resourceUserParametersData struct {
	Name  types.String `tfsdk:"name"`
	Value types.String `tfsdk:"value"`
}
