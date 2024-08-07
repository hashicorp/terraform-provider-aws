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

func TestAccFSxONTAPVolume_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var volume awstypes.Volume
	resourceName := "aws_fsx_ontap_volume.test"
	rName := fmt.Sprintf("tf_acc_test_%d", sdkacctest.RandInt())

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckPartitionHasService(t, names.FSxEndpointID) },
		ErrorCheck:               acctest.ErrorCheck(t, names.FSxServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckONTAPVolumeDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccONTAPVolumeConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckONTAPVolumeExists(ctx, resourceName, &volume),
					acctest.MatchResourceAttrRegionalARN(resourceName, names.AttrARN, "fsx", regexache.MustCompile(`volume/fs-.+/fsvol-.+`)),
					resource.TestCheckResourceAttr(resourceName, "bypass_snaplock_enterprise_retention", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "copy_tags_to_backups", acctest.CtFalse),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrFileSystemID),
					resource.TestCheckResourceAttr(resourceName, "junction_path", fmt.Sprintf("/%[1]s", rName)),
					resource.TestCheckResourceAttr(resourceName, "ontap_volume_type", "RW"),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, "security_style", ""),
					resource.TestCheckResourceAttr(resourceName, "size_in_megabytes", "1024"),
					resource.TestCheckResourceAttr(resourceName, "skip_final_backup", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "snaplock_configuration.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "snapshot_policy", "default"),
					resource.TestCheckResourceAttr(resourceName, "storage_efficiency_enabled", acctest.CtTrue),
					resource.TestCheckResourceAttrSet(resourceName, "storage_virtual_machine_id"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "tiering_policy.#", acctest.Ct1),
					resource.TestCheckResourceAttrSet(resourceName, "uuid"),
					resource.TestCheckResourceAttr(resourceName, names.AttrVolumeType, "ONTAP"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"final_backup_tags",
					"skip_final_backup",
				},
			},
		},
	})
}

func TestAccFSxONTAPVolume_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	var volume awstypes.Volume
	resourceName := "aws_fsx_ontap_volume.test"
	rName := fmt.Sprintf("tf_acc_test_%d", sdkacctest.RandInt())

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckPartitionHasService(t, names.FSxEndpointID) },
		ErrorCheck:               acctest.ErrorCheck(t, names.FSxServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckONTAPVolumeDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccONTAPVolumeConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckONTAPVolumeExists(ctx, resourceName, &volume),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tffsx.ResourceONTAPVolume(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccFSxONTAPVolume_aggregateConfiguration(t *testing.T) {
	ctx := acctest.Context(t)
	var volume1, volume2 awstypes.Volume
	ConstituentsPerAggregate := 10
	resourceName := "aws_fsx_ontap_volume.test"
	rName := fmt.Sprintf("tf_acc_test_%d", sdkacctest.RandInt())

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckPartitionHasService(t, names.FSxEndpointID) },
		ErrorCheck:               acctest.ErrorCheck(t, names.FSxServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckONTAPVolumeDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccONTAPVolumeConfig_aggregate(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckONTAPVolumeExists(ctx, resourceName, &volume1),
					resource.TestCheckResourceAttr(resourceName, "aggregate_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "aggregate_configuration.0.aggregates.#", acctest.Ct2),
					resource.TestCheckResourceAttr(resourceName, "aggregate_configuration.0.aggregates.0", "aggr1"),
					resource.TestCheckResourceAttr(resourceName, "aggregate_configuration.0.aggregates.1", "aggr2"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"bypass_snaplock_enterprise_retention",
					"final_backup_tags",
					"skip_final_backup",
				},
			},
			{
				Config: testAccONTAPVolumeConfig_aggregateConstituents(rName, ConstituentsPerAggregate, ConstituentsPerAggregate*204800),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckONTAPVolumeExists(ctx, resourceName, &volume2),
					testAccCheckONTAPVolumeRecreated(&volume1, &volume2),
					resource.TestCheckResourceAttr(resourceName, "aggregate_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "aggregate_configuration.0.aggregates.#", acctest.Ct2),
					resource.TestCheckResourceAttr(resourceName, "aggregate_configuration.0.aggregates.0", "aggr1"),
					resource.TestCheckResourceAttr(resourceName, "aggregate_configuration.0.aggregates.1", "aggr2"),
					resource.TestCheckResourceAttr(resourceName, "aggregate_configuration.0.total_constituents", fmt.Sprint(ConstituentsPerAggregate*2)),
					resource.TestCheckResourceAttr(resourceName, "size_in_megabytes", fmt.Sprint(ConstituentsPerAggregate*204800)),
				),
			},
		},
	})
}

