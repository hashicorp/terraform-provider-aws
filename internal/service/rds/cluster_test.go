package rds_test

import (
	"errors"
	"fmt"
	"log"
	"regexp"
	"strings"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/rds"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
)

func init() {
	acctest.RegisterServiceErrorCheckFunc(rds.EndpointsID, testAccErrorCheckSkip)

}

func testAccErrorCheckSkip(t *testing.T) resource.ErrorCheckFunc {
	return acctest.ErrorCheckSkipMessagesContaining(t,
		"engine mode serverless you requested is currently unavailable",
		"engine mode multimaster you requested is currently unavailable",
		"requested engine version was not found or does not support parallelquery functionality",
		"Backtrack is not enabled for the aurora engine",
		"Read replica DB clusters are not available in this region for engine aurora",
	)
}

func TestAccRDSCluster_basic(t *testing.T) {
	var dbCluster rds.DBCluster
	clusterName := sdkacctest.RandomWithPrefix("tf-aurora-cluster")
	resourceName := "aws_rds_cluster.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, rds.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckClusterDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccClusterConfig(clusterName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterExists(resourceName, &dbCluster),
					acctest.CheckResourceAttrRegionalARN(resourceName, "arn", "rds", fmt.Sprintf("cluster:%s", clusterName)),
					resource.TestCheckResourceAttr(resourceName, "backtrack_window", "0"),
					resource.TestCheckResourceAttr(resourceName, "copy_tags_to_snapshot", "false"),
					resource.TestCheckResourceAttr(resourceName, "storage_encrypted", "false"),
					resource.TestCheckResourceAttr(resourceName, "db_cluster_parameter_group_name", "default.aurora5.6"),
					resource.TestCheckResourceAttrSet(resourceName, "reader_endpoint"),
					resource.TestCheckResourceAttrSet(resourceName, "cluster_resource_id"),
					resource.TestCheckResourceAttr(resourceName, "engine", "aurora"),
					resource.TestCheckResourceAttrSet(resourceName, "engine_version"),
					resource.TestCheckResourceAttr(resourceName, "global_cluster_identifier", ""),
					resource.TestCheckResourceAttrSet(resourceName, "hosted_zone_id"),
					resource.TestCheckResourceAttr(resourceName, "enabled_cloudwatch_logs_exports.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "scaling_configuration.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
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
					"db_instance_parameter_group_name",
					"enable_global_write_forwarding",
					"master_password",
					"skip_final_snapshot",
				},
			},
		},
	})
}

func TestAccRDSCluster_allowMajorVersionUpgrade(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var dbCluster1, dbCluster2 rds.DBCluster
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_rds_cluster.test"
	// If these hardcoded versions become a maintenance burden, use DescribeDBEngineVersions
	// either by having a new data source created or implementing the testing similar
	// to TestAccDMSReplicationInstance_engineVersion
	engine := "aurora-postgresql"
	engineVersion1 := "12.9"
	engineVersion2 := "13.5"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, rds.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckClusterDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccClusterConfig_AllowMajorVersionUpgrade(rName, true, engine, engineVersion1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterExists(resourceName, &dbCluster1),
					resource.TestCheckResourceAttr(resourceName, "allow_major_version_upgrade", "true"),
					resource.TestCheckResourceAttr(resourceName, "engine", engine),
					resource.TestCheckResourceAttr(resourceName, "engine_version", engineVersion1),
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
					"db_instance_parameter_group_name",
					"enable_global_write_forwarding",
					"master_password",
					"skip_final_snapshot",
				},
			},
			{
				Config: testAccClusterConfig_AllowMajorVersionUpgrade(rName, true, engine, engineVersion2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterExists(resourceName, &dbCluster2),
					resource.TestCheckResourceAttr(resourceName, "allow_major_version_upgrade", "true"),
					resource.TestCheckResourceAttr(resourceName, "engine", engine),
					resource.TestCheckResourceAttr(resourceName, "engine_version", engineVersion2),
				),
			},
		},
	})
}

func TestAccRDSCluster_allowMajorVersionUpgradeWithCustomParametersApplyImm(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var dbCluster1, dbCluster2 rds.DBCluster
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_rds_cluster.test"
	// If these hardcoded versions become a maintenance burden, use DescribeDBEngineVersions
	// either by having a new data source created or implementing the testing similar
	// to TestAccDMSReplicationInstance_engineVersion
	engine := "aurora-postgresql"
	engineVersion1 := "12.9"
	engineVersion2 := "13.5"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, rds.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckClusterDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccClusterConfig_AllowMajorVersionUpgradeWithCustomParameters(rName, true, engine, engineVersion1, true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterExists(resourceName, &dbCluster1),
					resource.TestCheckResourceAttr(resourceName, "allow_major_version_upgrade", "true"),
					resource.TestCheckResourceAttr(resourceName, "engine", engine),
					resource.TestCheckResourceAttr(resourceName, "engine_version", engineVersion1),
				),
			},
			{
				Config: testAccClusterConfig_AllowMajorVersionUpgradeWithCustomParameters(rName, true, engine, engineVersion2, true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterExists(resourceName, &dbCluster2),
					resource.TestCheckResourceAttr(resourceName, "allow_major_version_upgrade", "true"),
					resource.TestCheckResourceAttr(resourceName, "engine", engine),
					resource.TestCheckResourceAttr(resourceName, "engine_version", engineVersion2),
				),
			},
		},
	})
}

func TestAccRDSCluster_allowMajorVersionUpgradeWithCustomParameters(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var dbCluster1, dbCluster2 rds.DBCluster
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_rds_cluster.test"
	// If these hardcoded versions become a maintenance burden, use DescribeDBEngineVersions
	// either by having a new data source created or implementing the testing similar
	// to TestAccDMSReplicationInstance_engineVersion
	engine := "aurora-postgresql"
	engineVersion1 := "12.9"
	engineVersion2 := "13.5"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, rds.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckClusterDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccClusterConfig_AllowMajorVersionUpgradeWithCustomParameters(rName, true, engine, engineVersion1, false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterExists(resourceName, &dbCluster1),
					resource.TestCheckResourceAttr(resourceName, "allow_major_version_upgrade", "true"),
					resource.TestCheckResourceAttr(resourceName, "engine", engine),
					resource.TestCheckResourceAttr(resourceName, "engine_version", engineVersion1),
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
					"db_instance_parameter_group_name",
					"enable_global_write_forwarding",
					"master_password",
					"skip_final_snapshot",
				},
			},
			{
				Config: testAccClusterConfig_AllowMajorVersionUpgradeWithCustomParameters(rName, true, engine, engineVersion2, false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterExists(resourceName, &dbCluster2),
					resource.TestCheckResourceAttr(resourceName, "allow_major_version_upgrade", "true"),
					resource.TestCheckResourceAttr(resourceName, "engine", engine),
					resource.TestCheckResourceAttr(resourceName, "engine_version", engineVersion2),
				),
			},
		},
	})
}

func TestAccRDSCluster_onlyMajorVersion(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var dbCluster1 rds.DBCluster
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_rds_cluster.test"
	// If these hardcoded versions become a maintenance burden, use DescribeDBEngineVersions
	// either by having a new data source created or implementing the testing similar
	// to TestAccDMSReplicationInstance_engineVersion
	engine := "aurora-postgresql"
	engineVersion1 := "11"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, rds.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckClusterDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccClusterConfig_MajorVersionOnly(rName, false, engine, engineVersion1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterExists(resourceName, &dbCluster1),
					resource.TestCheckResourceAttr(resourceName, "engine", engine),
					resource.TestCheckResourceAttr(resourceName, "engine_version", engineVersion1),
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
					"db_instance_parameter_group_name",
					"enable_global_write_forwarding",
					"engine_version",
					"master_password",
					"skip_final_snapshot",
				},
			},
		},
	})
}

func TestAccRDSCluster_availabilityZones(t *testing.T) {
	var dbCluster rds.DBCluster
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_rds_cluster.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, rds.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckClusterDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccClusterConfig_AvailabilityZones(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterExists(resourceName, &dbCluster),
				),
			},
		},
	})
}

func TestAccRDSCluster_storageType(t *testing.T) {
	var dbCluster rds.DBCluster
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_rds_cluster.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, rds.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckClusterDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccClusterConfig_StorageType(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterExists(resourceName, &dbCluster),
					resource.TestCheckResourceAttr(resourceName, "storage_type", "io1"),
				),
			},
		},
	})
}

func TestAccRDSCluster_allocatedStorage(t *testing.T) {
	var dbCluster rds.DBCluster
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_rds_cluster.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, rds.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckClusterDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccClusterConfig_AllocatedStorage(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterExists(resourceName, &dbCluster),
					resource.TestCheckResourceAttr(resourceName, "allocated_storage", "100"),
				),
			},
		},
	})
}

func TestAccRDSCluster_iops(t *testing.T) {
	var dbCluster rds.DBCluster
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_rds_cluster.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, rds.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckClusterDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccClusterConfig_Iops(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterExists(resourceName, &dbCluster),
					resource.TestCheckResourceAttr(resourceName, "iops", "1000"),
				),
			},
		},
	})
}

func TestAccRDSCluster_dbClusterInstanceClass(t *testing.T) {
	var dbCluster rds.DBCluster
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_rds_cluster.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, rds.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckClusterDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccClusterConfig_DbClusterInstanceClass(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterExists(resourceName, &dbCluster),
					resource.TestCheckResourceAttr(resourceName, "db_cluster_instance_class", "db.r6gd.xlarge"),
				),
			},
		},
	})
}

func TestAccRDSCluster_backtrackWindow(t *testing.T) {
	var dbCluster rds.DBCluster
	resourceName := "aws_rds_cluster.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, rds.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckClusterDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccClusterConfig_BacktrackWindow(43200),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterExists(resourceName, &dbCluster),
					resource.TestCheckResourceAttr(resourceName, "backtrack_window", "43200"),
				),
			},
			{
				Config: testAccClusterConfig_BacktrackWindow(86400),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterExists(resourceName, &dbCluster),
					resource.TestCheckResourceAttr(resourceName, "backtrack_window", "86400"),
				),
			},
		},
	})
}

func TestAccRDSCluster_clusterIdentifierPrefix(t *testing.T) {
	var v rds.DBCluster
	resourceName := "aws_rds_cluster.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, rds.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckClusterDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccClusterConfig_ClusterIdentifierPrefix("tf-test-"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterExists(resourceName, &v),
					resource.TestMatchResourceAttr(
						resourceName, "cluster_identifier", regexp.MustCompile("^tf-test-")),
				),
			},
		},
	})
}

func TestAccRDSCluster_dbSubnetGroupName(t *testing.T) {
	var dbCluster rds.DBCluster
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_rds_cluster.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, rds.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckClusterDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccClusterConfig_SubnetGroupName(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterExists(resourceName, &dbCluster),
				),
			},
		},
	})
}

func TestAccRDSCluster_s3Restore(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var v rds.DBCluster
	resourceName := "aws_rds_cluster.test"
	bucket := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	uniqueId := sdkacctest.RandomWithPrefix("tf-acc-s3-import-test")
	bucketPrefix := sdkacctest.RandString(5)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, rds.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckClusterDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccClusterConfig_s3Restore(bucket, bucketPrefix, uniqueId),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterExists(resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "engine", "aurora"),
				),
			},
		},
	})
}

func TestAccRDSCluster_pointInTimeRestore(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var v rds.DBCluster
	var c rds.DBCluster

	parentId := sdkacctest.RandomWithPrefix("tf-acc-point-in-time-restore-seed-test")
	restoredId := sdkacctest.RandomWithPrefix("tf-acc-point-in-time-restored-test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, rds.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckClusterDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccClusterConfig_pointInTimeRestoreSource(parentId, restoredId),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterExists("aws_rds_cluster.test", &v),
					testAccCheckClusterExists("aws_rds_cluster.restored_pit", &c),
					resource.TestCheckResourceAttr("aws_rds_cluster.restored_pit", "cluster_identifier", restoredId),
					resource.TestCheckResourceAttrPair("aws_rds_cluster.restored_pit", "engine", "aws_rds_cluster.test", "engine"),
				),
			},
		},
	})
}

func TestAccRDSCluster_PointInTimeRestore_enabledCloudWatchLogsExports(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var v rds.DBCluster
	var c rds.DBCluster

	parentId := sdkacctest.RandomWithPrefix("tf-acc-point-in-time-restore-seed-test")
	restoredId := sdkacctest.RandomWithPrefix("tf-acc-point-in-time-restored-test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, rds.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckClusterDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccClusterConfig_pointInTimeRestoreSource_enabled_cloudWatch_logs_exports(parentId, restoredId, "audit"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterExists("aws_rds_cluster.test", &v),
					testAccCheckClusterExists("aws_rds_cluster.restored_pit", &c),
					resource.TestCheckResourceAttr("aws_rds_cluster.restored_pit", "cluster_identifier", restoredId),
					resource.TestCheckResourceAttrPair("aws_rds_cluster.restored_pit", "engine", "aws_rds_cluster.test", "engine"),
					resource.TestCheckResourceAttr("aws_rds_cluster.restored_pit", "enabled_cloudwatch_logs_exports.#", "1"),
					resource.TestCheckTypeSetElemAttr("aws_rds_cluster.restored_pit", "enabled_cloudwatch_logs_exports.*", "audit"),
				),
			},
		},
	})
}

