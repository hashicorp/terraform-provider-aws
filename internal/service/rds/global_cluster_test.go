package rds_test

import (
	"errors"
	"fmt"
	"regexp"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/rds"
	"github.com/hashicorp/aws-sdk-go-base/tfawserr"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfrds "github.com/hashicorp/terraform-provider-aws/internal/service/rds"
)

func TestAccRDSGlobalCluster_basic(t *testing.T) {
	var globalCluster1 rds.GlobalCluster
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_rds_global_cluster.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t); testAccPreCheckGlobalCluster(t) },
		ErrorCheck:   acctest.ErrorCheck(t, rds.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckGlobalClusterDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccGlobalClusterConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGlobalClusterExists(resourceName, &globalCluster1),
					acctest.CheckResourceAttrGlobalARN(resourceName, "arn", "rds", fmt.Sprintf("global-cluster:%s", rName)),
					resource.TestCheckResourceAttr(resourceName, "database_name", ""),
					resource.TestCheckResourceAttr(resourceName, "deletion_protection", "false"),
					resource.TestCheckResourceAttrSet(resourceName, "engine"),
					resource.TestCheckResourceAttrSet(resourceName, "engine_version"),
					resource.TestCheckResourceAttr(resourceName, "global_cluster_identifier", rName),
					resource.TestMatchResourceAttr(resourceName, "global_cluster_resource_id", regexp.MustCompile(`cluster-.+`)),
					resource.TestCheckResourceAttr(resourceName, "storage_encrypted", "false"),
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
	var globalCluster1 rds.GlobalCluster
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_rds_global_cluster.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t); testAccPreCheckGlobalCluster(t) },
		ErrorCheck:   acctest.ErrorCheck(t, rds.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckGlobalClusterDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccGlobalClusterConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGlobalClusterExists(resourceName, &globalCluster1),
					testAccCheckGlobalClusterDisappears(&globalCluster1),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccRDSGlobalCluster_databaseName(t *testing.T) {
	var globalCluster1, globalCluster2 rds.GlobalCluster
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_rds_global_cluster.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t); testAccPreCheckGlobalCluster(t) },
		ErrorCheck:   acctest.ErrorCheck(t, rds.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckGlobalClusterDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccGlobalClusterDatabaseNameConfig(rName, "database1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGlobalClusterExists(resourceName, &globalCluster1),
					resource.TestCheckResourceAttr(resourceName, "database_name", "database1"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccGlobalClusterDatabaseNameConfig(rName, "database2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGlobalClusterExists(resourceName, &globalCluster2),
					testAccCheckGlobalClusterRecreated(&globalCluster1, &globalCluster2),
					resource.TestCheckResourceAttr(resourceName, "database_name", "database2"),
				),
			},
		},
	})
}

func TestAccRDSGlobalCluster_deletionProtection(t *testing.T) {
	var globalCluster1, globalCluster2 rds.GlobalCluster
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_rds_global_cluster.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t); testAccPreCheckGlobalCluster(t) },
		ErrorCheck:   acctest.ErrorCheck(t, rds.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckGlobalClusterDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccGlobalClusterDeletionProtectionConfig(rName, true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGlobalClusterExists(resourceName, &globalCluster1),
					resource.TestCheckResourceAttr(resourceName, "deletion_protection", "true"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccGlobalClusterDeletionProtectionConfig(rName, false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGlobalClusterExists(resourceName, &globalCluster2),
					testAccCheckGlobalClusterNotRecreated(&globalCluster1, &globalCluster2),
					resource.TestCheckResourceAttr(resourceName, "deletion_protection", "false"),
				),
			},
		},
	})
}

