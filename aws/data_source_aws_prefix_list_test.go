package aws

import (
	"fmt"
	"regexp"
	"strconv"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
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

		if attr["name"] != "com.amazonaws.us-west-2.s3" {
			return fmt.Errorf("bad name %s", attr["name"])
		}
		if attr["id"] != "pl-68a54001" {
			return fmt.Errorf("bad id %s", attr["id"])
		}

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

		if actual := attr["owner_id"]; actual != "AWS" {
			return fmt.Errorf("bad owner_id %s", actual)
		}

		if actual := attr["address_family"]; actual != "IPv4" {
			return fmt.Errorf("bad address_family %s", actual)
		}

		if actual := attr["arn"]; actual != "arn:aws:ec2:us-west-2:aws:prefix-list/pl-68a54001" {
			return fmt.Errorf("bad arn %s", actual)
		}

		if actual := attr["max_entries"]; actual != "" {
			return fmt.Errorf("unexpected max_entries %s", actual)
		}

		if attr["tags.%"] != "0" {
			return fmt.Errorf("expected 0 tags")
		}

		return nil
	}
}

const testAccDataSourceAwsPrefixListConfig = `
data "aws_prefix_list" "s3_by_id" {
  prefix_list_id = "pl-68a54001"
}

data "aws_prefix_list" "s3_by_name" {
 name = "com.amazonaws.us-west-2.s3"
}
`

const testAccDataSourceAwsPrefixListConfigFilter = `
data "aws_prefix_list" "s3_by_name" {
  filter {
    name   = "prefix-list-name"
    values = ["com.amazonaws.us-west-2.s3"]
  }
}

data "aws_prefix_list" "s3_by_id" {
  filter {
    name   = "prefix-list-id"
    values = ["pl-68a54001"]
  }
}
`

func TestAccDataSourceAwsPrefixList_matchesTooMany(t *testing.T) {
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config:      testAccDataSourceAwsPrefixListConfig_matchesTooMany,
				ExpectError: regexp.MustCompile(`more than one prefix list matched the given set of criteria`),
			},
		},
	})
}

const testAccDataSourceAwsPrefixListConfig_matchesTooMany = `
data "aws_prefix_list" "test" {}
`

func TestAccDataSourceAwsPrefixList_nameDoesNotOverrideFilter(t *testing.T) {
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				// The vanilla DescribePrefixLists API only supports filtering by
				// id and name. In this case, the `name` attribute and `prefix-list-id`
				// filter have been set up such that they conflict, thus proving
				// that both criteria took effect.
				Config:      testAccDataSourceAwsPrefixListConfig_nameDoesNotOverrideFilter,
				ExpectError: regexp.MustCompile(`no matching prefix list found`),
			},
		},
	})
}

const testAccDataSourceAwsPrefixListConfig_nameDoesNotOverrideFilter = `
data "aws_prefix_list" "test" {
  name = "com.amazonaws.us-west-2.s3" 
  filter {
    name = "prefix-list-id"
    values = ["pl-00a54069"]  # com.amazonaws.us-west-2.dynamodb
  }
}
`

func TestAccDataSourceAwsPrefixList_managedPrefixList(t *testing.T) {
	resourceName := "aws_prefix_list.test"
	dataSourceName := "data.aws_prefix_list.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSPrefixListDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceAwsPrefixListConfig_managedPrefixList,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(resourceName, "id", dataSourceName, "id"),
					resource.TestCheckResourceAttrPair(resourceName, "name", dataSourceName, "name"),
					resource.TestCheckResourceAttrPair(resourceName, "arn", dataSourceName, "arn"),
					resource.TestCheckResourceAttrPair(resourceName, "owner_id", dataSourceName, "owner_id"),
					testAccCheckResourceAttrAccountID(dataSourceName, "owner_id"),
					resource.TestCheckResourceAttrPair(resourceName, "name", dataSourceName, "name"),
					resource.TestCheckResourceAttrPair(resourceName, "address_family", dataSourceName, "address_family"),
					resource.TestCheckResourceAttrPair(resourceName, "max_entries", dataSourceName, "max_entries"),
					resource.TestCheckResourceAttr(dataSourceName, "cidr_blocks.#", "2"),
					resource.TestCheckResourceAttr(dataSourceName, "cidr_blocks.0", "1.0.0.0/8"),
					resource.TestCheckResourceAttr(dataSourceName, "cidr_blocks.1", "2.0.0.0/8"),
					resource.TestCheckResourceAttr(dataSourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(dataSourceName, "tags.Key1", "Value1"),
					resource.TestCheckResourceAttr(dataSourceName, "tags.Key2", "Value2"),
				),
			},
		},
	})
}

const testAccDataSourceAwsPrefixListConfig_managedPrefixList = `
resource "aws_prefix_list" "test" {
  name           = "tf-test-acc"
  max_entries    = 5
  address_family = "IPv4"
  entry {
    cidr_block = "1.0.0.0/8"
  }
  entry {
    cidr_block = "2.0.0.0/8"
  }
  tags = {
    Key1 = "Value1"
    Key2 = "Value2"
  }
}

data "aws_prefix_list" "test" {
  prefix_list_id = aws_prefix_list.test.id
}
`
