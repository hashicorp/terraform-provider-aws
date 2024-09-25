// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package bedrock_test

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/bedrock"
	"github.com/aws/aws-sdk-go-v2/service/bedrock/types"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/names"

	// TIP: You will often need to import the package that this test file lives
	// in. Since it is in the "test" context, it must import the package to use
	// any normal context constants, variables, or functions.
	tfbedrock "github.com/hashicorp/terraform-provider-aws/internal/service/bedrock"
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
					testAccCheckGuardrailVersionExists(ctx, resourceName, acctest.Ct1, &guardrailversion),
					resource.TestCheckResourceAttr(resourceName, "version", acctest.Ct1),
					resource.TestCheckResourceAttrSet(resourceName, "description"),
					resource.TestCheckResourceAttrSet(resourceName, "guardrail_identifier"),
				),
			},
			{
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateIdFunc:                    testAccGuardrailVersionImportStateIDFunc(resourceName),
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: "version",
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
					testAccCheckGuardrailVersionExists(ctx, resourceName, acctest.Ct1, &guardrailversion),
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
					testAccCheckGuardrailVersionExists(ctx, resourceName, acctest.Ct1, &guardrailversion),
					resource.TestCheckResourceAttr(resourceName, names.AttrVersion, acctest.Ct1),
				),
			},
			//Executes version resource again and validates first version exists
			{
				Config: testAccGuardrailVersion_skipDestroy(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGuardrailVersionExists(ctx, resourceName, acctest.Ct1, &guardrailversion),
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

			id := rs.Primary.Attributes["guardrail_identifier"]
			version := rs.Primary.Attributes[names.AttrVersion]
			_, err := tfbedrock.FindGuardrailByID(ctx, conn, id, version)
			if errs.IsA[*types.ResourceNotFoundException](err) {
				return nil
			}
			if err != nil {
				return create.Error(names.Bedrock, create.ErrActionCheckingDestroyed, tfbedrock.ResNameGuardrailVersion, rs.Primary.ID, err)
			}

			return create.Error(names.Bedrock, create.ErrActionCheckingDestroyed, tfbedrock.ResNameGuardrailVersion, rs.Primary.ID, errors.New("not destroyed"))
		}

		return nil
	}
}

func testAccCheckGuardrailVersionExists(ctx context.Context, name string, version string, guardrail *bedrock.GetGuardrailOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return create.Error(names.Bedrock, create.ErrActionCheckingExistence, tfbedrock.ResNameGuardrail, name, errors.New("not found"))
		}

		id := rs.Primary.Attributes["guardrail_identifier"]
		if id == "" {
			return create.Error(names.Bedrock, create.ErrActionCheckingExistence, tfbedrock.ResNameGuardrail, name, errors.New("guardrail_id not set"))
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).BedrockClient(ctx)

		out, err := tfbedrock.FindGuardrailByID(ctx, conn, id, version)
		if err != nil {
			return create.Error(names.Bedrock, create.ErrActionCheckingExistence, tfbedrock.ResNameGuardrail, rs.Primary.ID, err)
		}

		*guardrail = *out

		return nil
	}
}

func testAccGuardrailVersionImportStateIDFunc(resourceName string) resource.ImportStateIdFunc {
	return func(s *terraform.State) (string, error) {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return "", fmt.Errorf("Not found: %s", resourceName)
		}
		return fmt.Sprintf("%s,%s", rs.Primary.Attributes["guardrail_identifier"], rs.Primary.Attributes[names.AttrVersion]), nil
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
  description             = %[1]q
  guardrail_identifier    = aws_bedrock_guardrail.test.guardrail_arn
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
  description             = %[1]q
  guardrail_identifier    = aws_bedrock_guardrail.test.guardrail_arn
  skip_destroy 			  = true
}
`, rName)
}
