// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package cloudfront_test

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccCloudFrontDistributionTenantDataSource_basic(t *testing.T) {
	ctx := acctest.Context(t)
	dataSourceName := "data.aws_cloudfront_distribution_tenant.test"
	resourceName := "aws_cloudfront_distribution_tenant.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckPartitionHasService(t, names.CloudFrontEndpointID) },
		ErrorCheck:               acctest.ErrorCheck(t, names.CloudFrontServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccDistributionTenantDataSourceConfig_basic(t),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrPair(dataSourceName, names.AttrARN, resourceName, names.AttrARN),
					resource.TestCheckResourceAttrPair(dataSourceName, "connection_group_id", resourceName, "connection_group_id"),
					resource.TestCheckResourceAttrPair(dataSourceName, "distribution_id", resourceName, "distribution_id"),
					resource.TestCheckResourceAttrPair(dataSourceName, "domains.#", resourceName, "domains.#"),
					resource.TestCheckResourceAttrPair(dataSourceName, names.AttrEnabled, resourceName, names.AttrEnabled),
					resource.TestCheckResourceAttrPair(dataSourceName, "etag", resourceName, "etag"),
					resource.TestCheckResourceAttrPair(dataSourceName, "last_modified_time", resourceName, "last_modified_time"),
					resource.TestCheckResourceAttrPair(dataSourceName, names.AttrName, resourceName, names.AttrName),
					resource.TestCheckResourceAttrPair(dataSourceName, names.AttrStatus, resourceName, names.AttrStatus),
				),
			},
		},
	})
}

func testAccDistributionTenantDataSourceConfig_basic(t *testing.T) string {
	return acctest.ConfigCompose(testAccDistributionTenantConfig_basic(t), `
data "aws_cloudfront_distribution_tenant" "test" {
  id = aws_cloudfront_distribution_tenant.test.id
}
`)
}
