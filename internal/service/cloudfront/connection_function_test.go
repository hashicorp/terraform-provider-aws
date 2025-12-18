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
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	tfcloudfront "github.com/hashicorp/terraform-provider-aws/internal/service/cloudfront"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccCloudFrontConnectionFunction_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var connectionfunction cloudfront.DescribeConnectionFunctionOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_cloudfront_connection_function.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.CloudFrontEndpointID)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.CloudFrontServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckConnectionFunctionDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccConnectionFunctionConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckConnectionFunctionExists(ctx, resourceName, &connectionfunction),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					acctest.CheckResourceAttrGlobalARNFormat(ctx, resourceName, names.AttrARN, "cloudfront", "connection-function/{id}"),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrID),
					resource.TestCheckResourceAttrSet(resourceName, "etag"),
					resource.TestCheckResourceAttr(resourceName, "runtime", string(awstypes.FunctionRuntimeCloudfrontJs20)),
					resource.TestCheckResourceAttr(resourceName, "comment", "Test connection function"),
					resource.TestCheckResourceAttrSet(resourceName, "created_time"),
					resource.TestCheckResourceAttrSet(resourceName, "last_modified_time"),
					resource.TestCheckResourceAttr(resourceName, names.AttrStatus, "UNPUBLISHED"),
					resource.TestCheckResourceAttr(resourceName, "live_stage_etag", ""),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"location",
					"last_modified_time",
					"pubish",
				},
			},
		},
	})
}

func TestAccCloudFrontConnectionFunction_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	var connectionfunction cloudfront.DescribeConnectionFunctionOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_cloudfront_connection_function.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.CloudFrontEndpointID)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.CloudFrontServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckConnectionFunctionDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccConnectionFunctionConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckConnectionFunctionExists(ctx, resourceName, &connectionfunction),
					acctest.CheckFrameworkResourceDisappears(ctx, acctest.Provider, tfcloudfront.ResourceConnectionFunction, resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccCloudFrontConnectionFunction_update(t *testing.T) {
	ctx := acctest.Context(t)
	var connectionfunction1, connectionfunction2 cloudfront.DescribeConnectionFunctionOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_cloudfront_connection_function.test"
	kvsResourceName := "aws_cloudfront_key_value_store.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.CloudFrontEndpointID)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.CloudFrontServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckConnectionFunctionDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccConnectionFunctionConfig_updateInitial(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckConnectionFunctionExists(ctx, resourceName, &connectionfunction1),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, "runtime", string(awstypes.FunctionRuntimeCloudfrontJs20)),
					resource.TestCheckResourceAttr(resourceName, "comment", "Initial test connection function with runtime 2.0"),
					resource.TestCheckResourceAttr(resourceName, "key_value_store_associations.#", "0"),
					resource.TestCheckResourceAttr(resourceName, names.AttrStatus, "UNPUBLISHED"),
					resource.TestCheckResourceAttr(resourceName, "live_stage_etag", ""),
					resource.TestCheckResourceAttrSet(resourceName, "etag"),
				),
			},
			{
				Config: testAccConnectionFunctionConfig_updateComplete(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckConnectionFunctionExists(ctx, resourceName, &connectionfunction2),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, "runtime", string(awstypes.FunctionRuntimeCloudfrontJs20)),
					resource.TestCheckResourceAttr(resourceName, "comment", "Updated test connection function with all attributes"),
					resource.TestCheckResourceAttr(resourceName, "key_value_store_associations.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "key_value_store_associations.0.items.#", "1"),
					resource.TestCheckResourceAttrPair(resourceName, "key_value_store_associations.0.items.0.key_value_store_arn", kvsResourceName, names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, names.AttrStatus, "UNPUBLISHED"),
					resource.TestCheckResourceAttr(resourceName, "live_stage_etag", ""),
					resource.TestCheckResourceAttrSet(resourceName, "etag"),
					testAccCheckConnectionFunctionNotRecreated(&connectionfunction1, &connectionfunction2),
					testAccCheckConnectionFunctionEtagChanged(&connectionfunction1, &connectionfunction2),
				),
			},
		},
	})
}

