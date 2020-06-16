package aws

import (
	"fmt"
	"github.com/aws/aws-sdk-go/service/servicecatalog"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
	"log"
	"strings"
	"testing"
)

func TestAccAWSServiceCatalogPortfolioProductAssociation_Basic(t *testing.T) {
	productId := resource.UniqueId()
	portfolioName := resource.UniqueId()[16:]
	log.Printf("length: %d", len(portfolioName))
	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		CheckDestroy: testAccCheckServiceCatalogPortfolioProductAssociationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckAwsServiceCatalogPortfolioProductAssociationConfigBasic(productId, portfolioName),
				Check: testAccCheckAssociation(),
			},
			//{
			//	ResourceName: "association",
			//	ImportState: true,
			//	ImportStateVerify: true,
			//},
		},
	})
}

func testAccCheckAssociation() resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := testAccProvider.Meta().(*AWSClient).scconn
		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_servicecatalog_portfolio_product_association" {
				continue // not our monkey
			}
			productId, portfolioId := parseServiceCatalogPortfolioProductAssociationResourceId(rs.Primary.ID)
			input := servicecatalog.ListPortfoliosForProductInput{ProductId: &productId}
			portfolios, err := conn.ListPortfoliosForProduct(&input)
			if err != nil {
				return err
			}
			for _, portfolioDetail := range portfolios.PortfolioDetails {
				if *portfolioDetail.Id == portfolioId {
					return nil //is good
				}
			}
			return fmt.Errorf("association not found between portfolio %s and product %s", portfolioId, productId)
		}
		return fmt.Errorf("no associations found")
	}
}

func testAccCheckAwsServiceCatalogPortfolioProductAssociationConfigBasic(productId string, portfolioName string) string {
	//TODO: do I need to create both the portfolio and product first?
	template := fmt.Sprintf(`
data "aws_region" "current" { }

resource "aws_s3_bucket" "bucket" {
  bucket        = "%s"
  region        = "${data.aws_region.current.name}"
  acl           = "private"
  force_destroy = true
}

resource "aws_s3_bucket_object" "template1" {
  bucket  = "${aws_s3_bucket.bucket.id}"
  key     = "test_templates_for_terraform_sc_dev1.json"
  content = <<EOF
{
  "AWSTemplateFormatVersion": "2010-09-09",
  "Description": "Test CF teamplate for Service Catalog terraform dev",
  "Resources": {
    "Empty": {
      "Type": "AWS::CloudFormation::WaitConditionHandle"
    }
  }
}
EOF
}
resource "aws_servicecatalog_portfolio" "portfolio" {
  name = "%s"
  provider_name = "testing"
}
resource "aws_servicecatalog_product" "product" {
  name                = "product"
  owner               = "testing"
  product_type        = "CLOUD_FORMATION_TEMPLATE"
  provisioning_artifact {
    description = "description"
    name        = "artifact-name"
    info = {
      LoadTemplateFromURL = "https://s3.amazonaws.com/${aws_s3_bucket.bucket.id}/${aws_s3_bucket_object.template1.key}"
    }
  }
}
resource "aws_servicecatalog_portfolio_product_association" "association" {
  portfolio_id = aws_servicecatalog_portfolio.portfolio.id
  product_id = aws_servicecatalog_product.product.id
}
`, resource.UniqueId(), portfolioName)
	log.Print(template)
	return template
}

func testAccCheckServiceCatalogPortfolioProductAssociationDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*AWSClient).scconn
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_servicecatalog_portfolio_product_association" {
			continue // not our monkey
		}
		productId, portfolioId := parseServiceCatalogPortfolioProductAssociationResourceId(rs.Primary.ID)
		input := servicecatalog.ListPortfoliosForProductInput{ProductId: &productId}
		portfolios, err := conn.ListPortfoliosForProduct(&input)
		if err != nil {
			if isAWSErr(err, servicecatalog.ErrCodeResourceNotFoundException, "") {
				return nil // not found for product is good
			}
			return err // some other unexpected error
		}
		for _, portfolioDetail := range portfolios.PortfolioDetails {
			if *portfolioDetail.Id == portfolioId {
				return fmt.Errorf("expected AWs Service Catalog Portfolio Product Association to be gone, but it was still found")
			}
		}
	}
	return nil
}

func parseServiceCatalogPortfolioProductAssociationResourceId(id string) (string, string) {
	s := strings.Split(id, "-")
	productId := s[0]
	portfolioId := s[1]
	return productId, portfolioId
}
