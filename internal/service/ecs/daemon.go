// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package ecs

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/aws/arn"
	"github.com/aws/aws-sdk-go-v2/service/ecs"
	awstypes "github.com/aws/aws-sdk-go-v2/service/ecs/types"
	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework-validators/float64validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/boolplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64default"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/fwdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	fwflex "github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
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
	response.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			names.AttrARN: framework.ARNAttributeComputedOnly(),
			"capacity_provider_arns": schema.SetAttribute{
				CustomType: fwtypes.SetOfStringType,
				Required:   true,
			},
			"cluster_arn": schema.StringAttribute{
				CustomType: fwtypes.ARNType,
				Optional:   true,
				Computed:   true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"daemon_task_definition_arn": schema.StringAttribute{
				CustomType: fwtypes.ARNType,
				Required:   true,
			},
			"deployment_arn": schema.StringAttribute{
				CustomType: fwtypes.ARNType,
				Computed:   true,
			},
			"enable_ecs_managed_tags": schema.BoolAttribute{
				Optional: true,
			},
			"enable_execute_command": schema.BoolAttribute{
				Optional: true,
			},
			names.AttrName: schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
				Validators: []validator.String{
					stringvalidator.LengthBetween(1, 255),
					stringvalidator.RegexMatches(regexache.MustCompile("^[0-9A-Za-z_-]+$"), "must contain only alphanumeric characters, hyphens, and underscores"),
				},
			},
			names.AttrPropagateTags: schema.StringAttribute{
				CustomType: fwtypes.StringEnumType[awstypes.DaemonPropagateTags](),
				Optional:   true,
			},
			names.AttrStatus: schema.StringAttribute{
				CustomType: fwtypes.StringEnumType[awstypes.DaemonStatus](),
				Computed:   true,
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
							Computed: true,
							Default:  int64default.StaticInt64(0),
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
									"alarm_names": schema.ListAttribute{
										CustomType: fwtypes.ListOfStringType,
										Optional:   true,
									},
									"enable": schema.BoolAttribute{
										Optional: true,
										Computed: true,
										PlanModifiers: []planmodifier.Bool{
											boolplanmodifier.UseStateForUnknown(),
										},
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
	input.ClientToken = aws.String(create.UniqueId(ctx))
	input.Tags = getTagsIn(ctx)

	output, err := createDaemonWithRetry(ctx, conn, &input)

	// Some partitions (e.g. ISO) may not support tag-on-create.
	if input.Tags != nil && errs.IsUnsupportedOperationInPartitionError(r.Meta().Partition(ctx), err) {
		input.Tags = nil
		output, err = createDaemonWithRetry(ctx, conn, &input)
	}

	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("creating ECS Daemon (%s)", plan.DaemonName.ValueString()), err.Error())
		return
	}

	if output == nil || output.DaemonArn == nil {
		response.Diagnostics.AddError(fmt.Sprintf("creating ECS Daemon (%s)", plan.DaemonName.ValueString()), "empty output from API")
		return
	}

	// For partitions not supporting tag-on-create, attempt tag after create.
	if tags := getTagsIn(ctx); input.Tags == nil && len(tags) > 0 {
		if err := createTags(ctx, conn, aws.ToString(output.DaemonArn), tags); err != nil {
			response.Diagnostics.AddError(fmt.Sprintf("setting ECS Daemon (%s) tags", aws.ToString(output.DaemonArn)), err.Error())
			return
		}
	}

	plan.DaemonArn = types.StringValue(aws.ToString(output.DaemonArn))

	// Save ARN to state so Terraform can track the resource if the waiter times out.
	response.State.SetAttribute(ctx, path.Root(names.AttrARN), output.DaemonArn)

	createTimeout := r.CreateTimeout(ctx, plan.Timeouts)
	if err := waitDaemonActive(ctx, conn, plan.DaemonArn.ValueString(), createTimeout); err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("waiting for ECS Daemon (%s) create", plan.DaemonArn.ValueString()), err.Error())
		return
	}

	outputFind, err := findDaemonByARN(ctx, conn, plan.DaemonArn.ValueString())
	if retry.NotFound(err) {
		response.Diagnostics.Append(fwdiag.NewResourceNotFoundWarningDiagnostic(err))
		response.State.RemoveResource(ctx)
		return
	}
	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("reading ECS Daemon (%s)", plan.DaemonArn.ValueString()), err.Error())
		return
	}

	response.Diagnostics.Append(fwflex.Flatten(ctx, outputFind, &plan)...)
	plan.DaemonName = daemonNameFromARN(plan.DaemonArn.ValueString())
	if response.Diagnostics.HasError() {
		return
	}

	flattenDaemonCurrentRevision(ctx, conn, outputFind, &response.Diagnostics, &plan)
	if response.Diagnostics.HasError() {
		return
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

	output, err := findDaemonByARN(ctx, conn, state.DaemonArn.ValueString())
	if retry.NotFound(err) {
		response.Diagnostics.Append(fwdiag.NewResourceNotFoundWarningDiagnostic(err))
		response.State.RemoveResource(ctx)
		return
	}
	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("reading ECS Daemon (%s)", state.DaemonArn.ValueString()), err.Error())
		return
	}

	response.Diagnostics.Append(fwflex.Flatten(ctx, output, &state)...)
	state.DaemonName = daemonNameFromARN(state.DaemonArn.ValueString())
	if response.Diagnostics.HasError() {
		return
	}

	flattenDaemonCurrentRevision(ctx, conn, output, &response.Diagnostics, &state)
	if response.Diagnostics.HasError() {
		return
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

	diff, diags := fwflex.Diff(ctx, plan, state,
		fwflex.WithIgnoredField("EnableECSManagedTags"),
		fwflex.WithIgnoredField("EnableExecuteCommand"),
		fwflex.WithIgnoredField("PropagateTags"),
	)
	response.Diagnostics.Append(diags...)
	if response.Diagnostics.HasError() {
		return
	}

	if diff.HasChanges() {
		var input ecs.UpdateDaemonInput
		response.Diagnostics.Append(fwflex.Expand(ctx, plan, &input)...)
		if response.Diagnostics.HasError() {
			return
		}

		err := updateDaemonWithRetry(ctx, conn, &input)
		if err != nil {
			response.Diagnostics.AddError(fmt.Sprintf("updating ECS Daemon (%s)", plan.DaemonArn.ValueString()), err.Error())
			return
		}

		updateTimeout := r.UpdateTimeout(ctx, plan.Timeouts)
		if err := waitDaemonActive(ctx, conn, plan.DaemonArn.ValueString(), updateTimeout); err != nil {
			response.Diagnostics.AddError(fmt.Sprintf("waiting for ECS Daemon (%s) update", plan.DaemonArn.ValueString()), err.Error())
			return
		}
	}

	output, err := findDaemonByARN(ctx, conn, plan.DaemonArn.ValueString())
	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("reading ECS Daemon (%s)", plan.DaemonArn.ValueString()), err.Error())
		return
	}

	response.Diagnostics.Append(fwflex.Flatten(ctx, output, &plan)...)
	plan.DaemonName = daemonNameFromARN(plan.DaemonArn.ValueString())
	if response.Diagnostics.HasError() {
		return
	}

	flattenDaemonCurrentRevision(ctx, conn, output, &response.Diagnostics, &plan)
	if response.Diagnostics.HasError() {
		return
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

	input := ecs.DeleteDaemonInput{
		DaemonArn: state.DaemonArn.ValueStringPointer(),
	}

	_, err := conn.DeleteDaemon(ctx, &input)
	if errs.IsA[*awstypes.DaemonNotFoundException](err) || errs.IsAErrorMessageContains[*awstypes.ClientException](err, "not found") {
		return
	}

	switch {
	case err == nil, errs.IsAErrorMessageContains[*awstypes.InvalidParameterException](err, "deployment deletion is ongoing"):
		// no-op, continue on to waiter
	default:
		response.Diagnostics.AddError(fmt.Sprintf("deleting ECS Daemon (%s)", state.DaemonArn.ValueString()), err.Error())
		return
	}

	deleteTimeout := r.DeleteTimeout(ctx, state.Timeouts)
	if err := waitDaemonDeleted(ctx, conn, state.DaemonArn.ValueString(), deleteTimeout); err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("waiting for ECS Daemon (%s) delete", state.DaemonArn.ValueString()), err.Error())
	}
}

