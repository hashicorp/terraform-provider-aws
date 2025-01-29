// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package oam_test

import (
	"testing"

	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
)

func TestAccObservabilityAccessManager_serial(t *testing.T) {
	t.Parallel()

	testCases := map[string]map[string]func(t *testing.T){
		"Link": {
			acctest.CtBasic:         testAccObservabilityAccessManagerLink_basic,
			acctest.CtDisappears:    testAccObservabilityAccessManagerLink_disappears,
			"update":                testAccObservabilityAccessManagerLink_update,
			"tags":                  testAccObservabilityAccessManagerLink_tags,
			"logGroupConfiguration": testAccObservabilityAccessManagerLink_logGroupConfiguration,
			"metricConfiguration":   testAccObservabilityAccessManagerLink_metricConfiguration,
		},
		"LinkDataSource": {
			acctest.CtBasic:         testAccObservabilityAccessManagerLinkDataSource_basic,
			"logGroupConfiguration": testAccObservabilityAccessManagerLinkDataSource_logGroupConfiguration,
			"metricConfiguration":   testAccObservabilityAccessManagerLinkDataSource_metricConfiguration,
		},
		"LinksDataSource": {
			acctest.CtBasic: testAccObservabilityAccessManagerLinksDataSource_basic,
		},
		"Sink": {
			acctest.CtBasic:      testAccObservabilityAccessManagerSink_basic,
			acctest.CtDisappears: testAccObservabilityAccessManagerSink_disappears,
			"tags":               testAccObservabilityAccessManagerSink_tags,
		},
		"SinkDataSource": {
			acctest.CtBasic: testAccObservabilityAccessManagerSinkDataSource_basic,
		},
		"SinkPolicy": {
			acctest.CtBasic: testAccObservabilityAccessManagerSinkPolicy_basic,
			"update":        testAccObservabilityAccessManagerSinkPolicy_update,
		},
		"SinksDataSource": {
			acctest.CtBasic: testAccObservabilityAccessManagerSinksDataSource_basic,
		},
	}

	acctest.RunSerialTests2Levels(t, testCases, 0)
}
