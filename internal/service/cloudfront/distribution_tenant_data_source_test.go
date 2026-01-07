// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package cloudfront_test

import (
	"testing"

	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccCloudFrontDistributionTenantDataSource_basic(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	rootDomain := acctest.ACMCertificateDomainFromEnv(t)
	domain := acctest.ACMCertificateRandomSubDomain(rootDomain)
	dataSourceName := "data.aws_cloudfront_distribution_tenant.test"
	resourceName := "aws_cloudfront_distribution_tenant.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.CloudFrontEndpointID)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.CloudFrontServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccDistributionTenantDataSourceConfig_basic(rName, rootDomain, domain),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrPair(dataSourceName, names.AttrARN, resourceName, names.AttrARN),
					resource.TestCheckResourceAttrPair(dataSourceName, "connection_group_id", resourceName, "connection_group_id"),
					resource.TestCheckResourceAttrPair(dataSourceName, "customizations.#", resourceName, "customizations.#"),
					resource.TestCheckResourceAttrPair(dataSourceName, "distribution_id", resourceName, "distribution_id"),
					resource.TestCheckResourceAttrPair(dataSourceName, "domains.#", resourceName, "domain.#"),
					resource.TestCheckResourceAttrPair(dataSourceName, names.AttrEnabled, resourceName, names.AttrEnabled),
					resource.TestCheckResourceAttrPair(dataSourceName, "etag", resourceName, "etag"),
					resource.TestCheckResourceAttrPair(dataSourceName, "managed_certificate_request.#", resourceName, "managed_certificate_request.#"),
					resource.TestCheckResourceAttrPair(dataSourceName, names.AttrName, resourceName, names.AttrName),
					resource.TestCheckResourceAttrPair(dataSourceName, "parameters.#", resourceName, "parameter.#"),
					resource.TestCheckResourceAttrSet(dataSourceName, names.AttrStatus),
				),
			},
		},
	})
}

func TestAccCloudFrontDistributionTenantDataSource_byARN(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	rootDomain := acctest.ACMCertificateDomainFromEnv(t)
	domain := acctest.ACMCertificateRandomSubDomain(rootDomain)
	dataSourceName := "data.aws_cloudfront_distribution_tenant.test"
	resourceName := "aws_cloudfront_distribution_tenant.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.CloudFrontEndpointID)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.CloudFrontServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccDistributionTenantDataSourceConfig_byARN(rName, rootDomain, domain),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrPair(dataSourceName, names.AttrARN, resourceName, names.AttrARN),
					resource.TestCheckResourceAttrPair(dataSourceName, "connection_group_id", resourceName, "connection_group_id"),
					resource.TestCheckResourceAttrPair(dataSourceName, "customizations.#", resourceName, "customizations.#"),
					resource.TestCheckResourceAttrPair(dataSourceName, "distribution_id", resourceName, "distribution_id"),
					resource.TestCheckResourceAttrPair(dataSourceName, "domains.#", resourceName, "domain.#"),
					resource.TestCheckResourceAttrPair(dataSourceName, names.AttrEnabled, resourceName, names.AttrEnabled),
					resource.TestCheckResourceAttrPair(dataSourceName, "etag", resourceName, "etag"),
					resource.TestCheckResourceAttrPair(dataSourceName, "managed_certificate_request.#", resourceName, "managed_certificate_request.#"),
					resource.TestCheckResourceAttrPair(dataSourceName, names.AttrName, resourceName, names.AttrName),
					resource.TestCheckResourceAttrPair(dataSourceName, "parameters.#", resourceName, "parameter.#"),
					resource.TestCheckResourceAttrSet(dataSourceName, names.AttrStatus),
				),
			},
		},
	})
}

