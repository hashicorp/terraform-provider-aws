// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package organizations_test

import (
	"fmt"
	"testing"

	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccOrganizationsPoliciesForTargetDataSource_basic(t *testing.T) {
	ctx := acctest.Context(t)
	datasourceName := "data.aws_organizations_policies_for_target.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckOrganizationsAccount(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.OrganizationsServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccPoliciesForTargetDataSourceConfig_AttachQuery(rName),
				Check: resource.ComposeTestCheckFunc(
					acctest.CheckResourceAttrGreaterThanValue(datasourceName, "ids.#", 0),
				),
			},
		},
	})
}

func testAccPoliciesForTargetDataSourceConfig_AttachQuery(rName string) string {
	return fmt.Sprintf(`
resource "aws_organizations_organization" "test" {
  feature_set          = "ALL"
  enabled_policy_types = ["SERVICE_CONTROL_POLICY", "TAG_POLICY", "BACKUP_POLICY", "AISERVICES_OPT_OUT_POLICY"]
}

resource "aws_organizations_organizational_unit" "test" {
  name      = %[1]q
  parent_id = aws_organizations_organization.test.roots[0].id
}

resource "aws_organizations_policy" "test" {
  depends_on = [aws_organizations_organization.test]

  content = <<EOF
{
  "Version": "2012-10-17",
  "Statement": {
    "Effect": "Allow",
    "Action": "*",
    "Resource": "*"
  }
}
EOF

  name = %[1]q
}

resource "aws_organizations_policy_attachment" "test" {
  depends_on = [aws_organizations_policy.test]
  policy_id  = aws_organizations_policy.test.id
  target_id  = aws_organizations_organizational_unit.test.id
}

data "aws_organizations_policies_for_target" "test" {
  depends_on = [aws_organizations_policy_attachment.test]
  target_id  = aws_organizations_organizational_unit.test.id
  filter     = "SERVICE_CONTROL_POLICY"
}

data "aws_organizations_policy" "test" {
  policy_id = data.aws_organizations_policies_for_target.test.ids[0]
}
`, rName)
}
