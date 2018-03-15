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

	flattenedBudget, err := expandBudgetsBudgetFlatten(describeBudgetOutput.Budget)
	if err != nil {
		return fmt.Errorf("failed flattening budget output: %v", err)
	}

	if _, ok := d.GetOk("name"); ok {
		d.Set("name", flattenedBudget.name)
	}

	for k, v := range map[string]interface{}{
		"budget_type":                flattenedBudget.budgetType,
		"time_unit":                  flattenedBudget.timeUnit,
		"cost_filters":               convertCostFiltersToStringMap(flattenedBudget.costFilters),
		"limit_amount":               flattenedBudget.limitAmount,
		"limit_unit":                 flattenedBudget.limitUnit,
		"include_credit":             flattenedBudget.includeCredit,
		"include_other_subscription": flattenedBudget.includeOtherSubscription,
		"include_recurring":          flattenedBudget.includeRecurring,
		"include_refund":             flattenedBudget.includeRefund,
		"include_subscription":       flattenedBudget.includeSubscription,
		"include_support":            flattenedBudget.includeSupport,
		"include_tax":                flattenedBudget.includeTax,
		"include_upfront":            flattenedBudget.includeUpFront,
		"use_blended":                flattenedBudget.useBlended,
		"time_period_start":          flattenedBudget.timePeriodStart.Format("2006-01-02_15:04"),
		"time_period_end":            flattenedBudget.timePeriodEnd.Format("2006-01-02_15:04"),
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

type expandBudgetsBudgetFlattenedBudget struct {
	name                     *string
	budgetType               *string
	timeUnit                 *string
	costFilters              map[string][]*string
	limitAmount              *string
	limitUnit                *string
	includeCredit            *bool
	includeOtherSubscription *bool
	includeRecurring         *bool
	includeRefund            *bool
	includeSubscription      *bool
	includeSupport           *bool
	includeTax               *bool
	includeUpFront           *bool
	useBlended               *bool
	timePeriodStart          *time.Time
	timePeriodEnd            *time.Time
}

func expandBudgetsBudgetFlatten(budget *budgets.Budget) (*expandBudgetsBudgetFlattenedBudget, error) {
	if budget == nil {
		return nil, fmt.Errorf("empty budget returned from budget output: %v", budget)
	}

	budgetLimit := budget.BudgetLimit
	if budgetLimit == nil {
		return nil, fmt.Errorf("empty limit in budget: %v", budget)
	}

	budgetCostTypes := budget.CostTypes
	if budgetCostTypes == nil {
		return nil, fmt.Errorf("empty CostTypes in budget: %v", budget)
	}

	budgetTimePeriod := budget.TimePeriod
	if budgetTimePeriod == nil {
		return nil, fmt.Errorf("empty TimePeriod in budget: %v", budget)
	}

	budgetTimePeriodStart := budgetTimePeriod.Start
	if budgetTimePeriodStart == nil {
		return nil, fmt.Errorf("empty TimePeriodStart in budget: %v", budget)
	}

	budgetTimePeriodEnd := budgetTimePeriod.End
	if budgetTimePeriodEnd == nil {
		return nil, fmt.Errorf("empty TimePeriodEnd in budget: %v", budget)
	}

	return &expandBudgetsBudgetFlattenedBudget{
		name:                     budget.BudgetName,
		budgetType:               budget.BudgetType,
		timeUnit:                 budget.TimeUnit,
		costFilters:              budget.CostFilters,
		limitAmount:              budgetLimit.Amount,
		limitUnit:                budgetLimit.Unit,
		includeCredit:            budgetCostTypes.IncludeCredit,
		includeOtherSubscription: budgetCostTypes.IncludeOtherSubscription,
		includeRecurring:         budgetCostTypes.IncludeRecurring,
		includeRefund:            budgetCostTypes.IncludeRefund,
		includeSubscription:      budgetCostTypes.IncludeSubscription,
		includeSupport:           budgetCostTypes.IncludeSupport,
		includeTax:               budgetCostTypes.IncludeTax,
		includeUpFront:           budgetCostTypes.IncludeUpfront,
		useBlended:               budgetCostTypes.UseBlended,
		timePeriodStart:          budgetTimePeriodStart,
		timePeriodEnd:            budgetTimePeriodEnd,
	}, nil
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
