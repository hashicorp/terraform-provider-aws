// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package fsx_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/aws"
	awstypes "github.com/aws/aws-sdk-go-v2/service/fsx/types"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tffsx "github.com/hashicorp/terraform-provider-aws/internal/service/fsx"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func init() {
	acctest.RegisterServiceErrorCheckFunc(names.FSxServiceID, testAccErrorCheckSkip)
}

func testAccErrorCheckSkip(t *testing.T) resource.ErrorCheckFunc {
	return acctest.ErrorCheckSkipMessagesContaining(t,
		"Amazon FSx does not currently support OpenZFS file system creation in the following Availability Zones",
		// "ServiceLimitExceeded: Account 123456789012 can have at most 10240 MB/s of throughput capacity total across file systems"
		"throughput capacity total across file systems",
	)
}

func TestAccFSxOpenZFSFileSystem_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var filesystem awstypes.FileSystem
	resourceName := "aws_fsx_openzfs_file_system.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckPartitionHasService(t, names.FSxEndpointID) },
		ErrorCheck:               acctest.ErrorCheck(t, names.FSxServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckOpenZFSFileSystemDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccOpenZFSFileSystemConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckOpenZFSFileSystemExists(ctx, resourceName, &filesystem),
					acctest.MatchResourceAttrRegionalARN(resourceName, names.AttrARN, "fsx", regexache.MustCompile(`file-system/fs-.+`)),
					resource.TestCheckResourceAttr(resourceName, "automatic_backup_retention_days", acctest.Ct0),
					resource.TestCheckNoResourceAttr(resourceName, "backup_id"),
					resource.TestCheckResourceAttr(resourceName, "copy_tags_to_backups", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "copy_tags_to_volumes", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "daily_automatic_backup_start_time", ""),
					resource.TestCheckResourceAttr(resourceName, "deployment_type", string(awstypes.OpenZFSDeploymentTypeSingleAz1)),
					resource.TestCheckResourceAttr(resourceName, "disk_iops_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "disk_iops_configuration.0.iops", "192"),
					resource.TestCheckResourceAttr(resourceName, "disk_iops_configuration.0.mode", "AUTOMATIC"),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrDNSName),
					resource.TestCheckResourceAttr(resourceName, "endpoint_ip_address", ""),
					resource.TestCheckResourceAttr(resourceName, "endpoint_ip_address_range", ""),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrKMSKeyID),
					resource.TestCheckResourceAttr(resourceName, "network_interface_ids.#", acctest.Ct1),
					acctest.CheckResourceAttrAccountID(resourceName, names.AttrOwnerID),
					resource.TestCheckResourceAttr(resourceName, "preferred_subnet_id", ""),
					resource.TestCheckResourceAttr(resourceName, "root_volume_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "root_volume_configuration.0.data_compression_type", "NONE"),
					resource.TestCheckResourceAttr(resourceName, "root_volume_configuration.0.nfs_exports.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "root_volume_configuration.0.nfs_exports.0.client_configurations.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "root_volume_configuration.0.nfs_exports.0.client_configurations.0.clients", "*"),
					resource.TestCheckResourceAttr(resourceName, "root_volume_configuration.0.nfs_exports.0.client_configurations.0.options.#", acctest.Ct2),
					resource.TestCheckResourceAttr(resourceName, "root_volume_configuration.0.nfs_exports.0.client_configurations.0.options.0", "rw"),
					resource.TestCheckResourceAttr(resourceName, "root_volume_configuration.0.nfs_exports.0.client_configurations.0.options.1", "crossmnt"),
					resource.TestCheckResourceAttr(resourceName, "root_volume_configuration.0.read_only", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "root_volume_configuration.0.record_size_kib", "128"),
					resource.TestCheckResourceAttr(resourceName, "root_volume_configuration.0.user_and_group_quotas.#", acctest.Ct2),
					resource.TestCheckResourceAttrSet(resourceName, "root_volume_id"),
					resource.TestCheckResourceAttr(resourceName, "route_table_ids.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "security_group_ids.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "skip_final_backup", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "storage_capacity", "64"),
					resource.TestCheckResourceAttr(resourceName, names.AttrStorageType, string(awstypes.StorageTypeSsd)),
					resource.TestCheckResourceAttr(resourceName, "subnet_ids.#", acctest.Ct1),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "subnet_ids.*", "aws_subnet.test.0", names.AttrID),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "throughput_capacity", "64"),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrVPCID, "aws_vpc.test", names.AttrID),
					resource.TestMatchResourceAttr(resourceName, "weekly_maintenance_start_time", regexache.MustCompile(`^\d:\d\d:\d\d$`)),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"delete_options",
					"final_backup_tags",
					"skip_final_backup",
				},
			},
		},
	})
}

func TestAccFSxOpenZFSFileSystem_diskIops(t *testing.T) {
	ctx := acctest.Context(t)
	var filesystem awstypes.FileSystem
	resourceName := "aws_fsx_openzfs_file_system.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckPartitionHasService(t, names.FSxEndpointID) },
		ErrorCheck:               acctest.ErrorCheck(t, names.FSxServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckOpenZFSFileSystemDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccOpenZFSFileSystemConfig_diskIOPSConfiguration(rName, 192),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckOpenZFSFileSystemExists(ctx, resourceName, &filesystem),
					resource.TestCheckResourceAttr(resourceName, "disk_iops_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "disk_iops_configuration.0.mode", "USER_PROVISIONED"),
					resource.TestCheckResourceAttr(resourceName, "disk_iops_configuration.0.iops", "192"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					names.AttrSecurityGroupIDs,
					"delete_options",
					"final_backup_tags",
					"skip_final_backup",
				},
			},
			{
				Config: testAccOpenZFSFileSystemConfig_diskIOPSConfiguration(rName, 200),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckOpenZFSFileSystemExists(ctx, resourceName, &filesystem),
					resource.TestCheckResourceAttr(resourceName, "disk_iops_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "disk_iops_configuration.0.mode", "USER_PROVISIONED"),
					resource.TestCheckResourceAttr(resourceName, "disk_iops_configuration.0.iops", "200"),
				),
			},
		},
	})
}

