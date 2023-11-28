// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package route53

import (
	"context"
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/route53"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
)

// @SDKResource("aws_route53_hosted_zone_dnssec")
func ResourceHostedZoneDNSSEC() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceHostedZoneDNSSECCreate,
		ReadWithoutTimeout:   resourceHostedZoneDNSSECRead,
		UpdateWithoutTimeout: resourceHostedZoneDNSSECUpdate,
		DeleteWithoutTimeout: resourceHostedZoneDNSSECDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
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

func resourceHostedZoneDNSSECCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).Route53Conn(ctx)

	hostedZoneID := d.Get("hosted_zone_id").(string)
	signingStatus := d.Get("signing_status").(string)

	d.SetId(hostedZoneID)

	switch signingStatus {
	default:
		return sdkdiag.AppendErrorf(diags, "updating Route 53 Hosted Zone DNSSEC (%s) signing status: unknown status (%s)", d.Id(), signingStatus)
	case ServeSignatureSigning:
		if err := hostedZoneDNSSECEnable(ctx, conn, d.Id()); err != nil {
			return sdkdiag.AppendErrorf(diags, "enabling Route 53 Hosted Zone DNSSEC (%s): %s", d.Id(), err)
		}
	case ServeSignatureNotSigning:
		if err := hostedZoneDNSSECDisable(ctx, conn, d.Id()); err != nil {
			return sdkdiag.AppendErrorf(diags, "disabling Route 53 Hosted Zone DNSSEC (%s): %s", d.Id(), err)
		}
	}

	if _, err := waitHostedZoneDNSSECStatusUpdated(ctx, conn, d.Id(), signingStatus); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for Route 53 Hosted Zone DNSSEC (%s) signing status (%s): %s", d.Id(), signingStatus, err)
	}

	return append(diags, resourceHostedZoneDNSSECRead(ctx, d, meta)...)
}

func resourceHostedZoneDNSSECRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).Route53Conn(ctx)

	hostedZoneDnssec, err := FindHostedZoneDNSSEC(ctx, conn, d.Id())

	if !d.IsNewResource() && tfawserr.ErrCodeEquals(err, route53.ErrCodeDNSSECNotFound) {
		log.Printf("[WARN] Route 53 Hosted Zone DNSSEC (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if !d.IsNewResource() && tfawserr.ErrCodeEquals(err, route53.ErrCodeNoSuchHostedZone) {
		log.Printf("[WARN] Route 53 Hosted Zone DNSSEC (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading Route 53 Hosted Zone DNSSEC (%s): %s", d.Id(), err)
	}

	if hostedZoneDnssec == nil {
		if d.IsNewResource() {
			return sdkdiag.AppendErrorf(diags, "reading Route 53 Hosted Zone DNSSEC (%s): not found", d.Id())
		}

		log.Printf("[WARN] Route 53 Hosted Zone DNSSEC (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	d.Set("hosted_zone_id", d.Id())

	if hostedZoneDnssec.Status != nil {
		d.Set("signing_status", hostedZoneDnssec.Status.ServeSignature)
	}

	return diags
}

func resourceHostedZoneDNSSECUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).Route53Conn(ctx)

	if d.HasChange("signing_status") {
		signingStatus := d.Get("signing_status").(string)

		switch signingStatus {
		default:
			return sdkdiag.AppendErrorf(diags, "updating Route 53 Hosted Zone DNSSEC (%s) signing status: unknown status (%s)", d.Id(), signingStatus)
		case ServeSignatureSigning:
			if err := hostedZoneDNSSECEnable(ctx, conn, d.Id()); err != nil {
				return sdkdiag.AppendErrorf(diags, "enabling Route 53 Hosted Zone DNSSEC (%s): %s", d.Id(), err)
			}
		case ServeSignatureNotSigning:
			if err := hostedZoneDNSSECDisable(ctx, conn, d.Id()); err != nil {
				return sdkdiag.AppendErrorf(diags, "disabling Route 53 Hosted Zone DNSSEC (%s): %s", d.Id(), err)
			}
		}

		if _, err := waitHostedZoneDNSSECStatusUpdated(ctx, conn, d.Id(), signingStatus); err != nil {
			return sdkdiag.AppendErrorf(diags, "waiting for Route 53 Hosted Zone DNSSEC (%s) signing status (%s): %s", d.Id(), signingStatus, err)
		}
	}

	return append(diags, resourceHostedZoneDNSSECRead(ctx, d, meta)...)
}

func resourceHostedZoneDNSSECDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).Route53Conn(ctx)

	input := &route53.DisableHostedZoneDNSSECInput{
		HostedZoneId: aws.String(d.Id()),
	}

	output, err := conn.DisableHostedZoneDNSSECWithContext(ctx, input)

	if tfawserr.ErrCodeEquals(err, route53.ErrCodeDNSSECNotFound) {
		return diags
	}

	if tfawserr.ErrCodeEquals(err, route53.ErrCodeNoSuchHostedZone) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "disabling Route 53 Hosted Zone DNSSEC (%s): %s", d.Id(), err)
	}

	if output != nil && output.ChangeInfo != nil {
		if _, err := waitChangeInfoStatusInsync(ctx, conn, aws.StringValue(output.ChangeInfo.Id)); err != nil {
			return sdkdiag.AppendErrorf(diags, "waiting for Route 53 Hosted Zone DNSSEC (%s) disable: %s", d.Id(), err)
		}
	}

	return diags
}

func hostedZoneDNSSECDisable(ctx context.Context, conn *route53.Route53, hostedZoneID string) error {
	input := &route53.DisableHostedZoneDNSSECInput{
		HostedZoneId: aws.String(hostedZoneID),
	}

	output, err := conn.DisableHostedZoneDNSSECWithContext(ctx, input)

	if err != nil {
		return fmt.Errorf("disabling: %w", err)
	}

	if output != nil && output.ChangeInfo != nil {
		if _, err := waitChangeInfoStatusInsync(ctx, conn, aws.StringValue(output.ChangeInfo.Id)); err != nil {
			return fmt.Errorf("waiting for update: %w", err)
		}
	}

	return nil
}

func hostedZoneDNSSECEnable(ctx context.Context, conn *route53.Route53, hostedZoneID string) error {
	input := &route53.EnableHostedZoneDNSSECInput{
		HostedZoneId: aws.String(hostedZoneID),
	}

	output, err := conn.EnableHostedZoneDNSSECWithContext(ctx, input)

	if err != nil {
		return fmt.Errorf("enabling: %w", err)
	}

	if output != nil && output.ChangeInfo != nil {
		if _, err := waitChangeInfoStatusInsync(ctx, conn, aws.StringValue(output.ChangeInfo.Id)); err != nil {
			return fmt.Errorf("waiting for update: %w", err)
		}
	}

	return nil
}
