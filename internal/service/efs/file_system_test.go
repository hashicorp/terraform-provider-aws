// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package efs_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/YakDriver/regexache"
	awstypes "github.com/aws/aws-sdk-go-v2/service/efs/types"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfefs "github.com/hashicorp/terraform-provider-aws/internal/service/efs"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccEFSFileSystem_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var desc awstypes.FileSystemDescription
	resourceName := "aws_efs_file_system.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EFSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckFileSystemDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccFileSystemConfig_basic,
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckFileSystem(ctx, resourceName, &desc),
					acctest.MatchResourceAttrRegionalARN(resourceName, names.AttrARN, "elasticfilesystem", regexache.MustCompile(`file-system/fs-.+`)),
					resource.TestCheckResourceAttrSet(resourceName, "creation_token"),
					acctest.MatchResourceAttrRegionalHostname(resourceName, names.AttrDNSName, "efs", regexache.MustCompile(`fs-[^.]+`)),
					resource.TestCheckResourceAttr(resourceName, names.AttrEncrypted, acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "lifecycle_policy.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, ""),
					resource.TestCheckResourceAttr(resourceName, "number_of_mount_targets", acctest.Ct0),
					acctest.MatchResourceAttrAccountID(resourceName, names.AttrOwnerID),
					resource.TestCheckResourceAttr(resourceName, "performance_mode", "generalPurpose"),
					resource.TestCheckResourceAttr(resourceName, "protection.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "protection.0.replication_overwrite", "ENABLED"),
					resource.TestCheckResourceAttr(resourceName, "size_in_bytes.#", acctest.Ct1),
					resource.TestCheckResourceAttrSet(resourceName, "size_in_bytes.0.value"),
					resource.TestCheckResourceAttrSet(resourceName, "size_in_bytes.0.value_in_ia"),
					resource.TestCheckResourceAttrSet(resourceName, "size_in_bytes.0.value_in_standard"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "throughput_mode", string(awstypes.ThroughputModeBursting)),
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

func TestAccEFSFileSystem_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	var desc awstypes.FileSystemDescription
	resourceName := "aws_efs_file_system.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EFSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckFileSystemDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccFileSystemConfig_basic,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFileSystem(ctx, resourceName, &desc),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfefs.ResourceFileSystem(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccEFSFileSystem_performanceMode(t *testing.T) {
	ctx := acctest.Context(t)
	var desc awstypes.FileSystemDescription
	resourceName := "aws_efs_file_system.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EFSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckFileSystemDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccFileSystemConfig_performanceMode,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFileSystem(ctx, resourceName, &desc),
					resource.TestCheckResourceAttr(resourceName, "performance_mode", "maxIO"),
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

func TestAccEFSFileSystem_protection(t *testing.T) {
	ctx := acctest.Context(t)
	var desc awstypes.FileSystemDescription
	resourceName := "aws_efs_file_system.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EFSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckFileSystemDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccFileSystemConfig_protection("DISABLED"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFileSystem(ctx, resourceName, &desc),
					resource.TestCheckResourceAttr(resourceName, "protection.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "protection.0.replication_overwrite", "DISABLED"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccFileSystemConfig_protection("ENABLED"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFileSystem(ctx, resourceName, &desc),
					resource.TestCheckResourceAttr(resourceName, "protection.0.replication_overwrite", "ENABLED"),
				),
			},
		},
	})
}

func TestAccEFSFileSystem_availabilityZoneName(t *testing.T) {
	ctx := acctest.Context(t)
	var desc awstypes.FileSystemDescription
	resourceName := "aws_efs_file_system.test"
	rName := sdkacctest.RandomWithPrefix("tf-acc")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EFSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckFileSystemDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccFileSystemConfig_availabilityZoneName(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFileSystem(ctx, resourceName, &desc),
					resource.TestCheckResourceAttrPair(resourceName, "availability_zone_id", "data.aws_availability_zones.available", "zone_ids.0"),
					resource.TestCheckResourceAttrPair(resourceName, "availability_zone_name", "data.aws_availability_zones.available", "names.0"),
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

func TestAccEFSFileSystem_tags(t *testing.T) {
	ctx := acctest.Context(t)
	var desc awstypes.FileSystemDescription
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_efs_file_system.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EFSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckFileSystemDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccFileSystemConfig_tags1(rName, acctest.CtKey1, acctest.CtValue1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFileSystem(ctx, resourceName, &desc),
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
				Config: testAccFileSystemConfig_tags2(rName, acctest.CtKey1, acctest.CtValue1Updated, acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFileSystem(ctx, resourceName, &desc),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct2),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey1, acctest.CtValue1Updated),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey2, acctest.CtValue2),
				),
			},
			{
				Config: testAccFileSystemConfig_tags1(rName, acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFileSystem(ctx, resourceName, &desc),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey2, acctest.CtValue2),
				),
			},
			{
				Config: testAccFileSystemConfig_pagedTags(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFileSystem(ctx, resourceName, &desc),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct10),
				),
			},
			{
				Config: testAccFileSystemConfig_maxTags(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFileSystem(ctx, resourceName, &desc),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, "50"),
					resource.TestCheckResourceAttr(resourceName, "tags.Name", rName),
					resource.TestCheckResourceAttr(resourceName, "tags.Another", "tag"),
					resource.TestCheckResourceAttr(resourceName, "tags.Tag45", "TestTagValue"),
				),
			},
		},
	})
}

