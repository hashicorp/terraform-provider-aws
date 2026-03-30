// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package bedrockagent_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/service/bedrockagent"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	tfbedrockagent "github.com/hashicorp/terraform-provider-aws/internal/service/bedrockagent"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccBedrockAgentPrompt_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var prompt bedrockagent.GetPromptOutput
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_bedrockagent_prompt.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.BedrockEndpointID)
			testAccPreCheckPrompt(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.BedrockAgentServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckPromptDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccPromptConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckPromptExists(ctx, t, resourceName, &prompt),
					acctest.MatchResourceAttrRegionalARN(ctx, resourceName, names.AttrARN, "bedrock", regexache.MustCompile(`prompt/.+$`)),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrID),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrCreatedAt),
					resource.TestCheckResourceAttrSet(resourceName, "updated_at"),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrVersion),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, acctest.CtBasic),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccBedrockAgentPrompt_withEncryption(t *testing.T) {
	ctx := acctest.Context(t)
	var prompt bedrockagent.GetPromptOutput
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_bedrockagent_prompt.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.BedrockEndpointID)
			testAccPreCheckPrompt(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.BedrockAgentServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckPromptDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccPromptConfig_withEncryption(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckPromptExists(ctx, t, resourceName, &prompt),
					acctest.MatchResourceAttrRegionalARN(ctx, resourceName, names.AttrARN, "bedrock", regexache.MustCompile(`prompt/.+$`)),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrID),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrCreatedAt),
					resource.TestCheckResourceAttrSet(resourceName, "updated_at"),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrVersion),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttrPair(resourceName, "customer_encryption_key_arn", "aws_kms_key.test", names.AttrARN)),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccBedrockAgentPrompt_variants(t *testing.T) {
	ctx := acctest.Context(t)
	var prompt bedrockagent.GetPromptOutput
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_bedrockagent_prompt.test"
	foundationModel := "amazon.titan-text-express-v1"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.BedrockEndpointID)
			testAccPreCheckPrompt(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.BedrockAgentServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckPromptDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccPromptConfig_variants(rName, foundationModel),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckPromptExists(ctx, t, resourceName, &prompt),
					acctest.MatchResourceAttrRegionalARN(ctx, resourceName, names.AttrARN, "bedrock", regexache.MustCompile(`prompt/.+$`)),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrID),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrCreatedAt),
					resource.TestCheckResourceAttrSet(resourceName, "updated_at"),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrVersion),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, "default_variant", "text-variant"),
					resource.TestCheckResourceAttr(resourceName, "variant.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "variant.0.name", "text-variant"),
					resource.TestCheckResourceAttr(resourceName, "variant.0.model_id", foundationModel),
					resource.TestCheckResourceAttr(resourceName, "variant.0.template_type", "TEXT"),
					resource.TestCheckResourceAttr(resourceName, "variant.0.template_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "variant.0.template_configuration.0.text.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "variant.0.template_configuration.0.text.0.text", "{{prompt}}"),
					resource.TestCheckResourceAttr(resourceName, "variant.0.template_configuration.0.text.0.input_variable.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "variant.0.template_configuration.0.text.0.input_variable.0.name", "prompt"),
					resource.TestCheckResourceAttr(resourceName, "variant.1.name", "chat-variant"),
					resource.TestCheckResourceAttr(resourceName, "variant.1.model_id", foundationModel),
					resource.TestCheckResourceAttr(resourceName, "variant.1.template_type", "CHAT"),
					resource.TestCheckResourceAttr(resourceName, "variant.1.template_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "variant.1.template_configuration.0.chat.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "variant.1.template_configuration.0.chat.0.system.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "variant.1.template_configuration.0.chat.0.system.0.text", "test"),
					resource.TestCheckResourceAttr(resourceName, "variant.1.template_configuration.0.chat.0.message.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "variant.1.template_configuration.0.chat.0.message.0.role", "user"),
					resource.TestCheckResourceAttr(resourceName, "variant.1.template_configuration.0.chat.0.message.0.content.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "variant.1.template_configuration.0.chat.0.message.0.content.0.text", "test"),
					resource.TestCheckResourceAttr(resourceName, "variant.1.template_configuration.0.chat.0.message.1.role", "assistant"),
					resource.TestCheckResourceAttr(resourceName, "variant.1.template_configuration.0.chat.0.message.1.content.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "variant.1.template_configuration.0.chat.0.message.1.content.0.text", "test"),
					resource.TestCheckResourceAttr(resourceName, "variant.1.template_configuration.0.chat.0.tool_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "variant.1.template_configuration.0.chat.0.tool_configuration.0.tool_choice.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "variant.1.template_configuration.0.chat.0.tool_configuration.0.tool_choice.0.auto.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "variant.1.template_configuration.0.chat.0.tool_configuration.0.tool.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "variant.1.template_configuration.0.chat.0.tool_configuration.0.tool.0.tool_spec.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "variant.1.template_configuration.0.chat.0.tool_configuration.0.tool.0.tool_spec.0.name", "test"),
					resource.TestCheckResourceAttr(resourceName, "variant.1.template_configuration.0.chat.0.tool_configuration.0.tool.0.tool_spec.0.description", "test"),
					resource.TestCheckResourceAttr(resourceName, "variant.1.template_configuration.0.chat.0.tool_configuration.0.tool.0.tool_spec.0.input_schema.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "variant.1.template_configuration.0.chat.0.tool_configuration.0.tool.0.tool_spec.0.input_schema.0.json", "{\"Key1\":\"Value1\"}"),
				),
			},
		},
	})
}

func TestAccBedrockAgentPrompt_extraFields(t *testing.T) {
	ctx := acctest.Context(t)
	var prompt bedrockagent.GetPromptOutput
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_bedrockagent_prompt.test"
	foundationModel := "amazon.titan-text-express-v1"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.BedrockEndpointID)
			testAccPreCheckPrompt(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.BedrockAgentServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckPromptDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccPromptConfig_extraFields(rName, foundationModel),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckPromptExists(ctx, t, resourceName, &prompt),
					acctest.MatchResourceAttrRegionalARN(ctx, resourceName, names.AttrARN, "bedrock", regexache.MustCompile(`prompt/.+$`)),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrID),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrCreatedAt),
					resource.TestCheckResourceAttrSet(resourceName, "updated_at"),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrVersion),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, "default_variant", "test"),
					resource.TestCheckResourceAttr(resourceName, "variant.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "variant.0.name", "test"),
					resource.TestCheckResourceAttr(resourceName, "variant.0.additional_model_request_fields", "{\"Key1\":\"Value1\"}"),
					resource.TestCheckResourceAttr(resourceName, "variant.0.inference_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "variant.0.inference_configuration.0.text.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "variant.0.inference_configuration.0.text.0.max_tokens", "2048"),
					resource.TestCheckResourceAttr(resourceName, "variant.0.inference_configuration.0.text.0.stop_sequences.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "variant.0.inference_configuration.0.text.0.stop_sequences.0", "User:"),
					resource.TestCheckResourceAttr(resourceName, "variant.0.inference_configuration.0.text.0.top_p", "0.8999999761581421"),
					resource.TestCheckResourceAttr(resourceName, "variant.0.inference_configuration.0.text.0.temperature", "0"),
					resource.TestCheckResourceAttr(resourceName, "variant.0.metadata.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "variant.0.metadata.0.key", "Key1"),
					resource.TestCheckResourceAttr(resourceName, "variant.0.metadata.0.value", "Value1"),
					resource.TestCheckResourceAttr(resourceName, "variant.0.metadata.1.key", "Key2"),
					resource.TestCheckResourceAttr(resourceName, "variant.0.metadata.1.value", "Value2"),
					resource.TestCheckResourceAttr(resourceName, "variant.0.gen_ai_resource.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "variant.0.gen_ai_resource.0.agent.#", "1"),
					acctest.MatchResourceAttrRegionalARN(ctx, resourceName, "variant.0.gen_ai_resource.0.agent.0.agent_identifier", "bedrock", regexache.MustCompile(`agent-alias/.+$`)),
					resource.TestCheckResourceAttr(resourceName, "variant.0.template_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "variant.0.template_configuration.0.text.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "variant.0.template_configuration.0.text.0.text", "test"),
				),
			},
		},
	})
}

