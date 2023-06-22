package elbv2_test

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/elbv2"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	tfelbv2 "github.com/hashicorp/terraform-provider-aws/internal/service/elbv2"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccELBV2TargetRegistration_basic(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_lb_target_registration.test"
	targetGroupResourceName := "aws_lb_target_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, elbv2.EndpointsID)
		},
		ErrorCheck:               acctest.ErrorCheck(t, elbv2.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckTargetRegistrationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccTargetRegistrationConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTargetRegistrationExists(ctx, resourceName),
					resource.TestCheckResourceAttrPair(resourceName, "target_group_arn", targetGroupResourceName, "arn"),
					resource.TestCheckResourceAttr(resourceName, "target.#", "1"),
				),
			},
		},
	})
}

func TestAccELBV2TargetRegistration_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_lb_target_registration.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, elbv2.EndpointsID)
		},
		ErrorCheck:               acctest.ErrorCheck(t, elbv2.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckTargetRegistrationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccTargetRegistrationConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTargetRegistrationExists(ctx, resourceName),
					testAccCheckTargetRegistrationDisappears(ctx, resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccELBV2TargetRegistration_update(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_lb_target_registration.test"
	targetGroupResourceName := "aws_lb_target_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, elbv2.EndpointsID)
		},
		ErrorCheck:               acctest.ErrorCheck(t, elbv2.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckTargetRegistrationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccTargetRegistrationConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTargetRegistrationExists(ctx, resourceName),
					resource.TestCheckResourceAttrPair(resourceName, "target_group_arn", targetGroupResourceName, "arn"),
					resource.TestCheckResourceAttr(resourceName, "target.#", "1"),
				),
			},
			{
				Config: testAccTargetRegistrationConfig_update(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTargetRegistrationExists(ctx, resourceName),
					resource.TestCheckResourceAttrPair(resourceName, "target_group_arn", targetGroupResourceName, "arn"),
					resource.TestCheckResourceAttr(resourceName, "target.#", "3"),
				),
			},
			{
				Config: testAccTargetRegistrationConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTargetRegistrationExists(ctx, resourceName),
					resource.TestCheckResourceAttrPair(resourceName, "target_group_arn", targetGroupResourceName, "arn"),
					resource.TestCheckResourceAttr(resourceName, "target.#", "1"),
				),
			},
		},
	})
}

// testAccCheckTargetRegistrationDisappears is a custom variant of the shared acctest
// disappears helper. The shared function does not copy nested arguments into the resource
// to be deleted, resulting in an "InvalidParameter" exception as the nested Targets argument
// required for de-registration is empty.
func testAccCheckTargetRegistrationDisappears(ctx context.Context, n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("resource not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("resource ID missing: %s", n)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).ELBV2Conn(ctx)
		in := &elbv2.DeregisterTargetsInput{
			TargetGroupArn: aws.String(rs.Primary.Attributes["target_group_arn"]),
			Targets: []*elbv2.TargetDescription{
				{
					Id: aws.String(rs.Primary.Attributes["target.0.target_id"]),
				},
			},
		}

		_, err := conn.DeregisterTargetsWithContext(ctx, in)
		if err != nil && !tfawserr.ErrCodeEquals(err, elbv2.ErrCodeTargetGroupNotFoundException) {
			return fmt.Errorf("deregistering targets: %s", err)
		}

		return err
	}
}

func testAccCheckTargetRegistrationDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).ELBV2Conn(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_lb_target_registration" && rs.Type != "aws_alb_target_registration" {
				continue
			}

			targetGroupArn := rs.Primary.Attributes["target_group_arn"]

			// Extracting target data from nested object string attributes is complicated, so
			// lazily describe with only the target group ARN input and check the resulting
			// output count instead.
			out, err := conn.DescribeTargetHealthWithContext(ctx, &elbv2.DescribeTargetHealthInput{
				TargetGroupArn: aws.String(targetGroupArn),
			})
			if err == nil {
				if len(out.TargetHealthDescriptions) != 0 {
					return fmt.Errorf("Target Group %q still has registered targets", rs.Primary.ID)
				}
			}

			if tfawserr.ErrCodeEquals(err, elbv2.ErrCodeTargetGroupNotFoundException) || tfawserr.ErrCodeEquals(err, elbv2.ErrCodeInvalidTargetException) {
				return nil
			}

			return create.Error(names.ELBV2, create.ErrActionCheckingDestroyed, tfelbv2.ResNameTargetRegistration, rs.Primary.ID, errors.New("not destroyed"))
		}

		return nil
	}
}

