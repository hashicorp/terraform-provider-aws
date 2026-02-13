// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package bedrockagent_test

import (
	"context"
	"errors"
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
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	tfbedrockagent "github.com/hashicorp/terraform-provider-aws/internal/service/bedrockagent"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccBedrockAgentFlow_basic(t *testing.T) {
	ctx := acctest.Context(t)

	var flow bedrockagent.GetFlowOutput
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_bedrockagent_flow.test"
	foundationModel := "amazon.titan-text-express-v1"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.BedrockEndpointID)
			testAccPreCheckFlow(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.BedrockAgentServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckFlowDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccFlowConfig_basic(rName, foundationModel),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckFlowExists(ctx, t, resourceName, &flow),
					acctest.MatchResourceAttrRegionalARN(ctx, resourceName, names.AttrARN, "bedrock", regexache.MustCompile(`flow/.+$`)),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrID),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrCreatedAt),
					resource.TestCheckResourceAttrSet(resourceName, "updated_at"),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrVersion),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrStatus),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, acctest.CtBasic),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrExecutionRoleARN, "aws_iam_role.test", names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, "definition.#", "0"),
					resource.TestCheckNoResourceAttr(resourceName, "customer_encryption_key_arn"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccFlowConfig_basicUpdate(rName, foundationModel),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckFlowExists(ctx, t, resourceName, &flow),
					acctest.MatchResourceAttrRegionalARN(ctx, resourceName, names.AttrARN, "bedrock", regexache.MustCompile(`flow/.+$`)),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrID),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrCreatedAt),
					resource.TestCheckResourceAttrSet(resourceName, "updated_at"),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrVersion),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrStatus),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, "update"),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrExecutionRoleARN, "aws_iam_role.test", names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, "definition.#", "0"),
					resource.TestCheckNoResourceAttr(resourceName, "customer_encryption_key_arn"),
				),
			},
		},
	})
}

func TestAccBedrockAgentFlow_disappears(t *testing.T) {
	ctx := acctest.Context(t)

	var flow bedrockagent.GetFlowOutput
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_bedrockagent_flow.test"
	foundationModel := "amazon.titan-text-express-v1"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.BedrockEndpointID)
			testAccPreCheckFlow(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.BedrockAgentServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckFlowDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccFlowConfig_basic(rName, foundationModel),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckFlowExists(ctx, t, resourceName, &flow),
					acctest.CheckFrameworkResourceDisappears(ctx, t, tfbedrockagent.ResourceFlow, resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccBedrockAgentFlow_withEncryptionKey(t *testing.T) {
	ctx := acctest.Context(t)

	var flow bedrockagent.GetFlowOutput
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_bedrockagent_flow.test"
	foundationModel := "amazon.titan-text-express-v1"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.BedrockEndpointID)
			testAccPreCheckFlow(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.BedrockAgentServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckFlowDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccFlowConfig_withEncryptionKey(rName, foundationModel),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckFlowExists(ctx, t, resourceName, &flow),
					acctest.MatchResourceAttrRegionalARN(ctx, resourceName, names.AttrARN, "bedrock", regexache.MustCompile(`flow/.+$`)),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrID),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrCreatedAt),
					resource.TestCheckResourceAttrSet(resourceName, "updated_at"),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrVersion),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrStatus),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrExecutionRoleARN, "aws_iam_role.test", names.AttrARN),
					resource.TestCheckResourceAttrPair(resourceName, "customer_encryption_key_arn", "aws_kms_key.test", names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, "definition.#", "0"),
					resource.TestCheckNoResourceAttr(resourceName, names.AttrDescription),
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

func TestAccBedrockAgentFlow_tags(t *testing.T) {
	ctx := acctest.Context(t)

	var flow bedrockagent.GetFlowOutput
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_bedrockagent_flow.test"
	foundationModel := "amazon.titan-text-express-v1"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.BedrockEndpointID)
			testAccPreCheckFlow(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.BedrockAgentServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckFlowDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccFlowConfig_tags1(rName, foundationModel, acctest.CtKey1, acctest.CtValue1),
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
					testAccCheckFlowExists(ctx, t, resourceName, &flow),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccFlowConfig_tags2(rName, foundationModel, acctest.CtKey1, acctest.CtValue1Updated, acctest.CtKey2, acctest.CtValue2),
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
					testAccCheckFlowExists(ctx, t, resourceName, &flow),
				),
			},
			{
				Config: testAccFlowConfig_tags1(rName, foundationModel, acctest.CtKey2, acctest.CtValue2),
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
					testAccCheckFlowExists(ctx, t, resourceName, &flow),
				),
			},
		},
	})
}

