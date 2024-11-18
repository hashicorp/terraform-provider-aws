// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package batch

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"regexp"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/batch"
	awstypes "github.com/aws/aws-sdk-go-v2/service/batch/types"
	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	"github.com/hashicorp/terraform-plugin-framework-validators/setvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int32planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	"github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @Tags(identifierAttribute="arn")
// @FrameworkResource("aws_batch_job_definition", name="Job Definition")
func newResourceJobDefinition(_ context.Context) (resource.ResourceWithConfigure, error) {
	r := &resourceJobDefinition{}

	return r, nil
}

const (
	ResNameJobDefinition = "Job Definition"
)

type resourceJobDefinition struct {
	framework.ResourceWithConfigure
}

func (r *resourceJobDefinition) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = "aws_batch_job_definition"
}

func (r *resourceJobDefinition) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			names.AttrARN: framework.ARNAttributeComputedOnly(),
			names.AttrID:  framework.IDAttribute(),
			"arn_prefix": schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"container_properties": schema.StringAttribute{
				Optional:   true,
				CustomType: jsontypes.NormalizedType{},
				Validators: []validator.String{
					containerPropertiesValidator{},
					stringvalidator.ExactlyOneOf(
						path.MatchRoot("container_properties"),
						path.MatchRoot("ecs_properties"),
						path.MatchRoot("eks_properties"),
						path.MatchRoot("node_properties"),
					),
				},
				PlanModifiers: []planmodifier.String{
					ContainerPropertiesStringPlanModifier(),
					stringplanmodifier.RequiresReplace(),
				},
			},
			"deregister_on_new_revision": schema.BoolAttribute{
				Default:  booldefault.StaticBool(false),
				Optional: true,
				Computed: true,
			},
			"ecs_properties": schema.StringAttribute{
				Optional: true,
				Validators: []validator.String{
					ecsPropertiesValidator{},
				},
				PlanModifiers: []planmodifier.String{
					ECSStringPlanModifier(),
				},
			},
			"eks_properties": schema.ListAttribute{
				CustomType: fwtypes.NewListNestedObjectTypeOf[eksResourcePropertiesModel](ctx),
				Optional:   true,
				ElementType: types.ObjectType{
					AttrTypes: map[string]attr.Type{
						"pod_properties": fwtypes.NewListNestedObjectTypeOf[eksResourcePropertiesModel](ctx),
					},
				},
			},

			names.AttrName: schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
				Validators: []validator.String{
					stringvalidator.RegexMatches(
						regexp.MustCompile(`^[0-9A-Za-z]{1}[0-9A-Za-z_-]{0,127}$`),
						`must be up to 128 letters (uppercase and lowercase), numbers, underscores and dashes, and must start with an alphanumeric`,
					),
				},
			},
			"node_properties": schema.StringAttribute{
				Optional: true,
				Validators: []validator.String{
					nodePropertiesValidator{},
				},
				PlanModifiers: []planmodifier.String{
					NodePropertiesStringPlanModifier(),
				},
			},
			names.AttrParameters: schema.MapAttribute{
				Optional:    true,
				ElementType: types.StringType,
			},
			"retry_strategy": schema.ListAttribute{
				CustomType: fwtypes.NewListNestedObjectTypeOf[retryStrategyModel](ctx),
				Computed:   true,
				Optional:   true,
				ElementType: types.ObjectType{
					AttrTypes: map[string]attr.Type{
						"attempts":         types.Int64Type,
						"evaluate_on_exit": fwtypes.NewListNestedObjectTypeOf[evaluateOnExitModel](ctx),
					},
				},
			},
			"platform_capabilities": schema.SetAttribute{
				Optional:    true,
				ElementType: types.StringType,
				Validators: []validator.Set{
					setvalidator.ValueStringsAre(
						enum.FrameworkValidate[awstypes.PlatformCapability](),
					),
				},
			},
			names.AttrPropagateTags: schema.BoolAttribute{
				Optional: true,
				Computed: true,
				Default:  booldefault.StaticBool(false),
			},

			"revision": schema.Int32Attribute{
				Computed: true,
				PlanModifiers: []planmodifier.Int32{
					int32planmodifier.UseStateForUnknown(),
				},
			},
			"scheduling_priority": schema.Int32Attribute{
				Optional: true,
			},
			names.AttrTags:    tftags.TagsAttribute(),
			names.AttrTagsAll: tftags.TagsAttributeComputedOnly(),

			names.AttrType: schema.StringAttribute{
				Required: true,
				Validators: []validator.String{
					enum.FrameworkValidate[awstypes.JobDefinitionType](),
				},
			},
			names.AttrTimeout: schema.ListAttribute{
				CustomType: fwtypes.NewListNestedObjectTypeOf[jobTimeoutModel](ctx),
				Optional:   true,
				ElementType: types.ObjectType{
					AttrTypes: map[string]attr.Type{
						"attempt_duration_seconds": types.Int64Type,
					},
				},
			},
		},
	}
}

