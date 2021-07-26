package aws

import (
	"fmt"
	"log"
	"reflect"
	"regexp"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/budgets"
	"github.com/hashicorp/go-multierror"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/naming"
	tfbudgets "github.com/terraform-providers/terraform-provider-aws/aws/internal/service/budgets"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/service/budgets/finder"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/tfresource"
)

func init() {
	resource.AddTestSweepers("aws_budgets_budget", &resource.Sweeper{
		Name: "aws_budgets_budget",
		F:    testSweepBudgetsBudgets,
	})
}

func testSweepBudgetsBudgets(region string) error {
	client, err := sharedClientForRegion(region)
	if err != nil {
		return fmt.Errorf("error getting client: %w", err)
	}
	conn := client.(*AWSClient).budgetconn
	accountID := client.(*AWSClient).accountid
	input := &budgets.DescribeBudgetsInput{
		AccountId: aws.String(accountID),
	}
	var sweeperErrs *multierror.Error

	for {
		output, err := conn.DescribeBudgets(input)
		if testSweepSkipSweepError(err) {
			log.Printf("[WARN] Skipping Budgets sweep for %s: %s", region, err)
			return sweeperErrs.ErrorOrNil() // In case we have completed some pages, but had errors
		}
		if err != nil {
			sweeperErrs = multierror.Append(sweeperErrs, fmt.Errorf("error retrieving Budgets: %w", err))
			return sweeperErrs
		}

		for _, budget := range output.Budgets {
			name := aws.StringValue(budget.BudgetName)

			log.Printf("[INFO] Deleting Budget: %s", name)
			_, err := conn.DeleteBudget(&budgets.DeleteBudgetInput{
				AccountId:  aws.String(accountID),
				BudgetName: aws.String(name),
			})
			if isAWSErr(err, budgets.ErrCodeNotFoundException, "") {
				continue
			}
			if err != nil {
				sweeperErr := fmt.Errorf("error deleting Budget (%s): %w", name, err)
				log.Printf("[ERROR] %s", sweeperErr)
				sweeperErrs = multierror.Append(sweeperErrs, sweeperErr)
				continue
			}
		}

		if aws.StringValue(output.NextToken) == "" {
			break
		}
		input.NextToken = output.NextToken
	}

	return sweeperErrs.ErrorOrNil()
}

func TestAccAWSBudgetsBudget_basic(t *testing.T) {
	var budget budgets.Budget
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_budgets_budget.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPartitionHasServicePreCheck(budgets.EndpointsID, t) },
		ErrorCheck:   testAccErrorCheck(t, budgets.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccAWSBudgetsBudgetDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSBudgetsBudgetConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccAWSBudgetsBudgetExists(resourceName, &budget),
					testAccCheckResourceAttrAccountID(resourceName, "account_id"),
					testAccCheckResourceAttrGlobalARN(resourceName, "arn", "budgets", fmt.Sprintf(`budget/%s`, rName)),
					resource.TestCheckResourceAttr(resourceName, "budget_type", "RI_UTILIZATION"),
					resource.TestCheckResourceAttr(resourceName, "cost_filter.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "cost_filter.*", map[string]string{
						"name":     "Service",
						"values.#": "1",
						"values.0": "Amazon Elasticsearch Service",
					}),
					resource.TestCheckResourceAttr(resourceName, "cost_filters.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "cost_filters.Service", "Amazon Elasticsearch Service"),
					resource.TestCheckResourceAttr(resourceName, "limit_amount", "100.0"),
					resource.TestCheckResourceAttr(resourceName, "limit_unit", "PERCENTAGE"),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "notification.#", "0"),
					resource.TestCheckResourceAttrSet(resourceName, "time_period_end"),
					resource.TestCheckResourceAttrSet(resourceName, "time_period_start"),
					resource.TestCheckResourceAttr(resourceName, "time_unit", "QUARTERLY"),
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

