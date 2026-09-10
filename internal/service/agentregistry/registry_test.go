// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package agentregistry_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/YakDriver/regexache"
	awstypes "github.com/aws/aws-sdk-go-v2/service/agentregistrycontrol/types"
	"github.com/hashicorp/terraform-plugin-testing/config"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	tfknownvalue "github.com/hashicorp/terraform-provider-aws/internal/acctest/knownvalue"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	tfagentregistry "github.com/hashicorp/terraform-provider-aws/internal/service/agentregistry"
	"github.com/hashicorp/terraform-provider-aws/names"
)

var (
	checkRegistryARN                knownvalue.Check = tfknownvalue.RegionalARNRegexp("agent-registry", regexache.MustCompile(`registry/[a-zA-Z0-9]{12,16}`))
	checkRegistryARNAlternateRegion knownvalue.Check = tfknownvalue.RegionalARNAlternateRegionRegexp("agent-registry", regexache.MustCompile(`registry/[a-zA-Z0-9]{12,16}`))
)

func TestAccAgentRegistryRegistry_basic(t *testing.T) {
	ctx := acctest.Context(t)
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_agentregistry_registry.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.AgentRegistryServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckRegistryDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				ConfigDirectory: config.StaticDirectory("testdata/Registry/basic/"),
				ConfigVariables: config.Variables{
					acctest.CtRName: config.StringVariable(rName),
				},
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckRegistryExists(ctx, t, resourceName),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("approval_configuration"), knownvalue.ListSizeExact(0)),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("auto_detection_configuration"), knownvalue.ListSizeExact(0)),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrDescription), knownvalue.Null()),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("discovery_configuration"), knownvalue.ListExact([]knownvalue.Check{knownvalue.ObjectExact(map[string]knownvalue.Check{
						"authorizer_configuration": knownvalue.ListSizeExact(0),
						"authorizer_type":          tfknownvalue.StringExact(awstypes.RegistryAuthorizerTypeAwsIam),
					})})),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrName), knownvalue.StringExact(rName)),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("registry_arn"), checkRegistryARN),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("registry_id"), knownvalue.NotNull()),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrTags), knownvalue.Null()),
				},
			},
			{
				ConfigDirectory: config.StaticDirectory("testdata/Registry/basic/"),
				ConfigVariables: config.Variables{
					acctest.CtRName: config.StringVariable(rName),
				},
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateVerify:                    true,
				ImportStateIdFunc:                    acctest.AttrImportStateIdFunc(resourceName, "registry_id"),
				ImportStateVerifyIdentifierAttribute: "registry_id",
			},
		},
	})
}

func TestAccAgentRegistryRegistry_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_agentregistry_registry.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.AgentRegistryServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckRegistryDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				ConfigDirectory: config.StaticDirectory("testdata/Registry/basic/"),
				ConfigVariables: config.Variables{
					acctest.CtRName: config.StringVariable(rName),
				},
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckRegistryExists(ctx, t, resourceName),
					acctest.CheckFrameworkResourceDisappears(ctx, t, tfagentregistry.ResourceRegistry, resourceName),
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

