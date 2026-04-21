// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package ecs

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ecs"
	awstypes "github.com/aws/aws-sdk-go-v2/service/ecs/types"
	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework-validators/float64validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/id"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/fwdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	fwflex "github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkResource("aws_ecs_daemon", name="Daemon")
// @Tags(identifierAttribute="arn")
// @ArnIdentity
// @Testing(hasNoPreExistingResource=true)
func newDaemonResource(context.Context) (resource.ResourceWithConfigure, error) {
	r := &daemonResource{}

	r.SetDefaultCreateTimeout(20 * time.Minute)
	r.SetDefaultUpdateTimeout(20 * time.Minute)
	r.SetDefaultDeleteTimeout(20 * time.Minute)

	return r, nil
}

type daemonResource struct {
	framework.ResourceWithModel[daemonResourceModel]
	framework.WithTimeouts
	framework.WithImportByIdentity
}

func (r *daemonResource) Schema(ctx context.Context, request resource.SchemaRequest, response *resource.SchemaResponse) {
	s := schema.Schema{
		Attributes: map[string]schema.Attribute{
			names.AttrARN: schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"capacity_provider_arns": schema.ListAttribute{
				CustomType:  fwtypes.ListOfStringType,
				Required:    true,
				ElementType: types.StringType,
			},
			"cluster": schema.StringAttribute{
				Optional: true,
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"daemon_task_definition": schema.StringAttribute{
				Required: true,
			},
			"enable_ecs_managed_tags": schema.BoolAttribute{
				Optional: true,
				Computed: true,
				Default:  booldefault.StaticBool(false),
			},
			"enable_execute_command": schema.BoolAttribute{
				Optional: true,
				Computed: true,
				Default:  booldefault.StaticBool(false),
			},
			names.AttrName: schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"propagate_tags": schema.StringAttribute{
				Optional:   true,
				CustomType: fwtypes.StringEnumType[awstypes.DaemonPropagateTags](),
			},
			names.AttrStatus: schema.StringAttribute{
				Computed: true,
			},
			names.AttrTags:    tftags.TagsAttribute(),
			names.AttrTagsAll: tftags.TagsAttributeComputedOnly(),
		},
		Blocks: map[string]schema.Block{
			"deployment_configuration": schema.ListNestedBlock{
				Validators: []validator.List{
					listvalidator.SizeAtMost(1),
				},
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"bake_time_in_minutes": schema.Int64Attribute{
							Optional: true,
							Validators: []validator.Int64{
								int64validator.Between(0, 1440),
							},
						},
						"drain_percent": schema.Float64Attribute{
							Optional: true,
							Validators: []validator.Float64{
								float64validator.Between(0.0, 100.0),
							},
						},
					},
					Blocks: map[string]schema.Block{
						"alarms": schema.ListNestedBlock{
							Validators: []validator.List{
								listvalidator.SizeAtMost(1),
							},
							NestedObject: schema.NestedBlockObject{
								Attributes: map[string]schema.Attribute{
									"alarm_names": schema.SetAttribute{
										Required:    true,
										ElementType: types.StringType,
									},
									"enable": schema.BoolAttribute{
										Required: true,
									},
								},
							},
						},
					},
				},
			},
		},
	}

	s.Blocks[names.AttrTimeouts] = timeouts.Block(ctx, timeouts.Opts{
		Create: true,
		Update: true,
		Delete: true,
	})

	response.Schema = s
}

func (r *daemonResource) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	var plan daemonResourceModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &plan)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().ECSClient(ctx)

	var input ecs.CreateDaemonInput
	response.Diagnostics.Append(fwflex.Expand(ctx, plan, &input)...)
	if response.Diagnostics.HasError() {
		return
	}

	// Fields AutoFlex can't handle
	input.ClientToken = aws.String(id.UniqueId())
	input.Tags = getTagsIn(ctx)
	if len(plan.DeploymentConfiguration) > 0 {
		input.DeploymentConfiguration = expandDaemonDeploymentConfigurationFromModel(plan.DeploymentConfiguration[0])
	}

	output, err := conn.CreateDaemon(ctx, &input)
	if err != nil {
		response.Diagnostics.AddError("creating ECS Daemon ("+plan.DaemonName.ValueString()+")", err.Error())
		return
	}

	plan.DaemonArn = types.StringValue(aws.ToString(output.DaemonArn))

	createTimeout := r.CreateTimeout(ctx, plan.Timeouts)
	if _, err := waitDaemonActive(ctx, conn, plan.DaemonArn.ValueString(), createTimeout); err != nil {
		response.Diagnostics.AddError("waiting for ECS Daemon ("+plan.DaemonArn.ValueString()+") create", err.Error())
		return
	}

	origDTD := plan.DaemonTaskDefinitionArn

	daemon, err := findDaemonByARN(ctx, conn, plan.DaemonArn.ValueString())
	if err != nil {
		response.Diagnostics.AddError("reading ECS Daemon ("+plan.DaemonArn.ValueString()+")", err.Error())
		return
	}

	response.Diagnostics.Append(flattenDaemon(ctx, conn, daemon, &plan)...)
	if response.Diagnostics.HasError() {
		return
	}

	// Preserve the planned value — the API may report a stale revision
	// while the new revision is still rolling out.
	plan.DaemonTaskDefinitionArn = origDTD

	response.Diagnostics.Append(response.State.Set(ctx, &plan)...)
}

