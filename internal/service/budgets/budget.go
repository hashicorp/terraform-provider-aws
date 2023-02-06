package budgets

import (
	"context"
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/arn"
	"github.com/aws/aws-sdk-go/service/budgets"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/shopspring/decimal"
)

func ResourceBudget() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceBudgetCreate,
		ReadWithoutTimeout:   resourceBudgetRead,
		UpdateWithoutTimeout: resourceBudgetUpdate,
		DeleteWithoutTimeout: resourceBudgetDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			"account_id": {
				Type:         schema.TypeString,
				Computed:     true,
				Optional:     true,
				ForceNew:     true,
				ValidateFunc: verify.ValidAccountID,
			},
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"auto_adjust_data": {
				Type:     schema.TypeList,
				Optional: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"auto_adjust_type": {
							Type:         schema.TypeString,
							Required:     true,
							ValidateFunc: validation.StringInSlice(budgets.AutoAdjustType_Values(), false),
						},
						"historical_options": {
							Type:     schema.TypeList,
							Optional: true,
							MaxItems: 1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"budget_adjustment_period": {
										Type:         schema.TypeInt,
										Required:     true,
										ValidateFunc: validation.IntBetween(1, 60),
									},
									"lookback_available_periods": {
										Type:     schema.TypeInt,
										Computed: true,
									},
								},
							},
						},
						"last_auto_adjust_time": {
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
			},
			"budget_type": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringInSlice(budgets.BudgetType_Values(), false),
			},
			"cost_filters": {
				Type:          schema.TypeMap,
				Optional:      true,
				Computed:      true,
				Elem:          &schema.Schema{Type: schema.TypeString},
				ConflictsWith: []string{"cost_filter"},
				Deprecated:    "Use the attribute \"cost_filter\" instead.",
			},
			"cost_filter": {
				Type:     schema.TypeSet,
				Optional: true,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"name": {
							Type:     schema.TypeString,
							Required: true,
						},
						"values": {
							Type:     schema.TypeList,
							Required: true,
							Elem: &schema.Schema{
								Type: schema.TypeString,
							},
						},
					},
				},
				ConflictsWith: []string{"cost_filters"},
			},
			"cost_types": {
				Type:     schema.TypeList,
				Computed: true,
				Optional: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"include_credit": {
							Type:     schema.TypeBool,
							Optional: true,
							Default:  true,
						},
						"include_discount": {
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
						"use_amortized": {
							Type:     schema.TypeBool,
							Optional: true,
							Default:  false,
						},
						"use_blended": {
							Type:     schema.TypeBool,
							Optional: true,
							Default:  false,
						},
					},
				},
			},
			"limit_amount": {
				Type:             schema.TypeString,
				Optional:         true,
				Computed:         true,
				DiffSuppressFunc: suppressEquivalentBudgetLimitAmount,
				ConflictsWith:    []string{"planned_limit"},
			},
			"limit_unit": {
				Type:          schema.TypeString,
				Optional:      true,
				Computed:      true,
				ConflictsWith: []string{"planned_limit"},
			},
			"name": {
				Type:          schema.TypeString,
				Optional:      true,
				Computed:      true,
				ForceNew:      true,
				ConflictsWith: []string{"name_prefix"},
			},
			"name_prefix": {
				Type:          schema.TypeString,
				Optional:      true,
				Computed:      true,
				ForceNew:      true,
				ConflictsWith: []string{"name"},
			},
			"notification": {
				Type:     schema.TypeSet,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"comparison_operator": {
							Type:         schema.TypeString,
							Required:     true,
							ValidateFunc: validation.StringInSlice(budgets.ComparisonOperator_Values(), false),
						},
						"notification_type": {
							Type:         schema.TypeString,
							Required:     true,
							ValidateFunc: validation.StringInSlice(budgets.NotificationType_Values(), false),
						},
						"subscriber_email_addresses": {
							Type:     schema.TypeSet,
							Optional: true,
							Elem:     &schema.Schema{Type: schema.TypeString},
						},
						"subscriber_sns_topic_arns": {
							Type:     schema.TypeSet,
							Optional: true,
							Elem: &schema.Schema{
								Type:         schema.TypeString,
								ValidateFunc: verify.ValidARN,
							},
						},
						"threshold": {
							Type:     schema.TypeFloat,
							Required: true,
						},
						"threshold_type": {
							Type:         schema.TypeString,
							Required:     true,
							ValidateFunc: validation.StringInSlice(budgets.ThresholdType_Values(), false),
						},
					},
				},
			},
			"planned_limit": {
				Type:     schema.TypeSet,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"amount": {
							Type:             schema.TypeString,
							Required:         true,
							DiffSuppressFunc: suppressEquivalentBudgetLimitAmount,
						},
						"start_time": {
							Type:         schema.TypeString,
							Required:     true,
							ValidateFunc: validTimePeriodTimestamp,
						},
						"unit": {
							Type:     schema.TypeString,
							Required: true,
						},
					},
				},
				ConflictsWith: []string{"limit_amount", "limit_unit"},
			},
			"time_period_end": {
				Type:         schema.TypeString,
				Optional:     true,
				Default:      "2087-06-15_00:00",
				ValidateFunc: validTimePeriodTimestamp,
			},
			"time_period_start": {
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				ValidateFunc: validTimePeriodTimestamp,
			},
			"time_unit": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringInSlice(budgets.TimeUnit_Values(), false),
			},
		},
	}
}

func resourceBudgetCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).BudgetsConn()

	budget, err := expandBudgetUnmarshal(d)

	if err != nil {
		return diag.Errorf("expandBudgetUnmarshal: %s", err)
	}

	name := create.Name(d.Get("name").(string), d.Get("name_prefix").(string))
	budget.BudgetName = aws.String(name)

	accountID := d.Get("account_id").(string)
	if accountID == "" {
		accountID = meta.(*conns.AWSClient).AccountID
	}

	_, err = conn.CreateBudgetWithContext(ctx, &budgets.CreateBudgetInput{
		AccountId: aws.String(accountID),
		Budget:    budget,
	})

	if err != nil {
		return diag.Errorf("creating Budget (%s): %s", name, err)
	}

	d.SetId(BudgetCreateResourceID(accountID, aws.StringValue(budget.BudgetName)))

	notificationsRaw := d.Get("notification").(*schema.Set).List()
	notifications, subscribers := expandBudgetNotificationsUnmarshal(notificationsRaw)

	err = createBudgetNotifications(ctx, conn, notifications, subscribers, *budget.BudgetName, accountID)

	if err != nil {
		return diag.Errorf("creating Budget (%s) notifications: %s", d.Id(), err)
	}

	return resourceBudgetRead(ctx, d, meta)
}

func resourceBudgetRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).BudgetsConn()

	accountID, budgetName, err := BudgetParseResourceID(d.Id())

	if err != nil {
		return diag.FromErr(err)
	}

	budget, err := FindBudgetByTwoPartKey(ctx, conn, accountID, budgetName)

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] Budget (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return diag.Errorf("reading Budget (%s): %s", d.Id(), err)
	}

	d.Set("account_id", accountID)
	arn := arn.ARN{
		Partition: meta.(*conns.AWSClient).Partition,
		Service:   "budgets",
		AccountID: accountID,
		Resource:  fmt.Sprintf("budget/%s", budgetName),
	}
	d.Set("arn", arn.String())
	d.Set("budget_type", budget.BudgetType)

	// `cost_filters` should be removed in future releases
	if err := d.Set("cost_filter", convertCostFiltersToMap(budget.CostFilters)); err != nil {
		return diag.Errorf("setting cost_filter: %s", err)
	}
	if err := d.Set("cost_filters", convertCostFiltersToStringMap(budget.CostFilters)); err != nil {
		return diag.Errorf("setting cost_filters: %s", err)
	}
	if err := d.Set("cost_types", flattenCostTypes(budget.CostTypes)); err != nil {
		return diag.Errorf("setting cost_types: %s", err)
	}
	if err := d.Set("auto_adjust_data", flattenAutoAdjustData(budget.AutoAdjustData)); err != nil {
		return diag.Errorf("setting auto_adjust_data: %s", err)
	}

	if budget.BudgetLimit != nil {
		d.Set("limit_amount", budget.BudgetLimit.Amount)
		d.Set("limit_unit", budget.BudgetLimit.Unit)
	}

	d.Set("name", budget.BudgetName)
	d.Set("name_prefix", create.NamePrefixFromName(aws.StringValue(budget.BudgetName)))

	if err := d.Set("planned_limit", convertPlannedBudgetLimitsToSet(budget.PlannedBudgetLimits)); err != nil {
		return diag.Errorf("setting planned_limit: %s", err)
	}

	if budget.TimePeriod != nil {
		d.Set("time_period_end", TimePeriodTimestampToString(budget.TimePeriod.End))
		d.Set("time_period_start", TimePeriodTimestampToString(budget.TimePeriod.Start))
	}

	d.Set("time_unit", budget.TimeUnit)

	notifications, err := findNotifications(ctx, conn, accountID, budgetName)

	if tfresource.NotFound(err) {
		return nil
	}

	if err != nil {
		return diag.Errorf("reading Budget (%s) notifications: %s", d.Id(), err)
	}

	var tfList []interface{}

	for _, notification := range notifications {
		tfMap := make(map[string]interface{})

		tfMap["comparison_operator"] = aws.StringValue(notification.ComparisonOperator)
		tfMap["threshold"] = aws.Float64Value(notification.Threshold)
		tfMap["notification_type"] = aws.StringValue(notification.NotificationType)

		if notification.ThresholdType == nil {
			// The AWS API doesn't seem to return a ThresholdType if it's set to PERCENTAGE
			// Set it manually to make behavior more predictable
			tfMap["threshold_type"] = budgets.ThresholdTypePercentage
		} else {
			tfMap["threshold_type"] = aws.StringValue(notification.ThresholdType)
		}

		subscribers, err := findSubscribers(ctx, conn, accountID, budgetName, notification)

		if tfresource.NotFound(err) {
			tfList = append(tfList, tfMap)
			continue
		}

		if err != nil {
			return diag.Errorf("reading Budget (%s) subscribers: %s", d.Id(), err)
		}

		var emailSubscribers []string
		var snsSubscribers []string

		for _, subscriber := range subscribers {
			if aws.StringValue(subscriber.SubscriptionType) == budgets.SubscriptionTypeSns {
				snsSubscribers = append(snsSubscribers, aws.StringValue(subscriber.Address))
			} else if aws.StringValue(subscriber.SubscriptionType) == budgets.SubscriptionTypeEmail {
				emailSubscribers = append(emailSubscribers, aws.StringValue(subscriber.Address))
			}
		}

		tfMap["subscriber_email_addresses"] = emailSubscribers
		tfMap["subscriber_sns_topic_arns"] = snsSubscribers

		tfList = append(tfList, tfMap)
	}

	if err := d.Set("notification", tfList); err != nil {
		return diag.Errorf("setting notification: %s", err)
	}

	return nil
}

func resourceBudgetUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).BudgetsConn()

	accountID, _, err := BudgetParseResourceID(d.Id())

	if err != nil {
		return diag.FromErr(err)
	}

	budget, err := expandBudgetUnmarshal(d)

	if err != nil {
		return diag.Errorf("expandBudgetUnmarshal: %s", err)
	}

	_, err = conn.UpdateBudgetWithContext(ctx, &budgets.UpdateBudgetInput{
		AccountId: aws.String(accountID),
		NewBudget: budget,
	})

	if err != nil {
		return diag.Errorf("updating Budget (%s): %s", d.Id(), err)
	}

	err = updateBudgetNotifications(ctx, conn, d)

	if err != nil {
		return diag.Errorf("updating Budget (%s) notifications: %s", d.Id(), err)
	}

	return resourceBudgetRead(ctx, d, meta)
}

func resourceBudgetDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).BudgetsConn()

	accountID, budgetName, err := BudgetParseResourceID(d.Id())

	if err != nil {
		return diag.FromErr(err)
	}

	log.Printf("[DEBUG] Deleting Budget: %s", d.Id())
	_, err = conn.DeleteBudgetWithContext(ctx, &budgets.DeleteBudgetInput{
		AccountId:  aws.String(accountID),
		BudgetName: aws.String(budgetName),
	})

	if tfawserr.ErrCodeEquals(err, budgets.ErrCodeNotFoundException) {
		return nil
	}

	if err != nil {
		return diag.Errorf("deleting Budget (%s): %s", d.Id(), err)
	}

	return nil
}

const budgetResourceIDSeparator = ":"

func BudgetCreateResourceID(accountID, budgetName string) string {
	parts := []string{accountID, budgetName}
	id := strings.Join(parts, budgetResourceIDSeparator)

	return id
}

func BudgetParseResourceID(id string) (string, string, error) {
	parts := strings.Split(id, budgetResourceIDSeparator)

	if len(parts) == 2 && parts[0] != "" && parts[1] != "" {
		return parts[0], parts[1], nil
	}

	return "", "", fmt.Errorf("unexpected format for ID (%[1]s), expected AccountID%[2]sBudgetName", id, budgetActionResourceIDSeparator)
}

