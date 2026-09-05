// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package agentregistry

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/YakDriver/regexache"
	"github.com/YakDriver/smarterr"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/agentregistrycontrol"
	awstypes "github.com/aws/aws-sdk-go-v2/service/agentregistrycontrol/types"
	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/objectvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/setvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/fwdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	fwflex "github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/smerr"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// Function annotations are used for resource registration to the Provider. DO NOT EDIT.
// @FrameworkResource("aws_agentregistry_registry", name="Registry")
// @Tags(identifierAttribute="registry_arn")
// @Testing(hasNoPreExistingResource=true)
// @Testing(generator="randomWithPrefixAndUnderscore(t)")
// @Testing(importStateIdAttribute="registry_id")
// @IdentityAttribute("registry_id")
func newRegistryResource(_ context.Context) (resource.ResourceWithConfigure, error) {
	r := &registryResource{}

	r.SetDefaultCreateTimeout(30 * time.Minute)
	r.SetDefaultUpdateTimeout(30 * time.Minute)
	r.SetDefaultDeleteTimeout(30 * time.Minute)

	return r, nil
}

type registryResource struct {
	framework.ResourceWithModel[registryResourceModel]
	framework.WithTimeouts
	framework.WithImportByIdentity
}

func (r *registryResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			names.AttrDescription: schema.StringAttribute{
				Optional: true,
				Validators: []validator.String{
					stringvalidator.LengthBetween(1, 4096),
				},
			},
			names.AttrName: schema.StringAttribute{
				Required: true,
				Validators: []validator.String{
					stringvalidator.RegexMatches(
						regexache.MustCompile(`^[A-Za-z0-9_-]+$`),
						"must contain only letters, numbers, hyphens, and underscores",
					),
					stringvalidator.LengthBetween(1, 64),
				},
			},
			"registry_arn": framework.ARNAttributeComputedOnly(),
			"registry_id":  framework.IDAttribute(),
			names.AttrStatus: schema.StringAttribute{
				CustomType: fwtypes.StringEnumType[awstypes.RegistryStatus](),
				Computed:   true,
			},
			names.AttrTags:    tftags.TagsAttribute(),
			names.AttrTagsAll: tftags.TagsAttributeComputedOnly(),
		},
		Blocks: map[string]schema.Block{
			"approval_configuration": schema.ListNestedBlock{
				CustomType: fwtypes.NewListNestedObjectTypeOf[approvalConfigurationModel](ctx),
				Validators: []validator.List{
					listvalidator.SizeAtMost(1),
				},
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"auto_approval_rules": schema.SetAttribute{
							CustomType:  fwtypes.SetOfStringType,
							ElementType: types.StringType,
							Optional:    true,
							Validators: []validator.Set{
								setvalidator.ValueStringsAre(stringvalidator.OneOf("APPROVE_ALL")),
							},
						},
					},
				},
			},
			"discovery_configuration": schema.ListNestedBlock{
				CustomType: fwtypes.NewListNestedObjectTypeOf[discoveryConfigurationModel](ctx),
				Validators: []validator.List{
					listvalidator.IsRequired(),
					listvalidator.SizeAtMost(1),
				},
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"authorizer_type": schema.StringAttribute{
							CustomType: fwtypes.StringEnumType[awstypes.RegistryAuthorizerType](),
							Required:   true,
							PlanModifiers: []planmodifier.String{
								stringplanmodifier.RequiresReplace(),
							},
						},
					},
					Blocks: map[string]schema.Block{
						"authorizer_configuration": schema.ListNestedBlock{
							CustomType: fwtypes.NewListNestedObjectTypeOf[authorizerConfigurationModel](ctx),
							Validators: []validator.List{
								listvalidator.SizeAtMost(1),
							},
							NestedObject: schema.NestedBlockObject{
								Attributes: map[string]schema.Attribute{
									"allowed_audience": schema.ListAttribute{
										CustomType:  fwtypes.ListOfStringType,
										ElementType: types.StringType,
										Optional:    true,
									},
									"allowed_clients": schema.ListAttribute{
										CustomType:  fwtypes.ListOfStringType,
										ElementType: types.StringType,
										Optional:    true,
									},
									"allowed_scopes": schema.ListAttribute{
										CustomType:  fwtypes.ListOfStringType,
										ElementType: types.StringType,
										Optional:    true,
									},
									"discovery_url": schema.StringAttribute{
										Required: true,
									},
								},
								Blocks: map[string]schema.Block{
									"custom_claim": schema.SetNestedBlock{
										CustomType: fwtypes.NewSetNestedObjectTypeOf[customJWTAuthorizerCustomClaimModel](ctx),
										NestedObject: schema.NestedBlockObject{
											Attributes: map[string]schema.Attribute{
												"inbound_token_claim_name": schema.StringAttribute{
													Required: true,
													Validators: []validator.String{
														stringvalidator.LengthBetween(1, 255),
														stringvalidator.RegexMatches(regexache.MustCompile(`^[A-Za-z0-9_.:-]+$`), "must contain only letters, numbers, and the characters _ . - :"),
													},
												},
												"inbound_token_claim_value_type": schema.StringAttribute{
													CustomType: fwtypes.StringEnumType[awstypes.InboundTokenClaimValueType](),
													Required:   true,
												},
											},
											Blocks: map[string]schema.Block{
												"authorizing_claim_match_value": schema.ListNestedBlock{
													CustomType: fwtypes.NewListNestedObjectTypeOf[customJWTAuthorizerAuthorizingClaimMatchValueModel](ctx),
													Validators: []validator.List{
														listvalidator.IsRequired(),
														listvalidator.SizeAtMost(1),
													},
													NestedObject: schema.NestedBlockObject{
														Attributes: map[string]schema.Attribute{
															"claim_match_operator": schema.StringAttribute{
																CustomType: fwtypes.StringEnumType[awstypes.ClaimMatchOperatorType](),
																Required:   true,
															},
														},
														Blocks: map[string]schema.Block{
															"claim_match_value": schema.ListNestedBlock{
																CustomType: fwtypes.NewListNestedObjectTypeOf[customJWTAuthorizerClaimMatchValueModel](ctx),
																Validators: []validator.List{
																	listvalidator.IsRequired(),
																	listvalidator.SizeAtMost(1),
																},
																NestedObject: schema.NestedBlockObject{
																	Validators: []validator.Object{
																		objectvalidator.ExactlyOneOf(
																			path.MatchRelative().AtName("match_value_string"),
																			path.MatchRelative().AtName("match_value_string_list"),
																		),
																	},
																	Attributes: map[string]schema.Attribute{
																		"match_value_string": schema.StringAttribute{
																			Optional: true,
																			Validators: []validator.String{
																				stringvalidator.LengthBetween(1, 255),
																				stringvalidator.RegexMatches(regexache.MustCompile(`^[A-Za-z0-9_.:-]+$`), "must contain only letters, numbers, and the characters _ . - :"),
																			},
																		},
																		"match_value_string_list": schema.SetAttribute{
																			Optional:    true,
																			ElementType: types.StringType,
																			Validators: []validator.Set{
																				setvalidator.ValueStringsAre(
																					stringvalidator.LengthBetween(1, 255),
																					stringvalidator.RegexMatches(regexache.MustCompile(`^[A-Za-z0-9_.:-]+$`), "must contain only letters, numbers, and the characters _ . - :"),
																				),
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

func (r *registryResource) ValidateConfig(ctx context.Context, req resource.ValidateConfigRequest, resp *resource.ValidateConfigResponse) {
	var data registryResourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// The listvalidator.IsRequired schema validator handles a missing discovery
	// block. Resource-level validation only enforces the conditional relationship
	// that schema validators cannot express.
	if !data.DiscoveryConfiguration.IsNull() && !data.DiscoveryConfiguration.IsUnknown() {
		var discoveryConfig []discoveryConfigurationModel
		resp.Diagnostics.Append(data.DiscoveryConfiguration.ElementsAs(ctx, &discoveryConfig, false)...)
		if resp.Diagnostics.HasError() {
			return
		}

		if len(discoveryConfig) > 0 {
			config := discoveryConfig[0]
			if config.AuthorizerType.ValueEnum() == awstypes.RegistryAuthorizerTypeCustomJwt &&
				config.AuthorizerConfiguration.IsNull() {
				resp.Diagnostics.AddAttributeError(
					path.Root("discovery_configuration").AtListIndex(0).AtName("authorizer_configuration"),
					"Missing Required Block",
					"authorizer_configuration is required when authorizer_type is CUSTOM_JWT",
				)
			}
		}
	}
}

func (r *registryResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	conn := r.Meta().AgentRegistryClient(ctx)

	var plan registryResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	input := agentregistrycontrol.CreateRegistryInput{
		ClientToken: aws.String(create.UniqueId(ctx)),
		Name:        plan.Name.ValueStringPointer(),
		Description: plan.Description.ValueStringPointer(),
		Tags:        getTagsIn(ctx),
	}

	if !plan.ApprovalConfiguration.IsNull() {
		var approvalConfig []approvalConfigurationModel
		resp.Diagnostics.Append(plan.ApprovalConfiguration.ElementsAs(ctx, &approvalConfig, false)...)
		if resp.Diagnostics.HasError() {
			return
		}

		if len(approvalConfig) > 0 {
			var rules []awstypes.AutoApprovalRule
			resp.Diagnostics.Append(approvalConfig[0].AutoApprovalRules.ElementsAs(ctx, &rules, false)...)
			if resp.Diagnostics.HasError() {
				return
			}

			input.ApprovalConfiguration = &awstypes.ApprovalConfiguration{
				AutoApprovalRules: rules,
			}
		}
	}

	if !plan.DiscoveryConfiguration.IsNull() {
		var discoveryConfig []discoveryConfigurationModel
		resp.Diagnostics.Append(plan.DiscoveryConfiguration.ElementsAs(ctx, &discoveryConfig, false)...)
		if resp.Diagnostics.HasError() {
			return
		}

		if len(discoveryConfig) > 0 {
			config := discoveryConfig[0]
			input.DiscoveryConfiguration = &awstypes.DiscoveryConfiguration{
				AuthorizerType: config.AuthorizerType.ValueEnum(),
			}

			if !config.AuthorizerConfiguration.IsNull() {
				var authConfig []authorizerConfigurationModel
				resp.Diagnostics.Append(config.AuthorizerConfiguration.ElementsAs(ctx, &authConfig, false)...)
				if resp.Diagnostics.HasError() {
					return
				}

				if len(authConfig) > 0 {
					var allowedAudience, allowedClients, allowedScopes []string
					resp.Diagnostics.Append(authConfig[0].AllowedAudience.ElementsAs(ctx, &allowedAudience, false)...)
					resp.Diagnostics.Append(authConfig[0].AllowedClients.ElementsAs(ctx, &allowedClients, false)...)
					resp.Diagnostics.Append(authConfig[0].AllowedScopes.ElementsAs(ctx, &allowedScopes, false)...)
					if resp.Diagnostics.HasError() {
						return
					}

					jwtConfig := awstypes.CustomJWTAuthorizerConfiguration{
						DiscoveryUrl:    authConfig[0].DiscoveryURL.ValueStringPointer(),
						AllowedAudience: allowedAudience,
						AllowedClients:  allowedClients,
						AllowedScopes:   allowedScopes,
					}

					if !authConfig[0].CustomClaim.IsNull() {
						var customClaims []awstypes.CustomClaimValidationType
						smerr.AddEnrich(ctx, &resp.Diagnostics, fwflex.Expand(ctx, authConfig[0].CustomClaim, &customClaims))
						if resp.Diagnostics.HasError() {
							return
						}
						jwtConfig.CustomClaims = customClaims
					}

					input.DiscoveryConfiguration.AuthorizerConfiguration = &awstypes.AuthorizerConfigurationMemberCustomJWTAuthorizer{
						Value: jwtConfig,
					}
				}
			}
		}
	}

	out, err := conn.CreateRegistry(ctx, &input)
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.AgentRegistry, create.ErrActionCreating, "Registry", plan.Name.String(), err),
			err.Error(),
		)
		return
	}

	registryARN := aws.ToString(out.RegistryArn)

	created, err := waitRegistryCreated(ctx, conn, registryARN, r.CreateTimeout(ctx, plan.Timeouts))
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.AgentRegistry, create.ErrActionWaitingForCreation, "Registry", registryARN, err),
			err.Error(),
		)
		return
	}

	// Set values for unknowns
	plan.RegistryID = types.StringValue(aws.ToString(created.RegistryId))
	plan.RegistryARN = types.StringValue(aws.ToString(created.RegistryArn))
	plan.Status = fwtypes.StringEnumValue(created.Status)

	// Flatten the response back into state
	resp.Diagnostics.Append(flattenApprovalConfiguration(ctx, created.ApprovalConfiguration, &plan)...)
	resp.Diagnostics.Append(flattenDiscoveryConfiguration(ctx, created.DiscoveryConfiguration, &plan)...)

	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *registryResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	conn := r.Meta().AgentRegistryClient(ctx)

	var state registryResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	out, err := findRegistryByID(ctx, conn, state.RegistryID.ValueString())
	if retry.NotFound(err) {
		resp.Diagnostics.Append(fwdiag.NewResourceNotFoundWarningDiagnostic(err))
		resp.State.RemoveResource(ctx)
		return
	}
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.AgentRegistry, create.ErrActionReading, "Registry", state.RegistryID.String(), err),
			err.Error(),
		)
		return
	}

	// Set computed attributes
	state.RegistryID = types.StringValue(aws.ToString(out.RegistryId))
	state.RegistryARN = types.StringValue(aws.ToString(out.RegistryArn))
	state.Name = types.StringValue(aws.ToString(out.Name))
	state.Status = fwtypes.StringEnumValue(out.Status)

	if out.Description != nil {
		state.Description = types.StringValue(aws.ToString(out.Description))
	} else {
		state.Description = types.StringNull()
	}

	// Flatten nested structures
	resp.Diagnostics.Append(flattenApprovalConfiguration(ctx, out.ApprovalConfiguration, &state)...)
	resp.Diagnostics.Append(flattenDiscoveryConfiguration(ctx, out.DiscoveryConfiguration, &state)...)

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *registryResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	conn := r.Meta().AgentRegistryClient(ctx)

	var plan, state registryResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Framework tag updates can invoke Update with no Registry API field changes.
	// Computed attributes are unknown in the plan, so carry them forward unless
	// the service update below returns fresher values.
	plan.RegistryARN = state.RegistryARN
	plan.RegistryID = state.RegistryID
	plan.Status = state.Status

	if !plan.Name.Equal(state.Name) ||
		!plan.Description.Equal(state.Description) ||
		!plan.ApprovalConfiguration.Equal(state.ApprovalConfiguration) ||
		!plan.DiscoveryConfiguration.Equal(state.DiscoveryConfiguration) {
		input := agentregistrycontrol.UpdateRegistryInput{
			RegistryId: plan.RegistryID.ValueStringPointer(),
		}

		if !plan.Name.Equal(state.Name) {
			input.Name = plan.Name.ValueStringPointer()
		}

		if !plan.Description.Equal(state.Description) {
			input.Description = &awstypes.UpdatedDescription{
				OptionalValue: plan.Description.ValueStringPointer(),
			}
		}

		if !plan.ApprovalConfiguration.Equal(state.ApprovalConfiguration) {
			var approvalConfig []approvalConfigurationModel
			resp.Diagnostics.Append(plan.ApprovalConfiguration.ElementsAs(ctx, &approvalConfig, false)...)
			if resp.Diagnostics.HasError() {
				return
			}

			if len(approvalConfig) > 0 {
				var rules []awstypes.AutoApprovalRule
				resp.Diagnostics.Append(approvalConfig[0].AutoApprovalRules.ElementsAs(ctx, &rules, false)...)
				if resp.Diagnostics.HasError() {
					return
				}

				input.ApprovalConfiguration = &awstypes.UpdatedApprovalConfiguration{
					OptionalValue: &awstypes.ApprovalConfiguration{
						AutoApprovalRules: rules,
					},
				}
			} else {
				// Empty list means clear the approval configuration
				input.ApprovalConfiguration = &awstypes.UpdatedApprovalConfiguration{
					OptionalValue: &awstypes.ApprovalConfiguration{
						AutoApprovalRules: []awstypes.AutoApprovalRule{},
					},
				}
			}
		}

		if !plan.DiscoveryConfiguration.Equal(state.DiscoveryConfiguration) {
			var discoveryConfig []discoveryConfigurationModel
			resp.Diagnostics.Append(plan.DiscoveryConfiguration.ElementsAs(ctx, &discoveryConfig, false)...)
			if resp.Diagnostics.HasError() {
				return
			}

			if len(discoveryConfig) > 0 {
				config := discoveryConfig[0]

				if !config.AuthorizerConfiguration.IsNull() {
					var authConfig []authorizerConfigurationModel
					resp.Diagnostics.Append(config.AuthorizerConfiguration.ElementsAs(ctx, &authConfig, false)...)
					if resp.Diagnostics.HasError() {
						return
					}

					if len(authConfig) > 0 {
						var allowedAudience, allowedClients, allowedScopes []string
						resp.Diagnostics.Append(authConfig[0].AllowedAudience.ElementsAs(ctx, &allowedAudience, false)...)
						resp.Diagnostics.Append(authConfig[0].AllowedClients.ElementsAs(ctx, &allowedClients, false)...)
						resp.Diagnostics.Append(authConfig[0].AllowedScopes.ElementsAs(ctx, &allowedScopes, false)...)
						if resp.Diagnostics.HasError() {
							return
						}

						jwtConfig := awstypes.CustomJWTAuthorizerConfiguration{
							DiscoveryUrl:    authConfig[0].DiscoveryURL.ValueStringPointer(),
							AllowedAudience: allowedAudience,
							AllowedClients:  allowedClients,
							AllowedScopes:   allowedScopes,
						}

						if !authConfig[0].CustomClaim.IsNull() {
							var customClaims []awstypes.CustomClaimValidationType
							smerr.AddEnrich(ctx, &resp.Diagnostics, fwflex.Expand(ctx, authConfig[0].CustomClaim, &customClaims))
							if resp.Diagnostics.HasError() {
								return
							}
							jwtConfig.CustomClaims = customClaims
						}

						input.DiscoveryConfiguration = &awstypes.UpdatedDiscoveryConfiguration{
							AuthorizerConfiguration: &awstypes.UpdatedAuthorizerConfiguration{
								OptionalValue: &awstypes.AuthorizerConfigurationMemberCustomJWTAuthorizer{
									Value: jwtConfig,
								},
							},
						}
					}
				}
			}
		}

		_, err := conn.UpdateRegistry(ctx, &input)
		if err != nil {
			resp.Diagnostics.AddError(
				create.ProblemStandardMessage(names.AgentRegistry, create.ErrActionUpdating, "Registry", plan.RegistryID.String(), err),
				err.Error(),
			)
			return
		}

		updated, err := waitRegistryUpdated(ctx, conn, plan.RegistryID.ValueString(), r.UpdateTimeout(ctx, plan.Timeouts))
		if err != nil {
			resp.Diagnostics.AddError(
				create.ProblemStandardMessage(names.AgentRegistry, create.ErrActionWaitingForUpdate, "Registry", plan.RegistryID.String(), err),
				err.Error(),
			)
			return
		}

		plan.Status = fwtypes.StringEnumValue(updated.Status)
		resp.Diagnostics.Append(flattenApprovalConfiguration(ctx, updated.ApprovalConfiguration, &plan)...)
		resp.Diagnostics.Append(flattenDiscoveryConfiguration(ctx, updated.DiscoveryConfiguration, &plan)...)
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *registryResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	conn := r.Meta().AgentRegistryClient(ctx)

	var state registryResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	registryID := state.RegistryID.ValueString()

	input := agentregistrycontrol.DeleteRegistryInput{
		RegistryId: aws.String(registryID),
	}

	_, err := conn.DeleteRegistry(ctx, &input)
	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return
	}
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.AgentRegistry, create.ErrActionDeleting, "Registry", registryID, err),
			err.Error(),
		)
		return
	}

	if _, err := waitRegistryDeleted(ctx, conn, registryID, r.DeleteTimeout(ctx, state.Timeouts)); err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.AgentRegistry, create.ErrActionWaitingForDeletion, "Registry", registryID, err),
			err.Error(),
		)
		return
	}
}

