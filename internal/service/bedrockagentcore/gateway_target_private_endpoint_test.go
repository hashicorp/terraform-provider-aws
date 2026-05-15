// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package bedrockagentcore_test

import (
	"fmt"
	"testing"

	bedrockagentcorecontrol "github.com/aws/aws-sdk-go-v2/service/bedrockagentcorecontrol"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// TestAccBedrockAgentCoreGatewayTarget_privateEndpointManagedVPC verifies that a
// gateway target can be created with a managed VPC Lattice private endpoint.
func TestAccBedrockAgentCoreGatewayTarget_privateEndpointManagedVPC(t *testing.T) {
	ctx := acctest.Context(t)
	var gatewayTarget bedrockagentcorecontrol.GetGatewayTargetOutput
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_bedrockagentcore_gateway_target.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.BedrockEndpointID)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.BedrockAgentCoreServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckGatewayTargetDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccGatewayTargetConfig_privateEndpointManagedVPC(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckGatewayTargetExists(ctx, t, resourceName, &gatewayTarget),
					resource.TestCheckResourceAttr(resourceName, "private_endpoint.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "private_endpoint.0.managed_vpc_resource.#", "1"),
					resource.TestCheckResourceAttrPair(resourceName, "private_endpoint.0.managed_vpc_resource.0.vpc_id", "aws_vpc.test", names.AttrID),
					resource.TestCheckResourceAttr(resourceName, "private_endpoint.0.managed_vpc_resource.0.endpoint_ip_address_type", "IPV4"),
					resource.TestCheckResourceAttr(resourceName, "private_endpoint.0.self_managed_lattice_resource.#", "0"),
				),
			},
			{
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateIdFunc:                    acctest.AttrsImportStateIdFunc(resourceName, ",", "gateway_identifier", "target_id"),
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: "target_id",
			},
		},
	})
}

// TestAccBedrockAgentCoreGatewayTarget_privateEndpointSelfManagedLattice verifies that a
// gateway target can be created with a self-managed VPC Lattice resource configuration.
// The target may reach FAILED status if the Lattice resource configuration endpoint is not
// reachable as an MCP server — this is expected in test environments and only verifies that
// the provider correctly sends the self_managed_lattice_resource variant to the API.
func TestAccBedrockAgentCoreGatewayTarget_privateEndpointSelfManagedLattice(t *testing.T) {
	ctx := acctest.Context(t)
	var gatewayTarget bedrockagentcorecontrol.GetGatewayTargetOutput
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_bedrockagentcore_gateway_target.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.BedrockEndpointID)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.BedrockAgentCoreServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckGatewayTargetDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccGatewayTargetConfig_privateEndpointSelfManagedLattice(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckGatewayTargetExists(ctx, t, resourceName, &gatewayTarget),
					resource.TestCheckResourceAttr(resourceName, "private_endpoint.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "private_endpoint.0.self_managed_lattice_resource.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "private_endpoint.0.managed_vpc_resource.#", "0"),
				),
			},
		},
	})
}

// TestAccBedrockAgentCoreGatewayTarget_privateEndpointWithRoutingDomain verifies that
// the optional routing_domain field is correctly stored and read back.
func TestAccBedrockAgentCoreGatewayTarget_privateEndpointWithRoutingDomain(t *testing.T) {
	ctx := acctest.Context(t)
	var gatewayTarget bedrockagentcorecontrol.GetGatewayTargetOutput
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_bedrockagentcore_gateway_target.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.BedrockEndpointID)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.BedrockAgentCoreServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckGatewayTargetDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccGatewayTargetConfig_privateEndpointWithRoutingDomain(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckGatewayTargetExists(ctx, t, resourceName, &gatewayTarget),
					resource.TestCheckResourceAttr(resourceName, "private_endpoint.0.managed_vpc_resource.0.routing_domain", "my-alb.internal.example.com"),
				),
			},
		},
	})
}

// ---------------------------------------------------------------------------
// Config helpers
// ---------------------------------------------------------------------------

