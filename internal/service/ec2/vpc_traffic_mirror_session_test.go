// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package ec2_test

import (
	"context"
	"fmt"
	"strconv"
	"testing"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	awstypes "github.com/aws/aws-sdk-go-v2/service/ec2/types"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfec2 "github.com/hashicorp/terraform-provider-aws/internal/service/ec2"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccVPCTrafficMirrorSession_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var v awstypes.TrafficMirrorSession
	resourceName := "aws_ec2_traffic_mirror_session.test"
	description := "test session"
	session := sdkacctest.RandIntRange(1, 32766)
	rName := fmt.Sprintf("tf-acc-test-%s", sdkacctest.RandString(10))
	pLen := sdkacctest.RandIntRange(1, 255)
	vni := sdkacctest.RandIntRange(1, 16777216)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheckTrafficMirrorSession(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckTrafficMirrorSessionDestroy(ctx),
		Steps: []resource.TestStep{
			//create
			{
				Config: testAccVPCTrafficMirrorSessionConfig_basic(rName, session),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTrafficMirrorSessionExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, ""),
					resource.TestCheckResourceAttrSet(resourceName, "packet_length"),
					resource.TestCheckResourceAttr(resourceName, "session_number", strconv.Itoa(session)),
					resource.TestMatchResourceAttr(resourceName, "virtual_network_id", regexache.MustCompile(`\d+`)),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct0),
					acctest.CheckResourceAttrAccountID(resourceName, names.AttrOwnerID),
					acctest.MatchResourceAttrRegionalARN(resourceName, names.AttrARN, "ec2", regexache.MustCompile(`traffic-mirror-session/tms-.+`)),
				),
			},
			// update of description, packet length and VNI
			{
				Config: testAccVPCTrafficMirrorSessionConfig_optionals(description, rName, session, pLen, vni),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTrafficMirrorSessionExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, description),
					resource.TestCheckResourceAttr(resourceName, "packet_length", strconv.Itoa(pLen)),
					resource.TestCheckResourceAttr(resourceName, "session_number", strconv.Itoa(session)),
					resource.TestCheckResourceAttr(resourceName, "virtual_network_id", strconv.Itoa(vni)),
				),
			},
			// removal of description, packet length and VNI
			{
				Config: testAccVPCTrafficMirrorSessionConfig_basic(rName, session),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTrafficMirrorSessionExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, ""),
					resource.TestCheckResourceAttr(resourceName, "session_number", strconv.Itoa(session)),
					resource.TestMatchResourceAttr(resourceName, "virtual_network_id", regexache.MustCompile(`\d+`)),
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
	ctx := acctest.Context(t)
	var v awstypes.TrafficMirrorSession
	resourceName := "aws_ec2_traffic_mirror_session.test"
	session := sdkacctest.RandIntRange(1, 32766)
	rName := fmt.Sprintf("tf-acc-test-%s", sdkacctest.RandString(10))

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheckTrafficMirrorSession(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckTrafficMirrorSessionDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccVPCTrafficMirrorSessionConfig_tags1(rName, acctest.CtKey1, acctest.CtValue1, session),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTrafficMirrorSessionExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey1, acctest.CtValue1),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccVPCTrafficMirrorSessionConfig_tags2(rName, acctest.CtKey1, acctest.CtValue1Updated, acctest.CtKey2, acctest.CtValue2, session),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTrafficMirrorSessionExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct2),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey1, acctest.CtValue1Updated),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey2, acctest.CtValue2),
				),
			},
			{
				Config: testAccVPCTrafficMirrorSessionConfig_tags1(rName, acctest.CtKey2, acctest.CtValue2, session),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTrafficMirrorSessionExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey2, acctest.CtValue2),
				),
			},
		},
	})
}

