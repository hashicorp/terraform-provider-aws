package aws

import (
	"fmt"
	"log"
	"reflect"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/workspaces"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/aws/internal/service/workspaces/finder"
	"github.com/hashicorp/terraform-provider-aws/aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/provider"
)

func init() {
	resource.AddTestSweepers("aws_workspaces_directory", &resource.Sweeper{
		Name:         "aws_workspaces_directory",
		F:            testSweepWorkspacesDirectories,
		Dependencies: []string{"aws_workspaces_workspace", "aws_workspaces_ip_group"},
	})
}

func testSweepWorkspacesDirectories(region string) error {
	client, err := sharedClientForRegion(region)
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}
	conn := client.(*conns.AWSClient).WorkSpacesConn
	input := &workspaces.DescribeWorkspaceDirectoriesInput{}
	sweepResources := make([]*testSweepResource, 0)

	err = conn.DescribeWorkspaceDirectoriesPages(input, func(page *workspaces.DescribeWorkspaceDirectoriesOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, directory := range page.Directories {
			r := ResourceDirectory()
			d := r.Data(nil)
			d.SetId(aws.StringValue(directory.DirectoryId))

			sweepResources = append(sweepResources, NewTestSweepResource(r, d, client))
		}

		return !lastPage
	})

	if testSweepSkipSweepError(err) {
		log.Printf("[WARN] Skipping WorkSpaces Directory sweep for %s: %s", region, err)
		return nil
	}

	if err != nil {
		return fmt.Errorf("error listing WorkSpaces Directories (%s): %w", region, err)
	}

	err = testSweepResourceOrchestrator(sweepResources)

	if err != nil {
		return fmt.Errorf("error sweeping WorkSpaces Directories (%s): %w", region, err)
	}

	return nil
}

