// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package fsx_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/endpoints"
	"github.com/aws/aws-sdk-go/service/fsx"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tffsx "github.com/hashicorp/terraform-provider-aws/internal/service/fsx"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccFSxLustreFileSystem_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var filesystem fsx.FileSystem
	resourceName := "aws_fsx_lustre_file_system.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	deploymentType := fsx.LustreDeploymentTypeScratch1
	if acctest.Partition() == endpoints.AwsUsGovPartitionID {
		deploymentType = fsx.LustreDeploymentTypeScratch2 // SCRATCH_1 not supported in GovCloud
	}

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckPartitionHasService(t, fsx.EndpointsID) },
		ErrorCheck:               acctest.ErrorCheck(t, names.FSxServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckLustreFileSystemDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccLustreFileSystemConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckLustreFileSystemExists(ctx, resourceName, &filesystem),
					acctest.MatchResourceAttrRegionalARN(resourceName, names.AttrARN, "fsx", regexache.MustCompile(`file-system/fs-.+`)),
					resource.TestCheckResourceAttr(resourceName, "automatic_backup_retention_days", "0"),
					resource.TestCheckResourceAttr(resourceName, "copy_tags_to_backups", "false"),
					resource.TestCheckResourceAttr(resourceName, "data_compression_type", fsx.DataCompressionTypeNone),
					resource.TestCheckResourceAttr(resourceName, "deployment_type", deploymentType),
					resource.TestMatchResourceAttr(resourceName, "dns_name", regexache.MustCompile(`fs-.+\.fsx\.`)),
					resource.TestCheckResourceAttr(resourceName, "export_path", ""),
					resource.TestCheckResourceAttr(resourceName, "import_path", ""),
					resource.TestCheckResourceAttr(resourceName, "imported_file_chunk_size", "0"),
					resource.TestCheckResourceAttr(resourceName, "log_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "log_configuration.0.level", "DISABLED"),
					resource.TestCheckResourceAttrSet(resourceName, "mount_name"),
					resource.TestCheckResourceAttr(resourceName, "network_interface_ids.#", "2"),
					acctest.CheckResourceAttrAccountID(resourceName, names.AttrOwnerID),
					resource.TestCheckResourceAttr(resourceName, "security_group_ids.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "storage_capacity", "1200"),
					resource.TestCheckResourceAttr(resourceName, "storage_type", fsx.StorageTypeSsd),
					resource.TestCheckResourceAttr(resourceName, "subnet_ids.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
					resource.TestMatchResourceAttr(resourceName, names.AttrVPCID, regexache.MustCompile(`^vpc-.+`)),
					resource.TestMatchResourceAttr(resourceName, "weekly_maintenance_start_time", regexache.MustCompile(`^\d:\d\d:\d\d$`)),
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
	ctx := acctest.Context(t)
	var filesystem fsx.FileSystem
	resourceName := "aws_fsx_lustre_file_system.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckPartitionHasService(t, fsx.EndpointsID) },
		ErrorCheck:               acctest.ErrorCheck(t, names.FSxServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckLustreFileSystemDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccLustreFileSystemConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLustreFileSystemExists(ctx, resourceName, &filesystem),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tffsx.ResourceLustreFileSystem(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccFSxLustreFileSystem_dataCompression(t *testing.T) {
	ctx := acctest.Context(t)
	var filesystem fsx.FileSystem
	resourceName := "aws_fsx_lustre_file_system.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckPartitionHasService(t, fsx.EndpointsID) },
		ErrorCheck:               acctest.ErrorCheck(t, names.FSxServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckLustreFileSystemDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccLustreFileSystemConfig_compression(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLustreFileSystemExists(ctx, resourceName, &filesystem),
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
				Config: testAccLustreFileSystemConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLustreFileSystemExists(ctx, resourceName, &filesystem),
					resource.TestCheckResourceAttr(resourceName, "data_compression_type", fsx.DataCompressionTypeNone),
				),
			},
			{
				Config: testAccLustreFileSystemConfig_compression(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLustreFileSystemExists(ctx, resourceName, &filesystem),
					resource.TestCheckResourceAttr(resourceName, "data_compression_type", fsx.DataCompressionTypeLz4),
				),
			},
		},
	})
}

func TestAccFSxLustreFileSystem_exportPath(t *testing.T) {
	ctx := acctest.Context(t)
	var filesystem1, filesystem2 fsx.FileSystem
	resourceName := "aws_fsx_lustre_file_system.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckPartitionHasService(t, fsx.EndpointsID) },
		ErrorCheck:               acctest.ErrorCheck(t, names.FSxServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckLustreFileSystemDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccLustreFileSystemConfig_exportPath(rName, ""),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLustreFileSystemExists(ctx, resourceName, &filesystem1),
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
				Config: testAccLustreFileSystemConfig_exportPath(rName, "/prefix/"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLustreFileSystemExists(ctx, resourceName, &filesystem2),
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
	ctx := acctest.Context(t)
	var filesystem1, filesystem2 fsx.FileSystem
	resourceName := "aws_fsx_lustre_file_system.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckPartitionHasService(t, fsx.EndpointsID) },
		ErrorCheck:               acctest.ErrorCheck(t, names.FSxServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckLustreFileSystemDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccLustreFileSystemConfig_importPath(rName, ""),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLustreFileSystemExists(ctx, resourceName, &filesystem1),
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
				Config: testAccLustreFileSystemConfig_importPath(rName, "/prefix/"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLustreFileSystemExists(ctx, resourceName, &filesystem2),
					testAccCheckLustreFileSystemRecreated(&filesystem1, &filesystem2),
					resource.TestCheckResourceAttr(resourceName, "import_path", fmt.Sprintf("s3://%s/prefix/", rName)),
				),
			},
		},
	})
}

// lintignore: AT002
func TestAccFSxLustreFileSystem_importedFileChunkSize(t *testing.T) {
	ctx := acctest.Context(t)
	var filesystem1, filesystem2 fsx.FileSystem
	resourceName := "aws_fsx_lustre_file_system.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckPartitionHasService(t, fsx.EndpointsID) },
		ErrorCheck:               acctest.ErrorCheck(t, names.FSxServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckLustreFileSystemDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccLustreFileSystemConfig_importedChunkSize(rName, 2048),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLustreFileSystemExists(ctx, resourceName, &filesystem1),
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
				Config: testAccLustreFileSystemConfig_importedChunkSize(rName, 4096),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLustreFileSystemExists(ctx, resourceName, &filesystem2),
					testAccCheckLustreFileSystemRecreated(&filesystem1, &filesystem2),
					resource.TestCheckResourceAttr(resourceName, "imported_file_chunk_size", "4096"),
				),
			},
		},
	})
}

