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

func TestAccFSxWindowsFileSystem_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var filesystem awstypes.FileSystem
	resourceName := "aws_fsx_windows_file_system.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	domainName := acctest.RandomDomainName()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckPartitionHasService(t, names.FSxEndpointID) },
		ErrorCheck:               acctest.ErrorCheck(t, names.FSxServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckWindowsFileSystemDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccWindowsFileSystemConfig_basic(rName, domainName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckWindowsFileSystemExists(ctx, resourceName, &filesystem),
					acctest.MatchResourceAttrRegionalARN(resourceName, names.AttrARN, "fsx", regexache.MustCompile(`file-system/fs-.+`)),
					resource.TestCheckResourceAttr(resourceName, "aliases.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "audit_log_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "audit_log_configuration.0.file_access_audit_log_level", "DISABLED"),
					resource.TestCheckResourceAttr(resourceName, "audit_log_configuration.0.file_share_access_audit_log_level", "DISABLED"),
					resource.TestCheckResourceAttr(resourceName, "automatic_backup_retention_days", "7"),
					resource.TestCheckResourceAttr(resourceName, "copy_tags_to_backups", acctest.CtFalse),
					resource.TestMatchResourceAttr(resourceName, "daily_automatic_backup_start_time", regexache.MustCompile(`^\d\d:\d\d$`)),
					resource.TestCheckResourceAttr(resourceName, "deployment_type", "SINGLE_AZ_1"),
					resource.TestCheckResourceAttr(resourceName, "disk_iops_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "disk_iops_configuration.0.iops", "96"),
					resource.TestCheckResourceAttr(resourceName, "disk_iops_configuration.0.mode", "AUTOMATIC"),
					resource.TestMatchResourceAttr(resourceName, names.AttrDNSName, regexache.MustCompile(`fs-.+\..+`)),
					acctest.MatchResourceAttrRegionalARN(resourceName, names.AttrKMSKeyID, "kms", regexache.MustCompile(`key/.+`)),
					resource.TestCheckResourceAttr(resourceName, "network_interface_ids.#", acctest.Ct1),
					acctest.CheckResourceAttrAccountID(resourceName, names.AttrOwnerID),
					resource.TestCheckResourceAttr(resourceName, "security_group_ids.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "self_managed_active_directory.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "skip_final_backup", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "storage_capacity", "32"),
					resource.TestCheckResourceAttr(resourceName, names.AttrStorageType, "SSD"),
					resource.TestCheckResourceAttr(resourceName, "subnet_ids.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "throughput_capacity", "8"),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrVPCID, "aws_vpc.test", names.AttrID),
					resource.TestMatchResourceAttr(resourceName, "weekly_maintenance_start_time", regexache.MustCompile(`^\d:\d\d:\d\d$`)),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"final_backup_tags",
					names.AttrSecurityGroupIDs,
					"skip_final_backup",
				},
			},
		},
	})
}

func TestAccFSxWindowsFileSystem_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	var filesystem awstypes.FileSystem
	resourceName := "aws_fsx_windows_file_system.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	domainName := acctest.RandomDomainName()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckPartitionHasService(t, names.FSxEndpointID) },
		ErrorCheck:               acctest.ErrorCheck(t, names.FSxServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckWindowsFileSystemDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccWindowsFileSystemConfig_basic(rName, domainName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckWindowsFileSystemExists(ctx, resourceName, &filesystem),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tffsx.ResourceWindowsFileSystem(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccFSxWindowsFileSystem_singleAz2(t *testing.T) {
	ctx := acctest.Context(t)
	var filesystem awstypes.FileSystem
	resourceName := "aws_fsx_windows_file_system.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	domainName := acctest.RandomDomainName()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckPartitionHasService(t, names.FSxEndpointID) },
		ErrorCheck:               acctest.ErrorCheck(t, names.FSxServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckWindowsFileSystemDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccWindowsFileSystemConfig_subnetIDs1SingleType(rName, domainName, "SINGLE_AZ_2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckWindowsFileSystemExists(ctx, resourceName, &filesystem),
					acctest.MatchResourceAttrRegionalARN(resourceName, names.AttrARN, "fsx", regexache.MustCompile(`file-system/fs-.+`)),
					resource.TestCheckResourceAttr(resourceName, "aliases.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "automatic_backup_retention_days", "7"),
					resource.TestCheckResourceAttr(resourceName, "copy_tags_to_backups", acctest.CtFalse),
					resource.TestMatchResourceAttr(resourceName, "daily_automatic_backup_start_time", regexache.MustCompile(`^\d\d:\d\d$`)),
					resource.TestCheckResourceAttr(resourceName, "deployment_type", "SINGLE_AZ_2"),
					resource.TestMatchResourceAttr(resourceName, names.AttrDNSName, regexache.MustCompile(`^amznfsx\w{8}\.\w{8}\.test$`)),
					acctest.MatchResourceAttrRegionalARN(resourceName, names.AttrKMSKeyID, "kms", regexache.MustCompile(`key/.+`)),
					resource.TestCheckResourceAttr(resourceName, "network_interface_ids.#", acctest.Ct1),
					acctest.CheckResourceAttrAccountID(resourceName, names.AttrOwnerID),
					resource.TestCheckResourceAttr(resourceName, "security_group_ids.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "self_managed_active_directory.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "skip_final_backup", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "storage_capacity", "32"),
					resource.TestCheckResourceAttr(resourceName, names.AttrStorageType, "SSD"),
					resource.TestCheckResourceAttr(resourceName, "subnet_ids.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "throughput_capacity", "8"),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrVPCID, "aws_vpc.test", names.AttrID),
					resource.TestMatchResourceAttr(resourceName, "weekly_maintenance_start_time", regexache.MustCompile(`^\d:\d\d:\d\d$`)),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"final_backup_tags",
					names.AttrSecurityGroupIDs,
					"skip_final_backup",
				},
			},
		},
	})
}

func TestAccFSxWindowsFileSystem_storageTypeHdd(t *testing.T) {
	ctx := acctest.Context(t)
	var filesystem awstypes.FileSystem
	resourceName := "aws_fsx_windows_file_system.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	domainName := acctest.RandomDomainName()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckPartitionHasService(t, names.FSxEndpointID) },
		ErrorCheck:               acctest.ErrorCheck(t, names.FSxServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckWindowsFileSystemDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccWindowsFileSystemConfig_subnetIDs1StorageType(rName, domainName, "SINGLE_AZ_2", "HDD"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckWindowsFileSystemExists(ctx, resourceName, &filesystem),
					resource.TestCheckResourceAttr(resourceName, "deployment_type", "SINGLE_AZ_2"),
					resource.TestCheckResourceAttr(resourceName, names.AttrStorageType, "HDD"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"final_backup_tags",
					names.AttrSecurityGroupIDs,
					"skip_final_backup",
				},
			},
		},
	})
}

