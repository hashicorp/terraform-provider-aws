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

func TestAccAWSServiceCatalogPortfolioPrincipalAssociation_basic(t *testing.T) {
	salt := acctest.RandString(5)
	var portfolioId string
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAwsServiceCatalogPortfolioPrincipalAssociationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSServiceCatalogPortfolioPrincipalAssociationConfig_basic(salt, salt),
				Check:  testAccCheckAwsServiceCatalogPortfolioPrincipalAssociationExists(salt, salt, &portfolioId),
			},
			{
				ResourceName:      "aws_servicecatalog_portfolio_principal_association.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccAWSServiceCatalogPortfolioPrincipalAssociation_disappears(t *testing.T) {
	salt := acctest.RandString(5)
	var portfolioId string
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAwsServiceCatalogPortfolioPrincipalAssociationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSServiceCatalogPortfolioPrincipalAssociationConfig_basic(salt, salt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsServiceCatalogPortfolioPrincipalAssociationExists(salt, salt, &portfolioId),
					testAccCheckAwsServiceCatalogPortfolioPrincipalAssociationDisappears(),
					func(s *terraform.State) error {
						return testAccCheckAwsServiceCatalogPortfolioPrincipalAssociationNotPresentInAws(&portfolioId, "*")
					},
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccAWSServiceCatalogPortfolioPrincipalAssociation_Portfolio_update(t *testing.T) {
	salt := acctest.RandString(5)
	salt2 := acctest.RandString(5)
	var portfolioId1, portfolioId2 string
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAwsServiceCatalogPortfolioPrincipalAssociationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSServiceCatalogPortfolioPrincipalAssociationConfig_basic(salt, salt),
				Check:  testAccCheckAwsServiceCatalogPortfolioPrincipalAssociationExists(salt, salt, &portfolioId1),
			},
			{
				Config: testAccAWSServiceCatalogPortfolioPrincipalAssociationConfig_basic(salt2, salt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsServiceCatalogPortfolioPrincipalAssociationExists(salt2, salt, &portfolioId2),
					func(s *terraform.State) error {
						if portfolioId1 == portfolioId2 {
							return fmt.Errorf("Portfolio ID should have changed")
						}
						return testAccCheckAwsServiceCatalogPortfolioPrincipalAssociationNotPresentInAws(&portfolioId1, "*")
					},
				),
			},
		},
	})
}

func TestAccAWSServiceCatalogPortfolioPrincipalAssociation_Principal_update(t *testing.T) {
	salt := acctest.RandString(5)
	salt2 := acctest.RandString(5)
	var portfolioId1, portfolioId2 string
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAwsServiceCatalogPortfolioPrincipalAssociationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSServiceCatalogPortfolioPrincipalAssociationConfig_basic(salt, salt),
				Check:  testAccCheckAwsServiceCatalogPortfolioPrincipalAssociationExists(salt, salt, &portfolioId1),
			},
			{
				Config: testAccAWSServiceCatalogPortfolioPrincipalAssociationConfig_basic(salt, salt2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsServiceCatalogPortfolioPrincipalAssociationExists(salt, salt2, &portfolioId2),
					func(s *terraform.State) error {
						if portfolioId1 != portfolioId2 {
							return fmt.Errorf("Portfolio should not have changed")
						}
						return testAccCheckAwsServiceCatalogPortfolioPrincipalAssociationNotPresentInAws(&portfolioId2, salt)
					},
				),
			},
		},
	})
}

func TestAccAWSServiceCatalogPortfolioPrincipalAssociation_update_all(t *testing.T) {
	salt := acctest.RandString(5)
	salt2 := acctest.RandString(5)
	var portfolioId1, portfolioId2 string
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAwsServiceCatalogPortfolioPrincipalAssociationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSServiceCatalogPortfolioPrincipalAssociationConfig_basic(salt, salt),
				Check:  testAccCheckAwsServiceCatalogPortfolioPrincipalAssociationExists(salt, salt, &portfolioId1),
			},
			{
				Config: testAccAWSServiceCatalogPortfolioPrincipalAssociationConfig_basic(salt2, salt2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsServiceCatalogPortfolioPrincipalAssociationExists(salt2, salt2, &portfolioId2),
					func(s *terraform.State) error {
						if portfolioId1 == portfolioId2 {
							return fmt.Errorf("Portfolio should have changed")
						}
						return testAccCheckAwsServiceCatalogPortfolioPrincipalAssociationNotPresentInAws(&portfolioId2, salt)
					},
				),
			},
		},
	})
}

func testAccCheckAwsServiceCatalogPortfolioPrincipalAssociationExists(portfolioSalt, principalSalt string, portfolioIdToSet *string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := testAccProvider.Meta().(*AWSClient).scconn
		rs, ok := s.RootModule().Resources["aws_servicecatalog_portfolio_principal_association.test"]
		if !ok {
			return fmt.Errorf("association not found")
		}

		_, portfolioId, principalArn, err := parseServiceCatalogPortfolioPrincipalAssociationResourceId(rs.Primary.ID)
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

		if !strings.Contains(principalArn, principalSalt) {
			return fmt.Errorf("principal ARN from ID %s did not contain expected salt '%s'", rs.Primary.ID, principalSalt)
		}

		*portfolioIdToSet = portfolioId

		input := servicecatalog.ListPrincipalsForPortfolioInput{PortfolioId: &portfolioId}
		page, err := conn.ListPrincipalsForPortfolio(&input)
		if err != nil {
			return err
		}
		for _, principalDetail := range page.Principals {
			if aws.StringValue(principalDetail.PrincipalARN) == principalArn {
				// found
				return nil
			}
		}
		return fmt.Errorf("association not found between portfolio %s and principal %s; principals were: %v", portfolioId, principalArn, page.Principals)
	}
}

