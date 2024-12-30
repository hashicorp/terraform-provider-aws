// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package workspaces_test

import (
	"context"
	"fmt"
	"strings"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/workspaces"
	"github.com/aws/aws-sdk-go-v2/service/workspaces/types"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfworkspaces "github.com/hashicorp/terraform-provider-aws/internal/service/workspaces"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func testAccDirectory_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var v types.WorkspaceDirectory
	rName := sdkacctest.RandString(8)

	resourceName := "aws_workspaces_directory.main"
	directoryResourceName := "aws_directory_service_directory.main"
	iamRoleDataSourceName := "data.aws_iam_role.workspaces-default"

	domain := acctest.RandomDomainName()

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheckDirectory(ctx, t)
			acctest.PreCheckDirectoryServiceSimpleDirectory(ctx, t)
			acctest.PreCheckHasIAMRole(ctx, t, "workspaces_DefaultRole")
		},
		ErrorCheck:               acctest.ErrorCheck(t, strings.ToLower(workspaces.ServiceID)),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDirectoryDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccDirectoryConfig_basic(rName, domain),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckDirectoryExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrAlias, directoryResourceName, names.AttrAlias),
					resource.TestCheckResourceAttrPair(resourceName, "directory_id", directoryResourceName, names.AttrID),
					resource.TestCheckResourceAttrPair(resourceName, "directory_name", directoryResourceName, names.AttrName),
					resource.TestCheckResourceAttr(resourceName, "directory_type", string(types.WorkspaceDirectoryTypeSimpleAd)),
					resource.TestCheckResourceAttr(resourceName, "dns_ip_addresses.#", "2"),
					resource.TestCheckResourceAttrPair(resourceName, "iam_role_id", iamRoleDataSourceName, names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, "ip_group_ids.#", "0"),
					resource.TestCheckResourceAttrSet(resourceName, "registration_code"),
					resource.TestCheckResourceAttr(resourceName, "saml_properties.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "saml_properties.0.relay_state_parameter_name", "RelayState"),
					resource.TestCheckResourceAttr(resourceName, "saml_properties.0.status", "DISABLED"),
					resource.TestCheckResourceAttr(resourceName, "self_service_permissions.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "self_service_permissions.0.change_compute_type", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "self_service_permissions.0.increase_volume_size", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "self_service_permissions.0.rebuild_workspace", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "self_service_permissions.0.restart_workspace", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "self_service_permissions.0.switch_running_mode", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "subnet_ids.#", "2"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.Name", fmt.Sprintf("tf-testacc-workspaces-directory-%[1]s", rName)),
					resource.TestCheckResourceAttr(resourceName, "workspace_access_properties.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "workspace_access_properties.0.device_type_android", "ALLOW"),
					resource.TestCheckResourceAttr(resourceName, "workspace_access_properties.0.device_type_chromeos", "ALLOW"),
					resource.TestCheckResourceAttr(resourceName, "workspace_access_properties.0.device_type_ios", "ALLOW"),
					resource.TestCheckResourceAttr(resourceName, "workspace_access_properties.0.device_type_linux", "DENY"),
					resource.TestCheckResourceAttr(resourceName, "workspace_access_properties.0.device_type_osx", "ALLOW"),
					resource.TestCheckResourceAttr(resourceName, "workspace_access_properties.0.device_type_web", "DENY"),
					resource.TestCheckResourceAttr(resourceName, "workspace_access_properties.0.device_type_windows", "ALLOW"),
					resource.TestCheckResourceAttr(resourceName, "workspace_access_properties.0.device_type_zeroclient", "ALLOW"),
					resource.TestCheckResourceAttr(resourceName, "workspace_creation_properties.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "workspace_creation_properties.0.custom_security_group_id", ""),
					resource.TestCheckResourceAttr(resourceName, "workspace_creation_properties.0.default_ou", ""),
					resource.TestCheckResourceAttr(resourceName, "workspace_creation_properties.0.enable_internet_access", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "workspace_creation_properties.0.enable_maintenance_mode", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "workspace_creation_properties.0.user_enabled_as_local_administrator", acctest.CtTrue),
					resource.TestCheckResourceAttrSet(resourceName, "workspace_security_group_id"),
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

func testAccDirectory_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	var v types.WorkspaceDirectory
	rName := sdkacctest.RandString(8)

	resourceName := "aws_workspaces_directory.main"

	domain := acctest.RandomDomainName()

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheckDirectory(ctx, t)
			acctest.PreCheckDirectoryServiceSimpleDirectory(ctx, t)
			acctest.PreCheckHasIAMRole(ctx, t, "workspaces_DefaultRole")
		},
		ErrorCheck:               acctest.ErrorCheck(t, strings.ToLower(workspaces.ServiceID)),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDirectoryDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccDirectoryConfig_basic(rName, domain),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckDirectoryExists(ctx, resourceName, &v),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfworkspaces.ResourceDirectory(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccDirectory_subnetIDs(t *testing.T) {
	ctx := acctest.Context(t)
	var v types.WorkspaceDirectory
	rName := sdkacctest.RandString(8)

	resourceName := "aws_workspaces_directory.main"

	domain := acctest.RandomDomainName()

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheckDirectory(ctx, t)
			acctest.PreCheckDirectoryServiceSimpleDirectory(ctx, t)
			acctest.PreCheckHasIAMRole(ctx, t, "workspaces_DefaultRole")
		},
		ErrorCheck:               acctest.ErrorCheck(t, strings.ToLower(workspaces.ServiceID)),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDirectoryDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccDirectoryConfig_subnetIDs(rName, domain),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDirectoryExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "subnet_ids.#", "2"),
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

func testAccDirectory_tags(t *testing.T) {
	ctx := acctest.Context(t)
	var v types.WorkspaceDirectory
	rName := sdkacctest.RandString(8)

	resourceName := "aws_workspaces_directory.main"

	domain := acctest.RandomDomainName()

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheckDirectory(ctx, t)
			acctest.PreCheckDirectoryServiceSimpleDirectory(ctx, t)
			acctest.PreCheckHasIAMRole(ctx, t, "workspaces_DefaultRole")
		},
		ErrorCheck:               acctest.ErrorCheck(t, strings.ToLower(workspaces.ServiceID)),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDirectoryDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccDirectoryConfig_tags1(rName, domain, acctest.CtKey1, acctest.CtValue1),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckDirectoryExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, "1"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey1, acctest.CtValue1),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccDirectoryConfig_tags2(rName, domain, acctest.CtKey1, acctest.CtValue1Updated, acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckDirectoryExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, "2"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey1, acctest.CtValue1Updated),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey2, acctest.CtValue2),
				),
			},
			{
				Config: testAccDirectoryConfig_tags1(rName, domain, acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckDirectoryExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, "1"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey2, acctest.CtValue2),
				),
			},
		},
	})
}

