// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package ec2_test

import (
	"context"
	"fmt"
	"regexp"
	"testing"

	"github.com/aws/aws-sdk-go/service/ec2"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfec2 "github.com/hashicorp/terraform-provider-aws/internal/service/ec2"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

func TestAccVPCTrafficMirrorTarget_nlb(t *testing.T) {
	ctx := acctest.Context(t)
	var v ec2.TrafficMirrorTarget
	resourceName := "aws_ec2_traffic_mirror_target.test"
	description := "test nlb target"
	rName := fmt.Sprintf("tf-acc-test-%s", sdkacctest.RandString(10))

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheckTrafficMirrorTarget(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, ec2.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckTrafficMirrorTargetDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccVPCTrafficMirrorTargetConfig_nlb(rName, description),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTrafficMirrorTargetExists(ctx, resourceName, &v),
					acctest.CheckResourceAttrAccountID(resourceName, "owner_id"),
					acctest.MatchResourceAttrRegionalARN(resourceName, "arn", "ec2", regexp.MustCompile(`traffic-mirror-target/tmt-.+`)),
					resource.TestCheckResourceAttr(resourceName, "description", description),
					resource.TestCheckResourceAttrPair(resourceName, "network_load_balancer_arn", "aws_lb.test", "arn"),
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
	ctx := acctest.Context(t)
	var v ec2.TrafficMirrorTarget
	resourceName := "aws_ec2_traffic_mirror_target.test"
	rName := fmt.Sprintf("tf-acc-test-%s", sdkacctest.RandString(10))
	description := "test eni target"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheckTrafficMirrorTarget(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, ec2.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckTrafficMirrorTargetDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccVPCTrafficMirrorTargetConfig_eni(rName, description),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTrafficMirrorTargetExists(ctx, resourceName, &v),
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

func TestAccVPCTrafficMirrorTarget_tags(t *testing.T) {
	ctx := acctest.Context(t)
	var v ec2.TrafficMirrorTarget
	resourceName := "aws_ec2_traffic_mirror_target.test"
	description := "test nlb target"
	rName := fmt.Sprintf("tf-acc-test-%s", sdkacctest.RandString(10))

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheckTrafficMirrorTarget(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, ec2.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckTrafficMirrorTargetDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccVPCTrafficMirrorTargetConfig_tags1(rName, description, "key1", "value1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTrafficMirrorTargetExists(ctx, resourceName, &v),
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
					testAccCheckTrafficMirrorTargetExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1updated"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
			{
				Config: testAccVPCTrafficMirrorTargetConfig_tags1(rName, description, "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTrafficMirrorTargetExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
		},
	})
}

func TestAccVPCTrafficMirrorTarget_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	var v ec2.TrafficMirrorTarget
	resourceName := "aws_ec2_traffic_mirror_target.test"
	description := "test nlb target"
	rName := fmt.Sprintf("tf-acc-test-%s", sdkacctest.RandString(10))

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheckTrafficMirrorTarget(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, ec2.EndpointsID),
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
		ErrorCheck:               acctest.ErrorCheck(t, ec2.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckTrafficMirrorTargetDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccVPCTrafficMirrorTargetConfig_gwlb(rName, description),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "description", description),
					resource.TestCheckResourceAttrPair(resourceName, "gateway_load_balancer_endpoint_id", "aws_vpc_endpoint.test", "id"),
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

func testAccCheckTrafficMirrorTargetDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).EC2Conn(ctx)

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

func testAccCheckTrafficMirrorTargetExists(ctx context.Context, n string, v *ec2.TrafficMirrorTarget) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No EC2 Traffic Mirror Target ID is set")
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).EC2Conn(ctx)

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
		acctest.ConfigLatestAmazonLinuxHVMEBSAMI(),
		fmt.Sprintf(`
resource "aws_instance" "test" {
  ami           = data.aws_ami.amzn-ami-minimal-hvm-ebs.id
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

func testAccPreCheckTrafficMirrorTarget(ctx context.Context, t *testing.T) {
	conn := acctest.Provider.Meta().(*conns.AWSClient).EC2Conn(ctx)

	_, err := conn.DescribeTrafficMirrorTargetsWithContext(ctx, &ec2.DescribeTrafficMirrorTargetsInput{})

	if acctest.PreCheckSkipError(err) {
		t.Skip("skipping traffic mirror target acceptance test: ", err)
	}

	if err != nil {
		t.Fatal("Unexpected PreCheck error: ", err)
	}
}
