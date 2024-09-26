// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package dax_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/dax"
	awstypes "github.com/aws/aws-sdk-go-v2/service/dax/types"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	tfdax "github.com/hashicorp/terraform-provider-aws/internal/service/dax"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccDAXSubnetGroup_basic(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_dax_subnet_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.DAXServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckSubnetGroupDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccSubnetGroupConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSubnetGroupExists(ctx, "aws_dax_subnet_group.test"),
					resource.TestCheckResourceAttr("aws_dax_subnet_group.test", "subnet_ids.#", acctest.Ct2),
					resource.TestCheckResourceAttrSet("aws_dax_subnet_group.test", names.AttrVPCID),
				),
			},
			{
				Config: testAccSubnetGroupConfig_update(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSubnetGroupExists(ctx, "aws_dax_subnet_group.test"),
					resource.TestCheckResourceAttr("aws_dax_subnet_group.test", names.AttrDescription, "update"),
					resource.TestCheckResourceAttr("aws_dax_subnet_group.test", "subnet_ids.#", acctest.Ct3),
					resource.TestCheckResourceAttrSet("aws_dax_subnet_group.test", names.AttrVPCID),
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

func TestAccDAXSubnetGroup_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_dax_subnet_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.DAXServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckSubnetGroupDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccSubnetGroupConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSubnetGroupExists(ctx, resourceName),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfdax.ResourceSubnetGroup(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckSubnetGroupDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).DAXClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_dax_subnet_group" {
				continue
			}

			_, err := conn.DescribeSubnetGroups(ctx, &dax.DescribeSubnetGroupsInput{
				SubnetGroupNames: []string{rs.Primary.ID},
			})
			if err != nil {
				if errs.IsA[*awstypes.SubnetGroupNotFoundFault](err) {
					return nil
				}
				return err
			}
		}
		return nil
	}
}

func testAccCheckSubnetGroupExists(ctx context.Context, name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("Not found: %s", name)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).DAXClient(ctx)

		_, err := conn.DescribeSubnetGroups(ctx, &dax.DescribeSubnetGroupsInput{
			SubnetGroupNames: []string{rs.Primary.ID},
		})

		return err
	}
}

func testAccSubnetGroupConfig_basic(rName string) string {
	return fmt.Sprintf(`
resource "aws_vpc" "test" {
  cidr_block = "10.0.0.0/16"
}

resource "aws_subnet" "test1" {
  cidr_block = "10.0.1.0/24"
  vpc_id     = aws_vpc.test.id
}

resource "aws_subnet" "test2" {
  cidr_block = "10.0.2.0/24"
  vpc_id     = aws_vpc.test.id
}

resource "aws_dax_subnet_group" "test" {
  name = "%s"

  subnet_ids = [
    aws_subnet.test1.id,
    aws_subnet.test2.id,
  ]
}
`, rName)
}

func testAccSubnetGroupConfig_update(rName string) string {
	return fmt.Sprintf(`
resource "aws_vpc" "test" {
  cidr_block = "10.0.0.0/16"
}

resource "aws_subnet" "test1" {
  cidr_block = "10.0.1.0/24"
  vpc_id     = aws_vpc.test.id
}

resource "aws_subnet" "test2" {
  cidr_block = "10.0.2.0/24"
  vpc_id     = aws_vpc.test.id
}

resource "aws_subnet" "test3" {
  cidr_block = "10.0.3.0/24"
  vpc_id     = aws_vpc.test.id
}

resource "aws_dax_subnet_group" "test" {
  name        = "%s"
  description = "update"

  subnet_ids = [
    aws_subnet.test1.id,
    aws_subnet.test2.id,
    aws_subnet.test3.id,
  ]
}
`, rName)
}