func TestAccBedrockAgentFlow_withDefinition(t *testing.T) {
	ctx := acctest.Context(t)

	var flow bedrockagent.GetFlowOutput
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_bedrockagent_flow.test"
	foundationModel := "amazon.titan-text-express-v1"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.BedrockEndpointID)
			testAccPreCheckFlow(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.BedrockAgentServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckFlowDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccFlowConfig_withDefinition(rName, foundationModel),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckFlowExists(ctx, t, resourceName, &flow),
					acctest.MatchResourceAttrRegionalARN(ctx, resourceName, names.AttrARN, "bedrock", regexache.MustCompile(`flow/.+$`)),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrID),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrCreatedAt),
					resource.TestCheckResourceAttrSet(resourceName, "updated_at"),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrVersion),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrStatus),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrExecutionRoleARN, "aws_iam_role.test", names.AttrARN),
					resource.TestCheckNoResourceAttr(resourceName, "customer_encryption_key_arn"),
					resource.TestCheckNoResourceAttr(resourceName, names.AttrDescription),

					resource.TestCheckResourceAttr(resourceName, "definition.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "definition.0.connection.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "definition.0.node.#", "3"),

					resource.TestCheckResourceAttr(resourceName, "definition.0.connection.0.name", "FlowInputNodeFlowInputNode0ToPrompt_1PromptsNode0"),
					resource.TestCheckResourceAttr(resourceName, "definition.0.connection.0.source", "FlowInputNode"),
					resource.TestCheckResourceAttr(resourceName, "definition.0.connection.0.target", "Prompt_1"),
					resource.TestCheckResourceAttr(resourceName, "definition.0.connection.0.type", "Data"),
					resource.TestCheckResourceAttr(resourceName, "definition.0.connection.0.configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "definition.0.connection.0.configuration.0.data.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "definition.0.connection.0.configuration.0.data.0.source_output", "document"),
					resource.TestCheckResourceAttr(resourceName, "definition.0.connection.0.configuration.0.data.0.target_input", "topic"),

					resource.TestCheckResourceAttr(resourceName, "definition.0.connection.1.name", "Prompt_1PromptsNode0ToFlowOutputNodeFlowOutputNode0"),
					resource.TestCheckResourceAttr(resourceName, "definition.0.connection.1.source", "Prompt_1"),
					resource.TestCheckResourceAttr(resourceName, "definition.0.connection.1.target", "FlowOutputNode"),
					resource.TestCheckResourceAttr(resourceName, "definition.0.connection.1.type", "Data"),
					resource.TestCheckResourceAttr(resourceName, "definition.0.connection.1.configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "definition.0.connection.1.configuration.0.data.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "definition.0.connection.1.configuration.0.data.0.source_output", "modelCompletion"),
					resource.TestCheckResourceAttr(resourceName, "definition.0.connection.1.configuration.0.data.0.target_input", "document"),

					resource.TestCheckResourceAttr(resourceName, "definition.0.node.0.name", "FlowInputNode"),
					resource.TestCheckResourceAttr(resourceName, "definition.0.node.0.type", "Input"),
					resource.TestCheckResourceAttr(resourceName, "definition.0.node.0.configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "definition.0.node.0.configuration.0.input.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "definition.0.node.0.output.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "definition.0.node.0.output.0.name", "document"),
					resource.TestCheckResourceAttr(resourceName, "definition.0.node.0.output.0.type", "String"),

					resource.TestCheckResourceAttr(resourceName, "definition.0.node.1.name", "Prompt_1"),
					resource.TestCheckResourceAttr(resourceName, "definition.0.node.1.type", "Prompt"),
					resource.TestCheckResourceAttr(resourceName, "definition.0.node.1.configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "definition.0.node.1.configuration.0.prompt.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "definition.0.node.1.configuration.0.prompt.0.source_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "definition.0.node.1.configuration.0.prompt.0.source_configuration.0.inline.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "definition.0.node.1.configuration.0.prompt.0.source_configuration.0.inline.0.model_id", foundationModel),
					resource.TestCheckResourceAttr(resourceName, "definition.0.node.1.configuration.0.prompt.0.source_configuration.0.inline.0.template_type", "TEXT"),
					resource.TestCheckResourceAttr(resourceName, "definition.0.node.1.configuration.0.prompt.0.source_configuration.0.inline.0.inference_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "definition.0.node.1.configuration.0.prompt.0.source_configuration.0.inline.0.inference_configuration.0.text.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "definition.0.node.1.configuration.0.prompt.0.source_configuration.0.inline.0.inference_configuration.0.text.0.max_tokens", "2048"),
					resource.TestCheckResourceAttr(resourceName, "definition.0.node.1.configuration.0.prompt.0.source_configuration.0.inline.0.inference_configuration.0.text.0.stop_sequences.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "definition.0.node.1.configuration.0.prompt.0.source_configuration.0.inline.0.inference_configuration.0.text.0.stop_sequences.0", "User:"),
					resource.TestCheckResourceAttr(resourceName, "definition.0.node.1.configuration.0.prompt.0.source_configuration.0.inline.0.inference_configuration.0.text.0.temperature", "0"),
					resource.TestCheckResourceAttr(resourceName, "definition.0.node.1.configuration.0.prompt.0.source_configuration.0.inline.0.inference_configuration.0.text.0.top_p", "0.8999999761581421"),
					resource.TestCheckResourceAttr(resourceName, "definition.0.node.1.configuration.0.prompt.0.source_configuration.0.inline.0.template_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "definition.0.node.1.configuration.0.prompt.0.source_configuration.0.inline.0.template_configuration.0.text.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "definition.0.node.1.configuration.0.prompt.0.source_configuration.0.inline.0.template_configuration.0.text.0.text", "Write a paragraph about {{topic}}."),
					resource.TestCheckResourceAttr(resourceName, "definition.0.node.1.configuration.0.prompt.0.source_configuration.0.inline.0.template_configuration.0.text.0.input_variable.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "definition.0.node.1.configuration.0.prompt.0.source_configuration.0.inline.0.template_configuration.0.text.0.input_variable.0.name", "topic"),
					resource.TestCheckResourceAttr(resourceName, "definition.0.node.1.input.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "definition.0.node.1.input.0.expression", "$.data"),
					resource.TestCheckResourceAttr(resourceName, "definition.0.node.1.input.0.name", "topic"),
					resource.TestCheckResourceAttr(resourceName, "definition.0.node.1.input.0.type", "String"),
					resource.TestCheckResourceAttr(resourceName, "definition.0.node.1.output.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "definition.0.node.1.output.0.name", "modelCompletion"),
					resource.TestCheckResourceAttr(resourceName, "definition.0.node.1.output.0.type", "String"),

					resource.TestCheckResourceAttr(resourceName, "definition.0.node.2.name", "FlowOutputNode"),
					resource.TestCheckResourceAttr(resourceName, "definition.0.node.2.type", "Output"),
					resource.TestCheckResourceAttr(resourceName, "definition.0.node.2.configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "definition.0.node.2.configuration.0.output.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "definition.0.node.2.input.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "definition.0.node.2.input.0.expression", "$.data"),
					resource.TestCheckResourceAttr(resourceName, "definition.0.node.2.input.0.name", "document"),
					resource.TestCheckResourceAttr(resourceName, "definition.0.node.2.input.0.type", "String"),
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

func TestAccBedrockAgentFlow_withPromptResource(t *testing.T) {
	ctx := acctest.Context(t)

	var flow bedrockagent.GetFlowOutput
	flowName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	promptName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_bedrockagent_flow.test"
	foundationModel := "amazon.titan-text-express-v1"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.BedrockEndpointID)
			testAccPreCheckFlow(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.BedrockAgentServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckFlowDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccFlowConfig_withPromptResource(flowName, promptName, foundationModel),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckFlowExists(ctx, t, resourceName, &flow),
					acctest.MatchResourceAttrRegionalARN(ctx, resourceName, names.AttrARN, "bedrock", regexache.MustCompile(`flow/.+$`)),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrID),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrCreatedAt),
					resource.TestCheckResourceAttrSet(resourceName, "updated_at"),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrVersion),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrStatus),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, flowName),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrExecutionRoleARN, "aws_iam_role.test", names.AttrARN),
					resource.TestCheckNoResourceAttr(resourceName, "customer_encryption_key_arn"),
					resource.TestCheckNoResourceAttr(resourceName, names.AttrDescription),

					resource.TestCheckResourceAttr(resourceName, "definition.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "definition.0.connection.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "definition.0.node.#", "1"),

					resource.TestCheckResourceAttr(resourceName, "definition.0.node.0.name", "Prompt_1"),
					resource.TestCheckResourceAttr(resourceName, "definition.0.node.0.type", "Prompt"),
					resource.TestCheckResourceAttr(resourceName, "definition.0.node.0.configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "definition.0.node.0.configuration.0.prompt.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "definition.0.node.0.configuration.0.prompt.0.source_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "definition.0.node.0.configuration.0.prompt.0.source_configuration.0.resource.#", "1"),
					resource.TestCheckResourceAttrPair(resourceName, "definition.0.node.0.configuration.0.prompt.0.source_configuration.0.resource.0.prompt_arn", "aws_bedrockagent_prompt.test", names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, "definition.0.node.0.input.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "definition.0.node.0.input.0.expression", "$.data"),
					resource.TestCheckResourceAttr(resourceName, "definition.0.node.0.input.0.name", "topic"),
					resource.TestCheckResourceAttr(resourceName, "definition.0.node.0.input.0.type", "String"),
					resource.TestCheckResourceAttr(resourceName, "definition.0.node.0.output.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "definition.0.node.0.output.0.name", "modelCompletion"),
					resource.TestCheckResourceAttr(resourceName, "definition.0.node.0.output.0.type", "String"),
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

func testAccCheckFlowDestroy(ctx context.Context, t *testing.T) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.ProviderMeta(ctx, t).BedrockAgentClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_bedrockagent_flow" {
				continue
			}

			_, err := tfbedrockagent.FindFlowByID(ctx, conn, rs.Primary.ID)
			if retry.NotFound(err) {
				return nil
			}
			if err != nil {
				return create.Error(names.BedrockAgent, create.ErrActionCheckingDestroyed, tfbedrockagent.ResNameFlow, rs.Primary.ID, err)
			}

			return create.Error(names.BedrockAgent, create.ErrActionCheckingDestroyed, tfbedrockagent.ResNameFlow, rs.Primary.ID, errors.New("not destroyed"))
		}

		return nil
	}
}