func FindBudgetByTwoPartKey(ctx context.Context, conn *budgets.Budgets, accountID, budgetName string) (*budgets.Budget, error) {
	input := &budgets.DescribeBudgetInput{
		AccountId:  aws.String(accountID),
		BudgetName: aws.String(budgetName),
	}

	output, err := conn.DescribeBudgetWithContext(ctx, input)

	if tfawserr.ErrCodeEquals(err, budgets.ErrCodeNotFoundException) {
		return nil, &resource.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil || output.Budget == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	return output.Budget, nil
}

func findNotifications(ctx context.Context, conn *budgets.Budgets, accountID, budgetName string) ([]*budgets.Notification, error) {
	input := &budgets.DescribeNotificationsForBudgetInput{
		AccountId:  aws.String(accountID),
		BudgetName: aws.String(budgetName),
	}
	var output []*budgets.Notification

	err := conn.DescribeNotificationsForBudgetPagesWithContext(ctx, input, func(page *budgets.DescribeNotificationsForBudgetOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, notification := range page.Notifications {
			if notification == nil {
				continue
			}

			output = append(output, notification)
		}

		return !lastPage
	})

	if tfawserr.ErrCodeEquals(err, budgets.ErrCodeNotFoundException) {
		return nil, &resource.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if len(output) == 0 {
		return nil, &resource.NotFoundError{LastRequest: input}
	}

	return output, nil
}

func findSubscribers(ctx context.Context, conn *budgets.Budgets, accountID, budgetName string, notification *budgets.Notification) ([]*budgets.Subscriber, error) {
	input := &budgets.DescribeSubscribersForNotificationInput{
		AccountId:    aws.String(accountID),
		BudgetName:   aws.String(budgetName),
		Notification: notification,
	}
	var output []*budgets.Subscriber

	err := conn.DescribeSubscribersForNotificationPagesWithContext(ctx, input, func(page *budgets.DescribeSubscribersForNotificationOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, subscriber := range page.Subscribers {
			if subscriber == nil {
				continue
			}

			output = append(output, subscriber)
		}

		return !lastPage
	})

	if tfawserr.ErrCodeEquals(err, budgets.ErrCodeNotFoundException) {
		return nil, &resource.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if len(output) == 0 {
		return nil, &resource.NotFoundError{LastRequest: input}
	}

	return output, nil
}

func createBudgetNotifications(ctx context.Context, conn *budgets.Budgets, notifications []*budgets.Notification, subscribers [][]*budgets.Subscriber, budgetName string, accountID string) error {
	for i, notification := range notifications {
		subscribers := subscribers[i]

		if len(subscribers) == 0 {
			return fmt.Errorf("Budget notification must have at least one subscriber")
		}

		_, err := conn.CreateNotificationWithContext(ctx, &budgets.CreateNotificationInput{
			AccountId:    aws.String(accountID),
			BudgetName:   aws.String(budgetName),
			Notification: notification,
			Subscribers:  subscribers,
		})

		if err != nil {
			return err
		}
	}

	return nil
}

func updateBudgetNotifications(ctx context.Context, conn *budgets.Budgets, d *schema.ResourceData) error {
	accountID, budgetName, err := BudgetParseResourceID(d.Id())

	if err != nil {
		return err
	}

	if d.HasChange("notification") {
		o, n := d.GetChange("notification")
		os := o.(*schema.Set)
		ns := n.(*schema.Set)

		removeNotifications, _ := expandBudgetNotificationsUnmarshal(os.Difference(ns).List())
		addNotifications, addSubscribers := expandBudgetNotificationsUnmarshal(ns.Difference(os).List())

		for _, notification := range removeNotifications {
			input := &budgets.DeleteNotificationInput{
				AccountId:    aws.String(accountID),
				BudgetName:   aws.String(budgetName),
				Notification: notification,
			}

			_, err := conn.DeleteNotificationWithContext(ctx, input)

			if err != nil {
				return fmt.Errorf("deleting Budget (%s) notification: %s", d.Id(), err)
			}
		}

		err = createBudgetNotifications(ctx, conn, addNotifications, addSubscribers, budgetName, accountID)

		if err != nil {
			return fmt.Errorf("creating Budget (%s) notifications: %s", d.Id(), err)
		}
	}

	return nil
}

func flattenAutoAdjustData(autoAdjustData *budgets.AutoAdjustData) []map[string]interface{} {
	if autoAdjustData == nil {
		return []map[string]interface{}{}
	}

	attrs := map[string]interface{}{
		"auto_adjust_type":      aws.StringValue(autoAdjustData.AutoAdjustType),
		"last_auto_adjust_time": aws.TimeValue(autoAdjustData.LastAutoAdjustTime).Format(time.RFC3339),
	}

	if *autoAdjustData.HistoricalOptions != (budgets.HistoricalOptions{}) { // nosemgrep: ci.prefer-aws-go-sdk-pointer-conversion-conditional
		attrs["historical_options"] = flattenHistoricalOptions(autoAdjustData.HistoricalOptions)
	}

	return []map[string]interface{}{attrs}
}

func flattenHistoricalOptions(historicalOptions *budgets.HistoricalOptions) []map[string]interface{} {
	if historicalOptions == nil {
		return []map[string]interface{}{}
	}

	attrs := map[string]interface{}{
		"budget_adjustment_period":   aws.Int64Value(historicalOptions.BudgetAdjustmentPeriod),
		"lookback_available_periods": aws.Int64Value(historicalOptions.LookBackAvailablePeriods),
	}

	return []map[string]interface{}{attrs}
}

func flattenCostTypes(costTypes *budgets.CostTypes) []map[string]interface{} {
	if costTypes == nil {
		return []map[string]interface{}{}
	}

	m := map[string]interface{}{
		"include_credit":             aws.BoolValue(costTypes.IncludeCredit),
		"include_discount":           aws.BoolValue(costTypes.IncludeDiscount),
		"include_other_subscription": aws.BoolValue(costTypes.IncludeOtherSubscription),
		"include_recurring":          aws.BoolValue(costTypes.IncludeRecurring),
		"include_refund":             aws.BoolValue(costTypes.IncludeRefund),
		"include_subscription":       aws.BoolValue(costTypes.IncludeSubscription),
		"include_support":            aws.BoolValue(costTypes.IncludeSupport),
		"include_tax":                aws.BoolValue(costTypes.IncludeTax),
		"include_upfront":            aws.BoolValue(costTypes.IncludeUpfront),
		"use_amortized":              aws.BoolValue(costTypes.UseAmortized),
		"use_blended":                aws.BoolValue(costTypes.UseBlended),
	}
	return []map[string]interface{}{m}
}

func convertCostFiltersToMap(costFilters map[string][]*string) []map[string]interface{} {
	convertedCostFilters := make([]map[string]interface{}, 0)
	for k, v := range costFilters {
		convertedCostFilter := make(map[string]interface{})
		filterValues := make([]string, 0)
		for _, singleFilterValue := range v {
			filterValues = append(filterValues, *singleFilterValue)
		}
		convertedCostFilter["values"] = filterValues
		convertedCostFilter["name"] = k
		convertedCostFilters = append(convertedCostFilters, convertedCostFilter)
	}

	return convertedCostFilters
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

func convertPlannedBudgetLimitsToSet(plannedBudgetLimits map[string]*budgets.Spend) []interface{} {
	if plannedBudgetLimits == nil {
		return nil
	}

	convertedPlannedBudgetLimits := make([]interface{}, len(plannedBudgetLimits))
	i := 0

	for k, v := range plannedBudgetLimits {
		if v == nil {
			return nil
		}

		startTime, err := TimePeriodSecondsToString(k)
		if err != nil {
			return nil
		}

		convertedPlannedBudgetLimit := make(map[string]string)
		convertedPlannedBudgetLimit["amount"] = aws.StringValue(v.Amount)
		convertedPlannedBudgetLimit["start_time"] = startTime
		convertedPlannedBudgetLimit["unit"] = aws.StringValue(v.Unit)

		convertedPlannedBudgetLimits[i] = convertedPlannedBudgetLimit
		i++
	}

	return convertedPlannedBudgetLimits
}

func expandBudgetUnmarshal(d *schema.ResourceData) (*budgets.Budget, error) {
	budgetName := d.Get("name").(string)
	budgetType := d.Get("budget_type").(string)
	budgetTimeUnit := d.Get("time_unit").(string)
	budgetCostFilters := make(map[string][]*string)

	if costFilter, ok := d.GetOk("cost_filter"); ok {
		for _, v := range costFilter.(*schema.Set).List() {
			element := v.(map[string]interface{})
			key := element["name"].(string)
			for _, filterValue := range element["values"].([]interface{}) {
				budgetCostFilters[key] = append(budgetCostFilters[key], aws.String(filterValue.(string)))
			}
		}
	} else if costFilters, ok := d.GetOk("cost_filters"); ok {
		for k, v := range costFilters.(map[string]interface{}) {
			filterValue := v.(string)
			budgetCostFilters[k] = append(budgetCostFilters[k], aws.String(filterValue))
		}
	}

	budgetTimePeriodStart, err := timePeriodTimestampFromString(d.Get("time_period_start").(string))

	if err != nil {
		return nil, err
	}

	budgetTimePeriodEnd, err := timePeriodTimestampFromString(d.Get("time_period_end").(string))

	if err != nil {
		return nil, err
	}

	budget := &budgets.Budget{
		BudgetName: aws.String(budgetName),
		BudgetType: aws.String(budgetType),
		TimePeriod: &budgets.TimePeriod{
			End:   budgetTimePeriodEnd,
			Start: budgetTimePeriodStart,
		},
		TimeUnit:    aws.String(budgetTimeUnit),
		CostFilters: budgetCostFilters,
	}

	if v, ok := d.GetOk("auto_adjust_data"); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
		budget.AutoAdjustData = expandAutoAdjustData(v.([]interface{})[0].(map[string]interface{}))
	} else {
		if plannedBudgetLimitsRaw, ok := d.GetOk("planned_limit"); ok {
			plannedBudgetLimitsRaw := plannedBudgetLimitsRaw.(*schema.Set).List()

			plannedBudgetLimits, err := expandPlannedBudgetLimitsUnmarshal(plannedBudgetLimitsRaw)
			if err != nil {
				return nil, err
			}

			budget.PlannedBudgetLimits = plannedBudgetLimits
		} else {
			spendAmountValue, spendLimitOk := d.GetOk("limit_amount")
			spendUnitValue, spendUnitOk := d.GetOk("limit_unit")

			if spendUnitOk && spendLimitOk {
				budget.BudgetLimit = &budgets.Spend{
					Amount: aws.String(spendAmountValue.(string)),
					Unit:   aws.String(spendUnitValue.(string)),
				}
			}
		}
	}

	if v, ok := d.GetOk("cost_types"); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
		budget.CostTypes = expandCostTypes(v.([]interface{})[0].(map[string]interface{}))
	}

	return budget, nil
}

func expandAutoAdjustData(tfMap map[string]interface{}) *budgets.AutoAdjustData {
	if tfMap == nil {
		return nil
	}

	apiObject := &budgets.AutoAdjustData{}

	if v, ok := tfMap["auto_adjust_type"].(string); ok {
		apiObject.AutoAdjustType = aws.String(v)
	}

	if v, ok := tfMap["historical_options"].([]interface{}); ok && len(v) > 0 {
		apiObject.HistoricalOptions = expandHistoricalOptions(v)
	}

	return apiObject
}

func expandHistoricalOptions(l []interface{}) *budgets.HistoricalOptions {
	if len(l) == 0 || l[0] == nil {
		return nil
	}

	m := l[0].(map[string]interface{})

	apiObject := &budgets.HistoricalOptions{}

	if v, ok := m["budget_adjustment_period"].(int); ok && v != 0 {
		apiObject.BudgetAdjustmentPeriod = aws.Int64(int64(v))
	}

	return apiObject
}

func expandCostTypes(tfMap map[string]interface{}) *budgets.CostTypes {
	if tfMap == nil {
		return nil
	}

	apiObject := &budgets.CostTypes{}

	if v, ok := tfMap["include_credit"].(bool); ok {
		apiObject.IncludeCredit = aws.Bool(v)
	}
	if v, ok := tfMap["include_discount"].(bool); ok {
		apiObject.IncludeDiscount = aws.Bool(v)
	}
	if v, ok := tfMap["include_other_subscription"].(bool); ok {
		apiObject.IncludeOtherSubscription = aws.Bool(v)
	}
	if v, ok := tfMap["include_recurring"].(bool); ok {
		apiObject.IncludeRecurring = aws.Bool(v)
	}
	if v, ok := tfMap["include_refund"].(bool); ok {
		apiObject.IncludeRefund = aws.Bool(v)
	}
	if v, ok := tfMap["include_subscription"].(bool); ok {
		apiObject.IncludeSubscription = aws.Bool(v)
	}
	if v, ok := tfMap["include_support"].(bool); ok {
		apiObject.IncludeSupport = aws.Bool(v)
	}
	if v, ok := tfMap["include_tax"].(bool); ok {
		apiObject.IncludeTax = aws.Bool(v)
	}
	if v, ok := tfMap["include_upfront"].(bool); ok {
		apiObject.IncludeUpfront = aws.Bool(v)
	}
	if v, ok := tfMap["use_amortized"].(bool); ok {
		apiObject.UseAmortized = aws.Bool(v)
	}
	if v, ok := tfMap["use_blended"].(bool); ok {
		apiObject.UseBlended = aws.Bool(v)
	}

	return apiObject
}

func expandPlannedBudgetLimitsUnmarshal(plannedBudgetLimitsRaw []interface{}) (map[string]*budgets.Spend, error) {
	plannedBudgetLimits := make(map[string]*budgets.Spend, len(plannedBudgetLimitsRaw))

	for _, plannedBudgetLimit := range plannedBudgetLimitsRaw {
		plannedBudgetLimit := plannedBudgetLimit.(map[string]interface{})

		key, err := TimePeriodSecondsFromString(plannedBudgetLimit["start_time"].(string))
		if err != nil {
			return nil, err
		}

		amount := plannedBudgetLimit["amount"].(string)
		unit := plannedBudgetLimit["unit"].(string)

		plannedBudgetLimits[key] = &budgets.Spend{
			Amount: aws.String(amount),
			Unit:   aws.String(unit),
		}
	}

	return plannedBudgetLimits, nil
}

func expandBudgetNotificationsUnmarshal(notificationsRaw []interface{}) ([]*budgets.Notification, [][]*budgets.Subscriber) {
	notifications := make([]*budgets.Notification, len(notificationsRaw))
	subscribersForNotifications := make([][]*budgets.Subscriber, len(notificationsRaw))
	for i, notificationRaw := range notificationsRaw {
		notificationRaw := notificationRaw.(map[string]interface{})
		comparisonOperator := notificationRaw["comparison_operator"].(string)
		threshold := notificationRaw["threshold"].(float64)
		thresholdType := notificationRaw["threshold_type"].(string)
		notificationType := notificationRaw["notification_type"].(string)

		notifications[i] = &budgets.Notification{
			ComparisonOperator: aws.String(comparisonOperator),
			Threshold:          aws.Float64(threshold),
			ThresholdType:      aws.String(thresholdType),
			NotificationType:   aws.String(notificationType),
		}

		emailSubscribers := expandSubscribers(notificationRaw["subscriber_email_addresses"], budgets.SubscriptionTypeEmail)
		snsSubscribers := expandSubscribers(notificationRaw["subscriber_sns_topic_arns"], budgets.SubscriptionTypeSns)

		subscribersForNotifications[i] = append(emailSubscribers, snsSubscribers...)
	}
	return notifications, subscribersForNotifications
}

func expandSubscribers(rawList interface{}, subscriptionType string) []*budgets.Subscriber {
	result := make([]*budgets.Subscriber, 0)
	addrs := flex.ExpandStringSet(rawList.(*schema.Set))
	for _, addr := range addrs {
		result = append(result, &budgets.Subscriber{
			SubscriptionType: aws.String(subscriptionType),
			Address:          addr,
		})
	}
	return result
}

func suppressEquivalentBudgetLimitAmount(k, old, new string, d *schema.ResourceData) bool {
	d1, err := decimal.NewFromString(old)

	if err != nil {
		return false
	}

	d2, err := decimal.NewFromString(new)

	if err != nil {
		return false
	}

	return d1.Equal(d2)
}

func TimePeriodSecondsFromString(s string) (string, error) {
	if s == "" {
		return "", nil
	}

	ts, err := time.Parse(timePeriodLayout, s)

	if err != nil {
		return "", err
	}

	return strconv.FormatInt(aws.Time(ts).Unix(), 10), nil
}

func TimePeriodSecondsToString(s string) (string, error) {
	startTime, err := strconv.ParseInt(s, 10, 64)
	if err != nil {
		return "", err
	}

	startTime = startTime * 1000

	return aws.SecondsTimeValue(&startTime).UTC().Format(timePeriodLayout), nil
}

const (
	timePeriodLayout = "2006-01-02_15:04"
)

func timePeriodTimestampFromString(s string) (*time.Time, error) {
	if s == "" {
		return nil, nil
	}

	ts, err := time.Parse(timePeriodLayout, s)

	if err != nil {
		return nil, err
	}

	return aws.Time(ts), nil
}

func TimePeriodTimestampToString(ts *time.Time) string {
	if ts == nil {
		return ""
	}

	return aws.TimeValue(ts).Format(timePeriodLayout)
}

func validTimePeriodTimestamp(v interface{}, k string) (ws []string, errors []error) {
	_, err := time.Parse(timePeriodLayout, v.(string))

	if err != nil {
		errors = append(errors, fmt.Errorf("%q cannot be parsed as %q: %w", k, timePeriodLayout, err))
	}

	return
}