func TestAccFSxOpenZFSFileSystem_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	var filesystem awstypes.FileSystem
	resourceName := "aws_fsx_openzfs_file_system.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckPartitionHasService(t, names.FSxEndpointID) },
		ErrorCheck:               acctest.ErrorCheck(t, names.FSxServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckOpenZFSFileSystemDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccOpenZFSFileSystemConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckOpenZFSFileSystemExists(ctx, resourceName, &filesystem),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tffsx.ResourceOpenZFSFileSystem(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccFSxOpenZFSFileSystem_rootVolume(t *testing.T) {
	ctx := acctest.Context(t)
	var filesystem1, filesystem2, filesystem3 awstypes.FileSystem
	resourceName := "aws_fsx_openzfs_file_system.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckPartitionHasService(t, names.FSxEndpointID) },
		ErrorCheck:               acctest.ErrorCheck(t, names.FSxServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckOpenZFSFileSystemDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccOpenZFSFileSystemConfig_rootVolume1(rName, "NONE", acctest.CtFalse, 128),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckOpenZFSFileSystemExists(ctx, resourceName, &filesystem1),
					resource.TestCheckResourceAttr(resourceName, "root_volume_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "root_volume_configuration.0.data_compression_type", "NONE"),
					resource.TestCheckResourceAttr(resourceName, "root_volume_configuration.0.nfs_exports.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "root_volume_configuration.0.nfs_exports.0.client_configurations.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "root_volume_configuration.0.nfs_exports.0.client_configurations.0.clients", "10.0.1.0/24"),
					resource.TestCheckResourceAttr(resourceName, "root_volume_configuration.0.nfs_exports.0.client_configurations.0.options.#", acctest.Ct2),
					resource.TestCheckResourceAttr(resourceName, "root_volume_configuration.0.nfs_exports.0.client_configurations.0.options.0", "sync"),
					resource.TestCheckResourceAttr(resourceName, "root_volume_configuration.0.nfs_exports.0.client_configurations.0.options.1", "rw"),
					resource.TestCheckResourceAttr(resourceName, "root_volume_configuration.0.read_only", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "root_volume_configuration.0.record_size_kib", "128"),
					resource.TestCheckResourceAttr(resourceName, "root_volume_configuration.0.user_and_group_quotas.#", acctest.Ct3),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "root_volume_configuration.0.user_and_group_quotas.*", map[string]string{
						names.AttrID:                 acctest.Ct10,
						"storage_capacity_quota_gib": "128",
						names.AttrType:               "USER",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "root_volume_configuration.0.user_and_group_quotas.*", map[string]string{
						names.AttrID:                 acctest.Ct0,
						"storage_capacity_quota_gib": acctest.Ct0,
						names.AttrType:               "USER",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "root_volume_configuration.0.user_and_group_quotas.*", map[string]string{
						names.AttrID:                 acctest.Ct0,
						"storage_capacity_quota_gib": acctest.Ct0,
						names.AttrType:               "GROUP",
					}),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					names.AttrSecurityGroupIDs,
					"delete_options",
					"final_backup_tags",
					"skip_final_backup",
				},
			},
			{
				Config: testAccOpenZFSFileSystemConfig_rootVolume2(rName, "ZSTD", acctest.CtTrue, 256, 8),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckOpenZFSFileSystemExists(ctx, resourceName, &filesystem2),
					testAccCheckOpenZFSFileSystemNotRecreated(&filesystem1, &filesystem2),
					resource.TestCheckResourceAttr(resourceName, "root_volume_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "root_volume_configuration.0.data_compression_type", "ZSTD"),
					resource.TestCheckResourceAttr(resourceName, "root_volume_configuration.0.nfs_exports.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "root_volume_configuration.0.nfs_exports.0.client_configurations.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "root_volume_configuration.0.nfs_exports.0.client_configurations.0.clients", "10.0.1.0/24"),
					resource.TestCheckResourceAttr(resourceName, "root_volume_configuration.0.nfs_exports.0.client_configurations.0.options.#", acctest.Ct2),
					resource.TestCheckResourceAttr(resourceName, "root_volume_configuration.0.nfs_exports.0.client_configurations.0.options.0", "async"),
					resource.TestCheckResourceAttr(resourceName, "root_volume_configuration.0.nfs_exports.0.client_configurations.0.options.1", "rw"),
					resource.TestCheckResourceAttr(resourceName, "root_volume_configuration.0.read_only", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "root_volume_configuration.0.record_size_kib", "8"),
					resource.TestCheckResourceAttr(resourceName, "root_volume_configuration.0.user_and_group_quotas.#", acctest.Ct3),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "root_volume_configuration.0.user_and_group_quotas.*", map[string]string{
						names.AttrID:                 acctest.Ct10,
						"storage_capacity_quota_gib": "256",
						names.AttrType:               "USER",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "root_volume_configuration.0.user_and_group_quotas.*", map[string]string{
						names.AttrID:                 acctest.Ct0,
						"storage_capacity_quota_gib": acctest.Ct0,
						names.AttrType:               "USER",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "root_volume_configuration.0.user_and_group_quotas.*", map[string]string{
						names.AttrID:                 acctest.Ct0,
						"storage_capacity_quota_gib": acctest.Ct0,
						names.AttrType:               "GROUP",
					}),
				),
			},
			{
				Config: testAccOpenZFSFileSystemConfig_rootVolume3Client(rName, "NONE", acctest.CtFalse, 128, 1024, 512),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckOpenZFSFileSystemExists(ctx, resourceName, &filesystem3),
					testAccCheckOpenZFSFileSystemNotRecreated(&filesystem1, &filesystem3),
					resource.TestCheckResourceAttr(resourceName, "root_volume_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "root_volume_configuration.0.data_compression_type", "NONE"),
					resource.TestCheckResourceAttr(resourceName, "root_volume_configuration.0.nfs_exports.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "root_volume_configuration.0.nfs_exports.0.client_configurations.#", acctest.Ct2),
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
					resource.TestCheckResourceAttr(resourceName, "root_volume_configuration.0.read_only", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "root_volume_configuration.0.record_size_kib", "512"),
					resource.TestCheckResourceAttr(resourceName, "root_volume_configuration.0.user_and_group_quotas.#", "6"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "root_volume_configuration.0.user_and_group_quotas.*", map[string]string{
						names.AttrID:                 acctest.Ct10,
						"storage_capacity_quota_gib": "128",
						names.AttrType:               "USER",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "root_volume_configuration.0.user_and_group_quotas.*", map[string]string{
						names.AttrID:                 "20",
						"storage_capacity_quota_gib": "1024",
						names.AttrType:               "GROUP",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "root_volume_configuration.0.user_and_group_quotas.*", map[string]string{
						names.AttrID:                 "5",
						"storage_capacity_quota_gib": "1024",
						names.AttrType:               "GROUP",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "root_volume_configuration.0.user_and_group_quotas.*", map[string]string{
						names.AttrID:                 "100",
						"storage_capacity_quota_gib": "128",
						names.AttrType:               "USER",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "root_volume_configuration.0.user_and_group_quotas.*", map[string]string{
						names.AttrID:                 acctest.Ct0,
						"storage_capacity_quota_gib": acctest.Ct0,
						names.AttrType:               "USER",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "root_volume_configuration.0.user_and_group_quotas.*", map[string]string{
						names.AttrID:                 acctest.Ct0,
						"storage_capacity_quota_gib": acctest.Ct0,
						names.AttrType:               "GROUP",
					}),
				),
			},
			{
				Config: testAccOpenZFSFileSystemConfig_rootVolume4(rName, "NONE", acctest.CtFalse, 128, 1024),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckOpenZFSFileSystemExists(ctx, resourceName, &filesystem1),
					resource.TestCheckResourceAttr(resourceName, "root_volume_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "root_volume_configuration.0.data_compression_type", "NONE"),
					resource.TestCheckResourceAttr(resourceName, "root_volume_configuration.0.nfs_exports.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "root_volume_configuration.0.read_only", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "root_volume_configuration.0.record_size_kib", "128"),
					resource.TestCheckResourceAttr(resourceName, "root_volume_configuration.0.user_and_group_quotas.#", "6"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "root_volume_configuration.0.user_and_group_quotas.*", map[string]string{
						names.AttrID:                 acctest.Ct10,
						"storage_capacity_quota_gib": "128",
						names.AttrType:               "USER",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "root_volume_configuration.0.user_and_group_quotas.*", map[string]string{
						names.AttrID:                 "20",
						"storage_capacity_quota_gib": "1024",
						names.AttrType:               "GROUP",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "root_volume_configuration.0.user_and_group_quotas.*", map[string]string{
						names.AttrID:                 "5",
						"storage_capacity_quota_gib": "1024",
						names.AttrType:               "GROUP",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "root_volume_configuration.0.user_and_group_quotas.*", map[string]string{
						names.AttrID:                 "100",
						"storage_capacity_quota_gib": "128",
						names.AttrType:               "USER",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "root_volume_configuration.0.user_and_group_quotas.*", map[string]string{
						names.AttrID:                 acctest.Ct0,
						"storage_capacity_quota_gib": acctest.Ct0,
						names.AttrType:               "USER",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "root_volume_configuration.0.user_and_group_quotas.*", map[string]string{
						names.AttrID:                 acctest.Ct0,
						"storage_capacity_quota_gib": acctest.Ct0,
						names.AttrType:               "GROUP",
					}),
				),
			},
		},
	})
}

func TestAccFSxOpenZFSFileSystem_securityGroupIDs(t *testing.T) {
	ctx := acctest.Context(t)
	var filesystem1, filesystem2 awstypes.FileSystem
	resourceName := "aws_fsx_openzfs_file_system.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckPartitionHasService(t, names.FSxEndpointID) },
		ErrorCheck:               acctest.ErrorCheck(t, names.FSxServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckOpenZFSFileSystemDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccOpenZFSFileSystemConfig_securityGroupIDs1(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckOpenZFSFileSystemExists(ctx, resourceName, &filesystem1),
					resource.TestCheckResourceAttr(resourceName, "security_group_ids.#", acctest.Ct1),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					names.AttrSecurityGroupIDs,
					"delete_options",
					"final_backup_tags",
					"skip_final_backup",
				},
			},
			{
				Config: testAccOpenZFSFileSystemConfig_securityGroupIDs2(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckOpenZFSFileSystemExists(ctx, resourceName, &filesystem2),
					testAccCheckOpenZFSFileSystemRecreated(&filesystem1, &filesystem2),
					resource.TestCheckResourceAttr(resourceName, "security_group_ids.#", acctest.Ct2),
				),
			},
		},
	})
}

func TestAccFSxOpenZFSFileSystem_tags(t *testing.T) {
	ctx := acctest.Context(t)
	var filesystem1, filesystem2, filesystem3 awstypes.FileSystem
	resourceName := "aws_fsx_openzfs_file_system.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckPartitionHasService(t, names.FSxEndpointID) },
		ErrorCheck:               acctest.ErrorCheck(t, names.FSxServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckOpenZFSFileSystemDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccOpenZFSFileSystemConfig_tags1(rName, acctest.CtKey1, acctest.CtValue1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckOpenZFSFileSystemExists(ctx, resourceName, &filesystem1),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey1, acctest.CtValue1),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					names.AttrSecurityGroupIDs,
					"delete_options",
					"final_backup_tags",
					"skip_final_backup",
				},
			},
			{
				Config: testAccOpenZFSFileSystemConfig_tags2(rName, acctest.CtKey1, acctest.CtValue1Updated, acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckOpenZFSFileSystemExists(ctx, resourceName, &filesystem2),
					testAccCheckOpenZFSFileSystemNotRecreated(&filesystem1, &filesystem2),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct2),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey1, acctest.CtValue1Updated),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey2, acctest.CtValue2),
				),
			},
			{
				Config: testAccOpenZFSFileSystemConfig_tags1(rName, acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckOpenZFSFileSystemExists(ctx, resourceName, &filesystem3),
					testAccCheckOpenZFSFileSystemNotRecreated(&filesystem2, &filesystem3),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey2, acctest.CtValue2),
				),
			},
		},
	})
}

func TestAccFSxOpenZFSFileSystem_copyTags(t *testing.T) {
	ctx := acctest.Context(t)
	var filesystem1, filesystem2 awstypes.FileSystem
	resourceName := "aws_fsx_openzfs_file_system.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckPartitionHasService(t, names.FSxEndpointID) },
		ErrorCheck:               acctest.ErrorCheck(t, names.FSxServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckOpenZFSFileSystemDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccOpenZFSFileSystemConfig_copyTags(rName, acctest.CtKey1, acctest.CtValue1, acctest.CtTrue),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckOpenZFSFileSystemExists(ctx, resourceName, &filesystem1),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey1, acctest.CtValue1),
					resource.TestCheckResourceAttr(resourceName, "copy_tags_to_backups", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "copy_tags_to_volumes", acctest.CtTrue),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					names.AttrSecurityGroupIDs,
					"delete_options",
					"final_backup_tags",
					"skip_final_backup",
				},
			},
			{
				Config: testAccOpenZFSFileSystemConfig_copyTags(rName, acctest.CtKey1, acctest.CtValue1, acctest.CtFalse),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckOpenZFSFileSystemExists(ctx, resourceName, &filesystem2),
					testAccCheckOpenZFSFileSystemNotRecreated(&filesystem1, &filesystem2),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey1, acctest.CtValue1),
					resource.TestCheckResourceAttr(resourceName, "copy_tags_to_backups", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "copy_tags_to_volumes", acctest.CtFalse),
				),
			},
		},
	})
}