func testAccAwsWorkspacesDirectory_basic(t *testing.T) {
	var v workspaces.WorkspaceDirectory
	rName := sdkacctest.RandString(8)

	resourceName := "aws_workspaces_directory.main"
	directoryResourceName := "aws_directory_service_directory.main"
	iamRoleDataSourceName := "data.aws_iam_role.workspaces-default"

	domain := acctest.RandomDomainName()

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(t)
			testAccPreCheckWorkspacesDirectory(t)
			acctest.PreCheckDirectoryServiceSimpleDirectory(t)
			acctest.PreCheckHasIAMRole(t, "workspaces_DefaultRole")
		},
		ErrorCheck:   acctest.ErrorCheck(t, workspaces.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckAwsWorkspacesDirectoryDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccWorkspacesDirectoryConfig(rName, domain),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAwsWorkspacesDirectoryExists(resourceName, &v),
					resource.TestCheckResourceAttrPair(resourceName, "alias", directoryResourceName, "alias"),
					resource.TestCheckResourceAttrPair(resourceName, "directory_id", directoryResourceName, "id"),
					resource.TestCheckResourceAttrPair(resourceName, "directory_name", directoryResourceName, "name"),
					resource.TestCheckResourceAttr(resourceName, "directory_type", workspaces.WorkspaceDirectoryTypeSimpleAd),
					resource.TestCheckResourceAttr(resourceName, "dns_ip_addresses.#", "2"),
					resource.TestCheckResourceAttrPair(resourceName, "iam_role_id", iamRoleDataSourceName, "arn"),
					resource.TestCheckResourceAttr(resourceName, "ip_group_ids.#", "0"),
					resource.TestCheckResourceAttrSet(resourceName, "registration_code"),
					resource.TestCheckResourceAttr(resourceName, "self_service_permissions.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "self_service_permissions.0.change_compute_type", "false"),
					resource.TestCheckResourceAttr(resourceName, "self_service_permissions.0.increase_volume_size", "false"),
					resource.TestCheckResourceAttr(resourceName, "self_service_permissions.0.rebuild_workspace", "false"),
					resource.TestCheckResourceAttr(resourceName, "self_service_permissions.0.restart_workspace", "true"),
					resource.TestCheckResourceAttr(resourceName, "self_service_permissions.0.switch_running_mode", "false"),
					resource.TestCheckResourceAttr(resourceName, "subnet_ids.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
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
					resource.TestCheckResourceAttr(resourceName, "workspace_creation_properties.0.enable_internet_access", "false"),
					resource.TestCheckResourceAttr(resourceName, "workspace_creation_properties.0.enable_maintenance_mode", "true"),
					resource.TestCheckResourceAttr(resourceName, "workspace_creation_properties.0.user_enabled_as_local_administrator", "true"),
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

func testAccAwsWorkspacesDirectory_disappears(t *testing.T) {
	var v workspaces.WorkspaceDirectory
	rName := sdkacctest.RandString(8)

	resourceName := "aws_workspaces_directory.main"

	domain := acctest.RandomDomainName()

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(t)
			testAccPreCheckWorkspacesDirectory(t)
			acctest.PreCheckDirectoryServiceSimpleDirectory(t)
			acctest.PreCheckHasIAMRole(t, "workspaces_DefaultRole")
		},
		ErrorCheck:   acctest.ErrorCheck(t, workspaces.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckAwsWorkspacesDirectoryDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccWorkspacesDirectoryConfig(rName, domain),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAwsWorkspacesDirectoryExists(resourceName, &v),
					acctest.CheckResourceDisappears(acctest.Provider, ResourceDirectory(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccAwsWorkspacesDirectory_subnetIds(t *testing.T) {
	var v workspaces.WorkspaceDirectory
	rName := sdkacctest.RandString(8)

	resourceName := "aws_workspaces_directory.main"

	domain := acctest.RandomDomainName()

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(t)
			testAccPreCheckWorkspacesDirectory(t)
			acctest.PreCheckDirectoryServiceSimpleDirectory(t)
			acctest.PreCheckHasIAMRole(t, "workspaces_DefaultRole")
		},
		ErrorCheck:   acctest.ErrorCheck(t, workspaces.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckAwsWorkspacesDirectoryDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccWorkspacesDirectoryConfig_subnetIds(rName, domain),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsWorkspacesDirectoryExists(resourceName, &v),
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

func testAccAwsWorkspacesDirectory_tags(t *testing.T) {
	var v workspaces.WorkspaceDirectory
	rName := sdkacctest.RandString(8)

	resourceName := "aws_workspaces_directory.main"

	domain := acctest.RandomDomainName()

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(t)
			testAccPreCheckWorkspacesDirectory(t)
			acctest.PreCheckDirectoryServiceSimpleDirectory(t)
			acctest.PreCheckHasIAMRole(t, "workspaces_DefaultRole")
		},
		ErrorCheck:   acctest.ErrorCheck(t, workspaces.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckAwsWorkspacesDirectoryDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccWorkspacesDirectoryConfigTags1(rName, domain, "key1", "value1"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAwsWorkspacesDirectoryExists(resourceName, &v),
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
				Config: testAccWorkspacesDirectoryConfigTags2(rName, domain, "key1", "value1updated", "key2", "value2"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAwsWorkspacesDirectoryExists(resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1updated"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
			{
				Config: testAccWorkspacesDirectoryConfigTags1(rName, domain, "key2", "value2"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAwsWorkspacesDirectoryExists(resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
		},
	})
}

func testAccAwsWorkspacesDirectory_selfServicePermissions(t *testing.T) {
	var v workspaces.WorkspaceDirectory
	rName := sdkacctest.RandString(8)

	resourceName := "aws_workspaces_directory.main"

	domain := acctest.RandomDomainName()

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(t)
			testAccPreCheckWorkspacesDirectory(t)
			acctest.PreCheckDirectoryServiceSimpleDirectory(t)
			acctest.PreCheckHasIAMRole(t, "workspaces_DefaultRole")
		},
		ErrorCheck:   acctest.ErrorCheck(t, workspaces.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckAwsWorkspacesDirectoryDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccWorkspacesDirectory_selfServicePermissions(rName, domain),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAwsWorkspacesDirectoryExists(resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "self_service_permissions.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "self_service_permissions.0.change_compute_type", "false"),
					resource.TestCheckResourceAttr(resourceName, "self_service_permissions.0.increase_volume_size", "true"),
					resource.TestCheckResourceAttr(resourceName, "self_service_permissions.0.rebuild_workspace", "true"),
					resource.TestCheckResourceAttr(resourceName, "self_service_permissions.0.restart_workspace", "false"),
					resource.TestCheckResourceAttr(resourceName, "self_service_permissions.0.switch_running_mode", "true"),
				),
			},
		},
	})
}

func testAccAwsWorkspacesDirectory_workspaceAccessProperties(t *testing.T) {
	var v workspaces.WorkspaceDirectory
	rName := sdkacctest.RandString(8)

	resourceName := "aws_workspaces_directory.main"

	domain := acctest.RandomDomainName()

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(t)
			testAccPreCheckWorkspacesDirectory(t)
			acctest.PreCheckDirectoryServiceSimpleDirectory(t)
			acctest.PreCheckHasIAMRole(t, "workspaces_DefaultRole")
		},
		ErrorCheck:   acctest.ErrorCheck(t, workspaces.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckAwsWorkspacesDirectoryDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccWorkspacesDirectory_workspaceAccessProperties(rName, domain),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAwsWorkspacesDirectoryExists(resourceName, &v),
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

func testAccAwsWorkspacesDirectory_workspaceCreationProperties(t *testing.T) {
	var v workspaces.WorkspaceDirectory
	rName := sdkacctest.RandString(8)

	resourceName := "aws_workspaces_directory.main"
	resourceSecurityGroup := "aws_security_group.test"

	domain := acctest.RandomDomainName()

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(t)
			testAccPreCheckWorkspacesDirectory(t)
			acctest.PreCheckDirectoryServiceSimpleDirectory(t)
			acctest.PreCheckHasIAMRole(t, "workspaces_DefaultRole")
		},
		ErrorCheck:   acctest.ErrorCheck(t, workspaces.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckAwsWorkspacesDirectoryDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccWorkspacesDirectoryConfig_workspaceCreationProperties(rName, domain),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAwsWorkspacesDirectoryExists(resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "workspace_creation_properties.#", "1"),
					resource.TestCheckResourceAttrPair(resourceName, "workspace_creation_properties.0.custom_security_group_id", resourceSecurityGroup, "id"),
					resource.TestCheckResourceAttr(resourceName, "workspace_creation_properties.0.default_ou", "OU=AWS,DC=Workgroup,DC=Example,DC=com"),
					resource.TestCheckResourceAttr(resourceName, "workspace_creation_properties.0.enable_internet_access", "true"),
					resource.TestCheckResourceAttr(resourceName, "workspace_creation_properties.0.enable_maintenance_mode", "false"),
					resource.TestCheckResourceAttr(resourceName, "workspace_creation_properties.0.user_enabled_as_local_administrator", "false"),
				),
			},
		},
	})
}

func testAccAwsWorkspacesDirectory_workspaceCreationProperties_customSecurityGroupId_defaultOu(t *testing.T) {
	var v workspaces.WorkspaceDirectory
	rName := sdkacctest.RandString(8)

	resourceName := "aws_workspaces_directory.main"
	resourceSecurityGroup := "aws_security_group.test"

	domain := acctest.RandomDomainName()

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(t)
			testAccPreCheckWorkspacesDirectory(t)
			acctest.PreCheckDirectoryServiceSimpleDirectory(t)
			acctest.PreCheckHasIAMRole(t, "workspaces_DefaultRole")
		},
		ErrorCheck:   acctest.ErrorCheck(t, workspaces.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckAwsWorkspacesDirectoryDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccWorkspacesDirectoryConfig_workspaceCreationProperties_customSecurityGroupId_defaultOu_Absent(rName, domain),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAwsWorkspacesDirectoryExists(resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "workspace_creation_properties.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "workspace_creation_properties.0.custom_security_group_id", ""),
					resource.TestCheckResourceAttr(resourceName, "workspace_creation_properties.0.default_ou", ""),
				),
			},
			{
				Config: testAccWorkspacesDirectoryConfig_workspaceCreationProperties_customSecurityGroupId_defaultOu_Present(rName, domain),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAwsWorkspacesDirectoryExists(resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "workspace_creation_properties.#", "1"),
					resource.TestCheckResourceAttrPair(resourceName, "workspace_creation_properties.0.custom_security_group_id", resourceSecurityGroup, "id"),
					resource.TestCheckResourceAttr(resourceName, "workspace_creation_properties.0.default_ou", "OU=AWS,DC=Workgroup,DC=Example,DC=com"),
				),
			},
			{
				Config: testAccWorkspacesDirectoryConfig_workspaceCreationProperties_customSecurityGroupId_defaultOu_Absent(rName, domain),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAwsWorkspacesDirectoryExists(resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "workspace_creation_properties.#", "1"),
					resource.TestCheckResourceAttrSet(resourceName, "workspace_creation_properties.0.custom_security_group_id"),
					resource.TestCheckResourceAttrSet(resourceName, "workspace_creation_properties.0.default_ou"),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccAwsWorkspacesDirectory_ipGroupIds(t *testing.T) {
	var v workspaces.WorkspaceDirectory
	rName := sdkacctest.RandString(8)

	resourceName := "aws_workspaces_directory.test"

	domain := acctest.RandomDomainName()

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t); acctest.PreCheckHasIAMRole(t, "workspaces_DefaultRole") },
		ErrorCheck:   acctest.ErrorCheck(t, workspaces.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckAwsWorkspacesDirectoryDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccWorkspacesDirectoryConfig_ipGroupIds_create(rName, domain),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsWorkspacesDirectoryExists(resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "ip_group_ids.#", "1"),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "ip_group_ids.*", "aws_workspaces_ip_group.test_alpha", "id"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccWorkspacesDirectoryConfig_ipGroupIds_update(rName, domain),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsWorkspacesDirectoryExists(resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "ip_group_ids.#", "2"),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "ip_group_ids.*", "aws_workspaces_ip_group.test_beta", "id"),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "ip_group_ids.*", "aws_workspaces_ip_group.test_gamma", "id"),
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

