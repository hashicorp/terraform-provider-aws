package aws

import (
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/budgets"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/helper/schema"
)

func resourceAwsBudgetsBudget() *schema.Resource {
	return &schema.Resource{
		Schema: resourceAwsBudgetsBudgetSchema(),
		Create: resourceAwsBudgetsBudgetCreate,
		Read:   resourceAwsBudgetsBudgetRead,
		Update: resourceAwsBudgetsBudgetUpdate,
		Delete: resourceAwsBudgetsBudgetDelete,
	}
}

func resourceAwsBudgetsBudgetSchema() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		"name": {
			Type:          schema.TypeString,
			Optional:      true,
			ConflictsWith: []string{"name_prefix"},
		},
		"name_prefix": {
			Type:     schema.TypeString,
			Optional: true,
			ForceNew: true,
		},
		"budget_type": {
			Type:     schema.TypeString,
			Required: true,
		},
		"limit_amount": {
			Type:     schema.TypeString,
			Required: true,
		},
		"limit_unit": {
			Type:     schema.TypeString,
			Required: true,
		},
		"include_credit": {
			Type:     schema.TypeBool,
			Optional: true,
			Default:  true,
		},
		"include_other_subscription": {
			Type:     schema.TypeBool,
			Optional: true,
			Default:  true,
		},
		"include_recurring": {
			Type:     schema.TypeBool,
			Optional: true,
			Default:  true,
		},
		"include_refund": {
			Type:     schema.TypeBool,
			Optional: true,
			Default:  true,
		},
		"include_subscription": {
			Type:     schema.TypeBool,
			Optional: true,
			Default:  true,
		},
		"include_support": {
			Type:     schema.TypeBool,
			Optional: true,
			Default:  true,
		},
		"include_tax": {
			Type:     schema.TypeBool,
			Optional: true,
			Default:  true,
		},
		"include_upfront": {
			Type:     schema.TypeBool,
			Optional: true,
			Default:  true,
		},
		"use_blended": {
			Type:     schema.TypeBool,
			Optional: true,
			Default:  false,
		},
		"time_period_start": {
			Type:     schema.TypeString,
			Required: true,
		},
		"time_period_end": {
			Type:     schema.TypeString,
			Optional: true,
			Default:  "2087-06-15_00:00",
		},
		"time_unit": {
			Type:     schema.TypeString,
			Required: true,
		},
		"cost_filters": {
			Type:     schema.TypeMap,
			Optional: true,
			Computed: true,
		},
	}
}

func resourceAwsBudgetsBudgetCreate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*AWSClient).budgetconn
	accountID := meta.(*AWSClient).accountid
	budget, err := newBudgetsBudget(d)
	if err != nil {
		return fmt.Errorf("failed creating budget: %v", err)
	}

	_, err = client.CreateBudget(&budgets.CreateBudgetInput{
		AccountId: &accountID,
		Budget:    budget,
	})
	if err != nil {
		return fmt.Errorf("create budget failed: %v", err)
	}

	d.SetId(*budget.BudgetName)
	return resourceAwsBudgetsBudgetUpdate(d, meta)
}