func testAccCheckFlowExists(ctx context.Context, t *testing.T, name string, flow *bedrockagent.GetFlowOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return create.Error(names.BedrockAgent, create.ErrActionCheckingExistence, tfbedrockagent.ResNameFlow, name, errors.New("not found"))
		}

		if rs.Primary.ID == "" {
			return create.Error(names.BedrockAgent, create.ErrActionCheckingExistence, tfbedrockagent.ResNameFlow, name, errors.New("not set"))
		}

		conn := acctest.ProviderMeta(ctx, t).BedrockAgentClient(ctx)

		resp, err := tfbedrockagent.FindFlowByID(ctx, conn, rs.Primary.ID)
		if err != nil {
			return create.Error(names.BedrockAgent, create.ErrActionCheckingExistence, tfbedrockagent.ResNameFlow, rs.Primary.ID, err)
		}

		*flow = *resp

		return nil
	}
}

func testAccPreCheckFlow(ctx context.Context, t *testing.T) {
	conn := acctest.ProviderMeta(ctx, t).BedrockAgentClient(ctx)

	input := &bedrockagent.ListFlowsInput{}

	_, err := conn.ListFlows(ctx, input)

	if acctest.PreCheckSkipError(err) {
		t.Skipf("skipping acceptance testing: %s", err)
	}
	if err != nil {
		t.Fatalf("unexpected PreCheck error: %s", err)
	}
}

