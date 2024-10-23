// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package wafv2_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/YakDriver/regexache"
	awstypes "github.com/aws/aws-sdk-go-v2/service/wafv2/types"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfwafv2 "github.com/hashicorp/terraform-provider-aws/internal/service/wafv2"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccWAFV2IPSet_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var v awstypes.IPSet
	ipSetName := fmt.Sprintf("ip-set-%s", sdkacctest.RandString(5))
	resourceName := "aws_wafv2_ip_set.ip_set"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheckScopeRegional(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.WAFV2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckIPSetDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccIPSetConfig_basic(ipSetName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckIPSetExists(ctx, resourceName, &v),
					acctest.MatchResourceAttrRegionalARN(resourceName, names.AttrARN, "wafv2", regexache.MustCompile(`regional/ipset/.+$`)),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, ipSetName),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, ipSetName),
					resource.TestCheckResourceAttr(resourceName, names.AttrScope, string(awstypes.ScopeRegional)),
					resource.TestCheckResourceAttr(resourceName, "ip_address_version", string(awstypes.IPAddressVersionIpv4)),
					resource.TestCheckResourceAttr(resourceName, "addresses.#", acctest.Ct2),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct2),
					resource.TestCheckResourceAttr(resourceName, "tags.Tag1", "Value1"),
					resource.TestCheckResourceAttr(resourceName, "tags.Tag2", "Value2"),
				),
			},
			{
				Config: testAccIPSetConfig_update(ipSetName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckIPSetExists(ctx, resourceName, &v),
					acctest.MatchResourceAttrRegionalARN(resourceName, names.AttrARN, "wafv2", regexache.MustCompile(`regional/ipset/.+$`)),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, ipSetName),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, "Updated"),
					resource.TestCheckResourceAttr(resourceName, names.AttrScope, string(awstypes.ScopeRegional)),
					resource.TestCheckResourceAttr(resourceName, "ip_address_version", string(awstypes.IPAddressVersionIpv4)),
					resource.TestCheckResourceAttr(resourceName, "addresses.#", acctest.Ct3),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateIdFunc: testAccIPSetImportStateIdFunc(resourceName),
			},
		},
	})
}

func TestAccWAFV2IPSet_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	var r awstypes.IPSet
	ipSetName := fmt.Sprintf("ip-set-%s", sdkacctest.RandString(5))
	resourceName := "aws_wafv2_ip_set.ip_set"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheckScopeRegional(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.WAFV2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckIPSetDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccIPSetConfig_basic(ipSetName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckIPSetExists(ctx, resourceName, &r),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfwafv2.ResourceIPSet(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccWAFV2IPSet_ipv6(t *testing.T) {
	ctx := acctest.Context(t)
	var v awstypes.IPSet
	ipSetName := fmt.Sprintf("ip-set-%s", sdkacctest.RandString(5))
	resourceName := "aws_wafv2_ip_set.ip_set"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheckScopeRegional(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.WAFV2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckIPSetDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccIPSetConfig_v6(ipSetName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckIPSetExists(ctx, resourceName, &v),
					acctest.MatchResourceAttrRegionalARN(resourceName, names.AttrARN, "wafv2", regexache.MustCompile(`regional/ipset/.+$`)),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, ipSetName),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, ipSetName),
					resource.TestCheckResourceAttr(resourceName, names.AttrScope, string(awstypes.ScopeRegional)),
					resource.TestCheckResourceAttr(resourceName, "ip_address_version", string(awstypes.IPAddressVersionIpv6)),
					resource.TestCheckResourceAttr(resourceName, "addresses.#", acctest.Ct3),
					resource.TestCheckTypeSetElemAttr(resourceName, "addresses.*", "1234:5678:9abc:6811:0000:0000:0000:0000/64"),
					resource.TestCheckTypeSetElemAttr(resourceName, "addresses.*", "2001:0db8:0000:0000:0000:0000:0000:0000/32"),
					resource.TestCheckTypeSetElemAttr(resourceName, "addresses.*", "1111:0000:0000:0000:0000:0000:0000:0111/128"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateIdFunc: testAccIPSetImportStateIdFunc(resourceName),
			},
		},
	})
}

