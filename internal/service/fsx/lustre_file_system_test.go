package fsx_test

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/endpoints"
	"github.com/aws/aws-sdk-go/service/fsx"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tffsx "github.com/hashicorp/terraform-provider-aws/internal/service/fsx"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

func TestAccFSxLustreFileSystem_basic(t *testing.T) {
	var filesystem fsx.FileSystem
	resourceName := "aws_fsx_lustre_file_system.test"

	deploymentType := fsx.LustreDeploymentTypeScratch1
	if acctest.Partition() == endpoints.AwsUsGovPartitionID {
		deploymentType = fsx.LustreDeploymentTypeScratch2 // SCRATCH_1 not supported in GovCloud
	}

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); acctest.PreCheckPartitionHasService(fsx.EndpointsID, t) },
		ErrorCheck:        acctest.ErrorCheck(t, fsx.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckLustreFileSystemDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccLustreFileSystemSubnetIds1Config(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLustreFileSystemExists(resourceName, &filesystem),
					acctest.MatchResourceAttrRegionalARN(resourceName, "arn", "fsx", regexp.MustCompile(`file-system/fs-.+`)),
					resource.TestMatchResourceAttr(resourceName, "dns_name", regexp.MustCompile(`fs-.+\.fsx\.`)),
					resource.TestCheckResourceAttr(resourceName, "export_path", ""),
					resource.TestCheckResourceAttr(resourceName, "import_path", ""),
					resource.TestCheckResourceAttr(resourceName, "imported_file_chunk_size", "0"),
					resource.TestCheckResourceAttrSet(resourceName, "mount_name"),
					resource.TestCheckResourceAttr(resourceName, "network_interface_ids.#", "2"),
					acctest.CheckResourceAttrAccountID(resourceName, "owner_id"),
					resource.TestCheckResourceAttr(resourceName, "storage_capacity", "1200"),
					resource.TestCheckResourceAttr(resourceName, "subnet_ids.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "security_group_ids.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
					resource.TestMatchResourceAttr(resourceName, "vpc_id", regexp.MustCompile(`^vpc-.+`)),
					resource.TestMatchResourceAttr(resourceName, "weekly_maintenance_start_time", regexp.MustCompile(`^\d:\d\d:\d\d$`)),
					resource.TestCheckResourceAttr(resourceName, "deployment_type", deploymentType),
					resource.TestCheckResourceAttr(resourceName, "automatic_backup_retention_days", "0"),
					resource.TestCheckResourceAttr(resourceName, "storage_type", fsx.StorageTypeSsd),
					resource.TestCheckResourceAttr(resourceName, "copy_tags_to_backups", "false"),
					resource.TestCheckResourceAttr(resourceName, "data_compression_type", fsx.DataCompressionTypeNone),
					resource.TestCheckResourceAttr(resourceName, "log_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "log_configuration.0.level", "DISABLED"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"security_group_ids"},
			},
		},
	})
}

func TestAccFSxLustreFileSystem_disappears(t *testing.T) {
	var filesystem fsx.FileSystem
	resourceName := "aws_fsx_lustre_file_system.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); acctest.PreCheckPartitionHasService(fsx.EndpointsID, t) },
		ErrorCheck:        acctest.ErrorCheck(t, fsx.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckLustreFileSystemDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccLustreFileSystemSubnetIds1Config(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLustreFileSystemExists(resourceName, &filesystem),
					acctest.CheckResourceDisappears(acctest.Provider, tffsx.ResourceLustreFileSystem(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccFSxLustreFileSystem_dataCompression(t *testing.T) {
	var filesystem fsx.FileSystem
	resourceName := "aws_fsx_lustre_file_system.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); acctest.PreCheckPartitionHasService(fsx.EndpointsID, t) },
		ErrorCheck:        acctest.ErrorCheck(t, fsx.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckLustreFileSystemDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccLustreFileSystemCompressionConfig(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLustreFileSystemExists(resourceName, &filesystem),
					resource.TestCheckResourceAttr(resourceName, "data_compression_type", fsx.DataCompressionTypeLz4),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"security_group_ids"},
			},
			{
				Config: testAccLustreFileSystemSubnetIds1Config(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLustreFileSystemExists(resourceName, &filesystem),
					resource.TestCheckResourceAttr(resourceName, "data_compression_type", fsx.DataCompressionTypeNone),
				),
			},
			{
				Config: testAccLustreFileSystemCompressionConfig(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLustreFileSystemExists(resourceName, &filesystem),
					resource.TestCheckResourceAttr(resourceName, "data_compression_type", fsx.DataCompressionTypeLz4),
				),
			},
		},
	})
}

func TestAccFSxLustreFileSystem_exportPath(t *testing.T) {
	var filesystem1, filesystem2 fsx.FileSystem
	resourceName := "aws_fsx_lustre_file_system.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); acctest.PreCheckPartitionHasService(fsx.EndpointsID, t) },
		ErrorCheck:        acctest.ErrorCheck(t, fsx.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckLustreFileSystemDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccLustreFileSystemExportPathConfig(rName, ""),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLustreFileSystemExists(resourceName, &filesystem1),
					resource.TestCheckResourceAttr(resourceName, "export_path", fmt.Sprintf("s3://%s", rName)),
					resource.TestCheckResourceAttr(resourceName, "auto_import_policy", "NONE"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"security_group_ids"},
			},
			{
				Config: testAccLustreFileSystemExportPathConfig(rName, "/prefix/"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLustreFileSystemExists(resourceName, &filesystem2),
					testAccCheckLustreFileSystemRecreated(&filesystem1, &filesystem2),
					resource.TestCheckResourceAttr(resourceName, "export_path", fmt.Sprintf("s3://%s/prefix/", rName)),
					resource.TestCheckResourceAttr(resourceName, "auto_import_policy", "NONE"),
				),
			},
		},
	})
}

