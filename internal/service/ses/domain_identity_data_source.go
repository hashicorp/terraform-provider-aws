package ses

import (
	"context"
	"fmt"
	"regexp"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/arn"
	"github.com/aws/aws-sdk-go/service/ses"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
)

func DataSourceDomainIdentity() *schema.Resource {
	return &schema.Resource{
		ReadWithoutTimeout: dataSourceDomainIdentityRead,
		Schema: map[string]*schema.Schema{
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"domain": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringDoesNotMatch(regexp.MustCompile(`\.$`), "cannot end with a period"),
			},
			"verification_token": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func dataSourceDomainIdentityRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).SESConn()

	domainName := d.Get("domain").(string)
	d.SetId(domainName)
	d.Set("domain", domainName)

	readOpts := &ses.GetIdentityVerificationAttributesInput{
		Identities: []*string{
			aws.String(domainName),
		},
	}

	response, err := conn.GetIdentityVerificationAttributesWithContext(ctx, readOpts)
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "[WARN] Error fetching identity verification attributes for %s: %s", domainName, err)
	}

	verificationAttrs, ok := response.VerificationAttributes[domainName]
	if !ok {
		return sdkdiag.AppendErrorf(diags, "[WARN] Domain not listed in response when fetching verification attributes for %s", domainName)
	}

	arn := arn.ARN{
		Partition: meta.(*conns.AWSClient).Partition,
		Service:   "ses",
		Region:    meta.(*conns.AWSClient).Region,
		AccountID: meta.(*conns.AWSClient).AccountID,
		Resource:  fmt.Sprintf("identity/%s", domainName),
	}.String()
	d.Set("arn", arn)
	d.Set("verification_token", verificationAttrs.VerificationToken)
	return diags
}
