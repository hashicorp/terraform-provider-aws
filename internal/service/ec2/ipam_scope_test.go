package ec2_test

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfec2 "github.com/hashicorp/terraform-provider-aws/internal/service/ec2"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

func TestAccIPAMScope_basic(t *testing.T) {
	resourceName := "aws_vpc_ipam_scope.test"
	ipamName := "aws_vpc_ipam.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccIPAMPreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ec2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckVPCIpamScopeDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccVPCIpamScope("test"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "description", "test"),
					resource.TestCheckResourceAttr(resourceName, "pool_count", "0"),
					resource.TestCheckResourceAttr(resourceName, "is_default", "false"),
					resource.TestCheckResourceAttrPair(resourceName, "ipam_arn", ipamName, "arn"),
					resource.TestCheckResourceAttrPair(resourceName, "ipam_id", ipamName, "id"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccVPCIpamScope("test2"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "description", "test2"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
				),
			},
		},
	})
}

func TestAccIPAMScope_tags(t *testing.T) {
	resourceName := "aws_vpc_ipam_scope.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccIPAMPreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ec2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckVPCIpamScopeDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccVPCIpamScopeTagsConfig("key1", "value1"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccVPCIpamScopeTags2Config("key1", "value1updated", "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1updated"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
			{
				Config: testAccVPCIpamScopeTagsConfig("key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
		},
	})
}

func testAccCheckVPCIpamScopeDestroy(s *terraform.State) error {
	conn := acctest.Provider.Meta().(*conns.AWSClient).EC2Conn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_vpc_ipam_scope" {
			continue
		}

		id := aws.String(rs.Primary.ID)

		if _, err := tfec2.WaitIpamScopeDeleted(conn, *id, tfec2.IpamScopeDeleteTimeout); err != nil {
			if tfresource.NotFound(err) {
				return nil
			}
			return fmt.Errorf("error waiting for IPAM Scope (%s) to be deleted: %w", *id, err)
		}
	}

	return nil
}

const testAccVPCIpamScopeBase = `
data "aws_region" "current" {}

resource "aws_vpc_ipam" "test" {
  description = "test"
  operating_regions {
    region_name = data.aws_region.current.name
  }
}
`

func testAccVPCIpamScope(desc string) string {
	return testAccVPCIpamScopeBase + fmt.Sprintf(`
resource "aws_vpc_ipam_scope" "test" {
  ipam_id     = aws_vpc_ipam.test.id
  description = %[1]q
}
`, desc)
}

func testAccVPCIpamScopeTagsConfig(tagKey1, tagValue1 string) string {
	return testAccVPCIpamScopeBase + fmt.Sprintf(`
resource "aws_vpc_ipam_scope" "test" {
  ipam_id = aws_vpc_ipam.test.id
  tags = {
    %[1]q = %[2]q
  }
}
`, tagKey1, tagValue1)
}

func testAccVPCIpamScopeTags2Config(tagKey1, tagValue1, tagKey2, tagValue2 string) string {
	return testAccVPCIpamScopeBase + fmt.Sprintf(`


resource "aws_vpc_ipam_scope" "test" {
  ipam_id = aws_vpc_ipam.test.id
  tags = {
    %[1]q = %[2]q
    %[3]q = %[4]q
  }
}
	`, tagKey1, tagValue1, tagKey2, tagValue2)
}
