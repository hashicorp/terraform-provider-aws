package aws

import (
	"fmt"
	"log"
	"regexp"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/arn"
	"github.com/aws/aws-sdk-go/service/ses"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

func ResourceDomainIdentity() *schema.Resource {
	return &schema.Resource{
		Create: resourceDomainIdentityCreate,
		Read:   resourceDomainIdentityRead,
		Delete: resourceDomainIdentityDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"domain": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringDoesNotMatch(regexp.MustCompile(`\.$`), "cannot end with a period"),
			},
			"verification_token": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func resourceDomainIdentityCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).SESConn

	domainName := d.Get("domain").(string)

	createOpts := &ses.VerifyDomainIdentityInput{
		Domain: aws.String(domainName),
	}

	_, err := conn.VerifyDomainIdentity(createOpts)
	if err != nil {
		return fmt.Errorf("Error requesting SES domain identity verification: %s", err)
	}

	d.SetId(domainName)

	return resourceDomainIdentityRead(d, meta)
}

func resourceDomainIdentityRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).SESConn

	domainName := d.Id()
	d.Set("domain", domainName)

	readOpts := &ses.GetIdentityVerificationAttributesInput{
		Identities: []*string{
			aws.String(domainName),
		},
	}

	response, err := conn.GetIdentityVerificationAttributes(readOpts)
	if err != nil {
		log.Printf("[WARN] Error fetching identity verification attributes for %s: %s", d.Id(), err)
		return err
	}

	verificationAttrs, ok := response.VerificationAttributes[domainName]
	if !ok {
		log.Printf("[WARN] Domain not listed in response when fetching verification attributes for %s", d.Id())
		d.SetId("")
		return nil
	}

	arn := arn.ARN{
		Partition: meta.(*conns.AWSClient).Partition,
		Service:   "ses",
		Region:    meta.(*conns.AWSClient).Region,
		AccountID: meta.(*conns.AWSClient).AccountID,
		Resource:  fmt.Sprintf("identity/%s", d.Id()),
	}.String()
	d.Set("arn", arn)
	d.Set("verification_token", verificationAttrs.VerificationToken)
	return nil
}

func resourceDomainIdentityDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).SESConn

	domainName := d.Get("domain").(string)

	deleteOpts := &ses.DeleteIdentityInput{
		Identity: aws.String(domainName),
	}

	_, err := conn.DeleteIdentity(deleteOpts)
	if err != nil {
		return fmt.Errorf("Error deleting SES domain identity: %s", err)
	}

	return nil
}
