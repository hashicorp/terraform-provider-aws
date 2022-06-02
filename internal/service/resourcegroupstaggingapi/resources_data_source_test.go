package resourcegroupstaggingapi_test

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/service/resourcegroupstaggingapi"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
)

func TestAccResourceGroupsTaggingAPIResourcesDataSource_tagFilter(t *testing.T) {
	dataSourceName := "data.aws_resourcegroupstaggingapi_resources.test"
	resourceName := "aws_vpc.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, resourcegroupstaggingapi.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccResourcesTagFilterDataSourceConfig(rName),
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

func TestAccResourceGroupsTaggingAPIResourcesDataSource_includeComplianceDetails(t *testing.T) {
	dataSourceName := "data.aws_resourcegroupstaggingapi_resources.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, resourcegroupstaggingapi.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccResourcesIncludeComplianceDetailsDataSourceConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(dataSourceName, "resource_tag_mapping_list.0.compliance_details.#", "1"),
					resource.TestCheckResourceAttr(dataSourceName, "resource_tag_mapping_list.0.compliance_details.0.compliance_status", "true"),
				),
			},
		},
	})
}

func TestAccResourceGroupsTaggingAPIResourcesDataSource_resourceTypeFilters(t *testing.T) {
	dataSourceName := "data.aws_resourcegroupstaggingapi_resources.test"
	resourceName := "aws_vpc.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, resourcegroupstaggingapi.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccResourcesResourceTypeFiltersDataSourceConfig(rName),
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

func TestAccResourceGroupsTaggingAPIResourcesDataSource_resourceARNList(t *testing.T) {
	dataSourceName := "data.aws_resourcegroupstaggingapi_resources.test"
	resourceName := "aws_vpc.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, resourcegroupstaggingapi.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccResourcesResourceARNListDataSourceConfig(rName),
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

func testAccResourcesTagFilterDataSourceConfig(rName string) string {
	return fmt.Sprintf(`
resource "aws_vpc" "test" {
  cidr_block = "10.0.0.0/16"

  tags = {
    Key = %[1]q
  }
}

data "aws_resourcegroupstaggingapi_resources" "test" {
  tag_filter {
    key    = "Key"
    values = [aws_vpc.test.tags["Key"]]
  }
}
`, rName)
}

func testAccResourcesResourceTypeFiltersDataSourceConfig(rName string) string {
	return fmt.Sprintf(`
resource "aws_vpc" "test" {
  cidr_block = "10.0.0.0/16"

  tags = {
    Key = %[1]q
  }
}

data "aws_resourcegroupstaggingapi_resources" "test" {
  resource_type_filters = ["ec2:vpc"]

  depends_on = [aws_vpc.test]
}
`, rName)
}

func testAccResourcesResourceARNListDataSourceConfig(rName string) string {
	return fmt.Sprintf(`
resource "aws_vpc" "test" {
  cidr_block = "10.0.0.0/16"

  tags = {
    Key = %[1]q
  }
}

data "aws_resourcegroupstaggingapi_resources" "test" {
  resource_arn_list = [aws_vpc.test.arn]
}
`, rName)
}

func testAccResourcesIncludeComplianceDetailsDataSourceConfig(rName string) string {
	return fmt.Sprintf(`
resource "aws_vpc" "test" {
  cidr_block = "10.0.0.0/16"

  tags = {
    Key = %[1]q
  }
}

data "aws_resourcegroupstaggingapi_resources" "test" {
  include_compliance_details  = true
  exclude_compliant_resources = false
  resource_arn_list           = [aws_vpc.test.arn]
}
`, rName)
}
