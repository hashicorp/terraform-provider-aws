package aws

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccDataSourceAwsResourceGroupsTaggingAPIResources_basic(t *testing.T) {
	dataSourceName := "data.aws_resourcegroupstaggingapi_resources.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
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
	resourceName := "aws_lambda_function.test"
	rName := acctest.RandomWithPrefix("tf-acc-test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceAwsResourceGroupsTaggingAPIResourcesTagKeyFilterConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(dataSourceName, "resource_tag_mapping_list.#", "1"),
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
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
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
	resourceName := "aws_lambda_function.test"
	rName := acctest.RandomWithPrefix("tf-acc-test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceAwsResourceGroupsTaggingAPIResourcesResourceTypeFiltersConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(dataSourceName, "resource_tag_mapping_list.#", "1"),
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
	return testAccDataSourceAWSLambdaFunctionConfigBase(rName) + fmt.Sprintf(`
resource "aws_lambda_function" "test" {
  filename      = "test-fixtures/lambdatest.zip"
  function_name = %[1]q
  handler       = "exports.example"
  role          = aws_iam_role.lambda.arn
  runtime       = "nodejs12.x"

  tags = {
    Key = %[1]q
  }
}

data "aws_resourcegroupstaggingapi_resources" "test" {
  tag_filters {
    key = "Key"
  }

  depends_on = [aws_lambda_function.test]
}
`, rName)
}

func testAccDataSourceAwsResourceGroupsTaggingAPIResourcesResourceTypeFiltersConfig(rName string) string {
	return testAccDataSourceAWSLambdaFunctionConfigBase(rName) + fmt.Sprintf(`
resource "aws_lambda_function" "test" {
  filename      = "test-fixtures/lambdatest.zip"
  function_name = %[1]q
  handler       = "exports.example"
  role          = aws_iam_role.lambda.arn
  runtime       = "nodejs12.x"
}

data "aws_resourcegroupstaggingapi_resources" "test" {
  resource_type_filters = ["lambda"]

  depends_on = [aws_lambda_function.test]
}
`, rName)
}

const testAccDataSourceAwsResourceGroupsTaggingAPIResourcesComplianceConfig = `
data "aws_resourcegroupstaggingapi_resources" "test" {
  include_compliance_details  = true
  exclude_compliant_resources = false
}
`

const testAccDataSourceAwsResourceGroupsTaggingAPIResourcesResourceTypeConfig = `
data "aws_resourcegroupstaggingapi_resources" "test" {
  resource_type_filters = ["ec2:instance"]
}
`
