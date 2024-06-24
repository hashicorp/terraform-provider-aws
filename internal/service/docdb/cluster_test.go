// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package docdb_test

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/docdb"
	awstypes "github.com/aws/aws-sdk-go-v2/service/docdb/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfdocdb "github.com/hashicorp/terraform-provider-aws/internal/service/docdb"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func init() {
	acctest.RegisterServiceErrorCheckFunc(names.DocDBServiceID, testAccErrorCheckSkip)
}

func testAccErrorCheckSkip(t *testing.T) resource.ErrorCheckFunc {
	return acctest.ErrorCheckSkipMessagesContaining(t,
		"Global clusters are not supported",
	)
}

func TestAccDocDBCluster_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var dbCluster awstypes.DBCluster
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_docdb_cluster.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.DocDBServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckClusterDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccClusterConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckClusterExists(ctx, resourceName, &dbCluster),
					resource.TestCheckNoResourceAttr(resourceName, names.AttrAllowMajorVersionUpgrade),
					resource.TestCheckNoResourceAttr(resourceName, names.AttrApplyImmediately),
					acctest.MatchResourceAttrRegionalARN(resourceName, names.AttrARN, "rds", regexache.MustCompile(fmt.Sprintf("cluster:%s", rName))),
					resource.TestCheckResourceAttr(resourceName, "availability_zones.#", acctest.Ct3),
					resource.TestCheckResourceAttr(resourceName, "backup_retention_period", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, names.AttrClusterIdentifier, rName),
					resource.TestCheckResourceAttr(resourceName, "cluster_identifier_prefix", ""),
					resource.TestCheckResourceAttr(resourceName, "cluster_members.#", acctest.Ct0),
					resource.TestCheckResourceAttrSet(resourceName, "cluster_resource_id"),
					resource.TestCheckResourceAttrSet(resourceName, "db_cluster_parameter_group_name"),
					resource.TestCheckResourceAttr(resourceName, "db_subnet_group_name", "default"),
					resource.TestCheckResourceAttr(resourceName, names.AttrDeletionProtection, acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "enabled_cloudwatch_logs_exports.#", acctest.Ct2),
					resource.TestCheckResourceAttr(resourceName, "enabled_cloudwatch_logs_exports.0", "audit"),
					resource.TestCheckResourceAttr(resourceName, "enabled_cloudwatch_logs_exports.1", "profiler"),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrEndpoint),
					resource.TestCheckResourceAttr(resourceName, names.AttrEngine, "docdb"),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrEngineVersion),
					resource.TestCheckNoResourceAttr(resourceName, names.AttrFinalSnapshotIdentifier),
					resource.TestCheckResourceAttr(resourceName, "global_cluster_identifier", ""),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrHostedZoneID),
					resource.TestCheckResourceAttr(resourceName, names.AttrKMSKeyID, ""),
					resource.TestCheckResourceAttr(resourceName, "master_password", "avoid-plaintext-passwords"),
					resource.TestCheckResourceAttr(resourceName, "master_username", "tfacctest"),
					resource.TestCheckResourceAttr(resourceName, names.AttrPort, "27017"),
					resource.TestCheckResourceAttrSet(resourceName, "preferred_backup_window"),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrPreferredMaintenanceWindow),
					resource.TestCheckResourceAttrSet(resourceName, "reader_endpoint"),
					resource.TestCheckResourceAttr(resourceName, "skip_final_snapshot", acctest.CtTrue),
					resource.TestCheckNoResourceAttr(resourceName, "snapshot_identifier"),
					resource.TestCheckResourceAttr(resourceName, names.AttrStorageEncrypted, acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, names.AttrStorageType, ""),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "vpc_security_group_ids.#", acctest.Ct1),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					names.AttrAllowMajorVersionUpgrade,
					names.AttrApplyImmediately,
					names.AttrFinalSnapshotIdentifier,
					"master_password",
					"skip_final_snapshot",
				},
			},
		},
	})
}

func TestAccDocDBCluster_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	var v awstypes.DBCluster
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_docdb_cluster.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.DocDBServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckClusterDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccClusterConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterExists(ctx, resourceName, &v),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfdocdb.ResourceCluster(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccDocDBCluster_identifierGenerated(t *testing.T) {
	ctx := acctest.Context(t)
	var v awstypes.DBCluster
	resourceName := "aws_docdb_cluster.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.DocDBServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckClusterDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccClusterConfig_identifierGenerated(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterExists(ctx, resourceName, &v),
					acctest.CheckResourceAttrNameGeneratedWithPrefix(resourceName, names.AttrClusterIdentifier, "tf-"),
					resource.TestCheckResourceAttr(resourceName, "cluster_identifier_prefix", "tf-"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					names.AttrAllowMajorVersionUpgrade,
					names.AttrApplyImmediately,
					names.AttrFinalSnapshotIdentifier,
					"master_password",
					"skip_final_snapshot",
				},
			},
		},
	})
}

