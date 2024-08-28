// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package firehose_test

import (
	"context"
	"fmt"
	"strings"
	"testing"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/firehose"
	"github.com/aws/aws-sdk-go-v2/service/firehose/types"
	"github.com/aws/aws-sdk-go-v2/service/lambda"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tffirehose "github.com/hashicorp/terraform-provider-aws/internal/service/firehose"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccFirehoseDeliveryStream_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var stream types.DeliveryStreamDescription
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_kinesis_firehose_delivery_stream.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.FirehoseServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDeliveryStreamDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccDeliveryStreamConfig_extendedS3basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckDeliveryStreamExists(ctx, resourceName, &stream),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, names.AttrDestination, "extended_s3"),
					resource.TestCheckResourceAttrSet(resourceName, "destination_id"),
					resource.TestCheckResourceAttr(resourceName, "elasticsearch_configuration.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "extended_s3_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttrSet(resourceName, "extended_s3_configuration.0.bucket_arn"),
					resource.TestCheckResourceAttr(resourceName, "extended_s3_configuration.0.buffering_interval", "300"),
					resource.TestCheckResourceAttr(resourceName, "extended_s3_configuration.0.buffering_size", "5"),
					resource.TestCheckResourceAttr(resourceName, "extended_s3_configuration.0.cloudwatch_logging_options.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "extended_s3_configuration.0.cloudwatch_logging_options.0.enabled", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "extended_s3_configuration.0.cloudwatch_logging_options.0.log_group_name", ""),
					resource.TestCheckResourceAttr(resourceName, "extended_s3_configuration.0.cloudwatch_logging_options.0.log_stream_name", ""),
					resource.TestCheckResourceAttr(resourceName, "extended_s3_configuration.0.compression_format", "UNCOMPRESSED"),
					resource.TestCheckResourceAttr(resourceName, "extended_s3_configuration.0.custom_time_zone", "UTC"),
					resource.TestCheckResourceAttr(resourceName, "extended_s3_configuration.0.data_format_conversion_configuration.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "extended_s3_configuration.0.dynamic_partitioning_configuration.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "extended_s3_configuration.0.error_output_prefix", ""),
					resource.TestCheckResourceAttr(resourceName, "extended_s3_configuration.0.file_extension", ""),
					resource.TestCheckResourceAttr(resourceName, "extended_s3_configuration.0.kms_key_arn", ""),
					resource.TestCheckResourceAttr(resourceName, "extended_s3_configuration.0.prefix", ""),
					resource.TestCheckResourceAttr(resourceName, "extended_s3_configuration.0.processing_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "extended_s3_configuration.0.processing_configuration.0.enabled", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "extended_s3_configuration.0.processing_configuration.0.processors.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "extended_s3_configuration.0.processing_configuration.0.processors.0.type", "Lambda"),
					resource.TestCheckResourceAttr(resourceName, "extended_s3_configuration.0.processing_configuration.0.processors.0.parameters.#", acctest.Ct3),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "extended_s3_configuration.0.processing_configuration.0.processors.0.parameters.*", map[string]string{
						"parameter_name": "LambdaArn",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "extended_s3_configuration.0.processing_configuration.0.processors.0.parameters.*", map[string]string{
						"parameter_name":  "BufferSizeInMBs",
						"parameter_value": "1.1",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "extended_s3_configuration.0.processing_configuration.0.processors.0.parameters.*", map[string]string{
						"parameter_name":  "BufferIntervalInSeconds",
						"parameter_value": "70",
					}),
					resource.TestCheckResourceAttrSet(resourceName, "extended_s3_configuration.0.role_arn"),
					resource.TestCheckResourceAttr(resourceName, "extended_s3_configuration.0.s3_backup_configuration.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "extended_s3_configuration.0.s3_backup_mode", "Disabled"),
					resource.TestCheckResourceAttr(resourceName, "http_endpoint_configuration.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "kinesis_source_configuration.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "msk_source_configuration.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, "opensearch_configuration.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "opensearchserverless_configuration.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "redshift_configuration.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "server_side_encryption.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "server_side_encryption.0.enabled", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "server_side_encryption.0.key_arn", ""),
					resource.TestCheckResourceAttr(resourceName, "server_side_encryption.0.key_type", "AWS_OWNED_CMK"),
					resource.TestCheckResourceAttr(resourceName, "snowflake_configuration.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "splunk_configuration.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct0),
					resource.TestCheckResourceAttrSet(resourceName, "version_id"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			// Ensure we properly error on malformed import IDs
			{
				ResourceName:  resourceName,
				ImportState:   true,
				ImportStateId: "just-a-name",
				ExpectError:   regexache.MustCompile(`Expected ID in format`),
			},
			{
				ResourceName:  resourceName,
				ImportState:   true,
				ImportStateId: "arn:aws:firehose:us-east-1:123456789012:missing-slash", //lintignore:AWSAT003,AWSAT005
				ExpectError:   regexache.MustCompile(`Expected ID in format`),
			},
		},
	})
}

func TestAccFirehoseDeliveryStream_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	var stream types.DeliveryStreamDescription
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_kinesis_firehose_delivery_stream.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.FirehoseServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDeliveryStreamDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccDeliveryStreamConfig_extendedS3basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDeliveryStreamExists(ctx, resourceName, &stream),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tffirehose.ResourceDeliveryStream(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccFirehoseDeliveryStream_tags(t *testing.T) {
	ctx := acctest.Context(t)
	var stream types.DeliveryStreamDescription
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_kinesis_firehose_delivery_stream.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.FirehoseServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDeliveryStreamDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccDeliveryStreamConfig_tags1(rName, acctest.CtKey1, acctest.CtValue1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDeliveryStreamExists(ctx, resourceName, &stream),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey1, acctest.CtValue1),
				),
			},
			{
				Config: testAccDeliveryStreamConfig_tags2(rName, acctest.CtKey1, acctest.CtValue1Updated, acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDeliveryStreamExists(ctx, resourceName, &stream),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct2),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey1, acctest.CtValue1Updated),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey2, acctest.CtValue2),
				),
			},
			{
				Config: testAccDeliveryStreamConfig_tags1(rName, acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDeliveryStreamExists(ctx, resourceName, &stream),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey2, acctest.CtValue2),
				),
			},
		},
	})
}

func TestAccFirehoseDeliveryStream_s3WithCloudWatchLogging(t *testing.T) {
	ctx := acctest.Context(t)
	var stream types.DeliveryStreamDescription
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_kinesis_firehose_delivery_stream.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.FirehoseServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDeliveryStreamDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccDeliveryStreamConfig_s3CloudWatchLogging(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDeliveryStreamExists(ctx, resourceName, &stream),
					testAccCheckDeliveryStreamAttributes(&stream, nil, nil, nil, nil, nil, nil, nil),
				),
			},
		},
	})
}

func TestAccFirehoseDeliveryStream_extendedS3basic(t *testing.T) {
	ctx := acctest.Context(t)
	var stream types.DeliveryStreamDescription
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_kinesis_firehose_delivery_stream.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.FirehoseServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDeliveryStreamDestroy_ExtendedS3(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccDeliveryStreamConfig_extendedS3basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDeliveryStreamExists(ctx, resourceName, &stream),
					testAccCheckDeliveryStreamAttributes(&stream, nil, nil, nil, nil, nil, nil, nil),
					resource.TestCheckResourceAttr(resourceName, "extended_s3_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "extended_s3_configuration.0.error_output_prefix", ""),
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

func TestAccFirehoseDeliveryStream_ExtendedS3DataFormatConversion_enabled(t *testing.T) {
	ctx := acctest.Context(t)
	var stream types.DeliveryStreamDescription
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_kinesis_firehose_delivery_stream.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.FirehoseServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDeliveryStreamDestroy_ExtendedS3(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccDeliveryStreamConfig_extendedS3DataFormatConversionConfigurationEnabled(rName, false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDeliveryStreamExists(ctx, resourceName, &stream),
					resource.TestCheckResourceAttr(resourceName, "extended_s3_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "extended_s3_configuration.0.data_format_conversion_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "extended_s3_configuration.0.data_format_conversion_configuration.0.enabled", acctest.CtFalse),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccDeliveryStreamConfig_extendedS3DataFormatConversionConfigurationEnabled(rName, true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDeliveryStreamExists(ctx, resourceName, &stream),
					resource.TestCheckResourceAttr(resourceName, "extended_s3_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "extended_s3_configuration.0.data_format_conversion_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "extended_s3_configuration.0.data_format_conversion_configuration.0.enabled", acctest.CtTrue),
				),
			},
			{
				Config: testAccDeliveryStreamConfig_extendedS3DataFormatConversionConfigurationEnabled(rName, false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDeliveryStreamExists(ctx, resourceName, &stream),
					resource.TestCheckResourceAttr(resourceName, "extended_s3_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "extended_s3_configuration.0.data_format_conversion_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "extended_s3_configuration.0.data_format_conversion_configuration.0.enabled", acctest.CtFalse),
				),
			},
		},
	})
}

func TestAccFirehoseDeliveryStream_ExtendedS3_externalUpdate(t *testing.T) {
	ctx := acctest.Context(t)
	var stream types.DeliveryStreamDescription
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_kinesis_firehose_delivery_stream.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.FirehoseServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDeliveryStreamDestroy_ExtendedS3(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccDeliveryStreamConfig_extendedS3ExternalUpdate(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDeliveryStreamExists(ctx, resourceName, &stream),
					resource.TestCheckResourceAttr(resourceName, "extended_s3_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "extended_s3_configuration.0.data_format_conversion_configuration.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "extended_s3_configuration.0.processing_configuration.#", acctest.Ct1),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				PreConfig: func() {
					conn := acctest.Provider.Meta().(*conns.AWSClient).FirehoseClient(ctx)
					udi := firehose.UpdateDestinationInput{
						DeliveryStreamName:             aws.String(rName),
						DestinationId:                  aws.String("destinationId-000000000001"),
						CurrentDeliveryStreamVersionId: aws.String(acctest.Ct1),
						ExtendedS3DestinationUpdate: &types.ExtendedS3DestinationUpdate{
							DataFormatConversionConfiguration: &types.DataFormatConversionConfiguration{
								Enabled: aws.Bool(false),
							},
							ProcessingConfiguration: &types.ProcessingConfiguration{
								Enabled:    aws.Bool(false),
								Processors: []types.Processor{},
							},
						},
					}
					_, err := conn.UpdateDestination(ctx, &udi)
					if err != nil {
						t.Fatalf("Unable to update firehose destination: %s", err)
					}
				},
				Config: testAccDeliveryStreamConfig_extendedS3ExternalUpdate(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDeliveryStreamExists(ctx, resourceName, &stream),
					resource.TestCheckResourceAttr(resourceName, "extended_s3_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "extended_s3_configuration.0.data_format_conversion_configuration.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "extended_s3_configuration.0.processing_configuration.#", acctest.Ct1),
				),
			},
		},
	})
}

func TestAccFirehoseDeliveryStream_ExtendedS3DataFormatConversionDeserializer_update(t *testing.T) {
	ctx := acctest.Context(t)
	var stream types.DeliveryStreamDescription
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_kinesis_firehose_delivery_stream.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.FirehoseServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDeliveryStreamDestroy_ExtendedS3(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccDeliveryStreamConfig_extendedS3DataFormatConversionConfigurationHiveJSONSerDeEmpty(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDeliveryStreamExists(ctx, resourceName, &stream),
					resource.TestCheckResourceAttr(resourceName, "extended_s3_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "extended_s3_configuration.0.data_format_conversion_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "extended_s3_configuration.0.data_format_conversion_configuration.0.input_format_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "extended_s3_configuration.0.data_format_conversion_configuration.0.input_format_configuration.0.deserializer.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "extended_s3_configuration.0.data_format_conversion_configuration.0.input_format_configuration.0.deserializer.0.hive_json_ser_de.#", acctest.Ct1),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccDeliveryStreamConfig_extendedS3DataFormatConversionConfigurationOpenXJSONSerDeEmpty(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDeliveryStreamExists(ctx, resourceName, &stream),
					resource.TestCheckResourceAttr(resourceName, "extended_s3_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "extended_s3_configuration.0.data_format_conversion_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "extended_s3_configuration.0.data_format_conversion_configuration.0.input_format_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "extended_s3_configuration.0.data_format_conversion_configuration.0.input_format_configuration.0.deserializer.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "extended_s3_configuration.0.data_format_conversion_configuration.0.input_format_configuration.0.deserializer.0.open_x_json_ser_de.#", acctest.Ct1),
				),
			},
		},
	})
}

