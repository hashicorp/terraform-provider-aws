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
	"github.com/hashicorp/terraform-plugin-testing/config"
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

func init() {
	acctest.RegisterServiceErrorCheckFunc(names.BedrockAgentCoreServiceID, testAccErrorCheckSkip)
}

func testAccErrorCheckSkip(t *testing.T) resource.ErrorCheckFunc {
	return acctest.ErrorCheckSkipMessagesContaining(t,
		// aws_xray_trace_segment_destination not correctly configured:
		"X-Ray Delivery Destination is supported with CloudWatch",
		"log groups do not exist",
	)
}

func testAccRandomOnlineEvaluationConfigName(t *testing.T) string {
	return strings.ReplaceAll(acctest.RandomWithPrefix(t, acctest.ResourcePrefix), "-", "_")
}

func checkOnlineEvaluationConfigARN(name string) knownvalue.Check {
	return tfknownvalue.RegionalARNRegexp("bedrock-agentcore", regexache.MustCompile(`online-evaluation-config/`+name+`-[a-zA-Z0-9]{10}`))
}

func checkOnlineEvaluationConfigARNAlternateRegion(name string) knownvalue.Check {
	return tfknownvalue.RegionalARNAlternateRegionRegexp("bedrock-agentcore", regexache.MustCompile(`online-evaluation-config/`+name+`-[a-zA-Z0-9]{10}`))
}

func TestAccBedrockAgentCoreOnlineEvaluationConfig_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var v bedrockagentcorecontrol.GetOnlineEvaluationConfigOutput
	rName := testAccRandomOnlineEvaluationConfigName(t)
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
				ConfigDirectory: config.StaticDirectory("testdata/OnlineEvaluationConfig/basic/"),
				ConfigVariables: config.Variables{
					acctest.CtRName: config.StringVariable(rName),
				},
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckOnlineEvaluationConfigExists(ctx, t, resourceName, &v),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("online_evaluation_config_arn"), checkOnlineEvaluationConfigARN(rName)),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("online_evaluation_config_id"), knownvalue.NotNull()),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("online_evaluation_config_name"), knownvalue.StringExact(rName)),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrTags), knownvalue.Null()),
				},
			},
			{
				ConfigDirectory: config.StaticDirectory("testdata/OnlineEvaluationConfig/basic/"),
				ConfigVariables: config.Variables{
					acctest.CtRName: config.StringVariable(rName),
				},
				ResourceName:                         resourceName,
				ImportStateIdFunc:                    acctest.AttrImportStateIdFunc(resourceName, "online_evaluation_config_id"),
				ImportState:                          true,
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: "online_evaluation_config_id",
				ImportStateVerifyIgnore: []string{
					"enable_on_create",
				},
			},
		},
	})
}