func (r *resourceJobDefinition) readJobDefinitionIntoState(ctx context.Context, jd *awstypes.JobDefinition, state *resourceJobDefinitionModel) (resp diag.Diagnostics) {
	resp.Append(flex.Flatten(ctx, jd, state,
		flex.WithIgnoredFieldNamesAppend("TagsAll"),
		// Name and Arn are prefixed by JobDefinition
		flex.WithFieldNamePrefix("JobDefinition"),
	)...)
	if resp.HasError() {
		return resp
	}

	state.ID = types.StringPointerValue(jd.JobDefinitionArn)

	if jd.Revision != nil {
		arn, revision := aws.ToString(jd.JobDefinitionArn), aws.ToInt32(jd.Revision)
		state.ArnPrefix = types.StringValue(strings.TrimSuffix(arn, fmt.Sprintf(":%d", revision)))
	}

	// Convert the complex arguments to string
	// Future iterations will fully define the type
	if jd.ContainerProperties != nil {
		containerProps, err := flattenContainerProperties(jd.ContainerProperties)
		if err != nil {
			resp.AddError(
				create.ProblemStandardMessage(names.Batch, create.ErrActionSetting, ResNameJobDefinition, state.ID.String(), err),
				err.Error(),
			)
			return resp
		}
		state.ContainerProperties = jsontypes.NewNormalizedValue(string(containerProps))
	}

	if jd.EcsProperties != nil {
		ecsProps, err := flattenECSProperties(jd.EcsProperties)
		if err != nil {
			resp.AddError(
				create.ProblemStandardMessage(names.Batch, create.ErrActionSetting, ResNameJobDefinition, state.ID.String(), err),
				err.Error(),
			)
			return resp
		}
		state.ECSProperties = types.StringValue(ecsProps)
	}

	if jd.NodeProperties != nil {
		nodeProps, err := flattenNodeProperties(jd.NodeProperties)
		if err != nil {
			resp.AddError(
				create.ProblemStandardMessage(names.Batch, create.ErrActionSetting, ResNameJobDefinition, state.ID.String(), err),
				err.Error(),
			)
			return resp
		}
		state.NodeProperties = types.StringValue(string(nodeProps))
	}

	return resp
}

