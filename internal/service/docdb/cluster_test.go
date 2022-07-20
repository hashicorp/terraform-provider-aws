package docdb_test

import (
	"errors"
	"fmt"
	"log"
	"regexp"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/docdb"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
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
	var dbCluster docdb.DBCluster
	rInt := sdkacctest.RandInt()
	resourceName := "aws_docdb_cluster.default"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, docdb.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckClusterDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccClusterConfig_basic(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterExists(resourceName, &dbCluster),
					acctest.MatchResourceAttrRegionalARN(resourceName, "arn", "rds", regexp.MustCompile(`cluster:.+`)),
					resource.TestCheckResourceAttr(resourceName, "storage_encrypted", "false"),
					resource.TestCheckResourceAttr(resourceName, "db_cluster_parameter_group_name", "default.docdb4.0"),
					resource.TestCheckResourceAttrSet(resourceName, "reader_endpoint"),
					resource.TestCheckResourceAttrSet(resourceName, "cluster_resource_id"),
					resource.TestCheckResourceAttr(resourceName, "engine", "docdb"),
					resource.TestCheckResourceAttrSet(resourceName, "engine_version"),
					resource.TestCheckResourceAttrSet(resourceName, "hosted_zone_id"),
					resource.TestCheckResourceAttr(resourceName,
						"enabled_cloudwatch_logs_exports.0", "audit"),
					resource.TestCheckResourceAttr(resourceName,
						"enabled_cloudwatch_logs_exports.1", "profiler"),
					resource.TestCheckResourceAttr(resourceName, "deletion_protection", "false"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
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
	var v docdb.DBCluster

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, docdb.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckClusterDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccClusterConfig_namePrefix(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterExists("aws_docdb_cluster.test", &v),
					resource.TestMatchResourceAttr(
						"aws_docdb_cluster.test", "cluster_identifier", regexp.MustCompile("^tf-test-")),
				),
			},
			{
				ResourceName:      "aws_docdb_cluster.test",
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
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
	var v docdb.DBCluster

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, docdb.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckClusterDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccClusterConfig_generatedName(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterExists("aws_docdb_cluster.test", &v),
					resource.TestMatchResourceAttr(
						"aws_docdb_cluster.test", "cluster_identifier", regexp.MustCompile("^tf-")),
				),
			},
			{
				ResourceName:      "aws_docdb_cluster.test",
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
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
	var dbCluster1 docdb.DBCluster

	rName := sdkacctest.RandomWithPrefix("tf-acc-test")
	globalClusterResourceName := "aws_docdb_cluster.test"
	resourceName := "aws_docdb_cluster.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheckGlobalCluster(t) },
		ErrorCheck:        acctest.ErrorCheck(t, docdb.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckClusterDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccClusterConfig_globalIdentifier(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterExists(resourceName, &dbCluster1),
					resource.TestCheckResourceAttrPair(resourceName, "global_cluster_identifier", globalClusterResourceName, "id"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"apply_immediately",
					"cluster_identifier_prefix",
					"master_password",
					"skip_final_snapshot",
					"snapshot_identifier",
				},
			},
		},
	})
}

func TestAccDocDBCluster_GlobalClusterIdentifier_Add(t *testing.T) {
	var dbCluster1 docdb.DBCluster

	rName := sdkacctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_docdb_cluster.test"

	if acctest.Partition() == "aws-us-gov" {
		t.Skip("DocDB Global Cluster is not supported in GovCloud partition")
	}

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheckGlobalCluster(t) },
		ErrorCheck:        acctest.ErrorCheck(t, docdb.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckClusterDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccClusterConfig_globalCompatible(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterExists(resourceName, &dbCluster1),
					resource.TestCheckResourceAttr(resourceName, "global_cluster_identifier", ""),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"apply_immediately",
					"cluster_identifier_prefix",
					"master_password",
					"skip_final_snapshot",
					"snapshot_identifier",
				},
			},
			{
				Config:      testAccClusterConfig_globalIdentifier(rName),
				ExpectError: regexp.MustCompile(`existing DocDB Clusters cannot be added to an existing DocDB Global Cluster`),
			},
		},
	})
}

