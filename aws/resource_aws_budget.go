package aws

import (
	"fmt"
	"log"
	"time"

	"github.com/aws/aws-sdk-go/aws/awserr"
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
		"budget_name": {
			Type:     schema.TypeString,
			Required: true,
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
			Type:     schema.TypeString,
			Required: true,
		},
		"time_period_end": {
			Type:     schema.TypeString,
			Required: true,
		},
		"time_unit": {
			Type:     schema.TypeString,
			Required: true,
		},
		"cost_filters": &schema.Schema{
			Type:     schema.TypeMap,
			Optional: true,
			Computed: true,
		},
	}
}

func resourceAwsBudgetCreate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*AWSClient).budgetconn
	accountID := meta.(*AWSClient).accountid
	budget, err := newBudget(d)
	if err != nil {
		return fmt.Errorf("failed creating budget: %v", err)
	}

	createBudgetInput := new(budgets.CreateBudgetInput)
	createBudgetInput.SetAccountId(accountID)
	createBudgetInput.SetBudget(budget)
	_, err = client.CreateBudget(createBudgetInput)
	if err != nil {
		return fmt.Errorf("create budget failed: %v", err)
	}

	d.SetId(*budget.BudgetName)
	return resourceAwsBudgetUpdate(d, meta)
}

func resourceAwsBudgetRead(d *schema.ResourceData, meta interface{}) error {
	budgetName := d.Get("budget_name").(string)
	describeBudgetOutput, err := describeBudget(budgetName, meta)
	if isBudgetNotFoundException(err) {
		d.SetId("")
		return nil
	}

	if err != nil {
		return fmt.Errorf("describe budget failed: %v", err)
	}

	d.Set("budget_name", describeBudgetOutput.Budget.BudgetName)
	d.Set("budget_type", describeBudgetOutput.Budget.BudgetType)
	d.Set("limit_amount", describeBudgetOutput.Budget.BudgetLimit.Amount)
	d.Set("limit_unit", describeBudgetOutput.Budget.BudgetLimit.Unit)
	d.Set("include_tax", describeBudgetOutput.Budget.CostTypes.IncludeTax)
	d.Set("include_subscriptions", describeBudgetOutput.Budget.CostTypes.IncludeSubscription)
	d.Set("include_blended", describeBudgetOutput.Budget.CostTypes.UseBlended)
	d.Set("time_period_start", describeBudgetOutput.Budget.TimePeriod.Start)
	d.Set("time_period_end", describeBudgetOutput.Budget.TimePeriod.End)
	d.Set("time_unit", describeBudgetOutput.Budget.TimeUnit)
	d.Set("cost_filters", describeBudgetOutput.Budget.CostFilters)
	return nil
}

func resourceAwsBudgetUpdate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*AWSClient).budgetconn
	accountID := meta.(*AWSClient).accountid
	budget, err := newBudget(d)
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

	return resourceAwsBudgetRead(d, meta)
}

func resourceAwsBudgetDelete(d *schema.ResourceData, meta interface{}) error {
	budgetName := d.Get("budget_name").(string)
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

func newBudget(d *schema.ResourceData) (*budgets.Budget, error) {
	budgetName := d.Get("budget_name").(string)
	budgetType := d.Get("budget_type").(string)
	budgetLimitAmount := d.Get("limit_amount").(string)
	budgetLimitUnit := d.Get("limit_unit").(string)
	budgetIncludeTax := d.Get("include_tax").(bool)
	budgetIncludeSubscriptions := d.Get("include_subscriptions").(bool)
	budgetIncludeBlended := d.Get("include_blended").(bool)
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
		IncludeSubscription: &budgetIncludeSubscriptions,
		IncludeTax:          &budgetIncludeTax,
		UseBlended:          &budgetIncludeBlended,
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