func TestAccFirehoseDeliveryStream_ExtendedS3DataFormatConversionHiveJSONSerDe_empty(t *testing.T) {
	ctx := acctest.Context(t)
	var stream types.DeliveryStreamDescription
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_kinesis_firehose_delivery_stream.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.FirehoseServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDeliveryStreamDestroy_ExtendedS3(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccDeliveryStreamConfig_extendedS3DataFormatConversionConfigurationHiveJSONSerDeEmpty(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDeliveryStreamExists(ctx, resourceName, &stream),
					resource.TestCheckResourceAttr(resourceName, "extended_s3_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "extended_s3_configuration.0.data_format_conversion_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "extended_s3_configuration.0.data_format_conversion_configuration.0.input_format_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "extended_s3_configuration.0.data_format_conversion_configuration.0.input_format_configuration.0.deserializer.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "extended_s3_configuration.0.data_format_conversion_configuration.0.input_format_configuration.0.deserializer.0.hive_json_ser_de.#", acctest.Ct1),
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

func TestAccFirehoseDeliveryStream_ExtendedS3DataFormatConversionOpenXJSONSerDe_empty(t *testing.T) {
	ctx := acctest.Context(t)
	var stream types.DeliveryStreamDescription
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_kinesis_firehose_delivery_stream.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.FirehoseServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDeliveryStreamDestroy_ExtendedS3(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccDeliveryStreamConfig_extendedS3DataFormatConversionConfigurationOpenXJSONSerDeEmpty(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDeliveryStreamExists(ctx, resourceName, &stream),
					resource.TestCheckResourceAttr(resourceName, "extended_s3_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "extended_s3_configuration.0.data_format_conversion_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "extended_s3_configuration.0.data_format_conversion_configuration.0.input_format_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "extended_s3_configuration.0.data_format_conversion_configuration.0.input_format_configuration.0.deserializer.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "extended_s3_configuration.0.data_format_conversion_configuration.0.input_format_configuration.0.deserializer.0.open_x_json_ser_de.#", acctest.Ct1),
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

func TestAccFirehoseDeliveryStream_ExtendedS3DataFormatConversionOrcSerDe_empty(t *testing.T) {
	ctx := acctest.Context(t)
	var stream types.DeliveryStreamDescription
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_kinesis_firehose_delivery_stream.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.FirehoseServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDeliveryStreamDestroy_ExtendedS3(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccDeliveryStreamConfig_extendedS3DataFormatConversionConfigurationOrcSerDeEmpty(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDeliveryStreamExists(ctx, resourceName, &stream),
					resource.TestCheckResourceAttr(resourceName, "extended_s3_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "extended_s3_configuration.0.data_format_conversion_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "extended_s3_configuration.0.data_format_conversion_configuration.0.output_format_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "extended_s3_configuration.0.data_format_conversion_configuration.0.output_format_configuration.0.serializer.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "extended_s3_configuration.0.data_format_conversion_configuration.0.output_format_configuration.0.serializer.0.orc_ser_de.#", acctest.Ct1),
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

func TestAccFirehoseDeliveryStream_ExtendedS3DataFormatConversionParquetSerDe_empty(t *testing.T) {
	ctx := acctest.Context(t)
	var stream types.DeliveryStreamDescription
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_kinesis_firehose_delivery_stream.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.FirehoseServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDeliveryStreamDestroy_ExtendedS3(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccDeliveryStreamConfig_extendedS3DataFormatConversionConfigurationParquetSerDeEmpty(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDeliveryStreamExists(ctx, resourceName, &stream),
					resource.TestCheckResourceAttr(resourceName, "extended_s3_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "extended_s3_configuration.0.data_format_conversion_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "extended_s3_configuration.0.data_format_conversion_configuration.0.output_format_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "extended_s3_configuration.0.data_format_conversion_configuration.0.output_format_configuration.0.serializer.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "extended_s3_configuration.0.data_format_conversion_configuration.0.output_format_configuration.0.serializer.0.parquet_ser_de.#", acctest.Ct1),
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

func TestAccFirehoseDeliveryStream_ExtendedS3DataFormatConversionSerializer_update(t *testing.T) {
	ctx := acctest.Context(t)
	var stream types.DeliveryStreamDescription
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_kinesis_firehose_delivery_stream.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.FirehoseServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDeliveryStreamDestroy_ExtendedS3(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccDeliveryStreamConfig_extendedS3DataFormatConversionConfigurationOrcSerDeEmpty(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDeliveryStreamExists(ctx, resourceName, &stream),
					resource.TestCheckResourceAttr(resourceName, "extended_s3_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "extended_s3_configuration.0.data_format_conversion_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "extended_s3_configuration.0.data_format_conversion_configuration.0.output_format_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "extended_s3_configuration.0.data_format_conversion_configuration.0.output_format_configuration.0.serializer.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "extended_s3_configuration.0.data_format_conversion_configuration.0.output_format_configuration.0.serializer.0.orc_ser_de.#", acctest.Ct1),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccDeliveryStreamConfig_extendedS3DataFormatConversionConfigurationParquetSerDeEmpty(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDeliveryStreamExists(ctx, resourceName, &stream),
					resource.TestCheckResourceAttr(resourceName, "extended_s3_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "extended_s3_configuration.0.data_format_conversion_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "extended_s3_configuration.0.data_format_conversion_configuration.0.output_format_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "extended_s3_configuration.0.data_format_conversion_configuration.0.output_format_configuration.0.serializer.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "extended_s3_configuration.0.data_format_conversion_configuration.0.output_format_configuration.0.serializer.0.parquet_ser_de.#", acctest.Ct1),
				),
			},
		},
	})
}

func TestAccFirehoseDeliveryStream_ExtendedS3_errorOutputPrefix(t *testing.T) {
	ctx := acctest.Context(t)
	var stream types.DeliveryStreamDescription
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_kinesis_firehose_delivery_stream.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.FirehoseServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDeliveryStreamDestroy_ExtendedS3(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccDeliveryStreamConfig_extendedS3ErrorOutputPrefix(rName, "prefix1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDeliveryStreamExists(ctx, resourceName, &stream),
					resource.TestCheckResourceAttr(resourceName, "extended_s3_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "extended_s3_configuration.0.error_output_prefix", "prefix1"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccDeliveryStreamConfig_extendedS3ErrorOutputPrefix(rName, "prefix2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDeliveryStreamExists(ctx, resourceName, &stream),
					resource.TestCheckResourceAttr(resourceName, "extended_s3_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "extended_s3_configuration.0.error_output_prefix", "prefix2"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				// Ensure the ErrorOutputPrefix can be updated to an empty value
				// Reference: https://github.com/hashicorp/terraform-provider-aws/pull/11229#discussion_r356282765
				Config: testAccDeliveryStreamConfig_extendedS3ErrorOutputPrefix(rName, ""),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDeliveryStreamExists(ctx, resourceName, &stream),
					resource.TestCheckResourceAttr(resourceName, "extended_s3_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "extended_s3_configuration.0.error_output_prefix", ""),
				),
			},
		},
	})
}

func TestAccFirehoseDeliveryStream_ExtendedS3_S3BackupConfiguration_ErrorOutputPrefix(t *testing.T) {
	ctx := acctest.Context(t)
	var stream types.DeliveryStreamDescription
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_kinesis_firehose_delivery_stream.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.FirehoseServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDeliveryStreamDestroy_ExtendedS3(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccDeliveryStreamConfig_extendedS3S3BackUpConfigurationErrorOutputPrefix(rName, "prefix1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDeliveryStreamExists(ctx, resourceName, &stream),
					resource.TestCheckResourceAttr(resourceName, "extended_s3_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "extended_s3_configuration.0.s3_backup_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "extended_s3_configuration.0.s3_backup_configuration.0.error_output_prefix", "prefix1"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccDeliveryStreamConfig_extendedS3S3BackUpConfigurationErrorOutputPrefix(rName, "prefix2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDeliveryStreamExists(ctx, resourceName, &stream),
					resource.TestCheckResourceAttr(resourceName, "extended_s3_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "extended_s3_configuration.0.s3_backup_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "extended_s3_configuration.0.s3_backup_configuration.0.error_output_prefix", "prefix2")),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccDeliveryStreamConfig_extendedS3S3BackUpConfigurationErrorOutputPrefix(rName, ""),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDeliveryStreamExists(ctx, resourceName, &stream),
					resource.TestCheckResourceAttr(resourceName, "extended_s3_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "extended_s3_configuration.0.s3_backup_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "extended_s3_configuration.0.s3_backup_configuration.0.error_output_prefix", "")),
			},
		},
	})
}

// Reference: https://github.com/hashicorp/terraform-provider-aws/issues/12600
func TestAccFirehoseDeliveryStream_ExtendedS3Processing_empty(t *testing.T) {
	ctx := acctest.Context(t)
	var stream types.DeliveryStreamDescription
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_kinesis_firehose_delivery_stream.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.FirehoseServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDeliveryStreamDestroy_ExtendedS3(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccDeliveryStreamConfig_extendedS3ProcessingConfigurationEmpty(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDeliveryStreamExists(ctx, resourceName, &stream),
					resource.TestCheckResourceAttr(resourceName, "extended_s3_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "extended_s3_configuration.0.processing_configuration.#", acctest.Ct1),
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

func TestAccFirehoseDeliveryStream_extendedS3KMSKeyARN(t *testing.T) {
	ctx := acctest.Context(t)
	var stream types.DeliveryStreamDescription
	resourceName := "aws_kinesis_firehose_delivery_stream.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.FirehoseServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDeliveryStreamDestroy_ExtendedS3(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccDeliveryStreamConfig_extendedS3KMSKeyARN(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDeliveryStreamExists(ctx, resourceName, &stream),
					testAccCheckDeliveryStreamAttributes(&stream, nil, nil, nil, nil, nil, nil, nil),
					resource.TestCheckResourceAttrPair(resourceName, "extended_s3_configuration.0.kms_key_arn", "aws_kms_key.test", names.AttrARN),
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

func TestAccFirehoseDeliveryStream_extendedS3DynamicPartitioning(t *testing.T) {
	ctx := acctest.Context(t)
	var stream types.DeliveryStreamDescription
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_kinesis_firehose_delivery_stream.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.FirehoseServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDeliveryStreamDestroy_ExtendedS3(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccDeliveryStreamConfig_extendedS3DynamicPartitioning(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDeliveryStreamExists(ctx, resourceName, &stream),
					testAccCheckDeliveryStreamAttributes(&stream, nil, nil, nil, nil, nil, nil, nil),
					resource.TestCheckResourceAttr(resourceName, "extended_s3_configuration.0.processing_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "extended_s3_configuration.0.dynamic_partitioning_configuration.#", acctest.Ct1),
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

func TestAccFirehoseDeliveryStream_extendedS3DynamicPartitioningUpdate(t *testing.T) {
	ctx := acctest.Context(t)
	var stream types.DeliveryStreamDescription
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_kinesis_firehose_delivery_stream.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.FirehoseServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDeliveryStreamDestroy_ExtendedS3(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccDeliveryStreamConfig_extendedS3DynamicPartitioningBasic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDeliveryStreamExists(ctx, resourceName, &stream),
					testAccCheckDeliveryStreamAttributes(&stream, nil, nil, nil, nil, nil, nil, nil),
				),
			},
			{
				Config: testAccDeliveryStreamConfig_extendedS3DynamicPartitioning(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDeliveryStreamExists(ctx, resourceName, &stream),
					testAccCheckDeliveryStreamAttributes(&stream, nil, nil, nil, nil, nil, nil, nil),
					resource.TestCheckResourceAttr(resourceName, "extended_s3_configuration.0.processing_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "extended_s3_configuration.0.dynamic_partitioning_configuration.#", acctest.Ct1),
				),
			},
		},
	})
}

func TestAccFirehoseDeliveryStream_extendedS3Updates(t *testing.T) {
	ctx := acctest.Context(t)
	var stream types.DeliveryStreamDescription
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_kinesis_firehose_delivery_stream.test"

	firstUpdateExtendedS3DestinationConfig := &types.ExtendedS3DestinationDescription{
		BufferingHints: &types.BufferingHints{
			IntervalInSeconds: aws.Int32(400),
			SizeInMBs:         aws.Int32(10),
		},
		ProcessingConfiguration: &types.ProcessingConfiguration{
			Enabled: aws.Bool(true),
			Processors: []types.Processor{
				{
					Type: types.ProcessorTypeLambda,
					Parameters: []types.ProcessorParameter{
						{
							ParameterName:  types.ProcessorParameterNameLambdaArn,
							ParameterValue: aws.String("valueNotTested"),
						},
					},
				},
			},
		},
		S3BackupMode: types.S3BackupModeEnabled,
	}

	secondUpdateExtendedS3DestinationConfig := &types.ExtendedS3DestinationDescription{
		BufferingHints: &types.BufferingHints{
			IntervalInSeconds: aws.Int32(400),
			SizeInMBs:         aws.Int32(10),
		},
		ProcessingConfiguration: &types.ProcessingConfiguration{
			Enabled: aws.Bool(true),
			Processors: []types.Processor{
				{
					Type: types.ProcessorTypeLambda,
					Parameters: []types.ProcessorParameter{
						{
							ParameterName:  types.ProcessorParameterNameLambdaArn,
							ParameterValue: aws.String("valueNotTested"),
						},
						{
							ParameterName:  types.ProcessorParameterNameBufferIntervalInSeconds,
							ParameterValue: aws.String("201"),
						},
					},
				},
			},
		},
		S3BackupMode: types.S3BackupModeEnabled,
	}

	removeProcessorsExtendedS3DestinationConfig := &types.ExtendedS3DestinationDescription{
		BufferingHints: &types.BufferingHints{
			IntervalInSeconds: aws.Int32(400),
			SizeInMBs:         aws.Int32(10),
		},
		ProcessingConfiguration: &types.ProcessingConfiguration{
			Enabled:    aws.Bool(false),
			Processors: []types.Processor{},
		},
		S3BackupMode: types.S3BackupModeEnabled,
	}

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.FirehoseServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDeliveryStreamDestroy_ExtendedS3(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccDeliveryStreamConfig_extendedS3basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDeliveryStreamExists(ctx, resourceName, &stream),
					testAccCheckDeliveryStreamAttributes(&stream, nil, nil, nil, nil, nil, nil, nil),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccDeliveryStreamConfig_extendedS3UpdatesInitial(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDeliveryStreamExists(ctx, resourceName, &stream),
					testAccCheckDeliveryStreamAttributes(&stream, nil, firstUpdateExtendedS3DestinationConfig, nil, nil, nil, nil, nil),
				),
			},
			{
				Config: testAccDeliveryStreamConfig_extendedS3UpdatesSetBufferIntervalNoBufferSize(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDeliveryStreamExists(ctx, resourceName, &stream),
					testAccCheckDeliveryStreamAttributes(&stream, nil, secondUpdateExtendedS3DestinationConfig, nil, nil, nil, nil, nil),
				),
			},
			{
				Config: testAccDeliveryStreamConfig_extendedS3UpdatesRemoveProcessors(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDeliveryStreamExists(ctx, resourceName, &stream),
					testAccCheckDeliveryStreamAttributes(&stream, nil, removeProcessorsExtendedS3DestinationConfig, nil, nil, nil, nil, nil),
				),
			},
		},
	})
}

func TestAccFirehoseDeliveryStream_extendedS3CustomTimeZoneAndFileExtensionUpdates(t *testing.T) {
	ctx := acctest.Context(t)
	var stream types.DeliveryStreamDescription
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_kinesis_firehose_delivery_stream.test"
	customTimeZone := "America/Los_Angeles"
	fileExtension := ".json"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.FirehoseServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDeliveryStreamDestroy_ExtendedS3(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccDeliveryStreamConfig_extendedS3CustomTimeZoneAndFileExtensionUpdates(rName, customTimeZone, fileExtension),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckDeliveryStreamExists(ctx, resourceName, &stream),
					resource.TestCheckResourceAttr(resourceName, "extended_s3_configuration.0.custom_time_zone", customTimeZone),
					resource.TestCheckResourceAttr(resourceName, "extended_s3_configuration.0.file_extension", fileExtension),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccDeliveryStreamConfig_extendedS3CustomTimeZoneAndFileExtensionUpdatesNoValues(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckDeliveryStreamExists(ctx, resourceName, &stream),
					resource.TestCheckResourceAttr(resourceName, "extended_s3_configuration.0.custom_time_zone", "UTC"),
					resource.TestCheckResourceAttr(resourceName, "extended_s3_configuration.0.file_extension", ""),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccDeliveryStreamConfig_extendedS3CustomTimeZoneAndFileExtensionUpdates(rName, customTimeZone, fileExtension),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckDeliveryStreamExists(ctx, resourceName, &stream),
					resource.TestCheckResourceAttr(resourceName, "extended_s3_configuration.0.custom_time_zone", customTimeZone),
					resource.TestCheckResourceAttr(resourceName, "extended_s3_configuration.0.file_extension", fileExtension),
				),
			},
		},
	})
}

func TestAccFirehoseDeliveryStream_ExtendedS3_kinesisStreamSource(t *testing.T) {
	ctx := acctest.Context(t)
	var stream types.DeliveryStreamDescription
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_kinesis_firehose_delivery_stream.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.FirehoseServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDeliveryStreamDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccDeliveryStreamConfig_extendedS3KinesisSource(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDeliveryStreamExists(ctx, resourceName, &stream),
					testAccCheckDeliveryStreamAttributes(&stream, nil, nil, nil, nil, nil, nil, nil),
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

func TestAccFirehoseDeliveryStream_ExtendedS3_mskClusterSource(t *testing.T) {
	ctx := acctest.Context(t)
	var stream types.DeliveryStreamDescription
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_kinesis_firehose_delivery_stream.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.FirehoseServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDeliveryStreamDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccDeliveryStreamConfig_extendedS3MSKSource(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDeliveryStreamExists(ctx, resourceName, &stream),
					resource.TestCheckResourceAttr(resourceName, "kinesis_source_configuration.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "msk_source_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "msk_source_configuration.0.authentication_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "msk_source_configuration.0.authentication_configuration.0.connectivity", "PRIVATE"),
					resource.TestCheckResourceAttrSet(resourceName, "msk_source_configuration.0.authentication_configuration.0.role_arn"),
					resource.TestCheckResourceAttrSet(resourceName, "msk_source_configuration.0.msk_cluster_arn"),
					resource.TestCheckResourceAttr(resourceName, "msk_source_configuration.0.topic_name", "test"),
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

func TestAccFirehoseDeliveryStream_redshiftUpdates(t *testing.T) {
	ctx := acctest.Context(t)
	var stream types.DeliveryStreamDescription
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_kinesis_firehose_delivery_stream.test"

	updatedRedshiftConfig := &types.RedshiftDestinationDescription{
		CopyCommand: &types.CopyCommand{
			CopyOptions: aws.String("GZIP"),
		},
		S3BackupMode: types.RedshiftS3BackupModeEnabled,
		ProcessingConfiguration: &types.ProcessingConfiguration{
			Enabled: aws.Bool(true),
			Processors: []types.Processor{
				{
					Type: types.ProcessorTypeLambda,
					Parameters: []types.ProcessorParameter{
						{
							ParameterName:  types.ProcessorParameterNameLambdaArn,
							ParameterValue: aws.String("valueNotTested"),
						},
					},
				},
			},
		},
	}

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.FirehoseServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDeliveryStreamDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccDeliveryStreamConfig_redshift(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDeliveryStreamExists(ctx, resourceName, &stream),
					testAccCheckDeliveryStreamAttributes(&stream, nil, nil, nil, nil, nil, nil, nil),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"redshift_configuration.0.password"},
			},
			{
				Config: testAccDeliveryStreamConfig_redshiftUpdates(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDeliveryStreamExists(ctx, resourceName, &stream),
					testAccCheckDeliveryStreamAttributes(&stream, nil, nil, updatedRedshiftConfig, nil, nil, nil, nil),
				),
			},
		},
	})
}

func TestAccFirehoseDeliveryStream_Redshift_SecretsManagerConfiguration(t *testing.T) {
	ctx := acctest.Context(t)
	var stream types.DeliveryStreamDescription
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_kinesis_firehose_delivery_stream.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.FirehoseServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDeliveryStreamDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccDeliveryStreamConfig_redshiftSecretsManager(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDeliveryStreamExists(ctx, resourceName, &stream),
					resource.TestCheckResourceAttr(resourceName, "redshift_configuration.0.secrets_manager_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "redshift_configuration.0.secrets_manager_configuration.0.enabled", acctest.CtTrue),
					resource.TestCheckResourceAttrPair(resourceName, "redshift_configuration.0.secrets_manager_configuration.0.secret_arn", "aws_secretsmanager_secret.test", names.AttrARN),
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

func TestAccFirehoseDeliveryStream_snowflakeUpdates(t *testing.T) {
	ctx := acctest.Context(t)
	var stream types.DeliveryStreamDescription
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	key := acctest.TLSRSAPrivateKeyPEM(t, 4096)
	resourceName := "aws_kinesis_firehose_delivery_stream.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.FirehoseServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDeliveryStreamDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccDeliveryStreamConfig_snowflakeBasic(rName, key),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckDeliveryStreamExists(ctx, resourceName, &stream),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, names.AttrDestination, "snowflake"),
					resource.TestCheckResourceAttrSet(resourceName, "destination_id"),
					resource.TestCheckResourceAttr(resourceName, "elasticsearch_configuration.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "extended_s3_configuration.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "http_endpoint_configuration.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "kinesis_source_configuration.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, "opensearch_configuration.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "opensearchserverless_configuration.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "redshift_configuration.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "server_side_encryption.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "server_side_encryption.0.enabled", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "server_side_encryption.0.key_arn", ""),
					resource.TestCheckResourceAttr(resourceName, "server_side_encryption.0.key_type", "AWS_OWNED_CMK"),
					resource.TestCheckResourceAttr(resourceName, "snowflake_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "snowflake_configuration.0.account_url", fmt.Sprintf("https://%s.snowflakecomputing.com", rName)),
					resource.TestCheckResourceAttr(resourceName, "snowflake_configuration.0.cloudwatch_logging_options.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "snowflake_configuration.0.cloudwatch_logging_options.0.enabled", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "snowflake_configuration.0.cloudwatch_logging_options.0.log_group_name", ""),
					resource.TestCheckResourceAttr(resourceName, "snowflake_configuration.0.cloudwatch_logging_options.0.log_stream_name", ""),
					resource.TestCheckResourceAttr(resourceName, "snowflake_configuration.0.content_column_name", ""),
					resource.TestCheckResourceAttr(resourceName, "snowflake_configuration.0.data_loading_option", "JSON_MAPPING"),
					resource.TestCheckResourceAttr(resourceName, "snowflake_configuration.0.database", "test-db"),
					resource.TestCheckResourceAttr(resourceName, "snowflake_configuration.0.key_passphrase", ""),
					resource.TestCheckResourceAttr(resourceName, "snowflake_configuration.0.metadata_column_name", ""),
					resource.TestCheckResourceAttrSet(resourceName, "snowflake_configuration.0.private_key"),
					resource.TestCheckResourceAttr(resourceName, "snowflake_configuration.0.processing_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "snowflake_configuration.0.processing_configuration.0.enabled", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "snowflake_configuration.0.processing_configuration.0.processors.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "snowflake_configuration.0.retry_duration", "60"),
					resource.TestCheckResourceAttrSet(resourceName, "snowflake_configuration.0.role_arn"),
					resource.TestCheckResourceAttr(resourceName, "snowflake_configuration.0.s3_backup_mode", "FailedDataOnly"),
					resource.TestCheckResourceAttrSet(resourceName, "snowflake_configuration.0.s3_configuration.0.bucket_arn"),
					resource.TestCheckResourceAttr(resourceName, "snowflake_configuration.0.s3_configuration.0.buffering_interval", "400"),
					resource.TestCheckResourceAttr(resourceName, "snowflake_configuration.0.s3_configuration.0.buffering_size", acctest.Ct10),
					resource.TestCheckResourceAttr(resourceName, "snowflake_configuration.0.s3_configuration.0.cloudwatch_logging_options.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "snowflake_configuration.0.s3_configuration.0.cloudwatch_logging_options.0.enabled", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "snowflake_configuration.0.s3_configuration.0.cloudwatch_logging_options.0.log_group_name", ""),
					resource.TestCheckResourceAttr(resourceName, "snowflake_configuration.0.s3_configuration.0.cloudwatch_logging_options.0.log_stream_name", ""),
					resource.TestCheckResourceAttr(resourceName, "snowflake_configuration.0.s3_configuration.0.compression_format", "GZIP"),
					resource.TestCheckResourceAttr(resourceName, "snowflake_configuration.0.s3_configuration.0.error_output_prefix", ""),
					resource.TestCheckResourceAttr(resourceName, "snowflake_configuration.0.s3_configuration.0.kms_key_arn", ""),
					resource.TestCheckResourceAttr(resourceName, "snowflake_configuration.0.s3_configuration.0.prefix", ""),
					resource.TestCheckResourceAttrSet(resourceName, "snowflake_configuration.0.s3_configuration.0.role_arn"),
					resource.TestCheckResourceAttr(resourceName, "snowflake_configuration.0.schema", "test-schema"),
					resource.TestCheckResourceAttr(resourceName, "snowflake_configuration.0.snowflake_role_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "snowflake_configuration.0.snowflake_role_configuration.0.enabled", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "snowflake_configuration.0.snowflake_role_configuration.0.snowflake_role", ""),
					resource.TestCheckResourceAttr(resourceName, "snowflake_configuration.0.snowflake_vpc_configuration.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "snowflake_configuration.0.table", "test-table"),
					resource.TestCheckResourceAttr(resourceName, "snowflake_configuration.0.user", "test-usr"),
					resource.TestCheckResourceAttr(resourceName, "splunk_configuration.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct0),
					resource.TestCheckResourceAttrSet(resourceName, "version_id"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"snowflake_configuration.0.private_key"},
			},
			{
				Config: testAccDeliveryStreamConfig_snowflakeUpdate(rName, key),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckDeliveryStreamExists(ctx, resourceName, &stream),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, names.AttrDestination, "snowflake"),
					resource.TestCheckResourceAttrSet(resourceName, "destination_id"),
					resource.TestCheckResourceAttr(resourceName, "elasticsearch_configuration.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "extended_s3_configuration.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "http_endpoint_configuration.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "kinesis_source_configuration.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, "opensearch_configuration.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "opensearchserverless_configuration.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "redshift_configuration.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "server_side_encryption.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "server_side_encryption.0.enabled", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "server_side_encryption.0.key_arn", ""),
					resource.TestCheckResourceAttr(resourceName, "server_side_encryption.0.key_type", "AWS_OWNED_CMK"),
					resource.TestCheckResourceAttr(resourceName, "snowflake_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "snowflake_configuration.0.account_url", fmt.Sprintf("https://%s.snowflakecomputing.com", rName)),
					resource.TestCheckResourceAttr(resourceName, "snowflake_configuration.0.cloudwatch_logging_options.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "snowflake_configuration.0.cloudwatch_logging_options.0.enabled", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "snowflake_configuration.0.cloudwatch_logging_options.0.log_group_name", ""),
					resource.TestCheckResourceAttr(resourceName, "snowflake_configuration.0.cloudwatch_logging_options.0.log_stream_name", ""),
					resource.TestCheckResourceAttr(resourceName, "snowflake_configuration.0.content_column_name", "test-content"),
					resource.TestCheckResourceAttr(resourceName, "snowflake_configuration.0.data_loading_option", "VARIANT_CONTENT_MAPPING"),
					resource.TestCheckResourceAttr(resourceName, "snowflake_configuration.0.database", "test-db"),
					resource.TestCheckResourceAttr(resourceName, "snowflake_configuration.0.key_passphrase", ""),
					resource.TestCheckResourceAttr(resourceName, "snowflake_configuration.0.metadata_column_name", ""),
					resource.TestCheckResourceAttrSet(resourceName, "snowflake_configuration.0.private_key"),
					resource.TestCheckResourceAttr(resourceName, "snowflake_configuration.0.processing_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "snowflake_configuration.0.processing_configuration.0.enabled", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "snowflake_configuration.0.processing_configuration.0.processors.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "snowflake_configuration.0.processing_configuration.0.processors.0.type", "Lambda"),
					resource.TestCheckResourceAttr(resourceName, "snowflake_configuration.0.processing_configuration.0.processors.0.parameters.#", acctest.Ct3),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "snowflake_configuration.0.processing_configuration.0.processors.0.parameters.*", map[string]string{
						"parameter_name": "LambdaArn",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "snowflake_configuration.0.processing_configuration.0.processors.0.parameters.*", map[string]string{
						"parameter_name":  "BufferSizeInMBs",
						"parameter_value": "1.1",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "snowflake_configuration.0.processing_configuration.0.processors.0.parameters.*", map[string]string{
						"parameter_name":  "BufferIntervalInSeconds",
						"parameter_value": "70",
					}),
					resource.TestCheckResourceAttr(resourceName, "snowflake_configuration.0.retry_duration", "60"),
					resource.TestCheckResourceAttrSet(resourceName, "snowflake_configuration.0.role_arn"),
					resource.TestCheckResourceAttr(resourceName, "snowflake_configuration.0.s3_backup_mode", "FailedDataOnly"),
					resource.TestCheckResourceAttrSet(resourceName, "snowflake_configuration.0.s3_configuration.0.bucket_arn"),
					resource.TestCheckResourceAttr(resourceName, "snowflake_configuration.0.s3_configuration.0.buffering_interval", "400"),
					resource.TestCheckResourceAttr(resourceName, "snowflake_configuration.0.s3_configuration.0.buffering_size", acctest.Ct10),
					resource.TestCheckResourceAttr(resourceName, "snowflake_configuration.0.s3_configuration.0.cloudwatch_logging_options.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "snowflake_configuration.0.s3_configuration.0.cloudwatch_logging_options.0.enabled", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "snowflake_configuration.0.s3_configuration.0.cloudwatch_logging_options.0.log_group_name", ""),
					resource.TestCheckResourceAttr(resourceName, "snowflake_configuration.0.s3_configuration.0.cloudwatch_logging_options.0.log_stream_name", ""),
					resource.TestCheckResourceAttr(resourceName, "snowflake_configuration.0.s3_configuration.0.compression_format", "GZIP"),
					resource.TestCheckResourceAttr(resourceName, "snowflake_configuration.0.s3_configuration.0.error_output_prefix", ""),
					resource.TestCheckResourceAttr(resourceName, "snowflake_configuration.0.s3_configuration.0.kms_key_arn", ""),
					resource.TestCheckResourceAttr(resourceName, "snowflake_configuration.0.s3_configuration.0.prefix", ""),
					resource.TestCheckResourceAttrSet(resourceName, "snowflake_configuration.0.s3_configuration.0.role_arn"),
					resource.TestCheckResourceAttr(resourceName, "snowflake_configuration.0.schema", "test-schema"),
					resource.TestCheckResourceAttr(resourceName, "snowflake_configuration.0.snowflake_role_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "snowflake_configuration.0.snowflake_role_configuration.0.enabled", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "snowflake_configuration.0.snowflake_role_configuration.0.snowflake_role", "test-role"),
					resource.TestCheckResourceAttr(resourceName, "snowflake_configuration.0.snowflake_vpc_configuration.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "snowflake_configuration.0.table", "test-table"),
					resource.TestCheckResourceAttr(resourceName, "snowflake_configuration.0.user", "test-usr"),
					resource.TestCheckResourceAttr(resourceName, "splunk_configuration.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct0),
					resource.TestCheckResourceAttrSet(resourceName, "version_id"),
				),
			},
			{
				Config: testAccDeliveryStreamConfig_snowflakeUpdateSecretsManager(rName, key),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckDeliveryStreamExists(ctx, resourceName, &stream),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, names.AttrDestination, "snowflake"),
					resource.TestCheckResourceAttrSet(resourceName, "destination_id"),
					resource.TestCheckResourceAttr(resourceName, "elasticsearch_configuration.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "extended_s3_configuration.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "http_endpoint_configuration.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "kinesis_source_configuration.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, "opensearch_configuration.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "opensearchserverless_configuration.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "redshift_configuration.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "server_side_encryption.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "server_side_encryption.0.enabled", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "server_side_encryption.0.key_arn", ""),
					resource.TestCheckResourceAttr(resourceName, "server_side_encryption.0.key_type", "AWS_OWNED_CMK"),
					resource.TestCheckResourceAttr(resourceName, "snowflake_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "snowflake_configuration.0.account_url", fmt.Sprintf("https://%s.snowflakecomputing.com", rName)),
					resource.TestCheckResourceAttr(resourceName, "snowflake_configuration.0.cloudwatch_logging_options.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "snowflake_configuration.0.cloudwatch_logging_options.0.enabled", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "snowflake_configuration.0.cloudwatch_logging_options.0.log_group_name", ""),
					resource.TestCheckResourceAttr(resourceName, "snowflake_configuration.0.cloudwatch_logging_options.0.log_stream_name", ""),
					resource.TestCheckResourceAttr(resourceName, "snowflake_configuration.0.content_column_name", ""),
					resource.TestCheckResourceAttr(resourceName, "snowflake_configuration.0.data_loading_option", "JSON_MAPPING"),
					resource.TestCheckResourceAttr(resourceName, "snowflake_configuration.0.database", "test-db"),
					resource.TestCheckResourceAttr(resourceName, "snowflake_configuration.0.key_passphrase", ""),
					resource.TestCheckResourceAttr(resourceName, "snowflake_configuration.0.metadata_column_name", ""),
					resource.TestCheckResourceAttr(resourceName, "snowflake_configuration.0.private_key", ""),
					resource.TestCheckResourceAttr(resourceName, "snowflake_configuration.0.processing_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "snowflake_configuration.0.processing_configuration.0.enabled", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "snowflake_configuration.0.processing_configuration.0.processors.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "snowflake_configuration.0.retry_duration", "60"),
					resource.TestCheckResourceAttrSet(resourceName, "snowflake_configuration.0.role_arn"),
					resource.TestCheckResourceAttr(resourceName, "snowflake_configuration.0.s3_backup_mode", "FailedDataOnly"),
					resource.TestCheckResourceAttrSet(resourceName, "snowflake_configuration.0.s3_configuration.0.bucket_arn"),
					resource.TestCheckResourceAttr(resourceName, "snowflake_configuration.0.s3_configuration.0.buffering_interval", "400"),
					resource.TestCheckResourceAttr(resourceName, "snowflake_configuration.0.s3_configuration.0.buffering_size", acctest.Ct10),
					resource.TestCheckResourceAttr(resourceName, "snowflake_configuration.0.s3_configuration.0.cloudwatch_logging_options.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "snowflake_configuration.0.s3_configuration.0.cloudwatch_logging_options.0.enabled", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "snowflake_configuration.0.s3_configuration.0.cloudwatch_logging_options.0.log_group_name", ""),
					resource.TestCheckResourceAttr(resourceName, "snowflake_configuration.0.s3_configuration.0.cloudwatch_logging_options.0.log_stream_name", ""),
					resource.TestCheckResourceAttr(resourceName, "snowflake_configuration.0.s3_configuration.0.compression_format", "GZIP"),
					resource.TestCheckResourceAttr(resourceName, "snowflake_configuration.0.s3_configuration.0.error_output_prefix", ""),
					resource.TestCheckResourceAttr(resourceName, "snowflake_configuration.0.s3_configuration.0.kms_key_arn", ""),
					resource.TestCheckResourceAttr(resourceName, "snowflake_configuration.0.s3_configuration.0.prefix", ""),
					resource.TestCheckResourceAttrSet(resourceName, "snowflake_configuration.0.s3_configuration.0.role_arn"),
					resource.TestCheckResourceAttr(resourceName, "snowflake_configuration.0.schema", "test-schema"),
					resource.TestCheckResourceAttr(resourceName, "snowflake_configuration.0.secrets_manager_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "snowflake_configuration.0.secrets_manager_configuration.0.enabled", acctest.CtTrue),
					resource.TestCheckResourceAttrSet(resourceName, "snowflake_configuration.0.secrets_manager_configuration.0.role_arn"),
					resource.TestCheckResourceAttrSet(resourceName, "snowflake_configuration.0.secrets_manager_configuration.0.secret_arn"),
					resource.TestCheckResourceAttr(resourceName, "snowflake_configuration.0.snowflake_role_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "snowflake_configuration.0.snowflake_role_configuration.0.enabled", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "snowflake_configuration.0.snowflake_role_configuration.0.snowflake_role", ""),
					resource.TestCheckResourceAttr(resourceName, "snowflake_configuration.0.snowflake_vpc_configuration.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "snowflake_configuration.0.table", "test-table"),
					resource.TestCheckResourceAttr(resourceName, "snowflake_configuration.0.user", ""),
					resource.TestCheckResourceAttr(resourceName, "splunk_configuration.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct0),
					resource.TestCheckResourceAttrSet(resourceName, "version_id"),
				),
			},
		},
	})
}

func TestAccFirehoseDeliveryStream_splunkUpdates(t *testing.T) {
	ctx := acctest.Context(t)
	var stream types.DeliveryStreamDescription
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_kinesis_firehose_delivery_stream.test"

	updatedSplunkConfig := &types.SplunkDestinationDescription{
		BufferingHints: &types.SplunkBufferingHints{
			IntervalInSeconds: aws.Int32(45),
			SizeInMBs:         aws.Int32(3),
		},
		HECEndpointType:                   types.HECEndpointTypeEvent,
		HECAcknowledgmentTimeoutInSeconds: aws.Int32(600),
		S3BackupMode:                      types.SplunkS3BackupModeFailedEventsOnly,
		ProcessingConfiguration: &types.ProcessingConfiguration{
			Enabled: aws.Bool(true),
			Processors: []types.Processor{
				{
					Type: types.ProcessorTypeLambda,
					Parameters: []types.ProcessorParameter{
						{
							ParameterName:  types.ProcessorParameterNameLambdaArn,
							ParameterValue: aws.String("valueNotTested"),
						},
					},
				},
			},
		},
	}

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.FirehoseServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDeliveryStreamDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccDeliveryStreamConfig_splunkBasic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDeliveryStreamExists(ctx, resourceName, &stream),
					testAccCheckDeliveryStreamAttributes(&stream, nil, nil, nil, nil, nil, nil, nil),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccDeliveryStreamConfig_splunkUpdates(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDeliveryStreamExists(ctx, resourceName, &stream),
					testAccCheckDeliveryStreamAttributes(&stream, nil, nil, nil, nil, nil, updatedSplunkConfig, nil),
				),
			},
		},
	})
}

func TestAccFirehoseDeliveryStream_Splunk_ErrorOutputPrefix(t *testing.T) {
	ctx := acctest.Context(t)
	var stream types.DeliveryStreamDescription
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_kinesis_firehose_delivery_stream.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.FirehoseServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDeliveryStreamDestroy_ExtendedS3(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccDeliveryStreamConfig_splunkErrorOutputPrefix(rName, "prefix1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDeliveryStreamExists(ctx, resourceName, &stream),
					resource.TestCheckResourceAttr(resourceName, "splunk_configuration.0.s3_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "splunk_configuration.0.s3_configuration.0.error_output_prefix", "prefix1"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccDeliveryStreamConfig_splunkErrorOutputPrefix(rName, "prefix2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDeliveryStreamExists(ctx, resourceName, &stream),
					resource.TestCheckResourceAttr(resourceName, "splunk_configuration.0.s3_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "splunk_configuration.0.s3_configuration.0.error_output_prefix", "prefix2"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccDeliveryStreamConfig_splunkErrorOutputPrefix(rName, ""),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDeliveryStreamExists(ctx, resourceName, &stream),
					resource.TestCheckResourceAttr(resourceName, "splunk_configuration.0.s3_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "splunk_configuration.0.s3_configuration.0.error_output_prefix", ""),
				),
			},
		},
	})
}

func TestAccFirehoseDeliveryStream_Splunk_SecretsManagerConfiguration(t *testing.T) {
	ctx := acctest.Context(t)
	var stream types.DeliveryStreamDescription
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_kinesis_firehose_delivery_stream.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.FirehoseServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDeliveryStreamDestroy_ExtendedS3(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccDeliveryStreamConfig_splunkSecretsManager(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDeliveryStreamExists(ctx, resourceName, &stream),
					resource.TestCheckResourceAttr(resourceName, "splunk_configuration.0.secrets_manager_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "splunk_configuration.0.secrets_manager_configuration.0.enabled", acctest.CtTrue),
					resource.TestCheckResourceAttrPair(resourceName, "splunk_configuration.0.secrets_manager_configuration.0.secret_arn", "aws_secretsmanager_secret.test", names.AttrARN),
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

func TestAccFirehoseDeliveryStream_httpEndpoint(t *testing.T) {
	ctx := acctest.Context(t)
	var stream types.DeliveryStreamDescription
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_kinesis_firehose_delivery_stream.test"

	updatedHTTPEndpointConfig := &types.HttpEndpointDestinationDescription{
		EndpointConfiguration: &types.HttpEndpointDescription{
			Url:  aws.String("https://input-test.com:443"),
			Name: aws.String("HTTP_test"),
		},
		S3BackupMode: types.HttpEndpointS3BackupModeFailedDataOnly,
		ProcessingConfiguration: &types.ProcessingConfiguration{
			Enabled: aws.Bool(true),
			Processors: []types.Processor{
				{
					Type: types.ProcessorTypeLambda,
					Parameters: []types.ProcessorParameter{
						{
							ParameterName:  types.ProcessorParameterNameLambdaArn,
							ParameterValue: aws.String("valueNotTested"),
						},
					},
				},
			},
		},
	}

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.FirehoseServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDeliveryStreamDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccDeliveryStreamConfig_httpEndpointBasic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDeliveryStreamExists(ctx, resourceName, &stream),
					testAccCheckDeliveryStreamAttributes(&stream, nil, nil, nil, nil, nil, nil, nil),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccDeliveryStreamConfig_httpEndpointUpdates(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDeliveryStreamExists(ctx, resourceName, &stream),
					testAccCheckDeliveryStreamAttributes(&stream, nil, nil, nil, nil, nil, nil, updatedHTTPEndpointConfig),
				),
			},
		},
	})
}

func TestAccFirehoseDeliveryStream_HTTPEndpoint_ErrorOutputPrefix(t *testing.T) {
	ctx := acctest.Context(t)
	var stream types.DeliveryStreamDescription
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_kinesis_firehose_delivery_stream.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.FirehoseServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDeliveryStreamDestroy_ExtendedS3(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccDeliveryStreamConfig_httpEndpointErrorOutputPrefix(rName, "prefix1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDeliveryStreamExists(ctx, resourceName, &stream),
					resource.TestCheckResourceAttr(resourceName, "http_endpoint_configuration.0.s3_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "http_endpoint_configuration.0.s3_configuration.0.error_output_prefix", "prefix1"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccDeliveryStreamConfig_httpEndpointErrorOutputPrefix(rName, "prefix2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDeliveryStreamExists(ctx, resourceName, &stream),
					resource.TestCheckResourceAttr(resourceName, "http_endpoint_configuration.0.s3_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "http_endpoint_configuration.0.s3_configuration.0.error_output_prefix", "prefix2"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccDeliveryStreamConfig_httpEndpointErrorOutputPrefix(rName, ""),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDeliveryStreamExists(ctx, resourceName, &stream),
					resource.TestCheckResourceAttr(resourceName, "http_endpoint_configuration.0.s3_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "http_endpoint_configuration.0.s3_configuration.0.error_output_prefix", ""),
				),
			},
		},
	})
}

func TestAccFirehoseDeliveryStream_HTTPEndpoint_retryDuration(t *testing.T) {
	ctx := acctest.Context(t)
	var stream types.DeliveryStreamDescription
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_kinesis_firehose_delivery_stream.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.FirehoseServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDeliveryStreamDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccDeliveryStreamConfig_httpEndpointRetryDuration(rName, 301),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDeliveryStreamExists(ctx, resourceName, &stream),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccDeliveryStreamConfig_httpEndpointRetryDuration(rName, 302),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDeliveryStreamExists(ctx, resourceName, &stream),
				),
			},
		},
	})
}

func TestAccFirehoseDeliveryStream_HTTPEndpoint_SecretsManagerConfiguration(t *testing.T) {
	ctx := acctest.Context(t)
	var stream types.DeliveryStreamDescription
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_kinesis_firehose_delivery_stream.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.FirehoseServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDeliveryStreamDestroy_ExtendedS3(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccDeliveryStreamConfig_httpEndpointSecretsManager(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDeliveryStreamExists(ctx, resourceName, &stream),
					resource.TestCheckResourceAttr(resourceName, "http_endpoint_configuration.0.secrets_manager_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "http_endpoint_configuration.0.secrets_manager_configuration.0.enabled", acctest.CtTrue),
					resource.TestCheckResourceAttrPair(resourceName, "http_endpoint_configuration.0.secrets_manager_configuration.0.secret_arn", "aws_secretsmanager_secret.test", names.AttrARN),
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

func TestAccFirehoseDeliveryStream_elasticSearchUpdates(t *testing.T) {
	ctx := acctest.Context(t)
	var stream types.DeliveryStreamDescription
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_kinesis_firehose_delivery_stream.test"

	updatedElasticsearchConfig := &types.ElasticsearchDestinationDescription{
		BufferingHints: &types.ElasticsearchBufferingHints{
			IntervalInSeconds: aws.Int32(500),
		},
		ProcessingConfiguration: &types.ProcessingConfiguration{
			Enabled: aws.Bool(true),
			Processors: []types.Processor{
				{
					Type: types.ProcessorTypeLambda,
					Parameters: []types.ProcessorParameter{
						{
							ParameterName:  types.ProcessorParameterNameLambdaArn,
							ParameterValue: aws.String("valueNotTested"),
						},
					},
				},
			},
		},
	}

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckIAMServiceLinkedRole(ctx, t, "/aws-service-role/es") },
		ErrorCheck:               acctest.ErrorCheck(t, names.FirehoseServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDeliveryStreamDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccDeliveryStreamConfig_elasticsearchBasic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDeliveryStreamExists(ctx, resourceName, &stream),
					testAccCheckDeliveryStreamAttributes(&stream, nil, nil, nil, nil, nil, nil, nil),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccDeliveryStreamConfig_elasticsearchUpdate(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDeliveryStreamExists(ctx, resourceName, &stream),
					testAccCheckDeliveryStreamAttributes(&stream, nil, nil, nil, updatedElasticsearchConfig, nil, nil, nil),
				),
			},
		},
	})
}

func TestAccFirehoseDeliveryStream_elasticSearchEndpointUpdates(t *testing.T) {
	ctx := acctest.Context(t)
	var stream types.DeliveryStreamDescription
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_kinesis_firehose_delivery_stream.test"

	updatedElasticsearchConfig := &types.ElasticsearchDestinationDescription{
		BufferingHints: &types.ElasticsearchBufferingHints{
			IntervalInSeconds: aws.Int32(500),
		},
		ProcessingConfiguration: &types.ProcessingConfiguration{
			Enabled: aws.Bool(true),
			Processors: []types.Processor{
				{
					Type: types.ProcessorTypeLambda,
					Parameters: []types.ProcessorParameter{
						{
							ParameterName:  types.ProcessorParameterNameLambdaArn,
							ParameterValue: aws.String("valueNotTested"),
						},
					},
				},
			},
		},
	}

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckIAMServiceLinkedRole(ctx, t, "/aws-service-role/es") },
		ErrorCheck:               acctest.ErrorCheck(t, names.FirehoseServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDeliveryStreamDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccDeliveryStreamConfig_elasticsearchEndpoint(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDeliveryStreamExists(ctx, resourceName, &stream),
					testAccCheckDeliveryStreamAttributes(&stream, nil, nil, nil, nil, nil, nil, nil),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccDeliveryStreamConfig_elasticsearchEndpointUpdate(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDeliveryStreamExists(ctx, resourceName, &stream),
					testAccCheckDeliveryStreamAttributes(&stream, nil, nil, nil, updatedElasticsearchConfig, nil, nil, nil),
				),
			},
		},
	})
}

// This doesn't actually test updating VPC Configuration. It tests changing Elasticsearch configuration
// when the Kinesis Firehose delivery stream has a VPC Configuration.
func TestAccFirehoseDeliveryStream_elasticSearchWithVPCUpdates(t *testing.T) {
	ctx := acctest.Context(t)
	var stream types.DeliveryStreamDescription
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_kinesis_firehose_delivery_stream.test"

	updatedElasticsearchConfig := &types.ElasticsearchDestinationDescription{
		BufferingHints: &types.ElasticsearchBufferingHints{
			IntervalInSeconds: aws.Int32(500),
		},
		ProcessingConfiguration: &types.ProcessingConfiguration{
			Enabled: aws.Bool(true),
			Processors: []types.Processor{
				{
					Type: types.ProcessorTypeLambda,
					Parameters: []types.ProcessorParameter{
						{
							ParameterName:  types.ProcessorParameterNameLambdaArn,
							ParameterValue: aws.String("valueNotTested"),
						},
					},
				},
			},
		},
	}

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckIAMServiceLinkedRole(ctx, t, "/aws-service-role/es") },
		ErrorCheck:               acctest.ErrorCheck(t, names.FirehoseServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDeliveryStreamDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccDeliveryStreamConfig_elasticsearchVPCBasic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDeliveryStreamExists(ctx, resourceName, &stream),
					testAccCheckDeliveryStreamAttributes(&stream, nil, nil, nil, nil, nil, nil, nil),
					resource.TestCheckResourceAttrPair(resourceName, "elasticsearch_configuration.0.vpc_config.0.vpc_id", "aws_vpc.test", names.AttrID),
					resource.TestCheckResourceAttr(resourceName, "elasticsearch_configuration.0.vpc_config.0.subnet_ids.#", acctest.Ct2),
					resource.TestCheckResourceAttr(resourceName, "elasticsearch_configuration.0.vpc_config.0.security_group_ids.#", acctest.Ct2),
					resource.TestCheckResourceAttrPair(resourceName, "elasticsearch_configuration.0.vpc_config.0.role_arn", "aws_iam_role.firehose", names.AttrARN),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccDeliveryStreamConfig_elasticsearchVPCUpdate(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDeliveryStreamExists(ctx, resourceName, &stream),
					testAccCheckDeliveryStreamAttributes(&stream, nil, nil, nil, updatedElasticsearchConfig, nil, nil, nil),
					resource.TestCheckResourceAttrPair(resourceName, "elasticsearch_configuration.0.vpc_config.0.vpc_id", "aws_vpc.test", names.AttrID),
					resource.TestCheckResourceAttr(resourceName, "elasticsearch_configuration.0.vpc_config.0.subnet_ids.#", acctest.Ct2),
					resource.TestCheckResourceAttr(resourceName, "elasticsearch_configuration.0.vpc_config.0.security_group_ids.#", acctest.Ct2),
					resource.TestCheckResourceAttrPair(resourceName, "elasticsearch_configuration.0.vpc_config.0.role_arn", "aws_iam_role.firehose", names.AttrARN),
				),
			},
		},
	})
}

func TestAccFirehoseDeliveryStream_Elasticsearch_ErrorOutputPrefix(t *testing.T) {
	ctx := acctest.Context(t)
	var stream types.DeliveryStreamDescription
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_kinesis_firehose_delivery_stream.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckIAMServiceLinkedRole(ctx, t, "/aws-service-role/es") },
		ErrorCheck:               acctest.ErrorCheck(t, names.FirehoseServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDeliveryStreamDestroy_ExtendedS3(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccDeliveryStreamConfig_elasticsearchErrorOutputPrefix(rName, "prefix1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDeliveryStreamExists(ctx, resourceName, &stream),
					resource.TestCheckResourceAttr(resourceName, "elasticsearch_configuration.0.s3_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "elasticsearch_configuration.0.s3_configuration.0.error_output_prefix", "prefix1"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccDeliveryStreamConfig_elasticsearchErrorOutputPrefix(rName, "prefix2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDeliveryStreamExists(ctx, resourceName, &stream),
					resource.TestCheckResourceAttr(resourceName, "elasticsearch_configuration.0.s3_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "elasticsearch_configuration.0.s3_configuration.0.error_output_prefix", "prefix2"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccDeliveryStreamConfig_elasticsearchErrorOutputPrefix(rName, ""),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDeliveryStreamExists(ctx, resourceName, &stream),
					resource.TestCheckResourceAttr(resourceName, "elasticsearch_configuration.0.s3_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "elasticsearch_configuration.0.s3_configuration.0.error_output_prefix", ""),
				),
			},
		},
	})
}

func TestAccFirehoseDeliveryStream_openSearchUpdates(t *testing.T) {
	ctx := acctest.Context(t)
	var stream types.DeliveryStreamDescription
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_kinesis_firehose_delivery_stream.test"

	updatedOpensearchConfig := &types.AmazonopensearchserviceDestinationDescription{
		BufferingHints: &types.AmazonopensearchserviceBufferingHints{
			IntervalInSeconds: aws.Int32(500),
		},
		ProcessingConfiguration: &types.ProcessingConfiguration{
			Enabled: aws.Bool(true),
			Processors: []types.Processor{
				{
					Type: types.ProcessorTypeLambda,
					Parameters: []types.ProcessorParameter{
						{
							ParameterName:  types.ProcessorParameterNameLambdaArn,
							ParameterValue: aws.String("valueNotTested"),
						},
					},
				},
			},
		},
	}

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckIAMServiceLinkedRole(ctx, t, "/aws-service-role/opensearchservice")
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.FirehoseServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDeliveryStreamDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccDeliveryStreamConfig_opensearchBasic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDeliveryStreamExists(ctx, resourceName, &stream),
					testAccCheckDeliveryStreamAttributes(&stream, nil, nil, nil, nil, nil, nil, nil),
					resource.TestCheckResourceAttr(resourceName, "opensearch_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "opensearch_configuration.0.document_id_options.#", acctest.Ct0),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccDeliveryStreamConfig_opensearchUpdate(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDeliveryStreamExists(ctx, resourceName, &stream),
					testAccCheckDeliveryStreamAttributes(&stream, nil, nil, nil, nil, updatedOpensearchConfig, nil, nil),
				),
			},
		},
	})
}

func TestAccFirehoseDeliveryStream_openSearchEndpointUpdates(t *testing.T) {
	ctx := acctest.Context(t)
	var stream types.DeliveryStreamDescription
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_kinesis_firehose_delivery_stream.test"

	updatedOpensearchConfig := &types.AmazonopensearchserviceDestinationDescription{
		BufferingHints: &types.AmazonopensearchserviceBufferingHints{
			IntervalInSeconds: aws.Int32(500),
		},
		ProcessingConfiguration: &types.ProcessingConfiguration{
			Enabled: aws.Bool(true),
			Processors: []types.Processor{
				{
					Type: types.ProcessorTypeLambda,
					Parameters: []types.ProcessorParameter{
						{
							ParameterName:  types.ProcessorParameterNameLambdaArn,
							ParameterValue: aws.String("valueNotTested"),
						},
					},
				},
			},
		},
	}

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckIAMServiceLinkedRole(ctx, t, "/aws-service-role/opensearchservice")
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.FirehoseServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDeliveryStreamDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccDeliveryStreamConfig_opensearchEndpoint(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDeliveryStreamExists(ctx, resourceName, &stream),
					testAccCheckDeliveryStreamAttributes(&stream, nil, nil, nil, nil, nil, nil, nil),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccDeliveryStreamConfig_opensearchEndpointUpdate(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDeliveryStreamExists(ctx, resourceName, &stream),
					testAccCheckDeliveryStreamAttributes(&stream, nil, nil, nil, nil, updatedOpensearchConfig, nil, nil),
				),
			},
		},
	})
}

// This doesn't actually test updating VPC Configuration. It tests changing OpenSearch configuration
// when the Kinesis Firehose delivery stream has a VPC Configuration.
func TestAccFirehoseDeliveryStream_openSearchWithVPCUpdates(t *testing.T) {
	ctx := acctest.Context(t)
	var stream types.DeliveryStreamDescription
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_kinesis_firehose_delivery_stream.test"

	updatedOpensearchConfig := &types.AmazonopensearchserviceDestinationDescription{
		BufferingHints: &types.AmazonopensearchserviceBufferingHints{
			IntervalInSeconds: aws.Int32(500),
		},
		ProcessingConfiguration: &types.ProcessingConfiguration{
			Enabled: aws.Bool(true),
			Processors: []types.Processor{
				{
					Type: types.ProcessorTypeLambda,
					Parameters: []types.ProcessorParameter{
						{
							ParameterName:  types.ProcessorParameterNameLambdaArn,
							ParameterValue: aws.String("valueNotTested"),
						},
					},
				},
			},
		},
	}

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckIAMServiceLinkedRole(ctx, t, "/aws-service-role/opensearchservice")
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.FirehoseServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDeliveryStreamDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccDeliveryStreamConfig_opensearchVPCBasic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDeliveryStreamExists(ctx, resourceName, &stream),
					testAccCheckDeliveryStreamAttributes(&stream, nil, nil, nil, nil, nil, nil, nil),
					resource.TestCheckResourceAttrPair(resourceName, "opensearch_configuration.0.vpc_config.0.vpc_id", "aws_vpc.test", names.AttrID),
					resource.TestCheckResourceAttr(resourceName, "opensearch_configuration.0.vpc_config.0.subnet_ids.#", acctest.Ct2),
					resource.TestCheckResourceAttr(resourceName, "opensearch_configuration.0.vpc_config.0.security_group_ids.#", acctest.Ct2),
					resource.TestCheckResourceAttrPair(resourceName, "opensearch_configuration.0.vpc_config.0.role_arn", "aws_iam_role.firehose", names.AttrARN),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccDeliveryStreamConfig_opensearchVPCUpdate(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDeliveryStreamExists(ctx, resourceName, &stream),
					testAccCheckDeliveryStreamAttributes(&stream, nil, nil, nil, nil, updatedOpensearchConfig, nil, nil),
					resource.TestCheckResourceAttrPair(resourceName, "opensearch_configuration.0.vpc_config.0.vpc_id", "aws_vpc.test", names.AttrID),
					resource.TestCheckResourceAttr(resourceName, "opensearch_configuration.0.vpc_config.0.subnet_ids.#", acctest.Ct2),
					resource.TestCheckResourceAttr(resourceName, "opensearch_configuration.0.vpc_config.0.security_group_ids.#", acctest.Ct2),
					resource.TestCheckResourceAttrPair(resourceName, "opensearch_configuration.0.vpc_config.0.role_arn", "aws_iam_role.firehose", names.AttrARN),
				),
			},
		},
	})
}

