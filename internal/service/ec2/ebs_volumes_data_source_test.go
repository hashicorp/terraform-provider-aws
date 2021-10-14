package ec2_test

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/service/ec2"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
)

func TestAccDataSourceAwsEbsVolumes_basic(t *testing.T) {
	rInt := sdkacctest.RandIntRange(0, 256)
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, ec2.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckVolumeDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccEBSVolumeIDsDataSourceConfig(rInt),
			},
			{
				Config: testAccEBSVolumeIDsWithDataSourceDataSourceConfig(rInt),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("data.aws_ebs_volumes.subject_under_test", "ids.#", "2"),
				),
			},
			{
				// Force the destroy to not refresh the data source (leading to an error)
				Config: testAccEBSVolumeIDsDataSourceConfig(rInt),
			},
		},
	})
}

func testAccEBSVolumeIDsWithDataSourceDataSourceConfig(rInt int) string {
	return fmt.Sprintf(`
%s

data "aws_ebs_volumes" "subject_under_test" {
  tags = {
    TestIdentifierSet = "testAccDataSourceAwsEbsVolumes-%d"
  }
}
`, testAccEBSVolumeIDsDataSourceConfig(rInt), rInt)
}

func testAccEBSVolumeIDsDataSourceConfig(rInt int) string {
	return acctest.ConfigAvailableAZsNoOptIn() + fmt.Sprintf(`
data "aws_region" "current" {}

resource "aws_ebs_volume" "volume" {
  count = 2

  availability_zone = data.aws_availability_zones.available.names[0]
  size              = 1

  tags = {
    TestIdentifierSet = "testAccDataSourceAwsEbsVolumes-%d"
  }
}
`, rInt)
}
