// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package iam_test

import (
	"context"
	"fmt"
	"testing"

	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfiam "github.com/hashicorp/terraform-provider-aws/internal/service/iam"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccIAMOpenIDConnectProvider_basic(t *testing.T) {
	ctx := acctest.Context(t)
	rString := sdkacctest.RandString(5)
	url := "accounts.testle.com/" + rString
	resourceName := "aws_iam_openid_connect_provider.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.IAMServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckOpenIDConnectProviderDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccOpenIDConnectProviderConfig_basic(rString),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckOpenIDConnectProviderExists(ctx, resourceName),
					acctest.CheckResourceAttrGlobalARN(resourceName, "arn", "iam", fmt.Sprintf("oidc-provider/%s", url)),
					resource.TestCheckResourceAttr(resourceName, "url", url),
					resource.TestCheckResourceAttr(resourceName, "client_id_list.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "client_id_list.0",
						"266362248691-re108qaeld573ia0l6clj2i5ac7r7291.apps.testleusercontent.com"),
					resource.TestCheckResourceAttr(resourceName, "thumbprint_list.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccOpenIDConnectProviderConfig_modified(rString),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckOpenIDConnectProviderExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "url", url),
					resource.TestCheckResourceAttr(resourceName, "client_id_list.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "client_id_list.0",
						"266362248691-re108qaeld573ia0l6clj2i5ac7r7291.apps.testleusercontent.com"),
					resource.TestCheckResourceAttr(resourceName, "thumbprint_list.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "thumbprint_list.0", "cf23df2207d99a74fbe169e3eba035e633b65d94"),
					resource.TestCheckResourceAttr(resourceName, "thumbprint_list.1", "c784713d6f9cb67b55dd84f4e4af7832d42b8f55"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
				),
			},
		},
	})
}

func TestAccIAMOpenIDConnectProvider_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	rString := sdkacctest.RandString(5)
	resourceName := "aws_iam_openid_connect_provider.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.IAMServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckOpenIDConnectProviderDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccOpenIDConnectProviderConfig_basic(rString),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckOpenIDConnectProviderExists(ctx, resourceName),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfiam.ResourceOpenIDConnectProvider(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccIAMOpenIDConnectProvider_clientIDListOrder(t *testing.T) {
	ctx := acctest.Context(t)
	rString := sdkacctest.RandString(5)
	resourceName := "aws_iam_openid_connect_provider.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.IAMServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckOpenIDConnectProviderDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccOpenIDConnectProviderConfig_clientIDList_first(rString),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckOpenIDConnectProviderExists(ctx, resourceName),
				),
			},
			{
				Config: testAccOpenIDConnectProviderConfig_clientIDList_second(rString),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckOpenIDConnectProviderExists(ctx, resourceName),
				),
				ExpectNonEmptyPlan: false, // Expect an empty plan as only the order has been changed
				PlanOnly:           true,  // Expect an empty plan as only the order has been changed
			},
		},
	})
}

func testAccCheckOpenIDConnectProviderDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).IAMConn(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_iam_openid_connect_provider" {
				continue
			}

			_, err := tfiam.FindOpenIDConnectProviderByARN(ctx, conn, rs.Primary.ID)

			if tfresource.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("IAM OIDC Provider %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccCheckOpenIDConnectProviderExists(ctx context.Context, n string /*, v *iam.GetOpenIDConnectProviderOutput*/) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not Found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No IAM OIDC Provider ID is set")
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).IAMConn(ctx)

		_, err := tfiam.FindOpenIDConnectProviderByARN(ctx, conn, rs.Primary.ID)

		return err
	}
}

func testAccOpenIDConnectProviderConfig_basic(rName string) string {
	return fmt.Sprintf(`
resource "aws_iam_openid_connect_provider" "test" {
  url = "https://accounts.testle.com/%[1]s"

  client_id_list = [
    "266362248691-re108qaeld573ia0l6clj2i5ac7r7291.apps.testleusercontent.com",
  ]

  thumbprint_list = ["cf23df2207d99a74fbe169e3eba035e633b65d94"]
}
`, rName)
}

