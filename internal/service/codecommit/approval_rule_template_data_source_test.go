// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package codecommit_test

import (
	"fmt"
	"testing"

	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccCodeCommitApprovalRuleTemplateDataSource_basic(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_codecommit_approval_rule_template.test"
	datasourceName := "data.aws_codecommit_approval_rule_template.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.CodeCommitServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccApprovalRuleTemplateDataSourceConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(datasourceName, names.AttrDescription, resourceName, names.AttrDescription),
					resource.TestCheckResourceAttrPair(datasourceName, names.AttrName, resourceName, names.AttrName),
					resource.TestCheckResourceAttrPair(datasourceName, names.AttrContent, resourceName, names.AttrContent),
					resource.TestCheckResourceAttrPair(datasourceName, names.AttrCreationDate, resourceName, names.AttrCreationDate),
					resource.TestCheckResourceAttrPair(datasourceName, "last_modified_date", resourceName, "last_modified_date"),
					resource.TestCheckResourceAttrPair(datasourceName, "last_modified_user", resourceName, "last_modified_user"),
					resource.TestCheckResourceAttrPair(datasourceName, "rule_content_sha256", resourceName, "rule_content_sha256"),
				),
			},
		},
	})
}

func testAccApprovalRuleTemplateDataSourceConfig_basic(rName string) string {
	return fmt.Sprintf(`
data "aws_caller_identity" "current" {}

data "aws_partition" "current" {}

resource "aws_codecommit_approval_rule_template" "test" {
  description = %[1]q
  name        = %[1]q
  content     = <<EOF
{
	"Version": "2018-11-08",
	"DestinationReferences": ["refs/heads/master"],
	"Statements": [{
			"Type": "Approvers",
			"NumberOfApprovalsNeeded": 2,
			"ApprovalPoolMembers": ["arn:${data.aws_partition.current.partition}:sts::${data.aws_caller_identity.current.account_id}:assumed-role/CodeCommitReview/*"]
	}]
}
EOF
}

data "aws_codecommit_approval_rule_template" "test" {
  name = aws_codecommit_approval_rule_template.test.name
}
`, rName)
}
