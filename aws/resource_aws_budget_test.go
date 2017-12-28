package aws

import (
	"bytes"
	"fmt"
	"regexp"
	"strconv"
	"testing"
	"text/template"
	"time"

	"github.com/aws/aws-sdk-go/service/budgets"
	"github.com/hashicorp/terraform/helper/acctest"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/helper/schema"
	"github.com/hashicorp/terraform/terraform"
)

func TestAwsBudget_basic(t *testing.T) {
	name := fmt.Sprintf("test-budget-%d", acctest.RandInt())
	configBasic := budgetTestConfig{
		BudgetName:               name,
		BudgetType:               "COST",
		LimitAmount:              "100",
		LimitUnit:                "USD",
		FilterKey:                "AZ",
		FilterValue:              "us-east-1",
		IncludeCredit:            "false",
		IncludeOtherSubscription: "false",
		IncludeRecurring:         "false",
		IncludeRefund:            "false",
		IncludeSubscription:      "false",
		IncludeSupport:           "false",
		IncludeTax:               "false",
		IncludeUpfront:           "false",
		UseBlended:               "false",
		TimeUnit:                 "MONTHLY",
		TimePeriodStart:          "2017-01-01_12:00",
		TimePeriodEnd:            "2087-06-15_12:00",
	}
	configBasicUpdate := budgetTestConfig{
		BudgetName:               name,
		BudgetType:               "COST",
		LimitAmount:              "500",
		LimitUnit:                "USD",
		FilterKey:                "AZ",
		FilterValue:              "us-east-2",
		IncludeCredit:            "false",
		IncludeOtherSubscription: "false",
		IncludeRecurring:         "false",
		IncludeRefund:            "false",
		IncludeSubscription:      "false",
		IncludeSupport:           "false",
		IncludeTax:               "true",
		IncludeUpfront:           "false",
		UseBlended:               "false",
		TimeUnit:                 "MONTHLY",
		TimePeriodStart:          "2017-01-01_12:00",
		TimePeriodEnd:            "2018-01-01_12:00",
	}

	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		CheckDestroy: func(s *terraform.State) error {
			return testCheckBudgetDestroy(name, testAccProvider)
		},
		Steps: []resource.TestStep{
			{
				Config: testBudgetHCLBasicUseDefaults(configBasic),
				Check:  newComposedBudgetTestCheck(configBasic, testAccProvider, "name"),
			},

			{
				Config: testBudgetHCLBasic(configBasicUpdate),
				Check:  newComposedBudgetTestCheck(configBasicUpdate, testAccProvider, "name"),
			},
		},
	})
}

func TestAwsBudget_prefix(t *testing.T) {
	name := "test-budget-"
	configBasic := budgetTestConfig{
		BudgetName:               name,
		BudgetType:               "COST",
		LimitAmount:              "100",
		LimitUnit:                "USD",
		FilterKey:                "AZ",
		FilterValue:              "us-east-1",
		IncludeCredit:            "false",
		IncludeOtherSubscription: "false",
		IncludeRecurring:         "false",
		IncludeRefund:            "false",
		IncludeSubscription:      "false",
		IncludeSupport:           "false",
		IncludeTax:               "false",
		IncludeUpfront:           "false",
		UseBlended:               "false",
		TimeUnit:                 "MONTHLY",
		TimePeriodStart:          "2017-01-01_12:00",
		TimePeriodEnd:            "2087-06-15_12:00",
	}

	configBasicUpdate := budgetTestConfig{
		BudgetName:               name,
		BudgetType:               "COST",
		LimitAmount:              "500",
		LimitUnit:                "USD",
		FilterKey:                "AZ",
		FilterValue:              "us-east-2",
		IncludeCredit:            "false",
		IncludeOtherSubscription: "false",
		IncludeRecurring:         "false",
		IncludeRefund:            "false",
		IncludeSubscription:      "false",
		IncludeSupport:           "false",
		IncludeTax:               "true",
		IncludeUpfront:           "false",
		UseBlended:               "false",
		TimeUnit:                 "MONTHLY",
		TimePeriodStart:          "2017-01-01_12:00",
		TimePeriodEnd:            "2018-01-01_12:00",
	}

	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		CheckDestroy: func(s *terraform.State) error {
			return testCheckBudgetDestroy(name, testAccProvider)
		},
		Steps: []resource.TestStep{
			{
				Config: testBudgetHCLPrefixUseDefaults(configBasic),
				Check:  newComposedBudgetTestCheck(configBasic, testAccProvider, "name_prefix"),
			},

			{
				Config: testBudgetHCLPrefix(configBasicUpdate),
				Check:  newComposedBudgetTestCheck(configBasicUpdate, testAccProvider, "name_prefix"),
			},
		},
	})
}

