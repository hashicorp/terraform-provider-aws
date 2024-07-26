// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package rds_test

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/rds"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfrds "github.com/hashicorp/terraform-provider-aws/internal/service/rds"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestClusterIDRegionFromARN(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		TestName       string
		Input          string
		ExpectedID     string
		ExpectedRegion string
		ExpectedErr    bool
	}{
		{
			TestName:       "empty",
			Input:          "",
			ExpectedID:     "",
			ExpectedRegion: "",
			ExpectedErr:    true,
		},
		{
			TestName:       "normal ARN",
			Input:          "arn:aws:rds:us-west-2:012345678901:cluster:tf-acc-test-1467354933239945971", // lintignore:AWSAT003,AWSAT005
			ExpectedID:     "tf-acc-test-1467354933239945971",
			ExpectedRegion: "us-west-2", // lintignore:AWSAT003
			ExpectedErr:    false,
		},
		{
			TestName:       "another good ARN",
			Input:          "arn:aws:rds:us-east-1:012345678901:cluster:tf-acc-test-1467354933239945971", // lintignore:AWSAT003,AWSAT005
			ExpectedID:     "tf-acc-test-1467354933239945971",
			ExpectedRegion: "us-east-1", // lintignore:AWSAT003
			ExpectedErr:    false,
		},
		{
			TestName:       "no account",
			Input:          "arn:aws:rds:us-east-2::cluster:tf-acc-test-1467354933239945971", // lintignore:AWSAT003,AWSAT005
			ExpectedID:     "tf-acc-test-1467354933239945971",
			ExpectedRegion: "us-east-2", // lintignore:AWSAT003
			ExpectedErr:    false,
		},
		{
			TestName:       "wrong service",
			Input:          "arn:aws:connect:us-west-2:012345678901:instance/1032bdc4-d72c-5490-a9fa-3c9b4dba67bb", // lintignore:AWSAT003,AWSAT005
			ExpectedID:     "",
			ExpectedRegion: "",
			ExpectedErr:    true,
		},
		{
			TestName:       "wrong resource",
			Input:          "arn:aws:rds:us-east-2::notacluster:tf-acc-test-1467354933239945971", // lintignore:AWSAT003,AWSAT005
			ExpectedID:     "",
			ExpectedRegion: "",
			ExpectedErr:    true,
		},
	}

	for _, testCase := range testCases {
		testCase := testCase
		t.Run(testCase.TestName, func(t *testing.T) {
			t.Parallel()

			gotID, gotRegion, gotErr := tfrds.ClusterIDRegionFromARN(testCase.Input)

			if gotErr != nil && !testCase.ExpectedErr {
				t.Errorf("got no error, expected one: %s", testCase.Input)
			}

			if gotID != testCase.ExpectedID {
				t.Errorf("got %s, expected %s", gotID, testCase.ExpectedID)
			}

			if gotRegion != testCase.ExpectedRegion {
				t.Errorf("got %s, expected %s", gotRegion, testCase.ExpectedRegion)
			}
		})
	}
}

func TestAccRDSGlobalCluster_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var globalCluster1 rds.GlobalCluster
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_rds_global_cluster.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheckGlobalCluster(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.RDSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckGlobalClusterDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccGlobalClusterConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGlobalClusterExists(ctx, resourceName, &globalCluster1),
					acctest.CheckResourceAttrGlobalARN(resourceName, names.AttrARN, "rds", fmt.Sprintf("global-cluster:%s", rName)),
					resource.TestCheckResourceAttr(resourceName, names.AttrDatabaseName, ""),
					resource.TestCheckResourceAttr(resourceName, names.AttrDeletionProtection, acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, names.AttrEngine, "aurora-postgresql"),
					resource.TestCheckResourceAttr(resourceName, "engine_lifecycle_support", "open-source-rds-extended-support"),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrEngineVersion),
					resource.TestCheckResourceAttr(resourceName, "global_cluster_identifier", rName),
					resource.TestMatchResourceAttr(resourceName, "global_cluster_resource_id", regexache.MustCompile(`cluster-.+`)),
					resource.TestCheckResourceAttr(resourceName, names.AttrStorageEncrypted, acctest.CtFalse),
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

