package ec2_test

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/service/ec2"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
)

func TestAccEC2EBSSnapshotIDsDataSource_basic(t *testing.T) {
	dataSourceName := "data.aws_ebs_snapshot_ids.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ec2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccEBSSnapshotIdsDataSourceConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					acctest.CheckResourceAttrGreaterThanValue(dataSourceName, "ids.#", "0"),
					resource.TestCheckTypeSetElemAttrPair(dataSourceName, "ids.*", "aws_ebs_snapshot.test", "id"),
				),
			},
		},
	})
}

func TestAccEC2EBSSnapshotIDsDataSource_sorted(t *testing.T) {
	dataSourceName := "data.aws_ebs_snapshot_ids.test"
	resource1Name := "aws_ebs_snapshot.a"
	resource2Name := "aws_ebs_snapshot.b"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ec2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccEBSSnapshotIdsDataSourceConfig_sorted(rName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(dataSourceName, "ids.#", "2"),
					resource.TestCheckResourceAttrPair(dataSourceName, "ids.0", resource2Name, "id"),
					resource.TestCheckResourceAttrPair(dataSourceName, "ids.1", resource1Name, "id"),
				),
			},
		},
	})
}

func TestAccEC2EBSSnapshotIDsDataSource_empty(t *testing.T) {
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ec2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccEBSSnapshotIdsDataSourceConfig_empty,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("data.aws_ebs_snapshot_ids.empty", "ids.#", "0"),
				),
			},
		},
	})
}

func testAccEBSSnapshotIdsDataSourceConfig_basic(rName string) string {
	return acctest.ConfigCompose(acctest.ConfigAvailableAZsNoOptIn(), fmt.Sprintf(`
resource "aws_ebs_volume" "test" {
  availability_zone = data.aws_availability_zones.available.names[0]
  size              = 1

  tags = {
    Name = %[1]q
  }
}

resource "aws_ebs_snapshot" "test" {
  volume_id = aws_ebs_volume.test.id

  tags = {
    Name = %[1]q
  }
}

data "aws_ebs_snapshot_ids" "test" {
  owners = ["self"]

  depends_on = [aws_ebs_snapshot.test]
}
`, rName))
}

func testAccEBSSnapshotIdsDataSourceConfig_sorted(rName string) string {
	return acctest.ConfigCompose(acctest.ConfigAvailableAZsNoOptIn(), fmt.Sprintf(`
resource "aws_ebs_volume" "test" {
  availability_zone = data.aws_availability_zones.available.names[0]
  size              = 1

  count = 2

  tags = {
    Name = %[1]q
  }
}

resource "aws_ebs_snapshot" "a" {
  volume_id   = aws_ebs_volume.test.*.id[0]
  description = %[1]q

  tags = {
    Name = %[1]q
  }
}

resource "aws_ebs_snapshot" "b" {
  volume_id   = aws_ebs_volume.test.*.id[1]
  description = %[1]q

  # We want to ensure that 'aws_ebs_snapshot.a.creation_date' is less than
  # 'aws_ebs_snapshot.b.creation_date'/ so that we can ensure that the
  # snapshots are being sorted correctly.
  depends_on = [aws_ebs_snapshot.a]

  tags = {
    Name = %[1]q
  }
}

data "aws_ebs_snapshot_ids" "test" {
  owners = ["self"]

  filter {
    name   = "description"
    values = [%[1]q]
  }

  depends_on = [aws_ebs_snapshot.a, aws_ebs_snapshot.b]
}
`, rName))
}

const testAccEBSSnapshotIdsDataSourceConfig_empty = `
data "aws_ebs_snapshot_ids" "empty" {
  owners = ["000000000000"]
}
`
