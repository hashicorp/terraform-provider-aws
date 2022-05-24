package rds_test

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/endpoints"
	"github.com/aws/aws-sdk-go/service/rds"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfrds "github.com/hashicorp/terraform-provider-aws/internal/service/rds"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

func TestAccRDSClusterActivityStream_basic(t *testing.T) {
	var dbCluster rds.DBCluster
	clusterName := sdkacctest.RandomWithPrefix("tf-testacc-aurora-cluster")
	instanceName := sdkacctest.RandomWithPrefix("tf-testacc-aurora-instance")
	resourceName := "aws_rds_cluster_activity_stream.test"
	rdsClusterResourceName := "aws_rds_cluster.test"
	kmsKeyResourceName := "aws_kms_key.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(t)
			acctest.PreCheckPartition(t, endpoints.AwsPartitionID)
		},
		ErrorCheck:        acctest.ErrorCheck(t, rds.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckClusterActivityStreamDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccClusterActivityStreamConfig(clusterName, instanceName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterActivityStreamExists(resourceName, &dbCluster),
					testAccCheckClusterActivityStreamAttributes(&dbCluster),
					acctest.MatchResourceAttrRegionalARN(resourceName, "resource_arn", "rds", regexp.MustCompile("cluster:"+clusterName)),
					resource.TestCheckResourceAttrPair(resourceName, "resource_arn", rdsClusterResourceName, "arn"),
					resource.TestCheckResourceAttrPair(resourceName, "kms_key_id", kmsKeyResourceName, "key_id"),
					resource.TestCheckResourceAttrSet(resourceName, "kinesis_stream_name"),
					resource.TestCheckResourceAttr(resourceName, "mode", rds.ActivityStreamModeAsync),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"engine_native_audit_fields_included"},
			},
		},
	})
}

func TestAccRDSClusterActivityStream_disappears(t *testing.T) {
	var dbCluster rds.DBCluster
	clusterName := sdkacctest.RandomWithPrefix("tf-testacc-aurora-cluster")
	instanceName := sdkacctest.RandomWithPrefix("tf-testacc-aurora-instance")
	resourceName := "aws_rds_cluster_activity_stream.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(t)
			acctest.PreCheckPartition(t, endpoints.AwsPartitionID)
		},
		ErrorCheck:        acctest.ErrorCheck(t, rds.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckClusterActivityStreamDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccClusterActivityStreamConfig(clusterName, instanceName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterActivityStreamExists(resourceName, &dbCluster),
					acctest.CheckResourceDisappears(acctest.Provider, tfrds.ResourceClusterActivityStream(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckClusterActivityStreamExists(resourceName string, dbCluster *rds.DBCluster) resource.TestCheckFunc {
	return testAccCheckClusterActivityStreamExistsProvider(resourceName, dbCluster, acctest.Provider)
}

func testAccCheckClusterActivityStreamExistsProvider(resourceName string, dbCluster *rds.DBCluster, provider *schema.Provider) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("not found: %s", resourceName)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("DBCluster ID is not set")
		}

		conn := provider.Meta().(*conns.AWSClient).RDSConn

		response, err := tfrds.FindDBClusterWithActivityStream(conn, rs.Primary.ID)

		if err != nil {
			return err
		}

		*dbCluster = *response
		return nil
	}
}

func testAccCheckClusterActivityStreamAttributes(v *rds.DBCluster) resource.TestCheckFunc {
	return func(s *terraform.State) error {

		if aws.StringValue(v.DBClusterArn) == "" {
			return fmt.Errorf("empty RDS Cluster arn")
		}

		if aws.StringValue(v.ActivityStreamKmsKeyId) == "" {
			return fmt.Errorf("empty RDS Cluster activity stream kms key id")
		}

		if aws.StringValue(v.ActivityStreamKinesisStreamName) == "" {
			return fmt.Errorf("empty RDS Cluster activity stream kinesis stream name")
		}

		if aws.StringValue(v.ActivityStreamStatus) != rds.ActivityStreamStatusStarted {
			return fmt.Errorf("incorrect activity stream status: expected: %s, got: %s", rds.ActivityStreamStatusStarted, aws.StringValue(v.ActivityStreamStatus))
		}

		if aws.StringValue(v.ActivityStreamMode) != rds.ActivityStreamModeSync && aws.StringValue(v.ActivityStreamMode) != rds.ActivityStreamModeAsync {
			return fmt.Errorf("incorrect activity stream mode: expected: sync or async, got: %s", aws.StringValue(v.ActivityStreamMode))
		}

		return nil
	}
}

func testAccCheckClusterActivityStreamDestroy(s *terraform.State) error {
	return testAccCheckClusterActivityStreamDestroyWithProvider(s, acctest.Provider)
}

func testAccCheckClusterActivityStreamDestroyWithProvider(s *terraform.State, provider *schema.Provider) error {
	conn := provider.Meta().(*conns.AWSClient).RDSConn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_rds_cluster_activity_stream" {
			continue
		}

		var err error

		_, err = tfrds.FindDBClusterWithActivityStream(conn, rs.Primary.ID)
		if err != nil {
			// Return nil if the cluster is already destroyed
			if tfresource.NotFound(err) {
				return nil
			}
			return err
		}

		return err
	}

	return nil
}

func testAccClusterActivityStreamConfigBase(clusterName, instanceName string) string {
	return acctest.ConfigCompose(acctest.ConfigAvailableAZsNoOptIn(), fmt.Sprintf(`
resource "aws_kms_key" "test" {
  description             = "Testing for AWS RDS Cluster Activity Stream"
  deletion_window_in_days = 7
}

resource "aws_rds_cluster" "test" {
  cluster_identifier  = "%[1]s"
  availability_zones  = ["${data.aws_availability_zones.available.names[0]}", "${data.aws_availability_zones.available.names[1]}", "${data.aws_availability_zones.available.names[2]}"]
  master_username     = "foo"
  master_password     = "mustbeeightcharaters"
  skip_final_snapshot = true
  deletion_protection = false
  engine              = "aurora-postgresql"
  engine_version      = "11.9"
}

resource "aws_rds_cluster_instance" "test" {
  identifier         = "%[2]s"
  cluster_identifier = aws_rds_cluster.test.id
  engine             = aws_rds_cluster.test.engine
  instance_class     = "db.r6g.large"
}
`, clusterName, instanceName))
}

func testAccClusterActivityStreamConfig(clusterName, instanceName string) string {
	return acctest.ConfigCompose(testAccClusterActivityStreamConfigBase(clusterName, instanceName), `
resource "aws_rds_cluster_activity_stream" "test" {
  resource_arn = aws_rds_cluster.test.arn
  kms_key_id   = aws_kms_key.test.key_id
  mode         = "async"

  depends_on = [aws_rds_cluster_instance.test]
}
`)
}
