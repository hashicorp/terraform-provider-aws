// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package rds_test

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/YakDriver/regexache"
	rds_sdkv2 "github.com/aws/aws-sdk-go-v2/service/rds"
	"github.com/aws/aws-sdk-go-v2/service/rds/types"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/rds"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfrds "github.com/hashicorp/terraform-provider-aws/internal/service/rds"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func init() {
	acctest.RegisterServiceErrorCheckFunc(names.RDSServiceID, testAccErrorCheckSkip)
}

func testAccClusterImportStep(n string) resource.TestStep {
	return resource.TestStep{
		ResourceName:      n,
		ImportState:       true,
		ImportStateVerify: true,
		ImportStateVerifyIgnore: []string{
			names.AttrAllowMajorVersionUpgrade,
			names.AttrApplyImmediately,
			"cluster_members",
			"db_instance_parameter_group_name",
			"enable_global_write_forwarding",
			"enable_local_write_forwarding",
			"manage_master_user_password",
			"master_password",
			"master_user_secret_kms_key_id",
			"skip_final_snapshot",
		},
	}
}

func testAccErrorCheckSkip(t *testing.T) resource.ErrorCheckFunc {
	return acctest.ErrorCheckSkipMessagesContaining(t,
		"engine mode serverless you requested is currently unavailable",
		"engine mode multimaster you requested is currently unavailable",
		"requested engine version was not found or does not support parallelquery functionality",
		"Backtrack is not enabled for the aurora engine",
		"Read replica DB clusters are not available in this region for engine aurora",
		"no matching RDS Reserved Instance Offering found",
	)
}

func TestAccRDSCluster_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var dbCluster rds.DBCluster
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_rds_cluster.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.RDSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckClusterDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccClusterConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckClusterExists(ctx, resourceName, &dbCluster),
					acctest.CheckResourceAttrRegionalARN(resourceName, names.AttrARN, "rds", fmt.Sprintf("cluster:%s", rName)),
					resource.TestCheckResourceAttr(resourceName, "backtrack_window", acctest.Ct0),
					resource.TestCheckNoResourceAttr(resourceName, "ca_certificate_identifier"),
					resource.TestCheckNoResourceAttr(resourceName, "ca_certificate_valid_till"),
					resource.TestCheckResourceAttr(resourceName, names.AttrClusterIdentifier, rName),
					resource.TestCheckResourceAttr(resourceName, "cluster_identifier_prefix", ""),
					resource.TestCheckResourceAttrSet(resourceName, "cluster_resource_id"),
					resource.TestCheckResourceAttr(resourceName, "copy_tags_to_snapshot", acctest.CtFalse),
					resource.TestCheckResourceAttrSet(resourceName, "db_cluster_parameter_group_name"),
					resource.TestCheckResourceAttr(resourceName, "db_system_id", ""),
					resource.TestCheckResourceAttr(resourceName, "delete_automated_backups", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, names.AttrDomain, ""),
					resource.TestCheckResourceAttr(resourceName, "domain_iam_role_name", ""),
					resource.TestCheckResourceAttr(resourceName, "enabled_cloudwatch_logs_exports.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, names.AttrEngine, tfrds.ClusterEngineAuroraMySQL),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrEngineVersion),
					resource.TestCheckResourceAttr(resourceName, "global_cluster_identifier", ""),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrHostedZoneID),
					resource.TestCheckResourceAttr(resourceName, "network_type", "IPV4"),
					resource.TestCheckResourceAttrSet(resourceName, "reader_endpoint"),
					resource.TestCheckResourceAttr(resourceName, "scaling_configuration.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, names.AttrStorageEncrypted, acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct0),
				),
			},
			testAccClusterImportStep(resourceName),
		},
	})
}

func TestAccRDSCluster_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	var dbCluster rds.DBCluster
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_rds_cluster.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.RDSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckClusterDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccClusterConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterExists(ctx, resourceName, &dbCluster),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfrds.ResourceCluster(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccRDSCluster_identifierGenerated(t *testing.T) {
	ctx := acctest.Context(t)
	var v rds.DBCluster
	resourceName := "aws_rds_cluster.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.RDSServiceID),
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

func TestAccRDSCluster_identifierPrefix(t *testing.T) {
	ctx := acctest.Context(t)
	var v rds.DBCluster
	resourceName := "aws_rds_cluster.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.RDSServiceID),
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

func TestAccRDSCluster_tags(t *testing.T) {
	ctx := acctest.Context(t)
	var dbCluster1, dbCluster2, dbCluster3 rds.DBCluster
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_rds_cluster.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.RDSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckClusterDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccClusterConfig_tags1(rName, acctest.CtKey1, acctest.CtValue1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterExists(ctx, resourceName, &dbCluster1),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey1, acctest.CtValue1),
				),
			},
			testAccClusterImportStep(resourceName),
			{
				Config: testAccClusterConfig_tags2(rName, acctest.CtKey1, acctest.CtValue1Updated, acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterExists(ctx, resourceName, &dbCluster2),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct2),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey1, acctest.CtValue1Updated),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey2, acctest.CtValue2),
				),
			},
			{
				Config: testAccClusterConfig_tags1(rName, acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterExists(ctx, resourceName, &dbCluster3),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey2, acctest.CtValue2),
				),
			},
		},
	})
}

func TestAccRDSCluster_allowMajorVersionUpgrade(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var dbCluster1, dbCluster2 rds.DBCluster
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_rds_cluster.test"
	dataSourceName := "data.aws_rds_engine_version.test"
	dataSourceNameUpgrade := "data.aws_rds_engine_version.upgrade"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.RDSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckClusterDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccClusterConfig_allowMajorVersionUpgrade(rName, tfrds.ClusterEngineAuroraPostgreSQL, false, true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterExists(ctx, resourceName, &dbCluster1),
					resource.TestCheckResourceAttr(resourceName, names.AttrAllowMajorVersionUpgrade, acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, names.AttrEngine, tfrds.ClusterEngineAuroraPostgreSQL),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrEngineVersion, dataSourceName, "version_actual"),
				),
			},
			testAccClusterImportStep(resourceName),
			{
				Config: testAccClusterConfig_allowMajorVersionUpgrade(rName, tfrds.ClusterEngineAuroraPostgreSQL, true, true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterExists(ctx, resourceName, &dbCluster2),
					resource.TestCheckResourceAttr(resourceName, names.AttrAllowMajorVersionUpgrade, acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, names.AttrEngine, tfrds.ClusterEngineAuroraPostgreSQL),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrEngineVersion, dataSourceNameUpgrade, "version_actual"),
				),
			},
		},
	})
}

func TestAccRDSCluster_allowMajorVersionUpgradeNoApplyImmediately(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var dbCluster1, dbCluster2 rds.DBCluster
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_rds_cluster.test"
	dataSourceName := "data.aws_rds_engine_version.test"
	dataSourceNameUpgrade := "data.aws_rds_engine_version.upgrade"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.RDSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckClusterDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccClusterConfig_allowMajorVersionUpgrade(rName, tfrds.ClusterEngineAuroraPostgreSQL, false, false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterExists(ctx, resourceName, &dbCluster1),
					resource.TestCheckResourceAttr(resourceName, names.AttrAllowMajorVersionUpgrade, acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, names.AttrEngine, tfrds.ClusterEngineAuroraPostgreSQL),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrEngineVersion, dataSourceName, "version_actual"),
				),
			},
			testAccClusterImportStep(resourceName),
			{
				Config: testAccClusterConfig_allowMajorVersionUpgrade(rName, tfrds.ClusterEngineAuroraPostgreSQL, true, false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterExists(ctx, resourceName, &dbCluster2),
					resource.TestCheckResourceAttr(resourceName, names.AttrAllowMajorVersionUpgrade, acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, names.AttrEngine, tfrds.ClusterEngineAuroraPostgreSQL),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrEngineVersion, dataSourceNameUpgrade, "version_actual"),
				),
			},
		},
	})
}

func TestAccRDSCluster_allowMajorVersionUpgradeWithCustomParametersApplyImm(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var dbCluster1, dbCluster2 rds.DBCluster
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_rds_cluster.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.RDSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckClusterDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccClusterConfig_allowMajorVersionUpgradeCustomParameters(rName, false, tfrds.ClusterEngineAuroraPostgreSQL, true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterExists(ctx, resourceName, &dbCluster1),
					resource.TestCheckResourceAttr(resourceName, names.AttrAllowMajorVersionUpgrade, acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, names.AttrEngine, tfrds.ClusterEngineAuroraPostgreSQL),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrEngineVersion, "data.aws_rds_engine_version.test", "version_actual"),
				),
			},
			{
				Config: testAccClusterConfig_allowMajorVersionUpgradeCustomParameters(rName, true, tfrds.ClusterEngineAuroraPostgreSQL, true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterExists(ctx, resourceName, &dbCluster2),
					resource.TestCheckResourceAttr(resourceName, names.AttrAllowMajorVersionUpgrade, acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, names.AttrEngine, tfrds.ClusterEngineAuroraPostgreSQL),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrEngineVersion, "data.aws_rds_engine_version.upgrade", "version_actual"),
				),
			},
		},
	})
}

func TestAccRDSCluster_allowMajorVersionUpgradeWithCustomParameters(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var dbCluster1, dbCluster2 rds.DBCluster
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_rds_cluster.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.RDSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckClusterDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccClusterConfig_allowMajorVersionUpgradeCustomParameters(rName, false, tfrds.ClusterEngineAuroraPostgreSQL, false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterExists(ctx, resourceName, &dbCluster1),
					resource.TestCheckResourceAttr(resourceName, names.AttrAllowMajorVersionUpgrade, acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, names.AttrEngine, tfrds.ClusterEngineAuroraPostgreSQL),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrEngineVersion, "data.aws_rds_engine_version.test", "version_actual"),
				),
			},
			testAccClusterImportStep(resourceName),
			{
				Config: testAccClusterConfig_allowMajorVersionUpgradeCustomParameters(rName, true, tfrds.ClusterEngineAuroraPostgreSQL, false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterExists(ctx, resourceName, &dbCluster2),
					resource.TestCheckResourceAttr(resourceName, names.AttrAllowMajorVersionUpgrade, acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, names.AttrEngine, tfrds.ClusterEngineAuroraPostgreSQL),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrEngineVersion, "data.aws_rds_engine_version.upgrade", "version_actual"),
				),
			},
		},
	})
}

func TestAccRDSCluster_onlyMajorVersion(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var dbCluster1 rds.DBCluster
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_rds_cluster.test"
	dataSourceName := "data.aws_rds_engine_version.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.RDSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckClusterDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccClusterConfig_majorVersionOnly(rName, tfrds.ClusterEngineAuroraPostgreSQL, false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterExists(ctx, resourceName, &dbCluster1),
					resource.TestCheckResourceAttr(resourceName, names.AttrEngine, tfrds.ClusterEngineAuroraPostgreSQL),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrEngineVersion, dataSourceName, "version_actual"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					names.AttrAllowMajorVersionUpgrade,
					names.AttrApplyImmediately,
					"cluster_identifier_prefix",
					"cluster_members",
					"db_instance_parameter_group_name",
					"enable_global_write_forwarding",
					"enable_local_write_forwarding",
					names.AttrEngineVersion,
					"master_password",
					"skip_final_snapshot",
				},
			},
		},
	})
}

// https://github.com/hashicorp/terraform-provider-aws/issues/30605
func TestAccRDSCluster_minorVersion(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var dbCluster1, dbCluster2 rds.DBCluster
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_rds_cluster.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.RDSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckClusterDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccClusterConfig_minorVersion(rName, tfrds.ClusterEngineAuroraPostgreSQL, false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterExists(ctx, resourceName, &dbCluster1),
					resource.TestCheckResourceAttr(resourceName, names.AttrEngine, tfrds.ClusterEngineAuroraPostgreSQL),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrEngineVersion, "data.aws_rds_engine_version.test", "version_actual"),
				),
			},
			{
				Config: testAccClusterConfig_minorVersion(rName, tfrds.ClusterEngineAuroraPostgreSQL, true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterExists(ctx, resourceName, &dbCluster2),
					resource.TestCheckResourceAttr(resourceName, names.AttrEngine, tfrds.ClusterEngineAuroraPostgreSQL),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrEngineVersion, "data.aws_rds_engine_version.upgrade", "version_actual"),
				),
			},
		},
	})
}

func TestAccRDSCluster_availabilityZones(t *testing.T) {
	ctx := acctest.Context(t)
	var dbCluster rds.DBCluster
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_rds_cluster.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.RDSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckClusterDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccClusterConfig_availabilityZones(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterExists(ctx, resourceName, &dbCluster),
				),
			},
		},
	})
}

func TestAccRDSCluster_availabilityZones_caCertificateIdentifier(t *testing.T) {
	ctx := acctest.Context(t)
	var dbCluster rds.DBCluster
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_rds_cluster.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.RDSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckClusterDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccClusterConfig_availabilityZones_caCertificateIdentifier(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterExists(ctx, resourceName, &dbCluster),
					resource.TestCheckResourceAttrSet(resourceName, "ca_certificate_identifier"),
					resource.TestCheckResourceAttrSet(resourceName, "ca_certificate_valid_till"),
				),
			},
		},
	})
}

func TestAccRDSCluster_storageTypeIo1(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	ctx := acctest.Context(t)
	var dbCluster rds.DBCluster
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_rds_cluster.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.RDSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckClusterDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccClusterConfig_storageType(rName, "io1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterExists(ctx, resourceName, &dbCluster),
					resource.TestCheckResourceAttr(resourceName, names.AttrStorageType, "io1"),
				),
			},
		},
	})
}

func TestAccRDSCluster_storageTypeIo2(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	ctx := acctest.Context(t)
	var dbCluster rds.DBCluster
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_rds_cluster.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.RDSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckClusterDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccClusterConfig_storageType(rName, "io2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterExists(ctx, resourceName, &dbCluster),
					resource.TestCheckResourceAttr(resourceName, names.AttrStorageType, "io2"),
				),
			},
		},
	})
}

// For backwards compatibility, the control plane should always return a blank string even if sending "aurora" as the storage type
func TestAccRDSCluster_storageTypeAuroraReturnsBlank(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	ctx := acctest.Context(t)
	var dbCluster1 rds.DBCluster
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	storageTypeAurora := "aurora"
	storageTypeEmpty := ""
	resourceName := "aws_rds_cluster.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.RDSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckClusterDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccClusterConfig_auroraStorageType(rName, storageTypeAurora),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterExists(ctx, resourceName, &dbCluster1),
					resource.TestCheckResourceAttr(resourceName, names.AttrStorageType, storageTypeEmpty),
				),
			},
		},
	})
}

func TestAccRDSCluster_storageTypeAuroraIopt1(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	ctx := acctest.Context(t)
	var dbCluster rds.DBCluster
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	storageType := "aurora-iopt1"
	resourceName := "aws_rds_cluster.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.RDSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckClusterDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccClusterConfig_auroraStorageType(rName, storageType),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterExists(ctx, resourceName, &dbCluster),
					resource.TestCheckResourceAttr(resourceName, names.AttrStorageType, storageType),
				),
			},
		},
	})
}

func TestAccRDSCluster_storageTypeAuroraUpdateAuroraIopt1(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	ctx := acctest.Context(t)
	var dbCluster1, dbCluster2 rds.DBCluster
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	storageTypeEmpty := ""
	storageTypeAuroraIOPT1 := "aurora-iopt1"
	resourceName := "aws_rds_cluster.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.RDSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckClusterDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccClusterConfig_auroraStorageTypeNotDefined(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterExists(ctx, resourceName, &dbCluster1),
					resource.TestCheckResourceAttr(resourceName, names.AttrStorageType, storageTypeEmpty),
				),
			},
			{
				Config: testAccClusterConfig_auroraStorageType(rName, storageTypeAuroraIOPT1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterExists(ctx, resourceName, &dbCluster2),
					testAccCheckClusterNotRecreated(&dbCluster1, &dbCluster2),
					resource.TestCheckResourceAttr(resourceName, names.AttrStorageType, storageTypeAuroraIOPT1),
				),
			},
		},
	})
}

func TestAccRDSCluster_allocatedStorage(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	ctx := acctest.Context(t)
	var dbCluster rds.DBCluster
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_rds_cluster.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.RDSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckClusterDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccClusterConfig_allocatedStorage(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterExists(ctx, resourceName, &dbCluster),
					resource.TestCheckResourceAttr(resourceName, names.AttrAllocatedStorage, "100"),
				),
			},
		},
	})
}

// Verify storage_type from aurora-iopt1 to aurora
func TestAccRDSCluster_storageTypeAuroraIopt1UpdateAurora(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	ctx := acctest.Context(t)
	var dbCluster1, dbCluster2 rds.DBCluster
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	storageTypeAuroraIOPT1 := "aurora-iopt1"
	storageTypeAurora := "aurora"
	storageTypeEmpty := ""
	resourceName := "aws_rds_cluster.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.RDSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckClusterDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccClusterConfig_auroraStorageType(rName, storageTypeAuroraIOPT1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterExists(ctx, resourceName, &dbCluster1),
					resource.TestCheckResourceAttr(resourceName, names.AttrStorageType, storageTypeAuroraIOPT1),
				),
			},
			{
				Config: testAccClusterConfig_auroraStorageType(rName, storageTypeAurora),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterExists(ctx, resourceName, &dbCluster2),
					testAccCheckClusterNotRecreated(&dbCluster1, &dbCluster2),
					resource.TestCheckResourceAttr(resourceName, names.AttrStorageType, storageTypeEmpty),
				),
			},
		},
	})
}

func TestAccRDSCluster_iops(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	ctx := acctest.Context(t)
	var dbCluster rds.DBCluster
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_rds_cluster.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.RDSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckClusterDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccClusterConfig_iops(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterExists(ctx, resourceName, &dbCluster),
					resource.TestCheckResourceAttr(resourceName, names.AttrIOPS, "1000"),
				),
			},
		},
	})
}

func TestAccRDSCluster_dbClusterInstanceClass(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	ctx := acctest.Context(t)
	var dbCluster rds.DBCluster
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_rds_cluster.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.RDSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckClusterDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccClusterConfig_dbClusterInstanceClass(rName, true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterExists(ctx, resourceName, &dbCluster),
					resource.TestCheckResourceAttrSet(resourceName, "db_cluster_instance_class"),
				),
			},
			{
				Config: testAccClusterConfig_dbClusterInstanceClass(rName, false), // guaranteed to be a different instance class because the intersection of instance class lists is empty
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterExists(ctx, resourceName, &dbCluster),
					resource.TestCheckResourceAttrSet(resourceName, "db_cluster_instance_class"),
				),
			},
		},
	})
}

func TestAccRDSCluster_backtrackWindow(t *testing.T) {
	ctx := acctest.Context(t)
	var dbCluster rds.DBCluster
	resourceName := "aws_rds_cluster.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.RDSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckClusterDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccClusterConfig_backtrackWindow(43200),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterExists(ctx, resourceName, &dbCluster),
					resource.TestCheckResourceAttr(resourceName, "backtrack_window", "43200"),
				),
			},
			{
				Config: testAccClusterConfig_backtrackWindow(86400),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterExists(ctx, resourceName, &dbCluster),
					resource.TestCheckResourceAttr(resourceName, "backtrack_window", "86400"),
				),
			},
		},
	})
}

