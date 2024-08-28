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

func TestAccVPCSecurityGroupsDataSource_tag(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	dataSourceName := "data.aws_security_groups.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccVPCSecurityGroupsDataSourceConfig_tag(rName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(dataSourceName, "arns.#", acctest.Ct3),
					resource.TestCheckResourceAttr(dataSourceName, "ids.#", acctest.Ct3),
					resource.TestCheckResourceAttr(dataSourceName, "vpc_ids.#", acctest.Ct3),
				),
			},
		},
	})
}

func TestAccVPCSecurityGroupsDataSource_filter(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	dataSourceName := "data.aws_security_groups.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccVPCSecurityGroupsDataSourceConfig_filter(rName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(dataSourceName, "arns.#", acctest.Ct1),
					resource.TestCheckResourceAttr(dataSourceName, "ids.#", acctest.Ct1),
					resource.TestCheckResourceAttr(dataSourceName, "vpc_ids.#", acctest.Ct1),
				),
			},
		},
	})
}

func TestAccVPCSecurityGroupsDataSource_empty(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	dataSourceName := "data.aws_security_groups.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccVPCSecurityGroupsDataSourceConfig_empty(rName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(dataSourceName, "arns.#", acctest.Ct0),
					resource.TestCheckResourceAttr(dataSourceName, "ids.#", acctest.Ct0),
					resource.TestCheckResourceAttr(dataSourceName, "vpc_ids.#", acctest.Ct0),
				),
			},
		},
	})
}

func testAccVPCSecurityGroupsDataSourceConfig_tag(rName string) string {
	return fmt.Sprintf(`
resource "aws_vpc" "test" {
  cidr_block = "172.16.0.0/16"

  tags = {
    Name = %[1]q
  }
}

resource "aws_security_group" "test" {
  count  = 3
  vpc_id = aws_vpc.test.id
  name   = "%[1]s-${count.index}"

  tags = {
    Name = %[1]q
  }
}

data "aws_security_groups" "test" {
  tags = {
    Name = %[1]q
  }

  depends_on = [aws_security_group.test[0], aws_security_group.test[1], aws_security_group.test[2]]
}
`, rName)
}

func testAccVPCSecurityGroupsDataSourceConfig_filter(rName string) string {
	return fmt.Sprintf(`
resource "aws_vpc" "test" {
  cidr_block = "172.16.0.0/16"

  tags = {
    Name = %[1]q
  }
}

resource "aws_security_group" "test" {
  vpc_id = aws_vpc.test.id
  name   = %[1]q

  tags = {
    Name = %[1]q
  }
}

data "aws_security_groups" "test" {
  filter {
    name   = "vpc-id"
    values = [aws_vpc.test.id]
  }

  filter {
    name   = "group-name"
    values = [aws_security_group.test.name]
  }
}
`, rName)
}

func testAccVPCSecurityGroupsDataSourceConfig_empty(rName string) string {
	return fmt.Sprintf(`
data "aws_security_groups" "test" {
  tags = {
    Name = %[1]q
  }
}
`, rName)
}
