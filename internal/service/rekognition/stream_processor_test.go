// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package rekognition_test

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/rekognition"
	"github.com/aws/aws-sdk-go-v2/service/rekognition/types"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	tfrekognition "github.com/hashicorp/terraform-provider-aws/internal/service/rekognition"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccRekognitionStreamProcessor_basic(t *testing.T) {
	ctx := acctest.Context(t)

	var streamprocessor rekognition.DescribeStreamProcessorOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_rekognition_stream_processor.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.RekognitionEndpointID)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.RekognitionServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckStreamProcessorDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccStreamProcessorConfig_basic(testAccStreamProcessorConfig_setup(rName), rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckStreamProcessorExists(ctx, resourceName, &streamprocessor),
					// resource.TestCheckResourceAttr(resourceName, "auto_minor_version_upgrade", "false"),
					// resource.TestCheckResourceAttrSet(resourceName, "maintenance_window_start_time.0.day_of_week"),
					// resource.TestCheckTypeSetElemNestedAttrs(resourceName, "user.*", map[string]string{
					// 	"console_access": "false",
					// 	"groups.#":       "0",
					// 	"username":       "Test",
					// 	"password":       "TestTest1234",
					// }),
					// acctest.MatchResourceAttrRegionalARN(resourceName, "arn", "rekognition", regexache.MustCompile(`streamprocessor:+.`)),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"apply_immediately", "user"},
			},
		},
	})
}

// func TestAccRekognitionStreamProcessor_disappears(t *testing.T) {
// 	ctx := acctest.Context(t)
// 	if testing.Short() {
// 		t.Skip("skipping long-running test in short mode")
// 	}

// 	var streamprocessor rekognition.DescribeStreamProcessorResponse
// 	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
// 	resourceName := "aws_rekognition_stream_processor.test"

// 	resource.ParallelTest(t, resource.TestCase{
// 		PreCheck: func() {
// 			acctest.PreCheck(ctx, t)
// 			acctest.PreCheckPartitionHasService(t, names.RekognitionEndpointID)
// 			testAccPreCheck(t)
// 		},
// 		ErrorCheck:               acctest.ErrorCheck(t, names.RekognitionServiceID),
// 		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
// 		CheckDestroy:             testAccCheckStreamProcessorDestroy(ctx),
// 		Steps: []resource.TestStep{
// 			{
// 				Config: testAccStreamProcessorConfig_basic(rName, testAccStreamProcessorVersionNewer),
// 				Check: resource.ComposeTestCheckFunc(
// 					testAccCheckStreamProcessorExists(ctx, resourceName, &streamprocessor),
// 					// TIP: The Plugin-Framework disappears helper is similar to the Plugin-SDK version,
// 					// but expects a new resource factory function as the third argument. To expose this
// 					// private function to the testing package, you may need to add a line like the following
// 					// to exports_test.go:
// 					//
// 					//   var ResourceStreamProcessor = newResourceStreamProcessor
// 					acctest.CheckFrameworkResourceDisappears(ctx, acctest.Provider, tfrekognition.ResourceStreamProcessor, resourceName),
// 				),
// 				ExpectNonEmptyPlan: true,
// 			},
// 		},
// 	})
// }

func testAccCheckStreamProcessorDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).RekognitionClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_rekognition_stream_processor" {
				continue
			}

			_, err := conn.DescribeStreamProcessor(ctx, &rekognition.DescribeStreamProcessorInput{
				Name: aws.String(rs.Primary.ID),
			})
			if errs.IsA[*types.ResourceNotFoundException](err) {
				return nil
			}
			if err != nil {
				return create.Error(names.Rekognition, create.ErrActionCheckingDestroyed, tfrekognition.ResNameStreamProcessor, rs.Primary.ID, err)
			}

			return create.Error(names.Rekognition, create.ErrActionCheckingDestroyed, tfrekognition.ResNameStreamProcessor, rs.Primary.ID, errors.New("not destroyed"))
		}

		return nil
	}
}

func testAccCheckStreamProcessorExists(ctx context.Context, name string, streamprocessor *rekognition.DescribeStreamProcessorOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return create.Error(names.Rekognition, create.ErrActionCheckingExistence, tfrekognition.ResNameStreamProcessor, name, errors.New("not found"))
		}

		if rs.Primary.ID == "" {
			return create.Error(names.Rekognition, create.ErrActionCheckingExistence, tfrekognition.ResNameStreamProcessor, name, errors.New("not set"))
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).RekognitionClient(ctx)
		resp, err := conn.DescribeStreamProcessor(ctx, &rekognition.DescribeStreamProcessorInput{
			Name: aws.String(rs.Primary.ID),
		})

		if err != nil {
			return create.Error(names.Rekognition, create.ErrActionCheckingExistence, tfrekognition.ResNameStreamProcessor, rs.Primary.ID, err)
		}

		*streamprocessor = *resp

		return nil
	}
}

func testAccPreCheck(ctx context.Context, t *testing.T) {
	conn := acctest.Provider.Meta().(*conns.AWSClient).RekognitionClient(ctx)

	input := &rekognition.ListStreamProcessorsInput{}
	_, err := conn.ListStreamProcessors(ctx, input)

	if acctest.PreCheckSkipError(err) {
		t.Skipf("skipping acceptance testing: %s", err)
	}
	if err != nil {
		t.Fatalf("unexpected PreCheck error: %s", err)
	}
}

func testAccCheckStreamProcessorNotRecreated(before, after *rekognition.DescribeStreamProcessorOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if before, after := aws.ToString(before.StreamProcessorArn), aws.ToString(after.StreamProcessorArn); before != after {
			return create.Error(names.Rekognition, create.ErrActionCheckingNotRecreated, tfrekognition.ResNameStreamProcessor, aws.ToString(&before), errors.New("recreated"))
		}

		return nil
	}
}

func testAccStreamProcessorConfig_setup(rName string) string {
	return fmt.Sprintf(`
resource "aws_iam_role" "test" {
  name = "%[1]q-acctest-role"

  assume_role_policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Action = "sts:AssumeRole"
        Effect = "Allow"
        Sid    = ""
        Principal = {
          Service = "ec2.amazonaws.com"
        }
      },
    ]
  })
}

resource "aws_s3_bucket" "test" {
  bucket = "%[1]q-acctest-bucket"
}

resource "aws_sns_topic" "test" {
  name = "%[1]q-acctest-topic"
}

resource "aws_kinesis_video_stream" "test" {
  name                    = "%[1]q-acctest-kinesis-input"
  data_retention_in_hours = 1
  device_name             = "kinesis-video-device-name"
  media_type              = "video/h264"
}
	`, rName)
}

func testAccStreamProcessorConfig_basic(setup, rName string) string {
	return fmt.Sprintf(`
%[1]q

resource "aws_rekognition_stream_processor" "test" {
  role_arn = aws_iam_role.test.arn
  name     = "%[1]q-acctest-processor"

  data_sharing_preference {
    opt_in = true
  }

  output {
    s3_destination {
      bucket = aws_s3_bucket.test.bucket
    }
  }

  settings {
    connected_home {
      labels = ["PERSON", "ALL"]
    }
  }

  input {
    kinesis_video_stream {
      arn = aws_kinesis_video_stream.test.arn
    }
  }

  notification_channel {
    sns_topic_arn = aws_sns_topic.test.arn
  }
}
`, setup, rName)
}
