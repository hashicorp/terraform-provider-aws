package pinpoint_test

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/pinpoint"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfpinpoint "github.com/hashicorp/terraform-provider-aws/internal/service/pinpoint"
)

func TestAccPinpointEventStream_basic(t *testing.T) {
	var stream pinpoint.EventStream
	resourceName := "aws_pinpoint_event_stream.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	rName2 := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheckApp(t) },
		ErrorCheck:        acctest.ErrorCheck(t, pinpoint.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckEventStreamDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccEventStreamConfig_basic(rName, rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckEventStreamExists(resourceName, &stream),
					resource.TestCheckResourceAttrPair(resourceName, "application_id", "aws_pinpoint_app.test", "application_id"),
					resource.TestCheckResourceAttrPair(resourceName, "destination_stream_arn", "aws_kinesis_stream.test", "arn"),
					resource.TestCheckResourceAttrPair(resourceName, "role_arn", "aws_iam_role.test", "arn"),
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
					testAccCheckEventStreamExists(resourceName, &stream),
					resource.TestCheckResourceAttrPair(resourceName, "application_id", "aws_pinpoint_app.test", "application_id"),
					resource.TestCheckResourceAttrPair(resourceName, "destination_stream_arn", "aws_kinesis_stream.test", "arn"),
					resource.TestCheckResourceAttrPair(resourceName, "role_arn", "aws_iam_role.test", "arn"),
				),
			},
		},
	})
}

func TestAccPinpointEventStream_disappears(t *testing.T) {
	var stream pinpoint.EventStream
	resourceName := "aws_pinpoint_event_stream.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheckApp(t) },
		ErrorCheck:        acctest.ErrorCheck(t, pinpoint.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckEventStreamDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccEventStreamConfig_basic(rName, rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckEventStreamExists(resourceName, &stream),
					acctest.CheckResourceDisappears(acctest.Provider, tfpinpoint.ResourceEventStream(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckEventStreamExists(n string, stream *pinpoint.EventStream) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No Pinpoint event stream with that ID exists")
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).PinpointConn

		// Check if the app exists
		params := &pinpoint.GetEventStreamInput{
			ApplicationId: aws.String(rs.Primary.ID),
		}
		output, err := conn.GetEventStream(params)

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

func testAccCheckEventStreamDestroy(s *terraform.State) error {
	conn := acctest.Provider.Meta().(*conns.AWSClient).PinpointConn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_pinpoint_event_stream" {
			continue
		}

		// Check if the event stream exists
		params := &pinpoint.GetEventStreamInput{
			ApplicationId: aws.String(rs.Primary.ID),
		}
		_, err := conn.GetEventStream(params)
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
