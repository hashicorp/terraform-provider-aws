package rds_test

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/service/rds"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
)

func TestAccRDSClusterSnapshotDataSource_dbClusterSnapshotIdentifier(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	dataSourceName := "data.aws_db_cluster_snapshot.test"
	resourceName := "aws_db_cluster_snapshot.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, rds.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccClusterSnapshotDataSourceConfig_clusterSnapshotIdentifier(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterSnapshotExistsDataSource(dataSourceName),
					resource.TestCheckResourceAttrPair(dataSourceName, "allocated_storage", resourceName, "allocated_storage"),
					resource.TestCheckResourceAttrPair(dataSourceName, "availability_zones.#", resourceName, "availability_zones.#"),
					resource.TestCheckResourceAttrPair(dataSourceName, "db_cluster_identifier", resourceName, "db_cluster_identifier"),
					resource.TestCheckResourceAttrPair(dataSourceName, "db_cluster_snapshot_arn", resourceName, "db_cluster_snapshot_arn"),
					resource.TestCheckResourceAttrPair(dataSourceName, "db_cluster_snapshot_identifier", resourceName, "db_cluster_snapshot_identifier"),
					resource.TestCheckResourceAttrPair(dataSourceName, "engine", resourceName, "engine"),
					resource.TestCheckResourceAttrPair(dataSourceName, "engine_version", resourceName, "engine_version"),
					resource.TestCheckResourceAttrPair(dataSourceName, "kms_key_id", resourceName, "kms_key_id"),
					resource.TestCheckResourceAttrPair(dataSourceName, "license_model", resourceName, "license_model"),
					resource.TestCheckResourceAttrPair(dataSourceName, "port", resourceName, "port"),
					resource.TestCheckResourceAttrSet(dataSourceName, "snapshot_create_time"),
					resource.TestCheckResourceAttrPair(dataSourceName, "snapshot_type", resourceName, "snapshot_type"),
					resource.TestCheckResourceAttrPair(dataSourceName, "source_db_cluster_snapshot_arn", resourceName, "source_db_cluster_snapshot_arn"),
					resource.TestCheckResourceAttrPair(dataSourceName, "status", resourceName, "status"),
					resource.TestCheckResourceAttrPair(dataSourceName, "storage_encrypted", resourceName, "storage_encrypted"),
					resource.TestCheckResourceAttrPair(dataSourceName, "vpc_id", resourceName, "vpc_id"),
					resource.TestCheckResourceAttrPair(dataSourceName, "tags", resourceName, "tags"),
				),
			},
		},
	})
}

func TestAccRDSClusterSnapshotDataSource_dbClusterIdentifier(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	dataSourceName := "data.aws_db_cluster_snapshot.test"
	resourceName := "aws_db_cluster_snapshot.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, rds.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccClusterSnapshotDataSourceConfig_clusterIdentifier(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterSnapshotExistsDataSource(dataSourceName),
					resource.TestCheckResourceAttrPair(dataSourceName, "allocated_storage", resourceName, "allocated_storage"),
					resource.TestCheckResourceAttrPair(dataSourceName, "availability_zones.#", resourceName, "availability_zones.#"),
					resource.TestCheckResourceAttrPair(dataSourceName, "db_cluster_identifier", resourceName, "db_cluster_identifier"),
					resource.TestCheckResourceAttrPair(dataSourceName, "db_cluster_snapshot_arn", resourceName, "db_cluster_snapshot_arn"),
					resource.TestCheckResourceAttrPair(dataSourceName, "db_cluster_snapshot_identifier", resourceName, "db_cluster_snapshot_identifier"),
					resource.TestCheckResourceAttrPair(dataSourceName, "engine", resourceName, "engine"),
					resource.TestCheckResourceAttrPair(dataSourceName, "engine_version", resourceName, "engine_version"),
					resource.TestCheckResourceAttrPair(dataSourceName, "kms_key_id", resourceName, "kms_key_id"),
					resource.TestCheckResourceAttrPair(dataSourceName, "license_model", resourceName, "license_model"),
					resource.TestCheckResourceAttrPair(dataSourceName, "port", resourceName, "port"),
					resource.TestCheckResourceAttrSet(dataSourceName, "snapshot_create_time"),
					resource.TestCheckResourceAttrPair(dataSourceName, "snapshot_type", resourceName, "snapshot_type"),
					resource.TestCheckResourceAttrPair(dataSourceName, "source_db_cluster_snapshot_arn", resourceName, "source_db_cluster_snapshot_arn"),
					resource.TestCheckResourceAttrPair(dataSourceName, "status", resourceName, "status"),
					resource.TestCheckResourceAttrPair(dataSourceName, "storage_encrypted", resourceName, "storage_encrypted"),
					resource.TestCheckResourceAttrPair(dataSourceName, "vpc_id", resourceName, "vpc_id"),
					resource.TestCheckResourceAttrPair(dataSourceName, "tags", resourceName, "tags"),
				),
			},
		},
	})
}