func TestAccRDSGlobalCluster_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	var globalCluster1 rds.GlobalCluster
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_rds_global_cluster.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheckGlobalCluster(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.RDSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckGlobalClusterDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccGlobalClusterConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGlobalClusterExists(ctx, resourceName, &globalCluster1),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfrds.ResourceGlobalCluster(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccRDSGlobalCluster_databaseName(t *testing.T) {
	ctx := acctest.Context(t)
	var globalCluster1, globalCluster2 rds.GlobalCluster
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_rds_global_cluster.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheckGlobalCluster(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.RDSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckGlobalClusterDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccGlobalClusterConfig_databaseName(rName, "database1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGlobalClusterExists(ctx, resourceName, &globalCluster1),
					resource.TestCheckResourceAttr(resourceName, names.AttrDatabaseName, "database1"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccGlobalClusterConfig_databaseName(rName, "database2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGlobalClusterExists(ctx, resourceName, &globalCluster2),
					testAccCheckGlobalClusterRecreated(&globalCluster1, &globalCluster2),
					resource.TestCheckResourceAttr(resourceName, names.AttrDatabaseName, "database2"),
				),
			},
		},
	})
}

func TestAccRDSGlobalCluster_deletionProtection(t *testing.T) {
	ctx := acctest.Context(t)
	var globalCluster1, globalCluster2 rds.GlobalCluster
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_rds_global_cluster.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheckGlobalCluster(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.RDSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckGlobalClusterDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccGlobalClusterConfig_deletionProtection(rName, true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGlobalClusterExists(ctx, resourceName, &globalCluster1),
					resource.TestCheckResourceAttr(resourceName, names.AttrDeletionProtection, acctest.CtTrue),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccGlobalClusterConfig_deletionProtection(rName, false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGlobalClusterExists(ctx, resourceName, &globalCluster2),
					testAccCheckGlobalClusterNotRecreated(&globalCluster1, &globalCluster2),
					resource.TestCheckResourceAttr(resourceName, names.AttrDeletionProtection, acctest.CtFalse),
				),
			},
		},
	})
}

func TestAccRDSGlobalCluster_engineLifecycleSupport_disabled(t *testing.T) {
	ctx := acctest.Context(t)
	var globalCluster1 rds.GlobalCluster
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_rds_global_cluster.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheckGlobalCluster(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.RDSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckGlobalClusterDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccGlobalClusterConfig_engineLifecycleSupport_disabled(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGlobalClusterExists(ctx, resourceName, &globalCluster1),
					acctest.CheckResourceAttrGlobalARN(resourceName, names.AttrARN, "rds", fmt.Sprintf("global-cluster:%s", rName)),
					resource.TestCheckResourceAttr(resourceName, names.AttrDatabaseName, ""),
					resource.TestCheckResourceAttr(resourceName, names.AttrEngine, "aurora-postgresql"),
					resource.TestCheckResourceAttr(resourceName, "engine_lifecycle_support", "open-source-rds-extended-support-disabled"),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrEngineVersion),
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

func TestAccRDSGlobalCluster_EngineVersion_updateMinor(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var globalCluster1, globalCluster2 rds.GlobalCluster
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_rds_global_cluster.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheckGlobalCluster(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.RDSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckGlobalClusterDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccGlobalClusterConfig_primaryMinorEngineVersionDynamic(rName, tfrds.InstanceEngineAuroraPostgreSQL, false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGlobalClusterExists(ctx, resourceName, &globalCluster1),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrEngineVersion, "data.aws_rds_engine_version.test", "version_actual"),
				),
			},
			{
				Config: testAccGlobalClusterConfig_primaryMinorEngineVersionDynamic(rName, tfrds.InstanceEngineAuroraPostgreSQL, true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGlobalClusterExists(ctx, resourceName, &globalCluster2),
					testAccCheckGlobalClusterNotRecreated(&globalCluster1, &globalCluster2),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrEngineVersion, "data.aws_rds_engine_version.upgrade", "version_actual"),
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

func TestAccRDSGlobalCluster_EngineVersion_updateMajor(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var globalCluster1, globalCluster2 rds.GlobalCluster
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_rds_global_cluster.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheckGlobalCluster(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.RDSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckGlobalClusterDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccGlobalClusterConfig_primaryMajorEngineVersionDynamic(rName, tfrds.InstanceEngineAuroraMySQL, false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGlobalClusterExists(ctx, resourceName, &globalCluster1),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrEngineVersion, "data.aws_rds_engine_version.test", "version_actual"),
				),
			},
			{
				Config: testAccGlobalClusterConfig_primaryMajorEngineVersionDynamic(rName, tfrds.InstanceEngineAuroraMySQL, true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGlobalClusterExists(ctx, resourceName, &globalCluster2),
					testAccCheckGlobalClusterNotRecreated(&globalCluster1, &globalCluster2),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrEngineVersion, "data.aws_rds_engine_version.upgrade", "version_actual"),
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

func TestAccRDSGlobalCluster_EngineVersion_updateMinorMultiRegion(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	// aurora-mysql has issues with versions. Using the versions AWS provides with the data source
	// as compatible for minor version upgrade fails with
	// InvalidParameterValue: In-place minor version upgrade of Aurora MySQL global database cluster 'xyz' to Aurora MySQL engine version 8.0.mysql_aurora.3.05.2 isn't supported. The selected target version 8.0.mysql_aurora.3.05.2 supports a higher version of community MySQL that introduces changes incompatible with previous minor versions of Aurora MySQL. See the Aurora documentation for how to perform a minor version upgrade on global database clusters.

	var globalCluster1, globalCluster2 rds.GlobalCluster
	rNameGlobal := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix) // don't need to be unique but makes debugging easier
	rNamePrimary := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	rNameSecondary := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_rds_global_cluster.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheckGlobalCluster(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.RDSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5FactoriesAlternate(ctx, t),
		CheckDestroy:             testAccCheckGlobalClusterDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccGlobalClusterConfig_engineVersionMinorUpgradeMultiRegionDynamic(rNameGlobal, rNamePrimary, rNameSecondary, tfrds.ClusterEngineAuroraPostgreSQL, false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGlobalClusterExists(ctx, resourceName, &globalCluster1),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrEngineVersion, "data.aws_rds_engine_version.test", "version_actual"),
				),
			},
			{
				Config: testAccGlobalClusterConfig_engineVersionMinorUpgradeMultiRegionDynamic(rNameGlobal, rNamePrimary, rNameSecondary, tfrds.ClusterEngineAuroraPostgreSQL, true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGlobalClusterExists(ctx, resourceName, &globalCluster2),
					testAccCheckGlobalClusterNotRecreated(&globalCluster1, &globalCluster2),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrEngineVersion, "data.aws_rds_engine_version.upgrade", "version_actual"),
				),
			},
		},
	})
}

