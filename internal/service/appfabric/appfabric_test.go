// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package appfabric_test

import (
	"testing"
	"time"

	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
)

const serializeDelay = 10 * time.Second

// Serialize to limit API rate-limit exceeded errors (ServiceQuotaExceededException).
func TestAccAppFabric_serial(t *testing.T) {
	t.Parallel()

	testCases := map[string]map[string]func(t *testing.T){
		"AppBundle": {
			acctest.CtBasic:      testAccAppBundle_basic,
			acctest.CtDisappears: testAccAppBundle_disappears,
			"cmk":                testAccAppBundle_cmk,
			"tags":               testAccAppBundle_tags,
		},
		"AppAuthorization": {
			acctest.CtBasic:      testAccAppAuthorization_basic,
			acctest.CtDisappears: testAccAppAuthorization_disappears,
			"apiKeyUpdate":       testAccAppAuthorization_apiKeyUpdate,
			"oath2Update":        testAccAppAuthorization_oath2Update,
			"tags":               testAccAppAuthorization_tags,
		},
		"AppAuthorizationConnection": {
			acctest.CtBasic: testAccAppAuthorizationConnection_basic,
			"oath2Connect":  testAccAppAuthorizationConnection_OAuth2,
		},
	}

	acctest.RunSerialTests2Levels(t, testCases, serializeDelay)
}
