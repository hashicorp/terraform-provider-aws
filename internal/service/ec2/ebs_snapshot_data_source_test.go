package ec2_test

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
)

func TestAccEC2EBSSnapshotDataSource_basic(t *testing.T) {
	dataSourceName := "data.aws_ebs_snapshot.test"
	resourceName := "aws_ebs_snapshot.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ec2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckEBSSnapshotDataSourceConfig,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckEBSSnapshotIDDataSource(dataSourceName),
					resource.TestCheckResourceAttrPair(dataSourceName, "id", resourceName, "id"),
					resource.TestCheckResourceAttrPair(dataSourceName, "arn", resourceName, "arn"),
					resource.TestCheckResourceAttrPair(dataSourceName, "description", resourceName, "description"),
					resource.TestCheckResourceAttrPair(dataSourceName, "encrypted", resourceName, "encrypted"),
					resource.TestCheckResourceAttrPair(dataSourceName, "kms_key_id", resourceName, "kms_key_id"),
					resource.TestCheckResourceAttrPair(dataSourceName, "owner_alias", resourceName, "owner_alias"),
					resource.TestCheckResourceAttrPair(dataSourceName, "owner_id", resourceName, "owner_id"),
					resource.TestCheckResourceAttrPair(dataSourceName, "tags.%", resourceName, "tags.%"),
					resource.TestCheckResourceAttrPair(dataSourceName, "volume_id", resourceName, "volume_id"),
					resource.TestCheckResourceAttrPair(dataSourceName, "volume_size", resourceName, "volume_size"),
					resource.TestCheckResourceAttrPair(dataSourceName, "storage_tier", resourceName, "storage_tier"),
				),
			},
		},
	})
}

func TestAccEC2EBSSnapshotDataSource_filter(t *testing.T) {
	dataSourceName := "data.aws_ebs_snapshot.test"
	resourceName := "aws_ebs_snapshot.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ec2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckEBSSnapshotFilterDataSourceConfig,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckEBSSnapshotIDDataSource(dataSourceName),
					resource.TestCheckResourceAttrPair(dataSourceName, "id", resourceName, "id"),
				),
			},
		},
	})
}

func TestAccEC2EBSSnapshotDataSource_mostRecent(t *testing.T) {
	dataSourceName := "data.aws_ebs_snapshot.test"
	resourceName := "aws_ebs_snapshot.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ec2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckEBSSnapshotMostRecentDataSourceConfig,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckEBSSnapshotIDDataSource(dataSourceName),
					resource.TestCheckResourceAttrPair(dataSourceName, "id", resourceName, "id"),
				),
			},
		},
	})
}

func testAccCheckEBSSnapshotIDDataSource(n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Can't find snapshot data source: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("Snapshot data source ID not set")
		}
		return nil
	}
}

var testAccCheckEBSSnapshotDataSourceConfig = acctest.ConfigAvailableAZsNoOptIn() + `
resource "aws_ebs_volume" "test" {
  availability_zone = data.aws_availability_zones.available.names[0]
  type              = "gp2"
  size              = 1
}

resource "aws_ebs_snapshot" "test" {
  volume_id = aws_ebs_volume.test.id
}

data "aws_ebs_snapshot" "test" {
  snapshot_ids = [aws_ebs_snapshot.test.id]
}
`

var testAccCheckEBSSnapshotFilterDataSourceConfig = acctest.ConfigAvailableAZsNoOptIn() + `
resource "aws_ebs_volume" "test" {
  availability_zone = data.aws_availability_zones.available.names[0]
  type              = "gp2"
  size              = 1
}

resource "aws_ebs_snapshot" "test" {
  volume_id = aws_ebs_volume.test.id
}

data "aws_ebs_snapshot" "test" {
  filter {
    name   = "snapshot-id"
    values = [aws_ebs_snapshot.test.id]
  }
}
`

var testAccCheckEBSSnapshotMostRecentDataSourceConfig = acctest.ConfigAvailableAZsNoOptIn() + `
resource "aws_ebs_volume" "test" {
  availability_zone = data.aws_availability_zones.available.names[0]
  type              = "gp2"
  size              = 1
}

resource "aws_ebs_snapshot" "incorrect" {
  volume_id = aws_ebs_volume.test.id

  tags = {
    Name = "tf-acc-test-ec2-ebs-snapshot-data-source-most-recent"
  }
}

resource "aws_ebs_snapshot" "test" {
  volume_id = aws_ebs_snapshot.incorrect.volume_id

  tags = {
    Name = "tf-acc-test-ec2-ebs-snapshot-data-source-most-recent"
  }
}

data "aws_ebs_snapshot" "test" {
  most_recent = true

  filter {
    name   = "tag:Name"
    values = [aws_ebs_snapshot.test.tags.Name]
  }
}
`
