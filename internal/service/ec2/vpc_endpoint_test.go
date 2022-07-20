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

func TestAccVPCEndpoint_gatewayBasic(t *testing.T) {
	var endpoint ec2.VpcEndpoint
	resourceName := "aws_vpc_endpoint.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ec2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckVPCEndpointDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccVPCEndpointConfig_gatewayBasic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckVPCEndpointExists(resourceName, &endpoint),
					acctest.MatchResourceAttrRegionalARN(resourceName, "arn", "ec2", regexp.MustCompile(`vpc-endpoint/vpce-.+`)),
					acctest.CheckResourceAttrGreaterThanValue(resourceName, "cidr_blocks.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "dns_entry.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "dns_options.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "ip_address_type", ""),
					resource.TestCheckResourceAttr(resourceName, "network_interface_ids.#", "0"),
					acctest.CheckResourceAttrAccountID(resourceName, "owner_id"),
					resource.TestCheckResourceAttrSet(resourceName, "policy"),
					resource.TestCheckResourceAttrSet(resourceName, "prefix_list_id"),
					resource.TestCheckResourceAttr(resourceName, "private_dns_enabled", "false"),
					resource.TestCheckResourceAttr(resourceName, "requester_managed", "false"),
					resource.TestCheckResourceAttr(resourceName, "route_table_ids.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "security_group_ids.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "subnet_ids.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
					resource.TestCheckResourceAttr(resourceName, "vpc_endpoint_type", "Gateway"),
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

func TestAccVPCEndpoint_interfaceBasic(t *testing.T) {
	var endpoint ec2.VpcEndpoint
	resourceName := "aws_vpc_endpoint.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ec2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckVPCEndpointDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccVPCEndpointConfig_interfaceBasic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckVPCEndpointExists(resourceName, &endpoint),
					acctest.MatchResourceAttrRegionalARN(resourceName, "arn", "ec2", regexp.MustCompile(`vpc-endpoint/vpce-.+`)),
					resource.TestCheckResourceAttr(resourceName, "cidr_blocks.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "dns_entry.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "dns_options.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "dns_options.0.dns_record_ip_type", "ipv4"),
					resource.TestCheckResourceAttr(resourceName, "ip_address_type", "ipv4"),
					resource.TestCheckResourceAttr(resourceName, "network_interface_ids.#", "0"),
					acctest.CheckResourceAttrAccountID(resourceName, "owner_id"),
					resource.TestCheckResourceAttrSet(resourceName, "policy"),
					resource.TestCheckNoResourceAttr(resourceName, "prefix_list_id"),
					resource.TestCheckResourceAttr(resourceName, "private_dns_enabled", "false"),
					resource.TestCheckResourceAttr(resourceName, "requester_managed", "false"),
					resource.TestCheckResourceAttr(resourceName, "route_table_ids.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "security_group_ids.#", "1"), // Default SG.
					resource.TestCheckResourceAttr(resourceName, "subnet_ids.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
					resource.TestCheckResourceAttr(resourceName, "vpc_endpoint_type", "Interface"),
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

func TestAccVPCEndpoint_disappears(t *testing.T) {
	var endpoint ec2.VpcEndpoint
	resourceName := "aws_vpc_endpoint.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ec2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckVPCEndpointDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccVPCEndpointConfig_gatewayBasic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVPCEndpointExists(resourceName, &endpoint),
					acctest.CheckResourceDisappears(acctest.Provider, tfec2.ResourceVPCEndpoint(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccVPCEndpoint_tags(t *testing.T) {
	var endpoint ec2.VpcEndpoint
	resourceName := "aws_vpc_endpoint.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ec2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckVPCEndpointDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccVPCEndpointConfig_tags1(rName, "key1", "value1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVPCEndpointExists(resourceName, &endpoint),
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
				Config: testAccVPCEndpointConfig_tags2(rName, "key1", "value1updated", "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVPCEndpointExists(resourceName, &endpoint),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1updated"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
			{
				Config: testAccVPCEndpointConfig_tags1(rName, "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVPCEndpointExists(resourceName, &endpoint),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
		},
	})
}

func TestAccVPCEndpoint_gatewayWithRouteTableAndPolicy(t *testing.T) {
	var endpoint ec2.VpcEndpoint
	resourceName := "aws_vpc_endpoint.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ec2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckVPCEndpointDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccVPCEndpointConfig_gatewayRouteTableAndPolicy(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVPCEndpointExists(resourceName, &endpoint),
					resource.TestCheckResourceAttrSet(resourceName, "policy"),
					resource.TestCheckResourceAttr(resourceName, "route_table_ids.#", "1"),
				),
			},
			{
				Config: testAccVPCEndpointConfig_gatewayRouteTableAndPolicyModified(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVPCEndpointExists(resourceName, &endpoint),
					resource.TestCheckResourceAttrSet(resourceName, "policy"),
					resource.TestCheckResourceAttr(resourceName, "route_table_ids.#", "0"),
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

func TestAccVPCEndpoint_gatewayPolicy(t *testing.T) {
	var endpoint ec2.VpcEndpoint
	// This policy checks the DiffSuppressFunc
	policy1 := `
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Sid": "ReadOnly",
      "Principal": "*",
      "Action": [
        "dynamodb:DescribeTable",
        "dynamodb:ListTables"
      ],
      "Effect": "Allow",
      "Resource": "*"
    }
  ]
}
`

	policy2 := `
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Sid": "AllowAll",
      "Effect": "Allow",
      "Principal": {
        "AWS": "*"
      },
      "Action": "*",
      "Resource": "*"
    }
  ]
}
`

	resourceName := "aws_vpc_endpoint.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ec2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckVPCEndpointDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccVPCEndpointConfig_gatewayPolicy(rName, policy1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVPCEndpointExists(resourceName, &endpoint),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccVPCEndpointConfig_gatewayPolicy(rName, policy2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVPCEndpointExists(resourceName, &endpoint),
				),
			},
		},
	})
}

func TestAccVPCEndpoint_ignoreEquivalent(t *testing.T) {
	var endpoint ec2.VpcEndpoint
	resourceName := "aws_vpc_endpoint.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ec2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckVPCEndpointDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccVPCEndpointConfig_orderPolicy(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVPCEndpointExists(resourceName, &endpoint),
				),
			},
			{
				Config:   testAccVPCEndpointConfig_newOrderPolicy(rName),
				PlanOnly: true,
			},
		},
	})
}

func TestAccVPCEndpoint_ipAddressType(t *testing.T) {
	var endpoint ec2.VpcEndpoint
	resourceName := "aws_vpc_endpoint.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ec2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckVPCEndpointDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccVPCEndpointConfig_ipAddressType(rName, "ipv4"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVPCEndpointExists(resourceName, &endpoint),
					resource.TestCheckResourceAttr(resourceName, "dns_options.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "dns_options.0.dns_record_ip_type", "ipv4"),
					resource.TestCheckResourceAttr(resourceName, "ip_address_type", "ipv4"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"auto_accept"},
			},
			{
				Config: testAccVPCEndpointConfig_ipAddressType(rName, "dualstack"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVPCEndpointExists(resourceName, &endpoint),
					resource.TestCheckResourceAttr(resourceName, "dns_options.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "dns_options.0.dns_record_ip_type", "dualstack"),
					resource.TestCheckResourceAttr(resourceName, "ip_address_type", "dualstack"),
				),
			},
		},
	})
}

