package budgets

import (
	"fmt"
	"log"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/arn"
	"github.com/aws/aws-sdk-go/service/budgets"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
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
		Create: resourceBudgetCreate,
		Read:   resourceBudgetRead,
		Update: resourceBudgetUpdate,
		Delete: resourceBudgetDelete,

		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
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
				Required:         true,
				DiffSuppressFunc: suppressEquivalentBudgetLimitAmount,
			},
			"limit_unit": {
				Type:     schema.TypeString,
				Required: true,
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
			"time_period_end": {
				Type:         schema.TypeString,
				Optional:     true,
				Default:      "2087-06-15_00:00",
				ValidateFunc: ValidTimePeriodTimestamp,
			},
			"time_period_start": {
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				ValidateFunc: ValidTimePeriodTimestamp,
			},
			"time_unit": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringInSlice(budgets.TimeUnit_Values(), false),
			},
		},
	}
}

func resourceBudgetCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).BudgetsConn

	budget, err := expandBudgetUnmarshal(d)
	if err != nil {
		return fmt.Errorf("failed unmarshalling budget: %v", err)
	}

	name := create.Name(d.Get("name").(string), d.Get("name_prefix").(string))
	budget.BudgetName = aws.String(name)

	accountID := d.Get("account_id").(string)
	if accountID == "" {
		accountID = meta.(*conns.AWSClient).AccountID
	}

	_, err = conn.CreateBudget(&budgets.CreateBudgetInput{
		AccountId: aws.String(accountID),
		Budget:    budget,
	})

	if err != nil {
		return fmt.Errorf("error creating Budget (%s): %w", name, err)
	}

	d.SetId(fmt.Sprintf("%s:%s", accountID, aws.StringValue(budget.BudgetName)))

	notificationsRaw := d.Get("notification").(*schema.Set).List()
	notifications, subscribers := expandBudgetNotificationsUnmarshal(notificationsRaw)

	err = resourceBudgetNotificationsCreate(notifications, subscribers, *budget.BudgetName, accountID, meta)

	if err != nil {
		return fmt.Errorf("error creating Budget (%s) Notifications: %s", d.Id(), err)
	}

	return resourceBudgetRead(d, meta)
}

func resourceBudgetRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).BudgetsConn

	accountID, budgetName, err := BudgetParseResourceID(d.Id())

	if err != nil {
		return err
	}

	budget, err := FindBudgetByAccountIDAndBudgetName(conn, accountID, budgetName)

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] Budget (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return fmt.Errorf("error reading Budget (%s): %w", d.Id(), err)
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
		return fmt.Errorf("error setting cost_filter: %w", err)
	}
	if err := d.Set("cost_filters", convertCostFiltersToStringMap(budget.CostFilters)); err != nil {
		return fmt.Errorf("error setting cost_filters: %w", err)
	}

	if err := d.Set("cost_types", flattenCostTypes(budget.CostTypes)); err != nil {
		return fmt.Errorf("error setting cost_types: %w", err)
	}

	if budget.BudgetLimit != nil {
		d.Set("limit_amount", budget.BudgetLimit.Amount)
		d.Set("limit_unit", budget.BudgetLimit.Unit)
	}

	d.Set("name", budget.BudgetName)
	d.Set("name_prefix", create.NamePrefixFromName(aws.StringValue(budget.BudgetName)))

	if budget.TimePeriod != nil {
		d.Set("time_period_end", TimePeriodTimestampToString(budget.TimePeriod.End))
		d.Set("time_period_start", TimePeriodTimestampToString(budget.TimePeriod.Start))
	}

	d.Set("time_unit", budget.TimeUnit)

	notifications, err := FindNotificationsByAccountIDAndBudgetName(conn, accountID, budgetName)

	if tfresource.NotFound(err) {
		return nil
	}

	if err != nil {
		return fmt.Errorf("error reading Budget (%s) notifications: %w", d.Id(), err)
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

		subscribers, err := FindSubscribersByAccountIDBudgetNameAndNotification(conn, accountID, budgetName, notification)

		if tfresource.NotFound(err) {
			tfList = append(tfList, tfMap)
			continue
		}

		if err != nil {
			return fmt.Errorf("error reading Budget (%s) subscribers: %w", d.Id(), err)
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
		return fmt.Errorf("error setting notification: %w", err)
	}

	return nil
}

func resourceBudgetUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).BudgetsConn

	accountID, _, err := BudgetParseResourceID(d.Id())

	if err != nil {
		return err
	}

	budget, err := expandBudgetUnmarshal(d)
	if err != nil {
		return fmt.Errorf("could not create budget: %v", err)
	}

	_, err = conn.UpdateBudget(&budgets.UpdateBudgetInput{
		AccountId: aws.String(accountID),
		NewBudget: budget,
	})
	if err != nil {
		return fmt.Errorf("update budget failed: %v", err)
	}

	err = resourceBudgetNotificationsUpdate(d, meta)

	if err != nil {
		return fmt.Errorf("update budget notification failed: %v", err)
	}

	return resourceBudgetRead(d, meta)
}

func resourceBudgetDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).BudgetsConn

	accountID, budgetName, err := BudgetParseResourceID(d.Id())

	if err != nil {
		return err
	}

	log.Printf("[DEBUG] Deleting Budget: %s", d.Id())
	_, err = conn.DeleteBudget(&budgets.DeleteBudgetInput{
		AccountId:  aws.String(accountID),
		BudgetName: aws.String(budgetName),
	})

	if tfawserr.ErrCodeEquals(err, budgets.ErrCodeNotFoundException) {
		return nil
	}

	if err != nil {
		return fmt.Errorf("error deleting Budget (%s): %w", d.Id(), err)
	}

	return nil
}

func resourceBudgetNotificationsCreate(notifications []*budgets.Notification, subscribers [][]*budgets.Subscriber, budgetName string, accountID string, meta interface{}) error {
	conn := meta.(*conns.AWSClient).BudgetsConn

	for i, notification := range notifications {
		subscribers := subscribers[i]
		if len(subscribers) == 0 {
			return fmt.Errorf("Notification must have at least one subscriber!")
		}
		_, err := conn.CreateNotification(&budgets.CreateNotificationInput{
			BudgetName:   aws.String(budgetName),
			AccountId:    aws.String(accountID),
			Notification: notification,
			Subscribers:  subscribers,
		})

		if err != nil {
			return err
		}
	}

	return nil
}

func resourceBudgetNotificationsUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).BudgetsConn

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

			_, err := conn.DeleteNotification(input)

			if err != nil {
				return fmt.Errorf("error deleting Budget (%s) Notification: %s", d.Id(), err)
			}
		}

		err = resourceBudgetNotificationsCreate(addNotifications, addSubscribers, budgetName, accountID, meta)

		if err != nil {
			return fmt.Errorf("error creating Budget (%s) Notifications: %s", d.Id(), err)
		}
	}

	return nil
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

func expandBudgetUnmarshal(d *schema.ResourceData) (*budgets.Budget, error) {
	budgetName := d.Get("name").(string)
	budgetType := d.Get("budget_type").(string)
	budgetLimitAmount := d.Get("limit_amount").(string)
	budgetLimitUnit := d.Get("limit_unit").(string)
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

	budgetTimePeriodStart, err := TimePeriodTimestampFromString(d.Get("time_period_start").(string))

	if err != nil {
		return nil, err
	}

	budgetTimePeriodEnd, err := TimePeriodTimestampFromString(d.Get("time_period_end").(string))

	if err != nil {
		return nil, err
	}

	budget := &budgets.Budget{
		BudgetName: aws.String(budgetName),
		BudgetType: aws.String(budgetType),
		BudgetLimit: &budgets.Spend{
			Amount: aws.String(budgetLimitAmount),
			Unit:   aws.String(budgetLimitUnit),
		},
		TimePeriod: &budgets.TimePeriod{
			End:   budgetTimePeriodEnd,
			Start: budgetTimePeriodStart,
		},
		TimeUnit:    aws.String(budgetTimeUnit),
		CostFilters: budgetCostFilters,
	}

	if v, ok := d.GetOk("cost_types"); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
		budget.CostTypes = expandCostTypes(v.([]interface{})[0].(map[string]interface{}))
	}

	return budget, nil
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
