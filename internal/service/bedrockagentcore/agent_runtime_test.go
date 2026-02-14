// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package bedrockagentcore_test

import (
	"context"
	"fmt"
	"strings"
	"testing"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/service/bedrockagentcorecontrol"
	awstypes "github.com/aws/aws-sdk-go-v2/service/bedrockagentcorecontrol/types"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	tfknownvalue "github.com/hashicorp/terraform-provider-aws/internal/acctest/knownvalue"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	tfbedrockagentcore "github.com/hashicorp/terraform-provider-aws/internal/service/bedrockagentcore"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccBedrockAgentCoreAgentRuntime_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var agentRuntime bedrockagentcorecontrol.GetAgentRuntimeOutput
	rName := strings.ReplaceAll(acctest.RandomWithPrefix(t, acctest.ResourcePrefix), "-", "_")
	resourceName := "aws_bedrockagentcore_agent_runtime.test"
	rImageUri := acctest.SkipIfEnvVarNotSet(t, "AWS_BEDROCK_AGENTCORE_RUNTIME_IMAGE_V1_URI")

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.BedrockEndpointID)
			testAccPreCheckAgentRuntimes(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.BedrockAgentCoreServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckAgentRuntimeDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccAgentRuntimeConfig_basic(rName, rImageUri),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAgentRuntimeExists(ctx, t, resourceName, &agentRuntime),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("agent_runtime_arn"), tfknownvalue.RegionalARNRegexp("bedrock-agentcore", regexache.MustCompile(`runtime/.+`))),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("agent_runtime_id"), knownvalue.NotNull()),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("agent_runtime_name"), knownvalue.StringExact(rName)),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("agent_runtime_version"), knownvalue.NotNull()),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrTags), knownvalue.Null()),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("workload_identity_details"), knownvalue.ListSizeExact(1)),
				},
			},
			{
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateIdFunc:                    acctest.AttrImportStateIdFunc(resourceName, "agent_runtime_id"),
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: "agent_runtime_id",
			},
		},
	})
}

func TestAccBedrockAgentCoreAgentRuntime_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	var agentRuntime bedrockagentcorecontrol.GetAgentRuntimeOutput
	rName := strings.ReplaceAll(acctest.RandomWithPrefix(t, acctest.ResourcePrefix), "-", "_")
	resourceName := "aws_bedrockagentcore_agent_runtime.test"
	rImageUri := acctest.SkipIfEnvVarNotSet(t, "AWS_BEDROCK_AGENTCORE_RUNTIME_IMAGE_V1_URI")

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.BedrockEndpointID)
			testAccPreCheckAgentRuntimes(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.BedrockAgentCoreServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckAgentRuntimeDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccAgentRuntimeConfig_basic(rName, rImageUri),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAgentRuntimeExists(ctx, t, resourceName, &agentRuntime),
					acctest.CheckFrameworkResourceDisappears(ctx, t, tfbedrockagentcore.ResourceAgentRuntime, resourceName),
				),
				ExpectNonEmptyPlan: true,
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
					},
					PostApplyPostRefresh: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
					},
				},
			},
		},
	})
}

func TestAccBedrockAgentCoreAgentRuntime_tags(t *testing.T) {
	ctx := acctest.Context(t)
	var agentRuntime bedrockagentcorecontrol.GetAgentRuntimeOutput
	rName := strings.ReplaceAll(acctest.RandomWithPrefix(t, acctest.ResourcePrefix), "-", "_")
	resourceName := "aws_bedrockagentcore_agent_runtime.test"
	rImageUri := acctest.SkipIfEnvVarNotSet(t, "AWS_BEDROCK_AGENTCORE_RUNTIME_IMAGE_V1_URI")

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.BedrockEndpointID)
			testAccPreCheckAgentRuntimes(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.BedrockAgentCoreServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckAgentRuntimeDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccAgentRuntimeConfig_tags1(rName, rImageUri, acctest.CtKey1, acctest.CtValue1),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAgentRuntimeExists(ctx, t, resourceName, &agentRuntime),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrTags), knownvalue.MapExact(map[string]knownvalue.Check{
						acctest.CtKey1: knownvalue.StringExact(acctest.CtValue1),
					})),
				},
			},
			{
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateIdFunc:                    acctest.AttrImportStateIdFunc(resourceName, "agent_runtime_id"),
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: "agent_runtime_id",
			},
			{
				Config: testAccAgentRuntimeConfig_tags2(rName, rImageUri, acctest.CtKey1, acctest.CtValue1Updated, acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAgentRuntimeExists(ctx, t, resourceName, &agentRuntime),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionUpdate),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrTags), knownvalue.MapExact(map[string]knownvalue.Check{
						acctest.CtKey1: knownvalue.StringExact(acctest.CtValue1Updated),
						acctest.CtKey2: knownvalue.StringExact(acctest.CtValue2),
					})),
				},
			},
			{
				Config: testAccAgentRuntimeConfig_tags1(rName, rImageUri, acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAgentRuntimeExists(ctx, t, resourceName, &agentRuntime),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionUpdate),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrTags), knownvalue.MapExact(map[string]knownvalue.Check{
						acctest.CtKey2: knownvalue.StringExact(acctest.CtValue2),
					})),
				},
			},
		},
	})
}

