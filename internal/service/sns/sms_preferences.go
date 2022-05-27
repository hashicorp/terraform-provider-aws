package sns

import (
	"fmt"
	"log"
	"strconv"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/sns"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/attrmap"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
)

func validateMonthlySpend(v interface{}, k string) (ws []string, errors []error) {
	vInt := v.(int)
	if vInt < 0 {
		errors = append(errors, fmt.Errorf("error setting SMS preferences: monthly spend limit value [%d] must be >= 0", vInt))
	}
	return
}

func validateDeliverySamplingRate(v interface{}, k string) (ws []string, errors []error) {
	vInt, _ := strconv.Atoi(v.(string))
	if vInt < 0 || vInt > 100 {
		errors = append(errors, fmt.Errorf("error setting SMS preferences: default percentage of success to sample value [%d] must be between 0 and 100", vInt))
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
	}, smsPreferencesSchema)
)

func ResourceSMSPreferences() *schema.Resource {
	return &schema.Resource{
		Create: resourceSMSPreferencesSet,
		Read:   resourceSMSPreferencesGet,
		Update: resourceSMSPreferencesSet,
		Delete: resourceSMSPreferencesDelete,

		Schema: smsPreferencesSchema,
	}
}

func resourceSMSPreferencesSet(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).SNSConn

	attributes, err := SMSPreferencesAttributeMap.ResourceDataToAPIAttributesCreate(d)

	if err != nil {
		return err
	}

	input := &sns.SetSMSAttributesInput{
		Attributes: aws.StringMap(attributes),
	}

	log.Printf("[DEBUG] Setting SNS SMS Attributes: %s", input)
	if _, err := conn.SetSMSAttributes(input); err != nil {
		return fmt.Errorf("error setting SNS SMS Preferences: %w", err)
	}

	d.SetId("aws_sns_sms_id")

	return nil
}

func resourceSMSPreferencesGet(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).SNSConn

	output, err := conn.GetSMSAttributes(&sns.GetSMSAttributesInput{})

	if err != nil {
		return fmt.Errorf("error getting SNS SMS Preferences: %w", err)
	}

	err = SMSPreferencesAttributeMap.APIAttributesToResourceData(aws.StringValueMap(output.Attributes), d)

	if err != nil {
		return err
	}

	return nil
}

func resourceSMSPreferencesDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).SNSConn

	// Reset the attributes to their default value.
	attributes := make(map[string]string)
	for _, apiAttributeName := range SMSPreferencesAttributeMap.APIAttributeNames() {
		attributes[apiAttributeName] = ""
	}

	input := &sns.SetSMSAttributesInput{
		Attributes: aws.StringMap(attributes),
	}

	if _, err := conn.SetSMSAttributes(input); err != nil {
		return fmt.Errorf("error resetting SNS SMS Preferences: %w", err)
	}

	return nil
}
