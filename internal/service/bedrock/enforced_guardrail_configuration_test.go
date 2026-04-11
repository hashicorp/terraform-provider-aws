// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package bedrock_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	tfbedrock "github.com/hashicorp/terraform-provider-aws/internal/service/bedrock"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func testAccEnforcedGuardrailConfiguration_basic(t *testing.T) {
	ctx := acctest.Context(t)
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_bedrock_enforced_guardrail_configuration.test"
	guardrailResourceName := "aws_bedrock_guardrail.test"

	acctest.Test(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckPartitionHasService(t, names.BedrockEndpointID) },
		ErrorCheck:               acctest.ErrorCheck(t, names.BedrockServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckEnforcedGuardrailConfigurationDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccEnforcedGuardrailConfigurationConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckEnforcedGuardrailConfigurationExists(ctx, t, resourceName),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrID),
					resource.TestCheckResourceAttrSet(resourceName, "config_id"),
					resource.TestCheckResourceAttrPair(resourceName, "guardrail_arn", guardrailResourceName, "guardrail_arn"),
					resource.TestCheckResourceAttrSet(resourceName, "guardrail_id"),
					resource.TestCheckResourceAttr(resourceName, "guardrail_version", "DRAFT"),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrCreatedAt),
					resource.TestCheckResourceAttrSet(resourceName, "created_by"),
					resource.TestCheckResourceAttrSet(resourceName, "updated_at"),
					resource.TestCheckResourceAttrSet(resourceName, "updated_by"),
					resource.TestCheckResourceAttr(resourceName, "owner", "ACCOUNT"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"guardrail_identifier"},
			},
		},
	})
}

func testAccEnforcedGuardrailConfiguration_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_bedrock_enforced_guardrail_configuration.test"

	acctest.Test(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckPartitionHasService(t, names.BedrockEndpointID) },
		ErrorCheck:               acctest.ErrorCheck(t, names.BedrockServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckEnforcedGuardrailConfigurationDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccEnforcedGuardrailConfigurationConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckEnforcedGuardrailConfigurationExists(ctx, t, resourceName),
					acctest.CheckFrameworkResourceDisappears(ctx, t, tfbedrock.ResourceEnforcedGuardrailConfiguration, resourceName),
				),
				ExpectNonEmptyPlan: true,
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
					},
				},
			},
		},
	})
}

func testAccEnforcedGuardrailConfiguration_selectiveContentGuarding(t *testing.T) {
	ctx := acctest.Context(t)
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_bedrock_enforced_guardrail_configuration.test"

	acctest.Test(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckPartitionHasService(t, names.BedrockEndpointID) },
		ErrorCheck:               acctest.ErrorCheck(t, names.BedrockServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckEnforcedGuardrailConfigurationDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccEnforcedGuardrailConfigurationConfig_selectiveContentGuarding(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckEnforcedGuardrailConfigurationExists(ctx, t, resourceName),
					resource.TestCheckResourceAttr(resourceName, "selective_content_guarding.0.messages", "COMPREHENSIVE"),
					resource.TestCheckResourceAttr(resourceName, "selective_content_guarding.0.system", "COMPREHENSIVE"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"guardrail_identifier"},
			},
		},
	})
}

func testAccEnforcedGuardrailConfiguration_modelEnforcement(t *testing.T) {
	ctx := acctest.Context(t)
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_bedrock_enforced_guardrail_configuration.test"

	acctest.Test(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckPartitionHasService(t, names.BedrockEndpointID) },
		ErrorCheck:               acctest.ErrorCheck(t, names.BedrockServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckEnforcedGuardrailConfigurationDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccEnforcedGuardrailConfigurationConfig_modelEnforcement(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckEnforcedGuardrailConfigurationExists(ctx, t, resourceName),
					resource.TestCheckResourceAttr(resourceName, "model_enforcement.0.included_models.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "model_enforcement.0.included_models.0", "ALL"),
					resource.TestCheckResourceAttr(resourceName, "model_enforcement.0.excluded_models.#", "0"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"guardrail_identifier"},
			},
		},
	})
}

func testAccEnforcedGuardrailConfiguration_update(t *testing.T) {
	ctx := acctest.Context(t)
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_bedrock_enforced_guardrail_configuration.test"

	acctest.Test(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckPartitionHasService(t, names.BedrockEndpointID) },
		ErrorCheck:               acctest.ErrorCheck(t, names.BedrockServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckEnforcedGuardrailConfigurationDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccEnforcedGuardrailConfigurationConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckEnforcedGuardrailConfigurationExists(ctx, t, resourceName),
				),
			},
			{
				Config: testAccEnforcedGuardrailConfigurationConfig_selectiveContentGuarding(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckEnforcedGuardrailConfigurationExists(ctx, t, resourceName),
					resource.TestCheckResourceAttr(resourceName, "selective_content_guarding.0.messages", "COMPREHENSIVE"),
					resource.TestCheckResourceAttr(resourceName, "selective_content_guarding.0.system", "COMPREHENSIVE"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"guardrail_identifier"},
			},
		},
	})
}

func testAccCheckEnforcedGuardrailConfigurationExists(ctx context.Context, t *testing.T, n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		_, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.ProviderMeta(ctx, t).BedrockClient(ctx)

		_, err := tfbedrock.FindEnforcedGuardrailConfiguration(ctx, conn)

		return err
	}
}

func testAccCheckEnforcedGuardrailConfigurationDestroy(ctx context.Context, t *testing.T) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.ProviderMeta(ctx, t).BedrockClient(ctx)
		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_bedrock_enforced_guardrail_configuration" {
				continue
			}

			_, err := tfbedrock.FindEnforcedGuardrailConfiguration(ctx, conn)

			if retry.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("Bedrock Enforced Guardrail Configuration %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccEnforcedGuardrailConfigurationConfig_base(rName string) string {
	return fmt.Sprintf(`
resource "aws_bedrock_guardrail" "test" {
  name                      = %[1]q
  blocked_input_messaging   = "Blocked input"
  blocked_outputs_messaging = "Blocked output"
  description               = "Test guardrail for enforced guardrail configuration"

  word_policy_config {
    words_config {
      text = "deny"
    }
  }
}
`, rName)
}

func testAccEnforcedGuardrailConfigurationConfig_basic(rName string) string {
	return acctest.ConfigCompose(testAccEnforcedGuardrailConfigurationConfig_base(rName), `
resource "aws_bedrock_enforced_guardrail_configuration" "test" {
  guardrail_identifier = aws_bedrock_guardrail.test.guardrail_arn
  guardrail_version    = "DRAFT"
}
`)
}

func testAccEnforcedGuardrailConfigurationConfig_selectiveContentGuarding(rName string) string {
	return acctest.ConfigCompose(testAccEnforcedGuardrailConfigurationConfig_base(rName), `
resource "aws_bedrock_enforced_guardrail_configuration" "test" {
  guardrail_identifier = aws_bedrock_guardrail.test.guardrail_arn
  guardrail_version    = "DRAFT"

  selective_content_guarding {
    messages = "COMPREHENSIVE"
    system   = "COMPREHENSIVE"
  }
}
`)
}

func testAccEnforcedGuardrailConfigurationConfig_modelEnforcement(rName string) string {
	return acctest.ConfigCompose(testAccEnforcedGuardrailConfigurationConfig_base(rName), `
resource "aws_bedrock_enforced_guardrail_configuration" "test" {
  guardrail_identifier = aws_bedrock_guardrail.test.guardrail_arn
  guardrail_version    = "DRAFT"

  model_enforcement {
    included_models = ["ALL"]
    excluded_models = []
  }
}
`)
}