func TestAccBedrockAgentCoreAgentRuntime_description(t *testing.T) {
	ctx := acctest.Context(t)
	var agentRuntime bedrockagentcorecontrol.GetAgentRuntimeOutput
	rName := strings.ReplaceAll(acctest.RandomWithPrefix(t, acctest.ResourcePrefix), "-", "_")
	resourceName := "aws_bedrockagentcore_agent_runtime.test"
	rImageUri := acctest.SkipIfEnvVarNotSet(t, "AWS_BEDROCK_AGENTCORE_RUNTIME_IMAGE_V1_URI")

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.BedrockEndpointID)
			testAccPreCheckAgentRuntimes(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.BedrockAgentCoreServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckAgentRuntimeDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccAgentRuntimeConfig_description(rName, rImageUri, "Initial description"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAgentRuntimeExists(ctx, t, resourceName, &agentRuntime),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, "Initial description"),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrDescription), knownvalue.StringExact("Initial description")),
				},
			},
			{
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateIdFunc:                    acctest.AttrImportStateIdFunc(resourceName, "agent_runtime_id"),
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: "agent_runtime_id",
			},
			{
				Config: testAccAgentRuntimeConfig_description(rName, rImageUri, "Updated description"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAgentRuntimeExists(ctx, t, resourceName, &agentRuntime),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionUpdate),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrDescription), knownvalue.StringExact("Updated description")),
				},
			},
		},
	})
}

func TestAccBedrockAgentCoreAgentRuntime_environmentVariables(t *testing.T) {
	ctx := acctest.Context(t)
	var agentRuntime bedrockagentcorecontrol.GetAgentRuntimeOutput
	rName := strings.ReplaceAll(acctest.RandomWithPrefix(t, acctest.ResourcePrefix), "-", "_")
	resourceName := "aws_bedrockagentcore_agent_runtime.test"
	rImageUri := acctest.SkipIfEnvVarNotSet(t, "AWS_BEDROCK_AGENTCORE_RUNTIME_IMAGE_V1_URI")

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.BedrockEndpointID)
			testAccPreCheckAgentRuntimes(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.BedrockAgentCoreServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckAgentRuntimeDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccAgentRuntimeConfig_environmentVariables(rName, rImageUri, "ENV_KEY_1", "env_value_1"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAgentRuntimeExists(ctx, t, resourceName, &agentRuntime),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("environment_variables"), knownvalue.MapExact(map[string]knownvalue.Check{
						"ENV_KEY_1": knownvalue.StringExact("env_value_1"),
					})),
				},
			},
			{
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateIdFunc:                    acctest.AttrImportStateIdFunc(resourceName, "agent_runtime_id"),
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: "agent_runtime_id",
			},
			{
				Config: testAccAgentRuntimeConfig_environmentVariables(rName, rImageUri, "ENV_KEY_2", "env_value_2_updated"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAgentRuntimeExists(ctx, t, resourceName, &agentRuntime),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionUpdate),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("environment_variables"), knownvalue.MapExact(map[string]knownvalue.Check{
						"ENV_KEY_2": knownvalue.StringExact("env_value_2_updated"),
					})),
				},
			},
		},
	})
}