func TestAccVPCEndpoint_interfaceWithSubnetAndSecurityGroup(t *testing.T) {
	var endpoint ec2.VpcEndpoint
	resourceName := "aws_vpc_endpoint.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ec2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckVPCEndpointDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccVPCEndpointConfig_interfaceSubnet(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVPCEndpointExists(resourceName, &endpoint),
					resource.TestCheckResourceAttr(resourceName, "network_interface_ids.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "security_group_ids.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "subnet_ids.#", "1"),
				),
			},
			{
				Config: testAccVPCEndpointConfig_interfaceSubnetModified(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVPCEndpointExists(resourceName, &endpoint),
					resource.TestCheckResourceAttr(resourceName, "network_interface_ids.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "security_group_ids.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "subnet_ids.#", "3"),
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

func TestAccVPCEndpoint_interfaceNonAWSServiceAcceptOnCreate(t *testing.T) { // nosempgrep:aws-in-func-name
	var endpoint ec2.VpcEndpoint
	resourceName := "aws_vpc_endpoint.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ec2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckVPCEndpointDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccVPCEndpointConfig_interfaceNonAWSService(rName, true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVPCEndpointExists(resourceName, &endpoint),
					resource.TestCheckResourceAttr(resourceName, "state", "available"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"auto_accept"},
			},
		},
	})
}

