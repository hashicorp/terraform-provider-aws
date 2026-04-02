// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package auditmanager_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/service/auditmanager/types"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	tfauditmanager "github.com/hashicorp/terraform-provider-aws/internal/service/auditmanager"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccAuditManagerAssessment_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var assessment types.Assessment
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_auditmanager_assessment.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.AuditManagerEndpointID)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.AuditManagerServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckAssessmentDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccAssessmentConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAssessmentExists(ctx, t, resourceName, &assessment),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, "assessment_reports_destination.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "assessment_reports_destination.0.destination_type", "S3"),
					resource.TestCheckResourceAttrPair(resourceName, "framework_id", "aws_auditmanager_framework.test", names.AttrID),
					resource.TestCheckResourceAttr(resourceName, "scope.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "scope.0.aws_accounts.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "scope.0.aws_services.#", "0"),
					acctest.MatchResourceAttrRegionalARN(ctx, resourceName, names.AttrARN, "auditmanager", regexache.MustCompile(`assessment/.+$`)),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"roles"},
			},
		},
	})
}

func TestAccAuditManagerAssessment_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	var assessment types.Assessment
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_auditmanager_assessment.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.AuditManagerEndpointID)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.AuditManagerServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckAssessmentDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccAssessmentConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAssessmentExists(ctx, t, resourceName, &assessment),
					acctest.CheckFrameworkResourceDisappears(ctx, t, tfauditmanager.ResourceAssessment, resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccAuditManagerAssessment_tags(t *testing.T) {
	ctx := acctest.Context(t)
	var assessment types.Assessment
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_auditmanager_assessment.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.AuditManagerEndpointID)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.AuditManagerServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckAssessmentDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccAssessmentConfig_tags1(rName, acctest.CtKey1, acctest.CtValue1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAssessmentExists(ctx, t, resourceName, &assessment),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, "1"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey1, acctest.CtValue1),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"roles"},
			},
			{
				Config: testAccAssessmentConfig_tags2(rName, acctest.CtKey1, acctest.CtValue1Updated, acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAssessmentExists(ctx, t, resourceName, &assessment),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, "2"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey1, acctest.CtValue1Updated),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey2, acctest.CtValue2),
				),
			},
			{
				Config: testAccAssessmentConfig_tags1(rName, acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAssessmentExists(ctx, t, resourceName, &assessment),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, "1"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey2, acctest.CtValue2),
				),
			},
		},
	})
}

func TestAccAuditManagerAssessment_optional(t *testing.T) {
	ctx := acctest.Context(t)
	var assessment types.Assessment
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_auditmanager_assessment.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.AuditManagerEndpointID)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.AuditManagerServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckAssessmentDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccAssessmentConfig_optional(rName, "text"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAssessmentExists(ctx, t, resourceName, &assessment),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, "text"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"roles"},
			},
			{
				Config: testAccAssessmentConfig_optional(rName, "text updated"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAssessmentExists(ctx, t, resourceName, &assessment),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, "text updated"),
				),
			},
		},
	})
}

func testAccCheckAssessmentDestroy(ctx context.Context, t *testing.T) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.ProviderMeta(ctx, t).AuditManagerClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_auditmanager_assessment" {
				continue
			}

			_, err := tfauditmanager.FindAssessmentByID(ctx, conn, rs.Primary.ID)

			if retry.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("Audit Manager Assessment %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccCheckAssessmentExists(ctx context.Context, t *testing.T, n string, v *types.Assessment) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.ProviderMeta(ctx, t).AuditManagerClient(ctx)

		output, err := tfauditmanager.FindAssessmentByID(ctx, conn, rs.Primary.ID)

		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

func testAccAssessmentConfig_base(rName string) string {
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
`, rName)
}

func testAccAssessmentConfig_basic(rName string) string {
	return acctest.ConfigCompose(
		testAccAssessmentConfig_base(rName),
		fmt.Sprintf(`
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
`, rName))
}

func testAccAssessmentConfig_tags1(rName, tagKey1, tagValue1 string) string {
	return acctest.ConfigCompose(
		testAccAssessmentConfig_base(rName),
		fmt.Sprintf(`
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

  tags = {
    %[2]q = %[3]q
  }
}
`, rName, tagKey1, tagValue1))
}

func testAccAssessmentConfig_tags2(rName, tagKey1, tagValue1, tagKey2, tagValue2 string) string {
	return acctest.ConfigCompose(
		testAccAssessmentConfig_base(rName),
		fmt.Sprintf(`
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

  tags = {
    %[2]q = %[3]q
    %[4]q = %[5]q
  }
}
`, rName, tagKey1, tagValue1, tagKey2, tagValue2))
}

func testAccAssessmentConfig_optional(rName, description string) string {
	return acctest.ConfigCompose(
		testAccAssessmentConfig_base(rName),
		fmt.Sprintf(`
resource "aws_auditmanager_assessment" "test" {
  name        = %[1]q
  description = %[2]q

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
`, rName, description))
}