func TestAccBedrockAgentCoreAgentRuntime_authorizerConfiguration(t *testing.T) {
	ctx := acctest.Context(t)
	var agentRuntime bedrockagentcorecontrol.GetAgentRuntimeOutput
	rName := strings.ReplaceAll(acctest.RandomWithPrefix(t, acctest.ResourcePrefix), "-", "_")
	resourceName := "aws_bedrockagentcore_agent_runtime.test"
	rImageUri := acctest.SkipIfEnvVarNotSet(t, "AWS_BEDROCK_AGENTCORE_RUNTIME_IMAGE_V1_URI")

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.BedrockEndpointID)
			testAccPreCheckAgentRuntimes(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.BedrockAgentCoreServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckAgentRuntimeDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccAgentRuntimeConfig_authorizerConfiguration(rName, rImageUri, "https://accounts.google.com/.well-known/openid-configuration", "weather", "sports", "client-999", "client-888"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAgentRuntimeExists(ctx, t, resourceName, &agentRuntime),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("authorizer_configuration"), knownvalue.ListExact([]knownvalue.Check{
						knownvalue.ObjectExact(map[string]knownvalue.Check{
							"custom_jwt_authorizer": knownvalue.ListExact([]knownvalue.Check{
								knownvalue.ObjectExact(map[string]knownvalue.Check{
									"allowed_audience": knownvalue.SetExact([]knownvalue.Check{
										knownvalue.StringExact("sports"),
										knownvalue.StringExact("weather"),
									}),
									"allowed_clients": knownvalue.SetExact([]knownvalue.Check{
										knownvalue.StringExact("client-888"),
										knownvalue.StringExact("client-999"),
									}),
									"discovery_url": knownvalue.StringExact("https://accounts.google.com/.well-known/openid-configuration"),
								}),
							}),
						}),
					})),
				},
			},
			{
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateIdFunc:                    acctest.AttrImportStateIdFunc(resourceName, "agent_runtime_id"),
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: "agent_runtime_id",
			},
			{
				Config: testAccAgentRuntimeConfig_authorizerConfiguration(rName, rImageUri, "https://login.microsoftonline.com/common/v2.0/.well-known/openid-configuration", "finance", "technology", "client-111", "client-222"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAgentRuntimeExists(ctx, t, resourceName, &agentRuntime),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionUpdate),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("authorizer_configuration"), knownvalue.ListExact([]knownvalue.Check{
						knownvalue.ObjectExact(map[string]knownvalue.Check{
							"custom_jwt_authorizer": knownvalue.ListExact([]knownvalue.Check{
								knownvalue.ObjectExact(map[string]knownvalue.Check{
									"allowed_audience": knownvalue.SetExact([]knownvalue.Check{
										knownvalue.StringExact("finance"),
										knownvalue.StringExact("technology"),
									}),
									"allowed_clients": knownvalue.SetExact([]knownvalue.Check{
										knownvalue.StringExact("client-111"),
										knownvalue.StringExact("client-222"),
									}),
									"discovery_url": knownvalue.StringExact("https://login.microsoftonline.com/common/v2.0/.well-known/openid-configuration"),
								}),
							}),
						}),
					})),
				},
			},
		},
	})
}

func TestAccBedrockAgentCoreAgentRuntime_protocolConfiguration(t *testing.T) {
	ctx := acctest.Context(t)
	var agentRuntime bedrockagentcorecontrol.GetAgentRuntimeOutput
	rName := strings.ReplaceAll(acctest.RandomWithPrefix(t, acctest.ResourcePrefix), "-", "_")
	resourceName := "aws_bedrockagentcore_agent_runtime.test"
	rImageUri := acctest.SkipIfEnvVarNotSet(t, "AWS_BEDROCK_AGENTCORE_RUNTIME_IMAGE_V1_URI")

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.BedrockEndpointID)
			testAccPreCheckAgentRuntimes(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.BedrockAgentCoreServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckAgentRuntimeDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccAgentRuntimeConfig_protocolConfiguration(rName, rImageUri, "HTTP"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAgentRuntimeExists(ctx, t, resourceName, &agentRuntime),
					resource.TestCheckResourceAttr(resourceName, "protocol_configuration.0.server_protocol", "HTTP"),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("protocol_configuration"), knownvalue.ListExact([]knownvalue.Check{
						knownvalue.ObjectExact(map[string]knownvalue.Check{
							"server_protocol": tfknownvalue.StringExact(awstypes.ServerProtocolHttp),
						}),
					})),
				},
			},
			{
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateIdFunc:                    acctest.AttrImportStateIdFunc(resourceName, "agent_runtime_id"),
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: "agent_runtime_id",
			},
			{
				Config: testAccAgentRuntimeConfig_protocolConfiguration(rName, rImageUri, "MCP"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAgentRuntimeExists(ctx, t, resourceName, &agentRuntime),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionUpdate),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("protocol_configuration"), knownvalue.ListExact([]knownvalue.Check{
						knownvalue.ObjectExact(map[string]knownvalue.Check{
							"server_protocol": tfknownvalue.StringExact(awstypes.ServerProtocolMcp),
						}),
					})),
				},
			},
		},
	})
}

