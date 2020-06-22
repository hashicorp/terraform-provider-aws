package aws

import (
	"fmt"
	"strings"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/servicecatalog"
	"github.com/hashicorp/terraform-plugin-sdk/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
)

func TestAccAWSServiceCatalogPortfolioProductAssociation_basic(t *testing.T) {
	salt := acctest.RandString(5)
	var portfolioId, productId string
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAwsServiceCatalogPortfolioProductAssociationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSServiceCatalogPortfolioProductAssociationConfig_basic(salt, salt),
				Check:  testAccCheckAwsServiceCatalogPortfolioProductAssociationExists(salt, salt, &portfolioId, &productId),
			},
			{
				ResourceName:      "aws_servicecatalog_portfolio_product_association.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccAWSServiceCatalogPortfolioProductAssociation_disappears(t *testing.T) {
	salt := acctest.RandString(5)
	var portfolioId, productId string
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAwsServiceCatalogPortfolioProductAssociationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSServiceCatalogPortfolioProductAssociationConfig_basic(salt, salt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsServiceCatalogPortfolioProductAssociationExists(salt, salt, &portfolioId, &productId),
					testAccCheckAwsServiceCatalogPortfolioProductAssociationDisappears(),
					func(s *terraform.State) error {
						return testAccCheckAwsServiceCatalogPortfolioProductAssociationNotPresentInAws(&portfolioId, &productId)
					},
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccAWSServiceCatalogPortfolioProductAssociation_Portfolio_update(t *testing.T) {
	salt := acctest.RandString(5)
	salt2 := acctest.RandString(5)
	var portfolioId1, portfolioId2, productId1, productId2 string
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAwsServiceCatalogPortfolioProductAssociationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSServiceCatalogPortfolioProductAssociationConfig_basic(salt, salt),
				Check:  testAccCheckAwsServiceCatalogPortfolioProductAssociationExists(salt, salt, &portfolioId1, &productId1),
			},
			{
				Config: testAccAWSServiceCatalogPortfolioProductAssociationConfig_basic(salt2, salt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsServiceCatalogPortfolioProductAssociationExists(salt2, salt, &portfolioId2, &productId2),
					func(s *terraform.State) error {
						if portfolioId1 == portfolioId2 {
							return fmt.Errorf("Portfolio ID should have changed")
						}
						if productId1 != productId2 {
							return fmt.Errorf("Product ID should not have changed")
						}
						return testAccCheckAwsServiceCatalogPortfolioProductAssociationNotPresentInAws(&portfolioId1, &productId1)
					},
				),
			},
		},
	})
}

func TestAccAWSServiceCatalogPortfolioProductAssociation_Product_update(t *testing.T) {
	salt := acctest.RandString(5)
	salt2 := acctest.RandString(5)
	var portfolioId1, portfolioId2, productId1, productId2 string
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAwsServiceCatalogPortfolioProductAssociationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSServiceCatalogPortfolioProductAssociationConfig_basic(salt, salt),
				Check:  testAccCheckAwsServiceCatalogPortfolioProductAssociationExists(salt, salt, &portfolioId1, &productId1),
			},
			{
				Config: testAccAWSServiceCatalogPortfolioProductAssociationConfig_basic(salt, salt2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsServiceCatalogPortfolioProductAssociationExists(salt, salt2, &portfolioId2, &productId2),
					func(s *terraform.State) error {
						if portfolioId1 != portfolioId2 {
							return fmt.Errorf("Portfolio ID should not have changed")
						}
						if productId1 == productId2 {
							return fmt.Errorf("Product ID should have changed")
						}
						return testAccCheckAwsServiceCatalogPortfolioProductAssociationNotPresentInAws(&portfolioId2, &productId1)
					},
				),
			},
		},
	})
}

