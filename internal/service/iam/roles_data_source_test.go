// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package iam_test

import (
	"fmt"
	"strconv"
	"testing"

	"github.com/YakDriver/regexache"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccIAMRolesDataSource_basic(t *testing.T) {
	ctx := acctest.Context(t)
	dataSourceName := "data.aws_iam_roles.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.IAMServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccRolesDataSourceConfig_basic,
				Check: resource.ComposeTestCheckFunc(
					resource.TestMatchResourceAttr(dataSourceName, "names.#", regexache.MustCompile("[^0].*$")),
				),
			},
		},
	})
}

func TestAccIAMRolesDataSource_nameRegex(t *testing.T) {
	ctx := acctest.Context(t)
	rCount := strconv.Itoa(sdkacctest.RandIntRange(1, 4))
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	dataSourceName := "data.aws_iam_roles.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.IAMServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccRolesDataSourceConfig_nameRegex(rCount, rName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(dataSourceName, "names.#", rCount),
					resource.TestCheckResourceAttr(dataSourceName, "arns.#", rCount),
				),
			},
		},
	})
}

func TestAccIAMRolesDataSource_pathPrefix(t *testing.T) {
	ctx := acctest.Context(t)
	rCount := strconv.Itoa(sdkacctest.RandIntRange(1, 4))
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	rPathPrefix := sdkacctest.RandomWithPrefix("tf-acc-path")
	dataSourceName := "data.aws_iam_roles.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.IAMServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccRolesDataSourceConfig_pathPrefix(rCount, rName, rPathPrefix),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(dataSourceName, "names.#", rCount),
					resource.TestCheckResourceAttr(dataSourceName, "arns.#", rCount),
				),
			},
		},
	})
}

func TestAccIAMRolesDataSource_nonExistentPathPrefix(t *testing.T) {
	ctx := acctest.Context(t)
	dataSourceName := "data.aws_iam_roles.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.IAMServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccRolesDataSourceConfig_nonExistentPathPrefix,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(dataSourceName, "arns.#", acctest.Ct0),
					resource.TestCheckResourceAttr(dataSourceName, "names.#", acctest.Ct0),
				),
			},
		},
	})
}

func TestAccIAMRolesDataSource_nameRegexAndPathPrefix(t *testing.T) {
	ctx := acctest.Context(t)
	rCount := strconv.Itoa(sdkacctest.RandIntRange(1, 4))
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	rPathPrefix := sdkacctest.RandomWithPrefix("tf-acc-path")
	dataSourceName := "data.aws_iam_roles.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.IAMServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccRolesDataSourceConfig_nameRegexAndPathPrefix(rCount, rName, rPathPrefix, acctest.Ct0),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(dataSourceName, "names.#", acctest.Ct1),
					resource.TestCheckResourceAttr(dataSourceName, "arns.#", acctest.Ct1),
				),
			},
		},
	})
}

const testAccRolesDataSourceConfig_basic = `
data "aws_iam_roles" "test" {}
`

func testAccRolesDataSourceConfig_nameRegex(rCount, rName string) string {
	return fmt.Sprintf(`
data "aws_partition" "current" {}

resource "aws_iam_role" "test" {
  count = %[1]s
  name  = "%[2]s-${count.index}-role"

  assume_role_policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Action": "sts:AssumeRole",
      "Principal": {
        "Service": "ec2.${data.aws_partition.current.dns_suffix}"
      },
      "Effect": "Allow",
      "Sid": ""
    }
  ]
}
EOF
  tags = {
    Seed = %[2]q
  }
}

data "aws_iam_roles" "test" {
  name_regex = "${aws_iam_role.test[0].tags["Seed"]}-.*-role"
}
`, rCount, rName)
}

func testAccRolesDataSourceConfig_pathPrefix(rCount, rName, rPathPrefix string) string {
	return fmt.Sprintf(`
data "aws_partition" "current" {}

resource "aws_iam_role" "test" {
  count = %[1]s
  name  = "%[2]s-${count.index}-role"

  assume_role_policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Action": "sts:AssumeRole",
      "Principal": {
        "Service": "ec2.${data.aws_partition.current.dns_suffix}"
      },
      "Effect": "Allow",
      "Sid": ""
    }
  ]
}
EOF

  path = "/%[3]s/"
}

data "aws_iam_roles" "test" {
  path_prefix = aws_iam_role.test[0].path
}
`, rCount, rName, rPathPrefix)
}

const testAccRolesDataSourceConfig_nonExistentPathPrefix = `
data "aws_iam_roles" "test" {
  path_prefix = "/dne/path"
}
`

func testAccRolesDataSourceConfig_nameRegexAndPathPrefix(rCount, rName, rPathPrefix, rIndex string) string {
	return fmt.Sprintf(`
data "aws_partition" "current" {}

resource "aws_iam_role" "test" {
  count = %[1]s
  name  = "%[2]s-${count.index}-role"

  assume_role_policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Action": "sts:AssumeRole",
      "Principal": {
        "Service": "ec2.${data.aws_partition.current.dns_suffix}"
      },
      "Effect": "Allow",
      "Sid": ""
    }
  ]
}
EOF

  path = "/%[3]s/"
  tags = {
    Seed = %[2]q
  }
}

data "aws_iam_roles" "test" {
  name_regex  = "${aws_iam_role.test[0].tags["Seed"]}-%[4]s-role"
  path_prefix = aws_iam_role.test[0].path
}
`, rCount, rName, rPathPrefix, rIndex)
}