func TestAccFSxLustreFileSystem_securityGroupIDs(t *testing.T) {
	ctx := acctest.Context(t)
	var filesystem1, filesystem2 fsx.FileSystem
	resourceName := "aws_fsx_lustre_file_system.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckPartitionHasService(t, fsx.EndpointsID) },
		ErrorCheck:               acctest.ErrorCheck(t, names.FSxServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckLustreFileSystemDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccLustreFileSystemConfig_securityGroupIDs1(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLustreFileSystemExists(ctx, resourceName, &filesystem1),
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
				Config: testAccLustreFileSystemConfig_securityGroupIDs2(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLustreFileSystemExists(ctx, resourceName, &filesystem2),
					testAccCheckLustreFileSystemRecreated(&filesystem1, &filesystem2),
					resource.TestCheckResourceAttr(resourceName, "security_group_ids.#", "2"),
				),
			},
		},
	})
}

func TestAccFSxLustreFileSystem_storageCapacity(t *testing.T) {
	ctx := acctest.Context(t)
	var filesystem1, filesystem2 fsx.FileSystem
	resourceName := "aws_fsx_lustre_file_system.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckPartitionHasService(t, fsx.EndpointsID) },
		ErrorCheck:               acctest.ErrorCheck(t, names.FSxServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckLustreFileSystemDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccLustreFileSystemConfig_storageCapacity(rName, 7200),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLustreFileSystemExists(ctx, resourceName, &filesystem1),
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
				Config: testAccLustreFileSystemConfig_storageCapacity(rName, 1200),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLustreFileSystemExists(ctx, resourceName, &filesystem2),
					testAccCheckLustreFileSystemRecreated(&filesystem1, &filesystem2),
					resource.TestCheckResourceAttr(resourceName, "storage_capacity", "1200"),
				),
			},
		},
	})
}

func TestAccFSxLustreFileSystem_storageCapacityUpdate(t *testing.T) {
	ctx := acctest.Context(t)
	var filesystem1, filesystem2, filesystem3 fsx.FileSystem
	resourceName := "aws_fsx_lustre_file_system.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckPartitionHasService(t, fsx.EndpointsID) },
		ErrorCheck:               acctest.ErrorCheck(t, names.FSxServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckLustreFileSystemDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccLustreFileSystemConfig_storageCapacityScratch2(rName, 7200),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLustreFileSystemExists(ctx, resourceName, &filesystem1),
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
				Config: testAccLustreFileSystemConfig_storageCapacityScratch2(rName, 1200),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLustreFileSystemExists(ctx, resourceName, &filesystem2),
					testAccCheckLustreFileSystemRecreated(&filesystem1, &filesystem2),
					resource.TestCheckResourceAttr(resourceName, "storage_capacity", "1200"),
				),
			},
			{
				Config: testAccLustreFileSystemConfig_storageCapacityScratch2(rName, 7200),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLustreFileSystemExists(ctx, resourceName, &filesystem3),
					testAccCheckLustreFileSystemNotRecreated(&filesystem2, &filesystem3),
					resource.TestCheckResourceAttr(resourceName, "storage_capacity", "7200"),
				),
			},
		},
	})
}

func TestAccFSxLustreFileSystem_fileSystemTypeVersion(t *testing.T) {
	ctx := acctest.Context(t)
	var filesystem1, filesystem2 fsx.FileSystem
	resourceName := "aws_fsx_lustre_file_system.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckPartitionHasService(t, fsx.EndpointsID) },
		ErrorCheck:               acctest.ErrorCheck(t, names.FSxServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckLustreFileSystemDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccLustreFileSystemConfig_typeVersion(rName, "2.10"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLustreFileSystemExists(ctx, resourceName, &filesystem1),
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
				Config: testAccLustreFileSystemConfig_typeVersion(rName, "2.12"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLustreFileSystemExists(ctx, resourceName, &filesystem2),
					testAccCheckLustreFileSystemRecreated(&filesystem1, &filesystem2),
					resource.TestCheckResourceAttr(resourceName, "file_system_type_version", "2.12"),
				),
			},
		},
	})
}

