// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package bedrockagent

import (
	"context"
	"errors"
	"fmt"
	"github.com/hashicorp/terraform-plugin-framework-timetypes/timetypes"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/id"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/bedrockagent"
	awstypes "github.com/aws/aws-sdk-go-v2/service/bedrockagent/types"
	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/listplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
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

// @FrameworkResource(name="Bedrock Agent")
// @Tags(identifierAttribute="agent_arn")
func newBedrockAgentResource(context.Context) (resource.ResourceWithConfigure, error) {
	r := &bedrockAgentResource{}

	r.SetDefaultCreateTimeout(5 * time.Minute)
	r.SetDefaultUpdateTimeout(5 * time.Minute)
	r.SetDefaultDeleteTimeout(5 * time.Minute)

	return r, nil
}

type bedrockAgentResource struct {
	framework.ResourceWithConfigure
	framework.WithImportByID
	framework.WithTimeouts
}

func (r *bedrockAgentResource) Metadata(_ context.Context, request resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = "aws_bedrockagent_agent"
}

func (r *bedrockAgentResource) Schema(ctx context.Context, request resource.SchemaRequest, response *resource.SchemaResponse) {
	response.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"agent_arn": framework.ARNAttributeComputedOnly(),
			"agent_id":  framework.IDAttribute(),
			"agent_name": schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"agent_resource_role_arn": schema.StringAttribute{
				CustomType: fwtypes.ARNType,
				Required:   true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"agent_status": schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"agent_version": schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"created_at": schema.StringAttribute{
				CustomType: timetypes.RFC3339Type{},
				Computed:   true,
			},
			"customer_encryption_key_arn": schema.StringAttribute{
				CustomType: fwtypes.ARNType,
				Optional:   true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"description": schema.StringAttribute{
				Optional: true,
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"failure_reasons": schema.ListAttribute{
				ElementType: types.StringType,
				Computed:    true,
				PlanModifiers: []planmodifier.List{
					listplanmodifier.UseStateForUnknown(),
				},
			},
			"foundation_model": schema.StringAttribute{
				//CustomType: fwtypes.ARNType,
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			names.AttrID: framework.IDAttribute(),
			"idle_ttl": schema.Int64Attribute{
				Required: true,
			},
			"instruction": schema.StringAttribute{
				Optional: true,
				Computed: true,
			},
			"prepared_at": schema.StringAttribute{
				CustomType: timetypes.RFC3339Type{},
				Computed:   true,
			},
			"prompt_override_configuration": schema.ListAttribute{ // Limited here by V5 Protocol
				Computed:   true,
				Optional:   true,
				Validators: []validator.List{listvalidator.SizeAtMost(1)},
				CustomType: fwtypes.NewListNestedObjectTypeOf[poc](ctx),
				ElementType: types.ObjectType{
					AttrTypes: map[string]attr.Type{
						"override_lambda": types.StringType,
						"prompt_configurations": types.ListType{
							ElemType: types.ObjectType{
								AttrTypes: map[string]attr.Type{
									"base_prompt_template": types.StringType,
									"parser_mode":          types.StringType,
									"prompt_creation_mode": types.StringType,
									"prompt_state":         types.StringType,
									"prompt_type":          types.StringType,
									"inference_configuration": types.ObjectType{
										AttrTypes: map[string]attr.Type{
											"max_length": types.Int64Type,
											"stop_sequences": types.ListType{
												ElemType: types.StringType,
											},
											"temperature": types.Float64Type,
											"topk":        types.Int64Type,
											"topp":        types.Float64Type,
										},
									},
								},
							},
						},
					},
				},
			},
			"recommended_actions": schema.ListAttribute{
				ElementType: types.StringType,
				Computed:    true,
				PlanModifiers: []planmodifier.List{
					listplanmodifier.UseStateForUnknown(),
				},
			},
			"updated_at": schema.StringAttribute{
				CustomType: timetypes.RFC3339Type{},
				Computed:   true,
			},
			names.AttrTags:    tftags.TagsAttribute(),
			names.AttrTagsAll: tftags.TagsAttributeComputedOnly(),
		},
		Blocks: map[string]schema.Block{
			"timeouts": timeouts.Block(ctx, timeouts.Opts{
				Create: true,
				Update: true,
				Delete: true,
			}),
		},
	}
}

func (r *bedrockAgentResource) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	var data bedrockAgentResourceModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().BedrockAgentClient(ctx)

	input := &bedrockagent.CreateAgentInput{}
	response.Diagnostics.Append(fwflex.Expand(ctx, data, input)...)
	if response.Diagnostics.HasError() {
		return
	}

	input.ClientToken = aws.String(id.UniqueId())

	input.Tags = getTagsIn(ctx)

	output, err := conn.CreateAgent(ctx, input)

	if err != nil {
		response.Diagnostics.AddError("creating Bedrock Agent", err.Error())

		return
	}

	data.AgentARN = fwflex.StringToFramework(ctx, output.Agent.AgentArn)
	data.AgentId = fwflex.StringToFramework(ctx, output.Agent.AgentId)
	data.ID = data.AgentId
	agent, err := waitAgentCreated(ctx, conn, data.ID.ValueString(), r.CreateTimeout(ctx, data.Timeouts))
	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("waiting for Bedrock Agent (%s) create", data.ID.ValueString()), err.Error())

		return
	}

	response.Diagnostics.Append(fwflex.Flatten(ctx, agent.Agent, &data)...)
	response.Diagnostics.Append(response.State.Set(ctx, &data)...)
}