func (r *daemonResource) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	var state daemonResourceModel
	response.Diagnostics.Append(request.State.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().ECSClient(ctx)

	daemon, err := findDaemonByARN(ctx, conn, state.DaemonArn.ValueString())
	if tfresource.NotFound(err) {
		response.Diagnostics.Append(fwdiag.NewResourceNotFoundWarningDiagnostic(err))
		response.State.RemoveResource(ctx)
		return
	}
	if err != nil {
		response.Diagnostics.AddError("reading ECS Daemon ("+state.DaemonArn.ValueString()+")", err.Error())
		return
	}

	origDTD := state.DaemonTaskDefinitionArn

	response.Diagnostics.Append(flattenDaemon(ctx, conn, daemon, &state)...)
	if response.Diagnostics.HasError() {
		return
	}

	// Preserve the task definition ARN from state — the API's CurrentRevisions
	// may not reflect the latest update while a deployment is in progress.
	if !origDTD.IsNull() {
		state.DaemonTaskDefinitionArn = origDTD
	}

	// Set defaults for write-only fields not returned by the API (needed for import)
	if state.EnableECSManagedTags.IsNull() {
		state.EnableECSManagedTags = types.BoolValue(false)
	}
	if state.EnableExecuteCommand.IsNull() {
		state.EnableExecuteCommand = types.BoolValue(false)
	}

	response.Diagnostics.Append(response.State.Set(ctx, &state)...)
}

func (r *daemonResource) Update(ctx context.Context, request resource.UpdateRequest, response *resource.UpdateResponse) {
	var plan, state daemonResourceModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &plan)...)
	response.Diagnostics.Append(request.State.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().ECSClient(ctx)

	diff, d := fwflex.Diff(ctx, plan, state)
	response.Diagnostics.Append(d...)
	if response.Diagnostics.HasError() {
		return
	}

	if diff.HasChanges() {

		var input ecs.UpdateDaemonInput
		response.Diagnostics.Append(fwflex.Expand(ctx, plan, &input)...)
		if response.Diagnostics.HasError() {
			return
		}

		// Fields AutoFlex can't handle
		if len(plan.DeploymentConfiguration) > 0 {
			input.DeploymentConfiguration = expandDaemonDeploymentConfigurationFromModel(plan.DeploymentConfiguration[0])
		}

		if !plan.EnableECSManagedTags.Equal(state.EnableECSManagedTags) {
			input.EnableECSManagedTags = plan.EnableECSManagedTags.ValueBool()
		}

		if !plan.EnableExecuteCommand.Equal(state.EnableExecuteCommand) {
			input.EnableExecuteCommand = plan.EnableExecuteCommand.ValueBool()
		}

		_, err := conn.UpdateDaemon(ctx, &input)
		if err != nil {
			response.Diagnostics.AddError("updating ECS Daemon ("+plan.DaemonArn.ValueString()+")", err.Error())
			return
		}

		updateTimeout := r.UpdateTimeout(ctx, plan.Timeouts)
		if _, err := waitDaemonActive(ctx, conn, plan.DaemonArn.ValueString(), updateTimeout); err != nil {
			response.Diagnostics.AddError("waiting for ECS Daemon ("+plan.DaemonArn.ValueString()+") update", err.Error())
			return
		}
	}

	origDTD := plan.DaemonTaskDefinitionArn

	daemon, err := findDaemonByARN(ctx, conn, plan.DaemonArn.ValueString())
	if err != nil {
		response.Diagnostics.AddError("reading ECS Daemon ("+plan.DaemonArn.ValueString()+")", err.Error())
		return
	}

	response.Diagnostics.Append(flattenDaemon(ctx, conn, daemon, &plan)...)
	if response.Diagnostics.HasError() {
		return
	}

	plan.DaemonTaskDefinitionArn = origDTD

	response.Diagnostics.Append(response.State.Set(ctx, &plan)...)
}

