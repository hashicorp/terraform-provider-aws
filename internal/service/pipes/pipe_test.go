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
				Config: testAccPipeConfig_basic(rName),
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
				Config: testAccPipeConfig_basic(rName),
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
				Config: testAccPipeConfig_basic(rName),
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
				Config: testAccPipeConfig_basic(rName),
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
				Config: testAccPipeConfig_basic(rName),
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
				Config: testAccPipeConfig_basic(rName),
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
				Config: testAccPipeConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPipeExists(ctx, resourceName, &pipe),
					resource.TestCheckNoResourceAttr(resourceName, "target_parameters.0.input_template"),
				),
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

func testAccPipeConfig_basic(rName string) string {
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
