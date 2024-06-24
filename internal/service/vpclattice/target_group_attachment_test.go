// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package vpclattice_test

import (
	"context"
	"fmt"
	"strconv"
	"testing"

	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfvpclattice "github.com/hashicorp/terraform-provider-aws/internal/service/vpclattice"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccVPCLatticeTargetGroupAttachment_instance(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_vpclattice_target_group_attachment.test"
	instanceResourceName := "aws_instance.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.VPCLatticeEndpointID)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.VPCLatticeServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckRegisterTargetsDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccTargetGroupAttachmentConfig_instance(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckTargetsExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "target.#", acctest.Ct1),
					resource.TestCheckResourceAttrPair(resourceName, "target.0.id", instanceResourceName, names.AttrID),
					resource.TestCheckResourceAttr(resourceName, "target.0.port", "80"),
				),
			},
		},
	})
}

func TestAccVPCLatticeTargetGroupAttachment_ip(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_vpclattice_target_group_attachment.test"
	instanceResourceName := "aws_instance.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.VPCLatticeEndpointID)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.VPCLatticeServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckRegisterTargetsDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccTargetGroupAttachmentConfig_ip(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckTargetsExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "target.#", acctest.Ct1),
					resource.TestCheckResourceAttrPair(resourceName, "target.0.id", instanceResourceName, "private_ip"),
					resource.TestCheckResourceAttr(resourceName, "target.0.port", "8080"),
				),
			},
		},
	})
}

func TestAccVPCLatticeTargetGroupAttachment_lambda(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_vpclattice_target_group_attachment.test"
	lambdaResourceName := "aws_lambda_function.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.VPCLatticeEndpointID)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.VPCLatticeServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckRegisterTargetsDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccTargetGroupAttachmentConfig_lambda(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckTargetsExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "target.#", acctest.Ct1),
					resource.TestCheckResourceAttrPair(resourceName, "target.0.id", lambdaResourceName, names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, "target.0.port", acctest.Ct0),
				),
			},
		},
	})
}

func TestAccVPCLatticeTargetGroupAttachment_alb(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_vpclattice_target_group_attachment.test"
	albResourceName := "aws_lb.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.VPCLatticeEndpointID)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.VPCLatticeServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckRegisterTargetsDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccTargetGroupAttachmentConfig_alb(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckTargetsExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "target.#", acctest.Ct1),
					resource.TestCheckResourceAttrPair(resourceName, "target.0.id", albResourceName, names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, "target.0.port", "80"),
				),
			},
		},
	})
}

