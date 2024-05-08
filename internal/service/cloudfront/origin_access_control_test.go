// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package cloudfront_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/cloudfront"
	awstypes "github.com/aws/aws-sdk-go-v2/service/cloudfront/types"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfcloudfront "github.com/hashicorp/terraform-provider-aws/internal/service/cloudfront"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccCloudFrontOriginAccessControl_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var originaccesscontrol awstypes.OriginAccessControl
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_cloudfront_origin_access_control.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.CloudFrontEndpointID)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.CloudFrontServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckOriginAccessControlDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccOriginAccessControlConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckOriginAccessControlExists(ctx, resourceName, &originaccesscontrol),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, "Managed by Terraform"),
					resource.TestCheckResourceAttrSet(resourceName, "etag"),
					resource.TestCheckResourceAttrWith(resourceName, names.AttrID, func(value string) error {
						if value == "" {
							return fmt.Errorf("expected attribute to be set")
						}

						if id := aws.ToString(originaccesscontrol.Id); value != id {
							return fmt.Errorf("expected attribute to be equal to %s", id)
						}

						return nil
					}),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, "origin_access_control_origin_type", "s3"),
					resource.TestCheckResourceAttr(resourceName, "signing_behavior", "always"),
					resource.TestCheckResourceAttr(resourceName, "signing_protocol", "sigv4"),
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

func TestAccCloudFrontOriginAccessControl_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	var originaccesscontrol awstypes.OriginAccessControl
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_cloudfront_origin_access_control.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.CloudFrontEndpointID)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.CloudFrontServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckOriginAccessControlDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccOriginAccessControlConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckOriginAccessControlExists(ctx, resourceName, &originaccesscontrol),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfcloudfront.ResourceOriginAccessControl(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccCloudFrontOriginAccessControl_Name(t *testing.T) {
	ctx := acctest.Context(t)
	rName1 := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	rName2 := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_cloudfront_origin_access_control.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.CloudFrontEndpointID)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.CloudFrontServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckOriginAccessControlDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccOriginAccessControlConfig_name(rName1),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName1),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccOriginAccessControlConfig_name(rName2),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName2),
				),
			},
		},
	})
}

func TestAccCloudFrontOriginAccessControl_Description(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_cloudfront_origin_access_control.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.CloudFrontEndpointID)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.CloudFrontServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckOriginAccessControlDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccOriginAccessControlConfig_description(rName, "Acceptance Test 1"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, "Acceptance Test 1"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccOriginAccessControlConfig_description(rName, "Acceptance Test 2"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, "Acceptance Test 2"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccOriginAccessControlConfig_description(rName, ""),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, ""),
				),
			},
		},
	})
}

func TestAccCloudFrontOriginAccessControl_SigningBehavior(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_cloudfront_origin_access_control.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.CloudFrontEndpointID)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.CloudFrontServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckOriginAccessControlDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccOriginAccessControlConfig_signingBehavior(rName, "never"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "signing_behavior", "never"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccOriginAccessControlConfig_signingBehavior(rName, "no-override"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "signing_behavior", "no-override"),
				),
			},
		},
	})
}

func TestAccCloudFrontOriginAccessControl_lambdaOriginType(t *testing.T) {
	ctx := acctest.Context(t)
	var originaccesscontrol awstypes.OriginAccessControl
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_cloudfront_origin_access_control.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.CloudFrontEndpointID)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.CloudFrontServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckOriginAccessControlDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccOriginAccessControlConfig_originType(rName, "lambda"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckOriginAccessControlExists(ctx, resourceName, &originaccesscontrol),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, "Managed by Terraform"),
					resource.TestCheckResourceAttrSet(resourceName, "etag"),
					resource.TestCheckResourceAttrWith(resourceName, names.AttrID, func(value string) error {
						if value == "" {
							return fmt.Errorf("expected attribute to be set")
						}

						if id := aws.ToString(originaccesscontrol.Id); value != id {
							return fmt.Errorf("expected attribute to be equal to %s", id)
						}

						return nil
					}),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, "origin_access_control_origin_type", "lambda"),
					resource.TestCheckResourceAttr(resourceName, "signing_behavior", "always"),
					resource.TestCheckResourceAttr(resourceName, "signing_protocol", "sigv4"),
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