func TestAccRDSGlobalCluster_EngineVersion_updateMajorMultiRegion(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var globalCluster1, globalCluster2 rds.GlobalCluster
	rNameGlobal := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix) // don't need to be unique but makes debugging easier
	rNamePrimary := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	rNameSecondary := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_rds_global_cluster.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheckGlobalCluster(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.RDSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5FactoriesAlternate(ctx, t),
		CheckDestroy:             testAccCheckGlobalClusterDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccGlobalClusterConfig_engineVersionMajorUpgradeMultiRegionDynamic(rNameGlobal, rNamePrimary, rNameSecondary, tfrds.InstanceEngineAuroraMySQL, false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGlobalClusterExists(ctx, resourceName, &globalCluster1),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrEngineVersion, "data.aws_rds_engine_version.test", "version_actual"),
				),
			},
			{
				Config: testAccGlobalClusterConfig_engineVersionMajorUpgradeMultiRegionDynamic(rNameGlobal, rNamePrimary, rNameSecondary, tfrds.InstanceEngineAuroraMySQL, true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGlobalClusterExists(ctx, resourceName, &globalCluster2),
					testAccCheckGlobalClusterNotRecreated(&globalCluster1, &globalCluster2),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrEngineVersion, "data.aws_rds_engine_version.upgrade", "version_actual"),
				),
			},
		},
	})
}

func TestAccRDSGlobalCluster_EngineVersion_auroraMySQL(t *testing.T) {
	ctx := acctest.Context(t)
	var globalCluster1 rds.GlobalCluster
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_rds_global_cluster.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheckGlobalCluster(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.RDSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckGlobalClusterDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccGlobalClusterConfig_engineVersion(rName, tfrds.InstanceEngineAuroraMySQL),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGlobalClusterExists(ctx, resourceName, &globalCluster1),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrEngineVersion, "data.aws_rds_engine_version.default", names.AttrVersion),
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

func TestAccRDSGlobalCluster_EngineVersion_auroraPostgreSQL(t *testing.T) {
	ctx := acctest.Context(t)
	var globalCluster1 rds.GlobalCluster
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_rds_global_cluster.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheckGlobalCluster(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.RDSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckGlobalClusterDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccGlobalClusterConfig_engineVersion(rName, "aurora-postgresql"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGlobalClusterExists(ctx, resourceName, &globalCluster1),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrEngineVersion, "data.aws_rds_engine_version.default", names.AttrVersion),
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

func TestAccRDSGlobalCluster_forceDestroy(t *testing.T) {
	ctx := acctest.Context(t)
	var globalCluster1 rds.GlobalCluster
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_rds_global_cluster.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheckGlobalCluster(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.RDSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckGlobalClusterDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccGlobalClusterConfig_forceDestroy(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGlobalClusterExists(ctx, resourceName, &globalCluster1),
					resource.TestCheckResourceAttr(resourceName, names.AttrForceDestroy, acctest.CtTrue),
				),
			},
		},
	})
}

