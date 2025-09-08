// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package cloudfront_test

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccCloudFrontConnectionGroupDataSource_basic(t *testing.T) {
	ctx := acctest.Context(t)
	dataSourceName := "data.aws_cloudfront_connection_group.test"
	resourceName := "aws_cloudfront_connection_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckPartitionHasService(t, names.CloudFrontEndpointID) },
		ErrorCheck:               acctest.ErrorCheck(t, names.CloudFrontServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccConnectionGroupDataSourceConfig_basic(),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrPair(dataSourceName, "anycast_ip_list_id", resourceName, "anycast_ip_list_id"),
					resource.TestCheckResourceAttrPair(dataSourceName, names.AttrARN, resourceName, names.AttrARN),
					resource.TestCheckResourceAttrPair(dataSourceName, names.AttrEnabled, resourceName, names.AttrEnabled),
					resource.TestCheckResourceAttrPair(dataSourceName, "etag", resourceName, "etag"),
					resource.TestCheckResourceAttrPair(dataSourceName, "ipv6_enabled", resourceName, "ipv6_enabled"),
					resource.TestCheckResourceAttrPair(dataSourceName, "is_default", resourceName, "is_default"),
					resource.TestCheckResourceAttrPair(dataSourceName, "last_modified_time", resourceName, "last_modified_time"),
					resource.TestCheckResourceAttrPair(dataSourceName, names.AttrName, resourceName, names.AttrName),
					resource.TestCheckResourceAttrPair(dataSourceName, "routing_endpoint", resourceName, "routing_endpoint"),
					resource.TestCheckResourceAttrPair(dataSourceName, names.AttrStatus, resourceName, names.AttrStatus),
				),
			},
		},
	})
}

func TestAccCloudFrontConnectionGroupDataSource_anycast(t *testing.T) {
	ctx := acctest.Context(t)
	dataSourceName := "data.aws_cloudfront_connection_group.testip"
	resourceName := "aws_cloudfront_connection_group.testip"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckPartitionHasService(t, names.CloudFrontEndpointID) },
		ErrorCheck:               acctest.ErrorCheck(t, names.CloudFrontServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccConnectionGroupDataSourceConfig_anycast(),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrPair(dataSourceName, "anycast_ip_list_id", resourceName, "anycast_ip_list_id"),
					resource.TestCheckResourceAttrPair(dataSourceName, names.AttrARN, resourceName, names.AttrARN),
					resource.TestCheckResourceAttrPair(dataSourceName, names.AttrEnabled, resourceName, names.AttrEnabled),
					resource.TestCheckResourceAttrPair(dataSourceName, "etag", resourceName, "etag"),
					resource.TestCheckResourceAttrPair(dataSourceName, "ipv6_enabled", resourceName, "ipv6_enabled"),
					resource.TestCheckResourceAttrPair(dataSourceName, "is_default", resourceName, "is_default"),
					resource.TestCheckResourceAttrPair(dataSourceName, "last_modified_time", resourceName, "last_modified_time"),
					resource.TestCheckResourceAttrPair(dataSourceName, names.AttrName, resourceName, names.AttrName),
					resource.TestCheckResourceAttrPair(dataSourceName, "routing_endpoint", resourceName, "routing_endpoint"),
					resource.TestCheckResourceAttrPair(dataSourceName, names.AttrStatus, resourceName, names.AttrStatus),
				),
			},
		},
	})
}

func testAccConnectionGroupDataSourceConfig_basic() string {
	return acctest.ConfigCompose(testAccConnectionGroupConfig_basic(), `
data "aws_cloudfront_connection_group" "test" {
  id = aws_cloudfront_connection_group.test.id
}
`)
}

func testAccConnectionGroupDataSourceConfig_anycast() string {
	return acctest.ConfigCompose(testAccConnectionGroupConfig_anycastIpList(), `
data "aws_cloudfront_connection_group" "testip" {
  id = aws_cloudfront_connection_group.testip.id
}
`)
}