func TestAccCloudFrontConnectionFunction_publish1(t *testing.T) {
	ctx := acctest.Context(t)
	var connectionfunction1, connectionfunction2 cloudfront.DescribeConnectionFunctionOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_cloudfront_connection_function.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.CloudFrontEndpointID)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.CloudFrontServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckConnectionFunctionDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccConnectionFunctionConfig_publish(rName, false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckConnectionFunctionExists(ctx, resourceName, &connectionfunction1),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					acctest.CheckResourceAttrGlobalARNFormat(ctx, resourceName, names.AttrARN, "cloudfront", "connection-function/{id}"),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrID),
					resource.TestCheckResourceAttrSet(resourceName, "etag"),
					resource.TestCheckResourceAttr(resourceName, "runtime", string(awstypes.FunctionRuntimeCloudfrontJs20)),
					resource.TestCheckResourceAttr(resourceName, "comment", "Test connection function with publish"),
					resource.TestCheckResourceAttrSet(resourceName, "created_time"),
					resource.TestCheckResourceAttrSet(resourceName, "last_modified_time"),
					resource.TestCheckResourceAttr(resourceName, names.AttrStatus, "UNPUBLISHED"),
					resource.TestCheckResourceAttr(resourceName, "live_stage_etag", ""),
					resource.TestCheckResourceAttr(resourceName, "publish", "false"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"location",
					"publish", // publish is not returned by the API, so ignore during import
					"last_modified_time",
				},
			},
			{
				Config: testAccConnectionFunctionConfig_publish(rName, true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckConnectionFunctionExists(ctx, resourceName, &connectionfunction2),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					acctest.CheckResourceAttrGlobalARNFormat(ctx, resourceName, names.AttrARN, "cloudfront", "connection-function/{id}"),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrID),
					resource.TestCheckResourceAttrSet(resourceName, "etag"),
					resource.TestCheckResourceAttr(resourceName, "runtime", string(awstypes.FunctionRuntimeCloudfrontJs20)),
					resource.TestCheckResourceAttr(resourceName, "comment", "Test connection function with publish"),
					resource.TestCheckResourceAttrSet(resourceName, "created_time"),
					resource.TestCheckResourceAttrSet(resourceName, "last_modified_time"),
					resource.TestCheckResourceAttr(resourceName, names.AttrStatus, "UNASSOCIATED"),
					resource.TestCheckResourceAttrSet(resourceName, "live_stage_etag"),
					resource.TestCheckResourceAttr(resourceName, "publish", "true"),
				),
			},
		},
	})
}

func TestAccCloudFrontConnectionFunction_allAttributesWithKeyValueStore(t *testing.T) {
	ctx := acctest.Context(t)
	var connectionfunction1, connectionfunction2 cloudfront.DescribeConnectionFunctionOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_cloudfront_connection_function.test"
	kvsResourceName := "aws_cloudfront_key_value_store.test"
	kvsResourceName2 := "aws_cloudfront_key_value_store.test2"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.CloudFrontEndpointID)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.CloudFrontServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckConnectionFunctionDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccConnectionFunctionConfig_allAttributesWithKeyValueStore(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckConnectionFunctionExists(ctx, resourceName, &connectionfunction1),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					acctest.CheckResourceAttrGlobalARNFormat(ctx, resourceName, names.AttrARN, "cloudfront", "connection-function/{id}"),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrID),
					resource.TestCheckResourceAttrSet(resourceName, "etag"),
					resource.TestCheckResourceAttr(resourceName, "runtime", string(awstypes.FunctionRuntimeCloudfrontJs20)),
					resource.TestCheckResourceAttr(resourceName, "comment", "Initial test connection function with KVS"),
					resource.TestCheckResourceAttrSet(resourceName, "created_time"),
					resource.TestCheckResourceAttrSet(resourceName, "last_modified_time"),
					resource.TestCheckResourceAttr(resourceName, names.AttrStatus, "UNPUBLISHED"),
					resource.TestCheckResourceAttr(resourceName, "live_stage_etag", ""),
					resource.TestCheckResourceAttr(resourceName, "key_value_store_associations.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "key_value_store_associations.0.items.#", "1"),
					resource.TestCheckResourceAttrPair(resourceName, "key_value_store_associations.0.items.0.key_value_store_arn", kvsResourceName, names.AttrARN),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"location",
					"publish", // publish is not returned by the API, so ignore during import
					"last_modified_time",
				},
			},
			{
				Config: testAccConnectionFunctionConfig_allAttributesWithKeyValueStoreUpdated(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckConnectionFunctionExists(ctx, resourceName, &connectionfunction2),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, "runtime", string(awstypes.FunctionRuntimeCloudfrontJs20)),
					resource.TestCheckResourceAttr(resourceName, "comment", "Updated test connection function with two KVS"),
					resource.TestCheckResourceAttr(resourceName, "key_value_store_associations.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "key_value_store_associations.0.items.#", "1"),
					resource.TestCheckResourceAttrPair(resourceName, "key_value_store_associations.0.items.0.key_value_store_arn", kvsResourceName2, names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, names.AttrStatus, "UNPUBLISHED"),
					resource.TestCheckResourceAttr(resourceName, "live_stage_etag", ""),
					testAccCheckConnectionFunctionNotRecreated(&connectionfunction1, &connectionfunction2),
				),
			},
		},
	})
}

func testAccCheckConnectionFunctionDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).CloudFrontClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_cloudfront_connection_function" {
				continue
			}

			_, err := tfcloudfront.FindConnectionFunctionByTwoPartKey(ctx, conn, rs.Primary.ID, awstypes.FunctionStageDevelopment)
			if retry.NotFound(err) {
				continue
			}
			if err != nil {
				return create.Error(names.CloudFront, create.ErrActionCheckingDestroyed, tfcloudfront.ResNameConnectionFunction, rs.Primary.ID, err)
			}

			return create.Error(names.CloudFront, create.ErrActionCheckingDestroyed, tfcloudfront.ResNameConnectionFunction, rs.Primary.ID, errors.New("not destroyed"))
		}

		return nil
	}
}

func testAccCheckConnectionFunctionExists(ctx context.Context, name string, connectionfunction *cloudfront.DescribeConnectionFunctionOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return create.Error(names.CloudFront, create.ErrActionCheckingExistence, tfcloudfront.ResNameConnectionFunction, name, errors.New("not found"))
		}

		if rs.Primary.ID == "" {
			return create.Error(names.CloudFront, create.ErrActionCheckingExistence, tfcloudfront.ResNameConnectionFunction, name, errors.New("not set"))
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).CloudFrontClient(ctx)

		resp, err := tfcloudfront.FindConnectionFunctionByTwoPartKey(ctx, conn, rs.Primary.ID, awstypes.FunctionStageDevelopment)
		if err != nil {
			return create.Error(names.CloudFront, create.ErrActionCheckingExistence, tfcloudfront.ResNameConnectionFunction, rs.Primary.ID, err)
		}

		*connectionfunction = *resp

		return nil
	}
}

func testAccCheckConnectionFunctionNotRecreated(before, after *cloudfront.DescribeConnectionFunctionOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		beforeId := aws.ToString(before.ConnectionFunctionSummary.Id)
		afterId := aws.ToString(after.ConnectionFunctionSummary.Id)

		if beforeId != afterId {
			return create.Error(names.CloudFront, create.ErrActionCheckingNotRecreated, tfcloudfront.ResNameConnectionFunction, beforeId, fmt.Errorf("recreated: before ID %s, after ID %s", beforeId, afterId))
		}

		return nil
	}
}

func testAccCheckConnectionFunctionEtagChanged(before, after *cloudfront.DescribeConnectionFunctionOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		beforeEtag := aws.ToString(before.ETag)
		afterEtag := aws.ToString(after.ETag)

		if beforeEtag == "" {
			return create.Error(names.CloudFront, create.ErrActionChecking, tfcloudfront.ResNameConnectionFunction, "etag", fmt.Errorf("before etag is empty"))
		}

		if afterEtag == "" {
			return create.Error(names.CloudFront, create.ErrActionChecking, tfcloudfront.ResNameConnectionFunction, "etag", fmt.Errorf("after etag is empty"))
		}

		if beforeEtag == afterEtag {
			return create.Error(names.CloudFront, create.ErrActionChecking, tfcloudfront.ResNameConnectionFunction, "etag", fmt.Errorf("etag did not change: before %s, after %s", beforeEtag, afterEtag))
		}

		return nil
	}
}

func testAccConnectionFunctionConfig_basic(rName string) string {
	return fmt.Sprintf(`
resource "aws_cloudfront_connection_function" "test" {
  name    = %[1]q
  runtime = "cloudfront-js-2.0"
  comment = "Test connection function"
  code    = "function handler(event) { return event.request; }"
}
`, rName)
}

func testAccConnectionFunctionConfig_updateInitial(rName string) string {
	return fmt.Sprintf(`
resource "aws_cloudfront_connection_function" "test" {
  name    = %[1]q
  runtime = "cloudfront-js-2.0"
  comment = "Initial test connection function with runtime 2.0"
  code    = <<-EOT
function handler(event) {
  console.log("Initial function execution with runtime 2.0");
  return event.request;
}
EOT
}
`, rName)
}

func testAccConnectionFunctionConfig_updateComplete(rName string) string {
	return fmt.Sprintf(`
resource "aws_cloudfront_key_value_store" "test" {
  name    = %[1]q
  comment = "Test key value store for update test"
}

resource "aws_cloudfront_connection_function" "test" {
  name    = %[1]q
  runtime = "cloudfront-js-2.0"
  comment = "Updated test connection function with all attributes"
  code    = <<-EOT
function handler(event) {
  console.log("Updated function execution with KVS support");
  var kv = event.context.kvs;
  var testKey = "update-test-key";
  var value = kv.get(testKey);
  
  if (value) {
    console.log("Retrieved value from KVS: " + value);
    event.request.headers["x-kvs-value"] = {value: value};
  }
  
  event.request.headers["x-function-version"] = {value: "updated"};
  event.request.headers["x-timestamp"] = {value: new Date().toISOString()};
  
  return event.request;
}
EOT

  key_value_store_associations {
    items {
      key_value_store_arn = aws_cloudfront_key_value_store.test.arn
    }
  }
}
`, rName)
}

