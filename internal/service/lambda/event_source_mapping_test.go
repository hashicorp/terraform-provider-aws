// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package lambda_test

import (
	"context"
	"fmt"
	"strconv"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/kafka"
	"github.com/aws/aws-sdk-go-v2/service/lambda"
	awstypes "github.com/aws/aws-sdk-go-v2/service/lambda/types"
	"github.com/aws/aws-sdk-go-v2/service/mq"
	"github.com/aws/aws-sdk-go-v2/service/secretsmanager"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	tflambda "github.com/hashicorp/terraform-provider-aws/internal/service/lambda"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccLambdaEventSourceMapping_Kinesis_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var conf lambda.GetEventSourceMappingOutput
	resourceName := "aws_lambda_event_source_mapping.test"
	functionResourceName := "aws_lambda_function.test"
	functionResourceNameUpdated := "aws_lambda_function.test_update"
	eventSourceResourceName := "aws_kinesis_stream.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.LambdaServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckEventSourceMappingDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccEventSourceMappingConfig_kinesisBatchSize(rName, "100"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckEventSourceMappingExists(ctx, resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "batch_size", "100"),
					resource.TestCheckResourceAttr(resourceName, names.AttrEnabled, acctest.CtTrue),
					resource.TestCheckResourceAttrPair(resourceName, "event_source_arn", eventSourceResourceName, names.AttrARN),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrFunctionARN, functionResourceName, names.AttrARN),
					resource.TestCheckResourceAttrPair(resourceName, "function_name", functionResourceName, names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, "function_response_types.#", acctest.Ct0),
					acctest.CheckResourceAttrRFC3339(resourceName, "last_modified"),
					resource.TestCheckResourceAttr(resourceName, "tumbling_window_in_seconds", acctest.Ct0),
				),
			},
			// batch_size became optional.  Ensure that if the user supplies the default
			// value, but then moves to not providing the value, that we don't consider this
			// a diff.
			{
				PlanOnly: true,
				Config:   testAccEventSourceMappingConfig_kinesisBatchSize(rName, "null"),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"last_modified"},
			},
			{
				Config: testAccEventSourceMappingConfig_kinesisUpdateFunctionName(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckEventSourceMappingExists(ctx, resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "batch_size", "200"),
					resource.TestCheckResourceAttr(resourceName, names.AttrEnabled, acctest.CtFalse),
					resource.TestCheckResourceAttrPair(resourceName, "event_source_arn", eventSourceResourceName, names.AttrARN),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrFunctionARN, functionResourceNameUpdated, names.AttrARN),
					resource.TestCheckResourceAttrPair(resourceName, "function_name", functionResourceNameUpdated, names.AttrARN),
				),
			},
		},
	})
}

func TestAccLambdaEventSourceMapping_SQS_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var conf lambda.GetEventSourceMappingOutput
	resourceName := "aws_lambda_event_source_mapping.test"
	functionResourceName := "aws_lambda_function.test"
	functionResourceNameUpdated := "aws_lambda_function.test_update"
	eventSourceResourceName := "aws_sqs_queue.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.LambdaServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckEventSourceMappingDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccEventSourceMappingConfig_sqsBatchSize(rName, acctest.Ct10),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckEventSourceMappingExists(ctx, resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "batch_size", acctest.Ct10),
					resource.TestCheckResourceAttr(resourceName, names.AttrEnabled, acctest.CtTrue),
					resource.TestCheckResourceAttrPair(resourceName, "event_source_arn", eventSourceResourceName, names.AttrARN),
					resource.TestCheckResourceAttrPair(resourceName, "function_name", functionResourceName, names.AttrARN),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrFunctionARN, functionResourceName, names.AttrARN),
					acctest.CheckResourceAttrRFC3339(resourceName, "last_modified"),
					resource.TestCheckResourceAttr(resourceName, "filter_criteria.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "scaling_config.#", acctest.Ct0),
				),
			},
			// batch_size became optional.  Ensure that if the user supplies the default
			// value, but then moves to not providing the value, that we don't consider this
			// a diff.
			{
				PlanOnly: true,
				Config:   testAccEventSourceMappingConfig_sqsBatchSize(rName, "null"),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"last_modified"},
			},
			{
				Config: testAccEventSourceMappingConfig_sqsUpdateFunctionName(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckEventSourceMappingExists(ctx, resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "batch_size", "5"),
					resource.TestCheckResourceAttr(resourceName, names.AttrEnabled, acctest.CtFalse),
					resource.TestCheckResourceAttrPair(resourceName, "event_source_arn", eventSourceResourceName, names.AttrARN),
					resource.TestCheckResourceAttrPair(resourceName, "function_name", functionResourceNameUpdated, names.AttrARN),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrFunctionARN, functionResourceNameUpdated, names.AttrARN),
					acctest.CheckResourceAttrRFC3339(resourceName, "last_modified"),
				),
			},
		},
	})
}

func TestAccLambdaEventSourceMapping_DynamoDB_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var conf lambda.GetEventSourceMappingOutput
	resourceName := "aws_lambda_event_source_mapping.test"
	functionResourceName := "aws_lambda_function.test"
	eventSourceResourceName := "aws_dynamodb_table.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.LambdaServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckEventSourceMappingDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccEventSourceMappingConfig_dynamoDBBatchSize(rName, "100"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckEventSourceMappingExists(ctx, resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "batch_size", "100"),
					resource.TestCheckResourceAttr(resourceName, names.AttrEnabled, acctest.CtTrue),
					resource.TestCheckResourceAttrPair(resourceName, "event_source_arn", eventSourceResourceName, names.AttrStreamARN),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrFunctionARN, functionResourceName, names.AttrARN),
					resource.TestCheckResourceAttrPair(resourceName, "function_name", functionResourceName, names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, "function_response_types.#", acctest.Ct0),
					acctest.CheckResourceAttrRFC3339(resourceName, "last_modified"),
					resource.TestCheckResourceAttr(resourceName, "tumbling_window_in_seconds", acctest.Ct0),
				),
			},
			// batch_size became optional.  Ensure that if the user supplies the default
			// value, but then moves to not providing the value, that we don't consider this
			// a diff.
			{
				PlanOnly: true,
				Config:   testAccEventSourceMappingConfig_dynamoDBBatchSize(rName, "null"),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"last_modified"},
			},
		},
	})
}

func TestAccLambdaEventSourceMapping_DynamoDB_functionResponseTypes(t *testing.T) {
	ctx := acctest.Context(t)
	var conf lambda.GetEventSourceMappingOutput
	resourceName := "aws_lambda_event_source_mapping.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.LambdaServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckEventSourceMappingDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccEventSourceMappingConfig_dynamoDBFunctionResponseTypes(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckEventSourceMappingExists(ctx, resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "function_response_types.#", acctest.Ct1),
					resource.TestCheckTypeSetElemAttr(resourceName, "function_response_types.*", "ReportBatchItemFailures"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"last_modified"},
			},
			{
				Config: testAccEventSourceMappingConfig_dynamoDBNoFunctionResponseTypes(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckEventSourceMappingExists(ctx, resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "function_response_types.#", acctest.Ct0),
				),
			},
		},
	})
}

func TestAccLambdaEventSourceMapping_DynamoDB_streamAdded(t *testing.T) {
	ctx := acctest.Context(t)
	var conf lambda.GetEventSourceMappingOutput
	resourceName := "aws_dynamodb_table.test"
	mappingResourceName := "aws_lambda_event_source_mapping.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.LambdaServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckEventSourceMappingDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccEventSourceMappingConfig_dynamoDBStreamBase(rName, false),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"last_modified"},
			},
			{
				Config: testAccEventSourceMappingConfig_dynamoDBStreamEnabled(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckEventSourceMappingExists(ctx, mappingResourceName, &conf),
				),
			},
		},
	})
}

func TestAccLambdaEventSourceMapping_SQS_batchWindow(t *testing.T) {
	ctx := acctest.Context(t)
	var conf lambda.GetEventSourceMappingOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_lambda_event_source_mapping.test"
	batchWindow := int64(0)
	batchWindowUpdate := int64(100)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.LambdaServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckEventSourceMappingDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccEventSourceMappingConfig_sqsBatchWindow(rName, batchWindow),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckEventSourceMappingExists(ctx, resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "maximum_batching_window_in_seconds", strconv.Itoa(int(batchWindow))),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"last_modified"},
			},
			{
				Config: testAccEventSourceMappingConfig_sqsBatchWindow(rName, batchWindowUpdate),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckEventSourceMappingExists(ctx, resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "maximum_batching_window_in_seconds", strconv.Itoa(int(batchWindowUpdate))),
				),
			},
		},
	})
}

