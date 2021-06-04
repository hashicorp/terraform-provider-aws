package aws

import (
	"testing"

	"github.com/aws/aws-sdk-go/service/lakeformation"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/terraform-providers/terraform-provider-aws/atest"
)

func testAccAWSLakeFormationDataLakeSettingsDataSource_basic(t *testing.T) {
	callerIdentityName := "data.aws_caller_identity.current"
	resourceName := "data.aws_lakeformation_data_lake_settings.test"

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { atest.PreCheck(t); atest.PreCheckPartitionService(lakeformation.EndpointsID, t) },
		ErrorCheck:   atest.ErrorCheck(t, lakeformation.EndpointsID),
		Providers:    atest.Providers,
		CheckDestroy: testAccCheckAWSLakeFormationDataLakeSettingsDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSLakeFormationDataLakeSettingsDataSourceConfig_basic,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(resourceName, "catalog_id", callerIdentityName, "account_id"),
					resource.TestCheckResourceAttr(resourceName, "admins.#", "1"),
					resource.TestCheckResourceAttrPair(resourceName, "admins.0", callerIdentityName, "arn"),
				),
			},
		},
	})
}

const testAccAWSLakeFormationDataLakeSettingsDataSourceConfig_basic = `
data "aws_caller_identity" "current" {}

resource "aws_lakeformation_data_lake_settings" "test" {
  catalog_id = data.aws_caller_identity.current.account_id
  admins     = [data.aws_caller_identity.current.arn]
}

data "aws_lakeformation_data_lake_settings" "test" {
  catalog_id = aws_lakeformation_data_lake_settings.test.catalog_id
}
`