func TestAccWAFV2IPSet_minimal(t *testing.T) {
	ctx := acctest.Context(t)
	var v awstypes.IPSet
	ipSetName := fmt.Sprintf("ip-set-%s", sdkacctest.RandString(5))
	resourceName := "aws_wafv2_ip_set.ip_set"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheckScopeRegional(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.WAFV2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckIPSetDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccIPSetConfig_minimal(ipSetName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckIPSetExists(ctx, resourceName, &v),
					acctest.MatchResourceAttrRegionalARN(resourceName, names.AttrARN, "wafv2", regexache.MustCompile(`regional/ipset/.+$`)),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, ipSetName),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, ""),
					resource.TestCheckResourceAttr(resourceName, names.AttrScope, string(awstypes.ScopeRegional)),
					resource.TestCheckResourceAttr(resourceName, "ip_address_version", string(awstypes.IPAddressVersionIpv4)),
					resource.TestCheckResourceAttr(resourceName, "addresses.#", acctest.Ct0),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateIdFunc: testAccIPSetImportStateIdFunc(resourceName),
			},
		},
	})
}

func TestAccWAFV2IPSet_changeNameForceNew(t *testing.T) {
	ctx := acctest.Context(t)
	var before, after awstypes.IPSet
	ipSetName := fmt.Sprintf("ip-set-%s", sdkacctest.RandString(5))
	ipSetNewName := fmt.Sprintf("ip-set-%s", sdkacctest.RandString(5))
	resourceName := "aws_wafv2_ip_set.ip_set"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheckScopeRegional(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.WAFV2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckIPSetDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccIPSetConfig_basic(ipSetName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckIPSetExists(ctx, resourceName, &before),
					acctest.MatchResourceAttrRegionalARN(resourceName, names.AttrARN, "wafv2", regexache.MustCompile(`regional/ipset/.+$`)),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, ipSetName),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, ipSetName),
					resource.TestCheckResourceAttr(resourceName, names.AttrScope, string(awstypes.ScopeRegional)),
					resource.TestCheckResourceAttr(resourceName, "ip_address_version", string(awstypes.IPAddressVersionIpv4)),
					resource.TestCheckResourceAttr(resourceName, "addresses.#", acctest.Ct2),
				),
			},
			{
				Config: testAccIPSetConfig_basic(ipSetNewName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckIPSetExists(ctx, resourceName, &after),
					acctest.MatchResourceAttrRegionalARN(resourceName, names.AttrARN, "wafv2", regexache.MustCompile(`regional/ipset/.+$`)),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, ipSetNewName),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, ipSetNewName),
					resource.TestCheckResourceAttr(resourceName, names.AttrScope, string(awstypes.ScopeRegional)),
					resource.TestCheckResourceAttr(resourceName, "ip_address_version", string(awstypes.IPAddressVersionIpv4)),
					resource.TestCheckResourceAttr(resourceName, "addresses.#", acctest.Ct2),
				),
			},
		},
	})
}

func TestAccWAFV2IPSet_addresses(t *testing.T) {
	ctx := acctest.Context(t)
	var v awstypes.IPSet
	ipSetName := fmt.Sprintf("ip-set-%s", sdkacctest.RandString(5))
	resourceName := "aws_wafv2_ip_set.ip_set"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheckScopeRegional(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.WAFV2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckIPSetDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccIPSetConfig_addresses(ipSetName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckIPSetExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "addresses.#", acctest.Ct2),
				),
			},
		},
	})
}