func TestAccLambdaEventSourceMapping_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	var conf lambda.GetEventSourceMappingOutput
	resourceName := "aws_lambda_event_source_mapping.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.LambdaServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckEventSourceMappingDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccEventSourceMappingConfig_sqsBatchSize(rName, "7"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckEventSourceMappingExists(ctx, resourceName, &conf),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tflambda.ResourceEventSourceMapping(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccLambdaEventSourceMapping_SQS_changesInEnabledAreDetected(t *testing.T) {
	ctx := acctest.Context(t)
	var conf lambda.GetEventSourceMappingOutput
	resourceName := "aws_lambda_event_source_mapping.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.LambdaServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckEventSourceMappingDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccEventSourceMappingConfig_sqsBatchSize(rName, "9"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckEventSourceMappingExists(ctx, resourceName, &conf),
					testAccCheckEventSourceMappingIsBeingDisabled(ctx, &conf),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccLambdaEventSourceMapping_Kinesis_startingPositionTimestamp(t *testing.T) {
	ctx := acctest.Context(t)
	var conf lambda.GetEventSourceMappingOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_lambda_event_source_mapping.test"
	startingPositionTimestamp := time.Now().UTC().Format(time.RFC3339)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.LambdaServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckEventSourceMappingDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccEventSourceMappingConfig_kinesisStartingPositionTimestamp(rName, startingPositionTimestamp),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckEventSourceMappingExists(ctx, resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "starting_position", "AT_TIMESTAMP"),
					resource.TestCheckResourceAttr(resourceName, "starting_position_timestamp", startingPositionTimestamp),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"last_modified"},
			},
		},
	})
}

func TestAccLambdaEventSourceMapping_Kinesis_batchWindow(t *testing.T) {
	ctx := acctest.Context(t)
	var conf lambda.GetEventSourceMappingOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_lambda_event_source_mapping.test"
	batchWindow := int64(5)
	batchWindowUpdate := int64(10)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.LambdaServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckEventSourceMappingDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccEventSourceMappingConfig_kinesisBatchWindow(rName, batchWindow),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckEventSourceMappingExists(ctx, resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "maximum_batching_window_in_seconds", strconv.Itoa(int(batchWindow))),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"last_modified"},
			},
			{
				Config: testAccEventSourceMappingConfig_kinesisBatchWindow(rName, batchWindowUpdate),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckEventSourceMappingExists(ctx, resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "maximum_batching_window_in_seconds", strconv.Itoa(int(batchWindowUpdate))),
				),
			},
		},
	})
}

func TestAccLambdaEventSourceMapping_Kinesis_parallelizationFactor(t *testing.T) {
	ctx := acctest.Context(t)
	var conf lambda.GetEventSourceMappingOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_lambda_event_source_mapping.test"
	parallelizationFactor := int64(1)
	parallelizationFactorUpdate := int64(5)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.LambdaServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckEventSourceMappingDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccEventSourceMappingConfig_kinesisParallelizationFactor(rName, parallelizationFactor),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckEventSourceMappingExists(ctx, resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "parallelization_factor", strconv.Itoa(int(parallelizationFactor))),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"last_modified"},
			},
			{
				Config: testAccEventSourceMappingConfig_kinesisParallelizationFactor(rName, parallelizationFactorUpdate),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckEventSourceMappingExists(ctx, resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "parallelization_factor", strconv.Itoa(int(parallelizationFactorUpdate))),
				),
			},
		},
	})
}

func TestAccLambdaEventSourceMapping_Kinesis_tumblingWindowInSeconds(t *testing.T) {
	ctx := acctest.Context(t)
	var conf lambda.GetEventSourceMappingOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_lambda_event_source_mapping.test"
	tumblingWindowInSeconds := int64(30)
	tumblingWindowInSecondsUpdate := int64(300)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.LambdaServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckEventSourceMappingDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccEventSourceMappingConfig_kinesisTumblingWindowInSeconds(rName, tumblingWindowInSeconds),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckEventSourceMappingExists(ctx, resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "tumbling_window_in_seconds", strconv.Itoa(int(tumblingWindowInSeconds))),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"last_modified"},
			},
			{
				Config: testAccEventSourceMappingConfig_kinesisTumblingWindowInSeconds(rName, tumblingWindowInSecondsUpdate),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckEventSourceMappingExists(ctx, resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "tumbling_window_in_seconds", strconv.Itoa(int(tumblingWindowInSecondsUpdate))),
				),
			},
		},
	})
}

func TestAccLambdaEventSourceMapping_Kinesis_maximumRetryAttempts(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var conf lambda.GetEventSourceMappingOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_lambda_event_source_mapping.test"
	maximumRetryAttempts := int64(10000)
	maximumRetryAttemptsUpdate := int64(100)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.LambdaServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckEventSourceMappingDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccEventSourceMappingConfig_kinesisMaximumRetryAttempts(rName, maximumRetryAttempts),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckEventSourceMappingExists(ctx, resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "maximum_retry_attempts", strconv.Itoa(int(maximumRetryAttempts))),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"last_modified"},
			},
			{
				Config: testAccEventSourceMappingConfig_kinesisMaximumRetryAttempts(rName, maximumRetryAttemptsUpdate),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckEventSourceMappingExists(ctx, resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "maximum_retry_attempts", strconv.Itoa(int(maximumRetryAttemptsUpdate))),
				),
			},
		},
	})
}

func TestAccLambdaEventSourceMapping_Kinesis_maximumRetryAttemptsZero(t *testing.T) {
	ctx := acctest.Context(t)
	var conf lambda.GetEventSourceMappingOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_lambda_event_source_mapping.test"
	maximumRetryAttempts := int64(0)
	maximumRetryAttemptsUpdate := int64(100)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.LambdaServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckEventSourceMappingDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccEventSourceMappingConfig_kinesisMaximumRetryAttempts(rName, maximumRetryAttempts),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckEventSourceMappingExists(ctx, resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "maximum_retry_attempts", strconv.Itoa(int(maximumRetryAttempts))),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"last_modified"},
			},
			{
				Config: testAccEventSourceMappingConfig_kinesisMaximumRetryAttempts(rName, maximumRetryAttemptsUpdate),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckEventSourceMappingExists(ctx, resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "maximum_retry_attempts", strconv.Itoa(int(maximumRetryAttemptsUpdate))),
				),
			},
			{
				Config: testAccEventSourceMappingConfig_kinesisMaximumRetryAttempts(rName, maximumRetryAttempts),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckEventSourceMappingExists(ctx, resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "maximum_retry_attempts", strconv.Itoa(int(maximumRetryAttempts))),
				),
			},
		},
	})
}

func TestAccLambdaEventSourceMapping_Kinesis_maximumRetryAttemptsNegativeOne(t *testing.T) {
	ctx := acctest.Context(t)
	var conf lambda.GetEventSourceMappingOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_lambda_event_source_mapping.test"
	maximumRetryAttempts := int64(-1)
	maximumRetryAttemptsUpdate := int64(100)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.LambdaServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckEventSourceMappingDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccEventSourceMappingConfig_kinesisMaximumRetryAttempts(rName, maximumRetryAttempts),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckEventSourceMappingExists(ctx, resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "maximum_retry_attempts", strconv.Itoa(int(maximumRetryAttempts))),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"last_modified"},
			},
			{
				Config: testAccEventSourceMappingConfig_kinesisMaximumRetryAttempts(rName, maximumRetryAttemptsUpdate),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckEventSourceMappingExists(ctx, resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "maximum_retry_attempts", strconv.Itoa(int(maximumRetryAttemptsUpdate))),
				),
			},
			{
				Config: testAccEventSourceMappingConfig_kinesisMaximumRetryAttempts(rName, maximumRetryAttempts),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckEventSourceMappingExists(ctx, resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "maximum_retry_attempts", strconv.Itoa(int(maximumRetryAttempts))),
				),
			},
		},
	})
}

