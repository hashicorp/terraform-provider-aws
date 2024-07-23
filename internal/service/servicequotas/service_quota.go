// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package servicequotas

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/servicequotas"
	"github.com/aws/aws-sdk-go-v2/service/servicequotas/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_servicequotas_service_quota")
func ResourceServiceQuota() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceServiceQuotaCreate,
		ReadWithoutTimeout:   resourceServiceQuotaRead,
		UpdateWithoutTimeout: resourceServiceQuotaUpdate,
		DeleteWithoutTimeout: schema.NoopContext,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

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
			"quota_code": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
				ValidateFunc: validation.All(
					validation.StringLenBetween(1, 128),
					validation.StringMatch(regexache.MustCompile(`^[A-Za-z]`), "must begin with alphabetic character"),
					validation.StringMatch(regexache.MustCompile(`^[0-9A-Za-z-]+$`), "must contain only alphanumeric and hyphen characters"),
				),
			},
			"quota_name": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"request_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"request_status": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"service_code": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
				ValidateFunc: validation.All(
					validation.StringLenBetween(1, 63),
					validation.StringMatch(regexache.MustCompile(`^[A-Za-z]`), "must begin with alphabetic character"),
					validation.StringMatch(regexache.MustCompile(`^[0-9A-Za-z-]+$`), "must contain only alphanumeric and hyphen characters"),
				),
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
				Required: true,
			},
		},
	}
}

func resourceServiceQuotaCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ServiceQuotasClient(ctx)

	quotaCode := d.Get("quota_code").(string)
	serviceCode := d.Get("service_code").(string)
	value := d.Get(names.AttrValue).(float64)

	d.SetId(fmt.Sprintf("%s/%s", serviceCode, quotaCode))

	// A Service Quota will always have a default value, but will only have a current value if it has been set.
	// If it is not set, `GetServiceQuota` will return "NoSuchResourceException"
	defaultQuota, err := findServiceQuotaDefaultByID(ctx, conn, serviceCode, quotaCode)
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "getting Default Service Quota for (%s/%s): %s", serviceCode, quotaCode, err)
	}
	quotaValue := aws.ToFloat64(defaultQuota.Value)

	serviceQuota, err := findServiceQuotaByID(ctx, conn, serviceCode, quotaCode)
	if err != nil && !tfresource.NotFound(err) {
		return sdkdiag.AppendErrorf(diags, "getting Service Quota for (%s/%s): %s", serviceCode, quotaCode, err)
	}
	if serviceQuota != nil {
		quotaValue = aws.ToFloat64(serviceQuota.Value)
	}

	if value > quotaValue {
		input := &servicequotas.RequestServiceQuotaIncreaseInput{
			DesiredValue: aws.Float64(value),
			QuotaCode:    aws.String(quotaCode),
			ServiceCode:  aws.String(serviceCode),
		}

		output, err := conn.RequestServiceQuotaIncrease(ctx, input)

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "requesting Service Quota (%s) increase: %s", d.Id(), err)
		}

		if output == nil || output.RequestedQuota == nil {
			return sdkdiag.AppendErrorf(diags, "requesting Service Quota (%s) increase: empty result", d.Id())
		}

		d.Set("request_id", output.RequestedQuota.Id)
	}

	return append(diags, resourceServiceQuotaRead(ctx, d, meta)...)
}

func resourceServiceQuotaRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ServiceQuotasClient(ctx)

	serviceCode, quotaCode, err := resourceServiceQuotaParseID(d.Id())

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading Service Quota (%s): %s", d.Id(), err)
	}

	// A Service Quota will always have a default value, but will only have a current value if it has been set.
	// If it is not set, `GetServiceQuota` will return "NoSuchResourceException"
	defaultQuota, err := findServiceQuotaDefaultByID(ctx, conn, serviceCode, quotaCode)
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "getting Default Service Quota for (%s/%s): %s", serviceCode, quotaCode, err)
	}

	d.Set("adjustable", defaultQuota.Adjustable)
	d.Set(names.AttrARN, defaultQuota.QuotaArn)
	d.Set(names.AttrDefaultValue, defaultQuota.Value)
	d.Set("quota_code", defaultQuota.QuotaCode)
	d.Set("quota_name", defaultQuota.QuotaName)
	d.Set("service_code", defaultQuota.ServiceCode)
	d.Set(names.AttrServiceName, defaultQuota.ServiceName)
	d.Set(names.AttrValue, defaultQuota.Value)

	if err := d.Set("usage_metric", flattenUsageMetric(defaultQuota.UsageMetric)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting usage_metric for (%s/%s): %s", serviceCode, quotaCode, err)
	}

	serviceQuota, err := findServiceQuotaByID(ctx, conn, serviceCode, quotaCode)
	if err != nil && !tfresource.NotFound(err) {
		return sdkdiag.AppendErrorf(diags, "getting Service Quota for (%s/%s): %s", serviceCode, quotaCode, err)
	}

	if err == nil {
		d.Set(names.AttrARN, serviceQuota.QuotaArn)
		d.Set(names.AttrValue, serviceQuota.Value)
	}

	requestID := d.Get("request_id").(string)

	if requestID != "" {
		input := &servicequotas.GetRequestedServiceQuotaChangeInput{
			RequestId: aws.String(requestID),
		}

		output, err := conn.GetRequestedServiceQuotaChange(ctx, input)

		var nsr *types.NoSuchResourceException
		if errors.As(err, &nsr) {
			d.Set("request_id", "")
			d.Set("request_status", "")
			return diags
		}

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "getting Service Quotas Requested Service Quota Change (%s): %s", requestID, err)
		}

		if output == nil || output.RequestedQuota == nil {
			return sdkdiag.AppendErrorf(diags, "getting Service Quotas Requested Service Quota Change (%s): empty result", requestID)
		}

		d.Set("request_status", output.RequestedQuota.Status)

		switch output.RequestedQuota.Status {
		case types.RequestStatusApproved, types.RequestStatusCaseClosed, types.RequestStatusDenied:
			d.Set("request_id", "")
		case types.RequestStatusCaseOpened, types.RequestStatusPending:
			d.Set(names.AttrValue, output.RequestedQuota.DesiredValue)
		}
	}

	return diags
}

func resourceServiceQuotaUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ServiceQuotasClient(ctx)

	value := d.Get(names.AttrValue).(float64)
	serviceCode, quotaCode, err := resourceServiceQuotaParseID(d.Id())

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting Service Quota (%s): %s", d.Id(), err)
	}

	input := &servicequotas.RequestServiceQuotaIncreaseInput{
		DesiredValue: aws.Float64(value),
		QuotaCode:    aws.String(quotaCode),
		ServiceCode:  aws.String(serviceCode),
	}

	output, err := conn.RequestServiceQuotaIncrease(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "requesting Service Quota (%s) increase: %s", d.Id(), err)
	}

	if output == nil || output.RequestedQuota == nil {
		return sdkdiag.AppendErrorf(diags, "requesting Service Quota (%s) increase: empty result", d.Id())
	}

	d.Set("request_id", output.RequestedQuota.Id)

	return append(diags, resourceServiceQuotaRead(ctx, d, meta)...)
}

func resourceServiceQuotaParseID(id string) (string, string, error) {
	parts := strings.SplitN(id, "/", 2)

	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		return "", "", fmt.Errorf("unexpected format of ID (%s), expected SERVICE-CODE/QUOTA-CODE", id)
	}

	return parts[0], parts[1], nil
}
