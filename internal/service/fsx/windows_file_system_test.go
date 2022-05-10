package fsx_test

import (
	"fmt"
	"regexp"
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

func TestAccFSxWindowsFileSystem_basic(t *testing.T) {
	var filesystem fsx.FileSystem
	resourceName := "aws_fsx_windows_file_system.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); acctest.PreCheckPartitionHasService(fsx.EndpointsID, t) },
		ErrorCheck:        acctest.ErrorCheck(t, fsx.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckWindowsFileSystemDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccWindowsFileSystemSubnetIds1Config(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckWindowsFileSystemExists(resourceName, &filesystem),
					acctest.MatchResourceAttrRegionalARN(resourceName, "arn", "fsx", regexp.MustCompile(`file-system/fs-.+`)),
					resource.TestCheckResourceAttr(resourceName, "aliases.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "automatic_backup_retention_days", "7"),
					resource.TestCheckResourceAttr(resourceName, "copy_tags_to_backups", "false"),
					resource.TestMatchResourceAttr(resourceName, "daily_automatic_backup_start_time", regexp.MustCompile(`^\d\d:\d\d$`)),
					resource.TestMatchResourceAttr(resourceName, "dns_name", regexp.MustCompile(`fs-.+\..+`)),
					acctest.MatchResourceAttrRegionalARN(resourceName, "kms_key_id", "kms", regexp.MustCompile(`key/.+`)),
					resource.TestCheckResourceAttr(resourceName, "network_interface_ids.#", "1"),
					acctest.CheckResourceAttrAccountID(resourceName, "owner_id"),
					resource.TestCheckResourceAttr(resourceName, "security_group_ids.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "self_managed_active_directory.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "skip_final_backup", "true"),
					resource.TestCheckResourceAttr(resourceName, "storage_capacity", "32"),
					resource.TestCheckResourceAttr(resourceName, "subnet_ids.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
					resource.TestCheckResourceAttr(resourceName, "throughput_capacity", "8"),
					resource.TestCheckResourceAttrPair(resourceName, "vpc_id", "aws_vpc.test", "id"),
					resource.TestMatchResourceAttr(resourceName, "weekly_maintenance_start_time", regexp.MustCompile(`^\d:\d\d:\d\d$`)),
					resource.TestCheckResourceAttr(resourceName, "deployment_type", "SINGLE_AZ_1"),
					resource.TestCheckResourceAttr(resourceName, "storage_type", "SSD"),
					resource.TestCheckResourceAttr(resourceName, "audit_log_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "audit_log_configuration.0.file_access_audit_log_level", "DISABLED"),
					resource.TestCheckResourceAttr(resourceName, "audit_log_configuration.0.file_share_access_audit_log_level", "DISABLED"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"security_group_ids",
					"skip_final_backup",
				},
			},
			{
				Config:   testAccWindowsFileSystemSubnetIds1WithSingleTypeConfig("SINGLE_AZ_1"),
				PlanOnly: true,
			},
		},
	})
}

func TestAccFSxWindowsFileSystem_singleAz2(t *testing.T) {
	var filesystem fsx.FileSystem
	resourceName := "aws_fsx_windows_file_system.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); acctest.PreCheckPartitionHasService(fsx.EndpointsID, t) },
		ErrorCheck:        acctest.ErrorCheck(t, fsx.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckWindowsFileSystemDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccWindowsFileSystemSubnetIds1WithSingleTypeConfig("SINGLE_AZ_2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckWindowsFileSystemExists(resourceName, &filesystem),
					acctest.MatchResourceAttrRegionalARN(resourceName, "arn", "fsx", regexp.MustCompile(`file-system/fs-.+`)),
					resource.TestCheckResourceAttr(resourceName, "aliases.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "automatic_backup_retention_days", "7"),
					resource.TestCheckResourceAttr(resourceName, "copy_tags_to_backups", "false"),
					resource.TestMatchResourceAttr(resourceName, "daily_automatic_backup_start_time", regexp.MustCompile(`^\d\d:\d\d$`)),
					resource.TestMatchResourceAttr(resourceName, "dns_name", regexp.MustCompile(`^amznfsx\w{8}\.corp\.notexample\.com$`)),
					acctest.MatchResourceAttrRegionalARN(resourceName, "kms_key_id", "kms", regexp.MustCompile(`key/.+`)),
					resource.TestCheckResourceAttr(resourceName, "network_interface_ids.#", "1"),
					acctest.CheckResourceAttrAccountID(resourceName, "owner_id"),
					resource.TestCheckResourceAttr(resourceName, "security_group_ids.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "self_managed_active_directory.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "skip_final_backup", "true"),
					resource.TestCheckResourceAttr(resourceName, "storage_capacity", "32"),
					resource.TestCheckResourceAttr(resourceName, "subnet_ids.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
					resource.TestCheckResourceAttr(resourceName, "throughput_capacity", "8"),
					resource.TestCheckResourceAttrPair(resourceName, "vpc_id", "aws_vpc.test", "id"),
					resource.TestMatchResourceAttr(resourceName, "weekly_maintenance_start_time", regexp.MustCompile(`^\d:\d\d:\d\d$`)),
					resource.TestCheckResourceAttr(resourceName, "deployment_type", "SINGLE_AZ_2"),
					resource.TestCheckResourceAttr(resourceName, "storage_type", "SSD"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"security_group_ids",
					"skip_final_backup",
				},
			},
		},
	})
}

func TestAccFSxWindowsFileSystem_storageTypeHdd(t *testing.T) {
	var filesystem fsx.FileSystem
	resourceName := "aws_fsx_windows_file_system.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); acctest.PreCheckPartitionHasService(fsx.EndpointsID, t) },
		ErrorCheck:        acctest.ErrorCheck(t, fsx.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckWindowsFileSystemDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccWindowsFileSystemSubnetIds1WithStorageTypeConfig("SINGLE_AZ_2", "HDD"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckWindowsFileSystemExists(resourceName, &filesystem),
					resource.TestCheckResourceAttr(resourceName, "deployment_type", "SINGLE_AZ_2"),
					resource.TestCheckResourceAttr(resourceName, "storage_type", "HDD"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"security_group_ids",
					"skip_final_backup",
				},
			},
		},
	})
}

