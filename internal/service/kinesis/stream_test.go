// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package kinesis_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/kinesis"
	"github.com/aws/aws-sdk-go-v2/service/kinesis/types"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfkinesis "github.com/hashicorp/terraform-provider-aws/internal/service/kinesis"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccKinesisStream_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var stream types.StreamDescriptionSummary
	resourceName := "aws_kinesis_stream.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.KinesisServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckStreamDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccStreamConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckStreamExists(ctx, resourceName, &stream),
					acctest.CheckResourceAttrRegionalARN(resourceName, names.AttrARN, "kinesis", fmt.Sprintf("stream/%s", rName)),
					resource.TestCheckResourceAttr(resourceName, "encryption_type", "NONE"),
					resource.TestCheckResourceAttr(resourceName, "enforce_consumer_deletion", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, names.AttrKMSKeyID, ""),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, names.AttrRetentionPeriod, "24"),
					resource.TestCheckResourceAttr(resourceName, "shard_count", acctest.Ct2),
					resource.TestCheckResourceAttr(resourceName, "shard_level_metrics.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "stream_mode_details.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "stream_mode_details.0.stream_mode", "PROVISIONED"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct0),
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
	ctx := acctest.Context(t)
	var stream types.StreamDescriptionSummary
	resourceName := "aws_kinesis_stream.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.KinesisServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckStreamDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccStreamConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckStreamExists(ctx, resourceName, &stream),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfkinesis.ResourceStream(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccKinesisStream_createMultipleConcurrentStreams(t *testing.T) {
	ctx := acctest.Context(t)
	var stream types.StreamDescriptionSummary
	resourceName := "aws_kinesis_stream.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.KinesisServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckStreamDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccStreamConfig_concurrent(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckStreamExists(ctx, "aws_kinesis_stream.test.0", &stream),
					testAccCheckStreamExists(ctx, "aws_kinesis_stream.test.1", &stream),
					testAccCheckStreamExists(ctx, "aws_kinesis_stream.test.2", &stream),
					testAccCheckStreamExists(ctx, "aws_kinesis_stream.test.3", &stream),
					testAccCheckStreamExists(ctx, "aws_kinesis_stream.test.4", &stream),
					testAccCheckStreamExists(ctx, "aws_kinesis_stream.test.5", &stream),
					testAccCheckStreamExists(ctx, "aws_kinesis_stream.test.6", &stream),
					testAccCheckStreamExists(ctx, "aws_kinesis_stream.test.7", &stream),
					testAccCheckStreamExists(ctx, "aws_kinesis_stream.test.8", &stream),
					testAccCheckStreamExists(ctx, "aws_kinesis_stream.test.9", &stream),
					testAccCheckStreamExists(ctx, "aws_kinesis_stream.test.10", &stream),
					testAccCheckStreamExists(ctx, "aws_kinesis_stream.test.11", &stream),
					testAccCheckStreamExists(ctx, "aws_kinesis_stream.test.12", &stream),
					testAccCheckStreamExists(ctx, "aws_kinesis_stream.test.13", &stream),
					testAccCheckStreamExists(ctx, "aws_kinesis_stream.test.14", &stream),
					testAccCheckStreamExists(ctx, "aws_kinesis_stream.test.15", &stream),
					testAccCheckStreamExists(ctx, "aws_kinesis_stream.test.16", &stream),
					testAccCheckStreamExists(ctx, "aws_kinesis_stream.test.17", &stream),
					testAccCheckStreamExists(ctx, "aws_kinesis_stream.test.18", &stream),
					testAccCheckStreamExists(ctx, "aws_kinesis_stream.test.19", &stream),
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
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.KinesisServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckStreamDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config:      testAccStreamConfig_encryptionAndNoKMSKey(rName),
				ExpectError: regexache.MustCompile("KMS Key ID required when setting encryption_type is not set as NONE"),
			},
		},
	})
}

