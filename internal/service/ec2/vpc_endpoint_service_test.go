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
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfec2 "github.com/hashicorp/terraform-provider-aws/internal/service/ec2"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

func TestAccVPCEndpointService_basic(t *testing.T) {
	var svcCfg ec2.ServiceConfiguration
	resourceName := "aws_vpc_endpoint_service.test"
	rName := sdkacctest.RandomWithPrefix("tfacctest") // 32 character limit

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ec2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckVPCEndpointServiceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccVPCEndpointServiceConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckVPCEndpointServiceExists(resourceName, &svcCfg),
					resource.TestCheckResourceAttr(resourceName, "acceptance_required", "false"),
					resource.TestCheckResourceAttr(resourceName, "allowed_principals.#", "0"),
					acctest.MatchResourceAttrRegionalARN(resourceName, "arn", "ec2", regexp.MustCompile(`vpc-endpoint-service/vpce-svc-.+`)),
					acctest.CheckResourceAttrGreaterThanValue(resourceName, "availability_zones.#", "0"),
					acctest.CheckResourceAttrGreaterThanValue(resourceName, "base_endpoint_dns_names.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "gateway_load_balancer_arns.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "manages_vpc_endpoints", "false"),
					resource.TestCheckResourceAttr(resourceName, "network_load_balancer_arns.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "private_dns_name", ""),
					resource.TestCheckResourceAttr(resourceName, "private_dns_name_configuration.#", "0"),
					resource.TestCheckResourceAttrSet(resourceName, "service_name"),
					resource.TestCheckResourceAttr(resourceName, "service_type", "Interface"),
					resource.TestCheckResourceAttr(resourceName, "supported_ip_address_types.#", "1"),
					resource.TestCheckTypeSetElemAttr(resourceName, "supported_ip_address_types.*", "ipv4"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccVPCEndpointService_disappears(t *testing.T) {
	var svcCfg ec2.ServiceConfiguration
	resourceName := "aws_vpc_endpoint_service.test"
	rName := sdkacctest.RandomWithPrefix("tfacctest") // 32 character limit

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ec2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckVPCEndpointServiceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccVPCEndpointServiceConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVPCEndpointServiceExists(resourceName, &svcCfg),
					acctest.CheckResourceDisappears(acctest.Provider, tfec2.ResourceVPCEndpointService(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccVPCEndpointService_tags(t *testing.T) {
	var svcCfg ec2.ServiceConfiguration
	resourceName := "aws_vpc_endpoint_service.test"
	rName := sdkacctest.RandomWithPrefix("tfacctest") // 32 character limit

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ec2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckVPCEndpointServiceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccVPCEndpointServiceConfig_tags1(rName, "key1", "value1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVPCEndpointServiceExists(resourceName, &svcCfg),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccVPCEndpointServiceConfig_tags2(rName, "key1", "value1updated", "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVPCEndpointServiceExists(resourceName, &svcCfg),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1updated"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
			{
				Config: testAccVPCEndpointServiceConfig_tags1(rName, "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVPCEndpointServiceExists(resourceName, &svcCfg),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
		},
	})
}

func TestAccVPCEndpointService_networkLoadBalancerARNs(t *testing.T) {
	var svcCfg ec2.ServiceConfiguration
	resourceName := "aws_vpc_endpoint_service.test"
	rName := sdkacctest.RandomWithPrefix("tfacctest") // 32 character limit

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ec2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckVPCEndpointServiceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccVPCEndpointServiceConfig_networkLoadBalancerARNs(rName, 1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVPCEndpointServiceExists(resourceName, &svcCfg),
					resource.TestCheckResourceAttr(resourceName, "network_load_balancer_arns.#", "1"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccVPCEndpointServiceConfig_networkLoadBalancerARNs(rName, 2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVPCEndpointServiceExists(resourceName, &svcCfg),
					resource.TestCheckResourceAttr(resourceName, "network_load_balancer_arns.#", "2"),
				),
			},
		},
	})
}

func TestAccVPCEndpointService_supportedIPAddressTypes(t *testing.T) {
	var svcCfg ec2.ServiceConfiguration
	resourceName := "aws_vpc_endpoint_service.test"
	rName := sdkacctest.RandomWithPrefix("tfacctest") // 32 character limit

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ec2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckVPCEndpointServiceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccVPCEndpointServiceConfig_supportedIPAddressTypesIPv4(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVPCEndpointServiceExists(resourceName, &svcCfg),
					resource.TestCheckResourceAttr(resourceName, "supported_ip_address_types.#", "1"),
					resource.TestCheckTypeSetElemAttr(resourceName, "supported_ip_address_types.*", "ipv4"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccVPCEndpointServiceConfig_supportedIPAddressTypesIPv4AndIPv6(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVPCEndpointServiceExists(resourceName, &svcCfg),
					resource.TestCheckResourceAttr(resourceName, "supported_ip_address_types.#", "2"),
					resource.TestCheckTypeSetElemAttr(resourceName, "supported_ip_address_types.*", "ipv4"),
					resource.TestCheckTypeSetElemAttr(resourceName, "supported_ip_address_types.*", "ipv6"),
				),
			},
		},
	})
}

func TestAccVPCEndpointService_allowedPrincipals(t *testing.T) {
	var svcCfg ec2.ServiceConfiguration
	resourceName := "aws_vpc_endpoint_service.test"
	rName := sdkacctest.RandomWithPrefix("tfacctest") // 32 character limit

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ec2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckVPCEndpointServiceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccVPCEndpointServiceConfig_allowedPrincipals(rName, 1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVPCEndpointServiceExists(resourceName, &svcCfg),
					resource.TestCheckResourceAttr(resourceName, "allowed_principals.#", "1"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccVPCEndpointServiceConfig_allowedPrincipals(rName, 0),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVPCEndpointServiceExists(resourceName, &svcCfg),
					resource.TestCheckResourceAttr(resourceName, "allowed_principals.#", "0"),
				),
			},
			{
				Config: testAccVPCEndpointServiceConfig_allowedPrincipals(rName, 1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVPCEndpointServiceExists(resourceName, &svcCfg),
					resource.TestCheckResourceAttr(resourceName, "allowed_principals.#", "1"),
				),
			},
		},
	})
}

