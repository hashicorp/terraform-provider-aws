package dms

import (
	"encoding/base64"
	"fmt"
	"log"
	"regexp"

	"github.com/aws/aws-sdk-go/aws"
	dms "github.com/aws/aws-sdk-go/service/databasemigrationservice"
	"github.com/hashicorp/aws-sdk-go-base/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

func ResourceCertificate() *schema.Resource {
	return &schema.Resource{
		Create: resourceCertificateCreate,
		Read:   resourceCertificateRead,
		Update: resourceCertificateUpdate,
		Delete: resourceCertificateDelete,

		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"certificate_arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"certificate_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
				ValidateFunc: validation.All(
					validation.StringLenBetween(1, 255),
					validation.StringMatch(regexp.MustCompile("^[a-zA-Z][a-zA-Z0-9-]+$"), "must start with a letter, only contain alphanumeric characters and hyphens"),
					validation.StringDoesNotMatch(regexp.MustCompile(`--`), "cannot contain two consecutive hyphens"),
					validation.StringDoesNotMatch(regexp.MustCompile(`-$`), "cannot end in a hyphen"),
				),
			},
			"certificate_pem": {
				Type:      schema.TypeString,
				Optional:  true,
				ForceNew:  true,
				Sensitive: true,
			},
			"certificate_wallet": {
				Type:      schema.TypeString,
				Optional:  true,
				ForceNew:  true,
				Sensitive: true,
			},
			"tags":     tftags.TagsSchema(),
			"tags_all": tftags.TagsSchemaComputed(),
		},

		CustomizeDiff: verify.SetTagsDiff,
	}
}

func resourceCertificateCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).DMSConn
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	tags := defaultTagsConfig.MergeTags(tftags.New(d.Get("tags").(map[string]interface{})))
	certificateID := d.Get("certificate_id").(string)

	request := &dms.ImportCertificateInput{
		CertificateIdentifier: aws.String(certificateID),
		Tags:                  tags.IgnoreAws().DatabasemigrationserviceTags(),
	}

	pem, pemSet := d.GetOk("certificate_pem")
	wallet, walletSet := d.GetOk("certificate_wallet")

	if !pemSet && !walletSet {
		return fmt.Errorf("Must set either certificate_pem or certificate_wallet for DMS Certificate (%s)", certificateID)
	}
	if pemSet && walletSet {
		return fmt.Errorf("Cannot set both certificate_pem and certificate_wallet for DMS Certificate (%s)", certificateID)
	}

	if pemSet {
		request.CertificatePem = aws.String(pem.(string))
	}
	if walletSet {
		certWallet, err := base64.StdEncoding.DecodeString(wallet.(string))
		if err != nil {
			return fmt.Errorf("error Base64 decoding certificate_wallet for DMS Certificate (%s): %w", certificateID, err)
		}
		request.CertificateWallet = certWallet
	}

	_, err := conn.ImportCertificate(request)
	if err != nil {
		return fmt.Errorf("error creating DMS certificate (%s): %w", certificateID, err)
	}

	d.SetId(certificateID)
	return resourceCertificateRead(d, meta)
}

func resourceCertificateRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).DMSConn
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	ignoreTagsConfig := meta.(*conns.AWSClient).IgnoreTagsConfig

	response, err := conn.DescribeCertificates(&dms.DescribeCertificatesInput{
		Filters: []*dms.Filter{
			{
				Name:   aws.String("certificate-id"),
				Values: []*string{aws.String(d.Id())}, // Must use d.Id() to work with import.
			},
		},
	})

	if !d.IsNewResource() && tfawserr.ErrCodeEquals(err, dms.ErrCodeResourceNotFoundFault) {
		log.Printf("[WARN] DMS Certificate (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return fmt.Errorf("error reading DMS Certificate (%s): %w", d.Id(), err)
	}

	if response == nil || len(response.Certificates) == 0 || response.Certificates[0] == nil {
		if d.IsNewResource() {
			return fmt.Errorf("error reading DMS Certificate (%s): not found", d.Id())
		}
		log.Printf("[WARN] DMS Certificate (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	err = resourceAwsDmsCertificateSetState(d, response.Certificates[0])
	if err != nil {
		return err
	}

	tags, err := tftags.DatabasemigrationserviceListTags(conn, d.Get("certificate_arn").(string))

	if err != nil {
		return fmt.Errorf("error listing tags for DMS Certificate (%s): %w", d.Get("certificate_arn").(string), err)
	}

	tags = tags.IgnoreAws().IgnoreConfig(ignoreTagsConfig)

	//lintignore:AWSR002
	if err := d.Set("tags", tags.RemoveDefaultConfig(defaultTagsConfig).Map()); err != nil {
		return fmt.Errorf("error setting tags: %w", err)
	}

	if err := d.Set("tags_all", tags.Map()); err != nil {
		return fmt.Errorf("error setting tags_all: %w", err)
	}

	return nil
}

func resourceCertificateUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).DMSConn

	if d.HasChange("tags_all") {
		arn := d.Get("certificate_arn").(string)
		o, n := d.GetChange("tags_all")

		if err := tftags.DatabasemigrationserviceUpdateTags(conn, arn, o, n); err != nil {
			return fmt.Errorf("error updating DMS Certificate (%s) tags: %w", arn, err)
		}
	}

	return resourceCertificateRead(d, meta)
}

func resourceCertificateDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).DMSConn

	request := &dms.DeleteCertificateInput{
		CertificateArn: aws.String(d.Get("certificate_arn").(string)),
	}

	_, err := conn.DeleteCertificate(request)

	if err != nil {
		if tfawserr.ErrCodeEquals(err, dms.ErrCodeResourceNotFoundFault) {
			return nil
		}
		return fmt.Errorf("error deleting DMS Certificate (%s): %w", d.Id(), err)
	}

	return nil
}

func resourceAwsDmsCertificateSetState(d *schema.ResourceData, cert *dms.Certificate) error {
	d.SetId(aws.StringValue(cert.CertificateIdentifier))

	d.Set("certificate_id", cert.CertificateIdentifier)
	d.Set("certificate_arn", cert.CertificateArn)

	if cert.CertificatePem != nil && *cert.CertificatePem != "" {
		d.Set("certificate_pem", cert.CertificatePem)
	}
	if cert.CertificateWallet != nil && len(cert.CertificateWallet) != 0 {
		d.Set("certificate_wallet", verify.Base64Encode(cert.CertificateWallet))
	}

	return nil
}
