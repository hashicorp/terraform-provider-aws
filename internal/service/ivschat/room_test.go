// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package ivschat_test

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ivschat"
	"github.com/aws/aws-sdk-go-v2/service/ivschat/types"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	tfivschat "github.com/hashicorp/terraform-provider-aws/internal/service/ivschat"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccIVSChatRoom_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var room ivschat.GetRoomOutput
	resourceName := "aws_ivschat_room.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.IVSChatEndpointID)
			testAccPreCheckRoom(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.IVSChatEndpointID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckRoomDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccRoomConfig_basic(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRoomExists(ctx, resourceName, &room),
					resource.TestCheckResourceAttrSet(resourceName, "id"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
					resource.TestCheckResourceAttr(resourceName, "tags_all.%", "0"),
					acctest.MatchResourceAttrRegionalARN(resourceName, "arn", "ivschat", regexp.MustCompile(`room/.+`)),
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

func TestAccIVSChatRoom_tags(t *testing.T) {
	ctx := acctest.Context(t)
	var room ivschat.GetRoomOutput
	resourceName := "aws_ivschat_room.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.IVSChatEndpointID)
			testAccPreCheckRoom(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.IVSChatEndpointID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckRoomDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccRoomConfig_tags1("key1", "value1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRoomExists(ctx, resourceName, &room),
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
				Config: testAccRoomConfig_tags2("key1", "value1updated", "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRoomExists(ctx, resourceName, &room),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1updated"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
			{
				Config: testAccRoomConfig_tags1("key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRoomExists(ctx, resourceName, &room),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
		},
	})
}

func TestAccIVSChatRoom_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	var room ivschat.GetRoomOutput
	resourceName := "aws_ivschat_room.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.IVSChatEndpointID)
			testAccPreCheckRoom(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.IVSChatEndpointID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckRoomDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccRoomConfig_basic(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRoomExists(ctx, resourceName, &room),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfivschat.ResourceRoom(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccIVSChatRoom_update(t *testing.T) {
	ctx := acctest.Context(t)
	var room1, room2 ivschat.GetRoomOutput

	resourceName := "aws_ivschat_room.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	maximumMessageLength := "123"
	maximumMessageRatePerSecond := "5"
	fallbackResult := "DENY"
	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.IVSChatEndpointID)
			testAccPreCheckRoom(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.IVSChatEndpointID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckRoomDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccRoomConfig_basic(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRoomExists(ctx, resourceName, &room1),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccRoomConfig_update(rName, maximumMessageLength, maximumMessageRatePerSecond, fallbackResult),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRoomExists(ctx, resourceName, &room2),
					testAccCheckRoomNotRecreated(&room1, &room2),
					resource.TestCheckResourceAttr(resourceName, "maximum_message_length", maximumMessageLength),
					resource.TestCheckResourceAttr(resourceName, "maximum_message_rate_per_second", maximumMessageRatePerSecond),
					resource.TestCheckResourceAttr(resourceName, "message_review_handler.0.fallback_result", fallbackResult),
					acctest.CheckResourceAttrRegionalARN(resourceName, "message_review_handler.0.uri", "lambda", fmt.Sprintf("function:%s", rName)),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
				),
			},
		},
	})
}

func TestAccIVSChatRoom_loggingConfiguration(t *testing.T) {
	ctx := acctest.Context(t)
	var room1, room2, room3, room4 ivschat.GetRoomOutput

	resourceName := "aws_ivschat_room.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.IVSChatEndpointID)
			testAccPreCheckRoom(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.IVSChatEndpointID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckRoomDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccRoomConfig_loggingConfiguration1(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRoomExists(ctx, resourceName, &room1),
					resource.TestCheckResourceAttr(resourceName, "logging_configuration_identifiers.#", "1"),
					resource.TestCheckResourceAttrPair(resourceName, "logging_configuration_identifiers.0", "aws_ivschat_logging_configuration.test1", "arn"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccRoomConfig_loggingConfiguration2(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRoomExists(ctx, resourceName, &room2),
					testAccCheckRoomNotRecreated(&room1, &room2),
					resource.TestCheckResourceAttr(resourceName, "logging_configuration_identifiers.#", "2"),
					resource.TestCheckResourceAttrPair(resourceName, "logging_configuration_identifiers.0", "aws_ivschat_logging_configuration.test1", "arn"),
					resource.TestCheckResourceAttrPair(resourceName, "logging_configuration_identifiers.1", "aws_ivschat_logging_configuration.test2", "arn"),
				),
			},
			{
				Config: testAccRoomConfig_loggingConfiguration3(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRoomExists(ctx, resourceName, &room3),
					testAccCheckRoomNotRecreated(&room2, &room3),
					resource.TestCheckResourceAttr(resourceName, "logging_configuration_identifiers.#", "1"),
					resource.TestCheckResourceAttrPair(resourceName, "logging_configuration_identifiers.0", "aws_ivschat_logging_configuration.test3", "arn"),
				),
			},
			{
				Config: testAccRoomConfig_loggingConfiguration4(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRoomExists(ctx, resourceName, &room4),
					testAccCheckRoomNotRecreated(&room3, &room4),
					resource.TestCheckResourceAttr(resourceName, "logging_configuration_identifiers.#", "0"),
				),
			},
		},
	})
}

func TestAccIVSChatRoom_update_remove_messageReviewHandler_uri(t *testing.T) {
	ctx := acctest.Context(t)
	var room1, room2 ivschat.GetRoomOutput

	resourceName := "aws_ivschat_room.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.IVSChatEndpointID)
			testAccPreCheckRoom(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.IVSChatEndpointID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckRoomDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccRoomConfig_messageReviewHandler(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRoomExists(ctx, resourceName, &room1),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccRoomConfig_update_remove_messageReviewHandler_uri(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRoomExists(ctx, resourceName, &room2),
					testAccCheckRoomNotRecreated(&room1, &room2),
					resource.TestCheckResourceAttr(resourceName, "message_review_handler.0.uri", ""),
				),
			},
		},
	})
}

func testAccCheckRoomDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).IVSChatClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_ivschat_room" {
				continue
			}

			_, err := conn.GetRoom(ctx, &ivschat.GetRoomInput{
				Identifier: aws.String(rs.Primary.ID),
			})

			if err != nil {
				var nfe *types.ResourceNotFoundException
				if errors.As(err, &nfe) {
					return nil
				}
				return err
			}

			return create.Error(names.IVSChat, create.ErrActionCheckingDestroyed, tfivschat.ResNameRoom, rs.Primary.ID, errors.New("not destroyed"))
		}

		return nil
	}
}