func TestAccRDSCluster_dbSubnetGroupName(t *testing.T) {
	ctx := acctest.Context(t)
	var dbCluster rds.DBCluster
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_rds_cluster.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.RDSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckClusterDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccClusterConfig_subnetGroupName(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterExists(ctx, resourceName, &dbCluster),
				),
			},
		},
	})
}

func TestAccRDSCluster_pointInTimeRestore(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var sourceDBCluster, dbCluster rds.DBCluster
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	sourceResourceName := "aws_rds_cluster.test"
	resourceName := "aws_rds_cluster.restore"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.RDSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckClusterDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccClusterConfig_pointInTimeRestoreSource(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterExists(ctx, sourceResourceName, &sourceDBCluster),
					testAccCheckClusterExists(ctx, resourceName, &dbCluster),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrEngine, sourceResourceName, names.AttrEngine),
				),
			},
		},
	})
}

func TestAccRDSCluster_PointInTimeRestore_enabledCloudWatchLogsExports(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var sourceDBCluster, dbCluster rds.DBCluster
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	sourceResourceName := "aws_rds_cluster.test"
	resourceName := "aws_rds_cluster.restore"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.RDSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckClusterDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccClusterConfig_pointInTimeRestoreSource_enabledCloudWatchLogsExports(rName, "audit"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterExists(ctx, sourceResourceName, &sourceDBCluster),
					testAccCheckClusterExists(ctx, resourceName, &dbCluster),
					resource.TestCheckResourceAttr(resourceName, "enabled_cloudwatch_logs_exports.#", acctest.Ct1),
					resource.TestCheckTypeSetElemAttr(resourceName, "enabled_cloudwatch_logs_exports.*", "audit"),
				),
			},
		},
	})
}

func TestAccRDSCluster_takeFinalSnapshot(t *testing.T) {
	ctx := acctest.Context(t)
	var v rds.DBCluster
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_rds_cluster.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.RDSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckClusterDestroyWithFinalSnapshot(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccClusterConfig_finalSnapshot(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterExists(ctx, resourceName, &v),
				),
			},
		},
	})
}

// This is a regression test to make sure that we always cover the scenario as highlighted in
// https://github.com/hashicorp/terraform/issues/11568
// Expected error updated to match API response
func TestAccRDSCluster_missingUserNameCausesError(t *testing.T) {
	ctx := acctest.Context(t)
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.RDSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckClusterDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config:      testAccClusterConfig_withoutUserNameAndPassword(sdkacctest.RandInt()),
				ExpectError: regexache.MustCompile(`InvalidParameterValue: The parameter MasterUsername must be provided`),
			},
		},
	})
}

func TestAccRDSCluster_domain(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var dbCluster rds.DBCluster
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_rds_cluster.test"

	domain := acctest.RandomDomainName()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.RDSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckClusterDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccClusterConfig_domain(rName, domain),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterExists(ctx, resourceName, &dbCluster),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrDomain),
					resource.TestCheckResourceAttrSet(resourceName, "domain_iam_role_name"),
				),
			},
		},
	})
}

func TestAccRDSCluster_EnabledCloudWatchLogsExports_mySQL(t *testing.T) {
	ctx := acctest.Context(t)
	var dbCluster1, dbCluster2, dbCluster3 rds.DBCluster
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_rds_cluster.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.RDSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckClusterDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccClusterConfig_enabledCloudWatchLogsExports1(rName, "audit"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterExists(ctx, resourceName, &dbCluster1),
					resource.TestCheckResourceAttr(resourceName, "enabled_cloudwatch_logs_exports.#", acctest.Ct1),
					resource.TestCheckTypeSetElemAttr(resourceName, "enabled_cloudwatch_logs_exports.*", "audit"),
				),
			},
			{
				Config: testAccClusterConfig_enabledCloudWatchLogsExports2(rName, "slowquery", "error"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterExists(ctx, resourceName, &dbCluster2),
					resource.TestCheckResourceAttr(resourceName, "enabled_cloudwatch_logs_exports.#", acctest.Ct2),
					resource.TestCheckTypeSetElemAttr(resourceName, "enabled_cloudwatch_logs_exports.*", "error"),
					resource.TestCheckTypeSetElemAttr(resourceName, "enabled_cloudwatch_logs_exports.*", "slowquery"),
				),
			},
			{
				Config: testAccClusterConfig_enabledCloudWatchLogsExports1(rName, "error"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterExists(ctx, resourceName, &dbCluster3),
					resource.TestCheckResourceAttr(resourceName, "enabled_cloudwatch_logs_exports.#", acctest.Ct1),
					resource.TestCheckTypeSetElemAttr(resourceName, "enabled_cloudwatch_logs_exports.*", "error"),
				),
			},
		},
	})
}

func TestAccRDSCluster_EnabledCloudWatchLogsExports_postgresql(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var dbCluster1, dbCluster2 rds.DBCluster
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_rds_cluster.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.RDSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckClusterDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccClusterConfig_enabledCloudWatchLogsExportsPostgreSQL1(rName, "postgresql"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterExists(ctx, resourceName, &dbCluster1),
					resource.TestCheckResourceAttr(resourceName, "enabled_cloudwatch_logs_exports.#", acctest.Ct1),
					resource.TestCheckTypeSetElemAttr(resourceName, "enabled_cloudwatch_logs_exports.*", "postgresql"),
				),
			},
			{
				Config: testAccClusterConfig_enabledCloudWatchLogsExportsPostgreSQL2(rName, "postgresql", "upgrade"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterExists(ctx, resourceName, &dbCluster2),
					resource.TestCheckResourceAttr(resourceName, "enabled_cloudwatch_logs_exports.#", acctest.Ct2),
					resource.TestCheckTypeSetElemAttr(resourceName, "enabled_cloudwatch_logs_exports.*", "postgresql"),
					resource.TestCheckTypeSetElemAttr(resourceName, "enabled_cloudwatch_logs_exports.*", "upgrade"),
				),
			},
		},
	})
}

func TestAccRDSCluster_updateIAMRoles(t *testing.T) {
	ctx := acctest.Context(t)
	var v rds.DBCluster
	ri := sdkacctest.RandInt()
	resourceName := "aws_rds_cluster.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.RDSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckClusterDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccClusterConfig_includingIAMRoles(ri),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterExists(ctx, resourceName, &v),
				),
			},
			{
				Config: testAccClusterConfig_addIAMRoles(ri),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(
						resourceName, "iam_roles.#", acctest.Ct2),
				),
			},
			{
				Config: testAccClusterConfig_removeIAMRoles(ri),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(
						resourceName, "iam_roles.#", acctest.Ct1),
				),
			},
		},
	})
}

func TestAccRDSCluster_kmsKey(t *testing.T) {
	ctx := acctest.Context(t)
	var dbCluster1 rds.DBCluster
	kmsKeyResourceName := "aws_kms_key.foo"
	resourceName := "aws_rds_cluster.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.RDSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckClusterDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccClusterConfig_kmsKey(sdkacctest.RandInt()),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterExists(ctx, resourceName, &dbCluster1),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrKMSKeyID, kmsKeyResourceName, names.AttrARN),
				),
			},
		},
	})
}

func TestAccRDSCluster_networkType(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var dbCluster rds.DBCluster
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_rds_cluster.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.RDSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckClusterDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccClusterConfig_networkType(rName, "IPV4"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterExists(ctx, resourceName, &dbCluster),
					resource.TestCheckResourceAttr(resourceName, "network_type", "IPV4"),
				),
			},
			{
				Config: testAccClusterConfig_networkType(rName, "DUAL"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterExists(ctx, resourceName, &dbCluster),
					resource.TestCheckResourceAttr(resourceName, "network_type", "DUAL"),
				),
			},
		},
	})
}

func TestAccRDSCluster_encrypted(t *testing.T) {
	ctx := acctest.Context(t)
	var v rds.DBCluster
	resourceName := "aws_rds_cluster.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.RDSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckClusterDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccClusterConfig_encrypted(sdkacctest.RandInt()),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, names.AttrStorageEncrypted, acctest.CtTrue),
				),
			},
		},
	})
}

func TestAccRDSCluster_copyTagsToSnapshot(t *testing.T) {
	ctx := acctest.Context(t)
	var v rds.DBCluster
	rInt := sdkacctest.RandInt()
	resourceName := "aws_rds_cluster.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.RDSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckClusterDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccClusterConfig_copyTagsToSnapshot(rInt, true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "copy_tags_to_snapshot", acctest.CtTrue),
				),
			},
			{
				Config: testAccClusterConfig_copyTagsToSnapshot(rInt, false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "copy_tags_to_snapshot", acctest.CtFalse),
				),
			},
			{
				Config: testAccClusterConfig_copyTagsToSnapshot(rInt, true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "copy_tags_to_snapshot", acctest.CtTrue),
				),
			},
		},
	})
}

func TestAccRDSCluster_copyTagsToSnapshot_restorePointInTime(t *testing.T) {
	ctx := acctest.Context(t)
	var v rds.DBCluster
	rInt := sdkacctest.RandInt()
	resourceName := "aws_rds_cluster.test"
	resourceName2 := "aws_rds_cluster.clone"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.RDSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckClusterDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccClusterConfig_copyTagsToSnapshot_restorePointInTime(rInt, true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterExists(ctx, resourceName, &v),
					testAccCheckClusterExists(ctx, resourceName2, &v),
					resource.TestCheckResourceAttr(resourceName, "copy_tags_to_snapshot", acctest.CtTrue),
				),
			},
			{
				Config: testAccClusterConfig_copyTagsToSnapshot_restorePointInTime(rInt, false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterExists(ctx, resourceName, &v),
					testAccCheckClusterExists(ctx, resourceName2, &v),
					resource.TestCheckResourceAttr(resourceName2, "copy_tags_to_snapshot", acctest.CtFalse),
				),
			},
		},
	})
}

func TestAccRDSCluster_ReplicationSourceIdentifier_kmsKeyID(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var primaryCluster rds.DBCluster
	var replicaCluster rds.DBCluster
	resourceName := "aws_rds_cluster.test"
	resourceName2 := "aws_rds_cluster.alternate"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	// record the initialized providers so that we can use them to
	// check for the cluster in each region
	var providers []*schema.Provider

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckMultipleRegion(t, 2)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.RDSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5FactoriesPlusProvidersAlternate(ctx, t, &providers),
		CheckDestroy:             acctest.CheckWithProviders(testAccCheckClusterDestroyWithProvider(ctx), &providers),
		Steps: []resource.TestStep{
			{
				Config: testAccClusterConfig_replicationSourceIDKMSKeyID(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterExistsWithProvider(ctx, resourceName, &primaryCluster, acctest.RegionProviderFunc(acctest.Region(), &providers)),
					testAccCheckClusterExistsWithProvider(ctx, resourceName2, &replicaCluster, acctest.RegionProviderFunc(acctest.AlternateRegion(), &providers)),
				),
			},
		},
	})
}

func TestAccRDSCluster_backupsUpdate(t *testing.T) {
	ctx := acctest.Context(t)
	var v rds.DBCluster
	resourceName := "aws_rds_cluster.test"

	ri := sdkacctest.RandInt()
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.RDSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckClusterDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccClusterConfig_backups(ri),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(
						resourceName, "preferred_backup_window", "07:00-09:00"),
					resource.TestCheckResourceAttr(
						resourceName, "backup_retention_period", "5"),
					resource.TestCheckResourceAttr(
						resourceName, names.AttrPreferredMaintenanceWindow, "tue:04:00-tue:04:30"),
				),
			},
			{
				Config: testAccClusterConfig_backupsUpdate(ri),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(
						resourceName, "preferred_backup_window", "03:00-09:00"),
					resource.TestCheckResourceAttr(
						resourceName, "backup_retention_period", acctest.Ct10),
					resource.TestCheckResourceAttr(
						resourceName, names.AttrPreferredMaintenanceWindow, "wed:01:00-wed:01:30"),
				),
			},
		},
	})
}

func TestAccRDSCluster_iamAuth(t *testing.T) {
	ctx := acctest.Context(t)
	var v rds.DBCluster
	resourceName := "aws_rds_cluster.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.RDSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckClusterDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccClusterConfig_iamAuth(sdkacctest.RandInt()),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(
						resourceName, "iam_database_authentication_enabled", acctest.CtTrue),
				),
			},
		},
	})
}

func TestAccRDSCluster_deletionProtection(t *testing.T) {
	ctx := acctest.Context(t)
	var dbCluster1 rds.DBCluster
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_rds_cluster.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.RDSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckClusterDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccClusterConfig_deletionProtection(rName, true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterExists(ctx, resourceName, &dbCluster1),
					resource.TestCheckResourceAttr(resourceName, names.AttrDeletionProtection, acctest.CtTrue),
				),
			},
			{
				Config: testAccClusterConfig_deletionProtection(rName, false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterExists(ctx, resourceName, &dbCluster1),
					resource.TestCheckResourceAttr(resourceName, names.AttrDeletionProtection, acctest.CtFalse),
				),
			},
		},
	})
}

func TestAccRDSCluster_engineMode(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var dbCluster1, dbCluster2 rds.DBCluster

	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_rds_cluster.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.RDSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckClusterDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccClusterConfig_engineMode_serverless(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterExists(ctx, resourceName, &dbCluster1),
					resource.TestCheckResourceAttr(resourceName, "engine_mode", "serverless"),
					resource.TestCheckResourceAttr(resourceName, "scaling_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "serverlessv2_scaling_configuration.#", acctest.Ct0),
				),
			},
			{
				Config: testAccClusterConfig_engineMode(rName, "provisioned"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterExists(ctx, resourceName, &dbCluster2),
					testAccCheckClusterRecreated(&dbCluster1, &dbCluster2),
					resource.TestCheckResourceAttr(resourceName, "engine_mode", "provisioned"),
					resource.TestCheckResourceAttr(resourceName, "scaling_configuration.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "serverlessv2_scaling_configuration.#", acctest.Ct0),
				),
			},
		},
	})
}

func TestAccRDSCluster_engineVersion(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var dbCluster rds.DBCluster
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_rds_cluster.test"
	dataSourceName := "data.aws_rds_engine_version.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.RDSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckClusterDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccClusterConfig_engineVersion(rName, false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterExists(ctx, resourceName, &dbCluster),
					resource.TestCheckResourceAttr(resourceName, names.AttrEngine, tfrds.ClusterEngineAuroraPostgreSQL),
					resource.TestCheckResourceAttr(resourceName, "engine_lifecycle_support", "open-source-rds-extended-support"),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrEngineVersion, dataSourceName, names.AttrVersion),
				),
			},
			{
				Config:      testAccClusterConfig_engineVersion(rName, true),
				ExpectError: regexache.MustCompile(`Cannot modify engine version without a healthy primary instance in DB cluster`),
			},
		},
	})
}

func TestAccRDSCluster_engineVersionWithPrimaryInstance(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var dbCluster rds.DBCluster
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_rds_cluster.test"
	dataSourceName := "data.aws_rds_engine_version.test"
	dataSourceNameUpgrade := "data.aws_rds_engine_version.upgrade"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.RDSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckClusterDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccClusterConfig_engineVersionPrimaryInstance(rName, false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterExists(ctx, resourceName, &dbCluster),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrEngine, dataSourceName, names.AttrEngine),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrEngineVersion, dataSourceName, names.AttrVersion),
				),
			},
			{
				Config: testAccClusterConfig_engineVersionPrimaryInstance(rName, true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterExists(ctx, resourceName, &dbCluster),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrEngine, dataSourceNameUpgrade, names.AttrEngine),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrEngineVersion, dataSourceNameUpgrade, names.AttrVersion),
				),
			},
		},
	})
}

func TestAccRDSCluster_GlobalClusterIdentifierEngineMode_global(t *testing.T) {
	ctx := acctest.Context(t)
	var dbCluster1 rds.DBCluster

	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	globalClusterResourceName := "aws_rds_global_cluster.test"
	resourceName := "aws_rds_cluster.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheckGlobalCluster(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.RDSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckClusterDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccClusterConfig_GlobalClusterID_EngineMode_global(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterExists(ctx, resourceName, &dbCluster1),
					resource.TestCheckResourceAttrPair(resourceName, "global_cluster_identifier", globalClusterResourceName, names.AttrID),
				),
			},
		},
	})
}

func TestAccRDSCluster_GlobalClusterIdentifierEngineModeGlobal_add(t *testing.T) {
	ctx := acctest.Context(t)
	var dbCluster1 rds.DBCluster

	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_rds_cluster.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheckGlobalCluster(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.RDSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckClusterDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccClusterConfig_EngineMode_global(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterExists(ctx, resourceName, &dbCluster1),
					resource.TestCheckResourceAttr(resourceName, "global_cluster_identifier", ""),
				),
			},
			{
				Config:      testAccClusterConfig_GlobalClusterID_EngineMode_global(rName),
				ExpectError: regexache.MustCompile(`(?i)Existing RDS Clusters cannot be added to an existing RDS Global Cluster`),
			},
		},
	})
}

func TestAccRDSCluster_GlobalClusterIdentifierEngineModeGlobal_remove(t *testing.T) {
	ctx := acctest.Context(t)
	var dbCluster1 rds.DBCluster

	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	globalClusterResourceName := "aws_rds_global_cluster.test"
	resourceName := "aws_rds_cluster.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheckGlobalCluster(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.RDSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckClusterDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccClusterConfig_GlobalClusterID_EngineMode_global(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterExists(ctx, resourceName, &dbCluster1),
					resource.TestCheckResourceAttrPair(resourceName, "global_cluster_identifier", globalClusterResourceName, names.AttrID),
				),
			},
			{
				Config: testAccClusterConfig_EngineMode_global(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterExists(ctx, resourceName, &dbCluster1),
					resource.TestCheckResourceAttr(resourceName, "global_cluster_identifier", ""),
				),
			},
		},
	})
}

func TestAccRDSCluster_GlobalClusterIdentifierEngineModeGlobal_update(t *testing.T) {
	ctx := acctest.Context(t)
	var dbCluster1 rds.DBCluster

	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	globalClusterResourceName1 := "aws_rds_global_cluster.test.0"
	globalClusterResourceName2 := "aws_rds_global_cluster.test.1"
	resourceName := "aws_rds_cluster.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheckGlobalCluster(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.RDSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckClusterDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccClusterConfig_GlobalClusterID_EngineMode_globalUpdate(rName, globalClusterResourceName1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterExists(ctx, resourceName, &dbCluster1),
					resource.TestCheckResourceAttrPair(resourceName, "global_cluster_identifier", globalClusterResourceName1, names.AttrID),
				),
			},
			{
				Config:      testAccClusterConfig_GlobalClusterID_EngineMode_globalUpdate(rName, globalClusterResourceName2),
				ExpectError: regexache.MustCompile(`(?i)Existing RDS Clusters cannot be migrated between existing RDS Global Clusters`),
			},
		},
	})
}

