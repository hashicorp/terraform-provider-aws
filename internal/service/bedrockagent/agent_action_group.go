// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package bedrockagent

import (
	"context"
	"fmt"
	"github.com/hashicorp/terraform-plugin-framework-timetypes/timetypes"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	intflex "github.com/hashicorp/terraform-provider-aws/internal/flex"
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

const (
	agentActionGroupIdParts = 3
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
			"created_at": schema.StringAttribute{
				CustomType: timetypes.RFC3339Type{},
				Computed:   true,
			},
			"description": schema.StringAttribute{
				Optional: true,
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
			"skip_resource_in_use_check": schema.BoolAttribute{
				Computed: true,
				Default:  booldefault.StaticBool(false),
				Optional: true,
			},
			"updated_at": schema.StringAttribute{
				CustomType: timetypes.RFC3339Type{},
				Computed:   true,
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

	var input bedrockagent.CreateAgentActionGroupInput
	response.Diagnostics.Append(fwflex.Expand(ctx, data, &input)...)

	apiSchemaInput, diags := expandApiSchema(ctx, data.APISchema)
	response.Diagnostics.Append(diags...)

	if response.Diagnostics.HasError() {
		return
	}

	input.ApiSchema = apiSchemaInput

	actionGroupExecutorInput, diags := expandActionGroupExecutor(ctx, data.ActionGroupExecutor)
	response.Diagnostics.Append(diags...)

	if response.Diagnostics.HasError() {
		return
	}

	input.ActionGroupExecutor = actionGroupExecutorInput

	output, err := conn.CreateAgentActionGroup(ctx, &input)

	if err != nil {
		response.Diagnostics.AddError("creating Bedrock Agent Group", err.Error())

		return
	}

	response.Diagnostics.Append(fwflex.Flatten(ctx, output.AgentActionGroup, &data)...)

	err = data.setId()
	if err != nil {
		response.Diagnostics.AddError("creating Bedrock Agent Group", err.Error())

		return
	}

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

	output, err := findAgentActionGroupByID(ctx, conn, data.ID.ValueString())

	if tfresource.NotFound(err) {
		response.Diagnostics.Append(fwdiag.NewResourceNotFoundWarningDiagnostic(err))
		response.State.RemoveResource(ctx)

		return
	}

	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("reading Bedrock Agent (%s)", data.ID.ValueString()), err.Error())

		return
	}

	response.Diagnostics.Append(fwflex.Flatten(ctx, output.AgentActionGroup, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	apiSchemaData, moreDiags := flattenApiSchema(ctx, output.AgentActionGroup.ApiSchema)
	response.Diagnostics.Append(moreDiags...)
	if response.Diagnostics.HasError() {
		return
	}

	data.APISchema = apiSchemaData

	ageData, moreDiags := flattenActionGroupExecutor(ctx, output.AgentActionGroup.ActionGroupExecutor)
	response.Diagnostics.Append(moreDiags...)
	if response.Diagnostics.HasError() {
		return
	}

	data.ActionGroupExecutor = ageData

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

		apiConfig, diags := expandApiSchema(ctx, new.APISchema)
		response.Diagnostics.Append(diags...)

		if response.Diagnostics.HasError() {
			return
		}

		input.ApiSchema = apiConfig

		ageConfig, diags := expandActionGroupExecutor(ctx, new.ActionGroupExecutor)
		response.Diagnostics.Append(diags...)

		if response.Diagnostics.HasError() {
			return
		}

		input.ActionGroupExecutor = ageConfig

		_, err := conn.UpdateAgentActionGroup(ctx, input)

		if err != nil {
			response.Diagnostics.AddError(
				create.ProblemStandardMessage(names.BedrockAgent, create.ErrActionUpdating, "Bedrock Agent", old.AgentId.ValueString(), err),
				err.Error(),
			)
			return
		}
	}

	out, err := findAgentActionGroupByID(ctx, conn, old.ID.ValueString())
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
			AgentId:                fwflex.StringFromFramework(ctx, data.AgentId),
			ActionGroupId:          fwflex.StringFromFramework(ctx, data.ActionGroupId),
			AgentVersion:           fwflex.StringFromFramework(ctx, data.AgentVersion),
			SkipResourceInUseCheck: data.SkipResourceInUseCheck.ValueBool(),
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

func findAgentActionGroupByID(ctx context.Context, conn *bedrockagent.Client, id string) (*bedrockagent.GetAgentActionGroupOutput, error) {
	parts, err := intflex.ExpandResourceId(id, agentActionGroupIdParts, false)

	if err != nil {
		return nil, err
	}

	actionGroupId := parts[0]
	agentId := parts[1]
	agentVersion := parts[2]

	input := &bedrockagent.GetAgentActionGroupInput{
		ActionGroupId: aws.String(actionGroupId),
		AgentId:       aws.String(agentId),
		AgentVersion:  aws.String(agentVersion),
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
	CreatedAt                  timetypes.RFC3339                                    `tfsdk:"created_at"`
	Description                types.String                                         `tfsdk:"description"`
	ID                         types.String                                         `tfsdk:"id"`
	ParentActionGroupSignature types.String                                         `tfsdk:"parent_action_group_signature"`
	SkipResourceInUseCheck     types.Bool                                           `tfsdk:"skip_resource_in_use_check"`
	UpdatedAt                  timetypes.RFC3339                                    `tfsdk:"updated_at"`
}

func (a *actionGroupResourceModel) setId() error {
	id, err := intflex.FlattenResourceId([]string{a.ActionGroupId.ValueString(), a.AgentId.ValueString(), a.AgentVersion.ValueString()}, agentActionGroupIdParts, false)
	if err != nil {
		return err
	}
	a.ID = types.StringValue(id)
	return nil
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

func expandActionGroupExecutor(ctx context.Context, age fwtypes.ListNestedObjectValueOf[actionGroupExecutor]) (awstypes.ActionGroupExecutor, diag.Diagnostics) {
	var diags diag.Diagnostics
	var ageObject awstypes.ActionGroupExecutor
	planAge, moreDiags := age.ToPtr(ctx)

	diags.Append(moreDiags...)
	if diags.HasError() {
		return ageObject, diags
	}

	if !planAge.Lambda.IsNull() {
		var lambdaAge awstypes.ActionGroupExecutorMemberLambda
		diags.Append(fwflex.Expand(ctx, planAge.Lambda, &lambdaAge.Value)...)
		ageObject = &lambdaAge
	}

	return ageObject, diags
}

func flattenActionGroupExecutor(ctx context.Context, age awstypes.ActionGroupExecutor) (fwtypes.ListNestedObjectValueOf[actionGroupExecutor], diag.Diagnostics) {
	var ageData actionGroupExecutor
	switch v := age.(type) {
	case *awstypes.ActionGroupExecutorMemberLambda:
		ageData.Lambda = fwflex.StringValueToFramework(ctx, v.Value)
	}

	return fwtypes.NewListNestedObjectValueOfPtr(ctx, &ageData)
}

func expandApiSchema(ctx context.Context, api fwtypes.ListNestedObjectValueOf[apiSchema]) (awstypes.APISchema, diag.Diagnostics) {
	var diags diag.Diagnostics
	var apiObject awstypes.APISchema
	planApi, moreDiags := api.ToPtr(ctx)
	diags.Append(moreDiags...)
	if diags.HasError() {
		return apiObject, diags
	}

	if !planApi.S3.IsNull() {
		var s3 awstypes.APISchemaMemberS3
		diags.Append(fwflex.Expand(ctx, planApi.S3, &s3.Value)...)
		apiObject = &s3
	}

	if !planApi.Payload.IsNull() {
		var payload awstypes.APISchemaMemberPayload
		diags.Append(fwflex.Expand(ctx, planApi.Payload, &payload.Value)...)
		apiObject = &payload
	}
	return apiObject, diags
}

func flattenApiSchema(ctx context.Context, apiObject awstypes.APISchema) (fwtypes.ListNestedObjectValueOf[apiSchema], diag.Diagnostics) {
	var diags diag.Diagnostics
	apiSchemaData := &apiSchema{
		S3:      fwtypes.NewListNestedObjectValueOfNull[s3](ctx),
		Payload: types.StringNull(),
	}

	switch v := apiObject.(type) {
	case *awstypes.APISchemaMemberS3:
		var s3data s3
		diags.Append(fwflex.Flatten(ctx, v, s3data)...)
		apiSchemaData.S3 = fwtypes.NewListNestedObjectValueOfPtrMust(ctx, &s3data)
	case *awstypes.APISchemaMemberPayload:
		payloadValue := fwflex.StringValueToFramework(ctx, v.Value)
		apiSchemaData.Payload = payloadValue
	}

	return fwtypes.NewListNestedObjectValueOfPtrMust(ctx, apiSchemaData), diags
}
