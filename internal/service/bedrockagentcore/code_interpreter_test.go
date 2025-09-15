// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package bedrockagentcore_test

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"testing"
	"time"

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
					resource.TestCheckResourceAttrSet(resourceName, names.AttrID),
					resource.TestCheckNoResourceAttr(resourceName, names.AttrExecutionRoleARN),
					acctest.MatchResourceAttrRegionalARN(ctx, resourceName, names.AttrARN, "bedrock-agentcore", regexache.MustCompile(`code-interpreter-custom/.+$`)),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"client_token"},
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
				Config: testAccMemoryConfig_iamRole(rName),
				Check: resource.ComposeTestCheckFunc(
					acctest.CheckSleep(t, 5*time.Second),
				),
			},
			{
				Config: testAccCodeInterpreterConfig(rName, "test description", types.CodeInterpreterNetworkModeSandbox, true),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckCodeInterpreterExists(ctx, resourceName, &codeInterpreter),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, "test description"),
					resource.TestCheckResourceAttr(resourceName, "network_configuration.network_mode", string(types.CodeInterpreterNetworkModeSandbox)),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrID),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrExecutionRoleARN),
					acctest.MatchResourceAttrRegionalARN(ctx, resourceName, names.AttrARN, "bedrock-agentcore", regexache.MustCompile(`code-interpreter-custom/.+$`)),
				),
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
			_, err := tfbedrockagentcore.FindCodeInterpreterByID(ctx, conn, rs.Primary.ID)
			if tfresource.NotFound(err) {
				return nil
			}
			if err != nil {
				return create.Error(names.BedrockAgentCore, create.ErrActionCheckingDestroyed, tfbedrockagentcore.ResNameCodeInterpreter, rs.Primary.ID, err)
			}

			return create.Error(names.BedrockAgentCore, create.ErrActionCheckingDestroyed, tfbedrockagentcore.ResNameCodeInterpreter, rs.Primary.ID, errors.New("not destroyed"))
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

		if rs.Primary.ID == "" {
			return create.Error(names.BedrockAgentCore, create.ErrActionCheckingExistence, tfbedrockagentcore.ResNameCodeInterpreter, name, errors.New("not set"))
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).BedrockAgentCoreClient(ctx)

		resp, err := tfbedrockagentcore.FindCodeInterpreterByID(ctx, conn, rs.Primary.ID)
		if err != nil {
			return create.Error(names.BedrockAgentCore, create.ErrActionCheckingExistence, tfbedrockagentcore.ResNameCodeInterpreter, rs.Primary.ID, err)
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

func testAccCodeInterpreterConfig(rName, description string, networkMode types.CodeInterpreterNetworkMode, withRole bool) string {
	role := "aws_iam_role.test.arn"
	if !withRole {
		role = "null"
	}

	return acctest.ConfigCompose(testAccMemoryConfig_iamRole(rName), fmt.Sprintf(`
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
