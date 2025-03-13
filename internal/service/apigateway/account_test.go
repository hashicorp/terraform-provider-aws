// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package apigateway_test

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/apigateway"
	awstypes "github.com/aws/aws-sdk-go-v2/service/apigateway/types"
	"github.com/hashicorp/terraform-plugin-testing/compare"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfapigateway "github.com/hashicorp/terraform-provider-aws/internal/service/apigateway"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func testAccAccount_basic(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_api_gateway_account.test"

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			t.Cleanup(accountCleanup(ctx, t))
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.APIGatewayServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckAccountDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccAccountConfig_basic,
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("api_key_version"), knownvalue.NotNull()),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("cloudwatch_role_arn"), knownvalue.StringExact("")),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("features"), knownvalue.NotNull()),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrID), knownvalue.StringExact("api-gateway-account")),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("reset_on_delete"), knownvalue.Null()),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("throttle_settings"), knownvalue.ListExact([]knownvalue.Check{
						knownvalue.ObjectExact(map[string]knownvalue.Check{
							"burst_limit": knownvalue.Int32Exact(5000),
							"rate_limit":  knownvalue.Float64Exact(10000),
						}),
					})),
				},
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccAccount_cloudwatchRoleARN_value(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_api_gateway_account.test"

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			t.Cleanup(accountCleanup(ctx, t))
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.APIGatewayServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckAccountDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccAccountConfig_role0(rName),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.CompareValuePairs(
						resourceName, tfjsonpath.New("cloudwatch_role_arn"),
						"aws_iam_role.test[0]", tfjsonpath.New(names.AttrARN),
						compare.ValuesSame(),
					),
				},
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccAccountConfig_role1(rName),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.CompareValuePairs(
						resourceName, tfjsonpath.New("cloudwatch_role_arn"),
						"aws_iam_role.test[1]", tfjsonpath.New(names.AttrARN),
						compare.ValuesSame(),
					),
				},
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccAccountConfig_basic,
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("cloudwatch_role_arn"), knownvalue.StringExact("")),
				},
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccAccount_cloudwatchRoleARN_empty(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_api_gateway_account.test"

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			t.Cleanup(accountCleanup(ctx, t))
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.APIGatewayServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckAccountDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccAccountConfig_empty,
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("cloudwatch_role_arn"), knownvalue.StringExact("")),
				},
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccAccount_resetOnDelete_false(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_api_gateway_account.test"

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			t.Cleanup(accountCleanup(ctx, t))
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.APIGatewayServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckAccountNotDestroyed(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccAccountConfig_resetOnDelete(rName, false),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.CompareValuePairs(
						resourceName, tfjsonpath.New("cloudwatch_role_arn"),
						"aws_iam_role.test[0]", tfjsonpath.New(names.AttrARN),
						compare.ValuesSame(),
					),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("reset_on_delete"), knownvalue.Bool(false)),
				},
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"reset_on_delete",
				},
			},
		},
	})
}

func testAccAccount_resetOnDelete_true(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_api_gateway_account.test"

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			t.Cleanup(accountCleanup(ctx, t))
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.APIGatewayServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckAccountDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccAccountConfig_resetOnDelete(rName, true),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.CompareValuePairs(
						resourceName, tfjsonpath.New("cloudwatch_role_arn"),
						"aws_iam_role.test[0]", tfjsonpath.New(names.AttrARN),
						compare.ValuesSame(),
					),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("reset_on_delete"), knownvalue.Bool(true)),
				},
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"reset_on_delete",
				},
			},
		},
	})
}

func testAccAccount_frameworkMigration_basic(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_api_gateway_account.test"

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			t.Cleanup(accountCleanup(ctx, t))
		},
		ErrorCheck:   acctest.ErrorCheck(t, names.APIGatewayServiceID),
		CheckDestroy: testAccCheckAccountDestroy(ctx),
		Steps: []resource.TestStep{
			{
				ExternalProviders: map[string]resource.ExternalProvider{
					"aws": {
						Source:            "hashicorp/aws",
						VersionConstraint: "5.74.0",
					},
				},
				Config: testAccAccountConfig_basic,
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("cloudwatch_role_arn"), knownvalue.StringExact("")),
				},
			},
			{
				ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
				Config:                   testAccAccountConfig_basic,
				PlanOnly:                 true,
			},
		},
	})
}