func TestAccFSxLustreFileSystem_tags(t *testing.T) {
	ctx := acctest.Context(t)
	var filesystem1, filesystem2, filesystem3 fsx.FileSystem
	resourceName := "aws_fsx_lustre_file_system.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckPartitionHasService(t, fsx.EndpointsID) },
		ErrorCheck:               acctest.ErrorCheck(t, names.FSxServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckLustreFileSystemDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccLustreFileSystemConfig_tags1(rName, "key1", "value1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLustreFileSystemExists(ctx, resourceName, &filesystem1),
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
				Config: testAccLustreFileSystemConfig_tags2(rName, "key1", "value1updated", "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLustreFileSystemExists(ctx, resourceName, &filesystem2),
					testAccCheckLustreFileSystemNotRecreated(&filesystem1, &filesystem2),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1updated"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
			{
				Config: testAccLustreFileSystemConfig_tags1(rName, "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLustreFileSystemExists(ctx, resourceName, &filesystem3),
					testAccCheckLustreFileSystemNotRecreated(&filesystem2, &filesystem3),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
		},
	})
}

func TestAccFSxLustreFileSystem_weeklyMaintenanceStartTime(t *testing.T) {
	ctx := acctest.Context(t)
	var filesystem1, filesystem2 fsx.FileSystem
	resourceName := "aws_fsx_lustre_file_system.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckPartitionHasService(t, fsx.EndpointsID) },
		ErrorCheck:               acctest.ErrorCheck(t, names.FSxServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckLustreFileSystemDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccLustreFileSystemConfig_weeklyMaintenanceStartTime(rName, "1:01:01"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLustreFileSystemExists(ctx, resourceName, &filesystem1),
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
				Config: testAccLustreFileSystemConfig_weeklyMaintenanceStartTime(rName, "2:02:02"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLustreFileSystemExists(ctx, resourceName, &filesystem2),
					testAccCheckLustreFileSystemNotRecreated(&filesystem1, &filesystem2),
					resource.TestCheckResourceAttr(resourceName, "weekly_maintenance_start_time", "2:02:02"),
				),
			},
		},
	})
}

func TestAccFSxLustreFileSystem_automaticBackupRetentionDays(t *testing.T) {
	ctx := acctest.Context(t)
	var filesystem1, filesystem2 fsx.FileSystem
	resourceName := "aws_fsx_lustre_file_system.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckPartitionHasService(t, fsx.EndpointsID) },
		ErrorCheck:               acctest.ErrorCheck(t, names.FSxServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckLustreFileSystemDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccLustreFileSystemConfig_automaticBackupRetentionDays(rName, 90),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLustreFileSystemExists(ctx, resourceName, &filesystem1),
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
				Config: testAccLustreFileSystemConfig_automaticBackupRetentionDays(rName, 0),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLustreFileSystemExists(ctx, resourceName, &filesystem2),
					testAccCheckLustreFileSystemNotRecreated(&filesystem1, &filesystem2),
					resource.TestCheckResourceAttr(resourceName, "automatic_backup_retention_days", "0"),
				),
			},
			{
				Config: testAccLustreFileSystemConfig_automaticBackupRetentionDays(rName, 1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLustreFileSystemExists(ctx, resourceName, &filesystem1),
					resource.TestCheckResourceAttr(resourceName, "automatic_backup_retention_days", "1"),
				),
			},
		},
	})
}

func TestAccFSxLustreFileSystem_dailyAutomaticBackupStartTime(t *testing.T) {
	ctx := acctest.Context(t)
	var filesystem1, filesystem2 fsx.FileSystem
	resourceName := "aws_fsx_lustre_file_system.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckPartitionHasService(t, fsx.EndpointsID) },
		ErrorCheck:               acctest.ErrorCheck(t, names.FSxServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckLustreFileSystemDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccLustreFileSystemConfig_dailyAutomaticBackupStartTime(rName, "01:01"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLustreFileSystemExists(ctx, resourceName, &filesystem1),
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
				Config: testAccLustreFileSystemConfig_dailyAutomaticBackupStartTime(rName, "02:02"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLustreFileSystemExists(ctx, resourceName, &filesystem2),
					testAccCheckLustreFileSystemNotRecreated(&filesystem1, &filesystem2),
					resource.TestCheckResourceAttr(resourceName, "daily_automatic_backup_start_time", "02:02"),
				),
			},
		},
	})
}

func TestAccFSxLustreFileSystem_deploymentTypePersistent1(t *testing.T) {
	ctx := acctest.Context(t)
	var filesystem fsx.FileSystem
	resourceName := "aws_fsx_lustre_file_system.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckPartitionHasService(t, fsx.EndpointsID) },
		ErrorCheck:               acctest.ErrorCheck(t, names.FSxServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckLustreFileSystemDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccLustreFileSystemConfig_persistent1DeploymentType(rName, 50),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLustreFileSystemExists(ctx, resourceName, &filesystem),
					// per_unit_storage_throughput=50 is only available with deployment_type=PERSISTENT_1, so we test both here.
					resource.TestCheckResourceAttr(resourceName, "per_unit_storage_throughput", "50"),
					resource.TestCheckResourceAttr(resourceName, "deployment_type", fsx.LustreDeploymentTypePersistent1),
					resource.TestCheckResourceAttr(resourceName, "automatic_backup_retention_days", "0"),
					acctest.MatchResourceAttrRegionalARN(resourceName, names.AttrKMSKeyID, "kms", regexache.MustCompile(`key/.+`)),
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

func TestAccFSxLustreFileSystem_deploymentTypePersistent1_perUnitStorageThroughput(t *testing.T) {
	ctx := acctest.Context(t)
	var filesystem1, filesystem2 fsx.FileSystem
	resourceName := "aws_fsx_lustre_file_system.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckPartitionHasService(t, fsx.EndpointsID) },
		ErrorCheck:               acctest.ErrorCheck(t, names.FSxServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckLustreFileSystemDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccLustreFileSystemConfig_persistent1DeploymentType(rName, 50),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLustreFileSystemExists(ctx, resourceName, &filesystem1),
					// per_unit_storage_throughput=50 is only available with deployment_type=PERSISTENT_1, so we test both here.
					resource.TestCheckResourceAttr(resourceName, "per_unit_storage_throughput", "50"),
					resource.TestCheckResourceAttr(resourceName, "deployment_type", fsx.LustreDeploymentTypePersistent1),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"security_group_ids"},
			},
			{
				Config: testAccLustreFileSystemConfig_persistent1DeploymentType(rName, 100),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLustreFileSystemExists(ctx, resourceName, &filesystem2),
					testAccCheckLustreFileSystemNotRecreated(&filesystem1, &filesystem2),
					resource.TestCheckResourceAttr(resourceName, "per_unit_storage_throughput", "100"),
				),
			},
		},
	})
}