func TestAccRDSCluster_generatedName(t *testing.T) {
	var v rds.DBCluster
	resourceName := "aws_rds_cluster.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, rds.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckClusterDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccClusterConfig_generatedName(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterExists(resourceName, &v),
					resource.TestMatchResourceAttr(
						resourceName, "cluster_identifier", regexp.MustCompile("^tf-")),
				),
			},
		},
	})
}

func TestAccRDSCluster_takeFinalSnapshot(t *testing.T) {
	var v rds.DBCluster
	rInt := sdkacctest.RandInt()
	resourceName := "aws_rds_cluster.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, rds.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckClusterSnapshot(rInt),
		Steps: []resource.TestStep{
			{
				Config: testAccClusterWithFinalSnapshotConfig(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterExists(resourceName, &v),
				),
			},
		},
	})
}

// This is a regression test to make sure that we always cover the scenario as highlighted in
// https://github.com/hashicorp/terraform/issues/11568
// Expected error updated to match API response
func TestAccRDSCluster_missingUserNameCausesError(t *testing.T) {
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, rds.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckClusterDestroy,
		Steps: []resource.TestStep{
			{
				Config:      testAccClusterWithoutUserNameAndPasswordConfig(sdkacctest.RandInt()),
				ExpectError: regexp.MustCompile(`InvalidParameterValue: The parameter MasterUsername must be provided`),
			},
		},
	})
}

func TestAccRDSCluster_tags(t *testing.T) {
	var dbCluster1, dbCluster2, dbCluster3 rds.DBCluster
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_rds_cluster.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, rds.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckClusterDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccClusterTags1Config(rName, "key1", "value1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterExists(resourceName, &dbCluster1),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1"),
				),
			},
			{
				Config: testAccClusterTags2Config(rName, "key1", "value1updated", "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterExists(resourceName, &dbCluster2),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1updated"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
			{
				Config: testAccClusterTags1Config(rName, "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterExists(resourceName, &dbCluster3),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
		},
	})
}

func TestAccRDSCluster_EnabledCloudWatchLogsExports_mySQL(t *testing.T) {
	var dbCluster1, dbCluster2, dbCluster3 rds.DBCluster
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_rds_cluster.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, rds.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckClusterDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccClusterEnabledCloudWatchLogsExports1Config(rName, "audit"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterExists(resourceName, &dbCluster1),
					resource.TestCheckResourceAttr(resourceName, "enabled_cloudwatch_logs_exports.#", "1"),
					resource.TestCheckTypeSetElemAttr(resourceName, "enabled_cloudwatch_logs_exports.*", "audit"),
				),
			},
			{
				Config: testAccClusterEnabledCloudWatchLogsExports2Config(rName, "slowquery", "error"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterExists(resourceName, &dbCluster2),
					resource.TestCheckResourceAttr(resourceName, "enabled_cloudwatch_logs_exports.#", "2"),
					resource.TestCheckTypeSetElemAttr(resourceName, "enabled_cloudwatch_logs_exports.*", "error"),
					resource.TestCheckTypeSetElemAttr(resourceName, "enabled_cloudwatch_logs_exports.*", "slowquery"),
				),
			},
			{
				Config: testAccClusterEnabledCloudWatchLogsExports1Config(rName, "error"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterExists(resourceName, &dbCluster3),
					resource.TestCheckResourceAttr(resourceName, "enabled_cloudwatch_logs_exports.#", "1"),
					resource.TestCheckTypeSetElemAttr(resourceName, "enabled_cloudwatch_logs_exports.*", "error"),
				),
			},
		},
	})
}

func TestAccRDSCluster_EnabledCloudWatchLogsExports_postgresql(t *testing.T) {
	var dbCluster1 rds.DBCluster
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_rds_cluster.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, rds.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckClusterDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccClusterEnabledCloudWatchLogsExportsPostgres1Config(rName, "postgresql"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterExists(resourceName, &dbCluster1),
					resource.TestCheckResourceAttr(resourceName, "enabled_cloudwatch_logs_exports.#", "1"),
					resource.TestCheckTypeSetElemAttr(resourceName, "enabled_cloudwatch_logs_exports.*", "postgresql"),
				),
			},
		},
	})
}

func TestAccRDSCluster_updateIAMRoles(t *testing.T) {
	var v rds.DBCluster
	ri := sdkacctest.RandInt()
	resourceName := "aws_rds_cluster.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, rds.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckClusterDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccClusterIncludingIAMRolesConfig(ri),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterExists(resourceName, &v),
				),
			},
			{
				Config: testAccClusterAddIAMRolesConfig(ri),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterExists(resourceName, &v),
					resource.TestCheckResourceAttr(
						resourceName, "iam_roles.#", "2"),
				),
			},
			{
				Config: testAccClusterRemoveIAMRolesConfig(ri),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterExists(resourceName, &v),
					resource.TestCheckResourceAttr(
						resourceName, "iam_roles.#", "1"),
				),
			},
		},
	})
}

func TestAccRDSCluster_kmsKey(t *testing.T) {
	var dbCluster1 rds.DBCluster
	kmsKeyResourceName := "aws_kms_key.foo"
	resourceName := "aws_rds_cluster.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, rds.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckClusterDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccClusterConfig_kmsKey(sdkacctest.RandInt()),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterExists(resourceName, &dbCluster1),
					resource.TestCheckResourceAttrPair(resourceName, "kms_key_id", kmsKeyResourceName, "arn"),
				),
			},
		},
	})
}

func TestAccRDSCluster_encrypted(t *testing.T) {
	var v rds.DBCluster
	resourceName := "aws_rds_cluster.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, rds.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckClusterDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccClusterConfig_encrypted(sdkacctest.RandInt()),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterExists(resourceName, &v),
					resource.TestCheckResourceAttr(
						resourceName, "storage_encrypted", "true"),
					resource.TestCheckResourceAttr(
						resourceName, "db_cluster_parameter_group_name", "default.aurora5.6"),
				),
			},
		},
	})
}

func TestAccRDSCluster_copyTagsToSnapshot(t *testing.T) {
	var v rds.DBCluster
	rInt := sdkacctest.RandInt()
	resourceName := "aws_rds_cluster.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, rds.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckClusterDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccClusterWithCopyTagsToSnapshotConfig(rInt, true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterExists(resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "copy_tags_to_snapshot", "true"),
				),
			},
			{
				Config: testAccClusterWithCopyTagsToSnapshotConfig(rInt, false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterExists(resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "copy_tags_to_snapshot", "false"),
				),
			},
			{
				Config: testAccClusterWithCopyTagsToSnapshotConfig(rInt, true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterExists(resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "copy_tags_to_snapshot", "true"),
				),
			},
		},
	})
}

func TestAccRDSCluster_ReplicationSourceIdentifier_kmsKeyID(t *testing.T) {
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
			acctest.PreCheck(t)
			acctest.PreCheckMultipleRegion(t, 2)
		},
		ErrorCheck:        acctest.ErrorCheck(t, rds.EndpointsID),
		ProviderFactories: acctest.FactoriesAlternate(&providers),
		CheckDestroy:      acctest.CheckWithProviders(testAccCheckClusterDestroyWithProvider, &providers),
		Steps: []resource.TestStep{
			{
				Config: testAccClusterReplicationSourceIdentifierKMSKeyIDConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterExistsWithProvider(resourceName, &primaryCluster, acctest.RegionProviderFunc(acctest.Region(), &providers)),
					testAccCheckClusterExistsWithProvider(resourceName2, &replicaCluster, acctest.RegionProviderFunc(acctest.AlternateRegion(), &providers)),
				),
			},
		},
	})
}

func TestAccRDSCluster_backupsUpdate(t *testing.T) {
	var v rds.DBCluster
	resourceName := "aws_rds_cluster.test"

	ri := sdkacctest.RandInt()
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, rds.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckClusterDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccClusterConfig_backups(ri),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterExists(resourceName, &v),
					resource.TestCheckResourceAttr(
						resourceName, "preferred_backup_window", "07:00-09:00"),
					resource.TestCheckResourceAttr(
						resourceName, "backup_retention_period", "5"),
					resource.TestCheckResourceAttr(
						resourceName, "preferred_maintenance_window", "tue:04:00-tue:04:30"),
				),
			},
			{
				Config: testAccClusterConfig_backupsUpdate(ri),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterExists(resourceName, &v),
					resource.TestCheckResourceAttr(
						resourceName, "preferred_backup_window", "03:00-09:00"),
					resource.TestCheckResourceAttr(
						resourceName, "backup_retention_period", "10"),
					resource.TestCheckResourceAttr(
						resourceName, "preferred_maintenance_window", "wed:01:00-wed:01:30"),
				),
			},
		},
	})
}

func TestAccRDSCluster_iamAuth(t *testing.T) {
	var v rds.DBCluster
	resourceName := "aws_rds_cluster.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, rds.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckClusterDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccClusterConfig_iamAuth(sdkacctest.RandInt()),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterExists(resourceName, &v),
					resource.TestCheckResourceAttr(
						resourceName, "iam_database_authentication_enabled", "true"),
				),
			},
		},
	})
}

func TestAccRDSCluster_deletionProtection(t *testing.T) {
	var dbCluster1 rds.DBCluster
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_rds_cluster.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, rds.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckClusterDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccClusterConfig_DeletionProtection(rName, true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterExists(resourceName, &dbCluster1),
					resource.TestCheckResourceAttr(resourceName, "deletion_protection", "true"),
				),
			},
			{
				Config: testAccClusterConfig_DeletionProtection(rName, false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterExists(resourceName, &dbCluster1),
					resource.TestCheckResourceAttr(resourceName, "deletion_protection", "false"),
				),
			},
		},
	})
}

func TestAccRDSCluster_engineMode(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var dbCluster1, dbCluster2 rds.DBCluster

	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_rds_cluster.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, rds.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckClusterDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccClusterConfig_EngineMode(rName, "serverless"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterExists(resourceName, &dbCluster1),
					resource.TestCheckResourceAttr(resourceName, "engine_mode", "serverless"),
					resource.TestCheckResourceAttr(resourceName, "scaling_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "serverlessv2_scaling_configuration.#", "0"),
				),
			},
			{
				Config: testAccClusterConfig_EngineMode(rName, "provisioned"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterExists(resourceName, &dbCluster2),
					testAccCheckClusterRecreated(&dbCluster1, &dbCluster2),
					resource.TestCheckResourceAttr(resourceName, "engine_mode", "provisioned"),
					resource.TestCheckResourceAttr(resourceName, "scaling_configuration.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "serverlessv2_scaling_configuration.#", "0"),
				),
			},
		},
	})
}

func TestAccRDSCluster_EngineMode_global(t *testing.T) {
	var dbCluster1 rds.DBCluster

	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_rds_cluster.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheckGlobalCluster(t) },
		ErrorCheck:        acctest.ErrorCheck(t, rds.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckClusterDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccClusterConfig_EngineMode_Global(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterExists(resourceName, &dbCluster1),
					resource.TestCheckResourceAttr(resourceName, "engine_mode", "global"),
				),
			},
		},
	})
}

func TestAccRDSCluster_EngineMode_multiMaster(t *testing.T) {
	var dbCluster1 rds.DBCluster

	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_rds_cluster.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, rds.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckClusterDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccClusterConfig_EngineMode_Multimaster(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterExists(resourceName, &dbCluster1),
					resource.TestCheckResourceAttr(resourceName, "engine_mode", "multimaster"),
				),
			},
		},
	})
}

func TestAccRDSCluster_EngineMode_parallelQuery(t *testing.T) {
	var dbCluster1 rds.DBCluster

	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_rds_cluster.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, rds.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckClusterDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccClusterConfig_EngineMode(rName, "parallelquery"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterExists(resourceName, &dbCluster1),
					resource.TestCheckResourceAttr(resourceName, "engine_mode", "parallelquery"),
				),
			},
		},
	})
}

func TestAccRDSCluster_engineVersion(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var dbCluster rds.DBCluster
	rInt := sdkacctest.RandInt()
	resourceName := "aws_rds_cluster.test"
	dataSourceName := "data.aws_rds_engine_version.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, rds.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckClusterDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccClusterConfig_EngineVersion(false, rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterExists(resourceName, &dbCluster),
					resource.TestCheckResourceAttr(resourceName, "engine", "aurora-postgresql"),
					resource.TestCheckResourceAttrPair(resourceName, "engine_version", dataSourceName, "version"),
				),
			},
			{
				Config:      testAccClusterConfig_EngineVersion(true, rInt),
				ExpectError: regexp.MustCompile(`Cannot modify engine version without a healthy primary instance in DB cluster`),
			},
		},
	})
}

