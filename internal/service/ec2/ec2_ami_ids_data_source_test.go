// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package ec2_test

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
)

func TestAccEC2AMIIDsDataSource_basic(t *testing.T) {
	ctx := acctest.Context(t)
	datasourceName := "data.aws_ami_ids.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, ec2.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccAMIIDsDataSourceConfig_basic,
				Check: resource.ComposeTestCheckFunc(
					acctest.CheckResourceAttrGreaterThanValue(datasourceName, "ids.#", 0),
				),
			},
		},
	})
}

func TestAccEC2AMIIDsDataSource_sorted(t *testing.T) {
	ctx := acctest.Context(t)
	datasourceName := "data.aws_ami_ids.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, ec2.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccAMIIDsDataSourceConfig_sorted(false),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(datasourceName, "ids.#", "2"),
					resource.TestCheckResourceAttrPair(datasourceName, "ids.0", "data.aws_ami.test2", "id"),
					resource.TestCheckResourceAttrPair(datasourceName, "ids.1", "data.aws_ami.test1", "id"),
				),
			},
			{
				Config: testAccAMIIDsDataSourceConfig_sorted(true),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(datasourceName, "ids.#", "2"),
					resource.TestCheckResourceAttrPair(datasourceName, "ids.0", "data.aws_ami.test1", "id"),
					resource.TestCheckResourceAttrPair(datasourceName, "ids.1", "data.aws_ami.test2", "id"),
				),
			},
		},
	})
}

func TestAccEC2AMIIDsDataSource_includeDeprecated(t *testing.T) {
	ctx := acctest.Context(t)
	datasourceName := "data.aws_ami_ids.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, ec2.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccAMIIDsDataSourceConfig_includeDeprecated(true),
				Check: resource.ComposeTestCheckFunc(
					acctest.CheckResourceAttrGreaterThanValue(datasourceName, "ids.#", 0),
				),
			},
		},
	})
}

const testAccAMIIDsDataSourceConfig_basic = `
data "aws_ami_ids" "test" {
  owners = ["099720109477"]

  filter {
    name   = "name"
    values = ["ubuntu/images/ubuntu-*-*-amd64-server-*"]
  }
}
`

func testAccAMIIDsDataSourceConfig_sorted(sortAscending bool) string {
	return fmt.Sprintf(`
data "aws_ami" "test1" {
  owners = ["amazon"]

  filter {
    name   = "name"
    values = ["amzn-ami-hvm-2018.03.0.20221018.0-x86_64-gp2"]
  }
}

data "aws_ami" "test2" {
  owners = ["amazon"]

  filter {
    name   = "name"
    values = ["amzn-ami-hvm-2018.03.0.20221209.1-x86_64-gp2"]
  }
}

data "aws_ami_ids" "test" {
  owners = ["amazon"]

  filter {
    name   = "name"
    values = [data.aws_ami.test1.name, data.aws_ami.test2.name]
  }

  sort_ascending = %[1]t
}
`, sortAscending)
}

func testAccAMIIDsDataSourceConfig_includeDeprecated(includeDeprecated bool) string {
	return fmt.Sprintf(`
data "aws_ami_ids" "test" {
  owners             = ["099720109477"]
  include_deprecated = %[1]t

  filter {
    name   = "name"
    values = ["ubuntu/images/ubuntu-*-*-amd64-server-*"]
  }
}
`, includeDeprecated)
}
