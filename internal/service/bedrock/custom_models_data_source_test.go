// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package bedrock_test

import (
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/bedrock"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccBedrockCustomModelsDataSource_basic(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	datasourceName := "data.aws_bedrock_custom_models.test"
	resourceName := "aws_bedrock_custom_model.test"
	var v bedrock.GetModelCustomizationJobOutput

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckPartitionHasService(t, names.BedrockEndpointID) },
		ErrorCheck:               acctest.ErrorCheck(t, names.BedrockServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccCustomModelConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCustomModelExists(ctx, resourceName, &v),
				),
			},
			{
				PreConfig: func() {
					testAccWaitModelCustomizationJobCompleted(ctx, t, &v)
				},
				Config: testAccCustomModelsDataSourceConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(datasourceName, names.AttrID),
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
data "aws_bedrock_custom_models" "test" {
  depends_on = [aws_bedrock_custom_model.test]
}
`)
}
