// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package neptune_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/neptune"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfneptune "github.com/hashicorp/terraform-provider-aws/internal/service/neptune"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func testAccClusterImportStep(n string) resource.TestStep {
	return resource.TestStep{
		ResourceName:      n,
		ImportState:       true,
		ImportStateVerify: true,
		ImportStateVerifyIgnore: []string{
			names.AttrAllowMajorVersionUpgrade,
			names.AttrApplyImmediately,
			names.AttrFinalSnapshotIdentifier,
			"neptune_instance_parameter_group_name",
			"skip_final_snapshot",
			"snapshot_identifier",
		},
	}
}

func TestAccNeptuneCluster_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var dbCluster neptune.DBCluster
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_neptune_cluster.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.NeptuneServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckClusterDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccClusterConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckClusterExists(ctx, resourceName, &dbCluster),
					resource.TestCheckNoResourceAttr(resourceName, names.AttrAllowMajorVersionUpgrade),
					resource.TestCheckNoResourceAttr(resourceName, names.AttrApplyImmediately),
					acctest.MatchResourceAttrRegionalARN(resourceName, names.AttrARN, "rds", regexache.MustCompile(`cluster:.+`)),
					acctest.CheckResourceAttrGreaterThanValue(resourceName, "availability_zones.#", 0),
					resource.TestCheckResourceAttr(resourceName, "backup_retention_period", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, names.AttrClusterIdentifier, rName),
					resource.TestCheckResourceAttr(resourceName, "cluster_identifier_prefix", ""),
					resource.TestCheckResourceAttr(resourceName, "cluster_members.#", acctest.Ct0),
					resource.TestCheckResourceAttrSet(resourceName, "cluster_resource_id"),
					resource.TestCheckResourceAttr(resourceName, "copy_tags_to_snapshot", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, names.AttrDeletionProtection, acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "enable_cloudwatch_logs_exports.#", acctest.Ct0),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrEndpoint),
					resource.TestCheckResourceAttr(resourceName, names.AttrEngine, "neptune"),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrEngineVersion),
					resource.TestCheckNoResourceAttr(resourceName, names.AttrFinalSnapshotIdentifier),
					resource.TestCheckResourceAttr(resourceName, "global_cluster_identifier", ""),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrHostedZoneID),
					resource.TestCheckResourceAttr(resourceName, "iam_database_authentication_enabled", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "iam_roles.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, names.AttrKMSKeyARN, ""),
					resource.TestCheckResourceAttr(resourceName, "neptune_cluster_parameter_group_name", "default.neptune1.3"),
					resource.TestCheckNoResourceAttr(resourceName, "neptune_instance_parameter_group_name"),
					resource.TestCheckResourceAttr(resourceName, "neptune_subnet_group_name", "default"),
					resource.TestCheckResourceAttr(resourceName, names.AttrPort, "8182"),
					resource.TestCheckResourceAttrSet(resourceName, "preferred_backup_window"),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrPreferredMaintenanceWindow),
					resource.TestCheckResourceAttrSet(resourceName, "reader_endpoint"),
					resource.TestCheckResourceAttr(resourceName, "replication_source_identifier", ""),
					resource.TestCheckResourceAttr(resourceName, "serverless_v2_scaling_configuration.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "skip_final_snapshot", acctest.CtTrue),
					resource.TestCheckNoResourceAttr(resourceName, "snapshot_identifier"),
					resource.TestCheckResourceAttr(resourceName, names.AttrStorageEncrypted, acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, names.AttrStorageType, ""),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "vpc_security_group_ids.#", acctest.Ct1),
				),
			},
			testAccClusterImportStep(resourceName),
		},
	})
}

func TestAccNeptuneCluster_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	var dbCluster neptune.DBCluster
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_neptune_cluster.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.NeptuneServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckClusterDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccClusterConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterExists(ctx, resourceName, &dbCluster),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfneptune.ResourceCluster(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccNeptuneCluster_identifierGenerated(t *testing.T) {
	ctx := acctest.Context(t)
	var v neptune.DBCluster
	resourceName := "aws_neptune_cluster.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.NeptuneServiceID),
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
			testAccClusterImportStep(resourceName),
		},
	})
}

func TestAccNeptuneCluster_identifierPrefix(t *testing.T) {
	ctx := acctest.Context(t)
	var v neptune.DBCluster
	resourceName := "aws_neptune_cluster.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.NeptuneServiceID),
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
			testAccClusterImportStep(resourceName),
		},
	})
}

func TestAccNeptuneCluster_tags(t *testing.T) {
	ctx := acctest.Context(t)
	var v neptune.DBCluster
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_neptune_cluster.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.NeptuneServiceID),
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
			testAccClusterImportStep(resourceName),
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

