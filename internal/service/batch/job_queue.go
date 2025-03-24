// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package batch

import (
	"cmp"
	"context"
	"errors"
	"fmt"
	"slices"
	"time"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/batch"
	awstypes "github.com/aws/aws-sdk-go-v2/service/batch/types"
	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/resourcevalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/fwdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	fwflex "github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	tfslices "github.com/hashicorp/terraform-provider-aws/internal/slices"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkResource("aws_batch_job_queue", name="Job Queue")
// @Tags(identifierAttribute="arn")
// @Testing(existsType="github.com/aws/aws-sdk-go-v2/service/batch/types;types.JobQueueDetail")
func newJobQueueResource(_ context.Context) (resource.ResourceWithConfigure, error) {
	r := jobQueueResource{}

	r.SetDefaultCreateTimeout(10 * time.Minute)
	r.SetDefaultUpdateTimeout(10 * time.Minute)
	r.SetDefaultDeleteTimeout(10 * time.Minute)

	return &r, nil
}

type jobQueueResource struct {
	framework.ResourceWithConfigure
	framework.WithImportByID
	framework.WithTimeouts
}

func (r *jobQueueResource) Schema(ctx context.Context, request resource.SchemaRequest, response *resource.SchemaResponse) {
	response.Schema = schema.Schema{
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
			},
			names.AttrState: schema.StringAttribute{
				Required: true,
				Validators: []validator.String{
					enum.FrameworkValidateIgnoreCase[awstypes.JQState](),
				},
			},
			names.AttrTags:    tftags.TagsAttribute(),
			names.AttrTagsAll: tftags.TagsAttributeComputedOnly(),
		},
		Blocks: map[string]schema.Block{
			names.AttrTimeouts: timeouts.Block(ctx, timeouts.Opts{
				Create: true,
				Update: true,
				Delete: true,
			}),
			"compute_environment_order": schema.ListNestedBlock{
				CustomType: fwtypes.NewListNestedObjectTypeOf[computeEnvironmentOrderModel](ctx),
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"compute_environment": schema.StringAttribute{
							CustomType: fwtypes.ARNType,
							Required:   true,
						},
						"order": schema.Int64Attribute{
							Required: true,
						},
					},
				},
			},
			"job_state_time_limit_action": schema.ListNestedBlock{
				CustomType: fwtypes.NewListNestedObjectTypeOf[jobStateTimeLimitActionModel](ctx),
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						names.AttrAction: schema.StringAttribute{
							CustomType: fwtypes.StringEnumType[awstypes.JobStateTimeLimitActionsAction](),
							Required:   true,
						},
						"max_time_seconds": schema.Int64Attribute{
							Required: true,
							Validators: []validator.Int64{
								int64validator.Between(600, 86400),
							},
						},
						"reason": schema.StringAttribute{
							Required: true,
						},
						names.AttrState: schema.StringAttribute{
							CustomType: fwtypes.StringEnumType[awstypes.JobStateTimeLimitActionsState](),
							Required:   true,
						},
					},
				},
			},
		},
	}
}

func (r *jobQueueResource) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	var data jobQueueResourceModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().BatchClient(ctx)

	name := data.JobQueueName.ValueString()
	input := &batch.CreateJobQueueInput{
		JobQueueName: aws.String(name),
		Priority:     fwflex.Int32FromFrameworkInt64(ctx, data.Priority),
		State:        awstypes.JQState(data.State.ValueString()),
		Tags:         getTagsIn(ctx),
	}

	if !data.ComputeEnvironmentOrder.IsNull() {
		response.Diagnostics.Append(fwflex.Expand(ctx, data.ComputeEnvironmentOrder, &input.ComputeEnvironmentOrder)...)
		if response.Diagnostics.HasError() {
			return
		}
	} else {
		input.ComputeEnvironmentOrder = expandComputeEnvironments(ctx, data.ComputeEnvironments)
	}
	response.Diagnostics.Append(fwflex.Expand(ctx, data.JobStateTimeLimitActions, &input.JobStateTimeLimitActions)...)
	if response.Diagnostics.HasError() {
		return
	}
	if !data.SchedulingPolicyARN.IsNull() {
		input.SchedulingPolicyArn = fwflex.StringFromFramework(ctx, data.SchedulingPolicyARN)
	}

	output, err := conn.CreateJobQueue(ctx, input)

	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("creating Batch Job Queue (%s)", name), err.Error())

		return
	}

	// Set values for unknowns.
	data.JobQueueARN = fwflex.StringToFramework(ctx, output.JobQueueArn)
	data.setID()

	if _, err := waitJobQueueCreated(ctx, conn, data.ID.ValueString(), r.CreateTimeout(ctx, data.Timeouts)); err != nil {
		response.State.SetAttribute(ctx, path.Root(names.AttrID), data.ID) // Set 'id' so as to taint the resource.
		response.Diagnostics.AddError(fmt.Sprintf("waiting for Batch Job Queue (%s) create", data.ID.ValueString()), err.Error())

		return
	}

	response.Diagnostics.Append(response.State.Set(ctx, data)...)
}