func TestAccWAFV2IPSet_tags(t *testing.T) {
	ctx := acctest.Context(t)
	var v awstypes.IPSet
	ipSetName := fmt.Sprintf("ip-set-%s", sdkacctest.RandString(5))
	resourceName := "aws_wafv2_ip_set.ip_set"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheckScopeRegional(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.WAFV2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckIPSetDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccIPSetConfig_oneTag(ipSetName, "Tag1", "Value1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckIPSetExists(ctx, resourceName, &v),
					acctest.MatchResourceAttrRegionalARN(resourceName, names.AttrARN, "wafv2", regexache.MustCompile(`regional/ipset/.+$`)),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "tags.Tag1", "Value1"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateIdFunc: testAccIPSetImportStateIdFunc(resourceName),
			},
			{
				Config: testAccIPSetConfig_twoTags(ipSetName, "Tag1", "Value1Updated", "Tag2", "Value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckIPSetExists(ctx, resourceName, &v),
					acctest.MatchResourceAttrRegionalARN(resourceName, names.AttrARN, "wafv2", regexache.MustCompile(`regional/ipset/.+$`)),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct2),
					resource.TestCheckResourceAttr(resourceName, "tags.Tag1", "Value1Updated"),
					resource.TestCheckResourceAttr(resourceName, "tags.Tag2", "Value2"),
				),
			},
			{
				Config: testAccIPSetConfig_oneTag(ipSetName, "Tag2", "Value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckIPSetExists(ctx, resourceName, &v),
					acctest.MatchResourceAttrRegionalARN(resourceName, names.AttrARN, "wafv2", regexache.MustCompile(`regional/ipset/.+$`)),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "tags.Tag2", "Value2"),
				),
			},
		},
	})
}

func TestAccWAFV2IPSet_large(t *testing.T) {
	ctx := acctest.Context(t)
	var v awstypes.IPSet
	ipSetName := fmt.Sprintf("ip-set-%s", sdkacctest.RandString(5))
	resourceName := "aws_wafv2_ip_set.ip_set"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheckScopeRegional(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.WAFV2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckIPSetDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccIPSetConfig_large(ipSetName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckIPSetExists(ctx, resourceName, &v),
					acctest.MatchResourceAttrRegionalARN(resourceName, names.AttrARN, "wafv2", regexache.MustCompile(`regional/ipset/.+$`)),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, ipSetName),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, ipSetName),
					resource.TestCheckResourceAttr(resourceName, names.AttrScope, string(awstypes.ScopeRegional)),
					resource.TestCheckResourceAttr(resourceName, "ip_address_version", string(awstypes.IPAddressVersionIpv4)),
					resource.TestCheckResourceAttr(resourceName, "addresses.#", "50"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateIdFunc: testAccIPSetImportStateIdFunc(resourceName),
			},
		},
	})
}

func testAccCheckIPSetDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_wafv2_ip_set" {
				continue
			}

			conn := acctest.Provider.Meta().(*conns.AWSClient).WAFV2Client(ctx)

			_, err := tfwafv2.FindIPSetByThreePartKey(ctx, conn, rs.Primary.ID, rs.Primary.Attributes[names.AttrName], rs.Primary.Attributes[names.AttrScope])

			if tfresource.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("WAFv2 IPSet %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccCheckIPSetExists(ctx context.Context, n string, v *awstypes.IPSet) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No WAFv2 IPSet ID is set")
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).WAFV2Client(ctx)

		output, err := tfwafv2.FindIPSetByThreePartKey(ctx, conn, rs.Primary.ID, rs.Primary.Attributes[names.AttrName], rs.Primary.Attributes[names.AttrScope])

		if err != nil {
			return err
		}

		*v = *output.IPSet

		return nil
	}
}

func testAccIPSetConfig_basic(name string) string {
	return fmt.Sprintf(`
resource "aws_wafv2_ip_set" "ip_set" {
  name               = %[1]q
  description        = %[1]q
  scope              = "REGIONAL"
  ip_address_version = "IPV4"
  addresses          = ["1.2.3.4/32", "5.6.7.8/32"]

  tags = {
    Tag1 = "Value1"
    Tag2 = "Value2"
  }
}
`, name)
}

func testAccIPSetConfig_addresses(name string) string {
	return fmt.Sprintf(`
resource "aws_eip" "test" {
  domain = "vpc"
}

resource "aws_wafv2_ip_set" "ip_set" {
  name               = %[1]q
  description        = %[1]q
  scope              = "REGIONAL"
  ip_address_version = "IPV4"
  addresses          = ["1.2.3.4/32", "${aws_eip.test.public_ip}/32"]

  tags = {
    Tag1 = "Value1"
    Tag2 = "Value2"
  }
}
`, name)
}

func testAccIPSetConfig_update(name string) string {
	return fmt.Sprintf(`
resource "aws_wafv2_ip_set" "ip_set" {
  name               = %[1]q
  description        = "Updated"
  scope              = "REGIONAL"
  ip_address_version = "IPV4"
  addresses          = ["1.1.1.1/32", "2.2.2.2/32", "3.3.3.3/32"]
}
`, name)
}