func TestAccNeptuneCluster_copyTagsToSnapshot(t *testing.T) {
	ctx := acctest.Context(t)
	var dbCluster neptune.DBCluster
	rName := sdkacctest.RandomWithPrefix("tf-acc")
	resourceName := "aws_neptune_cluster.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.NeptuneServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckClusterDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccClusterConfig_copyTags(rName, true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterExists(ctx, resourceName, &dbCluster),
					resource.TestCheckResourceAttr(resourceName, "copy_tags_to_snapshot", acctest.CtTrue),
				),
			},
			testAccClusterImportStep(resourceName),
			{
				Config: testAccClusterConfig_copyTags(rName, false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterExists(ctx, resourceName, &dbCluster),
					resource.TestCheckResourceAttr(resourceName, "copy_tags_to_snapshot", acctest.CtFalse),
				),
			},
			{
				Config: testAccClusterConfig_copyTags(rName, true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterExists(ctx, resourceName, &dbCluster),
					resource.TestCheckResourceAttr(resourceName, "copy_tags_to_snapshot", acctest.CtTrue),
				),
			},
		},
	})
}

func TestAccNeptuneCluster_serverlessConfiguration(t *testing.T) {
	ctx := acctest.Context(t)
	var v neptune.DBCluster
	rName := sdkacctest.RandomWithPrefix("tf-acc")
	resourceName := "aws_neptune_cluster.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.NeptuneServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckClusterDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccClusterConfig_serverlessConfiguration(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "serverless_v2_scaling_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "serverless_v2_scaling_configuration.0.min_capacity", "4.5"),
					resource.TestCheckResourceAttr(resourceName, "serverless_v2_scaling_configuration.0.max_capacity", "12.5"),
				),
			},
			testAccClusterImportStep(resourceName),
		},
	})
}

func TestAccNeptuneCluster_takeFinalSnapshot(t *testing.T) {
	ctx := acctest.Context(t)
	var v neptune.DBCluster
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_neptune_cluster.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.NeptuneServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckClusterDestroyWithFinalSnapshot(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccClusterConfig_finalSnapshot(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterExists(ctx, resourceName, &v),
				),
			},
			testAccClusterImportStep(resourceName),
		},
	})
}

func TestAccNeptuneCluster_updateIAMRoles(t *testing.T) {
	ctx := acctest.Context(t)
	var v neptune.DBCluster
	rName := sdkacctest.RandomWithPrefix("tf-acc")
	resourceName := "aws_neptune_cluster.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.NeptuneServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckClusterDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccClusterConfig_includingIAMRoles(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterExists(ctx, resourceName, &v),
				),
			},
			{
				Config: testAccClusterConfig_addIAMRoles(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "iam_roles.#", acctest.Ct2),
				),
			},
			{
				Config: testAccClusterConfig_removeIAMRoles(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "iam_roles.#", acctest.Ct1),
				),
			},
			testAccClusterImportStep(resourceName),
		},
	})
}

func TestAccNeptuneCluster_kmsKey(t *testing.T) {
	ctx := acctest.Context(t)
	var v neptune.DBCluster
	resourceName := "aws_neptune_cluster.test"
	keyResourceName := "aws_kms_key.test"
	rName := sdkacctest.RandomWithPrefix("tf-acc")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.NeptuneServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckClusterDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccClusterConfig_kmsKey(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrKMSKeyARN, keyResourceName, names.AttrARN),
				),
			},
			testAccClusterImportStep(resourceName),
		},
	})
}

func TestAccNeptuneCluster_encrypted(t *testing.T) {
	ctx := acctest.Context(t)
	var v neptune.DBCluster
	resourceName := "aws_neptune_cluster.test"
	rName := sdkacctest.RandomWithPrefix("tf-acc")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.NeptuneServiceID),
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
			testAccClusterImportStep(resourceName),
		},
	})
}

func TestAccNeptuneCluster_backupsUpdate(t *testing.T) {
	ctx := acctest.Context(t)
	var v neptune.DBCluster
	resourceName := "aws_neptune_cluster.test"
	rName := sdkacctest.RandomWithPrefix("tf-acc")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.NeptuneServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckClusterDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccClusterConfig_backups(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "preferred_backup_window", "07:00-09:00"),
					resource.TestCheckResourceAttr(resourceName, "backup_retention_period", "5"),
					resource.TestCheckResourceAttr(resourceName, names.AttrPreferredMaintenanceWindow, "tue:04:00-tue:04:30"),
				),
			},
			{
				Config: testAccClusterConfig_backupsUpdate(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "preferred_backup_window", "03:00-09:00"),
					resource.TestCheckResourceAttr(resourceName, "backup_retention_period", acctest.Ct10),
					resource.TestCheckResourceAttr(resourceName, names.AttrPreferredMaintenanceWindow, "wed:01:00-wed:01:30"),
				),
			},
			testAccClusterImportStep(resourceName),
		},
	})
}