func (r *daemonResource) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	var state daemonResourceModel
	response.Diagnostics.Append(request.State.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().ECSClient(ctx)

	log.Printf("[DEBUG] Deleting ECS Daemon: %s", state.DaemonArn.ValueString())
	_, err := conn.DeleteDaemon(ctx, &ecs.DeleteDaemonInput{
		DaemonArn: aws.String(state.DaemonArn.ValueString()),
	})

	if errs.IsA[*awstypes.DaemonNotFoundException](err) || errs.IsAErrorMessageContains[*awstypes.ClientException](err, "not found") {
		return
	}

	if errs.IsAErrorMessageContains[*awstypes.InvalidParameterException](err, "deployment deletion is ongoing") {
		deleteTimeout := r.DeleteTimeout(ctx, state.Timeouts)
		if _, err := waitDaemonDeleted(ctx, conn, state.DaemonArn.ValueString(), deleteTimeout); err != nil {
			response.Diagnostics.AddError("waiting for ECS Daemon ("+state.DaemonArn.ValueString()+") delete", err.Error())
		}
		return
	}

	if err != nil {
		response.Diagnostics.AddError("deleting ECS Daemon ("+state.DaemonArn.ValueString()+")", err.Error())
		return
	}

	deleteTimeout := r.DeleteTimeout(ctx, state.Timeouts)
	if _, err := waitDaemonDeleted(ctx, conn, state.DaemonArn.ValueString(), deleteTimeout); err != nil {
		response.Diagnostics.AddError("waiting for ECS Daemon ("+state.DaemonArn.ValueString()+") delete", err.Error())
	}
}

func newNullObject(typ attr.Type) (obj basetypes.ObjectValue, diags diag.Diagnostics) {
	i, ok := typ.(attr.TypeWithAttributeTypes)
	if !ok {
		diags.AddError(
			"Internal Error",
			"An unexpected error occurred. "+
				"This is always an error in the provider. "+
				"Please report the following to the provider developer:\n\n"+
				fmt.Sprintf("Expected value type to implement attr.TypeWithAttributeTypes, got: %T", typ),
		)
		return
	}

	attrTypes := i.AttributeTypes()
	obj = basetypes.NewObjectNull(attrTypes)

	return obj, diags
}

// flattenDaemon populates the model from a DaemonDetail using AutoFlex for matching fields
// and manual handling for fields that require additional API calls or transformation.
func flattenDaemon(ctx context.Context, conn *ecs.Client, daemon *awstypes.DaemonDetail, model *daemonResourceModel) diag.Diagnostics {
	var diags diag.Diagnostics

	// AutoFlex handles DaemonArn, ClusterArn, Status
	diags.Append(fwflex.Flatten(ctx, daemon, model)...)
	if diags.HasError() {
		return diags
	}

	// Manual: extract daemon name from ARN
	if daemon.DaemonArn != nil {
		arnParts := strings.Split(aws.ToString(daemon.DaemonArn), "/")
		if len(arnParts) >= 3 {
			model.DaemonName = types.StringValue(arnParts[len(arnParts)-1])
		}
	}

	// Manual: get task definition and capacity providers from current revision
	if len(daemon.CurrentRevisions) > 0 {
		currentRevision := daemon.CurrentRevisions[0]

		if currentRevision.Arn != nil {
			revision, err := findDaemonRevisionByARN(ctx, conn, aws.ToString(currentRevision.Arn))
			if err != nil {
				diags.AddError("reading ECS Daemon Revision ("+aws.ToString(currentRevision.Arn)+")", err.Error())
				return diags
			}
			model.DaemonTaskDefinitionArn = types.StringPointerValue(revision.DaemonTaskDefinitionArn)
		}

		if len(currentRevision.CapacityProviders) > 0 {
			cpArns := make([]string, 0, len(currentRevision.CapacityProviders))
			for _, cp := range currentRevision.CapacityProviders {
				if cp.Arn != nil {
					cpArns = append(cpArns, aws.ToString(cp.Arn))
				}
			}
			model.CapacityProviderArns = fwflex.FlattenFrameworkStringValueListOfString(ctx, cpArns)
		}
	}

	return diags
}

func expandDaemonDeploymentConfigurationFromModel(m deploymentConfigurationModel) *awstypes.DaemonDeploymentConfiguration {
	apiObject := &awstypes.DaemonDeploymentConfiguration{}

	if !m.DrainPercent.IsNull() {
		apiObject.DrainPercent = aws.Float64(m.DrainPercent.ValueFloat64())
	}

	if !m.BakeTimeInMinutes.IsNull() {
		apiObject.BakeTimeInMinutes = int32(m.BakeTimeInMinutes.ValueInt64())
	}

	if len(m.Alarms) > 0 {
		alarm := m.Alarms[0]
		apiObject.Alarms = &awstypes.DaemonAlarmConfiguration{
			Enable: alarm.Enable.ValueBool(),
		}
		if !alarm.AlarmNames.IsNull() {
			var alarmNames []string
			for _, v := range alarm.AlarmNames.Elements() {
				if sv, ok := v.(types.String); ok {
					alarmNames = append(alarmNames, sv.ValueString())
				}
			}
			apiObject.Alarms.AlarmNames = alarmNames
		}
	}

	return apiObject
}

