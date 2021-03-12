package aws

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/service/kinesisanalyticsv2"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/service/kinesisanalyticsv2/finder"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/tfresource"
)

func TestAccAWSKinesisAnalyticsV2ApplicationSnapshot_basic(t *testing.T) {
	var v kinesisanalyticsv2.SnapshotDetails
	resourceName := "aws_kinesisanalyticsv2_application_snapshot.test"
	applicationResourceName := "aws_kinesisanalyticsv2_application.test"
	rName := acctest.RandomWithPrefix("tf-acc-test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSKinesisAnalyticsV2(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckKinesisAnalyticsV2ApplicationSnapshotDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccKinesisAnalyticsV2ApplicationSnapshotConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckKinesisAnalyticsV2ApplicationSnapshotExists(resourceName, &v),
					resource.TestCheckResourceAttrPair(resourceName, "application_name", applicationResourceName, "name"),
					resource.TestCheckResourceAttr(resourceName, "application_version_id", "1"),
					resource.TestCheckResourceAttrSet(resourceName, "snapshot_creation_timestamp"),
					resource.TestCheckResourceAttr(resourceName, "snapshot_name", rName),
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

func TestAccAWSKinesisAnalyticsV2ApplicationSnapshot_disappears(t *testing.T) {
	var v kinesisanalyticsv2.SnapshotDetails
	resourceName := "aws_kinesisanalyticsv2_application_snapshot.test"
	applicationResourceName := "aws_kinesisanalyticsv2_application.test"
	rName := acctest.RandomWithPrefix("tf-acc-test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSKinesisAnalyticsV2(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckKinesisAnalyticsV2ApplicationSnapshotDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccKinesisAnalyticsV2ApplicationSnapshotConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckKinesisAnalyticsV2ApplicationSnapshotExists(resourceName, &v),
					testAccCheckResourceDisappears(testAccProvider, resourceAwsKinesisAnalyticsV2ApplicationSnapshot(), applicationResourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccAWSKinesisAnalyticsV2ApplicationSnapshot_disappears_Application(t *testing.T) {
	var v kinesisanalyticsv2.SnapshotDetails
	resourceName := "aws_kinesisanalyticsv2_application_snapshot.test"
	rName := acctest.RandomWithPrefix("tf-acc-test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSKinesisAnalyticsV2(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckKinesisAnalyticsV2ApplicationSnapshotDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccKinesisAnalyticsV2ApplicationSnapshotConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckKinesisAnalyticsV2ApplicationSnapshotExists(resourceName, &v),
					testAccCheckResourceDisappears(testAccProvider, resourceAwsKinesisAnalyticsV2Application(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckKinesisAnalyticsV2ApplicationSnapshotDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*AWSClient).kinesisanalyticsv2conn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_kinesisanalyticsv2_application_snapshot" {
			continue
		}

		_, err := finder.SnapshotDetailsByApplicationAndSnapshotNames(conn, rs.Primary.Attributes["application_name"], rs.Primary.Attributes["snapshot_name"])

		if tfresource.NotFound(err) {
			continue
		}

		if err != nil {
			return err
		}

		return fmt.Errorf("Kinesis Analytics v2 Application Snapshot %s still exists", rs.Primary.ID)
	}
	return nil
}

func testAccCheckKinesisAnalyticsV2ApplicationSnapshotExists(n string, v *kinesisanalyticsv2.SnapshotDetails) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No Kinesis Analytics v2 Application Snapshot ID is set")
		}

		conn := testAccProvider.Meta().(*AWSClient).kinesisanalyticsv2conn

		application, err := finder.SnapshotDetailsByApplicationAndSnapshotNames(conn, rs.Primary.Attributes["application_name"], rs.Primary.Attributes["snapshot_name"])

		if err != nil {
			return err
		}

		*v = *application

		return nil
	}
}

func testAccKinesisAnalyticsV2ApplicationSnapshotConfig(rName string) string {
	return composeConfig(
		testAccKinesisAnalyticsV2ApplicationConfigBaseServiceExecutionIamRole(rName),
		fmt.Sprintf(`
data "aws_region" "current" {}

resource "aws_s3_bucket" "test" {
  bucket = %[1]q
}

resource "aws_s3_bucket_object" "test" {
  bucket = aws_s3_bucket.test.bucket
  key    = "aws-kinesis-analytics-java-apps-1.0.jar"
  source = "test-fixtures/aws-kinesis-analytics-java-apps-1.0.jar"
}

# The stream names are hard-coded into the application.
resource "aws_kinesis_stream" "input" {
  name        = "ExampleInputStream"
  shard_count = 1
}

resource "aws_kinesis_stream" "output" {
  name        = "ExampleOutputStream"
  shard_count = 1
}

resource "aws_kinesisanalyticsv2_application" "test" {
  name                   = %[1]q
  runtime_environment    = "FLINK-1_11"
  service_execution_role = aws_iam_role.test[0].arn

  application_configuration {
    application_code_configuration {
      code_content {
        s3_content_location {
          bucket_arn = aws_s3_bucket.test.arn
          file_key   = aws_s3_bucket_object.test.key
        }
      }

      code_content_type = "ZIPFILE"
    }

    application_snapshot_configuration {
      snapshots_enabled = true
    }

    environment_properties {
      property_group {
        property_group_id = "ProducerConfigProperties"

        property_map = {
          "flink.inputstream.initpos" = "LATEST"
          "aws.region"                = data.aws_region.current.name
          "AggregationEnabled"        = "false"
        }
      }
    }
  }

  start_application = true

  depends_on = [aws_kinesis_stream.input, aws_kinesis_stream.output]
}

resource "aws_kinesisanalyticsv2_application_snapshot" "test" {
  application_name = aws_kinesisanalyticsv2_application.test.name
  snapshot_name    = %[1]q
}
`, rName))
}
