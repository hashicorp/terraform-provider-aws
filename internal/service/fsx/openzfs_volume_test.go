// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package fsx_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/aws"
	awstypes "github.com/aws/aws-sdk-go-v2/service/fsx/types"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	tffsx "github.com/hashicorp/terraform-provider-aws/internal/service/fsx"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccFSxOpenZFSVolume_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var volume awstypes.Volume
	resourceName := "aws_fsx_openzfs_volume.test"
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckPartitionHasService(t, names.FSxEndpointID) },
		ErrorCheck:               acctest.ErrorCheck(t, names.FSxServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckOpenZFSVolumeDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccOpenZFSVolumeConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckOpenZFSVolumeExists(ctx, t, resourceName, &volume),
					acctest.MatchResourceAttrRegionalARN(ctx, resourceName, names.AttrARN, "fsx", regexache.MustCompile(`volume/fs-.+/fsvol-.+`)),
					resource.TestCheckResourceAttr(resourceName, "copy_tags_to_snapshots", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "data_compression_type", "NONE"),
					resource.TestCheckResourceAttr(resourceName, "delete_volume_options.#", "0"),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, "nfs_exports.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "nfs_exports.0.client_configurations.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "nfs_exports.0.client_configurations.0.clients", "*"),
					acctest.CheckResourceAttrGreaterThanValue(resourceName, "nfs_exports.0.client_configurations.0.options.#", 0),
					resource.TestCheckResourceAttrSet(resourceName, "parent_volume_id"),
					resource.TestCheckResourceAttr(resourceName, "read_only", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "record_size_kib", "128"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, "0"),
					resource.TestCheckResourceAttr(resourceName, "user_and_group_quotas.#", "2"),
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

func TestAccFSxOpenZFSVolume_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	var volume awstypes.Volume
	resourceName := "aws_fsx_openzfs_volume.test"
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckPartitionHasService(t, names.FSxEndpointID) },
		ErrorCheck:               acctest.ErrorCheck(t, names.FSxServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckOpenZFSVolumeDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccOpenZFSVolumeConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckOpenZFSVolumeExists(ctx, t, resourceName, &volume),
					acctest.CheckSDKResourceDisappears(ctx, t, tffsx.ResourceOpenZFSVolume(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccFSxOpenZFSVolume_parentVolume(t *testing.T) {
	ctx := acctest.Context(t)
	var volume, volume2 awstypes.Volume
	resourceName := "aws_fsx_openzfs_volume.test"
	resourceName2 := "aws_fsx_openzfs_volume.test2"
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	rName2 := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckPartitionHasService(t, names.FSxEndpointID) },
		ErrorCheck:               acctest.ErrorCheck(t, names.FSxServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckOpenZFSVolumeDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccOpenZFSVolumeConfig_parent(rName, rName2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckOpenZFSVolumeExists(ctx, t, resourceName, &volume),
					testAccCheckOpenZFSVolumeExists(ctx, t, resourceName2, &volume2),
					acctest.MatchResourceAttrRegionalARN(ctx, resourceName, names.AttrARN, "fsx", regexache.MustCompile(`volume/fs-.+/fsvol-.+`)),
					acctest.MatchResourceAttrRegionalARN(ctx, resourceName2, names.AttrARN, "fsx", regexache.MustCompile(`volume/fs-.+/fsvol-.+`)),
					resource.TestCheckResourceAttrPair(resourceName2, "parent_volume_id", resourceName, names.AttrID),
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

func TestAccFSxOpenZFSVolume_tags(t *testing.T) {
	ctx := acctest.Context(t)
	var volume1, volume2, volume3 awstypes.Volume
	resourceName := "aws_fsx_openzfs_volume.test"
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckPartitionHasService(t, names.FSxEndpointID) },
		ErrorCheck:               acctest.ErrorCheck(t, names.FSxServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckOpenZFSVolumeDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccOpenZFSVolumeConfig_tags1(rName, acctest.CtKey1, acctest.CtValue1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckOpenZFSVolumeExists(ctx, t, resourceName, &volume1),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, "1"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey1, acctest.CtValue1),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccOpenZFSVolumeConfig_tags2(rName, acctest.CtKey1, acctest.CtValue1Updated, acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckOpenZFSVolumeExists(ctx, t, resourceName, &volume2),
					testAccCheckOpenZFSVolumeNotRecreated(&volume1, &volume2),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, "2"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey1, acctest.CtValue1Updated),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey2, acctest.CtValue2),
				),
			},
			{
				Config: testAccOpenZFSVolumeConfig_tags1(rName, acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckOpenZFSVolumeExists(ctx, t, resourceName, &volume3),
					testAccCheckOpenZFSVolumeNotRecreated(&volume2, &volume3),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, "1"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey2, acctest.CtValue2),
				),
			},
		},
	})
}

