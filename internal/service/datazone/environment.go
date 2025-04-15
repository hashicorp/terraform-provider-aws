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
	"github.com/hashicorp/terraform-provider-aws/internal/errs/fwdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	fwflex "github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkResource("aws_datazone_environment", name="Environment")
func newResourceEnvironment(_ context.Context) (resource.ResourceWithConfigure, error) {
	r := &resourceEnvironment{}

	r.SetDefaultCreateTimeout(10 * time.Minute)
	r.SetDefaultUpdateTimeout(10 * time.Minute)
	r.SetDefaultDeleteTimeout(10 * time.Minute)

	return r, nil
}

const (
	ResNameEnvironment = "Environment"
)

type resourceEnvironment struct {
	framework.ResourceWithConfigure
	framework.WithTimeouts
}

func (r *resourceEnvironment) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"account_identifier": schema.StringAttribute{
				Optional: true,
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"account_region": schema.StringAttribute{
				Optional: true,
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"blueprint_identifier": schema.StringAttribute{
				Optional: true,
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
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
			},
			"domain_identifier": schema.StringAttribute{
				Required: true,
			},
			"profile_identifier": schema.StringAttribute{
				Required: true,
			},
			"glossary_terms": schema.ListAttribute{
				CustomType: fwtypes.ListOfStringType,
				Optional:   true,
			},
			names.AttrID: framework.IDAttribute(),
			"last_deployment": schema.ListAttribute{
				CustomType: fwtypes.NewListNestedObjectTypeOf[resourceLastDeployment](ctx),
				Computed:   true,
				PlanModifiers: []planmodifier.List{
					listplanmodifier.UseStateForUnknown(),
				},
			},
			names.AttrName: schema.StringAttribute{
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
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"provisioned_resources": schema.ListAttribute{
				CustomType: fwtypes.NewListNestedObjectTypeOf[resourceProvisionedResourcesData](ctx),
				Computed:   true,
				PlanModifiers: []planmodifier.List{
					listplanmodifier.UseStateForUnknown(),
				},
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
			names.AttrTimeouts: timeouts.Block(ctx, timeouts.Opts{
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

	in := &datazone.CreateEnvironmentInput{}
	resp.Diagnostics.Append(fwflex.Expand(ctx, &plan, in, fwflex.WithFieldNamePrefix("Environment"))...)
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

	state := plan
	state.Id = fwflex.StringToFramework(ctx, out.Id)

	// set partial state
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root(names.AttrID), out.Id)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("domain_identifier"), out.DomainId)...)

	if resp.Diagnostics.HasError() {
		return
	}

	createTimeout := r.CreateTimeout(ctx, plan.Timeouts)
	output, err := waitEnvironmentCreated(ctx, conn, state.DomainIdentifier.ValueString(), state.Id.ValueString(), createTimeout)

	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.DataZone, create.ErrActionWaitingForCreation, ResNameEnvironment, plan.Name.String(), err),
			err.Error(),
		)
		return
	}

	resp.Diagnostics.Append(fwflex.Flatten(ctx, output, &state, fwflex.WithIgnoredFieldNames([]string{"UserParameters"}))...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
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
		resp.Diagnostics.Append(fwdiag.NewResourceNotFoundWarningDiagnostic(err))
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

	resp.Diagnostics.Append(fwflex.Flatten(ctx, out, &state, fwflex.WithIgnoredFieldNamesAppend("UserParameters"),
		fwflex.WithFieldNamePrefix("Environment"),
	)...)

	if resp.Diagnostics.HasError() {
		return
	}

	state.AccountIdentifier = fwflex.StringToFramework(ctx, out.AwsAccountId)
	state.AccountRegion = fwflex.StringToFramework(ctx, out.AwsAccountRegion)
	state.ProjectIdentifier = fwflex.StringToFramework(ctx, out.ProjectId)
	state.ProfileIdentifier = fwflex.StringToFramework(ctx, out.EnvironmentProfileId)

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}
func (r *resourceEnvironment) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	conn := r.Meta().DataZoneClient(ctx)

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
		resp.Diagnostics.Append(fwflex.Expand(ctx, plan, in)...)

		if resp.Diagnostics.HasError() {
			return
		}
		in.Identifier = state.Id.ValueStringPointer()

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

		updateTimeout := r.UpdateTimeout(ctx, plan.Timeouts)
		_, err = waitEnvironmentUpdated(ctx, conn, plan.DomainIdentifier.ValueString(), plan.Id.ValueString(), updateTimeout)

		if err != nil {
			resp.Diagnostics.AddError(
				create.ProblemStandardMessage(names.DataZone, create.ErrActionWaitingForUpdate, ResNameEnvironment, plan.Id.String(), err),
				err.Error(),
			)
			return
		}
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

	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return
	}

	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.DataZone, create.ErrActionDeleting, ResNameEnvironment, state.Id.String(), err),
			err.Error(),
		)
		return
	}

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
	parts := strings.Split(req.ID, ",")

	if len(parts) != 2 {
		resp.Diagnostics.AddError("resource import invalid ID", fmt.Sprintf(`Unexpected format for import ID (%s), use: "DomainIdentifier,Id"`, req.ID))
		return
	}

	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("domain_identifier"), parts[0])...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root(names.AttrID), parts[1])...)
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
		if status, deployment := out.Status, out.LastDeployment; status == awstypes.EnvironmentStatusCreateFailed && deployment != nil {
			tfresource.SetLastError(err, fmt.Errorf("%s: %s", status, aws.ToString(deployment.FailureReason.Message)))
		}
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
		if status, deployment := out.Status, out.LastDeployment; status == awstypes.EnvironmentStatusUpdateFailed && deployment != nil {
			tfresource.SetLastError(err, fmt.Errorf("%s: %s", status, aws.ToString(deployment.FailureReason.Message)))
		}
		return out, err
	}

	return nil, err
}

