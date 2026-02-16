// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package fsx_test

import (
	"context"
	"fmt"
	"strings"
	"testing"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/aws"
	awstypes "github.com/aws/aws-sdk-go-v2/service/fsx/types"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	tffsx "github.com/hashicorp/terraform-provider-aws/internal/service/fsx"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccFSxONTAPStorageVirtualMachine_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var storageVirtualMachine awstypes.StorageVirtualMachine
	resourceName := "aws_fsx_ontap_storage_virtual_machine.test"
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckPartitionHasService(t, names.FSxEndpointID) },
		ErrorCheck:               acctest.ErrorCheck(t, names.FSxServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckONTAPStorageVirtualMachineDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccONTAPStorageVirtualMachineConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckONTAPStorageVirtualMachineExists(ctx, t, resourceName, &storageVirtualMachine),
					acctest.MatchResourceAttrRegionalARN(ctx, resourceName, names.AttrARN, "fsx", regexache.MustCompile(`storage-virtual-machine/fs-.+`)),
					resource.TestCheckResourceAttr(resourceName, "endpoints.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "endpoints.0.iscsi.#", "1"),
					resource.TestCheckResourceAttrSet(resourceName, "endpoints.0.iscsi.0.dns_name"),
					resource.TestCheckResourceAttr(resourceName, "endpoints.0.management.#", "1"),
					resource.TestCheckResourceAttrSet(resourceName, "endpoints.0.management.0.dns_name"),
					resource.TestCheckResourceAttr(resourceName, "endpoints.0.nfs.#", "1"),
					resource.TestCheckResourceAttrSet(resourceName, "endpoints.0.nfs.0.dns_name"),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrFileSystemID),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, "subtype", string(awstypes.StorageVirtualMachineSubtypeDefault)),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, "0"),
					resource.TestCheckResourceAttrSet(resourceName, "uuid"),
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

func TestAccFSxONTAPStorageVirtualMachine_rootVolumeSecurityStyle(t *testing.T) {
	ctx := acctest.Context(t)
	var storageVirtualMachine awstypes.StorageVirtualMachine
	resourceName := "aws_fsx_ontap_storage_virtual_machine.test"
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckPartitionHasService(t, names.FSxEndpointID) },
		ErrorCheck:               acctest.ErrorCheck(t, names.FSxServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckONTAPStorageVirtualMachineDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccONTAPStorageVirtualMachineConfig_rootVolumeSecurityStyle(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckONTAPStorageVirtualMachineExists(ctx, t, resourceName, &storageVirtualMachine),
					acctest.MatchResourceAttrRegionalARN(ctx, resourceName, names.AttrARN, "fsx", regexache.MustCompile(`storage-virtual-machine/fs-.+`)),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, "0"),
					resource.TestCheckResourceAttr(resourceName, "endpoints.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "endpoints.0.iscsi.#", "1"),
					resource.TestCheckResourceAttrSet(resourceName, "endpoints.0.iscsi.0.dns_name"),
					resource.TestCheckResourceAttr(resourceName, "endpoints.0.management.#", "1"),
					resource.TestCheckResourceAttrSet(resourceName, "endpoints.0.management.0.dns_name"),
					resource.TestCheckResourceAttr(resourceName, "endpoints.0.nfs.#", "1"),
					resource.TestCheckResourceAttrSet(resourceName, "endpoints.0.nfs.0.dns_name"),
					resource.TestCheckResourceAttrSet(resourceName, "root_volume_security_style"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"root_volume_security_style"},
			},
		},
	})
}

func TestAccFSxONTAPStorageVirtualMachine_svmAdminPassword(t *testing.T) {
	ctx := acctest.Context(t)
	var storageVirtualMachine1, storageVirtualMachine2 awstypes.StorageVirtualMachine
	resourceName := "aws_fsx_ontap_storage_virtual_machine.test"
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	pass1 := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	pass2 := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckPartitionHasService(t, names.FSxEndpointID) },
		ErrorCheck:               acctest.ErrorCheck(t, names.FSxServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckONTAPStorageVirtualMachineDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccONTAPStorageVirtualMachineConfig_svmAdminPassword(rName, pass1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckONTAPStorageVirtualMachineExists(ctx, t, resourceName, &storageVirtualMachine1),
					resource.TestCheckResourceAttr(resourceName, "svm_admin_password", pass1),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"svm_admin_password"},
			},
			{
				Config: testAccONTAPStorageVirtualMachineConfig_svmAdminPassword(rName, pass2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckONTAPStorageVirtualMachineExists(ctx, t, resourceName, &storageVirtualMachine2),
					testAccCheckONTAPStorageVirtualMachineNotRecreated(&storageVirtualMachine1, &storageVirtualMachine2),
					resource.TestCheckResourceAttr(resourceName, "svm_admin_password", pass2),
				),
			},
		},
	})
}

