// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package batch

import (
	"context"
	"errors"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go/service/batch"
	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework-validators/resourcevalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/fwdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	"github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkResource("aws_batch_job_queue", name="Job Queue")
// @Tags(identifierAttribute="arn")
// @Testing(existsType="github.com/aws/aws-sdk-go/service/batch;batch.JobQueueDetail")
func newResourceJobQueue(_ context.Context) (resource.ResourceWithConfigure, error) {
	r := resourceJobQueue{}

	r.SetDefaultCreateTimeout(10 * time.Minute)
	r.SetDefaultUpdateTimeout(10 * time.Minute)
	r.SetDefaultDeleteTimeout(10 * time.Minute)

	return &r, nil
}

const (
	ResNameJobQueue = "Job Queue"
	queueNotFound   = "does not exist"
)

type resourceJobQueue struct {
	framework.ResourceWithConfigure
	framework.WithTimeouts
}

func (r *resourceJobQueue) ConfigValidators(_ context.Context) []resource.ConfigValidator {
	return []resource.ConfigValidator{
		resourcevalidator.ExactlyOneOf(
			path.MatchRoot("compute_environments"),
			path.MatchRoot("compute_environment_order"),
		),
	}
}

func (r *resourceJobQueue) Metadata(_ context.Context, request resource.MetadataRequest, response *resource.MetadataResponse) {
	response.TypeName = "aws_batch_job_queue"
}

func (r *resourceJobQueue) Schema(ctx context.Context, request resource.SchemaRequest, response *resource.SchemaResponse) {
	s := schema.Schema{
		Version: 1,
		Attributes: map[string]schema.Attribute{
			names.AttrARN: framework.ARNAttributeComputedOnly(),
			"compute_environments": schema.ListAttribute{
				ElementType:        fwtypes.ARNType,
				Optional:           true,
				DeprecationMessage: "This parameter will be replaced by `compute_environment_order`.",
			},
			names.AttrID: framework.IDAttribute(),
			names.AttrName: schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
				Validators: []validator.String{
					stringvalidator.RegexMatches(regexache.MustCompile(`^[0-9A-Za-z]{1}[0-9A-Za-z_-]{0,127}$`),
						"must be up to 128 letters (uppercase and lowercase), numbers, underscores and dashes, and must start with an alphanumeric"),
				},
			},
			names.AttrPriority: schema.Int64Attribute{
				Required: true,
			},
			"scheduling_policy_arn": schema.StringAttribute{
				CustomType: fwtypes.ARNType,
				Optional:   true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			names.AttrState: schema.StringAttribute{
				Required: true,
				Validators: []validator.String{
					stringvalidator.OneOfCaseInsensitive(batch.JQState_Values()...),
				},
			},
			names.AttrTags:    tftags.TagsAttribute(),
			names.AttrTagsAll: tftags.TagsAttributeComputedOnly(),
		},
	}

	s.Blocks = map[string]schema.Block{
		names.AttrTimeouts: timeouts.Block(ctx, timeouts.Opts{
			Create: true,
			Update: true,
			Delete: true,
		}),
		"compute_environment_order": schema.ListNestedBlock{
			CustomType: fwtypes.NewListNestedObjectTypeOf[computeEnvironmentOrder](ctx),
			NestedObject: schema.NestedBlockObject{
				Attributes: map[string]schema.Attribute{
					"order": schema.Int64Attribute{
						Required: true,
					},
					"compute_environment": schema.StringAttribute{
						CustomType: fwtypes.ARNType,
						Required:   true,
					},
				},
			},
		},
	}

	response.Schema = s
}

func (r *resourceJobQueue) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	conn := r.Meta().BatchConn(ctx)
	var data resourceJobQueueData

	response.Diagnostics.Append(request.Plan.Get(ctx, &data)...)

	if response.Diagnostics.HasError() {
		return
	}

	input := batch.CreateJobQueueInput{
		JobQueueName: flex.StringFromFramework(ctx, data.JobQueueName),
		Priority:     flex.Int64FromFramework(ctx, data.Priority),
		State:        flex.StringFromFramework(ctx, data.State),
		Tags:         getTagsIn(ctx),
	}

	if !data.ComputeEnvironmentOrder.IsNull() {
		flex.Expand(ctx, data.ComputeEnvironmentOrder, &input.ComputeEnvironmentOrder)
	} else {
		input.ComputeEnvironmentOrder = expandComputeEnvironments(flex.ExpandFrameworkStringValueList(ctx, data.ComputeEnvironments))
	}
	if !data.SchedulingPolicyARN.IsNull() {
		input.SchedulingPolicyArn = flex.StringFromFramework(ctx, data.SchedulingPolicyARN)
	}

	output, err := conn.CreateJobQueueWithContext(ctx, &input)

	if err != nil {
		response.Diagnostics.AddError(
			create.ProblemStandardMessage(names.Batch, create.ErrActionCreating, ResNameJobQueue, data.JobQueueName.ValueString(), nil),
			err.Error(),
		)
		return
	}

	state := data
	state.ID = flex.StringToFramework(ctx, output.JobQueueArn)

	createTimeout := r.CreateTimeout(ctx, data.Timeouts)
	out, err := waitJobQueueCreated(ctx, conn, data.JobQueueName.ValueString(), createTimeout)

	if err != nil {
		response.Diagnostics.AddError(
			create.ProblemStandardMessage(names.Batch, create.ErrActionWaitingForCreation, ResNameJobQueue, data.JobQueueName.ValueString(), nil),
			err.Error(),
		)
		return
	}

	if !data.ComputeEnvironmentOrder.IsNull() {
		flex.Flatten(ctx, out.ComputeEnvironmentOrder, &data.ComputeEnvironmentOrder)
	} else {
		state.ComputeEnvironments = flex.FlattenFrameworkStringValueListLegacy(ctx, flattenComputeEnvironments(out.ComputeEnvironmentOrder))
	}
	response.Diagnostics.Append(state.refreshFromOutput(ctx, out)...)
	response.Diagnostics.Append(response.State.Set(ctx, &state)...)
}