func testAccFlowConfig_base(model string) string {
	return fmt.Sprintf(`
data "aws_caller_identity" "current" {}
data "aws_region" "current" {}
data "aws_partition" "current" {}

resource "aws_iam_role" "test" {
  name_prefix        = "AmazonBedrockExecutionRoleForFlows_tf"
  assume_role_policy = data.aws_iam_policy_document.test_trust.json
}

data "aws_iam_policy_document" "test_trust" {
  statement {
    actions = ["sts:AssumeRole"]
    principals {
      identifiers = ["bedrock.amazonaws.com"]
      type        = "Service"
    }
  }
}

resource "aws_iam_role_policy" "test_permissions" {
  role   = aws_iam_role.test.id
  policy = data.aws_iam_policy_document.test_permissions.json
}

data "aws_iam_policy_document" "test_permissions" {
  statement {
    actions   = ["bedrock:GetFlow"]
    effect    = "Allow"
    resources = ["*"]
  }
  statement {
    actions = ["bedrock:InvokeModel"]
    effect  = "Allow"
    resources = [
      "arn:${data.aws_partition.current.partition}:bedrock:${data.aws_region.current.name}::foundation-model/%[1]s",
    ]
  }
}`, model)
}

func testAccFlowConfig_basic(rName, model string) string {
	return acctest.ConfigCompose(testAccFlowConfig_base(model), fmt.Sprintf(`
resource "aws_bedrockagent_flow" "test" {
  name               = %[1]q
  execution_role_arn = aws_iam_role.test.arn
  description        = "basic"
}
`, rName))
}