func TestAccVPCLatticeTargetGroupAttachment_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_vpclattice_target_group_attachment.test"
	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.VPCLatticeEndpointID)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.VPCLatticeServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckRegisterTargetsDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccTargetGroupAttachmentConfig_instance(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTargetsExists(ctx, resourceName),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfvpclattice.ResourceTargetGroupAttachment(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccTargetGroupAttachmentConfig_baseInstance(rName string) string {
	return acctest.ConfigCompose(acctest.ConfigLatestAmazonLinux2HVMEBSX8664AMI(), acctest.ConfigVPCWithSubnets(rName, 1), fmt.Sprintf(`
resource "aws_instance" "test" {
  ami           = data.aws_ami.amzn2-ami-minimal-hvm-ebs-x86_64.id
  instance_type = "t2.small"
  subnet_id     = aws_subnet.test[0].id

  tags = {
    Name = %[1]q
  }
}
`, rName))
}

func testAccTargetGroupAttachmentConfig_instance(rName string) string {
	return acctest.ConfigCompose(testAccTargetGroupAttachmentConfig_baseInstance(rName), fmt.Sprintf(`
resource "aws_vpclattice_target_group" "test" {
  name = %[1]q
  type = "INSTANCE"

  config {
    port           = 80
    protocol       = "HTTP"
    vpc_identifier = aws_vpc.test.id
  }
}

resource "aws_vpclattice_target_group_attachment" "test" {
  target_group_identifier = aws_vpclattice_target_group.test.id

  target {
    id = aws_instance.test.id
  }
}
`, rName))
}

func testAccTargetGroupAttachmentConfig_ip(rName string) string {
	return acctest.ConfigCompose(testAccTargetGroupAttachmentConfig_baseInstance(rName), fmt.Sprintf(`
resource "aws_vpclattice_target_group" "test" {
  name = %[1]q
  type = "IP"

  config {
    port           = 80
    protocol       = "HTTP"
    vpc_identifier = aws_vpc.test.id
  }
}

resource "aws_vpclattice_target_group_attachment" "test" {
  target_group_identifier = aws_vpclattice_target_group.test.id

  target {
    id   = aws_instance.test.private_ip
    port = 8080
  }
}
`, rName))
}

func testAccTargetGroupAttachmentConfig_lambda(rName string) string {
	return fmt.Sprintf(`
data "aws_partition" "current" {}

resource "aws_vpclattice_target_group" "test" {
  name = %[1]q
  type = "LAMBDA"
}

resource "aws_lambda_function" "test" {
  filename      = "test-fixtures/lambda.zip"
  function_name = %[1]q
  role          = aws_iam_role.test.arn
  handler       = "test.handler"
  runtime       = "python3.7"
}

resource "aws_iam_role" "test" {
  name = %[1]q

  assume_role_policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [{
    "Action": "sts:AssumeRole",
    "Principal": {
      "Service": "lambda.${data.aws_partition.current.dns_suffix}"
    },
    "Effect": "Allow",
    "Sid": ""
  }]
}
EOF
}

resource "aws_vpclattice_target_group_attachment" "test" {
  target_group_identifier = aws_vpclattice_target_group.test.id

  target {
    id = aws_lambda_function.test.arn
  }
}
`, rName)
}

func testAccTargetGroupAttachmentConfig_alb(rName string) string {
	return acctest.ConfigCompose(acctest.ConfigVPCWithSubnets(rName, 2), fmt.Sprintf(`
resource "aws_vpclattice_target_group" "test" {
  name = %[1]q
  type = "ALB"

  config {
    port           = 80
    protocol       = "HTTP"
    vpc_identifier = aws_vpc.test.id
  }
}

resource "aws_lb" "test" {
  name               = %[1]q
  internal           = true
  load_balancer_type = "application"
  subnets            = aws_subnet.test[*].id

  enable_deletion_protection = false
}

resource "aws_lb_listener" "test" {
  load_balancer_arn = aws_lb.test.arn
  port              = "80"
  protocol          = "HTTP"

  default_action {
    type = "fixed-response"

    fixed_response {
      content_type = "text/plain"
      message_body = "Fixed response content"
      status_code  = "200"
    }
  }
}

resource "aws_vpclattice_target_group_attachment" "test" {
  target_group_identifier = aws_vpclattice_target_group.test.id

  target {
    id   = aws_lb.test.arn
    port = 80
  }
}
`, rName))
}

func testAccCheckTargetsExists(ctx context.Context, n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No VPC Lattice Target Group Attachment ID is set")
		}

		var err error
		var port int
		if v, ok := rs.Primary.Attributes["target.0.port"]; ok {
			port, err = strconv.Atoi(v)

			if err != nil {
				return err
			}
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).VPCLatticeClient(ctx)

		_, err = tfvpclattice.FindTargetByThreePartKey(ctx, conn, rs.Primary.Attributes["target_group_identifier"], rs.Primary.Attributes["target.0.id"], port)

		return err
	}
}

func testAccCheckRegisterTargetsDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).VPCLatticeClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_vpclattice_register_targets" {
				continue
			}

			var err error
			var port int
			if v, ok := rs.Primary.Attributes["target.0.port"]; ok {
				port, err = strconv.Atoi(v)

				if err != nil {
					return err
				}
			}

			_, err = tfvpclattice.FindTargetByThreePartKey(ctx, conn, rs.Primary.Attributes["target_group_identifier"], rs.Primary.Attributes["target.0.id"], port)

			if tfresource.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("VPC Lattice Target Group Attachment %s still exists", rs.Primary.ID)
		}

		return nil
	}
}