func TestAccBedrockAgentCoreOnlineEvaluationConfig_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	var v bedrockagentcorecontrol.GetOnlineEvaluationConfigOutput
	rName := testAccRandomOnlineEvaluationConfigName(t)
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
				ConfigDirectory: config.StaticDirectory("testdata/OnlineEvaluationConfig/basic/"),
				ConfigVariables: config.Variables{
					acctest.CtRName: config.StringVariable(rName),
				},
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckOnlineEvaluationConfigExists(ctx, t, resourceName, &v),
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
	var v bedrockagentcorecontrol.GetOnlineEvaluationConfigOutput
	rName := testAccRandomOnlineEvaluationConfigName(t)
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
					testAccCheckOnlineEvaluationConfigExists(ctx, t, resourceName, &v),
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrTags), knownvalue.MapExact(map[string]knownvalue.Check{
						acctest.CtKey1: knownvalue.StringExact(acctest.CtValue1),
					})),
				},
			},
			{
				ResourceName:                         resourceName,
				ImportStateIdFunc:                    acctest.AttrImportStateIdFunc(resourceName, "online_evaluation_config_id"),
				ImportState:                          true,
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: "online_evaluation_config_id",
				ImportStateVerifyIgnore: []string{
					"enable_on_create",
				},
			},
			{
				Config: testAccOnlineEvaluationConfigConfig_tags2(rName, acctest.CtKey1, acctest.CtValue1Updated, acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckOnlineEvaluationConfigExists(ctx, t, resourceName, &v),
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
					testAccCheckOnlineEvaluationConfigExists(ctx, t, resourceName, &v),
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
	var v bedrockagentcorecontrol.GetOnlineEvaluationConfigOutput
	rName := testAccRandomOnlineEvaluationConfigName(t)
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
				ConfigDirectory: config.StaticDirectory("testdata/OnlineEvaluationConfig/basic/"),
				ConfigVariables: config.Variables{
					acctest.CtRName: config.StringVariable(rName),
				},
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckOnlineEvaluationConfigExists(ctx, t, resourceName, &v),
				),
			},
			{
				ConfigDirectory: config.StaticDirectory("testdata/OnlineEvaluationConfig/updated/"),
				ConfigVariables: config.Variables{
					acctest.CtRName: config.StringVariable(rName),
				},
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckOnlineEvaluationConfigExists(ctx, t, resourceName, &v),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionUpdate),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrDescription), knownvalue.StringExact("Updated description")),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrRule).AtSliceIndex(0).AtMapKey("sampling_config").AtSliceIndex(0).AtMapKey("sampling_percentage"), knownvalue.Float64Exact(50.0)),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrRule).AtSliceIndex(0).AtMapKey("session_config").AtSliceIndex(0).AtMapKey("session_timeout_minutes"), knownvalue.Int64Exact(30)),
				},
			},
		},
	})
}

func TestAccBedrockAgentCoreOnlineEvaluationConfig_executionStatus(t *testing.T) {
	ctx := acctest.Context(t)
	var v bedrockagentcorecontrol.GetOnlineEvaluationConfigOutput
	rName := testAccRandomOnlineEvaluationConfigName(t)
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
					testAccCheckOnlineEvaluationConfigExists(ctx, t, resourceName, &v),
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("execution_status"), knownvalue.StringExact("DISABLED")),
				},
			},
			{
				Config: testAccOnlineEvaluationConfigConfig_executionStatus(rName, "ENABLED"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckOnlineEvaluationConfigExists(ctx, t, resourceName, &v),
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
					testAccCheckOnlineEvaluationConfigExists(ctx, t, resourceName, &v),
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
	var v bedrockagentcorecontrol.GetOnlineEvaluationConfigOutput
	rName := testAccRandomOnlineEvaluationConfigName(t)
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
					testAccCheckOnlineEvaluationConfigExists(ctx, t, resourceName, &v),
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrRule).AtSliceIndex(0).AtMapKey(names.AttrFilter).AtSliceIndex(0).AtMapKey(names.AttrKey), knownvalue.StringExact(names.AttrEnvironment)),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrRule).AtSliceIndex(0).AtMapKey(names.AttrFilter).AtSliceIndex(0).AtMapKey("operator"), knownvalue.StringExact("Equals")),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrRule).AtSliceIndex(0).AtMapKey("sampling_config").AtSliceIndex(0).AtMapKey("sampling_percentage"), knownvalue.Float64Exact(25.0)),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrRule).AtSliceIndex(0).AtMapKey("session_config").AtSliceIndex(0).AtMapKey("session_timeout_minutes"), knownvalue.Int64Exact(20)),
				},
			},
		},
	})
}

