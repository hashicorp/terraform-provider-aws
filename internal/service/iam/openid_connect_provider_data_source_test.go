// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package iam_test

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/service/iam"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
)

func TestAccIAMOpenidConnectProviderDataSource_basic(t *testing.T) {
	ctx := acctest.Context(t)
	rString := sdkacctest.RandString(5)
	dataSourceName := "data.aws_iam_openid_connect_provider.test"
	resourceName := "aws_iam_openid_connect_provider.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, iam.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckOpenIDConnectProviderDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccOpenIDConnectProviderDataSourceConfig_basic(rString),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckOpenIDConnectProviderExists(ctx, resourceName),
					resource.TestCheckResourceAttrPair(dataSourceName, "arn", resourceName, "arn"),
					resource.TestCheckResourceAttrPair(dataSourceName, "url", resourceName, "url"),
					resource.TestCheckResourceAttrPair(dataSourceName, "client_id_list", resourceName, "client_id_list"),
					resource.TestCheckResourceAttrPair(dataSourceName, "thumbprint_list", resourceName, "thumbprint_list"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
				),
			},
		},
	})
}

func TestAccIAMOpenidConnectProviderDataSource_url(t *testing.T) {
	ctx := acctest.Context(t)
	rString := sdkacctest.RandString(5)
	dataSourceName := "data.aws_iam_openid_connect_provider.test"
	resourceName := "aws_iam_openid_connect_provider.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, iam.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckOpenIDConnectProviderDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccOpenIDConnectProviderDataSourceConfig_url(rString),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckOpenIDConnectProviderExists(ctx, resourceName),
					resource.TestCheckResourceAttrPair(dataSourceName, "arn", resourceName, "arn"),
					resource.TestCheckResourceAttrPair(dataSourceName, "url", resourceName, "url"),
					resource.TestCheckResourceAttrPair(dataSourceName, "client_id_list", resourceName, "client_id_list"),
					resource.TestCheckResourceAttrPair(dataSourceName, "thumbprint_list", resourceName, "thumbprint_list"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
				),
			},
		},
	})
}

func TestAccIAMOpenidConnectProviderDataSource_tags(t *testing.T) {
	ctx := acctest.Context(t)
	rString := sdkacctest.RandString(5)
	dataSourceName := "data.aws_iam_openid_connect_provider.test"
	resourceName := "aws_iam_openid_connect_provider.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, iam.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckOpenIDConnectProviderDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccOpenIDConnectProviderDataSourceConfig_tags(rString),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckOpenIDConnectProviderExists(ctx, resourceName),
					resource.TestCheckResourceAttrPair(dataSourceName, "arn", resourceName, "arn"),
					resource.TestCheckResourceAttrPair(dataSourceName, "url", resourceName, "url"),
					resource.TestCheckResourceAttrPair(dataSourceName, "client_id_list", resourceName, "client_id_list"),
					resource.TestCheckResourceAttrPair(dataSourceName, "thumbprint_list", resourceName, "thumbprint_list"),
					resource.TestCheckResourceAttr(dataSourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(dataSourceName, "tags.tag1", "test-value1"),
					resource.TestCheckResourceAttr(dataSourceName, "tags.tag2", "test-value2")),
			},
		},
	})
}

func testAccOpenIDConnectProviderDataSourceConfig_basic(rName string) string {
	return fmt.Sprintf(`
resource "aws_iam_openid_connect_provider" "test" {
  url = "https://accounts.testle.com/%[1]s"

  client_id_list = [
    "266362248691-re108qaeld573ia0l6clj2i5ac7r7291.apps.testleusercontent.com",
  ]

  thumbprint_list = ["cf23df2207d99a74fbe169e3eba035e633b65d94"]
}

data "aws_iam_openid_connect_provider" "test" {
  arn = aws_iam_openid_connect_provider.test.arn
}
`, rName)
}

func testAccOpenIDConnectProviderDataSourceConfig_url(rName string) string {
	return fmt.Sprintf(`
resource "aws_iam_openid_connect_provider" "test" {
  url = "https://accounts.testle.com/%[1]s"

  client_id_list = [
    "266362248691-re108qaeld573ia0l6clj2i5ac7r7291.apps.testleusercontent.com",
  ]

  thumbprint_list = ["cf23df2207d99a74fbe169e3eba035e633b65d94"]
}

data "aws_iam_openid_connect_provider" "test" {
  url = "https://${aws_iam_openid_connect_provider.test.url}"
}
`, rName)
}

func testAccOpenIDConnectProviderDataSourceConfig_tags(rName string) string {
	return fmt.Sprintf(`
resource "aws_iam_openid_connect_provider" "test" {
  url = "https://accounts.testle.com/%[1]s"

  client_id_list = [
    "266362248691-re108qaeld573ia0l6clj2i5ac7r7291.apps.testleusercontent.com",
  ]

  thumbprint_list = ["cf23df2207d99a74fbe169e3eba035e633b65d94"]

  tags = {
    tag1 = "test-value1"
    tag2 = "test-value2"
  }
}

data "aws_iam_openid_connect_provider" "test" {
  arn = aws_iam_openid_connect_provider.test.arn
}
`, rName)
}
