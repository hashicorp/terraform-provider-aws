// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package resourcegroupstaggingapi_test

import (
	"fmt"
	"testing"

	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccResourceGroupsTaggingAPIResourcesDataSource_tagFilter(t *testing.T) {
	ctx := acctest.Context(t)
	dataSourceName := "data.aws_resourcegroupstaggingapi_resources.test"
	resourceName := "aws_vpc.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ResourceGroupsTaggingAPIServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccResourcesDataSourceConfig_tagFilter(rName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckTypeSetElemNestedAttrs(dataSourceName, "resource_tag_mapping_list.*", map[string]string{
						"tags.Key": rName,
					}),
					resource.TestCheckTypeSetElemAttrPair(dataSourceName, "resource_tag_mapping_list.*.resource_arn", resourceName, names.AttrARN),
				),
			},
		},
	})
}

func TestAccResourceGroupsTaggingAPIResourcesDataSource_includeComplianceDetails(t *testing.T) {
	ctx := acctest.Context(t)
	dataSourceName := "data.aws_resourcegroupstaggingapi_resources.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ResourceGroupsTaggingAPIServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccResourcesDataSourceConfig_includeComplianceDetails(rName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(dataSourceName, "resource_tag_mapping_list.0.compliance_details.#", acctest.Ct1),
					resource.TestCheckResourceAttr(dataSourceName, "resource_tag_mapping_list.0.compliance_details.0.compliance_status", acctest.CtTrue),
				),
			},
		},
	})
}

func TestAccResourceGroupsTaggingAPIResourcesDataSource_resourceTypeFilters(t *testing.T) {
	ctx := acctest.Context(t)
	dataSourceName := "data.aws_resourcegroupstaggingapi_resources.test"
	resourceName := "aws_vpc.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ResourceGroupsTaggingAPIServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccResourcesDataSourceConfig_resourceTypeFilters(rName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckTypeSetElemNestedAttrs(dataSourceName, "resource_tag_mapping_list.*", map[string]string{
						"tags.Key": rName,
					}),
					resource.TestCheckTypeSetElemAttrPair(dataSourceName, "resource_tag_mapping_list.*.resource_arn", resourceName, names.AttrARN),
				),
			},
		},
	})
}

func TestAccResourceGroupsTaggingAPIResourcesDataSource_resourceARNList(t *testing.T) {
	ctx := acctest.Context(t)
	dataSourceName := "data.aws_resourcegroupstaggingapi_resources.test"
	resourceName := "aws_vpc.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ResourceGroupsTaggingAPIServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccResourcesDataSourceConfig_resourceARNList(rName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckTypeSetElemNestedAttrs(dataSourceName, "resource_tag_mapping_list.*", map[string]string{
						"tags.Key": rName,
					}),
					resource.TestCheckTypeSetElemAttrPair(dataSourceName, "resource_tag_mapping_list.*.resource_arn", resourceName, names.AttrARN),
				),
			},
		},
	})
}

func testAccResourcesDataSourceConfig_tagFilter(rName string) string {
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

func testAccResourcesDataSourceConfig_resourceTypeFilters(rName string) string {
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

func testAccResourcesDataSourceConfig_resourceARNList(rName string) string {
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

func testAccResourcesDataSourceConfig_includeComplianceDetails(rName string) string {
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