func TestAccFSxOpenZFSVolume_copyTags(t *testing.T) {
	ctx := acctest.Context(t)
	var volume1, volume2 awstypes.Volume
	resourceName := "aws_fsx_openzfs_volume.test"
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckPartitionHasService(t, names.FSxEndpointID) },
		ErrorCheck:               acctest.ErrorCheck(t, names.FSxServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckOpenZFSVolumeDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccOpenZFSVolumeConfig_copyTags(rName, acctest.CtKey1, acctest.CtValue1, acctest.CtTrue),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckOpenZFSVolumeExists(ctx, t, resourceName, &volume1),
					resource.TestCheckResourceAttr(resourceName, "copy_tags_to_snapshots", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "delete_volume_options.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "delete_volume_options.0", "DELETE_CHILD_VOLUMES_AND_SNAPSHOTS"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, "1"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey1, acctest.CtValue1),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"delete_volume_options",
				},
			},
			{
				Config: testAccOpenZFSVolumeConfig_copyTags(rName, acctest.CtKey1, acctest.CtValue1, acctest.CtFalse),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckOpenZFSVolumeExists(ctx, t, resourceName, &volume2),
					testAccCheckOpenZFSVolumeRecreated(&volume1, &volume2),
					resource.TestCheckResourceAttr(resourceName, "copy_tags_to_snapshots", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "delete_volume_options.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "delete_volume_options.0", "DELETE_CHILD_VOLUMES_AND_SNAPSHOTS"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, "1"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey1, acctest.CtValue1),
				),
			},
		},
	})
}

func TestAccFSxOpenZFSVolume_name(t *testing.T) {
	ctx := acctest.Context(t)
	var volume1, volume2 awstypes.Volume
	resourceName := "aws_fsx_openzfs_volume.test"
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	rName2 := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckPartitionHasService(t, names.FSxEndpointID) },
		ErrorCheck:               acctest.ErrorCheck(t, names.FSxServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckOpenZFSVolumeDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccOpenZFSVolumeConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckOpenZFSVolumeExists(ctx, t, resourceName, &volume1),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccOpenZFSVolumeConfig_basic(rName2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckOpenZFSVolumeExists(ctx, t, resourceName, &volume2),
					testAccCheckOpenZFSVolumeNotRecreated(&volume1, &volume2),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName2),
				),
			},
		},
	})
}

func TestAccFSxOpenZFSVolume_dataCompressionType(t *testing.T) {
	ctx := acctest.Context(t)
	var volume1, volume2 awstypes.Volume
	resourceName := "aws_fsx_openzfs_volume.test"
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckPartitionHasService(t, names.FSxEndpointID) },
		ErrorCheck:               acctest.ErrorCheck(t, names.FSxServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckOpenZFSVolumeDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccOpenZFSVolumeConfig_dataCompression(rName, "ZSTD"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckOpenZFSVolumeExists(ctx, t, resourceName, &volume1),
					resource.TestCheckResourceAttr(resourceName, "data_compression_type", "ZSTD"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccOpenZFSVolumeConfig_dataCompression(rName, "NONE"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckOpenZFSVolumeExists(ctx, t, resourceName, &volume2),
					testAccCheckOpenZFSVolumeNotRecreated(&volume1, &volume2),
					resource.TestCheckResourceAttr(resourceName, "data_compression_type", "NONE"),
				),
			},
		},
	})
}

