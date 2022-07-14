package location_test

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/service/locationservice"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
)

func TestAccLocationPlaceIndexDataSource_indexName(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	dataSourceName := "data.aws_location_place_index.test"
	resourceName := "aws_location_place_index.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, locationservice.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckPlaceIndexDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccPlaceIndexDataSourceConfig_indexName(rName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(dataSourceName, "create_time", resourceName, "create_time"),
					resource.TestCheckResourceAttrPair(dataSourceName, "data_source", resourceName, "data_source"),
					resource.TestCheckResourceAttrPair(dataSourceName, "data_source_configuration.#", resourceName, "data_source_configuration.#"),
					resource.TestCheckResourceAttrPair(dataSourceName, "data_source_configuration.0.intended_use", resourceName, "data_source_configuration.0.intended_use"),
					resource.TestCheckResourceAttrPair(dataSourceName, "description", resourceName, "description"),
					resource.TestCheckResourceAttrPair(dataSourceName, "index_arn", resourceName, "index_arn"),
					resource.TestCheckResourceAttrPair(dataSourceName, "index_name", resourceName, "index_name"),
					resource.TestCheckResourceAttrPair(dataSourceName, "tags.%", resourceName, "tags.%"),
					resource.TestCheckResourceAttrPair(dataSourceName, "update_time", resourceName, "update_time"),
				),
			},
		},
	})
}

func testAccPlaceIndexDataSourceConfig_indexName(rName string) string {
	return fmt.Sprintf(`
resource "aws_location_place_index" "test" {
  data_source = "Here"
  index_name  = %[1]q
}

data "aws_location_place_index" "test" {
  index_name = aws_location_place_index.test.index_name
}
`, rName)
}
