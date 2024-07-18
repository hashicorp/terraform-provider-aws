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

func TestAccRekognitionStreamProcessor_basic(t *testing.T) {
	ctx := acctest.Context(t)

	var streamprocessor rekognition.DescribeStreamProcessorOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_rekognition_stream_processor.test"
	s3BucketResourceName := "aws_s3_bucket.test"
	kinesisVideoStreamResourceName := "aws_kinesis_video_stream.test"
	snsTopicResourceName := "aws_sns_topic.test"

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
				Config: testAccStreamProcessorConfig_connectedHome(rName, ""),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckStreamProcessorExists(ctx, resourceName, &streamprocessor),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, "data_sharing_preference.0.opt_in", acctest.CtTrue),
					resource.TestCheckResourceAttrPair(resourceName, "output.0.s3_destination.0.bucket", s3BucketResourceName, names.AttrBucket),
					resource.TestCheckResourceAttr(resourceName, "settings.0.connected_home.0.labels.#", acctest.Ct2),
					resource.TestCheckTypeSetElemAttr(resourceName, "settings.0.connected_home.0.labels.*", "PERSON"),
					resource.TestCheckTypeSetElemAttr(resourceName, "settings.0.connected_home.0.labels.*", "ALL"),
					resource.TestCheckResourceAttrPair(resourceName, "input.0.kinesis_video_stream.0.arn", kinesisVideoStreamResourceName, names.AttrARN),
					resource.TestCheckResourceAttrPair(resourceName, "notification_channel.0.sns_topic_arn", snsTopicResourceName, names.AttrARN),
					resource.TestCheckResourceAttrSet(resourceName, "stream_processor_arn"),
				),
			},
			{
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateIdFunc:                    testAccStreamProcessorImportStateIdFunc(resourceName),
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: names.AttrName,
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
				Config: testAccStreamProcessorConfig_connectedHome(rName, ""),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckStreamProcessorExists(ctx, resourceName, &streamprocessor),
					acctest.CheckFrameworkResourceDisappears(ctx, acctest.Provider, tfrekognition.ResourceStreamProcessor, resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccRekognitionStreamProcessor_connectedHome(t *testing.T) {
	ctx := acctest.Context(t)

	var streamprocessor, streamprocessor2 rekognition.DescribeStreamProcessorOutput
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
				Config: testAccStreamProcessorConfig_connectedHome(rName, testAccStreamProcessorConfig_boundingBox()),
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
			{
				Config: testAccStreamProcessorConfig_connectedHome(rName, testAccStreamProcessorConfig_polygons()),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckStreamProcessorExists(ctx, resourceName, &streamprocessor2),
					testAccCheckStreamProcessorNotRecreated(&streamprocessor, &streamprocessor2),
					resource.TestCheckResourceAttr(resourceName, "regions_of_interest.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "regions_of_interest.0.polygon.#", acctest.Ct3),
					resource.TestCheckResourceAttr(resourceName, "regions_of_interest.0.polygon.0.x", "0.5"),
					resource.TestCheckResourceAttr(resourceName, "regions_of_interest.0.polygon.0.y", "0.5"),
					resource.TestCheckResourceAttr(resourceName, "regions_of_interest.0.polygon.1.x", "0.5"),
					resource.TestCheckResourceAttr(resourceName, "regions_of_interest.0.polygon.1.y", "0.5"),
					resource.TestCheckResourceAttr(resourceName, "regions_of_interest.0.polygon.2.x", "0.5"),
					resource.TestCheckResourceAttr(resourceName, "regions_of_interest.0.polygon.2.y", "0.5"),
				),
			},
		},
	})
}

// NOTE: Stream Processors setup for Face Detection cannot be altered after the fact
func TestAccRekognitionStreamProcessor_faceRecognition(t *testing.T) {
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
				Config: testAccStreamProcessorConfig_faceRecognition(rName, ""),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckStreamProcessorExists(ctx, resourceName, &streamprocessor),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
				),
			},
			{
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateIdFunc:                    testAccStreamProcessorImportStateIdFunc(resourceName),
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: names.AttrName,
			},
		},
	})
}

