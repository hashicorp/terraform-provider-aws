// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package ses

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ses"
	awstypes "github.com/aws/aws-sdk-go-v2/service/ses/types"
	"github.com/hashicorp/aws-sdk-go-base/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_ses_identity_notification_topic", name="Identity Notification Topic")
func resourceIdentityNotificationTopic() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceIdentityNotificationTopicSet,
		ReadWithoutTimeout:   resourceIdentityNotificationTopicRead,
		UpdateWithoutTimeout: resourceIdentityNotificationTopicSet,
		DeleteWithoutTimeout: resourceIdentityNotificationTopicDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
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
			"notification_type": {
				Type:             schema.TypeString,
				Required:         true,
				ForceNew:         true,
				ValidateDiagFunc: enum.Validate[awstypes.NotificationType](),
			},
			names.AttrTopicARN: {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: verify.ValidARN,
			},
		},
	}
}

func resourceIdentityNotificationTopicSet(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).SESClient(ctx)

	identity := d.Get("identity").(string)
	notificationType := awstypes.NotificationType(d.Get("notification_type").(string))
	id := identityNotificationTopicCreateResourceID(identity, notificationType)
	inputSINT := &ses.SetIdentityNotificationTopicInput{
		Identity:         aws.String(identity),
		NotificationType: notificationType,
	}

	if v, ok := d.GetOk(names.AttrTopicARN); ok {
		inputSINT.SnsTopic = aws.String(v.(string))
	}

	_, err := conn.SetIdentityNotificationTopic(ctx, inputSINT)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "setting SES Identity Notification Topic (%s): %s", id, err)
	}

	if d.IsNewResource() {
		d.SetId(id)
	}

	inputSIHINE := &ses.SetIdentityHeadersInNotificationsEnabledInput{
		Enabled:          d.Get("include_original_headers").(bool),
		Identity:         aws.String(identity),
		NotificationType: notificationType,
	}

	_, err = conn.SetIdentityHeadersInNotificationsEnabled(ctx, inputSIHINE)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "setting SES Identity Notification Topic (%s) headers in notification: %s", d.Id(), err)
	}

	return append(diags, resourceIdentityNotificationTopicRead(ctx, d, meta)...)
}

func resourceIdentityNotificationTopicRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).SESClient(ctx)

	identity, notificationType, err := identityNotificationTopicParseResourceID(d.Id())
	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	notificationAttributes, err := findIdentityNotificationAttributesByIdentity(ctx, conn, identity)

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] SES Identity Notification Topic (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading SES Identity Notification Topic (%s): %s", d.Id(), err)
	}

	d.Set("identity", identity)
	d.Set("notification_type", notificationType)

	switch notificationType {
	case awstypes.NotificationTypeBounce:
		d.Set(names.AttrTopicARN, notificationAttributes.BounceTopic)
		d.Set("include_original_headers", notificationAttributes.HeadersInBounceNotificationsEnabled)
	case awstypes.NotificationTypeComplaint:
		d.Set(names.AttrTopicARN, notificationAttributes.ComplaintTopic)
		d.Set("include_original_headers", notificationAttributes.HeadersInComplaintNotificationsEnabled)
	case awstypes.NotificationTypeDelivery:
		d.Set(names.AttrTopicARN, notificationAttributes.DeliveryTopic)
		d.Set("include_original_headers", notificationAttributes.HeadersInDeliveryNotificationsEnabled)
	}

	return diags
}

func resourceIdentityNotificationTopicDelete(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).SESClient(ctx)

	identity, notificationType, err := identityNotificationTopicParseResourceID(d.Id())
	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	log.Printf("[DEBUG] Deleting SES Identity Notification Topic: %s", d.Id())
	_, err = conn.SetIdentityNotificationTopic(ctx, &ses.SetIdentityNotificationTopicInput{
		Identity:         aws.String(identity),
		NotificationType: notificationType,
	})

	if tfawserr.ErrMessageContains(err, errCodeInvalidParameterValue, "Must be a verified email address or domain") {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting SES Identity Notification Topic (%s): %s", d.Id(), err)
	}

	return diags
}

const identityNotificationTopicResourceIDSeparator = "|"

func identityNotificationTopicCreateResourceID(identity string, notificationType awstypes.NotificationType) string {
	parts := []string{identity, string(notificationType)} // nosemgrep:ci.typed-enum-conversion
	id := strings.Join(parts, identityNotificationTopicResourceIDSeparator)

	return id
}

func identityNotificationTopicParseResourceID(id string) (string, awstypes.NotificationType, error) {
	parts := strings.SplitN(id, identityNotificationTopicResourceIDSeparator, 2)

	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		return "", "", fmt.Errorf("unexpected format of ID (%[1]s), expected IDENTITY%[2]sNOTIFICATION_TYPE", id, identityNotificationTopicResourceIDSeparator)
	}

	return parts[0], awstypes.NotificationType(parts[1]), nil
}

func findIdentityNotificationAttributesByIdentity(ctx context.Context, conn *ses.Client, identity string) (*awstypes.IdentityNotificationAttributes, error) {
	input := &ses.GetIdentityNotificationAttributesInput{
		Identities: []string{identity},
	}
	output, err := findIdentityNotificationAttributes(ctx, conn, input)

	if err != nil {
		return nil, err
	}

	if v, ok := output[identity]; ok {
		return &v, nil
	}

	return nil, &retry.NotFoundError{}
}

func findIdentityNotificationAttributes(ctx context.Context, conn *ses.Client, input *ses.GetIdentityNotificationAttributesInput) (map[string]awstypes.IdentityNotificationAttributes, error) {
	output, err := conn.GetIdentityNotificationAttributes(ctx, input)

	if err != nil {
		return nil, err
	}

	if output == nil || output.NotificationAttributes == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	return output.NotificationAttributes, nil
}
