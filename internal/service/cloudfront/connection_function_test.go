// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package cloudfront_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/cloudfront"
	awstypes "github.com/aws/aws-sdk-go-v2/service/cloudfront/types"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	tfknownvalue "github.com/hashicorp/terraform-provider-aws/internal/acctest/knownvalue"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	tfcloudfront "github.com/hashicorp/terraform-provider-aws/internal/service/cloudfront"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccCloudFrontConnectionFunction_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var connectionfunction cloudfront.DescribeConnectionFunctionOutput
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_cloudfront_connection_function.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.CloudFrontEndpointID)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.CloudFrontServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckConnectionFunctionDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccConnectionFunctionConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckConnectionFunctionExists(ctx, t, resourceName, &connectionfunction),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("connection_function_arn"), tfknownvalue.GlobalARNRegexp("cloudfront", regexache.MustCompile(`connection-function/.+`))),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("etag"), knownvalue.NotNull()),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("live_stage_etag"), knownvalue.Null()),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrStatus), knownvalue.StringExact("UNPUBLISHED")),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrTags), knownvalue.Null()),
				},
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"publish",
				},
			},
		},
	})
}

func TestAccCloudFrontConnectionFunction_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	var connectionfunction cloudfront.DescribeConnectionFunctionOutput
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_cloudfront_connection_function.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.CloudFrontEndpointID)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.CloudFrontServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckConnectionFunctionDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccConnectionFunctionConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckConnectionFunctionExists(ctx, t, resourceName, &connectionfunction),
					acctest.CheckFrameworkResourceDisappears(ctx, t, tfcloudfront.ResourceConnectionFunction, resourceName),
				),
				ExpectNonEmptyPlan: true,
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
					},
					PostApplyPostRefresh: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
					},
				},
			},
		},
	})
}

func TestAccCloudFrontConnectionFunction_publishOnCreate(t *testing.T) {
	ctx := acctest.Context(t)
	var connectionfunction cloudfront.DescribeConnectionFunctionOutput
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_cloudfront_connection_function.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.CloudFrontEndpointID)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.CloudFrontServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckConnectionFunctionDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccConnectionFunctionConfig_publish(rName, true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckConnectionFunctionExists(ctx, t, resourceName, &connectionfunction),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("etag"), knownvalue.NotNull()),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("live_stage_etag"), knownvalue.NotNull()),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrStatus), knownvalue.StringExact("UNASSOCIATED")),
				},
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"publish",
				},
			},
		},
	})
}

func TestAccCloudFrontConnectionFunction_publishOnUpdate(t *testing.T) {
	ctx := acctest.Context(t)
	var connectionfunction cloudfront.DescribeConnectionFunctionOutput
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_cloudfront_connection_function.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.CloudFrontEndpointID)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.CloudFrontServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckConnectionFunctionDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccConnectionFunctionConfig_publish(rName, false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckConnectionFunctionExists(ctx, t, resourceName, &connectionfunction),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("etag"), knownvalue.NotNull()),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("live_stage_etag"), knownvalue.Null()),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrStatus), knownvalue.StringExact("UNPUBLISHED")),
				},
			},
			{
				Config: testAccConnectionFunctionConfig_publish(rName, true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckConnectionFunctionExists(ctx, t, resourceName, &connectionfunction),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionUpdate),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("etag"), knownvalue.NotNull()),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("live_stage_etag"), knownvalue.NotNull()),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrStatus), knownvalue.StringExact("UNASSOCIATED")),
				},
			},
		},
	})
}

func TestAccCloudFrontConnectionFunction_update(t *testing.T) {
	ctx := acctest.Context(t)
	var connectionfunction1, connectionfunction2 cloudfront.DescribeConnectionFunctionOutput
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_cloudfront_connection_function.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.CloudFrontEndpointID)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.CloudFrontServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckConnectionFunctionDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccConnectionFunctionConfig_updateInitial(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckConnectionFunctionExists(ctx, t, resourceName, &connectionfunction1),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
					},
				},
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"publish",
				},
			},
			{
				Config: testAccConnectionFunctionConfig_updateComplete(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckConnectionFunctionExists(ctx, t, resourceName, &connectionfunction2),
					testAccCheckConnectionFunctionEtagChanged(&connectionfunction1, &connectionfunction2),
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

func TestAccCloudFrontConnectionFunction_tags(t *testing.T) {
	ctx := acctest.Context(t)
	var connectionfunction cloudfront.DescribeConnectionFunctionOutput
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_cloudfront_connection_function.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.CloudFrontEndpointID)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.CloudFrontServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckConnectionFunctionDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccConnectionFunctionConfig_tags1(rName, acctest.CtKey1, acctest.CtValue1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckConnectionFunctionExists(ctx, t, resourceName, &connectionfunction),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrTags), knownvalue.MapExact(map[string]knownvalue.Check{
						acctest.CtKey1: knownvalue.StringExact(acctest.CtValue1),
					})),
				},
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"publish",
				},
			},
			{
				Config: testAccConnectionFunctionConfig_tags2(rName, acctest.CtKey1, acctest.CtValue1Updated, acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckConnectionFunctionExists(ctx, t, resourceName, &connectionfunction),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionUpdate),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrTags), knownvalue.MapExact(map[string]knownvalue.Check{
						acctest.CtKey1: knownvalue.StringExact(acctest.CtValue1Updated),
						acctest.CtKey2: knownvalue.StringExact(acctest.CtValue2),
					})),
				},
			},
			{
				Config: testAccConnectionFunctionConfig_tags1(rName, acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckConnectionFunctionExists(ctx, t, resourceName, &connectionfunction),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionUpdate),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrTags), knownvalue.MapExact(map[string]knownvalue.Check{
						acctest.CtKey2: knownvalue.StringExact(acctest.CtValue2),
					})),
				},
			},
		},
	})
}