func TestAccNeptuneCluster_iamAuth(t *testing.T) {
	ctx := acctest.Context(t)
	var v neptune.DBCluster
	resourceName := "aws_neptune_cluster.test"
	rName := sdkacctest.RandomWithPrefix("tf-acc")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.NeptuneServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckClusterDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccClusterConfig_iamAuth(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "iam_database_authentication_enabled", acctest.CtTrue),
				),
			},
			testAccClusterImportStep(resourceName),
		},
	})
}

func TestAccNeptuneCluster_updateCloudWatchLogsExports(t *testing.T) {
	ctx := acctest.Context(t)
	var dbCluster neptune.DBCluster
	rName := sdkacctest.RandomWithPrefix("tf-acc")
	resourceName := "aws_neptune_cluster.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.NeptuneServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckClusterDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccClusterConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterExists(ctx, resourceName, &dbCluster),
					resource.TestCheckNoResourceAttr(resourceName, "enable_cloudwatch_logs_exports.#"),
				),
			},
			{
				Config: testAccClusterConfig_cloudWatchLogsExports(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterExists(ctx, resourceName, &dbCluster),
					resource.TestCheckResourceAttr(resourceName, "enable_cloudwatch_logs_exports.#", acctest.Ct2),
					resource.TestCheckTypeSetElemAttr(resourceName, "enable_cloudwatch_logs_exports.*", "audit"),
					resource.TestCheckTypeSetElemAttr(resourceName, "enable_cloudwatch_logs_exports.*", "slowquery"),
				),
			},
			{
				Config: testAccClusterConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterExists(ctx, resourceName, &dbCluster),
					resource.TestCheckResourceAttr(resourceName, "enable_cloudwatch_logs_exports.#", acctest.Ct0),
				),
			},
			testAccClusterImportStep(resourceName),
		},
	})
}

func TestAccNeptuneCluster_updateEngineVersion(t *testing.T) {
	ctx := acctest.Context(t)
	var dbCluster neptune.DBCluster
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_neptune_cluster.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.NeptuneServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckClusterDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccClusterConfig_engineVersion(rName, "1.1.0.0", "default.neptune1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterExists(ctx, resourceName, &dbCluster),
					resource.TestCheckResourceAttr(resourceName, names.AttrEngineVersion, "1.1.0.0"),
				),
			},
			testAccClusterImportStep(resourceName),
			{
				Config: testAccClusterConfig_engineVersion(rName, "1.1.1.0", "default.neptune1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterExists(ctx, resourceName, &dbCluster),
					resource.TestCheckResourceAttr(resourceName, names.AttrEngineVersion, "1.1.1.0"),
				),
			},
			testAccClusterImportStep(resourceName),
		},
	})
}

func TestAccNeptuneCluster_updateEngineMajorVersion(t *testing.T) {
	ctx := acctest.Context(t)
	var dbCluster neptune.DBCluster
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_neptune_cluster.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.NeptuneServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckClusterDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccClusterConfig_engineVersion(rName, "1.1.1.0", "default.neptune1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterExists(ctx, resourceName, &dbCluster),
					resource.TestCheckResourceAttr(resourceName, names.AttrEngineVersion, "1.1.1.0"),
				),
			},
			testAccClusterImportStep(resourceName),
			{
				Config: testAccClusterConfig_engineMajorVersionUpdate(rName, "1.2.0.1", "default.neptune1.2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterExists(ctx, resourceName, &dbCluster),
					resource.TestCheckResourceAttr(resourceName, names.AttrEngineVersion, "1.2.0.1"),
				),
			},
			testAccClusterImportStep(resourceName),
		},
	})
}

func TestAccNeptuneCluster_GlobalClusterIdentifier_PrimarySecondaryClusters(t *testing.T) {
	ctx := acctest.Context(t)
	var providers []*schema.Provider
	var primaryDbCluster, secondaryDbCluster neptune.DBCluster

	rNameGlobal := sdkacctest.RandomWithPrefix("tf-acc-test-global")
	rNamePrimary := sdkacctest.RandomWithPrefix("tf-acc-test-primary")
	rNameSecondary := sdkacctest.RandomWithPrefix("tf-acc-test-secondary")

	resourceNamePrimary := "aws_neptune_cluster.primary"
	resourceNameSecondary := "aws_neptune_cluster.secondary"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckMultipleRegion(t, 2)
			testAccPreCheckGlobalCluster(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.NeptuneServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5FactoriesPlusProvidersAlternate(ctx, t, &providers),
		CheckDestroy:             testAccCheckClusterDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccClusterConfig_globalIdentifierPrimarySecondary(rNameGlobal, rNamePrimary, rNameSecondary),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterExistsWithProvider(ctx, resourceNamePrimary, &primaryDbCluster, acctest.RegionProviderFunc(acctest.Region(), &providers)),
					testAccCheckClusterExistsWithProvider(ctx, resourceNameSecondary, &secondaryDbCluster, acctest.RegionProviderFunc(acctest.AlternateRegion(), &providers)),
				),
			},
		},
	})
}

