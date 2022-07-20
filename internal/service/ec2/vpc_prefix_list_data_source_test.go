package ec2_test

import (
	"regexp"
	"testing"

	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
)

func TestAccVPCPrefixListDataSource_basic(t *testing.T) {
	ds1Name := "data.aws_prefix_list.s3_by_id"
	ds2Name := "data.aws_prefix_list.s3_by_name"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ec2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccVPCPrefixListDataSourceConfig_basic,
				Check: resource.ComposeTestCheckFunc(
					acctest.CheckResourceAttrGreaterThanValue(ds1Name, "cidr_blocks.#", "0"),
					resource.TestCheckResourceAttrSet(ds1Name, "name"),
					acctest.CheckResourceAttrGreaterThanValue(ds2Name, "cidr_blocks.#", "0"),
					resource.TestCheckResourceAttrSet(ds2Name, "name"),
				),
			},
		},
	})
}

func TestAccVPCPrefixListDataSource_filter(t *testing.T) {
	ds1Name := "data.aws_prefix_list.s3_by_id"
	ds2Name := "data.aws_prefix_list.s3_by_name"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ec2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccVPCPrefixListDataSourceConfig_filter,
				Check: resource.ComposeTestCheckFunc(
					acctest.CheckResourceAttrGreaterThanValue(ds1Name, "cidr_blocks.#", "0"),
					resource.TestCheckResourceAttrSet(ds1Name, "name"),
					acctest.CheckResourceAttrGreaterThanValue(ds2Name, "cidr_blocks.#", "0"),
					resource.TestCheckResourceAttrSet(ds2Name, "name"),
				),
			},
		},
	})
}

func TestAccVPCPrefixListDataSource_nameDoesNotOverrideFilter(t *testing.T) {
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ec2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config:      testAccVPCPrefixListDataSourceConfig_nameDoesNotOverrideFilter,
				ExpectError: regexp.MustCompile(`no matching EC2 Prefix List found`),
			},
		},
	})
}

const testAccVPCPrefixListDataSourceConfig_basic = `
data "aws_region" "current" {}

data "aws_prefix_list" "s3_by_id" {
  prefix_list_id = data.aws_prefix_list.s3_by_name.id
}

data "aws_prefix_list" "s3_by_name" {
  name = "com.amazonaws.${data.aws_region.current.name}.s3"
}
`

const testAccVPCPrefixListDataSourceConfig_filter = `
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

const testAccVPCPrefixListDataSourceConfig_nameDoesNotOverrideFilter = `
data "aws_region" "current" {}

data "aws_prefix_list" "test" {
  name = "com.amazonaws.${data.aws_region.current.name}.dynamodb"

  filter {
    name   = "prefix-list-name"
    values = ["com.amazonaws.${data.aws_region.current.name}.s3"]
  }
}
`
