// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package ssm_test

import (
	"fmt"
	"testing"

	awstypes "github.com/aws/aws-sdk-go-v2/service/ssm/types"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccSSMParameterDataSource_basic(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "data.aws_ssm_parameter.test"
	name := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.SSMServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccParameterDataSourceConfig_basic(name, false),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrPair(resourceName, names.AttrARN, "aws_ssm_parameter.test", names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, name),
					resource.TestCheckResourceAttr(resourceName, names.AttrType, "String"),
					resource.TestCheckResourceAttr(resourceName, names.AttrValue, "TestValue"),
					resource.TestCheckResourceAttr(resourceName, "with_decryption", acctest.CtFalse),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrVersion),
				),
			},
			{
				Config: testAccParameterDataSourceConfig_basic(name, true),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrPair(resourceName, names.AttrARN, "aws_ssm_parameter.test", names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, name),
					resource.TestCheckResourceAttr(resourceName, names.AttrType, "String"),
					resource.TestCheckResourceAttr(resourceName, names.AttrValue, "TestValue"),
					resource.TestCheckResourceAttr(resourceName, "with_decryption", acctest.CtTrue),
				),
			},
		},
	})
}

func TestAccSSMParameterDataSource_fullPath(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "data.aws_ssm_parameter.test"
	name := acctest.RandomWithPrefix(t, "/tf-acc-test/tf-acc-test")

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.SSMServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccParameterDataSourceConfig_basic(name, false),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrPair(resourceName, names.AttrARN, "aws_ssm_parameter.test", names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, name),
					resource.TestCheckResourceAttr(resourceName, names.AttrType, "String"),
					resource.TestCheckResourceAttr(resourceName, names.AttrValue, "TestValue"),
					resource.TestCheckResourceAttr(resourceName, "with_decryption", acctest.CtFalse),
				),
			},
		},
	})
}

func TestAccSSMParameterDataSource_insecureValue(t *testing.T) {
	ctx := acctest.Context(t)
	var param awstypes.Parameter
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_ssm_parameter.test"
	dataSourceName := "data.aws_ssm_parameter.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.SSMServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckParameterDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccParameterConfig_insecureValue(rName, "String"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckParameterExists(ctx, t, dataSourceName, &param),
					resource.TestCheckResourceAttrPair(dataSourceName, "insecure_value", resourceName, "insecure_value"),
				),
			},
		},
	})
}

func testAccParameterDataSourceConfig_basic(name string, withDecryption bool) string {
	return fmt.Sprintf(`
resource "aws_ssm_parameter" "test" {
  name  = %[1]q
  type  = "String"
  value = "TestValue"
}

data "aws_ssm_parameter" "test" {
  name            = aws_ssm_parameter.test.name
  with_decryption = %[2]t
}
`, name, withDecryption)
}

func testAccParameterConfig_insecureValue(rName, pType string) string {
	return fmt.Sprintf(`
resource "aws_ssm_parameter" "test" {
  name           = %[1]q
  type           = %[2]q
  insecure_value = "notsecret"
}

data "aws_ssm_parameter" "test" {
  name = aws_ssm_parameter.test.name
}
`, rName, pType)
}
