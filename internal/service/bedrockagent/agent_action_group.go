// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package bedrockagent

import (
	"context"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/bedrockagent"
	awstypes "github.com/aws/aws-sdk-go-v2/service/bedrockagent/types"
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/listplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/fwdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	fwflex "github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkResource(name="Bedrock Agent Action Group")
func newAgentActionGroupResource(context.Context) (resource.ResourceWithConfigure, error) {
	r := &agentActionGroupResource{}

	r.SetDefaultDeleteTimeout(120 * time.Minute)

	return r, nil
}

type agentActionGroupResource struct {
	framework.ResourceWithConfigure
	framework.WithImportByID
	framework.WithTimeouts
}

func (r *agentActionGroupResource) Metadata(_ context.Context, request resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = "aws_bedrockagent_agent_action_group"
}

func (r *agentActionGroupResource) Schema(ctx context.Context, request resource.SchemaRequest, response *resource.SchemaResponse) {
	response.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"action_group_id": framework.IDAttribute(),
			"action_group_name": schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"action_group_state": schema.StringAttribute{
				Optional: true,
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},

			"agent_id": schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"agent_version": schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},

			"client_token": schema.StringAttribute{
				Optional: true,
			},
			"created_at": schema.StringAttribute{
				Computed: true,
			},
			"description": schema.StringAttribute{
				Optional: true,
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			names.AttrID: framework.IDAttribute(),
			"parent_action_group_signature": schema.StringAttribute{
				Optional: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"prepared_at": schema.StringAttribute{
				Computed: true,
			},
			"updated_at": schema.StringAttribute{
				Computed: true,
			},
		},
		Blocks: map[string]schema.Block{
			"action_group_executor": schema.ListNestedBlock{
				CustomType: fwtypes.NewListNestedObjectTypeOf[actionGroupExecutor](ctx),
				PlanModifiers: []planmodifier.List{
					listplanmodifier.RequiresReplace(),
				},
				Validators: []validator.List{
					listvalidator.SizeAtMost(1),
				},
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"lambda": schema.StringAttribute{
							Required: true,
							PlanModifiers: []planmodifier.String{
								stringplanmodifier.RequiresReplace(),
							},
						},
					},
				},
			},
			"api_schema": schema.ListNestedBlock{
				CustomType: fwtypes.NewListNestedObjectTypeOf[apiSchema](ctx),
				PlanModifiers: []planmodifier.List{
					listplanmodifier.RequiresReplace(),
				},
				Validators: []validator.List{
					listvalidator.SizeAtMost(1),
				},
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"payload": schema.StringAttribute{
							Optional: true,
							Validators: []validator.String{
								stringvalidator.ConflictsWith(
									path.MatchRelative().AtParent().AtName("s3"),
								),
							},
							PlanModifiers: []planmodifier.String{
								stringplanmodifier.UseStateForUnknown(),
							},
						},
					},
					Blocks: map[string]schema.Block{
						"s3": schema.ListNestedBlock{
							CustomType: fwtypes.NewListNestedObjectTypeOf[s3](ctx),
							Validators: []validator.List{
								listvalidator.SizeAtMost(1),
								listvalidator.ConflictsWith(
									path.MatchRelative().AtParent().AtName("payload"),
								),
							},
							NestedObject: schema.NestedBlockObject{
								Attributes: map[string]schema.Attribute{
									"s3_bucket_name": schema.StringAttribute{
										Optional: true,
										PlanModifiers: []planmodifier.String{
											stringplanmodifier.UseStateForUnknown(),
										},
									},
									"s3_object_key": schema.StringAttribute{
										Optional: true,
										PlanModifiers: []planmodifier.String{
											stringplanmodifier.UseStateForUnknown(),
										},
									},
								},
							},
						},
					},
				},
			},
		},
	}
}

