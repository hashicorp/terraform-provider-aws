// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package grafana_test

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func init() {
	acctest.RegisterServiceErrorCheckFunc(names.GrafanaServiceID, testAccErrorCheckSkip)
}

func testAccErrorCheckSkip(t *testing.T) resource.ErrorCheckFunc {
	return acctest.ErrorCheckSkipMessagesContaining(t,
		"Active marketplace agreement not found",
	)
}

func TestAccGrafana_serial(t *testing.T) {
	t.Parallel()

	testCases := map[string]map[string]func(t *testing.T){
		"Workspace": {
			"saml":                     testAccWorkspace_saml,
			"sso":                      testAccWorkspace_sso,
			acctest.CtDisappears:       testAccWorkspace_disappears,
			"organization":             testAccWorkspace_organization,
			"dataSources":              testAccWorkspace_dataSources,
			"permissionType":           testAccWorkspace_permissionType,
			"notificationDestinations": testAccWorkspace_notificationDestinations,
			"tags":                     testAccWorkspace_tags,
			"vpc":                      testAccWorkspace_vpc,
			"configuration":            testAccWorkspace_configuration,
			"networkAccess":            testAccWorkspace_networkAccess,
			"version":                  testAccWorkspace_version,
		},
		"ApiKey": {
			acctest.CtBasic: testAccWorkspaceAPIKey_basic,
		},
		"DataSource": {
			acctest.CtBasic: testAccWorkspaceDataSource_basic,
		},
		"LicenseAssociation": {
			"enterpriseFreeTrial": testAccLicenseAssociation_freeTrial,
		},
		"SamlConfiguration": {
			acctest.CtBasic: testAccWorkspaceSAMLConfiguration_basic,
			"loginValidity": testAccWorkspaceSAMLConfiguration_loginValidity,
			"assertions":    testAccWorkspaceSAMLConfiguration_assertions,
		},
		"RoleAssociation": {
			"usersAdmin":           testAccRoleAssociation_usersAdmin,
			"usersEditor":          testAccRoleAssociation_usersEditor,
			"groupsAdmin":          testAccRoleAssociation_groupsAdmin,
			"groupsEditor":         testAccRoleAssociation_groupsEditor,
			"usersAndGroupsAdmin":  testAccRoleAssociation_usersAndGroupsAdmin,
			"usersAndGroupsEditor": testAccRoleAssociation_usersAndGroupsEditor,
		},
	}

	acctest.RunSerialTests2Levels(t, testCases, 0)
}