func resourceAwsBudgetsBudgetRead(d *schema.ResourceData, meta interface{}) error {
	budgetName := d.Id()
	describeBudgetOutput, err := describeBudget(budgetName, meta)
	if isBudgetNotFoundException(err) {
		d.SetId("")
		return nil
	}

	if err != nil {
		return fmt.Errorf("describe budget failed: %v", err)
	}

	budget := describeBudgetOutput.Budget
	if budget == nil {
		return fmt.Errorf("empty budget returned from budget output: %v", describeBudgetOutput)
	}

	budgetLimit := budget.BudgetLimit
	if budgetLimit == nil {
		return fmt.Errorf("empty limit in budget: %v", budget)
	}

	budgetCostTypes := budget.CostTypes
	if budgetCostTypes == nil {
		return fmt.Errorf("empty CostTypes in budget: %v", budget)
	}

	budgetTimePeriod := budget.TimePeriod
	if budgetTimePeriod == nil {
		return fmt.Errorf("empty TimePeriod in budget: %v", budget)
	}

	budgetTimePeriodStart := budgetTimePeriod.Start
	if budgetTimePeriodStart == nil {
		return fmt.Errorf("empty TimePeriodStart in budget: %v", budget)
	}

	budgetTimePeriodEnd := budgetTimePeriod.End
	if budgetTimePeriodEnd == nil {
		return fmt.Errorf("empty TimePeriodEnd in budget: %v", budget)
	}

	if _, ok := d.GetOk("name"); ok {
		d.Set("name", budget.BudgetName)
	}

	for k, v := range map[string]interface{}{
		"budget_type":                budget.BudgetType,
		"time_unit":                  budget.TimeUnit,
		"cost_filters":               convertCostFiltersToStringMap(budget.CostFilters),
		"limit_amount":               budgetLimit.Amount,
		"limit_unit":                 budgetLimit.Unit,
		"include_credit":             budgetCostTypes.IncludeCredit,
		"include_other_subscription": budgetCostTypes.IncludeOtherSubscription,
		"include_recurring":          budgetCostTypes.IncludeRecurring,
		"include_refund":             budgetCostTypes.IncludeRefund,
		"include_subscription":       budgetCostTypes.IncludeSubscription,
		"include_support":            budgetCostTypes.IncludeSupport,
		"include_tax":                budgetCostTypes.IncludeTax,
		"include_upfront":            budgetCostTypes.IncludeUpfront,
		"use_blended":                budgetCostTypes.UseBlended,
		"time_period_start":          budgetTimePeriodStart.Format("2006-01-02_15:04"),
		"time_period_end":            budgetTimePeriodEnd.Format("2006-01-02_15:04"),
	} {
		if err := d.Set(k, v); err != nil {
			return err
		}
	}
	return nil
}

func resourceAwsBudgetsBudgetUpdate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*AWSClient).budgetconn
	accountID := meta.(*AWSClient).accountid
	budget, err := newBudgetsBudget(d)
	if err != nil {
		return fmt.Errorf("could not create budget: %v", err)
	}

	updateBudgetInput := new(budgets.UpdateBudgetInput)
	updateBudgetInput.SetAccountId(accountID)
	updateBudgetInput.SetNewBudget(budget)
	_, err = client.UpdateBudget(updateBudgetInput)
	if err != nil {
		return fmt.Errorf("updaate budget failed: %v", err)
	}

	return resourceAwsBudgetsBudgetRead(d, meta)
}

func resourceAwsBudgetsBudgetDelete(d *schema.ResourceData, meta interface{}) error {
	budgetName := d.Id()
	if !budgetExists(budgetName, meta) {
		log.Printf("[INFO] budget %s could not be found. skipping delete.", d.Id())
		return nil
	}

	client := meta.(*AWSClient).budgetconn
	accountID := meta.(*AWSClient).accountid
	deleteBudgetInput := new(budgets.DeleteBudgetInput)
	deleteBudgetInput.SetBudgetName(budgetName)
	deleteBudgetInput.SetAccountId(accountID)
	_, err := client.DeleteBudget(deleteBudgetInput)
	if err != nil {
		return fmt.Errorf("delete budget failed: %v", err)
	}

	return nil
}

func convertCostFiltersToStringMap(costFilters map[string][]*string) map[string]string {
	convertedCostFilters := make(map[string]string)
	for k, v := range costFilters {
		filterValues := make([]string, 0)
		for _, singleFilterValue := range v {
			filterValues = append(filterValues, *singleFilterValue)
		}

		convertedCostFilters[k] = strings.Join(filterValues, ",")
	}

	return convertedCostFilters
}

