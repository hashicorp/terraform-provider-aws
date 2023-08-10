// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package ec2_test

import (
	"context"
	"fmt"
	"regexp"
	"testing"

	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfec2 "github.com/hashicorp/terraform-provider-aws/internal/service/ec2"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

func TestAccIPAMResourceDiscovery_serial(t *testing.T) {
	t.Parallel()

	testCases := map[string]map[string]func(t *testing.T){
		"ResourceDiscovery": {
			"basic":      testAccIPAMResourceDiscovery_basic,
			"modify":     testAccIPAMResourceDiscovery_modify,
			"disappears": testAccIPAMResourceDiscovery_disappears,
			"tags":       testAccIPAMResourceDiscovery_tags,
		},
		"ResourceDiscoveryAssociation": {
			"basic":      testAccIPAMResourceDiscoveryAssociation_basic,
			"disappears": testAccIPAMResourceDiscoveryAssociation_disappears,
			"tags":       testAccIPAMResourceDiscoveryAssociation_tags,
		},
	}

	acctest.RunSerialTests2Levels(t, testCases, 0)
}

func testAccIPAMResourceDiscovery_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var rd ec2.IpamResourceDiscovery
	resourceName := "aws_vpc_ipam_resource_discovery.test"
	dataSourceRegion := "data.aws_region.current"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, ec2.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckIPAMResourceDiscoveryDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccIPAMResourceDiscoveryConfig_base,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckIPAMResourceDiscoveryExists(ctx, resourceName, &rd),
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

func testAccIPAMResourceDiscovery_modify(t *testing.T) {
	ctx := acctest.Context(t)
	var rd ec2.IpamResourceDiscovery
	resourceName := "aws_vpc_ipam_resource_discovery.test"

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckMultipleRegion(t, 2)
		},
		ErrorCheck:               acctest.ErrorCheck(t, ec2.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5FactoriesMultipleRegions(ctx, t, 2),
		CheckDestroy:             testAccCheckIPAMResourceDiscoveryDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccIPAMResourceDiscoveryConfig_base,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckIPAMResourceDiscoveryExists(ctx, resourceName, &rd),
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

func testAccIPAMResourceDiscovery_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	var rd ec2.IpamResourceDiscovery
	resourceName := "aws_vpc_ipam_resource_discovery.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, ec2.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckIPAMResourceDiscoveryDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccIPAMResourceDiscoveryConfig_base,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckIPAMResourceDiscoveryExists(ctx, resourceName, &rd),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfec2.ResourceIPAMResourceDiscovery(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccIPAMResourceDiscovery_tags(t *testing.T) {
	ctx := acctest.Context(t)
	var rd ec2.IpamResourceDiscovery
	resourceName := "aws_vpc_ipam_resource_discovery.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, ec2.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckIPAMResourceDiscoveryDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccIPAMResourceDiscoveryConfig_tags("key1", "value1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckIPAMResourceDiscoveryExists(ctx, resourceName, &rd),
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

func testAccCheckIPAMResourceDiscoveryExists(ctx context.Context, n string, v *ec2.IpamResourceDiscovery) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No IPAM Resource Discovery ID is set")
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).EC2Conn(ctx)

		output, err := tfec2.FindIPAMResourceDiscoveryByID(ctx, conn, rs.Primary.ID)

		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

func testAccCheckIPAMResourceDiscoveryDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).EC2Conn(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_vpc_ipam_resource_discovery" {
				continue
			}

			_, err := tfec2.FindIPAMResourceDiscoveryByID(ctx, conn, rs.Primary.ID)

			if tfresource.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("IPAM Resource Discovery still exists: %s", rs.Primary.ID)
		}

		return nil
	}
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