func TestAccRDSGlobalCluster_Engine_aurora(t *testing.T) {
	var globalCluster1 rds.GlobalCluster
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_rds_global_cluster.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t); testAccPreCheckGlobalCluster(t) },
		ErrorCheck:   acctest.ErrorCheck(t, rds.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckGlobalClusterDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccGlobalClusterEngineConfig(rName, "aurora"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGlobalClusterExists(resourceName, &globalCluster1),
					resource.TestCheckResourceAttr(resourceName, "engine", "aurora"),
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

func TestAccRDSGlobalCluster_EngineVersion_aurora(t *testing.T) {
	var globalCluster1 rds.GlobalCluster
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_rds_global_cluster.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t); testAccPreCheckGlobalCluster(t) },
		ErrorCheck:   acctest.ErrorCheck(t, rds.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckGlobalClusterDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccGlobalClusterEngineVersionConfig(rName, "aurora", "5.6.10a"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGlobalClusterExists(resourceName, &globalCluster1),
					resource.TestCheckResourceAttr(resourceName, "engine_version", "5.6.10a"),
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

func TestAccRDSGlobalCluster_engineVersionUpdateMinor(t *testing.T) {
	var globalCluster1, globalCluster2 rds.GlobalCluster
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_rds_global_cluster.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t); testAccPreCheckGlobalCluster(t) },
		ErrorCheck:   acctest.ErrorCheck(t, rds.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckGlobalClusterDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccGlobalClusterWithPrimaryEngineVersionConfig(rName, "aurora", "5.6.mysql_aurora.1.22.2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGlobalClusterExists(resourceName, &globalCluster1),
				),
			},
			{
				Config: testAccGlobalClusterWithPrimaryEngineVersionConfig(rName, "aurora", "5.6.mysql_aurora.1.23.2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGlobalClusterExists(resourceName, &globalCluster2),
					testAccCheckGlobalClusterNotRecreated(&globalCluster1, &globalCluster2),
					resource.TestCheckResourceAttr(resourceName, "engine_version", "5.6.mysql_aurora.1.23.2"),
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

func TestAccRDSGlobalCluster_engineVersionUpdateMajor(t *testing.T) {
	var globalCluster1, globalCluster2 rds.GlobalCluster
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_rds_global_cluster.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t); testAccPreCheckGlobalCluster(t) },
		ErrorCheck:   acctest.ErrorCheck(t, rds.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckGlobalClusterDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccGlobalClusterWithPrimaryEngineVersionConfig(rName, "aurora", "5.6.mysql_aurora.1.22.2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGlobalClusterExists(resourceName, &globalCluster1),
				),
			},
			{
				Config:             testAccGlobalClusterWithPrimaryEngineVersionConfig(rName, "aurora", "5.7.mysql_aurora.2.07.2"),
				ExpectNonEmptyPlan: true,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGlobalClusterExists(resourceName, &globalCluster2),
					testAccCheckGlobalClusterNotRecreated(&globalCluster1, &globalCluster2),
					resource.TestCheckResourceAttr(resourceName, "engine_version", "5.7.mysql_aurora.2.07.2"),
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

func TestAccRDSGlobalCluster_EngineVersion_auroraMySQL(t *testing.T) {
	var globalCluster1 rds.GlobalCluster
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_rds_global_cluster.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t); testAccPreCheckGlobalCluster(t) },
		ErrorCheck:   acctest.ErrorCheck(t, rds.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckGlobalClusterDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccGlobalClusterEngineVersionConfig(rName, "aurora-mysql", "5.7.mysql_aurora.2.07.1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGlobalClusterExists(resourceName, &globalCluster1),
					resource.TestCheckResourceAttr(resourceName, "engine_version", "5.7.mysql_aurora.2.07.1"),
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

func TestAccRDSGlobalCluster_EngineVersion_auroraPostgresql(t *testing.T) {
	var globalCluster1 rds.GlobalCluster
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_rds_global_cluster.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t); testAccPreCheckGlobalCluster(t) },
		ErrorCheck:   acctest.ErrorCheck(t, rds.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckGlobalClusterDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccGlobalClusterEngineVersionConfig(rName, "aurora-postgresql", "10.11"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGlobalClusterExists(resourceName, &globalCluster1),
					resource.TestCheckResourceAttr(resourceName, "engine_version", "10.11"),
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

func TestAccRDSGlobalCluster_sourceDBClusterIdentifier(t *testing.T) {
	var globalCluster1 rds.GlobalCluster
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	clusterResourceName := "aws_rds_cluster.test"
	resourceName := "aws_rds_global_cluster.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t); testAccPreCheckGlobalCluster(t) },
		ErrorCheck:   acctest.ErrorCheck(t, rds.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckGlobalClusterDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccGlobalClusterSourceClusterIdentifierConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGlobalClusterExists(resourceName, &globalCluster1),
					resource.TestCheckResourceAttrPair(resourceName, "source_db_cluster_identifier", clusterResourceName, "arn"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"force_destroy", "source_db_cluster_identifier"},
			},
		},
	})
}

func TestAccRDSGlobalCluster_SourceDBClusterIdentifier_storageEncrypted(t *testing.T) {
	var globalCluster1 rds.GlobalCluster
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	clusterResourceName := "aws_rds_cluster.test"
	resourceName := "aws_rds_global_cluster.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t); testAccPreCheckGlobalCluster(t) },
		ErrorCheck:   acctest.ErrorCheck(t, rds.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckGlobalClusterDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccGlobalClusterSourceClusterIdentifierStorageEncryptedConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGlobalClusterExists(resourceName, &globalCluster1),
					resource.TestCheckResourceAttrPair(resourceName, "source_db_cluster_identifier", clusterResourceName, "arn"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"force_destroy", "source_db_cluster_identifier"},
			},
		},
	})
}