func TestAccBedrockAgentPrompt_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	var prompt bedrockagent.GetPromptOutput
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_bedrockagent_prompt.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.BedrockEndpointID)
			testAccPreCheckPrompt(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.BedrockAgentServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckPromptDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccPromptConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckPromptExists(ctx, t, resourceName, &prompt),
					acctest.CheckFrameworkResourceDisappears(ctx, t, tfbedrockagent.ResourcePrompt, resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccBedrockAgentPrompt_tags(t *testing.T) {
	ctx := acctest.Context(t)
	var prompt bedrockagent.GetPromptOutput
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_bedrockagent_prompt.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.BedrockEndpointID)
			testAccPreCheckPrompt(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.BedrockAgentServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckPromptDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccPromptConfig_tags1(rName, acctest.CtKey1, acctest.CtValue1),
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
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckPromptExists(ctx, t, resourceName, &prompt),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccPromptConfig_tags2(rName, acctest.CtKey1, acctest.CtValue1Updated, acctest.CtKey2, acctest.CtValue2),
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
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckPromptExists(ctx, t, resourceName, &prompt),
				),
			},
			{
				Config: testAccPromptConfig_tags1(rName, acctest.CtKey2, acctest.CtValue2),
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
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckPromptExists(ctx, t, resourceName, &prompt),
				),
			},
		},
	})
}

