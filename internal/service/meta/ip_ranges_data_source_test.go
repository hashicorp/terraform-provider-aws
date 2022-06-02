package meta_test

import (
	"fmt"
	"net"
	"regexp"
	"sort"
	"strconv"
	"testing"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	tfmeta "github.com/hashicorp/terraform-provider-aws/internal/service/meta"
)

func TestAccMetaIPRangesDataSource_basic(t *testing.T) {
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, tfmeta.PseudoServiceID),
		ProviderFactories: acctest.ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccIPRangesConfig,
				Check: resource.ComposeTestCheckFunc(
					testAccIPRangesCheckAttributes("data.aws_ip_ranges.some"),
					testAccIPRangesCheckCIDRBlocksAttribute("data.aws_ip_ranges.some", "cidr_blocks"),
					testAccIPRangesCheckCIDRBlocksAttribute("data.aws_ip_ranges.some", "ipv6_cidr_blocks"),
					resource.TestCheckResourceAttr("data.aws_ip_ranges.none", "cidr_blocks.#", "0"),
					resource.TestCheckResourceAttr("data.aws_ip_ranges.none", "ipv6_cidr_blocks.#", "0"),
				),
			},
		},
	})
}

func TestAccMetaIPRangesDataSource_url(t *testing.T) {
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, tfmeta.PseudoServiceID),
		ProviderFactories: acctest.ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccIPRangesURLConfig,
				Check: resource.ComposeTestCheckFunc(
					testAccIPRangesCheckAttributes("data.aws_ip_ranges.some"),
					testAccIPRangesCheckCIDRBlocksAttribute("data.aws_ip_ranges.some", "cidr_blocks"),
					testAccIPRangesCheckCIDRBlocksAttribute("data.aws_ip_ranges.some", "ipv6_cidr_blocks"),
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
			regionMember      = regexp.MustCompile(`regions\.\d+`)
			regions, services int
			serviceMember     = regexp.MustCompile(`services\.\d+`)
		)

		for k, v := range a {

			if regionMember.MatchString(k) {

				// lintignore:AWSAT003
				if !(v == "eu-west-1" || v == "eu-central-1") {
					return fmt.Errorf("unexpected region %s", v)
				}

				regions = regions + 1

			}

			if serviceMember.MatchString(k) {

				if v != "ec2" {
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
const testAccIPRangesConfig = `
data "aws_ip_ranges" "some" {
  regions  = ["eu-west-1", "eu-central-1"]
  services = ["ec2"]
}

data "aws_ip_ranges" "none" {
  regions  = ["mars-1"]
  services = ["blueorigin"]
}
`

// lintignore:AWSAT003
const testAccIPRangesURLConfig = `
data "aws_ip_ranges" "some" {
  regions  = ["eu-west-1", "eu-central-1"]
  services = ["ec2"]
  url      = "https://ip-ranges.amazonaws.com/ip-ranges.json"
}
`
