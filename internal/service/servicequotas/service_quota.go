// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

// DONOTCOPY: Copying old resources spreads bad habits. Use skaff instead.

package servicequotas

import (
	"context"
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"

	"github.com/YakDriver/regexache"
	"github.com/YakDriver/smarterr"
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

type serviceQuotaDefaultReader interface {
	GetAWSDefaultServiceQuota(context.Context, *servicequotas.GetAWSDefaultServiceQuotaInput, ...func(*servicequotas.Options)) (*servicequotas.GetAWSDefaultServiceQuotaOutput, error)
}

type serviceQuotaReader interface {
	GetServiceQuota(context.Context, *servicequotas.GetServiceQuotaInput, ...func(*servicequotas.Options)) (*servicequotas.GetServiceQuotaOutput, error)
}

type serviceQuotaRequestReader interface {
	GetRequestedServiceQuotaChange(context.Context, *servicequotas.GetRequestedServiceQuotaChangeInput, ...func(*servicequotas.Options)) (*servicequotas.GetRequestedServiceQuotaChangeOutput, error)
}

type serviceQuotaRequestHistoryReader interface {
	ListRequestedServiceQuotaChangeHistoryByQuota(context.Context, *servicequotas.ListRequestedServiceQuotaChangeHistoryByQuotaInput, ...func(*servicequotas.Options)) (*servicequotas.ListRequestedServiceQuotaChangeHistoryByQuotaOutput, error)
}

type serviceQuotaWaitClient interface {
	serviceQuotaReader
	serviceQuotaRequestReader
}

type serviceQuotaIncreaseClient interface {
	serviceQuotaWaitClient
	serviceQuotaRequestHistoryReader
	RequestServiceQuotaIncrease(context.Context, *servicequotas.RequestServiceQuotaIncreaseInput, ...func(*servicequotas.Options)) (*servicequotas.RequestServiceQuotaIncreaseOutput, error)
}

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

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(10 * time.Minute),
			Update: schema.DefaultTimeout(10 * time.Minute),
		},

		SchemaFunc: func() map[string]*schema.Schema {
			return map[string]*schema.Schema{
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
					DiffSuppressFunc: func(_, old, new string, d *schema.ResourceData) bool {
						if !d.Get("wait_for_fulfillment").(bool) {
							return false
						}
						applied, err := strconv.ParseFloat(old, 64)
						if err != nil {
							return false
						}
						requested, err := strconv.ParseFloat(new, 64)
						return err == nil && applied >= requested
					},
				},
				"wait_for_fulfillment": {
					Type:     schema.TypeBool,
					Optional: true,
					Default:  false,
				},
			}
		},
	}
}

func resourceServiceQuotaCreate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ServiceQuotasClient(ctx)
	if d.Get("wait_for_fulfillment").(bool) {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, d.Timeout(schema.TimeoutCreate))
		defer cancel()
	}

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

	if value < quotaValue && !d.Get("wait_for_fulfillment").(bool) {
		return sdkdiag.AppendErrorf(diags, "requesting Service Quotas Service Quota (%s) with value less than current", id)
	}

	if value > quotaValue {
		if err := increaseServiceQuota(ctx, conn, d, serviceCode, quotaCode, value, d.Timeout(schema.TimeoutCreate)); err != nil {
			return sdkdiag.AppendErrorf(diags, "requesting Service Quotas Service Quota (%s) increase: %s", id, err)
		}
	} else if d.Get("wait_for_fulfillment").(bool) {
		d.Set("request_id", "")
		d.Set("request_status", "")
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
			refreshServiceQuotaRequest(d, output)
		}
	}

	return diags
}