func TestAccRDSCluster_engineVersionWithPrimaryInstance(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var dbCluster rds.DBCluster
	rInt := sdkacctest.RandInt()
	resourceName := "aws_rds_cluster.test"
	dataSourceName := "data.aws_rds_engine_version.test"
	dataSourceNameUpgrade := "data.aws_rds_engine_version.upgrade"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, rds.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckClusterDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccClusterConfig_EngineVersionWithPrimaryInstance(false, rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterExists(resourceName, &dbCluster),
					resource.TestCheckResourceAttrPair(resourceName, "engine", dataSourceName, "engine"),
					resource.TestCheckResourceAttrPair(resourceName, "engine_version", dataSourceName, "version"),
				),
			},
			{
				Config: testAccClusterConfig_EngineVersionWithPrimaryInstance(true, rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterExists(resourceName, &dbCluster),
					resource.TestCheckResourceAttrPair(resourceName, "engine", dataSourceNameUpgrade, "engine"),
					resource.TestCheckResourceAttrPair(resourceName, "engine_version", dataSourceNameUpgrade, "version"),
				),
			},
		},
	})
}

func TestAccRDSCluster_GlobalClusterIdentifierEngineMode_global(t *testing.T) {
	var dbCluster1 rds.DBCluster

	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	globalClusterResourceName := "aws_rds_global_cluster.test"
	resourceName := "aws_rds_cluster.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheckGlobalCluster(t) },
		ErrorCheck:        acctest.ErrorCheck(t, rds.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckClusterDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccClusterConfig_GlobalClusterIdentifier_EngineMode_Global(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterExists(resourceName, &dbCluster1),
					resource.TestCheckResourceAttrPair(resourceName, "global_cluster_identifier", globalClusterResourceName, "id"),
				),
			},
		},
	})
}

func TestAccRDSCluster_GlobalClusterIdentifierEngineModeGlobal_add(t *testing.T) {
	var dbCluster1 rds.DBCluster

	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_rds_cluster.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheckGlobalCluster(t) },
		ErrorCheck:        acctest.ErrorCheck(t, rds.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckClusterDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccClusterConfig_EngineMode_Global(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterExists(resourceName, &dbCluster1),
					resource.TestCheckResourceAttr(resourceName, "global_cluster_identifier", ""),
				),
			},
			{
				Config:      testAccClusterConfig_GlobalClusterIdentifier_EngineMode_Global(rName),
				ExpectError: regexp.MustCompile(`Existing RDS Clusters cannot be added to an existing RDS Global Cluster`),
			},
		},
	})
}

func TestAccRDSCluster_GlobalClusterIdentifierEngineModeGlobal_remove(t *testing.T) {
	var dbCluster1 rds.DBCluster

	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	globalClusterResourceName := "aws_rds_global_cluster.test"
	resourceName := "aws_rds_cluster.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheckGlobalCluster(t) },
		ErrorCheck:        acctest.ErrorCheck(t, rds.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckClusterDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccClusterConfig_GlobalClusterIdentifier_EngineMode_Global(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterExists(resourceName, &dbCluster1),
					resource.TestCheckResourceAttrPair(resourceName, "global_cluster_identifier", globalClusterResourceName, "id"),
				),
			},
			{
				Config: testAccClusterConfig_EngineMode_Global(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterExists(resourceName, &dbCluster1),
					resource.TestCheckResourceAttr(resourceName, "global_cluster_identifier", ""),
				),
			},
		},
	})
}

func TestAccRDSCluster_GlobalClusterIdentifierEngineModeGlobal_update(t *testing.T) {
	var dbCluster1 rds.DBCluster

	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	globalClusterResourceName1 := "aws_rds_global_cluster.test.0"
	globalClusterResourceName2 := "aws_rds_global_cluster.test.1"
	resourceName := "aws_rds_cluster.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheckGlobalCluster(t) },
		ErrorCheck:        acctest.ErrorCheck(t, rds.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckClusterDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccClusterConfig_GlobalClusterIdentifier_EngineMode_Global_Update(rName, globalClusterResourceName1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterExists(resourceName, &dbCluster1),
					resource.TestCheckResourceAttrPair(resourceName, "global_cluster_identifier", globalClusterResourceName1, "id"),
				),
			},
			{
				Config:      testAccClusterConfig_GlobalClusterIdentifier_EngineMode_Global_Update(rName, globalClusterResourceName2),
				ExpectError: regexp.MustCompile(`Existing RDS Clusters cannot be migrated between existing RDS Global Clusters`),
			},
		},
	})
}

func TestAccRDSCluster_GlobalClusterIdentifierEngineMode_provisioned(t *testing.T) {
	var dbCluster1 rds.DBCluster

	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	globalClusterResourceName := "aws_rds_global_cluster.test"
	resourceName := "aws_rds_cluster.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheckGlobalCluster(t) },
		ErrorCheck:        acctest.ErrorCheck(t, rds.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckClusterDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccClusterConfig_GlobalClusterIdentifier_EngineMode_Provisioned(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterExists(resourceName, &dbCluster1),
					resource.TestCheckResourceAttrPair(resourceName, "global_cluster_identifier", globalClusterResourceName, "id"),
				),
			},
		},
	})
}

// Reference: https://github.com/hashicorp/terraform-provider-aws/issues/13126
func TestAccRDSCluster_GlobalClusterIdentifier_primarySecondaryClusters(t *testing.T) {
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
			acctest.PreCheck(t)
			acctest.PreCheckMultipleRegion(t, 2)
			testAccPreCheckGlobalCluster(t)
		},
		ErrorCheck:        acctest.ErrorCheck(t, rds.EndpointsID),
		ProviderFactories: acctest.FactoriesAlternate(&providers),
		CheckDestroy:      testAccCheckClusterDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccClusterConfig_GlobalClusterIdentifier_PrimarySecondaryClusters(rNameGlobal, rNamePrimary, rNameSecondary),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterExistsWithProvider(resourceNamePrimary, &primaryDbCluster, acctest.RegionProviderFunc(acctest.Region(), &providers)),
					testAccCheckClusterExistsWithProvider(resourceNameSecondary, &secondaryDbCluster, acctest.RegionProviderFunc(acctest.AlternateRegion(), &providers)),
				),
			},
		},
	})
}

// Reference: https://github.com/hashicorp/terraform-provider-aws/issues/13715
func TestAccRDSCluster_GlobalClusterIdentifier_replicationSourceIdentifier(t *testing.T) {
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
			acctest.PreCheck(t)
			acctest.PreCheckMultipleRegion(t, 2)
			testAccPreCheckGlobalCluster(t)
		},
		ErrorCheck:        acctest.ErrorCheck(t, rds.EndpointsID),
		ProviderFactories: acctest.FactoriesAlternate(&providers),
		CheckDestroy:      testAccCheckClusterDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccClusterConfig_GlobalClusterIdentifier_ReplicationSourceIdentifier(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterExistsWithProvider(resourceNamePrimary, &primaryDbCluster, acctest.RegionProviderFunc(acctest.Region(), &providers)),
					testAccCheckClusterExistsWithProvider(resourceNameSecondary, &secondaryDbCluster, acctest.RegionProviderFunc(acctest.AlternateRegion(), &providers)),
				),
			},
		},
	})
}

// Reference: https://github.com/hashicorp/terraform-provider-aws/issues/14457
func TestAccRDSCluster_GlobalClusterIdentifier_secondaryClustersWriteForwarding(t *testing.T) {
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
			acctest.PreCheck(t)
			acctest.PreCheckMultipleRegion(t, 2)
			testAccPreCheckGlobalCluster(t)
		},
		ErrorCheck:        acctest.ErrorCheck(t, rds.EndpointsID),
		ProviderFactories: acctest.FactoriesAlternate(&providers),
		CheckDestroy:      testAccCheckClusterDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccClusterConfig_GlobalClusterIdentifier_SecondaryClustersWriteForwarding(rNameGlobal, rNamePrimary, rNameSecondary),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterExistsWithProvider(resourceNamePrimary, &primaryDbCluster, acctest.RegionProviderFunc(acctest.Region(), &providers)),
					testAccCheckClusterExistsWithProvider(resourceNameSecondary, &secondaryDbCluster, acctest.RegionProviderFunc(acctest.AlternateRegion(), &providers)),
					resource.TestCheckResourceAttr(resourceNameSecondary, "enable_global_write_forwarding", "true"),
				),
			},
		},
	})
}

func TestAccRDSCluster_port(t *testing.T) {
	var dbCluster1, dbCluster2 rds.DBCluster
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_rds_cluster.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, rds.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckClusterDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccClusterConfig_Port(rName, 5432),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterExists(resourceName, &dbCluster1),
					resource.TestCheckResourceAttr(resourceName, "port", "5432"),
				),
			},
			{
				Config: testAccClusterConfig_Port(rName, 2345),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterExists(resourceName, &dbCluster2),
					resource.TestCheckResourceAttr(resourceName, "port", "2345"),
				),
			},
		},
	})
}

func TestAccRDSCluster_scaling(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var dbCluster rds.DBCluster

	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_rds_cluster.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, rds.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckClusterDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccClusterConfig_ScalingConfiguration(rName, false, 128, 4, 301, "RollbackCapacityChange"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterExists(resourceName, &dbCluster),
					resource.TestCheckResourceAttr(resourceName, "scaling_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "scaling_configuration.0.auto_pause", "false"),
					resource.TestCheckResourceAttr(resourceName, "scaling_configuration.0.max_capacity", "128"),
					resource.TestCheckResourceAttr(resourceName, "scaling_configuration.0.min_capacity", "4"),
					resource.TestCheckResourceAttr(resourceName, "scaling_configuration.0.seconds_until_auto_pause", "301"),
					resource.TestCheckResourceAttr(resourceName, "scaling_configuration.0.timeout_action", "RollbackCapacityChange"),
				),
			},
			{
				Config: testAccClusterConfig_ScalingConfiguration(rName, true, 256, 8, 86400, "ForceApplyCapacityChange"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterExists(resourceName, &dbCluster),
					resource.TestCheckResourceAttr(resourceName, "scaling_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "scaling_configuration.0.auto_pause", "true"),
					resource.TestCheckResourceAttr(resourceName, "scaling_configuration.0.max_capacity", "256"),
					resource.TestCheckResourceAttr(resourceName, "scaling_configuration.0.min_capacity", "8"),
					resource.TestCheckResourceAttr(resourceName, "scaling_configuration.0.seconds_until_auto_pause", "86400"),
					resource.TestCheckResourceAttr(resourceName, "scaling_configuration.0.timeout_action", "ForceApplyCapacityChange"),
				),
			},
		},
	})
}

func TestAccRDSCluster_serverlessV2ScalingConfiguration(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var dbCluster rds.DBCluster

	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_rds_cluster.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, rds.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckClusterDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccClusterServerlessV2ScalingConfigurationConfig(rName, 64.0, 0.5),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterExists(resourceName, &dbCluster),
					resource.TestCheckResourceAttr(resourceName, "serverlessv2_scaling_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "serverlessv2_scaling_configuration.0.max_capacity", "64"),
					resource.TestCheckResourceAttr(resourceName, "serverlessv2_scaling_configuration.0.min_capacity", "0.5"),
				),
			},
			{
				Config: testAccClusterServerlessV2ScalingConfigurationConfig(rName, 128.0, 8.5),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterExists(resourceName, &dbCluster),
					resource.TestCheckResourceAttr(resourceName, "serverlessv2_scaling_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "serverlessv2_scaling_configuration.0.max_capacity", "128"),
					resource.TestCheckResourceAttr(resourceName, "serverlessv2_scaling_configuration.0.min_capacity", "8.5"),
				),
			},
		},
	})
}

// Reference: https://github.com/hashicorp/terraform-provider-aws/issues/11698
func TestAccRDSCluster_Scaling_defaultMinCapacity(t *testing.T) {
	var dbCluster rds.DBCluster

	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_rds_cluster.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, rds.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckClusterDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccClusterConfig_ScalingConfiguration_DefaultMinCapacity(rName, false, 128, 301, "RollbackCapacityChange"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterExists(resourceName, &dbCluster),
					resource.TestCheckResourceAttr(resourceName, "scaling_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "scaling_configuration.0.auto_pause", "false"),
					resource.TestCheckResourceAttr(resourceName, "scaling_configuration.0.max_capacity", "128"),
					resource.TestCheckResourceAttr(resourceName, "scaling_configuration.0.min_capacity", "1"),
					resource.TestCheckResourceAttr(resourceName, "scaling_configuration.0.seconds_until_auto_pause", "301"),
					resource.TestCheckResourceAttr(resourceName, "scaling_configuration.0.timeout_action", "RollbackCapacityChange"),
				),
			},
		},
	})
}

func TestAccRDSCluster_snapshotIdentifier(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var dbCluster, sourceDbCluster rds.DBCluster
	var dbClusterSnapshot rds.DBClusterSnapshot

	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	sourceDbResourceName := "aws_rds_cluster.source"
	snapshotResourceName := "aws_db_cluster_snapshot.test"
	resourceName := "aws_rds_cluster.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, rds.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckInstanceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccClusterConfig_SnapshotIdentifier(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterExists(sourceDbResourceName, &sourceDbCluster),
					testAccCheckDbClusterSnapshotExists(snapshotResourceName, &dbClusterSnapshot),
					testAccCheckClusterExists(resourceName, &dbCluster),
				),
			},
		},
	})
}

