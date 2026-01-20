// Copyright IBM Corp. 2014, 2026
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
	dataSourceName := "data.aws_ssm_parameters_by_path.test"
	resourceName1 := "aws_ssm_parameter.test1"
	resourceName2 := "aws_ssm_parameter.test2"
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
					resource.TestCheckResourceAttr(dataSourceName, fmt.Sprintf("%s.#", names.AttrParameters), "2"),
					resource.TestCheckTypeSetElemAttrPair(dataSourceName, fmt.Sprintf("%s.*.%s", names.AttrParameters, names.AttrARN), resourceName1, names.AttrARN),
					resource.TestCheckTypeSetElemAttrPair(dataSourceName, fmt.Sprintf("%s.*.%s", names.AttrParameters, names.AttrName), resourceName1, names.AttrName),
					resource.TestCheckTypeSetElemAttrPair(dataSourceName, fmt.Sprintf("%s.*.%s", names.AttrParameters, names.AttrType), resourceName1, names.AttrType),
					resource.TestCheckTypeSetElemAttrPair(dataSourceName, fmt.Sprintf("%s.*.%s", names.AttrParameters, names.AttrValue), resourceName1, names.AttrValue),
					resource.TestCheckTypeSetElemAttrPair(dataSourceName, fmt.Sprintf("%s.*.%s", names.AttrParameters, names.AttrARN), resourceName2, names.AttrARN),
					resource.TestCheckTypeSetElemAttrPair(dataSourceName, fmt.Sprintf("%s.*.%s", names.AttrParameters, names.AttrName), resourceName2, names.AttrName),
					resource.TestCheckTypeSetElemAttrPair(dataSourceName, fmt.Sprintf("%s.*.%s", names.AttrParameters, names.AttrType), resourceName2, names.AttrType),
					resource.TestCheckTypeSetElemAttrPair(dataSourceName, fmt.Sprintf("%s.*.%s", names.AttrParameters, names.AttrValue), resourceName2, names.AttrValue),
					resource.TestCheckResourceAttr(dataSourceName, "with_decryption", acctest.CtFalse),
					resource.TestCheckResourceAttr(dataSourceName, "recursive", acctest.CtFalse),
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
	dataSourceName := "data.aws_ssm_parameters_by_path.recursive"
	resourceName1 := "aws_ssm_parameter.top_level"
	resourceName2 := "aws_ssm_parameter.nested"
	pathPrefix := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.SSMServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccParametersByPathDataSourceConfig_recursion(pathPrefix),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(dataSourceName, fmt.Sprintf("%s.#", names.AttrParameters), "2"),
					resource.TestCheckTypeSetElemAttrPair(dataSourceName, fmt.Sprintf("%s.*.%s", names.AttrParameters, names.AttrARN), resourceName1, names.AttrARN),
					resource.TestCheckTypeSetElemAttrPair(dataSourceName, fmt.Sprintf("%s.*.%s", names.AttrParameters, names.AttrName), resourceName1, names.AttrName),
					resource.TestCheckTypeSetElemAttrPair(dataSourceName, fmt.Sprintf("%s.*.%s", names.AttrParameters, names.AttrType), resourceName1, names.AttrType),
					resource.TestCheckTypeSetElemAttrPair(dataSourceName, fmt.Sprintf("%s.*.%s", names.AttrParameters, names.AttrValue), resourceName1, names.AttrValue),
					resource.TestCheckTypeSetElemAttrPair(dataSourceName, fmt.Sprintf("%s.*.%s", names.AttrParameters, names.AttrARN), resourceName2, names.AttrARN),
					resource.TestCheckTypeSetElemAttrPair(dataSourceName, fmt.Sprintf("%s.*.%s", names.AttrParameters, names.AttrName), resourceName2, names.AttrName),
					resource.TestCheckTypeSetElemAttrPair(dataSourceName, fmt.Sprintf("%s.*.%s", names.AttrParameters, names.AttrType), resourceName2, names.AttrType),
					resource.TestCheckTypeSetElemAttrPair(dataSourceName, fmt.Sprintf("%s.*.%s", names.AttrParameters, names.AttrValue), resourceName2, names.AttrValue),
					resource.TestCheckResourceAttr(dataSourceName, "recursive", acctest.CtTrue),
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
