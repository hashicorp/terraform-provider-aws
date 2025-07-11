// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package bedrockagent_test

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/bedrockagent"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	tfbedrockagent "github.com/hashicorp/terraform-provider-aws/internal/service/bedrockagent"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccBedrockAgentAgentPrepare_basic(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var agent bedrockagent.GetAgentOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_bedrockagent_agent_prepare.test"
	agentResourceName := "aws_bedrockagent_agent.test"
	model := "anthropic.claude-3-5-sonnet-20241022-v2:0"
	description := "basic claude"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.BedrockAgentEndpointID)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.BedrockAgentServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckAgentPrepareDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccAgentPrepareConfig_basic(rName, model, description),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAgentPrepareExists(ctx, resourceName, &agent),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrID, agentResourceName, names.AttrID),
					resource.TestCheckResourceAttrSet(resourceName, "prepared_at"),
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

func TestAccBedrockAgentAgentPrepare_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var agent bedrockagent.GetAgentOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_bedrockagent_agent_prepare.test"
	model := "anthropic.claude-3-5-sonnet-20241022-v2:0"
	description := "basic claude"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.BedrockAgentEndpointID)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.BedrockAgentServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckAgentPrepareDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccAgentPrepareConfig_basic(rName, model, description),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAgentPrepareExists(ctx, resourceName, &agent),
					acctest.CheckFrameworkResourceDisappears(ctx, acctest.Provider, tfbedrockagent.ResourceAgentPrepare, resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckAgentPrepareDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		// Since agent prepare is just an action, we don't need to check for destruction
		// The underlying agent may still exist, which is expected
		return nil
	}
}

func testAccCheckAgentPrepareExists(ctx context.Context, name string, agent *bedrockagent.GetAgentOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return create.Error(names.BedrockAgent, create.ErrActionCheckingExistence, tfbedrockagent.ResNameAgentPrepare, name, errors.New("not found"))
		}

		if rs.Primary.ID == "" {
			return create.Error(names.BedrockAgent, create.ErrActionCheckingExistence, tfbedrockagent.ResNameAgentPrepare, name, errors.New("id not set"))
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).BedrockAgentClient(ctx)

		resp, err := conn.GetAgent(ctx, &bedrockagent.GetAgentInput{
			AgentId: &rs.Primary.ID,
		})
		if err != nil {
			return create.Error(names.BedrockAgent, create.ErrActionCheckingExistence, tfbedrockagent.ResNameAgentPrepare, rs.Primary.ID, err)
		}

		*agent = *resp

		return nil
	}
}

func testAccAgentPrepareConfig_basic(rName, model, description string) string {
	return acctest.ConfigCompose(
		testAccAgent_base(rName, model),
		fmt.Sprintf(`
resource "aws_bedrockagent_agent" "test" {
  agent_name                  = "%[1]s"
  agent_resource_role_arn     = aws_iam_role.test_agent.arn
  description                 = %[3]q
  idle_session_ttl_in_seconds = 500
  instruction                 = file("${path.module}/test-fixtures/instruction.txt")
  foundation_model            = %[2]q
  prepare_agent               = false
}

resource "aws_bedrockagent_agent_prepare" "test" {
  id = aws_bedrockagent_agent.test.id
}`, rName, model, description))
}