func TestAccBedrockAgentCoreAgentRuntime_artifactContainer(t *testing.T) {
	ctx := acctest.Context(t)
	var agentRuntime bedrockagentcorecontrol.GetAgentRuntimeOutput
	rName := strings.ReplaceAll(acctest.RandomWithPrefix(t, acctest.ResourcePrefix), "-", "_")
	resourceName := "aws_bedrockagentcore_agent_runtime.test"
	rImageUriV1 := acctest.SkipIfEnvVarNotSet(t, "AWS_BEDROCK_AGENTCORE_RUNTIME_IMAGE_V1_URI")
	rImageUriV2 := acctest.SkipIfEnvVarNotSet(t, "AWS_BEDROCK_AGENTCORE_RUNTIME_IMAGE_V2_URI")

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.BedrockEndpointID)
			testAccPreCheckAgentRuntimes(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.BedrockAgentCoreServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckAgentRuntimeDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccAgentRuntimeConfig_basic(rName, rImageUriV1),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAgentRuntimeExists(ctx, t, resourceName, &agentRuntime),
					resource.TestCheckResourceAttr(resourceName, "agent_runtime_artifact.0.container_configuration.0.container_uri", rImageUriV1),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("agent_runtime_artifact"), knownvalue.ListExact([]knownvalue.Check{
						knownvalue.ObjectExact(map[string]knownvalue.Check{
							"code_configuration": knownvalue.ListExact([]knownvalue.Check{}),
							"container_configuration": knownvalue.ListExact([]knownvalue.Check{
								knownvalue.ObjectExact(map[string]knownvalue.Check{
									"container_uri": knownvalue.StringExact(rImageUriV1),
								}),
							}),
						}),
					})),
				},
			},
			{
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateIdFunc:                    acctest.AttrImportStateIdFunc(resourceName, "agent_runtime_id"),
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: "agent_runtime_id",
			},
			{
				Config: testAccAgentRuntimeConfig_basic(rName, rImageUriV2),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAgentRuntimeExists(ctx, t, resourceName, &agentRuntime),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionUpdate),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("agent_runtime_artifact"), knownvalue.ListExact([]knownvalue.Check{
						knownvalue.ObjectExact(map[string]knownvalue.Check{
							"code_configuration": knownvalue.ListExact([]knownvalue.Check{}),
							"container_configuration": knownvalue.ListExact([]knownvalue.Check{
								knownvalue.ObjectExact(map[string]knownvalue.Check{
									"container_uri": knownvalue.StringExact(rImageUriV2),
								}),
							}),
						}),
					})),
				},
			},
		},
	})
}

