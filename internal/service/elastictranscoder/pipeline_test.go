package elastictranscoder_test

import (
	"fmt"
	"reflect"
	"regexp"
	"sort"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/elastictranscoder"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfelastictranscoder "github.com/hashicorp/terraform-provider-aws/internal/service/elastictranscoder"
)

func TestAccElasticTranscoderPipeline_basic(t *testing.T) {
	pipeline := &elastictranscoder.Pipeline{}
	resourceName := "aws_elastictranscoder_pipeline.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, elastictranscoder.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckPipelineDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccPipelineBasicConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPipelineExists(resourceName, pipeline),
					acctest.MatchResourceAttrRegionalARN(resourceName, "arn", "elastictranscoder", regexp.MustCompile(`pipeline/.+`)),
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
	pipeline := &elastictranscoder.Pipeline{}
	resourceName := "aws_elastictranscoder_pipeline.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	keyResourceName := "aws_kms_key.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, elastictranscoder.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckPipelineDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccPipelineKMSKeyConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPipelineExists(resourceName, pipeline),
					resource.TestCheckResourceAttrPair(resourceName, "aws_kms_key_arn", keyResourceName, "arn"),
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
	pipeline := elastictranscoder.Pipeline{}
	resourceName := "aws_elastictranscoder_pipeline.test"

	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, elastictranscoder.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckPipelineDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccNotificationsConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPipelineExists(resourceName, &pipeline),
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
				Config: testAccNotificationsUpdateConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPipelineExists(resourceName, &pipeline),
					testAccCheckPipeline_notifications(&pipeline, []string{"completed"}),
				),
			},
		},
	})
}

// testAccCheckTags can be used to check the tags on a resource.
func testAccCheckPipeline_notifications(
	p *elastictranscoder.Pipeline, notifications []string) resource.TestCheckFunc {
	return func(s *terraform.State) error {

		var notes []string
		if aws.StringValue(p.Notifications.Completed) != "" {
			notes = append(notes, "completed")
		}
		if aws.StringValue(p.Notifications.Error) != "" {
			notes = append(notes, "error")
		}
		if aws.StringValue(p.Notifications.Progressing) != "" {
			notes = append(notes, "progressing")
		}
		if aws.StringValue(p.Notifications.Warning) != "" {
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
	pipeline := &elastictranscoder.Pipeline{}
	resourceName := "aws_elastictranscoder_pipeline.test"

	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, elastictranscoder.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckPipelineDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccPipelineWithContentConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPipelineExists(resourceName, pipeline),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccPipelineWithContentUpdateConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPipelineExists(resourceName, pipeline),
				),
			},
		},
	})
}

func TestAccElasticTranscoderPipeline_withPermissions(t *testing.T) {
	pipeline := &elastictranscoder.Pipeline{}
	resourceName := "aws_elastictranscoder_pipeline.test"

	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, elastictranscoder.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckPipelineDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccPipelineWithPermsConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPipelineExists(resourceName, pipeline),
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
	pipeline := &elastictranscoder.Pipeline{}
	resourceName := "aws_elastictranscoder_pipeline.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, elastictranscoder.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckPipelineDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccPipelineBasicConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPipelineExists(resourceName, pipeline),
					acctest.CheckResourceDisappears(acctest.Provider, tfelastictranscoder.ResourcePipeline(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckPipelineExists(n string, res *elastictranscoder.Pipeline) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No Pipeline ID is set")
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).ElasticTranscoderConn

		out, err := conn.ReadPipeline(&elastictranscoder.ReadPipelineInput{
			Id: aws.String(rs.Primary.ID),
		})

		if err != nil {
			return err
		}

		*res = *out.Pipeline

		return nil
	}
}

func testAccCheckPipelineDestroy(s *terraform.State) error {
	conn := acctest.Provider.Meta().(*conns.AWSClient).ElasticTranscoderConn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_elastictranscoder_pipline" {
			continue
		}

		out, err := conn.ReadPipeline(&elastictranscoder.ReadPipelineInput{
			Id: aws.String(rs.Primary.ID),
		})
		if tfawserr.ErrCodeEquals(err, elastictranscoder.ErrCodeResourceNotFoundException) {
			continue
		}
		if err != nil {
			return fmt.Errorf("unexpected error: %w", err)
		}

		if out.Pipeline != nil && aws.StringValue(out.Pipeline.Id) == rs.Primary.ID {
			return fmt.Errorf("Elastic Transcoder Pipeline still exists")
		}
	}
	return nil
}

func testAccPreCheck(t *testing.T) {
	conn := acctest.Provider.Meta().(*conns.AWSClient).ElasticTranscoderConn

	input := &elastictranscoder.ListPipelinesInput{}

	_, err := conn.ListPipelines(input)

	if acctest.PreCheckSkipError(err) {
		t.Skipf("skipping acceptance testing: %s", err)
	}

	if err != nil {
		t.Fatalf("unexpected PreCheck error: %s", err)
	}
}

func testAccPipelineBasicConfig(rName string) string {
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

resource "aws_s3_bucket_acl" "test" {
  bucket = aws_s3_bucket.test.id
  acl    = "private"
}
`, rName)
}

func testAccPipelineKMSKeyConfig(rName string) string {
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

resource "aws_s3_bucket_acl" "test" {
  bucket = aws_s3_bucket.test.id
  acl    = "private"
}
`, rName)
}

func testAccPipelineWithContentConfig(rName string) string {
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

resource "aws_s3_bucket_acl" "content_bucket_acl" {
  bucket = aws_s3_bucket.content_bucket.id
  acl    = "private"
}

resource "aws_s3_bucket" "input_bucket" {
  bucket = "%[1]s-input"
}

resource "aws_s3_bucket_acl" "input_bucket_acl" {
  bucket = aws_s3_bucket.input_bucket.id
  acl    = "private"
}

resource "aws_s3_bucket" "thumb_bucket" {
  bucket = "%[1]s-thumb"
}

resource "aws_s3_bucket_acl" "thumb_bucket_acl" {
  bucket = aws_s3_bucket.thumb_bucket.id
  acl    = "private"
}
`, rName)
}

func testAccPipelineWithContentUpdateConfig(rName string) string {
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

resource "aws_s3_bucket_acl" "content_bucket_acl" {
  bucket = aws_s3_bucket.content_bucket.id
  acl    = "private"
}

resource "aws_s3_bucket" "input_bucket" {
  bucket = "%[1]s-input"
}

resource "aws_s3_bucket_acl" "input_bucket_acl" {
  bucket = aws_s3_bucket.input_bucket.id
  acl    = "private"
}

resource "aws_s3_bucket" "thumb_bucket" {
  bucket = "%[1]s-thumb"
}

resource "aws_s3_bucket_acl" "thumb_bucket_acl" {
  bucket = aws_s3_bucket.thumb_bucket.id
  acl    = "private"
}
`, rName)
}

func testAccPipelineWithPermsConfig(rName string) string {
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

resource "aws_s3_bucket_acl" "test" {
  bucket = aws_s3_bucket.test.id
  acl    = "private"
}
`, rName)
}

func testAccNotificationsConfig(rName string) string {
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

resource "aws_s3_bucket_acl" "test" {
  bucket = aws_s3_bucket.test.id
  acl    = "private"
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

func testAccNotificationsUpdateConfig(rName string) string {
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

resource "aws_s3_bucket_acl" "test" {
  bucket = aws_s3_bucket.test.id
  acl    = "private"
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
