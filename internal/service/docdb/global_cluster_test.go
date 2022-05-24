package docdb_test

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/docdb"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfdocdb "github.com/hashicorp/terraform-provider-aws/internal/service/docdb"
)

func TestAccDocDBGlobalCluster_basic(t *testing.T) {
	var globalCluster1 docdb.GlobalCluster

	rName := sdkacctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_docdb_global_cluster.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheckGlobalCluster(t) },
		ErrorCheck:        acctest.ErrorCheck(t, docdb.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckGlobalClusterDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccGlobalClusterConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGlobalClusterExists(resourceName, &globalCluster1),
					//This is a rds arn
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

func TestAccDocDBGlobalCluster_disappears(t *testing.T) {
	var globalCluster1 docdb.GlobalCluster
	rName := sdkacctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_docdb_global_cluster.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheckGlobalCluster(t) },
		ErrorCheck:        acctest.ErrorCheck(t, docdb.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckGlobalClusterDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccGlobalClusterConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGlobalClusterExists(resourceName, &globalCluster1),
					testAccCheckGlobalClusterDisappears(&globalCluster1),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccDocDBGlobalCluster_DatabaseName(t *testing.T) {
	var globalCluster1, globalCluster2 docdb.GlobalCluster
	rName := sdkacctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_docdb_global_cluster.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheckGlobalCluster(t) },
		ErrorCheck:        acctest.ErrorCheck(t, docdb.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckGlobalClusterDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccGlobalClusterConfig_databaseName(rName, "database1"),
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
				Config: testAccGlobalClusterConfig_databaseName(rName, "database2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGlobalClusterExists(resourceName, &globalCluster2),
					testAccCheckGlobalClusterRecreated(&globalCluster1, &globalCluster2),
					resource.TestCheckResourceAttr(resourceName, "database_name", "database2"),
				),
			},
		},
	})
}

func TestAccDocDBGlobalCluster_DeletionProtection(t *testing.T) {
	var globalCluster1, globalCluster2 docdb.GlobalCluster
	rName := sdkacctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_docdb_global_cluster.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheckGlobalCluster(t) },
		ErrorCheck:        acctest.ErrorCheck(t, docdb.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckGlobalClusterDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccGlobalClusterConfig_deletionProtection(rName, true),
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
				Config: testAccGlobalClusterConfig_deletionProtection(rName, false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGlobalClusterExists(resourceName, &globalCluster2),
					testAccCheckGlobalClusterNotRecreated(&globalCluster1, &globalCluster2),
					resource.TestCheckResourceAttr(resourceName, "deletion_protection", "false"),
				),
			},
		},
	})
}

func TestAccDocDBGlobalCluster_Engine(t *testing.T) {
	var globalCluster1 docdb.GlobalCluster
	rName := sdkacctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_docdb_global_cluster.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheckGlobalCluster(t) },
		ErrorCheck:        acctest.ErrorCheck(t, docdb.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckGlobalClusterDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccGlobalClusterConfig_engine(rName, "docdb"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGlobalClusterExists(resourceName, &globalCluster1),
					resource.TestCheckResourceAttr(resourceName, "engine", "docdb"),
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

func TestAccDocDBGlobalCluster_EngineVersion(t *testing.T) {
	var globalCluster1 docdb.GlobalCluster
	rName := sdkacctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_docdb_global_cluster.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheckGlobalCluster(t) },
		ErrorCheck:        acctest.ErrorCheck(t, docdb.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckGlobalClusterDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccGlobalClusterConfig_engineVersion(rName, "docdb", "4.0.0"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGlobalClusterExists(resourceName, &globalCluster1),
					resource.TestCheckResourceAttr(resourceName, "engine_version", "4.0.0"),
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

func TestAccDocDBGlobalCluster_SourceDbClusterIdentifier(t *testing.T) {
	var globalCluster1 docdb.GlobalCluster
	rName := sdkacctest.RandomWithPrefix("tf-acc-test")
	clusterResourceName := "aws_docdb_cluster.test"
	resourceName := "aws_docdb_global_cluster.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheckGlobalCluster(t) },
		ErrorCheck:        acctest.ErrorCheck(t, docdb.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckGlobalClusterDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccGlobalClusterConfig_sourceDBClusterIdentifier(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGlobalClusterExists(resourceName, &globalCluster1),
					resource.TestCheckResourceAttrPair(resourceName, "source_db_cluster_identifier", clusterResourceName, "arn"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"source_db_cluster_identifier"},
			},
		},
	})
}

func TestAccDocDBGlobalCluster_SourceDbClusterIdentifier_StorageEncrypted(t *testing.T) {
	var globalCluster1 docdb.GlobalCluster
	rName := sdkacctest.RandomWithPrefix("tf-acc-test")
	clusterResourceName := "aws_docdb_cluster.test"
	resourceName := "aws_docdb_global_cluster.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheckGlobalCluster(t) },
		ErrorCheck:        acctest.ErrorCheck(t, docdb.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckGlobalClusterDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccGlobalClusterConfig_sourceDBClusterIdentifierStorageEncrypted(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGlobalClusterExists(resourceName, &globalCluster1),
					resource.TestCheckResourceAttrPair(resourceName, "source_db_cluster_identifier", clusterResourceName, "arn"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"source_db_cluster_identifier"},
			},
		},
	})
}