func TestAccFSxWindowsFileSystem_multiAz(t *testing.T) {
	ctx := acctest.Context(t)
	var filesystem awstypes.FileSystem
	resourceName := "aws_fsx_windows_file_system.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	domainName := acctest.RandomDomainName()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckPartitionHasService(t, names.FSxEndpointID) },
		ErrorCheck:               acctest.ErrorCheck(t, names.FSxServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckWindowsFileSystemDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccWindowsFileSystemConfig_subnetIDs2(rName, domainName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckWindowsFileSystemExists(ctx, resourceName, &filesystem),
					acctest.MatchResourceAttrRegionalARN(resourceName, names.AttrARN, "fsx", regexache.MustCompile(`file-system/fs-.+`)),
					resource.TestCheckResourceAttr(resourceName, "aliases.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "automatic_backup_retention_days", "7"),
					resource.TestCheckResourceAttr(resourceName, "copy_tags_to_backups", acctest.CtFalse),
					resource.TestMatchResourceAttr(resourceName, "daily_automatic_backup_start_time", regexache.MustCompile(`^\d\d:\d\d$`)),
					resource.TestCheckResourceAttr(resourceName, "deployment_type", "MULTI_AZ_1"),
					resource.TestMatchResourceAttr(resourceName, names.AttrDNSName, regexache.MustCompile(`^amznfsx\w{8}\.\w{8}\.test$`)),
					acctest.MatchResourceAttrRegionalARN(resourceName, names.AttrKMSKeyID, "kms", regexache.MustCompile(`key/.+`)),
					resource.TestCheckResourceAttr(resourceName, "network_interface_ids.#", acctest.Ct2),
					acctest.CheckResourceAttrAccountID(resourceName, names.AttrOwnerID),
					resource.TestCheckResourceAttr(resourceName, "self_managed_active_directory.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "skip_final_backup", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "storage_capacity", "32"),
					resource.TestCheckResourceAttr(resourceName, names.AttrStorageType, "SSD"),
					resource.TestCheckResourceAttr(resourceName, "subnet_ids.#", acctest.Ct2),
					resource.TestCheckResourceAttr(resourceName, "throughput_capacity", "8"),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrVPCID, "aws_vpc.test", names.AttrID),
					resource.TestMatchResourceAttr(resourceName, "weekly_maintenance_start_time", regexache.MustCompile(`^\d:\d\d:\d\d$`)),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"final_backup_tags",
					names.AttrSecurityGroupIDs,
					"skip_final_backup",
				},
			},
		},
	})
}

func TestAccFSxWindowsFileSystem_aliases(t *testing.T) {
	ctx := acctest.Context(t)
	var filesystem1, filesystem2, filesystem3 awstypes.FileSystem
	resourceName := "aws_fsx_windows_file_system.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	domainName := acctest.RandomDomainName()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckPartitionHasService(t, names.FSxEndpointID) },
		ErrorCheck:               acctest.ErrorCheck(t, names.FSxServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckWindowsFileSystemDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccWindowsFileSystemConfig_aliases1(rName, domainName, "filesystem1.example.com"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckWindowsFileSystemExists(ctx, resourceName, &filesystem1),
					resource.TestCheckResourceAttr(resourceName, "aliases.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "aliases.0", "filesystem1.example.com"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"final_backup_tags",
					names.AttrSecurityGroupIDs,
					"skip_final_backup",
				},
			},
			{
				Config: testAccWindowsFileSystemConfig_aliases2(rName, domainName, "filesystem2.example.com", "filesystem3.example.com"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckWindowsFileSystemExists(ctx, resourceName, &filesystem2),
					testAccCheckWindowsFileSystemNotRecreated(&filesystem1, &filesystem2),
					resource.TestCheckResourceAttr(resourceName, "aliases.#", acctest.Ct2),
					resource.TestCheckResourceAttr(resourceName, "aliases.0", "filesystem2.example.com"),
					resource.TestCheckResourceAttr(resourceName, "aliases.1", "filesystem3.example.com"),
				),
			},
			{
				Config: testAccWindowsFileSystemConfig_aliases1(rName, domainName, "filesystem3.example.com"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckWindowsFileSystemExists(ctx, resourceName, &filesystem3),
					testAccCheckWindowsFileSystemNotRecreated(&filesystem2, &filesystem3),
					resource.TestCheckResourceAttr(resourceName, "aliases.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "aliases.0", "filesystem3.example.com"),
				),
			},
		},
	})
}

