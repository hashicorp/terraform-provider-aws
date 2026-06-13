// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package bedrockagentcore_test

import (
	"context"
	"fmt"
	"strings"
	"testing"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/service/bedrockagentcorecontrol"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	tfknownvalue "github.com/hashicorp/terraform-provider-aws/internal/acctest/knownvalue"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	tfbedrockagentcore "github.com/hashicorp/terraform-provider-aws/internal/service/bedrockagentcore"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccBedrockAgentCoreEvaluator_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var evaluator bedrockagentcorecontrol.GetEvaluatorOutput
	rName := randomEvaluatorName(t)
	resourceName := "aws_bedrockagentcore_evaluator.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.BedrockEndpointID)
			testAccPreCheckEvaluators(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.BedrockAgentCoreServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckEvaluatorDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccEvaluatorConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckEvaluatorExists(ctx, t, resourceName, &evaluator),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("evaluator_arn"), tfknownvalue.RegionalARNRegexp("bedrock-agentcore", regexache.MustCompile(`evaluator/.+`))),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("evaluator_id"), knownvalue.NotNull()),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("evaluator_name"), knownvalue.StringExact(rName)),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("level"), knownvalue.StringExact("TRACE")),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrTags), knownvalue.Null()),
				},
			},
			{
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateIdFunc:                    acctest.AttrImportStateIdFunc(resourceName, "evaluator_id"),
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: "evaluator_id",
			},
		},
	})
}

func TestAccBedrockAgentCoreEvaluator_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	var evaluator bedrockagentcorecontrol.GetEvaluatorOutput
	rName := randomEvaluatorName(t)
	resourceName := "aws_bedrockagentcore_evaluator.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.BedrockEndpointID)
			testAccPreCheckEvaluators(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.BedrockAgentCoreServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckEvaluatorDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccEvaluatorConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckEvaluatorExists(ctx, t, resourceName, &evaluator),
					acctest.CheckFrameworkResourceDisappears(ctx, t, tfbedrockagentcore.ResourceEvaluator, resourceName),
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

func TestAccBedrockAgentCoreEvaluator_description(t *testing.T) {
	ctx := acctest.Context(t)
	var evaluator bedrockagentcorecontrol.GetEvaluatorOutput
	rName := randomEvaluatorName(t)
	resourceName := "aws_bedrockagentcore_evaluator.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.BedrockEndpointID)
			testAccPreCheckEvaluators(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.BedrockAgentCoreServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckEvaluatorDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccEvaluatorConfig_description(rName, "desc1"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckEvaluatorExists(ctx, t, resourceName, &evaluator),
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrDescription), knownvalue.StringExact("desc1")),
				},
			},
			{
				Config: testAccEvaluatorConfig_description(rName, "desc2"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckEvaluatorExists(ctx, t, resourceName, &evaluator),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionUpdate),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrDescription), knownvalue.StringExact("desc2")),
				},
			},
		},
	})
}

func testAccCheckEvaluatorDestroy(ctx context.Context, t *testing.T) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.ProviderMeta(ctx, t).BedrockAgentCoreClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_bedrockagentcore_evaluator" {
				continue
			}

			_, err := tfbedrockagentcore.FindEvaluatorByID(ctx, conn, rs.Primary.Attributes["evaluator_id"])
			if retry.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("Bedrock Agent Core Evaluator %s still exists", rs.Primary.Attributes["evaluator_id"])
		}

		return nil
	}
}

func testAccCheckEvaluatorExists(ctx context.Context, t *testing.T, n string, v *bedrockagentcorecontrol.GetEvaluatorOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.ProviderMeta(ctx, t).BedrockAgentCoreClient(ctx)

		resp, err := tfbedrockagentcore.FindEvaluatorByID(ctx, conn, rs.Primary.Attributes["evaluator_id"])
		if err != nil {
			return err
		}

		*v = *resp

		return nil
	}
}

func testAccPreCheckEvaluators(ctx context.Context, t *testing.T) {
	conn := acctest.ProviderMeta(ctx, t).BedrockAgentCoreClient(ctx)

	input := bedrockagentcorecontrol.ListEvaluatorsInput{}

	_, err := conn.ListEvaluators(ctx, &input)

	if acctest.PreCheckSkipError(err) {
		t.Skipf("skipping acceptance testing: %s", err)
	}
	if err != nil {
		t.Fatalf("unexpected PreCheck error: %s", err)
	}
}

