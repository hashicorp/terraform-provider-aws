// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package organizations_test

import (
	"fmt"
	"testing"

	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func testAccResourceTagsDataSource_basic(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_organizations_policy.test"
	dataSourceName := "data.aws_organizations_resource_tags.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckOrganizationsAccount(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.OrganizationsServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccResourceTagsDataSourceConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(resourceName, "tags.#", dataSourceName, "tags.#"),
				),
			},
		},
	})
}

func testAccResourceTagsDataSourceConfig_basic(rName string) string {
	return fmt.Sprintf(`
resource "aws_organizations_organization" "test" {}

resource "aws_organizations_policy" "test" {
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

  depends_on = [aws_organizations_organization.test]

  tags = {
    TerraformProviderAwsTest = true
    Alpha                    = 1
  }
}

data "aws_organizations_resource_tags" "test" {
  resource_id = aws_organizations_policy.test.id
}
`, rName)
}
