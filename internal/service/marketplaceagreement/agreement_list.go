// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package marketplaceagreement

import (
	"context"
	"fmt"
	"iter"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/marketplaceagreement"
	awstypes "github.com/aws/aws-sdk-go-v2/service/marketplaceagreement/types"
	"github.com/hashicorp/terraform-plugin-framework/list"
	listschema "github.com/hashicorp/terraform-plugin-framework/list/schema"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/fwdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	fwflex "github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	tfiter "github.com/hashicorp/terraform-provider-aws/internal/iter"
	"github.com/hashicorp/terraform-provider-aws/internal/logging"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/smerr"
)

// @FrameworkListResource("aws_marketplaceagreement_agreement")
func newAgreementResourceAsListResource() list.ListResourceWithConfigure {
	return &agreementListResource{}
}

var _ list.ListResource = &agreementListResource{}

type agreementListResource struct {
	agreementResource
	framework.WithList
}

func (l *agreementListResource) ListResourceConfigSchema(ctx context.Context, _ list.ListResourceSchemaRequest, response *list.ListResourceSchemaResponse) {
	response.Schema = listschema.Schema{}
}

func (l *agreementListResource) List(ctx context.Context, request list.ListRequest, stream *list.ListResultsStream) {
	conn := l.Meta().MarketplaceAgreementClient(ctx)

	// AgreementType is required, and Status is one of the few filters an acceptor may combine with it.
	input := marketplaceagreement.SearchAgreementsInput{
		Filters: []awstypes.Filter{
			{Name: aws.String("PartyType"), Values: []string{"Acceptor"}},
			{Name: aws.String("AgreementType"), Values: []string{"PurchaseAgreement"}},
			{Name: aws.String("Status"), Values: []string{string(awstypes.AgreementStatusActive)}},
		},
	}

	stream.Results = func(yield func(list.ListResult) bool) {
		for item, err := range listAgreements(ctx, conn, &input) {
			if err != nil {
				yield(fwdiag.NewListResultErrorDiagnostic(err))
				return
			}

			agreementID := aws.ToString(item.AgreementId)
			ctx := tflog.SetField(ctx, logging.ResourceAttributeKey("agreement_id"), agreementID)

			var (
				out   *marketplaceagreement.DescribeAgreementOutput
				terms []awstypes.AcceptedTerm
			)
			if request.IncludeResource {
				out, err = findAgreementByID(ctx, conn, agreementID)
				if retry.NotFound(err) {
					continue
				}
				if err != nil {
					yield(fwdiag.NewListResultErrorDiagnostic(err))
					return
				}

				terms, err = findAcceptedTermsByAgreementID(ctx, conn, agreementID)
				if err != nil {
					yield(fwdiag.NewListResultErrorDiagnostic(err))
					return
				}
			}

			result := request.NewListResult(ctx)
			var data agreementResourceModel
			l.SetResult(ctx, l.Meta(), request.IncludeResource, &data, &result, func() {
				data.AgreementID = fwflex.StringValueToFramework(ctx, agreementID)

				if request.IncludeResource {
					flattenAgreement(ctx, out, &data)

					smerr.AddEnrich(ctx, &result.Diagnostics, flattenAcceptedTerms(ctx, terms, &data))
					if result.Diagnostics.HasError() {
						return
					}
				}

				result.DisplayName = agreementID
			})

			if !yield(result) {
				return
			}
		}
	}
}

func listAgreements(ctx context.Context, conn *marketplaceagreement.Client, input *marketplaceagreement.SearchAgreementsInput, optFns ...func(*marketplaceagreement.Options)) iter.Seq2[awstypes.AgreementViewSummary, error] {
	return tfiter.ConcatValuesWithError(listAgreementPages(ctx, conn, input, optFns...))
}

func listAgreementPages(ctx context.Context, conn *marketplaceagreement.Client, input *marketplaceagreement.SearchAgreementsInput, optFns ...func(*marketplaceagreement.Options)) iter.Seq2[[]awstypes.AgreementViewSummary, error] {
	return func(yield func([]awstypes.AgreementViewSummary, error) bool) {
		pages := marketplaceagreement.NewSearchAgreementsPaginator(conn, input)
		for pages.HasMorePages() {
			page, err := pages.NextPage(ctx, optFns...)
			if err != nil {
				yield(nil, fmt.Errorf("listing Marketplace Agreement Agreements: %w", err))
				return
			}

			if !yield(page.AgreementViewSummaries, nil) {
				return
			}
		}
	}
}