func TestAccRDSGlobalCluster_sourceDBClusterIdentifier(t *testing.T) {
	ctx := acctest.Context(t)
	var globalCluster1 rds.GlobalCluster
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	clusterResourceName := "aws_rds_cluster.test"
	resourceName := "aws_rds_global_cluster.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheckGlobalCluster(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.RDSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckGlobalClusterDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccGlobalClusterConfig_sourceClusterID(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGlobalClusterExists(ctx, resourceName, &globalCluster1),
					resource.TestCheckResourceAttrPair(resourceName, "source_db_cluster_identifier", clusterResourceName, names.AttrARN),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{names.AttrForceDestroy, "source_db_cluster_identifier"},
			},
		},
	})
}

func TestAccRDSGlobalCluster_SourceDBClusterIdentifier_storageEncrypted(t *testing.T) {
	ctx := acctest.Context(t)
	var globalCluster1 rds.GlobalCluster
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	clusterResourceName := "aws_rds_cluster.test"
	resourceName := "aws_rds_global_cluster.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheckGlobalCluster(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.RDSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckGlobalClusterDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccGlobalClusterConfig_sourceClusterIDStorageEncrypted(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGlobalClusterExists(ctx, resourceName, &globalCluster1),
					resource.TestCheckResourceAttrPair(resourceName, "source_db_cluster_identifier", clusterResourceName, names.AttrARN),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{names.AttrForceDestroy, "source_db_cluster_identifier"},
			},
		},
	})
}

func TestAccRDSGlobalCluster_storageEncrypted(t *testing.T) {
	ctx := acctest.Context(t)
	var globalCluster1, globalCluster2 rds.GlobalCluster
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_rds_global_cluster.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheckGlobalCluster(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.RDSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckGlobalClusterDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccGlobalClusterConfig_storageEncrypted(rName, true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGlobalClusterExists(ctx, resourceName, &globalCluster1),
					resource.TestCheckResourceAttr(resourceName, names.AttrStorageEncrypted, acctest.CtTrue),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccGlobalClusterConfig_storageEncrypted(rName, false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGlobalClusterExists(ctx, resourceName, &globalCluster2),
					testAccCheckGlobalClusterRecreated(&globalCluster1, &globalCluster2),
					resource.TestCheckResourceAttr(resourceName, names.AttrStorageEncrypted, acctest.CtFalse),
				),
			},
		},
	})
}

func testAccCheckGlobalClusterExists(ctx context.Context, n string, v *rds.GlobalCluster) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).RDSConn(ctx)

		output, err := tfrds.FindGlobalClusterByID(ctx, conn, rs.Primary.ID)

		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

func testAccCheckGlobalClusterDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).RDSConn(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_rds_global_cluster" {
				continue
			}

			_, err := tfrds.FindGlobalClusterByID(ctx, conn, rs.Primary.ID)

			if tfresource.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("RDS Global Cluster %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccCheckGlobalClusterNotRecreated(i, j *rds.GlobalCluster) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if aws.StringValue(i.GlobalClusterArn) != aws.StringValue(j.GlobalClusterArn) {
			return fmt.Errorf("RDS Global Cluster was recreated. got: %s, expected: %s", aws.StringValue(i.GlobalClusterArn), aws.StringValue(j.GlobalClusterArn))
		}

		return nil
	}
}

func testAccCheckGlobalClusterRecreated(i, j *rds.GlobalCluster) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if aws.StringValue(i.GlobalClusterResourceId) == aws.StringValue(j.GlobalClusterResourceId) {
			return errors.New("RDS Global Cluster was not recreated")
		}

		return nil
	}
}

func testAccPreCheckGlobalCluster(ctx context.Context, t *testing.T) {
	conn := acctest.Provider.Meta().(*conns.AWSClient).RDSConn(ctx)

	input := &rds.DescribeGlobalClustersInput{}

	_, err := conn.DescribeGlobalClustersWithContext(ctx, input)

	if acctest.PreCheckSkipError(err) || tfawserr.ErrMessageContains(err, "InvalidParameterValue", "Access Denied to API Version: APIGlobalDatabases") {
		t.Skipf("skipping acceptance testing: %s", err)
	}

	if err != nil {
		t.Fatalf("unexpected PreCheck error: %s", err)
	}
}

