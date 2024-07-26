// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package servicecatalog_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/service/servicecatalog"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfservicecatalog "github.com/hashicorp/terraform-provider-aws/internal/service/servicecatalog"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func testAccPortfolioShare_basic(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_servicecatalog_portfolio_share.test"
	compareName := "aws_servicecatalog_portfolio.test"
	dataSourceName := "data.aws_caller_identity.alternate"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckAlternateAccount(t)
			acctest.PreCheckPartitionHasService(t, servicecatalog.EndpointsID)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.ServiceCatalogServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5FactoriesAlternate(ctx, t),
		CheckDestroy:             testAccCheckPortfolioShareDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccPortfolioShareConfig_basic(rName, true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPortfolioShareExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "accept_language", tfservicecatalog.AcceptLanguageEnglish),
					resource.TestCheckResourceAttr(resourceName, "accepted", acctest.CtFalse),
					resource.TestCheckResourceAttrPair(resourceName, "principal_id", dataSourceName, names.AttrAccountID),
					resource.TestCheckResourceAttrPair(resourceName, "portfolio_id", compareName, names.AttrID),
					resource.TestCheckResourceAttr(resourceName, "share_principals", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "share_tag_options", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, names.AttrType, servicecatalog.DescribePortfolioShareTypeAccount),
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
			{
				Config: testAccPortfolioShareConfig_basic(rName, false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPortfolioShareExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "accept_language", tfservicecatalog.AcceptLanguageEnglish),
					resource.TestCheckResourceAttr(resourceName, "accepted", acctest.CtFalse),
					resource.TestCheckResourceAttrPair(resourceName, "principal_id", dataSourceName, names.AttrAccountID),
					resource.TestCheckResourceAttrPair(resourceName, "portfolio_id", compareName, names.AttrID),
					resource.TestCheckResourceAttr(resourceName, "share_principals", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "share_tag_options", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, names.AttrType, servicecatalog.DescribePortfolioShareTypeAccount),
				),
			},
		},
	})
}

func testAccPortfolioShare_sharePrincipals(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_servicecatalog_portfolio_share.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckOrganizationsEnabled(ctx, t)
			acctest.PreCheckOrganizationManagementAccount(ctx, t)
			acctest.PreCheckPartitionHasService(t, servicecatalog.EndpointsID)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.ServiceCatalogServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5FactoriesAlternate(ctx, t),
		CheckDestroy:             testAccCheckPortfolioShareDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccPortfolioShareConfig_sharePrincipals(rName, true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPortfolioShareExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "share_principals", acctest.CtTrue),
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
			{
				Config: testAccPortfolioShareConfig_sharePrincipals(rName, false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPortfolioShareExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "share_principals", acctest.CtFalse),
				),
			},
		},
	})
}

func testAccPortfolioShare_organizationalUnit(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_servicecatalog_portfolio_share.test"
	compareName := "aws_servicecatalog_portfolio.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckOrganizationsEnabled(ctx, t)
			acctest.PreCheckOrganizationManagementAccount(ctx, t)
			acctest.PreCheckPartitionHasService(t, servicecatalog.EndpointsID)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.ServiceCatalogServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckPortfolioShareDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccPortfolioShareConfig_organizationalUnit(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPortfolioShareExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "accept_language", tfservicecatalog.AcceptLanguageEnglish),
					resource.TestCheckResourceAttr(resourceName, "accepted", acctest.CtTrue),
					resource.TestCheckResourceAttrPair(resourceName, "principal_id", "aws_organizations_organizational_unit.test", names.AttrID),
					resource.TestCheckResourceAttrPair(resourceName, "portfolio_id", compareName, names.AttrID),
					resource.TestCheckResourceAttr(resourceName, "share_tag_options", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, names.AttrType, servicecatalog.DescribePortfolioShareTypeOrganizationalUnit),
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

func testAccPortfolioShare_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_servicecatalog_portfolio_share.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckAlternateAccount(t)
			acctest.PreCheckPartitionHasService(t, servicecatalog.EndpointsID)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.ServiceCatalogServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5FactoriesAlternate(ctx, t),
		CheckDestroy:             testAccCheckPortfolioShareDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccPortfolioShareConfig_basic(rName, true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPortfolioShareExists(ctx, resourceName),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfservicecatalog.ResourcePortfolioShare(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckPortfolioShareDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).ServiceCatalogConn(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_servicecatalog_portfolio_share" {
				continue
			}

			_, err := tfservicecatalog.FindPortfolioShare(ctx, conn,
				rs.Primary.Attributes["portfolio_id"],
				rs.Primary.Attributes[names.AttrType],
				rs.Primary.Attributes["principal_id"],
			)

			if tfresource.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("Service Catalog Portfolio Share %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccCheckPortfolioShareExists(ctx context.Context, n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No Service Catalog Portfolio Share ID is set")
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).ServiceCatalogConn(ctx)

		_, err := tfservicecatalog.FindPortfolioShare(ctx, conn,
			rs.Primary.Attributes["portfolio_id"],
			rs.Primary.Attributes[names.AttrType],
			rs.Primary.Attributes["principal_id"],
		)

		return err
	}
}

func testAccPortfolioShareConfig_basic(rName string, share bool) string {
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
  share_tag_options   = %[2]t
  type                = "ACCOUNT"
  principal_id        = data.aws_caller_identity.alternate.account_id
  wait_for_acceptance = false
}
`, rName, share))
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

func testAccPortfolioShareConfig_sharePrincipals(rName string, share bool) string {
	return fmt.Sprintf(`
data "aws_organizations_organization" "current" {}

resource "aws_servicecatalog_organizations_access" "test" {
  enabled = "true"
}

resource "aws_servicecatalog_portfolio" "test" {
  name          = %[1]q
  description   = %[1]q
  provider_name = %[1]q
}

resource "aws_servicecatalog_portfolio_share" "test" {
  accept_language     = "en"
  portfolio_id        = aws_servicecatalog_portfolio.test.id
  share_principals    = %[2]t
  type                = "ORGANIZATION"
  principal_id        = data.aws_organizations_organization.current.arn
  wait_for_acceptance = false

  depends_on = [aws_servicecatalog_organizations_access.test]
}
`, rName, share)
}
