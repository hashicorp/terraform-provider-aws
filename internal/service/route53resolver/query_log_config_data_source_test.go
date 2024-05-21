// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package route53resolver_test

import (
	"testing"

	"github.com/aws/aws-sdk-go/service/route53resolver"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/names"
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
		ErrorCheck:               acctest.ErrorCheck(t, names.Route53ResolverServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccQueryLogConfigDataSourceConfig_basic(rName, acctest.CtKey1, acctest.CtValue1),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrPair(dataSourceName, names.AttrARN, resourceName, names.AttrARN),
					resource.TestCheckResourceAttrPair(dataSourceName, names.AttrDestinationARN, resourceName, names.AttrDestinationARN),
					resource.TestCheckResourceAttrPair(dataSourceName, names.AttrID, resourceName, names.AttrID),
					resource.TestCheckResourceAttrPair(dataSourceName, names.AttrName, resourceName, names.AttrName),
					resource.TestCheckResourceAttrPair(dataSourceName, names.AttrOwnerID, resourceName, names.AttrOwnerID),
					resource.TestCheckResourceAttrPair(dataSourceName, "resolver_query_log_config_id", resourceName, names.AttrID),
					resource.TestCheckResourceAttrPair(dataSourceName, "share_status", resourceName, "share_status"),
					resource.TestCheckResourceAttrPair(dataSourceName, acctest.CtTagsPercent, resourceName, acctest.CtTagsPercent),
					resource.TestCheckResourceAttrPair(dataSourceName, acctest.CtTagsKey1, resourceName, acctest.CtTagsKey1),
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
		ErrorCheck:               acctest.ErrorCheck(t, names.Route53ResolverServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccQueryLogConfigDataSourceConfig_filter(rName, acctest.CtKey1, acctest.CtValue1),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrPair(dataSourceName, names.AttrARN, resourceName, names.AttrARN),
					resource.TestCheckResourceAttrPair(dataSourceName, names.AttrDestinationARN, resourceName, names.AttrDestinationARN),
					resource.TestCheckResourceAttrPair(dataSourceName, names.AttrID, resourceName, names.AttrID),
					resource.TestCheckResourceAttrPair(dataSourceName, names.AttrName, resourceName, names.AttrName),
					resource.TestCheckResourceAttrPair(dataSourceName, names.AttrOwnerID, resourceName, names.AttrOwnerID),
					resource.TestCheckResourceAttrPair(dataSourceName, "resolver_query_log_config_id", resourceName, names.AttrID),
					resource.TestCheckResourceAttrPair(dataSourceName, "share_status", resourceName, "share_status"),
					resource.TestCheckResourceAttrPair(dataSourceName, acctest.CtTagsPercent, resourceName, acctest.CtTagsPercent),
					resource.TestCheckResourceAttrPair(dataSourceName, acctest.CtTagsKey1, resourceName, acctest.CtTagsKey1),
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