func TestAccFSxOpenZFSFileSystem_throughput(t *testing.T) {
	ctx := acctest.Context(t)
	var filesystem1, filesystem2 awstypes.FileSystem
	resourceName := "aws_fsx_openzfs_file_system.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckPartitionHasService(t, names.FSxEndpointID) },
		ErrorCheck:               acctest.ErrorCheck(t, names.FSxServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckOpenZFSFileSystemDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccOpenZFSFileSystemConfig_throughput(rName, 64),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckOpenZFSFileSystemExists(ctx, resourceName, &filesystem1),
					resource.TestCheckResourceAttr(resourceName, "throughput_capacity", "64"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					names.AttrSecurityGroupIDs,
					"delete_options",
					"final_backup_tags",
					"skip_final_backup",
				},
			},
			{
				Config: testAccOpenZFSFileSystemConfig_throughput(rName, 128),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckOpenZFSFileSystemExists(ctx, resourceName, &filesystem2),
					testAccCheckOpenZFSFileSystemNotRecreated(&filesystem1, &filesystem2),
					resource.TestCheckResourceAttr(resourceName, "throughput_capacity", "128"),
				),
			},
		},
	})
}

func TestAccFSxOpenZFSFileSystem_storageType(t *testing.T) {
	ctx := acctest.Context(t)
	var filesystem1 awstypes.FileSystem
	resourceName := "aws_fsx_openzfs_file_system.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckPartitionHasService(t, names.FSxEndpointID) },
		ErrorCheck:               acctest.ErrorCheck(t, names.FSxServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckOpenZFSFileSystemDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccOpenZFSFileSystemConfig_storageType(rName, "SSD"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckOpenZFSFileSystemExists(ctx, resourceName, &filesystem1),
					resource.TestCheckResourceAttr(resourceName, names.AttrStorageType, "SSD"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					names.AttrSecurityGroupIDs,
					"delete_options",
					"final_backup_tags",
					"skip_final_backup",
				},
			},
		},
	})
}

