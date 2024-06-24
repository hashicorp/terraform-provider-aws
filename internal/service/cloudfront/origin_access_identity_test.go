// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package cloudfront_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/service/cloudfront"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/id"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfcloudfront "github.com/hashicorp/terraform-provider-aws/internal/service/cloudfront"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccCloudFrontOriginAccessIdentity_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var origin cloudfront.GetCloudFrontOriginAccessIdentityOutput
	resourceName := "aws_cloudfront_origin_access_identity.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckPartitionHasService(t, names.CloudFrontEndpointID) },
		ErrorCheck:               acctest.ErrorCheck(t, names.CloudFrontServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckOriginAccessIdentityDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccOriginAccessIdentityConfig_basic,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckOriginAccessIdentityExists(ctx, resourceName, &origin),
					resource.TestCheckResourceAttr(resourceName, names.AttrComment, "some comment"),
					resource.TestMatchResourceAttr(resourceName, "caller_reference", regexache.MustCompile(fmt.Sprintf("^%s", id.UniqueIdPrefix))),
					resource.TestMatchResourceAttr(resourceName, "s3_canonical_user_id", regexache.MustCompile("^[0-9a-z]+")),
					resource.TestMatchResourceAttr(resourceName, "cloudfront_access_identity_path", regexache.MustCompile("^origin-access-identity/cloudfront/[0-9A-Z]+")),
					//lintignore:AWSAT001
					resource.TestMatchResourceAttr(resourceName, "iam_arn", regexache.MustCompile(fmt.Sprintf("^arn:%s:iam::cloudfront:user/CloudFront Origin Access Identity [0-9A-Z]+", acctest.Partition()))),
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

func TestAccCloudFrontOriginAccessIdentity_noComment(t *testing.T) {
	ctx := acctest.Context(t)
	var origin cloudfront.GetCloudFrontOriginAccessIdentityOutput
	resourceName := "aws_cloudfront_origin_access_identity.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckPartitionHasService(t, names.CloudFrontEndpointID) },
		ErrorCheck:               acctest.ErrorCheck(t, names.CloudFrontServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckOriginAccessIdentityDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccOriginAccessIdentityConfig_noComment,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckOriginAccessIdentityExists(ctx, resourceName, &origin),
					resource.TestMatchResourceAttr(resourceName, "caller_reference", regexache.MustCompile(fmt.Sprintf("^%s", id.UniqueIdPrefix))),
					resource.TestMatchResourceAttr(resourceName, "s3_canonical_user_id", regexache.MustCompile("^[0-9a-z]+")),
					resource.TestMatchResourceAttr(resourceName, "cloudfront_access_identity_path", regexache.MustCompile("^origin-access-identity/cloudfront/[0-9A-Z]+")),
					//lintignore:AWSAT001
					resource.TestMatchResourceAttr(resourceName, "iam_arn", regexache.MustCompile(fmt.Sprintf("^arn:%s:iam::cloudfront:user/CloudFront Origin Access Identity [0-9A-Z]+", acctest.Partition()))),
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

func TestAccCloudFrontOriginAccessIdentity_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	var origin cloudfront.GetCloudFrontOriginAccessIdentityOutput
	resourceName := "aws_cloudfront_origin_access_identity.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckPartitionHasService(t, names.CloudFrontEndpointID) },
		ErrorCheck:               acctest.ErrorCheck(t, names.CloudFrontServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckOriginAccessIdentityDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccOriginAccessIdentityConfig_basic,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckOriginAccessIdentityExists(ctx, resourceName, &origin),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfcloudfront.ResourceOriginAccessIdentity(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckOriginAccessIdentityDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).CloudFrontClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_cloudfront_origin_access_identity" {
				continue
			}

			_, err := tfcloudfront.FindOriginAccessIdentityByID(ctx, conn, rs.Primary.ID)

			if tfresource.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("CloudFront Origin Access Identity %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccCheckOriginAccessIdentityExists(ctx context.Context, n string, v *cloudfront.GetCloudFrontOriginAccessIdentityOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).CloudFrontClient(ctx)

		output, err := tfcloudfront.FindOriginAccessIdentityByID(ctx, conn, rs.Primary.ID)

		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

const testAccOriginAccessIdentityConfig_basic = `
resource "aws_cloudfront_origin_access_identity" "test" {
  comment = "some comment"
}
`

const testAccOriginAccessIdentityConfig_noComment = `
resource "aws_cloudfront_origin_access_identity" "test" {}
`
