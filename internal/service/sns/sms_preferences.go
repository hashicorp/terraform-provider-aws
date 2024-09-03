// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package sns

import (
	"context"
	"fmt"
	"strconv"

	"github.com/aws/aws-sdk-go-v2/service/sns"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/attrmap"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
)

func validateMonthlySpend(v interface{}, k string) (ws []string, errors []error) {
	vInt := v.(int)
	if vInt < 0 {
		errors = append(errors, fmt.Errorf("setting SMS preferences: monthly spend limit value [%d] must be >= 0", vInt))
	}
	return
}

func validateDeliverySamplingRate(v interface{}, k string) (ws []string, errors []error) {
	vInt, _ := strconv.Atoi(v.(string))
	if vInt < 0 || vInt > 100 {
		errors = append(errors, fmt.Errorf("setting SMS preferences: default percentage of success to sample value [%d] must be between 0 and 100", vInt))
	}
	return
}

var (
	smsPreferencesSchema = map[string]*schema.Schema{
		"default_sender_id": {
			Type:     schema.TypeString,
			Optional: true,
			AtLeastOneOf: []string{
				"default_sender_id",
				"default_sms_type",
				"delivery_status_iam_role_arn",
				"delivery_status_success_sampling_rate",
				"monthly_spend_limit",
				"usage_report_s3_bucket",
			},
		},
		"default_sms_type": {
			Type:         schema.TypeString,
			Optional:     true,
			ValidateFunc: validation.StringInSlice([]string{"Promotional", "Transactional"}, false),
			AtLeastOneOf: []string{
				"default_sender_id",
				"default_sms_type",
				"delivery_status_iam_role_arn",
				"delivery_status_success_sampling_rate",
				"monthly_spend_limit",
				"usage_report_s3_bucket",
			},
		},
		"delivery_status_iam_role_arn": {
			Type:     schema.TypeString,
			Optional: true,
			AtLeastOneOf: []string{
				"default_sender_id",
				"default_sms_type",
				"delivery_status_iam_role_arn",
				"delivery_status_success_sampling_rate",
				"monthly_spend_limit",
				"usage_report_s3_bucket",
			},
		},
		"delivery_status_success_sampling_rate": {
			Type:         schema.TypeString,
			Optional:     true,
			ValidateFunc: validateDeliverySamplingRate,
			AtLeastOneOf: []string{
				"default_sender_id",
				"default_sms_type",
				"delivery_status_iam_role_arn",
				"delivery_status_success_sampling_rate",
				"monthly_spend_limit",
				"usage_report_s3_bucket",
			},
		},
		"monthly_spend_limit": {
			Type:         schema.TypeInt,
			Optional:     true,
			Computed:     true,
			ValidateFunc: validateMonthlySpend,
			AtLeastOneOf: []string{
				"default_sender_id",
				"default_sms_type",
				"delivery_status_iam_role_arn",
				"delivery_status_success_sampling_rate",
				"monthly_spend_limit",
				"usage_report_s3_bucket",
			},
		},
		"usage_report_s3_bucket": {
			Type:     schema.TypeString,
			Optional: true,
			AtLeastOneOf: []string{
				"default_sender_id",
				"default_sms_type",
				"delivery_status_iam_role_arn",
				"delivery_status_success_sampling_rate",
				"monthly_spend_limit",
				"usage_report_s3_bucket",
			},
		},
	}

	SMSPreferencesAttributeMap = attrmap.New(map[string]string{
		"default_sender_id":                     "DefaultSenderID",
		"default_sms_type":                      "DefaultSMSType",
		"delivery_status_iam_role_arn":          "DeliveryStatusIAMRole",
		"delivery_status_success_sampling_rate": "DeliveryStatusSuccessSamplingRate",
		"monthly_spend_limit":                   "MonthlySpendLimit",
		"usage_report_s3_bucket":                "UsageReportS3Bucket",
	}, smsPreferencesSchema).WithMissingSetToNil("*")
)

// @SDKResource("aws_sns_sms_preferences")
func resourceSMSPreferences() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceSMSPreferencesSet,
		ReadWithoutTimeout:   resourceSMSPreferencesGet,
		UpdateWithoutTimeout: resourceSMSPreferencesSet,
		DeleteWithoutTimeout: resourceSMSPreferencesDelete,

		Schema: smsPreferencesSchema,
	}
}

func resourceSMSPreferencesSet(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).SNSClient(ctx)

	attributes, err := SMSPreferencesAttributeMap.ResourceDataToAPIAttributesCreate(d)
	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	input := &sns.SetSMSAttributesInput{
		Attributes: attributes,
	}

	_, err = conn.SetSMSAttributes(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "setting SNS SMS Preferences: %s", err)
	}

	d.SetId("aws_sns_sms_id")

	return diags
}

func resourceSMSPreferencesGet(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).SNSClient(ctx)

	output, err := conn.GetSMSAttributes(ctx, &sns.GetSMSAttributesInput{})

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading SNS SMS Preferences: %s", err)
	}

	return sdkdiag.AppendFromErr(diags, SMSPreferencesAttributeMap.APIAttributesToResourceData(output.Attributes, d))
}

func resourceSMSPreferencesDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).SNSClient(ctx)

	// Reset the attributes to their default value.
	attributes := make(map[string]string)
	for _, apiAttributeName := range SMSPreferencesAttributeMap.APIAttributeNames() {
		attributes[apiAttributeName] = ""
	}

	input := &sns.SetSMSAttributesInput{
		Attributes: attributes,
	}

	_, err := conn.SetSMSAttributes(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "resetting SNS SMS Preferences: %s", err)
	}

	return diags
}