func TestAccDocDBCluster_GlobalClusterIdentifier_Remove(t *testing.T) {
	var dbCluster1 docdb.DBCluster

	rName := sdkacctest.RandomWithPrefix("tf-acc-test")
	globalClusterResourceName := "aws_docdb_global_cluster.test"
	resourceName := "aws_docdb_cluster.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheckGlobalCluster(t) },
		ErrorCheck:        acctest.ErrorCheck(t, docdb.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckClusterDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccClusterConfig_globalIdentifier(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterExists(resourceName, &dbCluster1),
					resource.TestCheckResourceAttrPair(resourceName, "global_cluster_identifier", globalClusterResourceName, "id"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"apply_immediately",
					"cluster_identifier_prefix",
					"master_password",
					"skip_final_snapshot",
					"snapshot_identifier",
				},
			},
			{
				Config: testAccClusterConfig_globalCompatible(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterExists(resourceName, &dbCluster1),
					resource.TestCheckResourceAttr(resourceName, "global_cluster_identifier", ""),
				),
			},
		},
	})
}

func TestAccDocDBCluster_GlobalClusterIdentifier_Update(t *testing.T) {
	var dbCluster1 docdb.DBCluster

	rName := sdkacctest.RandomWithPrefix("tf-acc-test")
	globalClusterResourceName1 := "aws_docdb_global_cluster.test.0"
	globalClusterResourceName2 := "aws_docdb_global_cluster.test.1"
	resourceName := "aws_docdb_cluster.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheckGlobalCluster(t) },
		ErrorCheck:        acctest.ErrorCheck(t, docdb.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckClusterDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccClusterConfig_globalIdentifierUpdate(rName, globalClusterResourceName1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterExists(resourceName, &dbCluster1),
					resource.TestCheckResourceAttrPair(resourceName, "global_cluster_identifier", globalClusterResourceName1, "id"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"apply_immediately",
					"cluster_identifier_prefix",
					"master_password",
					"skip_final_snapshot",
					"snapshot_identifier",
				},
			},
			{
				Config:      testAccClusterConfig_globalIdentifierUpdate(rName, globalClusterResourceName2),
				ExpectError: regexp.MustCompile(`existing DocDB Clusters cannot be migrated between existing DocDB Global Clusters`),
			},
		},
	})
}

func TestAccDocDBCluster_GlobalClusterIdentifier_PrimarySecondaryClusters(t *testing.T) {
	var providers []*schema.Provider
	var primaryDbCluster, secondaryDbCluster docdb.DBCluster

	rNameGlobal := sdkacctest.RandomWithPrefix("tf-acc-test-global")
	rNamePrimary := sdkacctest.RandomWithPrefix("tf-acc-test-primary")
	rNameSecondary := sdkacctest.RandomWithPrefix("tf-acc-test-secondary")

	resourceNamePrimary := "aws_docdb_cluster.primary"
	resourceNameSecondary := "aws_docdb_cluster.secondary"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(t)
			acctest.PreCheckMultipleRegion(t, 2)
			testAccPreCheckGlobalCluster(t)
		},
		ErrorCheck:        acctest.ErrorCheck(t, docdb.EndpointsID),
		ProviderFactories: acctest.FactoriesAlternate(&providers),

		CheckDestroy: testAccCheckClusterDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccClusterConfig_globalIdentifierPrimarySecondary(rNameGlobal, rNamePrimary, rNameSecondary),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterExistsProvider(resourceNamePrimary, &primaryDbCluster, acctest.RegionProviderFunc(acctest.Region(), &providers)),
					testAccCheckClusterExistsProvider(resourceNameSecondary, &secondaryDbCluster, acctest.RegionProviderFunc(acctest.AlternateRegion(), &providers)),
				),
			},
		},
	})
}

