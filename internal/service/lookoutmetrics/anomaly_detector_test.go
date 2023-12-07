// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package lookoutmetrics_test

// // **PLEASE DELETE THIS AND ALL TIP COMMENTS BEFORE SUBMITTING A PR FOR REVIEW!**
// //
// // TIP: ==== INTRODUCTION ====
// // Thank you for trying the skaff tool!
// //
// // You have opted to include these helpful comments. They all include "TIP:"
// // to help you find and remove them when you're done with them.
// //
// // While some aspects of this file are customized to your input, the
// // scaffold tool does *not* look at the AWS API and ensure it has correct
// // function, structure, and variable names. It makes guesses based on
// // commonalities. You will need to make significant adjustments.
// //
// // In other words, as generated, this is a rough outline of the work you will
// // need to do. If something doesn't make sense for your situation, get rid of
// // it.
// //
// // Remember to register this new resource in the provider
// // (internal/provider/provider.go) once you finish. Otherwise, Terraform won't
// // know about it.

// import (
// 	// TIP: ==== IMPORTS ====
// 	// This is a common set of imports but not customized to your code since
// 	// your code hasn't been written yet. Make sure you, your IDE, or
// 	// goimports -w <file> fixes these imports.
// 	//
// 	// The provider linter wants your imports to be in two groups: first,
// 	// standard library (i.e., "fmt" or "strings"), second, everything else.
// 	//
// 	// Also, AWS Go SDK v2 may handle nested structures differently than v1,
// 	// using the services/lookoutmetrics/types package. If so, you'll
// 	// need to import types and reference the nested types, e.g., as
// 	// types.<Type Name>.
// 	"context"
// 	"errors"
// 	"fmt"
// 	"regexp"
// 	"testing"

// 	"github.com/aws/aws-sdk-go-v2/aws"
// 	"github.com/aws/aws-sdk-go-v2/service/lookoutmetrics"
// 	"github.com/aws/aws-sdk-go-v2/service/lookoutmetrics/types"
// 	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
// 	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
// 	"github.com/hashicorp/terraform-plugin-testing/terraform"
// 	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
// 	"github.com/hashicorp/terraform-provider-aws/internal/conns"
// 	"github.com/hashicorp/terraform-provider-aws/internal/create"

// 	// TIP: You will often need to import the package that this test file lives
// 	// in. Since it is in the "test" context, it must import the package to use
// 	// any normal context constants, variables, or functions.
// 	tflookoutmetrics "github.com/hashicorp/terraform-provider-aws/internal/service/lookoutmetrics"
// 	"github.com/hashicorp/terraform-provider-aws/names"
// )

// // TIP: File Structure. The basic outline for all test files should be as
// // follows. Improve this resource's maintainability by following this
// // outline.
// //
// // 1. Package declaration (add "_test" since this is a test file)
// // 2. Imports
// // 3. Unit tests
// // 4. Basic test
// // 5. Disappears test
// // 6. All the other tests
// // 7. Helper functions (exists, destroy, check, etc.)
// // 8. Functions that return Terraform configurations

// // TIP: ==== UNIT TESTS ====
// // This is an example of a unit test. Its name is not prefixed with
// // "TestAcc" like an acceptance test.
// //
// // Unlike acceptance tests, unit tests do not access AWS and are focused on a
// // function (or method). Because of this, they are quick and cheap to run.
// //
// // In designing a resource's implementation, isolate complex bits from AWS bits
// // so that they can be tested through a unit test. We encourage more unit tests
// // in the provider.
// //
// // Cut and dry functions using well-used patterns, like typical flatteners and
// // expanders, don't need unit testing. However, if they are complex or
// // intricate, they should be unit tested.
// func TestAnomalyDetectorExampleUnitTest(t *testing.T) {
// 	testCases := []struct {
// 		TestName string
// 		Input    string
// 		Expected string
// 		Error    bool
// 	}{
// 		{
// 			TestName: "empty",
// 			Input:    "",
// 			Expected: "",
// 			Error:    true,
// 		},
// 		{
// 			TestName: "descriptive name",
// 			Input:    "some input",
// 			Expected: "some output",
// 			Error:    false,
// 		},
// 		{
// 			TestName: "another descriptive name",
// 			Input:    "more input",
// 			Expected: "more output",
// 			Error:    false,
// 		},
// 	}

// 	for _, testCase := range testCases {
// 		t.Run(testCase.TestName, func(t *testing.T) {
// 			got, err := tflookoutmetrics.FunctionFromResource(testCase.Input)

