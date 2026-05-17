// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package elb_test

import (
	"context"
	"fmt"
	"testing"

	awstypes "github.com/aws/aws-sdk-go-v2/service/elasticloadbalancing/types"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	tfelb "github.com/hashicorp/terraform-provider-aws/internal/service/elb"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccELBAttachment_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var conf awstypes.LoadBalancerDescription
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	lbResourceName := "aws_elb.test"
	resourceName1 := "aws_elb_attachment.test1"
	resourceName2 := "aws_elb_attachment.test2"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ELBServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckAttachmentDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccAttachmentConfig_1(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAttachmentExists(ctx, t, resourceName1),
					testAccCheckLoadBalancerExists(ctx, t, lbResourceName, &conf),
					testAccAttachmentCheckInstanceCount(&conf, 1),
				),
			},
			{
				Config: testAccAttachmentConfig_2(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAttachmentExists(ctx, t, resourceName1),
					testAccCheckAttachmentExists(ctx, t, resourceName2),
					testAccCheckLoadBalancerExists(ctx, t, lbResourceName, &conf),
					testAccAttachmentCheckInstanceCount(&conf, 2),
				),
			},
		},
	})
}

func TestAccELBAttachment_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_elb_attachment.test1"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ELBServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckAttachmentDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccAttachmentConfig_1(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAttachmentExists(ctx, t, resourceName),
					acctest.CheckSDKResourceDisappears(ctx, t, tfelb.ResourceAttachment(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckAttachmentDestroy(ctx context.Context, t *testing.T) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.ProviderMeta(ctx, t).ELBClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_elb_attachment" {
				continue
			}

			err := tfelb.FindLoadBalancerAttachmentByTwoPartKey(ctx, conn, rs.Primary.Attributes["elb"], rs.Primary.Attributes["instance"])

			if retry.NotFound(err) {
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

func testAccCheckAttachmentExists(ctx context.Context, t *testing.T, n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.ProviderMeta(ctx, t).ELBClient(ctx)

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