func testAccCheckTargetRegistrationExists(ctx context.Context, name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return create.Error(names.ELBV2, create.ErrActionCheckingExistence, tfelbv2.ResNameTargetRegistration, name, errors.New("not found"))
		}

		if rs.Primary.ID == "" {
			return create.Error(names.ELBV2, create.ErrActionCheckingExistence, tfelbv2.ResNameTargetRegistration, name, errors.New("not set"))
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).ELBV2Conn(ctx)
		targetGroupArn := rs.Primary.Attributes["target_group_arn"]

		// Extracting target data from nested object string attributes is complicated, so
		// lazily describe with only the target group ARN input and check the resulting
		// output count instead.
		out, err := conn.DescribeTargetHealthWithContext(ctx, &elbv2.DescribeTargetHealthInput{
			TargetGroupArn: aws.String(targetGroupArn),
		})
		if err != nil {
			return create.Error(names.ELBV2, create.ErrActionCheckingExistence, tfelbv2.ResNameTargetRegistration, rs.Primary.ID, err)
		}
		if out.TargetHealthDescriptions == nil {
			return create.Error(names.ELBV2, create.ErrActionCheckingExistence, tfelbv2.ResNameTargetRegistration, rs.Primary.ID, errors.New("empty response"))
		}
		// TODO: check length of registered targets against count of configured target blocks?

		return nil
	}
}

func testAccTargetRegistrationConfig_basic(rName string) string {
	return acctest.ConfigCompose(
		acctest.ConfigLatestAmazonLinuxHVMEBSAMI(),
		acctest.AvailableEC2InstanceTypeForRegion("t3.micro", "t2.micro", "t1.micro", "m1.small"),
		acctest.ConfigVPCWithSubnets(rName, 1),
		fmt.Sprintf(`
resource "aws_instance" "test" {
  ami           = data.aws_ami.amzn-ami-minimal-hvm-ebs.id
  instance_type = data.aws_ec2_instance_type_offering.available.instance_type
  subnet_id     = aws_subnet.test[0].id
}

resource "aws_lb_target_group" "test" {
  name     = %[1]q
  port     = 80
  protocol = "HTTP"
  vpc_id   = aws_vpc.test.id
}

resource "aws_lb_target_registration" "test" {
  target_group_arn = aws_lb_target_group.test.arn

  target {
	target_id = aws_instance.test.id
  }
}
`, rName))
}

func testAccTargetRegistrationConfig_update(rName string) string {
	return acctest.ConfigCompose(
		acctest.ConfigLatestAmazonLinuxHVMEBSAMI(),
		acctest.AvailableEC2InstanceTypeForRegion("t3.micro", "t2.micro", "t1.micro", "m1.small"),
		acctest.ConfigVPCWithSubnets(rName, 1),
		fmt.Sprintf(`
resource "aws_instance" "test" {
  ami           = data.aws_ami.amzn-ami-minimal-hvm-ebs.id
  instance_type = data.aws_ec2_instance_type_offering.available.instance_type
  subnet_id     = aws_subnet.test[0].id
}

resource "aws_instance" "test2" {
  ami           = data.aws_ami.amzn-ami-minimal-hvm-ebs.id
  instance_type = data.aws_ec2_instance_type_offering.available.instance_type
  subnet_id     = aws_subnet.test[0].id
}

resource "aws_lb_target_group" "test" {
  name     = %[1]q
  port     = 80
  protocol = "HTTP"
  vpc_id   = aws_vpc.test.id
}

resource "aws_lb_target_registration" "test" {
  target_group_arn = aws_lb_target_group.test.arn

  target {
	target_id = aws_instance.test.id
  }
  target {
	target_id = aws_instance.test2.id
	port      = 80
  }
  target {
	target_id = aws_instance.test2.id
	port      = 8080
  }
}
`, rName))
}
