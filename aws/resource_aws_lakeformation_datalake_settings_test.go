package aws

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
)

func TestAccAWSLakeFormationDataLakeSettings_basic(t *testing.T) {
	callerIdentityName := "data.aws_caller_identity.current"
	resourceName := "aws_lakeformation_datalake_settings.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		// TODO: CheckDestroy: testAccCheckAWSLakeFormationDataLakeSettingsDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSLakeFormationDataLakeSettingsConfig_basic,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(callerIdentityName, "account_id", resourceName, "catalog_id"),
					resource.TestCheckResourceAttr(resourceName, "admins.#", "1"),
					resource.TestCheckResourceAttrPair(callerIdentityName, "arn", resourceName, "admins.0"),
				),
			},
		},
	})
}

const testAccAWSLakeFormationDataLakeSettingsConfig_basic = `
data "aws_caller_identity" "current" {}

resource "aws_lakeformation_datalake_settings" "test" {
  admins = ["${data.aws_caller_identity.current.arn}"]
}
`

func TestAccAWSLakeFormationDataLakeSettings_withCatalogId(t *testing.T) {
	callerIdentityName := "data.aws_caller_identity.current"
	resourceName := "aws_lakeformation_datalake_settings.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		// TODO: CheckDestroy: testAccCheckAWSLakeFormationDataLakeSettingsDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSLakeFormationDataLakeSettingsConfig_withCatalogId,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(callerIdentityName, "account_id", resourceName, "catalog_id"),
					resource.TestCheckResourceAttr(resourceName, "admins.#", "1"),
					resource.TestCheckResourceAttrPair(callerIdentityName, "arn", resourceName, "admins.0"),
				),
			},
		},
	})
}

const testAccAWSLakeFormationDataLakeSettingsConfig_withCatalogId = `
data "aws_caller_identity" "current" {}

resource "aws_lakeformation_datalake_settings" "test" {
  catalog_id = "${data.aws_caller_identity.current.account_id}"
  admins = ["${data.aws_caller_identity.current.arn}"]
}
`
