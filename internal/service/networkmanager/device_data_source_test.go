package networkmanager_test

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/service/networkmanager"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
)

func TestAccNetworkManagerDeviceDataSource_basic(t *testing.T) {
	dataSourceName := "data.aws_networkmanager_device.test"
	resourceName := "aws_networkmanager_device.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, networkmanager.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccDeviceDataSourceConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrPair(dataSourceName, "arn", resourceName, "arn"),
					resource.TestCheckResourceAttrPair(dataSourceName, "aws_location.#", resourceName, "aws_location.#"),
					resource.TestCheckResourceAttrPair(dataSourceName, "description", resourceName, "description"),
					resource.TestCheckResourceAttrPair(dataSourceName, "device_id", resourceName, "id"),
					resource.TestCheckResourceAttrPair(dataSourceName, "global_network_id", resourceName, "global_network_id"),
					resource.TestCheckResourceAttrPair(dataSourceName, "location.#", resourceName, "location.#"),
					resource.TestCheckResourceAttrPair(dataSourceName, "model", resourceName, "model"),
					resource.TestCheckResourceAttrPair(dataSourceName, "site_id", resourceName, "site_id"),
					resource.TestCheckResourceAttrPair(dataSourceName, "tags.%", resourceName, "tags.%"),
					resource.TestCheckResourceAttrPair(dataSourceName, "type", resourceName, "type"),
					resource.TestCheckResourceAttrPair(dataSourceName, "vendor", resourceName, "vendor"),
				),
			},
		},
	})
}

func testAccDeviceDataSourceConfig_basic(rName string) string {
	return fmt.Sprintf(`
resource "aws_networkmanager_global_network" "test" {
  tags = {
    Name = %[1]q
  }
}

resource "aws_networkmanager_device" "test" {
  global_network_id = aws_networkmanager_global_network.test.id

  description   = "description1"
  model         = "model1"
  serial_number = "sn1"
  type          = "type1"
  vendor        = "vendor1"

  location {
    address   = "Address 1"
    latitude  = "1.1"
    longitude = "-1.1"
  }

  tags = {
    Name = %[1]q
  }
}

data "aws_networkmanager_device" "test" {
  global_network_id = aws_networkmanager_global_network.test.id
  device_id         = aws_networkmanager_device.test.id
}
`, rName)
}