func TestAccFSxWindowsFileSystem_multiAz(t *testing.T) {
	var filesystem fsx.FileSystem
	resourceName := "aws_fsx_windows_file_system.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); acctest.PreCheckPartitionHasService(fsx.EndpointsID, t) },
		ErrorCheck:        acctest.ErrorCheck(t, fsx.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckWindowsFileSystemDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccWindowsFileSystemSubnetIds2Config(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckWindowsFileSystemExists(resourceName, &filesystem),
					acctest.MatchResourceAttrRegionalARN(resourceName, "arn", "fsx", regexp.MustCompile(`file-system/fs-.+`)),
					resource.TestCheckResourceAttr(resourceName, "aliases.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "automatic_backup_retention_days", "7"),
					resource.TestCheckResourceAttr(resourceName, "copy_tags_to_backups", "false"),
					resource.TestMatchResourceAttr(resourceName, "daily_automatic_backup_start_time", regexp.MustCompile(`^\d\d:\d\d$`)),
					resource.TestMatchResourceAttr(resourceName, "dns_name", regexp.MustCompile(`^amznfsx\w{8}\.corp\.notexample\.com$`)),
					acctest.MatchResourceAttrRegionalARN(resourceName, "kms_key_id", "kms", regexp.MustCompile(`key/.+`)),
					resource.TestCheckResourceAttr(resourceName, "network_interface_ids.#", "2"),
					acctest.CheckResourceAttrAccountID(resourceName, "owner_id"),
					resource.TestCheckResourceAttr(resourceName, "self_managed_active_directory.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "skip_final_backup", "true"),
					resource.TestCheckResourceAttr(resourceName, "storage_capacity", "32"),
					resource.TestCheckResourceAttr(resourceName, "subnet_ids.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
					resource.TestCheckResourceAttr(resourceName, "throughput_capacity", "8"),
					resource.TestCheckResourceAttrPair(resourceName, "vpc_id", "aws_vpc.test", "id"),
					resource.TestMatchResourceAttr(resourceName, "weekly_maintenance_start_time", regexp.MustCompile(`^\d:\d\d:\d\d$`)),
					resource.TestCheckResourceAttr(resourceName, "deployment_type", "MULTI_AZ_1"),
					resource.TestCheckResourceAttr(resourceName, "storage_type", "SSD"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"security_group_ids",
					"skip_final_backup",
				},
			},
		},
	})
}

func TestAccFSxWindowsFileSystem_disappears(t *testing.T) {
	var filesystem fsx.FileSystem
	resourceName := "aws_fsx_windows_file_system.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); acctest.PreCheckPartitionHasService(fsx.EndpointsID, t) },
		ErrorCheck:        acctest.ErrorCheck(t, fsx.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckWindowsFileSystemDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccWindowsFileSystemSubnetIds1Config(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckWindowsFileSystemExists(resourceName, &filesystem),
					acctest.CheckResourceDisappears(acctest.Provider, tffsx.ResourceWindowsFileSystem(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccFSxWindowsFileSystem_aliases(t *testing.T) {
	var filesystem1, filesystem2, filesystem3 fsx.FileSystem
	resourceName := "aws_fsx_windows_file_system.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); acctest.PreCheckPartitionHasService(fsx.EndpointsID, t) },
		ErrorCheck:        acctest.ErrorCheck(t, fsx.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckWindowsFileSystemDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccWindowsFileSystemAliases1Config("filesystem1.example.com"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckWindowsFileSystemExists(resourceName, &filesystem1),
					resource.TestCheckResourceAttr(resourceName, "aliases.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "aliases.0", "filesystem1.example.com"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"security_group_ids",
					"skip_final_backup",
				},
			},
			{
				Config: testAccWindowsFileSystemAliases2Config("filesystem2.example.com", "filesystem3.example.com"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckWindowsFileSystemExists(resourceName, &filesystem2),
					testAccCheckWindowsFileSystemNotRecreated(&filesystem1, &filesystem2),
					resource.TestCheckResourceAttr(resourceName, "aliases.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "aliases.0", "filesystem2.example.com"),
					resource.TestCheckResourceAttr(resourceName, "aliases.1", "filesystem3.example.com"),
				),
			},
			{
				Config: testAccWindowsFileSystemAliases1Config("filesystem3.example.com"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckWindowsFileSystemExists(resourceName, &filesystem3),
					testAccCheckWindowsFileSystemNotRecreated(&filesystem2, &filesystem3),
					resource.TestCheckResourceAttr(resourceName, "aliases.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "aliases.0", "filesystem3.example.com"),
				),
			},
		},
	})
}

func TestAccFSxWindowsFileSystem_automaticBackupRetentionDays(t *testing.T) {
	var filesystem1, filesystem2, filesystem3 fsx.FileSystem
	resourceName := "aws_fsx_windows_file_system.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); acctest.PreCheckPartitionHasService(fsx.EndpointsID, t) },
		ErrorCheck:        acctest.ErrorCheck(t, fsx.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckWindowsFileSystemDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccWindowsFileSystemAutomaticBackupRetentionDaysConfig(90),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckWindowsFileSystemExists(resourceName, &filesystem1),
					resource.TestCheckResourceAttr(resourceName, "automatic_backup_retention_days", "90"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"security_group_ids",
					"skip_final_backup",
				},
			},
			{
				Config: testAccWindowsFileSystemAutomaticBackupRetentionDaysConfig(0),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckWindowsFileSystemExists(resourceName, &filesystem2),
					testAccCheckWindowsFileSystemNotRecreated(&filesystem1, &filesystem2),
					resource.TestCheckResourceAttr(resourceName, "automatic_backup_retention_days", "0"),
				),
			},
			{
				Config: testAccWindowsFileSystemAutomaticBackupRetentionDaysConfig(14),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckWindowsFileSystemExists(resourceName, &filesystem3),
					testAccCheckWindowsFileSystemNotRecreated(&filesystem2, &filesystem3),
					resource.TestCheckResourceAttr(resourceName, "automatic_backup_retention_days", "14"),
				),
			},
		},
	})
}

func TestAccFSxWindowsFileSystem_copyTagsToBackups(t *testing.T) {
	var filesystem1, filesystem2 fsx.FileSystem
	resourceName := "aws_fsx_windows_file_system.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); acctest.PreCheckPartitionHasService(fsx.EndpointsID, t) },
		ErrorCheck:        acctest.ErrorCheck(t, fsx.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckWindowsFileSystemDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccWindowsFileSystemCopyTagsToBackupsConfig(true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckWindowsFileSystemExists(resourceName, &filesystem1),
					resource.TestCheckResourceAttr(resourceName, "copy_tags_to_backups", "true"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"security_group_ids",
					"skip_final_backup",
				},
			},
			{
				Config: testAccWindowsFileSystemCopyTagsToBackupsConfig(false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckWindowsFileSystemExists(resourceName, &filesystem2),
					testAccCheckWindowsFileSystemRecreated(&filesystem1, &filesystem2),
					resource.TestCheckResourceAttr(resourceName, "copy_tags_to_backups", "false"),
				),
			},
		},
	})
}

