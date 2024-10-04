// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package ec2_test

import (
	"fmt"
	"testing"

	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccVPCManagedPrefixListsDataSource_basic(t *testing.T) {
	ctx := acctest.Context(t)
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccVPCManagedPrefixListsDataSourceConfig_basic,
				Check: resource.ComposeTestCheckFunc(
					acctest.CheckResourceAttrGreaterThanValue("data.aws_ec2_managed_prefix_lists.test", "ids.#", 0),
				),
			},
		},
	})
}

func TestAccVPCManagedPrefixListsDataSource_tags(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccVPCManagedPrefixListsDataSourceConfig_tags(rName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("data.aws_ec2_managed_prefix_lists.test", "ids.#", acctest.Ct1),
				),
			},
		},
	})
}

func TestAccVPCManagedPrefixListsDataSource_noMatches(t *testing.T) {
	ctx := acctest.Context(t)
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccVPCManagedPrefixListsDataSourceConfig_noMatches,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("data.aws_ec2_managed_prefix_lists.test", "ids.#", acctest.Ct0),
				),
			},
		},
	})
}

const testAccVPCManagedPrefixListsDataSourceConfig_basic = `
data "aws_ec2_managed_prefix_lists" "test" {}
`

func testAccVPCManagedPrefixListsDataSourceConfig_tags(rName string) string {
	return fmt.Sprintf(`
resource "aws_ec2_managed_prefix_list" "test" {
  address_family = "IPv4"
  max_entries    = 1
  name           = %[1]q

  tags = {
    Name = %[1]q
  }
}

data "aws_ec2_managed_prefix_lists" "test" {
  tags = {
    Name = aws_ec2_managed_prefix_list.test.tags["Name"]
  }
}
`, rName)
}

const testAccVPCManagedPrefixListsDataSourceConfig_noMatches = `
data "aws_ec2_managed_prefix_lists" "test" {
  filter {
    name   = "prefix-list-name"
    values = ["no-match"]
  }
}
`