func (r *resourceJobDefinition) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	conn := r.Meta().BatchClient(ctx)

	var plan resourceJobDefinitionModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	input := &batch.RegisterJobDefinitionInput{
		JobDefinitionName: plan.Name.ValueStringPointer(),
		Type:              awstypes.JobDefinitionType(plan.Type.ValueString()),
		Tags:              getTagsIn(ctx),
	}

	// flex.WithIgnoredFieldNamesAppend("ArnPrefix"),
	// flex.WithIgnoredFieldNamesAppend("TagsAll"),
	// flex.WithIgnoredFieldNamesAppend("Type"),
	// // Name and Arn are prefixed by JobDefinition
	// flex.WithFieldNamePrefix("JobDefinition"),
	resp.Diagnostics.Append(flex.Expand(ctx, plan, input,
		flex.WithIgnoredFieldNamesAppend("Arn"),
		flex.WithIgnoredFieldNamesAppend("Name"), // JobDefinitionName is a separate field
		flex.WithIgnoredFieldNamesAppend("Revision"),
	)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Marshall from string to batch type
	// Future iterations will fully define the type
	if !plan.ContainerProperties.IsNull() {
		var containerProps awstypes.ContainerProperties
		err := json.Unmarshal([]byte(plan.ContainerProperties.ValueString()), &containerProps)
		if err != nil {
			resp.Diagnostics.AddError(
				create.ProblemStandardMessage(names.Batch, create.ErrActionCreating, ResNameJobDefinition, plan.Name.String(), err),
				err.Error(),
			)
			return
		}
		input.ContainerProperties = &containerProps
	}

	if !plan.ECSProperties.IsNull() {
		var ecsProps awstypes.EcsProperties
		err := json.Unmarshal([]byte(plan.ECSProperties.ValueString()), &ecsProps)
		if err != nil {
			resp.Diagnostics.AddError(
				create.ProblemStandardMessage(names.Batch, create.ErrActionCreating, ResNameJobDefinition, plan.Name.String(), err),
				err.Error(),
			)
			return
		}
		input.EcsProperties = &ecsProps
	}

	if !plan.NodeProperties.IsNull() {
		var nodeProps awstypes.NodeProperties
		err := json.Unmarshal([]byte(plan.NodeProperties.ValueString()), &nodeProps)
		if err != nil {
			resp.Diagnostics.AddError(
				create.ProblemStandardMessage(names.Batch, create.ErrActionCreating, ResNameJobDefinition, plan.Name.String(), err),
				err.Error(),
			)
			return
		}
		input.NodeProperties = &nodeProps
	}

	if resp.Diagnostics.HasError() {
		return
	}

	out, err := conn.RegisterJobDefinition(ctx, input)
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.Batch, create.ErrActionCreating, ResNameJobDefinition, plan.Name.String(), err),
			err.Error(),
		)
		return
	}
	if out == nil || out.JobDefinitionArn == nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.Batch, create.ErrActionCreating, ResNameJobDefinition, plan.Name.String(), nil),
			errors.New("empty output").Error(),
		)
		return
	}

	jd, err := findJobDefinitionByARN(ctx, conn, *out.JobDefinitionArn)
	if err != nil || jd == nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.Batch, create.ErrActionSetting, ResNameJobDefinition, plan.ID.String(), err),
			err.Error(),
		)
		return
	}

	resp.Diagnostics.Append(r.readJobDefinitionIntoState(ctx, jd, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *resourceJobDefinition) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	conn := r.Meta().BatchClient(ctx)

	var state resourceJobDefinitionModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	out, err := findJobDefinitionByARN(ctx, conn, state.ID.ValueString())
	if tfresource.NotFound(err) {
		resp.State.RemoveResource(ctx)
		return
	}
	if err != nil || out == nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.Batch, create.ErrActionReading, ResNameJobDefinition, state.ID.String(), err),
			err.Error(),
		)
		return
	}
	resp.Diagnostics.Append(r.readJobDefinitionIntoState(ctx, out, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(flex.Flatten(ctx, out, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *resourceJobDefinition) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	conn := r.Meta().BatchClient(ctx)

	// TIP: -- 2. Fetch the plan
	var plan, state resourceJobDefinitionModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	input := &batch.RegisterJobDefinitionInput{
		JobDefinitionName: state.Name.ValueStringPointer(),
		Tags:              getTagsIn(ctx),
		Type:              awstypes.JobDefinitionType(plan.Type.ValueString()),
	}

	resp.Diagnostics.Append(flex.Expand(ctx, plan, input)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Marshall from string to batch type
	// Future iterations will fully define the type
	if !plan.ContainerProperties.IsNull() {
		var containerProps awstypes.ContainerProperties
		err := json.Unmarshal([]byte(plan.ContainerProperties.ValueString()), &containerProps)
		if err != nil {
			resp.Diagnostics.AddError(
				create.ProblemStandardMessage(names.Batch, create.ErrActionCreating, ResNameJobDefinition, plan.Name.String(), err),
				err.Error(),
			)
			return
		}
		input.ContainerProperties = &containerProps
	}

	if !plan.ECSProperties.IsNull() {
		var ecsProps awstypes.EcsProperties
		err := json.Unmarshal([]byte(plan.ECSProperties.ValueString()), &ecsProps)
		if err != nil {
			resp.Diagnostics.AddError(
				create.ProblemStandardMessage(names.Batch, create.ErrActionCreating, ResNameJobDefinition, plan.Name.String(), err),
				err.Error(),
			)
			return
		}
		input.EcsProperties = &ecsProps
	}

	if !plan.NodeProperties.IsNull() {
		var nodeProps awstypes.NodeProperties
		err := json.Unmarshal([]byte(plan.NodeProperties.ValueString()), &nodeProps)
		if err != nil {
			resp.Diagnostics.AddError(
				create.ProblemStandardMessage(names.Batch, create.ErrActionCreating, ResNameJobDefinition, plan.Name.String(), err),
				err.Error(),
			)
			return
		}
		input.NodeProperties = &nodeProps
	}

	if resp.Diagnostics.HasError() {
		return
	}

	out, err := conn.RegisterJobDefinition(ctx, input)
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.Batch, create.ErrActionCreating, ResNameJobDefinition, plan.Name.String(), err),
			err.Error(),
		)
		return
	}
	if out == nil || out.JobDefinitionArn == nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.Batch, create.ErrActionCreating, ResNameJobDefinition, plan.Name.String(), nil),
			errors.New("empty output").Error(),
		)
		return
	}

	state.ID = types.StringPointerValue(out.JobDefinitionArn)
	state.Revision = types.Int32PointerValue(out.Revision)

	if plan.DeregisterOnNewRevision.ValueBool() {
		tflog.Debug(ctx, fmt.Sprintf("[DEBUG] Deleting previous Batch Job Definition: %s", *out.JobDefinitionArn))
		_, err := conn.DeregisterJobDefinition(ctx, &batch.DeregisterJobDefinitionInput{
			JobDefinition: out.JobDefinitionArn,
		})

		if err != nil {
			resp.Diagnostics.AddError(
				create.ProblemStandardMessage(names.Batch, create.ErrActionDeleting, ResNameJobDefinition, aws.ToString(out.JobDefinitionArn), nil),
				err.Error(),
			)
			return
		}
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *resourceJobDefinition) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	conn := r.Meta().BatchClient(ctx)

	var state resourceJobDefinitionModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	input := &batch.DescribeJobDefinitionsInput{
		JobDefinitionName: state.Name.ValueStringPointer(),
		Status:            aws.String(jobDefinitionStatusActive),
	}

	jds, err := findJobDefinitions(ctx, conn, input)

	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.Batch, create.ErrActionReading, ResNameJobDefinition, state.ID.String(), err),
			err.Error(),
		)
	}

	for i := range jds {
		arn := aws.ToString(jds[i].JobDefinitionArn)

		_, err := conn.DeregisterJobDefinition(ctx, &batch.DeregisterJobDefinitionInput{
			JobDefinition: aws.String(arn),
		})

		if err != nil {
			resp.Diagnostics.AddError(
				create.ProblemStandardMessage(names.Batch, create.ErrActionDeleting, ResNameJobDefinition, state.ID.String(), err),
				err.Error(),
			)
			return
		}
	}
}

