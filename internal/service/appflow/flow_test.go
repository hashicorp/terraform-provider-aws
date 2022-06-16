package appflow_test

import (
	"context"
	"fmt"
	"regexp"
	"testing"

	"github.com/aws/aws-sdk-go/service/appflow"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfappflow "github.com/hashicorp/terraform-provider-aws/internal/service/appflow"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

func TestAccAppFlowFlow_basic(t *testing.T) {
	var flowOutput appflow.FlowDefinition
	rSourceName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	rDestinationName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	rFlowName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_appflow_flow.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, appflow.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckFlowDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccFlowConfig_basic(rSourceName, rDestinationName, rFlowName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFlowExists(resourceName, &flowOutput),
					acctest.MatchResourceAttrRegionalARN(resourceName, "arn", "appflow", regexp.MustCompile(`flow/.+`)),
					resource.TestCheckResourceAttr(resourceName, "name", rFlowName),
					resource.TestCheckResourceAttrSet(resourceName, "destination_flow_config.#"),
					resource.TestCheckResourceAttrSet(resourceName, "destination_flow_config.0.connector_type"),
					resource.TestCheckResourceAttrSet(resourceName, "destination_flow_config.0.destination_connector_properties.#"),
					resource.TestCheckResourceAttrSet(resourceName, "source_flow_config.#"),
					resource.TestCheckResourceAttrSet(resourceName, "source_flow_config.0.connector_type"),
					resource.TestCheckResourceAttrSet(resourceName, "source_flow_config.0.source_connector_properties.#"),
					resource.TestCheckResourceAttrSet(resourceName, "task.#"),
					resource.TestCheckResourceAttrSet(resourceName, "task.0.source_fields.#"),
					resource.TestCheckResourceAttrSet(resourceName, "task.0.task_type"),
					resource.TestCheckResourceAttrSet(resourceName, "trigger_config.#"),
					resource.TestCheckResourceAttrSet(resourceName, "trigger_config.0.trigger_type"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
					resource.TestCheckResourceAttr(resourceName, "tags_all.%", "0"),
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

func TestAccAppFlowFlow_update(t *testing.T) {
	var flowOutput appflow.FlowDefinition
	rSourceName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	rDestinationName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	rFlowName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_appflow_flow.test"
	description := "test description"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, appflow.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckFlowDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccFlowConfig_basic(rSourceName, rDestinationName, rFlowName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFlowExists(resourceName, &flowOutput),
				),
			},
			{
				Config: testAccFlowConfig_update(rSourceName, rDestinationName, rFlowName, description),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFlowExists(resourceName, &flowOutput),
					resource.TestCheckResourceAttr(resourceName, "description", description),
				),
			},
		},
	})
}

func TestAccAppFlowFlow_TaskProperties(t *testing.T) {
	var flowOutput appflow.FlowDefinition
	rSourceName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	rDestinationName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	rFlowName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_appflow_flow.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, appflow.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckFlowDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccFlowConfig_taskProperties(rSourceName, rDestinationName, rFlowName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFlowExists(resourceName, &flowOutput),
					resource.TestCheckResourceAttr(resourceName, "task.0.task_properties.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "task.0.task_properties.SOURCE_DATA_TYPE", "CSV"),
					resource.TestCheckResourceAttr(resourceName, "task.0.task_properties.DESTINATION_DATA_TYPE", "CSV"),
				),
			},
		},
	})
}

