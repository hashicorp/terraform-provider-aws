// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package agentregistry_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/YakDriver/regexache"
	awstypes "github.com/aws/aws-sdk-go-v2/service/agentregistrycontrol/types"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	tfagentregistry "github.com/hashicorp/terraform-provider-aws/internal/service/agentregistry"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccAgentRegistryRegistry_basic(t *testing.T) {
	ctx := acctest.Context(t)
	rName := randomWithPrefixAndUnderscore(t)
	resourceName := "aws_agentregistry_registry.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.AgentRegistryServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckRegistryDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccRegistryConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckRegistryExists(ctx, t, resourceName),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttrSet(resourceName, "registry_id"),
					acctest.MatchResourceAttrRegionalARN(ctx, resourceName, "registry_arn", "agent-registry", regexache.MustCompile(`registry/.+`)),
					resource.TestCheckResourceAttr(resourceName, names.AttrStatus, string(awstypes.RegistryStatusReady)),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
					},
				},
			},
			{
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
	rName := randomWithPrefixAndUnderscore(t)
	resourceName := "aws_agentregistry_registry.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.AgentRegistryServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckRegistryDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccRegistryConfig_basic(rName),
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

func TestAccAgentRegistryRegistry_description(t *testing.T) {
	ctx := acctest.Context(t)
	rName := randomWithPrefixAndUnderscore(t)
	resourceName := "aws_agentregistry_registry.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.AgentRegistryServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckRegistryDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccRegistryConfig_description(rName, "initial description"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckRegistryExists(ctx, t, resourceName),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, "initial description"),
				),
			},
			{
				Config: testAccRegistryConfig_description(rName, "updated description"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckRegistryExists(ctx, t, resourceName),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, "updated description"),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionUpdate),
					},
				},
			},
			{
				Config: testAccRegistryConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckRegistryExists(ctx, t, resourceName),
					resource.TestCheckNoResourceAttr(resourceName, names.AttrDescription),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionUpdate),
					},
				},
			},
		},
	})
}

func TestAccAgentRegistryRegistry_approvalConfiguration(t *testing.T) {
	ctx := acctest.Context(t)
	rName := randomWithPrefixAndUnderscore(t)
	resourceName := "aws_agentregistry_registry.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.AgentRegistryServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckRegistryDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccRegistryConfig_approvalConfiguration(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckRegistryExists(ctx, t, resourceName),
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("approval_configuration").AtSliceIndex(0).AtMapKey("auto_approval_rules"), knownvalue.SetSizeExact(1)),
				},
			},
			{
				Config: testAccRegistryConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckRegistryExists(ctx, t, resourceName),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionUpdate),
					},
				},
			},
		},
	})
}

func TestAccAgentRegistryRegistry_discoveryIAM(t *testing.T) {
	ctx := acctest.Context(t)
	rName := randomWithPrefixAndUnderscore(t)
	resourceName := "aws_agentregistry_registry.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.AgentRegistryServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckRegistryDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccRegistryConfig_discoveryIAM(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckRegistryExists(ctx, t, resourceName),
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("discovery_configuration").AtSliceIndex(0).AtMapKey("authorizer_type"), knownvalue.StringExact("AWS_IAM")),
				},
			},
		},
	})
}