func TestAccDocDBCluster_identifierPrefix(t *testing.T) {
	ctx := acctest.Context(t)
	var v awstypes.DBCluster
	resourceName := "aws_docdb_cluster.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.DocDBServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckClusterDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccClusterConfig_identifierPrefix("tf-acc-test-prefix-"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterExists(ctx, resourceName, &v),
					acctest.CheckResourceAttrNameFromPrefix(resourceName, names.AttrClusterIdentifier, "tf-acc-test-prefix-"),
					resource.TestCheckResourceAttr(resourceName, "cluster_identifier_prefix", "tf-acc-test-prefix-"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					names.AttrAllowMajorVersionUpgrade,
					names.AttrApplyImmediately,
					names.AttrFinalSnapshotIdentifier,
					"master_password",
					"skip_final_snapshot",
				},
			},
		},
	})
}

func TestAccDocDBCluster_tags(t *testing.T) {
	ctx := acctest.Context(t)
	var v awstypes.DBCluster
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_docdb_cluster.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.DocDBServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckClusterDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccClusterConfig_tags1(rName, acctest.CtKey1, acctest.CtValue1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey1, acctest.CtValue1),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					names.AttrAllowMajorVersionUpgrade,
					names.AttrApplyImmediately,
					names.AttrFinalSnapshotIdentifier,
					"master_password",
					"skip_final_snapshot",
				},
			},
			{
				Config: testAccClusterConfig_tags2(rName, acctest.CtKey1, acctest.CtValue1Updated, acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct2),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey1, acctest.CtValue1Updated),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey2, acctest.CtValue2),
				),
			},
			{
				Config: testAccClusterConfig_tags1(rName, acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey2, acctest.CtValue2),
				),
			},
		},
	})
}

func TestAccDocDBCluster_takeFinalSnapshot(t *testing.T) {
	ctx := acctest.Context(t)
	var v awstypes.DBCluster
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	snapshotName := fmt.Sprintf("%s-snapshot", rName)
	resourceName := "aws_docdb_cluster.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.DocDBServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckClusterDestroyWithFinalSnapshot(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccClusterConfig_finalSnapshot(rName, snapshotName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterExists(ctx, resourceName, &v),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					names.AttrAllowMajorVersionUpgrade,
					names.AttrApplyImmediately,
					names.AttrFinalSnapshotIdentifier,
					"master_password",
					"skip_final_snapshot",
				},
			},
		},
	})
}

// This is a regression test to make sure that we always cover the scenario as hightlighted in
// https://github.com/hashicorp/terraform/issues/11568
func TestAccDocDBCluster_missingUserNameCausesError(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.DocDBServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckClusterDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config:      testAccClusterConfig_noUsernameOrPassword(rName),
				ExpectError: regexache.MustCompile(`required field is not set`),
			},
		},
	})
}

func TestAccDocDBCluster_updateCloudWatchLogsExports(t *testing.T) {
	ctx := acctest.Context(t)
	var v awstypes.DBCluster
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_docdb_cluster.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.DocDBServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckClusterDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccClusterConfig_noCloudWatchLogs(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "enabled_cloudwatch_logs_exports.#", acctest.Ct0),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					names.AttrAllowMajorVersionUpgrade,
					names.AttrApplyImmediately,
					names.AttrFinalSnapshotIdentifier,
					"master_password",
					"skip_final_snapshot",
				},
			},
			{
				Config: testAccClusterConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "enabled_cloudwatch_logs_exports.#", acctest.Ct2),
					resource.TestCheckResourceAttr(resourceName, "enabled_cloudwatch_logs_exports.0", "audit"),
					resource.TestCheckResourceAttr(resourceName, "enabled_cloudwatch_logs_exports.1", "profiler"),
				),
			},
		},
	})
}

func TestAccDocDBCluster_kmsKey(t *testing.T) {
	ctx := acctest.Context(t)
	var v awstypes.DBCluster
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_docdb_cluster.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.DocDBServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckClusterDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccClusterConfig_kmsKey(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrKMSKeyID, "aws_kms_key.test", names.AttrARN),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					names.AttrAllowMajorVersionUpgrade,
					names.AttrApplyImmediately,
					names.AttrFinalSnapshotIdentifier,
					"master_password",
					"skip_final_snapshot",
				},
			},
		},
	})
}

func TestAccDocDBCluster_encrypted(t *testing.T) {
	ctx := acctest.Context(t)
	var v awstypes.DBCluster
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_docdb_cluster.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.DocDBServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckClusterDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccClusterConfig_encrypted(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, names.AttrStorageEncrypted, acctest.CtTrue),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					names.AttrAllowMajorVersionUpgrade,
					names.AttrApplyImmediately,
					names.AttrFinalSnapshotIdentifier,
					"master_password",
					"skip_final_snapshot",
				},
			},
		},
	})
}

func TestAccDocDBCluster_backupsUpdate(t *testing.T) {
	ctx := acctest.Context(t)
	var v awstypes.DBCluster
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_docdb_cluster.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.DocDBServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckClusterDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccClusterConfig_backups(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "backup_retention_period", "5"),
					resource.TestCheckResourceAttr(resourceName, "preferred_backup_window", "07:00-09:00"),
					resource.TestCheckResourceAttr(resourceName, names.AttrPreferredMaintenanceWindow, "tue:04:00-tue:04:30"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					names.AttrAllowMajorVersionUpgrade,
					names.AttrApplyImmediately,
					names.AttrFinalSnapshotIdentifier,
					"master_password",
					"skip_final_snapshot",
				},
			},
			{
				Config: testAccClusterConfig_backupsUpdate(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "backup_retention_period", acctest.Ct10),
					resource.TestCheckResourceAttr(resourceName, "preferred_backup_window", "03:00-09:00"),
					resource.TestCheckResourceAttr(resourceName, names.AttrPreferredMaintenanceWindow, "wed:01:00-wed:01:30"),
				),
			},
		},
	})
}