func TestAccFSxOpenZFSFileSystem_weeklyMaintenanceStartTime(t *testing.T) {
	ctx := acctest.Context(t)
	var filesystem1, filesystem2 awstypes.FileSystem
	resourceName := "aws_fsx_openzfs_file_system.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckPartitionHasService(t, names.FSxEndpointID) },
		ErrorCheck:               acctest.ErrorCheck(t, names.FSxServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckOpenZFSFileSystemDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccOpenZFSFileSystemConfig_weeklyMaintenanceStartTime(rName, "1:01:01"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckOpenZFSFileSystemExists(ctx, resourceName, &filesystem1),
					resource.TestCheckResourceAttr(resourceName, "weekly_maintenance_start_time", "1:01:01"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					names.AttrSecurityGroupIDs,
					"delete_options",
					"final_backup_tags",
					"skip_final_backup",
				},
			},
			{
				Config: testAccOpenZFSFileSystemConfig_weeklyMaintenanceStartTime(rName, "2:02:02"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckOpenZFSFileSystemExists(ctx, resourceName, &filesystem2),
					testAccCheckOpenZFSFileSystemNotRecreated(&filesystem1, &filesystem2),
					resource.TestCheckResourceAttr(resourceName, "weekly_maintenance_start_time", "2:02:02"),
				),
			},
		},
	})
}

func TestAccFSxOpenZFSFileSystem_automaticBackupRetentionDays(t *testing.T) {
	ctx := acctest.Context(t)
	var filesystem1, filesystem2 awstypes.FileSystem
	resourceName := "aws_fsx_openzfs_file_system.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckPartitionHasService(t, names.FSxEndpointID) },
		ErrorCheck:               acctest.ErrorCheck(t, names.FSxServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckOpenZFSFileSystemDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccOpenZFSFileSystemConfig_automaticBackupRetentionDays(rName, 90),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckOpenZFSFileSystemExists(ctx, resourceName, &filesystem1),
					resource.TestCheckResourceAttr(resourceName, "automatic_backup_retention_days", "90"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					names.AttrSecurityGroupIDs,
					"delete_options",
					"final_backup_tags",
					"skip_final_backup",
				},
			},
			{
				Config: testAccOpenZFSFileSystemConfig_automaticBackupRetentionDays(rName, 0),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckOpenZFSFileSystemExists(ctx, resourceName, &filesystem2),
					testAccCheckOpenZFSFileSystemNotRecreated(&filesystem1, &filesystem2),
					resource.TestCheckResourceAttr(resourceName, "automatic_backup_retention_days", acctest.Ct0),
				),
			},
			{
				Config: testAccOpenZFSFileSystemConfig_automaticBackupRetentionDays(rName, 1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckOpenZFSFileSystemExists(ctx, resourceName, &filesystem1),
					resource.TestCheckResourceAttr(resourceName, "automatic_backup_retention_days", acctest.Ct1),
				),
			},
		},
	})
}

func TestAccFSxOpenZFSFileSystem_kmsKeyID(t *testing.T) {
	ctx := acctest.Context(t)
	var filesystem awstypes.FileSystem
	resourceName := "aws_fsx_openzfs_file_system.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckPartitionHasService(t, names.FSxEndpointID) },
		ErrorCheck:               acctest.ErrorCheck(t, names.FSxServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckOpenZFSFileSystemDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccOpenZFSFileSystemConfig_kmsKeyID(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckOpenZFSFileSystemExists(ctx, resourceName, &filesystem),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrKMSKeyID, "aws_kms_key.test", names.AttrARN),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					names.AttrSecurityGroupIDs,
					"delete_options",
					"final_backup_tags",
					"skip_final_backup",
				},
			},
		},
	})
}

func TestAccFSxOpenZFSFileSystem_dailyAutomaticBackupStartTime(t *testing.T) {
	ctx := acctest.Context(t)
	var filesystem1, filesystem2 awstypes.FileSystem
	resourceName := "aws_fsx_openzfs_file_system.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckPartitionHasService(t, names.FSxEndpointID) },
		ErrorCheck:               acctest.ErrorCheck(t, names.FSxServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckOpenZFSFileSystemDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccOpenZFSFileSystemConfig_dailyAutomaticBackupStartTime(rName, "01:01"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckOpenZFSFileSystemExists(ctx, resourceName, &filesystem1),
					resource.TestCheckResourceAttr(resourceName, "daily_automatic_backup_start_time", "01:01"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					names.AttrSecurityGroupIDs,
					"delete_options",
					"final_backup_tags",
					"skip_final_backup",
				},
			},
			{
				Config: testAccOpenZFSFileSystemConfig_dailyAutomaticBackupStartTime(rName, "02:02"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckOpenZFSFileSystemExists(ctx, resourceName, &filesystem2),
					testAccCheckOpenZFSFileSystemNotRecreated(&filesystem1, &filesystem2),
					resource.TestCheckResourceAttr(resourceName, "daily_automatic_backup_start_time", "02:02"),
				),
			},
		},
	})
}

