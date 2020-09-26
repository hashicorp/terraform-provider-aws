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

func TestAccAWSServiceCatalogPortfolioPrincipalAssociation_basic(t *testing.T) {
	saltedName := "tf-acc-test-" + acctest.RandString(5) // RandomWithPrefix exceeds max length 20
	resourceName := "aws_servicecatalog_portfolio_principal_association.test"
	var portfolioId string
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAwsServiceCatalogPortfolioPrincipalAssociationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSServiceCatalogPortfolioPrincipalAssociationConfig_basic("", "", saltedName, saltedName),
				Check:  testAccCheckAwsServiceCatalogPortfolioPrincipalAssociationExists("", "", saltedName, saltedName, &portfolioId),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccAWSServiceCatalogPortfolioPrincipalAssociation_disappears(t *testing.T) {
	saltedName := "tf-acc-test-" + acctest.RandString(5) // RandomWithPrefix exceeds max length 20
	resourceName := "aws_servicecatalog_portfolio_principal_association.test"
	var portfolioId string
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAwsServiceCatalogPortfolioPrincipalAssociationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSServiceCatalogPortfolioPrincipalAssociationConfig_basic("", "", saltedName, saltedName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsServiceCatalogPortfolioPrincipalAssociationExists("", "", saltedName, saltedName, &portfolioId),
					testAccCheckResourceDisappears(testAccProvider, resourceAwsServiceCatalogPortfolioPrincipalAssociation(), resourceName),
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
	saltedName := "tf-acc-test-" + acctest.RandString(5)   // RandomWithPrefix exceeds max length 20
	saltedName2 := "tf-acc-test2-" + acctest.RandString(5) // RandomWithPrefix exceeds max length 20
	var portfolioId1, portfolioId2 string
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAwsServiceCatalogPortfolioPrincipalAssociationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSServiceCatalogPortfolioPrincipalAssociationConfig_basic("1", "1", saltedName, saltedName),
				Check:  testAccCheckAwsServiceCatalogPortfolioPrincipalAssociationExists("1", "1", saltedName, saltedName, &portfolioId1),
			},
			{
				Config: testAccAWSServiceCatalogPortfolioPrincipalAssociationConfig_basic("2", "1", saltedName2, saltedName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsServiceCatalogPortfolioPrincipalAssociationExists("2", "1", saltedName2, saltedName, &portfolioId2),
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
	saltedName := "tf-acc-test-" + acctest.RandString(5)   // RandomWithPrefix exceeds max length 20
	saltedName2 := "tf-acc-test2-" + acctest.RandString(5) // RandomWithPrefix exceeds max length 20
	var portfolioId1, portfolioId2 string
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAwsServiceCatalogPortfolioPrincipalAssociationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSServiceCatalogPortfolioPrincipalAssociationConfig_basic("1", "1", saltedName, saltedName),
				Check:  testAccCheckAwsServiceCatalogPortfolioPrincipalAssociationExists("1", "1", saltedName, saltedName, &portfolioId1),
			},
			{
				Config: testAccAWSServiceCatalogPortfolioPrincipalAssociationConfig_basic("1", "2", saltedName, saltedName2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsServiceCatalogPortfolioPrincipalAssociationExists("1", "2", saltedName, saltedName2, &portfolioId2),
					func(s *terraform.State) error {
						if portfolioId1 != portfolioId2 {
							return fmt.Errorf("Portfolio should not have changed")
						}
						return testAccCheckAwsServiceCatalogPortfolioPrincipalAssociationNotPresentInAws(&portfolioId2, saltedName)
					},
				),
			},
		},
	})
}