func TestAccFSxONTAPVolume_copyTagsToBackups(t *testing.T) {
	ctx := acctest.Context(t)
	var volume1, volume2 awstypes.Volume
	resourceName := "aws_fsx_ontap_volume.test"
	rName := fmt.Sprintf("tf_acc_test_%d", sdkacctest.RandInt())

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckPartitionHasService(t, names.FSxEndpointID) },
		ErrorCheck:               acctest.ErrorCheck(t, names.FSxServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckONTAPVolumeDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccONTAPVolumeConfig_copyTagsToBackups(rName, true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckONTAPVolumeExists(ctx, resourceName, &volume1),
					resource.TestCheckResourceAttr(resourceName, "copy_tags_to_backups", acctest.CtTrue),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"bypass_snaplock_enterprise_retention",
					"final_backup_tags",
					"skip_final_backup",
				},
			},
			{
				Config: testAccONTAPVolumeConfig_copyTagsToBackups(rName, false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckONTAPVolumeExists(ctx, resourceName, &volume2),
					testAccCheckONTAPVolumeNotRecreated(&volume1, &volume2),
					resource.TestCheckResourceAttr(resourceName, "copy_tags_to_backups", acctest.CtFalse),
				),
			},
		},
	})
}

func TestAccFSxONTAPVolume_junctionPath(t *testing.T) {
	ctx := acctest.Context(t)
	var volume1, volume2 awstypes.Volume
	resourceName := "aws_fsx_ontap_volume.test"
	rName := fmt.Sprintf("tf_acc_test_%d", sdkacctest.RandInt())
	jPath1 := "/path1"
	jPath2 := "/path2"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckPartitionHasService(t, names.FSxEndpointID) },
		ErrorCheck:               acctest.ErrorCheck(t, names.FSxServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckONTAPVolumeDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccONTAPVolumeConfig_junctionPath(rName, jPath1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckONTAPVolumeExists(ctx, resourceName, &volume1),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, "junction_path", jPath1),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"bypass_snaplock_enterprise_retention",
					"final_backup_tags",
					"skip_final_backup",
				},
			},
			{
				Config: testAccONTAPVolumeConfig_junctionPath(rName, jPath2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckONTAPVolumeExists(ctx, resourceName, &volume2),
					testAccCheckONTAPVolumeNotRecreated(&volume1, &volume2),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, "junction_path", jPath2),
				),
			},
		},
	})
}

func TestAccFSxONTAPVolume_name(t *testing.T) {
	ctx := acctest.Context(t)
	var volume1, volume2 awstypes.Volume
	resourceName := "aws_fsx_ontap_volume.test"
	rName := fmt.Sprintf("tf_acc_test_%d", sdkacctest.RandInt())
	rName2 := fmt.Sprintf("tf_acc_test_%d", sdkacctest.RandInt())

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckPartitionHasService(t, names.FSxEndpointID) },
		ErrorCheck:               acctest.ErrorCheck(t, names.FSxServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckONTAPVolumeDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccONTAPVolumeConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckONTAPVolumeExists(ctx, resourceName, &volume1),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"bypass_snaplock_enterprise_retention",
					"final_backup_tags",
					"skip_final_backup",
				},
			},
			{
				Config: testAccONTAPVolumeConfig_basic(rName2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckONTAPVolumeExists(ctx, resourceName, &volume2),
					testAccCheckONTAPVolumeRecreated(&volume1, &volume2),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName2),
				),
			},
		},
	})
}

func TestAccFSxONTAPVolume_ontapVolumeType(t *testing.T) {
	ctx := acctest.Context(t)
	var volume awstypes.Volume
	resourceName := "aws_fsx_ontap_volume.test"
	rName := fmt.Sprintf("tf_acc_test_%d", sdkacctest.RandInt())

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckPartitionHasService(t, names.FSxEndpointID) },
		ErrorCheck:               acctest.ErrorCheck(t, names.FSxServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckONTAPVolumeDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccONTAPVolumeConfig_ontapVolumeTypeDP(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckONTAPVolumeExists(ctx, resourceName, &volume),
					resource.TestCheckResourceAttr(resourceName, "ontap_volume_type", "DP"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"bypass_snaplock_enterprise_retention",
					"final_backup_tags",
					"skip_final_backup",
				},
			},
		},
	})
}

func TestAccFSxONTAPVolume_securityStyle(t *testing.T) {
	ctx := acctest.Context(t)
	var volume1, volume2, volume3 awstypes.Volume
	resourceName := "aws_fsx_ontap_volume.test"
	rName := fmt.Sprintf("tf_acc_test_%d", sdkacctest.RandInt())

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckPartitionHasService(t, names.FSxEndpointID) },
		ErrorCheck:               acctest.ErrorCheck(t, names.FSxServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckONTAPVolumeDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccONTAPVolumeConfig_securityStyle(rName, "UNIX"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckONTAPVolumeExists(ctx, resourceName, &volume1),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, "security_style", "UNIX"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"bypass_snaplock_enterprise_retention",
					"final_backup_tags",
					"skip_final_backup",
				},
			},
			{
				Config: testAccONTAPVolumeConfig_securityStyle(rName, "NTFS"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckONTAPVolumeExists(ctx, resourceName, &volume2),
					testAccCheckONTAPVolumeNotRecreated(&volume1, &volume2),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, "security_style", "NTFS"),
				),
			},
			{
				Config: testAccONTAPVolumeConfig_securityStyle(rName, "MIXED"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckONTAPVolumeExists(ctx, resourceName, &volume3),
					testAccCheckONTAPVolumeNotRecreated(&volume1, &volume3),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, "security_style", "MIXED"),
				),
			},
		},
	})
}

