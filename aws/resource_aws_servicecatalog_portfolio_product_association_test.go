package aws

import (
	"fmt"
	"strings"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/servicecatalog"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccAWSServiceCatalogPortfolioProductAssociation_basic(t *testing.T) {
	saltedName := "tf-acc-test-" + acctest.RandString(5) // RandomWithPrefix exceeds max length 20
	var portfolioId, productId string
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAwsServiceCatalogPortfolioProductAssociationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSServiceCatalogPortfolioProductAssociationConfig_basic("", "", saltedName, saltedName),
				Check:  testAccCheckAwsServiceCatalogPortfolioProductAssociationExists("", "", saltedName, saltedName, &portfolioId, &productId),
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
	saltedName := "tf-acc-test-" + acctest.RandString(5) // RandomWithPrefix exceeds max length 20
	resourceName := "aws_servicecatalog_portfolio_product_association.test"
	var portfolioId, productId string
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAwsServiceCatalogPortfolioProductAssociationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSServiceCatalogPortfolioProductAssociationConfig_basic("", "", saltedName, saltedName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsServiceCatalogPortfolioProductAssociationExists("", "", saltedName, saltedName, &portfolioId, &productId),
					testAccCheckResourceDisappears(testAccProvider, resourceAwsServiceCatalogPortfolioProductAssociation(), resourceName),
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
	saltedName := "tf-acc-test-" + acctest.RandString(5)   // RandomWithPrefix exceeds max length 20
	saltedName2 := "tf-acc-test2-" + acctest.RandString(5) // RandomWithPrefix exceeds max length 20
	var portfolioId1, portfolioId2, productId1, productId2 string
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAwsServiceCatalogPortfolioProductAssociationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSServiceCatalogPortfolioProductAssociationConfig_basic("1", "1", saltedName, saltedName),
				Check:  testAccCheckAwsServiceCatalogPortfolioProductAssociationExists("1", "1", saltedName, saltedName, &portfolioId1, &productId1),
			},
			{
				Config: testAccAWSServiceCatalogPortfolioProductAssociationConfig_basic("2", "1", saltedName2, saltedName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsServiceCatalogPortfolioProductAssociationExists("2", "1", saltedName2, saltedName, &portfolioId2, &productId2),
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
	saltedName := "tf-acc-test-" + acctest.RandString(5)   // RandomWithPrefix exceeds max length 20
	saltedName2 := "tf-acc-test2-" + acctest.RandString(5) // RandomWithPrefix exceeds max length 20
	var portfolioId1, portfolioId2, productId1, productId2 string
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAwsServiceCatalogPortfolioProductAssociationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSServiceCatalogPortfolioProductAssociationConfig_basic("1", "1", saltedName, saltedName),
				Check:  testAccCheckAwsServiceCatalogPortfolioProductAssociationExists("1", "1", saltedName, saltedName, &portfolioId1, &productId1),
			},
			{
				Config: testAccAWSServiceCatalogPortfolioProductAssociationConfig_basic("1", "2", saltedName, saltedName2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsServiceCatalogPortfolioProductAssociationExists("1", "2", saltedName, saltedName2, &portfolioId2, &productId2),
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
	saltedName := "tf-acc-test-" + acctest.RandString(5)   // RandomWithPrefix exceeds max length 20
	saltedName2 := "tf-acc-test2-" + acctest.RandString(5) // RandomWithPrefix exceeds max length 20
	var portfolioId1, portfolioId2, productId1, productId2 string
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAwsServiceCatalogPortfolioProductAssociationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSServiceCatalogPortfolioProductAssociationConfig_basic("1", "1", saltedName, saltedName),
				Check:  testAccCheckAwsServiceCatalogPortfolioProductAssociationExists("1", "1", saltedName, saltedName, &portfolioId1, &productId1),
			},
			{
				Config: testAccAWSServiceCatalogPortfolioProductAssociationConfig_basic("2", "2", saltedName2, saltedName2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsServiceCatalogPortfolioProductAssociationExists("2", "2", saltedName2, saltedName2, &portfolioId2, &productId2),
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

func testAccCheckAwsServiceCatalogPortfolioProductAssociationExists(portfolioResourceSuffix, productResourceSuffix, portfolioSaltedName, productSaltedName string, portfolioIdToSet, productIdToSet *string) resource.TestCheckFunc {
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

		rsPortfolio, ok := s.RootModule().Resources["aws_servicecatalog_portfolio.test"+portfolioResourceSuffix]
		if !ok {
			return fmt.Errorf("portfolio %s not found", portfolioSaltedName)
		}
		if !strings.Contains(rsPortfolio.Primary.Attributes["name"], portfolioSaltedName) {
			return fmt.Errorf("portfolio from association ID %s did not contain expected salt '%s'", rs.Primary.ID, portfolioSaltedName)
		}

		rsProduct, ok := s.RootModule().Resources["aws_servicecatalog_product.test"+productResourceSuffix]
		if !ok {
			return fmt.Errorf("product %s not found", productSaltedName)
		}
		if !strings.Contains(rsProduct.Primary.Attributes["name"], productSaltedName) {
			return fmt.Errorf("product from association ID %s did not contain expected salt '%s'", rs.Primary.ID, productSaltedName)
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

func testAccAWSServiceCatalogPortfolioProductAssociationConfig_basic(portfolioResourceSuffix, productResourceSuffix, portfolioSaltedName, productSaltedName string) string {
	return composeConfig(
		testAccAWSServiceCatalogPortfolioProductAssociationConfig_portfolio(portfolioResourceSuffix, portfolioSaltedName),
		testAccAWSServiceCatalogPortfolioProductAssociationConfig_product(productResourceSuffix, productSaltedName),
		fmt.Sprintf(`
resource "aws_servicecatalog_portfolio_product_association" "test" {
    portfolio_id = aws_servicecatalog_portfolio.test%[1]s.id
    product_id = aws_servicecatalog_product.test%[2]s.id
}
`, portfolioResourceSuffix, productResourceSuffix))
}

func testAccAWSServiceCatalogPortfolioProductAssociationConfig_portfolio(suffix, saltedName string) string {
	// based on testAccAWSServiceCatalogPortfolioConfig_basic
	return fmt.Sprintf(`
resource "aws_servicecatalog_portfolio" "test%[2]s" {
  name          = "%[1]s"
  description   = "test-2"
  provider_name = "test-3"
}
`, saltedName, suffix)
}

func testAccAWSServiceCatalogPortfolioProductAssociationConfig_product(suffix, saltedName string) string {
	return fmt.Sprintf(`
# TODO region is recommended but not required, and breaks it with: config is invalid: "region": this field cannot be set
# data "aws_region" "current" {}

resource "aws_s3_bucket" "bucket" {
  bucket        = "%[1]s"
#  region        = data.aws_region.current.name
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

resource "aws_servicecatalog_product" "test%[2]s" {
  description         = "arbitrary product description"
  distributor         = "arbitrary distributor"
  name                = "%[1]s"
  owner               = "arbitrary owner"
  product_type        = "CLOUD_FORMATION_TEMPLATE"
  support_description = "arbitrary support description"
  support_email       = "arbitrary@email.com"
  support_url         = "http://arbitrary_url/foo.html"

  provisioning_artifact {
    description = "arbitrary description"
    name        = "%[1]s"
    info = {
      LoadTemplateFromURL = "https://s3.amazonaws.com/${aws_s3_bucket.bucket.id}/${aws_s3_bucket_object.template1.key}"
    }
  }

}
`, saltedName, suffix)
}