func (r *jobQueueResource) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	var data jobQueueResourceModel
	response.Diagnostics.Append(request.State.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	if err := data.InitFromID(); err != nil {
		response.Diagnostics.AddError("parsing resource ID", err.Error())

		return
	}

	conn := r.Meta().BatchClient(ctx)

	jobQueue, err := findJobQueueByID(ctx, conn, data.ID.ValueString())

	if tfresource.NotFound(err) {
		response.Diagnostics.Append(fwdiag.NewResourceNotFoundWarningDiagnostic(err))
		response.State.RemoveResource(ctx)

		return
	}

	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("reading Batch Job Queue (%s)", data.ID.ValueString()), err.Error())

		return
	}

	// Set attributes for import.
	if !data.ComputeEnvironmentOrder.IsNull() {
		response.Diagnostics.Append(fwflex.Flatten(ctx, jobQueue.ComputeEnvironmentOrder, &data.ComputeEnvironmentOrder)...)
		if response.Diagnostics.HasError() {
			return
		}
	} else {
		data.ComputeEnvironments = flattenComputeEnvironments(ctx, jobQueue.ComputeEnvironmentOrder)
	}

	data.JobQueueARN = fwflex.StringToFrameworkLegacy(ctx, jobQueue.JobQueueArn)
	data.JobQueueName = fwflex.StringToFramework(ctx, jobQueue.JobQueueName)
	response.Diagnostics.Append(fwflex.Flatten(ctx, jobQueue.JobStateTimeLimitActions, &data.JobStateTimeLimitActions)...)
	if response.Diagnostics.HasError() {
		return
	}
	data.Priority = fwflex.Int32ToFrameworkInt64Legacy(ctx, jobQueue.Priority)
	data.SchedulingPolicyARN = fwflex.StringToFrameworkARN(ctx, jobQueue.SchedulingPolicyArn)
	data.State = fwflex.StringValueToFramework(ctx, jobQueue.State)

	setTagsOut(ctx, jobQueue.Tags)

	response.Diagnostics.Append(response.State.Set(ctx, &data)...)
}

func (r *jobQueueResource) Update(ctx context.Context, request resource.UpdateRequest, response *resource.UpdateResponse) {
	var old, new jobQueueResourceModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &new)...)
	if response.Diagnostics.HasError() {
		return
	}
	response.Diagnostics.Append(request.State.Get(ctx, &old)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().BatchClient(ctx)

	var update bool
	input := &batch.UpdateJobQueueInput{
		JobQueue: fwflex.StringFromFramework(ctx, new.JobQueueName),
	}

	if !new.ComputeEnvironmentOrder.IsNull() && !new.ComputeEnvironmentOrder.Equal(old.ComputeEnvironmentOrder) {
		response.Diagnostics.Append(fwflex.Expand(ctx, new.ComputeEnvironmentOrder, &input.ComputeEnvironmentOrder)...)
		if response.Diagnostics.HasError() {
			return
		}
		update = true
	} else {
		if !new.ComputeEnvironments.Equal(old.ComputeEnvironments) {
			input.ComputeEnvironmentOrder = expandComputeEnvironments(ctx, new.ComputeEnvironments)
			update = true
		}
	}

	if !new.JobStateTimeLimitActions.Equal(old.JobStateTimeLimitActions) {
		response.Diagnostics.Append(fwflex.Expand(ctx, new.JobStateTimeLimitActions, &input.JobStateTimeLimitActions)...)
		if response.Diagnostics.HasError() {
			return
		}
		update = true
	}
	if !new.Priority.Equal(old.Priority) {
		input.Priority = fwflex.Int32FromFrameworkInt64(ctx, new.Priority)
		update = true
	}
	if !new.State.Equal(old.State) {
		input.State = awstypes.JQState(new.State.ValueString())
		update = true
	}
	if !old.SchedulingPolicyARN.IsNull() {
		input.SchedulingPolicyArn = fwflex.StringFromFramework(ctx, old.SchedulingPolicyARN)
		update = true
	}
	if !new.SchedulingPolicyARN.Equal(old.SchedulingPolicyARN) {
		if !new.SchedulingPolicyARN.IsNull() || !old.SchedulingPolicyARN.IsUnknown() {
			input.SchedulingPolicyArn = fwflex.StringFromFramework(ctx, new.SchedulingPolicyARN)
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
		_, err := conn.UpdateJobQueue(ctx, input)

		if err != nil {
			response.Diagnostics.AddError(fmt.Sprintf("updating Batch Job Queue (%s)", new.ID.ValueString()), err.Error())

			return
		}

		if _, err := waitJobQueueUpdated(ctx, conn, new.ID.ValueString(), r.UpdateTimeout(ctx, new.Timeouts)); err != nil {
			response.Diagnostics.AddError(fmt.Sprintf("waiting for Batch Job Queue (%s) update", new.ID.ValueString()), err.Error())

			return
		}
	}

	response.Diagnostics.Append(response.State.Set(ctx, &new)...)
}

func (r *jobQueueResource) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	var data jobQueueResourceModel
	response.Diagnostics.Append(request.State.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().BatchClient(ctx)

	updateInput := batch.UpdateJobQueueInput{
		JobQueue: fwflex.StringFromFramework(ctx, data.ID),
		State:    awstypes.JQStateDisabled,
	}
	_, err := conn.UpdateJobQueue(ctx, &updateInput)

	// "An error occurred (ClientException) when calling the UpdateJobQueue operation: ... does not exist".
	if errs.IsAErrorMessageContains[*awstypes.ClientException](err, "does not exist") {
		return
	}

	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("disabling Batch Job Queue (%s)", data.ID.ValueString()), err.Error())

		return
	}

	timeout := r.DeleteTimeout(ctx, data.Timeouts)
	if _, err := waitJobQueueUpdated(ctx, conn, data.ID.ValueString(), timeout); err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("waiting for Batch Job Queue (%s) disable", data.ID.ValueString()), err.Error())

		return
	}

	deleteInput := batch.DeleteJobQueueInput{
		JobQueue: fwflex.StringFromFramework(ctx, data.ID),
	}
	_, err = conn.DeleteJobQueue(ctx, &deleteInput)

	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("deleting Batch Job Queue (%s)", data.ID.ValueString()), err.Error())

		return
	}

	if _, err := waitJobQueueDeleted(ctx, conn, data.ID.ValueString(), timeout); err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("waiting for Batch Job Queue (%s) delete", data.ID.ValueString()), err.Error())

		return
	}
}

