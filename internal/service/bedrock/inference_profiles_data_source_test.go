// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package bedrock_test

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/bedrock/types"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccBedrockInferenceProfilesDataSource_basic(t *testing.T) {
	ctx := acctest.Context(t)
	datasourceName := "data.aws_bedrock_inference_profiles.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckPartitionHasService(t, names.BedrockEndpointID) },
		ErrorCheck:               acctest.ErrorCheck(t, names.BedrockServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccInferenceProfilesDataSourceConfig_basic(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckNoResourceAttr(datasourceName, names.AttrType),
					acctest.CheckResourceAttrGreaterThanValue(datasourceName, "inference_profile_summaries.#", 0),
				),
			},
		},
	})
}

func TestAccBedrockInferenceProfilesDataSource_type(t *testing.T) {
	ctx := acctest.Context(t)
	datasourceName := "data.aws_bedrock_inference_profiles.test"
	typeApplication := string(types.InferenceProfileTypeApplication)
	typeSystemDefined := string(types.InferenceProfileTypeSystemDefined)

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckPartitionHasService(t, names.BedrockEndpointID) },
		ErrorCheck:               acctest.ErrorCheck(t, names.BedrockServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccInferenceProfilesDataSourceConfig_type(typeApplication),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(datasourceName, names.AttrType, typeApplication),
					acctest.CheckResourceAttrGreaterThanOrEqualValue(datasourceName, "inference_profile_summaries.#", 0),
				),
			},
			{
				Config: testAccInferenceProfilesDataSourceConfig_type(typeSystemDefined),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(datasourceName, names.AttrType, typeSystemDefined),
					acctest.CheckResourceAttrGreaterThanValue(datasourceName, "inference_profile_summaries.#", 0),
				),
			},
		},
	})
}

func testAccInferenceProfilesDataSourceConfig_basic() string {
	return `
data "aws_bedrock_inference_profiles" "test" {}
`
}

func testAccInferenceProfilesDataSourceConfig_type(inferenceProfileType string) string {
	return fmt.Sprintf(`
data "aws_bedrock_inference_profiles" "test" {
  type = %[1]q
}
`, inferenceProfileType)
}