func TestAccFSxONTAPVolume_size(t *testing.T) {
	ctx := acctest.Context(t)
	var volume1, volume2, volume3, volume4 awstypes.Volume
	resourceName := "aws_fsx_ontap_volume.test"
	rName := fmt.Sprintf("tf_acc_test_%d", sdkacctest.RandInt())
	size1 := 1024
	size2 := 2048
	size3 := int64(549755813888000)
	size4 := int64(1125899906842623)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckPartitionHasService(t, names.FSxEndpointID) },
		ErrorCheck:               acctest.ErrorCheck(t, names.FSxServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckONTAPVolumeDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccONTAPVolumeConfig_size(rName, size1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckONTAPVolumeExists(ctx, resourceName, &volume1),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, "size_in_megabytes", fmt.Sprint(size1)),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"bypass_snaplock_enterprise_retention",
					"final_backup_tags",
					"skip_final_backup",
				},
			},
			{
				Config: testAccONTAPVolumeConfig_size(rName, size2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckONTAPVolumeExists(ctx, resourceName, &volume2),
					testAccCheckONTAPVolumeNotRecreated(&volume1, &volume2),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, "size_in_megabytes", fmt.Sprint(size2)),
				),
			},
			{
				Config: testAccONTAPVolumeConfig_sizeBytes(rName, size3),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckONTAPVolumeExists(ctx, resourceName, &volume3),
					testAccCheckONTAPVolumeRecreated(&volume2, &volume3),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, "size_in_bytes", fmt.Sprint(size3)),
				),
			},
			{
				Config: testAccONTAPVolumeConfig_sizeBytes(rName, size4),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckONTAPVolumeExists(ctx, resourceName, &volume4),
					testAccCheckONTAPVolumeNotRecreated(&volume3, &volume4),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, "size_in_bytes", fmt.Sprint(size4)),
				),
			},
		},
	})
}

func TestAccFSxONTAPVolume_snaplock(t *testing.T) {
	ctx := acctest.Context(t)
	var volume1 /*, volume2*/ awstypes.Volume
	resourceName := "aws_fsx_ontap_volume.test"
	rName := fmt.Sprintf("tf_acc_test_%d", sdkacctest.RandInt())

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckPartitionHasService(t, names.FSxEndpointID) },
		ErrorCheck:               acctest.ErrorCheck(t, names.FSxServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckONTAPVolumeDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccONTAPVolumeConfig_snaplockCreate(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckONTAPVolumeExists(ctx, resourceName, &volume1),
					resource.TestCheckResourceAttr(resourceName, "bypass_snaplock_enterprise_retention", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "snaplock_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "snaplock_configuration.0.audit_log_volume", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "snaplock_configuration.0.autocommit_period.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "snaplock_configuration.0.autocommit_period.0.type", "NONE"),
					resource.TestCheckResourceAttr(resourceName, "snaplock_configuration.0.autocommit_period.0.value", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "snaplock_configuration.0.privileged_delete", "DISABLED"),
					resource.TestCheckResourceAttr(resourceName, "snaplock_configuration.0.retention_period.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "snaplock_configuration.0.retention_period.0.default_retention.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "snaplock_configuration.0.retention_period.0.default_retention.0.type", "YEARS"),
					resource.TestCheckResourceAttr(resourceName, "snaplock_configuration.0.retention_period.0.default_retention.0.value", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "snaplock_configuration.0.retention_period.0.maximum_retention.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "snaplock_configuration.0.retention_period.0.maximum_retention.0.type", "YEARS"),
					resource.TestCheckResourceAttr(resourceName, "snaplock_configuration.0.retention_period.0.maximum_retention.0.value", "30"),
					resource.TestCheckResourceAttr(resourceName, "snaplock_configuration.0.retention_period.0.minimum_retention.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "snaplock_configuration.0.retention_period.0.minimum_retention.0.type", "YEARS"),
					resource.TestCheckResourceAttr(resourceName, "snaplock_configuration.0.retention_period.0.minimum_retention.0.value", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "snaplock_configuration.0.snaplock_type", "ENTERPRISE"),
					resource.TestCheckResourceAttr(resourceName, "snaplock_configuration.0.volume_append_mode_enabled", acctest.CtFalse),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"bypass_snaplock_enterprise_retention",
					"final_backup_tags",
					"skip_final_backup",
				},
			},
			/*
				See https://docs.aws.amazon.com/fsx/latest/ONTAPGuide/how-snaplock-works.html#snaplock-audit-log-volume.
				> The minimum retention period for a SnapLock audit log volume is six months. Until this retention period expires, the SnapLock audit log volume and the SVM and file system that are associated with it can't be deleted even if the volume was created in SnapLock Enterprise mode.

				{
					Config: testAccONTAPVolumeConfig_snaplockUpdate(rName),
					Check: resource.ComposeAggregateTestCheckFunc(
						testAccCheckONTAPVolumeExists(ctx, resourceName, &volume2),
						testAccCheckONTAPVolumeNotRecreated(&volume1, &volume2),
						resource.TestCheckResourceAttr(resourceName, "bypass_snaplock_enterprise_retention", "true"),
						resource.TestCheckResourceAttr(resourceName, "snaplock_configuration.#", "1"),
						resource.TestCheckResourceAttr(resourceName, "snaplock_configuration.0.audit_log_volume", "true"),
						resource.TestCheckResourceAttr(resourceName, "snaplock_configuration.0.autocommit_period.#", "1"),
						resource.TestCheckResourceAttr(resourceName, "snaplock_configuration.0.autocommit_period.0.type", "DAYS"),
						resource.TestCheckResourceAttr(resourceName, "snaplock_configuration.0.autocommit_period.0.value", "14"),
						resource.TestCheckResourceAttr(resourceName, "snaplock_configuration.0.privileged_delete", "PERMANENTLY_DISABLED"),
						resource.TestCheckResourceAttr(resourceName, "snaplock_configuration.0.retention_period.#", "1"),
						resource.TestCheckResourceAttr(resourceName, "snaplock_configuration.0.retention_period.0.default_retention.#", "1"),
						resource.TestCheckResourceAttr(resourceName, "snaplock_configuration.0.retention_period.0.default_retention.0.type", "DAYS"),
						resource.TestCheckResourceAttr(resourceName, "snaplock_configuration.0.retention_period.0.default_retention.0.value", "30"),
						resource.TestCheckResourceAttr(resourceName, "snaplock_configuration.0.retention_period.0.maximum_retention.#", "1"),
						resource.TestCheckResourceAttr(resourceName, "snaplock_configuration.0.retention_period.0.maximum_retention.0.type", "MONTHS"),
						resource.TestCheckResourceAttr(resourceName, "snaplock_configuration.0.retention_period.0.maximum_retention.0.value", "9"),
						resource.TestCheckResourceAttr(resourceName, "snaplock_configuration.0.retention_period.0.minimum_retention.#", "1"),
						resource.TestCheckResourceAttr(resourceName, "snaplock_configuration.0.retention_period.0.minimum_retention.0.type", "HOURS"),
						resource.TestCheckResourceAttr(resourceName, "snaplock_configuration.0.retention_period.0.minimum_retention.0.value", "24"),
						resource.TestCheckResourceAttr(resourceName, "snaplock_configuration.0.snaplock_type", "ENTERPRISE"),
						resource.TestCheckResourceAttr(resourceName, "snaplock_configuration.0.volume_append_mode_enabled", "true"),
					),
				},
			*/
		},
	})
}

