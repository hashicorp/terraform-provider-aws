// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package servicequotas

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/servicequotas"
	awstypes "github.com/aws/aws-sdk-go-v2/service/servicequotas/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	tfslices "github.com/hashicorp/terraform-provider-aws/internal/slices"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKDataSource("aws_servicequotas_service_quotas", name="Service Quotas")
func dataSourceServiceQuotas() *schema.Resource { // nosemgrep:ci.servicequotas-in-func-name
	return &schema.Resource{
		ReadWithoutTimeout: dataSourceServiceQuotasRead, // nosemgrep:ci.servicequotas-in-func-name

		Schema: map[string]*schema.Schema{
			"service_code": {
				Type:     schema.TypeString,
				Required: true,
			},
			"quotas": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
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
							Type:     schema.TypeString,
							Computed: true,
						},
						"quota_name": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"service_code": {
							Type:     schema.TypeString,
							Computed: true,
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
				},
			},
		},
	}
}

func dataSourceServiceQuotasRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics { // nosemgrep:ci.servicequotas-in-func-name
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ServiceQuotasClient(ctx)

	serviceCode := d.Get("service_code").(string)

	input := servicequotas.ListAWSDefaultServiceQuotasInput{
		ServiceCode: aws.String(serviceCode),
	}

	quotas, err := findDefaultServiceQuotas(ctx, conn, &input, tfslices.PredicateTrue[*awstypes.ServiceQuota]())

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "listing Service Quotas Service Quotas (%s): %s", serviceCode, err)
	}

	d.SetId(serviceCode)

	if err := d.Set("quotas", flattenServiceQuotas(quotas)); err != nil { // nosemgrep:ci.servicequotas-in-func-name
		return sdkdiag.AppendErrorf(diags, "setting quotas: %s", err)
	}

	return diags
}

func flattenServiceQuotas(apiObjects []awstypes.ServiceQuota) []any { // nosemgrep:ci.servicequotas-in-func-name
	var tfList []any

	for _, apiObject := range apiObjects {
		tfList = append(tfList, map[string]any{
			"adjustable":           apiObject.Adjustable,
			names.AttrARN:          aws.ToString(apiObject.QuotaArn),
			names.AttrDefaultValue: aws.ToFloat64(apiObject.Value),
			"global_quota":         apiObject.GlobalQuota,
			"quota_code":           aws.ToString(apiObject.QuotaCode),
			"quota_name":           aws.ToString(apiObject.QuotaName),
			"service_code":         aws.ToString(apiObject.ServiceCode),
			names.AttrServiceName:  aws.ToString(apiObject.ServiceName),
			"usage_metric":         flattenMetricInfo(apiObject.UsageMetric),
			names.AttrValue:        aws.ToFloat64(apiObject.Value),
		})
	}

	return tfList
}