func TestAccAppFlowFlow_tags(t *testing.T) {
	var flowOutput appflow.FlowDefinition
	rSourceName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	rDestinationName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	rFlowName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_appflow_flow.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, appflow.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckFlowDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccFlowConfig_tags1(rSourceName, rDestinationName, rFlowName, "key1", "value1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFlowExists(resourceName, &flowOutput),
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
				Config: testAccFlowConfig_tags2(rSourceName, rDestinationName, rFlowName, "key1", "value1updated", "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFlowExists(resourceName, &flowOutput),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1updated"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
			{
				Config: testAccFlowConfig_tags1(rSourceName, rDestinationName, rFlowName, "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFlowExists(resourceName, &flowOutput),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
		},
	})
}

func TestAccAppFlowFlow_disappears(t *testing.T) {
	var flowOutput appflow.FlowDefinition
	rSourceName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	rDestinationName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	rFlowName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_appflow_flow.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, appflow.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckFlowDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccFlowConfig_basic(rSourceName, rDestinationName, rFlowName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFlowExists(resourceName, &flowOutput),
					acctest.CheckResourceDisappears(acctest.Provider, tfappflow.ResourceFlow(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccConfigFlowBase(rSourceName string, rDestinationName string) string {
	return fmt.Sprintf(`
data "aws_partition" "current" {}

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
                "arn:${data.aws_partition.current.partition}:s3:::%[1]s",
                "arn:${data.aws_partition.current.partition}:s3:::%[1]s/*"
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
                "arn:${data.aws_partition.current.partition}:s3:::%[2]s",
                "arn:${data.aws_partition.current.partition}:s3:::%[2]s/*"
            ]
        }
    ],
	"Version": "2012-10-17"
}
EOF
}
`, rSourceName, rDestinationName)
}

func testAccFlowConfig_basic(rSourceName string, rDestinationName string, rFlowName string) string {
	return acctest.ConfigCompose(
		testAccConfigFlowBase(rSourceName, rDestinationName),
		fmt.Sprintf(`
resource "aws_appflow_flow" "test" {
  name = %[3]q

  source_flow_config {
    connector_type = "S3"
    source_connector_properties {
      s3 {
        bucket_name   = aws_s3_bucket_policy.test_source.bucket
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
    source_fields     = ["testField"]
    destination_field = "testField"
    task_type         = "Map"

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

func testAccFlowConfig_update(rSourceName string, rDestinationName string, rFlowName string, description string) string {
	return acctest.ConfigCompose(
		testAccConfigFlowBase(rSourceName, rDestinationName),
		fmt.Sprintf(`
resource "aws_appflow_flow" "test" {
  name        = %[3]q
  description = %[4]q

  source_flow_config {
    connector_type = "S3"
    source_connector_properties {
      s3 {
        bucket_name   = aws_s3_bucket_policy.test_source.bucket
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
    source_fields     = ["testField"]
    destination_field = "testField"
    task_type         = "Map"

    connector_operator {
      s3 = "NO_OP"
    }
  }

  trigger_config {
    trigger_type = "OnDemand"
  }
}
`, rSourceName, rDestinationName, rFlowName, description),
	)
}

func testAccFlowConfig_taskProperties(rSourceName string, rDestinationName string, rFlowName string) string {
	return acctest.ConfigCompose(
		testAccConfigFlowBase(rSourceName, rDestinationName),
		fmt.Sprintf(`
resource "aws_appflow_flow" "test" {
  name = %[3]q

  source_flow_config {
    connector_type = "S3"
    source_connector_properties {
      s3 {
        bucket_name   = aws_s3_bucket_policy.test_source.bucket
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
    source_fields     = ["testField"]
    destination_field = "testField"
    task_type         = "Map"

    task_properties = {
      SOURCE_DATA_TYPE      = "CSV"
      DESTINATION_DATA_TYPE = "CSV"
    }

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

func testAccFlowConfig_tags1(rSourceName string, rDestinationName string, rFlowName string, tagKey1 string, tagValue1 string) string {
	return acctest.ConfigCompose(
		testAccConfigFlowBase(rSourceName, rDestinationName),
		fmt.Sprintf(`
resource "aws_appflow_flow" "test" {
  name = %[3]q

  source_flow_config {
    connector_type = "S3"
    source_connector_properties {
      s3 {
        bucket_name   = aws_s3_bucket_policy.test_source.bucket
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
    source_fields     = ["testField"]
    destination_field = "testField"
    task_type         = "Map"

    connector_operator {
      s3 = "NO_OP"
    }
  }

  trigger_config {
    trigger_type = "OnDemand"
  }

  tags = {
    %[4]q = %[5]q
  }
}
`, rSourceName, rDestinationName, rFlowName, tagKey1, tagValue1),
	)
}

func testAccFlowConfig_tags2(rSourceName string, rDestinationName string, rFlowName string, tagKey1 string, tagValue1 string, tagKey2 string, tagValue2 string) string {
	return acctest.ConfigCompose(
		testAccConfigFlowBase(rSourceName, rDestinationName),
		fmt.Sprintf(`
resource "aws_appflow_flow" "test" {
  name = %[3]q

  source_flow_config {
    connector_type = "S3"
    source_connector_properties {
      s3 {
        bucket_name   = aws_s3_bucket_policy.test_source.bucket
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
    source_fields     = ["testField"]
    destination_field = "testField"
    task_type         = "Map"

    connector_operator {
      s3 = "NO_OP"
    }
  }

  trigger_config {
    trigger_type = "OnDemand"
  }

  tags = {
    %[4]q = %[5]q
    %[6]q = %[7]q
  }
}
`, rSourceName, rDestinationName, rFlowName, tagKey1, tagValue1, tagKey2, tagValue2),
	)
}

func testAccCheckFlowExists(resourceName string, flow *appflow.FlowDefinition) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("not found: %s", resourceName)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).AppFlowConn
		resp, err := tfappflow.FindFlowByARN(context.Background(), conn, rs.Primary.ID)

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

func testAccCheckFlowDestroy(s *terraform.State) error {
	conn := acctest.Provider.Meta().(*conns.AWSClient).AppFlowConn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_appflow_flow" {
			continue
		}

		_, err := tfappflow.FindFlowByARN(context.Background(), conn, rs.Primary.ID)

		if tfresource.NotFound(err) {
			continue
		}

		if err != nil {
			return err
		}

		return fmt.Errorf("Expected AppFlow Flow to be destroyed, %s found", rs.Primary.ID)
	}

	return nil
}