func TestAccFSxOpenZFSFileSystem_throughputCapacity(t *testing.T) {
	ctx := acctest.Context(t)
	var filesystem1, filesystem2 awstypes.FileSystem
	resourceName := "aws_fsx_openzfs_file_system.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckPartitionHasService(t, names.FSxEndpointID) },
		ErrorCheck:               acctest.ErrorCheck(t, names.FSxServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckOpenZFSFileSystemDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccOpenZFSFileSystemConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckOpenZFSFileSystemExists(ctx, resourceName, &filesystem1),
					resource.TestCheckResourceAttr(resourceName, "throughput_capacity", "64"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					names.AttrSecurityGroupIDs,
					"delete_options",
					"final_backup_tags",
					"skip_final_backup",
				},
			},
			{
				Config: testAccOpenZFSFileSystemConfig_throughputCapacity(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckOpenZFSFileSystemExists(ctx, resourceName, &filesystem2),
					testAccCheckOpenZFSFileSystemNotRecreated(&filesystem1, &filesystem2),
					resource.TestCheckResourceAttr(resourceName, "throughput_capacity", "128"),
				),
			},
		},
	})
}

func TestAccFSxOpenZFSFileSystem_storageCapacity(t *testing.T) {
	ctx := acctest.Context(t)
	var filesystem1, filesystem2 awstypes.FileSystem
	resourceName := "aws_fsx_openzfs_file_system.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckPartitionHasService(t, names.FSxEndpointID) },
		ErrorCheck:               acctest.ErrorCheck(t, names.FSxServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckOpenZFSFileSystemDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccOpenZFSFileSystemConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckOpenZFSFileSystemExists(ctx, resourceName, &filesystem1),
					resource.TestCheckResourceAttr(resourceName, "storage_capacity", "64"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					names.AttrSecurityGroupIDs,
					"delete_options",
					"final_backup_tags",
					"skip_final_backup",
				},
			},
			{
				Config: testAccOpenZFSFileSystemConfig_storageCapacity(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckOpenZFSFileSystemExists(ctx, resourceName, &filesystem2),
					testAccCheckOpenZFSFileSystemNotRecreated(&filesystem1, &filesystem2),
					resource.TestCheckResourceAttr(resourceName, "storage_capacity", "75"),
				),
			},
		},
	})
}

func TestAccFSxOpenZFSFileSystem_deploymentType(t *testing.T) {
	ctx := acctest.Context(t)
	var filesystem1, filesystem2 awstypes.FileSystem
	resourceName := "aws_fsx_openzfs_file_system.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckPartitionHasService(t, names.FSxEndpointID) },
		ErrorCheck:               acctest.ErrorCheck(t, names.FSxServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckOpenZFSFileSystemDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccOpenZFSFileSystemConfig_deploymentType(rName, "SINGLE_AZ_1", 64),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckOpenZFSFileSystemExists(ctx, resourceName, &filesystem1),
					resource.TestCheckResourceAttr(resourceName, "deployment_type", "SINGLE_AZ_1"),
					resource.TestCheckResourceAttr(resourceName, "throughput_capacity", "64"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					names.AttrSecurityGroupIDs,
					"delete_options",
					"final_backup_tags",
					"skip_final_backup",
				},
			},
			{
				Config: testAccOpenZFSFileSystemConfig_deploymentType(rName, "SINGLE_AZ_2", 160),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckOpenZFSFileSystemExists(ctx, resourceName, &filesystem2),
					testAccCheckOpenZFSFileSystemRecreated(&filesystem1, &filesystem2),
					resource.TestCheckResourceAttr(resourceName, "deployment_type", "SINGLE_AZ_2"),
					resource.TestCheckResourceAttr(resourceName, "throughput_capacity", "160"),
				),
			},
		},
	})
}

func TestAccFSxOpenZFSFileSystem_multiAZ(t *testing.T) {
	ctx := acctest.Context(t)
	var filesystem awstypes.FileSystem
	resourceName := "aws_fsx_openzfs_file_system.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckPartitionHasService(t, names.FSxEndpointID) },
		ErrorCheck:               acctest.ErrorCheck(t, names.FSxServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckOpenZFSFileSystemDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccOpenZFSFileSystemConfig_multiAZ(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckOpenZFSFileSystemExists(ctx, resourceName, &filesystem),
					acctest.MatchResourceAttrRegionalARN(resourceName, names.AttrARN, "fsx", regexache.MustCompile(`file-system/fs-.+`)),
					resource.TestCheckResourceAttr(resourceName, "automatic_backup_retention_days", acctest.Ct0),
					resource.TestCheckNoResourceAttr(resourceName, "backup_id"),
					resource.TestCheckResourceAttr(resourceName, "copy_tags_to_backups", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "copy_tags_to_volumes", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "daily_automatic_backup_start_time", ""),
					resource.TestCheckResourceAttr(resourceName, "deployment_type", string(awstypes.OpenZFSDeploymentTypeMultiAz1)),
					resource.TestCheckResourceAttr(resourceName, "disk_iops_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "disk_iops_configuration.0.iops", "192"),
					resource.TestCheckResourceAttr(resourceName, "disk_iops_configuration.0.mode", "AUTOMATIC"),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrDNSName),
					resource.TestCheckResourceAttrSet(resourceName, "endpoint_ip_address"),
					resource.TestCheckResourceAttrSet(resourceName, "endpoint_ip_address_range"),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrKMSKeyID),
					resource.TestCheckResourceAttr(resourceName, "network_interface_ids.#", acctest.Ct2),
					acctest.CheckResourceAttrAccountID(resourceName, names.AttrOwnerID),
					resource.TestCheckResourceAttrPair(resourceName, "preferred_subnet_id", "aws_subnet.test.0", names.AttrID),
					resource.TestCheckResourceAttr(resourceName, "root_volume_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "root_volume_configuration.0.data_compression_type", "NONE"),
					resource.TestCheckResourceAttr(resourceName, "root_volume_configuration.0.nfs_exports.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "root_volume_configuration.0.nfs_exports.0.client_configurations.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "root_volume_configuration.0.nfs_exports.0.client_configurations.0.clients", "*"),
					resource.TestCheckResourceAttr(resourceName, "root_volume_configuration.0.nfs_exports.0.client_configurations.0.options.#", acctest.Ct2),
					resource.TestCheckResourceAttr(resourceName, "root_volume_configuration.0.nfs_exports.0.client_configurations.0.options.0", "rw"),
					resource.TestCheckResourceAttr(resourceName, "root_volume_configuration.0.nfs_exports.0.client_configurations.0.options.1", "crossmnt"),
					resource.TestCheckResourceAttr(resourceName, "root_volume_configuration.0.read_only", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "root_volume_configuration.0.record_size_kib", "128"),
					resource.TestCheckResourceAttr(resourceName, "root_volume_configuration.0.user_and_group_quotas.#", acctest.Ct2),
					resource.TestCheckResourceAttrSet(resourceName, "root_volume_id"),
					resource.TestCheckResourceAttr(resourceName, "route_table_ids.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "security_group_ids.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "skip_final_backup", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "storage_capacity", "64"),
					resource.TestCheckResourceAttr(resourceName, names.AttrStorageType, string(awstypes.StorageTypeSsd)),
					resource.TestCheckResourceAttr(resourceName, "subnet_ids.#", acctest.Ct2),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "subnet_ids.*", "aws_subnet.test.0", names.AttrID),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "subnet_ids.*", "aws_subnet.test.1", names.AttrID),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "tags.Name", rName),
					resource.TestCheckResourceAttr(resourceName, "throughput_capacity", "160"),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrVPCID, "aws_vpc.test", names.AttrID),
					resource.TestMatchResourceAttr(resourceName, "weekly_maintenance_start_time", regexache.MustCompile(`^\d:\d\d:\d\d$`)),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"delete_options",
					"final_backup_tags",
					"skip_final_backup",
				},
			},
		},
	})
}

