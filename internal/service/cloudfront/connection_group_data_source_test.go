// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package cloudfront_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccCloudFrontConnectionGroupDataSource_basic(t *testing.T) {
	ctx := acctest.Context(t)
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	dataSourceName := "data.aws_cloudfront_connection_group.test"
	resourceName := "aws_cloudfront_connection_group.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.CloudFrontEndpointID)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.CloudFrontServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccConnectionGroupDataSourceConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrPair(dataSourceName, "anycast_ip_list_id", resourceName, "anycast_ip_list_id"),
					resource.TestCheckResourceAttrPair(dataSourceName, names.AttrARN, resourceName, names.AttrARN),
					resource.TestCheckResourceAttrPair(dataSourceName, names.AttrEnabled, resourceName, names.AttrEnabled),
					resource.TestCheckResourceAttrPair(dataSourceName, "etag", resourceName, "etag"),
					resource.TestCheckResourceAttrPair(dataSourceName, names.AttrID, resourceName, names.AttrID),
					resource.TestCheckResourceAttrPair(dataSourceName, "ipv6_enabled", resourceName, "ipv6_enabled"),
					resource.TestCheckResourceAttrPair(dataSourceName, "is_default", resourceName, "is_default"),
					resource.TestCheckResourceAttrPair(dataSourceName, "last_modified_time", resourceName, "last_modified_time"),
					resource.TestCheckResourceAttrPair(dataSourceName, names.AttrName, resourceName, names.AttrName),
					resource.TestCheckResourceAttrPair(dataSourceName, "routing_endpoint", resourceName, "routing_endpoint"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateIdFunc: acctest.AttrImportStateIdFunc(resourceName, names.AttrName),
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"wait_for_deployment",
					names.AttrStatus,
				},
			},
		},
	})
}

func TestAccCloudFrontConnectionGroupDataSource_byRoutingEndpoint(t *testing.T) {
	ctx := acctest.Context(t)
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	dataSourceName := "data.aws_cloudfront_connection_group.test"
	resourceName := "aws_cloudfront_connection_group.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckPartitionHasService(t, names.CloudFrontEndpointID) },
		ErrorCheck:               acctest.ErrorCheck(t, names.CloudFrontServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccConnectionGroupDataSourceConfig_byRoutingEndpoint(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrPair(dataSourceName, "anycast_ip_list_id", resourceName, "anycast_ip_list_id"),
					resource.TestCheckResourceAttrPair(dataSourceName, names.AttrARN, resourceName, names.AttrARN),
					resource.TestCheckResourceAttrPair(dataSourceName, names.AttrEnabled, resourceName, names.AttrEnabled),
					resource.TestCheckResourceAttrPair(dataSourceName, "etag", resourceName, "etag"),
					resource.TestCheckResourceAttrPair(dataSourceName, names.AttrID, resourceName, names.AttrID),
					resource.TestCheckResourceAttrPair(dataSourceName, "ipv6_enabled", resourceName, "ipv6_enabled"),
					resource.TestCheckResourceAttrPair(dataSourceName, "is_default", resourceName, "is_default"),
					resource.TestCheckResourceAttrPair(dataSourceName, "last_modified_time", resourceName, "last_modified_time"),
					resource.TestCheckResourceAttrPair(dataSourceName, names.AttrName, resourceName, names.AttrName),
					resource.TestCheckResourceAttrPair(dataSourceName, "routing_endpoint", resourceName, "routing_endpoint"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateIdFunc: acctest.AttrImportStateIdFunc(resourceName, names.AttrName),
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"wait_for_deployment",
					names.AttrStatus,
				},
			},
		},
	})
}

func testAccConnectionGroupDataSourceConfig_basic(rName string) string {
	return fmt.Sprintf(`
resource "aws_cloudfront_connection_group" "test" {
  name    = %[1]q
  enabled = true
}

data "aws_cloudfront_connection_group" "test" {
  id = aws_cloudfront_connection_group.test.id
}
`, rName)
}

func testAccConnectionGroupDataSourceConfig_byRoutingEndpoint(rName string) string {
	return fmt.Sprintf(`
resource "aws_cloudfront_connection_group" "test" {
  name    = %[1]q
  enabled = true
}

data "aws_cloudfront_connection_group" "test" {
  routing_endpoint = aws_cloudfront_connection_group.test.routing_endpoint
}
`, rName)
}
