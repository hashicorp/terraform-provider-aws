// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package docdb_test

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/docdb"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfdocdb "github.com/hashicorp/terraform-provider-aws/internal/service/docdb"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

func init() {
	acctest.RegisterServiceErrorCheckFunc(docdb.EndpointsID, testAccErrorCheckSkip)
}

func testAccErrorCheckSkip(t *testing.T) resource.ErrorCheckFunc {
	return acctest.ErrorCheckSkipMessagesContaining(t,
		"Global clusters are not supported",
	)
}

func TestAccDocDBCluster_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var dbCluster docdb.DBCluster
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_docdb_cluster.default"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, docdb.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckClusterDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccClusterConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterExists(ctx, resourceName, &dbCluster),
					acctest.MatchResourceAttrRegionalARN(resourceName, "arn", "rds", regexache.MustCompile(`cluster:.+`)),
					resource.TestCheckResourceAttr(resourceName, "storage_encrypted", "false"),
					resource.TestCheckResourceAttrSet(resourceName, "db_cluster_parameter_group_name"),
					resource.TestCheckResourceAttrSet(resourceName, "reader_endpoint"),
					resource.TestCheckResourceAttrSet(resourceName, "cluster_resource_id"),
					resource.TestCheckResourceAttr(resourceName, "engine", "docdb"),
					resource.TestCheckResourceAttrSet(resourceName, "engine_version"),
					resource.TestCheckResourceAttrSet(resourceName, "hosted_zone_id"),
					resource.TestCheckResourceAttr(resourceName, "enabled_cloudwatch_logs_exports.0", "audit"),
					resource.TestCheckResourceAttr(resourceName, "enabled_cloudwatch_logs_exports.1", "profiler"),
					resource.TestCheckResourceAttr(resourceName, "deletion_protection", "false"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"allow_major_version_upgrade",
					"apply_immediately",
					"cluster_identifier_prefix",
					"final_snapshot_identifier",
					"master_password",
					"skip_final_snapshot",
				},
			},
		},
	})
}

func TestAccDocDBCluster_namePrefix(t *testing.T) {
	ctx := acctest.Context(t)
	var v docdb.DBCluster

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, docdb.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckClusterDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccClusterConfig_namePrefix(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterExists(ctx, "aws_docdb_cluster.test", &v),
					resource.TestMatchResourceAttr(
						"aws_docdb_cluster.test", "cluster_identifier", regexache.MustCompile("^tf-test-")),
				),
			},
			{
				ResourceName:      "aws_docdb_cluster.test",
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"allow_major_version_upgrade",
					"apply_immediately",
					"cluster_identifier_prefix",
					"final_snapshot_identifier",
					"master_password",
					"skip_final_snapshot",
				},
			},
		},
	})
}

func TestAccDocDBCluster_generatedName(t *testing.T) {
	ctx := acctest.Context(t)
	var v docdb.DBCluster

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, docdb.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckClusterDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccClusterConfig_generatedName(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterExists(ctx, "aws_docdb_cluster.test", &v),
					resource.TestMatchResourceAttr(
						"aws_docdb_cluster.test", "cluster_identifier", regexache.MustCompile("^tf-")),
				),
			},
			{
				ResourceName:      "aws_docdb_cluster.test",
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"allow_major_version_upgrade",
					"apply_immediately",
					"cluster_identifier_prefix",
					"final_snapshot_identifier",
					"master_password",
					"skip_final_snapshot",
				},
			},
		},
	})
}

func TestAccDocDBCluster_GlobalClusterIdentifier(t *testing.T) {
	ctx := acctest.Context(t)
	var dbCluster1 docdb.DBCluster

	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	globalClusterResourceName := "aws_docdb_cluster.test"
	resourceName := "aws_docdb_cluster.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheckGlobalCluster(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, docdb.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckClusterDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccClusterConfig_globalIdentifier(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterExists(ctx, resourceName, &dbCluster1),
					resource.TestCheckResourceAttrPair(resourceName, "global_cluster_identifier", globalClusterResourceName, "id"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"allow_major_version_upgrade",
					"apply_immediately",
					"cluster_identifier_prefix",
					"final_snapshot_identifier",
					"master_password",
					"skip_final_snapshot",
				},
			},
		},
	})
}

