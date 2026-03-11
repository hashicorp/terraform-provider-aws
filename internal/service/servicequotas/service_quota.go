// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

// DONOTCOPY: Copying old resources spreads bad habits. Use skaff instead.

package servicequotas

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/servicequotas"
	awstypes "github.com/aws/aws-sdk-go-v2/service/servicequotas/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_servicequotas_service_quota", name="Service Quota")
func resourceServiceQuota() *schema.Resource {
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

func resourceServiceQuotaCreate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ServiceQuotasClient(ctx)

	serviceCode, quotaCode := d.Get("service_code").(string), d.Get("quota_code").(string)

	// A Service Quota will always have a default value, but will only have a current value if it has been set.
	defaultQuota, err := findDefaultServiceQuotaByServiceCodeAndQuotaCode(ctx, conn, serviceCode, quotaCode)
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading Service Quotas default Service Quota (%s/%s): %s", serviceCode, quotaCode, err)
	}

	quotaValue := aws.ToFloat64(defaultQuota.Value)

	serviceQuota, err := findServiceQuotaByServiceCodeAndQuotaCode(ctx, conn, serviceCode, quotaCode)

	switch {
	case retry.NotFound(err):
	case err != nil:
		return sdkdiag.AppendErrorf(diags, "reading Service Quotas Service Quota (%s/%s): %s", serviceCode, quotaCode, err)
	default:
		quotaValue = aws.ToFloat64(serviceQuota.Value)
	}

	id := serviceQuotaCreateResourceID(serviceCode, quotaCode)
	value := d.Get(names.AttrValue).(float64)

	if value < quotaValue {
		return sdkdiag.AppendErrorf(diags, "requesting Service Quotas Service Quota (%s) with value less than current", id)
	}

	if value > quotaValue {
		input := servicequotas.RequestServiceQuotaIncreaseInput{
			DesiredValue: aws.Float64(value),
			QuotaCode:    aws.String(quotaCode),
			ServiceCode:  aws.String(serviceCode),
		}

		output, err := conn.RequestServiceQuotaIncrease(ctx, &input)
		if err != nil {
			return sdkdiag.AppendErrorf(diags, "requesting Service Quotas Service Quota (%s) increase: %s", id, err)
		}

		d.Set("request_id", output.RequestedQuota.Id)
	}

	d.SetId(id)

	return append(diags, resourceServiceQuotaRead(ctx, d, meta)...)
}

func resourceServiceQuotaRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ServiceQuotasClient(ctx)

	serviceCode, quotaCode, err := serviceQuotaParseResourceID(d.Id())
	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	// A Service Quota will always have a default value, but will only have a current value if it has been set.
	defaultQuota, err := findDefaultServiceQuotaByServiceCodeAndQuotaCode(ctx, conn, serviceCode, quotaCode)

	if !d.IsNewResource() && retry.NotFound(err) {
		log.Printf("[WARN] Service Quotas default Service Quota (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading Service Quotas default Service Quota (%s/%s): %s", serviceCode, quotaCode, err)
	}

	d.Set("adjustable", defaultQuota.Adjustable)
	d.Set(names.AttrARN, defaultQuota.QuotaArn)
	d.Set(names.AttrDefaultValue, defaultQuota.Value)
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
		tflog.Debug(ctx, "No quota value set", map[string]any{
			"service_code": serviceCode,
			"quota_code":   quotaCode,
		})
	case err != nil:
		return sdkdiag.AppendErrorf(diags, "reading Service Quotas Service Quota (%s/%s): %s", serviceCode, quotaCode, err)
	default:
		d.Set(names.AttrARN, serviceQuota.QuotaArn)
		d.Set(names.AttrValue, serviceQuota.Value)
	}

	if requestID := d.Get("request_id").(string); requestID != "" {
		output, err := findRequestedServiceQuotaChangeByID(ctx, conn, requestID)

		switch {
		case retry.NotFound(err):
			d.Set("request_id", "")
			d.Set("request_status", "")

			return diags
		case err != nil:
			return sdkdiag.AppendErrorf(diags, "reading Service Quotas Requested Service Quota Change (%s): %s", requestID, err)
		default:
			d.Set("request_status", output.Status)
			switch output.Status {
			case awstypes.RequestStatusApproved, awstypes.RequestStatusCaseClosed, awstypes.RequestStatusDenied:
				d.Set("request_id", "")
			case awstypes.RequestStatusCaseOpened, awstypes.RequestStatusPending:
				d.Set(names.AttrValue, output.DesiredValue)
			}
		}
	}

	return diags
}

func resourceServiceQuotaUpdate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ServiceQuotasClient(ctx)

	serviceCode, quotaCode, err := serviceQuotaParseResourceID(d.Id())
	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	input := servicequotas.RequestServiceQuotaIncreaseInput{
		DesiredValue: aws.Float64(d.Get(names.AttrValue).(float64)),
		QuotaCode:    aws.String(quotaCode),
		ServiceCode:  aws.String(serviceCode),
	}

	output, err := conn.RequestServiceQuotaIncrease(ctx, &input)

	if errs.IsAErrorMessageContains[*awstypes.ResourceAlreadyExistsException](err, "Only one open service quota increase request is allowed per quota") {
		return sdkdiag.AppendWarningf(diags, "resource service quota %s already exists", d.Id())
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "requesting Service Quotas Service Quota (%s) increase: %s", d.Id(), err)
	}

	d.Set("request_id", output.RequestedQuota.Id)

	return append(diags, resourceServiceQuotaRead(ctx, d, meta)...)
}

