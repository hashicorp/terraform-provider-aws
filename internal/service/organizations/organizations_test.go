package organizations_test

import (
	"testing"
)

func TestAccOrganizations_serial(t *testing.T) {
	testCases := map[string]map[string]func(t *testing.T){
		"Organization": {
			"basic":                      testAccOrganization_basic,
			"AwsServiceAccessPrincipals": testAccOrganization_serviceAccessPrincipals,
			"EnabledPolicyTypes":         testAccOrganization_EnabledPolicyTypes,
			"FeatureSet_Basic":           testAccOrganization_FeatureSet,
			"FeatureSet_Update":          testAccOrganization_FeatureSetUpdate,
			"FeatureSet_ForcesNew":       testAccOrganization_FeatureSetForcesNew,
			"DataSource":                 testAccOrganizationDataSource_basic,
		},
		"Account": {
			"basic":           testAccAccount_basic,
			"CloseOnDeletion": testAccAccount_CloseOnDeletion,
			"ParentId":        testAccAccount_ParentID,
			"Tags":            testAccAccount_Tags,
			"GovCloud":        testAccAccount_govCloud,
		},
		"OrganizationalUnit": {
			"basic":      testAccOrganizationalUnit_basic,
			"disappears": testAccOrganizationalUnit_disappears,
			"Name":       testAccOrganizationalUnit_Name,
			"Tags":       testAccOrganizationalUnit_Tags,
		},
		"OrganizationalUnits": {
			"DataSource": testAccOrganizationalUnitsDataSource_basic,
		},
		"Policy": {
			"basic":                  testAccPolicy_basic,
			"concurrent":             testAccPolicy_concurrent,
			"Description":            testAccPolicy_description,
			"Tags":                   testAccPolicy_tags,
			"disappears":             testAccPolicy_disappears,
			"Type_AI_OPT_OUT":        testAccPolicy_type_AI_OPT_OUT,
			"Type_Backup":            testAccPolicy_type_Backup,
			"Type_SCP":               testAccPolicy_type_SCP,
			"Type_Tag":               testAccPolicy_type_Tag,
			"ImportAwsManagedPolicy": testAccPolicy_importManagedPolicy,
		},
		"PolicyAttachment": {
			"Account":            testAccPolicyAttachment_Account,
			"OrganizationalUnit": testAccPolicyAttachment_OrganizationalUnit,
			"Root":               testAccPolicyAttachment_Root,
		},
		"DelegatedAdministrator": {
			"basic":      testAccDelegatedAdministrator_basic,
			"disappears": testAccDelegatedAdministrator_disappears,
		},
		"ResourceTags": {
			"basic": testAccResourceTagsDataSource_basic,
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
