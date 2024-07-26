// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package ec2_test

import (
	"fmt"
	"testing"

	"github.com/YakDriver/regexache"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccVPCEndpointServiceDataSource_ServiceType_gateway(t *testing.T) {
	ctx := acctest.Context(t)
	datasourceName := "data.aws_vpc_endpoint_service.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccVPCEndpointServiceDataSourceConfig_type("dynamodb", "Gateway"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(datasourceName, "acceptance_required", acctest.CtFalse),
					acctest.MatchResourceAttrRegionalARN(datasourceName, names.AttrARN, "ec2", regexache.MustCompile(`vpc-endpoint-service/vpce-svc-.+`)),
					acctest.CheckResourceAttrGreaterThanValue(datasourceName, "availability_zones.#", 0),
					resource.TestCheckResourceAttr(datasourceName, "base_endpoint_dns_names.#", acctest.Ct1),
					resource.TestCheckResourceAttr(datasourceName, "manages_vpc_endpoints", acctest.CtFalse),
					resource.TestCheckResourceAttr(datasourceName, names.AttrOwner, "amazon"),
					resource.TestCheckResourceAttr(datasourceName, "private_dns_name", ""),
					testAccCheckResourceAttrRegionalReverseDNSService(datasourceName, names.AttrServiceName, "dynamodb"),
					resource.TestCheckResourceAttr(datasourceName, "service_type", "Gateway"),
					acctest.CheckResourceAttrGreaterThanValue(datasourceName, "supported_ip_address_types.#", 0),
					resource.TestCheckResourceAttr(datasourceName, acctest.CtTagsPercent, acctest.Ct0),
					resource.TestCheckResourceAttr(datasourceName, "vpc_endpoint_policy_supported", acctest.CtTrue),
				),
			},
		},
	})
}

func TestAccVPCEndpointServiceDataSource_ServiceType_interface(t *testing.T) {
	ctx := acctest.Context(t)
	datasourceName := "data.aws_vpc_endpoint_service.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccVPCEndpointServiceDataSourceConfig_type("ec2", "Interface"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(datasourceName, "acceptance_required", acctest.CtFalse),
					acctest.MatchResourceAttrRegionalARN(datasourceName, names.AttrARN, "ec2", regexache.MustCompile(`vpc-endpoint-service/vpce-svc-.+`)),
					acctest.CheckResourceAttrGreaterThanValue(datasourceName, "availability_zones.#", 0),
					resource.TestCheckResourceAttr(datasourceName, "base_endpoint_dns_names.#", acctest.Ct1),
					resource.TestCheckResourceAttr(datasourceName, "manages_vpc_endpoints", acctest.CtFalse),
					resource.TestCheckResourceAttr(datasourceName, names.AttrOwner, "amazon"),
					acctest.CheckResourceAttrRegionalHostnameService(datasourceName, "private_dns_name", "ec2"),
					testAccCheckResourceAttrRegionalReverseDNSService(datasourceName, names.AttrServiceName, "ec2"),
					resource.TestCheckResourceAttr(datasourceName, "service_type", "Interface"),
					acctest.CheckResourceAttrGreaterThanValue(datasourceName, "supported_ip_address_types.#", 0),
					resource.TestCheckResourceAttr(datasourceName, acctest.CtTagsPercent, acctest.Ct0),
					resource.TestCheckResourceAttr(datasourceName, "vpc_endpoint_policy_supported", acctest.CtTrue),
				),
			},
		},
	})
}

func TestAccVPCEndpointServiceDataSource_custom(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_vpc_endpoint_service.test"
	datasourceName := "data.aws_vpc_endpoint_service.test"
	rName := sdkacctest.RandomWithPrefix("tfacctest") // 32 character limit

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccVPCEndpointServiceDataSourceConfig_custom(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrPair(datasourceName, "acceptance_required", resourceName, "acceptance_required"),
					resource.TestCheckResourceAttrPair(datasourceName, names.AttrARN, resourceName, names.AttrARN),
					resource.TestCheckResourceAttrPair(datasourceName, "availability_zones.#", resourceName, "availability_zones.#"),
					resource.TestCheckResourceAttrPair(datasourceName, "base_endpoint_dns_names.#", resourceName, "base_endpoint_dns_names.#"),
					resource.TestCheckResourceAttrPair(datasourceName, "manages_vpc_endpoints", resourceName, "manages_vpc_endpoints"),
					acctest.CheckResourceAttrAccountID(datasourceName, names.AttrOwner),
					resource.TestCheckResourceAttrPair(datasourceName, "private_dns_name", resourceName, "private_dns_name"),
					resource.TestCheckResourceAttr(datasourceName, "service_type", "Interface"),
					resource.TestCheckResourceAttrPair(datasourceName, "supported_ip_address_types.#", resourceName, "supported_ip_address_types.#"),
					resource.TestCheckResourceAttrPair(datasourceName, acctest.CtTagsPercent, resourceName, acctest.CtTagsPercent),
					resource.TestCheckResourceAttr(datasourceName, "vpc_endpoint_policy_supported", acctest.CtFalse),
				),
			},
		},
	})
}

