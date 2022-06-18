package networkmanager_test

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/service/networkmanager"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
)

func TestAccNetworkManagerLinksDataSource_basic(t *testing.T) {
	dataSourceAllName := "data.aws_networkmanager_links.all"
	dataSourceByTagsName := "data.aws_networkmanager_links.by_tags"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, networkmanager.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccLinksDataSourceConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					acctest.CheckResourceAttrGreaterThanValue(dataSourceAllName, "ids.#", "1"),
					resource.TestCheckResourceAttr(dataSourceByTagsName, "ids.#", "1"),
				),
			},
		},
	})
}

func testAccLinksDataSourceConfig_basic(rName string) string {
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

resource "aws_networkmanager_link" "test1" {
  global_network_id = aws_networkmanager_global_network.test.id
  site_id           = aws_networkmanager_site.test.id

  bandwidth {
    download_speed = 50
    upload_speed   = 10
  }
}

resource "aws_networkmanager_link" "test2" {
  global_network_id = aws_networkmanager_global_network.test.id
  site_id           = aws_networkmanager_site.test.id

  bandwidth {
    download_speed = 75
    upload_speed   = 25
  }

  tags = {
    Name = %[1]q
  }
}

data "aws_networkmanager_links" "all" {
  global_network_id = aws_networkmanager_global_network.test.id

  depends_on = [aws_networkmanager_link.test1, aws_networkmanager_link.test2]
}

data "aws_networkmanager_links" "by_tags" {
  global_network_id = aws_networkmanager_global_network.test.id

  tags = {
    Name = %[1]q
  }

  depends_on = [aws_networkmanager_link.test1, aws_networkmanager_link.test2]
}
`, rName)
}
