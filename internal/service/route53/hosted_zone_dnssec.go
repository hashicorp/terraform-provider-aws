package route53

import (
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/route53"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
)

func ResourceHostedZoneDNSSEC() *schema.Resource {
	return &schema.Resource{
		Create: resourceHostedZoneDNSSECCreate,
		Read:   resourceHostedZoneDNSSECRead,
		Update: resourceHostedZoneDNSSECUpdate,
		Delete: resourceHostedZoneDNSSECDelete,

		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"hosted_zone_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"signing_status": {
				Type:     schema.TypeString,
				Optional: true,
				Default:  ServeSignatureSigning,
				ValidateFunc: validation.StringInSlice([]string{
					ServeSignatureSigning,
					ServeSignatureNotSigning,
				}, false),
			},
		},
	}
}

func resourceHostedZoneDNSSECCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).Route53Conn

	hostedZoneID := d.Get("hosted_zone_id").(string)
	signingStatus := d.Get("signing_status").(string)

	d.SetId(hostedZoneID)

	switch signingStatus {
	default:
		return fmt.Errorf("error updating Route 53 Hosted Zone DNSSEC (%s) signing status: unknown status (%s)", d.Id(), signingStatus)
	case ServeSignatureSigning:
		if err := hostedZoneDNSSECEnable(conn, d.Id()); err != nil {
			return fmt.Errorf("error enabling Route 53 Hosted Zone DNSSEC (%s): %w", d.Id(), err)
		}
	case ServeSignatureNotSigning:
		if err := hostedZoneDNSSECDisable(conn, d.Id()); err != nil {
			return fmt.Errorf("error disabling Route 53 Hosted Zone DNSSEC (%s): %w", d.Id(), err)
		}
	}

	if _, err := waitHostedZoneDNSSECStatusUpdated(conn, d.Id(), signingStatus); err != nil {
		return fmt.Errorf("error waiting for Route 53 Hosted Zone DNSSEC (%s) signing status (%s): %w", d.Id(), signingStatus, err)
	}

	return resourceHostedZoneDNSSECRead(d, meta)
}

func resourceHostedZoneDNSSECRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).Route53Conn

	hostedZoneDnssec, err := FindHostedZoneDNSSEC(conn, d.Id())

	if !d.IsNewResource() && tfawserr.ErrCodeEquals(err, route53.ErrCodeDNSSECNotFound) {
		log.Printf("[WARN] Route 53 Hosted Zone DNSSEC (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if !d.IsNewResource() && tfawserr.ErrCodeEquals(err, route53.ErrCodeNoSuchHostedZone) {
		log.Printf("[WARN] Route 53 Hosted Zone DNSSEC (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return fmt.Errorf("error reading Route 53 Hosted Zone DNSSEC (%s): %w", d.Id(), err)
	}

	if hostedZoneDnssec == nil {
		if d.IsNewResource() {
			return fmt.Errorf("error reading Route 53 Hosted Zone DNSSEC (%s): not found", d.Id())
		}

		log.Printf("[WARN] Route 53 Hosted Zone DNSSEC (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	d.Set("hosted_zone_id", d.Id())

	if hostedZoneDnssec.Status != nil {
		d.Set("signing_status", hostedZoneDnssec.Status.ServeSignature)
	}

	return nil
}

func resourceHostedZoneDNSSECUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).Route53Conn

	if d.HasChange("signing_status") {
		signingStatus := d.Get("signing_status").(string)

		switch signingStatus {
		default:
			return fmt.Errorf("error updating Route 53 Hosted Zone DNSSEC (%s) signing status: unknown status (%s)", d.Id(), signingStatus)
		case ServeSignatureSigning:
			if err := hostedZoneDNSSECEnable(conn, d.Id()); err != nil {
				return fmt.Errorf("error enabling Route 53 Hosted Zone DNSSEC (%s): %w", d.Id(), err)
			}
		case ServeSignatureNotSigning:
			if err := hostedZoneDNSSECDisable(conn, d.Id()); err != nil {
				return fmt.Errorf("error disabling Route 53 Hosted Zone DNSSEC (%s): %w", d.Id(), err)
			}
		}

		if _, err := waitHostedZoneDNSSECStatusUpdated(conn, d.Id(), signingStatus); err != nil {
			return fmt.Errorf("error waiting for Route 53 Hosted Zone DNSSEC (%s) signing status (%s): %w", d.Id(), signingStatus, err)
		}
	}

	return resourceHostedZoneDNSSECRead(d, meta)
}

func resourceHostedZoneDNSSECDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).Route53Conn

	input := &route53.DisableHostedZoneDNSSECInput{
		HostedZoneId: aws.String(d.Id()),
	}

	output, err := conn.DisableHostedZoneDNSSEC(input)

	if tfawserr.ErrCodeEquals(err, route53.ErrCodeDNSSECNotFound) {
		return nil
	}

	if tfawserr.ErrCodeEquals(err, route53.ErrCodeNoSuchHostedZone) {
		return nil
	}

	if err != nil {
		return fmt.Errorf("error disabling Route 53 Hosted Zone DNSSEC (%s): %w", d.Id(), err)
	}

	if output != nil && output.ChangeInfo != nil {
		if _, err := waitChangeInfoStatusInsync(conn, aws.StringValue(output.ChangeInfo.Id)); err != nil {
			return fmt.Errorf("error waiting for Route 53 Hosted Zone DNSSEC (%s) disable: %w", d.Id(), err)
		}
	}

	return nil
}

func hostedZoneDNSSECDisable(conn *route53.Route53, hostedZoneID string) error {
	input := &route53.DisableHostedZoneDNSSECInput{
		HostedZoneId: aws.String(hostedZoneID),
	}

	output, err := conn.DisableHostedZoneDNSSEC(input)

	if err != nil {
		return fmt.Errorf("error disabling: %w", err)
	}

	if output != nil && output.ChangeInfo != nil {
		if _, err := waitChangeInfoStatusInsync(conn, aws.StringValue(output.ChangeInfo.Id)); err != nil {
			return fmt.Errorf("error waiting for update: %w", err)
		}
	}

	return nil
}

func hostedZoneDNSSECEnable(conn *route53.Route53, hostedZoneID string) error {
	input := &route53.EnableHostedZoneDNSSECInput{
		HostedZoneId: aws.String(hostedZoneID),
	}

	output, err := conn.EnableHostedZoneDNSSEC(input)

	if err != nil {
		return fmt.Errorf("error enabling: %w", err)
	}

	if output != nil && output.ChangeInfo != nil {
		if _, err := waitChangeInfoStatusInsync(conn, aws.StringValue(output.ChangeInfo.Id)); err != nil {
			return fmt.Errorf("error waiting for update: %w", err)
		}
	}

	return nil
}