func newComposedBudgetTestCheck(config budgetTestConfig, provider *schema.Provider, nameField string) resource.TestCheckFunc {
	return resource.ComposeTestCheckFunc(
		resource.TestMatchResourceAttr("aws_budget.foo", nameField, regexp.MustCompile(config.BudgetName)),
		resource.TestCheckResourceAttr("aws_budget.foo", "budget_type", config.BudgetType),
		resource.TestCheckResourceAttr("aws_budget.foo", "limit_amount", config.LimitAmount),
		resource.TestCheckResourceAttr("aws_budget.foo", "limit_unit", config.LimitUnit),
		resource.TestCheckResourceAttr("aws_budget.foo", "time_period_start", config.TimePeriodStart),
		resource.TestCheckResourceAttr("aws_budget.foo", "time_period_end", config.TimePeriodEnd),
		resource.TestCheckResourceAttr("aws_budget.foo", "time_unit", config.TimeUnit),
		testBudgetExists(config, provider),
	)
}

func testBudgetExists(config budgetTestConfig, provider *schema.Provider) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources["aws_budget.foo"]
		if !ok {
			return fmt.Errorf("Not found: %s", "aws_budget.foo")
		}

		b, err := describeBudget(rs.Primary.ID, provider.Meta())
		if err != nil {
			return fmt.Errorf("Describebudget error: %v", err)
		}

		if b.Budget == nil {
			return fmt.Errorf("No budget returned %v in %v", b.Budget, b)
		}

		if *b.Budget.BudgetLimit.Amount != config.LimitAmount {
			return fmt.Errorf("budget limit incorrectly set %v != %v", config.LimitAmount, *b.Budget.BudgetLimit.Amount)
		}

		if err := checkBudgetCostTypes(config, *b.Budget.CostTypes); err != nil {
			return err
		}

		if err := checkBudgetTimePeriod(config, *b.Budget.TimePeriod); err != nil {
			return err
		}

		if v, ok := b.Budget.CostFilters[config.FilterKey]; !ok || *v[len(v)-1] != config.FilterValue {
			return fmt.Errorf("cost filter not set properly: %v", b.Budget.CostFilters)
		}

		return nil
	}
}

func checkBudgetTimePeriod(config budgetTestConfig, timePeriod budgets.TimePeriod) error {
	if end, _ := time.Parse("2006-01-02_15:04", config.TimePeriodEnd); *timePeriod.End != end {
		return fmt.Errorf("TimePeriodEnd not set properly '%v' should be '%v'", *timePeriod.End, end)
	}

	if start, _ := time.Parse("2006-01-02_15:04", config.TimePeriodStart); *timePeriod.Start != start {
		return fmt.Errorf("TimePeriodStart not set properly '%v' should be '%v'", *timePeriod.Start, start)
	}

	return nil
}

func checkBudgetCostTypes(config budgetTestConfig, costTypes budgets.CostTypes) error {
	if strconv.FormatBool(*costTypes.IncludeCredit) != config.IncludeCredit {
		return fmt.Errorf("IncludeCredit not set properly '%v' should be '%v'", *costTypes.IncludeCredit, config.IncludeCredit)
	}

	if strconv.FormatBool(*costTypes.IncludeOtherSubscription) != config.IncludeOtherSubscription {
		return fmt.Errorf("IncludeOtherSubscription not set properly '%v' should be '%v'", *costTypes.IncludeOtherSubscription, config.IncludeOtherSubscription)
	}

	if strconv.FormatBool(*costTypes.IncludeRecurring) != config.IncludeRecurring {
		return fmt.Errorf("IncludeRecurring not set properly  '%v' should be '%v'", *costTypes.IncludeRecurring, config.IncludeRecurring)
	}

	if strconv.FormatBool(*costTypes.IncludeRefund) != config.IncludeRefund {
		return fmt.Errorf("IncludeRefund not set properly '%v' should be '%v'", *costTypes.IncludeRefund, config.IncludeRefund)
	}

	if strconv.FormatBool(*costTypes.IncludeSubscription) != config.IncludeSubscription {
		return fmt.Errorf("IncludeSubscription not set properly '%v' should be '%v'", *costTypes.IncludeSubscription, config.IncludeSubscription)
	}

	if strconv.FormatBool(*costTypes.IncludeSupport) != config.IncludeSupport {
		return fmt.Errorf("IncludeSupport not set properly '%v' should be '%v'", *costTypes.IncludeSupport, config.IncludeSupport)
	}

	if strconv.FormatBool(*costTypes.IncludeTax) != config.IncludeTax {
		return fmt.Errorf("IncludeTax not set properly '%v' should be '%v'", *costTypes.IncludeTax, config.IncludeTax)
	}

	if strconv.FormatBool(*costTypes.IncludeUpfront) != config.IncludeUpfront {
		return fmt.Errorf("IncludeUpfront not set properly '%v' should be '%v'", *costTypes.IncludeUpfront, config.IncludeUpfront)
	}

	if strconv.FormatBool(*costTypes.UseBlended) != config.UseBlended {
		return fmt.Errorf("UseBlended not set properly '%v' should be '%v'", *costTypes.UseBlended, config.UseBlended)
	}

	return nil
}

