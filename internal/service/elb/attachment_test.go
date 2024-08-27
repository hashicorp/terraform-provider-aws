// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package elb_test

import (
	"context"
	"fmt"
	"testing"

	awstypes "github.com/aws/aws-sdk-go-v2/service/elasticloadbalancing/types"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfelb "github.com/hashicorp/terraform-provider-aws/internal/service/elb"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccELBAttachment_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var conf awstypes.LoadBalancerDescription
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	lbResourceName := "aws_elb.test"
	resourceName1 := "aws_elb_attachment.test1"
	resourceName2 := "aws_elb_attachment.test2"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ELBServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckAttachmentDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccAttachmentConfig_1(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAttachmentExists(ctx, resourceName1),
					testAccCheckLoadBalancerExists(ctx, lbResourceName, &conf),
					testAccAttachmentCheckInstanceCount(&conf, 1),
				),
			},
			{
				Config: testAccAttachmentConfig_2(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAttachmentExists(ctx, resourceName1),
					testAccCheckAttachmentExists(ctx, resourceName2),
					testAccCheckLoadBalancerExists(ctx, lbResourceName, &conf),
					testAccAttachmentCheckInstanceCount(&conf, 2),
				),
			},
		},
	})
}

func TestAccELBAttachment_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_elb_attachment.test1"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ELBServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckAttachmentDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccAttachmentConfig_1(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAttachmentExists(ctx, resourceName),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfelb.ResourceAttachment(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckAttachmentDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).ELBClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_elb_attachment" {
				continue
			}

			err := tfelb.FindLoadBalancerAttachmentByTwoPartKey(ctx, conn, rs.Primary.Attributes["elb"], rs.Primary.Attributes["instance"])

			if tfresource.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("ELB Classic Attachment %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccCheckAttachmentExists(ctx context.Context, n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).ELBClient(ctx)

		err := tfelb.FindLoadBalancerAttachmentByTwoPartKey(ctx, conn, rs.Primary.Attributes["elb"], rs.Primary.Attributes["instance"])

		return err
	}
}

func testAccAttachmentCheckInstanceCount(v *awstypes.LoadBalancerDescription, expected int) resource.TestCheckFunc {
	return func(*terraform.State) error {
		if actual := len(v.Instances); actual != expected {
			return fmt.Errorf("instance count does not match: expected %d, got %d", expected, actual)
		}
		return nil
	}
}

func testAccAttachmentConfig_base(rName string) string {
	return acctest.ConfigCompose(acctest.ConfigLatestAmazonLinux2HVMEBSX8664AMI(), acctest.ConfigAvailableAZsNoOptInDefaultExclude(), fmt.Sprintf(`
resource "aws_elb" "test" {
  availability_zones = data.aws_availability_zones.available.names

  name = %[1]q

  listener {
    instance_port     = 8000
    instance_protocol = "http"
    lb_port           = 80
    lb_protocol       = "http"
  }
}
`, rName))
}

func testAccAttachmentConfig_1(rName string) string {
	return acctest.ConfigCompose(testAccAttachmentConfig_base(rName), fmt.Sprintf(`
resource "aws_instance" "test1" {
  ami           = data.aws_ami.amzn2-ami-minimal-hvm-ebs-x86_64.id
  instance_type = "t2.micro"

  tags = {
    Name = %[1]q
  }
}

resource "aws_elb_attachment" "test1" {
  elb      = aws_elb.test.id
  instance = aws_instance.test1.id
}
`, rName))
}

func testAccAttachmentConfig_2(rName string) string {
	return acctest.ConfigCompose(testAccAttachmentConfig_1(rName), fmt.Sprintf(`
resource "aws_instance" "test2" {
  ami           = data.aws_ami.amzn2-ami-minimal-hvm-ebs-x86_64.id
  instance_type = "t2.micro"

  tags = {
    Name = %[1]q
  }
}

resource "aws_elb_attachment" "test2" {
  elb      = aws_elb.test.id
  instance = aws_instance.test2.id
}
`, rName))
}