func TestAccFSxLustreFileSystem_deploymentTypePersistent2(t *testing.T) {
	ctx := acctest.Context(t)
	var filesystem fsx.FileSystem
	resourceName := "aws_fsx_lustre_file_system.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckPartitionHasService(t, fsx.EndpointsID) },
		ErrorCheck:               acctest.ErrorCheck(t, names.FSxServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckLustreFileSystemDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccLustreFileSystemConfig_persistent2DeploymentType(rName, 125),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLustreFileSystemExists(ctx, resourceName, &filesystem),
					// per_unit_storage_throughput=125 is only available with deployment_type=PERSISTENT_2, so we test both here.
					resource.TestCheckResourceAttr(resourceName, "per_unit_storage_throughput", "125"),
					resource.TestCheckResourceAttr(resourceName, "deployment_type", fsx.LustreDeploymentTypePersistent2),
					resource.TestCheckResourceAttr(resourceName, "automatic_backup_retention_days", "0"),
					acctest.MatchResourceAttrRegionalARN(resourceName, names.AttrKMSKeyID, "kms", regexache.MustCompile(`key/.+`)),
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

func TestAccFSxLustreFileSystem_deploymentTypePersistent2_perUnitStorageThroughput(t *testing.T) {
	ctx := acctest.Context(t)
	var filesystem1, filesystem2 fsx.FileSystem
	resourceName := "aws_fsx_lustre_file_system.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckPartitionHasService(t, fsx.EndpointsID) },
		ErrorCheck:               acctest.ErrorCheck(t, names.FSxServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckLustreFileSystemDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccLustreFileSystemConfig_persistent2DeploymentType(rName, 125),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLustreFileSystemExists(ctx, resourceName, &filesystem1),
					// per_unit_storage_throughput=125 is only available with deployment_type=PERSISTENT_2, so we test both here.
					resource.TestCheckResourceAttr(resourceName, "per_unit_storage_throughput", "125"),
					resource.TestCheckResourceAttr(resourceName, "deployment_type", fsx.LustreDeploymentTypePersistent2),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"security_group_ids"},
			},
			{
				Config: testAccLustreFileSystemConfig_persistent2DeploymentType(rName, 250),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLustreFileSystemExists(ctx, resourceName, &filesystem2),
					testAccCheckLustreFileSystemNotRecreated(&filesystem1, &filesystem2),
					resource.TestCheckResourceAttr(resourceName, "per_unit_storage_throughput", "250"),
				),
			},
		},
	})
}

func TestAccFSxLustreFileSystem_logConfig(t *testing.T) {
	ctx := acctest.Context(t)
	var filesystem fsx.FileSystem
	resourceName := "aws_fsx_lustre_file_system.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckPartitionHasService(t, fsx.EndpointsID) },
		ErrorCheck:               acctest.ErrorCheck(t, names.FSxServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckLustreFileSystemDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccLustreFileSystemConfig_log(rName, "WARN_ONLY"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLustreFileSystemExists(ctx, resourceName, &filesystem),
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
				Config: testAccLustreFileSystemConfig_log(rName, "ERROR_ONLY"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLustreFileSystemExists(ctx, resourceName, &filesystem),
					resource.TestCheckResourceAttr(resourceName, "log_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "log_configuration.0.level", "ERROR_ONLY"),
					resource.TestCheckResourceAttrSet(resourceName, "log_configuration.0.destination"),
				),
			},
		},
	})
}

func TestAccFSxLustreFileSystem_rootSquashConfig(t *testing.T) {
	ctx := acctest.Context(t)
	var filesystem fsx.FileSystem
	resourceName := "aws_fsx_lustre_file_system.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckPartitionHasService(t, fsx.EndpointsID) },
		ErrorCheck:               acctest.ErrorCheck(t, names.FSxServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckLustreFileSystemDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccLustreFileSystemConfig_rootSquash(rName, "365534:65534"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLustreFileSystemExists(ctx, resourceName, &filesystem),
					resource.TestCheckResourceAttr(resourceName, "root_squash_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "root_squash_configuration.0.root_squash", "365534:65534"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"security_group_ids"},
			},
			{
				Config: testAccLustreFileSystemConfig_rootSquash(rName, "355534:64534"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLustreFileSystemExists(ctx, resourceName, &filesystem),
					resource.TestCheckResourceAttr(resourceName, "root_squash_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "root_squash_configuration.0.root_squash", "355534:64534"),
				),
			},
		},
	})
}