func TestAccFSxOpenZFSVolume_readOnly(t *testing.T) {
	ctx := acctest.Context(t)
	var volume1, volume2 awstypes.Volume
	resourceName := "aws_fsx_openzfs_volume.test"
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckPartitionHasService(t, names.FSxEndpointID) },
		ErrorCheck:               acctest.ErrorCheck(t, names.FSxServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckOpenZFSVolumeDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccOpenZFSVolumeConfig_readOnly(rName, acctest.CtFalse),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckOpenZFSVolumeExists(ctx, t, resourceName, &volume1),
					resource.TestCheckResourceAttr(resourceName, "read_only", acctest.CtFalse),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccOpenZFSVolumeConfig_readOnly(rName, acctest.CtTrue),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckOpenZFSVolumeExists(ctx, t, resourceName, &volume2),
					testAccCheckOpenZFSVolumeNotRecreated(&volume1, &volume2),
					resource.TestCheckResourceAttr(resourceName, "read_only", acctest.CtTrue),
				),
			},
		},
	})
}

func TestAccFSxOpenZFSVolume_recordSizeKib(t *testing.T) {
	ctx := acctest.Context(t)
	var volume1, volume2 awstypes.Volume
	resourceName := "aws_fsx_openzfs_volume.test"
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckPartitionHasService(t, names.FSxEndpointID) },
		ErrorCheck:               acctest.ErrorCheck(t, names.FSxServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckOpenZFSVolumeDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccOpenZFSVolumeConfig_recordSizeKib(rName, 8),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckOpenZFSVolumeExists(ctx, t, resourceName, &volume1),
					resource.TestCheckResourceAttr(resourceName, "record_size_kib", "8"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccOpenZFSVolumeConfig_recordSizeKib(rName, 1024),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckOpenZFSVolumeExists(ctx, t, resourceName, &volume2),
					testAccCheckOpenZFSVolumeNotRecreated(&volume1, &volume2),
					resource.TestCheckResourceAttr(resourceName, "record_size_kib", "1024"),
				),
			},
		},
	})
}

func TestAccFSxOpenZFSVolume_storageCapacity(t *testing.T) {
	ctx := acctest.Context(t)
	var volume1, volume2 awstypes.Volume
	resourceName := "aws_fsx_openzfs_volume.test"
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckPartitionHasService(t, names.FSxEndpointID) },
		ErrorCheck:               acctest.ErrorCheck(t, names.FSxServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckOpenZFSVolumeDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccOpenZFSVolumeConfig_storageCapacity(rName, 30, 20),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckOpenZFSVolumeExists(ctx, t, resourceName, &volume1),
					resource.TestCheckResourceAttr(resourceName, "storage_capacity_quota_gib", "30"),
					resource.TestCheckResourceAttr(resourceName, "storage_capacity_reservation_gib", "20"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccOpenZFSVolumeConfig_storageCapacity(rName, 40, 30),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckOpenZFSVolumeExists(ctx, t, resourceName, &volume2),
					testAccCheckOpenZFSVolumeNotRecreated(&volume1, &volume2),
					resource.TestCheckResourceAttr(resourceName, "storage_capacity_quota_gib", "40"),
					resource.TestCheckResourceAttr(resourceName, "storage_capacity_reservation_gib", "30"),
				),
			},
		},
	})
}

