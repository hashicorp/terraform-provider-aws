// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package acm

import (
	"context"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/acm"
	"github.com/aws/aws-sdk-go-v2/service/acm/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKDataSource("aws_acm_certificate", name="Certificate")
// @Tags(identifierAttribute="arn")
// @Testing(tlsKey=true, generator=false)
func dataSourceCertificate() *schema.Resource {
	return &schema.Resource{
		ReadWithoutTimeout: dataSourceCertificateRead,

		Schema: map[string]*schema.Schema{
			names.AttrARN: {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrCertificate: {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrCertificateChain: {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrDomain: {
				Type:     schema.TypeString,
				Required: true,
			},
			"key_types": {
				Type:     schema.TypeSet,
				Optional: true,
				Elem: &schema.Schema{
					Type:             schema.TypeString,
					ValidateDiagFunc: enum.Validate[types.KeyAlgorithm](),
				},
			},
			names.AttrMostRecent: {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},
			names.AttrStatus: {
				Type:     schema.TypeString,
				Computed: true,
			},
			"statuses": {
				Type:     schema.TypeList,
				Optional: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			names.AttrTags: tftags.TagsSchemaComputed(),
			"types": {
				Type:     schema.TypeList,
				Optional: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
		},
	}
}

func dataSourceCertificateRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	conn := meta.(*conns.AWSClient).ACMClient(ctx)
	ignoreTagsConfig := meta.(*conns.AWSClient).IgnoreTagsConfig

	domain := d.Get(names.AttrDomain).(string)
	input := acm.ListCertificatesInput{}

	if v, ok := d.GetOk("key_types"); ok && v.(*schema.Set).Len() > 0 {
		input.Includes = &types.Filters{
			KeyTypes: flex.ExpandStringyValueSet[types.KeyAlgorithm](v.(*schema.Set)),
		}
	}

	if v, ok := d.GetOk("statuses"); ok && len(v.([]interface{})) > 0 {
		input.CertificateStatuses = flex.ExpandStringyValueList[types.CertificateStatus](v.([]interface{}))
	} else {
		input.CertificateStatuses = []types.CertificateStatus{types.CertificateStatusIssued}
	}

	arns, err := tfresource.RetryGWhenNotFound(ctx, 1*time.Minute,
		func() ([]string, error) {
			return listCertificates(ctx, conn, &input, domain)
		},
	)
	if tfresource.NotFound(err) {
		sdkdiag.AppendErrorf(diags, "XXX no ACM Certificate matching domain (%s)", domain)
	} else if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading ACM Certificates: %s", err)
	}

	filterMostRecent := d.Get(names.AttrMostRecent).(bool)
	certificateTypes := flex.ExpandStringyValueList[types.CertificateType](d.Get("types").([]interface{}))

	if !filterMostRecent && len(certificateTypes) == 0 && len(arns) > 1 {
		return sdkdiag.AppendErrorf(diags, "multiple ACM Certificates matching domain (%s)", domain)
	}

	var matchedCertificate *types.CertificateDetail

	for _, arn := range arns {
		input := &acm.DescribeCertificateInput{
			CertificateArn: aws.String(arn),
		}

		certificate, err := findCertificate(ctx, conn, input)

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "reading ACM Certificate (%s): %s", arn, err)
		}

		if len(certificateTypes) > 0 {
			for _, certificateType := range certificateTypes {
				if certificate.Type == certificateType {
					// We do not have a candidate certificate.
					if matchedCertificate == nil {
						matchedCertificate = certificate

						break
					}

					// At this point, we already have a candidate certificate.
					// Check if we are filtering by most recent and update if necessary.
					if filterMostRecent {
						matchedCertificate, err = mostRecentCertificate(certificate, matchedCertificate)

						if err != nil {
							return sdkdiag.AppendFromErr(diags, err)
						}

						break
					}
					// Now we have multiple candidate certificates and we only allow one certificate.
					return sdkdiag.AppendErrorf(diags, "multiple ACM Certificates matching domain (%s)", domain)
				}
			}

			continue
		}

		// We do not have a candidate certificate.
		if matchedCertificate == nil {
			matchedCertificate = certificate

			continue
		}

		// At this point, we already have a candidate certificate.
		// Check if we are filtering by most recent and update if necessary.
		if filterMostRecent {
			matchedCertificate, err = mostRecentCertificate(certificate, matchedCertificate)

			if err != nil {
				return sdkdiag.AppendFromErr(diags, err)
			}

			continue
		}

		// Now we have multiple candidate certificates and we only allow one certificate.
		return sdkdiag.AppendErrorf(diags, "multiple ACM Certificates matching domain (%s)", domain)
	}

	if matchedCertificate == nil {
		return sdkdiag.AppendErrorf(diags, "YYY no ACM Certificate matching domain (%s)", domain)
	}

	// Get the certificate data if the status is issued
	var output *acm.GetCertificateOutput
	if matchedCertificate.Status == types.CertificateStatusIssued {
		arn := aws.ToString(matchedCertificate.CertificateArn)
		input := &acm.GetCertificateInput{
			CertificateArn: aws.String(arn),
		}
		var err error

		output, err = conn.GetCertificate(ctx, input)

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "reading ACM Certificate (%s): %s", arn, err)
		}
	}
	if output != nil {
		d.Set(names.AttrCertificate, output.Certificate)
		d.Set(names.AttrCertificateChain, output.CertificateChain)
	} else {
		d.Set(names.AttrCertificate, nil)
		d.Set(names.AttrCertificateChain, nil)
	}

	d.SetId(aws.ToString(matchedCertificate.CertificateArn))
	d.Set(names.AttrARN, matchedCertificate.CertificateArn)
	d.Set(names.AttrStatus, matchedCertificate.Status)

	tags, err := listTags(ctx, conn, aws.ToString(matchedCertificate.CertificateArn))

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "listing tags for ACM Certificate (%s): %s", d.Id(), err)
	}

	if err := d.Set(names.AttrTags, tags.IgnoreAWS().IgnoreConfig(ignoreTagsConfig).Map()); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting tags: %s", err)
	}

	return diags
}

func mostRecentCertificate(i, j *types.CertificateDetail) (*types.CertificateDetail, error) {
	if i.Status != j.Status {
		return nil, fmt.Errorf("most_recent filtering on different ACM certificate statues is not supported")
	}
	// Cover IMPORTED and ISSUED AMAZON_ISSUED certificates
	if i.Status == types.CertificateStatusIssued {
		if aws.ToTime(i.NotBefore).After(aws.ToTime(j.NotBefore)) {
			return i, nil
		}
		return j, nil
	}
	// Cover non-ISSUED AMAZON_ISSUED certificates
	if aws.ToTime(i.CreatedAt).After(aws.ToTime(j.CreatedAt)) {
		return i, nil
	}
	return j, nil
}

func listCertificates(ctx context.Context, conn *acm.Client, input *acm.ListCertificatesInput, domain string) ([]string, error) {
	var result []string

	pages := acm.NewListCertificatesPaginator(conn, input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if err != nil {
			return []string{}, err
		}

		for _, v := range page.CertificateSummaryList {
			if aws.ToString(v.DomainName) == domain {
				result = append(result, aws.ToString(v.CertificateArn))
			}
		}
	}

	if len(result) == 0 {
		return []string{}, tfresource.NewEmptyResultError(input)
	}

	return result, nil
}
