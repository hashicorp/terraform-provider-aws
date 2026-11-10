// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package ec2_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	tfec2 "github.com/hashicorp/terraform-provider-aws/internal/service/ec2"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccEC2AMIWatermark_basic(t *testing.T) {
	ctx := acctest.Context(t)
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_ami_watermark.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckAMIWatermarkDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccAMIWatermarkConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAMIWatermarkExists(ctx, t, resourceName),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("watermark_key"), knownvalue.NotNull()),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("image_id"), knownvalue.NotNull()),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrID), knownvalue.NotNull()),
				},
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateIdFunc: testAccAMIWatermarkImportStateIDFunc(resourceName),
				ImportStateVerify: true,
				// watermark_name is not returned by DescribeImages so cannot be verified on import
				ImportStateVerifyIgnore: []string{"watermark_name"},
			},
		},
	})
}

func TestAccEC2AMIWatermark_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_ami_watermark.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckAMIWatermarkDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccAMIWatermarkConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAMIWatermarkExists(ctx, t, resourceName),
					acctest.CheckFrameworkResourceDisappears(ctx, t, tfec2.ResourceAMIWatermark, resourceName),
				),
				ExpectNonEmptyPlan: true,
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PostApplyPostRefresh: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
					},
				},
			},
		},
	})
}

func testAccCheckAMIWatermarkExists(ctx context.Context, t *testing.T, n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.ProviderMeta(ctx, t).EC2Client(ctx)

		_, err := tfec2.FindImageWatermark(ctx, conn, rs.Primary.Attributes["image_id"], rs.Primary.Attributes["watermark_key"])

		return err
	}
}

func testAccCheckAMIWatermarkDestroy(ctx context.Context, t *testing.T) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.ProviderMeta(ctx, t).EC2Client(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_ami_watermark" {
				continue
			}

			_, err := tfec2.FindImageWatermark(ctx, conn, rs.Primary.Attributes["image_id"], rs.Primary.Attributes["watermark_key"])

			if retry.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("AMI Watermark %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccAMIWatermarkImportStateIDFunc(n string) resource.ImportStateIdFunc {
	return func(s *terraform.State) (string, error) {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return "", fmt.Errorf("Not Found: %s", n)
		}

		return fmt.Sprintf("%s,%s", rs.Primary.Attributes["image_id"], rs.Primary.Attributes["watermark_key"]), nil
	}
}

func testAccAMIWatermarkConfig_base(rName string) string {
	return acctest.ConfigCompose(acctest.ConfigLatestAmazonLinux2HVMEBSX8664AMI(), fmt.Sprintf(`
resource "aws_ami_copy" "test" {
  description       = %[1]q
  name              = %[1]q
  source_ami_id     = data.aws_ami.amzn2-ami-minimal-hvm-ebs-x86_64.id
  source_ami_region = data.aws_ami.amzn2-ami-minimal-hvm-ebs-x86_64.region
}
`, rName))
}

func testAccAMIWatermarkConfig_basic(rName string) string {
	return acctest.ConfigCompose(testAccAMIWatermarkConfig_base(rName), fmt.Sprintf(`
resource "aws_ami_watermark" "test" {
  image_id       = aws_ami_copy.test.id
  watermark_name = %[1]q
}
`, rName))
}
