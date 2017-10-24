package aws

import (
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ses"
	"github.com/hashicorp/terraform/helper/schema"
)

func resourceAwsSesDomainMailFrom() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsSesDomainMailFromCreate,
		Read:   resourceAwsSesDomainMailFromRead,
		Delete: resourceAwsSesDomainMailFromDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"domain": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"mail_from_domain": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
		},
	}
}

func resourceAwsSesDomainMailFromCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).sesConn

	domainName := d.Get("domain").(string)
	mailFromDomain := d.Get("mail_from_domain").(string)

	createOpts := &ses.SetIdentityMailFromDomainInput{
		BehaviorOnMXFailure: aws.String("UseDefaultValue"),
		Identity:            aws.String(domainName),
		MailFromDomain:      aws.String(mailFromDomain),
	}

	_, err := conn.SetIdentityMailFromDomain(createOpts)
	if err != nil {
		return fmt.Errorf("Error setting MAIL FROM domain: %s", err)
	}

	d.SetId(domainName)

	return resourceAwsSesDomainMailFromRead(d, meta)
}

func resourceAwsSesDomainMailFromRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).sesConn

	domainName := d.Id()
	d.Set("domain", domainName)

	readOpts := &ses.GetIdentityMailFromDomainAttributesInput{
		Identities: []*string{
			aws.String(domainName),
		},
	}

	_, err := conn.GetIdentityMailFromDomainAttributes(readOpts)
	if err != nil {
		log.Printf("[WARN] Error fetching MAIL FROM domain attributes for %s: %s", d.Id(), err)
		return err
	}

	return nil
}

func resourceAwsSesDomainMailFromDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).sesConn

	domainName := d.Get("domain").(string)

	deleteOpts := &ses.SetIdentityMailFromDomainInput{
		BehaviorOnMXFailure: aws.String("UseDefaultValue"),
		Identity:            aws.String(domainName),
		MailFromDomain:      nil,
	}

	_, err := conn.SetIdentityMailFromDomain(deleteOpts)
	if err != nil {
		return fmt.Errorf("Error deleting SES domain identity: %s", err)
	}

	return nil
}
