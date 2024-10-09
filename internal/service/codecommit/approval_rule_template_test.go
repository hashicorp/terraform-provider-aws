// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package codecommit_test

import (
	"context"
	"fmt"
	"testing"

	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfcodecommit "github.com/hashicorp/terraform-provider-aws/internal/service/codecommit"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccCodeCommitApprovalRuleTemplate_basic(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_codecommit_approval_rule_template.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.CodeCommitServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckApprovalRuleTemplateDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccApprovalRuleTemplateConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckApprovalRuleTemplateExists(ctx, resourceName),
					testAccCheckApprovalRuleTemplateContent(resourceName, 2),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttrSet(resourceName, "approval_rule_template_id"),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrCreationDate),
					resource.TestCheckResourceAttrSet(resourceName, "last_modified_date"),
					resource.TestCheckResourceAttrSet(resourceName, "last_modified_user"),
					resource.TestCheckResourceAttrSet(resourceName, "rule_content_sha256"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccCodeCommitApprovalRuleTemplate_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_codecommit_approval_rule_template.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.CodeCommitServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckApprovalRuleTemplateDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccApprovalRuleTemplateConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckApprovalRuleTemplateExists(ctx, resourceName),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfcodecommit.ResourceApprovalRuleTemplate(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccCodeCommitApprovalRuleTemplate_updateContentAndDescription(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_codecommit_approval_rule_template.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.CodeCommitServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckApprovalRuleTemplateDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccApprovalRuleTemplateConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckApprovalRuleTemplateExists(ctx, resourceName),
					testAccCheckApprovalRuleTemplateContent(resourceName, 2),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, ""),
				),
			},
			{
				Config: testAccApprovalRuleTemplateConfig_updateContentAndDescription(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckApprovalRuleTemplateExists(ctx, resourceName),
					testAccCheckApprovalRuleTemplateContent(resourceName, 1),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, "This is a test description"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccCodeCommitApprovalRuleTemplate_updateName(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	rNameUpdated := sdkacctest.RandomWithPrefix("tf-acc-test-update")
	resourceName := "aws_codecommit_approval_rule_template.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.CodeCommitServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckApprovalRuleTemplateDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccApprovalRuleTemplateConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckApprovalRuleTemplateExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
				),
			},
			{
				Config: testAccApprovalRuleTemplateConfig_basic(rNameUpdated),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckApprovalRuleTemplateExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rNameUpdated),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccCheckApprovalRuleTemplateContent(resourceName string, numApprovals int) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		expectedContent := fmt.Sprintf(`{"Version":"2018-11-08","DestinationReferences":["refs/heads/master"],"Statements":[{"Type":"Approvers","NumberOfApprovalsNeeded":%d,"ApprovalPoolMembers":["arn:%s:sts::%s:assumed-role/CodeCommitReview/*"]}]}`,
			numApprovals, acctest.Partition(), acctest.AccountID(),
		)
		return resource.TestCheckResourceAttr(resourceName, names.AttrContent, expectedContent)(s)
	}
}

func testAccCheckApprovalRuleTemplateExists(ctx context.Context, n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).CodeCommitClient(ctx)

		_, err := tfcodecommit.FindApprovalRuleTemplateByName(ctx, conn, rs.Primary.ID)

		return err
	}
}

func testAccCheckApprovalRuleTemplateDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).CodeCommitClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_codecommit_approval_rule_template" {
				continue
			}

			_, err := tfcodecommit.FindApprovalRuleTemplateByName(ctx, conn, rs.Primary.ID)

			if tfresource.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("CodeCommit Approval Rule Template (%s) still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccApprovalRuleTemplateConfig_basic(rName string) string {
	return fmt.Sprintf(`
data "aws_caller_identity" "current" {}

data "aws_partition" "current" {}

resource "aws_codecommit_approval_rule_template" "test" {
  name = %[1]q

  content = <<EOF
  {
	  "Version": "2018-11-08",
	  "DestinationReferences": ["refs/heads/master"],
	  "Statements": [{
			  "Type": "Approvers",
			  "NumberOfApprovalsNeeded": 2,
			  "ApprovalPoolMembers": ["arn:${data.aws_partition.current.partition}:sts::${data.aws_caller_identity.current.account_id}:assumed-role/CodeCommitReview/*"]}]
  }
  EOF 
}
`, rName)
}

func testAccApprovalRuleTemplateConfig_updateContentAndDescription(rName string) string {
	return fmt.Sprintf(`
data "aws_caller_identity" "current" {}

data "aws_partition" "current" {}

resource "aws_codecommit_approval_rule_template" "test" {
  name        = %[1]q
  description = "This is a test description"

  content = <<EOF
  {
	  "Version": "2018-11-08",
	  "DestinationReferences": ["refs/heads/master"],
	  "Statements": [{
			  "Type": "Approvers",
			  "NumberOfApprovalsNeeded": 1,
			  "ApprovalPoolMembers": ["arn:${data.aws_partition.current.partition}:sts::${data.aws_caller_identity.current.account_id}:assumed-role/CodeCommitReview/*"]}]
  }
  EOF 
}
`, rName)
}