func TestAccDocDBCluster_takeFinalSnapshot(t *testing.T) {
	var v docdb.DBCluster
	rInt := sdkacctest.RandInt()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, docdb.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckClusterSnapshot(rInt),
		Steps: []resource.TestStep{
			{
				Config: testAccClusterConfig_finalSnapshot(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterExists("aws_docdb_cluster.default", &v),
				),
			},
			{
				ResourceName:      "aws_docdb_cluster.default",
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
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

/// This is a regression test to make sure that we always cover the scenario as hightlighted in
/// https://github.com/hashicorp/terraform/issues/11568
func TestAccDocDBCluster_missingUserNameCausesError(t *testing.T) {
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, docdb.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckClusterDestroy,
		Steps: []resource.TestStep{
			{
				Config:      testAccClusterConfig_noUsernameOrPassword(sdkacctest.RandInt()),
				ExpectError: regexp.MustCompile(`required field is not set`),
			},
		},
	})
}

func TestAccDocDBCluster_updateTags(t *testing.T) {
	var v docdb.DBCluster
	ri := sdkacctest.RandInt()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, docdb.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckClusterDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccClusterConfig_basic(ri),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterExists("aws_docdb_cluster.default", &v),
					resource.TestCheckResourceAttr(
						"aws_docdb_cluster.default", "tags.%", "1"),
				),
			},
			{
				ResourceName:      "aws_docdb_cluster.default",
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"apply_immediately",
					"cluster_identifier_prefix",
					"final_snapshot_identifier",
					"master_password",
					"skip_final_snapshot",
				},
			},
			{
				Config: testAccClusterConfig_updatedTags(ri),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterExists("aws_docdb_cluster.default", &v),
					resource.TestCheckResourceAttr(
						"aws_docdb_cluster.default", "tags.%", "2"),
				),
			},
		},
	})
}

func TestAccDocDBCluster_updateCloudWatchLogsExports(t *testing.T) {
	var v docdb.DBCluster
	ri := sdkacctest.RandInt()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, docdb.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckClusterDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccClusterConfig_noCloudWatchLogs(ri),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterExists("aws_docdb_cluster.default", &v),
				),
			},
			{
				ResourceName:      "aws_docdb_cluster.default",
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"apply_immediately",
					"cluster_identifier_prefix",
					"final_snapshot_identifier",
					"master_password",
					"skip_final_snapshot",
				},
			},
			{
				Config: testAccClusterConfig_basic(ri),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterExists("aws_docdb_cluster.default", &v),
					resource.TestCheckResourceAttr("aws_docdb_cluster.default",
						"enabled_cloudwatch_logs_exports.0", "audit"),
				),
			},
		},
	})
}

func TestAccDocDBCluster_kmsKey(t *testing.T) {
	var v docdb.DBCluster

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, docdb.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckClusterDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccClusterConfig_kmsKey(sdkacctest.RandInt()),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterExists("aws_docdb_cluster.default", &v),
					resource.TestCheckResourceAttrPair("aws_docdb_cluster.default", "kms_key_id", "aws_kms_key.foo", "arn"),
				),
			},
			{
				ResourceName:      "aws_docdb_cluster.default",
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
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
	var v docdb.DBCluster

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, docdb.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckClusterDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccClusterConfig_encrypted(sdkacctest.RandInt()),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterExists("aws_docdb_cluster.default", &v),
					resource.TestCheckResourceAttr(
						"aws_docdb_cluster.default", "storage_encrypted", "true"),
					resource.TestCheckResourceAttr(
						"aws_docdb_cluster.default", "db_cluster_parameter_group_name", "default.docdb4.0"),
				),
			},
			{
				ResourceName:      "aws_docdb_cluster.default",
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
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
	var v docdb.DBCluster

	ri := sdkacctest.RandInt()
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, docdb.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckClusterDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccClusterConfig_backups(ri),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterExists("aws_docdb_cluster.default", &v),
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
					"apply_immediately",
					"cluster_identifier_prefix",
					"final_snapshot_identifier",
					"master_password",
					"skip_final_snapshot",
				},
			},
			{
				Config: testAccClusterConfig_backupsUpdate(ri),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterExists("aws_docdb_cluster.default", &v),
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
	var dbCluster1, dbCluster2 docdb.DBCluster
	rInt := sdkacctest.RandInt()
	resourceName := "aws_docdb_cluster.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, docdb.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckClusterDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccClusterConfig_port(rInt, 5432),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterExists(resourceName, &dbCluster1),
					resource.TestCheckResourceAttr(resourceName, "port", "5432"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"apply_immediately",
					"cluster_identifier_prefix",
					"final_snapshot_identifier",
					"master_password",
					"skip_final_snapshot",
				},
			},
			{
				Config: testAccClusterConfig_port(rInt, 2345),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterExists(resourceName, &dbCluster2),
					testAccCheckClusterRecreated(&dbCluster1, &dbCluster2),
					resource.TestCheckResourceAttr(resourceName, "port", "2345"),
				),
			},
		},
	})
}

