package aws

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccAWSVpcIpam_basicIpam(t *testing.T) {
	// var pool ec2.IpamPool
	resourceName := "aws_vpc_ipam.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		ErrorCheck:   testAccErrorCheck(t, ec2.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAwsVpcIpamDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAwsVpcIpam,
				Check: resource.ComposeTestCheckFunc(
					testAccMatchResourceAttrRegionalARN(resourceName, "arn", "ec2", regexp.MustCompile(`ipam/ipam-[a-z0-9]+`)),
					resource.TestCheckResourceAttr(resourceName, "operating_regions.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "scope_count", "2"),
					resource.TestMatchResourceAttr(resourceName, "private_default_scope_id", regexp.MustCompile("^ipam-scope-[a-z0-9]+")),
					resource.TestMatchResourceAttr(resourceName, "public_default_scope_id", regexp.MustCompile("^ipam-scope-[a-z0-9]+")),
					// resource.TestCheckResourceAttr(rName, "tags.%", "0"),
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

func TestAccAWSVpcIpam_modifyRegion(t *testing.T) {
	resourceName := "aws_vpc_ipam.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		ErrorCheck:   testAccErrorCheck(t, ec2.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAwsVpcIpamDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAwsVpcIpam,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "operating_regions.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "description", "test"),
					// resource.TestCheckResourceAttr(rName, "tags.%", "0"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccAwsVpcIpamOperatingRegion,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "operating_regions.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "description", "test ipam"),
				// resource.TestCheckResourceAttr(rName, "tags.%", "0"),
				),
			},
			{
				Config: testAccAwsVpcIpam,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "operating_regions.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "description", "test"),
					// resource.TestCheckResourceAttr(rName, "tags.%", "0"),
				),
			},
		},
	})
}

func testAccCheckAwsVpcIpamDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*AWSClient).ec2conn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_vpc_ipam" {
			continue
		}

		id := aws.String(rs.Primary.ID)

		if _, err := waiterIpamDeleted(conn, *id, IpamDeleteTimeout); err != nil {
			if isResourceNotFoundError(err) {
				return nil
			}
			return fmt.Errorf("error waiting for IPAM to be deleted: %w", err)
		}
	}

	return nil
}

const testAccAwsVpcIpam = `
resource "aws_vpc_ipam" "test" {
	description = "test"
	operating_regions {
	  region_name = "us-east-1"
	}
}
`

const testAccAwsVpcIpamOperatingRegion = `
resource "aws_vpc_ipam" "test" {
	description = "test ipam"
	operating_regions {
	  region_name = "us-east-1"
	}
	operating_regions {
	  region_name = "us-west-2"
	}
}
`