func TestAccRekognitionStreamProcessor_faceRecognition_boundingBox(t *testing.T) {
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
				Config: testAccStreamProcessorConfig_faceRecognition(rName, testAccStreamProcessorConfig_boundingBox()),
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

func TestAccRekognitionStreamProcessor_faceRecognition_polygon(t *testing.T) {
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
				Config: testAccStreamProcessorConfig_faceRecognition(rName, testAccStreamProcessorConfig_polygons()),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckStreamProcessorExists(ctx, resourceName, &streamprocessor),
					resource.TestCheckResourceAttr(resourceName, "regions_of_interest.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "regions_of_interest.0.polygon.#", acctest.Ct3),
					resource.TestCheckResourceAttr(resourceName, "regions_of_interest.0.polygon.0.x", "0.5"),
					resource.TestCheckResourceAttr(resourceName, "regions_of_interest.0.polygon.0.y", "0.5"),
					resource.TestCheckResourceAttr(resourceName, "regions_of_interest.0.polygon.1.x", "0.5"),
					resource.TestCheckResourceAttr(resourceName, "regions_of_interest.0.polygon.1.y", "0.5"),
					resource.TestCheckResourceAttr(resourceName, "regions_of_interest.0.polygon.2.x", "0.5"),
					resource.TestCheckResourceAttr(resourceName, "regions_of_interest.0.polygon.2.y", "0.5"),
				),
			},
		},
	})
}

func TestAccRekognitionStreamProcessor_tags(t *testing.T) {
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
				Config: testAccStreamProcessorConfig_tags1(rName, acctest.CtKey1, acctest.CtValue1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckStreamProcessorExists(ctx, resourceName, &streamprocessor),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey1, acctest.CtValue1),
				),
			},
			{
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateIdFunc:                    testAccStreamProcessorImportStateIdFunc(resourceName),
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: names.AttrName,
			},
			{
				Config: testAccStreamProcessorConfig_tags2(rName, acctest.CtKey1, acctest.CtValue1Updated, acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckStreamProcessorExists(ctx, resourceName, &streamprocessor),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct2),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey1, acctest.CtValue1Updated),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey2, acctest.CtValue2),
				),
			},
			{
				Config: testAccStreamProcessorConfig_tags1(rName, acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckStreamProcessorExists(ctx, resourceName, &streamprocessor),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey2, acctest.CtValue2),
				),
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

			streamName := rs.Primary.Attributes[names.AttrName]
			_, err := tfrekognition.FindStreamProcessorByName(ctx, conn, streamName)
			if tfresource.NotFound(err) {
				continue
			}

			if err != nil {
				return create.Error(names.Rekognition, create.ErrActionCheckingDestroyed, tfrekognition.ResNameStreamProcessor, streamName, err)
			}

			return create.Error(names.Rekognition, create.ErrActionCheckingDestroyed, tfrekognition.ResNameStreamProcessor, streamName, errors.New("not destroyed"))
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

		streamName := rs.Primary.Attributes[names.AttrName]
		if streamName == "" {
			return create.Error(names.Rekognition, create.ErrActionCheckingExistence, tfrekognition.ResNameStreamProcessor, name, errors.New("name not set"))
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).RekognitionClient(ctx)
		resp, err := tfrekognition.FindStreamProcessorByName(ctx, conn, streamName)
		if err != nil {
			return create.Error(names.Rekognition, create.ErrActionCheckingExistence, tfrekognition.ResNameStreamProcessor, streamName, err)
		}

		*streamprocessor = *resp

		return nil
	}
}

