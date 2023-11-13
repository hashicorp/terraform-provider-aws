// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package route53resolver_test

import (
	"testing"

	"github.com/aws/aws-sdk-go/service/route53resolver"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
)

func TestAccRoute53ResolverQueryLogConfigDataSource_basic(t *testing.T) {
	ctx := acctest.Context(t)

	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_route53_resolver_query_log_config.test"
	dataSourceName := "data.aws_route53_resolver_query_log_config.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, route53resolver.EndpointsID)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, route53resolver.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccQueryLogConfigDataSourceConfig_basic(rName, "key1", "value1"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrPair(dataSourceName, "arn", resourceName, "arn"),
					resource.TestCheckResourceAttrPair(dataSourceName, "destination_arn", resourceName, "destination_arn"),
					resource.TestCheckResourceAttrPair(dataSourceName, "id", resourceName, "id"),
					resource.TestCheckResourceAttrPair(dataSourceName, "name", resourceName, "name"),
					resource.TestCheckResourceAttrPair(dataSourceName, "owner_id", resourceName, "owner_id"),
					resource.TestCheckResourceAttrPair(dataSourceName, "resolver_query_log_config_id", resourceName, "id"),
					resource.TestCheckResourceAttrPair(dataSourceName, "share_status", resourceName, "share_status"),
					resource.TestCheckResourceAttrPair(dataSourceName, "tags.%", resourceName, "tags.%"),
					resource.TestCheckResourceAttrPair(dataSourceName, "tags.key1", resourceName, "tags.key1"),
				),
			},
		},
	})
}

func TestAccRoute53ResolverQueryLogConfigDataSource_filter(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_route53_resolver_query_log_config.test"
	dataSourceName := "data.aws_route53_resolver_query_log_config.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, route53resolver.EndpointsID)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, route53resolver.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccQueryLogConfigDataSourceConfig_filter(rName, "key1", "value1"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrPair(dataSourceName, "arn", resourceName, "arn"),
					resource.TestCheckResourceAttrPair(dataSourceName, "destination_arn", resourceName, "destination_arn"),
					resource.TestCheckResourceAttrPair(dataSourceName, "id", resourceName, "id"),
					resource.TestCheckResourceAttrPair(dataSourceName, "name", resourceName, "name"),
					resource.TestCheckResourceAttrPair(dataSourceName, "owner_id", resourceName, "owner_id"),
					resource.TestCheckResourceAttrPair(dataSourceName, "resolver_query_log_config_id", resourceName, "id"),
					resource.TestCheckResourceAttrPair(dataSourceName, "share_status", resourceName, "share_status"),
					resource.TestCheckResourceAttrPair(dataSourceName, "tags.%", resourceName, "tags.%"),
					resource.TestCheckResourceAttrPair(dataSourceName, "tags.key1", resourceName, "tags.key1"),
				),
			},
		},
	})
}

func testAccQueryLogConfigDataSourceConfig_basic(rName string, tagKey string, tagValue string) string {
	return acctest.ConfigCompose(testAccQueryLogConfigConfig_tags1(rName, tagKey, tagValue), `
data "aws_route53_resolver_query_log_config" "test" {
  resolver_query_log_config_id = aws_route53_resolver_query_log_config.test.id
}
`)
}

func testAccQueryLogConfigDataSourceConfig_filter(rName string, tagKey string, tagValue string) string {
	return acctest.ConfigCompose(testAccQueryLogConfigConfig_tags1(rName, tagKey, tagValue), `
data "aws_route53_resolver_query_log_config" "test" {
  filter {
    name   = "Name"
    values = [aws_route53_resolver_query_log_config.test.name]
  }
}
`)
}
