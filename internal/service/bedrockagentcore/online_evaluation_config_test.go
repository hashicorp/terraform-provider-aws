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

func TestAccBedrockAgentCoreOnlineEvaluationConfig_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var config bedrockagentcorecontrol.GetOnlineEvaluationConfigOutput
	rName := strings.ReplaceAll(acctest.RandomWithPrefix(t, acctest.ResourcePrefix), "-", "_")
	resourceName := "aws_bedrockagentcore_online_evaluation_config.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.BedrockEndpointID)
			testAccPreCheckOnlineEvaluationConfig(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.BedrockAgentCoreServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckOnlineEvaluationConfigDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccOnlineEvaluationConfigConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckOnlineEvaluationConfigExists(ctx, t, resourceName, &config),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("online_evaluation_config_arn"), tfknownvalue.RegionalARNRegexp("bedrock-agentcore", regexache.MustCompile(`online-evaluation-config/.+`))),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("online_evaluation_config_id"), knownvalue.NotNull()),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("online_evaluation_config_name"), knownvalue.StringExact(rName)),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrStatus), knownvalue.NotNull()),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrTags), knownvalue.Null()),
				},
			},
			{
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateVerify:                    true,
				ImportStateIdFunc:                    acctest.AttrImportStateIdFunc(resourceName, "online_evaluation_config_id"),
				ImportStateVerifyIdentifierAttribute: "online_evaluation_config_id",
				ImportStateVerifyIgnore:              []string{"enable_on_create"},
			},
		},
	})
}

func TestAccBedrockAgentCoreOnlineEvaluationConfig_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	var config bedrockagentcorecontrol.GetOnlineEvaluationConfigOutput
	rName := strings.ReplaceAll(acctest.RandomWithPrefix(t, acctest.ResourcePrefix), "-", "_")
	resourceName := "aws_bedrockagentcore_online_evaluation_config.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.BedrockEndpointID)
			testAccPreCheckOnlineEvaluationConfig(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.BedrockAgentCoreServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckOnlineEvaluationConfigDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccOnlineEvaluationConfigConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckOnlineEvaluationConfigExists(ctx, t, resourceName, &config),
					acctest.CheckFrameworkResourceDisappears(ctx, t, tfbedrockagentcore.ResourceOnlineEvaluationConfig, resourceName),
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

func TestAccBedrockAgentCoreOnlineEvaluationConfig_tags(t *testing.T) {
	ctx := acctest.Context(t)
	var config bedrockagentcorecontrol.GetOnlineEvaluationConfigOutput
	rName := strings.ReplaceAll(acctest.RandomWithPrefix(t, acctest.ResourcePrefix), "-", "_")
	resourceName := "aws_bedrockagentcore_online_evaluation_config.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.BedrockEndpointID)
			testAccPreCheckOnlineEvaluationConfig(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.BedrockAgentCoreServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckOnlineEvaluationConfigDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccOnlineEvaluationConfigConfig_tags1(rName, acctest.CtKey1, acctest.CtValue1),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckOnlineEvaluationConfigExists(ctx, t, resourceName, &config),
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrTags), knownvalue.MapExact(map[string]knownvalue.Check{
						acctest.CtKey1: knownvalue.StringExact(acctest.CtValue1),
					})),
				},
			},
			{
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateVerify:                    true,
				ImportStateIdFunc:                    acctest.AttrImportStateIdFunc(resourceName, "online_evaluation_config_id"),
				ImportStateVerifyIdentifierAttribute: "online_evaluation_config_id",
				ImportStateVerifyIgnore:              []string{"enable_on_create"},
			},
			{
				Config: testAccOnlineEvaluationConfigConfig_tags2(rName, acctest.CtKey1, acctest.CtValue1Updated, acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckOnlineEvaluationConfigExists(ctx, t, resourceName, &config),
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrTags), knownvalue.MapExact(map[string]knownvalue.Check{
						acctest.CtKey1: knownvalue.StringExact(acctest.CtValue1Updated),
						acctest.CtKey2: knownvalue.StringExact(acctest.CtValue2),
					})),
				},
			},
			{
				Config: testAccOnlineEvaluationConfigConfig_tags1(rName, acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckOnlineEvaluationConfigExists(ctx, t, resourceName, &config),
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrTags), knownvalue.MapExact(map[string]knownvalue.Check{
						acctest.CtKey2: knownvalue.StringExact(acctest.CtValue2),
					})),
				},
			},
		},
	})
}

