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

func testAccIPAMResourceDiscoveryAssociation_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var rda awstypes.IpamResourceDiscoveryAssociation
	resourceName := "aws_vpc_ipam_resource_discovery_association.test"
	ipamName := "aws_vpc_ipam.test"
	rdName := "aws_vpc_ipam_resource_discovery.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckIPAMResourceDiscoveryAssociationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccIPAMResourceDiscoveryAssociationConfig_basic(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckIPAMResourceDiscoveryAssociationExists(ctx, resourceName, &rda),
					acctest.MatchResourceAttrGlobalARN(resourceName, names.AttrARN, "ec2", regexache.MustCompile(`ipam-resource-discovery-association/ipam-res-disco-assoc-[0-9a-f]+$`)),
					resource.TestCheckResourceAttrPair(resourceName, "ipam_id", ipamName, names.AttrID),
					resource.TestCheckResourceAttrPair(resourceName, "ipam_resource_discovery_id", rdName, names.AttrID),
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

func testAccIPAMResourceDiscoveryAssociation_tags(t *testing.T) {
	ctx := acctest.Context(t)
	var rda awstypes.IpamResourceDiscoveryAssociation
	resourceName := "aws_vpc_ipam_resource_discovery_association.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckIPAMResourceDiscoveryAssociationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccIPAMResourceDiscoveryAssociationConfig_tags(acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckIPAMResourceDiscoveryAssociationExists(ctx, resourceName, &rda),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey2, acctest.CtValue2),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccIPAMResourceDiscoveryAssociationConfig_tags2(acctest.CtKey1, acctest.CtValue1Updated, acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct2),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey1, acctest.CtValue1Updated),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey2, acctest.CtValue2),
				),
			},
			{
				Config: testAccIPAMResourceDiscoveryAssociationConfig_tags(acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey2, acctest.CtValue2),
				),
			},
		},
	})
}

func testAccIPAMResourceDiscoveryAssociation_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	var rda awstypes.IpamResourceDiscoveryAssociation
	resourceName := "aws_vpc_ipam_resource_discovery_association.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckIPAMResourceDiscoveryAssociationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccIPAMResourceDiscoveryAssociationConfig_basic(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckIPAMResourceDiscoveryAssociationExists(ctx, resourceName, &rda),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfec2.ResourceIPAMResourceDiscoveryAssociation(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckIPAMResourceDiscoveryAssociationExists(ctx context.Context, n string, v *awstypes.IpamResourceDiscoveryAssociation) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No IPAM Resource Discovery Association ID is set")
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).EC2Client(ctx)

		output, err := tfec2.FindIPAMResourceDiscoveryAssociationByID(ctx, conn, rs.Primary.ID)

		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

func testAccCheckIPAMResourceDiscoveryAssociationDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).EC2Client(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_vpc_ipam_resource_discovery_association" {
				continue
			}

			_, err := tfec2.FindIPAMResourceDiscoveryAssociationByID(ctx, conn, rs.Primary.ID)

			if tfresource.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("IPAM Resource Discovery Association still exists: %s", rs.Primary.ID)
		}

		return nil
	}
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
