package appflow_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/service/appflow"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfappflow "github.com/hashicorp/terraform-provider-aws/internal/service/appflow"
)

func TestAccAppFlowFlow_basic(t *testing.T) {
	var flowOutput appflow.FlowDefinition
	rSourceName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	rDestinationName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	rFlowName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_appflow_flow.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:   func() { acctest.PreCheck(t) },
		ErrorCheck: acctest.ErrorCheck(t, appflow.EndpointsID),
		Providers:  acctest.Providers,
		Steps: []resource.TestStep{
			{
				Config: testAccConfigFlow_basic(rSourceName, rDestinationName, rFlowName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFlowExists(resourceName, &flowOutput),
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

func testAccConfigFlowBase(rSourceName string, rDestinationName string) string {
	return fmt.Sprintf(`
resource "aws_s3_bucket" "test_source" {
  bucket = %[1]q
}

resource "aws_s3_bucket_policy" "test_source" {
  bucket = aws_s3_bucket.test_source.id
  policy = <<EOF
{
    "Statement": [
        {
            "Effect": "Allow",
            "Sid": "AllowAppFlowSourceActions",
            "Principal": {
                "Service": "appflow.amazonaws.com"
            },
            "Action": [
                "s3:ListBucket",
                "s3:GetObject"
            ],
            "Resource": [
                "arn:aws:s3:::%[1]s",
                "arn:aws:s3:::%[1]s/*"
            ]
        }
    ],
	"Version": "2012-10-17"
}
EOF
}

resource "aws_s3_object" "test" {
  bucket = aws_s3_bucket.test_source.id
  key    = "flow_source.csv"
  source = "test-fixtures/flow_source.csv"
}

resource "aws_s3_bucket" "test_destination" {
  bucket = %[2]q
}

resource "aws_s3_bucket_policy" "test_destination" {
  bucket = aws_s3_bucket.test_destination.id
  policy = <<EOF

{
    "Statement": [
        {
            "Effect": "Allow",
            "Sid": "AllowAppFlowDestinationActions",
            "Principal": {
                "Service": "appflow.amazonaws.com"
            },
            "Action": [
                "s3:PutObject",
                "s3:AbortMultipartUpload",
                "s3:ListMultipartUploadParts",
                "s3:ListBucketMultipartUploads",
                "s3:GetBucketAcl",
                "s3:PutObjectAcl"
            ],
            "Resource": [
                "arn:aws:s3:::%[2]s",
                "arn:aws:s3:::%[2]s/*"
            ]
        }
    ],
	"Version": "2012-10-17"
}
EOF
}
`, rSourceName, rDestinationName)
}

func testAccConfigFlow_basic(rSourceName string, rDestinationName string, rFlowName string) string {
	return acctest.ConfigCompose(
		testAccConfigFlowBase(rSourceName, rDestinationName),
		fmt.Sprintf(`
resource "aws_appflow_flow" "test" {
  name = %[3]q

  source_flow_config {
    connector_type = "S3"
    source_connector_properties {
      s3 {
        bucket_name = aws_s3_bucket_policy.test_source.bucket
		bucket_prefix = "flow"
      }
    }
  }

  destination_flow_config {
    connector_type = "S3"
    destination_connector_properties {
      s3 {
        bucket_name = aws_s3_bucket_policy.test_destination.bucket

		s3_output_format_config {
		  prefix_config {
		    prefix_type = "PATH"
		  }
		}
      }
    }
  }

  task {
    source_fields = ["testField"]
	destination_field = "testField"
    task_type     = "Map"

	connector_operator {
	  s3 = "NO_OP"
	}
  }

  trigger_config {
    trigger_type = "OnDemand"
  }
}
`, rSourceName, rDestinationName, rFlowName),
	)
}

func testAccCheckFlowExists(resourceName string, flow *appflow.FlowDefinition) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("not found: %s", resourceName)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).AppFlowConn
		resp, err := tfappflow.FindFlowByArn(context.Background(), conn, rs.Primary.ID)

		if err != nil {
			return err
		}

		if resp == nil {
			return fmt.Errorf("AppFlow Flow %q does not exist", rs.Primary.ID)
		}

		*flow = *resp

		return nil
	}
}
