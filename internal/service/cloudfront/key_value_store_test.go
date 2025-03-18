// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package cloudfront_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
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

func TestAccCloudFrontKeyValueStore_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var keyvaluestore awstypes.KeyValueStore
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
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckKeyValueStoreExists(ctx, resourceName, &keyvaluestore),
					func(s *terraform.State) error {
						return acctest.CheckResourceAttrGlobalARN(ctx, resourceName, names.AttrARN, "cloudfront", "key-value-store/"+aws.ToString(keyvaluestore.Id))(s)
					},
					resource.TestCheckNoResourceAttr(resourceName, names.AttrComment),
					resource.TestCheckResourceAttrSet(resourceName, "etag"),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrID, resourceName, names.AttrName),
					resource.TestCheckResourceAttrSet(resourceName, "last_modified_time"),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
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

func TestAccCloudFrontKeyValueStore_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	var keyvaluestore awstypes.KeyValueStore
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

func TestAccCloudFrontKeyValueStore_comment(t *testing.T) {
	ctx := acctest.Context(t)
	var keyvaluestore awstypes.KeyValueStore
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
				Config: testAccKeyValueStoreConfig_comment(rName, comment1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckKeyValueStoreExists(ctx, resourceName, &keyvaluestore),
					resource.TestCheckResourceAttr(resourceName, names.AttrComment, comment1),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccKeyValueStoreConfig_comment(rName, comment2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckKeyValueStoreExists(ctx, resourceName, &keyvaluestore),
					resource.TestCheckResourceAttr(resourceName, names.AttrComment, comment2),
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

			_, err := tfcloudfront.FindKeyValueStoreByName(ctx, conn, rs.Primary.ID)

			if tfresource.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("CloudFront Key Value Store %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccCheckKeyValueStoreExists(ctx context.Context, n string, v *awstypes.KeyValueStore) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).CloudFrontClient(ctx)

		output, err := tfcloudfront.FindKeyValueStoreByName(ctx, conn, rs.Primary.ID)

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