func resourceServiceQuotaUpdate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ServiceQuotasClient(ctx)
	if d.Get("wait_for_fulfillment").(bool) {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, d.Timeout(schema.TimeoutUpdate))
		defer cancel()
	}

	serviceCode, quotaCode, err := serviceQuotaParseResourceID(d.Id())
	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

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

	requestedValue := d.Get(names.AttrValue).(float64)

	if requestedValue < quotaValue && !d.Get("wait_for_fulfillment").(bool) {
		return sdkdiag.AppendErrorf(diags, "updating Service Quotas Service Quota (%s) with value (%f) less than current (%f)", d.Id(), requestedValue, quotaValue)
	}

	if requestedValue <= quotaValue {
		tflog.Info(ctx, "Service Quota value already satisfied, skipping increase request", map[string]any{
			"service_code":    serviceCode,
			"quota_code":      quotaCode,
			"current_value":   quotaValue,
			"requested_value": requestedValue,
		})
		if d.Get("wait_for_fulfillment").(bool) {
			d.Set("request_id", "")
			d.Set("request_status", "")
		}

		return append(diags, resourceServiceQuotaRead(ctx, d, meta)...)
	}

	if err := increaseServiceQuota(ctx, conn, d, serviceCode, quotaCode, requestedValue, d.Timeout(schema.TimeoutUpdate)); err != nil {
		if !d.Get("wait_for_fulfillment").(bool) && errs.IsAErrorMessageContains[*awstypes.ResourceAlreadyExistsException](err, "Only one open service quota increase request is allowed per quota") {
			return sdkdiag.AppendWarningf(diags, "resource service quota %s already exists", d.Id())
		}
		return sdkdiag.AppendErrorf(diags, "requesting Service Quotas Service Quota (%s) increase: %s", d.Id(), err)
	}

	return append(diags, resourceServiceQuotaRead(ctx, d, meta)...)
}

func refreshServiceQuotaRequest(d *schema.ResourceData, request *awstypes.RequestedServiceQuotaChange) {
	d.Set("request_status", request.Status)
	if d.Get("wait_for_fulfillment").(bool) {
		switch request.Status {
		case awstypes.RequestStatusDenied, awstypes.RequestStatusNotApproved, awstypes.RequestStatusInvalidRequest:
			d.Set("request_id", "")
		default:
			// Refresh has the applied value, not the configured target. Apply can finish
			// at a lower configured target, but refresh must not discard a pending wait.
			if request.DesiredValue != nil && d.Get(names.AttrValue).(float64) >= aws.ToFloat64(request.DesiredValue) {
				d.Set("request_id", "")
			}
		}
		return
	}

	switch request.Status {
	case awstypes.RequestStatusApproved, awstypes.RequestStatusCaseClosed, awstypes.RequestStatusDenied:
		d.Set("request_id", "")
	case awstypes.RequestStatusCaseOpened, awstypes.RequestStatusPending:
		d.Set(names.AttrValue, request.DesiredValue)
	}
}

