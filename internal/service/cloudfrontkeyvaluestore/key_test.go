// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package cloudfrontkeyvaluestore_test

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"

	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
	"github.com/hashicorp/terraform-plugin-testing/tfversion"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	tfstatecheck "github.com/hashicorp/terraform-provider-aws/internal/acctest/statecheck"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
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
					resource.TestCheckResourceAttr(resourceName, names.AttrKey, rName),
					resource.TestCheckResourceAttrSet(resourceName, "key_value_store_arn"),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrID),
					resource.TestCheckResourceAttrSet(resourceName, "total_size_in_bytes"),
					resource.TestCheckResourceAttr(resourceName, names.AttrValue, value),
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

func TestAccCloudFrontKeyValueStoreKey_Identity_Basic(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	value := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_cloudfrontkeyvaluestore_key.test"

	resource.ParallelTest(t, resource.TestCase{
		TerraformVersionChecks: []tfversion.TerraformVersionCheck{
			tfversion.SkipBelow(tfversion.Version1_12_0),
		},
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.CloudFront)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.CloudFront),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckKeyDestroy(ctx),
		Steps: []resource.TestStep{
			// Step 1: Setup
			{
				Config: testAccKeyConfig_basic(rName, value),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckKeyExists(ctx, resourceName),
				),
				ConfigStateChecks: []statecheck.StateCheck{
					tfstatecheck.ExpectAttributeFormat(resourceName, tfjsonpath.New(names.AttrID), "{key_value_store_arn},{key}"),
					// statecheck.ExpectIdentity(resourceName, map[string]knownvalue.Check{
					// 	names.AttrAccountID:   tfknownvalue.AccountID(),
					// 	"key_value_store_arn": knownvalue.NotNull(),
					// 	names.AttrKey:         knownvalue.NotNull(),
					// }),
					// statecheck.ExpectIdentityValueMatchesState(resourceName, tfjsonpath.New("key_value_store_arn")),
					// statecheck.ExpectIdentityValueMatchesState(resourceName, tfjsonpath.New(names.AttrKey)),
				},
			},

			// Step 2: Import command
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateKind:   resource.ImportCommandWithID,
				ImportStateVerify: true,
			},

			// Step 3: Import block with Import ID
			{
				ResourceName:    resourceName,
				ImportState:     true,
				ImportStateKind: resource.ImportBlockWithID,
				ImportPlanChecks: resource.ImportPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectKnownValue(resourceName, tfjsonpath.New("key_value_store_arn"), knownvalue.NotNull()), // TODO: needs pair
						plancheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrKey), knownvalue.StringExact(rName)),
						plancheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrID), knownvalue.NotNull()), // TODO: needs tfplancheck.ExpectAttributeFormat
					},
				},
			},

			// // Step 4: Import block with Resource Identity
			// {
			// 	ResourceName:    resourceName,
			// 	ImportState:     true,
			// 	ImportStateKind: resource.ImportBlockWithResourceIdentity,
			// 	ImportPlanChecks: resource.ImportPlanChecks{
			// 		PreApply: []plancheck.PlanCheck{
			// 			plancheck.ExpectKnownValue(resourceName, tfjsonpath.New("key_value_store_arn"), knownvalue.NotNull()), // TODO: needs pair
			// 			plancheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrKey), knownvalue.StringExact(rName)),
			// 			plancheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrID), knownvalue.NotNull()), // TODO: needs tfplancheck.ExpectAttributeFormat
			// 		},
			// 	},
			// },
		},
	})
}

// This test is to verify the mutex lock is working correctly to allow serializing multiple keys being changed
func TestAccCloudFrontKeyValueStoreKey_mutex(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	var rNames []string
	for i := 1; i < 6; i++ {
		rNames = append(rNames, sdkacctest.RandomWithPrefix(acctest.ResourcePrefix))
	}
	value := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

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
				Config: testAccKeyConfig_mutex(rNames, rName, value),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("aws_cloudfrontkeyvaluestore_key.test.0", names.AttrKey, rNames[0]),
					resource.TestCheckResourceAttr("aws_cloudfrontkeyvaluestore_key.test.1", names.AttrKey, rNames[1]),
					resource.TestCheckResourceAttr("aws_cloudfrontkeyvaluestore_key.test.2", names.AttrKey, rNames[2]),
					resource.TestCheckResourceAttr("aws_cloudfrontkeyvaluestore_key.test.3", names.AttrKey, rNames[3]),
					resource.TestCheckResourceAttr("aws_cloudfrontkeyvaluestore_key.test.4", names.AttrKey, rNames[4]),
				),
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
					resource.TestCheckResourceAttr(resourceName, names.AttrKey, rName),
					resource.TestCheckResourceAttrSet(resourceName, "key_value_store_arn"),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrID),
					resource.TestCheckResourceAttrSet(resourceName, "total_size_in_bytes"),
					resource.TestCheckResourceAttr(resourceName, names.AttrValue, value1),
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
					resource.TestCheckResourceAttr(resourceName, names.AttrKey, rName),
					resource.TestCheckResourceAttrSet(resourceName, "key_value_store_arn"),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrID),
					resource.TestCheckResourceAttrSet(resourceName, "total_size_in_bytes"),
					resource.TestCheckResourceAttr(resourceName, names.AttrValue, value2),
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

			_, err := tfcloudfrontkeyvaluestore.FindKeyByTwoPartKey(ctx, conn, rs.Primary.Attributes["key_value_store_arn"], rs.Primary.Attributes[names.AttrKey])

			if tfresource.NotFound(err) {
				return nil
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("CloudFront KeyValueStore Key %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccCheckKeyExists(ctx context.Context, n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).CloudFrontKeyValueStoreClient(ctx)

		_, err := tfcloudfrontkeyvaluestore.FindKeyByTwoPartKey(ctx, conn, rs.Primary.Attributes["key_value_store_arn"], rs.Primary.Attributes[names.AttrKey])

		return err
	}
}

func testAccKeyConfig_basic(rName, value string) string {
	return fmt.Sprintf(`
resource "aws_cloudfrontkeyvaluestore_key" "test" {
  key                 = %[1]q
  key_value_store_arn = aws_cloudfront_key_value_store.test.arn
  value               = %[2]q
}

resource "aws_cloudfront_key_value_store" "test" {
  name = %[1]q
}
`, rName, value)
}

func testAccKeyConfig_mutex(rNames []string, rName, value string) string {
	rNameJson, _ := json.Marshal(rNames)
	rNameString := string(rNameJson)
	return fmt.Sprintf(`
resource "aws_cloudfrontkeyvaluestore_key" "test" {
  count               = length(local.key_list)
  key                 = local.key_list[count.index]
  key_value_store_arn = aws_cloudfront_key_value_store.test.arn
  value               = %[3]q
}

resource "aws_cloudfront_key_value_store" "test" {
  name = %[2]q
}

locals {
  key_list = %[1]s
}
`, rNameString, rName, value)
}
