package aws

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform/helper/acctest"
	"github.com/hashicorp/terraform/helper/resource"
)

func TestAccAWSInstance_importBasic(t *testing.T) {
	resourceName := "aws_instance.foo"

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckInstanceDestroy,
		Steps: []resource.TestStep{
			resource.TestStep{
				Config: testAccInstanceConfigVPC,
			},

			resource.TestStep{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"associate_public_ip_address", "user_data"},
			},
		},
	})
}

func TestAccAWSInstance_importInDefaultVpc(t *testing.T) {
	resourceName := "aws_instance.foo"
	rInt := acctest.RandInt()

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckInstanceDestroy,
		Steps: []resource.TestStep{
			resource.TestStep{
				Config: testAccInstanceConfigInDefaultVpc(rInt),
			},

			resource.TestStep{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccAWSInstance_importInEc2Classic(t *testing.T) {
	resourceName := "aws_instance.foo"
	rInt := acctest.RandInt()

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccEC2ClassicPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckInstanceDestroy,
		Steps: []resource.TestStep{
			resource.TestStep{
				Config: testAccInstanceConfigInEc2Classic(rInt),
			},

			resource.TestStep{
				Config:                  testAccInstanceConfigInEc2Classic(rInt),
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"network_interface", "source_dest_check"},
			},
		},
	})
}

func testAccInstanceConfigInDefaultVpc(rInt int) string {
	return fmt.Sprintf(`
data "aws_ami" "ubuntu" {
  most_recent = true

  filter {
    name   = "name"
    values = ["ubuntu/images/hvm-ssd/ubuntu-trusty-14.04-amd64-server-*"]
  }

  filter {
    name   = "virtualization-type"
    values = ["hvm"]
  }

  owners = ["099720109477"] # Canonical
}

resource "aws_security_group" "sg" {
  name = "tf_acc_test_%d"
  description = "Test security group"
}

resource "aws_instance" "foo" {
  ami             = "${data.aws_ami.ubuntu.id}"
  instance_type   = "t2.micro"
  security_groups = ["${aws_security_group.sg.name}"]
}
`, rInt)
}

func testAccInstanceConfigInEc2Classic(rInt int) string {
	return fmt.Sprintf(`
provider "aws" {
  region = "us-east-1"
}

data "aws_ami" "ubuntu" {
  most_recent = true

  filter {
    name   = "name"
    values = ["ubuntu/images/ubuntu-trusty-14.04-amd64-server-*"]
  }

  filter {
    name   = "virtualization-type"
    values = ["paravirtual"]
  }

  owners = ["099720109477"] # Canonical
}

resource "aws_security_group" "sg" {
  name = "tf_acc_test_%d"
  description = "Test security group"
}

resource "aws_instance" "foo" {
  ami             = "${data.aws_ami.ubuntu.id}"
  instance_type   = "m3.medium"
  security_groups = ["${aws_security_group.sg.name}"]
}
`, rInt)
}
