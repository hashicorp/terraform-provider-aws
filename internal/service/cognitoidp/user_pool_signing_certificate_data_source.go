package cognitoidp

import (
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/cognitoidentityprovider"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
)

func DataSourceUserPoolSigningCertificate() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceUserPoolSigningCertificateRead,

		Schema: map[string]*schema.Schema{
			"certificate": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"user_pool_id": {
				Type:     schema.TypeString,
				Required: true,
			},
		},
	}
}

func dataSourceUserPoolSigningCertificateRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).CognitoIDPConn

	userPoolID := d.Get("user_pool_id").(string)
	input := &cognitoidentityprovider.GetSigningCertificateInput{
		UserPoolId: aws.String(userPoolID),
	}

	output, err := conn.GetSigningCertificate(input)

	if err != nil {
		return fmt.Errorf("error reading Cognito User Pool (%s) Signing Certificate: %w", userPoolID, err)
	}

	d.SetId(userPoolID)
	d.Set("certificate", output.Certificate)

	return nil
}
