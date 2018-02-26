package aws

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
)

func TestAccAWSEBSSnapshot_basic(t *testing.T) {
	var v ec2.Snapshot
	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccAwsEbsSnapshotConfig,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSnapshotExists("aws_ebs_snapshot.test", &v),
					testAccCheckTags(&v.Tags, "Name", "testAccAwsEbsSnapshotConfig"),
				),
			},
		},
	})
}

func TestAccAWSEBSSnapshot_withDescription(t *testing.T) {
	var v ec2.Snapshot
	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccAwsEbsSnapshotConfigWithDescription,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSnapshotExists("aws_ebs_snapshot.test", &v),
					resource.TestCheckResourceAttr("aws_ebs_snapshot.test", "description", "EBS Snapshot Acceptance Test"),
				),
			},
		},
	})
}

func TestAccAWSEBSSnapshot_withKms(t *testing.T) {
	var v ec2.Snapshot
	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccAwsEbsSnapshotConfigWithKms,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSnapshotExists("aws_ebs_snapshot.test", &v),
					resource.TestMatchResourceAttr("aws_ebs_snapshot.test", "kms_key_id",
						regexp.MustCompile("^arn:aws:kms:[a-z]{2}-[a-z]+-\\d{1}:[0-9]{12}:key/[a-z0-9-]{36}$")),
				),
			},
		},
	})
}

func testAccCheckSnapshotExists(n string, v *ec2.Snapshot) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No ID is set")
		}

		conn := testAccProvider.Meta().(*AWSClient).ec2conn

		request := &ec2.DescribeSnapshotsInput{
			SnapshotIds: []*string{aws.String(rs.Primary.ID)},
		}

		response, err := conn.DescribeSnapshots(request)
		if err == nil {
			if response.Snapshots != nil && len(response.Snapshots) > 0 {
				*v = *response.Snapshots[0]
				return nil
			}
		}
		return fmt.Errorf("Error finding EC2 Snapshot %s", rs.Primary.ID)
	}
}

const testAccAwsEbsSnapshotConfig = `
resource "aws_ebs_volume" "test" {
  availability_zone = "us-west-2a"
  size              = 1
}

resource "aws_ebs_snapshot" "test" {
  volume_id = "${aws_ebs_volume.test.id}"

  tags {
    Name = "testAccAwsEbsSnapshotConfig"
  }
}
`

const testAccAwsEbsSnapshotConfigWithDescription = `
resource "aws_ebs_volume" "description_test" {
	availability_zone = "us-west-2a"
	size = 1
}

resource "aws_ebs_snapshot" "test" {
	volume_id = "${aws_ebs_volume.description_test.id}"
	description = "EBS Snapshot Acceptance Test"
}
`

const testAccAwsEbsSnapshotConfigWithKms = `
resource "aws_kms_key" "test" {
  deletion_window_in_days = 7

  tags {
    Name = "testAccAwsEbsSnapshotConfigWithKms"
  }
}

resource "aws_ebs_volume" "test" {
  availability_zone = "us-west-2a"
  size              = 1
  encrypted         = true
  kms_key_id        = "${aws_kms_key.test.arn}"

  tags {
    Name = "testAccAwsEbsSnapshotConfigWithKms"
  }
}

resource "aws_ebs_snapshot" "test" {
  volume_id = "${aws_ebs_volume.test.id}"

  tags {
    Name = "testAccAwsEbsSnapshotConfigWithKms"
  }
}
`