func testAccDirectory_SamlProperties(t *testing.T) {
	ctx := acctest.Context(t)
	var v types.WorkspaceDirectory
	rName := sdkacctest.RandString(8)

	resourceName := "aws_workspaces_directory.main"

	domain := acctest.RandomDomainName()
	rspn := sdkacctest.RandString(8)
	arspn := sdkacctest.RandString(8)
	uau := fmt.Sprintf("https://%s/", acctest.RandomDomainName())
	auau := fmt.Sprintf("https://%s/", acctest.RandomDomainName())

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheckDirectory(ctx, t)
			acctest.PreCheckDirectoryServiceSimpleDirectory(ctx, t)
			acctest.PreCheckHasIAMRole(ctx, t, "workspaces_DefaultRole")
		},
		ErrorCheck:               acctest.ErrorCheck(t, strings.ToLower(workspaces.ServiceID)),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDirectoryDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccDirectoryConfig_samlPropertiesFull(rName, domain, rspn, uau),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckDirectoryExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "saml_properties.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "saml_properties.0.relay_state_parameter_name", rspn),
					resource.TestCheckResourceAttr(resourceName, "saml_properties.0.user_access_url", uau),
					resource.TestCheckResourceAttr(resourceName, "saml_properties.0.status", "ENABLED"),
				),
			},
			{
				Config: testAccDirectoryConfig_samlPropertiesRSPN(rName, domain, arspn),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckDirectoryExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "saml_properties.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "saml_properties.0.relay_state_parameter_name", arspn),
					resource.TestCheckResourceAttr(resourceName, "saml_properties.0.user_access_url", ""),
					resource.TestCheckResourceAttr(resourceName, "saml_properties.0.status", "DISABLED"),
				),
			},
			{
				Config: testAccDirectoryConfig_samlPropertiesUAU(rName, domain, auau),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckDirectoryExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "saml_properties.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "saml_properties.0.relay_state_parameter_name", "RelayState"),
					resource.TestCheckResourceAttr(resourceName, "saml_properties.0.user_access_url", auau),
					resource.TestCheckResourceAttr(resourceName, "saml_properties.0.status", "ENABLED_WITH_DIRECTORY_LOGIN_FALLBACK"),
				),
			},
			{
				Config: testAccDirectoryConfig_samlPropertiesFull(rName, domain, rspn, uau),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckDirectoryExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "saml_properties.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "saml_properties.0.relay_state_parameter_name", rspn),
					resource.TestCheckResourceAttr(resourceName, "saml_properties.0.user_access_url", uau),
					resource.TestCheckResourceAttr(resourceName, "saml_properties.0.status", "ENABLED"),
				),
			},
			{
				Config: testAccDirectoryConfig_samlPropertiesEmpty(rName, domain),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckDirectoryExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "saml_properties.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "saml_properties.0.relay_state_parameter_name", "RelayState"),
					resource.TestCheckResourceAttr(resourceName, "saml_properties.0.user_access_url", ""),
					resource.TestCheckResourceAttr(resourceName, "saml_properties.0.status", "DISABLED"),
				),
			},
		},
	})
}

