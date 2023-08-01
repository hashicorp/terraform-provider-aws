// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package ec2_test

import (
	"context"
	"fmt"
	"regexp"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/id"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfec2 "github.com/hashicorp/terraform-provider-aws/internal/service/ec2"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

func TestAccIPAM_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var ipam ec2.Ipam
	resourceName := "aws_vpc_ipam.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, ec2.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckIPAMDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccIPAMConfig_basic,
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckIPAMExists(ctx, resourceName, &ipam),
					resource.TestCheckResourceAttrSet(resourceName, "arn"),
					resource.TestCheckResourceAttr(resourceName, "description", ""),
					resource.TestCheckResourceAttr(resourceName, "operating_regions.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "scope_count", "2"),
					resource.TestMatchResourceAttr(resourceName, "private_default_scope_id", regexp.MustCompile(`^ipam-scope-[\da-f]+`)),
					resource.TestMatchResourceAttr(resourceName, "public_default_scope_id", regexp.MustCompile(`^ipam-scope-[\da-f]+`)),
					resource.TestMatchResourceAttr(resourceName, "default_resource_discovery_association_id", regexp.MustCompile(`^ipam-res-disco-assoc-[\da-f]+`)),
					resource.TestMatchResourceAttr(resourceName, "default_resource_discovery_id", regexp.MustCompile(`^ipam-res-disco-[\da-f]+`)),
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

func TestAccIPAM_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	var ipam ec2.Ipam
	resourceName := "aws_vpc_ipam.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, ec2.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckIPAMDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccIPAMConfig_basic,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckIPAMExists(ctx, resourceName, &ipam),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfec2.ResourceIPAM(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccIPAM_description(t *testing.T) {
	ctx := acctest.Context(t)
	var ipam ec2.Ipam
	resourceName := "aws_vpc_ipam.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, ec2.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckIPAMDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccIPAMConfig_description("test1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckIPAMExists(ctx, resourceName, &ipam),
					resource.TestCheckResourceAttr(resourceName, "description", "test1"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccIPAMConfig_description("test2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckIPAMExists(ctx, resourceName, &ipam),
					resource.TestCheckResourceAttr(resourceName, "description", "test2"),
				),
			},
		},
	})
}

func TestAccIPAM_operatingRegions(t *testing.T) {
	ctx := acctest.Context(t)
	var ipam ec2.Ipam
	resourceName := "aws_vpc_ipam.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckMultipleRegion(t, 2) },
		ErrorCheck:               acctest.ErrorCheck(t, ec2.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5FactoriesMultipleRegions(ctx, t, 2),
		CheckDestroy:             testAccCheckIPAMDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccIPAMConfig_twoOperatingRegions(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckIPAMExists(ctx, resourceName, &ipam),
					resource.TestCheckResourceAttr(resourceName, "operating_regions.#", "2"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccIPAMConfig_basic,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckIPAMExists(ctx, resourceName, &ipam),
					resource.TestCheckResourceAttr(resourceName, "operating_regions.#", "1"),
				),
			},
			{
				Config: testAccIPAMConfig_twoOperatingRegions(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckIPAMExists(ctx, resourceName, &ipam),
					resource.TestCheckResourceAttr(resourceName, "operating_regions.#", "2"),
				),
			},
		},
	})
}

func TestAccIPAM_cascade(t *testing.T) {
	ctx := acctest.Context(t)
	var ipam ec2.Ipam
	resourceName := "aws_vpc_ipam.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, ec2.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckIPAMDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccIPAMConfig_cascade,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckIPAMExists(ctx, resourceName, &ipam),
					testAccCheckIPAMScopeCreate(ctx, &ipam),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"cascade"},
			},
		},
	})
}

func TestAccIPAM_tags(t *testing.T) {
	ctx := acctest.Context(t)
	var ipam ec2.Ipam
	resourceName := "aws_vpc_ipam.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, ec2.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckIPAMDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccIPAMConfig_tags("key1", "value1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckIPAMExists(ctx, resourceName, &ipam),
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
				Config: testAccIPAMConfig_tags2("key1", "value1updated", "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckIPAMExists(ctx, resourceName, &ipam),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1updated"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
			{
				Config: testAccIPAMConfig_tags("key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckIPAMExists(ctx, resourceName, &ipam),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
		},
	})
}

func testAccCheckIPAMExists(ctx context.Context, n string, v *ec2.Ipam) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No IPAM ID is set")
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).EC2Conn(ctx)

		output, err := tfec2.FindIPAMByID(ctx, conn, rs.Primary.ID)

		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

func testAccCheckIPAMDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).EC2Conn(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_vpc_ipam" {
				continue
			}

			_, err := tfec2.FindIPAMByID(ctx, conn, rs.Primary.ID)

			if tfresource.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("IPAM still exists: %s", rs.Primary.ID)
		}

		return nil
	}
}

func testAccCheckIPAMScopeCreate(ctx context.Context, ipam *ec2.Ipam) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).EC2Conn(ctx)

		_, err := conn.CreateIpamScopeWithContext(ctx, &ec2.CreateIpamScopeInput{
			ClientToken: aws.String(id.UniqueId()),
			IpamId:      aws.String(*ipam.IpamId),
		})

		return err
	}
}

const testAccIPAMConfig_basic = `
data "aws_region" "current" {}

resource "aws_vpc_ipam" "test" {
  operating_regions {
    region_name = data.aws_region.current.name
  }
}
`

const testAccIPAMConfig_cascade = `
data "aws_region" "current" {}

resource "aws_vpc_ipam" "test" {
  operating_regions {
    region_name = data.aws_region.current.name
  }
  cascade = true
}
`

func testAccIPAMConfig_description(description string) string {
	return fmt.Sprintf(`
data "aws_region" "current" {}

resource "aws_vpc_ipam" "test" {
  description = %[1]q

  operating_regions {
    region_name = data.aws_region.current.name
  }
}
`, description)
}

func testAccIPAMConfig_twoOperatingRegions() string {
	return acctest.ConfigCompose(acctest.ConfigMultipleRegionProvider(2), `
data "aws_region" "current" {}

data "aws_region" "alternate" {
  provider = awsalternate
}

resource "aws_vpc_ipam" "test" {
  operating_regions {
    region_name = data.aws_region.current.name
  }

  operating_regions {
    region_name = data.aws_region.alternate.name
  }
}
`)
}

func testAccIPAMConfig_tags(tagKey1, tagValue1 string) string {
	return fmt.Sprintf(`
data "aws_region" "current" {}

resource "aws_vpc_ipam" "test" {
  operating_regions {
    region_name = data.aws_region.current.name
  }

  tags = {
    %[1]q = %[2]q
  }
}
`, tagKey1, tagValue1)
}

func testAccIPAMConfig_tags2(tagKey1, tagValue1, tagKey2, tagValue2 string) string {
	return fmt.Sprintf(`
data "aws_region" "current" {}

resource "aws_vpc_ipam" "test" {
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
