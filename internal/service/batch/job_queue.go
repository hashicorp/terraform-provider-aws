// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package batch

import (
	"context"
	"errors"
	"fmt"
	"sort"
	"time"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/batch"
	awstypes "github.com/aws/aws-sdk-go-v2/service/batch/types"
	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
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

func (*jobQueueResource) Metadata(_ context.Context, request resource.MetadataRequest, response *resource.MetadataResponse) {
	response.TypeName = "aws_batch_job_queue"
}

func (r *jobQueueResource) Schema(ctx context.Context, request resource.SchemaRequest, response *resource.SchemaResponse) {
	response.Schema = schema.Schema{
		Version: 1,
		Attributes: map[string]schema.Attribute{
			names.AttrARN: framework.ARNAttributeComputedOnly(),
			"compute_environments": schema.ListAttribute{
				ElementType:        fwtypes.ARNType,
				Optional:           true,
				Computed:           true,
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
	input := &batch.CreateJobQueueInput{}
	response.Diagnostics.Append(fwflex.Expand(ctx, data, input)...)
	if response.Diagnostics.HasError() {
		return
	}

	if data.ComputeEnvironmentOrder.IsNull() {
		input.ComputeEnvironmentOrder = expandComputeEnvironments(ctx, data.ComputeEnvironments)
	}

	// Additional fields.
	input.Tags = getTagsIn(ctx)

	output, err := conn.CreateJobQueue(ctx, input)

	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("creating Batch Job Queue (%s)", name), err.Error())

		return
	}

	// Set values for unknowns.
	data.JobQueueARN = fwflex.StringToFramework(ctx, output.JobQueueArn)
	data.setID()

	jobQueue, err := waitJobQueueCreated(ctx, conn, data.ID.ValueString(), r.CreateTimeout(ctx, data.Timeouts))

	if err != nil {
		response.State.SetAttribute(ctx, path.Root(names.AttrID), data.ID) // Set 'id' so as to taint the resource.
		response.Diagnostics.AddError(fmt.Sprintf("waiting for Batch Job Queue (%s) create", data.ID.ValueString()), err.Error())

		return
	}

	data.ComputeEnvironments = flattenComputeEnvironments(ctx, jobQueue.ComputeEnvironmentOrder)

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

	jobQueue, err := findJobQueueByARN(ctx, conn, data.ID.ValueString())

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
	response.Diagnostics.Append(fwflex.Flatten(ctx, jobQueue, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	data.ComputeEnvironments = flattenComputeEnvironments(ctx, jobQueue.ComputeEnvironmentOrder)

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

	if !new.ComputeEnvironmentOrder.Equal(old.ComputeEnvironmentOrder) ||
		!new.ComputeEnvironments.Equal(old.ComputeEnvironments) ||
		!new.Priority.Equal(old.Priority) ||
		!new.SchedulingPolicyARN.Equal(old.SchedulingPolicyARN) ||
		!new.State.Equal(old.State) {
		if new.SchedulingPolicyARN.IsNull() {
			response.Diagnostics.AddError(
				"cannot remove the fair share scheduling policy",
				"cannot remove scheduling policy",
			)
			return
		}

		input := &batch.UpdateJobQueueInput{}
		response.Diagnostics.Append(fwflex.Expand(ctx, new, input)...)
		if response.Diagnostics.HasError() {
			return
		}

		_, err := conn.UpdateJobQueue(ctx, input)

		if err != nil {
			response.Diagnostics.AddError(fmt.Sprintf("updating Batch Job Queue (%s)", new.ID.ValueString()), err.Error())

			return
		}

		jobQueue, err := waitJobQueueUpdated(ctx, conn, new.ID.ValueString(), r.UpdateTimeout(ctx, new.Timeouts))

		if err != nil {
			response.Diagnostics.AddError(fmt.Sprintf("waiting for Batch Job Queue (%s) update", new.ID.ValueString()), err.Error())

			return
		}

		new.ComputeEnvironments = flattenComputeEnvironments(ctx, jobQueue.ComputeEnvironmentOrder)
	} else {
		new.ComputeEnvironments = old.ComputeEnvironments
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

	_, err := conn.UpdateJobQueue(ctx, &batch.UpdateJobQueueInput{
		JobQueue: fwflex.StringFromFramework(ctx, data.ID),
		State:    awstypes.JQStateDisabled,
	})

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

	_, err = conn.DeleteJobQueue(ctx, &batch.DeleteJobQueueInput{
		JobQueue: fwflex.StringFromFramework(ctx, data.ID),
	})

	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("deleting Batch Job Queue (%s)", data.ID.ValueString()), err.Error())

		return
	}

	if _, err := waitJobQueueDeleted(ctx, conn, data.ID.ValueString(), timeout); err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("waiting for Batch Job Queue (%s) delete", data.ID.ValueString()), err.Error())

		return
	}
}

func (r *jobQueueResource) ModifyPlan(ctx context.Context, request resource.ModifyPlanRequest, response *resource.ModifyPlanResponse) {
	r.SetTagsAll(ctx, request, response)
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

func findJobQueueByARN(ctx context.Context, conn *batch.Client, arn string) (*awstypes.JobQueueDetail, error) {
	input := &batch.DescribeJobQueuesInput{
		JobQueues: []string{arn},
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

func statusJobQueue(ctx context.Context, conn *batch.Client, arn string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		output, err := findJobQueueByARN(ctx, conn, arn)

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

func waitJobQueueUpdated(ctx context.Context, conn *batch.Client, arn string, timeout time.Duration) (*awstypes.JobQueueDetail, error) {
	stateConf := &retry.StateChangeConf{
		Pending:    enum.Slice(awstypes.JQStatusUpdating),
		Target:     enum.Slice(awstypes.JQStatusValid),
		Refresh:    statusJobQueue(ctx, conn, arn),
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
	ComputeEnvironments     types.List                                                    `tfsdk:"compute_environments"`
	ComputeEnvironmentOrder fwtypes.ListNestedObjectValueOf[computeEnvironmentOrderModel] `tfsdk:"compute_environment_order"`
	ID                      types.String                                                  `tfsdk:"id"`
	JobQueueARN             types.String                                                  `tfsdk:"arn"`
	JobQueueName            types.String                                                  `tfsdk:"name"`
	Priority                types.Int64                                                   `tfsdk:"priority"`
	SchedulingPolicyARN     fwtypes.ARN                                                   `tfsdk:"scheduling_policy_arn"`
	State                   types.String                                                  `tfsdk:"state"`
	Tags                    types.Map                                                     `tfsdk:"tags"`
	TagsAll                 types.Map                                                     `tfsdk:"tags_all"`
	Timeouts                timeouts.Value                                                `tfsdk:"timeouts"`
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
	sort.Slice(apiObjects, func(i, j int) bool {
		return aws.ToInt32(apiObjects[i].Order) < aws.ToInt32(apiObjects[j].Order)
	})

	return fwflex.FlattenFrameworkStringList(ctx, tfslices.ApplyToAll(apiObjects, func(v awstypes.ComputeEnvironmentOrder) *string {
		return v.ComputeEnvironment
	}))
}
