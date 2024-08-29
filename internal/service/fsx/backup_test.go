// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package fsx_test

import (
	"context"
	"fmt"
	"strings"
	"testing"

	"github.com/YakDriver/regexache"
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

func TestAccFSxBackup_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var backup awstypes.Backup
	resourceName := "aws_fsx_backup.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckPartitionHasService(t, names.FSxEndpointID) },
		ErrorCheck:               acctest.ErrorCheck(t, names.FSxServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckBackupDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccBackupConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBackupExists(ctx, resourceName, &backup),
					acctest.MatchResourceAttrRegionalARN(resourceName, names.AttrARN, "fsx", regexache.MustCompile(`backup/.+`)),
					acctest.CheckResourceAttrAccountID(resourceName, names.AttrOwnerID),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct1),
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
	ctx := acctest.Context(t)
	var backup awstypes.Backup
	resourceName := "aws_fsx_backup.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	//FSX ONTAP Volume Names can't use dash only underscore
	vName := strings.Replace(sdkacctest.RandomWithPrefix(acctest.ResourcePrefix), "-", "_", -1)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckPartitionHasService(t, names.FSxEndpointID) },
		ErrorCheck:               acctest.ErrorCheck(t, names.FSxServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckBackupDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccBackupConfig_ontapBasic(rName, vName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBackupExists(ctx, resourceName, &backup),
					acctest.MatchResourceAttrRegionalARN(resourceName, names.AttrARN, "fsx", regexache.MustCompile(`backup/.+`)),
					acctest.CheckResourceAttrAccountID(resourceName, names.AttrOwnerID),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct1),
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
	ctx := acctest.Context(t)
	var backup awstypes.Backup
	resourceName := "aws_fsx_backup.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckPartitionHasService(t, names.FSxEndpointID) },
		ErrorCheck:               acctest.ErrorCheck(t, names.FSxServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckBackupDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccBackupConfig_openZFSBasic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBackupExists(ctx, resourceName, &backup),
					acctest.MatchResourceAttrRegionalARN(resourceName, names.AttrARN, "fsx", regexache.MustCompile(`backup/.+`)),
					acctest.CheckResourceAttrAccountID(resourceName, names.AttrOwnerID),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct1),
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
	ctx := acctest.Context(t)
	var backup awstypes.Backup
	resourceName := "aws_fsx_backup.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	domainName := acctest.RandomDomainName()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckPartitionHasService(t, names.FSxEndpointID) },
		ErrorCheck:               acctest.ErrorCheck(t, names.FSxServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckBackupDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccBackupConfig_windowsBasic(rName, domainName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBackupExists(ctx, resourceName, &backup),
					acctest.MatchResourceAttrRegionalARN(resourceName, names.AttrARN, "fsx", regexache.MustCompile(`backup/.+`)),
					acctest.CheckResourceAttrAccountID(resourceName, names.AttrOwnerID),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct1),
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
	ctx := acctest.Context(t)
	var backup awstypes.Backup
	resourceName := "aws_fsx_backup.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckPartitionHasService(t, names.FSxEndpointID) },
		ErrorCheck:               acctest.ErrorCheck(t, names.FSxServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckBackupDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccBackupConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBackupExists(ctx, resourceName, &backup),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tffsx.ResourceBackup(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccFSxBackup_Disappears_filesystem(t *testing.T) {
	ctx := acctest.Context(t)
	var backup awstypes.Backup
	resourceName := "aws_fsx_backup.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckPartitionHasService(t, names.FSxEndpointID) },
		ErrorCheck:               acctest.ErrorCheck(t, names.FSxServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckBackupDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccBackupConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBackupExists(ctx, resourceName, &backup),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tffsx.ResourceLustreFileSystem(), "aws_fsx_lustre_file_system.test"),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccFSxBackup_tags(t *testing.T) {
	ctx := acctest.Context(t)
	var backup awstypes.Backup
	resourceName := "aws_fsx_backup.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckPartitionHasService(t, names.FSxEndpointID) },
		ErrorCheck:               acctest.ErrorCheck(t, names.FSxServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckBackupDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccBackupConfig_tags1(rName, acctest.CtKey1, acctest.CtValue1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBackupExists(ctx, resourceName, &backup),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey1, acctest.CtValue1),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccBackupConfig_tags2(rName, acctest.CtKey1, acctest.CtValue1Updated, acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBackupExists(ctx, resourceName, &backup),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct2),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey1, acctest.CtValue1Updated),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey2, acctest.CtValue2),
				),
			},
			{
				Config: testAccBackupConfig_tags1(rName, acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBackupExists(ctx, resourceName, &backup),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey2, acctest.CtValue2),
				),
			},
		},
	})
}

func TestAccFSxBackup_implicitTags(t *testing.T) {
	ctx := acctest.Context(t)
	var backup awstypes.Backup
	resourceName := "aws_fsx_backup.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckPartitionHasService(t, names.FSxEndpointID) },
		ErrorCheck:               acctest.ErrorCheck(t, names.FSxServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckBackupDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccBackupConfig_implicitTags(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBackupExists(ctx, resourceName, &backup),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "tags.Name", rName),
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

func testAccCheckBackupExists(ctx context.Context, n string, v *awstypes.Backup) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).FSxClient(ctx)

		output, err := tffsx.FindBackupByID(ctx, conn, rs.Primary.ID)

		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

func testAccCheckBackupDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).FSxClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_fsx_backup" {
				continue
			}

			_, err := tffsx.FindBackupByID(ctx, conn, rs.Primary.ID)

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
}

