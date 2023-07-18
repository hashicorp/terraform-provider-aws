// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package fsx_test

import (
	"context"
	"fmt"
	"regexp"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/fsx"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tffsx "github.com/hashicorp/terraform-provider-aws/internal/service/fsx"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

func TestAccFSxWindowsFileSystem_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var filesystem fsx.FileSystem
	resourceName := "aws_fsx_windows_file_system.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	domainName := acctest.RandomDomainName()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckPartitionHasService(t, fsx.EndpointsID) },
		ErrorCheck:               acctest.ErrorCheck(t, fsx.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckWindowsFileSystemDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccWindowsFileSystemConfig_subnetIDs1(rName, domainName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckWindowsFileSystemExists(ctx, resourceName, &filesystem),
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
				Config:   testAccWindowsFileSystemConfig_subnetIDs1SingleType(rName, domainName, "SINGLE_AZ_1"),
				PlanOnly: true,
			},
		},
	})
}

func TestAccFSxWindowsFileSystem_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	var filesystem fsx.FileSystem
	resourceName := "aws_fsx_windows_file_system.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	domainName := acctest.RandomDomainName()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckPartitionHasService(t, fsx.EndpointsID) },
		ErrorCheck:               acctest.ErrorCheck(t, fsx.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckWindowsFileSystemDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccWindowsFileSystemConfig_subnetIDs1(rName, domainName),
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
	var filesystem fsx.FileSystem
	resourceName := "aws_fsx_windows_file_system.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	domainName := acctest.RandomDomainName()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckPartitionHasService(t, fsx.EndpointsID) },
		ErrorCheck:               acctest.ErrorCheck(t, fsx.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckWindowsFileSystemDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccWindowsFileSystemConfig_subnetIDs1SingleType(rName, domainName, "SINGLE_AZ_2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckWindowsFileSystemExists(ctx, resourceName, &filesystem),
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
	ctx := acctest.Context(t)
	var filesystem fsx.FileSystem
	resourceName := "aws_fsx_windows_file_system.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	domainName := acctest.RandomDomainName()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckPartitionHasService(t, fsx.EndpointsID) },
		ErrorCheck:               acctest.ErrorCheck(t, fsx.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckWindowsFileSystemDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccWindowsFileSystemConfig_subnetIDs1StorageType(rName, domainName, "SINGLE_AZ_2", "HDD"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckWindowsFileSystemExists(ctx, resourceName, &filesystem),
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
	ctx := acctest.Context(t)
	var filesystem fsx.FileSystem
	resourceName := "aws_fsx_windows_file_system.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	domainName := acctest.RandomDomainName()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckPartitionHasService(t, fsx.EndpointsID) },
		ErrorCheck:               acctest.ErrorCheck(t, fsx.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckWindowsFileSystemDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccWindowsFileSystemConfig_subnetIDs2(rName, domainName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckWindowsFileSystemExists(ctx, resourceName, &filesystem),
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

func TestAccFSxWindowsFileSystem_aliases(t *testing.T) {
	ctx := acctest.Context(t)
	var filesystem1, filesystem2, filesystem3 fsx.FileSystem
	resourceName := "aws_fsx_windows_file_system.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	domainName := acctest.RandomDomainName()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckPartitionHasService(t, fsx.EndpointsID) },
		ErrorCheck:               acctest.ErrorCheck(t, fsx.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckWindowsFileSystemDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccWindowsFileSystemConfig_aliases1(rName, domainName, "filesystem1.example.com"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckWindowsFileSystemExists(ctx, resourceName, &filesystem1),
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
				Config: testAccWindowsFileSystemConfig_aliases2(rName, domainName, "filesystem2.example.com", "filesystem3.example.com"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckWindowsFileSystemExists(ctx, resourceName, &filesystem2),
					testAccCheckWindowsFileSystemNotRecreated(&filesystem1, &filesystem2),
					resource.TestCheckResourceAttr(resourceName, "aliases.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "aliases.0", "filesystem2.example.com"),
					resource.TestCheckResourceAttr(resourceName, "aliases.1", "filesystem3.example.com"),
				),
			},
			{
				Config: testAccWindowsFileSystemConfig_aliases1(rName, domainName, "filesystem3.example.com"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckWindowsFileSystemExists(ctx, resourceName, &filesystem3),
					testAccCheckWindowsFileSystemNotRecreated(&filesystem2, &filesystem3),
					resource.TestCheckResourceAttr(resourceName, "aliases.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "aliases.0", "filesystem3.example.com"),
				),
			},
		},
	})
}

func TestAccFSxWindowsFileSystem_automaticBackupRetentionDays(t *testing.T) {
	ctx := acctest.Context(t)
	var filesystem1, filesystem2, filesystem3 fsx.FileSystem
	resourceName := "aws_fsx_windows_file_system.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	domainName := acctest.RandomDomainName()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckPartitionHasService(t, fsx.EndpointsID) },
		ErrorCheck:               acctest.ErrorCheck(t, fsx.EndpointsID),
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
					"security_group_ids",
					"skip_final_backup",
				},
			},
			{
				Config: testAccWindowsFileSystemConfig_automaticBackupRetentionDays(rName, domainName, 0),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckWindowsFileSystemExists(ctx, resourceName, &filesystem2),
					testAccCheckWindowsFileSystemNotRecreated(&filesystem1, &filesystem2),
					resource.TestCheckResourceAttr(resourceName, "automatic_backup_retention_days", "0"),
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
	var filesystem1, filesystem2 fsx.FileSystem
	resourceName := "aws_fsx_windows_file_system.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	domainName := acctest.RandomDomainName()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckPartitionHasService(t, fsx.EndpointsID) },
		ErrorCheck:               acctest.ErrorCheck(t, fsx.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckWindowsFileSystemDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccWindowsFileSystemConfig_copyTagsToBackups(rName, domainName, true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckWindowsFileSystemExists(ctx, resourceName, &filesystem1),
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
				Config: testAccWindowsFileSystemConfig_copyTagsToBackups(rName, domainName, false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckWindowsFileSystemExists(ctx, resourceName, &filesystem2),
					testAccCheckWindowsFileSystemRecreated(&filesystem1, &filesystem2),
					resource.TestCheckResourceAttr(resourceName, "copy_tags_to_backups", "false"),
				),
			},
		},
	})
}

func TestAccFSxWindowsFileSystem_dailyAutomaticBackupStartTime(t *testing.T) {
	ctx := acctest.Context(t)
	var filesystem1, filesystem2 fsx.FileSystem
	resourceName := "aws_fsx_windows_file_system.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	domainName := acctest.RandomDomainName()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckPartitionHasService(t, fsx.EndpointsID) },
		ErrorCheck:               acctest.ErrorCheck(t, fsx.EndpointsID),
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
					"security_group_ids",
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

func TestAccFSxWindowsFileSystem_kmsKeyID(t *testing.T) {
	ctx := acctest.Context(t)
	var filesystem1, filesystem2 fsx.FileSystem
	kmsKeyResourceName1 := "aws_kms_key.test1"
	kmsKeyResourceName2 := "aws_kms_key.test2"
	resourceName := "aws_fsx_windows_file_system.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	domainName := acctest.RandomDomainName()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckPartitionHasService(t, fsx.EndpointsID) },
		ErrorCheck:               acctest.ErrorCheck(t, fsx.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckWindowsFileSystemDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccWindowsFileSystemConfig_kmsKeyID1(rName, domainName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckWindowsFileSystemExists(ctx, resourceName, &filesystem1),
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
				Config: testAccWindowsFileSystemConfig_kmsKeyID2(rName, domainName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckWindowsFileSystemExists(ctx, resourceName, &filesystem2),
					testAccCheckWindowsFileSystemRecreated(&filesystem1, &filesystem2),
					resource.TestCheckResourceAttrPair(resourceName, "kms_key_id", kmsKeyResourceName2, "arn"),
				),
			},
		},
	})
}

func TestAccFSxWindowsFileSystem_securityGroupIDs(t *testing.T) {
	ctx := acctest.Context(t)
	var filesystem1, filesystem2 fsx.FileSystem
	resourceName := "aws_fsx_windows_file_system.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	domainName := acctest.RandomDomainName()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckPartitionHasService(t, fsx.EndpointsID) },
		ErrorCheck:               acctest.ErrorCheck(t, fsx.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckWindowsFileSystemDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccWindowsFileSystemConfig_securityGroupIDs1(rName, domainName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckWindowsFileSystemExists(ctx, resourceName, &filesystem1),
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
				Config: testAccWindowsFileSystemConfig_securityGroupIDs2(rName, domainName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckWindowsFileSystemExists(ctx, resourceName, &filesystem2),
					testAccCheckWindowsFileSystemRecreated(&filesystem1, &filesystem2),
					resource.TestCheckResourceAttr(resourceName, "security_group_ids.#", "2"),
				),
			},
		},
	})
}

func TestAccFSxWindowsFileSystem_selfManagedActiveDirectory(t *testing.T) {
	ctx := acctest.Context(t)
	var filesystem fsx.FileSystem
	resourceName := "aws_fsx_windows_file_system.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	domainName := acctest.RandomDomainName()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckPartitionHasService(t, fsx.EndpointsID) },
		ErrorCheck:               acctest.ErrorCheck(t, fsx.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckWindowsFileSystemDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccWindowsFileSystemConfig_selfManagedActiveDirectory(rName, domainName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckWindowsFileSystemExists(ctx, resourceName, &filesystem),
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
	ctx := acctest.Context(t)
	var filesystem fsx.FileSystem
	resourceName := "aws_fsx_windows_file_system.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	domainName := acctest.RandomDomainName()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckPartitionHasService(t, fsx.EndpointsID) },
		ErrorCheck:               acctest.ErrorCheck(t, fsx.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckWindowsFileSystemDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccWindowsFileSystemConfig_selfManagedActiveDirectoryUsername(rName, domainName, "Admin"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckWindowsFileSystemExists(ctx, resourceName, &filesystem),
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
				Config: testAccWindowsFileSystemConfig_selfManagedActiveDirectoryUsername(rName, domainName, "Administrator"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckWindowsFileSystemExists(ctx, resourceName, &filesystem),
					resource.TestCheckResourceAttr(resourceName, "self_managed_active_directory.#", "1"),
				),
			},
		},
	})
}

func TestAccFSxWindowsFileSystem_storageCapacity(t *testing.T) {
	ctx := acctest.Context(t)
	var filesystem1, filesystem2 fsx.FileSystem
	resourceName := "aws_fsx_windows_file_system.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	domainName := acctest.RandomDomainName()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckPartitionHasService(t, fsx.EndpointsID) },
		ErrorCheck:               acctest.ErrorCheck(t, fsx.EndpointsID),
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
					"security_group_ids",
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
	var filesystem fsx.FileSystem
	resourceName := "aws_fsx_windows_file_system.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	domainName := acctest.RandomDomainName()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckPartitionHasService(t, fsx.EndpointsID) },
		ErrorCheck:               acctest.ErrorCheck(t, fsx.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckWindowsFileSystemDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccWindowsFileSystemConfig_fromBackup(rName, domainName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckWindowsFileSystemExists(ctx, resourceName, &filesystem),
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
	ctx := acctest.Context(t)
	var filesystem1, filesystem2, filesystem3 fsx.FileSystem
	resourceName := "aws_fsx_windows_file_system.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	domainName := acctest.RandomDomainName()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckPartitionHasService(t, fsx.EndpointsID) },
		ErrorCheck:               acctest.ErrorCheck(t, fsx.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckWindowsFileSystemDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccWindowsFileSystemConfig_tags1(rName, domainName, "key1", "value1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckWindowsFileSystemExists(ctx, resourceName, &filesystem1),
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
				Config: testAccWindowsFileSystemConfig_tags2(rName, domainName, "key1", "value1updated", "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckWindowsFileSystemExists(ctx, resourceName, &filesystem2),
					testAccCheckWindowsFileSystemNotRecreated(&filesystem1, &filesystem2),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1updated"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
			{
				Config: testAccWindowsFileSystemConfig_tags1(rName, domainName, "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckWindowsFileSystemExists(ctx, resourceName, &filesystem3),
					testAccCheckWindowsFileSystemNotRecreated(&filesystem2, &filesystem3),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
		},
	})
}

func TestAccFSxWindowsFileSystem_throughputCapacity(t *testing.T) {
	ctx := acctest.Context(t)
	var filesystem1, filesystem2 fsx.FileSystem
	resourceName := "aws_fsx_windows_file_system.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	domainName := acctest.RandomDomainName()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckPartitionHasService(t, fsx.EndpointsID) },
		ErrorCheck:               acctest.ErrorCheck(t, fsx.EndpointsID),
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
					"security_group_ids",
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
	var filesystem1, filesystem2 fsx.FileSystem
	resourceName := "aws_fsx_windows_file_system.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	domainName := acctest.RandomDomainName()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckPartitionHasService(t, fsx.EndpointsID) },
		ErrorCheck:               acctest.ErrorCheck(t, fsx.EndpointsID),
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
					"security_group_ids",
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
	var filesystem fsx.FileSystem
	resourceName := "aws_fsx_windows_file_system.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	domainName := acctest.RandomDomainName()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckPartitionHasService(t, fsx.EndpointsID) },
		ErrorCheck:               acctest.ErrorCheck(t, fsx.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckWindowsFileSystemDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccWindowsFileSystemConfig_audit(rName, domainName, "SUCCESS_ONLY"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckWindowsFileSystemExists(ctx, resourceName, &filesystem),
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
				Config: testAccWindowsFileSystemConfig_audit(rName, domainName, "SUCCESS_AND_FAILURE"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckWindowsFileSystemExists(ctx, resourceName, &filesystem),
					resource.TestCheckResourceAttr(resourceName, "audit_log_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "audit_log_configuration.0.file_access_audit_log_level", "SUCCESS_AND_FAILURE"),
					resource.TestCheckResourceAttr(resourceName, "audit_log_configuration.0.file_share_access_audit_log_level", "SUCCESS_AND_FAILURE"),
					resource.TestCheckResourceAttrSet(resourceName, "audit_log_configuration.0.audit_log_destination"),
				),
			},
		},
	})
}

func testAccCheckWindowsFileSystemExists(ctx context.Context, n string, v *fsx.FileSystem) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).FSxConn(ctx)

		output, err := tffsx.FindFileSystemByID(ctx, conn, rs.Primary.ID)

		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

func testAccCheckWindowsFileSystemDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).FSxConn(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_fsx_windows_file_system" {
				continue
			}

			_, err := tffsx.FindFileSystemByID(ctx, conn, rs.Primary.ID)

			if tfresource.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("FSx Windows File System (%s) still exists", rs.Primary.ID)
		}
		return nil
	}
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