func increaseServiceQuota(ctx context.Context, conn serviceQuotaIncreaseClient, d *schema.ResourceData, serviceCode, quotaCode string, value float64, timeout time.Duration) error {
	wait := d.Get("wait_for_fulfillment").(bool)
	var request *awstypes.RequestedServiceQuotaChange
	if wait {
		if requestID := d.Get("request_id").(string); requestID != "" {
			var err error
			request, err = findRequestedServiceQuotaChangeByID(ctx, conn, requestID)
			if err != nil && !retry.NotFound(err) {
				return smarterr.NewError(err)
			}
			if request != nil {
				switch request.Status {
				case awstypes.RequestStatusDenied, awstypes.RequestStatusNotApproved, awstypes.RequestStatusInvalidRequest:
					request = nil
				case awstypes.RequestStatusApproved, awstypes.RequestStatusCaseClosed:
					if request.DesiredValue != nil && aws.ToFloat64(request.DesiredValue) < value {
						request = nil
					}
				}
			}
			if request == nil {
				d.Set("request_id", "")
				d.Set("request_status", "")
			}
		}
		if request == nil {
			var err error
			request, err = findServiceQuotaRequestByQuota(ctx, conn, serviceCode, quotaCode, value)
			if err != nil && !retry.NotFound(err) {
				return smarterr.NewError(err)
			}
		}
	}

	if request == nil {
		input := servicequotas.RequestServiceQuotaIncreaseInput{
			DesiredValue: aws.Float64(value),
			QuotaCode:    aws.String(quotaCode),
			ServiceCode:  aws.String(serviceCode),
		}
		output, err := conn.RequestServiceQuotaIncrease(ctx, &input)
		if wait && errs.IsAErrorMessageContains[*awstypes.ResourceAlreadyExistsException](err, "Only one open service quota increase request is allowed per quota") {
			request, err = findServiceQuotaRequestByQuota(ctx, conn, serviceCode, quotaCode, value)
		} else if err == nil && output != nil {
			request = output.RequestedQuota
		}
		if err != nil {
			return smarterr.NewError(err)
		}
	}
	if request == nil || aws.ToString(request.Id) == "" {
		return smarterr.Errorf("response did not include a request ID")
	}
	if wait {
		if code := aws.ToString(request.ServiceCode); code != "" && code != serviceCode {
			return smarterr.Errorf("request (%s) belongs to another service", aws.ToString(request.Id))
		}
		if code := aws.ToString(request.QuotaCode); code != "" && code != quotaCode {
			return smarterr.Errorf("request (%s) belongs to another quota", aws.ToString(request.Id))
		}
	}

	d.SetId(serviceQuotaCreateResourceID(serviceCode, quotaCode))
	d.Set("request_id", request.Id)
	d.Set("request_status", request.Status)
	if !wait {
		return nil
	}
	if request.DesiredValue == nil || aws.ToFloat64(request.DesiredValue) < value {
		return smarterr.Errorf("existing request (%s) does not request the configured quota value (%f)", aws.ToString(request.Id), value)
	}
	if err := waitServiceQuotaFulfilled(ctx, conn, serviceCode, quotaCode, aws.ToString(request.Id), value, timeout); err != nil {
		return smarterr.Errorf("waiting for request (%s) to be fulfilled: %w", aws.ToString(request.Id), err)
	}
	d.Set("request_id", "")
	d.Set("request_status", "")
	return nil
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

func findDefaultServiceQuotaByServiceCodeAndQuotaCode(ctx context.Context, conn serviceQuotaDefaultReader, serviceCode, quotaCode string) (*awstypes.ServiceQuota, error) {
	input := servicequotas.GetAWSDefaultServiceQuotaInput{
		QuotaCode:   aws.String(quotaCode),
		ServiceCode: aws.String(serviceCode),
	}
	output, err := conn.GetAWSDefaultServiceQuota(ctx, &input)

	if errs.IsA[*awstypes.NoSuchResourceException](err) {
		return nil, &retry.NotFoundError{LastError: err}
	}

	if err != nil {
		return nil, err
	}

	if output == nil || output.Quota == nil {
		return nil, tfresource.NewEmptyResultError()
	}

	return output.Quota, nil
}

func findServiceQuotaByServiceCodeAndQuotaCode(ctx context.Context, conn serviceQuotaReader, serviceCode, quotaCode string) (*awstypes.ServiceQuota, error) {
	input := servicequotas.GetServiceQuotaInput{
		QuotaCode:   aws.String(quotaCode),
		ServiceCode: aws.String(serviceCode),
	}

	return findServiceQuota(ctx, conn, &input)
}

func findServiceQuota(ctx context.Context, conn serviceQuotaReader, input *servicequotas.GetServiceQuotaInput) (*awstypes.ServiceQuota, error) {
	output, err := conn.GetServiceQuota(ctx, input)

	if errs.IsA[*awstypes.NoSuchResourceException](err) {
		return nil, &retry.NotFoundError{LastError: err}
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

func findRequestedServiceQuotaChangeByID(ctx context.Context, conn serviceQuotaRequestReader, requestID string) (*awstypes.RequestedServiceQuotaChange, error) {
	input := servicequotas.GetRequestedServiceQuotaChangeInput{
		RequestId: aws.String(requestID),
	}

	return findRequestedServiceQuotaChange(ctx, conn, &input)
}

func findRequestedServiceQuotaChange(ctx context.Context, conn serviceQuotaRequestReader, input *servicequotas.GetRequestedServiceQuotaChangeInput) (*awstypes.RequestedServiceQuotaChange, error) {
	output, err := conn.GetRequestedServiceQuotaChange(ctx, input)

	if errs.IsA[*awstypes.NoSuchResourceException](err) {
		return nil, &retry.NotFoundError{LastError: err}
	}

	if err != nil {
		return nil, err
	}

	if output == nil || output.RequestedQuota == nil {
		return nil, tfresource.NewEmptyResultError()
	}

	return output.RequestedQuota, nil
}

func findServiceQuotaRequestByQuota(ctx context.Context, conn serviceQuotaRequestHistoryReader, serviceCode, quotaCode string, value float64) (*awstypes.RequestedServiceQuotaChange, error) {
	input := servicequotas.ListRequestedServiceQuotaChangeHistoryByQuotaInput{
		QuotaCode:   aws.String(quotaCode),
		ServiceCode: aws.String(serviceCode),
	}
	var latest *awstypes.RequestedServiceQuotaChange
	for {
		if err := ctx.Err(); err != nil {
			return nil, smarterr.NewError(err)
		}
		output, err := conn.ListRequestedServiceQuotaChangeHistoryByQuota(ctx, &input)

		if errs.IsA[*awstypes.NoSuchResourceException](err) {
			return nil, smarterr.NewError(&retry.NotFoundError{LastError: err})
		}

		if err != nil {
			return nil, smarterr.NewError(err)
		}

		if output == nil {
			return nil, smarterr.NewError(tfresource.NewEmptyResultError())
		}

		for i := range output.RequestedQuotas {
			request := &output.RequestedQuotas[i]

			switch request.Status {
			case awstypes.RequestStatusCaseOpened, awstypes.RequestStatusPending:
				return request, nil
			case awstypes.RequestStatusApproved, awstypes.RequestStatusCaseClosed:
				if aws.ToFloat64(request.DesiredValue) >= value && (latest == nil || aws.ToTime(request.Created).After(aws.ToTime(latest.Created))) {
					latest = request
				}
			}
		}

		if output.NextToken == nil || aws.ToString(output.NextToken) == "" {
			break
		}

		if aws.ToString(output.NextToken) == aws.ToString(input.NextToken) {
			return nil, smarterr.Errorf("quota request history returned a repeated pagination token")
		}
		input.NextToken = output.NextToken
	}
	if latest != nil {
		return latest, nil
	}

	return nil, smarterr.NewError(&retry.NotFoundError{})
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

func statusServiceQuotaFulfillment(conn serviceQuotaWaitClient, serviceCode, quotaCode, requestID string, value float64) retry.StateRefreshFunc {
	return func(ctx context.Context) (any, string, error) {
		quota, state, err := statusServiceQuotaValue(conn, serviceCode, quotaCode, value)(ctx)
		if err != nil || state == "updated" {
			return quota, state, smarterr.NewError(err)
		}
		output, err := findRequestedServiceQuotaChangeByID(ctx, conn, requestID)

		if retry.NotFound(err) {
			return quota, "pending", nil
		}

		if err != nil {
			return nil, "", smarterr.NewError(err)
		}

		switch output.Status {
		case awstypes.RequestStatusDenied:
			return output, "", smarterr.Errorf("service quota increase request was denied")
		case awstypes.RequestStatusNotApproved:
			return output, "", smarterr.Errorf("service quota increase request was not approved")
		case awstypes.RequestStatusInvalidRequest:
			return output, "", smarterr.Errorf("service quota increase request was invalid")
		case awstypes.RequestStatusPending, awstypes.RequestStatusCaseOpened, awstypes.RequestStatusApproved, awstypes.RequestStatusCaseClosed:
		default:
			return output, "", smarterr.Errorf("unexpected quota request status (%s)", output.Status)
		}

		return output, "pending", nil
	}
}

func waitServiceQuotaFulfilled(ctx context.Context, conn serviceQuotaWaitClient, serviceCode, quotaCode, requestID string, requestedValue float64, timeout time.Duration) error {
	waitCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	stateConf := &retry.StateChangeConf{
		Pending:    []string{"pending"},
		Target:     []string{"updated"},
		Refresh:    statusServiceQuotaFulfillment(conn, serviceCode, quotaCode, requestID, requestedValue),
		Timeout:    timeout,
		MinTimeout: 10 * time.Second,
	}

	_, err := stateConf.WaitForStateContext(waitCtx)
	return smarterr.NewError(err)
}

func statusServiceQuotaValue(conn serviceQuotaReader, serviceCode, quotaCode string, targetValue float64) retry.StateRefreshFunc {
	return func(ctx context.Context) (any, string, error) {
		serviceQuota, err := findServiceQuotaByServiceCodeAndQuotaCode(ctx, conn, serviceCode, quotaCode)

		if retry.NotFound(err) {
			tflog.Debug(ctx, "Quota value not yet available", map[string]any{
				"service_code": serviceCode,
				"quota_code":   quotaCode,
			})
			return nil, "pending", nil
		}

		if err != nil {
			return nil, "", err
		}

		currentValue := aws.ToFloat64(serviceQuota.Value)
		tflog.Debug(ctx, "Checking quota value", map[string]any{
			"service_code":  serviceCode,
			"quota_code":    quotaCode,
			"current_value": currentValue,
			"target_value":  targetValue,
		})

		if currentValue >= targetValue {
			return serviceQuota, "updated", nil
		}

		return serviceQuota, "pending", nil
	}
}
