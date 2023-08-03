// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package servicequotas

import (
	"context"
	"fmt"
	"regexp"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/servicequotas"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
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
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"default_value": {
				Type:     schema.TypeFloat,
				Computed: true,
			},
			"quota_code": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
				ValidateFunc: validation.All(
					validation.StringLenBetween(1, 128),
					validation.StringMatch(regexp.MustCompile(`^[a-zA-Z]`), "must begin with alphabetic character"),
					validation.StringMatch(regexp.MustCompile(`^[a-zA-Z0-9-]+$`), "must contain only alphanumeric and hyphen characters"),
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
					validation.StringMatch(regexp.MustCompile(`^[a-zA-Z]`), "must begin with alphabetic character"),
					validation.StringMatch(regexp.MustCompile(`^[a-zA-Z0-9-]+$`), "must contain only alphanumeric and hyphen characters"),
				),
			},
			"service_name": {
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
									"type": {
										Type:     schema.TypeString,
										Computed: true,
									},
								},
							},
						},
						"metric_name": {
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
			"value": {
				Type:     schema.TypeFloat,
				Required: true,
			},
		},
	}
}

func resourceServiceQuotaCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ServiceQuotasConn(ctx)

	quotaCode := d.Get("quota_code").(string)
	serviceCode := d.Get("service_code").(string)
	value := d.Get("value").(float64)

	d.SetId(fmt.Sprintf("%s/%s", serviceCode, quotaCode))

	// A Service Quota will always have a default value, but will only have a current value if it has been set.
	// If it is not set, `GetServiceQuota` will return "NoSuchResourceException"
	defaultQuota, err := findServiceQuotaDefaultByID(ctx, conn, serviceCode, quotaCode)
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "getting Default Service Quota for (%s/%s): %s", serviceCode, quotaCode, err)
	}
	quotaValue := aws.Float64Value(defaultQuota.Value)

	serviceQuota, err := findServiceQuotaByID(ctx, conn, serviceCode, quotaCode)
	if err != nil && !tfresource.NotFound(err) {
		return sdkdiag.AppendErrorf(diags, "getting Service Quota for (%s/%s): %s", serviceCode, quotaCode, err)
	}
	if serviceQuota != nil {
		quotaValue = aws.Float64Value(serviceQuota.Value)
	}

	if value > quotaValue {
		input := &servicequotas.RequestServiceQuotaIncreaseInput{
			DesiredValue: aws.Float64(value),
			QuotaCode:    aws.String(quotaCode),
			ServiceCode:  aws.String(serviceCode),
		}

		output, err := conn.RequestServiceQuotaIncreaseWithContext(ctx, input)

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
	conn := meta.(*conns.AWSClient).ServiceQuotasConn(ctx)

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
	d.Set("arn", defaultQuota.QuotaArn)
	d.Set("default_value", defaultQuota.Value)
	d.Set("quota_code", defaultQuota.QuotaCode)
	d.Set("quota_name", defaultQuota.QuotaName)
	d.Set("service_code", defaultQuota.ServiceCode)
	d.Set("service_name", defaultQuota.ServiceName)
	d.Set("value", defaultQuota.Value)

	if err := d.Set("usage_metric", flattenUsageMetric(defaultQuota.UsageMetric)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting usage_metric for (%s/%s): %s", serviceCode, quotaCode, err)
	}

	serviceQuota, err := findServiceQuotaByID(ctx, conn, serviceCode, quotaCode)
	if err != nil && !tfresource.NotFound(err) {
		return sdkdiag.AppendErrorf(diags, "getting Service Quota for (%s/%s): %s", serviceCode, quotaCode, err)
	}

	if err == nil {
		d.Set("arn", serviceQuota.QuotaArn)
		d.Set("value", serviceQuota.Value)
	}

	requestID := d.Get("request_id").(string)

	if requestID != "" {
		input := &servicequotas.GetRequestedServiceQuotaChangeInput{
			RequestId: aws.String(requestID),
		}

		output, err := conn.GetRequestedServiceQuotaChangeWithContext(ctx, input)

		if tfawserr.ErrCodeEquals(err, servicequotas.ErrCodeNoSuchResourceException) {
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

		requestStatus := aws.StringValue(output.RequestedQuota.Status)
		d.Set("request_status", requestStatus)

		switch requestStatus {
		case servicequotas.RequestStatusApproved, servicequotas.RequestStatusCaseClosed, servicequotas.RequestStatusDenied:
			d.Set("request_id", "")
		case servicequotas.RequestStatusCaseOpened, servicequotas.RequestStatusPending:
			d.Set("value", output.RequestedQuota.DesiredValue)
		}
	}

	return diags
}

func resourceServiceQuotaUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ServiceQuotasConn(ctx)

	value := d.Get("value").(float64)
	serviceCode, quotaCode, err := resourceServiceQuotaParseID(d.Id())

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting Service Quota (%s): %s", d.Id(), err)
	}

	input := &servicequotas.RequestServiceQuotaIncreaseInput{
		DesiredValue: aws.Float64(value),
		QuotaCode:    aws.String(quotaCode),
		ServiceCode:  aws.String(serviceCode),
	}

	output, err := conn.RequestServiceQuotaIncreaseWithContext(ctx, input)

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
