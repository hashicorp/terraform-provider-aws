// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package ec2_test

import (
	"fmt"
	"testing"

	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccEC2EBSVolumesDataSource_basic(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckVolumeDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccEBSVolumesDataSourceConfig_volumeIDs(rName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("data.aws_ebs_volumes.by_tags", "ids.#", acctest.Ct2),
					resource.TestCheckResourceAttr("data.aws_ebs_volumes.by_filter", "ids.#", acctest.Ct1),
					resource.TestCheckResourceAttr("data.aws_ebs_volumes.empty", "ids.#", acctest.Ct0),
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
