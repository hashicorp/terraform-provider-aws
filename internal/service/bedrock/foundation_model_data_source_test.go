// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package bedrock_test

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccBedrockFoundationModelDataSource_basic(t *testing.T) {
	ctx := acctest.Context(t)
	datasourceName := "data.aws_bedrock_foundation_model.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckPartitionHasService(t, names.BedrockEndpointID) },
		ErrorCheck:               acctest.ErrorCheck(t, names.BedrockServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccFoundationModelDataSourceConfig_basic(),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet(datasourceName, "customizations_supported.#"),
					resource.TestCheckResourceAttrSet(datasourceName, names.AttrID),
					resource.TestCheckResourceAttrSet(datasourceName, "inference_types_supported.#"),
					resource.TestCheckResourceAttrSet(datasourceName, "input_modalities.#"),
					resource.TestCheckResourceAttrSet(datasourceName, "model_arn"),
					resource.TestCheckResourceAttrSet(datasourceName, "model_name"),
					resource.TestCheckResourceAttrSet(datasourceName, "output_modalities.#"),
					resource.TestCheckResourceAttrSet(datasourceName, names.AttrProviderName),
					resource.TestCheckResourceAttrSet(datasourceName, "response_streaming_supported"),
				),
			},
		},
	})
}

func testAccFoundationModelDataSourceConfig_basic() string {
	return `
data "aws_bedrock_foundation_models" "test" {}

data "aws_bedrock_foundation_model" "test" {
  model_id = data.aws_bedrock_foundation_models.test.model_summaries[0].model_id
}
`
}
