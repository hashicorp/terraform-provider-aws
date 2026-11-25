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
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccIAMOpenIDConnectProviderDataSource_basic(t *testing.T) {
	ctx := acctest.Context(t)
	rString := acctest.RandString(t, 5)
	dataSourceName := "data.aws_iam_openid_connect_provider.test"
	resourceName := "aws_iam_openid_connect_provider.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.IAMServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckOpenIDConnectProviderDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccOpenIDConnectProviderDataSourceConfig_basic(rString),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckOpenIDConnectProviderExists(ctx, t, resourceName),
					resource.TestCheckResourceAttrPair(dataSourceName, names.AttrARN, resourceName, names.AttrARN),
					resource.TestCheckResourceAttrPair(dataSourceName, names.AttrURL, resourceName, names.AttrURL),
					resource.TestCheckResourceAttrPair(dataSourceName, "client_id_list", resourceName, "client_id_list"),
					resource.TestCheckResourceAttrPair(dataSourceName, "thumbprint_list", resourceName, "thumbprint_list"),
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(dataSourceName, tfjsonpath.New(names.AttrTags), knownvalue.MapExact(map[string]knownvalue.Check{})),
				},
			},
		},
	})
}

func TestAccIAMOpenIDConnectProviderDataSource_url(t *testing.T) {
	ctx := acctest.Context(t)
	rString := acctest.RandString(t, 5)
	dataSourceName := "data.aws_iam_openid_connect_provider.test"
	resourceName := "aws_iam_openid_connect_provider.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.IAMServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckOpenIDConnectProviderDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccOpenIDConnectProviderDataSourceConfig_url(rString),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckOpenIDConnectProviderExists(ctx, t, resourceName),
					resource.TestCheckResourceAttrPair(dataSourceName, names.AttrARN, resourceName, names.AttrARN),
					resource.TestCheckResourceAttrPair(dataSourceName, names.AttrURL, resourceName, names.AttrURL),
					resource.TestCheckResourceAttrPair(dataSourceName, "client_id_list", resourceName, "client_id_list"),
					resource.TestCheckResourceAttrPair(dataSourceName, "thumbprint_list", resourceName, "thumbprint_list"),
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(dataSourceName, tfjsonpath.New(names.AttrTags), knownvalue.MapExact(map[string]knownvalue.Check{})),
				},
			},
		},
	})
}

func TestAccIAMOpenIDConnectProviderDataSource_trustedIdentityPropagation(t *testing.T) {
	ctx := acctest.Context(t)
	rString := acctest.RandString(t, 5)
	dataSourceName := "data.aws_iam_openid_connect_provider.test"
	resourceName := "aws_iam_openid_connect_provider.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckSSOAdminInstances(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.IAMServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckOpenIDConnectProviderDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccOpenIDConnectProviderDataSourceConfig_trustedIdentityPropagation(rString),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckOpenIDConnectProviderExists(ctx, t, resourceName),
					resource.TestCheckResourceAttr(dataSourceName, "context_provider.#", "1"),
					resource.TestCheckResourceAttr(dataSourceName, "context_provider.0.subject_source_key", "sub"),
					resource.TestCheckResourceAttr(dataSourceName, "context_provider.0.subject_type_value", "user"),
					resource.TestCheckResourceAttr(dataSourceName, "context_provider.0.key_definition.#", "1"),
					resource.TestCheckResourceAttr(dataSourceName, "context_provider.0.key_definition.0.source_key", names.AttrEmail),
					resource.TestCheckResourceAttr(dataSourceName, "context_provider.0.key_definition.0.target_key", "Email"),
					resource.TestCheckResourceAttrPair(dataSourceName, "identity_center_instance_arn", resourceName, "identity_center_instance_arn"),
					resource.TestCheckResourceAttrPair(dataSourceName, "identity_store_arn", resourceName, "identity_store_arn"),
				),
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

func testAccOpenIDConnectProviderDataSourceConfig_trustedIdentityPropagation(rName string) string {
	return fmt.Sprintf(`
data "aws_ssoadmin_instances" "test" {}

resource "aws_iam_openid_connect_provider" "test" {
  url = "https://accounts.testle.com/%[1]s"

  client_id_list = [
    "266362248691-re108qaeld573ia0l6clj2i5ac7r7291.apps.testleusercontent.com",
  ]

  thumbprint_list = ["cf23df2207d99a74fbe169e3eba035e633b65d94"]

  identity_center_instance_arn = tolist(data.aws_ssoadmin_instances.test.arns)[0]

  context_provider {
    subject_source_key = "sub"
    subject_type_value = "user"

    key_definition {
      source_key = "email"
      target_key = "Email"
    }
  }
}

data "aws_iam_openid_connect_provider" "test" {
  arn = aws_iam_openid_connect_provider.test.arn
}
`, rName)
}
