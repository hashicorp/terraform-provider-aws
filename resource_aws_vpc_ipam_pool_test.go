package aws

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccAWSVpcIpamPool_basic(t *testing.T) {
	var pool ec2.IpamPool
	resourceName := "aws_vpc_ipam_pool.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		ErrorCheck:   testAccErrorCheck(t, ec2.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAwsVpcIpamPoolDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAwsVpcIpamPool,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsVpcIpamPoolExists(resourceName, &pool),
					resource.TestCheckResourceAttr(resourceName, "address_family", "ipv4"),
					resource.TestCheckResourceAttr(resourceName, "auto_import", "false"),
					resource.TestCheckResourceAttr(resourceName, "locale", "None"),
					resource.TestCheckResourceAttr(resourceName, "state", "create-complete"),
					// resource.TestCheckResourceAttr(rName, "tags.%", "0"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccAwsVpcIpamPoolUpdates,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "address_family", "ipv4"),
					resource.TestCheckResourceAttr(resourceName, "auto_import", "true"),
					resource.TestCheckResourceAttr(resourceName, "locale", "None"),
					resource.TestCheckResourceAttr(resourceName, "state", "modify-complete"),
					resource.TestCheckResourceAttr(resourceName, "allocation_default_netmask_length", "32"),
					resource.TestCheckResourceAttr(resourceName, "allocation_max_netmask_length", "32"),
					resource.TestCheckResourceAttr(resourceName, "allocation_min_netmask_length", "32"),
					resource.TestCheckResourceAttr(resourceName, "allocation_resource_tags.test", "1"),
				),
			},
		},
	})
}

func testAccCheckAwsVpcIpamPoolExists(n string, pool *ec2.IpamPool) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		id := rs.Primary.ID
		conn := testAccProvider.Meta().(*AWSClient).ec2conn
		found_pool, err := findIpamPoolById(conn, id)

		if err != nil {
			return err
		}
		*pool = *found_pool

		return nil
	}
}

// func TestAccAWSVpcIpamPool_ipv6(t *testing.T) {
// 	resourceName := "aws_vpc_ipam_pool.test"

// 	resource.ParallelTest(t, resource.TestCase{
// 		PreCheck:     func() { testAccPreCheck(t) },
// 		ErrorCheck:   testAccErrorCheck(t, ec2.EndpointsID),
// 		Providers:    testAccProviders,
// 		CheckDestroy: testAccCheckAwsVpcIpamPoolDestroy,
// 		Steps: []resource.TestStep{
// 			{
// 				Config: testAccAwsVpcIpamPool_ipv6,
// 				Check: resource.ComposeTestCheckFunc(
// 					// testAccCheckAwsVpcIpamExists(rName, &pool),
// 					resource.TestCheckResourceAttr(resourceName, "address_family", "ipv6"),
// 					// resource.TestCheckResourceAttr(resourceName, "auto_import", "false"),
// 					// resource.TestCheckResourceAttr(resourceName, "locale", "None"),
// 					// resource.TestCheckResourceAttr(resourceName, "state", "create-complete"),
// 					// resource.TestCheckResourceAttr(rName, "tags.%", "0"),
// 				),
// 			},
// 			{
// 				ResourceName:      resourceName,
// 				ImportState:       true,
// 				ImportStateVerify: true,
// 			},
// 		},
// 	})
// }

func testAccCheckAwsVpcIpamPoolDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*AWSClient).ec2conn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_vpc_ipam_pool" {
			continue
		}

		id := aws.String(rs.Primary.ID)

		if _, err := waitIpamPoolDeleted(conn, *id, IpamPoolDeleteTimeout); err != nil {
			if isResourceNotFoundError(err) {
				return nil
			}
			return fmt.Errorf("error waiting for IPAM Pool (%s) to be deleted: %w", *id, err)
		}
	}

	return nil
}

const testAccAwsVpcIpamPool = `
resource "aws_vpc_ipam" "test" {
	operating_regions {
	  region_name = "us-east-1"
	}
}
resource "aws_vpc_ipam_pool" "test" {
    address_family = "ipv4"
    ipam_scope_id  = aws_vpc_ipam.test.private_default_scope_id
}
`

const testAccAwsVpcIpamPoolUpdates = `
resource "aws_vpc_ipam" "test" {
	operating_regions {
	  region_name = "us-east-1"
	}
}
resource "aws_vpc_ipam_pool" "test" {
    address_family = "ipv4"
    ipam_scope_id  = aws_vpc_ipam.test.private_default_scope_id
	auto_import    = true
	allocation_default_netmask_length = 32
	allocation_max_netmask_length     = 32
	allocation_min_netmask_length     = 32
	allocation_resource_tags          = {
		test = "1"
	}
	description                       = "test"
}
`

const testAccAwsVpcIpamPool_ipv6 = `
resource "aws_vpc_ipam" "test" {
	operating_regions {
	  region_name = "us-east-1"
	}
}
resource "aws_vpc_ipam_pool" "test" {
	address_family = "ipv6"
	ipam_scope_id  =  aws_vpc_ipam.test.public_default_scope_id
	locale         = "us-east-1"
	description    = "ipv6 test"
	advertisable   = false
}
`