// 			if err != nil && !testCase.Error {
// 				t.Errorf("got error (%s), expected no error", err)
// 			}

// 			if err == nil && testCase.Error {
// 				t.Errorf("got (%s) and no error, expected error", got)
// 			}

// 			if got != testCase.Expected {
// 				t.Errorf("got %s, expected %s", got, testCase.Expected)
// 			}
// 		})
// 	}
// }

// // TIP: ==== ACCEPTANCE TESTS ====
// // This is an example of a basic acceptance test. This should test as much of
// // standard functionality of the resource as possible, and test importing, if
// // applicable. We prefix its name with "TestAcc", the service, and the
// // resource name.
// //
// // Acceptance test access AWS and cost money to run.
// func TestAccLookoutMetricsAnomalyDetector_basic(t *testing.T) {
// 	ctx := acctest.Context(t)
// 	// TIP: This is a long-running test guard for tests that run longer than
// 	// 300s (5 min) generally.
// 	if testing.Short() {
// 		t.Skip("skipping long-running test in short mode")
// 	}

// 	var anomalydetector lookoutmetrics.DescribeAnomalyDetectorResponse
// 	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
// 	resourceName := "aws_lookoutmetrics_anomaly_detector.test"

// 	resource.ParallelTest(t, resource.TestCase{
// 		PreCheck: func() {
// 			acctest.PreCheck(ctx, t)
// 			acctest.PreCheckPartitionHasService(t, names.LookoutMetricsEndpointID)
// 			testAccPreCheck(ctx, t)
// 		},
// 		ErrorCheck:               acctest.ErrorCheck(t, names.LookoutMetricsEndpointID),
// 		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
// 		CheckDestroy:             testAccCheckAnomalyDetectorDestroy(ctx),
// 		Steps: []resource.TestStep{
// 			{
// 				Config: testAccAnomalyDetectorConfig_basic(rName),
// 				Check: resource.ComposeTestCheckFunc(
// 					testAccCheckAnomalyDetectorExists(ctx, resourceName, &anomalydetector),
// 					resource.TestCheckResourceAttr(resourceName, "auto_minor_version_upgrade", "false"),
// 					resource.TestCheckResourceAttrSet(resourceName, "maintenance_window_start_time.0.day_of_week"),
// 					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "user.*", map[string]string{
// 						"console_access": "false",
// 						"groups.#":       "0",
// 						"username":       "Test",
// 						"password":       "TestTest1234",
// 					}),
// 					acctest.MatchResourceAttrRegionalARN(resourceName, "arn", "lookoutmetrics", regexp.MustCompile(`anomalydetector:+.`)),
// 				),
// 			},
// 			{
// 				ResourceName:            resourceName,
// 				ImportState:             true,
// 				ImportStateVerify:       true,
// 				ImportStateVerifyIgnore: []string{"apply_immediately", "user"},
// 			},
// 		},
// 	})
// }

// func TestAccLookoutMetricsAnomalyDetector_disappears(t *testing.T) {
// 	ctx := acctest.Context(t)
// 	if testing.Short() {
// 		t.Skip("skipping long-running test in short mode")
// 	}

// 	var anomalydetector lookoutmetrics.DescribeAnomalyDetectorResponse
// 	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
// 	resourceName := "aws_lookoutmetrics_anomaly_detector.test"

// 	resource.ParallelTest(t, resource.TestCase{
// 		PreCheck: func() {
// 			acctest.PreCheck(ctx, t)
// 			acctest.PreCheckPartitionHasService(t, names.LookoutMetricsEndpointID)
// 			testAccPreCheck(t)
// 		},
// 		ErrorCheck:               acctest.ErrorCheck(t, names.LookoutMetricsEndpointID),
// 		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
// 		CheckDestroy:             testAccCheckAnomalyDetectorDestroy(ctx),
// 		Steps: []resource.TestStep{
// 			{
// 				Config: testAccAnomalyDetectorConfig_basic(rName, testAccAnomalyDetectorVersionNewer),
// 				Check: resource.ComposeTestCheckFunc(
// 					testAccCheckAnomalyDetectorExists(ctx, resourceName, &anomalydetector),
// 					acctest.CheckResourceDisappears(ctx, acctest.Provider, tflookoutmetrics.ResourceAnomalyDetector(), resourceName),
// 				),
// 				ExpectNonEmptyPlan: true,
// 			},
// 		},
// 	})
// }

