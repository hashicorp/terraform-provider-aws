package aws

import (
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ses"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func resourceAwsSesIdentityMailFrom() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsSesIdentityMailFromSet,
		Read:   resourceAwsSesIdentityMailFromRead,
		Update: resourceAwsSesIdentityMailFromSet,
		Delete: resourceAwsSesIdentityMailFromDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"identity": {
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

func resourceAwsSesIdentityMailFromSet(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).sesconn

	behaviorOnMxFailure := d.Get("behavior_on_mx_failure").(string)
	identityName := d.Get("identity").(string)
	mailFromDomain := d.Get("mail_from_domain").(string)

	input := &ses.SetIdentityMailFromDomainInput{
		BehaviorOnMXFailure: aws.String(behaviorOnMxFailure),
		Identity:            aws.String(identityName),
		MailFromDomain:      aws.String(mailFromDomain),
	}

	_, err := conn.SetIdentityMailFromDomain(input)
	if err != nil {
		return fmt.Errorf("Error setting MAIL FROM domain: %s", err)
	}

	d.SetId(identityName)

	return resourceAwsSesIdentityMailFromRead(d, meta)
}

func resourceAwsSesIdentityMailFromRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).sesconn

	identityName := d.Id()

	readOpts := &ses.GetIdentityMailFromDomainAttributesInput{
		Identities: []*string{
			aws.String(identityName),
		},
	}

	out, err := conn.GetIdentityMailFromDomainAttributes(readOpts)

	if err != nil {
		return fmt.Errorf("error fetching SES MAIL FROM domain attributes for %s: %s", identityName, err)
	}

	if out == nil {
		return fmt.Errorf("error fetching SES MAIL FROM domain attributes for %s: empty response", identityName)
	}

	attributes, ok := out.MailFromDomainAttributes[identityName]

	if !ok {
		log.Printf("[WARN] SES Domain Identity (%s) not found, removing from state", identityName)
		d.SetId("")
		return nil
	}

	d.Set("behavior_on_mx_failure", attributes.BehaviorOnMXFailure)
	d.Set("identity", identityName)
	d.Set("mail_from_domain", attributes.MailFromDomain)

	return nil
}

func resourceAwsSesIdentityMailFromDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).sesconn

	identityName := d.Id()

	deleteOpts := &ses.SetIdentityMailFromDomainInput{
		Identity:       aws.String(identityName),
		MailFromDomain: nil,
	}

	_, err := conn.SetIdentityMailFromDomain(deleteOpts)
	if err != nil {
		return fmt.Errorf("Error deleting SES domain identity: %s", err)
	}

	return nil
}
