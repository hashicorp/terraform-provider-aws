package pipes_test

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/pipes"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"

	tfpipes "github.com/hashicorp/terraform-provider-aws/internal/service/pipes"
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
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttrPair(resourceName, "role_arn", "aws_iam_role.test", "arn"),
					resource.TestCheckResourceAttrPair(resourceName, "source", "aws_sqs_queue.source", "arn"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
					resource.TestCheckResourceAttrPair(resourceName, "target", "aws_sqs_queue.target", "arn"),
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
					resource.TestCheckResourceAttr(resourceName, "name_prefix", resource.UniqueIdPrefix),
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

func TestAccPipesPipe_target(t *testing.T) {
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
					resource.TestCheckResourceAttrPair(resourceName, "target", "aws_sqs_queue.target", "arn"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccPipeConfig_target(name),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPipeExists(ctx, resourceName, &pipe),
					resource.TestCheckResourceAttrPair(resourceName, "target", "aws_sqs_queue.target2", "arn"),
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

func testAccPipeConfig_target(name string) string {
	return acctest.ConfigCompose(
		testAccPipeConfig_base,
		testAccPipeConfig_base_sqsSource,
		testAccPipeConfig_base_sqsTarget,
		fmt.Sprintf(`
resource "aws_iam_role_policy" "target2" {
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
          aws_sqs_queue.target2.arn,
        ]
      },
    ]
  })
}

resource "aws_sqs_queue" "target2" {}

resource "aws_pipes_pipe" "test" {
  depends_on = [aws_iam_role_policy.source, aws_iam_role_policy.target2]
  name       = %[1]q
  role_arn   = aws_iam_role.test.arn
  source     = aws_sqs_queue.source.arn
  target     = aws_sqs_queue.target2.arn
}
`, name),
	)
}