func testAccCheckRoomExists(ctx context.Context, name string, room *ivschat.GetRoomOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return create.Error(names.IVSChat, create.ErrActionCheckingExistence, tfivschat.ResNameRoom, name, errors.New("not found"))
		}

		if rs.Primary.ID == "" {
			return create.Error(names.IVSChat, create.ErrActionCheckingExistence, tfivschat.ResNameRoom, name, errors.New("not set"))
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).IVSChatClient(ctx)

		resp, err := conn.GetRoom(ctx, &ivschat.GetRoomInput{
			Identifier: aws.String(rs.Primary.ID),
		})

		if err != nil {
			return create.Error(names.IVSChat, create.ErrActionCheckingExistence, tfivschat.ResNameRoom, rs.Primary.ID, err)
		}

		*room = *resp

		return nil
	}
}

func testAccPreCheckRoom(ctx context.Context, t *testing.T) {
	conn := acctest.Provider.Meta().(*conns.AWSClient).IVSChatClient(ctx)

	input := &ivschat.ListRoomsInput{}
	_, err := conn.ListRooms(ctx, input)

	if acctest.PreCheckSkipError(err) {
		t.Skipf("skipping acceptance testing: %s", err)
	}

	if err != nil {
		t.Fatalf("unexpected PreCheck error: %s", err)
	}
}

func testAccCheckRoomNotRecreated(before, after *ivschat.GetRoomOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if before, after := aws.ToString(before.Arn), aws.ToString(after.Arn); before != after {
			return create.Error(names.IVSChat, create.ErrActionCheckingNotRecreated, tfivschat.ResNameRoom, before, errors.New("recreated"))
		}

		return nil
	}
}

func testAccRoomConfig_basic() string {
	return `
resource "aws_ivschat_room" "test" {
}
`
}

