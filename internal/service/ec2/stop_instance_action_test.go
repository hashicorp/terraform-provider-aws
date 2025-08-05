// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package ec2_test

import (
	"context"
	"fmt"
	"testing"

	awstypes "github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfec2 "github.com/hashicorp/terraform-provider-aws/internal/service/ec2"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccEC2StopInstanceAction_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var v awstypes.Instance
	resourceName := "aws_instance.test"
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.EC2)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccStopInstanceActionConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInstanceExists(ctx, resourceName, &v),
					testAccCheckInstanceState(ctx, resourceName, awstypes.InstanceStateNameRunning),
				),
			},
			{
				Config: testAccStopInstanceActionConfig_withAction(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInstanceExists(ctx, resourceName, &v),
					testAccCheckInstanceState(ctx, resourceName, awstypes.InstanceStateNameStopped),
				),
			},
		},
	})
}

func TestAccEC2StopInstanceAction_force(t *testing.T) {
	ctx := acctest.Context(t)
	var v awstypes.Instance
	resourceName := "aws_instance.test"
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.EC2)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccStopInstanceActionConfig_force(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInstanceExists(ctx, resourceName, &v),
					testAccCheckInstanceState(ctx, resourceName, awstypes.InstanceStateNameStopped),
				),
			},
		},
	})
}

func TestAccEC2StopInstanceAction_customTimeout(t *testing.T) {
	ctx := acctest.Context(t)
	var v awstypes.Instance
	resourceName := "aws_instance.test"
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.EC2)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccStopInstanceActionConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInstanceExists(ctx, resourceName, &v),
					testAccCheckInstanceState(ctx, resourceName, awstypes.InstanceStateNameRunning),
				),
			},
			{
				Config: testAccStopInstanceActionConfig_withTimeout(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInstanceExists(ctx, resourceName, &v),
					testAccCheckInstanceState(ctx, resourceName, awstypes.InstanceStateNameStopped),
				),
			},
		},
	})
}

// testAccCheckInstanceState checks that an instance is in the expected state
func testAccCheckInstanceState(ctx context.Context, n string, expectedState awstypes.InstanceStateName) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No EC2 Instance ID is set")
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).EC2Client(ctx)

		instance, err := tfec2.FindInstanceByID(ctx, conn, rs.Primary.ID)
		if err != nil {
			return err
		}

		if instance.State.Name != expectedState {
			return fmt.Errorf("Expected instance state %s, got %s", expectedState, instance.State.Name)
		}

		return nil
	}
}

func testAccStopInstanceActionConfig_basic(rName string) string {
	return acctest.ConfigCompose(
		acctest.ConfigLatestAmazonLinux2HVMEBSX8664AMI(),
		acctest.ConfigAvailableAZsNoOptIn(),
		acctest.AvailableEC2InstanceTypeForAvailabilityZone("data.aws_availability_zones.available.names[0]", "t3.micro", "t2.micro"),
		fmt.Sprintf(`
resource "aws_instance" "test" {
  ami           = data.aws_ami.amzn2-ami-minimal-hvm-ebs-x86_64.id
  instance_type = data.aws_ec2_instance_type_offering.available.instance_type
  
  tags = {
    Name = %[1]q
  }
}
`, rName))
}

func testAccStopInstanceActionConfig_withAction(rName string) string {
	return acctest.ConfigCompose(
		testAccStopInstanceActionConfig_basic(rName),
		`
action "aws_ec2_stop_instance" "test" {
  config {
    instance_id = aws_instance.test.id
  }
}
`)
}

func testAccStopInstanceActionConfig_force(rName string) string {
	return acctest.ConfigCompose(
		acctest.ConfigLatestAmazonLinux2HVMEBSX8664AMI(),
		acctest.ConfigAvailableAZsNoOptIn(),
		acctest.AvailableEC2InstanceTypeForAvailabilityZone("data.aws_availability_zones.available.names[0]", "t3.micro", "t2.micro"),
		fmt.Sprintf(`
resource "aws_instance" "test" {
  ami           = data.aws_ami.amzn2-ami-minimal-hvm-ebs-x86_64.id
  instance_type = data.aws_ec2_instance_type_offering.available.instance_type
  
  tags = {
    Name = %[1]q
  }

  lifecycle {
    action_trigger {
      events = [after_create]
      actions = [action.aws_ec2_stop_instance.test]
    }
  }
}

action "aws_ec2_stop_instance" "test" {
  config {
    instance_id = aws_instance.test.id
    force       = true
  }
}
`, rName))
}

func testAccStopInstanceActionConfig_withTimeout(rName string) string {
	return acctest.ConfigCompose(
		testAccStopInstanceActionConfig_basic(rName),
		`
action "aws_ec2_stop_instance" "test" {
  config {
    instance_id = aws_instance.test.id
    timeout     = 300
  }
}
`)
}
