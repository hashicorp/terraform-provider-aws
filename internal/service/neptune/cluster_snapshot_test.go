package neptune_test

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/neptune"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
)

func TestAccNeptuneClusterSnapshot_basic(t *testing.T) {
	var dbClusterSnapshot neptune.DBClusterSnapshot
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_neptune_cluster_snapshot.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, neptune.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckClusterSnapshotDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccClusterSnapshotConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterSnapshotExists(resourceName, &dbClusterSnapshot),
					resource.TestCheckResourceAttrSet(resourceName, "allocated_storage"),
					resource.TestCheckResourceAttrSet(resourceName, "availability_zones.#"),
					acctest.CheckResourceAttrRegionalARN(resourceName, "db_cluster_snapshot_arn", "rds", fmt.Sprintf("cluster-snapshot:%s", rName)),
					resource.TestCheckResourceAttrSet(resourceName, "engine"),
					resource.TestCheckResourceAttrSet(resourceName, "engine_version"),
					resource.TestCheckResourceAttr(resourceName, "kms_key_id", ""),
					resource.TestCheckResourceAttrSet(resourceName, "license_model"),
					resource.TestCheckResourceAttrSet(resourceName, "port"),
					resource.TestCheckResourceAttr(resourceName, "snapshot_type", "manual"),
					resource.TestCheckResourceAttr(resourceName, "source_db_cluster_snapshot_arn", ""),
					resource.TestCheckResourceAttr(resourceName, "status", "available"),
					resource.TestCheckResourceAttr(resourceName, "storage_encrypted", "false"),
					resource.TestMatchResourceAttr(resourceName, "vpc_id", regexp.MustCompile(`^vpc-.+`)),
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

func testAccCheckClusterSnapshotDestroy(s *terraform.State) error {
	conn := acctest.Provider.Meta().(*conns.AWSClient).NeptuneConn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_neptune_cluster_snapshot" {
			continue
		}

		input := &neptune.DescribeDBClusterSnapshotsInput{
			DBClusterSnapshotIdentifier: aws.String(rs.Primary.ID),
		}

		output, err := conn.DescribeDBClusterSnapshots(input)
		if err != nil {
			if tfawserr.ErrCodeEquals(err, neptune.ErrCodeDBClusterSnapshotNotFoundFault) {
				continue
			}
			return err
		}

		if output != nil && len(output.DBClusterSnapshots) > 0 && output.DBClusterSnapshots[0] != nil && aws.StringValue(output.DBClusterSnapshots[0].DBClusterSnapshotIdentifier) == rs.Primary.ID {
			return fmt.Errorf("Neptune DB Cluster Snapshot %q still exists", rs.Primary.ID)
		}
	}

	return nil
}

func testAccCheckClusterSnapshotExists(resourceName string, dbClusterSnapshot *neptune.DBClusterSnapshot) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("Not found: %s", resourceName)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No ID is set for %s", resourceName)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).NeptuneConn

		input := &neptune.DescribeDBClusterSnapshotsInput{
			DBClusterSnapshotIdentifier: aws.String(rs.Primary.ID),
		}

		output, err := conn.DescribeDBClusterSnapshots(input)
		if err != nil {
			return err
		}

		if output == nil || len(output.DBClusterSnapshots) == 0 || output.DBClusterSnapshots[0] == nil || aws.StringValue(output.DBClusterSnapshots[0].DBClusterSnapshotIdentifier) != rs.Primary.ID {
			return fmt.Errorf("Neptune DB Cluster Snapshot %q not found", rs.Primary.ID)
		}

		*dbClusterSnapshot = *output.DBClusterSnapshots[0]

		return nil
	}
}

func testAccClusterSnapshotConfig(rName string) string {
	return fmt.Sprintf(`
resource "aws_neptune_cluster" "test" {
  cluster_identifier  = %[1]q
  skip_final_snapshot = true
}

resource "aws_neptune_cluster_snapshot" "test" {
  db_cluster_identifier          = aws_neptune_cluster.test.id
  db_cluster_snapshot_identifier = %[1]q
}
`, rName)
}