func TestAccDocDBCluster_pointInTimeRestore(t *testing.T) {
	ctx := acctest.Context(t)
	var dbCluster awstypes.DBCluster
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	sourceResourceName := "aws_docdb_cluster.test"
	resourceName := "aws_docdb_cluster.restore"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.DocDBServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckClusterDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccClusterConfig_pointInTimeRestoreSource(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckClusterExists(ctx, sourceResourceName, &dbCluster),
					testAccCheckClusterExists(ctx, resourceName, &dbCluster),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					names.AttrAllowMajorVersionUpgrade,
					names.AttrApplyImmediately,
					names.AttrFinalSnapshotIdentifier,
					"master_password",
					"restore_to_point_in_time",
					"skip_final_snapshot",
				},
			},
		},
	})
}

func TestAccDocDBCluster_port(t *testing.T) {
	ctx := acctest.Context(t)
	var dbCluster1, dbCluster2 awstypes.DBCluster
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_docdb_cluster.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.DocDBServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckClusterDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccClusterConfig_port(rName, 5432),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterExists(ctx, resourceName, &dbCluster1),
					resource.TestCheckResourceAttr(resourceName, names.AttrPort, "5432"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					names.AttrAllowMajorVersionUpgrade,
					names.AttrApplyImmediately,
					names.AttrFinalSnapshotIdentifier,
					"master_password",
					"skip_final_snapshot",
				},
			},
			{
				Config: testAccClusterConfig_port(rName, 2345),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterExists(ctx, resourceName, &dbCluster2),
					testAccCheckClusterRecreated(&dbCluster1, &dbCluster2),
					resource.TestCheckResourceAttr(resourceName, names.AttrPort, "2345"),
				),
			},
		},
	})
}

func TestAccDocDBCluster_deleteProtection(t *testing.T) {
	ctx := acctest.Context(t)
	var dbCluster awstypes.DBCluster
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_docdb_cluster.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.DocDBServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckClusterDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccClusterConfig_deleteProtection(rName, true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterExists(ctx, resourceName, &dbCluster),
					resource.TestCheckResourceAttr(resourceName, names.AttrDeletionProtection, acctest.CtTrue),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					names.AttrAllowMajorVersionUpgrade,
					names.AttrApplyImmediately,
					names.AttrFinalSnapshotIdentifier,
					"master_password",
					"skip_final_snapshot",
				},
			},
			{
				Config: testAccClusterConfig_deleteProtection(rName, false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterExists(ctx, resourceName, &dbCluster),
					resource.TestCheckResourceAttr(resourceName, names.AttrDeletionProtection, acctest.CtFalse),
				),
			},
			{
				Config: testAccClusterConfig_deleteProtection(rName, true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterExists(ctx, resourceName, &dbCluster),
					resource.TestCheckResourceAttr(resourceName, names.AttrDeletionProtection, acctest.CtTrue),
				),
			},
			{
				Config: testAccClusterConfig_deleteProtection(rName, false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterExists(ctx, resourceName, &dbCluster),
					resource.TestCheckResourceAttr(resourceName, names.AttrDeletionProtection, acctest.CtFalse),
				),
			},
		},
	})
}

func TestAccDocDBCluster_GlobalClusterIdentifier(t *testing.T) {
	ctx := acctest.Context(t)
	var dbCluster awstypes.DBCluster
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	globalClusterResourceName := "aws_docdb_cluster.test"
	resourceName := "aws_docdb_cluster.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheckGlobalCluster(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.DocDBServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckClusterDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccClusterConfig_globalIdentifier(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterExists(ctx, resourceName, &dbCluster),
					resource.TestCheckResourceAttrPair(resourceName, "global_cluster_identifier", globalClusterResourceName, names.AttrID),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					names.AttrAllowMajorVersionUpgrade,
					names.AttrApplyImmediately,
					names.AttrFinalSnapshotIdentifier,
					"master_password",
					"skip_final_snapshot",
				},
			},
		},
	})
}

func TestAccDocDBCluster_GlobalClusterIdentifier_Add(t *testing.T) {
	ctx := acctest.Context(t)
	var dbCluster awstypes.DBCluster
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_docdb_cluster.test"

	if acctest.Partition() == "aws-us-gov" {
		t.Skip("DocumentDB Global Cluster is not supported in GovCloud partition")
	}

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheckGlobalCluster(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.DocDBServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckClusterDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccClusterConfig_globalCompatible(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterExists(ctx, resourceName, &dbCluster),
					resource.TestCheckResourceAttr(resourceName, "global_cluster_identifier", ""),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					names.AttrAllowMajorVersionUpgrade,
					names.AttrApplyImmediately,
					names.AttrFinalSnapshotIdentifier,
					"master_password",
					"skip_final_snapshot",
				},
			},
			{
				Config:      testAccClusterConfig_globalIdentifier(rName),
				ExpectError: regexache.MustCompile(`existing DocumentDB Clusters cannot be added to an existing DocumentDB Global Cluster`),
			},
		},
	})
}