func TestAccDocDBGlobalCluster_StorageEncrypted(t *testing.T) {
	var globalCluster1, globalCluster2 docdb.GlobalCluster
	rName := sdkacctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_docdb_global_cluster.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheckGlobalCluster(t) },
		ErrorCheck:        acctest.ErrorCheck(t, docdb.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckGlobalClusterDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccGlobalClusterConfig_storageEncrypted(rName, true),
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
				Config: testAccGlobalClusterConfig_storageEncrypted(rName, false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGlobalClusterExists(resourceName, &globalCluster2),
					testAccCheckGlobalClusterRecreated(&globalCluster1, &globalCluster2),
					resource.TestCheckResourceAttr(resourceName, "storage_encrypted", "false"),
				),
			},
		},
	})
}

func testAccCheckGlobalClusterExists(resourceName string, globalCluster *docdb.GlobalCluster) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("Not found: %s", resourceName)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("no DocDB Global Cluster ID is set")
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).DocDBConn
		cluster, err := tfdocdb.FindGlobalClusterById(context.TODO(), conn, rs.Primary.ID)

		if err != nil {
			return err
		}

		if cluster == nil {
			return fmt.Errorf("docDB Global Cluster not found")
		}

		if aws.StringValue(cluster.Status) != "available" {
			return fmt.Errorf("docDB Global Cluster (%s) exists in non-available (%s) state", rs.Primary.ID, aws.StringValue(cluster.Status))
		}

		*globalCluster = *cluster

		return nil
	}
}

func testAccCheckGlobalClusterDestroy(s *terraform.State) error {
	conn := acctest.Provider.Meta().(*conns.AWSClient).DocDBConn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_docdb_global_cluster" {
			continue
		}

		globalCluster, err := tfdocdb.FindGlobalClusterById(context.TODO(), conn, rs.Primary.ID)

		if tfawserr.ErrCodeEquals(err, docdb.ErrCodeGlobalClusterNotFoundFault) {
			continue
		}

		if err != nil {
			return err
		}

		if globalCluster == nil {
			continue
		}

		return fmt.Errorf("docDB Global Cluster (%s) still exists in non-deleted (%s) state", rs.Primary.ID, aws.StringValue(globalCluster.Status))
	}

	return nil
}

func testAccCheckGlobalClusterDisappears(globalCluster *docdb.GlobalCluster) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).DocDBConn

		input := &docdb.DeleteGlobalClusterInput{
			GlobalClusterIdentifier: globalCluster.GlobalClusterIdentifier,
		}

		_, err := conn.DeleteGlobalCluster(input)

		if err != nil {
			return err
		}

		return tfdocdb.WaitForGlobalClusterDeletion(context.TODO(), conn, aws.StringValue(globalCluster.GlobalClusterIdentifier), tfdocdb.GlobalClusterDeleteTimeout)
	}
}

func testAccCheckGlobalClusterNotRecreated(i, j *docdb.GlobalCluster) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if aws.StringValue(i.GlobalClusterArn) != aws.StringValue(j.GlobalClusterArn) {
			return fmt.Errorf("docDB Global Cluster was recreated. got: %s, expected: %s", aws.StringValue(i.GlobalClusterArn), aws.StringValue(j.GlobalClusterArn))
		}

		return nil
	}
}

