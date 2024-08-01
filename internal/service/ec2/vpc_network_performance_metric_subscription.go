// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package ec2

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_vpc_network_performance_metric_subscription", name="VPC Network Performance Metric Subscription")
func resourceNetworkPerformanceMetricSubscription() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceNetworkPerformanceMetricSubscriptionCreate,
		ReadWithoutTimeout:   resourceNetworkPerformanceMetricSubscriptionRead,
		DeleteWithoutTimeout: resourceNetworkPerformanceMetricSubscriptionDelete,

		Schema: map[string]*schema.Schema{
			names.AttrDestination: {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"metric": {
				Type:             schema.TypeString,
				Optional:         true,
				ForceNew:         true,
				Default:          types.MetricTypeAggregateLatency,
				ValidateDiagFunc: enum.Validate[types.MetricType](),
			},
			"period": {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrSource: {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"statistic": {
				Type:             schema.TypeString,
				Optional:         true,
				ForceNew:         true,
				Default:          types.StatisticTypeP50,
				ValidateDiagFunc: enum.Validate[types.StatisticType](),
			},
		},
	}
}

func resourceNetworkPerformanceMetricSubscriptionCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).EC2Client(ctx)

	source := d.Get(names.AttrSource).(string)
	destination := d.Get(names.AttrDestination).(string)
	metric := d.Get("metric").(string)
	statistic := d.Get("statistic").(string)
	id := networkPerformanceMetricSubscriptionCreateResourceID(source, destination, metric, statistic)
	input := &ec2.EnableAwsNetworkPerformanceMetricSubscriptionInput{
		Destination: aws.String(destination),
		Metric:      types.MetricType(metric),
		Source:      aws.String(source),
		Statistic:   types.StatisticType(statistic),
	}

	_, err := conn.EnableAwsNetworkPerformanceMetricSubscription(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "enabling EC2 AWS Network Performance Metric Subscription (%s): %s", id, err)
	}

	d.SetId(id)

	return append(diags, resourceNetworkPerformanceMetricSubscriptionRead(ctx, d, meta)...)
}

func resourceNetworkPerformanceMetricSubscriptionRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).EC2Client(ctx)

	source, destination, metric, statistic, err := networkPerformanceMetricSubscriptionParseResourceID(d.Id())
	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	subscription, err := findNetworkPerformanceMetricSubscriptionByFourPartKey(ctx, conn, source, destination, metric, statistic)

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] EC2 AWS Network Performance Metric Subscription (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading EC2 AWS Network Performance Metric Subscription (%s): %s", d.Id(), err)
	}

	d.Set(names.AttrDestination, subscription.Destination)
	d.Set("metric", subscription.Metric)
	d.Set("period", subscription.Period)
	d.Set(names.AttrSource, subscription.Source)
	d.Set("statistic", subscription.Statistic)

	return diags
}

func resourceNetworkPerformanceMetricSubscriptionDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).EC2Client(ctx)

	source, destination, metric, statistic, err := networkPerformanceMetricSubscriptionParseResourceID(d.Id())
	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	log.Printf("[DEBUG] Deleting EC2 AWS Network Performance Metric Subscriptione: %s", d.Id())
	_, err = conn.DisableAwsNetworkPerformanceMetricSubscription(ctx, &ec2.DisableAwsNetworkPerformanceMetricSubscriptionInput{
		Destination: aws.String(destination),
		Metric:      types.MetricType(metric),
		Source:      aws.String(source),
		Statistic:   types.StatisticType(statistic),
	})

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "disabling EC2 AWS Network Performance Metric Subscription (%s): %s", d.Id(), err)
	}

	return diags
}

const networkPerformanceMetricSubscriptionRuleIDSeparator = "/"

func networkPerformanceMetricSubscriptionCreateResourceID(source, destination, metric, statistic string) string {
	parts := []string{source, destination, metric, statistic}
	id := strings.Join(parts, networkPerformanceMetricSubscriptionRuleIDSeparator)

	return id
}

func networkPerformanceMetricSubscriptionParseResourceID(id string) (string, string, string, string, error) {
	parts := strings.Split(id, networkPerformanceMetricSubscriptionRuleIDSeparator)

	if len(parts) == 4 && parts[0] != "" && parts[1] != "" && parts[2] != "" && parts[3] != "" {
		return parts[0], parts[1], parts[2], parts[3], nil
	}

	return "", "", "", "", fmt.Errorf("unexpected format for ID (%[1]s), expected source%[2]sdestination%[2]smetric%[2]sstatistic", id, networkPerformanceMetricSubscriptionRuleIDSeparator)
}
