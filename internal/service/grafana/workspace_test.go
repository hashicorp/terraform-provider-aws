// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package grafana_test

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/managedgrafana"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfgrafana "github.com/hashicorp/terraform-provider-aws/internal/service/grafana"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

func TestAccGrafana_serial(t *testing.T) {
	t.Parallel()

	testCases := map[string]map[string]func(t *testing.T){
		"Workspace": {
			"saml":                     testAccWorkspace_saml,
			"sso":                      testAccWorkspace_sso,
			"disappears":               testAccWorkspace_disappears,
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
			"basic": testAccWorkspaceAPIKey_basic,
		},
		"DataSource": {
			"basic": testAccWorkspaceDataSource_basic,
		},
		"LicenseAssociation": {
			"enterpriseFreeTrial": testAccLicenseAssociation_freeTrial,
		},
		"SamlConfiguration": {
			"basic":         testAccWorkspaceSAMLConfiguration_basic,
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

func testAccWorkspace_saml(t *testing.T) {
	ctx := acctest.Context(t)
	var v managedgrafana.WorkspaceDescription
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_grafana_workspace.test"
	iamRoleResourceName := "aws_iam_role.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckPartitionHasService(t, managedgrafana.EndpointsID) },
		ErrorCheck:               acctest.ErrorCheck(t, managedgrafana.EndpointsID),
		CheckDestroy:             testAccCheckWorkspaceDestroy(ctx),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccWorkspaceConfig_authenticationProvider(rName, "SAML"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckWorkspaceExists(ctx, resourceName, &v),
					acctest.MatchResourceAttrRegionalARN(resourceName, "arn", "grafana", regexp.MustCompile(`/workspaces/.+`)),
					resource.TestCheckResourceAttr(resourceName, "account_access_type", managedgrafana.AccountAccessTypeCurrentAccount),
					resource.TestCheckResourceAttr(resourceName, "authentication_providers.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "authentication_providers.0", managedgrafana.AuthenticationProviderTypesSaml),
					resource.TestCheckResourceAttr(resourceName, "data_sources.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "description", ""),
					resource.TestCheckResourceAttrSet(resourceName, "endpoint"),
					resource.TestCheckResourceAttrSet(resourceName, "grafana_version"),
					resource.TestCheckResourceAttrSet(resourceName, "name"),
					resource.TestCheckResourceAttr(resourceName, "notification_destinations.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "organization_role_name", ""),
					resource.TestCheckResourceAttr(resourceName, "organizational_units.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "permission_type", managedgrafana.PermissionTypeServiceManaged),
					resource.TestCheckResourceAttrPair(resourceName, "role_arn", iamRoleResourceName, "arn"),
					resource.TestCheckResourceAttr(resourceName, "saml_configuration_status", managedgrafana.SamlConfigurationStatusNotConfigured),
					resource.TestCheckResourceAttr(resourceName, "stack_set_name", ""),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
					resource.TestCheckResourceAttr(resourceName, "vpc_configuration.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "network_access_control.#", "0"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccWorkspace_vpc(t *testing.T) {
	ctx := acctest.Context(t)
	var v managedgrafana.WorkspaceDescription
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_grafana_workspace.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckPartitionHasService(t, managedgrafana.EndpointsID) },
		ErrorCheck:               acctest.ErrorCheck(t, managedgrafana.EndpointsID),
		CheckDestroy:             testAccCheckWorkspaceDestroy(ctx),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccWorkspaceConfig_vpc(rName, 2),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckWorkspaceExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "vpc_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "vpc_configuration.0.security_group_ids.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "vpc_configuration.0.subnet_ids.#", "2"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccWorkspaceConfig_vpc(rName, 3),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckWorkspaceExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "vpc_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "vpc_configuration.0.security_group_ids.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "vpc_configuration.0.subnet_ids.#", "3"),
				),
			},
			{
				Config: testAccWorkspaceConfig_vpcRemoved(rName, 3),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckWorkspaceExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "vpc_configuration.#", "0"),
				),
			},
		},
	})
}