func TestAccNeptuneCluster_deleteProtection(t *testing.T) {
	ctx := acctest.Context(t)
	var dbCluster neptune.DBCluster
	rName := sdkacctest.RandomWithPrefix("tf-acc")
	resourceName := "aws_neptune_cluster.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.NeptuneServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckClusterDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccClusterConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterExists(ctx, resourceName, &dbCluster),
					resource.TestCheckResourceAttr(resourceName, names.AttrDeletionProtection, acctest.CtFalse),
				),
			},
			testAccClusterImportStep(resourceName),
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

func TestAccNeptuneCluster_restoreFromSnapshot(t *testing.T) {
	ctx := acctest.Context(t)
	var dbCluster neptune.DBCluster
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_neptune_cluster.test"
	keyResourceName := "aws_kms_key.test2"
	parameterGroupResourceName := "aws_neptune_cluster_parameter_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.NeptuneServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckClusterDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccClusterConfig_restoreFromSnapshot(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckClusterExists(ctx, resourceName, &dbCluster),
					resource.TestCheckResourceAttr(resourceName, "backup_retention_period", "5"),
					resource.TestCheckResourceAttr(resourceName, names.AttrClusterIdentifier, rName),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrKMSKeyARN, keyResourceName, names.AttrARN),
					resource.TestCheckResourceAttrPair(resourceName, "neptune_cluster_parameter_group_name", parameterGroupResourceName, names.AttrID),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "tags.Name", rName),
					resource.TestCheckResourceAttr(resourceName, "vpc_security_group_ids.#", acctest.Ct2),
				),
			},
			testAccClusterImportStep(resourceName),
		},
	})
}

func TestAccNeptuneCluster_storageType(t *testing.T) {
	ctx := acctest.Context(t)
	var v neptune.DBCluster
	resourceName := "aws_neptune_cluster.test"
	rName := sdkacctest.RandomWithPrefix("tf-acc")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.NeptuneServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckClusterDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccClusterConfig_storageType(rName, "standard"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, names.AttrStorageType, ""),
				),
			},
			testAccClusterImportStep(resourceName),
			{
				Config: testAccClusterConfig_storageType(rName, "iopt1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, names.AttrStorageType, "iopt1"),
				),
			},
		},
	})
}

func testAccCheckClusterDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).NeptuneConn(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_neptune_cluster" {
				continue
			}

			_, err := tfneptune.FindDBClusterByID(ctx, conn, rs.Primary.ID)

			if tfresource.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("Neptune Cluster %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccCheckClusterExists(ctx context.Context, n string, v *neptune.DBCluster) resource.TestCheckFunc {
	return testAccCheckClusterExistsWithProvider(ctx, n, v, func() *schema.Provider { return acctest.Provider })
}

func testAccCheckClusterExistsWithProvider(ctx context.Context, n string, v *neptune.DBCluster, providerF func() *schema.Provider) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No Neptune Cluster ID is set")
		}

		conn := providerF().Meta().(*conns.AWSClient).NeptuneConn(ctx)

		output, err := tfneptune.FindDBClusterByID(ctx, conn, rs.Primary.ID)

		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

func testAccCheckClusterDestroyWithFinalSnapshot(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_neptune_cluster" {
				continue
			}

			conn := acctest.Provider.Meta().(*conns.AWSClient).NeptuneConn(ctx)

			finalSnapshotID := rs.Primary.Attributes[names.AttrFinalSnapshotIdentifier]
			_, err := tfneptune.FindClusterSnapshotByID(ctx, conn, finalSnapshotID)

			if err != nil {
				return err
			}

			_, err = conn.DeleteDBClusterSnapshotWithContext(ctx, &neptune.DeleteDBClusterSnapshotInput{
				DBClusterSnapshotIdentifier: aws.String(finalSnapshotID),
			})

			if err != nil {
				return err
			}

			_, err = tfneptune.FindDBClusterByID(ctx, conn, rs.Primary.ID)

			if tfresource.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("Neptune Cluster %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccClusterConfig_base() string {
	return acctest.ConfigCompose(acctest.ConfigAvailableAZsNoOptIn(), `
locals {
  availability_zone_names = slice(data.aws_availability_zones.available.names, 0, min(3, length(data.aws_availability_zones.available.names)))
}
`)
}

