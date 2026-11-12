// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package networkmanager_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccNetworkManagerCoreNetworksDataSource_basic(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	dataSourceAllName := "data.aws_networkmanager_core_networks.all"
	dataSourceByTagsName := "data.aws_networkmanager_core_networks.by_tags"
	resourceName := "aws_networkmanager_core_network.test1"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.NetworkManagerServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckCoreNetworkDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccCoreNetworksDataSourceConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					acctest.CheckResourceAttrGreaterThanValue(dataSourceAllName, "ids.#", 1),
					acctest.CheckResourceAttrGreaterThanValue(dataSourceAllName, "core_networks.#", 1),
					resource.TestCheckResourceAttr(dataSourceByTagsName, "ids.#", "1"),
					resource.TestCheckResourceAttr(dataSourceByTagsName, "core_networks.#", "1"),
					resource.TestCheckResourceAttrPair(dataSourceByTagsName, "ids.0", resourceName, names.AttrID),
					resource.TestCheckResourceAttrPair(dataSourceByTagsName, "core_networks.0.arn", resourceName, names.AttrARN),
					resource.TestCheckResourceAttrPair(dataSourceByTagsName, "core_networks.0.core_network_id", resourceName, names.AttrID),
					resource.TestCheckResourceAttrPair(dataSourceByTagsName, "core_networks.0.global_network_id", resourceName, "global_network_id"),
					resource.TestCheckResourceAttrPair(dataSourceByTagsName, "core_networks.0.state", resourceName, names.AttrState),
					resource.TestCheckResourceAttrSet(dataSourceByTagsName, "core_networks.0.owner_account_id"),
				),
			},
		},
	})
}

func testAccCoreNetworksDataSourceConfig_basic(rName string) string {
	return fmt.Sprintf(`
resource "aws_networkmanager_global_network" "test" {}

resource "aws_networkmanager_core_network" "test1" {
  global_network_id = aws_networkmanager_global_network.test.id

  tags = {
    Name = %[1]q
  }
}

resource "aws_networkmanager_core_network" "test2" {
  global_network_id = aws_networkmanager_global_network.test.id
}

data "aws_networkmanager_core_networks" "all" {
  depends_on = [aws_networkmanager_core_network.test1, aws_networkmanager_core_network.test2]
}

data "aws_networkmanager_core_networks" "by_tags" {
  tags = {
    Name = %[1]q
  }

  depends_on = [aws_networkmanager_core_network.test1, aws_networkmanager_core_network.test2]
}
`, rName)
}