func randomEvaluatorName(t *testing.T) string {
	return strings.ReplaceAll(fmt.Sprintf("tf_acc_test_%s", acctest.RandString(t, 10)), "-", "_")
}

func TestAccBedrockAgentCoreEvaluator_categorical(t *testing.T) {
	ctx := acctest.Context(t)
	var evaluator bedrockagentcorecontrol.GetEvaluatorOutput
	rName := randomEvaluatorName(t)
	resourceName := "aws_bedrockagentcore_evaluator.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.BedrockEndpointID)
			testAccPreCheckEvaluators(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.BedrockAgentCoreServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckEvaluatorDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccEvaluatorConfig_categorical(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckEvaluatorExists(ctx, t, resourceName, &evaluator),
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("evaluator_config").AtSliceIndex(0).AtMapKey("llm_as_a_judge").AtSliceIndex(0).AtMapKey("rating_scale").AtSliceIndex(0).AtMapKey("categorical").AtSliceIndex(0).AtMapKey("label"), knownvalue.StringExact("POSITIVE")),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("evaluator_config").AtSliceIndex(0).AtMapKey("llm_as_a_judge").AtSliceIndex(0).AtMapKey("rating_scale").AtSliceIndex(0).AtMapKey("categorical").AtSliceIndex(2).AtMapKey("label"), knownvalue.StringExact("NEGATIVE")),
				},
			},
			{
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateIdFunc:                    acctest.AttrImportStateIdFunc(resourceName, "evaluator_id"),
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: "evaluator_id",
			},
		},
	})
}

func TestAccBedrockAgentCoreEvaluator_inferenceConfig(t *testing.T) {
	ctx := acctest.Context(t)
	var evaluator bedrockagentcorecontrol.GetEvaluatorOutput
	rName := randomEvaluatorName(t)
	resourceName := "aws_bedrockagentcore_evaluator.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.BedrockEndpointID)
			testAccPreCheckEvaluators(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.BedrockAgentCoreServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckEvaluatorDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccEvaluatorConfig_inferenceConfig(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckEvaluatorExists(ctx, t, resourceName, &evaluator),
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("evaluator_config").AtSliceIndex(0).AtMapKey("llm_as_a_judge").AtSliceIndex(0).AtMapKey("model_config").AtSliceIndex(0).AtMapKey("bedrock_evaluator_model_config").AtSliceIndex(0).AtMapKey("inference_config").AtSliceIndex(0).AtMapKey("max_tokens"), knownvalue.Int32Exact(1024)),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("evaluator_config").AtSliceIndex(0).AtMapKey("llm_as_a_judge").AtSliceIndex(0).AtMapKey("model_config").AtSliceIndex(0).AtMapKey("bedrock_evaluator_model_config").AtSliceIndex(0).AtMapKey("inference_config").AtSliceIndex(0).AtMapKey("top_p"), knownvalue.Float64Exact(0.5)),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("evaluator_config").AtSliceIndex(0).AtMapKey("llm_as_a_judge").AtSliceIndex(0).AtMapKey("model_config").AtSliceIndex(0).AtMapKey("bedrock_evaluator_model_config").AtSliceIndex(0).AtMapKey("inference_config").AtSliceIndex(0).AtMapKey("stop_sequences"), knownvalue.ListExact([]knownvalue.Check{knownvalue.StringExact("STOP")})),
				},
			},
			{
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateIdFunc:                    acctest.AttrImportStateIdFunc(resourceName, "evaluator_id"),
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: "evaluator_id",
			},
		},
	})
}

func TestAccBedrockAgentCoreEvaluator_codeBased(t *testing.T) {
	ctx := acctest.Context(t)
	var evaluator bedrockagentcorecontrol.GetEvaluatorOutput
	rName := randomEvaluatorName(t)
	resourceName := "aws_bedrockagentcore_evaluator.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.BedrockEndpointID)
			testAccPreCheckEvaluators(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.BedrockAgentCoreServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckEvaluatorDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccEvaluatorConfig_codeBased(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckEvaluatorExists(ctx, t, resourceName, &evaluator),
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("evaluator_config").AtSliceIndex(0).AtMapKey("code_based").AtSliceIndex(0).AtMapKey("lambda_config").AtSliceIndex(0).AtMapKey("lambda_timeout_in_seconds"), knownvalue.Int32Exact(120)),
				},
			},
			{
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateIdFunc:                    acctest.AttrImportStateIdFunc(resourceName, "evaluator_id"),
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: "evaluator_id",
			},
		},
	})
}