func (r *resourceJobDefinition) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root(names.AttrID), req, resp)
}

func (r *resourceJobDefinition) ModifyPlan(ctx context.Context, request resource.ModifyPlanRequest, response *resource.ModifyPlanResponse) {
	r.SetTagsAll(ctx, request, response)
}

type resourceJobDefinitionModel struct {
	ARN                     types.String                                                `tfsdk:"arn"`
	ArnPrefix               types.String                                                `tfsdk:"arn_prefix" autoflex:"-"`
	ContainerProperties     jsontypes.Normalized                                        `tfsdk:"container_properties" autoflex:"-"`
	DeregisterOnNewRevision types.Bool                                                  `tfsdk:"deregister_on_new_revision" autoflex:"-"`
	ECSProperties           types.String                                                `tfsdk:"ecs_properties" autoflex:"-"`
	EKSProperties           fwtypes.ListNestedObjectValueOf[eksResourcePropertiesModel] `tfsdk:"eks_properties"`
	ID                      types.String                                                `tfsdk:"id" autoflex:"-"`
	Name                    types.String                                                `tfsdk:"name"`
	NodeProperties          types.String                                                `tfsdk:"node_properties" autoflex:"-"`
	Parameters              types.Map                                                   `tfsdk:"parameters" autoflex:",legacy"`
	PlatformCapabilities    types.Set                                                   `tfsdk:"platform_capabilities"`
	PropagateTags           types.Bool                                                  `tfsdk:"propagate_tags" autoflex:",legacy"`
	Revision                types.Int32                                                 `tfsdk:"revision"`
	RetryStrategy           fwtypes.ListNestedObjectValueOf[retryStrategyModel]         `tfsdk:"retry_strategy"`
	SchedulingPriority      types.Int32                                                 `tfsdk:"scheduling_priority"`
	Tags                    tftags.Map                                                  `tfsdk:"tags"`
	TagsAll                 tftags.Map                                                  `tfsdk:"tags_all"`
	Timeout                 fwtypes.ListNestedObjectValueOf[jobTimeoutModel]            `tfsdk:"timeout"`
	Type                    types.String                                                `tfsdk:"type"`
}

// The following custom validators are designed to ensure objects can marshal. Once converted away from string these can be removed
// Custom validator for ECS Properties JSON
type ecsPropertiesValidator struct{}

func (v ecsPropertiesValidator) Description(ctx context.Context) string {
	return "must be a valid ECS properties JSON document"
}

func (v ecsPropertiesValidator) MarkdownDescription(ctx context.Context) string {
	return "must be a valid ECS properties JSON document"
}