func (r *bedrockAgentResource) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	var data bedrockAgentResourceModel
	response.Diagnostics.Append(request.State.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	if data.ID.IsNull() {
		response.Diagnostics.AddError("parsing resource ID", "Bedrock Agent ID")

		return
	}

	conn := r.Meta().BedrockAgentClient(ctx)

	AgentId := data.ID.ValueString()
	output, err := findBedrockAgentByID(ctx, conn, AgentId)

	if tfresource.NotFound(err) {
		response.Diagnostics.Append(fwdiag.NewResourceNotFoundWarningDiagnostic(err))
		response.State.RemoveResource(ctx)

		return
	}

	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("reading Bedrock Agent (%s)", AgentId), err.Error())

		return
	}

	response.Diagnostics.Append(fwflex.Flatten(ctx, output.Agent, &data)...)

	if response.Diagnostics.HasError() {
		return
	}

	response.Diagnostics.Append(response.State.Set(ctx, &data)...)
}

func (r *bedrockAgentResource) Update(ctx context.Context, request resource.UpdateRequest, response *resource.UpdateResponse) {
	var old, new bedrockAgentResourceModel
	response.Diagnostics.Append(request.State.Get(ctx, &old)...)
	if response.Diagnostics.HasError() {
		return
	}
	response.Diagnostics.Append(request.Plan.Get(ctx, &new)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().BedrockAgentClient(ctx)

	if bedrockAgentHasChanges(ctx, old, new) {
		input := &bedrockagent.UpdateAgentInput{}
		response.Diagnostics.Append(fwflex.Expand(ctx, new, input)...)
		if response.Diagnostics.HasError() {
			return
		}

		_, err := conn.UpdateAgent(ctx, input)

		if err != nil {
			response.Diagnostics.AddError(
				create.ProblemStandardMessage(names.BedrockAgent, create.ErrActionUpdating, "Bedrock Agent", old.AgentId.ValueString(), err),
				err.Error(),
			)
			return
		}
	}

	out, err := findBedrockAgentByID(ctx, conn, old.ID.ValueString())
	if err != nil {
		response.Diagnostics.AddError(
			create.ProblemStandardMessage(names.BedrockAgent, create.ErrActionUpdating, "Bedrock Agent", old.AgentId.ValueString(), err),
			err.Error(),
		)
		return
	}
	response.Diagnostics.Append(fwflex.Flatten(ctx, out, &new)...)
	if response.Diagnostics.HasError() {
		return
	}

	response.Diagnostics.Append(response.State.Set(ctx, &new)...)
}

func (r *bedrockAgentResource) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	var data bedrockAgentResourceModel
	response.Diagnostics.Append(request.State.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().BedrockAgentClient(ctx)

	if !data.AgentId.IsNull() {
		AgentId := data.ID.ValueString()
		_, err := conn.DeleteAgent(ctx, &bedrockagent.DeleteAgentInput{
			AgentId: fwflex.StringFromFramework(ctx, data.AgentId),
		})

		if errs.IsA[*awstypes.ResourceNotFoundException](err) {
			return
		}

		if _, err := waitAgentDeleted(ctx, conn, AgentId, r.DeleteTimeout(ctx, data.Timeouts)); err != nil {
			response.Diagnostics.AddError(fmt.Sprintf("waiting for Bedrock Custom Model customization job (%s) stop", AgentId), err.Error())

			return
		}

		if err != nil {
			response.Diagnostics.AddError(fmt.Sprintf("deleting Bedrock Agent (%s)", data.ID.ValueString()), err.Error())

			return
		}
	}
}

func (r *bedrockAgentResource) ModifyPlan(ctx context.Context, request resource.ModifyPlanRequest, response *resource.ModifyPlanResponse) {
	r.SetTagsAll(ctx, request, response)
}

func findBedrockAgentByID(ctx context.Context, conn *bedrockagent.Client, id string) (*bedrockagent.GetAgentOutput, error) {
	input := &bedrockagent.GetAgentInput{
		AgentId: aws.String(id),
	}

	output, err := conn.GetAgent(ctx, input)

	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	return output, nil
}

func statusAgent(ctx context.Context, conn *bedrockagent.Client, id string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		output, err := findBedrockAgentByID(ctx, conn, id)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, string(output.Agent.AgentStatus), nil
	}
}

