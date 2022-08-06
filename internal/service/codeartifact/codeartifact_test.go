package codeartifact_test

import "testing"

func TestAccCodeArtifact_serial(t *testing.T) {
	testCases := map[string]map[string]func(t *testing.T){
		"AuthorizationTokenDataSource": {
			"basic":    testAccAuthorizationTokenDataSource_basic,
			"duration": testAccAuthorizationTokenDataSource_duration,
			"owner":    testAccAuthorizationTokenDataSource_owner,
		},
		"Domain": {
			"basic":                testAccDomain_basic,
			"defaultEncryptionKey": testAccDomain_defaultEncryptionKey,
			"disappears":           testAccDomain_disappears,
			"tags":                 testAccDomain_tags,
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

	for group, m := range testCases {
		m := m
		t.Run(group, func(t *testing.T) {
			for name, tc := range m {
				tc := tc
				t.Run(name, func(t *testing.T) {
					tc(t)
				})
			}
		})
	}
}