func testAccWindowsFileSystemConfig_aliases1(rName, domain, alias1 string) string {
	return acctest.ConfigCompose(testAccWindowsFileSystemConfig_base(rName, domain), fmt.Sprintf(`
resource "aws_fsx_windows_file_system" "test" {
  active_directory_id = aws_directory_service_directory.test.id
  skip_final_backup   = true
  storage_capacity    = 32
  subnet_ids          = [aws_subnet.test[0].id]
  throughput_capacity = 8

  aliases = [%[1]q]
}
`, alias1))
}

func testAccWindowsFileSystemConfig_aliases2(rName, domain, alias1, alias2 string) string {
	return acctest.ConfigCompose(testAccWindowsFileSystemConfig_base(rName, domain), fmt.Sprintf(`
resource "aws_fsx_windows_file_system" "test" {
  active_directory_id = aws_directory_service_directory.test.id
  skip_final_backup   = true
  storage_capacity    = 32
  subnet_ids          = [aws_subnet.test[0].id]
  throughput_capacity = 8

  aliases = [%[1]q, %[2]q]
}
`, alias1, alias2))
}

func testAccWindowsFileSystemConfig_automaticBackupRetentionDays(rName, domain string, automaticBackupRetentionDays int) string {
	return acctest.ConfigCompose(testAccWindowsFileSystemConfig_base(rName, domain), fmt.Sprintf(`
resource "aws_fsx_windows_file_system" "test" {
  active_directory_id             = aws_directory_service_directory.test.id
  automatic_backup_retention_days = %[1]d
  skip_final_backup               = true
  storage_capacity                = 32
  subnet_ids                      = [aws_subnet.test[0].id]
  throughput_capacity             = 8
}
`, automaticBackupRetentionDays))
}