func waitEnvironmentDeleted(ctx context.Context, conn *datazone.Client, domainId string, id string, timeout time.Duration) (*datazone.GetEnvironmentOutput, error) {
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(awstypes.EnvironmentStatusDeleting, awstypes.EnvironmentStatusActive),
		Target:  []string{},
		Refresh: statusEnvironment(ctx, conn, domainId, id),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)
	if out, ok := outputRaw.(*datazone.GetEnvironmentOutput); ok {
		if status, deployment := out.Status, out.LastDeployment; status == awstypes.EnvironmentStatusDeleteFailed && deployment != nil {
			tfresource.SetLastError(err, fmt.Errorf("%s: %s", status, aws.ToString(deployment.FailureReason.Message)))
		}
		return out, err
	}

	return nil, err
}

func statusEnvironment(ctx context.Context, conn *datazone.Client, domainId, id string) retry.StateRefreshFunc {
	return func() (any, string, error) {
		out, err := findEnvironmentByID(ctx, conn, domainId, id)
		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return out, string(out.Status), nil
	}
}

func findEnvironmentByID(ctx context.Context, conn *datazone.Client, domainId, id string) (*datazone.GetEnvironmentOutput, error) {
	in := &datazone.GetEnvironmentInput{
		DomainIdentifier: aws.String(domainId),
		Identifier:       aws.String(id),
	}

	out, err := conn.GetEnvironment(ctx, in)

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

type resourceEnvironmentData struct {
	AccountIdentifier    types.String                                                      `tfsdk:"account_identifier"`
	AccountRegion        types.String                                                      `tfsdk:"account_region"`
	BlueprintId          types.String                                                      `tfsdk:"blueprint_identifier"`
	CreatedAt            timetypes.RFC3339                                                 `tfsdk:"created_at"`
	CreatedBy            types.String                                                      `tfsdk:"created_by"`
	Description          types.String                                                      `tfsdk:"description"`
	DomainIdentifier     types.String                                                      `tfsdk:"domain_identifier"`
	ProfileIdentifier    types.String                                                      `tfsdk:"profile_identifier"`
	GlossaryTerms        fwtypes.ListValueOf[types.String]                                 `tfsdk:"glossary_terms"`
	Id                   types.String                                                      `tfsdk:"id"`
	LastDeployment       fwtypes.ListNestedObjectValueOf[resourceLastDeployment]           `tfsdk:"last_deployment"`
	Name                 types.String                                                      `tfsdk:"name"`
	ProjectIdentifier    types.String                                                      `tfsdk:"project_identifier"`
	Provider             types.String                                                      `tfsdk:"provider_environment"`
	ProvisionedResources fwtypes.ListNestedObjectValueOf[resourceProvisionedResourcesData] `tfsdk:"provisioned_resources"`
	Timeouts             timeouts.Value                                                    `tfsdk:"timeouts"`
	UserParameters       fwtypes.ListNestedObjectValueOf[resourceUserParametersData]       `tfsdk:"user_parameters"`
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

type resourceUserParametersData struct {
	Name  types.String `tfsdk:"name"`
	Value types.String `tfsdk:"value"`
}
