package aws

import (
	"context"
	"errors"
	"fmt"
	"github.com/aws/aws-sdk-go/service/docdb"
	"log"
	"regexp"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func init() {
	resource.AddTestSweepers("aws_docdb_global_cluster", &resource.Sweeper{
		Name: "aws_docdb_global_cluster",
		F:    testSweepDocDBGlobalClusters,
		Dependencies: []string{
			"aws_docdb_cluster",
		},
	})
}

func testSweepDocDBGlobalClusters(region string) error {
	client, err := sharedClientForRegion(region)

	if err != nil {
		return fmt.Errorf("error getting client: %w", err)
	}

	conn := client.(*AWSClient).docdbconn
	input := &docdb.DescribeGlobalClustersInput{}

	err = conn.DescribeGlobalClustersPages(input, func(out *docdb.DescribeGlobalClustersOutput, lastPage bool) bool {
		for _, globalCluster := range out.GlobalClusters {
			id := aws.StringValue(globalCluster.GlobalClusterIdentifier)
			input := &docdb.DeleteGlobalClusterInput{
				GlobalClusterIdentifier: globalCluster.GlobalClusterIdentifier,
			}

			log.Printf("[INFO] Deleting DocDB Global Cluster: %s", id)

			_, err := conn.DeleteGlobalCluster(input)

			if err != nil {
				log.Printf("[ERROR] Failed to delete DocDB Global Cluster (%s): %s", id, err)
				continue
			}

			if err := waitForDocDBGlobalClusterDeletion(context.TODO(), conn, id); err != nil {
				log.Printf("[ERROR] Failure while waiting for DocDB Global Cluster (%s) to be deleted: %s", id, err)
			}
		}
		return !lastPage
	})

	if testSweepSkipSweepError(err) {
		log.Printf("[WARN] Skipping DocDB Global Cluster sweep for %s: %s", region, err)
		return nil
	}

	if err != nil {
		return fmt.Errorf("error retrieving DocDB Global Clusters: %w", err)
	}

	return nil
}

func TestAccAWSDocDBGlobalCluster_basic(t *testing.T) {
	var globalCluster1 docdb.GlobalCluster
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_docdb_global_cluster.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSDocDBGlobalCluster(t) },
		ErrorCheck:   testAccErrorCheck(t, docdb.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSDocDBGlobalClusterDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSDocDBGlobalClusterConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSDocDBGlobalClusterExists(resourceName, &globalCluster1),
					//This is a rds arn
					testAccCheckResourceAttrGlobalARN(resourceName, "arn", "rds", fmt.Sprintf("global-cluster:%s", rName)),
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

func TestAccAWSDocDBGlobalCluster_disappears(t *testing.T) {
	var globalCluster1 docdb.GlobalCluster
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_docdb_global_cluster.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSDocDBGlobalCluster(t) },
		ErrorCheck:   testAccErrorCheck(t, docdb.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSDocDBGlobalClusterDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSDocDBGlobalClusterConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSDocDBGlobalClusterExists(resourceName, &globalCluster1),
					testAccCheckAWSDocDBGlobalClusterDisappears(&globalCluster1),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccAWSDocDBGlobalCluster_DatabaseName(t *testing.T) {
	var globalCluster1, globalCluster2 docdb.GlobalCluster
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_docdb_global_cluster.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSDocDBGlobalCluster(t) },
		ErrorCheck:   testAccErrorCheck(t, docdb.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSDocDBGlobalClusterDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSDocDBGlobalClusterConfigDatabaseName(rName, "database1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSDocDBGlobalClusterExists(resourceName, &globalCluster1),
					resource.TestCheckResourceAttr(resourceName, "database_name", "database1"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccAWSDocDBGlobalClusterConfigDatabaseName(rName, "database2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSDocDBGlobalClusterExists(resourceName, &globalCluster2),
					testAccCheckAWSDocDBGlobalClusterRecreated(&globalCluster1, &globalCluster2),
					resource.TestCheckResourceAttr(resourceName, "database_name", "database2"),
				),
			},
		},
	})
}

func TestAccAWSDocDBGlobalCluster_DeletionProtection(t *testing.T) {
	var globalCluster1, globalCluster2 docdb.GlobalCluster
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_docdb_global_cluster.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSDocDBGlobalCluster(t) },
		ErrorCheck:   testAccErrorCheck(t, docdb.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSDocDBGlobalClusterDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSDocDBGlobalClusterConfigDeletionProtection(rName, true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSDocDBGlobalClusterExists(resourceName, &globalCluster1),
					resource.TestCheckResourceAttr(resourceName, "deletion_protection", "true"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccAWSDocDBGlobalClusterConfigDeletionProtection(rName, false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSDocDBGlobalClusterExists(resourceName, &globalCluster2),
					testAccCheckAWSDocDBGlobalClusterNotRecreated(&globalCluster1, &globalCluster2),
					resource.TestCheckResourceAttr(resourceName, "deletion_protection", "false"),
				),
			},
		},
	})
}

