// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package cloudfront_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/cloudfront"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfcloudfront "github.com/hashicorp/terraform-provider-aws/internal/service/cloudfront"
)

func TestAccCloudFrontKeyGroup_basic(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_cloudfront_key_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckPartitionHasService(t, cloudfront.EndpointsID) },
		ErrorCheck:               acctest.ErrorCheck(t, cloudfront.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckKeyGroupDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccKeyGroupConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckKeyGroupExistence(ctx, resourceName),
					resource.TestCheckResourceAttr("aws_cloudfront_key_group.test", "comment", "test key group"),
					resource.TestCheckResourceAttrSet("aws_cloudfront_key_group.test", "etag"),
					resource.TestCheckResourceAttrSet("aws_cloudfront_key_group.test", "id"),
					resource.TestCheckResourceAttr("aws_cloudfront_key_group.test", "items.#", "1"),
					resource.TestCheckResourceAttr("aws_cloudfront_key_group.test", "name", rName),
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
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckPartitionHasService(t, cloudfront.EndpointsID) },
		ErrorCheck:               acctest.ErrorCheck(t, cloudfront.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckKeyGroupDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccKeyGroupConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckKeyGroupExistence(ctx, resourceName),
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
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckPartitionHasService(t, cloudfront.EndpointsID) },
		ErrorCheck:               acctest.ErrorCheck(t, cloudfront.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckKeyGroupDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccKeyGroupConfig_comment(rName, firstComment),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckKeyGroupExistence(ctx, resourceName),
					resource.TestCheckResourceAttr("aws_cloudfront_key_group.test", "comment", firstComment),
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
					testAccCheckKeyGroupExistence(ctx, resourceName),
					resource.TestCheckResourceAttr("aws_cloudfront_key_group.test", "comment", secondComment),
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
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckPartitionHasService(t, cloudfront.EndpointsID) },
		ErrorCheck:               acctest.ErrorCheck(t, cloudfront.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckKeyGroupDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccKeyGroupConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckKeyGroupExistence(ctx, resourceName),
					resource.TestCheckResourceAttr("aws_cloudfront_key_group.test", "items.#", "1"),
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
					testAccCheckKeyGroupExistence(ctx, resourceName),
					resource.TestCheckResourceAttr("aws_cloudfront_key_group.test", "items.#", "2"),
				),
			},
		},
	})
}

func testAccCheckKeyGroupExistence(ctx context.Context, r string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[r]
		if !ok {
			return fmt.Errorf("not found: %s", r)
		}
		if rs.Primary.ID == "" {
			return fmt.Errorf("no Id is set")
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).CloudFrontConn(ctx)

		input := &cloudfront.GetKeyGroupInput{
			Id: aws.String(rs.Primary.ID),
		}

		_, err := conn.GetKeyGroupWithContext(ctx, input)
		if err != nil {
			return fmt.Errorf("error retrieving CloudFront key group: %s", err)
		}
		return nil
	}
}

func testAccCheckKeyGroupDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).CloudFrontConn(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_cloudfront_key_group" {
				continue
			}

			input := &cloudfront.GetKeyGroupInput{
				Id: aws.String(rs.Primary.ID),
			}

			_, err := conn.GetKeyGroupWithContext(ctx, input)
			if tfawserr.ErrCodeEquals(err, cloudfront.ErrCodeNoSuchResource) {
				continue
			}
			if err != nil {
				return err
			}
			return fmt.Errorf("CloudFront key group (%s) was not deleted", rs.Primary.ID)
		}

		return nil
	}
}

func testAccKeyGroupBaseConfig(rName string) string {
	return fmt.Sprintf(`
resource "aws_cloudfront_public_key" "test" {
  comment     = "test key"
  encoded_key = file("test-fixtures/cloudfront-public-key.pem")
  name        = %q
}
`, rName)
}

func testAccKeyGroupConfig_basic(rName string) string {
	return testAccKeyGroupBaseConfig(rName) + fmt.Sprintf(`
resource "aws_cloudfront_key_group" "test" {
  comment = "test key group"
  items   = [aws_cloudfront_public_key.test.id]
  name    = %q
}
`, rName)
}

func testAccKeyGroupConfig_comment(rName string, comment string) string {
	return testAccKeyGroupBaseConfig(rName) + fmt.Sprintf(`
resource "aws_cloudfront_key_group" "test" {
  comment = %q
  items   = [aws_cloudfront_public_key.test.id]
  name    = %q
}
`, comment, rName)
}

func testAccKeyGroupConfig_items(rName string) string {
	return testAccKeyGroupBaseConfig(rName) + fmt.Sprintf(`
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
`, rName)
}
