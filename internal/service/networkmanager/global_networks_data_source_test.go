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

func TestAccNetworkManagerGlobalNetworksDataSource_basic(t *testing.T) {
	ctx := acctest.Context(t)
	dataSourceAllName := "data.aws_networkmanager_global_networks.all"
	dataSourceByTagsName := "data.aws_networkmanager_global_networks.by_tags"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.NetworkManagerServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccGlobalNetworksDataSourceConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					acctest.CheckResourceAttrGreaterThanValue(dataSourceAllName, "ids.#", 1),
					resource.TestCheckResourceAttr(dataSourceByTagsName, "ids.#", acctest.Ct1),
				),
			},
		},
	})
}

func testAccGlobalNetworksDataSourceConfig_basic(rName string) string {
	return fmt.Sprintf(`
resource "aws_networkmanager_global_network" "test1" {
  description = "test1"

  tags = {
    Name = %[1]q
  }
}

resource "aws_networkmanager_global_network" "test2" {
  description = "test2"
}

data "aws_networkmanager_global_networks" "all" {
  depends_on = [aws_networkmanager_global_network.test1, aws_networkmanager_global_network.test2]
}

data "aws_networkmanager_global_networks" "by_tags" {
  tags = {
    Name = %[1]q
  }

  depends_on = [aws_networkmanager_global_network.test1, aws_networkmanager_global_network.test2]
}
`, rName)
}
