package ce

import (
	"context"
	"regexp"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/costexplorer"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func ResourceAnomalySubscription() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceAnomalySubscriptionCreate,
		ReadContext:   resourceAnomalySubscriptionRead,
		UpdateContext: resourceAnomalySubscriptionUpdate,
		DeleteContext: resourceAnomalySubscriptionDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Schema: map[string]*schema.Schema{
			"account_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"frequency": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringInSlice([]string{costexplorer.AnomalySubscriptionFrequencyDaily, costexplorer.AnomalySubscriptionFrequencyImmediate, costexplorer.AnomalySubscriptionFrequencyWeekly}, false),
			},
			"monitor_arn_list": {
				Type:     schema.TypeList,
				Required: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
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
							Type:         schema.TypeString,
							Required:     true,
							ValidateFunc: validation.StringInSlice([]string{costexplorer.SubscriberTypeEmail, costexplorer.SubscriberTypeSns}, false),
						},
					},
				},
			},
			"threshold": {
				Type:     schema.TypeFloat,
				Required: true,
			},
			"name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
				ValidateFunc: validation.All(
					validation.StringLenBetween(1, 1024),
					validation.StringMatch(regexp.MustCompile(`[\\S\\s]*`), "Must be a valid Anomaly Subscription Name matching expression: [\\S\\s]*")),
			},
			// "tags":     tftags.TagsSchema(),
			// "tags_all": tftags.TagsSchemaComputed(),
		},

		// CustomizeDiff: verify.SetTagsDiff,
	}
}

func resourceAnomalySubscriptionCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).CEConn
	// defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	// tags := defaultTagsConfig.MergeTags(tftags.New(d.Get("tags").(map[string]interface{})))

	input := &costexplorer.CreateAnomalySubscriptionInput{
		AnomalySubscription: &costexplorer.AnomalySubscription{
			SubscriptionName: aws.String(d.Get("name").(string)),
			Frequency:        aws.String(d.Get("frequency").(string)),
			MonitorArnList:   aws.StringSlice(expandAnomalySubscriptionMonitorArnList(d.Get("monitor_arn_list").([]interface{}))),
			Subscribers:      expandAnomalySubscriptionSubscribers(d.Get("subscriber").(*schema.Set).List()),
			Threshold:        aws.Float64(d.Get("threshold").(float64)),
		},
	}

	// if len(tags) > 0 {
	// 	input.ResourceTags = Tags(tags.IgnoreAWS())
	// }

	resp, err := conn.CreateAnomalySubscriptionWithContext(ctx, input)

	if err != nil {
		return names.DiagError(names.CE, names.ErrActionUpdating, ResAnomalySubscription, d.Id(), err)
	}

	d.SetId(aws.StringValue(resp.SubscriptionArn))

	return resourceAnomalySubscriptionRead(ctx, d, meta)
}

func resourceAnomalySubscriptionRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).CEConn

	resp, err := conn.GetAnomalySubscriptionsWithContext(ctx, &costexplorer.GetAnomalySubscriptionsInput{SubscriptionArnList: aws.StringSlice([]string{d.Id()})})

	if !d.IsNewResource() && len(resp.AnomalySubscriptions) < 1 {
		names.LogNotFoundRemoveState(names.CE, names.ErrActionReading, ResAnomalySubscription, d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return names.DiagError(names.CE, names.ErrActionReading, ResAnomalySubscription, d.Id(), err)
	}

	anomalySubscription := resp.AnomalySubscriptions[0]

	d.Set("account_id", anomalySubscription.AccountId)
	d.Set("arn", anomalySubscription.SubscriptionArn)
	d.Set("frequency", anomalySubscription.Frequency)
	d.Set("monitor_arn_list", anomalySubscription.MonitorArnList)
	d.Set("subscriber", flattenAnomalySubscriptionSubscribers(anomalySubscription.Subscribers))
	d.Set("threshold", anomalySubscription.Threshold)
	d.Set("name", anomalySubscription.SubscriptionName)

	// tags, err := ListTags(conn, aws.StringValue(anomalySubscription.MonitorArn))

	// if err != nil {
	// 	return names.DiagError(names.CE, names.ErrActionReading, ResAnomalyMonitor, d.Id(), err)
	// }

	// //lintignore:AWSR002
	// if err := d.Set("tags", tags.RemoveDefaultConfig(defaultTagsConfig).Map()); err != nil {
	// 	return names.DiagError(names.CE, names.ErrActionUpdating, ResAnomalyMonitor, d.Id(), err)
	// }

	// if err := d.Set("tags_all", tags.Map()); err != nil {
	// 	return names.DiagError(names.CE, names.ErrActionUpdating, ResAnomalyMonitor, d.Id(), err)
	// }

	return nil
}

func resourceAnomalySubscriptionUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).CEConn
	requestUpdate := false

	input := &costexplorer.UpdateAnomalySubscriptionInput{
		SubscriptionArn: aws.String(d.Id()),
	}

	if d.HasChange("frequency") {
		input.Frequency = aws.String(d.Get("frequency").(string))
		requestUpdate = true
	}

	if d.HasChange("monitor_arn_list") {
		input.MonitorArnList = aws.StringSlice(d.Get("monitor_arn_list").([]string))
		requestUpdate = true
	}

	if d.HasChange("subscriber") {
		input.Subscribers = expandAnomalySubscriptionSubscribers(d.Get("subscriber").([]interface{}))
		requestUpdate = true
	}

	if d.HasChange("threshold") {
		input.Threshold = aws.Float64(d.Get("threshold").(float64))
		requestUpdate = true
	}

	if requestUpdate {
		_, err := conn.UpdateAnomalySubscriptionWithContext(ctx, input)

		if err != nil {
			return names.DiagError(names.CE, names.ErrActionReading, ResAnomalySubscription, d.Id(), err)
		}
	}

	return resourceAnomalySubscriptionRead(ctx, d, meta)
}

func resourceAnomalySubscriptionDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).CEConn

	_, err := conn.DeleteAnomalySubscriptionWithContext(ctx, &costexplorer.DeleteAnomalySubscriptionInput{SubscriptionArn: aws.String(d.Id())})

	if err != nil && tfawserr.ErrCodeEquals(err, costexplorer.ErrCodeResourceNotFoundException) {
		return nil
	}

	if err != nil {
		return names.DiagError(names.CE, names.ErrActionDeleting, ResAnomalySubscription, d.Id(), err)
	}

	return nil
}

func expandAnomalySubscriptionMonitorArnList(rawMonitorArnList []interface{}) []string {
	if len(rawMonitorArnList) == 0 {
		return nil
	}

	var monitorArns []string

	for _, arn := range rawMonitorArnList {

		monitorArns = append(monitorArns, arn.(string))
	}

	return monitorArns
}

func expandAnomalySubscriptionSubscribers(rawSubscribers []interface{}) []*costexplorer.Subscriber {
	if len(rawSubscribers) == 0 {
		return nil
	}

	var subscribers []*costexplorer.Subscriber

	for _, sub := range rawSubscribers {
		rawSubMap := sub.(map[string]interface{})
		subscriber := &costexplorer.Subscriber{Address: aws.String(rawSubMap["address"].(string)), Type: aws.String(rawSubMap["type"].(string))}
		subscribers = append(subscribers, subscriber)
	}

	return subscribers
}

func flattenAnomalySubscriptionSubscribers(subscribers []*costexplorer.Subscriber) []interface{} {
	if subscribers == nil {
		return []interface{}{}
	}

	var rawSubscribers []interface{}
	for _, subscriber := range subscribers {
		rawSubscriber := map[string]interface{}{
			"address": aws.StringValue(subscriber.Address),
			"type":    aws.StringValue(subscriber.Type),
		}

		rawSubscribers = append(rawSubscribers, rawSubscriber)
	}

	return rawSubscribers
}