func TestAccBedrockAgentCoreOnlineEvaluationConfig_insightsAndClustering(t *testing.T) {
	ctx := acctest.Context(t)
	var v bedrockagentcorecontrol.GetOnlineEvaluationConfigOutput
	rName := testAccRandomOnlineEvaluationConfigName(t)
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
				Config: testAccOnlineEvaluationConfigConfig_insightsAndClustering(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckOnlineEvaluationConfigExists(ctx, t, resourceName, &v),
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("insight").AtSliceIndex(0).AtMapKey("insight_id"), knownvalue.StringExact("Builtin.Insight.FailureAnalysis")),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("clustering_config").AtSliceIndex(0).AtMapKey("frequencies"), knownvalue.ListExact([]knownvalue.Check{
						knownvalue.StringExact("DAILY"),
						knownvalue.StringExact("WEEKLY"),
					})),
				},
			},
			{
				ResourceName:                         resourceName,
				ImportStateIdFunc:                    acctest.AttrImportStateIdFunc(resourceName, "online_evaluation_config_id"),
				ImportState:                          true,
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: "online_evaluation_config_id",
				ImportStateVerifyIgnore: []string{
					"enable_on_create",
				},
			},
		},
	})
}

func TestAccBedrockAgentCoreOnlineEvaluationConfig_insightOnly(t *testing.T) {
	ctx := acctest.Context(t)
	var v bedrockagentcorecontrol.GetOnlineEvaluationConfigOutput
	rName := testAccRandomOnlineEvaluationConfigName(t)
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
				// insight without clustering_config is a valid, service-accepted configuration.
				Config: testAccOnlineEvaluationConfigConfig_insightsWithoutClustering(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckOnlineEvaluationConfigExists(ctx, t, resourceName, &v),
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("insight").AtSliceIndex(0).AtMapKey("insight_id"), knownvalue.StringExact("Builtin.Insight.FailureAnalysis")),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("clustering_config"), knownvalue.ListSizeExact(0)),
				},
			},
			{
				ResourceName:                         resourceName,
				ImportStateIdFunc:                    acctest.AttrImportStateIdFunc(resourceName, "online_evaluation_config_id"),
				ImportState:                          true,
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: "online_evaluation_config_id",
				ImportStateVerifyIgnore: []string{
					"enable_on_create",
				},
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
data "aws_partition" "current" {}
data "aws_region" "current" {}
data "aws_caller_identity" "current" {}

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

resource "aws_iam_role_policy" "test" {
  role = aws_iam_role.test.name

  policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Sid": "CloudWatchLogReadStatement",
      "Effect": "Allow",
      "Action": [
        "logs:DescribeLogGroups",
        "logs:GetQueryResults",
        "logs:StartQuery"
      ],
      "Resource": "*"
    },
    {
      "Sid": "CloudWatchLogWriteStatement",
      "Effect": "Allow",
      "Action": [
        "logs:CreateLogGroup",
        "logs:CreateLogStream",
        "logs:PutLogEvents"
      ],
      "Resource": "arn:${data.aws_partition.current.partition}:logs:${data.aws_region.current.region}:${data.aws_caller_identity.current.account_id}:log-group:/aws/bedrock-agentcore/evaluations/*"
    },
    {
      "Sid": "CloudWatchIndexPolicyStatement",
      "Effect": "Allow",
      "Action": [
        "logs:DescribeIndexPolicies",
        "logs:PutIndexPolicy"
      ],
      "Resource": [
        "arn:${data.aws_partition.current.partition}:logs:${data.aws_region.current.region}:${data.aws_caller_identity.current.account_id}:log-group:aws/spans",
        "arn:${data.aws_partition.current.partition}:logs:${data.aws_region.current.region}:${data.aws_caller_identity.current.account_id}:log-group:aws/spans:*"
      ]
    },
    {
      "Sid": "BedrockInvokeStatement",
      "Effect": "Allow",
      "Action": [
        "bedrock:InvokeModel",
        "bedrock:InvokeModelWithResponseStream"
      ],
      "Resource": [
        "arn:${data.aws_partition.current.partition}:bedrock:${data.aws_region.current.region}::foundation-model/*",
        "arn:${data.aws_partition.current.partition}:bedrock:${data.aws_region.current.region}:${data.aws_caller_identity.current.account_id}:inference-profile/*"
      ]
    }
  ]
}
EOF
}