func TestAccFSxWindowsFileSystem_dailyAutomaticBackupStartTime(t *testing.T) {
	var filesystem1, filesystem2 fsx.FileSystem
	resourceName := "aws_fsx_windows_file_system.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); acctest.PreCheckPartitionHasService(fsx.EndpointsID, t) },
		ErrorCheck:        acctest.ErrorCheck(t, fsx.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckWindowsFileSystemDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccWindowsFileSystemDailyAutomaticBackupStartTimeConfig("01:01"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckWindowsFileSystemExists(resourceName, &filesystem1),
					resource.TestCheckResourceAttr(resourceName, "daily_automatic_backup_start_time", "01:01"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"security_group_ids",
					"skip_final_backup",
				},
			},
			{
				Config: testAccWindowsFileSystemDailyAutomaticBackupStartTimeConfig("02:02"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckWindowsFileSystemExists(resourceName, &filesystem2),
					testAccCheckWindowsFileSystemNotRecreated(&filesystem1, &filesystem2),
					resource.TestCheckResourceAttr(resourceName, "daily_automatic_backup_start_time", "02:02"),
				),
			},
		},
	})
}

func TestAccFSxWindowsFileSystem_kmsKeyID(t *testing.T) {
	var filesystem1, filesystem2 fsx.FileSystem
	kmsKeyResourceName1 := "aws_kms_key.test1"
	kmsKeyResourceName2 := "aws_kms_key.test2"
	resourceName := "aws_fsx_windows_file_system.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); acctest.PreCheckPartitionHasService(fsx.EndpointsID, t) },
		ErrorCheck:        acctest.ErrorCheck(t, fsx.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckWindowsFileSystemDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccWindowsFileSystemKMSKeyId1Config(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckWindowsFileSystemExists(resourceName, &filesystem1),
					resource.TestCheckResourceAttrPair(resourceName, "kms_key_id", kmsKeyResourceName1, "arn"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"security_group_ids",
					"skip_final_backup",
				},
			},
			{
				Config: testAccWindowsFileSystemKMSKeyId2Config(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckWindowsFileSystemExists(resourceName, &filesystem2),
					testAccCheckWindowsFileSystemRecreated(&filesystem1, &filesystem2),
					resource.TestCheckResourceAttrPair(resourceName, "kms_key_id", kmsKeyResourceName2, "arn"),
				),
			},
		},
	})
}

