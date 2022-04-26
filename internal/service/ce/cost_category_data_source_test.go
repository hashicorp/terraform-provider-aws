package ce_test

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/service/costexplorer"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
)

func TestAccCECostCategoryDataSource_basic(t *testing.T) {
	var output costexplorer.CostCategory
	resourceName := "aws_ce_cost_category.test"
	dataSourceName := "data.aws_ce_cost_category.test"
	rName := sdkacctest.RandomWithPrefix("tf-acc-test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ProviderFactories: acctest.ProviderFactories,
		ErrorCheck:        acctest.ErrorCheck(t, costexplorer.EndpointsID),
		Steps: []resource.TestStep{
			{
				Config: testAccCECostCategoryDataSourceConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCECostCategoryExists(resourceName, &output),
					resource.TestCheckResourceAttrPair(dataSourceName, "cost_category_arn", resourceName, "arn"),
					resource.TestCheckResourceAttrPair(dataSourceName, "name", resourceName, "name"),
					resource.TestCheckResourceAttrPair(dataSourceName, "rule_version", resourceName, "rule_version"),
					resource.TestCheckResourceAttrPair(dataSourceName, "rule.%", resourceName, "rule.%"),
				),
			},
		},
	})
}

func testAccCECostCategoryDataSourceConfig(name string) string {
	return fmt.Sprintf(testAccCECostCategoryConfig(name) + `
data "aws_ce_cost_category" "test" {
  cost_category_arn = aws_ce_cost_category.test.arn
}
`)
}