func testAccWorkspace_sso(t *testing.T) {
	ctx := acctest.Context(t)
	var v managedgrafana.WorkspaceDescription
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_grafana_workspace.test"
	iamRoleResourceName := "aws_iam_role.test"

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, managedgrafana.EndpointsID)
			acctest.PreCheckSSOAdminInstances(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, managedgrafana.EndpointsID),
		CheckDestroy:             testAccCheckWorkspaceDestroy(ctx),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccWorkspaceConfig_authenticationProvider(rName, "AWS_SSO"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckWorkspaceExists(ctx, resourceName, &v),
					acctest.MatchResourceAttrRegionalARN(resourceName, "arn", "grafana", regexp.MustCompile(`/workspaces/.+`)),
					resource.TestCheckResourceAttr(resourceName, "account_access_type", managedgrafana.AccountAccessTypeCurrentAccount),
					resource.TestCheckResourceAttr(resourceName, "authentication_providers.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "authentication_providers.0", managedgrafana.AuthenticationProviderTypesAwsSso),
					resource.TestCheckResourceAttr(resourceName, "data_sources.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "description", ""),
					resource.TestCheckResourceAttrSet(resourceName, "endpoint"),
					resource.TestCheckResourceAttrSet(resourceName, "grafana_version"),
					resource.TestCheckResourceAttrSet(resourceName, "name"),
					resource.TestCheckResourceAttr(resourceName, "notification_destinations.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "organization_role_name", ""),
					resource.TestCheckResourceAttr(resourceName, "organizational_units.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "permission_type", managedgrafana.PermissionTypeServiceManaged),
					resource.TestCheckResourceAttrPair(resourceName, "role_arn", iamRoleResourceName, "arn"),
					resource.TestCheckResourceAttr(resourceName, "saml_configuration_status", ""),
					resource.TestCheckResourceAttr(resourceName, "stack_set_name", ""),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccWorkspace_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	var v managedgrafana.WorkspaceDescription
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_grafana_workspace.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckPartitionHasService(t, managedgrafana.EndpointsID) },
		ErrorCheck:               acctest.ErrorCheck(t, managedgrafana.EndpointsID),
		CheckDestroy:             testAccCheckWorkspaceDestroy(ctx),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccWorkspaceConfig_authenticationProvider(rName, "SAML"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckWorkspaceExists(ctx, resourceName, &v),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfgrafana.ResourceWorkspace(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccWorkspace_organization(t *testing.T) {
	ctx := acctest.Context(t)
	var v managedgrafana.WorkspaceDescription
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_grafana_workspace.test"

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, managedgrafana.EndpointsID)
			acctest.PreCheckOrganizationsEnabled(ctx, t)
			acctest.PreCheckOrganizationManagementAccount(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, managedgrafana.EndpointsID),
		CheckDestroy:             testAccCheckWorkspaceDestroy(ctx),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccWorkspaceConfig_organization(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckWorkspaceExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "account_access_type", managedgrafana.AccountAccessTypeOrganization),
					resource.TestCheckResourceAttr(resourceName, "authentication_providers.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "authentication_providers.0", managedgrafana.AuthenticationProviderTypesSaml),
					resource.TestCheckResourceAttr(resourceName, "organization_role_name", ""),
					resource.TestCheckResourceAttr(resourceName, "organizational_units.#", "1"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccWorkspace_tags(t *testing.T) {
	ctx := acctest.Context(t)
	var v managedgrafana.WorkspaceDescription
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_grafana_workspace.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckPartitionHasService(t, managedgrafana.EndpointsID) },
		ErrorCheck:               acctest.ErrorCheck(t, managedgrafana.EndpointsID),
		CheckDestroy:             testAccCheckWorkspaceDestroy(ctx),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccWorkspaceConfig_tags1(rName, "key1", "value1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckWorkspaceExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccWorkspaceConfig_tags2(rName, "key1", "value1updated", "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckWorkspaceExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1updated"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
			{
				Config: testAccWorkspaceConfig_tags1(rName, "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckWorkspaceExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
		},
	})
}

func testAccWorkspace_dataSources(t *testing.T) {
	ctx := acctest.Context(t)
	var v managedgrafana.WorkspaceDescription
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_grafana_workspace.test"
	iamRoleResourceName := "aws_iam_role.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckPartitionHasService(t, managedgrafana.EndpointsID) },
		ErrorCheck:               acctest.ErrorCheck(t, managedgrafana.EndpointsID),
		CheckDestroy:             testAccCheckWorkspaceDestroy(ctx),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccWorkspaceConfig_dataSources(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckWorkspaceExists(ctx, resourceName, &v),
					acctest.MatchResourceAttrRegionalARN(resourceName, "arn", "grafana", regexp.MustCompile(`/workspaces/.+`)),
					resource.TestCheckResourceAttr(resourceName, "account_access_type", managedgrafana.AccountAccessTypeCurrentAccount),
					resource.TestCheckResourceAttr(resourceName, "authentication_providers.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "authentication_providers.0", managedgrafana.AuthenticationProviderTypesSaml),
					resource.TestCheckResourceAttr(resourceName, "data_sources.#", "3"),
					resource.TestCheckResourceAttr(resourceName, "data_sources.0", "CLOUDWATCH"),
					resource.TestCheckResourceAttr(resourceName, "data_sources.1", "PROMETHEUS"),
					resource.TestCheckResourceAttr(resourceName, "data_sources.2", "XRAY"),
					resource.TestCheckResourceAttr(resourceName, "description", rName),
					resource.TestCheckResourceAttrSet(resourceName, "endpoint"),
					resource.TestCheckResourceAttrSet(resourceName, "grafana_version"),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "notification_destinations.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "organization_role_name", ""),
					resource.TestCheckResourceAttr(resourceName, "organizational_units.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "permission_type", managedgrafana.PermissionTypeServiceManaged),
					resource.TestCheckResourceAttrPair(resourceName, "role_arn", iamRoleResourceName, "arn"),
					resource.TestCheckResourceAttr(resourceName, "saml_configuration_status", managedgrafana.SamlConfigurationStatusNotConfigured),
					resource.TestCheckResourceAttr(resourceName, "stack_set_name", ""),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccWorkspace_permissionType(t *testing.T) {
	ctx := acctest.Context(t)
	var v managedgrafana.WorkspaceDescription
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_grafana_workspace.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckPartitionHasService(t, managedgrafana.EndpointsID) },
		ErrorCheck:               acctest.ErrorCheck(t, managedgrafana.EndpointsID),
		CheckDestroy:             testAccCheckWorkspaceDestroy(ctx),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccWorkspaceConfig_permissionType(rName, "CUSTOMER_MANAGED"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckWorkspaceExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "permission_type", managedgrafana.PermissionTypeCustomerManaged),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccWorkspaceConfig_permissionType(rName, "SERVICE_MANAGED"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckWorkspaceExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "permission_type", managedgrafana.PermissionTypeServiceManaged),
				),
			},
		},
	})
}

func testAccWorkspace_notificationDestinations(t *testing.T) {
	ctx := acctest.Context(t)
	var v managedgrafana.WorkspaceDescription
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_grafana_workspace.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckPartitionHasService(t, managedgrafana.EndpointsID) },
		ErrorCheck:               acctest.ErrorCheck(t, managedgrafana.EndpointsID),
		CheckDestroy:             testAccCheckWorkspaceDestroy(ctx),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccWorkspaceConfig_authenticationProvider(rName, "SAML"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckWorkspaceExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "description", ""),
					resource.TestCheckResourceAttrSet(resourceName, "name"),
					resource.TestCheckResourceAttr(resourceName, "notification_destinations.#", "0"),
				),
			},
			{
				Config: testAccWorkspaceConfig_notificationDestinations(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckWorkspaceExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "description", rName),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "notification_destinations.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "notification_destinations.0", "SNS"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccWorkspace_configuration(t *testing.T) {
	ctx := acctest.Context(t)
	var v managedgrafana.WorkspaceDescription
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_grafana_workspace.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckPartitionHasService(t, managedgrafana.EndpointsID) },
		ErrorCheck:               acctest.ErrorCheck(t, managedgrafana.EndpointsID),
		CheckDestroy:             testAccCheckWorkspaceDestroy(ctx),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccWorkspaceConfig_configuration(rName, `{"unifiedAlerting": { "enabled": true }}`),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckWorkspaceExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "configuration", `{"unifiedAlerting":{"enabled":true}}`),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccWorkspaceConfig_configuration(rName, `{"unifiedAlerting": { "enabled": false }}`),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckWorkspaceExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "configuration", `{"unifiedAlerting":{"enabled":false}}`),
				),
			},
		},
	})
}