func TestAccFirehoseDeliveryStream_Opensearch_ErrorOutputPrefix(t *testing.T) {
	ctx := acctest.Context(t)
	var stream types.DeliveryStreamDescription
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_kinesis_firehose_delivery_stream.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckIAMServiceLinkedRole(ctx, t, "/aws-service-role/opensearchservice")
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.FirehoseServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDeliveryStreamDestroy_ExtendedS3(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccDeliveryStreamConfig_opensearchErrorOutputPrefix(rName, "prefix1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDeliveryStreamExists(ctx, resourceName, &stream),
					resource.TestCheckResourceAttr(resourceName, "opensearch_configuration.0.s3_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "opensearch_configuration.0.s3_configuration.0.error_output_prefix", "prefix1"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccDeliveryStreamConfig_opensearchErrorOutputPrefix(rName, "prefix2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDeliveryStreamExists(ctx, resourceName, &stream),
					resource.TestCheckResourceAttr(resourceName, "opensearch_configuration.0.s3_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "opensearch_configuration.0.s3_configuration.0.error_output_prefix", "prefix2"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccDeliveryStreamConfig_opensearchErrorOutputPrefix(rName, ""),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDeliveryStreamExists(ctx, resourceName, &stream),
					resource.TestCheckResourceAttr(resourceName, "opensearch_configuration.0.s3_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "opensearch_configuration.0.s3_configuration.0.error_output_prefix", ""),
				),
			},
		},
	})
}

