// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package bedrockagent

import (
	"context"
	"fmt"
	"time"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/bedrockagent"
	awstypes "github.com/aws/aws-sdk-go-v2/service/bedrockagent/types"
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/fwdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	fwflex "github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkResource(name="Agent Action Group")
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

func (*agentActionGroupResource) Metadata(_ context.Context, request resource.MetadataRequest, response *resource.MetadataResponse) {
	response.TypeName = "aws_bedrockagent_agent_action_group"
}

func (r *agentActionGroupResource) Schema(ctx context.Context, request resource.SchemaRequest, response *resource.SchemaResponse) {
	response.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"action_group_id": framework.IDAttribute(),
			"action_group_name": schema.StringAttribute{
				Required: true,
				Validators: []validator.String{
					stringvalidator.RegexMatches(regexache.MustCompile(`^([0-9a-zA-Z][_-]?){1,100}$`), "valid characters are a-z, A-Z, 0-9, _ (underscore) and - (hyphen). The name can have up to 100 characters"),
				},
			},
			"action_group_state": schema.StringAttribute{
				CustomType: fwtypes.StringEnumType[awstypes.ActionGroupState](),
				Optional:   true,
				Computed:   true,
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
			names.AttrDescription: schema.StringAttribute{
				Optional: true,
				Validators: []validator.String{
					stringvalidator.LengthBetween(1, 200),
				},
			},
			names.AttrID: framework.IDAttribute(),
			"parent_action_group_signature": schema.StringAttribute{
				CustomType: fwtypes.StringEnumType[awstypes.ActionGroupSignature](),
				Optional:   true,
			},
			"skip_resource_in_use_check": schema.BoolAttribute{
				Optional: true,
				Computed: true,
				Default:  booldefault.StaticBool(false),
			},
		},
		Blocks: map[string]schema.Block{
			"action_group_executor": schema.ListNestedBlock{
				CustomType: fwtypes.NewListNestedObjectTypeOf[actionGroupExecutorModel](ctx),
				Validators: []validator.List{
					listvalidator.SizeAtMost(1),
				},
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"lambda": schema.StringAttribute{
							CustomType: fwtypes.ARNType,
							Optional:   true,
						},
					},
				},
			},
			"api_schema": schema.ListNestedBlock{
				CustomType: fwtypes.NewListNestedObjectTypeOf[apiSchemaModel](ctx),
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
						},
					},
					Blocks: map[string]schema.Block{
						"s3": schema.ListNestedBlock{
							CustomType: fwtypes.NewListNestedObjectTypeOf[s3IdentifierModel](ctx),
							Validators: []validator.List{
								listvalidator.SizeAtMost(1),
								listvalidator.ConflictsWith(
									path.MatchRelative().AtParent().AtName("payload"),
								),
							},
							NestedObject: schema.NestedBlockObject{
								Attributes: map[string]schema.Attribute{
									names.AttrS3BucketName: schema.StringAttribute{
										Optional: true,
									},
									"s3_object_key": schema.StringAttribute{
										Optional: true,
									},
								},
							},
						},
					},
				},
			},
			"function_schema": schema.ListNestedBlock{
				CustomType: fwtypes.NewListNestedObjectTypeOf[functionSchemaModel](ctx),
				Validators: []validator.List{
					listvalidator.SizeAtMost(1),
				},
				NestedObject: schema.NestedBlockObject{
					Blocks: map[string]schema.Block{
						"functions": schema.ListNestedBlock{
							CustomType: fwtypes.NewListNestedObjectTypeOf[functionModel](ctx),
							NestedObject: schema.NestedBlockObject{
								Attributes: map[string]schema.Attribute{
									names.AttrName: schema.StringAttribute{
										Required: true,
										Validators: []validator.String{
											stringvalidator.RegexMatches(regexache.MustCompile(`^([0-9a-zA-Z][_-]?){1,100}$`), "valid characters are a-z, A-Z, 0-9, _ (underscore) and - (hyphen). The name can have up to 100 characters"),
										},
									},
									names.AttrDescription: schema.StringAttribute{
										Optional: true,
										Validators: []validator.String{
											stringvalidator.LengthBetween(1, 1200),
										},
									},
								},
								Blocks: map[string]schema.Block{
									names.AttrParameters: schema.SetNestedBlock{
										CustomType: fwtypes.NewSetNestedObjectTypeOf[parameterDetailModel](ctx),
										NestedObject: schema.NestedBlockObject{
											Attributes: map[string]schema.Attribute{
												"map_block_key": schema.StringAttribute{
													Required: true,
													Validators: []validator.String{
														stringvalidator.RegexMatches(regexache.MustCompile(`^([0-9a-zA-Z][_-]?){1,100}$`), "valid characters are a-z, A-Z, 0-9, _ (underscore) and - (hyphen). The name can have up to 100 characters"),
													},
												},
												names.AttrType: schema.StringAttribute{
													Required:   true,
													CustomType: fwtypes.StringEnumType[awstypes.Type](),
												},
												names.AttrDescription: schema.StringAttribute{
													Optional: true,
													Validators: []validator.String{
														stringvalidator.LengthBetween(1, 500),
													},
												},
												"required": schema.BoolAttribute{
													Optional: true,
												},
											},
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
	var data agentActionGroupResourceModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().BedrockAgentClient(ctx)

	input := &bedrockagent.CreateAgentActionGroupInput{}
	response.Diagnostics.Append(fwflex.Expand(ctx, data, input)...)
	if response.Diagnostics.HasError() {
		return
	}

	// AutoFlEx doesn't yet handle union types.
	if !data.ActionGroupExecutor.IsNull() {
		actionGroupExecutorData, diags := data.ActionGroupExecutor.ToPtr(ctx)
		response.Diagnostics.Append(diags...)
		if response.Diagnostics.HasError() {
			return
		}

		input.ActionGroupExecutor = expandActionGroupExecutor(ctx, actionGroupExecutorData)
	}

	if !data.APISchema.IsNull() {
		apiSchemaData, diags := data.APISchema.ToPtr(ctx)
		response.Diagnostics.Append(diags...)
		if response.Diagnostics.HasError() {
			return
		}

		input.ApiSchema = expandAPISchema(ctx, apiSchemaData)
	}

	if !data.FunctionSchema.IsNull() {
		functionSchemaData, diags := data.FunctionSchema.ToPtr(ctx)
		response.Diagnostics.Append(diags...)
		if response.Diagnostics.HasError() {
			return
		}

		functionSchema, diags := expandFunctionSchema(ctx, functionSchemaData)
		response.Diagnostics.Append(diags...)
		if response.Diagnostics.HasError() {
			return
		}
		input.FunctionSchema = functionSchema
	}

	output, err := conn.CreateAgentActionGroup(ctx, input)

	if err != nil {
		response.Diagnostics.AddError("creating Bedrock Agent Action Group", err.Error())

		return
	}

	// Set values for unknowns.
	data.ActionGroupID = fwflex.StringToFramework(ctx, output.AgentActionGroup.ActionGroupId)
	data.ActionGroupState = fwtypes.StringEnumValue(output.AgentActionGroup.ActionGroupState)
	data.setID()

	response.Diagnostics.Append(response.State.Set(ctx, &data)...)
}

func (r *agentActionGroupResource) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	var data agentActionGroupResourceModel
	response.Diagnostics.Append(request.State.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	if err := data.InitFromID(); err != nil {
		response.Diagnostics.AddError("parsing resource ID", err.Error())

		return
	}

	conn := r.Meta().BedrockAgentClient(ctx)

	output, err := findAgentActionGroupByThreePartKey(ctx, conn, data.ActionGroupID.ValueString(), data.AgentID.ValueString(), data.AgentVersion.ValueString())

	if tfresource.NotFound(err) {
		response.Diagnostics.Append(fwdiag.NewResourceNotFoundWarningDiagnostic(err))
		response.State.RemoveResource(ctx)

		return
	}

	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("reading Bedrock Agent Action Group (%s)", data.ID.ValueString()), err.Error())

		return
	}

	response.Diagnostics.Append(fwflex.Flatten(ctx, output, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	// AutoFlEx doesn't yet handle union types.
	data.ActionGroupExecutor = flattenActionGroupExecutor(ctx, output.ActionGroupExecutor)
	data.APISchema = flattenAPISchema(ctx, output.ApiSchema)

	functionSchema, diags := flattenFunctionSchema(ctx, output.FunctionSchema)
	response.Diagnostics.Append(diags...)
	if response.Diagnostics.HasError() {
		return
	}
	data.FunctionSchema = functionSchema

	response.Diagnostics.Append(response.State.Set(ctx, &data)...)
}

func (r *agentActionGroupResource) Update(ctx context.Context, request resource.UpdateRequest, response *resource.UpdateResponse) {
	var old, new agentActionGroupResourceModel
	response.Diagnostics.Append(request.State.Get(ctx, &old)...)
	if response.Diagnostics.HasError() {
		return
	}
	response.Diagnostics.Append(request.Plan.Get(ctx, &new)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().BedrockAgentClient(ctx)

	if !new.ActionGroupExecutor.Equal(old.ActionGroupExecutor) ||
		!new.ActionGroupName.Equal(old.ActionGroupName) ||
		!new.ActionGroupState.Equal(old.ActionGroupState) ||
		!new.APISchema.Equal(old.APISchema) ||
		!new.Description.Equal(old.Description) ||
		!new.FunctionSchema.Equal(old.FunctionSchema) ||
		!new.ParentActionGroupSignature.Equal(old.ParentActionGroupSignature) {
		input := &bedrockagent.UpdateAgentActionGroupInput{}
		response.Diagnostics.Append(fwflex.Expand(ctx, new, input)...)
		if response.Diagnostics.HasError() {
			return
		}

		// AutoFlEx doesn't yet handle union types.
		if !new.ActionGroupExecutor.IsNull() {
			actionGroupExecutorData, diags := new.ActionGroupExecutor.ToPtr(ctx)
			response.Diagnostics.Append(diags...)
			if response.Diagnostics.HasError() {
				return
			}

			input.ActionGroupExecutor = expandActionGroupExecutor(ctx, actionGroupExecutorData)
		}

		if !new.APISchema.IsNull() {
			apiSchemaData, diags := new.APISchema.ToPtr(ctx)
			response.Diagnostics.Append(diags...)
			if response.Diagnostics.HasError() {
				return
			}

			input.ApiSchema = expandAPISchema(ctx, apiSchemaData)
		}

		if !new.FunctionSchema.IsNull() {
			functionSchemaData, diags := new.FunctionSchema.ToPtr(ctx)
			response.Diagnostics.Append(diags...)
			if response.Diagnostics.HasError() {
				return
			}

			functionSchema, diags := expandFunctionSchema(ctx, functionSchemaData)
			response.Diagnostics.Append(diags...)
			if response.Diagnostics.HasError() {
				return
			}
			input.FunctionSchema = functionSchema
		}

		_, err := conn.UpdateAgentActionGroup(ctx, input)

		if err != nil {
			response.Diagnostics.AddError(fmt.Sprintf("updating Bedrock Agent Action Group (%s)", new.ID.ValueString()), err.Error())

			return
		}
	}

	output, err := findAgentActionGroupByThreePartKey(ctx, conn, new.ActionGroupID.ValueString(), new.AgentID.ValueString(), new.AgentVersion.ValueString())
	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("reading Bedrock Agent Action Group (%s)", new.ID.ValueString()), err.Error())

		return
	}

	new.ActionGroupState = fwtypes.StringEnumValue(output.ActionGroupState)

	response.Diagnostics.Append(response.State.Set(ctx, &new)...)
}

func (r *agentActionGroupResource) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	var data agentActionGroupResourceModel
	response.Diagnostics.Append(request.State.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().BedrockAgentClient(ctx)

	_, err := conn.DeleteAgentActionGroup(ctx, &bedrockagent.DeleteAgentActionGroupInput{
		ActionGroupId:          fwflex.StringFromFramework(ctx, data.ActionGroupID),
		AgentId:                fwflex.StringFromFramework(ctx, data.AgentID),
		AgentVersion:           fwflex.StringFromFramework(ctx, data.AgentVersion),
		SkipResourceInUseCheck: data.SkipResourceInUseCheck.ValueBool(),
	})

	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return
	}

	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("deleting Bedrock Agent Action Group (%s)", data.ID.ValueString()), err.Error())

		return
	}
}

func findAgentActionGroupByThreePartKey(ctx context.Context, conn *bedrockagent.Client, actionGroupID, agentID, agentVersion string) (*awstypes.AgentActionGroup, error) {
	input := &bedrockagent.GetAgentActionGroupInput{
		ActionGroupId: aws.String(actionGroupID),
		AgentId:       aws.String(agentID),
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

	if output == nil || output.AgentActionGroup == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	return output.AgentActionGroup, nil
}

type agentActionGroupResourceModel struct {
	ActionGroupID              types.String                                              `tfsdk:"action_group_id"`
	ActionGroupExecutor        fwtypes.ListNestedObjectValueOf[actionGroupExecutorModel] `tfsdk:"action_group_executor"`
	ActionGroupName            types.String                                              `tfsdk:"action_group_name"`
	ActionGroupState           fwtypes.StringEnum[awstypes.ActionGroupState]             `tfsdk:"action_group_state"`
	AgentID                    types.String                                              `tfsdk:"agent_id"`
	AgentVersion               types.String                                              `tfsdk:"agent_version"`
	APISchema                  fwtypes.ListNestedObjectValueOf[apiSchemaModel]           `tfsdk:"api_schema"`
	Description                types.String                                              `tfsdk:"description"`
	FunctionSchema             fwtypes.ListNestedObjectValueOf[functionSchemaModel]      `tfsdk:"function_schema"`
	ID                         types.String                                              `tfsdk:"id"`
	ParentActionGroupSignature fwtypes.StringEnum[awstypes.ActionGroupSignature]         `tfsdk:"parent_action_group_signature"`
	SkipResourceInUseCheck     types.Bool                                                `tfsdk:"skip_resource_in_use_check"`
}

const (
	agentActionGroupResourceIDPartCount = 3
)

func (m *agentActionGroupResourceModel) InitFromID() error {
	id := m.ID.ValueString()
	parts, err := flex.ExpandResourceId(id, agentActionGroupResourceIDPartCount, false)

	if err != nil {
		return err
	}

	m.ActionGroupID = types.StringValue(parts[0])
	m.AgentID = types.StringValue(parts[1])
	m.AgentVersion = types.StringValue(parts[2])

	return nil
}

func (m *agentActionGroupResourceModel) setID() {
	m.ID = types.StringValue(errs.Must(flex.FlattenResourceId([]string{m.ActionGroupID.ValueString(), m.AgentID.ValueString(), m.AgentVersion.ValueString()}, agentActionGroupResourceIDPartCount, false)))
}

type actionGroupExecutorModel struct {
	Lambda fwtypes.ARN `tfsdk:"lambda"`
}

type apiSchemaModel struct {
	Payload types.String                                       `tfsdk:"payload"`
	S3      fwtypes.ListNestedObjectValueOf[s3IdentifierModel] `tfsdk:"s3"`
}

type s3IdentifierModel struct {
	S3BucketName types.String `tfsdk:"s3_bucket_name"`
	S3ObjectKey  types.String `tfsdk:"s3_object_key"`
}

type functionSchemaModel struct {
	Functions fwtypes.ListNestedObjectValueOf[functionModel] `tfsdk:"functions"`
}

type functionModel struct {
	Name        types.String                                         `tfsdk:"name"`
	Description types.String                                         `tfsdk:"description"`
	Parameters  fwtypes.SetNestedObjectValueOf[parameterDetailModel] `tfsdk:"parameters"`
}
type parameterDetailModel struct {
	MapBlockKey types.String                      `tfsdk:"map_block_key"`
	Type        fwtypes.StringEnum[awstypes.Type] `tfsdk:"type"`
	Description types.String                      `tfsdk:"description"`
	Required    types.Bool                        `tfsdk:"required"`
}

func expandActionGroupExecutor(_ context.Context, actionGroupExecutorData *actionGroupExecutorModel) awstypes.ActionGroupExecutor {
	if !actionGroupExecutorData.Lambda.IsNull() {
		return &awstypes.ActionGroupExecutorMemberLambda{
			Value: actionGroupExecutorData.Lambda.ValueString(),
		}
	}

	return nil
}

func flattenActionGroupExecutor(ctx context.Context, apiObject awstypes.ActionGroupExecutor) fwtypes.ListNestedObjectValueOf[actionGroupExecutorModel] {
	if apiObject == nil {
		return fwtypes.NewListNestedObjectValueOfNull[actionGroupExecutorModel](ctx)
	}

	var actionGroupExecutorData actionGroupExecutorModel

	switch v := apiObject.(type) {
	case *awstypes.ActionGroupExecutorMemberLambda:
		actionGroupExecutorData.Lambda = fwtypes.ARNValue(v.Value)
	}

	return fwtypes.NewListNestedObjectValueOfPtrMust(ctx, &actionGroupExecutorData)
}

func expandAPISchema(ctx context.Context, apiSchemaData *apiSchemaModel) awstypes.APISchema {
	if !apiSchemaData.Payload.IsNull() {
		return &awstypes.APISchemaMemberPayload{
			Value: apiSchemaData.Payload.ValueString(),
		}
	}

	if !apiSchemaData.S3.IsNull() {
		s3IdentifierModel := fwdiag.Must(apiSchemaData.S3.ToPtr(ctx))

		return &awstypes.APISchemaMemberS3{
			Value: awstypes.S3Identifier{
				S3BucketName: fwflex.StringFromFramework(ctx, s3IdentifierModel.S3BucketName),
				S3ObjectKey:  fwflex.StringFromFramework(ctx, s3IdentifierModel.S3ObjectKey),
			},
		}
	}

	return nil
}

func flattenAPISchema(ctx context.Context, apiObject awstypes.APISchema) fwtypes.ListNestedObjectValueOf[apiSchemaModel] {
	if apiObject == nil {
		return fwtypes.NewListNestedObjectValueOfNull[apiSchemaModel](ctx)
	}

	var apiSchemaData apiSchemaModel

	switch v := apiObject.(type) {
	case *awstypes.APISchemaMemberPayload:
		apiSchemaData.Payload = fwflex.StringValueToFramework(ctx, v.Value)
		apiSchemaData.S3 = fwtypes.NewListNestedObjectValueOfNull[s3IdentifierModel](ctx)

	case *awstypes.APISchemaMemberS3:
		apiSchemaData.Payload = types.StringNull()
		apiSchemaData.S3 = fwtypes.NewListNestedObjectValueOfPtrMust(ctx, &s3IdentifierModel{
			S3BucketName: fwflex.StringToFramework(ctx, v.Value.S3BucketName),
			S3ObjectKey:  fwflex.StringToFramework(ctx, v.Value.S3ObjectKey),
		})
	}

	return fwtypes.NewListNestedObjectValueOfPtrMust(ctx, &apiSchemaData)
}

func expandFunctionSchema(ctx context.Context, functionSchemaData *functionSchemaModel) (awstypes.FunctionSchema, diag.Diagnostics) {
	var diags diag.Diagnostics

	if !functionSchemaData.Functions.IsNull() {
		functions := functionSchemaData.Functions
		memberFunctions := make([]awstypes.Function, len(functions.Elements()))
		diags.Append(fwflex.Expand(ctx, functions, &memberFunctions)...)
		if diags.HasError() {
			return nil, diags
		}

		return &awstypes.FunctionSchemaMemberFunctions{
			Value: memberFunctions,
		}, diags
	}

	return nil, diags
}

func flattenFunctionSchema(ctx context.Context, apiObject awstypes.FunctionSchema) (fwtypes.ListNestedObjectValueOf[functionSchemaModel], diag.Diagnostics) {
	var diags diag.Diagnostics

	if apiObject == nil {
		return fwtypes.NewListNestedObjectValueOfNull[functionSchemaModel](ctx), diags
	}

	var functionSchemaData functionSchemaModel

	switch v := apiObject.(type) {
	case *awstypes.FunctionSchemaMemberFunctions:
		memberFunctions := fwtypes.NewListNestedObjectValueOfSliceMust[functionModel](ctx, []*functionModel{})
		diags.Append(fwflex.Flatten(ctx, v.Value, &memberFunctions)...)
		if diags.HasError() {
			return fwtypes.NewListNestedObjectValueOfNull[functionSchemaModel](ctx), diags
		}

		functionSchemaData.Functions = memberFunctions
	}

	return fwtypes.NewListNestedObjectValueOfPtrMust(ctx, &functionSchemaData), diags
}