func findRegistryByID(ctx context.Context, conn *agentregistrycontrol.Client, id string) (*agentregistrycontrol.GetRegistryOutput, error) {
	input := agentregistrycontrol.GetRegistryInput{
		RegistryId: aws.String(id),
	}

	out, err := conn.GetRegistry(ctx, &input)

	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return nil, &retry.NotFoundError{
			LastError: err,
		}
	}

	if err != nil {
		return nil, smarterr.NewError(err)
	}

	if out == nil {
		return nil, tfresource.NewEmptyResultError()
	}

	return out, nil
}

func waitRegistryCreated(ctx context.Context, conn *agentregistrycontrol.Client, id string, timeout time.Duration) (*agentregistrycontrol.GetRegistryOutput, error) {
	stateConf := &retry.StateChangeConf{
		Pending:                   enum.Slice(awstypes.RegistryStatusCreating),
		Target:                    enum.Slice(awstypes.RegistryStatusReady),
		Refresh:                   statusRegistry(conn, id),
		Timeout:                   timeout,
		ContinuousTargetOccurence: 2,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)
	if out, ok := outputRaw.(*agentregistrycontrol.GetRegistryOutput); ok {
		if out.Status == awstypes.RegistryStatusCreateFailed {
			retry.SetLastError(err, errors.New(aws.ToString(out.StatusReason)))
		}
		return out, smarterr.NewError(err)
	}

	return nil, smarterr.NewError(err)
}