func TestAccDocDBCluster_GlobalClusterIdentifier_Add(t *testing.T) {
	ctx := acctest.Context(t)
	var dbCluster1 docdb.DBCluster

	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_docdb_cluster.test"

	if acctest.Partition() == "aws-us-gov" {
		t.Skip("DocumentDB Global Cluster is not supported in GovCloud partition")
	}

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheckGlobalCluster(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, docdb.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckClusterDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccClusterConfig_globalCompatible(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterExists(ctx, resourceName, &dbCluster1),
					resource.TestCheckResourceAttr(resourceName, "global_cluster_identifier", ""),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"allow_major_version_upgrade",
					"apply_immediately",
					"cluster_identifier_prefix",
					"final_snapshot_identifier",
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
	var dbCluster1 docdb.DBCluster

	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	globalClusterResourceName := "aws_docdb_global_cluster.test"
	resourceName := "aws_docdb_cluster.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheckGlobalCluster(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, docdb.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckClusterDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccClusterConfig_globalIdentifier(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterExists(ctx, resourceName, &dbCluster1),
					resource.TestCheckResourceAttrPair(resourceName, "global_cluster_identifier", globalClusterResourceName, "id"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"allow_major_version_upgrade",
					"apply_immediately",
					"cluster_identifier_prefix",
					"final_snapshot_identifier",
					"master_password",
					"skip_final_snapshot",
				},
			},
			{
				Config: testAccClusterConfig_globalCompatible(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterExists(ctx, resourceName, &dbCluster1),
					resource.TestCheckResourceAttr(resourceName, "global_cluster_identifier", ""),
				),
			},
		},
	})
}

func TestAccDocDBCluster_GlobalClusterIdentifier_Update(t *testing.T) {
	ctx := acctest.Context(t)
	var dbCluster1 docdb.DBCluster

	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	globalClusterResourceName1 := "aws_docdb_global_cluster.test.0"
	globalClusterResourceName2 := "aws_docdb_global_cluster.test.1"
	resourceName := "aws_docdb_cluster.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheckGlobalCluster(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, docdb.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckClusterDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccClusterConfig_globalIdentifierUpdate(rName, globalClusterResourceName1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterExists(ctx, resourceName, &dbCluster1),
					resource.TestCheckResourceAttrPair(resourceName, "global_cluster_identifier", globalClusterResourceName1, "id"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"allow_major_version_upgrade",
					"apply_immediately",
					"cluster_identifier_prefix",
					"final_snapshot_identifier",
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
	var primaryDbCluster, secondaryDbCluster docdb.DBCluster

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
		ErrorCheck:               acctest.ErrorCheck(t, docdb.EndpointsID),
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

func TestAccDocDBCluster_takeFinalSnapshot(t *testing.T) {
	ctx := acctest.Context(t)
	var v docdb.DBCluster
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	snapshotName := fmt.Sprintf("%s-snapshot", rName)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, docdb.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckClusterDestroyWithFinalSnapshot(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccClusterConfig_finalSnapshot(rName, snapshotName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterExists(ctx, "aws_docdb_cluster.default", &v),
				),
			},
			{
				ResourceName:      "aws_docdb_cluster.default",
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"allow_major_version_upgrade",
					"apply_immediately",
					"cluster_identifier_prefix",
					"final_snapshot_identifier",
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
		ErrorCheck:               acctest.ErrorCheck(t, docdb.EndpointsID),
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

func TestAccDocDBCluster_updateTags(t *testing.T) {
	ctx := acctest.Context(t)
	var v docdb.DBCluster
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, docdb.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckClusterDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccClusterConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterExists(ctx, "aws_docdb_cluster.default", &v),
					resource.TestCheckResourceAttr(
						"aws_docdb_cluster.default", "tags.%", "1"),
				),
			},
			{
				ResourceName:      "aws_docdb_cluster.default",
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"allow_major_version_upgrade",
					"apply_immediately",
					"cluster_identifier_prefix",
					"final_snapshot_identifier",
					"master_password",
					"skip_final_snapshot",
				},
			},
			{
				Config: testAccClusterConfig_updatedTags(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterExists(ctx, "aws_docdb_cluster.default", &v),
					resource.TestCheckResourceAttr(
						"aws_docdb_cluster.default", "tags.%", "2"),
				),
			},
		},
	})
}