func TestAccAWSServiceCatalogPortfolioProductAssociation_update_all(t *testing.T) {
	salt := acctest.RandString(5)
	salt2 := acctest.RandString(5)
	var portfolioId1, portfolioId2, productId1, productId2 string
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAwsServiceCatalogPortfolioProductAssociationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSServiceCatalogPortfolioProductAssociationConfig_basic(salt, salt),
				Check:  testAccCheckAwsServiceCatalogPortfolioProductAssociationExists(salt, salt, &portfolioId1, &productId1),
			},
			{
				Config: testAccAWSServiceCatalogPortfolioProductAssociationConfig_basic(salt2, salt2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsServiceCatalogPortfolioProductAssociationExists(salt2, salt2, &portfolioId2, &productId2),
					func(s *terraform.State) error {
						if portfolioId1 == portfolioId2 {
							return fmt.Errorf("Portfolio ID should have changed")
						}
						if productId1 == productId2 {
							return fmt.Errorf("Product ID should have changed")
						}
						return testAccCheckAwsServiceCatalogPortfolioProductAssociationNotPresentInAws(&portfolioId1, &productId1)
					},
				),
			},
		},
	})
}

func testAccCheckAwsServiceCatalogPortfolioProductAssociationExists(portfolioSalt, productSalt string, portfolioIdToSet, productIdToSet *string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := testAccProvider.Meta().(*AWSClient).scconn
		rs, ok := s.RootModule().Resources["aws_servicecatalog_portfolio_product_association.test"]
		if !ok {
			return fmt.Errorf("association not found")
		}

		_, portfolioId, productId, err := parseServiceCatalogPortfolioProductAssociationResourceId(rs.Primary.ID)
		if err != nil {
			return err
		}

		rsPortfolio, ok := s.RootModule().Resources["aws_servicecatalog_portfolio.test-"+portfolioSalt]
		if !ok {
			return fmt.Errorf("portfolio %s not found", portfolioSalt)
		}
		if !strings.Contains(rsPortfolio.Primary.Attributes["name"], portfolioSalt) {
			return fmt.Errorf("portfolio from association ID %s did not contain expected salt '%s'", rs.Primary.ID, portfolioSalt)
		}

		rsProduct, ok := s.RootModule().Resources["aws_servicecatalog_product.test-"+productSalt]
		if !ok {
			return fmt.Errorf("product %s not found", productSalt)
		}
		if !strings.Contains(rsProduct.Primary.Attributes["name"], productSalt) {
			return fmt.Errorf("product from association ID %s did not contain expected salt '%s'", rs.Primary.ID, productSalt)
		}

		*portfolioIdToSet = portfolioId
		*productIdToSet = productId

		input := servicecatalog.ListPortfoliosForProductInput{ProductId: &productId}
		page, err := conn.ListPortfoliosForProduct(&input)
		if err != nil {
			return err
		}
		for _, portfolioDetail := range page.PortfolioDetails {
			if aws.StringValue(portfolioDetail.Id) == portfolioId {
				// found
				return nil
			}
		}
		return fmt.Errorf("association not found between portfolio %s and product %s; portfolios were: %v", portfolioId, productId, page.PortfolioDetails)
	}
}

func testAccCheckAwsServiceCatalogPortfolioProductAssociationDisappears() resource.TestCheckFunc {
	return func(s *terraform.State) error {
		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_servicecatalog_portfolio_product_association" {
				continue // not our monkey
			}
			_, portfolioId, productId, err := parseServiceCatalogPortfolioProductAssociationResourceId(rs.Primary.ID)
			if err != nil {
				return err
			}
			conn := testAccProvider.Meta().(*AWSClient).scconn
			input := servicecatalog.DisassociateProductFromPortfolioInput{
				PortfolioId: aws.String(portfolioId),
				ProductId:   aws.String(productId),
			}
			_, err = conn.DisassociateProductFromPortfolio(&input)
			if err != nil {
				return fmt.Errorf("deleting Service Catalog Product(%s)/Portfolio(%s) Association failed: %s",
					productId, portfolioId, err.Error())
			}
			return nil
		}
		return fmt.Errorf("no matching resource found to make disappear")
	}
}

