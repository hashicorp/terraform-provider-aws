// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

// DONOTCOPY: Copying old resources spreads bad habits. Use skaff instead.

package bedrockagentcore

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/YakDriver/regexache"
	"github.com/YakDriver/smarterr"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/bedrockagentcorecontrol"
	awstypes "github.com/aws/aws-sdk-go-v2/service/bedrockagentcorecontrol/types"
	"github.com/aws/aws-sdk-go-v2/service/cloudwatchlogs"
	logstypes "github.com/aws/aws-sdk-go-v2/service/cloudwatchlogs/types"
	"github.com/aws/aws-sdk-go-v2/service/xray"
	xraytypes "github.com/aws/aws-sdk-go-v2/service/xray/types"
	"github.com/hashicorp/aws-sdk-go-base/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework-validators/int32validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/objectvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/setvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/listplanmodifier"
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

// @FrameworkResource("aws_bedrockagentcore_agent_runtime", name="Agent Runtime")
// @Tags(identifierAttribute="agent_runtime_arn")
// @Testing(tagsTest=false)
func newAgentRuntimeResource(_ context.Context) (resource.ResourceWithConfigure, error) {
	r := &agentRuntimeResource{}

	r.SetDefaultCreateTimeout(30 * time.Minute)
	r.SetDefaultUpdateTimeout(30 * time.Minute)
	r.SetDefaultDeleteTimeout(30 * time.Minute)

	return r, nil
}

type agentRuntimeResource struct {
	framework.ResourceWithModel[agentRuntimeResourceModel]
	framework.WithImportByIdentity
	framework.WithTimeouts
}

