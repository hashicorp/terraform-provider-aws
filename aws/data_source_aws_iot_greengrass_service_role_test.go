package aws

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
)

func TestAccAWSIotGreengrassServiceRoleDataSource_basic(t *testing.T) {
	dataSourceName := "data.aws_iot_greengrass_service_role.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSIotGreengrassServiceRoleConfig,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(dataSourceName, "role_arn", "arn_of_some_service_role"),
				),
			},
		},
	})
}

const testAccAWSIotGreengrassServiceRoleConfig = `
resource "aws_iam_role" "greengrass_service_role" {
  name = "greengrass_service_role"
  assume_role_policy = <<EOF
{
"Version": "2012-10-17",
"Statement": [
	{
	"Effect": "Allow",
	"Principal": {
		"Service": "greengrass.amazonaws.com"
	},
	"Action": "sts:AssumeRole"
	}
]
}
EOF
}

data "aws_iot_greengrass_service_role" "test" {}
`