type daemonResourceModel struct {
	DaemonArn               types.String                                     `tfsdk:"arn"`
	CapacityProviderArns    fwtypes.ListOfString                             `tfsdk:"capacity_provider_arns"`
	ClusterArn              types.String                                     `tfsdk:"cluster"`
	DaemonTaskDefinitionArn types.String                                     `tfsdk:"daemon_task_definition"`
	DeploymentConfiguration []deploymentConfigurationModel                   `tfsdk:"deployment_configuration" autoflex:"-"`
	EnableECSManagedTags    types.Bool                                       `tfsdk:"enable_ecs_managed_tags"`
	EnableExecuteCommand    types.Bool                                       `tfsdk:"enable_execute_command"`
	DaemonName              types.String                                     `tfsdk:"name"`
	PropagateTags           fwtypes.StringEnum[awstypes.DaemonPropagateTags] `tfsdk:"propagate_tags"`
	Region                  types.String                                     `tfsdk:"region"`
	Status                  types.String                                     `tfsdk:"status"`
	Tags                    tftags.Map                                       `tfsdk:"tags"`
	TagsAll                 tftags.Map                                       `tfsdk:"tags_all"`
	Timeouts                timeouts.Value                                   `tfsdk:"timeouts"`
}

type deploymentConfigurationModel struct {
	Alarms            []alarmConfigurationModel `tfsdk:"alarms"`
	BakeTimeInMinutes types.Int64               `tfsdk:"bake_time_in_minutes"`
	DrainPercent      types.Float64             `tfsdk:"drain_percent"`
}

type alarmConfigurationModel struct {
	AlarmNames types.Set  `tfsdk:"alarm_names"`
	Enable     types.Bool `tfsdk:"enable"`
}

// Finder, status, and waiter functions.

func findDaemonByARN(ctx context.Context, conn *ecs.Client, arn string) (*awstypes.DaemonDetail, error) {
	input := &ecs.DescribeDaemonInput{
		DaemonArn: aws.String(arn),
	}

	output, err := conn.DescribeDaemon(ctx, input)

	if errs.IsA[*awstypes.DaemonNotFoundException](err) {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil || output.Daemon == nil {
		return nil, tfresource.NewEmptyResultError()
	}

	if output.Daemon.Status == "DELETE_IN_PROGRESS" {
		return nil, &retry.NotFoundError{
			Message:     "DELETE_IN_PROGRESS",
			LastRequest: input,
		}
	}

	return output.Daemon, nil
}

func findDaemons(ctx context.Context, conn *ecs.Client, input *ecs.ListDaemonsInput) ([]awstypes.DaemonSummary, error) {
	var result []awstypes.DaemonSummary

	err := listDaemonsPages(ctx, conn, input, func(page *ecs.ListDaemonsOutput, lastPage bool) bool {
		result = append(result, page.DaemonSummariesList...)
		return !lastPage
	})

	if err != nil {
		return nil, err
	}

	return result, nil
}

func statusDaemon(ctx context.Context, conn *ecs.Client, arn string) retry.StateRefreshFunc {
	return func() (any, string, error) {
		output, err := findDaemonByARN(ctx, conn, arn)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, string(output.Status), nil
	}
}

func waitDaemonActive(ctx context.Context, conn *ecs.Client, arn string, timeout time.Duration) (*awstypes.DaemonDetail, error) {
	stateConf := &retry.StateChangeConf{
		Pending: []string{},
		Target:  enum.Slice(awstypes.DaemonStatusActive),
		Refresh: statusDaemon(ctx, conn, arn),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*awstypes.DaemonDetail); ok {
		return output, err
	}

	return nil, err
}

func waitDaemonDeleted(ctx context.Context, conn *ecs.Client, arn string, timeout time.Duration) (*awstypes.DaemonDetail, error) {
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(awstypes.DaemonStatusActive, "DELETE_IN_PROGRESS"),
		Target:  []string{},
		Refresh: statusDaemon(ctx, conn, arn),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*awstypes.DaemonDetail); ok {
		return output, err
	}

	return nil, err
}

func findDaemonRevisionByARN(ctx context.Context, conn *ecs.Client, arn string) (*awstypes.DaemonRevision, error) {
	input := &ecs.DescribeDaemonRevisionsInput{
		DaemonRevisionArns: []string{arn},
	}

	output, err := conn.DescribeDaemonRevisions(ctx, input)

	if err != nil {
		return nil, err
	}

	return tfresource.AssertSingleValueResult(output.DaemonRevisions)
}