func TestAccVPCEndpoint_interfaceNonAWSServiceAcceptOnUpdate(t *testing.T) { // nosempgrep:aws-in-func-name
	var endpoint ec2.VpcEndpoint
	resourceName := "aws_vpc_endpoint.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ec2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckVPCEndpointDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccVPCEndpointConfig_interfaceNonAWSService(rName, false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVPCEndpointExists(resourceName, &endpoint),
					resource.TestCheckResourceAttr(resourceName, "state", "pendingAcceptance"),
				),
			},
			{
				Config: testAccVPCEndpointConfig_interfaceNonAWSService(rName, true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVPCEndpointExists(resourceName, &endpoint),
					resource.TestCheckResourceAttr(resourceName, "state", "available"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"auto_accept"},
			},
		},
	})
}

func TestAccVPCEndpoint_VPCEndpointType_gatewayLoadBalancer(t *testing.T) {
	var endpoint ec2.VpcEndpoint
	vpcEndpointServiceResourceName := "aws_vpc_endpoint_service.test"
	resourceName := "aws_vpc_endpoint.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheckELBv2GatewayLoadBalancer(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ec2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckVPCEndpointDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccVPCEndpointConfig_gatewayLoadBalancer(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVPCEndpointExists(resourceName, &endpoint),
					resource.TestCheckResourceAttrPair(resourceName, "vpc_endpoint_type", vpcEndpointServiceResourceName, "service_type"),
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

func testAccCheckVPCEndpointDestroy(s *terraform.State) error {
	conn := acctest.Provider.Meta().(*conns.AWSClient).EC2Conn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_vpc_endpoint" {
			continue
		}

		_, err := tfec2.FindVPCEndpointByID(conn, rs.Primary.ID)

		if tfresource.NotFound(err) {
			continue
		}

		if err != nil {
			return err
		}
	}

	return nil
}

func testAccCheckVPCEndpointExists(n string, v *ec2.VpcEndpoint) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No EC2 VPC Endpoint ID is set")
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).EC2Conn

		output, err := tfec2.FindVPCEndpointByID(conn, rs.Primary.ID)

		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

func testAccVPCEndpointConfig_gatewayBasic(rName string) string {
	return fmt.Sprintf(`
resource "aws_vpc" "test" {
  cidr_block = "10.0.0.0/16"

  tags = {
    Name = %[1]q
  }
}

data "aws_region" "current" {}

resource "aws_vpc_endpoint" "test" {
  vpc_id       = aws_vpc.test.id
  service_name = "com.amazonaws.${data.aws_region.current.name}.s3"
}
`, rName)
}