func TestAccFSxLustreFileSystem_fromBackup(t *testing.T) {
	ctx := acctest.Context(t)
	var filesystem fsx.FileSystem
	resourceName := "aws_fsx_lustre_file_system.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckPartitionHasService(t, fsx.EndpointsID) },
		ErrorCheck:               acctest.ErrorCheck(t, names.FSxServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckLustreFileSystemDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccLustreFileSystemConfig_fromBackup(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLustreFileSystemExists(ctx, resourceName, &filesystem),
					resource.TestCheckResourceAttr(resourceName, "per_unit_storage_throughput", "50"),
					resource.TestCheckResourceAttr(resourceName, "deployment_type", fsx.LustreDeploymentTypePersistent1),
					resource.TestCheckResourceAttrPair(resourceName, "backup_id", "aws_fsx_backup.test", names.AttrID),
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
	ctx := acctest.Context(t)
	var filesystem1, filesystem2 fsx.FileSystem
	resourceName := "aws_fsx_lustre_file_system.test"
	kmsKeyResourceName1 := "aws_kms_key.test1"
	kmsKeyResourceName2 := "aws_kms_key.test2"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckPartitionHasService(t, fsx.EndpointsID) },
		ErrorCheck:               acctest.ErrorCheck(t, names.FSxServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckLustreFileSystemDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccLustreFileSystemConfig_kmsKeyID1(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLustreFileSystemExists(ctx, resourceName, &filesystem1),
					resource.TestCheckResourceAttr(resourceName, "deployment_type", fsx.LustreDeploymentTypePersistent1),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrKMSKeyID, kmsKeyResourceName1, names.AttrARN),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"security_group_ids"},
			},
			{
				Config: testAccLustreFileSystemConfig_kmsKeyID2(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLustreFileSystemExists(ctx, resourceName, &filesystem2),
					resource.TestCheckResourceAttr(resourceName, "deployment_type", fsx.LustreDeploymentTypePersistent1),
					testAccCheckLustreFileSystemRecreated(&filesystem1, &filesystem2),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrKMSKeyID, kmsKeyResourceName2, names.AttrARN),
				),
			},
		},
	})
}

func TestAccFSxLustreFileSystem_deploymentTypeScratch2(t *testing.T) {
	ctx := acctest.Context(t)
	var filesystem fsx.FileSystem
	resourceName := "aws_fsx_lustre_file_system.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckPartitionHasService(t, fsx.EndpointsID) },
		ErrorCheck:               acctest.ErrorCheck(t, names.FSxServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckLustreFileSystemDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccLustreFileSystemConfig_deploymentType(rName, fsx.LustreDeploymentTypeScratch2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLustreFileSystemExists(ctx, resourceName, &filesystem),
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
	ctx := acctest.Context(t)
	var filesystem fsx.FileSystem
	resourceName := "aws_fsx_lustre_file_system.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckPartitionHasService(t, fsx.EndpointsID) },
		ErrorCheck:               acctest.ErrorCheck(t, names.FSxServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckLustreFileSystemDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccLustreFileSystemConfig_hddStorageType(rName, fsx.DriveCacheTypeRead),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLustreFileSystemExists(ctx, resourceName, &filesystem),
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
	ctx := acctest.Context(t)
	var filesystem fsx.FileSystem
	resourceName := "aws_fsx_lustre_file_system.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckPartitionHasService(t, fsx.EndpointsID) },
		ErrorCheck:               acctest.ErrorCheck(t, names.FSxServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckLustreFileSystemDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccLustreFileSystemConfig_hddStorageType(rName, fsx.DriveCacheTypeNone),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLustreFileSystemExists(ctx, resourceName, &filesystem),
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
	ctx := acctest.Context(t)
	var filesystem fsx.FileSystem
	resourceName := "aws_fsx_lustre_file_system.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckPartitionHasService(t, fsx.EndpointsID) },
		ErrorCheck:               acctest.ErrorCheck(t, names.FSxServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckLustreFileSystemDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccLustreFileSystemConfig_copyTagsToBackups(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLustreFileSystemExists(ctx, resourceName, &filesystem),
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
	ctx := acctest.Context(t)
	var filesystem fsx.FileSystem
	resourceName := "aws_fsx_lustre_file_system.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckPartitionHasService(t, fsx.EndpointsID) },
		ErrorCheck:               acctest.ErrorCheck(t, names.FSxServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckLustreFileSystemDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccLustreFileSystemConfig_autoImportPolicy(rName, "", "NEW"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLustreFileSystemExists(ctx, resourceName, &filesystem),
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
				Config: testAccLustreFileSystemConfig_autoImportPolicy(rName, "", "NEW_CHANGED"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLustreFileSystemExists(ctx, resourceName, &filesystem),
					resource.TestCheckResourceAttr(resourceName, "auto_import_policy", "NEW_CHANGED"),
				),
			},
		},
	})
}

func testAccCheckLustreFileSystemExists(ctx context.Context, n string, v *fsx.FileSystem) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).FSxConn(ctx)

		output, err := tffsx.FindLustreFileSystemByID(ctx, conn, rs.Primary.ID)

		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

func testAccCheckLustreFileSystemDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).FSxConn(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_fsx_lustre_file_system" {
				continue
			}

			_, err := tffsx.FindLustreFileSystemByID(ctx, conn, rs.Primary.ID)

			if tfresource.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("FSx for Lustre File System %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccCheckLustreFileSystemNotRecreated(i, j *fsx.FileSystem) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if aws.StringValue(i.FileSystemId) != aws.StringValue(j.FileSystemId) {
			return fmt.Errorf("FSx for Lustre File System (%s) recreated", aws.StringValue(i.FileSystemId))
		}

		return nil
	}
}

func testAccCheckLustreFileSystemRecreated(i, j *fsx.FileSystem) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if aws.StringValue(i.FileSystemId) == aws.StringValue(j.FileSystemId) {
			return fmt.Errorf("FSx for Lustre File System (%s) not recreated", aws.StringValue(i.FileSystemId))
		}

		return nil
	}
}

func testAccLustreFileSystemConfig_base(rName string) string {
	return acctest.ConfigCompose(acctest.ConfigVPCWithSubnets(rName, 1), `
data "aws_partition" "current" {}
`)
}

