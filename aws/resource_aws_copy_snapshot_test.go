package aws

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/helper/schema"
	"github.com/hashicorp/terraform/terraform"
)

func TestAccAWSCopySnapshot_basic(t *testing.T) {
	var v ec2.Snapshot
	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccAwsCopySnapshotConfig,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCopySnapshotExists("aws_copy_snapshot.test", &v),
					testAccCheckTags(&v.Tags, "Name", "testAccAwsCopySnapshotConfig"),
				),
			},
		},
	})
}

func TestAccAWSCopySnapshot_withDescription(t *testing.T) {
	var v ec2.Snapshot
	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccAwsCopySnapshotConfigWithDescription,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCopySnapshotExists("aws_copy_snapshot.description_test", &v),
					resource.TestCheckResourceAttr("aws_copy_snapshot.description_test", "description", "Copy Snapshot Acceptance Test"),
				),
			},
		},
	})
}

func TestAccAWSCopySnapshot_withRegions(t *testing.T) {
	var v ec2.Snapshot

	// record the initialized providers so that we can use them to
	// check for the instances in each region
	var providers []*schema.Provider
	providerFactories := map[string]terraform.ResourceProviderFactory{
		"aws": func() (terraform.ResourceProvider, error) {
			p := Provider()
			providers = append(providers, p.(*schema.Provider))
			return p, nil
		},
	}

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: providerFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccAwsCopySnapshotConfigWithRegions,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCopySnapshotExistsWithProviders("aws_copy_snapshot.region_test", &v, &providers),
				),
			},
		},
	})

}

func testAccCheckCopySnapshotExists(n string, v *ec2.Snapshot) resource.TestCheckFunc {
	providers := []*schema.Provider{testAccProvider}
	return testAccCheckCopySnapshotExistsWithProviders(n, v, &providers)
}

func testAccCheckCopySnapshotExistsWithProviders(n string, v *ec2.Snapshot, providers *[]*schema.Provider) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No ID is set")
		}

		for _, provider := range *providers {
			// Ignore if Meta is empty, this can happen for validation providers
			if provider.Meta() == nil {
				continue
			}

			conn := provider.Meta().(*AWSClient).ec2conn

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
		}
		return fmt.Errorf("Error finding EC2 Snapshot %s", rs.Primary.ID)
	}
}

const testAccAwsCopySnapshotConfig = `
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

resource "aws_copy_snapshot" "test" {
  source_snapshot_id = "${aws_ebs_snapshot.test.id}"
	source_region      = "us-west-2"

  tags {
    Name = "testAccAwsCopySnapshotConfig"
  }
}
`

const testAccAwsCopySnapshotConfigWithDescription = `
resource "aws_ebs_volume" "description_test" {
	availability_zone = "us-west-2a"
	size              = 1

  tags {
    Name = "testAccAwsCopySnapshotConfigWithDescription"
  }
}

resource "aws_ebs_snapshot" "description_test" {
	volume_id   = "${aws_ebs_volume.description_test.id}"
	description = "EBS Snapshot Acceptance Test"

  tags {
    Name = "testAccAwsCopySnapshotConfigWithDescription"
  }
}

resource "aws_copy_snapshot" "description_test" {
	description        = "Copy Snapshot Acceptance Test"
  source_snapshot_id = "${aws_ebs_snapshot.description_test.id}"
	source_region      = "us-west-2"

  tags {
    Name = "testAccAwsCopySnapshotConfigWithDescription"
  }
}
`

const testAccAwsCopySnapshotConfigWithRegions = `
provider "aws" {
  region = "us-west-2"
  alias  = "uswest2"
}

provider "aws" {
  region = "us-east-1"
  alias  = "useast1"
}

resource "aws_ebs_volume" "region_test" {
  provider          = "aws.uswest2"
  availability_zone = "us-west-2a"
  size              = 1

  tags {
    Name = "testAccAwsCopySnapshotConfigWithRegions"
  }
}

resource "aws_ebs_snapshot" "region_test" {
  provider  = "aws.uswest2"
  volume_id = "${aws_ebs_volume.region_test.id}"

  tags {
    Name = "testAccAwsCopySnapshotConfigWithRegions"
  }
}

resource "aws_copy_snapshot" "region_test" {
  provider           = "aws.useast1"
  source_snapshot_id = "${aws_ebs_snapshot.region_test.id}"
  source_region      = "us-west-2"

  tags {
    Name = "testAccAwsCopySnapshotConfigWithRegions"
  }
}
`