func TestAccCloudFrontDistributionTenantDataSource_byName(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	rootDomain := acctest.ACMCertificateDomainFromEnv(t)
	domain := acctest.ACMCertificateRandomSubDomain(rootDomain)
	dataSourceName := "data.aws_cloudfront_distribution_tenant.test"
	resourceName := "aws_cloudfront_distribution_tenant.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.CloudFrontEndpointID)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.CloudFrontServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccDistributionTenantDataSourceConfig_byName(rName, rootDomain, domain),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrPair(dataSourceName, names.AttrARN, resourceName, names.AttrARN),
					resource.TestCheckResourceAttrPair(dataSourceName, "connection_group_id", resourceName, "connection_group_id"),
					resource.TestCheckResourceAttrPair(dataSourceName, "customizations.#", resourceName, "customizations.#"),
					resource.TestCheckResourceAttrPair(dataSourceName, "distribution_id", resourceName, "distribution_id"),
					resource.TestCheckResourceAttrPair(dataSourceName, "domains.#", resourceName, "domain.#"),
					resource.TestCheckResourceAttrPair(dataSourceName, names.AttrEnabled, resourceName, names.AttrEnabled),
					resource.TestCheckResourceAttrPair(dataSourceName, "etag", resourceName, "etag"),
					resource.TestCheckResourceAttrPair(dataSourceName, "managed_certificate_request.#", resourceName, "managed_certificate_request.#"),
					resource.TestCheckResourceAttrPair(dataSourceName, names.AttrName, resourceName, names.AttrName),
					resource.TestCheckResourceAttrPair(dataSourceName, "parameters.#", resourceName, "parameter.#"),
					resource.TestCheckResourceAttrSet(dataSourceName, names.AttrStatus),
				),
			},
		},
	})
}

func TestAccCloudFrontDistributionTenantDataSource_byDomain(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	rootDomain := acctest.ACMCertificateDomainFromEnv(t)
	domain := acctest.ACMCertificateRandomSubDomain(rootDomain)
	dataSourceName := "data.aws_cloudfront_distribution_tenant.test"
	resourceName := "aws_cloudfront_distribution_tenant.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.CloudFrontEndpointID)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.CloudFrontServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccDistributionTenantDataSourceConfig_byDomain(rName, rootDomain, domain),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrPair(dataSourceName, names.AttrARN, resourceName, names.AttrARN),
					resource.TestCheckResourceAttrPair(dataSourceName, "connection_group_id", resourceName, "connection_group_id"),
					resource.TestCheckResourceAttrPair(dataSourceName, "customizations.#", resourceName, "customizations.#"),
					resource.TestCheckResourceAttrPair(dataSourceName, "distribution_id", resourceName, "distribution_id"),
					resource.TestCheckResourceAttrPair(dataSourceName, "domains.#", resourceName, "domain.#"),
					resource.TestCheckResourceAttrPair(dataSourceName, names.AttrEnabled, resourceName, names.AttrEnabled),
					resource.TestCheckResourceAttrPair(dataSourceName, "etag", resourceName, "etag"),
					resource.TestCheckResourceAttrPair(dataSourceName, "managed_certificate_request.#", resourceName, "managed_certificate_request.#"),
					resource.TestCheckResourceAttrPair(dataSourceName, names.AttrName, resourceName, names.AttrName),
					resource.TestCheckResourceAttrPair(dataSourceName, "parameters.#", resourceName, "parameter.#"),
					resource.TestCheckResourceAttrSet(dataSourceName, names.AttrStatus),
				),
			},
		},
	})
}

func testAccDistributionTenantDataSourceConfig_basic(rName, rootDomain, tenantDomain string) string {
	return acctest.ConfigCompose(testAccDistributionTenantConfig_basic(rName, rootDomain, tenantDomain), `
data "aws_cloudfront_distribution_tenant" "test" {
  id = aws_cloudfront_distribution_tenant.test.id
}
`)
}

func testAccDistributionTenantDataSourceConfig_byARN(rName, rootDomain, tenantDomain string) string {
	return acctest.ConfigCompose(testAccDistributionTenantConfig_basic(rName, rootDomain, tenantDomain), `
data "aws_cloudfront_distribution_tenant" "test" {
  arn = aws_cloudfront_distribution_tenant.test.arn
}
`)
}

func testAccDistributionTenantDataSourceConfig_byName(rName, rootDomain, tenantDomain string) string {
	return acctest.ConfigCompose(testAccDistributionTenantConfig_basic(rName, rootDomain, tenantDomain), `
data "aws_cloudfront_distribution_tenant" "test" {
  name = aws_cloudfront_distribution_tenant.test.name
}
`)
}

func testAccDistributionTenantDataSourceConfig_byDomain(rName, rootDomain, tenantDomain string) string {
	return acctest.ConfigCompose(testAccDistributionTenantConfig_basic(rName, rootDomain, tenantDomain), `
data "aws_cloudfront_distribution_tenant" "test" {
  domain = aws_cloudfront_distribution_tenant.test.domain[0].domain
}
`)
}
