package aws

import (
	"testing"

	"github.com/aws/aws-sdk-go/service/servicecatalog"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

// THIS IS A HARD CODED TEST TO TEST FUNCTIONALITY WHILE WORKING THROUGH BLOCKED CODE

// !!FIX THIS!! BEFORE MERGING

func TestAccAWSServiceCatalogLaunchPathsDataSource_basic(t *testing.T) {
	dataSourceName := "data.aws_servicecatalog_launch_paths.test"
	rName := acctest.RandomWithPrefix("tf-acc-test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:   func() { testAccPreCheck(t) },
		ErrorCheck: testAccErrorCheck(t, servicecatalog.EndpointsID),
		Providers:  testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSServiceCatalogLaunchPathsDataSourceConfig_basic(rName, rName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(dataSourceName, "accept_language", "en"),
					resource.TestCheckResourceAttr(dataSourceName, "product_id", "prod-nui42zvkjm52a"),
					resource.TestCheckResourceAttr(dataSourceName, "summaries.#", "1"),
					resource.TestCheckResourceAttr(dataSourceName, "summaries.0.path_id", "lpv2-5aftplaxgosvk"),
					resource.TestCheckResourceAttr(dataSourceName, "summaries.0.name", "satsuki-11221122"),
				),
			},
		},
	})
}

func testAccAWSServiceCatalogLaunchPathsDataSourceConfig_basic(rName, description string) string {
	return `
data "aws_servicecatalog_launch_paths" "test" {
  product_id = "prod-nui42zvkjm52a"
}
`
}