func waitRegistryUpdated(ctx context.Context, conn *agentregistrycontrol.Client, id string, timeout time.Duration) (*agentregistrycontrol.GetRegistryOutput, error) {
	stateConf := &retry.StateChangeConf{
		Pending:                   enum.Slice(awstypes.RegistryStatusUpdating),
		Target:                    enum.Slice(awstypes.RegistryStatusReady),
		Refresh:                   statusRegistry(conn, id),
		Timeout:                   timeout,
		ContinuousTargetOccurence: 2,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)
	if out, ok := outputRaw.(*agentregistrycontrol.GetRegistryOutput); ok {
		if out.Status == awstypes.RegistryStatusUpdateFailed {
			retry.SetLastError(err, errors.New(aws.ToString(out.StatusReason)))
		}
		return out, smarterr.NewError(err)
	}

	return nil, smarterr.NewError(err)
}

func waitRegistryDeleted(ctx context.Context, conn *agentregistrycontrol.Client, id string, timeout time.Duration) (*agentregistrycontrol.GetRegistryOutput, error) {
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(awstypes.RegistryStatusDeleting, awstypes.RegistryStatusReady),
		Target:  []string{},
		Refresh: statusRegistry(conn, id),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)
	if out, ok := outputRaw.(*agentregistrycontrol.GetRegistryOutput); ok {
		if out.Status == awstypes.RegistryStatusDeleteFailed {
			retry.SetLastError(err, errors.New(aws.ToString(out.StatusReason)))
		}
		return out, smarterr.NewError(err)
	}

	return nil, smarterr.NewError(err)
}