func TestAccRDSCluster_GlobalClusterIdentifierEngineMode_provisioned(t *testing.T) {
	ctx := acctest.Context(t)
	var dbCluster1 rds.DBCluster

	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	globalClusterResourceName := "aws_rds_global_cluster.test"
	resourceName := "aws_rds_cluster.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheckGlobalCluster(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.RDSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckClusterDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccClusterConfig_GlobalClusterID_EngineMode_provisioned(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterExists(ctx, resourceName, &dbCluster1),
					resource.TestCheckResourceAttrPair(resourceName, "global_cluster_identifier", globalClusterResourceName, names.AttrID),
				),
			},
		},
	})
}

// Reference: https://github.com/hashicorp/terraform-provider-aws/issues/13126
func TestAccRDSCluster_GlobalClusterIdentifier_primarySecondaryClusters(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var providers []*schema.Provider
	var primaryDbCluster, secondaryDbCluster rds.DBCluster

	rNameGlobal := sdkacctest.RandomWithPrefix("tf-acc-test-global")
	rNamePrimary := sdkacctest.RandomWithPrefix("tf-acc-test-primary")
	rNameSecondary := sdkacctest.RandomWithPrefix("tf-acc-test-secondary")

	resourceNamePrimary := "aws_rds_cluster.primary"
	resourceNameSecondary := "aws_rds_cluster.secondary"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckMultipleRegion(t, 2)
			testAccPreCheckGlobalCluster(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.RDSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5FactoriesPlusProvidersAlternate(ctx, t, &providers),
		CheckDestroy:             testAccCheckClusterDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccClusterConfig_GlobalClusterID_primarySecondaryClusters(rNameGlobal, rNamePrimary, rNameSecondary),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterExistsWithProvider(ctx, resourceNamePrimary, &primaryDbCluster, acctest.RegionProviderFunc(acctest.Region(), &providers)),
					testAccCheckClusterExistsWithProvider(ctx, resourceNameSecondary, &secondaryDbCluster, acctest.RegionProviderFunc(acctest.AlternateRegion(), &providers)),
				),
			},
		},
	})
}

// Reference: https://github.com/hashicorp/terraform-provider-aws/issues/13715
func TestAccRDSCluster_GlobalClusterIdentifier_replicationSourceIdentifier(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var providers []*schema.Provider
	var primaryDbCluster, secondaryDbCluster rds.DBCluster

	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceNamePrimary := "aws_rds_cluster.primary"
	resourceNameSecondary := "aws_rds_cluster.secondary"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckMultipleRegion(t, 2)
			testAccPreCheckGlobalCluster(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.RDSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5FactoriesPlusProvidersAlternate(ctx, t, &providers),
		CheckDestroy:             testAccCheckClusterDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccClusterConfig_GlobalClusterID_replicationSourceID(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterExistsWithProvider(ctx, resourceNamePrimary, &primaryDbCluster, acctest.RegionProviderFunc(acctest.Region(), &providers)),
					testAccCheckClusterExistsWithProvider(ctx, resourceNameSecondary, &secondaryDbCluster, acctest.RegionProviderFunc(acctest.AlternateRegion(), &providers)),
				),
			},
		},
	})
}

// Reference: https://github.com/hashicorp/terraform-provider-aws/issues/14457
func TestAccRDSCluster_GlobalClusterIdentifier_secondaryClustersWriteForwarding(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var providers []*schema.Provider
	var primaryDbCluster, secondaryDbCluster rds.DBCluster

	rNameGlobal := sdkacctest.RandomWithPrefix("tf-acc-test-global")
	rNamePrimary := sdkacctest.RandomWithPrefix("tf-acc-test-primary")
	rNameSecondary := sdkacctest.RandomWithPrefix("tf-acc-test-secondary")

	resourceNamePrimary := "aws_rds_cluster.primary"
	resourceNameSecondary := "aws_rds_cluster.secondary"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckMultipleRegion(t, 2)
			testAccPreCheckGlobalCluster(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.RDSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5FactoriesPlusProvidersAlternate(ctx, t, &providers),
		CheckDestroy:             testAccCheckClusterDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccClusterConfig_GlobalClusterID_secondaryClustersWriteForwarding(rNameGlobal, rNamePrimary, rNameSecondary),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterExistsWithProvider(ctx, resourceNamePrimary, &primaryDbCluster, acctest.RegionProviderFunc(acctest.Region(), &providers)),
					testAccCheckClusterExistsWithProvider(ctx, resourceNameSecondary, &secondaryDbCluster, acctest.RegionProviderFunc(acctest.AlternateRegion(), &providers)),
					resource.TestCheckResourceAttr(resourceNameSecondary, "enable_global_write_forwarding", acctest.CtTrue),
				),
			},
		},
	})
}

func TestAccRDSCluster_ManagedMasterPassword_managed(t *testing.T) {
	ctx := acctest.Context(t)
	var dbCluster rds.DBCluster
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_rds_cluster.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.RDSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckClusterDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccClusterConfig_managedMasterPassword(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckClusterExists(ctx, resourceName, &dbCluster),
					acctest.CheckResourceAttrRegionalARN(resourceName, names.AttrARN, "rds", fmt.Sprintf("cluster:%s", rName)),
					resource.TestCheckResourceAttrSet(resourceName, "cluster_resource_id"),
					resource.TestCheckResourceAttr(resourceName, names.AttrEngine, tfrds.ClusterEngineAuroraMySQL),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrEngineVersion),
					resource.TestCheckResourceAttr(resourceName, "manage_master_user_password", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "master_user_secret.#", acctest.Ct1),
					resource.TestCheckResourceAttrSet(resourceName, "master_user_secret.0.kms_key_id"),
					resource.TestCheckResourceAttrSet(resourceName, "master_user_secret.0.secret_arn"),
					resource.TestCheckResourceAttrSet(resourceName, "master_user_secret.0.secret_status"),
				),
			},
			testAccClusterImportStep(resourceName),
		},
	})
}

func TestAccRDSCluster_ManagedMasterPassword_managedSpecificKMSKey(t *testing.T) {
	ctx := acctest.Context(t)
	var dbCluster rds.DBCluster
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_rds_cluster.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.RDSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckClusterDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccClusterConfig_managedMasterPasswordKMSKey(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckClusterExists(ctx, resourceName, &dbCluster),
					acctest.CheckResourceAttrRegionalARN(resourceName, names.AttrARN, "rds", fmt.Sprintf("cluster:%s", rName)),
					resource.TestCheckResourceAttrSet(resourceName, "cluster_resource_id"),
					resource.TestCheckResourceAttr(resourceName, names.AttrEngine, tfrds.ClusterEngineAuroraMySQL),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrEngineVersion),
					resource.TestCheckResourceAttr(resourceName, "manage_master_user_password", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "master_user_secret.#", acctest.Ct1),
					resource.TestCheckResourceAttrSet(resourceName, "master_user_secret.0.kms_key_id"),
					resource.TestCheckResourceAttrSet(resourceName, "master_user_secret.0.secret_arn"),
					resource.TestCheckResourceAttrSet(resourceName, "master_user_secret.0.secret_status"),
				),
			},
			testAccClusterImportStep(resourceName),
		},
	})
}

func TestAccRDSCluster_ManagedMasterPassword_convertToManaged(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var dbCluster1, dbCluster2 rds.DBCluster
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_rds_cluster.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.RDSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckClusterDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccClusterConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckClusterExists(ctx, resourceName, &dbCluster1),
					acctest.CheckResourceAttrRegionalARN(resourceName, names.AttrARN, "rds", fmt.Sprintf("cluster:%s", rName)),
					resource.TestCheckResourceAttrSet(resourceName, "cluster_resource_id"),
					resource.TestCheckNoResourceAttr(resourceName, "manage_master_user_password"),
				),
			},
			testAccClusterImportStep(resourceName),
			{
				Config: testAccClusterConfig_managedMasterPassword(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckClusterExists(ctx, resourceName, &dbCluster2),
					acctest.CheckResourceAttrRegionalARN(resourceName, names.AttrARN, "rds", fmt.Sprintf("cluster:%s", rName)),
					resource.TestCheckResourceAttrSet(resourceName, "cluster_resource_id"),
					resource.TestCheckResourceAttrSet(resourceName, "manage_master_user_password"),
					resource.TestCheckResourceAttr(resourceName, "manage_master_user_password", acctest.CtTrue),
				),
			},
		},
	})
}

func TestAccRDSCluster_port(t *testing.T) {
	ctx := acctest.Context(t)
	var dbCluster1, dbCluster2 rds.DBCluster
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_rds_cluster.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.RDSServiceID),
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
				Config: testAccClusterConfig_port(rName, 2345),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterExists(ctx, resourceName, &dbCluster2),
					resource.TestCheckResourceAttr(resourceName, names.AttrPort, "2345"),
				),
			},
		},
	})
}

func TestAccRDSCluster_scaling(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var dbCluster rds.DBCluster

	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_rds_cluster.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.RDSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckClusterDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccClusterConfig_scalingConfiguration(rName, false, 128, 4, 301, 301, "RollbackCapacityChange"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterExists(ctx, resourceName, &dbCluster),
					resource.TestCheckResourceAttr(resourceName, "scaling_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "scaling_configuration.0.auto_pause", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "scaling_configuration.0.max_capacity", "128"),
					resource.TestCheckResourceAttr(resourceName, "scaling_configuration.0.min_capacity", acctest.Ct4),
					resource.TestCheckResourceAttr(resourceName, "scaling_configuration.0.seconds_before_timeout", "301"),
					resource.TestCheckResourceAttr(resourceName, "scaling_configuration.0.seconds_until_auto_pause", "301"),
					resource.TestCheckResourceAttr(resourceName, "scaling_configuration.0.timeout_action", "RollbackCapacityChange"),
				),
			},
			{
				Config: testAccClusterConfig_scalingConfiguration(rName, true, 256, 8, 600, 86400, "ForceApplyCapacityChange"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterExists(ctx, resourceName, &dbCluster),
					resource.TestCheckResourceAttr(resourceName, "scaling_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "scaling_configuration.0.auto_pause", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "scaling_configuration.0.max_capacity", "256"),
					resource.TestCheckResourceAttr(resourceName, "scaling_configuration.0.min_capacity", "8"),
					resource.TestCheckResourceAttr(resourceName, "scaling_configuration.0.seconds_before_timeout", "600"),
					resource.TestCheckResourceAttr(resourceName, "scaling_configuration.0.seconds_until_auto_pause", "86400"),
					resource.TestCheckResourceAttr(resourceName, "scaling_configuration.0.timeout_action", "ForceApplyCapacityChange"),
				),
			},
		},
	})
}

func TestAccRDSCluster_serverlessV2ScalingConfiguration(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var dbCluster rds.DBCluster

	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_rds_cluster.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.RDSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckClusterDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccClusterConfig_serverlessV2ScalingConfiguration(rName, 64.0, 0.5),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterExists(ctx, resourceName, &dbCluster),
					resource.TestCheckResourceAttr(resourceName, "serverlessv2_scaling_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "serverlessv2_scaling_configuration.0.max_capacity", "64"),
					resource.TestCheckResourceAttr(resourceName, "serverlessv2_scaling_configuration.0.min_capacity", "0.5"),
				),
			},
			{
				Config: testAccClusterConfig_serverlessV2ScalingConfiguration(rName, 128.0, 8.5),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterExists(ctx, resourceName, &dbCluster),
					resource.TestCheckResourceAttr(resourceName, "serverlessv2_scaling_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "serverlessv2_scaling_configuration.0.max_capacity", "128"),
					resource.TestCheckResourceAttr(resourceName, "serverlessv2_scaling_configuration.0.min_capacity", "8.5"),
				),
			},
		},
	})
}

// Reference: https://github.com/hashicorp/terraform-provider-aws/issues/11698
func TestAccRDSCluster_Scaling_defaultMinCapacity(t *testing.T) {
	ctx := acctest.Context(t)
	var dbCluster rds.DBCluster

	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_rds_cluster.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.RDSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckClusterDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccClusterConfig_ScalingConfiguration_defaultMinCapacity(rName, false, 128, 301, "RollbackCapacityChange"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterExists(ctx, resourceName, &dbCluster),
					resource.TestCheckResourceAttr(resourceName, "scaling_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "scaling_configuration.0.auto_pause", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "scaling_configuration.0.max_capacity", "128"),
					resource.TestCheckResourceAttr(resourceName, "scaling_configuration.0.min_capacity", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "scaling_configuration.0.seconds_until_auto_pause", "301"),
					resource.TestCheckResourceAttr(resourceName, "scaling_configuration.0.timeout_action", "RollbackCapacityChange"),
				),
			},
		},
	})
}

func TestAccRDSCluster_snapshotIdentifier(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var dbCluster, sourceDbCluster rds.DBCluster
	var dbClusterSnapshot types.DBClusterSnapshot

	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	sourceDbResourceName := "aws_rds_cluster.source"
	snapshotResourceName := "aws_db_cluster_snapshot.test"
	resourceName := "aws_rds_cluster.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.RDSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDBInstanceDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccClusterConfig_snapshotID(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterExists(ctx, sourceDbResourceName, &sourceDbCluster),
					testAccCheckClusterSnapshotExists(ctx, snapshotResourceName, &dbClusterSnapshot),
					testAccCheckClusterExists(ctx, resourceName, &dbCluster),
				),
			},
		},
	})
}

func TestAccRDSCluster_SnapshotIdentifier_deletionProtection(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var dbCluster, sourceDbCluster rds.DBCluster
	var dbClusterSnapshot types.DBClusterSnapshot

	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	sourceDbResourceName := "aws_rds_cluster.source"
	snapshotResourceName := "aws_db_cluster_snapshot.test"
	resourceName := "aws_rds_cluster.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.RDSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDBInstanceDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccClusterConfig_SnapshotID_deletionProtection(rName, true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterExists(ctx, sourceDbResourceName, &sourceDbCluster),
					testAccCheckClusterSnapshotExists(ctx, snapshotResourceName, &dbClusterSnapshot),
					testAccCheckClusterExists(ctx, resourceName, &dbCluster),
					resource.TestCheckResourceAttr(resourceName, names.AttrDeletionProtection, acctest.CtTrue),
				),
			},
			// Ensure we disable deletion protection before attempting to delete :)
			{
				Config: testAccClusterConfig_SnapshotID_deletionProtection(rName, false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterExists(ctx, sourceDbResourceName, &sourceDbCluster),
					testAccCheckClusterSnapshotExists(ctx, snapshotResourceName, &dbClusterSnapshot),
					testAccCheckClusterExists(ctx, resourceName, &dbCluster),
					resource.TestCheckResourceAttr(resourceName, names.AttrDeletionProtection, acctest.CtFalse),
				),
			},
		},
	})
}

func TestAccRDSCluster_SnapshotIdentifierEngineMode_provisioned(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var dbCluster, sourceDbCluster rds.DBCluster
	var dbClusterSnapshot types.DBClusterSnapshot

	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	sourceDbResourceName := "aws_rds_cluster.source"
	snapshotResourceName := "aws_db_cluster_snapshot.test"
	resourceName := "aws_rds_cluster.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.RDSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDBInstanceDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccClusterConfig_SnapshotID_engineMode(rName, "provisioned"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterExists(ctx, sourceDbResourceName, &sourceDbCluster),
					testAccCheckClusterSnapshotExists(ctx, snapshotResourceName, &dbClusterSnapshot),
					testAccCheckClusterExists(ctx, resourceName, &dbCluster),
					resource.TestCheckResourceAttr(resourceName, "engine_mode", "provisioned"),
				),
			},
		},
	})
}

func TestAccRDSCluster_SnapshotIdentifierEngineMode_serverless(t *testing.T) {
	// The below is according to AWS Support. This test can be updated in the future
	// to initialize some data.
	t.Skip("serverless does not support snapshot restore on an empty volume")

	ctx := acctest.Context(t)

	var dbCluster, sourceDbCluster rds.DBCluster
	var dbClusterSnapshot types.DBClusterSnapshot

	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	sourceDbResourceName := "aws_rds_cluster.source"
	snapshotResourceName := "aws_db_cluster_snapshot.test"
	resourceName := "aws_rds_cluster.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.RDSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDBInstanceDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccClusterConfig_SnapshotID_engineMode(rName, "serverless"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterExists(ctx, sourceDbResourceName, &sourceDbCluster),
					testAccCheckClusterSnapshotExists(ctx, snapshotResourceName, &dbClusterSnapshot),
					testAccCheckClusterExists(ctx, resourceName, &dbCluster),
					resource.TestCheckResourceAttr(resourceName, "engine_mode", "serverless"),
				),
			},
		},
	})
}

// Reference: https://github.com/hashicorp/terraform-provider-aws/issues/6157
func TestAccRDSCluster_SnapshotIdentifierEngineVersion_different(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var dbCluster, sourceDbCluster rds.DBCluster
	var dbClusterSnapshot types.DBClusterSnapshot

	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	sourceDbResourceName := "aws_rds_cluster.source"
	snapshotResourceName := "aws_db_cluster_snapshot.test"
	resourceName := "aws_rds_cluster.test"
	dataSourceName := "data.aws_rds_engine_version.upgrade"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.RDSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDBInstanceDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccClusterConfig_SnapshotID_engineVersion(rName, false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterExists(ctx, sourceDbResourceName, &sourceDbCluster),
					testAccCheckClusterSnapshotExists(ctx, snapshotResourceName, &dbClusterSnapshot),
					testAccCheckClusterExists(ctx, resourceName, &dbCluster),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrEngineVersion, dataSourceName, names.AttrVersion),
				),
			},
		},
	})
}

// Reference: https://github.com/hashicorp/terraform-provider-aws/issues/6157
func TestAccRDSCluster_SnapshotIdentifierEngineVersion_equal(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var dbCluster, sourceDbCluster rds.DBCluster
	var dbClusterSnapshot types.DBClusterSnapshot

	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	sourceDbResourceName := "aws_rds_cluster.source"
	snapshotResourceName := "aws_db_cluster_snapshot.test"
	resourceName := "aws_rds_cluster.test"
	dataSourceName := "data.aws_rds_engine_version.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.RDSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDBInstanceDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccClusterConfig_SnapshotID_engineVersion(rName, true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterExists(ctx, sourceDbResourceName, &sourceDbCluster),
					testAccCheckClusterSnapshotExists(ctx, snapshotResourceName, &dbClusterSnapshot),
					testAccCheckClusterExists(ctx, resourceName, &dbCluster),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrEngineVersion, dataSourceName, names.AttrVersion),
				),
			},
		},
	})
}

func TestAccRDSCluster_SnapshotIdentifier_kmsKeyID(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var dbCluster, sourceDbCluster rds.DBCluster
	var dbClusterSnapshot types.DBClusterSnapshot

	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	kmsKeyResourceName := "aws_kms_key.test"
	sourceDbResourceName := "aws_rds_cluster.source"
	snapshotResourceName := "aws_db_cluster_snapshot.test"
	resourceName := "aws_rds_cluster.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.RDSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDBInstanceDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccClusterConfig_SnapshotID_kmsKeyID(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterExists(ctx, sourceDbResourceName, &sourceDbCluster),
					testAccCheckClusterSnapshotExists(ctx, snapshotResourceName, &dbClusterSnapshot),
					testAccCheckClusterExists(ctx, resourceName, &dbCluster),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrKMSKeyID, kmsKeyResourceName, names.AttrARN),
				),
			},
		},
	})
}

