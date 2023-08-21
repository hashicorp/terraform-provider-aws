// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package finspace_test

import (
	"testing"

	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
)

func TestAccFinSpace_serial(t *testing.T) {
	t.Parallel()

	testCases := map[string]map[string]func(t *testing.T){
		"KxCluster": {
			"autoScaling":          testAccKxCluster_autoScaling,
			"basic":                testAccKxCluster_basic,
			"cacheConfigurations":  testAccKxCluster_cacheConfigurations,
			"code":                 testAccKxCluster_code,
			"commandLineArgs":      testAccKxCluster_commandLineArgs,
			"database":             testAccKxCluster_database,
			"description":          testAccKxCluster_description,
			"disappears":           testAccKxCluster_disappears,
			"executionRole":        testAccKxCluster_executionRole,
			"initializationScript": testAccKxCluster_initializationScript,
			"multiAZ":              testAccKxCluster_multiAZ,
			"rdb":                  testAccKxCluster_rdb,
			"tags":                 testAccKxCluster_tags,
		},
		"KxDatabase": {
			"basic":       testAccKxDatabase_basic,
			"description": testAccKxDatabase_description,
			"disappears":  testAccKxDatabase_disappears,
			"tags":        testAccKxDatabase_tags,
		},
		"KxEnvironment": {
			"basic":          testAccKxEnvironment_basic,
			"customDNS":      testAccKxEnvironment_customDNS,
			"description":    testAccKxEnvironment_description,
			"disappears":     testAccKxEnvironment_disappears,
			"tags":           testAccKxEnvironment_tags,
			"transitGateway": testAccKxEnvironment_transitGateway,
			"updateName":     testAccKxEnvironment_updateName,
		},
		"KxUser": {
			"basic":      testAccKxUser_basic,
			"disappears": testAccKxUser_disappears,
			"tags":       testAccKxUser_tags,
			"updateRole": testAccKxUser_updateRole,
		},
	}

	acctest.RunSerialTests2Levels(t, testCases, 0)
}
