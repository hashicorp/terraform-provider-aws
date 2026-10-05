// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package marketplaceagreement_test

import (
	"context"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/marketplaceagreement"
	awstypes "github.com/aws/aws-sdk-go-v2/service/marketplaceagreement/types"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
)

// An account can hold only one active agreement per offer, and every test
// subscribes to the same offer, so the tests run one at a time.
func TestAccMarketplaceAgreement_serial(t *testing.T) {
	t.Parallel()

	testCases := map[string]map[string]func(t *testing.T){
		"Agreement": {
			acctest.CtBasic:        testAccMarketplaceAgreementAgreement_basic,
			acctest.CtDisappears:   testAccMarketplaceAgreementAgreement_disappears,
			"Identity":             testAccMarketplaceAgreementAgreement_identitySerial,
			"List_basic":           testAccMarketplaceAgreementAgreement_List_basic,
			"List_includeResource": testAccMarketplaceAgreementAgreement_List_includeResource,
		},
	}

	acctest.RunSerialTests2Levels(t, testCases, 5*time.Second)
}

func testAccPreCheck(ctx context.Context, t *testing.T) {
	conn := acctest.ProviderMeta(ctx, t).MarketplaceAgreementClient(ctx)

	input := marketplaceagreement.SearchAgreementsInput{
		Filters: []awstypes.Filter{
			{Name: aws.String("PartyType"), Values: []string{"Acceptor"}},
			{Name: aws.String("AgreementType"), Values: []string{"PurchaseAgreement"}},
		},
		MaxResults: aws.Int32(1),
	}
	_, err := conn.SearchAgreements(ctx, &input)

	if acctest.PreCheckSkipError(err) {
		t.Skipf("skipping acceptance testing: %s", err)
	}
	if err != nil {
		t.Fatalf("unexpected PreCheck error: %s", err)
	}
}