func TestAccFSxWindowsFileSystem_securityGroupIDs(t *testing.T) {
	var filesystem1, filesystem2 fsx.FileSystem
	resourceName := "aws_fsx_windows_file_system.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); acctest.PreCheckPartitionHasService(fsx.EndpointsID, t) },
		ErrorCheck:        acctest.ErrorCheck(t, fsx.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckWindowsFileSystemDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccWindowsFileSystemSecurityGroupIds1Config(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckWindowsFileSystemExists(resourceName, &filesystem1),
					resource.TestCheckResourceAttr(resourceName, "security_group_ids.#", "1"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"security_group_ids",
					"skip_final_backup",
				},
			},
			{
				Config: testAccWindowsFileSystemSecurityGroupIds2Config(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckWindowsFileSystemExists(resourceName, &filesystem2),
					testAccCheckWindowsFileSystemRecreated(&filesystem1, &filesystem2),
					resource.TestCheckResourceAttr(resourceName, "security_group_ids.#", "2"),
				),
			},
		},
	})
}

func TestAccFSxWindowsFileSystem_selfManagedActiveDirectory(t *testing.T) {
	var filesystem fsx.FileSystem
	resourceName := "aws_fsx_windows_file_system.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); acctest.PreCheckPartitionHasService(fsx.EndpointsID, t) },
		ErrorCheck:        acctest.ErrorCheck(t, fsx.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckWindowsFileSystemDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccWindowsFileSystemSelfManagedActiveDirectoryConfig(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckWindowsFileSystemExists(resourceName, &filesystem),
					resource.TestCheckResourceAttr(resourceName, "self_managed_active_directory.#", "1"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"security_group_ids",
					"self_managed_active_directory",
					"skip_final_backup",
				},
			},
		},
	})
}

func TestAccFSxWindowsFileSystem_SelfManagedActiveDirectory_username(t *testing.T) {
	var filesystem fsx.FileSystem
	resourceName := "aws_fsx_windows_file_system.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); acctest.PreCheckPartitionHasService(fsx.EndpointsID, t) },
		ErrorCheck:        acctest.ErrorCheck(t, fsx.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckWindowsFileSystemDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccWindowsFileSystemSelfManagedActiveDirectoryUsernameConfig("Admin"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckWindowsFileSystemExists(resourceName, &filesystem),
					resource.TestCheckResourceAttr(resourceName, "self_managed_active_directory.#", "1"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"security_group_ids",
					"self_managed_active_directory",
					"skip_final_backup",
				},
			},
			{
				Config: testAccWindowsFileSystemSelfManagedActiveDirectoryUsernameConfig("Administrator"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckWindowsFileSystemExists(resourceName, &filesystem),
					resource.TestCheckResourceAttr(resourceName, "self_managed_active_directory.#", "1"),
				),
			},
		},
	})
}

func TestAccFSxWindowsFileSystem_storageCapacity(t *testing.T) {
	var filesystem1, filesystem2 fsx.FileSystem
	resourceName := "aws_fsx_windows_file_system.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); acctest.PreCheckPartitionHasService(fsx.EndpointsID, t) },
		ErrorCheck:        acctest.ErrorCheck(t, fsx.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckWindowsFileSystemDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccWindowsFileSystemStorageCapacityConfig(32),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckWindowsFileSystemExists(resourceName, &filesystem1),
					resource.TestCheckResourceAttr(resourceName, "storage_capacity", "32"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"security_group_ids",
					"skip_final_backup",
				},
			},
			{
				Config: testAccWindowsFileSystemStorageCapacityConfig(36),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckWindowsFileSystemExists(resourceName, &filesystem2),
					testAccCheckWindowsFileSystemNotRecreated(&filesystem1, &filesystem2),
					resource.TestCheckResourceAttr(resourceName, "storage_capacity", "36"),
				),
			},
		},
	})
}