func TestAccRDSCluster_SnapshotIdentifier_masterPassword(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var dbCluster, sourceDbCluster rds.DBCluster
	var dbClusterSnapshot types.DBClusterSnapshot

	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	sourceDbResourceName := "aws_rds_cluster.source"
	snapshotResourceName := "aws_db_cluster_snapshot.test"
	resourceName := "aws_rds_cluster.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.RDSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDBInstanceDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccClusterConfig_SnapshotID_masterPassword(rName, "password1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterExists(ctx, sourceDbResourceName, &sourceDbCluster),
					testAccCheckClusterSnapshotExists(ctx, snapshotResourceName, &dbClusterSnapshot),
					testAccCheckClusterExists(ctx, resourceName, &dbCluster),
					resource.TestCheckResourceAttr(resourceName, "master_password", "password1"),
				),
			},
		},
	})
}

func TestAccRDSCluster_SnapshotIdentifier_masterUsername(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var dbCluster, sourceDbCluster rds.DBCluster
	var dbClusterSnapshot types.DBClusterSnapshot

	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	sourceDbResourceName := "aws_rds_cluster.source"
	snapshotResourceName := "aws_db_cluster_snapshot.test"
	resourceName := "aws_rds_cluster.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.RDSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDBInstanceDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccClusterConfig_SnapshotID_masterUsername(rName, "username1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterExists(ctx, sourceDbResourceName, &sourceDbCluster),
					testAccCheckClusterSnapshotExists(ctx, snapshotResourceName, &dbClusterSnapshot),
					testAccCheckClusterExists(ctx, resourceName, &dbCluster),
					resource.TestCheckResourceAttr(resourceName, "master_username", "foo"),
				),
				// It is not currently possible to update the master username in the RDS API
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccRDSCluster_SnapshotIdentifier_preferredBackupWindow(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var dbCluster, sourceDbCluster rds.DBCluster
	var dbClusterSnapshot types.DBClusterSnapshot

	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	sourceDbResourceName := "aws_rds_cluster.source"
	snapshotResourceName := "aws_db_cluster_snapshot.test"
	resourceName := "aws_rds_cluster.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.RDSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDBInstanceDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccClusterConfig_SnapshotID_preferredBackupWindow(rName, "00:00-08:00"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterExists(ctx, sourceDbResourceName, &sourceDbCluster),
					testAccCheckClusterSnapshotExists(ctx, snapshotResourceName, &dbClusterSnapshot),
					testAccCheckClusterExists(ctx, resourceName, &dbCluster),
					resource.TestCheckResourceAttr(resourceName, "preferred_backup_window", "00:00-08:00"),
				),
			},
		},
	})
}

func TestAccRDSCluster_SnapshotIdentifier_preferredMaintenanceWindow(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var dbCluster, sourceDbCluster rds.DBCluster
	var dbClusterSnapshot types.DBClusterSnapshot

	// This config is version agnostic. Use it as a model for fixing version errors
	// in other tests.

	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	sourceDbResourceName := "aws_rds_cluster.source"
	snapshotResourceName := "aws_db_cluster_snapshot.test"
	resourceName := "aws_rds_cluster.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.RDSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDBInstanceDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccClusterConfig_SnapshotID_preferredMaintenanceWindow(rName, "sun:01:00-sun:01:30"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterExists(ctx, sourceDbResourceName, &sourceDbCluster),
					testAccCheckClusterSnapshotExists(ctx, snapshotResourceName, &dbClusterSnapshot),
					testAccCheckClusterExists(ctx, resourceName, &dbCluster),
					resource.TestCheckResourceAttr(resourceName, names.AttrPreferredMaintenanceWindow, "sun:01:00-sun:01:30"),
				),
			},
		},
	})
}

func TestAccRDSCluster_SnapshotIdentifier_tags(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var dbCluster, sourceDbCluster rds.DBCluster
	var dbClusterSnapshot types.DBClusterSnapshot

	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	sourceDbResourceName := "aws_rds_cluster.source"
	snapshotResourceName := "aws_db_cluster_snapshot.test"
	resourceName := "aws_rds_cluster.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.RDSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDBInstanceDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccClusterConfig_SnapshotID_tags(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterExists(ctx, sourceDbResourceName, &sourceDbCluster),
					testAccCheckClusterSnapshotExists(ctx, snapshotResourceName, &dbClusterSnapshot),
					testAccCheckClusterExists(ctx, resourceName, &dbCluster),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey1, acctest.CtValue1),
				),
			},
		},
	})
}

func TestAccRDSCluster_SnapshotIdentifier_vpcSecurityGroupIDs(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var dbCluster, sourceDbCluster rds.DBCluster
	var dbClusterSnapshot types.DBClusterSnapshot

	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	sourceDbResourceName := "aws_rds_cluster.source"
	snapshotResourceName := "aws_db_cluster_snapshot.test"
	resourceName := "aws_rds_cluster.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.RDSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDBInstanceDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccClusterConfig_SnapshotID_vpcSecurityGroupIDs(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterExists(ctx, sourceDbResourceName, &sourceDbCluster),
					testAccCheckClusterSnapshotExists(ctx, snapshotResourceName, &dbClusterSnapshot),
					testAccCheckClusterExists(ctx, resourceName, &dbCluster),
				),
			},
		},
	})
}

// Regression reference: https://github.com/hashicorp/terraform-provider-aws/issues/5450
// This acceptance test explicitly tests when snapshot_identifier is set,
// vpc_security_group_ids is set (which triggered the resource update function),
// and tags is set which was missing its ARN used for tagging
func TestAccRDSCluster_SnapshotIdentifierVPCSecurityGroupIDs_tags(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var dbCluster, sourceDbCluster rds.DBCluster
	var dbClusterSnapshot types.DBClusterSnapshot

	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	sourceDbResourceName := "aws_rds_cluster.source"
	snapshotResourceName := "aws_db_cluster_snapshot.test"
	resourceName := "aws_rds_cluster.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.RDSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDBInstanceDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccClusterConfig_SnapshotID_VPCSecurityGroupIds_tags(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterExists(ctx, sourceDbResourceName, &sourceDbCluster),
					testAccCheckClusterSnapshotExists(ctx, snapshotResourceName, &dbClusterSnapshot),
					testAccCheckClusterExists(ctx, resourceName, &dbCluster),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey1, acctest.CtValue1),
				),
			},
		},
	})
}

func TestAccRDSCluster_SnapshotIdentifier_encryptedRestore(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var dbCluster, sourceDbCluster rds.DBCluster
	var dbClusterSnapshot types.DBClusterSnapshot

	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	kmsKeyResourceName := "aws_kms_key.test"
	sourceDbResourceName := "aws_rds_cluster.source"
	snapshotResourceName := "aws_db_cluster_snapshot.test"
	resourceName := "aws_rds_cluster.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.RDSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDBInstanceDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccClusterConfig_SnapshotID_encryptedRestore(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterExists(ctx, sourceDbResourceName, &sourceDbCluster),
					testAccCheckClusterSnapshotExists(ctx, snapshotResourceName, &dbClusterSnapshot),
					testAccCheckClusterExists(ctx, resourceName, &dbCluster),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrKMSKeyID, kmsKeyResourceName, names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, names.AttrStorageEncrypted, acctest.CtTrue),
				),
			},
		},
	})
}

func TestAccRDSCluster_enableHTTPEndpoint(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var dbCluster rds.DBCluster

	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_rds_cluster.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.RDSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckClusterDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccClusterConfig_enableHTTPEndpoint(rName, true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterExists(ctx, resourceName, &dbCluster),
					resource.TestCheckResourceAttr(resourceName, "enable_http_endpoint", acctest.CtTrue),
				),
			},
			{
				Config: testAccClusterConfig_enableHTTPEndpoint(rName, false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterExists(ctx, resourceName, &dbCluster),
					resource.TestCheckResourceAttr(resourceName, "enable_http_endpoint", acctest.CtFalse),
				),
			},
		},
	})
}

func TestAccRDSCluster_enableHTTPEndpointProvisioned(t *testing.T) {
	ctx := acctest.Context(t)

	var dbCluster rds.DBCluster

	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_rds_cluster.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.RDSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckClusterDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccClusterConfig_enableHTTPEndpointProvisioned(rName, true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterExists(ctx, resourceName, &dbCluster),
					resource.TestCheckResourceAttr(resourceName, "enable_http_endpoint", acctest.CtTrue),
				),
			},
		},
	})
}

func TestAccRDSCluster_password(t *testing.T) {
	ctx := acctest.Context(t)
	var dbCluster rds.DBCluster
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_rds_cluster.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.RDSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckClusterDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccClusterConfig_password(rName, "valid-password-1"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckClusterExists(ctx, resourceName, &dbCluster),
					resource.TestCheckResourceAttr(resourceName, "master_password", "valid-password-1"),
				),
			},
			{
				Config: testAccClusterConfig_password(rName, "valid-password-2"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckClusterExists(ctx, resourceName, &dbCluster),
					resource.TestCheckResourceAttr(resourceName, "master_password", "valid-password-2"),
				),
			},
		},
	})
}

func TestAccRDSCluster_NoDeleteAutomatedBackups(t *testing.T) {
	ctx := acctest.Context(t)
	var dbCluster rds.DBCluster
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_rds_cluster.test"

	backupWindowLength := 30 * time.Minute
	clusterProvisionTime := 5 * time.Minute

	// Unlike rds instance, there is no automated backup created when the cluster is deployed.
	// However, we can forcefully create one by moving maintenance window behind the deployment.
	now := time.Now().UTC()
	backupWindowStart := now.Add(clusterProvisionTime)
	backupWindowEnd := backupWindowStart.Add(backupWindowLength)
	preferredBackupWindow := fmt.Sprintf(
		"%02d:%02d-%02d:%02d",
		backupWindowStart.Hour(),
		backupWindowStart.Minute(),
		backupWindowEnd.Hour(),
		backupWindowEnd.Minute(),
	)

	waitUntilAutomatedBackupCreated := func(*terraform.State) error {
		ticker := time.NewTicker(1 * time.Minute)
		for {
			select {
			case <-time.After(backupWindowLength):
				return nil
			case <-ticker.C:
				conn := acctest.Provider.Meta().(*conns.AWSClient).RDSConn(ctx)
				output, _ := conn.DescribeDBClusterSnapshotsWithContext(ctx, &rds.DescribeDBClusterSnapshotsInput{
					DBClusterIdentifier: aws.String(rName),
					SnapshotType:        aws.String("automated"),
				})
				if output != nil && len(output.DBClusterSnapshots) > 0 {
					snapshot := output.DBClusterSnapshots[0]
					if aws.StringValue(snapshot.Status) == "available" {
						return nil
					}
				}
			}
		}
	}

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.RDSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckClusterAutomatedBackupsDelete(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccClusterConfig_noDeleteAutomatedBackups(rName, preferredBackupWindow),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterExists(ctx, resourceName, &dbCluster),
					waitUntilAutomatedBackupCreated,
				),
			},
		},
	})
}

func TestAccRDSCluster_engineLifecycleSupport_disabled(t *testing.T) {
	ctx := acctest.Context(t)
	var dbCluster rds.DBCluster
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_rds_cluster.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.RDSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckClusterDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccClusterConfig_engineLifecycleSupport_disabled(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckClusterExists(ctx, resourceName, &dbCluster),
					acctest.CheckResourceAttrRegionalARN(resourceName, names.AttrARN, "rds", fmt.Sprintf("cluster:%s", rName)),
					resource.TestCheckResourceAttr(resourceName, "backtrack_window", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, names.AttrClusterIdentifier, rName),
					resource.TestCheckResourceAttr(resourceName, names.AttrEngine, tfrds.ClusterEngineAuroraPostgreSQL),
					resource.TestCheckResourceAttr(resourceName, "engine_lifecycle_support", "open-source-rds-extended-support-disabled"),
				),
			},
			testAccClusterImportStep(resourceName),
		},
	})
}

func testAccCheckClusterDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		return testAccCheckClusterDestroyWithProvider(ctx)(s, acctest.Provider)
	}
}

func testAccCheckClusterAutomatedBackupsDelete(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).RDSConn(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_rds_cluster" {
				continue
			}
			describeOutput, err := conn.DescribeDBClusterSnapshotsWithContext(ctx, &rds.DescribeDBClusterSnapshotsInput{
				DBClusterIdentifier: aws.String(rs.Primary.Attributes[names.AttrClusterIdentifier]),
				SnapshotType:        aws.String("automated"),
			})
			if err != nil {
				return err
			}

			if describeOutput == nil || len(describeOutput.DBClusterSnapshots) == 0 {
				return fmt.Errorf("Automated backup for %s not found", rs.Primary.Attributes[names.AttrClusterIdentifier])
			}

			_, err = conn.DeleteDBClusterAutomatedBackupWithContext(ctx, &rds.DeleteDBClusterAutomatedBackupInput{
				DbClusterResourceId: aws.String(rs.Primary.Attributes["cluster_resource_id"]),
			})
			if err != nil {
				return err
			}
		}

		return testAccCheckClusterDestroyWithProvider(ctx)(s, acctest.Provider)
	}
}

func testAccCheckClusterDestroyWithProvider(ctx context.Context) acctest.TestCheckWithProviderFunc {
	return func(s *terraform.State, provider *schema.Provider) error {
		conn := provider.Meta().(*conns.AWSClient).RDSConn(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_rds_cluster" {
				continue
			}

			_, err := tfrds.FindDBClusterByID(ctx, conn, rs.Primary.ID)

			if tfresource.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("RDS Cluster %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccCheckClusterDestroyWithFinalSnapshot(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn1 := acctest.Provider.Meta().(*conns.AWSClient).RDSConn(ctx)
		conn2 := acctest.Provider.Meta().(*conns.AWSClient).RDSClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_rds_cluster" {
				continue
			}

			finalSnapshotID := rs.Primary.Attributes[names.AttrFinalSnapshotIdentifier]
			_, err := tfrds.FindDBClusterSnapshotByID(ctx, conn2, finalSnapshotID)
			if err != nil {
				return err
			}

			_, err = conn2.DeleteDBClusterSnapshot(ctx, &rds_sdkv2.DeleteDBClusterSnapshotInput{
				DBClusterSnapshotIdentifier: aws.String(finalSnapshotID),
			})

			if err != nil {
				return err
			}

			_, err = tfrds.FindDBClusterByID(ctx, conn1, rs.Primary.ID)

			if tfresource.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("RDS Cluster %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccCheckClusterExists(ctx context.Context, n string, v *rds.DBCluster) resource.TestCheckFunc {
	return testAccCheckClusterExistsWithProvider(ctx, n, v, func() *schema.Provider { return acctest.Provider })
}

func testAccCheckClusterExistsWithProvider(ctx context.Context, n string, v *rds.DBCluster, providerF func() *schema.Provider) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No RDS Cluster ID is set")
		}

		conn := providerF().Meta().(*conns.AWSClient).RDSConn(ctx)

		output, err := tfrds.FindDBClusterByID(ctx, conn, rs.Primary.ID)
		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

func testAccCheckClusterRecreated(i, j *rds.DBCluster) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if aws.TimeValue(i.ClusterCreateTime).Equal(aws.TimeValue(j.ClusterCreateTime)) {
			return errors.New("RDS Cluster was not recreated")
		}

		return nil
	}
}

func testAccCheckClusterNotRecreated(i, j *rds.DBCluster) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if !aws.TimeValue(i.ClusterCreateTime).Equal(aws.TimeValue(j.ClusterCreateTime)) {
			return errors.New("RDS Cluster was recreated")
		}

		return nil
	}
}

func TestAccRDSCluster_localWriteForwarding(t *testing.T) {
	ctx := acctest.Context(t)
	var dbCluster rds.DBCluster
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_rds_cluster.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.RDSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckClusterDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccClusterConfig_localWriteForwarding(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckClusterExists(ctx, resourceName, &dbCluster),
					acctest.CheckResourceAttrRegionalARN(resourceName, names.AttrARN, "rds", fmt.Sprintf("cluster:%s", rName)),
					resource.TestCheckResourceAttr(resourceName, "enable_local_write_forwarding", acctest.CtTrue),
				),
			},
			testAccClusterImportStep(resourceName),
		},
	})
}

func testAccClusterConfig_basic(rName string) string {
	return fmt.Sprintf(`
resource "aws_rds_cluster" "test" {
  cluster_identifier  = %[1]q
  database_name       = "test"
  engine              = %[2]q
  master_username     = "tfacctest"
  master_password     = "avoid-plaintext-passwords"
  skip_final_snapshot = true
}
`, rName, tfrds.ClusterEngineAuroraMySQL)
}

func testAccClusterConfig_tags1(rName, tagKey1, tagValue1 string) string {
	return fmt.Sprintf(`
resource "aws_rds_cluster" "test" {
  cluster_identifier  = %[1]q
  database_name       = "test"
  master_username     = "tfacctest"
  master_password     = "avoid-plaintext-passwords"
  engine              = %[2]q
  skip_final_snapshot = true

  tags = {
    %[3]q = %[4]q
  }
}
`, rName, tfrds.ClusterEngineAuroraMySQL, tagKey1, tagValue1)
}

func testAccClusterConfig_tags2(rName, tagKey1, tagValue1, tagKey2, tagValue2 string) string {
	return fmt.Sprintf(`
resource "aws_rds_cluster" "test" {
  cluster_identifier  = %[1]q
  database_name       = "test"
  master_username     = "tfacctest"
  master_password     = "avoid-plaintext-passwords"
  engine              = %[2]q
  skip_final_snapshot = true

  tags = {
    %[3]q = %[4]q
    %[5]q = %[6]q
  }
}
`, rName, tfrds.ClusterEngineAuroraMySQL, tagKey1, tagValue1, tagKey2, tagValue2)
}

func testAccClusterConfig_identifierGenerated() string {
	return fmt.Sprintf(`
resource "aws_rds_cluster" "test" {
  engine              = %[1]q
  master_username     = "tfacctest"
  master_password     = "avoid-plaintext-passwords"
  skip_final_snapshot = true
}
`, tfrds.ClusterEngineAuroraMySQL)
}

func testAccClusterConfig_identifierPrefix(identifierPrefix string) string {
	return fmt.Sprintf(`
resource "aws_rds_cluster" "test" {
  cluster_identifier_prefix = %[1]q
  engine                    = %[2]q
  master_username           = "tfacctest"
  master_password           = "avoid-plaintext-passwords"
  skip_final_snapshot       = true
}
`, identifierPrefix, tfrds.ClusterEngineAuroraMySQL)
}

func testAccClusterConfig_managedMasterPassword(rName string) string {
	return fmt.Sprintf(`
resource "aws_rds_cluster" "test" {
  cluster_identifier          = %[1]q
  database_name               = "test"
  manage_master_user_password = true
  master_username             = "tfacctest"
  engine                      = %[2]q
  skip_final_snapshot         = true
}
`, rName, tfrds.ClusterEngineAuroraMySQL)
}