func TestAccFSxWindowsFileSystem_automaticBackupRetentionDays(t *testing.T) {
	ctx := acctest.Context(t)
	var filesystem1, filesystem2, filesystem3 awstypes.FileSystem
	resourceName := "aws_fsx_windows_file_system.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	domainName := acctest.RandomDomainName()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckPartitionHasService(t, names.FSxEndpointID) },
		ErrorCheck:               acctest.ErrorCheck(t, names.FSxServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckWindowsFileSystemDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccWindowsFileSystemConfig_automaticBackupRetentionDays(rName, domainName, 90),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckWindowsFileSystemExists(ctx, resourceName, &filesystem1),
					resource.TestCheckResourceAttr(resourceName, "automatic_backup_retention_days", "90"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"final_backup_tags",
					names.AttrSecurityGroupIDs,
					"skip_final_backup",
				},
			},
			{
				Config: testAccWindowsFileSystemConfig_automaticBackupRetentionDays(rName, domainName, 0),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckWindowsFileSystemExists(ctx, resourceName, &filesystem2),
					testAccCheckWindowsFileSystemNotRecreated(&filesystem1, &filesystem2),
					resource.TestCheckResourceAttr(resourceName, "automatic_backup_retention_days", acctest.Ct0),
				),
			},
			{
				Config: testAccWindowsFileSystemConfig_automaticBackupRetentionDays(rName, domainName, 14),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckWindowsFileSystemExists(ctx, resourceName, &filesystem3),
					testAccCheckWindowsFileSystemNotRecreated(&filesystem2, &filesystem3),
					resource.TestCheckResourceAttr(resourceName, "automatic_backup_retention_days", "14"),
				),
			},
		},
	})
}

func TestAccFSxWindowsFileSystem_copyTagsToBackups(t *testing.T) {
	ctx := acctest.Context(t)
	var filesystem1, filesystem2 awstypes.FileSystem
	resourceName := "aws_fsx_windows_file_system.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	domainName := acctest.RandomDomainName()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckPartitionHasService(t, names.FSxEndpointID) },
		ErrorCheck:               acctest.ErrorCheck(t, names.FSxServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckWindowsFileSystemDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccWindowsFileSystemConfig_copyTagsToBackups(rName, domainName, true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckWindowsFileSystemExists(ctx, resourceName, &filesystem1),
					resource.TestCheckResourceAttr(resourceName, "copy_tags_to_backups", acctest.CtTrue),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"final_backup_tags",
					names.AttrSecurityGroupIDs,
					"skip_final_backup",
				},
			},
			{
				Config: testAccWindowsFileSystemConfig_copyTagsToBackups(rName, domainName, false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckWindowsFileSystemExists(ctx, resourceName, &filesystem2),
					testAccCheckWindowsFileSystemRecreated(&filesystem1, &filesystem2),
					resource.TestCheckResourceAttr(resourceName, "copy_tags_to_backups", acctest.CtFalse),
				),
			},
		},
	})
}

func TestAccFSxWindowsFileSystem_dailyAutomaticBackupStartTime(t *testing.T) {
	ctx := acctest.Context(t)
	var filesystem1, filesystem2 awstypes.FileSystem
	resourceName := "aws_fsx_windows_file_system.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	domainName := acctest.RandomDomainName()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckPartitionHasService(t, names.FSxEndpointID) },
		ErrorCheck:               acctest.ErrorCheck(t, names.FSxServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckWindowsFileSystemDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccWindowsFileSystemConfig_dailyAutomaticBackupStartTime(rName, domainName, "01:01"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckWindowsFileSystemExists(ctx, resourceName, &filesystem1),
					resource.TestCheckResourceAttr(resourceName, "daily_automatic_backup_start_time", "01:01"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"final_backup_tags",
					names.AttrSecurityGroupIDs,
					"skip_final_backup",
				},
			},
			{
				Config: testAccWindowsFileSystemConfig_dailyAutomaticBackupStartTime(rName, domainName, "02:02"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckWindowsFileSystemExists(ctx, resourceName, &filesystem2),
					testAccCheckWindowsFileSystemNotRecreated(&filesystem1, &filesystem2),
					resource.TestCheckResourceAttr(resourceName, "daily_automatic_backup_start_time", "02:02"),
				),
			},
		},
	})
}

func TestAccFSxWindowsFileSystem_deleteConfig(t *testing.T) {
	ctx := acctest.Context(t)
	var filesystem awstypes.FileSystem
	resourceName := "aws_fsx_windows_file_system.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	domainName := acctest.RandomDomainName()

	acctest.SkipIfEnvVarNotSet(t, "AWS_FSX_CREATE_FINAL_BACKUP")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckPartitionHasService(t, names.FSxEndpointID) },
		ErrorCheck:               acctest.ErrorCheck(t, names.FSxServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckWindowsFileSystemDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccWindowsFileSystemConfig_deleteConfig(rName, domainName, acctest.CtKey1, acctest.CtValue1, acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckWindowsFileSystemExists(ctx, resourceName, &filesystem),
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
					"final_backup_tags",
					names.AttrSecurityGroupIDs,
					"skip_final_backup",
				},
			},
		},
	})
}

func TestAccFSxWindowsFileSystem_kmsKeyID(t *testing.T) {
	ctx := acctest.Context(t)
	var filesystem1, filesystem2 awstypes.FileSystem
	kmsKeyResourceName1 := "aws_kms_key.test1"
	kmsKeyResourceName2 := "aws_kms_key.test2"
	resourceName := "aws_fsx_windows_file_system.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	domainName := acctest.RandomDomainName()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckPartitionHasService(t, names.FSxEndpointID) },
		ErrorCheck:               acctest.ErrorCheck(t, names.FSxServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckWindowsFileSystemDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccWindowsFileSystemConfig_kmsKeyID1(rName, domainName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckWindowsFileSystemExists(ctx, resourceName, &filesystem1),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrKMSKeyID, kmsKeyResourceName1, names.AttrARN),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"final_backup_tags",
					names.AttrSecurityGroupIDs,
					"skip_final_backup",
				},
			},
			{
				Config: testAccWindowsFileSystemConfig_kmsKeyID2(rName, domainName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckWindowsFileSystemExists(ctx, resourceName, &filesystem2),
					testAccCheckWindowsFileSystemRecreated(&filesystem1, &filesystem2),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrKMSKeyID, kmsKeyResourceName2, names.AttrARN),
				),
			},
		},
	})
}

