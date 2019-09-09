package aws

import (
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/servicecatalog"

	"github.com/hashicorp/terraform/helper/acctest"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"

	"testing"
)

func TestAccAWSServiceCatalogPortfolio_Basic(t *testing.T) {
	resourceName := "aws_servicecatalog_portfolio.test"
	name := acctest.RandString(5)
	var dpo servicecatalog.DescribePortfolioOutput

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckServiceCatlaogPortfolioDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckAwsServiceCatalogPortfolioResourceConfigBasic1(name),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPortfolio(resourceName, &dpo),
					resource.TestCheckResourceAttrSet(resourceName, "arn"),
					resource.TestCheckResourceAttrSet(resourceName, "created_time"),
					resource.TestCheckResourceAttr(resourceName, "name", name),
					resource.TestCheckResourceAttr(resourceName, "description", "test-2"),
					resource.TestCheckResourceAttr(resourceName, "provider_name", "test-3"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.Key1", "Value One"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccCheckAwsServiceCatalogPortfolioResourceConfigBasic2(name),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPortfolio(resourceName, &dpo),
					resource.TestCheckResourceAttrSet(resourceName, "arn"),
					resource.TestCheckResourceAttrSet(resourceName, "created_time"),
					resource.TestCheckResourceAttr(resourceName, "name", name),
					resource.TestCheckResourceAttr(resourceName, "description", "test-b"),
					resource.TestCheckResourceAttr(resourceName, "provider_name", "test-c"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.Key1", "Value 1"),
					resource.TestCheckResourceAttr(resourceName, "tags.Key2", "Value Two"),
				),
			},
			{
				Config: testAccCheckAwsServiceCatalogPortfolioResourceConfigBasic3(name),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPortfolio(resourceName, &dpo),
					resource.TestCheckResourceAttrSet(resourceName, "arn"),
					resource.TestCheckResourceAttrSet(resourceName, "created_time"),
					resource.TestCheckResourceAttr(resourceName, "name", name),
					resource.TestCheckResourceAttr(resourceName, "description", "test-only-change-me"),
					resource.TestCheckResourceAttr(resourceName, "provider_name", "test-c"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.Key3", "Value Three"),
				),
			},
		},
	})
}

func TestAccAWSServiceCatalogPortfolio_Disappears(t *testing.T) {
	name := acctest.RandString(5)
	resourceName := "aws_servicecatalog_portfolio.test"
	var dpo servicecatalog.DescribePortfolioOutput

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckServiceCatlaogPortfolioDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckAwsServiceCatalogPortfolioResourceConfigBasic1(name),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPortfolio(resourceName, &dpo),
					testAccCheckServiceCatlaogPortfolioDisappears(&dpo),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckPortfolio(pr string, dpo *servicecatalog.DescribePortfolioOutput) resource.TestCheckFunc {
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

		*dpo = *resp
		return nil
	}
}

func testAccCheckServiceCatlaogPortfolioDisappears(dpo *servicecatalog.DescribePortfolioOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := testAccProvider.Meta().(*AWSClient).scconn

		input := servicecatalog.DeletePortfolioInput{}
		input.Id = dpo.PortfolioDetail.Id

		_, err := conn.DeletePortfolio(&input)
		return err
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

func testAccCheckAwsServiceCatalogPortfolioResourceConfigBasic1(name string) string {
	return fmt.Sprintf(`
resource "aws_servicecatalog_portfolio" "test" {
  name          = "%s"
  description   = "test-2"
  provider_name = "test-3"

  tags = {
    Key1 = "Value One"
  }
}
`, name)
}

func testAccCheckAwsServiceCatalogPortfolioResourceConfigBasic2(name string) string {
	return fmt.Sprintf(`
resource "aws_servicecatalog_portfolio" "test" {
  name          = "%s"
  description   = "test-b"
  provider_name = "test-c"

  tags = {
    Key1 = "Value 1"
    Key2 = "Value Two"
  }
}
`, name)
}

func testAccCheckAwsServiceCatalogPortfolioResourceConfigBasic3(name string) string {
	return fmt.Sprintf(`
resource "aws_servicecatalog_portfolio" "test" {
  name          = "%s"
  description   = "test-only-change-me"
  provider_name = "test-c"

  tags = {
    Key3 = "Value Three"
  }
}
`, name)
}