func waitAgentCreated(ctx context.Context, conn *bedrockagent.Client, id string, timeout time.Duration) (*bedrockagent.GetAgentOutput, error) {
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(awstypes.AgentStatusCreating),
		Target:  enum.Slice(awstypes.AgentStatusNotPrepared),
		Refresh: statusAgent(ctx, conn, id),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*bedrockagent.GetAgentOutput); ok {
		tfresource.SetLastError(err, errors.New(aws.ToString((*string)(&output.Agent.AgentStatus))))

		return output, err
	}

	return nil, err
}

func waitAgentDeleted(ctx context.Context, conn *bedrockagent.Client, id string, timeout time.Duration) (*bedrockagent.GetAgentOutput, error) {
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(awstypes.AgentStatusDeleting, awstypes.AgentStatusCreating),
		Target:  []string{},
		Refresh: statusAgent(ctx, conn, id),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*bedrockagent.GetAgentOutput); ok {
		tfresource.SetLastError(err, errors.New(aws.ToString((*string)(&output.Agent.AgentStatus))))

		return output, err
	}

	return nil, err
}

type bedrockAgentResourceModel struct {
	AgentARN                    types.String                         `tfsdk:"agent_arn"`
	AgentId                     types.String                         `tfsdk:"agent_id"`
	AgentName                   types.String                         `tfsdk:"agent_name"`
	AgentResourceRoleARN        fwtypes.ARN                          `tfsdk:"agent_resource_role_arn"`
	AgentVersion                types.String                         `tfsdk:"agent_version"`
	AgentStatus                 types.String                         `tfsdk:"agent_status"`
	CreatedAt                   timetypes.RFC3339                    `tfsdk:"created_at"`
	CustomerEncryptionKeyARN    fwtypes.ARN                          `tfsdk:"customer_encryption_key_arn"`
	Description                 types.String                         `tfsdk:"description"`
	FailureReasons              types.List                           `tfsdk:"failure_reasons"`
	FoundationModel             types.String                         `tfsdk:"foundation_model"`
	ID                          types.String                         `tfsdk:"id"`
	IdleSessionTTLInSeconds     types.Int64                          `tfsdk:"idle_ttl"`
	Instruction                 types.String                         `tfsdk:"instruction"`
	PreparedAt                  timetypes.RFC3339                    `tfsdk:"prepared_at"`
	PromptOverrideConfiguration fwtypes.ListNestedObjectValueOf[poc] `tfsdk:"prompt_override_configuration"`
	RecommendedActions          types.List                           `tfsdk:"recommended_actions"`
	UpdatedAt                   timetypes.RFC3339                    `tfsdk:"updated_at"`
	Tags                        types.Map                            `tfsdk:"tags"`
	TagsAll                     types.Map                            `tfsdk:"tags_all"`
	Timeouts                    timeouts.Value                       `tfsdk:"timeouts"`
}

func (m *bedrockAgentResourceModel) setId() {
	m.ID = m.AgentId
}

type poc struct {
	OverrideLambda       fwtypes.ARN                         `tfsdk:"override_lambda"`
	PromptConfigurations fwtypes.ListNestedObjectValueOf[pc] `tfsdk:"prompt_configurations"`
}

type pc struct {
	BasePromptTemplate     types.String                        `tfsdk:"base_prompt_template"`
	ParserMode             types.String                        `tfsdk:"parser_mode"`
	PromptCreationMode     types.String                        `tfsdk:"prompt_creation_mode"`
	PromptState            types.String                        `tfsdk:"prompt_state"`
	PromptType             types.String                        `tfsdk:"prompt_type"`
	InferenceConfiguration fwtypes.ListNestedObjectValueOf[ic] `tfsdk:"inference_configuration"`
}

type ic struct {
	MaximumLength types.Int64                       `tfsdk:"max_length"`
	StopSequences fwtypes.ListValueOf[types.String] `tfsdk:"stop_sequences"`
	Temperature   types.Float64                     `tfsdk:"temperature"`
	TopK          types.Int64                       `tfsdk:"topk"`
	TopP          types.Float64                     `tfsdk:"topp"`
}

func bedrockAgentHasChanges(_ context.Context, plan, state bedrockAgentResourceModel) bool {
	return !plan.CustomerEncryptionKeyARN.Equal(state.CustomerEncryptionKeyARN) ||
		!plan.Description.Equal(state.Description)
}