func TestAccFSxOpenZFSVolume_nfsExports(t *testing.T) {
	ctx := acctest.Context(t)
	var volume1, volume2 awstypes.Volume
	resourceName := "aws_fsx_openzfs_volume.test"
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckPartitionHasService(t, names.FSxEndpointID) },
		ErrorCheck:               acctest.ErrorCheck(t, names.FSxServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckOpenZFSVolumeDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccOpenZFSVolumeConfig_nfsExports1(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckOpenZFSVolumeExists(ctx, t, resourceName, &volume1),
					resource.TestCheckResourceAttr(resourceName, "nfs_exports.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "nfs_exports.0.client_configurations.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "nfs_exports.0.client_configurations.0.clients", "10.0.1.0/24"),
					resource.TestCheckResourceAttr(resourceName, "nfs_exports.0.client_configurations.0.options.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "nfs_exports.0.client_configurations.0.options.0", "async"),
					resource.TestCheckResourceAttr(resourceName, "nfs_exports.0.client_configurations.0.options.1", "rw"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccOpenZFSVolumeConfig_nfsExports2(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckOpenZFSVolumeExists(ctx, t, resourceName, &volume2),
					testAccCheckOpenZFSVolumeNotRecreated(&volume1, &volume2),
					resource.TestCheckResourceAttr(resourceName, "nfs_exports.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "nfs_exports.0.client_configurations.#", "2"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "nfs_exports.0.client_configurations.*", map[string]string{
						"clients":   "10.0.1.0/24",
						"options.0": "async",
						"options.1": "rw",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "nfs_exports.0.client_configurations.*", map[string]string{
						"clients":   "*",
						"options.0": "sync",
						"options.1": "rw",
					}),
				),
			},
		},
	})
}

func TestAccFSxOpenZFSVolume_userAndGroupQuotas(t *testing.T) {
	ctx := acctest.Context(t)
	var volume1, volume2 awstypes.Volume
	resourceName := "aws_fsx_openzfs_volume.test"
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckPartitionHasService(t, names.FSxEndpointID) },
		ErrorCheck:               acctest.ErrorCheck(t, names.FSxServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckOpenZFSVolumeDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccOpenZFSVolumeConfig_userAndGroupQuotas1(rName, 256),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckOpenZFSVolumeExists(ctx, t, resourceName, &volume1),
					resource.TestCheckResourceAttr(resourceName, "user_and_group_quotas.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "user_and_group_quotas.*", map[string]string{
						names.AttrID:                 "10",
						"storage_capacity_quota_gib": "256",
						names.AttrType:               "USER",
					}),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccOpenZFSVolumeConfig_userAndGroupQuotas2(rName, 128, 1024),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckOpenZFSVolumeExists(ctx, t, resourceName, &volume2),
					testAccCheckOpenZFSVolumeNotRecreated(&volume1, &volume2),
					resource.TestCheckResourceAttr(resourceName, "user_and_group_quotas.#", "4"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "user_and_group_quotas.*", map[string]string{
						names.AttrID:                 "10",
						"storage_capacity_quota_gib": "128",
						names.AttrType:               "USER",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "user_and_group_quotas.*", map[string]string{
						names.AttrID:                 "20",
						"storage_capacity_quota_gib": "1024",
						names.AttrType:               "GROUP",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "user_and_group_quotas.*", map[string]string{
						names.AttrID:                 "5",
						"storage_capacity_quota_gib": "1024",
						names.AttrType:               "GROUP",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "user_and_group_quotas.*", map[string]string{
						names.AttrID:                 "100",
						"storage_capacity_quota_gib": "128",
						names.AttrType:               "USER",
					}),
				),
			},
		},
	})
}

func testAccCheckOpenZFSVolumeExists(ctx context.Context, t *testing.T, n string, v *awstypes.Volume) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.ProviderMeta(ctx, t).FSxClient(ctx)

		output, err := tffsx.FindOpenZFSVolumeByID(ctx, conn, rs.Primary.ID)

		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

func testAccCheckOpenZFSVolumeDestroy(ctx context.Context, t *testing.T) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.ProviderMeta(ctx, t).FSxClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_fsx_openzfs_volume" {
				continue
			}

			_, err := tffsx.FindOpenZFSVolumeByID(ctx, conn, rs.Primary.ID)

			if retry.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("FSx for OpenZFS Volume %s still exists", rs.Primary.ID)
		}
		return nil
	}
}

