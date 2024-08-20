// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package servicequotas

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/servicequotas/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKDataSource("aws_servicequotas_service_quota")
func DataSourceServiceQuota() *schema.Resource {
	return &schema.Resource{
		ReadWithoutTimeout: dataSourceServiceQuotaRead,

		Schema: map[string]*schema.Schema{
			"adjustable": {
				Type:     schema.TypeBool,
				Computed: true,
			},
			names.AttrARN: {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrDefaultValue: {
				Type:     schema.TypeFloat,
				Computed: true,
			},
			"global_quota": {
				Type:     schema.TypeBool,
				Computed: true,
			},
			"quota_code": {
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				ExactlyOneOf: []string{"quota_code", "quota_name"},
			},
			"quota_name": {
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				ExactlyOneOf: []string{"quota_code", "quota_name"},
			},
			"service_code": {
				Type:     schema.TypeString,
				Required: true,
			},
			names.AttrServiceName: {
				Type:     schema.TypeString,
				Computed: true,
			},
			"usage_metric": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"metric_dimensions": {
							Type:     schema.TypeList,
							Computed: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"class": {
										Type:     schema.TypeString,
										Computed: true,
									},
									"resource": {
										Type:     schema.TypeString,
										Computed: true,
									},
									"service": {
										Type:     schema.TypeString,
										Computed: true,
									},
									names.AttrType: {
										Type:     schema.TypeString,
										Computed: true,
									},
								},
							},
						},
						names.AttrMetricName: {
							Type:     schema.TypeString,
							Computed: true,
						},
						"metric_namespace": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"metric_statistic_recommendation": {
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
			},
			names.AttrValue: {
				Type:     schema.TypeFloat,
				Computed: true,
			},
		},
	}
}

func flattenUsageMetric(usageMetric *types.MetricInfo) []interface{} {
	if usageMetric == nil {
		return []interface{}{}
	}

	var usageMetrics []interface{}
	var metricDimensions []interface{}

	if usageMetric.MetricDimensions != nil && usageMetric.MetricDimensions["Service"] != "" {
		metricDimensions = append(metricDimensions, map[string]interface{}{
			"service":      usageMetric.MetricDimensions["Service"],
			"class":        usageMetric.MetricDimensions["Class"],
			names.AttrType: usageMetric.MetricDimensions["Type"],
			"resource":     usageMetric.MetricDimensions["Resource"],
		})
	} else {
		metricDimensions = append(metricDimensions, map[string]interface{}{})
	}

	usageMetrics = append(usageMetrics, map[string]interface{}{
		names.AttrMetricName:              usageMetric.MetricName,
		"metric_namespace":                usageMetric.MetricNamespace,
		"metric_statistic_recommendation": usageMetric.MetricStatisticRecommendation,
		"metric_dimensions":               metricDimensions,
	})

	return usageMetrics
}

func dataSourceServiceQuotaRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ServiceQuotasClient(ctx)

	quotaCode := d.Get("quota_code").(string)
	quotaName := d.Get("quota_name").(string)
	serviceCode := d.Get("service_code").(string)

	var err error
	var defaultQuota *types.ServiceQuota

	// A Service Quota will always have a default value, but will only have a current value if it has been set.
	// If it is not set, `GetServiceQuota` will return "NoSuchResourceException"
	if quotaName != "" {
		defaultQuota, err = findServiceQuotaDefaultByName(ctx, conn, serviceCode, quotaName)
		if err != nil {
			return sdkdiag.AppendErrorf(diags, "getting Default Service Quota for (%s/%s): %s", serviceCode, quotaName, err)
		}

		quotaCode = aws.ToString(defaultQuota.QuotaCode)
	} else {
		defaultQuota, err = findServiceQuotaDefaultByID(ctx, conn, serviceCode, quotaCode)
		if err != nil {
			return sdkdiag.AppendErrorf(diags, "getting Default Service Quota for (%s/%s): %s", serviceCode, quotaCode, err)
		}
	}

	d.SetId(aws.ToString(defaultQuota.QuotaArn))
	d.Set("adjustable", defaultQuota.Adjustable)
	d.Set(names.AttrARN, defaultQuota.QuotaArn)
	d.Set(names.AttrDefaultValue, defaultQuota.Value)
	d.Set("global_quota", defaultQuota.GlobalQuota)
	d.Set("quota_code", defaultQuota.QuotaCode)
	d.Set("quota_name", defaultQuota.QuotaName)
	d.Set("service_code", defaultQuota.ServiceCode)
	d.Set(names.AttrServiceName, defaultQuota.ServiceName)
	d.Set(names.AttrValue, defaultQuota.Value)

	if err := d.Set("usage_metric", flattenUsageMetric(defaultQuota.UsageMetric)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting usage_metric for (%s/%s): %s", serviceCode, quotaCode, err)
	}

	serviceQuota, err := findServiceQuotaByID(ctx, conn, serviceCode, quotaCode)
	if tfresource.NotFound(err) {
		return diags
	}
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "getting Service Quota for (%s/%s): %s", serviceCode, quotaCode, err)
	}

	d.Set(names.AttrARN, serviceQuota.QuotaArn)
	d.Set(names.AttrValue, serviceQuota.Value)

	return diags
}
