package kinesis_test

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/kinesis"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfkinesis "github.com/hashicorp/terraform-provider-aws/internal/service/kinesis"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

func TestAccKinesisStream_basic(t *testing.T) {
	var stream kinesis.StreamDescriptionSummary
	resourceName := "aws_kinesis_stream.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, kinesis.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckStreamDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccStreamConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckStreamExists(resourceName, &stream),
					acctest.CheckResourceAttrRegionalARN(resourceName, "arn", "kinesis", fmt.Sprintf("stream/%s", rName)),
					resource.TestCheckResourceAttr(resourceName, "encryption_type", "NONE"),
					resource.TestCheckResourceAttr(resourceName, "enforce_consumer_deletion", "false"),
					resource.TestCheckResourceAttr(resourceName, "kms_key_id", ""),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "retention_period", "24"),
					resource.TestCheckResourceAttr(resourceName, "shard_count", "2"),
					resource.TestCheckResourceAttr(resourceName, "shard_level_metrics.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "stream_mode_details.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "stream_mode_details.0.stream_mode", "PROVISIONED"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateId:           rName,
				ImportStateVerifyIgnore: []string{"enforce_consumer_deletion"},
			},
		},
	})
}

func TestAccKinesisStream_disappears(t *testing.T) {
	var stream kinesis.StreamDescriptionSummary
	resourceName := "aws_kinesis_stream.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, kinesis.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckStreamDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccStreamConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckStreamExists(resourceName, &stream),
					acctest.CheckResourceDisappears(acctest.Provider, tfkinesis.ResourceStream(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccKinesisStream_createMultipleConcurrentStreams(t *testing.T) {
	var stream kinesis.StreamDescriptionSummary
	resourceName := "aws_kinesis_stream.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, kinesis.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckStreamDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccStreamConfig_concurrent(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckStreamExists("aws_kinesis_stream.test.0", &stream),
					testAccCheckStreamExists("aws_kinesis_stream.test.1", &stream),
					testAccCheckStreamExists("aws_kinesis_stream.test.2", &stream),
					testAccCheckStreamExists("aws_kinesis_stream.test.3", &stream),
					testAccCheckStreamExists("aws_kinesis_stream.test.4", &stream),
					testAccCheckStreamExists("aws_kinesis_stream.test.5", &stream),
					testAccCheckStreamExists("aws_kinesis_stream.test.6", &stream),
					testAccCheckStreamExists("aws_kinesis_stream.test.7", &stream),
					testAccCheckStreamExists("aws_kinesis_stream.test.8", &stream),
					testAccCheckStreamExists("aws_kinesis_stream.test.9", &stream),
					testAccCheckStreamExists("aws_kinesis_stream.test.10", &stream),
					testAccCheckStreamExists("aws_kinesis_stream.test.11", &stream),
					testAccCheckStreamExists("aws_kinesis_stream.test.12", &stream),
					testAccCheckStreamExists("aws_kinesis_stream.test.13", &stream),
					testAccCheckStreamExists("aws_kinesis_stream.test.14", &stream),
					testAccCheckStreamExists("aws_kinesis_stream.test.15", &stream),
					testAccCheckStreamExists("aws_kinesis_stream.test.16", &stream),
					testAccCheckStreamExists("aws_kinesis_stream.test.17", &stream),
					testAccCheckStreamExists("aws_kinesis_stream.test.18", &stream),
					testAccCheckStreamExists("aws_kinesis_stream.test.19", &stream),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateId:           rName + "-0",
				ImportStateVerifyIgnore: []string{"enforce_consumer_deletion"},
			},
		},
	})
}

func TestAccKinesisStream_encryptionWithoutKMSKeyThrowsError(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, kinesis.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckStreamDestroy,
		Steps: []resource.TestStep{
			{
				Config:      testAccStreamConfig_encryptionAndNoKMSKey(rName),
				ExpectError: regexp.MustCompile("KMS Key Id required when setting encryption_type is not set as NONE"),
			},
		},
	})
}

func TestAccKinesisStream_encryption(t *testing.T) {
	var stream kinesis.StreamDescriptionSummary
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_kinesis_stream.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, kinesis.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckStreamDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccStreamConfig_encryption(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckStreamExists(resourceName, &stream),
					resource.TestCheckResourceAttr(resourceName, "encryption_type", "KMS"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateId:           rName,
				ImportStateVerifyIgnore: []string{"enforce_consumer_deletion"},
			},
			{
				Config: testAccStreamConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckStreamExists(resourceName, &stream),
					resource.TestCheckResourceAttr(resourceName, "encryption_type", "NONE"),
				),
			},
			{
				Config: testAccStreamConfig_encryption(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckStreamExists(resourceName, &stream),
					resource.TestCheckResourceAttr(resourceName, "encryption_type", "KMS"),
				),
			},
		},
	})
}

