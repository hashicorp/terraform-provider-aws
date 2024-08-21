// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package appfabric_test

import (
	"context"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/appfabric"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
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
		"Ingestion": {
			acctest.CtBasic:      testAccIngestion_basic,
			acctest.CtDisappears: testAccIngestion_disappears,
			"tags":               testAccIngestion_tags,
		},
		"IngestionDestination": {
			acctest.CtBasic:      testAccIngestionDestination_basic,
			acctest.CtDisappears: testAccIngestionDestination_disappears,
			"tags":               testAccIngestionDestination_tags,
			"update":             testAccIngestionDestination_update,
			"firehose":           testAccIngestionDestination_firehose,
		},
	}

	acctest.RunSerialTests2Levels(t, testCases, serializeDelay)
}

func testAccPreCheck(ctx context.Context, t *testing.T) {
	conn := acctest.Provider.Meta().(*conns.AWSClient).AppFabricClient(ctx)

	input := &appfabric.ListAppBundlesInput{}
	_, err := conn.ListAppBundles(ctx, input)

	if acctest.PreCheckSkipError(err) {
		t.Skipf("skipping acceptance testing: %s", err)
	}

	if err != nil {
		t.Fatalf("unexpected PreCheck error: %s", err)
	}
}