func TestAccFSxWindowsFileSystem_fromBackup(t *testing.T) {
	var filesystem fsx.FileSystem
	resourceName := "aws_fsx_windows_file_system.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); acctest.PreCheckPartitionHasService(fsx.EndpointsID, t) },
		ErrorCheck:        acctest.ErrorCheck(t, fsx.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckWindowsFileSystemDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccWindowsFileSystemFromBackup(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckWindowsFileSystemExists(resourceName, &filesystem),
					resource.TestCheckResourceAttrPair(resourceName, "backup_id", "aws_fsx_backup.test", "id"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"security_group_ids",
					"skip_final_backup",
					"backup_id",
				},
			},
		},
	})
}

func TestAccFSxWindowsFileSystem_tags(t *testing.T) {
	var filesystem1, filesystem2, filesystem3 fsx.FileSystem
	resourceName := "aws_fsx_windows_file_system.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); acctest.PreCheckPartitionHasService(fsx.EndpointsID, t) },
		ErrorCheck:        acctest.ErrorCheck(t, fsx.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckWindowsFileSystemDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccWindowsFileSystemTags1Config("key1", "value1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckWindowsFileSystemExists(resourceName, &filesystem1),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"security_group_ids",
					"skip_final_backup",
				},
			},
			{
				Config: testAccWindowsFileSystemTags2Config("key1", "value1updated", "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckWindowsFileSystemExists(resourceName, &filesystem2),
					testAccCheckWindowsFileSystemNotRecreated(&filesystem1, &filesystem2),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1updated"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
			{
				Config: testAccWindowsFileSystemTags1Config("key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckWindowsFileSystemExists(resourceName, &filesystem3),
					testAccCheckWindowsFileSystemNotRecreated(&filesystem2, &filesystem3),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
		},
	})
}

func TestAccFSxWindowsFileSystem_throughputCapacity(t *testing.T) {
	var filesystem1, filesystem2 fsx.FileSystem
	resourceName := "aws_fsx_windows_file_system.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); acctest.PreCheckPartitionHasService(fsx.EndpointsID, t) },
		ErrorCheck:        acctest.ErrorCheck(t, fsx.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckWindowsFileSystemDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccWindowsFileSystemThroughputCapacityConfig(16),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckWindowsFileSystemExists(resourceName, &filesystem1),
					resource.TestCheckResourceAttr(resourceName, "throughput_capacity", "16"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"security_group_ids",
					"skip_final_backup",
				},
			},
			{
				Config: testAccWindowsFileSystemThroughputCapacityConfig(32),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckWindowsFileSystemExists(resourceName, &filesystem2),
					testAccCheckWindowsFileSystemNotRecreated(&filesystem1, &filesystem2),
					resource.TestCheckResourceAttr(resourceName, "throughput_capacity", "32"),
				),
			},
		},
	})
}

func TestAccFSxWindowsFileSystem_weeklyMaintenanceStartTime(t *testing.T) {
	var filesystem1, filesystem2 fsx.FileSystem
	resourceName := "aws_fsx_windows_file_system.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); acctest.PreCheckPartitionHasService(fsx.EndpointsID, t) },
		ErrorCheck:        acctest.ErrorCheck(t, fsx.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckWindowsFileSystemDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccWindowsFileSystemWeeklyMaintenanceStartTimeConfig("1:01:01"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckWindowsFileSystemExists(resourceName, &filesystem1),
					resource.TestCheckResourceAttr(resourceName, "weekly_maintenance_start_time", "1:01:01"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"security_group_ids",
					"skip_final_backup",
				},
			},
			{
				Config: testAccWindowsFileSystemWeeklyMaintenanceStartTimeConfig("2:02:02"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckWindowsFileSystemExists(resourceName, &filesystem2),
					testAccCheckWindowsFileSystemNotRecreated(&filesystem1, &filesystem2),
					resource.TestCheckResourceAttr(resourceName, "weekly_maintenance_start_time", "2:02:02"),
				),
			},
		},
	})
}

func TestAccFSxWindowsFileSystem_audit(t *testing.T) {
	var filesystem fsx.FileSystem
	resourceName := "aws_fsx_windows_file_system.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); acctest.PreCheckPartitionHasService(fsx.EndpointsID, t) },
		ErrorCheck:        acctest.ErrorCheck(t, fsx.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckWindowsFileSystemDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccWindowsFileSystemAuditConfig(rName, "SUCCESS_ONLY"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckWindowsFileSystemExists(resourceName, &filesystem),
					resource.TestCheckResourceAttr(resourceName, "audit_log_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "audit_log_configuration.0.file_access_audit_log_level", "SUCCESS_ONLY"),
					resource.TestCheckResourceAttr(resourceName, "audit_log_configuration.0.file_share_access_audit_log_level", "SUCCESS_ONLY"),
					resource.TestCheckResourceAttrSet(resourceName, "audit_log_configuration.0.audit_log_destination"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"security_group_ids",
					"skip_final_backup",
				},
			},
			{
				Config: testAccWindowsFileSystemAuditConfig(rName, "SUCCESS_AND_FAILURE"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckWindowsFileSystemExists(resourceName, &filesystem),
					resource.TestCheckResourceAttr(resourceName, "audit_log_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "audit_log_configuration.0.file_access_audit_log_level", "SUCCESS_AND_FAILURE"),
					resource.TestCheckResourceAttr(resourceName, "audit_log_configuration.0.file_share_access_audit_log_level", "SUCCESS_AND_FAILURE"),
					resource.TestCheckResourceAttrSet(resourceName, "audit_log_configuration.0.audit_log_destination"),
				),
			},
		},
	})
}