func TestAccFSxOpenZFSFileSystem_routeTableIDs(t *testing.T) {
	ctx := acctest.Context(t)
	var filesystem awstypes.FileSystem
	resourceName := "aws_fsx_openzfs_file_system.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckPartitionHasService(t, names.FSxEndpointID) },
		ErrorCheck:               acctest.ErrorCheck(t, names.FSxServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckOpenZFSFileSystemDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccOpenZFSFileSystemConfig_routeTableIDs(rName, 1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckOpenZFSFileSystemExists(ctx, resourceName, &filesystem),
					resource.TestCheckResourceAttr(resourceName, "route_table_ids.#", acctest.Ct1),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "route_table_ids.*", "aws_route_table.test.0", names.AttrID),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					names.AttrSecurityGroupIDs,
					"delete_options",
					"final_backup_tags",
					"skip_final_backup",
				},
			},
			{
				Config: testAccOpenZFSFileSystemConfig_routeTableIDs(rName, 2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckOpenZFSFileSystemExists(ctx, resourceName, &filesystem),
					resource.TestCheckResourceAttr(resourceName, "route_table_ids.#", acctest.Ct2),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "route_table_ids.*", "aws_route_table.test.0", names.AttrID),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "route_table_ids.*", "aws_route_table.test.1", names.AttrID),
				),
			},
			{
				Config: testAccOpenZFSFileSystemConfig_routeTableIDs(rName, 1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckOpenZFSFileSystemExists(ctx, resourceName, &filesystem),
					resource.TestCheckResourceAttr(resourceName, "route_table_ids.#", acctest.Ct1),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "route_table_ids.*", "aws_route_table.test.0", names.AttrID),
				),
			},
		},
	})
}

func TestAccFSxOpenZFSFileSystem_deleteConfig(t *testing.T) {
	ctx := acctest.Context(t)
	var filesystem awstypes.FileSystem
	resourceName := "aws_fsx_openzfs_file_system.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	acctest.SkipIfEnvVarNotSet(t, "AWS_FSX_CREATE_FINAL_BACKUP")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckPartitionHasService(t, names.FSxEndpointID) },
		ErrorCheck:               acctest.ErrorCheck(t, names.FSxServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckOpenZFSFileSystemDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccOpenZFSFileSystemConfig_deleteConfig(rName, acctest.CtKey1, acctest.CtValue1, acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckOpenZFSFileSystemExists(ctx, resourceName, &filesystem),
					resource.TestCheckResourceAttr(resourceName, "delete_options.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "delete_options.0", "DELETE_CHILD_VOLUMES_AND_SNAPSHOTS"),
					resource.TestCheckResourceAttr(resourceName, "final_backup_tags.%", acctest.Ct2),
					resource.TestCheckResourceAttr(resourceName, "final_backup_tags."+acctest.CtKey1, acctest.CtValue1),
					resource.TestCheckResourceAttr(resourceName, "final_backup_tags."+acctest.CtKey2, acctest.CtValue2),
					resource.TestCheckResourceAttr(resourceName, "skip_final_backup", acctest.CtFalse),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					names.AttrSecurityGroupIDs,
					"delete_options",
					"final_backup_tags",
					"skip_final_backup",
				},
			},
		},
	})
}

func testAccCheckOpenZFSFileSystemExists(ctx context.Context, n string, v *awstypes.FileSystem) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).FSxClient(ctx)

		output, err := tffsx.FindOpenZFSFileSystemByID(ctx, conn, rs.Primary.ID)

		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

func testAccCheckOpenZFSFileSystemDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).FSxClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_fsx_openzfs_file_system" {
				continue
			}

			_, err := tffsx.FindOpenZFSFileSystemByID(ctx, conn, rs.Primary.ID)

			if tfresource.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("FSx for OpenZFS File System %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccCheckOpenZFSFileSystemNotRecreated(i, j *awstypes.FileSystem) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if aws.ToString(i.FileSystemId) != aws.ToString(j.FileSystemId) {
			return fmt.Errorf("FSx for OpenZFS File System (%s) recreated", aws.ToString(i.FileSystemId))
		}

		return nil
	}
}

func testAccCheckOpenZFSFileSystemRecreated(i, j *awstypes.FileSystem) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if aws.ToString(i.FileSystemId) == aws.ToString(j.FileSystemId) {
			return fmt.Errorf("FSx for OpenZFS File System (%s) not recreated", aws.ToString(i.FileSystemId))
		}

		return nil
	}
}

func testAccOpenZFSFileSystemConfig_baseSingleAZ(rName string) string {
	return acctest.ConfigVPCWithSubnets(rName, 1)
}

func testAccOpenZFSFileSystemConfig_baseMultiAZ(rName string) string {
	return acctest.ConfigVPCWithSubnets(rName, 2)
}

