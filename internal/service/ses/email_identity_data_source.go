package ses

import (
	"fmt"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/arn"
	"github.com/aws/aws-sdk-go/service/ses"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
)

func DataSourceEmailIdentity() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceEmailIdentityRead,
		Schema: map[string]*schema.Schema{
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"email": {
				Type:     schema.TypeString,
				Required: true,
				StateFunc: func(v interface{}) string {
					return strings.TrimSuffix(v.(string), ".")
				},
			},
		},
	}
}

func dataSourceEmailIdentityRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).SESConn

	email := d.Get("email").(string)
	email = strings.TrimSuffix(email, ".")

	d.SetId(email)
	d.Set("email", email)

	readOpts := &ses.GetIdentityVerificationAttributesInput{
		Identities: []*string{
			aws.String(email),
		},
	}

	response, err := conn.GetIdentityVerificationAttributes(readOpts)
	if err != nil {
		return fmt.Errorf("[WARN] Error fetching identity verification attributes for %s: %s", email, err)
	}

	_, ok := response.VerificationAttributes[email]
	if !ok {
		return fmt.Errorf("[WARN] Email not listed in response when fetching verification attributes for %s", d.Id())
	}

	arn := arn.ARN{
		AccountID: meta.(*conns.AWSClient).AccountID,
		Partition: meta.(*conns.AWSClient).Partition,
		Region:    meta.(*conns.AWSClient).Region,
		Resource:  fmt.Sprintf("identity/%s", email),
		Service:   "ses",
	}.String()
	d.Set("arn", arn)
	return nil
}
