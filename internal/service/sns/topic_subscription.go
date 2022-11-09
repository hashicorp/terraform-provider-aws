package sns

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awsutil"
	"github.com/aws/aws-sdk-go/service/sns"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/structure"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/attrmap"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

var (
	subscriptionSchema = map[string]*schema.Schema{
		"arn": {
			Type:     schema.TypeString,
			Computed: true,
		},
		"confirmation_timeout_in_minutes": {
			Type:     schema.TypeInt,
			Optional: true,
			Default:  1,
		},
		"confirmation_was_authenticated": {
			Type:     schema.TypeBool,
			Computed: true,
		},
		"delivery_policy": {
			Type:             schema.TypeString,
			Optional:         true,
			ValidateFunc:     validation.StringIsJSON,
			DiffSuppressFunc: SuppressEquivalentTopicSubscriptionDeliveryPolicy,
		},
		"endpoint": {
			Type:     schema.TypeString,
			Required: true,
			ForceNew: true,
		},
		"endpoint_auto_confirms": {
			Type:     schema.TypeBool,
			Optional: true,
			Default:  false,
		},
		"filter_policy": {
			Type:             schema.TypeString,
			Optional:         true,
			ValidateFunc:     validation.StringIsJSON,
			DiffSuppressFunc: verify.SuppressEquivalentJSONDiffs,
			StateFunc: func(v interface{}) string {
				json, _ := structure.NormalizeJsonString(v)
				return json
			},
		},
		"owner_id": {
			Type:     schema.TypeString,
			Computed: true,
		},
		"pending_confirmation": {
			Type:     schema.TypeBool,
			Computed: true,
		},
		"protocol": {
			Type:         schema.TypeString,
			Required:     true,
			ForceNew:     true,
			ValidateFunc: validation.StringInSlice(SubscriptionProtocol_Values(), false),
		},
		"raw_message_delivery": {
			Type:     schema.TypeBool,
			Optional: true,
			Default:  false,
		},
		"redrive_policy": {
			Type:             schema.TypeString,
			Optional:         true,
			ValidateFunc:     validation.StringIsJSON,
			DiffSuppressFunc: verify.SuppressEquivalentJSONDiffs,
		},
		"subscription_role_arn": {
			Type:         schema.TypeString,
			Optional:     true,
			ValidateFunc: verify.ValidARN,
		},
		"topic_arn": {
			Type:         schema.TypeString,
			Required:     true,
			ForceNew:     true,
			ValidateFunc: verify.ValidARN,
		},
	}

	subscriptionAttributeMap = attrmap.New(map[string]string{
		"arn":                            SubscriptionAttributeNameSubscriptionARN,
		"confirmation_was_authenticated": SubscriptionAttributeNameConfirmationWasAuthenticated,
		"delivery_policy":                SubscriptionAttributeNameDeliveryPolicy,
		"endpoint":                       SubscriptionAttributeNameEndpoint,
		"filter_policy":                  SubscriptionAttributeNameFilterPolicy,
		"owner_id":                       SubscriptionAttributeNameOwner,
		"pending_confirmation":           SubscriptionAttributeNamePendingConfirmation,
		"protocol":                       SubscriptionAttributeNameProtocol,
		"raw_message_delivery":           SubscriptionAttributeNameRawMessageDelivery,
		"redrive_policy":                 SubscriptionAttributeNameRedrivePolicy,
		"subscription_role_arn":          SubscriptionAttributeNameSubscriptionRoleARN,
		"topic_arn":                      SubscriptionAttributeNameTopicARN,
	}, subscriptionSchema).WithMissingSetToNil("*")
)

func ResourceTopicSubscription() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceTopicSubscriptionCreate,
		ReadWithoutTimeout:   resourceTopicSubscriptionRead,
		UpdateWithoutTimeout: resourceTopicSubscriptionUpdate,
		DeleteWithoutTimeout: resourceTopicSubscriptionDelete,

		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: subscriptionSchema,
	}
}

func resourceTopicSubscriptionCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).SNSConn

	attributes, err := subscriptionAttributeMap.ResourceDataToAPIAttributesCreate(d)

	if err != nil {
		return diag.FromErr(err)
	}

	// Endpoint, Protocol and TopicArn are not passed in Attributes.
	delete(attributes, SubscriptionAttributeNameEndpoint)
	delete(attributes, SubscriptionAttributeNameProtocol)
	delete(attributes, SubscriptionAttributeNameTopicARN)

	input := &sns.SubscribeInput{
		Attributes:            aws.StringMap(attributes),
		Endpoint:              aws.String(d.Get("endpoint").(string)),
		Protocol:              aws.String(d.Get("protocol").(string)),
		ReturnSubscriptionArn: aws.Bool(true), // even if not confirmed, will get ARN
		TopicArn:              aws.String(d.Get("topic_arn").(string)),
	}

	output, err := conn.SubscribeWithContext(ctx, input)

	if err != nil {
		return diag.Errorf("creating SNS Topic Subscription: %s", err)
	}

	d.SetId(aws.StringValue(output.SubscriptionArn))

	waitForConfirmation := true

	if !d.Get("endpoint_auto_confirms").(bool) && strings.Contains(d.Get("protocol").(string), "http") {
		waitForConfirmation = false
	}

	if strings.Contains(d.Get("protocol").(string), "email") {
		waitForConfirmation = false
	}

	timeout := subscriptionPendingConfirmationTimeout
	if strings.Contains(d.Get("protocol").(string), "http") {
		timeout = time.Duration(d.Get("confirmation_timeout_in_minutes").(int)) * time.Minute
	}

	if waitForConfirmation {
		if _, err := waitSubscriptionConfirmed(ctx, conn, d.Id(), timeout); err != nil {
			return diag.Errorf("waiting for SNS Topic Subscription (%s) confirmation: %s", d.Id(), err)
		}
	}

	return resourceTopicSubscriptionRead(ctx, d, meta)
}

func resourceTopicSubscriptionRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).SNSConn

	outputRaw, err := tfresource.RetryWhenNewResourceNotFoundContext(ctx, subscriptionCreateTimeout, func() (interface{}, error) {
		return FindSubscriptionAttributesByARN(ctx, conn, d.Id())
	}, d.IsNewResource())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] SNS Topic Subscription %s not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return diag.Errorf("reading SNS Topic Subscription (%s): %s", d.Id(), err)
	}

	attributes := outputRaw.(map[string]string)

	return diag.FromErr(subscriptionAttributeMap.APIAttributesToResourceData(attributes, d))
}

func resourceTopicSubscriptionUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).SNSConn

	attributes, err := subscriptionAttributeMap.ResourceDataToAPIAttributesUpdate(d)

	if err != nil {
		return diag.FromErr(err)
	}

	err = putSubscriptionAttributes(ctx, conn, d.Id(), attributes)

	if err != nil {
		return diag.FromErr(err)
	}

	return resourceTopicSubscriptionRead(ctx, d, meta)
}

func resourceTopicSubscriptionDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).SNSConn

	log.Printf("[DEBUG] Deleting SNS Topic Subscription: %s", d.Id())
	_, err := conn.UnsubscribeWithContext(ctx, &sns.UnsubscribeInput{
		SubscriptionArn: aws.String(d.Id()),
	})

	if tfawserr.ErrMessageContains(err, sns.ErrCodeInvalidParameterException, "Cannot unsubscribe a subscription that is pending confirmation") {
		return nil
	}

	if err != nil {
		return diag.Errorf("deleting SNS Topic Subscription (%s): %s", d.Id(), err)
	}

	if _, err := waitSubscriptionDeleted(ctx, conn, d.Id()); err != nil {
		return diag.Errorf("waiting for SNS Topic Subscription (%s) delete: %s", d.Id(), err)
	}

	return nil
}

func putSubscriptionAttributes(ctx context.Context, conn *sns.SNS, arn string, attributes map[string]string) error {
	for name, value := range attributes {
		err := putSubscriptionAttribute(ctx, conn, arn, name, value)

		if err != nil {
			return err
		}
	}

	return nil
}

func putSubscriptionAttribute(ctx context.Context, conn *sns.SNS, arn string, name, value string) error {
	// https://docs.aws.amazon.com/sns/latest/dg/message-filtering.html#message-filtering-policy-remove
	if name == SubscriptionAttributeNameFilterPolicy && value == "" {
		value = "{}"
	}

	input := &sns.SetSubscriptionAttributesInput{
		AttributeName:   aws.String(name),
		AttributeValue:  aws.String(value),
		SubscriptionArn: aws.String(arn),
	}

	// The AWS API requires a non-empty string value or nil for the RedrivePolicy attribute,
	// else throws an InvalidParameter error.
	if name == SubscriptionAttributeNameRedrivePolicy && value == "" {
		input.AttributeValue = nil
	}

	_, err := conn.SetSubscriptionAttributesWithContext(ctx, input)

	if err != nil {
		return fmt.Errorf("setting SNS Topic Subscription (%s) attribute (%s): %w", arn, name, err)
	}

	return nil
}