func (r *agentActionGroupResource) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	var data actionGroupResourceModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().BedrockAgentClient(ctx)

	// input := &bedrockagent.CreateAgentActionGroupInput{}
	// response.Diagnostics.Append(fwflex.Expand(ctx, data, input)...)
	// if response.Diagnostics.HasError() {
	// 	return
	// }

	// if !data.ActionGroupExecutor.IsNull() {
	// 	age, diags := data.ActionGroupExecutor.ToPtr(ctx)
	// 	response.Diagnostics.Append(diags...)
	// 	if response.Diagnostics.HasError() {
	// 		return
	// 	}
	// 	input.ActionGroupExecutor = expandAge(age)
	// }

	age, diags := data.ActionGroupExecutor.ToPtr(ctx)
	response.Diagnostics.Append(diags...)
	if response.Diagnostics.HasError() {
		return
	}
	lambda := &awstypes.ActionGroupExecutorMemberLambda{
		Value: *fwflex.StringFromFramework(ctx, age.Lambda),
	}

	apischema, diags := data.APISchema.ToPtr(ctx)
	response.Diagnostics.Append(diags...)
	if response.Diagnostics.HasError() {
		return
	}

	s3, diags := apischema.S3.ToPtr(ctx)
	response.Diagnostics.Append(diags...)
	if response.Diagnostics.HasError() {
		return
	}
	s3data := &awstypes.APISchemaMemberS3{}
	response.Diagnostics.Append(fwflex.Expand(ctx, s3, &s3data.Value)...)
	if response.Diagnostics.HasError() {
		return
	}

	input := &bedrockagent.CreateAgentActionGroupInput{
		ActionGroupName:     data.ActionGroupName.ValueStringPointer(),
		AgentId:             data.AgentId.ValueStringPointer(),
		AgentVersion:        data.AgentVersion.ValueStringPointer(),
		ActionGroupExecutor: lambda,
		ApiSchema:           s3data,
	}

	output, err := conn.CreateAgentActionGroup(ctx, input)

	if err != nil {
		response.Diagnostics.AddError("creating Bedrock Agent Group", err.Error())

		return
	}

	// Set values for unknowns.
	//data.ActionGroupId = fwflex.StringToFramework(ctx, output.AgentActionGroup.ActionGroupId)
	//data.ID = data.ActionGroupId

	var dataFromCreate actionGroupResourceModel
	response.Diagnostics.Append(fwflex.Flatten(ctx, output.AgentActionGroup, &dataFromCreate)...)
	data.CreatedAt = dataFromCreate.CreatedAt
	data.UpdatedAt = dataFromCreate.UpdatedAt
	data.PreparedAt = dataFromCreate.PreparedAt
	data.ActionGroupState = dataFromCreate.ActionGroupState
	data.Description = dataFromCreate.Description
	data.AgentId = dataFromCreate.AgentId
	data.ActionGroupId = dataFromCreate.ActionGroupId
	data.AgentVersion = dataFromCreate.AgentVersion
	data.ID = dataFromCreate.ActionGroupId

	response.Diagnostics.Append(response.State.Set(ctx, &data)...)
}