func TestExpandSelfServicePermissions(t *testing.T) {
	cases := []struct {
		input    []interface{}
		expected *workspaces.SelfservicePermissions
	}{
		// Empty
		{
			input:    []interface{}{},
			expected: nil,
		},
		// Full
		{
			input: []interface{}{
				map[string]interface{}{
					"change_compute_type":  false,
					"increase_volume_size": false,
					"rebuild_workspace":    true,
					"restart_workspace":    true,
					"switch_running_mode":  true,
				},
			},
			expected: &workspaces.SelfservicePermissions{
				ChangeComputeType:  aws.String(workspaces.ReconnectEnumDisabled),
				IncreaseVolumeSize: aws.String(workspaces.ReconnectEnumDisabled),
				RebuildWorkspace:   aws.String(workspaces.ReconnectEnumEnabled),
				RestartWorkspace:   aws.String(workspaces.ReconnectEnumEnabled),
				SwitchRunningMode:  aws.String(workspaces.ReconnectEnumEnabled),
			},
		},
	}

	for _, c := range cases {
		actual := expandSelfServicePermissions(c.input)
		if !reflect.DeepEqual(actual, c.expected) {
			t.Fatalf("expected\n\n%#+v\n\ngot\n\n%#+v", c.expected, actual)
		}
	}
}