func TestAccDocDBCluster_updateCloudWatchLogsExports(t *testing.T) {
	ctx := acctest.Context(t)
	var v docdb.DBCluster
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, docdb.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckClusterDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccClusterConfig_noCloudWatchLogs(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterExists(ctx, "aws_docdb_cluster.default", &v),
				),
			},
			{
				ResourceName:      "aws_docdb_cluster.default",
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"allow_major_version_upgrade",
					"apply_immediately",
					"cluster_identifier_prefix",
					"final_snapshot_identifier",
					"master_password",
					"skip_final_snapshot",
				},
			},
			{
				Config: testAccClusterConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterExists(ctx, "aws_docdb_cluster.default", &v),
					resource.TestCheckResourceAttr("aws_docdb_cluster.default",
						"enabled_cloudwatch_logs_exports.0", "audit"),
				),
			},
		},
	})
}

func TestAccDocDBCluster_kmsKey(t *testing.T) {
	ctx := acctest.Context(t)
	var v docdb.DBCluster
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, docdb.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckClusterDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccClusterConfig_kmsKey(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterExists(ctx, "aws_docdb_cluster.default", &v),
					resource.TestCheckResourceAttrPair("aws_docdb_cluster.default", "kms_key_id", "aws_kms_key.foo", "arn"),
				),
			},
			{
				ResourceName:      "aws_docdb_cluster.default",
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"allow_major_version_upgrade",
					"apply_immediately",
					"cluster_identifier_prefix",
					"final_snapshot_identifier",
					"master_password",
					"skip_final_snapshot",
				},
			},
		},
	})
}

func TestAccDocDBCluster_encrypted(t *testing.T) {
	ctx := acctest.Context(t)
	var v docdb.DBCluster
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, docdb.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckClusterDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccClusterConfig_encrypted(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterExists(ctx, "aws_docdb_cluster.default", &v),
					resource.TestCheckResourceAttr("aws_docdb_cluster.default", "storage_encrypted", "true"),
				),
			},
			{
				ResourceName:      "aws_docdb_cluster.default",
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"allow_major_version_upgrade",
					"apply_immediately",
					"cluster_identifier_prefix",
					"final_snapshot_identifier",
					"master_password",
					"skip_final_snapshot",
				},
			},
		},
	})
}

func TestAccDocDBCluster_backupsUpdate(t *testing.T) {
	ctx := acctest.Context(t)
	var v docdb.DBCluster
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, docdb.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckClusterDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccClusterConfig_backups(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterExists(ctx, "aws_docdb_cluster.default", &v),
					resource.TestCheckResourceAttr(
						"aws_docdb_cluster.default", "preferred_backup_window", "07:00-09:00"),
					resource.TestCheckResourceAttr(
						"aws_docdb_cluster.default", "backup_retention_period", "5"),
					resource.TestCheckResourceAttr(
						"aws_docdb_cluster.default", "preferred_maintenance_window", "tue:04:00-tue:04:30"),
				),
			},
			{
				ResourceName:      "aws_docdb_cluster.default",
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"allow_major_version_upgrade",
					"apply_immediately",
					"cluster_identifier_prefix",
					"final_snapshot_identifier",
					"master_password",
					"skip_final_snapshot",
				},
			},
			{
				Config: testAccClusterConfig_backupsUpdate(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterExists(ctx, "aws_docdb_cluster.default", &v),
					resource.TestCheckResourceAttr(
						"aws_docdb_cluster.default", "preferred_backup_window", "03:00-09:00"),
					resource.TestCheckResourceAttr(
						"aws_docdb_cluster.default", "backup_retention_period", "10"),
					resource.TestCheckResourceAttr(
						"aws_docdb_cluster.default", "preferred_maintenance_window", "wed:01:00-wed:01:30"),
				),
			},
		},
	})
}

