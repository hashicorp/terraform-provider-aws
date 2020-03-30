package aws

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
)

func TestAccDataSourceAwsResourceGroupsTaggingApiResources_basic(t *testing.T) {
	dataSourceName := "data.aws_resourcegroupstaggingapi_resources.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceAwsResourceGroupsTaggingApiResourcesBasicConfig,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(dataSourceName, "resource_tag_mapping_list.#"),
				),
			},
		},
	})
}

func TestAccDataSourceAwsResourceGroupsTaggingApiResources_tag_key_filter(t *testing.T) {
	dataSourceName := "data.aws_resourcegroupstaggingapi_resources.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceAwsResourceGroupsTaggingApiResourcesTagKeyFilterConfig,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(dataSourceName, "resource_tag_mapping_list.#"),
				),
			},
		},
	})
}

const testAccDataSourceAwsResourceGroupsTaggingApiResourcesBasicConfig = `
data "aws_resourcegroupstaggingapi_resources" "test" {}
`

const testAccDataSourceAwsResourceGroupsTaggingApiResourcesTagKeyFilterConfig = `
data "aws_resourcegroupstaggingapi_resources" "test" {
  tag_filters {
    key = "Name"
  }

}
`