func testAccOpenIDConnectProviderConfig_modified(rName string) string {
	return fmt.Sprintf(`
resource "aws_iam_openid_connect_provider" "test" {
  url = "https://accounts.testle.com/%[1]s"

  client_id_list = [
    "266362248691-re108qaeld573ia0l6clj2i5ac7r7291.apps.testleusercontent.com",
  ]

  thumbprint_list = ["cf23df2207d99a74fbe169e3eba035e633b65d94", "c784713d6f9cb67b55dd84f4e4af7832d42b8f55"]
}
`, rName)
}

func testAccOpenIDConnectProviderConfig_tags0(rName string) string {
	return fmt.Sprintf(`
resource "aws_iam_openid_connect_provider" "test" {
  url = "https://accounts.testle.com/%[1]s"

  client_id_list = [
    "266362248691-re108qaeld573ia0l6clj2i5ac7r7291.apps.testleusercontent.com",
  ]

  thumbprint_list = ["cf23df2207d99a74fbe169e3eba035e633b65d94"]
}
`, rName)
}

func testAccOpenIDConnectProviderConfig_tags1(rName, tagKey1, tagValue1 string) string {
	return fmt.Sprintf(`
resource "aws_iam_openid_connect_provider" "test" {
  url = "https://accounts.testle.com/%[1]s"

  client_id_list = [
    "266362248691-re108qaeld573ia0l6clj2i5ac7r7291.apps.testleusercontent.com",
  ]

  thumbprint_list = ["cf23df2207d99a74fbe169e3eba035e633b65d94"]

  tags = {
    %[2]q = %[3]q
  }
}
`, rName, tagKey1, tagValue1)
}

func testAccOpenIDConnectProviderConfig_tags2(rName, tagKey1, tagValue1, tagKey2, tagValue2 string) string {
	return fmt.Sprintf(`
resource "aws_iam_openid_connect_provider" "test" {
  url = "https://accounts.testle.com/%[1]s"

  client_id_list = [
    "266362248691-re108qaeld573ia0l6clj2i5ac7r7291.apps.testleusercontent.com",
  ]

  thumbprint_list = ["cf23df2207d99a74fbe169e3eba035e633b65d94"]

  tags = {
    %[2]q = %[3]q
    %[4]q = %[5]q
  }
}
`, rName, tagKey1, tagValue1, tagKey2, tagValue2)
}

func testAccOpenIDConnectProviderConfig_tagsNull(rName, tagKey1 string) string {
	return fmt.Sprintf(`
resource "aws_iam_openid_connect_provider" "test" {
  url = "https://accounts.testle.com/%[1]s"

  client_id_list = [
    "266362248691-re108qaeld573ia0l6clj2i5ac7r7291.apps.testleusercontent.com",
  ]

  thumbprint_list = ["cf23df2207d99a74fbe169e3eba035e633b65d94"]

  tags = {
    %[2]q = null
  }
}
`, rName, tagKey1)
}

func testAccOpenIDConnectProviderConfig_clientIDList_first(rName string) string {
	return fmt.Sprintf(`
resource "aws_iam_openid_connect_provider" "test" {
  url = "https://accounts.testle.com/%[1]s"

  client_id_list = [
    "abc.testle.com",
    "def.testle.com",
    "ghi.testle.com",
  ]

  thumbprint_list = ["oif8192f189fa2178f-testle.thumbprint.com"]
}
`, rName)
}

func testAccOpenIDConnectProviderConfig_clientIDList_second(rName string) string {
	return fmt.Sprintf(`
resource "aws_iam_openid_connect_provider" "test" {
  url = "https://accounts.testle.com/%[1]s"

  client_id_list = [
    "def.testle.com",
    "ghi.testle.com",
    "abc.testle.com",
  ]

  thumbprint_list = ["oif8192f189fa2178f-testle.thumbprint.com"]
}
`, rName)
}