func TestFlattenSelfServicePermissions(t *testing.T) {
	cases := []struct {
		input    *workspaces.SelfservicePermissions
		expected []interface{}
	}{
		// Empty
		{
			input:    nil,
			expected: []interface{}{},
		},
		// Full
		{
			input: &workspaces.SelfservicePermissions{
				ChangeComputeType:  aws.String(workspaces.ReconnectEnumDisabled),
				IncreaseVolumeSize: aws.String(workspaces.ReconnectEnumDisabled),
				RebuildWorkspace:   aws.String(workspaces.ReconnectEnumEnabled),
				RestartWorkspace:   aws.String(workspaces.ReconnectEnumEnabled),
				SwitchRunningMode:  aws.String(workspaces.ReconnectEnumEnabled),
			},
			expected: []interface{}{
				map[string]interface{}{
					"change_compute_type":  false,
					"increase_volume_size": false,
					"rebuild_workspace":    true,
					"restart_workspace":    true,
					"switch_running_mode":  true,
				},
			},
		},
	}

	for _, c := range cases {
		actual := flattenSelfServicePermissions(c.input)
		if !reflect.DeepEqual(actual, c.expected) {
			t.Fatalf("expected\n\n%#+v\n\ngot\n\n%#+v", c.expected, actual)
		}
	}
}

