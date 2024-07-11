// // Copyright (c) HashiCorp, Inc.
// // SPDX-License-Identifier: MPL-2.0

package bedrock_test

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/bedrock"
	"github.com/aws/aws-sdk-go-v2/service/bedrock/types"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	tfbedrock "github.com/hashicorp/terraform-provider-aws/internal/service/bedrock"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccBedrockGuardrail_basic(t *testing.T) {
	ctx := acctest.Context(t)
	// TIP: This is a long-running test guard for tests that run longer than
	// 300s (5 min) generally.
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var guardrail bedrock.GetGuardrailOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_bedrock_guardrail.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.BedrockEndpointID)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.BedrockEndpointID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckGuardrailDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccGuardrailConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGuardrailExists(ctx, resourceName, &guardrail),
					resource.TestCheckResourceAttrSet(resourceName, "arn"),
					resource.TestCheckResourceAttr(resourceName, "blocked_input_messaging", "test"),
					resource.TestCheckResourceAttr(resourceName, "blocked_outputs_messaging", "test"),
					resource.TestCheckResourceAttr(resourceName, "content_policy_config.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "content_policy_config.0.filters_config.#", acctest.Ct2),
					resource.TestCheckResourceAttrSet(resourceName, "created_at"),
					resource.TestCheckResourceAttr(resourceName, "description", "test"),
					resource.TestCheckNoResourceAttr(resourceName, "kms_key_arn"),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "sensitive_information_policy_config.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "sensitive_information_policy_config.0.pii_entities_config.#", acctest.Ct3),
					resource.TestCheckResourceAttr(resourceName, "sensitive_information_policy_config.0.regexes_config.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "status", "READY"),
					resource.TestCheckResourceAttr(resourceName, "topic_policy_config.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "topic_policy_config.0.topics_config.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "version", "DRAFT"),
					resource.TestCheckResourceAttr(resourceName, "word_policy_config.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "word_policy_config.0.managed_word_lists_config.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "word_policy_config.0.words_config.#", acctest.Ct1),
				),
			},
			{
				ResourceName:            resourceName,
				ImportStateIdFunc:       testAccGuardrailImportStateIDFunc(ctx, resourceName),
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"apply_immediately", "user"},
			},
		},
	})
}

func TestAccBedrockGuardrail_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var guardrail bedrock.GetGuardrailOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_bedrock_guardrail.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.BedrockEndpointID)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.BedrockEndpointID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckGuardrailDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccGuardrailConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGuardrailExists(ctx, resourceName, &guardrail),
					acctest.CheckFrameworkResourceDisappears(ctx, acctest.Provider, tfbedrock.ResourceGuardrail, resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccBedrockGuardrail_kmsKey(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_bedrock_guardrail.test"
	var guardrail bedrock.GetGuardrailOutput

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.BedrockEndpointID)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.BedrockEndpointID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckGuardrailDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccGuardrailConfig_kmsKey(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGuardrailExists(ctx, resourceName, &guardrail),
					resource.TestCheckResourceAttrSet(resourceName, "kms_key_arn"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportStateIdFunc:       testAccGuardrailImportStateIDFunc(ctx, resourceName),
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"apply_immediately", "user"},
			},
		},
	})
}

func TestAccBedrockGuardrail_tags(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_bedrock_guardrail.test"
	var guardrail bedrock.GetGuardrailOutput

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.BedrockEndpointID)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.BedrockEndpointID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckGuardrailDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccGuardrailConfig_tags(rName, acctest.CtKey1, acctest.CtValue1, acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGuardrailExists(ctx, resourceName, &guardrail),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct2),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey1, acctest.CtValue1),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey2, acctest.CtValue2),
				),
			},
			{
				Config: testAccGuardrailConfig_tags(rName, acctest.CtKey1, acctest.CtValue1Updated, acctest.CtKey2, acctest.CtValue1Updated),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGuardrailExists(ctx, resourceName, &guardrail),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct2),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey1, acctest.CtValue1Updated),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey2, acctest.CtValue1Updated),
				),
			},
		},
	})
}

