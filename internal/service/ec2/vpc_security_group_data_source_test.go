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

func TestAccVPCSecurityGroupDataSource_basic(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccVPCSecurityGroupDataSourceConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccSecurityGroupCheckDataSource("data.aws_security_group.by_id"),
					testAccSecurityGroupCheckDataSource("data.aws_security_group.by_tag"),
					testAccSecurityGroupCheckDataSource("data.aws_security_group.by_filter"),
					testAccSecurityGroupCheckDataSource("data.aws_security_group.by_name"),
				),
			},
		},
	})
}

func testAccSecurityGroupCheckDataSource(dataSourceName string) resource.TestCheckFunc {
	resourceName := "aws_security_group.test"

	return resource.ComposeAggregateTestCheckFunc(
		resource.TestCheckResourceAttrPair(dataSourceName, names.AttrARN, resourceName, names.AttrARN),
		resource.TestCheckResourceAttrPair(dataSourceName, names.AttrDescription, resourceName, names.AttrDescription),
		resource.TestCheckResourceAttrPair(dataSourceName, names.AttrName, resourceName, names.AttrName),
		resource.TestCheckResourceAttrPair(dataSourceName, acctest.CtTagsPercent, resourceName, acctest.CtTagsPercent),
		resource.TestCheckResourceAttrPair(dataSourceName, names.AttrVPCID, resourceName, names.AttrVPCID),
	)
}

func testAccVPCSecurityGroupDataSourceConfig_basic(rName string) string {
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

data "aws_security_group" "by_id" {
  id = aws_security_group.test.id
}

data "aws_security_group" "by_name" {
  name = aws_security_group.test.name
}

data "aws_security_group" "by_tag" {
  tags = {
    Name = aws_security_group.test.tags["Name"]
  }
}

data "aws_security_group" "by_filter" {
  filter {
    name   = "group-name"
    values = [aws_security_group.test.name]
  }
}
`, rName)
}
