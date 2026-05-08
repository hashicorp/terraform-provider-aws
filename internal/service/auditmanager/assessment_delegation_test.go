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

func TestAccAuditManagerAssessmentDelegation_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var delegation types.DelegationMetadata
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_auditmanager_assessment_delegation.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.AuditManagerEndpointID)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.AuditManagerServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckAssessmentDelegationDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccAssessmentDelegationConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAssessmentDelegationExists(ctx, t, resourceName, &delegation),
					resource.TestCheckResourceAttrPair(resourceName, "assessment_id", "aws_auditmanager_assessment.test", names.AttrID),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrRoleARN, "aws_iam_role.test_delegation", names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, "control_set_id", rName),
					resource.TestCheckResourceAttr(resourceName, "role_type", string(types.RoleTypeResourceOwner)),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"role_type"},
			},
		},
	})
}

func TestAccAuditManagerAssessmentDelegation_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	var delegation types.DelegationMetadata
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_auditmanager_assessment_delegation.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.AuditManagerEndpointID)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.AuditManagerServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckAssessmentDelegationDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccAssessmentDelegationConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAssessmentDelegationExists(ctx, t, resourceName, &delegation),
					acctest.CheckFrameworkResourceDisappears(ctx, t, tfauditmanager.ResourceAssessmentDelegation, resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccAuditManagerAssessmentDelegation_optional(t *testing.T) {
	ctx := acctest.Context(t)
	var delegation types.DelegationMetadata
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_auditmanager_assessment_delegation.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.AuditManagerEndpointID)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.AuditManagerServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckAssessmentDelegationDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccAssessmentDelegationConfig_optional(rName, "text"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAssessmentDelegationExists(ctx, t, resourceName, &delegation),
					resource.TestCheckResourceAttrPair(resourceName, "assessment_id", "aws_auditmanager_assessment.test", names.AttrID),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrRoleARN, "aws_iam_role.test_delegation", names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, "control_set_id", rName),
					resource.TestCheckResourceAttr(resourceName, "role_type", string(types.RoleTypeResourceOwner)),
					resource.TestCheckResourceAttr(resourceName, names.AttrComment, "text"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"role_type", names.AttrComment},
			},
			{
				Config: testAccAssessmentDelegationConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAssessmentDelegationExists(ctx, t, resourceName, &delegation),
					resource.TestCheckResourceAttrPair(resourceName, "assessment_id", "aws_auditmanager_assessment.test", names.AttrID),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrRoleARN, "aws_iam_role.test_delegation", names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, "control_set_id", rName),
					resource.TestCheckResourceAttr(resourceName, "role_type", string(types.RoleTypeResourceOwner)),
				),
			},
			{
				Config: testAccAssessmentDelegationConfig_optional(rName, "text-updated"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAssessmentDelegationExists(ctx, t, resourceName, &delegation),
					resource.TestCheckResourceAttrPair(resourceName, "assessment_id", "aws_auditmanager_assessment.test", names.AttrID),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrRoleARN, "aws_iam_role.test_delegation", names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, "control_set_id", rName),
					resource.TestCheckResourceAttr(resourceName, "role_type", string(types.RoleTypeResourceOwner)),
					resource.TestCheckResourceAttr(resourceName, names.AttrComment, "text-updated"),
				),
			},
		},
	})
}

func TestAccAuditManagerAssessmentDelegation_multiple(t *testing.T) {
	ctx := acctest.Context(t)
	var delegation1 types.DelegationMetadata
	var delegation2 types.DelegationMetadata
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName1 := "aws_auditmanager_assessment_delegation.test1"
	resourceName2 := "aws_auditmanager_assessment_delegation.test2"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.AuditManagerEndpointID)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.AuditManagerServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckAssessmentDelegationDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccAssessmentDelegationConfig_multiple(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAssessmentDelegationExists(ctx, t, resourceName1, &delegation1),
					resource.TestCheckResourceAttrPair(resourceName1, "assessment_id", "aws_auditmanager_assessment.test", names.AttrID),
					resource.TestCheckResourceAttrPair(resourceName1, names.AttrRoleARN, "aws_iam_role.test_delegation1", names.AttrARN),
					resource.TestCheckResourceAttr(resourceName1, "control_set_id", rName),
					resource.TestCheckResourceAttr(resourceName1, "role_type", string(types.RoleTypeResourceOwner)),
					testAccCheckAssessmentDelegationExists(ctx, t, resourceName2, &delegation2),
					resource.TestCheckResourceAttrPair(resourceName2, "assessment_id", "aws_auditmanager_assessment.test", names.AttrID),
					resource.TestCheckResourceAttrPair(resourceName2, names.AttrRoleARN, "aws_iam_role.test_delegation2", names.AttrARN),
					resource.TestCheckResourceAttr(resourceName2, "control_set_id", rName),
					resource.TestCheckResourceAttr(resourceName2, "role_type", string(types.RoleTypeResourceOwner)),
				),
			},
		},
	})
}