func TestAccAWSDocDBGlobalCluster_Engine(t *testing.T) {
	var globalCluster1 docdb.GlobalCluster
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_docdb_global_cluster.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSDocDBGlobalCluster(t) },
		ErrorCheck:   testAccErrorCheck(t, docdb.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSDocDBGlobalClusterDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSDocDBGlobalClusterConfigEngine(rName, "docdb"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSDocDBGlobalClusterExists(resourceName, &globalCluster1),
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

func TestAccAWSDocDBGlobalCluster_EngineVersion(t *testing.T) {
	var globalCluster1 docdb.GlobalCluster
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_docdb_global_cluster.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSDocDBGlobalCluster(t) },
		ErrorCheck:   testAccErrorCheck(t, docdb.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSDocDBGlobalClusterDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSDocDBGlobalClusterConfigEngineVersion(rName, "docdb", "4.0.0"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSDocDBGlobalClusterExists(resourceName, &globalCluster1),
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

func TestAccAWSDocDBGlobalCluster_SourceDbClusterIdentifier(t *testing.T) {
	var globalCluster1 docdb.GlobalCluster
	rName := acctest.RandomWithPrefix("tf-acc-test")
	clusterResourceName := "aws_docdb_cluster.test"
	resourceName := "aws_docdb_global_cluster.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSDocDBGlobalCluster(t) },
		ErrorCheck:   testAccErrorCheck(t, docdb.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSDocDBGlobalClusterDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSDocDBGlobalClusterConfigSourceDbClusterIdentifier(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSDocDBGlobalClusterExists(resourceName, &globalCluster1),
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

func TestAccAWSDocDBGlobalCluster_SourceDbClusterIdentifier_StorageEncrypted(t *testing.T) {
	var globalCluster1 docdb.GlobalCluster
	rName := acctest.RandomWithPrefix("tf-acc-test")
	clusterResourceName := "aws_docdb_cluster.test"
	resourceName := "aws_docdb_global_cluster.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSDocDBGlobalCluster(t) },
		ErrorCheck:   testAccErrorCheck(t, docdb.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSDocDBGlobalClusterDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSDocDBGlobalClusterConfigSourceDbClusterIdentifierStorageEncrypted(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSDocDBGlobalClusterExists(resourceName, &globalCluster1),
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

func TestAccAWSDocDBGlobalCluster_StorageEncrypted(t *testing.T) {
	var globalCluster1, globalCluster2 docdb.GlobalCluster
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_docdb_global_cluster.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSDocDBGlobalCluster(t) },
		ErrorCheck:   testAccErrorCheck(t, docdb.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSDocDBGlobalClusterDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSDocDBGlobalClusterConfigStorageEncrypted(rName, true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSDocDBGlobalClusterExists(resourceName, &globalCluster1),
					resource.TestCheckResourceAttr(resourceName, "storage_encrypted", "true"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccAWSDocDBGlobalClusterConfigStorageEncrypted(rName, false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSDocDBGlobalClusterExists(resourceName, &globalCluster2),
					testAccCheckAWSDocDBGlobalClusterRecreated(&globalCluster1, &globalCluster2),
					resource.TestCheckResourceAttr(resourceName, "storage_encrypted", "false"),
				),
			},
		},
	})
}

func testAccCheckAWSDocDBGlobalClusterExists(resourceName string, globalCluster *docdb.GlobalCluster) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("Not found: %s", resourceName)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("no DocDB Global Cluster ID is set")
		}

		conn := testAccProvider.Meta().(*AWSClient).docdbconn

		cluster, err := docDBDescribeGlobalCluster(context.TODO(), conn, rs.Primary.ID)

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

func testAccCheckAWSDocDBGlobalClusterDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*AWSClient).docdbconn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_docdb_global_cluster" {
			continue
		}

		globalCluster, err := docDBDescribeGlobalCluster(context.TODO(), conn, rs.Primary.ID)

		if isAWSErr(err, docdb.ErrCodeGlobalClusterNotFoundFault, "") {
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

func testAccCheckAWSDocDBGlobalClusterDisappears(globalCluster *docdb.GlobalCluster) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := testAccProvider.Meta().(*AWSClient).docdbconn

		input := &docdb.DeleteGlobalClusterInput{
			GlobalClusterIdentifier: globalCluster.GlobalClusterIdentifier,
		}

		_, err := conn.DeleteGlobalCluster(input)

		if err != nil {
			return err
		}

		return waitForDocDBGlobalClusterDeletion(context.TODO(), conn, aws.StringValue(globalCluster.GlobalClusterIdentifier))
	}
}

func testAccCheckAWSDocDBGlobalClusterNotRecreated(i, j *docdb.GlobalCluster) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if aws.StringValue(i.GlobalClusterArn) != aws.StringValue(j.GlobalClusterArn) {
			return fmt.Errorf("docDB Global Cluster was recreated. got: %s, expected: %s", aws.StringValue(i.GlobalClusterArn), aws.StringValue(j.GlobalClusterArn))
		}

		return nil
	}
}

func testAccCheckAWSDocDBGlobalClusterRecreated(i, j *docdb.GlobalCluster) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if aws.StringValue(i.GlobalClusterResourceId) == aws.StringValue(j.GlobalClusterResourceId) {
			return errors.New("docDB Global Cluster was not recreated")
		}

		return nil
	}
}