// lintignore: AT002
func TestAccFSxLustreFileSystem_importPath(t *testing.T) {
	var filesystem1, filesystem2 fsx.FileSystem
	resourceName := "aws_fsx_lustre_file_system.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); acctest.PreCheckPartitionHasService(fsx.EndpointsID, t) },
		ErrorCheck:        acctest.ErrorCheck(t, fsx.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckLustreFileSystemDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccLustreFileSystemImportPathConfig(rName, ""),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLustreFileSystemExists(resourceName, &filesystem1),
					resource.TestCheckResourceAttr(resourceName, "import_path", fmt.Sprintf("s3://%s", rName)),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"security_group_ids"},
			},
			{
				Config: testAccLustreFileSystemImportPathConfig(rName, "/prefix/"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLustreFileSystemExists(resourceName, &filesystem2),
					testAccCheckLustreFileSystemRecreated(&filesystem1, &filesystem2),
					resource.TestCheckResourceAttr(resourceName, "import_path", fmt.Sprintf("s3://%s/prefix/", rName)),
				),
			},
		},
	})
}

// lintignore: AT002
func TestAccFSxLustreFileSystem_importedFileChunkSize(t *testing.T) {
	var filesystem1, filesystem2 fsx.FileSystem
	resourceName := "aws_fsx_lustre_file_system.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); acctest.PreCheckPartitionHasService(fsx.EndpointsID, t) },
		ErrorCheck:        acctest.ErrorCheck(t, fsx.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckLustreFileSystemDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccLustreFileSystemImportedFileChunkSizeConfig(rName, 2048),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLustreFileSystemExists(resourceName, &filesystem1),
					resource.TestCheckResourceAttr(resourceName, "imported_file_chunk_size", "2048"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"security_group_ids"},
			},
			{
				Config: testAccLustreFileSystemImportedFileChunkSizeConfig(rName, 4096),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLustreFileSystemExists(resourceName, &filesystem2),
					testAccCheckLustreFileSystemRecreated(&filesystem1, &filesystem2),
					resource.TestCheckResourceAttr(resourceName, "imported_file_chunk_size", "4096"),
				),
			},
		},
	})
}

func TestAccFSxLustreFileSystem_securityGroupIDs(t *testing.T) {
	var filesystem1, filesystem2 fsx.FileSystem
	resourceName := "aws_fsx_lustre_file_system.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); acctest.PreCheckPartitionHasService(fsx.EndpointsID, t) },
		ErrorCheck:        acctest.ErrorCheck(t, fsx.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckLustreFileSystemDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccLustreFileSystemSecurityGroupIds1Config(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLustreFileSystemExists(resourceName, &filesystem1),
					resource.TestCheckResourceAttr(resourceName, "security_group_ids.#", "1"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"security_group_ids"},
			},
			{
				Config: testAccLustreFileSystemSecurityGroupIds2Config(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLustreFileSystemExists(resourceName, &filesystem2),
					testAccCheckLustreFileSystemRecreated(&filesystem1, &filesystem2),
					resource.TestCheckResourceAttr(resourceName, "security_group_ids.#", "2"),
				),
			},
		},
	})
}

func TestAccFSxLustreFileSystem_storageCapacity(t *testing.T) {
	var filesystem1, filesystem2 fsx.FileSystem
	resourceName := "aws_fsx_lustre_file_system.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); acctest.PreCheckPartitionHasService(fsx.EndpointsID, t) },
		ErrorCheck:        acctest.ErrorCheck(t, fsx.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckLustreFileSystemDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccLustreFileSystemStorageCapacityConfig(7200),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLustreFileSystemExists(resourceName, &filesystem1),
					resource.TestCheckResourceAttr(resourceName, "storage_capacity", "7200"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"security_group_ids"},
			},
			{
				Config: testAccLustreFileSystemStorageCapacityConfig(1200),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLustreFileSystemExists(resourceName, &filesystem2),
					testAccCheckLustreFileSystemRecreated(&filesystem1, &filesystem2),
					resource.TestCheckResourceAttr(resourceName, "storage_capacity", "1200"),
				),
			},
		},
	})
}

func TestAccFSxLustreFileSystem_storageCapacityUpdate(t *testing.T) {
	var filesystem1, filesystem2, filesystem3 fsx.FileSystem
	resourceName := "aws_fsx_lustre_file_system.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); acctest.PreCheckPartitionHasService(fsx.EndpointsID, t) },
		ErrorCheck:        acctest.ErrorCheck(t, fsx.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckLustreFileSystemDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccLustreFileSystemStorageCapacityScratch2Config(7200),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLustreFileSystemExists(resourceName, &filesystem1),
					resource.TestCheckResourceAttr(resourceName, "storage_capacity", "7200"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"security_group_ids"},
			},
			{
				Config: testAccLustreFileSystemStorageCapacityScratch2Config(1200),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLustreFileSystemExists(resourceName, &filesystem2),
					testAccCheckLustreFileSystemRecreated(&filesystem1, &filesystem2),
					resource.TestCheckResourceAttr(resourceName, "storage_capacity", "1200"),
				),
			},
			{
				Config: testAccLustreFileSystemStorageCapacityScratch2Config(7200),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLustreFileSystemExists(resourceName, &filesystem3),
					testAccCheckLustreFileSystemNotRecreated(&filesystem2, &filesystem3),
					resource.TestCheckResourceAttr(resourceName, "storage_capacity", "7200"),
				),
			},
		},
	})
}