func TestAccVPCEndpointService_gatewayLoadBalancerARNs(t *testing.T) {
	var svcCfg ec2.ServiceConfiguration
	resourceName := "aws_vpc_endpoint_service.test"
	rName := sdkacctest.RandomWithPrefix("tfacctest") // 32 character limit

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheckELBv2GatewayLoadBalancer(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ec2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckVPCEndpointServiceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccVPCEndpointServiceConfig_gatewayLoadBalancerARNs(rName, 1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVPCEndpointServiceExists(resourceName, &svcCfg),
					resource.TestCheckResourceAttr(resourceName, "gateway_load_balancer_arns.#", "1"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccVPCEndpointServiceConfig_gatewayLoadBalancerARNs(rName, 2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVPCEndpointServiceExists(resourceName, &svcCfg),
					resource.TestCheckResourceAttr(resourceName, "gateway_load_balancer_arns.#", "2"),
				),
			},
		},
	})
}

func TestAccVPCEndpointService_privateDNSName(t *testing.T) {
	var svcCfg ec2.ServiceConfiguration
	resourceName := "aws_vpc_endpoint_service.test"
	rName := sdkacctest.RandomWithPrefix("tfacctest") // 32 character limit
	domainName1 := acctest.RandomSubdomain()
	domainName2 := acctest.RandomSubdomain()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ec2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckVPCEndpointServiceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccVPCEndpointServiceConfig_privateDNSName(rName, domainName1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVPCEndpointServiceExists(resourceName, &svcCfg),
					resource.TestCheckResourceAttr(resourceName, "private_dns_name", domainName1),
					resource.TestCheckResourceAttr(resourceName, "private_dns_name_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "private_dns_name_configuration.0.type", "TXT"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccVPCEndpointServiceConfig_privateDNSName(rName, domainName2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVPCEndpointServiceExists(resourceName, &svcCfg),
					resource.TestCheckResourceAttr(resourceName, "private_dns_name", domainName2),
					resource.TestCheckResourceAttr(resourceName, "private_dns_name_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "private_dns_name_configuration.0.type", "TXT"),
				),
			},
		},
	})
}

