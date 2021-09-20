package servicecatalog_test

import (
	"fmt"
	"log"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/servicecatalog"
	multierror "github.com/hashicorp/go-multierror"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/provider"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	tfservicecatalog "github.com/hashicorp/terraform-provider-aws/internal/service/servicecatalog"
	tfservicecatalog "github.com/hashicorp/terraform-provider-aws/internal/service/servicecatalog"
)

// add sweeper to delete known test servicecat budget resource associations
func init() {
	resource.AddTestSweepers("aws_servicecatalog_budget_resource_association", &resource.Sweeper{
		Name:         "aws_servicecatalog_budget_resource_association",
		Dependencies: []string{},
		F:            testSweepServiceCatalogBudgetResourceAssociations,
	})
}

func testSweepServiceCatalogBudgetResourceAssociations(region string) error {
	client, err := acctest.SharedRegionalSweeperClient(region)

	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}

	conn := client.(*conns.AWSClient).ServiceCatalogConn
	sweepResources := make([]*acctest.SweepResource, 0)
	var errs *multierror.Error

	input := &servicecatalog.ListPortfoliosInput{}

	err = conn.ListPortfoliosPages(input, func(page *servicecatalog.ListPortfoliosOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, port := range page.PortfolioDetails {
			if port == nil {
				continue
			}

			resInput := &servicecatalog.ListBudgetsForResourceInput{
				ResourceId: port.Id,
			}

			err = conn.ListBudgetsForResourcePages(resInput, func(page *servicecatalog.ListBudgetsForResourceOutput, lastPage bool) bool {
				if page == nil {
					return !lastPage
				}

				for _, budget := range page.Budgets {
					if budget == nil {
						continue
					}

					r := ResourceBudgetResourceAssociation()
					d := r.Data(nil)
					d.SetId(tfservicecatalog.BudgetResourceAssociationID(aws.StringValue(budget.BudgetName), aws.StringValue(port.Id)))

					sweepResources = append(sweepResources, acctest.NewSweepResource(r, d, client))
				}

				return !lastPage
			})
		}

		return !lastPage
	})

	if err != nil {
		errs = multierror.Append(errs, fmt.Errorf("error describing Service Catalog Budget Resource (Portfolio) Associations for %s: %w", region, err))
	}

	prodInput := &servicecatalog.SearchProductsAsAdminInput{}

	err = conn.SearchProductsAsAdminPages(prodInput, func(page *servicecatalog.SearchProductsAsAdminOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, pvd := range page.ProductViewDetails {
			if pvd == nil || pvd.ProductViewSummary == nil {
				continue
			}

			resInput := &servicecatalog.ListBudgetsForResourceInput{
				ResourceId: pvd.ProductViewSummary.ProductId,
			}

			err = conn.ListBudgetsForResourcePages(resInput, func(page *servicecatalog.ListBudgetsForResourceOutput, lastPage bool) bool {
				if page == nil {
					return !lastPage
				}

				for _, budget := range page.Budgets {
					if budget == nil {
						continue
					}

					r := ResourceBudgetResourceAssociation()
					d := r.Data(nil)
					d.SetId(tfservicecatalog.BudgetResourceAssociationID(aws.StringValue(budget.BudgetName), aws.StringValue(pvd.ProductViewSummary.ProductId)))

					sweepResources = append(sweepResources, acctest.NewSweepResource(r, d, client))
				}

				return !lastPage
			})
		}

		return !lastPage
	})

	if err != nil {
		errs = multierror.Append(errs, fmt.Errorf("error describing Service Catalog Budget Resource (Product) Associations for %s: %w", region, err))
	}

	if err = acctest.SweepOrchestrator(sweepResources); err != nil {
		errs = multierror.Append(errs, fmt.Errorf("error sweeping Service Catalog Budget Resource Associations for %s: %w", region, err))
	}

	if acctest.SkipSweepError(errs.ErrorOrNil()) {
		log.Printf("[WARN] Skipping Service Catalog Budget Resource Associations sweep for %s: %s", region, errs)
		return nil
	}

	return errs.ErrorOrNil()
}

func TestAccAWSServiceCatalogBudgetResourceAssociation_basic(t *testing.T) {
	resourceName := "aws_servicecatalog_budget_resource_association.test"
	rName := sdkacctest.RandomWithPrefix("tf-acc-test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t); acctest.PreCheckPartitionHasService("budgets", t) },
		ErrorCheck:   acctest.ErrorCheck(t, servicecatalog.EndpointsID, "budgets"),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckAwsServiceCatalogBudgetResourceAssociationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSServiceCatalogBudgetResourceAssociationConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsServiceCatalogBudgetResourceAssociationExists(resourceName),
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

func TestAccAWSServiceCatalogBudgetResourceAssociation_disappears(t *testing.T) {
	resourceName := "aws_servicecatalog_budget_resource_association.test"
	rName := sdkacctest.RandomWithPrefix("tf-acc-test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t); acctest.PreCheckPartitionHasService("budgets", t) },
		ErrorCheck:   acctest.ErrorCheck(t, servicecatalog.EndpointsID, "budgets"),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckAwsServiceCatalogBudgetResourceAssociationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSServiceCatalogBudgetResourceAssociationConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsServiceCatalogBudgetResourceAssociationExists(resourceName),
					acctest.CheckResourceDisappears(acctest.Provider, ResourceBudgetResourceAssociation(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckAwsServiceCatalogBudgetResourceAssociationDestroy(s *terraform.State) error {
	conn := acctest.Provider.Meta().(*conns.AWSClient).ServiceCatalogConn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_servicecatalog_budget_resource_association" {
			continue
		}

		budgetName, resourceID, err := tfservicecatalog.BudgetResourceAssociationParseID(rs.Primary.ID)

		if err != nil {
			return fmt.Errorf("could not parse ID (%s): %w", rs.Primary.ID, err)
		}

		err = tfservicecatalog.WaitBudgetResourceAssociationDeleted(conn, budgetName, resourceID)

		if tfresource.NotFound(err) {
			continue
		}

		if err != nil {
			return fmt.Errorf("waiting for Service Catalog Budget Resource Association to be destroyed (%s): %w", rs.Primary.ID, err)
		}
	}

	return nil
}

func testAccCheckAwsServiceCatalogBudgetResourceAssociationExists(resourceName string) resource.TestCheckFunc {
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

		_, err = tfservicecatalog.WaitBudgetResourceAssociationReady(conn, budgetName, resourceID)

		if err != nil {
			return fmt.Errorf("waiting for Service Catalog Budget Resource Association existence (%s): %w", rs.Primary.ID, err)
		}

		return nil
	}
}

func testAccAWSServiceCatalogBudgetResourceAssociationConfig_base(rName, budgetType, limitAmount, limitUnit, timePeriodStart, timeUnit string) string {
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

func testAccAWSServiceCatalogBudgetResourceAssociationConfig_basic(rName string) string {
	return acctest.ConfigCompose(testAccAWSServiceCatalogBudgetResourceAssociationConfig_base(rName, "COST", "100.0", "USD", "2017-01-01_12:00", "MONTHLY"), fmt.Sprintf(`
resource "aws_servicecatalog_budget_resource_association" "test" {
  resource_id = aws_servicecatalog_portfolio.test.id
  budget_name = %[1]q
}
`, rName))
}
