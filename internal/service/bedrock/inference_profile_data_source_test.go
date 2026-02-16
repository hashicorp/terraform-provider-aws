// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package bedrock_test

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccBedrockInferenceProfileDataSource_basic(t *testing.T) {
	ctx := acctest.Context(t)
	datasourceName := "data.aws_bedrock_inference_profile.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckPartitionHasService(t, names.BedrockEndpointID) },
		ErrorCheck:               acctest.ErrorCheck(t, names.BedrockServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccInferenceProfileDataSourceConfig_basic(),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet(datasourceName, "inference_profile_arn"),
					resource.TestCheckResourceAttrSet(datasourceName, "inference_profile_id"),
					resource.TestCheckResourceAttrSet(datasourceName, "inference_profile_name"),
					resource.TestCheckResourceAttrSet(datasourceName, "models.#"),
					resource.TestCheckResourceAttrSet(datasourceName, names.AttrStatus),
					resource.TestCheckResourceAttrSet(datasourceName, names.AttrType),
					resource.TestCheckResourceAttrSet(datasourceName, names.AttrCreatedAt),
					resource.TestCheckResourceAttrSet(datasourceName, names.AttrDescription),
					resource.TestCheckResourceAttrSet(datasourceName, "updated_at"),
				),
			},
		},
	})
}

func testAccInferenceProfileDataSourceConfig_basic() string {
	return `
data "aws_bedrock_inference_profiles" "test" {}

data "aws_bedrock_inference_profile" "test" {
  inference_profile_id = data.aws_bedrock_inference_profiles.test.inference_profile_summaries[0].inference_profile_id
}
`
}