func TestAccAWSBudgetsBudget_Name_Generated(t *testing.T) {
	var budget budgets.Budget
	resourceName := "aws_budgets_budget.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPartitionHasServicePreCheck(budgets.EndpointsID, t) },
		ErrorCheck:   testAccErrorCheck(t, budgets.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccAWSBudgetsBudgetDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSBudgetsBudgetConfigNameGenerated(),
				Check: resource.ComposeTestCheckFunc(
					testAccAWSBudgetsBudgetExists(resourceName, &budget),
					resource.TestCheckResourceAttr(resourceName, "budget_type", "RI_COVERAGE"),
					resource.TestCheckResourceAttr(resourceName, "cost_filter.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "cost_filter.*", map[string]string{
						"name":     "Service",
						"values.#": "1",
						"values.0": "Amazon Redshift",
					}),
					resource.TestCheckResourceAttr(resourceName, "cost_filters.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "cost_filters.Service", "Amazon Redshift"),
					resource.TestCheckResourceAttr(resourceName, "limit_amount", "100.0"),
					resource.TestCheckResourceAttr(resourceName, "limit_unit", "PERCENTAGE"),
					naming.TestCheckResourceAttrNameGenerated(resourceName, "name"),
					resource.TestCheckResourceAttr(resourceName, "name_prefix", "terraform-"),
					resource.TestCheckResourceAttr(resourceName, "notification.#", "0"),
					resource.TestCheckResourceAttrSet(resourceName, "time_period_end"),
					resource.TestCheckResourceAttrSet(resourceName, "time_period_start"),
					resource.TestCheckResourceAttr(resourceName, "time_unit", "ANNUALLY"),
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

func TestAccAWSBudgetsBudget_NamePrefix(t *testing.T) {
	var budget budgets.Budget
	resourceName := "aws_budgets_budget.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPartitionHasServicePreCheck(budgets.EndpointsID, t) },
		ErrorCheck:   testAccErrorCheck(t, budgets.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccAWSBudgetsBudgetDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSBudgetsBudgetConfigNamePrefix("tf-acc-test-prefix-"),
				Check: resource.ComposeTestCheckFunc(
					testAccAWSBudgetsBudgetExists(resourceName, &budget),
					resource.TestCheckResourceAttr(resourceName, "budget_type", "SAVINGS_PLANS_UTILIZATION"),
					resource.TestCheckResourceAttr(resourceName, "cost_filter.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "cost_filters.%", "0"),
					resource.TestCheckResourceAttr(resourceName, "limit_amount", "100.0"),
					resource.TestCheckResourceAttr(resourceName, "limit_unit", "PERCENTAGE"),
					naming.TestCheckResourceAttrNameFromPrefix(resourceName, "name", "tf-acc-test-prefix-"),
					resource.TestCheckResourceAttr(resourceName, "name_prefix", "tf-acc-test-prefix-"),
					resource.TestCheckResourceAttr(resourceName, "notification.#", "0"),
					resource.TestCheckResourceAttrSet(resourceName, "time_period_end"),
					resource.TestCheckResourceAttrSet(resourceName, "time_period_start"),
					resource.TestCheckResourceAttr(resourceName, "time_unit", "MONTHLY"),
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

func TestAccAWSBudgetsBudget_disappears(t *testing.T) {
	var budget budgets.Budget
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_budgets_budget.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPartitionHasServicePreCheck(budgets.EndpointsID, t) },
		ErrorCheck:   testAccErrorCheck(t, budgets.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccAWSBudgetsBudgetDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSBudgetsBudgetConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccAWSBudgetsBudgetExists(resourceName, &budget),
					testAccCheckResourceDisappears(testAccProvider, resourceAwsBudgetsBudget(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccAWSBudgetsBudget_basicish(t *testing.T) {
	costFilterKey := "AZ"
	rName := acctest.RandomWithPrefix("tf-acc-test")
	configBasicDefaults := testAccAWSBudgetsBudgetConfigDefaults(rName)
	accountID := "012345678910"
	configBasicUpdate := testAccAWSBudgetsBudgetConfigUpdate(rName)
	resourceName := "aws_budgets_budget.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPartitionHasServicePreCheck(budgets.EndpointsID, t) },
		ErrorCheck:   testAccErrorCheck(t, budgets.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccAWSBudgetsBudgetDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSBudgetsBudgetConfig_BasicDefaults(configBasicDefaults, costFilterKey),
				Check: resource.ComposeTestCheckFunc(
					testAccAWSBudgetsBudgetExistsAndIsValid(resourceName, configBasicDefaults),
					testAccCheckResourceAttrGlobalARN(resourceName, "arn", "budgets", fmt.Sprintf(`budget/%s`, rName)),
					resource.TestMatchResourceAttr(resourceName, "name", regexp.MustCompile(*configBasicDefaults.BudgetName)),
					resource.TestCheckResourceAttr(resourceName, "budget_type", *configBasicDefaults.BudgetType),
					resource.TestCheckResourceAttr(resourceName, "limit_amount", *configBasicDefaults.BudgetLimit.Amount),
					resource.TestCheckResourceAttr(resourceName, "limit_unit", *configBasicDefaults.BudgetLimit.Unit),
					resource.TestCheckResourceAttr(resourceName, "time_period_start", configBasicDefaults.TimePeriod.Start.Format("2006-01-02_15:04")),
					resource.TestCheckResourceAttr(resourceName, "time_period_end", configBasicDefaults.TimePeriod.End.Format("2006-01-02_15:04")),
					resource.TestCheckResourceAttr(resourceName, "time_unit", *configBasicDefaults.TimeUnit),
				),
			},
			{
				PlanOnly:    true,
				Config:      testAccAWSBudgetsBudgetConfig_WithAccountID(configBasicDefaults, accountID, costFilterKey),
				ExpectError: regexp.MustCompile("account_id.*" + accountID),
			},
			{
				Config: testAccAWSBudgetsBudgetConfig_Basic(configBasicUpdate, costFilterKey),
				Check: resource.ComposeTestCheckFunc(
					testAccAWSBudgetsBudgetExistsAndIsValid(resourceName, configBasicUpdate),
					resource.TestMatchResourceAttr(resourceName, "name", regexp.MustCompile(*configBasicUpdate.BudgetName)),
					resource.TestCheckResourceAttr(resourceName, "budget_type", *configBasicUpdate.BudgetType),
					resource.TestCheckResourceAttr(resourceName, "limit_amount", *configBasicUpdate.BudgetLimit.Amount),
					resource.TestCheckResourceAttr(resourceName, "limit_unit", *configBasicUpdate.BudgetLimit.Unit),
					resource.TestCheckResourceAttr(resourceName, "time_period_start", configBasicUpdate.TimePeriod.Start.Format("2006-01-02_15:04")),
					resource.TestCheckResourceAttr(resourceName, "time_period_end", configBasicUpdate.TimePeriod.End.Format("2006-01-02_15:04")),
					resource.TestCheckResourceAttr(resourceName, "time_unit", *configBasicUpdate.TimeUnit),
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

func TestAccAWSBudgetsBudget_prefix(t *testing.T) {
	costFilterKey := "AZ"
	rName := acctest.RandomWithPrefix("tf-acc-test")
	configBasicDefaults := testAccAWSBudgetsBudgetConfigDefaults(rName)
	configBasicUpdate := testAccAWSBudgetsBudgetConfigUpdate(rName)
	resourceName := "aws_budgets_budget.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPartitionHasServicePreCheck(budgets.EndpointsID, t) },
		ErrorCheck:   testAccErrorCheck(t, budgets.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccAWSBudgetsBudgetDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSBudgetsBudgetConfig_PrefixDefaults(configBasicDefaults, costFilterKey),
				Check: resource.ComposeTestCheckFunc(
					testAccAWSBudgetsBudgetExistsAndIsValid(resourceName, configBasicDefaults),
					resource.TestMatchResourceAttr(resourceName, "name_prefix", regexp.MustCompile(*configBasicDefaults.BudgetName)),
					resource.TestCheckResourceAttr(resourceName, "budget_type", *configBasicDefaults.BudgetType),
					resource.TestCheckResourceAttr(resourceName, "limit_amount", *configBasicDefaults.BudgetLimit.Amount),
					resource.TestCheckResourceAttr(resourceName, "limit_unit", *configBasicDefaults.BudgetLimit.Unit),
					resource.TestCheckResourceAttr(resourceName, "time_period_start", configBasicDefaults.TimePeriod.Start.Format("2006-01-02_15:04")),
					resource.TestCheckResourceAttr(resourceName, "time_period_end", configBasicDefaults.TimePeriod.End.Format("2006-01-02_15:04")),
					resource.TestCheckResourceAttr(resourceName, "time_unit", *configBasicDefaults.TimeUnit),
				),
			},

			{
				Config: testAccAWSBudgetsBudgetConfig_Prefix(configBasicUpdate, costFilterKey),
				Check: resource.ComposeTestCheckFunc(
					testAccAWSBudgetsBudgetExistsAndIsValid(resourceName, configBasicUpdate),
					resource.TestMatchResourceAttr(resourceName, "name_prefix", regexp.MustCompile(*configBasicUpdate.BudgetName)),
					resource.TestCheckResourceAttr(resourceName, "budget_type", *configBasicUpdate.BudgetType),
					resource.TestCheckResourceAttr(resourceName, "limit_amount", *configBasicUpdate.BudgetLimit.Amount),
					resource.TestCheckResourceAttr(resourceName, "limit_unit", *configBasicUpdate.BudgetLimit.Unit),
					resource.TestCheckResourceAttr(resourceName, "time_period_start", configBasicUpdate.TimePeriod.Start.Format("2006-01-02_15:04")),
					resource.TestCheckResourceAttr(resourceName, "time_period_end", configBasicUpdate.TimePeriod.End.Format("2006-01-02_15:04")),
					resource.TestCheckResourceAttr(resourceName, "time_unit", *configBasicUpdate.TimeUnit),
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

func TestAccAWSBudgetsBudget_notification(t *testing.T) {
	rName := acctest.RandomWithPrefix("tf-acc-test")
	configBasicDefaults := testAccAWSBudgetsBudgetConfigDefaults(rName)
	configBasicDefaults.CostFilters = map[string][]*string{}
	resourceName := "aws_budgets_budget.test"

	notificationConfigDefaults := []budgets.Notification{testAccAWSBudgetsBudgetNotificationConfigDefaults()}
	notificationConfigUpdated := []budgets.Notification{testAccAWSBudgetsBudgetNotificationConfigUpdate()}
	twoNotificationConfigs := []budgets.Notification{
		testAccAWSBudgetsBudgetNotificationConfigUpdate(),
		testAccAWSBudgetsBudgetNotificationConfigDefaults(),
	}

	domain := testAccRandomDomainName()
	address1 := testAccRandomEmailAddress(domain)
	address2 := testAccRandomEmailAddress(domain)
	address3 := testAccRandomEmailAddress(domain)

	noEmails := []string{}
	oneEmail := []string{address1}
	oneOtherEmail := []string{address2}
	twoEmails := []string{address2, address3}
	noTopics := []string{}
	oneTopic := []string{"${aws_sns_topic.budget_notifications.arn}"}

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPartitionHasServicePreCheck(budgets.EndpointsID, t) },
		ErrorCheck:   testAccErrorCheck(t, budgets.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccAWSBudgetsBudgetDestroy,
		Steps: []resource.TestStep{
			// Can't create without at least one subscriber
			{
				Config:      testAccAWSBudgetsBudgetConfigWithNotification_Basic(configBasicDefaults, notificationConfigDefaults, noEmails, noTopics),
				ExpectError: regexp.MustCompile(`Notification must have at least one subscriber`),
				Check: resource.ComposeTestCheckFunc(
					testAccAWSBudgetsBudgetExistsAndIsValid(resourceName, configBasicDefaults),
				),
			},
			// Basic Notification with only email
			{
				Config: testAccAWSBudgetsBudgetConfigWithNotification_Basic(configBasicDefaults, notificationConfigDefaults, oneEmail, noTopics),
				Check: resource.ComposeTestCheckFunc(
					testAccAWSBudgetsBudgetExistsAndIsValid(resourceName, configBasicDefaults),
				),
			},
			// Change only subscriber to a different e-mail
			{
				Config: testAccAWSBudgetsBudgetConfigWithNotification_Basic(configBasicDefaults, notificationConfigDefaults, oneOtherEmail, noTopics),
				Check: resource.ComposeTestCheckFunc(
					testAccAWSBudgetsBudgetExistsAndIsValid(resourceName, configBasicDefaults),
				),
			},
			// Add a second e-mail and a topic
			{
				Config: testAccAWSBudgetsBudgetConfigWithNotification_Basic(configBasicDefaults, notificationConfigDefaults, twoEmails, oneTopic),
				Check: resource.ComposeTestCheckFunc(
					testAccAWSBudgetsBudgetExistsAndIsValid(resourceName, configBasicDefaults),
				),
			},
			// Delete both E-Mails
			{
				Config: testAccAWSBudgetsBudgetConfigWithNotification_Basic(configBasicDefaults, notificationConfigDefaults, noEmails, oneTopic),
				Check: resource.ComposeTestCheckFunc(
					testAccAWSBudgetsBudgetExistsAndIsValid(resourceName, configBasicDefaults),
				),
			},
			// Swap one Topic fo one E-Mail
			{
				Config: testAccAWSBudgetsBudgetConfigWithNotification_Basic(configBasicDefaults, notificationConfigDefaults, oneEmail, noTopics),
				Check: resource.ComposeTestCheckFunc(
					testAccAWSBudgetsBudgetExistsAndIsValid(resourceName, configBasicDefaults),
				),
			},
			// Can't update without at least one subscriber
			{
				Config:      testAccAWSBudgetsBudgetConfigWithNotification_Basic(configBasicDefaults, notificationConfigDefaults, noEmails, noTopics),
				ExpectError: regexp.MustCompile(`Notification must have at least one subscriber`),
				Check: resource.ComposeTestCheckFunc(
					testAccAWSBudgetsBudgetExistsAndIsValid(resourceName, configBasicDefaults),
				),
			},
			// Update all non-subscription parameters
			{
				Config:      testAccAWSBudgetsBudgetConfigWithNotification_Basic(configBasicDefaults, notificationConfigUpdated, noEmails, noTopics),
				ExpectError: regexp.MustCompile(`Notification must have at least one subscriber`),
				Check: resource.ComposeTestCheckFunc(
					testAccAWSBudgetsBudgetExistsAndIsValid(resourceName, configBasicDefaults),
				),
			},
			// Add a second subscription
			{
				Config:      testAccAWSBudgetsBudgetConfigWithNotification_Basic(configBasicDefaults, twoNotificationConfigs, noEmails, noTopics),
				ExpectError: regexp.MustCompile(`Notification must have at least one subscriber`),
				Check: resource.ComposeTestCheckFunc(
					testAccAWSBudgetsBudgetExistsAndIsValid(resourceName, configBasicDefaults),
				),
			},
		},
	})
}

func testAccAWSBudgetsBudgetExists(resourceName string, v *budgets.Budget) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("Not found: %s", resourceName)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No Budget ID is set")
		}

		conn := testAccProvider.Meta().(*AWSClient).budgetconn

		accountID, budgetName, err := tfbudgets.BudgetParseResourceID(rs.Primary.ID)

		if err != nil {
			return err
		}

		output, err := finder.BudgetByAccountIDAndBudgetName(conn, accountID, budgetName)

		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

func testAccAWSBudgetsBudgetExistsAndIsValid(resourceName string, config budgets.Budget) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("Not found: %s", resourceName)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No Budget ID is set")
		}

		conn := testAccProvider.Meta().(*AWSClient).budgetconn

		accountID, budgetName, err := tfbudgets.BudgetParseResourceID(rs.Primary.ID)

		if err != nil {
			return err
		}

		output, err := finder.BudgetByAccountIDAndBudgetName(conn, accountID, budgetName)

		if err != nil {
			return err
		}

		if aws.StringValue(output.BudgetLimit.Amount) != aws.StringValue(config.BudgetLimit.Amount) {
			return fmt.Errorf("budget limit incorrectly set %v != %v", aws.StringValue(config.BudgetLimit.Amount),
				aws.StringValue(output.BudgetLimit.Amount))
		}

		if err := testAccAWSBudgetsBudgetCheckCostTypes(config, *output.CostTypes); err != nil {
			return err
		}

		if err := testAccAWSBudgetsBudgetCheckTimePeriod(*config.TimePeriod, *output.TimePeriod); err != nil {
			return err
		}

		if !reflect.DeepEqual(output.CostFilters, config.CostFilters) {
			return fmt.Errorf("cost filter not set properly: %v != %v", output.CostFilters, config.CostFilters)
		}

		return nil
	}
}

func testAccAWSBudgetsBudgetCheckTimePeriod(configTimePeriod, timePeriod budgets.TimePeriod) error {
	if configTimePeriod.End.Format("2006-01-02_15:04") != timePeriod.End.Format("2006-01-02_15:04") {
		return fmt.Errorf("TimePeriodEnd not set properly '%v' should be '%v'", *timePeriod.End, *configTimePeriod.End)
	}

	if configTimePeriod.Start.Format("2006-01-02_15:04") != timePeriod.Start.Format("2006-01-02_15:04") {
		return fmt.Errorf("TimePeriodStart not set properly '%v' should be '%v'", *timePeriod.Start, *configTimePeriod.Start)
	}

	return nil
}

func testAccAWSBudgetsBudgetCheckCostTypes(config budgets.Budget, costTypes budgets.CostTypes) error {
	if *costTypes.IncludeCredit != *config.CostTypes.IncludeCredit {
		return fmt.Errorf("IncludeCredit not set properly '%v' should be '%v'", *costTypes.IncludeCredit, *config.CostTypes.IncludeCredit)
	}

	if *costTypes.IncludeOtherSubscription != *config.CostTypes.IncludeOtherSubscription {
		return fmt.Errorf("IncludeOtherSubscription not set properly '%v' should be '%v'", *costTypes.IncludeOtherSubscription, *config.CostTypes.IncludeOtherSubscription)
	}

	if *costTypes.IncludeRecurring != *config.CostTypes.IncludeRecurring {
		return fmt.Errorf("IncludeRecurring not set properly  '%v' should be '%v'", *costTypes.IncludeRecurring, *config.CostTypes.IncludeRecurring)
	}

	if *costTypes.IncludeRefund != *config.CostTypes.IncludeRefund {
		return fmt.Errorf("IncludeRefund not set properly '%v' should be '%v'", *costTypes.IncludeRefund, *config.CostTypes.IncludeRefund)
	}

	if *costTypes.IncludeSubscription != *config.CostTypes.IncludeSubscription {
		return fmt.Errorf("IncludeSubscription not set properly '%v' should be '%v'", *costTypes.IncludeSubscription, *config.CostTypes.IncludeSubscription)
	}

	if *costTypes.IncludeSupport != *config.CostTypes.IncludeSupport {
		return fmt.Errorf("IncludeSupport not set properly '%v' should be '%v'", *costTypes.IncludeSupport, *config.CostTypes.IncludeSupport)
	}

	if *costTypes.IncludeTax != *config.CostTypes.IncludeTax {
		return fmt.Errorf("IncludeTax not set properly '%v' should be '%v'", *costTypes.IncludeTax, *config.CostTypes.IncludeTax)
	}

	if *costTypes.IncludeUpfront != *config.CostTypes.IncludeUpfront {
		return fmt.Errorf("IncludeUpfront not set properly '%v' should be '%v'", *costTypes.IncludeUpfront, *config.CostTypes.IncludeUpfront)
	}

	if *costTypes.UseBlended != *config.CostTypes.UseBlended {
		return fmt.Errorf("UseBlended not set properly '%v' should be '%v'", *costTypes.UseBlended, *config.CostTypes.UseBlended)
	}

	return nil
}

func testAccAWSBudgetsBudgetDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*AWSClient).budgetconn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_budgets_budget" {
			continue
		}

		accountID, budgetName, err := tfbudgets.BudgetParseResourceID(rs.Primary.ID)

		if err != nil {
			return err
		}

		_, err = finder.BudgetByAccountIDAndBudgetName(conn, accountID, budgetName)

		if tfresource.NotFound(err) {
			continue
		}

		if err != nil {
			return err
		}

		return fmt.Errorf("Budget Action %s still exists", rs.Primary.ID)
	}

	return nil
}

func testAccAWSBudgetsBudgetConfigUpdate(name string) budgets.Budget {
	dateNow := time.Now().UTC()
	futureDate := dateNow.AddDate(0, 0, 14)
	startDate := dateNow.AddDate(0, 0, -14)
	return budgets.Budget{
		BudgetName: aws.String(name),
		BudgetType: aws.String("COST"),
		BudgetLimit: &budgets.Spend{
			Amount: aws.String("500.0"),
			Unit:   aws.String("USD"),
		},
		CostFilters: map[string][]*string{
			"AZ": {
				aws.String(testAccGetAlternateRegion()),
				aws.String(testAccGetThirdRegion()),
			},
		},
		CostTypes: &budgets.CostTypes{
			IncludeCredit:            aws.Bool(true),
			IncludeOtherSubscription: aws.Bool(true),
			IncludeRecurring:         aws.Bool(true),
			IncludeRefund:            aws.Bool(true),
			IncludeSubscription:      aws.Bool(false),
			IncludeSupport:           aws.Bool(true),
			IncludeTax:               aws.Bool(false),
			IncludeUpfront:           aws.Bool(true),
			UseBlended:               aws.Bool(false),
		},
		TimeUnit: aws.String("MONTHLY"),
		TimePeriod: &budgets.TimePeriod{
			End:   &futureDate,
			Start: &startDate,
		},
	}
}

func testAccAWSBudgetsBudgetConfigDefaults(name string) budgets.Budget {
	dateNow := time.Now().UTC()
	futureDate := time.Date(2087, 6, 15, 00, 0, 0, 0, time.UTC)
	startDate := dateNow.AddDate(0, 0, -14)
	return budgets.Budget{
		BudgetName: aws.String(name),
		BudgetType: aws.String("COST"),
		BudgetLimit: &budgets.Spend{
			Amount: aws.String("100.0"),
			Unit:   aws.String("USD"),
		},
		CostFilters: map[string][]*string{
			"AZ": {
				aws.String(testAccGetRegion()),
				aws.String(testAccGetThirdRegion()),
			},
		},
		CostTypes: &budgets.CostTypes{
			IncludeCredit:            aws.Bool(true),
			IncludeOtherSubscription: aws.Bool(true),
			IncludeRecurring:         aws.Bool(true),
			IncludeRefund:            aws.Bool(true),
			IncludeSubscription:      aws.Bool(true),
			IncludeSupport:           aws.Bool(true),
			IncludeTax:               aws.Bool(true),
			IncludeUpfront:           aws.Bool(true),
			UseBlended:               aws.Bool(false),
		},
		TimeUnit: aws.String("MONTHLY"),
		TimePeriod: &budgets.TimePeriod{
			End:   &futureDate,
			Start: &startDate,
		},
	}
}

func testAccAWSBudgetsBudgetNotificationConfigDefaults() budgets.Notification {
	return budgets.Notification{
		NotificationType:   aws.String(budgets.NotificationTypeActual),
		ThresholdType:      aws.String(budgets.ThresholdTypeAbsoluteValue),
		Threshold:          aws.Float64(100.0),
		ComparisonOperator: aws.String(budgets.ComparisonOperatorGreaterThan),
	}
}

func testAccAWSBudgetsBudgetNotificationConfigUpdate() budgets.Notification {
	return budgets.Notification{
		NotificationType:   aws.String(budgets.NotificationTypeForecasted),
		ThresholdType:      aws.String(budgets.ThresholdTypePercentage),
		Threshold:          aws.Float64(200.0),
		ComparisonOperator: aws.String(budgets.ComparisonOperatorLessThan),
	}
}

func testAccAWSBudgetsBudgetConfig(rName string) string {
	return fmt.Sprintf(`
resource "aws_budgets_budget" "test" {
  name         = %[1]q
  budget_type  = "RI_UTILIZATION"
  limit_amount = "100.0"
  limit_unit   = "PERCENTAGE"
  time_unit    = "QUARTERLY"

  cost_filters = {
    Service = "Amazon Elasticsearch Service"
  }
}
`, rName)
}

func testAccAWSBudgetsBudgetConfigNameGenerated() string {
	return `
resource "aws_budgets_budget" "test" {
  budget_type  = "RI_COVERAGE"
  limit_amount = "100.00"
  limit_unit   = "PERCENTAGE"
  time_unit    = "ANNUALLY"

  cost_filter {
    name   = "Service"
    values = ["Amazon Redshift"]
  }
}
`
}

func testAccAWSBudgetsBudgetConfigNamePrefix(namePrefix string) string {
	return fmt.Sprintf(`
resource "aws_budgets_budget" "test" {
  name_prefix  = %[1]q
  budget_type  = "SAVINGS_PLANS_UTILIZATION"
  limit_amount = "100"
  limit_unit   = "PERCENTAGE"
  time_unit    = "MONTHLY"
}
`, namePrefix)
}

func testAccAWSBudgetsBudgetConfig_WithAccountID(budgetConfig budgets.Budget, accountID, costFilterKey string) string {
	timePeriodStart := budgetConfig.TimePeriod.Start.Format("2006-01-02_15:04")
	costFilterValue1 := aws.StringValue(budgetConfig.CostFilters[costFilterKey][0])
	costFilterValue2 := aws.StringValue(budgetConfig.CostFilters[costFilterKey][1])

	return fmt.Sprintf(`
resource "aws_budgets_budget" "test" {
  account_id        = "%s"
  name_prefix       = "%s"
  budget_type       = "%s"
  limit_amount      = "%s"
  limit_unit        = "%s"
  time_period_start = "%s"
  time_unit         = "%s"

  cost_filter {
    name = "%s"
    values = [
      "%s",
      "%s",
    ]
  }
}
`, accountID, aws.StringValue(budgetConfig.BudgetName), aws.StringValue(budgetConfig.BudgetType), aws.StringValue(budgetConfig.BudgetLimit.Amount), aws.StringValue(budgetConfig.BudgetLimit.Unit), timePeriodStart, aws.StringValue(budgetConfig.TimeUnit), costFilterKey, costFilterValue1, costFilterValue2)
}

func testAccAWSBudgetsBudgetConfig_PrefixDefaults(budgetConfig budgets.Budget, costFilterKey string) string {
	timePeriodStart := budgetConfig.TimePeriod.Start.Format("2006-01-02_15:04")
	costFilterValue1 := aws.StringValue(budgetConfig.CostFilters[costFilterKey][0])
	costFilterValue2 := aws.StringValue(budgetConfig.CostFilters[costFilterKey][1])

	return fmt.Sprintf(`
resource "aws_budgets_budget" "test" {
  name_prefix       = "%s"
  budget_type       = "%s"
  limit_amount      = "%s"
  limit_unit        = "%s"
  time_period_start = "%s"
  time_unit         = "%s"

  cost_filter {
    name = "%s"
    values = [
      "%s",
      "%s",
    ]
  }
}
`, aws.StringValue(budgetConfig.BudgetName), aws.StringValue(budgetConfig.BudgetType), aws.StringValue(budgetConfig.BudgetLimit.Amount), aws.StringValue(budgetConfig.BudgetLimit.Unit), timePeriodStart, aws.StringValue(budgetConfig.TimeUnit), costFilterKey, costFilterValue1, costFilterValue2)
}

func testAccAWSBudgetsBudgetConfig_Prefix(budgetConfig budgets.Budget, costFilterKey string) string {
	timePeriodStart := budgetConfig.TimePeriod.Start.Format("2006-01-02_15:04")
	timePeriodEnd := budgetConfig.TimePeriod.End.Format("2006-01-02_15:04")
	costFilterValue1 := aws.StringValue(budgetConfig.CostFilters[costFilterKey][0])
	costFilterValue2 := aws.StringValue(budgetConfig.CostFilters[costFilterKey][1])

	return fmt.Sprintf(`
resource "aws_budgets_budget" "test" {
  name_prefix  = "%s"
  budget_type  = "%s"
  limit_amount = "%s"
  limit_unit   = "%s"

  cost_types {
    include_tax          = "%t"
    include_subscription = "%t"
    use_blended          = "%t"
  }

  time_period_start = "%s"
  time_period_end   = "%s"
  time_unit         = "%s"

  cost_filter {
    name = "%s"
    values = [
      "%s",
      "%s",
    ]
  }
}
`, aws.StringValue(budgetConfig.BudgetName), aws.StringValue(budgetConfig.BudgetType), aws.StringValue(budgetConfig.BudgetLimit.Amount), aws.StringValue(budgetConfig.BudgetLimit.Unit), aws.BoolValue(budgetConfig.CostTypes.IncludeTax), aws.BoolValue(budgetConfig.CostTypes.IncludeSubscription), aws.BoolValue(budgetConfig.CostTypes.UseBlended), timePeriodStart, timePeriodEnd, aws.StringValue(budgetConfig.TimeUnit), costFilterKey, costFilterValue1, costFilterValue2)
}

func testAccAWSBudgetsBudgetConfig_BasicDefaults(budgetConfig budgets.Budget, costFilterKey string) string {
	timePeriodStart := budgetConfig.TimePeriod.Start.Format("2006-01-02_15:04")
	costFilterValue1 := aws.StringValue(budgetConfig.CostFilters[costFilterKey][0])
	costFilterValue2 := aws.StringValue(budgetConfig.CostFilters[costFilterKey][1])

	return fmt.Sprintf(`
resource "aws_budgets_budget" "test" {
  name              = "%s"
  budget_type       = "%s"
  limit_amount      = "%s"
  limit_unit        = "%s"
  time_period_start = "%s"
  time_unit         = "%s"

  cost_filter {
    name = "%s"
    values = [
      "%s",
      "%s",
    ]
  }
}
`, aws.StringValue(budgetConfig.BudgetName), aws.StringValue(budgetConfig.BudgetType), aws.StringValue(budgetConfig.BudgetLimit.Amount), aws.StringValue(budgetConfig.BudgetLimit.Unit), timePeriodStart, aws.StringValue(budgetConfig.TimeUnit), costFilterKey, costFilterValue1, costFilterValue2)
}

func testAccAWSBudgetsBudgetConfig_Basic(budgetConfig budgets.Budget, costFilterKey string) string {
	timePeriodStart := budgetConfig.TimePeriod.Start.Format("2006-01-02_15:04")
	timePeriodEnd := budgetConfig.TimePeriod.End.Format("2006-01-02_15:04")
	costFilterValue1 := aws.StringValue(budgetConfig.CostFilters[costFilterKey][0])
	costFilterValue2 := aws.StringValue(budgetConfig.CostFilters[costFilterKey][1])

	return fmt.Sprintf(`
resource "aws_budgets_budget" "test" {
  name         = "%s"
  budget_type  = "%s"
  limit_amount = "%s"
  limit_unit   = "%s"

  cost_types {
    include_tax          = "%t"
    include_subscription = "%t"
    use_blended          = "%t"
  }

  time_period_start = "%s"
  time_period_end   = "%s"
  time_unit         = "%s"

  cost_filter {
    name = "%s"
    values = [
      "%s",
      "%s",
    ]
  }
}
`, aws.StringValue(budgetConfig.BudgetName), aws.StringValue(budgetConfig.BudgetType), aws.StringValue(budgetConfig.BudgetLimit.Amount), aws.StringValue(budgetConfig.BudgetLimit.Unit), aws.BoolValue(budgetConfig.CostTypes.IncludeTax), aws.BoolValue(budgetConfig.CostTypes.IncludeSubscription), aws.BoolValue(budgetConfig.CostTypes.UseBlended), timePeriodStart, timePeriodEnd, aws.StringValue(budgetConfig.TimeUnit), costFilterKey, costFilterValue1, costFilterValue2)
}

func testAccAWSBudgetsBudgetConfigWithNotification_Basic(budgetConfig budgets.Budget, notifications []budgets.Notification, emails []string, topics []string) string {
	timePeriodStart := budgetConfig.TimePeriod.Start.Format("2006-01-02_15:04")
	timePeriodEnd := budgetConfig.TimePeriod.End.Format("2006-01-02_15:04")
	notificationStrings := make([]string, len(notifications))

	for i, notification := range notifications {
		notificationStrings[i] = testAccAWSBudgetsBudgetConfigNotificationSnippet(notification, emails, topics)
	}

	return fmt.Sprintf(`
resource "aws_sns_topic" "budget_notifications" {
  name_prefix = "user-updates-topic"
}

resource "aws_budgets_budget" "test" {
  name         = "%s"
  budget_type  = "%s"
  limit_amount = "%s"
  limit_unit   = "%s"
  cost_types {
    include_tax          = "%t"
    include_subscription = "%t"
    use_blended          = "%t"
  }

  time_period_start = "%s"
  time_period_end   = "%s"
  time_unit         = "%s"
    %s
}
`, aws.StringValue(budgetConfig.BudgetName),
		aws.StringValue(budgetConfig.BudgetType),
		aws.StringValue(budgetConfig.BudgetLimit.Amount),
		aws.StringValue(budgetConfig.BudgetLimit.Unit),
		aws.BoolValue(budgetConfig.CostTypes.IncludeTax),
		aws.BoolValue(budgetConfig.CostTypes.IncludeSubscription),
		aws.BoolValue(budgetConfig.CostTypes.UseBlended),
		timePeriodStart,
		timePeriodEnd,
		aws.StringValue(budgetConfig.TimeUnit),
		strings.Join(notificationStrings, "\n"))
}

func testAccAWSBudgetsBudgetConfigNotificationSnippet(notification budgets.Notification, emails []string, topics []string) string {
	quotedEMails := make([]string, len(emails))
	for i, email := range emails {
		quotedEMails[i] = strconv.Quote(email)
	}

	quotedTopics := make([]string, len(topics))
	for i, topic := range topics {
		quotedTopics[i] = strconv.Quote(topic)
	}

	return fmt.Sprintf(`
notification {
  threshold                  = %f
  threshold_type             = "%s"
  notification_type          = "%s"
  subscriber_email_addresses = [%s]
  subscriber_sns_topic_arns  = [%s]
  comparison_operator        = "%s"
}
`, aws.Float64Value(notification.Threshold),
		aws.StringValue(notification.ThresholdType),
		aws.StringValue(notification.NotificationType),
		strings.Join(quotedEMails, ","),
		strings.Join(quotedTopics, ","),
		aws.StringValue(notification.ComparisonOperator))
}