func newBudgetsBudget(d *schema.ResourceData) (*budgets.Budget, error) {
	var budgetName string
	if id := d.Id(); id != "" {
		budgetName = id

	} else if v, ok := d.GetOk("name"); ok {
		budgetName = v.(string)

	} else if v, ok := d.GetOk("name_prefix"); ok {
		budgetName = resource.PrefixedUniqueId(v.(string))

	} else {
		budgetName = resource.UniqueId()
	}

	budgetType := d.Get("budget_type").(string)
	budgetLimitAmount := d.Get("limit_amount").(string)
	budgetLimitUnit := d.Get("limit_unit").(string)
	budgetIncludeCredit := d.Get("include_credit").(bool)
	budgetIncludeOtherSubscription := d.Get("include_other_subscription").(bool)
	budgetIncludeRecurring := d.Get("include_recurring").(bool)
	budgetIncludeRefund := d.Get("include_refund").(bool)
	budgetIncludeSubscription := d.Get("include_subscription").(bool)
	budgetIncludeSupport := d.Get("include_support").(bool)
	budgetIncludeTax := d.Get("include_tax").(bool)
	budgetIncludeUpfront := d.Get("include_upfront").(bool)
	budgetUseBlended := d.Get("use_blended").(bool)
	budgetTimeUnit := d.Get("time_unit").(string)
	budgetCostFilters := make(map[string][]*string)
	for k, v := range d.Get("cost_filters").(map[string]interface{}) {
		filterValue := v.(string)
		budgetCostFilters[k] = append(budgetCostFilters[k], &filterValue)
	}

	budgetTimePeriodStart, err := time.Parse("2006-01-02_15:04", d.Get("time_period_start").(string))
	if err != nil {
		return nil, fmt.Errorf("failure parsing time: %v", err)
	}

	budgetTimePeriodEnd, err := time.Parse("2006-01-02_15:04", d.Get("time_period_end").(string))
	if err != nil {
		return nil, fmt.Errorf("failure parsing time: %v", err)
	}

	budget := new(budgets.Budget)
	budget.SetBudgetName(budgetName)
	budget.SetBudgetType(budgetType)
	budget.SetBudgetLimit(&budgets.Spend{
		Amount: &budgetLimitAmount,
		Unit:   &budgetLimitUnit,
	})
	budget.SetCostTypes(&budgets.CostTypes{
		IncludeCredit:            &budgetIncludeCredit,
		IncludeOtherSubscription: &budgetIncludeOtherSubscription,
		IncludeRecurring:         &budgetIncludeRecurring,
		IncludeRefund:            &budgetIncludeRefund,
		IncludeSubscription:      &budgetIncludeSubscription,
		IncludeSupport:           &budgetIncludeSupport,
		IncludeTax:               &budgetIncludeTax,
		IncludeUpfront:           &budgetIncludeUpfront,
		UseBlended:               &budgetUseBlended,
	})
	budget.SetTimePeriod(&budgets.TimePeriod{
		End:   &budgetTimePeriodEnd,
		Start: &budgetTimePeriodStart,
	})
	budget.SetTimeUnit(budgetTimeUnit)
	budget.SetCostFilters(budgetCostFilters)
	return budget, nil
}

func budgetExists(budgetName string, meta interface{}) bool {
	_, err := describeBudget(budgetName, meta)
	if isBudgetNotFoundException(err) {
		return false
	}

	return true
}

func isBudgetNotFoundException(err error) bool {
	if awsErr, ok := err.(awserr.Error); ok && awsErr.Code() == budgets.ErrCodeNotFoundException {
		return true
	}

	return false
}

func describeBudget(budgetName string, meta interface{}) (*budgets.DescribeBudgetOutput, error) {
	client := meta.(*AWSClient).budgetconn
	accountID := meta.(*AWSClient).accountid
	describeBudgetInput := new(budgets.DescribeBudgetInput)
	describeBudgetInput.SetBudgetName(budgetName)
	describeBudgetInput.SetAccountId(accountID)
	return client.DescribeBudget(describeBudgetInput)
}