func TestAccFSxWindowsFileSystem_securityGroupIDs(t *testing.T) {
	ctx := acctest.Context(t)
	var filesystem1, filesystem2 awstypes.FileSystem
	resourceName := "aws_fsx_windows_file_system.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	domainName := acctest.RandomDomainName()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckPartitionHasService(t, names.FSxEndpointID) },
		ErrorCheck:               acctest.ErrorCheck(t, names.FSxServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckWindowsFileSystemDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccWindowsFileSystemConfig_securityGroupIDs1(rName, domainName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckWindowsFileSystemExists(ctx, resourceName, &filesystem1),
					resource.TestCheckResourceAttr(resourceName, "security_group_ids.#", acctest.Ct1),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"final_backup_tags",
					names.AttrSecurityGroupIDs,
					"skip_final_backup",
				},
			},
			{
				Config: testAccWindowsFileSystemConfig_securityGroupIDs2(rName, domainName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckWindowsFileSystemExists(ctx, resourceName, &filesystem2),
					testAccCheckWindowsFileSystemRecreated(&filesystem1, &filesystem2),
					resource.TestCheckResourceAttr(resourceName, "security_group_ids.#", acctest.Ct2),
				),
			},
		},
	})
}

func TestAccFSxWindowsFileSystem_selfManagedActiveDirectory(t *testing.T) {
	ctx := acctest.Context(t)
	var filesystem awstypes.FileSystem
	resourceName := "aws_fsx_windows_file_system.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	domainName := acctest.RandomDomainName()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckPartitionHasService(t, names.FSxEndpointID) },
		ErrorCheck:               acctest.ErrorCheck(t, names.FSxServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckWindowsFileSystemDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccWindowsFileSystemConfig_selfManagedActiveDirectory(rName, domainName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckWindowsFileSystemExists(ctx, resourceName, &filesystem),
					resource.TestCheckResourceAttr(resourceName, "self_managed_active_directory.#", acctest.Ct1),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"final_backup_tags",
					"self_managed_active_directory",
					names.AttrSecurityGroupIDs,
					"skip_final_backup",
				},
			},
		},
	})
}

func TestAccFSxWindowsFileSystem_storageCapacity(t *testing.T) {
	ctx := acctest.Context(t)
	var filesystem1, filesystem2 awstypes.FileSystem
	resourceName := "aws_fsx_windows_file_system.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	domainName := acctest.RandomDomainName()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckPartitionHasService(t, names.FSxEndpointID) },
		ErrorCheck:               acctest.ErrorCheck(t, names.FSxServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckWindowsFileSystemDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccWindowsFileSystemConfig_storageCapacity(rName, domainName, 32, 8),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckWindowsFileSystemExists(ctx, resourceName, &filesystem1),
					resource.TestCheckResourceAttr(resourceName, "storage_capacity", "32"),
					resource.TestCheckResourceAttr(resourceName, "throughput_capacity", "8"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"final_backup_tags",
					names.AttrSecurityGroupIDs,
					"skip_final_backup",
				},
			},
			{
				Config: testAccWindowsFileSystemConfig_storageCapacity(rName, domainName, 64, 16),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckWindowsFileSystemExists(ctx, resourceName, &filesystem2),
					testAccCheckWindowsFileSystemNotRecreated(&filesystem1, &filesystem2),
					resource.TestCheckResourceAttr(resourceName, "storage_capacity", "64"),
					resource.TestCheckResourceAttr(resourceName, "throughput_capacity", "16"),
				),
			},
		},
	})
}

func TestAccFSxWindowsFileSystem_fromBackup(t *testing.T) {
	ctx := acctest.Context(t)
	var filesystem awstypes.FileSystem
	resourceName := "aws_fsx_windows_file_system.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	domainName := acctest.RandomDomainName()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckPartitionHasService(t, names.FSxEndpointID) },
		ErrorCheck:               acctest.ErrorCheck(t, names.FSxServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckWindowsFileSystemDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccWindowsFileSystemConfig_fromBackup(rName, domainName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckWindowsFileSystemExists(ctx, resourceName, &filesystem),
					resource.TestCheckResourceAttrPair(resourceName, "backup_id", "aws_fsx_backup.test", names.AttrID),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"backup_id",
					"final_backup_tags",
					names.AttrSecurityGroupIDs,
					"skip_final_backup",
				},
			},
		},
	})
}

func TestAccFSxWindowsFileSystem_tags(t *testing.T) {
	ctx := acctest.Context(t)
	var filesystem1, filesystem2, filesystem3 awstypes.FileSystem
	resourceName := "aws_fsx_windows_file_system.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	domainName := acctest.RandomDomainName()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckPartitionHasService(t, names.FSxEndpointID) },
		ErrorCheck:               acctest.ErrorCheck(t, names.FSxServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckWindowsFileSystemDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccWindowsFileSystemConfig_tags1(rName, domainName, acctest.CtKey1, acctest.CtValue1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckWindowsFileSystemExists(ctx, resourceName, &filesystem1),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey1, acctest.CtValue1),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"final_backup_tags",
					names.AttrSecurityGroupIDs,
					"skip_final_backup",
				},
			},
			{
				Config: testAccWindowsFileSystemConfig_tags2(rName, domainName, acctest.CtKey1, acctest.CtValue1Updated, acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckWindowsFileSystemExists(ctx, resourceName, &filesystem2),
					testAccCheckWindowsFileSystemNotRecreated(&filesystem1, &filesystem2),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct2),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey1, acctest.CtValue1Updated),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey2, acctest.CtValue2),
				),
			},
			{
				Config: testAccWindowsFileSystemConfig_tags1(rName, domainName, acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckWindowsFileSystemExists(ctx, resourceName, &filesystem3),
					testAccCheckWindowsFileSystemNotRecreated(&filesystem2, &filesystem3),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey2, acctest.CtValue2),
				),
			},
		},
	})
}