func TestAccKinesisStream_shardCount(t *testing.T) {
	var stream kinesis.StreamDescriptionSummary
	var updatedStream kinesis.StreamDescriptionSummary

	testCheckStreamNotDestroyed := func() resource.TestCheckFunc {
		return func(*terraform.State) error {
			if *stream.StreamCreationTimestamp != *updatedStream.StreamCreationTimestamp {
				return fmt.Errorf("Creation timestamps dont match, stream was recreated")
			}
			return nil
		}
	}

	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_kinesis_stream.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, kinesis.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckStreamDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccStreamConfig_shardCount(rName, 128),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckStreamExists(resourceName, &stream),
					resource.TestCheckResourceAttr(resourceName, "shard_count", "128"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateId:           rName,
				ImportStateVerifyIgnore: []string{"enforce_consumer_deletion"},
			},
			{
				Config: testAccStreamConfig_shardCount(rName, 96),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckStreamExists(resourceName, &updatedStream),
					testCheckStreamNotDestroyed(),
					resource.TestCheckResourceAttr(resourceName, "shard_count", "96"),
				),
			},
		},
	})
}

func TestAccKinesisStream_retentionPeriod(t *testing.T) {
	var stream kinesis.StreamDescriptionSummary
	resourceName := "aws_kinesis_stream.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, kinesis.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckStreamDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccStreamConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckStreamExists(resourceName, &stream),
					resource.TestCheckResourceAttr(resourceName, "retention_period", "24"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateId:           rName,
				ImportStateVerifyIgnore: []string{"enforce_consumer_deletion"},
			},
			{
				Config: testAccStreamConfig_updateRetentionPeriod(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckStreamExists(resourceName, &stream),
					resource.TestCheckResourceAttr(resourceName, "retention_period", "8760"),
				),
			},

			{
				Config: testAccStreamConfig_decreaseRetentionPeriod(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckStreamExists(resourceName, &stream),
					resource.TestCheckResourceAttr(resourceName, "retention_period", "28"),
				),
			},
		},
	})
}

func TestAccKinesisStream_shardLevelMetrics(t *testing.T) {
	var stream kinesis.StreamDescriptionSummary
	resourceName := "aws_kinesis_stream.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, kinesis.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckStreamDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccStreamConfig_singleShardLevelMetric(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckStreamExists(resourceName, &stream),
					resource.TestCheckResourceAttr(resourceName, "shard_level_metrics.#", "1"),
					resource.TestCheckTypeSetElemAttr(resourceName, "shard_level_metrics.*", "IncomingBytes"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateId:           rName,
				ImportStateVerifyIgnore: []string{"enforce_consumer_deletion"},
			},
			{
				Config: testAccStreamConfig_allShardLevelMetrics(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckStreamExists(resourceName, &stream),
					resource.TestCheckResourceAttr(resourceName, "shard_level_metrics.#", "7"),
					resource.TestCheckTypeSetElemAttr(resourceName, "shard_level_metrics.*", "IncomingBytes"),
					resource.TestCheckTypeSetElemAttr(resourceName, "shard_level_metrics.*", "IncomingRecords"),
					resource.TestCheckTypeSetElemAttr(resourceName, "shard_level_metrics.*", "OutgoingBytes"),
					resource.TestCheckTypeSetElemAttr(resourceName, "shard_level_metrics.*", "OutgoingRecords"),
					resource.TestCheckTypeSetElemAttr(resourceName, "shard_level_metrics.*", "WriteProvisionedThroughputExceeded"),
					resource.TestCheckTypeSetElemAttr(resourceName, "shard_level_metrics.*", "ReadProvisionedThroughputExceeded"),
					resource.TestCheckTypeSetElemAttr(resourceName, "shard_level_metrics.*", "IteratorAgeMilliseconds"),
				),
			},
			{
				Config: testAccStreamConfig_singleShardLevelMetric(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckStreamExists(resourceName, &stream),
					resource.TestCheckResourceAttr(resourceName, "shard_level_metrics.#", "1"),
					resource.TestCheckTypeSetElemAttr(resourceName, "shard_level_metrics.*", "IncomingBytes"),
				),
			},
		},
	})
}