func TestAccLambdaEventSourceMapping_Kinesis_maximumRecordAgeInSeconds(t *testing.T) {
	ctx := acctest.Context(t)
	var conf lambda.GetEventSourceMappingOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_lambda_event_source_mapping.test"
	maximumRecordAgeInSeconds := int64(604800)
	maximumRecordAgeInSecondsUpdate := int64(3600)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.LambdaServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckEventSourceMappingDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccEventSourceMappingConfig_kinesisMaximumRecordAgeInSeconds(rName, maximumRecordAgeInSeconds),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckEventSourceMappingExists(ctx, resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "maximum_record_age_in_seconds", strconv.Itoa(int(maximumRecordAgeInSeconds))),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"last_modified"},
			},
			{
				Config: testAccEventSourceMappingConfig_kinesisMaximumRecordAgeInSeconds(rName, maximumRecordAgeInSecondsUpdate),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckEventSourceMappingExists(ctx, resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "maximum_record_age_in_seconds", strconv.Itoa(int(maximumRecordAgeInSecondsUpdate))),
				),
			},
		},
	})
}

func TestAccLambdaEventSourceMapping_Kinesis_maximumRecordAgeInSecondsNegativeOne(t *testing.T) {
	ctx := acctest.Context(t)
	var conf lambda.GetEventSourceMappingOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_lambda_event_source_mapping.test"
	maximumRecordAgeInSeconds := int64(-1)
	maximumRecordAgeInSecondsUpdate := int64(3600)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.LambdaServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckEventSourceMappingDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccEventSourceMappingConfig_kinesisMaximumRecordAgeInSeconds(rName, maximumRecordAgeInSeconds),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckEventSourceMappingExists(ctx, resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "maximum_record_age_in_seconds", strconv.Itoa(int(maximumRecordAgeInSeconds))),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"last_modified"},
			},
			{
				Config: testAccEventSourceMappingConfig_kinesisMaximumRecordAgeInSeconds(rName, maximumRecordAgeInSecondsUpdate),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckEventSourceMappingExists(ctx, resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "maximum_record_age_in_seconds", strconv.Itoa(int(maximumRecordAgeInSecondsUpdate))),
				),
			},
		},
	})
}

func TestAccLambdaEventSourceMapping_Kinesis_bisectBatch(t *testing.T) {
	ctx := acctest.Context(t)
	var conf lambda.GetEventSourceMappingOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_lambda_event_source_mapping.test"
	bisectBatch := false
	bisectBatchUpdate := true

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.LambdaServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckEventSourceMappingDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccEventSourceMappingConfig_kinesisBisectBatch(rName, bisectBatch),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckEventSourceMappingExists(ctx, resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "bisect_batch_on_function_error", strconv.FormatBool(bisectBatch)),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"last_modified"},
			},
			{
				Config: testAccEventSourceMappingConfig_kinesisBisectBatch(rName, bisectBatchUpdate),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckEventSourceMappingExists(ctx, resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "bisect_batch_on_function_error", strconv.FormatBool(bisectBatchUpdate)),
				),
			},
		},
	})
}

func TestAccLambdaEventSourceMapping_Kinesis_destination(t *testing.T) {
	ctx := acctest.Context(t)
	var conf lambda.GetEventSourceMappingOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_lambda_event_source_mapping.test"
	snsResourceName := "aws_sns_topic.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.LambdaServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckEventSourceMappingDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccEventSourceMappingConfig_kinesisDestination(rName, rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckEventSourceMappingExists(ctx, resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "destination_config.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "destination_config.0.on_failure.#", acctest.Ct1),
					resource.TestCheckResourceAttrPair(resourceName, "destination_config.0.on_failure.0.destination_arn", snsResourceName, names.AttrARN),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"last_modified"},
			},
			{
				Config: testAccEventSourceMappingConfig_kinesisDestination(rName, rName+"-update"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckEventSourceMappingExists(ctx, resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "destination_config.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "destination_config.0.on_failure.#", acctest.Ct1),
					resource.TestCheckResourceAttrPair(resourceName, "destination_config.0.on_failure.0.destination_arn", snsResourceName, names.AttrARN),
				),
			},
		},
	})
}

func TestAccLambdaEventSourceMapping_msk(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var v lambda.GetEventSourceMappingOutput
	resourceName := "aws_lambda_event_source_mapping.test"
	eventSourceResourceName := "aws_msk_cluster.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheckMSK(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.LambdaEndpointID, "kafka"),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckEventSourceMappingDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccEventSourceMappingConfig_msk(rName, "100"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckEventSourceMappingExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "batch_size", "100"),
					resource.TestCheckResourceAttrPair(resourceName, "event_source_arn", eventSourceResourceName, names.AttrARN),
					acctest.CheckResourceAttrRFC3339(resourceName, "last_modified"),
					resource.TestCheckResourceAttr(resourceName, "amazon_managed_kafka_event_source_config.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "topics.#", acctest.Ct1),
					resource.TestCheckTypeSetElemAttr(resourceName, "topics.*", "test"),
				),
			},
			// batch_size became optional.  Ensure that if the user supplies the default
			// value, but then moves to not providing the value, that we don't consider this
			// a diff.
			{
				PlanOnly: true,
				Config:   testAccEventSourceMappingConfig_msk(rName, "null"),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"last_modified"},
			},
			{
				Config: testAccEventSourceMappingConfig_msk(rName, "9999"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckEventSourceMappingExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "batch_size", "9999"),
					resource.TestCheckResourceAttrPair(resourceName, "event_source_arn", eventSourceResourceName, names.AttrARN),
					acctest.CheckResourceAttrRFC3339(resourceName, "last_modified"),
					resource.TestCheckResourceAttr(resourceName, "amazon_managed_kafka_event_source_config.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "topics.#", acctest.Ct1),
					resource.TestCheckTypeSetElemAttr(resourceName, "topics.*", "test"),
				),
			},
		},
	})
}

func TestAccLambdaEventSourceMapping_mskWithEventSourceConfig(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var v lambda.GetEventSourceMappingOutput
	resourceName := "aws_lambda_event_source_mapping.test"
	eventSourceResourceName := "aws_msk_cluster.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheckMSK(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.LambdaEndpointID, "kafka"),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckEventSourceMappingDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccEventSourceMappingConfig_mskWithEventSourceConfig(rName, "100"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckEventSourceMappingExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "batch_size", "100"),
					resource.TestCheckResourceAttrPair(resourceName, "event_source_arn", eventSourceResourceName, names.AttrARN),
					acctest.CheckResourceAttrRFC3339(resourceName, "last_modified"),
					resource.TestCheckResourceAttr(resourceName, "amazon_managed_kafka_event_source_config.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "amazon_managed_kafka_event_source_config.0.consumer_group_id", "amazon-managed-test-group-id"),
					resource.TestCheckResourceAttr(resourceName, "topics.#", acctest.Ct1),
					resource.TestCheckTypeSetElemAttr(resourceName, "topics.*", "test"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"last_modified"},
			},
		},
	})
}

func TestAccLambdaEventSourceMapping_selfManagedKafka(t *testing.T) {
	ctx := acctest.Context(t)
	var v lambda.GetEventSourceMappingOutput
	resourceName := "aws_lambda_event_source_mapping.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.LambdaServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckEventSourceMappingDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccEventSourceMappingConfig_selfManagedKafka(rName, "100", "test1:9092,test2:9092"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckEventSourceMappingExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "batch_size", "100"),
					resource.TestCheckResourceAttr(resourceName, names.AttrEnabled, acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "self_managed_event_source.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "self_managed_event_source.0.endpoints.KAFKA_BOOTSTRAP_SERVERS", "test1:9092,test2:9092"),
					resource.TestCheckResourceAttr(resourceName, "self_managed_kafka_event_source_config.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "source_access_configuration.#", acctest.Ct3),
					acctest.CheckResourceAttrRFC3339(resourceName, "last_modified"),
					resource.TestCheckResourceAttr(resourceName, "topics.#", acctest.Ct1),
					resource.TestCheckTypeSetElemAttr(resourceName, "topics.*", "test"),
				),
			},
			// batch_size became optional.  Ensure that if the user supplies the default
			// value, but then moves to not providing the value, that we don't consider this
			// a diff.
			// Verify also that bootstrap broker order does not matter.
			{
				PlanOnly: true,
				Config:   testAccEventSourceMappingConfig_selfManagedKafka(rName, "null", "test2:9092,test1:9092"),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"last_modified"},
			},
		},
	})
}