func testAccGlobalClusterConfig_basic(rName string) string {
	return fmt.Sprintf(`
resource "aws_rds_global_cluster" "test" {
  global_cluster_identifier = %[1]q
  engine                    = "aurora-postgresql"
}
`, rName)
}

func testAccGlobalClusterConfig_databaseName(rName, databaseName string) string {
	return fmt.Sprintf(`
resource "aws_rds_global_cluster" "test" {
  database_name             = %[1]q
  global_cluster_identifier = %[2]q
  engine                    = "aurora-postgresql"
}
`, databaseName, rName)
}

func testAccGlobalClusterConfig_deletionProtection(rName string, deletionProtection bool) string {
	return fmt.Sprintf(`
resource "aws_rds_global_cluster" "test" {
  deletion_protection       = %[1]t
  global_cluster_identifier = %[2]q
  engine                    = "aurora-postgresql"
}
`, deletionProtection, rName)
}

func testAccGlobalClusterConfig_engineLifecycleSupport_disabled(rName string) string {
	return fmt.Sprintf(`
resource "aws_rds_global_cluster" "test" {
  global_cluster_identifier = %[1]q
  engine                    = "aurora-postgresql"
  engine_lifecycle_support  = "open-source-rds-extended-support-disabled"
}
`, rName)
}

func testAccGlobalClusterConfig_engineVersion(rName, engine string) string {
	return fmt.Sprintf(`
data "aws_rds_engine_version" "default" {
  engine = %[1]q
}

resource "aws_rds_global_cluster" "test" {
  engine                    = data.aws_rds_engine_version.default.engine
  engine_version            = data.aws_rds_engine_version.default.version
  global_cluster_identifier = %[2]q
}
`, engine, rName)
}

func testAccGlobalClusterConfig_forceDestroy(rName string) string {
	return fmt.Sprintf(`
resource "aws_rds_global_cluster" "test" {
  global_cluster_identifier = %[1]q
  engine                    = "aurora-postgresql"
  force_destroy             = true
}
`, rName)
}

func testAccGlobalClusterConfig_primaryMajorEngineVersionDynamic(rName, engine string, upgrade bool) string {
	return fmt.Sprintf(`
data "aws_rds_engine_version" "test" {
  engine           = %[1]q
  has_major_target = true
  latest           = true
}

data "aws_rds_orderable_db_instance" "test" {
  engine                     = data.aws_rds_engine_version.test.engine
  engine_version             = data.aws_rds_engine_version.test.version_actual
  preferred_instance_classes = [%[2]s]
  supports_clusters          = true
  supports_global_databases  = true
}

data "aws_rds_engine_version" "upgrade" {
  engine             = data.aws_rds_engine_version.test.engine
  latest             = true
  preferred_versions = data.aws_rds_engine_version.test.valid_major_targets
}

data "aws_rds_orderable_db_instance" "upgrade" {
  engine                     = data.aws_rds_engine_version.test.engine
  engine_version             = data.aws_rds_engine_version.upgrade.version_actual
  preferred_instance_classes = [%[2]s]
  supports_clusters          = true
  supports_global_databases  = true
}

locals {
  engine_version = %[3]t ? data.aws_rds_engine_version.upgrade.version_actual : data.aws_rds_engine_version.test.version_actual
}

resource "aws_rds_global_cluster" "test" {
  engine                    = data.aws_rds_orderable_db_instance.test.engine
  engine_version            = local.engine_version
  global_cluster_identifier = %[4]q
}

resource "aws_rds_cluster" "test" {
  apply_immediately           = true
  allow_major_version_upgrade = true
  cluster_identifier          = %[4]q
  engine                      = data.aws_rds_orderable_db_instance.test.engine
  engine_version              = local.engine_version
  master_password             = "mustbeeightcharacters"
  master_username             = "test"
  skip_final_snapshot         = true

  global_cluster_identifier = aws_rds_global_cluster.test.id

  lifecycle {
    ignore_changes = [global_cluster_identifier]
  }
}

resource "aws_rds_cluster_instance" "test" {
  apply_immediately  = true
  cluster_identifier = aws_rds_cluster.test.id
  engine             = data.aws_rds_orderable_db_instance.test.engine
  engine_version     = local.engine_version
  identifier         = %[4]q
  instance_class     = data.aws_rds_orderable_db_instance.test.instance_class

  lifecycle {
    ignore_changes = [engine_version]
  }
}
`, engine, mainInstanceClasses, upgrade, rName)
}