func TestAccFirehoseDeliveryStream_openSearchServerlessUpdates(t *testing.T) {
	ctx := acctest.Context(t)
	var stream types.DeliveryStreamDescription
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_kinesis_firehose_delivery_stream.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.OpenSearchServerlessEndpointID)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.FirehoseServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDeliveryStreamDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccDeliveryStreamConfig_openSearchServerlessBasic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckDeliveryStreamExists(ctx, resourceName, &stream),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, names.AttrDestination, "opensearchserverless"),
					resource.TestCheckResourceAttrSet(resourceName, "destination_id"),
					resource.TestCheckResourceAttr(resourceName, "elasticsearch_configuration.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "extended_s3_configuration.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "http_endpoint_configuration.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "kinesis_source_configuration.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, "opensearch_configuration.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "opensearchserverless_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "opensearchserverless_configuration.0.buffering_interval", "300"),
					resource.TestCheckResourceAttr(resourceName, "opensearchserverless_configuration.0.buffering_size", "5"),
					resource.TestCheckResourceAttr(resourceName, "opensearchserverless_configuration.0.cloudwatch_logging_options.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "opensearchserverless_configuration.0.cloudwatch_logging_options.0.enabled", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "opensearchserverless_configuration.0.cloudwatch_logging_options.0.log_group_name", ""),
					resource.TestCheckResourceAttr(resourceName, "opensearchserverless_configuration.0.cloudwatch_logging_options.0.log_stream_name", ""),
					resource.TestCheckResourceAttrSet(resourceName, "opensearchserverless_configuration.0.collection_endpoint"),
					resource.TestCheckResourceAttr(resourceName, "opensearchserverless_configuration.0.index_name", "test"),
					resource.TestCheckResourceAttr(resourceName, "opensearchserverless_configuration.0.processing_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "opensearchserverless_configuration.0.processing_configuration.0.enabled", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "opensearchserverless_configuration.0.processing_configuration.0.processors.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "opensearchserverless_configuration.0.retry_duration", "300"),
					resource.TestCheckResourceAttrSet(resourceName, "opensearchserverless_configuration.0.role_arn"),
					resource.TestCheckResourceAttr(resourceName, "opensearchserverless_configuration.0.s3_backup_mode", "FailedDocumentsOnly"),
					resource.TestCheckResourceAttr(resourceName, "opensearchserverless_configuration.0.s3_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttrSet(resourceName, "opensearchserverless_configuration.0.s3_configuration.0.bucket_arn"),
					resource.TestCheckResourceAttr(resourceName, "opensearchserverless_configuration.0.s3_configuration.0.buffering_interval", "300"),
					resource.TestCheckResourceAttr(resourceName, "opensearchserverless_configuration.0.s3_configuration.0.buffering_size", "5"),
					resource.TestCheckResourceAttr(resourceName, "opensearchserverless_configuration.0.s3_configuration.0.cloudwatch_logging_options.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "opensearchserverless_configuration.0.s3_configuration.0.cloudwatch_logging_options.0.enabled", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "opensearchserverless_configuration.0.s3_configuration.0.cloudwatch_logging_options.0.log_group_name", ""),
					resource.TestCheckResourceAttr(resourceName, "opensearchserverless_configuration.0.s3_configuration.0.cloudwatch_logging_options.0.log_stream_name", ""),
					resource.TestCheckResourceAttr(resourceName, "opensearchserverless_configuration.0.s3_configuration.0.compression_format", "UNCOMPRESSED"),
					resource.TestCheckResourceAttr(resourceName, "opensearchserverless_configuration.0.s3_configuration.0.error_output_prefix", ""),
					resource.TestCheckResourceAttr(resourceName, "opensearchserverless_configuration.0.s3_configuration.0.kms_key_arn", ""),
					resource.TestCheckResourceAttr(resourceName, "opensearchserverless_configuration.0.s3_configuration.0.prefix", ""),
					resource.TestCheckResourceAttrSet(resourceName, "opensearchserverless_configuration.0.s3_configuration.0.role_arn"),
					resource.TestCheckResourceAttr(resourceName, "opensearchserverless_configuration.0.vpc_config.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "redshift_configuration.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "server_side_encryption.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "server_side_encryption.0.enabled", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "server_side_encryption.0.key_arn", ""),
					resource.TestCheckResourceAttr(resourceName, "server_side_encryption.0.key_type", "AWS_OWNED_CMK"),
					resource.TestCheckResourceAttr(resourceName, "snowflake_configuration.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "splunk_configuration.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct0),
					resource.TestCheckResourceAttrSet(resourceName, "version_id"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccDeliveryStreamConfig_openSearchServerlessUpdate(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckDeliveryStreamExists(ctx, resourceName, &stream),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, names.AttrDestination, "opensearchserverless"),
					resource.TestCheckResourceAttrSet(resourceName, "destination_id"),
					resource.TestCheckResourceAttr(resourceName, "elasticsearch_configuration.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "extended_s3_configuration.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "http_endpoint_configuration.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "kinesis_source_configuration.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "msk_source_configuration.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, "opensearch_configuration.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "opensearchserverless_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "opensearchserverless_configuration.0.buffering_interval", "500"),
					resource.TestCheckResourceAttr(resourceName, "opensearchserverless_configuration.0.buffering_size", acctest.Ct10),
					resource.TestCheckResourceAttr(resourceName, "opensearchserverless_configuration.0.cloudwatch_logging_options.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "opensearchserverless_configuration.0.cloudwatch_logging_options.0.enabled", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "opensearchserverless_configuration.0.cloudwatch_logging_options.0.log_group_name", ""),
					resource.TestCheckResourceAttr(resourceName, "opensearchserverless_configuration.0.cloudwatch_logging_options.0.log_stream_name", ""),
					resource.TestCheckResourceAttrSet(resourceName, "opensearchserverless_configuration.0.collection_endpoint"),
					resource.TestCheckResourceAttr(resourceName, "opensearchserverless_configuration.0.index_name", "test"),
					resource.TestCheckResourceAttr(resourceName, "opensearchserverless_configuration.0.processing_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "opensearchserverless_configuration.0.processing_configuration.0.enabled", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "opensearchserverless_configuration.0.processing_configuration.0.processors.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "opensearchserverless_configuration.0.processing_configuration.0.processors.0.type", "Lambda"),
					resource.TestCheckResourceAttr(resourceName, "opensearchserverless_configuration.0.processing_configuration.0.processors.0.parameters.#", acctest.Ct3),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "opensearchserverless_configuration.0.processing_configuration.0.processors.0.parameters.*", map[string]string{
						"parameter_name": "LambdaArn",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "opensearchserverless_configuration.0.processing_configuration.0.processors.0.parameters.*", map[string]string{
						"parameter_name":  "BufferSizeInMBs",
						"parameter_value": "1.1",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "opensearchserverless_configuration.0.processing_configuration.0.processors.0.parameters.*", map[string]string{
						"parameter_name":  "BufferIntervalInSeconds",
						"parameter_value": "70",
					}),
					resource.TestCheckResourceAttr(resourceName, "opensearchserverless_configuration.0.retry_duration", "300"),
					resource.TestCheckResourceAttrSet(resourceName, "opensearchserverless_configuration.0.role_arn"),
					resource.TestCheckResourceAttr(resourceName, "opensearchserverless_configuration.0.s3_backup_mode", "FailedDocumentsOnly"),
					resource.TestCheckResourceAttr(resourceName, "opensearchserverless_configuration.0.s3_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttrSet(resourceName, "opensearchserverless_configuration.0.s3_configuration.0.bucket_arn"),
					resource.TestCheckResourceAttr(resourceName, "opensearchserverless_configuration.0.s3_configuration.0.buffering_interval", "300"),
					resource.TestCheckResourceAttr(resourceName, "opensearchserverless_configuration.0.s3_configuration.0.buffering_size", "5"),
					resource.TestCheckResourceAttr(resourceName, "opensearchserverless_configuration.0.s3_configuration.0.cloudwatch_logging_options.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "opensearchserverless_configuration.0.s3_configuration.0.cloudwatch_logging_options.0.enabled", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "opensearchserverless_configuration.0.s3_configuration.0.cloudwatch_logging_options.0.log_group_name", ""),
					resource.TestCheckResourceAttr(resourceName, "opensearchserverless_configuration.0.s3_configuration.0.cloudwatch_logging_options.0.log_stream_name", ""),
					resource.TestCheckResourceAttr(resourceName, "opensearchserverless_configuration.0.s3_configuration.0.compression_format", "UNCOMPRESSED"),
					resource.TestCheckResourceAttr(resourceName, "opensearchserverless_configuration.0.s3_configuration.0.error_output_prefix", ""),
					resource.TestCheckResourceAttr(resourceName, "opensearchserverless_configuration.0.s3_configuration.0.kms_key_arn", ""),
					resource.TestCheckResourceAttr(resourceName, "opensearchserverless_configuration.0.s3_configuration.0.prefix", ""),
					resource.TestCheckResourceAttrSet(resourceName, "opensearchserverless_configuration.0.s3_configuration.0.role_arn"),
					resource.TestCheckResourceAttr(resourceName, "opensearchserverless_configuration.0.vpc_config.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "redshift_configuration.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "server_side_encryption.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "server_side_encryption.0.enabled", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "server_side_encryption.0.key_arn", ""),
					resource.TestCheckResourceAttr(resourceName, "server_side_encryption.0.key_type", "AWS_OWNED_CMK"),
					resource.TestCheckResourceAttr(resourceName, "snowflake_configuration.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "splunk_configuration.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct0),
					resource.TestCheckResourceAttrSet(resourceName, "version_id"),
				),
			},
		},
	})
}

// Regression test for https://github.com/hashicorp/terraform-provider-aws/issues/1657
func TestAccFirehoseDeliveryStream_missingProcessing(t *testing.T) {
	ctx := acctest.Context(t)
	var stream types.DeliveryStreamDescription
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_kinesis_firehose_delivery_stream.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.FirehoseServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDeliveryStreamDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccDeliveryStreamConfig_missingProcessingConfiguration(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDeliveryStreamExists(ctx, resourceName, &stream),
					testAccCheckDeliveryStreamAttributes(&stream, nil, nil, nil, nil, nil, nil, nil),
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

func testAccCheckDeliveryStreamExists(ctx context.Context, n string, v *types.DeliveryStreamDescription) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).FirehoseClient(ctx)

		output, err := tffirehose.FindDeliveryStreamByName(ctx, conn, rs.Primary.Attributes[names.AttrName])

		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

func testAccCheckDeliveryStreamDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_kinesis_firehose_delivery_stream" {
				continue
			}

			conn := acctest.Provider.Meta().(*conns.AWSClient).FirehoseClient(ctx)

			_, err := tffirehose.FindDeliveryStreamByName(ctx, conn, rs.Primary.Attributes[names.AttrName])

			if tfresource.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("Kinesis Firehose Delivery Stream %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccCheckDeliveryStreamAttributes(stream *types.DeliveryStreamDescription, s3config interface{}, extendedS3config interface{}, redshiftConfig interface{}, elasticsearchConfig interface{}, opensearchConfig interface{}, splunkConfig interface{}, httpEndpointConfig interface{}) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if !strings.HasPrefix(*stream.DeliveryStreamName, "terraform-kinesis-firehose") && !strings.HasPrefix(*stream.DeliveryStreamName, acctest.ResourcePrefix) {
			return fmt.Errorf("Bad Stream name: %s", *stream.DeliveryStreamName)
		}
		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_kinesis_firehose_delivery_stream" {
				continue
			}
			if *stream.DeliveryStreamARN != rs.Primary.Attributes[names.AttrARN] {
				return fmt.Errorf("Bad Delivery Stream ARN\n\t expected: %s\n\tgot: %s\n", rs.Primary.Attributes[names.AttrARN], *stream.DeliveryStreamARN)
			}

			if s3config != nil {
				s := s3config.(*types.S3DestinationDescription)
				// Range over the Stream Destinations, looking for the matching S3
				// destination. For simplicity, our test only have a single S3 or
				// Redshift destination, so at this time it's safe to match on the first
				// one
				var match bool
				for _, d := range stream.Destinations {
					if d.S3DestinationDescription != nil {
						if *d.S3DestinationDescription.BufferingHints.SizeInMBs == *s.BufferingHints.SizeInMBs {
							match = true
						}
					}
				}
				if !match {
					return fmt.Errorf("Mismatch s3 buffer size, expected: %v, got: %v", s, stream.Destinations)
				}
			}

			if extendedS3config != nil {
				es := extendedS3config.(*types.ExtendedS3DestinationDescription)

				// Range over the Stream Destinations, looking for the matching S3
				// destination. For simplicity, our test only have a single S3 or
				// Redshift destination, so at this time it's safe to match on the first
				// one
				var match, processingConfigMatch, matchS3BackupMode bool
				for _, d := range stream.Destinations {
					if d.ExtendedS3DestinationDescription != nil {
						if *d.ExtendedS3DestinationDescription.BufferingHints.SizeInMBs == *es.BufferingHints.SizeInMBs {
							match = true
						}
						if d.ExtendedS3DestinationDescription.S3BackupMode == es.S3BackupMode {
							matchS3BackupMode = true
						}

						processingConfigMatch = len(es.ProcessingConfiguration.Processors) == len(d.ExtendedS3DestinationDescription.ProcessingConfiguration.Processors)
					}
				}
				if !match {
					return fmt.Errorf("Mismatch extended s3 buffer size, expected: %v, got: %v", es, stream.Destinations)
				}
				if !processingConfigMatch {
					return fmt.Errorf("Mismatch extended s3 ProcessingConfiguration.Processors count, expected: %v, got: %v", es, stream.Destinations)
				}
				if !matchS3BackupMode {
					return fmt.Errorf("Mismatch extended s3 S3BackupMode, expected: %v, got: %v", es, stream.Destinations)
				}
			}

			if redshiftConfig != nil {
				r := redshiftConfig.(*types.RedshiftDestinationDescription)
				// Range over the Stream Destinations, looking for the matching Redshift
				// destination
				var matchCopyOptions, matchS3BackupMode, processingConfigMatch bool
				for _, d := range stream.Destinations {
					if d.RedshiftDestinationDescription != nil {
						if *d.RedshiftDestinationDescription.CopyCommand.CopyOptions == *r.CopyCommand.CopyOptions {
							matchCopyOptions = true
						}
						if d.RedshiftDestinationDescription.S3BackupMode == r.S3BackupMode {
							matchS3BackupMode = true
						}
						if r.ProcessingConfiguration != nil && d.RedshiftDestinationDescription.ProcessingConfiguration != nil {
							processingConfigMatch = len(r.ProcessingConfiguration.Processors) == len(d.RedshiftDestinationDescription.ProcessingConfiguration.Processors)
						}
					}
				}
				if !matchCopyOptions || !matchS3BackupMode {
					return fmt.Errorf("Mismatch Redshift CopyOptions or S3BackupMode, expected: %v, got: %v", r, stream.Destinations)
				}
				if !processingConfigMatch {
					return fmt.Errorf("Mismatch Redshift ProcessingConfiguration.Processors count, expected: %v, got: %v", r, stream.Destinations)
				}
			}

			if elasticsearchConfig != nil {
				es := elasticsearchConfig.(*types.ElasticsearchDestinationDescription)
				// Range over the Stream Destinations, looking for the matching Elasticsearch destination
				var match, processingConfigMatch bool
				for _, d := range stream.Destinations {
					if d.ElasticsearchDestinationDescription != nil {
						match = true
						if es.ProcessingConfiguration != nil && d.ElasticsearchDestinationDescription.ProcessingConfiguration != nil {
							processingConfigMatch = len(es.ProcessingConfiguration.Processors) == len(d.ElasticsearchDestinationDescription.ProcessingConfiguration.Processors)
						}
					}
				}
				if !match {
					return fmt.Errorf("Mismatch Elasticsearch Buffering Interval, expected: %v, got: %v", es, stream.Destinations)
				}
				if !processingConfigMatch {
					return fmt.Errorf("Mismatch Elasticsearch ProcessingConfiguration.Processors count, expected: %v, got: %v", es, stream.Destinations)
				}
			}

			if opensearchConfig != nil {
				es := opensearchConfig.(*types.AmazonopensearchserviceDestinationDescription)
				// Range over the Stream Destinations, looking for the matching Opensearch destination
				var match, processingConfigMatch bool
				for _, d := range stream.Destinations {
					if d.AmazonopensearchserviceDestinationDescription != nil {
						match = true
						if es.ProcessingConfiguration != nil && d.AmazonopensearchserviceDestinationDescription.ProcessingConfiguration != nil {
							processingConfigMatch = len(es.ProcessingConfiguration.Processors) == len(d.AmazonopensearchserviceDestinationDescription.ProcessingConfiguration.Processors)
						}
					}
				}
				if !match {
					return fmt.Errorf("Mismatch Opensearch Buffering Interval, expected: %v, got: %v", es, stream.Destinations)
				}
				if !processingConfigMatch {
					return fmt.Errorf("Mismatch Opensearch ProcessingConfiguration.Processors count, expected: %v, got: %v", es, stream.Destinations)
				}
			}

			if splunkConfig != nil {
				s := splunkConfig.(*types.SplunkDestinationDescription)
				// Range over the Stream Destinations, looking for the matching Splunk destination
				var match, matchHECEndpointType, matchHECAcknowledgmentTimeoutInSeconds, matchS3BackupMode, processingConfigMatch bool
				for _, d := range stream.Destinations {
					if d.SplunkDestinationDescription != nil {
						if *d.SplunkDestinationDescription.BufferingHints.SizeInMBs == *s.BufferingHints.SizeInMBs {
							match = true
						}
						if d.SplunkDestinationDescription.HECEndpointType == s.HECEndpointType {
							matchHECEndpointType = true
						}
						if *d.SplunkDestinationDescription.HECAcknowledgmentTimeoutInSeconds == *s.HECAcknowledgmentTimeoutInSeconds {
							matchHECAcknowledgmentTimeoutInSeconds = true
						}
						if d.SplunkDestinationDescription.S3BackupMode == s.S3BackupMode {
							matchS3BackupMode = true
						}
						if s.ProcessingConfiguration != nil && d.SplunkDestinationDescription.ProcessingConfiguration != nil {
							processingConfigMatch = len(s.ProcessingConfiguration.Processors) == len(d.SplunkDestinationDescription.ProcessingConfiguration.Processors)
						}
					}
				}
				if !match || !matchHECEndpointType || !matchHECAcknowledgmentTimeoutInSeconds || !matchS3BackupMode {
					return fmt.Errorf("Mismatch Splunk BufferingHints.SizeInMBs or HECEndpointType or HECAcknowledgmentTimeoutInSeconds or S3BackupMode, expected: %v, got: %v", s, stream.Destinations)
				}
				if !processingConfigMatch {
					return fmt.Errorf("Mismatch extended splunk ProcessingConfiguration.Processors count, expected: %v, got: %v", s, stream.Destinations)
				}
			}

			if httpEndpointConfig != nil {
				s := httpEndpointConfig.(*types.HttpEndpointDestinationDescription)
				// Range over the Stream Destinations, looking for the matching HttpEndpoint destination
				var matchS3BackupMode, matchUrl, matchName, processingConfigMatch bool
				for _, d := range stream.Destinations {
					if d.HttpEndpointDestinationDescription != nil {
						if d.HttpEndpointDestinationDescription.S3BackupMode == s.S3BackupMode {
							matchS3BackupMode = true
						}
						if *d.HttpEndpointDestinationDescription.EndpointConfiguration.Url == *s.EndpointConfiguration.Url {
							matchUrl = true
						}
						if *d.HttpEndpointDestinationDescription.EndpointConfiguration.Name == *s.EndpointConfiguration.Name {
							matchName = true
						}
						if s.ProcessingConfiguration != nil && d.HttpEndpointDestinationDescription.ProcessingConfiguration != nil {
							processingConfigMatch = len(s.ProcessingConfiguration.Processors) == len(d.HttpEndpointDestinationDescription.ProcessingConfiguration.Processors)
						}
					}
				}
				if !matchS3BackupMode {
					return fmt.Errorf("Mismatch HTTP Endpoint S3BackupMode, expected: %v, got: %v", s, stream.Destinations)
				}
				if !matchUrl || !matchName {
					return fmt.Errorf("Mismatch HTTP Endpoint EndpointConfiguration, expected: %v, got: %v", s, stream.Destinations)
				}
				if !processingConfigMatch {
					return fmt.Errorf("Mismatch HTTP Endpoint ProcessingConfiguration.Processors count, expected: %v, got: %v", s, stream.Destinations)
				}
			}
		}
		return nil
	}
}

func testAccCheckDeliveryStreamDestroy_ExtendedS3(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		err := testAccCheckDeliveryStreamDestroy(ctx)(s)

		if err == nil {
			err = testAccCheckLambdaFunctionDestroy(ctx)(s)
		}

		return err
	}
}

func testAccCheckLambdaFunctionDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).LambdaClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_lambda_function" {
				continue
			}

			_, err := conn.GetFunction(ctx, &lambda.GetFunctionInput{
				FunctionName: aws.String(rs.Primary.ID),
			})

			if err == nil {
				return fmt.Errorf("Lambda Function still exists")
			}
		}

		return nil
	}
}

func testAccDeliveryStreamConfig_base(rName string) string {
	return fmt.Sprintf(`
data "aws_caller_identity" "current" {}
data "aws_partition" "current" {}
data "aws_region" "current" {}

resource "aws_iam_role" "firehose" {
  name = %[1]q

  assume_role_policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Sid": "",
      "Effect": "Allow",
      "Principal": {
        "Service": "firehose.amazonaws.com"
      },
      "Action": "sts:AssumeRole",
      "Condition": {
        "StringEquals": {
          "sts:ExternalId": "${data.aws_caller_identity.current.account_id}"
        }
      }
    }
  ]
}
EOF
}

resource "aws_s3_bucket" "bucket" {
  bucket = %[1]q
}

resource "aws_iam_role_policy" "firehose" {
  name = %[1]q
  role = aws_iam_role.firehose.id

  policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Sid": "",
      "Effect": "Allow",
      "Action": [
        "s3:AbortMultipartUpload",
        "s3:GetBucketLocation",
        "s3:GetObject",
        "s3:ListBucket",
        "s3:ListBucketMultipartUploads",
        "s3:PutObject"
      ],
      "Resource": [
        "${aws_s3_bucket.bucket.arn}",
        "${aws_s3_bucket.bucket.arn}/*"
      ]
    },
    {
      "Sid": "GlueAccess",
      "Effect": "Allow",
      "Action": [
        "glue:GetTable",
        "glue:GetTableVersion",
        "glue:GetTableVersions"
      ],
      "Resource": [
        "*"
      ]
    },
    {
      "Sid": "LakeFormationDataAccess",
      "Effect": "Allow",
      "Action": [
        "lakeformation:GetDataAccess"
      ],
      "Resource": "*"
    }
  ]
}
EOF
}
`, rName)
}

func testAccDeliveryStreamConfig_baseLambda(rName string) string {
	return fmt.Sprintf(`
resource "aws_iam_role_policy" "iam_policy_for_lambda" {
  name = "%[1]s-lambda"
  role = aws_iam_role.iam_for_lambda.id

  policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Action": [
        "logs:CreateLogGroup",
        "logs:CreateLogStream",
        "logs:PutLogEvents"
      ],
      "Resource": "arn:${data.aws_partition.current.partition}:logs:*:*:*"
    },
    {
      "Effect": "Allow",
      "Action": [
        "xray:PutTraceSegments"
      ],
      "Resource": [
        "*"
      ]
    }
  ]
}
EOF
}

resource "aws_iam_role" "iam_for_lambda" {
  name = "%[1]s-lambda"

  assume_role_policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Action": "sts:AssumeRole",
      "Principal": {
        "Service": "lambda.amazonaws.com"
      },
      "Effect": "Allow",
      "Sid": ""
    }
  ]
}
EOF
}

resource "aws_lambda_function" "lambda_function_test" {
  filename      = "test-fixtures/lambdatest.zip"
  function_name = %[1]q
  role          = aws_iam_role.iam_for_lambda.arn
  handler       = "exports.example"
  runtime       = "nodejs16.x"
}
`, rName)
}

func testAccDeliveryStreamConfig_baseKinesisStreamSource(rName string) string {
	return fmt.Sprintf(`
resource "aws_kinesis_stream" "source" {
  name        = %[1]q
  shard_count = 1
}

resource "aws_iam_role" "kinesis_source" {
  name = "%[1]s-stream"

  assume_role_policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Sid": "",
      "Effect": "Allow",
      "Principal": {
        "Service": "firehose.amazonaws.com"
      },
      "Action": "sts:AssumeRole",
      "Condition": {
        "StringEquals": {
          "sts:ExternalId": "${data.aws_caller_identity.current.account_id}"
        }
      }
    }
  ]
}
EOF
}

resource "aws_iam_role_policy" "kinesis_source" {
  name = "%[1]s-stream"
  role = aws_iam_role.kinesis_source.id

  policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Sid": "",
      "Effect": "Allow",
      "Action": [
        "kinesis:DescribeStream",
        "kinesis:GetShardIterator",
        "kinesis:GetRecords"
      ],
      "Resource": [
        "${aws_kinesis_stream.source.arn}"
      ]
    }
  ]
}
EOF
}
`, rName)
}

func testAccDeliveryStreamConfig_baseMSKClusterSource(rName string) string {
	return acctest.ConfigCompose(acctest.ConfigAvailableAZsNoOptIn(), fmt.Sprintf(`
resource "aws_vpc" "test" {
  cidr_block = "10.0.0.0/16"

  enable_dns_hostnames = true

  tags = {
    Name = %[1]q
  }
}

resource "aws_subnet" "test" {
  count = 2

  vpc_id            = aws_vpc.test.id
  availability_zone = data.aws_availability_zones.available.names[count.index]
  cidr_block        = cidrsubnet(aws_vpc.test.cidr_block, 8, count.index)

  tags = {
    Name = %[1]q
  }
}

resource "aws_msk_serverless_cluster" "test" {
  cluster_name = %[1]q

  client_authentication {
    sasl {
      iam {
        enabled = true
      }
    }
  }

  vpc_config {
    subnet_ids = aws_subnet.test[*].id
  }
}

resource "aws_msk_cluster_policy" "test" {
  cluster_arn = aws_msk_serverless_cluster.test.arn

  policy = jsonencode({
    Version = "2012-10-17",
    Statement = [{
      Sid    = "testMskClusterPolicy"
      Effect = "Allow"
      Principal = {
        "Service" = "firehose.amazonaws.com"
      }
      Action = [
        "kafka:CreateVpcConnection",
      ]
      Resource = aws_msk_serverless_cluster.test.arn
    }]
  })
}

resource "aws_iam_role" "msk_source" {
  name = "%[1]s-cluster"

  assume_role_policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Sid": "",
      "Effect": "Allow",
      "Principal": {
        "Service": "firehose.amazonaws.com"
      },
      "Action": "sts:AssumeRole",
      "Condition": {
        "StringEquals": {
          "sts:ExternalId": "${data.aws_caller_identity.current.account_id}"
        }
      }
    }
  ]
}
EOF
}

resource "aws_iam_role_policy" "msk_source" {
  name = "%[1]s-cluster"
  role = aws_iam_role.msk_source.id

  policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Sid": "",
      "Effect": "Allow",
      "Action": [
        "kafka:GetBootstrapBrokers",
        "kafka:DescribeCluster",
        "kafka:DescribeClusterV2",
        "kafka-cluster:Connect"
      ],
      "Resource": [
        "${aws_msk_serverless_cluster.test.arn}"
      ]
    }
  ]
}
EOF
}
`, rName))
}

func testAccDeliveryStreamConfig_s3CloudWatchLogging(rName string) string {
	return fmt.Sprintf(`
data "aws_caller_identity" "current" {}
data "aws_partition" "current" {}

resource "aws_iam_role" "firehose" {
  name = %[1]q

  assume_role_policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Sid": "",
      "Effect": "Allow",
      "Principal": {
        "Service": "firehose.amazonaws.com"
      },
      "Action": "sts:AssumeRole",
      "Condition": {
        "StringEquals": {
          "sts:ExternalId": "${data.aws_caller_identity.current.account_id}"
        }
      }
    }
  ]
}
EOF
}

resource "aws_iam_role_policy" "firehose" {
  name = %[1]q
  role = aws_iam_role.firehose.id

  policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Sid": "",
      "Effect": "Allow",
      "Action": [
        "s3:AbortMultipartUpload",
        "s3:GetBucketLocation",
        "s3:GetObject",
        "s3:ListBucket",
        "s3:ListBucketMultipartUploads",
        "s3:PutObject"
      ],
      "Resource": [
        "${aws_s3_bucket.bucket.arn}",
        "${aws_s3_bucket.bucket.arn}/*"
      ]
    },
    {
      "Effect": "Allow",
      "Action": [
        "logs:putLogEvents"
      ],
      "Resource": [
        "arn:${data.aws_partition.current.partition}:logs::log-group:/aws/kinesisfirehose/*"
      ]
    }
  ]
}
EOF
}

resource "aws_s3_bucket" "bucket" {
  bucket = %[1]q
}

resource "aws_cloudwatch_log_group" "test" {
  name = %[1]q
}

resource "aws_cloudwatch_log_stream" "test" {
  name           = %[1]q
  log_group_name = aws_cloudwatch_log_group.test.name
}

resource "aws_kinesis_firehose_delivery_stream" "test" {
  depends_on  = [aws_iam_role_policy.firehose]
  name        = %[1]q
  destination = "extended_s3"

  extended_s3_configuration {
    role_arn   = aws_iam_role.firehose.arn
    bucket_arn = aws_s3_bucket.bucket.arn

    cloudwatch_logging_options {
      enabled         = true
      log_group_name  = aws_cloudwatch_log_group.test.name
      log_stream_name = aws_cloudwatch_log_stream.test.name
    }
  }
}
`, rName)
}

func testAccDeliveryStreamConfig_baseSecretsManager(rName, privateKey string) string {
	return fmt.Sprintf(`
data "aws_caller_identity" "current_for_secretsmanager" {}

resource "aws_iam_role" "iam_for_secretsmanager" {
  name = "%[1]s-secretsmanager"

  assume_role_policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Sid": "",
      "Effect": "Allow",
      "Principal": {
        "Service": "firehose.amazonaws.com"
      },
      "Action": "sts:AssumeRole",
      "Condition": {
        "StringEquals": {
          "sts:ExternalId": "${data.aws_caller_identity.current_for_secretsmanager.account_id}"
        }
      }
    }
  ]
}
EOF
}

resource "aws_secretsmanager_secret" "test" {
  name = %[1]q
}

resource "aws_secretsmanager_secret_version" "test" {
  secret_id = aws_secretsmanager_secret.test.id
  secret_string = jsonencode({
    user        = "%[1]s"
    private_key = "%[2]s"
  })
}


resource "aws_iam_role_policy" "iam_policy_for_secretsmanager" {
  name = "%[1]s-secretsmanager"
  role = aws_iam_role.iam_for_secretsmanager.id

  policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Sid": "",
      "Effect": "Allow",
      "Action": [
        "secret:GetSecretValue",
        "secret:ListSecretVersionIds"
      ],
      "Resource": [
        "${aws_secretsmanager_secret.test.arn}"
      ]
    }
  ]
}
EOF
}
`, rName, acctest.TLSPEMRemoveRSAPrivateKeyEncapsulationBoundaries(acctest.TLSPEMRemoveNewlines(privateKey)))
}

func testAccDeliveryStreamConfig_tags1(rName, tagKey1, tagValue1 string) string {
	return acctest.ConfigCompose(testAccDeliveryStreamConfig_base(rName), fmt.Sprintf(`
resource "aws_kinesis_firehose_delivery_stream" "test" {
  depends_on  = [aws_iam_role_policy.firehose]
  name        = %[1]q
  destination = "extended_s3"

  extended_s3_configuration {
    role_arn   = aws_iam_role.firehose.arn
    bucket_arn = aws_s3_bucket.bucket.arn
  }

  tags = {
    %[2]q = %[3]q
  }
}
`, rName, tagKey1, tagValue1))
}

func testAccDeliveryStreamConfig_tags2(rName, tagKey1, tagValue1, tagKey2, tagValue2 string) string {
	return acctest.ConfigCompose(testAccDeliveryStreamConfig_base(rName), fmt.Sprintf(`
resource "aws_kinesis_firehose_delivery_stream" "test" {
  depends_on  = [aws_iam_role_policy.firehose]
  name        = %[1]q
  destination = "extended_s3"

  extended_s3_configuration {
    role_arn   = aws_iam_role.firehose.arn
    bucket_arn = aws_s3_bucket.bucket.arn
  }

  tags = {
    %[2]q = %[3]q
    %[4]q = %[5]q
  }
}
`, rName, tagKey1, tagValue1, tagKey2, tagValue2))
}

func testAccDeliveryStreamConfig_extendedS3basic(rName string) string {
	return acctest.ConfigCompose(
		testAccDeliveryStreamConfig_baseLambda(rName),
		testAccDeliveryStreamConfig_base(rName),
		fmt.Sprintf(`
resource "aws_kinesis_firehose_delivery_stream" "test" {
  depends_on  = [aws_iam_role_policy.firehose]
  name        = %[1]q
  destination = "extended_s3"

  extended_s3_configuration {
    role_arn   = aws_iam_role.firehose.arn
    bucket_arn = aws_s3_bucket.bucket.arn

    processing_configuration {
      enabled = true

      processors {
        type = "Lambda"

        parameters {
          parameter_name  = "LambdaArn"
          parameter_value = "${aws_lambda_function.lambda_function_test.arn}:$LATEST"
        }
        parameters {
          parameter_name  = "BufferSizeInMBs"
          parameter_value = "1.1"
        }
        parameters {
          parameter_name  = "BufferIntervalInSeconds"
          parameter_value = "70"
        }
      }
    }

    s3_backup_mode = "Disabled"
  }
}
`, rName))
}

