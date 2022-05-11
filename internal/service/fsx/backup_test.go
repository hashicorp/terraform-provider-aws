package fsx_test

import (
	"fmt"
	"regexp"
	"strings"
	"testing"

	"github.com/aws/aws-sdk-go/service/fsx"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tffsx "github.com/hashicorp/terraform-provider-aws/internal/service/fsx"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

func TestAccFSxBackup_basic(t *testing.T) {
	var backup fsx.Backup
	resourceName := "aws_fsx_backup.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); acctest.PreCheckPartitionHasService(fsx.EndpointsID, t) },
		ErrorCheck:        acctest.ErrorCheck(t, fsx.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckBackupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccBackupBasicConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBackupExists(resourceName, &backup),
					acctest.MatchResourceAttrRegionalARN(resourceName, "arn", "fsx", regexp.MustCompile(`backup/.+`)),
					acctest.CheckResourceAttrAccountID(resourceName, "owner_id"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
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

func TestAccFSxBackup_ontapBasic(t *testing.T) {
	var backup fsx.Backup
	resourceName := "aws_fsx_backup.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	//FSX ONTAP Volume Names can't use dash only underscore
	vName := strings.Replace(sdkacctest.RandomWithPrefix(acctest.ResourcePrefix), "-", "_", -1)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); acctest.PreCheckPartitionHasService(fsx.EndpointsID, t) },
		ErrorCheck:        acctest.ErrorCheck(t, fsx.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckBackupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccBackupONTAPBasicConfig(rName, vName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBackupExists(resourceName, &backup),
					acctest.MatchResourceAttrRegionalARN(resourceName, "arn", "fsx", regexp.MustCompile(`backup/.+`)),
					acctest.CheckResourceAttrAccountID(resourceName, "owner_id"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
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

func TestAccFSxBackup_openzfsBasic(t *testing.T) {
	var backup fsx.Backup
	resourceName := "aws_fsx_backup.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); acctest.PreCheckPartitionHasService(fsx.EndpointsID, t) },
		ErrorCheck:        acctest.ErrorCheck(t, fsx.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckBackupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccBackupOpenzfsBasicConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBackupExists(resourceName, &backup),
					acctest.MatchResourceAttrRegionalARN(resourceName, "arn", "fsx", regexp.MustCompile(`backup/.+`)),
					acctest.CheckResourceAttrAccountID(resourceName, "owner_id"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
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

func TestAccFSxBackup_windowsBasic(t *testing.T) {
	var backup fsx.Backup
	resourceName := "aws_fsx_backup.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); acctest.PreCheckPartitionHasService(fsx.EndpointsID, t) },
		ErrorCheck:        acctest.ErrorCheck(t, fsx.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckBackupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccBackupWindowsBasicConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBackupExists(resourceName, &backup),
					acctest.MatchResourceAttrRegionalARN(resourceName, "arn", "fsx", regexp.MustCompile(`backup/.+`)),
					acctest.CheckResourceAttrAccountID(resourceName, "owner_id"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
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

func TestAccFSxBackup_disappears(t *testing.T) {
	var backup fsx.Backup
	resourceName := "aws_fsx_backup.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); acctest.PreCheckPartitionHasService(fsx.EndpointsID, t) },
		ErrorCheck:        acctest.ErrorCheck(t, fsx.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckBackupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccBackupBasicConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBackupExists(resourceName, &backup),
					acctest.CheckResourceDisappears(acctest.Provider, tffsx.ResourceBackup(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccFSxBackup_Disappears_filesystem(t *testing.T) {
	var backup fsx.Backup
	resourceName := "aws_fsx_backup.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); acctest.PreCheckPartitionHasService(fsx.EndpointsID, t) },
		ErrorCheck:        acctest.ErrorCheck(t, fsx.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckBackupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccBackupBasicConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBackupExists(resourceName, &backup),
					acctest.CheckResourceDisappears(acctest.Provider, tffsx.ResourceLustreFileSystem(), "aws_fsx_lustre_file_system.test"),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccFSxBackup_tags(t *testing.T) {
	var backup fsx.Backup
	resourceName := "aws_fsx_backup.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); acctest.PreCheckPartitionHasService(fsx.EndpointsID, t) },
		ErrorCheck:        acctest.ErrorCheck(t, fsx.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckBackupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccBackupTags1Config(rName, "key1", "value1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBackupExists(resourceName, &backup),
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
				Config: testAccBackupTags2Config(rName, "key1", "value1updated", "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBackupExists(resourceName, &backup),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1updated"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
			{
				Config: testAccBackupTags1Config(rName, "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBackupExists(resourceName, &backup),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
		},
	})
}

func TestAccFSxBackup_implicitTags(t *testing.T) {
	var backup fsx.Backup
	resourceName := "aws_fsx_backup.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); acctest.PreCheckPartitionHasService(fsx.EndpointsID, t) },
		ErrorCheck:        acctest.ErrorCheck(t, fsx.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckBackupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccBackupImplictTagsConfig("key1", "value1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBackupExists(resourceName, &backup),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1"),
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

func testAccCheckBackupExists(resourceName string, fs *fsx.Backup) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("Not found: %s", resourceName)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).FSxConn

		output, err := tffsx.FindBackupByID(conn, rs.Primary.ID)
		if err != nil {
			return err
		}

		if output == nil {
			return fmt.Errorf("FSx Backup (%s) not found", rs.Primary.ID)
		}

		*fs = *output

		return nil
	}
}

func testAccCheckBackupDestroy(s *terraform.State) error {
	conn := acctest.Provider.Meta().(*conns.AWSClient).FSxConn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_fsx_backup" {
			continue
		}

		_, err := tffsx.FindBackupByID(conn, rs.Primary.ID)
		if tfresource.NotFound(err) {
			continue
		}

		if err != nil {
			return err
		}

		return fmt.Errorf("FSx Backup %s still exists", rs.Primary.ID)
	}
	return nil
}

func testAccBackupBaseConfig() string {
	return acctest.ConfigCompose(acctest.ConfigAvailableAZsNoOptIn(), `
resource "aws_vpc" "test" {
  cidr_block = "10.0.0.0/16"
}

resource "aws_subnet" "test1" {
  vpc_id            = aws_vpc.test.id
  cidr_block        = "10.0.1.0/24"
  availability_zone = data.aws_availability_zones.available.names[0]
}

resource "aws_subnet" "test2" {
  vpc_id            = aws_vpc.test.id
  cidr_block        = "10.0.2.0/24"
  availability_zone = data.aws_availability_zones.available.names[1]
}
`)
}

func testAccBackupLustreBaseConfig(rName string) string {
	return acctest.ConfigCompose(testAccBackupBaseConfig(), fmt.Sprintf(`
resource "aws_fsx_lustre_file_system" "test" {
  storage_capacity            = 1200
  subnet_ids                  = [aws_subnet.test1.id]
  deployment_type             = "PERSISTENT_1"
  per_unit_storage_throughput = 50

  tags = {
    Name = %[1]q
  }
}
`, rName))
}

func testAccBackupONTAPBaseConfig(rName string, vName string) string {
	return acctest.ConfigCompose(testAccBackupBaseConfig(), fmt.Sprintf(`
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

resource "aws_fsx_ontap_storage_virtual_machine" "test" {
  file_system_id = aws_fsx_ontap_file_system.test.id
  name           = %[1]q
}

resource "aws_fsx_ontap_volume" "test" {
  name                       = %[2]q
  junction_path              = "/%[1]s"
  size_in_megabytes          = 1024
  storage_efficiency_enabled = true
  storage_virtual_machine_id = aws_fsx_ontap_storage_virtual_machine.test.id
}
`, rName, vName))
}

func testAccBackupOpenzfsBaseConfig(rName string) string {
	return acctest.ConfigCompose(testAccBackupBaseConfig(), fmt.Sprintf(`
resource "aws_fsx_openzfs_file_system" "test" {
  storage_capacity    = 64
  subnet_ids          = [aws_subnet.test1.id]
  deployment_type     = "SINGLE_AZ_1"
  throughput_capacity = 64


  tags = {
    Name = %[1]q
  }
}
`, rName))
}

func testAccBackupWindowsBaseConfig(rName string) string {
	return acctest.ConfigCompose(testAccBackupBaseConfig(), fmt.Sprintf(`
resource "aws_directory_service_directory" "test" {
  edition  = "Standard"
  name     = "corp.notexample.com"
  password = "SuperSecretPassw0rd"
  type     = "MicrosoftAD"

  vpc_settings {
    subnet_ids = [aws_subnet.test1.id, aws_subnet.test2.id]
    vpc_id     = aws_vpc.test.id
  }
}

resource "aws_fsx_windows_file_system" "test" {
  active_directory_id             = aws_directory_service_directory.test.id
  automatic_backup_retention_days = 0
  skip_final_backup               = true
  storage_capacity                = 32
  subnet_ids                      = [aws_subnet.test1.id]
  throughput_capacity             = 8

  tags = {
    Name = %[1]q
  }
}
`, rName))
}

func testAccBackupBasicConfig(rName string) string {
	return acctest.ConfigCompose(testAccBackupLustreBaseConfig(rName), fmt.Sprintf(`
resource "aws_fsx_backup" "test" {
  file_system_id = aws_fsx_lustre_file_system.test.id

  tags = {
    Name = %[1]q
  }
}
`, rName))
}

func testAccBackupONTAPBasicConfig(rName string, vName string) string {
	return acctest.ConfigCompose(testAccBackupONTAPBaseConfig(rName, vName), fmt.Sprintf(`
resource "aws_fsx_backup" "test" {
  volume_id = aws_fsx_ontap_volume.test.id

  tags = {
    Name = %[1]q
  }
}
`, rName))
}

func testAccBackupOpenzfsBasicConfig(rName string) string {
	return acctest.ConfigCompose(testAccBackupOpenzfsBaseConfig(rName), fmt.Sprintf(`
resource "aws_fsx_backup" "test" {
  file_system_id = aws_fsx_openzfs_file_system.test.id

  tags = {
    Name = %[1]q
  }
}
`, rName))
}

func testAccBackupWindowsBasicConfig(rName string) string {
	return acctest.ConfigCompose(testAccBackupWindowsBaseConfig(rName), fmt.Sprintf(`
resource "aws_fsx_backup" "test" {
  file_system_id = aws_fsx_windows_file_system.test.id

  tags = {
    Name = %[1]q
  }
}
`, rName))
}

func testAccBackupTags1Config(rName string, tagKey1, tagValue1 string) string {
	return acctest.ConfigCompose(testAccBackupLustreBaseConfig(rName), fmt.Sprintf(`
resource "aws_fsx_backup" "test" {
  file_system_id = aws_fsx_lustre_file_system.test.id

  tags = {
    %[1]q = %[2]q
  }
}
`, tagKey1, tagValue1))
}

func testAccBackupTags2Config(rName string, tagKey1, tagValue1, tagKey2, tagValue2 string) string {
	return acctest.ConfigCompose(testAccBackupLustreBaseConfig(rName), fmt.Sprintf(`
resource "aws_fsx_backup" "test" {
  file_system_id = aws_fsx_lustre_file_system.test.id

  tags = {
    %[1]q = %[2]q
    %[3]q = %[4]q
  }
}
`, tagKey1, tagValue1, tagKey2, tagValue2))
}

func testAccBackupImplictTagsConfig(tagKey1, tagValue1 string) string {
	return acctest.ConfigCompose(acctest.ConfigAvailableAZsNoOptIn(), fmt.Sprintf(`
resource "aws_vpc" "test" {
  cidr_block = "10.0.0.0/16"
}

resource "aws_subnet" "test1" {
  vpc_id            = aws_vpc.test.id
  cidr_block        = "10.0.1.0/24"
  availability_zone = data.aws_availability_zones.available.names[0]
}

resource "aws_fsx_lustre_file_system" "test" {
  storage_capacity            = 1200
  subnet_ids                  = [aws_subnet.test1.id]
  deployment_type             = "PERSISTENT_1"
  per_unit_storage_throughput = 50
  copy_tags_to_backups        = true

  tags = {
    %[1]q = %[2]q
  }
}

resource "aws_fsx_backup" "test" {
  file_system_id = aws_fsx_lustre_file_system.test.id
}
`, tagKey1, tagValue1))
}
