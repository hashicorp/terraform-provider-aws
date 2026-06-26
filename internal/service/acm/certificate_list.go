// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package acm

import (
	"context"
	"fmt"
	"iter"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/acm"
	awstypes "github.com/aws/aws-sdk-go-v2/service/acm/types"
	"github.com/hashicorp/terraform-plugin-framework/list"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/fwdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	"github.com/hashicorp/terraform-provider-aws/internal/logging"
	inttypes "github.com/hashicorp/terraform-provider-aws/internal/types"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKListResource("aws_acm_certificate")
func newCertificateResourceAsListResource() inttypes.ListResourceForSDK {
	l := certificateListResource{}
	l.SetResourceSchema(resourceCertificate())
	return &l
}

var _ list.ListResource = &certificateListResource{}

type certificateListResource struct {
	framework.ListResourceWithSDKv2Resource
}

type certificateListResourceModel struct {
	framework.WithRegionModel
}

func (l *certificateListResource) List(ctx context.Context, request list.ListRequest, stream *list.ListResultsStream) {
	awsClient := l.Meta()
	conn := awsClient.ACMClient(ctx)

	var query certificateListResourceModel
	if request.Config.Raw.IsKnown() && !request.Config.Raw.IsNull() {
		if diags := request.Config.Get(ctx, &query); diags.HasError() {
			stream.Results = list.ListResultsStreamDiagnostics(diags)
			return
		}
	}

	tflog.Info(ctx, "Listing ACM Certificates")

	stream.Results = func(yield func(list.ListResult) bool) {
		input := acm.ListCertificatesInput{
			Includes: &awstypes.Filters{
				KeyTypes: enum.EnumValues[awstypes.KeyAlgorithm](),
			},
		}
		for item, err := range listCertificates(ctx, conn, &input) {
			if err != nil {
				result := fwdiag.NewListResultErrorDiagnostic(err)
				yield(result)
				return
			}

			arn := aws.ToString(item.CertificateArn)
			ctx := tflog.SetField(ctx, logging.ResourceAttributeKey(names.AttrARN), arn)

			result := request.NewListResult(ctx)
			rd := l.ResourceData()
			rd.SetId(arn)
			rd.Set(names.AttrARN, arn)

			if request.IncludeResource {
				tflog.Info(ctx, "Reading ACM Certificate")
				certificate, err := findCertificateByARN(ctx, conn, arn)
				if err != nil {
					tflog.Error(ctx, "Reading ACM Certificate", map[string]any{
						"err": err.Error(),
					})
					continue
				}

				diags := resourceCertificateFlatten(ctx, rd, certificate)
				if diags.HasError() {
					tflog.Error(ctx, "Reading ACM Certificate", map[string]any{
						"diags": sdkdiag.DiagnosticsString(diags),
					})
					continue
				}
			}

			result.DisplayName = aws.ToString(item.DomainName)

			l.SetResult(ctx, awsClient, request.IncludeResource, rd, &result)
			if result.Diagnostics.HasError() {
				yield(result)
				return
			}

			if !yield(result) {
				return
			}
		}
	}
}

func listCertificates(ctx context.Context, conn *acm.Client, input *acm.ListCertificatesInput) iter.Seq2[awstypes.CertificateSummary, error] {
	return func(yield func(awstypes.CertificateSummary, error) bool) {
		pages := acm.NewListCertificatesPaginator(conn, input)
		for pages.HasMorePages() {
			page, err := pages.NextPage(ctx)
			if err != nil {
				yield(awstypes.CertificateSummary{}, fmt.Errorf("listing ACM Certificates: %w", err))
				return
			}

			for _, item := range page.CertificateSummaryList {
				if !yield(item, nil) {
					return
				}
			}
		}
	}
}