func TestAccDocDBCluster_port(t *testing.T) {
	ctx := acctest.Context(t)
	var dbCluster1, dbCluster2 docdb.DBCluster
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_docdb_cluster.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, docdb.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckClusterDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccClusterConfig_port(rName, 5432),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterExists(ctx, resourceName, &dbCluster1),
					resource.TestCheckResourceAttr(resourceName, "port", "5432"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"allow_major_version_upgrade",
					"apply_immediately",
					"cluster_identifier_prefix",
					"final_snapshot_identifier",
					"master_password",
					"skip_final_snapshot",
				},
			},
			{
				Config: testAccClusterConfig_port(rName, 2345),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterExists(ctx, resourceName, &dbCluster2),
					testAccCheckClusterRecreated(&dbCluster1, &dbCluster2),
					resource.TestCheckResourceAttr(resourceName, "port", "2345"),
				),
			},
		},
	})
}

func TestAccDocDBCluster_deleteProtection(t *testing.T) {
	ctx := acctest.Context(t)
	var dbCluster docdb.DBCluster
	resourceName := "aws_docdb_cluster.default"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, docdb.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckClusterDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccClusterConfig_deleteProtection(true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterExists(ctx, resourceName, &dbCluster),
					resource.TestCheckResourceAttr(resourceName, "deletion_protection", "true"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"allow_major_version_upgrade",
					"apply_immediately",
					"cluster_identifier_prefix",
					"final_snapshot_identifier",
					"master_password",
					"skip_final_snapshot",
				},
			},
			{
				Config: testAccClusterConfig_deleteProtection(false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterExists(ctx, resourceName, &dbCluster),
					resource.TestCheckResourceAttr(resourceName, "deletion_protection", "false"),
				),
			},
			{
				Config: testAccClusterConfig_deleteProtection(true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterExists(ctx, resourceName, &dbCluster),
					resource.TestCheckResourceAttr(resourceName, "deletion_protection", "true"),
				),
			},
			{
				Config: testAccClusterConfig_deleteProtection(false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterExists(ctx, resourceName, &dbCluster),
					resource.TestCheckResourceAttr(resourceName, "deletion_protection", "false"),
				),
			},
		},
	})
}

func testAccCheckClusterDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).DocDBConn(ctx)

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

func testAccCheckClusterExists(ctx context.Context, n string, v *docdb.DBCluster) resource.TestCheckFunc {
	return testAccCheckClusterExistsProvider(ctx, n, v, func() *schema.Provider { return acctest.Provider })
}

func testAccCheckClusterExistsProvider(ctx context.Context, n string, v *docdb.DBCluster, providerF func() *schema.Provider) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := providerF().Meta().(*conns.AWSClient).DocDBConn(ctx)

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
		conn := acctest.Provider.Meta().(*conns.AWSClient).DocDBConn(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_docdb_cluster" {
				continue
			}

			finalSnapshotID := rs.Primary.Attributes["final_snapshot_identifier"]
			_, err := conn.DeleteDBClusterSnapshotWithContext(ctx, &docdb.DeleteDBClusterSnapshotInput{
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

func testAccCheckClusterRecreated(i, j *docdb.DBCluster) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if aws.TimeValue(i.ClusterCreateTime).Equal(aws.TimeValue(j.ClusterCreateTime)) {
			return errors.New("DocumentDB Cluster was not recreated")
		}

		return nil
	}
}

func testAccClusterConfig_basic(rName string) string {
	return acctest.ConfigCompose(acctest.ConfigAvailableAZsNoOptIn(), fmt.Sprintf(`
resource "aws_docdb_cluster" "default" {
  cluster_identifier = %[1]q

  availability_zones = [
    data.aws_availability_zones.available.names[0],
    data.aws_availability_zones.available.names[1],
    data.aws_availability_zones.available.names[2]
  ]

  master_username     = "foo"
  master_password     = "mustbeeightcharaters"
  skip_final_snapshot = true

  tags = {
    Environment = "production"
  }

  enabled_cloudwatch_logs_exports = [
    "audit",
    "profiler",
  ]
}
`, rName))
}

func testAccClusterConfig_namePrefix() string {
	return `
resource "aws_docdb_cluster" "test" {
  cluster_identifier_prefix = "tf-test-"
  master_username           = "root"
  master_password           = "password"
  skip_final_snapshot       = true
}
`
}