func (r *resourceJobQueue) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	conn := r.Meta().BatchConn(ctx)
	var data resourceJobQueueData

	response.Diagnostics.Append(request.State.Get(ctx, &data)...)

	if response.Diagnostics.HasError() {
		return
	}

	out, err := findJobQueueByName(ctx, conn, data.ID.ValueString())

	if err != nil {
		response.Diagnostics.AddError(
			create.ProblemStandardMessage(names.Batch, create.ErrActionUpdating, ResNameJobQueue, data.JobQueueName.ValueString(), err),
			err.Error(),
		)
		return
	}

	if out == nil {
		response.Diagnostics.Append(fwdiag.NewResourceNotFoundWarningDiagnostic(errors.New("not found")))
		response.State.RemoveResource(ctx)
		return
	}

	if !data.ComputeEnvironmentOrder.IsNull() {
		flex.Flatten(ctx, out.ComputeEnvironmentOrder, &data.ComputeEnvironmentOrder)
	} else {
		data.ComputeEnvironments = flex.FlattenFrameworkStringValueListLegacy(ctx, flattenComputeEnvironments(out.ComputeEnvironmentOrder))
	}
	response.Diagnostics.Append(data.refreshFromOutput(ctx, out)...)
	response.Diagnostics.Append(response.State.Set(ctx, &data)...)
}

