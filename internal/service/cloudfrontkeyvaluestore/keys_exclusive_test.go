// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package cloudfrontkeyvaluestore_test

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/cloudfrontkeyvaluestore"
	"github.com/aws/aws-sdk-go-v2/service/cloudfrontkeyvaluestore/types"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	tfcloudfront "github.com/hashicorp/terraform-provider-aws/internal/service/cloudfront"
	tfcloudfrontkeyvaluestore "github.com/hashicorp/terraform-provider-aws/internal/service/cloudfrontkeyvaluestore"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccCloudFrontKeyValueStoreKeysExclusive_basic(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	var keys []string
	var values []string

	// Test with a large number of key value pairs to ensure batching is working correctly
	for i := 1; i < 170; i++ {
		keys = append(keys, sdkacctest.RandomWithPrefix(acctest.ResourcePrefix))
		values = append(values, sdkacctest.RandomWithPrefix(acctest.ResourcePrefix))
	}
	resourceName := "aws_cloudfrontkeyvaluestore_keys_exclusive.test"
	kvsResourceName := "aws_cloudfront_key_value_store.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.CloudFront)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.CloudFront),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckKeysExclusiveDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccKeysExclusiveConfig_basic(keys, values, rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckKeysExclusiveExists(ctx, resourceName),
					resource.TestCheckResourceAttrPair(resourceName, "key_value_store_arn", kvsResourceName, names.AttrARN),
					resource.TestCheckResourceAttrSet(resourceName, "total_size_in_bytes"),
					testCheckMultipleKeyValuePairs(keys, values, resourceName),
				),
			},
			{
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateIdFunc:                    acctest.AttrImportStateIdFunc(resourceName, "key_value_store_arn"),
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: "key_value_store_arn",
			},
		},
	})
}

func TestAccCloudFrontKeyValueStoreKeysExclusive_disappears_KeyValueStore(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	key := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	value := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_cloudfrontkeyvaluestore_keys_exclusive.test"
	kvsResourceName := "aws_cloudfront_key_value_store.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.CloudFront)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.CloudFront),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckKeysExclusiveDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccKeysExclusiveConfig_basic([]string{key}, []string{value}, rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckKeysExclusiveExists(ctx, resourceName),
					acctest.CheckFrameworkResourceDisappears(ctx, acctest.Provider, tfcloudfront.ResourceKeyValueStore, kvsResourceName),
				),
				ExpectNonEmptyPlan: true,
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PostApplyPostRefresh: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(kvsResourceName, plancheck.ResourceActionCreate),
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
					},
				},
			},
		},
	})
}

// A key value pair added out of band should be removed
func TestAccCloudFrontKeyValueStoreKeysExclusive_outOfBandAddition(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	key := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	value := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_cloudfrontkeyvaluestore_keys_exclusive.test"

	// add an additional random key out of band
	putKeys := []types.PutKeyRequestListItem{
		{
			Key:   aws.String(sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)),
			Value: aws.String(sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)),
		},
	}

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.CloudFront)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.CloudFront),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckKeysExclusiveDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccKeysExclusiveConfig_basic([]string{key}, []string{value}, rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckKeysExclusiveExists(ctx, resourceName),
					testAccCheckKeyValueStoreKeysExclusiveUpdate(ctx, resourceName, []types.DeleteKeyRequestListItem{}, putKeys),
				),
				ExpectNonEmptyPlan: true,
			},
			{
				Config: testAccKeysExclusiveConfig_basic([]string{key}, []string{value}, rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckKeysExclusiveExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "resource_key_value_pair.#", "1"),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionUpdate),
					},
				},
			},
		},
	})
}

// A key value pair removed out of band should be re-created
func TestAccCloudFrontKeyValueStoreKeysExclusive_outOfBandRemoval(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	key := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	value := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_cloudfrontkeyvaluestore_keys_exclusive.test"

	// remove the key created in our test
	deleteKeys := []types.DeleteKeyRequestListItem{
		{
			Key: aws.String(key),
		},
	}

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.CloudFront)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.CloudFront),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckKeysExclusiveDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccKeysExclusiveConfig_basic([]string{key}, []string{value}, rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckKeysExclusiveExists(ctx, resourceName),
					testAccCheckKeyValueStoreKeysExclusiveUpdate(ctx, resourceName, deleteKeys, []types.PutKeyRequestListItem{}),
				),
				ExpectNonEmptyPlan: true,
			},
			{
				Config: testAccKeysExclusiveConfig_basic([]string{key}, []string{value}, rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckKeysExclusiveExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "resource_key_value_pair.#", "1"),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionUpdate),
					},
				},
			},
		},
	})
}

