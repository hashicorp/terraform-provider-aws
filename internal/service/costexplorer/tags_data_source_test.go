package costexplorer_test

import (
	"fmt"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go/service/costexplorer"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
)

func TestAccCostExplorerTagsDataSource_basic(t *testing.T) {
	var output costexplorer.CostCategory
	resourceName := "aws_costexplorer_cost_category.test"
	dataSourceName := "data.aws_costexplorer_tags.test"
	rName := sdkacctest.RandomWithPrefix("tf-acc-test")

	formatDate := "2006-01-02"
	currentTime := time.Now()
	monthsAgo := currentTime.AddDate(0, -10, 0)
	startDate := monthsAgo.Format(formatDate)
	endDate := currentTime.Format(formatDate)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ProviderFactories: acctest.ProviderFactories,
		ErrorCheck:        acctest.ErrorCheck(t, costexplorer.EndpointsID),
		Steps: []resource.TestStep{
			{
				Config: testAccCostExplorerTagsSourceConfig(rName, startDate, endDate),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCostExplorerCostCategoryExists(resourceName, &output),
					resource.TestCheckResourceAttr(dataSourceName, "tags.#", "4"),
				),
			},
		},
	})
}

func TestAccCostExplorerTagsDataSource_filter(t *testing.T) {
	var output costexplorer.CostCategory
	resourceName := "aws_costexplorer_cost_category.test"
	dataSourceName := "data.aws_costexplorer_tags.test"
	rName := sdkacctest.RandomWithPrefix("tf-acc-test")

	formatDate := "2006-01-02"
	currentTime := time.Now()
	monthsAgo := currentTime.AddDate(0, -10, 0)
	startDate := monthsAgo.Format(formatDate)
	endDate := currentTime.Format(formatDate)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ProviderFactories: acctest.ProviderFactories,
		ErrorCheck:        acctest.ErrorCheck(t, costexplorer.EndpointsID),
		Steps: []resource.TestStep{
			{
				Config: testAccCostExplorerTagsSourceFilterConfig(rName, startDate, endDate),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCostExplorerCostCategoryExists(resourceName, &output),
					resource.TestCheckResourceAttr(dataSourceName, "tags.#", "3"),
				),
			},
		},
	})
}

func testAccCostExplorerTagsSourceConfig(name, start, end string) string {
	return fmt.Sprintf(testAccCostExplorerCostCategoryConfig(name)+`
data "aws_costexplorer_tags" "test" {
  time_period {
    start = %[1]q
    end   = %[2]q
  }
}
`, start, end)
}

func testAccCostExplorerTagsSourceFilterConfig(name, start, end string) string {
	return fmt.Sprintf(testAccCostExplorerCostCategoryConfig(name)+`
data "aws_region" "current" {}

data "aws_costexplorer_tags" "test" {
  time_period {
    start = %[1]q
    end   = %[2]q
  }
  filter {
    dimension {
      key           = "REGION"
      values        = [data.aws_region.current.name]
      match_options = ["EQUALS"]
    }
  }
}
`, start, end)
}
