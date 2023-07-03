// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package organizations_test

import (
	"fmt"
	"strconv"
	"testing"

	"github.com/aws/aws-sdk-go/service/organizations"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
)

func TestAccOrganizationsPoliciesDataSource_basic(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	serviceControlPolicyContent := `{"Version": "2012-10-17", "Statement": { "Effect": "Deny", "Action": "*", "Resource": "*"}}`
	datasourceName := "data.aws_organizations_policies.test"

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckOrganizationManagementAccount(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, organizations.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccPoliciesDataSourceConfig_ServiceControlPolicy(rName, organizations.PolicyTypeServiceControlPolicy, serviceControlPolicyContent),
				Check: resource.ComposeTestCheckFunc(
					acctest.CheckResourceAttrGreaterThanOrEqualValue(datasourceName, "ids.#", 1),
				),
			},
		},
	})
}

func testAccPoliciesDataSourceConfig_ServiceControlPolicy(rName, policyType, policyContent string) string {
	return fmt.Sprintf(`
resource "aws_organizations_policy" "test" {
  name    = %[1]q
  type    = %[2]q
  content = %[3]s
}

data "aws_organizations_policies" "test" {
  filter = aws_organizations_policy.test.type
}
`, rName, policyType, strconv.Quote(policyContent))
}
