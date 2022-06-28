package fsx_test

import (
	"fmt"
	"regexp"
	"strings"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/fsx"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tffsx "github.com/hashicorp/terraform-provider-aws/internal/service/fsx"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

func TestAccFSxOntapStorageVirtualMachine_basic(t *testing.T) {
	var storageVirtualMachine fsx.StorageVirtualMachine
	resourceName := "aws_fsx_ontap_storage_virtual_machine.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); acctest.PreCheckPartitionHasService(fsx.EndpointsID, t) },
		ErrorCheck:        acctest.ErrorCheck(t, fsx.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckOntapStorageVirtualMachineDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccONTAPStorageVirtualMachineConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckOntapStorageVirtualMachineExists(resourceName, &storageVirtualMachine),
					acctest.MatchResourceAttrRegionalARN(resourceName, "arn", "fsx", regexp.MustCompile(`storage-virtual-machine/fs-.+`)),
					resource.TestCheckResourceAttr(resourceName, "endpoints.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "endpoints.0.iscsi.#", "1"),
					resource.TestCheckResourceAttrSet(resourceName, "endpoints.0.iscsi.0.dns_name"),
					resource.TestCheckResourceAttr(resourceName, "endpoints.0.management.#", "1"),
					resource.TestCheckResourceAttrSet(resourceName, "endpoints.0.management.0.dns_name"),
					resource.TestCheckResourceAttr(resourceName, "endpoints.0.nfs.#", "1"),
					resource.TestCheckResourceAttrSet(resourceName, "endpoints.0.nfs.0.dns_name"),
					resource.TestCheckResourceAttrSet(resourceName, "file_system_id"),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "subtype", fsx.StorageVirtualMachineSubtypeDefault),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
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

func TestAccFSxOntapStorageVirtualMachine_rootVolumeSecurityStyle(t *testing.T) {
	var storageVirtualMachine fsx.StorageVirtualMachine
	resourceName := "aws_fsx_ontap_storage_virtual_machine.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); acctest.PreCheckPartitionHasService(fsx.EndpointsID, t) },
		ErrorCheck:        acctest.ErrorCheck(t, fsx.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckOntapStorageVirtualMachineDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccONTAPStorageVirtualMachineConfig_rootVolumeSecurityStyle(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckOntapStorageVirtualMachineExists(resourceName, &storageVirtualMachine),
					acctest.MatchResourceAttrRegionalARN(resourceName, "arn", "fsx", regexp.MustCompile(`storage-virtual-machine/fs-.+`)),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
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

func TestAccFSxOntapStorageVirtualMachine_svmAdminPassword(t *testing.T) {
	var storageVirtualMachine1, storageVirtualMachine2 fsx.StorageVirtualMachine
	resourceName := "aws_fsx_ontap_storage_virtual_machine.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	pass1 := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	pass2 := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); acctest.PreCheckPartitionHasService(fsx.EndpointsID, t) },
		ErrorCheck:        acctest.ErrorCheck(t, fsx.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckOntapStorageVirtualMachineDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccONTAPStorageVirtualMachineConfig_svmAdminPassword(rName, pass1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckOntapStorageVirtualMachineExists(resourceName, &storageVirtualMachine1),
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
					testAccCheckOntapStorageVirtualMachineExists(resourceName, &storageVirtualMachine2),
					testAccCheckOntapStorageVirtualMachineNotRecreated(&storageVirtualMachine1, &storageVirtualMachine2),
					resource.TestCheckResourceAttr(resourceName, "svm_admin_password", pass2),
				),
			},
		},
	})
}

func TestAccFSxOntapStorageVirtualMachine_disappears(t *testing.T) {
	var storageVirtualMachine fsx.StorageVirtualMachine
	resourceName := "aws_fsx_ontap_storage_virtual_machine.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); acctest.PreCheckPartitionHasService(fsx.EndpointsID, t) },
		ErrorCheck:        acctest.ErrorCheck(t, fsx.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckOntapStorageVirtualMachineDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccONTAPStorageVirtualMachineConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckOntapStorageVirtualMachineExists(resourceName, &storageVirtualMachine),
					acctest.CheckResourceDisappears(acctest.Provider, tffsx.ResourceOntapStorageVirtualMachine(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccFSxOntapStorageVirtualMachine_name(t *testing.T) {
	var storageVirtualMachine1, storageVirtualMachine2 fsx.StorageVirtualMachine
	resourceName := "aws_fsx_ontap_storage_virtual_machine.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	rName2 := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); acctest.PreCheckPartitionHasService(fsx.EndpointsID, t) },
		ErrorCheck:        acctest.ErrorCheck(t, fsx.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckOntapStorageVirtualMachineDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccONTAPStorageVirtualMachineConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckOntapStorageVirtualMachineExists(resourceName, &storageVirtualMachine1),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
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
					testAccCheckOntapStorageVirtualMachineExists(resourceName, &storageVirtualMachine2),
					testAccCheckOntapStorageVirtualMachineRecreated(&storageVirtualMachine1, &storageVirtualMachine2),
					resource.TestCheckResourceAttr(resourceName, "name", rName2),
				),
			},
		},
	})
}

func TestAccFSxOntapStorageVirtualMachine_tags(t *testing.T) {
	var storageVirtualMachine1, storageVirtualMachine2, storageVirtualMachine3 fsx.StorageVirtualMachine
	resourceName := "aws_fsx_ontap_storage_virtual_machine.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); acctest.PreCheckPartitionHasService(fsx.EndpointsID, t) },
		ErrorCheck:        acctest.ErrorCheck(t, fsx.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckOntapStorageVirtualMachineDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccONTAPStorageVirtualMachineConfig_tags1(rName, "key1", "value1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckOntapStorageVirtualMachineExists(resourceName, &storageVirtualMachine1),
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
				Config: testAccONTAPStorageVirtualMachineConfig_tags2(rName, "key1", "value1updated", "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckOntapStorageVirtualMachineExists(resourceName, &storageVirtualMachine2),
					testAccCheckOntapStorageVirtualMachineNotRecreated(&storageVirtualMachine1, &storageVirtualMachine2),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1updated"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
			{
				Config: testAccONTAPStorageVirtualMachineConfig_tags1(rName, "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckOntapStorageVirtualMachineExists(resourceName, &storageVirtualMachine3),
					testAccCheckOntapStorageVirtualMachineNotRecreated(&storageVirtualMachine2, &storageVirtualMachine3),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
		},
	})
}

func TestAccFSxOntapStorageVirtualMachine_activeDirectory(t *testing.T) {
	var storageVirtualMachine1 fsx.StorageVirtualMachine
	resourceName := "aws_fsx_ontap_storage_virtual_machine.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	netBiosName := "tftest-" + sdkacctest.RandString(7)
	domainNetbiosName := "tftestcorp"
	domainName := "tftestcorp.local"
	domainPassword1 := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); acctest.PreCheckPartitionHasService(fsx.EndpointsID, t) },
		ErrorCheck:        acctest.ErrorCheck(t, fsx.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckOntapStorageVirtualMachineDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccONTAPStorageVirtualMachineConfig_virutalSelfManagedActiveDirectory(rName, netBiosName, domainNetbiosName, domainName, domainPassword1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckOntapStorageVirtualMachineExists(resourceName, &storageVirtualMachine1),
					resource.TestCheckResourceAttr(resourceName, "active_directory_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "active_directory_configuration.0.netbios_name", strings.ToUpper(netBiosName)),
					resource.TestCheckResourceAttr(resourceName, "active_directory_configuration.0.self_managed_active_directory_configuration.0.domain_name", domainName),
					resource.TestCheckResourceAttr(resourceName, "endpoints.0.smb.#", "1"),
					resource.TestCheckResourceAttrSet(resourceName, "endpoints.0.smb.0.dns_name"),
					resource.TestCheckResourceAttr(resourceName, "active_directory_configuration.0.self_managed_active_directory_configuration.0.organizational_unit_distinguished_name", fmt.Sprintf("OU=computers,OU=%s", domainNetbiosName)),
					resource.TestCheckResourceAttr(resourceName, "active_directory_configuration.0.self_managed_active_directory_configuration.0.password", domainPassword1),
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

func testAccCheckOntapStorageVirtualMachineExists(resourceName string, svm *fsx.StorageVirtualMachine) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("Not found: %s", resourceName)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).FSxConn

		storageVirtualMachine, err := tffsx.FindStorageVirtualMachineByID(conn, rs.Primary.ID)
		if err != nil {
			return err
		}

		if storageVirtualMachine == nil {
			return fmt.Errorf("FSx ONTAP Storage Virtual Machine (%s) not found", rs.Primary.ID)
		}

		*svm = *storageVirtualMachine

		return nil
	}
}

func testAccCheckOntapStorageVirtualMachineDestroy(s *terraform.State) error {
	conn := acctest.Provider.Meta().(*conns.AWSClient).FSxConn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_storage_virtual_machine" {
			continue
		}

		storageVirtualMachine, err := tffsx.FindStorageVirtualMachineByID(conn, rs.Primary.ID)
		if tfresource.NotFound(err) {
			continue
		}

		if storageVirtualMachine != nil {
			return fmt.Errorf("FSx ONTAP Storage Virtual Machine (%s) still exists", rs.Primary.ID)
		}
	}
	return nil
}

func testAccCheckOntapStorageVirtualMachineNotRecreated(i, j *fsx.StorageVirtualMachine) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if aws.StringValue(i.StorageVirtualMachineId) != aws.StringValue(j.StorageVirtualMachineId) {
			return fmt.Errorf("FSx Storage Virtual Machine (%s) recreated", aws.StringValue(i.StorageVirtualMachineId))
		}

		return nil
	}
}