func statusRegistry(conn *agentregistrycontrol.Client, id string) retry.StateRefreshFunc {
	return func(ctx context.Context) (any, string, error) {
		out, err := findRegistryByID(ctx, conn, id)
		if retry.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", smarterr.NewError(err)
		}

		return out, string(out.Status), nil
	}
}

// The approval and discovery configurations are flattened by hand because an
// omitted block and an explicitly empty one both map to no auto-approval rules,
// a distinction that has to be preserved across refresh.
// nosemgrep:ci.semgrep.framework.manual-flattener-functions
func flattenApprovalConfiguration(ctx context.Context, apiObject *awstypes.ApprovalConfiguration, model *registryResourceModel) diag.Diagnostics {
	var diags diag.Diagnostics

	if apiObject == nil || len(apiObject.AutoApprovalRules) == 0 {
		// Omitted and explicit empty both mean manual approval to AWS. Preserve an
		// explicit configured block across refresh so `auto_approval_rules = []`
		// does not drift to null; imports have no prior block and stay null.
		if model.ApprovalConfiguration.IsNull() {
			model.ApprovalConfiguration = fwtypes.NewListNestedObjectValueOfNull[approvalConfigurationModel](ctx)
			return diags
		}

		emptyRules, d := fwtypes.NewSetValueOf[types.String](ctx, []attr.Value{})
		diags.Append(d...)
		model.ApprovalConfiguration = fwtypes.NewListNestedObjectValueOfValueSliceMust(ctx, []approvalConfigurationModel{{
			AutoApprovalRules: emptyRules,
		}})
		return diags
	}

	// Convert enum slice to attr.Value slice.
	ruleValues := make([]attr.Value, len(apiObject.AutoApprovalRules))
	for i, rule := range apiObject.AutoApprovalRules {
		ruleValues[i] = types.StringValue(string(rule))
	}

	rules, d := fwtypes.NewSetValueOf[types.String](ctx, ruleValues)
	diags.Append(d...)

	approvalConfig := approvalConfigurationModel{
		AutoApprovalRules: rules,
	}

	model.ApprovalConfiguration = fwtypes.NewListNestedObjectValueOfValueSliceMust(ctx, []approvalConfigurationModel{approvalConfig})
	return diags
}

