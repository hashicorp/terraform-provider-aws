// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package backup_test

import (
	"fmt"
	"testing"

	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccBackupSelectionDataSource_basic(t *testing.T) {
	ctx := acctest.Context(t)
	datasourceName := "data.aws_backup_selection.test"
	resourceName := "aws_backup_selection.test"
	rInt := sdkacctest.RandInt()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.BackupServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccSelectionDataSourceConfig_basic(rInt),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(datasourceName, names.AttrName, resourceName, names.AttrName),
					resource.TestCheckResourceAttrPair(datasourceName, names.AttrIAMRoleARN, resourceName, names.AttrIAMRoleARN),
					resource.TestCheckResourceAttrPair(datasourceName, "resources.#", resourceName, "resources.#"),
				),
			},
		},
	})
}

func testAccSelectionDataSourceConfig_basic(rInt int) string {
	return fmt.Sprintf(`
data "aws_caller_identity" "current" {}

data "aws_partition" "current" {}

data "aws_region" "current" {}

resource "aws_backup_vault" "test" {
  name = "tf_acc_test_backup_vault_%[1]d"
}

resource "aws_backup_plan" "test" {
  name = "tf_acc_test_backup_plan_%[1]d"

  rule {
    rule_name         = "tf_acc_test_backup_rule_%[1]d"
    target_vault_name = aws_backup_vault.test.name
    schedule          = "cron(0 12 * * ? *)"
  }
}

resource "aws_backup_selection" "test" {
  plan_id      = aws_backup_plan.test.id
  name         = "tf_acc_test_backup_selection_%[1]d"
  iam_role_arn = "arn:${data.aws_partition.current.partition}:iam::${data.aws_caller_identity.current.account_id}:role/service-role/AWSBackupDefaultServiceRole"

  selection_tag {
    type  = "STRINGEQUALS"
    key   = "foo"
    value = "bar"
  }

  condition {}
  not_resources = []

  resources = [
    "arn:${data.aws_partition.current.partition}:ec2:${data.aws_region.current.name}:${data.aws_caller_identity.current.account_id}:volume/*"
  ]
}

data "aws_backup_selection" "test" {
  plan_id      = aws_backup_plan.test.id
  selection_id = aws_backup_selection.test.id
}
`, rInt)
}