func (r *jobQueueResource) ConfigValidators(_ context.Context) []resource.ConfigValidator {
	return []resource.ConfigValidator{
		resourcevalidator.ExactlyOneOf(
			path.MatchRoot("compute_environments"),
			path.MatchRoot("compute_environment_order"),
		),
	}
}

func (r *jobQueueResource) UpgradeState(ctx context.Context) map[int64]resource.StateUpgrader {
	schemaV0 := jobQueueSchema0(ctx)

	return map[int64]resource.StateUpgrader{
		0: {
			PriorSchema:   &schemaV0,
			StateUpgrader: upgradeJobQueueResourceStateV0toV1,
		},
	}
}

func findJobQueueByID(ctx context.Context, conn *batch.Client, id string) (*awstypes.JobQueueDetail, error) {
	input := &batch.DescribeJobQueuesInput{
		JobQueues: []string{id},
	}

	output, err := findJobQueue(ctx, conn, input)

	if err != nil {
		return nil, err
	}

	if status := output.Status; status == awstypes.JQStatusDeleted {
		return nil, &retry.NotFoundError{
			Message:     string(status),
			LastRequest: input,
		}
	}

	return output, nil
}

func findJobQueue(ctx context.Context, conn *batch.Client, input *batch.DescribeJobQueuesInput) (*awstypes.JobQueueDetail, error) {
	output, err := findJobQueues(ctx, conn, input)

	if err != nil {
		return nil, err
	}

	return tfresource.AssertSingleValueResult(output)
}

func findJobQueues(ctx context.Context, conn *batch.Client, input *batch.DescribeJobQueuesInput) ([]awstypes.JobQueueDetail, error) {
	var output []awstypes.JobQueueDetail

	pages := batch.NewDescribeJobQueuesPaginator(conn, input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if err != nil {
			return nil, err
		}

		output = append(output, page.JobQueues...)
	}

	return output, nil
}

func statusJobQueue(ctx context.Context, conn *batch.Client, id string) retry.StateRefreshFunc {
	return func() (any, string, error) {
		output, err := findJobQueueByID(ctx, conn, id)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, string(output.Status), nil
	}
}

func waitJobQueueCreated(ctx context.Context, conn *batch.Client, id string, timeout time.Duration) (*awstypes.JobQueueDetail, error) {
	stateConf := &retry.StateChangeConf{
		Pending:    enum.Slice(awstypes.JQStatusCreating, awstypes.JQStatusUpdating),
		Target:     enum.Slice(awstypes.JQStatusValid),
		Refresh:    statusJobQueue(ctx, conn, id),
		Timeout:    timeout,
		MinTimeout: 10 * time.Second,
		Delay:      30 * time.Second,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*awstypes.JobQueueDetail); ok {
		tfresource.SetLastError(err, errors.New(aws.ToString(output.StatusReason)))

		return output, err
	}

	return nil, err
}

