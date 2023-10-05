// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package bedrock_test

import (
	"testing"

	"github.com/aws/aws-sdk-go/service/bedrock"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
)

func TestAccBedrockFoundationModelDataSource_basic(t *testing.T) {
	ctx := acctest.Context(t)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, bedrock.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccBedrockFoundationModelDataSourceConfig_basic(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.aws_bedrock_foundation_model.test", "id"),
					resource.TestCheckResourceAttrSet("data.aws_bedrock_foundation_model.test", "model_id"),
				),
			},
		},
	})
}

func testAccBedrockFoundationModelDataSourceConfig_basic() string {
	return `
data "aws_bedrock_foundation_models" "test" {

}

data "aws_bedrock_foundation_model" "test" {
  model_id = data.aws_bedrock_foundation_models.test.model_summaries[0].model_id
}
`
}
