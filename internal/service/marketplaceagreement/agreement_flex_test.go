// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package marketplaceagreement

import (
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	awstypes "github.com/aws/aws-sdk-go-v2/service/marketplaceagreement/types"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	fwflex "github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
)

// Read rebuilds requested_term from the accepted terms. If that doesn't expand back to
// exactly what Create sent, every plan after apply would replace the agreement.
func TestRequestedTermModel_roundTrip(t *testing.T) {
	t.Parallel()

	accepted := []awstypes.AcceptedTerm{
		&awstypes.AcceptedTermMemberLegalTerm{Value: awstypes.LegalTerm{
			Id: aws.String("term-legal"),
		}},
		&awstypes.AcceptedTermMemberConfigurableUpfrontPricingTerm{Value: awstypes.ConfigurableUpfrontPricingTerm{
			Id: aws.String("term-cupt"),
			Configuration: &awstypes.ConfigurableUpfrontPricingTermConfiguration{
				SelectorValue: aws.String("P12M"),
				Dimensions: []awstypes.Dimension{
					{DimensionKey: aws.String("AdminUsers"), DimensionValue: 5},
					{DimensionKey: aws.String("ReadOnlyUsers"), DimensionValue: 10},
				},
			},
		}},
		&awstypes.AcceptedTermMemberRenewalTerm{Value: awstypes.RenewalTerm{
			Id:            aws.String("term-renewal"),
			Configuration: &awstypes.RenewalTermConfiguration{EnableAutoRenew: aws.Bool(false)},
		}},
		&awstypes.AcceptedTermMemberVariablePaymentTerm{Value: awstypes.VariablePaymentTerm{
			Id: aws.String("term-vpt"),
			Configuration: &awstypes.VariablePaymentTermConfiguration{
				PaymentRequestApprovalStrategy: awstypes.PaymentRequestApprovalStrategyAutoApproveOnExpiration,
				ExpirationDuration:             aws.String("P10D"),
			},
		}},
	}
	want := []awstypes.RequestedTerm{
		{Id: aws.String("term-legal")},
		{
			Id: aws.String("term-cupt"),
			Configuration: &awstypes.RequestedTermConfigurationMemberConfigurableUpfrontPricingTermConfiguration{
				Value: awstypes.ConfigurableUpfrontPricingTermConfiguration{
					SelectorValue: aws.String("P12M"),
					Dimensions: []awstypes.Dimension{
						{DimensionKey: aws.String("AdminUsers"), DimensionValue: 5},
						{DimensionKey: aws.String("ReadOnlyUsers"), DimensionValue: 10},
					},
				},
			},
		},
		{
			Id: aws.String("term-renewal"),
			Configuration: &awstypes.RequestedTermConfigurationMemberRenewalTermConfiguration{
				Value: awstypes.RenewalTermConfiguration{EnableAutoRenew: aws.Bool(false)},
			},
		},
		{
			Id: aws.String("term-vpt"),
			Configuration: &awstypes.RequestedTermConfigurationMemberVariablePaymentTermConfiguration{
				Value: awstypes.VariablePaymentTermConfiguration{
					PaymentRequestApprovalStrategy: awstypes.PaymentRequestApprovalStrategyAutoApproveOnExpiration,
					ExpirationDuration:             aws.String("P10D"),
				},
			},
		},
	}

	ctx := t.Context()
	var model agreementResourceModel
	if diags := fwflex.Flatten(ctx, accepted, &model.RequestedTerms); diags.HasError() {
		t.Fatalf("flattening: %v", diags)
	}

	var got []awstypes.RequestedTerm
	if diags := fwflex.Expand(ctx, model.RequestedTerms, &got); diags.HasError() {
		t.Fatalf("expanding: %v", diags)
	}

	opts := []cmp.Option{
		cmpopts.IgnoreUnexported(
			awstypes.RequestedTerm{},
			awstypes.RequestedTermConfigurationMemberConfigurableUpfrontPricingTermConfiguration{},
			awstypes.RequestedTermConfigurationMemberRenewalTermConfiguration{},
			awstypes.RequestedTermConfigurationMemberVariablePaymentTermConfiguration{},
			awstypes.ConfigurableUpfrontPricingTermConfiguration{},
			awstypes.RenewalTermConfiguration{},
			awstypes.VariablePaymentTermConfiguration{},
			awstypes.Dimension{},
		),
		// requested_term and dimension are sets.
		cmpopts.SortSlices(func(a, b awstypes.RequestedTerm) bool { return aws.ToString(a.Id) < aws.ToString(b.Id) }),
		cmpopts.SortSlices(func(a, b awstypes.Dimension) bool { return aws.ToString(a.DimensionKey) < aws.ToString(b.DimensionKey) }),
	}
	if diff := cmp.Diff(want, got, opts...); diff != "" {
		t.Errorf("round trip mismatch (-want +got):\n%s", diff)
	}
}
