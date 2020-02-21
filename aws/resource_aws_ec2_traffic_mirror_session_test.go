package aws

import (
	"fmt"
	"regexp"
	"strconv"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/helper/acctest"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"

	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
)

func TestAccAWSEc2TrafficMirrorSession_basic(t *testing.T) {
	resourceName := "aws_ec2_traffic_mirror_session.session"
	description := "test session"
	session := acctest.RandIntRange(1, 32766)
	lbName := acctest.RandString(31)
	pLen := acctest.RandIntRange(1, 255)
	vni := acctest.RandIntRange(1, 16777216)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
			testAccPreCheckAWSEc2TrafficMirrorSession(t)
		},
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSEc2TrafficMirrorSessionDestroy,
		Steps: []resource.TestStep{
			//create
			{
				Config: testAccTrafficMirrorSessionConfig(lbName, session),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSEc2TrafficMirrorSessionExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "description", ""),
					resource.TestCheckResourceAttr(resourceName, "packet_length", "0"),
					resource.TestCheckResourceAttr(resourceName, "session_number", strconv.Itoa(session)),
					resource.TestMatchResourceAttr(resourceName, "virtual_network_id", regexp.MustCompile(`\d+`)),
				),
			},
			// update of description, packet length and VNI
			{
				Config: testAccTrafficMirrorSessionConfigWithOptionals(description, lbName, session, pLen, vni),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSEc2TrafficMirrorSessionExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "description", description),
					resource.TestCheckResourceAttr(resourceName, "packet_length", strconv.Itoa(pLen)),
					resource.TestCheckResourceAttr(resourceName, "session_number", strconv.Itoa(session)),
					resource.TestCheckResourceAttr(resourceName, "virtual_network_id", strconv.Itoa(vni)),
				),
			},
			// removal of description, packet length and VNI
			{
				Config: testAccTrafficMirrorSessionConfig(lbName, session),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSEc2TrafficMirrorSessionExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "description", ""),
					resource.TestCheckResourceAttr(resourceName, "packet_length", "0"),
					resource.TestCheckResourceAttr(resourceName, "session_number", strconv.Itoa(session)),
					resource.TestMatchResourceAttr(resourceName, "virtual_network_id", regexp.MustCompile(`\d+`)),
				),
			},
			// import test without VNI
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccCheckAWSEc2TrafficMirrorSessionExists(name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("Not found: %s", name)
		}
		if rs.Primary.ID == "" {
			return fmt.Errorf("No ID set for %s", name)
		}

		conn := testAccProvider.Meta().(*AWSClient).ec2conn
		out, err := conn.DescribeTrafficMirrorSessions(&ec2.DescribeTrafficMirrorSessionsInput{
			TrafficMirrorSessionIds: []*string{
				aws.String(rs.Primary.ID),
			},
		})

		if err != nil {
			return err
		}

		if 0 == len(out.TrafficMirrorSessions) {
			return fmt.Errorf("Traffic mirror session %s not found", rs.Primary.ID)
		}

		return nil
	}
}

func testAccTrafficMirrorSessionConfigBase(lbName string) string {
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
  instance_type = "m5.large" # m5.large required because only Nitro instances support mirroring
  subnet_id = "${aws_subnet.sub1.id}"
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

resource "aws_ec2_traffic_mirror_filter" "filter" {
}

resource "aws_ec2_traffic_mirror_target" "target" {
  network_load_balancer_arn = "${aws_lb.lb.arn}"
}

`, lbName)
}

func testAccTrafficMirrorSessionConfig(lbName string, session int) string {
	return fmt.Sprintf(`
%s

resource "aws_ec2_traffic_mirror_session" "session" {
  traffic_mirror_filter_id = "${aws_ec2_traffic_mirror_filter.filter.id}"
  traffic_mirror_target_id = "${aws_ec2_traffic_mirror_target.target.id}"
  network_interface_id = "${aws_instance.src.primary_network_interface_id}"
  session_number = %d
}
`, testAccTrafficMirrorSessionConfigBase(lbName), session)
}

func testAccTrafficMirrorSessionConfigWithOptionals(description string, lbName string, session, pLen, vni int) string {
	return fmt.Sprintf(`
%s

resource "aws_ec2_traffic_mirror_session" "session" {
  description = "%s"
  traffic_mirror_filter_id = "${aws_ec2_traffic_mirror_filter.filter.id}"
  traffic_mirror_target_id = "${aws_ec2_traffic_mirror_target.target.id}"
  network_interface_id = "${aws_instance.src.primary_network_interface_id}"
  session_number = %d
  packet_length = %d
  virtual_network_id = %d
}
`, testAccTrafficMirrorSessionConfigBase(lbName), description, session, pLen, vni)
}

func testAccPreCheckAWSEc2TrafficMirrorSession(t *testing.T) {
	conn := testAccProvider.Meta().(*AWSClient).ec2conn

	_, err := conn.DescribeTrafficMirrorSessions(&ec2.DescribeTrafficMirrorSessionsInput{})

	if testAccPreCheckSkipError(err) {
		t.Skip("skipping traffic mirror sessions acceprance test: ", err)
	}

	if err != nil {
		t.Fatal("Unexpected PreCheck error: ", err)
	}
}

func testAccCheckAWSEc2TrafficMirrorSessionDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*AWSClient).ec2conn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_ec2_traffic_mirror_session" {
			continue
		}

		out, err := conn.DescribeTrafficMirrorSessions(&ec2.DescribeTrafficMirrorSessionsInput{
			TrafficMirrorSessionIds: []*string{
				aws.String(rs.Primary.ID),
			},
		})

		if isAWSErr(err, "InvalidTrafficMirrorSessionId.NotFound", "") {
			continue
		}

		if err != nil {
			return err
		}

		if len(out.TrafficMirrorSessions) != 0 {
			return fmt.Errorf("Traffic mirror session %s still not destroyed", rs.Primary.ID)
		}
	}

	return nil
}
