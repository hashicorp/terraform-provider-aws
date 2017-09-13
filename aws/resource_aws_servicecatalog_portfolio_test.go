package aws

import (
	"github.com/hashicorp/terraform/helper/resource"

	"testing"
)

func TestAccAWSServiceCatalogPortfolio_basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			resource.TestStep{
				Config: testAccCheckAwsServiceCatalogPortfolioResourceConfig_basic,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("aws_servicecatalog_portfolio.test", "name", "test-1"),
					resource.TestCheckResourceAttr("aws_servicecatalog_portfolio.test", "description", "test-2"),
					resource.TestCheckResourceAttr("aws_servicecatalog_portfolio.test", "provider_name", "test-3"),
				),
			},
		},
	})
}

const testAccCheckAwsServiceCatalogPortfolioResourceConfig_basic = `
resource "aws_servicecatalog_portfolio" "test" {
  name = "test-1"
  description = "test-2"
  provider_name = "test-3"
}
`
