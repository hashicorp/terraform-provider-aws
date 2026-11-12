// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package continuum_test

import (
	"context"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/continuum"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
)

// Continuum allows one collector per account and Region, and creating an environment starts a scan,
// so every acceptance test runs serially.
func TestAccContinuum_serial(t *testing.T) {
	t.Parallel()

	testCases := map[string]map[string]func(t *testing.T){
		"Collector": {
			acctest.CtBasic:      testAccCollector_basic,
			acctest.CtDisappears: testAccCollector_disappears,
		},
	}

	acctest.RunSerialTests2Levels(t, testCases, 0)
}

func testAccPreCheck(ctx context.Context, t *testing.T) {
	conn := acctest.ProviderMeta(ctx, t).ContinuumClient(ctx)

	input := continuum.ListIntegrationsInput{}
	_, err := conn.ListIntegrations(ctx, &input)

	if acctest.PreCheckSkipError(err) {
		t.Skipf("skipping acceptance testing: %s", err)
	}
	if err != nil {
		t.Fatalf("unexpected PreCheck error: %s", err)
	}
}