func testAccCheckGlobalClusterRecreated(i, j *docdb.GlobalCluster) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if aws.StringValue(i.GlobalClusterResourceId) == aws.StringValue(j.GlobalClusterResourceId) {
			return errors.New("docDB Global Cluster was not recreated")
		}

		return nil
	}
}

func testAccPreCheckGlobalCluster(t *testing.T) {
	conn := acctest.Provider.Meta().(*conns.AWSClient).DocDBConn

	input := &docdb.DescribeGlobalClustersInput{}

	_, err := conn.DescribeGlobalClusters(input)

	if acctest.PreCheckSkipError(err) || tfawserr.ErrMessageContains(err, "InvalidParameterValue", "Access Denied to API Version: APIGlobalDatabases") {
		// Current Region/Partition does not support DocDB Global Clusters
		t.Skipf("skipping acceptance testing: %s", err)
	}

	if err != nil {
		t.Fatalf("unexpected PreCheck error: %s", err)
	}
}

func testAccGlobalClusterConfig_basic(rName string) string {
	return fmt.Sprintf(`
resource "aws_docdb_global_cluster" "test" {
  engine                    = "docdb"
  global_cluster_identifier = %q
}
`, rName)
}

func testAccGlobalClusterConfig_databaseName(rName, databaseName string) string {
	return fmt.Sprintf(`
resource "aws_docdb_global_cluster" "test" {
  engine                    = "docdb"
  database_name             = %q
  global_cluster_identifier = %q
}
`, databaseName, rName)
}

func testAccGlobalClusterConfig_deletionProtection(rName string, deletionProtection bool) string {
	return fmt.Sprintf(`
resource "aws_docdb_global_cluster" "test" {
  engine                    = "docdb"
  deletion_protection       = %t
  global_cluster_identifier = %q
}
`, deletionProtection, rName)
}

func testAccGlobalClusterConfig_engine(rName, engine string) string {
	return fmt.Sprintf(`
resource "aws_docdb_global_cluster" "test" {
  engine                    = %q
  global_cluster_identifier = %q
}
`, engine, rName)
}

func testAccGlobalClusterConfig_engineVersion(rName, engine, engineVersion string) string {
	return fmt.Sprintf(`
resource "aws_docdb_global_cluster" "test" {
  engine                    = %q
  engine_version            = %q
  global_cluster_identifier = %q
}
`, engine, engineVersion, rName)
}

func testAccGlobalClusterConfig_sourceDBClusterIdentifier(rName string) string {
	return fmt.Sprintf(`
resource "aws_docdb_cluster" "test" {
  cluster_identifier  = %[1]q
  engine              = "docdb"
  engine_version      = "4.0.0" # Minimum supported version for Global Clusters
  master_password     = "mustbeeightcharacters"
  master_username     = "test"
  skip_final_snapshot = true

  # global_cluster_identifier cannot be Computed

  lifecycle {
    ignore_changes = [global_cluster_identifier]
  }
}

resource "aws_docdb_global_cluster" "test" {
  global_cluster_identifier    = %[1]q
  source_db_cluster_identifier = aws_docdb_cluster.test.arn
}
`, rName)
}

func testAccGlobalClusterConfig_sourceDBClusterIdentifierStorageEncrypted(rName string) string {
	return fmt.Sprintf(`
resource "aws_docdb_cluster" "test" {
  cluster_identifier  = %[1]q
  engine              = "docdb"
  engine_version      = "4.0.0" # Minimum supported version for Global Clusters
  master_password     = "mustbeeightcharacters"
  master_username     = "test"
  skip_final_snapshot = true
  storage_encrypted   = true

  # global_cluster_identifier cannot be Computed

  lifecycle {
    ignore_changes = [global_cluster_identifier]
  }
}

resource "aws_docdb_global_cluster" "test" {
  global_cluster_identifier    = %[1]q
  source_db_cluster_identifier = aws_docdb_cluster.test.arn
}
`, rName)
}

func testAccGlobalClusterConfig_storageEncrypted(rName string, storageEncrypted bool) string {
	return fmt.Sprintf(`
resource "aws_docdb_global_cluster" "test" {
  global_cluster_identifier = %q
  engine                    = "docdb"
  storage_encrypted         = %t
}
`, rName, storageEncrypted)
}