const serviceQuotaResourceIDSeparator = "/"

func serviceQuotaCreateResourceID(serviceCode, quotaCode string) string {
	parts := []string{serviceCode, quotaCode}
	id := strings.Join(parts, serviceQuotaResourceIDSeparator)

	return id
}

func serviceQuotaParseResourceID(id string) (string, string, error) {
	parts := strings.SplitN(id, serviceQuotaResourceIDSeparator, 2)

	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		return "", "", fmt.Errorf("unexpected format for ID (%[1]s), expected SERVICE-CODE%[2]sQUOTA-CODE", id, serviceQuotaResourceIDSeparator)
	}

	return parts[0], parts[1], nil
}

func findDefaultServiceQuotaByServiceCodeAndQuotaCode(ctx context.Context, conn *servicequotas.Client, serviceCode, quotaCode string) (*awstypes.ServiceQuota, error) {
	input := servicequotas.GetAWSDefaultServiceQuotaInput{
		QuotaCode:   aws.String(quotaCode),
		ServiceCode: aws.String(serviceCode),
	}
	output, err := conn.GetAWSDefaultServiceQuota(ctx, &input)

	if errs.IsA[*awstypes.NoSuchResourceException](err) {
		return nil, &retry.NotFoundError{
			LastError: err,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil || output.Quota == nil {
		return nil, tfresource.NewEmptyResultError()
	}

	return output.Quota, nil
}

func findServiceQuotaByServiceCodeAndQuotaCode(ctx context.Context, conn *servicequotas.Client, serviceCode, quotaCode string) (*awstypes.ServiceQuota, error) {
	input := servicequotas.GetServiceQuotaInput{
		QuotaCode:   aws.String(quotaCode),
		ServiceCode: aws.String(serviceCode),
	}

	return findServiceQuota(ctx, conn, &input)
}

func findServiceQuota(ctx context.Context, conn *servicequotas.Client, input *servicequotas.GetServiceQuotaInput) (*awstypes.ServiceQuota, error) {
	output, err := conn.GetServiceQuota(ctx, input)

	if errs.IsA[*awstypes.NoSuchResourceException](err) {
		return nil, &retry.NotFoundError{
			LastError: err,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil || output.Quota == nil {
		return nil, tfresource.NewEmptyResultError()
	}

	if apiObject := output.Quota.ErrorReason; apiObject != nil {
		return nil, fmt.Errorf("%s: %s", apiObject.ErrorCode, aws.ToString(apiObject.ErrorMessage))
	}

	if output.Quota.Value == nil {
		return nil, tfresource.NewEmptyResultError()
	}

	return output.Quota, nil
}

func findRequestedServiceQuotaChangeByID(ctx context.Context, conn *servicequotas.Client, requestID string) (*awstypes.RequestedServiceQuotaChange, error) {
	input := servicequotas.GetRequestedServiceQuotaChangeInput{
		RequestId: aws.String(requestID),
	}

	return findRequestedServiceQuotaChange(ctx, conn, &input)
}

func findRequestedServiceQuotaChange(ctx context.Context, conn *servicequotas.Client, input *servicequotas.GetRequestedServiceQuotaChangeInput) (*awstypes.RequestedServiceQuotaChange, error) {
	output, err := conn.GetRequestedServiceQuotaChange(ctx, input)

	if errs.IsA[*awstypes.NoSuchResourceException](err) {
		return nil, &retry.NotFoundError{
			LastError: err,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil || output.RequestedQuota == nil {
		return nil, tfresource.NewEmptyResultError()
	}

	return output.RequestedQuota, nil
}

func flattenMetricInfo(apiObject *awstypes.MetricInfo) []any {
	if apiObject == nil {
		return []any{}
	}

	var tfList []any
	var tfListMetricDimensions []any

	if apiObject.MetricDimensions != nil && apiObject.MetricDimensions["Service"] != "" {
		tfListMetricDimensions = append(tfListMetricDimensions, map[string]any{
			"class":        apiObject.MetricDimensions["Class"],
			"resource":     apiObject.MetricDimensions["Resource"],
			"service":      apiObject.MetricDimensions["Service"],
			names.AttrType: apiObject.MetricDimensions["Type"],
		})
	} else {
		tfListMetricDimensions = append(tfListMetricDimensions, map[string]any{})
	}

	tfList = append(tfList, map[string]any{
		"metric_dimensions":               tfListMetricDimensions,
		names.AttrMetricName:              apiObject.MetricName,
		"metric_namespace":                apiObject.MetricNamespace,
		"metric_statistic_recommendation": apiObject.MetricStatisticRecommendation,
	})

	return tfList
}
