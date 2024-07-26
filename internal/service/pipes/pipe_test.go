// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package pipes_test

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"testing"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/service/pipes"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/id"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	tfpipes "github.com/hashicorp/terraform-provider-aws/internal/service/pipes"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccPipesPipe_basicSQS(t *testing.T) {
	ctx := acctest.Context(t)
	var pipe pipes.DescribePipeOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_pipes_pipe.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.PipesEndpointID)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.PipesServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckPipeDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccPipeConfig_basicSQS(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckPipeExists(ctx, resourceName, &pipe),
					acctest.MatchResourceAttrRegionalARN(resourceName, names.AttrARN, "pipes", regexache.MustCompile(regexp.QuoteMeta(`pipe/`+rName))),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, "Managed by Terraform"),
					resource.TestCheckResourceAttr(resourceName, "desired_state", "RUNNING"),
					resource.TestCheckResourceAttr(resourceName, "enrichment", ""),
					resource.TestCheckResourceAttr(resourceName, "enrichment_parameters.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrRoleARN, "aws_iam_role.test", names.AttrARN),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrSource, "aws_sqs_queue.source", names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, "source_parameters.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "source_parameters.0.activemq_broker_parameters.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "source_parameters.0.dynamodb_stream_parameters.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "source_parameters.0.filter_criteria.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "source_parameters.0.kinesis_stream_parameters.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "source_parameters.0.managed_streaming_kafka_parameters.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "source_parameters.0.rabbitmq_broker_parameters.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "source_parameters.0.self_managed_kafka_parameters.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "source_parameters.0.sqs_queue_parameters.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "source_parameters.0.sqs_queue_parameters.0.batch_size", acctest.Ct10),
					resource.TestCheckResourceAttr(resourceName, "source_parameters.0.sqs_queue_parameters.0.maximum_batching_window_in_seconds", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct0),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrTarget, "aws_sqs_queue.target", names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, "target_parameters.#", acctest.Ct0),
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

func TestAccPipesPipe_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	var pipe pipes.DescribePipeOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_pipes_pipe.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.PipesEndpointID)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.PipesServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckPipeDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccPipeConfig_basicSQS(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPipeExists(ctx, resourceName, &pipe),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfpipes.ResourcePipe(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccPipesPipe_description(t *testing.T) {
	ctx := acctest.Context(t)
	var pipe pipes.DescribePipeOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_pipes_pipe.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.PipesEndpointID)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.PipesServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckPipeDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccPipeConfig_description(rName, "Description 1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPipeExists(ctx, resourceName, &pipe),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, "Description 1"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccPipeConfig_description(rName, "Description 2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPipeExists(ctx, resourceName, &pipe),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, "Description 2"),
				),
			},
			{
				Config: testAccPipeConfig_description(rName, ""),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPipeExists(ctx, resourceName, &pipe),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, ""),
				),
			},
			{
				Config: testAccPipeConfig_basicSQS(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPipeExists(ctx, resourceName, &pipe),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, "Managed by Terraform"),
				),
			},
		},
	})
}

func TestAccPipesPipe_desiredState(t *testing.T) {
	ctx := acctest.Context(t)
	var pipe pipes.DescribePipeOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_pipes_pipe.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.PipesEndpointID)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.PipesServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckPipeDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccPipeConfig_desiredState(rName, "STOPPED"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPipeExists(ctx, resourceName, &pipe),
					resource.TestCheckResourceAttr(resourceName, "desired_state", "STOPPED"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccPipeConfig_desiredState(rName, "RUNNING"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPipeExists(ctx, resourceName, &pipe),
					resource.TestCheckResourceAttr(resourceName, "desired_state", "RUNNING"),
				),
			},
			{
				Config: testAccPipeConfig_desiredState(rName, "STOPPED"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPipeExists(ctx, resourceName, &pipe),
					resource.TestCheckResourceAttr(resourceName, "desired_state", "STOPPED"),
				),
			},
			{
				Config: testAccPipeConfig_basicSQS(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPipeExists(ctx, resourceName, &pipe),
					resource.TestCheckResourceAttr(resourceName, "desired_state", "RUNNING"),
				),
			},
		},
	})
}

func TestAccPipesPipe_enrichment(t *testing.T) {
	ctx := acctest.Context(t)
	var pipe pipes.DescribePipeOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_pipes_pipe.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.PipesEndpointID)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.PipesServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckPipeDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccPipeConfig_enrichment(rName, 0),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPipeExists(ctx, resourceName, &pipe),
					resource.TestCheckResourceAttrPair(resourceName, "enrichment", "aws_cloudwatch_event_api_destination.test.0", names.AttrARN),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccPipeConfig_enrichment(rName, 1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPipeExists(ctx, resourceName, &pipe),
					resource.TestCheckResourceAttrPair(resourceName, "enrichment", "aws_cloudwatch_event_api_destination.test.1", names.AttrARN),
				),
			},
			{
				Config: testAccPipeConfig_basicSQS(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPipeExists(ctx, resourceName, &pipe),
					resource.TestCheckResourceAttr(resourceName, "enrichment", ""),
				),
			},
		},
	})
}

func TestAccPipesPipe_enrichmentParameters(t *testing.T) {
	ctx := acctest.Context(t)
	var pipe pipes.DescribePipeOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_pipes_pipe.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.PipesEndpointID)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.PipesServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckPipeDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccPipeConfig_enrichmentParameters(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckPipeExists(ctx, resourceName, &pipe),
					resource.TestCheckResourceAttrPair(resourceName, "enrichment", "aws_cloudwatch_event_api_destination.test", names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, "enrichment_parameters.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "enrichment_parameters.0.http_parameters.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "enrichment_parameters.0.http_parameters.0.header_parameters.%", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "enrichment_parameters.0.http_parameters.0.header_parameters.X-Test-1", "Val1"),
					resource.TestCheckResourceAttr(resourceName, "enrichment_parameters.0.http_parameters.0.path_parameter_values.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "enrichment_parameters.0.http_parameters.0.path_parameter_values.0", "p1"),
					resource.TestCheckResourceAttr(resourceName, "enrichment_parameters.0.http_parameters.0.query_string_parameters.%", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "enrichment_parameters.0.http_parameters.0.query_string_parameters.q1", "abc"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccPipeConfig_enrichmentParametersUpdated(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckPipeExists(ctx, resourceName, &pipe),
					resource.TestCheckResourceAttrPair(resourceName, "enrichment", "aws_cloudwatch_event_api_destination.test", names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, "enrichment_parameters.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "enrichment_parameters.0.http_parameters.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "enrichment_parameters.0.http_parameters.0.header_parameters.%", acctest.Ct2),
					resource.TestCheckResourceAttr(resourceName, "enrichment_parameters.0.http_parameters.0.header_parameters.X-Test-1", "Val1"),
					resource.TestCheckResourceAttr(resourceName, "enrichment_parameters.0.http_parameters.0.header_parameters.X-Test-2", "Val2"),
					resource.TestCheckResourceAttr(resourceName, "enrichment_parameters.0.http_parameters.0.path_parameter_values.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "enrichment_parameters.0.http_parameters.0.path_parameter_values.0", "p2"),
					resource.TestCheckResourceAttr(resourceName, "enrichment_parameters.0.http_parameters.0.query_string_parameters.%", acctest.Ct0),
				),
			},
		},
	})
}

func TestAccPipesPipe_logConfiguration_cloudwatchLogsLogDestination(t *testing.T) {
	ctx := acctest.Context(t)
	var pipe pipes.DescribePipeOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_pipes_pipe.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.PipesEndpointID)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.PipesServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckPipeDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccPipeConfig_logConfiguration_cloudwatchLogsLogDestination(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckPipeExists(ctx, resourceName, &pipe),
					acctest.MatchResourceAttrRegionalARN(resourceName, names.AttrARN, "pipes", regexache.MustCompile(regexp.QuoteMeta(`pipe/`+rName))),
					resource.TestCheckResourceAttr(resourceName, "log_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "log_configuration.0.level", "INFO"),
					resource.TestCheckResourceAttr(resourceName, "log_configuration.0.cloudwatch_logs_log_destination.#", acctest.Ct1),
					resource.TestCheckResourceAttrSet(resourceName, "log_configuration.0.cloudwatch_logs_log_destination.0.log_group_arn"),
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

func TestAccPipesPipe_update_logConfiguration_cloudwatchLogsLogDestination(t *testing.T) {
	ctx := acctest.Context(t)
	var pipe pipes.DescribePipeOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_pipes_pipe.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.PipesEndpointID)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.PipesServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckPipeDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccPipeConfig_logConfiguration_cloudwatchLogsLogDestination(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckPipeExists(ctx, resourceName, &pipe),
					acctest.MatchResourceAttrRegionalARN(resourceName, names.AttrARN, "pipes", regexache.MustCompile(regexp.QuoteMeta(`pipe/`+rName))),
					resource.TestCheckResourceAttr(resourceName, "log_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "log_configuration.0.level", "INFO"),
					resource.TestCheckResourceAttr(resourceName, "log_configuration.0.cloudwatch_logs_log_destination.#", acctest.Ct1),
					resource.TestCheckResourceAttrSet(resourceName, "log_configuration.0.cloudwatch_logs_log_destination.0.log_group_arn"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccPipeConfig_logConfiguration_update_cloudwatchLogsLogDestination(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckPipeExists(ctx, resourceName, &pipe),
					acctest.MatchResourceAttrRegionalARN(resourceName, names.AttrARN, "pipes", regexache.MustCompile(regexp.QuoteMeta(`pipe/`+rName))),
					resource.TestCheckResourceAttr(resourceName, "log_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "log_configuration.0.level", "ERROR"),
					resource.TestCheckResourceAttr(resourceName, "log_configuration.0.cloudwatch_logs_log_destination.#", acctest.Ct1),
					resource.TestCheckResourceAttrSet(resourceName, "log_configuration.0.cloudwatch_logs_log_destination.0.log_group_arn"),
				),
			},
		},
	})
}

func TestAccPipesPipe_sourceParameters_filterCriteria(t *testing.T) {
	ctx := acctest.Context(t)
	var pipe pipes.DescribePipeOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_pipes_pipe.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.PipesEndpointID)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.PipesServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckPipeDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccPipeConfig_sourceParameters_filterCriteria1(rName, "test1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPipeExists(ctx, resourceName, &pipe),
					resource.TestCheckResourceAttr(resourceName, "source_parameters.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "source_parameters.0.filter_criteria.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "source_parameters.0.filter_criteria.0.filter.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "source_parameters.0.filter_criteria.0.filter.0.pattern", `{"source":["test1"]}`),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccPipeConfig_sourceParameters_filterCriteria2(rName, "test1", "test2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPipeExists(ctx, resourceName, &pipe),
					resource.TestCheckResourceAttr(resourceName, "source_parameters.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "source_parameters.0.filter_criteria.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "source_parameters.0.filter_criteria.0.filter.#", acctest.Ct2),
					resource.TestCheckResourceAttr(resourceName, "source_parameters.0.filter_criteria.0.filter.0.pattern", `{"source":["test1"]}`),
					resource.TestCheckResourceAttr(resourceName, "source_parameters.0.filter_criteria.0.filter.1.pattern", `{"source":["test2"]}`),
				),
			},
			{
				Config: testAccPipeConfig_sourceParameters_filterCriteria1(rName, "test2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPipeExists(ctx, resourceName, &pipe),
					resource.TestCheckResourceAttr(resourceName, "source_parameters.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "source_parameters.0.filter_criteria.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "source_parameters.0.filter_criteria.0.filter.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "source_parameters.0.filter_criteria.0.filter.0.pattern", `{"source":["test2"]}`),
				),
			},
			{
				Config: testAccPipeConfig_sourceParameters_filterCriteria0(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPipeExists(ctx, resourceName, &pipe),
					resource.TestCheckResourceAttr(resourceName, "source_parameters.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "source_parameters.0.filter_criteria.#", acctest.Ct0),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccPipeConfig_sourceParameters_filterCriteria1(rName, "test2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPipeExists(ctx, resourceName, &pipe),
					resource.TestCheckResourceAttr(resourceName, "source_parameters.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "source_parameters.0.filter_criteria.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "source_parameters.0.filter_criteria.0.filter.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "source_parameters.0.filter_criteria.0.filter.0.pattern", `{"source":["test2"]}`),
				),
			},
			{
				Config: testAccPipeConfig_basicSQS(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPipeExists(ctx, resourceName, &pipe),
					resource.TestCheckResourceAttr(resourceName, "source_parameters.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "source_parameters.0.filter_criteria.#", acctest.Ct1),
				),
			},
		},
	})
}

func TestAccPipesPipe_nameGenerated(t *testing.T) {
	ctx := acctest.Context(t)
	var pipe pipes.DescribePipeOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_pipes_pipe.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.PipesEndpointID)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.PipesServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckPipeDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccPipeConfig_nameGenerated(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPipeExists(ctx, resourceName, &pipe),
					acctest.CheckResourceAttrNameGenerated(resourceName, names.AttrName),
					resource.TestCheckResourceAttr(resourceName, names.AttrNamePrefix, id.UniqueIdPrefix),
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

func TestAccPipesPipe_namePrefix(t *testing.T) {
	ctx := acctest.Context(t)
	var pipe pipes.DescribePipeOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_pipes_pipe.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.PipesEndpointID)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.PipesServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckPipeDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccPipeConfig_namePrefix(rName, "tf-acc-test-prefix-"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPipeExists(ctx, resourceName, &pipe),
					acctest.CheckResourceAttrNameFromPrefix(resourceName, names.AttrName, "tf-acc-test-prefix-"),
					resource.TestCheckResourceAttr(resourceName, names.AttrNamePrefix, "tf-acc-test-prefix-"),
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

func TestAccPipesPipe_roleARN(t *testing.T) {
	ctx := acctest.Context(t)
	var pipe pipes.DescribePipeOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_pipes_pipe.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.PipesEndpointID)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.PipesServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckPipeDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccPipeConfig_basicSQS(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPipeExists(ctx, resourceName, &pipe),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrRoleARN, "aws_iam_role.test", names.AttrARN),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccPipeConfig_roleARN(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPipeExists(ctx, resourceName, &pipe),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrRoleARN, "aws_iam_role.test2", names.AttrARN),
				),
			},
		},
	})
}