func TestAccDocDBCluster_GlobalClusterIdentifier_Remove(t *testing.T) {
	ctx := acctest.Context(t)
	var dbCluster awstypes.DBCluster
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	globalClusterResourceName := "aws_docdb_global_cluster.test"
	resourceName := "aws_docdb_cluster.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheckGlobalCluster(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.DocDBServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckClusterDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccClusterConfig_globalIdentifier(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterExists(ctx, resourceName, &dbCluster),
					resource.TestCheckResourceAttrPair(resourceName, "global_cluster_identifier", globalClusterResourceName, names.AttrID),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					names.AttrAllowMajorVersionUpgrade,
					names.AttrApplyImmediately,
					names.AttrFinalSnapshotIdentifier,
					"master_password",
					"skip_final_snapshot",
				},
			},
			{
				Config: testAccClusterConfig_globalCompatible(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterExists(ctx, resourceName, &dbCluster),
					resource.TestCheckResourceAttr(resourceName, "global_cluster_identifier", ""),
				),
			},
		},
	})
}

func TestAccDocDBCluster_GlobalClusterIdentifier_Update(t *testing.T) {
	ctx := acctest.Context(t)
	var dbCluster awstypes.DBCluster
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	globalClusterResourceName1 := "aws_docdb_global_cluster.test.0"
	globalClusterResourceName2 := "aws_docdb_global_cluster.test.1"
	resourceName := "aws_docdb_cluster.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheckGlobalCluster(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.DocDBServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckClusterDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccClusterConfig_globalIdentifierUpdate(rName, globalClusterResourceName1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterExists(ctx, resourceName, &dbCluster),
					resource.TestCheckResourceAttrPair(resourceName, "global_cluster_identifier", globalClusterResourceName1, names.AttrID),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					names.AttrAllowMajorVersionUpgrade,
					names.AttrApplyImmediately,
					names.AttrFinalSnapshotIdentifier,
					"master_password",
					"skip_final_snapshot",
				},
			},
			{
				Config:      testAccClusterConfig_globalIdentifierUpdate(rName, globalClusterResourceName2),
				ExpectError: regexache.MustCompile(`existing DocumentDB Clusters cannot be migrated between existing DocumentDB Global Clusters`),
			},
		},
	})
}

func TestAccDocDBCluster_GlobalClusterIdentifier_PrimarySecondaryClusters(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	ctx := acctest.Context(t)
	var providers []*schema.Provider
	var primaryDbCluster, secondaryDbCluster awstypes.DBCluster
	rNameGlobal := sdkacctest.RandomWithPrefix("tf-acc-test-global")
	rNamePrimary := sdkacctest.RandomWithPrefix("tf-acc-test-primary")
	rNameSecondary := sdkacctest.RandomWithPrefix("tf-acc-test-secondary")
	resourceNamePrimary := "aws_docdb_cluster.primary"
	resourceNameSecondary := "aws_docdb_cluster.secondary"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckMultipleRegion(t, 2)
			testAccPreCheckGlobalCluster(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.DocDBServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5FactoriesPlusProvidersAlternate(ctx, t, &providers),
		CheckDestroy:             testAccCheckClusterDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccClusterConfig_globalIdentifierPrimarySecondary(rNameGlobal, rNamePrimary, rNameSecondary),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterExistsProvider(ctx, resourceNamePrimary, &primaryDbCluster, acctest.RegionProviderFunc(acctest.Region(), &providers)),
					testAccCheckClusterExistsProvider(ctx, resourceNameSecondary, &secondaryDbCluster, acctest.RegionProviderFunc(acctest.AlternateRegion(), &providers)),
				),
			},
		},
	})
}

