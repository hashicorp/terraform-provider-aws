// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package elastictranscoder_test

import (
	"context"
	"fmt"
	"reflect"
	"sort"
	"testing"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/elastictranscoder"
	awstypes "github.com/aws/aws-sdk-go-v2/service/elastictranscoder/types"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	tfelastictranscoder "github.com/hashicorp/terraform-provider-aws/internal/service/elastictranscoder"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccElasticTranscoderPipeline_basic(t *testing.T) {
	ctx := acctest.Context(t)
	pipeline := &awstypes.Pipeline{}
	resourceName := "aws_elastictranscoder_pipeline.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ElasticTranscoderServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckPipelineDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccPipelineConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPipelineExists(ctx, resourceName, pipeline),
					acctest.MatchResourceAttrRegionalARN(resourceName, names.AttrARN, "elastictranscoder", regexache.MustCompile(`pipeline/.+`)),
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

func TestAccElasticTranscoderPipeline_kmsKey(t *testing.T) {
	ctx := acctest.Context(t)
	pipeline := &awstypes.Pipeline{}
	resourceName := "aws_elastictranscoder_pipeline.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	keyResourceName := "aws_kms_key.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ElasticTranscoderServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckPipelineDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccPipelineConfig_kmsKey(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPipelineExists(ctx, resourceName, pipeline),
					resource.TestCheckResourceAttrPair(resourceName, "aws_kms_key_arn", keyResourceName, names.AttrARN),
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

func TestAccElasticTranscoderPipeline_notifications(t *testing.T) {
	ctx := acctest.Context(t)
	pipeline := awstypes.Pipeline{}
	resourceName := "aws_elastictranscoder_pipeline.test"

	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ElasticTranscoderServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckPipelineDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccPipelineConfig_notifications(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPipelineExists(ctx, resourceName, &pipeline),
					testAccCheckPipeline_notifications(&pipeline, []string{"warning", "completed"}),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			// update and check that we have 1 less notification
			{
				Config: testAccPipelineConfig_notificationsUpdate(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPipelineExists(ctx, resourceName, &pipeline),
					testAccCheckPipeline_notifications(&pipeline, []string{"completed"}),
				),
			},
		},
	})
}

// testAccCheckTags can be used to check the tags on a resource.
func testAccCheckPipeline_notifications(
	p *awstypes.Pipeline, notifications []string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		var notes []string
		if aws.ToString(p.Notifications.Completed) != "" {
			notes = append(notes, "completed")
		}
		if aws.ToString(p.Notifications.Error) != "" {
			notes = append(notes, "error")
		}
		if aws.ToString(p.Notifications.Progressing) != "" {
			notes = append(notes, "progressing")
		}
		if aws.ToString(p.Notifications.Warning) != "" {
			notes = append(notes, "warning")
		}

		if len(notes) != len(notifications) {
			return fmt.Errorf("ETC notifications didn't match:\n\texpected: %#v\n\tgot: %#v\n\n", notifications, notes)
		}

		sort.Strings(notes)
		sort.Strings(notifications)

		if !reflect.DeepEqual(notes, notifications) {
			return fmt.Errorf("ETC notifications were not equal:\n\texpected: %#v\n\tgot: %#v\n\n", notifications, notes)
		}

		return nil
	}
}

func TestAccElasticTranscoderPipeline_withContent(t *testing.T) {
	ctx := acctest.Context(t)
	pipeline := &awstypes.Pipeline{}
	resourceName := "aws_elastictranscoder_pipeline.test"

	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ElasticTranscoderServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckPipelineDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccPipelineConfig_content(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPipelineExists(ctx, resourceName, pipeline),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccPipelineConfig_contentUpdate(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPipelineExists(ctx, resourceName, pipeline),
				),
			},
		},
	})
}

func TestAccElasticTranscoderPipeline_withPermissions(t *testing.T) {
	ctx := acctest.Context(t)
	pipeline := &awstypes.Pipeline{}
	resourceName := "aws_elastictranscoder_pipeline.test"

	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ElasticTranscoderServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckPipelineDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccPipelineConfig_perms(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPipelineExists(ctx, resourceName, pipeline),
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

func TestAccElasticTranscoderPipeline_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	pipeline := &awstypes.Pipeline{}
	resourceName := "aws_elastictranscoder_pipeline.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ElasticTranscoderServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckPipelineDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccPipelineConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPipelineExists(ctx, resourceName, pipeline),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfelastictranscoder.ResourcePipeline(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckPipelineExists(ctx context.Context, n string, res *awstypes.Pipeline) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No Pipeline ID is set")
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).ElasticTranscoderClient(ctx)

		out, err := conn.ReadPipeline(ctx, &elastictranscoder.ReadPipelineInput{
			Id: aws.String(rs.Primary.ID),
		})

		if err != nil {
			return err
		}

		*res = *out.Pipeline

		return nil
	}
}

func testAccCheckPipelineDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).ElasticTranscoderClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_elastictranscoder_pipline" {
				continue
			}

			out, err := conn.ReadPipeline(ctx, &elastictranscoder.ReadPipelineInput{
				Id: aws.String(rs.Primary.ID),
			})
			if errs.IsA[*awstypes.ResourceNotFoundException](err) {
				continue
			}
			if err != nil {
				return fmt.Errorf("unexpected error: %w", err)
			}

			if out.Pipeline != nil && aws.ToString(out.Pipeline.Id) == rs.Primary.ID {
				return fmt.Errorf("Elastic Transcoder Pipeline still exists")
			}
		}
		return nil
	}
}

func testAccPreCheck(ctx context.Context, t *testing.T) {
	conn := acctest.Provider.Meta().(*conns.AWSClient).ElasticTranscoderClient(ctx)

	input := &elastictranscoder.ListPipelinesInput{}

	_, err := conn.ListPipelines(ctx, input)

	if acctest.PreCheckSkipError(err) {
		t.Skipf("skipping acceptance testing: %s", err)
	}

	if err != nil {
		t.Fatalf("unexpected PreCheck error: %s", err)
	}
}

func testAccPipelineConfig_basic(rName string) string {
	return fmt.Sprintf(`
resource "aws_elastictranscoder_pipeline" "test" {
  input_bucket  = aws_s3_bucket.test.bucket
  output_bucket = aws_s3_bucket.test.bucket
  name          = %[1]q
  role          = aws_iam_role.test.arn
}

resource "aws_iam_role" "test" {
  name = %[1]q

  assume_role_policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Action": "sts:AssumeRole",
      "Principal": {
        "Service": "ec2.amazonaws.com"
      },
      "Effect": "Allow",
      "Sid": ""
    }
  ]
}
EOF
}

resource "aws_s3_bucket" "test" {
  bucket = %[1]q
}
`, rName)
}

func testAccPipelineConfig_kmsKey(rName string) string {
	return fmt.Sprintf(`
resource "aws_kms_key" "test" {
  description = %[1]q
  policy      = <<POLICY
{
  "Version": "2012-10-17",
  "Id": "kms-tf-1",
  "Statement": [
    {
      "Sid": "Enable IAM User Permissions",
      "Effect": "Allow",
      "Principal": {
        "AWS": "*"
      },
      "Action": "kms:*",
      "Resource": "*"
    }
  ]
}
POLICY
}

resource "aws_elastictranscoder_pipeline" "test" {
  input_bucket    = aws_s3_bucket.test.bucket
  output_bucket   = aws_s3_bucket.test.bucket
  name            = %[1]q
  role            = aws_iam_role.test.arn
  aws_kms_key_arn = aws_kms_key.test.arn
}

resource "aws_iam_role" "test" {
  name = %[1]q

  assume_role_policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Action": "sts:AssumeRole",
      "Principal": {
        "Service": "ec2.amazonaws.com"
      },
      "Effect": "Allow",
      "Sid": ""
    }
  ]
}
EOF
}

resource "aws_s3_bucket" "test" {
  bucket = %[1]q
}
`, rName)
}

func testAccPipelineConfig_content(rName string) string {
	return fmt.Sprintf(`
resource "aws_elastictranscoder_pipeline" "test" {
  input_bucket = aws_s3_bucket.content_bucket.bucket
  name         = %[1]q
  role         = aws_iam_role.test.arn

  content_config {
    bucket        = aws_s3_bucket.content_bucket.bucket
    storage_class = "Standard"
  }

  thumbnail_config {
    bucket        = aws_s3_bucket.content_bucket.bucket
    storage_class = "Standard"
  }
}

resource "aws_iam_role" "test" {
  name = %[1]q

  assume_role_policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Action": "sts:AssumeRole",
      "Principal": {
        "Service": "ec2.amazonaws.com"
      },
      "Effect": "Allow",
      "Sid": ""
    }
  ]
}
EOF
}

resource "aws_s3_bucket" "content_bucket" {
  bucket = "%[1]s-content"
}

resource "aws_s3_bucket" "input_bucket" {
  bucket = "%[1]s-input"
}

resource "aws_s3_bucket" "thumb_bucket" {
  bucket = "%[1]s-thumb"
}
`, rName)
}

