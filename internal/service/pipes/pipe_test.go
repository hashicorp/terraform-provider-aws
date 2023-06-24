package pipes_test

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"testing"

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
		ErrorCheck:               acctest.ErrorCheck(t, names.PipesEndpointID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckPipeDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccPipeConfig_basicSQS(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckPipeExists(ctx, resourceName, &pipe),
					acctest.MatchResourceAttrRegionalARN(resourceName, "arn", "pipes", regexp.MustCompile(regexp.QuoteMeta(`pipe/`+rName))),
					resource.TestCheckResourceAttr(resourceName, "description", "Managed by Terraform"),
					resource.TestCheckResourceAttr(resourceName, "desired_state", "RUNNING"),
					resource.TestCheckResourceAttr(resourceName, "enrichment", ""),
					resource.TestCheckResourceAttr(resourceName, "enrichment_parameters.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttrPair(resourceName, "role_arn", "aws_iam_role.test", "arn"),
					resource.TestCheckResourceAttrPair(resourceName, "source", "aws_sqs_queue.source", "arn"),
					resource.TestCheckResourceAttr(resourceName, "source_parameters.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "source_parameters.0.activemq_broker_parameters.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "source_parameters.0.dynamodb_stream_parameters.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "source_parameters.0.filter_criteria.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "source_parameters.0.kinesis_stream_parameters.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "source_parameters.0.managed_streaming_kafka_parameters.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "source_parameters.0.rabbitmq_broker_parameters.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "source_parameters.0.self_managed_kafka_parameters.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "source_parameters.0.sqs_queue_parameters.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "source_parameters.0.sqs_queue_parameters.0.batch_size", "10"),
					resource.TestCheckResourceAttr(resourceName, "source_parameters.0.sqs_queue_parameters.0.maximum_batching_window_in_seconds", "0"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
					resource.TestCheckResourceAttrPair(resourceName, "target", "aws_sqs_queue.target", "arn"),
					resource.TestCheckResourceAttr(resourceName, "target_parameters.#", "0"),
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
		ErrorCheck:               acctest.ErrorCheck(t, names.PipesEndpointID),
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
		ErrorCheck:               acctest.ErrorCheck(t, names.PipesEndpointID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckPipeDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccPipeConfig_description(rName, "Description 1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPipeExists(ctx, resourceName, &pipe),
					resource.TestCheckResourceAttr(resourceName, "description", "Description 1"),
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
					resource.TestCheckResourceAttr(resourceName, "description", "Description 2"),
				),
			},
			{
				Config: testAccPipeConfig_description(rName, ""),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPipeExists(ctx, resourceName, &pipe),
					resource.TestCheckResourceAttr(resourceName, "description", ""),
				),
			},
			{
				Config: testAccPipeConfig_basicSQS(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPipeExists(ctx, resourceName, &pipe),
					resource.TestCheckResourceAttr(resourceName, "description", "Managed by Terraform"),
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
		ErrorCheck:               acctest.ErrorCheck(t, names.PipesEndpointID),
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
		ErrorCheck:               acctest.ErrorCheck(t, names.PipesEndpointID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckPipeDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccPipeConfig_enrichment(rName, 0),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPipeExists(ctx, resourceName, &pipe),
					resource.TestCheckResourceAttrPair(resourceName, "enrichment", "aws_cloudwatch_event_api_destination.test.0", "arn"),
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
					resource.TestCheckResourceAttrPair(resourceName, "enrichment", "aws_cloudwatch_event_api_destination.test.1", "arn"),
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
		ErrorCheck:               acctest.ErrorCheck(t, names.PipesEndpointID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckPipeDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccPipeConfig_sourceParameters_filterCriteria1(rName, "test1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPipeExists(ctx, resourceName, &pipe),
					resource.TestCheckResourceAttr(resourceName, "source_parameters.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "source_parameters.0.filter_criteria.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "source_parameters.0.filter_criteria.0.filter.#", "1"),
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
					resource.TestCheckResourceAttr(resourceName, "source_parameters.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "source_parameters.0.filter_criteria.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "source_parameters.0.filter_criteria.0.filter.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "source_parameters.0.filter_criteria.0.filter.0.pattern", `{"source":["test1"]}`),
					resource.TestCheckResourceAttr(resourceName, "source_parameters.0.filter_criteria.0.filter.1.pattern", `{"source":["test2"]}`),
				),
			},
			{
				Config: testAccPipeConfig_sourceParameters_filterCriteria1(rName, "test2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPipeExists(ctx, resourceName, &pipe),
					resource.TestCheckResourceAttr(resourceName, "source_parameters.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "source_parameters.0.filter_criteria.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "source_parameters.0.filter_criteria.0.filter.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "source_parameters.0.filter_criteria.0.filter.0.pattern", `{"source":["test2"]}`),
				),
			},
			{
				Config: testAccPipeConfig_sourceParameters_filterCriteria0(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPipeExists(ctx, resourceName, &pipe),
					resource.TestCheckResourceAttr(resourceName, "source_parameters.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "source_parameters.0.filter_criteria.#", "0"),
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
					resource.TestCheckResourceAttr(resourceName, "source_parameters.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "source_parameters.0.filter_criteria.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "source_parameters.0.filter_criteria.0.filter.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "source_parameters.0.filter_criteria.0.filter.0.pattern", `{"source":["test2"]}`),
				),
			},
			{
				Config: testAccPipeConfig_basicSQS(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPipeExists(ctx, resourceName, &pipe),
					resource.TestCheckResourceAttr(resourceName, "source_parameters.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "source_parameters.0.filter_criteria.#", "1"),
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
		ErrorCheck:               acctest.ErrorCheck(t, names.PipesEndpointID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckPipeDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccPipeConfig_nameGenerated(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPipeExists(ctx, resourceName, &pipe),
					acctest.CheckResourceAttrNameGenerated(resourceName, "name"),
					resource.TestCheckResourceAttr(resourceName, "name_prefix", id.UniqueIdPrefix),
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
		ErrorCheck:               acctest.ErrorCheck(t, names.PipesEndpointID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckPipeDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccPipeConfig_namePrefix(rName, "tf-acc-test-prefix-"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPipeExists(ctx, resourceName, &pipe),
					acctest.CheckResourceAttrNameFromPrefix(resourceName, "name", "tf-acc-test-prefix-"),
					resource.TestCheckResourceAttr(resourceName, "name_prefix", "tf-acc-test-prefix-"),
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
		ErrorCheck:               acctest.ErrorCheck(t, names.PipesEndpointID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckPipeDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccPipeConfig_basicSQS(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPipeExists(ctx, resourceName, &pipe),
					resource.TestCheckResourceAttrPair(resourceName, "role_arn", "aws_iam_role.test", "arn"),
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
					resource.TestCheckResourceAttrPair(resourceName, "role_arn", "aws_iam_role.test2", "arn"),
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
		ErrorCheck:               acctest.ErrorCheck(t, names.PipesEndpointID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckPipeDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccPipeConfig_tags1(rName, "key1", "value1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPipeExists(ctx, resourceName, &pipe),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccPipeConfig_tags2(rName, "key1", "value1updated", "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPipeExists(ctx, resourceName, &pipe),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1updated"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
			{
				Config: testAccPipeConfig_tags1(rName, "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPipeExists(ctx, resourceName, &pipe),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
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
		ErrorCheck:               acctest.ErrorCheck(t, names.PipesEndpointID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckPipeDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccPipeConfig_basicSQS(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPipeExists(ctx, resourceName, &pipe),
					resource.TestCheckResourceAttrPair(resourceName, "target", "aws_sqs_queue.target", "arn"),
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
					resource.TestCheckResourceAttrPair(resourceName, "target", "aws_sqs_queue.target2", "arn"),
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
		ErrorCheck:               acctest.ErrorCheck(t, names.PipesEndpointID),
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
		ErrorCheck:               acctest.ErrorCheck(t, names.PipesEndpointID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckPipeDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccPipeConfig_basicKinesis(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckPipeExists(ctx, resourceName, &pipe),
					acctest.MatchResourceAttrRegionalARN(resourceName, "arn", "pipes", regexp.MustCompile(regexp.QuoteMeta(`pipe/`+rName))),
					resource.TestCheckResourceAttr(resourceName, "description", "Managed by Terraform"),
					resource.TestCheckResourceAttr(resourceName, "desired_state", "RUNNING"),
					resource.TestCheckResourceAttr(resourceName, "enrichment", ""),
					resource.TestCheckResourceAttr(resourceName, "enrichment_parameters.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttrPair(resourceName, "role_arn", "aws_iam_role.test", "arn"),
					resource.TestCheckResourceAttrPair(resourceName, "source", "aws_kinesis_stream.source", "arn"),
					resource.TestCheckResourceAttr(resourceName, "source_parameters.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "source_parameters.0.activemq_broker_parameters.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "source_parameters.0.dynamodb_stream_parameters.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "source_parameters.0.filter_criteria.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "source_parameters.0.kinesis_stream_parameters.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "source_parameters.0.kinesis_stream_parameters.0.batch_size", "100"),
					resource.TestCheckResourceAttr(resourceName, "source_parameters.0.kinesis_stream_parameters.0.dead_letter_config.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "source_parameters.0.kinesis_stream_parameters.0.maximum_batching_window_in_seconds", "0"),
					resource.TestCheckResourceAttr(resourceName, "source_parameters.0.kinesis_stream_parameters.0.maximum_record_age_in_seconds", "0"),
					resource.TestCheckResourceAttr(resourceName, "source_parameters.0.kinesis_stream_parameters.0.maximum_retry_attempts", "0"),
					resource.TestCheckResourceAttr(resourceName, "source_parameters.0.kinesis_stream_parameters.0.on_partial_batch_item_failure", ""),
					resource.TestCheckResourceAttr(resourceName, "source_parameters.0.kinesis_stream_parameters.0.parallelization_factor", "0"),
					resource.TestCheckResourceAttr(resourceName, "source_parameters.0.kinesis_stream_parameters.0.starting_position", "LATEST"),
					resource.TestCheckResourceAttr(resourceName, "source_parameters.0.kinesis_stream_parameters.0.starting_position_timestamp", ""),
					resource.TestCheckResourceAttr(resourceName, "source_parameters.0.managed_streaming_kafka_parameters.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "source_parameters.0.rabbitmq_broker_parameters.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "source_parameters.0.self_managed_kafka_parameters.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "source_parameters.0.sqs_queue_parameters.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
					resource.TestCheckResourceAttrPair(resourceName, "target", "aws_kinesis_stream.target", "arn"),
					resource.TestCheckResourceAttr(resourceName, "target_parameters.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "target_parameters.0.batch_job_parameters.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "target_parameters.0.cloudwatch_logs_parameters.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "target_parameters.0.ecs_task_parameters.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "target_parameters.0.eventbridge_event_bus_parameters.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "target_parameters.0.http_parameters.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "target_parameters.0.input_template", ""),
					resource.TestCheckResourceAttr(resourceName, "target_parameters.0.kinesis_stream_parameters.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "target_parameters.0.kinesis_stream_parameters.0.partition_key", "test"),
					resource.TestCheckResourceAttr(resourceName, "target_parameters.0.lambda_function_parameters.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "target_parameters.0.redshift_data_parameters.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "target_parameters.0.sagemaker_pipeline_parameters.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "target_parameters.0.sqs_queue_parameters.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "target_parameters.0.step_function_state_machine_parameters.#", "0"),
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
		ErrorCheck:               acctest.ErrorCheck(t, names.PipesEndpointID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckPipeDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccPipeConfig_basicDynamoDBSourceCloudWatchLogsTarget(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckPipeExists(ctx, resourceName, &pipe),
					acctest.MatchResourceAttrRegionalARN(resourceName, "arn", "pipes", regexp.MustCompile(regexp.QuoteMeta(`pipe/`+rName))),
					resource.TestCheckResourceAttr(resourceName, "description", "Managed by Terraform"),
					resource.TestCheckResourceAttr(resourceName, "desired_state", "RUNNING"),
					resource.TestCheckResourceAttr(resourceName, "enrichment", ""),
					resource.TestCheckResourceAttr(resourceName, "enrichment_parameters.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttrPair(resourceName, "role_arn", "aws_iam_role.test", "arn"),
					resource.TestCheckResourceAttrPair(resourceName, "source", "aws_dynamodb_table.source", "stream_arn"),
					resource.TestCheckResourceAttr(resourceName, "source_parameters.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "source_parameters.0.activemq_broker_parameters.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "source_parameters.0.dynamodb_stream_parameters.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "source_parameters.0.dynamodb_stream_parameters.0.batch_size", "100"),
					resource.TestCheckResourceAttr(resourceName, "source_parameters.0.dynamodb_stream_parameters.0.dead_letter_config.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "source_parameters.0.dynamodb_stream_parameters.0.maximum_batching_window_in_seconds", "0"),
					resource.TestCheckResourceAttr(resourceName, "source_parameters.0.dynamodb_stream_parameters.0.maximum_record_age_in_seconds", "0"),
					resource.TestCheckResourceAttr(resourceName, "source_parameters.0.dynamodb_stream_parameters.0.maximum_retry_attempts", "0"),
					resource.TestCheckResourceAttr(resourceName, "source_parameters.0.dynamodb_stream_parameters.0.on_partial_batch_item_failure", ""),
					resource.TestCheckResourceAttr(resourceName, "source_parameters.0.dynamodb_stream_parameters.0.parallelization_factor", "0"),
					resource.TestCheckResourceAttr(resourceName, "source_parameters.0.dynamodb_stream_parameters.0.starting_position", "LATEST"),
					resource.TestCheckResourceAttr(resourceName, "source_parameters.0.filter_criteria.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "source_parameters.0.managed_streaming_kafka_parameters.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "source_parameters.0.rabbitmq_broker_parameters.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "source_parameters.0.self_managed_kafka_parameters.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "source_parameters.0.sqs_queue_parameters.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
					resource.TestCheckResourceAttrPair(resourceName, "target", "aws_cloudwatch_log_group.target", "arn"),
					resource.TestCheckResourceAttr(resourceName, "target_parameters.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "target_parameters.0.batch_job_parameters.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "target_parameters.0.cloudwatch_logs_parameters.#", "1"),
					resource.TestCheckResourceAttrPair(resourceName, "target_parameters.0.cloudwatch_logs_parameters.0.log_stream_name", "aws_cloudwatch_log_stream.target", "name"),
					resource.TestCheckResourceAttr(resourceName, "target_parameters.0.cloudwatch_logs_parameters.0.timestamp", ""),
					resource.TestCheckResourceAttr(resourceName, "target_parameters.0.ecs_task_parameters.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "target_parameters.0.eventbridge_event_bus_parameters.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "target_parameters.0.http_parameters.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "target_parameters.0.input_template", ""),
					resource.TestCheckResourceAttr(resourceName, "target_parameters.0.kinesis_stream_parameters.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "target_parameters.0.lambda_function_parameters.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "target_parameters.0.redshift_data_parameters.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "target_parameters.0.sagemaker_pipeline_parameters.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "target_parameters.0.sqs_queue_parameters.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "target_parameters.0.step_function_state_machine_parameters.#", "0"),
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
		ErrorCheck:               acctest.ErrorCheck(t, names.PipesEndpointID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckPipeDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccPipeConfig_basicActiveMQSourceStepFunctionTarget(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckPipeExists(ctx, resourceName, &pipe),
					acctest.MatchResourceAttrRegionalARN(resourceName, "arn", "pipes", regexp.MustCompile(regexp.QuoteMeta(`pipe/`+rName))),
					resource.TestCheckResourceAttr(resourceName, "description", "Managed by Terraform"),
					resource.TestCheckResourceAttr(resourceName, "desired_state", "RUNNING"),
					resource.TestCheckResourceAttr(resourceName, "enrichment", ""),
					resource.TestCheckResourceAttr(resourceName, "enrichment_parameters.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttrPair(resourceName, "role_arn", "aws_iam_role.test", "arn"),
					resource.TestCheckResourceAttrPair(resourceName, "source", "aws_mq_broker.source", "arn"),
					resource.TestCheckResourceAttr(resourceName, "source_parameters.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "source_parameters.0.activemq_broker_parameters.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "source_parameters.0.activemq_broker_parameters.0.batch_size", "100"),
					resource.TestCheckResourceAttr(resourceName, "source_parameters.0.activemq_broker_parameters.0.credentials.#", "1"),
					resource.TestCheckResourceAttrSet(resourceName, "source_parameters.0.activemq_broker_parameters.0.credentials.0.basic_auth"),
					resource.TestCheckResourceAttr(resourceName, "source_parameters.0.activemq_broker_parameters.0.maximum_batching_window_in_seconds", "0"),
					resource.TestCheckResourceAttr(resourceName, "source_parameters.0.activemq_broker_parameters.0.queue_name", "test"),
					resource.TestCheckResourceAttr(resourceName, "source_parameters.0.dynamodb_stream_parameters.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "source_parameters.0.filter_criteria.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "source_parameters.0.managed_streaming_kafka_parameters.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "source_parameters.0.rabbitmq_broker_parameters.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "source_parameters.0.self_managed_kafka_parameters.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "source_parameters.0.sqs_queue_parameters.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
					resource.TestCheckResourceAttrPair(resourceName, "target", "aws_sfn_state_machine.target", "arn"),
					resource.TestCheckResourceAttr(resourceName, "target_parameters.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "target_parameters.0.batch_job_parameters.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "target_parameters.0.cloudwatch_logs_parameters.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "target_parameters.0.ecs_task_parameters.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "target_parameters.0.eventbridge_event_bus_parameters.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "target_parameters.0.http_parameters.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "target_parameters.0.input_template", ""),
					resource.TestCheckResourceAttr(resourceName, "target_parameters.0.kinesis_stream_parameters.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "target_parameters.0.lambda_function_parameters.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "target_parameters.0.redshift_data_parameters.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "target_parameters.0.sagemaker_pipeline_parameters.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "target_parameters.0.sqs_queue_parameters.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "target_parameters.0.step_function_state_machine_parameters.#", "1"),
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

func testAccPipeConfig_baseSQSDeadLetter(rName string) string {
	return fmt.Sprintf(`
resource "aws_iam_role_policy" "deadletter" {
  role = aws_iam_role.test.id
  name = "%[1]s-deadletter"

  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Effect = "Allow"
        Action = [
          "sqs:*",
        ],
        Resource = [
          aws_sqs_queue.deadletter.arn,
        ]
      },
    ]
  })
}

resource "aws_sqs_queue" "deadletter" {
  name = "%[1]s-deadletter"
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
  depends_on  = [aws_iam_role_policy.source, aws_iam_role_policy.target]

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
  engine_version          = "5.15.0"
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
