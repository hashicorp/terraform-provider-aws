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

func TestAccCloudFrontPublicKey_basic(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_cloudfront_public_key.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckPartitionHasService(t, names.CloudFrontEndpointID) },
		ErrorCheck:               acctest.ErrorCheck(t, names.CloudFrontServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckPublicKeyDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccPublicKeyConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckPublicKeyExists(ctx, resourceName),
					resource.TestCheckResourceAttrSet(resourceName, "caller_reference"),
					resource.TestCheckResourceAttr(resourceName, names.AttrComment, ""),
					resource.TestCheckResourceAttrSet(resourceName, "encoded_key"),
					resource.TestCheckResourceAttrSet(resourceName, "etag"),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, names.AttrNamePrefix, ""),
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

func TestAccCloudFrontPublicKey_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_cloudfront_public_key.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckPartitionHasService(t, names.CloudFrontEndpointID) },
		ErrorCheck:               acctest.ErrorCheck(t, names.CloudFrontServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckPublicKeyDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccPublicKeyConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPublicKeyExists(ctx, resourceName),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfcloudfront.ResourcePublicKey(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccCloudFrontPublicKey_nameGenerated(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_cloudfront_public_key.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckPartitionHasService(t, names.CloudFrontEndpointID) },
		ErrorCheck:               acctest.ErrorCheck(t, names.CloudFrontServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckPublicKeyDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccPublicKeyConfig_nameGenerated(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPublicKeyExists(ctx, resourceName),
					acctest.CheckResourceAttrNameGeneratedWithPrefix(resourceName, names.AttrName, "tf-"),
					resource.TestCheckResourceAttr(resourceName, names.AttrNamePrefix, "tf-"),
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

func TestAccCloudFrontPublicKey_namePrefix(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_cloudfront_public_key.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckPartitionHasService(t, names.CloudFrontEndpointID) },
		ErrorCheck:               acctest.ErrorCheck(t, names.CloudFrontServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckPublicKeyDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccPublicKeyConfig_namePrefix("tf-acc-test-prefix-"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPublicKeyExists(ctx, resourceName),
					acctest.CheckResourceAttrNameFromPrefix(resourceName, names.AttrName, "tf-acc-test-prefix-"),
					resource.TestCheckResourceAttr(resourceName, names.AttrNamePrefix, "tf-acc-test-prefix-"),
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

func TestAccCloudFrontPublicKey_update(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_cloudfront_public_key.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckPartitionHasService(t, names.CloudFrontEndpointID) },
		ErrorCheck:               acctest.ErrorCheck(t, names.CloudFrontServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckPublicKeyDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccPublicKeyConfig_comment(rName, "comment 1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPublicKeyExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, names.AttrComment, "comment 1"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccPublicKeyConfig_comment(rName, "comment 2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPublicKeyExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, names.AttrComment, "comment 2"),
				),
			},
		},
	})
}

func testAccCheckPublicKeyExists(ctx context.Context, n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).CloudFrontClient(ctx)

		_, err := tfcloudfront.FindPublicKeyByID(ctx, conn, rs.Primary.ID)

		return err
	}
}

func testAccCheckPublicKeyDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).CloudFrontClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_cloudfront_public_key" {
				continue
			}

			_, err := tfcloudfront.FindPublicKeyByID(ctx, conn, rs.Primary.ID)

			if tfresource.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("CloudFront Public Key %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccPublicKeyConfig_basic(rName string) string {
	return fmt.Sprintf(`
resource "aws_cloudfront_public_key" "test" {
  encoded_key = file("test-fixtures/cloudfront-public-key.pem")
  name        = %[1]q
}
`, rName)
}

func testAccPublicKeyConfig_nameGenerated() string {
	return `
resource "aws_cloudfront_public_key" "test" {
  encoded_key = file("test-fixtures/cloudfront-public-key.pem")
}
`
}

func testAccPublicKeyConfig_namePrefix(namePrefix string) string {
	return fmt.Sprintf(`
resource "aws_cloudfront_public_key" "test" {
  encoded_key = file("test-fixtures/cloudfront-public-key.pem")
  name_prefix = %[1]q
}
`, namePrefix)
}

func testAccPublicKeyConfig_comment(rName, comment string) string {
	return fmt.Sprintf(`
resource "aws_cloudfront_public_key" "test" {
  comment     = %[2]q
  encoded_key = file("test-fixtures/cloudfront-public-key.pem")
  name        = %[1]q
}
`, rName, comment)
}
