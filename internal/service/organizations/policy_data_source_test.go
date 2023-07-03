// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package organizations_test

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/service/organizations"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
)

func testAccPolicyDataSource_UnattachedPolicy(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_organizations_policy.test"
	dataSourceName := "data.aws_organizations_policy.test"

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckOrganizationsAccount(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, organizations.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccPolicyDataSourceConfig_unattachedPolicy(rName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(resourceName, "id", dataSourceName, "policy_id"),
					resource.TestCheckResourceAttrPair(resourceName, "arn", dataSourceName, "arn"),
					resource.TestCheckResourceAttrPair(resourceName, "content", dataSourceName, "content"),
					resource.TestCheckResourceAttrPair(resourceName, "type", dataSourceName, "type"),
					resource.TestCheckResourceAttrPair(resourceName, "name", dataSourceName, "name"),
				),
			},
		},
	})
}

func testAccPolicyDataSourceConfig_unattachedPolicy(rName string) string {
	return fmt.Sprintf(`
resource "aws_organizations_organization" "test" {
  feature_set          = "ALL"
  enabled_policy_types = ["SERVICE_CONTROL_POLICY", "TAG_POLICY", "BACKUP_POLICY", "AISERVICES_OPT_OUT_POLICY"]
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

data "aws_organizations_policy" "test" {
  policy_id = aws_organizations_policy.test.id
}
`, rName)
}
