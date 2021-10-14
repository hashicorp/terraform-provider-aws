package ec2_test

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/hashicorp/aws-sdk-go-base/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfec2 "github.com/hashicorp/terraform-provider-aws/internal/service/ec2"
)

func TestAccEC2EBSSnapshotCopy_basic(t *testing.T) {
	var snapshot ec2.Snapshot
	resourceName := "aws_ebs_snapshot_copy.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, ec2.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckEbsSnapshotCopyDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccEBSSnapshotCopyConfig,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckEbsSnapshotCopyExists(resourceName, &snapshot),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
					acctest.MatchResourceAttrRegionalARNNoAccount(resourceName, "arn", "ec2", regexp.MustCompile(`snapshot/snap-.+`)),
				),
			},
		},
	})
}

func TestAccEC2EBSSnapshotCopy_tags(t *testing.T) {
	var snapshot ec2.Snapshot
	resourceName := "aws_ebs_snapshot_copy.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, ec2.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckEbsSnapshotCopyDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccEBSSnapshotCopyTags1Config("key1", "value1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckEbsSnapshotCopyExists(resourceName, &snapshot),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1"),
				),
			},
			{
				Config: testAccEBSSnapshotCopyTags2Config("key1", "value1updated", "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckEbsSnapshotCopyExists(resourceName, &snapshot),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "3"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1updated"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
			{
				Config: testAccEBSSnapshotCopyTags1Config("key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckEbsSnapshotCopyExists(resourceName, &snapshot),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
		},
	})
}

func TestAccEC2EBSSnapshotCopy_withDescription(t *testing.T) {
	var snapshot ec2.Snapshot
	resourceName := "aws_ebs_snapshot_copy.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, ec2.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckEbsSnapshotCopyDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccEBSSnapshotCopyWithDescriptionConfig,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckEbsSnapshotCopyExists(resourceName, &snapshot),
					resource.TestCheckResourceAttr(resourceName, "description", "Copy Snapshot Acceptance Test"),
				),
			},
		},
	})
}

func TestAccEC2EBSSnapshotCopy_withRegions(t *testing.T) {
	var providers []*schema.Provider
	var snapshot ec2.Snapshot
	resourceName := "aws_ebs_snapshot_copy.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(t)
			acctest.PreCheckMultipleRegion(t, 2)
		},
		ErrorCheck:        acctest.ErrorCheck(t, ec2.EndpointsID),
		ProviderFactories: acctest.FactoriesAlternate(&providers),
		CheckDestroy:      testAccCheckEbsSnapshotCopyDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccEBSSnapshotCopyWithRegionsConfig,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckEbsSnapshotCopyExists(resourceName, &snapshot),
				),
			},
		},
	})

}

func TestAccEC2EBSSnapshotCopy_withKMS(t *testing.T) {
	var snapshot ec2.Snapshot
	kmsKeyResourceName := "aws_kms_key.test"
	resourceName := "aws_ebs_snapshot_copy.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, ec2.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckEbsSnapshotCopyDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccEBSSnapshotCopyWithKMSConfig,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckEbsSnapshotCopyExists(resourceName, &snapshot),
					resource.TestCheckResourceAttrPair(resourceName, "kms_key_id", kmsKeyResourceName, "arn"),
				),
			},
		},
	})
}

func TestAccEC2EBSSnapshotCopy_disappears(t *testing.T) {
	var snapshot ec2.Snapshot
	resourceName := "aws_ebs_snapshot_copy.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, ec2.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckEbsSnapshotCopyDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccEBSSnapshotCopyConfig,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckEbsSnapshotCopyExists(resourceName, &snapshot),
					acctest.CheckResourceDisappears(acctest.Provider, tfec2.ResourceEBSSnapshotCopy(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckEbsSnapshotCopyDestroy(s *terraform.State) error {
	conn := acctest.Provider.Meta().(*conns.AWSClient).EC2Conn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_ebs_snapshot_copy" {
			continue
		}

		resp, err := conn.DescribeSnapshots(&ec2.DescribeSnapshotsInput{
			SnapshotIds: []*string{aws.String(rs.Primary.ID)},
		})

		if tfawserr.ErrMessageContains(err, "InvalidSnapshot.NotFound", "") {
			continue
		}

		if err == nil {
			for _, snapshot := range resp.Snapshots {
				if aws.StringValue(snapshot.SnapshotId) == rs.Primary.ID {
					return fmt.Errorf("EBS Snapshot still exists")
				}
			}
		}

		return err
	}

	return nil
}

func testAccCheckEbsSnapshotCopyExists(n string, v *ec2.Snapshot) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No ID is set")
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).EC2Conn

		input := &ec2.DescribeSnapshotsInput{
			SnapshotIds: []*string{aws.String(rs.Primary.ID)},
		}

		output, err := conn.DescribeSnapshots(input)

		if err != nil {
			return err
		}

		if output == nil || len(output.Snapshots) == 0 {
			return fmt.Errorf("Error finding EC2 Snapshot %s", rs.Primary.ID)
		}

		*v = *output.Snapshots[0]

		return nil
	}
}

const testAccEBSSnapshotCopyConfig = `
data "aws_availability_zones" "available" {
  state = "available"

  filter {
    name   = "opt-in-status"
    values = ["opt-in-not-required"]
  }
}

data "aws_region" "current" {}

resource "aws_ebs_volume" "test" {
  availability_zone = data.aws_availability_zones.available.names[0]
  size              = 1
}

resource "aws_ebs_snapshot" "test" {
  volume_id = aws_ebs_volume.test.id

  tags = {
    Name = "testAccEBSSnapshotCopyConfig"
  }
}

resource "aws_ebs_snapshot_copy" "test" {
  source_snapshot_id = aws_ebs_snapshot.test.id
  source_region      = data.aws_region.current.name
}
`