func TestAccFSxWindowsFileSystem_throughputCapacity(t *testing.T) {
	ctx := acctest.Context(t)
	var filesystem1, filesystem2 awstypes.FileSystem
	resourceName := "aws_fsx_windows_file_system.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	domainName := acctest.RandomDomainName()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckPartitionHasService(t, names.FSxEndpointID) },
		ErrorCheck:               acctest.ErrorCheck(t, names.FSxServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckWindowsFileSystemDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccWindowsFileSystemConfig_throughputCapacity(rName, domainName, 16),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckWindowsFileSystemExists(ctx, resourceName, &filesystem1),
					resource.TestCheckResourceAttr(resourceName, "throughput_capacity", "16"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"final_backup_tags",
					names.AttrSecurityGroupIDs,
					"skip_final_backup",
				},
			},
			{
				Config: testAccWindowsFileSystemConfig_throughputCapacity(rName, domainName, 32),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckWindowsFileSystemExists(ctx, resourceName, &filesystem2),
					testAccCheckWindowsFileSystemNotRecreated(&filesystem1, &filesystem2),
					resource.TestCheckResourceAttr(resourceName, "throughput_capacity", "32"),
				),
			},
		},
	})
}

func TestAccFSxWindowsFileSystem_weeklyMaintenanceStartTime(t *testing.T) {
	ctx := acctest.Context(t)
	var filesystem1, filesystem2 awstypes.FileSystem
	resourceName := "aws_fsx_windows_file_system.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	domainName := acctest.RandomDomainName()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckPartitionHasService(t, names.FSxEndpointID) },
		ErrorCheck:               acctest.ErrorCheck(t, names.FSxServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckWindowsFileSystemDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccWindowsFileSystemConfig_weeklyMaintenanceStartTime(rName, domainName, "1:01:01"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckWindowsFileSystemExists(ctx, resourceName, &filesystem1),
					resource.TestCheckResourceAttr(resourceName, "weekly_maintenance_start_time", "1:01:01"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"final_backup_tags",
					names.AttrSecurityGroupIDs,
					"skip_final_backup",
				},
			},
			{
				Config: testAccWindowsFileSystemConfig_weeklyMaintenanceStartTime(rName, domainName, "2:02:02"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckWindowsFileSystemExists(ctx, resourceName, &filesystem2),
					testAccCheckWindowsFileSystemNotRecreated(&filesystem1, &filesystem2),
					resource.TestCheckResourceAttr(resourceName, "weekly_maintenance_start_time", "2:02:02"),
				),
			},
		},
	})
}

func TestAccFSxWindowsFileSystem_audit(t *testing.T) {
	ctx := acctest.Context(t)
	var filesystem awstypes.FileSystem
	resourceName := "aws_fsx_windows_file_system.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	domainName := acctest.RandomDomainName()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckPartitionHasService(t, names.FSxEndpointID) },
		ErrorCheck:               acctest.ErrorCheck(t, names.FSxServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckWindowsFileSystemDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccWindowsFileSystemConfig_audit(rName, domainName, string(awstypes.WindowsAccessAuditLogLevelSuccessOnly)),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckWindowsFileSystemExists(ctx, resourceName, &filesystem),
					resource.TestCheckResourceAttr(resourceName, "audit_log_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "audit_log_configuration.0.file_access_audit_log_level", string(awstypes.WindowsAccessAuditLogLevelSuccessOnly)),
					resource.TestCheckResourceAttr(resourceName, "audit_log_configuration.0.file_share_access_audit_log_level", string(awstypes.WindowsAccessAuditLogLevelSuccessOnly)),
					resource.TestCheckResourceAttrSet(resourceName, "audit_log_configuration.0.audit_log_destination"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"final_backup_tags",
					names.AttrSecurityGroupIDs,
					"skip_final_backup",
				},
			},
			{
				Config: testAccWindowsFileSystemConfig_audit(rName, domainName, string(awstypes.WindowsAccessAuditLogLevelSuccessAndFailure)),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckWindowsFileSystemExists(ctx, resourceName, &filesystem),
					resource.TestCheckResourceAttr(resourceName, "audit_log_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "audit_log_configuration.0.file_access_audit_log_level", string(awstypes.WindowsAccessAuditLogLevelSuccessAndFailure)),
					resource.TestCheckResourceAttr(resourceName, "audit_log_configuration.0.file_share_access_audit_log_level", string(awstypes.WindowsAccessAuditLogLevelSuccessAndFailure)),
					resource.TestCheckResourceAttrSet(resourceName, "audit_log_configuration.0.audit_log_destination"),
				),
			},
			{
				Config: testAccWindowsFileSystemConfig_auditNoDestination(rName, domainName, string(awstypes.WindowsAccessAuditLogLevelDisabled)),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckWindowsFileSystemExists(ctx, resourceName, &filesystem),
					resource.TestCheckResourceAttr(resourceName, "audit_log_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "audit_log_configuration.0.file_access_audit_log_level", string(awstypes.WindowsAccessAuditLogLevelDisabled)),
					resource.TestCheckResourceAttr(resourceName, "audit_log_configuration.0.file_share_access_audit_log_level", string(awstypes.WindowsAccessAuditLogLevelDisabled)),
				),
			},
		},
	})
}

func TestAccFSxWindowsFileSystem_diskIops(t *testing.T) {
	ctx := acctest.Context(t)
	var filesystem1, filesystem2 awstypes.FileSystem
	resourceName := "aws_fsx_windows_file_system.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	domainName := acctest.RandomDomainName()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckPartitionHasService(t, names.FSxEndpointID) },
		ErrorCheck:               acctest.ErrorCheck(t, names.FSxServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckWindowsFileSystemDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccWindowsFileSystemConfig_diskIOPSConfiguration(rName, domainName, 192),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckWindowsFileSystemExists(ctx, resourceName, &filesystem1),
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
					"final_backup_tags",
					names.AttrSecurityGroupIDs,
					"skip_final_backup",
				},
			},
			{
				Config: testAccWindowsFileSystemConfig_diskIOPSConfiguration(rName, domainName, 256),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckWindowsFileSystemExists(ctx, resourceName, &filesystem2),
					testAccCheckWindowsFileSystemNotRecreated(&filesystem1, &filesystem2),
					resource.TestCheckResourceAttr(resourceName, "disk_iops_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "disk_iops_configuration.0.mode", "USER_PROVISIONED"),
					resource.TestCheckResourceAttr(resourceName, "disk_iops_configuration.0.iops", "256"),
				),
			},
		},
	})
}