func (r *resourceJobQueue) Update(ctx context.Context, request resource.UpdateRequest, response *resource.UpdateResponse) {
	conn := r.Meta().BatchConn(ctx)
	var plan, state resourceJobQueueData

	response.Diagnostics.Append(request.State.Get(ctx, &state)...)
	response.Diagnostics.Append(request.Plan.Get(ctx, &plan)...)

	if response.Diagnostics.HasError() {
		return
	}

	var update bool
	input := &batch.UpdateJobQueueInput{
		JobQueue: flex.StringFromFramework(ctx, plan.JobQueueName),
	}

	if !plan.ComputeEnvironmentOrder.IsNull() && !plan.ComputeEnvironmentOrder.Equal(state.ComputeEnvironmentOrder) {
		flex.Expand(ctx, plan.ComputeEnvironmentOrder, &input.ComputeEnvironmentOrder)
		update = true
	} else {
		if !plan.ComputeEnvironments.Equal(state.ComputeEnvironments) {
			ceo := flex.ExpandFrameworkStringValueList(ctx, plan.ComputeEnvironments)
			input.ComputeEnvironmentOrder = expandComputeEnvironments(ceo)
			update = true
		}
	}

	if !plan.Priority.Equal(state.Priority) {
		input.Priority = flex.Int64FromFramework(ctx, plan.Priority)

		update = true
	}

	if !plan.State.Equal(state.State) {
		input.State = flex.StringFromFramework(ctx, plan.State)

		update = true
	}

	if !state.SchedulingPolicyARN.IsNull() {
		input.SchedulingPolicyArn = flex.StringFromFramework(ctx, state.SchedulingPolicyARN)
		update = true
	}

	if !plan.SchedulingPolicyARN.Equal(state.SchedulingPolicyARN) {
		if !plan.SchedulingPolicyARN.IsNull() || !plan.SchedulingPolicyARN.IsUnknown() {
			input.SchedulingPolicyArn = flex.StringFromFramework(ctx, plan.SchedulingPolicyARN)

			update = true
		} else {
			response.Diagnostics.AddError(
				"cannot remove the fair share scheduling policy",
				"cannot remove scheduling policy",
			)
			return
		}
	}

	if update {
		_, err := conn.UpdateJobQueueWithContext(ctx, input)

		if err != nil {
			response.Diagnostics.AddError(
				create.ProblemStandardMessage(names.Batch, create.ErrActionUpdating, ResNameJobQueue, plan.JobQueueName.ValueString(), nil),
				err.Error(),
			)
			return
		}

		updateTimeout := r.UpdateTimeout(ctx, plan.Timeouts)
		out, err := waitJobQueueUpdated(ctx, conn, plan.ID.ValueString(), updateTimeout)

		if err != nil {
			response.Diagnostics.AddError(
				create.ProblemStandardMessage(names.Batch, create.ErrActionWaitingForCreation, ResNameJobQueue, plan.JobQueueName.ValueString(), nil),
				err.Error(),
			)
			return
		}

		response.Diagnostics.Append(plan.refreshFromOutput(ctx, out)...)
	}

	response.Diagnostics.Append(response.State.Set(ctx, &plan)...)
}

func (r *resourceJobQueue) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	conn := r.Meta().BatchConn(ctx)
	var data resourceJobQueueData

	response.Diagnostics.Append(request.State.Get(ctx, &data)...)

	if response.Diagnostics.HasError() {
		return
	}

	deleteTimeout := r.DeleteTimeout(ctx, data.Timeouts)
	found, err := disableJobQueue(ctx, conn, data.ID.ValueString(), deleteTimeout)

	if !found {
		return
	}

	if err != nil {
		response.Diagnostics.AddError(
			create.ProblemStandardMessage(names.Batch, create.ErrActionDeleting, ResNameJobQueue, data.JobQueueName.ValueString(), nil),
			err.Error(),
		)
		return
	}

	_, err = conn.DeleteJobQueueWithContext(ctx, &batch.DeleteJobQueueInput{
		JobQueue: flex.StringFromFramework(ctx, data.ID),
	})

	if err != nil {
		response.Diagnostics.AddError(
			create.ProblemStandardMessage(names.Batch, create.ErrActionDeleting, ResNameJobQueue, data.JobQueueName.ValueString(), nil),
			err.Error(),
		)
		return
	}

	_, err = waitJobQueueDeleted(ctx, conn, data.ID.ValueString(), deleteTimeout)

	if err != nil {
		response.Diagnostics.AddError(
			create.ProblemStandardMessage(names.Batch, create.ErrActionWaitingForDeletion, ResNameJobQueue, data.JobQueueName.ValueString(), nil),
			err.Error(),
		)
		return
	}
}