func testAccCheckAwsServiceCatalogPortfolioProductAssociationDestroy(s *terraform.State) error {
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_servicecatalog_portfolio_Product_association" {
			continue // not our monkey
		}
		_, portfolioId, productId, err := parseServiceCatalogPortfolioProductAssociationResourceId(rs.Primary.ID)
		if err != nil {
			return err
		}
		return testAccCheckAwsServiceCatalogPortfolioProductAssociationNotPresentInAws(&portfolioId, &productId)
	}
	return nil
}

func testAccCheckAwsServiceCatalogPortfolioProductAssociationNotPresentInAws(portfolioId *string, productId *string) error {
	conn := testAccProvider.Meta().(*AWSClient).scconn
	input := servicecatalog.ListPortfoliosForProductInput{ProductId: productId}
	portfolios, err := conn.ListPortfoliosForProduct(&input)
	if err != nil {
		if isAWSErr(err, servicecatalog.ErrCodeResourceNotFoundException, "") {
			return nil // not found is good
		}
		return err // some other unexpected error
	}
	if len(portfolios.PortfolioDetails) == 0 {
		return nil
	}
	if aws.StringValue(portfolioId) == "*" {
		return fmt.Errorf("expected AWS Service Catalog Portfolio Product Associations to be gone for %s, but still found some: %v", aws.StringValue(productId), portfolios.PortfolioDetails)
	} else {
		for _, portfolioDetail := range portfolios.PortfolioDetails {
			if aws.StringValue(portfolioDetail.Id) == aws.StringValue(portfolioId) {
				return fmt.Errorf("expected AWS Service Catalog Portfolio Product Association to be gone for %s, but it was still found: %s", aws.StringValue(productId), aws.StringValue(portfolioDetail.Id))
			}
		}
		// not found
		return nil
	}
}

func testAccAWSServiceCatalogPortfolioProductAssociationConfig_basic(portfolioSalt, productSalt string) string {
	return composeConfig(
		testAccAWSServiceCatalogPortfolioProductAssociationConfig_portfolio(portfolioSalt),
		testAccAWSServiceCatalogPortfolioProductAssociationConfig_product(productSalt),
		fmt.Sprintf(`
resource "aws_servicecatalog_portfolio_product_association" "test" {
    portfolio_id = aws_servicecatalog_portfolio.test-%s.id
    product_id = aws_servicecatalog_product.test-%s.id
}
`, portfolioSalt, productSalt))
}

func testAccAWSServiceCatalogPortfolioProductAssociationConfig_portfolio(salt string) string {
	// based on testAccAWSServiceCatalogPortfolioConfig_basic
	return fmt.Sprintf(`
resource "aws_servicecatalog_portfolio" "test-%s" {
  name          = "%s"
  description   = "test-2"
  provider_name = "test-3"
}
`, salt, "tfm-test-"+salt)
}

func testAccAWSServiceCatalogPortfolioProductAssociationConfig_product(salt string) string {
	// based on testAccAWSServiceCatalogProductResourceConfig_basic
	resourceName := "aws_servicecatalog_product.test-" + salt

	thisResourceParts := strings.Split(resourceName, ".")
	return fmt.Sprintf(`
data "aws_region" "current" {}

resource "aws_s3_bucket" "bucket" {
  bucket        = "bucket-%[3]s"
  region        = data.aws_region.current.name
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

resource "%[1]s" "%[2]s" {
  description         = "arbitrary product description"
  distributor         = "arbitrary distributor"
  name                = "product-%[3]s"
  owner               = "arbitrary owner"
  product_type        = "CLOUD_FORMATION_TEMPLATE"
  support_description = "arbitrary support description"
  support_email       = "arbitrary@email.com"
  support_url         = "http://arbitrary_url/foo.html"

  provisioning_artifact {
    description = "arbitrary description"
    name        = "pa-%[3]s"
    info = {
      LoadTemplateFromURL = "https://s3.amazonaws.com/${aws_s3_bucket.bucket.id}/${aws_s3_bucket_object.template1.key}"
    }
  }

}
`, thisResourceParts[0], thisResourceParts[1], salt)
}