func testAccLustreFileSystemConfig_basic(rName string) string {
	return acctest.ConfigCompose(testAccLustreFileSystemConfig_base(rName), `
resource "aws_fsx_lustre_file_system" "test" {
  storage_capacity = 1200
  subnet_ids       = aws_subnet.test[*].id
  deployment_type  = data.aws_partition.current.partition == "aws-us-gov" ? "SCRATCH_2" : null # GovCloud does not support SCRATCH_1
}
`)
}

func testAccLustreFileSystemConfig_exportPath(rName, exportPrefix string) string {
	return acctest.ConfigCompose(testAccLustreFileSystemConfig_base(rName), fmt.Sprintf(`
resource "aws_s3_bucket" "test" {
  bucket = %[1]q
}

resource "aws_fsx_lustre_file_system" "test" {
  export_path      = "s3://${aws_s3_bucket.test.bucket}%[2]s"
  import_path      = "s3://${aws_s3_bucket.test.bucket}"
  storage_capacity = 1200
  subnet_ids       = aws_subnet.test[*].id
  deployment_type  = data.aws_partition.current.partition == "aws-us-gov" ? "SCRATCH_2" : null # GovCloud does not support SCRATCH_1

  tags = {
    Name = %[1]q
  }
}
`, rName, exportPrefix))
}

func testAccLustreFileSystemConfig_importPath(rName, importPrefix string) string {
	return acctest.ConfigCompose(testAccLustreFileSystemConfig_base(rName), fmt.Sprintf(`
resource "aws_s3_bucket" "test" {
  bucket = %[1]q
}

resource "aws_fsx_lustre_file_system" "test" {
  import_path      = "s3://${aws_s3_bucket.test.bucket}%[2]s"
  storage_capacity = 1200
  subnet_ids       = aws_subnet.test[*].id
  deployment_type  = data.aws_partition.current.partition == "aws-us-gov" ? "SCRATCH_2" : null # GovCloud does not support SCRATCH_1

  tags = {
    Name = %[1]q
  }
}
`, rName, importPrefix))
}

func testAccLustreFileSystemConfig_importedChunkSize(rName string, importedFileChunkSize int) string {
	return acctest.ConfigCompose(testAccLustreFileSystemConfig_base(rName), fmt.Sprintf(`
resource "aws_s3_bucket" "test" {
  bucket = %[1]q
}

resource "aws_fsx_lustre_file_system" "test" {
  import_path              = "s3://${aws_s3_bucket.test.bucket}"
  imported_file_chunk_size = %[2]d
  storage_capacity         = 1200
  subnet_ids               = aws_subnet.test[*].id
  deployment_type          = data.aws_partition.current.partition == "aws-us-gov" ? "SCRATCH_2" : null # GovCloud does not support SCRATCH_1

  tags = {
    Name = %[1]q
  }
}
`, rName, importedFileChunkSize))
}

func testAccLustreFileSystemConfig_securityGroupIDs1(rName string) string {
	return acctest.ConfigCompose(testAccLustreFileSystemConfig_base(rName), fmt.Sprintf(`
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

resource "aws_fsx_lustre_file_system" "test" {
  security_group_ids = [aws_security_group.test1.id]
  storage_capacity   = 1200
  subnet_ids         = aws_subnet.test[*].id
  deployment_type    = data.aws_partition.current.partition == "aws-us-gov" ? "SCRATCH_2" : null # GovCloud does not support SCRATCH_1

  tags = {
    Name = %[1]q
  }
}
`, rName))
}

func testAccLustreFileSystemConfig_securityGroupIDs2(rName string) string {
	return acctest.ConfigCompose(testAccLustreFileSystemConfig_base(rName), fmt.Sprintf(`
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

resource "aws_fsx_lustre_file_system" "test" {
  security_group_ids = [aws_security_group.test1.id, aws_security_group.test2.id]
  storage_capacity   = 1200
  subnet_ids         = aws_subnet.test[*].id
  deployment_type    = data.aws_partition.current.partition == "aws-us-gov" ? "SCRATCH_2" : null # GovCloud does not support SCRATCH_1

  tags = {
    Name = %[1]q
  }
}
`, rName))
}

func testAccLustreFileSystemConfig_typeVersion(rName, fileSystemTypeVersion string) string {
	return acctest.ConfigCompose(testAccLustreFileSystemConfig_base(rName), fmt.Sprintf(`
resource "aws_fsx_lustre_file_system" "test" {
  file_system_type_version = %[2]q
  storage_capacity         = 1200
  subnet_ids               = aws_subnet.test[*].id
  deployment_type          = data.aws_partition.current.partition == "aws-us-gov" ? "SCRATCH_2" : null # GovCloud does not support SCRATCH_1

  tags = {
    Name = %[1]q
  }
}
`, rName, fileSystemTypeVersion))
}

func testAccLustreFileSystemConfig_storageCapacity(rName string, storageCapacity int) string {
	return acctest.ConfigCompose(testAccLustreFileSystemConfig_base(rName), fmt.Sprintf(`
resource "aws_fsx_lustre_file_system" "test" {
  storage_capacity = %[2]d
  subnet_ids       = aws_subnet.test[*].id
  deployment_type  = data.aws_partition.current.partition == "aws-us-gov" ? "SCRATCH_2" : null # GovCloud does not support SCRATCH_1

  tags = {
    Name = %[1]q
  }
}
`, rName, storageCapacity))
}

func testAccLustreFileSystemConfig_storageCapacityScratch2(rName string, storageCapacity int) string {
	return acctest.ConfigCompose(testAccLustreFileSystemConfig_base(rName), fmt.Sprintf(`
resource "aws_fsx_lustre_file_system" "test" {
  storage_capacity = %[2]d
  subnet_ids       = aws_subnet.test[*].id
  deployment_type  = "SCRATCH_2"

  tags = {
    Name = %[1]q
  }
}
`, rName, storageCapacity))
}