func TestAccPipesPipe_tags(t *testing.T) {
	ctx := acctest.Context(t)
	var pipe pipes.DescribePipeOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_pipes_pipe.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.PipesEndpointID)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.PipesServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckPipeDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccPipeConfig_tags1(rName, acctest.CtKey1, acctest.CtValue1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPipeExists(ctx, resourceName, &pipe),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey1, acctest.CtValue1),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccPipeConfig_tags2(rName, acctest.CtKey1, acctest.CtValue1Updated, acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPipeExists(ctx, resourceName, &pipe),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct2),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey1, acctest.CtValue1Updated),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey2, acctest.CtValue2),
				),
			},
			{
				Config: testAccPipeConfig_tags1(rName, acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPipeExists(ctx, resourceName, &pipe),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey2, acctest.CtValue2),
				),
			},
		},
	})
}

func TestAccPipesPipe_targetUpdate(t *testing.T) {
	ctx := acctest.Context(t)
	var pipe pipes.DescribePipeOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_pipes_pipe.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.PipesEndpointID)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.PipesServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckPipeDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccPipeConfig_basicSQS(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPipeExists(ctx, resourceName, &pipe),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrTarget, "aws_sqs_queue.target", names.AttrARN),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccPipeConfig_targetUpdated(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPipeExists(ctx, resourceName, &pipe),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrTarget, "aws_sqs_queue.target2", names.AttrARN),
				),
			},
		},
	})
}

func TestAccPipesPipe_targetParameters_inputTemplate(t *testing.T) {
	ctx := acctest.Context(t)
	var pipe pipes.DescribePipeOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_pipes_pipe.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.PipesEndpointID)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.PipesServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckPipeDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccPipeConfig_targetParameters_inputTemplate(rName, "$.first"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPipeExists(ctx, resourceName, &pipe),
					resource.TestCheckResourceAttr(resourceName, "target_parameters.0.input_template", "$.first"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccPipeConfig_targetParameters_inputTemplate(rName, "$.second"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPipeExists(ctx, resourceName, &pipe),
					resource.TestCheckResourceAttr(resourceName, "target_parameters.0.input_template", "$.second"),
				),
			},
			{
				Config: testAccPipeConfig_basicSQS(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPipeExists(ctx, resourceName, &pipe),
					resource.TestCheckNoResourceAttr(resourceName, "target_parameters.0.input_template"),
				),
			},
		},
	})
}

func TestAccPipesPipe_kinesisSourceAndTarget(t *testing.T) {
	ctx := acctest.Context(t)
	var pipe pipes.DescribePipeOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_pipes_pipe.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.PipesEndpointID)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.PipesServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckPipeDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccPipeConfig_basicKinesis(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckPipeExists(ctx, resourceName, &pipe),
					acctest.MatchResourceAttrRegionalARN(resourceName, names.AttrARN, "pipes", regexache.MustCompile(regexp.QuoteMeta(`pipe/`+rName))),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, "Managed by Terraform"),
					resource.TestCheckResourceAttr(resourceName, "desired_state", "RUNNING"),
					resource.TestCheckResourceAttr(resourceName, "enrichment", ""),
					resource.TestCheckResourceAttr(resourceName, "enrichment_parameters.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrRoleARN, "aws_iam_role.test", names.AttrARN),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrSource, "aws_kinesis_stream.source", names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, "source_parameters.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "source_parameters.0.activemq_broker_parameters.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "source_parameters.0.dynamodb_stream_parameters.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "source_parameters.0.filter_criteria.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "source_parameters.0.kinesis_stream_parameters.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "source_parameters.0.kinesis_stream_parameters.0.batch_size", "100"),
					resource.TestCheckResourceAttr(resourceName, "source_parameters.0.kinesis_stream_parameters.0.dead_letter_config.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "source_parameters.0.kinesis_stream_parameters.0.maximum_batching_window_in_seconds", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "source_parameters.0.kinesis_stream_parameters.0.maximum_record_age_in_seconds", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "source_parameters.0.kinesis_stream_parameters.0.maximum_retry_attempts", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "source_parameters.0.kinesis_stream_parameters.0.on_partial_batch_item_failure", ""),
					resource.TestCheckResourceAttr(resourceName, "source_parameters.0.kinesis_stream_parameters.0.parallelization_factor", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "source_parameters.0.kinesis_stream_parameters.0.starting_position", "LATEST"),
					resource.TestCheckResourceAttr(resourceName, "source_parameters.0.kinesis_stream_parameters.0.starting_position_timestamp", ""),
					resource.TestCheckResourceAttr(resourceName, "source_parameters.0.managed_streaming_kafka_parameters.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "source_parameters.0.rabbitmq_broker_parameters.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "source_parameters.0.self_managed_kafka_parameters.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "source_parameters.0.sqs_queue_parameters.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct0),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrTarget, "aws_kinesis_stream.target", names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, "target_parameters.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "target_parameters.0.batch_job_parameters.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "target_parameters.0.cloudwatch_logs_parameters.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "target_parameters.0.ecs_task_parameters.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "target_parameters.0.eventbridge_event_bus_parameters.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "target_parameters.0.http_parameters.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "target_parameters.0.input_template", ""),
					resource.TestCheckResourceAttr(resourceName, "target_parameters.0.kinesis_stream_parameters.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "target_parameters.0.kinesis_stream_parameters.0.partition_key", "test"),
					resource.TestCheckResourceAttr(resourceName, "target_parameters.0.lambda_function_parameters.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "target_parameters.0.redshift_data_parameters.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "target_parameters.0.sagemaker_pipeline_parameters.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "target_parameters.0.sqs_queue_parameters.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "target_parameters.0.step_function_state_machine_parameters.#", acctest.Ct0),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccPipeConfig_updateKinesis(rName, 10),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckPipeExists(ctx, resourceName, &pipe),
					acctest.MatchResourceAttrRegionalARN(resourceName, names.AttrARN, "pipes", regexache.MustCompile(regexp.QuoteMeta(`pipe/`+rName))),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, "Managed by Terraform"),
					resource.TestCheckResourceAttr(resourceName, "desired_state", "RUNNING"),
					resource.TestCheckResourceAttr(resourceName, "enrichment", ""),
					resource.TestCheckResourceAttr(resourceName, "enrichment_parameters.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrRoleARN, "aws_iam_role.test", names.AttrARN),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrSource, "aws_kinesis_stream.source", names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, "source_parameters.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "source_parameters.0.activemq_broker_parameters.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "source_parameters.0.dynamodb_stream_parameters.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "source_parameters.0.filter_criteria.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "source_parameters.0.kinesis_stream_parameters.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "source_parameters.0.kinesis_stream_parameters.0.batch_size", acctest.Ct10),
					resource.TestCheckResourceAttr(resourceName, "source_parameters.0.kinesis_stream_parameters.0.dead_letter_config.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "source_parameters.0.kinesis_stream_parameters.0.maximum_batching_window_in_seconds", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "source_parameters.0.kinesis_stream_parameters.0.maximum_record_age_in_seconds", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "source_parameters.0.kinesis_stream_parameters.0.maximum_retry_attempts", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "source_parameters.0.kinesis_stream_parameters.0.on_partial_batch_item_failure", ""),
					resource.TestCheckResourceAttr(resourceName, "source_parameters.0.kinesis_stream_parameters.0.parallelization_factor", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "source_parameters.0.kinesis_stream_parameters.0.starting_position", "LATEST"),
					resource.TestCheckResourceAttr(resourceName, "source_parameters.0.kinesis_stream_parameters.0.starting_position_timestamp", ""),
					resource.TestCheckResourceAttr(resourceName, "source_parameters.0.managed_streaming_kafka_parameters.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "source_parameters.0.rabbitmq_broker_parameters.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "source_parameters.0.self_managed_kafka_parameters.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "source_parameters.0.sqs_queue_parameters.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct0),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrTarget, "aws_kinesis_stream.target", names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, "target_parameters.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "target_parameters.0.batch_job_parameters.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "target_parameters.0.cloudwatch_logs_parameters.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "target_parameters.0.ecs_task_parameters.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "target_parameters.0.eventbridge_event_bus_parameters.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "target_parameters.0.http_parameters.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "target_parameters.0.input_template", ""),
					resource.TestCheckResourceAttr(resourceName, "target_parameters.0.kinesis_stream_parameters.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "target_parameters.0.kinesis_stream_parameters.0.partition_key", "test"),
					resource.TestCheckResourceAttr(resourceName, "target_parameters.0.lambda_function_parameters.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "target_parameters.0.redshift_data_parameters.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "target_parameters.0.sagemaker_pipeline_parameters.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "target_parameters.0.sqs_queue_parameters.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "target_parameters.0.step_function_state_machine_parameters.#", acctest.Ct0),
				),
			},
		},
	})
}

