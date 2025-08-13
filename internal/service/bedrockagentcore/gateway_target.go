// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package bedrockagentcore

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/YakDriver/smarterr"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/bedrockagentcorecontrol"
	awstypes "github.com/aws/aws-sdk-go-v2/service/bedrockagentcorecontrol/types"
	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/setvalidator"
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
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/fwdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	"github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/smerr"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// Function annotations are used for resource registration to the Provider. DO NOT EDIT.
// @FrameworkResource("aws_bedrockagentcore_gateway_target", name="Gateway Target")
func newResourceGatewayTarget(_ context.Context) (resource.ResourceWithConfigure, error) {
	r := &resourceGatewayTarget{}

	r.SetDefaultCreateTimeout(30 * time.Minute)
	r.SetDefaultUpdateTimeout(30 * time.Minute)
	r.SetDefaultDeleteTimeout(30 * time.Minute)

	return r, nil
}

const (
	ResNameGatewayTarget = "Gateway Target"
)

type resourceGatewayTarget struct {
	framework.ResourceWithModel[resourceGatewayTargetModel]
	framework.WithTimeouts
}

func jsonAttribute(conflictWith string) schema.StringAttribute {
	return schema.StringAttribute{
		Optional:      true,
		PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
		Validators: []validator.String{
			// Use AtParent() so the conflict path resolves to the sibling attribute,
			// not a child of the current attribute. Otherwise the framework builds
			// paths like `items_json.properties_json`, which is invalid.
			stringvalidator.ConflictsWith(path.MatchRelative().AtParent().AtName(conflictWith)),
		},
	}
}

func createLeafItemsBlock[T any](ctx context.Context) schema.ListNestedBlock {
	return schema.ListNestedBlock{
		CustomType: fwtypes.NewListNestedObjectTypeOf[T](ctx),
		Validators: []validator.List{
			listvalidator.SizeAtMost(1),
			listvalidator.ConflictsWith(path.MatchRelative().AtParent().AtName("property")),
		},
		NestedObject: schema.NestedBlockObject{
			Attributes: map[string]schema.Attribute{
				names.AttrDescription: schema.StringAttribute{Optional: true},
				names.AttrType: schema.StringAttribute{
					Required:   true,
					CustomType: fwtypes.StringEnumType[awstypes.SchemaType](),
				},
				"items_json":      jsonAttribute("properties_json"),
				"properties_json": jsonAttribute("items_json"),
			},
		},
	}
}

func createLeafPropertyBlock[T any](ctx context.Context) schema.SetNestedBlock {
	return schema.SetNestedBlock{
		CustomType: fwtypes.NewSetNestedObjectTypeOf[T](ctx),
		Validators: []validator.Set{
			setvalidator.ConflictsWith(path.MatchRelative().AtParent().AtName("items")),
		},
		NestedObject: schema.NestedBlockObject{
			Attributes: map[string]schema.Attribute{
				names.AttrName:        schema.StringAttribute{Required: true},
				names.AttrDescription: schema.StringAttribute{Optional: true},
				names.AttrType: schema.StringAttribute{
					Required:   true,
					CustomType: fwtypes.StringEnumType[awstypes.SchemaType](),
				},
				"required": schema.BoolAttribute{
					Optional: true,
					Computed: true,
					Default:  booldefault.StaticBool(false),
				},
				"items_json":      jsonAttribute("properties_json"),
				"properties_json": jsonAttribute("items_json"),
			},
		},
	}
}

func schemaDefinitionNestedBlock(ctx context.Context) schema.NestedBlockObject {
	return schema.NestedBlockObject{
		Attributes: map[string]schema.Attribute{
			names.AttrDescription: schema.StringAttribute{Optional: true},
			names.AttrType: schema.StringAttribute{
				Required:   true,
				CustomType: fwtypes.StringEnumType[awstypes.SchemaType](),
			},
		},
		Blocks: map[string]schema.Block{
			"property": schema.SetNestedBlock{
				CustomType: fwtypes.NewSetNestedObjectTypeOf[schemaPropertyModel](ctx),
				Validators: []validator.Set{
					setvalidator.ConflictsWith(path.MatchRelative().AtParent().AtName("items")),
				},
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						names.AttrName:        schema.StringAttribute{Required: true},
						names.AttrDescription: schema.StringAttribute{Optional: true},
						names.AttrType: schema.StringAttribute{
							Required:   true,
							CustomType: fwtypes.StringEnumType[awstypes.SchemaType](),
						},
						"required": schema.BoolAttribute{
							Optional: true,
							Computed: true,
							Default:  booldefault.StaticBool(false),
						},
					},
					Blocks: map[string]schema.Block{
						"items": schema.ListNestedBlock{
							CustomType: fwtypes.NewListNestedObjectTypeOf[schemaItemsModel](ctx),
							Validators: []validator.List{
								listvalidator.SizeAtMost(1),
								listvalidator.ConflictsWith(path.MatchRelative().AtParent().AtName("property")),
							},
							NestedObject: schema.NestedBlockObject{
								Attributes: map[string]schema.Attribute{
									names.AttrDescription: schema.StringAttribute{Optional: true},
									names.AttrType: schema.StringAttribute{
										Required:   true,
										CustomType: fwtypes.StringEnumType[awstypes.SchemaType](),
									},
								},
								Blocks: map[string]schema.Block{
									"items":    createLeafItemsBlock[schemaItemsLeafModel](ctx),
									"property": createLeafPropertyBlock[schemaPropertyLeafModel](ctx),
								},
							},
						},
						"property": createLeafPropertyBlock[schemaPropertyLeafModel](ctx),
					},
				},
			},
			"items": schema.ListNestedBlock{
				CustomType: fwtypes.NewListNestedObjectTypeOf[schemaItemsModel](ctx),
				Validators: []validator.List{
					listvalidator.SizeAtMost(1),
					listvalidator.ConflictsWith(path.MatchRelative().AtParent().AtName("property")),
				},
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						names.AttrDescription: schema.StringAttribute{Optional: true},
						names.AttrType: schema.StringAttribute{
							Required:   true,
							CustomType: fwtypes.StringEnumType[awstypes.SchemaType](),
						},
					},
					Blocks: map[string]schema.Block{
						"items":    createLeafItemsBlock[schemaItemsLeafModel](ctx),
						"property": createLeafPropertyBlock[schemaPropertyLeafModel](ctx),
					},
				},
			},
		},
	}
}

