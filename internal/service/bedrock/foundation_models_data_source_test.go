// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package bedrock_test

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/bedrock/types"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
)

func TestAccBedrockFoundationModelsDataSource_basic(t *testing.T) {
	ctx := acctest.Context(t)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
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

func TestAccBedrockFoundationModelsDataSource_byCustomizationType(t *testing.T) {
	ctx := acctest.Context(t)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccFoundationModelsDataSourceConfig_byCustomizationType(string(types.ModelCustomizationFineTuning)),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.aws_bedrock_foundation_models.test", "id"),
					acctest.CheckResourceAttrGreaterThanValue("data.aws_bedrock_foundation_models.test", "model_summaries.#", 0),
				),
			},
		},
	})
}

func TestAccBedrockFoundationModelsDataSource_byInferenceType(t *testing.T) {
	ctx := acctest.Context(t)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccFoundationModelsDataSourceConfig_byInferenceType(string(types.InferenceTypeOnDemand)),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.aws_bedrock_foundation_models.test", "id"),
					acctest.CheckResourceAttrGreaterThanValue("data.aws_bedrock_foundation_models.test", "model_summaries.#", 0),
				),
			},
		},
	})
}

func TestAccBedrockFoundationModelsDataSource_byOutputModality(t *testing.T) {
	ctx := acctest.Context(t)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccFoundationModelsDataSourceConfig_byOutputModality(string(types.ModelModalityText)),
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
data "aws_bedrock_foundation_models" "test" {}
`
}

func testAccFoundationModelsDataSourceConfig_byCustomizationType(customizationType string) string {
	return fmt.Sprintf(`
data "aws_bedrock_foundation_models" "test" {
  by_customization_type = %[1]q
}
`, customizationType)
}

func testAccFoundationModelsDataSourceConfig_byInferenceType(inferenceType string) string {
	return fmt.Sprintf(`
data "aws_bedrock_foundation_models" "test" {
  by_inference_type = %[1]q
}
`, inferenceType)
}

func testAccFoundationModelsDataSourceConfig_byOutputModality(outputModality string) string {
	return fmt.Sprintf(`
data "aws_bedrock_foundation_models" "test" {
  by_output_modality = %[1]q
}
`, outputModality)
}