func testAccWorkspace_networkAccess(t *testing.T) {
	ctx := acctest.Context(t)
	var v managedgrafana.WorkspaceDescription
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_grafana_workspace.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckPartitionHasService(t, managedgrafana.EndpointsID) },
		ErrorCheck:               acctest.ErrorCheck(t, managedgrafana.EndpointsID),
		CheckDestroy:             testAccCheckWorkspaceDestroy(ctx),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccWorkspaceConfig_networkAccess(rName, 1),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckWorkspaceExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "network_access_control.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "network_access_control.0.prefix_list_ids.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "network_access_control.0.vpce_ids.#", "1"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccWorkspaceConfig_networkAccess(rName, 2),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckWorkspaceExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "network_access_control.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "network_access_control.0.prefix_list_ids.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "network_access_control.0.vpce_ids.#", "2"),
				),
			},
			{
				Config: testAccWorkspaceConfig_networkAccessRemoved(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckWorkspaceExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "network_access_control.#", "0"),
				),
			},
		},
	})
}

func testAccWorkspace_version(t *testing.T) {
	ctx := acctest.Context(t)
	var v1, v2 managedgrafana.WorkspaceDescription
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_grafana_workspace.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckPartitionHasService(t, managedgrafana.EndpointsID) },
		ErrorCheck:               acctest.ErrorCheck(t, managedgrafana.EndpointsID),
		CheckDestroy:             testAccCheckWorkspaceDestroy(ctx),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccWorkspaceConfig_version(rName, "8.4"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckWorkspaceExists(ctx, resourceName, &v1),
					resource.TestCheckResourceAttr(resourceName, "grafana_version", "8.4"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccWorkspaceConfig_version(rName, "9.4"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckWorkspaceExists(ctx, resourceName, &v2),
					resource.TestCheckResourceAttr(resourceName, "grafana_version", "9.4"),
					testAccCheckWorkspaceNotRecreated(&v2, &v1),
				),
			},
		},
	})
}