func testAccClusterConfig_basic(rName string) string {
	return acctest.ConfigCompose(testAccClusterConfig_base(), fmt.Sprintf(`
resource "aws_neptune_cluster" "test" {
  cluster_identifier                   = %[1]q
  availability_zones                   = local.availability_zone_names
  engine                               = "neptune"
  neptune_cluster_parameter_group_name = "default.neptune1.3"
  skip_final_snapshot                  = true
}
`, rName))
}

func testAccClusterConfig_identifierGenerated() string {
	return `
resource "aws_neptune_cluster" "test" {
  engine                               = "neptune"
  neptune_cluster_parameter_group_name = "default.neptune1.3"
  skip_final_snapshot                  = true
}
`
}

func testAccClusterConfig_identifierPrefix(prefix string) string {
	return fmt.Sprintf(`
resource "aws_neptune_cluster" "test" {
  cluster_identifier_prefix            = %[1]q
  engine                               = "neptune"
  neptune_cluster_parameter_group_name = "default.neptune1.3"
  skip_final_snapshot                  = true
}
`, prefix)
}

func testAccClusterConfig_tags1(rName, tagKey1, tagValue1 string) string {
	return acctest.ConfigCompose(testAccClusterConfig_base(), fmt.Sprintf(`
resource "aws_neptune_cluster" "test" {
  cluster_identifier                   = %[1]q
  availability_zones                   = local.availability_zone_names
  engine                               = "neptune"
  neptune_cluster_parameter_group_name = "default.neptune1.3"
  skip_final_snapshot                  = true

  tags = {
    %[2]q = %[3]q
  }
}
`, rName, tagKey1, tagValue1))
}

func testAccClusterConfig_tags2(rName, tagKey1, tagValue1, tagKey2, tagValue2 string) string {
	return acctest.ConfigCompose(testAccClusterConfig_base(), fmt.Sprintf(`
resource "aws_neptune_cluster" "test" {
  cluster_identifier                   = %[1]q
  availability_zones                   = local.availability_zone_names
  engine                               = "neptune"
  neptune_cluster_parameter_group_name = "default.neptune1.3"
  skip_final_snapshot                  = true

  tags = {
    %[2]q = %[3]q
    %[4]q = %[5]q
  }
}
`, rName, tagKey1, tagValue1, tagKey2, tagValue2))
}

func testAccClusterConfig_copyTags(rName string, copy bool) string {
	return acctest.ConfigCompose(testAccClusterConfig_base(), fmt.Sprintf(`
resource "aws_neptune_cluster" "test" {
  cluster_identifier                   = %[1]q
  availability_zones                   = local.availability_zone_names
  engine                               = "neptune"
  neptune_cluster_parameter_group_name = "default.neptune1.3"
  skip_final_snapshot                  = true
  copy_tags_to_snapshot                = %[2]t
}
`, rName, copy))
}

func testAccClusterConfig_deleteProtection(rName string, isProtected bool) string {
	return acctest.ConfigCompose(testAccClusterConfig_base(), fmt.Sprintf(`
resource "aws_neptune_cluster" "test" {
  cluster_identifier                   = %[1]q
  availability_zones                   = local.availability_zone_names
  engine                               = "neptune"
  neptune_cluster_parameter_group_name = "default.neptune1.3"
  skip_final_snapshot                  = true
  deletion_protection                  = %[2]t
}
`, rName, isProtected))
}

func testAccClusterConfig_serverlessConfiguration(rName string) string {
	return fmt.Sprintf(`
resource "aws_neptune_cluster" "test" {
  cluster_identifier_prefix            = %[1]q
  engine                               = "neptune"
  engine_version                       = "1.2.0.1"
  neptune_cluster_parameter_group_name = "default.neptune1.2"
  skip_final_snapshot                  = true

  serverless_v2_scaling_configuration {
    min_capacity = 4.5
    max_capacity = 12.5
  }
}
`, rName)
}

func testAccClusterConfig_finalSnapshot(rName string) string {
	return acctest.ConfigCompose(testAccClusterConfig_base(), fmt.Sprintf(`
resource "aws_neptune_cluster" "test" {
  cluster_identifier                   = %[1]q
  availability_zones                   = local.availability_zone_names
  neptune_cluster_parameter_group_name = "default.neptune1.3"
  final_snapshot_identifier            = %[1]q
}
`, rName))
}