func testAccEBSSnapshotCopyTags1Config(tagKey1, tagValue1 string) string {
	return fmt.Sprintf(`
data "aws_availability_zones" "available" {
  state = "available"

  filter {
    name   = "opt-in-status"
    values = ["opt-in-not-required"]
  }
}

data "aws_region" "current" {
}

resource "aws_ebs_volume" "test" {
  availability_zone = data.aws_availability_zones.available.names[0]
  size              = 1
}

resource "aws_ebs_snapshot" "test" {
  volume_id = aws_ebs_volume.test.id

  tags = {
    Name = "testAccEBSSnapshotCopyConfig"
  }
}

resource "aws_ebs_snapshot_copy" "test" {
  source_snapshot_id = aws_ebs_snapshot.test.id
  source_region      = data.aws_region.current.name

  tags = {
    Name = "testAccEBSSnapshotCopyConfig"
    "%s" = "%s"
  }
}
`, tagKey1, tagValue1)
}

func testAccEBSSnapshotCopyTags2Config(tagKey1, tagValue1, tagKey2, tagValue2 string) string {
	return fmt.Sprintf(`
data "aws_availability_zones" "available" {
  state = "available"

  filter {
    name   = "opt-in-status"
    values = ["opt-in-not-required"]
  }
}

data "aws_region" "current" {
}

resource "aws_ebs_volume" "test" {
  availability_zone = data.aws_availability_zones.available.names[0]
  size              = 1
}

resource "aws_ebs_snapshot" "test" {
  volume_id = aws_ebs_volume.test.id

  tags = {
    Name = "testAccEBSSnapshotCopyConfig"
  }
}

resource "aws_ebs_snapshot_copy" "test" {
  source_snapshot_id = aws_ebs_snapshot.test.id
  source_region      = data.aws_region.current.name

  tags = {
    Name = "testAccEBSSnapshotCopyConfig"
    "%s" = "%s"
    "%s" = "%s"
  }
}
`, tagKey1, tagValue1, tagKey2, tagValue2)
}

const testAccEBSSnapshotCopyWithDescriptionConfig = `
data "aws_availability_zones" "available" {
  state = "available"

  filter {
    name   = "opt-in-status"
    values = ["opt-in-not-required"]
  }
}

data "aws_region" "current" {}

resource "aws_ebs_volume" "test" {
  availability_zone = data.aws_availability_zones.available.names[0]
  size              = 1

  tags = {
    Name = "testAccEBSSnapshotCopyWithDescriptionConfig"
  }
}

resource "aws_ebs_snapshot" "test" {
  volume_id   = aws_ebs_volume.test.id
  description = "EBS Snapshot Acceptance Test"

  tags = {
    Name = "testAccEBSSnapshotCopyWithDescriptionConfig"
  }
}

resource "aws_ebs_snapshot_copy" "test" {
  description        = "Copy Snapshot Acceptance Test"
  source_snapshot_id = aws_ebs_snapshot.test.id
  source_region      = data.aws_region.current.name

  tags = {
    Name = "testAccEBSSnapshotCopyWithDescriptionConfig"
  }
}
`

var testAccEBSSnapshotCopyWithRegionsConfig = acctest.ConfigAlternateRegionProvider() + `
data "aws_availability_zones" "alternate_available" {
  provider = "awsalternate"
  state    = "available"

  filter {
    name   = "opt-in-status"
    values = ["opt-in-not-required"]
  }
}

data "aws_region" "alternate" {
  provider = "awsalternate"
}

resource "aws_ebs_volume" "test" {
  provider          = "awsalternate"
  availability_zone = data.aws_availability_zones.alternate_available.names[0]
  size              = 1

  tags = {
    Name = "testAccEBSSnapshotCopyWithRegionsConfig"
  }
}

resource "aws_ebs_snapshot" "test" {
  provider  = "awsalternate"
  volume_id = aws_ebs_volume.test.id

  tags = {
    Name = "testAccEBSSnapshotCopyWithRegionsConfig"
  }
}

resource "aws_ebs_snapshot_copy" "test" {
  source_snapshot_id = aws_ebs_snapshot.test.id
  source_region      = data.aws_region.alternate.name

  tags = {
    Name = "testAccEBSSnapshotCopyWithRegionsConfig"
  }
}
`

const testAccEBSSnapshotCopyWithKMSConfig = `
data "aws_availability_zones" "available" {
  state = "available"

  filter {
    name   = "opt-in-status"
    values = ["opt-in-not-required"]
  }
}

data "aws_region" "current" {}

resource "aws_kms_key" "test" {
  description             = "testAccEBSSnapshotCopyWithKMSConfig"
  deletion_window_in_days = 7
}

resource "aws_ebs_volume" "test" {
  availability_zone = data.aws_availability_zones.available.names[0]
  size              = 1

  tags = {
    Name = "testAccEBSSnapshotCopyWithKMSConfig"
  }
}

resource "aws_ebs_snapshot" "test" {
  volume_id = aws_ebs_volume.test.id

  tags = {
    Name = "testAccEBSSnapshotCopyWithKMSConfig"
  }
}

resource "aws_ebs_snapshot_copy" "test" {
  source_snapshot_id = aws_ebs_snapshot.test.id
  source_region      = data.aws_region.current.name
  encrypted          = true
  kms_key_id         = aws_kms_key.test.arn

  tags = {
    Name = "testAccEBSSnapshotCopyWithKMSConfig"
  }
}
`