func testAccGlobalClusterConfig_primaryMinorEngineVersionDynamic(rName, engine string, upgrade bool) string {
	return fmt.Sprintf(`
data "aws_rds_engine_version" "test" {
  engine           = %[1]q
  has_minor_target = true
  latest           = true
}

data "aws_rds_orderable_db_instance" "test" {
  engine                     = data.aws_rds_engine_version.test.engine
  engine_version             = data.aws_rds_engine_version.test.version_actual
  preferred_instance_classes = [%[2]s]
  supports_clusters          = true
  supports_global_databases  = true
}

data "aws_rds_engine_version" "upgrade" {
  engine             = data.aws_rds_engine_version.test.engine
  latest             = true
  preferred_versions = data.aws_rds_engine_version.test.valid_minor_targets
}

data "aws_rds_orderable_db_instance" "upgrade" {
  engine                     = data.aws_rds_engine_version.test.engine
  engine_version             = data.aws_rds_engine_version.upgrade.version_actual
  preferred_instance_classes = [%[2]s]
  supports_clusters          = true
  supports_global_databases  = true
}

locals {
  engine_version = %[3]t ? data.aws_rds_engine_version.upgrade.version_actual : data.aws_rds_engine_version.test.version_actual
}

resource "aws_rds_global_cluster" "test" {
  engine                    = data.aws_rds_orderable_db_instance.test.engine
  engine_version            = local.engine_version
  global_cluster_identifier = %[4]q
}

resource "aws_rds_cluster" "test" {
  apply_immediately           = true
  allow_major_version_upgrade = true
  cluster_identifier          = %[4]q
  engine                      = data.aws_rds_orderable_db_instance.test.engine
  engine_version              = local.engine_version
  master_password             = "mustbeeightcharacters"
  master_username             = "test"
  skip_final_snapshot         = true

  global_cluster_identifier = aws_rds_global_cluster.test.id

  lifecycle {
    ignore_changes = [global_cluster_identifier]
  }
}

resource "aws_rds_cluster_instance" "test" {
  apply_immediately  = true
  cluster_identifier = aws_rds_cluster.test.id
  engine             = data.aws_rds_orderable_db_instance.test.engine
  engine_version     = local.engine_version
  identifier         = %[4]q
  instance_class     = data.aws_rds_orderable_db_instance.upgrade.instance_class

  lifecycle {
    ignore_changes = [engine_version]
  }
}
`, engine, mainInstanceClasses, upgrade, rName)
}