// This test requires a pre-uploaded S3 object containing a ZIP file.
// The test will be skipped if the relevant environment variables are not set.
// A sample ZIP file can be obtained from the AWS Management Console using “Start with a template”.
func TestAccBedrockAgentCoreAgentRuntime_artifactCode(t *testing.T) {
	ctx := acctest.Context(t)
	var agentRuntime bedrockagentcorecontrol.GetAgentRuntimeOutput
	rName := strings.ReplaceAll(acctest.RandomWithPrefix(t, acctest.ResourcePrefix), "-", "_")
	resourceName := "aws_bedrockagentcore_agent_runtime.test"
	rCodeS3BucketV1 := acctest.SkipIfEnvVarNotSet(t, "AWS_BEDROCK_AGENTCORE_RUNTIME_CODE_V1_S3_BUCKET")
	rCodeS3KeyV1 := acctest.SkipIfEnvVarNotSet(t, "AWS_BEDROCK_AGENTCORE_RUNTIME_CODE_V1_S3_KEY")
	rCodeS3BucketV2 := acctest.SkipIfEnvVarNotSet(t, "AWS_BEDROCK_AGENTCORE_RUNTIME_CODE_V2_S3_BUCKET")
	rCodeS3KeyV2 := acctest.SkipIfEnvVarNotSet(t, "AWS_BEDROCK_AGENTCORE_RUNTIME_CODE_V2_S3_KEY")

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.BedrockEndpointID)
			testAccPreCheckAgentRuntimes(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.BedrockAgentCoreServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckAgentRuntimeDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccAgentRuntimeConfig_codeConfiguration(rName, rCodeS3BucketV1, rCodeS3KeyV1),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAgentRuntimeExists(ctx, t, resourceName, &agentRuntime),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("agent_runtime_artifact"), knownvalue.ListExact([]knownvalue.Check{
						knownvalue.ObjectExact(map[string]knownvalue.Check{
							"code_configuration": knownvalue.ListExact([]knownvalue.Check{
								knownvalue.ObjectExact(map[string]knownvalue.Check{
									"entry_point": knownvalue.ListExact([]knownvalue.Check{
										knownvalue.StringExact("main.py"),
									}),
									"runtime": knownvalue.StringExact(string(awstypes.AgentManagedRuntimeTypePython313)),
									"code": knownvalue.ListExact([]knownvalue.Check{
										knownvalue.ObjectExact(map[string]knownvalue.Check{
											"s3": knownvalue.ListExact([]knownvalue.Check{
												knownvalue.ObjectExact(map[string]knownvalue.Check{
													names.AttrBucket: knownvalue.StringExact(rCodeS3BucketV1),
													names.AttrPrefix: knownvalue.StringExact(rCodeS3KeyV1),
													"version_id":     knownvalue.Null(),
												}),
											}),
										}),
									}),
								}),
							}),
							"container_configuration": knownvalue.ListExact([]knownvalue.Check{}),
						}),
					})),
				},
			},
			{
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateIdFunc:                    acctest.AttrImportStateIdFunc(resourceName, "agent_runtime_id"),
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: "agent_runtime_id",
			},
			{
				Config: testAccAgentRuntimeConfig_codeConfiguration(rName, rCodeS3BucketV2, rCodeS3KeyV2),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAgentRuntimeExists(ctx, t, resourceName, &agentRuntime),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionUpdate),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("agent_runtime_artifact"), knownvalue.ListExact([]knownvalue.Check{
						knownvalue.ObjectExact(map[string]knownvalue.Check{
							"code_configuration": knownvalue.ListExact([]knownvalue.Check{
								knownvalue.ObjectExact(map[string]knownvalue.Check{
									"entry_point": knownvalue.ListExact([]knownvalue.Check{
										knownvalue.StringExact("main.py"),
									}),
									"runtime": knownvalue.StringExact(string(awstypes.AgentManagedRuntimeTypePython313)),
									"code": knownvalue.ListExact([]knownvalue.Check{
										knownvalue.ObjectExact(map[string]knownvalue.Check{
											"s3": knownvalue.ListExact([]knownvalue.Check{
												knownvalue.ObjectExact(map[string]knownvalue.Check{
													names.AttrBucket: knownvalue.StringExact(rCodeS3BucketV2),
													names.AttrPrefix: knownvalue.StringExact(rCodeS3KeyV2),
													"version_id":     knownvalue.Null(),
												}),
											}),
										}),
									}),
								}),
							}),
							"container_configuration": knownvalue.ListExact([]knownvalue.Check{}),
						}),
					})),
				},
			},
		},
	})
}