func TestAccRDSGlobalCluster_storageEncrypted(t *testing.T) {
	var globalCluster1, globalCluster2 rds.GlobalCluster
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_rds_global_cluster.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t); testAccPreCheckGlobalCluster(t) },
		ErrorCheck:   acctest.ErrorCheck(t, rds.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckGlobalClusterDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccGlobalClusterStorageEncryptedConfig(rName, true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGlobalClusterExists(resourceName, &globalCluster1),
					resource.TestCheckResourceAttr(resourceName, "storage_encrypted", "true"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccGlobalClusterStorageEncryptedConfig(rName, false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGlobalClusterExists(resourceName, &globalCluster2),
					testAccCheckGlobalClusterRecreated(&globalCluster1, &globalCluster2),
					resource.TestCheckResourceAttr(resourceName, "storage_encrypted", "false"),
				),
			},
		},
	})
}

func testAccCheckGlobalClusterExists(resourceName string, globalCluster *rds.GlobalCluster) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("Not found: %s", resourceName)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No RDS Global Cluster ID is set")
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).RDSConn

		cluster, err := tfrds.DescribeGlobalCluster(conn, rs.Primary.ID)

		if err != nil {
			return err
		}

		if cluster == nil {
			return fmt.Errorf("RDS Global Cluster not found")
		}

		if aws.StringValue(cluster.Status) != "available" {
			return fmt.Errorf("RDS Global Cluster (%s) exists in non-available (%s) state", rs.Primary.ID, aws.StringValue(cluster.Status))
		}

		*globalCluster = *cluster

		return nil
	}
}

func testAccCheckGlobalClusterDestroy(s *terraform.State) error {
	conn := acctest.Provider.Meta().(*conns.AWSClient).RDSConn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_rds_global_cluster" {
			continue
		}

		globalCluster, err := tfrds.DescribeGlobalCluster(conn, rs.Primary.ID)

		if tfawserr.ErrMessageContains(err, rds.ErrCodeGlobalClusterNotFoundFault, "") {
			continue
		}

		if err != nil {
			return err
		}

		if globalCluster == nil {
			continue
		}

		return fmt.Errorf("RDS Global Cluster (%s) still exists in non-deleted (%s) state", rs.Primary.ID, aws.StringValue(globalCluster.Status))
	}

	return nil
}

func testAccCheckGlobalClusterDisappears(globalCluster *rds.GlobalCluster) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).RDSConn

		input := &rds.DeleteGlobalClusterInput{
			GlobalClusterIdentifier: globalCluster.GlobalClusterIdentifier,
		}

		_, err := conn.DeleteGlobalCluster(input)

		if err != nil {
			return err
		}

		return tfrds.WaitForGlobalClusterDeletion(conn, aws.StringValue(globalCluster.GlobalClusterIdentifier))
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