func testAccCheckWindowsFileSystemExists(ctx context.Context, n string, v *awstypes.FileSystem) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).FSxClient(ctx)

		output, err := tffsx.FindWindowsFileSystemByID(ctx, conn, rs.Primary.ID)

		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

func testAccCheckWindowsFileSystemDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).FSxClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_fsx_windows_file_system" {
				continue
			}

			_, err := tffsx.FindWindowsFileSystemByID(ctx, conn, rs.Primary.ID)

			if tfresource.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("FSx for Windows File Server File System (%s) still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccCheckWindowsFileSystemNotRecreated(i, j *awstypes.FileSystem) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if aws.ToString(i.FileSystemId) != aws.ToString(j.FileSystemId) {
			return fmt.Errorf("FSx for Windows File Server File System (%s) recreated", aws.ToString(i.FileSystemId))
		}

		return nil
	}
}

func testAccCheckWindowsFileSystemRecreated(i, j *awstypes.FileSystem) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if aws.ToString(i.FileSystemId) == aws.ToString(j.FileSystemId) {
			return fmt.Errorf("FSx for Windows File Server File System (%s) not recreated", aws.ToString(i.FileSystemId))
		}

		return nil
	}
}

func testAccWindowsFileSystemConfig_base(rName, domain string) string {
	return acctest.ConfigCompose(acctest.ConfigVPCWithSubnets(rName, 2), fmt.Sprintf(`
resource "aws_directory_service_directory" "test" {
  edition  = "Standard"
  name     = %[1]q
  password = "SuperSecretPassw0rd"
  type     = "MicrosoftAD"

  vpc_settings {
    subnet_ids = aws_subnet.test[*].id
    vpc_id     = aws_vpc.test.id
  }
}
`, domain))
}

func testAccWindowsFileSystemConfig_basic(rName, domain string) string {
	return acctest.ConfigCompose(testAccWindowsFileSystemConfig_base(rName, domain), `
resource "aws_fsx_windows_file_system" "test" {
  active_directory_id = aws_directory_service_directory.test.id
  skip_final_backup   = true
  storage_capacity    = 32
  subnet_ids          = [aws_subnet.test[0].id]
  throughput_capacity = 8
}
`)
}

func testAccWindowsFileSystemConfig_aliases1(rName, domain, alias1 string) string {
	return acctest.ConfigCompose(testAccWindowsFileSystemConfig_base(rName, domain), fmt.Sprintf(`
resource "aws_fsx_windows_file_system" "test" {
  active_directory_id = aws_directory_service_directory.test.id
  skip_final_backup   = true
  storage_capacity    = 32
  subnet_ids          = [aws_subnet.test[0].id]
  throughput_capacity = 8

  aliases = [%[2]q]

  tags = {
    Name = %[1]q
  }
}
`, rName, alias1))
}

func testAccWindowsFileSystemConfig_aliases2(rName, domain, alias1, alias2 string) string {
	return acctest.ConfigCompose(testAccWindowsFileSystemConfig_base(rName, domain), fmt.Sprintf(`
resource "aws_fsx_windows_file_system" "test" {
  active_directory_id = aws_directory_service_directory.test.id
  skip_final_backup   = true
  storage_capacity    = 32
  subnet_ids          = [aws_subnet.test[0].id]
  throughput_capacity = 8

  aliases = [%[2]q, %[3]q]

  tags = {
    Name = %[1]q
  }
}
`, rName, alias1, alias2))
}

func testAccWindowsFileSystemConfig_automaticBackupRetentionDays(rName, domain string, automaticBackupRetentionDays int) string {
	return acctest.ConfigCompose(testAccWindowsFileSystemConfig_base(rName, domain), fmt.Sprintf(`
resource "aws_fsx_windows_file_system" "test" {
  active_directory_id             = aws_directory_service_directory.test.id
  automatic_backup_retention_days = %[2]d
  skip_final_backup               = true
  storage_capacity                = 32
  subnet_ids                      = [aws_subnet.test[0].id]
  throughput_capacity             = 8

  tags = {
    Name = %[1]q
  }
}
`, rName, automaticBackupRetentionDays))
}

func testAccWindowsFileSystemConfig_copyTagsToBackups(rName, domain string, copyTagsToBackups bool) string {
	return acctest.ConfigCompose(testAccWindowsFileSystemConfig_base(rName, domain), fmt.Sprintf(`
resource "aws_fsx_windows_file_system" "test" {
  active_directory_id  = aws_directory_service_directory.test.id
  copy_tags_to_backups = %[2]t
  skip_final_backup    = true
  storage_capacity     = 32
  subnet_ids           = [aws_subnet.test[0].id]
  throughput_capacity  = 8

  tags = {
    Name = %[1]q
  }
}
`, rName, copyTagsToBackups))
}

func testAccWindowsFileSystemConfig_dailyAutomaticBackupStartTime(rName, domain, dailyAutomaticBackupStartTime string) string {
	return acctest.ConfigCompose(testAccWindowsFileSystemConfig_base(rName, domain), fmt.Sprintf(`
resource "aws_fsx_windows_file_system" "test" {
  active_directory_id               = aws_directory_service_directory.test.id
  daily_automatic_backup_start_time = %[2]q
  skip_final_backup                 = true
  storage_capacity                  = 32
  subnet_ids                        = [aws_subnet.test[0].id]
  throughput_capacity               = 8

  tags = {
    Name = %[1]q
  }
}
`, rName, dailyAutomaticBackupStartTime))
}

func testAccWindowsFileSystemConfig_deleteConfig(rName, domain, finalTagKey1, finalTagValue1, finalTagKey2, finalTagValue2 string) string {
	return acctest.ConfigCompose(testAccWindowsFileSystemConfig_base(rName, domain), fmt.Sprintf(`
resource "aws_fsx_windows_file_system" "test" {
  active_directory_id = aws_directory_service_directory.test.id
  skip_final_backup   = false
  storage_capacity    = 32
  subnet_ids          = [aws_subnet.test[0].id]
  throughput_capacity = 8

  final_backup_tags = {
    %[1]q = %[2]q
    %[3]q = %[4]q
  }
}
`, finalTagKey1, finalTagValue1, finalTagKey2, finalTagValue2))
}