func testAccClusterConfig_includingIAMRoles(rName string) string {
	return acctest.ConfigCompose(testAccClusterConfig_base(), fmt.Sprintf(`
resource "aws_iam_role" "test" {
  name = %[1]q
  path = "/"

  assume_role_policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Action": "sts:AssumeRole",
      "Principal": {
        "Service": "rds.amazonaws.com"
      },
      "Effect": "Allow",
      "Sid": ""
    }
  ]
}
EOF
}

resource "aws_iam_role_policy" "test" {
  name = %[1]q
  role = aws_iam_role.test.name

  policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": {
    "Effect": "Allow",
    "Action": "*",
    "Resource": "*"
  }
}
EOF
}

resource "aws_iam_role" "test-2" {
  name = "%[1]s-2"
  path = "/"

  assume_role_policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Action": "sts:AssumeRole",
      "Principal": {
        "Service": "rds.amazonaws.com"
      },
      "Effect": "Allow",
      "Sid": ""
    }
  ]
}
EOF
}

resource "aws_iam_role_policy" "test-2" {
  name = "%[1]s-2"
  role = aws_iam_role.test-2.name

  policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": {
    "Effect": "Allow",
    "Action": "*",
    "Resource": "*"
  }
}
EOF
}

resource "aws_neptune_cluster" "test" {
  cluster_identifier                   = %[1]q
  availability_zones                   = local.availability_zone_names
  neptune_cluster_parameter_group_name = "default.neptune1.3"
  skip_final_snapshot                  = true

  depends_on = [aws_iam_role.test, aws_iam_role.test-2]
}
`, rName))
}

func testAccClusterConfig_addIAMRoles(rName string) string {
	return acctest.ConfigCompose(testAccClusterConfig_base(), fmt.Sprintf(`
resource "aws_iam_role" "test" {
  name = %[1]q
  path = "/"

  assume_role_policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Action": "sts:AssumeRole",
      "Principal": {
        "Service": "rds.amazonaws.com"
      },
      "Effect": "Allow",
      "Sid": ""
    }
  ]
}
EOF
}

resource "aws_iam_role_policy" "test" {
  name = %[1]q
  role = aws_iam_role.test.name

  policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": {
    "Effect": "Allow",
    "Action": "*",
    "Resource": "*"
  }
}
EOF
}

resource "aws_iam_role" "test-2" {
  name = "%[1]s-2"
  path = "/"

  assume_role_policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Action": "sts:AssumeRole",
      "Principal": {
        "Service": "rds.amazonaws.com"
      },
      "Effect": "Allow",
      "Sid": ""
    }
  ]
}
EOF
}

resource "aws_iam_role_policy" "test-2" {
  name = "%[1]s-2"
  role = aws_iam_role.test-2.name

  policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": {
    "Effect": "Allow",
    "Action": "*",
    "Resource": "*"
  }
}
EOF
}

resource "aws_neptune_cluster" "test" {
  cluster_identifier  = %[1]q
  availability_zones  = local.availability_zone_names
  skip_final_snapshot = true
  iam_roles           = [aws_iam_role.test.arn, aws_iam_role.test-2.arn]

  neptune_cluster_parameter_group_name = "default.neptune1.3"

  depends_on = [aws_iam_role.test, aws_iam_role.test-2]
}
`, rName))
}

func testAccClusterConfig_removeIAMRoles(rName string) string {
	return acctest.ConfigCompose(testAccClusterConfig_base(), fmt.Sprintf(`
resource "aws_iam_role" "test" {
  name = %[1]q
  path = "/"

  assume_role_policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Action": "sts:AssumeRole",
      "Principal": {
        "Service": "rds.amazonaws.com"
      },
      "Effect": "Allow",
      "Sid": ""
    }
  ]
}
EOF
}

resource "aws_iam_role_policy" "test" {
  name = %[1]q
  role = aws_iam_role.test.name

  policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": {
    "Effect": "Allow",
    "Action": "*",
    "Resource": "*"
  }
}
EOF
}

resource "aws_neptune_cluster" "test" {
  cluster_identifier  = %[1]q
  availability_zones  = local.availability_zone_names
  skip_final_snapshot = true
  iam_roles           = [aws_iam_role.test.arn]

  neptune_cluster_parameter_group_name = "default.neptune1.3"

  depends_on = [aws_iam_role.test]
}
`, rName))
}

func testAccClusterConfig_kmsKey(rName string) string {
	return acctest.ConfigCompose(testAccClusterConfig_base(), fmt.Sprintf(`
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

resource "aws_neptune_cluster" "test" {
  cluster_identifier                   = %[1]q
  availability_zones                   = local.availability_zone_names
  neptune_cluster_parameter_group_name = "default.neptune1.3"
  storage_encrypted                    = true
  kms_key_arn                          = aws_kms_key.test.arn
  skip_final_snapshot                  = true
}
`, rName))
}

