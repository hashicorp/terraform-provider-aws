// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package bedrock_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccBedrockFoundationModelAgreementOffersDataSource_basic(t *testing.T) {
	ctx := acctest.Context(t)

	dataSourceName := "data.aws_bedrock_foundation_model_agreement_offers.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.BedrockEndpointID)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.BedrockServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccFoundationModelAgreementOffersDataSourceConfig_basic(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(dataSourceName, names.AttrID),
					acctest.CheckResourceAttrGreaterThanValue(dataSourceName, "offers.#", 0),
				),
			},
		},
	})
}

func TestAccBedrockFoundationModelAgreementOffersDataSource_offerType(t *testing.T) {
	ctx := acctest.Context(t)

	dataSourceName := "data.aws_bedrock_foundation_model_agreement_offers.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.BedrockEndpointID)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.BedrockServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccFoundationModelAgreementOffersDataSourceConfig_offerType("ALL"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(dataSourceName, names.AttrID),
					acctest.CheckResourceAttrGreaterThanValue(dataSourceName, "offers.#", 0),
				),
			},
			{
				Config: testAccFoundationModelAgreementOffersDataSourceConfig_offerType("PUBLIC"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(dataSourceName, names.AttrID),
					acctest.CheckResourceAttrGreaterThanValue(dataSourceName, "offers.#", 0),
				),
			},
		},
	})
}

func testAccFoundationModelAgreementOffersDataSourceConfig_basic() string {
	return `
data "aws_bedrock_foundation_models" "test" {}

data "aws_bedrock_foundation_model_agreement_offers" "test" {
  model_id = data.aws_bedrock_foundation_models.test.model_summaries[0].model_id
}
`
}

func testAccFoundationModelAgreementOffersDataSourceConfig_offerType(offerType string) string {
	return fmt.Sprintf(`
data "aws_bedrock_foundation_models" "test" {}

data "aws_bedrock_foundation_model_agreement_offers" "test" {
  model_id   = data.aws_bedrock_foundation_models.test.model_summaries[0].model_id
  offer_type = "%s"
}
`, offerType)
}