func (v ecsPropertiesValidator) ValidateString(ctx context.Context, req validator.StringRequest, resp *validator.StringResponse) {
	if req.ConfigValue.IsNull() || req.ConfigValue.IsUnknown() {
		return
	}

	if _, err := expandECSProperties(req.ConfigValue.ValueString()); err != nil {
		resp.Diagnostics.AddAttributeError(
			req.Path,
			"Invalid ECS Properties",
			fmt.Sprintf("Unable to parse ECS properties JSON: %s", err),
		)
	}
}

type containerPropertiesValidator struct{}

func (v containerPropertiesValidator) Description(ctx context.Context) string {
	return "must be a valid container properties JSON document"
}

func (v containerPropertiesValidator) MarkdownDescription(ctx context.Context) string {
	return "must be a valid container properties JSON document"
}

func (v containerPropertiesValidator) ValidateString(ctx context.Context, req validator.StringRequest, resp *validator.StringResponse) {
	if req.ConfigValue.IsNull() || req.ConfigValue.IsUnknown() {
		return
	}

	if _, err := expandContainerProperties(req.ConfigValue.ValueString()); err != nil {
		resp.Diagnostics.AddAttributeError(
			req.Path,
			"Invalid Container Properties",
			fmt.Sprintf("Unable to parse container properties JSON: %s", err),
		)
	}
}

type nodePropertiesValidator struct{}

func (v nodePropertiesValidator) Description(ctx context.Context) string {
	return "must be a valid node properties JSON document"
}
func (v nodePropertiesValidator) MarkdownDescription(ctx context.Context) string {
	return "must be a valid node properties JSON document"
}
func (v nodePropertiesValidator) ValidateString(ctx context.Context, req validator.StringRequest, resp *validator.StringResponse) {
	if req.ConfigValue.IsNull() || req.ConfigValue.IsUnknown() {
		return
	}

	if _, err := expandJobNodeProperties(req.ConfigValue.ValueString()); err != nil {
		resp.Diagnostics.AddAttributeError(
			req.Path,
			"Invalid Node Properties",
			fmt.Sprintf("Unable to parse node properties JSON: %s", err),
		)
	}
}

// Helper Functions
func findJobDefinitionByARN(ctx context.Context, conn *batch.Client, arn string) (*awstypes.JobDefinition, error) {
	const (
		jobDefinitionStatusInactive = "INACTIVE"
	)
	input := &batch.DescribeJobDefinitionsInput{
		JobDefinitions: []string{arn},
	}

	output, err := findJobDefinition(ctx, conn, input)

	if err != nil {
		return nil, err
	}

	if status := aws.ToString(output.Status); status == jobDefinitionStatusInactive {
		return nil, &retry.NotFoundError{
			Message:     status,
			LastRequest: input,
		}
	}

	return output, nil
}

func findJobDefinition(ctx context.Context, conn *batch.Client, input *batch.DescribeJobDefinitionsInput) (*awstypes.JobDefinition, error) {
	output, err := findJobDefinitions(ctx, conn, input)

	if err != nil {
		return nil, err
	}

	return tfresource.AssertSingleValueResult(output)
}

func findJobDefinitions(ctx context.Context, conn *batch.Client, input *batch.DescribeJobDefinitionsInput) ([]awstypes.JobDefinition, error) {
	var output []awstypes.JobDefinition

	pages := batch.NewDescribeJobDefinitionsPaginator(conn, input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if err != nil {
			return nil, err
		}

		output = append(output, page.JobDefinitions...)
	}

	return output, nil
}

type eksResourcePropertiesModel struct {
	PodProperties fwtypes.ListNestedObjectValueOf[eksResourcePodPropertiesModel] `tfsdk:"pod_properties"`
}

type eksResourcePodPropertiesModel struct {
	Containers  fwtypes.ListNestedObjectValueOf[eksContainerModel] `tfsdk:"containers"`
	DNSPolicy   types.String                                       `tfsdk:"dns_policy"`
	HostNetwork types.Bool                                         `tfsdk:"host_network"`
	// in the resource, its image_poll_secret but in the datasource its image_pull_secrets
	ImagePullSecrets      fwtypes.ListNestedObjectValueOf[eksImagePullSecrets] `tfsdk:"image_pull_secret"`
	InitContainers        fwtypes.ListNestedObjectValueOf[eksContainerModel]   `tfsdk:"init_containers"`
	Metadata              fwtypes.ListNestedObjectValueOf[eksMetadataModel]    `tfsdk:"metadata"`
	ServiceAccountName    types.String                                         `tfsdk:"service_account_name"`
	ShareProcessNamespace types.Bool                                           `tfsdk:"share_process_namespace"`
	Volumes               fwtypes.ListNestedObjectValueOf[eksVolumeModel]      `tfsdk:"volumes"`
}