func testAccDeliveryStreamConfig_extendedS3KinesisSource(rName string) string {
	return acctest.ConfigCompose(
		testAccDeliveryStreamConfig_base(rName),
		testAccDeliveryStreamConfig_baseKinesisStreamSource(rName),
		fmt.Sprintf(`
resource "aws_kinesis_firehose_delivery_stream" "test" {
  depends_on = [aws_iam_role_policy.firehose, aws_iam_role_policy.kinesis_source]
  name       = %[1]q

  kinesis_source_configuration {
    kinesis_stream_arn = aws_kinesis_stream.source.arn
    role_arn           = aws_iam_role.kinesis_source.arn
  }

  destination = "extended_s3"

  extended_s3_configuration {
    role_arn   = aws_iam_role.firehose.arn
    bucket_arn = aws_s3_bucket.bucket.arn
  }
}
`, rName))
}

func testAccDeliveryStreamConfig_extendedS3MSKSource(rName string) string {
	return acctest.ConfigCompose(
		testAccDeliveryStreamConfig_base(rName),
		testAccDeliveryStreamConfig_baseMSKClusterSource(rName),
		fmt.Sprintf(`
resource "aws_kinesis_firehose_delivery_stream" "test" {
  depends_on = [aws_iam_role_policy.firehose, aws_iam_role_policy.msk_source, aws_msk_cluster_policy.test]
  name       = %[1]q

  msk_source_configuration {
    authentication_configuration {
      connectivity = "PRIVATE"
      role_arn     = aws_iam_role.msk_source.arn
    }

    msk_cluster_arn = aws_msk_serverless_cluster.test.arn
    topic_name      = "test"
  }

  destination = "extended_s3"

  extended_s3_configuration {
    role_arn   = aws_iam_role.firehose.arn
    bucket_arn = aws_s3_bucket.bucket.arn
  }
}
`, rName))
}

func testAccDeliveryStreamConfig_extendedS3DataFormatConversionConfigurationEnabled(rName string, enabled bool) string {
	return acctest.ConfigCompose(testAccDeliveryStreamConfig_base(rName), fmt.Sprintf(`
resource "aws_glue_catalog_database" "test" {
  name = %[1]q
}

resource "aws_glue_catalog_table" "test" {
  database_name = aws_glue_catalog_database.test.name
  name          = %[1]q

  storage_descriptor {
    columns {
      name = "test"
      type = "string"
    }
  }
}

resource "aws_lakeformation_permissions" "test" {
  permissions = ["ALL"]
  principal   = aws_iam_role.firehose.arn

  table {
    database_name = aws_glue_catalog_database.test.name
    name          = aws_glue_catalog_table.test.name
  }
}

resource "aws_kinesis_firehose_delivery_stream" "test" {
  destination = "extended_s3"
  name        = %[1]q

  extended_s3_configuration {
    bucket_arn = aws_s3_bucket.bucket.arn
    # InvalidArgumentException: BufferingHints.SizeInMBs must be at least 64 when data format conversion is enabled.
    buffering_size = 128
    role_arn       = aws_iam_role.firehose.arn

    data_format_conversion_configuration {
      enabled = %[2]t

      input_format_configuration {
        deserializer {
          hive_json_ser_de {} # we have to pick one
        }
      }

      output_format_configuration {
        serializer {
          orc_ser_de {} # we have to pick one
        }
      }

      schema_configuration {
        database_name = aws_glue_catalog_table.test.database_name
        role_arn      = aws_iam_role.firehose.arn
        table_name    = aws_glue_catalog_table.test.name
      }
    }
  }

  depends_on = [aws_iam_role_policy.firehose, aws_lakeformation_permissions.test]
}
`, rName, enabled))
}

func testAccDeliveryStreamConfig_extendedS3DataFormatConversionConfigurationHiveJSONSerDeEmpty(rName string) string {
	return acctest.ConfigCompose(testAccDeliveryStreamConfig_base(rName), fmt.Sprintf(`
resource "aws_glue_catalog_database" "test" {
  name = %[1]q
}

resource "aws_glue_catalog_table" "test" {
  database_name = aws_glue_catalog_database.test.name
  name          = %[1]q

  storage_descriptor {
    columns {
      name = "test"
      type = "string"
    }
  }
}

resource "aws_lakeformation_permissions" "test" {
  permissions = ["ALL"]
  principal   = aws_iam_role.firehose.arn

  table {
    database_name = aws_glue_catalog_database.test.name
    name          = aws_glue_catalog_table.test.name
  }
}

resource "aws_kinesis_firehose_delivery_stream" "test" {
  destination = "extended_s3"
  name        = %[1]q

  extended_s3_configuration {
    bucket_arn = aws_s3_bucket.bucket.arn
    # InvalidArgumentException: BufferingHints.SizeInMBs must be at least 64 when data format conversion is enabled.
    buffering_size = 128
    role_arn       = aws_iam_role.firehose.arn

    data_format_conversion_configuration {
      input_format_configuration {
        deserializer {
          hive_json_ser_de {}
        }
      }

      output_format_configuration {
        serializer {
          orc_ser_de {} # we have to pick one
        }
      }

      schema_configuration {
        database_name = aws_glue_catalog_table.test.database_name
        role_arn      = aws_iam_role.firehose.arn
        table_name    = aws_glue_catalog_table.test.name
      }
    }
  }

  depends_on = [aws_iam_role_policy.firehose, aws_lakeformation_permissions.test]
}
`, rName))
}

func testAccDeliveryStreamConfig_extendedS3ExternalUpdate(rName string) string {
	return acctest.ConfigCompose(testAccDeliveryStreamConfig_base(rName), fmt.Sprintf(`
resource "aws_kinesis_firehose_delivery_stream" "test" {
  destination = "extended_s3"
  name        = %[1]q

  extended_s3_configuration {
    bucket_arn = aws_s3_bucket.bucket.arn
    role_arn   = aws_iam_role.firehose.arn
  }
}
`, rName))
}

func testAccDeliveryStreamConfig_extendedS3DataFormatConversionConfigurationOpenXJSONSerDeEmpty(rName string) string {
	return acctest.ConfigCompose(testAccDeliveryStreamConfig_base(rName), fmt.Sprintf(`
resource "aws_glue_catalog_database" "test" {
  name = %[1]q
}

resource "aws_glue_catalog_table" "test" {
  database_name = aws_glue_catalog_database.test.name
  name          = %[1]q

  storage_descriptor {
    columns {
      name = "test"
      type = "string"
    }
  }
}

resource "aws_lakeformation_permissions" "test" {
  permissions = ["ALL"]
  principal   = aws_iam_role.firehose.arn

  table {
    database_name = aws_glue_catalog_database.test.name
    name          = aws_glue_catalog_table.test.name
  }
}

resource "aws_kinesis_firehose_delivery_stream" "test" {
  destination = "extended_s3"
  name        = %[1]q

  extended_s3_configuration {
    bucket_arn = aws_s3_bucket.bucket.arn
    # InvalidArgumentException: BufferingHints.SizeInMBs must be at least 64 when data format conversion is enabled.
    buffering_size = 128
    role_arn       = aws_iam_role.firehose.arn

    data_format_conversion_configuration {
      input_format_configuration {
        deserializer {
          open_x_json_ser_de {}
        }
      }

      output_format_configuration {
        serializer {
          orc_ser_de {} # we have to pick one
        }
      }

      schema_configuration {
        database_name = aws_glue_catalog_table.test.database_name
        role_arn      = aws_iam_role.firehose.arn
        table_name    = aws_glue_catalog_table.test.name
      }
    }
  }

  depends_on = [aws_iam_role_policy.firehose, aws_lakeformation_permissions.test]
}
`, rName))
}

func testAccDeliveryStreamConfig_extendedS3DataFormatConversionConfigurationOrcSerDeEmpty(rName string) string {
	return acctest.ConfigCompose(testAccDeliveryStreamConfig_base(rName), fmt.Sprintf(`
resource "aws_glue_catalog_database" "test" {
  name = %[1]q
}

resource "aws_glue_catalog_table" "test" {
  database_name = aws_glue_catalog_database.test.name
  name          = %[1]q

  storage_descriptor {
    columns {
      name = "test"
      type = "string"
    }
  }
}

resource "aws_lakeformation_permissions" "test" {
  permissions = ["ALL"]
  principal   = aws_iam_role.firehose.arn

  table {
    database_name = aws_glue_catalog_database.test.name
    name          = aws_glue_catalog_table.test.name
  }
}

resource "aws_kinesis_firehose_delivery_stream" "test" {
  destination = "extended_s3"
  name        = %[1]q

  extended_s3_configuration {
    bucket_arn = aws_s3_bucket.bucket.arn
    # InvalidArgumentException: BufferingHints.SizeInMBs must be at least 64 when data format conversion is enabled.
    buffering_size = 128
    role_arn       = aws_iam_role.firehose.arn

    data_format_conversion_configuration {
      input_format_configuration {
        deserializer {
          hive_json_ser_de {} # we have to pick one
        }
      }

      output_format_configuration {
        serializer {
          orc_ser_de {}
        }
      }

      schema_configuration {
        database_name = aws_glue_catalog_table.test.database_name
        role_arn      = aws_iam_role.firehose.arn
        table_name    = aws_glue_catalog_table.test.name
      }
    }
  }

  depends_on = [aws_iam_role_policy.firehose, aws_lakeformation_permissions.test]
}
`, rName))
}

func testAccDeliveryStreamConfig_extendedS3DataFormatConversionConfigurationParquetSerDeEmpty(rName string) string {
	return acctest.ConfigCompose(testAccDeliveryStreamConfig_base(rName), fmt.Sprintf(`
resource "aws_glue_catalog_database" "test" {
  name = %[1]q
}

resource "aws_glue_catalog_table" "test" {
  database_name = aws_glue_catalog_database.test.name
  name          = %[1]q

  storage_descriptor {
    columns {
      name = "test"
      type = "string"
    }
  }
}

resource "aws_lakeformation_permissions" "test" {
  permissions = ["ALL"]
  principal   = aws_iam_role.firehose.arn

  table {
    database_name = aws_glue_catalog_database.test.name
    name          = aws_glue_catalog_table.test.name
  }
}

resource "aws_kinesis_firehose_delivery_stream" "test" {
  destination = "extended_s3"
  name        = %[1]q

  extended_s3_configuration {
    bucket_arn = aws_s3_bucket.bucket.arn
    # InvalidArgumentException: BufferingHints.SizeInMBs must be at least 64 when data format conversion is enabled.
    buffering_size = 128
    role_arn       = aws_iam_role.firehose.arn

    data_format_conversion_configuration {
      input_format_configuration {
        deserializer {
          hive_json_ser_de {} # we have to pick one
        }
      }

      output_format_configuration {
        serializer {
          parquet_ser_de {}
        }
      }

      schema_configuration {
        database_name = aws_glue_catalog_table.test.database_name
        role_arn      = aws_iam_role.firehose.arn
        table_name    = aws_glue_catalog_table.test.name
      }
    }
  }

  depends_on = [aws_iam_role_policy.firehose, aws_lakeformation_permissions.test]
}
`, rName))
}

func testAccDeliveryStreamConfig_extendedS3ErrorOutputPrefix(rName, errorOutputPrefix string) string {
	return acctest.ConfigCompose(testAccDeliveryStreamConfig_base(rName), fmt.Sprintf(`
resource "aws_kinesis_firehose_delivery_stream" "test" {
  destination = "extended_s3"
  name        = %[1]q

  extended_s3_configuration {
    bucket_arn          = aws_s3_bucket.bucket.arn
    error_output_prefix = %[2]q
    role_arn            = aws_iam_role.firehose.arn
  }

  depends_on = [aws_iam_role_policy.firehose]
}
`, rName, errorOutputPrefix))
}

func testAccDeliveryStreamConfig_extendedS3S3BackUpConfigurationErrorOutputPrefix(rName, errorOutputPrefix string) string {
	return acctest.ConfigCompose(testAccDeliveryStreamConfig_base(rName), fmt.Sprintf(`
resource "aws_kinesis_firehose_delivery_stream" "test" {
  destination = "extended_s3"
  name        = %[1]q

  extended_s3_configuration {
    bucket_arn     = aws_s3_bucket.bucket.arn
    role_arn       = aws_iam_role.firehose.arn
    s3_backup_mode = "Enabled"
    s3_backup_configuration {
      role_arn            = aws_iam_role.firehose.arn
      bucket_arn          = aws_s3_bucket.bucket.arn
      error_output_prefix = %[2]q
    }
  }

  depends_on = [aws_iam_role_policy.firehose]
}
`, rName, errorOutputPrefix))
}

func testAccDeliveryStreamConfig_extendedS3ProcessingConfigurationEmpty(rName string) string {
	return acctest.ConfigCompose(testAccDeliveryStreamConfig_base(rName), fmt.Sprintf(`
resource "aws_kinesis_firehose_delivery_stream" "test" {
  destination = "extended_s3"
  name        = %[1]q

  extended_s3_configuration {
    bucket_arn = aws_s3_bucket.bucket.arn
    role_arn   = aws_iam_role.firehose.arn

    processing_configuration {}
  }

  depends_on = [aws_iam_role_policy.firehose]
}
`, rName))
}

func testAccDeliveryStreamConfig_extendedS3KMSKeyARN(rName string) string {
	return acctest.ConfigCompose(
		testAccDeliveryStreamConfig_baseLambda(rName),
		testAccDeliveryStreamConfig_base(rName),
		fmt.Sprintf(`
resource "aws_kms_key" "test" {
  description = %[1]q
}

resource "aws_kinesis_firehose_delivery_stream" "test" {
  depends_on  = [aws_iam_role_policy.firehose]
  name        = %[1]q
  destination = "extended_s3"

  extended_s3_configuration {
    role_arn    = aws_iam_role.firehose.arn
    bucket_arn  = aws_s3_bucket.bucket.arn
    kms_key_arn = aws_kms_key.test.arn

    processing_configuration {
      enabled = false

      processors {
        type = "Lambda"

        parameters {
          parameter_name  = "LambdaArn"
          parameter_value = "${aws_lambda_function.lambda_function_test.arn}:$LATEST"
        }
        parameters {
          parameter_name  = "BufferSizeInMBs"
          parameter_value = "1.1"
        }
        parameters {
          parameter_name  = "BufferIntervalInSeconds"
          parameter_value = "70"
        }
      }
    }
  }
}
`, rName))
}

func testAccDeliveryStreamConfig_extendedS3DynamicPartitioning(rName string) string {
	return acctest.ConfigCompose(
		testAccDeliveryStreamConfig_baseLambda(rName),
		testAccDeliveryStreamConfig_base(rName),
		fmt.Sprintf(`
resource "aws_kinesis_firehose_delivery_stream" "test" {
  depends_on  = [aws_iam_role_policy.firehose]
  name        = %[1]q
  destination = "extended_s3"

  extended_s3_configuration {
    role_arn            = aws_iam_role.firehose.arn
    bucket_arn          = aws_s3_bucket.bucket.arn
    prefix              = "custom-prefix/customerId=!{partitionKeyFromLambda:customerId}/year=!{timestamp:yyyy}/month=!{timestamp:MM}/day=!{timestamp:dd}/hour=!{timestamp:HH}/"
    error_output_prefix = "prefix1"
    buffering_size      = 64

    dynamic_partitioning_configuration {
      enabled        = true
      retry_duration = 300
    }

    processing_configuration {
      enabled = true

      processors {
        type = "Lambda"

        parameters {
          parameter_name  = "LambdaArn"
          parameter_value = "${aws_lambda_function.lambda_function_test.arn}:$LATEST"
        }
        parameters {
          parameter_name  = "BufferSizeInMBs"
          parameter_value = "1.1"
        }
        parameters {
          parameter_name  = "BufferIntervalInSeconds"
          parameter_value = "70"
        }
      }

      processors {
        type = "RecordDeAggregation"

        parameters {
          parameter_name  = "SubRecordType"
          parameter_value = "JSON"
        }
      }
    }
  }
}
`, rName))
}

func testAccDeliveryStreamConfig_extendedS3DynamicPartitioningBasic(rName string) string {
	return acctest.ConfigCompose(
		testAccDeliveryStreamConfig_baseLambda(rName),
		testAccDeliveryStreamConfig_base(rName),
		fmt.Sprintf(`
resource "aws_kinesis_firehose_delivery_stream" "test" {
  depends_on  = [aws_iam_role_policy.firehose]
  name        = %[1]q
  destination = "extended_s3"

  extended_s3_configuration {
    role_arn            = aws_iam_role.firehose.arn
    bucket_arn          = aws_s3_bucket.bucket.arn
    error_output_prefix = "prefix1"
    buffering_size      = 64
  }
}
`, rName))
}

func testAccDeliveryStreamConfig_extendedS3UpdatesInitial(rName string) string {
	return acctest.ConfigCompose(
		testAccDeliveryStreamConfig_baseLambda(rName),
		testAccDeliveryStreamConfig_base(rName),
		fmt.Sprintf(`
resource "aws_kinesis_firehose_delivery_stream" "test" {
  depends_on  = [aws_iam_role_policy.firehose]
  name        = %[1]q
  destination = "extended_s3"

  extended_s3_configuration {
    role_arn   = aws_iam_role.firehose.arn
    bucket_arn = aws_s3_bucket.bucket.arn

    processing_configuration {
      enabled = true

      processors {
        type = "Lambda"

        parameters {
          parameter_name  = "BufferIntervalInSeconds"
          parameter_value = "70"
        }
        parameters {
          parameter_name  = "LambdaArn"
          parameter_value = "${aws_lambda_function.lambda_function_test.arn}:$LATEST"
        }
        parameters {
          parameter_name  = "BufferSizeInMBs"
          parameter_value = "1.1"
        }
      }
    }

    buffering_size     = 10
    buffering_interval = 400
    compression_format = "GZIP"
    s3_backup_mode     = "Enabled"

    s3_backup_configuration {
      role_arn   = aws_iam_role.firehose.arn
      bucket_arn = aws_s3_bucket.bucket.arn
    }
  }
}
`, rName))
}

func testAccDeliveryStreamConfig_extendedS3UpdatesSetBufferIntervalNoBufferSize(rName string) string {
	return acctest.ConfigCompose(
		testAccDeliveryStreamConfig_baseLambda(rName),
		testAccDeliveryStreamConfig_base(rName),
		fmt.Sprintf(`
resource "aws_kinesis_firehose_delivery_stream" "test" {
  depends_on  = [aws_iam_role_policy.firehose]
  name        = %[1]q
  destination = "extended_s3"

  extended_s3_configuration {
    role_arn   = aws_iam_role.firehose.arn
    bucket_arn = aws_s3_bucket.bucket.arn

    processing_configuration {
      enabled = true

      processors {
        type = "Lambda"

        parameters {
          parameter_name  = "LambdaArn"
          parameter_value = "${aws_lambda_function.lambda_function_test.arn}:$LATEST"
        }
        parameters {
          parameter_name  = "BufferIntervalInSeconds"
          parameter_value = 201
        }
      }
    }

    buffering_size     = 10
    buffering_interval = 400
    compression_format = "GZIP"
    s3_backup_mode     = "Enabled"

    s3_backup_configuration {
      role_arn   = aws_iam_role.firehose.arn
      bucket_arn = aws_s3_bucket.bucket.arn
    }
  }
}
`, rName))
}

func testAccDeliveryStreamConfig_extendedS3UpdatesRemoveProcessors(rName string) string {
	return acctest.ConfigCompose(
		testAccDeliveryStreamConfig_baseLambda(rName),
		testAccDeliveryStreamConfig_base(rName),
		fmt.Sprintf(`
resource "aws_kinesis_firehose_delivery_stream" "test" {
  depends_on  = [aws_iam_role_policy.firehose]
  name        = %[1]q
  destination = "extended_s3"

  extended_s3_configuration {
    role_arn           = aws_iam_role.firehose.arn
    bucket_arn         = aws_s3_bucket.bucket.arn
    buffering_size     = 10
    buffering_interval = 400
    compression_format = "GZIP"
    s3_backup_mode     = "Enabled"

    s3_backup_configuration {
      role_arn   = aws_iam_role.firehose.arn
      bucket_arn = aws_s3_bucket.bucket.arn
    }
  }
}
`, rName))
}

func testAccDeliveryStreamConfig_extendedS3CustomTimeZoneAndFileExtensionUpdates(rName, customTimeZone, fileExtension string) string {
	return acctest.ConfigCompose(testAccDeliveryStreamConfig_base(rName), fmt.Sprintf(`
resource "aws_kinesis_firehose_delivery_stream" "test" {
  depends_on  = [aws_iam_role_policy.firehose]
  name        = %[1]q
  destination = "extended_s3"

  extended_s3_configuration {
    role_arn         = aws_iam_role.firehose.arn
    bucket_arn       = aws_s3_bucket.bucket.arn
    custom_time_zone = %[2]q
    file_extension   = %[3]q
  }
}
`, rName, customTimeZone, fileExtension))
}

func testAccDeliveryStreamConfig_extendedS3CustomTimeZoneAndFileExtensionUpdatesNoValues(rName string) string {
	return acctest.ConfigCompose(testAccDeliveryStreamConfig_base(rName), fmt.Sprintf(`
resource "aws_kinesis_firehose_delivery_stream" "test" {
  depends_on  = [aws_iam_role_policy.firehose]
  name        = %[1]q
  destination = "extended_s3"

  extended_s3_configuration {
    role_arn   = aws_iam_role.firehose.arn
    bucket_arn = aws_s3_bucket.bucket.arn
  }
}
`, rName))
}