func testAccClusterConfig_generatedName() string {
	return `
resource "aws_docdb_cluster" "test" {
  master_username     = "root"
  master_password     = "password"
  skip_final_snapshot = true
}
`
}

func testAccClusterConfig_finalSnapshot(rName, snapshotName string) string {
	return acctest.ConfigCompose(acctest.ConfigAvailableAZsNoOptIn(), fmt.Sprintf(`
resource "aws_docdb_cluster" "default" {
  cluster_identifier = %[1]q

  availability_zones = [
    data.aws_availability_zones.available.names[0],
    data.aws_availability_zones.available.names[1],
    data.aws_availability_zones.available.names[2]
  ]

  master_username           = "foo"
  master_password           = "mustbeeightcharaters"
  final_snapshot_identifier = %[2]q

  tags = {
    Environment = "production"
  }
}
`, rName, snapshotName))
}

func testAccClusterConfig_noUsernameOrPassword(rName string) string {
	return acctest.ConfigCompose(acctest.ConfigAvailableAZsNoOptIn(), fmt.Sprintf(`
resource "aws_docdb_cluster" "default" {
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

func testAccClusterConfig_updatedTags(rName string) string {
	return acctest.ConfigCompose(acctest.ConfigAvailableAZsNoOptIn(), fmt.Sprintf(`
resource "aws_docdb_cluster" "default" {
  cluster_identifier = %[1]q

  availability_zones = [
    data.aws_availability_zones.available.names[0],
    data.aws_availability_zones.available.names[1],
    data.aws_availability_zones.available.names[2]
  ]

  master_username     = "foo"
  master_password     = "mustbeeightcharaters"
  skip_final_snapshot = true

  tags = {
    Environment = "production"
    AnotherTag  = "test"
  }
}
`, rName))
}

func testAccClusterConfig_noCloudWatchLogs(rName string) string {
	return acctest.ConfigCompose(acctest.ConfigAvailableAZsNoOptIn(), fmt.Sprintf(`
resource "aws_docdb_cluster" "default" {
  cluster_identifier = %[1]q

  availability_zones = [
    data.aws_availability_zones.available.names[0],
    data.aws_availability_zones.available.names[1],
    data.aws_availability_zones.available.names[2]
  ]

  master_username     = "foo"
  master_password     = "mustbeeightcharaters"
  skip_final_snapshot = true

  tags = {
    Environment = "production"
  }
}
`, rName))
}

func testAccClusterConfig_kmsKey(rName string) string {
	return acctest.ConfigCompose(acctest.ConfigAvailableAZsNoOptIn(), fmt.Sprintf(`
resource "aws_kms_key" "foo" {
  description = "Terraform acc test %[1]s"

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

resource "aws_docdb_cluster" "default" {
  cluster_identifier = %[1]q
  availability_zones = [
    data.aws_availability_zones.available.names[0],
    data.aws_availability_zones.available.names[1],
    data.aws_availability_zones.available.names[2]
  ]

  master_username     = "foo"
  master_password     = "mustbeeightcharaters"
  storage_encrypted   = true
  kms_key_id          = aws_kms_key.foo.arn
  skip_final_snapshot = true
}
`, rName))
}

func testAccClusterConfig_encrypted(rName string) string {
	return acctest.ConfigCompose(acctest.ConfigAvailableAZsNoOptIn(), fmt.Sprintf(`
resource "aws_docdb_cluster" "default" {
  cluster_identifier = %[1]q

  availability_zones = [
    data.aws_availability_zones.available.names[0],
    data.aws_availability_zones.available.names[1],
    data.aws_availability_zones.available.names[2]
  ]

  master_username     = "foo"
  master_password     = "mustbeeightcharaters"
  storage_encrypted   = true
  skip_final_snapshot = true
}
`, rName))
}

func testAccClusterConfig_backups(rName string) string {
	return acctest.ConfigCompose(acctest.ConfigAvailableAZsNoOptIn(), fmt.Sprintf(`
resource "aws_docdb_cluster" "default" {
  cluster_identifier = %[1]q

  availability_zones = [
    data.aws_availability_zones.available.names[0],
    data.aws_availability_zones.available.names[1],
    data.aws_availability_zones.available.names[2]
  ]

  master_username              = "foo"
  master_password              = "mustbeeightcharaters"
  backup_retention_period      = 5
  preferred_backup_window      = "07:00-09:00"
  preferred_maintenance_window = "tue:04:00-tue:04:30"
  skip_final_snapshot          = true
}
`, rName))
}

func testAccClusterConfig_backupsUpdate(rName string) string {
	return acctest.ConfigCompose(acctest.ConfigAvailableAZsNoOptIn(), fmt.Sprintf(`
resource "aws_docdb_cluster" "default" {
  cluster_identifier = %[1]q

  availability_zones = [
    data.aws_availability_zones.available.names[0],
    data.aws_availability_zones.available.names[1],
    data.aws_availability_zones.available.names[2]
  ]

  master_username              = "foo"
  master_password              = "mustbeeightcharaters"
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
  master_password     = "mustbeeightcharaters"
  master_username     = "foo"
  port                = %[2]d
  skip_final_snapshot = true
}
`, rName, port))
}

func testAccClusterConfig_deleteProtection(isProtected bool) string {
	return fmt.Sprintf(`
resource "aws_docdb_cluster" "default" {
  cluster_identifier_prefix = "tf-test-"
  master_username           = "root"
  master_password           = "password"
  skip_final_snapshot       = true
  deletion_protection       = %[1]t
}
`, isProtected)
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
  global_cluster_identifier = "%[1]s"
  engine                    = "docdb"
  engine_version            = "4.0.0"
}

resource "aws_docdb_cluster" "primary" {
  cluster_identifier        = "%[2]s"
  master_username           = "foo"
  master_password           = "barbarbar"
  skip_final_snapshot       = true
  global_cluster_identifier = aws_docdb_global_cluster.test.id
  engine                    = aws_docdb_global_cluster.test.engine
  engine_version            = aws_docdb_global_cluster.test.engine_version
}

resource "aws_docdb_cluster_instance" "primary" {
  identifier         = "%[2]s"
  cluster_identifier = aws_docdb_cluster.primary.id
  instance_class     = "db.r5.large"
}

resource "aws_vpc" "alternate" {
  provider   = "awsalternate"
  cidr_block = "10.0.0.0/16"

  tags = {
    Name = "%[3]s"
  }
}

resource "aws_subnet" "alternate" {
  provider          = "awsalternate"
  count             = 3
  vpc_id            = aws_vpc.alternate.id
  availability_zone = data.aws_availability_zones.alternate.names[count.index]
  cidr_block        = "10.0.${count.index}.0/24"

  tags = {
    Name = "%[3]s"
  }
}

resource "aws_docdb_subnet_group" "alternate" {
  provider   = "awsalternate"
  name       = "%[3]s"
  subnet_ids = aws_subnet.alternate[*].id
}

resource "aws_docdb_cluster" "secondary" {
  provider                  = "awsalternate"
  cluster_identifier        = "%[3]s"
  skip_final_snapshot       = true
  db_subnet_group_name      = aws_docdb_subnet_group.alternate.name
  global_cluster_identifier = aws_docdb_global_cluster.test.id
  engine                    = aws_docdb_global_cluster.test.engine
  engine_version            = aws_docdb_global_cluster.test.engine_version
  depends_on                = [aws_docdb_cluster_instance.primary]
}

resource "aws_docdb_cluster_instance" "secondary" {
  provider           = "awsalternate"
  identifier         = "%[3]s"
  cluster_identifier = aws_docdb_cluster.secondary.id
  instance_class     = "db.r5.large"
}
`, rNameGlobal, rNamePrimary, rNameSecondary))
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
  master_password           = "barbarbarbar"
  master_username           = "foo"
  skip_final_snapshot       = true
}
`, rName, globalClusterIdentifierResourceName)
}

func testAccClusterConfig_globalCompatible(rName string) string {
	return fmt.Sprintf(`
resource "aws_docdb_cluster" "test" {
  cluster_identifier  = %[1]q
  engine_version      = "4.0.0" # version compatible with global
  master_password     = "barbarbarbar"
  master_username     = "foo"
  skip_final_snapshot = true
}
`, rName)
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
  master_password           = "barbarbarbar"
  master_username           = "foo"
  skip_final_snapshot       = true
}
`, rName)
}