func (r *resourceJobQueue) ImportState(ctx context.Context, request resource.ImportStateRequest, response *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root(names.AttrID), request, response)
}

func (r *resourceJobQueue) ModifyPlan(ctx context.Context, request resource.ModifyPlanRequest, response *resource.ModifyPlanResponse) {
	r.SetTagsAll(ctx, request, response)
}

func (r *resourceJobQueue) UpgradeState(ctx context.Context) map[int64]resource.StateUpgrader {
	schemaV0 := jobQueueSchema0(ctx)

	return map[int64]resource.StateUpgrader{
		0: {
			PriorSchema:   &schemaV0,
			StateUpgrader: upgradeJobQueueResourceStateV0toV1,
		},
	}
}

type resourceJobQueueData struct {
	ARN                     types.String                                             `tfsdk:"arn"`
	ComputeEnvironments     types.List                                               `tfsdk:"compute_environments"`
	ComputeEnvironmentOrder fwtypes.ListNestedObjectValueOf[computeEnvironmentOrder] `tfsdk:"compute_environment_order"`
	ID                      types.String                                             `tfsdk:"id"`
	JobQueueName            types.String                                             `tfsdk:"name"`
	Priority                types.Int64                                              `tfsdk:"priority"`
	SchedulingPolicyARN     fwtypes.ARN                                              `tfsdk:"scheduling_policy_arn"`
	State                   types.String                                             `tfsdk:"state"`
	Tags                    types.Map                                                `tfsdk:"tags"`
	TagsAll                 types.Map                                                `tfsdk:"tags_all"`
	Timeouts                timeouts.Value                                           `tfsdk:"timeouts"`
}

type computeEnvironmentOrder struct {
	ComputeEnvironment fwtypes.ARN `tfsdk:"compute_environment"`
	Order              types.Int64 `tfsdk:"order"`
}

func (r *resourceJobQueueData) refreshFromOutput(ctx context.Context, out *batch.JobQueueDetail) diag.Diagnostics { //nolint:unparam
	var diags diag.Diagnostics

	r.ARN = flex.StringToFrameworkLegacy(ctx, out.JobQueueArn)
	r.JobQueueName = flex.StringToFramework(ctx, out.JobQueueName)
	r.Priority = flex.Int64ToFrameworkLegacy(ctx, out.Priority)
	r.SchedulingPolicyARN = flex.StringToFrameworkARN(ctx, out.SchedulingPolicyArn)
	r.State = flex.StringToFrameworkLegacy(ctx, out.State)

	setTagsOut(ctx, out.Tags)

	return diags
}

func expandComputeEnvironments(order []string) (envs []*batch.ComputeEnvironmentOrder) {
	for i, env := range order {
		envs = append(envs, &batch.ComputeEnvironmentOrder{
			Order:              aws.Int64(int64(i)),
			ComputeEnvironment: aws.String(env),
		})
	}
	return
}

func flattenComputeEnvironments(apiObject []*batch.ComputeEnvironmentOrder) []string {
	sort.Slice(apiObject, func(i, j int) bool {
		return aws.ToInt64(apiObject[i].Order) < aws.ToInt64(apiObject[j].Order)
	})

	computeEnvironments := make([]string, 0, len(apiObject))
	for _, v := range apiObject {
		computeEnvironments = append(computeEnvironments, aws.ToString(v.ComputeEnvironment))
	}

	return computeEnvironments
}