func testAccWindowsFileSystemConfig_copyTagsToBackups(rName, domain string, copyTagsToBackups bool) string {
	return acctest.ConfigCompose(testAccWindowsFileSystemConfig_base(rName, domain), fmt.Sprintf(`
resource "aws_fsx_windows_file_system" "test" {
  active_directory_id  = aws_directory_service_directory.test.id
  copy_tags_to_backups = %[1]t
  skip_final_backup    = true
  storage_capacity     = 32
  subnet_ids           = [aws_subnet.test[0].id]
  throughput_capacity  = 8
}
`, copyTagsToBackups))
}

func testAccWindowsFileSystemConfig_dailyAutomaticBackupStartTime(rName, domain, dailyAutomaticBackupStartTime string) string {
	return acctest.ConfigCompose(testAccWindowsFileSystemConfig_base(rName, domain), fmt.Sprintf(`
resource "aws_fsx_windows_file_system" "test" {
  active_directory_id               = aws_directory_service_directory.test.id
  daily_automatic_backup_start_time = %[1]q
  skip_final_backup                 = true
  storage_capacity                  = 32
  subnet_ids                        = [aws_subnet.test[0].id]
  throughput_capacity               = 8
}
`, dailyAutomaticBackupStartTime))
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
}
`, rName))
}

func testAccWindowsFileSystemConfig_selfManagedActiveDirectory(rName, domain string) string {
	return acctest.ConfigCompose(testAccWindowsFileSystemConfig_base(rName, domain), `
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
}
`)
}

func testAccWindowsFileSystemConfig_selfManagedActiveDirectoryUsername(rName, domain, username string) string {
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
    username    = %[1]q
  }
}
`, username))
}

