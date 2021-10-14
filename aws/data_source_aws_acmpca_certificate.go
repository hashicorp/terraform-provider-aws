package aws

import (
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/acmpca"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

func DataSourceCertificate() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceCertificateRead,

		Schema: map[string]*schema.Schema{
			"arn": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: verify.ValidARN,
			},
			"certificate_authority_arn": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: verify.ValidARN,
			},
			"certificate": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"certificate_chain": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func dataSourceCertificateRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).ACMPCAConn
	certificateArn := d.Get("arn").(string)

	getCertificateInput := &acmpca.GetCertificateInput{
		CertificateArn:          aws.String(certificateArn),
		CertificateAuthorityArn: aws.String(d.Get("certificate_authority_arn").(string)),
	}

	log.Printf("[DEBUG] Reading ACM PCA Certificate: %s", getCertificateInput)

	certificateOutput, err := conn.GetCertificate(getCertificateInput)
	if err != nil {
		return fmt.Errorf("error reading ACM PCA Certificate (%s): %w", certificateArn, err)
	}

	d.SetId(certificateArn)
	d.Set("certificate", certificateOutput.Certificate)
	d.Set("certificate_chain", certificateOutput.CertificateChain)

	return nil
}
