package aws

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/helper/acctest"
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
					resource.TestCheckResourceAttrSet(dataSourceName, "resource_tag_mapping_list.0.resource_arn"),
				),
			},
		},
	})
}

func TestAccDataSourceAwsResourceGroupsTaggingApiResources_tag_key_filter(t *testing.T) {
	dataSourceName := "data.aws_resourcegroupstaggingapi_resources.test"
	resourceName := "aws_vpc.test"
	key := acctest.RandString(10)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceAwsResourceGroupsTaggingApiResourcesTagKeyFilterConfig(key),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(dataSourceName, "resource_tag_mapping_list.#", "1"),
					resource.TestCheckResourceAttr(dataSourceName, fmt.Sprintf("resource_tag_mapping_list.0.tags.%s", key), key),
					resource.TestCheckResourceAttrPair(dataSourceName, "resource_tag_mapping_list.0.resource_arn", resourceName, "arn"),
				),
			},
		},
	})
}

func TestAccDataSourceAwsResourceGroupsTaggingApiResources_compliance(t *testing.T) {
	dataSourceName := "data.aws_resourcegroupstaggingapi_resources.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceAwsResourceGroupsTaggingApiResourcesComplianceConfig,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(dataSourceName, "resource_tag_mapping_list.0.compliance_details.#", "1"),
					resource.TestCheckResourceAttr(dataSourceName, "resource_tag_mapping_list.0.compliance_details.0.compliance_status", "true"),
					resource.TestCheckResourceAttrSet(dataSourceName, "resource_tag_mapping_list.0.resource_arn"),
				),
			},
		},
	})
}

func TestAccDataSourceAwsResourceGroupsTaggingApiResources_resource_type_filter(t *testing.T) {
	dataSourceName := "data.aws_resourcegroupstaggingapi_resources.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceAwsResourceGroupsTaggingApiResourcesResourceTypeConfig,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(dataSourceName, "resource_tag_mapping_list.#"),
					resource.TestCheckResourceAttrSet(dataSourceName, "resource_tag_mapping_list.0.resource_arn"),
				),
			},
		},
	})
}

const testAccDataSourceAwsResourceGroupsTaggingApiResourcesBasicConfig = `
data "aws_resourcegroupstaggingapi_resources" "test" {}
`

func testAccDataSourceAwsResourceGroupsTaggingApiResourcesTagKeyFilterConfig(key string) string {
	return fmt.Sprintf(`
resource "aws_vpc" "test" {
  cidr_block = "10.1.0.0/16"

  tags = {
    %[1]q = %[1]q
  }
}

data "aws_resourcegroupstaggingapi_resources" "test" {
  tag_filters {
    key = "${aws_vpc.test.tags[%[1]q]}"
  }
}
`, key)
}

const testAccDataSourceAwsResourceGroupsTaggingApiResourcesComplianceConfig = `
data "aws_resourcegroupstaggingapi_resources" "test" {
  include_compliance_details  = true
  exclude_compliant_resources = false
}
`

const testAccDataSourceAwsResourceGroupsTaggingApiResourcesResourceTypeConfig = `
data "aws_resourcegroupstaggingapi_resources" "test" {
  resource_type_filter = ["ec2:instance"]
}
`