func (r *agentActionGroupResource) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	var data actionGroupResourceModel
	response.Diagnostics.Append(request.State.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	if data.ID.IsNull() {
		response.Diagnostics.AddError("parsing resource ID", "Action Group ID")

		return
	}

	conn := r.Meta().BedrockAgentClient(ctx)

	AgentId := data.AgentId.ValueString()
	ActionGroupId := data.ActionGroupId.ValueString()
	AgentVersion := data.AgentVersion.ValueString()

	output, err := findAgentActionGroupByID(ctx, conn, ActionGroupId, AgentId, AgentVersion)

	if tfresource.NotFound(err) {
		response.Diagnostics.Append(fwdiag.NewResourceNotFoundWarningDiagnostic(err))
		response.State.RemoveResource(ctx)

		return
	}

	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("reading Bedrock Agent (%s)", AgentId), err.Error())

		return
	}

	response.Diagnostics.Append(fwflex.Flatten(ctx, output, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	response.Diagnostics.Append(response.State.Set(ctx, &data)...)
}

func (r *agentActionGroupResource) Update(ctx context.Context, request resource.UpdateRequest, response *resource.UpdateResponse) {
	var old, new actionGroupResourceModel
	response.Diagnostics.Append(request.State.Get(ctx, &old)...)
	if response.Diagnostics.HasError() {
		return
	}
	response.Diagnostics.Append(request.Plan.Get(ctx, &new)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().BedrockAgentClient(ctx)

	if agentActionGroupHasChanges(ctx, old, new) {
		input := &bedrockagent.UpdateAgentActionGroupInput{}
		response.Diagnostics.Append(fwflex.Expand(ctx, new, input)...)
		if response.Diagnostics.HasError() {
			return
		}

		_, err := conn.UpdateAgentActionGroup(ctx, input)

		if err != nil {
			response.Diagnostics.AddError(
				create.ProblemStandardMessage(names.BedrockAgent, create.ErrActionUpdating, "Bedrock Agent", old.AgentId.ValueString(), err),
				err.Error(),
			)
			return
		}
	}

	out, err := findAgentActionGroupByID(ctx, conn, old.ActionGroupId.String(), old.ID.ValueString(), old.AgentVersion.ValueString())
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

func (r *agentActionGroupResource) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	var data actionGroupResourceModel
	response.Diagnostics.Append(request.State.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().BedrockAgentClient(ctx)

	if !data.ActionGroupId.IsNull() {
		_, err := conn.DeleteAgentActionGroup(ctx, &bedrockagent.DeleteAgentActionGroupInput{
			AgentId:       fwflex.StringFromFramework(ctx, data.AgentId),
			ActionGroupId: fwflex.StringFromFramework(ctx, data.ActionGroupId),
			AgentVersion:  fwflex.StringFromFramework(ctx, data.AgentVersion),
		})

		if errs.IsA[*awstypes.ResourceNotFoundException](err) {
			return
		}

		if err != nil {
			response.Diagnostics.AddError(fmt.Sprintf("deleting Bedrock Agent (%s)", data.ID.ValueString()), err.Error())

			return
		}
	}
}

// func (r *agentActionGroupResource) ModifyPlan(ctx context.Context, request resource.ModifyPlanRequest, response *resource.ModifyPlanResponse) {
// 	r.SetTagsAll(ctx, request, response)
// }

func findAgentActionGroupByID(ctx context.Context, conn *bedrockagent.Client, grpid, agentid, agentversion string) (*bedrockagent.GetAgentActionGroupOutput, error) {
	input := &bedrockagent.GetAgentActionGroupInput{
		ActionGroupId: aws.String(grpid),
		AgentId:       aws.String(agentid),
		AgentVersion:  aws.String(agentversion),
	}

	output, err := conn.GetAgentActionGroup(ctx, input)

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

type actionGroupResourceModel struct {
	ActionGroupId              types.String                                         `tfsdk:"action_group_id"`
	ActionGroupExecutor        fwtypes.ListNestedObjectValueOf[actionGroupExecutor] `tfsdk:"action_group_executor"`
	ActionGroupName            types.String                                         `tfsdk:"action_group_name"`
	ActionGroupState           types.String                                         `tfsdk:"action_group_state"`
	AgentId                    types.String                                         `tfsdk:"agent_id"`
	AgentVersion               types.String                                         `tfsdk:"agent_version"`
	APISchema                  fwtypes.ListNestedObjectValueOf[apiSchema]           `tfsdk:"api_schema"`
	ClientToken                types.String                                         `tfsdk:"client_token"`
	CreatedAt                  types.String                                         `tfsdk:"created_at"`
	Description                types.String                                         `tfsdk:"description"`
	ID                         types.String                                         `tfsdk:"id"`
	ParentActionGroupSignature types.String                                         `tfsdk:"parent_action_group_signature"`
	PreparedAt                 types.String                                         `tfsdk:"prepared_at"`
	UpdatedAt                  types.String                                         `tfsdk:"updated_at"`
}

type actionGroupExecutor struct {
	Lambda types.String `tfsdk:"lambda"`
}

type apiSchema struct {
	Payload types.String                        `tfsdk:"payload"`
	S3      fwtypes.ListNestedObjectValueOf[s3] `tfsdk:"s3"`
}

type s3 struct {
	S3BucketName types.String `tfsdk:"s3_bucket_name"`
	S3ObjectKey  types.String `tfsdk:"s3_object_key"`
}

func agentActionGroupHasChanges(_ context.Context, plan, state actionGroupResourceModel) bool {
	return !plan.ActionGroupName.Equal(state.ActionGroupName) ||
		!plan.Description.Equal(state.Description)
}

func expandAge(agedata *actionGroupExecutor) awstypes.ActionGroupExecutor {
	if !agedata.Lambda.IsNull() {
		return &awstypes.ActionGroupExecutorMemberLambda{
			Value: agedata.Lambda.ValueString(),
		}
	}
	return nil
}
