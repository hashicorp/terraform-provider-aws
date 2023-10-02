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
)

func testAccPortfolioShareData_basic(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_servicecatalog_portfolio_share.test"
	compareName := "aws_servicecatalog_portfolio.test"
	dataSourceName := "data.aws_servicecatalog_portfolio_share.test" // "data.aws_caller_identity.alternate"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckAlternateAccount(t)
			acctest.PreCheckPartitionHasService(t, servicecatalog.EndpointsID)
		},
		ErrorCheck:               acctest.ErrorCheck(t, servicecatalog.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5FactoriesAlternate(ctx, t),
		CheckDestroy:             testAccCheckPortfolioShareDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccPortfolioShareDataConfig_basic(rName, true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPortfolioShareExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "accept_language", tfservicecatalog.AcceptLanguageEnglish),
					resource.TestCheckResourceAttr(resourceName, "accepted", "false"),
					resource.TestCheckResourceAttrPair(resourceName, "principal_id", dataSourceName, "account_id"),
					resource.TestCheckResourceAttrPair(resourceName, "portfolio_id", compareName, "id"),
					resource.TestCheckResourceAttr(resourceName, "share_principals", "false"),
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
			{
				Config: testAccPortfolioShareDataConfig_basic(rName, false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPortfolioShareExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "accept_language", tfservicecatalog.AcceptLanguageEnglish),
					resource.TestCheckResourceAttr(resourceName, "accepted", "false"),
					resource.TestCheckResourceAttrPair(resourceName, "principal_id", dataSourceName, "account_id"),
					resource.TestCheckResourceAttrPair(resourceName, "portfolio_id", compareName, "id"),
					resource.TestCheckResourceAttr(resourceName, "share_principals", "false"),
					resource.TestCheckResourceAttr(resourceName, "share_tag_options", "false"),
					resource.TestCheckResourceAttr(resourceName, "type", servicecatalog.DescribePortfolioShareTypeAccount),
				),
			},
		},
	})
}

func testAccPortfolioShareData_sharePrincipals(t *testing.T) {
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
		ErrorCheck:               acctest.ErrorCheck(t, servicecatalog.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5FactoriesAlternate(ctx, t),
		CheckDestroy:             testAccCheckPortfolioShareDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccPortfolioShareConfig_sharePrincipals(rName, true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPortfolioShareExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "share_principals", "true"),
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
					resource.TestCheckResourceAttr(resourceName, "share_principals", "false"),
				),
			},
		},
	})
}

func testAccPortfolioShareData_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_servicecatalog_portfolio_share.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckAlternateAccount(t)
			acctest.PreCheckPartitionHasService(t, servicecatalog.EndpointsID)
		},
		ErrorCheck:               acctest.ErrorCheck(t, servicecatalog.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5FactoriesAlternate(ctx, t),
		CheckDestroy:             testAccCheckPortfolioShareDataDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccPortfolioShareDataConfig_basic(rName, true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPortfolioShareDataExists(ctx, resourceName),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfservicecatalog.ResourcePortfolioShare(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckPortfolioShareDataDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).ServiceCatalogConn(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_servicecatalog_portfolio_share" {
				continue
			}

			_, err := tfservicecatalog.FindPortfolioShare(ctx, conn,
				rs.Primary.Attributes["portfolio_id"],
				rs.Primary.Attributes["type"],
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

func testAccCheckPortfolioShareDataExists(ctx context.Context, n string) resource.TestCheckFunc {
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
			rs.Primary.Attributes["type"],
			rs.Primary.Attributes["principal_id"],
		)

		return err
	}
}

func testAccPortfolioShareDataConfig_basic(rName string, share bool) string {
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

data "aws_servicecatalog_portfolio_share" "test" {
  portfolio_id = aws_servicecatalog_portfolio_share.test.id
  type         = aws_servicecatalog_portfolio_share.test.type
}
`, rName, share))
}