func TestAccKinesisStream_encryption(t *testing.T) {
	ctx := acctest.Context(t)
	var stream types.StreamDescriptionSummary
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_kinesis_stream.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.KinesisServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckStreamDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccStreamConfig_encryption(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckStreamExists(ctx, resourceName, &stream),
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
					testAccCheckStreamExists(ctx, resourceName, &stream),
					resource.TestCheckResourceAttr(resourceName, "encryption_type", "NONE"),
				),
			},
			{
				Config: testAccStreamConfig_encryption(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckStreamExists(ctx, resourceName, &stream),
					resource.TestCheckResourceAttr(resourceName, "encryption_type", "KMS"),
				),
			},
		},
	})
}

func TestAccKinesisStream_shardCount(t *testing.T) {
	ctx := acctest.Context(t)
	var stream types.StreamDescriptionSummary
	var updatedStream types.StreamDescriptionSummary

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
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.KinesisServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckStreamDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccStreamConfig_shardCount(rName, 128),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckStreamExists(ctx, resourceName, &stream),
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
					testAccCheckStreamExists(ctx, resourceName, &updatedStream),
					testCheckStreamNotDestroyed(),
					resource.TestCheckResourceAttr(resourceName, "shard_count", "96"),
				),
			},
		},
	})
}

func TestAccKinesisStream_retentionPeriod(t *testing.T) {
	ctx := acctest.Context(t)
	var stream types.StreamDescriptionSummary
	resourceName := "aws_kinesis_stream.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.KinesisServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckStreamDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccStreamConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckStreamExists(ctx, resourceName, &stream),
					resource.TestCheckResourceAttr(resourceName, names.AttrRetentionPeriod, "24"),
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
					testAccCheckStreamExists(ctx, resourceName, &stream),
					resource.TestCheckResourceAttr(resourceName, names.AttrRetentionPeriod, "8760"),
				),
			},

			{
				Config: testAccStreamConfig_decreaseRetentionPeriod(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckStreamExists(ctx, resourceName, &stream),
					resource.TestCheckResourceAttr(resourceName, names.AttrRetentionPeriod, "28"),
				),
			},
		},
	})
}

func TestAccKinesisStream_shardLevelMetrics(t *testing.T) {
	ctx := acctest.Context(t)
	var stream types.StreamDescriptionSummary
	resourceName := "aws_kinesis_stream.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.KinesisServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckStreamDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccStreamConfig_singleShardLevelMetric(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckStreamExists(ctx, resourceName, &stream),
					resource.TestCheckResourceAttr(resourceName, "shard_level_metrics.#", acctest.Ct1),
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
					testAccCheckStreamExists(ctx, resourceName, &stream),
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
					testAccCheckStreamExists(ctx, resourceName, &stream),
					resource.TestCheckResourceAttr(resourceName, "shard_level_metrics.#", acctest.Ct1),
					resource.TestCheckTypeSetElemAttr(resourceName, "shard_level_metrics.*", "IncomingBytes"),
				),
			},
		},
	})
}

func TestAccKinesisStream_enforceConsumerDeletion(t *testing.T) {
	ctx := acctest.Context(t)
	var stream types.StreamDescriptionSummary
	resourceName := "aws_kinesis_stream.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.KinesisServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckStreamDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccStreamConfig_enforceConsumerDeletion(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckStreamExists(ctx, resourceName, &stream),
					testAccStreamRegisterStreamConsumer(ctx, &stream, rName),
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
	ctx := acctest.Context(t)
	var stream types.StreamDescriptionSummary
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_kinesis_stream.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.KinesisServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckStreamDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccStreamConfig_tags1(rName, acctest.CtKey1, acctest.CtValue1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckStreamExists(ctx, resourceName, &stream),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey1, acctest.CtValue1),
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
				Config: testAccStreamConfig_tags2(rName, acctest.CtKey1, acctest.CtValue1Updated, acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckStreamExists(ctx, resourceName, &stream),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct2),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey1, acctest.CtValue1Updated),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey2, acctest.CtValue2),
				),
			},
			{
				Config: testAccStreamConfig_tags1(rName, acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckStreamExists(ctx, resourceName, &stream),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey2, acctest.CtValue2),
				),
			},
		},
	})
}