func testAccCheckWindowsFileSystemExists(resourceName string, fs *fsx.FileSystem) resource.TestCheckFunc {
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
			return fmt.Errorf("FSx Windows File System (%s) not found", rs.Primary.ID)
		}

		*fs = *filesystem

		return nil
	}
}

func testAccCheckWindowsFileSystemDestroy(s *terraform.State) error {
	conn := acctest.Provider.Meta().(*conns.AWSClient).FSxConn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_fsx_windows_file_system" {
			continue
		}

		filesystem, err := tffsx.FindFileSystemByID(conn, rs.Primary.ID)
		if tfresource.NotFound(err) {
			continue
		}

		if filesystem != nil {
			return fmt.Errorf("FSx Windows File System (%s) still exists", rs.Primary.ID)
		}
	}
	return nil
}

func testAccCheckWindowsFileSystemNotRecreated(i, j *fsx.FileSystem) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if aws.StringValue(i.FileSystemId) != aws.StringValue(j.FileSystemId) {
			return fmt.Errorf("FSx File System (%s) recreated", aws.StringValue(i.FileSystemId))
		}

		return nil
	}
}

func testAccCheckWindowsFileSystemRecreated(i, j *fsx.FileSystem) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if aws.StringValue(i.FileSystemId) == aws.StringValue(j.FileSystemId) {
			return fmt.Errorf("FSx File System (%s) not recreated", aws.StringValue(i.FileSystemId))
		}

		return nil
	}
}

func testAccWindowsFileSystemBaseConfig() string {
	return `
data "aws_availability_zones" "available" {
  state = "available"

  filter {
    name   = "opt-in-status"
    values = ["opt-in-not-required"]
  }
}

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
`
}

func testAccWindowsFileSystemAliases1Config(alias1 string) string {
	return testAccWindowsFileSystemBaseConfig() + fmt.Sprintf(`
resource "aws_fsx_windows_file_system" "test" {
  active_directory_id = aws_directory_service_directory.test.id
  skip_final_backup   = true
  storage_capacity    = 32
  subnet_ids          = [aws_subnet.test1.id]
  throughput_capacity = 8

  aliases = [%[1]q]
}
`, alias1)
}

func testAccWindowsFileSystemAliases2Config(alias1, alias2 string) string {
	return testAccWindowsFileSystemBaseConfig() + fmt.Sprintf(`
resource "aws_fsx_windows_file_system" "test" {
  active_directory_id = aws_directory_service_directory.test.id
  skip_final_backup   = true
  storage_capacity    = 32
  subnet_ids          = [aws_subnet.test1.id]
  throughput_capacity = 8

  aliases = [%[1]q, %[2]q]
}
`, alias1, alias2)
}

func testAccWindowsFileSystemAutomaticBackupRetentionDaysConfig(automaticBackupRetentionDays int) string {
	return testAccWindowsFileSystemBaseConfig() + fmt.Sprintf(`
resource "aws_fsx_windows_file_system" "test" {
  active_directory_id             = aws_directory_service_directory.test.id
  automatic_backup_retention_days = %[1]d
  skip_final_backup               = true
  storage_capacity                = 32
  subnet_ids                      = [aws_subnet.test1.id]
  throughput_capacity             = 8
}
`, automaticBackupRetentionDays)
}

func testAccWindowsFileSystemCopyTagsToBackupsConfig(copyTagsToBackups bool) string {
	return testAccWindowsFileSystemBaseConfig() + fmt.Sprintf(`
resource "aws_fsx_windows_file_system" "test" {
  active_directory_id  = aws_directory_service_directory.test.id
  copy_tags_to_backups = %[1]t
  skip_final_backup    = true
  storage_capacity     = 32
  subnet_ids           = [aws_subnet.test1.id]
  throughput_capacity  = 8
}
`, copyTagsToBackups)
}

func testAccWindowsFileSystemDailyAutomaticBackupStartTimeConfig(dailyAutomaticBackupStartTime string) string {
	return testAccWindowsFileSystemBaseConfig() + fmt.Sprintf(`
resource "aws_fsx_windows_file_system" "test" {
  active_directory_id               = aws_directory_service_directory.test.id
  daily_automatic_backup_start_time = %[1]q
  skip_final_backup                 = true
  storage_capacity                  = 32
  subnet_ids                        = [aws_subnet.test1.id]
  throughput_capacity               = 8
}
`, dailyAutomaticBackupStartTime)
}