func TestAccFSxONTAPVolume_snapshotPolicy(t *testing.T) {
	ctx := acctest.Context(t)
	var volume1, volume2 awstypes.Volume
	resourceName := "aws_fsx_ontap_volume.test"
	rName := fmt.Sprintf("tf_acc_test_%d", sdkacctest.RandInt())
	policy1 := "default"
	policy2 := "none"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckPartitionHasService(t, names.FSxEndpointID) },
		ErrorCheck:               acctest.ErrorCheck(t, names.FSxServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckONTAPVolumeDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccONTAPVolumeConfig_snapshotPolicy(rName, policy1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckONTAPVolumeExists(ctx, resourceName, &volume1),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, "snapshot_policy", fmt.Sprint(policy1)),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"bypass_snaplock_enterprise_retention",
					"final_backup_tags",
					"skip_final_backup",
				},
			},
			{
				Config: testAccONTAPVolumeConfig_snapshotPolicy(rName, policy2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckONTAPVolumeExists(ctx, resourceName, &volume2),
					testAccCheckONTAPVolumeNotRecreated(&volume1, &volume2),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, "snapshot_policy", fmt.Sprint(policy2)),
				),
			},
		},
	})
}

func TestAccFSxONTAPVolume_storageEfficiency(t *testing.T) {
	ctx := acctest.Context(t)
	var volume1, volume2 awstypes.Volume
	resourceName := "aws_fsx_ontap_volume.test"
	rName := fmt.Sprintf("tf_acc_test_%d", sdkacctest.RandInt())

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckPartitionHasService(t, names.FSxEndpointID) },
		ErrorCheck:               acctest.ErrorCheck(t, names.FSxServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckONTAPVolumeDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccONTAPVolumeConfig_storageEfficiency(rName, true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckONTAPVolumeExists(ctx, resourceName, &volume1),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, "storage_efficiency_enabled", acctest.CtTrue),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"bypass_snaplock_enterprise_retention",
					"final_backup_tags",
					"skip_final_backup",
				},
			},
			{
				Config: testAccONTAPVolumeConfig_storageEfficiency(rName, false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckONTAPVolumeExists(ctx, resourceName, &volume2),
					testAccCheckONTAPVolumeNotRecreated(&volume1, &volume2),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, "storage_efficiency_enabled", acctest.CtFalse),
				),
			},
		},
	})
}