func TestAccKinesisStream_updateKMSKeyID(t *testing.T) {
	ctx := acctest.Context(t)
	var stream types.StreamDescriptionSummary
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_kinesis_stream.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.KinesisServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckStreamDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccStreamConfig_updateKMSKeyID(rName, 0),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckStreamExists(ctx, resourceName, &stream),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrKMSKeyID, "aws_kms_key.key.0", names.AttrID),
				),
			},
			{
				Config: testAccStreamConfig_updateKMSKeyID(rName, 1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckStreamExists(ctx, resourceName, &stream),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrKMSKeyID, "aws_kms_key.key.1", names.AttrID),
				),
			},
		},
	})
}

func TestAccKinesisStream_basicOnDemand(t *testing.T) {
	ctx := acctest.Context(t)
	var stream types.StreamDescriptionSummary
	resourceName := "aws_kinesis_stream.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.KinesisServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckStreamDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccStreamConfig_basicOnDemand(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckStreamExists(ctx, resourceName, &stream),
					resource.TestCheckResourceAttr(resourceName, "shard_count", acctest.Ct0),
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
	ctx := acctest.Context(t)
	var stream types.StreamDescriptionSummary
	resourceName := "aws_kinesis_stream.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.KinesisServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckStreamDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccStreamConfig_changeProvisionedToOnDemand1(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckStreamExists(ctx, resourceName, &stream),
					resource.TestCheckResourceAttr(resourceName, "shard_count", acctest.Ct1),
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
					testAccCheckStreamExists(ctx, resourceName, &stream),
					resource.TestCheckResourceAttr(resourceName, "shard_count", acctest.Ct0),
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
					testAccCheckStreamExists(ctx, resourceName, &stream),
					resource.TestCheckResourceAttr(resourceName, "shard_count", acctest.Ct2),
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
					testAccCheckStreamExists(ctx, resourceName, &stream),
					resource.TestCheckResourceAttr(resourceName, "shard_count", acctest.Ct1),
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
	ctx := acctest.Context(t)
	var stream types.StreamDescriptionSummary
	resourceName := "aws_kinesis_stream.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.KinesisServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckStreamDestroy(ctx),
		Steps: []resource.TestStep{
			// Check that we can't create an invalid combination
			{
				Config:      testAccStreamConfig_failOnBadCountAndModeCombinationNothingSet(rName),
				ExpectError: regexache.MustCompile(`shard_count must be at least 1 when stream_mode is PROVISIONED`),
			},
			// Check that we can't create an invalid combination
			{
				Config:      testAccStreamConfig_failOnBadCountAndModeCombinationShardCountWhenOnDemand(rName),
				ExpectError: regexache.MustCompile(`shard_count must not be set when stream_mode is ON_DEMAND`),
			},
			// Prepare for updates...
			{
				Config: testAccStreamConfig_failOnBadCountAndModeCombination(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckStreamExists(ctx, resourceName, &stream),
					resource.TestCheckResourceAttr(resourceName, "shard_count", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "stream_mode_details.0.stream_mode", "PROVISIONED"),
				),
			},
			// Check that we can't update to an invalid combination
			{
				Config:      testAccStreamConfig_failOnBadCountAndModeCombinationNothingSet(rName),
				ExpectError: regexache.MustCompile(`shard_count must be at least 1 when stream_mode is PROVISIONED`),
			},
			// Check that we can't update to an invalid combination
			{
				Config:      testAccStreamConfig_failOnBadCountAndModeCombinationShardCountWhenOnDemand(rName),
				ExpectError: regexache.MustCompile(`shard_count must not be set when stream_mode is ON_DEMAND`),
			},
		},
	})
}

func testAccCheckStreamExists(ctx context.Context, n string, v *types.StreamDescriptionSummary) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).KinesisClient(ctx)

		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		output, err := tfkinesis.FindStreamByName(ctx, conn, rs.Primary.Attributes[names.AttrName])

		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

func testAccCheckStreamDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).KinesisClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_kinesis_stream" {
				continue
			}

			_, err := tfkinesis.FindStreamByName(ctx, conn, rs.Primary.Attributes[names.AttrName])

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
}

func testAccStreamRegisterStreamConsumer(ctx context.Context, stream *types.StreamDescriptionSummary, rName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).KinesisClient(ctx)

		if _, err := conn.RegisterStreamConsumer(ctx, &kinesis.RegisterStreamConsumerInput{
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