func TestAccBedrockAgentCoreOnlineEvaluationConfig_update(t *testing.T) {
	ctx := acctest.Context(t)
	var config bedrockagentcorecontrol.GetOnlineEvaluationConfigOutput
	rName := strings.ReplaceAll(acctest.RandomWithPrefix(t, acctest.ResourcePrefix), "-", "_")
	resourceName := "aws_bedrockagentcore_online_evaluation_config.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.BedrockEndpointID)
			testAccPreCheckOnlineEvaluationConfig(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.BedrockAgentCoreServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckOnlineEvaluationConfigDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccOnlineEvaluationConfigConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckOnlineEvaluationConfigExists(ctx, t, resourceName, &config),
				),
			},
			{
				Config: testAccOnlineEvaluationConfigConfig_updated(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckOnlineEvaluationConfigExists(ctx, t, resourceName, &config),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionUpdate),
					},
				},
			},
			{
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateVerify:                    true,
				ImportStateIdFunc:                    acctest.AttrImportStateIdFunc(resourceName, "online_evaluation_config_id"),
				ImportStateVerifyIdentifierAttribute: "online_evaluation_config_id",
				ImportStateVerifyIgnore:              []string{"enable_on_create"},
			},
		},
	})
}

func TestAccBedrockAgentCoreOnlineEvaluationConfig_executionStatus(t *testing.T) {
	ctx := acctest.Context(t)
	var config bedrockagentcorecontrol.GetOnlineEvaluationConfigOutput
	rName := strings.ReplaceAll(acctest.RandomWithPrefix(t, acctest.ResourcePrefix), "-", "_")
	resourceName := "aws_bedrockagentcore_online_evaluation_config.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.BedrockEndpointID)
			testAccPreCheckOnlineEvaluationConfig(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.BedrockAgentCoreServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckOnlineEvaluationConfigDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccOnlineEvaluationConfigConfig_executionStatus(rName, "DISABLED"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckOnlineEvaluationConfigExists(ctx, t, resourceName, &config),
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("execution_status"), knownvalue.StringExact("DISABLED")),
				},
			},
			{
				Config: testAccOnlineEvaluationConfigConfig_executionStatus(rName, "ENABLED"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckOnlineEvaluationConfigExists(ctx, t, resourceName, &config),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionUpdate),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("execution_status"), knownvalue.StringExact("ENABLED")),
				},
			},
			{
				Config: testAccOnlineEvaluationConfigConfig_executionStatus(rName, "DISABLED"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckOnlineEvaluationConfigExists(ctx, t, resourceName, &config),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionUpdate),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("execution_status"), knownvalue.StringExact("DISABLED")),
				},
			},
		},
	})
}

func TestAccBedrockAgentCoreOnlineEvaluationConfig_filters(t *testing.T) {
	ctx := acctest.Context(t)
	var config bedrockagentcorecontrol.GetOnlineEvaluationConfigOutput
	rName := strings.ReplaceAll(acctest.RandomWithPrefix(t, acctest.ResourcePrefix), "-", "_")
	resourceName := "aws_bedrockagentcore_online_evaluation_config.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.BedrockEndpointID)
			testAccPreCheckOnlineEvaluationConfig(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.BedrockAgentCoreServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckOnlineEvaluationConfigDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccOnlineEvaluationConfigConfig_filters(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckOnlineEvaluationConfigExists(ctx, t, resourceName, &config),
				),
			},
			{
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateVerify:                    true,
				ImportStateIdFunc:                    acctest.AttrImportStateIdFunc(resourceName, "online_evaluation_config_id"),
				ImportStateVerifyIdentifierAttribute: "online_evaluation_config_id",
				ImportStateVerifyIgnore:              []string{"enable_on_create"},
			},
		},
	})
}

