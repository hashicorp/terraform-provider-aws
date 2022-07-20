package ec2_test

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/service/ec2"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
)

func TestAccEC2EBSVolumesDataSource_basic(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ec2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckVolumeDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccEBSVolumesDataSourceConfig_volumeIDs(rName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("data.aws_ebs_volumes.by_tags", "ids.#", "2"),
					resource.TestCheckResourceAttr("data.aws_ebs_volumes.by_filter", "ids.#", "1"),
					resource.TestCheckResourceAttr("data.aws_ebs_volumes.empty", "ids.#", "0"),
				),
			},
		},
	})
}

func testAccEBSVolumesDataSourceConfig_volumeIDs(rName string) string {
	return acctest.ConfigCompose(acctest.ConfigAvailableAZsNoOptIn(), fmt.Sprintf(`
data "aws_region" "current" {}

resource "aws_ebs_volume" "test" {
  count = 2

  availability_zone = data.aws_availability_zones.available.names[0]
  size              = 1

  tags = {
    Name = %[1]q
  }
}

data "aws_ebs_volumes" "by_tags" {
  tags = {
    Name = %[1]q
  }

  depends_on = [aws_ebs_volume.test[0], aws_ebs_volume.test[1]]
}

data "aws_ebs_volumes" "by_filter" {
  filter {
    name   = "volume-id"
    values = [aws_ebs_volume.test[0].id]
  }

  depends_on = [aws_ebs_volume.test[0], aws_ebs_volume.test[1]]
}

data "aws_ebs_volumes" "empty" {
  filter {
    name   = "create-time"
    values = ["2000-01-01T00:00:00.000Z"]
  }

  depends_on = [aws_ebs_volume.test[0], aws_ebs_volume.test[1]]
}
`, rName))
}
