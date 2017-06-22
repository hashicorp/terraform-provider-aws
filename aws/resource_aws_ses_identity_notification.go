package aws

import (
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ses"
	"github.com/hashicorp/terraform/helper/schema"
)

func resourceAwsSesNotification() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsSesNotificationSet,
		Read:   resourceAwsSesNotificationRead,
		Update: resourceAwsSesNotificationSet,
		Delete: resourceAwsSesNotificationDelete,

		Schema: map[string]*schema.Schema{
			"topic_arn": &schema.Schema{
				Type:     schema.TypeString,
				Required: false,
			},

			"notification_type": &schema.Schema{
				Type:     schema.TypeSet,
				Required: true,
			},

			"identity": &schema.Schema{
				Type:     schema.TypeSet,
				Required: true,
			},
		},
	}
}

func resourceAwsSesNotificationSet(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).sesConn
	topic := d.Get("topic_arn").(string)
	notification := d.Get("notification_type").(string)
	identity := d.Get("identity").(string)

	setOpts := &ses.SetIdentityNotificationTopicInput{
		Identity:         aws.String(identity),
		NotificationType: aws.String(notification),
		SnsTopic:         aws.String(topic),
	}

	_, err := conn.SetIdentityNotificationTopicRequest(setOpts).Send()

	if err != nil {
		return fmt.Errorf("Error setting SES Identity Notification: %s", err)
	}

	return resourceAwsSesNotificationRead(d, meta)
}

func resourceAwsSesNotificationRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).sesConn
	notification := d.Get("notification_type").(*schema.Set)
	identity := d.Get("identity").(*schema.Set)

	getOpts := &ses.GetIdentityNotificationAttributesInput{
		Identities: []*string{aws.String(identity)},
	}

	response, err := conn.GetIdentityNotificationAttributes(getOpts)

	if err != nil {
		return fmt.Errorf("Error reading SES Identity Notification: %s", err)
	}

	notificationAttributes := response.NotificationAttributes[identity]
	switch notification {
	case "Bounce":
		if err := d.Set("topic", notificationAttributes.BounceTopic); err != nil {
			return err
		}
	case "Complain":
		if err := d.Set("topic", notificationAttributes.ComplaintTopic); err != nil {
			return err
		}
	case "Delivery":
		if err := d.Set("topic", notificationAttributes.DeliveryTopic); err != nil {
			return err
		}
	}

	return nil
}

func resourceAwsSesNotificationDelete(d *schema.ResourceData, meta interface{}) error {
	d.Set("topic_arn", nil)
	return resourceAwsSesNotificationSet(d, meta)
}