func testAccDirectory_selfServicePermissions(t *testing.T) {
	ctx := acctest.Context(t)
	var v types.WorkspaceDirectory
	rName := sdkacctest.RandString(8)

	resourceName := "aws_workspaces_directory.main"

	domain := acctest.RandomDomainName()

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheckDirectory(ctx, t)
			acctest.PreCheckDirectoryServiceSimpleDirectory(ctx, t)
			acctest.PreCheckHasIAMRole(ctx, t, "workspaces_DefaultRole")
		},
		ErrorCheck:               acctest.ErrorCheck(t, strings.ToLower(workspaces.ServiceID)),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDirectoryDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccDirectoryConfig_selfServicePermissions(rName, domain),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckDirectoryExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "self_service_permissions.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "self_service_permissions.0.change_compute_type", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "self_service_permissions.0.increase_volume_size", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "self_service_permissions.0.rebuild_workspace", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "self_service_permissions.0.restart_workspace", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "self_service_permissions.0.switch_running_mode", acctest.CtTrue),
				),
			},
		},
	})
}

func testAccDirectory_workspaceAccessProperties(t *testing.T) {
	ctx := acctest.Context(t)
	var v types.WorkspaceDirectory
	rName := sdkacctest.RandString(8)

	resourceName := "aws_workspaces_directory.main"

	domain := acctest.RandomDomainName()

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheckDirectory(ctx, t)
			acctest.PreCheckDirectoryServiceSimpleDirectory(ctx, t)
			acctest.PreCheckHasIAMRole(ctx, t, "workspaces_DefaultRole")
		},
		ErrorCheck:               acctest.ErrorCheck(t, strings.ToLower(workspaces.ServiceID)),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDirectoryDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccDirectoryConfig_workspaceAccessProperties(rName, domain),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckDirectoryExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "workspace_access_properties.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "workspace_access_properties.0.device_type_android", "ALLOW"),
					resource.TestCheckResourceAttr(resourceName, "workspace_access_properties.0.device_type_chromeos", "ALLOW"),
					resource.TestCheckResourceAttr(resourceName, "workspace_access_properties.0.device_type_ios", "ALLOW"),
					resource.TestCheckResourceAttr(resourceName, "workspace_access_properties.0.device_type_linux", "DENY"),
					resource.TestCheckResourceAttr(resourceName, "workspace_access_properties.0.device_type_osx", "ALLOW"),
					resource.TestCheckResourceAttr(resourceName, "workspace_access_properties.0.device_type_web", "DENY"),
					resource.TestCheckResourceAttr(resourceName, "workspace_access_properties.0.device_type_windows", "DENY"),
					resource.TestCheckResourceAttr(resourceName, "workspace_access_properties.0.device_type_zeroclient", "DENY"),
				),
			},
		},
	})
}