func TestAccFSxLustreFileSystem_fileSystemTypeVersion(t *testing.T) {
	var filesystem1, filesystem2 fsx.FileSystem
	resourceName := "aws_fsx_lustre_file_system.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); acctest.PreCheckPartitionHasService(fsx.EndpointsID, t) },
		ErrorCheck:        acctest.ErrorCheck(t, fsx.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckLustreFileSystemDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccLustreFileSystemFileSystemTypeVersionConfig("2.10"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLustreFileSystemExists(resourceName, &filesystem1),
					resource.TestCheckResourceAttr(resourceName, "file_system_type_version", "2.10"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"security_group_ids"},
			},
			{
				Config: testAccLustreFileSystemFileSystemTypeVersionConfig("2.12"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLustreFileSystemExists(resourceName, &filesystem2),
					testAccCheckLustreFileSystemRecreated(&filesystem1, &filesystem2),
					resource.TestCheckResourceAttr(resourceName, "file_system_type_version", "2.12"),
				),
			},
		},
	})
}

func TestAccFSxLustreFileSystem_tags(t *testing.T) {
	var filesystem1, filesystem2, filesystem3 fsx.FileSystem
	resourceName := "aws_fsx_lustre_file_system.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); acctest.PreCheckPartitionHasService(fsx.EndpointsID, t) },
		ErrorCheck:        acctest.ErrorCheck(t, fsx.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckLustreFileSystemDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccLustreFileSystemTags1Config("key1", "value1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLustreFileSystemExists(resourceName, &filesystem1),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"security_group_ids"},
			},
			{
				Config: testAccLustreFileSystemTags2Config("key1", "value1updated", "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLustreFileSystemExists(resourceName, &filesystem2),
					testAccCheckLustreFileSystemNotRecreated(&filesystem1, &filesystem2),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1updated"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
			{
				Config: testAccLustreFileSystemTags1Config("key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLustreFileSystemExists(resourceName, &filesystem3),
					testAccCheckLustreFileSystemNotRecreated(&filesystem2, &filesystem3),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
		},
	})
}

func TestAccFSxLustreFileSystem_weeklyMaintenanceStartTime(t *testing.T) {
	var filesystem1, filesystem2 fsx.FileSystem
	resourceName := "aws_fsx_lustre_file_system.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); acctest.PreCheckPartitionHasService(fsx.EndpointsID, t) },
		ErrorCheck:        acctest.ErrorCheck(t, fsx.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckLustreFileSystemDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccLustreFileSystemWeeklyMaintenanceStartTimeConfig("1:01:01"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLustreFileSystemExists(resourceName, &filesystem1),
					resource.TestCheckResourceAttr(resourceName, "weekly_maintenance_start_time", "1:01:01"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"security_group_ids"},
			},
			{
				Config: testAccLustreFileSystemWeeklyMaintenanceStartTimeConfig("2:02:02"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLustreFileSystemExists(resourceName, &filesystem2),
					testAccCheckLustreFileSystemNotRecreated(&filesystem1, &filesystem2),
					resource.TestCheckResourceAttr(resourceName, "weekly_maintenance_start_time", "2:02:02"),
				),
			},
		},
	})
}

func TestAccFSxLustreFileSystem_automaticBackupRetentionDays(t *testing.T) {
	var filesystem1, filesystem2 fsx.FileSystem
	resourceName := "aws_fsx_lustre_file_system.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); acctest.PreCheckPartitionHasService(fsx.EndpointsID, t) },
		ErrorCheck:        acctest.ErrorCheck(t, fsx.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckLustreFileSystemDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccLustreFileSystemAutomaticBackupRetentionDaysConfig(90),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLustreFileSystemExists(resourceName, &filesystem1),
					resource.TestCheckResourceAttr(resourceName, "automatic_backup_retention_days", "90"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"security_group_ids"},
			},
			{
				Config: testAccLustreFileSystemAutomaticBackupRetentionDaysConfig(0),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLustreFileSystemExists(resourceName, &filesystem2),
					testAccCheckLustreFileSystemNotRecreated(&filesystem1, &filesystem2),
					resource.TestCheckResourceAttr(resourceName, "automatic_backup_retention_days", "0"),
				),
			},
			{
				Config: testAccLustreFileSystemAutomaticBackupRetentionDaysConfig(1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLustreFileSystemExists(resourceName, &filesystem1),
					resource.TestCheckResourceAttr(resourceName, "automatic_backup_retention_days", "1"),
				),
			},
		},
	})
}

func TestAccFSxLustreFileSystem_dailyAutomaticBackupStartTime(t *testing.T) {
	var filesystem1, filesystem2 fsx.FileSystem
	resourceName := "aws_fsx_lustre_file_system.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); acctest.PreCheckPartitionHasService(fsx.EndpointsID, t) },
		ErrorCheck:        acctest.ErrorCheck(t, fsx.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckLustreFileSystemDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccLustreFileSystemDailyAutomaticBackupStartTimeConfig("01:01"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLustreFileSystemExists(resourceName, &filesystem1),
					resource.TestCheckResourceAttr(resourceName, "daily_automatic_backup_start_time", "01:01"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"security_group_ids"},
			},
			{
				Config: testAccLustreFileSystemDailyAutomaticBackupStartTimeConfig("02:02"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLustreFileSystemExists(resourceName, &filesystem2),
					testAccCheckLustreFileSystemNotRecreated(&filesystem1, &filesystem2),
					resource.TestCheckResourceAttr(resourceName, "daily_automatic_backup_start_time", "02:02"),
				),
			},
		},
	})
}

