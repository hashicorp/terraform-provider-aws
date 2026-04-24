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
				CustomType: fwtypes.NewListNestedObjectTypeOf[deploymentConfigurationModel](ctx),
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
							CustomType: fwtypes.NewListNestedObjectTypeOf[alarmConfigurationModel](ctx),
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
			names.AttrTimeouts: timeouts.Block(ctx, timeouts.Opts{
				Create: true,
				Update: true,
				Delete: true,
			}),
		},
	}

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

	output, err := conn.CreateDaemon(ctx, &input)
	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("creating ECS Daemon (%s)", plan.DaemonName.ValueString()), err.Error())
		return
	}

	plan.DaemonArn = types.StringValue(aws.ToString(output.DaemonArn))

	createTimeout := r.CreateTimeout(ctx, plan.Timeouts)
	if _, err := waitDaemonActive(ctx, conn, plan.DaemonArn.ValueString(), createTimeout); err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("waiting for ECS Daemon (%s) create", plan.DaemonArn.ValueString()), err.Error())
		return
	}

	daemon, err := findDaemonByARN(ctx, conn, plan.DaemonArn.ValueString())
	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("reading ECS Daemon (%s)", plan.DaemonArn.ValueString()), err.Error())
		return
	}

	response.Diagnostics.Append(flattenDaemon(ctx, daemon, &plan)...)
	if response.Diagnostics.HasError() {
		return
	}

	if len(daemon.CurrentRevisions) > 0 && daemon.CurrentRevisions[0].Arn != nil {
		revision, err := findDaemonRevisionByARN(ctx, conn, aws.ToString(daemon.CurrentRevisions[0].Arn))
		if err != nil {
			response.Diagnostics.AddError(fmt.Sprintf("reading ECS Daemon Revision (%s)", aws.ToString(daemon.CurrentRevisions[0].Arn)), err.Error())
			return
		}
		flattenDaemonRevision(ctx, revision, daemon.CurrentRevisions[0], &plan)
	}

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
		response.Diagnostics.AddError(fmt.Sprintf("reading ECS Daemon (%s)", state.DaemonArn.ValueString()), err.Error())
		return
	}

	response.Diagnostics.Append(flattenDaemon(ctx, daemon, &state)...)
	if response.Diagnostics.HasError() {
		return
	}

	if len(daemon.CurrentRevisions) > 0 && daemon.CurrentRevisions[0].Arn != nil {
		revision, err := findDaemonRevisionByARN(ctx, conn, aws.ToString(daemon.CurrentRevisions[0].Arn))
		if err != nil {
			response.Diagnostics.AddError(fmt.Sprintf("reading ECS Daemon Revision (%s)", aws.ToString(daemon.CurrentRevisions[0].Arn)), err.Error())
			return
		}
		flattenDaemonRevision(ctx, revision, daemon.CurrentRevisions[0], &state)
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
		if !plan.EnableECSManagedTags.Equal(state.EnableECSManagedTags) {
			input.EnableECSManagedTags = plan.EnableECSManagedTags.ValueBool()
		}

		if !plan.EnableExecuteCommand.Equal(state.EnableExecuteCommand) {
			input.EnableExecuteCommand = plan.EnableExecuteCommand.ValueBool()
		}

		_, err := conn.UpdateDaemon(ctx, &input)
		if err != nil {
			response.Diagnostics.AddError(fmt.Sprintf("updating ECS Daemon (%s)", plan.DaemonArn.ValueString()), err.Error())
			return
		}

		updateTimeout := r.UpdateTimeout(ctx, plan.Timeouts)
		if _, err := waitDaemonActive(ctx, conn, plan.DaemonArn.ValueString(), updateTimeout); err != nil {
			response.Diagnostics.AddError(fmt.Sprintf("waiting for ECS Daemon (%s) update", plan.DaemonArn.ValueString()), err.Error())
			return
		}
	}

	daemon, err := findDaemonByARN(ctx, conn, plan.DaemonArn.ValueString())
	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("reading ECS Daemon (%s)", plan.DaemonArn.ValueString()), err.Error())
		return
	}

	response.Diagnostics.Append(flattenDaemon(ctx, daemon, &plan)...)
	if response.Diagnostics.HasError() {
		return
	}

	if len(daemon.CurrentRevisions) > 0 && daemon.CurrentRevisions[0].Arn != nil {
		revision, err := findDaemonRevisionByARN(ctx, conn, aws.ToString(daemon.CurrentRevisions[0].Arn))
		if err != nil {
			response.Diagnostics.AddError(fmt.Sprintf("reading ECS Daemon Revision (%s)", aws.ToString(daemon.CurrentRevisions[0].Arn)), err.Error())
			return
		}
		flattenDaemonRevision(ctx, revision, daemon.CurrentRevisions[0], &plan)
	}

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

	switch {
	case err == nil:
		// no-op, continue on to waiter
	case errs.IsAErrorMessageContains[*awstypes.InvalidParameterException](err, "deployment deletion is ongoing"):
		// no-op, continue on to waiter
	default:
		response.Diagnostics.AddError(fmt.Sprintf("deleting ECS Daemon (%s)", state.DaemonArn.ValueString()), err.Error())
		return
	}

	deleteTimeout := r.DeleteTimeout(ctx, state.Timeouts)
	if _, err := waitDaemonDeleted(ctx, conn, state.DaemonArn.ValueString(), deleteTimeout); err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("waiting for ECS Daemon (%s) delete", state.DaemonArn.ValueString()), err.Error())
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
func flattenDaemon(ctx context.Context, daemon *awstypes.DaemonDetail, model *daemonResourceModel) diag.Diagnostics {
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

	return diags
}

