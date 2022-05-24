package docdb_test

import (
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
)

func TestAccDocDBClusterSnapshot_basic(t *testing.T) {
	var dbClusterSnapshot docdb.DBClusterSnapshot
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_docdb_cluster_snapshot.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, docdb.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckClusterSnapshotDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccClusterSnapshotConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterSnapshotExists(resourceName, &dbClusterSnapshot),
					resource.TestCheckResourceAttrSet(resourceName, "availability_zones.#"),
					acctest.MatchResourceAttrRegionalARN(resourceName, "db_cluster_snapshot_arn", "rds", regexp.MustCompile(`cluster-snapshot:.+`)),
					resource.TestCheckResourceAttrSet(resourceName, "engine"),
					resource.TestCheckResourceAttrSet(resourceName, "engine_version"),
					resource.TestCheckResourceAttr(resourceName, "kms_key_id", ""),
					resource.TestCheckResourceAttrSet(resourceName, "port"),
					resource.TestCheckResourceAttr(resourceName, "snapshot_type", "manual"),
					resource.TestCheckResourceAttr(resourceName, "source_db_cluster_snapshot_arn", ""),
					resource.TestCheckResourceAttr(resourceName, "status", "available"),
					resource.TestCheckResourceAttr(resourceName, "storage_encrypted", "false"),
					resource.TestCheckResourceAttrPair(resourceName, "vpc_id", "aws_vpc.test", "id"),
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
	conn := acctest.Provider.Meta().(*conns.AWSClient).DocDBConn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_docdb_cluster_snapshot" {
			continue
		}

		input := &docdb.DescribeDBClusterSnapshotsInput{
			DBClusterSnapshotIdentifier: aws.String(rs.Primary.ID),
		}

		output, err := conn.DescribeDBClusterSnapshots(input)
		if err != nil {
			if tfawserr.ErrCodeEquals(err, docdb.ErrCodeDBClusterSnapshotNotFoundFault) {
				continue
			}
			return err
		}

		if output != nil && len(output.DBClusterSnapshots) > 0 && output.DBClusterSnapshots[0] != nil && aws.StringValue(output.DBClusterSnapshots[0].DBClusterSnapshotIdentifier) == rs.Primary.ID {
			return fmt.Errorf("DocDB Cluster Snapshot %q still exists", rs.Primary.ID)
		}
	}

	return nil
}

func testAccCheckClusterSnapshotExists(resourceName string, dbClusterSnapshot *docdb.DBClusterSnapshot) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("Not found: %s", resourceName)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No ID is set for %s", resourceName)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).DocDBConn

		request := &docdb.DescribeDBClusterSnapshotsInput{
			DBClusterSnapshotIdentifier: aws.String(rs.Primary.ID),
		}

		response, err := conn.DescribeDBClusterSnapshots(request)
		if err != nil {
			return err
		}

		if response == nil || len(response.DBClusterSnapshots) == 0 || response.DBClusterSnapshots[0] == nil || aws.StringValue(response.DBClusterSnapshots[0].DBClusterSnapshotIdentifier) != rs.Primary.ID {
			return fmt.Errorf("DocDB Cluster Snapshot %q not found", rs.Primary.ID)
		}

		*dbClusterSnapshot = *response.DBClusterSnapshots[0]

		return nil
	}
}

func testAccClusterSnapshotConfig(rName string) string {
	return fmt.Sprintf(`
data "aws_availability_zones" "available" {
  state = "available"

  filter {
    name   = "opt-in-status"
    values = ["opt-in-not-required"]
  }
}

resource "aws_vpc" "test" {
  cidr_block = "192.168.0.0/16"

  tags = {
    Name = %q
  }
}

resource "aws_subnet" "test" {
  count = 2

  availability_zone = data.aws_availability_zones.available.names[count.index]
  cidr_block        = "192.168.${count.index}.0/24"
  vpc_id            = aws_vpc.test.id

  tags = {
    Name = %q
  }
}

resource "aws_docdb_subnet_group" "test" {
  name       = %q
  subnet_ids = aws_subnet.test[*].id
}

resource "aws_docdb_cluster" "test" {
  cluster_identifier   = %q
  db_subnet_group_name = aws_docdb_subnet_group.test.name
  master_password      = "barbarbarbar"
  master_username      = "foo"
  skip_final_snapshot  = true
}

resource "aws_docdb_cluster_snapshot" "test" {
  db_cluster_identifier          = aws_docdb_cluster.test.id
  db_cluster_snapshot_identifier = %q
}
`, rName, rName, rName, rName, rName)
}
