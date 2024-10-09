// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package elbv2

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/elasticloadbalancingv2"
	awstypes "github.com/aws/aws-sdk-go-v2/service/elasticloadbalancingv2/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	tfslices "github.com/hashicorp/terraform-provider-aws/internal/slices"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_alb_listener_certificate", name="Listener Certificate")
// @SDKResource("aws_lb_listener_certificate", name="Listener Certificate")
func resourceListenerCertificate() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceListenerCertificateCreate,
		ReadWithoutTimeout:   resourceListenerCertificateRead,
		DeleteWithoutTimeout: resourceListenerCertificateDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			names.AttrCertificateARN: {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: verify.ValidARN,
			},
			"listener_arn": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: verify.ValidARN,
			},
		},
	}
}

func resourceListenerCertificateCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ELBV2Client(ctx)

	listenerARN := d.Get("listener_arn").(string)
	certificateARN := d.Get(names.AttrCertificateARN).(string)
	id := listenerCertificateCreateResourceID(listenerARN, certificateARN)
	input := &elasticloadbalancingv2.AddListenerCertificatesInput{
		Certificates: []awstypes.Certificate{{
			CertificateArn: aws.String(certificateARN),
		}},
		ListenerArn: aws.String(listenerARN),
	}

	_, err := tfresource.RetryWhenIsA[*awstypes.CertificateNotFoundException](ctx, iamPropagationTimeout, func() (interface{}, error) {
		return conn.AddListenerCertificates(ctx, input)
	})

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating ELBv2 Listener Certificate (%s): %s", id, err)
	}

	d.SetId(id)

	return append(diags, resourceListenerCertificateRead(ctx, d, meta)...)
}

func resourceListenerCertificateRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ELBV2Client(ctx)

	listenerARN, certificateARN, err := listenerCertificateParseResourceID(d.Id())
	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	_, err = tfresource.RetryWhenNewResourceNotFound(ctx, elbv2PropagationTimeout, func() (interface{}, error) {
		return findListenerCertificateByTwoPartKey(ctx, conn, listenerARN, certificateARN)
	}, d.IsNewResource())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] ELBv2 Listener Certificate (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading ELBv2 Listener Certificate (%s): %s", d.Id(), err)
	}

	d.Set(names.AttrCertificateARN, certificateARN)
	d.Set("listener_arn", listenerARN)

	return diags
}

func resourceListenerCertificateDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ELBV2Client(ctx)

	listenerARN, certificateARN, err := listenerCertificateParseResourceID(d.Id())
	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	log.Printf("[INFO] Deleting ELBv2 Listener Certificate: %s", d.Id())
	_, err = conn.RemoveListenerCertificates(ctx, &elasticloadbalancingv2.RemoveListenerCertificatesInput{
		Certificates: []awstypes.Certificate{{
			CertificateArn: aws.String(certificateARN),
		}},
		ListenerArn: aws.String(listenerARN),
	})

	if errs.IsA[*awstypes.CertificateNotFoundException](err) || errs.IsA[*awstypes.ListenerNotFoundException](err) {
		return diags
	}

	// Even though we're not trying to remove the default certificate, AWS started returning this error around 2023-12-09.
	if errs.IsAErrorMessageContains[*awstypes.OperationNotPermittedException](err, "Default certificate cannot be removed") {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting ELBv2 Listener Certificate (%s): %s", d.Id(), err)
	}

	return diags
}

const listenerCertificateResourceIDSeparator = "_"

func listenerCertificateCreateResourceID(listenerARN, certificateARN string) string {
	parts := []string{listenerARN, certificateARN}
	id := strings.Join(parts, listenerCertificateResourceIDSeparator)

	return id
}

func listenerCertificateParseResourceID(id string) (string, string, error) {
	parts := strings.SplitN(id, listenerCertificateResourceIDSeparator, 2)

	if len(parts) == 2 && parts[0] != "" && parts[1] != "" {
		return parts[0], parts[1], nil
	}

	return "", "", fmt.Errorf("unexpected format for ID (%[1]s), expected LISTENER_ARN%[2]sCERTIFICATE_ARN", id, listenerCertificateResourceIDSeparator)
}

func findListenerCertificate(ctx context.Context, conn *elasticloadbalancingv2.Client, input *elasticloadbalancingv2.DescribeListenerCertificatesInput, filter tfslices.Predicate[*awstypes.Certificate]) (*awstypes.Certificate, error) {
	output, err := findListenerCertificates(ctx, conn, input, filter)

	if err != nil {
		return nil, err
	}

	return tfresource.AssertSingleValueResult(output)
}

func findListenerCertificates(ctx context.Context, conn *elasticloadbalancingv2.Client, input *elasticloadbalancingv2.DescribeListenerCertificatesInput, filter tfslices.Predicate[*awstypes.Certificate]) ([]awstypes.Certificate, error) {
	var output []awstypes.Certificate

	err := describeListenerCertificatesPages(ctx, conn, input, func(page *elasticloadbalancingv2.DescribeListenerCertificatesOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, v := range page.Certificates {
			if filter(&v) {
				output = append(output, v)
			}
		}

		return !lastPage
	})

	if errs.IsA[*awstypes.ListenerNotFoundException](err) {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	return output, nil
}

func findListenerCertificateByTwoPartKey(ctx context.Context, conn *elasticloadbalancingv2.Client, listenerARN, certificateARN string) (*awstypes.Certificate, error) {
	input := &elasticloadbalancingv2.DescribeListenerCertificatesInput{
		ListenerArn: aws.String(listenerARN),
		PageSize:    aws.Int32(400),
	}

	return findListenerCertificate(ctx, conn, input, func(v *awstypes.Certificate) bool {
		return !aws.ToBool(v.IsDefault) && aws.ToString(v.CertificateArn) == certificateARN
	})
}
