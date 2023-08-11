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

func testAccIPAMResourceDiscoveryAssociation_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var rda ec2.IpamResourceDiscoveryAssociation
	resourceName := "aws_vpc_ipam_resource_discovery_association.test"
	ipamName := "aws_vpc_ipam.test"
	rdName := "aws_vpc_ipam_resource_discovery.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, ec2.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckIPAMResourceDiscoveryAssociationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccIPAMResourceDiscoveryAssociationConfig_basic(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckIPAMResourceDiscoveryAssociationExists(ctx, resourceName, &rda),
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

func testAccIPAMResourceDiscoveryAssociation_tags(t *testing.T) {
	ctx := acctest.Context(t)
	var rda ec2.IpamResourceDiscoveryAssociation
	resourceName := "aws_vpc_ipam_resource_discovery_association.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, ec2.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckIPAMResourceDiscoveryAssociationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccIPAMResourceDiscoveryAssociationConfig_tags("key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckIPAMResourceDiscoveryAssociationExists(ctx, resourceName, &rda),
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

func testAccIPAMResourceDiscoveryAssociation_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	var rda ec2.IpamResourceDiscoveryAssociation
	resourceName := "aws_vpc_ipam_resource_discovery_association.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, ec2.EndpointsID),
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

func testAccCheckIPAMResourceDiscoveryAssociationExists(ctx context.Context, n string, v *ec2.IpamResourceDiscoveryAssociation) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No IPAM Resource Discovery Association ID is set")
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).EC2Conn(ctx)

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
		conn := acctest.Provider.Meta().(*conns.AWSClient).EC2Conn(ctx)

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
