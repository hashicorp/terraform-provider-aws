package ec2_test

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
)

func testAccManagedPrefixListGetIdByNameDataSource(name string, id *string, arn *string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).EC2Conn

		output, err := conn.DescribeManagedPrefixLists(&ec2.DescribeManagedPrefixListsInput{
			Filters: []*ec2.Filter{
				{
					Name:   aws.String("prefix-list-name"),
					Values: aws.StringSlice([]string{name}),
				},
			},
		})

		if err != nil {
			return err
		}

		*id = *output.PrefixLists[0].PrefixListId
		*arn = *output.PrefixLists[0].PrefixListArn
		return nil
	}
}

func TestAccVPCManagedPrefixListDataSource_basic(t *testing.T) {
	prefixListName := fmt.Sprintf("com.amazonaws.%s.s3", acctest.Region())
	prefixListId := ""
	prefixListArn := ""

	resourceByName := "data.aws_ec2_managed_prefix_list.s3_by_name"
	resourceById := "data.aws_ec2_managed_prefix_list.s3_by_id"
	prefixListResourceName := "data.aws_prefix_list.s3_by_id"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheckManagedPrefixList(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ec2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccVPCManagedPrefixListDataSourceConfig_basic,
				Check: resource.ComposeTestCheckFunc(
					testAccManagedPrefixListGetIdByNameDataSource(prefixListName, &prefixListId, &prefixListArn),

					resource.TestCheckResourceAttrPtr(resourceByName, "id", &prefixListId),
					resource.TestCheckResourceAttr(resourceByName, "name", prefixListName),
					resource.TestCheckResourceAttr(resourceByName, "owner_id", "AWS"),
					resource.TestCheckResourceAttr(resourceByName, "address_family", "IPv4"),
					resource.TestCheckResourceAttrPtr(resourceByName, "arn", &prefixListArn),
					resource.TestCheckResourceAttr(resourceByName, "max_entries", "0"),
					resource.TestCheckResourceAttr(resourceByName, "version", "0"),
					resource.TestCheckResourceAttr(resourceByName, "tags.%", "0"),

					resource.TestCheckResourceAttrPtr(resourceById, "id", &prefixListId),
					resource.TestCheckResourceAttr(resourceById, "name", prefixListName),

					resource.TestCheckResourceAttrPair(resourceByName, "id", prefixListResourceName, "id"),
					resource.TestCheckResourceAttrPair(resourceByName, "name", prefixListResourceName, "name"),
					resource.TestCheckResourceAttrPair(resourceByName, "entries.#", prefixListResourceName, "cidr_blocks.#"),
				),
			},
		},
	})
}

const testAccVPCManagedPrefixListDataSourceConfig_basic = `
data "aws_region" "current" {}

data "aws_ec2_managed_prefix_list" "s3_by_name" {
  name = "com.amazonaws.${data.aws_region.current.name}.s3"
}

data "aws_ec2_managed_prefix_list" "s3_by_id" {
  id = data.aws_ec2_managed_prefix_list.s3_by_name.id
}

data "aws_prefix_list" "s3_by_id" {
  prefix_list_id = data.aws_ec2_managed_prefix_list.s3_by_name.id
}
`

func TestAccVPCManagedPrefixListDataSource_filter(t *testing.T) {
	prefixListName := fmt.Sprintf("com.amazonaws.%s.s3", acctest.Region())
	prefixListId := ""
	prefixListArn := ""

	resourceByName := "data.aws_ec2_managed_prefix_list.s3_by_name"
	resourceById := "data.aws_ec2_managed_prefix_list.s3_by_id"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheckManagedPrefixList(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ec2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccVPCManagedPrefixListDataSourceConfig_filter,
				Check: resource.ComposeTestCheckFunc(
					testAccManagedPrefixListGetIdByNameDataSource(prefixListName, &prefixListId, &prefixListArn),
					resource.TestCheckResourceAttrPtr(resourceByName, "id", &prefixListId),
					resource.TestCheckResourceAttr(resourceByName, "name", prefixListName),
					resource.TestCheckResourceAttr(resourceByName, "owner_id", "AWS"),
					resource.TestCheckResourceAttr(resourceByName, "address_family", "IPv4"),
					resource.TestCheckResourceAttrPtr(resourceByName, "arn", &prefixListArn),
					resource.TestCheckResourceAttr(resourceByName, "max_entries", "0"),
					resource.TestCheckResourceAttr(resourceByName, "version", "0"),
					resource.TestCheckResourceAttr(resourceByName, "tags.%", "0"),

					resource.TestCheckResourceAttrPair(resourceByName, "id", resourceById, "id"),
					resource.TestCheckResourceAttrPair(resourceByName, "name", resourceById, "name"),
					resource.TestCheckResourceAttrPair(resourceByName, "entries", resourceById, "entries"),
					resource.TestCheckResourceAttrPair(resourceByName, "owner_id", resourceById, "owner_id"),
					resource.TestCheckResourceAttrPair(resourceByName, "address_family", resourceById, "address_family"),
					resource.TestCheckResourceAttrPair(resourceByName, "arn", resourceById, "arn"),
					resource.TestCheckResourceAttrPair(resourceByName, "max_entries", resourceById, "max_entries"),
					resource.TestCheckResourceAttrPair(resourceByName, "tags", resourceById, "tags"),
					resource.TestCheckResourceAttrPair(resourceByName, "version", resourceById, "version"),
				),
			},
		},
	})
}

const testAccVPCManagedPrefixListDataSourceConfig_filter = `
data "aws_region" "current" {}

data "aws_ec2_managed_prefix_list" "s3_by_name" {
  filter {
    name   = "prefix-list-name"
    values = ["com.amazonaws.${data.aws_region.current.name}.s3"]
  }
}

data "aws_ec2_managed_prefix_list" "s3_by_id" {
  filter {
    name   = "prefix-list-id"
    values = [data.aws_ec2_managed_prefix_list.s3_by_name.id]
  }
}
`

func TestAccVPCManagedPrefixListDataSource_matchesTooMany(t *testing.T) {
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheckManagedPrefixList(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ec2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config:      testAccVPCManagedPrefixListDataSourceConfig_matchesTooMany,
				ExpectError: regexp.MustCompile(`more than 1 prefix list matched the given criteria`),
			},
		},
	})
}

const testAccVPCManagedPrefixListDataSourceConfig_matchesTooMany = `
data "aws_ec2_managed_prefix_list" "test" {}
`