resource "aws_cloudwatch_log_group" "test" {
  name = "/aws/agentcore/%[1]s"
}
`, rName)
}

func testAccOnlineEvaluationConfigConfig_tags1(rName, tag1Key, tag1Value string) string {
	return acctest.ConfigCompose(testAccOnlineEvaluationConfigConfig_base(rName), fmt.Sprintf(`
resource "aws_bedrockagentcore_online_evaluation_config" "test" {
  online_evaluation_config_name = %[1]q
  enable_on_create              = false
  evaluation_execution_role_arn = aws_iam_role.test.arn

  data_source_config {
    cloudwatch_logs {
      log_group_names = [aws_cloudwatch_log_group.test.name]
      service_names   = ["strands_healthcare_single_agent.DEFAULT"]
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
    cloudwatch_logs {
      log_group_names = [aws_cloudwatch_log_group.test.name]
      service_names   = ["strands_healthcare_single_agent.DEFAULT"]
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
    cloudwatch_logs {
      log_group_names = [aws_cloudwatch_log_group.test.name]
      service_names   = ["strands_healthcare_single_agent.DEFAULT"]
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
    cloudwatch_logs {
      log_group_names = [aws_cloudwatch_log_group.test.name]
      service_names   = ["strands_healthcare_single_agent.DEFAULT"]
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

func testAccOnlineEvaluationConfigConfig_insightsAndClustering(rName string) string {
	return acctest.ConfigCompose(testAccOnlineEvaluationConfigConfig_base(rName), fmt.Sprintf(`
resource "aws_bedrockagentcore_online_evaluation_config" "test" {
  online_evaluation_config_name = %[1]q
  enable_on_create              = false
  evaluation_execution_role_arn = aws_iam_role.test.arn

  data_source_config {
    cloudwatch_logs {
      log_group_names = [aws_cloudwatch_log_group.test.name]
      service_names   = ["strands_healthcare_single_agent.DEFAULT"]
    }
  }

  insight {
    insight_id = "Builtin.Insight.FailureAnalysis"
  }

  clustering_config {
    frequencies = ["DAILY", "WEEKLY"]
  }

  rule {
    sampling_config {
      sampling_percentage = 10.0
    }
  }
}
`, rName))
}

func testAccOnlineEvaluationConfigConfig_insightsWithoutClustering(rName string) string {
	return acctest.ConfigCompose(testAccOnlineEvaluationConfigConfig_base(rName), fmt.Sprintf(`
resource "aws_bedrockagentcore_online_evaluation_config" "test" {
  online_evaluation_config_name = %[1]q
  enable_on_create              = false
  evaluation_execution_role_arn = aws_iam_role.test.arn

  data_source_config {
    cloudwatch_logs {
      log_group_names = [aws_cloudwatch_log_group.test.name]
      service_names   = ["strands_healthcare_single_agent.DEFAULT"]
    }
  }

  insight {
    insight_id = "Builtin.Insight.FailureAnalysis"
  }

  rule {
    sampling_config {
      sampling_percentage = 10.0
    }
  }
}
`, rName))
}

// TestAccBedrockAgentCoreOnlineEvaluationConfig_filterValueExactlyOneOf confirms
// the filter value union rejects multiple members (which silently dropped all but
// the first -> inconsistent result after apply) and an empty value {} (which
// surfaced an internal "always an error in the provider" message) at plan time.
func TestAccBedrockAgentCoreOnlineEvaluationConfig_filterValueExactlyOneOf(t *testing.T) {
	ctx := acctest.Context(t)
	rName := testAccRandomOnlineEvaluationConfigName(t)

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
				Config:      testAccOnlineEvaluationConfigConfig_filterValueMultiMember(rName),
				ExpectError: regexache.MustCompile("Invalid Attribute Combination"),
			},
			{
				Config:      testAccOnlineEvaluationConfigConfig_filterValueEmpty(rName),
				ExpectError: regexache.MustCompile("Invalid Attribute Combination"),
			},
		},
	})
}

// TestAccBedrockAgentCoreOnlineEvaluationConfig_frequenciesUnique confirms duplicate
// clustering frequencies are rejected at plan time instead of only by the API.
func TestAccBedrockAgentCoreOnlineEvaluationConfig_frequenciesUnique(t *testing.T) {
	ctx := acctest.Context(t)
	rName := testAccRandomOnlineEvaluationConfigName(t)

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
				Config:      testAccOnlineEvaluationConfigConfig_frequenciesDuplicate(rName),
				ExpectError: regexache.MustCompile("must not contain duplicate values"),
			},
		},
	})
}

