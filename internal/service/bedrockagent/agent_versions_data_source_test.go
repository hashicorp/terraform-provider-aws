// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package bedrockagent_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccBedrockAgentVersionsDataSource_basic(t *testing.T) {
	ctx := acctest.Context(t)
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	dataSourceName := "data.aws_bedrockagent_agent_versions.test"
	resourceName := "aws_bedrockagent_agent.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.BedrockAgentServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccAgentVersionsDataSourceConfig_basic(rName, "anthropic.claude-v2", "basic claude"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(dataSourceName, "agent_id", resourceName, "agent_id"),
					resource.TestCheckResourceAttrSet(dataSourceName, "agent_version_summaries.#"),
					resource.TestCheckResourceAttrSet(dataSourceName, "agent_version_summaries.0.%"),
					resource.TestCheckResourceAttrSet(dataSourceName, "agent_version_summaries.0.agent_name"),
					resource.TestCheckResourceAttrSet(dataSourceName, "agent_version_summaries.0.agent_version"),
				),
			},
		},
	})
}

func testAccAgentVersionsDataSourceConfig_basic(rName, model, desc string) string {
	return acctest.ConfigCompose(testAccAgent_base(rName, model), fmt.Sprintf(`
resource "aws_bedrockagent_agent" "test" {
  agent_name                  = %[1]q
  agent_resource_role_arn     = aws_iam_role.test_agent.arn
  description                 = %[3]q
  idle_session_ttl_in_seconds = 500
  instruction                 = file("${path.module}/test-fixtures/instruction.txt")
  foundation_model            = %[2]q
}

data "aws_bedrockagent_agent_versions" "test" {
  agent_id = aws_bedrockagent_agent.test.agent_id
}
`, rName, model, desc))
}