func TestAccLambdaEventSourceMapping_selfManagedKafkaWithEventSourceConfig(t *testing.T) {
	ctx := acctest.Context(t)
	var v lambda.GetEventSourceMappingOutput
	resourceName := "aws_lambda_event_source_mapping.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.LambdaServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckEventSourceMappingDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccEventSourceMappingConfig_selfManagedKafkaWithEventSourceConfig(rName, "100", "test1:9092,test2:9092"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckEventSourceMappingExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "batch_size", "100"),
					resource.TestCheckResourceAttr(resourceName, names.AttrEnabled, acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "self_managed_event_source.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "self_managed_event_source.0.endpoints.KAFKA_BOOTSTRAP_SERVERS", "test1:9092,test2:9092"),
					resource.TestCheckResourceAttr(resourceName, "self_managed_kafka_event_source_config.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "self_managed_kafka_event_source_config.0.consumer_group_id", "self-managed-test-group-id"),
					resource.TestCheckResourceAttr(resourceName, "source_access_configuration.#", acctest.Ct3),
					acctest.CheckResourceAttrRFC3339(resourceName, "last_modified"),
					resource.TestCheckResourceAttr(resourceName, "topics.#", acctest.Ct1),
					resource.TestCheckTypeSetElemAttr(resourceName, "topics.*", "test"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"last_modified"},
			},
		},
	})
}

func TestAccLambdaEventSourceMapping_activeMQ(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var v lambda.GetEventSourceMappingOutput
	resourceName := "aws_lambda_event_source_mapping.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheckSecretsManager(ctx, t)
			acctest.PreCheckPartitionHasService(t, "mq")
			testAccPreCheckMQ(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.LambdaEndpointID, "mq", "secretsmanager"),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckEventSourceMappingDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccEventSourceMappingConfig_activeMQ(rName, "100"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckEventSourceMappingExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "batch_size", "100"),
					resource.TestCheckResourceAttr(resourceName, names.AttrEnabled, acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "queues.#", acctest.Ct1),
					resource.TestCheckTypeSetElemAttr(resourceName, "queues.*", "test"),
					resource.TestCheckResourceAttr(resourceName, "source_access_configuration.#", acctest.Ct1),
				),
			},
			// batch_size became optional.  Ensure that if the user supplies the default
			// value, but then moves to not providing the value, that we don't consider this
			// a diff.
			{
				PlanOnly: true,
				Config:   testAccEventSourceMappingConfig_activeMQ(rName, "null"),
			},
		},
	})
}

func TestAccLambdaEventSourceMapping_rabbitMQ(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var v lambda.GetEventSourceMappingOutput
	resourceName := "aws_lambda_event_source_mapping.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheckSecretsManager(ctx, t)
			acctest.PreCheckPartitionHasService(t, "mq")
			testAccPreCheckMQ(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.LambdaEndpointID, "mq", "secretsmanager"),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckEventSourceMappingDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccEventSourceMappingConfig_rabbitMQ(rName, "100"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckEventSourceMappingExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "batch_size", "100"),
					resource.TestCheckResourceAttr(resourceName, names.AttrEnabled, acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "queues.#", acctest.Ct1),
					resource.TestCheckTypeSetElemAttr(resourceName, "queues.*", "test"),
					resource.TestCheckResourceAttr(resourceName, "source_access_configuration.#", acctest.Ct2),
				),
			},
			// batch_size became optional.  Ensure that if the user supplies the default
			// value, but then moves to not providing the value, that we don't consider this
			// a diff.
			{
				PlanOnly: true,
				Config:   testAccEventSourceMappingConfig_rabbitMQ(rName, "null"),
			},
		},
	})
}

func TestAccLambdaEventSourceMapping_SQS_filterCriteria(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var conf lambda.GetEventSourceMappingOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_lambda_event_source_mapping.test"
	pattern1 := "{\"Region\": [{\"prefix\": \"us-\"}]}"
	pattern2 := "{\"Location\": [\"New York\"], \"Day\": [\"Monday\"]}"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.LambdaServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckEventSourceMappingDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccEventSourceMappingConfig_sqsFilterCriteria1(rName, pattern1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckEventSourceMappingExists(ctx, resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "filter_criteria.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "filter_criteria.0.filter.#", acctest.Ct1),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "filter_criteria.0.filter.*", map[string]string{"pattern": pattern1}),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"last_modified"},
			},
			{
				Config: testAccEventSourceMappingConfig_sqsFilterCriteria2(rName, pattern1, pattern2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckEventSourceMappingExists(ctx, resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "filter_criteria.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "filter_criteria.0.filter.#", acctest.Ct2),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "filter_criteria.0.filter.*", map[string]string{"pattern": pattern1}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "filter_criteria.0.filter.*", map[string]string{"pattern": pattern2}),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"last_modified"},
			},
			{
				Config: testAccEventSourceMappingConfig_sqsFilterCriteria3(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckEventSourceMappingExists(ctx, resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "filter_criteria.#", acctest.Ct0),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"last_modified"},
			},
			{
				Config: testAccEventSourceMappingConfig_sqsFilterCriteria1(rName, pattern1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckEventSourceMappingExists(ctx, resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "filter_criteria.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "filter_criteria.0.filter.#", acctest.Ct1),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "filter_criteria.0.filter.*", map[string]string{"pattern": pattern1}),
				),
			},
		},
	})
}

func TestAccLambdaEventSourceMapping_SQS_scalingConfig(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var conf lambda.GetEventSourceMappingOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_lambda_event_source_mapping.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.LambdaServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckEventSourceMappingDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccEventSourceMappingConfig_sqsScalingConfig1(rName, 10),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckEventSourceMappingExists(ctx, resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "scaling_config.0.maximum_concurrency", acctest.Ct10),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"last_modified"},
			},
			{
				Config: testAccEventSourceMappingConfig_sqsScalingConfig1(rName, 15),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckEventSourceMappingExists(ctx, resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "scaling_config.0.maximum_concurrency", "15"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"last_modified"},
			},
			{
				Config: testAccEventSourceMappingConfig_sqsScalingConfig2(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckEventSourceMappingExists(ctx, resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "scaling_config.#", acctest.Ct0),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"last_modified"},
			},
		},
	})
}

func TestAccLambdaEventSourceMapping_documentDB(t *testing.T) {
	ctx := acctest.Context(t)
	var v lambda.GetEventSourceMappingOutput
	resourceName := "aws_lambda_event_source_mapping.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheckSecretsManager(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.LambdaServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckEventSourceMappingDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccEventSourceMappingConfig_documentDB(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckEventSourceMappingExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "batch_size", "100"),
					resource.TestCheckResourceAttr(resourceName, "document_db_event_source_config.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "document_db_event_source_config.0.collection_name", ""),
					resource.TestCheckResourceAttr(resourceName, "document_db_event_source_config.0.database_name", "test"),
					resource.TestCheckResourceAttr(resourceName, "document_db_event_source_config.0.full_document", "Default"),
					resource.TestCheckResourceAttr(resourceName, names.AttrEnabled, acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "source_access_configuration.#", acctest.Ct1),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"last_modified", "last_processing_result"},
			},
		},
	})
}

func testAccPreCheckMQ(ctx context.Context, t *testing.T) {
	conn := acctest.Provider.Meta().(*conns.AWSClient).MQClient(ctx)

	input := &mq.ListBrokersInput{}

	_, err := conn.ListBrokers(ctx, input)

	if acctest.PreCheckSkipError(err) {
		t.Skipf("skipping acceptance testing: %s", err)
	}

	if err != nil {
		t.Fatalf("unexpected PreCheck error: %s", err)
	}
}

func testAccPreCheckMSK(ctx context.Context, t *testing.T) {
	conn := acctest.Provider.Meta().(*conns.AWSClient).KafkaClient(ctx)

	input := &kafka.ListClustersInput{}

	_, err := conn.ListClusters(ctx, input)

	if acctest.PreCheckSkipError(err) {
		t.Skipf("skipping acceptance testing: %s", err)
	}

	if err != nil {
		t.Fatalf("unexpected PreCheck error: %s", err)
	}
}

func testAccPreCheckSecretsManager(ctx context.Context, t *testing.T) {
	conn := acctest.Provider.Meta().(*conns.AWSClient).SecretsManagerClient(ctx)

	input := &secretsmanager.ListSecretsInput{}

	_, err := conn.ListSecrets(ctx, input)

	if acctest.PreCheckSkipError(err) {
		t.Skipf("skipping acceptance testing: %s", err)
	}

	if err != nil {
		t.Fatalf("unexpected PreCheck error: %s", err)
	}
}