func testAccBackupConfig_baseLustre(rName string) string {
	return acctest.ConfigCompose(testAccLustreFileSystemConfig_base(rName), fmt.Sprintf(`
resource "aws_fsx_lustre_file_system" "test" {
  storage_capacity            = 1200
  subnet_ids                  = aws_subnet.test[*].id
  deployment_type             = "PERSISTENT_1"
  per_unit_storage_throughput = 50

  tags = {
    Name = %[1]q
  }
}
`, rName))
}

func testAccBackupConfig_baseONTAP(rName string, vName string) string {
	return acctest.ConfigCompose(testAccONTAPFileSystemConfig_base(rName), fmt.Sprintf(`
resource "aws_fsx_ontap_file_system" "test" {
  storage_capacity    = 1024
  subnet_ids          = aws_subnet.test[*].id
  deployment_type     = "MULTI_AZ_1"
  throughput_capacity = 128
  preferred_subnet_id = aws_subnet.test[0].id

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

func testAccBackupConfig_baseOpenZFS(rName string) string {
	return acctest.ConfigCompose(testAccOpenZFSFileSystemConfig_baseSingleAZ(rName), fmt.Sprintf(`
resource "aws_fsx_openzfs_file_system" "test" {
  storage_capacity    = 64
  subnet_ids          = aws_subnet.test[*].id
  deployment_type     = "SINGLE_AZ_1"
  throughput_capacity = 64
  skip_final_backup   = true

  tags = {
    Name = %[1]q
  }
}
`, rName))
}

func testAccBackupConfig_baseWindows(rName, domain string) string {
	return acctest.ConfigCompose(testAccWindowsFileSystemConfig_base(rName, domain), fmt.Sprintf(`
resource "aws_fsx_windows_file_system" "test" {
  active_directory_id             = aws_directory_service_directory.test.id
  automatic_backup_retention_days = 0
  skip_final_backup               = true
  storage_capacity                = 32
  subnet_ids                      = [aws_subnet.test[0].id]
  throughput_capacity             = 8

  tags = {
    Name = %[1]q
  }
}
`, rName))
}

func testAccBackupConfig_basic(rName string) string {
	return acctest.ConfigCompose(testAccBackupConfig_baseLustre(rName), fmt.Sprintf(`
resource "aws_fsx_backup" "test" {
  file_system_id = aws_fsx_lustre_file_system.test.id

  tags = {
    Name = %[1]q
  }
}
`, rName))
}

func testAccBackupConfig_ontapBasic(rName string, vName string) string {
	return acctest.ConfigCompose(testAccBackupConfig_baseONTAP(rName, vName), fmt.Sprintf(`
resource "aws_fsx_backup" "test" {
  volume_id = aws_fsx_ontap_volume.test.id

  tags = {
    Name = %[1]q
  }
}
`, rName))
}

func testAccBackupConfig_openZFSBasic(rName string) string {
	return acctest.ConfigCompose(testAccBackupConfig_baseOpenZFS(rName), fmt.Sprintf(`
resource "aws_fsx_backup" "test" {
  file_system_id = aws_fsx_openzfs_file_system.test.id

  tags = {
    Name = %[1]q
  }
}
`, rName))
}

func testAccBackupConfig_windowsBasic(rName, domain string) string {
	return acctest.ConfigCompose(testAccBackupConfig_baseWindows(rName, domain), fmt.Sprintf(`
resource "aws_fsx_backup" "test" {
  file_system_id = aws_fsx_windows_file_system.test.id

  tags = {
    Name = %[1]q
  }
}
`, rName))
}

func testAccBackupConfig_tags1(rName string, tagKey1, tagValue1 string) string {
	return acctest.ConfigCompose(testAccBackupConfig_baseLustre(rName), fmt.Sprintf(`
resource "aws_fsx_backup" "test" {
  file_system_id = aws_fsx_lustre_file_system.test.id

  tags = {
    %[1]q = %[2]q
  }
}
`, tagKey1, tagValue1))
}

func testAccBackupConfig_tags2(rName string, tagKey1, tagValue1, tagKey2, tagValue2 string) string {
	return acctest.ConfigCompose(testAccBackupConfig_baseLustre(rName), fmt.Sprintf(`
resource "aws_fsx_backup" "test" {
  file_system_id = aws_fsx_lustre_file_system.test.id

  tags = {
    %[1]q = %[2]q
    %[3]q = %[4]q
  }
}
`, tagKey1, tagValue1, tagKey2, tagValue2))
}

func testAccBackupConfig_implicitTags(rName string) string {
	return acctest.ConfigCompose(testAccLustreFileSystemConfig_base(rName), fmt.Sprintf(`
resource "aws_fsx_lustre_file_system" "test" {
  storage_capacity            = 1200
  subnet_ids                  = aws_subnet.test[*].id
  deployment_type             = "PERSISTENT_1"
  per_unit_storage_throughput = 50
  copy_tags_to_backups        = true

  tags = {
    Name = %[1]q
  }
}

resource "aws_fsx_backup" "test" {
  file_system_id = aws_fsx_lustre_file_system.test.id

  lifecycle {
    ignore_changes = [tags]
  }
}
`, rName))
}