func testAccClusterConfig_encrypted(rName string) string {
	return acctest.ConfigCompose(testAccClusterConfig_base(), fmt.Sprintf(`
resource "aws_neptune_cluster" "test" {
  cluster_identifier  = %[1]q
  availability_zones  = local.availability_zone_names
  storage_encrypted   = true
  skip_final_snapshot = true

  neptune_cluster_parameter_group_name = "default.neptune1.3"
}
`, rName))
}

func testAccClusterConfig_backups(rName string) string {
	return acctest.ConfigCompose(testAccClusterConfig_base(), fmt.Sprintf(`
resource "aws_neptune_cluster" "test" {
  cluster_identifier           = %[1]q
  availability_zones           = local.availability_zone_names
  backup_retention_period      = 5
  preferred_backup_window      = "07:00-09:00"
  preferred_maintenance_window = "tue:04:00-tue:04:30"
  skip_final_snapshot          = true

  neptune_cluster_parameter_group_name = "default.neptune1.3"
}
`, rName))
}

func testAccClusterConfig_backupsUpdate(rName string) string {
	return acctest.ConfigCompose(testAccClusterConfig_base(), fmt.Sprintf(`
resource "aws_neptune_cluster" "test" {
  cluster_identifier           = %[1]q
  availability_zones           = local.availability_zone_names
  backup_retention_period      = 10
  preferred_backup_window      = "03:00-09:00"
  preferred_maintenance_window = "wed:01:00-wed:01:30"
  apply_immediately            = true
  skip_final_snapshot          = true

  neptune_cluster_parameter_group_name = "default.neptune1.3"
}
`, rName))
}

func testAccClusterConfig_iamAuth(rName string) string {
	return acctest.ConfigCompose(testAccClusterConfig_base(), fmt.Sprintf(`
resource "aws_neptune_cluster" "test" {
  cluster_identifier                  = %[1]q
  availability_zones                  = local.availability_zone_names
  iam_database_authentication_enabled = true
  skip_final_snapshot                 = true

  neptune_cluster_parameter_group_name = "default.neptune1.3"
}
`, rName))
}

func testAccClusterConfig_cloudWatchLogsExports(rName string) string {
	return acctest.ConfigCompose(testAccClusterConfig_base(), fmt.Sprintf(`
resource "aws_neptune_cluster" "test" {
  cluster_identifier             = %[1]q
  availability_zones             = local.availability_zone_names
  skip_final_snapshot            = true
  enable_cloudwatch_logs_exports = ["audit", "slowquery"]

  neptune_cluster_parameter_group_name = "default.neptune1.3"
}
`, rName))
}

func testAccClusterConfig_engineVersionBase(rName string) string {
	return acctest.ConfigCompose(testAccClusterConfig_base(), fmt.Sprintf(`
data "aws_neptune_orderable_db_instance" "test" {
  engine         = "neptune"
  engine_version = aws_neptune_cluster.test.engine_version
  license_model  = "amazon-license"

  preferred_instance_classes = ["db.t3.medium", "db.r5.large", "db.r4.large"]
}

resource "aws_neptune_cluster_instance" "test" {
  identifier                   = %[1]q
  cluster_identifier           = aws_neptune_cluster.test.id
  apply_immediately            = true
  instance_class               = data.aws_neptune_orderable_db_instance.test.instance_class
  neptune_parameter_group_name = aws_neptune_cluster.test.neptune_cluster_parameter_group_name
  promotion_tier               = "3"
}
`, rName))
}

func testAccClusterConfig_engineVersion(rName, engineVersion, clusterParameterGroupName string) string {
	return acctest.ConfigCompose(testAccClusterConfig_engineVersionBase(rName), fmt.Sprintf(`
resource "aws_neptune_cluster" "test" {
  cluster_identifier                   = %[1]q
  apply_immediately                    = true
  availability_zones                   = local.availability_zone_names
  engine_version                       = %[2]q
  neptune_cluster_parameter_group_name = %[3]q
  skip_final_snapshot                  = true
}
`, rName, engineVersion, clusterParameterGroupName))
}