// nosemgrep:ci.semgrep.framework.manual-flattener-functions
func flattenDiscoveryConfiguration(ctx context.Context, apiObject *awstypes.DiscoveryConfiguration, model *registryResourceModel) diag.Diagnostics {
	var diags diag.Diagnostics

	if apiObject == nil {
		model.DiscoveryConfiguration = fwtypes.NewListNestedObjectValueOfNull[discoveryConfigurationModel](ctx)
		return diags
	}

	discoveryConfig := discoveryConfigurationModel{
		AuthorizerType: fwtypes.StringEnumValue(apiObject.AuthorizerType),
	}

	if apiObject.AuthorizerConfiguration != nil {
		if member, ok := apiObject.AuthorizerConfiguration.(*awstypes.AuthorizerConfigurationMemberCustomJWTAuthorizer); ok {
			authConfig := authorizerConfigurationModel{
				DiscoveryURL: types.StringValue(aws.ToString(member.Value.DiscoveryUrl)),
			}

			if len(member.Value.AllowedAudience) > 0 {
				// Convert []string to []attr.Value
				audienceVals := make([]attr.Value, len(member.Value.AllowedAudience))
				for i, v := range member.Value.AllowedAudience {
					audienceVals[i] = types.StringValue(v)
				}
				audience, d := fwtypes.NewListValueOf[types.String](ctx, audienceVals)
				diags.Append(d...)
				authConfig.AllowedAudience = audience
			} else {
				authConfig.AllowedAudience = fwtypes.NewListValueOfNull[types.String](ctx)
			}

			if len(member.Value.AllowedClients) > 0 {
				clientVals := make([]attr.Value, len(member.Value.AllowedClients))
				for i, v := range member.Value.AllowedClients {
					clientVals[i] = types.StringValue(v)
				}
				clients, d := fwtypes.NewListValueOf[types.String](ctx, clientVals)
				diags.Append(d...)
				authConfig.AllowedClients = clients
			} else {
				authConfig.AllowedClients = fwtypes.NewListValueOfNull[types.String](ctx)
			}

			if len(member.Value.AllowedScopes) > 0 {
				scopeVals := make([]attr.Value, len(member.Value.AllowedScopes))
				for i, v := range member.Value.AllowedScopes {
					scopeVals[i] = types.StringValue(v)
				}
				scopes, d := fwtypes.NewListValueOf[types.String](ctx, scopeVals)
				diags.Append(d...)
				authConfig.AllowedScopes = scopes
			} else {
				authConfig.AllowedScopes = fwtypes.NewListValueOfNull[types.String](ctx)
			}

			if len(member.Value.CustomClaims) > 0 {
				var customClaims fwtypes.SetNestedObjectValueOf[customJWTAuthorizerCustomClaimModel]
				smerr.AddEnrich(ctx, &diags, fwflex.Flatten(ctx, member.Value.CustomClaims, &customClaims))
				authConfig.CustomClaim = customClaims
			} else {
				authConfig.CustomClaim = fwtypes.NewSetNestedObjectValueOfNull[customJWTAuthorizerCustomClaimModel](ctx)
			}

			discoveryConfig.AuthorizerConfiguration = fwtypes.NewListNestedObjectValueOfValueSliceMust(ctx, []authorizerConfigurationModel{authConfig})
		}
	} else {
		discoveryConfig.AuthorizerConfiguration = fwtypes.NewListNestedObjectValueOfNull[authorizerConfigurationModel](ctx)
	}

	model.DiscoveryConfiguration = fwtypes.NewListNestedObjectValueOfValueSliceMust(ctx, []discoveryConfigurationModel{discoveryConfig})
	return diags
}

