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

func TestAccFSxOntapVolume_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var volume fsx.Volume
	resourceName := "aws_fsx_ontap_volume.test"
	rName := fmt.Sprintf("tf_acc_test_%d", sdkacctest.RandInt())

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckPartitionHasService(t, fsx.EndpointsID) },
		ErrorCheck:               acctest.ErrorCheck(t, fsx.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckOntapVolumeDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccONTAPVolumeConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckOntapVolumeExists(ctx, resourceName, &volume),
					acctest.MatchResourceAttrRegionalARN(resourceName, "arn", "fsx", regexp.MustCompile(`volume/fs-.+/fsvol-.+`)),
					resource.TestCheckResourceAttrSet(resourceName, "file_system_id"),
					resource.TestCheckResourceAttr(resourceName, "junction_path", fmt.Sprintf("/%[1]s", rName)),
					resource.TestCheckResourceAttr(resourceName, "ontap_volume_type", "RW"),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "security_style", ""),
					resource.TestCheckResourceAttr(resourceName, "size_in_megabytes", "1024"),
					resource.TestCheckResourceAttr(resourceName, "skip_final_backup", "false"),
					resource.TestCheckResourceAttr(resourceName, "storage_efficiency_enabled", "true"),
					resource.TestCheckResourceAttrSet(resourceName, "storage_virtual_machine_id"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
					resource.TestCheckResourceAttrSet(resourceName, "uuid"),
					resource.TestCheckResourceAttr(resourceName, "volume_type", "ONTAP"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"skip_final_backup"},
			},
		},
	})
}

func TestAccFSxOntapVolume_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	var volume fsx.Volume
	resourceName := "aws_fsx_ontap_volume.test"
	rName := fmt.Sprintf("tf_acc_test_%d", sdkacctest.RandInt())

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckPartitionHasService(t, fsx.EndpointsID) },
		ErrorCheck:               acctest.ErrorCheck(t, fsx.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckOntapVolumeDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccONTAPVolumeConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckOntapVolumeExists(ctx, resourceName, &volume),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tffsx.ResourceOntapVolume(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccFSxOntapVolume_name(t *testing.T) {
	ctx := acctest.Context(t)
	var volume1, volume2 fsx.Volume
	resourceName := "aws_fsx_ontap_volume.test"
	rName := fmt.Sprintf("tf_acc_test_%d", sdkacctest.RandInt())
	rName2 := fmt.Sprintf("tf_acc_test_%d", sdkacctest.RandInt())

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckPartitionHasService(t, fsx.EndpointsID) },
		ErrorCheck:               acctest.ErrorCheck(t, fsx.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckOntapVolumeDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccONTAPVolumeConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckOntapVolumeExists(ctx, resourceName, &volume1),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"skip_final_backup"},
			},
			{
				Config: testAccONTAPVolumeConfig_basic(rName2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckOntapVolumeExists(ctx, resourceName, &volume2),
					testAccCheckOntapVolumeRecreated(&volume1, &volume2),
					resource.TestCheckResourceAttr(resourceName, "name", rName2),
				),
			},
		},
	})
}

func TestAccFSxOntapVolume_junctionPath(t *testing.T) {
	ctx := acctest.Context(t)
	var volume1, volume2 fsx.Volume
	resourceName := "aws_fsx_ontap_volume.test"
	rName := fmt.Sprintf("tf_acc_test_%d", sdkacctest.RandInt())
	jPath1 := "/path1"
	jPath2 := "/path2"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckPartitionHasService(t, fsx.EndpointsID) },
		ErrorCheck:               acctest.ErrorCheck(t, fsx.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckOntapVolumeDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccONTAPVolumeConfig_junctionPath(rName, jPath1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckOntapVolumeExists(ctx, resourceName, &volume1),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "junction_path", jPath1),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"skip_final_backup"},
			},
			{
				Config: testAccONTAPVolumeConfig_junctionPath(rName, jPath2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckOntapVolumeExists(ctx, resourceName, &volume2),
					testAccCheckOntapVolumeNotRecreated(&volume1, &volume2),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "junction_path", jPath2),
				),
			},
		},
	})
}

func TestAccFSxOntapVolume_ontapVolumeType(t *testing.T) {
	ctx := acctest.Context(t)
	var volume fsx.Volume
	resourceName := "aws_fsx_ontap_volume.test"
	rName := fmt.Sprintf("tf_acc_test_%d", sdkacctest.RandInt())

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckPartitionHasService(t, fsx.EndpointsID) },
		ErrorCheck:               acctest.ErrorCheck(t, fsx.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckOntapVolumeDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccONTAPVolumeConfig_ontapVolumeTypeDP(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckOntapVolumeExists(ctx, resourceName, &volume),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "ontap_volume_type", "DP"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"skip_final_backup"},
			},
		},
	})
}

