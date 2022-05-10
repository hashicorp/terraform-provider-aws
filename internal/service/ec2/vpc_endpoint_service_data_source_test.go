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
				Config: testAccVPCEndpointServiceGatewayDataSourceConfig,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckResourceAttrRegionalReverseDNSService(datasourceName, "service_name", "dynamodb"),
					resource.TestCheckResourceAttr(datasourceName, "acceptance_required", "false"),
					resource.TestCheckResourceAttrPair(datasourceName, "availability_zones.#", "data.aws_availability_zones.available", "names.#"),
					resource.TestCheckResourceAttr(datasourceName, "base_endpoint_dns_names.#", "1"),
					resource.TestCheckResourceAttr(datasourceName, "manages_vpc_endpoints", "false"),
					resource.TestCheckResourceAttr(datasourceName, "owner", "amazon"),
					resource.TestCheckResourceAttr(datasourceName, "private_dns_name", ""),
					resource.TestCheckResourceAttr(datasourceName, "service_type", "Gateway"),
					resource.TestCheckResourceAttr(datasourceName, "vpc_endpoint_policy_supported", "true"),
					resource.TestCheckResourceAttr(datasourceName, "tags.%", "0"),
					acctest.MatchResourceAttrRegionalARN(datasourceName, "arn", "ec2", regexp.MustCompile(`vpc-endpoint-service/vpce-svc-.+`)),
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
				Config: testAccVPCEndpointServiceInterfaceDataSourceConfig,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckResourceAttrRegionalReverseDNSService(datasourceName, "service_name", "ec2"),
					resource.TestCheckResourceAttr(datasourceName, "acceptance_required", "false"),
					resource.TestCheckResourceAttr(datasourceName, "base_endpoint_dns_names.#", "1"),
					resource.TestCheckResourceAttr(datasourceName, "manages_vpc_endpoints", "false"),
					resource.TestCheckResourceAttr(datasourceName, "owner", "amazon"),
					acctest.CheckResourceAttrRegionalHostnameService(datasourceName, "private_dns_name", "ec2"),
					resource.TestCheckResourceAttr(datasourceName, "service_type", "Interface"),
					resource.TestCheckResourceAttr(datasourceName, "vpc_endpoint_policy_supported", "true"),
					resource.TestCheckResourceAttr(datasourceName, "tags.%", "0"),
					acctest.MatchResourceAttrRegionalARN(datasourceName, "arn", "ec2", regexp.MustCompile(`vpc-endpoint-service/vpce-svc-.+`)),
				),
			},
		},
	})
}

func TestAccVPCEndpointServiceDataSource_custom(t *testing.T) {
	datasourceName := "data.aws_vpc_endpoint_service.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ec2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccVPCEndpointServiceCustomDataSourceConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(datasourceName, "acceptance_required", "true"),
					resource.TestCheckResourceAttr(datasourceName, "availability_zones.#", "2"),
					resource.TestCheckResourceAttr(datasourceName, "manages_vpc_endpoints", "false"),
					acctest.CheckResourceAttrAccountID(datasourceName, "owner"),
					resource.TestCheckResourceAttr(datasourceName, "service_type", "Interface"),
					resource.TestCheckResourceAttr(datasourceName, "vpc_endpoint_policy_supported", "false"),
					resource.TestCheckResourceAttr(datasourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(datasourceName, "tags.Name", rName),
					acctest.MatchResourceAttrRegionalARN(datasourceName, "arn", "ec2", regexp.MustCompile(`vpc-endpoint-service/vpce-svc-.+`)),
				),
			},
		},
	})
}

func TestAccVPCEndpointServiceDataSource_Custom_filter(t *testing.T) {
	datasourceName := "data.aws_vpc_endpoint_service.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ec2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccVPCEndpointServiceCustomFilterDataSourceConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(datasourceName, "acceptance_required", "true"),
					resource.TestCheckResourceAttr(datasourceName, "availability_zones.#", "2"),
					resource.TestCheckResourceAttr(datasourceName, "manages_vpc_endpoints", "false"),
					acctest.CheckResourceAttrAccountID(datasourceName, "owner"),
					resource.TestCheckResourceAttr(datasourceName, "service_type", "Interface"),
					resource.TestCheckResourceAttr(datasourceName, "vpc_endpoint_policy_supported", "false"),
					resource.TestCheckResourceAttr(datasourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(datasourceName, "tags.Name", rName),
					acctest.MatchResourceAttrRegionalARN(datasourceName, "arn", "ec2", regexp.MustCompile(`vpc-endpoint-service/vpce-svc-.+`)),
				),
			},
		},
	})
}

