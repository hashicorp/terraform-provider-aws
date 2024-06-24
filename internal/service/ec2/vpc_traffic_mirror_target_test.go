// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package ec2_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/YakDriver/regexache"
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

func TestAccVPCTrafficMirrorTarget_nlb(t *testing.T) {
	ctx := acctest.Context(t)
	var v awstypes.TrafficMirrorTarget
	resourceName := "aws_ec2_traffic_mirror_target.test"
	description := "test nlb target"
	rName := fmt.Sprintf("tf-acc-test-%s", sdkacctest.RandString(10))

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheckTrafficMirrorTarget(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckTrafficMirrorTargetDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccVPCTrafficMirrorTargetConfig_nlb(rName, description),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTrafficMirrorTargetExists(ctx, resourceName, &v),
					acctest.CheckResourceAttrAccountID(resourceName, names.AttrOwnerID),
					acctest.MatchResourceAttrRegionalARN(resourceName, names.AttrARN, "ec2", regexache.MustCompile(`traffic-mirror-target/tmt-.+`)),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, description),
					resource.TestCheckResourceAttrPair(resourceName, "network_load_balancer_arn", "aws_lb.test", names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct0),
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
	ctx := acctest.Context(t)
	var v awstypes.TrafficMirrorTarget
	resourceName := "aws_ec2_traffic_mirror_target.test"
	rName := fmt.Sprintf("tf-acc-test-%s", sdkacctest.RandString(10))
	description := "test eni target"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheckTrafficMirrorTarget(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckTrafficMirrorTargetDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccVPCTrafficMirrorTargetConfig_eni(rName, description),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTrafficMirrorTargetExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, description),
					resource.TestMatchResourceAttr(resourceName, names.AttrNetworkInterfaceID, regexache.MustCompile("eni-.*")),
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
	ctx := acctest.Context(t)
	var v awstypes.TrafficMirrorTarget
	resourceName := "aws_ec2_traffic_mirror_target.test"
	description := "test nlb target"
	rName := fmt.Sprintf("tf-acc-test-%s", sdkacctest.RandString(10))

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheckTrafficMirrorTarget(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckTrafficMirrorTargetDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccVPCTrafficMirrorTargetConfig_tags1(rName, description, acctest.CtKey1, acctest.CtValue1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTrafficMirrorTargetExists(ctx, resourceName, &v),
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
				Config: testAccVPCTrafficMirrorTargetConfig_tags2(rName, description, acctest.CtKey1, acctest.CtValue1Updated, acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTrafficMirrorTargetExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct2),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey1, acctest.CtValue1Updated),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey2, acctest.CtValue2),
				),
			},
			{
				Config: testAccVPCTrafficMirrorTargetConfig_tags1(rName, description, acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTrafficMirrorTargetExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey2, acctest.CtValue2),
				),
			},
		},
	})
}

