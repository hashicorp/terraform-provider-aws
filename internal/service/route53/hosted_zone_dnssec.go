// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package route53

import (
	"context"
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/route53"
	awstypes "github.com/aws/aws-sdk-go-v2/service/route53/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

const (
	serveSignatureActionNeeded    = "ACTION_NEEDED"
	serveSignatureDeleting        = "DELETING"
	serveSignatureInternalFailure = "INTERNAL_FAILURE"
	serveSignatureNotSigning      = "NOT_SIGNING"
	serveSignatureSigning         = "SIGNING"
)

// @SDKResource("aws_route53_hosted_zone_dnssec", name="Hosted Zone DNSSEC")
func resourceHostedZoneDNSSEC() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceHostedZoneDNSSECCreate,
		ReadWithoutTimeout:   resourceHostedZoneDNSSECRead,
		UpdateWithoutTimeout: resourceHostedZoneDNSSECUpdate,
		DeleteWithoutTimeout: resourceHostedZoneDNSSECDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			names.AttrHostedZoneID: {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"signing_status": {
				Type:     schema.TypeString,
				Optional: true,
				Default:  serveSignatureSigning,
				ValidateFunc: validation.StringInSlice([]string{
					serveSignatureSigning,
					serveSignatureNotSigning,
				}, false),
			},
		},
	}
}

func resourceHostedZoneDNSSECCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).Route53Client(ctx)

	hostedZoneID := d.Get(names.AttrHostedZoneID).(string)
	signingStatus := d.Get("signing_status").(string)

	d.SetId(hostedZoneID)

	if signingStatus == serveSignatureSigning {
		if err := hostedZoneDNSSECEnable(ctx, conn, d.Id()); err != nil {
			return sdkdiag.AppendFromErr(diags, err)
		}
	} else {
		if err := hostedZoneDNSSECDisable(ctx, conn, d.Id()); err != nil {
			return sdkdiag.AppendFromErr(diags, err)
		}
	}

	if _, err := waitHostedZoneDNSSECStatusUpdated(ctx, conn, d.Id(), signingStatus); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for Route 53 Hosted Zone DNSSEC (%s) signing status update: %s", d.Id(), err)
	}

	return append(diags, resourceHostedZoneDNSSECRead(ctx, d, meta)...)
}

func resourceHostedZoneDNSSECRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).Route53Client(ctx)

	hostedZoneDNSSEC, err := findHostedZoneDNSSECByZoneID(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] Route 53 Hosted Zone DNSSEC (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading Route 53 Hosted Zone DNSSEC (%s): %s", d.Id(), err)
	}

	d.Set(names.AttrHostedZoneID, d.Id())
	d.Set("signing_status", hostedZoneDNSSEC.Status.ServeSignature)

	return diags
}

func resourceHostedZoneDNSSECUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).Route53Client(ctx)

	if d.HasChange("signing_status") {
		signingStatus := d.Get("signing_status").(string)

		if signingStatus == serveSignatureSigning {
			if err := hostedZoneDNSSECEnable(ctx, conn, d.Id()); err != nil {
				return sdkdiag.AppendFromErr(diags, err)
			}
		} else {
			if err := hostedZoneDNSSECDisable(ctx, conn, d.Id()); err != nil {
				return sdkdiag.AppendFromErr(diags, err)
			}
		}

		if _, err := waitHostedZoneDNSSECStatusUpdated(ctx, conn, d.Id(), signingStatus); err != nil {
			return sdkdiag.AppendErrorf(diags, "waiting for Route 53 Hosted Zone DNSSEC (%s) signing status update: %s", d.Id(), err)
		}
	}

	return append(diags, resourceHostedZoneDNSSECRead(ctx, d, meta)...)
}

func resourceHostedZoneDNSSECDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).Route53Client(ctx)

	log.Printf("[DEBUG] Deleting Route 53 Hosted Zone DNSSEC: %s", d.Id())
	output, err := conn.DisableHostedZoneDNSSEC(ctx, &route53.DisableHostedZoneDNSSECInput{
		HostedZoneId: aws.String(d.Id()),
	})

	if errs.IsA[*awstypes.DNSSECNotFound](err) || errs.IsA[*awstypes.NoSuchHostedZone](err) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "disabling Route 53 Hosted Zone DNSSEC (%s): %s", d.Id(), err)
	}

	if output.ChangeInfo != nil {
		if _, err := waitChangeInsync(ctx, conn, aws.ToString(output.ChangeInfo.Id)); err != nil {
			return sdkdiag.AppendErrorf(diags, "waiting for Route 53 Hosted Zone DNSSEC (%s) synchronize: %s", d.Id(), err)
		}
	}

	return diags
}

func hostedZoneDNSSECDisable(ctx context.Context, conn *route53.Client, hostedZoneID string) error {
	input := &route53.DisableHostedZoneDNSSECInput{
		HostedZoneId: aws.String(hostedZoneID),
	}

	const (
		timeout = 5 * time.Minute
	)
	outputRaw, err := tfresource.RetryWhenIsA[*awstypes.KeySigningKeyInParentDSRecord](ctx, timeout, func() (interface{}, error) {
		return conn.DisableHostedZoneDNSSEC(ctx, input)
	})

	if err != nil {
		return fmt.Errorf("disabling Route 53 Hosted Zone DNSSEC (%s): %w", hostedZoneID, err)
	}

	if output := outputRaw.(*route53.DisableHostedZoneDNSSECOutput); output.ChangeInfo != nil {
		if _, err := waitChangeInsync(ctx, conn, aws.ToString(output.ChangeInfo.Id)); err != nil {
			return fmt.Errorf("waiting for Route 53 Hosted Zone DNSSEC (%s) synchronize: %w", hostedZoneID, err)
		}
	}

	return nil
}

func hostedZoneDNSSECEnable(ctx context.Context, conn *route53.Client, hostedZoneID string) error {
	input := &route53.EnableHostedZoneDNSSECInput{
		HostedZoneId: aws.String(hostedZoneID),
	}

	output, err := conn.EnableHostedZoneDNSSEC(ctx, input)

	if err != nil {
		return fmt.Errorf("enabling Route 53 Hosted Zone DNSSEC (%s): %w", hostedZoneID, err)
	}

	if output.ChangeInfo != nil {
		if _, err := waitChangeInsync(ctx, conn, aws.ToString(output.ChangeInfo.Id)); err != nil {
			return fmt.Errorf("waiting for Route 53 Hosted Zone DNSSEC (%s) synchronize: %w", hostedZoneID, err)
		}
	}

	return nil
}

func findHostedZoneDNSSECByZoneID(ctx context.Context, conn *route53.Client, hostedZoneID string) (*route53.GetDNSSECOutput, error) {
	input := &route53.GetDNSSECInput{
		HostedZoneId: aws.String(hostedZoneID),
	}

	output, err := conn.GetDNSSEC(ctx, input)

	if errs.IsA[*awstypes.DNSSECNotFound](err) || errs.IsA[*awstypes.NoSuchHostedZone](err) {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil || output.Status == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	return output, nil
}

func statusHostedZoneDNSSEC(ctx context.Context, conn *route53.Client, hostedZoneID string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		output, err := findHostedZoneDNSSECByZoneID(ctx, conn, hostedZoneID)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output.Status, aws.ToString(output.Status.ServeSignature), nil
	}
}

func waitHostedZoneDNSSECStatusUpdated(ctx context.Context, conn *route53.Client, hostedZoneID, status string) (*awstypes.DNSSECStatus, error) { //nolint:unparam
	const (
		timeout = 5 * time.Minute
	)
	stateConf := &retry.StateChangeConf{
		Target:     []string{status},
		Refresh:    statusHostedZoneDNSSEC(ctx, conn, hostedZoneID),
		MinTimeout: 5 * time.Second,
		Timeout:    timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*awstypes.DNSSECStatus); ok {
		if serveSignature := aws.ToString(output.ServeSignature); serveSignature == serveSignatureInternalFailure {
			tfresource.SetLastError(err, errors.New(aws.ToString(output.StatusMessage)))
		}

		return output, err
	}

	return nil, err
}
