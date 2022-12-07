package ec2_test

import (
	"fmt"
	"os"
	"testing"

	"github.com/aws/aws-sdk-go/service/ec2"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
)

func TestAccVPCPublicIpv4PoolsDataSource_main(t *testing.T) {
	if os.Getenv("PUBLIC_IPV4_POOLS_MESSAGE") == "" ||
		os.Getenv("PUBLIC_IPV4_POOLS_SIGNATURE") == "" ||
		os.Getenv("PUBLIC_IPV4_POOLS_PROVISIONED_CIDR") == "" ||
		os.Getenv("PUBLIC_IPV4_POOLS_NETMASK_LENGTH") == "" ||
		os.Getenv("PUBLIC_IPV4_POOLS_VPC_CIDR") == "" {
		t.Skip("Environment variable PUBLIC_IPV4_POOLS_MESSAGE, PUBLIC_IPV4_POOLS_SIGNATURE, PUBLIC_IPV4_POOLS_NETMASK_LENGTH, PUBLIC_IPV4_POOLS_VPC_CIDR, or PUBLIC_IPV4_POOLS_PROVISIONED_CIDR is not set")
	}

	var m string
	var s string
	var p string
	var v string
	var l string

	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	m = os.Getenv("PUBLIC_IPV4_POOLS_MESSAGE")
	s = os.Getenv("PUBLIC_IPV4_POOLS_SIGNATURE")

	p = os.Getenv("PUBLIC_IPV4_POOLS_PROVISIONED_CIDR")
	v = os.Getenv("PUBLIC_IPV4_POOLS_VPC_CIDR")
	l = os.Getenv("PUBLIC_IPV4_POOLS_NETMASK_LENGTH")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ErrorCheck:               acctest.ErrorCheck(t, ec2.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckVPCIPv4CIDRBlockAssociationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccVPCPublicIpv4PoolsDataSource_base(p, m, s, v, l),
				Check: resource.ComposeTestCheckFunc(
					// Check that contents of the test pool are what we expect:
					resource.TestCheckResourceAttr("data.aws_vpc_public_ipv4_pools.test", "pools.#", "2"),
				),
			},
			{
				Config: testAccVPCPublicIpv4PoolsDataSource_filter(p, m, s, v, l),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("data.aws_vpc_public_ipv4_pools.test", "pools.#", "1"),
				),
			},
			{
				Config: testAccVPCPublicIpv4PoolsDataSource_tags(p, m, s, v, l),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("data.aws_vpc_public_ipv4_pools.test", "pools.#", "1"),
				),
			},
			{
				Config: testAccVPCPublicIpv4PoolsDataSource_empty(rName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("data.aws_vpc_public_ipv4_pools.test", "pools.#", "0"),
				),
			},
		},
	})
}

func testAccVPCPublicIpv4PoolsDataSource_base(cidr, msg, signature, vpcCidr string, netmaskLength string) string {
	return fmt.Sprintf(`
data "aws_region" "current" {}

resource "aws_vpc_ipam" "test" {
  operating_regions {
    region_name = data.aws_region.current.name
  }
}

resource "aws_vpc_ipam_pool" "test_1" {
  address_family                    = "ipv4"
  ipam_scope_id                     = aws_vpc_ipam.test.public_default_scope_id
  locale                            = data.aws_region.current.name
  aws_service                       = "ec2"
  allocation_default_netmask_length = %[5]s
}

resource "aws_vpc_ipam_pool" "test_2" {
  address_family                    = "ipv4"
  ipam_scope_id                     = aws_vpc_ipam.test.public_default_scope_id
  locale                            = data.aws_region.current.name
  aws_service                       = "ec2"
  allocation_default_netmask_length = %[5]s
}

resource "aws_vpc_ipam_pool_cidr" "test_1" {
  ipam_pool_id = aws_vpc_ipam_pool.test_1.id
  cidr         = %[1]q

  cidr_authorization_context {
    message   = %[2]q
    signature = %[3]q
  }
}

resource "aws_vpc" "test" {
  cidr_block = "10.0.0.0/16"
}

resource "aws_vpc_ipv4_cidr_block_association" "test_1" {
  ipv4_ipam_pool_id = aws_vpc_ipam_pool.test_1.id
  ipv4_cidr_block   = %[4]q
  vpc_id            = aws_vpc.test.id
  depends_on = [
    aws_vpc_ipam_pool_cidr.test
  ]
}

resource "aws_vpc_ipv4_cidr_block_association" "test_2" {
  ipv4_ipam_pool_id = aws_vpc_ipam_pool.test_2.id
  ipv4_cidr_block   = %[4]q
  vpc_id            = aws_vpc.test.id
  depends_on = [
	aws_vpc_ipam_pool_cidr.test
  ]
}

data "aws_vpc_public_ipv4_pools" "test" {
  pool_ids = [aws_vpc_ipam_pool.test_1.id, aws_vpc_ipam_pool.test_2.id]
}
	`, cidr, msg, signature, vpcCidr, netmaskLength)
}

