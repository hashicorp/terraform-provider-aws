package vpclattice_test

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/vpclattice"
	"github.com/aws/aws-sdk-go-v2/service/vpclattice/types"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	tfvpclattice "github.com/hashicorp/terraform-provider-aws/internal/service/vpclattice"

	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccVPCLatticeRegisterTargets_instance(t *testing.T) {
	ctx := acctest.Context(t)
	var targets vpclattice.ListTargetsOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_vpclattice_register_targets.test"
	instanceId := "aws_instance.test_instance"
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
				Config: testAccRegisterTargets_instance(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckTargetsExists(ctx, resourceName, &targets),
					resource.TestCheckResourceAttrPair(resourceName, "default_action.0.forward.0.target_groups.0.target_group_identifier", targetGroupResourceName, "target_group_identifier"),
					resource.TestCheckResourceAttrPair(resourceName, "targets.0.id", instanceId, "id"),
					resource.TestCheckResourceAttr(resourceName, "targets.0.port", "80"),
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

func TestAccVPCLatticeRegisterTargets_ip(t *testing.T) {
	ctx := acctest.Context(t)
	var targets vpclattice.ListTargetsOutput
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
					testAccCheckTargetsExists(ctx, resourceName, &targets),
					resource.TestCheckResourceAttrPair(resourceName, "default_action.0.forward.0.target_groups.0.target_group_identifier", targetGroupResourceName, "target_group_identifier"),
					resource.TestCheckResourceAttrPair(resourceName, "targets.0.id", instanceIP, "private_ip"),
					resource.TestCheckResourceAttr(resourceName, "targets.0.port", "80"),
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

func TestAccVPCLatticeRegisterTargets_lambda(t *testing.T) {
	ctx := acctest.Context(t)
	var targets vpclattice.ListTargetsOutput
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
					testAccCheckTargetsExists(ctx, resourceName, &targets),
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

func TestAccVPCLatticeRegisterTargets_alb(t *testing.T) {
	ctx := acctest.Context(t)
	var targets vpclattice.ListTargetsOutput
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
					testAccCheckTargetsExists(ctx, resourceName, &targets),
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
	  
	  resource "aws_instance" "test_instance" {
		ami           = data.aws_ami.amzn-ami-minimal-hvm-ebs.id
		instance_type = "t2.micro"
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
`, rName))
}

func testAccRegisterTargets_instance(rName string) string {
	return acctest.ConfigCompose(testAccRegisterTargetsConfig_instance(rName), `
	resource "aws_vpclattice_register_targets" "test" {
		depends_on = [aws_instance.test_instance]
		target_group_identifier = aws_vpclattice_target_group.test.id
	  
		targets {
		  id   = aws_instance.test_instance.id
		  port = 80
			}
	  }

`)
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
		instance_type = "t2.micro"
		subnet_id     = aws_subnet.test[0].id
	  }

resource "aws_vpclattice_target_group" "test" {
  name = %[1]q
  type = "IP"

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
		depends_on = [aws_instance.test_ip]
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
	name = %[1]q
	type = "LAMBDA"
  
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
		depends_on = [aws_lambda_function.test]
		target_group_identifier = aws_vpclattice_target_group.test.id
	  
		targets {
		  id   = aws_lambda_function.test.arn
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
	subnets            =  [aws_subnet.test[0].id, aws_subnet.test[1].id]
  
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
		depends_on = [aws_lb.test]
		target_group_identifier = aws_vpclattice_target_group.test.id
	  
		targets {
		  id   = aws_lb.test.arn
		  port = 80
			}
	  }

`)
}

func testAccCheckTargetsExists(ctx context.Context, name string, targets *vpclattice.ListTargetsOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return create.Error(names.VPCLattice, create.ErrActionCheckingExistence, tfvpclattice.ResNameRegisterTargets, name, errors.New("not found"))
		}

		if rs.Primary.ID == "" {
			return create.Error(names.VPCLattice, create.ErrActionCheckingExistence, tfvpclattice.ResNameRegisterTargets, name, errors.New("not set"))
		}

		targetGroupIdentifier := rs.Primary.Attributes["target_group_identifier"]
		targetID := rs.Primary.Attributes["targets.0.id"]
		portStr, hasPort := rs.Primary.Attributes["targets.0.port"]
		port, _ := strconv.Atoi(portStr)
		hasValidPort := hasPort && port > 0

		if targetID == "" {
			return fmt.Errorf("Error: target ID is empty")
		}

		target := types.Target{
			Id: aws.String(targetID),
		}

		if hasValidPort {
			portStr := rs.Primary.Attributes["targets.0.port"]
			port, err := strconv.Atoi(portStr)
			if err != nil {
				return fmt.Errorf("Error parsing target port: %s", err)
			}
			target.Port = aws.Int32(int32(port))
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).VPCLatticeClient()
		resp, err := conn.ListTargets(ctx, &vpclattice.ListTargetsInput{
			TargetGroupIdentifier: aws.String(targetGroupIdentifier),
			Targets:               []types.Target{target},
		})

		if err != nil {
			return create.Error(names.VPCLattice, create.ErrActionCheckingExistence, tfvpclattice.ResNameRegisterTargets, rs.Primary.ID, err)
		}

		*targets = *resp

		return nil
	}
}

func testAccCheckRegisterTargetsDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).VPCLatticeClient()

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_vpclattice_register_targets" {
				continue
			}

			targetGroupIdentifier := rs.Primary.Attributes["target_group_identifier"]
			targetID := rs.Primary.Attributes["targets.0.id"]
			portStr, hasPort := rs.Primary.Attributes["targets.0.port"]
			port, _ := strconv.Atoi(portStr)
			hasValidPort := hasPort && port > 0

			target := types.Target{
				Id: aws.String(targetID),
			}

			if hasValidPort {
				portStr := rs.Primary.Attributes["targets.0.port"]
				port, err := strconv.Atoi(portStr)
				if err != nil {
					return fmt.Errorf("Error parsing target port: %s", err)
				}
				target.Port = aws.Int32(int32(port))
			}

			_, err := conn.ListTargets(ctx, &vpclattice.ListTargetsInput{
				TargetGroupIdentifier: aws.String(targetGroupIdentifier),
				Targets:               []types.Target{target},
			})

			if err != nil {
				var nfe *types.ResourceNotFoundException
				if errors.As(err, &nfe) {
					return nil
				}
				return err
			}

			return create.Error(names.VPCLattice, create.ErrActionCheckingDestroyed, tfvpclattice.ResNameRegisterTargets, rs.Primary.ID, errors.New("not destroyed"))
		}

		return nil
	}
}
