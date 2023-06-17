package organizations_test

import (
	"testing"

	"github.com/aws/aws-sdk-go/service/organizations"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
)

func testAccOrganizationDataSource_basic(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_organizations_organization.test"
	dataSourceName := "data.aws_organizations_organization.test"

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckOrganizationsAccount(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, organizations.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccOrganizationDataSourceConfig_basic,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrPair(resourceName, "accounts.#", dataSourceName, "accounts.#"),
					resource.TestCheckResourceAttrPair(resourceName, "arn", dataSourceName, "arn"),
					resource.TestCheckResourceAttrPair(resourceName, "aws_service_access_principals.#", dataSourceName, "aws_service_access_principals.#"),
					resource.TestCheckResourceAttrPair(resourceName, "enabled_policy_types.#", dataSourceName, "enabled_policy_types.#"),
					resource.TestCheckResourceAttrPair(resourceName, "feature_set", dataSourceName, "feature_set"),
					resource.TestCheckResourceAttrPair(resourceName, "id", dataSourceName, "id"),
					resource.TestCheckResourceAttrPair(resourceName, "status", dataSourceName, "status"),
					resource.TestCheckResourceAttrPair(resourceName, "master_account_arn", dataSourceName, "master_account_arn"),
					resource.TestCheckResourceAttrPair(resourceName, "master_account_email", dataSourceName, "master_account_email"),
					resource.TestCheckResourceAttrPair(resourceName, "master_account_id", dataSourceName, "master_account_id"),
					resource.TestCheckResourceAttrPair(resourceName, "non_master_accounts.#", dataSourceName, "non_master_accounts.#"),
					resource.TestCheckResourceAttrPair(resourceName, "roots.#", dataSourceName, "roots.#"),
				),
			},
		},
	})
}

// Create an new organization so that we are its management account.
const testAccOrganizationDataSourceConfig_basic = `
resource "aws_organizations_organization" "test" {}

data "aws_organizations_organization" "test" {
  depends_on = [aws_organizations_organization.test]
}
`