func testAccClusterConfig_managedMasterPasswordKMSKey(rName string) string {
	return fmt.Sprintf(`
data "aws_caller_identity" "current" {}
data "aws_partition" "current" {}

resource "aws_kms_key" "example" {
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
        "AWS": "arn:${data.aws_partition.current.partition}:iam::${data.aws_caller_identity.current.account_id}:root"
      },
      "Action": "kms:*",
      "Resource": "*"
    }
  ]
}
 POLICY

}

resource "aws_rds_cluster" "test" {
  cluster_identifier            = %[1]q
  database_name                 = "test"
  engine                        = %[2]q
  manage_master_user_password   = true
  master_username               = "tfacctest"
  master_user_secret_kms_key_id = aws_kms_key.example.arn
  skip_final_snapshot           = true
}
`, rName, tfrds.ClusterEngineAuroraMySQL)
}

func testAccClusterConfig_allowMajorVersionUpgrade(rName string, engine string, upgrade, applyImmediately bool) string {
	return fmt.Sprintf(`
data "aws_rds_engine_version" "test" {
  engine                  = data.aws_rds_engine_version.upgrade.engine
  latest                  = true
  preferred_major_targets = [data.aws_rds_engine_version.upgrade.version_actual]
}

data "aws_rds_engine_version" "upgrade" {
  engine = %[1]q
}

locals {
  engine_version = %[2]t ? data.aws_rds_engine_version.upgrade.version_actual : data.aws_rds_engine_version.test.version_actual
}

data "aws_rds_orderable_db_instance" "test" {
  engine                     = data.aws_rds_engine_version.test.engine
  engine_version             = local.engine_version
  preferred_instance_classes = [%[3]s]
}

resource "aws_rds_cluster" "test" {
  allow_major_version_upgrade = true
  apply_immediately           = %[4]t
  cluster_identifier          = %[5]q
  engine                      = data.aws_rds_engine_version.test.engine
  engine_version              = local.engine_version
  master_password             = "avoid-plaintext-passwords"
  master_username             = "tfacctest"
  skip_final_snapshot         = true
}

# Upgrading requires a healthy primary instance
resource "aws_rds_cluster_instance" "test" {
  apply_immediately  = %[4]t
  cluster_identifier = aws_rds_cluster.test.id
  engine             = data.aws_rds_engine_version.test.engine
  engine_version     = data.aws_rds_orderable_db_instance.test.engine_version
  identifier         = %[5]q
  instance_class     = data.aws_rds_orderable_db_instance.test.instance_class

  lifecycle {
    ignore_changes = [engine_version]
  }
}
`, engine, upgrade, mainInstanceClasses, applyImmediately, rName)
}

func testAccClusterConfig_allowMajorVersionUpgradeCustomParameters(rName string, upgrade bool, engine string, applyImmediate bool) string {
	return fmt.Sprintf(`
data "aws_rds_engine_version" "test" {
  engine                  = data.aws_rds_engine_version.upgrade.engine
  latest                  = true
  preferred_major_targets = [data.aws_rds_engine_version.upgrade.version_actual]
}

data "aws_rds_engine_version" "upgrade" {
  engine = %[1]q
}

locals {
  parameter_group_family = %[2]t ? data.aws_rds_engine_version.upgrade.parameter_group_family : data.aws_rds_engine_version.test.parameter_group_family
  engine_version         = %[2]t ? data.aws_rds_engine_version.upgrade.version_actual : data.aws_rds_engine_version.test.version_actual
}

data "aws_rds_orderable_db_instance" "test" {
  engine                     = data.aws_rds_engine_version.test.engine
  engine_version             = local.engine_version
  preferred_instance_classes = [%[3]s]
}

resource "aws_rds_cluster" "test" {
  allow_major_version_upgrade      = true
  apply_immediately                = true
  cluster_identifier               = %[4]q
  db_cluster_parameter_group_name  = aws_rds_cluster_parameter_group.test.name
  db_instance_parameter_group_name = aws_db_parameter_group.test.name
  engine                           = data.aws_rds_engine_version.test.engine
  engine_version                   = local.engine_version
  master_password                  = "mustbeeightcharaters"
  master_username                  = "test"
  skip_final_snapshot              = true
}

# Upgrading requires a healthy primary instance
resource "aws_rds_cluster_instance" "test" {
  apply_immediately       = %[5]t
  cluster_identifier      = aws_rds_cluster.test.id
  db_parameter_group_name = aws_db_parameter_group.test.name
  engine                  = data.aws_rds_orderable_db_instance.test.engine
  engine_version          = data.aws_rds_orderable_db_instance.test.engine_version
  identifier              = %[4]q
  instance_class          = data.aws_rds_orderable_db_instance.test.instance_class

  lifecycle {
    ignore_changes = [engine_version]
  }
}

resource "aws_rds_cluster_parameter_group" "test" {
  name_prefix = %[4]q
  family      = local.parameter_group_family

  lifecycle {
    create_before_destroy = true
  }
}

resource "aws_db_parameter_group" "test" {
  name_prefix = %[4]q
  family      = local.parameter_group_family

  lifecycle {
    create_before_destroy = true
  }
}
`, engine, upgrade, mainInstanceClasses, rName, applyImmediate)
}

func testAccClusterConfig_majorVersionOnly(rName string, engine string, allowMajorVersionUpgrade bool) string {
	return fmt.Sprintf(`
data "aws_rds_engine_version" "test" {
  engine                  = data.aws_rds_engine_version.upgrade.engine
  latest                  = true
  preferred_major_targets = [data.aws_rds_engine_version.upgrade.version_actual]
}

data "aws_rds_engine_version" "upgrade" {
  engine = %[1]q
}

data "aws_rds_orderable_db_instance" "test" {
  engine                     = data.aws_rds_engine_version.test.engine
  engine_version             = data.aws_rds_engine_version.test.version_actual
  preferred_instance_classes = [%[2]s]
}

resource "aws_rds_cluster" "test" {
  allow_major_version_upgrade = %[3]t
  apply_immediately           = true
  cluster_identifier          = %[4]q
  engine                      = data.aws_rds_engine_version.test.engine
  engine_version              = data.aws_rds_engine_version.test.version_actual
  master_password             = "mustbeeightcharaters"
  master_username             = "test"
  skip_final_snapshot         = true
}

# Upgrading requires a healthy primary instance
resource "aws_rds_cluster_instance" "test" {
  cluster_identifier = aws_rds_cluster.test.id
  engine             = aws_rds_cluster.test.engine
  engine_version     = aws_rds_cluster.test.engine_version
  identifier         = %[4]q
  instance_class     = data.aws_rds_orderable_db_instance.test.instance_class
}
`, engine, mainInstanceClasses, allowMajorVersionUpgrade, rName)
}

func testAccClusterConfig_minorVersion(rName, engine string, upgrade bool) string {
	return fmt.Sprintf(`
data "aws_rds_engine_version" "test" {
  engine                    = %[1]q
  latest                    = true
  preferred_upgrade_targets = [data.aws_rds_engine_version.upgrade.version_actual]
}

data "aws_rds_engine_version" "upgrade" {
  engine = %[1]q
}

locals {
  parameter_group_name = %[2]t ? data.aws_rds_engine_version.upgrade.parameter_group_family : data.aws_rds_engine_version.test.parameter_group_family
  engine_version       = %[2]t ? data.aws_rds_engine_version.upgrade.version_actual : data.aws_rds_engine_version.test.version_actual
}

data "aws_rds_orderable_db_instance" "test" {
  engine                     = data.aws_rds_engine_version.test.engine
  engine_version             = local.engine_version
  preferred_instance_classes = [%[3]s]
}

resource "aws_rds_cluster" "test" {
  cluster_identifier               = %[4]q
  engine                           = data.aws_rds_engine_version.test.engine
  engine_version                   = local.engine_version
  db_instance_parameter_group_name = "default.${local.parameter_group_name}"
  master_username                  = "tfacctest"
  master_password                  = %[4]q
  skip_final_snapshot              = true
}

# Upgrading requires a healthy primary instance
resource "aws_rds_cluster_instance" "test" {
  apply_immediately       = true
  cluster_identifier      = aws_rds_cluster.test.id
  db_parameter_group_name = aws_rds_cluster.test.db_instance_parameter_group_name
  engine                  = data.aws_rds_orderable_db_instance.test.engine
  engine_version          = data.aws_rds_orderable_db_instance.test.engine_version
  identifier              = %[4]q
  instance_class          = data.aws_rds_orderable_db_instance.test.instance_class

  lifecycle {
    ignore_changes = [engine_version]
  }
}
`, engine, upgrade, mainInstanceClasses, rName)
}

func testAccClusterConfig_availabilityZones(rName string) string {
	return acctest.ConfigCompose(acctest.ConfigAvailableAZsNoOptIn(), fmt.Sprintf(`
resource "aws_rds_cluster" "test" {
  apply_immediately   = true
  availability_zones  = [data.aws_availability_zones.available.names[0], data.aws_availability_zones.available.names[1], data.aws_availability_zones.available.names[2]]
  cluster_identifier  = %[1]q
  engine              = %[2]q
  master_password     = "avoid-plaintext-passwords"
  master_username     = "tfacctest"
  skip_final_snapshot = true
}
`, rName, tfrds.ClusterEngineAuroraMySQL))
}

func testAccClusterConfig_availabilityZones_caCertificateIdentifier(rName string) string {
	return acctest.ConfigCompose(testAccConfig_ClusterSubnetGroup(rName), fmt.Sprintf(`
data "aws_rds_orderable_db_instance" "test" {
  engine                     = %[2]q
  engine_latest_version      = true
  preferred_instance_classes = [%[3]s]
  supports_clusters          = true
  supports_iops              = true
}

resource "aws_rds_cluster" "test" {
  apply_immediately         = true
  db_subnet_group_name      = aws_db_subnet_group.test.name
  ca_certificate_identifier = "rds-ca-2019"
  cluster_identifier        = %[1]q
  engine                    = data.aws_rds_orderable_db_instance.test.engine
  engine_version            = data.aws_rds_orderable_db_instance.test.engine_version
  storage_type              = data.aws_rds_orderable_db_instance.test.storage_type
  allocated_storage         = 100
  iops                      = 1000
  db_cluster_instance_class = "db.r6gd.xlarge"
  master_password           = "avoid-plaintext-passwords"
  master_username           = "tfacctest"
  skip_final_snapshot       = true
}
`, rName, tfrds.ClusterEngineMySQL, mainInstanceClasses))
}

func testAccClusterConfig_storageType(rName string, sType string) string {
	return acctest.ConfigCompose(
		testAccConfig_ClusterSubnetGroup(rName),
		fmt.Sprintf(`
data "aws_rds_orderable_db_instance" "test" {
  engine                     = %[1]q
  engine_latest_version      = true
  preferred_instance_classes = [%[2]s]
  storage_type               = %[4]q
  supports_iops              = true
  supports_clusters          = true
}

resource "aws_rds_cluster" "test" {
  apply_immediately         = true
  cluster_identifier        = %[3]q
  db_cluster_instance_class = data.aws_rds_orderable_db_instance.test.instance_class
  db_subnet_group_name      = aws_db_subnet_group.test.name
  engine                    = data.aws_rds_orderable_db_instance.test.engine
  engine_version            = data.aws_rds_orderable_db_instance.test.engine_version
  storage_type              = data.aws_rds_orderable_db_instance.test.storage_type
  allocated_storage         = 100
  iops                      = 1000
  master_password           = "mustbeeightcharaters"
  master_username           = "test"
  skip_final_snapshot       = true
}
`, tfrds.ClusterEngineMySQL, mainInstanceClasses, rName, sType))
}

func testAccClusterConfig_allocatedStorage(rName string) string {
	return acctest.ConfigCompose(
		testAccConfig_ClusterSubnetGroup(rName),
		fmt.Sprintf(`
data "aws_rds_orderable_db_instance" "test" {
  engine                     = %[1]q
  engine_latest_version      = true
  preferred_instance_classes = [%[2]s]
  storage_type               = "io1"
  supports_iops              = true
  supports_clusters          = true
}

resource "aws_rds_cluster" "test" {
  apply_immediately         = true
  cluster_identifier        = %[3]q
  db_cluster_instance_class = data.aws_rds_orderable_db_instance.test.instance_class
  db_subnet_group_name      = aws_db_subnet_group.test.name
  engine                    = data.aws_rds_orderable_db_instance.test.engine
  engine_version            = data.aws_rds_orderable_db_instance.test.engine_version
  storage_type              = data.aws_rds_orderable_db_instance.test.storage_type
  allocated_storage         = 100
  iops                      = 1000
  master_password           = "mustbeeightcharaters"
  master_username           = "test"
  skip_final_snapshot       = true
}
`, tfrds.ClusterEngineMySQL, mainInstanceClasses, rName))
}

func testAccClusterConfig_iops(rName string) string {
	return acctest.ConfigCompose(
		testAccConfig_ClusterSubnetGroup(rName),
		fmt.Sprintf(`
data "aws_rds_orderable_db_instance" "test" {
  engine                     = %[1]q
  engine_latest_version      = true
  preferred_instance_classes = [%[2]s]
  storage_type               = "io1"
  supports_iops              = true
  supports_clusters          = true
}

resource "aws_rds_cluster" "test" {
  apply_immediately         = true
  cluster_identifier        = %[3]q
  db_cluster_instance_class = data.aws_rds_orderable_db_instance.test.instance_class
  db_subnet_group_name      = aws_db_subnet_group.test.name
  engine                    = data.aws_rds_orderable_db_instance.test.engine
  engine_version            = data.aws_rds_orderable_db_instance.test.engine_version
  storage_type              = data.aws_rds_orderable_db_instance.test.storage_type
  allocated_storage         = 100
  iops                      = 1000
  master_password           = "mustbeeightcharaters"
  master_username           = "test"
  skip_final_snapshot       = true
}
`, tfrds.ClusterEngineMySQL, mainInstanceClasses, rName))
}

func testAccClusterConfig_dbClusterInstanceClass(rName string, oddClasses bool) string {
	var halfClasses []string
	start := 0
	if oddClasses {
		start = 1
	}
	for i := start; i < len(instanceClassesSlice); i += 2 {
		halfClasses = append(halfClasses, instanceClassesSlice[i])
	}
	halfPreferredClasses := strings.Join(halfClasses, ", ")
	return acctest.ConfigCompose(
		testAccConfig_ClusterSubnetGroup(rName),
		fmt.Sprintf(`
data "aws_rds_orderable_db_instance" "test" {
  engine                     = %[1]q
  engine_latest_version      = true
  preferred_instance_classes = [%[2]s]
  storage_type               = "io1"
  supports_iops              = true
  supports_clusters          = true
}

# this test is about updating the instance class. not sure how to dynamically make it give a different value?
resource "aws_rds_cluster" "test" {
  apply_immediately         = true
  cluster_identifier        = %[3]q
  db_cluster_instance_class = data.aws_rds_orderable_db_instance.test.instance_class
  db_subnet_group_name      = aws_db_subnet_group.test.name
  engine                    = %[1]q
  storage_type              = "io1"
  allocated_storage         = 100
  iops                      = 1000
  master_password           = "mustbeeightcharaters"
  master_username           = "test"
  skip_final_snapshot       = true
}
`, tfrds.ClusterEngineMySQL, halfPreferredClasses, rName))
}

func testAccClusterConfig_backtrackWindow(backtrackWindow int) string {
	return fmt.Sprintf(`
resource "aws_rds_cluster" "test" {
  apply_immediately         = true
  backtrack_window          = %[1]d
  cluster_identifier_prefix = "tf-acc-test-"
  engine                    = %[2]q
  master_password           = "mustbeeightcharaters"
  master_username           = "test"
  skip_final_snapshot       = true
}
`, backtrackWindow, tfrds.ClusterEngineAuroraMySQL)
}

func testAccClusterConfig_subnetGroupName(rName string) string {
	return acctest.ConfigCompose(
		testAccConfig_ClusterSubnetGroup(rName),
		fmt.Sprintf(`
resource "aws_rds_cluster" "test" {
  cluster_identifier   = %[1]q
  engine               = %[2]q
  master_username      = "tfacctest"
  master_password      = "avoid-plaintext-passwords"
  db_subnet_group_name = aws_db_subnet_group.test.name
  skip_final_snapshot  = true
}
`, rName, tfrds.ClusterEngineAuroraMySQL))
}

func testAccClusterConfig_finalSnapshot(rName string) string {
	return fmt.Sprintf(`
resource "aws_rds_cluster" "test" {
  cluster_identifier        = %[1]q
  database_name             = "test"
  engine                    = %[2]q
  master_username           = "tfacctest"
  master_password           = "avoid-plaintext-passwords"
  final_snapshot_identifier = %[1]q
}
`, rName, tfrds.ClusterEngineAuroraMySQL)
}

func testAccClusterConfig_withoutUserNameAndPassword(n int) string {
	return fmt.Sprintf(`
resource "aws_rds_cluster" "default" {
  cluster_identifier  = "tf-aurora-cluster-%[1]d"
  engine              = %[2]q
  database_name       = "mydb"
  skip_final_snapshot = true
}
`, n, tfrds.ClusterEngineAuroraMySQL)
}

func testAccClusterConfig_baseForPITR(rName string) string {
	return acctest.ConfigCompose(
		testAccConfig_ClusterSubnetGroup(rName),
		fmt.Sprintf(`
resource "aws_rds_cluster" "test" {
  cluster_identifier   = %[1]q
  master_username      = "tfacctest"
  master_password      = "avoid-plaintext-passwords"
  db_subnet_group_name = aws_db_subnet_group.test.name
  skip_final_snapshot  = true
  engine               = %[2]q
}
`, rName, tfrds.ClusterEngineAuroraMySQL))
}

func testAccClusterConfig_pointInTimeRestoreSource(rName string) string {
	return acctest.ConfigCompose(testAccClusterConfig_baseForPITR(rName), fmt.Sprintf(`
resource "aws_rds_cluster" "restore" {
  cluster_identifier  = "%[1]s-restore"
  skip_final_snapshot = true
  engine              = aws_rds_cluster.test.engine

  restore_to_point_in_time {
    source_cluster_identifier  = aws_rds_cluster.test.cluster_identifier
    restore_type               = "full-copy"
    use_latest_restorable_time = true
  }
}
`, rName))
}

func testAccClusterConfig_pointInTimeRestoreSource_enabledCloudWatchLogsExports(rName, enabledCloudwatchLogExports string) string {
	return acctest.ConfigCompose(testAccClusterConfig_baseForPITR(rName), fmt.Sprintf(`
resource "aws_rds_cluster" "restore" {
  cluster_identifier              = "%[1]s-restore"
  skip_final_snapshot             = true
  engine                          = aws_rds_cluster.test.engine
  enabled_cloudwatch_logs_exports = [%[2]q]

  restore_to_point_in_time {
    source_cluster_identifier  = aws_rds_cluster.test.cluster_identifier
    restore_type               = "full-copy"
    use_latest_restorable_time = true
  }
}
`, rName, enabledCloudwatchLogExports))
}