func (r *resourceGatewayTarget) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"client_token": schema.StringAttribute{
				Optional: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			names.AttrDescription: schema.StringAttribute{
				Optional: true,
			},
			"gateway_identifier": schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			names.AttrID: framework.IDAttribute(),
			names.AttrName: schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
		},
		Blocks: map[string]schema.Block{
			"credential_provider_configuration": schema.ListNestedBlock{
				CustomType: fwtypes.NewListNestedObjectTypeOf[credentialProviderConfigurationModel](ctx),
				Validators: []validator.List{
					listvalidator.IsRequired(),
					listvalidator.SizeAtMost(1),
				},
				NestedObject: schema.NestedBlockObject{
					Blocks: map[string]schema.Block{
						"api_key": schema.ListNestedBlock{
							CustomType: fwtypes.NewListNestedObjectTypeOf[apiKeyCredentialProviderModel](ctx),
							Validators: []validator.List{
								listvalidator.SizeAtMost(1),
								listvalidator.ConflictsWith(
									path.MatchRelative().AtParent().AtName("oauth"),
									path.MatchRelative().AtParent().AtName("gateway_iam_role"),
								),
							},
							NestedObject: schema.NestedBlockObject{
								Attributes: map[string]schema.Attribute{
									"credential_location": schema.StringAttribute{
										Optional:   true,
										CustomType: fwtypes.StringEnumType[awstypes.ApiKeyCredentialLocation](),
									},
									"credential_parameter_name": schema.StringAttribute{
										Optional: true,
									},
									"credential_prefix": schema.StringAttribute{
										Optional: true,
									},
									"provider_arn": schema.StringAttribute{
										Required: true,
									},
								},
							},
						},
						"oauth": schema.ListNestedBlock{
							CustomType: fwtypes.NewListNestedObjectTypeOf[oauthCredentialProviderModel](ctx),
							Validators: []validator.List{
								listvalidator.SizeAtMost(1),
								listvalidator.ConflictsWith(
									path.MatchRelative().AtParent().AtName("api_key"),
									path.MatchRelative().AtParent().AtName("gateway_iam_role"),
								),
							},
							NestedObject: schema.NestedBlockObject{
								Attributes: map[string]schema.Attribute{
									"custom_parameters": schema.MapAttribute{
										CustomType: fwtypes.MapOfStringType,
										Optional:   true,
									},
									"provider_arn": schema.StringAttribute{
										Required: true,
									},
									"scopes": schema.SetAttribute{
										CustomType: fwtypes.SetOfStringType,
										Required:   true,
									},
								},
							},
						},
						"gateway_iam_role": schema.ListNestedBlock{
							CustomType: fwtypes.NewListNestedObjectTypeOf[gatewayIAMRoleProviderModel](ctx),
							Validators: []validator.List{
								listvalidator.SizeAtMost(1),
								listvalidator.ConflictsWith(
									path.MatchRelative().AtParent().AtName("api_key"),
									path.MatchRelative().AtParent().AtName("oauth"),
								),
							},
							NestedObject: schema.NestedBlockObject{
								// Empty block - no attributes needed for Gateway IAM Role
							},
						},
					},
				},
			},
			"target_configuration": schema.ListNestedBlock{
				CustomType: fwtypes.NewListNestedObjectTypeOf[targetConfigurationModel](ctx),
				Validators: []validator.List{
					listvalidator.IsRequired(),
					listvalidator.SizeAtMost(1),
				},
				NestedObject: schema.NestedBlockObject{
					Blocks: map[string]schema.Block{
						"mcp": schema.ListNestedBlock{
							CustomType: fwtypes.NewListNestedObjectTypeOf[mcpConfigurationModel](ctx),
							Validators: []validator.List{
								listvalidator.SizeAtMost(1),
							},
							NestedObject: schema.NestedBlockObject{
								Blocks: map[string]schema.Block{
									"lambda": schema.ListNestedBlock{
										CustomType: fwtypes.NewListNestedObjectTypeOf[mcpLambdaConfigurationModel](ctx),
										Validators: []validator.List{
											listvalidator.SizeAtMost(1),
										},
										NestedObject: schema.NestedBlockObject{
											Attributes: map[string]schema.Attribute{
												"lambda_arn": schema.StringAttribute{
													Required: true,
												},
											},
											Blocks: map[string]schema.Block{
												"tool_schema": schema.ListNestedBlock{
													CustomType: fwtypes.NewListNestedObjectTypeOf[toolSchemaModel](ctx),
													Validators: []validator.List{
														listvalidator.IsRequired(),
														listvalidator.SizeAtMost(1),
													},
													NestedObject: schema.NestedBlockObject{
														Blocks: map[string]schema.Block{
															"inline_payload": schema.ListNestedBlock{
																CustomType: fwtypes.NewListNestedObjectTypeOf[toolDefinitionModel](ctx),
																NestedObject: schema.NestedBlockObject{
																	Attributes: map[string]schema.Attribute{
																		names.AttrDescription: schema.StringAttribute{
																			Required: true,
																		},
																		names.AttrName: schema.StringAttribute{
																			Required: true,
																		},
																	},
																	Blocks: map[string]schema.Block{
																		"input_schema": schema.ListNestedBlock{
																			CustomType: fwtypes.NewListNestedObjectTypeOf[schemaDefinitionModel](ctx),
																			Validators: []validator.List{
																				listvalidator.IsRequired(),
																				listvalidator.SizeAtMost(1),
																			},
																			NestedObject: schemaDefinitionNestedBlock(ctx),
																		},
																		"output_schema": schema.ListNestedBlock{
																			CustomType: fwtypes.NewListNestedObjectTypeOf[schemaDefinitionModel](ctx),
																			Validators: []validator.List{
																				listvalidator.SizeAtMost(1),
																			},
																			NestedObject: schemaDefinitionNestedBlock(ctx),
																		},
																	},
																},
															},
															"s3": schema.ListNestedBlock{
																CustomType: fwtypes.NewListNestedObjectTypeOf[s3ConfigurationModel](ctx),
																Validators: []validator.List{
																	listvalidator.SizeAtMost(1),
																},
																NestedObject: schema.NestedBlockObject{
																	Attributes: map[string]schema.Attribute{
																		"bucket_owner_account_id": schema.StringAttribute{
																			Optional: true,
																		},
																		names.AttrURI: schema.StringAttribute{
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
									"open_api_schema": schema.ListNestedBlock{
										CustomType: fwtypes.NewListNestedObjectTypeOf[apiSchemaConfigurationModel](ctx),
										Validators: []validator.List{
											listvalidator.SizeAtMost(1),
										},
										NestedObject: schema.NestedBlockObject{
											Blocks: map[string]schema.Block{
												"inline_payload": schema.ListNestedBlock{
													CustomType: fwtypes.NewListNestedObjectTypeOf[inlinePayloadModel](ctx),
													Validators: []validator.List{
														listvalidator.SizeAtMost(1),
													},
													NestedObject: schema.NestedBlockObject{
														Attributes: map[string]schema.Attribute{
															"payload": schema.StringAttribute{
																Required: true,
															},
														},
													},
												},
												"s3": schema.ListNestedBlock{
													CustomType: fwtypes.NewListNestedObjectTypeOf[s3ConfigurationModel](ctx),
													Validators: []validator.List{
														listvalidator.SizeAtMost(1),
													},
													NestedObject: schema.NestedBlockObject{
														Attributes: map[string]schema.Attribute{
															"bucket_owner_account_id": schema.StringAttribute{
																Optional: true,
															},
															names.AttrURI: schema.StringAttribute{
																Optional: true,
															},
														},
													},
												},
											},
										},
									},
									"smithy_model": schema.ListNestedBlock{
										CustomType: fwtypes.NewListNestedObjectTypeOf[apiSchemaConfigurationModel](ctx),
										Validators: []validator.List{
											listvalidator.SizeAtMost(1),
										},
										NestedObject: schema.NestedBlockObject{
											Blocks: map[string]schema.Block{
												"inline_payload": schema.ListNestedBlock{
													CustomType: fwtypes.NewListNestedObjectTypeOf[inlinePayloadModel](ctx),
													Validators: []validator.List{
														listvalidator.SizeAtMost(1),
													},
													NestedObject: schema.NestedBlockObject{
														Attributes: map[string]schema.Attribute{
															"payload": schema.StringAttribute{
																Required: true,
															},
														},
													},
												},
												"s3": schema.ListNestedBlock{
													CustomType: fwtypes.NewListNestedObjectTypeOf[s3ConfigurationModel](ctx),
													Validators: []validator.List{
														listvalidator.SizeAtMost(1),
													},
													NestedObject: schema.NestedBlockObject{
														Attributes: map[string]schema.Attribute{
															"bucket_owner_account_id": schema.StringAttribute{
																Optional: true,
															},
															names.AttrURI: schema.StringAttribute{
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
				},
			},
			names.AttrTimeouts: timeouts.Block(ctx, timeouts.Opts{
				Create: true,
				Update: true,
				Delete: true,
			}),
		},
	}
}

func (r resourceGatewayTarget) ModifyPlan(ctx context.Context, req resource.ModifyPlanRequest, resp *resource.ModifyPlanResponse) {
	// on create, state is null; on destroy, plan is null - nothing to compare
	if req.State.Raw.IsNull() || req.Plan.Raw.IsNull() {
		return
	}

	// Force replacement if target configuration changes between lambda, smithy_model, and open_api_schema
	var plan, state resourceGatewayTargetModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	planTargetData, diags := plan.TargetConfiguration.ToPtr(ctx)
	resp.Diagnostics.Append(diags...)
	stateTargetData, diags := state.TargetConfiguration.ToPtr(ctx)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	planTargetType := planTargetData.GetConfigurationType(ctx)
	stateTargetType := stateTargetData.GetConfigurationType(ctx)

	if planTargetType != stateTargetType {
		resp.RequiresReplace = append(resp.RequiresReplace, path.Root("target_configuration"))
	}
}

func (r *resourceGatewayTarget) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	conn := r.Meta().BedrockAgentCoreClient(ctx)

	var plan resourceGatewayTargetModel
	smerr.EnrichAppend(ctx, &resp.Diagnostics, req.Plan.Get(ctx, &plan))
	if resp.Diagnostics.HasError() {
		return
	}

	var input bedrockagentcorecontrol.CreateGatewayTargetInput
	smerr.EnrichAppend(ctx, &resp.Diagnostics, flex.Expand(ctx, plan, &input))
	if resp.Diagnostics.HasError() {
		return
	}

	out, err := conn.CreateGatewayTarget(ctx, &input)
	if err != nil {
		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, plan.Name.String())
		return
	}
	if out == nil {
		smerr.AddError(ctx, &resp.Diagnostics, errors.New("empty output"), smerr.ID, plan.Name.String())
		return
	}

	smerr.EnrichAppend(ctx, &resp.Diagnostics, flex.Flatten(ctx, out, &plan, flex.WithIgnoredFieldNames([]string{"GatewayArn"})))
	if resp.Diagnostics.HasError() {
		return
	}
	plan.ID = flex.StringToFramework(ctx, out.TargetId)

	createTimeout := r.CreateTimeout(ctx, plan.Timeouts)
	_, err = waitGatewayTargetCreated(ctx, conn, plan.GatewayIdentifier.ValueString(), plan.ID.ValueString(), createTimeout)
	if err != nil {
		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, plan.Name.String())
		return
	}

	smerr.EnrichAppend(ctx, &resp.Diagnostics, resp.State.Set(ctx, plan))
}

func (r *resourceGatewayTarget) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	conn := r.Meta().BedrockAgentCoreClient(ctx)

	var state resourceGatewayTargetModel
	smerr.EnrichAppend(ctx, &resp.Diagnostics, req.State.Get(ctx, &state))
	if resp.Diagnostics.HasError() {
		return
	}

	out, err := findGatewayTargetByID(ctx, conn, state.GatewayIdentifier.ValueString(), state.ID.ValueString())
	if tfresource.NotFound(err) {
		resp.Diagnostics.Append(fwdiag.NewResourceNotFoundWarningDiagnostic(err))
		resp.State.RemoveResource(ctx)
		return
	}
	if err != nil {
		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, state.ID.String())
		return
	}

	smerr.EnrichAppend(ctx, &resp.Diagnostics, flex.Flatten(ctx, out, &state, flex.WithIgnoredFieldNames([]string{"GatewayArn"})))
	if resp.Diagnostics.HasError() {
		return
	}
	state.ID = flex.StringToFramework(ctx, out.TargetId)

	smerr.EnrichAppend(ctx, &resp.Diagnostics, resp.State.Set(ctx, &state))
}

func (r *resourceGatewayTarget) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	conn := r.Meta().BedrockAgentCoreClient(ctx)

	var plan, state resourceGatewayTargetModel
	smerr.EnrichAppend(ctx, &resp.Diagnostics, req.Plan.Get(ctx, &plan))
	smerr.EnrichAppend(ctx, &resp.Diagnostics, req.State.Get(ctx, &state))
	if resp.Diagnostics.HasError() {
		return
	}

	diff, d := flex.Diff(ctx, plan, state)
	smerr.EnrichAppend(ctx, &resp.Diagnostics, d)
	if resp.Diagnostics.HasError() {
		return
	}

	if diff.HasChanges() {
		var input bedrockagentcorecontrol.UpdateGatewayTargetInput
		smerr.EnrichAppend(ctx, &resp.Diagnostics, flex.Expand(ctx, plan, &input))
		if resp.Diagnostics.HasError() {
			return
		}
		input.TargetId = plan.ID.ValueStringPointer()

		out, err := conn.UpdateGatewayTarget(ctx, &input)
		if err != nil {
			smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, plan.ID.String())
			return
		}
		if out == nil {
			smerr.AddError(ctx, &resp.Diagnostics, errors.New("empty output"), smerr.ID, plan.ID.String())
			return
		}

		smerr.EnrichAppend(ctx, &resp.Diagnostics, flex.Flatten(ctx, out, &plan, flex.WithIgnoredFieldNames([]string{"GatewayArn"})))
		if resp.Diagnostics.HasError() {
			return
		}
		plan.ID = flex.StringToFramework(ctx, out.TargetId)

		updateTimeout := r.UpdateTimeout(ctx, plan.Timeouts)
		_, err = waitGatewayTargetUpdated(ctx, conn, plan.GatewayIdentifier.ValueString(), plan.ID.ValueString(), updateTimeout)
		if err != nil {
			smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, plan.ID.String())
			return
		}
	}

	smerr.EnrichAppend(ctx, &resp.Diagnostics, resp.State.Set(ctx, &plan))
}

func (r *resourceGatewayTarget) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	conn := r.Meta().BedrockAgentCoreClient(ctx)

	var state resourceGatewayTargetModel
	smerr.EnrichAppend(ctx, &resp.Diagnostics, req.State.Get(ctx, &state))
	if resp.Diagnostics.HasError() {
		return
	}

	input := bedrockagentcorecontrol.DeleteGatewayTargetInput{
		GatewayIdentifier: state.GatewayIdentifier.ValueStringPointer(),
		TargetId:          state.ID.ValueStringPointer(),
	}

	_, err := conn.DeleteGatewayTarget(ctx, &input)
	if err != nil {
		if errs.IsA[*awstypes.ResourceNotFoundException](err) {
			return
		}

		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, state.ID.String())
		return
	}

	deleteTimeout := r.DeleteTimeout(ctx, state.Timeouts)
	_, err = waitGatewayTargetDeleted(ctx, conn, state.GatewayIdentifier.ValueString(), state.ID.ValueString(), deleteTimeout)
	if err != nil {
		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, state.ID.String())
		return
	}
}

func (r *resourceGatewayTarget) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	parts := strings.Split(req.ID, ",")

	if len(parts) != 2 {
		resp.Diagnostics.AddError("Resource Import Invalid ID", fmt.Sprintf(`Unexpected format for import ID (%s), use: "GatewayIdentifier,TargetId"`, req.ID))
		return
	}

	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("gateway_identifier"), parts[0])...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root(names.AttrID), parts[1])...)
}

func waitGatewayTargetCreated(ctx context.Context, conn *bedrockagentcorecontrol.Client, gatewayIdentifier, targetId string, timeout time.Duration) (*bedrockagentcorecontrol.GetGatewayTargetOutput, error) {
	stateConf := &retry.StateChangeConf{
		Pending:                   enum.Slice(awstypes.TargetStatusCreating),
		Target:                    enum.Slice(awstypes.TargetStatusReady),
		Refresh:                   statusGatewayTarget(ctx, conn, gatewayIdentifier, targetId),
		Timeout:                   timeout,
		NotFoundChecks:            20,
		ContinuousTargetOccurence: 2,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)
	if out, ok := outputRaw.(*bedrockagentcorecontrol.GetGatewayTargetOutput); ok {
		return out, smarterr.NewError(err)
	}

	return nil, smarterr.NewError(err)
}

func waitGatewayTargetUpdated(ctx context.Context, conn *bedrockagentcorecontrol.Client, gatewayIdentifier, targetId string, timeout time.Duration) (*bedrockagentcorecontrol.GetGatewayTargetOutput, error) {
	stateConf := &retry.StateChangeConf{
		Pending:                   enum.Slice(awstypes.TargetStatusUpdating),
		Target:                    enum.Slice(awstypes.TargetStatusReady),
		Refresh:                   statusGatewayTarget(ctx, conn, gatewayIdentifier, targetId),
		Timeout:                   timeout,
		NotFoundChecks:            20,
		ContinuousTargetOccurence: 2,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)
	if out, ok := outputRaw.(*bedrockagentcorecontrol.GetGatewayTargetOutput); ok {
		return out, smarterr.NewError(err)
	}

	return nil, smarterr.NewError(err)
}

func waitGatewayTargetDeleted(ctx context.Context, conn *bedrockagentcorecontrol.Client, gatewayIdentifier, targetId string, timeout time.Duration) (*bedrockagentcorecontrol.GetGatewayTargetOutput, error) {
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(awstypes.TargetStatusDeleting, awstypes.TargetStatusReady),
		Target:  []string{},
		Refresh: statusGatewayTarget(ctx, conn, gatewayIdentifier, targetId),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)
	if out, ok := outputRaw.(*bedrockagentcorecontrol.GetGatewayTargetOutput); ok {
		return out, smarterr.NewError(err)
	}

	return nil, smarterr.NewError(err)
}

func statusGatewayTarget(ctx context.Context, conn *bedrockagentcorecontrol.Client, gatewayIdentifier, targetId string) retry.StateRefreshFunc {
	return func() (any, string, error) {
		out, err := findGatewayTargetByID(ctx, conn, gatewayIdentifier, targetId)
		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", smarterr.NewError(err)
		}

		return out, string(out.Status), nil
	}
}

func findGatewayTargetByID(ctx context.Context, conn *bedrockagentcorecontrol.Client, gatewayIdentifier, targetId string) (*bedrockagentcorecontrol.GetGatewayTargetOutput, error) {
	input := bedrockagentcorecontrol.GetGatewayTargetInput{
		GatewayIdentifier: aws.String(gatewayIdentifier),
		TargetId:          aws.String(targetId),
	}

	out, err := conn.GetGatewayTarget(ctx, &input)
	if err != nil {
		if errs.IsA[*awstypes.ResourceNotFoundException](err) {
			return nil, smarterr.NewError(&retry.NotFoundError{
				LastError:   err,
				LastRequest: &input,
			})
		}

		return nil, smarterr.NewError(err)
	}

	if out == nil {
		return nil, smarterr.NewError(tfresource.NewEmptyResultError(&input))
	}

	return out, nil
}

type resourceGatewayTargetModel struct {
	framework.WithRegionModel
	ClientToken                     types.String                                                          `tfsdk:"client_token"`
	Description                     types.String                                                          `tfsdk:"description"`
	GatewayIdentifier               types.String                                                          `tfsdk:"gateway_identifier"`
	ID                              types.String                                                          `tfsdk:"id"`
	Name                            types.String                                                          `tfsdk:"name"`
	CredentialProviderConfiguration fwtypes.ListNestedObjectValueOf[credentialProviderConfigurationModel] `tfsdk:"credential_provider_configuration"`
	TargetConfiguration             fwtypes.ListNestedObjectValueOf[targetConfigurationModel]             `tfsdk:"target_configuration"`
	Timeouts                        timeouts.Value                                                        `tfsdk:"timeouts"`
}

type credentialProviderConfigurationModel struct {
	ApiKey         fwtypes.ListNestedObjectValueOf[apiKeyCredentialProviderModel] `tfsdk:"api_key"`
	OAuth          fwtypes.ListNestedObjectValueOf[oauthCredentialProviderModel]  `tfsdk:"oauth"`
	GatewayIAMRole fwtypes.ListNestedObjectValueOf[gatewayIAMRoleProviderModel]   `tfsdk:"gateway_iam_role"`
}

var (
	_ flex.Expander  = credentialProviderConfigurationModel{}
	_ flex.Flattener = &credentialProviderConfigurationModel{}
)

func (m *credentialProviderConfigurationModel) Flatten(ctx context.Context, v any) (diags diag.Diagnostics) {
	switch t := v.(type) {
	case awstypes.CredentialProviderConfiguration:
		switch t.CredentialProviderType {
		case awstypes.CredentialProviderTypeApiKey:
			if apiKeyProvider, ok := t.CredentialProvider.(*awstypes.CredentialProviderMemberApiKeyCredentialProvider); ok {
				var model apiKeyCredentialProviderModel
				d := flex.Flatten(ctx, apiKeyProvider.Value, &model)
				diags.Append(d...)
				if diags.HasError() {
					return diags
				}
				m.ApiKey = fwtypes.NewListNestedObjectValueOfPtrMust(ctx, &model)
			}
			return diags

		case awstypes.CredentialProviderTypeOauth:
			if oauthProvider, ok := t.CredentialProvider.(*awstypes.CredentialProviderMemberOauthCredentialProvider); ok {
				var model oauthCredentialProviderModel
				d := flex.Flatten(ctx, oauthProvider.Value, &model)
				diags.Append(d...)
				if diags.HasError() {
					return diags
				}
				m.OAuth = fwtypes.NewListNestedObjectValueOfPtrMust(ctx, &model)
			}
			return diags

		case awstypes.CredentialProviderTypeGatewayIamRole:
			m.GatewayIAMRole = fwtypes.NewListNestedObjectValueOfPtrMust(ctx, &gatewayIAMRoleProviderModel{})
			return diags

		default:
			diags.AddError(
				"Unknown Credential Provider Type",
				fmt.Sprintf("Received unknown credential provider type: %s", t.CredentialProviderType),
			)
			return diags
		}

	default:
		diags.AddError(
			"Invalid Credential Provider Configuration",
			fmt.Sprintf("Received unexpected type: %T", v),
		)
		return diags
	}
}

func (m credentialProviderConfigurationModel) Expand(ctx context.Context) (result any, diags diag.Diagnostics) {
	var c awstypes.CredentialProviderConfiguration
	switch {
	case !m.ApiKey.IsNull():
		apiKeyCredentialProviderConfigurationData, d := m.ApiKey.ToPtr(ctx)
		diags.Append(d...)
		if diags.HasError() {
			return nil, diags
		}

		var r awstypes.CredentialProviderMemberApiKeyCredentialProvider
		diags.Append(flex.Expand(ctx, apiKeyCredentialProviderConfigurationData, &r.Value)...)
		if diags.HasError() {
			return nil, diags
		}
		c.CredentialProviderType = awstypes.CredentialProviderTypeApiKey
		c.CredentialProvider = &r
		return &c, diags

	case !m.OAuth.IsNull():
		oauthCredentialProviderConfigurationData, d := m.OAuth.ToPtr(ctx)
		diags.Append(d...)
		if diags.HasError() {
			return nil, diags
		}

		var r awstypes.CredentialProviderMemberOauthCredentialProvider
		diags.Append(flex.Expand(ctx, oauthCredentialProviderConfigurationData, &r.Value)...)
		if diags.HasError() {
			return nil, diags
		}
		c.CredentialProviderType = awstypes.CredentialProviderTypeOauth
		c.CredentialProvider = &r
		return &c, diags

	case !m.GatewayIAMRole.IsNull():
		c.CredentialProviderType = awstypes.CredentialProviderTypeGatewayIamRole
		c.CredentialProvider = nil
		return &c, diags

	default:
		diags.AddError(
			"Invalid Credential Provider Configuration",
			"At least one credential provider must be configured: api_key, oauth, or gateway_iam_role",
		)
		// Return a non-nil, well-typed value so the framework doesn't raise a generic
		// error before our diagnostic reaches the user. Use GatewayIAMRole as the
		// neutral variant (it has no fields) while still surfacing the diagnostic.
		var c awstypes.CredentialProviderConfiguration
		c.CredentialProviderType = awstypes.CredentialProviderTypeGatewayIamRole
		c.CredentialProvider = nil
		return &c, diags
	}
}

type targetConfigurationModel struct {
	MCP fwtypes.ListNestedObjectValueOf[mcpConfigurationModel] `tfsdk:"mcp"`
}

func (m *targetConfigurationModel) GetConfigurationType(ctx context.Context) string {
	switch mcpData, _ := m.MCP.ToPtr(ctx); {
	case !mcpData.Lambda.IsNull():
		return "lambda"
	case !mcpData.OpenApiSchema.IsNull():
		return "open_api_schema"
	case !mcpData.SmithyModel.IsNull():
		return "smithy_model"
	default:
		return "unknown"
	}
}

var (
	_ flex.Expander  = targetConfigurationModel{}
	_ flex.Flattener = &targetConfigurationModel{}
)

func (m *targetConfigurationModel) Flatten(ctx context.Context, v any) (diags diag.Diagnostics) {
	switch t := v.(type) {
	case awstypes.TargetConfigurationMemberMcp:
		var model mcpConfigurationModel
		d := flex.Flatten(ctx, t.Value, &model)
		diags.Append(d...)
		if diags.HasError() {
			return diags
		}

		m.MCP = fwtypes.NewListNestedObjectValueOfPtrMust(ctx, &model)

		return diags
	default:
		return diags
	}
}

func (m targetConfigurationModel) Expand(ctx context.Context) (result any, diags diag.Diagnostics) {
	switch {
	case !m.MCP.IsNull():
		mcpConfigurationData, d := m.MCP.ToPtr(ctx)
		diags.Append(d...)
		if diags.HasError() {
			return nil, diags
		}

		var r awstypes.TargetConfigurationMemberMcp
		diags.Append(flex.Expand(ctx, mcpConfigurationData, &r.Value)...)
		if diags.HasError() {
			return nil, diags
		}

		return &r, diags
	}

	return nil, diags
}

type mcpConfigurationModel struct {
	Lambda        fwtypes.ListNestedObjectValueOf[mcpLambdaConfigurationModel] `tfsdk:"lambda"`
	SmithyModel   fwtypes.ListNestedObjectValueOf[apiSchemaConfigurationModel] `tfsdk:"smithy_model"`
	OpenApiSchema fwtypes.ListNestedObjectValueOf[apiSchemaConfigurationModel] `tfsdk:"open_api_schema"`
}

var (
	_ flex.Expander  = mcpConfigurationModel{}
	_ flex.Flattener = &mcpConfigurationModel{}
)

func (m *mcpConfigurationModel) Flatten(ctx context.Context, v any) (diags diag.Diagnostics) {
	switch t := v.(type) {
	case awstypes.McpTargetConfigurationMemberLambda:
		var model mcpLambdaConfigurationModel
		d := flex.Flatten(ctx, t.Value, &model)
		diags.Append(d...)
		if diags.HasError() {
			return diags
		}
		m.Lambda = fwtypes.NewListNestedObjectValueOfPtrMust(ctx, &model)
		return diags

	case awstypes.McpTargetConfigurationMemberOpenApiSchema:
		var model apiSchemaConfigurationModel
		d := flex.Flatten(ctx, t.Value, &model)
		diags.Append(d...)
		if diags.HasError() {
			return diags
		}
		m.OpenApiSchema = fwtypes.NewListNestedObjectValueOfPtrMust(ctx, &model)
		return diags

	case awstypes.McpTargetConfigurationMemberSmithyModel:
		var model apiSchemaConfigurationModel
		d := flex.Flatten(ctx, t.Value, &model)
		diags.Append(d...)
		if diags.HasError() {
			return diags
		}
		m.SmithyModel = fwtypes.NewListNestedObjectValueOfPtrMust(ctx, &model)
		return diags

	default:
		return diags
	}
}

func (m mcpConfigurationModel) Expand(ctx context.Context) (result any, diags diag.Diagnostics) {
	switch {
	case !m.Lambda.IsNull():
		lambdaMCPConfigurationData, d := m.Lambda.ToPtr(ctx)
		diags.Append(d...)
		if diags.HasError() {
			return nil, diags
		}

		var r awstypes.McpTargetConfigurationMemberLambda
		diags.Append(flex.Expand(ctx, lambdaMCPConfigurationData, &r.Value)...)
		if diags.HasError() {
			return nil, diags
		}
		return &r, diags

	case !m.OpenApiSchema.IsNull():
		openApiMCPConfigurationData, d := m.OpenApiSchema.ToPtr(ctx)
		diags.Append(d...)
		if diags.HasError() {
			return nil, diags
		}

		var r awstypes.McpTargetConfigurationMemberOpenApiSchema
		diags.Append(flex.Expand(ctx, openApiMCPConfigurationData, &r.Value)...)
		if diags.HasError() {
			return nil, diags
		}
		return &r, diags

	case !m.SmithyModel.IsNull():
		smithyMCPConfigurationData, d := m.SmithyModel.ToPtr(ctx)
		diags.Append(d...)
		if diags.HasError() {
			return nil, diags
		}
		var r awstypes.McpTargetConfigurationMemberSmithyModel
		diags.Append(flex.Expand(ctx, smithyMCPConfigurationData, &r.Value)...)
		if diags.HasError() {
			return nil, diags
		}
		return &r, diags
	}
	return nil, diags
}

type apiKeyCredentialProviderModel struct {
	CredentialLocation      fwtypes.StringEnum[awstypes.ApiKeyCredentialLocation] `tfsdk:"credential_location"`
	CredentialParameterName types.String                                          `tfsdk:"credential_parameter_name"`
	CredentialPrefix        types.String                                          `tfsdk:"credential_prefix"`
	ProviderArn             types.String                                          `tfsdk:"provider_arn"`
}

type oauthCredentialProviderModel struct {
	CustomParameters fwtypes.MapOfString `tfsdk:"custom_parameters"`
	ProviderArn      types.String        `tfsdk:"provider_arn"`
	Scopes           fwtypes.SetOfString `tfsdk:"scopes"`
}

type gatewayIAMRoleProviderModel struct {
	// Empty struct - Gateway IAM Role provider requires no configuration
}

type mcpLambdaConfigurationModel struct {
	LambdaArn  types.String                                     `tfsdk:"lambda_arn"`
	ToolSchema fwtypes.ListNestedObjectValueOf[toolSchemaModel] `tfsdk:"tool_schema"`
}

type toolSchemaModel struct {
	InlinePayload fwtypes.ListNestedObjectValueOf[toolDefinitionModel]  `tfsdk:"inline_payload"`
	S3            fwtypes.ListNestedObjectValueOf[s3ConfigurationModel] `tfsdk:"s3"`
}

var (
	_ flex.Expander  = toolSchemaModel{}
	_ flex.Flattener = &toolSchemaModel{}
)

func (m *toolSchemaModel) Flatten(ctx context.Context, v any) (diags diag.Diagnostics) {
	switch t := v.(type) {
	case awstypes.ToolSchemaMemberInlinePayload:
		var toolDefModels []*toolDefinitionModel
		for _, toolDef := range t.Value {
			var model toolDefinitionModel
			d := flex.Flatten(ctx, toolDef, &model)
			diags.Append(d...)
			if diags.HasError() {
				return diags
			}
			toolDefModels = append(toolDefModels, &model)
		}
		m.InlinePayload = fwtypes.NewListNestedObjectValueOfSliceMust(ctx, toolDefModels)
		return diags

	case awstypes.ToolSchemaMemberS3:
		var model s3ConfigurationModel
		d := flex.Flatten(ctx, t.Value, &model)
		diags.Append(d...)
		if diags.HasError() {
			return diags
		}
		m.S3 = fwtypes.NewListNestedObjectValueOfPtrMust(ctx, &model)
		return diags

	default:
		return diags
	}
}

func (m toolSchemaModel) Expand(ctx context.Context) (result any, diags diag.Diagnostics) {
	switch {
	case !m.InlinePayload.IsNull():
		inlinePayloadToolSchemaData, d := m.InlinePayload.ToSlice(ctx)
		diags.Append(d...)
		if diags.HasError() {
			return nil, diags
		}

		var toolDefs []awstypes.ToolDefinition
		for _, toolDefModel := range inlinePayloadToolSchemaData {
			var toolDef awstypes.ToolDefinition
			diags.Append(flex.Expand(ctx, toolDefModel, &toolDef)...)
			if diags.HasError() {
				return nil, diags
			}
			toolDefs = append(toolDefs, toolDef)
		}

		var r awstypes.ToolSchemaMemberInlinePayload
		r.Value = toolDefs
		return &r, diags

	case !m.S3.IsNull():
		s3ToolSchemaData, d := m.S3.ToPtr(ctx)
		diags.Append(d...)
		if diags.HasError() {
			return nil, diags
		}

		var r awstypes.ToolSchemaMemberS3
		diags.Append(flex.Expand(ctx, s3ToolSchemaData, &r.Value)...)
		if diags.HasError() {
			return nil, diags
		}
		return &r, diags
	}
	return nil, diags
}

type toolDefinitionModel struct {
	Description  types.String                                           `tfsdk:"description"`
	Name         types.String                                           `tfsdk:"name"`
	InputSchema  fwtypes.ListNestedObjectValueOf[schemaDefinitionModel] `tfsdk:"input_schema"`
	OutputSchema fwtypes.ListNestedObjectValueOf[schemaDefinitionModel] `tfsdk:"output_schema"`
}

type schemaDefinitionCoreModel struct {
	Description types.String                                      `tfsdk:"description"`
	Type        fwtypes.StringEnum[awstypes.SchemaType]           `tfsdk:"type"`
	Items       fwtypes.ListNestedObjectValueOf[schemaItemsModel] `tfsdk:"items"`
}

type schemaDefinitionModel struct {
	schemaDefinitionCoreModel
	Properties fwtypes.SetNestedObjectValueOf[schemaPropertyModel] `tfsdk:"property"`
}

var (
	_ flex.Expander  = schemaDefinitionModel{}
	_ flex.Flattener = &schemaDefinitionModel{}
)

func (m *schemaDefinitionModel) Flatten(ctx context.Context, v any) (diags diag.Diagnostics) {
	// Ensure Properties is a typed Null when absent to avoid zero-value Set panics during state encoding
	m.Properties = fwtypes.NewSetNestedObjectValueOfNull[schemaPropertyModel](ctx)
	switch t := v.(type) {
	case awstypes.SchemaDefinition:
		diags.Append(flex.Flatten(ctx, v, &m.schemaDefinitionCoreModel)...)
		if diags.HasError() {
			return diags
		}

		// Normalize: when API omits Items, return an empty list (not null)
		if t.Items == nil {
			m.Items = fwtypes.NewListNestedObjectValueOfSliceMust(ctx, []*schemaItemsModel{})
		}

		if t.Properties != nil {
			properties, d := flattenTargetSchemaProperties(ctx, t.Properties, t.Required)
			diags.Append(d...)
			if diags.HasError() {
				return diags
			}
			m.Properties = properties
		}
	}

	return diags
}

func (m schemaDefinitionModel) Expand(ctx context.Context) (result any, diags diag.Diagnostics) {
	schemaDefinitionData := &awstypes.SchemaDefinition{}

	diags.Append(flex.Expand(ctx, m.schemaDefinitionCoreModel, schemaDefinitionData)...)
	if diags.HasError() {
		return nil, diags
	}

	if !m.Properties.IsNull() {
		properties, requiredProps, d := expandTargetSchemaProperties(ctx, m.Properties)
		diags.Append(d...)
		if diags.HasError() {
			return nil, diags
		}
		schemaDefinitionData.Properties = properties
		schemaDefinitionData.Required = requiredProps
	}

	return schemaDefinitionData, diags
}

type schemaItemsCoreModel struct {
	Description types.String                                          `tfsdk:"description"`
	Type        fwtypes.StringEnum[awstypes.SchemaType]               `tfsdk:"type"`
	Items       fwtypes.ListNestedObjectValueOf[schemaItemsLeafModel] `tfsdk:"items"`
}

type schemaItemsModel struct {
	schemaItemsCoreModel
	Properties fwtypes.SetNestedObjectValueOf[schemaPropertyLeafModel] `tfsdk:"property"`
}

var (
	_ flex.Expander  = schemaItemsModel{}
	_ flex.Flattener = &schemaItemsModel{}
)

func (m *schemaItemsModel) Flatten(ctx context.Context, v any) (diags diag.Diagnostics) {
	// Ensure Properties is a typed Null when absent to avoid zero-value Set panics during state encoding
	m.Properties = fwtypes.NewSetNestedObjectValueOfNull[schemaPropertyLeafModel](ctx)
	switch t := v.(type) {
	case awstypes.SchemaDefinition:
		diags.Append(flex.Flatten(ctx, v, &m.schemaItemsCoreModel)...)
		if diags.HasError() {
			return diags
		}

		// Normalize: when API omits Items, return an empty list (not null)
		if t.Items == nil {
			m.Items = fwtypes.NewListNestedObjectValueOfSliceMust(ctx, []*schemaItemsLeafModel{})
		}

		if t.Properties != nil {
			properties, d := flattenTargetSchemaLeafProperties(ctx, t.Properties, t.Required)
			diags.Append(d...)
			if diags.HasError() {
				return diags
			}
			m.Properties = properties
		}
	}

	return diags
}

func (m schemaItemsModel) Expand(ctx context.Context) (result any, diags diag.Diagnostics) {
	schemaDefinitionData := &awstypes.SchemaDefinition{}

	diags.Append(flex.Expand(ctx, m.schemaItemsCoreModel, schemaDefinitionData)...)
	if diags.HasError() {
		return nil, diags
	}

	if !m.Properties.IsNull() {
		properties, requiredProps, d := expandTargetSchemaLeafProperties(ctx, m.Properties)
		diags.Append(d...)
		if diags.HasError() {
			return nil, diags
		}
		schemaDefinitionData.Properties = properties
		schemaDefinitionData.Required = requiredProps
	}

	return schemaDefinitionData, diags
}

type schemaPropertyCoreModel struct {
	Name        types.String                                      `tfsdk:"name"`
	Description types.String                                      `tfsdk:"description"`
	Type        fwtypes.StringEnum[awstypes.SchemaType]           `tfsdk:"type"`
	Items       fwtypes.ListNestedObjectValueOf[schemaItemsModel] `tfsdk:"items"`
}

type schemaPropertyModel struct {
	schemaPropertyCoreModel
	Required   types.Bool                                              `tfsdk:"required"`
	Properties fwtypes.SetNestedObjectValueOf[schemaPropertyLeafModel] `tfsdk:"property"`
}

var (
	_ flex.Expander  = schemaPropertyModel{}
	_ flex.Flattener = &schemaPropertyModel{}
)

func (m *schemaPropertyModel) Flatten(ctx context.Context, v any) (diags diag.Diagnostics) {
	// Ensure Properties is a typed Null when absent to avoid zero-value Set panics during state encoding
	m.Properties = fwtypes.NewSetNestedObjectValueOfNull[schemaPropertyLeafModel](ctx)
	switch t := v.(type) {
	case awstypes.SchemaDefinition:
		diags.Append(flex.Flatten(ctx, v, &m.schemaPropertyCoreModel)...)
		if diags.HasError() {
			return diags
		}

		// Normalize: when API omits Items, return an empty list (not null)
		if t.Items == nil {
			m.Items = fwtypes.NewListNestedObjectValueOfSliceMust(ctx, []*schemaItemsModel{})
		}

		if t.Properties != nil {
			properties, d := flattenTargetSchemaLeafProperties(ctx, t.Properties, t.Required)
			diags.Append(d...)
			if diags.HasError() {
				return diags
			}
			m.Properties = properties
		}
	}
	return diags
}

func (m schemaPropertyModel) Expand(ctx context.Context) (result any, diags diag.Diagnostics) {
	var schemaDefinitionLeafData = awstypes.SchemaDefinition{}
	diags.Append(flex.Expand(ctx, m.schemaPropertyCoreModel, &schemaDefinitionLeafData)...)
	if diags.HasError() {
		return nil, diags
	}

	if !m.Properties.IsNull() {
		properties, requiredProps, d := expandTargetSchemaLeafProperties(ctx, m.Properties)
		diags.Append(d...)
		if diags.HasError() {
			return nil, diags
		}
		schemaDefinitionLeafData.Properties = properties
		schemaDefinitionLeafData.Required = requiredProps
	}

	return schemaDefinitionLeafData, diags
}

type schemaItemsLeafCoreModel struct {
	Description types.String                            `tfsdk:"description"`
	Type        fwtypes.StringEnum[awstypes.SchemaType] `tfsdk:"type"`
}

type schemaItemsLeafModel struct {
	schemaItemsLeafCoreModel
	// JSON serialized schema for deeper nesting
	ItemsJSON      types.String `tfsdk:"items_json"`
	PropertiesJSON types.String `tfsdk:"properties_json"`
}

var (
	_ flex.Expander  = schemaItemsLeafModel{}
	_ flex.Flattener = &schemaItemsLeafModel{}
)

func (m *schemaItemsLeafModel) Flatten(ctx context.Context, v any) (diags diag.Diagnostics) {
	switch t := v.(type) {
	case awstypes.SchemaDefinition:
		diags.Append(flex.Flatten(ctx, v, &m.schemaItemsLeafCoreModel)...)
		if diags.HasError() {
			return diags
		}
		// Populate ItemsJSON
		if t.Items != nil {
			jsonItems := convertToJSONSchemaDefinition(t.Items)
			b, err := json.Marshal(jsonItems)
			if err != nil {
				diags.AddWarning("Failed to marshal items for items_json", err.Error())
				m.ItemsJSON = types.StringNull()
			} else {
				m.ItemsJSON = types.StringValue(string(b))
			}
		} else {
			m.ItemsJSON = types.StringNull()
		}
		// Populate PropertiesJSON
		if t.Properties != nil || len(t.Required) > 0 {
			propObj := awstypes.SchemaDefinition{
				Properties: t.Properties,
				Required:   t.Required,
			}
			jsonProps := convertToJSONSchemaDefinition(&propObj)
			b, err := json.Marshal(jsonProps)
			if err != nil {
				diags.AddWarning("Failed to marshal properties for properties_json", err.Error())
				m.PropertiesJSON = types.StringNull()
			} else {
				m.PropertiesJSON = types.StringValue(string(b))
			}
		} else {
			m.PropertiesJSON = types.StringNull()
		}
	}
	return diags
}

func (m schemaItemsLeafModel) Expand(ctx context.Context) (result any, diags diag.Diagnostics) {
	var sd awstypes.SchemaDefinition
	// Expand core (type/description)
	diags.Append(flex.Expand(ctx, m.schemaItemsLeafCoreModel, &sd)...)
	if diags.HasError() {
		return nil, diags
	}

	if isNonEmpty(m.ItemsJSON) {
		jsd, d := parseJSONSchemaDefinition(m.ItemsJSON.ValueString())
		diags.Append(d...)
		if diags.HasError() {
			return nil, diags
		}
		sd.Items = jsd
	}
	if isNonEmpty(m.PropertiesJSON) {
		jsd, d := parseJSONSchemaDefinition(m.PropertiesJSON.ValueString())
		diags.Append(d...)
		if diags.HasError() {
			return nil, diags
		}
		sd.Properties = jsd.Properties
		sd.Required = jsd.Required
	}

	return &sd, diags
}

type schemaPropertyLeafCoreModel struct {
	Name        types.String                            `tfsdk:"name"`
	Description types.String                            `tfsdk:"description"`
	Type        fwtypes.StringEnum[awstypes.SchemaType] `tfsdk:"type"`
}

type schemaPropertyLeafModel struct {
	schemaPropertyLeafCoreModel
	Required types.Bool `tfsdk:"required"`
	// JSON serialized schema for deeper nesting
	ItemsJSON      types.String `tfsdk:"items_json"`
	PropertiesJSON types.String `tfsdk:"properties_json"`
}

var (
	_ flex.Expander  = schemaPropertyLeafModel{}
	_ flex.Flattener = &schemaPropertyLeafModel{}
)

func (m *schemaPropertyLeafModel) Flatten(ctx context.Context, v any) (diags diag.Diagnostics) {
	switch t := v.(type) {
	case awstypes.SchemaDefinition:
		diags.Append(flex.Flatten(ctx, v, &m.schemaPropertyLeafCoreModel)...)
		if diags.HasError() {
			return diags
		}
		// Populate ItemsJSON
		if t.Items != nil {
			jsonItems := convertToJSONSchemaDefinition(t.Items)
			b, err := json.Marshal(jsonItems)
			if err != nil {
				diags.AddWarning("Failed to marshal items for items_json", err.Error())
				m.ItemsJSON = types.StringNull()
			} else {
				m.ItemsJSON = types.StringValue(string(b))
			}
		} else {
			m.ItemsJSON = types.StringNull()
		}
		// Populate PropertiesJSON
		if t.Properties != nil || len(t.Required) > 0 {
			propObj := awstypes.SchemaDefinition{
				Properties: t.Properties,
				Required:   t.Required,
			}
			jsonProps := convertToJSONSchemaDefinition(&propObj)
			b, err := json.Marshal(jsonProps)
			if err != nil {
				diags.AddWarning("Failed to marshal properties for properties_json", err.Error())
				m.PropertiesJSON = types.StringNull()
			} else {
				m.PropertiesJSON = types.StringValue(string(b))
			}
		} else {
			m.PropertiesJSON = types.StringNull()
		}
	}
	return diags
}

func (m schemaPropertyLeafModel) Expand(ctx context.Context) (result any, diags diag.Diagnostics) {
	var schemaDefinitionData = awstypes.SchemaDefinition{}

	diags.Append(flex.Expand(ctx, m.schemaPropertyLeafCoreModel, &schemaDefinitionData)...)
	if diags.HasError() {
		return nil, diags
	}

	if isNonEmpty(m.ItemsJSON) {
		jsd, d := parseJSONSchemaDefinition(m.ItemsJSON.ValueString())
		diags.Append(d...)
		if diags.HasError() {
			return nil, diags
		}
		schemaDefinitionData.Items = jsd
	}
	if isNonEmpty(m.PropertiesJSON) {
		jsd, d := parseJSONSchemaDefinition(m.PropertiesJSON.ValueString())
		diags.Append(d...)
		if diags.HasError() {
			return nil, diags
		}
		schemaDefinitionData.Properties = jsd.Properties
		schemaDefinitionData.Required = jsd.Required
	}
	return schemaDefinitionData, diags
}

type s3ConfigurationModel struct {
	BucketOwnerAccountId types.String `tfsdk:"bucket_owner_account_id"`
	Uri                  types.String `tfsdk:"uri"`
}

type apiSchemaConfigurationModel struct {
	InlinePayload fwtypes.ListNestedObjectValueOf[inlinePayloadModel]   `tfsdk:"inline_payload"`
	S3            fwtypes.ListNestedObjectValueOf[s3ConfigurationModel] `tfsdk:"s3"`
}

var (
	_ flex.Expander  = apiSchemaConfigurationModel{}
	_ flex.Flattener = &apiSchemaConfigurationModel{}
)

func (m *apiSchemaConfigurationModel) Flatten(ctx context.Context, v any) (diags diag.Diagnostics) {
	switch t := v.(type) {
	case awstypes.ApiSchemaConfigurationMemberInlinePayload:
		var model inlinePayloadModel
		model.Payload = types.StringValue(t.Value)
		m.InlinePayload = fwtypes.NewListNestedObjectValueOfPtrMust(ctx, &model)
		return diags

	case awstypes.ApiSchemaConfigurationMemberS3:
		var model s3ConfigurationModel
		d := flex.Flatten(ctx, t.Value, &model)
		diags.Append(d...)
		if diags.HasError() {
			return diags
		}
		m.S3 = fwtypes.NewListNestedObjectValueOfPtrMust(ctx, &model)
		return diags

	default:
		return diags
	}
}

func (m apiSchemaConfigurationModel) Expand(ctx context.Context) (result any, diags diag.Diagnostics) {
	switch {
	case !m.InlinePayload.IsNull():
		inlinePayloadApiSchemaConfigurationData, d := m.InlinePayload.ToPtr(ctx)
		diags.Append(d...)
		if diags.HasError() {
			return nil, diags
		}

		var r awstypes.ApiSchemaConfigurationMemberInlinePayload
		r.Value = inlinePayloadApiSchemaConfigurationData.Payload.ValueString()
		return &r, diags

	case !m.S3.IsNull():
		s3ApiSchemaConfigurationData, d := m.S3.ToPtr(ctx)
		diags.Append(d...)
		if diags.HasError() {
			return nil, diags
		}

		var r awstypes.ApiSchemaConfigurationMemberS3
		diags.Append(flex.Expand(ctx, s3ApiSchemaConfigurationData, &r.Value)...)
		if diags.HasError() {
			return nil, diags
		}
		return &r, diags
	}
	return nil, diags
}

type inlinePayloadModel struct {
	Payload types.String `tfsdk:"payload"`
}

// Helper functions for PropertiesJSON map conversion
func flattenTargetSchemaProperties(
	ctx context.Context,
	properties map[string]awstypes.SchemaDefinition,
	required []string,
) (fwtypes.SetNestedObjectValueOf[schemaPropertyModel], diag.Diagnostics) {
	var diags diag.Diagnostics
	if len(properties) == 0 {
		return fwtypes.NewSetNestedObjectValueOfNull[schemaPropertyModel](ctx), diags
	}

	requiredSet := map[string]bool{}
	for _, n := range required {
		requiredSet[n] = true
	}

	var propertyModels []*schemaPropertyModel
	for name, schemaDefn := range properties {
		pm := &schemaPropertyModel{}
		d := pm.Flatten(ctx, schemaDefn)
		diags.Append(d...)
		if diags.HasError() {
			return fwtypes.NewSetNestedObjectValueOfNull[schemaPropertyModel](ctx), diags
		}

		pm.Name = types.StringValue(name)
		pm.Required = types.BoolValue(requiredSet[name])

		propertyModels = append(propertyModels, pm)
	}

	return fwtypes.NewSetNestedObjectValueOfSliceMust(ctx, propertyModels), diags
}

func expandTargetSchemaProperties(ctx context.Context, properties fwtypes.SetNestedObjectValueOf[schemaPropertyModel]) (map[string]awstypes.SchemaDefinition, []string, diag.Diagnostics) {
	var diags diag.Diagnostics
	result := make(map[string]awstypes.SchemaDefinition)
	var requiredProps []string

	propertySlice, d := properties.ToSlice(ctx)
	diags.Append(d...)
	if diags.HasError() {
		return nil, nil, diags
	}

	for _, propertyModel := range propertySlice {
		expandedValue, d := propertyModel.Expand(ctx)
		diags.Append(d...)
		if diags.HasError() {
			return nil, nil, diags
		}

		if schemaDefn, ok := expandedValue.(awstypes.SchemaDefinition); ok {
			name := propertyModel.Name.ValueString()
			result[name] = schemaDefn

			// Since we always set required to explicit boolean, we can check it directly
			if propertyModel.Required.ValueBool() {
				requiredProps = append(requiredProps, name)
			}
		}
	}
	return result, requiredProps, diags
}

// Helper functions for Leaf PropertiesJSON map conversion
func flattenTargetSchemaLeafProperties(ctx context.Context, properties map[string]awstypes.SchemaDefinition, requiredProps []string) (fwtypes.SetNestedObjectValueOf[schemaPropertyLeafModel], diag.Diagnostics) {
	var diags diag.Diagnostics
	requiredSet := make(map[string]bool)
	for _, prop := range requiredProps {
		requiredSet[prop] = true
	}

	var propertyModels []*schemaPropertyLeafModel
	for name, schemaDefn := range properties {
		pm := &schemaPropertyLeafModel{}
		d := pm.Flatten(ctx, schemaDefn)
		diags.Append(d...)
		if diags.HasError() {
			return fwtypes.NewSetNestedObjectValueOfNull[schemaPropertyLeafModel](ctx), diags
		}
		pm.Name = types.StringValue(name)
		pm.Required = types.BoolValue(requiredSet[name])
		propertyModels = append(propertyModels, pm)
	}
	return fwtypes.NewSetNestedObjectValueOfSliceMust(ctx, propertyModels), diags
}

func expandTargetSchemaLeafProperties(ctx context.Context, properties fwtypes.SetNestedObjectValueOf[schemaPropertyLeafModel]) (map[string]awstypes.SchemaDefinition, []string, diag.Diagnostics) {
	var diags diag.Diagnostics
	result := make(map[string]awstypes.SchemaDefinition)
	var requiredProps []string

	propertySlice, d := properties.ToSlice(ctx)
	diags.Append(d...)
	if diags.HasError() {
		return nil, nil, diags
	}

	for _, propertyModel := range propertySlice {
		expandedValue, d := propertyModel.Expand(ctx)
		diags.Append(d...)
		if diags.HasError() {
			return nil, nil, diags
		}

		if schemaDefn, ok := expandedValue.(awstypes.SchemaDefinition); ok {
			name := propertyModel.Name.ValueString()
			result[name] = schemaDefn

			if propertyModel.Required.ValueBool() {
				requiredProps = append(requiredProps, name)
			}
		}
	}

	return result, requiredProps, diags
}

func parseJSONSchemaDefinition(s string) (*awstypes.SchemaDefinition, diag.Diagnostics) {
	var diags diag.Diagnostics
	s = strings.TrimSpace(s)
	if s == "" {
		diags.AddError("Invalid JSON", "JSON schema must be a non-empty string")
		return nil, diags
	}
	var sd awstypes.SchemaDefinition
	if err := json.Unmarshal([]byte(s), &sd); err != nil {
		diags.AddError("Invalid JSON", err.Error())
		return nil, diags
	}
	return &sd, diags
}

func isNonEmpty(s types.String) bool {
	return !s.IsNull() && !s.IsUnknown() && strings.TrimSpace(s.ValueString()) != ""
}

// jsonSchemaDefinition is a helper struct for JSON serialization with lowercase field names
type jsonSchemaDefinition struct {
	Type        string                           `json:"type,omitempty"`
	Description *string                          `json:"description,omitempty"`
	Items       *jsonSchemaDefinition            `json:"items,omitempty"`
	Properties  map[string]*jsonSchemaDefinition `json:"properties,omitempty"`
	Required    []string                         `json:"required,omitempty"`
}

// convertToJSONSchemaDefinition converts AWS SDK SchemaDefinition to our JSON-friendly version
func convertToJSONSchemaDefinition(sd *awstypes.SchemaDefinition) *jsonSchemaDefinition {
	if sd == nil {
		return nil
	}

	jsd := &jsonSchemaDefinition{
		Type: string(sd.Type), // Convert SchemaType enum to string
	}

	// Only set non-nil values to avoid null fields in JSON
	if sd.Description != nil && aws.ToString(sd.Description) != "" {
		jsd.Description = sd.Description
	}
	if len(sd.Required) > 0 {
		jsd.Required = sd.Required
	}

	if sd.Items != nil {
		jsd.Items = convertToJSONSchemaDefinition(sd.Items)
	}

	if sd.Properties != nil {
		jsd.Properties = make(map[string]*jsonSchemaDefinition)
		for k, v := range sd.Properties {
			if converted := convertToJSONSchemaDefinition(&v); converted != nil {
				jsd.Properties[k] = converted
			}
		}
		// If no properties were added, don't include the properties field
		if len(jsd.Properties) == 0 {
			jsd.Properties = nil
		}
	}

	return jsd
}

// func sweepGatewayTargets(ctx context.Context, client *conns.AWSClient) ([]sweep.Sweepable, error) {
// 	input := bedrockagentcorecontrol.ListGatewayTargetsInput{}
// 	conn := client.BedrockAgentCoreClient(ctx)
// 	var sweepResources []sweep.Sweepable

// 	pages := bedrockagentcorecontrol.NewListGatewayTargetsPaginator(conn, &input)
// 	for pages.HasMorePages() {
// 		page, err := pages.NextPage(ctx)
// 		if err != nil {
// 			return nil, smarterr.NewError(err)
// 		}

// 		for _, v := range page.ItemsJSON {
// 			sweepResources = append(sweepResources, sweepfw.NewSweepResource(newResourceGatewayTarget, client,
// 				sweepfw.NewAttribute(names.AttrID, aws.ToString(v.TargetId)), sweepfw.NewAttribute(names.AttrID, aws.ToString(v.GatewayIdentifier))),
// 			)
// 		}
// 	}

// 	return sweepResources, nil
// }