func testCheckBudgetDestroy(budgetName string, provider *schema.Provider) error {
	if budgetExists(budgetName, provider.Meta()) {
		return fmt.Errorf("Budget '%s' was not deleted properly", budgetName)
	}

	return nil
}

type budgetTestConfig struct {
	BudgetName               string
	BudgetType               string
	LimitAmount              string
	LimitUnit                string
	IncludeCredit            string
	IncludeOtherSubscription string
	IncludeRecurring         string
	IncludeRefund            string
	IncludeSubscription      string
	IncludeSupport           string
	IncludeTax               string
	IncludeUpfront           string
	UseBlended               string
	TimeUnit                 string
	TimePeriodStart          string
	TimePeriodEnd            string
	FilterKey                string
	FilterValue              string
}

func testBudgetHCLPrefixUseDefaults(budgetConfig budgetTestConfig) string {
	t := template.Must(template.New("t1").
		Parse(`
resource "aws_budget" "foo" {
	name_prefix = "{{.BudgetName}}"
	budget_type = "{{.BudgetType}}"
 	limit_amount = "{{.LimitAmount}}"
 	limit_unit = "{{.LimitUnit}}"
	time_period_start = "{{.TimePeriodStart}}" 
 	time_unit = "{{.TimeUnit}}"
	cost_filters {
		{{.FilterKey}} = "{{.FilterValue}}"
	}
}
`))
	var doc bytes.Buffer
	t.Execute(&doc, budgetConfig)
	return doc.String()
}

func testBudgetHCLPrefix(budgetConfig budgetTestConfig) string {
	t := template.Must(template.New("t1").
		Parse(`
resource "aws_budget" "foo" {
	name_prefix = "{{.BudgetName}}"
	budget_type = "{{.BudgetType}}"
 	limit_amount = "{{.LimitAmount}}"
 	limit_unit = "{{.LimitUnit}}"
	include_tax = "{{.IncludeTax}}"
	include_subscription = "{{.IncludeSubscription}}"
	use_blended = "{{.UseBlended}}"
	time_period_start = "{{.TimePeriodStart}}" 
	time_period_end = "{{.TimePeriodEnd}}"
 	time_unit = "{{.TimeUnit}}"
	cost_filters {
		{{.FilterKey}} = "{{.FilterValue}}"
	}
}
`))
	var doc bytes.Buffer
	t.Execute(&doc, budgetConfig)
	return doc.String()
}

func testBudgetHCLBasicUseDefaults(budgetConfig budgetTestConfig) string {
	t := template.Must(template.New("t1").
		Parse(`
resource "aws_budget" "foo" {
	name = "{{.BudgetName}}"
	budget_type = "{{.BudgetType}}"
 	limit_amount = "{{.LimitAmount}}"
 	limit_unit = "{{.LimitUnit}}"
	time_period_start = "{{.TimePeriodStart}}" 
 	time_unit = "{{.TimeUnit}}"
	cost_filters {
		{{.FilterKey}} = "{{.FilterValue}}"
	}
}
`))
	var doc bytes.Buffer
	t.Execute(&doc, budgetConfig)
	return doc.String()
}

func testBudgetHCLBasic(budgetConfig budgetTestConfig) string {
	t := template.Must(template.New("t1").
		Parse(`
resource "aws_budget" "foo" {
	name = "{{.BudgetName}}"
	budget_type = "{{.BudgetType}}"
 	limit_amount = "{{.LimitAmount}}"
 	limit_unit = "{{.LimitUnit}}"
	include_tax = "{{.IncludeTax}}"
	include_subscription = "{{.IncludeSubscription}}"
	use_blended = "{{.UseBlended}}"
	time_period_start = "{{.TimePeriodStart}}" 
	time_period_end = "{{.TimePeriodEnd}}"
 	time_unit = "{{.TimeUnit}}"
	cost_filters {
		{{.FilterKey}} = "{{.FilterValue}}"
	}
}
`))
	var doc bytes.Buffer
	t.Execute(&doc, budgetConfig)
	return doc.String()
}