func testAccDeliveryStreamConfig_baseRedshift(rName string) string {
	return acctest.ConfigCompose(
		testAccDeliveryStreamConfig_base(rName),
		acctest.ConfigVPCWithSubnets(rName, 1),
		fmt.Sprintf(`
resource "aws_redshift_subnet_group" "test" {
  name        = %[1]q
  description = "test"
  subnet_ids  = aws_subnet.test[*].id
}

resource "aws_redshift_cluster" "test" {
  cluster_identifier        = %[1]q
  availability_zone         = data.aws_availability_zones.available.names[0]
  database_name             = "test"
  master_username           = "testuser"
  master_password           = "T3stPass"
  node_type                 = "dc2.large"
  cluster_type              = "single-node"
  skip_final_snapshot       = true
  cluster_subnet_group_name = aws_redshift_subnet_group.test.id
  publicly_accessible       = false
}
`, rName))
}

func testAccDeliveryStreamConfig_redshift(rName string) string {
	return acctest.ConfigCompose(testAccDeliveryStreamConfig_baseRedshift(rName), fmt.Sprintf(`
resource "aws_kinesis_firehose_delivery_stream" "test" {
  name        = %[1]q
  destination = "redshift"

  redshift_configuration {
    role_arn        = aws_iam_role.firehose.arn
    cluster_jdbcurl = "jdbc:redshift://${aws_redshift_cluster.test.endpoint}/${aws_redshift_cluster.test.database_name}"
    username        = "testuser"
    password        = "T3stPass"
    data_table_name = "test-table"

    s3_configuration {
      role_arn   = aws_iam_role.firehose.arn
      bucket_arn = aws_s3_bucket.bucket.arn
    }
  }
}
`, rName))
}

func testAccDeliveryStreamConfig_redshiftUpdates(rName string) string {
	return acctest.ConfigCompose(
		testAccDeliveryStreamConfig_baseLambda(rName),
		testAccDeliveryStreamConfig_baseRedshift(rName),
		fmt.Sprintf(`
resource "aws_kinesis_firehose_delivery_stream" "test" {
  name        = %[1]q
  destination = "redshift"

  redshift_configuration {
    role_arn        = aws_iam_role.firehose.arn
    cluster_jdbcurl = "jdbc:redshift://${aws_redshift_cluster.test.endpoint}/${aws_redshift_cluster.test.database_name}"
    username        = "testuser"
    password        = "T3stPass"
    s3_backup_mode  = "Enabled"

    s3_backup_configuration {
      role_arn   = aws_iam_role.firehose.arn
      bucket_arn = aws_s3_bucket.bucket.arn
    }

    s3_configuration {
      role_arn           = aws_iam_role.firehose.arn
      bucket_arn         = aws_s3_bucket.bucket.arn
      buffering_size     = 10
      buffering_interval = 400
      compression_format = "GZIP"
    }


    data_table_name    = "test-table"
    copy_options       = "GZIP"
    data_table_columns = "test-col"

    processing_configuration {
      enabled = false

      processors {
        type = "Lambda"

        parameters {
          parameter_name  = "LambdaArn"
          parameter_value = "${aws_lambda_function.lambda_function_test.arn}:$LATEST"
        }
        parameters {
          parameter_name  = "BufferSizeInMBs"
          parameter_value = "1.1"
        }
        parameters {
          parameter_name  = "BufferIntervalInSeconds"
          parameter_value = "70"
        }
      }
    }
  }
}
`, rName))
}

func testAccDeliveryStreamConfig_redshiftSecretsManager(rName string) string {
	return acctest.ConfigCompose(testAccDeliveryStreamConfig_baseRedshift(rName), fmt.Sprintf(`
resource "aws_secretsmanager_secret" "test" {
  name = %[1]q
}

resource "aws_kinesis_firehose_delivery_stream" "test" {
  name        = %[1]q
  destination = "redshift"

  redshift_configuration {
    role_arn        = aws_iam_role.firehose.arn
    cluster_jdbcurl = "jdbc:redshift://${aws_redshift_cluster.test.endpoint}/${aws_redshift_cluster.test.database_name}"
    data_table_name = "test-table"

    s3_configuration {
      role_arn   = aws_iam_role.firehose.arn
      bucket_arn = aws_s3_bucket.bucket.arn
    }

    secrets_manager_configuration {
      enabled    = true
      role_arn   = aws_iam_role.firehose.arn
      secret_arn = aws_secretsmanager_secret.test.arn
    }
  }
}
`, rName))
}

func testAccDeliveryStreamConfig_snowflakeBasic(rName, privateKey string) string {
	return acctest.ConfigCompose(testAccDeliveryStreamConfig_base(rName), fmt.Sprintf(`
resource "aws_kinesis_firehose_delivery_stream" "test" {
  depends_on  = [aws_iam_role_policy.firehose]
  name        = %[1]q
  destination = "snowflake"

  snowflake_configuration {
    account_url = "https://%[1]s.snowflakecomputing.com"
    database    = "test-db"
    private_key = "%[2]s"
    role_arn    = aws_iam_role.firehose.arn
    schema      = "test-schema"
    table       = "test-table"
    user        = "test-usr"

    s3_configuration {
      role_arn           = aws_iam_role.firehose.arn
      bucket_arn         = aws_s3_bucket.bucket.arn
      buffering_size     = 10
      buffering_interval = 400
      compression_format = "GZIP"
    }
  }
}
`, rName, acctest.TLSPEMRemoveRSAPrivateKeyEncapsulationBoundaries(acctest.TLSPEMRemoveNewlines(privateKey))))
}

func testAccDeliveryStreamConfig_snowflakeUpdate(rName, privateKey string) string {
	return acctest.ConfigCompose(testAccDeliveryStreamConfig_base(rName), testAccDeliveryStreamConfig_baseLambda(rName), fmt.Sprintf(`
resource "aws_kinesis_firehose_delivery_stream" "test" {
  depends_on  = [aws_iam_role_policy.firehose]
  name        = %[1]q
  destination = "snowflake"

  snowflake_configuration {
    account_url = "https://%[1]s.snowflakecomputing.com"
    database    = "test-db"
    private_key = "%[2]s"
    role_arn    = aws_iam_role.firehose.arn
    schema      = "test-schema"
    table       = "test-table"
    user        = "test-usr"

    s3_configuration {
      role_arn           = aws_iam_role.firehose.arn
      bucket_arn         = aws_s3_bucket.bucket.arn
      buffering_size     = 10
      buffering_interval = 400
      compression_format = "GZIP"
    }

    data_loading_option = "VARIANT_CONTENT_MAPPING"
    content_column_name = "test-content"

    snowflake_role_configuration {
      enabled        = true
      snowflake_role = "test-role"
    }

    processing_configuration {
      enabled = false

      processors {
        type = "Lambda"

        parameters {
          parameter_name  = "LambdaArn"
          parameter_value = "${aws_lambda_function.lambda_function_test.arn}:$LATEST"
        }
        parameters {
          parameter_name  = "BufferSizeInMBs"
          parameter_value = "1.1"
        }
        parameters {
          parameter_name  = "BufferIntervalInSeconds"
          parameter_value = "70"
        }
      }
    }
  }
}
`, rName, acctest.TLSPEMRemoveRSAPrivateKeyEncapsulationBoundaries(acctest.TLSPEMRemoveNewlines(privateKey))))
}

func testAccDeliveryStreamConfig_snowflakeUpdateSecretsManager(rName, privateKey string) string {
	return acctest.ConfigCompose(testAccDeliveryStreamConfig_base(rName), testAccDeliveryStreamConfig_baseSecretsManager(rName, privateKey), fmt.Sprintf(`
resource "aws_kinesis_firehose_delivery_stream" "test" {
  depends_on  = [aws_iam_role_policy.firehose, aws_iam_role_policy.iam_policy_for_secretsmanager]
  name        = %[1]q
  destination = "snowflake"

  snowflake_configuration {
    account_url = "https://%[1]s.snowflakecomputing.com"
    database    = "test-db"
    role_arn    = aws_iam_role.firehose.arn
    schema      = "test-schema"
    table       = "test-table"

    secrets_manager_configuration {
      enabled    = true
      secret_arn = aws_secretsmanager_secret.test.arn
      role_arn   = aws_iam_role.iam_for_secretsmanager.arn
    }

    s3_configuration {
      role_arn           = aws_iam_role.firehose.arn
      bucket_arn         = aws_s3_bucket.bucket.arn
      buffering_size     = 10
      buffering_interval = 400
      compression_format = "GZIP"
    }
  }
}
`, rName))
}

func testAccDeliveryStreamConfig_splunkBasic(rName string) string {
	return acctest.ConfigCompose(testAccDeliveryStreamConfig_base(rName), fmt.Sprintf(`
resource "aws_kinesis_firehose_delivery_stream" "test" {
  depends_on  = [aws_iam_role_policy.firehose]
  name        = %[1]q
  destination = "splunk"

  splunk_configuration {
    hec_endpoint = "https://input-test.com:443"
    hec_token    = "51D4DA16-C61B-4F5F-8EC7-ED4301342A4A"

    s3_configuration {
      role_arn   = aws_iam_role.firehose.arn
      bucket_arn = aws_s3_bucket.bucket.arn
    }
  }
}
`, rName))
}

func testAccDeliveryStreamConfig_splunkUpdates(rName string) string {
	return acctest.ConfigCompose(
		testAccDeliveryStreamConfig_baseLambda(rName),
		testAccDeliveryStreamConfig_base(rName),
		fmt.Sprintf(`
resource "aws_kinesis_firehose_delivery_stream" "test" {
  depends_on  = [aws_iam_role_policy.firehose]
  name        = %[1]q
  destination = "splunk"

  splunk_configuration {
    buffering_interval         = 45
    buffering_size             = 3
    hec_endpoint               = "https://input-test.com:443"
    hec_token                  = "51D4DA16-C61B-4F5F-8EC7-ED4301342A4A"
    hec_acknowledgment_timeout = 600
    hec_endpoint_type          = "Event"
    s3_backup_mode             = "FailedEventsOnly"

    s3_configuration {
      role_arn           = aws_iam_role.firehose.arn
      bucket_arn         = aws_s3_bucket.bucket.arn
      buffering_size     = 10
      buffering_interval = 400
      compression_format = "GZIP"
    }

    processing_configuration {
      enabled = true

      processors {
        type = "Lambda"

        parameters {
          parameter_name  = "LambdaArn"
          parameter_value = "${aws_lambda_function.lambda_function_test.arn}:$LATEST"
        }

        parameters {
          parameter_name  = "RoleArn"
          parameter_value = aws_iam_role.firehose.arn
        }

        parameters {
          parameter_name  = "BufferSizeInMBs"
          parameter_value = "1.1"
        }

        parameters {
          parameter_name  = "BufferIntervalInSeconds"
          parameter_value = "70"
        }
      }
    }
  }
}
`, rName))
}

func testAccDeliveryStreamConfig_splunkSecretsManager(rName string) string {
	return acctest.ConfigCompose(testAccDeliveryStreamConfig_base(rName), fmt.Sprintf(`
resource "aws_secretsmanager_secret" "test" {
  name = %[1]q
}

resource "aws_kinesis_firehose_delivery_stream" "test" {
  depends_on  = [aws_iam_role_policy.firehose]
  name        = %[1]q
  destination = "splunk"

  splunk_configuration {
    hec_endpoint = "https://input-test.com:443"

    s3_configuration {
      role_arn   = aws_iam_role.firehose.arn
      bucket_arn = aws_s3_bucket.bucket.arn
    }

    secrets_manager_configuration {
      enabled    = true
      role_arn   = aws_iam_role.firehose.arn
      secret_arn = aws_secretsmanager_secret.test.arn
    }
  }
}
`, rName))
}

func testAccDeliveryStreamConfig_splunkErrorOutputPrefix(rName, errorOutputPrefix string) string {
	return acctest.ConfigCompose(testAccDeliveryStreamConfig_base(rName), fmt.Sprintf(`
resource "aws_kinesis_firehose_delivery_stream" "test" {
  depends_on  = [aws_iam_role_policy.firehose]
  name        = %[1]q
  destination = "splunk"

  splunk_configuration {
    hec_endpoint = "https://input-test.com:443"
    hec_token    = "51D4DA16-C61B-4F5F-8EC7-ED4301342A4A"

    s3_configuration {
      role_arn            = aws_iam_role.firehose.arn
      bucket_arn          = aws_s3_bucket.bucket.arn
      error_output_prefix = %[2]q
    }
  }
}
`, rName, errorOutputPrefix))
}

func testAccDeliveryStreamConfig_httpEndpointBasic(rName string) string {
	return acctest.ConfigCompose(testAccDeliveryStreamConfig_base(rName), fmt.Sprintf(`
resource "aws_kinesis_firehose_delivery_stream" "test" {
  depends_on  = [aws_iam_role_policy.firehose]
  name        = %[1]q
  destination = "http_endpoint"

  http_endpoint_configuration {
    url      = "https://input-test.com:443"
    name     = "HTTP_test"
    role_arn = aws_iam_role.firehose.arn

    s3_configuration {
      role_arn   = aws_iam_role.firehose.arn
      bucket_arn = aws_s3_bucket.bucket.arn
    }
  }
}
`, rName))
}

func testAccDeliveryStreamConfig_httpEndpointErrorOutputPrefix(rName, errorOutputPrefix string) string {
	return acctest.ConfigCompose(testAccDeliveryStreamConfig_base(rName), fmt.Sprintf(`
resource "aws_kinesis_firehose_delivery_stream" "test" {
  depends_on  = [aws_iam_role_policy.firehose]
  name        = %[1]q
  destination = "http_endpoint"

  http_endpoint_configuration {
    url      = "https://input-test.com:443"
    name     = "HTTP_test"
    role_arn = aws_iam_role.firehose.arn

    s3_configuration {
      role_arn            = aws_iam_role.firehose.arn
      bucket_arn          = aws_s3_bucket.bucket.arn
      error_output_prefix = %[2]q
    }
  }
}
`, rName, errorOutputPrefix))
}

func testAccDeliveryStreamConfig_httpEndpointRetryDuration(rName string, retryDuration int) string {
	return acctest.ConfigCompose(testAccDeliveryStreamConfig_base(rName), fmt.Sprintf(`
resource "aws_kinesis_firehose_delivery_stream" "test" {
  depends_on  = [aws_iam_role_policy.firehose]
  name        = %[1]q
  destination = "http_endpoint"

  http_endpoint_configuration {
    url            = "https://input-test.com:443"
    name           = "HTTP_test"
    retry_duration = %[2]d
    role_arn       = aws_iam_role.firehose.arn

    s3_configuration {
      role_arn   = aws_iam_role.firehose.arn
      bucket_arn = aws_s3_bucket.bucket.arn
    }
  }
}
`, rName, retryDuration))
}

func testAccDeliveryStreamConfig_httpEndpointUpdates(rName string) string {
	return acctest.ConfigCompose(
		testAccDeliveryStreamConfig_baseLambda(rName),
		testAccDeliveryStreamConfig_base(rName),
		fmt.Sprintf(`
resource "aws_kinesis_firehose_delivery_stream" "test" {
  depends_on  = [aws_iam_role_policy.firehose]
  name        = %[1]q
  destination = "http_endpoint"

  http_endpoint_configuration {
    url            = "https://input-test.com:443"
    name           = "HTTP_test"
    access_key     = "test_key"
    role_arn       = aws_iam_role.firehose.arn
    s3_backup_mode = "FailedDataOnly"

    s3_configuration {
      role_arn           = aws_iam_role.firehose.arn
      bucket_arn         = aws_s3_bucket.bucket.arn
      buffering_size     = 10
      buffering_interval = 400
      compression_format = "GZIP"
    }

    processing_configuration {
      enabled = true

      processors {
        type = "Lambda"

        parameters {
          parameter_name  = "LambdaArn"
          parameter_value = "${aws_lambda_function.lambda_function_test.arn}:$LATEST"
        }

        parameters {
          parameter_name  = "BufferSizeInMBs"
          parameter_value = "1.1"
        }

        parameters {
          parameter_name  = "BufferIntervalInSeconds"
          parameter_value = "70"
        }
      }
    }
  }
}
`, rName))
}

func testAccDeliveryStreamConfig_httpEndpointSecretsManager(rName string) string {
	return acctest.ConfigCompose(testAccDeliveryStreamConfig_base(rName), fmt.Sprintf(`
resource "aws_secretsmanager_secret" "test" {
  name = %[1]q
}

resource "aws_kinesis_firehose_delivery_stream" "test" {
  depends_on = [aws_iam_role_policy.firehose]

  name        = %[1]q
  destination = "http_endpoint"

  http_endpoint_configuration {
    url      = "https://input-test.com:443"
    name     = "HTTP_test"
    role_arn = aws_iam_role.firehose.arn

    s3_configuration {
      role_arn   = aws_iam_role.firehose.arn
      bucket_arn = aws_s3_bucket.bucket.arn
    }

    secrets_manager_configuration {
      enabled    = true
      role_arn   = aws_iam_role.firehose.arn
      secret_arn = aws_secretsmanager_secret.test.arn
    }
  }
}
`, rName))
}

func testAccDeliveryStreamConfig_baseElasticsearch(rName string) string {
	return acctest.ConfigCompose(testAccDeliveryStreamConfig_base(rName), fmt.Sprintf(`
resource "aws_elasticsearch_domain" "test_cluster" {
  domain_name = substr(%[1]q, 0, 28)

  cluster_config {
    instance_type = "m4.large.elasticsearch"
  }

  domain_endpoint_options {
    enforce_https       = true
    tls_security_policy = "Policy-Min-TLS-1-2-2019-07"
  }

  ebs_options {
    ebs_enabled = true
    volume_size = 10
  }
}

resource "aws_iam_role_policy" "firehose-elasticsearch" {
  name   = "%[1]s-es"
  role   = aws_iam_role.firehose.id
  policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Action": [
        "es:*"
      ],
      "Resource": [
        "${aws_elasticsearch_domain.test_cluster.arn}",
        "${aws_elasticsearch_domain.test_cluster.arn}/*"
      ]
    }
  ]
}
EOF
}
`, rName))
}

func testAccDeliveryStreamConfig_baseElasticsearchVPC(rName string) string {
	return acctest.ConfigCompose(
		testAccDeliveryStreamConfig_base(rName),
		acctest.ConfigVPCWithSubnets(rName, 2),
		fmt.Sprintf(`
resource "aws_security_group" "test" {
  count = 2

  name   = "%[1]s-${count.index}"
  vpc_id = aws_vpc.test.id

  tags = {
    Name = %[1]q
  }
}

resource "aws_elasticsearch_domain" "test_cluster" {
  domain_name = substr(%[1]q, 0, 28)

  cluster_config {
    instance_count         = 2
    zone_awareness_enabled = true
    instance_type          = "t2.small.elasticsearch"
  }

  ebs_options {
    ebs_enabled = true
    volume_size = 10
  }

  vpc_options {
    security_group_ids = aws_security_group.test[*].id
    subnet_ids         = aws_subnet.test[*].id
  }
}

resource "aws_iam_role_policy" "firehose-elasticsearch" {
  name   = "%[1]s-es"
  role   = aws_iam_role.firehose.id
  policy = <<EOF
{
	"Version":"2012-10-17",
	"Statement":[
	   {
		  "Effect":"Allow",
		  "Action":[
			 "es:*"
		  ],
		  "Resource":[
			"${aws_elasticsearch_domain.test_cluster.arn}",
			"${aws_elasticsearch_domain.test_cluster.arn}/*"
		  ]
	   },
	   {
		  "Effect":"Allow",
		  "Action":[
			 "ec2:Describe*",
			 "ec2:CreateNetworkInterface",
			 "ec2:CreateNetworkInterfacePermission",
			 "ec2:DeleteNetworkInterface"
		  ],
		  "Resource":[
			 "*"
		  ]
	   }
	]
}
EOF
}
`, rName))
}

func testAccDeliveryStreamConfig_elasticsearchBasic(rName string) string {
	return acctest.ConfigCompose(testAccDeliveryStreamConfig_baseElasticsearch(rName), fmt.Sprintf(`
resource "aws_kinesis_firehose_delivery_stream" "test" {
  depends_on = [aws_iam_role_policy.firehose-elasticsearch]

  name        = %[1]q
  destination = "elasticsearch"

  elasticsearch_configuration {
    domain_arn = aws_elasticsearch_domain.test_cluster.arn
    role_arn   = aws_iam_role.firehose.arn
    index_name = "test"
    type_name  = "test"

    s3_configuration {
      role_arn   = aws_iam_role.firehose.arn
      bucket_arn = aws_s3_bucket.bucket.arn
    }
  }
}
`, rName))
}

func testAccDeliveryStreamConfig_elasticsearchErrorOutputPrefix(rName, errorOutputPrefix string) string {
	return acctest.ConfigCompose(testAccDeliveryStreamConfig_baseElasticsearch(rName), fmt.Sprintf(`
resource "aws_kinesis_firehose_delivery_stream" "test" {
  depends_on = [aws_iam_role_policy.firehose-elasticsearch]

  name        = %[1]q
  destination = "elasticsearch"

  elasticsearch_configuration {
    domain_arn = aws_elasticsearch_domain.test_cluster.arn
    role_arn   = aws_iam_role.firehose.arn
    index_name = "test"
    type_name  = "test"

    s3_configuration {
      role_arn            = aws_iam_role.firehose.arn
      bucket_arn          = aws_s3_bucket.bucket.arn
      error_output_prefix = %[2]q
    }
  }
}
`, rName, errorOutputPrefix))
}