func TestAccFSxONTAPVolume_tags(t *testing.T) {
	ctx := acctest.Context(t)
	var volume1, volume2, volume3 awstypes.Volume
	resourceName := "aws_fsx_ontap_volume.test"
	rName := fmt.Sprintf("tf_acc_test_%d", sdkacctest.RandInt())

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckPartitionHasService(t, names.FSxEndpointID) },
		ErrorCheck:               acctest.ErrorCheck(t, names.FSxServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckONTAPVolumeDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccONTAPVolumeConfig_tags1(rName, acctest.CtKey1, acctest.CtValue1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckONTAPVolumeExists(ctx, resourceName, &volume1),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey1, acctest.CtValue1),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"bypass_snaplock_enterprise_retention",
					"final_backup_tags",
					"skip_final_backup",
				},
			},
			{
				Config: testAccONTAPVolumeConfig_tags2(rName, acctest.CtKey1, acctest.CtValue1Updated, acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckONTAPVolumeExists(ctx, resourceName, &volume2),
					testAccCheckONTAPVolumeNotRecreated(&volume1, &volume2),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct2),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey1, acctest.CtValue1Updated),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey2, acctest.CtValue2),
				),
			},
			{
				Config: testAccONTAPVolumeConfig_tags1(rName, acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckONTAPVolumeExists(ctx, resourceName, &volume3),
					testAccCheckONTAPVolumeNotRecreated(&volume2, &volume3),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey2, acctest.CtValue2),
				),
			},
		},
	})
}

func TestAccFSxONTAPVolume_tieringPolicy(t *testing.T) {
	ctx := acctest.Context(t)
	var volume1, volume2, volume3, volume4 awstypes.Volume
	resourceName := "aws_fsx_ontap_volume.test"
	rName := fmt.Sprintf("tf_acc_test_%d", sdkacctest.RandInt())

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckPartitionHasService(t, names.FSxEndpointID) },
		ErrorCheck:               acctest.ErrorCheck(t, names.FSxServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckONTAPVolumeDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccONTAPVolumeConfig_tieringPolicyNoCooling(rName, "NONE"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckONTAPVolumeExists(ctx, resourceName, &volume1),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, "tiering_policy.0.name", "NONE"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"bypass_snaplock_enterprise_retention",
					"final_backup_tags",
					"skip_final_backup",
				},
			},
			{
				Config: testAccONTAPVolumeConfig_tieringPolicy(rName, "SNAPSHOT_ONLY", 10),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckONTAPVolumeExists(ctx, resourceName, &volume2),
					testAccCheckONTAPVolumeNotRecreated(&volume1, &volume2),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, "tiering_policy.0.name", "SNAPSHOT_ONLY"),
					resource.TestCheckResourceAttr(resourceName, "tiering_policy.0.cooling_period", acctest.Ct10),
				),
			},
			{
				Config: testAccONTAPVolumeConfig_tieringPolicy(rName, "AUTO", 60),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckONTAPVolumeExists(ctx, resourceName, &volume3),
					testAccCheckONTAPVolumeNotRecreated(&volume1, &volume3),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, "tiering_policy.0.name", "AUTO"),
					resource.TestCheckResourceAttr(resourceName, "tiering_policy.0.cooling_period", "60"),
				),
			},
			{
				Config: testAccONTAPVolumeConfig_tieringPolicyNoCooling(rName, "ALL"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckONTAPVolumeExists(ctx, resourceName, &volume4),
					testAccCheckONTAPVolumeNotRecreated(&volume1, &volume4),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, "tiering_policy.0.name", "ALL"),
				),
			},
		},
	})
}

func TestAccFSxONTAPVolume_volumeStyle(t *testing.T) {
	ctx := acctest.Context(t)
	var volume awstypes.Volume
	resourceName := "aws_fsx_ontap_volume.test"
	style1 := "FLEXVOL"
	style2 := "FLEXGROUP"
	rName := fmt.Sprintf("tf_acc_test_%d", sdkacctest.RandInt())

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckPartitionHasService(t, names.FSxEndpointID) },
		ErrorCheck:               acctest.ErrorCheck(t, names.FSxServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckONTAPVolumeDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccONTAPVolumeConfig_ontapStyle(rName, style1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckONTAPVolumeExists(ctx, resourceName, &volume),
					resource.TestCheckResourceAttr(resourceName, "volume_style", style1),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"bypass_snaplock_enterprise_retention",
					"final_backup_tags",
					"skip_final_backup",
				},
			},
			{
				Config: testAccONTAPVolumeConfig_ontapStyle(rName, style2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckONTAPVolumeExists(ctx, resourceName, &volume),
					resource.TestCheckResourceAttr(resourceName, "volume_style", style2),
				),
			},
		},
	})
}

