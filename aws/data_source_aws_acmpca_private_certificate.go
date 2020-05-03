package aws

import (
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/acmpca"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
)

func dataSourceAwsAcmpcaPrivateCertificate() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceAwsAcmpcaPrivateCertificateRead,

		Schema: map[string]*schema.Schema{
			"arn": {
				Type:     schema.TypeString,
				Required: true,
			},
			"certificate_authority_arn": {
				Type:     schema.TypeString,
				Required: true,
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

func dataSourceAwsAcmpcaPrivateCertificateRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).acmpcaconn
	certificateArn := d.Get("arn").(string)

	getCertificateInput := &acmpca.GetCertificateInput{
		CertificateArn:          aws.String(certificateArn),
		CertificateAuthorityArn: aws.String(d.Get("certificate_authority_arn").(string)),
	}

	log.Printf("[DEBUG] Reading ACMPCA Certificate: %s", getCertificateInput)

	certificateOutput, err := conn.GetCertificate(getCertificateInput)
	if err != nil {
		return fmt.Errorf("error reading ACMPCA Certificate: %s", err)
	}

	d.SetId(certificateArn)
	d.Set("certificate", aws.StringValue(certificateOutput.Certificate))
	d.Set("certificate_chain", aws.StringValue(certificateOutput.CertificateChain))

	return nil
}