func testAccDeliveryStreamConfig_elasticsearchVPCBasic(rName string) string {
	return acctest.ConfigCompose(testAccDeliveryStreamConfig_baseElasticsearchVPC(rName), fmt.Sprintf(`
resource "aws_kinesis_firehose_delivery_stream" "test" {
  depends_on = [aws_iam_role_policy.firehose-elasticsearch]

  name        = %[1]q
  destination = "elasticsearch"

  elasticsearch_configuration {
    domain_arn = aws_elasticsearch_domain.test_cluster.arn
    role_arn   = aws_iam_role.firehose.arn
    index_name = "test"
    type_name  = "test"

    s3_configuration {
      role_arn   = aws_iam_role.firehose.arn
      bucket_arn = aws_s3_bucket.bucket.arn
    }

    vpc_config {
      subnet_ids         = aws_subnet.test[*].id
      security_group_ids = aws_security_group.test[*].id
      role_arn           = aws_iam_role.firehose.arn
    }
  }
}
`, rName))
}

func testAccDeliveryStreamConfig_elasticsearchUpdate(rName string) string {
	return acctest.ConfigCompose(
		testAccDeliveryStreamConfig_baseLambda(rName),
		testAccDeliveryStreamConfig_baseElasticsearch(rName),
		fmt.Sprintf(`
resource "aws_kinesis_firehose_delivery_stream" "test" {
  depends_on = [aws_iam_role_policy.firehose-elasticsearch]

  name        = %[1]q
  destination = "elasticsearch"

  elasticsearch_configuration {
    domain_arn         = aws_elasticsearch_domain.test_cluster.arn
    role_arn           = aws_iam_role.firehose.arn
    index_name         = "test"
    type_name          = "test"
    buffering_interval = 500

    s3_configuration {
      role_arn   = aws_iam_role.firehose.arn
      bucket_arn = aws_s3_bucket.bucket.arn
    }

    processing_configuration {
      enabled = false

      processors {
        type = "Lambda"

        parameters {
          parameter_name  = "LambdaArn"
          parameter_value = "${aws_lambda_function.lambda_function_test.arn}:$LATEST"
        }
        parameters {
          parameter_name  = "BufferSizeInMBs"
          parameter_value = "1.1"
        }
        parameters {
          parameter_name  = "BufferIntervalInSeconds"
          parameter_value = "70"
        }
      }
    }
  }
}
`, rName))
}

func testAccDeliveryStreamConfig_elasticsearchVPCUpdate(rName string) string {
	return acctest.ConfigCompose(
		testAccDeliveryStreamConfig_baseLambda(rName),
		testAccDeliveryStreamConfig_baseElasticsearchVPC(rName),
		fmt.Sprintf(`
resource "aws_kinesis_firehose_delivery_stream" "test" {
  depends_on = [aws_iam_role_policy.firehose-elasticsearch]

  name        = %[1]q
  destination = "elasticsearch"

  elasticsearch_configuration {
    domain_arn         = aws_elasticsearch_domain.test_cluster.arn
    role_arn           = aws_iam_role.firehose.arn
    index_name         = "test"
    type_name          = "test"
    buffering_interval = 500

    s3_configuration {
      role_arn   = aws_iam_role.firehose.arn
      bucket_arn = aws_s3_bucket.bucket.arn
    }

    vpc_config {
      subnet_ids         = aws_subnet.test[*].id
      security_group_ids = aws_security_group.test[*].id
      role_arn           = aws_iam_role.firehose.arn
    }

    processing_configuration {
      enabled = false
      processors {
        type = "Lambda"
        parameters {
          parameter_name  = "LambdaArn"
          parameter_value = "${aws_lambda_function.lambda_function_test.arn}:$LATEST"
        }
        parameters {
          parameter_name  = "BufferSizeInMBs"
          parameter_value = "1.1"
        }
        parameters {
          parameter_name  = "BufferIntervalInSeconds"
          parameter_value = "70"
        }
      }
    }
  }
}`, rName))
}

func testAccDeliveryStreamConfig_elasticsearchEndpoint(rName string) string {
	return acctest.ConfigCompose(testAccDeliveryStreamConfig_baseElasticsearch(rName), fmt.Sprintf(`
resource "aws_kinesis_firehose_delivery_stream" "test" {
  depends_on = [aws_iam_role_policy.firehose-elasticsearch]

  name        = %[1]q
  destination = "elasticsearch"

  elasticsearch_configuration {
    cluster_endpoint = "https://${aws_elasticsearch_domain.test_cluster.endpoint}"
    role_arn         = aws_iam_role.firehose.arn
    index_name       = "test"
    type_name        = "test"

    s3_configuration {
      role_arn   = aws_iam_role.firehose.arn
      bucket_arn = aws_s3_bucket.bucket.arn
    }
  }
}`, rName))
}

func testAccDeliveryStreamConfig_elasticsearchEndpointUpdate(rName string) string {
	return acctest.ConfigCompose(
		testAccDeliveryStreamConfig_baseLambda(rName),
		testAccDeliveryStreamConfig_baseElasticsearch(rName),
		fmt.Sprintf(`
resource "aws_kinesis_firehose_delivery_stream" "test" {
  depends_on = [aws_iam_role_policy.firehose-elasticsearch]

  name        = %[1]q
  destination = "elasticsearch"

  elasticsearch_configuration {
    cluster_endpoint   = "https://${aws_elasticsearch_domain.test_cluster.endpoint}"
    role_arn           = aws_iam_role.firehose.arn
    index_name         = "test"
    type_name          = "test"
    buffering_interval = 500

    s3_configuration {
      role_arn   = aws_iam_role.firehose.arn
      bucket_arn = aws_s3_bucket.bucket.arn
    }

    processing_configuration {
      enabled = false

      processors {
        type = "Lambda"

        parameters {
          parameter_name  = "LambdaArn"
          parameter_value = "${aws_lambda_function.lambda_function_test.arn}:$LATEST"
        }
        parameters {
          parameter_name  = "BufferSizeInMBs"
          parameter_value = "1.1"
        }
        parameters {
          parameter_name  = "BufferIntervalInSeconds"
          parameter_value = "70"
        }
      }
    }
  }
}`, rName))
}

func testAccDeliveryStreamConfig_baseOpensearch(rName string) string {
	return acctest.ConfigCompose(testAccDeliveryStreamConfig_base(rName), fmt.Sprintf(`
resource "aws_opensearch_domain" "test_cluster" {
  domain_name = substr(%[1]q, 0, 28)

  cluster_config {
    instance_type = "m4.large.search"
  }

  domain_endpoint_options {
    enforce_https       = true
    tls_security_policy = "Policy-Min-TLS-1-2-2019-07"
  }

  ebs_options {
    ebs_enabled = true
    volume_size = 10
  }
}

resource "aws_iam_role_policy" "firehose-opensearch" {
  name   = "%[1]s-es"
  role   = aws_iam_role.firehose.id
  policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Action": [
        "es:*"
      ],
      "Resource": [
        "${aws_opensearch_domain.test_cluster.arn}",
        "${aws_opensearch_domain.test_cluster.arn}/*"
      ]
    }
  ]
}
EOF
}
`, rName))
}

func testAccDeliveryStreamConfig_baseOpensearchVPC(rName string) string {
	return acctest.ConfigCompose(
		testAccDeliveryStreamConfig_base(rName),
		acctest.ConfigVPCWithSubnets(rName, 2),
		fmt.Sprintf(`
resource "aws_security_group" "test" {
  count = 2

  name   = "%[1]s-${count.index}"
  vpc_id = aws_vpc.test.id

  tags = {
    Name = %[1]q
  }
}

resource "aws_opensearch_domain" "test_cluster" {
  domain_name = substr(%[1]q, 0, 28)

  cluster_config {
    instance_count         = 2
    zone_awareness_enabled = true
    instance_type          = "t2.small.search"
  }

  ebs_options {
    ebs_enabled = true
    volume_size = 10
  }

  vpc_options {
    security_group_ids = aws_security_group.test[*].id
    subnet_ids         = aws_subnet.test[*].id
  }
}

resource "aws_iam_role_policy" "firehose-opensearch" {
  name   = "%[1]s-os"
  role   = aws_iam_role.firehose.id
  policy = <<EOF
{
	"Version":"2012-10-17",
	"Statement":[
	   {
		  "Effect":"Allow",
		  "Action":[
			 "es:*"
		  ],
		  "Resource":[
			"${aws_opensearch_domain.test_cluster.arn}",
			"${aws_opensearch_domain.test_cluster.arn}/*"
		  ]
	   },
	   {
		  "Effect":"Allow",
		  "Action":[
			 "ec2:Describe*",
			 "ec2:CreateNetworkInterface",
			 "ec2:CreateNetworkInterfacePermission",
			 "ec2:DeleteNetworkInterface"
		  ],
		  "Resource":[
			 "*"
		  ]
	   }
	]
}
EOF
}
`, rName))
}

func testAccDeliveryStreamConfig_opensearchBasic(rName string) string {
	return acctest.ConfigCompose(testAccDeliveryStreamConfig_baseOpensearch(rName), fmt.Sprintf(`
resource "aws_kinesis_firehose_delivery_stream" "test" {
  depends_on = [aws_iam_role_policy.firehose-opensearch]

  name        = %[1]q
  destination = "opensearch"

  opensearch_configuration {
    domain_arn = aws_opensearch_domain.test_cluster.arn
    role_arn   = aws_iam_role.firehose.arn
    index_name = "test"

    s3_configuration {
      role_arn   = aws_iam_role.firehose.arn
      bucket_arn = aws_s3_bucket.bucket.arn
    }
  }
}
`, rName))
}

func testAccDeliveryStreamConfig_opensearchEndpoint(rName string) string {
	return acctest.ConfigCompose(testAccDeliveryStreamConfig_baseOpensearch(rName), fmt.Sprintf(`
resource "aws_kinesis_firehose_delivery_stream" "test" {
  depends_on = [aws_iam_role_policy.firehose-opensearch]

  name        = %[1]q
  destination = "opensearch"

  opensearch_configuration {
    cluster_endpoint = "https://${aws_opensearch_domain.test_cluster.endpoint}"
    role_arn         = aws_iam_role.firehose.arn
    index_name       = "test"

    s3_configuration {
      role_arn   = aws_iam_role.firehose.arn
      bucket_arn = aws_s3_bucket.bucket.arn
    }
  }
}`, rName))
}

func testAccDeliveryStreamConfig_opensearchEndpointUpdate(rName string) string {
	return acctest.ConfigCompose(
		testAccDeliveryStreamConfig_baseLambda(rName),
		testAccDeliveryStreamConfig_baseOpensearch(rName),
		fmt.Sprintf(`
resource "aws_kinesis_firehose_delivery_stream" "test" {
  depends_on = [aws_iam_role_policy.firehose-opensearch]

  name        = %[1]q
  destination = "opensearch"

  opensearch_configuration {
    cluster_endpoint   = "https://${aws_opensearch_domain.test_cluster.endpoint}"
    role_arn           = aws_iam_role.firehose.arn
    index_name         = "test"
    buffering_interval = 500

    s3_configuration {
      role_arn   = aws_iam_role.firehose.arn
      bucket_arn = aws_s3_bucket.bucket.arn
    }

    processing_configuration {
      enabled = false

      processors {
        type = "Lambda"

        parameters {
          parameter_name  = "LambdaArn"
          parameter_value = "${aws_lambda_function.lambda_function_test.arn}:$LATEST"
        }
        parameters {
          parameter_name  = "BufferSizeInMBs"
          parameter_value = "1.1"
        }
        parameters {
          parameter_name  = "BufferIntervalInSeconds"
          parameter_value = "70"
        }
      }
    }
  }
}`, rName))
}

func testAccDeliveryStreamConfig_opensearchErrorOutputPrefix(rName, errorOutputPrefix string) string {
	return acctest.ConfigCompose(testAccDeliveryStreamConfig_baseOpensearch(rName), fmt.Sprintf(`
resource "aws_kinesis_firehose_delivery_stream" "test" {
  depends_on = [aws_iam_role_policy.firehose-opensearch]

  name        = %[1]q
  destination = "opensearch"

  opensearch_configuration {
    domain_arn = aws_opensearch_domain.test_cluster.arn
    role_arn   = aws_iam_role.firehose.arn
    index_name = "test"

    s3_configuration {
      role_arn            = aws_iam_role.firehose.arn
      bucket_arn          = aws_s3_bucket.bucket.arn
      error_output_prefix = %[2]q
    }
  }
}
`, rName, errorOutputPrefix))
}

func testAccDeliveryStreamConfig_opensearchVPCBasic(rName string) string {
	return acctest.ConfigCompose(testAccDeliveryStreamConfig_baseOpensearchVPC(rName), fmt.Sprintf(`
resource "aws_kinesis_firehose_delivery_stream" "test" {
  depends_on = [aws_iam_role_policy.firehose-opensearch]

  name        = %[1]q
  destination = "opensearch"

  opensearch_configuration {
    domain_arn = aws_opensearch_domain.test_cluster.arn
    role_arn   = aws_iam_role.firehose.arn
    index_name = "test"

    s3_configuration {
      role_arn   = aws_iam_role.firehose.arn
      bucket_arn = aws_s3_bucket.bucket.arn
    }

    vpc_config {
      subnet_ids         = aws_subnet.test[*].id
      security_group_ids = aws_security_group.test[*].id
      role_arn           = aws_iam_role.firehose.arn
    }
  }
}
`, rName))
}

func testAccDeliveryStreamConfig_opensearchUpdate(rName string) string {
	return acctest.ConfigCompose(
		testAccDeliveryStreamConfig_baseLambda(rName),
		testAccDeliveryStreamConfig_baseOpensearch(rName),
		fmt.Sprintf(`
resource "aws_kinesis_firehose_delivery_stream" "test" {
  depends_on = [aws_iam_role_policy.firehose-opensearch]

  name        = %[1]q
  destination = "opensearch"

  opensearch_configuration {
    domain_arn         = aws_opensearch_domain.test_cluster.arn
    role_arn           = aws_iam_role.firehose.arn
    index_name         = "test"
    buffering_interval = 500

    s3_configuration {
      role_arn   = aws_iam_role.firehose.arn
      bucket_arn = aws_s3_bucket.bucket.arn
    }

    processing_configuration {
      enabled = false

      processors {
        type = "Lambda"

        parameters {
          parameter_name  = "LambdaArn"
          parameter_value = "${aws_lambda_function.lambda_function_test.arn}:$LATEST"
        }
        parameters {
          parameter_name  = "BufferSizeInMBs"
          parameter_value = "1.1"
        }
        parameters {
          parameter_name  = "BufferIntervalInSeconds"
          parameter_value = "70"
        }
      }
    }
  }
}
`, rName))
}

func testAccDeliveryStreamConfig_opensearchVPCUpdate(rName string) string {
	return acctest.ConfigCompose(
		testAccDeliveryStreamConfig_baseLambda(rName),
		testAccDeliveryStreamConfig_baseOpensearchVPC(rName),
		fmt.Sprintf(`
resource "aws_kinesis_firehose_delivery_stream" "test" {
  depends_on = [aws_iam_role_policy.firehose-opensearch]

  name        = %[1]q
  destination = "opensearch"

  opensearch_configuration {
    domain_arn         = aws_opensearch_domain.test_cluster.arn
    role_arn           = aws_iam_role.firehose.arn
    index_name         = "test"
    buffering_interval = 500

    s3_configuration {
      role_arn   = aws_iam_role.firehose.arn
      bucket_arn = aws_s3_bucket.bucket.arn
    }

    vpc_config {
      subnet_ids         = aws_subnet.test[*].id
      security_group_ids = aws_security_group.test[*].id
      role_arn           = aws_iam_role.firehose.arn
    }

    processing_configuration {
      enabled = false
      processors {
        type = "Lambda"
        parameters {
          parameter_name  = "LambdaArn"
          parameter_value = "${aws_lambda_function.lambda_function_test.arn}:$LATEST"
        }
        parameters {
          parameter_name  = "BufferSizeInMBs"
          parameter_value = "1.1"
        }
        parameters {
          parameter_name  = "BufferIntervalInSeconds"
          parameter_value = "70"
        }
      }
    }
  }
}`, rName))
}

func testAccDeliveryStreamConfig_baseOpenSearchServerless(rName string) string {
	return acctest.ConfigCompose(testAccDeliveryStreamConfig_base(rName), fmt.Sprintf(`
resource "aws_opensearchserverless_security_policy" "test" {
  name = %[1]q
  type = "encryption"
  policy = jsonencode({
    "Rules" = [
      {
        "Resource" = [
          "collection/%[1]s"
        ],
        "ResourceType" = "collection"
      }
    ],
    "AWSOwnedKey" = true
  })
}

resource "aws_opensearchserverless_collection" "test" {
  name = %[1]q

  depends_on = [aws_opensearchserverless_security_policy.test]
}

resource "aws_iam_role_policy" "firehose-aoss" {
  name   = "%[1]s-aoss"
  role   = aws_iam_role.firehose.id
  policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Action": [
        "aoss:*"
      ],
      "Resource": [
        "${aws_opensearchserverless_collection.test.arn}"
      ]
    }
  ]
}
EOF
}
`, rName))
}

func testAccDeliveryStreamConfig_openSearchServerlessBasic(rName string) string {
	return acctest.ConfigCompose(testAccDeliveryStreamConfig_baseOpenSearchServerless(rName), fmt.Sprintf(`
resource "aws_kinesis_firehose_delivery_stream" "test" {
  depends_on = [aws_iam_role_policy.firehose-aoss]

  name        = %[1]q
  destination = "opensearchserverless"

  opensearchserverless_configuration {
    collection_endpoint = aws_opensearchserverless_collection.test.collection_endpoint
    role_arn            = aws_iam_role.firehose.arn
    index_name          = "test"

    s3_configuration {
      role_arn   = aws_iam_role.firehose.arn
      bucket_arn = aws_s3_bucket.bucket.arn
    }
  }
}
`, rName))
}

func testAccDeliveryStreamConfig_openSearchServerlessUpdate(rName string) string {
	return acctest.ConfigCompose(testAccDeliveryStreamConfig_baseOpenSearchServerless(rName), testAccDeliveryStreamConfig_baseLambda(rName), fmt.Sprintf(`
resource "aws_kinesis_firehose_delivery_stream" "test" {
  depends_on = [aws_iam_role_policy.firehose-aoss]

  name        = %[1]q
  destination = "opensearchserverless"

  opensearchserverless_configuration {
    collection_endpoint = aws_opensearchserverless_collection.test.collection_endpoint
    role_arn            = aws_iam_role.firehose.arn
    index_name          = "test"
    buffering_interval  = 500
    buffering_size      = 10

    s3_configuration {
      role_arn   = aws_iam_role.firehose.arn
      bucket_arn = aws_s3_bucket.bucket.arn
    }

    processing_configuration {
      enabled = false

      processors {
        type = "Lambda"

        parameters {
          parameter_name  = "LambdaArn"
          parameter_value = "${aws_lambda_function.lambda_function_test.arn}:$LATEST"
        }
        parameters {
          parameter_name  = "BufferSizeInMBs"
          parameter_value = "1.1"
        }
        parameters {
          parameter_name  = "BufferIntervalInSeconds"
          parameter_value = "70"
        }
      }
    }
  }
}
`, rName))
}

func testAccDeliveryStreamConfig_missingProcessingConfiguration(rName string) string {
	return fmt.Sprintf(`
data "aws_partition" "current" {}

resource "aws_iam_role" "firehose" {
  name = %[1]q

  assume_role_policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Sid": "",
      "Effect": "Allow",
      "Principal": {
        "Service": "firehose.amazonaws.com"
      },
      "Action": "sts:AssumeRole"
    }
  ]
}
EOF
}

resource "aws_iam_role_policy" "firehose" {
  name = %[1]q
  role = aws_iam_role.firehose.id

  policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Sid": "",
      "Effect": "Allow",
      "Action": [
        "s3:AbortMultipartUpload",
        "s3:GetBucketLocation",
        "s3:GetObject",
        "s3:ListBucket",
        "s3:ListBucketMultipartUploads",
        "s3:PutObject"
      ],
      "Resource": [
        "${aws_s3_bucket.bucket.arn}",
        "${aws_s3_bucket.bucket.arn}/*"
      ]
    },
    {
      "Effect": "Allow",
      "Action": [
        "logs:putLogEvents"
      ],
      "Resource": [
        "arn:${data.aws_partition.current.partition}:logs::log-group:/aws/kinesisfirehose/*"
      ]
    }
  ]
}
EOF
}

resource "aws_s3_bucket" "bucket" {
  bucket = %[1]q
}

resource "aws_kinesis_firehose_delivery_stream" "test" {
  name        = %[1]q
  destination = "extended_s3"

  extended_s3_configuration {
    role_arn           = aws_iam_role.firehose.arn
    prefix             = "tracking/autocomplete_stream/"
    buffering_interval = 300
    buffering_size     = 5
    compression_format = "GZIP"
    bucket_arn         = aws_s3_bucket.bucket.arn
  }
}
`, rName)
}
