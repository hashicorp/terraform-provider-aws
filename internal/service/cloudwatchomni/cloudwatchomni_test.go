// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package cloudwatchomni_test

import (
	"context"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/cloudwatchomni"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
)

// Tests are serialized because an account can have at most one CloudWatch Omni domain.
func TestAccCloudWatchOmni_serial(t *testing.T) {
	t.Parallel()

	testCases := map[string]map[string]func(t *testing.T){
		"Domain": {
			acctest.CtBasic:      testAccCloudWatchOmniDomain_basic,
			acctest.CtDisappears: testAccCloudWatchOmniDomain_disappears,
			acctest.CtName:       testAccCloudWatchOmniDomain_name,
			"identityCenter":     testAccCloudWatchOmniDomain_identityCenter,
			"Identity":           testAccCloudWatchOmniDomain_identitySerial,
		},
	}

	acctest.RunSerialTests2Levels(t, testCases, 0)
}

func testAccPreCheck(ctx context.Context, t *testing.T) {
	conn := acctest.ProviderMeta(ctx, t).CloudWatchOmniClient(ctx)

	var input cloudwatchomni.ListDomainsInput
	out, err := conn.ListDomains(ctx, &input)

	if acctest.PreCheckSkipError(err) {
		t.Skipf("skipping acceptance testing: %s", err)
	}
	if err != nil {
		t.Fatalf("unexpected PreCheck error: %s", err)
	}

	// Creating a domain fails if the account already has one.
	if len(out.Items) > 0 {
		t.Skip("skipping acceptance testing: account already has a CloudWatch Omni domain")
	}
}
