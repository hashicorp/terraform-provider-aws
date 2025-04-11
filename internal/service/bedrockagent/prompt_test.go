// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package bedrockagent_test

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/service/bedrockagent"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"

	tfbedrockagent "github.com/hashicorp/terraform-provider-aws/internal/service/bedrockagent"
)

func TestAccBedrockAgentPrompt_basic(t *testing.T) {
	ctx := acctest.Context(t)

	var prompt bedrockagent.GetPromptOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_bedrockagent_prompt.test"
	foundationModel := "amazon.titan-text-express-v1"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.BedrockEndpointID)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.BedrockAgentServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckPromptDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccPromptConfig_basic(rName, foundationModel),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckPromptExists(ctx, resourceName, &prompt),
					acctest.MatchResourceAttrRegionalARN(ctx, resourceName, names.AttrARN, "bedrock", regexache.MustCompile(`prompt/.+$`)),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrID),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrCreatedAt),
					resource.TestCheckResourceAttrSet(resourceName, "updated_at"),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrVersion),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, "basic"),
					resource.TestCheckResourceAttr(resourceName, "default_variant", "test-variant"),
					resource.TestCheckResourceAttr(resourceName, "variant.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "variant.0.name", "test-variant"),
					resource.TestCheckResourceAttr(resourceName, "variant.0.model_id", foundationModel),
					resource.TestCheckResourceAttr(resourceName, "variant.0.metadata.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "variant.0.metadata.0.key", "Key1"),
					resource.TestCheckResourceAttr(resourceName, "variant.0.metadata.0.value", "Value1"),
					resource.TestCheckResourceAttr(resourceName, "variant.0.metadata.1.key", "Key2"),
					resource.TestCheckResourceAttr(resourceName, "variant.0.metadata.1.value", "Value2"),
					resource.TestCheckResourceAttr(resourceName, "variant.0.additional_model_request_fields", "{\"Key1\":\"Value1\"}"),
					resource.TestCheckResourceAttr(resourceName, "variant.0.template_type", "TEXT"),
					resource.TestCheckResourceAttr(resourceName, "variant.0.template_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "variant.0.template_configuration.0.text.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "variant.0.template_configuration.0.text.0.text", "{{prompt}}"),
					resource.TestCheckResourceAttr(resourceName, "variant.0.template_configuration.0.text.0.input_variable.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "variant.0.template_configuration.0.text.0.input_variable.0.name", "prompt"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"apply_immediately", "user"},
			},
		},
	})
}

func TestAccBedrockAgentPrompt_disappears(t *testing.T) {
	ctx := acctest.Context(t)

	var prompt bedrockagent.GetPromptOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_bedrockagent_prompt.test"
	foundationModel := "amazon.titan-text-express-v1"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.BedrockEndpointID)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.BedrockAgentServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckPromptDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccPromptConfig_basic(rName, foundationModel),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckPromptExists(ctx, resourceName, &prompt),
					acctest.CheckFrameworkResourceDisappears(ctx, acctest.Provider, tfbedrockagent.ResourcePrompt, resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckPromptDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).BedrockAgentClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_bedrockagent_prompt" {
				continue
			}

			_, err := tfbedrockagent.FindPromptByID(ctx, conn, rs.Primary.ID)
			if tfresource.NotFound(err) {
				return nil
			}
			if err != nil {
				return create.Error(names.BedrockAgent, create.ErrActionCheckingDestroyed, tfbedrockagent.ResNamePrompt, rs.Primary.ID, err)
			}

			return create.Error(names.BedrockAgent, create.ErrActionCheckingDestroyed, tfbedrockagent.ResNamePrompt, rs.Primary.ID, errors.New("not destroyed"))
		}

		return nil
	}
}

func testAccCheckPromptExists(ctx context.Context, name string, prompt *bedrockagent.GetPromptOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return create.Error(names.BedrockAgent, create.ErrActionCheckingExistence, tfbedrockagent.ResNamePrompt, name, errors.New("not found"))
		}

		if rs.Primary.ID == "" {
			return create.Error(names.BedrockAgent, create.ErrActionCheckingExistence, tfbedrockagent.ResNamePrompt, name, errors.New("not set"))
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).BedrockAgentClient(ctx)

		resp, err := tfbedrockagent.FindPromptByID(ctx, conn, rs.Primary.ID)
		if err != nil {
			return create.Error(names.BedrockAgent, create.ErrActionCheckingExistence, tfbedrockagent.ResNamePrompt, rs.Primary.ID, err)
		}

		*prompt = *resp

		return nil
	}
}

func testAccPreCheck(ctx context.Context, t *testing.T) {
	conn := acctest.Provider.Meta().(*conns.AWSClient).BedrockAgentClient(ctx)

	input := &bedrockagent.ListPromptsInput{}

	_, err := conn.ListPrompts(ctx, input)

	if acctest.PreCheckSkipError(err) {
		t.Skipf("skipping acceptance testing: %s", err)
	}
	if err != nil {
		t.Fatalf("unexpected PreCheck error: %s", err)
	}
}

func testAccPromptConfig_basic(rName, model string) string {
	return fmt.Sprintf(`
resource "aws_bedrockagent_prompt" "test" {
  name            = %[1]q
  description     = "basic"
  default_variant = "test-variant"

  variant {
    name                            = "test-variant"
    model_id                        = %[2]q
    additional_model_request_fields = jsonencode({ "Key1" = "Value1" })

    metadata {
      key   = "Key1"
      value = "Value1"
    }

    metadata {
      key   = "Key2"
      value = "Value2"
    }

    template_type = "TEXT"
    template_configuration {
      text {
        text = "{{prompt}}"

        input_variable {
          name = "prompt"
        }
      }
    }
  }
}
`, rName, model)
}
