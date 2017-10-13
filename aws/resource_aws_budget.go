package aws

import (
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go/service/budgets"
	"github.com/hashicorp/terraform/helper/schema"
)

func resourceAwsBudget() *schema.Resource {
	return &schema.Resource{
		Schema: resourceAwsBudgetSchema(),
		Create: resourceAwsBudgetCreate,
		Read:   resourceAwsBudgetRead,
		Update: resourceAwsBudgetUpdate,
		Delete: resourceAwsBudgetDelete,
	}
}

func resourceAwsBudgetSchema() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		"name": {
			Type:     schema.TypeString,
			Required: true,
		},
		"type": {
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
		"include_tax": {
			Type:     schema.TypeBool,
			Required: true,
		},
		"include_subscriptions": {
			Type:     schema.TypeBool,
			Required: true,
		},
		"include_blended": {
			Type:     schema.TypeBool,
			Required: true,
		},
		"time_period_start": {
			Type:     schema.TypeInt,
			Required: true,
		},
		"time_period_end": {
			Type:     schema.TypeInt,
			Required: true,
		},
		"time_unit": {
			Type:     schema.TypeString,
			Required: true,
		},
	}
}

func resourceAwsBudgetCreate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*AWSClient).budgetconn
	accountID := meta.(*AWSClient).accountid
	budgetName := d.Get("name").(string)
	budgetType := d.Get("type").(string)
	budgetLimitAmount := d.Get("limit_amount").(string)
	budgetLimitUnit := d.Get("limit_unit").(string)
	budgetIncludeTax := d.Get("include_tax").(bool)
	budgetIncludeSubscriptions := d.Get("include_subscriptions").(bool)
	budgetIncludeBlended := d.Get("include_blended").(bool)
	budgetTimePeriodStart := time.Unix(int64(d.Get("time_period_start").(int)), 0)
	budgetTimePeriodEnd := time.Unix(int64(d.Get("time_period_end").(int)), 0)
	budgetTimeUnit := d.Get("time_unit").(string)

	budget := new(budgets.Budget)
	budget.SetBudgetName(budgetName)
	budget.SetBudgetType(budgetType)
	budget.SetBudgetLimit(&budgets.Spend{
		Amount: &budgetLimitAmount,
		Unit:   &budgetLimitUnit,
	})
	budget.SetCostTypes(&budgets.CostTypes{
		IncludeSubscription: &budgetIncludeSubscriptions,
		IncludeTax:          &budgetIncludeTax,
		UseBlended:          &budgetIncludeBlended,
	})
	budget.SetTimePeriod(&budgets.TimePeriod{
		End:   &budgetTimePeriodEnd,
		Start: &budgetTimePeriodStart,
	})
	budget.SetTimeUnit(budgetTimeUnit)
	createBudgetInput := new(budgets.CreateBudgetInput)
	createBudgetInput.SetAccountId(accountID)
	createBudgetInput.SetBudget(budget)
	_, err := client.CreateBudget(createBudgetInput)
	if err != nil {
		return fmt.Errorf("create budget failed: ", err)
	}

	return nil
}

func resourceAwsBudgetRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*AWSClient).budgetconn
	return fmt.Errorf("not yet implemented", client)
}

func resourceAwsBudgetUpdate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*AWSClient).budgetconn
	return fmt.Errorf("not yet implemented", client)
}

func resourceAwsBudgetDelete(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*AWSClient).budgetconn
	return fmt.Errorf("not yet implemented", client)
}
