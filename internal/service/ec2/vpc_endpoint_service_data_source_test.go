package ec2_test

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/aws/aws-sdk-go/service/ec2"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
)

func TestAccVPCEndpointServiceDataSource_gateway(t *testing.T) {
	datasourceName := "data.aws_vpc_endpoint_service.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ec2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccVPCEndpointServiceDataSourceConfig_gateway,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(datasourceName, "acceptance_required", "false"),
					acctest.MatchResourceAttrRegionalARN(datasourceName, "arn", "ec2", regexp.MustCompile(`vpc-endpoint-service/vpce-svc-.+`)),
					acctest.CheckResourceAttrGreaterThanValue(datasourceName, "availability_zones.#", "0"),
					resource.TestCheckResourceAttr(datasourceName, "base_endpoint_dns_names.#", "1"),
					resource.TestCheckResourceAttr(datasourceName, "manages_vpc_endpoints", "false"),
					resource.TestCheckResourceAttr(datasourceName, "owner", "amazon"),
					resource.TestCheckResourceAttr(datasourceName, "private_dns_name", ""),
					testAccCheckResourceAttrRegionalReverseDNSService(datasourceName, "service_name", "dynamodb"),
					resource.TestCheckResourceAttr(datasourceName, "service_type", "Gateway"),
					acctest.CheckResourceAttrGreaterThanValue(datasourceName, "supported_ip_address_types.#", "0"),
					resource.TestCheckResourceAttr(datasourceName, "tags.%", "0"),
					resource.TestCheckResourceAttr(datasourceName, "vpc_endpoint_policy_supported", "true"),
				),
			},
		},
	})
}

func TestAccVPCEndpointServiceDataSource_interface(t *testing.T) {
	datasourceName := "data.aws_vpc_endpoint_service.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ec2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccVPCEndpointServiceDataSourceConfig_interface,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(datasourceName, "acceptance_required", "false"),
					acctest.MatchResourceAttrRegionalARN(datasourceName, "arn", "ec2", regexp.MustCompile(`vpc-endpoint-service/vpce-svc-.+`)),
					acctest.CheckResourceAttrGreaterThanValue(datasourceName, "availability_zones.#", "0"),
					resource.TestCheckResourceAttr(datasourceName, "base_endpoint_dns_names.#", "1"),
					resource.TestCheckResourceAttr(datasourceName, "manages_vpc_endpoints", "false"),
					resource.TestCheckResourceAttr(datasourceName, "owner", "amazon"),
					acctest.CheckResourceAttrRegionalHostnameService(datasourceName, "private_dns_name", "ec2"),
					testAccCheckResourceAttrRegionalReverseDNSService(datasourceName, "service_name", "ec2"),
					resource.TestCheckResourceAttr(datasourceName, "service_type", "Interface"),
					acctest.CheckResourceAttrGreaterThanValue(datasourceName, "supported_ip_address_types.#", "0"),
					resource.TestCheckResourceAttr(datasourceName, "tags.%", "0"),
					resource.TestCheckResourceAttr(datasourceName, "vpc_endpoint_policy_supported", "true"),
				),
			},
		},
	})
}

func TestAccVPCEndpointServiceDataSource_custom(t *testing.T) {
	resourceName := "aws_vpc_endpoint_service.test"
	datasourceName := "data.aws_vpc_endpoint_service.test"
	rName := sdkacctest.RandomWithPrefix("tfacctest") // 32 character limit

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ec2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccVPCEndpointServiceDataSourceConfig_custom(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrPair(datasourceName, "acceptance_required", resourceName, "acceptance_required"),
					resource.TestCheckResourceAttrPair(datasourceName, "arn", resourceName, "arn"),
					resource.TestCheckResourceAttrPair(datasourceName, "availability_zones.#", resourceName, "availability_zones.#"),
					resource.TestCheckResourceAttrPair(datasourceName, "base_endpoint_dns_names.#", resourceName, "base_endpoint_dns_names.#"),
					resource.TestCheckResourceAttrPair(datasourceName, "manages_vpc_endpoints", resourceName, "manages_vpc_endpoints"),
					acctest.CheckResourceAttrAccountID(datasourceName, "owner"),
					resource.TestCheckResourceAttrPair(datasourceName, "private_dns_name", resourceName, "private_dns_name"),
					resource.TestCheckResourceAttr(datasourceName, "service_type", "Interface"),
					resource.TestCheckResourceAttrPair(datasourceName, "supported_ip_address_types.#", resourceName, "supported_ip_address_types.#"),
					resource.TestCheckResourceAttrPair(datasourceName, "tags.%", resourceName, "tags.%"),
					resource.TestCheckResourceAttr(datasourceName, "vpc_endpoint_policy_supported", "false"),
				),
			},
		},
	})
}

