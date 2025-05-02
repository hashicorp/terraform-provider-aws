// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package rds

import (
	"context"
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/rds"
	"github.com/aws/aws-sdk-go-v2/service/rds/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	tfslices "github.com/hashicorp/terraform-provider-aws/internal/slices"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_rds_certificate", name="Default Certificate")
// @SingletonIdentity
func resourceCertificate() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceCertificatePut,
		ReadWithoutTimeout:   resourceCertificateRead,
		UpdateWithoutTimeout: resourceCertificatePut,
		DeleteWithoutTimeout: resourceCertificateDelete,

		Importer: &schema.ResourceImporter{
			StateContext: func(ctx context.Context, rd *schema.ResourceData, meta any) ([]*schema.ResourceData, error) {
				if rd.Id() != "" {
					rd.Set("region", rd.Id())
					return []*schema.ResourceData{rd}, nil
				}

				identity, err := rd.Identity()
				if err != nil {
					return nil, err
				}

				accountIDRaw, ok := identity.GetOk(names.AttrAccountID)
				var accountID string
				if ok {
					accountID, ok = accountIDRaw.(string)
					if !ok {
						return nil, fmt.Errorf("identity attribute %q: expected string, got %T", names.AttrAccountID, accountIDRaw)
					}
					client := meta.(*conns.AWSClient)
					if accountID != client.AccountID(ctx) {
						return nil, fmt.Errorf("Unable to import\n\nidentity attribute %q: Provider configured with Account ID %q, got %q", names.AttrAccountID, client.AccountID(ctx), accountID)
					}
				}

				regionRaw, ok := identity.GetOk("region")
				var region string
				if ok {
					region, ok = regionRaw.(string)
					if !ok {
						return nil, fmt.Errorf("identity attribute %q: expected string, got %T", "region", regionRaw)
					}
				} else {
					client := meta.(*conns.AWSClient)
					region = client.Region(ctx)
				}
				rd.Set("region", region)
				rd.SetId(region)

				return []*schema.ResourceData{rd}, nil
			},
		},

		Schema: map[string]*schema.Schema{
			"certificate_identifier": {
				Type:     schema.TypeString,
				Required: true,
			},
		},
	}
}

func resourceCertificatePut(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).RDSClient(ctx)

	certificateID := d.Get("certificate_identifier").(string)
	input := &rds.ModifyCertificatesInput{
		CertificateIdentifier: aws.String(certificateID),
	}

	_, err := conn.ModifyCertificates(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "setting RDS Default Certificate (%s): %s", certificateID, err)
	}

	if d.IsNewResource() {
		d.SetId(meta.(*conns.AWSClient).Region(ctx))
	}

	return append(diags, resourceCertificateRead(ctx, d, meta)...)
}

func resourceCertificateRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).RDSClient(ctx)

	output, err := findDefaultCertificate(ctx, conn)

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] RDS Default Certificate (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading RDS Default Certificate (%s): %s", d.Id(), err)
	}

	d.Set("certificate_identifier", output.CertificateIdentifier)

	return diags
}

func resourceCertificateDelete(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).RDSClient(ctx)

	log.Printf("[DEBUG] Deleting RDS Default Certificate: %s", d.Id())
	_, err := conn.ModifyCertificates(ctx, &rds.ModifyCertificatesInput{
		RemoveCustomerOverride: aws.Bool(true),
	})

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "removing RDS Default Certificate (%s): %s", d.Id(), err)
	}

	return diags
}

func findCertificate(ctx context.Context, conn *rds.Client, input *rds.DescribeCertificatesInput, filter tfslices.Predicate[*types.Certificate]) (*types.Certificate, error) {
	output, err := findCertificates(ctx, conn, input, filter)

	if err != nil {
		return nil, err
	}

	return tfresource.AssertSingleValueResult(output)
}

func findCertificates(ctx context.Context, conn *rds.Client, input *rds.DescribeCertificatesInput, filter tfslices.Predicate[*types.Certificate]) ([]types.Certificate, error) {
	var output []types.Certificate

	pages := rds.NewDescribeCertificatesPaginator(conn, input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if errs.IsA[*types.CertificateNotFoundFault](err) {
			return nil, &retry.NotFoundError{
				LastError:   err,
				LastRequest: input,
			}
		}

		if err != nil {
			return nil, err
		}

		for _, v := range page.Certificates {
			if filter(&v) {
				output = append(output, v)
			}
		}
	}

	return output, nil
}

func findDefaultCertificate(ctx context.Context, conn *rds.Client) (*types.Certificate, error) {
	input := &rds.DescribeCertificatesInput{}

	return findCertificate(ctx, conn, input, func(v *types.Certificate) bool {
		return aws.ToBool(v.CustomerOverride)
	})
}
