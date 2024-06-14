// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package servicecatalog_test

import (
	"context"
	"fmt"
	"testing"

	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfservicecatalog "github.com/hashicorp/terraform-provider-aws/internal/service/servicecatalog"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// add sweeper to delete known test servicecat product portfolio associations

func TestAccServiceCatalogProductPortfolioAssociation_basic(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_servicecatalog_product_portfolio_association.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	domain := fmt.Sprintf("http://%s", acctest.RandomDomainName())

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ServiceCatalogServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckProductPortfolioAssociationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccProductPortfolioAssociationConfig_basic(rName, domain, acctest.DefaultEmailAddress),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckProductPortfolioAssociationExists(ctx, resourceName),
					resource.TestCheckResourceAttrPair(resourceName, "portfolio_id", "aws_servicecatalog_portfolio.test", names.AttrID),
					resource.TestCheckResourceAttrPair(resourceName, "product_id", "aws_servicecatalog_product.test", names.AttrID),
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

func TestAccServiceCatalogProductPortfolioAssociation_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_servicecatalog_product_portfolio_association.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	domain := fmt.Sprintf("http://%s", acctest.RandomDomainName())

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ServiceCatalogServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckProductPortfolioAssociationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccProductPortfolioAssociationConfig_basic(rName, domain, acctest.DefaultEmailAddress),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckProductPortfolioAssociationExists(ctx, resourceName),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfservicecatalog.ResourceProductPortfolioAssociation(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckProductPortfolioAssociationDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).ServiceCatalogConn(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_servicecatalog_product_portfolio_association" {
				continue
			}

			acceptLanguage, portfolioID, productID, err := tfservicecatalog.ProductPortfolioAssociationParseID(rs.Primary.ID)

			if err != nil {
				return fmt.Errorf("could not parse ID (%s): %w", rs.Primary.ID, err)
			}

			err = tfservicecatalog.WaitProductPortfolioAssociationDeleted(ctx, conn, acceptLanguage, portfolioID, productID, tfservicecatalog.ProductPortfolioAssociationDeleteTimeout)

			if tfresource.NotFound(err) {
				continue
			}

			if err != nil {
				return fmt.Errorf("waiting for Service Catalog Product Portfolio Association to be destroyed (%s): %w", rs.Primary.ID, err)
			}
		}

		return nil
	}
}

func testAccCheckProductPortfolioAssociationExists(ctx context.Context, resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]

		if !ok {
			return fmt.Errorf("resource not found: %s", resourceName)
		}

		acceptLanguage, portfolioID, productID, err := tfservicecatalog.ProductPortfolioAssociationParseID(rs.Primary.ID)

		if err != nil {
			return fmt.Errorf("could not parse ID (%s): %w", rs.Primary.ID, err)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).ServiceCatalogConn(ctx)

		_, err = tfservicecatalog.WaitProductPortfolioAssociationReady(ctx, conn, acceptLanguage, portfolioID, productID, tfservicecatalog.ProductPortfolioAssociationReadyTimeout)

		if err != nil {
			return fmt.Errorf("waiting for Service Catalog Product Portfolio Association existence (%s): %w", rs.Primary.ID, err)
		}

		return nil
	}
}

func testAccProductPortfolioAssociationConfig_base(rName, domain, email string) string {
	return fmt.Sprintf(`
resource "aws_cloudformation_stack" "test" {
  name = %[1]q

  template_body = jsonencode({
    AWSTemplateFormatVersion = "2010-09-09"

    Resources = {
      MyVPC = {
        Type = "AWS::EC2::VPC"
        Properties = {
          CidrBlock = "10.1.0.0/16"
        }
      }
    }

    Outputs = {
      VpcID = {
        Description = "VPC ID"
        Value = {
          Ref = "MyVPC"
        }
      }
    }
  })
}

resource "aws_servicecatalog_product" "test" {
  description         = "beskrivning"
  distributor         = "distributör"
  name                = %[1]q
  owner               = "ägare"
  type                = "CLOUD_FORMATION_TEMPLATE"
  support_description = "supportbeskrivning"
  support_email       = %[3]q
  support_url         = %[2]q

  provisioning_artifact_parameters {
    description          = "artefaktbeskrivning"
    name                 = %[1]q
    template_physical_id = aws_cloudformation_stack.test.id
    type                 = "CLOUD_FORMATION_TEMPLATE"
  }

  tags = {
    Name = %[1]q
  }
}

resource "aws_servicecatalog_portfolio" "test" {
  name          = %[1]q
  provider_name = %[1]q
}
`, rName, domain, email)
}

func testAccProductPortfolioAssociationConfig_basic(rName, domain, email string) string {
	return acctest.ConfigCompose(testAccProductPortfolioAssociationConfig_base(rName, domain, email), `
resource "aws_servicecatalog_product_portfolio_association" "test" {
  portfolio_id = aws_servicecatalog_portfolio.test.id
  product_id   = aws_servicecatalog_product.test.id
}
`)
}
