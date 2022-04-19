package ses

import (
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/sesv2"
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
			"selector": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"private_key": {
				Type:     schema.TypeString,
				Optional: true,
			},
		},
	}
}

func resourceDomainDKIMCreate(d *schema.ResourceData, meta interface{}) error {
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

	createOpts := &sesv2.CreateEmailIdentityInput{
		EmailIdentity: aws.String(domainName),
	}
	// ByoDKIM:
	if selector != "" {
		createOpts = &sesv2.CreateEmailIdentityInput{
			EmailIdentity: aws.String(domainName),
			DkimSigningAttributes: &sesv2.DkimSigningAttributes{
				DomainSigningPrivateKey: aws.String(privateKey),
				DomainSigningSelector:   aws.String(selector),
			},
		}
	}
	_, err := conn.CreateEmailIdentity(createOpts)
	if err != nil {
		return fmt.Errorf("error requesting SES domain identity verification: %s", err)
	}

	d.SetId(domainName)

	return resourceDomainDKIMRead(d, meta)
}

func resourceDomainDKIMRead(d *schema.ResourceData, meta interface{}) error {
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
	return nil
}

func resourceDomainDKIMDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).SESV2Conn

	domainName := d.Id()
	deleteOpts := &sesv2.DeleteEmailIdentityInput{EmailIdentity: aws.String(domainName)}
	_, err := conn.DeleteEmailIdentity(deleteOpts)
	if err != nil {
		return fmt.Errorf("error deleting identity %s: %s", d.Id(), err)
	}
	return nil
}
