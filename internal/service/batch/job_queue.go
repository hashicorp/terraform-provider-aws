// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package batch

import (
	"context"
	"errors"
	"fmt"
	"iter"
	"slices"
	"time"
	"unique"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/batch"
	awstypes "github.com/aws/aws-sdk-go-v2/service/batch/types"
	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/list"
	listschema "github.com/hashicorp/terraform-plugin-framework/list/schema"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/fwdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	fwflex "github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/provider/interceptors"
	tfslices "github.com/hashicorp/terraform-provider-aws/internal/slices"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	inttypes "github.com/hashicorp/terraform-provider-aws/internal/types"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkResource("aws_batch_job_queue", name="Job Queue")
// @Tags(identifierAttribute="arn")
// @ArnIdentity(identityDuplicateAttributes="id")
// @ArnFormat("job-queue/{name}")
// @Testing(existsType="github.com/aws/aws-sdk-go-v2/service/batch/types;types.JobQueueDetail")
// @Testing(preIdentityVersion="v5.100.0")
func newJobQueueResource(_ context.Context) (resource.ResourceWithConfigure, error) {
	r := jobQueueResource{}

	r.SetDefaultCreateTimeout(10 * time.Minute)
	r.SetDefaultUpdateTimeout(10 * time.Minute)
	r.SetDefaultDeleteTimeout(10 * time.Minute)

	return &r, nil
}

// @List
func JobQueueResourceAsListResource() list.ListResourceWithConfigure {
	return &jobQueueResource{}
}

var _ list.ListResource = &jobQueueResource{}

type jobQueueResource struct {
	framework.ResourceWithModel[jobQueueResourceModel]
	framework.WithTimeouts
	framework.WithImportByIdentity
}

func (r *jobQueueResource) Schema(ctx context.Context, request resource.SchemaRequest, response *resource.SchemaResponse) {
	response.Schema = schema.Schema{
		Version: 2,
		Attributes: map[string]schema.Attribute{
			names.AttrARN: framework.ARNAttributeComputedOnly(),
			names.AttrID:  framework.IDAttributeDeprecatedWithAlternate(path.Root(names.AttrARN)),
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

	response.Diagnostics.Append(fwflex.Flatten(ctx, jobQueue, &data, fwflex.WithFieldNamePrefix("JobQueue"))...)
	if response.Diagnostics.HasError() {
		return
	}

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

func (r *jobQueueResource) UpgradeState(ctx context.Context) map[int64]resource.StateUpgrader {
	schemaV0 := jobQueueSchema0(ctx)
	schemaV1 := jobQueueSchema1(ctx)

	return map[int64]resource.StateUpgrader{
		0: {
			PriorSchema:   &schemaV0,
			StateUpgrader: upgradeJobQueueResourceStateV0toV1,
		},
		1: {
			PriorSchema:   &schemaV1,
			StateUpgrader: upgradeJobQueueResourceStateV1toV2,
		},
	}
}

func findJobQueueByID(ctx context.Context, conn *batch.Client, id string) (*awstypes.JobQueueDetail, error) {
	input := batch.DescribeJobQueuesInput{
		JobQueues: []string{id},
	}

	output, err := findJobQueue(ctx, conn, &input)

	if err != nil {
		return nil, err
	}

	if status := output.Status; status == awstypes.JQStatusDeleted {
		return nil, &retry.NotFoundError{
			Message: string(status),
		}
	}

	return output, nil
}

func findJobQueue(ctx context.Context, conn *batch.Client, input *batch.DescribeJobQueuesInput) (*awstypes.JobQueueDetail, error) {
	return tfresource.AssertSingleValueResultIterErr(listJobQueues(ctx, conn, input))
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
	framework.WithRegionModel
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

// DescribeJobQueues is an "All-Or-Some" call.
func listJobQueues(ctx context.Context, conn *batch.Client, input *batch.DescribeJobQueuesInput) iter.Seq2[awstypes.JobQueueDetail, error] {
	return func(yield func(awstypes.JobQueueDetail, error) bool) {
		pages := batch.NewDescribeJobQueuesPaginator(conn, input)
		for pages.HasMorePages() {
			page, err := pages.NextPage(ctx)
			if err != nil {
				yield(awstypes.JobQueueDetail{}, fmt.Errorf("listing Batch Job Queues: %w", err))
				return
			}

			for _, jobQueue := range page.JobQueues {
				if !yield(jobQueue, nil) {
					return
				}
			}
		}
	}
}

func (r jobQueueResource) ListResourceConfigSchema(_ context.Context, request list.ListResourceSchemaRequest, response *list.ListResourceSchemaResponse) {
	response.Schema = listschema.Schema{
		Attributes: map[string]listschema.Attribute{},
	}
}

func (r jobQueueResource) List(ctx context.Context, request list.ListRequest, stream *list.ListResultsStream) {
	var query jobQueueListModel

	if request.Config.Raw.IsKnown() && !request.Config.Raw.IsNull() {
		if diags := request.Config.Get(ctx, &query); diags.HasError() {
			stream.Results = list.ListResultsStreamDiagnostics(diags)
			return
		}
	}

	awsClient := r.Meta()
	conn := awsClient.BatchClient(ctx)

	interceptors := []listResultInterceptor{
		populateTimeoutsInterceptor{},
		populateIdentityInterceptor{},
		identityInterceptor{
			attributes: r.IdentitySpec().Attributes,
		},
		setResourceInterceptor{},
		tagsInterceptor{
			HTags: interceptors.HTags(unique.Make(inttypes.ServicePackageResourceTags{
				IdentifierAttribute: names.AttrARN,
			})),
		},
	}

	stream.Results = func(yield func(list.ListResult) bool) {
		result := request.NewListResult()
		var input batch.DescribeJobQueuesInput
		for jobQueue, err := range listJobQueues(ctx, conn, &input) {
			if err != nil {
				result = list.ListResult{
					Diagnostics: diag.Diagnostics{
						diag.NewErrorDiagnostic(
							"Error Listing Remote Resources",
							fmt.Sprintf("Error: %s", err),
						),
					},
				}
				yield(result)
				return
			}

			ctx = tftags.NewContext(ctx, awsClient.DefaultTagsConfig(ctx), awsClient.IgnoreTagsConfig(ctx))

			var data jobQueueResourceModel

			params := interceptorParams{
				c:      awsClient,
				result: &result,
				object: &data,
			}

			params.when = Before
			for interceptor := range slices.Values(interceptors) {
				d := interceptor.read(ctx, params)
				result.Diagnostics.Append(d...)
				if d.HasError() {
					result = list.ListResult{Diagnostics: result.Diagnostics}
					yield(result)
					return
				}
			}

			if diags := fwflex.Flatten(ctx, jobQueue, &data, fwflex.WithFieldNamePrefix("JobQueue")); diags.HasError() {
				result.Diagnostics.Append(diags...)
			}
			setTagsOut(ctx, jobQueue.Tags)

			result.DisplayName = fmt.Sprintf("x%s (%s)x", data.JobQueueName.ValueString(), data.JobQueueARN.ValueString())

			params.when = After
			for interceptor := range tfslices.BackwardValues(interceptors) {
				d := interceptor.read(ctx, params)
				result.Diagnostics.Append(d...)
				if d.HasError() {
					result = list.ListResult{Diagnostics: result.Diagnostics}
					yield(result)
					return
				}
			}

			if result.Diagnostics.HasError() {
				result = list.ListResult{Diagnostics: result.Diagnostics}
				yield(result)
				return
			}

			if !yield(result) {
				return
			}
		}
	}
}

type jobQueueListModel struct {
	// TODO: factor out
	Region types.String `tfsdk:"region"`
}

// when represents the point in the CRUD request lifecycle that an interceptor is run.
// Multiple values can be ORed together.
type when uint16

const (
	Before  when = 1 << iota // Interceptor is invoked before call to method in schema
	After                    // Interceptor is invoked after successful call to method in schema
	OnError                  // Interceptor is invoked after unsuccessful call to method in schema
	Finally                  // Interceptor is invoked after After or OnError
)

type interceptorParams struct {
	c      *conns.AWSClient
	result *list.ListResult
	object *jobQueueResourceModel // Because tfsdk.Resource doesn't have SetAttribute
	when   when
}

type listResultInterceptor interface {
	read(ctx context.Context, params interceptorParams) diag.Diagnostics
}

type tagsInterceptor struct {
	interceptors.HTags
}

// Copied from tagsResourceInterceptor.read()
func (r tagsInterceptor) read(ctx context.Context, params interceptorParams) (diags diag.Diagnostics) {
	sp, serviceName, resourceName, tagsInContext, ok := interceptors.InfoFromContext(ctx, params.c)
	if !ok {
		return
	}

	switch params.when {
	case After:
		// If the R handler didn't set tags, try and read them from the service API.
		if tagsInContext.TagsOut.IsNone() {
			// Some old resources may not have the required attribute set after Read:
			// https://github.com/hashicorp/terraform-provider-aws/issues/31180
			if identifier := r.GetIdentifierFramework(ctx, params.result.Resource); identifier != "" {
				if err := r.ListTags(ctx, sp, params.c, identifier); err != nil {
					diags.AddError(fmt.Sprintf("listing tags for %s %s (%s)", serviceName, resourceName, identifier), err.Error())

					return
				}
			}
		}

		apiTags := tagsInContext.TagsOut.UnwrapOrDefault()

		// AWS APIs often return empty lists of tags when none have been configured.
		var stateTags tftags.Map
		params.result.Resource.GetAttribute(ctx, path.Root(names.AttrTags), &stateTags)
		// Remove any provider configured ignore_tags and system tags from those returned from the service API.
		// The resource's configured tags do not include any provider configured default_tags.
		if v := apiTags.IgnoreSystem(sp.ServicePackageName()).IgnoreConfig(params.c.IgnoreTagsConfig(ctx)).ResolveDuplicatesFramework(ctx, params.c.DefaultTagsConfig(ctx), params.c.IgnoreTagsConfig(ctx), stateTags, &diags).Map(); len(v) > 0 {
			stateTags = tftags.NewMapFromMapValue(fwflex.FlattenFrameworkStringValueMapLegacy(ctx, v))
		}
		params.object.Tags = stateTags

		// Computed tags_all do.
		stateTagsAll := fwflex.FlattenFrameworkStringValueMapLegacy(ctx, apiTags.IgnoreSystem(sp.ServicePackageName()).IgnoreConfig(params.c.IgnoreTagsConfig(ctx)).Map())
		params.object.TagsAll = tftags.NewMapFromMapValue(stateTagsAll)
	}

	return
}

// This interceptor will not be needed if Framework pre-populates the Resource as it does with CRUD operations
type populateTimeoutsInterceptor struct{}

func (r populateTimeoutsInterceptor) read(ctx context.Context, params interceptorParams) (diags diag.Diagnostics) {
	switch params.when {
	case Before:
		timeoutsType, d := params.result.Resource.Schema.TypeAtPath(ctx, path.Root("timeouts"))
		diags.Append(d...)
		if d.HasError() {
			return
		}

		obj, d := newEmptyObject(timeoutsType)
		diags.Append(d...)
		if d.HasError() {
			return
		}
		params.object.Timeouts.Object = obj
	}

	return
}

// This interceptor will not be needed if Framework pre-populates the Identity as it does with CRUD operations
type populateIdentityInterceptor struct{}

func (r populateIdentityInterceptor) read(ctx context.Context, params interceptorParams) (diags diag.Diagnostics) {
	switch params.when {
	case Before:
		identityType := params.result.Identity.Schema.Type()

		obj, d := newEmptyObject(identityType)
		diags.Append(d...)
		if diags.HasError() {
			return
		}

		diags.Append(params.result.Identity.Set(ctx, obj)...)
		if diags.HasError() {
			return
		}
	}

	return
}

// This interceptor will not be needed if the Resource value is of a type that implements SetAttribute
type setResourceInterceptor struct{}

func (r setResourceInterceptor) read(ctx context.Context, params interceptorParams) (diags diag.Diagnostics) {
	switch params.when {
	case After:
		diags.Append(params.result.Resource.Set(ctx, params.object)...)
	}

	return
}

type identityInterceptor struct {
	attributes []inttypes.IdentityAttribute
}

func (r identityInterceptor) read(ctx context.Context, params interceptorParams) (diags diag.Diagnostics) {
	awsClient := params.c

	switch params.when {
	case After:
		for _, att := range r.attributes {
			switch att.Name() {
			case names.AttrAccountID:
				diags.Append(params.result.Identity.SetAttribute(ctx, path.Root(att.Name()), awsClient.AccountID(ctx))...)
				if diags.HasError() {
					return
				}

			case names.AttrRegion:
				diags.Append(params.result.Identity.SetAttribute(ctx, path.Root(att.Name()), awsClient.Region(ctx))...)
				if diags.HasError() {
					return
				}

			default:
				var attrVal attr.Value
				diags.Append(params.result.Resource.GetAttribute(ctx, path.Root(att.ResourceAttributeName()), &attrVal)...)
				if diags.HasError() {
					return
				}

				diags.Append(params.result.Identity.SetAttribute(ctx, path.Root(att.Name()), attrVal)...)
				if diags.HasError() {
					return
				}
			}
		}
	}

	return
}

func newEmptyObject(typ attr.Type) (obj basetypes.ObjectValue, diags diag.Diagnostics) {
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
	attrValues := make(map[string]attr.Value, len(attrTypes))
	for attrName := range attrTypes {
		attrValues[attrName] = types.StringNull()
	}
	obj, d := basetypes.NewObjectValue(attrTypes, attrValues)
	diags.Append(d...)
	if d.HasError() {
		return basetypes.ObjectValue{}, diags
	}

	return obj, diags
}
