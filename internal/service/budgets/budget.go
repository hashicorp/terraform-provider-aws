// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

// DONOTCOPY: Copying old resources spreads bad habits. Use skaff instead.

package budgets

import (
	"cmp"
	"context"
	"errors"
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/budgets"
	awstypes "github.com/aws/aws-sdk-go-v2/service/budgets/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	inttypes "github.com/hashicorp/terraform-provider-aws/internal/types"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
	"github.com/shopspring/decimal"
)

const filterExpressionDepth = 3

// @SDKResource("aws_budgets_budget", name="Budget")
// @Tags(identifierAttribute="arn")
// @Testing(existsType="github.com/aws/aws-sdk-go-v2/service/budgets/types;awstypes;awstypes.Budget")
func resourceBudget() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceBudgetCreate,
		ReadWithoutTimeout:   resourceBudgetRead,
		UpdateWithoutTimeout: resourceBudgetUpdate,
		DeleteWithoutTimeout: resourceBudgetDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		CustomizeDiff: validateFilterExpressionDiff,

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
			"billing_view_arn": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: verify.ValidARN,
			},
			"budget_type": {
				Type:             schema.TypeString,
				Required:         true,
				ValidateDiagFunc: enum.Validate[awstypes.BudgetType](),
			},
			"cost_filter": {
				Type:          schema.TypeSet,
				Optional:      true,
				Computed:      true,
				ConflictsWith: []string{"filter_expression"},
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
			"filter_expression": {
				Type:          schema.TypeList,
				Optional:      true,
				MaxItems:      1,
				ConflictsWith: []string{"cost_filter"},
				// AWS Budgets API enforces: "Expression nested depth cannot be more than 2" which is not mentioned in docs
				// Schema level 3 = AWS depth 2 (because operators added when level > 1)
				Elem: filterExpressionElem(filterExpressionDepth),
			},
		},
	}
}

func filterExpressionElem(level int) *schema.Resource {
	// This is the non-recursive part of the schema.
	expressionSchema := map[string]*schema.Schema{
		"cost_categories": {
			Type:     schema.TypeList,
			Optional: true,
			MaxItems: 1,
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					names.AttrKey: {
						Type:     schema.TypeString,
						Optional: true,
					},
					"match_options": {
						Type:     schema.TypeList,
						Optional: true,
						Elem: &schema.Schema{
							Type:             schema.TypeString,
							ValidateDiagFunc: enum.Validate[awstypes.MatchOption](),
						},
					},
					names.AttrValues: {
						Type:     schema.TypeList,
						Optional: true,
						MinItems: 1,
						Elem: &schema.Schema{
							Type:         schema.TypeString,
							ValidateFunc: validation.StringLenBetween(0, 1024),
						},
					},
				},
			},
		},
		"dimensions": {
			Type:     schema.TypeList,
			Optional: true,
			MaxItems: 1,
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					names.AttrKey: {
						Type:             schema.TypeString,
						Required:         true,
						ValidateDiagFunc: enum.Validate[awstypes.Dimension](),
					},
					"match_options": {
						Type:     schema.TypeList,
						Optional: true,
						Elem: &schema.Schema{
							Type:             schema.TypeString,
							ValidateDiagFunc: enum.Validate[awstypes.MatchOption](),
						},
					},
					names.AttrValues: {
						Type:     schema.TypeList,
						Required: true,
						MinItems: 1,
						Elem: &schema.Schema{
							Type:         schema.TypeString,
							ValidateFunc: validation.StringLenBetween(0, 1024),
						},
					},
				},
			},
		},
		names.AttrTags: {
			Type:     schema.TypeList,
			Optional: true,
			MaxItems: 1,
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					names.AttrKey: {
						Type:         schema.TypeString,
						Optional:     true,
						ValidateFunc: validation.StringLenBetween(0, 1024),
					},
					"match_options": {
						Type:     schema.TypeList,
						Optional: true,
						Elem: &schema.Schema{
							Type:             schema.TypeString,
							ValidateDiagFunc: enum.Validate[awstypes.MatchOption](),
						},
					},
					names.AttrValues: {
						Type:     schema.TypeList,
						Optional: true,
						MinItems: 1,
						Elem: &schema.Schema{
							Type:         schema.TypeString,
							ValidateFunc: validation.StringLenBetween(0, 1024),
						},
					},
				},
			},
		},
	}

	if level > 1 {
		// Add in the recursive part of the schema
		expressionSchema["and"] = &schema.Schema{
			Type:     schema.TypeList,
			Optional: true,
			Elem:     filterExpressionElem(level - 1),
		}
		expressionSchema["not"] = &schema.Schema{
			Type:     schema.TypeList,
			Optional: true,
			MaxItems: 1,
			Elem:     filterExpressionElem(level - 1),
		}
		expressionSchema["or"] = &schema.Schema{
			Type:     schema.TypeList,
			Optional: true,
			Elem:     filterExpressionElem(level - 1),
		}
	}

	return &schema.Resource{
		Schema: expressionSchema,
	}
}

