package ses

import (
	"fmt"
	"log"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/arn"
	"github.com/aws/aws-sdk-go/service/sesv2"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
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
	conn := meta.(*conns.AWSClient).SESV2Conn

	email := d.Get("email").(string)
	email = strings.TrimSuffix(email, ".")

	createOpts := &sesv2.CreateEmailIdentityInput{
		EmailIdentity: aws.String(email),
	}

	_, err := conn.CreateEmailIdentity(createOpts)
	if err != nil {
		return fmt.Errorf("error creating SES email identity: %w", err)
	}

	d.SetId(email)

	return resourceEmailIdentityRead(d, meta)
}

func resourceEmailIdentityRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).SESV2Conn

	email := d.Id()
	d.Set("email", email)

	readOpts := &sesv2.GetEmailIdentityInput{
		EmailIdentity: aws.String(email),
	}

	_, err := conn.GetEmailIdentity(readOpts)
	if err != nil {
		log.Printf("[WARN] Error reading SES email identity %s: %s", d.Id(), err)
		return err
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
	conn := meta.(*conns.AWSClient).SESV2Conn

	email := d.Get("email").(string)

	deleteOpts := &sesv2.DeleteEmailIdentityInput{
		EmailIdentity: aws.String(email),
	}

	_, err := conn.DeleteEmailIdentity(deleteOpts)
	if err != nil {
		return fmt.Errorf("error deleting SES email identity: %w", err)
	}

	return nil
}
