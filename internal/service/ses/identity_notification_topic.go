// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package ses

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ses"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_ses_identity_notification_topic")
func ResourceIdentityNotificationTopic() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceNotificationTopicSet,
		ReadWithoutTimeout:   resourceIdentityNotificationTopicRead,
		UpdateWithoutTimeout: resourceNotificationTopicSet,
		DeleteWithoutTimeout: resourceIdentityNotificationTopicDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			names.AttrTopicARN: {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: verify.ValidARN,
			},

			"notification_type": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
				ValidateFunc: validation.StringInSlice([]string{
					ses.NotificationTypeBounce,
					ses.NotificationTypeComplaint,
					ses.NotificationTypeDelivery,
				}, false),
			},

			"identity": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.NoZeroValues,
			},

			"include_original_headers": {
				Type:     schema.TypeBool,
				Optional: true,
			},
		},
	}
}

func resourceNotificationTopicSet(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).SESConn(ctx)
	notification := d.Get("notification_type").(string)
	identity := d.Get("identity").(string)
	includeOriginalHeaders := d.Get("include_original_headers").(bool)

	setOpts := &ses.SetIdentityNotificationTopicInput{
		Identity:         aws.String(identity),
		NotificationType: aws.String(notification),
	}

	if v, ok := d.GetOk(names.AttrTopicARN); ok {
		setOpts.SnsTopic = aws.String(v.(string))
	}

	d.SetId(fmt.Sprintf("%s|%s", identity, notification))

	log.Printf("[DEBUG] Setting SES Identity Notification Topic: %#v", setOpts)

	if _, err := conn.SetIdentityNotificationTopicWithContext(ctx, setOpts); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting SES Identity Notification Topic: %s", err)
	}

	setHeadersOpts := &ses.SetIdentityHeadersInNotificationsEnabledInput{
		Identity:         aws.String(identity),
		NotificationType: aws.String(notification),
		Enabled:          aws.Bool(includeOriginalHeaders),
	}

	log.Printf("[DEBUG] Setting SES Identity Notification Topic Headers: %#v", setHeadersOpts)

	if _, err := conn.SetIdentityHeadersInNotificationsEnabledWithContext(ctx, setHeadersOpts); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting SES Identity Notification Topic Headers Forwarding: %s", err)
	}

	return append(diags, resourceIdentityNotificationTopicRead(ctx, d, meta)...)
}

func resourceIdentityNotificationTopicRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).SESConn(ctx)

	identity, notificationType, err := decodeIdentityNotificationTopicID(d.Id())
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading SES Identity Notification Topic (%s): %s", d.Id(), err)
	}

	d.Set("identity", identity)
	d.Set("notification_type", notificationType)

	getOpts := &ses.GetIdentityNotificationAttributesInput{
		Identities: []*string{aws.String(identity)},
	}

	response, err := conn.GetIdentityNotificationAttributesWithContext(ctx, getOpts)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading SES Identity Notification Topic (%s): %s", d.Id(), err)
	}

	d.Set(names.AttrTopicARN, "")
	if response == nil {
		return diags
	}

	notificationAttributes, notificationAttributesOk := response.NotificationAttributes[identity]
	if !notificationAttributesOk {
		return diags
	}

	switch notificationType {
	case ses.NotificationTypeBounce:
		d.Set(names.AttrTopicARN, notificationAttributes.BounceTopic)
		d.Set("include_original_headers", notificationAttributes.HeadersInBounceNotificationsEnabled)
	case ses.NotificationTypeComplaint:
		d.Set(names.AttrTopicARN, notificationAttributes.ComplaintTopic)
		d.Set("include_original_headers", notificationAttributes.HeadersInComplaintNotificationsEnabled)
	case ses.NotificationTypeDelivery:
		d.Set(names.AttrTopicARN, notificationAttributes.DeliveryTopic)
		d.Set("include_original_headers", notificationAttributes.HeadersInDeliveryNotificationsEnabled)
	}

	return diags
}

func resourceIdentityNotificationTopicDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).SESConn(ctx)

	identity, notificationType, err := decodeIdentityNotificationTopicID(d.Id())
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting SES Identity Notification Topic (%s): %s", d.Id(), err)
	}

	setOpts := &ses.SetIdentityNotificationTopicInput{
		Identity:         aws.String(identity),
		NotificationType: aws.String(notificationType),
		SnsTopic:         nil,
	}

	if _, err := conn.SetIdentityNotificationTopicWithContext(ctx, setOpts); err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting SES Identity Notification Topic (%s): %s", d.Id(), err)
	}

	return append(diags, resourceIdentityNotificationTopicRead(ctx, d, meta)...)
}

func decodeIdentityNotificationTopicID(id string) (string, string, error) {
	parts := strings.Split(id, "|")
	if len(parts) != 2 {
		return "", "", fmt.Errorf("Unexpected format of ID (%q), expected IDENTITY|TYPE", id)
	}
	return parts[0], parts[1], nil
}