func testAccClusterConfig_domain(rName, domain string) string {
	return acctest.ConfigCompose(
		testAccConfig_ClusterSubnetGroup(rName),
		testAccConfig_ServiceRole(rName),
		testAccConfig_DirectoryService(rName, domain),
		fmt.Sprintf(`
resource "aws_rds_cluster" "test" {
  cluster_identifier     = %[1]q
  engine                 = %[2]q
  master_username        = "tfacctest"
  master_password        = "avoid-plaintext-passwords"
  db_subnet_group_name   = aws_db_subnet_group.test.name
  skip_final_snapshot    = true
  vpc_security_group_ids = [aws_security_group.test.id]
  domain                 = aws_directory_service_directory.directory.id
  domain_iam_role_name   = aws_iam_role.role.name
}
`, rName, tfrds.ClusterEngineAuroraPostgreSQL),
	)
}

func testAccClusterConfig_enabledCloudWatchLogsExports1(rName, enabledCloudwatchLogExports1 string) string {
	return fmt.Sprintf(`
resource "aws_rds_cluster" "test" {
  cluster_identifier              = %[1]q
  enabled_cloudwatch_logs_exports = [%[2]q]
  engine                          = %[3]q
  master_username                 = "foo"
  master_password                 = "mustbeeightcharaters"
  skip_final_snapshot             = true
}
`, rName, enabledCloudwatchLogExports1, tfrds.ClusterEngineAuroraMySQL)
}

func testAccClusterConfig_enabledCloudWatchLogsExports2(rName, enabledCloudwatchLogExports1, enabledCloudwatchLogExports2 string) string {
	return fmt.Sprintf(`
resource "aws_rds_cluster" "test" {
  cluster_identifier              = %[1]q
  enabled_cloudwatch_logs_exports = [%[2]q, %[3]q]
  engine                          = %[4]q
  master_username                 = "foo"
  master_password                 = "mustbeeightcharaters"
  skip_final_snapshot             = true
}
`, rName, enabledCloudwatchLogExports1, enabledCloudwatchLogExports2, tfrds.ClusterEngineAuroraMySQL)
}

func testAccClusterConfig_enabledCloudWatchLogsExportsPostgreSQL1(rName, enabledCloudwatchLogExports1 string) string {
	return acctest.ConfigCompose(
		testAccConfig_ClusterSubnetGroup(rName),
		fmt.Sprintf(`
data "aws_rds_orderable_db_instance" "test" {
  engine                     = %[1]q
  engine_latest_version      = true
  preferred_instance_classes = [%[2]s]
  storage_type               = "io1"
  supports_iops              = true
  supports_clusters          = true
}

resource "aws_rds_cluster" "test" {
  cluster_identifier              = %[3]q
  enabled_cloudwatch_logs_exports = [%[4]q]
  master_username                 = "tfacctest"
  master_password                 = "avoid-plaintext-passwords"
  skip_final_snapshot             = true
  allocated_storage               = 100
  storage_type                    = data.aws_rds_orderable_db_instance.test.storage_type
  iops                            = 1000
  db_cluster_instance_class       = data.aws_rds_orderable_db_instance.test.instance_class
  db_subnet_group_name            = aws_db_subnet_group.test.name
  engine                          = data.aws_rds_orderable_db_instance.test.engine
  engine_mode                     = "provisioned"
  engine_version                  = data.aws_rds_orderable_db_instance.test.engine_version
}
`, tfrds.ClusterEnginePostgres, mainInstanceClasses, rName, enabledCloudwatchLogExports1))
}

func testAccClusterConfig_enabledCloudWatchLogsExportsPostgreSQL2(rName, enabledCloudwatchLogExports1, enabledCloudwatchLogExports2 string) string {
	return acctest.ConfigCompose(
		testAccConfig_ClusterSubnetGroup(rName),
		fmt.Sprintf(`
data "aws_rds_orderable_db_instance" "test" {
  engine                     = %[1]q
  engine_latest_version      = true
  preferred_instance_classes = [%[2]s]
  storage_type               = "io1"
  supports_iops              = true
  supports_clusters          = true
}

resource "aws_rds_cluster" "test" {
  cluster_identifier              = %[3]q
  enabled_cloudwatch_logs_exports = [%[4]q, %[5]q]
  master_username                 = "tfacctest"
  master_password                 = "avoid-plaintext-passwords"
  skip_final_snapshot             = true
  allocated_storage               = 100
  storage_type                    = data.aws_rds_orderable_db_instance.test.storage_type
  iops                            = 1000
  db_cluster_instance_class       = data.aws_rds_orderable_db_instance.test.instance_class
  db_subnet_group_name            = aws_db_subnet_group.test.name
  engine                          = data.aws_rds_orderable_db_instance.test.engine
  engine_mode                     = "provisioned"
  engine_version                  = data.aws_rds_orderable_db_instance.test.engine_version
}
`, tfrds.ClusterEnginePostgres, mainInstanceClasses, rName, enabledCloudwatchLogExports1, enabledCloudwatchLogExports2))
}

func testAccClusterConfig_kmsKey(n int) string {
	return fmt.Sprintf(`
data "aws_caller_identity" "current" {}
data "aws_partition" "current" {}

resource "aws_kms_key" "foo" {
  description = "Terraform acc test %[1]d"

  policy = <<POLICY
{
  "Version": "2012-10-17",
  "Id": "kms-tf-1",
  "Statement": [
    {
      "Sid": "Enable IAM User Permissions",
      "Effect": "Allow",
      "Principal": {
        "AWS": "arn:${data.aws_partition.current.partition}:iam::${data.aws_caller_identity.current.account_id}:root"
      },
      "Action": "kms:*",
      "Resource": "*"
    }
  ]
}
 POLICY

}

resource "aws_rds_cluster" "test" {
  cluster_identifier  = "tf-aurora-cluster-%[1]d"
  database_name       = "mydb"
  engine              = %[2]q
  master_username     = "foo"
  master_password     = "mustbeeightcharaters"
  storage_encrypted   = true
  kms_key_id          = aws_kms_key.foo.arn
  skip_final_snapshot = true
}
`, n, tfrds.ClusterEngineAuroraMySQL)
}

func testAccClusterConfig_encrypted(n int) string {
	return fmt.Sprintf(`
resource "aws_rds_cluster" "test" {
  cluster_identifier  = "tf-aurora-cluster-%d"
  database_name       = "mydb"
  engine              = %[2]q
  master_username     = "foo"
  master_password     = "mustbeeightcharaters"
  storage_encrypted   = true
  skip_final_snapshot = true
}
`, n, tfrds.ClusterEngineAuroraMySQL)
}

func testAccClusterConfig_backups(n int) string {
	return fmt.Sprintf(`
resource "aws_rds_cluster" "test" {
  cluster_identifier           = "tf-aurora-cluster-%[1]d"
  database_name                = "mydb"
  engine                       = %[2]q
  master_username              = "foo"
  master_password              = "mustbeeightcharaters"
  backup_retention_period      = 5
  preferred_backup_window      = "07:00-09:00"
  preferred_maintenance_window = "tue:04:00-tue:04:30"
  skip_final_snapshot          = true
}
`, n, tfrds.ClusterEngineAuroraMySQL)
}

func testAccClusterConfig_backupsUpdate(n int) string {
	return fmt.Sprintf(`
resource "aws_rds_cluster" "test" {
  cluster_identifier           = "tf-aurora-cluster-%[1]d"
  database_name                = "mydb"
  engine                       = %[2]q
  master_username              = "foo"
  master_password              = "mustbeeightcharaters"
  backup_retention_period      = 10
  preferred_backup_window      = "03:00-09:00"
  preferred_maintenance_window = "wed:01:00-wed:01:30"
  apply_immediately            = true
  skip_final_snapshot          = true
}
`, n, tfrds.ClusterEngineAuroraMySQL)
}

func testAccClusterConfig_iamAuth(n int) string {
	return fmt.Sprintf(`
resource "aws_rds_cluster" "test" {
  cluster_identifier                  = "tf-aurora-cluster-%[1]d"
  database_name                       = "mydb"
  engine                              = %[2]q
  master_username                     = "foo"
  master_password                     = "mustbeeightcharaters"
  iam_database_authentication_enabled = true
  skip_final_snapshot                 = true
}
`, n, tfrds.ClusterEngineAuroraMySQL)
}

func testAccClusterConfig_engineVersion(rName string, upgrade bool) string {
	return fmt.Sprintf(`
data "aws_rds_engine_version" "test" {
  engine                    = %[1]q
  latest                    = true
  preferred_upgrade_targets = [data.aws_rds_engine_version.upgrade.version_actual]
}

data "aws_rds_engine_version" "upgrade" {
  engine = %[1]q
}

locals {
  parameter_group_name = %[2]t ? data.aws_rds_engine_version.upgrade.parameter_group_family : data.aws_rds_engine_version.test.parameter_group_family
  engine_version       = %[2]t ? data.aws_rds_engine_version.upgrade.version : data.aws_rds_engine_version.test.version
}

resource "aws_rds_cluster" "test" {
  cluster_identifier              = %[3]q
  database_name                   = "test"
  db_cluster_parameter_group_name = "default.${local.parameter_group_name}"
  engine                          = data.aws_rds_engine_version.test.engine
  engine_version                  = local.engine_version
  master_password                 = "avoid-plaintext-passwords"
  master_username                 = "tfacctest"
  skip_final_snapshot             = true
  apply_immediately               = true
}
`, tfrds.ClusterEngineAuroraPostgreSQL, upgrade, rName)
}

func testAccClusterConfig_engineVersionPrimaryInstance(rName string, upgrade bool) string {
	return fmt.Sprintf(`
data "aws_rds_engine_version" "test" {
  engine                    = %[1]q
  latest                    = true
  preferred_upgrade_targets = [data.aws_rds_engine_version.upgrade.version_actual]
}

data "aws_rds_engine_version" "upgrade" {
  engine = %[1]q
}

locals {
  parameter_group_name = %[2]t ? data.aws_rds_engine_version.upgrade.parameter_group_family : data.aws_rds_engine_version.test.parameter_group_family
  engine_version       = %[2]t ? data.aws_rds_engine_version.upgrade.version : data.aws_rds_engine_version.test.version
}

resource "aws_rds_cluster" "test" {
  cluster_identifier              = %[3]q
  database_name                   = "test"
  db_cluster_parameter_group_name = "default.${local.parameter_group_name}"
  engine                          = data.aws_rds_engine_version.test.engine
  engine_version                  = local.engine_version
  master_password                 = "avoid-plaintext-passwords"
  master_username                 = "tfacctest"
  skip_final_snapshot             = true
  apply_immediately               = true
}

data "aws_rds_orderable_db_instance" "test" {
  engine                     = data.aws_rds_engine_version.test.engine
  engine_version             = data.aws_rds_engine_version.test.version
  preferred_instance_classes = [%[4]s]
}

resource "aws_rds_cluster_instance" "test" {
  identifier         = %[3]q
  cluster_identifier = aws_rds_cluster.test.cluster_identifier
  engine             = aws_rds_cluster.test.engine
  instance_class     = data.aws_rds_orderable_db_instance.test.instance_class
}
`, tfrds.ClusterEngineAuroraPostgreSQL, upgrade, rName, mainInstanceClasses)
}

func testAccClusterConfig_port(rName string, port int) string {
	return fmt.Sprintf(`
resource "aws_rds_cluster" "test" {
  cluster_identifier  = %[1]q
  database_name       = "mydb"
  engine              = %[2]q
  master_password     = "mustbeeightcharaters"
  master_username     = "foo"
  port                = %[3]d
  skip_final_snapshot = true
}
`, rName, tfrds.ClusterEngineAuroraPostgreSQL, port)
}

func testAccClusterConfig_includingIAMRoles(n int) string {
	return fmt.Sprintf(`
resource "aws_iam_role" "rds_sample_role" {
  name = "rds_sample_role_%[1]d"
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

resource "aws_iam_role_policy" "rds_policy" {
  name = "rds_sample_role_policy_%[1]d"
  role = aws_iam_role.rds_sample_role.name

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

resource "aws_iam_role" "another_rds_sample_role" {
  name = "another_rds_sample_role_%[1]d"
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

resource "aws_iam_role_policy" "another_rds_policy" {
  name = "another_rds_sample_role_policy_%[1]d"
  role = aws_iam_role.another_rds_sample_role.name

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

resource "aws_rds_cluster" "test" {
  cluster_identifier  = "tf-aurora-cluster-%[1]d"
  database_name       = "mydb"
  engine              = %[2]q
  master_username     = "foo"
  master_password     = "mustbeeightcharaters"
  skip_final_snapshot = true

  tags = {
    Environment = "production"
  }

  depends_on = [aws_iam_role.another_rds_sample_role, aws_iam_role.rds_sample_role]
}
`, n, tfrds.ClusterEngineAuroraMySQL)
}

func testAccClusterConfig_addIAMRoles(n int) string {
	return fmt.Sprintf(`
resource "aws_iam_role" "rds_sample_role" {
  name = "rds_sample_role_%[1]d"
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

resource "aws_iam_role_policy" "rds_policy" {
  name = "rds_sample_role_policy_%[1]d"
  role = aws_iam_role.rds_sample_role.name

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

resource "aws_iam_role" "another_rds_sample_role" {
  name = "another_rds_sample_role_%[1]d"
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

resource "aws_iam_role_policy" "another_rds_policy" {
  name = "another_rds_sample_role_policy_%[1]d"
  role = aws_iam_role.another_rds_sample_role.name

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

resource "aws_rds_cluster" "test" {
  cluster_identifier  = "tf-aurora-cluster-%[1]d"
  database_name       = "mydb"
  engine              = %[2]q
  master_username     = "foo"
  master_password     = "mustbeeightcharaters"
  skip_final_snapshot = true
  iam_roles           = [aws_iam_role.rds_sample_role.arn, aws_iam_role.another_rds_sample_role.arn]

  tags = {
    Environment = "production"
  }

  depends_on = [aws_iam_role.another_rds_sample_role, aws_iam_role.rds_sample_role]
}
`, n, tfrds.ClusterEngineAuroraMySQL)
}

func testAccClusterConfig_removeIAMRoles(n int) string {
	return fmt.Sprintf(`
resource "aws_iam_role" "another_rds_sample_role" {
  name = "another_rds_sample_role_%[1]d"
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

resource "aws_iam_role_policy" "another_rds_policy" {
  name = "another_rds_sample_role_policy_%[1]d"
  role = aws_iam_role.another_rds_sample_role.name

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

resource "aws_rds_cluster" "test" {
  cluster_identifier  = "tf-aurora-cluster-%[1]d"
  database_name       = "mydb"
  engine              = %[2]q
  master_username     = "foo"
  master_password     = "mustbeeightcharaters"
  skip_final_snapshot = true
  iam_roles           = [aws_iam_role.another_rds_sample_role.arn]

  tags = {
    Environment = "production"
  }

  depends_on = [aws_iam_role.another_rds_sample_role]
}
`, n, tfrds.ClusterEngineAuroraMySQL)
}

func testAccClusterConfig_replicationSourceIDKMSKeyID(rName string) string {
	return acctest.ConfigCompose(acctest.ConfigMultipleRegionProvider(2), fmt.Sprintf(`
data "aws_availability_zones" "available" {
  provider = "awsalternate"

  state = "available"

  filter {
    name   = "opt-in-status"
    values = ["opt-in-not-required"]
  }
}

data "aws_caller_identity" "current" {}
data "aws_partition" "current" {}
data "aws_region" "current" {}

data "aws_rds_engine_version" "default" {
  engine = %[1]q
}

data "aws_rds_orderable_db_instance" "test" {
  engine                     = data.aws_rds_engine_version.default.engine
  engine_version             = data.aws_rds_engine_version.default.version
  preferred_instance_classes = [%[2]s]
}

resource "aws_rds_cluster_parameter_group" "test" {
  name        = %[3]q
  family      = data.aws_rds_engine_version.default.parameter_group_family
  description = "RDS default cluster parameter group"

  parameter {
    name         = "binlog_format"
    value        = "STATEMENT"
    apply_method = "pending-reboot"
  }
}

resource "aws_rds_cluster" "test" {
  cluster_identifier              = "%[3]s-primary"
  db_cluster_parameter_group_name = aws_rds_cluster_parameter_group.test.name
  database_name                   = "test"
  engine                          = data.aws_rds_engine_version.default.engine
  master_username                 = "tfacctest"
  master_password                 = "avoid-plaintext-passwords"
  storage_encrypted               = true
  skip_final_snapshot             = true
}

resource "aws_rds_cluster_instance" "test" {
  identifier         = "%[3]s-primary"
  cluster_identifier = aws_rds_cluster.test.id
  instance_class     = data.aws_rds_orderable_db_instance.test.instance_class
  engine             = aws_rds_cluster.test.engine
  engine_version     = aws_rds_cluster.test.engine_version
}

resource "aws_kms_key" "test" {
  provider = "awsalternate"

  description = %[3]q

  policy = <<POLICY
{
  "Version": "2012-10-17",
  "Id": "kms-tf-1",
  "Statement": [
    {
      "Sid": "Enable IAM User Permissions",
      "Effect": "Allow",
      "Principal": {
        "AWS": "arn:${data.aws_partition.current.partition}:iam::${data.aws_caller_identity.current.account_id}:root"
      },
      "Action": "kms:*",
      "Resource": "*"
    }
  ]
}
  POLICY
}

resource "aws_vpc" "test" {
  provider = "awsalternate"

  cidr_block = "10.0.0.0/16"

  tags = {
    Name = %[3]q
  }
}

resource "aws_subnet" "test" {
  provider = "awsalternate"

  count = 3

  vpc_id            = aws_vpc.test.id
  availability_zone = data.aws_availability_zones.available.names[count.index]
  cidr_block        = "10.0.${count.index}.0/24"

  tags = {
    Name = %[3]q
  }
}

resource "aws_db_subnet_group" "test" {
  provider = "awsalternate"

  name       = %[3]q
  subnet_ids = aws_subnet.test[*].id
}

resource "aws_rds_cluster" "alternate" {
  provider = "awsalternate"

  cluster_identifier            = "%[3]s-replica"
  db_subnet_group_name          = aws_db_subnet_group.test.name
  engine                        = %[1]q
  kms_key_id                    = aws_kms_key.test.arn
  storage_encrypted             = true
  skip_final_snapshot           = true
  replication_source_identifier = aws_rds_cluster.test.arn
  source_region                 = data.aws_region.current.name

  depends_on = [
    aws_rds_cluster_instance.test,
  ]
}
`, tfrds.ClusterEngineAuroraMySQL, mainInstanceClasses, rName))
}

func testAccClusterConfig_networkType(rName string, networkType string) string {
	return acctest.ConfigCompose(
		acctest.ConfigVPCWithSubnetsIPv6(rName, 2),
		fmt.Sprintf(`
resource "aws_db_subnet_group" "test" {
  name       = %[1]q
  subnet_ids = aws_subnet.test[*].id
}

data "aws_rds_orderable_db_instance" "test" {
  engine                     = %[2]q
  engine_latest_version      = true
  preferred_instance_classes = [%[3]s]
  supported_network_types    = [%[4]q]
}

resource "aws_rds_cluster" "test" {
  cluster_identifier   = %[1]q
  db_subnet_group_name = aws_db_subnet_group.test.name
  network_type         = %[4]q
  engine               = %[2]q
  engine_version       = data.aws_rds_orderable_db_instance.test.engine_version
  master_password      = "avoid-plaintext-passwords"
  master_username      = "tfacctest"
  skip_final_snapshot  = true
  apply_immediately    = true
}
`, rName, tfrds.ClusterEngineAuroraPostgreSQL, mainInstanceClasses, networkType))
}