func daemonNameFromARN(arnStr string) types.String {
	parsed, err := arn.Parse(arnStr)
	if err != nil {
		return types.StringNull()
	}
	parts := strings.Split(parsed.Resource, "/")
	if len(parts) == 3 {
		return types.StringValue(parts[2])
	}
	return types.StringNull()
}

// flattenDaemonRevision populates task definition ARN and capacity
// provider ARNs from a DaemonRevision and DaemonRevisionDetail. DaemonTaskDefinitionArn is only
// set when the model's value is null (e.g., during import) to avoid overwriting
// the plan value with potentially stale revision data during Create/Update.
func flattenDaemonRevision(ctx context.Context, revision *awstypes.DaemonRevision, revisionDetail awstypes.DaemonRevisionDetail, model *daemonResourceModel) { // nosemgrep:ci.semgrep.framework.manual-flattener-functions
	if model.DaemonTaskDefinitionArn.IsNull() {
		model.DaemonTaskDefinitionArn = fwtypes.ARNValue(aws.ToString(revision.DaemonTaskDefinitionArn))
	}

	if len(revisionDetail.CapacityProviders) > 0 {
		cpArns := make([]string, 0, len(revisionDetail.CapacityProviders))
		for _, cp := range revisionDetail.CapacityProviders {
			if cp.Arn != nil {
				cpArns = append(cpArns, aws.ToString(cp.Arn))
			}
		}
		model.CapacityProviderArns = fwflex.FlattenFrameworkStringValueSetOfString(ctx, cpArns)
	}
}

