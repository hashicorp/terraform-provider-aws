// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package route53resolver_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/service/route53resolver"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfroute53resolver "github.com/hashicorp/terraform-provider-aws/internal/service/route53resolver"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccRoute53ResolverEndpoint_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var ep route53resolver.ResolverEndpoint
	resourceName := "aws_route53_resolver_endpoint.test"
	vpcResourceName := "aws_vpc.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.Route53ResolverServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckEndpointDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccEndpointConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckEndpointExists(ctx, resourceName, &ep),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, "direction", "INBOUND"),
					resource.TestCheckResourceAttrPair(resourceName, "host_vpc_id", vpcResourceName, names.AttrID),
					resource.TestCheckResourceAttr(resourceName, "ip_address.#", acctest.Ct3),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, ""),
					resource.TestCheckResourceAttr(resourceName, "protocols.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "resolver_endpoint_type", "IPV4"),
					resource.TestCheckResourceAttr(resourceName, "security_group_ids.#", acctest.Ct2),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct0),
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

func TestAccRoute53ResolverEndpoint_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	var ep route53resolver.ResolverEndpoint
	resourceName := "aws_route53_resolver_endpoint.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.Route53ResolverServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckEndpointDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccEndpointConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckEndpointExists(ctx, resourceName, &ep),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfroute53resolver.ResourceEndpoint(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccRoute53ResolverEndpoint_tags(t *testing.T) {
	ctx := acctest.Context(t)
	var ep route53resolver.ResolverEndpoint
	resourceName := "aws_route53_resolver_endpoint.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.Route53ResolverServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckEndpointDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccEndpointConfig_tags1(rName, acctest.CtKey1, acctest.CtValue1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckEndpointExists(ctx, resourceName, &ep),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey1, acctest.CtValue1),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccEndpointConfig_tags2(rName, acctest.CtKey1, acctest.CtValue1Updated, acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckEndpointExists(ctx, resourceName, &ep),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct2),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey1, acctest.CtValue1Updated),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey2, acctest.CtValue2),
				),
			},
			{
				Config: testAccEndpointConfig_tags1(rName, acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckEndpointExists(ctx, resourceName, &ep),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey2, acctest.CtValue2),
				),
			},
		},
	})
}

func TestAccRoute53ResolverEndpoint_updateOutbound(t *testing.T) {
	ctx := acctest.Context(t)
	var ep route53resolver.ResolverEndpoint
	resourceName := "aws_route53_resolver_endpoint.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	initialName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	updatedName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.Route53ResolverServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckEndpointDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccEndpointConfig_outbound(rName, initialName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckEndpointExists(ctx, resourceName, &ep),
					resource.TestCheckResourceAttr(resourceName, "direction", "OUTBOUND"),
					resource.TestCheckResourceAttr(resourceName, "ip_address.#", acctest.Ct2),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, initialName),
					resource.TestCheckResourceAttr(resourceName, "protocols.#", acctest.Ct1),
				),
			},
			{
				Config: testAccEndpointConfig_updatedOutbound(rName, updatedName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckEndpointExists(ctx, resourceName, &ep),
					resource.TestCheckResourceAttr(resourceName, "direction", "OUTBOUND"),
					resource.TestCheckResourceAttr(resourceName, "ip_address.#", acctest.Ct3),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, updatedName),
					resource.TestCheckResourceAttr(resourceName, "protocols.#", acctest.Ct2),
				),
			},
		},
	})
}

func TestAccRoute53ResolverEndpoint_resolverEndpointType(t *testing.T) {
	ctx := acctest.Context(t)
	var ep route53resolver.ResolverEndpoint
	resourceName := "aws_route53_resolver_endpoint.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.Route53ResolverServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckEndpointDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccEndpointConfig_resolverEndpointType(rName, "DUALSTACK"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckEndpointExists(ctx, resourceName, &ep),
					resource.TestCheckResourceAttr(resourceName, "resolver_endpoint_type", "DUALSTACK"),
				),
			},
			{
				Config: testAccEndpointConfig_resolverEndpointType(rName, "IPV4"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckEndpointExists(ctx, resourceName, &ep),
					resource.TestCheckResourceAttr(resourceName, "resolver_endpoint_type", "IPV4"),
				),
			},
		},
	})
}

func testAccCheckEndpointDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).Route53ResolverConn(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_route53_resolver_endpoint" {
				continue
			}

			_, err := tfroute53resolver.FindResolverEndpointByID(ctx, conn, rs.Primary.ID)

			if tfresource.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("Route53 Resolver Endpoint still exists: %s", rs.Primary.ID)
		}

		return nil
	}
}

func testAccCheckEndpointExists(ctx context.Context, n string, v *route53resolver.ResolverEndpoint) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No Route53 Resolver Endpoint ID is set")
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).Route53ResolverConn(ctx)

		output, err := tfroute53resolver.FindResolverEndpointByID(ctx, conn, rs.Primary.ID)

		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

func testAccPreCheck(ctx context.Context, t *testing.T) {
	conn := acctest.Provider.Meta().(*conns.AWSClient).Route53ResolverConn(ctx)

	input := &route53resolver.ListResolverEndpointsInput{}

	_, err := conn.ListResolverEndpointsWithContext(ctx, input)

	if acctest.PreCheckSkipError(err) {
		t.Skipf("skipping acceptance testing: %s", err)
	}

	if err != nil {
		t.Fatalf("unexpected PreCheck error: %s", err)
	}
}

