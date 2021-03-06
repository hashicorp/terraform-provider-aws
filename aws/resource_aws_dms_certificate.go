package aws

import (
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	dms "github.com/aws/aws-sdk-go/service/databasemigrationservice"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/keyvaluetags"
)

func resourceAwsDmsCertificate() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsDmsCertificateCreate,
		Read:   resourceAwsDmsCertificateRead,
		Update: resourceAwsDmsCertificateUpdate,
		Delete: resourceAwsDmsCertificateDelete,

		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"certificate_arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"certificate_id": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validateDmsCertificateId,
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
			"tags": tagsSchema(),
		},
	}
}

func resourceAwsDmsCertificateCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).dmsconn

	request := &dms.ImportCertificateInput{
		CertificateIdentifier: aws.String(d.Get("certificate_id").(string)),
		Tags:                  keyvaluetags.New(d.Get("tags").(map[string]interface{})).IgnoreAws().DatabasemigrationserviceTags(),
	}

	pem, pemSet := d.GetOk("certificate_pem")
	wallet, walletSet := d.GetOk("certificate_wallet")

	if !pemSet && !walletSet {
		return fmt.Errorf("Must set either certificate_pem and certificate_wallet.")
	}
	if pemSet && walletSet {
		return fmt.Errorf("Cannot set both certificate_pem and certificate_wallet.")
	}

	if pemSet {
		request.CertificatePem = aws.String(pem.(string))
	}
	if walletSet {
		request.CertificateWallet = []byte(wallet.(string))
	}

	log.Println("[DEBUG] DMS import certificate:", request)

	_, err := conn.ImportCertificate(request)
	if err != nil {
		return err
	}

	d.SetId(d.Get("certificate_id").(string))
	return resourceAwsDmsCertificateRead(d, meta)
}

func resourceAwsDmsCertificateRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).dmsconn
	ignoreTagsConfig := meta.(*AWSClient).IgnoreTagsConfig

	response, err := conn.DescribeCertificates(&dms.DescribeCertificatesInput{
		Filters: []*dms.Filter{
			{
				Name:   aws.String("certificate-id"),
				Values: []*string{aws.String(d.Id())}, // Must use d.Id() to work with import.
			},
		},
	})
	if err != nil {
		if dmserr, ok := err.(awserr.Error); ok && dmserr.Code() == "ResourceNotFoundFault" {
			d.SetId("")
			return nil
		}
		return err
	}

	err = resourceAwsDmsCertificateSetState(d, response.Certificates[0])
	if err != nil {
		return err
	}

	tags, err := keyvaluetags.DatabasemigrationserviceListTags(conn, d.Get("certificate_arn").(string))

	if err != nil {
		return fmt.Errorf("error listing tags for DMS Certificate (%s): %w", d.Get("certificate_arn").(string), err)
	}

	if err := d.Set("tags", tags.IgnoreAws().IgnoreConfig(ignoreTagsConfig).Map()); err != nil {
		return fmt.Errorf("error setting tags: %w", err)
	}

	return nil
}

func resourceAwsDmsCertificateUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).dmsconn

	if d.HasChange("tags") {
		arn := d.Get("certificate_arn").(string)
		o, n := d.GetChange("tags")

		if err := keyvaluetags.DatabasemigrationserviceUpdateTags(conn, arn, o, n); err != nil {
			return fmt.Errorf("error updating DMS Certificate (%s) tags: %w", arn, err)
		}
	}

	return resourceAwsDmsCertificateRead(d, meta)
}

func resourceAwsDmsCertificateDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).dmsconn

	request := &dms.DeleteCertificateInput{
		CertificateArn: aws.String(d.Get("certificate_arn").(string)),
	}

	log.Printf("[DEBUG] DMS delete certificate: %#v", request)

	_, err := conn.DeleteCertificate(request)
	return err
}

func resourceAwsDmsCertificateSetState(d *schema.ResourceData, cert *dms.Certificate) error {
	d.SetId(aws.StringValue(cert.CertificateIdentifier))

	d.Set("certificate_id", cert.CertificateIdentifier)
	d.Set("certificate_arn", cert.CertificateArn)

	if cert.CertificatePem != nil && *cert.CertificatePem != "" {
		d.Set("certificate_pem", cert.CertificatePem)
	}
	if cert.CertificateWallet != nil && len(cert.CertificateWallet) != 0 {
		d.Set("certificate_wallet", string(cert.CertificateWallet))
	}

	return nil
}