func TestAccFSxONTAPStorageVirtualMachine_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	var storageVirtualMachine awstypes.StorageVirtualMachine
	resourceName := "aws_fsx_ontap_storage_virtual_machine.test"
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckPartitionHasService(t, names.FSxEndpointID) },
		ErrorCheck:               acctest.ErrorCheck(t, names.FSxServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckONTAPStorageVirtualMachineDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccONTAPStorageVirtualMachineConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckONTAPStorageVirtualMachineExists(ctx, t, resourceName, &storageVirtualMachine),
					acctest.CheckSDKResourceDisappears(ctx, t, tffsx.ResourceONTAPStorageVirtualMachine(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccFSxONTAPStorageVirtualMachine_name(t *testing.T) {
	ctx := acctest.Context(t)
	var storageVirtualMachine1, storageVirtualMachine2 awstypes.StorageVirtualMachine
	resourceName := "aws_fsx_ontap_storage_virtual_machine.test"
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	rName2 := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckPartitionHasService(t, names.FSxEndpointID) },
		ErrorCheck:               acctest.ErrorCheck(t, names.FSxServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckONTAPStorageVirtualMachineDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccONTAPStorageVirtualMachineConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckONTAPStorageVirtualMachineExists(ctx, t, resourceName, &storageVirtualMachine1),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccONTAPStorageVirtualMachineConfig_basic(rName2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckONTAPStorageVirtualMachineExists(ctx, t, resourceName, &storageVirtualMachine2),
					testAccCheckONTAPStorageVirtualMachineRecreated(&storageVirtualMachine1, &storageVirtualMachine2),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName2),
				),
			},
		},
	})
}

func TestAccFSxONTAPStorageVirtualMachine_tags(t *testing.T) {
	ctx := acctest.Context(t)
	var storageVirtualMachine1, storageVirtualMachine2, storageVirtualMachine3 awstypes.StorageVirtualMachine
	resourceName := "aws_fsx_ontap_storage_virtual_machine.test"
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckPartitionHasService(t, names.FSxEndpointID) },
		ErrorCheck:               acctest.ErrorCheck(t, names.FSxServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckONTAPStorageVirtualMachineDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccONTAPStorageVirtualMachineConfig_tags1(rName, acctest.CtKey1, acctest.CtValue1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckONTAPStorageVirtualMachineExists(ctx, t, resourceName, &storageVirtualMachine1),
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
				Config: testAccONTAPStorageVirtualMachineConfig_tags2(rName, acctest.CtKey1, acctest.CtValue1Updated, acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckONTAPStorageVirtualMachineExists(ctx, t, resourceName, &storageVirtualMachine2),
					testAccCheckONTAPStorageVirtualMachineNotRecreated(&storageVirtualMachine1, &storageVirtualMachine2),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, "2"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey1, acctest.CtValue1Updated),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey2, acctest.CtValue2),
				),
			},
			{
				Config: testAccONTAPStorageVirtualMachineConfig_tags1(rName, acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckONTAPStorageVirtualMachineExists(ctx, t, resourceName, &storageVirtualMachine3),
					testAccCheckONTAPStorageVirtualMachineNotRecreated(&storageVirtualMachine2, &storageVirtualMachine3),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, "1"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey2, acctest.CtValue2),
				),
			},
		},
	})
}

func TestAccFSxONTAPStorageVirtualMachine_activeDirectoryCreate(t *testing.T) {
	ctx := acctest.Context(t)
	var storageVirtualMachine awstypes.StorageVirtualMachine
	resourceName := "aws_fsx_ontap_storage_virtual_machine.test"
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	netBiosName := "tftest-" + sdkacctest.RandString(7)
	domainNetbiosName := "tftest" + sdkacctest.RandString(4)
	domainName := domainNetbiosName + ".local"
	domainPassword := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckPartitionHasService(t, names.FSxEndpointID) },
		ErrorCheck:               acctest.ErrorCheck(t, names.FSxServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckONTAPStorageVirtualMachineDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccONTAPStorageVirtualMachineConfig_selfManagedActiveDirectory(rName, netBiosName, domainNetbiosName, domainName, domainPassword),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckONTAPStorageVirtualMachineExists(ctx, t, resourceName, &storageVirtualMachine),
					resource.TestCheckResourceAttr(resourceName, "active_directory_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "active_directory_configuration.0.netbios_name", strings.ToUpper(netBiosName)),
					resource.TestCheckResourceAttr(resourceName, "active_directory_configuration.0.self_managed_active_directory_configuration.0.domain_name", domainName),
					resource.TestCheckResourceAttr(resourceName, "active_directory_configuration.0.self_managed_active_directory_configuration.0.file_system_administrators_group", "Admins"),
					resource.TestCheckResourceAttr(resourceName, "active_directory_configuration.0.self_managed_active_directory_configuration.0.organizational_unit_distinguished_name", fmt.Sprintf("OU=computers,OU=%s", domainNetbiosName)),
					resource.TestCheckResourceAttr(resourceName, "active_directory_configuration.0.self_managed_active_directory_configuration.0.password", domainPassword),
					resource.TestCheckResourceAttr(resourceName, "active_directory_configuration.0.self_managed_active_directory_configuration.0.username", "Admin"),
					resource.TestCheckResourceAttr(resourceName, "endpoints.0.smb.#", "1"),
					resource.TestCheckResourceAttrSet(resourceName, "endpoints.0.smb.0.dns_name"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"active_directory_configuration",
				},
			},
		},
	})
}