func TestAccFSxONTAPVolume_deleteConfig(t *testing.T) {
	ctx := acctest.Context(t)
	var volume awstypes.Volume
	resourceName := "aws_fsx_ontap_volume.test"
	rName := fmt.Sprintf("tf_acc_test_%d", sdkacctest.RandInt())

	acctest.SkipIfEnvVarNotSet(t, "AWS_FSX_CREATE_FINAL_BACKUP")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckPartitionHasService(t, names.FSxEndpointID) },
		ErrorCheck:               acctest.ErrorCheck(t, names.FSxServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckONTAPVolumeDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccONTAPVolumeConfig_deleteConfig(rName, acctest.CtKey1, acctest.CtValue1, acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckONTAPVolumeExists(ctx, resourceName, &volume),
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
					"bypass_snaplock_enterprise_retention",
					"final_backup_tags",
					"skip_final_backup",
				},
			},
		},
	})
}

func testAccCheckONTAPVolumeExists(ctx context.Context, n string, v *awstypes.Volume) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).FSxClient(ctx)

		output, err := tffsx.FindONTAPVolumeByID(ctx, conn, rs.Primary.ID)

		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

func testAccCheckONTAPVolumeDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).FSxClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_fsx_ontap_volume" {
				continue
			}

			_, err := tffsx.FindONTAPVolumeByID(ctx, conn, rs.Primary.ID)

			if tfresource.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("FSx for NetApp ONTAP Volume %s still exists", rs.Primary.ID)
		}
		return nil
	}
}

func testAccCheckONTAPVolumeNotRecreated(i, j *awstypes.Volume) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if aws.ToString(i.VolumeId) != aws.ToString(j.VolumeId) {
			return fmt.Errorf("FSx for NetApp ONTAP Volume (%s) recreated", aws.ToString(i.VolumeId))
		}

		return nil
	}
}

func testAccCheckONTAPVolumeRecreated(i, j *awstypes.Volume) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if aws.ToString(i.VolumeId) == aws.ToString(j.VolumeId) {
			return fmt.Errorf("FSx for NetApp ONTAP Volume (%s) not recreated", aws.ToString(i.VolumeId))
		}

		return nil
	}
}