func testAccCheckPromptDestroy(ctx context.Context, t *testing.T) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.ProviderMeta(ctx, t).BedrockAgentClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_bedrockagent_prompt" {
				continue
			}

			_, err := tfbedrockagent.FindPromptByID(ctx, conn, rs.Primary.ID)

			if retry.NotFound(err) {
				return nil
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("Bedrock Agent Prompt %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccCheckPromptExists(ctx context.Context, t *testing.T, n string, v *bedrockagent.GetPromptOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.ProviderMeta(ctx, t).BedrockAgentClient(ctx)

		output, err := tfbedrockagent.FindPromptByID(ctx, conn, rs.Primary.ID)

		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

func testAccPreCheckPrompt(ctx context.Context, t *testing.T) {
	conn := acctest.ProviderMeta(ctx, t).BedrockAgentClient(ctx)

	input := &bedrockagent.ListPromptsInput{}

	_, err := conn.ListPrompts(ctx, input)

	if acctest.PreCheckSkipError(err) {
		t.Skipf("skipping acceptance testing: %s", err)
	}
	if err != nil {
		t.Fatalf("unexpected PreCheck error: %s", err)
	}
}

func testAccPromptConfig_basic(rName string) string {
	return fmt.Sprintf(`
resource "aws_bedrockagent_prompt" "test" {
  name        = %[1]q
  description = "basic"
}
`, rName)
}

func testAccPromptConfig_withEncryption(rName string) string {
	return fmt.Sprintf(`
resource "aws_kms_key" "test" {
  description             = %[1]q
  deletion_window_in_days = 7
}

resource "aws_bedrockagent_prompt" "test" {
  name                        = %[1]q
  customer_encryption_key_arn = aws_kms_key.test.arn
}
`, rName)
}

func testAccPromptConfig_tags1(rName, tag1Key, tag1Value string) string {
	return fmt.Sprintf(`
resource "aws_bedrockagent_prompt" "test" {
  name = %[1]q

  tags = {
    %[2]q = %[3]q
  }
}
`, rName, tag1Key, tag1Value)
}

func testAccPromptConfig_tags2(rName, tag1Key, tag1Value, tag2Key, tag2Value string) string {
	return fmt.Sprintf(`
resource "aws_bedrockagent_prompt" "test" {
  name = %[1]q

  tags = {
    %[2]q = %[3]q
    %[4]q = %[5]q
  }
}
`, rName, tag1Key, tag1Value, tag2Key, tag2Value)
}

func testAccPromptConfig_variants(rName, model string) string {
	return fmt.Sprintf(`
resource "aws_bedrockagent_prompt" "test" {
  name            = %[1]q
  default_variant = "text-variant"

  variant {
    name          = "text-variant"
    model_id      = %[2]q
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

  variant {
    name          = "chat-variant"
    model_id      = %[2]q
    template_type = "CHAT"

    template_configuration {
      chat {
        system {
          text = "test"
        }

        message {
          role = "user"

          content {
            text = "test"
          }
        }

        message {
          role = "assistant"

          content {
            text = "test"
          }
        }

        tool_configuration {
          tool_choice {
            auto {}
          }

          tool {
            tool_spec {
              name        = "test"
              description = "test"

              input_schema {
                json = jsonencode({ "Key1" = "Value1" })
              }
            }
          }
        }
      }
    }
  }
}
`, rName, model)
}

func testAccPromptConfig_extraFields(rName, model string) string {
	return fmt.Sprintf(`
resource "aws_iam_role" "test_agent" {
  assume_role_policy = data.aws_iam_policy_document.test_agent_trust.json
  name_prefix        = "AmazonBedrockExecutionRoleForAgents_tf"
}

data "aws_iam_policy_document" "test_agent_trust" {
  statement {
    actions = ["sts:AssumeRole"]
    principals {
      identifiers = ["bedrock.amazonaws.com"]
      type        = "Service"
    }
  }
}

resource "aws_bedrockagent_agent" "test" {
  agent_name                  = %[1]q
  agent_resource_role_arn     = aws_iam_role.test_agent.arn
  idle_session_ttl_in_seconds = 500
  instruction                 = file("${path.module}/test-fixtures/instruction.txt")
  foundation_model            = %[2]q
}

resource "aws_bedrockagent_agent_alias" "test" {
  agent_alias_name = %[1]q
  agent_id         = aws_bedrockagent_agent.test.agent_id
}

resource "aws_bedrockagent_prompt" "test" {
  name            = %[1]q
  default_variant = "test"

  variant {
    name                            = "test"
    additional_model_request_fields = jsonencode({ "Key1" = "Value1" })

    inference_configuration {
      text {
        max_tokens     = 2048
        stop_sequences = ["User:"]
        top_p          = 0.8999999761581421
        temperature    = 0
      }
    }

    metadata {
      key   = "Key1"
      value = "Value1"
    }

    metadata {
      key   = "Key2"
      value = "Value2"
    }

    gen_ai_resource {
      agent {
        agent_identifier = aws_bedrockagent_agent_alias.test.agent_alias_arn
      }
    }

    template_type = "TEXT"
    template_configuration {
      text {
        text = "test"
      }
    }
  }
}
`, rName, model)
}
