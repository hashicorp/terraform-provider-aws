package aws

import (
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/acmpca"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/aws/internal/service/acmpca/finder"
	"github.com/hashicorp/terraform-provider-aws/aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

func ResourceCertificateAuthorityCertificate() *schema.Resource {
	return &schema.Resource{
		Create: resourceCertificateAuthorityCertificateCreate,
		Read:   resourceCertificateAuthorityCertificateRead,
		Delete: schema.Noop,

		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"certificate": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringLenBetween(1, 32768),
			},
			"certificate_authority_arn": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: verify.ValidARN,
			},
			"certificate_chain": {
				Type:         schema.TypeString,
				Optional:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringLenBetween(0, 2097152),
			},
		},
	}
}

func resourceCertificateAuthorityCertificateCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).ACMPCAConn

	certificateAuthorityArn := d.Get("certificate_authority_arn").(string)

	input := &acmpca.ImportCertificateAuthorityCertificateInput{
		Certificate:             []byte(d.Get("certificate").(string)),
		CertificateAuthorityArn: aws.String(certificateAuthorityArn),
	}
	if v, ok := d.Get("certificate_chain").(string); ok && v != "" {
		input.CertificateChain = []byte(v)
	}

	_, err := conn.ImportCertificateAuthorityCertificate(input)
	if err != nil {
		return fmt.Errorf("error associating ACM PCA Certificate with Certificate Authority (%s): %w", certificateAuthorityArn, err)
	}

	d.SetId(certificateAuthorityArn)

	return resourceCertificateAuthorityCertificateRead(d, meta)
}

func resourceCertificateAuthorityCertificateRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).ACMPCAConn

	output, err := finder.CertificateAuthorityCertificateByARN(conn, d.Id())
	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] ACM PCA Certificate Authority Certificate (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}
	if err != nil {
		return fmt.Errorf("error reading ACM PCA Certificate Authority Certificate (%s): %w", d.Id(), err)
	}

	d.Set("certificate_authority_arn", d.Id())
	d.Set("certificate", output.Certificate)
	d.Set("certificate_chain", output.CertificateChain)

	return nil
}