func testAccConnectionFunctionConfig_allAttributesWithKeyValueStore(rName string) string {
	return fmt.Sprintf(`
resource "aws_cloudfront_key_value_store" "test" {
  name    = %[1]q
  comment = "Test key value store for connection function"
}

resource "aws_cloudfront_connection_function" "test" {
  name    = %[1]q
  runtime = "cloudfront-js-2.0"
  comment = "Initial test connection function with KVS"
  code    = <<-EOT
function handler(event) {
  var kv = event.context.kvs;
  var key = "test-key";
  var value = kv.get(key);
  console.log("Retrieved value: " + value);
  return event.request;
}
EOT

  key_value_store_associations {
    items {
      key_value_store_arn = aws_cloudfront_key_value_store.test.arn
    }
  }
}
`, rName)
}

func testAccConnectionFunctionConfig_publish(rName string, publish bool) string {
	return fmt.Sprintf(`
resource "aws_cloudfront_connection_function" "test" {
  name    = %[1]q
  runtime = "cloudfront-js-2.0"
  comment = "Test connection function with publish"
  code    = "function handler(event) { return event.request; }"
  publish = %[2]t
}
`, rName, publish)
}

func TestAccCloudFrontConnectionFunction_publishTransition(t *testing.T) {
	ctx := acctest.Context(t)
	var connectionfunction1, connectionfunction2 cloudfront.DescribeConnectionFunctionOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_cloudfront_connection_function.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.CloudFrontEndpointID)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.CloudFrontServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckConnectionFunctionDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccConnectionFunctionConfig_publish(rName, false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckConnectionFunctionExists(ctx, resourceName, &connectionfunction1),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, names.AttrStatus, "UNPUBLISHED"),
					resource.TestCheckResourceAttr(resourceName, "live_stage_etag", ""),
					resource.TestCheckResourceAttr(resourceName, "publish", "false"),
				),
			},
			{
				Config: testAccConnectionFunctionConfig_publish(rName, true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckConnectionFunctionExists(ctx, resourceName, &connectionfunction2),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, names.AttrStatus, "UNASSOCIATED"),
					resource.TestCheckResourceAttrSet(resourceName, "live_stage_etag"),
					resource.TestCheckResourceAttr(resourceName, "publish", "true"),
					testAccCheckConnectionFunctionNotRecreated(&connectionfunction1, &connectionfunction2),
				),
			},
		},
	})
}

func TestAccCloudFrontConnectionFunction_waiter(t *testing.T) {
	ctx := acctest.Context(t)
	var connectionfunction cloudfront.DescribeConnectionFunctionOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_cloudfront_connection_function.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.CloudFrontEndpointID)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.CloudFrontServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckConnectionFunctionDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccConnectionFunctionConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckConnectionFunctionExists(ctx, resourceName, &connectionfunction),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, names.AttrStatus, "UNPUBLISHED"),
					resource.TestCheckResourceAttr(resourceName, "live_stage_etag", ""),
					// Verify that the waiter worked and we have a stable status
					func(s *terraform.State) error {
						// Additional check to ensure status is one of the expected stable states
						status := connectionfunction.ConnectionFunctionSummary.Status
						if status == nil {
							return fmt.Errorf("status is nil")
						}
						statusStr := aws.ToString(status)
						if statusStr != "UNPUBLISHED" && statusStr != "UNASSOCIATED" && statusStr != "DEPLOYED" {
							return fmt.Errorf("unexpected status: %s", statusStr)
						}
						return nil
					},
				),
			},
		},
	})
}

func testAccConnectionFunctionConfig_allAttributesWithKeyValueStoreUpdated(rName string) string {
	return fmt.Sprintf(`
resource "aws_cloudfront_key_value_store" "test" {
  name    = %[1]q
  comment = "Test key value store for connection function"
}

resource "aws_cloudfront_key_value_store" "test2" {
  name    = "%[1]s-new"
  comment = "New test key value store for connection function"
}

resource "aws_cloudfront_connection_function" "test" {
  name    = %[1]q
  runtime = "cloudfront-js-2.0"
  comment = "Updated test connection function with two KVS"
  code    = <<-EOT
function handler(event) {
  var kv = event.context.kvs;
  var originalKey = "test-key";
  var newKey = "new-key";
  var originalValue = kv.get(originalKey);
  var newValue = kv.get(newKey);
  console.log("Original value: " + originalValue);
  console.log("New value: " + newValue);
  event.request.headers["x-original-header"] = {value: originalValue};
  event.request.headers["x-new-header"] = {value: newValue};
  return event.request;
}
EOT

  key_value_store_associations {
    items {
      key_value_store_arn = aws_cloudfront_key_value_store.test2.arn
    }
  }
}
`, rName)
}