func TestAccDocDBCluster_deleteProtection(t *testing.T) {
	var dbCluster docdb.DBCluster
	resourceName := "aws_docdb_cluster.default"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, docdb.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckClusterDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccClusterConfig_deleteProtection(true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterExists(resourceName, &dbCluster),
					resource.TestCheckResourceAttr(resourceName, "deletion_protection", "true"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
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
					testAccCheckClusterExists(resourceName, &dbCluster),
					resource.TestCheckResourceAttr(resourceName, "deletion_protection", "false"),
				),
			},
			{
				Config: testAccClusterConfig_deleteProtection(true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterExists(resourceName, &dbCluster),
					resource.TestCheckResourceAttr(resourceName, "deletion_protection", "true"),
				),
			},
			{
				Config: testAccClusterConfig_deleteProtection(false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterExists(resourceName, &dbCluster),
					resource.TestCheckResourceAttr(resourceName, "deletion_protection", "false"),
				),
			},
		},
	})
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

func testAccCheckClusterDestroy(s *terraform.State) error {
	return testAccCheckClusterDestroyProvider(s, acctest.Provider)
}

func testAccCheckClusterDestroyProvider(s *terraform.State, provider *schema.Provider) error {
	conn := provider.Meta().(*conns.AWSClient).DocDBConn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_docdb_cluster" {
			continue
		}

		// Try to find the Group
		var err error
		resp, err := conn.DescribeDBClusters(
			&docdb.DescribeDBClustersInput{
				DBClusterIdentifier: aws.String(rs.Primary.ID),
			})

		if err == nil {
			if len(resp.DBClusters) != 0 &&
				*resp.DBClusters[0].DBClusterIdentifier == rs.Primary.ID {
				return fmt.Errorf("DB Cluster %s still exists", rs.Primary.ID)
			}
		}

		if tfawserr.ErrCodeEquals(err, docdb.ErrCodeDBClusterNotFoundFault) {
			continue
		}

		return err
	}

	return nil
}

func testAccCheckClusterExists(n string, v *docdb.DBCluster) resource.TestCheckFunc {
	return testAccCheckClusterExistsProvider(n, v, func() *schema.Provider { return acctest.Provider })
}

func testAccCheckClusterExistsProvider(n string, v *docdb.DBCluster, providerF func() *schema.Provider) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No DB Instance ID is set")
		}

		provider := providerF()
		conn := provider.Meta().(*conns.AWSClient).DocDBConn
		resp, err := conn.DescribeDBClusters(&docdb.DescribeDBClustersInput{
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

func testAccCheckClusterRecreated(i, j *docdb.DBCluster) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if aws.TimeValue(i.ClusterCreateTime).Equal(aws.TimeValue(j.ClusterCreateTime)) {
			return errors.New("DocDB Cluster was not recreated")
		}

		return nil
	}
}

