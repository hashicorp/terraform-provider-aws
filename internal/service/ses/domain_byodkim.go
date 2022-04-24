package ses

import (
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/sesv2"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
)

func ResourceDomainBYODKIM() *schema.Resource {
	return &schema.Resource{
		Create: resourceDomainBYODDKIMCreate,
		Read:   resourceDomainBYODDKIMRead,
		Update: resourceDomainBYODDKIMUpdate,
		Delete: resourceDomainBYODDKIMDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"domain": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"selector": {
				Type:     schema.TypeString,
				ForceNew: true,
				Required: true,
			},
			"private_key": {
				Type:     schema.TypeString,
				ForceNew: true,
				Required: true,
			},
			"dkim_tokens": {
				Type:     schema.TypeList,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
				Optional: true,
			},
			"public_key_name": {
				Type:        schema.TypeString,
				ForceNew:    true,
				Optional:    true,
				Description: "In order to complete the domain configuration, use public_key_name as the NAME for the TXT record, and the public key as the value.",
			},
		},
	}
}

func resourceDomainBYODDKIMCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).SESV2Conn

	domainName := d.Get("domain").(string)

	var selector string
	if v, ok := d.GetOk("selector"); ok {
		selector = v.(string)
	}

	var privateKey string
	if v, ok := d.GetOk("private_key"); ok {
		privateKey = v.(string)
	}

	input := &sesv2.GetEmailIdentityInput{
		EmailIdentity: aws.String(domainName),
	}

	_, err := conn.GetEmailIdentity(input)
	if err != nil {
		return fmt.Errorf("can't get domain %s: %v", domainName, err)
	}

	opts := &sesv2.PutEmailIdentityDkimSigningAttributesInput{
		EmailIdentity: aws.String(domainName),
		SigningAttributes: &sesv2.DkimSigningAttributes{
			DomainSigningPrivateKey: aws.String(privateKey),
			DomainSigningSelector:   aws.String(selector),
		},
		SigningAttributesOrigin: aws.String("EXTERNAL"),
	}
	_, err = conn.PutEmailIdentityDkimSigningAttributes(opts)
	if err != nil {
		return fmt.Errorf("can't put dkim: %v", err)
	}

	d.SetId(domainName)

	return resourceDomainBYODDKIMRead(d, meta)
}

func resourceDomainBYODDKIMRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).SESV2Conn

	domainName := d.Id()
	d.Set("domain", domainName)

	readOpts := &sesv2.GetEmailIdentityInput{
		EmailIdentity: aws.String(domainName),
	}

	res, err := conn.GetEmailIdentity(readOpts)
	if err != nil {
		log.Printf("[WARN] Error fetching identity verification attributes for %s: %s", d.Id(), err)
		return err
	}

	if res.DkimAttributes == nil {
		log.Printf("[WARN] Domain not listed in response when fetching verification attributes for %s", d.Id())
		d.SetId("")
		return nil
	}

	d.Set("dkim_tokens", aws.StringValueSlice(res.DkimAttributes.Tokens))

	var selector string
	if v, ok := d.GetOk("selector"); ok {
		selector = v.(string)
	}

	nameRes := selector + "._domainkey." + domainName
	d.Set("public_key_name", nameRes)

	return nil
}

func resourceDomainBYODDKIMUpdate(d *schema.ResourceData, meta interface{}) error {
	return nil
}

func resourceDomainBYODDKIMDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).SESV2Conn

	domainName := d.Id()
	deleteOpts := &sesv2.DeleteEmailIdentityInput{EmailIdentity: aws.String(domainName)}
	_, err := conn.DeleteEmailIdentity(deleteOpts)
	if err != nil {
		return fmt.Errorf("error deleting identity %s: %s", d.Id(), err)
	}
	return nil
}