func resourceBudgetCreate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	c := meta.(*conns.AWSClient)
	conn := c.BudgetsClient(ctx)

	budget, err := expandBudgetUnmarshal(d)
	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	name := create.Name(ctx, d.Get(names.AttrName).(string), d.Get(names.AttrNamePrefix).(string))
	budget.BudgetName = aws.String(name)

	accountID := cmp.Or(d.Get(names.AttrAccountID).(string), c.AccountID(ctx))
	input := budgets.CreateBudgetInput{
		AccountId:    aws.String(accountID),
		Budget:       budget,
		ResourceTags: getTagsIn(ctx),
	}
	_, err = conn.CreateBudget(ctx, &input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating Budget (%s): %s", name, err)
	}

	d.SetId(budgetCreateResourceID(accountID, aws.ToString(budget.BudgetName)))

	_, err = findWithDelay(ctx, func(context.Context) (*awstypes.Budget, error) {
		return findBudgetByTwoPartKey(ctx, conn, accountID, name)
	})

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading Budget (%s): %s", d.Id(), err)
	}

	notificationsRaw := d.Get("notification").(*schema.Set).List()
	notifications, subscribers := expandBudgetNotificationsUnmarshal(notificationsRaw)

	err = createBudgetNotifications(ctx, conn, notifications, subscribers, *budget.BudgetName, accountID)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating Budget (%s) notifications: %s", d.Id(), err)
	}

	return append(diags, resourceBudgetRead(ctx, d, meta)...)
}

func resourceBudgetRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	c := meta.(*conns.AWSClient)
	conn := c.BudgetsClient(ctx)

	accountID, budgetName, err := budgetParseResourceID(d.Id())
	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	budget, err := findBudgetByTwoPartKey(ctx, conn, accountID, budgetName)

	if !d.IsNewResource() && retry.NotFound(err) {
		log.Printf("[WARN] Budget (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading Budget (%s): %s", d.Id(), err)
	}

	d.Set(names.AttrAccountID, accountID)
	d.Set(names.AttrARN, budgetARN(ctx, c, accountID, budgetName))
	d.Set("budget_type", budget.BudgetType)
	d.Set("billing_view_arn", budget.BillingViewArn)

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

	if err := d.Set("filter_expression", flattenFilterExpression(budget.FilterExpression)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting filter_expression: %s", err)
	}

	notifications, err := findNotificationsByTwoPartKey(ctx, conn, accountID, budgetName)

	if retry.NotFound(err) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading Budget (%s) notifications: %s", d.Id(), err)
	}

	var tfList []any

	for _, notification := range notifications {
		tfMap := make(map[string]any)

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

		subscribers, err := findSubscribersByThreePartKey(ctx, conn, accountID, budgetName, notification)

		if retry.NotFound(err) {
			tfList = append(tfList, tfMap)
			continue
		}

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "reading Budget (%s) subscribers: %s", d.Id(), err)
		}

		var emailSubscribers []string
		var snsSubscribers []string

		for _, subscriber := range subscribers {
			switch subscriber.SubscriptionType {
			case awstypes.SubscriptionTypeSns:
				snsSubscribers = append(snsSubscribers, aws.ToString(subscriber.Address))
			case awstypes.SubscriptionTypeEmail:
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

func resourceBudgetUpdate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).BudgetsClient(ctx)

	accountID, _, err := budgetParseResourceID(d.Id())
	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	budget, err := expandBudgetUnmarshal(d)
	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	input := budgets.UpdateBudgetInput{
		AccountId: aws.String(accountID),
		NewBudget: budget,
	}
	_, err = conn.UpdateBudget(ctx, &input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "updating Budget (%s): %s", d.Id(), err)
	}

	err = updateBudgetNotifications(ctx, conn, d)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "updating Budget (%s) notifications: %s", d.Id(), err)
	}

	return append(diags, resourceBudgetRead(ctx, d, meta)...)
}

