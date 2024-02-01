// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package bedrock_test

import (
	"testing"

	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccBedrockCustomModelsDataSource_basic(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	datasourceName := "data.aws_bedrock_custom_models.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckPartitionHasService(t, names.BedrockEndpointID) },
		ErrorCheck:               acctest.ErrorCheck(t, names.BedrockEndpointID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			// TODO: Create custom model and wait for it to be created.
			{
				Config: testAccCustomModelsDataSourceConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(datasourceName, "id"),
					acctest.CheckResourceAttrGreaterThanValue(datasourceName, "model_summaries.#", 0),
					resource.TestCheckResourceAttrSet(datasourceName, "model_summaries.0.creation_time"),
					resource.TestCheckResourceAttrSet(datasourceName, "model_summaries.0.model_arn"),
					resource.TestCheckResourceAttrSet(datasourceName, "model_summaries.0.model_name"),
				),
			},
		},
	})
}

func testAccCustomModelsDataSourceConfig_basic(rName string) string {
	return acctest.ConfigCompose(testAccCustomModelConfig_basic(rName), `
data "aws_bedrock_custom_models" "test" {}
`)
}
