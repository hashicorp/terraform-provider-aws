package organizations_test

import (
	"testing"

	"github.com/aws/aws-sdk-go/service/organizations"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
)

func TestAccOrganizationPolicyDataSource_basic(t *testing.T) {
	ctx := acctest.Context(t)
	dataSourceName := "data.aws_organizations_policy.test"

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckOrganizationManagementAccount(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, organizations.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccOrganizationPolicyDataSourceConfig_basic,
				Check: resource.ComposeTestCheckFunc(
					acctest.CheckResourceAttrGreaterThanValue(dataSourceName, "policy.#", "0"),
				),
			},
		},
	})
}

const testAccOrganizationPolicyDataSourceConfig_basic = `
data "aws_organizations_organization" "current" {}

data "aws_organizations_organizational_policies" "current" {
  target_id = data.aws_organizations_organization.current.roots[0].id
  filter = "SERVICE_CONTROL_POLICY"
}
data "aws_organizations_policies" "test" {
	policy_id= data.aws_organizations_oorganizational_policies.current.policies[0].id
}
`
