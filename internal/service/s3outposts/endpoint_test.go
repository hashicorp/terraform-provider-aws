package s3outposts_test

import (
	"context"
	"fmt"
	"regexp"
	"testing"

	"github.com/aws/aws-sdk-go/service/s3outposts"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfs3outposts "github.com/hashicorp/terraform-provider-aws/internal/service/s3outposts"
)

func TestAccS3OutpostsEndpoint_basic(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_s3outposts_endpoint.test"
	rInt := sdkacctest.RandIntRange(0, 255)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckOutpostsOutposts(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, s3outposts.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckEndpointDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccEndpointConfig_basic(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckEndpointExists(ctx, resourceName),
					acctest.MatchResourceAttrRegionalARN(resourceName, "arn", "s3-outposts", regexp.MustCompile(`outpost/[^/]+/endpoint/[a-z0-9]+`)),
					resource.TestCheckResourceAttrSet(resourceName, "creation_time"),
					resource.TestCheckResourceAttrPair(resourceName, "cidr_block", "aws_vpc.test", "cidr_block"),
					resource.TestCheckResourceAttr(resourceName, "network_interfaces.#", "4"),
					resource.TestCheckResourceAttrPair(resourceName, "outpost_id", "data.aws_outposts_outpost.test", "id"),
					resource.TestCheckResourceAttrPair(resourceName, "security_group_id", "aws_security_group.test", "id"),
					resource.TestCheckResourceAttrPair(resourceName, "subnet_id", "aws_subnet.test", "id"),
					resource.TestCheckResourceAttr(resourceName, "access_type", "Private"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateIdFunc: testAccEndpointImportStateIdFunc(resourceName),
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccS3OutpostsEndpoint_private(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_s3outposts_endpoint.test"
	rInt := sdkacctest.RandIntRange(0, 255)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckOutpostsOutposts(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, s3outposts.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckEndpointDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccEndpointPrivateConfig(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckEndpointExists(ctx, resourceName),
					acctest.MatchResourceAttrRegionalARN(resourceName, "arn", "s3-outposts", regexp.MustCompile(`outpost/[^/]+/endpoint/[a-z0-9]+`)),
					resource.TestCheckResourceAttrSet(resourceName, "creation_time"),
					resource.TestCheckResourceAttrPair(resourceName, "cidr_block", "aws_vpc.test", "cidr_block"),
					resource.TestCheckResourceAttr(resourceName, "network_interfaces.#", "4"),
					resource.TestCheckResourceAttrPair(resourceName, "outpost_id", "data.aws_outposts_outpost.test", "id"),
					resource.TestCheckResourceAttrPair(resourceName, "security_group_id", "aws_security_group.test", "id"),
					resource.TestCheckResourceAttrPair(resourceName, "subnet_id", "aws_subnet.test", "id"),
					resource.TestCheckResourceAttr(resourceName, "access_type", "Private"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateIdFunc: testAccEndpointImportStateIdFunc(resourceName),
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccS3OutpostsEndpoint_customerOwnedIPv4Pool(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_s3outposts_endpoint.test"
	rInt := sdkacctest.RandIntRange(0, 255)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckOutpostsOutposts(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, s3outposts.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckEndpointDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccEndpointCustomerOwnedIPv4PoolConfig(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckEndpointExists(ctx, resourceName),
					acctest.MatchResourceAttrRegionalARN(resourceName, "arn", "s3-outposts", regexp.MustCompile(`outpost/[^/]+/endpoint/[a-z0-9]+`)),
					resource.TestCheckResourceAttrSet(resourceName, "creation_time"),
					resource.TestCheckResourceAttrPair(resourceName, "cidr_block", "aws_vpc.test", "cidr_block"),
					resource.TestCheckResourceAttr(resourceName, "network_interfaces.#", "4"),
					resource.TestCheckResourceAttrPair(resourceName, "outpost_id", "data.aws_outposts_outpost.test", "id"),
					resource.TestCheckResourceAttrPair(resourceName, "security_group_id", "aws_security_group.test", "id"),
					resource.TestCheckResourceAttrPair(resourceName, "subnet_id", "aws_subnet.test", "id"),
					resource.TestCheckResourceAttr(resourceName, "access_type", "CustomerOwnedIp"),
					resource.TestMatchResourceAttr(resourceName, "customer_owned_ipv4_pool", regexp.MustCompile(`^ipv4pool-coip-.+$`)),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateIdFunc: testAccEndpointImportStateIdFunc(resourceName),
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccS3OutpostsEndpoint_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_s3outposts_endpoint.test"
	rInt := sdkacctest.RandIntRange(0, 255)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckOutpostsOutposts(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, s3outposts.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckEndpointDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccEndpointConfig_basic(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckEndpointExists(ctx, resourceName),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfs3outposts.ResourceEndpoint(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckEndpointDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).S3OutpostsConn()

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_s3outposts_endpoint" {
				continue
			}

			endpoint, err := tfs3outposts.FindEndpoint(ctx, conn, rs.Primary.ID)

			if err != nil {
				return err
			}

			if endpoint != nil {
				return fmt.Errorf("S3 Outposts Endpoint (%s) still exists", rs.Primary.ID)
			}
		}

		return nil
	}
}

func testAccCheckEndpointExists(ctx context.Context, resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("not found: %s", resourceName)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("no resource ID is set")
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).S3OutpostsConn()

		endpoint, err := tfs3outposts.FindEndpoint(ctx, conn, rs.Primary.ID)

		if err != nil {
			return err
		}

		if endpoint == nil {
			return fmt.Errorf("S3 Outposts Endpoint (%s) not found", rs.Primary.ID)
		}

		return nil
	}
}

func testAccEndpointImportStateIdFunc(resourceName string) resource.ImportStateIdFunc {
	return func(s *terraform.State) (string, error) {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return "", fmt.Errorf("Not found: %s", resourceName)
		}

		return fmt.Sprintf("%s,%s,%s", rs.Primary.ID, rs.Primary.Attributes["security_group_id"], rs.Primary.Attributes["subnet_id"]), nil
	}
}

func testAccEndpointConfig_basic(rInt int) string {
	return fmt.Sprintf(`
data "aws_outposts_outposts" "test" {}

data "aws_outposts_outpost" "test" {
  id = tolist(data.aws_outposts_outposts.test.ids)[0]
}

resource "aws_vpc" "test" {
  cidr_block = "10.%[1]d.0.0/16"
}

resource "aws_security_group" "test" {
  vpc_id = aws_vpc.test.id
}

resource "aws_subnet" "test" {
  availability_zone = data.aws_outposts_outpost.test.availability_zone
  cidr_block        = cidrsubnet(aws_vpc.test.cidr_block, 8, 0)
  outpost_arn       = data.aws_outposts_outpost.test.arn
  vpc_id            = aws_vpc.test.id
}

resource "aws_s3outposts_endpoint" "test" {
  outpost_id        = data.aws_outposts_outpost.test.id
  security_group_id = aws_security_group.test.id
  subnet_id         = aws_subnet.test.id
}
`, rInt)
}

func testAccEndpointPrivateConfig(rInt int) string {
	return fmt.Sprintf(`
data "aws_outposts_outposts" "test" {}

data "aws_outposts_outpost" "test" {
  id = tolist(data.aws_outposts_outposts.test.ids)[0]
}

resource "aws_vpc" "test" {
  cidr_block = "10.%[1]d.0.0/16"
}

resource "aws_security_group" "test" {
  vpc_id = aws_vpc.test.id
}

resource "aws_subnet" "test" {
  availability_zone = data.aws_outposts_outpost.test.availability_zone
  cidr_block        = cidrsubnet(aws_vpc.test.cidr_block, 8, 0)
  outpost_arn       = data.aws_outposts_outpost.test.arn
  vpc_id            = aws_vpc.test.id
}

resource "aws_s3outposts_endpoint" "test" {
  outpost_id        = data.aws_outposts_outpost.test.id
  security_group_id = aws_security_group.test.id
  subnet_id         = aws_subnet.test.id
  access_type       = "Private"
}
`, rInt)
}

func testAccEndpointCustomerOwnedIPv4PoolConfig(rInt int) string {
	return fmt.Sprintf(`
data "aws_outposts_outposts" "test" {}

data "aws_outposts_outpost" "test" {
  id = tolist(data.aws_outposts_outposts.test.ids)[0]
}

data "aws_ec2_coip_pools" "test" {}

data "aws_ec2_local_gateway_route_table" "test" {
  outpost_arn = data.aws_outposts_outpost.test.arn
}

resource "aws_vpc" "test" {
  cidr_block = "10.%[1]d.0.0/16"
}

resource "aws_ec2_local_gateway_route_table_vpc_association" "test" {
  local_gateway_route_table_id = data.aws_ec2_local_gateway_route_table.test.id
  vpc_id                       = aws_vpc.test.id
}

resource "aws_security_group" "test" {
  vpc_id = aws_vpc.test.id
}

resource "aws_subnet" "test" {
  availability_zone = data.aws_outposts_outpost.test.availability_zone
  cidr_block        = cidrsubnet(aws_vpc.test.cidr_block, 8, 0)
  outpost_arn       = data.aws_outposts_outpost.test.arn
  vpc_id            = aws_vpc.test.id
}

resource "aws_s3outposts_endpoint" "test" {
  depends_on = [ aws_ec2_local_gateway_route_table_vpc_association.test ]
  outpost_id               = data.aws_outposts_outpost.test.id
  security_group_id        = aws_security_group.test.id
  subnet_id                = aws_subnet.test.id
  access_type              = "CustomerOwnedIp"
  customer_owned_ipv4_pool = tolist(data.aws_ec2_coip_pools.test.pool_ids)[0]
}
`, rInt)
}
