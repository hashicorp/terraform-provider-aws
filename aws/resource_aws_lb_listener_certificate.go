package aws

import (
	"errors"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/elbv2"
	"github.com/hashicorp/errwrap"
	"github.com/hashicorp/terraform/helper/schema"
)

func resourceAwsLbListenerCertificate() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsLbListenerCertificateCreate,
		Read:   resourceAwsLbListenerCertificateRead,
		Delete: resourceAwsLbListenerCertificateDelete,

		Schema: map[string]*schema.Schema{
			"listener_arn": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"certificate_arn": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
		},
	}
}

func resourceAwsLbListenerCertificateCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).elbv2conn

	params := &elbv2.AddListenerCertificatesInput{
		ListenerArn: aws.String(d.Get("listener_arn").(string)),
		Certificates: []*elbv2.Certificate{
			&elbv2.Certificate{
				CertificateArn: aws.String(d.Get("certificate_arn").(string)),
			},
		},
	}

	log.Printf("[DEBUG] Adding certificate: %s of listener: %s", d.Get("certificate_arn").(string), d.Get("listener_arn").(string))
	resp, err := conn.AddListenerCertificates(params)
	if err != nil {
		return errwrap.Wrapf("Error creating LB Listener Certificate: {{err}}", err)
	}

	if len(resp.Certificates) == 0 {
		return errors.New("Error creating LB Listener Certificate: no certificates returned in response")
	}

	d.SetId(d.Get("listener_arn").(string) + "_" + d.Get("certificate_arn").(string))

	return resourceAwsLbListenerCertificateRead(d, meta)
}

func resourceAwsLbListenerCertificateRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).elbv2conn
	log.Printf("[DEBUG] Reading certificate: %s of listener: %s", d.Get("certificate_arn").(string), d.Get("listener_arn").(string))

	params := &elbv2.DescribeListenerCertificatesInput{
		ListenerArn: aws.String(d.Get("listener_arn").(string)),
		PageSize:    aws.Int64(400),
	}

	morePages := true
	found := false
	for morePages && !found {
		resp, err := conn.DescribeListenerCertificates(params)
		if err != nil {
			return err
		}

		for _, cert := range resp.Certificates {
			// We don't care about the default certificate.
			if *cert.IsDefault {
				continue
			}

			if *cert.CertificateArn == d.Get("certificate_arn").(string) {
				found = true
			}
		}

		if resp.NextMarker != nil {
			params.Marker = resp.NextMarker
		} else {
			morePages = false
		}
	}

	if !found {
		log.Printf("[WARN] DescribeListenerCertificates - removing %s from state", d.Id())
		d.SetId("")
		return nil
	}

	return nil
}

func resourceAwsLbListenerCertificateDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).elbv2conn
	log.Printf("[DEBUG] Deleting certificate: %s of listener: %s", d.Get("certificate_arn").(string), d.Get("listener_arn").(string))

	params := &elbv2.RemoveListenerCertificatesInput{
		ListenerArn: aws.String(d.Get("listener_arn").(string)),
		Certificates: []*elbv2.Certificate{
			&elbv2.Certificate{
				CertificateArn: aws.String(d.Get("certificate_arn").(string)),
			},
		},
	}

	_, err := conn.RemoveListenerCertificates(params)
	if err != nil {
		return errwrap.Wrapf("Error removing LB Listener Certificate: {{err}}", err)
	}

	return nil
}
