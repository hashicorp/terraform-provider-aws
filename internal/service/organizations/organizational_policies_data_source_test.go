package organizations_test

import (
	"testing"

	"github.com/aws/aws-sdk-go/service/organizations"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
)

func testAccOrganizationalPoliciesDataSource_basic(t *testing.T) {
	ctx := acctest.Context(t)
	organizationPolicyDataSourceName := "data.aws_organizations_organizational_policies.test"

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckOrganizationManagementAccount(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, organizations.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccOrganizationalPoliciesDataSourceConfig_basic,
				Check: resource.ComposeTestCheckFunc(
					acctest.CheckResourceAttrGreaterThanValue(organizationPolicyDataSourceName, "policies.#", "0"),
				),
			},
		},
	})
}

const testAccOrganizationalPoliciesDataSourceConfig_basic = `
data "aws_organizations_organization" "current" {}

resource "aws_organizations_organizational_unit" "test" {
  name      = "test"
  target_id = data.aws_organizations_organization.current.roots[0].id
}
`
