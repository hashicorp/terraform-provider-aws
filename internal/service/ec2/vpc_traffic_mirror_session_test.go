package ec2_test

import (
	"fmt"
	"regexp"
	"strconv"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfec2 "github.com/hashicorp/terraform-provider-aws/internal/service/ec2"
)

func TestAccVPCTrafficMirrorSession_basic(t *testing.T) {
	var v ec2.TrafficMirrorSession
	resourceName := "aws_ec2_traffic_mirror_session.test"
	description := "test session"
	session := sdkacctest.RandIntRange(1, 32766)
	rName := fmt.Sprintf("tf-acc-test-%s", sdkacctest.RandString(10))
	pLen := sdkacctest.RandIntRange(1, 255)
	vni := sdkacctest.RandIntRange(1, 16777216)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(t)
			testAccPreCheckTrafficMirrorSession(t)
		},
		ErrorCheck:        acctest.ErrorCheck(t, ec2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckTrafficMirrorSessionDestroy,
		Steps: []resource.TestStep{
			//create
			{
				Config: testAccVPCTrafficMirrorSessionConfig_basic(rName, session),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTrafficMirrorSessionExists(resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "description", ""),
					resource.TestCheckResourceAttr(resourceName, "packet_length", "0"),
					resource.TestCheckResourceAttr(resourceName, "session_number", strconv.Itoa(session)),
					resource.TestMatchResourceAttr(resourceName, "virtual_network_id", regexp.MustCompile(`\d+`)),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
					acctest.CheckResourceAttrAccountID(resourceName, "owner_id"),
					acctest.MatchResourceAttrRegionalARN(resourceName, "arn", "ec2", regexp.MustCompile(`traffic-mirror-session/tms-.+`)),
				),
			},
			// update of description, packet length and VNI
			{
				Config: testAccVPCTrafficMirrorSessionConfig_optionals(description, rName, session, pLen, vni),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTrafficMirrorSessionExists(resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "description", description),
					resource.TestCheckResourceAttr(resourceName, "packet_length", strconv.Itoa(pLen)),
					resource.TestCheckResourceAttr(resourceName, "session_number", strconv.Itoa(session)),
					resource.TestCheckResourceAttr(resourceName, "virtual_network_id", strconv.Itoa(vni)),
				),
			},
			// removal of description, packet length and VNI
			{
				Config: testAccVPCTrafficMirrorSessionConfig_basic(rName, session),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTrafficMirrorSessionExists(resourceName, &v),
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

func TestAccVPCTrafficMirrorSession_tags(t *testing.T) {
	var v ec2.TrafficMirrorSession
	resourceName := "aws_ec2_traffic_mirror_session.test"
	session := sdkacctest.RandIntRange(1, 32766)
	rName := fmt.Sprintf("tf-acc-test-%s", sdkacctest.RandString(10))

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(t)
			testAccPreCheckTrafficMirrorSession(t)
		},
		ErrorCheck:        acctest.ErrorCheck(t, ec2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckTrafficMirrorSessionDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccVPCTrafficMirrorSessionConfig_tags1(rName, "key1", "value1", session),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTrafficMirrorSessionExists(resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccVPCTrafficMirrorSessionConfig_tags2(rName, "key1", "value1updated", "key2", "value2", session),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTrafficMirrorSessionExists(resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1updated"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
			{
				Config: testAccVPCTrafficMirrorSessionConfig_tags1(rName, "key2", "value2", session),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTrafficMirrorSessionExists(resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
		},
	})
}

func TestAccVPCTrafficMirrorSession_disappears(t *testing.T) {
	var v ec2.TrafficMirrorSession
	resourceName := "aws_ec2_traffic_mirror_session.test"
	session := sdkacctest.RandIntRange(1, 32766)
	rName := fmt.Sprintf("tf-acc-test-%s", sdkacctest.RandString(10))

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(t)
			testAccPreCheckTrafficMirrorSession(t)
		},
		ErrorCheck:        acctest.ErrorCheck(t, ec2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckTrafficMirrorSessionDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccVPCTrafficMirrorSessionConfig_basic(rName, session),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTrafficMirrorSessionExists(resourceName, &v),
					acctest.CheckResourceDisappears(acctest.Provider, tfec2.ResourceTrafficMirrorSession(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckTrafficMirrorSessionExists(name string, session *ec2.TrafficMirrorSession) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("Not found: %s", name)
		}
		if rs.Primary.ID == "" {
			return fmt.Errorf("No ID set for %s", name)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).EC2Conn
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

		*session = *out.TrafficMirrorSessions[0]

		return nil
	}
}

func testAccTrafficMirrorSessionConfigBase(rName string) string {
	return acctest.ConfigCompose(acctest.ConfigLatestAmazonLinuxHVMEBSAMI(), fmt.Sprintf(`
data "aws_availability_zones" "azs" {
  state = "available"

  filter {
    name   = "opt-in-status"
    values = ["opt-in-not-required"]
  }
}

resource "aws_vpc" "vpc" {
  cidr_block = "10.0.0.0/16"

  tags = {
    Name = %[1]q
  }
}

resource "aws_subnet" "sub1" {
  vpc_id            = aws_vpc.vpc.id
  cidr_block        = "10.0.0.0/24"
  availability_zone = data.aws_availability_zones.azs.names[0]

  tags = {
    Name = %[1]q
  }
}

resource "aws_subnet" "sub2" {
  vpc_id            = aws_vpc.vpc.id
  cidr_block        = "10.0.1.0/24"
  availability_zone = data.aws_availability_zones.azs.names[1]

  tags = {
    Name = %[1]q
  }
}

resource "aws_instance" "src" {
  ami           = data.aws_ami.amzn-ami-minimal-hvm-ebs.id
  instance_type = "m5.large" # m5.large required because only Nitro instances support mirroring
  subnet_id     = aws_subnet.sub1.id

  tags = {
    Name = %[1]q
  }
}

resource "aws_lb" "lb" {
  name               = %[1]q
  internal           = true
  load_balancer_type = "network"
  subnets            = [aws_subnet.sub1.id, aws_subnet.sub2.id]

  enable_deletion_protection = false

  tags = {
    Name        = %[1]q
    Environment = "production"
  }
}

resource "aws_ec2_traffic_mirror_filter" "filter" {
}

resource "aws_ec2_traffic_mirror_target" "target" {
  network_load_balancer_arn = aws_lb.lb.arn
}
`, rName))
}

func testAccVPCTrafficMirrorSessionConfig_basic(rName string, session int) string {
	return acctest.ConfigCompose(testAccTrafficMirrorSessionConfigBase(rName), fmt.Sprintf(`
resource "aws_ec2_traffic_mirror_session" "test" {
  traffic_mirror_filter_id = aws_ec2_traffic_mirror_filter.filter.id
  traffic_mirror_target_id = aws_ec2_traffic_mirror_target.target.id
  network_interface_id     = aws_instance.src.primary_network_interface_id
  session_number           = %d
}
`, session))
}

func testAccVPCTrafficMirrorSessionConfig_tags1(rName, tagKey1, tagValue1 string, session int) string {
	return acctest.ConfigCompose(testAccTrafficMirrorSessionConfigBase(rName), fmt.Sprintf(`
resource "aws_ec2_traffic_mirror_session" "test" {
  traffic_mirror_filter_id = aws_ec2_traffic_mirror_filter.filter.id
  traffic_mirror_target_id = aws_ec2_traffic_mirror_target.target.id
  network_interface_id     = aws_instance.src.primary_network_interface_id
  session_number           = %[3]d

  tags = {
    %[1]q = %[2]q
  }
}
`, tagKey1, tagValue1, session))
}

func testAccVPCTrafficMirrorSessionConfig_tags2(rName, tagKey1, tagValue1, tagKey2, tagValue2 string, session int) string {
	return acctest.ConfigCompose(testAccTrafficMirrorSessionConfigBase(rName), fmt.Sprintf(`
resource "aws_ec2_traffic_mirror_session" "test" {
  traffic_mirror_filter_id = aws_ec2_traffic_mirror_filter.filter.id
  traffic_mirror_target_id = aws_ec2_traffic_mirror_target.target.id
  network_interface_id     = aws_instance.src.primary_network_interface_id
  session_number           = %[5]d

  tags = {
    %[1]q = %[2]q
    %[3]q = %[4]q
  }
}
`, tagKey1, tagValue1, tagKey2, tagValue2, session))
}

func testAccVPCTrafficMirrorSessionConfig_optionals(description string, rName string, session, pLen, vni int) string {
	return acctest.ConfigCompose(testAccTrafficMirrorSessionConfigBase(rName), fmt.Sprintf(`
resource "aws_ec2_traffic_mirror_session" "test" {
  description              = "%s"
  traffic_mirror_filter_id = aws_ec2_traffic_mirror_filter.filter.id
  traffic_mirror_target_id = aws_ec2_traffic_mirror_target.target.id
  network_interface_id     = aws_instance.src.primary_network_interface_id
  session_number           = %d
  packet_length            = %d
  virtual_network_id       = %d
}
`, description, session, pLen, vni))
}

func testAccPreCheckTrafficMirrorSession(t *testing.T) {
	conn := acctest.Provider.Meta().(*conns.AWSClient).EC2Conn

	_, err := conn.DescribeTrafficMirrorSessions(&ec2.DescribeTrafficMirrorSessionsInput{})

	if acctest.PreCheckSkipError(err) {
		t.Skip("skipping traffic mirror sessions acceptance test: ", err)
	}

	if err != nil {
		t.Fatal("Unexpected PreCheck error: ", err)
	}
}

func testAccCheckTrafficMirrorSessionDestroy(s *terraform.State) error {
	conn := acctest.Provider.Meta().(*conns.AWSClient).EC2Conn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_ec2_traffic_mirror_session" {
			continue
		}

		out, err := conn.DescribeTrafficMirrorSessions(&ec2.DescribeTrafficMirrorSessionsInput{
			TrafficMirrorSessionIds: []*string{
				aws.String(rs.Primary.ID),
			},
		})

		if tfawserr.ErrCodeEquals(err, "InvalidTrafficMirrorSessionId.NotFound") {
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