// Test helper functions.

func testAccCheckOnlineEvaluationConfigDestroy(ctx context.Context, t *testing.T) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.ProviderMeta(ctx, t).BedrockAgentCoreClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_bedrockagentcore_online_evaluation_config" {
				continue
			}

			_, err := tfbedrockagentcore.FindOnlineEvaluationConfigByID(ctx, conn, rs.Primary.Attributes["online_evaluation_config_id"])
			if retry.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("Bedrock AgentCore Online Evaluation Config %s still exists", rs.Primary.Attributes["online_evaluation_config_id"])
		}

		return nil
	}
}

func testAccCheckOnlineEvaluationConfigExists(ctx context.Context, t *testing.T, n string, v *bedrockagentcorecontrol.GetOnlineEvaluationConfigOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.ProviderMeta(ctx, t).BedrockAgentCoreClient(ctx)

		resp, err := tfbedrockagentcore.FindOnlineEvaluationConfigByID(ctx, conn, rs.Primary.Attributes["online_evaluation_config_id"])
		if err != nil {
			return err
		}

		*v = *resp

		return nil
	}
}

func testAccPreCheckOnlineEvaluationConfig(ctx context.Context, t *testing.T) {
	conn := acctest.ProviderMeta(ctx, t).BedrockAgentCoreClient(ctx)

	input := bedrockagentcorecontrol.ListOnlineEvaluationConfigsInput{}

	_, err := conn.ListOnlineEvaluationConfigs(ctx, &input)

	if acctest.PreCheckSkipError(err) {
		t.Skipf("skipping acceptance testing: %s", err)
	}
	if err != nil {
		t.Fatalf("unexpected PreCheck error: %s", err)
	}
}

// Test configurations.

func testAccOnlineEvaluationConfigConfig_base(rName string) string {
	return fmt.Sprintf(`
data "aws_iam_policy_document" "test" {
  statement {
    effect  = "Allow"
    actions = ["sts:AssumeRole"]
    principals {
      type        = "Service"
      identifiers = ["bedrock-agentcore.amazonaws.com"]
    }
  }
}

resource "aws_iam_role" "test" {
  name               = %[1]q
  assume_role_policy = data.aws_iam_policy_document.test.json
}

resource "aws_cloudwatch_log_group" "test" {
  name = "/aws/agentcore/%[1]s"
}
`, rName)
}

func testAccOnlineEvaluationConfigConfig_basic(rName string) string {
	return acctest.ConfigCompose(testAccOnlineEvaluationConfigConfig_base(rName), fmt.Sprintf(`
resource "aws_bedrockagentcore_online_evaluation_config" "test" {
  online_evaluation_config_name = %[1]q
  enable_on_create              = false
  evaluation_execution_role_arn = aws_iam_role.test.arn

  data_source_config {
    cloud_watch_logs {
      log_group_names = [aws_cloudwatch_log_group.test.name]
      service_names   = ["test_service"]
    }
  }

  evaluator {
    evaluator_id = "Builtin.Helpfulness"
  }

  rule {
    sampling_config {
      sampling_percentage = 10.0
    }
  }
}
`, rName))
}

func testAccOnlineEvaluationConfigConfig_updated(rName string) string {
	return acctest.ConfigCompose(testAccOnlineEvaluationConfigConfig_base(rName), fmt.Sprintf(`
resource "aws_bedrockagentcore_online_evaluation_config" "test" {
  online_evaluation_config_name = %[1]q
  description                   = "Updated description"
  enable_on_create              = false
  evaluation_execution_role_arn = aws_iam_role.test.arn

  data_source_config {
    cloud_watch_logs {
      log_group_names = [aws_cloudwatch_log_group.test.name]
      service_names   = ["test_service"]
    }
  }

  evaluator {
    evaluator_id = "Builtin.Helpfulness"
  }

  evaluator {
    evaluator_id = "Builtin.GoalSuccessRate"
  }

  rule {
    sampling_config {
      sampling_percentage = 50.0
    }

    session_config {
      session_timeout_minutes = 30
    }
  }
}
`, rName))
}