func TestAccFSxOntapVolume_securityStyle(t *testing.T) {
	ctx := acctest.Context(t)
	var volume1, volume2, volume3 fsx.Volume
	resourceName := "aws_fsx_ontap_volume.test"
	rName := fmt.Sprintf("tf_acc_test_%d", sdkacctest.RandInt())

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckPartitionHasService(t, fsx.EndpointsID) },
		ErrorCheck:               acctest.ErrorCheck(t, fsx.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckOntapVolumeDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccONTAPVolumeConfig_securityStyle(rName, "UNIX"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckOntapVolumeExists(ctx, resourceName, &volume1),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "security_style", "UNIX"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"skip_final_backup"},
			},
			{
				Config: testAccONTAPVolumeConfig_securityStyle(rName, "NTFS"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckOntapVolumeExists(ctx, resourceName, &volume2),
					testAccCheckOntapVolumeNotRecreated(&volume1, &volume2),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "security_style", "NTFS"),
				),
			},
			{
				Config: testAccONTAPVolumeConfig_securityStyle(rName, "MIXED"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckOntapVolumeExists(ctx, resourceName, &volume3),
					testAccCheckOntapVolumeNotRecreated(&volume1, &volume3),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "security_style", "MIXED"),
				),
			},
		},
	})
}

func TestAccFSxOntapVolume_size(t *testing.T) {
	ctx := acctest.Context(t)
	var volume1, volume2 fsx.Volume
	resourceName := "aws_fsx_ontap_volume.test"
	rName := fmt.Sprintf("tf_acc_test_%d", sdkacctest.RandInt())
	size1 := 1024
	size2 := 2048

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckPartitionHasService(t, fsx.EndpointsID) },
		ErrorCheck:               acctest.ErrorCheck(t, fsx.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckOntapVolumeDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccONTAPVolumeConfig_size(rName, size1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckOntapVolumeExists(ctx, resourceName, &volume1),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "size_in_megabytes", fmt.Sprint(size1)),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"skip_final_backup"},
			},
			{
				Config: testAccONTAPVolumeConfig_size(rName, size2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckOntapVolumeExists(ctx, resourceName, &volume2),
					testAccCheckOntapVolumeNotRecreated(&volume1, &volume2),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "size_in_megabytes", fmt.Sprint(size2)),
				),
			},
		},
	})
}

func TestAccFSxOntapVolume_storageEfficiency(t *testing.T) {
	ctx := acctest.Context(t)
	var volume1, volume2 fsx.Volume
	resourceName := "aws_fsx_ontap_volume.test"
	rName := fmt.Sprintf("tf_acc_test_%d", sdkacctest.RandInt())

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckPartitionHasService(t, fsx.EndpointsID) },
		ErrorCheck:               acctest.ErrorCheck(t, fsx.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckOntapVolumeDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccONTAPVolumeConfig_storageEfficiency(rName, true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckOntapVolumeExists(ctx, resourceName, &volume1),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "storage_efficiency_enabled", "true"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"skip_final_backup"},
			},
			{
				Config: testAccONTAPVolumeConfig_storageEfficiency(rName, false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckOntapVolumeExists(ctx, resourceName, &volume2),
					testAccCheckOntapVolumeNotRecreated(&volume1, &volume2),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "storage_efficiency_enabled", "false"),
				),
			},
		},
	})
}

