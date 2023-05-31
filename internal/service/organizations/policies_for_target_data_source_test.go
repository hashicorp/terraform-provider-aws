package organizations_test

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/service/organizations"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
)

func TestAccOrganizationsPoliciesForTargetDataSource_basic(t *testing.T) {
	ctx := acctest.Context(t)
	dataSourceName := "data.aws_organizations_policies_for_target.test"
	policyResourceName := "data.aws_organizations_policy.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckOrganizationManagementAccount(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, organizations.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccPoliciesForTargetDataSourceConfig_AttachQuery(rName),
				Check: resource.ComposeTestCheckFunc(
					acctest.CheckResourceAttrGreaterThanValue("data.aws_organizations_policies_for_target.test", "policies.#", "0"),
					resource.TestCheckResourceAttrPair(dataSourceName, "policies.0.arn", policyResourceName, "arn"),
					resource.TestCheckResourceAttrPair(dataSourceName, "policies.0.id", policyResourceName, "id"),
					resource.TestCheckResourceAttrPair(dataSourceName, "policies.0.name", policyResourceName, "name"),
					resource.TestCheckResourceAttrPair(dataSourceName, "policies.0.description", policyResourceName, "description"),
					resource.TestCheckResourceAttrPair(dataSourceName, "policies.0.type", policyResourceName, "type"),
				),
			},
		},
	})
}

func testAccPoliciesForTargetDataSourceConfig_AttachQuery(rName string) string {
	return fmt.Sprintf(`
data "aws_organizations_organization" "test" {
}

resource "aws_organizations_organizational_unit" "test" {
  name      = %[1]q
  parent_id = data.aws_organizations_organization.test.roots[0].id
}

resource "aws_organizations_policy" "test" {
  depends_on = [data.aws_organizations_organization.test]

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
  policy_id = data.aws_organizations_policies_for_target.test.policies[0].id
}
`, rName)
}