func TestAccBedrockAgentCoreEvaluator_kmsKey(t *testing.T) {
	ctx := acctest.Context(t)
	var evaluator bedrockagentcorecontrol.GetEvaluatorOutput
	rName := randomEvaluatorName(t)
	resourceName := "aws_bedrockagentcore_evaluator.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.BedrockEndpointID)
			testAccPreCheckEvaluators(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.BedrockAgentCoreServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckEvaluatorDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccEvaluatorConfig_kmsKey(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckEvaluatorExists(ctx, t, resourceName, &evaluator),
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrKMSKeyARN), tfknownvalue.RegionalARNRegexp("kms", regexache.MustCompile(`key/.+`))),
				},
			},
		},
	})
}

func TestAccBedrockAgentCoreEvaluator_additionalModelRequestFields(t *testing.T) {
	ctx := acctest.Context(t)
	var evaluator bedrockagentcorecontrol.GetEvaluatorOutput
	rName := randomEvaluatorName(t)
	resourceName := "aws_bedrockagentcore_evaluator.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.BedrockEndpointID)
			testAccPreCheckEvaluators(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.BedrockAgentCoreServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckEvaluatorDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccEvaluatorConfig_additionalModelRequestFields(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckEvaluatorExists(ctx, t, resourceName, &evaluator),
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("evaluator_config").AtSliceIndex(0).AtMapKey("llm_as_a_judge").AtSliceIndex(0).AtMapKey("model_config").AtSliceIndex(0).AtMapKey("bedrock_evaluator_model_config").AtSliceIndex(0).AtMapKey("additional_model_request_fields"), knownvalue.StringExact(`{"top_k":50}`)),
				},
			},
			{
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateIdFunc:                    acctest.AttrImportStateIdFunc(resourceName, "evaluator_id"),
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: "evaluator_id",
			},
		},
	})
}

func testAccEvaluatorConfig_basic(rName string) string {
	return fmt.Sprintf(`
resource "aws_bedrockagentcore_evaluator" "test" {
  evaluator_name = %[1]q
  level          = "TRACE"

  evaluator_config {
    llm_as_a_judge {
      instructions = "Given the {context} and the {assistant_turn}, compare against {expected_response} and rate from 1 to 5."

      rating_scale {
        numerical {
          definition = "Not helpful at all."
          value      = 1
          label      = "1"
        }
        numerical {
          definition = "Extremely helpful."
          value      = 5
          label      = "5"
        }
      }

      model_config {
        bedrock_evaluator_model_config {
          model_id = "global.anthropic.claude-sonnet-4-5-20250929-v1:0"
        }
      }
    }
  }
}
`, rName)
}

func testAccEvaluatorConfig_description(rName, description string) string {
	return fmt.Sprintf(`
resource "aws_bedrockagentcore_evaluator" "test" {
  evaluator_name = %[1]q
  description    = %[2]q
  level          = "TRACE"

  evaluator_config {
    llm_as_a_judge {
      instructions = "Given the {context} and the {assistant_turn}, compare against {expected_response} and rate from 1 to 5."

      rating_scale {
        numerical {
          definition = "Not helpful at all."
          value      = 1
          label      = "1"
        }
        numerical {
          definition = "Extremely helpful."
          value      = 5
          label      = "5"
        }
      }

      model_config {
        bedrock_evaluator_model_config {
          model_id = "global.anthropic.claude-sonnet-4-5-20250929-v1:0"
        }
      }
    }
  }
}
`, rName, description)
}

func testAccEvaluatorConfig_categorical(rName string) string {
	return fmt.Sprintf(`
resource "aws_bedrockagentcore_evaluator" "test" {
  evaluator_name = %[1]q
  level          = "TRACE"

  evaluator_config {
    llm_as_a_judge {
      instructions = "Classify the tone of the {assistant_turn} given the {context} and {expected_response}."

      rating_scale {
        categorical {
          definition = "Friendly, helpful tone."
          label      = "POSITIVE"
        }
        categorical {
          definition = "Neutral or terse tone."
          label      = "NEUTRAL"
        }
        categorical {
          definition = "Unhelpful or dismissive tone."
          label      = "NEGATIVE"
        }
      }

      model_config {
        bedrock_evaluator_model_config {
          model_id = "global.anthropic.claude-sonnet-4-5-20250929-v1:0"
        }
      }
    }
  }
}
`, rName)
}