func testAccCheckClusterSnapshot(rInt int) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_docdb_cluster" {
				continue
			}

			// Try and delete the snapshot before we check for the cluster not found
			snapshot_identifier := fmt.Sprintf("tf-acctest-docdbcluster-snapshot-%d", rInt)

			awsClient := acctest.Provider.Meta().(*conns.AWSClient)
			conn := awsClient.DocDBConn

			log.Printf("[INFO] Deleting the Snapshot %s", snapshot_identifier)
			_, snapDeleteErr := conn.DeleteDBClusterSnapshot(
				&docdb.DeleteDBClusterSnapshotInput{
					DBClusterSnapshotIdentifier: aws.String(snapshot_identifier),
				})
			if snapDeleteErr != nil {
				return snapDeleteErr
			}

			// Try to find the Group
			var err error
			resp, err := conn.DescribeDBClusters(
				&docdb.DescribeDBClustersInput{
					DBClusterIdentifier: aws.String(rs.Primary.ID),
				})

			if err == nil {
				if len(resp.DBClusters) != 0 &&
					*resp.DBClusters[0].DBClusterIdentifier == rs.Primary.ID {
					return fmt.Errorf("DB Cluster %s still exists", rs.Primary.ID)
				}
			}

			if tfawserr.ErrCodeEquals(err, docdb.ErrCodeDBClusterNotFoundFault) {
				continue
			}

			return err
		}

		return nil
	}
}