func TestAccPipesPipe_dynamoDBSourceCloudWatchLogsTarget(t *testing.T) {
	ctx := acctest.Context(t)
	var pipe pipes.DescribePipeOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_pipes_pipe.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.PipesEndpointID)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.PipesServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckPipeDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccPipeConfig_basicDynamoDBSourceCloudWatchLogsTarget(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckPipeExists(ctx, resourceName, &pipe),
					acctest.MatchResourceAttrRegionalARN(resourceName, names.AttrARN, "pipes", regexache.MustCompile(regexp.QuoteMeta(`pipe/`+rName))),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, "Managed by Terraform"),
					resource.TestCheckResourceAttr(resourceName, "desired_state", "RUNNING"),
					resource.TestCheckResourceAttr(resourceName, "enrichment", ""),
					resource.TestCheckResourceAttr(resourceName, "enrichment_parameters.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrRoleARN, "aws_iam_role.test", names.AttrARN),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrSource, "aws_dynamodb_table.source", names.AttrStreamARN),
					resource.TestCheckResourceAttr(resourceName, "source_parameters.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "source_parameters.0.activemq_broker_parameters.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "source_parameters.0.dynamodb_stream_parameters.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "source_parameters.0.dynamodb_stream_parameters.0.batch_size", "100"),
					resource.TestCheckResourceAttr(resourceName, "source_parameters.0.dynamodb_stream_parameters.0.dead_letter_config.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "source_parameters.0.dynamodb_stream_parameters.0.maximum_batching_window_in_seconds", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "source_parameters.0.dynamodb_stream_parameters.0.maximum_record_age_in_seconds", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "source_parameters.0.dynamodb_stream_parameters.0.maximum_retry_attempts", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "source_parameters.0.dynamodb_stream_parameters.0.on_partial_batch_item_failure", ""),
					resource.TestCheckResourceAttr(resourceName, "source_parameters.0.dynamodb_stream_parameters.0.parallelization_factor", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "source_parameters.0.dynamodb_stream_parameters.0.starting_position", "LATEST"),
					resource.TestCheckResourceAttr(resourceName, "source_parameters.0.filter_criteria.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "source_parameters.0.managed_streaming_kafka_parameters.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "source_parameters.0.rabbitmq_broker_parameters.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "source_parameters.0.self_managed_kafka_parameters.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "source_parameters.0.sqs_queue_parameters.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct0),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrTarget, "aws_cloudwatch_log_group.target", names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, "target_parameters.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "target_parameters.0.batch_job_parameters.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "target_parameters.0.cloudwatch_logs_parameters.#", acctest.Ct1),
					resource.TestCheckResourceAttrPair(resourceName, "target_parameters.0.cloudwatch_logs_parameters.0.log_stream_name", "aws_cloudwatch_log_stream.target", names.AttrName),
					resource.TestCheckResourceAttr(resourceName, "target_parameters.0.cloudwatch_logs_parameters.0.timestamp", ""),
					resource.TestCheckResourceAttr(resourceName, "target_parameters.0.ecs_task_parameters.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "target_parameters.0.eventbridge_event_bus_parameters.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "target_parameters.0.http_parameters.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "target_parameters.0.input_template", ""),
					resource.TestCheckResourceAttr(resourceName, "target_parameters.0.kinesis_stream_parameters.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "target_parameters.0.lambda_function_parameters.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "target_parameters.0.redshift_data_parameters.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "target_parameters.0.sagemaker_pipeline_parameters.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "target_parameters.0.sqs_queue_parameters.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "target_parameters.0.step_function_state_machine_parameters.#", acctest.Ct0),
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

func TestAccPipesPipe_activeMQSourceStepFunctionTarget(t *testing.T) {
	ctx := acctest.Context(t)
	var pipe pipes.DescribePipeOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_pipes_pipe.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.PipesEndpointID)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.PipesServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckPipeDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccPipeConfig_basicActiveMQSourceStepFunctionTarget(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckPipeExists(ctx, resourceName, &pipe),
					acctest.MatchResourceAttrRegionalARN(resourceName, names.AttrARN, "pipes", regexache.MustCompile(regexp.QuoteMeta(`pipe/`+rName))),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, "Managed by Terraform"),
					resource.TestCheckResourceAttr(resourceName, "desired_state", "RUNNING"),
					resource.TestCheckResourceAttr(resourceName, "enrichment", ""),
					resource.TestCheckResourceAttr(resourceName, "enrichment_parameters.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrRoleARN, "aws_iam_role.test", names.AttrARN),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrSource, "aws_mq_broker.source", names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, "source_parameters.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "source_parameters.0.activemq_broker_parameters.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "source_parameters.0.activemq_broker_parameters.0.batch_size", "100"),
					resource.TestCheckResourceAttr(resourceName, "source_parameters.0.activemq_broker_parameters.0.credentials.#", acctest.Ct1),
					resource.TestCheckResourceAttrSet(resourceName, "source_parameters.0.activemq_broker_parameters.0.credentials.0.basic_auth"),
					resource.TestCheckResourceAttr(resourceName, "source_parameters.0.activemq_broker_parameters.0.maximum_batching_window_in_seconds", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "source_parameters.0.activemq_broker_parameters.0.queue_name", "test"),
					resource.TestCheckResourceAttr(resourceName, "source_parameters.0.dynamodb_stream_parameters.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "source_parameters.0.filter_criteria.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "source_parameters.0.managed_streaming_kafka_parameters.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "source_parameters.0.rabbitmq_broker_parameters.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "source_parameters.0.self_managed_kafka_parameters.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "source_parameters.0.sqs_queue_parameters.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct0),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrTarget, "aws_sfn_state_machine.target", names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, "target_parameters.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "target_parameters.0.batch_job_parameters.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "target_parameters.0.cloudwatch_logs_parameters.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "target_parameters.0.ecs_task_parameters.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "target_parameters.0.eventbridge_event_bus_parameters.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "target_parameters.0.http_parameters.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "target_parameters.0.input_template", ""),
					resource.TestCheckResourceAttr(resourceName, "target_parameters.0.kinesis_stream_parameters.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "target_parameters.0.lambda_function_parameters.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "target_parameters.0.redshift_data_parameters.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "target_parameters.0.sagemaker_pipeline_parameters.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "target_parameters.0.sqs_queue_parameters.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "target_parameters.0.step_function_state_machine_parameters.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "target_parameters.0.step_function_state_machine_parameters.0.invocation_type", "REQUEST_RESPONSE"),
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

func TestAccPipesPipe_rabbitMQSourceEventBusTarget(t *testing.T) {
	ctx := acctest.Context(t)
	var pipe pipes.DescribePipeOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_pipes_pipe.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.PipesEndpointID)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.PipesServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckPipeDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccPipeConfig_basicRabbitMQSourceEventBusTarget(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckPipeExists(ctx, resourceName, &pipe),
					acctest.MatchResourceAttrRegionalARN(resourceName, names.AttrARN, "pipes", regexache.MustCompile(regexp.QuoteMeta(`pipe/`+rName))),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, "Managed by Terraform"),
					resource.TestCheckResourceAttr(resourceName, "desired_state", "RUNNING"),
					resource.TestCheckResourceAttr(resourceName, "enrichment", ""),
					resource.TestCheckResourceAttr(resourceName, "enrichment_parameters.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrRoleARN, "aws_iam_role.test", names.AttrARN),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrSource, "aws_mq_broker.source", names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, "source_parameters.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "source_parameters.0.activemq_broker_parameters.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "source_parameters.0.dynamodb_stream_parameters.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "source_parameters.0.filter_criteria.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "source_parameters.0.managed_streaming_kafka_parameters.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "source_parameters.0.rabbitmq_broker_parameters.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "source_parameters.0.rabbitmq_broker_parameters.0.batch_size", acctest.Ct10),
					resource.TestCheckResourceAttr(resourceName, "source_parameters.0.rabbitmq_broker_parameters.0.credentials.#", acctest.Ct1),
					resource.TestCheckResourceAttrSet(resourceName, "source_parameters.0.rabbitmq_broker_parameters.0.credentials.0.basic_auth"),
					resource.TestCheckResourceAttr(resourceName, "source_parameters.0.rabbitmq_broker_parameters.0.maximum_batching_window_in_seconds", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "source_parameters.0.rabbitmq_broker_parameters.0.queue_name", "test"),
					resource.TestCheckResourceAttr(resourceName, "source_parameters.0.rabbitmq_broker_parameters.0.virtual_host", ""),
					resource.TestCheckResourceAttr(resourceName, "source_parameters.0.self_managed_kafka_parameters.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "source_parameters.0.sqs_queue_parameters.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct0),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrTarget, "aws_cloudwatch_event_bus.target", names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, "target_parameters.#", acctest.Ct0),
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

func TestAccPipesPipe_mskSourceHTTPTarget(t *testing.T) {
	acctest.Skip(t, "DependencyViolation errors deleting subnets and security group")

	ctx := acctest.Context(t)
	var pipe pipes.DescribePipeOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_pipes_pipe.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.PipesEndpointID)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.PipesServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckPipeDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccPipeConfig_basicMSKSourceHTTPTarget(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckPipeExists(ctx, resourceName, &pipe),
					acctest.MatchResourceAttrRegionalARN(resourceName, names.AttrARN, "pipes", regexache.MustCompile(regexp.QuoteMeta(`pipe/`+rName))),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, "Managed by Terraform"),
					resource.TestCheckResourceAttr(resourceName, "desired_state", "RUNNING"),
					resource.TestCheckResourceAttr(resourceName, "enrichment", ""),
					resource.TestCheckResourceAttr(resourceName, "enrichment_parameters.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrRoleARN, "aws_iam_role.test", names.AttrARN),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrSource, "aws_msk_cluster.source", names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, "source_parameters.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "source_parameters.0.activemq_broker_parameters.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "source_parameters.0.dynamodb_stream_parameters.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "source_parameters.0.filter_criteria.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "source_parameters.0.managed_streaming_kafka_parameters.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "source_parameters.0.managed_streaming_kafka_parameters.0.batch_size", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "source_parameters.0.managed_streaming_kafka_parameters.0.consumer_group_id", ""),
					resource.TestCheckResourceAttr(resourceName, "source_parameters.0.managed_streaming_kafka_parameters.0.credentials.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "source_parameters.0.managed_streaming_kafka_parameters.0.maximum_batching_window_in_seconds", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "source_parameters.0.managed_streaming_kafka_parameters.0.starting_position", ""),
					resource.TestCheckResourceAttr(resourceName, "source_parameters.0.managed_streaming_kafka_parameters.0.topic_name", "test"),
					resource.TestCheckResourceAttr(resourceName, "source_parameters.0.rabbitmq_broker_parameters.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "source_parameters.0.self_managed_kafka_parameters.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "source_parameters.0.sqs_queue_parameters.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct0),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrTarget),
					resource.TestCheckResourceAttr(resourceName, "target_parameters.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "target_parameters.0.batch_job_parameters.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "target_parameters.0.cloudwatch_logs_parameters.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "target_parameters.0.ecs_task_parameters.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "target_parameters.0.eventbridge_event_bus_parameters.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "target_parameters.0.http_parameters.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "target_parameters.0.http_parameters.0.header_parameters.%", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "target_parameters.0.http_parameters.0.header_parameters.X-Test", "test"),
					resource.TestCheckResourceAttr(resourceName, "target_parameters.0.http_parameters.0.path_parameter_values.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "target_parameters.0.http_parameters.0.path_parameter_values.0", "p1"),
					resource.TestCheckResourceAttr(resourceName, "target_parameters.0.http_parameters.0.query_string_parameters.%", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "target_parameters.0.http_parameters.0.query_string_parameters.testing", "yes"),
					resource.TestCheckResourceAttr(resourceName, "target_parameters.0.input_template", ""),
					resource.TestCheckResourceAttr(resourceName, "target_parameters.0.kinesis_stream_parameters.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "target_parameters.0.lambda_function_parameters.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "target_parameters.0.redshift_data_parameters.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "target_parameters.0.sagemaker_pipeline_parameters.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "target_parameters.0.sqs_queue_parameters.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "target_parameters.0.step_function_state_machine_parameters.#", acctest.Ct0),
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

func TestAccPipesPipe_selfManagedKafkaSourceLambdaFunctionTarget(t *testing.T) {
	acctest.Skip(t, "DependencyViolation errors deleting subnets and security group")

	ctx := acctest.Context(t)
	var pipe pipes.DescribePipeOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_pipes_pipe.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.PipesEndpointID)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.PipesServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckPipeDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccPipeConfig_basicSelfManagedKafkaSourceLambdaFunctionTarget(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckPipeExists(ctx, resourceName, &pipe),
					acctest.MatchResourceAttrRegionalARN(resourceName, names.AttrARN, "pipes", regexache.MustCompile(regexp.QuoteMeta(`pipe/`+rName))),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, "Managed by Terraform"),
					resource.TestCheckResourceAttr(resourceName, "desired_state", "RUNNING"),
					resource.TestCheckResourceAttr(resourceName, "enrichment", ""),
					resource.TestCheckResourceAttr(resourceName, "enrichment_parameters.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrRoleARN, "aws_iam_role.test", names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, names.AttrSource, "smk://test1:9092,test2:9092"),
					resource.TestCheckResourceAttr(resourceName, "source_parameters.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "source_parameters.0.activemq_broker_parameters.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "source_parameters.0.dynamodb_stream_parameters.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "source_parameters.0.filter_criteria.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "source_parameters.0.managed_streaming_kafka_parameters.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "source_parameters.0.rabbitmq_broker_parameters.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "source_parameters.0.self_managed_kafka_parameters.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "source_parameters.0.self_managed_kafka_parameters.0.additional_bootstrap_servers.#", acctest.Ct1),
					resource.TestCheckTypeSetElemAttr(resourceName, "source_parameters.0.self_managed_kafka_parameters.0.additional_bootstrap_servers.*", "testing:1234"),
					resource.TestCheckResourceAttr(resourceName, "source_parameters.0.self_managed_kafka_parameters.0.batch_size", "100"),
					resource.TestCheckResourceAttr(resourceName, "source_parameters.0.self_managed_kafka_parameters.0.consumer_group_id", "self-managed-test-group-id"),
					resource.TestCheckResourceAttr(resourceName, "source_parameters.0.self_managed_kafka_parameters.0.credentials.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "source_parameters.0.self_managed_kafka_parameters.0.maximum_batching_window_in_seconds", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "source_parameters.0.self_managed_kafka_parameters.0.server_root_ca_certificate", ""),
					resource.TestCheckResourceAttr(resourceName, "source_parameters.0.self_managed_kafka_parameters.0.starting_position", ""),
					resource.TestCheckResourceAttr(resourceName, "source_parameters.0.self_managed_kafka_parameters.0.topic_name", "test"),
					resource.TestCheckResourceAttr(resourceName, "source_parameters.0.self_managed_kafka_parameters.0.vpc.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "source_parameters.0.self_managed_kafka_parameters.0.vpc.0.security_groups.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "source_parameters.0.self_managed_kafka_parameters.0.vpc.0.subnets.#", acctest.Ct2),
					resource.TestCheckResourceAttr(resourceName, "source_parameters.0.sqs_queue_parameters.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct0),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrTarget, "aws_lambda_function.target", names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, "target_parameters.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "target_parameters.0.batch_job_parameters.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "target_parameters.0.cloudwatch_logs_parameters.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "target_parameters.0.ecs_task_parameters.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "target_parameters.0.eventbridge_event_bus_parameters.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "target_parameters.0.http_parameters.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "target_parameters.0.input_template", ""),
					resource.TestCheckResourceAttr(resourceName, "target_parameters.0.kinesis_stream_parameters.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "target_parameters.0.lambda_function_parameters.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "target_parameters.0.lambda_function_parameters.0.invocation_type", "REQUEST_RESPONSE"),
					resource.TestCheckResourceAttr(resourceName, "target_parameters.0.redshift_data_parameters.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "target_parameters.0.sagemaker_pipeline_parameters.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "target_parameters.0.sqs_queue_parameters.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "target_parameters.0.step_function_state_machine_parameters.#", acctest.Ct0),
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

func TestAccPipesPipe_sqsSourceRedshiftTarget(t *testing.T) {
	ctx := acctest.Context(t)
	var pipe pipes.DescribePipeOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_pipes_pipe.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.PipesEndpointID)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.PipesServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckPipeDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccPipeConfig_basicSQSSourceRedshiftTarget(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckPipeExists(ctx, resourceName, &pipe),
					acctest.MatchResourceAttrRegionalARN(resourceName, names.AttrARN, "pipes", regexache.MustCompile(regexp.QuoteMeta(`pipe/`+rName))),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, "Managed by Terraform"),
					resource.TestCheckResourceAttr(resourceName, "desired_state", "RUNNING"),
					resource.TestCheckResourceAttr(resourceName, "enrichment", ""),
					resource.TestCheckResourceAttr(resourceName, "enrichment_parameters.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrRoleARN, "aws_iam_role.test", names.AttrARN),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrSource, "aws_sqs_queue.source", names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, "source_parameters.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "source_parameters.0.activemq_broker_parameters.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "source_parameters.0.dynamodb_stream_parameters.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "source_parameters.0.filter_criteria.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "source_parameters.0.managed_streaming_kafka_parameters.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "source_parameters.0.rabbitmq_broker_parameters.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "source_parameters.0.self_managed_kafka_parameters.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "source_parameters.0.sqs_queue_parameters.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "source_parameters.0.sqs_queue_parameters.0.batch_size", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "source_parameters.0.sqs_queue_parameters.0.maximum_batching_window_in_seconds", "90"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct0),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrTarget, "aws_redshift_cluster.target", names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, "target_parameters.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "target_parameters.0.batch_job_parameters.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "target_parameters.0.cloudwatch_logs_parameters.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "target_parameters.0.ecs_task_parameters.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "target_parameters.0.eventbridge_event_bus_parameters.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "target_parameters.0.http_parameters.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "target_parameters.0.input_template", ""),
					resource.TestCheckResourceAttr(resourceName, "target_parameters.0.kinesis_stream_parameters.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "target_parameters.0.lambda_function_parameters.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "target_parameters.0.redshift_data_parameters.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "target_parameters.0.redshift_data_parameters.0.database", "db1"),
					resource.TestCheckResourceAttr(resourceName, "target_parameters.0.redshift_data_parameters.0.db_user", "user1"),
					resource.TestCheckResourceAttr(resourceName, "target_parameters.0.redshift_data_parameters.0.secret_manager_arn", ""),
					resource.TestCheckResourceAttr(resourceName, "target_parameters.0.redshift_data_parameters.0.sqls.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "target_parameters.0.redshift_data_parameters.0.statement_name", "SelectAll"),
					resource.TestCheckResourceAttr(resourceName, "target_parameters.0.redshift_data_parameters.0.with_event", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "target_parameters.0.sagemaker_pipeline_parameters.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "target_parameters.0.sqs_queue_parameters.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "target_parameters.0.step_function_state_machine_parameters.#", acctest.Ct0),
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

func TestAccPipesPipe_SourceSageMakerTarget(t *testing.T) {
	acctest.Skip(t, "aws_sagemaker_pipeline resource not yet implemented")

	ctx := acctest.Context(t)
	var pipe pipes.DescribePipeOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_pipes_pipe.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.PipesEndpointID)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.PipesServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckPipeDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccPipeConfig_basicSQSSourceSageMakerTarget(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckPipeExists(ctx, resourceName, &pipe),
					acctest.MatchResourceAttrRegionalARN(resourceName, names.AttrARN, "pipes", regexache.MustCompile(regexp.QuoteMeta(`pipe/`+rName))),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, "Managed by Terraform"),
					resource.TestCheckResourceAttr(resourceName, "desired_state", "RUNNING"),
					resource.TestCheckResourceAttr(resourceName, "enrichment", ""),
					resource.TestCheckResourceAttr(resourceName, "enrichment_parameters.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrRoleARN, "aws_iam_role.test", names.AttrARN),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrSource, "aws_sqs_queue.source", names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, "source_parameters.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "source_parameters.0.activemq_broker_parameters.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "source_parameters.0.dynamodb_stream_parameters.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "source_parameters.0.filter_criteria.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "source_parameters.0.managed_streaming_kafka_parameters.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "source_parameters.0.rabbitmq_broker_parameters.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "source_parameters.0.self_managed_kafka_parameters.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "source_parameters.0.sqs_queue_parameters.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct0),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrTarget, "aws_sagemaker_pipeline.target", names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, "target_parameters.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "target_parameters.0.batch_job_parameters.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "target_parameters.0.cloudwatch_logs_parameters.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "target_parameters.0.ecs_task_parameters.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "target_parameters.0.eventbridge_event_bus_parameters.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "target_parameters.0.http_parameters.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "target_parameters.0.input_template", ""),
					resource.TestCheckResourceAttr(resourceName, "target_parameters.0.kinesis_stream_parameters.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "target_parameters.0.lambda_function_parameters.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "target_parameters.0.redshift_data_parameters.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "target_parameters.0.sagemaker_pipeline_parameters.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "target_parameters.0.sagemaker_pipeline_parameters.0.pipeline_parameter.#", acctest.Ct2),
					resource.TestCheckResourceAttr(resourceName, "target_parameters.0.sagemaker_pipeline_parameters.0.pipeline_parameter.0.name", "p1"),
					resource.TestCheckResourceAttr(resourceName, "target_parameters.0.sagemaker_pipeline_parameters.0.pipeline_parameter.0.value", "v1"),
					resource.TestCheckResourceAttr(resourceName, "target_parameters.0.sagemaker_pipeline_parameters.0.pipeline_parameter.1.name", "p2"),
					resource.TestCheckResourceAttr(resourceName, "target_parameters.0.sagemaker_pipeline_parameters.0.pipeline_parameter.1.value", "v2"),
					resource.TestCheckResourceAttr(resourceName, "target_parameters.0.sqs_queue_parameters.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "target_parameters.0.step_function_state_machine_parameters.#", acctest.Ct0),
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

func TestAccPipesPipe_sqsSourceBatchJobTarget(t *testing.T) {
	ctx := acctest.Context(t)
	var pipe pipes.DescribePipeOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_pipes_pipe.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.PipesEndpointID)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.PipesServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckPipeDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccPipeConfig_basicSQSSourceBatchJobTarget(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckPipeExists(ctx, resourceName, &pipe),
					acctest.MatchResourceAttrRegionalARN(resourceName, names.AttrARN, "pipes", regexache.MustCompile(regexp.QuoteMeta(`pipe/`+rName))),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, "Managed by Terraform"),
					resource.TestCheckResourceAttr(resourceName, "desired_state", "RUNNING"),
					resource.TestCheckResourceAttr(resourceName, "enrichment", ""),
					resource.TestCheckResourceAttr(resourceName, "enrichment_parameters.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrRoleARN, "aws_iam_role.test", names.AttrARN),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrSource, "aws_sqs_queue.source", names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, "source_parameters.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "source_parameters.0.activemq_broker_parameters.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "source_parameters.0.dynamodb_stream_parameters.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "source_parameters.0.filter_criteria.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "source_parameters.0.managed_streaming_kafka_parameters.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "source_parameters.0.rabbitmq_broker_parameters.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "source_parameters.0.self_managed_kafka_parameters.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "source_parameters.0.sqs_queue_parameters.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct0),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrTarget, "aws_batch_job_queue.target", names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, "target_parameters.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "target_parameters.0.batch_job_parameters.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "target_parameters.0.batch_job_parameters.0.array_properties.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "target_parameters.0.batch_job_parameters.0.array_properties.0.size", "512"),
					resource.TestCheckResourceAttr(resourceName, "target_parameters.0.batch_job_parameters.0.container_overrides.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "target_parameters.0.batch_job_parameters.0.container_overrides.0.command.#", acctest.Ct3),
					resource.TestCheckResourceAttr(resourceName, "target_parameters.0.batch_job_parameters.0.container_overrides.0.command.0", "rm"),
					resource.TestCheckResourceAttr(resourceName, "target_parameters.0.batch_job_parameters.0.container_overrides.0.command.1", "-fr"),
					resource.TestCheckResourceAttr(resourceName, "target_parameters.0.batch_job_parameters.0.container_overrides.0.command.2", "/"),
					resource.TestCheckResourceAttr(resourceName, "target_parameters.0.batch_job_parameters.0.container_overrides.0.environment.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "target_parameters.0.batch_job_parameters.0.container_overrides.0.environment.0.name", "TMP"),
					resource.TestCheckResourceAttr(resourceName, "target_parameters.0.batch_job_parameters.0.container_overrides.0.environment.0.value", "/tmp2"),
					resource.TestCheckResourceAttr(resourceName, "target_parameters.0.batch_job_parameters.0.container_overrides.0.instance_type", ""),
					resource.TestCheckResourceAttr(resourceName, "target_parameters.0.batch_job_parameters.0.container_overrides.0.resource_requirement.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "target_parameters.0.batch_job_parameters.0.container_overrides.0.resource_requirement.0.type", "GPU"),
					resource.TestCheckResourceAttr(resourceName, "target_parameters.0.batch_job_parameters.0.container_overrides.0.resource_requirement.0.value", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "target_parameters.0.batch_job_parameters.0.depends_on.#", acctest.Ct0),
					resource.TestCheckResourceAttrSet(resourceName, "target_parameters.0.batch_job_parameters.0.job_definition"),
					resource.TestCheckResourceAttr(resourceName, "target_parameters.0.batch_job_parameters.0.job_name", "testing"),
					resource.TestCheckResourceAttr(resourceName, "target_parameters.0.batch_job_parameters.0.parameters.%", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "target_parameters.0.batch_job_parameters.0.parameters.Key1", "Value1"),
					resource.TestCheckResourceAttr(resourceName, "target_parameters.0.batch_job_parameters.0.retry_strategy.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "target_parameters.0.cloudwatch_logs_parameters.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "target_parameters.0.ecs_task_parameters.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "target_parameters.0.eventbridge_event_bus_parameters.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "target_parameters.0.http_parameters.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "target_parameters.0.input_template", ""),
					resource.TestCheckResourceAttr(resourceName, "target_parameters.0.kinesis_stream_parameters.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "target_parameters.0.lambda_function_parameters.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "target_parameters.0.redshift_data_parameters.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "target_parameters.0.sagemaker_pipeline_parameters.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "target_parameters.0.sqs_queue_parameters.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "target_parameters.0.step_function_state_machine_parameters.#", acctest.Ct0),
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

func TestAccPipesPipe_sqsSourceECSTaskTarget(t *testing.T) {
	acctest.Skip(t, "ValidationException: [numeric instance is lower than the required minimum (minimum: 1, found: 0)]")

	ctx := acctest.Context(t)
	var pipe pipes.DescribePipeOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_pipes_pipe.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.PipesEndpointID)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.PipesServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckPipeDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccPipeConfig_basicSQSSourceECSTaskTarget(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckPipeExists(ctx, resourceName, &pipe),
					acctest.MatchResourceAttrRegionalARN(resourceName, names.AttrARN, "pipes", regexache.MustCompile(regexp.QuoteMeta(`pipe/`+rName))),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, "Managed by Terraform"),
					resource.TestCheckResourceAttr(resourceName, "desired_state", "RUNNING"),
					resource.TestCheckResourceAttr(resourceName, "enrichment", ""),
					resource.TestCheckResourceAttr(resourceName, "enrichment_parameters.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrRoleARN, "aws_iam_role.test", names.AttrARN),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrSource, "aws_sqs_queue.source", names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, "source_parameters.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "source_parameters.0.activemq_broker_parameters.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "source_parameters.0.dynamodb_stream_parameters.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "source_parameters.0.filter_criteria.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "source_parameters.0.managed_streaming_kafka_parameters.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "source_parameters.0.rabbitmq_broker_parameters.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "source_parameters.0.self_managed_kafka_parameters.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "source_parameters.0.sqs_queue_parameters.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct0),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrTarget, "aws_ecs_cluster.target", names.AttrID),
					resource.TestCheckResourceAttr(resourceName, "target_parameters.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "target_parameters.0.batch_job_parameters.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "target_parameters.0.cloudwatch_logs_parameters.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "target_parameters.0.ecs_task_parameters.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "target_parameters.0.ecs_task_parameters.0.capacity_provider_strategy.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "target_parameters.0.ecs_task_parameters.0.enable_ecs_managed_tags", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "target_parameters.0.ecs_task_parameters.0.enable_execute_command", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "target_parameters.0.ecs_task_parameters.0.group", "g1"),
					resource.TestCheckResourceAttr(resourceName, "target_parameters.0.ecs_task_parameters.0.launch_type", "FARGATE"),
					resource.TestCheckResourceAttr(resourceName, "target_parameters.0.ecs_task_parameters.0.network_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "target_parameters.0.ecs_task_parameters.0.network_configuration.0.aws_vpc_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "target_parameters.0.ecs_task_parameters.0.network_configuration.0.aws_vpc_configuration.0.assign_public_ip", ""),
					resource.TestCheckResourceAttr(resourceName, "target_parameters.0.ecs_task_parameters.0.network_configuration.0.aws_vpc_configuration.0.security_groups.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "target_parameters.0.ecs_task_parameters.0.network_configuration.0.aws_vpc_configuration.0.subnets.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "target_parameters.0.ecs_task_parameters.0.overrides.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "target_parameters.0.ecs_task_parameters.0.overrides.0.container_override.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "target_parameters.0.ecs_task_parameters.0.overrides.0.container_override.0.command.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "target_parameters.0.ecs_task_parameters.0.overrides.0.container_override.0.cpu", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "target_parameters.0.ecs_task_parameters.0.overrides.0.container_override.0.environment.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "target_parameters.0.ecs_task_parameters.0.overrides.0.container_override.0.environment.0.name", "TMP"),
					resource.TestCheckResourceAttr(resourceName, "target_parameters.0.ecs_task_parameters.0.overrides.0.container_override.0.environment.0.value", "/tmp2"),
					resource.TestCheckResourceAttr(resourceName, "target_parameters.0.ecs_task_parameters.0.overrides.0.container_override.0.environment_file.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "target_parameters.0.ecs_task_parameters.0.overrides.0.container_override.0.memory", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "target_parameters.0.ecs_task_parameters.0.overrides.0.container_override.0.memory_reservation", "1024"),
					resource.TestCheckResourceAttr(resourceName, "target_parameters.0.ecs_task_parameters.0.overrides.0.container_override.0.name", "first"),
					resource.TestCheckResourceAttr(resourceName, "target_parameters.0.ecs_task_parameters.0.overrides.0.container_override.0.resource_requirement.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "target_parameters.0.ecs_task_parameters.0.overrides.0.container_override.0.resource_requirement.0.type", "GPU"),
					resource.TestCheckResourceAttr(resourceName, "target_parameters.0.ecs_task_parameters.0.overrides.0.container_override.0.resource_requirement.0.value", acctest.Ct2),
					resource.TestCheckResourceAttr(resourceName, "target_parameters.0.ecs_task_parameters.0.overrides.0.cpu", ""),
					resource.TestCheckResourceAttr(resourceName, "target_parameters.0.ecs_task_parameters.0.overrides.0.ephemeral_storage.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "target_parameters.0.ecs_task_parameters.0.overrides.0.ephemeral_storage.0.size_in_gib", "32"),
					resource.TestCheckResourceAttr(resourceName, "target_parameters.0.ecs_task_parameters.0.overrides.0.execution_role_arn", ""),
					resource.TestCheckResourceAttr(resourceName, "target_parameters.0.ecs_task_parameters.0.overrides.0.inference_accelerator_override.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "target_parameters.0.ecs_task_parameters.0.overrides.0.memory", ""),
					resource.TestCheckResourceAttr(resourceName, "target_parameters.0.ecs_task_parameters.0.overrides.0.task_role_arn", ""),
					resource.TestCheckResourceAttr(resourceName, "target_parameters.0.ecs_task_parameters.0.placement_constraint.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "target_parameters.0.ecs_task_parameters.0.placement_strategy.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "target_parameters.0.ecs_task_parameters.0.placement_strategy.0.field", "cpu"),
					resource.TestCheckResourceAttr(resourceName, "target_parameters.0.ecs_task_parameters.0.placement_strategy.0.type", "binpack"),
					resource.TestCheckResourceAttr(resourceName, "target_parameters.0.ecs_task_parameters.0.platform_version", ""),
					resource.TestCheckResourceAttr(resourceName, "target_parameters.0.ecs_task_parameters.0.propagate_tags", "TASK_DEFINITION"),
					resource.TestCheckResourceAttr(resourceName, "target_parameters.0.ecs_task_parameters.0.reference_id", "refid"),
					resource.TestCheckResourceAttr(resourceName, "target_parameters.0.ecs_task_parameters.0.tags.%", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "target_parameters.0.ecs_task_parameters.0.tags.Name", rName),
					resource.TestCheckResourceAttr(resourceName, "target_parameters.0.ecs_task_parameters.0.task_count", acctest.Ct1),
					resource.TestCheckResourceAttrSet(resourceName, "target_parameters.0.ecs_task_parameters.0.task_definition_arn"),
					resource.TestCheckResourceAttr(resourceName, "target_parameters.0.eventbridge_event_bus_parameters.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "target_parameters.0.http_parameters.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "target_parameters.0.input_template", ""),
					resource.TestCheckResourceAttr(resourceName, "target_parameters.0.kinesis_stream_parameters.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "target_parameters.0.lambda_function_parameters.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "target_parameters.0.redshift_data_parameters.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "target_parameters.0.sagemaker_pipeline_parameters.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "target_parameters.0.sqs_queue_parameters.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "target_parameters.0.step_function_state_machine_parameters.#", acctest.Ct0),
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

func testAccCheckPipeDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).PipesClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_pipes_pipe" {
				continue
			}

			_, err := tfpipes.FindPipeByName(ctx, conn, rs.Primary.ID)

			if tfresource.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return create.Error(names.Pipes, create.ErrActionCheckingDestroyed, tfpipes.ResNamePipe, rs.Primary.ID, errors.New("not destroyed"))
		}

		return nil
	}
}

func testAccCheckPipeExists(ctx context.Context, name string, pipe *pipes.DescribePipeOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return create.Error(names.Pipes, create.ErrActionCheckingExistence, tfpipes.ResNamePipe, name, errors.New("not found"))
		}

		if rs.Primary.ID == "" {
			return create.Error(names.Pipes, create.ErrActionCheckingExistence, tfpipes.ResNamePipe, name, errors.New("not set"))
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).PipesClient(ctx)

		output, err := tfpipes.FindPipeByName(ctx, conn, rs.Primary.ID)

		if err != nil {
			return err
		}

		*pipe = *output

		return nil
	}
}

func testAccPreCheck(ctx context.Context, t *testing.T) {
	conn := acctest.Provider.Meta().(*conns.AWSClient).PipesClient(ctx)

	input := &pipes.ListPipesInput{}
	_, err := conn.ListPipes(ctx, input)

	if acctest.PreCheckSkipError(err) {
		t.Skipf("skipping acceptance testing: %s", err)
	}

	if err != nil {
		t.Fatalf("unexpected PreCheck error: %s", err)
	}
}

func testAccPipeConfig_base(rName string) string {
	return fmt.Sprintf(`
data "aws_caller_identity" "main" {}
data "aws_partition" "main" {}

resource "aws_iam_role" "test" {
  name = %[1]q

  assume_role_policy = jsonencode({
    Version = "2012-10-17"
    Statement = {
      Effect = "Allow"
      Action = "sts:AssumeRole"
      Principal = {
        Service = "pipes.${data.aws_partition.main.dns_suffix}"
      }
      Condition = {
        StringEquals = {
          "aws:SourceAccount" = data.aws_caller_identity.main.account_id
        }
      }
    }
  })
}
`, rName)
}

func testAccPipeConfig_baseSQSSource(rName string) string {
	return fmt.Sprintf(`
resource "aws_iam_role_policy" "source" {
  role = aws_iam_role.test.id
  name = "%[1]s-source"

  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Effect = "Allow"
        Action = [
          "sqs:DeleteMessage",
          "sqs:GetQueueAttributes",
          "sqs:ReceiveMessage",
        ],
        Resource = [
          aws_sqs_queue.source.arn,
        ]
      },
    ]
  })
}

resource "aws_sqs_queue" "source" {
  name = "%[1]s-source"
}
`, rName)
}

func testAccPipeConfig_baseSQSTarget(rName string) string {
	return fmt.Sprintf(`
resource "aws_iam_role_policy" "target" {
  role = aws_iam_role.test.id
  name = "%[1]s-target"

  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Effect = "Allow"
        Action = [
          "sqs:SendMessage",
        ],
        Resource = [
          aws_sqs_queue.target.arn,
        ]
      },
    ]
  })
}

resource "aws_sqs_queue" "target" {
  name = "%[1]s-target"
}
`, rName)
}

func testAccPipeConfig_baseKinesisSource(rName string) string {
	return fmt.Sprintf(`
resource "aws_iam_role_policy" "source" {
  role = aws_iam_role.test.id
  name = "%[1]s-source"

  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Effect = "Allow"
        Action = [
          "kinesis:DescribeStream",
          "kinesis:DescribeStreamSummary",
          "kinesis:GetRecords",
          "kinesis:GetShardIterator",
          "kinesis:ListShards",
          "kinesis:ListStreams",
          "kinesis:SubscribeToShard",
        ],
        Resource = [
          aws_kinesis_stream.source.arn,
        ]
      },
    ]
  })
}

resource "aws_kinesis_stream" "source" {
  name = "%[1]s-source"

  stream_mode_details {
    stream_mode = "ON_DEMAND"
  }
}
`, rName)
}

func testAccPipeConfig_baseKinesisTarget(rName string) string {
	return fmt.Sprintf(`
resource "aws_iam_role_policy" "target" {
  role = aws_iam_role.test.id
  name = "%[1]s-target"

  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Effect = "Allow"
        Action = [
          "kinesis:PutRecord",
        ],
        Resource = [
          aws_kinesis_stream.target.arn,
        ]
      },
    ]
  })
}

resource "aws_kinesis_stream" "target" {
  name = "%[1]s-target"

  stream_mode_details {
    stream_mode = "ON_DEMAND"
  }
}
`, rName)
}

func testAccPipeConfig_basicSQS(rName string) string {
	return acctest.ConfigCompose(
		testAccPipeConfig_base(rName),
		testAccPipeConfig_baseSQSSource(rName),
		testAccPipeConfig_baseSQSTarget(rName),
		fmt.Sprintf(`
resource "aws_pipes_pipe" "test" {
  depends_on = [aws_iam_role_policy.source, aws_iam_role_policy.target]

  name     = %[1]q
  role_arn = aws_iam_role.test.arn
  source   = aws_sqs_queue.source.arn
  target   = aws_sqs_queue.target.arn
}
`, rName))
}

func testAccPipeConfig_description(rName, description string) string {
	return acctest.ConfigCompose(
		testAccPipeConfig_base(rName),
		testAccPipeConfig_baseSQSSource(rName),
		testAccPipeConfig_baseSQSTarget(rName),
		fmt.Sprintf(`
resource "aws_pipes_pipe" "test" {
  depends_on = [aws_iam_role_policy.source, aws_iam_role_policy.target]

  name     = %[1]q
  role_arn = aws_iam_role.test.arn
  source   = aws_sqs_queue.source.arn
  target   = aws_sqs_queue.target.arn

  description = %[2]q
}
`, rName, description))
}

func testAccPipeConfig_desiredState(rName, state string) string {
	return acctest.ConfigCompose(
		testAccPipeConfig_base(rName),
		testAccPipeConfig_baseSQSSource(rName),
		testAccPipeConfig_baseSQSTarget(rName),
		fmt.Sprintf(`
resource "aws_pipes_pipe" "test" {
  depends_on = [aws_iam_role_policy.source, aws_iam_role_policy.target]

  name     = %[1]q
  role_arn = aws_iam_role.test.arn
  source   = aws_sqs_queue.source.arn
  target   = aws_sqs_queue.target.arn

  desired_state = %[2]q
}
`, rName, state))
}

func testAccPipeConfig_enrichment(rName string, i int) string {
	return acctest.ConfigCompose(
		testAccPipeConfig_base(rName),
		testAccPipeConfig_baseSQSSource(rName),
		testAccPipeConfig_baseSQSTarget(rName),
		fmt.Sprintf(`
resource "aws_cloudwatch_event_connection" "test" {
  name               = %[1]q
  authorization_type = "API_KEY"

  auth_parameters {
    api_key {
      key   = "testKey"
      value = "testValue"
    }
  }
}

resource "aws_cloudwatch_event_api_destination" "test" {
  count               = 2
  name                = "%[1]s-${count.index}"
  invocation_endpoint = "https://example.com/${count.index}"
  http_method         = "POST"
  connection_arn      = aws_cloudwatch_event_connection.test.arn
}

resource "aws_pipes_pipe" "test" {
  depends_on = [aws_iam_role_policy.source, aws_iam_role_policy.target]

  name     = %[1]q
  role_arn = aws_iam_role.test.arn
  source   = aws_sqs_queue.source.arn
  target   = aws_sqs_queue.target.arn

  enrichment = aws_cloudwatch_event_api_destination.test[%[2]d].arn
}
`, rName, i))
}

func testAccPipeConfig_enrichmentParameters(rName string) string {
	return acctest.ConfigCompose(
		testAccPipeConfig_base(rName),
		testAccPipeConfig_baseSQSSource(rName),
		testAccPipeConfig_baseSQSTarget(rName),
		fmt.Sprintf(`
resource "aws_cloudwatch_event_connection" "test" {
  name               = %[1]q
  authorization_type = "API_KEY"

  auth_parameters {
    api_key {
      key   = "testKey"
      value = "testValue"
    }
  }
}

resource "aws_cloudwatch_event_api_destination" "test" {
  name                = %[1]q
  invocation_endpoint = "https://example.com/"
  http_method         = "POST"
  connection_arn      = aws_cloudwatch_event_connection.test.arn
}

resource "aws_pipes_pipe" "test" {
  depends_on = [aws_iam_role_policy.source, aws_iam_role_policy.target]

  name     = %[1]q
  role_arn = aws_iam_role.test.arn
  source   = aws_sqs_queue.source.arn
  target   = aws_sqs_queue.target.arn

  enrichment = aws_cloudwatch_event_api_destination.test.arn

  enrichment_parameters {
    http_parameters {
      header_parameters = {
        "X-Test-1" = "Val1"
      }

      path_parameter_values = ["p1"]

      query_string_parameters = {
        "q1" = "abc"
      }
    }
  }
}
`, rName))
}

func testAccPipeConfig_enrichmentParametersUpdated(rName string) string {
	return acctest.ConfigCompose(
		testAccPipeConfig_base(rName),
		testAccPipeConfig_baseSQSSource(rName),
		testAccPipeConfig_baseSQSTarget(rName),
		fmt.Sprintf(`
resource "aws_cloudwatch_event_connection" "test" {
  name               = %[1]q
  authorization_type = "API_KEY"

  auth_parameters {
    api_key {
      key   = "testKey"
      value = "testValue"
    }
  }
}

resource "aws_cloudwatch_event_api_destination" "test" {
  name                = %[1]q
  invocation_endpoint = "https://example.com/"
  http_method         = "POST"
  connection_arn      = aws_cloudwatch_event_connection.test.arn
}

resource "aws_pipes_pipe" "test" {
  depends_on = [aws_iam_role_policy.source, aws_iam_role_policy.target]

  name     = %[1]q
  role_arn = aws_iam_role.test.arn
  source   = aws_sqs_queue.source.arn
  target   = aws_sqs_queue.target.arn

  enrichment = aws_cloudwatch_event_api_destination.test.arn

  enrichment_parameters {
    http_parameters {
      header_parameters = {
        "X-Test-1" = "Val1"
        "X-Test-2" = "Val2"
      }

      path_parameter_values = ["p2"]
    }
  }
}
`, rName))
}

func testAccPipeConfig_logConfiguration_cloudwatchLogsLogDestination(rName string) string {
	return acctest.ConfigCompose(
		testAccPipeConfig_base(rName),
		testAccPipeConfig_baseSQSSource(rName),
		testAccPipeConfig_baseSQSTarget(rName),
		fmt.Sprintf(`
resource "aws_pipes_pipe" "test" {
  depends_on = [aws_iam_role_policy.source, aws_iam_role_policy.target]

  name     = %[1]q
  role_arn = aws_iam_role.test.arn
  source   = aws_sqs_queue.source.arn
  target   = aws_sqs_queue.target.arn
  log_configuration {
    level = "INFO"
    cloudwatch_logs_log_destination {
      log_group_arn = aws_cloudwatch_log_group.target.arn
    }
  }
}

resource "aws_cloudwatch_log_group" "target" {
  name = "%[1]s-target"
}
`, rName))
}

func testAccPipeConfig_logConfiguration_update_cloudwatchLogsLogDestination(rName string) string {
	return acctest.ConfigCompose(
		testAccPipeConfig_base(rName),
		testAccPipeConfig_baseSQSSource(rName),
		testAccPipeConfig_baseSQSTarget(rName),
		fmt.Sprintf(`
resource "aws_pipes_pipe" "test" {
  depends_on = [aws_iam_role_policy.source, aws_iam_role_policy.target]

  name     = %[1]q
  role_arn = aws_iam_role.test.arn
  source   = aws_sqs_queue.source.arn
  target   = aws_sqs_queue.target.arn
  log_configuration {
    level = "ERROR"
    cloudwatch_logs_log_destination {
      log_group_arn = aws_cloudwatch_log_group.target.arn
    }
  }
}

resource "aws_cloudwatch_log_group" "target" {
  name = "%[1]s-target"
}
`, rName))
}

func testAccPipeConfig_sourceParameters_filterCriteria1(rName, criteria1 string) string {
	return acctest.ConfigCompose(
		testAccPipeConfig_base(rName),
		testAccPipeConfig_baseSQSSource(rName),
		testAccPipeConfig_baseSQSTarget(rName),
		fmt.Sprintf(`
resource "aws_pipes_pipe" "test" {
  depends_on = [aws_iam_role_policy.source, aws_iam_role_policy.target]

  name     = %[1]q
  role_arn = aws_iam_role.test.arn
  source   = aws_sqs_queue.source.arn
  target   = aws_sqs_queue.target.arn

  source_parameters {
    filter_criteria {
      filter {
        pattern = jsonencode({
          source = [%[2]q]
        })
      }
    }
  }
}
`, rName, criteria1))
}

func testAccPipeConfig_sourceParameters_filterCriteria2(rName, criteria1, criteria2 string) string {
	return acctest.ConfigCompose(
		testAccPipeConfig_base(rName),
		testAccPipeConfig_baseSQSSource(rName),
		testAccPipeConfig_baseSQSTarget(rName),
		fmt.Sprintf(`
resource "aws_pipes_pipe" "test" {
  depends_on = [aws_iam_role_policy.source, aws_iam_role_policy.target]

  name     = %[1]q
  role_arn = aws_iam_role.test.arn
  source   = aws_sqs_queue.source.arn
  target   = aws_sqs_queue.target.arn

  source_parameters {
    filter_criteria {
      filter {
        pattern = jsonencode({
          source = [%[2]q]
        })
      }

      filter {
        pattern = jsonencode({
          source = [%[3]q]
        })
      }
    }
  }
}
`, rName, criteria1, criteria2))
}

func testAccPipeConfig_sourceParameters_filterCriteria0(rName string) string {
	return acctest.ConfigCompose(
		testAccPipeConfig_base(rName),
		testAccPipeConfig_baseSQSSource(rName),
		testAccPipeConfig_baseSQSTarget(rName),
		fmt.Sprintf(`
resource "aws_pipes_pipe" "test" {
  depends_on = [aws_iam_role_policy.source, aws_iam_role_policy.target]
  name       = %[1]q
  role_arn   = aws_iam_role.test.arn
  source     = aws_sqs_queue.source.arn
  target     = aws_sqs_queue.target.arn

  source_parameters {
    filter_criteria {}
  }
}
`, rName))
}

func testAccPipeConfig_nameGenerated(rName string) string {
	return acctest.ConfigCompose(
		testAccPipeConfig_base(rName),
		testAccPipeConfig_baseSQSSource(rName),
		testAccPipeConfig_baseSQSTarget(rName),
		`
resource "aws_pipes_pipe" "test" {
  depends_on = [aws_iam_role_policy.source, aws_iam_role_policy.target]

  role_arn = aws_iam_role.test.arn
  source   = aws_sqs_queue.source.arn
  target   = aws_sqs_queue.target.arn
}
`,
	)
}

func testAccPipeConfig_namePrefix(rName, namePrefix string) string {
	return acctest.ConfigCompose(
		testAccPipeConfig_base(rName),
		testAccPipeConfig_baseSQSSource(rName),
		testAccPipeConfig_baseSQSTarget(rName),
		fmt.Sprintf(`
resource "aws_pipes_pipe" "test" {
  depends_on = [aws_iam_role_policy.source, aws_iam_role_policy.target]

  name_prefix = %[1]q
  role_arn    = aws_iam_role.test.arn
  source      = aws_sqs_queue.source.arn
  target      = aws_sqs_queue.target.arn
}
`, namePrefix))
}

func testAccPipeConfig_roleARN(rName string) string {
	return acctest.ConfigCompose(
		testAccPipeConfig_base(rName),
		testAccPipeConfig_baseSQSSource(rName),
		testAccPipeConfig_baseSQSTarget(rName),
		fmt.Sprintf(`
resource "aws_iam_role" "test2" {
  name = "%[1]s-2"

  assume_role_policy = jsonencode({
    Version = "2012-10-17"
    Statement = {
      Effect = "Allow"
      Action = "sts:AssumeRole"
      Principal = {
        Service = "pipes.${data.aws_partition.main.dns_suffix}"
      }
      Condition = {
        StringEquals = {
          "aws:SourceAccount" = data.aws_caller_identity.main.account_id
        }
      }
    }
  })
}

resource "aws_iam_role_policy" "source2" {
  role = aws_iam_role.test2.id
  name = "%[1]s-source2"

  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Effect = "Allow"
        Action = [
          "sqs:DeleteMessage",
          "sqs:GetQueueAttributes",
          "sqs:ReceiveMessage",
        ],
        Resource = [
          aws_sqs_queue.source.arn,
        ]
      },
    ]
  })
}

resource "aws_pipes_pipe" "test" {
  depends_on = [aws_iam_role_policy.source2, aws_iam_role_policy.target]

  name     = %[1]q
  role_arn = aws_iam_role.test2.arn
  source   = aws_sqs_queue.source.arn
  target   = aws_sqs_queue.target.arn
}
`, rName))
}

func testAccPipeConfig_tags1(rName, tag1Key, tag1Value string) string {
	return acctest.ConfigCompose(
		testAccPipeConfig_base(rName),
		testAccPipeConfig_baseSQSSource(rName),
		testAccPipeConfig_baseSQSTarget(rName),
		fmt.Sprintf(`
resource "aws_pipes_pipe" "test" {
  depends_on = [aws_iam_role_policy.source, aws_iam_role_policy.target]

  name     = %[1]q
  role_arn = aws_iam_role.test.arn
  source   = aws_sqs_queue.source.arn
  target   = aws_sqs_queue.target.arn

  tags = {
    %[2]q = %[3]q
  }
}
`, rName, tag1Key, tag1Value))
}

func testAccPipeConfig_tags2(rName, tag1Key, tag1Value, tag2Key, tag2Value string) string {
	return acctest.ConfigCompose(
		testAccPipeConfig_base(rName),
		testAccPipeConfig_baseSQSSource(rName),
		testAccPipeConfig_baseSQSTarget(rName),
		fmt.Sprintf(`
resource "aws_pipes_pipe" "test" {
  depends_on = [aws_iam_role_policy.source, aws_iam_role_policy.target]

  name     = %[1]q
  role_arn = aws_iam_role.test.arn
  source   = aws_sqs_queue.source.arn
  target   = aws_sqs_queue.target.arn

  tags = {
    %[2]q = %[3]q
    %[4]q = %[5]q
  }
}
`, rName, tag1Key, tag1Value, tag2Key, tag2Value))
}

func testAccPipeConfig_targetUpdated(rName string) string {
	return acctest.ConfigCompose(
		testAccPipeConfig_base(rName),
		testAccPipeConfig_baseSQSSource(rName),
		testAccPipeConfig_baseSQSTarget(rName),
		fmt.Sprintf(`
resource "aws_iam_role_policy" "target2" {
  role = aws_iam_role.test.id
  name = "%[1]s-target2"

  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Effect = "Allow"
        Action = [
          "sqs:SendMessage",
        ],
        Resource = [
          aws_sqs_queue.target2.arn,
        ]
      },
    ]
  })
}

resource "aws_sqs_queue" "target2" {
  name = "%[1]s-target2"
}

resource "aws_pipes_pipe" "test" {
  depends_on = [aws_iam_role_policy.source, aws_iam_role_policy.target2]

  name     = %[1]q
  role_arn = aws_iam_role.test.arn
  source   = aws_sqs_queue.source.arn
  target   = aws_sqs_queue.target2.arn
}
`, rName))
}

func testAccPipeConfig_targetParameters_inputTemplate(rName, template string) string {
	return acctest.ConfigCompose(
		testAccPipeConfig_base(rName),
		testAccPipeConfig_baseSQSSource(rName),
		testAccPipeConfig_baseSQSTarget(rName),
		fmt.Sprintf(`
resource "aws_pipes_pipe" "test" {
  depends_on = [aws_iam_role_policy.source, aws_iam_role_policy.target]

  name     = %[1]q
  role_arn = aws_iam_role.test.arn
  source   = aws_sqs_queue.source.arn
  target   = aws_sqs_queue.target.arn

  target_parameters {
    input_template = %[2]q
  }
}
`, rName, template))
}

func testAccPipeConfig_basicKinesis(rName string) string {
	return acctest.ConfigCompose(
		testAccPipeConfig_base(rName),
		testAccPipeConfig_baseKinesisSource(rName),
		testAccPipeConfig_baseKinesisTarget(rName),
		fmt.Sprintf(`
resource "aws_pipes_pipe" "test" {
  depends_on = [aws_iam_role_policy.source, aws_iam_role_policy.target]

  name     = %[1]q
  role_arn = aws_iam_role.test.arn
  source   = aws_kinesis_stream.source.arn
  target   = aws_kinesis_stream.target.arn

  source_parameters {
    kinesis_stream_parameters {
      starting_position = "LATEST"
    }
  }

  target_parameters {
    kinesis_stream_parameters {
      partition_key = "test"
    }
  }
}
`, rName))
}

func testAccPipeConfig_updateKinesis(rName string, batchSize int) string {
	return acctest.ConfigCompose(
		testAccPipeConfig_base(rName),
		testAccPipeConfig_baseKinesisSource(rName),
		testAccPipeConfig_baseKinesisTarget(rName),
		fmt.Sprintf(`
resource "aws_pipes_pipe" "test" {
  depends_on = [aws_iam_role_policy.source, aws_iam_role_policy.target]

  name     = %[1]q
  role_arn = aws_iam_role.test.arn
  source   = aws_kinesis_stream.source.arn
  target   = aws_kinesis_stream.target.arn

  source_parameters {
    kinesis_stream_parameters {
      batch_size        = %[2]d
      starting_position = "LATEST"
    }
  }

  target_parameters {
    kinesis_stream_parameters {
      partition_key = "test"
    }
  }
}
`, rName, batchSize))
}

func testAccPipeConfig_basicDynamoDBSourceCloudWatchLogsTarget(rName string) string {
	return acctest.ConfigCompose(
		testAccPipeConfig_base(rName),
		fmt.Sprintf(`
resource "aws_iam_role_policy" "source" {
  role = aws_iam_role.test.id
  name = "%[1]s-source"

  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Effect = "Allow"
        Action = [
          "dynamodb:DescribeStream",
          "dynamodb:GetRecords",
          "dynamodb:GetShardIterator",
          "dynamodb:ListStreams",
        ],
        Resource = [
          aws_dynamodb_table.source.stream_arn,
          "${aws_dynamodb_table.source.stream_arn}/*"
        ]
      },
    ]
  })
}

resource "aws_dynamodb_table" "source" {
  name             = "%[1]s-source"
  billing_mode     = "PAY_PER_REQUEST"
  hash_key         = "PK"
  range_key        = "SK"
  stream_enabled   = true
  stream_view_type = "NEW_AND_OLD_IMAGES"

  attribute {
    name = "PK"
    type = "S"
  }

  attribute {
    name = "SK"
    type = "S"
  }
}

resource "aws_iam_role_policy" "target" {
  role = aws_iam_role.test.id
  name = "%[1]s-target"

  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Effect = "Allow"
        Action = [
          "logs:PutLogEvents",
        ],
        Resource = [
          aws_cloudwatch_log_stream.target.arn,
        ]
      },
    ]
  })
}

resource "aws_cloudwatch_log_group" "target" {
  name = "%[1]s-target"
}

resource "aws_cloudwatch_log_stream" "target" {
  name           = "%[1]s-target"
  log_group_name = aws_cloudwatch_log_group.target.name
}

resource "aws_pipes_pipe" "test" {
  depends_on = [aws_iam_role_policy.source, aws_iam_role_policy.target]

  name     = %[1]q
  role_arn = aws_iam_role.test.arn
  source   = aws_dynamodb_table.source.stream_arn
  target   = aws_cloudwatch_log_group.target.arn

  source_parameters {
    dynamodb_stream_parameters {
      starting_position = "LATEST"
    }
  }

  target_parameters {
    cloudwatch_logs_parameters {
      log_stream_name = aws_cloudwatch_log_stream.target.name
    }
  }
}
`, rName))
}

func testAccPipeConfig_basicActiveMQSourceStepFunctionTarget(rName string) string {
	return acctest.ConfigCompose(
		testAccPipeConfig_base(rName),
		fmt.Sprintf(`
resource "aws_iam_role_policy" "source" {
  role = aws_iam_role.test.id
  name = "%[1]s-source"

  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Effect = "Allow"
        Action = [
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
        Resource = [
          "*"
        ]
      },
    ]
  })

  depends_on = [aws_mq_broker.source]
}

resource "aws_security_group" "source" {
  name = "%[1]s-source"

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

resource "aws_mq_broker" "source" {
  broker_name             = "%[1]s-source"
  engine_type             = "ActiveMQ"
  engine_version          = "5.17.6"
  host_instance_type      = "mq.t2.micro"
  security_groups         = [aws_security_group.source.id]
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

resource "aws_secretsmanager_secret" "source" {
  name = "%[1]s-source"
}

resource "aws_secretsmanager_secret_version" "source" {
  secret_id     = aws_secretsmanager_secret.source.id
  secret_string = jsonencode({ username = "Test", password = "TestTest1234" })
}

resource "aws_iam_role" "target" {
  name = "%[1]s-target"

  assume_role_policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Action": "sts:AssumeRole",
      "Principal": {
        "Service": "states.${data.aws_partition.main.dns_suffix}"
      },
      "Effect": "Allow",
      "Sid": ""
    }
  ]
}
EOF
}

resource "aws_sfn_state_machine" "target" {
  name     = "%[1]s-target"
  role_arn = aws_iam_role.target.arn

  definition = <<EOF
{
  "Comment": "A Hello World example of the Amazon States Language using an AWS Lambda Function",
  "StartAt": "HelloWorld",
  "States": {
    "HelloWorld": {
      "Type": "Pass",
      "End": true
    }
  }
}
EOF
}

resource "aws_pipes_pipe" "test" {
  depends_on = [aws_iam_role_policy.source]

  name     = %[1]q
  role_arn = aws_iam_role.test.arn
  source   = aws_mq_broker.source.arn
  target   = aws_sfn_state_machine.target.arn

  source_parameters {
    activemq_broker_parameters {
      queue_name = "test"

      credentials {
        basic_auth = aws_secretsmanager_secret_version.source.arn
      }
    }
  }

  target_parameters {
    step_function_state_machine_parameters {
      invocation_type = "REQUEST_RESPONSE"
    }
  }
}
`, rName))
}

func testAccPipeConfig_basicRabbitMQSourceEventBusTarget(rName string) string {
	return acctest.ConfigCompose(
		testAccPipeConfig_base(rName),
		fmt.Sprintf(`
resource "aws_iam_role_policy" "source" {
  role = aws_iam_role.test.id
  name = "%[1]s-source"

  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Effect = "Allow"
        Action = [
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
        Resource = [
          "*"
        ]
      },
    ]
  })

  depends_on = [aws_mq_broker.source]
}

resource "aws_mq_broker" "source" {
  broker_name             = "%[1]s-source"
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

resource "aws_secretsmanager_secret" "source" {
  name = "%[1]s-source"
}

resource "aws_secretsmanager_secret_version" "source" {
  secret_id     = aws_secretsmanager_secret.source.id
  secret_string = jsonencode({ username = "Test", password = "TestTest1234" })
}

resource "aws_iam_role_policy" "target" {
  role = aws_iam_role.test.id
  name = "%[1]s-target"

  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Effect = "Allow"
        Action = [
          "events:PutEvent",
        ],
        Resource = [
          aws_cloudwatch_event_bus.target.arn,
        ]
      },
    ]
  })
}

resource "aws_cloudwatch_event_bus" "target" {
  name = "%[1]s-target"
}

resource "aws_pipes_pipe" "test" {
  depends_on = [aws_iam_role_policy.source, aws_iam_role_policy.target]

  name     = %[1]q
  role_arn = aws_iam_role.test.arn
  source   = aws_mq_broker.source.arn
  target   = aws_cloudwatch_event_bus.target.arn

  source_parameters {
    rabbitmq_broker_parameters {
      queue_name = "test"

      credentials {
        basic_auth = aws_secretsmanager_secret_version.source.arn
      }
    }
  }
}
`, rName))
}

func testAccPipeConfig_basicMSKSourceHTTPTarget(rName string) string {
	return acctest.ConfigCompose(
		testAccPipeConfig_base(rName),
		acctest.ConfigVPCWithSubnets(rName, 3),
		fmt.Sprintf(`
resource "aws_iam_role_policy" "source" {
  role = aws_iam_role.test.id
  name = "%[1]s-source"

  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Effect = "Allow"
        Action = [
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
        Resource = [
          "*"
        ]
      },
    ]
  })

  depends_on = [aws_msk_cluster.source]
}

resource "aws_security_group" "source" {
  name   = "%[1]s-source"
  vpc_id = aws_vpc.test.id

  tags = {
    Name = "%[1]s-source"
  }
}

resource "aws_msk_cluster" "source" {
  cluster_name           = "%[1]s-source"
  kafka_version          = "2.7.1"
  number_of_broker_nodes = 3

  broker_node_group_info {
    client_subnets  = aws_subnet.test[*].id
    instance_type   = "kafka.m5.large"
    security_groups = [aws_security_group.source.id]

    storage_info {
      ebs_storage_info {
        volume_size = 10
      }
    }
  }
}

resource "aws_api_gateway_rest_api" "target" {
  name = "%[1]s-target"

  body = jsonencode({
    openapi = "3.0.1"
    info = {
      title   = "example"
      version = "1.0"
    }
    paths = {
      "/" = {
        get = {
          x-amazon-apigateway-integration = {
            httpMethod           = "GET"
            payloadFormatVersion = "1.0"
            type                 = "HTTP_PROXY"
            uri                  = "https://ip-ranges.amazonaws.com"
          }
        }
      }
    }
  })
}

resource "aws_api_gateway_deployment" "target" {
  rest_api_id = aws_api_gateway_rest_api.target.id

  triggers = {
    redeployment = sha1(jsonencode(aws_api_gateway_rest_api.target.body))
  }

  lifecycle {
    create_before_destroy = true
  }
}

resource "aws_api_gateway_stage" "target" {
  deployment_id = aws_api_gateway_deployment.target.id
  rest_api_id   = aws_api_gateway_rest_api.target.id
  stage_name    = "test"
}

resource "aws_pipes_pipe" "test" {
  depends_on = [aws_iam_role_policy.source]

  name     = %[1]q
  role_arn = aws_iam_role.test.arn
  source   = aws_msk_cluster.source.arn
  target   = "${aws_api_gateway_stage.target.execution_arn}/GET/*"

  source_parameters {
    managed_streaming_kafka_parameters {
      topic_name = "test"
    }
  }

  target_parameters {
    http_parameters {
      header_parameters = {
        "X-Test" = "test"
      }

      path_parameter_values = ["p1"]

      query_string_parameters = {
        "testing" = "yes"
      }
    }
  }
}
`, rName))
}

func testAccPipeConfig_basicSelfManagedKafkaSourceLambdaFunctionTarget(rName string) string {
	return acctest.ConfigCompose(
		testAccPipeConfig_base(rName),
		acctest.ConfigVPCWithSubnets(rName, 2),
		fmt.Sprintf(`
resource "aws_iam_role_policy" "source" {
  role = aws_iam_role.test.id
  name = "%[1]s-source"

  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Effect = "Allow"
        Action = [
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
        Resource = [
          "*"
        ]
      },
    ]
  })
}

resource "aws_security_group" "source" {
  name   = "%[1]s-source"
  vpc_id = aws_vpc.test.id

  tags = {
    Name = "%[1]s-source"
  }
}

resource "aws_iam_role" "target" {
  name = "%[1]s-target"

  assume_role_policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Action": "sts:AssumeRole",
      "Principal": {
        "Service": "lambda.${data.aws_partition.main.dns_suffix}"
      },
      "Effect": "Allow",
      "Sid": ""
    }
  ]
}
EOF
}

resource "aws_lambda_function" "target" {
  filename      = "test-fixtures/lambdatest.zip"
  function_name = "%[1]s-target"
  role          = aws_iam_role.target.arn
  handler       = "index.handler"
  runtime       = "nodejs16.x"
}

resource "aws_pipes_pipe" "test" {
  depends_on = [aws_iam_role_policy.source]

  name     = %[1]q
  role_arn = aws_iam_role.test.arn
  source   = "smk://test1:9092,test2:9092"
  target   = aws_lambda_function.target.arn

  source_parameters {
    self_managed_kafka_parameters {
      additional_bootstrap_servers = ["testing:1234"]
      consumer_group_id            = "self-managed-test-group-id"
      topic_name                   = "test"

      vpc {
        security_groups = [aws_security_group.source.id]
        subnets         = aws_subnet.test[*].id
      }
    }
  }

  target_parameters {
    lambda_function_parameters {
      invocation_type = "REQUEST_RESPONSE"
    }
  }
}
`, rName))
}

func testAccPipeConfig_basicSQSSourceRedshiftTarget(rName string) string {
	return acctest.ConfigCompose(
		testAccPipeConfig_base(rName),
		testAccPipeConfig_baseSQSSource(rName),
		acctest.ConfigAvailableAZsNoOptInExclude("usw2-az2"),
		fmt.Sprintf(`
resource "aws_redshift_cluster" "target" {
  cluster_identifier                  = "%[1]s-target"
  availability_zone                   = data.aws_availability_zones.available.names[0]
  database_name                       = "test"
  master_username                     = "tfacctest"
  master_password                     = "Mustbe8characters"
  node_type                           = "dc2.large"
  automated_snapshot_retention_period = 0
  allow_version_upgrade               = false
  skip_final_snapshot                 = true
}

resource "aws_pipes_pipe" "test" {
  depends_on = [aws_iam_role_policy.source]

  name     = %[1]q
  role_arn = aws_iam_role.test.arn
  source   = aws_sqs_queue.source.arn
  target   = aws_redshift_cluster.target.arn

  source_parameters {
    sqs_queue_parameters {
      batch_size                         = 1
      maximum_batching_window_in_seconds = 90
    }
  }

  target_parameters {
    redshift_data_parameters {
      database       = "db1"
      db_user        = "user1"
      sqls           = ["SELECT * FROM table"]
      statement_name = "SelectAll"
    }
  }
}
`, rName))
}

func testAccPipeConfig_basicSQSSourceSageMakerTarget(rName string) string {
	return acctest.ConfigCompose(
		testAccPipeConfig_base(rName),
		testAccPipeConfig_baseSQSSource(rName),
		fmt.Sprintf(`
# TODO Add aws_sagemaker_pipeline resource.

resource "aws_pipes_pipe" "test" {
  depends_on = [aws_iam_role_policy.source]

  name     = %[1]q
  role_arn = aws_iam_role.test.arn
  source   = aws_sqs_queue.source.arn
  target   = aws_sagemaker_pipeline.target.arn

  target_parameters {
    sagemaker_pipeline_parameters {
      pipeline_parameter {
        name  = "p1"
        value = "v1"
      }

      pipeline_parameter {
        name  = "p2"
        value = "v2"
      }
    }
  }
}
`, rName))
}

func testAccPipeConfig_basicSQSSourceBatchJobTarget(rName string) string {
	return acctest.ConfigCompose(
		testAccPipeConfig_base(rName),
		testAccPipeConfig_baseSQSSource(rName),
		acctest.ConfigVPCWithSubnets(rName, 2),
		fmt.Sprintf(`
resource "aws_iam_role" "target" {
  name               = "%[1]s-target"
  assume_role_policy = <<EOF
{
    "Version": "2012-10-17",
    "Statement": [
    {
        "Action": "sts:AssumeRole",
        "Effect": "Allow",
        "Principal": {
        "Service": "batch.${data.aws_partition.main.dns_suffix}"
        }
    }
    ]
}
EOF
}

resource "aws_iam_role_policy_attachment" "target" {
  role       = aws_iam_role.target.name
  policy_arn = "arn:${data.aws_partition.main.partition}:iam::aws:policy/service-role/AWSBatchServiceRole"
}

resource "aws_iam_role" "ecs_instance_role" {
  name = "%[1]s-ecs"

  assume_role_policy = <<EOF
{
    "Version": "2012-10-17",
    "Statement": [
    {
        "Action": "sts:AssumeRole",
        "Effect": "Allow",
        "Principal": {
        "Service": "ec2.${data.aws_partition.main.dns_suffix}"
        }
    }
    ]
}
EOF
}

resource "aws_iam_role_policy_attachment" "ecs_instance_role" {
  role       = aws_iam_role.ecs_instance_role.name
  policy_arn = "arn:${data.aws_partition.main.partition}:iam::aws:policy/service-role/AmazonEC2ContainerServiceforEC2Role"
}

resource "aws_iam_instance_profile" "ecs_instance_role" {
  name = aws_iam_role.ecs_instance_role.name
  role = aws_iam_role_policy_attachment.ecs_instance_role.role
}

resource "aws_security_group" "target" {
  name   = %[1]q
  vpc_id = aws_vpc.test.id

  tags = {
    Name = %[1]q
  }
}

resource "aws_batch_compute_environment" "target" {
  compute_environment_name = "%[1]s-target"
  service_role             = aws_iam_role.target.arn
  type                     = "MANAGED"

  compute_resources {
    instance_role      = aws_iam_instance_profile.ecs_instance_role.arn
    instance_type      = ["c5", "m5", "r5"]
    max_vcpus          = 1
    min_vcpus          = 0
    security_group_ids = [aws_security_group.target.id]
    subnets            = aws_subnet.test[*].id
    type               = "EC2"
  }

  depends_on = [aws_iam_role_policy_attachment.target]
}

resource "aws_batch_job_queue" "target" {
  compute_environments = [aws_batch_compute_environment.target.arn]
  name                 = "%[1]s-target"
  priority             = 1
  state                = "ENABLED"
}

resource "aws_batch_job_definition" "target" {
  container_properties = jsonencode({
    command = ["echo", "test"]
    image   = "busybox"
    memory  = 128
    vcpus   = 1
  })
  name = "%[1]s-target"
  type = "container"
}

resource "aws_pipes_pipe" "test" {
  depends_on = [aws_iam_role_policy.source]

  name     = %[1]q
  role_arn = aws_iam_role.test.arn
  source   = aws_sqs_queue.source.arn
  target   = aws_batch_job_queue.target.arn

  target_parameters {
    batch_job_parameters {
      array_properties {
        size = 512
      }

      container_overrides {
        command = ["rm", "-fr", "/"]

        environment {
          name  = "TMP"
          value = "/tmp2"
        }

        resource_requirement {
          type  = "GPU"
          value = "1"
        }
      }

      job_definition = aws_batch_job_definition.target.arn
      job_name       = "testing"

      parameters = {
        "Key1" = "Value1"
      }
    }
  }
}
`, rName))
}

func testAccPipeConfig_basicSQSSourceECSTaskTarget(rName string) string {
	return acctest.ConfigCompose(
		testAccPipeConfig_base(rName),
		testAccPipeConfig_baseSQSSource(rName),
		acctest.ConfigVPCWithSubnets(rName, 1),
		fmt.Sprintf(`
resource "aws_iam_role_policy" "target" {
  role = aws_iam_role.test.id
  name = "%[1]s-target"

  policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [{
    "Effect": "Allow",
    "Action": ["ecs:RunTask"],
    "Resource": ["*"]
  }]
}
EOF
}

resource "aws_ecs_cluster" "target" {
  name = "%[1]s-target"
}

resource "aws_ecs_cluster_capacity_providers" "target" {
  cluster_name       = aws_ecs_cluster.target.name
  capacity_providers = ["FARGATE"]
}

resource "aws_ecs_task_definition" "target" {
  family                   = "%[1]s-target"
  cpu                      = 256
  memory                   = 512
  requires_compatibilities = ["FARGATE"]
  network_mode             = "awsvpc"

  container_definitions = jsonencode([
    {
      name      = "sleep"
      image     = "busybox"
      cpu       = 10
      command   = ["sleep", "300"]
      memory    = 10
      essential = true
      portMappings = [
        {
          protocol      = "tcp"
          containerPort = 8000
        }
      ]
    }
  ])

  depends_on = [aws_ecs_cluster_capacity_providers.target]
}

resource "aws_pipes_pipe" "test" {
  depends_on = [aws_iam_role_policy.source, aws_iam_role_policy.target]

  name     = %[1]q
  role_arn = aws_iam_role.test.arn
  source   = aws_sqs_queue.source.arn
  target   = aws_ecs_cluster.target.id

  target_parameters {
    ecs_task_parameters {
      task_definition_arn = aws_ecs_task_definition.target.arn

      enable_ecs_managed_tags = true
      group                   = "g1"
      launch_type             = "FARGATE"

      network_configuration {
        aws_vpc_configuration {
          subnets = aws_subnet.test[*].id
        }
      }

      overrides {
        container_override {
          environment {
            name  = "TMP"
            value = "/tmp2"
          }

          memory_reservation = 1024
          name               = "first"

          resource_requirement {
            type  = "GPU"
            value = 2
          }
        }

        ephemeral_storage {
          size_in_gib = 32
        }
      }

      placement_strategy {
        field = "cpu"
        type  = "binpack"
      }

      propagate_tags = "TASK_DEFINITION"
      reference_id   = "refid"

      tags = {
        Name = %[1]q
      }
    }
  }
}
`, rName))
}
