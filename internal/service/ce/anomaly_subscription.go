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
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
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
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				ValidateFunc: validation.StringMatch(regexp.MustCompile(`[\\S\\s]*`), "Must be a valid AWS Account ID matching expression: [\\S\\s]*"),
			},
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"frequency": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringInSlice(costexplorer.AnomalySubscriptionFrequency_Values(), false),
			},
			"monitor_arn_list": {
				Type:     schema.TypeList,
				Required: true,
				Elem: &schema.Schema{
					Type:         schema.TypeString,
					ValidateFunc: validation.StringMatch(regexp.MustCompile(`arn:aws[-a-z0-9]*:[a-z0-9]+:[-a-z0-9]*:[0-9]{12}:[-a-zA-Z0-9/:_]+`), "Must be a valid anomaly monitor ARN"),
				},
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
				Type:         schema.TypeFloat,
				Required:     true,
				ValidateFunc: validation.FloatAtLeast(0.0),
			},
			"name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
				ValidateFunc: validation.All(
					validation.StringLenBetween(1, 1024),
					validation.StringMatch(regexp.MustCompile(`[\\S\\s]*`), "Must be a valid Anomaly Subscription Name matching expression: [\\S\\s]*")),
			},
			"tags":     tftags.TagsSchema(),
			"tags_all": tftags.TagsSchemaComputed(),
		},

		CustomizeDiff: verify.SetTagsDiff,
	}
}

func resourceAnomalySubscriptionCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).CEConn
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	tags := defaultTagsConfig.MergeTags(tftags.New(d.Get("tags").(map[string]interface{})))

	input := &costexplorer.CreateAnomalySubscriptionInput{
		AnomalySubscription: &costexplorer.AnomalySubscription{
			SubscriptionName: aws.String(d.Get("name").(string)),
			Frequency:        aws.String(d.Get("frequency").(string)),
			MonitorArnList:   aws.StringSlice(expandAnomalySubscriptionMonitorARNList(d.Get("monitor_arn_list").([]interface{}))),
			Subscribers:      expandAnomalySubscriptionSubscribers(d.Get("subscriber").(*schema.Set).List()),
			Threshold:        aws.Float64(d.Get("threshold").(float64)),
		},
	}

	if v, ok := d.GetOk("account_id"); ok {
		input.AnomalySubscription.AccountId = aws.String(v.(string))
	}

	if len(tags) > 0 {
		input.ResourceTags = Tags(tags.IgnoreAWS())
	}

	resp, err := conn.CreateAnomalySubscriptionWithContext(ctx, input)

	if err != nil {
		return names.DiagError(names.CE, names.ErrActionCreating, ResAnomalySubscription, d.Id(), err)
	}

	if resp == nil || resp.SubscriptionArn == nil {
		return diag.Errorf("creating Cost Explorer Anomaly Subscription resource (%s): empty output", d.Get("name").(string))
	}

	d.SetId(aws.StringValue(resp.SubscriptionArn))

	return resourceAnomalySubscriptionRead(ctx, d, meta)
}

func resourceAnomalySubscriptionRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).CEConn
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	ignoreTagsConfig := meta.(*conns.AWSClient).IgnoreTagsConfig

	subscription, err := FindAnomalySubscriptionByARN(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		names.LogNotFoundRemoveState(names.CE, names.ErrActionReading, ResAnomalySubscription, d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return names.DiagError(names.CE, names.ErrActionReading, ResAnomalySubscription, d.Id(), err)
	}

	d.Set("account_id", subscription.AccountId)
	d.Set("arn", subscription.SubscriptionArn)
	d.Set("frequency", subscription.Frequency)
	d.Set("monitor_arn_list", subscription.MonitorArnList)
	d.Set("subscriber", flattenAnomalySubscriptionSubscribers(subscription.Subscribers))
	d.Set("threshold", subscription.Threshold)
	d.Set("name", subscription.SubscriptionName)

	tags, err := ListTags(conn, aws.StringValue(subscription.SubscriptionArn))
	tags = tags.IgnoreAWS().IgnoreConfig(ignoreTagsConfig)

	if err != nil {
		return names.DiagError(names.CE, names.ErrActionReading, ResTags, d.Id(), err)
	}

	//lintignore:AWSR002
	if err := d.Set("tags", tags.RemoveDefaultConfig(defaultTagsConfig).Map()); err != nil {
		return names.DiagError(names.CE, names.ErrActionUpdating, ResTags, d.Id(), err)
	}

	if err := d.Set("tags_all", tags.Map()); err != nil {
		return names.DiagError(names.CE, names.ErrActionUpdating, ResTags, d.Id(), err)
	}

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
		input.Subscribers = expandAnomalySubscriptionSubscribers(d.Get("subscriber").(*schema.Set).List())
		requestUpdate = true
	}

	if d.HasChange("threshold") {
		input.Threshold = aws.Float64(d.Get("threshold").(float64))
		requestUpdate = true
	}

	if d.HasChange("tags") {
		o, n := d.GetChange("tags")

		if err := UpdateTags(conn, d.Id(), o, n); err != nil {
			return names.DiagError(names.CE, names.ErrActionUpdating, ResTags, d.Id(), err)
		}
	}

	if d.HasChange("tags_all") {
		o, n := d.GetChange("tags_all")

		if err := UpdateTags(conn, d.Id(), o, n); err != nil {
			return names.DiagError(names.CE, names.ErrActionUpdating, ResTags, d.Id(), err)
		}
	}

	if requestUpdate {
		_, err := conn.UpdateAnomalySubscriptionWithContext(ctx, input)

		if err != nil {
			return names.DiagError(names.CE, names.ErrActionUpdating, ResAnomalySubscription, d.Id(), err)
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