// flattenDaemonRevision populates task definition ARN and capacity provider ARNs
// from a DaemonRevision and DaemonRevisionDetail. DaemonTaskDefinitionArn is only
// set when the model's value is null (e.g., during import) to avoid overwriting
// the plan value with potentially stale revision data during Create/Update.
func flattenDaemonRevision(ctx context.Context, revision *awstypes.DaemonRevision, revisionDetail awstypes.DaemonRevisionDetail, model *daemonResourceModel) {
	if model.DaemonTaskDefinitionArn.IsNull() {
		model.DaemonTaskDefinitionArn = types.StringPointerValue(revision.DaemonTaskDefinitionArn)
	}

	if len(revisionDetail.CapacityProviders) > 0 {
		cpArns := make([]string, 0, len(revisionDetail.CapacityProviders))
		for _, cp := range revisionDetail.CapacityProviders {
			if cp.Arn != nil {
				cpArns = append(cpArns, aws.ToString(cp.Arn))
			}
		}
		model.CapacityProviderArns = fwflex.FlattenFrameworkStringValueListOfString(ctx, cpArns)
	}
}

type daemonResourceModel struct {
	DaemonArn               types.String                                     `tfsdk:"arn"`
	CapacityProviderArns    fwtypes.ListOfString                             `tfsdk:"capacity_provider_arns"`
	ClusterArn              types.String                                     `tfsdk:"cluster"`
	DaemonTaskDefinitionArn types.String                                     `tfsdk:"daemon_task_definition"`
	DeploymentConfiguration fwtypes.ListNestedObjectValueOf[deploymentConfigurationModel] `tfsdk:"deployment_configuration"`
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
	Alarms            fwtypes.ListNestedObjectValueOf[alarmConfigurationModel] `tfsdk:"alarms"`
	BakeTimeInMinutes types.Int64                                              `tfsdk:"bake_time_in_minutes"`
	DrainPercent      types.Float64                                            `tfsdk:"drain_percent"`
}

type alarmConfigurationModel struct {
	AlarmNames fwtypes.SetOfString `tfsdk:"alarm_names"`
	Enable     types.Bool          `tfsdk:"enable"`
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
