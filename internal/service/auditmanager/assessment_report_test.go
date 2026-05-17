// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package auditmanager_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/auditmanager/types"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	tfauditmanager "github.com/hashicorp/terraform-provider-aws/internal/service/auditmanager"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccAuditManagerAssessmentReport_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var assessmentReport types.AssessmentReportMetadata
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_auditmanager_assessment_report.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.AuditManagerEndpointID)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.AuditManagerServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckAssessmentReportDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccAssessmentReportConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAssessmentReportExists(ctx, t, resourceName, &assessmentReport),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttrPair(resourceName, "assessment_id", "aws_auditmanager_assessment.test", names.AttrID),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{names.AttrStatus},
			},
		},
	})
}

func TestAccAuditManagerAssessmentReport_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	var assessmentReport types.AssessmentReportMetadata
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_auditmanager_assessment_report.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.AuditManagerEndpointID)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.AuditManagerServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckAssessmentReportDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccAssessmentReportConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAssessmentReportExists(ctx, t, resourceName, &assessmentReport),
					acctest.CheckFrameworkResourceDisappears(ctx, t, tfauditmanager.ResourceAssessmentReport, resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccAuditManagerAssessmentReport_optional(t *testing.T) {
	ctx := acctest.Context(t)
	var assessmentReport types.AssessmentReportMetadata
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_auditmanager_assessment_report.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.AuditManagerEndpointID)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.AuditManagerServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckAssessmentReportDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccAssessmentReportConfig_optional(rName, "text"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAssessmentReportExists(ctx, t, resourceName, &assessmentReport),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttrPair(resourceName, "assessment_id", "aws_auditmanager_assessment.test", names.AttrID),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, "text"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{names.AttrStatus},
			},
			{
				Config: testAccAssessmentReportConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAssessmentReportExists(ctx, t, resourceName, &assessmentReport),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttrPair(resourceName, "assessment_id", "aws_auditmanager_assessment.test", names.AttrID),
				),
			},
			{
				Config: testAccAssessmentReportConfig_optional(rName, "text-updated"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAssessmentReportExists(ctx, t, resourceName, &assessmentReport),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttrPair(resourceName, "assessment_id", "aws_auditmanager_assessment.test", names.AttrID),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, "text-updated"),
				),
			},
		},
	})
}

func testAccCheckAssessmentReportDestroy(ctx context.Context, t *testing.T) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.ProviderMeta(ctx, t).AuditManagerClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_auditmanager_assessment_report" {
				continue
			}

			_, err := tfauditmanager.FindAssessmentReportByID(ctx, conn, rs.Primary.ID)

			if retry.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("Audit Manager Assessment Report %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccCheckAssessmentReportExists(ctx context.Context, t *testing.T, n string, v *types.AssessmentReportMetadata) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.ProviderMeta(ctx, t).AuditManagerClient(ctx)

		output, err := tfauditmanager.FindAssessmentReportByID(ctx, conn, rs.Primary.ID)

		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

func testAccAssessmentReportConfig_base(rName string) string {
	return fmt.Sprintf(`
data "aws_caller_identity" "current" {}

resource "aws_s3_bucket" "test" {
  bucket        = %[1]q
  force_destroy = true
}

resource "aws_iam_role" "test" {
  name = %[1]q

  assume_role_policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Action = "sts:AssumeRole"
        Effect = "Allow"
        Sid    = ""
        Principal = {
          Service = "ec2.amazonaws.com"
        }
      },
    ]
  })
}

resource "aws_auditmanager_control" "test" {
  name = %[1]q

  control_mapping_sources {
    source_name          = %[1]q
    source_set_up_option = "Procedural_Controls_Mapping"
    source_type          = "MANUAL"
  }
}

resource "aws_auditmanager_framework" "test" {
  name = %[1]q

  control_sets {
    name = %[1]q
    controls {
      id = aws_auditmanager_control.test.id
    }
  }
}

resource "aws_auditmanager_assessment" "test" {
  name = %[1]q

  assessment_reports_destination {
    destination      = "s3://${aws_s3_bucket.test.id}"
    destination_type = "S3"
  }

  framework_id = aws_auditmanager_framework.test.id

  roles {
    role_arn  = aws_iam_role.test.arn
    role_type = "PROCESS_OWNER"
  }

  scope {
    aws_accounts {
      id = data.aws_caller_identity.current.account_id
    }
  }
}
`, rName)
}

func testAccAssessmentReportConfig_basic(rName string) string {
	return acctest.ConfigCompose(
		testAccAssessmentReportConfig_base(rName),
		fmt.Sprintf(`
resource "aws_auditmanager_assessment_report" "test" {
  name          = %[1]q
  assessment_id = aws_auditmanager_assessment.test.id
}
`, rName))
}

func testAccAssessmentReportConfig_optional(rName, description string) string {
	return acctest.ConfigCompose(
		testAccAssessmentReportConfig_base(rName),
		fmt.Sprintf(`
resource "aws_auditmanager_assessment_report" "test" {
  name          = %[1]q
  assessment_id = aws_auditmanager_assessment.test.id

  description = %[2]q
}
`, rName, description))
}