func testAccVPCEndpointConfig_gatewayRouteTableAndPolicy(rName string) string {
	return fmt.Sprintf(`
resource "aws_vpc" "test" {
  cidr_block = "10.0.0.0/16"

  tags = {
    Name = %[1]q
  }
}

resource "aws_subnet" "test" {
  vpc_id     = aws_vpc.test.id
  cidr_block = "10.0.1.0/24"

  tags = {
    Name = %[1]q
  }
}

data "aws_region" "current" {}

resource "aws_vpc_endpoint" "test" {
  vpc_id       = aws_vpc.test.id
  service_name = "com.amazonaws.${data.aws_region.current.name}.s3"

  route_table_ids = [
    aws_route_table.test.id,
  ]

  policy = <<POLICY
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Sid": "AllowAll",
      "Effect": "Allow",
      "Principal": {
        "AWS": "*"
      },
      "Action": "*",
      "Resource": "*"
    }
  ]
}
POLICY

  tags = {
    Name = %[1]q
  }
}

resource "aws_route_table" "test" {
  vpc_id = aws_vpc.test.id

  tags = {
    Name = %[1]q
  }
}

resource "aws_route_table_association" "test" {
  subnet_id      = aws_subnet.test.id
  route_table_id = aws_route_table.test.id
}
`, rName)
}

func testAccVPCEndpointConfig_gatewayRouteTableAndPolicyModified(rName string) string {
	return fmt.Sprintf(`
resource "aws_vpc" "test" {
  cidr_block = "10.0.0.0/16"

  tags = {
    Name = %[1]q
  }
}

resource "aws_subnet" "test" {
  vpc_id     = aws_vpc.test.id
  cidr_block = "10.0.1.0/24"

  tags = {
    Name = %[1]q
  }
}

data "aws_region" "current" {}

resource "aws_vpc_endpoint" "test" {
  vpc_id       = aws_vpc.test.id
  service_name = "com.amazonaws.${data.aws_region.current.name}.s3"

  route_table_ids = []

  policy = ""

  tags = {
    Name = %[1]q
  }
}

resource "aws_internet_gateway" "test" {
  vpc_id = aws_vpc.test.id

  tags = {
    Name = %[1]q
  }
}

resource "aws_route_table" "test" {
  vpc_id = aws_vpc.test.id

  route {
    cidr_block = "0.0.0.0/0"
    gateway_id = aws_internet_gateway.test.id
  }

  tags = {
    Name = %[1]q
  }
}

resource "aws_route_table_association" "test" {
  subnet_id      = aws_subnet.test.id
  route_table_id = aws_route_table.test.id
}
`, rName)
}

func testAccVPCEndpointConfig_interfaceBasic(rName string) string {
	return fmt.Sprintf(`
resource "aws_vpc" "test" {
  cidr_block = "10.0.0.0/16"

  tags = {
    Name = %[1]q
  }
}

data "aws_region" "current" {}

resource "aws_vpc_endpoint" "test" {
  vpc_id            = aws_vpc.test.id
  service_name      = "com.amazonaws.${data.aws_region.current.name}.ec2"
  vpc_endpoint_type = "Interface"
}
`, rName)
}

func testAccVPCEndpointConfig_ipAddressType(rName, addressType string) string {
	return acctest.ConfigCompose(testAccVPCEndpointServiceConfig_supportedIPAddressTypesBase(rName), fmt.Sprintf(`
resource "aws_vpc_endpoint_service" "test" {
  acceptance_required        = false
  network_load_balancer_arns = aws_lb.test[*].arn
  supported_ip_address_types = ["ipv4", "ipv6"]

  tags = {
    Name = %[1]q
  }
}

resource "aws_vpc_endpoint" "test" {
  vpc_id              = aws_vpc.test.id
  service_name        = aws_vpc_endpoint_service.test.service_name
  vpc_endpoint_type   = "Interface"
  private_dns_enabled = false
  auto_accept         = true
  ip_address_type     = %[2]q

  dns_options {
    dns_record_ip_type = %[2]q
  }
}
`, rName, addressType))
}

func testAccVPCEndpointConfig_gatewayPolicy(rName, policy string) string {
	return fmt.Sprintf(`
data "aws_vpc_endpoint_service" "test" {
  service = "dynamodb"
}

resource "aws_vpc" "test" {
  cidr_block = "10.0.0.0/16"

  tags = {
    Name = %[1]q
  }
}

resource "aws_vpc_endpoint" "test" {
  policy       = <<POLICY
%[2]s
POLICY
  service_name = data.aws_vpc_endpoint_service.test.service_name
  vpc_id       = aws_vpc.test.id

  tags = {
    Name = %[1]q
  }
}
`, rName, policy)
}