type registryResourceModel struct {
	framework.WithRegionModel
	ApprovalConfiguration  fwtypes.ListNestedObjectValueOf[approvalConfigurationModel]  `tfsdk:"approval_configuration"`
	Description            types.String                                                 `tfsdk:"description"`
	DiscoveryConfiguration fwtypes.ListNestedObjectValueOf[discoveryConfigurationModel] `tfsdk:"discovery_configuration"`
	Name                   types.String                                                 `tfsdk:"name"`
	RegistryARN            types.String                                                 `tfsdk:"registry_arn"`
	RegistryID             types.String                                                 `tfsdk:"registry_id"`
	Status                 fwtypes.StringEnum[awstypes.RegistryStatus]                  `tfsdk:"status"`
	Tags                   tftags.Map                                                   `tfsdk:"tags"`
	TagsAll                tftags.Map                                                   `tfsdk:"tags_all"`
	Timeouts               timeouts.Value                                               `tfsdk:"timeouts"`
}

type approvalConfigurationModel struct {
	AutoApprovalRules fwtypes.SetOfString `tfsdk:"auto_approval_rules"`
}

type discoveryConfigurationModel struct {
	AuthorizerType          fwtypes.StringEnum[awstypes.RegistryAuthorizerType]           `tfsdk:"authorizer_type"`
	AuthorizerConfiguration fwtypes.ListNestedObjectValueOf[authorizerConfigurationModel] `tfsdk:"authorizer_configuration"`
}