func testAccOpenZFSFileSystemConfig_basic(rName string) string {
	return acctest.ConfigCompose(testAccOpenZFSFileSystemConfig_baseSingleAZ(rName), `
resource "aws_fsx_openzfs_file_system" "test" {
  skip_final_backup   = true
  storage_capacity    = 64
  subnet_ids          = aws_subnet.test[*].id
  deployment_type     = "SINGLE_AZ_1"
  throughput_capacity = 64
}
`)
}

func testAccOpenZFSFileSystemConfig_diskIOPSConfiguration(rName string, iops int) string {
	return acctest.ConfigCompose(testAccOpenZFSFileSystemConfig_baseSingleAZ(rName), fmt.Sprintf(`
resource "aws_fsx_openzfs_file_system" "test" {
  skip_final_backup   = true
  storage_capacity    = 64
  storage_type        = "SSD"
  subnet_ids          = aws_subnet.test[*].id
  deployment_type     = "SINGLE_AZ_1"
  throughput_capacity = 64
  disk_iops_configuration {
    mode = "USER_PROVISIONED"
    iops = %[2]d
  }

  tags = {
    Name = %[1]q
  }
}
`, rName, iops))
}

func testAccOpenZFSFileSystemConfig_securityGroupIDs1(rName string) string {
	return acctest.ConfigCompose(testAccOpenZFSFileSystemConfig_baseSingleAZ(rName), fmt.Sprintf(`
resource "aws_security_group" "test1" {
  name   = "%[1]s-1"
  vpc_id = aws_vpc.test.id

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
  skip_final_backup   = true
  storage_capacity    = 64
  subnet_ids          = aws_subnet.test[*].id
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
	return acctest.ConfigCompose(testAccOpenZFSFileSystemConfig_baseSingleAZ(rName), fmt.Sprintf(`
resource "aws_security_group" "test1" {
  name   = "%[1]s-1"
  vpc_id = aws_vpc.test.id

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
  name   = "%[1]s-2"
  vpc_id = aws_vpc.test.id

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
  skip_final_backup   = true
  storage_capacity    = 64
  subnet_ids          = aws_subnet.test[*].id
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
	return acctest.ConfigCompose(testAccOpenZFSFileSystemConfig_baseSingleAZ(rName), fmt.Sprintf(`
resource "aws_fsx_openzfs_file_system" "test" {
  skip_final_backup   = true
  storage_capacity    = 64
  subnet_ids          = aws_subnet.test[*].id
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
	return acctest.ConfigCompose(testAccOpenZFSFileSystemConfig_baseSingleAZ(rName), fmt.Sprintf(`
resource "aws_fsx_openzfs_file_system" "test" {
  skip_final_backup   = true
  storage_capacity    = 64
  subnet_ids          = aws_subnet.test[*].id
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
	return acctest.ConfigCompose(testAccOpenZFSFileSystemConfig_baseSingleAZ(rName), fmt.Sprintf(`
resource "aws_fsx_openzfs_file_system" "test" {
  skip_final_backup    = true
  storage_capacity     = 64
  subnet_ids           = aws_subnet.test[*].id
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
	return acctest.ConfigCompose(testAccOpenZFSFileSystemConfig_baseSingleAZ(rName), fmt.Sprintf(`
resource "aws_fsx_openzfs_file_system" "test" {
  skip_final_backup             = true
  storage_capacity              = 64
  subnet_ids                    = aws_subnet.test[*].id
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
	return acctest.ConfigCompose(testAccOpenZFSFileSystemConfig_baseSingleAZ(rName), fmt.Sprintf(`
resource "aws_fsx_openzfs_file_system" "test" {
  skip_final_backup                 = true
  storage_capacity                  = 64
  subnet_ids                        = aws_subnet.test[*].id
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
	return acctest.ConfigCompose(testAccOpenZFSFileSystemConfig_baseSingleAZ(rName), fmt.Sprintf(`
resource "aws_fsx_openzfs_file_system" "test" {
  skip_final_backup               = true
  storage_capacity                = 64
  subnet_ids                      = aws_subnet.test[*].id
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
	return acctest.ConfigCompose(testAccOpenZFSFileSystemConfig_baseSingleAZ(rName), fmt.Sprintf(`
resource "aws_kms_key" "test" {
  description             = %[1]q
  deletion_window_in_days = 7
}

resource "aws_fsx_openzfs_file_system" "test" {
  skip_final_backup   = true
  storage_capacity    = 64
  subnet_ids          = aws_subnet.test[*].id
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
	return acctest.ConfigCompose(testAccOpenZFSFileSystemConfig_baseSingleAZ(rName), fmt.Sprintf(`
resource "aws_fsx_openzfs_file_system" "test" {
  skip_final_backup   = true
  storage_capacity    = 64
  subnet_ids          = aws_subnet.test[*].id
  deployment_type     = "SINGLE_AZ_1"
  throughput_capacity = %[2]d

  tags = {
    Name = %[1]q
  }
}
`, rName, throughput))
}

func testAccOpenZFSFileSystemConfig_storageType(rName, storageType string) string {
	return acctest.ConfigCompose(testAccOpenZFSFileSystemConfig_baseSingleAZ(rName), fmt.Sprintf(`
resource "aws_fsx_openzfs_file_system" "test" {
  skip_final_backup   = true
  storage_capacity    = 64
  subnet_ids          = aws_subnet.test[*].id
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
	return acctest.ConfigCompose(testAccOpenZFSFileSystemConfig_baseSingleAZ(rName), fmt.Sprintf(`
resource "aws_fsx_openzfs_file_system" "test" {
  skip_final_backup   = true
  storage_capacity    = 64
  subnet_ids          = aws_subnet.test[*].id
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
    user_and_group_quotas {
      id                         = 0
      storage_capacity_quota_gib = 0
      type                       = "USER"
    }
    user_and_group_quotas {
      id                         = 0
      storage_capacity_quota_gib = 0
      type                       = "GROUP"
    }
  }

  tags = {
    Name = %[1]q
  }
}
`, rName, dataCompression, readOnly, quotaSize))
}

func testAccOpenZFSFileSystemConfig_rootVolume2(rName, dataCompression, readOnly string, quotaSize, recordSizeKiB int) string {
	return acctest.ConfigCompose(testAccOpenZFSFileSystemConfig_baseSingleAZ(rName), fmt.Sprintf(`
resource "aws_fsx_openzfs_file_system" "test" {
  skip_final_backup   = true
  storage_capacity    = 64
  subnet_ids          = aws_subnet.test[*].id
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

    read_only       = %[3]s
    record_size_kib = %[5]d

    user_and_group_quotas {
      id                         = 10
      storage_capacity_quota_gib = %[4]d
      type                       = "USER"
    }
    user_and_group_quotas {
      id                         = 0
      storage_capacity_quota_gib = 0
      type                       = "USER"
    }
    user_and_group_quotas {
      id                         = 0
      storage_capacity_quota_gib = 0
      type                       = "GROUP"
    }
  }

  tags = {
    Name = %[1]q
  }
}
`, rName, dataCompression, readOnly, quotaSize, recordSizeKiB))
}

func testAccOpenZFSFileSystemConfig_rootVolume3Client(rName, dataCompression, readOnly string, userQuota, groupQuota, recordSizeKiB int) string {
	return acctest.ConfigCompose(testAccOpenZFSFileSystemConfig_baseSingleAZ(rName), fmt.Sprintf(`
resource "aws_fsx_openzfs_file_system" "test" {
  skip_final_backup   = true
  storage_capacity    = 64
  subnet_ids          = aws_subnet.test[*].id
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

    read_only       = %[3]s
    record_size_kib = %[6]d

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
    user_and_group_quotas {
      id                         = 0
      storage_capacity_quota_gib = 0
      type                       = "USER"
    }
    user_and_group_quotas {
      id                         = 0
      storage_capacity_quota_gib = 0
      type                       = "GROUP"
    }
  }

  tags = {
    Name = %[1]q
  }
}
`, rName, dataCompression, readOnly, userQuota, groupQuota, recordSizeKiB))
}

func testAccOpenZFSFileSystemConfig_rootVolume4(rName, dataCompression, readOnly string, userQuota, groupQuota int) string {
	return acctest.ConfigCompose(testAccOpenZFSFileSystemConfig_baseSingleAZ(rName), fmt.Sprintf(`
resource "aws_fsx_openzfs_file_system" "test" {
  skip_final_backup   = true
  storage_capacity    = 64
  subnet_ids          = aws_subnet.test[*].id
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
    user_and_group_quotas {
      id                         = 0
      storage_capacity_quota_gib = 0
      type                       = "USER"
    }
    user_and_group_quotas {
      id                         = 0
      storage_capacity_quota_gib = 0
      type                       = "GROUP"
    }
  }

  tags = {
    Name = %[1]q
  }
}
`, rName, dataCompression, readOnly, userQuota, groupQuota))
}

func testAccOpenZFSFileSystemConfig_throughputCapacity(rName string) string {
	return acctest.ConfigCompose(testAccOpenZFSFileSystemConfig_baseSingleAZ(rName), fmt.Sprintf(`
resource "aws_fsx_openzfs_file_system" "test" {
  skip_final_backup   = true
  storage_capacity    = 64
  subnet_ids          = aws_subnet.test[*].id
  deployment_type     = "SINGLE_AZ_1"
  throughput_capacity = 128

  tags = {
    Name = %[1]q
  }
}
`, rName))
}

func testAccOpenZFSFileSystemConfig_storageCapacity(rName string) string {
	return acctest.ConfigCompose(testAccOpenZFSFileSystemConfig_baseSingleAZ(rName), fmt.Sprintf(`
resource "aws_fsx_openzfs_file_system" "test" {
  skip_final_backup   = true
  storage_capacity    = 75
  subnet_ids          = aws_subnet.test[*].id
  deployment_type     = "SINGLE_AZ_1"
  throughput_capacity = 64

  tags = {
    Name = %[1]q
  }
}
`, rName))
}

func testAccOpenZFSFileSystemConfig_deploymentType(rName, deploymentType string, throughput int) string {
	return acctest.ConfigCompose(testAccOpenZFSFileSystemConfig_baseSingleAZ(rName), fmt.Sprintf(`
resource "aws_fsx_openzfs_file_system" "test" {
  skip_final_backup   = true
  storage_capacity    = 64
  subnet_ids          = aws_subnet.test[*].id
  deployment_type     = %[2]q
  throughput_capacity = %[3]d

  tags = {
    Name = %[1]q
  }
}
`, rName, deploymentType, throughput))
}

func testAccOpenZFSFileSystemConfig_multiAZ(rName string) string {
	return acctest.ConfigCompose(testAccOpenZFSFileSystemConfig_baseMultiAZ(rName), fmt.Sprintf(`
resource "aws_fsx_openzfs_file_system" "test" {
  skip_final_backup   = true
  storage_capacity    = 64
  subnet_ids          = aws_subnet.test[*].id
  preferred_subnet_id = aws_subnet.test[0].id
  deployment_type     = "MULTI_AZ_1"
  throughput_capacity = 160

  tags = {
    Name = %[1]q
  }
}
`, rName))
}

func testAccOpenZFSFileSystemConfig_routeTableIDs(rName string, n int) string {
	return acctest.ConfigCompose(testAccOpenZFSFileSystemConfig_baseMultiAZ(rName), fmt.Sprintf(`
resource "aws_internet_gateway" "test" {
  vpc_id = aws_vpc.test.id

  tags = {
    Name = %[1]q
  }
}

resource "aws_route_table" "test" {
  count = %[2]d

  vpc_id = aws_vpc.test.id

  route {
    cidr_block = "0.0.0.0/0"
    gateway_id = aws_internet_gateway.test.id
  }

  tags = {
    Name = %[1]q
  }
}

resource "aws_fsx_openzfs_file_system" "test" {
  skip_final_backup   = true
  storage_capacity    = 64
  subnet_ids          = aws_subnet.test[*].id
  preferred_subnet_id = aws_subnet.test[0].id
  deployment_type     = "MULTI_AZ_1"
  throughput_capacity = 160
  route_table_ids     = aws_route_table.test[*].id

  tags = {
    Name = %[1]q
  }
}
`, rName, n))
}

func testAccOpenZFSFileSystemConfig_deleteConfig(rName, finalTagKey1, finalTagValue1, finalTagKey2, finalTagValue2 string) string {
	return acctest.ConfigCompose(testAccOpenZFSFileSystemConfig_baseSingleAZ(rName), fmt.Sprintf(`
resource "aws_fsx_openzfs_file_system" "test" {
  skip_final_backup   = false
  storage_capacity    = 64
  subnet_ids          = aws_subnet.test[*].id
  deployment_type     = "SINGLE_AZ_1"
  throughput_capacity = 64
  delete_options      = ["DELETE_CHILD_VOLUMES_AND_SNAPSHOTS"]

  final_backup_tags = {
    %[1]q = %[2]q
    %[3]q = %[4]q
  }
}
`, finalTagKey1, finalTagValue1, finalTagKey2, finalTagValue2))
}
