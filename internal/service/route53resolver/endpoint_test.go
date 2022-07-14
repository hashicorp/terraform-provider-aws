package route53resolver_test

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/route53resolver"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
)

func TestAccRoute53ResolverEndpoint_basicInbound(t *testing.T) {
	var ep route53resolver.ResolverEndpoint
	resourceName := "aws_route53_resolver_endpoint.foo"
	rInt := sdkacctest.RandInt()
	name := fmt.Sprintf("terraform-testacc-r53-resolver-%d", rInt)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, route53resolver.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckEndpointDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccEndpointConfig_initial(rInt, "INBOUND", name),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckEndpointExists(resourceName, &ep),
					resource.TestCheckResourceAttr(resourceName, "name", name),
					resource.TestCheckResourceAttr(resourceName, "ip_address.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "security_group_ids.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.Usage", "original"),
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

func TestAccRoute53ResolverEndpoint_updateOutbound(t *testing.T) {
	var ep route53resolver.ResolverEndpoint
	resourceName := "aws_route53_resolver_endpoint.foo"
	rInt := sdkacctest.RandInt()
	initialName := fmt.Sprintf("terraform-testacc-r53-resolver-%d", rInt)
	updatedName := fmt.Sprintf("terraform-testacc-r53-rupdated-%d", rInt)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, route53resolver.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckEndpointDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccEndpointConfig_initial(rInt, "OUTBOUND", initialName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckEndpointExists(resourceName, &ep),
					resource.TestCheckResourceAttr(resourceName, "name", initialName),
					resource.TestCheckResourceAttr(resourceName, "ip_address.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "security_group_ids.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.Usage", "original"),
				),
			},
			{
				Config: testAccEndpointConfig_updated(rInt, "OUTBOUND", updatedName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckEndpointExists(resourceName, &ep),
					resource.TestCheckResourceAttr(resourceName, "name", updatedName),
					resource.TestCheckResourceAttr(resourceName, "ip_address.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "security_group_ids.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.Usage", "changed"),
				),
			},
		},
	})
}

func testAccCheckEndpointDestroy(s *terraform.State) error {
	conn := acctest.Provider.Meta().(*conns.AWSClient).Route53ResolverConn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_route53_resolver_endpoint" {
			continue
		}

		// Try to find the resource
		_, err := conn.GetResolverEndpoint(&route53resolver.GetResolverEndpointInput{
			ResolverEndpointId: aws.String(rs.Primary.ID),
		})
		// Verify the error is what we want
		if tfawserr.ErrCodeEquals(err, route53resolver.ErrCodeResourceNotFoundException) {
			continue
		}
		if err != nil {
			return err
		}
		return fmt.Errorf("Route 53 Resolver endpoint still exists: %s", rs.Primary.ID)
	}

	return nil
}

func testAccCheckEndpointExists(n string, ep *route53resolver.ResolverEndpoint) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No Route 53 Resolver endpoint ID is set")
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).Route53ResolverConn
		resp, err := conn.GetResolverEndpoint(&route53resolver.GetResolverEndpointInput{
			ResolverEndpointId: aws.String(rs.Primary.ID),
		})
		if err != nil {
			return err
		}

		*ep = *resp.ResolverEndpoint

		return nil
	}
}

func testAccPreCheck(t *testing.T) {
	conn := acctest.Provider.Meta().(*conns.AWSClient).Route53ResolverConn

	input := &route53resolver.ListResolverEndpointsInput{}

	_, err := conn.ListResolverEndpoints(input)

	if acctest.PreCheckSkipError(err) {
		t.Skipf("skipping acceptance testing: %s", err)
	}

	if err != nil {
		t.Fatalf("unexpected PreCheck error: %s", err)
	}
}

func testAccEndpointConfig_base(rInt int) string {
	return fmt.Sprintf(`
resource "aws_vpc" "foo" {
  cidr_block           = "10.0.0.0/16"
  enable_dns_support   = true
  enable_dns_hostnames = true

  tags = {
    Name = "terraform-testacc-r53-resolver-vpc-%[1]d"
  }
}

data "aws_availability_zones" "available" {
  state = "available"

  filter {
    name   = "opt-in-status"
    values = ["opt-in-not-required"]
  }
}

resource "aws_subnet" "sn1" {
  vpc_id            = aws_vpc.foo.id
  cidr_block        = cidrsubnet(aws_vpc.foo.cidr_block, 2, 0)
  availability_zone = data.aws_availability_zones.available.names[0]

  tags = {
    Name = "tf-acc-r53-resolver-sn1-%[1]d"
  }
}

resource "aws_subnet" "sn2" {
  vpc_id            = aws_vpc.foo.id
  cidr_block        = cidrsubnet(aws_vpc.foo.cidr_block, 2, 1)
  availability_zone = data.aws_availability_zones.available.names[1]

  tags = {
    Name = "tf-acc-r53-resolver-sn2-%[1]d"
  }
}

resource "aws_subnet" "sn3" {
  vpc_id            = aws_vpc.foo.id
  cidr_block        = cidrsubnet(aws_vpc.foo.cidr_block, 2, 2)
  availability_zone = data.aws_availability_zones.available.names[2]

  tags = {
    Name = "tf-acc-r53-resolver-sn3-%[1]d"
  }
}

resource "aws_security_group" "sg1" {
  vpc_id = aws_vpc.foo.id
  name   = "tf-acc-r53-resolver-sg1-%[1]d"

  tags = {
    Name = "tf-acc-r53-resolver-sg1-%[1]d"
  }
}

resource "aws_security_group" "sg2" {
  vpc_id = aws_vpc.foo.id
  name   = "tf-acc-r53-resolver-sg2-%[1]d"

  tags = {
    Name = "tf-acc-r53-resolver-sg2-%[1]d"
  }
}
`, rInt)
}

func testAccEndpointConfig_initial(rInt int, direction, name string) string {
	return fmt.Sprintf(`
%s

resource "aws_route53_resolver_endpoint" "foo" {
  direction = "%s"
  name      = "%s"

  security_group_ids = [
    aws_security_group.sg1.id,
    aws_security_group.sg2.id,
  ]

  ip_address {
    subnet_id = aws_subnet.sn1.id
  }

  ip_address {
    subnet_id = aws_subnet.sn2.id
    ip        = cidrhost(aws_subnet.sn2.cidr_block, 8)
  }

  tags = {
    Environment = "production"
    Usage       = "original"
  }
}
`, testAccEndpointConfig_base(rInt), direction, name)
}

func testAccEndpointConfig_updated(rInt int, direction, name string) string {
	return fmt.Sprintf(`
%s

resource "aws_route53_resolver_endpoint" "foo" {
  direction = "%s"
  name      = "%s"

  security_group_ids = [
    aws_security_group.sg1.id,
    aws_security_group.sg2.id,
  ]

  ip_address {
    subnet_id = aws_subnet.sn1.id
  }

  ip_address {
    subnet_id = aws_subnet.sn3.id
  }

  tags = {
    Usage = "changed"
  }
}
`, testAccEndpointConfig_base(rInt), direction, name)
}
