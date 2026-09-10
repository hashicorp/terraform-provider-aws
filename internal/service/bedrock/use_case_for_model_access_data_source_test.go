// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package bedrock_test

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccBedrockUseCaseForModelAccessDataSource_basic(t *testing.T) {
	ctx := acctest.Context(t)

	dataSourceName := "data.aws_bedrock_use_case_for_model_access.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.BedrockEndpointID)
			testAccPreCheck(ctx, t)
			testAccPreCheckFoundationModelUseCase(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.BedrockServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccUseCaseForModelAccessDataSourceConfig_basic(),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckUseCaseForModelAccessExists(ctx, t, dataSourceName),
					resource.TestCheckResourceAttrSet(dataSourceName, "form_data"),
				),
			},
		},
	})
}

func testAccUseCaseForModelAccessDataSourceConfig_basic() string {
	return `
data "aws_bedrock_use_case_for_model_access" "test" {
}
`
}
