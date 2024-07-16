// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package servicecatalog_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/YakDriver/regexache"
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

func TestAccServiceCatalogPortfolio_basic(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_servicecatalog_portfolio.test"
	name := sdkacctest.RandString(5)
	var dpo servicecatalog.DescribePortfolioOutput

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ServiceCatalogServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckPortfolioDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccPortfolioConfig_basic(name),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPortfolioExists(ctx, resourceName, &dpo),
					acctest.MatchResourceAttrRegionalARN(resourceName, names.AttrARN, "catalog", regexache.MustCompile(`portfolio/.+`)),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrCreatedTime),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, name),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, "test-2"),
					resource.TestCheckResourceAttr(resourceName, names.AttrProviderName, "test-3"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct0),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccServiceCatalogPortfolio_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	name := sdkacctest.RandString(5)
	resourceName := "aws_servicecatalog_portfolio.test"
	var dpo servicecatalog.DescribePortfolioOutput

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ServiceCatalogServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckPortfolioDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccPortfolioConfig_basic(name),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPortfolioExists(ctx, resourceName, &dpo),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfservicecatalog.ResourcePortfolio(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckPortfolioExists(ctx context.Context, n string, v *servicecatalog.DescribePortfolioOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No Service Catalog Portfolio ID is set")
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).ServiceCatalogConn(ctx)

		output, err := tfservicecatalog.FindPortfolioByID(ctx, conn, rs.Primary.ID)

		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

func testAccCheckPortfolioDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).ServiceCatalogConn(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_servicecatalog_portfolio" {
				continue
			}

			_, err := tfservicecatalog.FindPortfolioByID(ctx, conn, rs.Primary.ID)

			if tfresource.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("Service Catalog Portfolio %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccPortfolioConfig_basic(name string) string {
	return fmt.Sprintf(`
resource "aws_servicecatalog_portfolio" "test" {
  name          = "%s"
  description   = "test-2"
  provider_name = "test-3"
}
`, name)
}
