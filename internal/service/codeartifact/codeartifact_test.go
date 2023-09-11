// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package codeartifact_test

import (
	"testing"

	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
)

func TestAccCodeArtifact_serial(t *testing.T) {
	t.Parallel()

	testCases := map[string]map[string]func(t *testing.T){
		"AuthorizationTokenDataSource": {
			"basic":    testAccAuthorizationTokenDataSource_basic,
			"duration": testAccAuthorizationTokenDataSource_duration,
			"owner":    testAccAuthorizationTokenDataSource_owner,
		},
		"Domain": {
			"basic":                         testAccDomain_basic,
			"defaultEncryptionKey":          testAccDomain_defaultEncryptionKey,
			"disappears":                    testAccDomain_disappears,
			"migrateAssetSizeBytesToString": testAccDomain_MigrateAssetSizeBytesToString,
			"tags":                          testAccDomain_tags,
		},
		"DomainPermissionsPolicy": {
			"basic":            testAccDomainPermissionsPolicy_basic,
			"disappears":       testAccDomainPermissionsPolicy_disappears,
			"owner":            testAccDomainPermissionsPolicy_owner,
			"disappearsDomain": testAccDomainPermissionsPolicy_Disappears_domain,
			"ignoreEquivalent": testAccDomainPermissionsPolicy_ignoreEquivalent,
		},
		"Repository": {
			"basic":              testAccRepository_basic,
			"description":        testAccRepository_description,
			"disappears":         testAccRepository_disappears,
			"externalConnection": testAccRepository_externalConnection,
			"owner":              testAccRepository_owner,
			"tags":               testAccRepository_tags,
			"upstreams":          testAccRepository_upstreams,
		},
		"RepositoryEndpointDataSource": {
			"basic": testAccRepositoryEndpointDataSource_basic,
			"owner": testAccRepositoryEndpointDataSource_owner,
		},
		"RepositoryPermissionsPolicy": {
			"basic":            testAccRepositoryPermissionsPolicy_basic,
			"disappears":       testAccRepositoryPermissionsPolicy_disappears,
			"owner":            testAccRepositoryPermissionsPolicy_owner,
			"disappearsDomain": testAccRepositoryPermissionsPolicy_Disappears_domain,
			"ignoreEquivalent": testAccRepositoryPermissionsPolicy_ignoreEquivalent,
		},
	}

	acctest.RunSerialTests2Levels(t, testCases, 0)
}
