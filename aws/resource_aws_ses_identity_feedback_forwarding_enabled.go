package aws

import (
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ses"
	"github.com/aws/aws-sdk-go/service/sesv2"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func resourceAwsSesIdentityFeedbackForwardingEnabled() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsSesIdentityFeedbackForwardingEnabledSet,
		Read:   resourceAwsSesIdentityFeedbackForwardingEnabledRead,
		Update: resourceAwsSesIdentityFeedbackForwardingEnabledSet,
		Delete: resourceAwsSesIdentityFeedbackForwardingEnabledDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"identity": {
				Type:     schema.TypeString,
				Required: true,
			},
			"enabled": {
				Type:     schema.TypeBool,
				Required: true,
			},
		},
	}
}

func resourceAwsSesIdentityFeedbackForwardingEnabledSet(d *schema.ResourceData, meta interface{}) error {
	fmt.Println("resourceAwsSesIdentityFeedbackForwardingEnabledSet")
	conn := meta.(*AWSClient).sesconn

	identity := d.Get("identity").(string)
	enabled := d.Get("enabled").(bool)

	input := &ses.SetIdentityFeedbackForwardingEnabledInput{
		Identity:          aws.String(identity),
		ForwardingEnabled: aws.Bool(enabled),
	}
	fmt.Printf("input: %v\n", input)
	_, err := conn.SetIdentityFeedbackForwardingEnabled(input)
	if err != nil {
		return fmt.Errorf("Error setting Feedback Forwarding identity: %s", err)
	}

	d.SetId(identity)

	return resourceAwsSesIdentityFeedbackForwardingEnabledRead(d, meta)
}

func resourceAwsSesIdentityFeedbackForwardingEnabledRead(d *schema.ResourceData, meta interface{}) error {
	fmt.Println("resourceAwsSesIdentityFeedbackForwardingEnabledRead")
	conn := meta.(*AWSClient).sesv2conn

	identity := d.Id()
	d.Set("identity", identity)

	input := &sesv2.GetEmailIdentityInput{
		EmailIdentity: aws.String(identity),
	}

	response, err := conn.GetEmailIdentity(input)
	if err != nil {
		return fmt.Errorf("[WARN] Error fetching email identity for %s: %s", d.Id(), err)
	}

	d.Set("enabled", response.FeedbackForwardingStatus)
	return nil
}

func resourceAwsSesIdentityFeedbackForwardingEnabledDelete(d *schema.ResourceData, meta interface{}) error {
	fmt.Println("resourceAwsSesIdentityFeedbackForwardingEnabledDelete")
	conn := meta.(*AWSClient).sesconn
	identity := d.Id()

	input := &ses.SetIdentityFeedbackForwardingEnabledInput{
		Identity:          aws.String(identity),
		ForwardingEnabled: aws.Bool(true),
	}
	_, err := conn.SetIdentityFeedbackForwardingEnabled(input)
	if err != nil {
		return err
	}
	return nil
}