func TestAccRDSCluster_SnapshotIdentifier_deletionProtection(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var dbCluster, sourceDbCluster rds.DBCluster
	var dbClusterSnapshot rds.DBClusterSnapshot

	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	sourceDbResourceName := "aws_rds_cluster.source"
	snapshotResourceName := "aws_db_cluster_snapshot.test"
	resourceName := "aws_rds_cluster.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, rds.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckInstanceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccClusterConfig_SnapshotIdentifier_DeletionProtection(rName, true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterExists(sourceDbResourceName, &sourceDbCluster),
					testAccCheckDbClusterSnapshotExists(snapshotResourceName, &dbClusterSnapshot),
					testAccCheckClusterExists(resourceName, &dbCluster),
					resource.TestCheckResourceAttr(resourceName, "deletion_protection", "true"),
				),
			},
			// Ensure we disable deletion protection before attempting to delete :)
			{
				Config: testAccClusterConfig_SnapshotIdentifier_DeletionProtection(rName, false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterExists(sourceDbResourceName, &sourceDbCluster),
					testAccCheckDbClusterSnapshotExists(snapshotResourceName, &dbClusterSnapshot),
					testAccCheckClusterExists(resourceName, &dbCluster),
					resource.TestCheckResourceAttr(resourceName, "deletion_protection", "false"),
				),
			},
		},
	})
}

func TestAccRDSCluster_SnapshotIdentifierEngineMode_parallelQuery(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var dbCluster, sourceDbCluster rds.DBCluster
	var dbClusterSnapshot rds.DBClusterSnapshot

	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	sourceDbResourceName := "aws_rds_cluster.source"
	snapshotResourceName := "aws_db_cluster_snapshot.test"
	resourceName := "aws_rds_cluster.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, rds.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckInstanceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccClusterConfig_SnapshotIdentifier_EngineMode(rName, "parallelquery"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterExists(sourceDbResourceName, &sourceDbCluster),
					testAccCheckDbClusterSnapshotExists(snapshotResourceName, &dbClusterSnapshot),
					testAccCheckClusterExists(resourceName, &dbCluster),
					resource.TestCheckResourceAttr(resourceName, "engine_mode", "parallelquery"),
				),
			},
		},
	})
}

func TestAccRDSCluster_SnapshotIdentifierEngineMode_provisioned(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var dbCluster, sourceDbCluster rds.DBCluster
	var dbClusterSnapshot rds.DBClusterSnapshot

	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	sourceDbResourceName := "aws_rds_cluster.source"
	snapshotResourceName := "aws_db_cluster_snapshot.test"
	resourceName := "aws_rds_cluster.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, rds.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckInstanceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccClusterConfig_SnapshotIdentifier_EngineMode(rName, "provisioned"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterExists(sourceDbResourceName, &sourceDbCluster),
					testAccCheckDbClusterSnapshotExists(snapshotResourceName, &dbClusterSnapshot),
					testAccCheckClusterExists(resourceName, &dbCluster),
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

	var dbCluster, sourceDbCluster rds.DBCluster
	var dbClusterSnapshot rds.DBClusterSnapshot

	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	sourceDbResourceName := "aws_rds_cluster.source"
	snapshotResourceName := "aws_db_cluster_snapshot.test"
	resourceName := "aws_rds_cluster.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, rds.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckInstanceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccClusterConfig_SnapshotIdentifier_EngineMode(rName, "serverless"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterExists(sourceDbResourceName, &sourceDbCluster),
					testAccCheckDbClusterSnapshotExists(snapshotResourceName, &dbClusterSnapshot),
					testAccCheckClusterExists(resourceName, &dbCluster),
					resource.TestCheckResourceAttr(resourceName, "engine_mode", "serverless"),
				),
			},
		},
	})
}

// Reference: https://github.com/hashicorp/terraform-provider-aws/issues/6157
func TestAccRDSCluster_SnapshotIdentifierEngineVersion_different(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var dbCluster, sourceDbCluster rds.DBCluster
	var dbClusterSnapshot rds.DBClusterSnapshot

	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	sourceDbResourceName := "aws_rds_cluster.source"
	snapshotResourceName := "aws_db_cluster_snapshot.test"
	resourceName := "aws_rds_cluster.test"
	dataSourceName := "data.aws_rds_engine_version.upgrade"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, rds.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckInstanceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccClusterConfig_SnapshotIdentifier_EngineVersion(rName, false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterExists(sourceDbResourceName, &sourceDbCluster),
					testAccCheckDbClusterSnapshotExists(snapshotResourceName, &dbClusterSnapshot),
					testAccCheckClusterExists(resourceName, &dbCluster),
					resource.TestCheckResourceAttrPair(resourceName, "engine_version", dataSourceName, "version"),
				),
			},
		},
	})
}

// Reference: https://github.com/hashicorp/terraform-provider-aws/issues/6157
func TestAccRDSCluster_SnapshotIdentifierEngineVersion_equal(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var dbCluster, sourceDbCluster rds.DBCluster
	var dbClusterSnapshot rds.DBClusterSnapshot

	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	sourceDbResourceName := "aws_rds_cluster.source"
	snapshotResourceName := "aws_db_cluster_snapshot.test"
	resourceName := "aws_rds_cluster.test"
	dataSourceName := "data.aws_rds_engine_version.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, rds.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckInstanceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccClusterConfig_SnapshotIdentifier_EngineVersion(rName, true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterExists(sourceDbResourceName, &sourceDbCluster),
					testAccCheckDbClusterSnapshotExists(snapshotResourceName, &dbClusterSnapshot),
					testAccCheckClusterExists(resourceName, &dbCluster),
					resource.TestCheckResourceAttrPair(resourceName, "engine_version", dataSourceName, "version"),
				),
			},
		},
	})
}

func TestAccRDSCluster_SnapshotIdentifier_kmsKeyID(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var dbCluster, sourceDbCluster rds.DBCluster
	var dbClusterSnapshot rds.DBClusterSnapshot

	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	kmsKeyResourceName := "aws_kms_key.test"
	sourceDbResourceName := "aws_rds_cluster.source"
	snapshotResourceName := "aws_db_cluster_snapshot.test"
	resourceName := "aws_rds_cluster.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, rds.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckInstanceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccClusterConfig_SnapshotIdentifier_KMSKeyID(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterExists(sourceDbResourceName, &sourceDbCluster),
					testAccCheckDbClusterSnapshotExists(snapshotResourceName, &dbClusterSnapshot),
					testAccCheckClusterExists(resourceName, &dbCluster),
					resource.TestCheckResourceAttrPair(resourceName, "kms_key_id", kmsKeyResourceName, "arn"),
				),
			},
		},
	})
}

func TestAccRDSCluster_SnapshotIdentifier_masterPassword(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var dbCluster, sourceDbCluster rds.DBCluster
	var dbClusterSnapshot rds.DBClusterSnapshot

	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	sourceDbResourceName := "aws_rds_cluster.source"
	snapshotResourceName := "aws_db_cluster_snapshot.test"
	resourceName := "aws_rds_cluster.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, rds.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckInstanceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccClusterConfig_SnapshotIdentifier_MasterPassword(rName, "password1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterExists(sourceDbResourceName, &sourceDbCluster),
					testAccCheckDbClusterSnapshotExists(snapshotResourceName, &dbClusterSnapshot),
					testAccCheckClusterExists(resourceName, &dbCluster),
					resource.TestCheckResourceAttr(resourceName, "master_password", "password1"),
				),
			},
		},
	})
}

func TestAccRDSCluster_SnapshotIdentifier_masterUsername(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var dbCluster, sourceDbCluster rds.DBCluster
	var dbClusterSnapshot rds.DBClusterSnapshot

	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	sourceDbResourceName := "aws_rds_cluster.source"
	snapshotResourceName := "aws_db_cluster_snapshot.test"
	resourceName := "aws_rds_cluster.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, rds.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckInstanceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccClusterConfig_SnapshotIdentifier_MasterUsername(rName, "username1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterExists(sourceDbResourceName, &sourceDbCluster),
					testAccCheckDbClusterSnapshotExists(snapshotResourceName, &dbClusterSnapshot),
					testAccCheckClusterExists(resourceName, &dbCluster),
					resource.TestCheckResourceAttr(resourceName, "master_username", "foo"),
				),
				// It is not currently possible to update the master username in the RDS API
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccRDSCluster_SnapshotIdentifier_preferredBackupWindow(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var dbCluster, sourceDbCluster rds.DBCluster
	var dbClusterSnapshot rds.DBClusterSnapshot

	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	sourceDbResourceName := "aws_rds_cluster.source"
	snapshotResourceName := "aws_db_cluster_snapshot.test"
	resourceName := "aws_rds_cluster.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, rds.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckInstanceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccClusterConfig_SnapshotIdentifier_PreferredBackupWindow(rName, "00:00-08:00"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterExists(sourceDbResourceName, &sourceDbCluster),
					testAccCheckDbClusterSnapshotExists(snapshotResourceName, &dbClusterSnapshot),
					testAccCheckClusterExists(resourceName, &dbCluster),
					resource.TestCheckResourceAttr(resourceName, "preferred_backup_window", "00:00-08:00"),
				),
			},
		},
	})
}

func TestAccRDSCluster_SnapshotIdentifier_preferredMaintenanceWindow(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var dbCluster, sourceDbCluster rds.DBCluster
	var dbClusterSnapshot rds.DBClusterSnapshot

	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	sourceDbResourceName := "aws_rds_cluster.source"
	snapshotResourceName := "aws_db_cluster_snapshot.test"
	resourceName := "aws_rds_cluster.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, rds.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckInstanceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccClusterConfig_SnapshotIdentifier_PreferredMaintenanceWindow(rName, "sun:01:00-sun:01:30"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterExists(sourceDbResourceName, &sourceDbCluster),
					testAccCheckDbClusterSnapshotExists(snapshotResourceName, &dbClusterSnapshot),
					testAccCheckClusterExists(resourceName, &dbCluster),
					resource.TestCheckResourceAttr(resourceName, "preferred_maintenance_window", "sun:01:00-sun:01:30"),
				),
			},
		},
	})
}

func TestAccRDSCluster_SnapshotIdentifier_tags(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var dbCluster, sourceDbCluster rds.DBCluster
	var dbClusterSnapshot rds.DBClusterSnapshot

	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	sourceDbResourceName := "aws_rds_cluster.source"
	snapshotResourceName := "aws_db_cluster_snapshot.test"
	resourceName := "aws_rds_cluster.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, rds.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckInstanceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccClusterConfig_SnapshotIdentifier_Tags(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterExists(sourceDbResourceName, &sourceDbCluster),
					testAccCheckDbClusterSnapshotExists(snapshotResourceName, &dbClusterSnapshot),
					testAccCheckClusterExists(resourceName, &dbCluster),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1"),
				),
			},
		},
	})
}

func TestAccRDSCluster_SnapshotIdentifier_vpcSecurityGroupIDs(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var dbCluster, sourceDbCluster rds.DBCluster
	var dbClusterSnapshot rds.DBClusterSnapshot

	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	sourceDbResourceName := "aws_rds_cluster.source"
	snapshotResourceName := "aws_db_cluster_snapshot.test"
	resourceName := "aws_rds_cluster.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, rds.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckInstanceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccClusterConfig_SnapshotIdentifier_VPCSecurityGroupIDs(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterExists(sourceDbResourceName, &sourceDbCluster),
					testAccCheckDbClusterSnapshotExists(snapshotResourceName, &dbClusterSnapshot),
					testAccCheckClusterExists(resourceName, &dbCluster),
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
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var dbCluster, sourceDbCluster rds.DBCluster
	var dbClusterSnapshot rds.DBClusterSnapshot

	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	sourceDbResourceName := "aws_rds_cluster.source"
	snapshotResourceName := "aws_db_cluster_snapshot.test"
	resourceName := "aws_rds_cluster.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, rds.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckInstanceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccClusterConfig_SnapshotIdentifier_VPCSecurityGroupIds_Tags(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterExists(sourceDbResourceName, &sourceDbCluster),
					testAccCheckDbClusterSnapshotExists(snapshotResourceName, &dbClusterSnapshot),
					testAccCheckClusterExists(resourceName, &dbCluster),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1"),
				),
			},
		},
	})
}

func TestAccRDSCluster_SnapshotIdentifier_encryptedRestore(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var dbCluster, sourceDbCluster rds.DBCluster
	var dbClusterSnapshot rds.DBClusterSnapshot

	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	kmsKeyResourceName := "aws_kms_key.test"
	sourceDbResourceName := "aws_rds_cluster.source"
	snapshotResourceName := "aws_db_cluster_snapshot.test"
	resourceName := "aws_rds_cluster.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, rds.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckInstanceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccClusterConfig_SnapshotIdentifier_EncryptedRestore(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterExists(sourceDbResourceName, &sourceDbCluster),
					testAccCheckDbClusterSnapshotExists(snapshotResourceName, &dbClusterSnapshot),
					testAccCheckClusterExists(resourceName, &dbCluster),
					resource.TestCheckResourceAttrPair(resourceName, "kms_key_id", kmsKeyResourceName, "arn"),
					resource.TestCheckResourceAttr(resourceName, "storage_encrypted", "true"),
				),
			},
		},
	})
}

