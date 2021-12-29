package organizations_test

import (
	"testing"

	"github.com/aws/aws-sdk-go/service/organizations"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
)

func testAccOrganizationResourceTagsDataSource_basic(t *testing.T) {
	resourceName := "aws_organizations_resource_tag.test"
	dataSourceName := "data.aws_organizations_resource_tags.test"
	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(t)
			acctest.PreCheckOrganizationsAccount(t)
		},
		ErrorCheck: acctest.ErrorCheck(t, organizations.EndpointsID),
		Providers:  acctest.Providers,
		Steps: []resource.TestStep{
			{
				Config: testAccOrganizationResourceTagsDataSourceConfig,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(dataSourceName, "tags.#", "2"),
					resource.TestCheckResourceAttrPair(resourceName, "id", dataSourceName, "children.0.id"),
				),
			},
			{
				// This is to make sure the data source isn't around trying to read the resource
				// when the resource is being destroyed
				Config: testAccCheckAWSOrganizationResourceOnlyConfig,
			},
		},
	})
}

const testAccOrganizationResourceTagsDataSourceConfig = `
resource "aws_organizations_resource_tags" "test" {
  "id" = "12345"
  "resource_id" = "12345"
  "tags" = tomap({
    "testtagkey1" = "testtagval1"
    "testtagkey2" = "testtagval2"
  })
}

data "aws_organizations_resource_tags" "test" {
  resource_id = aws_organizations_resource_tags.test.resource_id
}
`
