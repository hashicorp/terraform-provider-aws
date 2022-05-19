package ec2_test

import (
	"testing"

	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
)

func TestAccIPAMPoolDataSource_basic(t *testing.T) {
	resourceName := "aws_vpc_ipam_pool.test"
	dataSourceName := "data.aws_vpc_ipam_pool.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccIPAMPreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ec2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccIPAMPoolOptions_basic,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrPair(dataSourceName, "address_family", resourceName, "address_family"),
					resource.TestCheckResourceAttrPair(dataSourceName, "allocation_default_netmask_length", resourceName, "allocation_default_netmask_length"),
					resource.TestCheckResourceAttrPair(dataSourceName, "allocation_max_netmask_length", resourceName, "allocation_max_netmask_length"),
					resource.TestCheckResourceAttrPair(dataSourceName, "allocation_min_netmask_length", resourceName, "allocation_min_netmask_length"),
					resource.TestCheckResourceAttrPair(dataSourceName, "allocation_resource_tags.%", resourceName, "allocation_resource_tags.%"),
					resource.TestCheckResourceAttrPair(dataSourceName, "arn", resourceName, "arn"),
					resource.TestCheckResourceAttrPair(dataSourceName, "auto_import", resourceName, "auto_import"),
					resource.TestCheckResourceAttrPair(dataSourceName, "aws_service", resourceName, "aws_service"),
					resource.TestCheckResourceAttrPair(dataSourceName, "description", resourceName, "description"),
					resource.TestCheckResourceAttrPair(dataSourceName, "ipam_scope_id", resourceName, "ipam_scope_id"),
					resource.TestCheckResourceAttrPair(dataSourceName, "ipam_scope_type", resourceName, "ipam_scope_type"),
					resource.TestCheckResourceAttrPair(dataSourceName, "locale", resourceName, "locale"),
					resource.TestCheckResourceAttrPair(dataSourceName, "pool_depth", resourceName, "pool_depth"),
					resource.TestCheckResourceAttrPair(dataSourceName, "publicly_advertisable", resourceName, "publicly_advertisable"),
					resource.TestCheckResourceAttrPair(dataSourceName, "source_ipam_pool_id", resourceName, "source_ipam_pool_id"),
					resource.TestCheckResourceAttrPair(dataSourceName, "state", resourceName, "state"),
					resource.TestCheckResourceAttrPair(dataSourceName, "tags.%", resourceName, "tags.%"),
				),
			},
		},
	})
}

var testAccIPAMPoolOptions_basic = acctest.ConfigCompose(testAccIPAMPoolConfig_base, `
resource "aws_vpc_ipam_pool" "test" {
  address_family                    = "ipv4"
  ipam_scope_id                     = aws_vpc_ipam.test.private_default_scope_id
  auto_import                       = true
  allocation_default_netmask_length = 32
  allocation_max_netmask_length     = 32
  allocation_min_netmask_length     = 32
  allocation_resource_tags = {
    test = "1"
  }
  description = "test"
}

data "aws_vpc_ipam_pool" "test" {
  ipam_pool_id = aws_vpc_ipam_pool.test.id
}
`)
