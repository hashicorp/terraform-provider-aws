// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package meta_test

import (
	"fmt"
	"net"
	"sort"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/YakDriver/regexache"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	tfmeta "github.com/hashicorp/terraform-provider-aws/internal/service/meta"
)

func TestAccMetaIPRangesDataSource_basic(t *testing.T) {
	ctx := acctest.Context(t)
	dataSourceName := "data.aws_ip_ranges.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, tfmeta.PseudoServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccIPRangesDataSourceConfig_basic,
				Check: resource.ComposeTestCheckFunc(
					testAccIPRangesCheckAttributes(dataSourceName),
					testAccIPRangesCheckCIDRBlocksAttribute(dataSourceName, "cidr_blocks"),
					testAccIPRangesCheckCIDRBlocksAttribute(dataSourceName, "ipv6_cidr_blocks"),
				),
			},
		},
	})
}

func TestAccMetaIPRangesDataSource_none(t *testing.T) {
	ctx := acctest.Context(t)
	dataSourceName := "data.aws_ip_ranges.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, tfmeta.PseudoServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccIPRangesDataSourceConfig_none,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(dataSourceName, "cidr_blocks.#", acctest.Ct0),
					resource.TestCheckResourceAttr(dataSourceName, "ipv6_cidr_blocks.#", acctest.Ct0),
				),
			},
		},
	})
}

func TestAccMetaIPRangesDataSource_url(t *testing.T) {
	ctx := acctest.Context(t)
	dataSourceName := "data.aws_ip_ranges.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, tfmeta.PseudoServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccIPRangesDataSourceConfig_url,
				Check: resource.ComposeTestCheckFunc(
					testAccIPRangesCheckAttributes(dataSourceName),
					testAccIPRangesCheckCIDRBlocksAttribute(dataSourceName, "cidr_blocks"),
					testAccIPRangesCheckCIDRBlocksAttribute(dataSourceName, "ipv6_cidr_blocks"),
				),
			},
		},
	})
}

func TestAccMetaIPRangesDataSource_uppercase(t *testing.T) {
	ctx := acctest.Context(t)
	dataSourceName := "data.aws_ip_ranges.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, tfmeta.PseudoServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccIPRangesDataSourceConfig_uppercase,
				Check: resource.ComposeTestCheckFunc(
					testAccIPRangesCheckAttributes(dataSourceName),
					testAccIPRangesCheckCIDRBlocksAttribute(dataSourceName, "cidr_blocks"),
					testAccIPRangesCheckCIDRBlocksAttribute(dataSourceName, "ipv6_cidr_blocks"),
				),
			},
		},
	})
}

func testAccIPRangesCheckAttributes(n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		r := s.RootModule().Resources[n]
		a := r.Primary.Attributes

		var (
			createDate time.Time
			err        error
			syncToken  int
		)

		if createDate, err = time.Parse("2006-01-02-15-04-05", a["create_date"]); err != nil {
			return err
		}

		if syncToken, err = strconv.Atoi(a["sync_token"]); err != nil {
			return err
		}

		if syncToken != int(createDate.Unix()) {
			return fmt.Errorf("sync_token %d does not match create_date %s", syncToken, createDate)
		}

		var (
			regionMember      = regexache.MustCompile(`regions\.\d+`)
			regions, services int
			serviceMember     = regexache.MustCompile(`services\.\d+`)
		)

		for k, v := range a {
			if regionMember.MatchString(k) {
				// lintignore:AWSAT003
				if v := strings.ToLower(v); !(v == "eu-west-1" || v == "eu-central-1") {
					return fmt.Errorf("unexpected region %s", v)
				}

				regions = regions + 1
			}

			if serviceMember.MatchString(k) {
				if v := strings.ToLower(v); !(v == "ec2" || v == "amazon") {
					return fmt.Errorf("unexpected service %s", v)
				}

				services = services + 1
			}
		}

		if regions != 2 {
			return fmt.Errorf("unexpected number of regions: %d", regions)
		}

		if services != 1 {
			return fmt.Errorf("unexpected number of services: %d", services)
		}

		return nil
	}
}

func testAccIPRangesCheckCIDRBlocksAttribute(name, attribute string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		r := s.RootModule().Resources[name]
		a := r.Primary.Attributes

		var (
			cidrBlockSize int
			cidrBlocks    sort.StringSlice
			err           error
		)

		if cidrBlockSize, err = strconv.Atoi(a[fmt.Sprintf("%s.#", attribute)]); err != nil {
			return err
		}

		if cidrBlockSize < 5 {
			return fmt.Errorf("%s for eu-west-1 seem suspiciously low: %d", attribute, cidrBlockSize) // lintignore:AWSAT003
		}

		cidrBlocks = make([]string, cidrBlockSize)

		for i := range cidrBlocks {
			cidrBlock := a[fmt.Sprintf("%s.%d", attribute, i)]

			_, _, err := net.ParseCIDR(cidrBlock)
			if err != nil {
				return fmt.Errorf("malformed CIDR block %s in %s: %s", cidrBlock, attribute, err)
			}

			cidrBlocks[i] = cidrBlock
		}

		if !sort.IsSorted(cidrBlocks) {
			return fmt.Errorf("unexpected order of %s: %s", attribute, cidrBlocks)
		}

		return nil
	}
}

// lintignore:AWSAT003
const testAccIPRangesDataSourceConfig_basic = `
data "aws_ip_ranges" "test" {
  regions  = ["eu-west-1", "eu-central-1"]
  services = ["ec2"]
}
`

const testAccIPRangesDataSourceConfig_none = `
data "aws_ip_ranges" "test" {
  regions  = ["mars-1"]
  services = ["blueorigin"]
}
`

// lintignore:AWSAT003
const testAccIPRangesDataSourceConfig_url = `
data "aws_ip_ranges" "test" {
  regions  = ["eu-west-1", "eu-central-1"]
  services = ["ec2"]
  url      = "https://ip-ranges.amazonaws.com/ip-ranges.json"
}
`

// lintignore:AWSAT003
const testAccIPRangesDataSourceConfig_uppercase = `
data "aws_ip_ranges" "test" {
  regions  = ["EU-WEST-1", "EU-CENTRAL-1"]
  services = ["AMAZON"]
}
`