func TestAccRDSCluster_enableHTTPEndpoint(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var dbCluster rds.DBCluster

	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_rds_cluster.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, rds.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckClusterDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccClusterConfig_EnableHTTPEndpoint(rName, true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterExists(resourceName, &dbCluster),
					resource.TestCheckResourceAttr(resourceName, "enable_http_endpoint", "true"),
				),
			},
			{
				Config: testAccClusterConfig_EnableHTTPEndpoint(rName, false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterExists(resourceName, &dbCluster),
					resource.TestCheckResourceAttr(resourceName, "enable_http_endpoint", "false"),
				),
			},
		},
	})
}

func testAccCheckClusterDestroy(s *terraform.State) error {
	return testAccCheckClusterDestroyWithProvider(s, acctest.Provider)
}

func testAccCheckClusterDestroyWithProvider(s *terraform.State, provider *schema.Provider) error {
	conn := provider.Meta().(*conns.AWSClient).RDSConn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_rds_cluster" {
			continue
		}

		// Try to find the Group
		var err error
		resp, err := conn.DescribeDBClusters(
			&rds.DescribeDBClustersInput{
				DBClusterIdentifier: aws.String(rs.Primary.ID),
			})

		if err == nil {
			if len(resp.DBClusters) != 0 &&
				*resp.DBClusters[0].DBClusterIdentifier == rs.Primary.ID {
				return fmt.Errorf("DB Cluster %s still exists", rs.Primary.ID)
			}
		}

		if tfawserr.ErrCodeEquals(err, rds.ErrCodeDBClusterNotFoundFault) {
			continue
		}

		return err
	}

	return nil
}

func testAccCheckClusterSnapshot(rInt int) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_rds_cluster" {
				continue
			}

			// Try and delete the snapshot before we check for the cluster not found
			snapshot_identifier := fmt.Sprintf("tf-acctest-rdscluster-snapshot-%d", rInt)

			awsClient := acctest.Provider.Meta().(*conns.AWSClient)
			conn := awsClient.RDSConn

			log.Printf("[INFO] Deleting the Snapshot %s", snapshot_identifier)
			_, snapDeleteErr := conn.DeleteDBClusterSnapshot(
				&rds.DeleteDBClusterSnapshotInput{
					DBClusterSnapshotIdentifier: aws.String(snapshot_identifier),
				})
			if snapDeleteErr != nil {
				return snapDeleteErr
			}

			// Try to find the Group
			var err error
			resp, err := conn.DescribeDBClusters(
				&rds.DescribeDBClustersInput{
					DBClusterIdentifier: aws.String(rs.Primary.ID),
				})

			if tfawserr.ErrCodeEquals(err, rds.ErrCodeDBClusterNotFoundFault) {
				continue
			}

			if err == nil {
				if len(resp.DBClusters) != 0 &&
					*resp.DBClusters[0].DBClusterIdentifier == rs.Primary.ID {
					return fmt.Errorf("DB Cluster %s still exists", rs.Primary.ID)
				}
			}

			return err
		}

		return nil
	}
}

func testAccCheckClusterExists(n string, v *rds.DBCluster) resource.TestCheckFunc {
	return testAccCheckClusterExistsWithProvider(n, v, func() *schema.Provider { return acctest.Provider })
}

func testAccCheckClusterExistsWithProvider(n string, v *rds.DBCluster, providerF func() *schema.Provider) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No DB Instance ID is set")
		}

		provider := providerF()
		conn := provider.Meta().(*conns.AWSClient).RDSConn
		resp, err := conn.DescribeDBClusters(&rds.DescribeDBClustersInput{
			DBClusterIdentifier: aws.String(rs.Primary.ID),
		})

		if err != nil {
			return err
		}

		for _, c := range resp.DBClusters {
			if *c.DBClusterIdentifier == rs.Primary.ID {
				*v = *c
				return nil
			}
		}

		return fmt.Errorf("DB Cluster (%s) not found", rs.Primary.ID)
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

func testAccClusterConfig(rName string) string {
	return fmt.Sprintf(`
resource "aws_rds_cluster" "test" {
  cluster_identifier              = %q
  database_name                   = "mydb"
  master_username                 = "foo"
  master_password                 = "mustbeeightcharaters"
  db_cluster_parameter_group_name = "default.aurora5.6"
  skip_final_snapshot             = true
}
`, rName)
}

func testAccClusterConfig_AllowMajorVersionUpgrade(rName string, allowMajorVersionUpgrade bool, engine string, engineVersion string) string {
	return fmt.Sprintf(`
resource "aws_rds_cluster" "test" {
  allow_major_version_upgrade = %[1]t
  apply_immediately           = true
  cluster_identifier          = %[2]q
  engine                      = %[3]q
  engine_version              = %[4]q
  master_password             = "mustbeeightcharaters"
  master_username             = "test"
  skip_final_snapshot         = true
}

data "aws_rds_orderable_db_instance" "test" {
  engine                     = aws_rds_cluster.test.engine
  engine_version             = aws_rds_cluster.test.engine_version
  preferred_instance_classes = ["db.t3.medium", "db.r5.large", "db.r4.large"]
}

# Upgrading requires a healthy primary instance
resource "aws_rds_cluster_instance" "test" {
  cluster_identifier = aws_rds_cluster.test.id
  engine             = data.aws_rds_orderable_db_instance.test.engine
  engine_version     = data.aws_rds_orderable_db_instance.test.engine_version
  identifier         = %[2]q
  instance_class     = data.aws_rds_orderable_db_instance.test.instance_class

  lifecycle {
    ignore_changes = [engine_version]
  }
}
`, allowMajorVersionUpgrade, rName, engine, engineVersion)
}

func testAccClusterConfig_AllowMajorVersionUpgradeWithCustomParameters(rName string, allowMajorVersionUpgrade bool, engine string, engineVersion string, applyImmediate bool) string {
	return fmt.Sprintf(`
resource "aws_rds_cluster" "test" {
  allow_major_version_upgrade      = %[1]t
  apply_immediately                = true
  cluster_identifier               = %[2]q
  db_cluster_parameter_group_name  = aws_rds_cluster_parameter_group.test.name
  db_instance_parameter_group_name = aws_db_parameter_group.test.name
  engine                           = %[3]q
  engine_version                   = %[4]q
  master_password                  = "mustbeeightcharaters"
  master_username                  = "test"
  skip_final_snapshot              = true
}

data "aws_rds_orderable_db_instance" "test" {
  engine                     = aws_rds_cluster.test.engine
  engine_version             = aws_rds_cluster.test.engine_version
  preferred_instance_classes = ["db.t3.medium", "db.r5.large", "db.r6g.large"]
}

# Upgrading requires a healthy primary instance
resource "aws_rds_cluster_instance" "test" {
  apply_immediately       = %[5]t
  cluster_identifier      = aws_rds_cluster.test.id
  db_parameter_group_name = aws_db_parameter_group.test.name
  engine                  = data.aws_rds_orderable_db_instance.test.engine
  engine_version          = data.aws_rds_orderable_db_instance.test.engine_version
  identifier              = %[2]q
  instance_class          = data.aws_rds_orderable_db_instance.test.instance_class

  lifecycle {
    ignore_changes = [engine_version]
  }
}

resource "aws_rds_cluster_parameter_group" "test" {
  name_prefix = %[2]q
  family      = %[6]q

  lifecycle {
    create_before_destroy = true
  }
}

resource "aws_db_parameter_group" "test" {
  name_prefix = %[2]q
  family      = %[6]q

  lifecycle {
    create_before_destroy = true
  }
}
`, allowMajorVersionUpgrade, rName, engine, engineVersion, applyImmediate, engine+strings.Split(engineVersion, ".")[0])
}

func testAccClusterConfig_MajorVersionOnly(rName string, allowMajorVersionUpgrade bool, engine string, engineVersion string) string {
	return fmt.Sprintf(`
resource "aws_rds_cluster" "test" {
  allow_major_version_upgrade = %[1]t
  apply_immediately           = true
  cluster_identifier          = %[2]q
  engine                      = %[3]q
  engine_version              = %[4]q
  master_password             = "mustbeeightcharaters"
  master_username             = "test"
  skip_final_snapshot         = true
}

# Upgrading requires a healthy primary instance
resource "aws_rds_cluster_instance" "test" {
  cluster_identifier = aws_rds_cluster.test.id
  engine             = aws_rds_cluster.test.engine
  engine_version     = aws_rds_cluster.test.engine_version
  identifier         = %[2]q
  instance_class     = "db.r4.large"
}
`, allowMajorVersionUpgrade, rName, engine, engineVersion)
}

func testAccClusterConfig_AvailabilityZones(rName string) string {
	return fmt.Sprintf(`
data "aws_availability_zones" "available" {
  state = "available"

  filter {
    name   = "opt-in-status"
    values = ["opt-in-not-required"]
  }
}

resource "aws_rds_cluster" "test" {
  apply_immediately   = true
  availability_zones  = [data.aws_availability_zones.available.names[0], data.aws_availability_zones.available.names[1], data.aws_availability_zones.available.names[2]]
  cluster_identifier  = %q
  master_password     = "mustbeeightcharaters"
  master_username     = "test"
  skip_final_snapshot = true
}
`, rName)
}

func testAccClusterConfig_StorageType(rName string) string {
	return fmt.Sprintf(`
resource "aws_rds_cluster" "test" {
  apply_immediately         = true
  cluster_identifier        = %[1]q
  db_cluster_instance_class = "db.r6gd.xlarge"
  engine                    = "mysql"
  storage_type              = "io1"
  allocated_storage         = 100
  iops                      = 1000
  master_password           = "mustbeeightcharaters"
  master_username           = "test"
  skip_final_snapshot       = true
}
`, rName)
}

func testAccClusterConfig_AllocatedStorage(rName string) string {
	return fmt.Sprintf(`
resource "aws_rds_cluster" "test" {
  apply_immediately         = true
  cluster_identifier        = %[1]q
  db_cluster_instance_class = "db.r6gd.xlarge"
  engine                    = "mysql"
  storage_type              = "io1"
  allocated_storage         = 100
  iops                      = 1000
  master_password           = "mustbeeightcharaters"
  master_username           = "test"
  skip_final_snapshot       = true
}
`, rName)
}

func testAccClusterConfig_Iops(rName string) string {
	return fmt.Sprintf(`
resource "aws_rds_cluster" "test" {
  apply_immediately         = true
  cluster_identifier        = %[1]q
  db_cluster_instance_class = "db.r6gd.xlarge"
  engine                    = "mysql"
  storage_type              = "io1"
  allocated_storage         = 100
  iops                      = 1000
  master_password           = "mustbeeightcharaters"
  master_username           = "test"
  skip_final_snapshot       = true
}
`, rName)
}

func testAccClusterConfig_DbClusterInstanceClass(rName string) string {
	return fmt.Sprintf(`
resource "aws_rds_cluster" "test" {
  apply_immediately         = true
  cluster_identifier        = %[1]q
  db_cluster_instance_class = "db.r6gd.xlarge"
  engine                    = "mysql"
  storage_type              = "io1"
  allocated_storage         = 100
  iops                      = 1000
  master_password           = "mustbeeightcharaters"
  master_username           = "test"
  skip_final_snapshot       = true
}
`, rName)
}

func testAccClusterConfig_BacktrackWindow(backtrackWindow int) string {
	return fmt.Sprintf(`
resource "aws_rds_cluster" "test" {
  apply_immediately         = true
  backtrack_window          = %d
  cluster_identifier_prefix = "tf-acc-test-"
  master_password           = "mustbeeightcharaters"
  master_username           = "test"
  skip_final_snapshot       = true
}
`, backtrackWindow)
}

func testAccClusterConfig_ClusterIdentifierPrefix(clusterIdentifierPrefix string) string {
	return fmt.Sprintf(`
resource "aws_rds_cluster" "test" {
  cluster_identifier_prefix = %q
  master_username           = "root"
  master_password           = "password"
  skip_final_snapshot       = true
}
`, clusterIdentifierPrefix)
}

func testAccClusterConfig_SubnetGroupName(rName string) string {
	return fmt.Sprintf(`
data "aws_availability_zones" "available" {
  state = "available"

  filter {
    name   = "opt-in-status"
    values = ["opt-in-not-required"]
  }
}

resource "aws_rds_cluster" "test" {
  cluster_identifier   = %[1]q
  master_username      = "root"
  master_password      = "password"
  db_subnet_group_name = aws_db_subnet_group.test.name
  skip_final_snapshot  = true
}

resource "aws_vpc" "test" {
  cidr_block = "10.0.0.0/16"

  tags = {
    Name = "terraform-testacc-rds-cluster-name-prefix"
  }
}

resource "aws_subnet" "a" {
  vpc_id            = aws_vpc.test.id
  cidr_block        = "10.0.0.0/24"
  availability_zone = data.aws_availability_zones.available.names[0]

  tags = {
    Name = "tf-acc-rds-cluster-name-prefix-a"
  }
}

resource "aws_subnet" "b" {
  vpc_id            = aws_vpc.test.id
  cidr_block        = "10.0.1.0/24"
  availability_zone = data.aws_availability_zones.available.names[1]

  tags = {
    Name = "tf-acc-rds-cluster-name-prefix-b"
  }
}

resource "aws_db_subnet_group" "test" {
  name       = %[1]q
  subnet_ids = [aws_subnet.a.id, aws_subnet.b.id]
}
`, rName)
}