func testAccCheckAssessmentDelegationDestroy(ctx context.Context, t *testing.T) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.ProviderMeta(ctx, t).AuditManagerClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_auditmanager_assessment_delegation" {
				continue
			}

			_, err := tfauditmanager.FindAssessmentDelegationByThreePartKey(ctx, conn, rs.Primary.Attributes["assessment_id"], rs.Primary.Attributes[names.AttrRoleARN], rs.Primary.Attributes["control_set_id"])

			if retry.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("Audit Manager Assessment Delegation %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccCheckAssessmentDelegationExists(ctx context.Context, t *testing.T, n string, v *types.DelegationMetadata) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.ProviderMeta(ctx, t).AuditManagerClient(ctx)

		output, err := tfauditmanager.FindAssessmentDelegationByThreePartKey(ctx, conn, rs.Primary.Attributes["assessment_id"], rs.Primary.Attributes[names.AttrRoleARN], rs.Primary.Attributes["control_set_id"])

		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

func testAccAssessmentDelegationConfig_base(rName string) string {
	return fmt.Sprintf(`
data "aws_caller_identity" "current" {}

resource "aws_s3_bucket" "test" {
  bucket        = %[1]q
  force_destroy = true
}

data "aws_iam_policy_document" "test" {
  statement {
    actions = ["sts:AssumeRole"]
    principals {
      type        = "Service"
      identifiers = ["auditmanager.amazonaws.com"]
    }
  }
}

resource "aws_iam_role" "test" {
  name               = %[1]q
  assume_role_policy = data.aws_iam_policy_document.test.json
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

func testAccAssessmentDelegationConfig_basic(rName string) string {
	return acctest.ConfigCompose(
		testAccAssessmentDelegationConfig_base(rName),
		fmt.Sprintf(`
resource "aws_iam_role" "test_delegation" {
  name               = "%[1]s-delegation"
  assume_role_policy = data.aws_iam_policy_document.test.json
}

resource "aws_auditmanager_assessment_delegation" "test" {
  assessment_id  = aws_auditmanager_assessment.test.id
  role_arn       = aws_iam_role.test_delegation.arn
  role_type      = "RESOURCE_OWNER"
  control_set_id = %[1]q
}
`, rName))
}

func testAccAssessmentDelegationConfig_optional(rName, comment string) string {
	return acctest.ConfigCompose(
		testAccAssessmentDelegationConfig_base(rName),
		fmt.Sprintf(`
resource "aws_iam_role" "test_delegation" {
  name               = "%[1]s-delegation"
  assume_role_policy = data.aws_iam_policy_document.test.json
}

resource "aws_auditmanager_assessment_delegation" "test" {
  assessment_id  = aws_auditmanager_assessment.test.id
  role_arn       = aws_iam_role.test_delegation.arn
  role_type      = "RESOURCE_OWNER"
  control_set_id = %[1]q

  comment = %[2]q
}
`, rName, comment))
}

func testAccAssessmentDelegationConfig_multiple(rName string) string {
	return acctest.ConfigCompose(
		testAccAssessmentDelegationConfig_base(rName),
		fmt.Sprintf(`
resource "aws_iam_role" "test_delegation1" {
  name               = "%[1]s-delegation1"
  assume_role_policy = data.aws_iam_policy_document.test.json
}

resource "aws_auditmanager_assessment_delegation" "test1" {
  assessment_id  = aws_auditmanager_assessment.test.id
  role_arn       = aws_iam_role.test_delegation1.arn
  role_type      = "RESOURCE_OWNER"
  control_set_id = %[1]q
}

resource "aws_iam_role" "test_delegation2" {
  name               = "%[1]s-delegation2"
  assume_role_policy = data.aws_iam_policy_document.test.json
}

resource "aws_auditmanager_assessment_delegation" "test2" {
  assessment_id  = aws_auditmanager_assessment.test.id
  role_arn       = aws_iam_role.test_delegation2.arn
  role_type      = "RESOURCE_OWNER"
  control_set_id = %[1]q
}
`, rName))
}
