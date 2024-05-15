// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package codeartifact_test

import (
	"testing"

	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccCodeArtifact_serial(t *testing.T) {
	t.Parallel()

	testCases := map[string]map[string]func(t *testing.T){
		"AuthorizationTokenDataSource": {
			acctest.CtBasic:    testAccAuthorizationTokenDataSource_basic,
			names.AttrDuration: testAccAuthorizationTokenDataSource_duration,
			names.AttrOwner:    testAccAuthorizationTokenDataSource_owner,
		},
		"Domain": {
			acctest.CtBasic:                 testAccDomain_basic,
			"defaultEncryptionKey":          testAccDomain_defaultEncryptionKey,
			"disappears":                    testAccDomain_disappears,
			"migrateAssetSizeBytesToString": testAccDomain_MigrateAssetSizeBytesToString,
			names.AttrTags:                  testAccDomain_tags,
		},
		"DomainPermissionsPolicy": {
			acctest.CtBasic:    testAccDomainPermissionsPolicy_basic,
			"disappears":       testAccDomainPermissionsPolicy_disappears,
			names.AttrOwner:    testAccDomainPermissionsPolicy_owner,
			"disappearsDomain": testAccDomainPermissionsPolicy_Disappears_domain,
			"ignoreEquivalent": testAccDomainPermissionsPolicy_ignoreEquivalent,
		},
		"Repository": {
			acctest.CtBasic:       testAccRepository_basic,
			names.AttrDescription: testAccRepository_description,
			"disappears":          testAccRepository_disappears,
			"externalConnection":  testAccRepository_externalConnection,
			names.AttrOwner:       testAccRepository_owner,
			names.AttrTags:        testAccRepository_tags,
			"upstreams":           testAccRepository_upstreams,
		},
		"RepositoryEndpointDataSource": {
			acctest.CtBasic: testAccRepositoryEndpointDataSource_basic,
			names.AttrOwner: testAccRepositoryEndpointDataSource_owner,
		},
		"RepositoryPermissionsPolicy": {
			acctest.CtBasic:    testAccRepositoryPermissionsPolicy_basic,
			"disappears":       testAccRepositoryPermissionsPolicy_disappears,
			names.AttrOwner:    testAccRepositoryPermissionsPolicy_owner,
			"disappearsDomain": testAccRepositoryPermissionsPolicy_Disappears_domain,
			"ignoreEquivalent": testAccRepositoryPermissionsPolicy_ignoreEquivalent,
		},
	}

	acctest.RunSerialTests2Levels(t, testCases, 0)
}