func testAccGlobalClusterConfig_engineVersionMajorUpgradeMultiRegionDynamic(rNameGlobal, rNamePrimary, rNameSecondary, engine string, upgrade bool) string {
	return acctest.ConfigCompose(acctest.ConfigMultipleRegionProvider(2), fmt.Sprintf(`
data "aws_region" "current" {}

data "aws_availability_zones" "alternate" {
  provider = "awsalternate"
  state    = "available"

  filter {
    name   = "opt-in-status"
    values = ["opt-in-not-required"]
  }
}

data "aws_rds_engine_version" "test" {
  engine           = %[1]q
  has_major_target = true
  latest           = true
}

data "aws_rds_orderable_db_instance" "test" {
  engine                     = data.aws_rds_engine_version.test.engine
  engine_version             = data.aws_rds_engine_version.test.version_actual
  preferred_instance_classes = [%[2]s]
  supports_clusters          = true
  supports_global_databases  = true
}

data "aws_rds_engine_version" "upgrade" {
  engine             = data.aws_rds_engine_version.test.engine
  latest             = true
  preferred_versions = data.aws_rds_engine_version.test.valid_major_targets
}

data "aws_rds_orderable_db_instance" "upgrade" {
  engine                     = data.aws_rds_engine_version.test.engine
  engine_version             = data.aws_rds_engine_version.upgrade.version_actual
  preferred_instance_classes = [%[2]s]
  supports_clusters          = true
  supports_global_databases  = true
}

locals {
  engine_version = %[3]t ? data.aws_rds_orderable_db_instance.upgrade.engine_version : data.aws_rds_engine_version.test.version
  instance_class = %[3]t ? data.aws_rds_orderable_db_instance.upgrade.instance_class : data.aws_rds_orderable_db_instance.test.instance_class
}

resource "aws_rds_global_cluster" "test" {
  global_cluster_identifier = %[4]q
  engine                    = data.aws_rds_orderable_db_instance.upgrade.engine
  engine_version            = local.engine_version
}

resource "aws_rds_cluster" "primary" {
  allow_major_version_upgrade = true
  apply_immediately           = true
  cluster_identifier          = %[5]q
  engine                      = aws_rds_global_cluster.test.engine
  engine_version              = aws_rds_global_cluster.test.engine_version
  global_cluster_identifier   = aws_rds_global_cluster.test.id
  master_password             = "avoid-plaintext-passwords"
  master_username             = "tfacctest"
  skip_final_snapshot         = true

  lifecycle {
    ignore_changes = [engine_version]
  }
}

resource "aws_rds_cluster_instance" "primary" {
  apply_immediately  = true
  cluster_identifier = aws_rds_cluster.primary.id
  engine             = aws_rds_cluster.primary.engine
  engine_version     = aws_rds_cluster.primary.engine_version
  identifier         = %[5]q
  instance_class     = local.instance_class
}

resource "aws_vpc" "alternate" {
  provider   = "awsalternate"
  cidr_block = "10.0.0.0/16"

  tags = {
    Name = %[6]q
  }
}

resource "aws_subnet" "alternate" {
  provider          = "awsalternate"
  count             = 3
  vpc_id            = aws_vpc.alternate.id
  availability_zone = data.aws_availability_zones.alternate.names[count.index]
  cidr_block        = "10.0.${count.index}.0/24"

  tags = {
    Name = %[6]q
  }
}

resource "aws_db_subnet_group" "alternate" {
  provider   = "awsalternate"
  name       = %[6]q
  subnet_ids = aws_subnet.alternate[*].id
}

resource "aws_rds_cluster" "secondary" {
  provider                    = "awsalternate"
  allow_major_version_upgrade = true
  apply_immediately           = true
  cluster_identifier          = %[6]q
  engine                      = aws_rds_global_cluster.test.engine
  engine_version              = aws_rds_global_cluster.test.engine_version
  global_cluster_identifier   = aws_rds_global_cluster.test.id
  skip_final_snapshot         = true

  lifecycle {
    ignore_changes = [
      replication_source_identifier,
      engine_version,
    ]
  }

  depends_on = [aws_rds_cluster_instance.primary]
}

resource "aws_rds_cluster_instance" "secondary" {
  provider           = "awsalternate"
  apply_immediately  = true
  cluster_identifier = aws_rds_cluster.secondary.id
  engine             = aws_rds_cluster.secondary.engine
  engine_version     = aws_rds_cluster.secondary.engine_version
  identifier         = %[6]q
  instance_class     = local.instance_class
}
`, engine, mainInstanceClasses, upgrade, rNameGlobal, rNamePrimary, rNameSecondary))
}

