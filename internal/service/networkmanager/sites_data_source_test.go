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

func TestAccNetworkManagerSitesDataSource_basic(t *testing.T) {
	ctx := acctest.Context(t)
	dataSourceAllName := "data.aws_networkmanager_sites.all"
	dataSourceByTagsName := "data.aws_networkmanager_sites.by_tags"
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.NetworkManagerServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccSitesDataSourceConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					acctest.CheckResourceAttrGreaterThanValue(dataSourceAllName, "ids.#", 1),
					resource.TestCheckResourceAttr(dataSourceByTagsName, "ids.#", "1"),
				),
			},
		},
	})
}

func testAccSitesDataSourceConfig_basic(rName string) string {
	return fmt.Sprintf(`
resource "aws_networkmanager_global_network" "test" {
  tags = {
    Name = %[1]q
  }
}

resource "aws_networkmanager_site" "test1" {
  global_network_id = aws_networkmanager_global_network.test.id

  tags = {
    Name = %[1]q
  }
}

resource "aws_networkmanager_site" "test2" {
  global_network_id = aws_networkmanager_global_network.test.id
}

data "aws_networkmanager_sites" "all" {
  global_network_id = aws_networkmanager_global_network.test.id

  depends_on = [aws_networkmanager_site.test1, aws_networkmanager_site.test2]
}

data "aws_networkmanager_sites" "by_tags" {
  global_network_id = aws_networkmanager_global_network.test.id

  tags = {
    Name = %[1]q
  }

  depends_on = [aws_networkmanager_site.test1, aws_networkmanager_site.test2]
}
`, rName)
}