func resourceBudgetDelete(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).BudgetsClient(ctx)

	accountID, budgetName, err := budgetParseResourceID(d.Id())
	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	log.Printf("[DEBUG] Deleting Budget: %s", d.Id())
	input := budgets.DeleteBudgetInput{
		AccountId:  aws.String(accountID),
		BudgetName: aws.String(budgetName),
	}
	_, err = conn.DeleteBudget(ctx, &input)

	if errs.IsA[*awstypes.NotFoundException](err) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting Budget (%s): %s", d.Id(), err)
	}

	return diags
}

const budgetResourceIDSeparator = ":"

func budgetCreateResourceID(accountID, budgetName string) string {
	parts := []string{accountID, budgetName}
	id := strings.Join(parts, budgetResourceIDSeparator)

	return id
}

func budgetParseResourceID(id string) (string, string, error) {
	parts := strings.Split(id, budgetResourceIDSeparator)

	if len(parts) == 2 && parts[0] != "" && parts[1] != "" {
		return parts[0], parts[1], nil
	}

	return "", "", fmt.Errorf("unexpected format for ID (%[1]s), expected AccountID%[2]sBudgetName", id, budgetActionResourceIDSeparator)
}

func createBudgetNotifications(ctx context.Context, conn *budgets.Client, notifications []*awstypes.Notification, subscribers [][]awstypes.Subscriber, budgetName string, accountID string) error {
	for i, notification := range notifications {
		subscribers := subscribers[i]

		if len(subscribers) == 0 {
			return errors.New("Budget notification must have at least one subscriber")
		}

		input := budgets.CreateNotificationInput{
			AccountId:    aws.String(accountID),
			BudgetName:   aws.String(budgetName),
			Notification: notification,
			Subscribers:  subscribers,
		}
		_, err := conn.CreateNotification(ctx, &input)

		if err != nil {
			return err
		}
	}

	return nil
}

func updateBudgetNotifications(ctx context.Context, conn *budgets.Client, d *schema.ResourceData) error {
	accountID, budgetName, err := budgetParseResourceID(d.Id())

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
				return fmt.Errorf("deleting Budget (%s) notification: %w", d.Id(), err)
			}
		}

		err = createBudgetNotifications(ctx, conn, addNotifications, addSubscribers, budgetName, accountID)

		if err != nil {
			return fmt.Errorf("creating Budget (%s) notifications: %w", d.Id(), err)
		}
	}

	return nil
}

