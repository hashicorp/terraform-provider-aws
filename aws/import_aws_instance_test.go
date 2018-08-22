package aws

import (
	"fmt"
	"os"
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
			{
				Config: testAccInstanceConfigVPC,
			},

			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"associate_public_ip_address", "user_data"},
			},
		},
	})
}

func TestAccAWSInstance_importInDefaultVpcBySgName(t *testing.T) {
	resourceName := "aws_instance.foo"
	rInt := acctest.RandInt()

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckInstanceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccInstanceConfigInDefaultVpcBySgName(rInt),
			},

			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccAWSInstance_importInDefaultVpcBySgId(t *testing.T) {
	resourceName := "aws_instance.foo"
	rInt := acctest.RandInt()

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckInstanceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccInstanceConfigInDefaultVpcBySgId(rInt),
			},

			{
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

	// EC2 Classic enabled
	oldvar := os.Getenv("AWS_DEFAULT_REGION")
	os.Setenv("AWS_DEFAULT_REGION", "us-east-1")
	defer os.Setenv("AWS_DEFAULT_REGION", oldvar)

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccEC2ClassicPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckInstanceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccInstanceConfigInEc2Classic(rInt),
			},

			{
				Config:                  testAccInstanceConfigInEc2Classic(rInt),
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"network_interface", "source_dest_check"},
			},
		},
	})
}

func testAccInstanceConfigInDefaultVpcBySgName(rInt int) string {
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

data "aws_vpc" "default" {
	default = true
}

resource "aws_security_group" "sg" {
  name = "tf_acc_test_%d"
  description = "Test security group"
	vpc_id = "${data.aws_vpc.default.id}"
}

resource "aws_instance" "foo" {
  ami             = "${data.aws_ami.ubuntu.id}"
  instance_type   = "t2.micro"
  security_groups = ["${aws_security_group.sg.name}"]
}
`, rInt)
}

func testAccInstanceConfigInDefaultVpcBySgId(rInt int) string {
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

data "aws_vpc" "default" {
	default = true
}

resource "aws_security_group" "sg" {
  name = "tf_acc_test_%d"
  description = "Test security group"
	vpc_id = "${data.aws_vpc.default.id}"
}

resource "aws_instance" "foo" {
  ami             = "${data.aws_ami.ubuntu.id}"
  instance_type   = "t2.micro"
  vpc_security_group_ids = ["${aws_security_group.sg.id}"]
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