func testAccCheckEventSourceMappingIsBeingDisabled(ctx context.Context, v *lambda.GetEventSourceMappingOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).LambdaClient(ctx)
		// Disable enabled state
		err := retry.RetryContext(ctx, 10*time.Minute, func() *retry.RetryError {
			input := &lambda.UpdateEventSourceMappingInput{
				Enabled: aws.Bool(false),
				UUID:    v.UUID,
			}

			_, err := conn.UpdateEventSourceMapping(ctx, input)

			if errs.IsA[*awstypes.ResourceInUseException](err) {
				return retry.RetryableError(fmt.Errorf(
					"Waiting for Lambda Event Source Mapping to be ready to be updated: %v", v.UUID))
			}

			if err != nil {
				return retry.NonRetryableError(
					fmt.Errorf("Error updating Lambda Event Source Mapping: %w", err))
			}

			return nil
		})

		if err != nil {
			return err
		}

		// wait for state to be propagated
		return retry.RetryContext(ctx, 10*time.Minute, func() *retry.RetryError {
			output, err := tflambda.FindEventSourceMappingByID(ctx, conn, aws.ToString(v.UUID))

			if err != nil {
				return retry.NonRetryableError(
					fmt.Errorf("Error getting Lambda Event Source Mapping: %s", err))
			}

			if state := aws.ToString(output.State); state != "Disabled" {
				return retry.RetryableError(fmt.Errorf(
					"Waiting to get Lambda Event Source Mapping to be fully enabled, it's currently %s: %v", state, v.UUID))
			}

			return nil
		})
	}
}

func testAccCheckEventSourceMappingDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).LambdaClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_lambda_event_source_mapping" {
				continue
			}

			_, err := tflambda.FindEventSourceMappingByID(ctx, conn, rs.Primary.ID)

			if tfresource.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("Lambda Event Source Mapping %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccCheckEventSourceMappingExists(ctx context.Context, n string, v *lambda.GetEventSourceMappingOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf(" Lambda Event Source Mapping resource not found: %s", n)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).LambdaClient(ctx)

		output, err := tflambda.FindEventSourceMappingByID(ctx, conn, rs.Primary.ID)

		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

func testAccEventSourceMappingConfig_kinesisBase(rName string) string {
	return fmt.Sprintf(`
resource "aws_iam_role" "test" {
  name = %[1]q

  assume_role_policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [{
    "Action": "sts:AssumeRole",
    "Principal": {
      "Service": "lambda.amazonaws.com"
    },
    "Effect": "Allow"
  }]
}
EOF
}

resource "aws_iam_role_policy" "test" {
  role = aws_iam_role.test.name

  policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Action": [
        "kinesis:*"
      ],
      "Resource": "*"
    },
    {
      "Effect": "Allow",
      "Action": [
        "sns:*"
      ],
      "Resource": "*"
    }
  ]
}
EOF
}

resource "aws_kinesis_stream" "test" {
  name        = %[1]q
  shard_count = 1
}

resource "aws_lambda_function" "test" {
  filename      = "test-fixtures/lambdatest.zip"
  function_name = %[1]q
  handler       = "exports.example"
  role          = aws_iam_role.test.arn
  runtime       = "nodejs16.x"
}
`, rName)
}

func testAccEventSourceMappingConfig_sqsBase(rName string) string {
	return fmt.Sprintf(`
resource "aws_iam_role" "test" {
  name = %[1]q

  assume_role_policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [{
    "Action": "sts:AssumeRole",
    "Principal": {
      "Service": "lambda.amazonaws.com"
    },
    "Effect": "Allow"
  }]
}
EOF
}

resource "aws_iam_role_policy" "test" {
  role = aws_iam_role.test.name

  policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Action": [
        "sqs:*"
      ],
      "Resource": "*"
    }
  ]
}
EOF
}

resource "aws_sqs_queue" "test" {
  name = %[1]q
}

resource "aws_lambda_function" "test" {
  filename      = "test-fixtures/lambdatest.zip"
  function_name = %[1]q
  handler       = "exports.example"
  role          = aws_iam_role.test.arn
  runtime       = "nodejs16.x"
}
`, rName)
}

func testAccEventSourceMappingConfig_dynamoDBBase(rName string) string {
	return fmt.Sprintf(`
resource "aws_iam_role" "test" {
  name = %[1]q

  assume_role_policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [{
    "Action": "sts:AssumeRole",
    "Principal": {
      "Service": "lambda.amazonaws.com"
    },
    "Effect": "Allow"
  }]
}
EOF
}

resource "aws_iam_role_policy" "test" {
  role = aws_iam_role.test.name

  policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Action": [
        "dynamodb:*"
      ],
      "Resource": "*"
    }
  ]
}
EOF
}

resource "aws_dynamodb_table" "test" {
  name           = %[1]q
  read_capacity  = 1
  write_capacity = 2
  hash_key       = "TestTableHashKey"

  attribute {
    name = "TestTableHashKey"
    type = "S"
  }

  stream_enabled   = true
  stream_view_type = "KEYS_ONLY"
}

resource "aws_lambda_function" "test" {
  filename      = "test-fixtures/lambdatest.zip"
  function_name = %[1]q
  handler       = "exports.example"
  role          = aws_iam_role.test.arn
  runtime       = "nodejs16.x"
}
`, rName)
}

func testAccEventSourceMappingConfig_dynamoDBStreamBase(rName string, streamEnabled bool) string {
	var streamStatus string
	var streamViewType string
	if streamEnabled {
		streamStatus = "stream_enabled   = true"
		streamViewType = "stream_view_type = \"KEYS_ONLY\""
	} else {
		streamStatus = ""
		streamViewType = ""
	}

	return fmt.Sprintf(`
resource "aws_iam_role" "test" {
  name = %[1]q

  assume_role_policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [{
    "Action": "sts:AssumeRole",
    "Principal": {
      "Service": "lambda.amazonaws.com"
    },
    "Effect": "Allow"
  }]
}
EOF
}

resource "aws_iam_role_policy" "test" {
  role = aws_iam_role.test.name

  policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Action": [
        "dynamodb:*"
      ],
      "Resource": "*"
    }
  ]
}
EOF
}

resource "aws_dynamodb_table" "test" {
  name           = %[1]q
  read_capacity  = 1
  write_capacity = 2
  hash_key       = "TestTableHashKey"

  attribute {
    name = "TestTableHashKey"
    type = "S"
  }
  %[2]s
  %[3]s
}

resource "aws_lambda_function" "test" {
  filename      = "test-fixtures/lambdatest.zip"
  function_name = %[1]q
  handler       = "exports.example"
  role          = aws_iam_role.test.arn
  runtime       = "nodejs18.x"
}
`, rName, streamStatus, streamViewType)
}

func testAccEventSourceMappingConfig_kafkaBase(rName string) string {
	return acctest.ConfigCompose(acctest.ConfigVPCWithSubnets(rName, 2), fmt.Sprintf(`
resource "aws_iam_role" "test" {
  name = %[1]q

  assume_role_policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [{
    "Action": "sts:AssumeRole",
    "Principal": {
      "Service": "lambda.amazonaws.com"
    },
    "Effect": "Allow"
  }]
}
EOF
}

resource "aws_iam_policy" "test" {
  name = %[1]q

  policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Action": [
        "kafka:DescribeCluster",
        "kafka:GetBootstrapBrokers",
        "ec2:CreateNetworkInterface",
        "ec2:DeleteNetworkInterface",
        "ec2:DescribeNetworkInterfaces",
        "ec2:DescribeSecurityGroups",
        "ec2:DescribeSubnets",
        "ec2:DescribeVpcs",
        "logs:CreateLogGroup",
        "logs:CreateLogStream",
        "logs:PutLogEvents"
      ],
      "Resource": "*"
    }
  ]
}
EOF
}

resource "aws_iam_policy_attachment" "test" {
  name       = %[1]q
  roles      = [aws_iam_role.test.name]
  policy_arn = aws_iam_policy.test.arn
}

resource "aws_security_group" "test" {
  name   = %[1]q
  vpc_id = aws_vpc.test.id

  tags = {
    Name = %[1]q
  }
}

resource "aws_lambda_function" "test" {
  filename      = "test-fixtures/lambdatest.zip"
  function_name = %[1]q
  handler       = "exports.example"
  role          = aws_iam_role.test.arn
  runtime       = "nodejs16.x"
}
`, rName))
}

func testAccEventSourceMappingConfig_activeMQBase(rName string) string {
	return fmt.Sprintf(`
resource "aws_iam_role" "test" {
  name = %[1]q

  assume_role_policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [{
    "Action": "sts:AssumeRole",
    "Principal": {
      "Service": "lambda.amazonaws.com"
    },
    "Effect": "Allow"
  }]
}
EOF
}

resource "aws_iam_role_policy" "test" {
  role = aws_iam_role.test.name

  policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Action": [
        "mq:DescribeBroker",
        "secretsmanager:GetSecretValue",
        "ec2:CreateNetworkInterface",
        "ec2:DescribeNetworkInterfaces",
        "ec2:DescribeVpcs",
        "ec2:DeleteNetworkInterface",
        "ec2:DescribeSubnets",
        "ec2:DescribeSecurityGroups",
        "logs:CreateLogGroup",
        "logs:CreateLogStream",
        "logs:PutLogEvents"
      ],
      "Resource": "*"
    }
  ]
}
EOF
}

resource "aws_lambda_function" "test" {
  filename      = "test-fixtures/lambdatest.zip"
  function_name = %[1]q
  handler       = "exports.example"
  role          = aws_iam_role.test.arn
  runtime       = "nodejs16.x"
}

resource "aws_security_group" "test" {
  name = %[1]q

  ingress {
    from_port   = 61617
    to_port     = 61617
    protocol    = "tcp"
    cidr_blocks = ["0.0.0.0/0"]
  }

  tags = {
    Name = %[1]q
  }
}

resource "aws_mq_broker" "test" {
  broker_name             = %[1]q
  engine_type             = "ActiveMQ"
  engine_version          = "5.17.6"
  host_instance_type      = "mq.t2.micro"
  security_groups         = [aws_security_group.test.id]
  authentication_strategy = "simple"
  storage_type            = "efs"

  logs {
    general = true
  }

  user {
    username = "Test"
    password = "TestTest1234"
  }

  publicly_accessible = true
}

resource "aws_secretsmanager_secret" "test" {
  name = %[1]q
}

resource "aws_secretsmanager_secret_version" "test" {
  secret_id     = aws_secretsmanager_secret.test.id
  secret_string = jsonencode({ username = "Test", password = "TestTest1234" })
}
`, rName)
}

func testAccEventSourceMappingConfig_rabbitMQBase(rName string) string {
	return fmt.Sprintf(`
resource "aws_iam_role" "test" {
  name = %[1]q

  assume_role_policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [{
    "Action": "sts:AssumeRole",
    "Principal": {
      "Service": "lambda.amazonaws.com"
    },
    "Effect": "Allow"
  }]
}
EOF
}

resource "aws_iam_role_policy" "test" {
  role = aws_iam_role.test.name

  policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Action": [
        "mq:DescribeBroker",
        "secretsmanager:GetSecretValue",
        "ec2:CreateNetworkInterface",
        "ec2:DescribeNetworkInterfaces",
        "ec2:DescribeVpcs",
        "ec2:DeleteNetworkInterface",
        "ec2:DescribeSubnets",
        "ec2:DescribeSecurityGroups",
        "logs:CreateLogGroup",
        "logs:CreateLogStream",
        "logs:PutLogEvents"
      ],
      "Resource": "*"
    }
  ]
}
EOF
}

resource "aws_lambda_function" "test" {
  filename      = "test-fixtures/lambdatest.zip"
  function_name = %[1]q
  handler       = "exports.example"
  role          = aws_iam_role.test.arn
  runtime       = "nodejs16.x"
}

resource "aws_mq_broker" "test" {
  broker_name             = %[1]q
  engine_type             = "RabbitMQ"
  engine_version          = "3.8.11"
  host_instance_type      = "mq.t3.micro"
  authentication_strategy = "simple"

  logs {
    general = true
  }

  user {
    username = "Test"
    password = "TestTest1234"
  }

  publicly_accessible = true
}

resource "aws_secretsmanager_secret" "test" {
  name = %[1]q
}

resource "aws_secretsmanager_secret_version" "test" {
  secret_id     = aws_secretsmanager_secret.test.id
  secret_string = jsonencode({ username = "Test", password = "TestTest1234" })
}
`, rName)
}

func testAccEventSourceMappingConfig_documentDBBase(rName string) string {
	return acctest.ConfigCompose(acctest.ConfigVPCWithSubnets(rName, 2), fmt.Sprintf(`
resource "aws_iam_role" "test" {
  name = %[1]q

  assume_role_policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [{
    "Action": "sts:AssumeRole",
    "Principal": {
      "Service": "lambda.amazonaws.com"
    },
    "Effect": "Allow"
  }]
}
EOF
}

resource "aws_iam_policy" "test" {
  name = %[1]q

  policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Action": [
		"rds:DescribeDBClusters",
		"rds:DescribeDBClusterParameters",
		"rds:DescribeDBSubnetGroups",
		"ec2:CreateNetworkInterface",
		"ec2:DescribeNetworkInterfaces",
		"ec2:DescribeVpcs",
		"ec2:DeleteNetworkInterface",
		"ec2:DescribeSubnets",
		"ec2:DescribeSecurityGroups",
		"kms:Decrypt",
		"secretsmanager:GetSecretValue"
      ],
      "Resource": "*"
    }
  ]
}
EOF
}

resource "aws_iam_policy_attachment" "test" {
  name       = %[1]q
  roles      = [aws_iam_role.test.name]
  policy_arn = aws_iam_policy.test.arn
}

resource "aws_secretsmanager_secret" "test" {
  name = %[1]q
}

resource "aws_secretsmanager_secret_version" "test" {
  secret_id     = aws_secretsmanager_secret.test.id
  secret_string = jsonencode({ username = "tfacctest", password = "avoid-plaintext-passwords" })
}

resource "aws_lambda_function" "test" {
  filename      = "test-fixtures/lambdatest.zip"
  function_name = %[1]q
  handler       = "exports.example"
  role          = aws_iam_role.test.arn
  runtime       = "nodejs16.x"
}

resource "aws_docdb_subnet_group" "test" {
  name       = %[1]q
  subnet_ids = aws_subnet.test[*].id
}

resource "aws_docdb_cluster" "test" {
  cluster_identifier = %[1]q

  db_subnet_group_name = aws_docdb_subnet_group.test.name
  master_username      = "tfacctest"
  master_password      = "avoid-plaintext-passwords"
  skip_final_snapshot  = true

  enabled_cloudwatch_logs_exports = [
    "audit",
    "profiler",
  ]
}
`, rName))
}

func testAccEventSourceMappingConfig_kinesisStartingPositionTimestamp(rName, startingPositionTimestamp string) string {
	return acctest.ConfigCompose(testAccEventSourceMappingConfig_kinesisBase(rName), fmt.Sprintf(`
resource "aws_lambda_event_source_mapping" "test" {
  batch_size                  = 100
  enabled                     = true
  event_source_arn            = aws_kinesis_stream.test.arn
  function_name               = aws_lambda_function.test.arn
  starting_position           = "AT_TIMESTAMP"
  starting_position_timestamp = %[1]q
}
`, startingPositionTimestamp))
}

func testAccEventSourceMappingConfig_kinesisBatchWindow(rName string, batchWindow int64) string {
	return acctest.ConfigCompose(testAccEventSourceMappingConfig_kinesisBase(rName), fmt.Sprintf(`
resource "aws_lambda_event_source_mapping" "test" {
  batch_size                         = 100
  maximum_batching_window_in_seconds = %[1]d
  enabled                            = true
  event_source_arn                   = aws_kinesis_stream.test.arn
  function_name                      = aws_lambda_function.test.arn
  starting_position                  = "TRIM_HORIZON"
}
`, batchWindow))
}

func testAccEventSourceMappingConfig_kinesisParallelizationFactor(rName string, parallelizationFactor int64) string {
	return acctest.ConfigCompose(testAccEventSourceMappingConfig_kinesisBase(rName), fmt.Sprintf(`
resource "aws_lambda_event_source_mapping" "test" {
  batch_size             = 100
  parallelization_factor = %[1]d
  enabled                = true
  event_source_arn       = aws_kinesis_stream.test.arn
  function_name          = aws_lambda_function.test.arn
  starting_position      = "TRIM_HORIZON"
}
`, parallelizationFactor))
}

func testAccEventSourceMappingConfig_kinesisTumblingWindowInSeconds(rName string, tumblingWindowInSeconds int64) string {
	return acctest.ConfigCompose(testAccEventSourceMappingConfig_kinesisBase(rName), fmt.Sprintf(`
resource "aws_lambda_event_source_mapping" "test" {
  batch_size                 = 100
  tumbling_window_in_seconds = %[1]d
  enabled                    = true
  event_source_arn           = aws_kinesis_stream.test.arn
  function_name              = aws_lambda_function.test.arn
  starting_position          = "TRIM_HORIZON"
}
`, tumblingWindowInSeconds))
}

func testAccEventSourceMappingConfig_kinesisMaximumRetryAttempts(rName string, maximumRetryAttempts int64) string {
	return acctest.ConfigCompose(testAccEventSourceMappingConfig_kinesisBase(rName), fmt.Sprintf(`
resource "aws_lambda_event_source_mapping" "test" {
  batch_size             = 100
  maximum_retry_attempts = %[1]d
  enabled                = true
  event_source_arn       = aws_kinesis_stream.test.arn
  function_name          = aws_lambda_function.test.arn
  starting_position      = "TRIM_HORIZON"
}
`, maximumRetryAttempts))
}

func testAccEventSourceMappingConfig_kinesisBisectBatch(rName string, bisectBatch bool) string {
	return acctest.ConfigCompose(testAccEventSourceMappingConfig_kinesisBase(rName), fmt.Sprintf(`
resource "aws_lambda_event_source_mapping" "test" {
  batch_size                     = 100
  bisect_batch_on_function_error = %[1]t
  enabled                        = true
  event_source_arn               = aws_kinesis_stream.test.arn
  function_name                  = aws_lambda_function.test.arn
  starting_position              = "TRIM_HORIZON"
}
`, bisectBatch))
}

func testAccEventSourceMappingConfig_kinesisMaximumRecordAgeInSeconds(rName string, maximumRecordAgeInSeconds int64) string {
	return acctest.ConfigCompose(testAccEventSourceMappingConfig_kinesisBase(rName), fmt.Sprintf(`
resource "aws_lambda_event_source_mapping" "test" {
  batch_size                    = 100
  maximum_record_age_in_seconds = %[1]d
  enabled                       = true
  event_source_arn              = aws_kinesis_stream.test.arn
  function_name                 = aws_lambda_function.test.arn
  starting_position             = "TRIM_HORIZON"
}
`, maximumRecordAgeInSeconds))
}

func testAccEventSourceMappingConfig_kinesisDestination(rName string, topicName string) string {
	return acctest.ConfigCompose(testAccEventSourceMappingConfig_kinesisBase(rName), fmt.Sprintf(`
resource "aws_sns_topic" "test" {
  name = %[1]q
}

resource "aws_lambda_event_source_mapping" "test" {
  batch_size        = 100
  enabled           = true
  event_source_arn  = aws_kinesis_stream.test.arn
  function_name     = aws_lambda_function.test.arn
  starting_position = "TRIM_HORIZON"

  destination_config {
    on_failure {
      destination_arn = aws_sns_topic.test.arn
    }
  }
}
`, topicName))
}

func testAccEventSourceMappingConfig_kinesisBatchSize(rName, batchSize string) string {
	if batchSize == "" {
		batchSize = "null"
	}

	return acctest.ConfigCompose(testAccEventSourceMappingConfig_kinesisBase(rName), fmt.Sprintf(`
resource "aws_lambda_event_source_mapping" "test" {
  batch_size        = %[1]s
  enabled           = true
  event_source_arn  = aws_kinesis_stream.test.arn
  function_name     = aws_lambda_function.test.function_name
  starting_position = "TRIM_HORIZON"
}
`, batchSize))
}

func testAccEventSourceMappingConfig_kinesisUpdateFunctionName(rName string) string {
	return acctest.ConfigCompose(testAccEventSourceMappingConfig_kinesisBase(rName), fmt.Sprintf(`
resource "aws_lambda_function" "test_update" {
  filename      = "test-fixtures/lambdatest.zip"
  function_name = "%[1]s-update"
  role          = aws_iam_role.test.arn
  handler       = "exports.example"
  runtime       = "nodejs16.x"
}

resource "aws_lambda_event_source_mapping" "test" {
  batch_size        = 200
  enabled           = false
  event_source_arn  = aws_kinesis_stream.test.arn
  function_name     = aws_lambda_function.test_update.arn
  starting_position = "TRIM_HORIZON"
}
`, rName))
}

func testAccEventSourceMappingConfig_sqsBatchSize(rName, batchSize string) string {
	if batchSize == "" {
		batchSize = "null"
	}

	return acctest.ConfigCompose(testAccEventSourceMappingConfig_sqsBase(rName), fmt.Sprintf(`
resource "aws_lambda_event_source_mapping" "test" {
  batch_size       = %[1]s
  enabled          = true
  event_source_arn = aws_sqs_queue.test.arn
  function_name    = aws_lambda_function.test.function_name
}
`, batchSize))
}

func testAccEventSourceMappingConfig_sqsUpdateFunctionName(rName string) string {
	return acctest.ConfigCompose(testAccEventSourceMappingConfig_sqsBase(rName), fmt.Sprintf(`
resource "aws_iam_role" "test_update" {
  name = "%[1]s-update"

  assume_role_policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [{
    "Action": "sts:AssumeRole",
    "Principal": {
      "Service": "lambda.amazonaws.com"
    },
    "Effect": "Allow"
  }]
}
EOF
}

resource "aws_iam_role_policy" "test_update" {
  role = aws_iam_role.test_update.name

  policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [{
    "Effect": "Allow",
    "Action": ["sqs:*"],
    "Resource": "*"
  }]
}
EOF

  depends_on = [aws_lambda_function.test_update]
}

resource "aws_lambda_function" "test_update" {
  filename      = "test-fixtures/lambdatest.zip"
  function_name = "%[1]s-update"
  role          = aws_iam_role.test_update.arn
  handler       = "exports.example"
  runtime       = "nodejs16.x"
}

resource "aws_lambda_event_source_mapping" "test" {
  batch_size       = 5
  enabled          = false
  event_source_arn = aws_sqs_queue.test.arn
  function_name    = aws_lambda_function.test_update.arn
}
`, rName))
}

func testAccEventSourceMappingConfig_sqsBatchWindow(rName string, batchWindow int64) string {
	return acctest.ConfigCompose(testAccEventSourceMappingConfig_sqsBase(rName), fmt.Sprintf(`
resource "aws_lambda_event_source_mapping" "test" {
  batch_size                         = 10
  maximum_batching_window_in_seconds = %[1]d
  event_source_arn                   = aws_sqs_queue.test.arn
  enabled                            = false
  function_name                      = aws_lambda_function.test.arn
}
`, batchWindow))
}

func testAccEventSourceMappingConfig_msk(rName, batchSize string) string {
	if batchSize == "" {
		batchSize = "null"
	}

	return acctest.ConfigCompose(testAccEventSourceMappingConfig_kafkaBase(rName), fmt.Sprintf(`
resource "aws_msk_cluster" "test" {
  cluster_name           = %[1]q
  kafka_version          = "2.7.1"
  number_of_broker_nodes = 2

  broker_node_group_info {
    client_subnets  = aws_subnet.test[*].id
    instance_type   = "kafka.m5.large"
    security_groups = [aws_security_group.test.id]

    storage_info {
      ebs_storage_info {
        volume_size = 10
      }
    }
  }
}

resource "aws_lambda_event_source_mapping" "test" {
  batch_size        = %[2]s
  event_source_arn  = aws_msk_cluster.test.arn
  enabled           = true
  function_name     = aws_lambda_function.test.arn
  topics            = ["test"]
  starting_position = "TRIM_HORIZON"

  depends_on = [aws_iam_policy_attachment.test]
}
`, rName, batchSize))
}

func testAccEventSourceMappingConfig_mskWithEventSourceConfig(rName, batchSize string) string {
	if batchSize == "" {
		batchSize = "null"
	}

	return acctest.ConfigCompose(testAccEventSourceMappingConfig_kafkaBase(rName), fmt.Sprintf(`
resource "aws_msk_cluster" "test" {
  cluster_name           = %[1]q
  kafka_version          = "2.7.1"
  number_of_broker_nodes = 2

  broker_node_group_info {
    client_subnets  = aws_subnet.test[*].id
    instance_type   = "kafka.m5.large"
    security_groups = [aws_security_group.test.id]

    storage_info {
      ebs_storage_info {
        volume_size = 10
      }
    }
  }
}

resource "aws_lambda_event_source_mapping" "test" {
  batch_size        = %[2]s
  event_source_arn  = aws_msk_cluster.test.arn
  enabled           = true
  function_name     = aws_lambda_function.test.arn
  topics            = ["test"]
  starting_position = "TRIM_HORIZON"

  amazon_managed_kafka_event_source_config {
    consumer_group_id = "amazon-managed-test-group-id"
  }

  depends_on = [aws_iam_policy_attachment.test]
}
`, rName, batchSize))
}

func testAccEventSourceMappingConfig_selfManagedKafka(rName, batchSize, kafkaBootstrapServers string) string {
	if batchSize == "" {
		batchSize = "null"
	}

	return acctest.ConfigCompose(testAccEventSourceMappingConfig_kafkaBase(rName), fmt.Sprintf(`
resource "aws_lambda_event_source_mapping" "test" {
  batch_size        = %[2]s
  enabled           = false
  function_name     = aws_lambda_function.test.arn
  topics            = ["test"]
  starting_position = "TRIM_HORIZON"

  self_managed_event_source {
    endpoints = {
      KAFKA_BOOTSTRAP_SERVERS = %[3]q
    }
  }

  dynamic "source_access_configuration" {
    for_each = aws_subnet.test[*].id
    content {
      type = "VPC_SUBNET"
      uri  = "subnet:${source_access_configuration.value}"
    }
  }

  source_access_configuration {
    type = "VPC_SECURITY_GROUP"
    uri  = aws_security_group.test.id
  }
}
`, rName, batchSize, kafkaBootstrapServers))
}

func testAccEventSourceMappingConfig_selfManagedKafkaWithEventSourceConfig(rName, batchSize, kafkaBootstrapServers string) string {
	if batchSize == "" {
		batchSize = "null"
	}

	return acctest.ConfigCompose(testAccEventSourceMappingConfig_kafkaBase(rName), fmt.Sprintf(`
resource "aws_lambda_event_source_mapping" "test" {
  batch_size        = %[2]s
  enabled           = false
  function_name     = aws_lambda_function.test.arn
  topics            = ["test"]
  starting_position = "TRIM_HORIZON"

  self_managed_kafka_event_source_config {
    consumer_group_id = "self-managed-test-group-id"
  }

  self_managed_event_source {
    endpoints = {
      KAFKA_BOOTSTRAP_SERVERS = %[3]q
    }
  }

  dynamic "source_access_configuration" {
    for_each = aws_subnet.test[*].id
    content {
      type = "VPC_SUBNET"
      uri  = "subnet:${source_access_configuration.value}"
    }
  }

  source_access_configuration {
    type = "VPC_SECURITY_GROUP"
    uri  = aws_security_group.test.id
  }
}
`, rName, batchSize, kafkaBootstrapServers))
}

func testAccEventSourceMappingConfig_dynamoDBBatchSize(rName, batchSize string) string {
	if batchSize == "" {
		batchSize = "null"
	}

	return acctest.ConfigCompose(testAccEventSourceMappingConfig_dynamoDBBase(rName), fmt.Sprintf(`
resource "aws_lambda_event_source_mapping" "test" {
  batch_size        = %[1]s
  enabled           = true
  event_source_arn  = aws_dynamodb_table.test.stream_arn
  function_name     = aws_lambda_function.test.function_name
  starting_position = "LATEST"
}
`, batchSize))
}

func testAccEventSourceMappingConfig_dynamoDBFunctionResponseTypes(rName string) string {
	return acctest.ConfigCompose(testAccEventSourceMappingConfig_dynamoDBBase(rName), `
resource "aws_lambda_event_source_mapping" "test" {
  batch_size        = 150
  enabled           = true
  event_source_arn  = aws_dynamodb_table.test.stream_arn
  function_name     = aws_lambda_function.test.function_name
  starting_position = "LATEST"

  function_response_types = ["ReportBatchItemFailures"]
}
`)
}

func testAccEventSourceMappingConfig_dynamoDBNoFunctionResponseTypes(rName string) string {
	return acctest.ConfigCompose(testAccEventSourceMappingConfig_dynamoDBBase(rName), `
resource "aws_lambda_event_source_mapping" "test" {
  batch_size        = 150
  enabled           = true
  event_source_arn  = aws_dynamodb_table.test.stream_arn
  function_name     = aws_lambda_function.test.function_name
  starting_position = "LATEST"
}
`)
}

func testAccEventSourceMappingConfig_dynamoDBStreamEnabled(rName string) string {
	return acctest.ConfigCompose(testAccEventSourceMappingConfig_dynamoDBStreamBase(rName, true), `
resource "aws_lambda_event_source_mapping" "test" {
  batch_size        = 150
  enabled           = true
  event_source_arn  = aws_dynamodb_table.test.stream_arn
  function_name     = aws_lambda_function.test.function_name
  starting_position = "LATEST"
}
`)
}

func testAccEventSourceMappingConfig_activeMQ(rName, batchSize string) string {
	if batchSize == "" {
		batchSize = "null"
	}

	return acctest.ConfigCompose(testAccEventSourceMappingConfig_activeMQBase(rName), fmt.Sprintf(`
resource "aws_lambda_event_source_mapping" "test" {
  batch_size       = %[1]s
  event_source_arn = aws_mq_broker.test.arn
  enabled          = true
  function_name    = aws_lambda_function.test.arn
  queues           = ["test"]

  source_access_configuration {
    type = "BASIC_AUTH"
    uri  = aws_secretsmanager_secret_version.test.arn
  }
}
`, batchSize))
}

func testAccEventSourceMappingConfig_rabbitMQ(rName, batchSize string) string {
	if batchSize == "" {
		batchSize = "null"
	}

	return acctest.ConfigCompose(testAccEventSourceMappingConfig_rabbitMQBase(rName), fmt.Sprintf(`
resource "aws_lambda_event_source_mapping" "test" {
  batch_size       = %[1]s
  event_source_arn = aws_mq_broker.test.arn
  enabled          = true
  function_name    = aws_lambda_function.test.arn
  queues           = ["test"]

  source_access_configuration {
    type = "VIRTUAL_HOST"
    uri  = "/vhost"
  }

  source_access_configuration {
    type = "BASIC_AUTH"
    uri  = aws_secretsmanager_secret_version.test.arn
  }
}
`, batchSize))
}

func testAccEventSourceMappingConfig_documentDB(rName string) string {
	return acctest.ConfigCompose(testAccEventSourceMappingConfig_documentDBBase(rName), `
resource "aws_lambda_event_source_mapping" "test" {
  enabled          = true
  event_source_arn = aws_docdb_cluster.test.arn
  function_name    = aws_lambda_function.test.arn

  document_db_event_source_config {
    database_name = "test"
  }

  source_access_configuration {
    type = "BASIC_AUTH"
    uri  = aws_secretsmanager_secret_version.test.arn
  }

  depends_on = [aws_iam_policy_attachment.test]
}
`)
}

func testAccEventSourceMappingConfig_sqsFilterCriteria1(rName string, pattern1 string) string {
	return acctest.ConfigCompose(testAccEventSourceMappingConfig_sqsBase(rName), fmt.Sprintf(`
resource "aws_lambda_event_source_mapping" "test" {
  event_source_arn = aws_sqs_queue.test.arn
  function_name    = aws_lambda_function.test.arn

  filter_criteria {
    filter {
      pattern = %q
    }
  }
}
`, pattern1))
}

func testAccEventSourceMappingConfig_sqsFilterCriteria2(rName string, pattern1, pattern2 string) string {
	return acctest.ConfigCompose(testAccEventSourceMappingConfig_sqsBase(rName), fmt.Sprintf(`
resource "aws_lambda_event_source_mapping" "test" {
  event_source_arn = aws_sqs_queue.test.arn
  function_name    = aws_lambda_function.test.arn

  filter_criteria {
    filter {
      pattern = %q
    }

    filter {
      pattern = %q
    }
  }
}
`, pattern1, pattern2))
}

func testAccEventSourceMappingConfig_sqsFilterCriteria3(rName string) string {
	return acctest.ConfigCompose(testAccEventSourceMappingConfig_sqsBase(rName), `
resource "aws_lambda_event_source_mapping" "test" {
  event_source_arn = aws_sqs_queue.test.arn
  function_name    = aws_lambda_function.test.arn
}
`)
}

func testAccEventSourceMappingConfig_sqsScalingConfig1(rName string, maximumConcurrency int) string {
	return acctest.ConfigCompose(testAccEventSourceMappingConfig_sqsBase(rName), fmt.Sprintf(`
resource "aws_lambda_event_source_mapping" "test" {
  event_source_arn = aws_sqs_queue.test.arn
  function_name    = aws_lambda_function.test.arn

  scaling_config {
    maximum_concurrency = %[1]d
  }
}
`, maximumConcurrency))
}

func testAccEventSourceMappingConfig_sqsScalingConfig2(rName string) string {
	return acctest.ConfigCompose(testAccEventSourceMappingConfig_sqsBase(rName), `
resource "aws_lambda_event_source_mapping" "test" {
  event_source_arn = aws_sqs_queue.test.arn
  function_name    = aws_lambda_function.test.arn
}
`)
}
