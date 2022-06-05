package networkmanager_test

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/service/networkmanager"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
)

func TestAccNetworkManagerGlobalNetworksDataSource_basic(t *testing.T) {
	dataSourceAllName := "data.aws_networkmanager_global_networks.all"
	dataSourceByTagsName := "data.aws_networkmanager_global_networks.by_tags"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, networkmanager.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccGlobalNetworksDataSourceConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					acctest.CheckResourceAttrGreaterThanValue(dataSourceAllName, "ids.#", "1"),
					resource.TestCheckResourceAttr(dataSourceByTagsName, "ids.#", "1"),
				),
			},
		},
	})
}

func testAccGlobalNetworksDataSourceConfig(rName string) string {
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