func TestAccAgentRegistryRegistry_name(t *testing.T) {
	ctx := acctest.Context(t)
	rName1 := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	rName2 := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_agentregistry_registry.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.AgentRegistryServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckRegistryDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				ConfigDirectory: config.StaticDirectory("testdata/Registry/basic/"),
				ConfigVariables: config.Variables{
					acctest.CtRName: config.StringVariable(rName1),
				},
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckRegistryExists(ctx, t, resourceName),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("approval_configuration"), knownvalue.ListSizeExact(0)),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("auto_detection_configuration"), knownvalue.ListSizeExact(0)),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrDescription), knownvalue.Null()),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("discovery_configuration"), knownvalue.ListExact([]knownvalue.Check{knownvalue.ObjectExact(map[string]knownvalue.Check{
						"authorizer_configuration": knownvalue.ListSizeExact(0),
						"authorizer_type":          tfknownvalue.StringExact(awstypes.RegistryAuthorizerTypeAwsIam),
					})})),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrName), knownvalue.StringExact(rName1)),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("registry_arn"), checkRegistryARN),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("registry_id"), knownvalue.NotNull()),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrTags), knownvalue.Null()),
				},
			},
			{
				ConfigDirectory: config.StaticDirectory("testdata/Registry/basic/"),
				ConfigVariables: config.Variables{
					acctest.CtRName: config.StringVariable(rName1),
				},
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateVerify:                    true,
				ImportStateIdFunc:                    acctest.AttrImportStateIdFunc(resourceName, "registry_id"),
				ImportStateVerifyIdentifierAttribute: "registry_id",
			},
			{
				ConfigDirectory: config.StaticDirectory("testdata/Registry/basic/"),
				ConfigVariables: config.Variables{
					acctest.CtRName: config.StringVariable(rName2),
				},
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckRegistryExists(ctx, t, resourceName),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionUpdate),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("approval_configuration"), knownvalue.ListSizeExact(0)),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("auto_detection_configuration"), knownvalue.ListSizeExact(0)),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrDescription), knownvalue.Null()),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("discovery_configuration"), knownvalue.ListExact([]knownvalue.Check{knownvalue.ObjectExact(map[string]knownvalue.Check{
						"authorizer_configuration": knownvalue.ListSizeExact(0),
						"authorizer_type":          tfknownvalue.StringExact(awstypes.RegistryAuthorizerTypeAwsIam),
					})})),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrName), knownvalue.StringExact(rName2)),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("registry_arn"), checkRegistryARN),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("registry_id"), knownvalue.NotNull()),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrTags), knownvalue.Null()),
				},
			},
		},
	})
}

func TestAccAgentRegistryRegistry_description(t *testing.T) {
	ctx := acctest.Context(t)
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_agentregistry_registry.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.AgentRegistryServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckRegistryDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				ConfigDirectory: config.StaticDirectory("testdata/Registry/description/"),
				ConfigVariables: config.Variables{
					acctest.CtRName: config.StringVariable(rName),
					"description":   config.StringVariable("description1"),
				},
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckRegistryExists(ctx, t, resourceName),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrDescription), knownvalue.StringExact("description1")),
				},
			},
			{
				ConfigDirectory: config.StaticDirectory("testdata/Registry/description/"),
				ConfigVariables: config.Variables{
					acctest.CtRName: config.StringVariable(rName),
					"description":   config.StringVariable("description2"),
				},
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckRegistryExists(ctx, t, resourceName),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionUpdate),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrDescription), knownvalue.StringExact("description2")),
				},
			},
			{
				ConfigDirectory: config.StaticDirectory("testdata/Registry/basic/"),
				ConfigVariables: config.Variables{
					acctest.CtRName: config.StringVariable(rName),
				},
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckRegistryExists(ctx, t, resourceName),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionUpdate),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrDescription), knownvalue.Null()),
				},
			},
			{
				ConfigDirectory: config.StaticDirectory("testdata/Registry/description/"),
				ConfigVariables: config.Variables{
					acctest.CtRName: config.StringVariable(rName),
					"description":   config.StringVariable("description1"),
				},
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckRegistryExists(ctx, t, resourceName),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionUpdate),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrDescription), knownvalue.StringExact("description1")),
				},
			},
		},
	})
}