func TestAccRDSClusterSnapshotDataSource_mostRecent(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	dataSourceName := "data.aws_db_cluster_snapshot.test"
	resourceName := "aws_db_cluster_snapshot.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, rds.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccClusterSnapshotDataSourceConfig_mostRecent(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterSnapshotExistsDataSource(dataSourceName),
					resource.TestCheckResourceAttrPair(dataSourceName, "db_cluster_snapshot_arn", resourceName, "db_cluster_snapshot_arn"),
					resource.TestCheckResourceAttrPair(dataSourceName, "db_cluster_snapshot_identifier", resourceName, "db_cluster_snapshot_identifier"),
				),
			},
		},
	})
}

func testAccCheckClusterSnapshotExistsDataSource(dataSourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[dataSourceName]
		if !ok {
			return fmt.Errorf("Can't find data source: %s", dataSourceName)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("Snapshot data source ID not set")
		}
		return nil
	}
}

func testAccClusterSnapshotDataSourceConfig_clusterSnapshotIdentifier(rName string) string {
	return acctest.ConfigAvailableAZsNoOptIn() + fmt.Sprintf(`
resource "aws_vpc" "test" {
  cidr_block = "192.168.0.0/16"

  tags = {
    Name = %[1]q
  }
}

resource "aws_subnet" "test" {
  count = 2

  availability_zone = data.aws_availability_zones.available.names[count.index]
  cidr_block        = "192.168.${count.index}.0/24"
  vpc_id            = aws_vpc.test.id

  tags = {
    Name = %[1]q
  }
}

resource "aws_db_subnet_group" "test" {
  name       = %[1]q
  subnet_ids = [aws_subnet.test.*.id[0], aws_subnet.test.*.id[1]]
}

resource "aws_rds_cluster" "test" {
  cluster_identifier   = %[1]q
  db_subnet_group_name = aws_db_subnet_group.test.name
  master_password      = "barbarbarbar"
  master_username      = "foo"
  skip_final_snapshot  = true
}

resource "aws_db_cluster_snapshot" "test" {
  db_cluster_identifier          = aws_rds_cluster.test.id
  db_cluster_snapshot_identifier = %[1]q

  tags = {
    Name = %[1]q
  }
}

data "aws_db_cluster_snapshot" "test" {
  db_cluster_snapshot_identifier = aws_db_cluster_snapshot.test.id
}
`, rName)
}

func testAccClusterSnapshotDataSourceConfig_clusterIdentifier(rName string) string {
	return acctest.ConfigAvailableAZsNoOptIn() + fmt.Sprintf(`
resource "aws_vpc" "test" {
  cidr_block = "192.168.0.0/16"

  tags = {
    Name = %[1]q
  }
}

resource "aws_subnet" "test" {
  count = 2

  availability_zone = data.aws_availability_zones.available.names[count.index]
  cidr_block        = "192.168.${count.index}.0/24"
  vpc_id            = aws_vpc.test.id

  tags = {
    Name = %[1]q
  }
}

resource "aws_db_subnet_group" "test" {
  name       = %[1]q
  subnet_ids = [aws_subnet.test.*.id[0], aws_subnet.test.*.id[1]]
}

resource "aws_rds_cluster" "test" {
  cluster_identifier   = %[1]q
  db_subnet_group_name = aws_db_subnet_group.test.name
  master_password      = "barbarbarbar"
  master_username      = "foo"
  skip_final_snapshot  = true
}

resource "aws_db_cluster_snapshot" "test" {
  db_cluster_identifier          = aws_rds_cluster.test.id
  db_cluster_snapshot_identifier = %[1]q

  tags = {
    Name = %[1]q
  }
}

data "aws_db_cluster_snapshot" "test" {
  db_cluster_identifier = aws_db_cluster_snapshot.test.db_cluster_identifier
}
`, rName)
}

func testAccClusterSnapshotDataSourceConfig_mostRecent(rName string) string {
	return acctest.ConfigAvailableAZsNoOptIn() + fmt.Sprintf(`
resource "aws_vpc" "test" {
  cidr_block = "192.168.0.0/16"

  tags = {
    Name = %[1]q
  }
}

resource "aws_subnet" "test" {
  count = 2

  availability_zone = data.aws_availability_zones.available.names[count.index]
  cidr_block        = "192.168.${count.index}.0/24"
  vpc_id            = aws_vpc.test.id

  tags = {
    Name = %[1]q
  }
}

resource "aws_db_subnet_group" "test" {
  name       = %[1]q
  subnet_ids = [aws_subnet.test.*.id[0], aws_subnet.test.*.id[1]]
}

resource "aws_rds_cluster" "test" {
  cluster_identifier   = %[1]q
  db_subnet_group_name = aws_db_subnet_group.test.name
  master_password      = "barbarbarbar"
  master_username      = "foo"
  skip_final_snapshot  = true
}

resource "aws_db_cluster_snapshot" "incorrect" {
  db_cluster_identifier          = aws_rds_cluster.test.id
  db_cluster_snapshot_identifier = "%[1]s-incorrect"
}

resource "aws_db_cluster_snapshot" "test" {
  db_cluster_identifier          = aws_db_cluster_snapshot.incorrect.db_cluster_identifier
  db_cluster_snapshot_identifier = %[1]q
}

data "aws_db_cluster_snapshot" "test" {
  db_cluster_identifier = aws_db_cluster_snapshot.test.db_cluster_identifier
  most_recent           = true
}
`, rName)
}