func testAccClusterConfig_deletionProtection(rName string, deletionProtection bool) string {
	return fmt.Sprintf(`
resource "aws_rds_cluster" "test" {
  cluster_identifier  = %[1]q
  deletion_protection = %[2]t
  engine              = %[3]q
  master_password     = "barbarbarbar"
  master_username     = "foo"
  skip_final_snapshot = true
}
`, rName, deletionProtection, tfrds.ClusterEngineAuroraMySQL)
}

func testAccClusterConfig_engineMode(rName, engineMode string) string {
	return fmt.Sprintf(`
resource "aws_rds_cluster" "test" {
  cluster_identifier  = %[1]q
  engine              = %[2]q
  engine_mode         = %[3]q
  master_password     = "barbarbarbar"
  master_username     = "foo"
  skip_final_snapshot = true
}
`, rName, tfrds.ClusterEngineAuroraMySQL, engineMode)
}

func testAccClusterConfig_engineMode_serverless(rName string) string {
	return fmt.Sprintf(`
resource "aws_rds_cluster" "test" {
  cluster_identifier  = %[1]q
  engine              = %[2]q
  engine_mode         = "serverless"
  master_password     = "barbarbarbar"
  master_username     = "foo"
  skip_final_snapshot = true

  scaling_configuration {
    min_capacity = 2
  }
}
`, rName, tfrds.ClusterEngineAuroraMySQL)
}

func testAccClusterConfig_EngineMode_global(rName string) string {
	return fmt.Sprintf(`
resource "aws_rds_cluster" "test" {
  cluster_identifier  = %[1]q
  engine              = %[2]q
  engine_version      = data.aws_rds_engine_version.default.version
  master_password     = "avoid-plaintext-passwords"
  master_username     = "tfacctest"
  skip_final_snapshot = true
}

data "aws_rds_engine_version" "default" {
  engine = %[2]q
}
`, rName, tfrds.ClusterEngineAuroraMySQL)
}

func testAccClusterConfig_GlobalClusterID_EngineMode_global(rName string) string {
	return fmt.Sprintf(`
resource "aws_rds_global_cluster" "test" {
  engine                    = %[1]q
  engine_version            = data.aws_rds_engine_version.default.version
  force_destroy             = true # Partial configuration removal ordering fix for after Terraform 0.12
  global_cluster_identifier = %[2]q
}

resource "aws_rds_cluster" "test" {
  cluster_identifier        = %[2]q
  global_cluster_identifier = aws_rds_global_cluster.test.id
  engine                    = %[1]q
  engine_version            = data.aws_rds_engine_version.default.version
  master_password           = "avoid-plaintext-passwords"
  master_username           = "tfacctest"
  skip_final_snapshot       = true
}

data "aws_rds_engine_version" "default" {
  engine = %[1]q
}
`, tfrds.ClusterEngineAuroraMySQL, rName)
}

func testAccClusterConfig_GlobalClusterID_EngineMode_globalUpdate(rName, globalClusterIdentifierResourceName string) string {
	return fmt.Sprintf(`
resource "aws_rds_global_cluster" "test" {
  count = 2

  engine                    = %[1]q
  engine_version            = data.aws_rds_engine_version.default.version
  global_cluster_identifier = "%[2]s-${count.index}"
}

resource "aws_rds_cluster" "test" {
  cluster_identifier        = %[2]q
  global_cluster_identifier = %[3]s.id
  engine                    = %[1]q
  engine_version            = data.aws_rds_engine_version.default.version
  master_password           = "avoid-plaintext-passwords"
  master_username           = "tfacctest"
  skip_final_snapshot       = true
}

data "aws_rds_engine_version" "default" {
  engine = %[1]q
}
`, tfrds.ClusterEngineAuroraMySQL, rName, globalClusterIdentifierResourceName)
}

func testAccClusterConfig_GlobalClusterID_EngineMode_provisioned(rName string) string {
	return fmt.Sprintf(`
data "aws_rds_orderable_db_instance" "test" {
  engine                     = %[1]q
  engine_latest_version      = true
  preferred_instance_classes = [%[2]s]
  supported_engine_modes     = ["provisioned"]
  supports_clusters          = true
}

resource "aws_rds_global_cluster" "test" {
  engine                    = %[1]q
  engine_version            = data.aws_rds_orderable_db_instance.test.engine_version
  global_cluster_identifier = %[3]q
}

resource "aws_rds_cluster" "test" {
  cluster_identifier        = %[3]q
  engine                    = aws_rds_global_cluster.test.engine
  engine_version            = aws_rds_global_cluster.test.engine_version
  global_cluster_identifier = aws_rds_global_cluster.test.id
  master_password           = "barbarbarbar"
  master_username           = "foo"
  skip_final_snapshot       = true
}
`, tfrds.ClusterEngineAuroraPostgreSQL, mainInstanceClasses, rName)
}

func testAccClusterConfig_GlobalClusterID_primarySecondaryClusters(rNameGlobal, rNamePrimary, rNameSecondary string) string {
	return acctest.ConfigCompose(
		acctest.ConfigMultipleRegionProvider(2),
		fmt.Sprintf(`
data "aws_region" "current" {}

data "aws_availability_zones" "alternate" {
  provider = "awsalternate"
  state    = "available"

  filter {
    name   = "opt-in-status"
    values = ["opt-in-not-required"]
  }
}

data "aws_rds_orderable_db_instance" "test" {
  engine                     = %[1]q
  engine_latest_version      = true
  preferred_instance_classes = [%[2]s]
  supports_clusters          = true
  supports_global_databases  = true
}

resource "aws_rds_global_cluster" "test" {
  global_cluster_identifier = %[3]q
  engine                    = %[1]q
  engine_version            = data.aws_rds_orderable_db_instance.test.engine_version
}

resource "aws_rds_cluster" "primary" {
  cluster_identifier        = %[4]q
  database_name             = "mydb"
  master_username           = "foo"
  master_password           = "barbarbar"
  skip_final_snapshot       = true
  global_cluster_identifier = aws_rds_global_cluster.test.id
  engine                    = aws_rds_global_cluster.test.engine
  engine_version            = aws_rds_global_cluster.test.engine_version
}

resource "aws_rds_cluster_instance" "primary" {
  identifier         = %[4]q
  cluster_identifier = aws_rds_cluster.primary.id
  instance_class     = data.aws_rds_orderable_db_instance.test.instance_class
  engine             = aws_rds_cluster.primary.engine
  engine_version     = aws_rds_cluster.primary.engine_version
}

resource "aws_vpc" "alternate" {
  provider   = "awsalternate"
  cidr_block = "10.0.0.0/16"

  tags = {
    Name = %[5]q
  }
}

resource "aws_subnet" "alternate" {
  provider          = "awsalternate"
  count             = 3
  vpc_id            = aws_vpc.alternate.id
  availability_zone = data.aws_availability_zones.alternate.names[count.index]
  cidr_block        = "10.0.${count.index}.0/24"

  tags = {
    Name = %[5]q
  }
}

resource "aws_db_subnet_group" "alternate" {
  provider   = "awsalternate"
  name       = %[5]q
  subnet_ids = aws_subnet.alternate[*].id
}

resource "aws_rds_cluster" "secondary" {
  provider                  = "awsalternate"
  cluster_identifier        = %[5]q
  db_subnet_group_name      = aws_db_subnet_group.alternate.name
  skip_final_snapshot       = true
  source_region             = data.aws_region.current.name
  global_cluster_identifier = aws_rds_global_cluster.test.id
  engine                    = aws_rds_global_cluster.test.engine
  engine_version            = aws_rds_global_cluster.test.engine_version
  depends_on                = [aws_rds_cluster_instance.primary]

  lifecycle {
    ignore_changes = [
      replication_source_identifier,
    ]
  }
}

resource "aws_rds_cluster_instance" "secondary" {
  provider           = "awsalternate"
  identifier         = %[5]q
  cluster_identifier = aws_rds_cluster.secondary.id
  instance_class     = data.aws_rds_orderable_db_instance.test.instance_class
  engine             = aws_rds_cluster.secondary.engine
  engine_version     = aws_rds_cluster.secondary.engine_version
}
`, tfrds.ClusterEngineAuroraMySQL, mainInstanceClasses, rNameGlobal, rNamePrimary, rNameSecondary))
}

func testAccClusterConfig_GlobalClusterID_secondaryClustersWriteForwarding(rNameGlobal, rNamePrimary, rNameSecondary string) string {
	return acctest.ConfigCompose(
		acctest.ConfigMultipleRegionProvider(2),
		fmt.Sprintf(`
data "aws_region" "current" {}

data "aws_availability_zones" "alternate" {
  provider = "awsalternate"
  state    = "available"

  filter {
    name   = "opt-in-status"
    values = ["opt-in-not-required"]
  }
}

data "aws_rds_orderable_db_instance" "test" {
  engine                     = %[1]q
  engine_latest_version      = true
  preferred_instance_classes = [%[2]s]
  supports_clusters          = true
  supports_global_databases  = true
}

resource "aws_rds_global_cluster" "test" {
  global_cluster_identifier = %[3]q
  engine                    = data.aws_rds_orderable_db_instance.test.engine
  engine_version            = data.aws_rds_orderable_db_instance.test.engine_version
}

resource "aws_rds_cluster" "primary" {
  cluster_identifier        = %[4]q
  database_name             = "mydb"
  master_username           = "foo"
  master_password           = "barbarbar"
  skip_final_snapshot       = true
  global_cluster_identifier = aws_rds_global_cluster.test.id
  engine                    = aws_rds_global_cluster.test.engine
  engine_version            = aws_rds_global_cluster.test.engine_version
}

resource "aws_rds_cluster_instance" "primary" {
  identifier         = %[4]q
  cluster_identifier = aws_rds_cluster.primary.id
  instance_class     = data.aws_rds_orderable_db_instance.test.instance_class
  engine             = aws_rds_cluster.primary.engine
  engine_version     = aws_rds_cluster.primary.engine_version
}

resource "aws_vpc" "alternate" {
  provider   = "awsalternate"
  cidr_block = "10.0.0.0/16"

  tags = {
    Name = %[5]q
  }
}

resource "aws_subnet" "alternate" {
  provider          = "awsalternate"
  count             = 3
  vpc_id            = aws_vpc.alternate.id
  availability_zone = data.aws_availability_zones.alternate.names[count.index]
  cidr_block        = "10.0.${count.index}.0/24"

  tags = {
    Name = %[5]q
  }
}

resource "aws_db_subnet_group" "alternate" {
  provider   = "awsalternate"
  name       = %[5]q
  subnet_ids = aws_subnet.alternate[*].id
}

resource "aws_rds_cluster" "secondary" {
  provider                       = "awsalternate"
  cluster_identifier             = %[5]q
  db_subnet_group_name           = aws_db_subnet_group.alternate.name
  skip_final_snapshot            = true
  source_region                  = data.aws_region.current.name
  global_cluster_identifier      = aws_rds_global_cluster.test.id
  enable_global_write_forwarding = true
  engine                         = aws_rds_global_cluster.test.engine
  engine_version                 = aws_rds_global_cluster.test.engine_version
  depends_on                     = [aws_rds_cluster_instance.primary]

  lifecycle {
    ignore_changes = [
      replication_source_identifier,
    ]
  }
}

resource "aws_rds_cluster_instance" "secondary" {
  provider           = "awsalternate"
  identifier         = %[5]q
  cluster_identifier = aws_rds_cluster.secondary.id
  instance_class     = data.aws_rds_orderable_db_instance.test.instance_class
  engine             = aws_rds_cluster.secondary.engine
  engine_version     = aws_rds_cluster.secondary.engine_version
}
`, tfrds.ClusterEngineAuroraMySQL, mainInstanceClasses, rNameGlobal, rNamePrimary, rNameSecondary))
}

func testAccClusterConfig_GlobalClusterID_replicationSourceID(rName string) string {
	return acctest.ConfigCompose(
		acctest.ConfigMultipleRegionProvider(2),
		fmt.Sprintf(`
data "aws_region" "current" {}

data "aws_availability_zones" "alternate" {
  provider = "awsalternate"
  state    = "available"

  filter {
    name   = "opt-in-status"
    values = ["opt-in-not-required"]
  }
}

data "aws_rds_engine_version" "default" {
  engine = %[1]q
}

data "aws_rds_orderable_db_instance" "test" {
  engine                     = data.aws_rds_engine_version.default.engine
  engine_version             = data.aws_rds_engine_version.default.version
  preferred_instance_classes = [%[2]s]
  supports_global_databases  = true
}

resource "aws_rds_global_cluster" "test" {
  global_cluster_identifier = %[3]q
  engine                    = data.aws_rds_engine_version.default.engine
  engine_version            = data.aws_rds_engine_version.default.version
}

resource "aws_rds_cluster" "primary" {
  cluster_identifier        = "%[3]s-primary"
  database_name             = "mydb"
  engine                    = aws_rds_global_cluster.test.engine
  engine_version            = aws_rds_global_cluster.test.engine_version
  global_cluster_identifier = aws_rds_global_cluster.test.id
  master_password           = "barbarbar"
  master_username           = "foo"
  skip_final_snapshot       = true
}

resource "aws_rds_cluster_instance" "primary" {
  cluster_identifier = aws_rds_cluster.primary.id
  engine             = aws_rds_cluster.primary.engine
  engine_version     = aws_rds_cluster.primary.engine_version
  identifier         = "%[3]s-primary"
  instance_class     = data.aws_rds_orderable_db_instance.test.instance_class
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

resource "aws_db_subnet_group" "alternate" {
  provider   = "awsalternate"
  name       = %[3]q
  subnet_ids = aws_subnet.alternate[*].id
}

resource "aws_rds_cluster" "secondary" {
  provider   = "awsalternate"
  depends_on = [aws_rds_cluster_instance.primary]

  cluster_identifier            = "%[3]s-secondary"
  db_subnet_group_name          = aws_db_subnet_group.alternate.name
  engine                        = aws_rds_global_cluster.test.engine
  engine_version                = aws_rds_global_cluster.test.engine_version
  global_cluster_identifier     = aws_rds_global_cluster.test.id
  replication_source_identifier = aws_rds_cluster.primary.arn
  skip_final_snapshot           = true
  source_region                 = data.aws_region.current.name
}

resource "aws_rds_cluster_instance" "secondary" {
  provider = "awsalternate"

  cluster_identifier = aws_rds_cluster.secondary.id
  engine             = aws_rds_cluster.secondary.engine
  engine_version     = aws_rds_cluster.secondary.engine_version
  identifier         = "%[3]s-secondary"
  instance_class     = aws_rds_cluster_instance.primary.instance_class
}
`, tfrds.ClusterEngineAuroraPostgreSQL, mainInstanceClasses, rName))
}

func testAccClusterConfig_scalingConfiguration(rName string, autoPause bool, maxCapacity, minCapacity, secondsBeforeTimeout, secondsUntilAutoPause int, timeoutAction string) string {
	return fmt.Sprintf(`
resource "aws_rds_cluster" "test" {
  cluster_identifier  = %[1]q
  engine              = %[2]q
  engine_mode         = "serverless"
  master_password     = "barbarbarbar"
  master_username     = "foo"
  skip_final_snapshot = true

  scaling_configuration {
    auto_pause               = %[3]t
    max_capacity             = %[4]d
    min_capacity             = %[5]d
    seconds_before_timeout   = %[6]d
    seconds_until_auto_pause = %[7]d
    timeout_action           = %[8]q
  }
}
`, rName, tfrds.ClusterEngineAuroraMySQL, autoPause, maxCapacity, minCapacity, secondsBeforeTimeout, secondsUntilAutoPause, timeoutAction)
}

func testAccClusterConfig_serverlessV2ScalingConfiguration(rName string, maxCapacity, minCapacity float64) string {
	return fmt.Sprintf(`
data "aws_rds_engine_version" "test" {
  engine      = %[1]q
  latest      = true
  include_all = true

  filter {
    name   = "engine-mode"
    values = ["serverless"]
  }
}

resource "aws_rds_cluster" "test" {
  cluster_identifier  = %[2]q
  master_password     = "barbarbarbar"
  master_username     = "foo"
  skip_final_snapshot = true
  engine              = data.aws_rds_engine_version.test.engine
  engine_version      = data.aws_rds_engine_version.test.version

  serverlessv2_scaling_configuration {
    max_capacity = %[3]f
    min_capacity = %[4]f
  }
}
`, tfrds.ClusterEngineAuroraPostgreSQL, rName, maxCapacity, minCapacity)
}

func testAccClusterConfig_ScalingConfiguration_defaultMinCapacity(rName string, autoPause bool, maxCapacity, secondsUntilAutoPause int, timeoutAction string) string {
	return fmt.Sprintf(`
resource "aws_rds_cluster" "test" {
  cluster_identifier  = %[1]q
  engine              = %[2]q
  engine_mode         = "serverless"
  master_password     = "barbarbarbar"
  master_username     = "foo"
  skip_final_snapshot = true

  scaling_configuration {
    auto_pause               = %[3]t
    max_capacity             = %[4]d
    seconds_until_auto_pause = %[5]d
    timeout_action           = %[6]q
  }
}
`, rName, tfrds.ClusterEngineAuroraMySQL, autoPause, maxCapacity, secondsUntilAutoPause, timeoutAction)
}

func testAccClusterConfig_snapshotID(rName string) string {
	return fmt.Sprintf(`
resource "aws_rds_cluster" "source" {
  cluster_identifier  = "%[1]s-source"
  engine              = %[2]q
  master_password     = "barbarbarbar"
  master_username     = "foo"
  skip_final_snapshot = true
}

resource "aws_db_cluster_snapshot" "test" {
  db_cluster_identifier          = aws_rds_cluster.source.id
  db_cluster_snapshot_identifier = %[1]q
}

resource "aws_rds_cluster" "test" {
  cluster_identifier  = %[1]q
  engine              = %[2]q
  skip_final_snapshot = true
  snapshot_identifier = aws_db_cluster_snapshot.test.id
}
`, rName, tfrds.ClusterEngineAuroraMySQL)
}

func testAccClusterConfig_SnapshotID_deletionProtection(rName string, deletionProtection bool) string {
	return fmt.Sprintf(`
resource "aws_rds_cluster" "source" {
  cluster_identifier  = "%[1]s-source"
  engine              = %[2]q
  master_password     = "barbarbarbar"
  master_username     = "foo"
  skip_final_snapshot = true
}

resource "aws_db_cluster_snapshot" "test" {
  db_cluster_identifier          = aws_rds_cluster.source.id
  db_cluster_snapshot_identifier = %[1]q
}

resource "aws_rds_cluster" "test" {
  cluster_identifier  = %[1]q
  engine              = %[2]q
  deletion_protection = %[3]t
  skip_final_snapshot = true
  snapshot_identifier = aws_db_cluster_snapshot.test.id
}
`, rName, tfrds.ClusterEngineAuroraMySQL, deletionProtection)
}

