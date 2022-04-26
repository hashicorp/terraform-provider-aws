package ses

import (
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/sesv2"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
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
			"origin": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
				ValidateFunc: validation.StringInSlice([]string{
					sesv2.DkimSigningAttributesOriginAwsSes,
				}, false),
			},
			"dkim_tokens": {
				Type:     schema.TypeList,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"status": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func resourceDomainDKIMCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).SESV2Conn

	domainName := d.Get("domain").(string)

	createOpts := &sesv2.PutEmailIdentityDkimSigningAttributesInput{
		EmailIdentity: aws.String(domainName),
	}

	if v, ok := d.GetOk("origin"); ok {
		createOpts.SigningAttributesOrigin = aws.String(v.(string))
	}

	_, err := conn.PutEmailIdentityDkimSigningAttributes(createOpts)
	if err != nil {
		return fmt.Errorf("Error requesting SES domain identity verification: %s", err)
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

	response, err := conn.GetEmailIdentity(readOpts)
	if err != nil {
		log.Printf("[WARN] Error fetching identity verification attributes for %s: %s", d.Id(), err)
		return err
	}

	if response.DkimAttributes == nil {
		log.Printf("[WARN] Domain not listed in response when fetching verification attributes for %s", d.Id())
		d.SetId("")
		return nil
	}

	d.Set("origin", aws.StringValue(response.DkimAttributes.SigningAttributesOrigin))
	d.Set("dkim_tokens", aws.StringValueSlice(response.DkimAttributes.Tokens))
	d.Set("status", aws.StringValue(response.DkimAttributes.Status))
	return nil
}

func resourceDomainDKIMDelete(d *schema.ResourceData, meta interface{}) error {

	return nil
}