func TestAccBedrockGuardrail_update(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_bedrock_guardrail.test"
	var guardrail bedrock.GetGuardrailOutput

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.BedrockEndpointID)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.BedrockEndpointID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckGuardrailDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccGuardrailConfig_wordConfig_only(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGuardrailExists(ctx, resourceName, &guardrail),
					resource.TestCheckResourceAttrSet(resourceName, "arn"),
					resource.TestCheckResourceAttr(resourceName, "blocked_input_messaging", "test"),
					resource.TestCheckResourceAttr(resourceName, "blocked_outputs_messaging", "test"),
					resource.TestCheckResourceAttr(resourceName, "content_policy_config.#", acctest.Ct0),
					resource.TestCheckResourceAttrSet(resourceName, "created_at"),
					resource.TestCheckResourceAttr(resourceName, "description", "test"),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "sensitive_information_policy_config.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "status", "READY"),
					resource.TestCheckResourceAttr(resourceName, "topic_policy_config.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "version", "DRAFT"),
					resource.TestCheckResourceAttr(resourceName, "word_policy_config.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "word_policy_config.0.managed_word_lists_config.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "word_policy_config.0.words_config.#", acctest.Ct1),
				),
			},
			{
				Config: testAccGuardrailConfig_update(rName, "test", "test", "MEDIUM", "^\\d{3}-\\d{2}-\\d{4}$", "NAME", "investment_topic", "HATE"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGuardrailExists(ctx, resourceName, &guardrail),
					resource.TestCheckResourceAttr(resourceName, "blocked_input_messaging", "test"),
					resource.TestCheckResourceAttr(resourceName, "blocked_outputs_messaging", "test"),
					resource.TestCheckResourceAttr(resourceName, "content_policy_config.0.filters_config.0.input_strength", "MEDIUM"),
					resource.TestCheckResourceAttr(resourceName, "sensitive_information_policy_config.0.regexes_config.0.pattern", "^\\d{3}-\\d{2}-\\d{4}$"),
					resource.TestCheckResourceAttr(resourceName, "sensitive_information_policy_config.0.pii_entities_config.0.type", "NAME"),
					resource.TestCheckResourceAttr(resourceName, "topic_policy_config.0.topics_config.0.name", "investment_topic"),
					resource.TestCheckResourceAttr(resourceName, "word_policy_config.0.words_config.0.text", "HATE"),
				),
			},
			{
				Config: testAccGuardrailConfig_update(rName, "update", "update", "HIGH", "^\\d{4}-\\d{2}-\\d{4}$", "USERNAME", "earnings_topic", "HATRED"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGuardrailExists(ctx, resourceName, &guardrail),
					resource.TestCheckResourceAttr(resourceName, "blocked_input_messaging", "update"),
					resource.TestCheckResourceAttr(resourceName, "blocked_outputs_messaging", "update"),
					resource.TestCheckResourceAttr(resourceName, "content_policy_config.0.filters_config.0.input_strength", "HIGH"),
					resource.TestCheckResourceAttr(resourceName, "sensitive_information_policy_config.0.regexes_config.0.pattern", "^\\d{4}-\\d{2}-\\d{4}$"),
					resource.TestCheckResourceAttr(resourceName, "sensitive_information_policy_config.0.pii_entities_config.0.type", "USERNAME"),
					resource.TestCheckResourceAttr(resourceName, "topic_policy_config.0.topics_config.0.name", "earnings_topic"),
					resource.TestCheckResourceAttr(resourceName, "word_policy_config.0.words_config.0.text", "HATRED"),
				),
			},
		},
	})
}

func testAccGuardrailImportStateIDFunc(ctx context.Context, n string) resource.ImportStateIdFunc {
	return func(s *terraform.State) (string, error) {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return "", fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).BedrockClient(ctx)

		version := rs.Primary.Attributes[names.AttrVersion]
		id := rs.Primary.ID
		_, err := tfbedrock.FindGuardrailByID(ctx, conn, id, version)

		if err != nil {
			return "", err
		}

		return fmt.Sprintf("%s:%s", id, version), nil
	}
}

func testAccCheckGuardrailDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).BedrockClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_bedrock_guardrail" {
				continue
			}

			_, err := conn.GetGuardrail(ctx, &bedrock.GetGuardrailInput{
				GuardrailIdentifier: aws.String(rs.Primary.ID),
				GuardrailVersion:    aws.String(rs.Primary.Attributes[names.AttrVersion]),
			})
			if errs.IsA[*types.ResourceNotFoundException](err) {
				return nil
			}
			if err != nil {
				return create.Error(names.Bedrock, create.ErrActionCheckingDestroyed, tfbedrock.ResNameGuardrail, rs.Primary.ID, err)
			}

			return create.Error(names.Bedrock, create.ErrActionCheckingDestroyed, tfbedrock.ResNameGuardrail, rs.Primary.ID, errors.New("not destroyed"))
		}

		return nil
	}
}

func testAccCheckGuardrailExists(ctx context.Context, name string, guardrail *bedrock.GetGuardrailOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return create.Error(names.Bedrock, create.ErrActionCheckingExistence, tfbedrock.ResNameGuardrail, name, errors.New("not found"))
		}

		if rs.Primary.ID == "" {
			return create.Error(names.Bedrock, create.ErrActionCheckingExistence, tfbedrock.ResNameGuardrail, name, errors.New("not set"))
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).BedrockClient(ctx)
		resp, err := conn.GetGuardrail(ctx, &bedrock.GetGuardrailInput{
			GuardrailIdentifier: aws.String(rs.Primary.ID),
			GuardrailVersion:    aws.String(rs.Primary.Attributes[names.AttrVersion]),
		})

		if err != nil {
			return create.Error(names.Bedrock, create.ErrActionCheckingExistence, tfbedrock.ResNameGuardrail, rs.Primary.ID, err)
		}

		*guardrail = *resp

		return nil
	}
}

func testAccPreCheck(ctx context.Context, t *testing.T) {
	conn := acctest.Provider.Meta().(*conns.AWSClient).BedrockClient(ctx)

	input := &bedrock.ListGuardrailsInput{}
	_, err := conn.ListGuardrails(ctx, input)

	if acctest.PreCheckSkipError(err) {
		t.Skipf("skipping acceptance testing: %s", err)
	}
	if err != nil {
		t.Fatalf("unexpected PreCheck error: %s", err)
	}
}

// func testAccCheckGuardrailNotRecreated(before, after *bedrock.GetGuardrailOutput) resource.TestCheckFunc {
// 	return func(s *terraform.State) error {
// 		if before, after := aws.ToString(before.GuardrailId), aws.ToString(after.GuardrailId); before != after {
// 			return create.Error(names.Bedrock, create.ErrActionCheckingNotRecreated, tfbedrock.ResNameGuardrail, aws.ToString(before.GuardrailId), errors.New("recreated"))
// 		}

// 		return nil
// 	}
// }

func testAccGuardrailConfig_basic(rName string) string {
	return fmt.Sprintf(`

resource "aws_bedrock_guardrail" "test" {
  name                      = %[1]q
  blocked_input_messaging   = "test"
  blocked_outputs_messaging = "test"
  description               = "test"
  content_policy_config {
    filters_config {
      input_strength  = "MEDIUM"
      output_strength = "MEDIUM"
      type            = "HATE"
    }
    filters_config {
      input_strength  = "HIGH"
      output_strength = "HIGH"
      type            = "VIOLENCE"
    }
  }
  contextual_grounding_policy_config {
    filters_config {
      threshold  = 0.4
      type       = "GROUNDING"
    }
  }
  sensitive_information_policy_config {
    pii_entities_config {
      action = "BLOCK"
      type   = "NAME"
    }
    pii_entities_config {
      action = "BLOCK"
      type   = "DRIVER_ID"
    }
    pii_entities_config {
      action = "ANONYMIZE"
      type   = "USERNAME"
    }

    regexes_config {
      action      = "BLOCK"
      description = "example regex"
      name        = "regex_example"
      pattern     = "^\\d{3}-\\d{2}-\\d{4}$"
    }
  }
  topic_policy_config {
    topics_config {
      name       = "investment_topic"
      examples   = ["Where should I invest my money ?"]
      type       = "DENY"
      definition = "Investment advice refers to inquiries, guidance, or recommendations regarding the management or allocation of funds or assets with the goal of generating returns ."
    }
  }
  word_policy_config {
    managed_word_lists_config {
      type = "PROFANITY"
    }
    words_config {
      text = "HATE"
    }
  }
  tags = {
    "Modified By" = "terraform"
  }
}
`, rName)
}