func TestAccCloudFrontKeyValueStoreKeysExclusive_value(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	var keys []string
	var values []string
	for i := 1; i < 6; i++ {
		keys = append(keys, sdkacctest.RandomWithPrefix(acctest.ResourcePrefix))
		values = append(values, sdkacctest.RandomWithPrefix(acctest.ResourcePrefix))
	}

	resourceName := "aws_cloudfrontkeyvaluestore_keys_exclusive.test"

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
				Config: testAccKeysExclusiveConfig_basic([]string{keys[0], keys[1]}, []string{values[0], values[0]}, rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckKeysExclusiveExists(ctx, resourceName),
					testCheckMultipleKeyValuePairs([]string{keys[0], keys[1]}, []string{values[0], values[0]}, resourceName),
				),
			},
			{
				Config: testAccKeysExclusiveConfig_basic([]string{keys[0], keys[2]}, []string{values[0], values[2]}, rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckKeysExclusiveExists(ctx, resourceName),
					testCheckMultipleKeyValuePairs([]string{keys[0], keys[2]}, []string{values[0], values[2]}, resourceName),
				),
			},
		},
	})
}

func TestAccCloudFrontKeyValueStoreKeysExclusive_empty(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resourceName := "aws_cloudfrontkeyvaluestore_keys_exclusive.test"
	keyResourceName := "aws_cloudfrontkeyvaluestore_key.test"

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
				Config: testAccKeysExclusiveConfig_empty(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckKeysExclusiveExists(ctx, resourceName)),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						resourceName,
						tfjsonpath.New("resource_key_value_pair"),
						knownvalue.SetExact([]knownvalue.Check{}),
					),
				},
				// The _exclusive resource will remove the key value pair created by the _key resource,
				// resulting in a non-empty plan.
				ExpectNonEmptyPlan: true,
			},
			{
				Config: testAccKeysExclusiveConfig_empty(rName),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionNoop),
						plancheck.ExpectResourceAction(keyResourceName, plancheck.ResourceActionCreate),
					},
				},
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccCloudFrontKeyValueStoreKeysExclusive_maxBatchSize(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	maxBatchSize := sdkacctest.RandIntRange(35, 49)
	var keys []string
	var values []string
	// Test with a large number of key value pairs to ensure batching is working correctly
	for i := 1; i < 170; i++ {
		keys = append(keys, sdkacctest.RandomWithPrefix(acctest.ResourcePrefix))
		values = append(values, sdkacctest.RandomWithPrefix(acctest.ResourcePrefix))
	}
	resourceName := "aws_cloudfrontkeyvaluestore_keys_exclusive.test"
	kvsResourceName := "aws_cloudfront_key_value_store.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.CloudFront)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.CloudFront),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckKeysExclusiveDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccKeysExclusiveConfig_maxBatchSize(keys, values, rName, maxBatchSize),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckKeysExclusiveExists(ctx, resourceName),
					resource.TestCheckResourceAttrPair(resourceName, "key_value_store_arn", kvsResourceName, names.AttrARN),
					resource.TestCheckResourceAttrSet(resourceName, "total_size_in_bytes"),
					testCheckMultipleKeyValuePairs(keys, values, resourceName),
				),
			},
			{
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateIdFunc:                    acctest.AttrImportStateIdFunc(resourceName, "key_value_store_arn"),
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: "key_value_store_arn",
				ImportStateVerifyIgnore:              []string{"max_batch_size"},
			},
		},
	})
}

func testAccCheckKeysExclusiveDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).CloudFrontKeyValueStoreClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_cloudfrontkeyvaluestore_keys_exclusive" {
				continue
			}

			_, err := tfcloudfrontkeyvaluestore.FindKeyValueStoreByARN(ctx, conn, rs.Primary.Attributes["key_value_store_arn"])

			if tfresource.NotFound(err) {
				return nil
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("CloudFront KeyValueStore Key %s still exists", rs.Primary.Attributes["key_value_store_arn"])
		}

		return nil
	}
}

