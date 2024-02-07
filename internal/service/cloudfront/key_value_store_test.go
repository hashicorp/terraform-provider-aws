// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package cloudfront_test

import (
	"context"
	"errors"
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
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/names"

	tfcloudfront "github.com/hashicorp/terraform-provider-aws/internal/service/cloudfront"
)

func TestAccCloudFrontKeyValueStore_basic(t *testing.T) {
	ctx := acctest.Context(t)

	var keyvaluestore cloudfront.DescribeKeyValueStoreOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_cloudfront_key_value_store.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.CloudFront)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.CloudFront),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckKeyValueStoreDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccKeyValueStoreConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckKeyValueStoreExists(ctx, resourceName, &keyvaluestore),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"last_modified_time"},
			},
		},
	})
}

func TestAccCloudFrontKeyValueStore_disappears(t *testing.T) {
	ctx := acctest.Context(t)

	var keyvaluestore cloudfront.DescribeKeyValueStoreOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_cloudfront_key_value_store.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.CloudFront)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.CloudFront),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckKeyValueStoreDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccKeyValueStoreConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckKeyValueStoreExists(ctx, resourceName, &keyvaluestore),
					acctest.CheckFrameworkResourceDisappears(ctx, acctest.Provider, tfcloudfront.ResourceKeyValueStore, resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccCloudFrontKeyValueStore_Comment(t *testing.T) {
	ctx := acctest.Context(t)

	var keyvaluestore cloudfront.DescribeKeyValueStoreOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_cloudfront_key_value_store.test"

	comment1 := "comment1"
	comment2 := "comment2"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.CloudFront),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckKeyValueStoreDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccCloudFrontKeyValueStoreConfigComment(rName, comment1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckKeyValueStoreExists(ctx, resourceName, &keyvaluestore),
					resource.TestCheckResourceAttr(resourceName, "comment", comment1),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"last_modified_time"},
			},
			{
				Config: testAccCloudFrontKeyValueStoreConfigComment(rName, comment2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckKeyValueStoreExists(ctx, resourceName, &keyvaluestore),
					resource.TestCheckResourceAttr(resourceName, "comment", comment2),
				),
			},
		},
	})
}

func testAccCheckKeyValueStoreDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).CloudFrontClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_cloudfront_key_value_store" {
				continue
			}

			_, err := conn.DescribeKeyValueStore(ctx, &cloudfront.DescribeKeyValueStoreInput{
				Name: aws.String(rs.Primary.ID),
			})
			if errs.IsA[*awstypes.EntityNotFound](err) {
				return nil
			}
			if err != nil {
				return create.Error(names.CloudFront, create.ErrActionCheckingDestroyed, tfcloudfront.ResNameKeyValueStore, rs.Primary.ID, err)
			}

			return create.Error(names.CloudFront, create.ErrActionCheckingDestroyed, tfcloudfront.ResNameKeyValueStore, rs.Primary.ID, errors.New("not destroyed"))
		}

		return nil
	}
}

func testAccCheckKeyValueStoreExists(ctx context.Context, name string, keyvaluestore *cloudfront.DescribeKeyValueStoreOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return create.Error(names.CloudFront, create.ErrActionCheckingExistence, tfcloudfront.ResNameKeyValueStore, name, errors.New("not found"))
		}

		if rs.Primary.ID == "" {
			return create.Error(names.CloudFront, create.ErrActionCheckingExistence, tfcloudfront.ResNameKeyValueStore, name, errors.New("not set"))
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).CloudFrontClient(ctx)
		resp, err := conn.DescribeKeyValueStore(ctx, &cloudfront.DescribeKeyValueStoreInput{
			Name: aws.String(rs.Primary.ID),
		})

		if err != nil {
			return create.Error(names.CloudFront, create.ErrActionCheckingExistence, tfcloudfront.ResNameKeyValueStore, rs.Primary.ID, err)
		}

		*keyvaluestore = *resp

		return nil
	}
}

func testAccCloudFrontKeyValueStoreConfigComment(rName string, comment string) string {
	return fmt.Sprintf(`
resource "aws_cloudfront_key_value_store" "test" {
	name    = %[1]q
  comment = %[2]q
}
`, rName, comment)
}

func testAccKeyValueStoreConfig_basic(rName string) string {
	return fmt.Sprintf(`
resource "aws_cloudfront_key_value_store" "test" {
  name = %[1]q
}
`, rName)
}
