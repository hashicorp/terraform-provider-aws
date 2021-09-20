package aws

import (
	"testing"

	"github.com/aws/aws-sdk-go/service/efs"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/provider"
)

func TestAccDataSourceAWSEFSAccessPoints_basic(t *testing.T) {
	dataSourceName := "data.aws_efs_access_points.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, efs.EndpointsID),
		Providers:    acctest.Providers,
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
