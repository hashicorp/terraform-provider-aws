package cognitoidp

import (
	"context"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/cognitoidentityprovider"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
)

func DataSourceUserPoolSigningCertificate() *schema.Resource {
	return &schema.Resource{
		ReadWithoutTimeout: dataSourceUserPoolSigningCertificateRead,

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

func dataSourceUserPoolSigningCertificateRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).CognitoIDPConn()

	userPoolID := d.Get("user_pool_id").(string)
	input := &cognitoidentityprovider.GetSigningCertificateInput{
		UserPoolId: aws.String(userPoolID),
	}

	output, err := conn.GetSigningCertificateWithContext(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading Cognito User Pool (%s) Signing Certificate: %s", userPoolID, err)
	}

	d.SetId(userPoolID)
	d.Set("certificate", output.Certificate)

	return diags
}
