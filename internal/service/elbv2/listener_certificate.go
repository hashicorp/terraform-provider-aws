// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package elbv2

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/elbv2"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_alb_listener_certificate")
// @SDKResource("aws_lb_listener_certificate")
func ResourceListenerCertificate() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceListenerCertificateCreate,
		ReadWithoutTimeout:   resourceListenerCertificateRead,
		DeleteWithoutTimeout: resourceListenerCertificateDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			"listener_arn": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: verify.ValidARN,
			},
			"certificate_arn": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: verify.ValidARN,
			},
		},
	}
}

const (
	ResNameListenerCertificate  = "Listener Certificate"
	ListenerCertificateNotFound = "ListenerCertificateNotFound"
)

func resourceListenerCertificateCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ELBV2Conn(ctx)

	listenerArn := d.Get("listener_arn").(string)
	certificateArn := d.Get("certificate_arn").(string)

	params := &elbv2.AddListenerCertificatesInput{
		ListenerArn: aws.String(listenerArn),
		Certificates: []*elbv2.Certificate{{
			CertificateArn: aws.String(certificateArn),
		}},
	}

	log.Printf("[DEBUG] Adding certificate: %s of listener: %s", certificateArn, listenerArn)

	err := retry.RetryContext(ctx, 1*time.Minute, func() *retry.RetryError {
		_, err := conn.AddListenerCertificatesWithContext(ctx, params)

		// Retry for IAM Server Certificate eventual consistency
		if tfawserr.ErrCodeEquals(err, elbv2.ErrCodeCertificateNotFoundException) {
			return retry.RetryableError(err)
		}

		if err != nil {
			return retry.NonRetryableError(err)
		}

		return nil
	})

	if tfresource.TimedOut(err) {
		_, err = conn.AddListenerCertificatesWithContext(ctx, params)
	}

	if err != nil {
		return create.DiagError(names.ELBV2, create.ErrActionCreating, ResNameListenerCertificate, d.Id(), err)
	}

	d.SetId(listenerCertificateCreateID(listenerArn, certificateArn))

	return append(diags, resourceListenerCertificateRead(ctx, d, meta)...)
}

func resourceListenerCertificateRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ELBV2Conn(ctx)

	listenerArn, certificateArn, err := listenerCertificateParseID(d.Id())
	if err != nil {
		return create.DiagError(names.ELBV2, create.ErrActionReading, ResNameListenerCertificate, d.Id(), err)
	}

	log.Printf("[DEBUG] Reading certificate: %s of listener: %s", certificateArn, listenerArn)

	err = retry.RetryContext(ctx, 1*time.Minute, func() *retry.RetryError {
		err := findListenerCertificate(ctx, conn, certificateArn, listenerArn, true, nil)
		if tfresource.NotFound(err) && d.IsNewResource() {
			return retry.RetryableError(err)
		}

		if err != nil {
			return retry.NonRetryableError(err)
		}

		return nil
	})

	if tfresource.TimedOut(err) {
		err = findListenerCertificate(ctx, conn, certificateArn, listenerArn, true, nil)
	}

	if !d.IsNewResource() && tfresource.NotFound(err) {
		create.LogNotFoundRemoveState(names.ELBV2, create.ErrActionReading, ResNameListenerCertificate, d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return create.DiagError(names.ELBV2, create.ErrActionReading, ResNameListenerCertificate, d.Id(), err)
	}

	d.Set("certificate_arn", certificateArn)
	d.Set("listener_arn", listenerArn)

	return diags
}

func resourceListenerCertificateDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ELBV2Conn(ctx)

	certificateArn := d.Get("certificate_arn").(string)
	listenerArn := d.Get("listener_arn").(string)

	log.Printf("[DEBUG] Deleting certificate: %s of listener: %s", certificateArn, listenerArn)

	params := &elbv2.RemoveListenerCertificatesInput{
		ListenerArn: aws.String(listenerArn),
		Certificates: []*elbv2.Certificate{{
			CertificateArn: aws.String(certificateArn),
		}},
	}

	_, err := conn.RemoveListenerCertificatesWithContext(ctx, params)
	if err != nil {
		if tfawserr.ErrCodeEquals(err, elbv2.ErrCodeCertificateNotFoundException) {
			return diags
		}
		if tfawserr.ErrCodeEquals(err, elbv2.ErrCodeListenerNotFoundException) {
			return diags
		}

		return create.DiagError(names.ELBV2, create.ErrActionDeleting, ResNameListenerCertificate, d.Id(), err)
	}

	return diags
}

func findListenerCertificate(ctx context.Context, conn *elbv2.ELBV2, certificateArn, listenerArn string, skipDefault bool, nextMarker *string) error {
	params := &elbv2.DescribeListenerCertificatesInput{
		ListenerArn: aws.String(listenerArn),
		PageSize:    aws.Int64(400),
	}
	if nextMarker != nil {
		params.Marker = nextMarker
	}

	resp, err := conn.DescribeListenerCertificatesWithContext(ctx, params)
	if err != nil {
		return err
	}

	for _, cert := range resp.Certificates {
		if skipDefault && aws.BoolValue(cert.IsDefault) {
			continue
		}

		if aws.StringValue(cert.CertificateArn) == certificateArn {
			return nil
		}
	}

	if resp.NextMarker != nil {
		return findListenerCertificate(ctx, conn, certificateArn, listenerArn, skipDefault, resp.NextMarker)
	}

	return &retry.NotFoundError{
		LastRequest: params,
		Message:     fmt.Sprintf("%s: certificate %s for listener %s not found", ListenerCertificateNotFound, certificateArn, listenerArn),
	}
}
