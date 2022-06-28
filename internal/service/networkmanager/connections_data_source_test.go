package networkmanager_test

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/service/networkmanager"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
)

func TestAccNetworkManagerConnectionsDataSource_basic(t *testing.T) {
	dataSourceAllName := "data.aws_networkmanager_connections.all"
	dataSourceByTagsName := "data.aws_networkmanager_connections.by_tags"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, networkmanager.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccConnectionsDataSourceConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					acctest.CheckResourceAttrGreaterThanValue(dataSourceAllName, "ids.#", "1"),
					resource.TestCheckResourceAttr(dataSourceByTagsName, "ids.#", "1"),
				),
			},
		},
	})
}

func testAccConnectionsDataSourceConfig_basic(rName string) string {
	return fmt.Sprintf(`
resource "aws_networkmanager_global_network" "test" {
  description = %[1]q

  tags = {
    Name = %[1]q
  }
}

resource "aws_networkmanager_site" "test" {
  global_network_id = aws_networkmanager_global_network.test.id

  description = %[1]q

  tags = {
    Name = %[1]q
  }
}

resource "aws_networkmanager_device" "test1" {
  global_network_id = aws_networkmanager_global_network.test.id

  description = "%[1]s-1"

  site_id = aws_networkmanager_site.test.id

  tags = {
    Name = %[1]q
  }
}

resource "aws_networkmanager_device" "test2" {
  global_network_id = aws_networkmanager_global_network.test.id

  description = "%[1]s-2"

  site_id = aws_networkmanager_site.test.id

  tags = {
    Name = %[1]q
  }

  # Create one device at a time.
  depends_on = [aws_networkmanager_device.test1]
}

resource "aws_networkmanager_device" "test3" {
  global_network_id = aws_networkmanager_global_network.test.id

  description = "%[1]s-3"

  site_id = aws_networkmanager_site.test.id

  tags = {
    Name = %[1]q
  }

  # Create one device at a time.
  depends_on = [aws_networkmanager_device.test2]
}

resource "aws_networkmanager_device" "test4" {
  global_network_id = aws_networkmanager_global_network.test.id

  description = "%[1]s-4"

  site_id = aws_networkmanager_site.test.id

  tags = {
    Name = %[1]q
  }

  # Create one device at a time.
  depends_on = [aws_networkmanager_device.test3]
}

resource "aws_networkmanager_connection" "test1" {
  global_network_id   = aws_networkmanager_global_network.test.id
  device_id           = aws_networkmanager_device.test1.id
  connected_device_id = aws_networkmanager_device.test2.id

  description = "%[1]s-1"
}

resource "aws_networkmanager_connection" "test2" {
  global_network_id   = aws_networkmanager_global_network.test.id
  device_id           = aws_networkmanager_device.test3.id
  connected_device_id = aws_networkmanager_device.test4.id

  description = "%[1]s-2"

  tags = {
    Name = %[1]q
  }

  # Create one connection at a time.
  depends_on = [aws_networkmanager_connection.test1]
}

data "aws_networkmanager_connections" "all" {
  global_network_id = aws_networkmanager_global_network.test.id

  depends_on = [aws_networkmanager_connection.test1, aws_networkmanager_connection.test2]
}

data "aws_networkmanager_connections" "by_tags" {
  global_network_id = aws_networkmanager_global_network.test.id

  tags = {
    Name = %[1]q
  }

  depends_on = [aws_networkmanager_connection.test1, aws_networkmanager_connection.test2]
}
`, rName)
}
