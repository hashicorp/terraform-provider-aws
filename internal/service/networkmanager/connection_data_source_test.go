package networkmanager_test

import (
	"testing"

	"github.com/aws/aws-sdk-go/service/networkmanager"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
)

func TestAccNetworkManagerConnectionDataSource_basic(t *testing.T) {
	dataSourceName := "data.aws_networkmanager_connection.test"
	resourceName := "aws_networkmanager_connection.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, networkmanager.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccConnectionDataSourceConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrPair(dataSourceName, "arn", resourceName, "arn"),
					resource.TestCheckResourceAttrPair(dataSourceName, "connected_device_id", resourceName, "connected_device_id"),
					resource.TestCheckResourceAttrPair(dataSourceName, "connected_link_id", resourceName, "connected_link_id"),
					resource.TestCheckResourceAttrPair(dataSourceName, "connection_id", resourceName, "id"),
					resource.TestCheckResourceAttrPair(dataSourceName, "description", resourceName, "description"),
					resource.TestCheckResourceAttrPair(dataSourceName, "global_network_id", resourceName, "global_network_id"),
					resource.TestCheckResourceAttrPair(dataSourceName, "link_id", resourceName, "link_id"),
					resource.TestCheckResourceAttrPair(dataSourceName, "tags.%", resourceName, "tags.%"),
				),
			},
		},
	})
}

func testAccConnectionDataSourceConfig_basic(rName string) string {
	return acctest.ConfigCompose(testAccConnectionConfig_descriptionAndLinks(rName), `
data "aws_networkmanager_connection" "test" {
  global_network_id = aws_networkmanager_global_network.test.id
  connection_id     = aws_networkmanager_connection.test.id
}
`)
}
