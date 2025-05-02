// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package acm

import (
	"context"
	"fmt"
	"slices"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/acm"
	awstypes "github.com/aws/aws-sdk-go-v2/service/acm/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	tfslices "github.com/hashicorp/terraform-provider-aws/internal/slices"
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
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				AtLeastOneOf: []string{names.AttrDomain, names.AttrTags},
			},
			"key_types": {
				Type:     schema.TypeSet,
				Optional: true,
				Elem: &schema.Schema{
					Type:             schema.TypeString,
					ValidateDiagFunc: enum.Validate[awstypes.KeyAlgorithm](),
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
			names.AttrTags: {
				Type:         schema.TypeMap,
				Optional:     true,
				Computed:     true,
				Elem:         &schema.Schema{Type: schema.TypeString},
				AtLeastOneOf: []string{names.AttrDomain, names.AttrTags},
			},
			"types": {
				Type:     schema.TypeList,
				Optional: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
		},
	}
}

func dataSourceCertificateRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ACMClient(ctx)

	input := acm.ListCertificatesInput{}

	if v, ok := d.GetOk("key_types"); ok && v.(*schema.Set).Len() > 0 {
		filters := awstypes.Filters{
			KeyTypes: flex.ExpandStringyValueSet[awstypes.KeyAlgorithm](v.(*schema.Set)),
		}
		input.Includes = &filters
	}

	if v, ok := d.GetOk("statuses"); ok && len(v.([]any)) > 0 {
		input.CertificateStatuses = flex.ExpandStringyValueList[awstypes.CertificateStatus](v.([]any))
	} else {
		input.CertificateStatuses = []awstypes.CertificateStatus{awstypes.CertificateStatusIssued}
	}

	f := tfslices.PredicateTrue[*awstypes.CertificateSummary]()
	if domain, ok := d.GetOk(names.AttrDomain); ok {
		f = func(v *awstypes.CertificateSummary) bool {
			return aws.ToString(v.DomainName) == domain
		}
	}
	if certificateTypes := flex.ExpandStringyValueList[awstypes.CertificateType](d.Get("types").([]any)); len(certificateTypes) > 0 {
		f = tfslices.PredicateAnd(f, func(v *awstypes.CertificateSummary) bool {
			return slices.Contains(certificateTypes, v.Type)
		})
	}

	const (
		timeout = 1 * time.Minute
	)
	certificateSummaries, err := tfresource.RetryGWhenNotFound(ctx, timeout,
		func() ([]awstypes.CertificateSummary, error) {
			output, err := findCertificates(ctx, conn, &input, f)
			switch {
			case err != nil:
				return nil, err
			case len(output) == 0:
				return nil, tfresource.NewEmptyResultError(input)
			default:
				return output, nil
			}
		},
	)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading ACM Certificates: %s", err)
	}

	var certificates []*awstypes.CertificateDetail
	for _, certificateSummary := range certificateSummaries {
		certificateARN := aws.ToString(certificateSummary.CertificateArn)
		certificate, err := findCertificateByARN(ctx, conn, certificateARN)

		if tfresource.NotFound(err) {
			continue
		}

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "reading ACM Certificate (%s): %s", certificateARN, err)
		}

		if tagsToMatch := getTagsIn(ctx); len(tagsToMatch) > 0 {
			tags, err := listTags(ctx, conn, certificateARN)

			if errs.IsA[*awstypes.ResourceNotFoundException](err) {
				continue
			}

			if err != nil {
				return sdkdiag.AppendErrorf(diags, "listing tags for ACM Certificate (%s): %s", certificateARN, err)
			}

			if !tags.ContainsAll(keyValueTags(ctx, tagsToMatch)) {
				continue
			}
		}

		certificates = append(certificates, certificate)
	}

	if len(certificates) == 0 {
		return sdkdiag.AppendErrorf(diags, "no matching ACM Certificate found")
	}

	var matchedCertificate *awstypes.CertificateDetail
	if d.Get(names.AttrMostRecent).(bool) {
		matchedCertificate = certificates[0]

		for _, certificate := range certificates {
			matchedCertificate, err = mostRecentCertificate(certificate, matchedCertificate)

			if err != nil {
				return sdkdiag.AppendFromErr(diags, err)
			}
		}
	} else if n := len(certificates); n > 1 {
		return sdkdiag.AppendErrorf(diags, "%d matching ACM Certificates found", n)
	} else {
		matchedCertificate = certificates[0]
	}

	// Get the certificate data if the status is issued
	var output *acm.GetCertificateOutput
	if matchedCertificate.Status == awstypes.CertificateStatusIssued {
		arn := aws.ToString(matchedCertificate.CertificateArn)
		input := acm.GetCertificateInput{
			CertificateArn: aws.String(arn),
		}
		var err error

		output, err = conn.GetCertificate(ctx, &input)

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
	d.Set(names.AttrDomain, matchedCertificate.DomainName)
	d.Set(names.AttrStatus, matchedCertificate.Status)

	return diags
}

func mostRecentCertificate(i, j *awstypes.CertificateDetail) (*awstypes.CertificateDetail, error) {
	if i.Status != j.Status {
		return nil, fmt.Errorf("most_recent filtering on different ACM certificate statuses is not supported")
	}
	// Cover IMPORTED and ISSUED AMAZON_ISSUED certificates
	if i.Status == awstypes.CertificateStatusIssued {
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

func findCertificates(ctx context.Context, conn *acm.Client, input *acm.ListCertificatesInput, filter tfslices.Predicate[*awstypes.CertificateSummary]) ([]awstypes.CertificateSummary, error) {
	var output []awstypes.CertificateSummary

	pages := acm.NewListCertificatesPaginator(conn, input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if err != nil {
			return nil, err
		}

		for _, v := range page.CertificateSummaryList {
			if filter(&v) {
				output = append(output, v)
			}
		}
	}

	return output, nil
}