func testAccWindowsFileSystemConfig_kmsKeyID1(rName, domain string) string {
	return acctest.ConfigCompose(testAccWindowsFileSystemConfig_base(rName, domain), fmt.Sprintf(`
resource "aws_kms_key" "test1" {
  description             = "%[1]s-1"
  deletion_window_in_days = 7
}

resource "aws_fsx_windows_file_system" "test" {
  active_directory_id = aws_directory_service_directory.test.id
  kms_key_id          = aws_kms_key.test1.arn
  skip_final_backup   = true
  storage_capacity    = 32
  subnet_ids          = [aws_subnet.test[0].id]
  throughput_capacity = 8

  tags = {
    Name = %[1]q
  }
}
`, rName))
}

func testAccWindowsFileSystemConfig_kmsKeyID2(rName, domain string) string {
	return acctest.ConfigCompose(testAccWindowsFileSystemConfig_base(rName, domain), fmt.Sprintf(`
resource "aws_kms_key" "test2" {
  description             = "%[1]s-2"
  deletion_window_in_days = 7
}

resource "aws_fsx_windows_file_system" "test" {
  active_directory_id = aws_directory_service_directory.test.id
  kms_key_id          = aws_kms_key.test2.arn
  skip_final_backup   = true
  storage_capacity    = 32
  subnet_ids          = [aws_subnet.test[0].id]
  throughput_capacity = 8

  tags = {
    Name = %[1]q
  }
}
`, rName))
}

func testAccWindowsFileSystemConfig_securityGroupIDs1(rName, domain string) string {
	return acctest.ConfigCompose(testAccWindowsFileSystemConfig_base(rName, domain), fmt.Sprintf(`
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

resource "aws_fsx_windows_file_system" "test" {
  active_directory_id = aws_directory_service_directory.test.id
  security_group_ids  = [aws_security_group.test1.id]
  skip_final_backup   = true
  storage_capacity    = 32
  subnet_ids          = [aws_subnet.test[0].id]
  throughput_capacity = 8

  tags = {
    Name = %[1]q
  }
}
`, rName))
}

func testAccWindowsFileSystemConfig_securityGroupIDs2(rName, domain string) string {
	return acctest.ConfigCompose(testAccWindowsFileSystemConfig_base(rName, domain), fmt.Sprintf(`
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

resource "aws_fsx_windows_file_system" "test" {
  active_directory_id = aws_directory_service_directory.test.id
  security_group_ids  = [aws_security_group.test1.id, aws_security_group.test2.id]
  skip_final_backup   = true
  storage_capacity    = 32
  subnet_ids          = [aws_subnet.test[0].id]
  throughput_capacity = 8

  tags = {
    Name = %[1]q
  }
}
`, rName))
}

func testAccWindowsFileSystemConfig_selfManagedActiveDirectory(rName, domain string) string {
	return acctest.ConfigCompose(testAccWindowsFileSystemConfig_base(rName, domain), fmt.Sprintf(`
resource "aws_fsx_windows_file_system" "test" {
  skip_final_backup   = true
  storage_capacity    = 32
  subnet_ids          = [aws_subnet.test[0].id]
  throughput_capacity = 8

  self_managed_active_directory {
    dns_ips     = aws_directory_service_directory.test.dns_ip_addresses
    domain_name = aws_directory_service_directory.test.name
    password    = aws_directory_service_directory.test.password
    username    = "Admin"
  }

  tags = {
    Name = %[1]q
  }
}
`, rName))
}

func testAccWindowsFileSystemConfig_storageCapacity(rName, domain string, storageCapacity, throughputCapacity int) string {
	return acctest.ConfigCompose(testAccWindowsFileSystemConfig_base(rName, domain), fmt.Sprintf(`
resource "aws_fsx_windows_file_system" "test" {
  active_directory_id = aws_directory_service_directory.test.id
  skip_final_backup   = true
  storage_capacity    = %[2]d
  subnet_ids          = [aws_subnet.test[0].id]
  throughput_capacity = %[3]d

  tags = {
    Name = %[1]q
  }
}
`, rName, storageCapacity, throughputCapacity))
}

func testAccWindowsFileSystemConfig_subnetIDs1SingleType(rName, domain, azType string) string {
	return acctest.ConfigCompose(testAccWindowsFileSystemConfig_base(rName, domain), fmt.Sprintf(`
resource "aws_fsx_windows_file_system" "test" {
  active_directory_id = aws_directory_service_directory.test.id
  skip_final_backup   = true
  storage_capacity    = 32
  deployment_type     = %[2]q
  subnet_ids          = [aws_subnet.test[0].id]
  throughput_capacity = 8

  tags = {
    Name = %[1]q
  }
}
`, rName, azType))
}

func testAccWindowsFileSystemConfig_subnetIDs1StorageType(rName, domain, azType, storageType string) string {
	return acctest.ConfigCompose(testAccWindowsFileSystemConfig_base(rName, domain), fmt.Sprintf(`
resource "aws_fsx_windows_file_system" "test" {
  active_directory_id = aws_directory_service_directory.test.id
  skip_final_backup   = true
  storage_capacity    = 2000
  deployment_type     = %[2]q
  subnet_ids          = [aws_subnet.test[0].id]
  throughput_capacity = 8
  storage_type        = %[3]q

  tags = {
    Name = %[1]q
  }
}
`, rName, azType, storageType))
}

func testAccWindowsFileSystemConfig_subnetIDs2(rName, domain string) string {
	return acctest.ConfigCompose(testAccWindowsFileSystemConfig_base(rName, domain), fmt.Sprintf(`
resource "aws_fsx_windows_file_system" "test" {
  active_directory_id = aws_directory_service_directory.test.id
  skip_final_backup   = true
  storage_capacity    = 32
  deployment_type     = "MULTI_AZ_1"
  subnet_ids          = [aws_subnet.test[0].id, aws_subnet.test[1].id]
  preferred_subnet_id = aws_subnet.test[0].id
  throughput_capacity = 8

  tags = {
    Name = %[1]q
  }
}
`, rName))
}

