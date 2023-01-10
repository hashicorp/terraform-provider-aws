package ses

import (
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ses"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
)

func ResourceDomainDKIM() *schema.Resource {
	return &schema.Resource{
		Create: resourceDomainDKIMCreate,
		Read:   resourceDomainDKIMRead,
		Delete: resourceDomainDKIMDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"domain": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"dkim_tokens": {
				Type:     schema.TypeList,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
		},
	}
}

func resourceDomainDKIMCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).SESConn()

	domainName := d.Get("domain").(string)

	createOpts := &ses.VerifyDomainDkimInput{
		Domain: aws.String(domainName),
	}

	_, err := conn.VerifyDomainDkim(createOpts)
	if err != nil {
		return fmt.Errorf("Error requesting SES domain identity verification: %s", err)
	}

	d.SetId(domainName)

	return resourceDomainDKIMRead(d, meta)
}

func resourceDomainDKIMRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).SESConn()

	domainName := d.Id()
	d.Set("domain", domainName)

	readOpts := &ses.GetIdentityDkimAttributesInput{
		Identities: []*string{
			aws.String(domainName),
		},
	}

	response, err := conn.GetIdentityDkimAttributes(readOpts)
	if err != nil {
		return fmt.Errorf("reading SES Domain DKIM (%s): %w", d.Id(), err)
	}

	verificationAttrs, ok := response.DkimAttributes[domainName]
	if !ok {
		log.Printf("[WARN] SES Domain DKIM (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	d.Set("dkim_tokens", aws.StringValueSlice(verificationAttrs.DkimTokens))
	return nil
}

func resourceDomainDKIMDelete(d *schema.ResourceData, meta interface{}) error {
	return nil
}
