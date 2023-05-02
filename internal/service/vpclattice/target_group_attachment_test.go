package vpclattice_test

import (
	"context"
	"fmt"
	"strconv"
	"testing"

	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
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
	targetGroupResourceName := "aws_vpclattice_target_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.VPCLatticeEndpointID)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.VPCLatticeEndpointID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckRegisterTargetsDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccRegisterTargetsConfig_instance(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckTargetsExists(ctx, resourceName),
					resource.TestCheckResourceAttrPair(resourceName, "default_action.0.forward.0.target_groups.0.target_group_identifier", targetGroupResourceName, "target_group_identifier"),
					resource.TestCheckResourceAttrPair(resourceName, "target.0.id", instanceResourceName, "id"),
					resource.TestCheckResourceAttr(resourceName, "target.0.port", "80"),
				),
			},
		},
	})
}

func TestAccVPCLatticeTargetGroupAttachment_ip(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_vpclattice_register_targets.test"
	instanceIP := "aws_instance.test_ip"
	targetGroupResourceName := "aws_vpclattice_target_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.VPCLatticeEndpointID)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.VPCLatticeEndpointID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckRegisterTargetsDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccRegisterTargets_ip(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckTargetsExists(ctx, resourceName),
					resource.TestCheckResourceAttrPair(resourceName, "default_action.0.forward.0.target_groups.0.target_group_identifier", targetGroupResourceName, "target_group_identifier"),
					resource.TestCheckResourceAttrPair(resourceName, "targets.0.id", instanceIP, "private_ip"),
					resource.TestCheckResourceAttr(resourceName, "targets.0.port", "80"),
				),
			},
		},
	})
}

func TestAccVPCLatticeTargetGroupAttachment_lambda(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_vpclattice_register_targets.test"
	lambdaName := "aws_lambda_function.test"
	targetGroupResourceName := "aws_vpclattice_target_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.VPCLatticeEndpointID)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.VPCLatticeEndpointID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckRegisterTargetsDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccRegisterTargets_lambda(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckTargetsExists(ctx, resourceName),
					resource.TestCheckResourceAttrPair(resourceName, "default_action.0.forward.0.target_groups.0.target_group_identifier", targetGroupResourceName, "target_group_identifier"),
					resource.TestCheckResourceAttrPair(resourceName, "targets.0.id", lambdaName, "arn"),
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

func TestAccVPCLatticeTargetGroupAttachment_alb(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_vpclattice_register_targets.test"
	albName := "aws_lb.test"
	targetGroupResourceName := "aws_vpclattice_target_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.VPCLatticeEndpointID)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.VPCLatticeEndpointID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckRegisterTargetsDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccRegisterTargets_alb(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckTargetsExists(ctx, resourceName),
					resource.TestCheckResourceAttrPair(resourceName, "default_action.0.forward.0.target_groups.0.target_group_identifier", targetGroupResourceName, "target_group_identifier"),
					resource.TestCheckResourceAttrPair(resourceName, "targets.0.id", albName, "arn"),
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

func testAccRegisterTargetsConfig_instance(rName string) string {
	return acctest.ConfigCompose(acctest.ConfigLatestAmazonLinuxHVMEBSAMI(), acctest.ConfigVPCWithSubnets(rName, 1), fmt.Sprintf(`
resource "aws_instance" "test" {
  ami           = data.aws_ami.amzn-ami-minimal-hvm-ebs.id
  instance_type = "t2.small"
  subnet_id     = aws_subnet.test[0].id
}

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

func testAccRegisterTargetsConfig_ipAddress(rName string) string {
	return acctest.ConfigCompose(acctest.ConfigVPCWithSubnets(rName, 1), fmt.Sprintf(`


data "aws_ami" "amzn-ami-minimal-hvm-ebs" {
  most_recent = true
  owners      = ["amazon"]

  filter {
    name   = "name"
    values = ["amzn-ami-minimal-hvm-*"]
  }

  filter {
    name   = "root-device-type"
    values = ["ebs"]
  }
}

resource "aws_instance" "test_ip" {
  ami           = data.aws_ami.amzn-ami-minimal-hvm-ebs.id
  instance_type = "t3.large"
  subnet_id     = aws_subnet.test[0].id
}

resource "aws_vpclattice_target_group" "test" {
  depends_on = [aws_instance.test_ip]
  name       = %[1]q
  type       = "IP"

  config {
    port           = 80
    protocol       = "HTTP"
    vpc_identifier = aws_vpc.test.id
  }
}
`, rName))
}

func testAccRegisterTargets_ip(rName string) string {
	return acctest.ConfigCompose(testAccRegisterTargetsConfig_ipAddress(rName), `


resource "aws_vpclattice_register_targets" "test" {
  depends_on              = [aws_vpclattice_target_group.test]
  target_group_identifier = aws_vpclattice_target_group.test.id

  targets {
    id   = aws_instance.test_ip.private_ip
    port = 80
  }
}


`)
}

func testAccRegisterTargetsConfig_lambda(rName string) string {
	return fmt.Sprintf(`
data "aws_partition" "current" {}

resource "aws_vpclattice_target_group" "test" {
  depends_on = [aws_lambda_function.test]
  name       = %[1]q
  type       = "LAMBDA"

}

resource "aws_lambda_function" "test" {
  filename      = "test-fixtures/lambda.zip"
  function_name = %[1]q
  role          = aws_iam_role.test.arn
  handler       = "lambda_elb.lambda_handler"
  runtime       = "python3.7"
}



resource "aws_iam_role" "test" {
  assume_role_policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Action": "sts:AssumeRole",
      "Principal": {
        "Service": "lambda.${data.aws_partition.current.dns_suffix}"
      },
      "Effect": "Allow",
      "Sid": ""
    }
  ]
}
	EOF

}
`, rName)
}

func testAccRegisterTargets_lambda(rName string) string {
	return acctest.ConfigCompose(testAccRegisterTargetsConfig_lambda(rName), `
resource "aws_vpclattice_register_targets" "test" {
  depends_on              = [aws_vpclattice_target_group.test]
  target_group_identifier = aws_vpclattice_target_group.test.id

  targets {
    id = aws_lambda_function.test.arn
  }
}


`)
}

func testAccRegisterTargetsConfig_alb(rName string) string {
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
  subnets            = [aws_subnet.test[0].id, aws_subnet.test[1].id]

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


`, rName))
}

func testAccRegisterTargets_alb(rName string) string {
	return acctest.ConfigCompose(testAccRegisterTargetsConfig_alb(rName), `
resource "aws_vpclattice_register_targets" "test" {
  depends_on              = [aws_lb.test]
  target_group_identifier = aws_vpclattice_target_group.test.id

  targets {
    id   = aws_lb.test.arn
    port = 80
  }
}


`)
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

		conn := acctest.Provider.Meta().(*conns.AWSClient).VPCLatticeClient()

		_, err = tfvpclattice.FindTargetByThreePartKey(ctx, conn, rs.Primary.Attributes["target_group_identifier"], rs.Primary.Attributes["target.0.id"], port)

		return err
	}
}

func testAccCheckRegisterTargetsDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).VPCLatticeClient()

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
