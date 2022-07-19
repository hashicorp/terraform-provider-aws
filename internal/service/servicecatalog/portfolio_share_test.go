package servicecatalog_test

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/service/servicecatalog"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfservicecatalog "github.com/hashicorp/terraform-provider-aws/internal/service/servicecatalog"
)

func TestAccServiceCatalogPortfolioShare_basic(t *testing.T) {
	var providers []*schema.Provider
	resourceName := "aws_servicecatalog_portfolio_share.test"
	compareName := "aws_servicecatalog_portfolio.test"
	dataSourceName := "data.aws_caller_identity.alternate"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(t)
			acctest.PreCheckAlternateAccount(t)
			acctest.PreCheckPartitionHasService(servicecatalog.EndpointsID, t)
		},
		ErrorCheck:        acctest.ErrorCheck(t, servicecatalog.EndpointsID),
		ProviderFactories: acctest.FactoriesAlternate(&providers),
		CheckDestroy:      testAccCheckPortfolioShareDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccPortfolioShareConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPortfolioShareExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "accept_language", tfservicecatalog.AcceptLanguageEnglish),
					resource.TestCheckResourceAttr(resourceName, "accepted", "false"),
					resource.TestCheckResourceAttrPair(resourceName, "principal_id", dataSourceName, "account_id"),
					resource.TestCheckResourceAttrPair(resourceName, "portfolio_id", compareName, "id"),
					resource.TestCheckResourceAttr(resourceName, "share_tag_options", "true"),
					resource.TestCheckResourceAttr(resourceName, "type", servicecatalog.DescribePortfolioShareTypeAccount),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"accept_language",
				},
			},
		},
	})
}

func TestAccServiceCatalogPortfolioShare_organizationalUnit(t *testing.T) {
	resourceName := "aws_servicecatalog_portfolio_share.test"
	compareName := "aws_servicecatalog_portfolio.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(t)
			acctest.PreCheckOrganizationsEnabled(t)
			acctest.PreCheckOrganizationManagementAccount(t)
			acctest.PreCheckPartitionHasService(servicecatalog.EndpointsID, t)
		},
		ErrorCheck:        acctest.ErrorCheck(t, servicecatalog.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckPortfolioShareDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccPortfolioShareConfig_organizationalUnit(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPortfolioShareExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "accept_language", tfservicecatalog.AcceptLanguageEnglish),
					resource.TestCheckResourceAttr(resourceName, "accepted", "true"),
					resource.TestCheckResourceAttrPair(resourceName, "principal_id", "aws_organizations_organizational_unit.test", "id"),
					resource.TestCheckResourceAttrPair(resourceName, "portfolio_id", compareName, "id"),
					resource.TestCheckResourceAttr(resourceName, "share_tag_options", "true"),
					resource.TestCheckResourceAttr(resourceName, "type", servicecatalog.DescribePortfolioShareTypeOrganizationalUnit),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"accept_language",
				},
			},
		},
	})
}

func testAccCheckPortfolioShareDestroy(s *terraform.State) error {
	conn := acctest.Provider.Meta().(*conns.AWSClient).ServiceCatalogConn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_servicecatalog_portfolio_share" {
			continue
		}

		output, err := tfservicecatalog.FindPortfolioShare(
			conn,
			rs.Primary.Attributes["portfolio_id"],
			rs.Primary.Attributes["type"],
			rs.Primary.Attributes["principal_id"],
		)

		if tfawserr.ErrCodeEquals(err, servicecatalog.ErrCodeResourceNotFoundException) {
			return nil
		}

		if err != nil {
			return fmt.Errorf("error getting Service Catalog Portfolio Share (%s): %w", rs.Primary.ID, err)
		}

		if output != nil {
			return fmt.Errorf("Service Catalog Portfolio Share (%s) still exists", rs.Primary.ID)
		}
	}

	return nil
}

func testAccCheckPortfolioShareExists(resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]

		if !ok {
			return fmt.Errorf("resource not found: %s", resourceName)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).ServiceCatalogConn

		_, err := tfservicecatalog.FindPortfolioShare(
			conn,
			rs.Primary.Attributes["portfolio_id"],
			rs.Primary.Attributes["type"],
			rs.Primary.Attributes["principal_id"],
		)

		if tfawserr.ErrCodeEquals(err, servicecatalog.ErrCodeResourceNotFoundException) {
			return fmt.Errorf("Service Catalog Portfolio Share (%s) not found", rs.Primary.ID)
		}

		if err != nil {
			return fmt.Errorf("error getting Service Catalog Portfolio Share (%s): %w", rs.Primary.ID, err)
		}

		return nil
	}
}

func testAccPortfolioShareConfig_basic(rName string) string {
	return acctest.ConfigCompose(acctest.ConfigAlternateAccountProvider(), fmt.Sprintf(`
data "aws_caller_identity" "alternate" {
  provider = "awsalternate"
}

resource "aws_servicecatalog_portfolio" "test" {
  name          = %[1]q
  description   = %[1]q
  provider_name = %[1]q
}

resource "aws_servicecatalog_portfolio_share" "test" {
  accept_language     = "en"
  portfolio_id        = aws_servicecatalog_portfolio.test.id
  share_tag_options   = true
  type                = "ACCOUNT"
  principal_id        = data.aws_caller_identity.alternate.account_id
  wait_for_acceptance = false
}
`, rName))
}

func testAccPortfolioShareConfig_organizationalUnit(rName string) string {
	return fmt.Sprintf(`
data "aws_partition" "current" {}

resource "aws_servicecatalog_organizations_access" "test" {
  enabled = "true"
}

resource "aws_servicecatalog_portfolio" "test" {
  name          = %[1]q
  description   = %[1]q
  provider_name = %[1]q
}

data "aws_organizations_organization" "test" {}

resource "aws_organizations_organizational_unit" "test" {
  name      = %[1]q
  parent_id = data.aws_organizations_organization.test.roots[0].id
}

resource "aws_servicecatalog_portfolio_share" "test" {
  accept_language   = "en"
  portfolio_id      = aws_servicecatalog_portfolio.test.id
  share_tag_options = true
  type              = "ORGANIZATIONAL_UNIT"
  principal_id      = aws_organizations_organizational_unit.test.arn
}
`, rName)
}