// func testAccCheckAnomalyDetectorDestroy(ctx context.Context) resource.TestCheckFunc {
// 	return func(s *terraform.State) error {
// 		conn := acctest.Provider.Meta().(*conns.AWSClient).LookoutMetricsClient()

// 		for _, rs := range s.RootModule().Resources {
// 			if rs.Type != "aws_lookoutmetrics_anomaly_detector" {
// 				continue
// 			}

// 			input := &lookoutmetrics.DescribeAnomalyDetectorInput{
// 				AnomalyDetectorId: aws.String(rs.Primary.ID),
// 			}
// 			_, err := conn.DescribeAnomalyDetector(ctx, &lookoutmetrics.DescribeAnomalyDetectorInput{
// 				AnomalyDetectorId: aws.String(rs.Primary.ID),
// 			})
// 			if err != nil {
// 				var nfe *types.ResourceNotFoundException
// 				if errors.As(err, &nfe) {
// 					return nil
// 				}
// 				return err
// 			}

// 			return create.Error(names.LookoutMetrics, create.ErrActionCheckingDestroyed, tflookoutmetrics.ResNameAnomalyDetector, rs.Primary.ID, errors.New("not destroyed"))
// 		}

// 		return nil
// 	}
// }

// func testAccCheckAnomalyDetectorExists(ctx context.Context, name string, anomalydetector *lookoutmetrics.DescribeAnomalyDetectorResponse) resource.TestCheckFunc {
// 	return func(s *terraform.State) error {
// 		rs, ok := s.RootModule().Resources[name]
// 		if !ok {
// 			return create.Error(names.LookoutMetrics, create.ErrActionCheckingExistence, tflookoutmetrics.ResNameAnomalyDetector, name, errors.New("not found"))
// 		}

// 		if rs.Primary.ID == "" {
// 			return create.Error(names.LookoutMetrics, create.ErrActionCheckingExistence, tflookoutmetrics.ResNameAnomalyDetector, name, errors.New("not set"))
// 		}

// 		conn := acctest.Provider.Meta().(*conns.AWSClient).LookoutMetricsClient()
// 		resp, err := conn.DescribeAnomalyDetector(ctx, &lookoutmetrics.DescribeAnomalyDetectorInput{
// 			AnomalyDetectorId: aws.String(rs.Primary.ID),
// 		})

// 		if err != nil {
// 			return create.Error(names.LookoutMetrics, create.ErrActionCheckingExistence, tflookoutmetrics.ResNameAnomalyDetector, rs.Primary.ID, err)
// 		}

// 		*anomalydetector = *resp

// 		return nil
// 	}
// }

// func testAccPreCheck(ctx context.Context, t *testing.T) {
// 	conn := acctest.Provider.Meta().(*conns.AWSClient).LookoutMetricsClient()

// 	input := &lookoutmetrics.ListAnomalyDetectorsInput{}
// 	_, err := conn.ListAnomalyDetectors(ctx, input)

// 	if acctest.PreCheckSkipError(err) {
// 		t.Skipf("skipping acceptance testing: %s", err)
// 	}

// 	if err != nil {
// 		t.Fatalf("unexpected PreCheck error: %s", err)
// 	}
// }

// func testAccCheckAnomalyDetectorNotRecreated(before, after *lookoutmetrics.DescribeAnomalyDetectorResponse) resource.TestCheckFunc {
// 	return func(s *terraform.State) error {
// 		if before, after := aws.StringValue(before.AnomalyDetectorId), aws.StringValue(after.AnomalyDetectorId); before != after {
// 			return create.Error(names.LookoutMetrics, create.ErrActionCheckingNotRecreated, tflookoutmetrics.ResNameAnomalyDetector, aws.StringValue(before.AnomalyDetectorId), errors.New("recreated"))
// 		}

// 		return nil
// 	}
// }

// func testAccAnomalyDetectorConfig_basic(rName, version string) string {
// 	return fmt.Sprintf(`
// resource "aws_security_group" "test" {
//   name = %[1]q
// }

// resource "aws_lookoutmetrics_anomaly_detector" "test" {
//   anomaly_detector_name   = %[1]q
//   engine_type             = "ActiveLookoutMetrics"
//   engine_version          = %[2]q
//   host_instance_type      = "lookoutmetrics.t2.micro"
//   security_groups         = [aws_security_group.test.id]
//   authentication_strategy = "simple"
//   storage_type            = "efs"

//   logs {
//     general = true
//   }

//   user {
//     username = "Test"
//     password = "TestTest1234"
//   }
// }
// `, rName, version)
// }
