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

func init() {
	acctest.RegisterServiceErrorCheckFunc(fsx.EndpointsID, testAccErrorCheckSkip)
}

func testAccErrorCheckSkip(t *testing.T) resource.ErrorCheckFunc {
	return acctest.ErrorCheckSkipMessagesContaining(t,
		"Amazon FSx does not currently support OpenZFS file system creation in the following Availability Zones",
	)
}

func TestAccFSxOpenzfsFileSystem_basic(t *testing.T) {
	var filesystem fsx.FileSystem
	resourceName := "aws_fsx_openzfs_file_system.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); acctest.PreCheckPartitionHasService(fsx.EndpointsID, t) },
		ErrorCheck:        acctest.ErrorCheck(t, fsx.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckOpenzfsFileSystemDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccOpenZFSFileSystemConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckOpenzfsFileSystemExists(resourceName, &filesystem),
					acctest.MatchResourceAttrRegionalARN(resourceName, "arn", "fsx", regexp.MustCompile(`file-system/fs-.+`)),
					resource.TestCheckResourceAttr(resourceName, "network_interface_ids.#", "1"),
					acctest.CheckResourceAttrAccountID(resourceName, "owner_id"),
					resource.TestCheckResourceAttrSet(resourceName, "root_volume_id"),
					resource.TestCheckResourceAttrSet(resourceName, "dns_name"),
					resource.TestCheckResourceAttr(resourceName, "storage_capacity", "64"),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "subnet_ids.*", "aws_subnet.test1", "id"),
					resource.TestCheckResourceAttr(resourceName, "security_group_ids.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
					resource.TestCheckResourceAttrPair(resourceName, "vpc_id", "aws_vpc.test", "id"),
					resource.TestMatchResourceAttr(resourceName, "weekly_maintenance_start_time", regexp.MustCompile(`^\d:\d\d:\d\d$`)),
					resource.TestCheckResourceAttr(resourceName, "deployment_type", fsx.OpenZFSDeploymentTypeSingleAz1),
					resource.TestCheckResourceAttr(resourceName, "automatic_backup_retention_days", "0"),
					resource.TestCheckResourceAttr(resourceName, "storage_type", fsx.StorageTypeSsd),
					resource.TestCheckResourceAttrSet(resourceName, "kms_key_id"),
					resource.TestCheckResourceAttr(resourceName, "throughput_capacity", "64"),
					resource.TestCheckResourceAttr(resourceName, "disk_iops_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "disk_iops_configuration.0.mode", "AUTOMATIC"),
					resource.TestCheckResourceAttr(resourceName, "disk_iops_configuration.0.iops", "192"),
					resource.TestCheckResourceAttr(resourceName, "root_volume_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "root_volume_configuration.0.data_compression_type", "NONE"),
					resource.TestCheckResourceAttr(resourceName, "root_volume_configuration.0.nfs_exports.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "root_volume_configuration.0.nfs_exports.0.client_configurations.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "root_volume_configuration.0.nfs_exports.0.client_configurations.0.clients", "*"),
					resource.TestCheckResourceAttr(resourceName, "root_volume_configuration.0.nfs_exports.0.client_configurations.0.options.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "root_volume_configuration.0.nfs_exports.0.client_configurations.0.options.0", "crossmnt"),
					resource.TestCheckResourceAttr(resourceName, "root_volume_configuration.0.read_only", "false"),
					resource.TestCheckResourceAttr(resourceName, "root_volume_configuration.0.user_and_group_quotas.#", "0"),
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

func TestAccFSxOpenzfsFileSystem_diskIops(t *testing.T) {
	var filesystem fsx.FileSystem
	resourceName := "aws_fsx_openzfs_file_system.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); acctest.PreCheckPartitionHasService(fsx.EndpointsID, t) },
		ErrorCheck:        acctest.ErrorCheck(t, fsx.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckOpenzfsFileSystemDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccOpenZFSFileSystemConfig_diskIOPSConfiguration(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckOpenzfsFileSystemExists(resourceName, &filesystem),
					resource.TestCheckResourceAttr(resourceName, "disk_iops_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "disk_iops_configuration.0.mode", "USER_PROVISIONED"),
					resource.TestCheckResourceAttr(resourceName, "disk_iops_configuration.0.iops", "3072"),
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

func TestAccFSxOpenzfsFileSystem_disappears(t *testing.T) {
	var filesystem fsx.FileSystem
	resourceName := "aws_fsx_openzfs_file_system.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); acctest.PreCheckPartitionHasService(fsx.EndpointsID, t) },
		ErrorCheck:        acctest.ErrorCheck(t, fsx.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckOpenzfsFileSystemDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccOpenZFSFileSystemConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckOpenzfsFileSystemExists(resourceName, &filesystem),
					acctest.CheckResourceDisappears(acctest.Provider, tffsx.ResourceOpenzfsFileSystem(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccFSxOpenzfsFileSystem_rootVolume(t *testing.T) {
	var filesystem1, filesystem2, filesystem3 fsx.FileSystem
	resourceName := "aws_fsx_openzfs_file_system.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); acctest.PreCheckPartitionHasService(fsx.EndpointsID, t) },
		ErrorCheck:        acctest.ErrorCheck(t, fsx.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckOpenzfsFileSystemDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccOpenZFSFileSystemConfig_rootVolume1(rName, "NONE", "false", 128),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckOpenzfsFileSystemExists(resourceName, &filesystem1),
					resource.TestCheckResourceAttr(resourceName, "root_volume_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "root_volume_configuration.0.data_compression_type", "NONE"),
					resource.TestCheckResourceAttr(resourceName, "root_volume_configuration.0.nfs_exports.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "root_volume_configuration.0.nfs_exports.0.client_configurations.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "root_volume_configuration.0.nfs_exports.0.client_configurations.0.clients", "10.0.1.0/24"),
					resource.TestCheckResourceAttr(resourceName, "root_volume_configuration.0.nfs_exports.0.client_configurations.0.options.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "root_volume_configuration.0.nfs_exports.0.client_configurations.0.options.0", "sync"),
					resource.TestCheckResourceAttr(resourceName, "root_volume_configuration.0.nfs_exports.0.client_configurations.0.options.1", "rw"),
					resource.TestCheckResourceAttr(resourceName, "root_volume_configuration.0.read_only", "false"),
					resource.TestCheckResourceAttr(resourceName, "root_volume_configuration.0.user_and_group_quotas.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "root_volume_configuration.0.user_and_group_quotas.*", map[string]string{
						"id":                         "10",
						"storage_capacity_quota_gib": "128",
						"type":                       "USER",
					}),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"security_group_ids"},
			},
			{
				Config: testAccOpenZFSFileSystemConfig_rootVolume2(rName, "ZSTD", "true", 256),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckOpenzfsFileSystemExists(resourceName, &filesystem2),
					testAccCheckOpenzfsFileSystemNotRecreated(&filesystem1, &filesystem2),
					resource.TestCheckResourceAttr(resourceName, "root_volume_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "root_volume_configuration.0.data_compression_type", "ZSTD"),
					resource.TestCheckResourceAttr(resourceName, "root_volume_configuration.0.nfs_exports.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "root_volume_configuration.0.nfs_exports.0.client_configurations.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "root_volume_configuration.0.nfs_exports.0.client_configurations.0.clients", "10.0.1.0/24"),
					resource.TestCheckResourceAttr(resourceName, "root_volume_configuration.0.nfs_exports.0.client_configurations.0.options.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "root_volume_configuration.0.nfs_exports.0.client_configurations.0.options.0", "async"),
					resource.TestCheckResourceAttr(resourceName, "root_volume_configuration.0.nfs_exports.0.client_configurations.0.options.1", "rw"),
					resource.TestCheckResourceAttr(resourceName, "root_volume_configuration.0.read_only", "true"),
					resource.TestCheckResourceAttr(resourceName, "root_volume_configuration.0.user_and_group_quotas.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "root_volume_configuration.0.user_and_group_quotas.*", map[string]string{
						"id":                         "10",
						"storage_capacity_quota_gib": "256",
						"type":                       "USER",
					}),
				),
			},
			{
				Config: testAccOpenZFSFileSystemConfig_rootVolume3Client(rName, "NONE", "false", 128, 1024),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckOpenzfsFileSystemExists(resourceName, &filesystem3),
					testAccCheckOpenzfsFileSystemNotRecreated(&filesystem1, &filesystem3),
					resource.TestCheckResourceAttr(resourceName, "root_volume_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "root_volume_configuration.0.data_compression_type", "NONE"),
					resource.TestCheckResourceAttr(resourceName, "root_volume_configuration.0.nfs_exports.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "root_volume_configuration.0.nfs_exports.0.client_configurations.#", "2"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "root_volume_configuration.0.nfs_exports.0.client_configurations.*", map[string]string{
						"clients":   "10.0.1.0/24",
						"options.0": "async",
						"options.1": "rw",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "root_volume_configuration.0.nfs_exports.0.client_configurations.*", map[string]string{
						"clients":   "*",
						"options.0": "sync",
						"options.1": "rw",
					}),
					resource.TestCheckResourceAttr(resourceName, "root_volume_configuration.0.read_only", "false"),
					resource.TestCheckResourceAttr(resourceName, "root_volume_configuration.0.user_and_group_quotas.#", "4"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "root_volume_configuration.0.user_and_group_quotas.*", map[string]string{
						"id":                         "10",
						"storage_capacity_quota_gib": "128",
						"type":                       "USER",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "root_volume_configuration.0.user_and_group_quotas.*", map[string]string{
						"id":                         "20",
						"storage_capacity_quota_gib": "1024",
						"type":                       "GROUP",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "root_volume_configuration.0.user_and_group_quotas.*", map[string]string{
						"id":                         "5",
						"storage_capacity_quota_gib": "1024",
						"type":                       "GROUP",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "root_volume_configuration.0.user_and_group_quotas.*", map[string]string{
						"id":                         "100",
						"storage_capacity_quota_gib": "128",
						"type":                       "USER",
					}),
				),
			},
			{
				Config: testAccOpenZFSFileSystemConfig_rootVolume4(rName, "NONE", "false", 128, 1024),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckOpenzfsFileSystemExists(resourceName, &filesystem1),
					resource.TestCheckResourceAttr(resourceName, "root_volume_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "root_volume_configuration.0.data_compression_type", "NONE"),
					resource.TestCheckResourceAttr(resourceName, "root_volume_configuration.0.nfs_exports.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "root_volume_configuration.0.read_only", "false"),
					resource.TestCheckResourceAttr(resourceName, "root_volume_configuration.0.user_and_group_quotas.#", "4"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "root_volume_configuration.0.user_and_group_quotas.*", map[string]string{
						"id":                         "10",
						"storage_capacity_quota_gib": "128",
						"type":                       "USER",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "root_volume_configuration.0.user_and_group_quotas.*", map[string]string{
						"id":                         "20",
						"storage_capacity_quota_gib": "1024",
						"type":                       "GROUP",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "root_volume_configuration.0.user_and_group_quotas.*", map[string]string{
						"id":                         "5",
						"storage_capacity_quota_gib": "1024",
						"type":                       "GROUP",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "root_volume_configuration.0.user_and_group_quotas.*", map[string]string{
						"id":                         "100",
						"storage_capacity_quota_gib": "128",
						"type":                       "USER",
					}),
				),
			},
		},
	})
}

func TestAccFSxOpenzfsFileSystem_securityGroupIDs(t *testing.T) {
	var filesystem1, filesystem2 fsx.FileSystem
	resourceName := "aws_fsx_openzfs_file_system.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); acctest.PreCheckPartitionHasService(fsx.EndpointsID, t) },
		ErrorCheck:        acctest.ErrorCheck(t, fsx.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckOpenzfsFileSystemDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccOpenZFSFileSystemConfig_securityGroupIDs1(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckOpenzfsFileSystemExists(resourceName, &filesystem1),
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
				Config: testAccOpenZFSFileSystemConfig_securityGroupIDs2(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckOpenzfsFileSystemExists(resourceName, &filesystem2),
					testAccCheckOpenzfsFileSystemRecreated(&filesystem1, &filesystem2),
					resource.TestCheckResourceAttr(resourceName, "security_group_ids.#", "2"),
				),
			},
		},
	})
}

func TestAccFSxOpenzfsFileSystem_tags(t *testing.T) {
	var filesystem1, filesystem2, filesystem3 fsx.FileSystem
	resourceName := "aws_fsx_openzfs_file_system.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); acctest.PreCheckPartitionHasService(fsx.EndpointsID, t) },
		ErrorCheck:        acctest.ErrorCheck(t, fsx.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckOpenzfsFileSystemDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccOpenZFSFileSystemConfig_tags1(rName, "key1", "value1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckOpenzfsFileSystemExists(resourceName, &filesystem1),
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
				Config: testAccOpenZFSFileSystemConfig_tags2(rName, "key1", "value1updated", "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckOpenzfsFileSystemExists(resourceName, &filesystem2),
					testAccCheckOpenzfsFileSystemNotRecreated(&filesystem1, &filesystem2),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1updated"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
			{
				Config: testAccOpenZFSFileSystemConfig_tags1(rName, "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckOpenzfsFileSystemExists(resourceName, &filesystem3),
					testAccCheckOpenzfsFileSystemNotRecreated(&filesystem2, &filesystem3),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
		},
	})
}

func TestAccFSxOpenzfsFileSystem_copyTags(t *testing.T) {
	var filesystem1, filesystem2 fsx.FileSystem
	resourceName := "aws_fsx_openzfs_file_system.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); acctest.PreCheckPartitionHasService(fsx.EndpointsID, t) },
		ErrorCheck:        acctest.ErrorCheck(t, fsx.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckOpenzfsFileSystemDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccOpenZFSFileSystemConfig_copyTags(rName, "key1", "value1", "true"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckOpenzfsFileSystemExists(resourceName, &filesystem1),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1"),
					resource.TestCheckResourceAttr(resourceName, "copy_tags_to_backups", "true"),
					resource.TestCheckResourceAttr(resourceName, "copy_tags_to_volumes", "true"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"security_group_ids"},
			},
			{
				Config: testAccOpenZFSFileSystemConfig_copyTags(rName, "key1", "value1", "false"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckOpenzfsFileSystemExists(resourceName, &filesystem2),
					testAccCheckOpenzfsFileSystemNotRecreated(&filesystem1, &filesystem2),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1"),
					resource.TestCheckResourceAttr(resourceName, "copy_tags_to_backups", "false"),
					resource.TestCheckResourceAttr(resourceName, "copy_tags_to_volumes", "false"),
				),
			},
		},
	})
}

func TestAccFSxOpenzfsFileSystem_throughput(t *testing.T) {
	var filesystem1, filesystem2 fsx.FileSystem
	resourceName := "aws_fsx_openzfs_file_system.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); acctest.PreCheckPartitionHasService(fsx.EndpointsID, t) },
		ErrorCheck:        acctest.ErrorCheck(t, fsx.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckOpenzfsFileSystemDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccOpenZFSFileSystemConfig_throughput(rName, 64),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckOpenzfsFileSystemExists(resourceName, &filesystem1),
					resource.TestCheckResourceAttr(resourceName, "throughput_capacity", "64"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"security_group_ids"},
			},
			{
				Config: testAccOpenZFSFileSystemConfig_throughput(rName, 128),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckOpenzfsFileSystemExists(resourceName, &filesystem2),
					testAccCheckOpenzfsFileSystemNotRecreated(&filesystem1, &filesystem2),
					resource.TestCheckResourceAttr(resourceName, "throughput_capacity", "128"),
				),
			},
		},
	})
}

func TestAccFSxOpenzfsFileSystem_storageType(t *testing.T) {
	var filesystem1 fsx.FileSystem
	resourceName := "aws_fsx_openzfs_file_system.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); acctest.PreCheckPartitionHasService(fsx.EndpointsID, t) },
		ErrorCheck:        acctest.ErrorCheck(t, fsx.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckOpenzfsFileSystemDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccOpenZFSFileSystemConfig_storageType(rName, "SSD"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckOpenzfsFileSystemExists(resourceName, &filesystem1),
					resource.TestCheckResourceAttr(resourceName, "storage_type", "SSD"),
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

func TestAccFSxOpenzfsFileSystem_weeklyMaintenanceStartTime(t *testing.T) {
	var filesystem1, filesystem2 fsx.FileSystem
	resourceName := "aws_fsx_openzfs_file_system.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); acctest.PreCheckPartitionHasService(fsx.EndpointsID, t) },
		ErrorCheck:        acctest.ErrorCheck(t, fsx.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckOpenzfsFileSystemDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccOpenZFSFileSystemConfig_weeklyMaintenanceStartTime(rName, "1:01:01"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckOpenzfsFileSystemExists(resourceName, &filesystem1),
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
				Config: testAccOpenZFSFileSystemConfig_weeklyMaintenanceStartTime(rName, "2:02:02"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckOpenzfsFileSystemExists(resourceName, &filesystem2),
					testAccCheckOpenzfsFileSystemNotRecreated(&filesystem1, &filesystem2),
					resource.TestCheckResourceAttr(resourceName, "weekly_maintenance_start_time", "2:02:02"),
				),
			},
		},
	})
}

func TestAccFSxOpenzfsFileSystem_automaticBackupRetentionDays(t *testing.T) {
	var filesystem1, filesystem2 fsx.FileSystem
	resourceName := "aws_fsx_openzfs_file_system.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); acctest.PreCheckPartitionHasService(fsx.EndpointsID, t) },
		ErrorCheck:        acctest.ErrorCheck(t, fsx.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckOpenzfsFileSystemDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccOpenZFSFileSystemConfig_automaticBackupRetentionDays(rName, 90),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckOpenzfsFileSystemExists(resourceName, &filesystem1),
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
				Config: testAccOpenZFSFileSystemConfig_automaticBackupRetentionDays(rName, 0),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckOpenzfsFileSystemExists(resourceName, &filesystem2),
					testAccCheckOpenzfsFileSystemNotRecreated(&filesystem1, &filesystem2),
					resource.TestCheckResourceAttr(resourceName, "automatic_backup_retention_days", "0"),
				),
			},
			{
				Config: testAccOpenZFSFileSystemConfig_automaticBackupRetentionDays(rName, 1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckOpenzfsFileSystemExists(resourceName, &filesystem1),
					resource.TestCheckResourceAttr(resourceName, "automatic_backup_retention_days", "1"),
				),
			},
		},
	})
}

func TestAccFSxOpenzfsFileSystem_kmsKeyID(t *testing.T) {
	var filesystem fsx.FileSystem
	resourceName := "aws_fsx_openzfs_file_system.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); acctest.PreCheckPartitionHasService(fsx.EndpointsID, t) },
		ErrorCheck:        acctest.ErrorCheck(t, fsx.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckOpenzfsFileSystemDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccOpenZFSFileSystemConfig_kmsKeyID(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckOpenzfsFileSystemExists(resourceName, &filesystem),
					resource.TestCheckResourceAttrPair(resourceName, "kms_key_id", "aws_kms_key.test", "arn"),
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

func TestAccFSxOpenzfsFileSystem_dailyAutomaticBackupStartTime(t *testing.T) {
	var filesystem1, filesystem2 fsx.FileSystem
	resourceName := "aws_fsx_openzfs_file_system.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); acctest.PreCheckPartitionHasService(fsx.EndpointsID, t) },
		ErrorCheck:        acctest.ErrorCheck(t, fsx.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckOpenzfsFileSystemDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccOpenZFSFileSystemConfig_dailyAutomaticBackupStartTime(rName, "01:01"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckOpenzfsFileSystemExists(resourceName, &filesystem1),
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
				Config: testAccOpenZFSFileSystemConfig_dailyAutomaticBackupStartTime(rName, "02:02"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckOpenzfsFileSystemExists(resourceName, &filesystem2),
					testAccCheckOpenzfsFileSystemNotRecreated(&filesystem1, &filesystem2),
					resource.TestCheckResourceAttr(resourceName, "daily_automatic_backup_start_time", "02:02"),
				),
			},
		},
	})
}

func testAccCheckOpenzfsFileSystemExists(resourceName string, fs *fsx.FileSystem) resource.TestCheckFunc {
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
			return fmt.Errorf("FSx Openzfs File System (%s) not found", rs.Primary.ID)
		}

		*fs = *filesystem

		return nil
	}
}

func testAccCheckOpenzfsFileSystemNotRecreated(i, j *fsx.FileSystem) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if aws.StringValue(i.FileSystemId) != aws.StringValue(j.FileSystemId) {
			return fmt.Errorf("FSx OpenZFS File System (%s) recreated", aws.StringValue(i.FileSystemId))
		}

		return nil
	}
}

func testAccCheckOpenzfsFileSystemRecreated(i, j *fsx.FileSystem) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if aws.StringValue(i.FileSystemId) == aws.StringValue(j.FileSystemId) {
			return fmt.Errorf("FSx OpenZFS File System (%s) not recreated", aws.StringValue(i.FileSystemId))
		}

		return nil
	}
}

func testAccCheckOpenzfsFileSystemDestroy(s *terraform.State) error {
	conn := acctest.Provider.Meta().(*conns.AWSClient).FSxConn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_fsx_openzfs_file_system" {
			continue
		}

		filesystem, err := tffsx.FindFileSystemByID(conn, rs.Primary.ID)
		if tfresource.NotFound(err) {
			continue
		}

		if filesystem != nil {
			return fmt.Errorf("FSx OpenZFS File System (%s) still exists", rs.Primary.ID)
		}
	}
	return nil
}

func testAccOpenzfsFileSystemBaseConfig(rName string) string {
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

`, rName))
}

func testAccOpenZFSFileSystemConfig_basic(rName string) string {
	return acctest.ConfigCompose(testAccOpenzfsFileSystemBaseConfig(rName), `
resource "aws_fsx_openzfs_file_system" "test" {
  storage_capacity    = 64
  subnet_ids          = [aws_subnet.test1.id]
  deployment_type     = "SINGLE_AZ_1"
  throughput_capacity = 64
}
`)
}

func testAccOpenZFSFileSystemConfig_diskIOPSConfiguration(rName string) string {
	return acctest.ConfigCompose(testAccOpenzfsFileSystemBaseConfig(rName), `
resource "aws_fsx_openzfs_file_system" "test" {
  storage_capacity    = 64
  storage_type        = "SSD"
  subnet_ids          = [aws_subnet.test1.id]
  deployment_type     = "SINGLE_AZ_1"
  throughput_capacity = 64
  disk_iops_configuration {
    mode = "USER_PROVISIONED"
    iops = 3072
  }
}
`)
}

func testAccOpenZFSFileSystemConfig_securityGroupIDs1(rName string) string {
	return acctest.ConfigCompose(testAccOpenzfsFileSystemBaseConfig(rName), fmt.Sprintf(`
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

  tags = {
    Name = %[1]q
  }
}

resource "aws_fsx_openzfs_file_system" "test" {
  security_group_ids  = [aws_security_group.test1.id]
  storage_capacity    = 64
  subnet_ids          = [aws_subnet.test1.id]
  deployment_type     = "SINGLE_AZ_1"
  throughput_capacity = 512
  storage_type        = "SSD"

  tags = {
    Name = %[1]q
  }
}
`, rName))
}

func testAccOpenZFSFileSystemConfig_securityGroupIDs2(rName string) string {
	return acctest.ConfigCompose(testAccOpenzfsFileSystemBaseConfig(rName), fmt.Sprintf(`
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

  tags = {
    Name = %[1]q
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

  tags = {
    Name = %[1]q
  }
}

resource "aws_fsx_openzfs_file_system" "test" {
  security_group_ids  = [aws_security_group.test1.id, aws_security_group.test2.id]
  storage_capacity    = 64
  subnet_ids          = [aws_subnet.test1.id]
  deployment_type     = "SINGLE_AZ_1"
  throughput_capacity = 512
  storage_type        = "SSD"

  tags = {
    Name = %[1]q
  }
}
`, rName))
}

func testAccOpenZFSFileSystemConfig_tags1(rName, tagKey1, tagValue1 string) string {
	return acctest.ConfigCompose(testAccOpenzfsFileSystemBaseConfig(rName), fmt.Sprintf(`
resource "aws_fsx_openzfs_file_system" "test" {
  storage_capacity    = 64
  subnet_ids          = [aws_subnet.test1.id]
  deployment_type     = "SINGLE_AZ_1"
  throughput_capacity = 512
  storage_type        = "SSD"

  tags = {
    %[1]q = %[2]q
  }
}
`, tagKey1, tagValue1))
}

func testAccOpenZFSFileSystemConfig_tags2(rName, tagKey1, tagValue1, tagKey2, tagValue2 string) string {
	return acctest.ConfigCompose(testAccOpenzfsFileSystemBaseConfig(rName), fmt.Sprintf(`
resource "aws_fsx_openzfs_file_system" "test" {
  storage_capacity    = 64
  subnet_ids          = [aws_subnet.test1.id]
  deployment_type     = "SINGLE_AZ_1"
  throughput_capacity = 512
  storage_type        = "SSD"

  tags = {
    %[1]q = %[2]q
    %[3]q = %[4]q
  }
}
`, tagKey1, tagValue1, tagKey2, tagValue2))
}

func testAccOpenZFSFileSystemConfig_copyTags(rName, tagKey1, tagValue1, copyTags string) string {
	return acctest.ConfigCompose(testAccOpenzfsFileSystemBaseConfig(rName), fmt.Sprintf(`
resource "aws_fsx_openzfs_file_system" "test" {
  storage_capacity     = 64
  subnet_ids           = [aws_subnet.test1.id]
  deployment_type      = "SINGLE_AZ_1"
  throughput_capacity  = 512
  storage_type         = "SSD"
  copy_tags_to_backups = %[3]s
  copy_tags_to_volumes = %[3]s

  tags = {
    %[1]q = %[2]q
  }
}
`, tagKey1, tagValue1, copyTags))
}

func testAccOpenZFSFileSystemConfig_weeklyMaintenanceStartTime(rName, weeklyMaintenanceStartTime string) string {
	return acctest.ConfigCompose(testAccOpenzfsFileSystemBaseConfig(rName), fmt.Sprintf(`
resource "aws_fsx_openzfs_file_system" "test" {
  storage_capacity              = 64
  subnet_ids                    = [aws_subnet.test1.id]
  deployment_type               = "SINGLE_AZ_1"
  throughput_capacity           = 512
  storage_type                  = "SSD"
  weekly_maintenance_start_time = %[2]q

  tags = {
    Name = %[1]q
  }
}
`, rName, weeklyMaintenanceStartTime))
}

func testAccOpenZFSFileSystemConfig_dailyAutomaticBackupStartTime(rName, dailyAutomaticBackupStartTime string) string {
	return acctest.ConfigCompose(testAccOpenzfsFileSystemBaseConfig(rName), fmt.Sprintf(`
resource "aws_fsx_openzfs_file_system" "test" {
  storage_capacity                  = 64
  subnet_ids                        = [aws_subnet.test1.id]
  deployment_type                   = "SINGLE_AZ_1"
  throughput_capacity               = 512
  storage_type                      = "SSD"
  daily_automatic_backup_start_time = %[2]q
  automatic_backup_retention_days   = 1

  tags = {
    Name = %[1]q
  }
}
`, rName, dailyAutomaticBackupStartTime))
}

func testAccOpenZFSFileSystemConfig_automaticBackupRetentionDays(rName string, retention int) string {
	return acctest.ConfigCompose(testAccOpenzfsFileSystemBaseConfig(rName), fmt.Sprintf(`
resource "aws_fsx_openzfs_file_system" "test" {
  storage_capacity                = 64
  subnet_ids                      = [aws_subnet.test1.id]
  deployment_type                 = "SINGLE_AZ_1"
  throughput_capacity             = 512
  storage_type                    = "SSD"
  automatic_backup_retention_days = %[2]d

  tags = {
    Name = %[1]q
  }
}
`, rName, retention))
}

func testAccOpenZFSFileSystemConfig_kmsKeyID(rName string) string {
	return acctest.ConfigCompose(testAccOpenzfsFileSystemBaseConfig(rName), fmt.Sprintf(`
resource "aws_kms_key" "test" {
  description             = %[1]q
  deletion_window_in_days = 7
}

resource "aws_fsx_openzfs_file_system" "test" {
  storage_capacity    = 64
  subnet_ids          = [aws_subnet.test1.id]
  deployment_type     = "SINGLE_AZ_1"
  throughput_capacity = 64
  storage_type        = "SSD"
  kms_key_id          = aws_kms_key.test.arn

  tags = {
    Name = %[1]q
  }
}
`, rName))
}

func testAccOpenZFSFileSystemConfig_throughput(rName string, throughput int) string {
	return acctest.ConfigCompose(testAccOpenzfsFileSystemBaseConfig(rName), fmt.Sprintf(`
resource "aws_fsx_openzfs_file_system" "test" {
  storage_capacity    = 64
  subnet_ids          = [aws_subnet.test1.id]
  deployment_type     = "SINGLE_AZ_1"
  throughput_capacity = %[2]d

  tags = {
    Name = %[1]q
  }
}
`, rName, throughput))
}

func testAccOpenZFSFileSystemConfig_storageType(rName, storageType string) string {
	return acctest.ConfigCompose(testAccOpenzfsFileSystemBaseConfig(rName), fmt.Sprintf(`
resource "aws_fsx_openzfs_file_system" "test" {
  storage_capacity    = 64
  subnet_ids          = [aws_subnet.test1.id]
  deployment_type     = "SINGLE_AZ_1"
  storage_type        = %[2]q
  throughput_capacity = 64

  tags = {
    Name = %[1]q
  }
}
`, rName, storageType))
}

func testAccOpenZFSFileSystemConfig_rootVolume1(rName, dataCompression, readOnly string, quotaSize int) string {
	return acctest.ConfigCompose(testAccOpenzfsFileSystemBaseConfig(rName), fmt.Sprintf(`
resource "aws_fsx_openzfs_file_system" "test" {
  storage_capacity    = 64
  subnet_ids          = [aws_subnet.test1.id]
  deployment_type     = "SINGLE_AZ_1"
  throughput_capacity = 64
  root_volume_configuration {
    copy_tags_to_snapshots = true
    data_compression_type  = %[2]q
    nfs_exports {
      client_configurations {
        clients = "10.0.1.0/24"
        options = ["sync", "rw"]
      }
    }
    read_only = %[3]s
    user_and_group_quotas {
      id                         = 10
      storage_capacity_quota_gib = %[4]d
      type                       = "USER"
    }
  }

  tags = {
    Name = %[1]q
  }
}
`, rName, dataCompression, readOnly, quotaSize))
}

func testAccOpenZFSFileSystemConfig_rootVolume2(rName, dataCompression, readOnly string, quotaSize int) string {
	return acctest.ConfigCompose(testAccOpenzfsFileSystemBaseConfig(rName), fmt.Sprintf(`
resource "aws_fsx_openzfs_file_system" "test" {
  storage_capacity    = 64
  subnet_ids          = [aws_subnet.test1.id]
  deployment_type     = "SINGLE_AZ_1"
  throughput_capacity = 64
  root_volume_configuration {
    copy_tags_to_snapshots = true
    data_compression_type  = %[2]q
    nfs_exports {
      client_configurations {
        clients = "10.0.1.0/24"
        options = ["async", "rw"]
      }
    }
    read_only = %[3]s
    user_and_group_quotas {
      id                         = 10
      storage_capacity_quota_gib = %[4]d
      type                       = "USER"
    }
  }

  tags = {
    Name = %[1]q
  }
}
`, rName, dataCompression, readOnly, quotaSize))
}

func testAccOpenZFSFileSystemConfig_rootVolume3Client(rName, dataCompression, readOnly string, userQuota, groupQuota int) string {
	return acctest.ConfigCompose(testAccOpenzfsFileSystemBaseConfig(rName), fmt.Sprintf(`
resource "aws_fsx_openzfs_file_system" "test" {
  storage_capacity    = 64
  subnet_ids          = [aws_subnet.test1.id]
  deployment_type     = "SINGLE_AZ_1"
  throughput_capacity = 64
  root_volume_configuration {
    copy_tags_to_snapshots = true
    data_compression_type  = %[2]q
    nfs_exports {
      client_configurations {
        clients = "10.0.1.0/24"
        options = ["async", "rw"]
      }
      client_configurations {
        clients = "*"
        options = ["sync", "rw"]
      }
    }
    read_only = %[3]s
    user_and_group_quotas {
      id                         = 10
      storage_capacity_quota_gib = %[4]d
      type                       = "USER"
    }
    user_and_group_quotas {
      id                         = 20
      storage_capacity_quota_gib = %[5]d
      type                       = "GROUP"
    }
    user_and_group_quotas {
      id                         = 5
      storage_capacity_quota_gib = %[5]d
      type                       = "GROUP"
    }
    user_and_group_quotas {
      id                         = 100
      storage_capacity_quota_gib = %[4]d
      type                       = "USER"
    }
  }

  tags = {
    Name = %[1]q
  }
}
`, rName, dataCompression, readOnly, userQuota, groupQuota))
}

func testAccOpenZFSFileSystemConfig_rootVolume4(rName, dataCompression, readOnly string, userQuota, groupQuota int) string {
	return acctest.ConfigCompose(testAccOpenzfsFileSystemBaseConfig(rName), fmt.Sprintf(`
resource "aws_fsx_openzfs_file_system" "test" {
  storage_capacity    = 64
  subnet_ids          = [aws_subnet.test1.id]
  deployment_type     = "SINGLE_AZ_1"
  throughput_capacity = 64
  root_volume_configuration {
    copy_tags_to_snapshots = true
    data_compression_type  = %[2]q

    user_and_group_quotas {
      id                         = 10
      storage_capacity_quota_gib = %[4]d
      type                       = "USER"
    }
    user_and_group_quotas {
      id                         = 20
      storage_capacity_quota_gib = %[5]d
      type                       = "GROUP"
    }
    user_and_group_quotas {
      id                         = 5
      storage_capacity_quota_gib = %[5]d
      type                       = "GROUP"
    }
    user_and_group_quotas {
      id                         = 100
      storage_capacity_quota_gib = %[4]d
      type                       = "USER"
    }
  }

  tags = {
    Name = %[1]q
  }
}
`, rName, dataCompression, readOnly, userQuota, groupQuota))
}
