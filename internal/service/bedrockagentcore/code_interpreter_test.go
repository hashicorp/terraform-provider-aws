// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package bedrockagentcore_test

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"testing"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/service/bedrockagentcorecontrol"
	"github.com/aws/aws-sdk-go-v2/service/bedrockagentcorecontrol/types"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	tfbedrockagentcore "github.com/hashicorp/terraform-provider-aws/internal/service/bedrockagentcore"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccBedrockAgentCoreCodeInterpreter_basic(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var codeInterpreter bedrockagentcorecontrol.GetCodeInterpreterOutput
	rName := strings.ReplaceAll(sdkacctest.RandomWithPrefix(acctest.ResourcePrefix), "-", "_")
	resourceName := "aws_bedrockagentcore_code_interpreter.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.BedrockEndpointID)
			testAccPreCheckCodeInterpreters(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.BedrockAgentCoreServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckCodeInterpreterDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccCodeInterpreterConfig(rName, "test description", types.CodeInterpreterNetworkModePublic, false),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckCodeInterpreterExists(ctx, resourceName, &codeInterpreter),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, "test description"),
					resource.TestCheckResourceAttr(resourceName, "network_configuration.network_mode", string(types.CodeInterpreterNetworkModePublic)),
					resource.TestCheckResourceAttrSet(resourceName, "code_interpreter_id"),
					resource.TestCheckNoResourceAttr(resourceName, names.AttrExecutionRoleARN),
					acctest.MatchResourceAttrRegionalARN(ctx, resourceName, "code_interpreter_arn", "bedrock-agentcore", regexache.MustCompile(`code-interpreter-custom/.+$`)),
				),
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

func TestAccBedrockAgentCoreCodeInterpreter_role(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var codeInterpreter bedrockagentcorecontrol.GetCodeInterpreterOutput
	rName := strings.ReplaceAll(sdkacctest.RandomWithPrefix(acctest.ResourcePrefix), "-", "_")
	resourceName := "aws_bedrockagentcore_code_interpreter.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.BedrockEndpointID)
			testAccPreCheckCodeInterpreters(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.BedrockAgentCoreServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckCodeInterpreterDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccCodeInterpreterConfig(rName, "test description", types.CodeInterpreterNetworkModeSandbox, true),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckCodeInterpreterExists(ctx, resourceName, &codeInterpreter),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, "test description"),
					resource.TestCheckResourceAttr(resourceName, "network_configuration.network_mode", string(types.CodeInterpreterNetworkModeSandbox)),
					resource.TestCheckResourceAttrSet(resourceName, "code_interpreter_id"),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrExecutionRoleARN),
					acctest.MatchResourceAttrRegionalARN(ctx, resourceName, "code_interpreter_arn", "bedrock-agentcore", regexache.MustCompile(`code-interpreter-custom/.+$`)),
				),
			},
		},
	})
}

func TestAccBedrockAgentCoreCodeInterpreter_tags(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var codeInterpreter bedrockagentcorecontrol.GetCodeInterpreterOutput
	rName := strings.ReplaceAll(sdkacctest.RandomWithPrefix(acctest.ResourcePrefix), "-", "_")
	resourceName := "aws_bedrockagentcore_code_interpreter.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.BedrockEndpointID)
			testAccPreCheckCodeInterpreters(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.BedrockAgentCoreServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckCodeInterpreterDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccCodeInterpreterConfig_tags(rName, acctest.CtKey1, acctest.CtValue1),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckCodeInterpreterExists(ctx, resourceName, &codeInterpreter),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey1, acctest.CtValue1),
				),
			},
			{
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateIdFunc:                    acctest.AttrImportStateIdFunc(resourceName, "code_interpreter_id"),
				ImportStateVerify:                    true,
				ImportStateVerifyIgnore:              []string{"client_token"},
				ImportStateVerifyIdentifierAttribute: "code_interpreter_id",
			},
		},
	})
}