func testAccAccount_frameworkMigration_cloudwatchRoleARN(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_api_gateway_account.test"

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			t.Cleanup(accountCleanup(ctx, t))
		},
		ErrorCheck:   acctest.ErrorCheck(t, names.APIGatewayServiceID),
		CheckDestroy: testAccCheckAccountNotDestroyed(ctx),
		Steps: []resource.TestStep{
			{
				ExternalProviders: map[string]resource.ExternalProvider{
					"aws": {
						Source:            "hashicorp/aws",
						VersionConstraint: "5.74.0",
					},
				},
				Config: testAccAccountConfig_role0(rName),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.CompareValuePairs(
						resourceName, tfjsonpath.New("cloudwatch_role_arn"),
						"aws_iam_role.test[0]", tfjsonpath.New(names.AttrARN),
						compare.ValuesSame(),
					),
				},
			},
			{
				ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
				Config:                   testAccAccountConfig_role0(rName),
				PlanOnly:                 true,
			},
		},
	})
}

func testAccCheckAccountDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).APIGatewayClient(ctx)

		account, err := tfapigateway.FindAccount(ctx, conn)
		if err != nil {
			return err
		}

		if account.CloudwatchRoleArn == nil {
			// Settings have been reset
			return nil
		}

		return errors.New("API Gateway Account still exists")
	}
}

func testAccCheckAccountNotDestroyed(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).APIGatewayClient(ctx)

		account, err := tfapigateway.FindAccount(ctx, conn)
		if err != nil {
			return err
		}

		if account.CloudwatchRoleArn != nil {
			// Settings have not been reset
			return nil
		}

		return errors.New("API Gateway Account was reset")
	}
}

func accountCleanup(ctx context.Context, t *testing.T) func() {
	return func() {
		t.Helper()

		conn := acctest.Provider.Meta().(*conns.AWSClient).APIGatewayClient(ctx)

		input := apigateway.UpdateAccountInput{
			PatchOperations: []awstypes.PatchOperation{
				{
					Op:    awstypes.OpReplace,
					Path:  aws.String("/cloudwatchRoleArn"),
					Value: nil,
				},
			},
		}

		if _, err := conn.UpdateAccount(ctx, &input); err != nil {
			t.Errorf("API Gateway Account cleanup: %s", err)
		}
	}
}

const testAccAccountConfig_basic = `
resource "aws_api_gateway_account" "test" {}
`

func testAccAccountConfig_base(rName string) string {
	return fmt.Sprintf(`
data "aws_partition" "current" {}

resource "aws_iam_role" "test" {
  count = 2

  name = "%[1]s-${count.index}"

  assume_role_policy = jsonencode({
    Version = "2012-10-17"
    Statement = [{
      Action = "sts:AssumeRole"
      Effect = "Allow"
      Principal = {
        Service = "apigateway.amazonaws.com"
      }
    }]
  })

  managed_policy_arns = ["arn:${data.aws_partition.current.partition}:iam::aws:policy/service-role/AmazonAPIGatewayPushToCloudWatchLogs"]
}
`, rName)
}

func testAccAccountConfig_role0(rName string) string {
	return acctest.ConfigCompose(
		testAccAccountConfig_base(rName), `
resource "aws_api_gateway_account" "test" {
  cloudwatch_role_arn = aws_iam_role.test[0].arn
}
`)
}

func testAccAccountConfig_role1(rName string) string {
	return acctest.ConfigCompose(
		testAccAccountConfig_base(rName), `
resource "aws_api_gateway_account" "test" {
  cloudwatch_role_arn = aws_iam_role.test[1].arn
}
`)
}

const testAccAccountConfig_empty = `
resource "aws_api_gateway_account" "test" {
  cloudwatch_role_arn = ""
}
`

func testAccAccountConfig_resetOnDelete(rName string, reset bool) string {
	return acctest.ConfigCompose(
		testAccAccountConfig_base(rName),
		fmt.Sprintf(`
resource "aws_api_gateway_account" "test" {
  cloudwatch_role_arn = aws_iam_role.test[0].arn
  reset_on_delete     = %[1]t
}
`, reset))
}