func TestAccFSxOntapVolume_tags(t *testing.T) {
	ctx := acctest.Context(t)
	var volume1, volume2, volume3 fsx.Volume
	resourceName := "aws_fsx_ontap_volume.test"
	rName := fmt.Sprintf("tf_acc_test_%d", sdkacctest.RandInt())

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckPartitionHasService(t, fsx.EndpointsID) },
		ErrorCheck:               acctest.ErrorCheck(t, fsx.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckOntapVolumeDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccONTAPVolumeConfig_tags1(rName, "key1", "value1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckOntapVolumeExists(ctx, resourceName, &volume1),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"skip_final_backup"},
			},
			{
				Config: testAccONTAPVolumeConfig_tags2(rName, "key1", "value1updated", "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckOntapVolumeExists(ctx, resourceName, &volume2),
					testAccCheckOntapVolumeNotRecreated(&volume1, &volume2),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1updated"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
			{
				Config: testAccONTAPVolumeConfig_tags1(rName, "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckOntapVolumeExists(ctx, resourceName, &volume3),
					testAccCheckOntapVolumeNotRecreated(&volume2, &volume3),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
		},
	})
}

func TestAccFSxOntapVolume_tieringPolicy(t *testing.T) {
	ctx := acctest.Context(t)
	var volume1, volume2, volume3, volume4 fsx.Volume
	resourceName := "aws_fsx_ontap_volume.test"
	rName := fmt.Sprintf("tf_acc_test_%d", sdkacctest.RandInt())

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckPartitionHasService(t, fsx.EndpointsID) },
		ErrorCheck:               acctest.ErrorCheck(t, fsx.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckOntapVolumeDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccONTAPVolumeConfig_tieringPolicyNoCooling(rName, "NONE"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckOntapVolumeExists(ctx, resourceName, &volume1),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "tiering_policy.0.name", "NONE"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"skip_final_backup"},
			},
			{
				Config: testAccONTAPVolumeConfig_tieringPolicy(rName, "SNAPSHOT_ONLY", 10),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckOntapVolumeExists(ctx, resourceName, &volume2),
					testAccCheckOntapVolumeNotRecreated(&volume1, &volume2),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "tiering_policy.0.name", "SNAPSHOT_ONLY"),
					resource.TestCheckResourceAttr(resourceName, "tiering_policy.0.cooling_period", "10"),
				),
			},
			{
				Config: testAccONTAPVolumeConfig_tieringPolicy(rName, "AUTO", 60),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckOntapVolumeExists(ctx, resourceName, &volume3),
					testAccCheckOntapVolumeNotRecreated(&volume1, &volume3),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "tiering_policy.0.name", "AUTO"),
					resource.TestCheckResourceAttr(resourceName, "tiering_policy.0.cooling_period", "60"),
				),
			},
			{
				Config: testAccONTAPVolumeConfig_tieringPolicyNoCooling(rName, "ALL"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckOntapVolumeExists(ctx, resourceName, &volume4),
					testAccCheckOntapVolumeNotRecreated(&volume1, &volume4),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "tiering_policy.0.name", "ALL"),
				),
			},
		},
	})
}

func testAccCheckOntapVolumeExists(ctx context.Context, n string, v *fsx.Volume) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).FSxConn(ctx)

		output, err := tffsx.FindVolumeByID(ctx, conn, rs.Primary.ID)

		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

func testAccCheckOntapVolumeDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).FSxConn(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_fsx_ontap_volume" {
				continue
			}

			volume, err := tffsx.FindVolumeByID(ctx, conn, rs.Primary.ID)

			if tfresource.NotFound(err) {
				continue
			}

			if volume != nil {
				return fmt.Errorf("FSx ONTAP Volume (%s) still exists", rs.Primary.ID)
			}
		}
		return nil
	}
}

func testAccCheckOntapVolumeNotRecreated(i, j *fsx.Volume) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if aws.StringValue(i.VolumeId) != aws.StringValue(j.VolumeId) {
			return fmt.Errorf("FSx ONTAP Volume (%s) recreated", aws.StringValue(i.VolumeId))
		}

		return nil
	}
}

func testAccCheckOntapVolumeRecreated(i, j *fsx.Volume) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if aws.StringValue(i.VolumeId) == aws.StringValue(j.VolumeId) {
			return fmt.Errorf("FSx ONTAP Volume (%s) not recreated", aws.StringValue(i.VolumeId))
		}

		return nil
	}
}

func testAccOntapVolumeConfig_base(rName string) string {
	return acctest.ConfigCompose(acctest.ConfigVPCWithSubnets(rName, 2), fmt.Sprintf(`
resource "aws_fsx_ontap_file_system" "test" {
  storage_capacity    = 1024
  subnet_ids          = aws_subnet.test[*].id
  deployment_type     = "MULTI_AZ_1"
  throughput_capacity = 512
  preferred_subnet_id = aws_subnet.test[0].id

  tags = {
    Name = %[1]q
  }
}

resource "aws_fsx_ontap_storage_virtual_machine" "test" {
  file_system_id = aws_fsx_ontap_file_system.test.id
  name           = %[1]q
}
`, rName))
}

func testAccONTAPVolumeConfig_basic(rName string) string {
	return acctest.ConfigCompose(testAccOntapVolumeConfig_base(rName), fmt.Sprintf(`
resource "aws_fsx_ontap_volume" "test" {
  name                       = %[1]q
  junction_path              = "/%[1]s"
  size_in_megabytes          = 1024
  storage_efficiency_enabled = true
  storage_virtual_machine_id = aws_fsx_ontap_storage_virtual_machine.test.id
}
`, rName))
}

func testAccONTAPVolumeConfig_junctionPath(rName string, junctionPath string) string {
	return acctest.ConfigCompose(testAccOntapVolumeConfig_base(rName), fmt.Sprintf(`
resource "aws_fsx_ontap_volume" "test" {
  name                       = %[1]q
  junction_path              = %[2]q
  size_in_megabytes          = 1024
  storage_efficiency_enabled = true
  storage_virtual_machine_id = aws_fsx_ontap_storage_virtual_machine.test.id
}
`, rName, junctionPath))
}