func TestAccCloudFrontOriginAccessControl_mediaPackageV2Type(t *testing.T) {
	ctx := acctest.Context(t)
	var originaccesscontrol awstypes.OriginAccessControl
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_cloudfront_origin_access_control.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.CloudFrontEndpointID)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.CloudFrontServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckOriginAccessControlDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccOriginAccessControlConfig_originType(rName, "mediapackagev2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckOriginAccessControlExists(ctx, resourceName, &originaccesscontrol),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, "Managed by Terraform"),
					resource.TestCheckResourceAttrSet(resourceName, "etag"),
					resource.TestCheckResourceAttrWith(resourceName, names.AttrID, func(value string) error {
						if value == "" {
							return fmt.Errorf("expected attribute to be set")
						}

						if id := aws.ToString(originaccesscontrol.Id); value != id {
							return fmt.Errorf("expected attribute to be equal to %s", id)
						}

						return nil
					}),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, "origin_access_control_origin_type", "mediapackagev2"),
					resource.TestCheckResourceAttr(resourceName, "signing_behavior", "always"),
					resource.TestCheckResourceAttr(resourceName, "signing_protocol", "sigv4"),
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

func testAccCheckOriginAccessControlDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).CloudFrontClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_cloudfront_origin_access_control" {
				continue
			}

			_, err := tfcloudfront.FindOriginAccessControlByID(ctx, conn, rs.Primary.ID)

			if tfresource.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("CloudFront Origin Access Control %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccCheckOriginAccessControlExists(ctx context.Context, n string, v *awstypes.OriginAccessControl) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).CloudFrontClient(ctx)

		output, err := tfcloudfront.FindOriginAccessControlByID(ctx, conn, rs.Primary.ID)

		if err != nil {
			return err
		}

		*v = *output.OriginAccessControl

		return nil
	}
}

func testAccPreCheck(ctx context.Context, t *testing.T) {
	conn := acctest.Provider.Meta().(*conns.AWSClient).CloudFrontClient(ctx)

	input := &cloudfront.ListOriginAccessControlsInput{}
	_, err := conn.ListOriginAccessControls(ctx, input)

	if acctest.PreCheckSkipError(err) {
		t.Skipf("skipping acceptance testing: %s", err)
	}

	if err != nil {
		t.Fatalf("unexpected PreCheck error: %s", err)
	}
}

func testAccOriginAccessControlConfig_basic(rName string) string {
	return fmt.Sprintf(`
resource "aws_cloudfront_origin_access_control" "test" {
  name                              = %[1]q
  origin_access_control_origin_type = "s3"
  signing_behavior                  = "always"
  signing_protocol                  = "sigv4"
}
`, rName)
}

func testAccOriginAccessControlConfig_name(rName string) string {
	return fmt.Sprintf(`
resource "aws_cloudfront_origin_access_control" "test" {
  name                              = %[1]q
  origin_access_control_origin_type = "s3"
  signing_behavior                  = "always"
  signing_protocol                  = "sigv4"
}
`, rName)
}

func testAccOriginAccessControlConfig_description(rName, description string) string {
	return fmt.Sprintf(`
resource "aws_cloudfront_origin_access_control" "test" {
  name                              = %[1]q
  description                       = %[2]q
  origin_access_control_origin_type = "s3"
  signing_behavior                  = "always"
  signing_protocol                  = "sigv4"
}
`, rName, description)
}

func testAccOriginAccessControlConfig_signingBehavior(rName, signingBehavior string) string {
	return fmt.Sprintf(`
resource "aws_cloudfront_origin_access_control" "test" {
  name                              = %[1]q
  origin_access_control_origin_type = "s3"
  signing_behavior                  = %[2]q
  signing_protocol                  = "sigv4"
}
`, rName, signingBehavior)
}

func testAccOriginAccessControlConfig_originType(rName, originType string) string {
	return fmt.Sprintf(`
resource "aws_cloudfront_origin_access_control" "test" {
  name                              = %[1]q
  origin_access_control_origin_type = %[2]q
  signing_behavior                  = "always"
  signing_protocol                  = "sigv4"
}
`, rName, originType)
}