// flattenDaemonCurrentRevision fetches the current revision for a daemon and
// flattens its fields into the model. This is shared across Create, Read,
// Update, and the list resource.
func flattenDaemonCurrentRevision(ctx context.Context, conn *ecs.Client, daemon *awstypes.DaemonDetail, diags *diag.Diagnostics, model *daemonResourceModel) { // nosemgrep:ci.semgrep.framework.manual-flattener-functions
	if len(daemon.CurrentRevisions) == 0 || daemon.CurrentRevisions[0].Arn == nil {
		return
	}
	revision, err := findDaemonRevisionByARN(ctx, conn, aws.ToString(daemon.CurrentRevisions[0].Arn))
	if err != nil {
		diags.AddError(fmt.Sprintf("reading ECS Daemon Revision (%s)", aws.ToString(daemon.CurrentRevisions[0].Arn)), err.Error())
		return
	}
	flattenDaemonRevision(ctx, revision, daemon.CurrentRevisions[0], model)
}

func createDaemonWithRetry(ctx context.Context, conn *ecs.Client, input *ecs.CreateDaemonInput) (*ecs.CreateDaemonOutput, error) {
	output, err := tfresource.RetryWhen[*ecs.CreateDaemonOutput](ctx, propagationTimeout,
		func(ctx context.Context) (*ecs.CreateDaemonOutput, error) {
			return conn.CreateDaemon(ctx, input)
		},
		func(err error) (bool, error) {
			if errs.IsA[*awstypes.ClusterNotFoundException](err) {
				return true, err
			}

			if errs.IsAErrorMessageContains[*awstypes.InvalidParameterException](err, "Unable to assume the service linked role") {
				return true, err
			}

			return false, err
		},
	)
	if err != nil {
		return nil, err
	}

	return output, nil
}

func updateDaemonWithRetry(ctx context.Context, conn *ecs.Client, input *ecs.UpdateDaemonInput) error {
	_, err := tfresource.RetryWhenIsAErrorMessageContains[*ecs.UpdateDaemonOutput, *awstypes.InvalidParameterException](ctx, propagationTimeout,
		func(ctx context.Context) (*ecs.UpdateDaemonOutput, error) {
			return conn.UpdateDaemon(ctx, input)
		},
		"Unable to assume the service linked role",
	)

	return err
}