func testAccPipelineConfig_contentUpdate(rName string) string {
	return fmt.Sprintf(`
resource "aws_elastictranscoder_pipeline" "test" {
  input_bucket = aws_s3_bucket.input_bucket.bucket
  name         = %[1]q
  role         = aws_iam_role.test.arn

  content_config {
    bucket        = aws_s3_bucket.content_bucket.bucket
    storage_class = "Standard"
  }

  thumbnail_config {
    bucket        = aws_s3_bucket.thumb_bucket.bucket
    storage_class = "Standard"
  }
}

resource "aws_iam_role" "test" {
  name = %[1]q

  assume_role_policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Action": "sts:AssumeRole",
      "Principal": {
        "Service": "ec2.amazonaws.com"
      },
      "Effect": "Allow",
      "Sid": ""
    }
  ]
}
EOF
}

resource "aws_s3_bucket" "content_bucket" {
  bucket = "%[1]s-content"
}

resource "aws_s3_bucket" "input_bucket" {
  bucket = "%[1]s-input"
}

resource "aws_s3_bucket" "thumb_bucket" {
  bucket = "%[1]s-thumb"
}
`, rName)
}

func testAccPipelineConfig_perms(rName string) string {
	return fmt.Sprintf(`
resource "aws_elastictranscoder_pipeline" "test" {
  input_bucket = aws_s3_bucket.test.bucket
  name         = %[1]q
  role         = aws_iam_role.test.arn

  content_config {
    bucket        = aws_s3_bucket.test.bucket
    storage_class = "Standard"
  }

  content_config_permissions {
    grantee_type = "Group"
    grantee      = "AuthenticatedUsers"
    access       = ["FullControl"]
  }

  thumbnail_config {
    bucket        = aws_s3_bucket.test.bucket
    storage_class = "Standard"
  }

  thumbnail_config_permissions {
    grantee_type = "Group"
    grantee      = "AuthenticatedUsers"
    access       = ["FullControl"]
  }
}

resource "aws_iam_role" "test" {
  name = %[1]q

  assume_role_policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Action": "sts:AssumeRole",
      "Principal": {
        "Service": "ec2.amazonaws.com"
      },
      "Effect": "Allow",
      "Sid": ""
    }
  ]
}
EOF
}

resource "aws_s3_bucket" "test" {
  bucket = %[1]q
}
`, rName)
}

func testAccPipelineConfig_notifications(rName string) string {
	return fmt.Sprintf(`
resource "aws_elastictranscoder_pipeline" "test" {
  input_bucket  = aws_s3_bucket.test.bucket
  output_bucket = aws_s3_bucket.test.bucket
  name          = %[1]q
  role          = aws_iam_role.test.arn

  notifications {
    completed = aws_sns_topic.test.arn
    warning   = aws_sns_topic.test.arn
  }
}

resource "aws_iam_role" "test" {
  name = %[1]q

  assume_role_policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Action": "sts:AssumeRole",
      "Principal": {
        "Service": "ec2.amazonaws.com"
      },
      "Effect": "Allow",
      "Sid": ""
    }
  ]
}
EOF
}

resource "aws_s3_bucket" "test" {
  bucket = %[1]q
}

resource "aws_sns_topic" "test" {
  name = %[1]q

  policy = <<EOF
{
  "Version": "2012-10-17",
  "Id": "AWSAccountTopicAccess",
  "Statement": [
    {
      "Sid": "*",
      "Effect": "Allow",
      "Principal": "*",
      "Action": "sns:Publish",
      "Resource": "*"
    }
  ]
}
EOF
}
`, rName)
}

func testAccPipelineConfig_notificationsUpdate(rName string) string {
	return fmt.Sprintf(`
resource "aws_elastictranscoder_pipeline" "test" {
  input_bucket  = aws_s3_bucket.test.bucket
  output_bucket = aws_s3_bucket.test.bucket
  name          = %[1]q
  role          = aws_iam_role.test.arn

  notifications {
    completed = aws_sns_topic.test.arn
  }
}

resource "aws_iam_role" "test" {
  name = %[1]q

  assume_role_policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Action": "sts:AssumeRole",
      "Principal": {
        "Service": "ec2.amazonaws.com"
      },
      "Effect": "Allow",
      "Sid": ""
    }
  ]
}
EOF
}

resource "aws_s3_bucket" "test" {
  bucket = %[1]q
}

resource "aws_sns_topic" "test" {
  name = %[1]q

  policy = <<EOF
{
  "Version": "2012-10-17",
  "Id": "AWSAccountTopicAccess",
  "Statement": [
    {
      "Sid": "*",
      "Effect": "Allow",
      "Principal": "*",
      "Action": "sns:Publish",
      "Resource": "*"
    }
  ]
}
EOF
}
`, rName)
}
