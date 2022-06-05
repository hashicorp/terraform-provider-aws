package lakeformation_test

import (
	"testing"

	"github.com/aws/aws-sdk-go/service/lakeformation"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
)

func testAccDataLakeSettingsDataSource_basic(t *testing.T) {
	resourceName := "data.aws_lakeformation_data_lake_settings.test"

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); acctest.PreCheckPartitionHasService(lakeformation.EndpointsID, t) },
		ErrorCheck:        acctest.ErrorCheck(t, lakeformation.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckDataLakeSettingsDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDataLakeSettingsDataSourceConfig_basic,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(resourceName, "catalog_id", "data.aws_caller_identity.current", "account_id"),
					resource.TestCheckResourceAttr(resourceName, "admins.#", "1"),
					resource.TestCheckResourceAttrPair(resourceName, "admins.0", "data.aws_iam_session_context.current", "issuer_arn"),
				),
			},
		},
	})
}

const testAccDataLakeSettingsDataSourceConfig_basic = `
data "aws_caller_identity" "current" {}

data "aws_iam_session_context" "current" {
  arn = data.aws_caller_identity.current.arn
}

resource "aws_lakeformation_data_lake_settings" "test" {
  catalog_id = data.aws_caller_identity.current.account_id
  admins     = [data.aws_iam_session_context.current.issuer_arn]
}

data "aws_lakeformation_data_lake_settings" "test" {
  catalog_id = aws_lakeformation_data_lake_settings.test.catalog_id
}
`