func TestAccAgentRegistryRegistry_authorizerConfiguration(t *testing.T) {
	ctx := acctest.Context(t)
	rName := randomWithPrefixAndUnderscore(t)
	resourceName := "aws_agentregistry_registry.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.AgentRegistryServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckRegistryDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccRegistryConfig_authorizerConfigurationCreate(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckRegistryExists(ctx, t, resourceName),
					resource.TestCheckResourceAttr(resourceName, "discovery_configuration.0.authorizer_type", "CUSTOM_JWT"),
					resource.TestCheckResourceAttr(resourceName, "discovery_configuration.0.authorizer_configuration.0.discovery_url", "https://accounts.google.com/.well-known/openid-configuration"),
					resource.TestCheckResourceAttr(resourceName, "discovery_configuration.0.authorizer_configuration.0.allowed_audience.0", "audience-1"),
					resource.TestCheckResourceAttr(resourceName, "discovery_configuration.0.authorizer_configuration.0.custom_claim.#", "1"),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("discovery_configuration").AtSliceIndex(0).AtMapKey("authorizer_type"), knownvalue.StringExact("CUSTOM_JWT")),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("discovery_configuration").AtSliceIndex(0).AtMapKey("authorizer_configuration").AtSliceIndex(0).AtMapKey("custom_claim"), knownvalue.SetSizeExact(1)),
				},
			},
			{
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateVerify:                    true,
				ImportStateIdFunc:                    acctest.AttrImportStateIdFunc(resourceName, "registry_id"),
				ImportStateVerifyIdentifierAttribute: "registry_id",
			},
			{
				Config: testAccRegistryConfig_authorizerConfigurationUpdate(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckRegistryExists(ctx, t, resourceName),
					resource.TestCheckResourceAttr(resourceName, "discovery_configuration.0.authorizer_configuration.0.allowed_audience.0", "audience-2"),
					resource.TestCheckResourceAttr(resourceName, "discovery_configuration.0.authorizer_configuration.0.custom_claim.#", "1"),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionUpdate),
					},
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

			return fmt.Errorf("Agent Registry Registry %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccRegistryConfig_basic(rName string) string {
	return fmt.Sprintf(`
resource "aws_agentregistry_registry" "test" {
  name = %[1]q

  discovery_configuration {
    authorizer_type = "AWS_IAM"
  }
}
`, rName)
}

func testAccRegistryConfig_description(rName, description string) string {
	return fmt.Sprintf(`
resource "aws_agentregistry_registry" "test" {
  name        = %[1]q
  description = %[2]q

  discovery_configuration {
    authorizer_type = "AWS_IAM"
  }
}
`, rName, description)
}

func testAccRegistryConfig_approvalConfiguration(rName string) string {
	return fmt.Sprintf(`
resource "aws_agentregistry_registry" "test" {
  name = %[1]q

  approval_configuration {
    auto_approval_rules = ["APPROVE_ALL"]
  }

  discovery_configuration {
    authorizer_type = "AWS_IAM"
  }
}
`, rName)
}

func testAccRegistryConfig_discoveryIAM(rName string) string {
	return testAccRegistryConfig_basic(rName)
}

func testAccRegistryConfig_authorizerConfigurationCreate(rName string) string {
	return fmt.Sprintf(`
resource "aws_agentregistry_registry" "test" {
  name = %[1]q

  discovery_configuration {
    authorizer_type = "CUSTOM_JWT"

    authorizer_configuration {
      discovery_url    = "https://accounts.google.com/.well-known/openid-configuration"
      allowed_audience = ["audience-1"]

      custom_claim {
        inbound_token_claim_name       = "sub"
        inbound_token_claim_value_type = "STRING"

        authorizing_claim_match_value {
          claim_match_operator = "EQUALS"

          claim_match_value {
            match_value_string = "test-user"
          }
        }
      }
    }
  }
}
`, rName)
}

func testAccRegistryConfig_authorizerConfigurationUpdate(rName string) string {
	return fmt.Sprintf(`
resource "aws_agentregistry_registry" "test" {
  name = %[1]q

  discovery_configuration {
    authorizer_type = "CUSTOM_JWT"

    authorizer_configuration {
      discovery_url    = "https://accounts.google.com/.well-known/openid-configuration"
      allowed_audience = ["audience-2"]

      custom_claim {
        inbound_token_claim_name       = "sub"
        inbound_token_claim_value_type = "STRING"

        authorizing_claim_match_value {
          claim_match_operator = "EQUALS"

          claim_match_value {
            match_value_string = "test-user"
          }
        }
      }
    }
  }
}
`, rName)
}

func randomWithPrefixAndUnderscore(t *testing.T) string {
	// Several descriptive test names exceed the API's 64-character registry
	// name limit before the random suffix is appended.
	return acctest.RandomWithPrefix(t, "tf-acc-agentregistry")
}