func testAccClusterConfig_s3Restore(bucketName string, bucketPrefix string, uniqueId string) string {
	return fmt.Sprintf(`
data "aws_availability_zones" "available" {
  state = "available"

  filter {
    name   = "opt-in-status"
    values = ["opt-in-not-required"]
  }
}

resource "aws_s3_bucket" "xtrabackup" {
  bucket = %[1]q
}

resource "aws_s3_object" "xtrabackup_db" {
  bucket = aws_s3_bucket.xtrabackup.id
  key    = "%[2]s/mysql-5-6-xtrabackup.tar.gz"
  source = "./testdata/mysql-5-6-xtrabackup.tar.gz"
  etag   = filemd5("./testdata/mysql-5-6-xtrabackup.tar.gz")
}

resource "aws_iam_role" "rds_s3_access_role" {
  name = "%[3]s-role"

  assume_role_policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Sid": "",
      "Effect": "Allow",
      "Principal": {
        "Service": "rds.amazonaws.com"
      },
      "Action": "sts:AssumeRole"
    }
  ]
}
EOF
}

resource "aws_iam_policy" "test" {
  name = "%[3]s-policy"

  policy = <<POLICY
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Action": [
        "s3:*"
      ],
      "Resource": [
        "${aws_s3_bucket.xtrabackup.arn}",
        "${aws_s3_bucket.xtrabackup.arn}/*"
      ]
    }
  ]
}
POLICY
}

resource "aws_iam_policy_attachment" "test-attach" {
  name = "%[3]s-policy-attachment"

  roles = [
    aws_iam_role.rds_s3_access_role.name,
  ]

  policy_arn = aws_iam_policy.test.arn
}

resource "aws_rds_cluster" "test" {
  cluster_identifier_prefix = "tf-test-"
  master_username           = "root"
  master_password           = "password"
  skip_final_snapshot       = true

  s3_import {
    source_engine         = "mysql"
    source_engine_version = "5.6"

    bucket_name    = aws_s3_bucket.xtrabackup.bucket
    bucket_prefix  = "%[2]s"
    ingestion_role = aws_iam_role.rds_s3_access_role.arn
  }
}
`, bucketName, bucketPrefix, uniqueId)
}

func testAccClusterConfig_generatedName() string {
	return `
resource "aws_rds_cluster" "test" {
  master_username     = "root"
  master_password     = "password"
  skip_final_snapshot = true
}
`
}

func testAccClusterWithFinalSnapshotConfig(n int) string {
	return fmt.Sprintf(`
resource "aws_rds_cluster" "test" {
  cluster_identifier              = "tf-aurora-cluster-%[1]d"
  database_name                   = "mydb"
  master_username                 = "foo"
  master_password                 = "mustbeeightcharaters"
  db_cluster_parameter_group_name = "default.aurora5.6"
  final_snapshot_identifier       = "tf-acctest-rdscluster-snapshot-%[1]d"

  tags = {
    Environment = "production"
  }
}
`, n)
}

func testAccClusterWithoutUserNameAndPasswordConfig(n int) string {
	return fmt.Sprintf(`
resource "aws_rds_cluster" "default" {
  cluster_identifier  = "tf-aurora-cluster-%d"
  database_name       = "mydb"
  skip_final_snapshot = true
}
`, n)
}

func testAccClusterConfig_pointInTimeRestoreSource(parentId, childId string) string {
	return acctest.ConfigCompose(acctest.ConfigAvailableAZsNoOptIn(), fmt.Sprintf(`
resource "aws_rds_cluster" "test" {
  cluster_identifier   = "%[1]s"
  master_username      = "root"
  master_password      = "password"
  db_subnet_group_name = aws_db_subnet_group.test.name
  skip_final_snapshot  = true
  engine               = "aurora-mysql"
}

resource "aws_vpc" "test" {
  cidr_block = "10.0.0.0/16"
  tags = {
    Name = "%[1]s-vpc"
  }
}

resource "aws_subnet" "subnets" {
  count             = length(data.aws_availability_zones.available.names)
  vpc_id            = aws_vpc.test.id
  cidr_block        = "10.0.${count.index}.0/24"
  availability_zone = data.aws_availability_zones.available.names[count.index]
  tags = {
    Name = "%[1]s-subnet-${count.index}"
  }
}

resource "aws_db_subnet_group" "test" {
  name       = "%[1]s-db-subnet-group"
  subnet_ids = aws_subnet.subnets[*].id
}

resource "aws_rds_cluster" "restored_pit" {
  cluster_identifier  = "%s"
  skip_final_snapshot = true
  engine              = aws_rds_cluster.test.engine
  restore_to_point_in_time {
    source_cluster_identifier  = aws_rds_cluster.test.cluster_identifier
    restore_type               = "full-copy"
    use_latest_restorable_time = true
  }
}
`, parentId, childId))
}

func testAccClusterConfig_pointInTimeRestoreSource_enabled_cloudWatch_logs_exports(parentId, childId, enabledCloudwatchLogExports string) string {
	return acctest.ConfigCompose(acctest.ConfigAvailableAZsNoOptIn(), fmt.Sprintf(`
resource "aws_rds_cluster" "test" {
  cluster_identifier   = "%[1]s"
  master_username      = "root"
  master_password      = "password"
  db_subnet_group_name = aws_db_subnet_group.test.name
  skip_final_snapshot  = true
  engine               = "aurora-mysql"
}

resource "aws_vpc" "test" {
  cidr_block = "10.0.0.0/16"
  tags = {
    Name = "%[1]s-vpc"
  }
}

resource "aws_subnet" "subnets" {
  count             = length(data.aws_availability_zones.available.names)
  vpc_id            = aws_vpc.test.id
  cidr_block        = "10.0.${count.index}.0/24"
  availability_zone = data.aws_availability_zones.available.names[count.index]
  tags = {
    Name = "%[1]s-subnet-${count.index}"
  }
}

resource "aws_db_subnet_group" "test" {
  name       = "%[1]s-db-subnet-group"
  subnet_ids = aws_subnet.subnets[*].id
}

resource "aws_rds_cluster" "restored_pit" {
  cluster_identifier              = "%s"
  skip_final_snapshot             = true
  engine                          = aws_rds_cluster.test.engine
  enabled_cloudwatch_logs_exports = [%q]
  restore_to_point_in_time {
    source_cluster_identifier  = aws_rds_cluster.test.cluster_identifier
    restore_type               = "full-copy"
    use_latest_restorable_time = true
  }
}
`, parentId, childId, enabledCloudwatchLogExports))
}

func testAccClusterTags1Config(rName, tagKey1, tagValue1 string) string {
	return fmt.Sprintf(`
resource "aws_rds_cluster" "test" {
  cluster_identifier  = %q
  master_username     = "foo"
  master_password     = "mustbeeightcharaters"
  skip_final_snapshot = true

  tags = {
    %q = %q
  }
}
`, rName, tagKey1, tagValue1)
}

func testAccClusterTags2Config(rName, tagKey1, tagValue1, tagKey2, tagValue2 string) string {
	return fmt.Sprintf(`
resource "aws_rds_cluster" "test" {
  cluster_identifier  = %q
  master_username     = "foo"
  master_password     = "mustbeeightcharaters"
  skip_final_snapshot = true

  tags = {
    %q = %q
    %q = %q
  }
}
`, rName, tagKey1, tagValue1, tagKey2, tagValue2)
}

func testAccClusterEnabledCloudWatchLogsExports1Config(rName, enabledCloudwatchLogExports1 string) string {
	return fmt.Sprintf(`
resource "aws_rds_cluster" "test" {
  cluster_identifier              = %q
  enabled_cloudwatch_logs_exports = [%q]
  master_username                 = "foo"
  master_password                 = "mustbeeightcharaters"
  skip_final_snapshot             = true
}
`, rName, enabledCloudwatchLogExports1)
}

func testAccClusterEnabledCloudWatchLogsExports2Config(rName, enabledCloudwatchLogExports1, enabledCloudwatchLogExports2 string) string {
	return fmt.Sprintf(`
resource "aws_rds_cluster" "test" {
  cluster_identifier              = %q
  enabled_cloudwatch_logs_exports = [%q, %q]
  master_username                 = "foo"
  master_password                 = "mustbeeightcharaters"
  skip_final_snapshot             = true
}
`, rName, enabledCloudwatchLogExports1, enabledCloudwatchLogExports2)
}

func testAccClusterEnabledCloudWatchLogsExportsPostgres1Config(rName, enabledCloudwatchLogExports1 string) string {
	return fmt.Sprintf(`
resource "aws_rds_cluster" "test" {
  cluster_identifier              = %q
  enabled_cloudwatch_logs_exports = [%q]
  engine                          = "aurora-postgresql"
  master_username                 = "foo"
  master_password                 = "mustbeeightcharaters"
  skip_final_snapshot             = true
}
`, rName, enabledCloudwatchLogExports1)
}

func testAccClusterConfig_kmsKey(n int) string {
	return fmt.Sprintf(`

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
        "AWS": "*"
      },
      "Action": "kms:*",
      "Resource": "*"
    }
  ]
}
 POLICY

}

resource "aws_rds_cluster" "test" {
  cluster_identifier              = "tf-aurora-cluster-%[1]d"
  database_name                   = "mydb"
  master_username                 = "foo"
  master_password                 = "mustbeeightcharaters"
  db_cluster_parameter_group_name = "default.aurora5.6"
  storage_encrypted               = true
  kms_key_id                      = aws_kms_key.foo.arn
  skip_final_snapshot             = true
}
`, n)
}

func testAccClusterConfig_encrypted(n int) string {
	return fmt.Sprintf(`
resource "aws_rds_cluster" "test" {
  cluster_identifier  = "tf-aurora-cluster-%d"
  database_name       = "mydb"
  master_username     = "foo"
  master_password     = "mustbeeightcharaters"
  storage_encrypted   = true
  skip_final_snapshot = true
}
`, n)
}

func testAccClusterConfig_backups(n int) string {
	return fmt.Sprintf(`
resource "aws_rds_cluster" "test" {
  cluster_identifier           = "tf-aurora-cluster-%d"
  database_name                = "mydb"
  master_username              = "foo"
  master_password              = "mustbeeightcharaters"
  backup_retention_period      = 5
  preferred_backup_window      = "07:00-09:00"
  preferred_maintenance_window = "tue:04:00-tue:04:30"
  skip_final_snapshot          = true
}
`, n)
}

func testAccClusterConfig_backupsUpdate(n int) string {
	return fmt.Sprintf(`
resource "aws_rds_cluster" "test" {
  cluster_identifier           = "tf-aurora-cluster-%d"
  database_name                = "mydb"
  master_username              = "foo"
  master_password              = "mustbeeightcharaters"
  backup_retention_period      = 10
  preferred_backup_window      = "03:00-09:00"
  preferred_maintenance_window = "wed:01:00-wed:01:30"
  apply_immediately            = true
  skip_final_snapshot          = true
}
`, n)
}

func testAccClusterConfig_iamAuth(n int) string {
	return fmt.Sprintf(`
resource "aws_rds_cluster" "test" {
  cluster_identifier                  = "tf-aurora-cluster-%d"
  database_name                       = "mydb"
  master_username                     = "foo"
  master_password                     = "mustbeeightcharaters"
  iam_database_authentication_enabled = true
  skip_final_snapshot                 = true
}
`, n)
}

func testAccClusterConfig_EngineVersion(upgrade bool, rInt int) string {
	return fmt.Sprintf(`
data "aws_rds_engine_version" "test" {
  engine             = "aurora-postgresql"
  preferred_versions = ["11.6", "11.7", "11.9"]
}

data "aws_rds_engine_version" "upgrade" {
  engine             = data.aws_rds_engine_version.test.engine
  preferred_versions = data.aws_rds_engine_version.test.valid_upgrade_targets
}

locals {
  parameter_group_name = %[1]t ? data.aws_rds_engine_version.upgrade.parameter_group_family : data.aws_rds_engine_version.test.parameter_group_family
  engine_version       = %[1]t ? data.aws_rds_engine_version.upgrade.version : data.aws_rds_engine_version.test.version
}

resource "aws_rds_cluster" "test" {
  cluster_identifier              = "tf-acc-test-%[2]d"
  database_name                   = "mydb"
  db_cluster_parameter_group_name = "default.${local.parameter_group_name}"
  engine                          = data.aws_rds_engine_version.test.engine
  engine_version                  = local.engine_version
  master_password                 = "mustbeeightcharaters"
  master_username                 = "foo"
  skip_final_snapshot             = true
  apply_immediately               = true
}
`, upgrade, rInt)
}