func testAccCheckVPCEndpointServiceDestroy(s *terraform.State) error {
	conn := acctest.Provider.Meta().(*conns.AWSClient).EC2Conn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_vpc_endpoint_service" {
			continue
		}

		_, err := tfec2.FindVPCEndpointServiceConfigurationByID(conn, rs.Primary.ID)

		if tfresource.NotFound(err) {
			continue
		}

		if err != nil {
			return err
		}

		return fmt.Errorf("EC2 VPC Endpoint Service %s still exists", rs.Primary.ID)
	}

	return nil
}

func testAccCheckVPCEndpointServiceExists(n string, v *ec2.ServiceConfiguration) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No EC2 VPC Endpoint Service ID is set")
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).EC2Conn

		output, err := tfec2.FindVPCEndpointServiceConfigurationByID(conn, rs.Primary.ID)

		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

func testAccVPCEndpointServiceConfig_networkLoadBalancerBase(rName string, count int) string {
	return acctest.ConfigCompose(acctest.ConfigAvailableAZsNoOptIn(), fmt.Sprintf(`
resource "aws_vpc" "test" {
  cidr_block = "10.0.0.0/16"

  tags = {
    Name = %[1]q
  }
}

resource "aws_subnet" "test" {
  count = 2

  vpc_id            = aws_vpc.test.id
  cidr_block        = cidrsubnet(aws_vpc.test.cidr_block, 8, count.index)
  availability_zone = data.aws_availability_zones.available.names[count.index]

  tags = {
    Name = %[1]q
  }
}

resource "aws_lb" "test" {
  count = %[2]d

  load_balancer_type = "network"
  name               = "%[1]s-${count.index}"

  subnets = aws_subnet.test[*].id

  internal                   = true
  idle_timeout               = 60
  enable_deletion_protection = false

  tags = {
    Name = %[1]q
  }
}
`, rName, count))
}

func testAccVPCEndpointServiceConfig_supportedIPAddressTypesBase(rName string) string {
	return acctest.ConfigCompose(acctest.ConfigAvailableAZsNoOptIn(), fmt.Sprintf(`
resource "aws_vpc" "test" {
  cidr_block                       = "10.0.0.0/16"
  assign_generated_ipv6_cidr_block = true

  tags = {
    Name = %[1]q
  }
}

resource "aws_subnet" "test" {
  count = 2

  vpc_id            = aws_vpc.test.id
  cidr_block        = cidrsubnet(aws_vpc.test.cidr_block, 8, count.index)
  availability_zone = data.aws_availability_zones.available.names[count.index]
  ipv6_cidr_block   = cidrsubnet(aws_vpc.test.ipv6_cidr_block, 8, count.index)

  tags = {
    Name = %[1]q
  }
}

resource "aws_lb" "test" {
  load_balancer_type = "network"
  name               = %[1]q

  subnets = aws_subnet.test[*].id

  internal                   = true
  idle_timeout               = 60
  enable_deletion_protection = false

  ip_address_type = "dualstack"

  tags = {
    Name = %[1]q
  }
}
`, rName))
}

func testAccVPCEndpointServiceConfig_basic(rName string) string {
	return acctest.ConfigCompose(testAccVPCEndpointServiceConfig_networkLoadBalancerBase(rName, 1), `
resource "aws_vpc_endpoint_service" "test" {
  acceptance_required        = false
  network_load_balancer_arns = aws_lb.test[*].arn
}
`)
}

func testAccVPCEndpointServiceConfig_gatewayLoadBalancerARNs(rName string, count int) string {
	return acctest.ConfigCompose(acctest.ConfigAvailableAZsNoOptIn(), fmt.Sprintf(`
resource "aws_vpc" "test" {
  cidr_block = "10.10.10.0/25"

  tags = {
    Name = %[1]q
  }
}

resource "aws_subnet" "test" {
  availability_zone = data.aws_availability_zones.available.names[0]
  cidr_block        = cidrsubnet(aws_vpc.test.cidr_block, 2, 0)
  vpc_id            = aws_vpc.test.id

  tags = {
    Name = %[1]q
  }
}

resource "aws_lb" "test" {
  count = %[2]d

  load_balancer_type = "gateway"
  name               = "%[1]s-${count.index}"

  subnet_mapping {
    subnet_id = aws_subnet.test.id
  }
}

resource "aws_vpc_endpoint_service" "test" {
  acceptance_required        = false
  gateway_load_balancer_arns = aws_lb.test[*].arn

  tags = {
    Name = %[1]q
  }
}
`, rName, count))
}