func TestAccFSxLustreFileSystem_deploymentTypePersistent1(t *testing.T) {
	var filesystem fsx.FileSystem
	resourceName := "aws_fsx_lustre_file_system.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); acctest.PreCheckPartitionHasService(fsx.EndpointsID, t) },
		ErrorCheck:        acctest.ErrorCheck(t, fsx.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckLustreFileSystemDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccLustreFileSystemPersistent1DeploymentType(50),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLustreFileSystemExists(resourceName, &filesystem),
					// per_unit_storage_throughput=50 is only available with deployment_type=PERSISTENT_1, so we test both here.
					resource.TestCheckResourceAttr(resourceName, "per_unit_storage_throughput", "50"),
					resource.TestCheckResourceAttr(resourceName, "deployment_type", fsx.LustreDeploymentTypePersistent1),
					resource.TestCheckResourceAttr(resourceName, "automatic_backup_retention_days", "0"),
					acctest.MatchResourceAttrRegionalARN(resourceName, "kms_key_id", "kms", regexp.MustCompile(`key/.+`)),
					// We don't know the randomly generated mount_name ahead of time like for SCRATCH_1 deployment types.
					resource.TestCheckResourceAttrSet(resourceName, "mount_name"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"security_group_ids"},
			},
		},
	})
}

func TestAccFSxLustreFileSystem_deploymentTypePersistent2(t *testing.T) {
	var filesystem fsx.FileSystem
	resourceName := "aws_fsx_lustre_file_system.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); acctest.PreCheckPartitionHasService(fsx.EndpointsID, t) },
		ErrorCheck:        acctest.ErrorCheck(t, fsx.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckLustreFileSystemDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccLustreFileSystemPersistent2DeploymentType(125),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLustreFileSystemExists(resourceName, &filesystem),
					// per_unit_storage_throughput=125 is only available with deployment_type=PERSISTENT_2, so we test both here.
					resource.TestCheckResourceAttr(resourceName, "per_unit_storage_throughput", "125"),
					resource.TestCheckResourceAttr(resourceName, "deployment_type", fsx.LustreDeploymentTypePersistent2),
					resource.TestCheckResourceAttr(resourceName, "automatic_backup_retention_days", "0"),
					acctest.MatchResourceAttrRegionalARN(resourceName, "kms_key_id", "kms", regexp.MustCompile(`key/.+`)),
					// We don't know the randomly generated mount_name ahead of time like for SCRATCH_1 deployment types.
					resource.TestCheckResourceAttrSet(resourceName, "mount_name"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"security_group_ids"},
			},
		},
	})
}

func TestAccFSxLustreFileSystem_logConfig(t *testing.T) {
	var filesystem fsx.FileSystem
	resourceName := "aws_fsx_lustre_file_system.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); acctest.PreCheckPartitionHasService(fsx.EndpointsID, t) },
		ErrorCheck:        acctest.ErrorCheck(t, fsx.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckLustreFileSystemDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccLustreFileSystemLogConfig(rName, "WARN_ONLY"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLustreFileSystemExists(resourceName, &filesystem),
					resource.TestCheckResourceAttr(resourceName, "log_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "log_configuration.0.level", "WARN_ONLY"),
					resource.TestCheckResourceAttrSet(resourceName, "log_configuration.0.destination"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"security_group_ids"},
			},
			{
				Config: testAccLustreFileSystemLogConfig(rName, "ERROR_ONLY"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLustreFileSystemExists(resourceName, &filesystem),
					resource.TestCheckResourceAttr(resourceName, "log_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "log_configuration.0.level", "ERROR_ONLY"),
					resource.TestCheckResourceAttrSet(resourceName, "log_configuration.0.destination"),
				),
			},
		},
	})
}

func TestAccFSxLustreFileSystem_fromBackup(t *testing.T) {
	var filesystem fsx.FileSystem
	resourceName := "aws_fsx_lustre_file_system.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); acctest.PreCheckPartitionHasService(fsx.EndpointsID, t) },
		ErrorCheck:        acctest.ErrorCheck(t, fsx.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckLustreFileSystemDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccLustreFileSystemFromBackup(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLustreFileSystemExists(resourceName, &filesystem),
					resource.TestCheckResourceAttr(resourceName, "per_unit_storage_throughput", "50"),
					resource.TestCheckResourceAttr(resourceName, "deployment_type", fsx.LustreDeploymentTypePersistent1),
					resource.TestCheckResourceAttrPair(resourceName, "backup_id", "aws_fsx_backup.test", "id"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"security_group_ids", "backup_id"},
			},
		},
	})
}

func TestAccFSxLustreFileSystem_kmsKeyID(t *testing.T) {
	var filesystem1, filesystem2 fsx.FileSystem
	resourceName := "aws_fsx_lustre_file_system.test"
	kmsKeyResourceName1 := "aws_kms_key.test1"
	kmsKeyResourceName2 := "aws_kms_key.test2"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); acctest.PreCheckPartitionHasService(fsx.EndpointsID, t) },
		ErrorCheck:        acctest.ErrorCheck(t, fsx.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckLustreFileSystemDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccLustreFileSystemKMSKeyId1Config(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLustreFileSystemExists(resourceName, &filesystem1),
					resource.TestCheckResourceAttr(resourceName, "deployment_type", fsx.LustreDeploymentTypePersistent1),
					resource.TestCheckResourceAttrPair(resourceName, "kms_key_id", kmsKeyResourceName1, "arn"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"security_group_ids"},
			},
			{
				Config: testAccLustreFileSystemKMSKeyId2Config(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLustreFileSystemExists(resourceName, &filesystem2),
					resource.TestCheckResourceAttr(resourceName, "deployment_type", fsx.LustreDeploymentTypePersistent1),
					testAccCheckWindowsFileSystemRecreated(&filesystem1, &filesystem2),
					resource.TestCheckResourceAttrPair(resourceName, "kms_key_id", kmsKeyResourceName2, "arn"),
				),
			},
		},
	})
}