type TopicSubscriptionDeliveryPolicy struct {
	Guaranteed         bool                                                 `json:"guaranteed,omitempty"`
	HealthyRetryPolicy *TopicSubscriptionDeliveryPolicyHealthyRetryPolicy   `json:"healthyRetryPolicy,omitempty"`
	SicklyRetryPolicy  *snsTopicSubscriptionDeliveryPolicySicklyRetryPolicy `json:"sicklyRetryPolicy,omitempty"`
	ThrottlePolicy     *snsTopicSubscriptionDeliveryPolicyThrottlePolicy    `json:"throttlePolicy,omitempty"`
}

func (s TopicSubscriptionDeliveryPolicy) String() string {
	return awsutil.Prettify(s)
}

func (s TopicSubscriptionDeliveryPolicy) GoString() string {
	return s.String()
}

type TopicSubscriptionDeliveryPolicyHealthyRetryPolicy struct {
	BackoffFunction    string `json:"backoffFunction,omitempty"`
	MaxDelayTarget     int    `json:"maxDelayTarget,omitempty"`
	MinDelayTarget     int    `json:"minDelayTarget,omitempty"`
	NumMaxDelayRetries int    `json:"numMaxDelayRetries,omitempty"`
	NumMinDelayRetries int    `json:"numMinDelayRetries,omitempty"`
	NumNoDelayRetries  int    `json:"numNoDelayRetries,omitempty"`
	NumRetries         int    `json:"numRetries,omitempty"`
}

func (s TopicSubscriptionDeliveryPolicyHealthyRetryPolicy) String() string {
	return awsutil.Prettify(s)
}

func (s TopicSubscriptionDeliveryPolicyHealthyRetryPolicy) GoString() string {
	return s.String()
}

type snsTopicSubscriptionDeliveryPolicySicklyRetryPolicy struct {
	BackoffFunction    string `json:"backoffFunction,omitempty"`
	MaxDelayTarget     int    `json:"maxDelayTarget,omitempty"`
	MinDelayTarget     int    `json:"minDelayTarget,omitempty"`
	NumMaxDelayRetries int    `json:"numMaxDelayRetries,omitempty"`
	NumMinDelayRetries int    `json:"numMinDelayRetries,omitempty"`
	NumNoDelayRetries  int    `json:"numNoDelayRetries,omitempty"`
	NumRetries         int    `json:"numRetries,omitempty"`
}

func (s snsTopicSubscriptionDeliveryPolicySicklyRetryPolicy) String() string {
	return awsutil.Prettify(s)
}

func (s snsTopicSubscriptionDeliveryPolicySicklyRetryPolicy) GoString() string {
	return s.String()
}

type snsTopicSubscriptionDeliveryPolicyThrottlePolicy struct {
	MaxReceivesPerSecond int `json:"maxReceivesPerSecond,omitempty"`
}

func (s snsTopicSubscriptionDeliveryPolicyThrottlePolicy) String() string {
	return awsutil.Prettify(s)
}

func (s snsTopicSubscriptionDeliveryPolicyThrottlePolicy) GoString() string {
	return s.String()
}

type TopicSubscriptionRedrivePolicy struct {
	DeadLetterTargetArn string `json:"deadLetterTargetArn,omitempty"`
}

func SuppressEquivalentTopicSubscriptionDeliveryPolicy(k, old, new string, d *schema.ResourceData) bool {
	ob, err := normalizeTopicSubscriptionDeliveryPolicy(old)
	if err != nil {
		log.Print(err)
		return false
	}

	nb, err := normalizeTopicSubscriptionDeliveryPolicy(new)
	if err != nil {
		log.Print(err)
		return false
	}

	return verify.JSONBytesEqual(ob, nb)
}

func normalizeTopicSubscriptionDeliveryPolicy(policy string) ([]byte, error) {
	var deliveryPolicy TopicSubscriptionDeliveryPolicy

	if err := json.Unmarshal([]byte(policy), &deliveryPolicy); err != nil {
		return nil, fmt.Errorf("[WARN] Unable to unmarshal SNS Topic Subscription delivery policy JSON: %s", err)
	}

	normalizedDeliveryPolicy, err := json.Marshal(deliveryPolicy)

	if err != nil {
		return nil, fmt.Errorf("[WARN] Unable to marshal SNS Topic Subscription delivery policy back to JSON: %s", err)
	}

	b := bytes.NewBufferString("")
	if err := json.Compact(b, normalizedDeliveryPolicy); err != nil {
		return nil, fmt.Errorf("[WARN] Unable to marshal SNS Topic Subscription delivery policy back to JSON: %s", err)
	}

	return b.Bytes(), nil
}
