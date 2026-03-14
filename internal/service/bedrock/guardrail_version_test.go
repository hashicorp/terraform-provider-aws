// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package bedrock_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/bedrock"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	tfbedrock "github.com/hashicorp/terraform-provider-aws/internal/service/bedrock"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccBedrockGuardrailVersion_basic(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var guardrailversion bedrock.GetGuardrailOutput
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_bedrock_guardrail_version.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.BedrockEndpointID)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.BedrockServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckGuardrailVersionDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccGuardrailVersion_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGuardrailVersionExists(ctx, t, resourceName, &guardrailversion),
					resource.TestCheckResourceAttr(resourceName, names.AttrVersion, "1"),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrDescription),
					resource.TestCheckResourceAttrSet(resourceName, "guardrail_arn"),
				),
			},
			{
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateIdFunc:                    testAccGuardrailVersionImportStateIDFunc(resourceName),
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: names.AttrVersion,
			},
		},
	})
}

func TestAccBedrockGuardrailVersion_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var guardrailversion bedrock.GetGuardrailOutput
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_bedrock_guardrail_version.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.BedrockEndpointID)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.BedrockServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckGuardrailVersionDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccGuardrailVersion_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGuardrailVersionExists(ctx, t, resourceName, &guardrailversion),
					acctest.CheckFrameworkResourceDisappears(ctx, t, tfbedrock.ResourceGuardrailVersion, resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccBedrockGuardrailVersion_skipDestroy(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_bedrock_guardrail_version.test"
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	var guardrailversion bedrock.GetGuardrailOutput
	acctest.Test(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.BedrockEndpointID)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.ServiceQuotasServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckGuardrailVersionDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccGuardrailVersion_skipDestroy(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGuardrailVersionExists(ctx, t, resourceName, &guardrailversion),
					resource.TestCheckResourceAttr(resourceName, names.AttrVersion, "1"),
				),
			},
			//Executes version resource again and validates first version exists
			{
				Config: testAccGuardrailVersion_skipDestroy(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGuardrailVersionExists(ctx, t, resourceName, &guardrailversion),
					resource.TestCheckResourceAttr(resourceName, names.AttrVersion, "1"),
				),
			},
		},
	})
}

func testAccCheckGuardrailVersionDestroy(ctx context.Context, t *testing.T) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.ProviderMeta(ctx, t).BedrockClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_bedrock_guardrail_version" {
				continue
			}

			_, err := tfbedrock.FindGuardrailByTwoPartKey(ctx, conn, rs.Primary.Attributes["guardrail_arn"], rs.Primary.Attributes[names.AttrVersion])

			if retry.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("Bedrock Guardrail Version %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccCheckGuardrailVersionExists(ctx context.Context, t *testing.T, n string, v *bedrock.GetGuardrailOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.ProviderMeta(ctx, t).BedrockClient(ctx)

		output, err := tfbedrock.FindGuardrailByTwoPartKey(ctx, conn, rs.Primary.Attributes["guardrail_arn"], rs.Primary.Attributes[names.AttrVersion])

		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

func testAccGuardrailVersionImportStateIDFunc(resourceName string) resource.ImportStateIdFunc {
	return func(s *terraform.State) (string, error) {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return "", fmt.Errorf("Not found: %s", resourceName)
		}
		return fmt.Sprintf("%s,%s", rs.Primary.Attributes["guardrail_arn"], rs.Primary.Attributes[names.AttrVersion]), nil
	}
}

func testAccGuardrailVersion_basic(rName string) string {
	return fmt.Sprintf(`
resource "aws_bedrock_guardrail" "test" {
  name                      = %[1]q
  blocked_input_messaging   = "test"
  blocked_outputs_messaging = "test"
  description               = "test"

  word_policy_config {
    managed_word_lists_config {
      type = "PROFANITY"
    }
    words_config {
      text = "HATE"
    }
  }
}

resource "aws_bedrock_guardrail_version" "test" {
  description   = %[1]q
  guardrail_arn = aws_bedrock_guardrail.test.guardrail_arn
}
`, rName)
}

func testAccGuardrailVersion_skipDestroy(rName string) string {
	return fmt.Sprintf(`
resource "aws_bedrock_guardrail" "test" {
  name                      = %[1]q
  blocked_input_messaging   = "test"
  blocked_outputs_messaging = "test"
  description               = "test"

  word_policy_config {
    managed_word_lists_config {
      type = "PROFANITY"
    }
    words_config {
      text = "HATE"
    }
  }
}

resource "aws_bedrock_guardrail_version" "test" {
  description   = %[1]q
  guardrail_arn = aws_bedrock_guardrail.test.guardrail_arn
  skip_destroy  = true
}
`, rName)
}