func TestExpandWorkspaceAccessProperties(t *testing.T) {
	cases := []struct {
		input    []interface{}
		expected *workspaces.WorkspaceAccessProperties
	}{
		// Empty
		{
			input:    []interface{}{},
			expected: nil,
		},
		// Full
		{
			input: []interface{}{
				map[string]interface{}{
					"device_type_android":    "ALLOW",
					"device_type_chromeos":   "ALLOW",
					"device_type_ios":        "ALLOW",
					"device_type_linux":      "DENY",
					"device_type_osx":        "ALLOW",
					"device_type_web":        "DENY",
					"device_type_windows":    "DENY",
					"device_type_zeroclient": "DENY",
				},
			},
			expected: &workspaces.WorkspaceAccessProperties{
				DeviceTypeAndroid:    aws.String("ALLOW"),
				DeviceTypeChromeOs:   aws.String("ALLOW"),
				DeviceTypeIos:        aws.String("ALLOW"),
				DeviceTypeLinux:      aws.String("DENY"),
				DeviceTypeOsx:        aws.String("ALLOW"),
				DeviceTypeWeb:        aws.String("DENY"),
				DeviceTypeWindows:    aws.String("DENY"),
				DeviceTypeZeroClient: aws.String("DENY"),
			},
		},
	}

	for _, c := range cases {
		actual := expandWorkspaceAccessProperties(c.input)
		if !reflect.DeepEqual(actual, c.expected) {
			t.Fatalf("expected\n\n%#+v\n\ngot\n\n%#+v", c.expected, actual)
		}
	}
}

func TestFlattenWorkspaceAccessProperties(t *testing.T) {
	cases := []struct {
		input    *workspaces.WorkspaceAccessProperties
		expected []interface{}
	}{
		// Empty
		{
			input:    nil,
			expected: []interface{}{},
		},
		// Full
		{
			input: &workspaces.WorkspaceAccessProperties{
				DeviceTypeAndroid:    aws.String("ALLOW"),
				DeviceTypeChromeOs:   aws.String("ALLOW"),
				DeviceTypeIos:        aws.String("ALLOW"),
				DeviceTypeLinux:      aws.String("DENY"),
				DeviceTypeOsx:        aws.String("ALLOW"),
				DeviceTypeWeb:        aws.String("DENY"),
				DeviceTypeWindows:    aws.String("DENY"),
				DeviceTypeZeroClient: aws.String("DENY"),
			},
			expected: []interface{}{
				map[string]interface{}{
					"device_type_android":    "ALLOW",
					"device_type_chromeos":   "ALLOW",
					"device_type_ios":        "ALLOW",
					"device_type_linux":      "DENY",
					"device_type_osx":        "ALLOW",
					"device_type_web":        "DENY",
					"device_type_windows":    "DENY",
					"device_type_zeroclient": "DENY",
				},
			},
		},
	}

	for _, c := range cases {
		actual := flattenWorkspaceAccessProperties(c.input)
		if !reflect.DeepEqual(actual, c.expected) {
			t.Fatalf("expected\n\n%#+v\n\ngot\n\n%#+v", c.expected, actual)
		}
	}
}