func TestAccDocDBCluster_updateEngineMajorVersion(t *testing.T) {
	// https://docs.aws.amazon.com/documentdb/latest/developerguide/docdb-mvu.html.
	acctest.Skip(t, "Amazon DocumentDB has identified an issue and is temporarily disallowing major version upgrades (MVU) in all regions.")

	ctx := acctest.Context(t)
	var dbCluster awstypes.DBCluster
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_docdb_cluster.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.DocDBServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckClusterDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccClusterConfig_engineVersion(rName, "4.0.0"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckClusterExists(ctx, resourceName, &dbCluster),
					resource.TestCheckResourceAttr(resourceName, names.AttrAllowMajorVersionUpgrade, acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, names.AttrApplyImmediately, acctest.CtTrue),
					acctest.MatchResourceAttrRegionalARN(resourceName, names.AttrARN, "rds", regexache.MustCompile(fmt.Sprintf("cluster:%s", rName))),
					resource.TestCheckResourceAttr(resourceName, "availability_zones.#", acctest.Ct3),
					resource.TestCheckResourceAttr(resourceName, "backup_retention_period", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, names.AttrClusterIdentifier, rName),
					resource.TestCheckResourceAttr(resourceName, "cluster_identifier_prefix", ""),
					resource.TestCheckResourceAttr(resourceName, "cluster_members.#", acctest.Ct0),
					resource.TestCheckResourceAttrSet(resourceName, "cluster_resource_id"),
					resource.TestCheckResourceAttr(resourceName, "db_cluster_parameter_group_name", "default.docdb4.0"),
					resource.TestCheckResourceAttr(resourceName, "db_subnet_group_name", "default"),
					resource.TestCheckResourceAttr(resourceName, names.AttrDeletionProtection, acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "enabled_cloudwatch_logs_exports.#", acctest.Ct0),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrEndpoint),
					resource.TestCheckResourceAttr(resourceName, names.AttrEngine, "docdb"),
					resource.TestCheckResourceAttr(resourceName, names.AttrEngineVersion, "4.0.0"),
					resource.TestCheckNoResourceAttr(resourceName, names.AttrFinalSnapshotIdentifier),
					resource.TestCheckResourceAttr(resourceName, "global_cluster_identifier", ""),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrHostedZoneID),
					resource.TestCheckResourceAttr(resourceName, names.AttrKMSKeyID, ""),
					resource.TestCheckResourceAttr(resourceName, "master_password", "avoid-plaintext-passwords"),
					resource.TestCheckResourceAttr(resourceName, "master_username", "tfacctest"),
					resource.TestCheckResourceAttr(resourceName, names.AttrPort, "27017"),
					resource.TestCheckResourceAttrSet(resourceName, "preferred_backup_window"),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrPreferredMaintenanceWindow),
					resource.TestCheckResourceAttrSet(resourceName, "reader_endpoint"),
					resource.TestCheckResourceAttr(resourceName, "skip_final_snapshot", acctest.CtTrue),
					resource.TestCheckNoResourceAttr(resourceName, "snapshot_identifier"),
					resource.TestCheckResourceAttr(resourceName, names.AttrStorageEncrypted, acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, names.AttrStorageType, ""),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "vpc_security_group_ids.#", acctest.Ct1),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					names.AttrAllowMajorVersionUpgrade,
					names.AttrApplyImmediately,
					names.AttrFinalSnapshotIdentifier,
					"master_password",
					"skip_final_snapshot",
				},
			},
			{
				Config: testAccClusterConfig_engineVersion(rName, "5.0.0"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterExists(ctx, resourceName, &dbCluster),
					resource.TestCheckResourceAttr(resourceName, "cluster_members.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "db_cluster_parameter_group_name", "default.docdb5.0"),
					resource.TestCheckResourceAttr(resourceName, names.AttrEngineVersion, "5.0.0"),
				),
			},
		},
	})
}

func TestAccDocDBCluster_storageType(t *testing.T) {
	ctx := acctest.Context(t)
	var dbCluster awstypes.DBCluster
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_docdb_cluster.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.DocDBServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckClusterDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccClusterConfig_storageType(rName, "standard"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterExists(ctx, resourceName, &dbCluster),
					resource.TestCheckResourceAttr(resourceName, names.AttrStorageType, ""),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					names.AttrAllowMajorVersionUpgrade,
					names.AttrApplyImmediately,
					names.AttrFinalSnapshotIdentifier,
					"master_password",
					"skip_final_snapshot",
				},
			},
			{
				Config: testAccClusterConfig_storageType(rName, "iopt1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterExists(ctx, resourceName, &dbCluster),
					resource.TestCheckResourceAttr(resourceName, names.AttrStorageType, "iopt1"),
				),
			},
			{
				Config: testAccClusterConfig_storageType(rName, "standard"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterExists(ctx, resourceName, &dbCluster),
					resource.TestCheckResourceAttr(resourceName, names.AttrStorageType, ""),
				),
			},
		},
	})
}

func testAccCheckClusterDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).DocDBClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_docdb_cluster" {
				continue
			}

			_, err := tfdocdb.FindDBClusterByID(ctx, conn, rs.Primary.ID)

			if tfresource.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("DocumentDB Cluster %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccCheckClusterExists(ctx context.Context, n string, v *awstypes.DBCluster) resource.TestCheckFunc {
	return testAccCheckClusterExistsProvider(ctx, n, v, func() *schema.Provider { return acctest.Provider })
}

func testAccCheckClusterExistsProvider(ctx context.Context, n string, v *awstypes.DBCluster, providerF func() *schema.Provider) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := providerF().Meta().(*conns.AWSClient).DocDBClient(ctx)

		output, err := tfdocdb.FindDBClusterByID(ctx, conn, rs.Primary.ID)

		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

func testAccCheckClusterDestroyWithFinalSnapshot(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).DocDBClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_docdb_cluster" {
				continue
			}

			finalSnapshotID := rs.Primary.Attributes[names.AttrFinalSnapshotIdentifier]
			_, err := conn.DeleteDBClusterSnapshot(ctx, &docdb.DeleteDBClusterSnapshotInput{
				DBClusterSnapshotIdentifier: aws.String(finalSnapshotID),
			})

			if err != nil {
				return err
			}

			_, err = tfdocdb.FindDBClusterByID(ctx, conn, rs.Primary.ID)

			if tfresource.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("DocumentDB Cluster %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccCheckClusterRecreated(i, j *awstypes.DBCluster) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if aws.ToTime(i.ClusterCreateTime).Equal(aws.ToTime(j.ClusterCreateTime)) {
			return errors.New("DocumentDB Cluster was not recreated")
		}

		return nil
	}
}

