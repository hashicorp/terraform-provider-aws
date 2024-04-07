// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package ce

import (
	"context"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/costexplorer"
	awstypes "github.com/aws/aws-sdk-go-v2/service/costexplorer/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_ce_anomaly_subscription", name="Anomaly Subscription")
// @Tags(identifierAttribute="id")
func ResourceAnomalySubscription() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceAnomalySubscriptionCreate,
		ReadWithoutTimeout:   resourceAnomalySubscriptionRead,
		UpdateWithoutTimeout: resourceAnomalySubscriptionUpdate,
		DeleteWithoutTimeout: resourceAnomalySubscriptionDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Schema: map[string]*schema.Schema{
			"account_id": {
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				ValidateFunc: verify.ValidAccountID,
			},
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"frequency": {
				Type:             schema.TypeString,
				Required:         true,
				ValidateDiagFunc: enum.Validate[awstypes.AnomalySubscriptionFrequency](),
			},
			"monitor_arn_list": {
				Type:     schema.TypeList,
				Required: true,
				Elem: &schema.Schema{
					Type:         schema.TypeString,
					ValidateFunc: verify.ValidARN,
				},
			},
			"name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
				ValidateFunc: validation.All(
					validation.StringLenBetween(1, 1024),
					validation.StringMatch(regexache.MustCompile(`[\\S\\s]*`), "Must be a valid Anomaly Subscription Name matching expression: [\\S\\s]*")),
			},
			"subscriber": {
				Type:     schema.TypeSet,
				Required: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"address": {
							Type:     schema.TypeString,
							Required: true,
						},
						"type": {
							Type:             schema.TypeString,
							Required:         true,
							ValidateDiagFunc: enum.Validate[awstypes.SubscriberType](),
						},
					},
				},
			},
			"threshold_expression": {
				Type:     schema.TypeList,
				MaxItems: 1,
				Computed: true,
				Optional: true,
				Elem:     schemaCostCategoryRule(),
			},
			names.AttrTags:    tftags.TagsSchema(),
			names.AttrTagsAll: tftags.TagsSchemaComputed(),
		},

		CustomizeDiff: verify.SetTagsDiff,
	}
}

func resourceAnomalySubscriptionCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	conn := meta.(*conns.AWSClient).CEClient(ctx)

	input := &costexplorer.CreateAnomalySubscriptionInput{
		AnomalySubscription: &awstypes.AnomalySubscription{
			SubscriptionName: aws.String(d.Get("name").(string)),
			Frequency:        awstypes.AnomalySubscriptionFrequency(d.Get("frequency").(string)),
			MonitorArnList:   expandAnomalySubscriptionMonitorARNList(d.Get("monitor_arn_list").([]interface{})),
			Subscribers:      expandAnomalySubscriptionSubscribers(d.Get("subscriber").(*schema.Set).List()),
		},
		ResourceTags: getTagsIn(ctx),
	}

	if v, ok := d.GetOk("account_id"); ok {
		input.AnomalySubscription.AccountId = aws.String(v.(string))
	}

	if v, ok := d.GetOk("threshold_expression"); ok {
		input.AnomalySubscription.ThresholdExpression = expandCostExpression(v.([]interface{})[0].(map[string]interface{}))
	}

	resp, err := conn.CreateAnomalySubscription(ctx, input)

	if err != nil {
		return create.AppendDiagError(diags, names.CE, create.ErrActionCreating, ResNameAnomalySubscription, d.Id(), err)
	}

	if resp == nil || resp.SubscriptionArn == nil {
		return sdkdiag.AppendErrorf(diags, "creating Cost Explorer Anomaly Subscription resource (%s): empty output", d.Get("name").(string))
	}

	d.SetId(aws.ToString(resp.SubscriptionArn))

	return append(diags, resourceAnomalySubscriptionRead(ctx, d, meta)...)
}

func resourceAnomalySubscriptionRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	conn := meta.(*conns.AWSClient).CEClient(ctx)

	subscription, err := FindAnomalySubscriptionByARN(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		create.LogNotFoundRemoveState(names.CE, create.ErrActionReading, ResNameAnomalySubscription, d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return create.AppendDiagError(diags, names.CE, create.ErrActionReading, ResNameAnomalySubscription, d.Id(), err)
	}

	d.Set("account_id", subscription.AccountId)
	d.Set("arn", subscription.SubscriptionArn)
	d.Set("frequency", subscription.Frequency)
	d.Set("monitor_arn_list", subscription.MonitorArnList)
	d.Set("subscriber", flattenAnomalySubscriptionSubscribers(subscription.Subscribers))
	d.Set("name", subscription.SubscriptionName)

	if err = d.Set("threshold_expression", []interface{}{flattenCostCategoryRuleExpression(subscription.ThresholdExpression)}); err != nil {
		return create.AppendDiagError(diags, names.CE, "setting threshold_expression", ResNameAnomalySubscription, d.Id(), err)
	}

	return diags
}

func resourceAnomalySubscriptionUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	conn := meta.(*conns.AWSClient).CEClient(ctx)

	if d.HasChangesExcept("tags", "tags_All") {
		input := &costexplorer.UpdateAnomalySubscriptionInput{
			SubscriptionArn: aws.String(d.Id()),
		}

		if d.HasChange("frequency") {
			input.Frequency = awstypes.AnomalySubscriptionFrequency(d.Get("frequency").(string))
		}

		if d.HasChange("monitor_arn_list") {
			input.MonitorArnList = expandAnomalySubscriptionMonitorARNList(d.Get("monitor_arn_list").([]interface{}))
		}

		if d.HasChange("subscriber") {
			input.Subscribers = expandAnomalySubscriptionSubscribers(d.Get("subscriber").(*schema.Set).List())
		}

		if d.HasChange("threshold_expression") {
			input.ThresholdExpression = expandCostExpression(d.Get("threshold_expression").([]interface{})[0].(map[string]interface{}))
		}

		_, err := conn.UpdateAnomalySubscription(ctx, input)

		if err != nil {
			return create.AppendDiagError(diags, names.CE, create.ErrActionUpdating, ResNameAnomalySubscription, d.Id(), err)
		}
	}

	return append(diags, resourceAnomalySubscriptionRead(ctx, d, meta)...)
}

func resourceAnomalySubscriptionDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	conn := meta.(*conns.AWSClient).CEClient(ctx)

	_, err := conn.DeleteAnomalySubscription(ctx, &costexplorer.DeleteAnomalySubscriptionInput{SubscriptionArn: aws.String(d.Id())})

	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return diags
	}

	if err != nil {
		return create.AppendDiagError(diags, names.CE, create.ErrActionDeleting, ResNameAnomalySubscription, d.Id(), err)
	}

	return diags
}

func expandAnomalySubscriptionMonitorARNList(rawMonitorArnList []interface{}) []string {
	if len(rawMonitorArnList) == 0 {
		return nil
	}

	var monitorArns []string

	for _, arn := range rawMonitorArnList {
		monitorArns = append(monitorArns, arn.(string))
	}

	return monitorArns
}

func expandAnomalySubscriptionSubscribers(rawSubscribers []interface{}) []awstypes.Subscriber {
	if len(rawSubscribers) == 0 {
		return nil
	}

	var subscribers []awstypes.Subscriber

	for _, sub := range rawSubscribers {
		rawSubMap := sub.(map[string]interface{})
		subscriber := awstypes.Subscriber{Address: aws.String(rawSubMap["address"].(string)), Type: awstypes.SubscriberType(rawSubMap["type"].(string))}
		subscribers = append(subscribers, subscriber)
	}

	return subscribers
}

func flattenAnomalySubscriptionSubscribers(subscribers []awstypes.Subscriber) []interface{} {
	if subscribers == nil {
		return []interface{}{}
	}

	var rawSubscribers []interface{}
	for _, subscriber := range subscribers {
		rawSubscriber := map[string]interface{}{
			"address": aws.ToString(subscriber.Address),
			"type":    string(subscriber.Type),
		}

		rawSubscribers = append(rawSubscribers, rawSubscriber)
	}

	return rawSubscribers
}