func TestAccVPCTrafficMirrorSession_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	var v awstypes.TrafficMirrorSession
	resourceName := "aws_ec2_traffic_mirror_session.test"
	session := sdkacctest.RandIntRange(1, 32766)
	rName := fmt.Sprintf("tf-acc-test-%s", sdkacctest.RandString(10))

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheckTrafficMirrorSession(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckTrafficMirrorSessionDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccVPCTrafficMirrorSessionConfig_basic(rName, session),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTrafficMirrorSessionExists(ctx, resourceName, &v),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfec2.ResourceTrafficMirrorSession(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccVPCTrafficMirrorSession_updateTrafficMirrorTarget(t *testing.T) {
	ctx := acctest.Context(t)
	var v1, v2 awstypes.TrafficMirrorSession
	resourceName := "aws_ec2_traffic_mirror_session.test"
	session := sdkacctest.RandIntRange(1, 32766)
	rName := fmt.Sprintf("tf-acc-test-%s", sdkacctest.RandString(10))

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheckTrafficMirrorSession(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckTrafficMirrorSessionDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccTrafficMirrorSessionConfig_trafficMirrorTarget(rName, 0, session),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTrafficMirrorSessionExists(ctx, resourceName, &v1),
				),
			},
			{
				Config: testAccTrafficMirrorSessionConfig_trafficMirrorTarget(rName, 1, session),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTrafficMirrorSessionExists(ctx, resourceName, &v2),
					testAccCheckTrafficMirrorSessionNotRecreated(t, &v1, &v2),
				),
			},
		},
	})
}

func testAccPreCheckTrafficMirrorSession(ctx context.Context, t *testing.T) {
	conn := acctest.Provider.Meta().(*conns.AWSClient).EC2Client(ctx)

	_, err := conn.DescribeTrafficMirrorSessions(ctx, &ec2.DescribeTrafficMirrorSessionsInput{})

	if acctest.PreCheckSkipError(err) {
		t.Skip("skipping traffic mirror sessions acceptance test: ", err)
	}

	if err != nil {
		t.Fatal("Unexpected PreCheck error: ", err)
	}
}

func testAccCheckTrafficMirrorSessionNotRecreated(t *testing.T, before, after *awstypes.TrafficMirrorSession) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if before, after := aws.ToString(before.TrafficMirrorSessionId), aws.ToString(after.TrafficMirrorSessionId); before != after {
			t.Fatalf("Expected TrafficMirrorSessionIDs not to change, but both got before: %s and after: %s", before, after)
		}

		return nil
	}
}

func testAccCheckTrafficMirrorSessionDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).EC2Client(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_ec2_traffic_mirror_session" {
				continue
			}

			_, err := tfec2.FindTrafficMirrorSessionByID(ctx, conn, rs.Primary.ID)

			if tfresource.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("EC2 Traffic Mirror Session %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccCheckTrafficMirrorSessionExists(ctx context.Context, n string, v *awstypes.TrafficMirrorSession) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).EC2Client(ctx)

		output, err := tfec2.FindTrafficMirrorSessionByID(ctx, conn, rs.Primary.ID)

		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

func testAccTrafficMirrorSessionConfig_base(rName string) string {
	return acctest.ConfigCompose(acctest.ConfigVPCWithSubnets(rName, 2), acctest.ConfigLatestAmazonLinux2HVMEBSX8664AMI(), fmt.Sprintf(`
resource "aws_instance" "test" {
  ami           = data.aws_ami.amzn2-ami-minimal-hvm-ebs-x86_64.id
  instance_type = "m5.large" # m5.large required because only Nitro instances support mirroring
  subnet_id     = aws_subnet.test[0].id

  tags = {
    Name = %[1]q
  }
}

resource "aws_lb" "test" {
  name               = %[1]q
  internal           = true
  load_balancer_type = "network"
  subnets            = aws_subnet.test[*].id

  enable_deletion_protection = false
}

resource "aws_ec2_traffic_mirror_filter" "test" {
  tags = {
    Name = %[1]q
  }
}

resource "aws_ec2_traffic_mirror_target" "test" {
  network_load_balancer_arn = aws_lb.test.arn

  tags = {
    Name = %[1]q
  }
}
`, rName))
}

func testAccVPCTrafficMirrorSessionConfig_basic(rName string, session int) string {
	return acctest.ConfigCompose(testAccTrafficMirrorSessionConfig_base(rName), fmt.Sprintf(`
resource "aws_ec2_traffic_mirror_session" "test" {
  traffic_mirror_filter_id = aws_ec2_traffic_mirror_filter.test.id
  traffic_mirror_target_id = aws_ec2_traffic_mirror_target.test.id
  network_interface_id     = aws_instance.test.primary_network_interface_id
  session_number           = %[1]d
}
`, session))
}

func testAccVPCTrafficMirrorSessionConfig_tags1(rName, tagKey1, tagValue1 string, session int) string {
	return acctest.ConfigCompose(testAccTrafficMirrorSessionConfig_base(rName), fmt.Sprintf(`
resource "aws_ec2_traffic_mirror_session" "test" {
  traffic_mirror_filter_id = aws_ec2_traffic_mirror_filter.test.id
  traffic_mirror_target_id = aws_ec2_traffic_mirror_target.test.id
  network_interface_id     = aws_instance.test.primary_network_interface_id
  session_number           = %[3]d

  tags = {
    %[1]q = %[2]q
  }
}
`, tagKey1, tagValue1, session))
}

func testAccVPCTrafficMirrorSessionConfig_tags2(rName, tagKey1, tagValue1, tagKey2, tagValue2 string, session int) string {
	return acctest.ConfigCompose(testAccTrafficMirrorSessionConfig_base(rName), fmt.Sprintf(`
resource "aws_ec2_traffic_mirror_session" "test" {
  traffic_mirror_filter_id = aws_ec2_traffic_mirror_filter.test.id
  traffic_mirror_target_id = aws_ec2_traffic_mirror_target.test.id
  network_interface_id     = aws_instance.test.primary_network_interface_id
  session_number           = %[5]d

  tags = {
    %[1]q = %[2]q
    %[3]q = %[4]q
  }
}
`, tagKey1, tagValue1, tagKey2, tagValue2, session))
}

func testAccVPCTrafficMirrorSessionConfig_optionals(description string, rName string, session, pLen, vni int) string {
	return acctest.ConfigCompose(testAccTrafficMirrorSessionConfig_base(rName), fmt.Sprintf(`
resource "aws_ec2_traffic_mirror_session" "test" {
  description              = %[1]q
  traffic_mirror_filter_id = aws_ec2_traffic_mirror_filter.test.id
  traffic_mirror_target_id = aws_ec2_traffic_mirror_target.test.id
  network_interface_id     = aws_instance.test.primary_network_interface_id
  session_number           = %[2]d
  packet_length            = %[3]d
  virtual_network_id       = %[4]d
}
`, description, session, pLen, vni))
}

func testAccTrafficMirrorSessionConfig_trafficMirrorTarget(rName string, idx, session int) string {
	return acctest.ConfigCompose(acctest.ConfigVPCWithSubnets(rName, 1), acctest.ConfigLatestAmazonLinux2HVMEBSX8664AMI(), fmt.Sprintf(`
resource "aws_instance" "target" {
  count = 2

  ami           = data.aws_ami.amzn2-ami-minimal-hvm-ebs-x86_64.id
  instance_type = "t2.micro"
  subnet_id     = aws_subnet.test[0].id

  tags = {
    Name = %[1]q
  }
}

resource "aws_instance" "test" {
  ami           = data.aws_ami.amzn2-ami-minimal-hvm-ebs-x86_64.id
  instance_type = "m5.large" # m5.large required because only Nitro instances support mirroring
  subnet_id     = aws_subnet.test[0].id

  tags = {
    Name = %[1]q
  }
}

resource "aws_ec2_traffic_mirror_target" "test" {
  count = 2

  network_interface_id = aws_instance.target[count.index].primary_network_interface_id

  tags = {
    Name = %[1]q
  }
}

resource "aws_ec2_traffic_mirror_filter" "test" {
  tags = {
    Name = %[1]q
  }
}

resource "aws_ec2_traffic_mirror_session" "test" {
  description              = %[1]q
  traffic_mirror_filter_id = aws_ec2_traffic_mirror_filter.test.id
  traffic_mirror_target_id = aws_ec2_traffic_mirror_target.test[%[2]d].id
  network_interface_id     = aws_instance.test.primary_network_interface_id
  session_number           = %[3]d

  tags = {
    Name = %[1]q
  }
}
`, rName, idx, session))
}