func findJobQueueByName(ctx context.Context, conn *batch.Batch, sn string) (*batch.JobQueueDetail, error) {
	describeOpts := &batch.DescribeJobQueuesInput{
		JobQueues: []*string{aws.String(sn)},
	}
	resp, err := conn.DescribeJobQueuesWithContext(ctx, describeOpts)
	if err != nil {
		return nil, err
	}

	numJobQueues := len(resp.JobQueues)
	switch {
	case numJobQueues == 0:
		return nil, nil
	case numJobQueues == 1:
		return resp.JobQueues[0], nil
	case numJobQueues > 1:
		return nil, fmt.Errorf("Multiple Job Queues with name %s", sn)
	}
	return nil, nil
}

func disableJobQueue(ctx context.Context, conn *batch.Batch, id string, timeout time.Duration) (bool, error) {
	_, err := conn.UpdateJobQueueWithContext(ctx, &batch.UpdateJobQueueInput{
		JobQueue: aws.String(id),
		State:    aws.String(batch.JQStateDisabled),
	})

	if err != nil {
		if strings.Contains(err.Error(), queueNotFound) {
			return false, nil
		}
		return true, err
	}

	stateChangeConf := &retry.StateChangeConf{
		Pending:    []string{batch.JQStatusUpdating},
		Target:     []string{batch.JQStatusValid},
		Refresh:    jobQueueRefreshStatusFunc(ctx, conn, id),
		Timeout:    timeout,
		Delay:      10 * time.Second,
		MinTimeout: 3 * time.Second,
	}
	_, err = stateChangeConf.WaitForStateContext(ctx)
	return true, err
}

func waitJobQueueCreated(ctx context.Context, conn *batch.Batch, id string, timeout time.Duration) (*batch.JobQueueDetail, error) {
	stateConf := &retry.StateChangeConf{
		Pending:    []string{batch.JQStatusCreating, batch.JQStatusUpdating},
		Target:     []string{batch.JQStatusValid},
		Refresh:    jobQueueRefreshStatusFunc(ctx, conn, id),
		Timeout:    timeout,
		MinTimeout: 10 * time.Second,
		Delay:      30 * time.Second,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*batch.JobQueueDetail); ok {
		return output, err
	}

	return nil, err
}

func waitJobQueueUpdated(ctx context.Context, conn *batch.Batch, id string, timeout time.Duration) (*batch.JobQueueDetail, error) {
	stateConf := &retry.StateChangeConf{
		Pending:    []string{batch.JQStatusUpdating},
		Target:     []string{batch.JQStatusValid},
		Refresh:    jobQueueRefreshStatusFunc(ctx, conn, id),
		Timeout:    timeout,
		MinTimeout: 10 * time.Second,
		Delay:      30 * time.Second,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*batch.JobQueueDetail); ok {
		return output, err
	}

	return nil, err
}

func waitJobQueueDeleted(ctx context.Context, conn *batch.Batch, id string, timeout time.Duration) (*batch.JobQueueDetail, error) {
	stateConf := &retry.StateChangeConf{
		Pending:    []string{batch.JQStateDisabled, batch.JQStatusDeleting},
		Target:     []string{batch.JQStatusDeleted},
		Refresh:    jobQueueRefreshStatusFunc(ctx, conn, id),
		Timeout:    timeout,
		MinTimeout: 10 * time.Second,
		Delay:      30 * time.Second,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*batch.JobQueueDetail); ok {
		return output, err
	}

	return nil, err
}

func jobQueueRefreshStatusFunc(ctx context.Context, conn *batch.Batch, sn string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		ce, err := findJobQueueByName(ctx, conn, sn)
		if err != nil {
			return nil, "", err
		}

		if ce == nil {
			return 42, batch.JQStatusDeleted, nil
		}

		return ce, *ce.Status, nil
	}
}
