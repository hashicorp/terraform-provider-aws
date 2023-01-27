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

func TestAccIPAMResourceDiscovery_basic(t *testing.T) {
	resourceName := "aws_vpc_ipam_resource_discovery.test"
	dataSourceRegion := "data.aws_region.current"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ErrorCheck:               acctest.ErrorCheck(t, ec2.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckIPAMResourceDiscoveryDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccIPAMResourceDiscoveryConfig_base,
				Check: resource.ComposeTestCheckFunc(
					acctest.MatchResourceAttrGlobalARN(resourceName, "arn", "ec2", regexp.MustCompile(`ipam-resource-discovery/ipam-res-disco-[\da-f]+$`)),
					resource.TestCheckResourceAttr(resourceName, "description", "test"),
					resource.TestCheckResourceAttrPair(resourceName, "ipam_resource_discovery_region", dataSourceRegion, "name"),
					resource.TestCheckResourceAttr(resourceName, "is_default", "false"),
					resource.TestCheckResourceAttr(resourceName, "operating_regions.#", "1"),
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

func TestAccIPAMResourceDiscovery_modify(t *testing.T) {
	resourceName := "aws_vpc_ipam_resource_discovery.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(t)

			acctest.PreCheckMultipleRegion(t, 2)
		},
		ErrorCheck:               acctest.ErrorCheck(t, ec2.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5FactoriesMultipleRegions(t, 2),
		CheckDestroy:             testAccCheckIPAMResourceDiscoveryDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccIPAMResourceDiscoveryConfig_base,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "description", "test"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccIPAMResourceDiscoveryConfig_operatingRegion(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "description", "test"),
				),
			},
			{
				Config: testAccIPAMResourceDiscoveryConfig_base,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "description", "test"),
				),
			},
			{
				Config: testAccIPAMResourceDiscoveryConfig_baseAlternateDescription,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "description", "test ipam"),
				),
			},
		},
	})
}

func TestAccIPAMResourceDiscovery_tags(t *testing.T) {
	resourceName := "aws_vpc_ipam_resource_discovery.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ErrorCheck:               acctest.ErrorCheck(t, ec2.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckIPAMResourceDiscoveryDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccIPAMResourceDiscoveryConfig_tags("key1", "value1"),
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
				Config: testAccIPAMResourceDiscoveryConfig_tags2("key1", "value1updated", "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1updated"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
			{
				Config: testAccIPAMResourceDiscoveryConfig_tags("key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
		},
	})
}

func testAccCheckIPAMResourceDiscoveryDestroy(s *terraform.State) error {
	conn := acctest.Provider.Meta().(*conns.AWSClient).EC2Conn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_vpc_ipam_resource_discovery" {
			continue
		}

		id := aws.String(rs.Primary.ID)

		if _, err := tfec2.WaiterIPAMResourceDiscoveryDeleted(conn, *id, tfec2.IPAMResourceDiscoveryDeleteTimeout); err != nil {
			if tfresource.NotFound(err) {
				return nil
			}
			return fmt.Errorf("waiting for IPAMResourceDiscovery to be deleted: %w", err)
		}
	}

	return nil
}

const testAccIPAMResourceDiscoveryConfig_base = `
data "aws_region" "current" {}

resource "aws_vpc_ipam_resource_discovery" "test" {
  description = "test"
  operating_regions {
    region_name = data.aws_region.current.name
  }
}
`

const testAccIPAMResourceDiscoveryConfig_baseAlternateDescription = `
data "aws_region" "current" {}

resource "aws_vpc_ipam_resource_discovery" "test" {
  description = "test ipam"
  operating_regions {
    region_name = data.aws_region.current.name
  }
}
`

func testAccIPAMResourceDiscoveryConfig_operatingRegion() string {
	return acctest.ConfigCompose(
		acctest.ConfigMultipleRegionProvider(2), `
data "aws_region" "current" {}

data "aws_region" "alternate" {
  provider = awsalternate
}

resource "aws_vpc_ipam_resource_discovery" "test" {
  description = "test"
  operating_regions {
    region_name = data.aws_region.current.name
  }
  operating_regions {
    region_name = data.aws_region.alternate.name
  }
}
`)
}

func testAccIPAMResourceDiscoveryConfig_tags(tagKey1, tagValue1 string) string {
	return fmt.Sprintf(`
data "aws_region" "current" {}

resource "aws_vpc_ipam_resource_discovery" "test" {
  description = "test"
  operating_regions {
    region_name = data.aws_region.current.name
  }
  tags = {
    %[1]q = %[2]q
  }
}
`, tagKey1, tagValue1)
}

func testAccIPAMResourceDiscoveryConfig_tags2(tagKey1, tagValue1, tagKey2, tagValue2 string) string {
	return fmt.Sprintf(`
data "aws_region" "current" {}

resource "aws_vpc_ipam_resource_discovery" "test" {
  description = "test"
  operating_regions {
    region_name = data.aws_region.current.name
  }
  tags = {
    %[1]q = %[2]q
    %[3]q = %[4]q
  }
}
	`, tagKey1, tagValue1, tagKey2, tagValue2)
}
