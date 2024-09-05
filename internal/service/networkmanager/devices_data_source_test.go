// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package networkmanager_test

import (
	"fmt"
	"testing"

	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccNetworkManagerDevicesDataSource_basic(t *testing.T) {
	ctx := acctest.Context(t)
	dataSourceAllName := "data.aws_networkmanager_devices.all"
	dataSourceByTagsName := "data.aws_networkmanager_devices.by_tags"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.NetworkManagerServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccDevicesDataSourceConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					acctest.CheckResourceAttrGreaterThanValue(dataSourceAllName, "ids.#", 1),
					resource.TestCheckResourceAttr(dataSourceByTagsName, "ids.#", acctest.Ct1),
				),
			},
		},
	})
}

func testAccDevicesDataSourceConfig_basic(rName string) string {
	return fmt.Sprintf(`
resource "aws_networkmanager_global_network" "test" {
  tags = {
    Name = %[1]q
  }
}

resource "aws_networkmanager_device" "test1" {
  global_network_id = aws_networkmanager_global_network.test.id

  tags = {
    Name = %[1]q
  }
}

resource "aws_networkmanager_device" "test2" {
  global_network_id = aws_networkmanager_global_network.test.id
}

data "aws_networkmanager_devices" "all" {
  global_network_id = aws_networkmanager_global_network.test.id

  depends_on = [aws_networkmanager_device.test1, aws_networkmanager_device.test2]
}

data "aws_networkmanager_devices" "by_tags" {
  global_network_id = aws_networkmanager_global_network.test.id

  tags = {
    Name = %[1]q
  }

  depends_on = [aws_networkmanager_device.test1, aws_networkmanager_device.test2]
}
`, rName)
}
