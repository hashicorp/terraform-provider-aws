package ec2_test

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/service/ec2"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
)

func TestAccVPCPublicIpv4PoolsDataSource_filter(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ErrorCheck:               acctest.ErrorCheck(t, ec2.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckVPCDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccVPCPublicIpv4PoolsDataSourceConfig_filter(rName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("data.aws_vpc_public_ipv4_pools.test", "pools.#", "1"),
				),
			},
		},
	})
}

func TestAccVPCPublicIpv4PoolsDataSource_tags(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ErrorCheck:               acctest.ErrorCheck(t, ec2.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckVPCDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccVPCPublicIpv4PoolsDataSourceConfig_tags(rName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("data.aws_vpc_public_ipv4_pools.test", "pools.#", "1"),
				),
			},
		},
	})
}

func TestAccVPCPublicIpv4PoolsDataSource_empty(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ErrorCheck:               acctest.ErrorCheck(t, ec2.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckVPCDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccVPCPublicIpv4PoolsDataSourceConfig_empty(rName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("data.aws_vpc_public_ipv4_pools.test", "pools.#", "0"),
				),
			},
		},
	})
}

func testAccVPCPublicIpv4PoolsDataSourceConfig_Base(rName string) string {
	return fmt.Sprintf(`
data "aws_region" "current" {}

resource "aws_vpc_ipam" "test_ipam" {
	operating_regions {
		region_name = data.aws_region.current.name
	}
}

resource "aws_vpc_ipam_pool" "test_pool_1" {
	address_family = "ipv4"
	ipam_scope_id  = aws_vpc_ipam.test_ipam.private_default_scope_id
	locale         = data.aws_region.current.name
	tags = {
		Name = "test-1"
	  }
}

resource "aws_vpc_ipam_pool" "test_pool_2" {
	address_family = "ipv4"
	ipam_scope_id  = aws_vpc_ipam.test_ipam.private_default_scope_id
	locale         = data.aws_region.current.name
	tags = {
		Name = "test-2"
		UniqueTagKey = "unimportant"
	  }
}

resource "aws_vpc_ipam_pool_cidr" "test_cidr_1" {
	ipam_pool_id = aws_vpc_ipam_pool.test_pool_1.id
	cidr         = "172.1.0.0/16"
}

resource "aws_vpc_ipam_pool_cidr" "test_cidr_2" {
	ipam_pool_id = aws_vpc_ipam_pool.test_pool_1.id
	cidr         = "172.2.0.0/16"
}

resource "aws_vpc_ipam_pool_cidr" "test_cidr_3" {
	ipam_pool_id = aws_vpc_ipam_pool.test_pool_2.id
	cidr         = "172.3.0.0/16"
}

resource "aws_vpc_ipam_pool_cidr" "test_cidr_4" {
	ipam_pool_id = aws_vpc_ipam_pool.test_pool_2.id
	cidr         = "172.4.0.0/16"
}
`, rName)
}

func testAccVPCPublicIpv4PoolsDataSourceConfig_filter(rName string) string {
	return acctest.ConfigCompose(testAccVPCPublicIpv4PoolsDataSourceConfig_Base(rName), `
data "aws_vpc_public_ipv4_pools" "test" {
  filter {
    name   = "tag-key"
    values = ["UniqueTagKey"]
  }
}
`)
}

func testAccVPCPublicIpv4PoolsDataSourceConfig_tags(rName string) string {
	return acctest.ConfigCompose(testAccVPCPublicIpv4PoolsDataSourceConfig_Base(rName), `
data "aws_vpc_public_ipv4_pools" "test" {
  tags = {
    Name = aws_vpc_ipam_pool.test_pool_1.tags.Name
  }
}
`)
}

func testAccVPCPublicIpv4PoolsDataSourceConfig_empty(rName string) string {
	return fmt.Sprintf(`
data "aws_vpc_public_ipv4_pools" "test" {
  tags = {
    Name = "not_a_real_tag"
  }
}
`, rName)
}
