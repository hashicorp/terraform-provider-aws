package aws

import (
    "fmt"
    "github.com/hashicorp/terraform-plugin-sdk/helper/acctest"
    "github.com/aws/aws-sdk-go/service/servicecatalog"
    "github.com/hashicorp/terraform-plugin-sdk/helper/resource"
    "github.com/hashicorp/terraform-plugin-sdk/terraform"
    "testing"
)

func TestAccAWSServiceCatalogPortfolioPrincipalAssociation_Basic(t *testing.T) {
    salt := acctest.RandString(5)
    resource.ParallelTest(t, resource.TestCase{
        PreCheck: func() { testAccPreCheck(t) },
        Providers: testAccProviders,
        CheckDestroy: testAccCheckServiceCatalogPortfolioPrincipalAssociationDestroy,
        Steps: []resource.TestStep{
            {
                Config: testAccCheckAwsServiceCatalogPortfolioPrincipalAssociationConfigBasic(salt),
                Check: testAccCheckAwsServiceCatalogPortfolioPrincipalAssociation(),
            },
            {
                ResourceName: "aws_servicecatalog_portfolio_principal_association.association",
                ImportState: true,
                ImportStateVerify: true,
            },
        },
    })
}

func testAccCheckAwsServiceCatalogPortfolioPrincipalAssociation() resource.TestCheckFunc {
    return func(s *terraform.State) error {
        conn := testAccProvider.Meta().(*AWSClient).scconn
        for _, rs := range s.RootModule().Resources {
            if rs.Type != "aws_servicecatalog_portfolio_principal_association" {
                continue // not our monkey
            }
            _, portfolioId, principalArn := parseServiceCatalogPortfolioPrincipalAssociationResourceId(rs.Primary.ID)
            input := servicecatalog.ListPrincipalsForPortfolioInput{PortfolioId: &portfolioId}
            page, err := conn.ListPrincipalsForPortfolio(&input)
            if err != nil {
                return err
            }
            for _, principalDetail := range page.Principals {
                if *principalDetail.PrincipalARN == principalArn {
                    return nil //is good
                }
            }
            return fmt.Errorf("association not found between portfolio %s and principal %s", portfolioId, principalArn)
        }
        return fmt.Errorf("no associations found")
    }
}

func testAccCheckServiceCatalogPortfolioPrincipalAssociationDestroy(s *terraform.State) error {
    conn := testAccProvider.Meta().(*AWSClient).scconn
    for _, rs := range s.RootModule().Resources {
        if rs.Type != "aws_servicecatalog_portfolio_principal_association" {
            continue // not our monkey
        }
        _, portfolioId, principalArn := parseServiceCatalogPortfolioPrincipalAssociationResourceId(rs.Primary.ID)
        input := servicecatalog.ListPrincipalsForPortfolioInput{PortfolioId: &portfolioId}
        page, err := conn.ListPrincipalsForPortfolio(&input)
        if err != nil {
            if isAWSErr(err, servicecatalog.ErrCodeResourceNotFoundException, "") {
                return nil // not found for principal is good
            }
            return err // some other unexpected error
        }
        for _, principalDetail := range page.Principals {
            if *principalDetail.PrincipalARN == principalArn {
                return fmt.Errorf("expected AWS Service Catalog Portfolio Principal Association to be gone, but it was still found")
            }
        }
    }
    return nil
}

func testAccCheckAwsServiceCatalogPortfolioPrincipalAssociationConfigBasic(salt string) string {
    return testAccCheckAwsServiceCatalogPortfolioResourceConfigBasic("tfm_automated_test-"+salt) + "\n" + 
        testAccCheckAwsServiceCatalogPortfolioPrincipalAssociationConfigRoleAndAssociation(salt)
}

func testAccCheckAwsServiceCatalogPortfolioPrincipalAssociationConfigRole(salt string) string {
    roleName := "tfm-sc-tester-" + salt;
    return fmt.Sprintf(`
# IAM
resource "aws_iam_role" "tfm-sc-tester" {
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
`, roleName)
}

func testAccCheckAwsServiceCatalogPortfolioPrincipalAssociationConfigAssociation() string {
    return fmt.Sprintf(`
resource "aws_servicecatalog_portfolio_principal_association" "association" {
    portfolio_id = aws_servicecatalog_portfolio.test.id
    principal_arn = aws_iam_role.tfm-sc-tester.arn
}`)
}

func testAccCheckAwsServiceCatalogPortfolioPrincipalAssociationConfigRoleAndAssociation(salt string) string {
    return testAccCheckAwsServiceCatalogPortfolioPrincipalAssociationConfigRole(salt) +
        testAccCheckAwsServiceCatalogPortfolioPrincipalAssociationConfigAssociation()
}
