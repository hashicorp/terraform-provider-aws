// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package cloudfrontkeyvaluestore_test

import (
	"context"
	"errors"
	"fmt"
	"testing"

	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	tfcloudfrontkeyvaluestore "github.com/hashicorp/terraform-provider-aws/internal/service/cloudfrontkeyvaluestore"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccCloudFrontKeyValueStoreKey_basic(t *testing.T) {
	ctx := acctest.Context(t)

	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	value := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_cloudfrontkeyvaluestore_key.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.CloudFront)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.CloudFront),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckKeyDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccKeyConfig_basic(rName, value),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckKeyExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "key", rName),
					resource.TestCheckResourceAttrSet(resourceName, "key_value_store_arn"),
					resource.TestCheckResourceAttrSet(resourceName, "id"),
					resource.TestCheckResourceAttrSet(resourceName, "total_size_in_bytes"),
					resource.TestCheckResourceAttr(resourceName, "value", value),
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

func TestAccCloudFrontKeyValueStoreKey_value(t *testing.T) {
	ctx := acctest.Context(t)

	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	value1 := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	value2 := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resourceName := "aws_cloudfrontkeyvaluestore_key.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.CloudFront)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.CloudFront),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckKeyDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccKeyConfig_basic(rName, value1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckKeyExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "key", rName),
					resource.TestCheckResourceAttrSet(resourceName, "key_value_store_arn"),
					resource.TestCheckResourceAttrSet(resourceName, "id"),
					resource.TestCheckResourceAttrSet(resourceName, "total_size_in_bytes"),
					resource.TestCheckResourceAttr(resourceName, "value", value1),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccKeyConfig_basic(rName, value2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckKeyExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "key", rName),
					resource.TestCheckResourceAttrSet(resourceName, "key_value_store_arn"),
					resource.TestCheckResourceAttrSet(resourceName, "id"),
					resource.TestCheckResourceAttrSet(resourceName, "total_size_in_bytes"),
					resource.TestCheckResourceAttr(resourceName, "value", value2),
				),
			},
		},
	})
}

func TestAccCloudFrontKeyValueStoreKey_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	value := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_cloudfrontkeyvaluestore_key.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.CloudFront)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.CloudFront),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckKeyDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccKeyConfig_basic(rName, value),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckKeyExists(ctx, resourceName),
					acctest.CheckFrameworkResourceDisappears(ctx, acctest.Provider, tfcloudfrontkeyvaluestore.ResourceKey, resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckKeyDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).CloudFrontKeyValueStoreClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_cloudfrontkeyvaluestore_key" {
				continue
			}

			_, _, err := tfcloudfrontkeyvaluestore.FindKeyByID(ctx, conn, rs.Primary.ID)
			if tfresource.NotFound(err) {
				return nil
			}

			if err != nil {
				return create.Error(names.CloudFrontKeyValueStore, create.ErrActionCheckingDestroyed, tfcloudfrontkeyvaluestore.ResNameKey, rs.Primary.ID, err)
			}

			return create.Error(names.CloudFrontKeyValueStore, create.ErrActionCheckingDestroyed, tfcloudfrontkeyvaluestore.ResNameKey, rs.Primary.ID, errors.New("not destroyed"))
		}

		return nil
	}
}

func testAccCheckKeyExists(ctx context.Context, name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return create.Error(names.CloudFrontKeyValueStore, create.ErrActionCheckingExistence, tfcloudfrontkeyvaluestore.ResNameKey, name, errors.New("not found"))
		}

		if rs.Primary.ID == "" {
			return create.Error(names.CloudFrontKeyValueStore, create.ErrActionCheckingExistence, tfcloudfrontkeyvaluestore.ResNameKey, name, errors.New("not set"))
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).CloudFrontKeyValueStoreClient(ctx)
		_, _, err := tfcloudfrontkeyvaluestore.FindKeyByID(ctx, conn, rs.Primary.ID)
		if tfresource.NotFound(err) {
			return create.Error(names.CloudFrontKeyValueStore, create.ErrActionCheckingExistence, tfcloudfrontkeyvaluestore.ResNameKey, rs.Primary.ID, errors.New("Resource Not Found"))
		}

		if err != nil {
			return create.Error(names.CloudFrontKeyValueStore, create.ErrActionCheckingExistence, tfcloudfrontkeyvaluestore.ResNameKey, rs.Primary.ID, err)
		}

		return nil
	}
}

func testAccKeyConfig_basic(rName, value string) string {
	return fmt.Sprintf(`
resource "aws_cloudfront_key_value_store" "test" {
  name = %[1]q
}
resource "aws_cloudfrontkeyvaluestore_key" "test" {
  key                 = %[1]q
  key_value_store_arn = aws_cloudfront_key_value_store.test.arn
  value               = %[2]q
}
`, rName, value)
}
