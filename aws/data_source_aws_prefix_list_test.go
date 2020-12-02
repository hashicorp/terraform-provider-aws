package aws

import (
	"fmt"
	"regexp"
	"strconv"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccDataSourceAwsPrefixList_basic(t *testing.T) {
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceAwsPrefixListConfig,
				Check: resource.ComposeTestCheckFunc(
					testAccDataSourceAwsPrefixListCheck("data.aws_prefix_list.s3_by_id"),
					testAccDataSourceAwsPrefixListCheck("data.aws_prefix_list.s3_by_name"),
					resource.TestMatchResourceAttr("data.aws_prefix_list.s3_by_name", "id", regexp.MustCompile(`^pl-[0-9a-z]{8}$`)),
					testAccCheckResourceAttrRegionalReverseDnsService("data.aws_prefix_list.s3_by_name", "name", "s3"),
					resource.TestMatchResourceAttr("data.aws_prefix_list.s3_by_id", "id", regexp.MustCompile(`^pl-[0-9a-z]{8}$`)),
					testAccCheckResourceAttrRegionalReverseDnsService("data.aws_prefix_list.s3_by_id", "name", "s3"),
				),
			},
		},
	})
}

func TestAccDataSourceAwsPrefixList_filter(t *testing.T) {
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceAwsPrefixListConfigFilter,
				Check: resource.ComposeTestCheckFunc(
					testAccDataSourceAwsPrefixListCheck("data.aws_prefix_list.s3_by_id"),
					testAccDataSourceAwsPrefixListCheck("data.aws_prefix_list.s3_by_name"),
					resource.TestMatchResourceAttr("data.aws_prefix_list.s3_by_name", "id", regexp.MustCompile(`^pl-[0-9a-z]{8}$`)),
					testAccCheckResourceAttrRegionalReverseDnsService("data.aws_prefix_list.s3_by_name", "name", "s3"),
					resource.TestMatchResourceAttr("data.aws_prefix_list.s3_by_id", "id", regexp.MustCompile(`^pl-[0-9a-z]{8}$`)),
					testAccCheckResourceAttrRegionalReverseDnsService("data.aws_prefix_list.s3_by_id", "name", "s3"),
				),
			},
		},
	})
}

func testAccDataSourceAwsPrefixListCheck(name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("root module has no resource called %s", name)
		}

		attr := rs.Primary.Attributes

		var (
			cidrBlockSize int
			err           error
		)

		if cidrBlockSize, err = strconv.Atoi(attr["cidr_blocks.#"]); err != nil {
			return err
		}
		if cidrBlockSize < 1 {
			return fmt.Errorf("cidr_blocks seem suspiciously low: %d", cidrBlockSize)
		}

		return nil
	}
}

const testAccDataSourceAwsPrefixListConfig = `
data "aws_region" "current" {}

data "aws_prefix_list" "s3_by_name" {
  name = "com.amazonaws.${data.aws_region.current.name}.s3"
}

data "aws_prefix_list" "s3_by_id" {
  prefix_list_id = data.aws_prefix_list.s3_by_name.id
}
`

const testAccDataSourceAwsPrefixListConfigFilter = `
data "aws_region" "current" {}

data "aws_prefix_list" "s3_by_name" {
  filter {
    name   = "prefix-list-name"
    values = ["com.amazonaws.${data.aws_region.current.name}.s3"]
  }
}

data "aws_prefix_list" "s3_by_id" {
  filter {
    name   = "prefix-list-id"
    values = [data.aws_prefix_list.s3_by_name.id]
  }
}
`
