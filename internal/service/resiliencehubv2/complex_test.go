// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package resiliencehubv2_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccResilienceHubV2_complex(t *testing.T) {
	ctx := acctest.Context(t)
	rName := acctest.RandomWithPrefix(t, "tf-acc-test")

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.ResilienceHubV2),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccComplexConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					// Policy
					resource.TestCheckResourceAttr("aws_resiliencehubv2_policy.test", names.AttrName, rName+"-policy"),
					resource.TestCheckResourceAttr("aws_resiliencehubv2_policy.test", "availability_slo.0.target", "99.9"),
					resource.TestCheckResourceAttr("aws_resiliencehubv2_policy.test", "multi_az.0.disaster_recovery_approach", "ACTIVE_ACTIVE"),

					// System
					resource.TestCheckResourceAttr("aws_resiliencehubv2_system.test", names.AttrName, rName+"-system"),

					// Service references policy
					resource.TestCheckResourceAttr("aws_resiliencehubv2_service.test", names.AttrName, rName+"-service"),
					resource.TestCheckResourceAttrPair("aws_resiliencehubv2_service.test", "policy_arn", "aws_resiliencehubv2_policy.test", names.AttrARN),
					resource.TestCheckResourceAttr("aws_resiliencehubv2_service.test", "permission_model.0.invoker_role_name", "AWSResilienceHubAssessmentRole"),

					// UserJourney references system
					resource.TestCheckResourceAttr("aws_resiliencehubv2_user_journey.test", names.AttrName, rName+"-journey"),
					resource.TestCheckResourceAttrPair("aws_resiliencehubv2_user_journey.test", "system_arn", "aws_resiliencehubv2_system.test", names.AttrARN),

					// ServiceFunction references service
					resource.TestCheckResourceAttr("aws_resiliencehubv2_service_function.test", names.AttrName, rName+"-function"),
					resource.TestCheckResourceAttr("aws_resiliencehubv2_service_function.test", "criticality", "PRIMARY"),
					resource.TestCheckResourceAttrPair("aws_resiliencehubv2_service_function.test", "service_arn", "aws_resiliencehubv2_service.test", names.AttrARN),

					// InputSource references service
					resource.TestCheckResourceAttrSet("aws_resiliencehubv2_input_source.test", "input_source_id"),
					resource.TestCheckResourceAttrPair("aws_resiliencehubv2_input_source.test", "service_arn", "aws_resiliencehubv2_service.test", names.AttrARN),
				),
			},
		},
	})
}

func testAccComplexConfig(rName string) string {
	return fmt.Sprintf(`
data "aws_region" "current" {}

resource "aws_resiliencehubv2_policy" "test" {
  name = "%[1]s-policy"

  availability_slo {
    target = 99.9
  }

  multi_az {
    disaster_recovery_approach = "ACTIVE_ACTIVE"
    rpo_in_minutes             = 5
    rto_in_minutes             = 10
  }
}

resource "aws_resiliencehubv2_system" "test" {
  name        = "%[1]s-system"
  description = "Complex test system"
}

resource "aws_resiliencehubv2_service" "test" {
  name    = "%[1]s-service"
  regions = [data.aws_region.current.name]

  policy_arn = aws_resiliencehubv2_policy.test.arn

  permission_model {
    invoker_role_name = "AWSResilienceHubAssessmentRole"
  }
}

resource "aws_resiliencehubv2_user_journey" "test" {
  name       = "%[1]s-journey"
  system_arn = aws_resiliencehubv2_system.test.arn
}

resource "aws_resiliencehubv2_service_function" "test" {
  name        = "%[1]s-function"
  service_arn = aws_resiliencehubv2_service.test.arn
  criticality = "PRIMARY"
}

resource "aws_cloudformation_stack" "test" {
  name = "%[1]s-stack"

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

resource "aws_resiliencehubv2_input_source" "test" {
  service_arn   = aws_resiliencehubv2_service.test.arn
  cfn_stack_arn = aws_cloudformation_stack.test.id
}
`, rName)
}
