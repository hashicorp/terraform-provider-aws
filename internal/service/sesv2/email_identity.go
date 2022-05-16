package sesv2

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
			"identity": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
				StateFunc: func(v interface{}) string {
					return strings.TrimSuffix(v.(string), ".")
				},
			},
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"identity_type": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"origin": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"dkim_tokens": {
				Type:     schema.TypeList,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
		},
	}
}

func resourceEmailIdentityCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).SESV2Conn

	identity := d.Get("identity").(string)
	identity = strings.TrimSuffix(identity, ".")

	createOpts := &sesv2.CreateEmailIdentityInput{
		EmailIdentity: aws.String(identity),
	}

	_, err := conn.CreateEmailIdentity(createOpts)
	if err != nil {
		return fmt.Errorf("Error requesting SES identity verification: %s", err)
	}

	d.SetId(identity)

	return resourceEmailIdentityRead(d, meta)
}

func resourceEmailIdentityRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).SESV2Conn

	identity := d.Id()
	d.Set("identity", identity)

	readOpts := &sesv2.GetEmailIdentityInput{
		EmailIdentity: aws.String(identity),
	}

	response, err := conn.GetEmailIdentity(readOpts)
	if err != nil {
		log.Printf("[WARN] Error fetching identity verification attributes for %s: %s", d.Id(), err)
		return err
	}

	var origin string
	var tokens []string
	if response.DkimAttributes != nil {
		origin = aws.StringValue(response.DkimAttributes.SigningAttributesOrigin)
		tokens = aws.StringValueSlice(response.DkimAttributes.Tokens)
	}
	d.Set("origin", origin)
	d.Set("dkim_tokens", tokens)
	d.Set("identity_type", aws.StringValue(response.IdentityType))

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

	identity := d.Get("identity").(string)

	deleteOpts := &sesv2.DeleteEmailIdentityInput{
		EmailIdentity: aws.String(identity),
	}

	_, err := conn.DeleteEmailIdentity(deleteOpts)
	if err != nil {
		if _, ok := err.(*sesv2.NotFoundException); ok {
			return nil
		}
		return fmt.Errorf("Error deleting SES identity: %s", err)
	}

	return nil
}
