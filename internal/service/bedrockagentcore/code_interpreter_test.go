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

func TestAccBedrockAgentCoreCodeInterpreter_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var codeInterpreter bedrockagentcorecontrol.GetCodeInterpreterOutput
	rName := strings.ReplaceAll(acctest.RandomWithPrefix(t, acctest.ResourcePrefix), "-", "_")
	resourceName := "aws_bedrockagentcore_code_interpreter.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.BedrockEndpointID)
			testAccPreCheckCodeInterpreters(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.BedrockAgentCoreServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckCodeInterpreterDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccCodeInterpreterConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckCodeInterpreterExists(ctx, t, resourceName, &codeInterpreter),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("code_interpreter_arn"), tfknownvalue.RegionalARNRegexp("bedrock-agentcore", regexache.MustCompile(`code-interpreter-custom/.+`))),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("code_interpreter_id"), knownvalue.NotNull()),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrTags), knownvalue.Null()),
				},
			},
			{
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateIdFunc:                    acctest.AttrImportStateIdFunc(resourceName, "code_interpreter_id"),
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: "code_interpreter_id",
			},
		},
	})
}

func TestAccBedrockAgentCoreCodeInterpreter_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	var codeinterpreter bedrockagentcorecontrol.GetCodeInterpreterOutput
	rName := strings.ReplaceAll(acctest.RandomWithPrefix(t, acctest.ResourcePrefix), "-", "_")
	resourceName := "aws_bedrockagentcore_code_interpreter.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.BedrockEndpointID)
			testAccPreCheckCodeInterpreters(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.BedrockAgentCoreServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckCodeInterpreterDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccCodeInterpreterConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckCodeInterpreterExists(ctx, t, resourceName, &codeinterpreter),
					acctest.CheckFrameworkResourceDisappears(ctx, t, tfbedrockagentcore.ResourceCodeInterpreter, resourceName),
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

func TestAccBedrockAgentCoreCodeInterpreter_tags(t *testing.T) {
	ctx := acctest.Context(t)
	var codeInterpreter bedrockagentcorecontrol.GetCodeInterpreterOutput
	rName := strings.ReplaceAll(acctest.RandomWithPrefix(t, acctest.ResourcePrefix), "-", "_")
	resourceName := "aws_bedrockagentcore_code_interpreter.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.BedrockEndpointID)
			testAccPreCheckCodeInterpreters(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.BedrockAgentCoreServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckCodeInterpreterDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccCodeInterpreterConfig_tags1(rName, acctest.CtKey1, acctest.CtValue1),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckCodeInterpreterExists(ctx, t, resourceName, &codeInterpreter),
				),
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
			},
			{
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateIdFunc:                    acctest.AttrImportStateIdFunc(resourceName, "code_interpreter_id"),
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: "code_interpreter_id",
			},
			{
				Config: testAccCodeInterpreterConfig_tags2(rName, acctest.CtKey1, acctest.CtValue1Updated, acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckCodeInterpreterExists(ctx, t, resourceName, &codeInterpreter),
				),
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
			},
			{
				Config: testAccCodeInterpreterConfig_tags1(rName, acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckCodeInterpreterExists(ctx, t, resourceName, &codeInterpreter),
				),
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
			},
		},
	})
}

func TestAccBedrockAgentCoreCodeInterpreter_full(t *testing.T) {
	ctx := acctest.Context(t)
	var codeInterpreter bedrockagentcorecontrol.GetCodeInterpreterOutput
	rName := strings.ReplaceAll(acctest.RandomWithPrefix(t, acctest.ResourcePrefix), "-", "_")
	resourceName := "aws_bedrockagentcore_code_interpreter.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.BedrockEndpointID)
			testAccPreCheckCodeInterpreters(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.BedrockAgentCoreServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckCodeInterpreterDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccCodeInterpreterConfig_full(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckCodeInterpreterExists(ctx, t, resourceName, &codeInterpreter),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
					},
				},
			},
			{
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateIdFunc:                    acctest.AttrImportStateIdFunc(resourceName, "code_interpreter_id"),
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: "code_interpreter_id",
			},
		},
	})
}

func testAccCheckCodeInterpreterDestroy(ctx context.Context, t *testing.T) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.ProviderMeta(ctx, t).BedrockAgentCoreClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_bedrockagentcore_code_interpreter" {
				continue
			}

			_, err := tfbedrockagentcore.FindCodeInterpreterByID(ctx, conn, rs.Primary.Attributes["code_interpreter_id"])
			if retry.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("Bedrock Agent Core Code Interpreter %s still exists", rs.Primary.Attributes["code_interpreter_id"])
		}

		return nil
	}
}

func testAccCheckCodeInterpreterExists(ctx context.Context, t *testing.T, n string, v *bedrockagentcorecontrol.GetCodeInterpreterOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.ProviderMeta(ctx, t).BedrockAgentCoreClient(ctx)

		resp, err := tfbedrockagentcore.FindCodeInterpreterByID(ctx, conn, rs.Primary.Attributes["code_interpreter_id"])
		if err != nil {
			return err
		}

		*v = *resp

		return nil
	}
}

func testAccPreCheckCodeInterpreters(ctx context.Context, t *testing.T) {
	conn := acctest.ProviderMeta(ctx, t).BedrockAgentCoreClient(ctx)

	input := bedrockagentcorecontrol.ListCodeInterpretersInput{}

	_, err := conn.ListCodeInterpreters(ctx, &input)

	if acctest.PreCheckSkipError(err) {
		t.Skipf("skipping acceptance testing: %s", err)
	}
	if err != nil {
		t.Fatalf("unexpected PreCheck error: %s", err)
	}
}

func testAccCodeInterpreterConfig_full(rName string) string {
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

resource "aws_bedrockagentcore_code_interpreter" "test" {
  name               = %[1]q
  description        = "test"
  execution_role_arn = aws_iam_role.test.arn

  network_configuration {
    network_mode = "SANDBOX"
  }
}
`, rName)
}

func testAccCodeInterpreterConfig_basic(rName string) string {
	return fmt.Sprintf(`
resource "aws_bedrockagentcore_code_interpreter" "test" {
  name = %[1]q

  network_configuration {
    network_mode = "SANDBOX"
  }
}
`, rName)
}

func testAccCodeInterpreterConfig_tags1(rName, tagKey1, tagValue1 string) string {
	return fmt.Sprintf(`
resource "aws_bedrockagentcore_code_interpreter" "test" {
  name = %[1]q

  network_configuration {
    network_mode = "PUBLIC"
  }

  tags = {
    %[2]q = %[3]q
  }
}
`, rName, tagKey1, tagValue1)
}

func testAccCodeInterpreterConfig_tags2(rName, tagKey1, tagValue1, tag2Key, tag2Value string) string {
	return fmt.Sprintf(`
resource "aws_bedrockagentcore_code_interpreter" "test" {
  name = %[1]q

  network_configuration {
    network_mode = "PUBLIC"
  }

  tags = {
    %[2]q = %[3]q
    %[4]q = %[5]q
  }
}
`, rName, tagKey1, tagValue1, tag2Key, tag2Value)
}
