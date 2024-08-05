// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package auditmanager_test

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/auditmanager/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	tfauditmanager "github.com/hashicorp/terraform-provider-aws/internal/service/auditmanager"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccAuditManagerAssessmentReport_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var assessmentReport types.AssessmentReportMetadata
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_auditmanager_assessment_report.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.AuditManagerEndpointID)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.AuditManagerServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckAssessmentReportDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccAssessmentReportConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAssessmentReportExists(ctx, resourceName, &assessmentReport),
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
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_auditmanager_assessment_report.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.AuditManagerEndpointID)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.AuditManagerServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckAssessmentReportDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccAssessmentReportConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAssessmentReportExists(ctx, resourceName, &assessmentReport),
					acctest.CheckFrameworkResourceDisappears(ctx, acctest.Provider, tfauditmanager.ResourceAssessmentReport, resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccAuditManagerAssessmentReport_optional(t *testing.T) {
	ctx := acctest.Context(t)
	var assessmentReport types.AssessmentReportMetadata
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_auditmanager_assessment_report.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.AuditManagerEndpointID)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.AuditManagerServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckAssessmentReportDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccAssessmentReportConfig_optional(rName, "text"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAssessmentReportExists(ctx, resourceName, &assessmentReport),
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
					testAccCheckAssessmentReportExists(ctx, resourceName, &assessmentReport),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttrPair(resourceName, "assessment_id", "aws_auditmanager_assessment.test", names.AttrID),
				),
			},
			{
				Config: testAccAssessmentReportConfig_optional(rName, "text-updated"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAssessmentReportExists(ctx, resourceName, &assessmentReport),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttrPair(resourceName, "assessment_id", "aws_auditmanager_assessment.test", names.AttrID),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, "text-updated"),
				),
			},
		},
	})
}

func testAccCheckAssessmentReportDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).AuditManagerClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_auditmanager_assessment_report" {
				continue
			}

			_, err := tfauditmanager.FindAssessmentReportByID(ctx, conn, rs.Primary.ID)
			if err != nil {
				var nfe *retry.NotFoundError
				if errors.As(err, &nfe) {
					return nil
				}
				return err
			}

			return create.Error(names.AuditManager, create.ErrActionCheckingDestroyed, tfauditmanager.ResNameAssessmentReport, rs.Primary.ID, errors.New("not destroyed"))
		}

		return nil
	}
}

func testAccCheckAssessmentReportExists(ctx context.Context, name string, assessmentReport *types.AssessmentReportMetadata) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return create.Error(names.AuditManager, create.ErrActionCheckingExistence, tfauditmanager.ResNameAssessmentReport, name, errors.New("not found"))
		}

		if rs.Primary.ID == "" {
			return create.Error(names.AuditManager, create.ErrActionCheckingExistence, tfauditmanager.ResNameAssessmentReport, name, errors.New("not set"))
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).AuditManagerClient(ctx)
		resp, err := tfauditmanager.FindAssessmentReportByID(ctx, conn, rs.Primary.ID)
		if err != nil {
			return create.Error(names.AuditManager, create.ErrActionCheckingExistence, tfauditmanager.ResNameAssessmentReport, rs.Primary.ID, err)
		}

		*assessmentReport = *resp

		return nil
	}
}

func testAccAssessmentReportConfigBase(rName string) string {
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
    aws_services {
      service_name = "S3"
    }
  }
}
`, rName)
}

func testAccAssessmentReportConfig_basic(rName string) string {
	return acctest.ConfigCompose(
		testAccAssessmentReportConfigBase(rName),
		fmt.Sprintf(`
resource "aws_auditmanager_assessment_report" "test" {
  name          = %[1]q
  assessment_id = aws_auditmanager_assessment.test.id
}
`, rName))
}

func testAccAssessmentReportConfig_optional(rName, description string) string {
	return acctest.ConfigCompose(
		testAccAssessmentReportConfigBase(rName),
		fmt.Sprintf(`
resource "aws_auditmanager_assessment_report" "test" {
  name          = %[1]q
  assessment_id = aws_auditmanager_assessment.test.id

  description = %[2]q
}
`, rName, description))
}