func testAccCheckKeysExclusiveExists(ctx context.Context, n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).CloudFrontKeyValueStoreClient(ctx)

		_, out, err := tfcloudfrontkeyvaluestore.FindResourceKeyValuePairsForKeyValueStore(ctx, conn, rs.Primary.Attributes["key_value_store_arn"])

		if err != nil {
			return create.Error(names.CloudFrontKeyValueStore, create.ErrActionCheckingExistence, tfcloudfrontkeyvaluestore.ResNameKeysExclusive, rs.Primary.Attributes["key_value_store_arn"], err)
		}

		kvPairCount := rs.Primary.Attributes["resource_key_value_pair.#"]
		if kvPairCount != strconv.Itoa(len(out)) {
			return create.Error(names.CloudFrontKeyValueStore, create.ErrActionCheckingExistence, tfcloudfrontkeyvaluestore.ResNameKeysExclusive, rs.Primary.Attributes["key_value_store_arn"], errors.New("unexpected resource_key_value_pair count"))
		}

		return nil
	}
}

func testAccCheckKeyValueStoreKeysExclusiveUpdate(ctx context.Context, n string, deletes []types.DeleteKeyRequestListItem, puts []types.PutKeyRequestListItem) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).CloudFrontKeyValueStoreClient(ctx)

		resp, err := tfcloudfrontkeyvaluestore.FindKeyValueStoreByARN(ctx, conn, rs.Primary.Attributes["key_value_store_arn"])

		if err != nil {
			return fmt.Errorf("error finding Cloudfront KeyValueStore in out of band test")
		}

		input := cloudfrontkeyvaluestore.UpdateKeysInput{
			KvsARN:  resp.KvsARN,
			IfMatch: resp.ETag,
			Deletes: deletes,
			Puts:    puts,
		}

		_, err = conn.UpdateKeys(ctx, &input)

		if err != nil {
			return fmt.Errorf("Error updating CloudFront KeyValueStore %s Key for out of band tests", rs.Primary.Attributes["key_value_store_arn"])
		}

		return nil
	}
}

func testCheckMultipleKeyValuePairs(keys, values []string, resourceName string) resource.TestCheckFunc {
	for i := range keys {
		return resource.TestCheckTypeSetElemNestedAttrs(resourceName, "resource_key_value_pair.*", map[string]string{
			names.AttrKey:   keys[i],
			names.AttrValue: values[i],
		})
	}

	return nil
}

func testAccKeysExclusiveConfig_basic(keys, values []string, rName string) string {
	keysJson, _ := json.Marshal(keys)
	keysString := string(keysJson)
	valuesJson, _ := json.Marshal(values)
	valuesString := string(valuesJson)
	return fmt.Sprintf(`
locals {
  key_list      = %[1]s
  value_list    = %[2]s
  key_value_set = { for i, v in local.key_list : local.key_list[i] => local.value_list[i] }
}

resource "aws_cloudfront_key_value_store" "test" {
  name = %[3]q
}

resource "aws_cloudfrontkeyvaluestore_keys_exclusive" "test" {
  key_value_store_arn = aws_cloudfront_key_value_store.test.arn

  dynamic "resource_key_value_pair" {
    for_each = local.key_value_set
    content {
      key   = resource_key_value_pair.key
      value = resource_key_value_pair.value

    }
  }
}
`, keysString, valuesString, rName)
}

func testAccKeysExclusiveConfig_empty(rName string) string {
	return fmt.Sprintf(`
resource "aws_cloudfront_key_value_store" "test" {
  name = %[1]q
}

resource "aws_cloudfrontkeyvaluestore_key" "test" {
  key                 = %[1]q
  key_value_store_arn = aws_cloudfront_key_value_store.test.arn
  value               = %[1]q
}

resource "aws_cloudfrontkeyvaluestore_keys_exclusive" "test" {
  key_value_store_arn = aws_cloudfront_key_value_store.test.arn
  depends_on = [
    aws_cloudfrontkeyvaluestore_key.test
  ]
}
`, rName)
}

func testAccKeysExclusiveConfig_maxBatchSize(keys, values []string, rName string, maxBatchSize int) string {
	keysJson, _ := json.Marshal(keys)
	keysString := string(keysJson)
	valuesJson, _ := json.Marshal(values)
	valuesString := string(valuesJson)
	return fmt.Sprintf(`
locals {
  key_list      = %[1]s
  value_list    = %[2]s
  key_value_set = { for i, v in local.key_list : local.key_list[i] => local.value_list[i] }
}

resource "aws_cloudfront_key_value_store" "test" {
  name = %[3]q
}

resource "aws_cloudfrontkeyvaluestore_keys_exclusive" "test" {
  key_value_store_arn = aws_cloudfront_key_value_store.test.arn
  max_batch_size      = %[4]d
  dynamic "resource_key_value_pair" {
    for_each = local.key_value_set
    content {
      key   = resource_key_value_pair.key
      value = resource_key_value_pair.value

    }
  }
}
`, keysString, valuesString, rName, maxBatchSize)
}