// Ensure that changing the artifact type forces a new resource.
func TestAccBedrockAgentCoreAgentRuntime_artifactTypeChanged(t *testing.T) {
	ctx := acctest.Context(t)
	var agentRuntime bedrockagentcorecontrol.GetAgentRuntimeOutput
	rName := strings.ReplaceAll(acctest.RandomWithPrefix(t, acctest.ResourcePrefix), "-", "_")
	resourceName := "aws_bedrockagentcore_agent_runtime.test"
	rImageUriV1 := acctest.SkipIfEnvVarNotSet(t, "AWS_BEDROCK_AGENTCORE_RUNTIME_IMAGE_V1_URI")
	rCodeS3BucketV1 := acctest.SkipIfEnvVarNotSet(t, "AWS_BEDROCK_AGENTCORE_RUNTIME_CODE_V1_S3_BUCKET")
	rCodeS3KeyV1 := acctest.SkipIfEnvVarNotSet(t, "AWS_BEDROCK_AGENTCORE_RUNTIME_CODE_V1_S3_KEY")

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.BedrockEndpointID)
			testAccPreCheckAgentRuntimes(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.BedrockAgentCoreServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckAgentRuntimeDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccAgentRuntimeConfig_basic(rName, rImageUriV1),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAgentRuntimeExists(ctx, t, resourceName, &agentRuntime),
					resource.TestCheckResourceAttr(resourceName, "agent_runtime_artifact.0.container_configuration.0.container_uri", rImageUriV1),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("agent_runtime_artifact"), knownvalue.ListExact([]knownvalue.Check{
						knownvalue.ObjectExact(map[string]knownvalue.Check{
							"code_configuration": knownvalue.ListExact([]knownvalue.Check{}),
							"container_configuration": knownvalue.ListExact([]knownvalue.Check{
								knownvalue.ObjectExact(map[string]knownvalue.Check{
									"container_uri": knownvalue.StringExact(rImageUriV1),
								}),
							}),
						}),
					})),
				},
			},
			{
				// Switch to code artifact, expect destroy/create
				Config: testAccAgentRuntimeConfig_codeConfiguration(rName, rCodeS3BucketV1, rCodeS3KeyV1),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAgentRuntimeExists(ctx, t, resourceName, &agentRuntime),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionDestroyBeforeCreate),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("agent_runtime_artifact"), knownvalue.ListExact([]knownvalue.Check{
						knownvalue.ObjectExact(map[string]knownvalue.Check{
							"code_configuration": knownvalue.ListExact([]knownvalue.Check{
								knownvalue.ObjectExact(map[string]knownvalue.Check{
									"entry_point": knownvalue.ListExact([]knownvalue.Check{
										knownvalue.StringExact("main.py"),
									}),
									"runtime": knownvalue.StringExact(string(awstypes.AgentManagedRuntimeTypePython313)),
									"code": knownvalue.ListExact([]knownvalue.Check{
										knownvalue.ObjectExact(map[string]knownvalue.Check{
											"s3": knownvalue.ListExact([]knownvalue.Check{
												knownvalue.ObjectExact(map[string]knownvalue.Check{
													names.AttrBucket: knownvalue.StringExact(rCodeS3BucketV1),
													names.AttrPrefix: knownvalue.StringExact(rCodeS3KeyV1),
													"version_id":     knownvalue.Null(),
												}),
											}),
										}),
									}),
								}),
							}),
							"container_configuration": knownvalue.ListExact([]knownvalue.Check{}),
						}),
					})),
				},
			},
			{
				// Switch back to container artifact, expect destroy/create
				Config: testAccAgentRuntimeConfig_basic(rName, rImageUriV1),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAgentRuntimeExists(ctx, t, resourceName, &agentRuntime),
					resource.TestCheckResourceAttr(resourceName, "agent_runtime_artifact.0.container_configuration.0.container_uri", rImageUriV1),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionDestroyBeforeCreate),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("agent_runtime_artifact"), knownvalue.ListExact([]knownvalue.Check{
						knownvalue.ObjectExact(map[string]knownvalue.Check{
							"code_configuration": knownvalue.ListExact([]knownvalue.Check{}),
							"container_configuration": knownvalue.ListExact([]knownvalue.Check{
								knownvalue.ObjectExact(map[string]knownvalue.Check{
									"container_uri": knownvalue.StringExact(rImageUriV1),
								}),
							}),
						}),
					})),
				},
			},
		},
	})
}

