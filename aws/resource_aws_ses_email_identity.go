package aws

import (
	"fmt"
	"log"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/arn"
	"github.com/aws/aws-sdk-go/service/ses"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
)

func ResourceEmailIdentity() *schema.Resource {
	return &schema.Resource{
		Create: resourceEmailIdentityCreate,
		Read:   resourceEmailIdentityRead,
		Delete: resourceEmailIdentityDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"email": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
				StateFunc: func(v interface{}) string {
					return strings.TrimSuffix(v.(string), ".")
				},
			},
		},
	}
}

func resourceEmailIdentityCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).SESConn

	email := d.Get("email").(string)
	email = strings.TrimSuffix(email, ".")

	createOpts := &ses.VerifyEmailIdentityInput{
		EmailAddress: aws.String(email),
	}

	_, err := conn.VerifyEmailIdentity(createOpts)
	if err != nil {
		return fmt.Errorf("Error requesting SES email identity verification: %s", err)
	}

	d.SetId(email)

	return resourceEmailIdentityRead(d, meta)
}

func resourceEmailIdentityRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).SESConn

	email := d.Id()
	d.Set("email", email)

	readOpts := &ses.GetIdentityVerificationAttributesInput{
		Identities: []*string{
			aws.String(email),
		},
	}

	response, err := conn.GetIdentityVerificationAttributes(readOpts)
	if err != nil {
		log.Printf("[WARN] Error fetching identity verification attributes for %s: %s", d.Id(), err)
		return err
	}

	_, ok := response.VerificationAttributes[email]
	if !ok {
		log.Printf("[WARN] Email not listed in response when fetching verification attributes for %s", d.Id())
		d.SetId("")
		return nil
	}

	arn := arn.ARN{
		AccountID: meta.(*conns.AWSClient).AccountID,
		Partition: meta.(*conns.AWSClient).Partition,
		Region:    meta.(*conns.AWSClient).Region,
		Resource:  fmt.Sprintf("identity/%s", d.Id()),
		Service:   "ses",
	}.String()
	d.Set("arn", arn)
	return nil
}

func resourceEmailIdentityDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).SESConn

	email := d.Get("email").(string)

	deleteOpts := &ses.DeleteIdentityInput{
		Identity: aws.String(email),
	}

	_, err := conn.DeleteIdentity(deleteOpts)
	if err != nil {
		return fmt.Errorf("Error deleting SES email identity: %s", err)
	}

	return nil
}