func TestAccKinesisStream_enforceConsumerDeletion(t *testing.T) {
	var stream kinesis.StreamDescriptionSummary
	resourceName := "aws_kinesis_stream.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, kinesis.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckStreamDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccStreamConfig_enforceConsumerDeletion(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckStreamExists(resourceName, &stream),
					testAccStreamRegisterStreamConsumer(&stream, rName),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateId:           rName,
				ImportStateVerifyIgnore: []string{"enforce_consumer_deletion"},
			},
		},
	})
}

func TestAccKinesisStream_tags(t *testing.T) {
	var stream kinesis.StreamDescriptionSummary
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_kinesis_stream.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, kinesis.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckStreamDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccStreamConfig_tags1(rName, "key1", "value1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckStreamExists(resourceName, &stream),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateId:           rName,
				ImportStateVerifyIgnore: []string{"enforce_consumer_deletion"},
			},
			{
				Config: testAccStreamConfig_tags2(rName, "key1", "value1updated", "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckStreamExists(resourceName, &stream),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1updated"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
			{
				Config: testAccStreamConfig_tags1(rName, "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckStreamExists(resourceName, &stream),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
		},
	})
}

func TestAccKinesisStream_updateKMSKeyID(t *testing.T) {
	var stream kinesis.StreamDescriptionSummary
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_kinesis_stream.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, kinesis.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckStreamDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccStreamConfig_updateKMSKeyID(rName, 0),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckStreamExists(resourceName, &stream),
					resource.TestCheckResourceAttrPair(resourceName, "kms_key_id", "aws_kms_key.key.0", "id"),
				),
			},
			{
				Config: testAccStreamConfig_updateKMSKeyID(rName, 1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckStreamExists(resourceName, &stream),
					resource.TestCheckResourceAttrPair(resourceName, "kms_key_id", "aws_kms_key.key.1", "id"),
				),
			},
		},
	})
}

func TestAccKinesisStream_basicOnDemand(t *testing.T) {
	var stream kinesis.StreamDescriptionSummary
	resourceName := "aws_kinesis_stream.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, kinesis.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckStreamDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccStreamConfig_basicOnDemand(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckStreamExists(resourceName, &stream),
					resource.TestCheckResourceAttr(resourceName, "shard_count", "0"),
					resource.TestCheckResourceAttr(resourceName, "stream_mode_details.0.stream_mode", "ON_DEMAND"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateId:           rName,
				ImportStateVerifyIgnore: []string{"enforce_consumer_deletion"},
			},
		},
	})
}

func TestAccKinesisStream_switchBetweenProvisionedAndOnDemand(t *testing.T) {
	var stream kinesis.StreamDescriptionSummary
	resourceName := "aws_kinesis_stream.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, kinesis.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckStreamDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccStreamConfig_changeProvisionedToOnDemand1(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckStreamExists(resourceName, &stream),
					resource.TestCheckResourceAttr(resourceName, "shard_count", "1"),
					resource.TestCheckResourceAttr(resourceName, "stream_mode_details.0.stream_mode", "PROVISIONED"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateId:           rName,
				ImportStateVerifyIgnore: []string{"enforce_consumer_deletion"},
			},
			{
				Config: testAccStreamConfig_changeProvisionedToOnDemand2(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckStreamExists(resourceName, &stream),
					resource.TestCheckResourceAttr(resourceName, "shard_count", "0"),
					resource.TestCheckResourceAttr(resourceName, "stream_mode_details.0.stream_mode", "ON_DEMAND"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateId:           rName,
				ImportStateVerifyIgnore: []string{"enforce_consumer_deletion"},
			},
			{
				Config: testAccStreamConfig_changeProvisionedToOnDemand3(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckStreamExists(resourceName, &stream),
					resource.TestCheckResourceAttr(resourceName, "shard_count", "2"),
					resource.TestCheckResourceAttr(resourceName, "stream_mode_details.0.stream_mode", "PROVISIONED"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateId:           rName,
				ImportStateVerifyIgnore: []string{"enforce_consumer_deletion"},
			},
			{
				Config: testAccStreamConfig_changeProvisionedToOnDemand1(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckStreamExists(resourceName, &stream),
					resource.TestCheckResourceAttr(resourceName, "shard_count", "1"),
					resource.TestCheckResourceAttr(resourceName, "stream_mode_details.0.stream_mode", "PROVISIONED"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateId:           rName,
				ImportStateVerifyIgnore: []string{"enforce_consumer_deletion"},
			},
		},
	})
}

func TestAccKinesisStream_failOnBadStreamCountAndStreamModeCombination(t *testing.T) {
	var stream kinesis.StreamDescriptionSummary
	resourceName := "aws_kinesis_stream.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, kinesis.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckStreamDestroy,
		Steps: []resource.TestStep{
			// Check that we can't create an invalid combination
			{
				Config:      testAccStreamConfig_failOnBadCountAndModeCombinationNothingSet(rName),
				ExpectError: regexp.MustCompile(`shard_count must be at least 1 when stream_mode is PROVISIONED`),
			},
			// Check that we can't create an invalid combination
			{
				Config:      testAccStreamConfig_failOnBadCountAndModeCombinationShardCountWhenOnDemand(rName),
				ExpectError: regexp.MustCompile(`shard_count must not be set when stream_mode is ON_DEMAND`),
			},
			// Prepare for updates...
			{
				Config: testAccStreamConfig_failOnBadCountAndModeCombination(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckStreamExists(resourceName, &stream),
					resource.TestCheckResourceAttr(resourceName, "shard_count", "1"),
					resource.TestCheckResourceAttr(resourceName, "stream_mode_details.0.stream_mode", "PROVISIONED"),
				),
			},
			// Check that we can't update to an invalid combination
			{
				Config:      testAccStreamConfig_failOnBadCountAndModeCombinationNothingSet(rName),
				ExpectError: regexp.MustCompile(`shard_count must be at least 1 when stream_mode is PROVISIONED`),
			},
			// Check that we can't update to an invalid combination
			{
				Config:      testAccStreamConfig_failOnBadCountAndModeCombinationShardCountWhenOnDemand(rName),
				ExpectError: regexp.MustCompile(`shard_count must not be set when stream_mode is ON_DEMAND`),
			},
		},
	})
}

func testAccCheckStreamExists(n string, v *kinesis.StreamDescriptionSummary) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).KinesisConn

		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No Kinesis Stream ID is set")
		}

		output, err := tfkinesis.FindStreamByName(conn, rs.Primary.Attributes["name"])

		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