func testAccStreamProcessorImportStateIdFunc(resourceName string) resource.ImportStateIdFunc {
	return func(s *terraform.State) (string, error) {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return "", fmt.Errorf("Not found: %s", resourceName)
		}

		return rs.Primary.Attributes[names.AttrName], nil
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

func testAccStreamProcessorConfigBase_connectedHome(rName string) string {
	return fmt.Sprintf(`
resource "aws_iam_role" "test" {
  name = %[1]q

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
  bucket = %[1]q
}

resource "aws_sns_topic" "test" {
  name = %[1]q
}

resource "aws_kinesis_video_stream" "test" {
  name                    = %[1]q
  data_retention_in_hours = 1
  device_name             = "kinesis-video-device-name"
  media_type              = "video/h264"
}
`, rName)
}

func testAccStreamProcessorConfigBase_faceRecognition(rName string) string {
	return fmt.Sprintf(`
resource "aws_kinesis_video_stream" "test" {
  name                    = %[1]q
  data_retention_in_hours = 1
  device_name             = "kinesis-video-device-name"
  media_type              = "video/h264"
}

resource "aws_kinesis_stream" "test_output" {
  name        = %[1]q
  shard_count = 1
}

resource "aws_iam_role" "test" {
  name = %[1]q

  inline_policy {
    name = "Rekognition-Access"
    policy = jsonencode({
      Version = "2012-10-17"
      Statement = [
        {
          Action = [
            "kinesis:Get*",
            "kinesis:DescribeStreamSummary"
          ]
          Effect   = "Allow"
          Resource = ["${aws_kinesis_video_stream.test.arn}"]
        },
        {
          Action = [
            "kinesis:PutRecord"
          ]
          Effect   = "Allow"
          Resource = ["${aws_kinesis_stream.test_output.arn}"]
        },
      ]
    })
  }

  assume_role_policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Action = "sts:AssumeRole"
        Effect = "Allow"
        Principal = {
          Service = "rekognition.amazonaws.com"
        }
      },
    ]
  })
}

resource "aws_rekognition_collection" "test" {
  collection_id = %[1]q
}
`, rName)
}

func testAccStreamProcessorConfig_polygons() string {
	return `
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
}`
}

func testAccStreamProcessorConfig_boundingBox() string {
	return `
regions_of_interest {
  bounding_box {
    left   = 0.5
    top    = 0.5
    height = 0.5
    width  = 0.5
  }
}`
}

func testAccStreamProcessorConfig_connectedHome(rName, regionsOfInterest string) string {
	return acctest.ConfigCompose(
		testAccStreamProcessorConfigBase_connectedHome(rName),
		fmt.Sprintf(`
resource "aws_rekognition_stream_processor" "test" {
  role_arn = aws_iam_role.test.arn
  name     = %[1]q

  data_sharing_preference {
    opt_in = true
  }

  output {
    s3_destination {
      bucket = aws_s3_bucket.test.bucket
    }
  }

%[2]s

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
`, rName, regionsOfInterest))
}

func testAccStreamProcessorConfig_faceRecognition(rName, regionsOfInterest string) string {
	return acctest.ConfigCompose(
		testAccStreamProcessorConfigBase_faceRecognition(rName),
		fmt.Sprintf(`
resource "aws_rekognition_stream_processor" "test" {
  role_arn = aws_iam_role.test.arn
  name     = %[1]q

  data_sharing_preference {
    opt_in = false
  }

%[2]s

  input {
    kinesis_video_stream {
      arn = aws_kinesis_video_stream.test.arn
    }
  }

  output {
    kinesis_data_stream {
      arn = aws_kinesis_stream.test_output.arn
    }
  }

  settings {
    face_search {
      collection_id = aws_rekognition_collection.test.id
    }
  }
}
`, rName, regionsOfInterest))
}

func testAccStreamProcessorConfig_tags1(rName, tagKey1, tagValue1 string) string {
	return acctest.ConfigCompose(
		testAccStreamProcessorConfigBase_connectedHome(rName),
		fmt.Sprintf(`
resource "aws_rekognition_stream_processor" "test" {
  role_arn = aws_iam_role.test.arn
  name     = %[1]q

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

  tags = {
    %[2]q = %[3]q
  }
}
`, rName, tagKey1, tagValue1))
}

func testAccStreamProcessorConfig_tags2(rName, tagKey1, tagValue1, tagKey2, tagValue2 string) string {
	return acctest.ConfigCompose(
		testAccStreamProcessorConfigBase_connectedHome(rName),
		fmt.Sprintf(`
resource "aws_rekognition_stream_processor" "test" {
  role_arn = aws_iam_role.test.arn
  name     = %[1]q

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

  tags = {
    %[2]q = %[3]q
    %[4]q = %[5]q
  }
}
`, rName, tagKey1, tagValue1, tagKey2, tagValue2))
}