func testAccONTAPVolumeConfig_ontapVolumeTypeDP(rName string) string {
	return acctest.ConfigCompose(testAccOntapVolumeConfig_base(rName), fmt.Sprintf(`
resource "aws_fsx_ontap_volume" "test" {
  name                       = %[1]q
  ontap_volume_type          = "DP"
  size_in_megabytes          = 1024
  skip_final_backup          = true
  storage_virtual_machine_id = aws_fsx_ontap_storage_virtual_machine.test.id
}
`, rName))
}

func testAccONTAPVolumeConfig_securityStyle(rName string, securityStyle string) string {
	return acctest.ConfigCompose(testAccOntapVolumeConfig_base(rName), fmt.Sprintf(`
resource "aws_fsx_ontap_volume" "test" {
  name                       = %[1]q
  junction_path              = "/%[1]s"
  size_in_megabytes          = 1024
  security_style             = %[2]q
  storage_efficiency_enabled = true
  storage_virtual_machine_id = aws_fsx_ontap_storage_virtual_machine.test.id
}
`, rName, securityStyle))
}

func testAccONTAPVolumeConfig_size(rName string, size int) string {
	return acctest.ConfigCompose(testAccOntapVolumeConfig_base(rName), fmt.Sprintf(`
resource "aws_fsx_ontap_volume" "test" {
  name                       = %[1]q
  junction_path              = "/%[1]s"
  size_in_megabytes          = %[2]d
  storage_efficiency_enabled = true
  storage_virtual_machine_id = aws_fsx_ontap_storage_virtual_machine.test.id
}
`, rName, size))
}

func testAccONTAPVolumeConfig_storageEfficiency(rName string, storageEfficiencyEnabled bool) string {
	return acctest.ConfigCompose(testAccOntapVolumeConfig_base(rName), fmt.Sprintf(`
resource "aws_fsx_ontap_volume" "test" {
  name                       = %[1]q
  junction_path              = "/%[1]s"
  size_in_megabytes          = 1024
  storage_efficiency_enabled = %[2]t
  storage_virtual_machine_id = aws_fsx_ontap_storage_virtual_machine.test.id
}
`, rName, storageEfficiencyEnabled))
}

func testAccONTAPVolumeConfig_tieringPolicy(rName string, policy string, coolingPeriod int) string {
	return acctest.ConfigCompose(testAccOntapVolumeConfig_base(rName), fmt.Sprintf(`
resource "aws_fsx_ontap_volume" "test" {
  name                       = %[1]q
  junction_path              = "/%[1]s"
  size_in_megabytes          = 1024
  storage_efficiency_enabled = true
  storage_virtual_machine_id = aws_fsx_ontap_storage_virtual_machine.test.id

  tiering_policy {
    name           = %[2]q
    cooling_period = %[3]d
  }
}
`, rName, policy, coolingPeriod))
}

func testAccONTAPVolumeConfig_tieringPolicyNoCooling(rName string, policy string) string {
	return acctest.ConfigCompose(testAccOntapVolumeConfig_base(rName), fmt.Sprintf(`
resource "aws_fsx_ontap_volume" "test" {
  name                       = %[1]q
  junction_path              = "/%[1]s"
  size_in_megabytes          = 1024
  storage_efficiency_enabled = true
  storage_virtual_machine_id = aws_fsx_ontap_storage_virtual_machine.test.id

  tiering_policy {
    name = %[2]q
  }
}
`, rName, policy))
}

func testAccONTAPVolumeConfig_tags1(rName, tagKey1, tagValue1 string) string {
	return acctest.ConfigCompose(testAccOntapVolumeConfig_base(rName), fmt.Sprintf(`
resource "aws_fsx_ontap_volume" "test" {
  name                       = %[1]q
  junction_path              = "/%[1]s"
  size_in_megabytes          = 1024
  storage_efficiency_enabled = true
  storage_virtual_machine_id = aws_fsx_ontap_storage_virtual_machine.test.id

  tags = {
    %[2]q = %[3]q
  }
}
`, rName, tagKey1, tagValue1))
}

func testAccONTAPVolumeConfig_tags2(rName, tagKey1, tagValue1, tagKey2, tagValue2 string) string {
	return acctest.ConfigCompose(testAccOntapVolumeConfig_base(rName), fmt.Sprintf(`
resource "aws_fsx_ontap_volume" "test" {
  name                       = %[1]q
  junction_path              = "/%[1]s"
  size_in_megabytes          = 1024
  storage_efficiency_enabled = true
  storage_virtual_machine_id = aws_fsx_ontap_storage_virtual_machine.test.id

  tags = {
    %[2]q = %[3]q
    %[4]q = %[5]q
  }
}
`, rName, tagKey1, tagValue1, tagKey2, tagValue2))
}