func testAccClusterConfig_basic(n int) string {
	return acctest.ConfigCompose(acctest.ConfigAvailableAZsNoOptIn(), fmt.Sprintf(`
resource "aws_docdb_cluster" "default" {
  cluster_identifier = "tf-docdb-cluster-%d"

  availability_zones = [
    data.aws_availability_zones.available.names[0],
    data.aws_availability_zones.available.names[1],
    data.aws_availability_zones.available.names[2]
  ]

  master_username                 = "foo"
  master_password                 = "mustbeeightcharaters"
  db_cluster_parameter_group_name = "default.docdb4.0"
  skip_final_snapshot             = true

  tags = {
    Environment = "production"
  }

  enabled_cloudwatch_logs_exports = [
    "audit",
    "profiler",
  ]
}
`, n))
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

func testAccClusterConfig_finalSnapshot(n int) string {
	return acctest.ConfigCompose(acctest.ConfigAvailableAZsNoOptIn(), fmt.Sprintf(`
resource "aws_docdb_cluster" "default" {
  cluster_identifier = "tf-docdb-cluster-%[1]d"

  availability_zones = [
    data.aws_availability_zones.available.names[0],
    data.aws_availability_zones.available.names[1],
    data.aws_availability_zones.available.names[2]
  ]

  master_username                 = "foo"
  master_password                 = "mustbeeightcharaters"
  db_cluster_parameter_group_name = "default.docdb4.0"
  final_snapshot_identifier       = "tf-acctest-docdbcluster-snapshot-%[1]d"

  tags = {
    Environment = "production"
  }
}
`, n))
}

func testAccClusterConfig_noUsernameOrPassword(n int) string {
	return acctest.ConfigCompose(acctest.ConfigAvailableAZsNoOptIn(), fmt.Sprintf(`
resource "aws_docdb_cluster" "default" {
  cluster_identifier = "tf-docdb-cluster-%d"

  availability_zones = [
    data.aws_availability_zones.available.names[0],
    data.aws_availability_zones.available.names[1],
    data.aws_availability_zones.available.names[2]
  ]

  skip_final_snapshot = true
}
`, n))
}

func testAccClusterConfig_updatedTags(n int) string {
	return acctest.ConfigCompose(acctest.ConfigAvailableAZsNoOptIn(), fmt.Sprintf(`
resource "aws_docdb_cluster" "default" {
  cluster_identifier = "tf-docdb-cluster-%d"

  availability_zones = [
    data.aws_availability_zones.available.names[0],
    data.aws_availability_zones.available.names[1],
    data.aws_availability_zones.available.names[2]
  ]

  master_username                 = "foo"
  master_password                 = "mustbeeightcharaters"
  db_cluster_parameter_group_name = "default.docdb4.0"
  skip_final_snapshot             = true

  tags = {
    Environment = "production"
    AnotherTag  = "test"
  }
}
`, n))
}

func testAccClusterConfig_noCloudWatchLogs(n int) string {
	return acctest.ConfigCompose(acctest.ConfigAvailableAZsNoOptIn(), fmt.Sprintf(`
resource "aws_docdb_cluster" "default" {
  cluster_identifier = "tf-docdb-cluster-%d"

  availability_zones = [
    data.aws_availability_zones.available.names[0],
    data.aws_availability_zones.available.names[1],
    data.aws_availability_zones.available.names[2]
  ]

  master_username                 = "foo"
  master_password                 = "mustbeeightcharaters"
  db_cluster_parameter_group_name = "default.docdb4.0"
  skip_final_snapshot             = true

  tags = {
    Environment = "production"
  }
}
`, n))
}

func testAccClusterConfig_kmsKey(n int) string {
	return acctest.ConfigCompose(acctest.ConfigAvailableAZsNoOptIn(), fmt.Sprintf(`
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

resource "aws_docdb_cluster" "default" {
  cluster_identifier = "tf-docdb-cluster-%[1]d"
  availability_zones = [
    data.aws_availability_zones.available.names[0],
    data.aws_availability_zones.available.names[1],
    data.aws_availability_zones.available.names[2]
  ]

  master_username                 = "foo"
  master_password                 = "mustbeeightcharaters"
  db_cluster_parameter_group_name = "default.docdb4.0"
  storage_encrypted               = true
  kms_key_id                      = aws_kms_key.foo.arn
  skip_final_snapshot             = true
}
`, n))
}

func testAccClusterConfig_encrypted(n int) string {
	return acctest.ConfigCompose(acctest.ConfigAvailableAZsNoOptIn(), fmt.Sprintf(`
resource "aws_docdb_cluster" "default" {
  cluster_identifier = "tf-docdb-cluster-%d"

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
`, n))
}

func testAccClusterConfig_backups(n int) string {
	return acctest.ConfigCompose(acctest.ConfigAvailableAZsNoOptIn(), fmt.Sprintf(`
resource "aws_docdb_cluster" "default" {
  cluster_identifier = "tf-docdb-cluster-%d"

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
`, n))
}

func testAccClusterConfig_backupsUpdate(n int) string {
	return acctest.ConfigCompose(acctest.ConfigAvailableAZsNoOptIn(), fmt.Sprintf(`
resource "aws_docdb_cluster" "default" {
  cluster_identifier = "tf-docdb-cluster-%d"

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
`, n))
}

func testAccClusterConfig_port(rInt, port int) string {
	return acctest.ConfigCompose(acctest.ConfigAvailableAZsNoOptIn(), fmt.Sprintf(`
resource "aws_docdb_cluster" "test" {
  availability_zones = [
    data.aws_availability_zones.available.names[0],
    data.aws_availability_zones.available.names[1],
    data.aws_availability_zones.available.names[2]
  ]

  cluster_identifier              = "tf-acc-test-%d"
  db_cluster_parameter_group_name = "default.docdb4.0"
  engine                          = "docdb"
  master_password                 = "mustbeeightcharaters"
  master_username                 = "foo"
  port                            = %d
  skip_final_snapshot             = true
}
`, rInt, port))
}

func testAccClusterConfig_deleteProtection(isProtected bool) string {
	return fmt.Sprintf(`
resource "aws_docdb_cluster" "default" {
  cluster_identifier_prefix = "tf-test-"
  master_username           = "root"
  master_password           = "password"
  skip_final_snapshot       = true
  deletion_protection       = %t
}
`, isProtected)
}
