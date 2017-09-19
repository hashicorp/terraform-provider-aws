package aws

import (
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/servicecatalog"

	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"

	"testing"
)

func TestAccAWSServiceCatalogPortfolio_basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckServiceCatlaogPortfolioDestroy,
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

func TestAccAWSServiceCatalogPortfolio_disappears(t *testing.T) {
	var portfolioDetail servicecatalog.PortfolioDetail
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckServiceCatlaogPortfolioDestroy,
		Steps: []resource.TestStep{
			resource.TestStep{
				Config: testAccCheckAwsServiceCatalogPortfolioResourceConfig_basic1,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPortfolio("aws_servicecatalog_portfolio.test", &portfolioDetail),
					testAccCheckServiceCatlaogPortfolioDisappears(&portfolioDetail),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckPortfolio(pr string, pd *servicecatalog.PortfolioDetail) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := testAccProvider.Meta().(*AWSClient).scconn
		rs, ok := s.RootModule().Resources[pr]
		if !ok {
			return fmt.Errorf("Not found: %s", pr)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No ID is set")
		}

		input := servicecatalog.DescribePortfolioInput{}
		input.Id = aws.String(rs.Primary.ID)

		resp, err := conn.DescribePortfolio(&input)
		if err != nil {
			return err
		}

		*pd = *resp.PortfolioDetail
		return nil
	}
}

func testAccCheckServiceCatlaogPortfolioDisappears(pd *servicecatalog.PortfolioDetail) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := testAccProvider.Meta().(*AWSClient).scconn

		input := servicecatalog.DeletePortfolioInput{}
		input.Id = pd.Id

		_, err := conn.DeletePortfolio(&input)
		if err != nil {
			return err
		}

		return nil
	}
}

func testAccCheckServiceCatlaogPortfolioDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*AWSClient).scconn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_servicecatalog_portfolio" {
			continue
		}
		input := servicecatalog.DescribePortfolioInput{}
		input.Id = aws.String(rs.Primary.ID)

		_, err := conn.DescribePortfolio(&input)
		if err == nil {
			return fmt.Errorf("Portfolio still exists")
		}
	}

	return nil
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
