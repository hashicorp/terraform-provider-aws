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

// add sweeper to delete known test servicecat budget resource associations

func TestAccServiceCatalogBudgetResourceAssociation_basic(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_servicecatalog_budget_resource_association.test"
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckPartitionHasService(t, "budgets") },
		ErrorCheck:               acctest.ErrorCheck(t, names.ServiceCatalogEndpointID, "budgets"),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckBudgetResourceAssociationDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccBudgetResourceAssociationConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBudgetResourceAssociationExists(ctx, t, resourceName),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrResourceID, "aws_servicecatalog_portfolio.test", names.AttrID),
					resource.TestCheckResourceAttrPair(resourceName, "budget_name", "aws_budgets_budget.test", names.AttrName),
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

func TestAccServiceCatalogBudgetResourceAssociation_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_servicecatalog_budget_resource_association.test"
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckPartitionHasService(t, "budgets") },
		ErrorCheck:               acctest.ErrorCheck(t, names.ServiceCatalogEndpointID, "budgets"),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckBudgetResourceAssociationDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccBudgetResourceAssociationConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBudgetResourceAssociationExists(ctx, t, resourceName),
					acctest.CheckSDKResourceDisappears(ctx, t, tfservicecatalog.ResourceBudgetResourceAssociation(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckBudgetResourceAssociationDestroy(ctx context.Context, t *testing.T) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.ProviderMeta(ctx, t).ServiceCatalogClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_servicecatalog_budget_resource_association" {
				continue
			}

			budgetName, resourceID, err := tfservicecatalog.BudgetResourceAssociationParseID(rs.Primary.ID)

			if err != nil {
				return fmt.Errorf("could not parse ID (%s): %w", rs.Primary.ID, err)
			}

			err = tfservicecatalog.WaitBudgetResourceAssociationDeleted(ctx, conn, budgetName, resourceID, tfservicecatalog.BudgetResourceAssociationDeleteTimeout)

			if retry.NotFound(err) {
				continue
			}

			if err != nil {
				return fmt.Errorf("waiting for Service Catalog Budget Resource Association to be destroyed (%s): %w", rs.Primary.ID, err)
			}
		}

		return nil
	}
}

func testAccCheckBudgetResourceAssociationExists(ctx context.Context, t *testing.T, resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]

		if !ok {
			return fmt.Errorf("resource not found: %s", resourceName)
		}

		budgetName, resourceID, err := tfservicecatalog.BudgetResourceAssociationParseID(rs.Primary.ID)

		if err != nil {
			return fmt.Errorf("could not parse ID (%s): %w", rs.Primary.ID, err)
		}

		conn := acctest.ProviderMeta(ctx, t).ServiceCatalogClient(ctx)

		_, err = tfservicecatalog.WaitBudgetResourceAssociationReady(ctx, conn, budgetName, resourceID, tfservicecatalog.BudgetResourceAssociationReadyTimeout)

		if err != nil {
			return fmt.Errorf("waiting for Service Catalog Budget Resource Association existence (%s): %w", rs.Primary.ID, err)
		}

		return nil
	}
}

func testAccBudgetResourceAssociationConfig_base(rName, budgetType, limitAmount, limitUnit, timePeriodStart, timeUnit string) string {
	return fmt.Sprintf(`
resource "aws_servicecatalog_portfolio" "test" {
  name          = %[1]q
  description   = %[1]q
  provider_name = %[1]q
}

resource "aws_budgets_budget" "test" {
  name              = %[1]q
  budget_type       = %[2]q
  limit_amount      = %[3]q
  limit_unit        = %[4]q
  time_period_start = %[5]q
  time_unit         = %[6]q
}
`, rName, budgetType, limitAmount, limitUnit, timePeriodStart, timeUnit)
}

func testAccBudgetResourceAssociationConfig_basic(rName string) string {
	return acctest.ConfigCompose(testAccBudgetResourceAssociationConfig_base(rName, "COST", "100.0", "USD", "2017-01-01_12:00", "MONTHLY"), fmt.Sprintf(`
resource "aws_servicecatalog_budget_resource_association" "test" {
  resource_id = aws_servicecatalog_portfolio.test.id
  budget_name = %[1]q
}
`, rName))
}