func testAccClusterConfig_EngineVersionWithPrimaryInstance(upgrade bool, rInt int) string {
	return fmt.Sprintf(`
data "aws_rds_engine_version" "test" {
  engine             = "aurora-postgresql"
  preferred_versions = ["10.7", "10.13", "11.6"]
}

data "aws_rds_engine_version" "upgrade" {
  engine             = data.aws_rds_engine_version.test.engine
  preferred_versions = data.aws_rds_engine_version.test.valid_upgrade_targets
}

locals {
  parameter_group_name = %[1]t ? data.aws_rds_engine_version.upgrade.parameter_group_family : data.aws_rds_engine_version.test.parameter_group_family
  engine_version       = %[1]t ? data.aws_rds_engine_version.upgrade.version : data.aws_rds_engine_version.test.version
}

resource "aws_rds_cluster" "test" {
  cluster_identifier              = "tf-acc-test-%[2]d"
  database_name                   = "mydb"
  db_cluster_parameter_group_name = "default.${local.parameter_group_name}"
  engine                          = data.aws_rds_engine_version.test.engine
  engine_version                  = local.engine_version
  master_password                 = "mustbeeightcharaters"
  master_username                 = "foo"
  skip_final_snapshot             = true
  apply_immediately               = true
}

data "aws_rds_orderable_db_instance" "test" {
  engine                     = data.aws_rds_engine_version.test.engine
  engine_version             = data.aws_rds_engine_version.test.version
  preferred_instance_classes = ["db.t2.small", "db.t3.medium", "db.r4.large"]
}

resource "aws_rds_cluster_instance" "test" {
  identifier         = "tf-acc-test-%[2]d"
  cluster_identifier = aws_rds_cluster.test.cluster_identifier
  engine             = aws_rds_cluster.test.engine
  instance_class     = data.aws_rds_orderable_db_instance.test.instance_class
}
`, upgrade, rInt)
}

func testAccClusterConfig_Port(rName string, port int) string {
	return fmt.Sprintf(`
resource "aws_rds_cluster" "test" {
  cluster_identifier              = %[1]q
  database_name                   = "mydb"
  db_cluster_parameter_group_name = "default.aurora-postgresql13"
  engine                          = "aurora-postgresql"
  master_password                 = "mustbeeightcharaters"
  master_username                 = "foo"
  port                            = %[2]d
  skip_final_snapshot             = true
}
`, rName, port)
}

func testAccClusterIncludingIAMRolesConfig(n int) string {
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
  cluster_identifier              = "tf-aurora-cluster-%[1]d"
  database_name                   = "mydb"
  master_username                 = "foo"
  master_password                 = "mustbeeightcharaters"
  db_cluster_parameter_group_name = "default.aurora5.6"
  skip_final_snapshot             = true

  tags = {
    Environment = "production"
  }

  depends_on = [aws_iam_role.another_rds_sample_role, aws_iam_role.rds_sample_role]
}
`, n)
}

func testAccClusterAddIAMRolesConfig(n int) string {
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
  cluster_identifier              = "tf-aurora-cluster-%[1]d"
  database_name                   = "mydb"
  master_username                 = "foo"
  master_password                 = "mustbeeightcharaters"
  db_cluster_parameter_group_name = "default.aurora5.6"
  skip_final_snapshot             = true
  iam_roles                       = [aws_iam_role.rds_sample_role.arn, aws_iam_role.another_rds_sample_role.arn]

  tags = {
    Environment = "production"
  }

  depends_on = [aws_iam_role.another_rds_sample_role, aws_iam_role.rds_sample_role]
}
`, n)
}

func testAccClusterRemoveIAMRolesConfig(n int) string {
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
  cluster_identifier              = "tf-aurora-cluster-%[1]d"
  database_name                   = "mydb"
  master_username                 = "foo"
  master_password                 = "mustbeeightcharaters"
  db_cluster_parameter_group_name = "default.aurora5.6"
  skip_final_snapshot             = true
  iam_roles                       = [aws_iam_role.another_rds_sample_role.arn]

  tags = {
    Environment = "production"
  }

  depends_on = [aws_iam_role.another_rds_sample_role]
}
`, n)
}

func testAccClusterReplicationSourceIdentifierKMSKeyIDConfig(rName string) string {
	return acctest.ConfigCompose(
		acctest.ConfigMultipleRegionProvider(2),
		fmt.Sprintf(`
data "aws_availability_zones" "alternate" {
  provider = "awsalternate"

  state = "available"

  filter {
    name   = "opt-in-status"
    values = ["opt-in-not-required"]
  }
}

data "aws_caller_identity" "current" {}

data "aws_region" "current" {}

resource "aws_rds_cluster_parameter_group" "test" {
  name        = %[1]q
  family      = "aurora5.6"
  description = "RDS default cluster parameter group"

  parameter {
    name         = "binlog_format"
    value        = "STATEMENT"
    apply_method = "pending-reboot"
  }
}

resource "aws_rds_cluster" "test" {
  cluster_identifier              = "%[1]s-primary"
  db_cluster_parameter_group_name = aws_rds_cluster_parameter_group.test.name
  database_name                   = "mydb"
  master_username                 = "foo"
  master_password                 = "mustbeeightcharaters"
  storage_encrypted               = true
  skip_final_snapshot             = true
}

resource "aws_rds_cluster_instance" "test" {
  identifier         = "%[1]s-primary"
  cluster_identifier = aws_rds_cluster.test.id
  instance_class     = "db.t2.small"
}

resource "aws_kms_key" "alternate" {
  provider    = "awsalternate"
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

resource "aws_vpc" "alternate" {
  provider   = "awsalternate"
  cidr_block = "10.0.0.0/16"

  tags = {
    Name = "terraform-acctest-rds-cluster-encrypted-cross-region-replica"
  }
}

resource "aws_subnet" "alternate" {
  provider          = "awsalternate"
  count             = 3
  vpc_id            = aws_vpc.alternate.id
  availability_zone = data.aws_availability_zones.alternate.names[count.index]
  cidr_block        = "10.0.${count.index}.0/24"

  tags = {
    Name = "tf-acc-rds-cluster-encrypted-cross-region-replica-${count.index}"
  }
}

resource "aws_db_subnet_group" "alternate" {
  provider   = "awsalternate"
  name       = %[1]q
  subnet_ids = aws_subnet.alternate[*].id
}

resource "aws_rds_cluster" "alternate" {
  provider                      = "awsalternate"
  cluster_identifier            = "%[1]s-replica"
  db_subnet_group_name          = aws_db_subnet_group.alternate.name
  kms_key_id                    = aws_kms_key.alternate.arn
  storage_encrypted             = true
  skip_final_snapshot           = true
  replication_source_identifier = aws_rds_cluster.test.arn
  source_region                 = data.aws_region.current.name

  depends_on = [
    aws_rds_cluster_instance.test,
  ]
}
`, rName))
}

func testAccClusterConfig_DeletionProtection(rName string, deletionProtection bool) string {
	return fmt.Sprintf(`
resource "aws_rds_cluster" "test" {
  cluster_identifier  = %q
  deletion_protection = %t
  master_password     = "barbarbarbar"
  master_username     = "foo"
  skip_final_snapshot = true
}
`, rName, deletionProtection)
}

func testAccClusterConfig_EngineMode(rName, engineMode string) string {
	return fmt.Sprintf(`
resource "aws_rds_cluster" "test" {
  cluster_identifier  = %q
  engine_mode         = %q
  master_password     = "barbarbarbar"
  master_username     = "foo"
  skip_final_snapshot = true

  #   scaling_configuration {
  # 	min_capacity = 2
  #   }
}
`, rName, engineMode)
}

func testAccClusterConfig_EngineMode_Global(rName string) string {
	return fmt.Sprintf(`
resource "aws_rds_cluster" "test" {
  cluster_identifier  = %q
  engine_mode         = "global"
  engine_version      = "5.6.10a" # version compatible with engine_mode = "global"
  master_password     = "barbarbarbar"
  master_username     = "foo"
  skip_final_snapshot = true
}
`, rName)
}

func testAccClusterConfig_EngineMode_Multimaster(rName string) string {
	return fmt.Sprintf(`
data "aws_availability_zones" "available" {
  state = "available"

  filter {
    name   = "opt-in-status"
    values = ["opt-in-not-required"]
  }
}

resource "aws_vpc" "test" {
  cidr_block = "10.0.0.0/16"

  tags = {
    Name = "tf-acc-test-rds-cluster-multimaster"
  }
}

resource "aws_subnet" "test1" {
  vpc_id            = aws_vpc.test.id
  cidr_block        = "10.0.0.0/24"
  availability_zone = data.aws_availability_zones.available.names[0]

  tags = {
    Name = "tf-acc-test-rds-cluster-multimaster"
  }
}

resource "aws_subnet" "test2" {
  vpc_id            = aws_vpc.test.id
  cidr_block        = "10.0.1.0/24"
  availability_zone = data.aws_availability_zones.available.names[1]

  tags = {
    Name = "tf-acc-test-rds-cluster-multimaster"
  }
}

resource "aws_db_subnet_group" "test" {
  name       = %[1]q
  subnet_ids = [aws_subnet.test1.id, aws_subnet.test2.id]
}

# multimaster requires db_subnet_group_name
resource "aws_rds_cluster" "test" {
  cluster_identifier   = %[1]q
  db_subnet_group_name = aws_db_subnet_group.test.name
  engine_mode          = "multimaster"
  master_password      = "barbarbarbar"
  master_username      = "foo"
  skip_final_snapshot  = true
}
`, rName)
}

func testAccClusterConfig_GlobalClusterIdentifier_EngineMode_Global(rName string) string {
	return fmt.Sprintf(`
resource "aws_rds_global_cluster" "test" {
  engine_version            = "5.6.10a" # version compatible with engine_mode = "global"
  force_destroy             = true      # Partial configuration removal ordering fix for after Terraform 0.12
  global_cluster_identifier = %[1]q
}

resource "aws_rds_cluster" "test" {
  cluster_identifier        = %[1]q
  global_cluster_identifier = aws_rds_global_cluster.test.id
  engine_mode               = "global"
  engine_version            = aws_rds_global_cluster.test.engine_version
  master_password           = "barbarbarbar"
  master_username           = "foo"
  skip_final_snapshot       = true
}
`, rName)
}

func testAccClusterConfig_GlobalClusterIdentifier_EngineMode_Global_Update(rName, globalClusterIdentifierResourceName string) string {
	return fmt.Sprintf(`
resource "aws_rds_global_cluster" "test" {
  count = 2

  engine_version            = "5.6.10a" # version compatible with engine_mode = "global"
  global_cluster_identifier = "%[1]s-${count.index}"
}

resource "aws_rds_cluster" "test" {
  cluster_identifier        = %[1]q
  global_cluster_identifier = %[2]s.id
  engine_mode               = "global"
  engine_version            = %[2]s.engine_version
  master_password           = "barbarbarbar"
  master_username           = "foo"
  skip_final_snapshot       = true
}
`, rName, globalClusterIdentifierResourceName)
}

func testAccClusterConfig_GlobalClusterIdentifier_EngineMode_Provisioned(rName string) string {
	return fmt.Sprintf(`
resource "aws_rds_global_cluster" "test" {
  engine                    = "aurora-postgresql"
  engine_version            = "12.9"
  global_cluster_identifier = %[1]q
}

resource "aws_rds_cluster" "test" {
  cluster_identifier        = %[1]q
  engine                    = aws_rds_global_cluster.test.engine
  engine_version            = aws_rds_global_cluster.test.engine_version
  global_cluster_identifier = aws_rds_global_cluster.test.id
  master_password           = "barbarbarbar"
  master_username           = "foo"
  skip_final_snapshot       = true
}
`, rName)
}

func testAccClusterConfig_GlobalClusterIdentifier_PrimarySecondaryClusters(rNameGlobal, rNamePrimary, rNameSecondary string) string {
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

resource "aws_rds_global_cluster" "test" {
  global_cluster_identifier = "%[1]s"
  engine                    = "aurora-mysql"
  engine_version            = "5.7.mysql_aurora.2.07.1"
}

resource "aws_rds_cluster" "primary" {
  cluster_identifier        = "%[2]s"
  database_name             = "mydb"
  master_username           = "foo"
  master_password           = "barbarbar"
  skip_final_snapshot       = true
  global_cluster_identifier = aws_rds_global_cluster.test.id
  engine                    = aws_rds_global_cluster.test.engine
  engine_version            = aws_rds_global_cluster.test.engine_version
}

resource "aws_rds_cluster_instance" "primary" {
  identifier         = "%[2]s"
  cluster_identifier = aws_rds_cluster.primary.id
  instance_class     = "db.r4.large" # only db.r4 or db.r5 are valid for Aurora global db
  engine             = aws_rds_cluster.primary.engine
  engine_version     = aws_rds_cluster.primary.engine_version
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

resource "aws_db_subnet_group" "alternate" {
  provider   = "awsalternate"
  name       = "%[3]s"
  subnet_ids = aws_subnet.alternate[*].id
}

resource "aws_rds_cluster" "secondary" {
  provider                  = "awsalternate"
  cluster_identifier        = "%[3]s"
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
  identifier         = "%[3]s"
  cluster_identifier = aws_rds_cluster.secondary.id
  instance_class     = "db.r4.large" # only db.r4 or db.r5 are valid for Aurora global db
  engine             = aws_rds_cluster.secondary.engine
  engine_version     = aws_rds_cluster.secondary.engine_version
}
`, rNameGlobal, rNamePrimary, rNameSecondary))
}