func testAccCheckOntapStorageVirtualMachineRecreated(i, j *fsx.StorageVirtualMachine) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if aws.StringValue(i.StorageVirtualMachineId) == aws.StringValue(j.StorageVirtualMachineId) {
			return fmt.Errorf("FSx Storage Virtual Machine (%s) not recreated", aws.StringValue(i.StorageVirtualMachineId))
		}

		return nil
	}
}

func testAccOntapStorageVirtualMachineBaseConfig(rName string) string {
	return acctest.ConfigCompose(acctest.ConfigAvailableAZsNoOptIn(), fmt.Sprintf(`
data "aws_partition" "current" {}

resource "aws_vpc" "test" {
  cidr_block = "10.0.0.0/16"

  tags = {
    Name = %[1]q
  }
}

resource "aws_subnet" "test1" {
  vpc_id            = aws_vpc.test.id
  cidr_block        = "10.0.1.0/24"
  availability_zone = data.aws_availability_zones.available.names[0]

  tags = {
    Name = %[1]q
  }
}

resource "aws_subnet" "test2" {
  vpc_id            = aws_vpc.test.id
  cidr_block        = "10.0.2.0/24"
  availability_zone = data.aws_availability_zones.available.names[1]

  tags = {
    Name = %[1]q
  }
}

resource "aws_fsx_ontap_file_system" "test" {
  storage_capacity    = 1024
  subnet_ids          = [aws_subnet.test1.id, aws_subnet.test2.id]
  deployment_type     = "MULTI_AZ_1"
  throughput_capacity = 512
  preferred_subnet_id = aws_subnet.test1.id

  tags = {
    Name = %[1]q
  }
}
`, rName))
}