func TestAccEFSFileSystem_kmsKey(t *testing.T) {
	ctx := acctest.Context(t)
	var desc awstypes.FileSystemDescription
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	kmsKeyResourceName := "aws_kms_key.test"
	resourceName := "aws_efs_file_system.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EFSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckFileSystemDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccFileSystemConfig_kmsKey(rName, true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFileSystem(ctx, resourceName, &desc),
					resource.TestCheckResourceAttr(resourceName, names.AttrEncrypted, acctest.CtTrue),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrKMSKeyID, kmsKeyResourceName, names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
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

func TestAccEFSFileSystem_kmsWithoutEncryption(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EFSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckFileSystemDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config:      testAccFileSystemConfig_kmsKey(rName, false),
				ExpectError: regexache.MustCompile(`encrypted must be set to true when kms_key_id is specified`),
			},
		},
	})
}

func TestAccEFSFileSystem_provisionedThroughputInMibps(t *testing.T) {
	ctx := acctest.Context(t)
	var desc awstypes.FileSystemDescription
	resourceName := "aws_efs_file_system.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EFSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckFileSystemDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccFileSystemConfig_provisionedThroughputInMibps(1.0),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFileSystem(ctx, resourceName, &desc),
					resource.TestCheckResourceAttr(resourceName, "provisioned_throughput_in_mibps", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "throughput_mode", string(awstypes.ThroughputModeProvisioned)),
				),
			},
			{
				Config: testAccFileSystemConfig_provisionedThroughputInMibps(2.0),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFileSystem(ctx, resourceName, &desc),
					resource.TestCheckResourceAttr(resourceName, "provisioned_throughput_in_mibps", acctest.Ct2),
					resource.TestCheckResourceAttr(resourceName, "throughput_mode", string(awstypes.ThroughputModeProvisioned)),
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

func TestAccEFSFileSystem_throughputMode(t *testing.T) {
	ctx := acctest.Context(t)
	var desc awstypes.FileSystemDescription
	resourceName := "aws_efs_file_system.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EFSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckFileSystemDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccFileSystemConfig_provisionedThroughputInMibps(1.0),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFileSystem(ctx, resourceName, &desc),
					resource.TestCheckResourceAttr(resourceName, "provisioned_throughput_in_mibps", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "throughput_mode", string(awstypes.ThroughputModeProvisioned)),
				),
			},
			{
				Config: testAccFileSystemConfig_throughputMode(string(awstypes.ThroughputModeBursting)),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFileSystem(ctx, resourceName, &desc),
					resource.TestCheckResourceAttr(resourceName, "provisioned_throughput_in_mibps", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "throughput_mode", string(awstypes.ThroughputModeBursting)),
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

func TestAccEFSFileSystem_lifecyclePolicy(t *testing.T) {
	ctx := acctest.Context(t)
	var desc awstypes.FileSystemDescription
	resourceName := "aws_efs_file_system.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EFSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckFileSystemDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccFileSystemConfig_lifecyclePolicy(
					"transition_to_ia",
					"invalid_value",
				),
				ExpectError: regexache.MustCompile(`got invalid_value`),
			},
			{
				Config: testAccFileSystemConfig_lifecyclePolicy(
					"transition_to_ia",
					string(awstypes.TransitionToIARulesAfter30Days),
				),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFileSystem(ctx, resourceName, &desc),
					resource.TestCheckResourceAttr(resourceName, "lifecycle_policy.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "lifecycle_policy.0.transition_to_archive", ""),
					resource.TestCheckResourceAttr(resourceName, "lifecycle_policy.0.transition_to_ia", string(awstypes.TransitionToIARulesAfter30Days)),
					resource.TestCheckResourceAttr(resourceName, "lifecycle_policy.0.transition_to_primary_storage_class", ""),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccFileSystemConfig_lifecyclePolicy(
					"transition_to_primary_storage_class",
					string(awstypes.TransitionToPrimaryStorageClassRulesAfter1Access),
				),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFileSystem(ctx, resourceName, &desc),
					resource.TestCheckResourceAttr(resourceName, "lifecycle_policy.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "lifecycle_policy.0.transition_to_archive", ""),
					resource.TestCheckResourceAttr(resourceName, "lifecycle_policy.0.transition_to_ia", ""),
					resource.TestCheckResourceAttr(resourceName, "lifecycle_policy.0.transition_to_primary_storage_class", string(awstypes.TransitionToPrimaryStorageClassRulesAfter1Access)),
				),
			},
			{
				Config: testAccFileSystemConfig_removedLifecyclePolicy,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFileSystem(ctx, resourceName, &desc),
					resource.TestCheckResourceAttr(resourceName, "lifecycle_policy.#", acctest.Ct0),
				),
			},
			{
				Config: testAccFileSystemConfig_lifecyclePolicyMulti(
					"transition_to_primary_storage_class",
					string(awstypes.TransitionToPrimaryStorageClassRulesAfter1Access),
					"transition_to_ia",
					string(awstypes.TransitionToIARulesAfter30Days),
				),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFileSystem(ctx, resourceName, &desc),
					resource.TestCheckResourceAttr(resourceName, "lifecycle_policy.#", acctest.Ct2),
					resource.TestCheckResourceAttr(resourceName, "lifecycle_policy.0.transition_to_archive", ""),
					resource.TestCheckResourceAttr(resourceName, "lifecycle_policy.0.transition_to_ia", ""),
					resource.TestCheckResourceAttr(resourceName, "lifecycle_policy.0.transition_to_primary_storage_class", string(awstypes.TransitionToPrimaryStorageClassRulesAfter1Access)),
					resource.TestCheckResourceAttr(resourceName, "lifecycle_policy.1.transition_to_archive", ""),
					resource.TestCheckResourceAttr(resourceName, "lifecycle_policy.1.transition_to_ia", string(awstypes.TransitionToIARulesAfter30Days)),
					resource.TestCheckResourceAttr(resourceName, "lifecycle_policy.1.transition_to_primary_storage_class", ""),
				),
			},
			{
				Config: testAccFileSystemConfig_lifecyclePolicyAll(
					"transition_to_primary_storage_class",
					string(awstypes.TransitionToPrimaryStorageClassRulesAfter1Access),
					"transition_to_ia",
					string(awstypes.TransitionToIARulesAfter30Days),
					"transition_to_archive",
					string(awstypes.TransitionToArchiveRulesAfter60Days),
				),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFileSystem(ctx, resourceName, &desc),
					resource.TestCheckResourceAttr(resourceName, "lifecycle_policy.#", acctest.Ct3),
					resource.TestCheckResourceAttr(resourceName, "lifecycle_policy.0.transition_to_archive", ""),
					resource.TestCheckResourceAttr(resourceName, "lifecycle_policy.0.transition_to_ia", ""),
					resource.TestCheckResourceAttr(resourceName, "lifecycle_policy.0.transition_to_primary_storage_class", string(awstypes.TransitionToPrimaryStorageClassRulesAfter1Access)),
					resource.TestCheckResourceAttr(resourceName, "lifecycle_policy.1.transition_to_archive", ""),
					resource.TestCheckResourceAttr(resourceName, "lifecycle_policy.1.transition_to_ia", string(awstypes.TransitionToIARulesAfter30Days)),
					resource.TestCheckResourceAttr(resourceName, "lifecycle_policy.1.transition_to_primary_storage_class", ""),
					resource.TestCheckResourceAttr(resourceName, "lifecycle_policy.2.transition_to_archive", string(awstypes.TransitionToArchiveRulesAfter60Days)),
					resource.TestCheckResourceAttr(resourceName, "lifecycle_policy.2.transition_to_ia", ""),
					resource.TestCheckResourceAttr(resourceName, "lifecycle_policy.2.transition_to_primary_storage_class", ""),
				),
			},
			{
				Config: testAccFileSystemConfig_lifecyclePolicyTransitionToArchive(
					"transition_to_ia",
					string(awstypes.TransitionToIARulesAfter30Days),
					"transition_to_archive",
					string(awstypes.TransitionToArchiveRulesAfter60Days),
				),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFileSystem(ctx, resourceName, &desc),
					resource.TestCheckResourceAttr(resourceName, "lifecycle_policy.#", acctest.Ct2),
					resource.TestCheckResourceAttr(resourceName, "lifecycle_policy.0.transition_to_archive", ""),
					resource.TestCheckResourceAttr(resourceName, "lifecycle_policy.0.transition_to_ia", string(awstypes.TransitionToIARulesAfter30Days)),
					resource.TestCheckResourceAttr(resourceName, "lifecycle_policy.0.transition_to_primary_storage_class", ""),
					resource.TestCheckResourceAttr(resourceName, "lifecycle_policy.1.transition_to_archive", string(awstypes.TransitionToArchiveRulesAfter60Days)),
					resource.TestCheckResourceAttr(resourceName, "lifecycle_policy.1.transition_to_ia", ""),
					resource.TestCheckResourceAttr(resourceName, "lifecycle_policy.1.transition_to_primary_storage_class", ""),
				),
			},
		},
	})
}

func testAccCheckFileSystemDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).EFSClient(ctx)
		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_efs_file_system" {
				continue
			}

			_, err := tfefs.FindFileSystemByID(ctx, conn, rs.Primary.ID)

			if tfresource.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("EFS File System %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccCheckFileSystem(ctx context.Context, n string, v *awstypes.FileSystemDescription) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).EFSClient(ctx)

		output, err := tfefs.FindFileSystemByID(ctx, conn, rs.Primary.ID)

		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

const testAccFileSystemConfig_basic = `
resource "aws_efs_file_system" "test" {}
`

const testAccFileSystemConfig_performanceMode = `
resource "aws_efs_file_system" "test" {
  performance_mode = "maxIO"
}
`

func testAccFileSystemConfig_protection(replicationOverwwrite string) string {
	return fmt.Sprintf(`
resource "aws_efs_file_system" "test" {
  protection {
    replication_overwrite = %[1]q
  }
}
`, replicationOverwwrite)
}

func testAccFileSystemConfig_availabilityZoneName(rName string) string {
	return acctest.ConfigCompose(acctest.ConfigAvailableAZsNoOptIn(), fmt.Sprintf(`
resource "aws_efs_file_system" "test" {
  creation_token         = %[1]q
  availability_zone_name = data.aws_availability_zones.available.names[0]
}
`, rName))
}

func testAccFileSystemConfig_tags1(rName, tagKey1, tagValue1 string) string {
	return fmt.Sprintf(`
resource "aws_efs_file_system" "test" {
  creation_token = %[1]q

  tags = {
    %[2]q = %[3]q
  }
}
`, rName, tagKey1, tagValue1)
}

func testAccFileSystemConfig_tags2(rName, tagKey1, tagValue1, tagKey2, tagValue2 string) string {
	return fmt.Sprintf(`
resource "aws_efs_file_system" "test" {
  creation_token = %[1]q

  tags = {
    %[2]q = %[3]q
    %[4]q = %[5]q
  }
}
`, rName, tagKey1, tagValue1, tagKey2, tagValue2)
}

func testAccFileSystemConfig_pagedTags(rName string) string {
	return fmt.Sprintf(`
resource "aws_efs_file_system" "test" {
  tags = {
    Name           = %[1]q
    Another        = "tag"
    Test           = "yes"
    User           = "root"
    Page           = "1"
    Environment    = "prod"
    CostCenter     = "terraform"
    AcceptanceTest = "PagedTags"
    CreationToken  = "radek"
    PerfMode       = "max"
  }
}
`, rName)
}

func testAccFileSystemConfig_maxTags(rName string) string {
	return fmt.Sprintf(`
resource "aws_efs_file_system" "test" {
  tags = {
    Name    = %[1]q
    Another = "tag"
    Tag00   = "TestTagValue"
    Tag01   = "TestTagValue"
    Tag02   = "TestTagValue"
    Tag03   = "TestTagValue"
    Tag04   = "TestTagValue"
    Tag05   = "TestTagValue"
    Tag06   = "TestTagValue"
    Tag07   = "TestTagValue"
    Tag08   = "TestTagValue"
    Tag09   = "TestTagValue"
    Tag10   = "TestTagValue"
    Tag11   = "TestTagValue"
    Tag12   = "TestTagValue"
    Tag13   = "TestTagValue"
    Tag14   = "TestTagValue"
    Tag15   = "TestTagValue"
    Tag16   = "TestTagValue"
    Tag17   = "TestTagValue"
    Tag18   = "TestTagValue"
    Tag19   = "TestTagValue"
    Tag20   = "TestTagValue"
    Tag21   = "TestTagValue"
    Tag22   = "TestTagValue"
    Tag23   = "TestTagValue"
    Tag24   = "TestTagValue"
    Tag25   = "TestTagValue"
    Tag26   = "TestTagValue"
    Tag27   = "TestTagValue"
    Tag28   = "TestTagValue"
    Tag29   = "TestTagValue"
    Tag30   = "TestTagValue"
    Tag31   = "TestTagValue"
    Tag32   = "TestTagValue"
    Tag33   = "TestTagValue"
    Tag34   = "TestTagValue"
    Tag35   = "TestTagValue"
    Tag36   = "TestTagValue"
    Tag37   = "TestTagValue"
    Tag38   = "TestTagValue"
    Tag39   = "TestTagValue"
    Tag40   = "TestTagValue"
    Tag41   = "TestTagValue"
    Tag42   = "TestTagValue"
    Tag43   = "TestTagValue"
    Tag44   = "TestTagValue"
    Tag45   = "TestTagValue"
    Tag46   = "TestTagValue"
    Tag47   = "TestTagValue"
  }
}
`, rName)
}

func testAccFileSystemConfig_kmsKey(rName string, enable bool) string {
	return fmt.Sprintf(`
resource "aws_kms_key" "test" {
  description = %[1]q
}

resource "aws_efs_file_system" "test" {
  encrypted  = %[2]t
  kms_key_id = aws_kms_key.test.arn

  tags = {
    Name = %[1]q
  }
}
`, rName, enable)
}

func testAccFileSystemConfig_throughputMode(throughputMode string) string {
	return fmt.Sprintf(`
resource "aws_efs_file_system" "test" {
  throughput_mode = %[1]q
}
`, throughputMode)
}

func testAccFileSystemConfig_provisionedThroughputInMibps(provisionedThroughputInMibps float64) string {
	return fmt.Sprintf(`
resource "aws_efs_file_system" "test" {
  provisioned_throughput_in_mibps = %[1]f
  throughput_mode                 = "provisioned"
}
`, provisionedThroughputInMibps)
}

func testAccFileSystemConfig_lifecyclePolicy(lpName, lpVal string) string {
	return fmt.Sprintf(`
resource "aws_efs_file_system" "test" {
  lifecycle_policy {
    %[1]s = %[2]q
  }
}
`, lpName, lpVal)
}

func testAccFileSystemConfig_lifecyclePolicyMulti(lpName1, lpVal1, lpName2, lpVal2 string) string {
	return fmt.Sprintf(`
resource "aws_efs_file_system" "test" {
  lifecycle_policy {
    %[1]s = %[2]q
  }

  lifecycle_policy {
    %[3]s = %[4]q
  }
}
`, lpName1, lpVal1, lpName2, lpVal2)
}

func testAccFileSystemConfig_lifecyclePolicyAll(lpName1, lpVal1, lpName2, lpVal2, lpName3, lpVal3 string) string {
	return fmt.Sprintf(`
resource "aws_efs_file_system" "test" {
  throughput_mode = "elastic"

  lifecycle_policy {
    %[1]s = %[2]q
  }

  lifecycle_policy {
    %[3]s = %[4]q
  }

  lifecycle_policy {
    %[5]s = %[6]q
  }
}
`, lpName1, lpVal1, lpName2, lpVal2, lpName3, lpVal3)
}

func testAccFileSystemConfig_lifecyclePolicyTransitionToArchive(lpName1, lpVal1, lpName2, lpVal2 string) string {
	return fmt.Sprintf(`
resource "aws_efs_file_system" "test" {
  throughput_mode = "elastic"

  lifecycle_policy {
    %[1]s = %[2]q
  }

  lifecycle_policy {
    %[3]s = %[4]q
  }
}
`, lpName1, lpVal1, lpName2, lpVal2)
}

const testAccFileSystemConfig_removedLifecyclePolicy = `
resource "aws_efs_file_system" "test" {}
`