func waitJobQueueUpdated(ctx context.Context, conn *batch.Client, id string, timeout time.Duration) (*awstypes.JobQueueDetail, error) { //nolint:unparam
	stateConf := &retry.StateChangeConf{
		Pending:    enum.Slice(awstypes.JQStatusUpdating),
		Target:     enum.Slice(awstypes.JQStatusValid),
		Refresh:    statusJobQueue(ctx, conn, id),
		Timeout:    timeout,
		MinTimeout: 10 * time.Second,
		Delay:      30 * time.Second,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*awstypes.JobQueueDetail); ok {
		tfresource.SetLastError(err, errors.New(aws.ToString(output.StatusReason)))

		return output, err
	}

	return nil, err
}

func waitJobQueueDeleted(ctx context.Context, conn *batch.Client, id string, timeout time.Duration) (*awstypes.JobQueueDetail, error) {
	stateConf := &retry.StateChangeConf{
		Pending:    enum.Slice(awstypes.JQStatusDeleting),
		Target:     []string{},
		Refresh:    statusJobQueue(ctx, conn, id),
		Timeout:    timeout,
		MinTimeout: 10 * time.Second,
		Delay:      30 * time.Second,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*awstypes.JobQueueDetail); ok {
		tfresource.SetLastError(err, errors.New(aws.ToString(output.StatusReason)))

		return output, err
	}

	return nil, err
}

type jobQueueResourceModel struct {
	ComputeEnvironments      types.List                                                    `tfsdk:"compute_environments"`
	ComputeEnvironmentOrder  fwtypes.ListNestedObjectValueOf[computeEnvironmentOrderModel] `tfsdk:"compute_environment_order"`
	ID                       types.String                                                  `tfsdk:"id"`
	JobQueueARN              types.String                                                  `tfsdk:"arn"`
	JobQueueName             types.String                                                  `tfsdk:"name"`
	JobStateTimeLimitActions fwtypes.ListNestedObjectValueOf[jobStateTimeLimitActionModel] `tfsdk:"job_state_time_limit_action"`
	Priority                 types.Int64                                                   `tfsdk:"priority"`
	SchedulingPolicyARN      fwtypes.ARN                                                   `tfsdk:"scheduling_policy_arn"`
	State                    types.String                                                  `tfsdk:"state"`
	Tags                     tftags.Map                                                    `tfsdk:"tags"`
	TagsAll                  tftags.Map                                                    `tfsdk:"tags_all"`
	Timeouts                 timeouts.Value                                                `tfsdk:"timeouts"`
}

func (model *jobQueueResourceModel) InitFromID() error {
	model.JobQueueARN = model.ID

	return nil
}

func (model *jobQueueResourceModel) setID() {
	model.ID = model.JobQueueARN
}

type computeEnvironmentOrderModel struct {
	ComputeEnvironment fwtypes.ARN `tfsdk:"compute_environment"`
	Order              types.Int64 `tfsdk:"order"`
}

type jobStateTimeLimitActionModel struct {
	Action         fwtypes.StringEnum[awstypes.JobStateTimeLimitActionsAction] `tfsdk:"action"`
	MaxTimeSeconds types.Int64                                                 `tfsdk:"max_time_seconds"`
	Reason         types.String                                                `tfsdk:"reason"`
	State          fwtypes.StringEnum[awstypes.JobStateTimeLimitActionsState]  `tfsdk:"state"`
}

func expandComputeEnvironments(ctx context.Context, tfList types.List) []awstypes.ComputeEnvironmentOrder {
	var apiObjects []awstypes.ComputeEnvironmentOrder

	for i, env := range fwflex.ExpandFrameworkStringList(ctx, tfList) {
		apiObjects = append(apiObjects, awstypes.ComputeEnvironmentOrder{
			ComputeEnvironment: env,
			Order:              aws.Int32(int32(i)),
		})
	}

	return apiObjects
}

func flattenComputeEnvironments(ctx context.Context, apiObjects []awstypes.ComputeEnvironmentOrder) types.List {
	slices.SortFunc(apiObjects, func(a, b awstypes.ComputeEnvironmentOrder) int {
		return cmp.Compare(aws.ToInt32(a.Order), aws.ToInt32(b.Order))
	})

	return fwflex.FlattenFrameworkStringListLegacy(ctx, tfslices.ApplyToAll(apiObjects, func(v awstypes.ComputeEnvironmentOrder) *string {
		return v.ComputeEnvironment
	}))
}