type authorizerConfigurationModel struct {
	AllowedAudience fwtypes.ListOfString                                                `tfsdk:"allowed_audience"`
	AllowedClients  fwtypes.ListOfString                                                `tfsdk:"allowed_clients"`
	AllowedScopes   fwtypes.ListOfString                                                `tfsdk:"allowed_scopes"`
	CustomClaim     fwtypes.SetNestedObjectValueOf[customJWTAuthorizerCustomClaimModel] `tfsdk:"custom_claim"`
	DiscoveryURL    types.String                                                        `tfsdk:"discovery_url"`
}

type customJWTAuthorizerCustomClaimModel struct {
	InboundTokenClaimName      types.String                                                                        `tfsdk:"inbound_token_claim_name"`
	InboundTokenClaimValueType fwtypes.StringEnum[awstypes.InboundTokenClaimValueType]                             `tfsdk:"inbound_token_claim_value_type"`
	AuthorizingClaimMatchValue fwtypes.ListNestedObjectValueOf[customJWTAuthorizerAuthorizingClaimMatchValueModel] `tfsdk:"authorizing_claim_match_value"`
}

type customJWTAuthorizerAuthorizingClaimMatchValueModel struct {
	ClaimMatchOperator fwtypes.StringEnum[awstypes.ClaimMatchOperatorType]                      `tfsdk:"claim_match_operator"`
	ClaimMatchValue    fwtypes.ListNestedObjectValueOf[customJWTAuthorizerClaimMatchValueModel] `tfsdk:"claim_match_value"`
}