func testAccLustreFileSystemConfig_tags1(rName, tagKey1, tagValue1 string) string {
	return acctest.ConfigCompose(testAccLustreFileSystemConfig_base(rName), fmt.Sprintf(`
resource "aws_fsx_lustre_file_system" "test" {
  storage_capacity = 1200
  subnet_ids       = aws_subnet.test[*].id
  deployment_type  = data.aws_partition.current.partition == "aws-us-gov" ? "SCRATCH_2" : null # GovCloud does not support SCRATCH_1

  tags = {
    %[1]q = %[2]q
  }
}
`, tagKey1, tagValue1))
}

func testAccLustreFileSystemConfig_tags2(rName, tagKey1, tagValue1, tagKey2, tagValue2 string) string {
	return acctest.ConfigCompose(testAccLustreFileSystemConfig_base(rName), fmt.Sprintf(`
resource "aws_fsx_lustre_file_system" "test" {
  storage_capacity = 1200
  subnet_ids       = aws_subnet.test[*].id
  deployment_type  = data.aws_partition.current.partition == "aws-us-gov" ? "SCRATCH_2" : null # GovCloud does not support SCRATCH_1

  tags = {
    %[1]q = %[2]q
    %[3]q = %[4]q
  }
}
`, tagKey1, tagValue1, tagKey2, tagValue2))
}

func testAccLustreFileSystemConfig_weeklyMaintenanceStartTime(rName, weeklyMaintenanceStartTime string) string {
	return acctest.ConfigCompose(testAccLustreFileSystemConfig_base(rName), fmt.Sprintf(`
resource "aws_fsx_lustre_file_system" "test" {
  storage_capacity              = 1200
  subnet_ids                    = aws_subnet.test[*].id
  weekly_maintenance_start_time = %[2]q
  deployment_type               = data.aws_partition.current.partition == "aws-us-gov" ? "SCRATCH_2" : null # GovCloud does not support SCRATCH_1

  tags = {
    Name = %[1]q
  }
}
`, rName, weeklyMaintenanceStartTime))
}

func testAccLustreFileSystemConfig_dailyAutomaticBackupStartTime(rName, dailyAutomaticBackupStartTime string) string {
	return acctest.ConfigCompose(testAccLustreFileSystemConfig_base(rName), fmt.Sprintf(`
resource "aws_fsx_lustre_file_system" "test" {
  storage_capacity                  = 1200
  subnet_ids                        = aws_subnet.test[*].id
  deployment_type                   = "PERSISTENT_1"
  per_unit_storage_throughput       = 50
  daily_automatic_backup_start_time = %[2]q
  automatic_backup_retention_days   = 1

  tags = {
    Name = %[1]q
  }
}
`, rName, dailyAutomaticBackupStartTime))
}

func testAccLustreFileSystemConfig_automaticBackupRetentionDays(rName string, retention int) string {
	return acctest.ConfigCompose(testAccLustreFileSystemConfig_base(rName), fmt.Sprintf(`
resource "aws_fsx_lustre_file_system" "test" {
  storage_capacity                = 1200
  subnet_ids                      = aws_subnet.test[*].id
  deployment_type                 = "PERSISTENT_1"
  per_unit_storage_throughput     = 50
  automatic_backup_retention_days = %[2]d

  tags = {
    Name = %[1]q
  }
}
`, rName, retention))
}

func testAccLustreFileSystemConfig_deploymentType(rName, deploymentType string) string {
	return acctest.ConfigCompose(testAccLustreFileSystemConfig_base(rName), fmt.Sprintf(`
resource "aws_fsx_lustre_file_system" "test" {
  storage_capacity = 1200
  subnet_ids       = aws_subnet.test[*].id
  deployment_type  = %[2]q

  tags = {
    Name = %[1]q
  }
}
`, rName, deploymentType))
}

func testAccLustreFileSystemConfig_persistent1DeploymentType(rName string, perUnitStorageThroughput int) string {
	return acctest.ConfigCompose(testAccLustreFileSystemConfig_base(rName), fmt.Sprintf(`
resource "aws_fsx_lustre_file_system" "test" {
  storage_capacity            = 1200
  subnet_ids                  = aws_subnet.test[*].id
  deployment_type             = "PERSISTENT_1"
  per_unit_storage_throughput = %[2]d

  tags = {
    Name = %[1]q
  }
}
`, rName, perUnitStorageThroughput))
}

func testAccLustreFileSystemConfig_persistent2DeploymentType(rName string, perUnitStorageThroughput int) string {
	return acctest.ConfigCompose(testAccLustreFileSystemConfig_base(rName), fmt.Sprintf(`
resource "aws_fsx_lustre_file_system" "test" {
  storage_capacity            = 1200
  subnet_ids                  = aws_subnet.test[*].id
  deployment_type             = "PERSISTENT_2"
  per_unit_storage_throughput = %[2]d

  tags = {
    Name = %[1]q
  }
}
`, rName, perUnitStorageThroughput))
}

func testAccLustreFileSystemConfig_fromBackup(rName string) string {
	return acctest.ConfigCompose(testAccLustreFileSystemConfig_base(rName), fmt.Sprintf(`
resource "aws_fsx_lustre_file_system" "base" {
  storage_capacity            = 1200
  subnet_ids                  = aws_subnet.test[*].id
  deployment_type             = "PERSISTENT_1"
  per_unit_storage_throughput = 50

  tags = {
    Name = %[1]q
  }
}

resource "aws_fsx_backup" "test" {
  file_system_id = aws_fsx_lustre_file_system.base.id

  tags = {
    Name = %[1]q
  }
}

resource "aws_fsx_lustre_file_system" "test" {
  storage_capacity            = 1200
  subnet_ids                  = aws_subnet.test[*].id
  deployment_type             = "PERSISTENT_1"
  per_unit_storage_throughput = 50
  backup_id                   = aws_fsx_backup.test.id

  tags = {
    Name = %[1]q
  }
}
`, rName))
}