func testAccCheckWorkspaceExists(ctx context.Context, n string, v *managedgrafana.WorkspaceDescription) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No Grafana Workspace ID is set")
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).GrafanaConn(ctx)

		output, err := tfgrafana.FindWorkspaceByID(ctx, conn, rs.Primary.ID)

		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

func testAccCheckWorkspaceDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).GrafanaConn(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_grafana_workspace" {
				continue
			}

			_, err := tfgrafana.FindWorkspaceByID(ctx, conn, rs.Primary.ID)

			if tfresource.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("Grafana Workspace %s still exists", rs.Primary.ID)
		}
		return nil
	}
}

func testAccCheckWorkspaceNotRecreated(i, j *managedgrafana.WorkspaceDescription) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if aws.StringValue(i.Id) != aws.StringValue(j.Id) {
			return errors.New("Grafana Workspace was recreated")
		}

		return nil
	}
}

func testAccWorkspaceConfig_base(rName string) string {
	return fmt.Sprintf(`
resource "aws_iam_role" "test" {
  name = %[1]q

  assume_role_policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Action = "sts:AssumeRole"
        Effect = "Allow"
        Sid    = ""
        Principal = {
          Service = "grafana.amazonaws.com"
        }
      },
    ]
  })
}
`, rName)
}

func testAccWorkspaceConfig_authenticationProvider(rName, authenticationProvider string) string {
	return acctest.ConfigCompose(testAccWorkspaceConfig_base(rName), fmt.Sprintf(`
resource "aws_grafana_workspace" "test" {
  account_access_type      = "CURRENT_ACCOUNT"
  authentication_providers = [%[1]q]
  permission_type          = "SERVICE_MANAGED"
  role_arn                 = aws_iam_role.test.arn
}
`, authenticationProvider))
}

