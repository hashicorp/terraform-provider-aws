// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package ec2_test

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfec2 "github.com/hashicorp/terraform-provider-aws/internal/service/ec2"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccEC2InstanceState_basic(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_ec2_instance_state.test"
	state := "stopped"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckInstanceDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccInstanceStateConfig_basic(state, acctest.CtFalse),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInstanceStateExists(ctx, resourceName),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrInstanceID),
					resource.TestCheckResourceAttr(resourceName, names.AttrState, state),
				),
			},
		},
	})
}

func TestAccEC2InstanceState_state(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_ec2_instance_state.test"
	stateStopped := "stopped"
	stateRunning := "running"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckInstanceDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccInstanceStateConfig_basic(stateStopped, acctest.CtFalse),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInstanceStateExists(ctx, resourceName),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrInstanceID),
					resource.TestCheckResourceAttr(resourceName, names.AttrState, stateStopped),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccInstanceStateConfig_basic(stateRunning, acctest.CtFalse),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInstanceStateExists(ctx, resourceName),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrInstanceID),
					resource.TestCheckResourceAttr(resourceName, names.AttrState, stateRunning),
				),
			},
		},
	})
}

func TestAccEC2InstanceState_disappears_Instance(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_ec2_instance_state.test"
	parentResourceName := "aws_instance.test"
	state := "stopped"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckInstanceDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccInstanceStateConfig_basic(state, acctest.CtFalse),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInstanceStateExists(ctx, resourceName),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfec2.ResourceInstance(), parentResourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckInstanceStateExists(ctx context.Context, n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return errors.New("No EC2InstanceState ID is set")
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).EC2Client(ctx)

		out, err := tfec2.FindInstanceStateByID(ctx, conn, rs.Primary.ID)

		if err != nil {
			return err
		}

		if out == nil {
			return fmt.Errorf("Instance State %q does not exist", rs.Primary.ID)
		}

		return nil
	}
}

func testAccInstanceStateConfig_basic(state string, force string) string {
	return acctest.ConfigCompose(
		acctest.ConfigLatestAmazonLinux2HVMEBSX8664AMI(),
		acctest.AvailableEC2InstanceTypeForRegion("t3.micro", "t2.micro", "t1.micro", "m1.small"),
		fmt.Sprintf(`
resource "aws_instance" "test" {
  ami           = data.aws_ami.amzn2-ami-minimal-hvm-ebs-x86_64.id
  instance_type = data.aws_ec2_instance_type_offering.available.instance_type
}

resource "aws_ec2_instance_state" "test" {
  instance_id = aws_instance.test.id
  state       = %[1]q
  force       = %[2]s
}
`, state, force))
}