func testAccWindowsFileSystemKMSKeyId1Config() string {
	return testAccWindowsFileSystemBaseConfig() + `
resource "aws_kms_key" "test1" {
  description             = "FSx KMS Testing key"
  deletion_window_in_days = 7
}

resource "aws_fsx_windows_file_system" "test" {
  active_directory_id = aws_directory_service_directory.test.id
  kms_key_id          = aws_kms_key.test1.arn
  skip_final_backup   = true
  storage_capacity    = 32
  subnet_ids          = [aws_subnet.test1.id]
  throughput_capacity = 8
}
`
}

func testAccWindowsFileSystemKMSKeyId2Config() string {
	return testAccWindowsFileSystemBaseConfig() + `
resource "aws_kms_key" "test2" {
  description             = "FSx KMS Testing key"
  deletion_window_in_days = 7
}

resource "aws_fsx_windows_file_system" "test" {
  active_directory_id = aws_directory_service_directory.test.id
  kms_key_id          = aws_kms_key.test2.arn
  skip_final_backup   = true
  storage_capacity    = 32
  subnet_ids          = [aws_subnet.test1.id]
  throughput_capacity = 8
}
`
}

func testAccWindowsFileSystemSecurityGroupIds1Config() string {
	return testAccWindowsFileSystemBaseConfig() + `
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

resource "aws_fsx_windows_file_system" "test" {
  active_directory_id = aws_directory_service_directory.test.id
  security_group_ids  = [aws_security_group.test1.id]
  skip_final_backup   = true
  storage_capacity    = 32
  subnet_ids          = [aws_subnet.test1.id]
  throughput_capacity = 8
}
`
}

func testAccWindowsFileSystemSecurityGroupIds2Config() string {
	return testAccWindowsFileSystemBaseConfig() + `
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

resource "aws_fsx_windows_file_system" "test" {
  active_directory_id = aws_directory_service_directory.test.id
  security_group_ids  = [aws_security_group.test1.id, aws_security_group.test2.id]
  skip_final_backup   = true
  storage_capacity    = 32
  subnet_ids          = [aws_subnet.test1.id]
  throughput_capacity = 8
}
`
}

func testAccWindowsFileSystemSelfManagedActiveDirectoryConfig() string {
	return testAccWindowsFileSystemBaseConfig() + `
resource "aws_fsx_windows_file_system" "test" {
  skip_final_backup   = true
  storage_capacity    = 32
  subnet_ids          = [aws_subnet.test1.id]
  throughput_capacity = 8

  self_managed_active_directory {
    dns_ips     = aws_directory_service_directory.test.dns_ip_addresses
    domain_name = aws_directory_service_directory.test.name
    password    = aws_directory_service_directory.test.password
    username    = "Admin"
  }
}
`
}

func testAccWindowsFileSystemSelfManagedActiveDirectoryUsernameConfig(username string) string {
	return testAccWindowsFileSystemBaseConfig() + fmt.Sprintf(`
resource "aws_fsx_windows_file_system" "test" {
  skip_final_backup   = true
  storage_capacity    = 32
  subnet_ids          = [aws_subnet.test1.id]
  throughput_capacity = 8

  self_managed_active_directory {
    dns_ips     = aws_directory_service_directory.test.dns_ip_addresses
    domain_name = aws_directory_service_directory.test.name
    password    = aws_directory_service_directory.test.password
    username    = %[1]q
  }
}
`, username)
}

func testAccWindowsFileSystemStorageCapacityConfig(storageCapacity int) string {
	return testAccWindowsFileSystemBaseConfig() + fmt.Sprintf(`
resource "aws_fsx_windows_file_system" "test" {
  active_directory_id = aws_directory_service_directory.test.id
  skip_final_backup   = true
  storage_capacity    = %[1]d
  subnet_ids          = [aws_subnet.test1.id]
  throughput_capacity = 16
}
`, storageCapacity)
}

func testAccWindowsFileSystemSubnetIds1Config() string {
	return testAccWindowsFileSystemBaseConfig() + `
resource "aws_fsx_windows_file_system" "test" {
  active_directory_id = aws_directory_service_directory.test.id
  skip_final_backup   = true
  storage_capacity    = 32
  subnet_ids          = [aws_subnet.test1.id]
  throughput_capacity = 8
}
`
}

func testAccWindowsFileSystemSubnetIds1WithSingleTypeConfig(azType string) string {
	return testAccWindowsFileSystemBaseConfig() + fmt.Sprintf(`
resource "aws_fsx_windows_file_system" "test" {
  active_directory_id = aws_directory_service_directory.test.id
  skip_final_backup   = true
  storage_capacity    = 32
  deployment_type     = %[1]q
  subnet_ids          = [aws_subnet.test1.id]
  throughput_capacity = 8
}
`, azType)
}

