package aws

import (
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/acm"
	"github.com/hashicorp/terraform/helper/schema"
)

func resourceAwsAcmCertificateImported() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsAcmCertificateImportedCreate,
		Read:   resourceAwsAcmCertificateRead(false),
		Update: resourceAwsAcmCertificateImportedUpdate,
		Delete: resourceAwsAcmCertificateDelete,

		Schema: map[string]*schema.Schema{
			"private_key": {
				Type:      schema.TypeString,
				Required:  true,
				StateFunc: normalizeCert,
				Sensitive: true,
			},
			"certificate_body": {
				Type:      schema.TypeString,
				Required:  true,
				StateFunc: normalizeCert,
			},
			"certificate_chain": {
				Type:      schema.TypeString,
				Optional:  true,
				StateFunc: normalizeCert,
			},
			"domain_name": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"subject_alternative_names": {
				Type:     schema.TypeList,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"tags": tagsSchema(),
		},
	}
}

func resourceAwsAcmCertificateImportedCreate(d *schema.ResourceData, meta interface{}) error {
	acmconn := meta.(*AWSClient).acmconn
	resp, err := resourceAwsAcmCertificateImport(acmconn, d, false)
	if err != nil {
		return fmt.Errorf("Error importing certificate: %s", err)
	}

	d.SetId(*resp.CertificateArn)
	if v, ok := d.GetOk("tags"); ok {
		params := &acm.AddTagsToCertificateInput{
			CertificateArn: resp.CertificateArn,
			Tags:           tagsFromMapACM(v.(map[string]interface{})),
		}
		_, err := acmconn.AddTagsToCertificate(params)

		if err != nil {
			return fmt.Errorf("Error requesting certificate: %s", err)
		}
	}

	return resourceAwsAcmCertificateRead(false)(d, meta)
}

func resourceAwsAcmCertificateImportedUpdate(d *schema.ResourceData, meta interface{}) error {
	acmconn := meta.(*AWSClient).acmconn

	if d.HasChange("private_key") || d.HasChange("certificate_body") || d.HasChange("certificate_chain") {
		_, err := resourceAwsAcmCertificateImport(acmconn, d, true)
		if err != nil {
			return fmt.Errorf("Error updating certificate: %s", err)
		}
	}

	if d.HasChange("tags") {
		err := setTagsACM(acmconn, d)
		if err != nil {
			return err
		}
	}
	return nil
}

func resourceAwsAcmCertificateImport(conn *acm.ACM, d *schema.ResourceData, update bool) (*acm.ImportCertificateOutput, error) {
	params := &acm.ImportCertificateInput{
		PrivateKey:  []byte(d.Get("private_key").(string)),
		Certificate: []byte(d.Get("certificate_body").(string)),
	}
	if chain, ok := d.GetOk("certificate_chain"); ok {
		params.CertificateChain = []byte(chain.(string))
	}
	if update {
		params.CertificateArn = aws.String(d.Get("arn").(string))
	}

	log.Printf("[DEBUG] ACM Certificate Import: %#v", params)
	return conn.ImportCertificate(params)
}
