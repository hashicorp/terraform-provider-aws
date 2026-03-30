// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package cloudfront_test

import (
	"context"
	"fmt"
	"testing"

	awstypes "github.com/aws/aws-sdk-go-v2/service/cloudfront/types"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	tfcloudfront "github.com/hashicorp/terraform-provider-aws/internal/service/cloudfront"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccCloudFrontKeyValueStore_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var keyvaluestore awstypes.KeyValueStore
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_cloudfront_key_value_store.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.CloudFront)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.CloudFront),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckKeyValueStoreDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccKeyValueStoreConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckKeyValueStoreExists(ctx, t, resourceName, &keyvaluestore),
					acctest.CheckResourceAttrGlobalARNFormat(ctx, resourceName, names.AttrARN, "cloudfront", "key-value-store/{id}"),
					resource.TestCheckNoResourceAttr(resourceName, names.AttrComment),
					resource.TestCheckResourceAttrSet(resourceName, "etag"),
					resource.TestCheckResourceAttrSet(resourceName, "last_modified_time"),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateIdFunc: acctest.AttrImportStateIdFunc(resourceName, names.AttrName),
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccCloudFrontKeyValueStore_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	var keyvaluestore awstypes.KeyValueStore
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_cloudfront_key_value_store.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.CloudFront)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.CloudFront),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckKeyValueStoreDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccKeyValueStoreConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckKeyValueStoreExists(ctx, t, resourceName, &keyvaluestore),
					acctest.CheckFrameworkResourceDisappears(ctx, t, tfcloudfront.ResourceKeyValueStore, resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccCloudFrontKeyValueStore_comment(t *testing.T) {
	ctx := acctest.Context(t)
	var keyvaluestore awstypes.KeyValueStore
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_cloudfront_key_value_store.test"
	comment1 := "comment1"
	comment2 := "comment2"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.CloudFront),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckKeyValueStoreDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccKeyValueStoreConfig_comment(rName, comment1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckKeyValueStoreExists(ctx, t, resourceName, &keyvaluestore),
					resource.TestCheckResourceAttr(resourceName, names.AttrComment, comment1),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateIdFunc: acctest.AttrImportStateIdFunc(resourceName, names.AttrName),
				ImportStateVerify: true,
			},
			{
				Config: testAccKeyValueStoreConfig_comment(rName, comment2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckKeyValueStoreExists(ctx, t, resourceName, &keyvaluestore),
					resource.TestCheckResourceAttr(resourceName, names.AttrComment, comment2),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateIdFunc: acctest.AttrImportStateIdFunc(resourceName, names.AttrName),
				ImportStateVerify: true,
			},
		},
	})
}

func testAccCheckKeyValueStoreDestroy(ctx context.Context, t *testing.T) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.ProviderMeta(ctx, t).CloudFrontClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_cloudfront_key_value_store" {
				continue
			}

			_, err := tfcloudfront.FindKeyValueStoreByName(ctx, conn, rs.Primary.Attributes[names.AttrName])

			if retry.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("CloudFront Key Value Store %q still exists", rs.Primary.Attributes[names.AttrName])
		}

		return nil
	}
}

func testAccCheckKeyValueStoreExists(ctx context.Context, t *testing.T, n string, v *awstypes.KeyValueStore) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.ProviderMeta(ctx, t).CloudFrontClient(ctx)

		output, err := tfcloudfront.FindKeyValueStoreByName(ctx, conn, rs.Primary.Attributes[names.AttrName])

		if err != nil {
			return err
		}

		*v = *output.KeyValueStore

		return nil
	}
}

func testAccKeyValueStoreConfig_basic(rName string) string {
	return fmt.Sprintf(`
resource "aws_cloudfront_key_value_store" "test" {
  name = %[1]q
}
`, rName)
}

func testAccKeyValueStoreConfig_comment(rName string, comment string) string {
	return fmt.Sprintf(`
resource "aws_cloudfront_key_value_store" "test" {
  name    = %[1]q
  comment = %[2]q
}
`, rName, comment)
}