func testAccEvaluatorConfig_inferenceConfig(rName string) string {
	return fmt.Sprintf(`
resource "aws_bedrockagentcore_evaluator" "test" {
  evaluator_name = %[1]q
  level          = "TRACE"

  evaluator_config {
    llm_as_a_judge {
      instructions = "Given the {context} and the {assistant_turn}, compare against {expected_response} and rate from 1 to 5."

      rating_scale {
        numerical {
          definition = "Not helpful at all."
          value      = 1
          label      = "1"
        }
        numerical {
          definition = "Extremely helpful."
          value      = 5
          label      = "5"
        }
      }

      model_config {
        bedrock_evaluator_model_config {
          model_id = "global.anthropic.claude-sonnet-4-5-20250929-v1:0"

          inference_config {
            max_tokens     = 1024
            top_p          = 0.5
            stop_sequences = ["STOP"]
          }
        }
      }
    }
  }
}
`, rName)
}

func testAccEvaluatorConfig_additionalModelRequestFields(rName string) string {
	return fmt.Sprintf(`
resource "aws_bedrockagentcore_evaluator" "test" {
  evaluator_name = %[1]q
  level          = "TRACE"

  evaluator_config {
    llm_as_a_judge {
      instructions = "Given the {context} and the {assistant_turn}, compare against {expected_response} and rate from 1 to 5."

      rating_scale {
        numerical {
          definition = "Not helpful at all."
          value      = 1
          label      = "1"
        }
        numerical {
          definition = "Extremely helpful."
          value      = 5
          label      = "5"
        }
      }

      model_config {
        bedrock_evaluator_model_config {
          model_id                        = "global.anthropic.claude-sonnet-4-5-20250929-v1:0"
          additional_model_request_fields = jsonencode({ top_k = 50 })
        }
      }
    }
  }
}
`, rName)
}

func testAccEvaluatorConfig_codeBased(rName string) string {
	return acctest.ConfigCompose(testAccEvaluatorConfig_lambda(rName), fmt.Sprintf(`
resource "aws_bedrockagentcore_evaluator" "test" {
  evaluator_name = %[1]q
  level          = "TRACE"

  evaluator_config {
    code_based {
      lambda_config {
        lambda_arn                = aws_lambda_function.test.arn
        lambda_timeout_in_seconds = 120
      }
    }
  }
}
`, rName))
}

func testAccEvaluatorConfig_kmsKey(rName string) string {
	return fmt.Sprintf(`
resource "aws_kms_key" "test" {
  description             = "Test key for %[1]s"
  deletion_window_in_days = 7
}

resource "aws_bedrockagentcore_evaluator" "test" {
  evaluator_name = %[1]q
  level          = "TRACE"
  kms_key_arn    = aws_kms_key.test.arn

  evaluator_config {
    llm_as_a_judge {
      instructions = "Given the {context} and the {assistant_turn}, compare against {expected_response} and rate from 1 to 5."

      rating_scale {
        numerical {
          definition = "Not helpful at all."
          value      = 1
          label      = "1"
        }
        numerical {
          definition = "Extremely helpful."
          value      = 5
          label      = "5"
        }
      }

      model_config {
        bedrock_evaluator_model_config {
          model_id = "global.anthropic.claude-sonnet-4-5-20250929-v1:0"
        }
      }
    }
  }
}
`, rName)
}

func testAccEvaluatorConfig_lambda(rName string) string {
	return fmt.Sprintf(`
data "aws_iam_policy_document" "lambda_assume" {
  statement {
    effect  = "Allow"
    actions = ["sts:AssumeRole"]
    principals {
      type        = "Service"
      identifiers = ["lambda.amazonaws.com"]
    }
  }
}

resource "aws_iam_role" "lambda" {
  name               = "%[1]s-lambda"
  assume_role_policy = data.aws_iam_policy_document.lambda_assume.json
}

resource "aws_lambda_function" "test" {
  filename      = "test-fixtures/lambdatest.zip"
  function_name = %[1]q
  role          = aws_iam_role.lambda.arn
  handler       = "lambdatest.handler"
  runtime       = "nodejs24.x"
}
`, rName)
}
