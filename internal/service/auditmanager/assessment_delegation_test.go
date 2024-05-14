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

func TestAccAuditManagerAssessmentDelegation_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var delegation types.DelegationMetadata
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_auditmanager_assessment_delegation.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.AuditManagerEndpointID)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.AuditManagerServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckAssessmentDelegationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccAssessmentDelegationConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAssessmentDelegationExists(ctx, resourceName, &delegation),
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
				ImportStateVerifyIgnore: []string{"control_set_id", "role_type"},
			},
		},
	})
}

func TestAccAuditManagerAssessmentDelegation_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	var delegation types.DelegationMetadata
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_auditmanager_assessment_delegation.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.AuditManagerEndpointID)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.AuditManagerServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckAssessmentDelegationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccAssessmentDelegationConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAssessmentDelegationExists(ctx, resourceName, &delegation),
					acctest.CheckFrameworkResourceDisappears(ctx, acctest.Provider, tfauditmanager.ResourceAssessmentDelegation, resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccAuditManagerAssessmentDelegation_optional(t *testing.T) {
	ctx := acctest.Context(t)
	var delegation types.DelegationMetadata
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_auditmanager_assessment_delegation.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.AuditManagerEndpointID)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.AuditManagerServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckAssessmentDelegationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccAssessmentDelegationConfig_optional(rName, "text"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAssessmentDelegationExists(ctx, resourceName, &delegation),
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
				ImportStateVerifyIgnore: []string{"control_set_id", "role_type", names.AttrComment},
			},
			{
				Config: testAccAssessmentDelegationConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAssessmentDelegationExists(ctx, resourceName, &delegation),
					resource.TestCheckResourceAttrPair(resourceName, "assessment_id", "aws_auditmanager_assessment.test", names.AttrID),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrRoleARN, "aws_iam_role.test_delegation", names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, "control_set_id", rName),
					resource.TestCheckResourceAttr(resourceName, "role_type", string(types.RoleTypeResourceOwner)),
				),
			},
			{
				Config: testAccAssessmentDelegationConfig_optional(rName, "text-updated"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAssessmentDelegationExists(ctx, resourceName, &delegation),
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
	var delegation types.DelegationMetadata
	var delegation2 types.DelegationMetadata
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_auditmanager_assessment_delegation.test"
	resourceName2 := "aws_auditmanager_assessment_delegation.test2"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.AuditManagerEndpointID)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.AuditManagerServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckAssessmentDelegationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccAssessmentDelegationConfig_multiple(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAssessmentDelegationExists(ctx, resourceName, &delegation),
					resource.TestCheckResourceAttrPair(resourceName, "assessment_id", "aws_auditmanager_assessment.test", names.AttrID),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrRoleARN, "aws_iam_role.test_delegation", names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, "control_set_id", rName),
					resource.TestCheckResourceAttr(resourceName, "role_type", string(types.RoleTypeResourceOwner)),
					testAccCheckAssessmentDelegationExists(ctx, resourceName2, &delegation2),
					resource.TestCheckResourceAttrPair(resourceName2, "assessment_id", "aws_auditmanager_assessment.test", names.AttrID),
					resource.TestCheckResourceAttrPair(resourceName2, names.AttrRoleARN, "aws_iam_role.test_delegation2", names.AttrARN),
					resource.TestCheckResourceAttr(resourceName2, "control_set_id", rName),
					resource.TestCheckResourceAttr(resourceName2, "role_type", string(types.RoleTypeResourceOwner)),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"control_set_id", "role_type"},
			},
			{
				ResourceName:            resourceName2,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"control_set_id", "role_type"},
			},
		},
	})
}

func testAccCheckAssessmentDelegationDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).AuditManagerClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_auditmanager_assessment_delegation" {
				continue
			}

			_, err := tfauditmanager.FindAssessmentDelegationByID(ctx, conn, rs.Primary.ID)
			if err != nil {
				var nfe *retry.NotFoundError
				if errors.As(err, &nfe) {
					return nil
				}
				return err
			}

			return create.Error(names.AuditManager, create.ErrActionCheckingDestroyed, tfauditmanager.ResNameAssessmentDelegation, rs.Primary.ID, errors.New("not destroyed"))
		}

		return nil
	}
}

func testAccCheckAssessmentDelegationExists(ctx context.Context, name string, delegation *types.DelegationMetadata) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return create.Error(names.AuditManager, create.ErrActionCheckingExistence, tfauditmanager.ResNameAssessmentDelegation, name, errors.New("not found"))
		}

		if rs.Primary.ID == "" {
			return create.Error(names.AuditManager, create.ErrActionCheckingExistence, tfauditmanager.ResNameAssessmentDelegation, name, errors.New("not set"))
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).AuditManagerClient(ctx)
		resp, err := tfauditmanager.FindAssessmentDelegationByID(ctx, conn, rs.Primary.ID)
		if err != nil {
			return create.Error(names.AuditManager, create.ErrActionCheckingExistence, tfauditmanager.ResNameAssessmentDelegation, rs.Primary.ID, err)
		}

		*delegation = *resp

		return nil
	}
}

func testAccAssessmentDelegationConfigBase(rName string) string {
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
    aws_services {
      service_name = "S3"
    }
  }
}
`, rName)
}

func testAccAssessmentDelegationConfig_basic(rName string) string {
	return acctest.ConfigCompose(
		testAccAssessmentDelegationConfigBase(rName),
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
		testAccAssessmentDelegationConfigBase(rName),
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
		testAccAssessmentDelegationConfigBase(rName),
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