func testAccRoomConfig_lambda(rName string) string {
	return fmt.Sprintf(`
resource "aws_iam_role" "test" {
  name               = %[1]q
  assume_role_policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [{
    "Effect": "Allow",
    "Action": ["sts:AssumeRole"],
    "Principal": {"Service": "lambda.amazonaws.com"}
  }]
}
EOF
}

resource "aws_lambda_function" "test" {
  filename         = "test-fixtures/lambdatest.zip"
  function_name    = %[1]q
  role             = aws_iam_role.test.arn
  source_code_hash = filebase64sha256("test-fixtures/lambdatest.zip")
  runtime          = "nodejs16.x"
  handler          = "index.handler"
}

resource "aws_lambda_permission" "test" {
  statement_id  = %[1]q
  action        = "lambda:InvokeFunction"
  function_name = aws_lambda_function.test.function_name
  principal     = "ivschat.amazonaws.com"
}
`, rName)
}

func testAccRoomConfig_update(rName, maximumMessageLength, maximumMessageRatePerSecond, fallbackResult string) string {
	return acctest.ConfigCompose(
		testAccRoomConfig_lambda(rName),
		fmt.Sprintf(`
resource "aws_ivschat_room" "test" {
  depends_on                      = [aws_lambda_permission.test]
  name                            = %[1]q
  maximum_message_length          = %[2]s
  maximum_message_rate_per_second = %[3]s
  message_review_handler {
    uri             = aws_lambda_function.test.arn
    fallback_result = %[4]q
  }
}
`, rName, maximumMessageLength, maximumMessageRatePerSecond, fallbackResult))
}

func testAccRoomConfig_messageReviewHandler(rName string) string {
	return acctest.ConfigCompose(
		testAccRoomConfig_lambda(rName),
		`
resource "aws_ivschat_room" "test" {
  depends_on = [aws_lambda_permission.test]
  message_review_handler {
    uri = aws_lambda_function.test.arn
  }
}
`)
}

func testAccRoomConfig_update_remove_messageReviewHandler_uri() string {
	return `
resource "aws_ivschat_room" "test" {
  message_review_handler {
    uri = ""
  }
}
`
}

func testAccRoomConfig_tags1(tagKey1, tagValue1 string) string {
	return fmt.Sprintf(`
resource "aws_ivschat_room" "test" {
  tags = {
    %[1]q = %[2]q
  }
}
`, tagKey1, tagValue1)
}

func testAccRoomConfig_tags2(tagKey1, tagValue1, tagKey2, tagValue2 string) string {
	return fmt.Sprintf(`
resource "aws_ivschat_room" "test" {
  tags = {
    %[1]q = %[2]q
    %[3]q = %[4]q
  }
}
`, tagKey1, tagValue1, tagKey2, tagValue2)
}

func testAccRoomConfig_loggingConfiguration_base(rName string) string {
	return fmt.Sprintf(`
resource "aws_s3_bucket" "test" {
  bucket        = %[1]q
  force_destroy = true
}

resource "aws_ivschat_logging_configuration" "test1" {
  destination_configuration {
    s3 {
      bucket_name = aws_s3_bucket.test.id
    }
  }
}

resource "aws_ivschat_logging_configuration" "test2" {
  destination_configuration {
    s3 {
      bucket_name = aws_s3_bucket.test.id
    }
  }
}

resource "aws_ivschat_logging_configuration" "test3" {
  destination_configuration {
    s3 {
      bucket_name = aws_s3_bucket.test.id
    }
  }
}
`, rName)
}

func testAccRoomConfig_loggingConfiguration1(rName string) string {
	return acctest.ConfigCompose(
		testAccRoomConfig_loggingConfiguration_base(rName),
		`
resource "aws_ivschat_room" "test" {
  logging_configuration_identifiers = [
    aws_ivschat_logging_configuration.test1.arn
  ]
}
`)
}

func testAccRoomConfig_loggingConfiguration2(rName string) string {
	return acctest.ConfigCompose(
		testAccRoomConfig_loggingConfiguration_base(rName),
		`
resource "aws_ivschat_room" "test" {
  logging_configuration_identifiers = [
    aws_ivschat_logging_configuration.test1.arn,
    aws_ivschat_logging_configuration.test2.arn,
  ]
}
`)
}

func testAccRoomConfig_loggingConfiguration3(rName string) string {
	return acctest.ConfigCompose(
		testAccRoomConfig_loggingConfiguration_base(rName),
		`
resource "aws_ivschat_room" "test" {
  logging_configuration_identifiers = [
    aws_ivschat_logging_configuration.test3.arn
  ]
}
`)
}

func testAccRoomConfig_loggingConfiguration4(rName string) string {
	return acctest.ConfigCompose(
		testAccRoomConfig_loggingConfiguration_base(rName),
		`
resource "aws_ivschat_room" "test" {
  logging_configuration_identifiers = []
}
`)
}
