// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package lambda_test

//import (
//	"context"
//	"errors"
//	"fmt"
//	"testing"
//
//	"github.com/YakDriver/regexache"
//	"github.com/aws/aws-sdk-go-v2/aws"
//	"github.com/aws/aws-sdk-go-v2/service/lambda"
//	"github.com/aws/aws-sdk-go-v2/service/lambda/types"
//	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
//	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
//	"github.com/hashicorp/terraform-plugin-testing/plancheck"
//	"github.com/hashicorp/terraform-plugin-testing/terraform"
//	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
//	"github.com/hashicorp/terraform-provider-aws/internal/conns"
//	"github.com/hashicorp/terraform-provider-aws/internal/create"
//	"github.com/hashicorp/terraform-provider-aws/internal/errs"
//	tflambda "github.com/hashicorp/terraform-provider-aws/internal/service/lambda"
//	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
//	"github.com/hashicorp/terraform-provider-aws/names"
//)
//
//func TestAccLambdaCapacityProvider_basic(t *testing.T) {
//	ctx := acctest.Context(t)
//	if testing.Short() {
//		t.Skip("skipping long-running test in short mode")
//	}
//
//	var capacityprovider lambda.DescribeCapacityProviderResponse
//	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
//	resourceName := "aws_lambda_capacity_provider.test"
//
//	resource.ParallelTest(t, resource.TestCase{
//		PreCheck: func() {
//			acctest.PreCheck(ctx, t)
//			acctest.PreCheckPartitionHasService(t, names.LambdaEndpointID)
//			testAccPreCheck(ctx, t)
//		},
//		ErrorCheck:               acctest.ErrorCheck(t, names.LambdaServiceID),
//		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
//		CheckDestroy:             testAccCheckCapacityProviderDestroy(ctx),
//		Steps: []resource.TestStep{
//			{
//				Config: testAccCapacityProviderConfig_basic(rName),
//				Check: resource.ComposeAggregateTestCheckFunc(
//					testAccCheckCapacityProviderExists(ctx, resourceName, &capacityprovider),
//					resource.TestCheckResourceAttr(resourceName, "auto_minor_version_upgrade", "false"),
//					resource.TestCheckResourceAttrSet(resourceName, "maintenance_window_start_time.0.day_of_week"),
//					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "user.*", map[string]string{
//						"console_access": "false",
//						"groups.#":       "0",
//						"username":       "Test",
//						"password":       "TestTest1234",
//					}),
//					acctest.MatchResourceAttrRegionalARN(ctx, resourceName, names.AttrARN, "lambda", regexache.MustCompile(`capacityprovider:.+$`)),
//				),
//			},
//			{
//				ResourceName:            resourceName,
//				ImportState:             true,
//				ImportStateVerify:       true,
//				ImportStateVerifyIgnore: []string{"apply_immediately", "user"},
//			},
//		},
//	})
//}
//
//func TestAccLambdaCapacityProvider_disappears(t *testing.T) {
//	ctx := acctest.Context(t)
//	if testing.Short() {
//		t.Skip("skipping long-running test in short mode")
//	}
//
//	var capacityprovider lambda.DescribeCapacityProviderResponse
//	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
//	resourceName := "aws_lambda_capacity_provider.test"
//
//	resource.ParallelTest(t, resource.TestCase{
//		PreCheck: func() {
//			acctest.PreCheck(ctx, t)
//			acctest.PreCheckPartitionHasService(t, names.LambdaEndpointID)
//			testAccPreCheck(ctx, t)
//		},
//		ErrorCheck:               acctest.ErrorCheck(t, names.LambdaServiceID),
//		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
//		CheckDestroy:             testAccCheckCapacityProviderDestroy(ctx),
//		Steps: []resource.TestStep{
//			{
//				Config: testAccCapacityProviderConfig_basic(rName, testAccCapacityProviderVersionNewer),
//				Check: resource.ComposeAggregateTestCheckFunc(
//					testAccCheckCapacityProviderExists(ctx, resourceName, &capacityprovider),
//					acctest.CheckFrameworkResourceDisappears(ctx, acctest.Provider, tflambda.ResourceCapacityProvider, resourceName),
//				),
//				ExpectNonEmptyPlan: true,
//				ConfigPlanChecks: resource.ConfigPlanChecks{
//					PostApplyPostRefresh: []plancheck.PlanCheck{
//						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
//					},
//				},
//			},
//		},
//	})
//}
//
//func testAccCheckCapacityProviderDestroy(ctx context.Context) resource.TestCheckFunc {
//	return func(s *terraform.State) error {
//		conn := acctest.Provider.Meta().(*conns.AWSClient).LambdaClient(ctx)
//
//		for _, rs := range s.RootModule().Resources {
//			if rs.Type != "aws_lambda_capacity_provider" {
//				continue
//			}
//
//			_, err := tflambda.FindCapacityProviderByID(ctx, conn, rs.Primary.ID)
//			if tfresource.NotFound(err) {
//				return nil
//			}
//			if err != nil {
//				return create.Error(names.Lambda, create.ErrActionCheckingDestroyed, tflambda.ResNameCapacityProvider, rs.Primary.ID, err)
//			}
//
//			return create.Error(names.Lambda, create.ErrActionCheckingDestroyed, tflambda.ResNameCapacityProvider, rs.Primary.ID, errors.New("not destroyed"))
//		}
//
//		return nil
//	}
//}
//
//func testAccCheckCapacityProviderExists(ctx context.Context, name string, capacityprovider *lambda.DescribeCapacityProviderResponse) resource.TestCheckFunc {
//	return func(s *terraform.State) error {
//		rs, ok := s.RootModule().Resources[name]
//		if !ok {
//			return create.Error(names.Lambda, create.ErrActionCheckingExistence, tflambda.ResNameCapacityProvider, name, errors.New("not found"))
//		}
//
//		if rs.Primary.ID == "" {
//			return create.Error(names.Lambda, create.ErrActionCheckingExistence, tflambda.ResNameCapacityProvider, name, errors.New("not set"))
//		}
//
//		conn := acctest.Provider.Meta().(*conns.AWSClient).LambdaClient(ctx)
//
//		resp, err := tflambda.FindCapacityProviderByID(ctx, conn, rs.Primary.ID)
//		if err != nil {
//			return create.Error(names.Lambda, create.ErrActionCheckingExistence, tflambda.ResNameCapacityProvider, rs.Primary.ID, err)
//		}
//
//		*capacityprovider = *resp
//
//		return nil
//	}
//}
//
//func testAccPreCheck(ctx context.Context, t *testing.T) {
//	conn := acctest.Provider.Meta().(*conns.AWSClient).LambdaClient(ctx)
//
//	input := &lambda.ListCapacityProvidersInput{}
//
//	_, err := conn.ListCapacityProviders(ctx, input)
//
//	if acctest.PreCheckSkipError(err) {
//		t.Skipf("skipping acceptance testing: %s", err)
//	}
//	if err != nil {
//		t.Fatalf("unexpected PreCheck error: %s", err)
//	}
//}
//
//func testAccCheckCapacityProviderNotRecreated(before, after *lambda.DescribeCapacityProviderResponse) resource.TestCheckFunc {
//	return func(s *terraform.State) error {
//		if before, after := aws.ToString(before.CapacityProviderId), aws.ToString(after.CapacityProviderId); before != after {
//			return create.Error(names.Lambda, create.ErrActionCheckingNotRecreated, tflambda.ResNameCapacityProvider, aws.ToString(before.CapacityProviderId), errors.New("recreated"))
//		}
//
//		return nil
//	}
//}
//
//func testAccCapacityProviderConfig_basic(rName, version string) string {
//	return fmt.Sprintf(`
//resource "aws_security_group" "test" {
//  name = %[1]q
//}
//
//resource "aws_lambda_capacity_provider" "test" {
//  capacity_provider_name             = %[1]q
//  engine_type             = "ActiveLambda"
//  engine_version          = %[2]q
//  host_instance_type      = "lambda.t2.micro"
//  security_groups         = [aws_security_group.test.id]
//  authentication_strategy = "simple"
//  storage_type            = "efs"
//
//  logs {
//    general = true
//  }
//
//  user {
//    username = "Test"
//    password = "TestTest1234"
//  }
//}
//`, rName, version)
//}
