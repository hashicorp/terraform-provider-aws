package aws

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/helper/acctest"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"

	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
)

func TestAccAWSEc2TrafficMirrorTarget_nlb(t *testing.T) {
	resourceName := "aws_ec2_traffic_mirror_target.target"
	description := "test nlb target"
	lbName := acctest.RandString(32)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
			testAccPreCheckAWSEc2TrafficMirrorTarget(t)
		},
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSEc2TrafficMirrorTargetDestroy,
		Steps: []resource.TestStep{
			//create
			{
				Config: testAccTrafficMirrorTargetConfigNlb(description, lbName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSEc2TrafficMirrorTargetExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "description", description),
					resource.TestMatchResourceAttr(resourceName, "network_load_balancer_arn", regexp.MustCompile("arn:aws:elasticloadbalancing:.*")),
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

func TestAccAWSEc2TrafficMirrorTarget_eni(t *testing.T) {
	resourceName := "aws_ec2_traffic_mirror_target.target"
	description := "test eni target"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
			testAccPreCheckAWSEc2TrafficMirrorTarget(t)
		},
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSEc2TrafficMirrorTargetDestroy,
		Steps: []resource.TestStep{
			//create
			{
				Config: testAccTrafficMirrorTargetConfigEni(description),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSEc2TrafficMirrorTargetExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "description", description),
					resource.TestMatchResourceAttr(resourceName, "network_interface_id", regexp.MustCompile("eni-.*")),
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

func testAccCheckAWSEc2TrafficMirrorTargetExists(name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("Not found: %s", name)
		}
		if rs.Primary.ID == "" {
			return fmt.Errorf("No ID set for %s", name)
		}

		conn := testAccProvider.Meta().(*AWSClient).ec2conn
		out, err := conn.DescribeTrafficMirrorTargets(&ec2.DescribeTrafficMirrorTargetsInput{
			TrafficMirrorTargetIds: []*string{
				aws.String(rs.Primary.ID),
			},
		})

		if err != nil {
			return err
		}

		if 0 == len(out.TrafficMirrorTargets) {
			return fmt.Errorf("Traffic mirror target %s not found", rs.Primary.ID)
		}

		return nil
	}
}

func testAccTrafficMirrorTargetConfigNlb(description string, lbName string) string {
	return fmt.Sprintf(`
data "aws_availability_zones" "azs" {
  state = "available"
}

resource "aws_vpc" "vpc" {
  cidr_block = "10.0.0.0/16"
}

resource "aws_subnet" "sub1" {
  vpc_id = "${aws_vpc.vpc.id}"
  cidr_block = "10.0.0.0/24"
  availability_zone = "${data.aws_availability_zones.azs.names[0]}"
}

resource "aws_subnet" "sub2" {
  vpc_id = "${aws_vpc.vpc.id}"
  cidr_block = "10.0.1.0/24"
  availability_zone = "${data.aws_availability_zones.azs.names[1]}"
}

resource "aws_lb" "lb" {
  name               = "%s"
  internal           = true
  load_balancer_type = "network"
  subnets            = ["${aws_subnet.sub1.id}", "${aws_subnet.sub2.id}"]

  enable_deletion_protection  = false

  tags = {
    Environment = "production"
  }
}

resource "aws_ec2_traffic_mirror_target" "target" {
  description = "%s"
  network_load_balancer_arn = "${aws_lb.lb.arn}"
}
`, lbName, description)
}

func testAccTrafficMirrorTargetConfigEni(description string) string {
	return fmt.Sprintf(`
data "aws_availability_zones" "azs" {
  state = "available"
}

data "aws_ami" "amzn-linux" {
  most_recent = true

  filter {
    name = "name"
    values = ["amzn2-ami-hvm-2.0*"]
  }

  filter {
    name = "architecture"
    values = ["x86_64"]
  }

  owners = ["137112412989"]
}

resource "aws_vpc" "vpc" {
  cidr_block = "10.0.0.0/16"
}

resource "aws_subnet" "sub1" {
  vpc_id = "${aws_vpc.vpc.id}"
  cidr_block = "10.0.0.0/24"
  availability_zone = "${data.aws_availability_zones.azs.names[0]}"
}

resource "aws_subnet" "sub2" {
  vpc_id = "${aws_vpc.vpc.id}"
  cidr_block = "10.0.1.0/24"
  availability_zone = "${data.aws_availability_zones.azs.names[1]}"
}

resource "aws_instance" "src" {
  ami = "${data.aws_ami.amzn-linux.id}"
  instance_type = "t2.micro"
  subnet_id = "${aws_subnet.sub1.id}"
}

resource "aws_ec2_traffic_mirror_target" "target" {
  description = "%s"
  network_interface_id = "${aws_instance.src.primary_network_interface_id}"
}
`, description)
}

func testAccPreCheckAWSEc2TrafficMirrorTarget(t *testing.T) {
	conn := testAccProvider.Meta().(*AWSClient).ec2conn

	_, err := conn.DescribeTrafficMirrorTargets(&ec2.DescribeTrafficMirrorTargetsInput{})

	if testAccPreCheckSkipError(err) {
		t.Skip("skipping traffic mirror target acceprance test: ", err)
	}

	if err != nil {
		t.Fatal("Unexpected PreCheck error: ", err)
	}
}

func testAccCheckAWSEc2TrafficMirrorTargetDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*AWSClient).ec2conn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_ec2_traffic_mirror_target" {
			continue
		}

		out, err := conn.DescribeTrafficMirrorTargets(&ec2.DescribeTrafficMirrorTargetsInput{
			TrafficMirrorTargetIds: []*string{
				aws.String(rs.Primary.ID),
			},
		})

		if isAWSErr(err, "InvalidTrafficMirrorTargetId.NotFound", "") {
			continue
		}

		if err != nil {
			return err
		}

		if len(out.TrafficMirrorTargets) != 0 {
			return fmt.Errorf("Traffic mirror target %s still not destroyed", rs.Primary.ID)
		}
	}

	return nil
}