func testAccWorkspaceConfig_organization(rName string) string {
	return acctest.ConfigCompose(testAccWorkspaceConfig_base(rName), fmt.Sprintf(`
resource "aws_grafana_workspace" "test" {
  account_access_type      = "ORGANIZATION"
  authentication_providers = ["SAML"]
  permission_type          = "SERVICE_MANAGED"
  role_arn                 = aws_iam_role.test.arn
  organizational_units     = [aws_organizations_organizational_unit.test.id]
}

data "aws_organizations_organization" "test" {}

resource "aws_organizations_organizational_unit" "test" {
  name      = %[1]q
  parent_id = data.aws_organizations_organization.test.roots[0].id
}
`, rName))
}

func testAccWorkspaceConfig_tags1(rName, tagKey1, tagValue1 string) string {
	return acctest.ConfigCompose(testAccWorkspaceConfig_base(rName), fmt.Sprintf(`
resource "aws_grafana_workspace" "test" {
  account_access_type      = "CURRENT_ACCOUNT"
  authentication_providers = ["SAML"]
  permission_type          = "SERVICE_MANAGED"
  name                     = %[1]q
  role_arn                 = aws_iam_role.test.arn

  tags = {
    %[2]q = %[3]q
  }

}
  `, rName, tagKey1, tagValue1))
}

func testAccWorkspaceConfig_tags2(rName, tagKey1, tagValue1, tagKey2, tagValue2 string) string {
	return acctest.ConfigCompose(testAccWorkspaceConfig_base(rName), fmt.Sprintf(`
resource "aws_grafana_workspace" "test" {
  account_access_type      = "CURRENT_ACCOUNT"
  authentication_providers = ["SAML"]
  permission_type          = "SERVICE_MANAGED"
  name                     = %[1]q
  role_arn                 = aws_iam_role.test.arn

  tags = {
    %[2]q = %[3]q
    %[4]q = %[5]q
  }

}
  `, rName, tagKey1, tagValue1, tagKey2, tagValue2))
}

func testAccWorkspaceConfig_dataSources(rName string) string {
	return acctest.ConfigCompose(testAccWorkspaceConfig_base(rName), fmt.Sprintf(`
resource "aws_grafana_workspace" "test" {
  account_access_type      = "CURRENT_ACCOUNT"
  authentication_providers = ["SAML"]
  permission_type          = "SERVICE_MANAGED"
  name                     = %[1]q
  description              = %[1]q
  data_sources             = ["CLOUDWATCH", "PROMETHEUS", "XRAY"]
  role_arn                 = aws_iam_role.test.arn
}
`, rName))
}

func testAccWorkspaceConfig_permissionType(rName, permissionType string) string {
	return acctest.ConfigCompose(testAccWorkspaceConfig_base(rName), fmt.Sprintf(`
resource "aws_grafana_workspace" "test" {
  account_access_type      = "CURRENT_ACCOUNT"
  authentication_providers = ["SAML"]
  permission_type          = %[1]q
  role_arn                 = aws_iam_role.test.arn
}
`, permissionType))
}

func testAccWorkspaceConfig_notificationDestinations(rName string) string {
	return acctest.ConfigCompose(testAccWorkspaceConfig_base(rName), fmt.Sprintf(`
resource "aws_grafana_workspace" "test" {
  account_access_type       = "CURRENT_ACCOUNT"
  authentication_providers  = ["SAML"]
  permission_type           = "SERVICE_MANAGED"
  name                      = %[1]q
  description               = %[1]q
  notification_destinations = ["SNS"]
  role_arn                  = aws_iam_role.test.arn
}
`, rName))
}

func testAccWorkspaceConfig_networkAccess(rName string, endpoints int) string {
	return acctest.ConfigCompose(testAccWorkspaceConfig_base(rName), acctest.ConfigVPCWithSubnets(rName, endpoints), fmt.Sprintf(`
resource "aws_ec2_managed_prefix_list" "test" {
  name           = %[1]q
  address_family = "IPv4"
  max_entries    = 5
}

resource "aws_security_group" "test" {
  description = %[1]q
  vpc_id      = aws_vpc.test.id
}

data "aws_region" "current" {}

resource "aws_vpc_endpoint" "test" {
  count = %[2]d

  private_dns_enabled = false
  security_group_ids  = [aws_security_group.test.id]
  service_name        = "com.amazonaws.${data.aws_region.current.name}.grafana-workspace"
  subnet_ids          = [aws_subnet.test[count.index].id]
  vpc_endpoint_type   = "Interface"
  vpc_id              = aws_vpc.test.id
}

resource "aws_grafana_workspace" "test" {
  account_access_type      = "CURRENT_ACCOUNT"
  authentication_providers = ["SAML"]
  permission_type          = "SERVICE_MANAGED"
  name                     = %[1]q
  description              = %[1]q
  role_arn                 = aws_iam_role.test.arn

  network_access_control {
    prefix_list_ids = [aws_ec2_managed_prefix_list.test.id]
    vpce_ids        = aws_vpc_endpoint.test[*].id
  }
}
`, rName, endpoints))
}

