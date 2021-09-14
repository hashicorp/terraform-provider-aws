package aws

import (
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ses"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func resourceAwsSesEmailMailFrom() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsSesEmailMailFromSet,
		Read:   resourceAwsSesEmailMailFromRead,
		Update: resourceAwsSesEmailMailFromSet,
		Delete: resourceAwsSesEmailMailFromDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"email": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"mail_from_domain": {
				Type:     schema.TypeString,
				Required: true,
			},
			"behavior_on_mx_failure": {
				Type:     schema.TypeString,
				Optional: true,
				Default:  ses.BehaviorOnMXFailureUseDefaultValue,
			},
		},
	}
}

func resourceAwsSesEmailMailFromSet(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).sesconn

	behaviorOnMxFailure := d.Get("behavior_on_mx_failure").(string)
	emailName := d.Get("email").(string)
	mailFromDomain := d.Get("mail_from_domain").(string)

	input := &ses.SetIdentityMailFromDomainInput{
		BehaviorOnMXFailure: aws.String(behaviorOnMxFailure),
		Identity:            aws.String(emailName),
		MailFromDomain:      aws.String(mailFromDomain),
	}

	_, err := conn.SetIdentityMailFromDomain(input)
	if err != nil {
		return fmt.Errorf("Error setting MAIL FROM domain: %s", err)
	}

	d.SetId(emailName)

	return resourceAwsSesEmailMailFromRead(d, meta)
}

func resourceAwsSesEmailMailFromRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).sesconn

	emailName := d.Id()

	readOpts := &ses.GetIdentityMailFromDomainAttributesInput{
		Identities: []*string{
			aws.String(emailName),
		},
	}

	out, err := conn.GetIdentityMailFromDomainAttributes(readOpts)

	if err != nil {
		return fmt.Errorf("error fetching SES MAIL FROM domain attributes for %s: %s", emailName, err)
	}

	if out == nil {
		return fmt.Errorf("error fetching SES MAIL FROM domain attributes for %s: empty response", emailName)
	}

	attributes, ok := out.MailFromDomainAttributes[emailName]

	if !ok {
		log.Printf("[WARN] SES Domain Identity (%s) not found, removing from state", emailName)
		d.SetId("")
		return nil
	}

	d.Set("behavior_on_mx_failure", attributes.BehaviorOnMXFailure)
	d.Set("email", emailName)
	d.Set("mail_from_domain", attributes.MailFromDomain)

	return nil
}

func resourceAwsSesEmailMailFromDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).sesconn

	emailName := d.Id()

	deleteOpts := &ses.SetIdentityMailFromDomainInput{
		Identity:       aws.String(emailName),
		MailFromDomain: nil,
	}

	_, err := conn.SetIdentityMailFromDomain(deleteOpts)
	if err != nil {
		return fmt.Errorf("Error deleting SES email identity: %s", err)
	}

	return nil
}