func TestAccVPCEndpointServiceDataSource_Custom_filter(t *testing.T) {
	resourceName := "aws_vpc_endpoint_service.test"
	datasourceName := "data.aws_vpc_endpoint_service.test"
	rName := sdkacctest.RandomWithPrefix("tfacctest") // 32 character limit

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ec2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccVPCEndpointServiceDataSourceConfig_customFilter(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrPair(datasourceName, "acceptance_required", resourceName, "acceptance_required"),
					resource.TestCheckResourceAttrPair(datasourceName, "arn", resourceName, "arn"),
					resource.TestCheckResourceAttrPair(datasourceName, "availability_zones.#", resourceName, "availability_zones.#"),
					resource.TestCheckResourceAttrPair(datasourceName, "base_endpoint_dns_names.#", resourceName, "base_endpoint_dns_names.#"),
					resource.TestCheckResourceAttrPair(datasourceName, "manages_vpc_endpoints", resourceName, "manages_vpc_endpoints"),
					acctest.CheckResourceAttrAccountID(datasourceName, "owner"),
					resource.TestCheckResourceAttrPair(datasourceName, "private_dns_name", resourceName, "private_dns_name"),
					resource.TestCheckResourceAttr(datasourceName, "service_type", "Interface"),
					resource.TestCheckResourceAttrPair(datasourceName, "supported_ip_address_types.#", resourceName, "supported_ip_address_types.#"),
					resource.TestCheckResourceAttrPair(datasourceName, "tags.%", resourceName, "tags.%"),
					resource.TestCheckResourceAttr(datasourceName, "vpc_endpoint_policy_supported", "false"),
				),
			},
		},
	})
}

func TestAccVPCEndpointServiceDataSource_CustomFilter_tags(t *testing.T) {
	resourceName := "aws_vpc_endpoint_service.test"
	datasourceName := "data.aws_vpc_endpoint_service.test"
	rName := sdkacctest.RandomWithPrefix("tfacctest") // 32 character limit

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ec2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccVPCEndpointServiceDataSourceConfig_customFilterTags(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrPair(datasourceName, "acceptance_required", resourceName, "acceptance_required"),
					resource.TestCheckResourceAttrPair(datasourceName, "arn", resourceName, "arn"),
					resource.TestCheckResourceAttrPair(datasourceName, "availability_zones.#", resourceName, "availability_zones.#"),
					resource.TestCheckResourceAttrPair(datasourceName, "base_endpoint_dns_names.#", resourceName, "base_endpoint_dns_names.#"),
					resource.TestCheckResourceAttrPair(datasourceName, "manages_vpc_endpoints", resourceName, "manages_vpc_endpoints"),
					acctest.CheckResourceAttrAccountID(datasourceName, "owner"),
					resource.TestCheckResourceAttrPair(datasourceName, "private_dns_name", resourceName, "private_dns_name"),
					resource.TestCheckResourceAttr(datasourceName, "service_type", "Interface"),
					resource.TestCheckResourceAttrPair(datasourceName, "supported_ip_address_types.#", resourceName, "supported_ip_address_types.#"),
					resource.TestCheckResourceAttrPair(datasourceName, "tags.%", resourceName, "tags.%"),
					resource.TestCheckResourceAttr(datasourceName, "vpc_endpoint_policy_supported", "false"),
				),
			},
		},
	})
}

func TestAccVPCEndpointServiceDataSource_ServiceType_gateway(t *testing.T) {
	datasourceName := "data.aws_vpc_endpoint_service.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ec2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccVPCEndpointServiceDataSourceConfig_type("s3", "Gateway"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckResourceAttrRegionalReverseDNSService(datasourceName, "service_name", "s3"),
					resource.TestCheckResourceAttr(datasourceName, "service_type", "Gateway"),
				),
			},
		},
	})
}

func TestAccVPCEndpointServiceDataSource_ServiceType_interface(t *testing.T) {
	datasourceName := "data.aws_vpc_endpoint_service.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ec2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccVPCEndpointServiceDataSourceConfig_type("ec2", "Interface"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckResourceAttrRegionalReverseDNSService(datasourceName, "service_name", "ec2"),
					resource.TestCheckResourceAttr(datasourceName, "service_type", "Interface"),
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

const testAccVPCEndpointServiceDataSourceConfig_gateway = `
data "aws_vpc_endpoint_service" "test" {
  service = "dynamodb"
}
`

const testAccVPCEndpointServiceDataSourceConfig_interface = `
data "aws_vpc_endpoint_service" "test" {
  service = "ec2"
}
`

func testAccVPCEndpointServiceDataSourceConfig_type(service string, serviceType string) string {
	return fmt.Sprintf(`
data "aws_vpc_endpoint_service" "test" {
  service      = %[1]q
  service_type = %[2]q
}
`, service, serviceType)
}

func testAccVPCEndpointServiceDataSourceConfig_customBase(rName string) string {
	return acctest.ConfigCompose(testAccVPCEndpointServiceConfig_networkLoadBalancerBase(rName, 1), fmt.Sprintf(`
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
