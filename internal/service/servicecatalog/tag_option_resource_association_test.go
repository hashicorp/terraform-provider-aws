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

// add sweeper to delete known test servicecat tag option resource associations

func TestAccServiceCatalogTagOptionResourceAssociation_basic(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_servicecatalog_tag_option_resource_association.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ServiceCatalogServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckTagOptionResourceAssociationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccTagOptionResourceAssociationConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTagOptionResourceAssociationExists(ctx, resourceName),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrResourceID, "aws_servicecatalog_portfolio.test", names.AttrID),
					resource.TestCheckResourceAttrPair(resourceName, "tag_option_id", "aws_servicecatalog_tag_option.test", names.AttrID),
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

func TestAccServiceCatalogTagOptionResourceAssociation_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_servicecatalog_tag_option_resource_association.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ServiceCatalogServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckTagOptionResourceAssociationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccTagOptionResourceAssociationConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTagOptionResourceAssociationExists(ctx, resourceName),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfservicecatalog.ResourceTagOptionResourceAssociation(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckTagOptionResourceAssociationDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).ServiceCatalogConn(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_servicecatalog_tag_option_resource_association" {
				continue
			}

			tagOptionID, resourceID, err := tfservicecatalog.TagOptionResourceAssociationParseID(rs.Primary.ID)

			if err != nil {
				return fmt.Errorf("could not parse ID (%s): %w", rs.Primary.ID, err)
			}

			err = tfservicecatalog.WaitTagOptionResourceAssociationDeleted(ctx, conn, tagOptionID, resourceID, tfservicecatalog.TagOptionResourceAssociationDeleteTimeout)

			if tfresource.NotFound(err) {
				continue
			}

			if err != nil {
				return fmt.Errorf("waiting for Service Catalog Tag Option Resource Association to be destroyed (%s): %w", rs.Primary.ID, err)
			}
		}

		return nil
	}
}

func testAccCheckTagOptionResourceAssociationExists(ctx context.Context, resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]

		if !ok {
			return fmt.Errorf("resource not found: %s", resourceName)
		}

		tagOptionID, resourceID, err := tfservicecatalog.TagOptionResourceAssociationParseID(rs.Primary.ID)

		if err != nil {
			return fmt.Errorf("could not parse ID (%s): %w", rs.Primary.ID, err)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).ServiceCatalogConn(ctx)

		_, err = tfservicecatalog.WaitTagOptionResourceAssociationReady(ctx, conn, tagOptionID, resourceID, tfservicecatalog.TagOptionResourceAssociationReadyTimeout)

		if err != nil {
			return fmt.Errorf("waiting for Service Catalog Tag Option Resource Association existence (%s): %w", rs.Primary.ID, err)
		}

		return nil
	}
}

func testAccTagOptionResourceAssociationConfig_base(rName string) string {
	return fmt.Sprintf(`
resource "aws_servicecatalog_portfolio" "test" {
  name          = %[1]q
  description   = %[1]q
  provider_name = %[1]q
}

resource "aws_servicecatalog_tag_option" "test" {
  key   = %[1]q
  value = %[1]q
}
`, rName)
}

func testAccTagOptionResourceAssociationConfig_basic(rName string) string {
	return acctest.ConfigCompose(testAccTagOptionResourceAssociationConfig_base(rName), `
resource "aws_servicecatalog_tag_option_resource_association" "test" {
  resource_id   = aws_servicecatalog_portfolio.test.id
  tag_option_id = aws_servicecatalog_tag_option.test.id
}
`)
}