func testAccClusterConfig_basic(rName string) string {
	return acctest.ConfigCompose(acctest.ConfigAvailableAZsNoOptIn(), fmt.Sprintf(`
resource "aws_docdb_cluster" "test" {
  cluster_identifier = %[1]q

  availability_zones = [
    data.aws_availability_zones.available.names[0],
    data.aws_availability_zones.available.names[1],
    data.aws_availability_zones.available.names[2]
  ]

  master_password     = "avoid-plaintext-passwords"
  master_username     = "tfacctest"
  skip_final_snapshot = true

  enabled_cloudwatch_logs_exports = [
    "audit",
    "profiler",
  ]
}
`, rName))
}

func testAccClusterConfig_identifierGenerated() string {
	return `
resource "aws_docdb_cluster" "test" {
  master_password     = "avoid-plaintext-passwords"
  master_username     = "tfacctest"
  skip_final_snapshot = true
}
`
}

func testAccClusterConfig_identifierPrefix(prefix string) string {
	return fmt.Sprintf(`
resource "aws_docdb_cluster" "test" {
  cluster_identifier_prefix = %[1]q
  master_password           = "avoid-plaintext-passwords"
  master_username           = "tfacctest"
  skip_final_snapshot       = true
}
`, prefix)
}

func testAccClusterConfig_tags1(rName, tagKey1, tagValue1 string) string {
	return fmt.Sprintf(`
resource "aws_docdb_cluster" "test" {
  cluster_identifier  = %[1]q
  master_password     = "avoid-plaintext-passwords"
  master_username     = "tfacctest"
  skip_final_snapshot = true

  tags = {
    %[2]q = %[3]q
  }
}
`, rName, tagKey1, tagValue1)
}

func testAccClusterConfig_tags2(rName, tagKey1, tagValue1, tagKey2, tagValue2 string) string {
	return fmt.Sprintf(`
resource "aws_docdb_cluster" "test" {
  cluster_identifier  = %[1]q
  master_password     = "avoid-plaintext-passwords"
  master_username     = "tfacctest"
  skip_final_snapshot = true

  tags = {
    %[2]q = %[3]q
    %[4]q = %[5]q
  }
}
`, rName, tagKey1, tagValue1, tagKey2, tagValue2)
}

func testAccClusterConfig_finalSnapshot(rName, snapshotName string) string {
	return acctest.ConfigCompose(acctest.ConfigAvailableAZsNoOptIn(), fmt.Sprintf(`
resource "aws_docdb_cluster" "test" {
  cluster_identifier = %[1]q

  availability_zones = [
    data.aws_availability_zones.available.names[0],
    data.aws_availability_zones.available.names[1],
    data.aws_availability_zones.available.names[2]
  ]

  master_password           = "avoid-plaintext-passwords"
  master_username           = "tfacctest"
  final_snapshot_identifier = %[2]q
}
`, rName, snapshotName))
}

func testAccClusterConfig_noUsernameOrPassword(rName string) string {
	return acctest.ConfigCompose(acctest.ConfigAvailableAZsNoOptIn(), fmt.Sprintf(`
resource "aws_docdb_cluster" "test" {
  cluster_identifier = %[1]q

  availability_zones = [
    data.aws_availability_zones.available.names[0],
    data.aws_availability_zones.available.names[1],
    data.aws_availability_zones.available.names[2]
  ]

  skip_final_snapshot = true
}
`, rName))
}

func testAccClusterConfig_noCloudWatchLogs(rName string) string {
	return acctest.ConfigCompose(acctest.ConfigAvailableAZsNoOptIn(), fmt.Sprintf(`
resource "aws_docdb_cluster" "test" {
  cluster_identifier = %[1]q

  availability_zones = [
    data.aws_availability_zones.available.names[0],
    data.aws_availability_zones.available.names[1],
    data.aws_availability_zones.available.names[2]
  ]

  master_password     = "avoid-plaintext-passwords"
  master_username     = "tfacctest"
  skip_final_snapshot = true
}
`, rName))
}

func testAccClusterConfig_kmsKey(rName string) string {
	return acctest.ConfigCompose(acctest.ConfigAvailableAZsNoOptIn(), fmt.Sprintf(`
resource "aws_kms_key" "test" {
  description = %[1]q

  policy = <<POLICY
{
  "Version": "2012-10-17",
  "Id": "kms-tf-1",
  "Statement": [
    {
      "Sid": "Enable IAM User Permissions",
      "Effect": "Allow",
      "Principal": {
        "AWS": "*"
      },
      "Action": "kms:*",
      "Resource": "*"
    }
  ]
}
POLICY
}

resource "aws_docdb_cluster" "test" {
  cluster_identifier = %[1]q
  availability_zones = [
    data.aws_availability_zones.available.names[0],
    data.aws_availability_zones.available.names[1],
    data.aws_availability_zones.available.names[2]
  ]

  master_password     = "avoid-plaintext-passwords"
  master_username     = "tfacctest"
  storage_encrypted   = true
  kms_key_id          = aws_kms_key.test.arn
  skip_final_snapshot = true
}
`, rName))
}

