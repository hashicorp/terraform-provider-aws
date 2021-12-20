package cognitoidp

import (
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/cognitoidentityprovider"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
)

func DataSourceSigningCert() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceUserPoolSigningCertRead,
		Schema: map[string]*schema.Schema{
			"user_pool_id": {
				Type:     schema.TypeString,
				Required: true,
			},
			"certificate": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func dataSourceUserPoolSigningCertRead(d *schema.ResourceData, meta interface{}) error {
	id := d.Get("user_pool_id").(string)
	conn := meta.(*conns.AWSClient).CognitoIDPConn
	result, err := conn.GetSigningCertificate(&cognitoidentityprovider.GetSigningCertificateInput{
		UserPoolId: aws.String(id),
	})
	if err != nil {
		return fmt.Errorf("Error getting signing cert from user pool: %w", err)
	}
	d.SetId(id)
	d.Set("certificate", *result.Certificate)
	return nil
}