func TestAccAgentRegistryRegistry_approvalConfiguration(t *testing.T) {
	ctx := acctest.Context(t)
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_agentregistry_registry.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.AgentRegistryServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckRegistryDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				ConfigDirectory: config.StaticDirectory("testdata/Registry/approval_configuration/"),
				ConfigVariables: config.Variables{
					acctest.CtRName: config.StringVariable(rName),
				},
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckRegistryExists(ctx, t, resourceName),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("approval_configuration"), knownvalue.ListExact([]knownvalue.Check{knownvalue.ObjectExact(map[string]knownvalue.Check{
						"auto_approval_rules": knownvalue.SetExact([]knownvalue.Check{
							tfknownvalue.StringExact(awstypes.AutoApprovalRuleApproveAll),
						}),
					})})),
				},
			},
			{
				ConfigDirectory: config.StaticDirectory("testdata/Registry/approval_configuration/"),
				ConfigVariables: config.Variables{
					acctest.CtRName: config.StringVariable(rName),
				},
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateVerify:                    true,
				ImportStateIdFunc:                    acctest.AttrImportStateIdFunc(resourceName, "registry_id"),
				ImportStateVerifyIdentifierAttribute: "registry_id",
			},
			{
				ConfigDirectory: config.StaticDirectory("testdata/Registry/basic/"),
				ConfigVariables: config.Variables{
					acctest.CtRName: config.StringVariable(rName),
				},
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckRegistryExists(ctx, t, resourceName),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionUpdate),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("approval_configuration"), knownvalue.ListSizeExact(0)),
				},
			},
			{
				ConfigDirectory: config.StaticDirectory("testdata/Registry/approval_configuration/"),
				ConfigVariables: config.Variables{
					acctest.CtRName: config.StringVariable(rName),
				},
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckRegistryExists(ctx, t, resourceName),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionUpdate),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("approval_configuration"), knownvalue.ListExact([]knownvalue.Check{knownvalue.ObjectExact(map[string]knownvalue.Check{
						"auto_approval_rules": knownvalue.SetExact([]knownvalue.Check{
							tfknownvalue.StringExact(awstypes.AutoApprovalRuleApproveAll),
						}),
					})})),
				},
			},
		},
	})
}

func TestAccAgentRegistryRegistry_customJWTAuthorizerBasic(t *testing.T) {
	ctx := acctest.Context(t)
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_agentregistry_registry.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.AgentRegistryServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckRegistryDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				ConfigDirectory: config.StaticDirectory("testdata/Registry/custom_jwt_authorizer/"),
				ConfigVariables: config.Variables{
					acctest.CtRName: config.StringVariable(rName),
				},
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckRegistryExists(ctx, t, resourceName),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("discovery_configuration"), knownvalue.ListExact([]knownvalue.Check{knownvalue.ObjectExact(map[string]knownvalue.Check{
						"authorizer_configuration": knownvalue.ListExact([]knownvalue.Check{knownvalue.ObjectExact(map[string]knownvalue.Check{
							"custom_jwt_authorizer": knownvalue.ListExact([]knownvalue.Check{knownvalue.ObjectExact(map[string]knownvalue.Check{
								"allowed_audience": knownvalue.Null(),
								"allowed_clients": knownvalue.ListExact([]knownvalue.Check{
									knownvalue.StringExact("client1"),
								}),
								"allowed_scopes": knownvalue.Null(),
								"custom_claim":   knownvalue.SetSizeExact(0),
								"discovery_url":  knownvalue.StringExact("https://accounts.google.com/.well-known/openid-configuration"),
							})}),
						})}),
						"authorizer_type": tfknownvalue.StringExact(awstypes.RegistryAuthorizerTypeCustomJwt),
					})})),
				},
			},
			{
				ConfigDirectory: config.StaticDirectory("testdata/Registry/custom_jwt_authorizer/"),
				ConfigVariables: config.Variables{
					acctest.CtRName: config.StringVariable(rName),
				},
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateVerify:                    true,
				ImportStateIdFunc:                    acctest.AttrImportStateIdFunc(resourceName, "registry_id"),
				ImportStateVerifyIdentifierAttribute: "registry_id",
			},
		},
	})
}

