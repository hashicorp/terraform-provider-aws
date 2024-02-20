// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package redshiftserverless_test

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/service/redshiftserverless"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
)

func TestAccRedshiftServerlessNamespaceDataSource_basic(t *testing.T) {
	ctx := acctest.Context(t)
	dataSourceName := "data.aws_redshiftserverless_namespace.test"
	resourceName := "aws_redshiftserverless_namespace.test"

	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, redshiftserverless.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccNamespaceDataSourceConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrPair(dataSourceName, "admin_username", resourceName, "admin_username"),
					resource.TestCheckResourceAttrPair(dataSourceName, "arn", resourceName, "arn"),
					resource.TestCheckResourceAttrPair(dataSourceName, "db_name", resourceName, "db_name"),
					resource.TestCheckResourceAttrPair(dataSourceName, "default_iam_role_arn", resourceName, "default_iam_role_arn"),
					resource.TestCheckResourceAttr(dataSourceName, "iam_roles.#", "0"),
					resource.TestCheckResourceAttrPair(dataSourceName, "kms_key_id", resourceName, "kms_key_id"),
					resource.TestCheckResourceAttr(dataSourceName, "log_exports.#", "0"),
					resource.TestCheckResourceAttrPair(dataSourceName, "namespace_id", resourceName, "namespace_id"),
					resource.TestCheckResourceAttrPair(dataSourceName, "namespace_name", resourceName, "namespace_name"),
				),
			},
		},
	})
}

func TestAccRedshiftServerlessNamespaceDataSource_iamRole(t *testing.T) {
	ctx := acctest.Context(t)
	dataSourceName := "data.aws_redshiftserverless_namespace.test"
	resourceName := "aws_redshiftserverless_namespace.test"

	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, redshiftserverless.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccNamespaceDataSourceConfig_defaultIAMRole(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrPair(dataSourceName, "namespace_name", resourceName, "namespace_name"),
					resource.TestCheckResourceAttrPair(dataSourceName, "default_iam_role_arn", resourceName, "default_iam_role_arn"),
					resource.TestCheckResourceAttr(dataSourceName, "iam_roles.#", "1"),
					resource.TestCheckTypeSetElemAttrPair(dataSourceName, "iam_roles.*", "aws_iam_role.test", "arn"),
				),
			},
		},
	})
}

func TestAccRedshiftServerlessNamespaceDataSource_user(t *testing.T) {
	ctx := acctest.Context(t)
	dataSourceName := "data.aws_redshiftserverless_namespace.test"
	resourceName := "aws_redshiftserverless_namespace.test"

	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	username := "admin_user"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, redshiftserverless.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccNamespaceDataSourceConfig_user(rName, username),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrPair(dataSourceName, "namespace_name", resourceName, "namespace_name"),
					resource.TestCheckResourceAttr(dataSourceName, "admin_username", username),
				),
			},
		},
	})
}

func TestAccRedshiftServerlessNamespaceDataSource_logExports(t *testing.T) {
	ctx := acctest.Context(t)
	dataSourceName := "data.aws_redshiftserverless_namespace.test"
	resourceName := "aws_redshiftserverless_namespace.test"

	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	logExport := "userlog"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, redshiftserverless.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccNamespaceDataSourceConfig_logExports(rName, logExport),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrPair(dataSourceName, "namespace_name", resourceName, "namespace_name"),
					resource.TestCheckResourceAttr(dataSourceName, "log_exports.#", "1"),
					resource.TestCheckResourceAttr(dataSourceName, "log_exports.0", logExport),
				),
			},
		},
	})
}

func testAccNamespaceDataSourceConfig_basic(rName string) string {
	return fmt.Sprintf(`
resource "aws_redshiftserverless_namespace" "test" {
  namespace_name = %[1]q
}

data "aws_redshiftserverless_namespace" "test" {
  namespace_name = aws_redshiftserverless_namespace.test.namespace_name
}
`, rName)
}

func testAccNamespaceDataSourceConfig_defaultIAMRole(rName string) string {
	return fmt.Sprintf(`
resource "aws_iam_role" "test" {
  name = %[1]q

  assume_role_policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Principal": {
        "Service": [
          "ec2.amazonaws.com"
        ]
      },
      "Action": [
        "sts:AssumeRole"
      ]
    }
  ]
}
EOF
}

resource "aws_redshiftserverless_namespace" "test" {
  namespace_name       = %[1]q
  default_iam_role_arn = aws_iam_role.test.arn
  iam_roles            = [aws_iam_role.test.arn]
}

data "aws_redshiftserverless_namespace" "test" {
  namespace_name = aws_redshiftserverless_namespace.test.namespace_name
}
`, rName)
}

func testAccNamespaceDataSourceConfig_user(rName string, username string) string {
	return fmt.Sprintf(`
resource "aws_redshiftserverless_namespace" "test" {
  namespace_name      = %[1]q
  admin_username      = %[2]q
  admin_user_password = "Test_Password_123"
}

data "aws_redshiftserverless_namespace" "test" {
  namespace_name = aws_redshiftserverless_namespace.test.namespace_name
}
`, rName, username)
}

func testAccNamespaceDataSourceConfig_logExports(rName string, logExport string) string {
	return fmt.Sprintf(`
resource "aws_redshiftserverless_namespace" "test" {
  namespace_name = %[1]q
  log_exports    = [%[2]q]
}

data "aws_redshiftserverless_namespace" "test" {
  namespace_name = aws_redshiftserverless_namespace.test.namespace_name
}
`, rName, logExport)
}
