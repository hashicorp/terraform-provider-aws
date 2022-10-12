package ec2_test

import (
	"fmt"
	"regexp"
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

func TestAccIPAMResourceDiscoveryAssociation_basic(t *testing.T) {
	resourceName := "aws_vpc_ipam_resource_discovery_association.test"
	ipamName := "aws_vpc_ipam.test"
	rdName := "aws_vpc_ipam_resource_discovery.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t); testAccIPAMPreCheck(t) },
		ErrorCheck:               acctest.ErrorCheck(t, ec2.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckIPAMResourceDiscoveryDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccIPAMResourceDiscoveryAssociationConfig_basic(),
				Check: resource.ComposeTestCheckFunc(
					acctest.MatchResourceAttrGlobalARN(resourceName, "arn", "ec2", regexp.MustCompile(`ipam-resource-discovery-association/ipam-res-disco-assoc-[\da-f]+$`)),
					resource.TestCheckResourceAttrPair(resourceName, "ipam_id", ipamName, "id"),
					resource.TestCheckResourceAttrPair(resourceName, "ipam_resource_discovery_id", rdName, "id"),
					acctest.CheckResourceAttrAccountID(resourceName, "owner_id"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
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

func TestAccIPAMResourceDiscoveryAssociation_tags(t *testing.T) {
	resourceName := "aws_vpc_ipam_resource_discovery_association.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t); testAccIPAMPreCheck(t) },
		ErrorCheck:               acctest.ErrorCheck(t, ec2.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckIPAMResourceDiscoveryDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccIPAMResourceDiscoveryAssociationConfig_tags("key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccIPAMResourceDiscoveryAssociationConfig_tags2("key1", "value1updated", "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1updated"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
			{
				Config: testAccIPAMResourceDiscoveryAssociationConfig_tags("key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
		},
	})
}

func testAccCheckIPAMResourceDiscoveryAssociationDestroy(s *terraform.State) error {
	conn := acctest.Provider.Meta().(*conns.AWSClient).EC2Conn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_vpc_ipam_resource_discovery_association" {
			continue
		}

		id := aws.String(rs.Primary.ID)

		if _, err := tfec2.WaiterIPAMResourceDiscoveryAssociationDeleted(conn, *id, tfec2.IPAMResourceDiscoveryAssociationDeleteTimeout); err != nil {
			if tfresource.NotFound(err) {
				return nil
			}
			return fmt.Errorf("waiting for IPAMResourceDiscovery to be dissociated: %w", err)
		}
	}

	return nil
}

const testAccIPAMResourceDiscoveryAssociationConfig_base = `
data "aws_region" "current" {}

resource "aws_vpc_ipam" "test" {
	description = "test ipam"
	operating_regions {
	  region_name = data.aws_region.current.name
	}
}

resource "aws_vpc_ipam_resource_discovery" "test" {
  description = "test ipam resource discovery"
  operating_regions {
    region_name = data.aws_region.current.name
  }
}
`

func testAccIPAMResourceDiscoveryAssociationConfig_basic() string {
	return acctest.ConfigCompose(testAccIPAMResourceDiscoveryAssociationConfig_base, `
resource "aws_vpc_ipam_resource_discovery_association" "test" {
  ipam_id                    = aws_vpc_ipam.test.id
  ipam_resource_discovery_id = aws_vpc_ipam_resource_discovery.test.id
}
`)
}

func testAccIPAMResourceDiscoveryAssociationConfig_tags(tagKey1, tagValue1 string) string {
	return acctest.ConfigCompose(testAccIPAMResourceDiscoveryAssociationConfig_base, fmt.Sprintf(`
resource "aws_vpc_ipam_resource_discovery_association" "test" {
  ipam_id                    = aws_vpc_ipam.test.id
  ipam_resource_discovery_id = aws_vpc_ipam_resource_discovery.test.id
  tags = {
    %[1]q = %[2]q
  }
}
`, tagKey1, tagValue1))
}

func testAccIPAMResourceDiscoveryAssociationConfig_tags2(tagKey1, tagValue1, tagKey2, tagValue2 string) string {
	return acctest.ConfigCompose(testAccIPAMResourceDiscoveryAssociationConfig_base, fmt.Sprintf(`
resource "aws_vpc_ipam_resource_discovery_association" "test" {
  ipam_id                    = aws_vpc_ipam.test.id
  ipam_resource_discovery_id = aws_vpc_ipam_resource_discovery.test.id
  tags = {
    %[1]q = %[2]q
    %[3]q = %[4]q
  }
}
	`, tagKey1, tagValue1, tagKey2, tagValue2))
}
