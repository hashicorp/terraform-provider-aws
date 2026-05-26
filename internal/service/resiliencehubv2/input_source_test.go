// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package resiliencehubv2_test

import (
	"context"
	"fmt"
	"testing"

	awstypes "github.com/aws/aws-sdk-go-v2/service/resiliencehubv2/types"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	tfresiliencehubv2 "github.com/hashicorp/terraform-provider-aws/internal/service/resiliencehubv2"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccResilienceHubV2InputSource_cfnStack(t *testing.T) {
	ctx := acctest.Context(t)
	var is awstypes.InputSourceSummary
	rName := acctest.RandomWithPrefix(t, "tf-acc-test")
	resourceName := "aws_resiliencehubv2_input_source.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.ResilienceHubV2),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckInputSourceDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccInputSourceConfig_cfnStack(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInputSourceExists(ctx, t, resourceName, &is),
					resource.TestCheckResourceAttrSet(resourceName, "input_source_id"),
					resource.TestCheckResourceAttrSet(resourceName, "cfn_stack_arn"),
				),
			},
		},
	})
}

func testAccCheckInputSourceDestroy(ctx context.Context, t *testing.T) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.ProviderMeta(ctx, t).ResilienceHubV2Client(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_resiliencehubv2_input_source" {
				continue
			}

			_, err := tfresiliencehubv2.FindInputSourceByID(ctx, conn, rs.Primary.Attributes["service_arn"], rs.Primary.Attributes["input_source_id"])
			if err == nil {
				return fmt.Errorf("Resilience Hub V2 Input Source %s still exists", rs.Primary.Attributes[names.AttrID])
			}
		}

		return nil
	}
}

func testAccCheckInputSourceExists(ctx context.Context, t *testing.T, n string, v *awstypes.InputSourceSummary) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Input Source not found: %s", n)
		}

		conn := acctest.ProviderMeta(ctx, t).ResilienceHubV2Client(ctx)

		output, err := tfresiliencehubv2.FindInputSourceByID(ctx, conn, rs.Primary.Attributes["service_arn"], rs.Primary.Attributes["input_source_id"])
		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

func testAccInputSourceConfig_cfnStack(rName string) string {
	return fmt.Sprintf(`
resource "aws_cloudformation_stack" "test" {
  name = %[1]q

  template_body = jsonencode({
    AWSTemplateFormatVersion = "2010-09-09"
    Description              = "Test stack for NGRH input source"
    Resources = {
      WaitHandle = {
        Type = "AWS::CloudFormation::WaitConditionHandle"
      }
    }
  })
}

resource "aws_resiliencehubv2_policy" "test" {
  name = "%[1]s-policy"

  availability_slo {
    target = 99.9
  }
}

resource "aws_resiliencehubv2_service" "test" {
  name    = "%[1]s-service"
  regions = ["us-west-2"]

  policy_arn = aws_resiliencehubv2_policy.test.arn

  permission_model {
    invoker_role_name = "AWSResilienceHubAssessmentRole"
  }
}

resource "aws_resiliencehubv2_input_source" "test" {
  service_arn   = aws_resiliencehubv2_service.test.arn
  cfn_stack_arn = aws_cloudformation_stack.test.id
}
`, rName)
}