func testAccCheckStreamDestroy(s *terraform.State) error {
	conn := acctest.Provider.Meta().(*conns.AWSClient).KinesisConn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_kinesis_stream" {
			continue
		}

		_, err := tfkinesis.FindStreamByName(conn, rs.Primary.Attributes["name"])

		if tfresource.NotFound(err) {
			continue
		}

		if err != nil {
			return err
		}

		return fmt.Errorf("Kinesis Stream %s still exists", rs.Primary.ID)
	}

	return nil
}

func testAccStreamRegisterStreamConsumer(stream *kinesis.StreamDescriptionSummary, rName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).KinesisConn

		if _, err := conn.RegisterStreamConsumer(&kinesis.RegisterStreamConsumerInput{
			ConsumerName: aws.String(rName),
			StreamARN:    stream.StreamARN,
		}); err != nil {
			return err
		}

		return nil
	}
}

func testAccStreamConfig_basic(rName string) string {
	return fmt.Sprintf(`
resource "aws_kinesis_stream" "test" {
  name        = %[1]q
  shard_count = 2
}
`, rName)
}

func testAccStreamConfig_concurrent(rName string) string {
	return fmt.Sprintf(`
resource "aws_kinesis_stream" "test" {
  count       = 20
  name        = "%[1]s-${count.index}"
  shard_count = 2
}
`, rName)
}

func testAccStreamConfig_encryptionAndNoKMSKey(rName string) string {
	return fmt.Sprintf(`
resource "aws_kinesis_stream" "test" {
  name            = %[1]q
  shard_count     = 2
  encryption_type = "KMS"
}
`, rName)
}

func testAccStreamConfig_encryption(rName string) string {
	return fmt.Sprintf(`
resource "aws_kinesis_stream" "test" {
  name            = %[1]q
  shard_count     = 2
  encryption_type = "KMS"
  kms_key_id      = aws_kms_key.test.id
}

resource "aws_kms_key" "test" {
  description             = %[1]q
  deletion_window_in_days = 7

  policy = <<POLICY
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
`, rName)
}

func testAccStreamConfig_shardCount(rName string, shardCount int) string {
	return fmt.Sprintf(`
resource "aws_kinesis_stream" "test" {
  name        = %[1]q
  shard_count = %[2]d
}
`, rName, shardCount)
}

func testAccStreamConfig_updateRetentionPeriod(rName string) string {
	return fmt.Sprintf(`
resource "aws_kinesis_stream" "test" {
  name             = %[1]q
  shard_count      = 2
  retention_period = 8760
}
`, rName)
}