func testAccVPCPublicIpv4PoolsDataSource_filter(cidr, msg, signature, vpcCidr string, netmaskLength string) string {
	return fmt.Sprintf(`
data "aws_region" "current" {}

resource "aws_vpc_ipam" "test" {
  operating_regions {
    region_name = data.aws_region.current.name
  }
}

resource "aws_vpc_ipam_pool" "test_1" {
  address_family                    = "ipv4"
  ipam_scope_id                     = aws_vpc_ipam.test.public_default_scope_id
  locale                            = data.aws_region.current.name
  aws_service                       = "ec2"
  allocation_default_netmask_length = %[5]s
  tags = {
	UniqueTagKey = "UnimportantValue"
  }
}

resource "aws_vpc_ipam_pool_cidr" "test_1" {
  ipam_pool_id = aws_vpc_ipam_pool.test_1.id
  cidr         = %[1]q
  
  cidr_authorization_context {
    message   = %[2]q
    signature = %[3]q
  }
}

resource "aws_vpc" "test" {
  cidr_block = "10.0.0.0/16"
}

resource "aws_vpc_ipv4_cidr_block_association" "test_1" {
  ipv4_ipam_pool_id = aws_vpc_ipam_pool.test_1.id
  ipv4_cidr_block   = %[4]q
  vpc_id            = aws_vpc.test.id
  depends_on = [
    aws_vpc_ipam_pool_cidr.test
  ]
}

data "aws_vpc_public_ipv4_pools" "test" {
  filter {
    name   = "tag-key"
	values = ["UniqueTagKey"]
  }
}
	`, cidr, msg, signature, vpcCidr, netmaskLength)
}

func testAccVPCPublicIpv4PoolsDataSource_tags(cidr, msg, signature, vpcCidr string, netmaskLength string) string {
	return fmt.Sprintf(`
data "aws_region" "current" {}

resource "aws_vpc_ipam" "test" {
  operating_regions {
    region_name = data.aws_region.current.name
  }
}

resource "aws_vpc_ipam_pool" "test_1" {
  address_family                    = "ipv4"
  ipam_scope_id                     = aws_vpc_ipam.test.public_default_scope_id
  locale                            = data.aws_region.current.name
  aws_service                       = "ec2"
  allocation_default_netmask_length = %[5]s

  tags = {
	Name = "ipv4_pool_test_2"
  }
}

resource "aws_vpc_ipam_pool_cidr" "test_1" {
  ipam_pool_id = aws_vpc_ipam_pool.test_1.id
  cidr         = %[1]q

  cidr_authorization_context {
    message   = %[2]q
    signature = %[3]q
  }
}

resource "aws_vpc" "test" {
  cidr_block = "10.0.0.0/16"
}

resource "aws_vpc_ipv4_cidr_block_association" "test_1" {
  ipv4_ipam_pool_id = aws_vpc_ipam_pool.test_1.id
  ipv4_cidr_block   = %[4]q
  vpc_id            = aws_vpc.test.id
  depends_on = [
    aws_vpc_ipam_pool_cidr.test
  ]
}

data "aws_vpc_public_ipv4_pools" "test" {
  tags = {
    Name = aws_vpc_ipam_pool.test_1.tags.Name
  }
}
	`, cidr, msg, signature, vpcCidr, netmaskLength)
}

func testAccVPCPublicIpv4PoolsDataSource_empty(rName string) string {
	return fmt.Sprintf(`
data "aws_vpc_public_ipv4_pools" "test" {
  tags = {
    Name = %[1]q
  }
}
`, rName)
}
