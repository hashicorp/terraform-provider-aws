package networkmanager_test

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/service/networkmanager"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
)

func TestAccNetworkManagerLinkDataSource_basic(t *testing.T) {
	dataSourceName := "data.aws_networkmanager_link.test"
	resourceName := "aws_networkmanager_link.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, networkmanager.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccLinkDataSourceConfig(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrPair(dataSourceName, "arn", resourceName, "arn"),
					resource.TestCheckResourceAttrPair(dataSourceName, "bandwidth.#", resourceName, "bandwidth.#"),
					resource.TestCheckResourceAttrPair(dataSourceName, "description", resourceName, "description"),
					resource.TestCheckResourceAttrPair(dataSourceName, "global_network_id", resourceName, "global_network_id"),
					resource.TestCheckResourceAttrPair(dataSourceName, "link_id", resourceName, "id"),
					resource.TestCheckResourceAttrPair(dataSourceName, "provider_name", resourceName, "provider_name"),
					resource.TestCheckResourceAttrPair(dataSourceName, "site_id", resourceName, "site_id"),
					resource.TestCheckResourceAttrPair(dataSourceName, "tags.%", resourceName, "tags.%"),
					resource.TestCheckResourceAttrPair(dataSourceName, "type", resourceName, "type"),
				),
			},
		},
	})
}

func testAccLinkDataSourceConfig(rName string) string {
	return fmt.Sprintf(`
resource "aws_networkmanager_global_network" "test" {
  tags = {
    Name = %[1]q
  }
}

resource "aws_networkmanager_site" "test" {
  global_network_id = aws_networkmanager_global_network.test.id

  tags = {
    Name = %[1]q
  }
}

resource "aws_networkmanager_link" "test" {
  global_network_id = aws_networkmanager_global_network.test.id
  site_id           = aws_networkmanager_site.test.id

  bandwidth {
    download_speed = 50
    upload_speed   = 10
  }

  description   = "description1"
  provider_name = "provider1"
  type          = "type1"

  tags = {
    Name = %[1]q
  }
}

data "aws_networkmanager_link" "test" {
  global_network_id = aws_networkmanager_global_network.test.id
  link_id           = aws_networkmanager_link.test.id
}
`, rName)
}