func TestAccFSxLustreFileSystem_deploymentTypeScratch2(t *testing.T) {
	var filesystem fsx.FileSystem
	resourceName := "aws_fsx_lustre_file_system.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); acctest.PreCheckPartitionHasService(fsx.EndpointsID, t) },
		ErrorCheck:        acctest.ErrorCheck(t, fsx.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckLustreFileSystemDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccLustreFileSystemDeploymentType(fsx.LustreDeploymentTypeScratch2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLustreFileSystemExists(resourceName, &filesystem),
					resource.TestCheckResourceAttr(resourceName, "deployment_type", fsx.LustreDeploymentTypeScratch2),
					// We don't know the randomly generated mount_name ahead of time like for SCRATCH_1 deployment types.
					resource.TestCheckResourceAttrSet(resourceName, "mount_name"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"security_group_ids"},
			},
		},
	})
}

func TestAccFSxLustreFileSystem_storageTypeHddDriveCacheRead(t *testing.T) {
	var filesystem fsx.FileSystem
	resourceName := "aws_fsx_lustre_file_system.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); acctest.PreCheckPartitionHasService(fsx.EndpointsID, t) },
		ErrorCheck:        acctest.ErrorCheck(t, fsx.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckLustreFileSystemDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccLustreFileSystemHddStorageType(fsx.DriveCacheTypeRead),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLustreFileSystemExists(resourceName, &filesystem),
					resource.TestCheckResourceAttr(resourceName, "storage_type", fsx.StorageTypeHdd),
					resource.TestCheckResourceAttr(resourceName, "drive_cache_type", fsx.DriveCacheTypeRead),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"security_group_ids"},
			},
		},
	})
}

func TestAccFSxLustreFileSystem_storageTypeHddDriveCacheNone(t *testing.T) {
	var filesystem fsx.FileSystem
	resourceName := "aws_fsx_lustre_file_system.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); acctest.PreCheckPartitionHasService(fsx.EndpointsID, t) },
		ErrorCheck:        acctest.ErrorCheck(t, fsx.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckLustreFileSystemDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccLustreFileSystemHddStorageType(fsx.DriveCacheTypeNone),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLustreFileSystemExists(resourceName, &filesystem),
					resource.TestCheckResourceAttr(resourceName, "storage_type", fsx.StorageTypeHdd),
					resource.TestCheckResourceAttr(resourceName, "drive_cache_type", fsx.DriveCacheTypeNone),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"security_group_ids"},
			},
		},
	})
}

func TestAccFSxLustreFileSystem_copyTagsToBackups(t *testing.T) {
	var filesystem fsx.FileSystem
	resourceName := "aws_fsx_lustre_file_system.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); acctest.PreCheckPartitionHasService(fsx.EndpointsID, t) },
		ErrorCheck:        acctest.ErrorCheck(t, fsx.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckLustreFileSystemDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccLustreFileSystemCopyTagsToBackups(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLustreFileSystemExists(resourceName, &filesystem),
					resource.TestCheckResourceAttr(resourceName, "copy_tags_to_backups", "true"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"security_group_ids"},
			},
		},
	})
}

func TestAccFSxLustreFileSystem_autoImportPolicy(t *testing.T) {
	var filesystem fsx.FileSystem
	resourceName := "aws_fsx_lustre_file_system.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); acctest.PreCheckPartitionHasService(fsx.EndpointsID, t) },
		ErrorCheck:        acctest.ErrorCheck(t, fsx.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckLustreFileSystemDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccLustreFileSystemAutoImportPolicyConfig(rName, "", "NEW"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLustreFileSystemExists(resourceName, &filesystem),
					resource.TestCheckResourceAttr(resourceName, "auto_import_policy", "NEW"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"security_group_ids"},
			},
			{
				Config: testAccLustreFileSystemAutoImportPolicyConfig(rName, "", "NEW_CHANGED"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLustreFileSystemExists(resourceName, &filesystem),
					resource.TestCheckResourceAttr(resourceName, "auto_import_policy", "NEW_CHANGED"),
				),
			},
		},
	})
}

func testAccCheckLustreFileSystemExists(resourceName string, fs *fsx.FileSystem) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("Not found: %s", resourceName)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).FSxConn

		filesystem, err := tffsx.FindFileSystemByID(conn, rs.Primary.ID)
		if err != nil {
			return err
		}

		if filesystem == nil {
			return fmt.Errorf("FSx Lustre File System (%s) not found", rs.Primary.ID)
		}

		*fs = *filesystem

		return nil
	}
}

func testAccCheckLustreFileSystemDestroy(s *terraform.State) error {
	conn := acctest.Provider.Meta().(*conns.AWSClient).FSxConn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_fsx_lustre_file_system" {
			continue
		}

		filesystem, err := tffsx.FindFileSystemByID(conn, rs.Primary.ID)
		if tfresource.NotFound(err) {
			continue
		}

		if filesystem != nil {
			return fmt.Errorf("FSx Lustre File System (%s) still exists", rs.Primary.ID)
		}
	}
	return nil
}

func testAccCheckLustreFileSystemNotRecreated(i, j *fsx.FileSystem) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if aws.StringValue(i.FileSystemId) != aws.StringValue(j.FileSystemId) {
			return fmt.Errorf("FSx File System (%s) recreated", aws.StringValue(i.FileSystemId))
		}

		return nil
	}
}

func testAccCheckLustreFileSystemRecreated(i, j *fsx.FileSystem) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if aws.StringValue(i.FileSystemId) == aws.StringValue(j.FileSystemId) {
			return fmt.Errorf("FSx File System (%s) not recreated", aws.StringValue(i.FileSystemId))
		}

		return nil
	}
}