func testAccPreCheckGlobalCluster(t *testing.T) {
	conn := acctest.Provider.Meta().(*conns.AWSClient).RDSConn

	input := &rds.DescribeGlobalClustersInput{}

	_, err := conn.DescribeGlobalClusters(input)

	if acctest.PreCheckSkipError(err) || tfawserr.ErrMessageContains(err, "InvalidParameterValue", "Access Denied to API Version: APIGlobalDatabases") {
		t.Skipf("skipping acceptance testing: %s", err)
	}

	if err != nil {
		t.Fatalf("unexpected PreCheck error: %s", err)
	}
}

func testAccGlobalClusterConfig(rName string) string {
	return fmt.Sprintf(`
resource "aws_rds_global_cluster" "test" {
  global_cluster_identifier = %q
}
`, rName)
}

func testAccGlobalClusterDatabaseNameConfig(rName, databaseName string) string {
	return fmt.Sprintf(`
resource "aws_rds_global_cluster" "test" {
  database_name             = %q
  global_cluster_identifier = %q
}
`, databaseName, rName)
}

func testAccGlobalClusterDeletionProtectionConfig(rName string, deletionProtection bool) string {
	return fmt.Sprintf(`
resource "aws_rds_global_cluster" "test" {
  deletion_protection       = %t
  global_cluster_identifier = %q
}
`, deletionProtection, rName)
}

func testAccGlobalClusterEngineConfig(rName, engine string) string {
	return fmt.Sprintf(`
resource "aws_rds_global_cluster" "test" {
  engine                    = %q
  global_cluster_identifier = %q
}
`, engine, rName)
}

func testAccGlobalClusterEngineVersionConfig(rName, engine, engineVersion string) string {
	return fmt.Sprintf(`
resource "aws_rds_global_cluster" "test" {
  engine                    = %q
  engine_version            = %q
  global_cluster_identifier = %q
}
`, engine, engineVersion, rName)
}

func testAccGlobalClusterWithPrimaryEngineVersionConfig(rName, engine, engineVersion string) string {
	return fmt.Sprintf(`
resource "aws_rds_global_cluster" "test" {
  engine                    = %[1]q
  engine_version            = %[2]q
  global_cluster_identifier = %[3]q
}

resource "aws_rds_cluster" "test" {
  apply_immediately           = true
  allow_major_version_upgrade = true
  cluster_identifier          = %[3]q
  master_password             = "mustbeeightcharacters"
  master_username             = "test"
  skip_final_snapshot         = true

  global_cluster_identifier = aws_rds_global_cluster.test.global_cluster_identifier

  lifecycle {
    ignore_changes = [global_cluster_identifier]
  }
}

resource "aws_rds_cluster_instance" "test" {
  apply_immediately  = true
  cluster_identifier = aws_rds_cluster.test.id
  identifier         = %[3]q
  instance_class     = "db.r3.large"

  lifecycle {
    ignore_changes = [engine_version]
  }
}
`, engine, engineVersion, rName)
}

func testAccGlobalClusterSourceClusterIdentifierConfig(rName string) string {
	return fmt.Sprintf(`
resource "aws_rds_cluster" "test" {
  cluster_identifier  = %[1]q
  engine              = "aurora-postgresql"
  engine_version      = "10.11" # Minimum supported version for Global Clusters
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

func testAccGlobalClusterSourceClusterIdentifierStorageEncryptedConfig(rName string) string {
	return fmt.Sprintf(`
resource "aws_rds_cluster" "test" {
  cluster_identifier  = %[1]q
  engine              = "aurora-postgresql"
  engine_version      = "10.11" # Minimum supported version for Global Clusters
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

func testAccGlobalClusterStorageEncryptedConfig(rName string, storageEncrypted bool) string {
	return fmt.Sprintf(`
resource "aws_rds_global_cluster" "test" {
  global_cluster_identifier = %q
  storage_encrypted         = %t
}
`, rName, storageEncrypted)
}
