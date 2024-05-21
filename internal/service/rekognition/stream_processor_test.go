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
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	tfrekognition "github.com/hashicorp/terraform-provider-aws/internal/service/rekognition"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccRekognitionStreamProcessor_connectedHome(t *testing.T) {
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
				Config: testAccStreamProcessorConfig_connectedHome(testAccStreamProcessorConfig_connectedHome_setup(rName), rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckStreamProcessorExists(ctx, resourceName, &streamprocessor),
					resource.TestCheckResourceAttr(resourceName, names.AttrID, fmt.Sprintf("%[1]s-acctest-processor", rName)),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, fmt.Sprintf("%[1]s-acctest-processor", rName)),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{names.AttrARN},
			},
		},
	})
}

func TestAccRekognitionStreamProcessor_connectedHome_polygon(t *testing.T) {
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
				Config: testAccStreamProcessorConfig_connectedHome_polygons(testAccStreamProcessorConfig_connectedHome_setup(rName), rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckStreamProcessorExists(ctx, resourceName, &streamprocessor),
					resource.TestCheckResourceAttr(resourceName, "regions_of_interest.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "regions_of_interest.0.polygon.#", acctest.Ct3),
				),
			},
		},
	})
}

func TestAccRekognitionStreamProcessor_connectedHome_boundingBox(t *testing.T) {
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
				Config: testAccStreamProcessorConfig_connectedHome_boundingBox(testAccStreamProcessorConfig_connectedHome_setup(rName), rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckStreamProcessorExists(ctx, resourceName, &streamprocessor),
					resource.TestCheckResourceAttr(resourceName, "regions_of_interest.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "regions_of_interest.0.polygon.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "regions_of_interest.0.bounding_box.left", "0.5"),
					resource.TestCheckResourceAttr(resourceName, "regions_of_interest.0.bounding_box.top", "0.5"),
					resource.TestCheckResourceAttr(resourceName, "regions_of_interest.0.bounding_box.height", "0.5"),
					resource.TestCheckResourceAttr(resourceName, "regions_of_interest.0.bounding_box.width", "0.5"),
				),
			},
		},
	})
}

func TestAccRekognitionStreamProcessor_disappears(t *testing.T) {
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
				Config: testAccStreamProcessorConfig_connectedHome(testAccStreamProcessorConfig_connectedHome_setup(rName), rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckStreamProcessorExists(ctx, resourceName, &streamprocessor),
					acctest.CheckFrameworkResourceDisappears(ctx, acctest.Provider, tfrekognition.ResourceStreamProcessor, resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckStreamProcessorDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).RekognitionClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_rekognition_stream_processor" {
				continue
			}

			_, err := tfrekognition.FindCollectionByID(ctx, conn, rs.Primary.ID)
			if tfresource.NotFound(err) {
				continue
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

// func testAccCheckStreamProcessorNotRecreated(before, after *rekognition.DescribeStreamProcessorOutput) resource.TestCheckFunc {
// 	return func(s *terraform.State) error {
// 		if before, after := aws.ToString(before.StreamProcessorArn), aws.ToString(after.StreamProcessorArn); before != after {
// 			return create.Error(names.Rekognition, create.ErrActionCheckingNotRecreated, tfrekognition.ResNameStreamProcessor, aws.ToString(&before), errors.New("recreated"))
// 		}

// 		return nil
// 	}
// }

func testAccStreamProcessorConfig_connectedHome_setup(rName string) string {
	return fmt.Sprintf(`
resource "aws_iam_role" "test" {
  name = "%[1]s-acctest-role"

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
  bucket = "%[1]s-acctest-bucket"
}

resource "aws_sns_topic" "test" {
  name = "%[1]s-acctest-topic"
}

resource "aws_kinesis_video_stream" "test" {
  name                    = "%[1]s-acctest-kinesis-input"
  data_retention_in_hours = 1
  device_name             = "kinesis-video-device-name"
  media_type              = "video/h264"
}
	`, rName)
}

func testAccStreamProcessorConfig_connectedHome(setup, rName string) string {
	return fmt.Sprintf(`
%[1]s

resource "aws_rekognition_stream_processor" "test" {
  role_arn = aws_iam_role.test.arn
  name     = "%[2]s-acctest-processor"

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

func testAccStreamProcessorConfig_connectedHome_polygons(setup, rName string) string {
	return fmt.Sprintf(`
%[1]s

resource "aws_rekognition_stream_processor" "test" {
  role_arn = aws_iam_role.test.arn
  name     = "%[2]s-acctest-processor"

  data_sharing_preference {
    opt_in = true
  }

  output {
    s3_destination {
      bucket = aws_s3_bucket.test.bucket
    }
  }

  regions_of_interest {
    polygon {
      x = 0.5
      y = 0.5
    }
    polygon {
      x = 0.5
      y = 0.5
    }
    polygon {
      x = 0.5
      y = 0.5
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

func testAccStreamProcessorConfig_connectedHome_boundingBox(setup, rName string) string {
	return fmt.Sprintf(`
%[1]s

resource "aws_rekognition_stream_processor" "test" {
  role_arn = aws_iam_role.test.arn
  name     = "%[2]s-acctest-processor"

  data_sharing_preference {
    opt_in = true
  }

  output {
    s3_destination {
      bucket = aws_s3_bucket.test.bucket
    }
  }

  regions_of_interest {
    bounding_box {
      left   = 0.5
      top    = 0.5
      height = 0.5
      width  = 0.5
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