func testAccLustreFileSystemBaseConfig() string {
	return acctest.ConfigCompose(acctest.ConfigAvailableAZsNoOptIn(), `
data "aws_partition" "current" {}

resource "aws_vpc" "test" {
  cidr_block = "10.0.0.0/16"
}

resource "aws_subnet" "test1" {
  vpc_id            = aws_vpc.test.id
  cidr_block        = "10.0.1.0/24"
  availability_zone = data.aws_availability_zones.available.names[0]
}
`)
}

func testAccLustreFileSystemExportPathConfig(rName, exportPrefix string) string {
	return acctest.ConfigCompose(testAccLustreFileSystemBaseConfig(), fmt.Sprintf(`
resource "aws_s3_bucket" "test" {
  bucket = %[1]q
}

resource "aws_s3_bucket_acl" "test" {
  bucket = aws_s3_bucket.test.id
  acl    = "private"
}

resource "aws_fsx_lustre_file_system" "test" {
  export_path      = "s3://${aws_s3_bucket.test.bucket}%[2]s"
  import_path      = "s3://${aws_s3_bucket.test.bucket}"
  storage_capacity = 1200
  subnet_ids       = [aws_subnet.test1.id]
  deployment_type  = data.aws_partition.current.partition == "aws-us-gov" ? "SCRATCH_2" : null # GovCloud does not support SCRATCH_1
}
`, rName, exportPrefix))
}

func testAccLustreFileSystemImportPathConfig(rName, importPrefix string) string {
	return acctest.ConfigCompose(testAccLustreFileSystemBaseConfig(), fmt.Sprintf(`
resource "aws_s3_bucket" "test" {
  bucket = %[1]q
}

resource "aws_s3_bucket_acl" "test" {
  bucket = aws_s3_bucket.test.id
  acl    = "private"
}

resource "aws_fsx_lustre_file_system" "test" {
  import_path      = "s3://${aws_s3_bucket.test.bucket}%[2]s"
  storage_capacity = 1200
  subnet_ids       = [aws_subnet.test1.id]
  deployment_type  = data.aws_partition.current.partition == "aws-us-gov" ? "SCRATCH_2" : null # GovCloud does not support SCRATCH_1
}
`, rName, importPrefix))
}

func testAccLustreFileSystemImportedFileChunkSizeConfig(rName string, importedFileChunkSize int) string {
	return acctest.ConfigCompose(testAccLustreFileSystemBaseConfig(), fmt.Sprintf(`
resource "aws_s3_bucket" "test" {
  bucket = %[1]q
}

resource "aws_s3_bucket_acl" "test" {
  bucket = aws_s3_bucket.test.id
  acl    = "private"
}

resource "aws_fsx_lustre_file_system" "test" {
  import_path              = "s3://${aws_s3_bucket.test.bucket}"
  imported_file_chunk_size = %[2]d
  storage_capacity         = 1200
  subnet_ids               = [aws_subnet.test1.id]
  deployment_type          = data.aws_partition.current.partition == "aws-us-gov" ? "SCRATCH_2" : null # GovCloud does not support SCRATCH_1
}
`, rName, importedFileChunkSize))
}

func testAccLustreFileSystemSecurityGroupIds1Config() string {
	return acctest.ConfigCompose(testAccLustreFileSystemBaseConfig(), `
resource "aws_security_group" "test1" {
  description = "security group for FSx testing"
  vpc_id      = aws_vpc.test.id

  ingress {
    cidr_blocks = [aws_vpc.test.cidr_block]
    from_port   = 0
    protocol    = -1
    to_port     = 0
  }

  egress {
    cidr_blocks = ["0.0.0.0/0"]
    from_port   = 0
    protocol    = "-1"
    to_port     = 0
  }
}

resource "aws_fsx_lustre_file_system" "test" {
  security_group_ids = [aws_security_group.test1.id]
  storage_capacity   = 1200
  subnet_ids         = [aws_subnet.test1.id]
  deployment_type    = data.aws_partition.current.partition == "aws-us-gov" ? "SCRATCH_2" : null # GovCloud does not support SCRATCH_1
}
`)
}

func testAccLustreFileSystemSecurityGroupIds2Config() string {
	return acctest.ConfigCompose(testAccLustreFileSystemBaseConfig(), `
resource "aws_security_group" "test1" {
  description = "security group for FSx testing"
  vpc_id      = aws_vpc.test.id

  ingress {
    cidr_blocks = [aws_vpc.test.cidr_block]
    from_port   = 0
    protocol    = -1
    to_port     = 0
  }

  egress {
    cidr_blocks = ["0.0.0.0/0"]
    from_port   = 0
    protocol    = "-1"
    to_port     = 0
  }
}

resource "aws_security_group" "test2" {
  description = "security group for FSx testing"
  vpc_id      = aws_vpc.test.id

  ingress {
    cidr_blocks = [aws_vpc.test.cidr_block]
    from_port   = 0
    protocol    = -1
    to_port     = 0
  }

  egress {
    cidr_blocks = ["0.0.0.0/0"]
    from_port   = 0
    protocol    = "-1"
    to_port     = 0
  }
}

resource "aws_fsx_lustre_file_system" "test" {
  security_group_ids = [aws_security_group.test1.id, aws_security_group.test2.id]
  storage_capacity   = 1200
  subnet_ids         = [aws_subnet.test1.id]
  deployment_type    = data.aws_partition.current.partition == "aws-us-gov" ? "SCRATCH_2" : null # GovCloud does not support SCRATCH_1
}
`)
}