func testAccVPCEndpointServiceConfig_networkLoadBalancerARNs(rName string, count int) string {
	return acctest.ConfigCompose(testAccVPCEndpointServiceConfig_networkLoadBalancerBase(rName, count), fmt.Sprintf(`
resource "aws_vpc_endpoint_service" "test" {
  acceptance_required        = false
  network_load_balancer_arns = aws_lb.test[*].arn

  tags = {
    Name = %[1]q
  }
}
`, rName))
}

func testAccVPCEndpointServiceConfig_supportedIPAddressTypesIPv4(rName string) string {
	return acctest.ConfigCompose(testAccVPCEndpointServiceConfig_supportedIPAddressTypesBase(rName), fmt.Sprintf(`
resource "aws_vpc_endpoint_service" "test" {
  acceptance_required        = false
  network_load_balancer_arns = aws_lb.test[*].arn
  supported_ip_address_types = ["ipv4"]

  tags = {
    Name = %[1]q
  }
}
`, rName))
}

func testAccVPCEndpointServiceConfig_supportedIPAddressTypesIPv4AndIPv6(rName string) string {
	return acctest.ConfigCompose(testAccVPCEndpointServiceConfig_supportedIPAddressTypesBase(rName), fmt.Sprintf(`
resource "aws_vpc_endpoint_service" "test" {
  acceptance_required        = false
  network_load_balancer_arns = aws_lb.test[*].arn
  supported_ip_address_types = ["ipv4", "ipv6"]

  tags = {
    Name = %[1]q
  }
}
`, rName))
}

func testAccVPCEndpointServiceConfig_allowedPrincipals(rName string, count int) string {
	return acctest.ConfigCompose(testAccVPCEndpointServiceConfig_networkLoadBalancerBase(rName, 1), fmt.Sprintf(`
data "aws_caller_identity" "current" {}

data "aws_iam_session_context" "current" {
  arn = data.aws_caller_identity.current.arn
}

resource "aws_vpc_endpoint_service" "test" {
  acceptance_required        = false
  network_load_balancer_arns = aws_lb.test[*].arn

  allowed_principals = (%[2]d == 0 ? [] : [data.aws_iam_session_context.current.issuer_arn])

  tags = {
    Name = %[1]q
  }
}
`, rName, count))
}

func testAccVPCEndpointServiceConfig_tags1(rName, tagKey1, tagValue1 string) string {
	return acctest.ConfigCompose(testAccVPCEndpointServiceConfig_networkLoadBalancerBase(rName, 1), fmt.Sprintf(`
resource "aws_vpc_endpoint_service" "test" {
  acceptance_required        = false
  network_load_balancer_arns = aws_lb.test[*].arn

  tags = {
    %[1]q = %[2]q
  }
}
`, tagKey1, tagValue1))
}

func testAccVPCEndpointServiceConfig_tags2(rName, tagKey1, tagValue1, tagKey2, tagValue2 string) string {
	return acctest.ConfigCompose(testAccVPCEndpointServiceConfig_networkLoadBalancerBase(rName, 1), fmt.Sprintf(`
resource "aws_vpc_endpoint_service" "test" {
  acceptance_required        = false
  network_load_balancer_arns = aws_lb.test[*].arn

  tags = {
    %[1]q = %[2]q
    %[3]q = %[4]q
  }
}
`, tagKey1, tagValue1, tagKey2, tagValue2))
}

func testAccVPCEndpointServiceConfig_privateDNSName(rName, dnsName string) string {
	return acctest.ConfigCompose(testAccVPCEndpointServiceConfig_networkLoadBalancerBase(rName, 1), fmt.Sprintf(`
resource "aws_vpc_endpoint_service" "test" {
  acceptance_required        = false
  network_load_balancer_arns = aws_lb.test[*].arn
  private_dns_name           = %[2]q

  tags = {
    Name = %[1]q
  }
}
`, rName, dnsName))
}