func testAccPreCheckAWSDocDBGlobalCluster(t *testing.T) {
	conn := testAccProvider.Meta().(*AWSClient).docdbconn

	input := &docdb.DescribeGlobalClustersInput{}

	_, err := conn.DescribeGlobalClusters(input)

	if testAccPreCheckSkipError(err) || isAWSErr(err, "InvalidParameterValue", "Access Denied to API Version: APIGlobalDatabases") {
		// Current Region/Partition does not support DocDB Global Clusters
		t.Skipf("skipping acceptance testing: %s", err)
	}

	if err != nil {
		t.Fatalf("unexpected PreCheck error: %s", err)
	}
}

func testAccAWSDocDBGlobalClusterConfig(rName string) string {
	return fmt.Sprintf(`
resource "aws_docdb_global_cluster" "test" {
  global_cluster_identifier = %q
}
`, rName)
}

func testAccAWSDocDBGlobalClusterConfigDatabaseName(rName, databaseName string) string {
	return fmt.Sprintf(`
resource "aws_docdb_global_cluster" "test" {
  database_name             = %q
  global_cluster_identifier = %q
}
`, databaseName, rName)
}

func testAccAWSDocDBGlobalClusterConfigDeletionProtection(rName string, deletionProtection bool) string {
	return fmt.Sprintf(`
resource "aws_docdb_global_cluster" "test" {
  deletion_protection       = %t
  global_cluster_identifier = %q
}
`, deletionProtection, rName)
}

func testAccAWSDocDBGlobalClusterConfigEngine(rName, engine string) string {
	return fmt.Sprintf(`
resource "aws_docdb_global_cluster" "test" {
  engine                    = %q
  global_cluster_identifier = %q
}
`, engine, rName)
}

func testAccAWSDocDBGlobalClusterConfigEngineVersion(rName, engine, engineVersion string) string {
	return fmt.Sprintf(`
resource "aws_docdb_global_cluster" "test" {
  engine                    = %q
  engine_version            = %q
  global_cluster_identifier = %q
}
`, engine, engineVersion, rName)
}

func testAccAWSDocDBGlobalClusterConfigSourceDbClusterIdentifier(rName string) string {
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
  force_destroy                = true
  global_cluster_identifier    = %[1]q
  source_db_cluster_identifier = aws_docdb_cluster.test.arn
}
`, rName)
}

func testAccAWSDocDBGlobalClusterConfigSourceDbClusterIdentifierStorageEncrypted(rName string) string {
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
  force_destroy                = true
  global_cluster_identifier    = %[1]q
  source_db_cluster_identifier = aws_docdb_cluster.test.arn
}
`, rName)
}

func testAccAWSDocDBGlobalClusterConfigStorageEncrypted(rName string, storageEncrypted bool) string {
	return fmt.Sprintf(`
resource "aws_docdb_global_cluster" "test" {
  global_cluster_identifier = %q
  storage_encrypted         = %t
}
`, rName, storageEncrypted)
}