func testAccEndpointConfig_base(rName string) string {
	return acctest.ConfigCompose(acctest.ConfigAvailableAZsNoOptIn(), fmt.Sprintf(`
resource "aws_vpc" "test" {
  cidr_block           = "10.0.0.0/16"
  enable_dns_support   = true
  enable_dns_hostnames = true

  tags = {
    Name = %[1]q
  }
}

resource "aws_subnet" "test" {
  count = 3

  vpc_id            = aws_vpc.test.id
  availability_zone = data.aws_availability_zones.available.names[count.index]
  cidr_block        = cidrsubnet(aws_vpc.test.cidr_block, 8, count.index)

  tags = {
    Name = %[1]q
  }
}

resource "aws_security_group" "test" {
  count = 2

  vpc_id = aws_vpc.test.id
  name   = "%[1]s-${count.index}"

  tags = {
    Name = %[1]q
  }
}
`, rName))
}

func testAccEndpointConfig_basic(rName string) string {
	return acctest.ConfigCompose(testAccEndpointConfig_base(rName), `
resource "aws_route53_resolver_endpoint" "test" {
  direction = "INBOUND"

  security_group_ids = aws_security_group.test[*].id

  ip_address {
    subnet_id = aws_subnet.test[0].id
  }

  ip_address {
    subnet_id = aws_subnet.test[1].id
  }

  ip_address {
    subnet_id = aws_subnet.test[2].id
  }
}
`)
}

func testAccEndpointConfig_tags1(rName, tagKey1, tagValue1 string) string {
	return acctest.ConfigCompose(testAccEndpointConfig_base(rName), fmt.Sprintf(`
resource "aws_route53_resolver_endpoint" "test" {
  direction = "INBOUND"

  security_group_ids = aws_security_group.test[*].id

  ip_address {
    subnet_id = aws_subnet.test[0].id
  }

  ip_address {
    subnet_id = aws_subnet.test[1].id
  }

  ip_address {
    subnet_id = aws_subnet.test[2].id
  }

  tags = {
    %[1]q = %[2]q
  }
}
`, tagKey1, tagValue1))
}

func testAccEndpointConfig_tags2(rName, tagKey1, tagValue1, tagKey2, tagValue2 string) string {
	return acctest.ConfigCompose(testAccEndpointConfig_base(rName), fmt.Sprintf(`
resource "aws_route53_resolver_endpoint" "test" {
  direction = "INBOUND"

  security_group_ids = aws_security_group.test[*].id

  ip_address {
    subnet_id = aws_subnet.test[0].id
  }

  ip_address {
    subnet_id = aws_subnet.test[1].id
  }

  ip_address {
    subnet_id = aws_subnet.test[2].id
  }

  tags = {
    %[1]q = %[2]q
    %[3]q = %[4]q
  }
}
`, tagKey1, tagValue1, tagKey2, tagValue2))
}

func testAccEndpointConfig_outbound(rName, name string) string {
	return acctest.ConfigCompose(testAccEndpointConfig_base(rName), fmt.Sprintf(`
resource "aws_route53_resolver_endpoint" "test" {
  direction = "OUTBOUND"
  name      = %[1]q

  security_group_ids = aws_security_group.test[*].id

  ip_address {
    subnet_id = aws_subnet.test[0].id
  }

  ip_address {
    subnet_id = aws_subnet.test[1].id
    ip        = cidrhost(aws_subnet.test[1].cidr_block, 8)
  }
}
`, name))
}

func testAccEndpointConfig_updatedOutbound(rName, name string) string {
	return acctest.ConfigCompose(testAccEndpointConfig_base(rName), fmt.Sprintf(`
resource "aws_route53_resolver_endpoint" "test" {
  direction = "OUTBOUND"
  name      = %[1]q

  security_group_ids = aws_security_group.test[*].id

  ip_address {
    subnet_id = aws_subnet.test[2].id
  }

  ip_address {
    subnet_id = aws_subnet.test[1].id
  }

  ip_address {
    subnet_id = aws_subnet.test[0].id
  }

  protocols = ["Do53", "DoH"]
}
`, name))
}

func testAccEndpointConfig_resolverEndpointType(rName, resolverEndpointType string) string {
	return acctest.ConfigCompose(acctest.ConfigAvailableAZsNoOptIn(), fmt.Sprintf(`
resource "aws_vpc" "test" {
  cidr_block           = "10.0.0.0/16"
  enable_dns_support   = true
  enable_dns_hostnames = true

  assign_generated_ipv6_cidr_block = true

  tags = {
    Name = %[1]q
  }
}

resource "aws_subnet" "test" {
  count = 3

  vpc_id            = aws_vpc.test.id
  availability_zone = data.aws_availability_zones.available.names[count.index]
  cidr_block        = cidrsubnet(aws_vpc.test.cidr_block, 8, count.index)

  ipv6_cidr_block                 = cidrsubnet(aws_vpc.test.ipv6_cidr_block, 8, count.index)
  assign_ipv6_address_on_creation = true

  tags = {
    Name = %[1]q
  }
}

resource "aws_security_group" "test" {
  count = 2

  vpc_id = aws_vpc.test.id
  name   = "%[1]s-${count.index}"

  tags = {
    Name = %[1]q
  }
}

resource "aws_route53_resolver_endpoint" "test" {
  direction = "INBOUND"

  security_group_ids = aws_security_group.test[*].id

  ip_address {
    subnet_id = aws_subnet.test[0].id
  }

  ip_address {
    subnet_id = aws_subnet.test[1].id
  }

  ip_address {
    subnet_id = aws_subnet.test[2].id
  }

  resolver_endpoint_type = %[2]q
}
`, rName, resolverEndpointType))
}
