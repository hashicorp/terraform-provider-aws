package networkmanager_test

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/service/networkmanager"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
)

func TestAccNetworkManagerSiteDataSource_basic(t *testing.T) {
	dataSourceName := "data.aws_networkmanager_site.test"
	resourceName := "aws_networkmanager_site.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, networkmanager.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccSiteDataSourceConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrPair(dataSourceName, "arn", resourceName, "arn"),
					resource.TestCheckResourceAttrPair(dataSourceName, "description", resourceName, "description"),
					resource.TestCheckResourceAttrPair(dataSourceName, "global_network_id", resourceName, "global_network_id"),
					resource.TestCheckResourceAttrPair(dataSourceName, "location.#", resourceName, "location.#"),
					resource.TestCheckResourceAttrPair(dataSourceName, "site_id", resourceName, "id"),
					resource.TestCheckResourceAttrPair(dataSourceName, "tags.%", resourceName, "tags.%"),
				),
			},
		},
	})
}

func testAccSiteDataSourceConfig_basic(rName string) string {
	return fmt.Sprintf(`
resource "aws_networkmanager_global_network" "test" {
  tags = {
    Name = %[1]q
  }
}

resource "aws_networkmanager_site" "test" {
  description       = "test"
  global_network_id = aws_networkmanager_global_network.test.id

  location {
    latitude  = "18.0029784"
    longitude = "-76.7897987"
  }

  tags = {
    Name = %[1]q
  }
}

data "aws_networkmanager_site" "test" {
  global_network_id = aws_networkmanager_global_network.test.id
  site_id           = aws_networkmanager_site.test.id
}
`, rName)
}