func testAccONTAPVolumeConfig_base(rName string) string {
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

func testAccONTAPVolumeConfig_baseScaleOut(rName string) string {
	return acctest.ConfigCompose(acctest.ConfigVPCWithSubnets(rName, 2), fmt.Sprintf(`
resource "aws_fsx_ontap_file_system" "test" {
  storage_capacity                = 2048
  subnet_ids                      = [aws_subnet.test[0].id]
  deployment_type                 = "SINGLE_AZ_2"
  ha_pairs                        = 2
  throughput_capacity_per_ha_pair = 3072
  preferred_subnet_id             = aws_subnet.test[0].id

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
	return acctest.ConfigCompose(testAccONTAPVolumeConfig_base(rName), fmt.Sprintf(`
resource "aws_fsx_ontap_volume" "test" {
  name                       = %[1]q
  junction_path              = "/%[1]s"
  size_in_megabytes          = 1024
  skip_final_backup          = true
  storage_efficiency_enabled = true
  storage_virtual_machine_id = aws_fsx_ontap_storage_virtual_machine.test.id
}
`, rName))
}

func testAccONTAPVolumeConfig_aggregate(rName string) string {
	return acctest.ConfigCompose(testAccONTAPVolumeConfig_baseScaleOut(rName), fmt.Sprintf(`
resource "aws_fsx_ontap_volume" "test" {
  name                       = %[1]q
  junction_path              = "/%[1]s"
  size_in_megabytes          = 16 * 102400
  skip_final_backup          = true
  storage_efficiency_enabled = true
  storage_virtual_machine_id = aws_fsx_ontap_storage_virtual_machine.test.id
  volume_style               = "FLEXGROUP"

  aggregate_configuration {
    aggregates = ["aggr1", "aggr2"]
  }

}
`, rName))
}

func testAccONTAPVolumeConfig_aggregateConstituents(rName string, ConstituentsPerAggregate int, size int) string {
	return acctest.ConfigCompose(testAccONTAPVolumeConfig_baseScaleOut(rName), fmt.Sprintf(`
resource "aws_fsx_ontap_volume" "test" {
  name                       = %[1]q
  junction_path              = "/%[1]s"
  size_in_megabytes          = %[3]d
  skip_final_backup          = true
  storage_efficiency_enabled = true
  storage_virtual_machine_id = aws_fsx_ontap_storage_virtual_machine.test.id
  volume_style               = "FLEXGROUP"

  aggregate_configuration {
    aggregates                 = ["aggr1", "aggr2"]
    constituents_per_aggregate = %[2]d
  }

}
`, rName, ConstituentsPerAggregate, size))
}

func testAccONTAPVolumeConfig_copyTagsToBackups(rName string, copyTagsToBackups bool) string {
	return acctest.ConfigCompose(testAccONTAPVolumeConfig_base(rName), fmt.Sprintf(`
resource "aws_fsx_ontap_volume" "test" {
  name                       = %[1]q
  junction_path              = "/%[1]s"
  size_in_megabytes          = 1024
  skip_final_backup          = true
  storage_efficiency_enabled = true
  storage_virtual_machine_id = aws_fsx_ontap_storage_virtual_machine.test.id
  copy_tags_to_backups       = %[2]t
}
`, rName, copyTagsToBackups))
}

func testAccONTAPVolumeConfig_junctionPath(rName, junctionPath string) string {
	return acctest.ConfigCompose(testAccONTAPVolumeConfig_base(rName), fmt.Sprintf(`
resource "aws_fsx_ontap_volume" "test" {
  name                       = %[1]q
  junction_path              = %[2]q
  skip_final_backup          = true
  size_in_megabytes          = 1024
  storage_efficiency_enabled = true
  storage_virtual_machine_id = aws_fsx_ontap_storage_virtual_machine.test.id
}
`, rName, junctionPath))
}

func testAccONTAPVolumeConfig_ontapVolumeTypeDP(rName string) string {
	return acctest.ConfigCompose(testAccONTAPVolumeConfig_base(rName), fmt.Sprintf(`
resource "aws_fsx_ontap_volume" "test" {
  name                       = %[1]q
  ontap_volume_type          = "DP"
  size_in_megabytes          = 1024
  skip_final_backup          = true
  storage_virtual_machine_id = aws_fsx_ontap_storage_virtual_machine.test.id
}
`, rName))
}

func testAccONTAPVolumeConfig_securityStyle(rName, securityStyle string) string {
	return acctest.ConfigCompose(testAccONTAPVolumeConfig_base(rName), fmt.Sprintf(`
resource "aws_fsx_ontap_volume" "test" {
  name                       = %[1]q
  junction_path              = "/%[1]s"
  size_in_megabytes          = 1024
  security_style             = %[2]q
  skip_final_backup          = true
  storage_efficiency_enabled = true
  storage_virtual_machine_id = aws_fsx_ontap_storage_virtual_machine.test.id
}
`, rName, securityStyle))
}

func testAccONTAPVolumeConfig_size(rName string, size int) string {
	return acctest.ConfigCompose(testAccONTAPVolumeConfig_base(rName), fmt.Sprintf(`
resource "aws_fsx_ontap_volume" "test" {
  name                       = %[1]q
  junction_path              = "/%[1]s"
  size_in_megabytes          = %[2]d
  skip_final_backup          = true
  storage_efficiency_enabled = true
  storage_virtual_machine_id = aws_fsx_ontap_storage_virtual_machine.test.id
}
`, rName, size))
}

func testAccONTAPVolumeConfig_sizeBytes(rName string, size int64) string {
	return acctest.ConfigCompose(testAccONTAPVolumeConfig_base(rName), fmt.Sprintf(`
resource "aws_fsx_ontap_volume" "test" {
  name                       = %[1]q
  junction_path              = "/%[1]s"
  size_in_bytes              = %[2]d
  skip_final_backup          = true
  storage_efficiency_enabled = true
  storage_virtual_machine_id = aws_fsx_ontap_storage_virtual_machine.test.id
  volume_style               = "FLEXGROUP"
}
`, rName, size))
}

func testAccONTAPVolumeConfig_snaplockCreate(rName string) string {
	return acctest.ConfigCompose(testAccONTAPVolumeConfig_base(rName), fmt.Sprintf(`
resource "aws_fsx_ontap_volume" "test" {
  name                       = %[1]q
  junction_path              = "/%[1]s"
  size_in_megabytes          = 1024
  skip_final_backup          = true
  storage_efficiency_enabled = true
  storage_virtual_machine_id = aws_fsx_ontap_storage_virtual_machine.test.id

  snaplock_configuration {
    snaplock_type = "ENTERPRISE"
  }

  bypass_snaplock_enterprise_retention = true
}
`, rName))
}

/*
func testAccONTAPVolumeConfig_snaplockUpdate(rName string) string {
	return acctest.ConfigCompose(testAccONTAPVolumeConfig_base(rName), fmt.Sprintf(`
resource "aws_fsx_ontap_volume" "test" {
  name                       = %[1]q
  junction_path              = "/snaplock_audit_log"
  size_in_megabytes          = 1024
  skip_final_backup          = true
  storage_efficiency_enabled = true
  storage_virtual_machine_id = aws_fsx_ontap_storage_virtual_machine.test.id

  snaplock_configuration {
    audit_log_volume           = true
    privileged_delete          = "PERMANENTLY_DISABLED"
    snaplock_type              = "ENTERPRISE"
    volume_append_mode_enabled = true

    autocommit_period {
      type  = "DAYS"
      value = 14
    }

    retention_period {
      default_retention {
        type  = "DAYS"
        value = 30
      }

      maximum_retention {
        type  = "MONTHS"
        value = 9
      }

      minimum_retention {
        type  = "HOURS"
        value = 24
      }
    }
  }

  bypass_snaplock_enterprise_retention = true
}
`, rName))
}
*/

func testAccONTAPVolumeConfig_snapshotPolicy(rName, snapshotPolicy string) string {
	return acctest.ConfigCompose(testAccONTAPVolumeConfig_base(rName), fmt.Sprintf(`
resource "aws_fsx_ontap_volume" "test" {
  name                       = %[1]q
  junction_path              = "/%[1]s"
  size_in_megabytes          = 1024
  skip_final_backup          = true
  snapshot_policy            = %[2]q
  storage_efficiency_enabled = true
  storage_virtual_machine_id = aws_fsx_ontap_storage_virtual_machine.test.id
}
`, rName, snapshotPolicy))
}

func testAccONTAPVolumeConfig_storageEfficiency(rName string, storageEfficiencyEnabled bool) string {
	return acctest.ConfigCompose(testAccONTAPVolumeConfig_base(rName), fmt.Sprintf(`
resource "aws_fsx_ontap_volume" "test" {
  name                       = %[1]q
  junction_path              = "/%[1]s"
  size_in_megabytes          = 1024
  skip_final_backup          = true
  storage_efficiency_enabled = %[2]t
  storage_virtual_machine_id = aws_fsx_ontap_storage_virtual_machine.test.id
}
`, rName, storageEfficiencyEnabled))
}

func testAccONTAPVolumeConfig_tieringPolicy(rName, policy string, coolingPeriod int) string {
	return acctest.ConfigCompose(testAccONTAPVolumeConfig_base(rName), fmt.Sprintf(`
resource "aws_fsx_ontap_volume" "test" {
  name                       = %[1]q
  junction_path              = "/%[1]s"
  size_in_megabytes          = 1024
  skip_final_backup          = true
  storage_efficiency_enabled = true
  storage_virtual_machine_id = aws_fsx_ontap_storage_virtual_machine.test.id

  tiering_policy {
    name           = %[2]q
    cooling_period = %[3]d
  }
}
`, rName, policy, coolingPeriod))
}

func testAccONTAPVolumeConfig_tieringPolicyNoCooling(rName, policy string) string {
	return acctest.ConfigCompose(testAccONTAPVolumeConfig_base(rName), fmt.Sprintf(`
resource "aws_fsx_ontap_volume" "test" {
  name                       = %[1]q
  junction_path              = "/%[1]s"
  size_in_megabytes          = 1024
  skip_final_backup          = true
  storage_efficiency_enabled = true
  storage_virtual_machine_id = aws_fsx_ontap_storage_virtual_machine.test.id

  tiering_policy {
    name = %[2]q
  }
}
`, rName, policy))
}

func testAccONTAPVolumeConfig_tags1(rName, tagKey1, tagValue1 string) string {
	return acctest.ConfigCompose(testAccONTAPVolumeConfig_base(rName), fmt.Sprintf(`
resource "aws_fsx_ontap_volume" "test" {
  name                       = %[1]q
  junction_path              = "/%[1]s"
  size_in_megabytes          = 1024
  skip_final_backup          = true
  storage_efficiency_enabled = true
  storage_virtual_machine_id = aws_fsx_ontap_storage_virtual_machine.test.id

  tags = {
    %[2]q = %[3]q
  }
}
`, rName, tagKey1, tagValue1))
}

func testAccONTAPVolumeConfig_tags2(rName, tagKey1, tagValue1, tagKey2, tagValue2 string) string {
	return acctest.ConfigCompose(testAccONTAPVolumeConfig_base(rName), fmt.Sprintf(`
resource "aws_fsx_ontap_volume" "test" {
  name                       = %[1]q
  junction_path              = "/%[1]s"
  size_in_megabytes          = 1024
  skip_final_backup          = true
  storage_efficiency_enabled = true
  storage_virtual_machine_id = aws_fsx_ontap_storage_virtual_machine.test.id

  tags = {
    %[2]q = %[3]q
    %[4]q = %[5]q
  }
}
`, rName, tagKey1, tagValue1, tagKey2, tagValue2))
}

func testAccONTAPVolumeConfig_ontapStyle(rName string, style string) string {
	return acctest.ConfigCompose(testAccONTAPVolumeConfig_base(rName), fmt.Sprintf(`
resource "aws_fsx_ontap_volume" "test" {
  name                       = %[1]q
  junction_path              = "/%[1]s"
  size_in_megabytes          = 1048576
  skip_final_backup          = true
  volume_style               = %[2]q
  storage_efficiency_enabled = true
  storage_virtual_machine_id = aws_fsx_ontap_storage_virtual_machine.test.id
}
`, rName, style))
}

func testAccONTAPVolumeConfig_deleteConfig(rName, finalTagKey1, finalTagValue1, finalTagKey2, finalTagValue2 string) string {
	return acctest.ConfigCompose(testAccONTAPVolumeConfig_base(rName), fmt.Sprintf(`
resource "aws_fsx_ontap_volume" "test" {
  name                       = %[1]q
  junction_path              = "/%[1]s"
  size_in_megabytes          = 1024
  skip_final_backup          = false
  storage_efficiency_enabled = true
  storage_virtual_machine_id = aws_fsx_ontap_storage_virtual_machine.test.id

  final_backup_tags = {
    %[2]q = %[3]q
    %[4]q = %[5]q
  }
}
`, rName, finalTagKey1, finalTagValue1, finalTagKey2, finalTagValue2))
}
