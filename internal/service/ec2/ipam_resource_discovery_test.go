// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package ec2_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/YakDriver/regexache"
	awstypes "github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfec2 "github.com/hashicorp/terraform-provider-aws/internal/service/ec2"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccIPAMResourceDiscovery_serial(t *testing.T) {
	t.Parallel()

	testCases := map[string]map[string]func(t *testing.T){
		"ResourceDiscovery": {
			acctest.CtBasic:      testAccIPAMResourceDiscovery_basic,
			"modify":             testAccIPAMResourceDiscovery_modify,
			acctest.CtDisappears: testAccIPAMResourceDiscovery_disappears,
			"tags":               testAccIPAMResourceDiscovery_tags,
		},
		"ResourceDiscoveryAssociation": {
			acctest.CtBasic:      testAccIPAMResourceDiscoveryAssociation_basic,
			acctest.CtDisappears: testAccIPAMResourceDiscoveryAssociation_disappears,
			"tags":               testAccIPAMResourceDiscoveryAssociation_tags,
		},
	}

	acctest.RunSerialTests2Levels(t, testCases, 0)
}

func testAccIPAMResourceDiscovery_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var rd awstypes.IpamResourceDiscovery
	resourceName := "aws_vpc_ipam_resource_discovery.test"
	dataSourceRegion := "data.aws_region.current"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckIPAMResourceDiscoveryDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccIPAMResourceDiscoveryConfig_base,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckIPAMResourceDiscoveryExists(ctx, resourceName, &rd),
					acctest.MatchResourceAttrGlobalARN(resourceName, names.AttrARN, "ec2", regexache.MustCompile(`ipam-resource-discovery/ipam-res-disco-[0-9a-f]+$`)),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, "test"),
					resource.TestCheckResourceAttrPair(resourceName, "ipam_resource_discovery_region", dataSourceRegion, names.AttrName),
					resource.TestCheckResourceAttr(resourceName, "is_default", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "operating_regions.#", acctest.Ct1),
					acctest.CheckResourceAttrAccountID(resourceName, names.AttrOwnerID),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct0),
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
	var rd awstypes.IpamResourceDiscovery
	resourceName := "aws_vpc_ipam_resource_discovery.test"

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckMultipleRegion(t, 2)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5FactoriesMultipleRegions(ctx, t, 2),
		CheckDestroy:             testAccCheckIPAMResourceDiscoveryDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccIPAMResourceDiscoveryConfig_base,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckIPAMResourceDiscoveryExists(ctx, resourceName, &rd),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, "test"),
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
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, "test"),
				),
			},
			{
				Config: testAccIPAMResourceDiscoveryConfig_base,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, "test"),
				),
			},
			{
				Config: testAccIPAMResourceDiscoveryConfig_baseAlternateDescription,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, "test ipam"),
				),
			},
		},
	})
}

func testAccIPAMResourceDiscovery_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	var rd awstypes.IpamResourceDiscovery
	resourceName := "aws_vpc_ipam_resource_discovery.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
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
	var rd awstypes.IpamResourceDiscovery
	resourceName := "aws_vpc_ipam_resource_discovery.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckIPAMResourceDiscoveryDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccIPAMResourceDiscoveryConfig_tags(acctest.CtKey1, acctest.CtValue1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckIPAMResourceDiscoveryExists(ctx, resourceName, &rd),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey1, acctest.CtValue1),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccIPAMResourceDiscoveryConfig_tags2(acctest.CtKey1, acctest.CtValue1Updated, acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct2),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey1, acctest.CtValue1Updated),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey2, acctest.CtValue2),
				),
			},
			{
				Config: testAccIPAMResourceDiscoveryConfig_tags(acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey2, acctest.CtValue2),
				),
			},
		},
	})
}

func testAccCheckIPAMResourceDiscoveryExists(ctx context.Context, n string, v *awstypes.IpamResourceDiscovery) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No IPAM Resource Discovery ID is set")
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).EC2Client(ctx)

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
		conn := acctest.Provider.Meta().(*conns.AWSClient).EC2Client(ctx)

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
