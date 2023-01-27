package cloudfront

import (
	"context"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/cloudfront"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

func ResourceMonitoringSubscription() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceMonitoringSubscriptionCreate,
		ReadWithoutTimeout:   resourceMonitoringSubscriptionRead,
		UpdateWithoutTimeout: resourceMonitoringSubscriptionCreate,
		DeleteWithoutTimeout: resourceMonitoringSubscriptionDelete,
		Importer: &schema.ResourceImporter{
			StateContext: resourceMonitoringSubscriptionImport,
		},

		Schema: map[string]*schema.Schema{
			"distribution_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"monitoring_subscription": {
				Type:     schema.TypeList,
				Required: true,
				MinItems: 1,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"realtime_metrics_subscription_config": {
							Type:     schema.TypeList,
							Required: true,
							MinItems: 1,
							MaxItems: 1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"realtime_metrics_subscription_status": {
										Type:         schema.TypeString,
										Required:     true,
										ValidateFunc: validation.StringInSlice(cloudfront.RealtimeMetricsSubscriptionStatus_Values(), false),
									},
								},
							},
						},
					},
				},
			},
		},
	}
}

func resourceMonitoringSubscriptionCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).CloudFrontConn()

	id := d.Get("distribution_id").(string)
	input := &cloudfront.CreateMonitoringSubscriptionInput{
		DistributionId: aws.String(id),
	}

	if v, ok := d.GetOk("monitoring_subscription"); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
		input.MonitoringSubscription = expandMonitoringSubscription(v.([]interface{})[0].(map[string]interface{}))
	}

	log.Printf("[DEBUG] Creating CloudFront Monitoring Subscription: %s", input)
	_, err := conn.CreateMonitoringSubscriptionWithContext(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating CloudFront Monitoring Subscription (%s): %s", id, err)
	}

	d.SetId(id)

	return append(diags, resourceMonitoringSubscriptionRead(ctx, d, meta)...)
}

func resourceMonitoringSubscriptionRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).CloudFrontConn()

	output, err := FindMonitoringSubscriptionByDistributionID(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] CloudFront Monitoring Subscription (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading CloudFront Monitoring Subscription (%s): %s", d.Id(), err)
	}

	if output.MonitoringSubscription != nil {
		if err := d.Set("monitoring_subscription", []interface{}{flattenMonitoringSubscription(output.MonitoringSubscription)}); err != nil {
			return sdkdiag.AppendErrorf(diags, "setting monitoring_subscription: %s", err)
		}
	} else {
		d.Set("monitoring_subscription", nil)
	}

	return diags
}

func resourceMonitoringSubscriptionDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).CloudFrontConn()

	log.Printf("[DEBUG] Deleting CloudFront Monitoring Subscription (%s)", d.Id())
	_, err := conn.DeleteMonitoringSubscriptionWithContext(ctx, &cloudfront.DeleteMonitoringSubscriptionInput{
		DistributionId: aws.String(d.Id()),
	})

	if tfawserr.ErrCodeEquals(err, cloudfront.ErrCodeNoSuchDistribution) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting CloudFront Monitoring Subscription (%s): %s", d.Id(), err)
	}

	return diags
}

func resourceMonitoringSubscriptionImport(ctx context.Context, d *schema.ResourceData, m interface{}) ([]*schema.ResourceData, error) {
	d.Set("distribution_id", d.Id())
	return []*schema.ResourceData{d}, nil
}

func expandMonitoringSubscription(tfMap map[string]interface{}) *cloudfront.MonitoringSubscription {
	if tfMap == nil {
		return nil
	}

	apiObject := &cloudfront.MonitoringSubscription{}

	if v, ok := tfMap["realtime_metrics_subscription_config"].([]interface{}); ok && len(v) > 0 {
		apiObject.RealtimeMetricsSubscriptionConfig = expandRealtimeMetricsSubscriptionConfig(v[0].(map[string]interface{}))
	}

	return apiObject
}

func expandRealtimeMetricsSubscriptionConfig(tfMap map[string]interface{}) *cloudfront.RealtimeMetricsSubscriptionConfig {
	if tfMap == nil {
		return nil
	}

	apiObject := &cloudfront.RealtimeMetricsSubscriptionConfig{}

	if v, ok := tfMap["realtime_metrics_subscription_status"].(string); ok && v != "" {
		apiObject.RealtimeMetricsSubscriptionStatus = aws.String(v)
	}

	return apiObject
}

func flattenMonitoringSubscription(apiObject *cloudfront.MonitoringSubscription) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if v := flattenRealtimeMetricsSubscriptionConfig(apiObject.RealtimeMetricsSubscriptionConfig); len(v) > 0 {
		tfMap["realtime_metrics_subscription_config"] = []interface{}{v}
	}

	return tfMap
}

func flattenRealtimeMetricsSubscriptionConfig(apiObject *cloudfront.RealtimeMetricsSubscriptionConfig) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if v := apiObject.RealtimeMetricsSubscriptionStatus; v != nil {
		tfMap["realtime_metrics_subscription_status"] = aws.StringValue(v)
	}

	return tfMap
}
