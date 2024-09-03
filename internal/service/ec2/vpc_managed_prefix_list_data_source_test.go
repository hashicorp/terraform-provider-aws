// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package ec2_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	awstypes "github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func testAccManagedPrefixListGetIdByNameDataSource(ctx context.Context, name string, id *string, arn *string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).EC2Client(ctx)

		output, err := conn.DescribeManagedPrefixLists(ctx, &ec2.DescribeManagedPrefixListsInput{
			Filters: []awstypes.Filter{
				{
					Name:   aws.String("prefix-list-name"),
					Values: []string{name},
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
	ctx := acctest.Context(t)
	prefixListName := fmt.Sprintf("com.amazonaws.%s.s3", acctest.Region())
	prefixListId := ""
	prefixListArn := ""

	resourceByName := "data.aws_ec2_managed_prefix_list.s3_by_name"
	resourceById := "data.aws_ec2_managed_prefix_list.s3_by_id"
	prefixListResourceName := "data.aws_prefix_list.s3_by_id"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheckManagedPrefixList(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccVPCManagedPrefixListDataSourceConfig_basic,
				Check: resource.ComposeTestCheckFunc(
					testAccManagedPrefixListGetIdByNameDataSource(ctx, prefixListName, &prefixListId, &prefixListArn),

					resource.TestCheckResourceAttrPtr(resourceByName, names.AttrID, &prefixListId),
					resource.TestCheckResourceAttr(resourceByName, names.AttrName, prefixListName),
					resource.TestCheckResourceAttr(resourceByName, names.AttrOwnerID, "AWS"),
					resource.TestCheckResourceAttr(resourceByName, "address_family", "IPv4"),
					resource.TestCheckResourceAttrPtr(resourceByName, names.AttrARN, &prefixListArn),
					resource.TestCheckResourceAttr(resourceByName, "max_entries", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceByName, names.AttrVersion, acctest.Ct0),
					resource.TestCheckResourceAttr(resourceByName, acctest.CtTagsPercent, acctest.Ct0),

					resource.TestCheckResourceAttrPtr(resourceById, names.AttrID, &prefixListId),
					resource.TestCheckResourceAttr(resourceById, names.AttrName, prefixListName),

					resource.TestCheckResourceAttrPair(resourceByName, names.AttrID, prefixListResourceName, names.AttrID),
					resource.TestCheckResourceAttrPair(resourceByName, names.AttrName, prefixListResourceName, names.AttrName),
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
	ctx := acctest.Context(t)
	prefixListName := fmt.Sprintf("com.amazonaws.%s.s3", acctest.Region())
	prefixListId := ""
	prefixListArn := ""

	resourceByName := "data.aws_ec2_managed_prefix_list.s3_by_name"
	resourceById := "data.aws_ec2_managed_prefix_list.s3_by_id"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheckManagedPrefixList(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccVPCManagedPrefixListDataSourceConfig_filter,
				Check: resource.ComposeTestCheckFunc(
					testAccManagedPrefixListGetIdByNameDataSource(ctx, prefixListName, &prefixListId, &prefixListArn),
					resource.TestCheckResourceAttrPtr(resourceByName, names.AttrID, &prefixListId),
					resource.TestCheckResourceAttr(resourceByName, names.AttrName, prefixListName),
					resource.TestCheckResourceAttr(resourceByName, names.AttrOwnerID, "AWS"),
					resource.TestCheckResourceAttr(resourceByName, "address_family", "IPv4"),
					resource.TestCheckResourceAttrPtr(resourceByName, names.AttrARN, &prefixListArn),
					resource.TestCheckResourceAttr(resourceByName, "max_entries", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceByName, names.AttrVersion, acctest.Ct0),
					resource.TestCheckResourceAttr(resourceByName, acctest.CtTagsPercent, acctest.Ct0),

					resource.TestCheckResourceAttrPair(resourceByName, names.AttrID, resourceById, names.AttrID),
					resource.TestCheckResourceAttrPair(resourceByName, names.AttrName, resourceById, names.AttrName),
					resource.TestCheckResourceAttrPair(resourceByName, "entries", resourceById, "entries"),
					resource.TestCheckResourceAttrPair(resourceByName, names.AttrOwnerID, resourceById, names.AttrOwnerID),
					resource.TestCheckResourceAttrPair(resourceByName, "address_family", resourceById, "address_family"),
					resource.TestCheckResourceAttrPair(resourceByName, names.AttrARN, resourceById, names.AttrARN),
					resource.TestCheckResourceAttrPair(resourceByName, "max_entries", resourceById, "max_entries"),
					resource.TestCheckResourceAttrPair(resourceByName, names.AttrTags, resourceById, names.AttrTags),
					resource.TestCheckResourceAttrPair(resourceByName, names.AttrVersion, resourceById, names.AttrVersion),
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
	ctx := acctest.Context(t)
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheckManagedPrefixList(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config:      testAccVPCManagedPrefixListDataSourceConfig_matchesTooMany,
				ExpectError: regexache.MustCompile(`multiple EC2 Managed Prefix Lists matched`),
			},
		},
	})
}

const testAccVPCManagedPrefixListDataSourceConfig_matchesTooMany = `
data "aws_ec2_managed_prefix_list" "test" {}
`
