// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

// DONOTCOPY: Copying old resources spreads bad habits. Use skaff instead.

package servicequotas

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/servicequotas"
	awstypes "github.com/aws/aws-sdk-go-v2/service/servicequotas/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	tfslices "github.com/hashicorp/terraform-provider-aws/internal/slices"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKDataSource("aws_servicequotas_service_quota", name="Service Quota")
func dataSourceServiceQuota() *schema.Resource {
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

func dataSourceServiceQuotaRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ServiceQuotasClient(ctx)

	quotaCode := d.Get("quota_code").(string)
	quotaName := d.Get("quota_name").(string)
	serviceCode := d.Get("service_code").(string)

	var err error
	var defaultQuota *awstypes.ServiceQuota

	// A Service Quota will always have a default value, but will only have a current value if it has been set.
	if quotaName != "" {
		defaultQuota, err = findDefaultServiceQuotaByServiceCodeAndQuotaName(ctx, conn, serviceCode, quotaName)
	} else {
		defaultQuota, err = findDefaultServiceQuotaByServiceCodeAndQuotaCode(ctx, conn, serviceCode, quotaCode)
	}

	if err != nil {
		return sdkdiag.AppendFromErr(diags, tfresource.SingularDataSourceFindError("Service Quotas Service Quota", err))
	}

	if quotaName != "" {
		quotaCode = aws.ToString(defaultQuota.QuotaCode)
	}

	arn := aws.ToString(defaultQuota.QuotaArn)
	d.SetId(arn)
	d.Set("adjustable", defaultQuota.Adjustable)
	d.Set(names.AttrARN, arn)
	d.Set(names.AttrDefaultValue, defaultQuota.Value)
	d.Set("global_quota", defaultQuota.GlobalQuota)
	d.Set("quota_code", defaultQuota.QuotaCode)
	d.Set("quota_name", defaultQuota.QuotaName)
	d.Set("service_code", defaultQuota.ServiceCode)
	d.Set(names.AttrServiceName, defaultQuota.ServiceName)
	if err := d.Set("usage_metric", flattenMetricInfo(defaultQuota.UsageMetric)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting usage_metric: %s", err)
	}
	d.Set(names.AttrValue, defaultQuota.Value)

	serviceQuota, err := findServiceQuotaByServiceCodeAndQuotaCode(ctx, conn, serviceCode, quotaCode)

	switch {
	case retry.NotFound(err):
	case err != nil:
		return sdkdiag.AppendErrorf(diags, "reading Service Quotas Service Quota (%s/%s): %s", serviceCode, quotaCode, err)
	default:
		d.Set(names.AttrARN, serviceQuota.QuotaArn)
		d.Set(names.AttrValue, serviceQuota.Value)
	}

	return diags
}

func findDefaultServiceQuotaByServiceCodeAndQuotaName(ctx context.Context, conn *servicequotas.Client, serviceCode, quotaName string) (*awstypes.ServiceQuota, error) {
	input := servicequotas.ListAWSDefaultServiceQuotasInput{
		ServiceCode: aws.String(serviceCode),
	}

	return findDefaultServiceQuota(ctx, conn, &input, func(v *awstypes.ServiceQuota) bool {
		return aws.ToString(v.QuotaName) == quotaName
	})
}

func findDefaultServiceQuota(ctx context.Context, conn *servicequotas.Client, input *servicequotas.ListAWSDefaultServiceQuotasInput, filter tfslices.Predicate[*awstypes.ServiceQuota]) (*awstypes.ServiceQuota, error) {
	output, err := findDefaultServiceQuotas(ctx, conn, input, filter)

	if err != nil {
		return nil, err
	}

	return tfresource.AssertSingleValueResult(output)
}

func findDefaultServiceQuotas(ctx context.Context, conn *servicequotas.Client, input *servicequotas.ListAWSDefaultServiceQuotasInput, filter tfslices.Predicate[*awstypes.ServiceQuota]) ([]awstypes.ServiceQuota, error) { // nosemgrep:ci.servicequotas-in-func-name
	var output []awstypes.ServiceQuota

	pages := servicequotas.NewListAWSDefaultServiceQuotasPaginator(conn, input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if errs.IsA[*awstypes.NoSuchResourceException](err) {
			return nil, &retry.NotFoundError{
				LastError: err,
			}
		}

		if err != nil {
			return nil, err
		}

		for _, v := range page.Quotas {
			if filter(&v) {
				output = append(output, v)
			}
		}
	}

	return output, nil
}