func testAccLustreFileSystemFileSystemTypeVersionConfig(fileSystemTypeVersion string) string {
	return acctest.ConfigCompose(testAccLustreFileSystemBaseConfig(), fmt.Sprintf(`
resource "aws_fsx_lustre_file_system" "test" {
  file_system_type_version = %[1]q
  storage_capacity         = 1200
  subnet_ids               = [aws_subnet.test1.id]
  deployment_type          = data.aws_partition.current.partition == "aws-us-gov" ? "SCRATCH_2" : null # GovCloud does not support SCRATCH_1
}
`, fileSystemTypeVersion))
}

func testAccLustreFileSystemStorageCapacityConfig(storageCapacity int) string {
	return acctest.ConfigCompose(testAccLustreFileSystemBaseConfig(), fmt.Sprintf(`
resource "aws_fsx_lustre_file_system" "test" {
  storage_capacity = %[1]d
  subnet_ids       = [aws_subnet.test1.id]
  deployment_type  = data.aws_partition.current.partition == "aws-us-gov" ? "SCRATCH_2" : null # GovCloud does not support SCRATCH_1
}
`, storageCapacity))
}

func testAccLustreFileSystemStorageCapacityScratch2Config(storageCapacity int) string {
	return acctest.ConfigCompose(testAccLustreFileSystemBaseConfig(), fmt.Sprintf(`
resource "aws_fsx_lustre_file_system" "test" {
  storage_capacity = %[1]d
  subnet_ids       = [aws_subnet.test1.id]
  deployment_type  = "SCRATCH_2"
}
`, storageCapacity))
}

func testAccLustreFileSystemSubnetIds1Config() string {
	return acctest.ConfigCompose(testAccLustreFileSystemBaseConfig(), `
resource "aws_fsx_lustre_file_system" "test" {
  storage_capacity = 1200
  subnet_ids       = [aws_subnet.test1.id]
  deployment_type  = data.aws_partition.current.partition == "aws-us-gov" ? "SCRATCH_2" : null # GovCloud does not support SCRATCH_1
}
`)
}

func testAccLustreFileSystemTags1Config(tagKey1, tagValue1 string) string {
	return acctest.ConfigCompose(testAccLustreFileSystemBaseConfig(), fmt.Sprintf(`
resource "aws_fsx_lustre_file_system" "test" {
  storage_capacity = 1200
  subnet_ids       = [aws_subnet.test1.id]
  deployment_type  = data.aws_partition.current.partition == "aws-us-gov" ? "SCRATCH_2" : null # GovCloud does not support SCRATCH_1

  tags = {
    %[1]q = %[2]q
  }
}
`, tagKey1, tagValue1))
}

func testAccLustreFileSystemTags2Config(tagKey1, tagValue1, tagKey2, tagValue2 string) string {
	return acctest.ConfigCompose(testAccLustreFileSystemBaseConfig(), fmt.Sprintf(`
resource "aws_fsx_lustre_file_system" "test" {
  storage_capacity = 1200
  subnet_ids       = [aws_subnet.test1.id]
  deployment_type  = data.aws_partition.current.partition == "aws-us-gov" ? "SCRATCH_2" : null # GovCloud does not support SCRATCH_1

  tags = {
    %[1]q = %[2]q
    %[3]q = %[4]q
  }
}
`, tagKey1, tagValue1, tagKey2, tagValue2))
}

func testAccLustreFileSystemWeeklyMaintenanceStartTimeConfig(weeklyMaintenanceStartTime string) string {
	return acctest.ConfigCompose(testAccLustreFileSystemBaseConfig(), fmt.Sprintf(`
resource "aws_fsx_lustre_file_system" "test" {
  storage_capacity              = 1200
  subnet_ids                    = [aws_subnet.test1.id]
  weekly_maintenance_start_time = %[1]q
  deployment_type               = data.aws_partition.current.partition == "aws-us-gov" ? "SCRATCH_2" : null # GovCloud does not support SCRATCH_1
}
`, weeklyMaintenanceStartTime))
}

func testAccLustreFileSystemDailyAutomaticBackupStartTimeConfig(dailyAutomaticBackupStartTime string) string {
	return acctest.ConfigCompose(testAccLustreFileSystemBaseConfig(), fmt.Sprintf(`
resource "aws_fsx_lustre_file_system" "test" {
  storage_capacity                  = 1200
  subnet_ids                        = [aws_subnet.test1.id]
  deployment_type                   = "PERSISTENT_1"
  per_unit_storage_throughput       = 50
  daily_automatic_backup_start_time = %[1]q
  automatic_backup_retention_days   = 1
}
`, dailyAutomaticBackupStartTime))
}

func testAccLustreFileSystemAutomaticBackupRetentionDaysConfig(retention int) string {
	return acctest.ConfigCompose(testAccLustreFileSystemBaseConfig(), fmt.Sprintf(`
resource "aws_fsx_lustre_file_system" "test" {
  storage_capacity                = 1200
  subnet_ids                      = ["${aws_subnet.test1.id}"]
  deployment_type                 = "PERSISTENT_1"
  per_unit_storage_throughput     = 50
  automatic_backup_retention_days = %[1]d
}
`, retention))
}

func testAccLustreFileSystemDeploymentType(deploymentType string) string {
	return acctest.ConfigCompose(testAccLustreFileSystemBaseConfig(), fmt.Sprintf(`
resource "aws_fsx_lustre_file_system" "test" {
  storage_capacity = 1200
  subnet_ids       = [aws_subnet.test1.id]
  deployment_type  = %[1]q
}
`, deploymentType))
}

func testAccLustreFileSystemPersistent1DeploymentType(perUnitStorageThroughput int) string {
	return acctest.ConfigCompose(testAccLustreFileSystemBaseConfig(), fmt.Sprintf(`
resource "aws_fsx_lustre_file_system" "test" {
  storage_capacity            = 1200
  subnet_ids                  = [aws_subnet.test1.id]
  deployment_type             = "PERSISTENT_1"
  per_unit_storage_throughput = %[1]d
}
`, perUnitStorageThroughput))
}

