package ce_test

import (
	"fmt"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go/service/costexplorer"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
)

func TestAccCETagsDataSource_basic(t *testing.T) {
	var output costexplorer.CostCategory
	resourceName := "aws_ce_cost_category.test"
	dataSourceName := "data.aws_ce_tags.test"
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
				Config: testAccTagsDataSourceConfig_basic(rName, startDate, endDate),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCostCategoryExists(resourceName, &output),
					resource.TestCheckResourceAttr(dataSourceName, "tags.#", "1"),
				),
			},
		},
	})
}

func TestAccCETagsDataSource_filter(t *testing.T) {
	var output costexplorer.CostCategory
	resourceName := "aws_ce_cost_category.test"
	dataSourceName := "data.aws_ce_tags.test"
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
				Config: testAccTagsDataSourceConfig_filter(rName, startDate, endDate),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCostCategoryExists(resourceName, &output),
					resource.TestCheckResourceAttr(dataSourceName, "tags.#", "1"),
				),
			},
		},
	})
}

func testAccTagsDataSourceConfig_basic(rName, start, end string) string {
	return fmt.Sprintf(`
resource "aws_ce_cost_category" "test" {
  name         = %[1]q
  rule_version = "CostCategoryExpression.v1"
  rule {
    value = "production"
    rule {
      tags {
        key    = %[1]q
        values = ["abc", "123"]
      }
    }
    type = "REGULAR"
  }
}

data "aws_ce_tags" "test" {
  time_period {
    start = %[2]q
    end   = %[3]q
  }
  tag_key = %[1]q
}
`, rName, start, end)
}

func testAccTagsDataSourceConfig_filter(rName, start, end string) string {
	return fmt.Sprintf(`
data "aws_region" "current" {}


resource "aws_ce_cost_category" "test" {
  name         = %[1]q
  rule_version = "CostCategoryExpression.v1"
  rule {
    value = "production"
    rule {
      tags {
        key    = %[1]q
        values = ["abc", "123"]
      }
    }
    type = "REGULAR"
  }
}

data "aws_ce_tags" "test" {
  time_period {
    start = %[2]q
    end   = %[3]q
  }
  filter {
    dimension {
      key           = "REGION"
      values        = [data.aws_region.current.name]
      match_options = ["EQUALS"]
    }
  }
  tag_key = %[1]q
}
`, rName, start, end)
}
