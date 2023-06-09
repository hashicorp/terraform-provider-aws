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

func TestAccPipesPipe_basic(t *testing.T) {
	ctx := acctest.Context(t)

	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

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
				Config: testAccPipeConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPipeExists(ctx, resourceName, &pipe),
					acctest.MatchResourceAttrRegionalARN(resourceName, "arn", "pipes", regexp.MustCompile(regexp.QuoteMeta(`pipe/`+rName))),
					resource.TestCheckResourceAttr(resourceName, "description", "Managed by Terraform"),
					resource.TestCheckResourceAttr(resourceName, "desired_state", "RUNNING"),
					resource.TestCheckResourceAttr(resourceName, "enrichment", ""),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttrPair(resourceName, "role_arn", "aws_iam_role.test", "arn"),
					resource.TestCheckResourceAttrPair(resourceName, "source", "aws_sqs_queue.source", "arn"),
					resource.TestCheckResourceAttr(resourceName, "source_parameters.#", "1"),
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
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

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
				Config: testAccPipeConfig_basic(rName),
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
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var pipe pipes.DescribePipeOutput
	name := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
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
				Config: testAccPipeConfig_description(name, "Description 1"),
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
				Config: testAccPipeConfig_description(name, "Description 2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPipeExists(ctx, resourceName, &pipe),
					resource.TestCheckResourceAttr(resourceName, "description", "Description 2"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccPipeConfig_description(name, ""),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPipeExists(ctx, resourceName, &pipe),
					resource.TestCheckResourceAttr(resourceName, "description", ""),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccPipeConfig_basic(name),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPipeExists(ctx, resourceName, &pipe),
					resource.TestCheckResourceAttr(resourceName, "description", "Managed by Terraform"),
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

func TestAccPipesPipe_desiredState(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var pipe pipes.DescribePipeOutput
	name := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
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
				Config: testAccPipeConfig_desiredState(name, "STOPPED"),
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
				Config: testAccPipeConfig_desiredState(name, "RUNNING"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPipeExists(ctx, resourceName, &pipe),
					resource.TestCheckResourceAttr(resourceName, "desired_state", "RUNNING"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccPipeConfig_desiredState(name, "STOPPED"),
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
				Config: testAccPipeConfig_basic(name),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPipeExists(ctx, resourceName, &pipe),
					resource.TestCheckResourceAttr(resourceName, "desired_state", "RUNNING"),
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

func TestAccPipesPipe_enrichment(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var pipe pipes.DescribePipeOutput
	name := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_pipes_pipe.test"

	headerKey := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	headerValue := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	queryStringKey := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	queryStringValue := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	headerKeyModified := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	headerValueModified := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	queryStringKeyModified := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	queryStringValueModified := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

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
				Config: testAccPipeConfig_enrichment(
					name,
					0,
					headerKey,
					headerValue,
					queryStringKey,
					queryStringValue,
				),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPipeExists(ctx, resourceName, &pipe),
					resource.TestCheckResourceAttrPair(resourceName, "enrichment", "aws_cloudwatch_event_api_destination.test.0", "arn"),
					resource.TestCheckResourceAttr(resourceName, "enrichment_parameters.0.http_parameters.0.header.0.key", headerKey),
					resource.TestCheckResourceAttr(resourceName, "enrichment_parameters.0.http_parameters.0.header.0.value", headerValue),
					resource.TestCheckResourceAttr(resourceName, "enrichment_parameters.0.http_parameters.0.path_parameters.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "enrichment_parameters.0.http_parameters.0.query_string.0.key", queryStringKey),
					resource.TestCheckResourceAttr(resourceName, "enrichment_parameters.0.http_parameters.0.query_string.0.value", queryStringValue),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccPipeConfig_enrichment(
					name,
					1,
					headerKeyModified,
					headerValueModified,
					queryStringKeyModified,
					queryStringValueModified,
				),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPipeExists(ctx, resourceName, &pipe),
					resource.TestCheckResourceAttrPair(resourceName, "enrichment", "aws_cloudwatch_event_api_destination.test.1", "arn"),
					resource.TestCheckResourceAttr(resourceName, "enrichment_parameters.0.http_parameters.0.header.0.key", headerKeyModified),
					resource.TestCheckResourceAttr(resourceName, "enrichment_parameters.0.http_parameters.0.header.0.value", headerValueModified),
					resource.TestCheckResourceAttr(resourceName, "enrichment_parameters.0.http_parameters.0.path_parameters.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "enrichment_parameters.0.http_parameters.0.query_string.0.key", queryStringKeyModified),
					resource.TestCheckResourceAttr(resourceName, "enrichment_parameters.0.http_parameters.0.query_string.0.value", queryStringValueModified),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccPipeConfig_basic(name),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPipeExists(ctx, resourceName, &pipe),
					resource.TestCheckResourceAttr(resourceName, "enrichment", ""),
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

func TestAccPipesPipe_sourceParameters_filterCriteria(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var pipe pipes.DescribePipeOutput
	name := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
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
				Config: testAccPipeConfig_sourceParameters_filterCriteria1(name, "test1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPipeExists(ctx, resourceName, &pipe),
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
				Config: testAccPipeConfig_sourceParameters_filterCriteria2(name, "test1", "test2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPipeExists(ctx, resourceName, &pipe),
					resource.TestCheckResourceAttr(resourceName, "source_parameters.0.filter_criteria.0.filter.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "source_parameters.0.filter_criteria.0.filter.0.pattern", `{"source":["test1"]}`),
					resource.TestCheckResourceAttr(resourceName, "source_parameters.0.filter_criteria.0.filter.1.pattern", `{"source":["test2"]}`),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccPipeConfig_sourceParameters_filterCriteria1(name, "test2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPipeExists(ctx, resourceName, &pipe),
					resource.TestCheckResourceAttr(resourceName, "source_parameters.0.filter_criteria.0.filter.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "source_parameters.0.filter_criteria.0.filter.0.pattern", `{"source":["test2"]}`),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccPipeConfig_sourceParameters_filterCriteria1(name, "test2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPipeExists(ctx, resourceName, &pipe),
					resource.TestCheckResourceAttr(resourceName, "source_parameters.0.filter_criteria.0.filter.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "source_parameters.0.filter_criteria.0.filter.0.pattern", `{"source":["test2"]}`),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccPipeConfig_basic(name),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPipeExists(ctx, resourceName, &pipe),
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

func TestAccPipesPipe_nameGenerated(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var pipe pipes.DescribePipeOutput
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
				Config: testAccPipeConfig_nameGenerated(),
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
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var pipe pipes.DescribePipeOutput
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
				Config: testAccPipeConfig_namePrefix("tf-acc-test-prefix-"),
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
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var pipe pipes.DescribePipeOutput
	name := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
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
				Config: testAccPipeConfig_basic(name),
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
				Config: testAccPipeConfig_roleARN(name),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPipeExists(ctx, resourceName, &pipe),
					resource.TestCheckResourceAttrPair(resourceName, "role_arn", "aws_iam_role.test2", "arn"),
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

func TestAccPipesPipe_tags(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var pipe pipes.DescribePipeOutput
	name := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
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
				Config: testAccPipeConfig_tags1(name, "key1", "value1"),
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
				Config: testAccPipeConfig_tags2(name, "key1", "value1updated", "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPipeExists(ctx, resourceName, &pipe),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1updated"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccPipeConfig_tags1(name, "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPipeExists(ctx, resourceName, &pipe),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
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

func TestAccPipesPipe_source_sqs_target_sqs(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var pipe pipes.DescribePipeOutput
	name := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_pipes_pipe.test"

	batchSize := 8
	batchWindow := 5
	dedupeID := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	messageGroupID := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	batchSizeModified := 9
	batchWindowModified := 6
	dedupeIDModified := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	messageGroupIDModified := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

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
				Config: testAccPipeConfig_source_sqs_target_sqs(name, batchSize, batchWindow, dedupeID, messageGroupID),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPipeExists(ctx, resourceName, &pipe),
					resource.TestCheckResourceAttrPair(resourceName, "source", "aws_sqs_queue.source", "arn"),
					resource.TestCheckResourceAttrPair(resourceName, "target", "aws_sqs_queue.target_fifo", "arn"),
					resource.TestCheckResourceAttr(resourceName, "source_parameters.0.sqs_queue.0.batch_size", fmt.Sprintf("%d", batchSize)),
					resource.TestCheckResourceAttr(resourceName, "source_parameters.0.sqs_queue.0.maximum_batching_window_in_seconds", fmt.Sprintf("%d", batchWindow)),
					resource.TestCheckResourceAttr(resourceName, "target_parameters.0.sqs_queue.0.message_deduplication_id", dedupeID),
					resource.TestCheckResourceAttr(resourceName, "target_parameters.0.sqs_queue.0.message_group_id", messageGroupID),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccPipeConfig_source_sqs_target_sqs(name, batchSizeModified, batchWindowModified, dedupeIDModified, messageGroupIDModified),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPipeExists(ctx, resourceName, &pipe),
					resource.TestCheckResourceAttrPair(resourceName, "source", "aws_sqs_queue.source", "arn"),
					resource.TestCheckResourceAttrPair(resourceName, "target", "aws_sqs_queue.target_fifo", "arn"),
					resource.TestCheckResourceAttr(resourceName, "source_parameters.0.sqs_queue.0.batch_size", fmt.Sprintf("%d", batchSizeModified)),
					resource.TestCheckResourceAttr(resourceName, "source_parameters.0.sqs_queue.0.maximum_batching_window_in_seconds", fmt.Sprintf("%d", batchWindowModified)),
					resource.TestCheckResourceAttr(resourceName, "target_parameters.0.sqs_queue.0.message_deduplication_id", dedupeIDModified),
					resource.TestCheckResourceAttr(resourceName, "target_parameters.0.sqs_queue.0.message_group_id", messageGroupIDModified),
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

func TestAccPipesPipe_source_kinesis_target_kinesis(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var pipe pipes.DescribePipeOutput
	name := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_pipes_pipe.test"

	batchSize := 10
	batchWindow := 5
	maxRecordAge := -1
	parallelization := 2
	retries := 3
	partitionKey := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	batchSizeModified := 11
	batchWindowModified := 6
	maxRecordAgeModified := 65
	parallelizationModified := 3
	retriesModified := 4
	partitionKeyModified := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

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
				Config: testAccPipeConfig_source_kinesis_target_kinesis(name, batchSize, batchWindow, maxRecordAge, parallelization, retries, partitionKey),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPipeExists(ctx, resourceName, &pipe),
					resource.TestCheckResourceAttrPair(resourceName, "source", "aws_kinesis_stream.source", "arn"),
					resource.TestCheckResourceAttrPair(resourceName, "target", "aws_kinesis_stream.target", "arn"),
					resource.TestCheckResourceAttrPair(resourceName, "source_parameters.0.kinesis_stream.0.dead_letter_config.0.arn", "aws_sqs_queue.deadletter", "arn"),
					resource.TestCheckResourceAttr(resourceName, "source_parameters.0.kinesis_stream.0.batch_size", fmt.Sprintf("%d", batchSize)),
					resource.TestCheckResourceAttr(resourceName, "source_parameters.0.kinesis_stream.0.maximum_batching_window_in_seconds", fmt.Sprintf("%d", batchWindow)),
					resource.TestCheckResourceAttr(resourceName, "source_parameters.0.kinesis_stream.0.maximum_record_age_in_seconds", fmt.Sprintf("%d", maxRecordAge)),
					resource.TestCheckResourceAttr(resourceName, "source_parameters.0.kinesis_stream.0.parallelization_factor", fmt.Sprintf("%d", parallelization)),
					resource.TestCheckResourceAttr(resourceName, "source_parameters.0.kinesis_stream.0.maximum_retry_attempts", fmt.Sprintf("%d", retries)),
					resource.TestCheckResourceAttr(resourceName, "source_parameters.0.kinesis_stream.0.starting_position", "AT_TIMESTAMP"),
					resource.TestCheckResourceAttr(resourceName, "source_parameters.0.kinesis_stream.0.on_partial_batch_item_failure", "AUTOMATIC_BISECT"),
					resource.TestCheckResourceAttr(resourceName, "source_parameters.0.kinesis_stream.0.starting_position_timestamp", "2023-01-01T00:00:00Z"),
					resource.TestCheckResourceAttr(resourceName, "target_parameters.0.kinesis_stream.0.partition_key", partitionKey),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccPipeConfig_source_kinesis_target_kinesis(name, batchSizeModified, batchWindowModified, maxRecordAgeModified, parallelizationModified, retriesModified, partitionKeyModified),
				Check: resource.ComposeTestCheckFunc(

					testAccCheckPipeExists(ctx, resourceName, &pipe),
					resource.TestCheckResourceAttrPair(resourceName, "source", "aws_kinesis_stream.source", "arn"),
					resource.TestCheckResourceAttrPair(resourceName, "target", "aws_kinesis_stream.target", "arn"),
					resource.TestCheckResourceAttrPair(resourceName, "source_parameters.0.kinesis_stream.0.dead_letter_config.0.arn", "aws_sqs_queue.deadletter", "arn"),
					resource.TestCheckResourceAttr(resourceName, "source_parameters.0.kinesis_stream.0.batch_size", fmt.Sprintf("%d", batchSizeModified)),
					resource.TestCheckResourceAttr(resourceName, "source_parameters.0.kinesis_stream.0.maximum_batching_window_in_seconds", fmt.Sprintf("%d", batchWindowModified)),
					resource.TestCheckResourceAttr(resourceName, "source_parameters.0.kinesis_stream.0.maximum_record_age_in_seconds", fmt.Sprintf("%d", maxRecordAgeModified)),
					resource.TestCheckResourceAttr(resourceName, "source_parameters.0.kinesis_stream.0.parallelization_factor", fmt.Sprintf("%d", parallelizationModified)),
					resource.TestCheckResourceAttr(resourceName, "source_parameters.0.kinesis_stream.0.maximum_retry_attempts", fmt.Sprintf("%d", retriesModified)),
					resource.TestCheckResourceAttr(resourceName, "source_parameters.0.kinesis_stream.0.starting_position", "AT_TIMESTAMP"),
					resource.TestCheckResourceAttr(resourceName, "source_parameters.0.kinesis_stream.0.on_partial_batch_item_failure", "AUTOMATIC_BISECT"),
					resource.TestCheckResourceAttr(resourceName, "source_parameters.0.kinesis_stream.0.starting_position_timestamp", "2023-01-01T00:00:00Z"),
					resource.TestCheckResourceAttr(resourceName, "target_parameters.0.kinesis_stream.0.partition_key", partitionKeyModified),
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

func TestAccPipesPipe_source_dynamo_target_cloudwatch_logs(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var pipe pipes.DescribePipeOutput
	name := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_pipes_pipe.test"

	batchSize := 8
	batchWindow := 5
	maxRecordAge := -1
	parallelization := 2
	retries := 3

	batchSizeModified := 9
	batchWindowModified := 6
	maxRecordAgeModified := 65
	parallelizationModified := 3
	retriesModified := 4

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
				Config: testAccPipeConfig_source_dynamo_target_cloudwatch_logs(name, batchSize, batchWindow, maxRecordAge, parallelization, retries),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPipeExists(ctx, resourceName, &pipe),
					resource.TestCheckResourceAttrPair(resourceName, "source", "aws_dynamodb_table.source", "stream_arn"),
					resource.TestCheckResourceAttrPair(resourceName, "target", "aws_cloudwatch_log_group.target", "arn"),
					resource.TestCheckResourceAttrPair(resourceName, "source_parameters.0.dynamo_db_stream.0.dead_letter_config.0.arn", "aws_sqs_queue.deadletter", "arn"),
					resource.TestCheckResourceAttr(resourceName, "source_parameters.0.dynamo_db_stream.0.batch_size", fmt.Sprintf("%d", batchSize)),
					resource.TestCheckResourceAttr(resourceName, "source_parameters.0.dynamo_db_stream.0.maximum_batching_window_in_seconds", fmt.Sprintf("%d", batchWindow)),
					resource.TestCheckResourceAttr(resourceName, "source_parameters.0.dynamo_db_stream.0.maximum_record_age_in_seconds", fmt.Sprintf("%d", maxRecordAge)),
					resource.TestCheckResourceAttr(resourceName, "source_parameters.0.dynamo_db_stream.0.parallelization_factor", fmt.Sprintf("%d", parallelization)),
					resource.TestCheckResourceAttr(resourceName, "source_parameters.0.dynamo_db_stream.0.maximum_retry_attempts", fmt.Sprintf("%d", retries)),
					resource.TestCheckResourceAttr(resourceName, "source_parameters.0.dynamo_db_stream.0.starting_position", "TRIM_HORIZON"),
					resource.TestCheckResourceAttr(resourceName, "source_parameters.0.dynamo_db_stream.0.on_partial_batch_item_failure", "AUTOMATIC_BISECT"),
					resource.TestCheckResourceAttr(resourceName, "target_parameters.0.cloudwatch_logs.0.log_stream_name", name),
					resource.TestCheckResourceAttr(resourceName, "target_parameters.0.cloudwatch_logs.0.timestamp", "$.detail.timestamp"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccPipeConfig_source_dynamo_target_cloudwatch_logs(name, batchSizeModified, batchWindowModified, maxRecordAgeModified, parallelizationModified, retriesModified),
				Check: resource.ComposeTestCheckFunc(

					testAccCheckPipeExists(ctx, resourceName, &pipe),
					resource.TestCheckResourceAttrPair(resourceName, "source", "aws_dynamodb_table.source", "stream_arn"),
					resource.TestCheckResourceAttrPair(resourceName, "target", "aws_cloudwatch_log_group.target", "arn"),
					resource.TestCheckResourceAttrPair(resourceName, "source_parameters.0.dynamo_db_stream.0.dead_letter_config.0.arn", "aws_sqs_queue.deadletter", "arn"),
					resource.TestCheckResourceAttr(resourceName, "source_parameters.0.dynamo_db_stream.0.batch_size", fmt.Sprintf("%d", batchSizeModified)),
					resource.TestCheckResourceAttr(resourceName, "source_parameters.0.dynamo_db_stream.0.maximum_batching_window_in_seconds", fmt.Sprintf("%d", batchWindowModified)),
					resource.TestCheckResourceAttr(resourceName, "source_parameters.0.dynamo_db_stream.0.maximum_record_age_in_seconds", fmt.Sprintf("%d", maxRecordAgeModified)),
					resource.TestCheckResourceAttr(resourceName, "source_parameters.0.dynamo_db_stream.0.parallelization_factor", fmt.Sprintf("%d", parallelizationModified)),
					resource.TestCheckResourceAttr(resourceName, "source_parameters.0.dynamo_db_stream.0.maximum_retry_attempts", fmt.Sprintf("%d", retriesModified)),
					resource.TestCheckResourceAttr(resourceName, "source_parameters.0.dynamo_db_stream.0.starting_position", "TRIM_HORIZON"),
					resource.TestCheckResourceAttr(resourceName, "source_parameters.0.dynamo_db_stream.0.on_partial_batch_item_failure", "AUTOMATIC_BISECT"),
					resource.TestCheckResourceAttr(resourceName, "target_parameters.0.cloudwatch_logs.0.log_stream_name", name),
					resource.TestCheckResourceAttr(resourceName, "target_parameters.0.cloudwatch_logs.0.timestamp", "$.detail.timestamp"),
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

func TestAccPipesPipe_source_active_mq_target_sqs(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var pipe pipes.DescribePipeOutput
	name := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_pipes_pipe.test"

	batchSize := 8
	batchWindow := 5

	batchSizeModified := 9
	batchWindowModified := 6

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.PipesEndpointID)
			acctest.PreCheckPartitionHasService(t, names.MQ)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.PipesEndpointID, names.MQ),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckPipeDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccPipeConfig_source_active_mq_target_sqs(name, batchSize, batchWindow),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPipeExists(ctx, resourceName, &pipe),
					resource.TestCheckResourceAttrPair(resourceName, "source", "aws_mq_broker.test", "arn"),
					resource.TestCheckResourceAttrPair(resourceName, "target", "aws_sqs_queue.target", "arn"),
					resource.TestCheckResourceAttr(resourceName, "source_parameters.0.active_mq_broker.0.batch_size", fmt.Sprintf("%d", batchSize)),
					resource.TestCheckResourceAttr(resourceName, "source_parameters.0.active_mq_broker.0.maximum_batching_window_in_seconds", fmt.Sprintf("%d", batchWindow)),
					resource.TestCheckResourceAttr(resourceName, "source_parameters.0.active_mq_broker.0.queue", "test"),
					resource.TestCheckResourceAttrPair(resourceName, "source_parameters.0.active_mq_broker.0.credentials.0.basic_auth", "aws_secretsmanager_secret_version.test", "arn"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccPipeConfig_source_active_mq_target_sqs(name, batchSizeModified, batchWindowModified),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPipeExists(ctx, resourceName, &pipe),
					resource.TestCheckResourceAttrPair(resourceName, "source", "aws_mq_broker.test", "arn"),
					resource.TestCheckResourceAttrPair(resourceName, "target", "aws_sqs_queue.target", "arn"),
					resource.TestCheckResourceAttr(resourceName, "source_parameters.0.active_mq_broker.0.batch_size", fmt.Sprintf("%d", batchSizeModified)),
					resource.TestCheckResourceAttr(resourceName, "source_parameters.0.active_mq_broker.0.maximum_batching_window_in_seconds", fmt.Sprintf("%d", batchWindowModified)),
					resource.TestCheckResourceAttr(resourceName, "source_parameters.0.active_mq_broker.0.queue", "test"),
					resource.TestCheckResourceAttrPair(resourceName, "source_parameters.0.active_mq_broker.0.credentials.0.basic_auth", "aws_secretsmanager_secret_version.test", "arn"),
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

func TestAccPipesPipe_source_rabbit_mq_target_sqs(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var pipe pipes.DescribePipeOutput
	name := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_pipes_pipe.test"

	batchSize := 8
	batchWindow := 5

	batchSizeModified := 9
	batchWindowModified := 6

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.PipesEndpointID)
			acctest.PreCheckPartitionHasService(t, names.MQ)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.PipesEndpointID, names.MQ),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckPipeDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccPipeConfig_source_rabbit_mq_target_sqs(name, batchSize, batchWindow),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPipeExists(ctx, resourceName, &pipe),
					resource.TestCheckResourceAttrPair(resourceName, "source", "aws_mq_broker.test", "arn"),
					resource.TestCheckResourceAttrPair(resourceName, "target", "aws_sqs_queue.target", "arn"),
					resource.TestCheckResourceAttr(resourceName, "source_parameters.0.rabbit_mq_broker.0.batch_size", fmt.Sprintf("%d", batchSize)),
					resource.TestCheckResourceAttr(resourceName, "source_parameters.0.rabbit_mq_broker.0.maximum_batching_window_in_seconds", fmt.Sprintf("%d", batchWindow)),
					resource.TestCheckResourceAttr(resourceName, "source_parameters.0.rabbit_mq_broker.0.queue", "test"),
					resource.TestCheckResourceAttr(resourceName, "source_parameters.0.rabbit_mq_broker.0.virtual_host", "/vhost"),
					resource.TestCheckResourceAttrPair(resourceName, "source_parameters.0.rabbit_mq_broker.0.credentials.0.basic_auth", "aws_secretsmanager_secret_version.test", "arn"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccPipeConfig_source_rabbit_mq_target_sqs(name, batchSizeModified, batchWindowModified),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPipeExists(ctx, resourceName, &pipe),
					resource.TestCheckResourceAttrPair(resourceName, "source", "aws_mq_broker.test", "arn"),
					resource.TestCheckResourceAttrPair(resourceName, "target", "aws_sqs_queue.target", "arn"),
					resource.TestCheckResourceAttr(resourceName, "source_parameters.0.rabbit_mq_broker.0.batch_size", fmt.Sprintf("%d", batchSizeModified)),
					resource.TestCheckResourceAttr(resourceName, "source_parameters.0.rabbit_mq_broker.0.maximum_batching_window_in_seconds", fmt.Sprintf("%d", batchWindowModified)),
					resource.TestCheckResourceAttr(resourceName, "source_parameters.0.rabbit_mq_broker.0.queue", "test"),
					resource.TestCheckResourceAttr(resourceName, "source_parameters.0.rabbit_mq_broker.0.virtual_host", "/vhost"),
					resource.TestCheckResourceAttrPair(resourceName, "source_parameters.0.rabbit_mq_broker.0.credentials.0.basic_auth", "aws_secretsmanager_secret_version.test", "arn"),
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

func TestAccPipesPipe_source_managed_streaming_kafka_target_sqs(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var pipe pipes.DescribePipeOutput
	name := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_pipes_pipe.test"

	batchSize := 8
	batchWindow := 5

	batchSizeModified := 9
	batchWindowModified := 6

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.PipesEndpointID)
			acctest.PreCheckPartitionHasService(t, names.Kafka)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.PipesEndpointID, names.Kafka),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckPipeDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccPipeConfig_source_managed_streaming_kafka_target_sqs(name, batchSize, batchWindow),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPipeExists(ctx, resourceName, &pipe),
					resource.TestCheckResourceAttrPair(resourceName, "source", "aws_msk_cluster.test", "arn"),
					resource.TestCheckResourceAttrPair(resourceName, "target", "aws_sqs_queue.target", "arn"),
					resource.TestCheckResourceAttr(resourceName, "source_parameters.0.managed_streaming_kafka.0.batch_size", fmt.Sprintf("%d", batchSize)),
					resource.TestCheckResourceAttr(resourceName, "source_parameters.0.managed_streaming_kafka.0.maximum_batching_window_in_seconds", fmt.Sprintf("%d", batchWindow)),
					resource.TestCheckResourceAttr(resourceName, "source_parameters.0.managed_streaming_kafka.0.topic", "test"),
					resource.TestCheckResourceAttr(resourceName, "source_parameters.0.managed_streaming_kafka.0.consumer_group_id", "amazon-managed-test-group-id"),
					resource.TestCheckResourceAttr(resourceName, "source_parameters.0.managed_streaming_kafka.0.starting_position", "TRIM_HORIZON"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccPipeConfig_source_managed_streaming_kafka_target_sqs(name, batchSizeModified, batchWindowModified),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPipeExists(ctx, resourceName, &pipe),
					resource.TestCheckResourceAttrPair(resourceName, "source", "aws_msk_cluster.test", "arn"),
					resource.TestCheckResourceAttrPair(resourceName, "target", "aws_sqs_queue.target", "arn"),
					resource.TestCheckResourceAttr(resourceName, "source_parameters.0.managed_streaming_kafka.0.batch_size", fmt.Sprintf("%d", batchSizeModified)),
					resource.TestCheckResourceAttr(resourceName, "source_parameters.0.managed_streaming_kafka.0.maximum_batching_window_in_seconds", fmt.Sprintf("%d", batchWindowModified)),
					resource.TestCheckResourceAttr(resourceName, "source_parameters.0.managed_streaming_kafka.0.topic", "test"),
					resource.TestCheckResourceAttr(resourceName, "source_parameters.0.managed_streaming_kafka.0.consumer_group_id", "amazon-managed-test-group-id"),
					resource.TestCheckResourceAttr(resourceName, "source_parameters.0.managed_streaming_kafka.0.starting_position", "TRIM_HORIZON"),
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

func TestAccPipesPipe_source_self_managed_streaming_kafka_target_sqs(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var pipe pipes.DescribePipeOutput
	name := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_pipes_pipe.test"

	servers := "smk://test1:9092,test2:9092"
	batchSize := 8
	batchWindow := 5

	batchSizeModified := 9
	batchWindowModified := 6

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.PipesEndpointID)
			acctest.PreCheckPartitionHasService(t, names.Kafka)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.PipesEndpointID, names.Kafka),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckPipeDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccPipeConfig_source_self_managed_streaming_kafka_target_sqs(name, batchSize, batchWindow, servers),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPipeExists(ctx, resourceName, &pipe),
					resource.TestCheckResourceAttr(resourceName, "source", servers),
					resource.TestCheckResourceAttrPair(resourceName, "target", "aws_sqs_queue.target", "arn"),
					resource.TestCheckResourceAttr(resourceName, "source_parameters.0.self_managed_kafka.0.batch_size", fmt.Sprintf("%d", batchSize)),
					resource.TestCheckResourceAttr(resourceName, "source_parameters.0.self_managed_kafka.0.maximum_batching_window_in_seconds", fmt.Sprintf("%d", batchWindow)),
					resource.TestCheckResourceAttr(resourceName, "source_parameters.0.self_managed_kafka.0.topic", "test"),
					resource.TestCheckResourceAttr(resourceName, "source_parameters.0.self_managed_kafka.0.consumer_group_id", "self-managed-test-group-id"),
					resource.TestCheckResourceAttr(resourceName, "source_parameters.0.self_managed_kafka.0.servers.0", "test:1234"),
					resource.TestCheckResourceAttrPair(resourceName, "source_parameters.0.self_managed_kafka.0.vpc.0.security_groups.0", "aws_security_group.test", "id"),
					resource.TestCheckResourceAttr(resourceName, "source_parameters.0.self_managed_kafka.0.vpc.0.subnets.#", "2"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccPipeConfig_source_self_managed_streaming_kafka_target_sqs(name, batchSizeModified, batchWindowModified, servers),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPipeExists(ctx, resourceName, &pipe),
					resource.TestCheckResourceAttr(resourceName, "source", servers),
					resource.TestCheckResourceAttrPair(resourceName, "target", "aws_sqs_queue.target", "arn"),
					resource.TestCheckResourceAttr(resourceName, "source_parameters.0.self_managed_kafka.0.batch_size", fmt.Sprintf("%d", batchSizeModified)),
					resource.TestCheckResourceAttr(resourceName, "source_parameters.0.self_managed_kafka.0.maximum_batching_window_in_seconds", fmt.Sprintf("%d", batchWindowModified)),
					resource.TestCheckResourceAttr(resourceName, "source_parameters.0.self_managed_kafka.0.consumer_group_id", "self-managed-test-group-id"),
					resource.TestCheckResourceAttr(resourceName, "source_parameters.0.self_managed_kafka.0.servers.0", "test:1234"),
					resource.TestCheckResourceAttrPair(resourceName, "source_parameters.0.self_managed_kafka.0.vpc.0.security_groups.0", "aws_security_group.test", "id"),
					resource.TestCheckResourceAttr(resourceName, "source_parameters.0.self_managed_kafka.0.vpc.0.subnets.#", "2"),
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

func TestAccPipesPipe_source_sqs_target_batch_job(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var pipe pipes.DescribePipeOutput
	name := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_pipes_pipe.test"

	attempts := 2
	size := 3
	parameterKey := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	parameterValue := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	command := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	environmentName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	environmentValue := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	instanceType := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	attemptsModified := 4
	sizeModified := 5
	parameterKeyModified := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	parameterValueModified := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	commandModified := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	environmentNameModified := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	environmentValueModified := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	instanceTypeModified := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

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
				Config: testAccPipeConfig_source_sqs_target_batch_job(
					name,
					attempts,
					size,
					parameterKey,
					parameterValue,
					command,
					environmentName,
					environmentValue,
					instanceType,
				),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPipeExists(ctx, resourceName, &pipe),
					resource.TestCheckResourceAttrPair(resourceName, "source", "aws_sqs_queue.source", "arn"),
					resource.TestCheckResourceAttrPair(resourceName, "target", "aws_batch_job_queue.test", "arn"),
					resource.TestCheckResourceAttrPair(resourceName, "target_parameters.0.batch_target.0.job_definition", "aws_batch_job_definition.test", "arn"),
					resource.TestCheckResourceAttr(resourceName, "target_parameters.0.batch_target.0.job_name", name),
					resource.TestCheckResourceAttr(resourceName, "target_parameters.0.batch_target.0.retry_strategy.0.attempts", fmt.Sprintf("%d", attempts)),
					resource.TestCheckResourceAttr(resourceName, "target_parameters.0.batch_target.0.array_properties.0.size", fmt.Sprintf("%d", size)),
					resource.TestCheckResourceAttr(resourceName, "target_parameters.0.batch_target.0.parameters.0.key", parameterKey),
					resource.TestCheckResourceAttr(resourceName, "target_parameters.0.batch_target.0.parameters.0.value", parameterValue),
					resource.TestCheckResourceAttr(resourceName, "target_parameters.0.batch_target.0.depends_on.0.job_id", name),
					resource.TestCheckResourceAttr(resourceName, "target_parameters.0.batch_target.0.depends_on.0.type", "SEQUENTIAL"),
					resource.TestCheckResourceAttr(resourceName, "target_parameters.0.batch_target.0.container_overrides.0.command.0", command),
					resource.TestCheckResourceAttr(resourceName, "target_parameters.0.batch_target.0.container_overrides.0.environment.0.name", environmentName),
					resource.TestCheckResourceAttr(resourceName, "target_parameters.0.batch_target.0.container_overrides.0.environment.0.value", environmentValue),
					resource.TestCheckResourceAttr(resourceName, "target_parameters.0.batch_target.0.container_overrides.0.instance_type", instanceType),
					resource.TestCheckResourceAttr(resourceName, "target_parameters.0.batch_target.0.container_overrides.0.resource_requirements.0.type", "VCPU"),
					resource.TestCheckResourceAttr(resourceName, "target_parameters.0.batch_target.0.container_overrides.0.resource_requirements.0.value", "4"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccPipeConfig_source_sqs_target_batch_job(
					name,
					attemptsModified,
					sizeModified,
					parameterKeyModified,
					parameterValueModified,
					commandModified,
					environmentNameModified,
					environmentValueModified,
					instanceTypeModified,
				),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPipeExists(ctx, resourceName, &pipe),
					resource.TestCheckResourceAttrPair(resourceName, "source", "aws_sqs_queue.source", "arn"),
					resource.TestCheckResourceAttrPair(resourceName, "target", "aws_batch_job_queue.test", "arn"),
					resource.TestCheckResourceAttrPair(resourceName, "target_parameters.0.batch_target.0.job_definition", "aws_batch_job_definition.test", "arn"),
					resource.TestCheckResourceAttr(resourceName, "target_parameters.0.batch_target.0.job_name", name),
					resource.TestCheckResourceAttr(resourceName, "target_parameters.0.batch_target.0.retry_strategy.0.attempts", fmt.Sprintf("%d", attemptsModified)),
					resource.TestCheckResourceAttr(resourceName, "target_parameters.0.batch_target.0.array_properties.0.size", fmt.Sprintf("%d", sizeModified)),
					resource.TestCheckResourceAttr(resourceName, "target_parameters.0.batch_target.0.parameters.0.key", parameterKeyModified),
					resource.TestCheckResourceAttr(resourceName, "target_parameters.0.batch_target.0.parameters.0.value", parameterValueModified),
					resource.TestCheckResourceAttr(resourceName, "target_parameters.0.batch_target.0.depends_on.0.job_id", name),
					resource.TestCheckResourceAttr(resourceName, "target_parameters.0.batch_target.0.depends_on.0.type", "SEQUENTIAL"),
					resource.TestCheckResourceAttr(resourceName, "target_parameters.0.batch_target.0.container_overrides.0.command.0", commandModified),
					resource.TestCheckResourceAttr(resourceName, "target_parameters.0.batch_target.0.container_overrides.0.environment.0.name", environmentNameModified),
					resource.TestCheckResourceAttr(resourceName, "target_parameters.0.batch_target.0.container_overrides.0.environment.0.value", environmentValueModified),
					resource.TestCheckResourceAttr(resourceName, "target_parameters.0.batch_target.0.container_overrides.0.instance_type", instanceTypeModified),
					resource.TestCheckResourceAttr(resourceName, "target_parameters.0.batch_target.0.container_overrides.0.resource_requirements.0.type", "VCPU"),
					resource.TestCheckResourceAttr(resourceName, "target_parameters.0.batch_target.0.container_overrides.0.resource_requirements.0.value", "4"),
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

func TestAccPipesPipe_source_sqs_target_event_bridge_event_bus(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var pipe pipes.DescribePipeOutput
	name := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_pipes_pipe.test"

	detailType := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	source := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	time := "$.detail.time"

	detailTypeModified := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	sourceModified := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	timeModified := "$.detail.timestamp"

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
				Config: testAccPipeConfig_source_sqs_target_event_bridge_event_bus(
					name,
					detailType,
					source,
					time,
				),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPipeExists(ctx, resourceName, &pipe),
					resource.TestCheckResourceAttrPair(resourceName, "source", "aws_sqs_queue.source", "arn"),
					resource.TestCheckResourceAttrPair(resourceName, "target", "aws_cloudwatch_event_bus.test", "arn"),
					resource.TestCheckResourceAttr(resourceName, "target_parameters.0.event_bridge_event_bus.0.detail_type", detailType),
					resource.TestCheckResourceAttr(resourceName, "target_parameters.0.event_bridge_event_bus.0.endpoint_id", ""),
					resource.TestCheckResourceAttrPair(resourceName, "target_parameters.0.event_bridge_event_bus.0.resources.0", "aws_sqs_queue.source", "arn"),
					resource.TestCheckResourceAttr(resourceName, "target_parameters.0.event_bridge_event_bus.0.source", source),
					resource.TestCheckResourceAttr(resourceName, "target_parameters.0.event_bridge_event_bus.0.time", time),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccPipeConfig_source_sqs_target_event_bridge_event_bus(
					name,
					detailTypeModified,
					sourceModified,
					timeModified,
				),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPipeExists(ctx, resourceName, &pipe),
					resource.TestCheckResourceAttrPair(resourceName, "source", "aws_sqs_queue.source", "arn"),
					resource.TestCheckResourceAttrPair(resourceName, "target", "aws_cloudwatch_event_bus.test", "arn"),
					resource.TestCheckResourceAttr(resourceName, "target_parameters.0.event_bridge_event_bus.0.detail_type", detailTypeModified),
					resource.TestCheckResourceAttr(resourceName, "target_parameters.0.event_bridge_event_bus.0.endpoint_id", ""),
					resource.TestCheckResourceAttrPair(resourceName, "target_parameters.0.event_bridge_event_bus.0.resources.0", "aws_sqs_queue.source", "arn"),
					resource.TestCheckResourceAttr(resourceName, "target_parameters.0.event_bridge_event_bus.0.source", sourceModified),
					resource.TestCheckResourceAttr(resourceName, "target_parameters.0.event_bridge_event_bus.0.time", timeModified),
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

func TestAccPipesPipe_source_sqs_target_http(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var pipe pipes.DescribePipeOutput
	name := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_pipes_pipe.test"

	headerKey := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	headerValue := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	queryStringKey := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	queryStringValue := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	headerKeyModified := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	headerValueModified := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	queryStringKeyModified := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	queryStringValueModified := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

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
				Config: testAccPipeConfig_source_sqs_target_http(
					name,
					headerKey,
					headerValue,
					queryStringKey,
					queryStringValue,
				),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPipeExists(ctx, resourceName, &pipe),
					resource.TestCheckResourceAttrPair(resourceName, "source", "aws_sqs_queue.source", "arn"),
					resource.TestCheckResourceAttr(resourceName, "target_parameters.0.http_parameters.0.header.0.key", headerKey),
					resource.TestCheckResourceAttr(resourceName, "target_parameters.0.http_parameters.0.header.0.value", headerValue),
					resource.TestCheckResourceAttr(resourceName, "target_parameters.0.http_parameters.0.path_parameters.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "target_parameters.0.http_parameters.0.query_string.0.key", queryStringKey),
					resource.TestCheckResourceAttr(resourceName, "target_parameters.0.http_parameters.0.query_string.0.value", queryStringValue),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccPipeConfig_source_sqs_target_http(
					name,
					headerKeyModified,
					headerValueModified,
					queryStringKeyModified,
					queryStringValueModified,
				),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPipeExists(ctx, resourceName, &pipe),
					resource.TestCheckResourceAttr(resourceName, "target_parameters.0.http_parameters.0.header.0.key", headerKeyModified),
					resource.TestCheckResourceAttr(resourceName, "target_parameters.0.http_parameters.0.header.0.value", headerValueModified),
					resource.TestCheckResourceAttr(resourceName, "target_parameters.0.http_parameters.0.path_parameters.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "target_parameters.0.http_parameters.0.query_string.0.key", queryStringKeyModified),
					resource.TestCheckResourceAttr(resourceName, "target_parameters.0.http_parameters.0.query_string.0.value", queryStringValueModified),
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

func TestAccPipesPipe_source_sqs_target_lambda_function(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var pipe pipes.DescribePipeOutput
	name := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_pipes_pipe.test"

	invocationType := "REQUEST_RESPONSE"
	invocationTypeModified := "FIRE_AND_FORGET"

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
				Config: testAccPipeConfig_source_sqs_target_lambda_function(
					name,
					invocationType,
				),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPipeExists(ctx, resourceName, &pipe),
					resource.TestCheckResourceAttrPair(resourceName, "source", "aws_sqs_queue.source", "arn"),
					resource.TestCheckResourceAttrPair(resourceName, "target", "aws_lambda_function.test", "arn"),
					resource.TestCheckResourceAttr(resourceName, "target_parameters.0.lambda_function.0.invocation_type", invocationType),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccPipeConfig_source_sqs_target_lambda_function(
					name,
					invocationTypeModified,
				),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPipeExists(ctx, resourceName, &pipe),
					resource.TestCheckResourceAttrPair(resourceName, "source", "aws_sqs_queue.source", "arn"),
					resource.TestCheckResourceAttrPair(resourceName, "target", "aws_lambda_function.test", "arn"),
					resource.TestCheckResourceAttr(resourceName, "target_parameters.0.lambda_function.0.invocation_type", invocationTypeModified),
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

func TestAccPipesPipe_source_sqs_target_redshift(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var pipe pipes.DescribePipeOutput
	name := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_pipes_pipe.test"

	withEvent := false
	withEventModified := true

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
				Config: testAccPipeConfig_source_sqs_target_redshift(
					name,
					withEvent,
				),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPipeExists(ctx, resourceName, &pipe),
					resource.TestCheckResourceAttrPair(resourceName, "source", "aws_sqs_queue.source", "arn"),
					resource.TestCheckResourceAttrPair(resourceName, "target", "aws_redshift_cluster.test", "arn"),
					resource.TestCheckResourceAttr(resourceName, "target_parameters.0.redshift_data.0.database", "redshiftdb"),
					resource.TestCheckResourceAttr(resourceName, "target_parameters.0.redshift_data.0.sqls.0", "SELECT * FROM table"),
					resource.TestCheckResourceAttr(resourceName, "target_parameters.0.redshift_data.0.statement_name", "NewStatement"),
					resource.TestCheckResourceAttr(resourceName, "target_parameters.0.redshift_data.0.database_user", "someUser"),
					resource.TestCheckResourceAttrPair(resourceName, "target_parameters.0.redshift_data.0.secret_manager_arn", "aws_secretsmanager_secret_version.test", "arn"),
					resource.TestCheckResourceAttr(resourceName, "target_parameters.0.redshift_data.0.with_event", fmt.Sprintf("%t", withEvent)),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccPipeConfig_source_sqs_target_redshift(
					name,
					withEventModified,
				),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPipeExists(ctx, resourceName, &pipe),
					resource.TestCheckResourceAttrPair(resourceName, "source", "aws_sqs_queue.source", "arn"),
					resource.TestCheckResourceAttrPair(resourceName, "target", "aws_redshift_cluster.test", "arn"),
					resource.TestCheckResourceAttr(resourceName, "target_parameters.0.redshift_data.0.database", "redshiftdb"),
					resource.TestCheckResourceAttr(resourceName, "target_parameters.0.redshift_data.0.sqls.0", "SELECT * FROM table"),
					resource.TestCheckResourceAttr(resourceName, "target_parameters.0.redshift_data.0.statement_name", "NewStatement"),
					resource.TestCheckResourceAttr(resourceName, "target_parameters.0.redshift_data.0.database_user", "someUser"),
					resource.TestCheckResourceAttrPair(resourceName, "target_parameters.0.redshift_data.0.secret_manager_arn", "aws_secretsmanager_secret_version.test", "arn"),
					resource.TestCheckResourceAttr(resourceName, "target_parameters.0.redshift_data.0.with_event", fmt.Sprintf("%t", withEventModified)),
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

func TestAccPipesPipe_source_sqs_target_step_function(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var pipe pipes.DescribePipeOutput
	name := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_pipes_pipe.test"

	invocationType := "REQUEST_RESPONSE"
	invocationTypeModified := "FIRE_AND_FORGET"

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
				Config: testAccPipeConfig_source_sqs_target_step_function(
					name,
					invocationType,
				),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPipeExists(ctx, resourceName, &pipe),
					resource.TestCheckResourceAttrPair(resourceName, "source", "aws_sqs_queue.source", "arn"),
					resource.TestCheckResourceAttrPair(resourceName, "target", "aws_sfn_state_machine.test", "arn"),
					resource.TestCheckResourceAttr(resourceName, "target_parameters.0.step_function.0.invocation_type", invocationType),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccPipeConfig_source_sqs_target_step_function(
					name,
					invocationTypeModified,
				),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPipeExists(ctx, resourceName, &pipe),
					resource.TestCheckResourceAttrPair(resourceName, "source", "aws_sqs_queue.source", "arn"),
					resource.TestCheckResourceAttrPair(resourceName, "target", "aws_sfn_state_machine.test", "arn"),
					resource.TestCheckResourceAttr(resourceName, "target_parameters.0.step_function.0.invocation_type", invocationTypeModified),
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

func TestAccPipesPipe_source_sqs_target_ecs_task(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var pipe pipes.DescribePipeOutput
	name := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_pipes_pipe.test"

	enableEcsManagedTags := true
	enableExecuteCommand := false
	launchType := "FARGATE"
	propagateTags := "TASK_DEFINITION"
	referenceId := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	taskCount := 1
	tagKey := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	tagValue := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	enableEcsManagedTagsModified := false
	enableExecuteCommandModified := true
	referenceIdModified := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	taskCountModified := 2
	tagKeyModified := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	tagValueModified := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

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
				Config: testAccPipeConfig_source_sqs_target_ecs_task(
					name,
					enableEcsManagedTags,
					enableExecuteCommand,
					launchType,
					propagateTags,
					referenceId,
					taskCount,
					tagKey,
					tagValue,
				),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPipeExists(ctx, resourceName, &pipe),
					resource.TestCheckResourceAttrPair(resourceName, "source", "aws_sqs_queue.source", "arn"),
					resource.TestCheckResourceAttrPair(resourceName, "target", "aws_ecs_cluster.test", "id"),
					resource.TestCheckResourceAttrPair(resourceName, "target_parameters.0.ecs_task.0.task_definition_arn", "aws_ecs_task_definition.task", "arn"),
					resource.TestCheckResourceAttr(resourceName, "target_parameters.0.ecs_task.0.enable_ecs_managed_tags", fmt.Sprintf("%t", enableEcsManagedTags)),
					resource.TestCheckResourceAttr(resourceName, "target_parameters.0.ecs_task.0.enable_execute_command", fmt.Sprintf("%t", enableExecuteCommand)),
					resource.TestCheckNoResourceAttr(resourceName, "target_parameters.0.ecs_task.0.group"),
					resource.TestCheckResourceAttr(resourceName, "target_parameters.0.ecs_task.0.launch_type", launchType),
					resource.TestCheckNoResourceAttr(resourceName, "target_parameters.0.ecs_task.0.network_configuration.0.aws_vpc_configuration.0.assign_public_ip"),
					resource.TestCheckResourceAttr(resourceName, "target_parameters.0.ecs_task.0.network_configuration.0.aws_vpc_configuration.0.security_groups.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "target_parameters.0.ecs_task.0.network_configuration.0.aws_vpc_configuration.0.subnets.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "target_parameters.0.ecs_task.0.propagate_tags", propagateTags),
					resource.TestCheckResourceAttr(resourceName, "target_parameters.0.ecs_task.0.reference_id", referenceId),
					resource.TestCheckResourceAttr(resourceName, "target_parameters.0.ecs_task.0.task_count", fmt.Sprintf("%d", taskCount)),
					resource.TestCheckResourceAttr(resourceName, "target_parameters.0.ecs_task.0.tags.0.key", tagKey),
					resource.TestCheckResourceAttr(resourceName, "target_parameters.0.ecs_task.0.tags.0.value", tagValue),
					resource.TestCheckNoResourceAttr(resourceName, "target_parameters.0.ecs_task.0.overrides"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccPipeConfig_source_sqs_target_ecs_task(
					name,
					enableEcsManagedTagsModified,
					enableExecuteCommandModified,
					launchType,
					propagateTags,
					referenceIdModified,
					taskCountModified,
					tagKeyModified,
					tagValueModified,
				),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPipeExists(ctx, resourceName, &pipe),
					resource.TestCheckResourceAttrPair(resourceName, "source", "aws_sqs_queue.source", "arn"),
					resource.TestCheckResourceAttrPair(resourceName, "target", "aws_ecs_cluster.test", "id"),
					resource.TestCheckResourceAttrPair(resourceName, "target_parameters.0.ecs_task.0.task_definition_arn", "aws_ecs_task_definition.task", "arn"),
					resource.TestCheckResourceAttr(resourceName, "target_parameters.0.ecs_task.0.enable_ecs_managed_tags", fmt.Sprintf("%t", enableEcsManagedTagsModified)),
					resource.TestCheckResourceAttr(resourceName, "target_parameters.0.ecs_task.0.enable_execute_command", fmt.Sprintf("%t", enableExecuteCommandModified)),
					resource.TestCheckNoResourceAttr(resourceName, "target_parameters.0.ecs_task.0.group"),
					resource.TestCheckResourceAttr(resourceName, "target_parameters.0.ecs_task.0.launch_type", launchType),
					resource.TestCheckNoResourceAttr(resourceName, "target_parameters.0.ecs_task.0.network_configuration.0.aws_vpc_configuration.0.assign_public_ip"),
					resource.TestCheckResourceAttr(resourceName, "target_parameters.0.ecs_task.0.network_configuration.0.aws_vpc_configuration.0.security_groups.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "target_parameters.0.ecs_task.0.network_configuration.0.aws_vpc_configuration.0.subnets.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "target_parameters.0.ecs_task.0.propagate_tags", propagateTags),
					resource.TestCheckResourceAttr(resourceName, "target_parameters.0.ecs_task.0.reference_id", referenceIdModified),
					resource.TestCheckResourceAttr(resourceName, "target_parameters.0.ecs_task.0.task_count", fmt.Sprintf("%d", taskCountModified)),
					resource.TestCheckResourceAttr(resourceName, "target_parameters.0.ecs_task.0.tags.0.key", tagKeyModified),
					resource.TestCheckResourceAttr(resourceName, "target_parameters.0.ecs_task.0.tags.0.value", tagValueModified),
					resource.TestCheckNoResourceAttr(resourceName, "target_parameters.0.ecs_task.0.overrides"),
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

func TestAccPipesPipe_targetParameters_inputTemplate(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var pipe pipes.DescribePipeOutput
	name := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
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
				Config: testAccPipeConfig_targetParameters_inputTemplate(name, "$.first"),
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
				Config: testAccPipeConfig_targetParameters_inputTemplate(name, "$.second"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPipeExists(ctx, resourceName, &pipe),
					resource.TestCheckResourceAttr(resourceName, "target_parameters.0.input_template", "$.second"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccPipeConfig_basic(name),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPipeExists(ctx, resourceName, &pipe),
					resource.TestCheckNoResourceAttr(resourceName, "target_parameters.0.input_template"),
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
		conn := acctest.Provider.Meta().(*conns.AWSClient).PipesClient()

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

		conn := acctest.Provider.Meta().(*conns.AWSClient).PipesClient()

		output, err := tfpipes.FindPipeByName(ctx, conn, rs.Primary.ID)
		if err != nil {
			return err
		}

		*pipe = *output

		return nil
	}
}

func testAccPreCheck(ctx context.Context, t *testing.T) {
	conn := acctest.Provider.Meta().(*conns.AWSClient).PipesClient()

	input := &pipes.ListPipesInput{}
	_, err := conn.ListPipes(ctx, input)

	if acctest.PreCheckSkipError(err) {
		t.Skipf("skipping acceptance testing: %s", err)
	}

	if err != nil {
		t.Fatalf("unexpected PreCheck error: %s", err)
	}
}

const testAccPipeConfig_base = `
data "aws_caller_identity" "main" {}
data "aws_partition" "main" {}

resource "aws_iam_role" "test" {
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
`

const testAccPipeConfig_base_sqsSource = `
resource "aws_iam_role_policy" "source" {
  role = aws_iam_role.test.id
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

resource "aws_sqs_queue" "source" {}
`

const testAccPipeConfig_base_sqsTarget = `
resource "aws_iam_role_policy" "target" {
  role = aws_iam_role.test.id
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

resource "aws_sqs_queue" "target" {}
`

const testAccPipeConfig_base_deadletter = `
resource "aws_iam_role_policy" "deadletter" {
  role = aws_iam_role.test.id
  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Effect = "Allow"
        Action = [
          "sqs:*"
        ],
        Resource = [
          aws_sqs_queue.deadletter.arn
        ]
      },
    ]
  })
}

resource "aws_sqs_queue" "deadletter" {}`

func testAccPipeConfig_base_kafka(name string) string {
	return acctest.ConfigCompose(acctest.ConfigVPCWithSubnets(name, 2), fmt.Sprintf(`
resource "aws_iam_role_policy" "test" {
  role = aws_iam_role.test.name

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

resource "aws_security_group" "test" {
  name   = %[1]q
  vpc_id = aws_vpc.test.id

  tags = {
    Name = %[1]q
  }
}
`, name))
}

func testAccPipeConfig_base_activeMQ(name string) string {
	return fmt.Sprintf(`
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
  engine_version          = "5.15.0"
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
`, name)
}

func testAccPipeConfig_base_rabbitMQ(name string) string {
	return fmt.Sprintf(`
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
`, name)
}

func testAccPipeConfig_base_batch(name string) string {
	return acctest.ConfigCompose(
		acctest.ConfigVPCWithSubnets(name, 2),
		fmt.Sprintf(`
data "aws_partition" "current" {}

resource "aws_iam_role" "ecs_iam_role" {
  name = "ecs_%[1]s"

  assume_role_policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Action": "sts:AssumeRole",
      "Effect": "Allow",
      "Principal": {
        "Service": "ec2.${data.aws_partition.current.dns_suffix}"
      }
    }
  ]
}
EOF
}

resource "aws_iam_role_policy_attachment" "ecs_policy_attachment" {
  role       = aws_iam_role.ecs_iam_role.name
  policy_arn = "arn:${data.aws_partition.current.partition}:iam::aws:policy/service-role/AmazonEC2ContainerServiceforEC2Role"
}

resource "aws_iam_instance_profile" "iam_instance_profile" {
  name = "ecs_%[1]s"
  role = aws_iam_role.ecs_iam_role.name
}

resource "aws_iam_role" "batch_iam_role" {
  name = "batch_%[1]s"

  assume_role_policy = <<EOF
{
    "Version": "2012-10-17",
    "Statement": [
    {
        "Action": "sts:AssumeRole",
        "Effect": "Allow",
        "Principal": {
          "Service": "batch.${data.aws_partition.current.dns_suffix}"
        }
    }
    ]
}
EOF
}

resource "aws_iam_role_policy_attachment" "batch_policy_attachment" {
  role       = aws_iam_role.batch_iam_role.name
  policy_arn = "arn:${data.aws_partition.current.partition}:iam::aws:policy/service-role/AWSBatchServiceRole"
}

resource "aws_batch_compute_environment" "test" {
  compute_environment_name = "%[1]s"

  compute_resources {
    instance_role = aws_iam_instance_profile.iam_instance_profile.arn

    instance_type = [
      "c4.large",
    ]

    max_vcpus = 16
    min_vcpus = 0

    security_group_ids = [
      aws_security_group.test.id,
    ]

    subnets = aws_subnet.test[*].id

    type = "EC2"
  }

  service_role = aws_iam_role.batch_iam_role.arn
  type         = "MANAGED"
  depends_on   = [aws_iam_role_policy_attachment.batch_policy_attachment]
}

resource "aws_batch_job_queue" "test" {
  name                 = "%[1]s"
  state                = "ENABLED"
  priority             = 1
  compute_environments = [aws_batch_compute_environment.test.arn]
}

resource "aws_batch_job_definition" "test" {
  name = "%[1]s"
  type = "container"

  container_properties = <<CONTAINER_PROPERTIES
{
  "command": ["ls", "-la"],
  "image": "busybox",
  "memory": 512,
  "vcpus": 1,
  "volumes": [ ],
  "environment": [ ],
  "mountPoints": [ ],
  "ulimits": [ ]
}
CONTAINER_PROPERTIES
}
`, name),
	)
}

func testAccPipeConfig_base_ecs(name string) string {
	return acctest.ConfigCompose(
		acctest.ConfigVPCWithSubnets(name, 2),
		acctest.ConfigLatestAmazonLinuxHVMEBSAMI(),
		fmt.Sprintf(`
resource "aws_iam_role_policy" "test" {
  name = %[1]q
  role = aws_iam_role.test.id

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

resource "aws_ecs_cluster" "test" {
  name = %[1]q
}

resource "aws_ecs_task_definition" "task" {
  family                   = %[1]q
  cpu                      = 256
  memory                   = 512
  requires_compatibilities = ["FARGATE"]
  network_mode             = "awsvpc"

  container_definitions = <<EOF
[
  {
    "name": "first",
    "image": "service-first",
    "cpu": 10,
    "memory": 512,
    "essential": true
  }
]
EOF
}

data "aws_partition" "current" {}

resource "aws_launch_template" "test" {
  name          = %[1]q
  image_id      = data.aws_ami.amzn-ami-minimal-hvm-ebs.id
  instance_type = "t2.micro"
}
`, name))
}

func testAccPipeConfig_base_http(name string) string {
	return fmt.Sprintf(`
resource "aws_api_gateway_rest_api" "test" {
  name = %[1]q
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

resource "aws_api_gateway_deployment" "test" {
  rest_api_id = aws_api_gateway_rest_api.test.id

  triggers = {
    redeployment = sha1(jsonencode(aws_api_gateway_rest_api.test.body))
  }

  lifecycle {
    create_before_destroy = true
  }
}

resource "aws_api_gateway_stage" "test" {
  deployment_id = aws_api_gateway_deployment.test.id
  rest_api_id   = aws_api_gateway_rest_api.test.id
  stage_name    = "test"
}

data "aws_partition" "current" {}
`, name)
}

func testAccPipeConfig_base_lambda(name string) string {
	return fmt.Sprintf(`
resource "aws_lambda_function" "test" {
  function_name    = %[1]q
  filename         = "test-fixtures/lambdatest.zip"
  source_code_hash = filebase64sha256("test-fixtures/lambdatest.zip")
  role             = aws_iam_role.lambda.arn
  handler          = "exports.example"
  runtime          = "nodejs16.x"
}

resource "aws_iam_role" "lambda" {
  name = %[1]q

  assume_role_policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Action": "sts:AssumeRole",
      "Principal": {
        "Service": "lambda.${data.aws_partition.current.dns_suffix}"
      },
      "Effect": "Allow",
      "Sid": ""
    }
  ]
}
EOF
}

data "aws_partition" "current" {}
`, name)
}

func testAccPipeConfig_base_redshift(name string) string {
	return acctest.ConfigCompose(acctest.ConfigVPCWithSubnets(name, 2), fmt.Sprintf(`
data "aws_partition" "current" {}

resource "aws_redshift_subnet_group" "test" {
  name       = %[1]q
  subnet_ids = aws_subnet.test[*].id
}

resource "aws_internet_gateway" "test" {
  vpc_id = aws_vpc.test.id

  tags = {
    Name = %[1]q
  }
}

resource "aws_redshift_cluster" "test" {
  cluster_identifier                  = %[1]q
  cluster_subnet_group_name           = aws_redshift_subnet_group.test.name
  database_name                       = "test"
  master_username                     = "tfacctest"
  master_password                     = "Mustbe8characters"
  node_type                           = "dc2.large"
  automated_snapshot_retention_period = 0
  allow_version_upgrade               = false
  skip_final_snapshot                 = true

  depends_on = [aws_internet_gateway.test, aws_subnet.test, aws_vpc.test]
}

resource "aws_secretsmanager_secret" "test" {
  name = %[1]q
}

resource "aws_secretsmanager_secret_version" "test" {
  secret_id     = aws_secretsmanager_secret.test.id
  secret_string = jsonencode({ password = "Mustbe8characters" })
}
`, name))
}

func testAccPipeConfig_base_step_function(name string) string {
	return fmt.Sprintf(`
resource "aws_sfn_state_machine" "test" {
  name     = "my-state-machine"
  role_arn = aws_iam_role.function.arn

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

resource "aws_iam_role" "function" {
  name = %[1]q

  assume_role_policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Action": "sts:AssumeRole",
      "Principal": {
        "Service": "states.${data.aws_partition.current.dns_suffix}"
      },
      "Effect": "Allow",
      "Sid": ""
    }
  ]
}
EOF
}

data "aws_partition" "current" {}
`, name)
}

func testAccPipeConfig_basic(name string) string {
	return acctest.ConfigCompose(
		testAccPipeConfig_base,
		testAccPipeConfig_base_sqsSource,
		testAccPipeConfig_base_sqsTarget,
		fmt.Sprintf(`
resource "aws_pipes_pipe" "test" {
  depends_on = [aws_iam_role_policy.source, aws_iam_role_policy.target]
  name       = %[1]q
  role_arn   = aws_iam_role.test.arn
  source     = aws_sqs_queue.source.arn
  target     = aws_sqs_queue.target.arn
}
`, name),
	)
}

func testAccPipeConfig_description(name, description string) string {
	return acctest.ConfigCompose(
		testAccPipeConfig_base,
		testAccPipeConfig_base_sqsSource,
		testAccPipeConfig_base_sqsTarget,
		fmt.Sprintf(`
resource "aws_pipes_pipe" "test" {
  depends_on = [aws_iam_role_policy.source, aws_iam_role_policy.target]
  name       = %[1]q
  role_arn   = aws_iam_role.test.arn
  source     = aws_sqs_queue.source.arn
  target     = aws_sqs_queue.target.arn

  description = %[2]q
}
`, name, description),
	)
}

func testAccPipeConfig_desiredState(name, state string) string {
	return acctest.ConfigCompose(
		testAccPipeConfig_base,
		testAccPipeConfig_base_sqsSource,
		testAccPipeConfig_base_sqsTarget,
		fmt.Sprintf(`
resource "aws_pipes_pipe" "test" {
  depends_on = [aws_iam_role_policy.source, aws_iam_role_policy.target]
  name       = %[1]q
  role_arn   = aws_iam_role.test.arn
  source     = aws_sqs_queue.source.arn
  target     = aws_sqs_queue.target.arn

  desired_state = %[2]q
}
`, name, state),
	)
}

func testAccPipeConfig_enrichment(
	name string,
	i int,
	headerKey string,
	headerValue string,
	queryStringKey string,
	queryStringValue string,
) string {
	return acctest.ConfigCompose(
		testAccPipeConfig_base,
		testAccPipeConfig_base_sqsSource,
		testAccPipeConfig_base_sqsTarget,
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

locals {
  name_prefix = %[1]q
}

resource "aws_cloudwatch_event_api_destination" "test" {
  count               = 2
  name                = "${local.name_prefix}-${count.index}"
  invocation_endpoint = "https://example.com/${count.index}"
  http_method         = "POST"
  connection_arn      = aws_cloudwatch_event_connection.test.arn
}

resource "aws_pipes_pipe" "test" {
  depends_on = [aws_iam_role_policy.source, aws_iam_role_policy.target]
  name       = %[1]q
  role_arn   = aws_iam_role.test.arn
  source     = aws_sqs_queue.source.arn
  target     = aws_sqs_queue.target.arn

  source_parameters {
	sqs_queue {
		batch_size 						   = 1
		maximum_batching_window_in_seconds = 0
	}
  }

  enrichment = aws_cloudwatch_event_api_destination.test[%[2]d].arn

  enrichment_parameters {
    http_parameters {
      header_parameters = {
        %[3]q = %[4]q
      }

      path_parameter_values = ["parameter1"]

      query_string_parameters = {
        %[5]q = %[6]q
      }
    }
  }
}
`, name,
			i,
			headerKey,
			headerValue,
			queryStringKey,
			queryStringValue),
	)
}

func testAccPipeConfig_sourceParameters_filterCriteria1(name, criteria1 string) string {
	return acctest.ConfigCompose(
		testAccPipeConfig_base,
		testAccPipeConfig_base_sqsSource,
		testAccPipeConfig_base_sqsTarget,
		fmt.Sprintf(`
resource "aws_pipes_pipe" "test" {
  depends_on = [aws_iam_role_policy.source, aws_iam_role_policy.target]
  name       = %[1]q
  role_arn   = aws_iam_role.test.arn
  source     = aws_sqs_queue.source.arn
  target     = aws_sqs_queue.target.arn

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
`, name, criteria1),
	)
}

func testAccPipeConfig_sourceParameters_filterCriteria2(name, criteria1, criteria2 string) string {
	return acctest.ConfigCompose(
		testAccPipeConfig_base,
		testAccPipeConfig_base_sqsSource,
		testAccPipeConfig_base_sqsTarget,
		fmt.Sprintf(`
resource "aws_pipes_pipe" "test" {
  depends_on = [aws_iam_role_policy.source, aws_iam_role_policy.target]
  name       = %[1]q
  role_arn   = aws_iam_role.test.arn
  source     = aws_sqs_queue.source.arn
  target     = aws_sqs_queue.target.arn

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
`, name, criteria1, criteria2),
	)
}

func testAccPipeConfig_nameGenerated() string {
	return acctest.ConfigCompose(
		testAccPipeConfig_base,
		testAccPipeConfig_base_sqsSource,
		testAccPipeConfig_base_sqsTarget,
		`
resource "aws_pipes_pipe" "test" {
  depends_on = [aws_iam_role_policy.source, aws_iam_role_policy.target]
  role_arn   = aws_iam_role.test.arn
  source     = aws_sqs_queue.source.arn
  target     = aws_sqs_queue.target.arn
}
`,
	)
}

func testAccPipeConfig_namePrefix(namePrefix string) string {
	return acctest.ConfigCompose(
		testAccPipeConfig_base,
		testAccPipeConfig_base_sqsSource,
		testAccPipeConfig_base_sqsTarget,
		fmt.Sprintf(`
resource "aws_pipes_pipe" "test" {
  depends_on  = [aws_iam_role_policy.source, aws_iam_role_policy.target]
  name_prefix = %[1]q
  role_arn    = aws_iam_role.test.arn
  source      = aws_sqs_queue.source.arn
  target      = aws_sqs_queue.target.arn
}
`, namePrefix),
	)
}

func testAccPipeConfig_roleARN(name string) string {
	return acctest.ConfigCompose(
		testAccPipeConfig_base,
		testAccPipeConfig_base_sqsSource,
		testAccPipeConfig_base_sqsTarget,
		fmt.Sprintf(`
resource "aws_iam_role" "test2" {
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
  name       = %[1]q
  role_arn   = aws_iam_role.test2.arn
  source     = aws_sqs_queue.source.arn
  target     = aws_sqs_queue.target.arn
}
`, name),
	)
}

func testAccPipeConfig_tags1(name, tag1Key, tag1Value string) string {
	return acctest.ConfigCompose(
		testAccPipeConfig_base,
		testAccPipeConfig_base_sqsSource,
		testAccPipeConfig_base_sqsTarget,
		fmt.Sprintf(`
resource "aws_pipes_pipe" "test" {
  depends_on = [aws_iam_role_policy.source, aws_iam_role_policy.target]
  name       = %[1]q
  role_arn   = aws_iam_role.test.arn
  source     = aws_sqs_queue.source.arn
  target     = aws_sqs_queue.target.arn

  tags = {
    %[2]q = %[3]q
  }
}
`, name, tag1Key, tag1Value),
	)
}

func testAccPipeConfig_tags2(name, tag1Key, tag1Value, tag2Key, tag2Value string) string {
	return acctest.ConfigCompose(
		testAccPipeConfig_base,
		testAccPipeConfig_base_sqsSource,
		testAccPipeConfig_base_sqsTarget,
		fmt.Sprintf(`
resource "aws_pipes_pipe" "test" {
  depends_on = [aws_iam_role_policy.source, aws_iam_role_policy.target]
  name       = %[1]q
  role_arn   = aws_iam_role.test.arn
  source     = aws_sqs_queue.source.arn
  target     = aws_sqs_queue.target.arn

  tags = {
    %[2]q = %[3]q
    %[4]q = %[5]q
  }
}
`, name, tag1Key, tag1Value, tag2Key, tag2Value),
	)
}

func testAccPipeConfig_source_sqs_target_sqs(
	name string,
	batchSize int,
	batchingWindow int,
	dedupeID string,
	groupID string,
) string {
	return acctest.ConfigCompose(
		testAccPipeConfig_base,
		testAccPipeConfig_base_sqsSource,
		testAccPipeConfig_base_sqsTarget,
		fmt.Sprintf(`
resource "aws_iam_role_policy" "target_fifo" {
  role = aws_iam_role.test.id
  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Effect = "Allow"
        Action = [
          "sqs:SendMessage",
        ],
        Resource = [
          aws_sqs_queue.target_fifo.arn,
        ]
      },
    ]
  })
}

resource "aws_sqs_queue" "target_fifo" {
  fifo_queue                  = true
  content_based_deduplication = true
}

resource "aws_pipes_pipe" "test" {
  depends_on = [aws_iam_role_policy.source, aws_iam_role_policy.target_fifo]
  name       = %[1]q
  role_arn   = aws_iam_role.test.arn
  source     = aws_sqs_queue.source.arn
  target     = aws_sqs_queue.target_fifo.arn

  source_parameters {
    sqs_queue {
      batch_size                         = %[2]d
      maximum_batching_window_in_seconds = %[3]d
    }
  }

  target_parameters {
    sqs_queue {
      message_deduplication_id = %[4]q
      message_group_id         = %[5]q
    }
  }
}
`, name, batchSize, batchingWindow, dedupeID, groupID),
	)
}

func testAccPipeConfig_source_kinesis_target_kinesis(
	name string,
	batchSize int,
	batchingWindow int,
	maxRecordAge int,
	parallelization int,
	retries int,
	partitionKey string,
) string {
	return acctest.ConfigCompose(
		testAccPipeConfig_base,
		testAccPipeConfig_base_deadletter,
		fmt.Sprintf(`
resource "aws_iam_role_policy" "source" {
  role = aws_iam_role.test.id
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
  name = "source-%[1]s"

  stream_mode_details {
    stream_mode = "ON_DEMAND"
  }
}

resource "aws_iam_role_policy" "target" {
  role = aws_iam_role.test.id
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
  name = "target-%[1]s"

  stream_mode_details {
    stream_mode = "ON_DEMAND"
  }
}

resource "aws_pipes_pipe" "test" {
  depends_on = [aws_iam_role_policy.source, aws_iam_role_policy.target, aws_iam_role_policy.deadletter, aws_sqs_queue.deadletter]
  name       = %[1]q
  role_arn   = aws_iam_role.test.arn
  source     = aws_kinesis_stream.source.arn
  target     = aws_kinesis_stream.target.arn

  source_parameters {
    kinesis_stream {
      batch_size                         = %[2]d
      maximum_batching_window_in_seconds = %[3]d
      maximum_record_age_in_seconds      = %[4]d
      parallelization_factor             = %[5]d
      maximum_retry_attempts             = %[6]d
      on_partial_batch_item_failure      = "AUTOMATIC_BISECT"
      starting_position                  = "AT_TIMESTAMP"
      starting_position_timestamp        = "2023-01-01T00:00:00Z"

      dead_letter_config {
        arn = aws_sqs_queue.deadletter.arn
      }
    }
  }

  target_parameters {
    kinesis_stream {
      partition_key = %[7]q
    }
  }
}
`, name,
			batchSize,
			batchingWindow,
			maxRecordAge,
			parallelization,
			retries,
			partitionKey,
		),
	)
}

func testAccPipeConfig_source_dynamo_target_cloudwatch_logs(
	name string,
	batchSize int,
	batchingWindow int,
	maxRecordAge int,
	parallelization int,
	retries int,
) string {
	return acctest.ConfigCompose(
		testAccPipeConfig_base,
		testAccPipeConfig_base_deadletter,
		fmt.Sprintf(`
resource "aws_iam_role_policy" "source" {
  role = aws_iam_role.test.id
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
  name             = %[1]q
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

resource "aws_cloudwatch_log_group" "target" {
  name = %[1]q
}

resource "aws_cloudwatch_log_stream" "target" {
  name           = %[1]q
  log_group_name = aws_cloudwatch_log_group.target.name
}

resource "aws_iam_role_policy" "target" {
  role = aws_iam_role.test.id
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

resource "aws_pipes_pipe" "test" {
  depends_on = [aws_iam_role_policy.source, aws_iam_role_policy.target, aws_iam_role_policy.deadletter, aws_sqs_queue.deadletter]
  name       = %[1]q
  role_arn   = aws_iam_role.test.arn
  source     = aws_dynamodb_table.source.stream_arn
  target     = aws_cloudwatch_log_group.target.arn

  source_parameters {
    dynamo_db_stream {
      batch_size                         = %[2]d
      maximum_batching_window_in_seconds = %[3]d
      maximum_record_age_in_seconds      = %[4]d
      parallelization_factor             = %[5]d
      maximum_retry_attempts             = %[6]d
      on_partial_batch_item_failure      = "AUTOMATIC_BISECT"
      starting_position                  = "TRIM_HORIZON"

      dead_letter_config {
        arn = aws_sqs_queue.deadletter.arn
      }
    }
  }

  target_parameters {
    cloudwatch_logs {
      log_stream_name = %[1]q
      timestamp       = "$.detail.timestamp"
    }
  }
}
`, name,
			batchSize,
			batchingWindow,
			maxRecordAge,
			parallelization,
			retries),
	)
}

func testAccPipeConfig_source_active_mq_target_sqs(
	name string,
	batchSize int,
	batchingWindow int,
) string {
	return acctest.ConfigCompose(
		testAccPipeConfig_base,
		testAccPipeConfig_base_activeMQ(name),
		testAccPipeConfig_base_sqsTarget,
		fmt.Sprintf(`
resource "aws_pipes_pipe" "test" {
  depends_on  = [aws_iam_role_policy.test, aws_iam_role_policy.target]
  name_prefix = %[1]q
  role_arn    = aws_iam_role.test.arn
  source      = aws_mq_broker.test.arn
  target      = aws_sqs_queue.target.arn

  source_parameters {
    active_mq_broker {
      batch_size                         = %[2]d
      maximum_batching_window_in_seconds = %[3]d
      queue                              = "test"

      credentials {
        basic_auth = aws_secretsmanager_secret_version.test.arn
      }
    }
  }
}
`, name, batchSize, batchingWindow),
	)
}

func testAccPipeConfig_source_rabbit_mq_target_sqs(
	name string,
	batchSize int,
	batchingWindow int,
) string {
	return acctest.ConfigCompose(
		testAccPipeConfig_base,
		testAccPipeConfig_base_rabbitMQ(name),
		testAccPipeConfig_base_sqsTarget,
		fmt.Sprintf(`
resource "aws_pipes_pipe" "test" {
  depends_on  = [aws_iam_role_policy.test, aws_iam_role_policy.target]
  name_prefix = %[1]q
  role_arn    = aws_iam_role.test.arn
  source      = aws_mq_broker.test.arn
  target      = aws_sqs_queue.target.arn

  source_parameters {
    rabbit_mq_broker {
      batch_size                         = %[2]d
      maximum_batching_window_in_seconds = %[3]d
      queue                              = "test"
      virtual_host                       = "/vhost"

      credentials {
        basic_auth = aws_secretsmanager_secret_version.test.arn
      }
    }
  }
}
`, name, batchSize, batchingWindow),
	)
}

func testAccPipeConfig_source_managed_streaming_kafka_target_sqs(
	name string,
	batchSize int,
	batchingWindow int,
) string {
	return acctest.ConfigCompose(
		testAccPipeConfig_base,
		testAccPipeConfig_base_kafka(name),
		testAccPipeConfig_base_sqsTarget,
		fmt.Sprintf(`
resource "aws_msk_cluster" "test" {
  cluster_name           = %[1]q
  kafka_version          = "2.7.1"
  number_of_broker_nodes = 2
  depends_on = [aws_subnet.test, aws_vpc.test, aws_security_group.test]

  broker_node_group_info {
    client_subnets  = aws_subnet.test[*].id
    ebs_volume_size = 10
    instance_type   = "kafka.t3.small"
    security_groups = [aws_security_group.test.id]
  }
}

resource "aws_pipes_pipe" "test" {
  depends_on  = [aws_iam_role_policy.test, aws_iam_role_policy.target, aws_msk_cluster.test]
  name_prefix = %[1]q
  role_arn    = aws_iam_role.test.arn
  source      = aws_msk_cluster.test.arn
  target      = aws_sqs_queue.target.arn

  source_parameters {
    managed_streaming_kafka {
      batch_size                         = %[2]d
      maximum_batching_window_in_seconds = %[3]d
      topic                              = "test"
      consumer_group_id                  = "amazon-managed-test-group-id"
      starting_position                  = "TRIM_HORIZON"
    }
  }
}
`, name, batchSize, batchingWindow),
	)
}

func testAccPipeConfig_source_self_managed_streaming_kafka_target_sqs(
	name string,
	batchSize int,
	batchingWindow int,
	servers string,
) string {
	return acctest.ConfigCompose(
		testAccPipeConfig_base,
		testAccPipeConfig_base_kafka(name),
		testAccPipeConfig_base_sqsTarget,
		fmt.Sprintf(`
resource "aws_pipes_pipe" "test" {
  depends_on  = [aws_iam_role_policy.test, aws_iam_role_policy.target]
  name_prefix = %[1]q
  role_arn    = aws_iam_role.test.arn
  source      = %[4]q
  target      = aws_sqs_queue.target.arn

  source_parameters {
    self_managed_kafka {
      batch_size                         = %[2]d
      maximum_batching_window_in_seconds = %[3]d
      topic                              = "test"
      consumer_group_id                  = "self-managed-test-group-id"
      starting_position                  = "TRIM_HORIZON"
      servers                            = ["test:1234"]

      vpc {
        security_groups = [aws_security_group.test.id]
        subnets         = aws_subnet.test[*].id
      }
    }
  }
}
`, name, batchSize, batchingWindow, servers),
	)
}

func testAccPipeConfig_source_sqs_target_batch_job(
	name string,
	attempts int,
	size int,
	parameterKey string,
	parameterValue string,
	command string,
	environmentName string,
	environmentValue string,
	instanceType string,
) string {
	return acctest.ConfigCompose(
		testAccPipeConfig_base,
		testAccPipeConfig_base_sqsSource,
		testAccPipeConfig_base_batch(name),
		fmt.Sprintf(`
resource "aws_pipes_pipe" "test" {
  depends_on = [aws_iam_role_policy.source, aws_batch_job_queue.test, aws_batch_job_definition.test]
  name       = %[1]q
  role_arn   = aws_iam_role.test.arn
  source     = aws_sqs_queue.source.arn
  target     = aws_batch_job_queue.test.arn

  target_parameters {
    batch_target {
      job_definition = aws_batch_job_definition.test.arn
      job_name       = %[1]q
      retry_strategy {
        attempts = %[2]d
      }
      array_properties {
        size = %[3]d
      }
      parameters {
        key   = %[4]q
        value = %[5]q
      }
      depends_on {
        job_id = %[1]q
        type   = "SEQUENTIAL"
      }
      container_overrides {
        command = [%[6]q]
        environment {
          name  = %[7]q
          value = %[8]q
        }
        instance_type = %[9]q
        resource_requirements {
          type  = "VCPU"
          value = "4"
        }
      }
    }
  }
}
`, name,
			attempts,
			size,
			parameterKey,
			parameterValue,
			command,
			environmentName,
			environmentValue,
			instanceType,
		),
	)
}

func testAccPipeConfig_source_sqs_target_event_bridge_event_bus(
	name string,
	detailType string,
	source string,
	time string,
) string {
	return acctest.ConfigCompose(
		testAccPipeConfig_base,
		testAccPipeConfig_base_sqsSource,
		fmt.Sprintf(`
resource "aws_cloudwatch_event_bus" "test" {
  name = %[1]q
}

resource "aws_iam_role_policy" "target" {
  role = aws_iam_role.test.id
  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Effect = "Allow"
        Action = [
          "events:PutEvent",
        ],
        Resource = [
          aws_cloudwatch_event_bus.test.arn,
        ]
      },
    ]
  })
}

resource "aws_pipes_pipe" "test" {
  depends_on = [aws_iam_role_policy.source, aws_iam_role_policy.target, aws_cloudwatch_event_bus.test]
  name       = %[1]q
  role_arn   = aws_iam_role.test.arn
  source     = aws_sqs_queue.source.arn
  target     = aws_cloudwatch_event_bus.test.arn

  target_parameters {
    event_bridge_event_bus {
      detail_type = %[2]q
      resources   = [aws_sqs_queue.source.arn]
      source      = %[3]q
      time        = %[4]q
    }
  }
}
`, name,
			detailType,
			source,
			time,
		),
	)
}

func testAccPipeConfig_source_sqs_target_http(
	name string,
	headerKey string,
	headerValue string,
	queryStringKey string,
	queryStringValue string,
) string {
	return acctest.ConfigCompose(
		testAccPipeConfig_base,
		testAccPipeConfig_base_sqsSource,
		testAccPipeConfig_base_http(name),
		fmt.Sprintf(`
resource "aws_pipes_pipe" "test" {
  depends_on = [aws_iam_role_policy.source, aws_api_gateway_stage.test]
  name       = %[1]q
  role_arn   = aws_iam_role.test.arn
  source     = aws_sqs_queue.source.arn
  target     = "${aws_api_gateway_stage.test.execution_arn}/GET/*"

  target_parameters {
    http_parameters {
      header {
        key   = %[2]q
        value = %[3]q
      }

      path_parameters = ["param"]

      query_string {
        key   = %[4]q
        value = %[5]q
      }
    }
  }
}
`, name,
			headerKey,
			headerValue,
			queryStringKey,
			queryStringValue,
		),
	)
}

func testAccPipeConfig_source_sqs_target_lambda_function(
	name string,
	invocationType string,
) string {
	return acctest.ConfigCompose(
		testAccPipeConfig_base,
		testAccPipeConfig_base_sqsSource,
		testAccPipeConfig_base_lambda(name),
		fmt.Sprintf(`
resource "aws_pipes_pipe" "test" {
  depends_on = [aws_iam_role_policy.source, aws_lambda_function.test]
  name       = %[1]q
  role_arn   = aws_iam_role.test.arn
  source     = aws_sqs_queue.source.arn
  target     = aws_lambda_function.test.arn

  target_parameters {
    lambda_function {
      invocation_type = %[2]q
    }
  }
}
`, name,
			invocationType,
		),
	)
}

func testAccPipeConfig_source_sqs_target_redshift(
	name string,
	withEvent bool,
) string {
	return acctest.ConfigCompose(
		testAccPipeConfig_base,
		testAccPipeConfig_base_sqsSource,
		testAccPipeConfig_base_redshift(name),
		fmt.Sprintf(`
resource "aws_pipes_pipe" "test" {
  depends_on = [aws_iam_role_policy.source, aws_redshift_cluster.test]
  name       = %[1]q
  role_arn   = aws_iam_role.test.arn
  source     = aws_sqs_queue.source.arn
  target     = aws_redshift_cluster.test.arn

  target_parameters {
    redshift_data {
      database           = "redshiftdb"
      sqls               = ["SELECT * FROM table"]
      statement_name     = "NewStatement"
      database_user      = "someUser"
      secret_manager_arn = aws_secretsmanager_secret_version.test.arn
      with_event         = %[2]t
    }
  }
}
`, name,
			withEvent,
		),
	)
}

func testAccPipeConfig_source_sqs_target_step_function(
	name string,
	invocationType string,
) string {
	return acctest.ConfigCompose(
		testAccPipeConfig_base,
		testAccPipeConfig_base_sqsSource,
		testAccPipeConfig_base_step_function(name),
		fmt.Sprintf(`
resource "aws_pipes_pipe" "test" {
  depends_on = [aws_iam_role_policy.source, aws_sfn_state_machine.test]
  name       = %[1]q
  role_arn   = aws_iam_role.test.arn
  source     = aws_sqs_queue.source.arn
  target     = aws_sfn_state_machine.test.arn

  target_parameters {
    step_function {
      invocation_type = %[2]q
    }
  }
}
`, name,
			invocationType,
		),
	)
}

func testAccPipeConfig_source_sqs_target_ecs_task(
	name string,
	enableEcsManagedTags bool,
	enableExecuteCommand bool,
	launchType string,
	propagateTags string,
	referenceId string,
	taskCount int,
	tagKey string,
	tagValue string,
) string {
	return acctest.ConfigCompose(
		testAccPipeConfig_base,
		testAccPipeConfig_base_sqsSource,
		testAccPipeConfig_base_ecs(name),
		fmt.Sprintf(`
resource "aws_pipes_pipe" "test" {
  depends_on = [aws_iam_role_policy.source, aws_ecs_cluster.test]
  name       = %[1]q
  role_arn   = aws_iam_role.test.arn
  source     = aws_sqs_queue.source.arn
  target     = aws_ecs_cluster.test.id

  target_parameters {
    ecs_task {
      task_definition_arn = aws_ecs_task_definition.task.arn

      enable_ecs_managed_tags = %[2]t
      enable_execute_command  = %[3]t
      launch_type             = %[4]q
      propagate_tags          = %[5]q
      reference_id            = %[6]q
      task_count              = %[7]d

      network_configuration {
        aws_vpc_configuration {
          subnets = aws_subnet.test[*].id
        }
      }

      tags {
        key   = %[8]q
        value = %[9]q
      }
    }
  }
}
`, name,
			enableEcsManagedTags,
			enableExecuteCommand,
			launchType,
			propagateTags,
			referenceId,
			taskCount,
			tagKey,
			tagValue,
		),
	)
}

func testAccPipeConfig_targetParameters_inputTemplate(name, template string) string {
	return acctest.ConfigCompose(
		testAccPipeConfig_base,
		testAccPipeConfig_base_sqsSource,
		testAccPipeConfig_base_sqsTarget,
		fmt.Sprintf(`
resource "aws_pipes_pipe" "test" {
  depends_on = [aws_iam_role_policy.source, aws_iam_role_policy.target]
  name       = %[1]q
  role_arn   = aws_iam_role.test.arn
  source     = aws_sqs_queue.source.arn
  target     = aws_sqs_queue.target.arn

  target_parameters {
    input_template = %[2]q
  }
}
`, name, template),
	)
}
