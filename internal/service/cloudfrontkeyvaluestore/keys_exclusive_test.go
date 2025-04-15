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
	tfcloudfrontkeyvaluestore "github.com/hashicorp/terraform-provider-aws/internal/service/cloudfrontkeyvaluestore"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccCloudFrontKeyValueStoreKeysExclusive_basic(t *testing.T) {
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
		CheckDestroy:             testAccCheckKeysExclusiveDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccKeysExclusiveConfig_basic(keys, values, rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckKeysExclusiveExists(ctx, resourceName),
					resource.TestCheckResourceAttrSet(resourceName, "key_value_store_arn"),
					resource.TestCheckResourceAttrSet(resourceName, "total_size_in_bytes"),
					testCheckMultipleKeyValuePairs(keys, values, resourceName),
				),
			},
			{
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateIdFunc:                    testAccKeysExclusiveImportStateIdFunc(resourceName),
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: "key_value_store_arn",
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
					resource.TestCheckResourceAttrSet(resourceName, "key_value_store_arn"),
					resource.TestCheckResourceAttrSet(resourceName, "total_size_in_bytes"),
					testCheckMultipleKeyValuePairs([]string{keys[0], keys[1]}, []string{values[0], values[0]}, resourceName),
				),
			},
			{
				Config: testAccKeysExclusiveConfig_basic([]string{keys[0], keys[2]}, []string{values[0], values[2]}, rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckKeysExclusiveExists(ctx, resourceName),
					resource.TestCheckResourceAttrSet(resourceName, "key_value_store_arn"),
					resource.TestCheckResourceAttrSet(resourceName, "total_size_in_bytes"),
					testCheckMultipleKeyValuePairs([]string{keys[0], keys[2]}, []string{values[0], values[2]}, resourceName),
				),
			},
		},
	})
}

func TestAccCloudFrontKeyValueStoreKeysExclusive_empty(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	var keys []string
	var values []string
	for i := 1; i < 6; i++ {
		keys = append(keys, sdkacctest.RandomWithPrefix(acctest.ResourcePrefix))
		values = append(values, sdkacctest.RandomWithPrefix(acctest.ResourcePrefix))
	}

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

func testAccKeysExclusiveImportStateIdFunc(resourceName string) resource.ImportStateIdFunc {
	return func(s *terraform.State) (string, error) {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return "", fmt.Errorf("Not found: %s", resourceName)
		}

		return rs.Primary.Attributes["key_value_store_arn"], nil
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
  depends_on =[
  aws_cloudfrontkeyvaluestore_key.test
  ]
}
`, rName)
}
