// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package meta_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/aws-sdk-go-base/v2/endpoints"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	tfmeta "github.com/hashicorp/terraform-provider-aws/internal/service/meta"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccMetaServiceDataSource_basic(t *testing.T) {
	ctx := acctest.Context(t)
	dataSourceName := "data.aws_service.test"
	serviceID := "ec2"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, tfmeta.PseudoServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccServiceDataSourceConfig_serviceID(serviceID),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(dataSourceName, names.AttrDNSName, fmt.Sprintf("%s.%s.%s", serviceID, acctest.Region(), "amazonaws.com")),
					resource.TestCheckResourceAttr(dataSourceName, "partition", acctest.Partition()),
					resource.TestCheckResourceAttr(dataSourceName, "reverse_dns_prefix", "com.amazonaws"),
					resource.TestCheckResourceAttr(dataSourceName, names.AttrRegion, acctest.Region()),
					resource.TestCheckResourceAttr(dataSourceName, "reverse_dns_name", fmt.Sprintf("%s.%s.%s", "com.amazonaws", acctest.Region(), serviceID)),
					resource.TestCheckResourceAttr(dataSourceName, "service_id", serviceID),
					resource.TestCheckResourceAttr(dataSourceName, "supported", acctest.CtTrue),
				),
			},
		},
	})
}

func TestAccMetaServiceDataSource_irregularServiceID(t *testing.T) {
	ctx := acctest.Context(t)
	dataSourceName := "data.aws_service.test"
	serviceID := "resource-explorer-2"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, tfmeta.PseudoServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccServiceDataSourceConfig_serviceID(serviceID),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(dataSourceName, names.AttrDNSName, fmt.Sprintf("%s.%s.%s", serviceID, acctest.Region(), "amazonaws.com")),
					resource.TestCheckResourceAttr(dataSourceName, "partition", acctest.Partition()),
					resource.TestCheckResourceAttr(dataSourceName, "reverse_dns_prefix", "com.amazonaws"),
					resource.TestCheckResourceAttr(dataSourceName, names.AttrRegion, acctest.Region()),
					resource.TestCheckResourceAttr(dataSourceName, "reverse_dns_name", fmt.Sprintf("%s.%s.%s", "com.amazonaws", acctest.Region(), serviceID)),
					resource.TestCheckResourceAttr(dataSourceName, "service_id", serviceID),
					resource.TestCheckResourceAttr(dataSourceName, "supported", acctest.CtTrue),
				),
			},
		},
	})
}
func TestAccMetaServiceDataSource_irregularServiceIDUnsupported(t *testing.T) {
	ctx := acctest.Context(t)
	dataSourceName := "data.aws_service.test"
	serviceID := "resourceexplorer2"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, tfmeta.PseudoServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccServiceDataSourceConfig_serviceID(serviceID),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(dataSourceName, names.AttrDNSName, fmt.Sprintf("%s.%s.%s", serviceID, acctest.Region(), "amazonaws.com")),
					resource.TestCheckResourceAttr(dataSourceName, "partition", acctest.Partition()),
					resource.TestCheckResourceAttr(dataSourceName, "reverse_dns_prefix", "com.amazonaws"),
					resource.TestCheckResourceAttr(dataSourceName, names.AttrRegion, acctest.Region()),
					resource.TestCheckResourceAttr(dataSourceName, "reverse_dns_name", fmt.Sprintf("%s.%s.%s", "com.amazonaws", acctest.Region(), serviceID)),
					resource.TestCheckResourceAttr(dataSourceName, "service_id", serviceID),
					resource.TestCheckResourceAttr(dataSourceName, "supported", acctest.CtFalse),
				),
			},
		},
	})
}

func TestAccMetaServiceDataSource_byReverseDNSName(t *testing.T) {
	ctx := acctest.Context(t)
	dataSourceName := "data.aws_service.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, tfmeta.PseudoServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccServiceDataSourceConfig_byReverseDNSName(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(dataSourceName, names.AttrRegion, endpoints.CnNorth1RegionID),
					resource.TestCheckResourceAttr(dataSourceName, "reverse_dns_name", fmt.Sprintf("%s.%s.%s", "cn.com.amazonaws", endpoints.CnNorth1RegionID, names.S3)),
					resource.TestCheckResourceAttr(dataSourceName, "reverse_dns_prefix", "cn.com.amazonaws"),
					resource.TestCheckResourceAttr(dataSourceName, "service_id", names.S3),
					resource.TestCheckResourceAttr(dataSourceName, "supported", acctest.CtTrue),
				),
			},
		},
	})
}

func TestAccMetaServiceDataSource_byDNSName(t *testing.T) {
	ctx := acctest.Context(t)
	dataSourceName := "data.aws_service.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, tfmeta.PseudoServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccServiceDataSourceConfig_byDNSName(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(dataSourceName, names.AttrRegion, endpoints.UsEast1RegionID),
					resource.TestCheckResourceAttr(dataSourceName, "reverse_dns_name", fmt.Sprintf("%s.%s.%s", "com.amazonaws", endpoints.UsEast1RegionID, names.RDS)),
					resource.TestCheckResourceAttr(dataSourceName, "reverse_dns_prefix", "com.amazonaws"),
					resource.TestCheckResourceAttr(dataSourceName, "service_id", names.RDS),
					resource.TestCheckResourceAttr(dataSourceName, "supported", acctest.CtTrue),
				),
			},
		},
	})
}

func TestAccMetaServiceDataSource_byParts(t *testing.T) {
	ctx := acctest.Context(t)
	dataSourceName := "data.aws_service.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, tfmeta.PseudoServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccServiceDataSourceConfig_byPart(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(dataSourceName, names.AttrDNSName, fmt.Sprintf("%s.%s.%s", names.S3, acctest.Region(), "amazonaws.com")),
					resource.TestCheckResourceAttr(dataSourceName, "reverse_dns_name", fmt.Sprintf("%s.%s.%s", "com.amazonaws", acctest.Region(), names.S3)),
					resource.TestCheckResourceAttr(dataSourceName, "supported", acctest.CtTrue),
				),
			},
		},
	})
}

func TestAccMetaServiceDataSource_unsupported(t *testing.T) {
	ctx := acctest.Context(t)
	dataSourceName := "data.aws_service.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, tfmeta.PseudoServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccServiceDataSourceConfig_unsupported(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(dataSourceName, names.AttrDNSName, fmt.Sprintf("%s.%s.%s", names.WAFEndpointID, endpoints.UsGovWest1RegionID, "amazonaws.com")),
					resource.TestCheckResourceAttr(dataSourceName, "partition", endpoints.AwsUsGovPartitionID),
					resource.TestCheckResourceAttr(dataSourceName, "reverse_dns_prefix", "com.amazonaws"),
					resource.TestCheckResourceAttr(dataSourceName, names.AttrRegion, endpoints.UsGovWest1RegionID),
					resource.TestCheckResourceAttr(dataSourceName, "reverse_dns_name", fmt.Sprintf("%s.%s.%s", "com.amazonaws", endpoints.UsGovWest1RegionID, names.WAFEndpointID)),
					resource.TestCheckResourceAttr(dataSourceName, "service_id", names.WAFEndpointID),
					resource.TestCheckResourceAttr(dataSourceName, "supported", acctest.CtFalse),
				),
			},
		},
	})
}

func testAccServiceDataSourceConfig_serviceID(id string) string {
	return fmt.Sprintf(`
data "aws_service" "test" {
  service_id = %[1]q
}
`, id)
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
  region             = data.aws_region.current.region
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