func testAccClusterConfig_engineMajorVersionUpdate(rName, engineVersion, clusterParameterGroupName string) string {
	return acctest.ConfigCompose(testAccClusterConfig_engineVersionBase(rName), fmt.Sprintf(`
resource "aws_neptune_cluster" "test" {
  cluster_identifier                   = %[1]q
  apply_immediately                    = true
  availability_zones                   = local.availability_zone_names
  engine_version                       = %[2]q
  neptune_cluster_parameter_group_name = %[3]q
  skip_final_snapshot                  = true
  allow_major_version_upgrade          = true
}
`, rName, engineVersion, clusterParameterGroupName))
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

resource "aws_neptune_global_cluster" "test" {
  global_cluster_identifier = %[1]q
  engine                    = "neptune"
  engine_version            = "1.2.0.0"
}

resource "aws_neptune_cluster" "primary" {
  cluster_identifier                   = %[2]q
  skip_final_snapshot                  = true
  global_cluster_identifier            = aws_neptune_global_cluster.test.id
  engine                               = aws_neptune_global_cluster.test.engine
  engine_version                       = aws_neptune_global_cluster.test.engine_version
  neptune_cluster_parameter_group_name = "default.neptune1.2"
}

resource "aws_neptune_cluster_instance" "primary" {
  identifier                   = %[2]q
  cluster_identifier           = aws_neptune_cluster.primary.id
  instance_class               = "db.r5.large"
  neptune_parameter_group_name = "default.neptune1.2"
  engine_version               = aws_neptune_global_cluster.test.engine_version
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

resource "aws_neptune_subnet_group" "alternate" {
  provider   = "awsalternate"
  name       = %[3]q
  subnet_ids = aws_subnet.alternate[*].id
}


resource "aws_neptune_cluster" "secondary" {
  provider                             = "awsalternate"
  cluster_identifier                   = %[3]q
  skip_final_snapshot                  = true
  neptune_subnet_group_name            = aws_neptune_subnet_group.alternate.name
  global_cluster_identifier            = aws_neptune_global_cluster.test.id
  engine                               = aws_neptune_global_cluster.test.engine
  engine_version                       = aws_neptune_global_cluster.test.engine_version
  neptune_cluster_parameter_group_name = "default.neptune1.2"

  depends_on = [aws_neptune_cluster_instance.primary]

  lifecycle {
    ignore_changes = [replication_source_identifier]
  }
}

resource "aws_neptune_cluster_instance" "secondary" {
  provider                     = "awsalternate"
  identifier                   = %[3]q
  cluster_identifier           = aws_neptune_cluster.secondary.id
  neptune_parameter_group_name = "default.neptune1.2"
  engine_version               = aws_neptune_global_cluster.test.engine_version
  instance_class               = "db.r5.large"
}
`, rNameGlobal, rNamePrimary, rNameSecondary))
}

func testAccClusterConfig_restoreFromSnapshot(rName string) string {
	return fmt.Sprintf(`
resource "aws_kms_key" "test1" {
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

resource "aws_kms_key" "test2" {
  description = %[1]q

  policy = <<POLICY
{
  "Version": "2012-10-17",
  "Id": "kms-tf-2",
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

resource "aws_default_vpc" "test" {}

resource "aws_security_group" "test" {
  count = 2

  name   = "%[1]s-${count.index}"
  vpc_id = aws_default_vpc.test.id

  tags = {
    Name = %[1]q
  }
}

resource "aws_neptune_cluster" "source" {
  cluster_identifier                   = "%[1]s-src"
  neptune_cluster_parameter_group_name = "default.neptune1.3"
  skip_final_snapshot                  = true
  storage_encrypted                    = true
  kms_key_arn                          = aws_kms_key.test1.arn
}

resource "aws_neptune_cluster_snapshot" "test" {
  db_cluster_identifier          = aws_neptune_cluster.source.id
  db_cluster_snapshot_identifier = %[1]q
}

resource "aws_neptune_cluster_parameter_group" "test" {
  family = "neptune1.2"
  name   = %[1]q

  parameter {
    name  = "neptune_enable_audit_log"
    value = "1"
  }
}

resource "aws_neptune_cluster" "test" {
  cluster_identifier                   = %[1]q
  skip_final_snapshot                  = true
  storage_encrypted                    = true
  snapshot_identifier                  = aws_neptune_cluster_snapshot.test.id
  kms_key_arn                          = aws_kms_key.test2.arn
  backup_retention_period              = 5
  neptune_cluster_parameter_group_name = aws_neptune_cluster_parameter_group.test.id
  vpc_security_group_ids               = aws_security_group.test[*].id

  tags = {
    Name = %[1]q
  }
}
`, rName)
}

func testAccClusterConfig_storageType(rName, storageType string) string {
	return acctest.ConfigCompose(testAccClusterConfig_base(), fmt.Sprintf(`
resource "aws_neptune_cluster" "test" {
  cluster_identifier                   = %[1]q
  availability_zones                   = local.availability_zone_names
  engine                               = "neptune"
  engine_version                       = "1.3.0.0"
  neptune_cluster_parameter_group_name = "default.neptune1.3"
  skip_final_snapshot                  = true
  storage_type                         = %[2]q
  apply_immediately                    = true
}
`, rName, storageType))
}
