package aws

import (
	"testing"

	"github.com/aws/aws-sdk-go/service/efs"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/terraform-providers/terraform-provider-aws/atest"
)

func TestAccDataSourceAWSEFSAccessPoints_basic(t *testing.T) {
	dataSourceName := "data.aws_efs_access_points.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { atest.PreCheck(t) },
		ErrorCheck:   atest.ErrorCheck(t, efs.EndpointsID),
		Providers:    atest.Providers,
		CheckDestroy: testAccCheckEfsAccessPointDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceAWSEFSAccessPointsConfig(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(dataSourceName, "arns.#", "1"),
					resource.TestCheckResourceAttr(dataSourceName, "ids.#", "1"),
				),
			},
		},
	})
}

func testAccDataSourceAWSEFSAccessPointsConfig() string {
	return `
resource "aws_efs_file_system" "test" {}

resource "aws_efs_access_point" "test" {
  file_system_id = aws_efs_file_system.test.id
}

data "aws_efs_access_points" "test" {
  file_system_id = aws_efs_access_point.test.file_system_id
}
`
}
