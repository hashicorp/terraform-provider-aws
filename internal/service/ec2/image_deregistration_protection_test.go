// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package ec2_test

import (
	"fmt"
	"regexp"
	"testing"

	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccEc2ImageDeregistrationProtection_basic(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_ec2_image_deregistration_protection.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckAMIDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccEc2ImageDeregistrationProtectionBasic(rName, "t2.medium"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "with_cooldown", "false"),
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownOutputValueAtPath(
						"deregistration_protection",
						tfjsonpath.New("deregistration_protection"),
						knownvalue.StringRegexp(regexp.MustCompile("enabled-without-cooldown")),
					),
				},
			},
		},
	})
}

func testAccEc2ImageDeregistrationProtectionBasic(rName, instanceType string) string {
	return acctest.ConfigCompose(
		acctest.ConfigLatestAmazonLinux2HVMEBSX8664AMI(),
		testAccInstanceVPCConfig(rName, false, 0),
		fmt.Sprintf(`
resource "aws_instance" "test" {
  ami       = data.aws_ami.amzn2-ami-minimal-hvm-ebs-x86_64.id
  subnet_id = aws_subnet.test.id

  instance_type = %[1]q

  tags = {
    Name = %[2]q
  }
}

resource "aws_ami_from_instance" "test" {
  name               = "terraform-example-ami"
  source_instance_id = aws_instance.test.id
}

resource "aws_ec2_image_deregistration_protection" "test" {
  ami_id = aws_ami_from_instance.test.id
}

output "deregistration_protection_output" {
  value = aws_ec2_image_deregistration_protection.test
}	
`, instanceType, rName))
}
