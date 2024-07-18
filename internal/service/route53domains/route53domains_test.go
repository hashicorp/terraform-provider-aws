// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package route53domains_test

import (
	"context"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/route53domains"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccRoute53Domains_serial(t *testing.T) {
	t.Parallel()

	testCases := map[string]map[string]func(t *testing.T){
		"RegisteredDomain": {
			"tags":           testAccRegisteredDomain_tags,
			"autoRenew":      testAccRegisteredDomain_autoRenew,
			"contacts":       testAccRegisteredDomain_contacts,
			"contactPrivacy": testAccRegisteredDomain_contactPrivacy,
			"nameservers":    testAccRegisteredDomain_nameservers,
			"transferLock":   testAccRegisteredDomain_transferLock,
		},
		"DelegationSignerRecord": {
			acctest.CtBasic:      testAccDelegationSignerRecord_basic,
			acctest.CtDisappears: testAccDelegationSignerRecord_disappears,
		},
	}

	acctest.RunSerialTests2Levels(t, testCases, 0)
}

func testAccPreCheck(ctx context.Context, t *testing.T) {
	acctest.PreCheckPartitionHasService(t, names.Route53DomainsEndpointID)

	conn := acctest.Provider.Meta().(*conns.AWSClient).Route53DomainsClient(ctx)

	input := &route53domains.ListDomainsInput{}

	_, err := conn.ListDomains(ctx, input)

	if acctest.PreCheckSkipError(err) {
		t.Skipf("skipping acceptance testing: %s", err)
	}

	if err != nil {
		t.Fatalf("unexpected PreCheck error: %s", err)
	}
}