type customJWTAuthorizerClaimMatchValueModel struct {
	MatchValueString     types.String        `tfsdk:"match_value_string"`
	MatchValueStringList fwtypes.SetOfString `tfsdk:"match_value_string_list"`
}

var (
	_ fwflex.Expander  = customJWTAuthorizerClaimMatchValueModel{}
	_ fwflex.Flattener = &customJWTAuthorizerClaimMatchValueModel{}
)

func (m *customJWTAuthorizerClaimMatchValueModel) Flatten(ctx context.Context, v any) diag.Diagnostics {
	var diags diag.Diagnostics
	switch t := v.(type) {
	case awstypes.ClaimMatchValueTypeMemberMatchValueString:
		m.MatchValueString = types.StringValue(t.Value)
	case awstypes.ClaimMatchValueTypeMemberMatchValueStringList:
		m.MatchValueStringList = fwflex.FlattenFrameworkStringValueSetOfString(ctx, t.Value)

	default:
		diags.AddError(
			"Unsupported Type",
			fmt.Sprintf("claim match value flatten: %T", v),
		)
	}
	return diags
}

func (m customJWTAuthorizerClaimMatchValueModel) Expand(ctx context.Context) (any, diag.Diagnostics) {
	var diags diag.Diagnostics
	switch {
	case !m.MatchValueString.IsNull():
		var r awstypes.ClaimMatchValueTypeMemberMatchValueString
		r.Value = fwflex.StringValueFromFramework(ctx, m.MatchValueString)
		return &r, diags
	case !m.MatchValueStringList.IsNull():
		var r awstypes.ClaimMatchValueTypeMemberMatchValueStringList
		r.Value = fwflex.ExpandFrameworkStringValueSet(ctx, m.MatchValueStringList)
		return &r, diags
	}
	return nil, diags
}
