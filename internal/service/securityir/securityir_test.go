// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package securityir_test

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/securityir"
	awstypes "github.com/aws/aws-sdk-go-v2/service/securityir/types"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	tfsecurityir "github.com/hashicorp/terraform-provider-aws/internal/service/securityir"
)

// serializeDelay is applied between serialized subtests. An account or organization
// can have only one Security Incident Response membership, so every test contends for it.
const serializeDelay = 5 * time.Second

func TestAccSecurityIR_serial(t *testing.T) {
	t.Parallel()

	testCases := map[string]map[string]func(t *testing.T){
		"Membership": {
			acctest.CtBasic:           testAccMembership_basic,
			acctest.CtDisappears:      testAccMembership_disappears,
			"membershipName":          testAccMembership_membershipName,
			"incidentResponseTeam":    testAccMembership_incidentResponseTeam,
			"optInFeature":            testAccMembership_optInFeature,
			"coverEntireOrganization": testAccMembership_coverEntireOrganization,
			"organizationalUnits":     testAccMembership_organizationalUnits,
			"tags":                    testAccSecurityIRMembership_tagsSerial,
			"Identity":                testAccSecurityIRMembership_identitySerial,
			"List_basic":              testAccMembership_List_basic,
			"List_includeResource":    testAccMembership_List_includeResource,
			"List_regionOverride":     testAccMembership_List_regionOverride,
		},
	}

	acctest.RunSerialTests2Levels(t, testCases, serializeDelay)
}

// testAccPreCheck skips when the account already has an active membership that was not
// created by an acceptance test, as tests would otherwise fail or affect a real membership.
func testAccPreCheck(ctx context.Context, t *testing.T) {
	conn := acctest.ProviderMeta(ctx, t).SecurityIRClient(ctx)

	var input securityir.ListMembershipsInput
	var items []awstypes.ListMembershipItem
	pages := securityir.NewListMembershipsPaginator(conn, &input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if acctest.PreCheckSkipError(err) {
			t.Skipf("skipping acceptance testing: %s", err)
		}
		if err != nil {
			t.Fatalf("unexpected PreCheck error: %s", err)
		}

		items = append(items, page.Items...)
	}

	for _, v := range items {
		if v.MembershipStatus != awstypes.MembershipStatusActive {
			continue
		}

		id := aws.ToString(v.MembershipId)
		membership, err := tfsecurityir.FindMembershipByID(ctx, conn, id)
		if retry.NotFound(err) {
			continue
		}
		// A membership that cannot be read here (for example, one in another Region) is assumed not to be a test membership.
		if err != nil {
			t.Skipf("skipping acceptance testing: active Security Incident Response membership %s exists: %s", id, err)
		}

		if name := aws.ToString(membership.MembershipName); !strings.HasPrefix(name, acctest.ResourcePrefix) {
			t.Skipf("skipping acceptance testing: active Security Incident Response membership %q (%s) exists", name, id)
		}
	}
}