// TestAccBedrockAgentCoreOnlineEvaluationConfig_clusteringRequiresInsight confirms
// clustering_config is rejected without insight at plan time (the service only
// accepts clusteringConfig when insights are provided).
func TestAccBedrockAgentCoreOnlineEvaluationConfig_clusteringRequiresInsight(t *testing.T) {
	ctx := acctest.Context(t)
	rName := testAccRandomOnlineEvaluationConfigName(t)

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
				Config:      testAccOnlineEvaluationConfigConfig_evaluatorClustering(rName),
				ExpectError: regexache.MustCompile("Invalid Attribute Combination"),
			},
		},
	})
}

// TestAccBedrockAgentCoreOnlineEvaluationConfig_evaluatorInsightSwitch confirms that
// switching between evaluators and insights forces replacement. The service cannot
// switch via update ("Cannot switch between evaluators and insights via update.
// Delete and recreate the config."), which previously failed at apply.
func TestAccBedrockAgentCoreOnlineEvaluationConfig_evaluatorInsightSwitch(t *testing.T) {
	ctx := acctest.Context(t)
	var v bedrockagentcorecontrol.GetOnlineEvaluationConfigOutput
	rName := testAccRandomOnlineEvaluationConfigName(t)
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
				Config: testAccOnlineEvaluationConfigConfig_evaluator(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckOnlineEvaluationConfigExists(ctx, t, resourceName, &v),
				),
			},
			{
				Config: testAccOnlineEvaluationConfigConfig_insightsAndClustering(rName),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionReplace),
					},
				},
			},
		},
	})
}

// TestAccBedrockAgentCoreOnlineEvaluationConfig_descriptionClear confirms the
// set->unset transition for description. The Update API retains description when
// omitted and rejects an empty string, so description is Optional+Computed;
// removing it from configuration must not produce a perpetual diff.
func TestAccBedrockAgentCoreOnlineEvaluationConfig_descriptionClear(t *testing.T) {
	ctx := acctest.Context(t)
	var v bedrockagentcorecontrol.GetOnlineEvaluationConfigOutput
	rName := testAccRandomOnlineEvaluationConfigName(t)
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
				Config: testAccOnlineEvaluationConfigConfig_evaluatorDescription(rName, "initial description"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckOnlineEvaluationConfigExists(ctx, t, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, "initial description"),
				),
			},
			{
				Config: testAccOnlineEvaluationConfigConfig_evaluator(rName),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PostApplyPostRefresh: []plancheck.PlanCheck{
						plancheck.ExpectEmptyPlan(),
					},
				},
			},
		},
	})
}