func testAccGlobalClusterConfig_engineVersionMinorUpgradeMultiRegionDynamic(rNameGlobal, rNamePrimary, rNameSecondary, engine string, upgrade bool) string {
	return acctest.ConfigCompose(acctest.ConfigMultipleRegionProvider(2), fmt.Sprintf(`
data "aws_region" "current" {}

data "aws_availability_zones" "alternate" {
  provider = "awsalternate"
  state    = "available"

  filter {
    name   = "opt-in-status"
    values = ["opt-in-not-required"]
  }
}

data "aws_rds_engine_version" "test" {
  engine           = %[1]q
  has_minor_target = true
  latest           = true
}

data "aws_rds_orderable_db_instance" "test" {
  engine                     = data.aws_rds_engine_version.test.engine
  engine_version             = data.aws_rds_engine_version.test.version_actual
  preferred_instance_classes = [%[2]s]
  supports_clusters          = true
  supports_global_databases  = true
}

data "aws_rds_engine_version" "upgrade" {
  engine             = data.aws_rds_engine_version.test.engine
  latest             = true
  preferred_versions = data.aws_rds_engine_version.test.valid_minor_targets
}

data "aws_rds_orderable_db_instance" "upgrade" {
  engine                     = data.aws_rds_engine_version.test.engine
  engine_version             = data.aws_rds_engine_version.upgrade.version_actual
  preferred_instance_classes = [%[2]s]
  supports_clusters          = true
  supports_global_databases  = true
}

locals {
  engine_version = %[3]t ? data.aws_rds_engine_version.upgrade.version_actual : data.aws_rds_engine_version.test.version_actual
}

resource "aws_rds_global_cluster" "test" {
  global_cluster_identifier = %[4]q
  engine                    = data.aws_rds_orderable_db_instance.upgrade.engine
  engine_version            = local.engine_version
}

resource "aws_rds_cluster" "primary" {
  allow_major_version_upgrade = true
  apply_immediately           = true
  cluster_identifier          = %[5]q
  engine                      = aws_rds_global_cluster.test.engine
  engine_version              = local.engine_version
  global_cluster_identifier   = aws_rds_global_cluster.test.id
  master_password             = "avoid-plaintext-passwords"
  master_username             = "tfacctest"
  skip_final_snapshot         = true

  lifecycle {
    ignore_changes = [engine_version]
  }
}

resource "aws_rds_cluster_instance" "primary" {
  apply_immediately  = true
  cluster_identifier = aws_rds_cluster.primary.id
  engine             = aws_rds_cluster.primary.engine
  engine_version     = local.engine_version
  identifier         = %[5]q
  instance_class     = data.aws_rds_orderable_db_instance.upgrade.instance_class
}

resource "aws_vpc" "alternate" {
  provider   = "awsalternate"
  cidr_block = "10.0.0.0/16"

  tags = {
    Name = %[6]q
  }
}

resource "aws_subnet" "alternate" {
  provider          = "awsalternate"
  count             = 3
  vpc_id            = aws_vpc.alternate.id
  availability_zone = data.aws_availability_zones.alternate.names[count.index]
  cidr_block        = "10.0.${count.index}.0/24"

  tags = {
    Name = %[6]q
  }
}

resource "aws_db_subnet_group" "alternate" {
  provider   = "awsalternate"
  name       = %[6]q
  subnet_ids = aws_subnet.alternate[*].id
}

resource "aws_rds_cluster" "secondary" {
  provider                    = "awsalternate"
  allow_major_version_upgrade = true
  apply_immediately           = true
  cluster_identifier          = %[6]q
  engine                      = aws_rds_global_cluster.test.engine
  engine_version              = local.engine_version
  global_cluster_identifier   = aws_rds_global_cluster.test.id
  skip_final_snapshot         = true

  lifecycle {
    ignore_changes = [
      replication_source_identifier,
      engine_version,
    ]
  }

  depends_on = [aws_rds_cluster_instance.primary]
}

resource "aws_rds_cluster_instance" "secondary" {
  provider           = "awsalternate"
  apply_immediately  = true
  cluster_identifier = aws_rds_cluster.secondary.id
  engine             = aws_rds_cluster.secondary.engine
  engine_version     = local.engine_version
  identifier         = %[6]q
  instance_class     = data.aws_rds_orderable_db_instance.upgrade.instance_class
}
`, engine, mainInstanceClasses, upgrade, rNameGlobal, rNamePrimary, rNameSecondary))
}

func testAccGlobalClusterConfig_sourceClusterID(rName string) string {
	return fmt.Sprintf(`
data "aws_rds_engine_version" "default" {
  engine = "aurora-postgresql"
}

resource "aws_rds_cluster" "test" {
  cluster_identifier  = %[1]q
  engine              = data.aws_rds_engine_version.default.engine
  engine_version      = data.aws_rds_engine_version.default.version
  master_password     = "mustbeeightcharacters"
  master_username     = "test"
  skip_final_snapshot = true

  # global_cluster_identifier cannot be Computed

  lifecycle {
    ignore_changes = [global_cluster_identifier]
  }
}

resource "aws_rds_global_cluster" "test" {
  force_destroy                = true
  global_cluster_identifier    = %[1]q
  source_db_cluster_identifier = aws_rds_cluster.test.arn
}
`, rName)
}

func testAccGlobalClusterConfig_sourceClusterIDStorageEncrypted(rName string) string {
	return fmt.Sprintf(`
data "aws_rds_engine_version" "default" {
  engine = "aurora-postgresql"
}

resource "aws_rds_cluster" "test" {
  cluster_identifier  = %[1]q
  engine              = data.aws_rds_engine_version.default.engine
  engine_version      = data.aws_rds_engine_version.default.version
  master_password     = "mustbeeightcharacters"
  master_username     = "test"
  skip_final_snapshot = true
  storage_encrypted   = true

  # global_cluster_identifier cannot be Computed

  lifecycle {
    ignore_changes = [global_cluster_identifier]
  }
}

resource "aws_rds_global_cluster" "test" {
  force_destroy                = true
  global_cluster_identifier    = %[1]q
  source_db_cluster_identifier = aws_rds_cluster.test.arn
}
`, rName)
}

func testAccGlobalClusterConfig_storageEncrypted(rName string, storageEncrypted bool) string {
	return fmt.Sprintf(`
resource "aws_rds_global_cluster" "test" {
  global_cluster_identifier = %[1]q
  storage_encrypted         = %[2]t
  engine                    = "aurora-postgresql"
}
`, rName, storageEncrypted)
}