func TestExpandWorkspaceCreationProperties(t *testing.T) {
	cases := []struct {
		input    []interface{}
		expected *workspaces.WorkspaceCreationProperties
	}{
		// Empty
		{
			input:    []interface{}{},
			expected: nil,
		},
		// Full
		{
			input: []interface{}{
				map[string]interface{}{
					"custom_security_group_id":            "sg-123456789012",
					"default_ou":                          "OU=AWS,DC=Workgroup,DC=Example,DC=com",
					"enable_internet_access":              true,
					"enable_maintenance_mode":             true,
					"user_enabled_as_local_administrator": true,
				},
			},
			expected: &workspaces.WorkspaceCreationProperties{
				CustomSecurityGroupId:           aws.String("sg-123456789012"),
				DefaultOu:                       aws.String("OU=AWS,DC=Workgroup,DC=Example,DC=com"),
				EnableInternetAccess:            aws.Bool(true),
				EnableMaintenanceMode:           aws.Bool(true),
				UserEnabledAsLocalAdministrator: aws.Bool(true),
			},
		},
		// Without Custom Security Group ID & Default OU
		{
			input: []interface{}{
				map[string]interface{}{
					"custom_security_group_id":            "",
					"default_ou":                          "",
					"enable_internet_access":              true,
					"enable_maintenance_mode":             true,
					"user_enabled_as_local_administrator": true,
				},
			},
			expected: &workspaces.WorkspaceCreationProperties{
				CustomSecurityGroupId:           nil,
				DefaultOu:                       nil,
				EnableInternetAccess:            aws.Bool(true),
				EnableMaintenanceMode:           aws.Bool(true),
				UserEnabledAsLocalAdministrator: aws.Bool(true),
			},
		},
	}

	for _, c := range cases {
		actual := expandWorkspaceCreationProperties(c.input)
		if !reflect.DeepEqual(actual, c.expected) {
			t.Fatalf("expected\n\n%#+v\n\ngot\n\n%#+v", c.expected, actual)
		}
	}
}

func TestFlattenWorkspaceCreationProperties(t *testing.T) {
	cases := []struct {
		input    *workspaces.DefaultWorkspaceCreationProperties
		expected []interface{}
	}{
		// Empty
		{
			input:    nil,
			expected: []interface{}{},
		},
		// Full
		{
			input: &workspaces.DefaultWorkspaceCreationProperties{
				CustomSecurityGroupId:           aws.String("sg-123456789012"),
				DefaultOu:                       aws.String("OU=AWS,DC=Workgroup,DC=Example,DC=com"),
				EnableInternetAccess:            aws.Bool(true),
				EnableMaintenanceMode:           aws.Bool(true),
				UserEnabledAsLocalAdministrator: aws.Bool(true),
			},
			expected: []interface{}{
				map[string]interface{}{
					"custom_security_group_id":            "sg-123456789012",
					"default_ou":                          "OU=AWS,DC=Workgroup,DC=Example,DC=com",
					"enable_internet_access":              true,
					"enable_maintenance_mode":             true,
					"user_enabled_as_local_administrator": true,
				},
			},
		},
	}

	for _, c := range cases {
		actual := flattenWorkspaceCreationProperties(c.input)
		if !reflect.DeepEqual(actual, c.expected) {
			t.Fatalf("expected\n\n%#+v\n\ngot\n\n%#+v", c.expected, actual)
		}
	}
}