func testAccLustreFileSystemConfig_kmsKeyID1(rName string) string {
	return acctest.ConfigCompose(testAccLustreFileSystemConfig_base(rName), fmt.Sprintf(`
resource "aws_kms_key" "test1" {
  description             = "%[1]s-1"
  deletion_window_in_days = 7
}

resource "aws_fsx_lustre_file_system" "test" {
  storage_capacity            = 1200
  subnet_ids                  = aws_subnet.test[*].id
  deployment_type             = "PERSISTENT_1"
  per_unit_storage_throughput = 50
  kms_key_id                  = aws_kms_key.test1.arn

  tags = {
    Name = %[1]q
  }
}
`, rName))
}

func testAccLustreFileSystemConfig_kmsKeyID2(rName string) string {
	return acctest.ConfigCompose(testAccLustreFileSystemConfig_base(rName), fmt.Sprintf(`
resource "aws_kms_key" "test2" {
  description             = "%[1]s-2"
  deletion_window_in_days = 7
}

resource "aws_fsx_lustre_file_system" "test" {
  storage_capacity            = 1200
  subnet_ids                  = aws_subnet.test[*].id
  deployment_type             = "PERSISTENT_1"
  per_unit_storage_throughput = 50
  kms_key_id                  = aws_kms_key.test2.arn

  tags = {
    Name = %[1]q
  }
}
`, rName))
}

func testAccLustreFileSystemConfig_hddStorageType(rName, driveCacheType string) string {
	return acctest.ConfigCompose(testAccLustreFileSystemConfig_base(rName), fmt.Sprintf(`
resource "aws_fsx_lustre_file_system" "test" {
  storage_capacity            = 6000
  subnet_ids                  = aws_subnet.test[*].id
  deployment_type             = "PERSISTENT_1"
  per_unit_storage_throughput = 12
  storage_type                = "HDD"
  drive_cache_type            = %[2]q

  tags = {
    Name = %[1]q
  }
}
`, rName, driveCacheType))
}

func testAccLustreFileSystemConfig_autoImportPolicy(rName, exportPrefix, policy string) string {
	return acctest.ConfigCompose(testAccLustreFileSystemConfig_base(rName), fmt.Sprintf(`
resource "aws_s3_bucket" "test" {
  bucket = %[1]q
}

resource "aws_fsx_lustre_file_system" "test" {
  export_path        = "s3://${aws_s3_bucket.test.bucket}%[2]s"
  import_path        = "s3://${aws_s3_bucket.test.bucket}"
  auto_import_policy = %[3]q
  storage_capacity   = 1200
  subnet_ids         = aws_subnet.test[*].id
  deployment_type    = data.aws_partition.current.partition == "aws-us-gov" ? "SCRATCH_2" : null # GovCloud does not support SCRATCH_1

  tags = {
    Name = %[1]q
  }
}
`, rName, exportPrefix, policy))
}

func testAccLustreFileSystemConfig_copyTagsToBackups(rName string) string {
	return acctest.ConfigCompose(testAccLustreFileSystemConfig_base(rName), fmt.Sprintf(`
resource "aws_fsx_lustre_file_system" "test" {
  storage_capacity            = 1200
  deployment_type             = "PERSISTENT_1"
  subnet_ids                  = aws_subnet.test[*].id
  per_unit_storage_throughput = 50
  copy_tags_to_backups        = true

  tags = {
    Name = %[1]q
  }
}
`, rName))
}

func testAccLustreFileSystemConfig_compression(rName string) string {
	return acctest.ConfigCompose(testAccLustreFileSystemConfig_base(rName), fmt.Sprintf(`
resource "aws_fsx_lustre_file_system" "test" {
  storage_capacity      = 1200
  subnet_ids            = aws_subnet.test[*].id
  deployment_type       = data.aws_partition.current.partition == "aws-us-gov" ? "SCRATCH_2" : null # GovCloud does not support SCRATCH_1
  data_compression_type = "LZ4"

  tags = {
    Name = %[1]q
  }
}
`, rName))
}

func testAccLustreFileSystemConfig_log(rName, status string) string {
	return acctest.ConfigCompose(testAccLustreFileSystemConfig_base(rName), fmt.Sprintf(`
resource aws_cloudwatch_log_group "test" {
  name = "/aws/fsx/%[1]s"
}

resource "aws_fsx_lustre_file_system" "test" {
  storage_capacity = 1200
  subnet_ids       = aws_subnet.test[*].id
  deployment_type  = data.aws_partition.current.partition == "aws-us-gov" ? "SCRATCH_2" : null # GovCloud does not support SCRATCH_1

  log_configuration {
    destination = aws_cloudwatch_log_group.test.arn
    level       = %[2]q
  }

  tags = {
    Name = %[1]q
  }
}
`, rName, status))
}

func testAccLustreFileSystemConfig_rootSquash(rName, uid string) string {
	return acctest.ConfigCompose(testAccLustreFileSystemConfig_base(rName), fmt.Sprintf(`
resource "aws_fsx_lustre_file_system" "test" {
  storage_capacity = 1200
  subnet_ids       = aws_subnet.test[*].id
  deployment_type  = data.aws_partition.current.partition == "aws-us-gov" ? "SCRATCH_2" : null # GovCloud does not support SCRATCH_1

  root_squash_configuration {
    root_squash = %[2]q
  }

  tags = {
    Name = %[1]q
  }
}
`, rName, uid))
}
