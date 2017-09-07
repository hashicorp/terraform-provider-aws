package aws

import (
	"testing"

	"github.com/hashicorp/terraform/helper/resource"
)

func TestAccAWSRedshiftServiceAccount_basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			resource.TestStep{
				Config: testAccCheckAwsRedshiftServiceAccountConfig,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("data.aws_redshift_service_account.main", "id", "902366379725"),
				),
			},
			resource.TestStep{
				Config: testAccCheckAwsRedshiftServiceAccountExplicitRegionConfig,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("data.aws_redshift_service_account.regional", "id", "307160386991"),
				),
			},
			resource.TestStep{
				Config: testAccCheckAwsRedshiftServiceAccountExplicitRegionConfig,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("data.aws_redshift_service_account.regional", "id", "075028567923"),
				),
			},
		},
	})
}

const testAccCheckAwsRedshiftServiceAccountConfig = `
data "aws_redshift_service_account" "main" { }
`

const testAccCheckAwsRedshiftServiceAccountExplicitRegionConfig = `
data "aws_redshift_service_account" "regional" {
	region = "eu-west-2"
}
`

const testAccCheckAwsRedshiftServiceAccountExplicitRegionConfig = `
data "aws_redshift_service_account" "regional" {
	region = "sa-east-1"
}
`