func testAccFlowConfig_basicUpdate(rName, model string) string {
	return acctest.ConfigCompose(testAccFlowConfig_base(model), fmt.Sprintf(`
resource "aws_bedrockagent_flow" "test" {
  name               = %[1]q
  execution_role_arn = aws_iam_role.test.arn
  description        = "update"
}
`, rName))
}

func testAccFlowConfig_withEncryptionKey(rName, model string) string {
	return acctest.ConfigCompose(testAccFlowConfig_base(model), fmt.Sprintf(`
resource "aws_kms_key" "test" {
  description             = %[1]q
  deletion_window_in_days = 7
}

resource "aws_bedrockagent_flow" "test" {
  name                        = %[1]q
  execution_role_arn          = aws_iam_role.test.arn
  customer_encryption_key_arn = aws_kms_key.test.arn
}
`, rName))
}

func testAccFlowConfig_tags1(rName, model, tag1Key, tag1Value string) string {
	return acctest.ConfigCompose(testAccFlowConfig_base(model), fmt.Sprintf(`
resource "aws_bedrockagent_flow" "test" {
  name               = %[1]q
  execution_role_arn = aws_iam_role.test.arn

  tags = {
    %[2]q = %[3]q
  }
}
`, rName, tag1Key, tag1Value))
}

func testAccFlowConfig_tags2(rName, model, tag1Key, tag1Value, tag2Key, tag2Value string) string {
	return acctest.ConfigCompose(testAccFlowConfig_base(model), fmt.Sprintf(`
resource "aws_bedrockagent_flow" "test" {
  name               = %[1]q
  execution_role_arn = aws_iam_role.test.arn

  tags = {
    %[2]q = %[3]q
    %[4]q = %[5]q
  }
}
`, rName, tag1Key, tag1Value, tag2Key, tag2Value))
}

func testAccFlowConfig_withDefinition(rName, model string) string {
	return acctest.ConfigCompose(testAccFlowConfig_base(model), fmt.Sprintf(`
resource "aws_bedrockagent_flow" "test" {
  name               = %[1]q
  execution_role_arn = aws_iam_role.test.arn

  definition {
    connection {
      name   = "FlowInputNodeFlowInputNode0ToPrompt_1PromptsNode0"
      source = "FlowInputNode"
      target = "Prompt_1"
      type   = "Data"

      configuration {
        data {
          source_output = "document"
          target_input  = "topic"
        }
      }
    }
    connection {
      name   = "Prompt_1PromptsNode0ToFlowOutputNodeFlowOutputNode0"
      source = "Prompt_1"
      target = "FlowOutputNode"
      type   = "Data"

      configuration {
        data {
          source_output = "modelCompletion"
          target_input  = "document"
        }
      }
    }
    node {
      name = "FlowInputNode"
      type = "Input"

      configuration {
        input {}
      }

      output {
        name = "document"
        type = "String"
      }
    }
    node {
      name = "Prompt_1"
      type = "Prompt"

      configuration {
        prompt {
          source_configuration {
            inline {
              model_id      = %[2]q
              template_type = "TEXT"

              inference_configuration {
                text {
                  max_tokens     = 2048
                  stop_sequences = ["User:"]
                  temperature    = 0
                  top_p          = 0.8999999761581421
                }
              }

              template_configuration {
                text {
                  text = "Write a paragraph about {{topic}}."

                  input_variable {
                    name = "topic"
                  }
                }
              }
            }
          }
        }
      }

      input {
        expression = "$.data"
        name       = "topic"
        type       = "String"
      }

      output {
        name = "modelCompletion"
        type = "String"
      }
    }
    node {
      name = "FlowOutputNode"
      type = "Output"

      configuration {
        output {}
      }

      input {
        expression = "$.data"
        name       = "document"
        type       = "String"
      }
    }
  }
}
`, rName, model))
}

func testAccFlowConfig_withPromptResource(flowName, promptName, model string) string {
	return acctest.ConfigCompose(testAccFlowConfig_base(model), fmt.Sprintf(`
resource "aws_bedrockagent_prompt" "test" {
  name        = %[1]q
  description = "My prompt description."
}

resource "aws_bedrockagent_flow" "test" {
  name               = %[2]q
  execution_role_arn = aws_iam_role.test.arn

  definition {
    node {
      name = "Prompt_1"
      type = "Prompt"

      configuration {
        prompt {
          source_configuration {
            resource {
              prompt_arn = aws_bedrockagent_prompt.test.arn
            }
          }
        }
      }

      input {
        expression = "$.data"
        name       = "topic"
        type       = "String"
      }

      output {
        name = "modelCompletion"
        type = "String"
      }
    }
  }
}
`, promptName, flowName, model))
}