func testAccWindowsFileSystemConfig_fromBackup(rName, domain string) string {
	return acctest.ConfigCompose(testAccWindowsFileSystemConfig_base(rName, domain), fmt.Sprintf(`
resource "aws_fsx_windows_file_system" "base" {
  active_directory_id = aws_directory_service_directory.test.id
  skip_final_backup   = true
  storage_capacity    = 32
  subnet_ids          = [aws_subnet.test[0].id]
  throughput_capacity = 8

  tags = {
    Name = %[1]q
  }
}

resource "aws_fsx_backup" "test" {
  file_system_id = aws_fsx_windows_file_system.base.id

  tags = {
    Name = %[1]q
  }
}

resource "aws_fsx_windows_file_system" "test" {
  active_directory_id = aws_directory_service_directory.test.id
  backup_id           = aws_fsx_backup.test.id
  skip_final_backup   = true
  subnet_ids          = [aws_subnet.test[0].id]
  throughput_capacity = 8

  tags = {
    Name = %[1]q
  }
}
`, rName))
}

func testAccWindowsFileSystemConfig_tags1(rName, domain, tagKey1, tagValue1 string) string {
	return acctest.ConfigCompose(testAccWindowsFileSystemConfig_base(rName, domain), fmt.Sprintf(`
resource "aws_fsx_windows_file_system" "test" {
  active_directory_id = aws_directory_service_directory.test.id
  skip_final_backup   = true
  storage_capacity    = 32
  subnet_ids          = [aws_subnet.test[0].id]
  throughput_capacity = 8

  tags = {
    %[1]q = %[2]q
  }
}
`, tagKey1, tagValue1))
}

func testAccWindowsFileSystemConfig_tags2(rName, domain, tagKey1, tagValue1, tagKey2, tagValue2 string) string {
	return acctest.ConfigCompose(testAccWindowsFileSystemConfig_base(rName, domain), fmt.Sprintf(`
resource "aws_fsx_windows_file_system" "test" {
  active_directory_id = aws_directory_service_directory.test.id
  skip_final_backup   = true
  storage_capacity    = 32
  subnet_ids          = [aws_subnet.test[0].id]
  throughput_capacity = 8

  tags = {
    %[1]q = %[2]q
    %[3]q = %[4]q
  }
}
`, tagKey1, tagValue1, tagKey2, tagValue2))
}

func testAccWindowsFileSystemConfig_throughputCapacity(rName, domain string, throughputCapacity int) string {
	return acctest.ConfigCompose(testAccWindowsFileSystemConfig_base(rName, domain), fmt.Sprintf(`
resource "aws_fsx_windows_file_system" "test" {
  active_directory_id = aws_directory_service_directory.test.id
  skip_final_backup   = true
  storage_capacity    = 32
  subnet_ids          = [aws_subnet.test[0].id]
  throughput_capacity = %[2]d

  tags = {
    Name = %[1]q
  }
}
`, rName, throughputCapacity))
}

func testAccWindowsFileSystemConfig_weeklyMaintenanceStartTime(rName, domain, weeklyMaintenanceStartTime string) string {
	return acctest.ConfigCompose(testAccWindowsFileSystemConfig_base(rName, domain), fmt.Sprintf(`
resource "aws_fsx_windows_file_system" "test" {
  active_directory_id           = aws_directory_service_directory.test.id
  skip_final_backup             = true
  storage_capacity              = 32
  subnet_ids                    = [aws_subnet.test[0].id]
  throughput_capacity           = 8
  weekly_maintenance_start_time = %[2]q

  tags = {
    Name = %[1]q
  }
}
`, rName, weeklyMaintenanceStartTime))
}

func testAccWindowsFileSystemConfig_audit(rName, domain, status string) string {
	return acctest.ConfigCompose(testAccWindowsFileSystemConfig_base(rName, domain), fmt.Sprintf(`
resource aws_cloudwatch_log_group "test" {
  name = "/aws/fsx/%[1]s"
}

resource "aws_fsx_windows_file_system" "test" {
  active_directory_id = aws_directory_service_directory.test.id
  skip_final_backup   = true
  storage_capacity    = 32
  subnet_ids          = [aws_subnet.test[0].id]
  throughput_capacity = 32

  audit_log_configuration {
    audit_log_destination             = aws_cloudwatch_log_group.test.arn
    file_access_audit_log_level       = %[2]q
    file_share_access_audit_log_level = %[2]q
  }

  tags = {
    Name = %[1]q
  }
}
`, rName, status))
}

func testAccWindowsFileSystemConfig_auditNoDestination(rName, domain, status string) string {
	return acctest.ConfigCompose(testAccWindowsFileSystemConfig_base(rName, domain), fmt.Sprintf(`
resource aws_cloudwatch_log_group "test" {
  name = "/aws/fsx/%[1]s"
}

resource "aws_fsx_windows_file_system" "test" {
  active_directory_id = aws_directory_service_directory.test.id
  skip_final_backup   = true
  storage_capacity    = 32
  subnet_ids          = [aws_subnet.test[0].id]
  throughput_capacity = 32

  audit_log_configuration {
    file_access_audit_log_level       = %[2]q
    file_share_access_audit_log_level = %[2]q
  }

  tags = {
    Name = %[1]q
  }
}
`, rName, status))
}

func testAccWindowsFileSystemConfig_diskIOPSConfiguration(rName, domain string, iops int) string {
	return acctest.ConfigCompose(testAccWindowsFileSystemConfig_base(rName, domain), fmt.Sprintf(`
resource "aws_fsx_windows_file_system" "test" {
  active_directory_id = aws_directory_service_directory.test.id
  skip_final_backup   = true
  storage_capacity    = 64
  storage_type        = "SSD"
  subnet_ids          = [aws_subnet.test[0].id]
  throughput_capacity = 8

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