func testAccCheckOpenZFSVolumeNotRecreated(i, j *awstypes.Volume) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if aws.ToString(i.VolumeId) != aws.ToString(j.VolumeId) {
			return fmt.Errorf("FSx for OpenZFS Volume (%s) recreated", aws.ToString(i.VolumeId))
		}

		return nil
	}
}

func testAccCheckOpenZFSVolumeRecreated(i, j *awstypes.Volume) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if aws.ToString(i.VolumeId) == aws.ToString(j.VolumeId) {
			return fmt.Errorf("FSx for OpenZFS Volume (%s) not recreated", aws.ToString(i.VolumeId))
		}

		return nil
	}
}

func testAccOpenZFSVolumeConfig_base(rName string) string {
	return acctest.ConfigCompose(acctest.ConfigVPCWithSubnets(rName, 1), fmt.Sprintf(`
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

func testAccOpenZFSVolumeConfig_basic(rName string) string {
	return acctest.ConfigCompose(testAccOpenZFSVolumeConfig_base(rName), fmt.Sprintf(`
resource "aws_fsx_openzfs_volume" "test" {
  name             = %[1]q
  parent_volume_id = aws_fsx_openzfs_file_system.test.root_volume_id
}
`, rName))
}

func testAccOpenZFSVolumeConfig_parent(rName, rName2 string) string {
	return acctest.ConfigCompose(testAccOpenZFSVolumeConfig_base(rName), fmt.Sprintf(`
resource "aws_fsx_openzfs_volume" "test" {
  name             = %[1]q
  parent_volume_id = aws_fsx_openzfs_file_system.test.root_volume_id
}

resource "aws_fsx_openzfs_volume" "test2" {
  name             = %[2]q
  parent_volume_id = aws_fsx_openzfs_volume.test.id
}
`, rName, rName2))
}

func testAccOpenZFSVolumeConfig_tags1(rName, tagKey1, tagValue1 string) string {
	return acctest.ConfigCompose(testAccOpenZFSVolumeConfig_base(rName), fmt.Sprintf(`
resource "aws_fsx_openzfs_volume" "test" {
  name             = %[1]q
  parent_volume_id = aws_fsx_openzfs_file_system.test.root_volume_id

  tags = {
    %[2]q = %[3]q
  }
}
`, rName, tagKey1, tagValue1))
}

func testAccOpenZFSVolumeConfig_tags2(rName, tagKey1, tagValue1, tagKey2, tagValue2 string) string {
	return acctest.ConfigCompose(testAccOpenZFSVolumeConfig_base(rName), fmt.Sprintf(`
resource "aws_fsx_openzfs_volume" "test" {
  name             = %[1]q
  parent_volume_id = aws_fsx_openzfs_file_system.test.root_volume_id


  tags = {
    %[2]q = %[3]q
    %[4]q = %[5]q
  }
}
`, rName, tagKey1, tagValue1, tagKey2, tagValue2))
}

func testAccOpenZFSVolumeConfig_copyTags(rName, tagKey1, tagValue1, copyTags string) string {
	return acctest.ConfigCompose(testAccOpenZFSVolumeConfig_base(rName), fmt.Sprintf(`
resource "aws_fsx_openzfs_volume" "test" {
  name                   = %[1]q
  parent_volume_id       = aws_fsx_openzfs_file_system.test.root_volume_id
  copy_tags_to_snapshots = %[4]s

  tags = {
    %[2]q = %[3]q
  }

  delete_volume_options = ["DELETE_CHILD_VOLUMES_AND_SNAPSHOTS"]
}
`, rName, tagKey1, tagValue1, copyTags))
}

func testAccOpenZFSVolumeConfig_dataCompression(rName, dType string) string {
	return acctest.ConfigCompose(testAccOpenZFSVolumeConfig_base(rName), fmt.Sprintf(`
resource "aws_fsx_openzfs_volume" "test" {
  name                  = %[1]q
  parent_volume_id      = aws_fsx_openzfs_file_system.test.root_volume_id
  data_compression_type = %[2]q
}
`, rName, dType))
}

func testAccOpenZFSVolumeConfig_readOnly(rName, readOnly string) string {
	return acctest.ConfigCompose(testAccOpenZFSVolumeConfig_base(rName), fmt.Sprintf(`
resource "aws_fsx_openzfs_volume" "test" {
  name             = %[1]q
  parent_volume_id = aws_fsx_openzfs_file_system.test.root_volume_id
  read_only        = %[2]s
}
`, rName, readOnly))
}

func testAccOpenZFSVolumeConfig_recordSizeKib(rName string, recordSizeKib int) string {
	return acctest.ConfigCompose(testAccOpenZFSVolumeConfig_base(rName), fmt.Sprintf(`
resource "aws_fsx_openzfs_volume" "test" {
  name             = %[1]q
  parent_volume_id = aws_fsx_openzfs_file_system.test.root_volume_id
  record_size_kib  = %[2]d
}
`, rName, recordSizeKib))
}

func testAccOpenZFSVolumeConfig_storageCapacity(rName string, storageQuota, storageReservation int) string {
	return acctest.ConfigCompose(testAccOpenZFSVolumeConfig_base(rName), fmt.Sprintf(`
resource "aws_fsx_openzfs_volume" "test" {
  name                             = %[1]q
  parent_volume_id                 = aws_fsx_openzfs_file_system.test.root_volume_id
  storage_capacity_quota_gib       = %[2]d
  storage_capacity_reservation_gib = %[3]d
}
`, rName, storageQuota, storageReservation))
}

func testAccOpenZFSVolumeConfig_nfsExports1(rName string) string {
	return acctest.ConfigCompose(testAccOpenZFSVolumeConfig_base(rName), fmt.Sprintf(`
resource "aws_fsx_openzfs_volume" "test" {
  name             = %[1]q
  parent_volume_id = aws_fsx_openzfs_file_system.test.root_volume_id
  nfs_exports {
    client_configurations {
      clients = "10.0.1.0/24"
      options = ["async", "rw"]
    }
  }

}
`, rName))
}

func testAccOpenZFSVolumeConfig_nfsExports2(rName string) string {
	return acctest.ConfigCompose(testAccOpenZFSVolumeConfig_base(rName), fmt.Sprintf(`
resource "aws_fsx_openzfs_volume" "test" {
  name             = %[1]q
  parent_volume_id = aws_fsx_openzfs_file_system.test.root_volume_id
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
}
`, rName))
}

func testAccOpenZFSVolumeConfig_userAndGroupQuotas1(rName string, quotaSize int) string {
	return acctest.ConfigCompose(testAccOpenZFSVolumeConfig_base(rName), fmt.Sprintf(`
resource "aws_fsx_openzfs_volume" "test" {
  name             = %[1]q
  parent_volume_id = aws_fsx_openzfs_file_system.test.root_volume_id
  user_and_group_quotas {
    id                         = 10
    storage_capacity_quota_gib = %[2]d
    type                       = "USER"
  }
}
`, rName, quotaSize))
}

func testAccOpenZFSVolumeConfig_userAndGroupQuotas2(rName string, userQuota, groupQuota int) string {
	return acctest.ConfigCompose(testAccOpenZFSVolumeConfig_base(rName), fmt.Sprintf(`
resource "aws_fsx_openzfs_volume" "test" {
  name             = %[1]q
  parent_volume_id = aws_fsx_openzfs_file_system.test.root_volume_id
  user_and_group_quotas {
    id                         = 10
    storage_capacity_quota_gib = %[2]d
    type                       = "USER"
  }
  user_and_group_quotas {
    id                         = 20
    storage_capacity_quota_gib = %[3]d
    type                       = "GROUP"
  }
  user_and_group_quotas {
    id                         = 5
    storage_capacity_quota_gib = %[3]d
    type                       = "GROUP"
  }
  user_and_group_quotas {
    id                         = 100
    storage_capacity_quota_gib = %[2]d
    type                       = "USER"
  }
}
`, rName, userQuota, groupQuota))
}