func TestAccAgentRegistryRegistry_customJWTAuthorizerFull(t *testing.T) {
	ctx := acctest.Context(t)
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_agentregistry_registry.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.AgentRegistryServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckRegistryDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				ConfigDirectory: config.StaticDirectory("testdata/Registry/custom_jwt_authorizer.full/"),
				ConfigVariables: config.Variables{
					acctest.CtRName: config.StringVariable(rName),
				},
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckRegistryExists(ctx, t, resourceName),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("discovery_configuration"), knownvalue.ListExact([]knownvalue.Check{knownvalue.ObjectExact(map[string]knownvalue.Check{
						"authorizer_configuration": knownvalue.ListExact([]knownvalue.Check{knownvalue.ObjectExact(map[string]knownvalue.Check{
							"custom_jwt_authorizer": knownvalue.ListExact([]knownvalue.Check{knownvalue.ObjectExact(map[string]knownvalue.Check{
								"allowed_audience": knownvalue.ListExact([]knownvalue.Check{
									knownvalue.StringExact("audience-Z"),
									knownvalue.StringExact("audience-A"),
								}),
								"allowed_clients": knownvalue.ListExact([]knownvalue.Check{
									knownvalue.StringExact("client-Z"),
									knownvalue.StringExact("client-A"),
								}),
								"allowed_scopes": knownvalue.ListExact([]knownvalue.Check{
									knownvalue.StringExact("scope-99"),
									knownvalue.StringExact("scope-01"),
								}),
								"custom_claim": knownvalue.SetExact([]knownvalue.Check{
									knownvalue.MapExact(map[string]knownvalue.Check{
										"authorizing_claim_match_value": knownvalue.ListExact([]knownvalue.Check{knownvalue.ObjectExact(map[string]knownvalue.Check{
											"claim_match_operator": tfknownvalue.StringExact(awstypes.ClaimMatchOperatorTypeEquals),
											"claim_match_value": knownvalue.ListExact([]knownvalue.Check{knownvalue.ObjectExact(map[string]knownvalue.Check{
												"match_value_string":      knownvalue.StringExact("value01"),
												"match_value_string_list": knownvalue.Null(),
											})}),
										})}),
										"inbound_token_claim_name":       knownvalue.StringExact("claim_99"),
										"inbound_token_claim_value_type": tfknownvalue.StringExact(awstypes.InboundTokenClaimValueTypeString),
									}),
									knownvalue.MapExact(map[string]knownvalue.Check{
										"authorizing_claim_match_value": knownvalue.ListExact([]knownvalue.Check{knownvalue.ObjectExact(map[string]knownvalue.Check{
											"claim_match_operator": tfknownvalue.StringExact(awstypes.ClaimMatchOperatorTypeContainsAny),
											"claim_match_value": knownvalue.ListExact([]knownvalue.Check{knownvalue.ObjectExact(map[string]knownvalue.Check{
												"match_value_string": knownvalue.Null(),
												"match_value_string_list": knownvalue.SetExact([]knownvalue.Check{
													knownvalue.StringExact("value99"),
													knownvalue.StringExact("value01"),
													knownvalue.StringExact("12345"),
												}),
											})}),
										})}),
										"inbound_token_claim_name":       knownvalue.StringExact("claim_01"),
										"inbound_token_claim_value_type": tfknownvalue.StringExact(awstypes.InboundTokenClaimValueTypeStringArray),
									}),
								}),
								"discovery_url": knownvalue.StringExact("https://accounts.google.com/.well-known/openid-configuration"),
							})}),
						})}),
						"authorizer_type": tfknownvalue.StringExact(awstypes.RegistryAuthorizerTypeCustomJwt),
					})})),
				},
			},
			{
				ConfigDirectory: config.StaticDirectory("testdata/Registry/custom_jwt_authorizer.full/"),
				ConfigVariables: config.Variables{
					acctest.CtRName: config.StringVariable(rName),
				},
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateVerify:                    true,
				ImportStateIdFunc:                    acctest.AttrImportStateIdFunc(resourceName, "registry_id"),
				ImportStateVerifyIdentifierAttribute: "registry_id",
			},
		},
	})
}