func testAccVPCEndpointConfig_vpcBase(rName string) string {
	return acctest.ConfigCompose(acctest.ConfigAvailableAZsNoOptIn(), fmt.Sprintf(`
resource "aws_vpc" "test" {
  cidr_block           = "10.0.0.0/16"
  enable_dns_support   = true
  enable_dns_hostnames = true

  tags = {
    Name = %[1]q
  }
}

data "aws_region" "current" {}

resource "aws_subnet" "test" {
  count = 3

  vpc_id            = aws_vpc.test.id
  cidr_block        = cidrsubnet(aws_vpc.test.cidr_block, 2, count.index)
  availability_zone = data.aws_availability_zones.available.names[count.index]

  tags = {
    Name = %[1]q
  }
}

resource "aws_security_group" "test" {
  count = 2

  vpc_id = aws_vpc.test.id

  tags = {
    Name = %[1]q
  }
}
`, rName))
}

func testAccVPCEndpointConfig_interfaceSubnet(rName string) string {
	return acctest.ConfigCompose(
		testAccVPCEndpointConfig_vpcBase(rName),
		fmt.Sprintf(`
resource "aws_vpc_endpoint" "test" {
  vpc_id              = aws_vpc.test.id
  service_name        = "com.amazonaws.${data.aws_region.current.name}.ec2"
  vpc_endpoint_type   = "Interface"
  private_dns_enabled = false

  subnet_ids = [
    aws_subnet.test[0].id,
  ]

  security_group_ids = [
    aws_security_group.test[0].id,
    aws_security_group.test[1].id,
  ]

  tags = {
    Name = %[1]q
  }
}
`, rName))
}

func testAccVPCEndpointConfig_interfaceSubnetModified(rName string) string {
	return acctest.ConfigCompose(
		testAccVPCEndpointConfig_vpcBase(rName),
		fmt.Sprintf(`
resource "aws_vpc_endpoint" "test" {
  vpc_id              = aws_vpc.test.id
  service_name        = "com.amazonaws.${data.aws_region.current.name}.ec2"
  vpc_endpoint_type   = "Interface"
  private_dns_enabled = true

  subnet_ids = [
    aws_subnet.test[2].id,
    aws_subnet.test[1].id,
    aws_subnet.test[0].id,
  ]

  security_group_ids = [
    aws_security_group.test[1].id,
  ]

  tags = {
    Name = %[1]q
  }
}
`, rName))
}

func testAccVPCEndpointConfig_interfaceNonAWSService(rName string, autoAccept bool) string { // nosemgrep:aws-in-func-name
	return acctest.ConfigCompose(
		testAccVPCEndpointConfig_vpcBase(rName),
		fmt.Sprintf(`
resource "aws_lb" "test" {
  name = %[1]q

  subnets = [
    aws_subnet.test[0].id,
    aws_subnet.test[1].id,
  ]

  load_balancer_type         = "network"
  internal                   = true
  idle_timeout               = 60
  enable_deletion_protection = false

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

resource "aws_vpc_endpoint" "test" {
  vpc_id              = aws_vpc.test.id
  service_name        = aws_vpc_endpoint_service.test.service_name
  vpc_endpoint_type   = "Interface"
  private_dns_enabled = false
  auto_accept         = %[2]t

  security_group_ids = [
    aws_security_group.test[0].id,
  ]

  tags = {
    Name = %[1]q
  }
}
`, rName, autoAccept))
}

