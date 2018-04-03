package aws

import (
	"fmt"
	"log"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ses"
	"github.com/hashicorp/terraform/helper/schema"
	"github.com/hashicorp/terraform/helper/validation"
)

func resourceAwsSesNotification() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsSesNotificationSet,
		Read:   resourceAwsSesNotificationRead,
		Update: resourceAwsSesNotificationSet,
		Delete: resourceAwsSesNotificationDelete,

		Schema: map[string]*schema.Schema{
			"topic_arn": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validateArn,
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
		},
	}
}

func resourceAwsSesNotificationSet(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).sesConn
	notification := d.Get("notification_type").(string)
	identity := d.Get("identity").(string)

	setOpts := &ses.SetIdentityNotificationTopicInput{
		Identity:         aws.String(identity),
		NotificationType: aws.String(notification),
	}

	if v, ok := d.GetOk("topic_arn"); ok && v.(string) != "" {
		setOpts.SnsTopic = aws.String(v.(string))
	}

	d.SetId(fmt.Sprintf("%s|%s", identity, notification))

	log.Printf("[DEBUG] Setting SES Identity Notification: %#v", setOpts)

	if _, err := conn.SetIdentityNotificationTopic(setOpts); err != nil {
		return fmt.Errorf("Error setting SES Identity Notification: %s", err)
	}

	return resourceAwsSesNotificationRead(d, meta)
}

func resourceAwsSesNotificationRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).sesConn

	identity, notificationType, err := decodeSesIdentityNotificationId(d.Id())
	if err != nil {
		return err
	}

	getOpts := &ses.GetIdentityNotificationAttributesInput{
		Identities: []*string{aws.String(identity)},
	}

	log.Printf("[DEBUG] Reading SES Identity Notification Attributes: %#v", getOpts)

	response, err := conn.GetIdentityNotificationAttributes(getOpts)

	if err != nil {
		return fmt.Errorf("Error reading SES Identity Notification: %s", err)
	}

	notificationAttributes := response.NotificationAttributes[identity]
	switch notificationType {
	case ses.NotificationTypeBounce:
		d.Set("topic_arn", notificationAttributes.BounceTopic)
	case ses.NotificationTypeComplaint:
		d.Set("topic_arn", notificationAttributes.ComplaintTopic)
	case ses.NotificationTypeDelivery:
		d.Set("topic_arn", notificationAttributes.DeliveryTopic)
	}

	return nil
}

func resourceAwsSesNotificationDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).sesConn

	identity, notificationType, err := decodeSesIdentityNotificationId(d.Id())
	if err != nil {
		return err
	}

	setOpts := &ses.SetIdentityNotificationTopicInput{
		Identity:         aws.String(identity),
		NotificationType: aws.String(notificationType),
		SnsTopic:         nil,
	}

	log.Printf("[DEBUG] Deleting SES Identity Notification: %#v", setOpts)

	if _, err := conn.SetIdentityNotificationTopic(setOpts); err != nil {
		return fmt.Errorf("Error deleting SES Identity Notification: %s", err)
	}

	return resourceAwsSesNotificationRead(d, meta)
}

func decodeSesIdentityNotificationId(id string) (string, string, error) {
	parts := strings.Split(id, "|")
	if len(parts) != 2 {
		return "", "", fmt.Errorf("Unexpected format of ID (%q), expected IDENTITY|TYPE", id)
	}
	return parts[0], parts[1], nil
}