func TestAccBedrockAgentCoreCodeInterpreter_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var codeinterpreter bedrockagentcorecontrol.GetCodeInterpreterOutput
	rName := strings.ReplaceAll(sdkacctest.RandomWithPrefix(acctest.ResourcePrefix), "-", "_")
	resourceName := "aws_bedrockagentcore_code_interpreter.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.BedrockEndpointID)
			testAccPreCheckCodeInterpreters(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.BedrockAgentCoreServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckCodeInterpreterDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccCodeInterpreterConfig(rName, "test description", types.CodeInterpreterNetworkModePublic, false),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckCodeInterpreterExists(ctx, resourceName, &codeinterpreter),
					acctest.CheckFrameworkResourceDisappears(ctx, acctest.Provider, tfbedrockagentcore.ResourceCodeInterpreter, resourceName),
				),
				ExpectNonEmptyPlan: true,
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PostApplyPostRefresh: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
					},
				},
			},
		},
	})
}

func testAccCheckCodeInterpreterDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).BedrockAgentCoreClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_bedrockagentcore_code_interpreter" {
				continue
			}
			id := rs.Primary.Attributes["code_interpreter_id"]
			_, err := tfbedrockagentcore.FindCodeInterpreterByID(ctx, conn, id)
			if tfresource.NotFound(err) {
				return nil
			}
			if err != nil {
				return create.Error(names.BedrockAgentCore, create.ErrActionCheckingDestroyed, tfbedrockagentcore.ResNameCodeInterpreter, id, err)
			}

			return create.Error(names.BedrockAgentCore, create.ErrActionCheckingDestroyed, tfbedrockagentcore.ResNameCodeInterpreter, id, errors.New("not destroyed"))
		}

		return nil
	}
}

func testAccCheckCodeInterpreterExists(ctx context.Context, name string, codeinterpreter *bedrockagentcorecontrol.GetCodeInterpreterOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return create.Error(names.BedrockAgentCore, create.ErrActionCheckingExistence, tfbedrockagentcore.ResNameCodeInterpreter, name, errors.New("not found"))
		}

		id := rs.Primary.Attributes["code_interpreter_id"]
		if id == "" {
			return create.Error(names.BedrockAgentCore, create.ErrActionCheckingExistence, tfbedrockagentcore.ResNameCodeInterpreter, name, errors.New("not set"))
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).BedrockAgentCoreClient(ctx)

		resp, err := tfbedrockagentcore.FindCodeInterpreterByID(ctx, conn, id)
		if err != nil {
			return create.Error(names.BedrockAgentCore, create.ErrActionCheckingExistence, tfbedrockagentcore.ResNameCodeInterpreter, id, err)
		}

		*codeinterpreter = *resp

		return nil
	}
}

func testAccPreCheckCodeInterpreters(ctx context.Context, t *testing.T) {
	conn := acctest.Provider.Meta().(*conns.AWSClient).BedrockAgentCoreClient(ctx)

	input := bedrockagentcorecontrol.ListCodeInterpretersInput{}

	_, err := conn.ListCodeInterpreters(ctx, &input)

	if acctest.PreCheckSkipError(err) {
		t.Skipf("skipping acceptance testing: %s", err)
	}
	if err != nil {
		t.Fatalf("unexpected PreCheck error: %s", err)
	}
}

func testAccCodeInspectorConfig_IAMRole(rName string) string {
	return fmt.Sprintf(`
data "aws_partition" "current" {}

data "aws_iam_policy_document" "test_assume" {
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
  assume_role_policy = data.aws_iam_policy_document.test_assume.json
}
`, rName)
}

func testAccCodeInterpreterConfig(rName, description string, networkMode types.CodeInterpreterNetworkMode, withRole bool) string {
	role := "aws_iam_role.test.arn"
	if !withRole {
		role = "null"
	}

	return acctest.ConfigCompose(testAccCodeInspectorConfig_IAMRole(rName), fmt.Sprintf(`
resource "aws_bedrockagentcore_code_interpreter" "test" {
  name               = %[1]q
  description        = %[2]q
  execution_role_arn = %[3]s

  network_configuration = {
    network_mode = %[4]q
  }
}
`, rName, description, role, networkMode))
}

func testAccCodeInterpreterConfig_tags(rName, tagKey1, tagValue1 string) string {
	return acctest.ConfigCompose(testAccCodeInspectorConfig_IAMRole(rName), fmt.Sprintf(`
resource "aws_bedrockagentcore_code_interpreter" "test" {
  name               = %[1]q
  description        = "test description"
  execution_role_arn = null

  network_configuration = {
    network_mode = %[2]q
  }

  tags = {
    %[3]q = %[4]q
  }
}
`, rName, types.CodeInterpreterNetworkModePublic, tagKey1, tagValue1))
}