func testAccCheckAwsWorkspacesDirectoryDestroy(s *terraform.State) error {
	conn := acctest.Provider.Meta().(*conns.AWSClient).WorkSpacesConn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_workspaces_directory" {
			continue
		}

		_, err := finder.DirectoryByID(conn, rs.Primary.ID)

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

func testAccCheckAwsWorkspacesDirectoryExists(n string, v *workspaces.WorkspaceDirectory) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No WorkSpaces Directory ID is set")
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).WorkSpacesConn

		output, err := finder.DirectoryByID(conn, rs.Primary.ID)

		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

func testAccPreCheckWorkspacesDirectory(t *testing.T) {
	conn := acctest.Provider.Meta().(*conns.AWSClient).WorkSpacesConn

	input := &workspaces.DescribeWorkspaceDirectoriesInput{}

	_, err := conn.DescribeWorkspaceDirectories(input)

	if acctest.PreCheckSkipError(err) {
		t.Skipf("skipping acceptance testing: %s", err)
	}

	if err != nil {
		t.Fatalf("unexpected PreCheck error: %s", err)
	}
}

func testAccAwsWorkspacesDirectoryConfig_Prerequisites(rName, domain string) string {
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

func testAccWorkspacesDirectoryConfig(rName, domain string) string {
	return acctest.ConfigCompose(
		testAccAwsWorkspacesDirectoryConfig_Prerequisites(rName, domain),
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

func testAccWorkspacesDirectory_selfServicePermissions(rName, domain string) string {
	return acctest.ConfigCompose(
		testAccAwsWorkspacesDirectoryConfig_Prerequisites(rName, domain),
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

func testAccWorkspacesDirectoryConfig_subnetIds(rName, domain string) string {
	return acctest.ConfigCompose(
		testAccAwsWorkspacesDirectoryConfig_Prerequisites(rName, domain),
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

func testAccWorkspacesDirectoryConfigTags1(rName, domain, tagKey1, tagValue1 string) string {
	return acctest.ConfigCompose(
		testAccAwsWorkspacesDirectoryConfig_Prerequisites(rName, domain),
		fmt.Sprintf(`
resource "aws_workspaces_directory" "main" {
  directory_id = aws_directory_service_directory.main.id

  tags = {
    %[1]q = %[2]q
  }
}
`, tagKey1, tagValue1))
}

func testAccWorkspacesDirectoryConfigTags2(rName, domain, tagKey1, tagValue1, tagKey2, tagValue2 string) string {
	return acctest.ConfigCompose(
		testAccAwsWorkspacesDirectoryConfig_Prerequisites(rName, domain),
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

func testAccWorkspacesDirectory_workspaceAccessProperties(rName, domain string) string {
	return acctest.ConfigCompose(
		testAccAwsWorkspacesDirectoryConfig_Prerequisites(rName, domain),
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

func testAccWorkspacesDirectoryConfig_workspaceCreationProperties(rName, domain string) string {
	return acctest.ConfigCompose(
		testAccAwsWorkspacesDirectoryConfig_Prerequisites(rName, domain),
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

func testAccWorkspacesDirectoryConfig_workspaceCreationProperties_customSecurityGroupId_defaultOu_Absent(rName, domain string) string {
	return acctest.ConfigCompose(
		testAccAwsWorkspacesDirectoryConfig_Prerequisites(rName, domain),
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

func testAccWorkspacesDirectoryConfig_workspaceCreationProperties_customSecurityGroupId_defaultOu_Present(rName, domain string) string {
	return acctest.ConfigCompose(
		testAccAwsWorkspacesDirectoryConfig_Prerequisites(rName, domain),
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

func testAccWorkspacesDirectoryConfig_ipGroupIds_create(rName, domain string) string {
	return acctest.ConfigCompose(
		testAccAwsWorkspacesDirectoryConfig_Prerequisites(rName, domain),
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

func testAccWorkspacesDirectoryConfig_ipGroupIds_update(rName, domain string) string {
	return acctest.ConfigCompose(
		testAccAwsWorkspacesDirectoryConfig_Prerequisites(rName, domain),
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