func testAccOntapStorageVirtualMachineADConfig(rName string, domainName string, domainPassword string) string {
	return acctest.ConfigCompose(testAccOntapStorageVirtualMachineBaseConfig(rName), fmt.Sprintf(`

resource "aws_directory_service_directory" "test" {
  edition  = "Standard"
  name     = %[2]q
  password = %[3]q
  type     = "MicrosoftAD"

  vpc_settings {
    subnet_ids = [aws_subnet.test1.id, aws_subnet.test2.id]
    vpc_id     = aws_vpc.test.id
  }
}
`, rName, domainName, domainPassword))
}

func testAccONTAPStorageVirtualMachineConfig_basic(rName string) string {
	return acctest.ConfigCompose(testAccOntapStorageVirtualMachineBaseConfig(rName), fmt.Sprintf(`
resource "aws_fsx_ontap_storage_virtual_machine" "test" {
  file_system_id = aws_fsx_ontap_file_system.test.id
  name           = %[1]q
}
`, rName))
}

func testAccONTAPStorageVirtualMachineConfig_rootVolumeSecurityStyle(rName string) string {
	return acctest.ConfigCompose(testAccOntapStorageVirtualMachineBaseConfig(rName), fmt.Sprintf(`
resource "aws_fsx_ontap_storage_virtual_machine" "test" {
  file_system_id             = aws_fsx_ontap_file_system.test.id
  name                       = %[1]q
  root_volume_security_style = "NTFS"
}
`, rName))
}

func testAccONTAPStorageVirtualMachineConfig_svmAdminPassword(rName string, pass string) string {
	return acctest.ConfigCompose(testAccOntapStorageVirtualMachineBaseConfig(rName), fmt.Sprintf(`
resource "aws_fsx_ontap_storage_virtual_machine" "test" {
  file_system_id     = aws_fsx_ontap_file_system.test.id
  name               = %[1]q
  svm_admin_password = %[2]q
}
`, rName, pass))
}

func testAccONTAPStorageVirtualMachineConfig_tags1(rName, tagKey1, tagValue1 string) string {
	return acctest.ConfigCompose(testAccOntapStorageVirtualMachineBaseConfig(rName), fmt.Sprintf(`
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
	return acctest.ConfigCompose(testAccOntapStorageVirtualMachineBaseConfig(rName), fmt.Sprintf(`
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

func testAccONTAPStorageVirtualMachineConfig_virutalSelfManagedActiveDirectory(rName string, netBiosName string, domainNetbiosName string, domainName string, domainPassword string) string {
	return acctest.ConfigCompose(testAccOntapStorageVirtualMachineADConfig(rName, domainName, domainPassword), fmt.Sprintf(`
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
    }
  }
}
`, rName, netBiosName, domainName, domainPassword, domainNetbiosName))
}