func TestAccVPCTrafficMirrorTarget_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	var v awstypes.TrafficMirrorTarget
	resourceName := "aws_ec2_traffic_mirror_target.test"
	description := "test nlb target"
	rName := fmt.Sprintf("tf-acc-test-%s", sdkacctest.RandString(10))

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheckTrafficMirrorTarget(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckTrafficMirrorTargetDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccVPCTrafficMirrorTargetConfig_nlb(rName, description),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTrafficMirrorTargetExists(ctx, resourceName, &v),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfec2.ResourceTrafficMirrorTarget(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccVPCTrafficMirrorTarget_gwlb(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_ec2_traffic_mirror_target.test"
	rName := fmt.Sprintf("tf-acc-test-%s", sdkacctest.RandString(10))
	description := "test gwlb endpoint target"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheckTrafficMirrorTarget(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckTrafficMirrorTargetDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccVPCTrafficMirrorTargetConfig_gwlb(rName, description),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, description),
					resource.TestCheckResourceAttrPair(resourceName, "gateway_load_balancer_endpoint_id", "aws_vpc_endpoint.test", names.AttrID),
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

func testAccPreCheckTrafficMirrorTarget(ctx context.Context, t *testing.T) {
	conn := acctest.Provider.Meta().(*conns.AWSClient).EC2Client(ctx)

	_, err := conn.DescribeTrafficMirrorTargets(ctx, &ec2.DescribeTrafficMirrorTargetsInput{})

	if acctest.PreCheckSkipError(err) {
		t.Skip("skipping traffic mirror target acceptance test: ", err)
	}

	if err != nil {
		t.Fatal("Unexpected PreCheck error: ", err)
	}
}

func testAccCheckTrafficMirrorTargetDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).EC2Client(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_ec2_traffic_mirror_target" {
				continue
			}

			_, err := tfec2.FindTrafficMirrorTargetByID(ctx, conn, rs.Primary.ID)

			if tfresource.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("EC2 Traffic Mirror Target %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccCheckTrafficMirrorTargetExists(ctx context.Context, n string, v *awstypes.TrafficMirrorTarget) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).EC2Client(ctx)

		output, err := tfec2.FindTrafficMirrorTargetByID(ctx, conn, rs.Primary.ID)

		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

func testAccVPCTrafficMirrorTargetConfig_nlb(rName, description string) string {
	return acctest.ConfigCompose(acctest.ConfigVPCWithSubnets(rName, 2), fmt.Sprintf(`
resource "aws_lb" "test" {
  name               = %[1]q
  internal           = true
  load_balancer_type = "network"
  subnets            = aws_subnet.test[*].id

  enable_deletion_protection = false
}

resource "aws_ec2_traffic_mirror_target" "test" {
  description               = %[2]q
  network_load_balancer_arn = aws_lb.test.arn
}
`, rName, description))
}

func testAccVPCTrafficMirrorTargetConfig_eni(rName, description string) string {
	return acctest.ConfigCompose(
		acctest.ConfigVPCWithSubnets(rName, 1),
		acctest.ConfigLatestAmazonLinux2HVMEBSX8664AMI(),
		fmt.Sprintf(`
resource "aws_instance" "test" {
  ami           = data.aws_ami.amzn2-ami-minimal-hvm-ebs-x86_64.id
  instance_type = "t2.micro"
  subnet_id     = aws_subnet.test[0].id

  tags = {
    Name = %[1]q
  }
}

resource "aws_ec2_traffic_mirror_target" "test" {
  description          = %[2]q
  network_interface_id = aws_instance.test.primary_network_interface_id

  tags = {
    Name = %[1]q
  }
}
`, rName, description))
}

func testAccVPCTrafficMirrorTargetConfig_tags1(rName, description, tagKey1, tagValue1 string) string {
	return acctest.ConfigCompose(acctest.ConfigVPCWithSubnets(rName, 2), fmt.Sprintf(`
resource "aws_lb" "test" {
  name               = %[1]q
  internal           = true
  load_balancer_type = "network"
  subnets            = aws_subnet.test[*].id

  enable_deletion_protection = false
}

resource "aws_ec2_traffic_mirror_target" "test" {
  description               = %[2]q
  network_load_balancer_arn = aws_lb.test.arn

  tags = {
    %[3]q = %[4]q
  }
}
`, rName, description, tagKey1, tagValue1))
}

func testAccVPCTrafficMirrorTargetConfig_tags2(rName, description, tagKey1, tagValue1, tagKey2, tagValue2 string) string {
	return acctest.ConfigCompose(acctest.ConfigVPCWithSubnets(rName, 2), fmt.Sprintf(`
resource "aws_lb" "test" {
  name               = %[1]q
  internal           = true
  load_balancer_type = "network"
  subnets            = aws_subnet.test[*].id

  enable_deletion_protection = false
}

resource "aws_ec2_traffic_mirror_target" "test" {
  description               = %[2]q
  network_load_balancer_arn = aws_lb.test.arn

  tags = {
    %[3]q = %[4]q
    %[5]q = %[6]q
  }
}
`, rName, description, tagKey1, tagValue1, tagKey2, tagValue2))
}

func testAccVPCTrafficMirrorTargetConfig_gwlb(rName, description string) string {
	return acctest.ConfigCompose(
		testAccVPCEndpointConfig_gatewayLoadBalancer(rName),
		fmt.Sprintf(`
resource "aws_ec2_traffic_mirror_target" "test" {
  description                       = %[2]q
  gateway_load_balancer_endpoint_id = aws_vpc_endpoint.test.id

  tags = {
    Name = %[1]q
  }
}
`, rName, description))
}