func testAccCheckConnectionFunctionDestroy(ctx context.Context, t *testing.T) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.ProviderMeta(ctx, t).CloudFrontClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_cloudfront_connection_function" {
				continue
			}

			_, err := tfcloudfront.FindConnectionFunctionByTwoPartKey(ctx, conn, rs.Primary.ID, awstypes.FunctionStageDevelopment)
			if retry.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("CloudFront Connection Function %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccCheckConnectionFunctionExists(ctx context.Context, t *testing.T, n string, v *cloudfront.DescribeConnectionFunctionOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.ProviderMeta(ctx, t).CloudFrontClient(ctx)

		output, err := tfcloudfront.FindConnectionFunctionByTwoPartKey(ctx, conn, rs.Primary.ID, awstypes.FunctionStageDevelopment)

		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

func testAccCheckConnectionFunctionEtagChanged(before, after *cloudfront.DescribeConnectionFunctionOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if before, after := aws.ToString(before.ETag), aws.ToString(after.ETag); after == before {
			return fmt.Errorf("Etag did not change: %s", after)
		}

		return nil
	}
}

func testAccConnectionFunctionConfig_basic(rName string) string {
	return fmt.Sprintf(`
resource "aws_cloudfront_connection_function" "test" {
  name                     = %[1]q
  connection_function_code = "function handler(event) { return event.request; }"

  connection_function_config {
    comment = "Test connection function"
    runtime = "cloudfront-js-2.0"
  }
}
`, rName)
}

func testAccConnectionFunctionConfig_publish(rName string, publish bool) string {
	return fmt.Sprintf(`
resource "aws_cloudfront_connection_function" "test" {
  name                     = %[1]q
  connection_function_code = "function handler(event) { return event.request; }"

  connection_function_config {
    comment = "Test connection function"
    runtime = "cloudfront-js-2.0"
  }

  publish = %[2]t
}
`, rName, publish)
}

func testAccConnectionFunctionConfig_updateInitial(rName string) string {
	return fmt.Sprintf(`
resource "aws_cloudfront_key_value_store" "test1" {
  name    = "%[1]s-1"
  comment = "Test key value store for update test 1"
}

resource "aws_cloudfront_connection_function" "test" {
  name                     = %[1]q
  connection_function_code = <<-EOT
function handler(event) {
  console.log("Initial function execution with runtime 2.0");
  return event.request;
}
EOT

  connection_function_config {
    comment = "Initial test connection function with runtime 2.0"
    runtime = "cloudfront-js-2.0"

    key_value_store_association {
      key_value_store_arn = aws_cloudfront_key_value_store.test1.arn
    }
  }
}
`, rName)
}

func testAccConnectionFunctionConfig_updateComplete(rName string) string {
	return fmt.Sprintf(`
resource "aws_cloudfront_key_value_store" "test1" {
  name    = "%[1]s-1"
  comment = "Test key value store for update test 1"
}

resource "aws_cloudfront_key_value_store" "test2" {
  name    = "%[1]s-2"
  comment = "Test key value store for update test 2"
}

resource "aws_cloudfront_connection_function" "test" {
  name                     = %[1]q
  connection_function_code = <<-EOT
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

  connection_function_config {
    comment = "Updated test connection function with all attributes"
    runtime = "cloudfront-js-2.0"

    key_value_store_association {
      key_value_store_arn = aws_cloudfront_key_value_store.test2.arn
    }
  }
}
`, rName)
}

func testAccConnectionFunctionConfig_tags1(rName, tagKey1, tagValue1 string) string {
	return fmt.Sprintf(`
resource "aws_cloudfront_connection_function" "test" {
  name                     = %[1]q
  connection_function_code = "function handler(event) { return event.request; }"

  connection_function_config {
    comment = "Test connection function"
    runtime = "cloudfront-js-2.0"
  }

  tags = {
    %[2]q = %[3]q
  }
}
`, rName, tagKey1, tagValue1)
}

func testAccConnectionFunctionConfig_tags2(rName, tagKey1, tagValue1, tagKey2, tagValue2 string) string {
	return fmt.Sprintf(`
resource "aws_cloudfront_connection_function" "test" {
  name                     = %[1]q
  connection_function_code = "function handler(event) { return event.request; }"

  connection_function_config {
    comment = "Test connection function"
    runtime = "cloudfront-js-2.0"
  }

  tags = {
    %[2]q = %[3]q
    %[4]q = %[5]q
  }
}
`, rName, tagKey1, tagValue1, tagKey2, tagValue2)
}