func testAccIPSetConfig_v6(name string) string {
	return fmt.Sprintf(`
resource "aws_wafv2_ip_set" "ip_set" {
  name               = %[1]q
  description        = %[1]q
  scope              = "REGIONAL"
  ip_address_version = "IPV6"
  addresses = [
    "1111:0000:0000:0000:0000:0000:0000:0111/128",
    "1234:5678:9abc:6811:0000:0000:0000:0000/64",
    "2001:0db8:0000:0000:0000:0000:0000:0000/32"
  ]
}
`, name)
}

func testAccIPSetConfig_minimal(name string) string {
	return fmt.Sprintf(`
resource "aws_wafv2_ip_set" "ip_set" {
  name               = %[1]q
  scope              = "REGIONAL"
  ip_address_version = "IPV4"
}
`, name)
}

func testAccIPSetConfig_oneTag(name, tagKey, tagValue string) string {
	return fmt.Sprintf(`
resource "aws_wafv2_ip_set" "ip_set" {
  name               = %[1]q
  description        = %[1]q
  scope              = "REGIONAL"
  ip_address_version = "IPV4"
  addresses          = ["1.2.3.4/32", "5.6.7.8/32"]

  tags = {
    %[2]q = %[3]q
  }
}
`, name, tagKey, tagValue)
}

func testAccIPSetConfig_twoTags(name, tag1Key, tag1Value, tag2Key, tag2Value string) string {
	return fmt.Sprintf(`
resource "aws_wafv2_ip_set" "ip_set" {
  name               = %[1]q
  description        = %[1]q
  scope              = "REGIONAL"
  ip_address_version = "IPV4"
  addresses          = ["1.2.3.4/32", "5.6.7.8/32"]

  tags = {
    %[2]q = %[3]q
    %[4]q = %[5]q
  }
}
`, name, tag1Key, tag1Value, tag2Key, tag2Value)
}

func testAccIPSetConfig_large(name string) string {
	return fmt.Sprintf(`
resource "aws_wafv2_ip_set" "ip_set" {
  name               = %[1]q
  description        = %[1]q
  scope              = "REGIONAL"
  ip_address_version = "IPV4"
  addresses = [
    "1.1.1.50/32", "1.1.1.73/32", "1.1.1.15/32", "2.2.2.30/32", "1.1.1.38/32",
    "2.2.2.53/32", "1.1.1.21/32", "2.2.2.24/32", "1.1.1.44/32", "1.1.1.1/32",
    "1.1.1.67/32", "2.2.2.76/32", "2.2.2.99/32", "1.1.1.26/32", "2.2.2.93/32",
    "2.2.2.64/32", "1.1.1.32/32", "2.2.2.12/32", "2.2.2.47/32", "1.1.1.91/32",
    "1.1.1.78/32", "2.2.2.82/32", "2.2.2.58/32", "1.1.1.85/32", "2.2.2.4/32",
    "2.2.2.65/32", "2.2.2.23/32", "2.2.2.17/32", "2.2.2.42/32", "1.1.1.56/32",
    "1.1.1.79/32", "2.2.2.81/32", "2.2.2.36/32", "2.2.2.59/32", "2.2.2.9/32",
    "1.1.1.7/32", "1.1.1.84/32", "1.1.1.51/32", "2.2.2.70/32", "2.2.2.87/32",
    "1.1.1.39/32", "1.1.1.90/32", "2.2.2.31/32", "1.1.1.62/32", "1.1.1.14/32",
    "1.1.1.20/32", "2.2.2.25/32", "1.1.1.45/32", "1.1.1.2/32", "2.2.2.98/32"
  ]
}
`, name)
}

func testAccIPSetImportStateIdFunc(resourceName string) resource.ImportStateIdFunc {
	return func(s *terraform.State) (string, error) {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return "", fmt.Errorf("Not found: %s", resourceName)
		}

		return fmt.Sprintf("%s/%s/%s", rs.Primary.ID, rs.Primary.Attributes[names.AttrName], rs.Primary.Attributes[names.AttrScope]), nil
	}
}
