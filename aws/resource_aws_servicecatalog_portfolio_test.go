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
				Config: testAccCheckAwsServiceCatalogPortfolioResourceConfig_basic1,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("aws_servicecatalog_portfolio.test", "name", "test-1"),
					resource.TestCheckResourceAttr("aws_servicecatalog_portfolio.test", "description", "test-2"),
					resource.TestCheckResourceAttr("aws_servicecatalog_portfolio.test", "provider_name", "test-3"),
				),
			},
			resource.TestStep{
				Config: testAccCheckAwsServiceCatalogPortfolioResourceConfig_basic2,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("aws_servicecatalog_portfolio.test", "name", "test-a"),
					resource.TestCheckResourceAttr("aws_servicecatalog_portfolio.test", "description", "test-b"),
					resource.TestCheckResourceAttr("aws_servicecatalog_portfolio.test", "provider_name", "test-c"),
				),
			},
			resource.TestStep{
				Config: testAccCheckAwsServiceCatalogPortfolioResourceConfig_basic3,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("aws_servicecatalog_portfolio.test", "name", "test-a"),
					resource.TestCheckResourceAttr("aws_servicecatalog_portfolio.test", "description", "test-only-change-me"),
					resource.TestCheckResourceAttr("aws_servicecatalog_portfolio.test", "provider_name", "test-c"),
				),
			},
		},
	})
}

const testAccCheckAwsServiceCatalogPortfolioResourceConfig_basic1 = `
resource "aws_servicecatalog_portfolio" "test" {
  name = "test-1"
  description = "test-2"
  provider_name = "test-3"
}
`

const testAccCheckAwsServiceCatalogPortfolioResourceConfig_basic2 = `
resource "aws_servicecatalog_portfolio" "test" {
  name = "test-a"
  description = "test-b"
  provider_name = "test-c"
}
`

const testAccCheckAwsServiceCatalogPortfolioResourceConfig_basic3 = `
resource "aws_servicecatalog_portfolio" "test" {
  name = "test-a"
  description = "test-only-change-me"
  provider_name = "test-c"
}
`
