package ec2_test

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfec2 "github.com/hashicorp/terraform-provider-aws/internal/service/ec2"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"

	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccVPCIpam_basic(t *testing.T) {
	// var pool ec2.IpamPool
	resourceName := "aws_vpc_ipam.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, ec2.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckVPCIpamDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccVPCIpam,
				Check: resource.ComposeTestCheckFunc(
					acctest.MatchResourceAttrRegionalARN(resourceName, "arn", "ec2", regexp.MustCompile(`ipam/ipam-[a-z0-9]+`)),
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

func TestAccVPCIpam_modifyRegion(t *testing.T) {
	resourceName := "aws_vpc_ipam.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, ec2.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckVPCIpamDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccVPCIpam,
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
				Config: testAccVPCIpamOperatingRegion,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "operating_regions.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "description", "test ipam"),
				// resource.TestCheckResourceAttr(rName, "tags.%", "0"),
				),
			},
			{
				Config: testAccVPCIpam,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "operating_regions.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "description", "test"),
					// resource.TestCheckResourceAttr(rName, "tags.%", "0"),
				),
			},
		},
	})
}

func testAccCheckVPCIpamDestroy(s *terraform.State) error {
	conn := acctest.Provider.Meta().(*conns.AWSClient).EC2Conn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_vpc_ipam" {
			continue
		}

		id := aws.String(rs.Primary.ID)

		if _, err := tfec2.WaiterIpamDeleted(conn, *id, tfec2.IpamDeleteTimeout); err != nil {
			if tfresource.NotFound(err) {
				return nil
			}
			return fmt.Errorf("error waiting for IPAM to be deleted: %w", err)
		}
	}

	return nil
}

const testAccVPCIpam = `
data "aws_region" "current" {}

resource "aws_vpc_ipam" "test" {
	description = "test"
	operating_regions {
	  region_name = data.aws_region.current.name
	}
}
`

const testAccVPCIpamOperatingRegion = `
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