func testAccLustreFileSystemPersistent2DeploymentType(perUnitStorageThroughput int) string {
	return acctest.ConfigCompose(testAccLustreFileSystemBaseConfig(), fmt.Sprintf(`
resource "aws_fsx_lustre_file_system" "test" {
  storage_capacity            = 1200
  subnet_ids                  = [aws_subnet.test1.id]
  deployment_type             = "PERSISTENT_2"
  per_unit_storage_throughput = %[1]d
}
`, perUnitStorageThroughput))
}

func testAccLustreFileSystemFromBackup() string {
	return acctest.ConfigCompose(testAccLustreFileSystemBaseConfig(), `
resource "aws_fsx_lustre_file_system" "base" {
  storage_capacity            = 1200
  subnet_ids                  = [aws_subnet.test1.id]
  deployment_type             = "PERSISTENT_1"
  per_unit_storage_throughput = 50
}

resource "aws_fsx_backup" "test" {
  file_system_id = aws_fsx_lustre_file_system.base.id
}

resource "aws_fsx_lustre_file_system" "test" {
  storage_capacity            = 1200
  subnet_ids                  = [aws_subnet.test1.id]
  deployment_type             = "PERSISTENT_1"
  per_unit_storage_throughput = 50
  backup_id                   = aws_fsx_backup.test.id
}
`)
}

func testAccLustreFileSystemKMSKeyId1Config() string {
	return acctest.ConfigCompose(testAccLustreFileSystemBaseConfig(), `
resource "aws_kms_key" "test1" {
  description             = "FSx KMS Testing key"
  deletion_window_in_days = 7
}

resource "aws_fsx_lustre_file_system" "test" {
  storage_capacity            = 1200
  subnet_ids                  = [aws_subnet.test1.id]
  deployment_type             = "PERSISTENT_1"
  per_unit_storage_throughput = 50
  kms_key_id                  = aws_kms_key.test1.arn
}
`)
}

func testAccLustreFileSystemKMSKeyId2Config() string {
	return acctest.ConfigCompose(testAccLustreFileSystemBaseConfig(), `
resource "aws_kms_key" "test2" {
  description             = "FSx KMS Testing key"
  deletion_window_in_days = 7
}

resource "aws_fsx_lustre_file_system" "test" {
  storage_capacity            = 1200
  subnet_ids                  = [aws_subnet.test1.id]
  deployment_type             = "PERSISTENT_1"
  per_unit_storage_throughput = 50
  kms_key_id                  = aws_kms_key.test2.arn
}
`)
}

func testAccLustreFileSystemHddStorageType(drive_cache_type string) string {
	return acctest.ConfigCompose(testAccLustreFileSystemBaseConfig(), fmt.Sprintf(`
resource "aws_fsx_lustre_file_system" "test" {
  storage_capacity            = 6000
  subnet_ids                  = [aws_subnet.test1.id]
  deployment_type             = "PERSISTENT_1"
  per_unit_storage_throughput = 12
  storage_type                = "HDD"
  drive_cache_type            = %[1]q
}
`, drive_cache_type))
}

func testAccLustreFileSystemAutoImportPolicyConfig(rName, exportPrefix, policy string) string {
	return acctest.ConfigCompose(testAccLustreFileSystemBaseConfig(), fmt.Sprintf(`
resource "aws_s3_bucket" "test" {
  bucket = %[1]q
}

resource "aws_s3_bucket_acl" "test" {
  bucket = aws_s3_bucket.test.id
  acl    = "private"
}

resource "aws_fsx_lustre_file_system" "test" {
  export_path        = "s3://${aws_s3_bucket.test.bucket}%[2]s"
  import_path        = "s3://${aws_s3_bucket.test.bucket}"
  auto_import_policy = %[3]q
  storage_capacity   = 1200
  subnet_ids         = [aws_subnet.test1.id]
  deployment_type    = data.aws_partition.current.partition == "aws-us-gov" ? "SCRATCH_2" : null # GovCloud does not support SCRATCH_1
}
`, rName, exportPrefix, policy))
}

func testAccLustreFileSystemCopyTagsToBackups() string {
	return acctest.ConfigCompose(testAccLustreFileSystemBaseConfig(), `
resource "aws_fsx_lustre_file_system" "test" {
  storage_capacity            = 1200
  deployment_type             = "PERSISTENT_1"
  subnet_ids                  = [aws_subnet.test1.id]
  per_unit_storage_throughput = 50
  copy_tags_to_backups        = true
}
`)
}

func testAccLustreFileSystemCompressionConfig() string {
	return acctest.ConfigCompose(testAccLustreFileSystemBaseConfig(), `
resource "aws_fsx_lustre_file_system" "test" {
  storage_capacity      = 1200
  subnet_ids            = [aws_subnet.test1.id]
  deployment_type       = data.aws_partition.current.partition == "aws-us-gov" ? "SCRATCH_2" : null # GovCloud does not support SCRATCH_1
  data_compression_type = "LZ4"
}
`)
}

func testAccLustreFileSystemLogConfig(rName, status string) string {
	return acctest.ConfigCompose(testAccLustreFileSystemBaseConfig(), fmt.Sprintf(`
resource aws_cloudwatch_log_group "test" {
  name = "/aws/fsx/%[1]s"
}

resource "aws_fsx_lustre_file_system" "test" {
  storage_capacity = 1200
  subnet_ids       = [aws_subnet.test1.id]
  deployment_type  = data.aws_partition.current.partition == "aws-us-gov" ? "SCRATCH_2" : null # GovCloud does not support SCRATCH_1

  log_configuration {
    destination = aws_cloudwatch_log_group.test.arn
    level       = %[2]q
  }
}
`, rName, status))
}
