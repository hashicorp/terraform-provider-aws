package aws

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/service/resourcegroupstaggingapi"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccDataSourceAwsResourceGroupsTaggingAPIResources_basic(t *testing.T) {
	dataSourceName := "data.aws_resourcegroupstaggingapi_resources.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:   func() { testAccPreCheck(t) },
		ErrorCheck: testAccErrorCheck(t, resourcegroupstaggingapi.EndpointsID),
		Providers:  testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceAwsResourceGroupsTaggingAPIResourcesBasicConfig,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(dataSourceName, "resource_tag_mapping_list.#"),
					resource.TestCheckResourceAttrSet(dataSourceName, "resource_tag_mapping_list.0.resource_arn"),
				),
			},
		},
	})
}

func TestAccDataSourceAwsResourceGroupsTaggingAPIResources_tag_key_filter(t *testing.T) {
	dataSourceName := "data.aws_resourcegroupstaggingapi_resources.test"
	resourceName := "aws_api_gateway_rest_api.test"
	rName := acctest.RandomWithPrefix("tf-acc-test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:   func() { testAccPreCheck(t) },
		ErrorCheck: testAccErrorCheck(t, resourcegroupstaggingapi.EndpointsID),
		Providers:  testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceAwsResourceGroupsTaggingAPIResourcesTagKeyFilterConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckTypeSetElemNestedAttrs(dataSourceName, "resource_tag_mapping_list.*", map[string]string{
						"tags.Key": rName,
					}),
					resource.TestCheckTypeSetElemAttrPair(dataSourceName, "resource_tag_mapping_list.*.resource_arn", resourceName, "arn"),
				),
			},
		},
	})
}

func TestAccDataSourceAwsResourceGroupsTaggingAPIResources_compliance(t *testing.T) {
	dataSourceName := "data.aws_resourcegroupstaggingapi_resources.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:   func() { testAccPreCheck(t) },
		ErrorCheck: testAccErrorCheck(t, resourcegroupstaggingapi.EndpointsID),
		Providers:  testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceAwsResourceGroupsTaggingAPIResourcesComplianceConfig,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(dataSourceName, "resource_tag_mapping_list.0.compliance_details.#", "1"),
					resource.TestCheckResourceAttr(dataSourceName, "resource_tag_mapping_list.0.compliance_details.0.compliance_status", "true"),
					resource.TestCheckResourceAttrSet(dataSourceName, "resource_tag_mapping_list.0.resource_arn"),
				),
			},
		},
	})
}

func TestAccDataSourceAwsResourceGroupsTaggingAPIResources_resource_type_filters(t *testing.T) {
	dataSourceName := "data.aws_resourcegroupstaggingapi_resources.test"
	resourceName := "aws_api_gateway_rest_api.test"
	rName := acctest.RandomWithPrefix("tf-acc-test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:   func() { testAccPreCheck(t) },
		ErrorCheck: testAccErrorCheck(t, resourcegroupstaggingapi.EndpointsID),
		Providers:  testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceAwsResourceGroupsTaggingAPIResourcesResourceTypeFiltersConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckTypeSetElemNestedAttrs(dataSourceName, "resource_tag_mapping_list.*", map[string]string{
						"tags.Key": rName,
					}),
					resource.TestCheckTypeSetElemAttrPair(dataSourceName, "resource_tag_mapping_list.*.resource_arn", resourceName, "arn"),
				),
			},
		},
	})
}

func TestAccDataSourceAwsResourceGroupsTaggingAPIResources_resource_arn_list(t *testing.T) {
	dataSourceName := "data.aws_resourcegroupstaggingapi_resources.test"
	resourceName := "aws_api_gateway_rest_api.test"
	rName := acctest.RandomWithPrefix("tf-acc-test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:   func() { testAccPreCheck(t) },
		ErrorCheck: testAccErrorCheck(t, resourcegroupstaggingapi.EndpointsID),
		Providers:  testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceAwsResourceGroupsTaggingAPIResourcesResourceARNListConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckTypeSetElemNestedAttrs(dataSourceName, "resource_tag_mapping_list.*", map[string]string{
						"tags.Key": rName,
					}),
					resource.TestCheckTypeSetElemAttrPair(dataSourceName, "resource_tag_mapping_list.*.resource_arn", resourceName, "arn"),
				),
			},
		},
	})
}

const testAccDataSourceAwsResourceGroupsTaggingAPIResourcesBasicConfig = `
data "aws_resourcegroupstaggingapi_resources" "test" {}
`

func testAccDataSourceAwsResourceGroupsTaggingAPIResourcesTagKeyFilterConfig(rName string) string {
	return fmt.Sprintf(`
resource "aws_api_gateway_rest_api" "test" {
  name = %[1]q

  tags = {
    Key = %[1]q
  }
}

data "aws_resourcegroupstaggingapi_resources" "test" {
  filter {
    key = "Key"
  }

  depends_on = [aws_api_gateway_rest_api.test]
}
`, rName)
}

func testAccDataSourceAwsResourceGroupsTaggingAPIResourcesResourceTypeFiltersConfig(rName string) string {
	return fmt.Sprintf(`
resource "aws_api_gateway_rest_api" "test" {
  name = %[1]q

  tags = {
    Key = %[1]q
  }
}

data "aws_resourcegroupstaggingapi_resources" "test" {
  resource_type_filters = ["apigateway"]

  depends_on = [aws_api_gateway_rest_api.test]
}
`, rName)
}

func testAccDataSourceAwsResourceGroupsTaggingAPIResourcesResourceARNListConfig(rName string) string {
	return fmt.Sprintf(`
resource "aws_api_gateway_rest_api" "test" {
  name = %[1]q

  tags = {
    Key = %[1]q
  }
}

data "aws_resourcegroupstaggingapi_resources" "test" {
  resource_arn_list = [aws_api_gateway_rest_api.test.arn]
}
`, rName)
}

const testAccDataSourceAwsResourceGroupsTaggingAPIResourcesComplianceConfig = `
data "aws_resourcegroupstaggingapi_resources" "test" {
  include_compliance_details  = true
  exclude_compliant_resources = false
}
`