func testAccGuardrailConfig_kmsKey(rName string) string {
	return acctest.ConfigCompose(testAccCustomModelConfig_base(rName), fmt.Sprintf(`
resource "aws_kms_key" "test" {
  description             = %[1]q
  deletion_window_in_days = 7
}

resource "aws_bedrock_guardrail" "test" {
  name                      = %[1]q
  blocked_input_messaging   = "test"
  blocked_outputs_messaging = "test"
  description               = "test"
  kms_key_arn   			= aws_kms_key.test.arn
  content_policy_config {
    filters_config {
      input_strength  = "MEDIUM"
      output_strength = "MEDIUM"
      type            = "HATE"
    }
    filters_config {
      input_strength  = "HIGH"
      output_strength = "HIGH"
      type            = "VIOLENCE"
    }
  }
  word_policy_config {
    managed_word_lists_config {
      type = "PROFANITY"
    }
    words_config {
      text = "HATE"
    }
  }
  tags = {
    "Modified By" = "terraform"
  }
}
`, rName))
}

func testAccGuardrailConfig_tags(rName, tagKey1, tagValue1, tagKey2, tagValue2 string) string {
	return acctest.ConfigCompose(testAccCustomModelConfig_base(rName), fmt.Sprintf(`
resource "aws_kms_key" "test" {
  description             = %[1]q
  deletion_window_in_days = 7
}

resource "aws_bedrock_guardrail" "test" {
  name                      = %[1]q
  blocked_input_messaging   = "test"
  blocked_outputs_messaging = "test"
  description               = "test"
  kms_key_arn   			= aws_kms_key.test.arn
  content_policy_config {
    filters_config {
      input_strength  = "MEDIUM"
      output_strength = "MEDIUM"
      type            = "HATE"
    }
    filters_config {
      input_strength  = "HIGH"
      output_strength = "HIGH"
      type            = "VIOLENCE"
    }
  }
  word_policy_config {
    managed_word_lists_config {
      type = "PROFANITY"
    }
    words_config {
      text = "HATE"
    }
  }
  tags = {
    %[2]q = %[3]q
	%[4]q = %[5]q
  }
}
`, rName, tagKey1, tagValue1, tagKey2, tagValue2))
}

func testAccGuardrailConfig_update(rName, blockedInputMessaging, blockedOutputMessaging, inputStrength, regexPattern, piiType, topicName, wordConfig string) string {
	return fmt.Sprintf(`

resource "aws_bedrock_guardrail" "test" {
  name                      = %[1]q
  blocked_input_messaging   = %[2]q
  blocked_outputs_messaging = %[3]q
  description               = "test"
  content_policy_config {
    filters_config {
      input_strength  = %[4]q
      output_strength = "MEDIUM"
      type            = "HATE"
    }
  }
  sensitive_information_policy_config {
    pii_entities_config {
      action = "BLOCK"
      type   = %[6]q
    }

    regexes_config {
      action      = "BLOCK"
      description = "example regex"
      name        = "regex_example"
      pattern     = %[5]q
    }
  }
  topic_policy_config {
    topics_config {
      name       = %[7]q
      examples   = ["Where should I invest my money ?"]
      type       = "DENY"
      definition = "Investment advice refers to inquiries, guidance, or recommendations regarding the management or allocation of funds or assets with the goal of generating returns ."
    }
  }
  word_policy_config {
    managed_word_lists_config {
      type = "PROFANITY"
    }
    words_config {
      text = %[8]q
    }
  }
  tags = {
    "Modified By" = "terraform"
  }
}
`, rName, blockedInputMessaging, blockedOutputMessaging, inputStrength, regexPattern, piiType, topicName, wordConfig)
}

func testAccGuardrailConfig_wordConfig_only(rName string) string {
	return acctest.ConfigCompose(testAccCustomModelConfig_base(rName), fmt.Sprintf(`
resource "aws_kms_key" "test" {
  description             = %[1]q
  deletion_window_in_days = 7
}

resource "aws_bedrock_guardrail" "test" {
  name                      = %[1]q
  blocked_input_messaging   = "test"
  blocked_outputs_messaging = "test"
  description               = "test"
  kms_key_arn   			= aws_kms_key.test.arn
  word_policy_config {
    managed_word_lists_config {
      type = "PROFANITY"
    }
    words_config {
      text = "HATE"
    }
  }
  tags = {
    "Modified By" = "terraform"
  }
}
`, rName))
}
