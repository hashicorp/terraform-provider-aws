// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package cloudfront

import (
	"context"
	"log"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/cloudfront"
	awstypes "github.com/aws/aws-sdk-go-v2/service/cloudfront/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

// @SDKResource("aws_cloudfront_monitoring_subscription", name="Monitoring Subscription")
func resourceMonitoringSubscription() *schema.Resource {
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
										Type:             schema.TypeString,
										Required:         true,
										ValidateDiagFunc: enum.Validate[awstypes.RealtimeMetricsSubscriptionStatus](),
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
	conn := meta.(*conns.AWSClient).CloudFrontClient(ctx)

	id := d.Get("distribution_id").(string)
	input := &cloudfront.CreateMonitoringSubscriptionInput{
		DistributionId: aws.String(id),
	}

	if v, ok := d.GetOk("monitoring_subscription"); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
		input.MonitoringSubscription = expandMonitoringSubscription(v.([]interface{})[0].(map[string]interface{}))
	}

	_, err := conn.CreateMonitoringSubscription(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating CloudFront Monitoring Subscription (%s): %s", id, err)
	}

	if d.IsNewResource() {
		d.SetId(id)
	}

	return append(diags, resourceMonitoringSubscriptionRead(ctx, d, meta)...)
}

func resourceMonitoringSubscriptionRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).CloudFrontClient(ctx)

	output, err := findMonitoringSubscriptionByDistributionID(ctx, conn, d.Id())

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
	conn := meta.(*conns.AWSClient).CloudFrontClient(ctx)

	log.Printf("[DEBUG] Deleting CloudFront Monitoring Subscription: %s", d.Id())
	_, err := conn.DeleteMonitoringSubscription(ctx, &cloudfront.DeleteMonitoringSubscriptionInput{
		DistributionId: aws.String(d.Id()),
	})

	if errs.IsA[*awstypes.NoSuchDistribution](err) || errs.IsA[*awstypes.NoSuchMonitoringSubscription](err) {
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

func findMonitoringSubscriptionByDistributionID(ctx context.Context, conn *cloudfront.Client, id string) (*cloudfront.GetMonitoringSubscriptionOutput, error) {
	input := &cloudfront.GetMonitoringSubscriptionInput{
		DistributionId: aws.String(id),
	}

	output, err := conn.GetMonitoringSubscription(ctx, input)

	if errs.IsA[*awstypes.NoSuchDistribution](err) || errs.IsA[*awstypes.NoSuchMonitoringSubscription](err) {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	return output, nil
}

func expandMonitoringSubscription(tfMap map[string]interface{}) *awstypes.MonitoringSubscription {
	if tfMap == nil {
		return nil
	}

	apiObject := &awstypes.MonitoringSubscription{}

	if v, ok := tfMap["realtime_metrics_subscription_config"].([]interface{}); ok && len(v) > 0 {
		apiObject.RealtimeMetricsSubscriptionConfig = expandRealtimeMetricsSubscriptionConfig(v[0].(map[string]interface{}))
	}

	return apiObject
}

func expandRealtimeMetricsSubscriptionConfig(tfMap map[string]interface{}) *awstypes.RealtimeMetricsSubscriptionConfig {
	if tfMap == nil {
		return nil
	}

	apiObject := &awstypes.RealtimeMetricsSubscriptionConfig{}

	if v, ok := tfMap["realtime_metrics_subscription_status"].(string); ok && v != "" {
		apiObject.RealtimeMetricsSubscriptionStatus = awstypes.RealtimeMetricsSubscriptionStatus(v)
	}

	return apiObject
}

func flattenMonitoringSubscription(apiObject *awstypes.MonitoringSubscription) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if v := flattenRealtimeMetricsSubscriptionConfig(apiObject.RealtimeMetricsSubscriptionConfig); len(v) > 0 {
		tfMap["realtime_metrics_subscription_config"] = []interface{}{v}
	}

	return tfMap
}

func flattenRealtimeMetricsSubscriptionConfig(apiObject *awstypes.RealtimeMetricsSubscriptionConfig) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{
		"realtime_metrics_subscription_status": apiObject.RealtimeMetricsSubscriptionStatus,
	}

	return tfMap
}
