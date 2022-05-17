package organizations_test

import (
	"testing"

	"github.com/aws/aws-sdk-go/service/organizations"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
)

func testAccOrganizationDataSource_basic(t *testing.T) {
	resourceName := "aws_organizations_organization.test"
	dataSourceName := "data.aws_organizations_organization.test"

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(t)
			acctest.PreCheckOrganizationsAccount(t)
		},
		ErrorCheck:        acctest.ErrorCheck(t, organizations.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccOrganizationConfig_resourceOnly,
			},
			{
				Config: testAccOrganizationConfig_basic,
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
			{
				// This is to make sure the data source isn't around trying to read the resource
				// when the resource is being destroyed
				Config: testAccOrganizationConfig_resourceOnly,
			},
		},
	})
}

const testAccOrganizationConfig_resourceOnly = `
resource "aws_organizations_organization" "test" {}
`

const testAccOrganizationConfig_basic = `
resource "aws_organizations_organization" "test" {}

data "aws_organizations_organization" "test" {}
`