func testAccClusterConfig_encrypted(rName string) string {
	return acctest.ConfigCompose(acctest.ConfigAvailableAZsNoOptIn(), fmt.Sprintf(`
resource "aws_docdb_cluster" "test" {
  cluster_identifier = %[1]q

  availability_zones = [
    data.aws_availability_zones.available.names[0],
    data.aws_availability_zones.available.names[1],
    data.aws_availability_zones.available.names[2]
  ]

  master_password     = "avoid-plaintext-passwords"
  master_username     = "tfacctest"
  storage_encrypted   = true
  skip_final_snapshot = true
}
`, rName))
}

func testAccClusterConfig_backups(rName string) string {
	return acctest.ConfigCompose(acctest.ConfigAvailableAZsNoOptIn(), fmt.Sprintf(`
resource "aws_docdb_cluster" "test" {
  cluster_identifier = %[1]q

  availability_zones = [
    data.aws_availability_zones.available.names[0],
    data.aws_availability_zones.available.names[1],
    data.aws_availability_zones.available.names[2]
  ]

  master_password              = "avoid-plaintext-passwords"
  master_username              = "tfacctest"
  backup_retention_period      = 5
  preferred_backup_window      = "07:00-09:00"
  preferred_maintenance_window = "tue:04:00-tue:04:30"
  skip_final_snapshot          = true
}
`, rName))
}

func testAccClusterConfig_backupsUpdate(rName string) string {
	return acctest.ConfigCompose(acctest.ConfigAvailableAZsNoOptIn(), fmt.Sprintf(`
resource "aws_docdb_cluster" "test" {
  cluster_identifier = %[1]q

  availability_zones = [
    data.aws_availability_zones.available.names[0],
    data.aws_availability_zones.available.names[1],
    data.aws_availability_zones.available.names[2]
  ]

  master_password              = "avoid-plaintext-passwords"
  master_username              = "tfacctest"
  backup_retention_period      = 10
  preferred_backup_window      = "03:00-09:00"
  preferred_maintenance_window = "wed:01:00-wed:01:30"
  apply_immediately            = true
  skip_final_snapshot          = true
}
`, rName))
}

func testAccClusterConfig_port(rName string, port int) string {
	return acctest.ConfigCompose(acctest.ConfigAvailableAZsNoOptIn(), fmt.Sprintf(`
resource "aws_docdb_cluster" "test" {
  availability_zones = [
    data.aws_availability_zones.available.names[0],
    data.aws_availability_zones.available.names[1],
    data.aws_availability_zones.available.names[2]
  ]

  cluster_identifier  = %[1]q
  engine              = "docdb"
  master_password     = "avoid-plaintext-passwords"
  master_username     = "tfacctest"
  port                = %[2]d
  skip_final_snapshot = true
}
`, rName, port))
}

func testAccClusterConfig_baseForPITR(rName string) string {
	return acctest.ConfigCompose(acctest.ConfigAvailableAZsNoOptIn(), fmt.Sprintf(`
resource "aws_docdb_cluster" "test" {
  cluster_identifier = %[1]q

  availability_zones = [
    data.aws_availability_zones.available.names[0],
    data.aws_availability_zones.available.names[1],
    data.aws_availability_zones.available.names[2]
  ]

  master_password     = "avoid-plaintext-passwords"
  master_username     = "tfacctest"
  skip_final_snapshot = true

  enabled_cloudwatch_logs_exports = [
    "audit",
    "profiler",
  ]
}
`, rName))
}

func testAccClusterConfig_pointInTimeRestoreSource(rName string) string {
	return acctest.ConfigCompose(testAccClusterConfig_baseForPITR(rName), fmt.Sprintf(`
resource "aws_docdb_cluster" "restore" {
  cluster_identifier = "%[1]s-restore"

  restore_to_point_in_time {
    source_cluster_identifier  = aws_docdb_cluster.test.cluster_identifier
    restore_type               = "full-copy"
    use_latest_restorable_time = true
  }

  skip_final_snapshot = true
}
`, rName))
}

func testAccClusterConfig_deleteProtection(rName string, isProtected bool) string {
	return fmt.Sprintf(`
resource "aws_docdb_cluster" "test" {
  cluster_identifier  = %[1]q
  master_username     = "tfacctest"
  master_password     = "avoid-plaintext-passwords"
  skip_final_snapshot = true
  deletion_protection = %[2]t
}
`, rName, isProtected)
}

func testAccClusterConfig_globalIdentifier(rName string) string {
	return fmt.Sprintf(`
resource "aws_docdb_global_cluster" "test" {
  engine_version            = "4.0.0" # version compatible
  engine                    = "docdb"
  global_cluster_identifier = %[1]q
}

resource "aws_docdb_cluster" "test" {
  cluster_identifier        = %[1]q
  global_cluster_identifier = aws_docdb_global_cluster.test.id
  engine_version            = aws_docdb_global_cluster.test.engine_version
  master_password           = "avoid-plaintext-passwords"
  master_username           = "tfacctest"
  skip_final_snapshot       = true
}
`, rName)
}

func testAccClusterConfig_globalCompatible(rName string) string {
	return fmt.Sprintf(`
resource "aws_docdb_cluster" "test" {
  cluster_identifier  = %[1]q
  engine_version      = "4.0.0" # version compatible with global
  master_password     = "avoid-plaintext-passwords"
  master_username     = "tfacctest"
  skip_final_snapshot = true
}
`, rName)
}