func testAccWindowsFileSystemConfig_storageCapacity(rName, domain string, storageCapacity, throughputCapacity int) string {
	return acctest.ConfigCompose(testAccWindowsFileSystemConfig_base(rName, domain), fmt.Sprintf(`
resource "aws_fsx_windows_file_system" "test" {
  active_directory_id = aws_directory_service_directory.test.id
  skip_final_backup   = true
  storage_capacity    = %[1]d
  subnet_ids          = [aws_subnet.test[0].id]
  throughput_capacity = %[2]d
}
`, storageCapacity, throughputCapacity))
}

func testAccWindowsFileSystemConfig_subnetIDs1(rName, domain string) string {
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

func testAccWindowsFileSystemConfig_subnetIDs1SingleType(rName, domain, azType string) string {
	return acctest.ConfigCompose(testAccWindowsFileSystemConfig_base(rName, domain), fmt.Sprintf(`
resource "aws_fsx_windows_file_system" "test" {
  active_directory_id = aws_directory_service_directory.test.id
  skip_final_backup   = true
  storage_capacity    = 32
  deployment_type     = %[1]q
  subnet_ids          = [aws_subnet.test[0].id]
  throughput_capacity = 8
}
`, azType))
}

func testAccWindowsFileSystemConfig_subnetIDs1StorageType(rName, domain, azType, storageType string) string {
	return acctest.ConfigCompose(testAccWindowsFileSystemConfig_base(rName, domain), fmt.Sprintf(`
resource "aws_fsx_windows_file_system" "test" {
  active_directory_id = aws_directory_service_directory.test.id
  skip_final_backup   = true
  storage_capacity    = 2000
  deployment_type     = %[1]q
  subnet_ids          = [aws_subnet.test[0].id]
  throughput_capacity = 8
  storage_type        = %[2]q
}
`, azType, storageType))
}

func testAccWindowsFileSystemConfig_subnetIDs2(rName, domain string) string {
	return acctest.ConfigCompose(testAccWindowsFileSystemConfig_base(rName, domain), `
resource "aws_fsx_windows_file_system" "test" {
  active_directory_id = aws_directory_service_directory.test.id
  skip_final_backup   = true
  storage_capacity    = 32
  deployment_type     = "MULTI_AZ_1"
  subnet_ids          = [aws_subnet.test[0].id, aws_subnet.test[1].id]
  preferred_subnet_id = aws_subnet.test[0].id
  throughput_capacity = 8
}
`)
}

func testAccWindowsFileSystemConfig_fromBackup(rName, domain string) string {
	return acctest.ConfigCompose(testAccWindowsFileSystemConfig_base(rName, domain), `
resource "aws_fsx_windows_file_system" "base" {
  active_directory_id = aws_directory_service_directory.test.id
  skip_final_backup   = true
  storage_capacity    = 32
  subnet_ids          = [aws_subnet.test[0].id]
  throughput_capacity = 8
}

resource "aws_fsx_backup" "test" {
  file_system_id = aws_fsx_windows_file_system.base.id
}

resource "aws_fsx_windows_file_system" "test" {
  active_directory_id = aws_directory_service_directory.test.id
  backup_id           = aws_fsx_backup.test.id
  skip_final_backup   = true
  subnet_ids          = [aws_subnet.test[0].id]
  throughput_capacity = 8
}
`)
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
  throughput_capacity = %[1]d
}
`, throughputCapacity))
}

func testAccWindowsFileSystemConfig_weeklyMaintenanceStartTime(rName, domain, weeklyMaintenanceStartTime string) string {
	return acctest.ConfigCompose(testAccWindowsFileSystemConfig_base(rName, domain), fmt.Sprintf(`
resource "aws_fsx_windows_file_system" "test" {
  active_directory_id           = aws_directory_service_directory.test.id
  skip_final_backup             = true
  storage_capacity              = 32
  subnet_ids                    = [aws_subnet.test[0].id]
  throughput_capacity           = 8
  weekly_maintenance_start_time = %[1]q
}
`, weeklyMaintenanceStartTime))
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
}
`, rName, status))
}
