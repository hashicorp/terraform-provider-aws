// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package bedrockagent_test

import (
	"fmt"
	"testing"

	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccBedrockAgentAgentPrepare_basic(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_bedrockagent_agent_prepare.test"
	agentResourceName := "aws_bedrockagent_agent.test"
	model := "anthropic.claude-3-5-sonnet-20241022-v2:0"
	description := "basic claude"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.BedrockEndpointID)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.BedrockServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccAgentPrepareConfig_basic(rName, model, description),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrPair(resourceName, names.AttrID, agentResourceName, names.AttrID),
					resource.TestCheckResourceAttrSet(resourceName, "prepared_at"),
				),
			},
		},
	})
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
