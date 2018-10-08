package aws

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/hashicorp/terraform/helper/acctest"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
)

func TestAccAWSAMICopy_basic(t *testing.T) {
	var image ec2.Image
	resourceName := "aws_ami_copy.test"

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSAMICopyDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSAMICopyConfig,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSAMICopyExists(resourceName, &image),
					testAccCheckAWSAMICopyAttributes(&image),
				),
			},
		},
	})
}

func TestAccAWSAMICopy_EnaSupport(t *testing.T) {
	var image ec2.Image
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_ami_copy.test"

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSAMICopyDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSAMICopyConfig_ENASupport(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSAMICopyExists(resourceName, &image),
					resource.TestCheckResourceAttr(resourceName, "ena_support", "true"),
				),
			},
		},
	})
}

func testAccCheckAWSAMICopyExists(resourceName string, image *ec2.Image) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("Not found: %s", resourceName)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No ID set for %s", resourceName)
		}

		conn := testAccProvider.Meta().(*AWSClient).ec2conn

		input := &ec2.DescribeImagesInput{
			ImageIds: []*string{aws.String(rs.Primary.ID)},
		}
		output, err := conn.DescribeImages(input)
		if err != nil {
			return err
		}

		if len(output.Images) == 0 || aws.StringValue(output.Images[0].ImageId) != rs.Primary.ID {
			return fmt.Errorf("AMI %q not found", rs.Primary.ID)
		}

		*image = *output.Images[0]

		return nil
	}
}

func testAccCheckAWSAMICopyDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*AWSClient).ec2conn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_ami_copy" {
			continue
		}

		input := &ec2.DescribeImagesInput{
			ImageIds: []*string{aws.String(rs.Primary.ID)},
		}
		output, err := conn.DescribeImages(input)
		if err != nil {
			return err
		}

		if output != nil && len(output.Images) > 0 && aws.StringValue(output.Images[0].ImageId) == rs.Primary.ID {
			return fmt.Errorf("AMI %q still exists in state: %s", rs.Primary.ID, aws.StringValue(output.Images[0].State))
		}
	}

	// Check for managed EBS snapshots
	return testAccCheckAWSEbsSnapshotDestroy(s)
}

func testAccCheckAWSAMICopyAttributes(image *ec2.Image) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if expected := "available"; aws.StringValue(image.State) != expected {
			return fmt.Errorf("invalid image state; expected %s, got %s", expected, aws.StringValue(image.State))
		}
		if expected := "machine"; aws.StringValue(image.ImageType) != expected {
			return fmt.Errorf("wrong image type; expected %s, got %s", expected, aws.StringValue(image.ImageType))
		}
		if expected := "terraform-acc-ami-copy"; aws.StringValue(image.Name) != expected {
			return fmt.Errorf("wrong name; expected %s, got %s", expected, aws.StringValue(image.Name))
		}

		snapshots := []string{}
		for _, bdm := range image.BlockDeviceMappings {
			// The snapshot ID might not be set,
			// even for a block device that is an
			// EBS volume.
			if bdm.Ebs != nil && bdm.Ebs.SnapshotId != nil {
				snapshots = append(snapshots, aws.StringValue(bdm.Ebs.SnapshotId))
			}
		}

		if expected := 1; len(snapshots) != expected {
			return fmt.Errorf("wrong number of snapshots; expected %v, got %v", expected, len(snapshots))
		}

		return nil
	}
}

var testAccAWSAMICopyConfig = `
provider "aws" {
	region = "us-east-1"
}

// An AMI can't be directly copied from one account to another, and
// we can't rely on any particular AMI being available since anyone
// can run this test in whatever account they like.
// Therefore we jump through some hoops here:
//  - Spin up an EC2 instance based on a public AMI
//  - Create an AMI by snapshotting that EC2 instance, using
//    aws_ami_from_instance .
//  - Copy the new AMI using aws_ami_copy .
//
// Thus this test can only succeed if the aws_ami_from_instance resource
// is working. If it's misbehaving it will likely cause this test to fail too.

// Since we're booting a t2.micro HVM instance we need a VPC for it to boot
// up into.

resource "aws_vpc" "foo" {
	cidr_block = "10.1.0.0/16"
	tags {
		Name = "terraform-testacc-ami-copy"
	}
}

resource "aws_subnet" "foo" {
	cidr_block = "10.1.1.0/24"
	vpc_id = "${aws_vpc.foo.id}"
	tags {
		Name = "tf-acc-ami-copy"
	}
}

resource "aws_instance" "test" {
    // This AMI has one block device mapping, so we expect to have
    // one snapshot in our created AMI.
    // This is an Ubuntu Linux HVM AMI. A public HVM AMI is required
    // because paravirtual images cannot be copied between accounts.
    ami = "ami-0f8bce65"
    instance_type = "t2.micro"
    tags {
        Name = "terraform-acc-ami-copy-victim"
    }

    subnet_id = "${aws_subnet.foo.id}"
}

resource "aws_ami_from_instance" "test" {
    name = "terraform-acc-ami-copy-victim"
    description = "Testing Terraform aws_ami_from_instance resource"
    source_instance_id = "${aws_instance.test.id}"
}

resource "aws_ami_copy" "test" {
    name = "terraform-acc-ami-copy"
    description = "Testing Terraform aws_ami_copy resource"
    source_ami_id = "${aws_ami_from_instance.test.id}"
    source_ami_region = "us-east-1"
}
`

func testAccAWSAMICopyConfig_ENASupport(rName string) string {
	return fmt.Sprintf(`
data "aws_availability_zones" "available" {}
data "aws_region" "current" {}

resource "aws_ebs_volume" "test" {
  availability_zone = "${data.aws_availability_zones.available.names[0]}"
  size              = 1

  tags {
    Name = %q
  }
}

resource "aws_ebs_snapshot" "test" {
  volume_id = "${aws_ebs_volume.test.id}"

  tags {
    Name = %q
  }
}

resource "aws_ami" "test" {
  ena_support         = true
  name                = "%s-source"
  virtualization_type = "hvm"
  root_device_name    = "/dev/sda1"

  ebs_block_device {
    device_name = "/dev/sda1"
    snapshot_id = "${aws_ebs_snapshot.test.id}"
  }
}

resource "aws_ami_copy" "test" {
    name              = "%s-copy"
    source_ami_id     = "${aws_ami.test.id}"
    source_ami_region = "${data.aws_region.current.name}"
}
`, rName, rName, rName, rName)
}