func testAccCheckAgentRuntimeDestroy(ctx context.Context, t *testing.T) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.ProviderMeta(ctx, t).BedrockAgentCoreClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_bedrockagentcore_agent_runtime" {
				continue
			}

			_, err := tfbedrockagentcore.FindAgentRuntimeByID(ctx, conn, rs.Primary.Attributes["agent_runtime_id"])
			if retry.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("Bedrock Agent Core Agent Runtime %s still exists", rs.Primary.Attributes["agent_runtime_id"])
		}

		return nil
	}
}

func testAccCheckAgentRuntimeExists(ctx context.Context, t *testing.T, n string, v *bedrockagentcorecontrol.GetAgentRuntimeOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.ProviderMeta(ctx, t).BedrockAgentCoreClient(ctx)

		resp, err := tfbedrockagentcore.FindAgentRuntimeByID(ctx, conn, rs.Primary.Attributes["agent_runtime_id"])
		if err != nil {
			return err
		}

		*v = *resp

		return nil
	}
}

func testAccPreCheckAgentRuntimes(ctx context.Context, t *testing.T) {
	conn := acctest.ProviderMeta(ctx, t).BedrockAgentCoreClient(ctx)

	var input bedrockagentcorecontrol.ListAgentRuntimesInput

	_, err := conn.ListAgentRuntimes(ctx, &input)

	if acctest.PreCheckSkipError(err) {
		t.Skipf("skipping acceptance testing: %s", err)
	}
	if err != nil {
		t.Fatalf("unexpected PreCheck error: %s", err)
	}
}

func testAccAgentRuntimeConfig_baseIAMRole(rName string) string {
	return fmt.Sprintf(`
data "aws_iam_policy_document" "test_assume" {
  statement {
    effect  = "Allow"
    actions = ["sts:AssumeRole"]
    principals {
      type        = "Service"
      identifiers = ["bedrock-agentcore.amazonaws.com"]
    }
  }
}

data "aws_iam_policy_document" "test" {
  statement {
    actions = [
      "ecr:GetAuthorizationToken",
      "ecr:BatchGetImage",
      "ecr:GetDownloadUrlForLayer"
    ]
    effect    = "Allow"
    resources = ["*"]
  }
}

resource "aws_iam_role" "test" {
  name               = %[1]q
  assume_role_policy = data.aws_iam_policy_document.test_assume.json
}

resource "aws_iam_role_policy" "test" {
  role   = aws_iam_role.test.id
  policy = data.aws_iam_policy_document.test.json
}

resource "aws_iam_role" "test2" {
  name               = "%[1]s-2"
  assume_role_policy = data.aws_iam_policy_document.test_assume.json
}

resource "aws_iam_role_policy" "test2" {
  role   = aws_iam_role.test2.id
  policy = data.aws_iam_policy_document.test.json
}
`, rName)
}

func testAccAgentRuntimeConfig_basic(rName, rImageUri string) string {
	return acctest.ConfigCompose(testAccAgentRuntimeConfig_baseIAMRole(rName), fmt.Sprintf(`
resource "aws_bedrockagentcore_agent_runtime" "test" {
  agent_runtime_name = %[1]q
  role_arn           = aws_iam_role.test.arn

  agent_runtime_artifact {
    container_configuration {
      container_uri = %[2]q
    }
  }

  network_configuration {
    network_mode = "PUBLIC"
  }
}
`, rName, rImageUri))
}

func testAccAgentRuntimeConfig_tags1(rName, rImageUri, tag1Key, tag1Value string) string {
	return acctest.ConfigCompose(testAccAgentRuntimeConfig_baseIAMRole(rName), fmt.Sprintf(`
resource "aws_bedrockagentcore_agent_runtime" "test" {
  agent_runtime_name = %[1]q
  role_arn           = aws_iam_role.test.arn

  agent_runtime_artifact {
    container_configuration {
      container_uri = %[2]q
    }
  }

  network_configuration {
    network_mode = "PUBLIC"
  }

  tags = {
    %[3]q = %[4]q
  }
}
`, rName, rImageUri, tag1Key, tag1Value))
}