func testAccDirectory_workspaceCreationProperties(t *testing.T) {
	ctx := acctest.Context(t)
	var v types.WorkspaceDirectory
	rName := sdkacctest.RandString(8)

	resourceName := "aws_workspaces_directory.main"
	resourceSecurityGroup := "aws_security_group.test"

	domain := acctest.RandomDomainName()

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheckDirectory(ctx, t)
			acctest.PreCheckDirectoryServiceSimpleDirectory(ctx, t)
			acctest.PreCheckHasIAMRole(ctx, t, "workspaces_DefaultRole")
		},
		ErrorCheck:               acctest.ErrorCheck(t, strings.ToLower(workspaces.ServiceID)),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDirectoryDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccDirectoryConfig_workspaceCreationProperties(rName, domain),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckDirectoryExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "workspace_creation_properties.#", "1"),
					resource.TestCheckResourceAttrPair(resourceName, "workspace_creation_properties.0.custom_security_group_id", resourceSecurityGroup, names.AttrID),
					resource.TestCheckResourceAttr(resourceName, "workspace_creation_properties.0.default_ou", "OU=AWS,DC=Workgroup,DC=Example,DC=com"),
					resource.TestCheckResourceAttr(resourceName, "workspace_creation_properties.0.enable_internet_access", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "workspace_creation_properties.0.enable_maintenance_mode", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "workspace_creation_properties.0.user_enabled_as_local_administrator", acctest.CtFalse),
				),
			},
		},
	})
}

func testAccDirectory_workspaceCreationProperties_customSecurityGroupId_defaultOu(t *testing.T) {
	ctx := acctest.Context(t)
	var v types.WorkspaceDirectory
	rName := sdkacctest.RandString(8)

	resourceName := "aws_workspaces_directory.main"
	resourceSecurityGroup := "aws_security_group.test"

	domain := acctest.RandomDomainName()

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheckDirectory(ctx, t)
			acctest.PreCheckDirectoryServiceSimpleDirectory(ctx, t)
			acctest.PreCheckHasIAMRole(ctx, t, "workspaces_DefaultRole")
		},
		ErrorCheck:               acctest.ErrorCheck(t, strings.ToLower(workspaces.ServiceID)),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDirectoryDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccDirectoryConfig_creationPropertiesCustomSecurityGroupIdDefaultOUAbsent(rName, domain),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckDirectoryExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "workspace_creation_properties.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "workspace_creation_properties.0.custom_security_group_id", ""),
					resource.TestCheckResourceAttr(resourceName, "workspace_creation_properties.0.default_ou", ""),
				),
			},
			{
				Config: testAccDirectoryConfig_creationPropertiesCustomSecurityGroupIdDefaultOUPresent(rName, domain),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckDirectoryExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "workspace_creation_properties.#", "1"),
					resource.TestCheckResourceAttrPair(resourceName, "workspace_creation_properties.0.custom_security_group_id", resourceSecurityGroup, names.AttrID),
					resource.TestCheckResourceAttr(resourceName, "workspace_creation_properties.0.default_ou", "OU=AWS,DC=Workgroup,DC=Example,DC=com"),
				),
			},
			{
				Config: testAccDirectoryConfig_creationPropertiesCustomSecurityGroupIdDefaultOUAbsent(rName, domain),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckDirectoryExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "workspace_creation_properties.#", "1"),
					resource.TestCheckResourceAttrSet(resourceName, "workspace_creation_properties.0.custom_security_group_id"),
					resource.TestCheckResourceAttrSet(resourceName, "workspace_creation_properties.0.default_ou"),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccDirectory_ipGroupIDs(t *testing.T) {
	ctx := acctest.Context(t)
	var v types.WorkspaceDirectory
	rName := sdkacctest.RandString(8)

	resourceName := "aws_workspaces_directory.test"

	domain := acctest.RandomDomainName()

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckHasIAMRole(ctx, t, "workspaces_DefaultRole") },
		ErrorCheck:               acctest.ErrorCheck(t, strings.ToLower(workspaces.ServiceID)),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDirectoryDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccDirectoryConfig_ipGroupIdsCreate(rName, domain),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDirectoryExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "ip_group_ids.#", "1"),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "ip_group_ids.*", "aws_workspaces_ip_group.test_alpha", names.AttrID),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccDirectoryConfig_ipGroupIdsUpdate(rName, domain),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDirectoryExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "ip_group_ids.#", "2"),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "ip_group_ids.*", "aws_workspaces_ip_group.test_beta", names.AttrID),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "ip_group_ids.*", "aws_workspaces_ip_group.test_gamma", names.AttrID),
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

func testAccCheckDirectoryDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).WorkSpacesClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_workspaces_directory" {
				continue
			}

			_, err := tfworkspaces.FindDirectoryByID(ctx, conn, rs.Primary.ID)

			if tfresource.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("WorkSpaces Directory %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccCheckDirectoryExists(ctx context.Context, n string, v *types.WorkspaceDirectory) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).WorkSpacesClient(ctx)

		output, err := tfworkspaces.FindDirectoryByID(ctx, conn, rs.Primary.ID)

		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

func testAccPreCheckDirectory(ctx context.Context, t *testing.T) {
	conn := acctest.Provider.Meta().(*conns.AWSClient).WorkSpacesClient(ctx)

	input := &workspaces.DescribeWorkspaceDirectoriesInput{}

	_, err := conn.DescribeWorkspaceDirectories(ctx, input)

	if acctest.PreCheckSkipError(err) {
		t.Skipf("skipping acceptance testing: %s", err)
	}

	if err != nil {
		t.Fatalf("unexpected PreCheck error: %s", err)
	}
}

func testAccDirectoryConfig_Prerequisites(rName, domain string) string {
	return acctest.ConfigCompose(
		acctest.ConfigAvailableAZsNoOptIn(),
		//lintignore:AWSAT003
		fmt.Sprintf(`
data "aws_region" "current" {}

locals {
  region_workspaces_az_ids = {
    "us-east-1" = formatlist("use1-az%%d", [2, 4, 6])
  }

  workspaces_az_ids = lookup(local.region_workspaces_az_ids, data.aws_region.current.name, data.aws_availability_zones.available.zone_ids)
}

resource "aws_vpc" "main" {
  cidr_block = "10.0.0.0/16"

  tags = {
    Name = "tf-testacc-workspaces-directory-%[1]s"
  }
}

resource "aws_subnet" "primary" {
  vpc_id               = aws_vpc.main.id
  availability_zone_id = local.workspaces_az_ids[0]
  cidr_block           = "10.0.1.0/24"

  tags = {
    Name = "tf-testacc-workspaces-directory-%[1]s-primary"
  }
}

resource "aws_subnet" "secondary" {
  vpc_id               = aws_vpc.main.id
  availability_zone_id = local.workspaces_az_ids[1]
  cidr_block           = "10.0.2.0/24"

  tags = {
    Name = "tf-testacc-workspaces-directory-%[1]s-secondary"
  }
}

resource "aws_directory_service_directory" "main" {
  size     = "Small"
  name     = %[2]q
  password = "#S1ncerely"

  vpc_settings {
    vpc_id     = aws_vpc.main.id
    subnet_ids = [aws_subnet.primary.id, aws_subnet.secondary.id]
  }

  tags = {
    Name = "tf-testacc-workspaces-directory-%[1]s"
  }
}
`, rName, domain))
}

func testAccDirectoryConfig_basic(rName, domain string) string {
	return acctest.ConfigCompose(
		testAccDirectoryConfig_Prerequisites(rName, domain),
		fmt.Sprintf(`
resource "aws_workspaces_directory" "main" {
  directory_id = aws_directory_service_directory.main.id

  tags = {
    Name = "tf-testacc-workspaces-directory-%[1]s"
  }
}

data "aws_iam_role" "workspaces-default" {
  name = "workspaces_DefaultRole"
}
`, rName))
}

func testAccDirectoryConfig_samlPropertiesFull(rName, domain, rspn, uau string) string {
	return acctest.ConfigCompose(
		testAccDirectoryConfig_Prerequisites(rName, domain),
		fmt.Sprintf(`
resource "aws_workspaces_directory" "main" {
  directory_id = aws_directory_service_directory.main.id

  saml_properties {
    relay_state_parameter_name = %[2]q
    user_access_url            = %[3]q
    status                     = "ENABLED"
  }

  tags = {
    Name = "tf-testacc-workspaces-directory-%[1]s"
  }
}
`, rName, rspn, uau))
}

func testAccDirectoryConfig_samlPropertiesRSPN(rName, domain, rspn string) string {
	return acctest.ConfigCompose(
		testAccDirectoryConfig_Prerequisites(rName, domain),
		fmt.Sprintf(`
resource "aws_workspaces_directory" "main" {
  directory_id = aws_directory_service_directory.main.id

  saml_properties {
    relay_state_parameter_name = %[2]q
    status                     = "DISABLED"
  }

  tags = {
    Name = "tf-testacc-workspaces-directory-%[1]s"
  }
}
`, rName, rspn))
}

func testAccDirectoryConfig_samlPropertiesUAU(rName, domain, uau string) string {
	return acctest.ConfigCompose(
		testAccDirectoryConfig_Prerequisites(rName, domain),
		fmt.Sprintf(`
resource "aws_workspaces_directory" "main" {
  directory_id = aws_directory_service_directory.main.id

  saml_properties {
    user_access_url = %[2]q
    status          = "ENABLED_WITH_DIRECTORY_LOGIN_FALLBACK"
  }

  tags = {
    Name = "tf-testacc-workspaces-directory-%[1]s"
  }
}
`, rName, uau))
}

func testAccDirectoryConfig_samlPropertiesEmpty(rName, domain string) string {
	return acctest.ConfigCompose(
		testAccDirectoryConfig_Prerequisites(rName, domain),
		fmt.Sprintf(`
resource "aws_workspaces_directory" "main" {
  directory_id = aws_directory_service_directory.main.id

  saml_properties {}

  tags = {
    Name = "tf-testacc-workspaces-directory-%[1]s"
  }
}
`, rName))
}

func testAccDirectoryConfig_selfServicePermissions(rName, domain string) string {
	return acctest.ConfigCompose(
		testAccDirectoryConfig_Prerequisites(rName, domain),
		fmt.Sprintf(`
resource "aws_workspaces_directory" "main" {
  directory_id = aws_directory_service_directory.main.id

  self_service_permissions {
    change_compute_type  = false
    increase_volume_size = true
    rebuild_workspace    = true
    restart_workspace    = false
    switch_running_mode  = true
  }

  tags = {
    Name = "tf-testacc-workspaces-directory-%[1]s"
  }
}
`, rName))
}

func testAccDirectoryConfig_subnetIDs(rName, domain string) string {
	return acctest.ConfigCompose(
		testAccDirectoryConfig_Prerequisites(rName, domain),
		fmt.Sprintf(`
resource "aws_workspaces_directory" "main" {
  directory_id = aws_directory_service_directory.main.id
  subnet_ids   = [aws_subnet.primary.id, aws_subnet.secondary.id]

  tags = {
    Name = "tf-testacc-workspaces-directory-%[1]s"
  }
}
`, rName))
}

func testAccDirectoryConfig_tags1(rName, domain, tagKey1, tagValue1 string) string {
	return acctest.ConfigCompose(
		testAccDirectoryConfig_Prerequisites(rName, domain),
		fmt.Sprintf(`
resource "aws_workspaces_directory" "main" {
  directory_id = aws_directory_service_directory.main.id

  tags = {
    %[1]q = %[2]q
  }
}
`, tagKey1, tagValue1))
}

func testAccDirectoryConfig_tags2(rName, domain, tagKey1, tagValue1, tagKey2, tagValue2 string) string {
	return acctest.ConfigCompose(
		testAccDirectoryConfig_Prerequisites(rName, domain),
		fmt.Sprintf(`
resource "aws_workspaces_directory" "main" {
  directory_id = aws_directory_service_directory.main.id

  tags = {
    %[1]q = %[2]q
    %[3]q = %[4]q
  }
}
`, tagKey1, tagValue1, tagKey2, tagValue2))
}

func testAccDirectoryConfig_workspaceAccessProperties(rName, domain string) string {
	return acctest.ConfigCompose(
		testAccDirectoryConfig_Prerequisites(rName, domain),
		fmt.Sprintf(`
resource "aws_workspaces_directory" "main" {
  directory_id = aws_directory_service_directory.main.id

  workspace_access_properties {
    device_type_android    = "ALLOW"
    device_type_chromeos   = "ALLOW"
    device_type_ios        = "ALLOW"
    device_type_linux      = "DENY"
    device_type_osx        = "ALLOW"
    device_type_web        = "DENY"
    device_type_windows    = "DENY"
    device_type_zeroclient = "DENY"
  }

  tags = {
    Name = "tf-testacc-workspaces-directory-%[1]s"
  }
}
`, rName))
}

func testAccDirectoryConfig_workspaceCreationProperties(rName, domain string) string {
	return acctest.ConfigCompose(
		testAccDirectoryConfig_Prerequisites(rName, domain),
		fmt.Sprintf(`
resource "aws_security_group" "test" {
  name   = "tf-acctest-%[1]s"
  vpc_id = aws_vpc.main.id
}

resource "aws_workspaces_directory" "main" {
  directory_id = aws_directory_service_directory.main.id

  workspace_creation_properties {
    custom_security_group_id            = aws_security_group.test.id
    default_ou                          = "OU=AWS,DC=Workgroup,DC=Example,DC=com"
    enable_internet_access              = true
    enable_maintenance_mode             = false
    user_enabled_as_local_administrator = false
  }

  tags = {
    Name = "tf-testacc-workspaces-directory-%[1]s"
  }
}
`, rName))
}

func testAccDirectoryConfig_creationPropertiesCustomSecurityGroupIdDefaultOUAbsent(rName, domain string) string {
	return acctest.ConfigCompose(
		testAccDirectoryConfig_Prerequisites(rName, domain),
		fmt.Sprintf(`
resource "aws_workspaces_directory" "main" {
  directory_id = aws_directory_service_directory.main.id

  workspace_creation_properties {
    enable_internet_access              = true
    enable_maintenance_mode             = false
    user_enabled_as_local_administrator = false
  }

  tags = {
    Name = "tf-testacc-workspaces-directory-%[1]s"
  }
}
`, rName))
}

func testAccDirectoryConfig_creationPropertiesCustomSecurityGroupIdDefaultOUPresent(rName, domain string) string {
	return acctest.ConfigCompose(
		testAccDirectoryConfig_Prerequisites(rName, domain),
		fmt.Sprintf(`
resource "aws_security_group" "test" {
  vpc_id = aws_vpc.main.id
  name   = "tf-acctest-%[1]s"
}

resource "aws_workspaces_directory" "main" {
  directory_id = aws_directory_service_directory.main.id

  workspace_creation_properties {
    custom_security_group_id            = aws_security_group.test.id
    default_ou                          = "OU=AWS,DC=Workgroup,DC=Example,DC=com"
    enable_internet_access              = true
    enable_maintenance_mode             = false
    user_enabled_as_local_administrator = false
  }

  tags = {
    Name = "tf-testacc-workspaces-directory-%[1]s"
  }
}
`, rName))
}

func testAccDirectoryConfig_ipGroupIdsCreate(rName, domain string) string {
	return acctest.ConfigCompose(
		testAccDirectoryConfig_Prerequisites(rName, domain),
		fmt.Sprintf(`
resource "aws_workspaces_ip_group" "test_alpha" {
  name = "%[1]s-alpha"
}

resource "aws_workspaces_directory" "test" {
  directory_id = aws_directory_service_directory.main.id

  ip_group_ids = [
    aws_workspaces_ip_group.test_alpha.id
  ]

  tags = {
    Name = "tf-testacc-workspaces-directory-%[1]s"
  }
}
`, rName))
}

func testAccDirectoryConfig_ipGroupIdsUpdate(rName, domain string) string {
	return acctest.ConfigCompose(
		testAccDirectoryConfig_Prerequisites(rName, domain),
		fmt.Sprintf(`
resource "aws_workspaces_ip_group" "test_beta" {
  name = "%[1]s-beta"
}

resource "aws_workspaces_ip_group" "test_gamma" {
  name = "%[1]s-gamma"
}

resource "aws_workspaces_directory" "test" {
  directory_id = aws_directory_service_directory.main.id

  ip_group_ids = [
    aws_workspaces_ip_group.test_beta.id,
    aws_workspaces_ip_group.test_gamma.id
  ]

  tags = {
    Name = "tf-testacc-workspaces-directory-%[1]s"
  }
}
`, rName))
}
