// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package ssm_test

import (
	"fmt"
	"testing"

	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccSSMParametersByPathDataSource_basic(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "data.aws_ssm_parameters_by_path.test"
	rName1 := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	rName2 := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.SSMServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccParametersByPathDataSourceConfig_basic(rName1, rName2, false),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "arns.#", acctest.Ct2),
					resource.TestCheckResourceAttr(resourceName, "names.#", acctest.Ct2),
					resource.TestCheckResourceAttr(resourceName, "types.#", acctest.Ct2),
					resource.TestCheckResourceAttr(resourceName, "values.#", acctest.Ct2),
					resource.TestCheckResourceAttr(resourceName, "with_decryption", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "recursive", acctest.CtFalse),
				),
			},
		},
	})
}

func testAccParametersByPathDataSourceConfig_basic(rName1, rName2 string, withDecryption bool) string {
	return fmt.Sprintf(`
resource "aws_ssm_parameter" "test1" {
  name  = "/%[1]s/param-a"
  type  = "String"
  value = "TestValueA"
}

resource "aws_ssm_parameter" "test2" {
  name  = "/%[1]s/param-b"
  type  = "String"
  value = "TestValueB"
}

resource "aws_ssm_parameter" "test3" {
  name  = "/%[2]s/param-c"
  type  = "String"
  value = "TestValueC"
}

data "aws_ssm_parameters_by_path" "test" {
  path            = "/%[1]s"
  with_decryption = %[3]t

  depends_on = [
    aws_ssm_parameter.test1,
    aws_ssm_parameter.test2,
    aws_ssm_parameter.test3,
  ]
}
`, rName1, rName2, withDecryption)
}

func TestAccSSMParametersByPathDataSource_withRecursion(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "data.aws_ssm_parameters_by_path.recursive"
	pathPrefix := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.SSMServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccParametersByPathDataSourceConfig_recursion(pathPrefix),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "arns.#", acctest.Ct2),
					resource.TestCheckResourceAttr(resourceName, "names.#", acctest.Ct2),
					resource.TestCheckResourceAttr(resourceName, "types.#", acctest.Ct2),
					resource.TestCheckResourceAttr(resourceName, "values.#", acctest.Ct2),
					resource.TestCheckResourceAttr(resourceName, "recursive", acctest.CtTrue),
				),
			},
		},
	})
}

func testAccParametersByPathDataSourceConfig_recursion(pathPrefix string) string {
	return fmt.Sprintf(`
resource "aws_ssm_parameter" "top_level" {
  name  = "/%[1]s/top_param"
  type  = "String"
  value = "TestValueA"
}

resource "aws_ssm_parameter" "nested" {
  name  = "/%[1]s/nested/param"
  type  = "String"
  value = "TestValueB"
}

data "aws_ssm_parameters_by_path" "recursive" {
  path      = "/%[1]s/"
  recursive = true

  depends_on = [
    aws_ssm_parameter.top_level,
    aws_ssm_parameter.nested,
  ]
}
`, pathPrefix)
}