func TestAccAgentRegistryRegistry_autoDetectionConfiguration(t *testing.T) {
	ctx := acctest.Context(t)
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_agentregistry_registry.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckOrganizationManagementAccount(ctx, t)
			// https://docs.aws.amazon.com/bedrock-agentcore/latest/devguide/registry-organizations.html.
			acctest.PreCheckIAMServiceLinkedRole(ctx, t, "/aws-service-role/agent-registry.amazonaws.com")
			acctest.PreCheckOrganizationsEnabledServicePrincipal(ctx, t, "agent-registry.amazonaws.com")
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.AgentRegistryServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckRegistryDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				ConfigDirectory: config.StaticDirectory("testdata/Registry/auto_detection_configuration/"),
				ConfigVariables: config.Variables{
					acctest.CtRName: config.StringVariable(rName),
					"enabled":       config.BoolVariable(true),
				},
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckRegistryExists(ctx, t, resourceName),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("auto_detection_configuration"), knownvalue.ListExact([]knownvalue.Check{knownvalue.ObjectExact(map[string]knownvalue.Check{
						names.AttrEnabled: knownvalue.Bool(true),
						names.AttrScope:   tfknownvalue.StringExact(awstypes.AutoDetectionScopeOrganization),
					})})),
				},
			},
			{
				ConfigDirectory: config.StaticDirectory("testdata/Registry/auto_detection_configuration/"),
				ConfigVariables: config.Variables{
					acctest.CtRName: config.StringVariable(rName),
					"enabled":       config.BoolVariable(true),
				},
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateVerify:                    true,
				ImportStateIdFunc:                    acctest.AttrImportStateIdFunc(resourceName, "registry_id"),
				ImportStateVerifyIdentifierAttribute: "registry_id",
			},
			{
				ConfigDirectory: config.StaticDirectory("testdata/Registry/auto_detection_configuration/"),
				ConfigVariables: config.Variables{
					acctest.CtRName: config.StringVariable(rName),
					"enabled":       config.BoolVariable(false),
				},
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckRegistryExists(ctx, t, resourceName),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionUpdate),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("auto_detection_configuration"), knownvalue.ListExact([]knownvalue.Check{knownvalue.ObjectExact(map[string]knownvalue.Check{
						names.AttrEnabled: knownvalue.Bool(false),
						names.AttrScope:   tfknownvalue.StringExact(awstypes.AutoDetectionScopeOrganization),
					})})),
				},
			},
			{
				ConfigDirectory: config.StaticDirectory("testdata/Registry/basic/"),
				ConfigVariables: config.Variables{
					acctest.CtRName: config.StringVariable(rName),
				},
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckRegistryExists(ctx, t, resourceName),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionUpdate),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("auto_detection_configuration"), knownvalue.ListSizeExact(0)),
				},
			},
			{
				ConfigDirectory: config.StaticDirectory("testdata/Registry/auto_detection_configuration/"),
				ConfigVariables: config.Variables{
					acctest.CtRName: config.StringVariable(rName),
					"enabled":       config.BoolVariable(true),
				},
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckRegistryExists(ctx, t, resourceName),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionUpdate),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("auto_detection_configuration"), knownvalue.ListExact([]knownvalue.Check{knownvalue.ObjectExact(map[string]knownvalue.Check{
						names.AttrEnabled: knownvalue.Bool(true),
						names.AttrScope:   tfknownvalue.StringExact(awstypes.AutoDetectionScopeOrganization),
					})})),
				},
			},
			{
				ConfigDirectory: config.StaticDirectory("testdata/Registry/auto_detection_configuration/"),
				ConfigVariables: config.Variables{
					acctest.CtRName: config.StringVariable(rName),
					"enabled":       config.BoolVariable(false),
				},
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckRegistryExists(ctx, t, resourceName),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionUpdate),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("auto_detection_configuration"), knownvalue.ListExact([]knownvalue.Check{knownvalue.ObjectExact(map[string]knownvalue.Check{
						names.AttrEnabled: knownvalue.Bool(false),
						names.AttrScope:   tfknownvalue.StringExact(awstypes.AutoDetectionScopeOrganization),
					})})),
				},
			},
		},
	})
}

func testAccCheckRegistryExists(ctx context.Context, t *testing.T, n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.ProviderMeta(ctx, t).AgentRegistryClient(ctx)

		_, err := tfagentregistry.FindRegistryByID(ctx, conn, rs.Primary.Attributes["registry_id"])

		return err
	}
}

func testAccCheckRegistryDestroy(ctx context.Context, t *testing.T) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.ProviderMeta(ctx, t).AgentRegistryClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_agentregistry_registry" {
				continue
			}

			_, err := tfagentregistry.FindRegistryByID(ctx, conn, rs.Primary.Attributes["registry_id"])

			if retry.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("Agent Registry Registry %s still exists", rs.Primary.Attributes["registry_id"])
		}

		return nil
	}
}