func testAccOnlineEvaluationConfigConfig_tags1(rName, tag1Key, tag1Value string) string {
	return acctest.ConfigCompose(testAccOnlineEvaluationConfigConfig_base(rName), fmt.Sprintf(`
resource "aws_bedrockagentcore_online_evaluation_config" "test" {
  online_evaluation_config_name = %[1]q
  enable_on_create              = false
  evaluation_execution_role_arn = aws_iam_role.test.arn

  data_source_config {
    cloud_watch_logs {
      log_group_names = [aws_cloudwatch_log_group.test.name]
      service_names   = ["test_service"]
    }
  }

  evaluator {
    evaluator_id = "Builtin.Helpfulness"
  }

  rule {
    sampling_config {
      sampling_percentage = 10.0
    }
  }

  tags = {
    %[2]q = %[3]q
  }
}
`, rName, tag1Key, tag1Value))
}

func testAccOnlineEvaluationConfigConfig_tags2(rName, tag1Key, tag1Value, tag2Key, tag2Value string) string {
	return acctest.ConfigCompose(testAccOnlineEvaluationConfigConfig_base(rName), fmt.Sprintf(`
resource "aws_bedrockagentcore_online_evaluation_config" "test" {
  online_evaluation_config_name = %[1]q
  enable_on_create              = false
  evaluation_execution_role_arn = aws_iam_role.test.arn

  data_source_config {
    cloud_watch_logs {
      log_group_names = [aws_cloudwatch_log_group.test.name]
      service_names   = ["test_service"]
    }
  }

  evaluator {
    evaluator_id = "Builtin.Helpfulness"
  }

  rule {
    sampling_config {
      sampling_percentage = 10.0
    }
  }

  tags = {
    %[2]q = %[3]q
    %[4]q = %[5]q
  }
}
`, rName, tag1Key, tag1Value, tag2Key, tag2Value))
}

func testAccOnlineEvaluationConfigConfig_executionStatus(rName, executionStatus string) string {
	return acctest.ConfigCompose(testAccOnlineEvaluationConfigConfig_base(rName), fmt.Sprintf(`
resource "aws_bedrockagentcore_online_evaluation_config" "test" {
  online_evaluation_config_name = %[1]q
  enable_on_create              = false
  execution_status              = %[2]q
  evaluation_execution_role_arn = aws_iam_role.test.arn

  data_source_config {
    cloud_watch_logs {
      log_group_names = [aws_cloudwatch_log_group.test.name]
      service_names   = ["test_service"]
    }
  }

  evaluator {
    evaluator_id = "Builtin.Helpfulness"
  }

  rule {
    sampling_config {
      sampling_percentage = 10.0
    }
  }
}
`, rName, executionStatus))
}

func testAccOnlineEvaluationConfigConfig_filters(rName string) string {
	return acctest.ConfigCompose(testAccOnlineEvaluationConfigConfig_base(rName), fmt.Sprintf(`
resource "aws_bedrockagentcore_online_evaluation_config" "test" {
  online_evaluation_config_name = %[1]q
  enable_on_create              = false
  evaluation_execution_role_arn = aws_iam_role.test.arn

  data_source_config {
    cloud_watch_logs {
      log_group_names = [aws_cloudwatch_log_group.test.name]
      service_names   = ["test_service"]
    }
  }

  evaluator {
    evaluator_id = "Builtin.Helpfulness"
  }

  rule {
    sampling_config {
      sampling_percentage = 25.0
    }

    filter {
      key      = "environment"
      operator = "Equals"

      value {
        string_value = "production"
      }
    }

    session_config {
      session_timeout_minutes = 20
    }
  }
}
`, rName))
}