func testAccCheckAwsServiceCatalogPortfolioPrincipalAssociationDisappears() resource.TestCheckFunc {
	return func(s *terraform.State) error {
		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_servicecatalog_portfolio_principal_association" {
				continue // not our monkey
			}
			_, portfolioId, principalArn, err := parseServiceCatalogPortfolioPrincipalAssociationResourceId(rs.Primary.ID)
			if err != nil {
				return err
			}
			conn := testAccProvider.Meta().(*AWSClient).scconn
			input := servicecatalog.DisassociatePrincipalFromPortfolioInput{
				PortfolioId:  aws.String(portfolioId),
				PrincipalARN: aws.String(principalArn),
			}
			_, err = conn.DisassociatePrincipalFromPortfolio(&input)
			if err != nil {
				return fmt.Errorf("deleting Service Catalog Principal(%s)/Portfolio(%s) Association failed: %s",
					principalArn, portfolioId, err.Error())
			}
			return nil
		}
		return fmt.Errorf("no matching resource found to make disappear")
	}
}

func testAccCheckAwsServiceCatalogPortfolioPrincipalAssociationDestroy(s *terraform.State) error {
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_servicecatalog_portfolio_principal_association" {
			continue // not our monkey
		}
		_, portfolioId, principalArn, err := parseServiceCatalogPortfolioPrincipalAssociationResourceId(rs.Primary.ID)
		if err != nil {
			return err
		}
		return testAccCheckAwsServiceCatalogPortfolioPrincipalAssociationNotPresentInAws(&portfolioId, principalArn)
	}
	return nil
}

func testAccCheckAwsServiceCatalogPortfolioPrincipalAssociationNotPresentInAws(portfolioId *string, principalArnSubstring string) error {
	conn := testAccProvider.Meta().(*AWSClient).scconn
	input := servicecatalog.ListPrincipalsForPortfolioInput{PortfolioId: portfolioId}
	page, err := conn.ListPrincipalsForPortfolio(&input)
	if err != nil {
		if isAWSErr(err, servicecatalog.ErrCodeResourceNotFoundException, "") {
			return nil // not found is good
		}
		return err // some other unexpected error
	}
	if len(page.Principals) == 0 {
		return nil
	}
	if principalArnSubstring == "*" {
		return fmt.Errorf("expected AWS Service Catalog Portfolio Principal Associations to be gone for %s, but still found some: %v", *portfolioId, page.Principals)
	} else {
		for _, principalDetail := range page.Principals {
			if strings.Contains(aws.StringValue(principalDetail.PrincipalARN), principalArnSubstring) {
				return fmt.Errorf("expected AWS Service Catalog Portfolio Principal Association to be gone, but it was still found: %s", aws.StringValue(principalDetail.PrincipalARN))
			}
		}
		// not found
		return nil
	}
}

func testAccAWSServiceCatalogPortfolioPrincipalAssociationConfig_basic(portfolioSalt, principalSalt string) string {
	return composeConfig(
		testAccAWSServiceCatalogPortfolioPrincipalAssociationConfig_portfolio(portfolioSalt),
		testAccAWSServiceCatalogPortfolioPrincipalAssociationConfig_role(principalSalt),
		testAccAWSServiceCatalogPortfolioPrincipalAssociationConfig_association(portfolioSalt, principalSalt))
}

func testAccAWSServiceCatalogPortfolioPrincipalAssociationConfig_portfolio(salt string) string {
	// based on testAccAWSServiceCatalogPortfolioConfig_basic
	return fmt.Sprintf(`
resource "aws_servicecatalog_portfolio" "test-%s" {
  name          = "%s"
  description   = "test-2"
  provider_name = "test-3"
}
`, salt, "tfm-test-"+salt)
}

func testAccAWSServiceCatalogPortfolioPrincipalAssociationConfig_role(salt string) string {
	roleName := "tfm-sc-tester-" + salt
	return fmt.Sprintf(`
# IAM
resource "aws_iam_role" "test-%s" {
  name = "%s"
  assume_role_policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Action": "sts:AssumeRole",
      "Principal": { "AWS": "*" },
      "Effect": "Allow",
      "Sid": ""
    }
  ]
}
EOF
}
`, salt, roleName)
}

func testAccAWSServiceCatalogPortfolioPrincipalAssociationConfig_association(portfolioSalt, principalSalt string) string {
	return fmt.Sprintf(`
resource "aws_servicecatalog_portfolio_principal_association" "test" {
    portfolio_id = aws_servicecatalog_portfolio.test-%s.id
    principal_arn = aws_iam_role.test-%s.arn
}
`, portfolioSalt, principalSalt)
}
