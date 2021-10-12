package aws

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/service/greengrass"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccAWSIotGreengrassServiceRoleDataSource_basic(t *testing.T) {
	dataSourceName := "data.aws_iot_greengrass_service_role.test"
	resourceName := "aws_iam_role.greengrass_service_role"
	rInt := acctest.RandInt()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                  func() { testAccPreCheck(t) },
		ErrorCheck:                testAccErrorCheck(t, greengrass.EndpointsID),
		Providers:                 testAccProviders,
		PreventPostDestroyRefresh: true,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSIotGreengrassServiceRoleDataSource_ConfigResources(rInt),
			},
			{
				Config: testAccAWSIotGreengrassServiceRoleDataSourceConfig(rInt),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(dataSourceName, "role_arn", resourceName, "arn"),
				),
			},
		},
	})
}

func testAccAWSIotGreengrassServiceRoleDataSource_ConfigResources(rInt int) string {
	return fmt.Sprintf(`
resource "aws_iam_role" "greengrass_service_role" {
  name               = "greengrass_service_role_test_%[1]d"
  assume_role_policy = <<-EOT
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
  EOT
}

resource "aws_iot_greengrass_service_role" "test" {
  role_arn = aws_iam_role.greengrass_service_role.arn
}
`, rInt)
}

func testAccAWSIotGreengrassServiceRoleDataSourceConfig(rInt int) string {
	return fmt.Sprintf(`
%s

data "aws_iot_greengrass_service_role" "test" {}
`, testAccAWSIotGreengrassServiceRoleDataSource_ConfigResources(rInt))
}