func testAccStreamConfig_decreaseRetentionPeriod(rName string) string {
	return fmt.Sprintf(`
resource "aws_kinesis_stream" "test" {
  name             = %[1]q
  shard_count      = 2
  retention_period = 28
}
`, rName)
}

func testAccStreamConfig_allShardLevelMetrics(rName string) string {
	return fmt.Sprintf(`
resource "aws_kinesis_stream" "test" {
  name        = %[1]q
  shard_count = 2

  shard_level_metrics = [
    "IncomingBytes",
    "IncomingRecords",
    "OutgoingBytes",
    "OutgoingRecords",
    "WriteProvisionedThroughputExceeded",
    "ReadProvisionedThroughputExceeded",
    "IteratorAgeMilliseconds",
  ]
}
`, rName)
}

func testAccStreamConfig_singleShardLevelMetric(rName string) string {
	return fmt.Sprintf(`
resource "aws_kinesis_stream" "test" {
  name        = %[1]q
  shard_count = 2

  shard_level_metrics = [
    "IncomingBytes",
  ]
}
`, rName)
}

func testAccStreamConfig_tags1(rName, tagKey1, tagValue1 string) string {
	return fmt.Sprintf(`
resource "aws_kinesis_stream" "test" {
  name        = %[1]q
  shard_count = 2

  tags = {
    %[2]q = %[3]q
  }
}
`, rName, tagKey1, tagValue1)
}

func testAccStreamConfig_tags2(rName, tagKey1, tagValue1, tagKey2, tagValue2 string) string {
	return fmt.Sprintf(`
resource "aws_kinesis_stream" "test" {
  name        = %[1]q
  shard_count = 2

  tags = {
    %[2]q = %[3]q
    %[4]q = %[5]q
  }
}
`, rName, tagKey1, tagValue1, tagKey2, tagValue2)
}

func testAccStreamConfig_enforceConsumerDeletion(rName string) string {
	return fmt.Sprintf(`
resource "aws_kinesis_stream" "test" {
  name                      = %[1]q
  shard_count               = 2
  enforce_consumer_deletion = true
}
`, rName)
}

func testAccStreamConfig_updateKMSKeyID(rName string, keyIdx int) string {
	return fmt.Sprintf(`
resource "aws_kms_key" "key" {
  count = 2

  description             = "%[1]s-${count.index}"
  deletion_window_in_days = 10
}

resource "aws_kinesis_stream" "test" {
  name            = %[1]q
  shard_count     = 1
  encryption_type = "KMS"
  kms_key_id      = aws_kms_key.key[%[2]d].id
}
`, rName, keyIdx)
}

func testAccStreamConfig_basicOnDemand(rName string) string {
	return fmt.Sprintf(`
resource "aws_kinesis_stream" "test" {
  name = %[1]q

  stream_mode_details {
    stream_mode = "ON_DEMAND"
  }
}
`, rName)
}

func testAccStreamConfig_changeProvisionedToOnDemand1(rName string) string {
	return fmt.Sprintf(`
resource "aws_kinesis_stream" "test" {
  name        = %[1]q
  shard_count = 1
}
`, rName)
}

func testAccStreamConfig_changeProvisionedToOnDemand2(rName string) string {
	return fmt.Sprintf(`
resource "aws_kinesis_stream" "test" {
  name = %[1]q

  stream_mode_details {
    stream_mode = "ON_DEMAND"
  }
}
`, rName)
}

func testAccStreamConfig_changeProvisionedToOnDemand3(rName string) string {
	return fmt.Sprintf(`
resource "aws_kinesis_stream" "test" {
  name        = %[1]q
  shard_count = 2

  stream_mode_details {
    stream_mode = "PROVISIONED"
  }
}
`, rName)
}

func testAccStreamConfig_failOnBadCountAndModeCombinationNothingSet(rName string) string {
	return fmt.Sprintf(`
resource "aws_kinesis_stream" "test" {
  name = %[1]q
}
`, rName)
}

func testAccStreamConfig_failOnBadCountAndModeCombinationShardCountWhenOnDemand(rName string) string {
	return fmt.Sprintf(`
resource "aws_kinesis_stream" "test" {
  name        = %[1]q
  shard_count = 1

  stream_mode_details {
    stream_mode = "ON_DEMAND"
  }
}
`, rName)
}

func testAccStreamConfig_failOnBadCountAndModeCombination(rName string) string {
	return fmt.Sprintf(`
resource "aws_kinesis_stream" "test" {
  name        = %[1]q
  shard_count = 1
}
`, rName)
}