func TestAccVPCEndpointServiceDataSource_CustomFilter_tags(t *testing.T) {
	datasourceName := "data.aws_vpc_endpoint_service.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ec2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccVPCEndpointServiceCustomFilterTagsDataSourceConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(datasourceName, "acceptance_required", "true"),
					resource.TestCheckResourceAttr(datasourceName, "availability_zones.#", "2"),
					resource.TestCheckResourceAttr(datasourceName, "manages_vpc_endpoints", "false"),
					acctest.CheckResourceAttrAccountID(datasourceName, "owner"),
					resource.TestCheckResourceAttr(datasourceName, "service_type", "Interface"),
					resource.TestCheckResourceAttr(datasourceName, "vpc_endpoint_policy_supported", "false"),
					resource.TestCheckResourceAttr(datasourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(datasourceName, "tags.Name", rName),
					acctest.MatchResourceAttrRegionalARN(datasourceName, "arn", "ec2", regexp.MustCompile(`vpc-endpoint-service/vpce-svc-.+`)),
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
				Config: testAccVPCEndpointServiceDataSourceConfig_ServiceType("s3", "Gateway"),
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
				Config: testAccVPCEndpointServiceDataSourceConfig_ServiceType("ec2", "Interface"),
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

const testAccVPCEndpointServiceGatewayDataSourceConfig = `
data "aws_availability_zones" "available" {}

data "aws_vpc_endpoint_service" "test" {
  service = "dynamodb"
}
`

const testAccVPCEndpointServiceInterfaceDataSourceConfig = `
data "aws_vpc_endpoint_service" "test" {
  service = "ec2"
}
`

func testAccVPCEndpointServiceDataSourceConfig_ServiceType(service string, serviceType string) string {
	return fmt.Sprintf(`
data "aws_vpc_endpoint_service" "test" {
  service      = %[1]q
  service_type = %[2]q
}
`, service, serviceType)
}

func testAccVPCEndpointServiceCustomBaseDataSourceConfig(rName string) string {
	return fmt.Sprintf(`
resource "aws_vpc" "test" {
  cidr_block = "10.0.0.0/16"

  tags = {
    Name = %[1]q
  }
}

resource "aws_lb" "test" {
  name = %[1]q

  subnets = [
    aws_subnet.test1.id,
    aws_subnet.test2.id,
  ]

  load_balancer_type         = "network"
  internal                   = true
  idle_timeout               = 60
  enable_deletion_protection = false

  tags = {
    Name = %[1]q
  }
}

data "aws_availability_zones" "available" {
  state = "available"

  filter {
    name   = "opt-in-status"
    values = ["opt-in-not-required"]
  }
}

resource "aws_subnet" "test1" {
  vpc_id            = aws_vpc.test.id
  cidr_block        = "10.0.1.0/24"
  availability_zone = data.aws_availability_zones.available.names[0]

  tags = {
    Name = %[1]q
  }
}

resource "aws_subnet" "test2" {
  vpc_id            = aws_vpc.test.id
  cidr_block        = "10.0.2.0/24"
  availability_zone = data.aws_availability_zones.available.names[1]

  tags = {
    Name = %[1]q
  }
}

resource "aws_vpc_endpoint_service" "test" {
  acceptance_required = true

  network_load_balancer_arns = [
    aws_lb.test.id,
  ]

  tags = {
    Name = %[1]q
  }
}
`, rName)
}

func testAccVPCEndpointServiceCustomDataSourceConfig(rName string) string {
	return testAccVPCEndpointServiceCustomBaseDataSourceConfig(rName) + `
data "aws_vpc_endpoint_service" "test" {
  service_name = aws_vpc_endpoint_service.test.service_name
}
`
}

func testAccVPCEndpointServiceCustomFilterDataSourceConfig(rName string) string {
	return testAccVPCEndpointServiceCustomBaseDataSourceConfig(rName) + `
data "aws_vpc_endpoint_service" "test" {
  filter {
    name   = "service-name"
    values = [aws_vpc_endpoint_service.test.service_name]
  }
}
`
}

func testAccVPCEndpointServiceCustomFilterTagsDataSourceConfig(rName string) string {
	return testAccVPCEndpointServiceCustomBaseDataSourceConfig(rName) + `
data "aws_vpc_endpoint_service" "test" {
  tags = {
    Name = aws_vpc_endpoint_service.test.tags["Name"]
  }
}
`
}
