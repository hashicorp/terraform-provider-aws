package aws

import (
	"testing"

	"github.com/hashicorp/terraform/helper/resource"
)

func TestAccAWSCloudTrailServiceAccount_basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckAwsCloudTrailServiceAccountConfig,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("data.aws_cloudtrail_service_account.main", "id", "113285607260"),
					resource.TestCheckResourceAttr("data.aws_cloudtrail_service_account.main", "arn", "arn:aws:iam::113285607260:root"),
				),
			},
			{
				Config: testAccCheckAwsCloudTrailServiceAccountExplicitRegionConfig,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("data.aws_cloudtrail_service_account.regional", "id", "282025262664"),
					resource.TestCheckResourceAttr("data.aws_cloudtrail_service_account.regional", "arn", "arn:aws:iam::282025262664:root"),
				),
			},
		},
	})
}

const testAccCheckAwsCloudTrailServiceAccountConfig = `
data "aws_cloudtrail_service_account" "main" { }
`

const testAccCheckAwsCloudTrailServiceAccountExplicitRegionConfig = `
data "aws_cloudtrail_service_account" "regional" {
	region = "eu-west-2"
}
`
