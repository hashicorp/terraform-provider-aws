// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

// DONOTCOPY: Copying old resources spreads bad habits. Use skaff instead.

package bedrockagentcore

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/YakDriver/regexache"
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
	sdkid "github.com/hashicorp/terraform-plugin-sdk/v2/helper/id"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/fwdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	fwflex "github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	tfjson "github.com/hashicorp/terraform-provider-aws/internal/json"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/smerr"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkResource("aws_bedrockagentcore_gateway_target", name="Gateway Target")
func newGatewayTargetResource(_ context.Context) (resource.ResourceWithConfigure, error) {
	r := &gatewayTargetResource{}

	r.SetDefaultCreateTimeout(30 * time.Minute)
	r.SetDefaultUpdateTimeout(30 * time.Minute)
	r.SetDefaultDeleteTimeout(30 * time.Minute)

	return r, nil
}

type gatewayTargetResource struct {
	framework.ResourceWithModel[gatewayTargetResourceModel]
	framework.WithTimeouts
}

func jsonAttribute(conflictWith string) schema.StringAttribute {
	return schema.StringAttribute{
		Optional:      true,
		PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
		Validators: []validator.String{
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

func (r *gatewayTargetResource) Schema(ctx context.Context, request resource.SchemaRequest, response *resource.SchemaResponse) {
	response.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			names.AttrDescription: schema.StringAttribute{
				Optional: true,
				Validators: []validator.String{
					stringvalidator.LengthBetween(1, 200),
				},
			},
			"gateway_identifier": schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			names.AttrName: schema.StringAttribute{
				Required: true,
				Validators: []validator.String{
					stringvalidator.RegexMatches(regexache.MustCompile(`^([0-9a-zA-Z][-]?){1,100}$`), ""),
				},
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"target_id": framework.IDAttribute(),
		},
		Blocks: map[string]schema.Block{
			"credential_provider_configuration": schema.ListNestedBlock{
				CustomType: fwtypes.NewListNestedObjectTypeOf[credentialProviderConfigurationModel](ctx),
				Validators: []validator.List{
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
										CustomType: fwtypes.ARNType,
										Required:   true,
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
									"mcp_server": schema.ListNestedBlock{
										CustomType: fwtypes.NewListNestedObjectTypeOf[mcpServerConfigurationModel](ctx),
										Validators: []validator.List{
											listvalidator.SizeAtMost(1),
										},
										NestedObject: schema.NestedBlockObject{
											Attributes: map[string]schema.Attribute{
												names.AttrEndpoint: schema.StringAttribute{
													Required: true,
													Validators: []validator.String{
														stringvalidator.RegexMatches(
															regexache.MustCompile(`https://.*`),
															"Must start with https://",
														),
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

func (r *gatewayTargetResource) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	var data gatewayTargetResourceModel
	smerr.AddEnrich(ctx, &response.Diagnostics, request.Plan.Get(ctx, &data))
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().BedrockAgentCoreClient(ctx)

	gatewayIdentifier := fwflex.StringValueFromFramework(ctx, data.GatewayIdentifier)
	var input bedrockagentcorecontrol.CreateGatewayTargetInput
	smerr.AddEnrich(ctx, &response.Diagnostics, fwflex.Expand(ctx, data, &input))
	if response.Diagnostics.HasError() {
		return
	}

	// Additional fields.
	input.ClientToken = aws.String(sdkid.UniqueId())

	out, err := conn.CreateGatewayTarget(ctx, &input)
	if err != nil {
		smerr.AddError(ctx, &response.Diagnostics, err, smerr.ID, data.Name.String())
		return
	}

	targetID := aws.ToString(out.TargetId)
	data.TargetID = fwflex.StringValueToFramework(ctx, targetID)

	if _, err := waitGatewayTargetCreated(ctx, conn, gatewayIdentifier, targetID, r.CreateTimeout(ctx, data.Timeouts)); err != nil {
		smerr.AddError(ctx, &response.Diagnostics, err, smerr.ID, targetID)
		return
	}

	smerr.AddEnrich(ctx, &response.Diagnostics, response.State.Set(ctx, data))
}

func (r *gatewayTargetResource) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	var data gatewayTargetResourceModel
	smerr.AddEnrich(ctx, &response.Diagnostics, request.State.Get(ctx, &data))
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().BedrockAgentCoreClient(ctx)

	gatewayIdentifier, targetID := fwflex.StringValueFromFramework(ctx, data.GatewayIdentifier), fwflex.StringValueFromFramework(ctx, data.TargetID)
	out, err := findGatewayTargetByTwoPartKey(ctx, conn, gatewayIdentifier, targetID)
	if retry.NotFound(err) {
		smerr.AddOne(ctx, &response.Diagnostics, fwdiag.NewResourceNotFoundWarningDiagnostic(err))
		response.State.RemoveResource(ctx)
		return
	}
	if err != nil {
		smerr.AddError(ctx, &response.Diagnostics, err, smerr.ID, targetID)
		return
	}

	smerr.AddEnrich(ctx, &response.Diagnostics, fwflex.Flatten(ctx, out, &data, fwflex.WithIgnoredFieldNames([]string{"GatewayArn"})))
	if response.Diagnostics.HasError() {
		return
	}

	smerr.AddEnrich(ctx, &response.Diagnostics, response.State.Set(ctx, &data))
}

func (r *gatewayTargetResource) Update(ctx context.Context, request resource.UpdateRequest, response *resource.UpdateResponse) {
	var new, old gatewayTargetResourceModel
	smerr.AddEnrich(ctx, &response.Diagnostics, request.Plan.Get(ctx, &new))
	smerr.AddEnrich(ctx, &response.Diagnostics, request.State.Get(ctx, &old))
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().BedrockAgentCoreClient(ctx)

	diff, d := fwflex.Diff(ctx, new, old)
	smerr.AddEnrich(ctx, &response.Diagnostics, d)
	if response.Diagnostics.HasError() {
		return
	}

	if diff.HasChanges() {
		gatewayIdentifier, targetID := fwflex.StringValueFromFramework(ctx, new.GatewayIdentifier), fwflex.StringValueFromFramework(ctx, new.TargetID)
		var input bedrockagentcorecontrol.UpdateGatewayTargetInput
		smerr.AddEnrich(ctx, &response.Diagnostics, fwflex.Expand(ctx, new, &input))
		if response.Diagnostics.HasError() {
			return
		}

		_, err := conn.UpdateGatewayTarget(ctx, &input)
		if err != nil {
			smerr.AddError(ctx, &response.Diagnostics, err, smerr.ID, targetID)
			return
		}

		_, err = waitGatewayTargetUpdated(ctx, conn, gatewayIdentifier, targetID, r.UpdateTimeout(ctx, new.Timeouts))
		if err != nil {
			smerr.AddError(ctx, &response.Diagnostics, err, smerr.ID, targetID)
			return
		}
	}

	smerr.AddEnrich(ctx, &response.Diagnostics, response.State.Set(ctx, &new))
}

func (r *gatewayTargetResource) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	var data gatewayTargetResourceModel
	smerr.AddEnrich(ctx, &response.Diagnostics, request.State.Get(ctx, &data))
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().BedrockAgentCoreClient(ctx)

	gatewayIdentifier, targetID := fwflex.StringValueFromFramework(ctx, data.GatewayIdentifier), fwflex.StringValueFromFramework(ctx, data.TargetID)
	input := bedrockagentcorecontrol.DeleteGatewayTargetInput{
		GatewayIdentifier: aws.String(gatewayIdentifier),
		TargetId:          aws.String(targetID),
	}
	_, err := conn.DeleteGatewayTarget(ctx, &input)
	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return
	}
	if err != nil {
		smerr.AddError(ctx, &response.Diagnostics, err, smerr.ID, targetID)
		return
	}

	if _, err := waitGatewayTargetDeleted(ctx, conn, gatewayIdentifier, targetID, r.DeleteTimeout(ctx, data.Timeouts)); err != nil {
		smerr.AddError(ctx, &response.Diagnostics, err, smerr.ID, targetID)
		return
	}
}

func (r *gatewayTargetResource) ImportState(ctx context.Context, request resource.ImportStateRequest, response *resource.ImportStateResponse) {
	parts := strings.Split(request.ID, ",")

	if len(parts) != 2 {
		smerr.AddError(ctx, &response.Diagnostics, fmt.Errorf(`Unexpected format for import ID (%s), use: "GatewayIdentifier,TargetId"`, request.ID))
		return
	}

	smerr.AddEnrich(ctx, &response.Diagnostics, response.State.SetAttribute(ctx, path.Root("gateway_identifier"), parts[0]))
	smerr.AddEnrich(ctx, &response.Diagnostics, response.State.SetAttribute(ctx, path.Root("target_id"), parts[1]))
}

func (r gatewayTargetResource) ModifyPlan(ctx context.Context, request resource.ModifyPlanRequest, response *resource.ModifyPlanResponse) {
	if request.State.Raw.IsNull() || request.Plan.Raw.IsNull() {
		return
	}

	// Force replacement if target configuration changes between lambda, smithy_model, and open_api_schema
	var plan, state gatewayTargetResourceModel
	smerr.AddEnrich(ctx, &response.Diagnostics, request.Plan.Get(ctx, &plan))
	smerr.AddEnrich(ctx, &response.Diagnostics, request.State.Get(ctx, &state))
	if response.Diagnostics.HasError() {
		return
	}

	planTargetData, diags := plan.TargetConfiguration.ToPtr(ctx)
	smerr.AddEnrich(ctx, &response.Diagnostics, diags)
	stateTargetData, diags := state.TargetConfiguration.ToPtr(ctx)
	smerr.AddEnrich(ctx, &response.Diagnostics, diags)
	if response.Diagnostics.HasError() {
		return
	}

	planTargetType := planTargetData.GetConfigurationType(ctx)
	stateTargetType := stateTargetData.GetConfigurationType(ctx)

	if planTargetType != stateTargetType {
		response.RequiresReplace = append(response.RequiresReplace, path.Root("target_configuration"))
	}
}

func waitGatewayTargetCreated(ctx context.Context, conn *bedrockagentcorecontrol.Client, gatewayIdentifier, targetID string, timeout time.Duration) (*bedrockagentcorecontrol.GetGatewayTargetOutput, error) {
	stateConf := &retry.StateChangeConf{
		Pending:                   enum.Slice(awstypes.TargetStatusCreating),
		Target:                    enum.Slice(awstypes.TargetStatusReady),
		Refresh:                   statusGatewayTarget(conn, gatewayIdentifier, targetID),
		Timeout:                   timeout,
		ContinuousTargetOccurence: 2,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)
	if out, ok := outputRaw.(*bedrockagentcorecontrol.GetGatewayTargetOutput); ok {
		retry.SetLastError(err, errors.New(strings.Join(out.StatusReasons, "; ")))
		return out, smarterr.NewError(err)
	}

	return nil, smarterr.NewError(err)
}

func waitGatewayTargetUpdated(ctx context.Context, conn *bedrockagentcorecontrol.Client, gatewayIdentifier, targetID string, timeout time.Duration) (*bedrockagentcorecontrol.GetGatewayTargetOutput, error) {
	stateConf := &retry.StateChangeConf{
		Pending:                   enum.Slice(awstypes.TargetStatusUpdating),
		Target:                    enum.Slice(awstypes.TargetStatusReady),
		Refresh:                   statusGatewayTarget(conn, gatewayIdentifier, targetID),
		Timeout:                   timeout,
		ContinuousTargetOccurence: 2,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)
	if out, ok := outputRaw.(*bedrockagentcorecontrol.GetGatewayTargetOutput); ok {
		retry.SetLastError(err, errors.New(strings.Join(out.StatusReasons, "; ")))
		return out, smarterr.NewError(err)
	}

	return nil, smarterr.NewError(err)
}

func waitGatewayTargetDeleted(ctx context.Context, conn *bedrockagentcorecontrol.Client, gatewayIdentifier, targetID string, timeout time.Duration) (*bedrockagentcorecontrol.GetGatewayTargetOutput, error) {
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(awstypes.TargetStatusDeleting, awstypes.TargetStatusReady),
		Target:  []string{},
		Refresh: statusGatewayTarget(conn, gatewayIdentifier, targetID),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)
	if out, ok := outputRaw.(*bedrockagentcorecontrol.GetGatewayTargetOutput); ok {
		retry.SetLastError(err, errors.New(strings.Join(out.StatusReasons, "; ")))
		return out, smarterr.NewError(err)
	}

	return nil, smarterr.NewError(err)
}

func statusGatewayTarget(conn *bedrockagentcorecontrol.Client, gatewayIdentifier, targetID string) retry.StateRefreshFunc {
	return func(ctx context.Context) (any, string, error) {
		out, err := findGatewayTargetByTwoPartKey(ctx, conn, gatewayIdentifier, targetID)
		if retry.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", smarterr.NewError(err)
		}

		return out, string(out.Status), nil
	}
}

func findGatewayTargetByTwoPartKey(ctx context.Context, conn *bedrockagentcorecontrol.Client, gatewayIdentifier, targetID string) (*bedrockagentcorecontrol.GetGatewayTargetOutput, error) {
	input := bedrockagentcorecontrol.GetGatewayTargetInput{
		GatewayIdentifier: aws.String(gatewayIdentifier),
		TargetId:          aws.String(targetID),
	}

	return findGatewayTarget(ctx, conn, &input)
}

func findGatewayTarget(ctx context.Context, conn *bedrockagentcorecontrol.Client, input *bedrockagentcorecontrol.GetGatewayTargetInput) (*bedrockagentcorecontrol.GetGatewayTargetOutput, error) {
	out, err := conn.GetGatewayTarget(ctx, input)

	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return nil, smarterr.NewError(&retry.NotFoundError{
			LastError: err,
		})
	}

	if err != nil {
		return nil, smarterr.NewError(err)
	}

	if out == nil {
		return nil, smarterr.NewError(tfresource.NewEmptyResultError())
	}

	return out, nil
}

type gatewayTargetResourceModel struct {
	framework.WithRegionModel
	CredentialProviderConfiguration fwtypes.ListNestedObjectValueOf[credentialProviderConfigurationModel] `tfsdk:"credential_provider_configuration"`
	Description                     types.String                                                          `tfsdk:"description"`
	GatewayIdentifier               types.String                                                          `tfsdk:"gateway_identifier"`
	Name                            types.String                                                          `tfsdk:"name"`
	TargetConfiguration             fwtypes.ListNestedObjectValueOf[targetConfigurationModel]             `tfsdk:"target_configuration"`
	TargetID                        types.String                                                          `tfsdk:"target_id"`
	Timeouts                        timeouts.Value                                                        `tfsdk:"timeouts"`
}

type credentialProviderConfigurationModel struct {
	ApiKey         fwtypes.ListNestedObjectValueOf[apiKeyCredentialProviderModel] `tfsdk:"api_key"`
	OAuth          fwtypes.ListNestedObjectValueOf[oauthCredentialProviderModel]  `tfsdk:"oauth"`
	GatewayIAMRole fwtypes.ListNestedObjectValueOf[gatewayIAMRoleProviderModel]   `tfsdk:"gateway_iam_role"`
}

var (
	_ fwflex.Expander  = credentialProviderConfigurationModel{}
	_ fwflex.Flattener = &credentialProviderConfigurationModel{}
)

func (m *credentialProviderConfigurationModel) Flatten(ctx context.Context, v any) diag.Diagnostics {
	var diags diag.Diagnostics
	switch t := v.(type) {
	case awstypes.CredentialProviderConfiguration:
		switch t.CredentialProviderType {
		case awstypes.CredentialProviderTypeApiKey:
			if apiKeyProvider, ok := t.CredentialProvider.(*awstypes.CredentialProviderMemberApiKeyCredentialProvider); ok {
				var model apiKeyCredentialProviderModel
				smerr.AddEnrich(ctx, &diags, fwflex.Flatten(ctx, apiKeyProvider.Value, &model))
				if diags.HasError() {
					return diags
				}
				m.ApiKey = fwtypes.NewListNestedObjectValueOfPtrMust(ctx, &model)
			}

		case awstypes.CredentialProviderTypeOauth:
			if oauthProvider, ok := t.CredentialProvider.(*awstypes.CredentialProviderMemberOauthCredentialProvider); ok {
				var model oauthCredentialProviderModel
				smerr.AddEnrich(ctx, &diags, fwflex.Flatten(ctx, oauthProvider.Value, &model))
				if diags.HasError() {
					return diags
				}
				m.OAuth = fwtypes.NewListNestedObjectValueOfPtrMust(ctx, &model)
			}

		case awstypes.CredentialProviderTypeGatewayIamRole:
			m.GatewayIAMRole = fwtypes.NewListNestedObjectValueOfPtrMust(ctx, &gatewayIAMRoleProviderModel{})

		default:
			diags.AddError(
				"Unknown Credential Provider Type",
				fmt.Sprintf("Received unknown credential provider type: %s", t.CredentialProviderType),
			)
		}

	default:
		diags.AddError(
			"Invalid Credential Provider Configuration",
			fmt.Sprintf("Received unexpected type: %T", v),
		)
	}
	return diags
}

func (m credentialProviderConfigurationModel) Expand(ctx context.Context) (any, diag.Diagnostics) {
	var diags diag.Diagnostics
	var c awstypes.CredentialProviderConfiguration
	switch {
	case !m.ApiKey.IsNull():
		apiKeyCredentialProviderConfigurationData, d := m.ApiKey.ToPtr(ctx)
		smerr.AddEnrich(ctx, &diags, d)
		if diags.HasError() {
			return nil, diags
		}

		var r awstypes.CredentialProviderMemberApiKeyCredentialProvider
		smerr.AddEnrich(ctx, &diags, fwflex.Expand(ctx, apiKeyCredentialProviderConfigurationData, &r.Value))
		if diags.HasError() {
			return nil, diags
		}
		c.CredentialProviderType = awstypes.CredentialProviderTypeApiKey
		c.CredentialProvider = &r
		return &c, diags

	case !m.OAuth.IsNull():
		oauthCredentialProviderConfigurationData, d := m.OAuth.ToPtr(ctx)
		smerr.AddEnrich(ctx, &diags, d)
		if diags.HasError() {
			return nil, diags
		}

		var r awstypes.CredentialProviderMemberOauthCredentialProvider
		smerr.AddEnrich(ctx, &diags, fwflex.Expand(ctx, oauthCredentialProviderConfigurationData, &r.Value))
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
		return nil, diags
	}
}

type targetConfigurationModel struct {
	MCP fwtypes.ListNestedObjectValueOf[mcpConfigurationModel] `tfsdk:"mcp"`
}

func (m *targetConfigurationModel) GetConfigurationType(ctx context.Context) string {
	switch mcpData, _ := m.MCP.ToPtr(ctx); {
	case !mcpData.Lambda.IsNull():
		return "lambda"
	case !mcpData.MCPServer.IsNull():
		return "mcp_server"
	case !mcpData.OpenApiSchema.IsNull():
		return "open_api_schema"
	case !mcpData.SmithyModel.IsNull():
		return "smithy_model"
	default:
		return "unknown"
	}
}

var (
	_ fwflex.Expander  = targetConfigurationModel{}
	_ fwflex.Flattener = &targetConfigurationModel{}
)

func (m *targetConfigurationModel) Flatten(ctx context.Context, v any) diag.Diagnostics {
	var diags diag.Diagnostics
	switch t := v.(type) {
	case awstypes.TargetConfigurationMemberMcp:
		var model mcpConfigurationModel
		smerr.AddEnrich(ctx, &diags, fwflex.Flatten(ctx, t.Value, &model))
		if diags.HasError() {
			return diags
		}

		m.MCP = fwtypes.NewListNestedObjectValueOfPtrMust(ctx, &model)
	default:
		diags.AddError(
			"Unsupported Type",
			fmt.Sprintf("target configuration flatten: %T", v),
		)
	}
	return diags
}

func (m targetConfigurationModel) Expand(ctx context.Context) (any, diag.Diagnostics) {
	var diags diag.Diagnostics
	switch {
	case !m.MCP.IsNull():
		mcpConfigurationData, d := m.MCP.ToPtr(ctx)
		smerr.AddEnrich(ctx, &diags, d)
		if diags.HasError() {
			return nil, diags
		}

		var r awstypes.TargetConfigurationMemberMcp
		smerr.AddEnrich(ctx, &diags, fwflex.Expand(ctx, mcpConfigurationData, &r.Value))
		if diags.HasError() {
			return nil, diags
		}

		return &r, diags
	}

	return nil, diags
}

type mcpConfigurationModel struct {
	Lambda        fwtypes.ListNestedObjectValueOf[mcpLambdaConfigurationModel] `tfsdk:"lambda"`
	MCPServer     fwtypes.ListNestedObjectValueOf[mcpServerConfigurationModel] `tfsdk:"mcp_server"`
	SmithyModel   fwtypes.ListNestedObjectValueOf[apiSchemaConfigurationModel] `tfsdk:"smithy_model"`
	OpenApiSchema fwtypes.ListNestedObjectValueOf[apiSchemaConfigurationModel] `tfsdk:"open_api_schema"`
}

var (
	_ fwflex.Expander  = mcpConfigurationModel{}
	_ fwflex.Flattener = &mcpConfigurationModel{}
)

func (m *mcpConfigurationModel) Flatten(ctx context.Context, v any) diag.Diagnostics {
	var diags diag.Diagnostics
	switch t := v.(type) {
	case awstypes.McpTargetConfigurationMemberLambda:
		var model mcpLambdaConfigurationModel
		smerr.AddEnrich(ctx, &diags, fwflex.Flatten(ctx, t.Value, &model))
		if diags.HasError() {
			return diags
		}
		m.Lambda = fwtypes.NewListNestedObjectValueOfPtrMust(ctx, &model)

	case awstypes.McpTargetConfigurationMemberMcpServer:
		var model mcpServerConfigurationModel
		smerr.AddEnrich(ctx, &diags, fwflex.Flatten(ctx, t.Value, &model))
		if diags.HasError() {
			return diags
		}
		m.MCPServer = fwtypes.NewListNestedObjectValueOfPtrMust(ctx, &model)

	case awstypes.McpTargetConfigurationMemberOpenApiSchema:
		var model apiSchemaConfigurationModel
		smerr.AddEnrich(ctx, &diags, fwflex.Flatten(ctx, t.Value, &model))
		if diags.HasError() {
			return diags
		}
		m.OpenApiSchema = fwtypes.NewListNestedObjectValueOfPtrMust(ctx, &model)

	case awstypes.McpTargetConfigurationMemberSmithyModel:
		var model apiSchemaConfigurationModel
		smerr.AddEnrich(ctx, &diags, fwflex.Flatten(ctx, t.Value, &model))
		if diags.HasError() {
			return diags
		}
		m.SmithyModel = fwtypes.NewListNestedObjectValueOfPtrMust(ctx, &model)

	default:
		diags.AddError(
			"Unsupported Type",
			fmt.Sprintf("mcp configuration flatten: %T", v),
		)
	}
	return diags
}

func (m mcpConfigurationModel) Expand(ctx context.Context) (any, diag.Diagnostics) {
	var diags diag.Diagnostics
	switch {
	case !m.Lambda.IsNull():
		lambdaMCPConfigurationData, d := m.Lambda.ToPtr(ctx)
		smerr.AddEnrich(ctx, &diags, d)
		if diags.HasError() {
			return nil, diags
		}

		var r awstypes.McpTargetConfigurationMemberLambda
		smerr.AddEnrich(ctx, &diags, fwflex.Expand(ctx, lambdaMCPConfigurationData, &r.Value))
		if diags.HasError() {
			return nil, diags
		}
		return &r, diags

	case !m.MCPServer.IsNull():
		mcpServerConfigurationData, d := m.MCPServer.ToPtr(ctx)
		smerr.AddEnrich(ctx, &diags, d)
		if diags.HasError() {
			return nil, diags
		}

		var r awstypes.McpTargetConfigurationMemberMcpServer
		smerr.AddEnrich(ctx, &diags, fwflex.Expand(ctx, mcpServerConfigurationData, &r.Value))
		if diags.HasError() {
			return nil, diags
		}
		return &r, diags

	case !m.OpenApiSchema.IsNull():
		openApiMCPConfigurationData, d := m.OpenApiSchema.ToPtr(ctx)
		smerr.AddEnrich(ctx, &diags, d)
		if diags.HasError() {
			return nil, diags
		}

		var r awstypes.McpTargetConfigurationMemberOpenApiSchema
		smerr.AddEnrich(ctx, &diags, fwflex.Expand(ctx, openApiMCPConfigurationData, &r.Value))
		if diags.HasError() {
			return nil, diags
		}
		return &r, diags

	case !m.SmithyModel.IsNull():
		smithyMCPConfigurationData, d := m.SmithyModel.ToPtr(ctx)
		smerr.AddEnrich(ctx, &diags, d)
		if diags.HasError() {
			return nil, diags
		}
		var r awstypes.McpTargetConfigurationMemberSmithyModel
		smerr.AddEnrich(ctx, &diags, fwflex.Expand(ctx, smithyMCPConfigurationData, &r.Value))
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
	ProviderARN             fwtypes.ARN                                           `tfsdk:"provider_arn"`
}

type oauthCredentialProviderModel struct {
	CustomParameters fwtypes.MapOfString `tfsdk:"custom_parameters"`
	ProviderARN      fwtypes.ARN         `tfsdk:"provider_arn"`
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
	_ fwflex.Expander  = toolSchemaModel{}
	_ fwflex.Flattener = &toolSchemaModel{}
)

func (m *toolSchemaModel) Flatten(ctx context.Context, v any) diag.Diagnostics {
	var diags diag.Diagnostics
	switch t := v.(type) {
	case awstypes.ToolSchemaMemberInlinePayload:
		var toolDefModels []*toolDefinitionModel
		for _, toolDef := range t.Value {
			var model toolDefinitionModel
			smerr.AddEnrich(ctx, &diags, fwflex.Flatten(ctx, toolDef, &model))
			if diags.HasError() {
				return diags
			}
			toolDefModels = append(toolDefModels, &model)
		}
		m.InlinePayload = fwtypes.NewListNestedObjectValueOfSliceMust(ctx, toolDefModels)

	case awstypes.ToolSchemaMemberS3:
		var model s3ConfigurationModel
		smerr.AddEnrich(ctx, &diags, fwflex.Flatten(ctx, t.Value, &model))
		if diags.HasError() {
			return diags
		}
		m.S3 = fwtypes.NewListNestedObjectValueOfPtrMust(ctx, &model)

	default:
		diags.AddError(
			"Unsupported Type",
			fmt.Sprintf("tool schema configuration flatten: %T", v),
		)
	}
	return diags
}

func (m toolSchemaModel) Expand(ctx context.Context) (any, diag.Diagnostics) {
	var diags diag.Diagnostics
	switch {
	case !m.InlinePayload.IsNull():
		inlinePayloadToolSchemaData, d := m.InlinePayload.ToSlice(ctx)
		smerr.AddEnrich(ctx, &diags, d)
		if diags.HasError() {
			return nil, diags
		}

		var toolDefs []awstypes.ToolDefinition
		for _, toolDefModel := range inlinePayloadToolSchemaData {
			var toolDef awstypes.ToolDefinition
			smerr.AddEnrich(ctx, &diags, fwflex.Expand(ctx, toolDefModel, &toolDef))
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
		smerr.AddEnrich(ctx, &diags, d)
		if diags.HasError() {
			return nil, diags
		}

		var r awstypes.ToolSchemaMemberS3
		smerr.AddEnrich(ctx, &diags, fwflex.Expand(ctx, s3ToolSchemaData, &r.Value))
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
	_ fwflex.Expander  = schemaDefinitionModel{}
	_ fwflex.Flattener = &schemaDefinitionModel{}
)

func (m *schemaDefinitionModel) Flatten(ctx context.Context, v any) diag.Diagnostics {
	var diags diag.Diagnostics
	// Ensure Properties is a typed Null when absent to avoid zero-value Set panics during state encoding
	m.Properties = fwtypes.NewSetNestedObjectValueOfNull[schemaPropertyModel](ctx)
	switch t := v.(type) {
	case awstypes.SchemaDefinition:
		smerr.AddEnrich(ctx, &diags, fwflex.Flatten(ctx, v, &m.schemaDefinitionCoreModel))
		if diags.HasError() {
			return diags
		}

		// Normalize: when API omits Items, return an empty list (not null)
		if t.Items == nil {
			m.Items = fwtypes.NewListNestedObjectValueOfSliceMust(ctx, []*schemaItemsModel{})
		}

		if t.Properties != nil {
			properties, d := flattenTargetSchemaProperties(ctx, t.Properties, t.Required)
			smerr.AddEnrich(ctx, &diags, d)
			if diags.HasError() {
				return diags
			}
			m.Properties = properties
		}
	default:
		diags.AddError(
			"Unsupported Type",
			fmt.Sprintf("schema definition flatten: %T", v),
		)
	}

	return diags
}

func (m schemaDefinitionModel) Expand(ctx context.Context) (any, diag.Diagnostics) {
	var diags diag.Diagnostics
	schemaDefinitionData := &awstypes.SchemaDefinition{}

	diags.Append(fwflex.Expand(ctx, m.schemaDefinitionCoreModel, schemaDefinitionData)...)
	if diags.HasError() {
		return nil, diags
	}

	if !m.Properties.IsNull() {
		properties, requiredProps, d := expandTargetSchemaProperties(ctx, m.Properties)
		smerr.AddEnrich(ctx, &diags, d)
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
	_ fwflex.Expander  = schemaItemsModel{}
	_ fwflex.Flattener = &schemaItemsModel{}
)

func (m *schemaItemsModel) Flatten(ctx context.Context, v any) diag.Diagnostics {
	var diags diag.Diagnostics
	// Ensure Properties is a typed Null when absent to avoid zero-value Set panics during state encoding
	m.Properties = fwtypes.NewSetNestedObjectValueOfNull[schemaPropertyLeafModel](ctx)
	switch t := v.(type) {
	case awstypes.SchemaDefinition:
		smerr.AddEnrich(ctx, &diags, fwflex.Flatten(ctx, v, &m.schemaItemsCoreModel))
		if diags.HasError() {
			return diags
		}

		// Normalize: when API omits Items, return an empty list (not null)
		if t.Items == nil {
			m.Items = fwtypes.NewListNestedObjectValueOfSliceMust(ctx, []*schemaItemsLeafModel{})
		}

		if t.Properties != nil {
			properties, d := flattenTargetSchemaLeafProperties(ctx, t.Properties, t.Required)
			smerr.AddEnrich(ctx, &diags, d)
			if diags.HasError() {
				return diags
			}
			m.Properties = properties
		}
	default:
		diags.AddError(
			"Unsupported Type",
			fmt.Sprintf("schema items flatten: %T", v),
		)
	}

	return diags
}

func (m schemaItemsModel) Expand(ctx context.Context) (any, diag.Diagnostics) {
	var diags diag.Diagnostics
	schemaDefinitionData := &awstypes.SchemaDefinition{}

	diags.Append(fwflex.Expand(ctx, m.schemaItemsCoreModel, schemaDefinitionData)...)
	if diags.HasError() {
		return nil, diags
	}

	if !m.Properties.IsNull() {
		properties, requiredProps, d := expandTargetSchemaLeafProperties(ctx, m.Properties)
		smerr.AddEnrich(ctx, &diags, d)
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
	_ fwflex.Expander  = schemaPropertyModel{}
	_ fwflex.Flattener = &schemaPropertyModel{}
)

func (m *schemaPropertyModel) Flatten(ctx context.Context, v any) diag.Diagnostics {
	var diags diag.Diagnostics
	// Ensure Properties is a typed Null when absent to avoid zero-value Set panics during state encoding
	m.Properties = fwtypes.NewSetNestedObjectValueOfNull[schemaPropertyLeafModel](ctx)
	switch t := v.(type) {
	case awstypes.SchemaDefinition:
		smerr.AddEnrich(ctx, &diags, fwflex.Flatten(ctx, v, &m.schemaPropertyCoreModel))
		if diags.HasError() {
			return diags
		}

		// Normalize: when API omits Items, return an empty list (not null)
		if t.Items == nil {
			m.Items = fwtypes.NewListNestedObjectValueOfSliceMust(ctx, []*schemaItemsModel{})
		}

		if t.Properties != nil {
			properties, d := flattenTargetSchemaLeafProperties(ctx, t.Properties, t.Required)
			smerr.AddEnrich(ctx, &diags, d)
			if diags.HasError() {
				return diags
			}
			m.Properties = properties
		}
	default:
		diags.AddError(
			"Unsupported Type",
			fmt.Sprintf("schema property flatten: %T", v),
		)
	}
	return diags
}

func (m schemaPropertyModel) Expand(ctx context.Context) (any, diag.Diagnostics) {
	var diags diag.Diagnostics
	var schemaDefinitionLeafData = awstypes.SchemaDefinition{}
	smerr.AddEnrich(ctx, &diags, fwflex.Expand(ctx, m.schemaPropertyCoreModel, &schemaDefinitionLeafData))
	if diags.HasError() {
		return nil, diags
	}

	if !m.Properties.IsNull() {
		properties, requiredProps, d := expandTargetSchemaLeafProperties(ctx, m.Properties)
		smerr.AddEnrich(ctx, &diags, d)
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
	_ fwflex.Expander  = schemaItemsLeafModel{}
	_ fwflex.Flattener = &schemaItemsLeafModel{}
)

func (m *schemaItemsLeafModel) Flatten(ctx context.Context, v any) diag.Diagnostics {
	var diags diag.Diagnostics
	switch t := v.(type) {
	case awstypes.SchemaDefinition:
		smerr.AddEnrich(ctx, &diags, fwflex.Flatten(ctx, v, &m.schemaItemsLeafCoreModel))
		if diags.HasError() {
			return diags
		}
		// Populate ItemsJSON
		if t.Items != nil {
			jsonItems := convertToJSONSchemaDefinition(t.Items)
			s, err := tfjson.EncodeToString(jsonItems)
			if err != nil {
				diags.AddWarning("Failed to marshal items for items_json", err.Error())
				m.ItemsJSON = types.StringNull()
			} else {
				m.ItemsJSON = types.StringValue(s)
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
			s, err := tfjson.EncodeToString(jsonProps)
			if err != nil {
				diags.AddWarning("Failed to marshal properties for properties_json", err.Error())
				m.PropertiesJSON = types.StringNull()
			} else {
				m.PropertiesJSON = types.StringValue(s)
			}
		} else {
			m.PropertiesJSON = types.StringNull()
		}
	default:
		diags.AddError(
			"Unsupported Type",
			fmt.Sprintf("schema items leaf flatten: %T", v),
		)
	}
	return diags
}

func (m schemaItemsLeafModel) Expand(ctx context.Context) (any, diag.Diagnostics) {
	var diags diag.Diagnostics
	var sd awstypes.SchemaDefinition
	// Expand core (type/description)
	smerr.AddEnrich(ctx, &diags, fwflex.Expand(ctx, m.schemaItemsLeafCoreModel, &sd))
	if diags.HasError() {
		return nil, diags
	}

	if isNonEmpty(m.ItemsJSON) {
		jsd, d := parseJSONSchemaDefinition(m.ItemsJSON.ValueString())
		smerr.AddEnrich(ctx, &diags, d)
		if diags.HasError() {
			return nil, diags
		}
		sd.Items = jsd
	}
	if isNonEmpty(m.PropertiesJSON) {
		jsd, d := parseJSONSchemaDefinition(m.PropertiesJSON.ValueString())
		smerr.AddEnrich(ctx, &diags, d)
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
	_ fwflex.Expander  = schemaPropertyLeafModel{}
	_ fwflex.Flattener = &schemaPropertyLeafModel{}
)

func (m *schemaPropertyLeafModel) Flatten(ctx context.Context, v any) diag.Diagnostics {
	var diags diag.Diagnostics
	switch t := v.(type) {
	case awstypes.SchemaDefinition:
		smerr.AddEnrich(ctx, &diags, fwflex.Flatten(ctx, v, &m.schemaPropertyLeafCoreModel))
		if diags.HasError() {
			return diags
		}
		// Populate ItemsJSON
		if t.Items != nil {
			jsonItems := convertToJSONSchemaDefinition(t.Items)
			s, err := tfjson.EncodeToString(jsonItems)
			if err != nil {
				diags.AddWarning("Failed to marshal items for items_json", err.Error())
				m.ItemsJSON = types.StringNull()
			} else {
				m.ItemsJSON = types.StringValue(strings.TrimSpace(s))
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
			s, err := tfjson.EncodeToString(jsonProps)
			if err != nil {
				diags.AddWarning("Failed to marshal properties for properties_json", err.Error())
				m.PropertiesJSON = types.StringNull()
			} else {
				m.PropertiesJSON = types.StringValue(s)
			}
		} else {
			m.PropertiesJSON = types.StringNull()
		}
	default:
		diags.AddError(
			"Unsupported Type",
			fmt.Sprintf("schema property leaf flatten: %T", v),
		)
	}
	return diags
}

func (m schemaPropertyLeafModel) Expand(ctx context.Context) (any, diag.Diagnostics) {
	var diags diag.Diagnostics
	var schemaDefinitionData = awstypes.SchemaDefinition{}

	smerr.AddEnrich(ctx, &diags, fwflex.Expand(ctx, m.schemaPropertyLeafCoreModel, &schemaDefinitionData))
	if diags.HasError() {
		return nil, diags
	}

	if isNonEmpty(m.ItemsJSON) {
		jsd, d := parseJSONSchemaDefinition(m.ItemsJSON.ValueString())
		smerr.AddEnrich(ctx, &diags, d)
		if diags.HasError() {
			return nil, diags
		}
		schemaDefinitionData.Items = jsd
	}
	if isNonEmpty(m.PropertiesJSON) {
		jsd, d := parseJSONSchemaDefinition(m.PropertiesJSON.ValueString())
		smerr.AddEnrich(ctx, &diags, d)
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

type mcpServerConfigurationModel struct {
	Endpoint types.String `tfsdk:"endpoint"`
}

type apiSchemaConfigurationModel struct {
	InlinePayload fwtypes.ListNestedObjectValueOf[inlinePayloadModel]   `tfsdk:"inline_payload"`
	S3            fwtypes.ListNestedObjectValueOf[s3ConfigurationModel] `tfsdk:"s3"`
}

var (
	_ fwflex.Expander  = apiSchemaConfigurationModel{}
	_ fwflex.Flattener = &apiSchemaConfigurationModel{}
)

func (m *apiSchemaConfigurationModel) Flatten(ctx context.Context, v any) diag.Diagnostics {
	var diags diag.Diagnostics
	switch t := v.(type) {
	case awstypes.ApiSchemaConfigurationMemberInlinePayload:
		var model inlinePayloadModel
		model.Payload = types.StringValue(t.Value)
		m.InlinePayload = fwtypes.NewListNestedObjectValueOfPtrMust(ctx, &model)
		return diags

	case awstypes.ApiSchemaConfigurationMemberS3:
		var model s3ConfigurationModel
		d := fwflex.Flatten(ctx, t.Value, &model)
		smerr.AddEnrich(ctx, &diags, d)
		if diags.HasError() {
			return diags
		}
		m.S3 = fwtypes.NewListNestedObjectValueOfPtrMust(ctx, &model)

	default:
		diags.AddError(
			"Unsupported Type",
			fmt.Sprintf("api schema configuration flatten: %T", v),
		)
	}
	return diags
}

func (m apiSchemaConfigurationModel) Expand(ctx context.Context) (any, diag.Diagnostics) {
	var diags diag.Diagnostics
	switch {
	case !m.InlinePayload.IsNull():
		inlinePayloadApiSchemaConfigurationData, d := m.InlinePayload.ToPtr(ctx)
		smerr.AddEnrich(ctx, &diags, d)
		if diags.HasError() {
			return nil, diags
		}

		var r awstypes.ApiSchemaConfigurationMemberInlinePayload
		r.Value = inlinePayloadApiSchemaConfigurationData.Payload.ValueString()
		return &r, diags

	case !m.S3.IsNull():
		s3ApiSchemaConfigurationData, d := m.S3.ToPtr(ctx)
		smerr.AddEnrich(ctx, &diags, d)
		if diags.HasError() {
			return nil, diags
		}

		var r awstypes.ApiSchemaConfigurationMemberS3
		smerr.AddEnrich(ctx, &diags, fwflex.Expand(ctx, s3ApiSchemaConfigurationData, &r.Value))
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
		smerr.AddEnrich(ctx, &diags, d)
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
	smerr.AddEnrich(ctx, &diags, d)
	if diags.HasError() {
		return nil, nil, diags
	}

	for _, propertyModel := range propertySlice {
		expandedValue, d := propertyModel.Expand(ctx)
		smerr.AddEnrich(ctx, &diags, d)
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
		smerr.AddEnrich(ctx, &diags, d)
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
	smerr.AddEnrich(ctx, &diags, d)
	if diags.HasError() {
		return nil, nil, diags
	}

	for _, propertyModel := range propertySlice {
		expandedValue, d := propertyModel.Expand(ctx)
		smerr.AddEnrich(ctx, &diags, d)
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
	if err := tfjson.DecodeFromString(s, &sd); err != nil {
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
