// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package budgets

import (
	"context"
	"errors"
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/aws/arn"
	"github.com/aws/aws-sdk-go-v2/service/budgets"
	awstypes "github.com/aws/aws-sdk-go-v2/service/budgets/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
	"github.com/shopspring/decimal"
)

// @SDKResource("aws_budgets_budget")
// @Tags(identifierAttribute="arn")
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
			names.AttrAccountID: {
				Type:         schema.TypeString,
				Computed:     true,
				Optional:     true,
				ForceNew:     true,
				ValidateFunc: verify.ValidAccountID,
			},
			names.AttrARN: {
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
							Type:             schema.TypeString,
							Required:         true,
							ValidateDiagFunc: enum.Validate[awstypes.AutoAdjustType](),
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
				Type:             schema.TypeString,
				Required:         true,
				ValidateDiagFunc: enum.Validate[awstypes.BudgetType](),
			},
			"cost_filter": {
				Type:     schema.TypeSet,
				Optional: true,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						names.AttrName: {
							Type:     schema.TypeString,
							Required: true,
						},
						names.AttrValues: {
							Type:     schema.TypeList,
							Required: true,
							Elem: &schema.Schema{
								Type: schema.TypeString,
							},
						},
					},
				},
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
			names.AttrName: {
				Type:          schema.TypeString,
				Optional:      true,
				Computed:      true,
				ForceNew:      true,
				ConflictsWith: []string{names.AttrNamePrefix},
			},
			names.AttrNamePrefix: {
				Type:          schema.TypeString,
				Optional:      true,
				Computed:      true,
				ForceNew:      true,
				ConflictsWith: []string{names.AttrName},
			},
			"notification": {
				Type:     schema.TypeSet,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"comparison_operator": {
							Type:             schema.TypeString,
							Required:         true,
							ValidateDiagFunc: enum.Validate[awstypes.ComparisonOperator](),
						},
						"notification_type": {
							Type:             schema.TypeString,
							Required:         true,
							ValidateDiagFunc: enum.Validate[awstypes.NotificationType](),
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
							Type:             schema.TypeString,
							Required:         true,
							ValidateDiagFunc: enum.Validate[awstypes.ThresholdType](),
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
						names.AttrStartTime: {
							Type:         schema.TypeString,
							Required:     true,
							ValidateFunc: validTimePeriodTimestamp,
						},
						names.AttrUnit: {
							Type:     schema.TypeString,
							Required: true,
						},
					},
				},
				ConflictsWith: []string{"limit_amount", "limit_unit"},
			},
			names.AttrTags:    tftags.TagsSchema(),
			names.AttrTagsAll: tftags.TagsSchemaComputed(),
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
				Type:             schema.TypeString,
				Required:         true,
				ValidateDiagFunc: enum.Validate[awstypes.TimeUnit](),
			},
		},
		CustomizeDiff: verify.SetTagsDiff,
	}
}

func resourceBudgetCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	conn := meta.(*conns.AWSClient).BudgetsClient(ctx)

	budget, err := expandBudgetUnmarshal(d)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "expandBudgetUnmarshal: %s", err)
	}

	name := create.Name(d.Get(names.AttrName).(string), d.Get(names.AttrNamePrefix).(string))
	budget.BudgetName = aws.String(name)

	accountID := d.Get(names.AttrAccountID).(string)
	if accountID == "" {
		accountID = meta.(*conns.AWSClient).AccountID
	}

	_, err = conn.CreateBudget(ctx, &budgets.CreateBudgetInput{
		AccountId:    aws.String(accountID),
		Budget:       budget,
		ResourceTags: getTagsIn(ctx),
	})

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating Budget (%s): %s", name, err)
	}

	d.SetId(BudgetCreateResourceID(accountID, aws.ToString(budget.BudgetName)))

	notificationsRaw := d.Get("notification").(*schema.Set).List()
	notifications, subscribers := expandBudgetNotificationsUnmarshal(notificationsRaw)

	err = createBudgetNotifications(ctx, conn, notifications, subscribers, *budget.BudgetName, accountID)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating Budget (%s) notifications: %s", d.Id(), err)
	}

	return append(diags, resourceBudgetRead(ctx, d, meta)...)
}

func resourceBudgetRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	conn := meta.(*conns.AWSClient).BudgetsClient(ctx)

	accountID, budgetName, err := BudgetParseResourceID(d.Id())

	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	//budget, err := FindBudgetByTwoPartKey(ctx, conn, accountID, budgetName)

	budget, err := FindBudgetWithDelay(ctx, func() (*awstypes.Budget, error) {
		return FindBudgetByTwoPartKey(ctx, conn, accountID, budgetName)
	})

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] Budget (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading Budget (%s): %s", d.Id(), err)
	}

	d.Set(names.AttrAccountID, accountID)
	arn := arn.ARN{
		Partition: meta.(*conns.AWSClient).Partition,
		Service:   "budgets",
		AccountID: accountID,
		Resource:  "budget/" + budgetName,
	}
	d.Set(names.AttrARN, arn.String())
	d.Set("budget_type", budget.BudgetType)

	if err := d.Set("cost_filter", convertCostFiltersToMap(budget.CostFilters)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting cost_filter: %s", err)
	}
	if err := d.Set("cost_types", flattenCostTypes(budget.CostTypes)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting cost_types: %s", err)
	}
	if err := d.Set("auto_adjust_data", flattenAutoAdjustData(budget.AutoAdjustData)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting auto_adjust_data: %s", err)
	}

	if budget.BudgetLimit != nil {
		d.Set("limit_amount", budget.BudgetLimit.Amount)
		d.Set("limit_unit", budget.BudgetLimit.Unit)
	}

	d.Set(names.AttrName, budget.BudgetName)
	d.Set(names.AttrNamePrefix, create.NamePrefixFromName(aws.ToString(budget.BudgetName)))

	if err := d.Set("planned_limit", convertPlannedBudgetLimitsToSet(budget.PlannedBudgetLimits)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting planned_limit: %s", err)
	}

	if budget.TimePeriod != nil {
		d.Set("time_period_end", TimePeriodTimestampToString(budget.TimePeriod.End))
		d.Set("time_period_start", TimePeriodTimestampToString(budget.TimePeriod.Start))
	}

	d.Set("time_unit", budget.TimeUnit)

	notifications, err := findNotifications(ctx, conn, accountID, budgetName)

	if tfresource.NotFound(err) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading Budget (%s) notifications: %s", d.Id(), err)
	}

	var tfList []interface{}

	for _, notification := range notifications {
		tfMap := make(map[string]interface{})

		tfMap["comparison_operator"] = string(notification.ComparisonOperator)
		tfMap["threshold"] = notification.Threshold
		tfMap["notification_type"] = string(notification.NotificationType)

		if notification.ThresholdType == "" {
			// The AWS API doesn't seem to return a ThresholdType if it's set to PERCENTAGE
			// Set it manually to make behavior more predictable
			tfMap["threshold_type"] = awstypes.ThresholdTypePercentage
		} else {
			tfMap["threshold_type"] = string(notification.ThresholdType)
		}

		subscribers, err := findSubscribers(ctx, conn, accountID, budgetName, notification)

		if tfresource.NotFound(err) {
			tfList = append(tfList, tfMap)
			continue
		}

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "reading Budget (%s) subscribers: %s", d.Id(), err)
		}

		var emailSubscribers []string
		var snsSubscribers []string

		for _, subscriber := range subscribers {
			if subscriber.SubscriptionType == awstypes.SubscriptionTypeSns {
				snsSubscribers = append(snsSubscribers, aws.ToString(subscriber.Address))
			} else if subscriber.SubscriptionType == awstypes.SubscriptionTypeEmail {
				emailSubscribers = append(emailSubscribers, aws.ToString(subscriber.Address))
			}
		}

		tfMap["subscriber_email_addresses"] = emailSubscribers
		tfMap["subscriber_sns_topic_arns"] = snsSubscribers

		tfList = append(tfList, tfMap)
	}

	if err := d.Set("notification", tfList); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting notification: %s", err)
	}

	return diags
}

func resourceBudgetUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	conn := meta.(*conns.AWSClient).BudgetsClient(ctx)

	accountID, _, err := BudgetParseResourceID(d.Id())

	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	budget, err := expandBudgetUnmarshal(d)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "expandBudgetUnmarshal: %s", err)
	}

	_, err = conn.UpdateBudget(ctx, &budgets.UpdateBudgetInput{
		AccountId: aws.String(accountID),
		NewBudget: budget,
	})

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "updating Budget (%s): %s", d.Id(), err)
	}

	err = updateBudgetNotifications(ctx, conn, d)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "updating Budget (%s) notifications: %s", d.Id(), err)
	}

	return append(diags, resourceBudgetRead(ctx, d, meta)...)
}

func resourceBudgetDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	conn := meta.(*conns.AWSClient).BudgetsClient(ctx)

	accountID, budgetName, err := BudgetParseResourceID(d.Id())

	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	log.Printf("[DEBUG] Deleting Budget: %s", d.Id())
	_, err = conn.DeleteBudget(ctx, &budgets.DeleteBudgetInput{
		AccountId:  aws.String(accountID),
		BudgetName: aws.String(budgetName),
	})

	if errs.IsA[*awstypes.NotFoundException](err) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting Budget (%s): %s", d.Id(), err)
	}

	return diags
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

func FindBudgetByTwoPartKey(ctx context.Context, conn *budgets.Client, accountID, budgetName string) (*awstypes.Budget, error) {
	input := &budgets.DescribeBudgetInput{
		AccountId:  aws.String(accountID),
		BudgetName: aws.String(budgetName),
	}

	output, err := conn.DescribeBudget(ctx, input)

	if errs.IsA[*awstypes.NotFoundException](err) {
		return nil, &retry.NotFoundError{
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

func findNotifications(ctx context.Context, conn *budgets.Client, accountID, budgetName string) ([]awstypes.Notification, error) {
	input := &budgets.DescribeNotificationsForBudgetInput{
		AccountId:  aws.String(accountID),
		BudgetName: aws.String(budgetName),
	}
	var output []awstypes.Notification

	pages := budgets.NewDescribeNotificationsForBudgetPaginator(conn, input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if errs.IsA[*awstypes.NotFoundException](err) {
			return nil, &retry.NotFoundError{
				LastError:   err,
				LastRequest: input,
			}
		}

		if err != nil {
			return nil, err
		}

		for _, np := range page.Notifications {
			if np == (awstypes.Notification{}) {
				continue
			}

			output = append(output, np)
		}
	}

	if len(output) == 0 {
		return nil, &retry.NotFoundError{LastRequest: input}
	}

	return output, nil
}

func findSubscribers(ctx context.Context, conn *budgets.Client, accountID, budgetName string, notification awstypes.Notification) ([]awstypes.Subscriber, error) {
	input := &budgets.DescribeSubscribersForNotificationInput{
		AccountId:    aws.String(accountID),
		BudgetName:   aws.String(budgetName),
		Notification: &notification,
	}
	var output []awstypes.Subscriber

	pages := budgets.NewDescribeSubscribersForNotificationPaginator(conn, input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if errs.IsA[*awstypes.NotFoundException](err) {
			return nil, &retry.NotFoundError{
				LastError:   err,
				LastRequest: input,
			}
		}

		if err != nil {
			return nil, err
		}

		for _, subscriber := range page.Subscribers {
			if subscriber == (awstypes.Subscriber{}) {
				continue
			}

			output = append(output, subscriber)
		}
	}

	if len(output) == 0 {
		return nil, &retry.NotFoundError{LastRequest: input}
	}

	return output, nil
}

func createBudgetNotifications(ctx context.Context, conn *budgets.Client, notifications []*awstypes.Notification, subscribers [][]awstypes.Subscriber, budgetName string, accountID string) error {
	for i, notification := range notifications {
		subscribers := subscribers[i]

		if len(subscribers) == 0 {
			return errors.New("Budget notification must have at least one subscriber")
		}

		_, err := conn.CreateNotification(ctx, &budgets.CreateNotificationInput{
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

func updateBudgetNotifications(ctx context.Context, conn *budgets.Client, d *schema.ResourceData) error {
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

			_, err := conn.DeleteNotification(ctx, input)

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

func flattenAutoAdjustData(autoAdjustData *awstypes.AutoAdjustData) []map[string]interface{} {
	if autoAdjustData == nil {
		return []map[string]interface{}{}
	}

	attrs := map[string]interface{}{
		"auto_adjust_type":      string(autoAdjustData.AutoAdjustType),
		"last_auto_adjust_time": aws.ToTime(autoAdjustData.LastAutoAdjustTime).Format(time.RFC3339),
	}

	if *autoAdjustData.HistoricalOptions != (awstypes.HistoricalOptions{}) { // nosemgrep:ci.semgrep.aws.prefer-pointer-conversion-conditional
		attrs["historical_options"] = flattenHistoricalOptions(autoAdjustData.HistoricalOptions)
	}

	return []map[string]interface{}{attrs}
}

func flattenHistoricalOptions(historicalOptions *awstypes.HistoricalOptions) []map[string]interface{} {
	if historicalOptions == nil {
		return []map[string]interface{}{}
	}

	attrs := map[string]interface{}{
		"budget_adjustment_period":   int64(aws.ToInt32(historicalOptions.BudgetAdjustmentPeriod)),
		"lookback_available_periods": int64(aws.ToInt32(historicalOptions.LookBackAvailablePeriods)),
	}

	return []map[string]interface{}{attrs}
}

func flattenCostTypes(costTypes *awstypes.CostTypes) []map[string]interface{} {
	if costTypes == nil {
		return []map[string]interface{}{}
	}

	m := map[string]interface{}{
		"include_credit":             aws.ToBool(costTypes.IncludeCredit),
		"include_discount":           aws.ToBool(costTypes.IncludeDiscount),
		"include_other_subscription": aws.ToBool(costTypes.IncludeOtherSubscription),
		"include_recurring":          aws.ToBool(costTypes.IncludeRecurring),
		"include_refund":             aws.ToBool(costTypes.IncludeRefund),
		"include_subscription":       aws.ToBool(costTypes.IncludeSubscription),
		"include_support":            aws.ToBool(costTypes.IncludeSupport),
		"include_tax":                aws.ToBool(costTypes.IncludeTax),
		"include_upfront":            aws.ToBool(costTypes.IncludeUpfront),
		"use_amortized":              aws.ToBool(costTypes.UseAmortized),
		"use_blended":                aws.ToBool(costTypes.UseBlended),
	}
	return []map[string]interface{}{m}
}

func convertCostFiltersToMap(costFilters map[string][]string) []map[string]interface{} {
	convertedCostFilters := make([]map[string]interface{}, 0)
	for k, v := range costFilters {
		convertedCostFilter := make(map[string]interface{})
		filterValues := make([]string, 0)
		filterValues = append(filterValues, v...)

		convertedCostFilter[names.AttrValues] = filterValues
		convertedCostFilter[names.AttrName] = k
		convertedCostFilters = append(convertedCostFilters, convertedCostFilter)
	}

	return convertedCostFilters
}

func convertPlannedBudgetLimitsToSet(plannedBudgetLimits map[string]awstypes.Spend) []interface{} {
	if plannedBudgetLimits == nil {
		return nil
	}

	convertedPlannedBudgetLimits := make([]interface{}, len(plannedBudgetLimits))
	i := 0

	for k, v := range plannedBudgetLimits {
		if v == (awstypes.Spend{}) {
			return nil
		}

		startTime, err := TimePeriodSecondsToString(k)
		if err != nil {
			return nil
		}

		convertedPlannedBudgetLimit := make(map[string]string)
		convertedPlannedBudgetLimit["amount"] = aws.ToString(v.Amount)
		convertedPlannedBudgetLimit[names.AttrStartTime] = startTime
		convertedPlannedBudgetLimit[names.AttrUnit] = aws.ToString(v.Unit)

		convertedPlannedBudgetLimits[i] = convertedPlannedBudgetLimit
		i++
	}

	return convertedPlannedBudgetLimits
}

func expandBudgetUnmarshal(d *schema.ResourceData) (*awstypes.Budget, error) {
	budgetName := d.Get(names.AttrName).(string)
	budgetType := d.Get("budget_type").(string)
	budgetTimeUnit := d.Get("time_unit").(string)
	budgetCostFilters := make(map[string][]string)

	if costFilter, ok := d.GetOk("cost_filter"); ok {
		for _, v := range costFilter.(*schema.Set).List() {
			element := v.(map[string]interface{})
			key := element[names.AttrName].(string)
			for _, filterValue := range element[names.AttrValues].([]interface{}) {
				budgetCostFilters[key] = append(budgetCostFilters[key], filterValue.(string))
			}
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

	budget := &awstypes.Budget{
		BudgetName: aws.String(budgetName),
		BudgetType: awstypes.BudgetType(budgetType),
		TimePeriod: &awstypes.TimePeriod{
			End:   budgetTimePeriodEnd,
			Start: budgetTimePeriodStart,
		},
		TimeUnit:    awstypes.TimeUnit(budgetTimeUnit),
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
				budget.BudgetLimit = &awstypes.Spend{
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

func expandAutoAdjustData(tfMap map[string]interface{}) *awstypes.AutoAdjustData {
	if tfMap == nil {
		return nil
	}

	apiObject := &awstypes.AutoAdjustData{}

	if v, ok := tfMap["auto_adjust_type"].(string); ok {
		apiObject.AutoAdjustType = awstypes.AutoAdjustType(v)
	}

	if v, ok := tfMap["historical_options"].([]interface{}); ok && len(v) > 0 {
		apiObject.HistoricalOptions = expandHistoricalOptions(v)
	}

	return apiObject
}

func expandHistoricalOptions(l []interface{}) *awstypes.HistoricalOptions {
	if len(l) == 0 || l[0] == nil {
		return nil
	}

	m := l[0].(map[string]interface{})

	apiObject := &awstypes.HistoricalOptions{}

	if v, ok := m["budget_adjustment_period"].(int); ok && v != 0 {
		apiObject.BudgetAdjustmentPeriod = aws.Int32(int32(v))
	}

	return apiObject
}

func expandCostTypes(tfMap map[string]interface{}) *awstypes.CostTypes {
	if tfMap == nil {
		return nil
	}

	apiObject := &awstypes.CostTypes{}

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

func expandPlannedBudgetLimitsUnmarshal(plannedBudgetLimitsRaw []interface{}) (map[string]awstypes.Spend, error) {
	plannedBudgetLimits := make(map[string]awstypes.Spend, len(plannedBudgetLimitsRaw))

	for _, plannedBudgetLimit := range plannedBudgetLimitsRaw {
		plannedBudgetLimit := plannedBudgetLimit.(map[string]interface{})

		key, err := TimePeriodSecondsFromString(plannedBudgetLimit[names.AttrStartTime].(string))
		if err != nil {
			return nil, err
		}

		amount := plannedBudgetLimit["amount"].(string)
		unit := plannedBudgetLimit[names.AttrUnit].(string)

		plannedBudgetLimits[key] = awstypes.Spend{
			Amount: aws.String(amount),
			Unit:   aws.String(unit),
		}
	}

	return plannedBudgetLimits, nil
}

func expandBudgetNotificationsUnmarshal(notificationsRaw []interface{}) ([]*awstypes.Notification, [][]awstypes.Subscriber) {
	notifications := make([]*awstypes.Notification, len(notificationsRaw))
	subscribersForNotifications := make([][]awstypes.Subscriber, len(notificationsRaw))
	for i, notificationRaw := range notificationsRaw {
		notificationRaw := notificationRaw.(map[string]interface{})
		comparisonOperator := notificationRaw["comparison_operator"].(string)
		threshold := notificationRaw["threshold"].(float64)
		thresholdType := notificationRaw["threshold_type"].(string)
		notificationType := notificationRaw["notification_type"].(string)

		notifications[i] = &awstypes.Notification{
			ComparisonOperator: awstypes.ComparisonOperator(comparisonOperator),
			Threshold:          threshold,
			ThresholdType:      awstypes.ThresholdType(thresholdType),
			NotificationType:   awstypes.NotificationType(notificationType),
		}

		emailSubscribers := expandSubscribers(notificationRaw["subscriber_email_addresses"], awstypes.SubscriptionTypeEmail)
		snsSubscribers := expandSubscribers(notificationRaw["subscriber_sns_topic_arns"], awstypes.SubscriptionTypeSns)

		subscribersForNotifications[i] = append(emailSubscribers, snsSubscribers...)
	}
	return notifications, subscribersForNotifications
}

func expandSubscribers(rawList interface{}, subscriptionType awstypes.SubscriptionType) []awstypes.Subscriber {
	result := make([]awstypes.Subscriber, 0)
	addrs := flex.ExpandStringSet(rawList.(*schema.Set))
	for _, addr := range addrs {
		result = append(result, awstypes.Subscriber{
			SubscriptionType: subscriptionType,
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

	return time.Unix(startTime, 0).UTC().Format(timePeriodLayout), nil
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

	return aws.ToTime(ts).Format(timePeriodLayout)
}

func validTimePeriodTimestamp(v interface{}, k string) (ws []string, errors []error) {
	_, err := time.Parse(timePeriodLayout, v.(string))

	if err != nil {
		errors = append(errors, fmt.Errorf("%q cannot be parsed as %q: %w", k, timePeriodLayout, err))
	}

	return
}