func testAccClusterConfig_GlobalClusterIdentifier_SecondaryClustersWriteForwarding(rNameGlobal, rNamePrimary, rNameSecondary string) string {
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

resource "aws_rds_global_cluster" "test" {
  global_cluster_identifier = "%[1]s"
  engine                    = "aurora-mysql"
  engine_version            = "5.7.mysql_aurora.2.08.1"
}

resource "aws_rds_cluster" "primary" {
  cluster_identifier        = "%[2]s"
  database_name             = "mydb"
  master_username           = "foo"
  master_password           = "barbarbar"
  skip_final_snapshot       = true
  global_cluster_identifier = aws_rds_global_cluster.test.id
  engine                    = aws_rds_global_cluster.test.engine
  engine_version            = aws_rds_global_cluster.test.engine_version
}

resource "aws_rds_cluster_instance" "primary" {
  identifier         = "%[2]s"
  cluster_identifier = aws_rds_cluster.primary.id
  instance_class     = "db.r4.large" # only db.r4 or db.r5 are valid for Aurora global db
  engine             = aws_rds_cluster.primary.engine
  engine_version     = aws_rds_cluster.primary.engine_version
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

resource "aws_db_subnet_group" "alternate" {
  provider   = "awsalternate"
  name       = "%[3]s"
  subnet_ids = aws_subnet.alternate[*].id
}

resource "aws_rds_cluster" "secondary" {
  provider                       = "awsalternate"
  cluster_identifier             = "%[3]s"
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
  identifier         = "%[3]s"
  cluster_identifier = aws_rds_cluster.secondary.id
  instance_class     = "db.r4.large" # only db.r4 or db.r5 are valid for Aurora global db
  engine             = aws_rds_cluster.secondary.engine
  engine_version     = aws_rds_cluster.secondary.engine_version
}
`, rNameGlobal, rNamePrimary, rNameSecondary))
}

func testAccClusterConfig_GlobalClusterIdentifier_ReplicationSourceIdentifier(rName string) string {
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
  engine = "aurora-postgresql"
}

data "aws_rds_orderable_db_instance" "test" {
  engine                     = data.aws_rds_engine_version.default.engine
  engine_version             = data.aws_rds_engine_version.default.version
  preferred_instance_classes = ["db.r5.large", "db.r5.xlarge", "db.r6g.large"] # Aurora global db may be limited to rx
}

resource "aws_rds_global_cluster" "test" {
  global_cluster_identifier = %[1]q
  engine                    = data.aws_rds_engine_version.default.engine
  engine_version            = data.aws_rds_engine_version.default.version
}

resource "aws_rds_cluster" "primary" {
  cluster_identifier        = "%[1]s-primary"
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
  identifier         = "%[1]s-primary"
  instance_class     = data.aws_rds_orderable_db_instance.test.instance_class
}

resource "aws_vpc" "alternate" {
  provider   = "awsalternate"
  cidr_block = "10.0.0.0/16"

  tags = {
    Name = %[1]q
  }
}

resource "aws_subnet" "alternate" {
  provider          = "awsalternate"
  count             = 3
  vpc_id            = aws_vpc.alternate.id
  availability_zone = data.aws_availability_zones.alternate.names[count.index]
  cidr_block        = "10.0.${count.index}.0/24"

  tags = {
    Name = %[1]q
  }
}

resource "aws_db_subnet_group" "alternate" {
  provider   = "awsalternate"
  name       = "%[1]s"
  subnet_ids = aws_subnet.alternate[*].id
}

resource "aws_rds_cluster" "secondary" {
  provider   = "awsalternate"
  depends_on = [aws_rds_cluster_instance.primary]

  cluster_identifier            = "%[1]s-secondary"
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
  identifier         = "%[1]s-secondary"
  instance_class     = aws_rds_cluster_instance.primary.instance_class
}
`, rName))
}

func testAccClusterConfig_ScalingConfiguration(rName string, autoPause bool, maxCapacity, minCapacity, secondsUntilAutoPause int, timeoutAction string) string {
	return fmt.Sprintf(`
resource "aws_rds_cluster" "test" {
  cluster_identifier  = %q
  engine_mode         = "serverless"
  master_password     = "barbarbarbar"
  master_username     = "foo"
  skip_final_snapshot = true

  scaling_configuration {
    auto_pause               = %t
    max_capacity             = %d
    min_capacity             = %d
    seconds_until_auto_pause = %d
    timeout_action           = "%s"
  }
}
`, rName, autoPause, maxCapacity, minCapacity, secondsUntilAutoPause, timeoutAction)
}

func testAccClusterServerlessV2ScalingConfigurationConfig(rName string, maxCapacity, minCapacity float64) string {
	return fmt.Sprintf(`
data "aws_rds_engine_version" "test" {
  engine             = "aurora-postgresql"
  preferred_versions = ["13.6"]
}

resource "aws_rds_cluster" "test" {
  cluster_identifier  = %[1]q
  master_password     = "barbarbarbar"
  master_username     = "foo"
  skip_final_snapshot = true
  engine              = data.aws_rds_engine_version.test.engine
  engine_version      = data.aws_rds_engine_version.test.version

  serverlessv2_scaling_configuration {
    max_capacity = %[2]f
    min_capacity = %[3]f
  }
}
`, rName, maxCapacity, minCapacity)
}

func testAccClusterConfig_ScalingConfiguration_DefaultMinCapacity(rName string, autoPause bool, maxCapacity, secondsUntilAutoPause int, timeoutAction string) string {
	return fmt.Sprintf(`
resource "aws_rds_cluster" "test" {
  cluster_identifier  = %q
  engine_mode         = "serverless"
  master_password     = "barbarbarbar"
  master_username     = "foo"
  skip_final_snapshot = true

  scaling_configuration {
    auto_pause               = %t
    max_capacity             = %d
    seconds_until_auto_pause = %d
    timeout_action           = "%s"
  }
}
`, rName, autoPause, maxCapacity, secondsUntilAutoPause, timeoutAction)
}

func testAccClusterConfig_SnapshotIdentifier(rName string) string {
	return fmt.Sprintf(`
resource "aws_rds_cluster" "source" {
  cluster_identifier  = "%[1]s-source"
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
  skip_final_snapshot = true
  snapshot_identifier = aws_db_cluster_snapshot.test.id
}
`, rName)
}

func testAccClusterConfig_SnapshotIdentifier_DeletionProtection(rName string, deletionProtection bool) string {
	return fmt.Sprintf(`
resource "aws_rds_cluster" "source" {
  cluster_identifier  = "%[1]s-source"
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
  deletion_protection = %[2]t
  skip_final_snapshot = true
  snapshot_identifier = aws_db_cluster_snapshot.test.id
}
`, rName, deletionProtection)
}

func testAccClusterConfig_SnapshotIdentifier_EngineMode(rName, engineMode string) string {
	return fmt.Sprintf(`
resource "aws_rds_cluster" "source" {
  cluster_identifier  = "%[1]s-source"
  engine_mode         = %[2]q
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
  engine_mode         = %[2]q
  skip_final_snapshot = true
  snapshot_identifier = aws_db_cluster_snapshot.test.id
}
`, rName, engineMode)
}

func testAccClusterConfig_SnapshotIdentifier_EngineVersion(rName string, same bool) string {
	return fmt.Sprintf(`
data "aws_rds_engine_version" "test" {
  engine             = "aurora-postgresql"
  preferred_versions = ["13.3", "12.9", "11.14"]
}

data "aws_rds_engine_version" "upgrade" {
  engine             = data.aws_rds_engine_version.test.engine
  preferred_versions = data.aws_rds_engine_version.test.valid_upgrade_targets
}

resource "aws_rds_cluster" "source" {
  cluster_identifier  = "%[1]s-source"
  engine              = data.aws_rds_engine_version.test.engine
  engine_version      = data.aws_rds_engine_version.test.version
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
  engine              = data.aws_rds_engine_version.test.engine
  engine_version      = %[2]t ? data.aws_rds_engine_version.test.version : data.aws_rds_engine_version.upgrade.version
  skip_final_snapshot = true
  snapshot_identifier = aws_db_cluster_snapshot.test.id
}
`, rName, same)
}

func testAccClusterConfig_SnapshotIdentifier_KMSKeyID(rName string) string {
	return fmt.Sprintf(`
resource "aws_kms_key" "test" {
  deletion_window_in_days = 7
}

resource "aws_rds_cluster" "source" {
  cluster_identifier  = "%[1]s-source"
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
  kms_key_id          = aws_kms_key.test.arn
  skip_final_snapshot = true
  snapshot_identifier = aws_db_cluster_snapshot.test.id
}
`, rName)
}

func testAccClusterConfig_SnapshotIdentifier_MasterPassword(rName, masterPassword string) string {
	return fmt.Sprintf(`
resource "aws_rds_cluster" "source" {
  cluster_identifier  = "%[1]s-source"
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
  master_password     = %[2]q
  skip_final_snapshot = true
  snapshot_identifier = aws_db_cluster_snapshot.test.id
}
`, rName, masterPassword)
}

func testAccClusterConfig_SnapshotIdentifier_MasterUsername(rName, masterUsername string) string {
	return fmt.Sprintf(`
resource "aws_rds_cluster" "source" {
  cluster_identifier  = "%[1]s-source"
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
  master_username     = %[2]q
  skip_final_snapshot = true
  snapshot_identifier = aws_db_cluster_snapshot.test.id
}
`, rName, masterUsername)
}

func testAccClusterConfig_SnapshotIdentifier_PreferredBackupWindow(rName, preferredBackupWindow string) string {
	return fmt.Sprintf(`
resource "aws_rds_cluster" "source" {
  cluster_identifier  = "%[1]s-source"
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
  preferred_backup_window      = %[2]q
  preferred_maintenance_window = "sun:09:00-sun:09:30"
  skip_final_snapshot          = true
  snapshot_identifier          = aws_db_cluster_snapshot.test.id
}
`, rName, preferredBackupWindow)
}

func testAccClusterConfig_SnapshotIdentifier_PreferredMaintenanceWindow(rName, preferredMaintenanceWindow string) string {
	return fmt.Sprintf(`
resource "aws_rds_cluster" "source" {
  cluster_identifier  = "%[1]s-source"
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
  preferred_maintenance_window = %[2]q
  skip_final_snapshot          = true
  snapshot_identifier          = aws_db_cluster_snapshot.test.id
}
`, rName, preferredMaintenanceWindow)
}

func testAccClusterConfig_SnapshotIdentifier_Tags(rName string) string {
	return fmt.Sprintf(`
resource "aws_rds_cluster" "source" {
  cluster_identifier  = "%[1]s-source"
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
  skip_final_snapshot = true
  snapshot_identifier = aws_db_cluster_snapshot.test.id

  tags = {
    key1 = "value1"
  }
}
`, rName)
}

func testAccClusterConfig_SnapshotIdentifier_VPCSecurityGroupIDs(rName string) string {
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
  skip_final_snapshot    = true
  snapshot_identifier    = aws_db_cluster_snapshot.test.id
  vpc_security_group_ids = [data.aws_security_group.default.id]
}
`, rName)
}

func testAccClusterConfig_SnapshotIdentifier_VPCSecurityGroupIds_Tags(rName string) string {
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
  skip_final_snapshot    = true
  snapshot_identifier    = aws_db_cluster_snapshot.test.id
  vpc_security_group_ids = [data.aws_security_group.default.id]

  tags = {
    key1 = "value1"
  }
}
`, rName)
}

func testAccClusterConfig_SnapshotIdentifier_EncryptedRestore(rName string) string {
	return fmt.Sprintf(`
resource "aws_kms_key" "test" {}

resource "aws_rds_cluster" "source" {
  cluster_identifier  = "%[1]s-source"
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
  skip_final_snapshot = true
  snapshot_identifier = aws_db_cluster_snapshot.test.id

  storage_encrypted = true
  kms_key_id        = aws_kms_key.test.arn
}
`, rName)
}

func testAccClusterWithCopyTagsToSnapshotConfig(n int, f bool) string {
	return fmt.Sprintf(`
resource "aws_rds_cluster" "test" {
  cluster_identifier              = "tf-aurora-cluster-%d"
  database_name                   = "mydb"
  master_username                 = "foo"
  master_password                 = "mustbeeightcharaters"
  db_cluster_parameter_group_name = "default.aurora5.6"
  copy_tags_to_snapshot           = %t
  skip_final_snapshot             = true
}
`, n, f)
}

func testAccClusterConfig_EnableHTTPEndpoint(rName string, enableHttpEndpoint bool) string {
	return fmt.Sprintf(`
resource "aws_rds_cluster" "test" {
  cluster_identifier   = %q
  engine_mode          = "serverless"
  master_password      = "barbarbarbar"
  master_username      = "foo"
  skip_final_snapshot  = true
  enable_http_endpoint = %t

  scaling_configuration {
    auto_pause               = false
    max_capacity             = 128
    min_capacity             = 4
    seconds_until_auto_pause = 301
    timeout_action           = "RollbackCapacityChange"
  }
}
`, rName, enableHttpEndpoint)
}
