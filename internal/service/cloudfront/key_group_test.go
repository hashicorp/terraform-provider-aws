// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package cloudfront_test

import (
	"context"
	"fmt"
	"testing"

	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfcloudfront "github.com/hashicorp/terraform-provider-aws/internal/service/cloudfront"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccCloudFrontKeyGroup_basic(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_cloudfront_key_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckPartitionHasService(t, names.CloudFrontEndpointID) },
		ErrorCheck:               acctest.ErrorCheck(t, names.CloudFrontServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckKeyGroupDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccKeyGroupConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckKeyGroupExists(ctx, resourceName),
					resource.TestCheckResourceAttr("aws_cloudfront_key_group.test", names.AttrComment, "test key group"),
					resource.TestCheckResourceAttrSet("aws_cloudfront_key_group.test", "etag"),
					resource.TestCheckResourceAttrSet("aws_cloudfront_key_group.test", names.AttrID),
					resource.TestCheckResourceAttr("aws_cloudfront_key_group.test", "items.#", acctest.Ct1),
					resource.TestCheckResourceAttr("aws_cloudfront_key_group.test", names.AttrName, rName),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccCloudFrontKeyGroup_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_cloudfront_key_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckPartitionHasService(t, names.CloudFrontEndpointID) },
		ErrorCheck:               acctest.ErrorCheck(t, names.CloudFrontServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckKeyGroupDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccKeyGroupConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckKeyGroupExists(ctx, resourceName),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfcloudfront.ResourceKeyGroup(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccCloudFrontKeyGroup_comment(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_cloudfront_key_group.test"

	firstComment := "first comment"
	secondComment := "second comment"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckPartitionHasService(t, names.CloudFrontEndpointID) },
		ErrorCheck:               acctest.ErrorCheck(t, names.CloudFrontServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckKeyGroupDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccKeyGroupConfig_comment(rName, firstComment),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckKeyGroupExists(ctx, resourceName),
					resource.TestCheckResourceAttr("aws_cloudfront_key_group.test", names.AttrComment, firstComment),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccKeyGroupConfig_comment(rName, secondComment),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckKeyGroupExists(ctx, resourceName),
					resource.TestCheckResourceAttr("aws_cloudfront_key_group.test", names.AttrComment, secondComment),
				),
			},
		},
	})
}

func TestAccCloudFrontKeyGroup_items(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_cloudfront_key_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckPartitionHasService(t, names.CloudFrontEndpointID) },
		ErrorCheck:               acctest.ErrorCheck(t, names.CloudFrontServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckKeyGroupDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccKeyGroupConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckKeyGroupExists(ctx, resourceName),
					resource.TestCheckResourceAttr("aws_cloudfront_key_group.test", "items.#", acctest.Ct1),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccKeyGroupConfig_items(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckKeyGroupExists(ctx, resourceName),
					resource.TestCheckResourceAttr("aws_cloudfront_key_group.test", "items.#", acctest.Ct2),
				),
			},
		},
	})
}

func testAccCheckKeyGroupExists(ctx context.Context, n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).CloudFrontClient(ctx)

		_, err := tfcloudfront.FindKeyGroupByID(ctx, conn, rs.Primary.ID)

		return err
	}
}

func testAccCheckKeyGroupDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).CloudFrontClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_cloudfront_key_group" {
				continue
			}

			_, err := tfcloudfront.FindKeyGroupByID(ctx, conn, rs.Primary.ID)

			if tfresource.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("CloudFront Key Group %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccKeyGroupConfig_base(rName string) string {
	return fmt.Sprintf(`
resource "aws_cloudfront_public_key" "test" {
  comment     = "test key"
  encoded_key = file("test-fixtures/cloudfront-public-key.pem")
  name        = %[1]q
}
`, rName)
}

func testAccKeyGroupConfig_basic(rName string) string {
	return acctest.ConfigCompose(testAccKeyGroupConfig_base(rName), fmt.Sprintf(`
resource "aws_cloudfront_key_group" "test" {
  comment = "test key group"
  items   = [aws_cloudfront_public_key.test.id]
  name    = %[1]q
}
`, rName))
}

func testAccKeyGroupConfig_comment(rName, comment string) string {
	return acctest.ConfigCompose(testAccKeyGroupConfig_base(rName), fmt.Sprintf(`
resource "aws_cloudfront_key_group" "test" {
  comment = %[2]q
  items   = [aws_cloudfront_public_key.test.id]
  name    = %[1]q
}
`, rName, comment))
}

func testAccKeyGroupConfig_items(rName string) string {
	return acctest.ConfigCompose(testAccKeyGroupConfig_base(rName), fmt.Sprintf(`
resource "aws_cloudfront_public_key" "test2" {
  comment     = "second test key"
  encoded_key = file("test-fixtures/cloudfront-public-key.pem")
  name        = "%[1]s-second"
}

resource "aws_cloudfront_key_group" "test" {
  comment = "test key group"
  items   = [aws_cloudfront_public_key.test.id, aws_cloudfront_public_key.test2.id]
  name    = %[1]q
}
`, rName))
}