func testAccVPCEndpointConfig_tags1(rName, tagKey1, tagValue1 string) string {
	return fmt.Sprintf(`
resource "aws_vpc" "test" {
  cidr_block = "10.0.0.0/16"

  tags = {
    Name = %[1]q
  }
}

data "aws_region" "current" {}

resource "aws_vpc_endpoint" "test" {
  vpc_id       = aws_vpc.test.id
  service_name = "com.amazonaws.${data.aws_region.current.name}.s3"

  tags = {
    %[2]q = %[3]q
  }
}
`, rName, tagKey1, tagValue1)
}

func testAccVPCEndpointConfig_tags2(rName, tagKey1, tagValue1, tagKey2, tagValue2 string) string {
	return fmt.Sprintf(`
resource "aws_vpc" "test" {
  cidr_block = "10.0.0.0/16"

  tags = {
    Name = %[1]q
  }
}

data "aws_region" "current" {}

resource "aws_vpc_endpoint" "test" {
  vpc_id       = aws_vpc.test.id
  service_name = "com.amazonaws.${data.aws_region.current.name}.s3"

  tags = {
    %[2]q = %[3]q
    %[4]q = %[5]q
  }
}
`, rName, tagKey1, tagValue1, tagKey2, tagValue2)
}

func testAccVPCEndpointConfig_gatewayLoadBalancer(rName string) string {
	return acctest.ConfigCompose(acctest.ConfigAvailableAZsNoOptIn(), fmt.Sprintf(`
data "aws_caller_identity" "current" {}

data "aws_iam_session_context" "current" {
  arn = data.aws_caller_identity.current.arn
}

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
  load_balancer_type = "gateway"
  name               = %[1]q

  subnet_mapping {
    subnet_id = aws_subnet.test.id
  }

  tags = {
    Name = %[1]q
  }
}

resource "aws_vpc_endpoint_service" "test" {
  acceptance_required        = false
  allowed_principals         = [data.aws_iam_session_context.current.issuer_arn]
  gateway_load_balancer_arns = [aws_lb.test.arn]

  tags = {
    Name = %[1]q
  }
}

resource "aws_vpc_endpoint" "test" {
  service_name      = aws_vpc_endpoint_service.test.service_name
  subnet_ids        = [aws_subnet.test.id]
  vpc_endpoint_type = aws_vpc_endpoint_service.test.service_type
  vpc_id            = aws_vpc.test.id

  tags = {
    Name = %[1]q
  }
}
`, rName))
}

func testAccVPCEndpointConfig_orderPolicy(rName string) string {
	return fmt.Sprintf(`
data "aws_vpc_endpoint_service" "test" {
  service = "dynamodb"
}

resource "aws_vpc" "test" {
  cidr_block = "10.0.0.0/16"

  tags = {
    Name = %[1]q
  }
}

resource "aws_vpc_endpoint" "test" {
  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [{
      Sid       = "ReadOnly"
      Principal = "*"
      Action = [
        "dynamodb:DescribeTable",
        "dynamodb:ListTables",
        "dynamodb:ListTagsOfResource",
      ]
      Effect   = "Allow"
      Resource = "*"
    }]
  })
  service_name = data.aws_vpc_endpoint_service.test.service_name
  vpc_id       = aws_vpc.test.id

  tags = {
    Name = %[1]q
  }
}
`, rName)
}

func testAccVPCEndpointConfig_newOrderPolicy(rName string) string {
	return fmt.Sprintf(`
data "aws_vpc_endpoint_service" "test" {
  service = "dynamodb"
}

resource "aws_vpc" "test" {
  cidr_block = "10.0.0.0/16"

  tags = {
    Name = %[1]q
  }
}

resource "aws_vpc_endpoint" "test" {
  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [{
      Sid       = "ReadOnly"
      Principal = "*"
      Action = [
        "dynamodb:ListTables",
        "dynamodb:ListTagsOfResource",
        "dynamodb:DescribeTable",
      ]
      Effect   = "Allow"
      Resource = "*"
    }]
  })
  service_name = data.aws_vpc_endpoint_service.test.service_name
  vpc_id       = aws_vpc.test.id

  tags = {
    Name = %[1]q
  }
}
`, rName)
}
