// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package meta_test

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/service/rds"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	tfmeta "github.com/hashicorp/terraform-provider-aws/internal/service/meta"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccMetaService_basic(t *testing.T) {
	ctx := acctest.Context(t)
	dataSourceName := "data.aws_service.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, tfmeta.PseudoServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccServiceDataSourceConfig_basic(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(dataSourceName, names.AttrDNSName, fmt.Sprintf("%s.%s.%s", names.EC2, acctest.Region(), "amazonaws.com")),
					resource.TestCheckResourceAttr(dataSourceName, "partition", acctest.Partition()),
					resource.TestCheckResourceAttr(dataSourceName, "reverse_dns_prefix", "com.amazonaws"),
					resource.TestCheckResourceAttr(dataSourceName, names.AttrRegion, acctest.Region()),
					resource.TestCheckResourceAttr(dataSourceName, "reverse_dns_name", fmt.Sprintf("%s.%s.%s", "com.amazonaws", acctest.Region(), names.EC2)),
					resource.TestCheckResourceAttr(dataSourceName, "service_id", names.EC2),
					resource.TestCheckResourceAttr(dataSourceName, "supported", acctest.CtTrue),
				),
			},
		},
	})
}

func TestAccMetaService_byReverseDNSName(t *testing.T) {
	ctx := acctest.Context(t)
	dataSourceName := "data.aws_service.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, tfmeta.PseudoServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccServiceDataSourceConfig_byReverseDNSName(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(dataSourceName, names.AttrRegion, names.CNNorth1RegionID),
					resource.TestCheckResourceAttr(dataSourceName, "reverse_dns_name", fmt.Sprintf("%s.%s.%s", "cn.com.amazonaws", names.CNNorth1RegionID, s3.EndpointsID)),
					resource.TestCheckResourceAttr(dataSourceName, "reverse_dns_prefix", "cn.com.amazonaws"),
					resource.TestCheckResourceAttr(dataSourceName, "service_id", s3.EndpointsID),
					resource.TestCheckResourceAttr(dataSourceName, "supported", acctest.CtTrue),
				),
			},
		},
	})
}

func TestAccMetaService_byDNSName(t *testing.T) {
	ctx := acctest.Context(t)
	dataSourceName := "data.aws_service.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, tfmeta.PseudoServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccServiceDataSourceConfig_byDNSName(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(dataSourceName, names.AttrRegion, names.USEast1RegionID),
					resource.TestCheckResourceAttr(dataSourceName, "reverse_dns_name", fmt.Sprintf("%s.%s.%s", "com.amazonaws", names.USEast1RegionID, rds.EndpointsID)),
					resource.TestCheckResourceAttr(dataSourceName, "reverse_dns_prefix", "com.amazonaws"),
					resource.TestCheckResourceAttr(dataSourceName, "service_id", rds.EndpointsID),
					resource.TestCheckResourceAttr(dataSourceName, "supported", acctest.CtTrue),
				),
			},
		},
	})
}

func TestAccMetaService_byParts(t *testing.T) {
	ctx := acctest.Context(t)
	dataSourceName := "data.aws_service.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, tfmeta.PseudoServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccServiceDataSourceConfig_byPart(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(dataSourceName, names.AttrDNSName, fmt.Sprintf("%s.%s.%s", s3.EndpointsID, acctest.Region(), "amazonaws.com")),
					resource.TestCheckResourceAttr(dataSourceName, "reverse_dns_name", fmt.Sprintf("%s.%s.%s", "com.amazonaws", acctest.Region(), s3.EndpointsID)),
					resource.TestCheckResourceAttr(dataSourceName, "supported", acctest.CtTrue),
				),
			},
		},
	})
}

func TestAccMetaService_unsupported(t *testing.T) {
	ctx := acctest.Context(t)
	dataSourceName := "data.aws_service.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, tfmeta.PseudoServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccServiceDataSourceConfig_unsupported(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(dataSourceName, names.AttrDNSName, fmt.Sprintf("%s.%s.%s", names.WAFEndpointID, names.USGovWest1RegionID, "amazonaws.com")),
					resource.TestCheckResourceAttr(dataSourceName, "partition", names.USGovCloudPartitionID),
					resource.TestCheckResourceAttr(dataSourceName, "reverse_dns_prefix", "com.amazonaws"),
					resource.TestCheckResourceAttr(dataSourceName, names.AttrRegion, names.USGovWest1RegionID),
					resource.TestCheckResourceAttr(dataSourceName, "reverse_dns_name", fmt.Sprintf("%s.%s.%s", "com.amazonaws", names.USGovWest1RegionID, names.WAFEndpointID)),
					resource.TestCheckResourceAttr(dataSourceName, "service_id", names.WAFEndpointID),
					resource.TestCheckResourceAttr(dataSourceName, "supported", acctest.CtFalse),
				),
			},
		},
	})
}

func testAccServiceDataSourceConfig_basic() string {
	return fmt.Sprintf(`
data "aws_service" "test" {
  service_id = %[1]q
}
`, names.EC2)
}

func testAccServiceDataSourceConfig_byReverseDNSName() string {
	// lintignore:AWSAT003
	return `
data "aws_service" "test" {
  reverse_dns_name = "cn.com.amazonaws.cn-north-1.s3"
}
`
}

func testAccServiceDataSourceConfig_byDNSName() string {
	// lintignore:AWSAT003
	return `
data "aws_service" "test" {
  dns_name = "rds.us-east-1.amazonaws.com"
}
`
}

func testAccServiceDataSourceConfig_byPart() string {
	return `
data "aws_region" "current" {}

data "aws_service" "test" {
  reverse_dns_prefix = "com.amazonaws"
  region             = data.aws_region.current.name
  service_id         = "s3"
}
`
}

func testAccServiceDataSourceConfig_unsupported() string {
	// lintignore:AWSAT003
	return `
data "aws_service" "test" {
  reverse_dns_name = "com.amazonaws.us-gov-west-1.waf"
}
`
}