// TestAccBedrockAgentCoreOnlineEvaluationConfig_clusteringConfigClear confirms the
// set->unset transition for clustering_config converges. The Update API PATCH-merges and
// retains clustering_config when omitted (the SDK requires Frequencies, so an empty value
// is not a clear signal), so clustering_config is Optional+Computed and removing it must
// produce an empty re-plan instead of a perpetual diff.
func TestAccBedrockAgentCoreOnlineEvaluationConfig_clusteringConfigClear(t *testing.T) {
	ctx := acctest.Context(t)
	var v bedrockagentcorecontrol.GetOnlineEvaluationConfigOutput
	rName := testAccRandomOnlineEvaluationConfigName(t)
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
				Config: testAccOnlineEvaluationConfigConfig_insightsAndClustering(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckOnlineEvaluationConfigExists(ctx, t, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "clustering_config.#", "1"),
				),
			},
			{
				// Remove clustering_config (keep insight). Optional+Computed absorbs the
				// server-retained value, so the post-apply refresh plan is empty.
				Config: testAccOnlineEvaluationConfigConfig_insightsWithoutClustering(rName),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PostApplyPostRefresh: []plancheck.PlanCheck{
						plancheck.ExpectEmptyPlan(),
					},
				},
			},
		},
	})
}

// TestAccBedrockAgentCoreOnlineEvaluationConfig_enableOnCreateForcesNew confirms that
// changing enable_on_create forces replacement. It is a create-only field (absent
// from the Update input and Get output), so an in-place change never reached the API.
func TestAccBedrockAgentCoreOnlineEvaluationConfig_enableOnCreateForcesNew(t *testing.T) {
	ctx := acctest.Context(t)
	var v bedrockagentcorecontrol.GetOnlineEvaluationConfigOutput
	rName := testAccRandomOnlineEvaluationConfigName(t)
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
				Config: testAccOnlineEvaluationConfigConfig_enableOnCreate(rName, false),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckOnlineEvaluationConfigExists(ctx, t, resourceName, &v),
				),
			},
			{
				Config: testAccOnlineEvaluationConfigConfig_enableOnCreate(rName, true),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionReplace),
					},
				},
			},
		},
	})
}

// TestAccBedrockAgentCoreOnlineEvaluationConfig_filterClearing confirms the
// set->unset transition for the optional rule.filter block. Update re-sends the
// full rule, so removing the filter block must converge to an empty plan rather
// than a perpetual diff or an inconsistent result after apply.
func TestAccBedrockAgentCoreOnlineEvaluationConfig_filterClearing(t *testing.T) {
	ctx := acctest.Context(t)
	var v bedrockagentcorecontrol.GetOnlineEvaluationConfigOutput
	rName := testAccRandomOnlineEvaluationConfigName(t)
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
				Config: testAccOnlineEvaluationConfigConfig_evaluatorFilter(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckOnlineEvaluationConfigExists(ctx, t, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "rule.0.filter.#", "1"),
				),
			},
			{
				Config: testAccOnlineEvaluationConfigConfig_evaluator(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckOnlineEvaluationConfigExists(ctx, t, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "rule.0.filter.#", "0"),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PostApplyPostRefresh: []plancheck.PlanCheck{
						plancheck.ExpectEmptyPlan(),
					},
				},
			},
		},
	})
}

// TestAccBedrockAgentCoreOnlineEvaluationConfig_sessionConfigClearing confirms the
// set->unset transition for the optional rule.session_config block. Update re-sends
// the full rule, so removing the session_config block must converge to an empty plan
// rather than a perpetual diff or an inconsistent result after apply.
func TestAccBedrockAgentCoreOnlineEvaluationConfig_sessionConfigClearing(t *testing.T) {
	ctx := acctest.Context(t)
	var v bedrockagentcorecontrol.GetOnlineEvaluationConfigOutput
	rName := testAccRandomOnlineEvaluationConfigName(t)
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
				Config: testAccOnlineEvaluationConfigConfig_evaluatorSessionConfig(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckOnlineEvaluationConfigExists(ctx, t, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "rule.0.session_config.#", "1"),
				),
			},
			{
				Config: testAccOnlineEvaluationConfigConfig_evaluator(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckOnlineEvaluationConfigExists(ctx, t, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "rule.0.session_config.#", "0"),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PostApplyPostRefresh: []plancheck.PlanCheck{
						plancheck.ExpectEmptyPlan(),
					},
				},
			},
		},
	})
}