func testAccClusterConfig_SnapshotID_engineMode(rName, engineMode string) string {
	return fmt.Sprintf(`
resource "aws_rds_cluster" "source" {
  cluster_identifier  = "%[1]s-source"
  engine              = %[2]q
  engine_mode         = %[3]q
  master_password     = "barbarbarbar"
  master_username     = "foo"
  skip_final_snapshot = true
}

resource "aws_db_cluster_snapshot" "test" {
  db_cluster_identifier          = aws_rds_cluster.source.id
  db_cluster_snapshot_identifier = %[1]q
}

resource "aws_rds_cluster" "test" {
  cluster_identifier  = %[1]q
  engine              = %[2]q
  engine_mode         = %[3]q
  skip_final_snapshot = true
  snapshot_identifier = aws_db_cluster_snapshot.test.id
}
`, rName, tfrds.ClusterEngineAuroraMySQL, engineMode)
}

func testAccClusterConfig_SnapshotID_engineVersion(rName string, same bool) string {
	return fmt.Sprintf(`
data "aws_rds_engine_version" "test" {
  engine                    = data.aws_rds_engine_version.upgrade.engine
  latest                    = true
  preferred_upgrade_targets = [data.aws_rds_engine_version.upgrade.version_actual]
}

data "aws_rds_engine_version" "upgrade" {
  engine = %[1]q
}

resource "aws_rds_cluster" "source" {
  cluster_identifier  = "%[2]s-source"
  engine              = data.aws_rds_engine_version.test.engine
  engine_version      = data.aws_rds_engine_version.test.version
  master_password     = "barbarbarbar"
  master_username     = "foo"
  skip_final_snapshot = true
}

resource "aws_db_cluster_snapshot" "test" {
  db_cluster_identifier          = aws_rds_cluster.source.id
  db_cluster_snapshot_identifier = %[2]q
}

resource "aws_rds_cluster" "test" {
  cluster_identifier  = %[2]q
  engine              = data.aws_rds_engine_version.test.engine
  engine_version      = %[3]t ? data.aws_rds_engine_version.test.version : data.aws_rds_engine_version.upgrade.version
  skip_final_snapshot = true
  snapshot_identifier = aws_db_cluster_snapshot.test.id
}
`, tfrds.ClusterEngineAuroraPostgreSQL, rName, same)
}

func testAccClusterConfig_SnapshotID_kmsKeyID(rName string) string {
	return fmt.Sprintf(`
resource "aws_kms_key" "test" {
  deletion_window_in_days = 7
}

resource "aws_rds_cluster" "source" {
  cluster_identifier  = "%[1]s-source"
  engine              = %[2]q
  master_password     = "barbarbarbar"
  master_username     = "foo"
  skip_final_snapshot = true
}

resource "aws_db_cluster_snapshot" "test" {
  db_cluster_identifier          = aws_rds_cluster.source.id
  db_cluster_snapshot_identifier = %[1]q
}

resource "aws_rds_cluster" "test" {
  cluster_identifier  = %[1]q
  engine              = %[2]q
  kms_key_id          = aws_kms_key.test.arn
  skip_final_snapshot = true
  snapshot_identifier = aws_db_cluster_snapshot.test.id
}
`, rName, tfrds.ClusterEngineAuroraMySQL)
}

func testAccClusterConfig_SnapshotID_masterPassword(rName, masterPassword string) string {
	return fmt.Sprintf(`
resource "aws_rds_cluster" "source" {
  cluster_identifier  = "%[1]s-source"
  engine              = %[2]q
  master_password     = "barbarbarbar"
  master_username     = "foo"
  skip_final_snapshot = true
}

resource "aws_db_cluster_snapshot" "test" {
  db_cluster_identifier          = aws_rds_cluster.source.id
  db_cluster_snapshot_identifier = %[1]q
}

resource "aws_rds_cluster" "test" {
  cluster_identifier  = %[1]q
  engine              = %[2]q
  master_password     = %[3]q
  skip_final_snapshot = true
  snapshot_identifier = aws_db_cluster_snapshot.test.id
}
`, rName, tfrds.ClusterEngineAuroraMySQL, masterPassword)
}

func testAccClusterConfig_SnapshotID_masterUsername(rName, masterUsername string) string {
	return fmt.Sprintf(`
resource "aws_rds_cluster" "source" {
  cluster_identifier  = "%[1]s-source"
  engine              = %[2]q
  master_password     = "barbarbarbar"
  master_username     = "foo"
  skip_final_snapshot = true
}

resource "aws_db_cluster_snapshot" "test" {
  db_cluster_identifier          = aws_rds_cluster.source.id
  db_cluster_snapshot_identifier = %[1]q
}

resource "aws_rds_cluster" "test" {
  cluster_identifier  = %[1]q
  engine              = %[2]q
  master_username     = %[3]q
  skip_final_snapshot = true
  snapshot_identifier = aws_db_cluster_snapshot.test.id
}
`, rName, tfrds.ClusterEngineAuroraMySQL, masterUsername)
}

func testAccClusterConfig_SnapshotID_preferredBackupWindow(rName, preferredBackupWindow string) string {
	return fmt.Sprintf(`
resource "aws_rds_cluster" "source" {
  cluster_identifier  = "%[1]s-source"
  engine              = %[2]q
  master_password     = "barbarbarbar"
  master_username     = "foo"
  skip_final_snapshot = true
}

resource "aws_db_cluster_snapshot" "test" {
  db_cluster_identifier          = aws_rds_cluster.source.id
  db_cluster_snapshot_identifier = %[1]q
}

resource "aws_rds_cluster" "test" {
  cluster_identifier           = %[1]q
  engine                       = %[2]q
  preferred_backup_window      = %[3]q
  preferred_maintenance_window = "sun:09:00-sun:09:30"
  skip_final_snapshot          = true
  snapshot_identifier          = aws_db_cluster_snapshot.test.id
}
`, rName, tfrds.ClusterEngineAuroraMySQL, preferredBackupWindow)
}

func testAccClusterConfig_SnapshotID_preferredMaintenanceWindow(rName, preferredMaintenanceWindow string) string {
	// This config will never need the version updated. Use as a model for changing the other
	// tests.
	return fmt.Sprintf(`
data "aws_rds_engine_version" "test" {
  engine = %[1]q
}

resource "aws_rds_cluster" "source" {
  cluster_identifier  = "%[2]s-source"
  master_password     = "barbarbarbar"
  master_username     = "foo"
  skip_final_snapshot = true
  engine              = data.aws_rds_engine_version.test.engine
}

resource "aws_db_cluster_snapshot" "test" {
  db_cluster_identifier          = aws_rds_cluster.source.id
  db_cluster_snapshot_identifier = %[2]q
}

resource "aws_rds_cluster" "test" {
  cluster_identifier           = %[2]q
  preferred_maintenance_window = %[3]q
  skip_final_snapshot          = true
  snapshot_identifier          = aws_db_cluster_snapshot.test.id
  engine                       = data.aws_rds_engine_version.test.engine
}
`, tfrds.ClusterEngineAuroraMySQL, rName, preferredMaintenanceWindow)
}

func testAccClusterConfig_SnapshotID_tags(rName string) string {
	return fmt.Sprintf(`
resource "aws_rds_cluster" "source" {
  cluster_identifier  = "%[1]s-source"
  engine              = %[2]q
  master_password     = "barbarbarbar"
  master_username     = "foo"
  skip_final_snapshot = true
}

resource "aws_db_cluster_snapshot" "test" {
  db_cluster_identifier          = aws_rds_cluster.source.id
  db_cluster_snapshot_identifier = %[1]q
}

resource "aws_rds_cluster" "test" {
  cluster_identifier  = %[1]q
  engine              = %[2]q
  skip_final_snapshot = true
  snapshot_identifier = aws_db_cluster_snapshot.test.id

  tags = {
    key1 = "value1"
  }
}
`, rName, tfrds.ClusterEngineAuroraMySQL)
}

func testAccClusterConfig_SnapshotID_vpcSecurityGroupIDs(rName string) string {
	return fmt.Sprintf(`
data "aws_vpc" "default" {
  default = true
}

data "aws_security_group" "default" {
  name   = "default"
  vpc_id = data.aws_vpc.default.id
}

resource "aws_rds_cluster" "source" {
  cluster_identifier  = "%[1]s-source"
  engine              = %[2]q
  master_password     = "barbarbarbar"
  master_username     = "foo"
  skip_final_snapshot = true
}

resource "aws_db_cluster_snapshot" "test" {
  db_cluster_identifier          = aws_rds_cluster.source.id
  db_cluster_snapshot_identifier = %[1]q
}

resource "aws_rds_cluster" "test" {
  cluster_identifier     = %[1]q
  engine                 = %[2]q
  skip_final_snapshot    = true
  snapshot_identifier    = aws_db_cluster_snapshot.test.id
  vpc_security_group_ids = [data.aws_security_group.default.id]
}
`, rName, tfrds.ClusterEngineAuroraMySQL)
}

func testAccClusterConfig_SnapshotID_VPCSecurityGroupIds_tags(rName string) string {
	return fmt.Sprintf(`
data "aws_vpc" "default" {
  default = true
}

data "aws_security_group" "default" {
  name   = "default"
  vpc_id = data.aws_vpc.default.id
}

resource "aws_rds_cluster" "source" {
  cluster_identifier  = "%[1]s-source"
  engine              = %[2]q
  master_password     = "barbarbarbar"
  master_username     = "foo"
  skip_final_snapshot = true
}

resource "aws_db_cluster_snapshot" "test" {
  db_cluster_identifier          = aws_rds_cluster.source.id
  db_cluster_snapshot_identifier = %[1]q
}

resource "aws_rds_cluster" "test" {
  cluster_identifier     = %[1]q
  engine                 = %[2]q
  skip_final_snapshot    = true
  snapshot_identifier    = aws_db_cluster_snapshot.test.id
  vpc_security_group_ids = [data.aws_security_group.default.id]

  tags = {
    key1 = "value1"
  }
}
`, rName, tfrds.ClusterEngineAuroraMySQL)
}

func testAccClusterConfig_SnapshotID_encryptedRestore(rName string) string {
	return fmt.Sprintf(`
resource "aws_kms_key" "test" {}

resource "aws_rds_cluster" "source" {
  cluster_identifier  = "%[1]s-source"
  engine              = %[2]q
  master_password     = "barbarbarbar"
  master_username     = "foo"
  skip_final_snapshot = true
}

resource "aws_db_cluster_snapshot" "test" {
  db_cluster_identifier          = aws_rds_cluster.source.id
  db_cluster_snapshot_identifier = %[1]q
}

resource "aws_rds_cluster" "test" {
  cluster_identifier  = %[1]q
  engine              = %[2]q
  skip_final_snapshot = true
  snapshot_identifier = aws_db_cluster_snapshot.test.id

  storage_encrypted = true
  kms_key_id        = aws_kms_key.test.arn
}
`, rName, tfrds.ClusterEngineAuroraMySQL)
}

func testAccClusterConfig_copyTagsToSnapshot(n int, f bool) string {
	return fmt.Sprintf(`
resource "aws_rds_cluster" "test" {
  cluster_identifier    = "tf-aurora-cluster-%[1]d"
  engine                = %[2]q
  database_name         = "mydb"
  master_username       = "foo"
  master_password       = "mustbeeightcharaters"
  copy_tags_to_snapshot = %[3]t
  skip_final_snapshot   = true
}
`, n, tfrds.ClusterEngineAuroraMySQL, f)
}

func testAccClusterConfig_copyTagsToSnapshot_restorePointInTime(n int, f bool) string {
	return fmt.Sprintf(`
resource "aws_rds_cluster" "test" {
  cluster_identifier    = "tf-aurora-cluster-%[1]d"
  engine                = %[2]q
  database_name         = "mydb"
  master_username       = "foo"
  master_password       = "mustbeeightcharaters"
  copy_tags_to_snapshot = true
  skip_final_snapshot   = true
}

resource "aws_rds_cluster" "clone" {
  cluster_identifier    = "tf-aurora-cluster-clone-%[1]d"
  engine                = %[2]q
  database_name         = "mydb"
  master_username       = "foo"
  master_password       = "mustbeeightcharaters"
  copy_tags_to_snapshot = %[3]t
  skip_final_snapshot   = true
  restore_to_point_in_time {
    source_cluster_identifier  = aws_rds_cluster.test.cluster_identifier
    use_latest_restorable_time = true
  }
}
`, n, tfrds.ClusterEngineAuroraMySQL, f)
}

func testAccClusterConfig_auroraStorageType(rName, storageType string) string {
	return fmt.Sprintf(`
data "aws_rds_engine_version" "default" {
  engine = %[1]q
}

resource "aws_rds_cluster" "test" {
  apply_immediately   = true
  cluster_identifier  = %[2]q
  engine              = data.aws_rds_engine_version.default.engine
  engine_version      = data.aws_rds_engine_version.default.version
  master_password     = "avoid-plaintext-passwords"
  master_username     = "tfacctest"
  skip_final_snapshot = true
  storage_type        = %[3]q
}

`, tfrds.ClusterEngineAuroraPostgreSQL, rName, storageType)
}

func testAccClusterConfig_auroraStorageTypeNotDefined(rName string) string {
	return fmt.Sprintf(`
data "aws_rds_engine_version" "default" {
  engine = %[1]q
}

data "aws_rds_orderable_db_instance" "default" {
  engine                     = data.aws_rds_engine_version.default.engine
  engine_version             = data.aws_rds_engine_version.default.version
  preferred_instance_classes = [%[2]s]
}

resource "aws_rds_cluster" "test" {
  apply_immediately   = true
  cluster_identifier  = %[3]q
  engine              = data.aws_rds_engine_version.default.engine
  engine_version      = data.aws_rds_engine_version.default.version
  master_password     = "avoid-plaintext-passwords"
  master_username     = "tfacctest"
  skip_final_snapshot = true
}

`, tfrds.ClusterEngineAuroraPostgreSQL, mainInstanceClasses, rName)
}

func testAccClusterConfig_enableHTTPEndpoint(rName string, enableHttpEndpoint bool) string {
	return fmt.Sprintf(`
resource "aws_rds_cluster" "test" {
  cluster_identifier   = %[1]q
  engine               = %[2]q
  engine_mode          = "serverless"
  master_password      = "barbarbarbar"
  master_username      = "foo"
  skip_final_snapshot  = true
  enable_http_endpoint = %[3]t

  scaling_configuration {
    auto_pause               = false
    max_capacity             = 128
    min_capacity             = 4
    seconds_until_auto_pause = 301
    timeout_action           = "RollbackCapacityChange"
  }
}
`, rName, tfrds.ClusterEngineAuroraMySQL, enableHttpEndpoint)
}

func testAccClusterConfig_enableHTTPEndpointProvisioned(rName string, enableHttpEndpoint bool) string {
	return fmt.Sprintf(`
resource "aws_rds_cluster" "test" {
  cluster_identifier   = %[1]q
  engine               = %[2]q
  engine_mode          = "provisioned"
  master_password      = "barbarbarbar"
  master_username      = "foo"
  skip_final_snapshot  = true
  enable_http_endpoint = %[3]t

  serverlessv2_scaling_configuration {
    max_capacity = 1.0
    min_capacity = 0.5
  }
}
`, rName, tfrds.ClusterEngineAuroraPostgreSQL, enableHttpEndpoint)
}

func testAccClusterConfig_password(rName, password string) string {
	return fmt.Sprintf(`
resource "aws_rds_cluster" "test" {
  cluster_identifier  = %[1]q
  database_name       = "test"
  master_username     = "tfacctest"
  master_password     = %[2]q
  engine              = %[3]q
  skip_final_snapshot = true
}
`, rName, password, tfrds.ClusterEngineAuroraMySQL)
}

func testAccConfig_ClusterSubnetGroup(rName string) string {
	return acctest.ConfigCompose(
		acctest.ConfigVPCWithSubnets(rName, 3),
		fmt.Sprintf(`
resource "aws_db_subnet_group" "test" {
  name       = %[1]q
  subnet_ids = aws_subnet.test[*].id
}
`, rName),
	)
}

func testAccConfig_ServiceRole(rName string) string {
	return fmt.Sprintf(`
resource "aws_iam_role" "role" {
  name = %[1]q

  assume_role_policy = <<EOF
{
	"Version": "2012-10-17",
	"Statement": [
		{
		"Action": "sts:AssumeRole",
		"Principal": {
			"Service": [
				"directoryservice.rds.${data.aws_partition.current.dns_suffix}",				
				"rds.${data.aws_partition.current.dns_suffix}"
			]
		},
		"Effect": "Allow",
		"Sid": ""
		}
	]
}
EOF
}

resource "aws_iam_role_policy_attachment" "attatch-policy" {
  role       = aws_iam_role.role.name
  policy_arn = "arn:${data.aws_partition.current.partition}:iam::aws:policy/service-role/AmazonRDSDirectoryServiceAccess"
}
`, rName)
}

func testAccConfig_DirectoryService(rName, domain string) string {
	return fmt.Sprintf(`
resource "aws_security_group" "test" {
  name   = %[1]q
  vpc_id = aws_vpc.test.id

  tags = {
    Name = %[1]q
  }
}

resource "aws_security_group_rule" "test" {
  type        = "egress"
  from_port   = 0
  to_port     = 0
  protocol    = "-1"
  cidr_blocks = ["0.0.0.0/0"]

  security_group_id = aws_security_group.test.id
}

resource "aws_directory_service_directory" "directory" {
  name     = %[2]q
  password = "SuperSecretPassw0rd"
  size     = "Small"
  type     = "MicrosoftAD"
  edition  = "Standard"

  vpc_settings {
    vpc_id     = aws_vpc.test.id
    subnet_ids = [aws_subnet.test[0].id, aws_subnet.test[1].id]
  }
}

data "aws_partition" "current" {}
`, rName, domain)
}

func testAccClusterConfig_noDeleteAutomatedBackups(rName, preferredBackupWindow string) string {
	return fmt.Sprintf(`
resource "aws_rds_cluster" "test" {
  cluster_identifier       = %[1]q
  database_name            = "test"
  engine                   = %[2]q
  master_username          = "tfacctest"
  master_password          = "avoid-plaintext-passwords"
  preferred_backup_window  = %[3]q
  skip_final_snapshot      = true
  delete_automated_backups = false
}
`, rName, tfrds.ClusterEngineAuroraMySQL, preferredBackupWindow)
}

func testAccClusterConfig_localWriteForwarding(rName string) string {
	return fmt.Sprintf(`
resource "aws_rds_cluster" "test" {
  cluster_identifier            = %[1]q
  database_name                 = "test"
  engine                        = "aurora-mysql"
  engine_version                = "8.0.mysql_aurora.3.04.0"
  enable_local_write_forwarding = true
  master_username               = "tfacctest"
  master_password               = "avoid-plaintext-passwords"
  skip_final_snapshot           = true
}
`, rName)
}

func testAccClusterConfig_engineLifecycleSupport_disabled(rName string) string {
	return fmt.Sprintf(`
resource "aws_rds_cluster" "test" {
  cluster_identifier       = %[1]q
  database_name            = "test"
  engine                   = %[2]q
  master_username          = "tfacctest"
  master_password          = "avoid-plaintext-passwords"
  skip_final_snapshot      = true
  engine_lifecycle_support = "open-source-rds-extended-support-disabled"
}
`, rName, tfrds.ClusterEngineAuroraPostgreSQL)
}
