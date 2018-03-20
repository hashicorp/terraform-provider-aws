package aws

import (
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
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
		"cost_types": {
			Type:     schema.TypeList,
			Optional: true,
			MaxItems: 1,
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
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
				},
			},
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
	budget, err := expandBudgetsBudget(d)
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
	client := meta.(*AWSClient).budgetconn
	accountID := meta.(*AWSClient).accountid
	describeBudgetOutput, err := client.DescribeBudget(&budgets.DescribeBudgetInput{
		BudgetName: &budgetName,
		AccountId:  &accountID,
	})
	if isAWSErr(err, budgets.ErrCodeNotFoundException, "") {
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
		"budget_type":       flattenedBudget.budgetType,
		"time_unit":         flattenedBudget.timeUnit,
		"cost_filters":      convertCostFiltersToStringMap(flattenedBudget.costFilters),
		"limit_amount":      flattenedBudget.limitAmount,
		"limit_unit":        flattenedBudget.limitUnit,
		"cost_types":        []interface{}{flattenedBudget.costTypes},
		"time_period_start": flattenedBudget.timePeriodStart.Format("2006-01-02_15:04"),
		"time_period_end":   flattenedBudget.timePeriodEnd.Format("2006-01-02_15:04"),
	} {
		if _, ok := d.GetOk(k); ok {
			if err := d.Set(k, v); err != nil {
				return err
			}
		}
	}
	return nil
}

func resourceAwsBudgetsBudgetUpdate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*AWSClient).budgetconn
	accountID := meta.(*AWSClient).accountid
	budget, err := expandBudgetsBudget(d)
	if err != nil {
		return fmt.Errorf("could not create budget: %v", err)
	}

	_, err = client.UpdateBudget(&budgets.UpdateBudgetInput{
		AccountId: &accountID,
		NewBudget: budget,
	})
	if err != nil {
		return fmt.Errorf("update budget failed: %v", err)
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
	_, err := client.DeleteBudget(&budgets.DeleteBudgetInput{
		BudgetName: &budgetName,
		AccountId:  &accountID,
	})
	if err != nil {
		return fmt.Errorf("delete budget failed: %v", err)
	}

	return nil
}

type flattenedBudgetsBudget struct {
	name            *string
	budgetType      *string
	timeUnit        *string
	costFilters     map[string][]*string
	limitAmount     *string
	limitUnit       *string
	costTypes       map[string]bool
	timePeriodStart *time.Time
	timePeriodEnd   *time.Time
}

func expandBudgetsBudgetFlatten(budget *budgets.Budget) (*flattenedBudgetsBudget, error) {
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
	costTypesMap := map[string]bool{
		"include_credit":             *budgetCostTypes.IncludeCredit,
		"include_other_subscription": *budgetCostTypes.IncludeOtherSubscription,
		"include_recurring":          *budgetCostTypes.IncludeRecurring,
		"include_refund":             *budgetCostTypes.IncludeRefund,
		"include_subscription":       *budgetCostTypes.IncludeSubscription,
		"include_support":            *budgetCostTypes.IncludeSupport,
		"include_tax":                *budgetCostTypes.IncludeTax,
		"include_upfront":            *budgetCostTypes.IncludeUpfront,
		"use_blended":                *budgetCostTypes.UseBlended,
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

	return &flattenedBudgetsBudget{
		name:            budget.BudgetName,
		budgetType:      budget.BudgetType,
		timeUnit:        budget.TimeUnit,
		costFilters:     budget.CostFilters,
		limitAmount:     budgetLimit.Amount,
		limitUnit:       budgetLimit.Unit,
		costTypes:       costTypesMap,
		timePeriodStart: budgetTimePeriodStart,
		timePeriodEnd:   budgetTimePeriodEnd,
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

func expandBudgetsBudget(d *schema.ResourceData) (*budgets.Budget, error) {
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
	budgetCostTypes := d.Get("cost_types").([]interface{})
	costTypes := &budgets.CostTypes{
		IncludeCredit:            aws.Bool(true),
		IncludeOtherSubscription: aws.Bool(true),
		IncludeRecurring:         aws.Bool(true),
		IncludeRefund:            aws.Bool(true),
		IncludeSubscription:      aws.Bool(true),
		IncludeSupport:           aws.Bool(true),
		IncludeTax:               aws.Bool(true),
		IncludeUpfront:           aws.Bool(true),
		UseBlended:               aws.Bool(false),
	}
	if len(budgetCostTypes) == 1 {
		costTypesMap := budgetCostTypes[0].(map[string]interface{})
		for k, v := range map[string]*bool{
			"include_credit":             costTypes.IncludeCredit,
			"include_other_subscription": costTypes.IncludeOtherSubscription,
			"include_recurring":          costTypes.IncludeRecurring,
			"include_refund":             costTypes.IncludeRefund,
			"include_subscription":       costTypes.IncludeSubscription,
			"include_support":            costTypes.IncludeSupport,
			"include_tax":                costTypes.IncludeTax,
			"include_upfront":            costTypes.IncludeUpfront,
			"use_blended":                costTypes.UseBlended,
		} {
			if val, ok := costTypesMap[k]; ok {
				*v = val.(bool)
			}
		}
	}

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

	budget := &budgets.Budget{
		BudgetName: &budgetName,
		BudgetType: &budgetType,
		BudgetLimit: &budgets.Spend{
			Amount: &budgetLimitAmount,
			Unit:   &budgetLimitUnit,
		},
		CostTypes: costTypes,
		TimePeriod: &budgets.TimePeriod{
			End:   &budgetTimePeriodEnd,
			Start: &budgetTimePeriodStart,
		},
		TimeUnit:    &budgetTimeUnit,
		CostFilters: budgetCostFilters,
	}
	return budget, nil
}

func budgetExists(budgetName string, meta interface{}) bool {
	client := meta.(*AWSClient).budgetconn
	accountID := meta.(*AWSClient).accountid
	_, err := client.DescribeBudget(&budgets.DescribeBudgetInput{
		BudgetName: &budgetName,
		AccountId:  &accountID,
	})
	if isAWSErr(err, budgets.ErrCodeNotFoundException, "") {
		return false
	}

	return true
}
