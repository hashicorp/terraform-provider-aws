// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package iam_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	tfknownvalue "github.com/hashicorp/terraform-provider-aws/internal/acctest/knownvalue"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccIAMRolesDataSource_basic(t *testing.T) {
	ctx := acctest.Context(t)
	dataSourceName := "data.aws_iam_roles.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.IAMServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccRolesDataSourceConfig_basic,
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(dataSourceName, tfjsonpath.New(names.AttrARNs), tfknownvalue.ListNotEmpty()),
					statecheck.ExpectKnownValue(dataSourceName, tfjsonpath.New("name_regex"), knownvalue.Null()),
					statecheck.ExpectKnownValue(dataSourceName, tfjsonpath.New(names.AttrNames), tfknownvalue.ListNotEmpty()),
					statecheck.ExpectKnownValue(dataSourceName, tfjsonpath.New("path_prefix"), knownvalue.Null()),
				},
			},
		},
	})
}

func TestAccIAMRolesDataSource_nameRegex(t *testing.T) {
	ctx := acctest.Context(t)
	rCount := acctest.RandIntRange(t, 1, 4)
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	dataSourceName := "data.aws_iam_roles.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.IAMServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccRolesDataSourceConfig_nameRegex(rCount, rName),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(dataSourceName, tfjsonpath.New(names.AttrARNs), knownvalue.ListSizeExact(rCount)),
					statecheck.ExpectKnownValue(dataSourceName, tfjsonpath.New("name_regex"), knownvalue.StringExact(fmt.Sprintf("%s-.*-role", rName))),
					statecheck.ExpectKnownValue(dataSourceName, tfjsonpath.New(names.AttrNames), knownvalue.ListSizeExact(rCount)),
					statecheck.ExpectKnownValue(dataSourceName, tfjsonpath.New("path_prefix"), knownvalue.Null()),
				},
			},
		},
	})
}

func TestAccIAMRolesDataSource_pathPrefix(t *testing.T) {
	ctx := acctest.Context(t)
	rCount := acctest.RandIntRange(t, 1, 4)
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	rPathPrefix := acctest.RandomWithPrefix(t, "tf-acc-path")
	dataSourceName := "data.aws_iam_roles.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.IAMServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccRolesDataSourceConfig_pathPrefix(rCount, rName, rPathPrefix),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(dataSourceName, tfjsonpath.New(names.AttrARNs), knownvalue.ListSizeExact(rCount)),
					statecheck.ExpectKnownValue(dataSourceName, tfjsonpath.New("name_regex"), knownvalue.Null()),
					statecheck.ExpectKnownValue(dataSourceName, tfjsonpath.New(names.AttrNames), knownvalue.ListSizeExact(rCount)),
					statecheck.ExpectKnownValue(dataSourceName, tfjsonpath.New("path_prefix"), knownvalue.StringExact(fmt.Sprintf("/%s/", rPathPrefix))),
				},
			},
		},
	})
}

func TestAccIAMRolesDataSource_nonExistentPathPrefix(t *testing.T) {
	ctx := acctest.Context(t)
	dataSourceName := "data.aws_iam_roles.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.IAMServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccRolesDataSourceConfig_nonExistentPathPrefix,
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(dataSourceName, tfjsonpath.New(names.AttrARNs), knownvalue.ListExact([]knownvalue.Check{})),
					statecheck.ExpectKnownValue(dataSourceName, tfjsonpath.New(names.AttrNames), knownvalue.ListExact([]knownvalue.Check{})),
				},
			},
		},
	})
}

func TestAccIAMRolesDataSource_nameRegexAndPathPrefix(t *testing.T) {
	ctx := acctest.Context(t)
	rCount := acctest.RandIntRange(t, 1, 4)
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	rPathPrefix := acctest.RandomWithPrefix(t, "tf-acc-path")
	dataSourceName := "data.aws_iam_roles.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.IAMServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccRolesDataSourceConfig_nameRegexAndPathPrefix(rCount, rName, rPathPrefix, "0"),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(dataSourceName, tfjsonpath.New(names.AttrARNs), knownvalue.ListSizeExact(1)),
					statecheck.ExpectKnownValue(dataSourceName, tfjsonpath.New("name_regex"), knownvalue.StringExact(fmt.Sprintf("%s-0-role", rName))),
					statecheck.ExpectKnownValue(dataSourceName, tfjsonpath.New(names.AttrNames), knownvalue.ListSizeExact(1)),
					statecheck.ExpectKnownValue(dataSourceName, tfjsonpath.New("path_prefix"), knownvalue.StringExact(fmt.Sprintf("/%s/", rPathPrefix))),
				},
			},
		},
	})
}

const testAccRolesDataSourceConfig_basic = `
data "aws_iam_roles" "test" {}
`

func testAccRolesDataSourceConfig_nameRegex(rCount int, rName string) string {
	return fmt.Sprintf(`
data "aws_service_principal" "ec2" {
  service_name = "ec2"
}

resource "aws_iam_role" "test" {
  count = %[1]d
  name  = "%[2]s-${count.index}-role"

  assume_role_policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Action": "sts:AssumeRole",
      "Principal": {
        "Service": "${data.aws_service_principal.ec2.name}"
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

func testAccRolesDataSourceConfig_pathPrefix(rCount int, rName, rPathPrefix string) string {
	return fmt.Sprintf(`
data "aws_service_principal" "ec2" {
  service_name = "ec2"
}

resource "aws_iam_role" "test" {
  count = %[1]d
  name  = "%[2]s-${count.index}-role"

  assume_role_policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Action": "sts:AssumeRole",
      "Principal": {
        "Service": "${data.aws_service_principal.ec2.name}"
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

func testAccRolesDataSourceConfig_nameRegexAndPathPrefix(rCount int, rName, rPathPrefix, rIndex string) string {
	return fmt.Sprintf(`
data "aws_service_principal" "ec2" {
  service_name = "ec2"
}

resource "aws_iam_role" "test" {
  count = %[1]d
  name  = "%[2]s-${count.index}-role"

  assume_role_policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Action": "sts:AssumeRole",
      "Principal": {
        "Service": "${data.aws_service_principal.ec2.name}"
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
