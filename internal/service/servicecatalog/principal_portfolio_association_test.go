// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package servicecatalog_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	tfservicecatalog "github.com/hashicorp/terraform-provider-aws/internal/service/servicecatalog"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccServiceCatalogPrincipalPortfolioAssociation_basic(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_servicecatalog_principal_portfolio_association.test"
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ServiceCatalogServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckPrincipalPortfolioAssociationDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccPrincipalPortfolioAssociationConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPrincipalPortfolioAssociationExists(ctx, t, resourceName),
					resource.TestCheckResourceAttrPair(resourceName, "portfolio_id", "aws_servicecatalog_portfolio.test", names.AttrID),
					resource.TestCheckResourceAttrPair(resourceName, "principal_arn", "aws_iam_role.test", names.AttrARN),
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

func TestAccServiceCatalogPrincipalPortfolioAssociation_iam_pattern(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_servicecatalog_principal_portfolio_association.test"
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ServiceCatalogServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckPrincipalPortfolioAssociationDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccPrincipalPortfolioAssociationConfig_iam_pattern(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPrincipalPortfolioAssociationExists(ctx, t, resourceName),
					resource.TestCheckResourceAttrPair(resourceName, "portfolio_id", "aws_servicecatalog_portfolio.test", names.AttrID),
					resource.TestCheckResourceAttr(resourceName, "principal_type", "IAM_PATTERN"),
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
func TestAccServiceCatalogPrincipalPortfolioAssociation_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_servicecatalog_principal_portfolio_association.test"
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ServiceCatalogServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckPrincipalPortfolioAssociationDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccPrincipalPortfolioAssociationConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPrincipalPortfolioAssociationExists(ctx, t, resourceName),
					acctest.CheckSDKResourceDisappears(ctx, t, tfservicecatalog.ResourcePrincipalPortfolioAssociation(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccServiceCatalogPrincipalPortfolioAssociation_migrateV0(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_servicecatalog_principal_portfolio_association.test"
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:   acctest.ErrorCheck(t, names.ServiceCatalogServiceID),
		CheckDestroy: testAccCheckPrincipalPortfolioAssociationDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				ExternalProviders: map[string]resource.ExternalProvider{
					"aws": {
						Source:            "hashicorp/aws",
						VersionConstraint: "5.15.0",
					},
				},
				Config: testAccPrincipalPortfolioAssociationConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					// Can't call this as the old ID format is invalid.
					// testAccCheckPrincipalPortfolioAssociationExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "principal_type", "IAM"),
				),
			},
			{
				ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
				Config:                   testAccPrincipalPortfolioAssociationConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPrincipalPortfolioAssociationExists(ctx, t, resourceName),
				),
			},
		},
	})
}

func testAccCheckPrincipalPortfolioAssociationDestroy(ctx context.Context, t *testing.T) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.ProviderMeta(ctx, t).ServiceCatalogClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_servicecatalog_principal_portfolio_association" {
				continue
			}

			acceptLanguage, principalARN, portfolioID, principalType, err := tfservicecatalog.PrincipalPortfolioAssociationParseResourceID(rs.Primary.ID)
			if err != nil {
				return err
			}

			_, err = tfservicecatalog.FindPrincipalPortfolioAssociation(ctx, conn, acceptLanguage, principalARN, portfolioID, principalType)

			if retry.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("Service Catalog Principal Portfolio Association (%s) still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccCheckPrincipalPortfolioAssociationExists(ctx context.Context, t *testing.T, n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		acceptLanguage, principalARN, portfolioID, principalType, err := tfservicecatalog.PrincipalPortfolioAssociationParseResourceID(rs.Primary.ID)
		if err != nil {
			return err
		}

		conn := acctest.ProviderMeta(ctx, t).ServiceCatalogClient(ctx)

		_, err = tfservicecatalog.FindPrincipalPortfolioAssociation(ctx, conn, acceptLanguage, principalARN, portfolioID, principalType)

		return err
	}
}

func testAccPrincipalPortfolioAssociationConfig_base(rName string) string {
	return fmt.Sprintf(`
data "aws_partition" "current" {}

resource "aws_iam_role" "test" {
  name = %[1]q

  assume_role_policy = jsonencode({
    Version = "2012-10-17"
    Statement = [{
      Action = "sts:AssumeRole"
      Effect = "Allow"
      Principal = {
        Service = "servicecatalog.${data.aws_partition.current.dns_suffix}"
      }
      Sid = ""
    }]
  })
}

resource "aws_servicecatalog_portfolio" "test" {
  name          = %[1]q
  provider_name = %[1]q
}
`, rName)
}

func testAccPrincipalPortfolioAssociationConfig_basic(rName string) string {
	return acctest.ConfigCompose(testAccPrincipalPortfolioAssociationConfig_base(rName), `
resource "aws_servicecatalog_principal_portfolio_association" "test" {
  portfolio_id  = aws_servicecatalog_portfolio.test.id
  principal_arn = aws_iam_role.test.arn
}
`)
}

func testAccPrincipalPortfolioAssociationConfig_iam_pattern(rName string) string {
	return acctest.ConfigCompose(testAccPrincipalPortfolioAssociationConfig_base(rName), `
resource "aws_servicecatalog_principal_portfolio_association" "test" {
  portfolio_id   = aws_servicecatalog_portfolio.test.id
  principal_arn  = "arn:${data.aws_partition.current.partition}:iam:::role/${aws_iam_role.test.name}"
  principal_type = "IAM_PATTERN"
}
`)
}
