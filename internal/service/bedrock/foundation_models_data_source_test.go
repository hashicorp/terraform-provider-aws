// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package bedrock_test

import (
	"testing"

	"github.com/aws/aws-sdk-go/service/bedrock"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
)

func TestAccBedrockFoundationModelsDataSource_basic(t *testing.T) {
	ctx := acctest.Context(t)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, bedrock.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccFoundationModelsDataSourceConfig_basic(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.aws_bedrock_foundation_models.test", "id"),
					acctest.CheckResourceAttrGreaterThanValue("data.aws_bedrock_foundation_models.test", "model_summaries.#", 0),
				),
			},
		},
	})
}

func testAccFoundationModelsDataSourceConfig_basic() string {
	return `
data "aws_bedrock_foundation_models" "test" {
}
`
}