func testAccWorkspaceConfig_networkAccessRemoved(rName string) string {
	return acctest.ConfigCompose(testAccWorkspaceConfig_base(rName), acctest.ConfigVPCWithSubnets(rName, 1), fmt.Sprintf(`
resource "aws_ec2_managed_prefix_list" "test" {
  name           = %[1]q
  address_family = "IPv4"
  max_entries    = 5
}

resource "aws_security_group" "test" {
  description = %[1]q
  vpc_id      = aws_vpc.test.id
}

data "aws_region" "current" {}

resource "aws_vpc_endpoint" "test" {
  private_dns_enabled = false
  security_group_ids  = [aws_security_group.test.id]
  service_name        = "com.amazonaws.${data.aws_region.current.name}.grafana-workspace"
  subnet_ids          = aws_subnet.test[*].id
  vpc_endpoint_type   = "Interface"
  vpc_id              = aws_vpc.test.id
}

resource "aws_grafana_workspace" "test" {
  account_access_type      = "CURRENT_ACCOUNT"
  authentication_providers = ["SAML"]
  permission_type          = "SERVICE_MANAGED"
  name                     = %[1]q
  description              = %[1]q
  role_arn                 = aws_iam_role.test.arn
}
`, rName))
}

func testAccWorkspaceConfig_vpc(rName string, subnets int) string {
	return acctest.ConfigCompose(testAccWorkspaceConfig_base(rName), acctest.ConfigVPCWithSubnets(rName, subnets), fmt.Sprintf(`
resource "aws_security_group" "test" {
  description = %[1]q
  vpc_id      = aws_vpc.test.id
}

resource "aws_grafana_workspace" "test" {
  account_access_type      = "CURRENT_ACCOUNT"
  authentication_providers = ["SAML"]
  permission_type          = "SERVICE_MANAGED"
  role_arn                 = aws_iam_role.test.arn

  vpc_configuration {
    subnet_ids         = aws_subnet.test[*].id
    security_group_ids = [aws_security_group.test.id]
  }
}
`, rName))
}

func testAccWorkspaceConfig_vpcRemoved(rName string, subnets int) string {
	return acctest.ConfigCompose(testAccWorkspaceConfig_base(rName), acctest.ConfigVPCWithSubnets(rName, subnets), fmt.Sprintf(`
resource "aws_security_group" "test" {
  description = %[1]q
  vpc_id      = aws_vpc.test.id
}

resource "aws_grafana_workspace" "test" {
  account_access_type      = "CURRENT_ACCOUNT"
  authentication_providers = ["SAML"]
  permission_type          = "SERVICE_MANAGED"
  role_arn                 = aws_iam_role.test.arn
}
`, rName))
}

func testAccWorkspaceConfig_configuration(rName, configuration string) string {
	return acctest.ConfigCompose(testAccWorkspaceConfig_base(rName), fmt.Sprintf(`
resource "aws_grafana_workspace" "test" {
  account_access_type      = "CURRENT_ACCOUNT"
  authentication_providers = ["SAML"]
  permission_type          = "SERVICE_MANAGED"
  role_arn                 = aws_iam_role.test.arn
  configuration            = %[1]q
}
`, configuration))
}

func testAccWorkspaceConfig_version(rName, version string) string {
	return acctest.ConfigCompose(testAccWorkspaceConfig_base(rName), fmt.Sprintf(`
resource "aws_grafana_workspace" "test" {
  account_access_type      = "CURRENT_ACCOUNT"
  authentication_providers = ["SAML"]
  permission_type          = "SERVICE_MANAGED"
  role_arn                 = aws_iam_role.test.arn
  grafana_version          = %[1]q
}
`, version))
}
