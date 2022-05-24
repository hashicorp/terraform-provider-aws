package ec2_test

import (
	"fmt"
	"regexp"
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

func TestAccVPCTrafficMirrorTarget_nlb(t *testing.T) {
	var v ec2.TrafficMirrorTarget
	resourceName := "aws_ec2_traffic_mirror_target.test"
	description := "test nlb target"
	rName := fmt.Sprintf("tf-acc-test-%s", sdkacctest.RandString(10))

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(t)
			testAccPreCheckTrafficMirrorTarget(t)
		},
		ErrorCheck:        acctest.ErrorCheck(t, ec2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckTrafficMirrorTargetDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccVPCTrafficMirrorTargetConfig_nlb(rName, description),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTrafficMirrorTargetExists(resourceName, &v),
					acctest.CheckResourceAttrAccountID(resourceName, "owner_id"),
					acctest.MatchResourceAttrRegionalARN(resourceName, "arn", "ec2", regexp.MustCompile(`traffic-mirror-target/tmt-.+`)),
					resource.TestCheckResourceAttr(resourceName, "description", description),
					resource.TestCheckResourceAttrPair(resourceName, "network_load_balancer_arn", "aws_lb.lb", "arn"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
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

func TestAccVPCTrafficMirrorTarget_eni(t *testing.T) {
	var v ec2.TrafficMirrorTarget
	resourceName := "aws_ec2_traffic_mirror_target.test"
	rName := fmt.Sprintf("tf-acc-test-%s", sdkacctest.RandString(10))
	description := "test eni target"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(t)
			testAccPreCheckTrafficMirrorTarget(t)
		},
		ErrorCheck:        acctest.ErrorCheck(t, ec2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckTrafficMirrorTargetDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccVPCTrafficMirrorTargetConfig_eni(rName, description),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTrafficMirrorTargetExists(resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "description", description),
					resource.TestMatchResourceAttr(resourceName, "network_interface_id", regexp.MustCompile("eni-.*")),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
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

func TestAccVPCTrafficMirrorTarget_tags(t *testing.T) {
	var v ec2.TrafficMirrorTarget
	resourceName := "aws_ec2_traffic_mirror_target.test"
	description := "test nlb target"
	rName := fmt.Sprintf("tf-acc-test-%s", sdkacctest.RandString(10))

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(t)
			testAccPreCheckTrafficMirrorTarget(t)
		},
		ErrorCheck:        acctest.ErrorCheck(t, ec2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckTrafficMirrorTargetDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccVPCTrafficMirrorTargetConfig_tags1(rName, description, "key1", "value1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTrafficMirrorTargetExists(resourceName, &v),
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
				Config: testAccVPCTrafficMirrorTargetConfig_tags2(rName, description, "key1", "value1updated", "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTrafficMirrorTargetExists(resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1updated"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
			{
				Config: testAccVPCTrafficMirrorTargetConfig_tags1(rName, description, "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTrafficMirrorTargetExists(resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
		},
	})
}

func TestAccVPCTrafficMirrorTarget_disappears(t *testing.T) {
	var v ec2.TrafficMirrorTarget
	resourceName := "aws_ec2_traffic_mirror_target.test"
	description := "test nlb target"
	rName := fmt.Sprintf("tf-acc-test-%s", sdkacctest.RandString(10))

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(t)
			testAccPreCheckTrafficMirrorTarget(t)
		},
		ErrorCheck:        acctest.ErrorCheck(t, ec2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckTrafficMirrorTargetDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccVPCTrafficMirrorTargetConfig_nlb(rName, description),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTrafficMirrorTargetExists(resourceName, &v),
					acctest.CheckResourceDisappears(acctest.Provider, tfec2.ResourceTrafficMirrorTarget(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckTrafficMirrorTargetExists(name string, target *ec2.TrafficMirrorTarget) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("Not found: %s", name)
		}
		if rs.Primary.ID == "" {
			return fmt.Errorf("No ID set for %s", name)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).EC2Conn
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

		*target = *out.TrafficMirrorTargets[0]

		return nil
	}
}

func testAccTrafficMirrorTargetConfigBase(rName string) string {
	return fmt.Sprintf(`
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
`, rName)
}

func testAccVPCTrafficMirrorTargetConfig_nlb(rName, description string) string {
	return acctest.ConfigCompose(testAccTrafficMirrorTargetConfigBase(rName), fmt.Sprintf(`
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

resource "aws_ec2_traffic_mirror_target" "test" {
  description               = %[2]q
  network_load_balancer_arn = aws_lb.lb.arn
}
`, rName, description))
}

func testAccVPCTrafficMirrorTargetConfig_eni(rName, description string) string {
	return acctest.ConfigCompose(
		testAccTrafficMirrorTargetConfigBase(rName),
		acctest.ConfigLatestAmazonLinuxHVMEBSAMI(),
		fmt.Sprintf(`
resource "aws_instance" "src" {
  ami           = data.aws_ami.amzn-ami-minimal-hvm-ebs.id
  instance_type = "t2.micro"
  subnet_id     = aws_subnet.sub1.id

  tags = {
    Name = %[1]q
  }
}

resource "aws_ec2_traffic_mirror_target" "test" {
  description          = %[2]q
  network_interface_id = aws_instance.src.primary_network_interface_id
}
`, rName, description))
}

func testAccVPCTrafficMirrorTargetConfig_tags1(rName, description, tagKey1, tagValue1 string) string {
	return acctest.ConfigCompose(testAccTrafficMirrorTargetConfigBase(rName), fmt.Sprintf(`
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

resource "aws_ec2_traffic_mirror_target" "test" {
  description               = %[2]q
  network_load_balancer_arn = aws_lb.lb.arn

  tags = {
    %[3]q = %[4]q
  }
}
`, rName, description, tagKey1, tagValue1))
}

func testAccVPCTrafficMirrorTargetConfig_tags2(rName, description, tagKey1, tagValue1, tagKey2, tagValue2 string) string {
	return acctest.ConfigCompose(testAccTrafficMirrorTargetConfigBase(rName), fmt.Sprintf(`
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

resource "aws_ec2_traffic_mirror_target" "test" {
  description               = %[2]q
  network_load_balancer_arn = aws_lb.lb.arn

  tags = {
    %[3]q = %[4]q
    %[5]q = %[6]q
  }
}
`, rName, description, tagKey1, tagValue1, tagKey2, tagValue2))
}

func testAccPreCheckTrafficMirrorTarget(t *testing.T) {
	conn := acctest.Provider.Meta().(*conns.AWSClient).EC2Conn

	_, err := conn.DescribeTrafficMirrorTargets(&ec2.DescribeTrafficMirrorTargetsInput{})

	if acctest.PreCheckSkipError(err) {
		t.Skip("skipping traffic mirror target acceptance test: ", err)
	}

	if err != nil {
		t.Fatal("Unexpected PreCheck error: ", err)
	}
}

func testAccCheckTrafficMirrorTargetDestroy(s *terraform.State) error {
	conn := acctest.Provider.Meta().(*conns.AWSClient).EC2Conn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_ec2_traffic_mirror_target" {
			continue
		}

		out, err := conn.DescribeTrafficMirrorTargets(&ec2.DescribeTrafficMirrorTargetsInput{
			TrafficMirrorTargetIds: []*string{
				aws.String(rs.Primary.ID),
			},
		})

		if tfawserr.ErrCodeEquals(err, "InvalidTrafficMirrorTargetId.NotFound") {
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