func testAccClusterConfig_globalIdentifierUpdate(rName, globalClusterIdentifierResourceName string) string {
	return fmt.Sprintf(`
resource "aws_docdb_global_cluster" "test" {
  count                     = 2
  engine                    = "docdb"
  engine_version            = "4.0.0" # version compatible with global
  global_cluster_identifier = "%[1]s-${count.index}"
}

resource "aws_docdb_cluster" "test" {
  cluster_identifier        = %[1]q
  global_cluster_identifier = %[2]s.id
  engine_version            = %[2]s.engine_version
  master_password           = "avoid-plaintext-passwords"
  master_username           = "tfacctest"
  skip_final_snapshot       = true
}
`, rName, globalClusterIdentifierResourceName)
}

func testAccClusterConfig_globalIdentifierPrimarySecondary(rNameGlobal, rNamePrimary, rNameSecondary string) string {
	return acctest.ConfigCompose(
		acctest.ConfigMultipleRegionProvider(2),
		fmt.Sprintf(`
data "aws_availability_zones" "alternate" {
  provider = "awsalternate"
  state    = "available"

  filter {
    name   = "opt-in-status"
    values = ["opt-in-not-required"]
  }
}

resource "aws_docdb_global_cluster" "test" {
  global_cluster_identifier = %[1]q
  engine                    = "docdb"
  engine_version            = "4.0.0"
}

resource "aws_docdb_cluster" "primary" {
  cluster_identifier        = %[2]q
  master_password           = "avoid-plaintext-passwords"
  master_username           = "tfacctest"
  skip_final_snapshot       = true
  global_cluster_identifier = aws_docdb_global_cluster.test.id
  engine                    = aws_docdb_global_cluster.test.engine
  engine_version            = aws_docdb_global_cluster.test.engine_version
}

resource "aws_docdb_cluster_instance" "primary" {
  identifier         = %[2]q
  cluster_identifier = aws_docdb_cluster.primary.id
  instance_class     = "db.r5.large"
}

resource "aws_vpc" "alternate" {
  provider   = "awsalternate"
  cidr_block = "10.0.0.0/16"

  tags = {
    Name = %[3]q
  }
}

resource "aws_subnet" "alternate" {
  provider          = "awsalternate"
  count             = 3
  vpc_id            = aws_vpc.alternate.id
  availability_zone = data.aws_availability_zones.alternate.names[count.index]
  cidr_block        = "10.0.${count.index}.0/24"

  tags = {
    Name = %[3]q
  }
}

resource "aws_docdb_subnet_group" "alternate" {
  provider   = "awsalternate"
  name       = %[3]q
  subnet_ids = aws_subnet.alternate[*].id
}

resource "aws_docdb_cluster" "secondary" {
  provider                  = "awsalternate"
  cluster_identifier        = %[3]q
  skip_final_snapshot       = true
  db_subnet_group_name      = aws_docdb_subnet_group.alternate.name
  global_cluster_identifier = aws_docdb_global_cluster.test.id
  engine                    = aws_docdb_global_cluster.test.engine
  engine_version            = aws_docdb_global_cluster.test.engine_version
  depends_on                = [aws_docdb_cluster_instance.primary]
}

resource "aws_docdb_cluster_instance" "secondary" {
  provider           = "awsalternate"
  identifier         = %[3]q
  cluster_identifier = aws_docdb_cluster.secondary.id
  instance_class     = "db.r5.large"
}
`, rNameGlobal, rNamePrimary, rNameSecondary))
}

func testAccClusterConfig_engineVersion(rName, engineVersion string) string {
	return fmt.Sprintf(`
resource "aws_docdb_cluster" "test" {
  cluster_identifier          = %[1]q
  engine_version              = %[2]q
  master_password             = "avoid-plaintext-passwords"
  master_username             = "tfacctest"
  skip_final_snapshot         = true
  apply_immediately           = true
  allow_major_version_upgrade = true
}

data "aws_docdb_orderable_db_instance" "test" {
  engine                     = aws_docdb_cluster.test.engine
  preferred_instance_classes = ["db.t3.medium", "db.4tg.medium", "db.r5.large", "db.r6g.large"]
}

resource "aws_docdb_cluster_instance" "test" {
  identifier         = %[1]q
  cluster_identifier = aws_docdb_cluster.test.id
  instance_class     = data.aws_docdb_orderable_db_instance.test.instance_class
}
`, rName, engineVersion)
}

func testAccClusterConfig_storageType(rName, storageType string) string {
	return acctest.ConfigCompose(acctest.ConfigAvailableAZsNoOptIn(), fmt.Sprintf(`
resource "aws_docdb_cluster" "test" {
  availability_zones = [
    data.aws_availability_zones.available.names[0],
    data.aws_availability_zones.available.names[1],
    data.aws_availability_zones.available.names[2]
  ]

  cluster_identifier  = %[1]q
  engine              = "docdb"
  master_password     = "avoid-plaintext-passwords"
  master_username     = "tfacctest"
  storage_type        = %[2]q
  apply_immediately   = true
  skip_final_snapshot = true
}
`, rName, storageType))
}
