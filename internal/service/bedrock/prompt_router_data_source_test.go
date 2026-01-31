// Copyright IBM Corp. 2014, 2025
// SPDX-License-Identifier: MPL-2.0

package bedrock_test

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccBedrockPromptRouterDataSource_basic(t *testing.T) {
	ctx := acctest.Context(t)
	datasourceName := "data.aws_bedrock_prompt_router.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckPartitionHasService(t, names.BedrockEndpointID) },
		ErrorCheck:               acctest.ErrorCheck(t, names.BedrockServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccPromptRouterDataSourceConfig_basic(),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet(datasourceName, "prompt_router_arn"),
					resource.TestCheckResourceAttrSet(datasourceName, "prompt_router_name"),
					resource.TestCheckResourceAttrSet(datasourceName, names.AttrStatus),
					resource.TestCheckResourceAttrSet(datasourceName, names.AttrType),
					resource.TestCheckResourceAttrSet(datasourceName, names.AttrCreatedAt),
					resource.TestCheckResourceAttrSet(datasourceName, "updated_at"),
				),
			},
		},
	})
}

func testAccPromptRouterDataSourceConfig_basic() string {
	return `
data "aws_bedrock_prompt_routers" "test" {}

data "aws_bedrock_prompt_router" "test" {
  prompt_router_arn = data.aws_bedrock_prompt_routers.test.prompt_router_summaries[0].prompt_router_arn
}
`
}
