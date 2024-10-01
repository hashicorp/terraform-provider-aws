// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package bedrock_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/bedrock"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfbedrock "github.com/hashicorp/terraform-provider-aws/internal/service/bedrock"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccBedrockGuardrailVersion_basic(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var guardrailversion bedrock.GetGuardrailOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_bedrock_guardrail_version.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.BedrockEndpointID)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.BedrockServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckGuardrailVersionDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccGuardrailVersion_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGuardrailVersionExists(ctx, resourceName, &guardrailversion),
					resource.TestCheckResourceAttr(resourceName, names.AttrVersion, acctest.Ct1),
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
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_bedrock_guardrail_version.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.BedrockEndpointID)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.BedrockServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckGuardrailVersionDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccGuardrailVersion_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGuardrailVersionExists(ctx, resourceName, &guardrailversion),
					acctest.CheckFrameworkResourceDisappears(ctx, acctest.Provider, tfbedrock.ResourceGuardrailVersion, resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccBedrockGuardrailVersion_skipDestroy(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_bedrock_guardrail_version.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	var guardrailversion bedrock.GetGuardrailOutput
	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.BedrockEndpointID)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.ServiceQuotasServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckGuardrailVersionDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccGuardrailVersion_skipDestroy(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGuardrailVersionExists(ctx, resourceName, &guardrailversion),
					resource.TestCheckResourceAttr(resourceName, names.AttrVersion, acctest.Ct1),
				),
			},
			//Executes version resource again and validates first version exists
			{
				Config: testAccGuardrailVersion_skipDestroy(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGuardrailVersionExists(ctx, resourceName, &guardrailversion),
					resource.TestCheckResourceAttr(resourceName, names.AttrVersion, acctest.Ct1),
				),
			},
		},
	})
}

func testAccCheckGuardrailVersionDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).BedrockClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_bedrock_guardrail_version" {
				continue
			}

			_, err := tfbedrock.FindGuardrailByTwoPartKey(ctx, conn, rs.Primary.Attributes["guardrail_arn"], rs.Primary.Attributes[names.AttrVersion])

			if tfresource.NotFound(err) {
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

func testAccCheckGuardrailVersionExists(ctx context.Context, n string, v *bedrock.GetGuardrailOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).BedrockClient(ctx)

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