func TestAccFSxONTAPStorageVirtualMachine_activeDirectoryJoin(t *testing.T) {
	ctx := acctest.Context(t)
	var storageVirtualMachine1, storageVirtualMachine2 awstypes.StorageVirtualMachine
	resourceName := "aws_fsx_ontap_storage_virtual_machine.test"
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	netBiosName := "tftest-" + sdkacctest.RandString(7)
	domainNetbiosName := "tftest" + sdkacctest.RandString(4)
	domainName := domainNetbiosName + ".local"
	domainPassword := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckPartitionHasService(t, names.FSxEndpointID) },
		ErrorCheck:               acctest.ErrorCheck(t, names.FSxServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckONTAPStorageVirtualMachineDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccONTAPStorageVirtualMachineConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckONTAPStorageVirtualMachineExists(ctx, t, resourceName, &storageVirtualMachine1),
					resource.TestCheckResourceAttr(resourceName, "active_directory_configuration.#", "0"),
				),
			},
			{
				Config: testAccONTAPStorageVirtualMachineConfig_selfManagedActiveDirectory(rName, netBiosName, domainNetbiosName, domainName, domainPassword),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckONTAPStorageVirtualMachineExists(ctx, t, resourceName, &storageVirtualMachine2),
					testAccCheckONTAPStorageVirtualMachineNotRecreated(&storageVirtualMachine1, &storageVirtualMachine2),
					resource.TestCheckResourceAttr(resourceName, "active_directory_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "active_directory_configuration.0.netbios_name", strings.ToUpper(netBiosName)),
					resource.TestCheckResourceAttr(resourceName, "active_directory_configuration.0.self_managed_active_directory_configuration.0.domain_name", domainName),
					resource.TestCheckResourceAttr(resourceName, "active_directory_configuration.0.self_managed_active_directory_configuration.0.file_system_administrators_group", "Admins"),
					resource.TestCheckResourceAttr(resourceName, "active_directory_configuration.0.self_managed_active_directory_configuration.0.organizational_unit_distinguished_name", fmt.Sprintf("OU=computers,OU=%s", domainNetbiosName)),
					resource.TestCheckResourceAttr(resourceName, "active_directory_configuration.0.self_managed_active_directory_configuration.0.password", domainPassword),
					resource.TestCheckResourceAttr(resourceName, "active_directory_configuration.0.self_managed_active_directory_configuration.0.username", "Admin"),
				),
			},
		},
	})
}

func testAccCheckONTAPStorageVirtualMachineExists(ctx context.Context, t *testing.T, n string, v *awstypes.StorageVirtualMachine) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.ProviderMeta(ctx, t).FSxClient(ctx)

		output, err := tffsx.FindStorageVirtualMachineByID(ctx, conn, rs.Primary.ID)

		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

func testAccCheckONTAPStorageVirtualMachineDestroy(ctx context.Context, t *testing.T) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.ProviderMeta(ctx, t).FSxClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_storage_virtual_machine" {
				continue
			}

			_, err := tffsx.FindStorageVirtualMachineByID(ctx, conn, rs.Primary.ID)

			if retry.NotFound(err) {
				continue
			}

			return fmt.Errorf("FSx ONTAP Storage Virtual Machine (%s) still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccCheckONTAPStorageVirtualMachineNotRecreated(i, j *awstypes.StorageVirtualMachine) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if aws.ToString(i.StorageVirtualMachineId) != aws.ToString(j.StorageVirtualMachineId) {
			return fmt.Errorf("FSx ONTAP Storage Virtual Machine (%s) recreated", aws.ToString(i.StorageVirtualMachineId))
		}

		return nil
	}
}

func testAccCheckONTAPStorageVirtualMachineRecreated(i, j *awstypes.StorageVirtualMachine) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if aws.ToString(i.StorageVirtualMachineId) == aws.ToString(j.StorageVirtualMachineId) {
			return fmt.Errorf("FSx ONTAP Storage Virtual Machine (%s) not recreated", aws.ToString(i.StorageVirtualMachineId))
		}

		return nil
	}
}

