package iot

import (
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/iot"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
)

func ResourceCertificate() *schema.Resource {
	return &schema.Resource{
		Create: resourceCertificateCreate,
		Read:   resourceCertificateRead,
		Update: resourceCertificateUpdate,
		Delete: resourceCertificateDelete,
		Schema: map[string]*schema.Schema{
			"csr": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
			},
			"active": {
				Type:     schema.TypeBool,
				Required: true,
			},
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"certificate_pem": {
				Type:      schema.TypeString,
				Optional:  true,
				Computed:  true,
				Sensitive: true,
			},
			"ca_pem": {
				Type:      schema.TypeString,
				Optional:  true,
				Sensitive: true,
			},
			"public_key": {
				Type:      schema.TypeString,
				Computed:  true,
				Sensitive: true,
			},
			"private_key": {
				Type:      schema.TypeString,
				Computed:  true,
				Sensitive: true,
			},
		},
	}
}

func resourceCertificateCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).IoTConn

	_, okcert := d.GetOk("certificate_pem")
	_, okCA := d.GetOk("ca_pem")

	cert_status := "INACTIVE"
	if d.Get("active").(bool) {
		cert_status = "ACTIVE"
	}

	if _, ok := d.GetOk("csr"); ok {
		log.Printf("[DEBUG] Creating certificate from CSR")
		out, err := conn.CreateCertificateFromCsr(&iot.CreateCertificateFromCsrInput{
			CertificateSigningRequest: aws.String(d.Get("csr").(string)),
			SetAsActive:               aws.Bool(d.Get("active").(bool)),
		})
		if err != nil {
			return fmt.Errorf("error creating certificate from CSR: %v", err)
		}
		log.Printf("[DEBUG] Created certificate from CSR")

		d.SetId(aws.StringValue(out.CertificateId))
	} else if okcert && okCA {
		log.Printf("[DEBUG] Registering certificate with CA")
		out, err := conn.RegisterCertificate(&iot.RegisterCertificateInput{
			CaCertificatePem: aws.String(d.Get("ca_pem").(string)),
			CertificatePem:   aws.String(d.Get("certificate_pem").(string)),
			Status:           aws.String(cert_status),
		})
		if err != nil {
			return fmt.Errorf("error registering certificate with CA: %v", err)
		}
		log.Printf("[DEBUG] Certificate with CA registered")

		d.SetId(aws.StringValue(out.CertificateId))
	} else if okcert {
		log.Printf("[DEBUG] Registering certificate without CA")
		out, err := conn.RegisterCertificateWithoutCA(&iot.RegisterCertificateWithoutCAInput{
			CertificatePem: aws.String(d.Get("certificate_pem").(string)),
			Status:         aws.String(cert_status),
		})
		if err != nil {
			return fmt.Errorf("error registering certificate without CA: %v", err)
		}
		log.Printf("[DEBUG] Certificate without CA registered")

		d.SetId(aws.StringValue(out.CertificateId))
	} else {
		log.Printf("[DEBUG] Creating keys and certificate")
		out, err := conn.CreateKeysAndCertificate(&iot.CreateKeysAndCertificateInput{
			SetAsActive: aws.Bool(d.Get("active").(bool)),
		})
		if err != nil {
			return fmt.Errorf("error creating keys and certificate: %v", err)
		}
		log.Printf("[DEBUG] Created keys and certificate")

		d.SetId(aws.StringValue(out.CertificateId))
		d.Set("public_key", out.KeyPair.PublicKey)
		d.Set("private_key", out.KeyPair.PrivateKey)
	}

	return resourceCertificateRead(d, meta)
}

func resourceCertificateRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).IoTConn

	out, err := conn.DescribeCertificate(&iot.DescribeCertificateInput{
		CertificateId: aws.String(d.Id()),
	})
	if err != nil {
		return fmt.Errorf("error reading certificate details: %v", err)
	}

	d.Set("active", aws.Bool(aws.StringValue(out.CertificateDescription.Status) == iot.CertificateStatusActive))
	d.Set("arn", out.CertificateDescription.CertificateArn)
	d.Set("certificate_pem", out.CertificateDescription.CertificatePem)

	return nil
}

func resourceCertificateUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).IoTConn

	if d.HasChange("active") {
		status := iot.CertificateStatusInactive
		if d.Get("active").(bool) {
			status = iot.CertificateStatusActive
		}

		_, err := conn.UpdateCertificate(&iot.UpdateCertificateInput{
			CertificateId: aws.String(d.Id()),
			NewStatus:     aws.String(status),
		})
		if err != nil {
			return fmt.Errorf("error updating certificate: %v", err)
		}
	}

	return resourceCertificateRead(d, meta)
}

func resourceCertificateDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).IoTConn

	_, err := conn.UpdateCertificate(&iot.UpdateCertificateInput{
		CertificateId: aws.String(d.Id()),
		NewStatus:     aws.String("INACTIVE"),
	})
	if err != nil {
		return fmt.Errorf("error inactivating certificate: %v", err)
	}

	_, err = conn.DeleteCertificate(&iot.DeleteCertificateInput{
		CertificateId: aws.String(d.Id()),
	})
	if err != nil {
		return fmt.Errorf("error deleting certificate: %v", err)
	}

	return nil
}