func (r *agentRuntimeResource) Schema(ctx context.Context, request resource.SchemaRequest, response *resource.SchemaResponse) {
	response.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"agent_runtime_arn": framework.ARNAttributeComputedOnly(),
			"agent_runtime_id":  framework.IDAttribute(),
			"agent_runtime_name": schema.StringAttribute{
				Required: true,
				Validators: []validator.String{
					stringvalidator.RegexMatches(regexache.MustCompile(`^[a-zA-Z][a-zA-Z0-9_]{0,47}$`), ""),
				},
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"agent_runtime_version": schema.StringAttribute{
				Computed: true,
			},
			names.AttrDescription: schema.StringAttribute{
				Optional: true,
				Validators: []validator.String{
					stringvalidator.LengthBetween(1, 4096),
				},
			},
			"environment_variables": schema.MapAttribute{
				CustomType: fwtypes.MapOfStringType,
				Optional:   true,
			},
			"lifecycle_configuration": framework.ResourceOptionalComputedListOfObjectsAttribute[lifecycleConfigurationModel](ctx, 1, nil, listplanmodifier.UseStateForUnknown()),
			names.AttrRoleARN: schema.StringAttribute{
				CustomType: fwtypes.ARNType,
				Required:   true,
			},
			names.AttrTags:              tftags.TagsAttribute(),
			names.AttrTagsAll:           tftags.TagsAttributeComputedOnly(),
			"workload_identity_details": framework.ResourceComputedListOfObjectsAttribute[workloadIdentityDetailsModel](ctx, listplanmodifier.UseStateForUnknown()),
		},

		Blocks: map[string]schema.Block{
			"agent_runtime_artifact": schema.ListNestedBlock{
				CustomType: fwtypes.NewListNestedObjectTypeOf[agentRuntimeArtifactModel](ctx),
				Validators: []validator.List{
					listvalidator.IsRequired(),
					listvalidator.SizeAtMost(1),
				},
				PlanModifiers: []planmodifier.List{
					listplanmodifier.RequiresReplaceIf(
						func(ctx context.Context, request planmodifier.ListRequest, response *listplanmodifier.RequiresReplaceIfFuncResponse) {
							// If code_configuration was set in the previous configuration and container_configuration is set in the planned configuration, a replacement is required — and vice versa.
							var prev, plan agentRuntimeArtifactModel
							smerr.AddEnrich(ctx, &response.Diagnostics, request.State.GetAttribute(ctx, path.Root("agent_runtime_artifact").AtListIndex(0), &prev))
							smerr.AddEnrich(ctx, &response.Diagnostics, request.Plan.GetAttribute(ctx, path.Root("agent_runtime_artifact").AtListIndex(0), &plan))
							if response.Diagnostics.HasError() {
								return
							}
							if (!prev.ContainerConfiguration.IsNull() && !plan.CodeConfiguration.IsNull()) ||
								(!prev.CodeConfiguration.IsNull() && !plan.ContainerConfiguration.IsNull()) {
								response.RequiresReplace = true
							}
						},
						"Artifact type change between code_configuration and container_configuration requires replacement",
						"",
					),
				},
				NestedObject: schema.NestedBlockObject{
					Blocks: map[string]schema.Block{
						"code_configuration": schema.ListNestedBlock{
							CustomType: fwtypes.NewListNestedObjectTypeOf[codeConfigurationModel](ctx),
							Validators: []validator.List{
								listvalidator.SizeAtMost(1),
								listvalidator.ExactlyOneOf(
									path.MatchRelative().AtParent().AtName("container_configuration"),
									path.MatchRelative().AtParent().AtName("code_configuration"),
								),
							},
							NestedObject: schema.NestedBlockObject{
								Attributes: map[string]schema.Attribute{
									"entry_point": schema.ListAttribute{
										CustomType: fwtypes.ListOfStringType,
										Required:   true,
										Validators: []validator.List{
											listvalidator.SizeAtLeast(1),
											listvalidator.SizeAtMost(2),
											listvalidator.ValueStringsAre(stringvalidator.LengthBetween(1, 128)),
										},
										ElementType: types.StringType,
									},
									"runtime": schema.StringAttribute{
										Required:   true,
										CustomType: fwtypes.StringEnumType[awstypes.AgentManagedRuntimeType](),
									},
								},
								Blocks: map[string]schema.Block{
									"code": schema.ListNestedBlock{
										CustomType: fwtypes.NewListNestedObjectTypeOf[codeConfigurationCodeModel](ctx),
										Validators: []validator.List{
											listvalidator.SizeAtMost(1),
										},
										NestedObject: schema.NestedBlockObject{
											Blocks: map[string]schema.Block{
												"s3": schema.ListNestedBlock{
													CustomType: fwtypes.NewListNestedObjectTypeOf[s3CodeConfigurationModel](ctx),
													Validators: []validator.List{
														listvalidator.SizeAtMost(1),
														listvalidator.ExactlyOneOf(
															// If another member is added to the union, this will need to be updated.
															path.MatchRelative().AtParent().AtName("s3"),
														),
													},
													NestedObject: schema.NestedBlockObject{
														Attributes: map[string]schema.Attribute{
															names.AttrBucket: schema.StringAttribute{
																Required: true,
																Validators: []validator.String{
																	stringvalidator.RegexMatches(regexache.MustCompile(`^[a-z0-9][a-z0-9.-]{1,61}[a-z0-9]$`), "must be a valid S3 bucket name"),
																},
															},
															names.AttrPrefix: schema.StringAttribute{
																Required: true,
																Validators: []validator.String{
																	stringvalidator.LengthBetween(1, 1024),
																},
															},
															"version_id": schema.StringAttribute{
																Optional: true,
																Validators: []validator.String{
																	stringvalidator.LengthBetween(3, 1024),
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
						"container_configuration": schema.ListNestedBlock{
							CustomType: fwtypes.NewListNestedObjectTypeOf[containerConfigurationModel](ctx),
							Validators: []validator.List{
								listvalidator.SizeAtMost(1),
								listvalidator.ExactlyOneOf(
									path.MatchRelative().AtParent().AtName("container_configuration"),
									path.MatchRelative().AtParent().AtName("code_configuration"),
								),
							},
							NestedObject: schema.NestedBlockObject{
								Attributes: map[string]schema.Attribute{
									"container_uri": schema.StringAttribute{
										Required: true,
									},
								},
							},
						},
					},
				},
			},
			"authorizer_configuration": authorizerConfigurationSchema(ctx),
			names.AttrNetworkConfiguration: schema.ListNestedBlock{
				CustomType: fwtypes.NewListNestedObjectTypeOf[networkConfigurationModel](ctx),
				Validators: []validator.List{
					listvalidator.IsRequired(),
					listvalidator.SizeAtMost(1),
				},
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"network_mode": schema.StringAttribute{
							CustomType: fwtypes.StringEnumType[awstypes.NetworkMode](),
							Required:   true,
						},
					},
					Blocks: map[string]schema.Block{
						"network_mode_config": schema.ListNestedBlock{
							CustomType: fwtypes.NewListNestedObjectTypeOf[vpcConfigModel](ctx),
							Validators: []validator.List{
								listvalidator.SizeAtMost(1),
							},
							NestedObject: schema.NestedBlockObject{
								Attributes: map[string]schema.Attribute{
									names.AttrSecurityGroups: schema.SetAttribute{
										CustomType: fwtypes.SetOfStringType,
										Required:   true,
									},
									names.AttrSubnets: schema.SetAttribute{
										CustomType: fwtypes.SetOfStringType,
										Required:   true,
									},
								},
							},
						},
					},
				},
			},
			"protocol_configuration": schema.ListNestedBlock{
				CustomType: fwtypes.NewListNestedObjectTypeOf[protocolConfigurationModel](ctx),
				Validators: []validator.List{
					listvalidator.SizeAtMost(1),
				},
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"server_protocol": schema.StringAttribute{
							Optional:   true,
							CustomType: fwtypes.StringEnumType[awstypes.ServerProtocol](),
						},
					},
				},
			},
			"request_header_configuration": schema.ListNestedBlock{
				CustomType: fwtypes.NewListNestedObjectTypeOf[requestHeaderConfigurationModel](ctx),
				Validators: []validator.List{
					listvalidator.SizeAtMost(1),
				},
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"request_header_allowlist": schema.SetAttribute{
							CustomType: fwtypes.SetOfStringType,
							Optional:   true,
						},
					},
				},
			},
			"observability": schema.ListNestedBlock{
				CustomType: fwtypes.NewListNestedObjectTypeOf[observabilityConfigurationModel](ctx),
				Validators: []validator.List{
					listvalidator.SizeAtMost(1),
				},
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"enabled": schema.BoolAttribute{
							Required: true,
						},
						"runtime_language": schema.StringAttribute{
							Optional: true,
							Validators: []validator.String{
								stringvalidator.OneOf("python"),
							},
						},
					},
					Blocks: map[string]schema.Block{
						"cloudwatch_logs": schema.ListNestedBlock{
							CustomType: fwtypes.NewListNestedObjectTypeOf[cloudwatchLogsConfigurationModel](ctx),
							Validators: []validator.List{
								listvalidator.SizeAtMost(1),
							},
							NestedObject: schema.NestedBlockObject{
								Attributes: map[string]schema.Attribute{
									"log_group_name": schema.StringAttribute{
										Optional: true,
										Computed: true,
										Validators: []validator.String{
											stringvalidator.LengthBetween(1, 512),
											stringvalidator.RegexMatches(
												regexache.MustCompile(`^[a-zA-Z0-9._/\-#]+$`),
												"must contain only alphanumeric characters, periods, hyphens, underscores, forward slashes, and hash signs",
											),
										},
									},
									"retention_in_days": schema.Int32Attribute{
										Optional: true,
										Validators: []validator.Int32{
											int32validator.OneOf(1, 3, 5, 7, 14, 30, 60, 90, 120, 150, 180, 365, 400, 545, 731, 1096, 1827, 2192, 2557, 2922, 3288, 3653),
										},
									},
								},
							},
						},
						"xray": schema.ListNestedBlock{
							CustomType: fwtypes.NewListNestedObjectTypeOf[xrayConfigurationModel](ctx),
							Validators: []validator.List{
								listvalidator.SizeAtMost(1),
							},
							NestedObject: schema.NestedBlockObject{
								Attributes: map[string]schema.Attribute{
									"sampling_percentage": schema.Int32Attribute{
										Optional: true,
										Validators: []validator.Int32{
											int32validator.Between(0, 100),
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

// Note that this function and the models used within it are also used in gateway.go.
func authorizerConfigurationSchema(ctx context.Context) schema.ListNestedBlock {
	return schema.ListNestedBlock{
		CustomType: fwtypes.NewListNestedObjectTypeOf[authorizerConfigurationModel](ctx),
		Validators: []validator.List{
			listvalidator.SizeAtMost(1),
		},
		NestedObject: schema.NestedBlockObject{
			Blocks: map[string]schema.Block{
				"custom_jwt_authorizer": schema.ListNestedBlock{
					CustomType: fwtypes.NewListNestedObjectTypeOf[customJWTAuthorizerConfigurationModel](ctx),
					Validators: []validator.List{
						listvalidator.SizeAtMost(1),
					},
					NestedObject: schema.NestedBlockObject{
						Attributes: map[string]schema.Attribute{
							"allowed_audience": schema.SetAttribute{
								CustomType: fwtypes.SetOfStringType,
								Optional:   true,
							},
							"allowed_clients": schema.SetAttribute{
								CustomType: fwtypes.SetOfStringType,
								Optional:   true,
							},
							"allowed_scopes": schema.SetAttribute{
								CustomType: fwtypes.SetOfStringType,
								Optional:   true,
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
												stringvalidator.RegexMatches(regexache.MustCompile(`^[A-Za-z0-9_.-:]+$`), "must contain only letters, numbers, and the characters _ . - :"),
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
																		stringvalidator.RegexMatches(regexache.MustCompile(`^[A-Za-z0-9_.-]+$`), "must contain only letters, numbers, and the characters _ . -"),
																	},
																},
																"match_value_string_list": schema.SetAttribute{
																	Optional:    true,
																	ElementType: types.StringType,
																	Validators: []validator.Set{
																		setvalidator.ValueStringsAre(
																			stringvalidator.LengthBetween(1, 255),
																			stringvalidator.RegexMatches(regexache.MustCompile(`^[A-Za-z0-9_.-]+$`), "must contain only letters, numbers, and the characters _ . -"),
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
	}
}

func (r *agentRuntimeResource) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	var data agentRuntimeResourceModel
	smerr.AddEnrich(ctx, &response.Diagnostics, request.Plan.Get(ctx, &data))
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().BedrockAgentCoreClient(ctx)

	var input bedrockagentcorecontrol.CreateAgentRuntimeInput
	smerr.AddEnrich(ctx, &response.Diagnostics, fwflex.Expand(ctx, data, &input, fwflex.WithFieldNamePrefix("AgentRuntime")))
	if response.Diagnostics.HasError() {
		return
	}

	// Additional fields.
	input.ClientToken = aws.String(create.UniqueId(ctx))
	input.Tags = getTagsIn(ctx)

	var (
		out *bedrockagentcorecontrol.CreateAgentRuntimeOutput
		err error
	)
	err = tfresource.Retry(ctx, propagationTimeout, func(ctx context.Context) *tfresource.RetryError {
		out, err = conn.CreateAgentRuntime(ctx, &input)

		// IAM propagation.
		if tfawserr.ErrMessageContains(err, errCodeValidationException, "Role validation failed") {
			return tfresource.RetryableError(err)
		}
		if tfawserr.ErrMessageContains(err, errCodeValidationException, "Access denied while validating ECR URI") {
			return tfresource.RetryableError(err)
		}

		if err != nil {
			return tfresource.NonRetryableError(err)
		}

		return nil
	})
	if err != nil {
		smerr.AddError(ctx, &response.Diagnostics, err, smerr.ID, data.AgentRuntimeName.String())
		return
	}

	agentRuntimeID := aws.ToString(out.AgentRuntimeId)

	if _, err := waitAgentRuntimeCreated(ctx, conn, agentRuntimeID, r.CreateTimeout(ctx, data.Timeouts)); err != nil {
		smerr.AddError(ctx, &response.Diagnostics, err, smerr.ID, agentRuntimeID)
		return
	}

	runtime, err := findAgentRuntimeByID(ctx, conn, agentRuntimeID)
	if err != nil {
		smerr.AddError(ctx, &response.Diagnostics, err, smerr.ID, agentRuntimeID)
		return
	}

	// Persist the created runtime to state immediately so that a subsequent destroy
	// can remove it even if post-create observability configuration fails.
	// Save with null observability to avoid unknown computed values (e.g. log_group_name).
	userEnvVars := data.EnvironmentVariables
	configObservability := data.Observability
	smerr.AddEnrich(ctx, &response.Diagnostics, fwflex.Flatten(ctx, runtime, &data, fwflex.WithFieldNamePrefix("AgentRuntime")))
	if response.Diagnostics.HasError() {
		return
	}
	data.Observability = fwtypes.NewListNestedObjectValueOfNull[observabilityConfigurationModel](ctx)
	smerr.AddEnrich(ctx, &response.Diagnostics, response.State.Set(ctx, data))
	if response.Diagnostics.HasError() {
		return
	}
	data.Observability = configObservability

	if !data.Observability.IsNull() {
		obsConfig, d := data.Observability.ToPtr(ctx)
		smerr.AddEnrich(ctx, &response.Diagnostics, d)
		if response.Diagnostics.HasError() {
			return
		}
		if obsConfig != nil && obsConfig.Enabled.ValueBool() {
			if err := r.ensureXRayResourcePolicy(ctx); err != nil {
				smerr.AddError(ctx, &response.Diagnostics, err, smerr.ID, agentRuntimeID)
				return
			}
			if err := r.waitForXRayResourcePolicy(ctx, propagationTimeout); err != nil {
				smerr.AddError(ctx, &response.Diagnostics, fmt.Errorf("waiting for X-Ray resource policy: %w", err), smerr.ID, agentRuntimeID)
				return
			}
			resolvedLogGroup, err := configureObservability(ctx, conn, r.Meta().LogsClient(ctx), r.Meta().XRayClient(ctx), r.Meta().Region(ctx), runtime, obsConfig, data.EnvironmentVariables)
			if err != nil {
				smerr.AddError(ctx, &response.Diagnostics, err, smerr.ID, agentRuntimeID)
				return
			}
			// Persist computed log_group_name when the user did not explicitly set one.
			if !obsConfig.CloudwatchLogs.IsNull() {
				cwConfig, d := obsConfig.CloudwatchLogs.ToPtr(ctx)
				smerr.AddEnrich(ctx, &response.Diagnostics, d)
				if response.Diagnostics.HasError() {
					return
				}
				if cwConfig != nil && (cwConfig.LogGroupName.IsNull() || cwConfig.LogGroupName.IsUnknown()) {
					cwConfig.LogGroupName = types.StringValue(resolvedLogGroup)
					obsConfig.CloudwatchLogs = fwtypes.NewListNestedObjectValueOfPtrMust(ctx, cwConfig)
				}
			}
			data.Observability = fwtypes.NewListNestedObjectValueOfPtrMust(ctx, obsConfig)
			runtime, err = findAgentRuntimeByID(ctx, conn, agentRuntimeID)
			if err != nil {
				smerr.AddError(ctx, &response.Diagnostics, err, smerr.ID, agentRuntimeID)
				return
			}
		}
	}

	// resolvedObservability holds the fully resolved observability block (known log_group_name).
	resolvedObservability := data.Observability

	// Re-flatten after observability to pick up any updated API fields.
	smerr.AddEnrich(ctx, &response.Diagnostics, fwflex.Flatten(ctx, runtime, &data, fwflex.WithFieldNamePrefix("AgentRuntime")))
	if response.Diagnostics.HasError() {
		return
	}
	// fwflex.Flatten operates on API fields only; restore Terraform-managed fields.
	data.Observability = resolvedObservability
	if !data.Observability.IsNull() {
		obsConfig, d := data.Observability.ToPtr(ctx)
		smerr.AddEnrich(ctx, &response.Diagnostics, d)
		if !response.Diagnostics.HasError() && obsConfig != nil && obsConfig.Enabled.ValueBool() {
			data.EnvironmentVariables = userEnvVars
		}
	}

	smerr.AddEnrich(ctx, &response.Diagnostics, response.State.Set(ctx, data))
}

func (r *agentRuntimeResource) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	var data agentRuntimeResourceModel
	smerr.AddEnrich(ctx, &response.Diagnostics, request.State.Get(ctx, &data))
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().BedrockAgentCoreClient(ctx)

	agentRuntimeID := fwflex.StringValueFromFramework(ctx, data.AgentRuntimeID)
	out, err := findAgentRuntimeByID(ctx, conn, agentRuntimeID)
	if retry.NotFound(err) {
		smerr.AddOne(ctx, &response.Diagnostics, fwdiag.NewResourceNotFoundWarningDiagnostic(err))
		response.State.RemoveResource(ctx)
		return
	}
	if err != nil {
		smerr.AddError(ctx, &response.Diagnostics, err, smerr.ID, agentRuntimeID)
		return
	}

	savedObservability := data.Observability
	userEnvVars := data.EnvironmentVariables
	smerr.AddEnrich(ctx, &response.Diagnostics, fwflex.Flatten(ctx, out, &data, fwflex.WithFieldNamePrefix("AgentRuntime")))
	if response.Diagnostics.HasError() {
		return
	}
	// fwflex.Flatten operates on API fields only; restore Terraform-managed fields.
	data.Observability = savedObservability

	if !data.Observability.IsNull() {
		obsConfig, d := data.Observability.ToPtr(ctx)
		smerr.AddEnrich(ctx, &response.Diagnostics, d)
		if response.Diagnostics.HasError() {
			return
		}
		if obsConfig != nil && obsConfig.Enabled.ValueBool() {
			data.EnvironmentVariables = userEnvVars

			// Drift detection: sync retention_in_days from CloudWatch Logs.
			if !obsConfig.CloudwatchLogs.IsNull() {
				cwConfig, d := obsConfig.CloudwatchLogs.ToPtr(ctx)
				smerr.AddEnrich(ctx, &response.Diagnostics, d)
				if response.Diagnostics.HasError() {
					return
				}
				if cwConfig != nil && !cwConfig.LogGroupName.IsNull() && !cwConfig.RetentionInDays.IsNull() {
					retention, err := readLogGroupRetention(ctx, r.Meta().LogsClient(ctx), cwConfig.LogGroupName.ValueString())
					if err != nil {
						smerr.AddError(ctx, &response.Diagnostics, err, smerr.ID, agentRuntimeID)
						return
					}
					if retention > 0 {
						cwConfig.RetentionInDays = types.Int32Value(retention)
					} else {
						cwConfig.RetentionInDays = types.Int32Null()
					}
					obsConfig.CloudwatchLogs = fwtypes.NewListNestedObjectValueOfPtrMust(ctx, cwConfig)
					data.Observability = fwtypes.NewListNestedObjectValueOfPtrMust(ctx, obsConfig)
				}
			}

			// Drift detection: sync sampling_percentage from X-Ray.
			if !obsConfig.Xray.IsNull() {
				xrayConfig, d := obsConfig.Xray.ToPtr(ctx)
				smerr.AddEnrich(ctx, &response.Diagnostics, d)
				if response.Diagnostics.HasError() {
					return
				}
				if xrayConfig != nil && !xrayConfig.SamplingPercentage.IsNull() {
					percentage, found, err := readXRaySamplingPercentage(ctx, r.Meta().XRayClient(ctx), agentRuntimeID)
					if err != nil {
						smerr.AddError(ctx, &response.Diagnostics, err, smerr.ID, agentRuntimeID)
						return
					}
					if found {
						xrayConfig.SamplingPercentage = types.Int32Value(percentage)
					} else {
						xrayConfig.SamplingPercentage = types.Int32Null()
					}
					obsConfig.Xray = fwtypes.NewListNestedObjectValueOfPtrMust(ctx, xrayConfig)
					data.Observability = fwtypes.NewListNestedObjectValueOfPtrMust(ctx, obsConfig)
				}
			}
		}
	} else if out.EnvironmentVariables["AGENT_OBSERVABILITY_ENABLED"] == "true" {
		// Import path: observability was enabled but no prior Terraform state exists.
		// Reconstruct the observability block from OTEL environment variables injected
		// by configureObservability.
		obsConfig, err := reconstructObservabilityFromEnvVars(ctx, r.Meta().LogsClient(ctx), r.Meta().XRayClient(ctx), out.EnvironmentVariables, agentRuntimeID)
		if err != nil {
			smerr.AddError(ctx, &response.Diagnostics, err, smerr.ID, agentRuntimeID)
			return
		}
		data.Observability = fwtypes.NewListNestedObjectValueOfPtrMust(ctx, obsConfig)

		// Strip OTEL env vars so environment_variables reflects the user-managed set only.
		filteredAttrs := make(map[string]attr.Value)
		for k, v := range data.EnvironmentVariables.Elements() {
			if !isOtelEnvVar(k) {
				filteredAttrs[k] = v
			}
		}
		data.EnvironmentVariables = fwtypes.NewMapValueOfMust[types.String](ctx, filteredAttrs)
	}

	smerr.AddEnrich(ctx, &response.Diagnostics, response.State.Set(ctx, &data))
}

func reconstructObservabilityFromEnvVars(ctx context.Context, logsConn *cloudwatchlogs.Client, xrayConn *xray.Client, apiEnvVars map[string]string, runtimeID string) (*observabilityConfigurationModel, error) {
	// Detect runtime language from language-specific env vars.
	runtimeLanguage := ""
	if _, ok := apiEnvVars["OTEL_PYTHON_DISTRO"]; ok {
		runtimeLanguage = "python"
	}

	// Extract log group from OTEL_EXPORTER_OTLP_LOGS_HEADERS.
	logGroup := ""
	if headers, ok := apiEnvVars["OTEL_EXPORTER_OTLP_LOGS_HEADERS"]; ok {
		for _, part := range strings.Split(headers, ",") {
			if strings.HasPrefix(part, "x-aws-log-group=") {
				logGroup = strings.TrimPrefix(part, "x-aws-log-group=")
				break
			}
		}
	}

	// Build cloudwatch_logs config, reading retention from the API.
	cwConfig := &cloudwatchLogsConfigurationModel{
		LogGroupName:    types.StringValue(logGroup),
		RetentionInDays: types.Int32Null(),
	}
	if logGroup != "" {
		retention, err := readLogGroupRetention(ctx, logsConn, logGroup)
		if err != nil {
			return nil, fmt.Errorf("reading log group retention during import: %w", err)
		}
		if retention > 0 {
			cwConfig.RetentionInDays = types.Int32Value(retention)
		}
	}

	// Build xray config, reading sampling percentage from the API.
	xrayConfig := &xrayConfigurationModel{SamplingPercentage: types.Int32Null()}
	percentage, found, err := readXRaySamplingPercentage(ctx, xrayConn, runtimeID)
	if err != nil {
		return nil, fmt.Errorf("reading X-Ray sampling percentage during import: %w", err)
	}
	if found {
		xrayConfig.SamplingPercentage = types.Int32Value(percentage)
	}

	return &observabilityConfigurationModel{
		Enabled:         types.BoolValue(true),
		RuntimeLanguage: types.StringValue(runtimeLanguage),
		CloudwatchLogs:  fwtypes.NewListNestedObjectValueOfPtrMust(ctx, cwConfig),
		Xray:            fwtypes.NewListNestedObjectValueOfPtrMust(ctx, xrayConfig),
	}, nil
}

func (r *agentRuntimeResource) Update(ctx context.Context, request resource.UpdateRequest, response *resource.UpdateResponse) {
	var new, old agentRuntimeResourceModel
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
		agentRuntimeID := fwflex.StringValueFromFramework(ctx, new.AgentRuntimeID)
		var input bedrockagentcorecontrol.UpdateAgentRuntimeInput
		smerr.AddEnrich(ctx, &response.Diagnostics, fwflex.Expand(ctx, new, &input, fwflex.WithFieldNamePrefix("AgentRuntime")))
		if response.Diagnostics.HasError() {
			return
		}

		// Additional fields.
		input.ClientToken = aws.String(create.UniqueId(ctx))

		_, err := conn.UpdateAgentRuntime(ctx, &input)
		if err != nil {
			smerr.AddError(ctx, &response.Diagnostics, err, smerr.ID, agentRuntimeID)
			return
		}

		if _, err := waitAgentRuntimeUpdated(ctx, conn, agentRuntimeID, r.UpdateTimeout(ctx, new.Timeouts)); err != nil {
			smerr.AddError(ctx, &response.Diagnostics, err, smerr.ID, agentRuntimeID)
			return
		}

		if !new.Observability.IsNull() {
			obsConfig, d := new.Observability.ToPtr(ctx)
			smerr.AddEnrich(ctx, &response.Diagnostics, d)
			if response.Diagnostics.HasError() {
				return
			}
			if obsConfig != nil && obsConfig.Enabled.ValueBool() {
				runtime, err := findAgentRuntimeByID(ctx, conn, agentRuntimeID)
				if err != nil {
					smerr.AddError(ctx, &response.Diagnostics, err, smerr.ID, agentRuntimeID)
					return
				}
				if err := r.ensureXRayResourcePolicy(ctx); err != nil {
					smerr.AddError(ctx, &response.Diagnostics, err, smerr.ID, agentRuntimeID)
					return
				}
				resolvedLogGroup, err := configureObservability(ctx, conn, r.Meta().LogsClient(ctx), r.Meta().XRayClient(ctx), r.Meta().Region(ctx), runtime, obsConfig, new.EnvironmentVariables)
				if err != nil {
					smerr.AddError(ctx, &response.Diagnostics, err, smerr.ID, agentRuntimeID)
					return
				}
				// Persist computed log_group_name when the user did not explicitly set one.
				if !obsConfig.CloudwatchLogs.IsNull() {
					cwConfig, d := obsConfig.CloudwatchLogs.ToPtr(ctx)
					smerr.AddEnrich(ctx, &response.Diagnostics, d)
					if response.Diagnostics.HasError() {
						return
					}
					if cwConfig != nil && (cwConfig.LogGroupName.IsNull() || cwConfig.LogGroupName.IsUnknown()) {
						cwConfig.LogGroupName = types.StringValue(resolvedLogGroup)
						obsConfig.CloudwatchLogs = fwtypes.NewListNestedObjectValueOfPtrMust(ctx, cwConfig)
					}
				}
				new.Observability = fwtypes.NewListNestedObjectValueOfPtrMust(ctx, obsConfig)
				if _, err := waitAgentRuntimeUpdated(ctx, conn, agentRuntimeID, r.UpdateTimeout(ctx, new.Timeouts)); err != nil {
					smerr.AddError(ctx, &response.Diagnostics, err, smerr.ID, agentRuntimeID)
					return
				}
			} else if obsConfig != nil && !obsConfig.Enabled.ValueBool() {
				oldObsConfig, d := old.Observability.ToPtr(ctx)
				smerr.AddEnrich(ctx, &response.Diagnostics, d)
				if response.Diagnostics.HasError() {
					return
				}
				if oldObsConfig != nil && oldObsConfig.Enabled.ValueBool() {
					runtime, err := findAgentRuntimeByID(ctx, conn, agentRuntimeID)
					if err != nil {
						smerr.AddError(ctx, &response.Diagnostics, err, smerr.ID, agentRuntimeID)
						return
					}
					if err := disableObservability(ctx, conn, r.Meta().XRayClient(ctx), runtime, agentRuntimeID); err != nil {
						smerr.AddError(ctx, &response.Diagnostics, err, smerr.ID, agentRuntimeID)
						return
					}
					if _, err := waitAgentRuntimeUpdated(ctx, conn, agentRuntimeID, r.UpdateTimeout(ctx, new.Timeouts)); err != nil {
						smerr.AddError(ctx, &response.Diagnostics, err, smerr.ID, agentRuntimeID)
						return
					}
				}
				// Resolve computed log_group_name: preserve from old state or derive from runtime ID.
				if !obsConfig.CloudwatchLogs.IsNull() {
					cwConfig, d := obsConfig.CloudwatchLogs.ToPtr(ctx)
					smerr.AddEnrich(ctx, &response.Diagnostics, d)
					if response.Diagnostics.HasError() {
						return
					}
					if cwConfig != nil && (cwConfig.LogGroupName.IsNull() || cwConfig.LogGroupName.IsUnknown()) {
						resolvedName := fmt.Sprintf("/aws/bedrock-agentcore/runtimes/%s", agentRuntimeID)
						if oldObsConfig != nil && !oldObsConfig.CloudwatchLogs.IsNull() {
							oldCwConfig, d := oldObsConfig.CloudwatchLogs.ToPtr(ctx)
							smerr.AddEnrich(ctx, &response.Diagnostics, d)
							if !response.Diagnostics.HasError() && oldCwConfig != nil && !oldCwConfig.LogGroupName.IsNull() && !oldCwConfig.LogGroupName.IsUnknown() {
								resolvedName = oldCwConfig.LogGroupName.ValueString()
							}
						}
						cwConfig.LogGroupName = types.StringValue(resolvedName)
						obsConfig.CloudwatchLogs = fwtypes.NewListNestedObjectValueOfPtrMust(ctx, cwConfig)
						new.Observability = fwtypes.NewListNestedObjectValueOfPtrMust(ctx, obsConfig)
					}
				}
			}
		}

		runtime, err := findAgentRuntimeByID(ctx, conn, agentRuntimeID)
		if err != nil {
			smerr.AddError(ctx, &response.Diagnostics, err, smerr.ID, agentRuntimeID)
			return
		}

		savedObservability := new.Observability
		userEnvVars := new.EnvironmentVariables
		smerr.AddEnrich(ctx, &response.Diagnostics, fwflex.Flatten(ctx, runtime, &new, fwflex.WithFieldNamePrefix("AgentRuntime")))
		if response.Diagnostics.HasError() {
			return
		}
		// fwflex.Flatten operates on API fields only; restore Terraform-managed fields.
		new.Observability = savedObservability
		if !new.Observability.IsNull() {
			obsConfig, d := new.Observability.ToPtr(ctx)
			smerr.AddEnrich(ctx, &response.Diagnostics, d)
			if !response.Diagnostics.HasError() && obsConfig != nil && obsConfig.Enabled.ValueBool() {
				new.EnvironmentVariables = userEnvVars
			}
		}
	} else {
		new.AgentRuntimeVersion = old.AgentRuntimeVersion
	}

	smerr.AddEnrich(ctx, &response.Diagnostics, response.State.Set(ctx, &new))
}

func (r *agentRuntimeResource) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	var data agentRuntimeResourceModel
	smerr.AddEnrich(ctx, &response.Diagnostics, request.State.Get(ctx, &data))
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().BedrockAgentCoreClient(ctx)

	agentRuntimeID := fwflex.StringValueFromFramework(ctx, data.AgentRuntimeID)
	input := bedrockagentcorecontrol.DeleteAgentRuntimeInput{
		AgentRuntimeId: aws.String(agentRuntimeID),
	}

	_, err := conn.DeleteAgentRuntime(ctx, &input)
	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return
	}
	if err != nil {
		smerr.AddError(ctx, &response.Diagnostics, err, smerr.ID, agentRuntimeID)
		return
	}

	if _, err := waitAgentRuntimeDeleted(ctx, conn, agentRuntimeID, r.DeleteTimeout(ctx, data.Timeouts)); err != nil {
		smerr.AddError(ctx, &response.Diagnostics, err, smerr.ID, agentRuntimeID)
		return
	}

	// Clean up X-Ray sampling rule if observability was configured with xray.
	if !data.Observability.IsNull() {
		obsConfig, d := data.Observability.ToPtr(ctx)
		smerr.AddEnrich(ctx, &response.Diagnostics, d)
		if !response.Diagnostics.HasError() && obsConfig != nil && !obsConfig.Xray.IsNull() {
			if err := deleteXRaySamplingRule(ctx, r.Meta().XRayClient(ctx), agentRuntimeID); err != nil {
				smerr.AddError(ctx, &response.Diagnostics, err, smerr.ID, agentRuntimeID)
				return
			}
		}
	}
}

func (r *agentRuntimeResource) ImportState(ctx context.Context, request resource.ImportStateRequest, response *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("agent_runtime_id"), request, response)
}

func waitAgentRuntimeCreated(ctx context.Context, conn *bedrockagentcorecontrol.Client, id string, timeout time.Duration) (*bedrockagentcorecontrol.GetAgentRuntimeOutput, error) {
	stateConf := &retry.StateChangeConf{
		Pending:                   enum.Slice(awstypes.AgentRuntimeStatusCreating),
		Target:                    enum.Slice(awstypes.AgentRuntimeStatusReady),
		Refresh:                   statusAgentRuntime(conn, id),
		Timeout:                   timeout,
		ContinuousTargetOccurence: 2,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)
	if out, ok := outputRaw.(*bedrockagentcorecontrol.GetAgentRuntimeOutput); ok {
		return out, smarterr.NewError(err)
	}

	return nil, smarterr.NewError(err)
}

func waitAgentRuntimeUpdated(ctx context.Context, conn *bedrockagentcorecontrol.Client, id string, timeout time.Duration) (*bedrockagentcorecontrol.GetAgentRuntimeOutput, error) {
	stateConf := &retry.StateChangeConf{
		Pending:                   enum.Slice(awstypes.AgentRuntimeStatusUpdating),
		Target:                    enum.Slice(awstypes.AgentRuntimeStatusReady),
		Refresh:                   statusAgentRuntime(conn, id),
		Timeout:                   timeout,
		ContinuousTargetOccurence: 2,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)
	if out, ok := outputRaw.(*bedrockagentcorecontrol.GetAgentRuntimeOutput); ok {
		return out, smarterr.NewError(err)
	}

	return nil, smarterr.NewError(err)
}

func waitAgentRuntimeDeleted(ctx context.Context, conn *bedrockagentcorecontrol.Client, id string, timeout time.Duration) (*bedrockagentcorecontrol.GetAgentRuntimeOutput, error) {
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(awstypes.AgentRuntimeStatusDeleting, awstypes.AgentRuntimeStatusReady),
		Target:  []string{},
		Refresh: statusAgentRuntime(conn, id),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)
	if out, ok := outputRaw.(*bedrockagentcorecontrol.GetAgentRuntimeOutput); ok {
		return out, smarterr.NewError(err)
	}

	return nil, smarterr.NewError(err)
}

func statusAgentRuntime(conn *bedrockagentcorecontrol.Client, id string) retry.StateRefreshFunc {
	return func(ctx context.Context) (any, string, error) {
		out, err := findAgentRuntimeByID(ctx, conn, id)
		if retry.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", smarterr.NewError(err)
		}

		return out, string(out.Status), nil
	}
}

func findAgentRuntimeByID(ctx context.Context, conn *bedrockagentcorecontrol.Client, id string) (*bedrockagentcorecontrol.GetAgentRuntimeOutput, error) {
	input := bedrockagentcorecontrol.GetAgentRuntimeInput{
		AgentRuntimeId: aws.String(id),
	}

	return findAgentRuntime(ctx, conn, &input)
}

func findAgentRuntime(ctx context.Context, conn *bedrockagentcorecontrol.Client, input *bedrockagentcorecontrol.GetAgentRuntimeInput) (*bedrockagentcorecontrol.GetAgentRuntimeOutput, error) {
	out, err := conn.GetAgentRuntime(ctx, input)

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

type agentRuntimeResourceModel struct {
	framework.WithRegionModel
	AgentRuntimeARN            types.String                                                     `tfsdk:"agent_runtime_arn"`
	AgentRuntimeArtifact       fwtypes.ListNestedObjectValueOf[agentRuntimeArtifactModel]       `tfsdk:"agent_runtime_artifact"`
	AgentRuntimeID             types.String                                                     `tfsdk:"agent_runtime_id"`
	AgentRuntimeName           types.String                                                     `tfsdk:"agent_runtime_name"`
	AgentRuntimeVersion        types.String                                                     `tfsdk:"agent_runtime_version"`
	AuthorizerConfiguration    fwtypes.ListNestedObjectValueOf[authorizerConfigurationModel]    `tfsdk:"authorizer_configuration"`
	Description                types.String                                                     `tfsdk:"description"`
	EnvironmentVariables       fwtypes.MapOfString                                              `tfsdk:"environment_variables"`
	LifecycleConfiguration     fwtypes.ListNestedObjectValueOf[lifecycleConfigurationModel]     `tfsdk:"lifecycle_configuration"`
	NetworkConfiguration       fwtypes.ListNestedObjectValueOf[networkConfigurationModel]       `tfsdk:"network_configuration"`
	Observability              fwtypes.ListNestedObjectValueOf[observabilityConfigurationModel] `tfsdk:"observability"`
	ProtocolConfiguration      fwtypes.ListNestedObjectValueOf[protocolConfigurationModel]      `tfsdk:"protocol_configuration"`
	RequestHeaderConfiguration fwtypes.ListNestedObjectValueOf[requestHeaderConfigurationModel] `tfsdk:"request_header_configuration"`
	RoleARN                    fwtypes.ARN                                                      `tfsdk:"role_arn"`
	Tags                       tftags.Map                                                       `tfsdk:"tags"`
	TagsAll                    tftags.Map                                                       `tfsdk:"tags_all"`
	Timeouts                   timeouts.Value                                                   `tfsdk:"timeouts"`
	WorkloadIdentityDetails    fwtypes.ListNestedObjectValueOf[workloadIdentityDetailsModel]    `tfsdk:"workload_identity_details"`
}

type agentRuntimeArtifactModel struct {
	CodeConfiguration      fwtypes.ListNestedObjectValueOf[codeConfigurationModel]      `tfsdk:"code_configuration"`
	ContainerConfiguration fwtypes.ListNestedObjectValueOf[containerConfigurationModel] `tfsdk:"container_configuration"`
}

var (
	_ fwflex.Expander  = agentRuntimeArtifactModel{}
	_ fwflex.Flattener = &agentRuntimeArtifactModel{}
)

func (m *agentRuntimeArtifactModel) Flatten(ctx context.Context, v any) diag.Diagnostics {
	var diags diag.Diagnostics
	switch t := v.(type) {
	case awstypes.AgentRuntimeArtifactMemberCodeConfiguration:
		var data codeConfigurationModel
		smerr.AddEnrich(ctx, &diags, fwflex.Flatten(ctx, t.Value, &data))
		if diags.HasError() {
			return diags
		}
		m.CodeConfiguration = fwtypes.NewListNestedObjectValueOfPtrMust(ctx, &data)
	case awstypes.AgentRuntimeArtifactMemberContainerConfiguration:
		var data containerConfigurationModel
		smerr.AddEnrich(ctx, &diags, fwflex.Flatten(ctx, t.Value, &data))
		if diags.HasError() {
			return diags
		}
		m.ContainerConfiguration = fwtypes.NewListNestedObjectValueOfPtrMust(ctx, &data)

	default:
		diags.AddError(
			"Unsupported Type",
			fmt.Sprintf("artifact flatten: %T", v),
		)
	}
	return diags
}

func (m agentRuntimeArtifactModel) Expand(ctx context.Context) (any, diag.Diagnostics) {
	var diags diag.Diagnostics
	switch {
	case !m.CodeConfiguration.IsNull():
		data, d := m.CodeConfiguration.ToPtr(ctx)
		smerr.AddEnrich(ctx, &diags, d)
		if diags.HasError() {
			return nil, diags
		}
		var r awstypes.AgentRuntimeArtifactMemberCodeConfiguration
		smerr.AddEnrich(ctx, &diags, fwflex.Expand(ctx, data, &r.Value))
		if diags.HasError() {
			return nil, diags
		}
		return &r, diags
	case !m.ContainerConfiguration.IsNull():
		data, d := m.ContainerConfiguration.ToPtr(ctx)
		smerr.AddEnrich(ctx, &diags, d)
		if diags.HasError() {
			return nil, diags
		}
		var r awstypes.AgentRuntimeArtifactMemberContainerConfiguration
		smerr.AddEnrich(ctx, &diags, fwflex.Expand(ctx, data, &r.Value))
		if diags.HasError() {
			return nil, diags
		}
		return &r, diags
	}
	return nil, diags
}

type codeConfigurationModel struct {
	Code       fwtypes.ListNestedObjectValueOf[codeConfigurationCodeModel] `tfsdk:"code"`
	EntryPoint fwtypes.ListOfString                                        `tfsdk:"entry_point"`
	Runtime    fwtypes.StringEnum[awstypes.AgentManagedRuntimeType]        `tfsdk:"runtime"`
}

type codeConfigurationCodeModel struct {
	S3 fwtypes.ListNestedObjectValueOf[s3CodeConfigurationModel] `tfsdk:"s3"`
}

type s3CodeConfigurationModel struct {
	Bucket    types.String `tfsdk:"bucket"`
	Prefix    types.String `tfsdk:"prefix"`
	VersionID types.String `tfsdk:"version_id"`
}

var (
	_ fwflex.Expander  = codeConfigurationCodeModel{}
	_ fwflex.Flattener = &codeConfigurationCodeModel{}
)

func (m *codeConfigurationCodeModel) Flatten(ctx context.Context, v any) diag.Diagnostics {
	var diags diag.Diagnostics
	switch t := v.(type) {
	case awstypes.CodeMemberS3:
		var data s3CodeConfigurationModel
		smerr.AddEnrich(ctx, &diags, fwflex.Flatten(ctx, t.Value, &data))
		if diags.HasError() {
			return diags
		}
		m.S3 = fwtypes.NewListNestedObjectValueOfPtrMust(ctx, &data)

	default:
		diags.AddError(
			"Unsupported Type",
			fmt.Sprintf("code configuration code flatten: %T", v),
		)
	}
	return diags
}

func (m codeConfigurationCodeModel) Expand(ctx context.Context) (any, diag.Diagnostics) {
	var diags diag.Diagnostics
	switch {
	case !m.S3.IsNull():
		data, d := m.S3.ToPtr(ctx)
		smerr.AddEnrich(ctx, &diags, d)
		if diags.HasError() {
			return nil, diags
		}
		var r awstypes.CodeMemberS3
		smerr.AddEnrich(ctx, &diags, fwflex.Expand(ctx, data, &r.Value))
		if diags.HasError() {
			return nil, diags
		}
		return &r, diags
	}
	return nil, diags
}

type containerConfigurationModel struct {
	ContainerURI types.String `tfsdk:"container_uri"`
}

type authorizerConfigurationModel struct {
	CustomJWTAuthorizer fwtypes.ListNestedObjectValueOf[customJWTAuthorizerConfigurationModel] `tfsdk:"custom_jwt_authorizer"`
}

var (
	_ fwflex.Expander  = authorizerConfigurationModel{}
	_ fwflex.Flattener = &authorizerConfigurationModel{}
)

func (m *authorizerConfigurationModel) Flatten(ctx context.Context, v any) diag.Diagnostics {
	var diags diag.Diagnostics
	switch t := v.(type) {
	case awstypes.AuthorizerConfigurationMemberCustomJWTAuthorizer:
		var data customJWTAuthorizerConfigurationModel
		smerr.AddEnrich(ctx, &diags, fwflex.Flatten(ctx, t.Value, &data))
		if diags.HasError() {
			return diags
		}
		m.CustomJWTAuthorizer = fwtypes.NewListNestedObjectValueOfPtrMust(ctx, &data)

	default:
		diags.AddError(
			"Unsupported Type",
			fmt.Sprintf("authorization configuration flatten: %T", v),
		)
	}
	return diags
}

func (m authorizerConfigurationModel) Expand(ctx context.Context) (any, diag.Diagnostics) {
	var diags diag.Diagnostics
	switch {
	case !m.CustomJWTAuthorizer.IsNull():
		data, d := m.CustomJWTAuthorizer.ToPtr(ctx)
		smerr.AddEnrich(ctx, &diags, d)
		if diags.HasError() {
			return nil, diags
		}
		var r awstypes.AuthorizerConfigurationMemberCustomJWTAuthorizer
		smerr.AddEnrich(ctx, &diags, fwflex.Expand(ctx, data, &r.Value))
		if diags.HasError() {
			return nil, diags
		}
		return &r, diags
	}
	return nil, diags
}

type customJWTAuthorizerConfigurationModel struct {
	AllowedAudience fwtypes.SetOfString                                                 `tfsdk:"allowed_audience"`
	AllowedClients  fwtypes.SetOfString                                                 `tfsdk:"allowed_clients"`
	AllowedScopes   fwtypes.SetOfString                                                 `tfsdk:"allowed_scopes"`
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

type lifecycleConfigurationModel struct {
	IdleRuntimeSessionTimeout types.Int32 `tfsdk:"idle_runtime_session_timeout"`
	MaxLifetime               types.Int32 `tfsdk:"max_lifetime"`
}

type networkConfigurationModel struct {
	NetworkMode       fwtypes.StringEnum[awstypes.NetworkMode]        `tfsdk:"network_mode"`
	NetworkModeConfig fwtypes.ListNestedObjectValueOf[vpcConfigModel] `tfsdk:"network_mode_config"`
}

type vpcConfigModel struct {
	SecurityGroups fwtypes.SetOfString `tfsdk:"security_groups"`
	Subnets        fwtypes.SetOfString `tfsdk:"subnets"`
}

type protocolConfigurationModel struct {
	ServerProtocol fwtypes.StringEnum[awstypes.ServerProtocol] `tfsdk:"server_protocol"`
}

type requestHeaderConfigurationModel struct {
	RequestHeaderAllowlist fwtypes.SetOfString `tfsdk:"request_header_allowlist"`
}

var (
	_ fwflex.Expander  = requestHeaderConfigurationModel{}
	_ fwflex.Flattener = &requestHeaderConfigurationModel{}
)

func (m *requestHeaderConfigurationModel) Flatten(ctx context.Context, v any) diag.Diagnostics {
	var diags diag.Diagnostics
	switch t := v.(type) {
	case awstypes.RequestHeaderConfigurationMemberRequestHeaderAllowlist:
		m.RequestHeaderAllowlist = fwflex.FlattenFrameworkStringValueSetOfString(ctx, t.Value)

	default:
		diags.AddError(
			"Unsupported Type",
			fmt.Sprintf("artifact flatten: %T", v),
		)
	}
	return diags
}

func (m requestHeaderConfigurationModel) Expand(ctx context.Context) (any, diag.Diagnostics) {
	var diags diag.Diagnostics
	switch {
	case !m.RequestHeaderAllowlist.IsNull():
		var diags diag.Diagnostics
		var r awstypes.RequestHeaderConfigurationMemberRequestHeaderAllowlist
		r.Value = fwflex.ExpandFrameworkStringValueSet(ctx, m.RequestHeaderAllowlist)
		return &r, diags
	}
	return nil, diags
}

type workloadIdentityDetailsModel struct {
	WorkloadIdentityARN types.String `tfsdk:"workload_identity_arn"`
}

type observabilityConfigurationModel struct {
	CloudwatchLogs  fwtypes.ListNestedObjectValueOf[cloudwatchLogsConfigurationModel] `tfsdk:"cloudwatch_logs"`
	Enabled         types.Bool                                                        `tfsdk:"enabled"`
	RuntimeLanguage types.String                                                      `tfsdk:"runtime_language"`
	Xray            fwtypes.ListNestedObjectValueOf[xrayConfigurationModel]           `tfsdk:"xray"`
}

type xrayConfigurationModel struct {
	SamplingPercentage types.Int32 `tfsdk:"sampling_percentage"`
}

type cloudwatchLogsConfigurationModel struct {
	LogGroupName    types.String `tfsdk:"log_group_name"`
	RetentionInDays types.Int32  `tfsdk:"retention_in_days"`
}

const xrayResourcePolicyName = "BedrockAgentCoreXRayPolicy"

// configureObservability injects OTEL environment variables, configures X-Ray, and applies
// any CloudWatch Logs settings. It returns the resolved log group name so callers can
// persist it as a computed value in state.
func configureObservability(ctx context.Context, conn *bedrockagentcorecontrol.Client, logsConn *cloudwatchlogs.Client, xrayConn *xray.Client, region string, runtime *bedrockagentcorecontrol.GetAgentRuntimeOutput, obsConfig *observabilityConfigurationModel, existingEnvVars fwtypes.MapOfString) (string, error) {
	runtimeID := aws.ToString(runtime.AgentRuntimeId)
	runtimeName := aws.ToString(runtime.AgentRuntimeName)

	// Resolve log group name: use user-provided value if set, otherwise derive from runtime ID.
	logGroup := fmt.Sprintf("/aws/bedrock-agentcore/runtimes/%s", runtimeID)
	var retentionInDays *int32

	if !obsConfig.CloudwatchLogs.IsNull() {
		cwConfig, d := obsConfig.CloudwatchLogs.ToPtr(ctx)
		if d.HasError() {
			return "", fmt.Errorf("reading cloudwatch_logs configuration: %s", d)
		}
		if cwConfig != nil {
			if !cwConfig.LogGroupName.IsNull() && !cwConfig.LogGroupName.IsUnknown() {
				logGroup = cwConfig.LogGroupName.ValueString()
			}
			if !cwConfig.RetentionInDays.IsNull() {
				v := cwConfig.RetentionInDays.ValueInt32()
				retentionInDays = &v
			}
		}
	}

	obsEnvVars := map[string]string{
		"AGENT_OBSERVABILITY_ENABLED":        "true",
		"OTEL_EXPORTER_OTLP_LOGS_ENDPOINT":   fmt.Sprintf("https://logs.%s.amazonaws.com/v1/logs", region),
		"OTEL_EXPORTER_OTLP_LOGS_HEADERS":    fmt.Sprintf("x-aws-log-group=%s,x-aws-log-stream=runtime-logs,x-aws-metric-namespace=bedrock-agentcore", logGroup),
		"OTEL_EXPORTER_OTLP_PROTOCOL":        "http/protobuf",
		"OTEL_EXPORTER_OTLP_TRACES_ENDPOINT": fmt.Sprintf("https://xray.%s.amazonaws.com/v1/traces", region),
		"OTEL_RESOURCE_ATTRIBUTES":           fmt.Sprintf("service.name=%s,aws.log.group.names=%s", runtimeName, logGroup),
		"OTEL_TRACES_EXPORTER":               "otlp",
	}

	// Inject language-specific OTEL env vars.
	switch obsConfig.RuntimeLanguage.ValueString() {
	case "python":
		obsEnvVars["OTEL_PYTHON_DISTRO"] = "aws_distro"
		obsEnvVars["OTEL_PYTHON_CONFIGURATOR"] = "aws_configurator"
		obsEnvVars["OTEL_PYTHON_LOGGING_AUTO_INSTRUMENTATION_ENABLED"] = "true"
	}

	mergedEnvVars := make(map[string]string)
	if !existingEnvVars.IsNull() {
		for k, v := range existingEnvVars.Elements() {
			if strVal, ok := v.(types.String); ok {
				mergedEnvVars[k] = strVal.ValueString()
			}
		}
	}
	for k, v := range obsEnvVars {
		mergedEnvVars[k] = v
	}

	updateInput := &bedrockagentcorecontrol.UpdateAgentRuntimeInput{
		AgentRuntimeId:          aws.String(runtimeID),
		AgentRuntimeArtifact:    runtime.AgentRuntimeArtifact,
		AuthorizerConfiguration: runtime.AuthorizerConfiguration,
		Description:             runtime.Description,
		EnvironmentVariables:    mergedEnvVars,
		NetworkConfiguration:    runtime.NetworkConfiguration,
		ProtocolConfiguration:   runtime.ProtocolConfiguration,
		RoleArn:                 runtime.RoleArn,
	}

	if _, err := conn.UpdateAgentRuntime(ctx, updateInput); err != nil {
		return "", fmt.Errorf("updating agent runtime with observability environment variables: %w", err)
	}

	if err := configureXRayForObservability(ctx, xrayConn); err != nil {
		return "", fmt.Errorf("configuring X-Ray for observability: %w", err)
	}

	if !obsConfig.Xray.IsNull() {
		xrayConfig, d := obsConfig.Xray.ToPtr(ctx)
		if d.HasError() {
			return "", fmt.Errorf("reading xray configuration: %s", d)
		}
		if xrayConfig != nil && !xrayConfig.SamplingPercentage.IsNull() {
			if err := applyXRaySamplingRule(ctx, xrayConn, runtimeID, xrayConfig.SamplingPercentage.ValueInt32()); err != nil {
				return "", fmt.Errorf("configuring X-Ray sampling rule: %w", err)
			}
		}
	}

	if retentionInDays != nil {
		if err := applyLogGroupRetention(ctx, logsConn, logGroup, *retentionInDays); err != nil {
			return "", err
		}
	}

	return logGroup, nil
}

func configureXRayForObservability(ctx context.Context, xrayConn *xray.Client) error {
	out, err := xrayConn.GetTraceSegmentDestination(ctx, &xray.GetTraceSegmentDestinationInput{})
	if err != nil {
		return fmt.Errorf("getting X-Ray trace segment destination: %w", err)
	}

	if out.Destination != xraytypes.TraceSegmentDestinationCloudWatchLogs {
		if _, err := xrayConn.UpdateTraceSegmentDestination(ctx, &xray.UpdateTraceSegmentDestinationInput{
			Destination: xraytypes.TraceSegmentDestinationCloudWatchLogs,
		}); err != nil {
			return fmt.Errorf("updating X-Ray trace segment destination: %w", err)
		}
	}

	return nil
}

func (r *agentRuntimeResource) ensureXRayResourcePolicy(ctx context.Context) error {
	meta := r.Meta()
	logsConn := meta.LogsClient(ctx)
	region := meta.Region(ctx)
	accountID := meta.AccountID(ctx)

	policyDocument := fmt.Sprintf(`{
		"Version": "2012-10-17",
		"Statement": [
			{
				"Sid": "TransactionSearchXRayAccess",
				"Effect": "Allow",
				"Principal": {
					"Service": "xray.amazonaws.com"
				},
				"Action": "logs:PutLogEvents",
				"Resource": [
					"arn:aws:logs:%[1]s:%[2]s:log-group:aws/spans:*",
					"arn:aws:logs:%[1]s:%[2]s:log-group:/aws/application-signals/data:*"
				],
				"Condition": {
					"ArnLike": {
						"aws:SourceArn": "arn:aws:xray:%[1]s:%[2]s:*"
					},
					"StringEquals": {
						"aws:SourceAccount": "%[2]s"
					}
				}
			}
		]
	}`, region, accountID)

	_, err := logsConn.PutResourcePolicy(ctx, &cloudwatchlogs.PutResourcePolicyInput{
		PolicyName:     aws.String(xrayResourcePolicyName),
		PolicyDocument: aws.String(policyDocument),
	})
	if errs.IsA[*logstypes.ResourceAlreadyExistsException](err) {
		return nil
	}
	if err != nil {
		return fmt.Errorf("putting CloudWatch Logs resource policy for X-Ray: %w", err)
	}

	return nil
}

func (r *agentRuntimeResource) waitForXRayResourcePolicy(ctx context.Context, timeout time.Duration) error {
	stateConf := &retry.StateChangeConf{
		Pending: []string{"notfound"},
		Target:  []string{"found"},
		Timeout: timeout,
		Refresh: func(ctx context.Context) (any, string, error) {
			logsConn := r.Meta().LogsClient(ctx)
			var nextToken *string
			for {
				out, err := logsConn.DescribeResourcePolicies(ctx, &cloudwatchlogs.DescribeResourcePoliciesInput{
					Limit:     aws.Int32(50),
					NextToken: nextToken,
				})
				if err != nil {
					return nil, "", err
				}
				for _, policy := range out.ResourcePolicies {
					if aws.ToString(policy.PolicyName) == xrayResourcePolicyName {
						return policy, "found", nil
					}
				}
				if out.NextToken == nil {
					break
				}
				nextToken = out.NextToken
			}
			return nil, "notfound", nil
		},
	}
	_, err := stateConf.WaitForStateContext(ctx)
	return err
}

// applyLogGroupRetention sets the retention policy on a CloudWatch Logs log group.
// If the log group does not yet exist it is created first, then the retention policy is applied.
func applyLogGroupRetention(ctx context.Context, logsConn *cloudwatchlogs.Client, logGroupName string, retentionInDays int32) error {
	_, err := logsConn.PutRetentionPolicy(ctx, &cloudwatchlogs.PutRetentionPolicyInput{
		LogGroupName:    aws.String(logGroupName),
		RetentionInDays: aws.Int32(retentionInDays),
	})
	if err == nil {
		return nil
	}
	if !errs.IsA[*logstypes.ResourceNotFoundException](err) {
		return fmt.Errorf("setting retention policy on log group %q: %w", logGroupName, err)
	}
	// Log group does not exist yet — create it, then apply the retention policy.
	if _, cerr := logsConn.CreateLogGroup(ctx, &cloudwatchlogs.CreateLogGroupInput{
		LogGroupName: aws.String(logGroupName),
	}); cerr != nil && !errs.IsA[*logstypes.ResourceAlreadyExistsException](cerr) {
		return fmt.Errorf("creating log group %q: %w", logGroupName, cerr)
	}
	if _, err := logsConn.PutRetentionPolicy(ctx, &cloudwatchlogs.PutRetentionPolicyInput{
		LogGroupName:    aws.String(logGroupName),
		RetentionInDays: aws.Int32(retentionInDays),
	}); err != nil {
		return fmt.Errorf("setting retention policy on log group %q: %w", logGroupName, err)
	}
	return nil
}

// readLogGroupRetention returns the current retention-in-days for the given log group,
// or 0 if no retention policy is set (never expire).
// Returns an error only for API failures; a missing log group is treated as 0, no error.
func readLogGroupRetention(ctx context.Context, logsConn *cloudwatchlogs.Client, logGroupName string) (int32, error) {
	var nextToken *string
	for {
		out, err := logsConn.DescribeLogGroups(ctx, &cloudwatchlogs.DescribeLogGroupsInput{
			LogGroupNamePrefix: aws.String(logGroupName),
			NextToken:          nextToken,
		})
		if err != nil {
			return 0, fmt.Errorf("describing log group %q: %w", logGroupName, err)
		}
		for _, lg := range out.LogGroups {
			if aws.ToString(lg.LogGroupName) == logGroupName {
				if lg.RetentionInDays != nil {
					return *lg.RetentionInDays, nil
				}
				return 0, nil
			}
		}
		if out.NextToken == nil {
			break
		}
		nextToken = out.NextToken
	}
	// Log group does not exist yet (first log event hasn't arrived).
	return 0, nil
}

const (
	xraySamplingRulePrefix = "bedrock-agentcore-"
	xraySamplingRuleMaxLen = 32
)

func xraySamplingRuleName(runtimeID string) string {
	name := xraySamplingRulePrefix + runtimeID
	if len(name) > xraySamplingRuleMaxLen {
		name = name[:xraySamplingRuleMaxLen]
	}
	return name
}

// applyXRaySamplingRule creates or updates an X-Ray sampling rule for the given runtime,
// setting the fixed sampling rate to samplingPercentage/100.
func applyXRaySamplingRule(ctx context.Context, xrayConn *xray.Client, runtimeID string, samplingPercentage int32) error {
	ruleName := xraySamplingRuleName(runtimeID)
	fixedRate := float64(samplingPercentage) / 100.0

	// Check whether the rule already exists.
	var nextToken *string
	for {
		out, err := xrayConn.GetSamplingRules(ctx, &xray.GetSamplingRulesInput{
			NextToken: nextToken,
		})
		if err != nil {
			return fmt.Errorf("getting X-Ray sampling rules: %w", err)
		}
		for _, record := range out.SamplingRuleRecords {
			if record.SamplingRule != nil && aws.ToString(record.SamplingRule.RuleName) == ruleName {
				// Rule exists — update the fixed rate.
				if _, err := xrayConn.UpdateSamplingRule(ctx, &xray.UpdateSamplingRuleInput{
					SamplingRuleUpdate: &xraytypes.SamplingRuleUpdate{
						RuleName:  aws.String(ruleName),
						FixedRate: aws.Float64(fixedRate),
					},
				}); err != nil {
					return fmt.Errorf("updating X-Ray sampling rule %q: %w", ruleName, err)
				}
				return nil
			}
		}
		if out.NextToken == nil {
			break
		}
		nextToken = out.NextToken
	}

	// Rule does not exist — create it.
	if _, err := xrayConn.CreateSamplingRule(ctx, &xray.CreateSamplingRuleInput{
		SamplingRule: &xraytypes.SamplingRule{
			RuleName:      aws.String(ruleName),
			FixedRate:     fixedRate,
			HTTPMethod:    aws.String("*"),
			Host:          aws.String("*"),
			Priority:      aws.Int32(9000),
			ReservoirSize: 5,
			ResourceARN:   aws.String("*"),
			ServiceName:   aws.String("*"),
			ServiceType:   aws.String("*"),
			URLPath:       aws.String("*"),
			Version:       aws.Int32(1),
		},
	}); err != nil {
		return fmt.Errorf("creating X-Ray sampling rule %q: %w", ruleName, err)
	}
	return nil
}

// readXRaySamplingPercentage returns the current sampling percentage (0–100) for the
// per-runtime sampling rule, and whether the rule was found. Returns an error only for
// API failures; a missing rule is treated as (0, false, nil).
func readXRaySamplingPercentage(ctx context.Context, xrayConn *xray.Client, runtimeID string) (int32, bool, error) {
	ruleName := xraySamplingRuleName(runtimeID)
	var nextToken *string
	for {
		out, err := xrayConn.GetSamplingRules(ctx, &xray.GetSamplingRulesInput{
			NextToken: nextToken,
		})
		if err != nil {
			return 0, false, fmt.Errorf("getting X-Ray sampling rules: %w", err)
		}
		for _, record := range out.SamplingRuleRecords {
			if record.SamplingRule != nil && aws.ToString(record.SamplingRule.RuleName) == ruleName {
				percentage := int32(record.SamplingRule.FixedRate * 100)
				return percentage, true, nil
			}
		}
		if out.NextToken == nil {
			break
		}
		nextToken = out.NextToken
	}
	return 0, false, nil
}

// deleteXRaySamplingRule removes the per-runtime X-Ray sampling rule. A missing rule is
// treated as a no-op so that deletes remain idempotent.
func deleteXRaySamplingRule(ctx context.Context, xrayConn *xray.Client, runtimeID string) error {
	ruleName := xraySamplingRuleName(runtimeID)
	if _, err := xrayConn.DeleteSamplingRule(ctx, &xray.DeleteSamplingRuleInput{
		RuleName: aws.String(ruleName),
	}); err != nil {
		if tfawserr.ErrCodeEquals(err, "InvalidRequestException") {
			return nil // rule does not exist
		}
		return fmt.Errorf("deleting X-Ray sampling rule %q: %w", ruleName, err)
	}
	return nil
}

// disableObservability removes OTEL environment variables injected during observability
// configuration and deletes the per-runtime X-Ray sampling rule.
func disableObservability(ctx context.Context, conn *bedrockagentcorecontrol.Client, xrayConn *xray.Client, runtime *bedrockagentcorecontrol.GetAgentRuntimeOutput, runtimeID string) error {
	cleaned := make(map[string]string, len(runtime.EnvironmentVariables))
	for k, v := range runtime.EnvironmentVariables {
		if !isOtelEnvVar(k) {
			cleaned[k] = v
		}
	}

	updateInput := &bedrockagentcorecontrol.UpdateAgentRuntimeInput{
		AgentRuntimeId:          aws.String(runtimeID),
		AgentRuntimeArtifact:    runtime.AgentRuntimeArtifact,
		AuthorizerConfiguration: runtime.AuthorizerConfiguration,
		Description:             runtime.Description,
		EnvironmentVariables:    cleaned,
		NetworkConfiguration:    runtime.NetworkConfiguration,
		ProtocolConfiguration:   runtime.ProtocolConfiguration,
		RoleArn:                 runtime.RoleArn,
	}
	if _, err := conn.UpdateAgentRuntime(ctx, updateInput); err != nil {
		return fmt.Errorf("removing observability environment variables from agent runtime: %w", err)
	}

	return deleteXRaySamplingRule(ctx, xrayConn, runtimeID)
}

// isOtelEnvVar reports whether key is an OTEL env var injected by configureObservability.
func isOtelEnvVar(key string) bool {
	switch key {
	case "AGENT_OBSERVABILITY_ENABLED",
		"OTEL_EXPORTER_OTLP_LOGS_ENDPOINT",
		"OTEL_EXPORTER_OTLP_LOGS_HEADERS",
		"OTEL_EXPORTER_OTLP_PROTOCOL",
		"OTEL_EXPORTER_OTLP_TRACES_ENDPOINT",
		"OTEL_RESOURCE_ATTRIBUTES",
		"OTEL_TRACES_EXPORTER",
		"OTEL_PYTHON_DISTRO",
		"OTEL_PYTHON_CONFIGURATOR",
		"OTEL_PYTHON_LOGGING_AUTO_INSTRUMENTATION_ENABLED":
		return true
	}
	return false
}
