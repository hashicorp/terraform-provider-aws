package aws

import (
	//"errors"
	"fmt"
	"log"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform/helper/acctest"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/helper/schema"
	"github.com/hashicorp/terraform/terraform"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/neptune"
)

func TestAccAWSNeptuneCluster_basic(t *testing.T) {
	var dbCluster neptune.DBCluster
	rInt := acctest.RandInt()
	resourceName := "aws_neptune_cluster.default"

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSNeptuneClusterDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSNeptuneClusterConfig(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSNeptuneClusterExists(resourceName, &dbCluster),
					resource.TestCheckResourceAttr(resourceName, "storage_encrypted", "false"),
					resource.TestCheckResourceAttr(resourceName, "neptune_cluster_parameter_group_name", "default.neptune1"),
					resource.TestCheckResourceAttrSet(resourceName, "reader_endpoint"),
					resource.TestCheckResourceAttrSet(resourceName, "cluster_resource_id"),
					resource.TestCheckResourceAttr(resourceName, "engine", "neptune"),
					resource.TestCheckResourceAttrSet(resourceName, "engine_version"),
					resource.TestCheckResourceAttrSet(resourceName, "hosted_zone_id"),
				),
			},
		},
	})
}

func TestAccAWSNeptuneCluster_namePrefix(t *testing.T) {
	var v neptune.DBCluster

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSNeptuneClusterDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSNeptuneClusterConfig_namePrefix(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSNeptuneClusterExists("aws_neptune_cluster.test", &v),
					resource.TestMatchResourceAttr(
						"aws_neptune_cluster.test", "cluster_identifier", regexp.MustCompile("^tf-test-")),
				),
			},
		},
	})
}

func TestAccAWSNeptuneCluster_takeFinalSnapshot(t *testing.T) {
	var v neptune.DBCluster
	rInt := acctest.RandInt()

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSNeptuneClusterSnapshot(rInt),
		Steps: []resource.TestStep{
			{
				Config: testAccAWSNeptuneClusterConfigWithFinalSnapshot(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSNeptuneClusterExists("aws_neptune_cluster.default", &v),
				),
			},
		},
	})
}

func testAccCheckAWSNeptuneClusterDestroy(s *terraform.State) error {
	return testAccCheckAWSNeptuneClusterDestroyWithProvider(s, testAccProvider)
}

func testAccCheckAWSNeptuneClusterDestroyWithProvider(s *terraform.State, provider *schema.Provider) error {
	conn := provider.Meta().(*AWSClient).neptuneconn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_neptune_cluster" {
			continue
		}

		// Try to find the Group
		var err error
		resp, err := conn.DescribeDBClusters(
			&neptune.DescribeDBClustersInput{
				DBClusterIdentifier: aws.String(rs.Primary.ID),
			})

		if err == nil {
			if len(resp.DBClusters) != 0 &&
				aws.StringValue(resp.DBClusters[0].DBClusterIdentifier) == rs.Primary.ID {
				return fmt.Errorf("Neptune Cluster %s still exists", rs.Primary.ID)
			}
		}

		// Return nil if the cluster is already destroyed
		if err != nil {
			if isAWSErr(err, neptune.ErrCodeDBClusterNotFoundFault, "") {
				return nil
			}
		}

		return err
	}

	return nil
}

func testAccCheckAWSNeptuneClusterExists(n string, v *neptune.DBCluster) resource.TestCheckFunc {
	return testAccCheckAWSNeptuneClusterExistsWithProvider(n, v, func() *schema.Provider { return testAccProvider })
}

func testAccCheckAWSNeptuneClusterExistsWithProvider(n string, v *neptune.DBCluster, providerF func() *schema.Provider) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No Neptune Instance ID is set")
		}

		provider := providerF()
		conn := provider.Meta().(*AWSClient).neptuneconn
		resp, err := conn.DescribeDBClusters(&neptune.DescribeDBClustersInput{
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

		return fmt.Errorf("Neptune Cluster (%s) not found", rs.Primary.ID)
	}
}

func testAccCheckAWSNeptuneClusterSnapshot(rInt int) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_neptune_cluster" {
				continue
			}

			// Try and delete the snapshot before we check for the cluster not found
			snapshot_identifier := fmt.Sprintf("tf-acctest-neptunecluster-snapshot-%d", rInt)

			awsClient := testAccProvider.Meta().(*AWSClient)
			conn := awsClient.neptuneconn

			log.Printf("[INFO] Deleting the Snapshot %s", snapshot_identifier)
			_, snapDeleteErr := conn.DeleteDBClusterSnapshot(
				&neptune.DeleteDBClusterSnapshotInput{
					DBClusterSnapshotIdentifier: aws.String(snapshot_identifier),
				})
			if snapDeleteErr != nil {
				return snapDeleteErr
			}

			// Try to find the Group
			var err error
			resp, err := conn.DescribeDBClusters(
				&neptune.DescribeDBClustersInput{
					DBClusterIdentifier: aws.String(rs.Primary.ID),
				})

			if err == nil {
				if len(resp.DBClusters) != 0 &&
					aws.StringValue(resp.DBClusters[0].DBClusterIdentifier) == rs.Primary.ID {
					return fmt.Errorf("Neptune Cluster %s still exists", rs.Primary.ID)
				}
			}

			// Return nil if the cluster is already destroyed
			if err != nil {
				if isAWSErr(err, neptune.ErrCodeDBClusterNotFoundFault, "") {
					return nil
				}
			}

			return err
		}

		return nil
	}
}

func testAccAWSNeptuneClusterConfig(n int) string {
	return fmt.Sprintf(`
resource "aws_neptune_cluster" "default" {
  cluster_identifier = "tf-neptune-cluster-%d"
  availability_zones = ["us-west-2a","us-west-2b","us-west-2c"]
  engine = "neptune"
  neptune_cluster_parameter_group_name = "default.neptune1"
  skip_final_snapshot = true
  tags {
    Environment = "production"
  }
}`, n)
}

func testAccAWSNeptuneClusterConfig_namePrefix() string {
	return fmt.Sprintf(`
resource "aws_neptune_cluster" "test" {
  cluster_identifier_prefix = "tf-test-"
  engine = "neptune"
  neptune_cluster_parameter_group_name = "default.neptune1"
  skip_final_snapshot = true
}
`)
}

func testAccAWSNeptuneClusterConfigWithFinalSnapshot(n int) string {
	return fmt.Sprintf(`
resource "aws_neptune_cluster" "default" {
  cluster_identifier = "tf-neptune-cluster-%d"
  availability_zones = ["us-west-2a","us-west-2b","us-west-2c"]
  neptune_cluster_parameter_group_name = "default.neptune1"
  final_snapshot_identifier = "tf-acctest-neptunecluster-snapshot-%d"
  tags {
    Environment = "production"
  }
}`, n, n)
}
