// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package bedrockagentcore_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	tfbedrockagentcore "github.com/hashicorp/terraform-provider-aws/internal/service/bedrockagentcore"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccBedrockAgentCoreConfigurationBundle_basic(t *testing.T) {
	ctx := acctest.Context(t)
	rName := randomWithPrefixAndUnderscore(t)
	resourceName := "aws_bedrockagentcore_configuration_bundle.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.BedrockEndpointID)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.BedrockAgentCoreServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckConfigurationBundleDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccConfigurationBundleConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckConfigurationBundleExists(ctx, t, resourceName),
					resource.TestCheckResourceAttr(resourceName, "bundle_name", rName),
					resource.TestCheckResourceAttrSet(resourceName, "bundle_id"),
					resource.TestCheckResourceAttrSet(resourceName, "bundle_arn"),
					resource.TestCheckResourceAttrSet(resourceName, "version_id"),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrCreatedAt),
					resource.TestCheckResourceAttr(resourceName, "component.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "lineage_metadata.#", "1"),
				),
			},
			{
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateIdFunc:                    acctest.AttrImportStateIdFunc(resourceName, "bundle_id"),
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: "bundle_id",
				// Write-only version-metadata inputs are not returned by the API at
				// the top level (they surface only in lineage_metadata).
				ImportStateVerifyIgnore: []string{"branch_name", "commit_message", "created_by"},
			},
		},
	})
}

func TestAccBedrockAgentCoreConfigurationBundle_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	rName := randomWithPrefixAndUnderscore(t)
	resourceName := "aws_bedrockagentcore_configuration_bundle.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.BedrockEndpointID)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.BedrockAgentCoreServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckConfigurationBundleDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccConfigurationBundleConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckConfigurationBundleExists(ctx, t, resourceName),
					acctest.CheckFrameworkResourceDisappears(ctx, t, tfbedrockagentcore.ResourceConfigurationBundle, resourceName),
				),
				ExpectNonEmptyPlan: true,
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PostApplyPostRefresh: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
					},
				},
			},
		},
	})
}

func TestAccBedrockAgentCoreConfigurationBundle_update(t *testing.T) {
	ctx := acctest.Context(t)
	rName := randomWithPrefixAndUnderscore(t)
	resourceName := "aws_bedrockagentcore_configuration_bundle.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.BedrockEndpointID)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.BedrockAgentCoreServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckConfigurationBundleDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccConfigurationBundleConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckConfigurationBundleExists(ctx, t, resourceName),
					resource.TestCheckNoResourceAttr(resourceName, names.AttrDescription),
					resource.TestCheckResourceAttr(resourceName, "component.#", "1"),
				),
			},
			{
				// Change the component configuration (cuts a new version) and add a
				// description. The Update path must supply parentVersionIds +
				// commitMessage, then re-read the advanced version_id.
				Config: testAccConfigurationBundleConfig_updated(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckConfigurationBundleExists(ctx, t, resourceName),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, "updated bundle"),
					resource.TestCheckResourceAttr(resourceName, "component.#", "2"),
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

func TestAccBedrockAgentCoreConfigurationBundle_kmsKey(t *testing.T) {
	ctx := acctest.Context(t)
	rName := randomWithPrefixAndUnderscore(t)
	resourceName := "aws_bedrockagentcore_configuration_bundle.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.BedrockEndpointID)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.BedrockAgentCoreServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckConfigurationBundleDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccConfigurationBundleConfig_kmsKey(rName, 0),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckConfigurationBundleExists(ctx, t, resourceName),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrKMSKeyARN, "aws_kms_key.test.0", names.AttrARN),
				),
			},
			{
				// Rotate to a different KMS key. UpdateConfigurationBundle accepts a
				// changed kmsKeyArn, so this is an in-place update, not a replacement.
				Config: testAccConfigurationBundleConfig_kmsKey(rName, 1),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckConfigurationBundleExists(ctx, t, resourceName),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrKMSKeyARN, "aws_kms_key.test.1", names.AttrARN),
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

func testAccCheckConfigurationBundleDestroy(ctx context.Context, t *testing.T) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.ProviderMeta(ctx, t).BedrockAgentCoreClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_bedrockagentcore_configuration_bundle" {
				continue
			}

			_, err := tfbedrockagentcore.FindConfigurationBundleByID(ctx, conn, rs.Primary.Attributes["bundle_id"], nil)
			if retry.NotFound(err) {
				continue
			}
			if err != nil {
				return err
			}
			return fmt.Errorf("Bedrock AgentCore Configuration Bundle %s still exists", rs.Primary.Attributes["bundle_id"])
		}

		return nil
	}
}

func testAccCheckConfigurationBundleExists(ctx context.Context, t *testing.T, n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.ProviderMeta(ctx, t).BedrockAgentCoreClient(ctx)

		_, err := tfbedrockagentcore.FindConfigurationBundleByID(ctx, conn, rs.Primary.Attributes["bundle_id"], nil)
		return err
	}
}

func testAccConfigurationBundleConfig_basic(rName string) string {
	return fmt.Sprintf(`
resource "aws_bedrockagentcore_configuration_bundle" "test" {
  bundle_name = %[1]q

  component {
    component_identifier = "arn:aws:bedrock-agentcore:::evaluator/Builtin.Helpfulness"
    configuration        = jsonencode({ threshold = 0.7 })
  }
}
`, rName)
}

func testAccConfigurationBundleConfig_kmsKey(rName string, keyIndex int) string {
	return fmt.Sprintf(`
resource "aws_kms_key" "test" {
  count                   = 2
  description             = "%[1]s-${count.index}"
  deletion_window_in_days = 7
}

resource "aws_bedrockagentcore_configuration_bundle" "test" {
  bundle_name = %[1]q
  kms_key_arn = aws_kms_key.test[%[2]d].arn

  component {
    component_identifier = "arn:aws:bedrock-agentcore:::evaluator/Builtin.Helpfulness"
    configuration        = jsonencode({ threshold = 0.7 })
  }
}
`, rName, keyIndex)
}

func testAccConfigurationBundleConfig_updated(rName string) string {
	return fmt.Sprintf(`
resource "aws_bedrockagentcore_configuration_bundle" "test" {
  bundle_name = %[1]q
  description = "updated bundle"

  component {
    component_identifier = "arn:aws:bedrock-agentcore:::evaluator/Builtin.Helpfulness"
    configuration        = jsonencode({ threshold = 0.9 })
  }

  component {
    component_identifier = "arn:aws:bedrock-agentcore:::evaluator/Builtin.GoalSuccessRate"
    configuration        = jsonencode({ threshold = 0.5 })
  }
}
`, rName)
}
