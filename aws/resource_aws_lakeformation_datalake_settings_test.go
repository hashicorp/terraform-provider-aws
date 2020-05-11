package aws

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/lakeformation"

	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
)

func TestAccAWSLakeFormationDataLakeSettings_basic(t *testing.T) {
	callerIdentityName := "data.aws_caller_identity.current"
	resourceName := "aws_lakeformation_datalake_settings.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSLakeFormationDataLakeSettingsEmpty,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSLakeFormationDataLakeSettingsConfig_basic,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckResourceAttrAccountID(resourceName, "catalog_id"),
					resource.TestCheckResourceAttrPair(callerIdentityName, "account_id", resourceName, "catalog_id"),
					resource.TestCheckResourceAttr(resourceName, "admins.#", "1"),
					testAccCheckAWSLakeFormationDataLakePrincipal(callerIdentityName, "arn", resourceName, "admins"),
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
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSLakeFormationDataLakeSettingsEmpty,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSLakeFormationDataLakeSettingsConfig_withCatalogId,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckResourceAttrAccountID(resourceName, "catalog_id"),
					resource.TestCheckResourceAttrPair(callerIdentityName, "account_id", resourceName, "catalog_id"),
					resource.TestCheckResourceAttr(resourceName, "admins.#", "1"),
					testAccCheckAWSLakeFormationDataLakePrincipal(callerIdentityName, "arn", resourceName, "admins"),
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

func testAccCheckAWSLakeFormationDataLakeSettingsEmpty(s *terraform.State) error {
	conn := testAccProvider.Meta().(*AWSClient).lakeformationconn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_lakeformation_datalake_settings" {
			continue
		}

		input := &lakeformation.GetDataLakeSettingsInput{
			CatalogId: aws.String(testAccProvider.Meta().(*AWSClient).accountid),
		}

		out, err := conn.GetDataLakeSettings(input)
		if err != nil {
			return fmt.Errorf("Error reading DataLakeSettings: %s", err)
		}

		if len(out.DataLakeSettings.DataLakeAdmins) > 0 {
			return fmt.Errorf("Error admins list not empty in DataLakeSettings: %s", out)
		}
	}

	return nil
}

func testAccCheckAWSLakeFormationDataLakePrincipal(nameFirst, keyFirst, nameSecond, keySecond string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		isFirst, err := primaryInstanceState(s, nameFirst)
		if err != nil {
			return err
		}

		valueFirst, okFirst := isFirst.Attributes[keyFirst]
		if !okFirst {
			return fmt.Errorf("%s: Attribute %q not set", nameFirst, keyFirst)
		}

		expandedKey := fmt.Sprintf("%s.%d", keySecond, schema.HashString(valueFirst))
		return resource.TestCheckResourceAttr(nameSecond, expandedKey, valueFirst)(s)
	}
}