func testAccGatewayTargetConfig_privateEndpointManagedVPC(rName string) string {
	return acctest.ConfigCompose(
		testAccGatewayTargetConfig_infra(rName),
		testAccVPCConfig(rName),
		fmt.Sprintf(`
resource "aws_bedrockagentcore_gateway_target" "test" {
  gateway_identifier = aws_bedrockagentcore_gateway.test.gateway_id
  name               = %[1]q

  target_configuration {
    mcp {
      mcp_server {
        endpoint = "https://mcp.internal.example.com/mcp"
      }
    }
  }

  private_endpoint {
    managed_vpc_resource {
      vpc_id                   = aws_vpc.test.id
      subnet_ids               = [aws_subnet.test.id]
      endpoint_ip_address_type = "IPV4"
      security_group_ids       = [aws_security_group.test.id]
    }
  }
}
`, rName),
	)
}

func testAccGatewayTargetConfig_privateEndpointSelfManagedLattice(rName string) string {
	return acctest.ConfigCompose(
		testAccGatewayTargetConfig_infra(rName),
		testAccVPCLatticeResourceConfigConfig(rName),
		fmt.Sprintf(`
resource "aws_bedrockagentcore_gateway_target" "test" {
  gateway_identifier = aws_bedrockagentcore_gateway.test.gateway_id
  name               = %[1]q

  target_configuration {
    mcp {
      mcp_server {
        endpoint = "https://mcp.internal.example.com/mcp"
      }
    }
  }

  private_endpoint {
    self_managed_lattice_resource {
      resource_configuration_identifier = aws_vpclattice_resource_configuration.test.arn
    }
  }
}
`, rName),
	)
}

func testAccGatewayTargetConfig_privateEndpointWithRoutingDomain(rName string) string {
	return acctest.ConfigCompose(
		testAccGatewayTargetConfig_infra(rName),
		testAccVPCConfig(rName),
		fmt.Sprintf(`
resource "aws_bedrockagentcore_gateway_target" "test" {
  gateway_identifier = aws_bedrockagentcore_gateway.test.gateway_id
  name               = %[1]q

  target_configuration {
    mcp {
      mcp_server {
        endpoint = "https://mcp.internal.example.com/mcp"
      }
    }
  }

  private_endpoint {
    managed_vpc_resource {
      vpc_id                   = aws_vpc.test.id
      subnet_ids               = [aws_subnet.test.id]
      endpoint_ip_address_type = "IPV4"
      routing_domain           = "my-alb.internal.example.com"
    }
  }
}
`, rName),
	)
}

func testAccVPCConfig(rName string) string {
	return fmt.Sprintf(`
resource "aws_vpc" "test" {
  cidr_block = "10.0.0.0/16"

  tags = {
    Name = %[1]q
  }
}

resource "aws_subnet" "test" {
  vpc_id            = aws_vpc.test.id
  cidr_block        = "10.0.1.0/24"
  availability_zone = data.aws_availability_zones.available.names[0]

  tags = {
    Name = %[1]q
  }
}

resource "aws_security_group" "test" {
  name   = %[1]q
  vpc_id = aws_vpc.test.id
}

data "aws_availability_zones" "available" {
  state = "available"
}
`, rName)
}

func testAccVPCLatticeResourceConfigConfig(rName string) string {
	return acctest.ConfigCompose(
		testAccVPCConfig(rName),
		fmt.Sprintf(`
resource "aws_vpclattice_resource_configuration" "test" {
  name = %[1]q
  type = "SINGLE"

  resource_gateway_identifier = aws_vpclattice_resource_gateway.test.id

  port_ranges = ["443"]
  protocol    = "TCP"

  resource_configuration_definition {
    ip_resource {
      ip_address = "10.0.1.100"
    }
  }
}

resource "aws_vpclattice_resource_gateway" "test" {
  name       = %[1]q
  vpc_id     = aws_vpc.test.id
  subnet_ids = [aws_subnet.test.id]
}
`, rName),
	)
}