func TestAccAWSServiceCatalogPortfolioPrincipalAssociation_update_all(t *testing.T) {
	saltedName := "tf-acc-test-" + acctest.RandString(5)   // RandomWithPrefix exceeds max length 20
	saltedName2 := "tf-acc-test2-" + acctest.RandString(5) // RandomWithPrefix exceeds max length 20
	var portfolioId1, portfolioId2 string
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAwsServiceCatalogPortfolioPrincipalAssociationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSServiceCatalogPortfolioPrincipalAssociationConfig_basic("1", "1", saltedName, saltedName),
				Check:  testAccCheckAwsServiceCatalogPortfolioPrincipalAssociationExists("1", "1", saltedName, saltedName, &portfolioId1),
			},
			{
				Config: testAccAWSServiceCatalogPortfolioPrincipalAssociationConfig_basic("2", "2", saltedName2, saltedName2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsServiceCatalogPortfolioPrincipalAssociationExists("2", "2", saltedName2, saltedName2, &portfolioId2),
					func(s *terraform.State) error {
						if portfolioId1 == portfolioId2 {
							return fmt.Errorf("Portfolio should have changed")
						}
						return testAccCheckAwsServiceCatalogPortfolioPrincipalAssociationNotPresentInAws(&portfolioId2, saltedName)
					},
				),
			},
		},
	})
}

func testAccCheckAwsServiceCatalogPortfolioPrincipalAssociationExists(portfolioSuffix, principalSuffix, portfolioSaltedName, principalSaltedName string, portfolioIdToSet *string) resource.TestCheckFunc {
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

		rsPortfolio, ok := s.RootModule().Resources["aws_servicecatalog_portfolio.test"+portfolioSuffix]
		if !ok {
			return fmt.Errorf("portfolio %s not found", portfolioSaltedName)
		}
		if !strings.Contains(rsPortfolio.Primary.Attributes["name"], portfolioSaltedName) {
			return fmt.Errorf("portfolio from association ID %s did not contain expected salt '%s'", rs.Primary.ID, portfolioSaltedName)
		}

		rsPrincipal, ok2 := s.RootModule().Resources["aws_iam_role.test"+principalSuffix]
		if !ok2 {
			return fmt.Errorf("principal %s not found", principalSaltedName)
		}
		if !strings.Contains(rsPrincipal.Primary.Attributes["name"], principalSaltedName) {
			return fmt.Errorf("principal from association ID %s did not contain expected salt '%s'", rs.Primary.ID, principalSaltedName)
		}

		if !strings.Contains(principalArn, principalSaltedName) {
			return fmt.Errorf("principal ARN from ID %s did not contain expected salt '%s'", rs.Primary.ID, principalSaltedName)
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

func testAccAWSServiceCatalogPortfolioPrincipalAssociationConfig_basic(portfolioSuffix, principalSuffix, portfolioSaltedName, principalSaltedName string) string {
	return composeConfig(
		testAccAWSServiceCatalogPortfolioPrincipalAssociationConfig_portfolio(portfolioSuffix, portfolioSaltedName),
		testAccAWSServiceCatalogPortfolioPrincipalAssociationConfig_role(principalSuffix, principalSaltedName),
		testAccAWSServiceCatalogPortfolioPrincipalAssociationConfig_association(portfolioSuffix, principalSuffix))
}

func testAccAWSServiceCatalogPortfolioPrincipalAssociationConfig_portfolio(suffix, saltedName string) string {
	// based on testAccAWSServiceCatalogPortfolioConfig_basic
	return fmt.Sprintf(`
resource "aws_servicecatalog_portfolio" "test%[2]s" {
  name          = "%[1]s"
  description   = "test-2"
  provider_name = "test-3"
}
`, saltedName, suffix)
}

func testAccAWSServiceCatalogPortfolioPrincipalAssociationConfig_role(suffix, saltedName string) string {
	return fmt.Sprintf(`
# IAM
resource "aws_iam_role" "test%[2]s" {
  name = "%[1]s"
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
`, saltedName, suffix)
}

func testAccAWSServiceCatalogPortfolioPrincipalAssociationConfig_association(portfolioSuffix, principalSuffix string) string {
	return fmt.Sprintf(`
resource "aws_servicecatalog_portfolio_principal_association" "test" {
    portfolio_id = aws_servicecatalog_portfolio.test%[1]s.id
    principal_arn = aws_iam_role.test%[2]s.arn
}
`, portfolioSuffix, principalSuffix)
}
