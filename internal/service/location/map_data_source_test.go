package location_test

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/service/locationservice"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
)

func TestAccLocationMapDataSource_mapName(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	dataSourceName := "data.aws_location_map.test"
	resourceName := "aws_location_map.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, locationservice.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckMapDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccMapDataSourceConfig_name(rName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(dataSourceName, "configuration.#", resourceName, "configuration.#"),
					resource.TestCheckResourceAttrPair(dataSourceName, "configuration.0.style", resourceName, "configuration.0.style"),
					resource.TestCheckResourceAttrPair(dataSourceName, "create_time", resourceName, "create_time"),
					resource.TestCheckResourceAttrPair(dataSourceName, "description", resourceName, "description"),
					resource.TestCheckResourceAttrPair(dataSourceName, "map_arn", resourceName, "map_arn"),
					resource.TestCheckResourceAttrPair(dataSourceName, "map_name", resourceName, "map_name"),
					resource.TestCheckResourceAttrPair(dataSourceName, "tags.%", resourceName, "tags.%"),
					resource.TestCheckResourceAttrPair(dataSourceName, "update_time", resourceName, "update_time"),
				),
			},
		},
	})
}

func testAccMapDataSourceConfig_name(rName string) string {
	return fmt.Sprintf(`
resource "aws_location_map" "test" {
  configuration {
    style = "VectorHereBerlin"
  }

  map_name = %[1]q
}

data "aws_location_map" "test" {
  map_name = aws_location_map.test.map_name
}
`, rName)
}