func testAccWindowsFileSystemSubnetIds1WithStorageTypeConfig(azType, storageType string) string {
	return testAccWindowsFileSystemBaseConfig() + fmt.Sprintf(`
resource "aws_fsx_windows_file_system" "test" {
  active_directory_id = aws_directory_service_directory.test.id
  skip_final_backup   = true
  storage_capacity    = 2000
  deployment_type     = %[1]q
  subnet_ids          = [aws_subnet.test1.id]
  throughput_capacity = 8
  storage_type        = %[2]q
}
`, azType, storageType)
}

func testAccWindowsFileSystemSubnetIds2Config() string {
	return acctest.ConfigCompose(testAccWindowsFileSystemBaseConfig(), `
resource "aws_fsx_windows_file_system" "test" {
  active_directory_id = aws_directory_service_directory.test.id
  skip_final_backup   = true
  storage_capacity    = 32
  deployment_type     = "MULTI_AZ_1"
  subnet_ids          = [aws_subnet.test1.id, aws_subnet.test2.id]
  preferred_subnet_id = aws_subnet.test1.id
  throughput_capacity = 8
}
`)
}

func testAccWindowsFileSystemFromBackup() string {
	return testAccWindowsFileSystemBaseConfig() + `
resource "aws_fsx_windows_file_system" "base" {
  active_directory_id = aws_directory_service_directory.test.id
  skip_final_backup   = true
  storage_capacity    = 32
  subnet_ids          = [aws_subnet.test1.id]
  throughput_capacity = 8
}

resource "aws_fsx_backup" "test" {
  file_system_id = aws_fsx_windows_file_system.base.id
}

resource "aws_fsx_windows_file_system" "test" {
  active_directory_id = aws_directory_service_directory.test.id
  backup_id           = aws_fsx_backup.test.id
  skip_final_backup   = true
  subnet_ids          = [aws_subnet.test1.id]
  throughput_capacity = 8
}
`
}

func testAccWindowsFileSystemTags1Config(tagKey1, tagValue1 string) string {
	return testAccWindowsFileSystemBaseConfig() + fmt.Sprintf(`
resource "aws_fsx_windows_file_system" "test" {
  active_directory_id = aws_directory_service_directory.test.id
  skip_final_backup   = true
  storage_capacity    = 32
  subnet_ids          = [aws_subnet.test1.id]
  throughput_capacity = 8

  tags = {
    %[1]q = %[2]q
  }
}
`, tagKey1, tagValue1)
}

func testAccWindowsFileSystemTags2Config(tagKey1, tagValue1, tagKey2, tagValue2 string) string {
	return testAccWindowsFileSystemBaseConfig() + fmt.Sprintf(`
resource "aws_fsx_windows_file_system" "test" {
  active_directory_id = aws_directory_service_directory.test.id
  skip_final_backup   = true
  storage_capacity    = 32
  subnet_ids          = [aws_subnet.test1.id]
  throughput_capacity = 8

  tags = {
    %[1]q = %[2]q
    %[3]q = %[4]q
  }
}
`, tagKey1, tagValue1, tagKey2, tagValue2)
}

func testAccWindowsFileSystemThroughputCapacityConfig(throughputCapacity int) string {
	return testAccWindowsFileSystemBaseConfig() + fmt.Sprintf(`
resource "aws_fsx_windows_file_system" "test" {
  active_directory_id = aws_directory_service_directory.test.id
  skip_final_backup   = true
  storage_capacity    = 32
  subnet_ids          = [aws_subnet.test1.id]
  throughput_capacity = %[1]d
}
`, throughputCapacity)
}

func testAccWindowsFileSystemWeeklyMaintenanceStartTimeConfig(weeklyMaintenanceStartTime string) string {
	return testAccWindowsFileSystemBaseConfig() + fmt.Sprintf(`
resource "aws_fsx_windows_file_system" "test" {
  active_directory_id           = aws_directory_service_directory.test.id
  skip_final_backup             = true
  storage_capacity              = 32
  subnet_ids                    = [aws_subnet.test1.id]
  throughput_capacity           = 8
  weekly_maintenance_start_time = %[1]q
}
`, weeklyMaintenanceStartTime)
}

func testAccWindowsFileSystemAuditConfig(rName, status string) string {
	return testAccWindowsFileSystemBaseConfig() + fmt.Sprintf(`
resource aws_cloudwatch_log_group "test" {
  name = "/aws/fsx/%[1]s"
}

resource "aws_fsx_windows_file_system" "test" {
  active_directory_id = aws_directory_service_directory.test.id
  skip_final_backup   = true
  storage_capacity    = 32
  subnet_ids          = [aws_subnet.test1.id]
  throughput_capacity = 32

  audit_log_configuration {
    audit_log_destination             = aws_cloudwatch_log_group.test.arn
    file_access_audit_log_level       = %[2]q
    file_share_access_audit_log_level = %[2]q
  }
}
`, rName, status)
}
