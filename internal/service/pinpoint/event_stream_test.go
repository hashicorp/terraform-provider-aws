// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package pinpoint_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/pinpoint"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfpinpoint "github.com/hashicorp/terraform-provider-aws/internal/service/pinpoint"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccPinpointEventStream_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var stream pinpoint.EventStream
	resourceName := "aws_pinpoint_event_stream.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	rName2 := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheckApp(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.PinpointServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckEventStreamDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccEventStreamConfig_basic(rName, rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckEventStreamExists(ctx, resourceName, &stream),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrApplicationID, "aws_pinpoint_app.test", names.AttrApplicationID),
					resource.TestCheckResourceAttrPair(resourceName, "destination_stream_arn", "aws_kinesis_stream.test", names.AttrARN),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrRoleARN, "aws_iam_role.test", names.AttrARN),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccEventStreamConfig_basic(rName, rName2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckEventStreamExists(ctx, resourceName, &stream),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrApplicationID, "aws_pinpoint_app.test", names.AttrApplicationID),
					resource.TestCheckResourceAttrPair(resourceName, "destination_stream_arn", "aws_kinesis_stream.test", names.AttrARN),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrRoleARN, "aws_iam_role.test", names.AttrARN),
				),
			},
		},
	})
}

func TestAccPinpointEventStream_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	var stream pinpoint.EventStream
	resourceName := "aws_pinpoint_event_stream.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheckApp(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.PinpointServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckEventStreamDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccEventStreamConfig_basic(rName, rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckEventStreamExists(ctx, resourceName, &stream),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfpinpoint.ResourceEventStream(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckEventStreamExists(ctx context.Context, n string, stream *pinpoint.EventStream) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No Pinpoint event stream with that ID exists")
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).PinpointConn(ctx)

		// Check if the app exists
		params := &pinpoint.GetEventStreamInput{
			ApplicationId: aws.String(rs.Primary.ID),
		}
		output, err := conn.GetEventStreamWithContext(ctx, params)

		if err != nil {
			return err
		}

		*stream = *output.EventStream

		return nil
	}
}

func testAccEventStreamConfig_basic(rName, streamName string) string {
	return fmt.Sprintf(`
resource "aws_pinpoint_app" "test" {}

resource "aws_pinpoint_event_stream" "test" {
  application_id         = aws_pinpoint_app.test.application_id
  destination_stream_arn = aws_kinesis_stream.test.arn
  role_arn               = aws_iam_role.test.arn
}

resource "aws_kinesis_stream" "test" {
  name        = %[2]q
  shard_count = 1
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
        "Service": "pinpoint.amazonaws.com"
      },
      "Effect": "Allow",
      "Sid": ""
    }
  ]
}
EOF
}

resource "aws_iam_role_policy" "test" {
  name = %[1]q
  role = aws_iam_role.test.id

  policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": {
    "Action": [
      "kinesis:PutRecords",
      "kinesis:DescribeStream"
    ],
    "Effect": "Allow",
    "Resource": [
      "${aws_kinesis_stream.test.arn}"
    ]
  }
}
EOF
}
`, rName, streamName)
}

func testAccCheckEventStreamDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).PinpointConn(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_pinpoint_event_stream" {
				continue
			}

			// Check if the event stream exists
			params := &pinpoint.GetEventStreamInput{
				ApplicationId: aws.String(rs.Primary.ID),
			}
			_, err := conn.GetEventStreamWithContext(ctx, params)
			if err != nil {
				if tfawserr.ErrCodeEquals(err, pinpoint.ErrCodeNotFoundException) {
					continue
				}
				return err
			}
			return fmt.Errorf("Event stream exists when it should be destroyed!")
		}

		return nil
	}
}