func testAccOnlineEvaluationConfigConfig_evaluator(rName string) string {
	return acctest.ConfigCompose(testAccOnlineEvaluationConfigConfig_base(rName), fmt.Sprintf(`
resource "aws_bedrockagentcore_online_evaluation_config" "test" {
  online_evaluation_config_name = %[1]q
  enable_on_create              = false
  evaluation_execution_role_arn = aws_iam_role.test.arn

  data_source_config {
    cloudwatch_logs {
      log_group_names = [aws_cloudwatch_log_group.test.name]
      service_names   = ["strands_healthcare_single_agent.DEFAULT"]
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

func testAccOnlineEvaluationConfigConfig_evaluatorFilter(rName string) string {
	return acctest.ConfigCompose(testAccOnlineEvaluationConfigConfig_base(rName), fmt.Sprintf(`
resource "aws_bedrockagentcore_online_evaluation_config" "test" {
  online_evaluation_config_name = %[1]q
  enable_on_create              = false
  evaluation_execution_role_arn = aws_iam_role.test.arn

  data_source_config {
    cloudwatch_logs {
      log_group_names = [aws_cloudwatch_log_group.test.name]
      service_names   = ["strands_healthcare_single_agent.DEFAULT"]
    }
  }

  evaluator {
    evaluator_id = "Builtin.Helpfulness"
  }

  rule {
    sampling_config {
      sampling_percentage = 10.0
    }

    filter {
      key      = "environment"
      operator = "Equals"

      value {
        string_value = "production"
      }
    }
  }
}
`, rName))
}

func testAccOnlineEvaluationConfigConfig_evaluatorSessionConfig(rName string) string {
	return acctest.ConfigCompose(testAccOnlineEvaluationConfigConfig_base(rName), fmt.Sprintf(`
resource "aws_bedrockagentcore_online_evaluation_config" "test" {
  online_evaluation_config_name = %[1]q
  enable_on_create              = false
  evaluation_execution_role_arn = aws_iam_role.test.arn

  data_source_config {
    cloudwatch_logs {
      log_group_names = [aws_cloudwatch_log_group.test.name]
      service_names   = ["strands_healthcare_single_agent.DEFAULT"]
    }
  }

  evaluator {
    evaluator_id = "Builtin.Helpfulness"
  }

  rule {
    sampling_config {
      sampling_percentage = 10.0
    }

    session_config {
      session_timeout_minutes = 20
    }
  }
}
`, rName))
}

func testAccOnlineEvaluationConfigConfig_evaluatorDescription(rName, description string) string {
	return acctest.ConfigCompose(testAccOnlineEvaluationConfigConfig_base(rName), fmt.Sprintf(`
resource "aws_bedrockagentcore_online_evaluation_config" "test" {
  online_evaluation_config_name = %[1]q
  enable_on_create              = false
  description                   = %[2]q
  evaluation_execution_role_arn = aws_iam_role.test.arn

  data_source_config {
    cloudwatch_logs {
      log_group_names = [aws_cloudwatch_log_group.test.name]
      service_names   = ["strands_healthcare_single_agent.DEFAULT"]
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
`, rName, description))
}

func testAccOnlineEvaluationConfigConfig_enableOnCreate(rName string, enable bool) string {
	return acctest.ConfigCompose(testAccOnlineEvaluationConfigConfig_base(rName), fmt.Sprintf(`
resource "aws_bedrockagentcore_online_evaluation_config" "test" {
  online_evaluation_config_name = %[1]q
  enable_on_create              = %[2]t
  evaluation_execution_role_arn = aws_iam_role.test.arn

  data_source_config {
    cloudwatch_logs {
      log_group_names = [aws_cloudwatch_log_group.test.name]
      service_names   = ["strands_healthcare_single_agent.DEFAULT"]
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
`, rName, enable))
}

func testAccOnlineEvaluationConfigConfig_filterValueMultiMember(rName string) string {
	return acctest.ConfigCompose(testAccOnlineEvaluationConfigConfig_base(rName), fmt.Sprintf(`
resource "aws_bedrockagentcore_online_evaluation_config" "test" {
  online_evaluation_config_name = %[1]q
  enable_on_create              = false
  evaluation_execution_role_arn = aws_iam_role.test.arn

  data_source_config {
    cloudwatch_logs {
      log_group_names = [aws_cloudwatch_log_group.test.name]
      service_names   = ["strands_healthcare_single_agent.DEFAULT"]
    }
  }

  evaluator {
    evaluator_id = "Builtin.Helpfulness"
  }

  rule {
    sampling_config {
      sampling_percentage = 10.0
    }

    filter {
      key      = "environment"
      operator = "Equals"

      value {
        string_value  = "production"
        boolean_value = true
      }
    }
  }
}
`, rName))
}

func testAccOnlineEvaluationConfigConfig_filterValueEmpty(rName string) string {
	return acctest.ConfigCompose(testAccOnlineEvaluationConfigConfig_base(rName), fmt.Sprintf(`
resource "aws_bedrockagentcore_online_evaluation_config" "test" {
  online_evaluation_config_name = %[1]q
  enable_on_create              = false
  evaluation_execution_role_arn = aws_iam_role.test.arn

  data_source_config {
    cloudwatch_logs {
      log_group_names = [aws_cloudwatch_log_group.test.name]
      service_names   = ["strands_healthcare_single_agent.DEFAULT"]
    }
  }

  evaluator {
    evaluator_id = "Builtin.Helpfulness"
  }

  rule {
    sampling_config {
      sampling_percentage = 10.0
    }

    filter {
      key      = "environment"
      operator = "Equals"

      value {}
    }
  }
}
`, rName))
}

func testAccOnlineEvaluationConfigConfig_frequenciesDuplicate(rName string) string {
	return acctest.ConfigCompose(testAccOnlineEvaluationConfigConfig_base(rName), fmt.Sprintf(`
resource "aws_bedrockagentcore_online_evaluation_config" "test" {
  online_evaluation_config_name = %[1]q
  enable_on_create              = false
  evaluation_execution_role_arn = aws_iam_role.test.arn

  data_source_config {
    cloudwatch_logs {
      log_group_names = [aws_cloudwatch_log_group.test.name]
      service_names   = ["strands_healthcare_single_agent.DEFAULT"]
    }
  }

  insight {
    insight_id = "Builtin.Insight.FailureAnalysis"
  }

  clustering_config {
    frequencies = ["MONTHLY", "MONTHLY"]
  }

  rule {
    sampling_config {
      sampling_percentage = 10.0
    }
  }
}
`, rName))
}

func testAccOnlineEvaluationConfigConfig_evaluatorClustering(rName string) string {
	return acctest.ConfigCompose(testAccOnlineEvaluationConfigConfig_base(rName), fmt.Sprintf(`
resource "aws_bedrockagentcore_online_evaluation_config" "test" {
  online_evaluation_config_name = %[1]q
  enable_on_create              = false
  evaluation_execution_role_arn = aws_iam_role.test.arn

  data_source_config {
    cloudwatch_logs {
      log_group_names = [aws_cloudwatch_log_group.test.name]
      service_names   = ["strands_healthcare_single_agent.DEFAULT"]
    }
  }

  evaluator {
    evaluator_id = "Builtin.Helpfulness"
  }

  clustering_config {
    frequencies = ["DAILY"]
  }

  rule {
    sampling_config {
      sampling_percentage = 10.0
    }
  }
}
`, rName))
}
