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
			acctest.CtBasic: testAccAuthorizationTokenDataSource_basic,
			"duration":      testAccAuthorizationTokenDataSource_duration,
			"owner":         testAccAuthorizationTokenDataSource_owner,
		},
		"Domain": {
			acctest.CtBasic:                 testAccDomain_basic,
			"defaultEncryptionKey":          testAccDomain_defaultEncryptionKey,
			acctest.CtDisappears:            testAccDomain_disappears,
			"migrateAssetSizeBytesToString": testAccDomain_MigrateAssetSizeBytesToString,
			"tags":                          testAccDomain_tags,
		},
		"DomainPermissionsPolicy": {
			acctest.CtBasic:      testAccDomainPermissionsPolicy_basic,
			acctest.CtDisappears: testAccDomainPermissionsPolicy_disappears,
			"owner":              testAccDomainPermissionsPolicy_owner,
			"disappearsDomain":   testAccDomainPermissionsPolicy_Disappears_domain,
			"ignoreEquivalent":   testAccDomainPermissionsPolicy_ignoreEquivalent,
		},
		"Repository": {
			acctest.CtBasic:      testAccRepository_basic,
			"description":        testAccRepository_description,
			acctest.CtDisappears: testAccRepository_disappears,
			"externalConnection": testAccRepository_externalConnection,
			"owner":              testAccRepository_owner,
			"tags":               testAccRepository_tags,
			"upstreams":          testAccRepository_upstreams,
		},
		"RepositoryEndpointDataSource": {
			acctest.CtBasic: testAccRepositoryEndpointDataSource_basic,
			"owner":         testAccRepositoryEndpointDataSource_owner,
		},
		"RepositoryPermissionsPolicy": {
			acctest.CtBasic:      testAccRepositoryPermissionsPolicy_basic,
			acctest.CtDisappears: testAccRepositoryPermissionsPolicy_disappears,
			"owner":              testAccRepositoryPermissionsPolicy_owner,
			"disappearsDomain":   testAccRepositoryPermissionsPolicy_Disappears_domain,
			"ignoreEquivalent":   testAccRepositoryPermissionsPolicy_ignoreEquivalent,
		},
	}

	acctest.RunSerialTests2Levels(t, testCases, 0)
}