func testAccAgentRuntimeConfig_tags2(rName, rImageUri, tag1Key, tag1Value, tag2Key, tag2Value string) string {
	return acctest.ConfigCompose(testAccAgentRuntimeConfig_baseIAMRole(rName), fmt.Sprintf(`
resource "aws_bedrockagentcore_agent_runtime" "test" {
  agent_runtime_name = %[1]q
  role_arn           = aws_iam_role.test.arn

  agent_runtime_artifact {
    container_configuration {
      container_uri = %[2]q
    }
  }

  network_configuration {
    network_mode = "PUBLIC"
  }

  tags = {
    %[3]q = %[4]q
    %[5]q = %[6]q
  }
}
`, rName, rImageUri, tag1Key, tag1Value, tag2Key, tag2Value))
}

func testAccAgentRuntimeConfig_description(rName, rImageUri, description string) string {
	return acctest.ConfigCompose(testAccAgentRuntimeConfig_baseIAMRole(rName), fmt.Sprintf(`
resource "aws_bedrockagentcore_agent_runtime" "test" {
  agent_runtime_name = %[1]q
  role_arn           = aws_iam_role.test.arn
  description        = %[2]q

  agent_runtime_artifact {
    container_configuration {
      container_uri = %[3]q
    }
  }

  network_configuration {
    network_mode = "PUBLIC"
  }
}
`, rName, description, rImageUri))
}

func testAccAgentRuntimeConfig_environmentVariables(rName, rImageUri, envKey, envValue string) string {
	return acctest.ConfigCompose(testAccAgentRuntimeConfig_baseIAMRole(rName), fmt.Sprintf(`
resource "aws_bedrockagentcore_agent_runtime" "test" {
  agent_runtime_name = %[1]q
  role_arn           = aws_iam_role.test.arn

  environment_variables = {
    %[2]s = %[3]q
  }

  agent_runtime_artifact {
    container_configuration {
      container_uri = %[4]q
    }
  }

  network_configuration {
    network_mode = "PUBLIC"
  }
}
`, rName, envKey, envValue, rImageUri))
}

func testAccAgentRuntimeConfig_authorizerConfiguration(rName, rImageUri, discoveryUrl, audience1, audience2, client1, client2 string) string {
	return acctest.ConfigCompose(testAccAgentRuntimeConfig_baseIAMRole(rName), fmt.Sprintf(`
resource "aws_bedrockagentcore_agent_runtime" "test" {
  agent_runtime_name = %[1]q
  role_arn           = aws_iam_role.test.arn

  agent_runtime_artifact {
    container_configuration {
      container_uri = %[2]q
    }
  }

  authorizer_configuration {
    custom_jwt_authorizer {
      discovery_url    = %[3]q
      allowed_audience = [%[4]q, %[5]q]
      allowed_clients  = [%[6]q, %[7]q]
    }
  }

  network_configuration {
    network_mode = "PUBLIC"
  }
}
`, rName, rImageUri, discoveryUrl, audience1, audience2, client1, client2))
}

func testAccAgentRuntimeConfig_protocolConfiguration(rName, rImageUri, serverProtocol string) string {
	return acctest.ConfigCompose(testAccAgentRuntimeConfig_baseIAMRole(rName), fmt.Sprintf(`
resource "aws_bedrockagentcore_agent_runtime" "test" {
  agent_runtime_name = %[1]q
  role_arn           = aws_iam_role.test.arn

  agent_runtime_artifact {
    container_configuration {
      container_uri = %[2]q
    }
  }

  network_configuration {
    network_mode = "PUBLIC"
  }

  protocol_configuration {
    server_protocol = %[3]q
  }
}
`, rName, rImageUri, serverProtocol))
}

func testAccAgentRuntimeConfig_codeConfiguration(rName, s3Bucket, s3Key string) string {
	return acctest.ConfigCompose(testAccAgentRuntimeConfig_baseIAMRole(rName), fmt.Sprintf(`
resource "aws_bedrockagentcore_agent_runtime" "test" {
  agent_runtime_name = %[1]q
  role_arn           = aws_iam_role.test.arn

  agent_runtime_artifact {
    code_configuration {
      entry_point = ["main.py"]
      runtime     = "PYTHON_3_13"
      code {
        s3 {
          bucket = %[2]q
          prefix = %[3]q
        }
      }
    }
  }

  network_configuration {
    network_mode = "PUBLIC"
  }
}
`, rName, s3Bucket, s3Key))
}