func findDaemonByARN(ctx context.Context, conn *ecs.Client, arn string) (*awstypes.DaemonDetail, error) {
	input := &ecs.DescribeDaemonInput{
		DaemonArn: aws.String(arn),
	}

	output, err := conn.DescribeDaemon(ctx, input)

	if errs.IsA[*awstypes.DaemonNotFoundException](err) {
		return nil, &retry.NotFoundError{
			LastError: err,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil || output.Daemon == nil {
		return nil, tfresource.NewEmptyResultError()
	}

	if output.Daemon.Status == awstypes.DaemonStatusDeleteInProgress {
		return nil, &retry.NotFoundError{
			Message: string(awstypes.DaemonStatusDeleteInProgress),
		}
	}

	return output.Daemon, nil
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

func statusDaemon(conn *ecs.Client, arn string) retry.StateRefreshFunc {
	return func(ctx context.Context) (any, string, error) {
		output, err := findDaemonByARN(ctx, conn, arn)

		if retry.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, string(output.Status), nil
	}
}

func waitDaemonActive(ctx context.Context, conn *ecs.Client, arn string, timeout time.Duration) error {
	stateConf := &retry.StateChangeConf{
		Pending: []string{},
		Target:  enum.Slice(awstypes.DaemonStatusActive),
		Refresh: statusDaemon(conn, arn),
		Timeout: timeout,
	}

	_, err := stateConf.WaitForStateContext(ctx)

	return err
}

func waitDaemonDeleted(ctx context.Context, conn *ecs.Client, arn string, timeout time.Duration) error {
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(awstypes.DaemonStatusActive, awstypes.DaemonStatusDeleteInProgress),
		Target:  []string{},
		Refresh: statusDaemon(conn, arn),
		Timeout: timeout,
	}

	_, err := stateConf.WaitForStateContext(ctx)

	return err
}

type daemonResourceModel struct {
	framework.WithRegionModel
	DaemonArn               types.String                                                  `tfsdk:"arn"`
	CapacityProviderArns    fwtypes.SetOfString                                           `tfsdk:"capacity_provider_arns"`
	ClusterArn              fwtypes.ARN                                                   `tfsdk:"cluster_arn"`
	DaemonTaskDefinitionArn fwtypes.ARN                                                   `tfsdk:"daemon_task_definition_arn"`
	DeploymentConfiguration fwtypes.ListNestedObjectValueOf[deploymentConfigurationModel] `tfsdk:"deployment_configuration"`
	DeploymentArn           fwtypes.ARN                                                   `tfsdk:"deployment_arn"`
	EnableECSManagedTags    types.Bool                                                    `tfsdk:"enable_ecs_managed_tags"`
	EnableExecuteCommand    types.Bool                                                    `tfsdk:"enable_execute_command"`
	DaemonName              types.String                                                  `tfsdk:"name"`
	PropagateTags           fwtypes.StringEnum[awstypes.DaemonPropagateTags]              `tfsdk:"propagate_tags"`
	Status                  fwtypes.StringEnum[awstypes.DaemonStatus]                     `tfsdk:"status"`
	Tags                    tftags.Map                                                    `tfsdk:"tags"`
	TagsAll                 tftags.Map                                                    `tfsdk:"tags_all"`
	Timeouts                timeouts.Value                                                `tfsdk:"timeouts"`
}

type deploymentConfigurationModel struct {
	Alarms            fwtypes.ListNestedObjectValueOf[alarmConfigurationModel] `tfsdk:"alarms"`
	BakeTimeInMinutes types.Int64                                              `tfsdk:"bake_time_in_minutes"`
	DrainPercent      types.Float64                                            `tfsdk:"drain_percent"`
}

type alarmConfigurationModel struct {
	AlarmNames fwtypes.ListOfString `tfsdk:"alarm_names"`
	Enable     types.Bool           `tfsdk:"enable"`
}