func TestAccVPCEndpointServiceDataSource_Custom_filter(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_vpc_endpoint_service.test"
	datasourceName := "data.aws_vpc_endpoint_service.test"
	rName := sdkacctest.RandomWithPrefix("tfacctest") // 32 character limit

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccVPCEndpointServiceDataSourceConfig_customFilter(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrPair(datasourceName, "acceptance_required", resourceName, "acceptance_required"),
					resource.TestCheckResourceAttrPair(datasourceName, names.AttrARN, resourceName, names.AttrARN),
					resource.TestCheckResourceAttrPair(datasourceName, "availability_zones.#", resourceName, "availability_zones.#"),
					resource.TestCheckResourceAttrPair(datasourceName, "base_endpoint_dns_names.#", resourceName, "base_endpoint_dns_names.#"),
					resource.TestCheckResourceAttrPair(datasourceName, "manages_vpc_endpoints", resourceName, "manages_vpc_endpoints"),
					acctest.CheckResourceAttrAccountID(datasourceName, names.AttrOwner),
					resource.TestCheckResourceAttrPair(datasourceName, "private_dns_name", resourceName, "private_dns_name"),
					resource.TestCheckResourceAttr(datasourceName, "service_type", "Interface"),
					resource.TestCheckResourceAttrPair(datasourceName, "supported_ip_address_types.#", resourceName, "supported_ip_address_types.#"),
					resource.TestCheckResourceAttrPair(datasourceName, acctest.CtTagsPercent, resourceName, acctest.CtTagsPercent),
					resource.TestCheckResourceAttr(datasourceName, "vpc_endpoint_policy_supported", acctest.CtFalse),
				),
			},
		},
	})
}

func TestAccVPCEndpointServiceDataSource_CustomFilter_tags(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_vpc_endpoint_service.test"
	datasourceName := "data.aws_vpc_endpoint_service.test"
	rName := sdkacctest.RandomWithPrefix("tfacctest") // 32 character limit

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccVPCEndpointServiceDataSourceConfig_customFilterTags(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrPair(datasourceName, "acceptance_required", resourceName, "acceptance_required"),
					resource.TestCheckResourceAttrPair(datasourceName, names.AttrARN, resourceName, names.AttrARN),
					resource.TestCheckResourceAttrPair(datasourceName, "availability_zones.#", resourceName, "availability_zones.#"),
					resource.TestCheckResourceAttrPair(datasourceName, "base_endpoint_dns_names.#", resourceName, "base_endpoint_dns_names.#"),
					resource.TestCheckResourceAttrPair(datasourceName, "manages_vpc_endpoints", resourceName, "manages_vpc_endpoints"),
					acctest.CheckResourceAttrAccountID(datasourceName, names.AttrOwner),
					resource.TestCheckResourceAttrPair(datasourceName, "private_dns_name", resourceName, "private_dns_name"),
					resource.TestCheckResourceAttr(datasourceName, "service_type", "Interface"),
					resource.TestCheckResourceAttrPair(datasourceName, "supported_ip_address_types.#", resourceName, "supported_ip_address_types.#"),
					resource.TestCheckResourceAttrPair(datasourceName, acctest.CtTagsPercent, resourceName, acctest.CtTagsPercent),
					resource.TestCheckResourceAttr(datasourceName, "vpc_endpoint_policy_supported", acctest.CtFalse),
				),
			},
		},
	})
}

// testAccCheckResourceAttrRegionalReverseDNSService ensures the Terraform state exactly matches a service reverse DNS hostname with region and partition DNS suffix
//
// For example: com.amazonaws.us-west-2.s3
func testAccCheckResourceAttrRegionalReverseDNSService(resourceName, attributeName, serviceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		reverseDns := fmt.Sprintf("%s.%s.%s", acctest.PartitionReverseDNSPrefix(), acctest.Region(), serviceName)

		return resource.TestCheckResourceAttr(resourceName, attributeName, reverseDns)(s)
	}
}

func testAccVPCEndpointServiceDataSourceConfig_type(service string, serviceType string) string {
	return fmt.Sprintf(`
data "aws_vpc_endpoint_service" "test" {
  service      = %[1]q
  service_type = %[2]q
}
`, service, serviceType)
}

func testAccVPCEndpointServiceDataSourceConfig_customBase(rName string) string {
	return acctest.ConfigCompose(testAccVPCEndpointServiceConfig_baseNetworkLoadBalancer(rName, 1), fmt.Sprintf(`
resource "aws_vpc_endpoint_service" "test" {
  acceptance_required        = false
  network_load_balancer_arns = aws_lb.test[*].arn

  tags = {
    Name = %[1]q
  }
}
`, rName))
}

func testAccVPCEndpointServiceDataSourceConfig_custom(rName string) string {
	return acctest.ConfigCompose(testAccVPCEndpointServiceDataSourceConfig_customBase(rName), `
data "aws_vpc_endpoint_service" "test" {
  service_name = aws_vpc_endpoint_service.test.service_name
}
`)
}

func testAccVPCEndpointServiceDataSourceConfig_customFilter(rName string) string {
	return acctest.ConfigCompose(testAccVPCEndpointServiceDataSourceConfig_customBase(rName), `
data "aws_vpc_endpoint_service" "test" {
  filter {
    name   = "service-name"
    values = [aws_vpc_endpoint_service.test.service_name]
  }
}
`)
}

func testAccVPCEndpointServiceDataSourceConfig_customFilterTags(rName string) string {
	return acctest.ConfigCompose(testAccVPCEndpointServiceDataSourceConfig_customBase(rName), `
data "aws_vpc_endpoint_service" "test" {
  tags = {
    Name = aws_vpc_endpoint_service.test.tags["Name"]
  }
}
`)
}