func findBudget(ctx context.Context, conn *budgets.Client, input *budgets.DescribeBudgetInput) (*awstypes.Budget, error) {
	output, err := conn.DescribeBudget(ctx, input)

	if errs.IsA[*awstypes.NotFoundException](err) {
		return nil, &retry.NotFoundError{
			LastError: err,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil || output.Budget == nil {
		return nil, tfresource.NewEmptyResultError()
	}

	return output.Budget, nil
}

func findBudgetByTwoPartKey(ctx context.Context, conn *budgets.Client, accountID, budgetName string) (*awstypes.Budget, error) {
	input := budgets.DescribeBudgetInput{
		AccountId:  aws.String(accountID),
		BudgetName: aws.String(budgetName),
	}

	return findBudget(ctx, conn, &input)
}

func findNotifications(ctx context.Context, conn *budgets.Client, input *budgets.DescribeNotificationsForBudgetInput) ([]awstypes.Notification, error) {
	var output []awstypes.Notification

	pages := budgets.NewDescribeNotificationsForBudgetPaginator(conn, input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if errs.IsA[*awstypes.NotFoundException](err) {
			return nil, &retry.NotFoundError{
				LastError: err,
			}
		}

		if err != nil {
			return nil, err
		}

		for _, v := range page.Notifications {
			if inttypes.IsZero(v) {
				continue
			}

			output = append(output, v)
		}
	}

	return output, nil
}

func findNotificationsByTwoPartKey(ctx context.Context, conn *budgets.Client, accountID, budgetName string) ([]awstypes.Notification, error) {
	input := budgets.DescribeNotificationsForBudgetInput{
		AccountId:  aws.String(accountID),
		BudgetName: aws.String(budgetName),
	}
	output, err := findNotifications(ctx, conn, &input)

	if err != nil {
		return nil, err
	}

	if len(output) == 0 {
		return nil, tfresource.NewEmptyResultError()
	}

	return output, nil
}

func findSubscribers(ctx context.Context, conn *budgets.Client, input *budgets.DescribeSubscribersForNotificationInput) ([]awstypes.Subscriber, error) {
	var output []awstypes.Subscriber

	pages := budgets.NewDescribeSubscribersForNotificationPaginator(conn, input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if errs.IsA[*awstypes.NotFoundException](err) {
			return nil, &retry.NotFoundError{
				LastError: err,
			}
		}

		if err != nil {
			return nil, err
		}

		for _, v := range page.Subscribers {
			if inttypes.IsZero(v) {
				continue
			}

			output = append(output, v)
		}
	}

	return output, nil
}

func findSubscribersByThreePartKey(ctx context.Context, conn *budgets.Client, accountID, budgetName string, notification awstypes.Notification) ([]awstypes.Subscriber, error) {
	input := budgets.DescribeSubscribersForNotificationInput{
		AccountId:    aws.String(accountID),
		BudgetName:   aws.String(budgetName),
		Notification: &notification,
	}
	output, err := findSubscribers(ctx, conn, &input)

	if err != nil {
		return nil, err
	}

	if len(output) == 0 {
		return nil, tfresource.NewEmptyResultError()
	}

	return output, nil
}

func flattenAutoAdjustData(autoAdjustData *awstypes.AutoAdjustData) []map[string]any {
	if autoAdjustData == nil {
		return []map[string]any{}
	}

	attrs := map[string]any{
		"auto_adjust_type":      string(autoAdjustData.AutoAdjustType),
		"last_auto_adjust_time": aws.ToTime(autoAdjustData.LastAutoAdjustTime).Format(time.RFC3339),
	}

	if *autoAdjustData.HistoricalOptions != (awstypes.HistoricalOptions{}) { // nosemgrep:ci.semgrep.aws.prefer-pointer-conversion-conditional
		attrs["historical_options"] = flattenHistoricalOptions(autoAdjustData.HistoricalOptions)
	}

	return []map[string]any{attrs}
}

func flattenHistoricalOptions(historicalOptions *awstypes.HistoricalOptions) []map[string]any {
	if historicalOptions == nil {
		return []map[string]any{}
	}

	attrs := map[string]any{
		"budget_adjustment_period":   int64(aws.ToInt32(historicalOptions.BudgetAdjustmentPeriod)),
		"lookback_available_periods": int64(aws.ToInt32(historicalOptions.LookBackAvailablePeriods)),
	}

	return []map[string]any{attrs}
}

func flattenCostTypes(costTypes *awstypes.CostTypes) []map[string]any {
	if costTypes == nil {
		return []map[string]any{}
	}

	m := map[string]any{
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
	return []map[string]any{m}
}

func convertCostFiltersToMap(costFilters map[string][]string) []map[string]any {
	convertedCostFilters := make([]map[string]any, 0)
	for k, v := range costFilters {
		convertedCostFilter := make(map[string]any)
		filterValues := make([]string, 0)
		filterValues = append(filterValues, v...)

		convertedCostFilter[names.AttrValues] = filterValues
		convertedCostFilter[names.AttrName] = k
		convertedCostFilters = append(convertedCostFilters, convertedCostFilter)
	}

	return convertedCostFilters
}

func convertPlannedBudgetLimitsToSet(plannedBudgetLimits map[string]awstypes.Spend) []any {
	if plannedBudgetLimits == nil {
		return nil
	}

	convertedPlannedBudgetLimits := make([]any, len(plannedBudgetLimits))
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

func flattenFilterExpression(apiObject *awstypes.Expression) []map[string]any {
	if apiObject == nil {
		return nil
	}

	tfMap := make(map[string]any)

	if apiObject.And != nil {
		tfMap["and"] = flattenFilterExpressions(apiObject.And)
	}
	if apiObject.CostCategories != nil {
		tfMap["cost_categories"] = []map[string]any{flattenCostCategoryValues(apiObject.CostCategories)}
	}
	if apiObject.Dimensions != nil {
		tfMap["dimensions"] = []map[string]any{flattenExpressionDimensionValues(apiObject.Dimensions)}
	}
	if apiObject.Not != nil {
		tfMap["not"] = flattenFilterExpression(apiObject.Not)
	}
	if apiObject.Or != nil {
		tfMap["or"] = flattenFilterExpressions(apiObject.Or)
	}
	if apiObject.Tags != nil {
		tfMap[names.AttrTags] = []map[string]any{flattenTagValues(apiObject.Tags)}
	}

	return []map[string]any{tfMap}
}

func flattenFilterExpressions(apiObjects []awstypes.Expression) []map[string]any {
	if len(apiObjects) == 0 {
		return nil
	}

	result := make([]map[string]any, 0, len(apiObjects))
	for i := range apiObjects {
		result = append(result, flattenFilterExpression(&apiObjects[i])[0])
	}

	return result
}

func flattenCostCategoryValues(apiObject *awstypes.CostCategoryValues) map[string]any {
	if apiObject == nil {
		return nil
	}
	tfMap := make(map[string]any)
	if apiObject.Key != nil {
		tfMap[names.AttrKey] = aws.ToString(apiObject.Key)
	}
	if apiObject.MatchOptions != nil {
		tfMap["match_options"] = flex.FlattenStringyValueList(apiObject.MatchOptions)
	}
	if apiObject.Values != nil {
		tfMap[names.AttrValues] = apiObject.Values
	}
	return tfMap
}

func flattenExpressionDimensionValues(apiObject *awstypes.ExpressionDimensionValues) map[string]any {
	if apiObject == nil {
		return nil
	}
	tfMap := make(map[string]any)
	tfMap[names.AttrKey] = string(apiObject.Key)
	if apiObject.MatchOptions != nil {
		tfMap["match_options"] = flex.FlattenStringyValueList(apiObject.MatchOptions)
	}
	if apiObject.Values != nil {
		tfMap[names.AttrValues] = apiObject.Values
	}
	return tfMap
}

func flattenTagValues(apiObject *awstypes.TagValues) map[string]any {
	if apiObject == nil {
		return nil
	}
	tfMap := make(map[string]any)
	if apiObject.Key != nil {
		tfMap[names.AttrKey] = aws.ToString(apiObject.Key)
	}
	if apiObject.MatchOptions != nil {
		tfMap["match_options"] = flex.FlattenStringyValueList(apiObject.MatchOptions)
	}
	if apiObject.Values != nil {
		tfMap[names.AttrValues] = apiObject.Values
	}
	return tfMap
}

func expandBudgetUnmarshal(d *schema.ResourceData) (*awstypes.Budget, error) {
	budgetName := d.Get(names.AttrName).(string)
	budgetType := d.Get("budget_type").(string)
	budgetTimeUnit := d.Get("time_unit").(string)
	var budgetCostFilters map[string][]string

	if costFilter, ok := d.GetOk("cost_filter"); ok {
		budgetCostFilters = make(map[string][]string)
		for _, v := range costFilter.(*schema.Set).List() {
			element := v.(map[string]any)
			key := element[names.AttrName].(string)
			for _, filterValue := range element[names.AttrValues].([]any) {
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
		TimeUnit: awstypes.TimeUnit(budgetTimeUnit),
	}

	if budgetCostFilters != nil {
		budget.CostFilters = budgetCostFilters
	}

	if v, ok := d.GetOk("auto_adjust_data"); ok && len(v.([]any)) > 0 && v.([]any)[0] != nil {
		budget.AutoAdjustData = expandAutoAdjustData(v.([]any)[0].(map[string]any))
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

	if v, ok := d.GetOk("billing_view_arn"); ok {
		budget.BillingViewArn = aws.String(v.(string))
	}

	if v, ok := d.GetOk("cost_types"); ok && len(v.([]any)) > 0 && v.([]any)[0] != nil {
		budget.CostTypes = expandCostTypes(v.([]any)[0].(map[string]any))
	}

	if v, ok := d.GetOk("filter_expression"); ok && len(v.([]any)) > 0 && v.([]any)[0] != nil {
		budget.FilterExpression = expandFilterExpression(v.([]any)[0].(map[string]any))
	}

	return budget, nil
}

func expandAutoAdjustData(tfMap map[string]any) *awstypes.AutoAdjustData {
	if tfMap == nil {
		return nil
	}

	apiObject := &awstypes.AutoAdjustData{}

	if v, ok := tfMap["auto_adjust_type"].(string); ok {
		apiObject.AutoAdjustType = awstypes.AutoAdjustType(v)
	}

	if v, ok := tfMap["historical_options"].([]any); ok && len(v) > 0 {
		apiObject.HistoricalOptions = expandHistoricalOptions(v)
	}

	return apiObject
}

func expandHistoricalOptions(l []any) *awstypes.HistoricalOptions {
	if len(l) == 0 || l[0] == nil {
		return nil
	}

	m := l[0].(map[string]any)

	apiObject := &awstypes.HistoricalOptions{}

	if v, ok := m["budget_adjustment_period"].(int); ok && v != 0 {
		apiObject.BudgetAdjustmentPeriod = aws.Int32(int32(v))
	}

	return apiObject
}

func expandCostTypes(tfMap map[string]any) *awstypes.CostTypes {
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

func expandFilterExpression(tfMap map[string]any) *awstypes.Expression {
	if tfMap == nil {
		return nil
	}

	apiObject := &awstypes.Expression{}

	if v, ok := tfMap["dimensions"].([]any); ok && len(v) > 0 {
		apiObject.Dimensions = expandExpressionDimensionValues(v[0].(map[string]any))
	}

	if v, ok := tfMap[names.AttrTags].([]any); ok && len(v) > 0 {
		apiObject.Tags = expandTagValues(v[0].(map[string]any))
	}

	if v, ok := tfMap["cost_categories"].([]any); ok && len(v) > 0 {
		apiObject.CostCategories = expandCostCategoryValues(v[0].(map[string]any))
	}

	if v, ok := tfMap["and"].([]any); ok && len(v) > 0 {
		apiObject.And = make([]awstypes.Expression, 0, len(v))
		for _, sub := range v {
			apiObject.And = append(apiObject.And, *expandFilterExpression(sub.(map[string]any)))
		}
	}

	if v, ok := tfMap["or"].([]any); ok && len(v) > 0 {
		apiObject.Or = make([]awstypes.Expression, 0, len(v))
		for _, sub := range v {
			apiObject.Or = append(apiObject.Or, *expandFilterExpression(sub.(map[string]any)))
		}
	}

	if v, ok := tfMap["not"].([]any); ok && len(v) > 0 && v[0] != nil {
		apiObject.Not = expandFilterExpression(v[0].(map[string]any))
	}

	return apiObject
}

func expandCostCategoryValues(tfMap map[string]any) *awstypes.CostCategoryValues {
	if tfMap == nil {
		return nil
	}

	apiObject := &awstypes.CostCategoryValues{}

	if v, ok := tfMap[names.AttrKey].(string); ok {
		apiObject.Key = aws.String(v)
	}

	if v, ok := tfMap["match_options"].([]any); ok && len(v) > 0 {
		apiObject.MatchOptions = make([]awstypes.MatchOption, len(v))
		for i, option := range v {
			apiObject.MatchOptions[i] = awstypes.MatchOption(option.(string))
		}
	}

	if v, ok := tfMap[names.AttrValues].([]any); ok && len(v) > 0 {
		apiObject.Values = make([]string, len(v))
		for i, value := range v {
			apiObject.Values[i] = value.(string)
		}
	}

	return apiObject
}

func expandExpressionDimensionValues(tfMap map[string]any) *awstypes.ExpressionDimensionValues {
	if tfMap == nil {
		return nil
	}

	apiObject := &awstypes.ExpressionDimensionValues{}

	if v, ok := tfMap[names.AttrKey].(string); ok {
		apiObject.Key = awstypes.Dimension(v)
	}

	if v, ok := tfMap["match_options"].([]any); ok && len(v) > 0 {
		apiObject.MatchOptions = make([]awstypes.MatchOption, len(v))
		for i, option := range v {
			apiObject.MatchOptions[i] = awstypes.MatchOption(option.(string))
		}
	}

	if v, ok := tfMap[names.AttrValues].([]any); ok && len(v) > 0 {
		apiObject.Values = make([]string, len(v))
		for i, value := range v {
			apiObject.Values[i] = value.(string)
		}
	}

	return apiObject
}

func expandTagValues(tfMap map[string]any) *awstypes.TagValues {
	if tfMap == nil {
		return nil
	}

	apiObject := &awstypes.TagValues{}

	if v, ok := tfMap[names.AttrKey].(string); ok {
		apiObject.Key = aws.String(v)
	}

	if v, ok := tfMap["match_options"].([]any); ok && len(v) > 0 {
		apiObject.MatchOptions = make([]awstypes.MatchOption, len(v))
		for i, option := range v {
			apiObject.MatchOptions[i] = awstypes.MatchOption(option.(string))
		}
	}

	if v, ok := tfMap[names.AttrValues].([]any); ok && len(v) > 0 {
		apiObject.Values = make([]string, len(v))
		for i, value := range v {
			apiObject.Values[i] = value.(string)
		}
	}

	return apiObject
}

func expandPlannedBudgetLimitsUnmarshal(plannedBudgetLimitsRaw []any) (map[string]awstypes.Spend, error) {
	plannedBudgetLimits := make(map[string]awstypes.Spend, len(plannedBudgetLimitsRaw))

	for _, plannedBudgetLimit := range plannedBudgetLimitsRaw {
		plannedBudgetLimit := plannedBudgetLimit.(map[string]any)

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

func expandBudgetNotificationsUnmarshal(notificationsRaw []any) ([]*awstypes.Notification, [][]awstypes.Subscriber) {
	notifications := make([]*awstypes.Notification, len(notificationsRaw))
	subscribersForNotifications := make([][]awstypes.Subscriber, len(notificationsRaw))
	for i, notificationRaw := range notificationsRaw {
		notificationRaw := notificationRaw.(map[string]any)
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

func expandSubscribers(rawList any, subscriptionType awstypes.SubscriptionType) []awstypes.Subscriber {
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

func validTimePeriodTimestamp(v any, k string) (ws []string, errors []error) {
	_, err := time.Parse(timePeriodLayout, v.(string))

	if err != nil {
		errors = append(errors, fmt.Errorf("%q cannot be parsed as %q: %w", k, timePeriodLayout, err))
	}

	return
}

func validateFilterExpressionDiff(_ context.Context, diff *schema.ResourceDiff, _ any) error {
	if v, ok := diff.GetOk("filter_expression"); ok && len(v.([]any)) > 0 && v.([]any)[0] != nil {
		if err := validateFilterExpressionForAbsent(v.([]any)[0].(map[string]any), "filter_expression"); err != nil {
			return err
		}
	}
	return nil
}

func validateFilterExpressionForAbsent(expr map[string]any, path string) error {
	// AWS error WITHOUT 'values' in combination with ABSENT : Missing required parameter "Values"
	// AWS error WITH 'values' in combination with ABSENT : [Dimensions|Tags|CostCategories] expression must not have values set when ABSENT is provided
	if dims, ok := expr["dimensions"].([]any); ok {
		for i, dim := range dims {
			if dimMap, ok := dim.(map[string]any); ok {
				if matchOpts, ok := dimMap["match_options"].([]any); ok {
					for _, opt := range matchOpts {
						if opt.(string) == string(awstypes.MatchOptionAbsent) {
							return fmt.Errorf("%s.dimensions[%d]: ABSENT match_option is not supported", path, i)
						}
					}
				}
			}
		}
	}

	if tags, ok := expr[names.AttrTags].([]any); ok {
		for i, tag := range tags {
			if tagMap, ok := tag.(map[string]any); ok {
				if matchOpts, ok := tagMap["match_options"].([]any); ok {
					for _, opt := range matchOpts {
						if opt.(string) == string(awstypes.MatchOptionAbsent) {
							return fmt.Errorf("%s.tags[%d]: ABSENT match_option is not supported", path, i)
						}
					}
				}
			}
		}
	}

	if costCats, ok := expr["cost_categories"].([]any); ok {
		for i, cc := range costCats {
			if ccMap, ok := cc.(map[string]any); ok {
				if matchOpts, ok := ccMap["match_options"].([]any); ok {
					for _, opt := range matchOpts {
						if opt.(string) == string(awstypes.MatchOptionAbsent) {
							return fmt.Errorf("%s.cost_categories[%d]: ABSENT match_option is not supported", path, i)
						}
					}
				}
			}
		}
	}

	if andExprs, ok := expr["and"].([]any); ok {
		for i, andExpr := range andExprs {
			if andMap, ok := andExpr.(map[string]any); ok {
				if err := validateFilterExpressionForAbsent(andMap, fmt.Sprintf("%s.and[%d]", path, i)); err != nil {
					return err
				}
			}
		}
	}

	if orExprs, ok := expr["or"].([]any); ok {
		for i, orExpr := range orExprs {
			if orMap, ok := orExpr.(map[string]any); ok {
				if err := validateFilterExpressionForAbsent(orMap, fmt.Sprintf("%s.or[%d]", path, i)); err != nil {
					return err
				}
			}
		}
	}

	if notExprs, ok := expr["not"].([]any); ok && len(notExprs) > 0 {
		if notMap, ok := notExprs[0].(map[string]any); ok {
			if err := validateFilterExpressionForAbsent(notMap, fmt.Sprintf("%s.not", path)); err != nil {
				return err
			}
		}
	}

	return nil
}

func budgetARN(ctx context.Context, c *conns.AWSClient, accountID, budgetName string) string {
	return c.GlobalARNWithAccount(ctx, "budgets", accountID, "budget/"+budgetName)
}
