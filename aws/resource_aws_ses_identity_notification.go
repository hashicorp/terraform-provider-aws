package aws

import (
	"fmt"
	"log"
	"strings"

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
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validateArn,
			},

			"notification_type": &schema.Schema{
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validateNotificationType,
			},

			"identity": &schema.Schema{
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validateIdentity,
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

	log.Printf("[DEBUG] Setting SES Identity Notification: %#v", setOpts)

	if _, err := conn.SetIdentityNotificationTopic(setOpts); err != nil {
		return fmt.Errorf("Error setting SES Identity Notification: %s", err)
	}

	return resourceAwsSesNotificationRead(d, meta)
}

func resourceAwsSesNotificationRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).sesConn
	notification := d.Get("notification_type").(string)
	identity := d.Get("identity").(string)

	getOpts := &ses.GetIdentityNotificationAttributesInput{
		Identities: []*string{aws.String(identity)},
	}

	log.Printf("[DEBUG] Reading SES Identity Notification Attributes: %#v", getOpts)

	response, err := conn.GetIdentityNotificationAttributes(getOpts)

	if err != nil {
		return fmt.Errorf("Error reading SES Identity Notification: %s", err)
	}

	notificationAttributes := response.NotificationAttributes[identity]
	switch notification {
	case ses.NotificationTypeBounce:
		if err := d.Set("topic_arn", notificationAttributes.BounceTopic); err != nil {
			return err
		}
	case ses.NotificationTypeComplaint:
		if err := d.Set("topic_arn", notificationAttributes.ComplaintTopic); err != nil {
			return err
		}
	case ses.NotificationTypeDelivery:
		if err := d.Set("topic_arn", notificationAttributes.DeliveryTopic); err != nil {
			return err
		}
	}

	return nil
}

func resourceAwsSesNotificationDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).sesConn
	notification := d.Get("notification_type").(string)
	identity := d.Get("identity").(string)

	setOpts := &ses.SetIdentityNotificationTopicInput{
		Identity:         aws.String(identity),
		NotificationType: aws.String(notification),
		SnsTopic:         nil,
	}

	log.Printf("[DEBUG] Deleting SES Identity Notification: %#v", setOpts)

	if _, err := conn.SetIdentityNotificationTopic(setOpts); err != nil {
		return fmt.Errorf("Error deleting SES Identity Notification: %s", err)
	}

	return resourceAwsSesNotificationRead(d, meta)
}

func validateNotificationType(v interface{}, k string) (ws []string, errors []error) {
	value := strings.Title(strings.ToLower(v.(string)))
	if value == "Bounce" || value == "Complaint" || value == "Delivery" {
		return
	}

	errors = append(errors, fmt.Errorf("%q must be either %q, %q or %q", k, "Bounce", "Complaint", "Delivery"))
	return
}

func validateIdentity(v interface{}, k string) (ws []string, errors []error) {
	value := strings.ToLower(v.(string))
	if value != "" {
		return
	}

	errors = append(errors, fmt.Errorf("%q must not be empty", k))
	return
}
