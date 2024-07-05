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
						"custom_control": schema.StringAttribute{
							CustomType: fwtypes.StringEnumType[awstypes.CustomControlMethod](),
							Optional:   true,
						},
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
						"member_functions": schema.ListNestedBlock{
							CustomType: fwtypes.NewListNestedObjectTypeOf[memberFunctionsModel](ctx),
							Validators: []validator.List{
								listvalidator.SizeAtMost(1),
							},
							NestedObject: schema.NestedBlockObject{
								Blocks: map[string]schema.Block{
									"functions": schema.ListNestedBlock{
										CustomType: fwtypes.NewListNestedObjectTypeOf[functionModel](ctx),
										NestedObject: schema.NestedBlockObject{
											Attributes: map[string]schema.Attribute{
												names.AttrDescription: schema.StringAttribute{
													Optional: true,
													Validators: []validator.String{
														stringvalidator.LengthBetween(1, 1200),
													},
												},
												names.AttrName: schema.StringAttribute{
													Required: true,
													Validators: []validator.String{
														stringvalidator.RegexMatches(regexache.MustCompile(`^([0-9a-zA-Z][_-]?){1,100}$`), "valid characters are a-z, A-Z, 0-9, _ (underscore) and - (hyphen). The name can have up to 100 characters"),
													},
												},
											},
											Blocks: map[string]schema.Block{
												names.AttrParameters: schema.SetNestedBlock{
													CustomType: fwtypes.NewSetNestedObjectTypeOf[parameterDetailModel](ctx),
													NestedObject: schema.NestedBlockObject{
														Attributes: map[string]schema.Attribute{
															names.AttrDescription: schema.StringAttribute{
																Optional: true,
																Validators: []validator.String{
																	stringvalidator.LengthBetween(1, 500),
																},
															},
															"map_block_key": schema.StringAttribute{
																Required: true,
																Validators: []validator.String{
																	stringvalidator.RegexMatches(regexache.MustCompile(`^([0-9a-zA-Z][_-]?){1,100}$`), "valid characters are a-z, A-Z, 0-9, _ (underscore) and - (hyphen). The name can have up to 100 characters"),
																},
															},
															"required": schema.BoolAttribute{
																Optional: true,
															},
															names.AttrType: schema.StringAttribute{
																Required:   true,
																CustomType: fwtypes.StringEnumType[awstypes.Type](),
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
	CustomControl fwtypes.StringEnum[awstypes.CustomControlMethod] `tfsdk:"custom_control"`
	Lambda        fwtypes.ARN                                      `tfsdk:"lambda"`
}

var (
	_ fwflex.Expander  = actionGroupExecutorModel{}
	_ fwflex.Flattener = &actionGroupExecutorModel{}
)

func (m actionGroupExecutorModel) Expand(ctx context.Context) (result any, diags diag.Diagnostics) {
	switch {
	case !m.CustomControl.IsNull():
		return &awstypes.ActionGroupExecutorMemberCustomControl{
			Value: m.CustomControl.ValueEnum(),
		}, diags
	case !m.Lambda.IsNull():
		return &awstypes.ActionGroupExecutorMemberLambda{
			Value: m.Lambda.ValueString(),
		}, diags
	}

	return nil, diags
}

func (m *actionGroupExecutorModel) Flatten(ctx context.Context, v any) (diags diag.Diagnostics) {
	switch t := v.(type) {
	case awstypes.ActionGroupExecutorMemberCustomControl:
		m.CustomControl = fwtypes.StringEnumValue(t.Value)

	case awstypes.ActionGroupExecutorMemberLambda:
		m.Lambda = fwtypes.ARNValue(t.Value)

		return diags
	}

	return diags
}

type apiSchemaModel struct {
	Payload types.String                                       `tfsdk:"payload"`
	S3      fwtypes.ListNestedObjectValueOf[s3IdentifierModel] `tfsdk:"s3"`
}

var (
	_ fwflex.Expander  = apiSchemaModel{}
	_ fwflex.Flattener = &apiSchemaModel{}
)

func (m apiSchemaModel) Expand(ctx context.Context) (result any, diags diag.Diagnostics) {
	switch {
	case !m.Payload.IsNull():
		return &awstypes.APISchemaMemberPayload{
			Value: m.Payload.ValueString(),
		}, diags

	case !m.S3.IsNull():
		s3IdentifierModel := fwdiag.Must(m.S3.ToPtr(ctx))

		return &awstypes.APISchemaMemberS3{
			Value: awstypes.S3Identifier{
				S3BucketName: fwflex.StringFromFramework(ctx, s3IdentifierModel.S3BucketName),
				S3ObjectKey:  fwflex.StringFromFramework(ctx, s3IdentifierModel.S3ObjectKey),
			},
		}, diags
	}

	return nil, diags
}

func (m *apiSchemaModel) Flatten(ctx context.Context, v any) (diags diag.Diagnostics) {
	switch t := v.(type) {
	case awstypes.APISchemaMemberPayload:
		m.Payload = fwflex.StringToFramework(ctx, &t.Value)

		return diags

	case awstypes.APISchemaMemberS3:
		var model s3IdentifierModel
		d := fwflex.Flatten(ctx, t.Value, &model)
		diags.Append(d...)
		if diags.HasError() {
			return diags
		}

		m.S3 = fwtypes.NewListNestedObjectValueOfPtrMust(ctx, &model)

		return diags
	}

	return diags
}

type s3IdentifierModel struct {
	S3BucketName types.String `tfsdk:"s3_bucket_name"`
	S3ObjectKey  types.String `tfsdk:"s3_object_key"`
}

var (
	_ fwflex.Expander  = functionSchemaModel{}
	_ fwflex.Flattener = &functionSchemaModel{}
)

type functionSchemaModel struct {
	MemberFunctions fwtypes.ListNestedObjectValueOf[memberFunctionsModel] `tfsdk:"member_functions"`
}

func (m functionSchemaModel) Expand(ctx context.Context) (result any, diags diag.Diagnostics) {
	switch {
	case !m.MemberFunctions.IsNull():
		memberFunctionsModel := fwdiag.Must(m.MemberFunctions.ToPtr(ctx))
		var functions []awstypes.Function
		diags.Append(fwflex.Expand(ctx, memberFunctionsModel.Functions, &functions)...)
		if diags.HasError() {
			return nil, diags
		}

		return &awstypes.FunctionSchemaMemberFunctions{
			Value: functions,
		}, diags
	}

	return nil, diags
}

func (m *functionSchemaModel) Flatten(ctx context.Context, v any) (diags diag.Diagnostics) {
	m.MemberFunctions = fwtypes.NewListNestedObjectValueOfNull[memberFunctionsModel](ctx)

	switch t := v.(type) {
	case awstypes.FunctionSchemaMemberFunctions:
		var functions fwtypes.ListNestedObjectValueOf[functionModel]
		diags.Append(fwflex.Flatten(ctx, t.Value, &functions)...)
		if diags.HasError() {
			return diags
		}

		memberFunctions := memberFunctionsModel{
			Functions: functions,
		}

		m.MemberFunctions = fwtypes.NewListNestedObjectValueOfPtrMust(ctx, &memberFunctions)

		return diags
	}

	return diags
}

type memberFunctionsModel struct {
	Functions fwtypes.ListNestedObjectValueOf[functionModel] `tfsdk:"functions"`
}

type functionModel struct {
	Description types.String                                         `tfsdk:"description"`
	Name        types.String                                         `tfsdk:"name"`
	Parameters  fwtypes.SetNestedObjectValueOf[parameterDetailModel] `tfsdk:"parameters"`
}

type parameterDetailModel struct {
	Description types.String                      `tfsdk:"description"`
	MapBlockKey types.String                      `tfsdk:"map_block_key"`
	Required    types.Bool                        `tfsdk:"required"`
	Type        fwtypes.StringEnum[awstypes.Type] `tfsdk:"type"`
}
