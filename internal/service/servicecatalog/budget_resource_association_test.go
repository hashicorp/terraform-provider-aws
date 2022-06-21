package servicecatalog_test

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/service/servicecatalog"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfservicecatalog "github.com/hashicorp/terraform-provider-aws/internal/service/servicecatalog"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

// add sweeper to delete known test servicecat budget resource associations

func TestAccServiceCatalogBudgetResourceAssociation_basic(t *testing.T) {
	resourceName := "aws_servicecatalog_budget_resource_association.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); acctest.PreCheckPartitionHasService("budgets", t) },
		ErrorCheck:        acctest.ErrorCheck(t, servicecatalog.EndpointsID, "budgets"),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckBudgetResourceAssociationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccBudgetResourceAssociationConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBudgetResourceAssociationExists(resourceName),
					resource.TestCheckResourceAttrPair(resourceName, "resource_id", "aws_servicecatalog_portfolio.test", "id"),
					resource.TestCheckResourceAttrPair(resourceName, "budget_name", "aws_budgets_budget.test", "name"),
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
	resourceName := "aws_servicecatalog_budget_resource_association.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); acctest.PreCheckPartitionHasService("budgets", t) },
		ErrorCheck:        acctest.ErrorCheck(t, servicecatalog.EndpointsID, "budgets"),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckBudgetResourceAssociationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccBudgetResourceAssociationConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBudgetResourceAssociationExists(resourceName),
					acctest.CheckResourceDisappears(acctest.Provider, tfservicecatalog.ResourceBudgetResourceAssociation(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckBudgetResourceAssociationDestroy(s *terraform.State) error {
	conn := acctest.Provider.Meta().(*conns.AWSClient).ServiceCatalogConn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_servicecatalog_budget_resource_association" {
			continue
		}

		budgetName, resourceID, err := tfservicecatalog.BudgetResourceAssociationParseID(rs.Primary.ID)

		if err != nil {
			return fmt.Errorf("could not parse ID (%s): %w", rs.Primary.ID, err)
		}

		err = tfservicecatalog.WaitBudgetResourceAssociationDeleted(conn, budgetName, resourceID, tfservicecatalog.BudgetResourceAssociationDeleteTimeout)

		if tfresource.NotFound(err) {
			continue
		}

		if err != nil {
			return fmt.Errorf("waiting for Service Catalog Budget Resource Association to be destroyed (%s): %w", rs.Primary.ID, err)
		}
	}

	return nil
}

func testAccCheckBudgetResourceAssociationExists(resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]

		if !ok {
			return fmt.Errorf("resource not found: %s", resourceName)
		}

		budgetName, resourceID, err := tfservicecatalog.BudgetResourceAssociationParseID(rs.Primary.ID)

		if err != nil {
			return fmt.Errorf("could not parse ID (%s): %w", rs.Primary.ID, err)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).ServiceCatalogConn

		_, err = tfservicecatalog.WaitBudgetResourceAssociationReady(conn, budgetName, resourceID, tfservicecatalog.BudgetResourceAssociationReadyTimeout)

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
