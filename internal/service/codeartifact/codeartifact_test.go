package codeartifact_test

import "testing"

func TestAccCodeArtifact_serial(t *testing.T) {
	testCases := map[string]map[string]func(t *testing.T){
		"AuthorizationTokenDataSource": {
			"basic":    testAccCodeArtifactAuthorizationTokenDataSource_basic,
			"duration": testAccCodeArtifactAuthorizationTokenDataSource_duration,
			"owner":    testAccCodeArtifactAuthorizationTokenDataSource_owner,
		},
		"Domain": {
			"basic":                testAccCodeArtifactDomain_basic,
			"defaultEncryptionKey": testAccCodeArtifactDomain_defaultEncryptionKey,
			"disappears":           testAccCodeArtifactDomain_disappears,
			"tags":                 testAccCodeArtifactDomain_tags,
		},
		"DomainPermissionsPolicy": {
			"basic":            testAccCodeArtifactDomainPermissionsPolicy_basic,
			"disappears":       testAccCodeArtifactDomainPermissionsPolicy_disappears,
			"owner":            testAccCodeArtifactDomainPermissionsPolicy_owner,
			"disappearsDomain": testAccCodeArtifactDomainPermissionsPolicy_Disappears_domain,
			"ignoreEquivalent": testAccCodeArtifactDomainPermissionsPolicy_ignoreEquivalent,
		},
		"Repository": {
			"basic":              testAccCodeArtifactRepository_basic,
			"description":        testAccCodeArtifactRepository_description,
			"disappears":         testAccCodeArtifactRepository_disappears,
			"externalConnection": testAccCodeArtifactRepository_externalConnection,
			"owner":              testAccCodeArtifactRepository_owner,
			"tags":               testAccCodeArtifactRepository_tags,
			"upstreams":          testAccCodeArtifactRepository_upstreams,
		},
		"RepositoryEndpointDataSource": {
			"basic": testAccCodeArtifactRepositoryEndpointDataSource_basic,
			"owner": testAccCodeArtifactRepositoryEndpointDataSource_owner,
		},
		"RepositoryPermissionsPolicy": {
			"basic":            testAccCodeArtifactRepositoryPermissionsPolicy_basic,
			"disappears":       testAccCodeArtifactRepositoryPermissionsPolicy_disappears,
			"owner":            testAccCodeArtifactRepositoryPermissionsPolicy_owner,
			"disappearsDomain": testAccCodeArtifactRepositoryPermissionsPolicy_Disappears_domain,
			"ignoreEquivalent": testAccCodeArtifactRepositoryPermissionsPolicy_ignoreEquivalent,
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