func testAccONTAPStorageVirtualMachineConfig_base(rName string) string {
	return acctest.ConfigCompose(testAccONTAPFileSystemConfig_base(rName), fmt.Sprintf(`
resource "aws_fsx_ontap_file_system" "test" {
  storage_capacity    = 1024
  subnet_ids          = aws_subnet.test[*].id
  deployment_type     = "MULTI_AZ_1"
  throughput_capacity = 512
  preferred_subnet_id = aws_subnet.test[0].id

  tags = {
    Name = %[1]q
  }
}
`, rName))
}

func testAccONTAPStorageVirtualMachineADConfig_base(rName, domainName, domainPassword string) string {
	return acctest.ConfigCompose(testAccONTAPStorageVirtualMachineConfig_base(rName), fmt.Sprintf(`
resource "aws_directory_service_directory" "test" {
  edition  = "Standard"
  name     = %[1]q
  password = %[2]q
  type     = "MicrosoftAD"

  vpc_settings {
    subnet_ids = aws_subnet.test[*].id
    vpc_id     = aws_vpc.test.id
  }
}
`, domainName, domainPassword))
}

func testAccONTAPStorageVirtualMachineConfig_basic(rName string) string {
	return acctest.ConfigCompose(testAccONTAPStorageVirtualMachineConfig_base(rName), fmt.Sprintf(`
resource "aws_fsx_ontap_storage_virtual_machine" "test" {
  file_system_id = aws_fsx_ontap_file_system.test.id
  name           = %[1]q
}
`, rName))
}

func testAccONTAPStorageVirtualMachineConfig_rootVolumeSecurityStyle(rName string) string {
	return acctest.ConfigCompose(testAccONTAPStorageVirtualMachineConfig_base(rName), fmt.Sprintf(`
resource "aws_fsx_ontap_storage_virtual_machine" "test" {
  file_system_id             = aws_fsx_ontap_file_system.test.id
  name                       = %[1]q
  root_volume_security_style = "NTFS"
}
`, rName))
}

func testAccONTAPStorageVirtualMachineConfig_svmAdminPassword(rName, pass string) string {
	return acctest.ConfigCompose(testAccONTAPStorageVirtualMachineConfig_base(rName), fmt.Sprintf(`
resource "aws_fsx_ontap_storage_virtual_machine" "test" {
  file_system_id     = aws_fsx_ontap_file_system.test.id
  name               = %[1]q
  svm_admin_password = %[2]q
}
`, rName, pass))
}

func testAccONTAPStorageVirtualMachineConfig_tags1(rName, tagKey1, tagValue1 string) string {
	return acctest.ConfigCompose(testAccONTAPStorageVirtualMachineConfig_base(rName), fmt.Sprintf(`
resource "aws_fsx_ontap_storage_virtual_machine" "test" {
  file_system_id = aws_fsx_ontap_file_system.test.id
  name           = %[1]q

  tags = {
    %[2]q = %[3]q
  }
}
`, rName, tagKey1, tagValue1))
}

func testAccONTAPStorageVirtualMachineConfig_tags2(rName, tagKey1, tagValue1, tagKey2, tagValue2 string) string {
	return acctest.ConfigCompose(testAccONTAPStorageVirtualMachineConfig_base(rName), fmt.Sprintf(`
resource "aws_fsx_ontap_storage_virtual_machine" "test" {
  file_system_id = aws_fsx_ontap_file_system.test.id
  name           = %[1]q

  tags = {
    %[2]q = %[3]q
    %[4]q = %[5]q
  }
}
`, rName, tagKey1, tagValue1, tagKey2, tagValue2))
}

func testAccONTAPStorageVirtualMachineConfig_selfManagedActiveDirectory(rName, netBiosName, domainNetbiosName, domainName, domainPassword string) string {
	return acctest.ConfigCompose(testAccONTAPStorageVirtualMachineADConfig_base(rName, domainName, domainPassword), fmt.Sprintf(`
resource "aws_fsx_ontap_storage_virtual_machine" "test" {
  file_system_id = aws_fsx_ontap_file_system.test.id
  name           = %[1]q
  depends_on     = [aws_directory_service_directory.test]

  active_directory_configuration {
    netbios_name = %[2]q
    self_managed_active_directory_configuration {
      dns_ips                                = aws_directory_service_directory.test.dns_ip_addresses
      domain_name                            = %[3]q
      password                               = %[4]q
      username                               = "Admin"
      organizational_unit_distinguished_name = "OU=computers,OU=%[5]s"
      file_system_administrators_group       = "Admins"
    }
  }
}
`, rName, netBiosName, domainName, domainPassword, domainNetbiosName))
}
