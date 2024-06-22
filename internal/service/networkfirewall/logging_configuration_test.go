// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package networkfirewall_test

import (
	"context"
	"fmt"
	"testing"

	awstypes "github.com/aws/aws-sdk-go-v2/service/networkfirewall/types"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfnetworkfirewall "github.com/hashicorp/terraform-provider-aws/internal/service/networkfirewall"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccNetworkFirewallLoggingConfiguration_CloudWatchLogDestination_logGroup(t *testing.T) {
	ctx := acctest.Context(t)
	logGroupName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	updatedLogGroupName := fmt.Sprintf("%s-updated", logGroupName)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_networkfirewall_logging_configuration.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.NetworkFirewallServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckLoggingConfigurationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccLoggingConfigurationConfig_cloudWatch(logGroupName, rName, string(awstypes.LogDestinationTypeCloudwatchLogs), string(awstypes.LogTypeFlow)),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLoggingConfigurationExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "logging_configuration.#", acctest.Ct1),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "logging_configuration.0.log_destination_config.*", map[string]string{
						"log_destination.%":        acctest.Ct1,
						"log_destination.logGroup": logGroupName,
						"log_destination_type":     string(awstypes.LogDestinationTypeCloudwatchLogs),
					}),
				),
			},
			{
				Config: testAccLoggingConfigurationConfig_cloudWatch(updatedLogGroupName, rName, string(awstypes.LogDestinationTypeCloudwatchLogs), string(awstypes.LogTypeFlow)),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLoggingConfigurationExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "logging_configuration.#", acctest.Ct1),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "logging_configuration.0.log_destination_config.*", map[string]string{
						"log_destination.%":        acctest.Ct1,
						"log_destination.logGroup": updatedLogGroupName,
						"log_destination_type":     string(awstypes.LogDestinationTypeCloudwatchLogs),
					}),
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

func TestAccNetworkFirewallLoggingConfiguration_CloudWatchLogDestination_logType(t *testing.T) {
	ctx := acctest.Context(t)
	logGroupName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_networkfirewall_logging_configuration.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.NetworkFirewallServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckLoggingConfigurationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccLoggingConfigurationConfig_cloudWatch(logGroupName, rName, string(awstypes.LogDestinationTypeCloudwatchLogs), string(awstypes.LogTypeFlow)),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLoggingConfigurationExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "logging_configuration.#", acctest.Ct1),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "logging_configuration.0.log_destination_config.*", map[string]string{
						"log_type": string(awstypes.LogTypeFlow),
					}),
				),
			},
			{
				Config: testAccLoggingConfigurationConfig_cloudWatch(logGroupName, rName, string(awstypes.LogDestinationTypeCloudwatchLogs), string(awstypes.LogTypeAlert)),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLoggingConfigurationExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "logging_configuration.#", acctest.Ct1),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "logging_configuration.0.log_destination_config.*", map[string]string{
						"log_type": string(awstypes.LogTypeAlert),
					}),
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

func TestAccNetworkFirewallLoggingConfiguration_KinesisLogDestination_deliveryStream(t *testing.T) {
	ctx := acctest.Context(t)
	streamName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	updatedStreamName := fmt.Sprintf("%s-updated", streamName)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_networkfirewall_logging_configuration.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.NetworkFirewallServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckLoggingConfigurationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccLoggingConfigurationConfig_kinesis(streamName, rName, string(awstypes.LogDestinationTypeKinesisDataFirehose), string(awstypes.LogTypeFlow)),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLoggingConfigurationExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "logging_configuration.#", acctest.Ct1),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "logging_configuration.0.log_destination_config.*", map[string]string{
						"log_destination.%":              acctest.Ct1,
						"log_destination.deliveryStream": streamName,
						"log_destination_type":           string(awstypes.LogDestinationTypeKinesisDataFirehose),
					}),
				),
			},
			{
				Config: testAccLoggingConfigurationConfig_kinesis(updatedStreamName, rName, string(awstypes.LogDestinationTypeKinesisDataFirehose), string(awstypes.LogTypeFlow)),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLoggingConfigurationExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "logging_configuration.#", acctest.Ct1),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "logging_configuration.0.log_destination_config.*", map[string]string{
						"log_destination.%":              acctest.Ct1,
						"log_destination.deliveryStream": updatedStreamName,
						"log_destination_type":           string(awstypes.LogDestinationTypeKinesisDataFirehose),
					}),
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

func TestAccNetworkFirewallLoggingConfiguration_KinesisLogDestination_logType(t *testing.T) {
	ctx := acctest.Context(t)
	streamName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_networkfirewall_logging_configuration.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.NetworkFirewallServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckLoggingConfigurationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccLoggingConfigurationConfig_kinesis(streamName, rName, string(awstypes.LogDestinationTypeKinesisDataFirehose), string(awstypes.LogTypeFlow)),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLoggingConfigurationExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "logging_configuration.#", acctest.Ct1),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "logging_configuration.0.log_destination_config.*", map[string]string{
						"log_type": string(awstypes.LogTypeFlow),
					}),
				),
			},
			{
				Config: testAccLoggingConfigurationConfig_kinesis(streamName, rName, string(awstypes.LogDestinationTypeKinesisDataFirehose), string(awstypes.LogTypeAlert)),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLoggingConfigurationExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "logging_configuration.#", acctest.Ct1),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "logging_configuration.0.log_destination_config.*", map[string]string{
						"log_type": string(awstypes.LogTypeAlert),
					}),
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

func TestAccNetworkFirewallLoggingConfiguration_S3LogDestination_bucketName(t *testing.T) {
	ctx := acctest.Context(t)
	bucketName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	updatedBucketName := fmt.Sprintf("%s-updated", bucketName)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_networkfirewall_logging_configuration.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.NetworkFirewallServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckLoggingConfigurationDestroy(ctx),
		Steps: []resource.TestStep{

			{
				Config: testAccLoggingConfigurationConfig_s3(bucketName, rName, string(awstypes.LogDestinationTypeS3), string(awstypes.LogTypeFlow)),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLoggingConfigurationExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "logging_configuration.#", acctest.Ct1),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "logging_configuration.0.log_destination_config.*", map[string]string{
						"log_destination.%":          acctest.Ct1,
						"log_destination.bucketName": bucketName,
					}),
				),
			},
			{
				Config: testAccLoggingConfigurationConfig_s3(updatedBucketName, rName, string(awstypes.LogDestinationTypeS3), string(awstypes.LogTypeFlow)),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLoggingConfigurationExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "logging_configuration.#", acctest.Ct1),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "logging_configuration.0.log_destination_config.*", map[string]string{
						"log_destination.%":          acctest.Ct1,
						"log_destination.bucketName": updatedBucketName,
					}),
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

func TestAccNetworkFirewallLoggingConfiguration_S3LogDestination_logType(t *testing.T) {
	ctx := acctest.Context(t)
	bucketName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_networkfirewall_logging_configuration.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.NetworkFirewallServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckLoggingConfigurationDestroy(ctx),
		Steps: []resource.TestStep{

			{
				Config: testAccLoggingConfigurationConfig_s3(bucketName, rName, string(awstypes.LogDestinationTypeS3), string(awstypes.LogTypeFlow)),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLoggingConfigurationExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "logging_configuration.#", acctest.Ct1),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "logging_configuration.0.log_destination_config.*", map[string]string{
						"log_type": string(awstypes.LogTypeFlow),
					}),
				),
			},
			{
				Config: testAccLoggingConfigurationConfig_s3(bucketName, rName, string(awstypes.LogDestinationTypeS3), string(awstypes.LogTypeAlert)),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLoggingConfigurationExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "logging_configuration.#", acctest.Ct1),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "logging_configuration.0.log_destination_config.*", map[string]string{
						"log_type": string(awstypes.LogTypeAlert),
					}),
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

func TestAccNetworkFirewallLoggingConfiguration_S3LogDestination_prefix(t *testing.T) {
	ctx := acctest.Context(t)
	bucketName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_networkfirewall_logging_configuration.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.NetworkFirewallServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckLoggingConfigurationDestroy(ctx),
		Steps: []resource.TestStep{

			{
				Config: testAccLoggingConfigurationConfig_s3(bucketName, rName, string(awstypes.LogDestinationTypeS3), string(awstypes.LogTypeFlow)),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLoggingConfigurationExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "logging_configuration.#", acctest.Ct1),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "logging_configuration.0.log_destination_config.*", map[string]string{
						"log_destination.%":          acctest.Ct1,
						"log_destination.bucketName": bucketName,
					}),
				),
			},
			{
				Config: testAccLoggingConfigurationConfig_s3UpdatePrefix(bucketName, rName, string(awstypes.LogDestinationTypeS3), string(awstypes.LogTypeFlow)),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLoggingConfigurationExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "logging_configuration.#", acctest.Ct1),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "logging_configuration.0.log_destination_config.*", map[string]string{
						"log_destination.%":          acctest.Ct2,
						"log_destination.bucketName": bucketName,
						"log_destination.prefix":     "update-prefix",
					}),
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

func TestAccNetworkFirewallLoggingConfiguration_updateFirewallARN(t *testing.T) {
	ctx := acctest.Context(t)
	bucketName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_networkfirewall_logging_configuration.test"
	firewallResourceName := "aws_networkfirewall_firewall.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.NetworkFirewallServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckLoggingConfigurationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccLoggingConfigurationConfig_s3(bucketName, rName, string(awstypes.LogDestinationTypeS3), string(awstypes.LogTypeFlow)),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLoggingConfigurationExists(ctx, resourceName),
					resource.TestCheckResourceAttrPair(resourceName, "firewall_arn", firewallResourceName, names.AttrARN),
				),
			},
			{
				// ForceNew Firewall i.e. LoggingConfiguration Resource
				Config: testAccLoggingConfigurationConfig_s3UpdateFirewallARN(bucketName, rName, string(awstypes.LogDestinationTypeS3), string(awstypes.LogTypeFlow)),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLoggingConfigurationExists(ctx, resourceName),
					resource.TestCheckResourceAttrPair(resourceName, "firewall_arn", firewallResourceName, names.AttrARN),
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

func TestAccNetworkFirewallLoggingConfiguration_updateLogDestinationType(t *testing.T) {
	ctx := acctest.Context(t)
	bucketName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	logGroupName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	streamName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_networkfirewall_logging_configuration.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.NetworkFirewallServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckLoggingConfigurationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccLoggingConfigurationConfig_cloudWatch(logGroupName, rName, string(awstypes.LogDestinationTypeCloudwatchLogs), string(awstypes.LogTypeFlow)),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLoggingConfigurationExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "logging_configuration.#", acctest.Ct1),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "logging_configuration.0.log_destination_config.*", map[string]string{
						"log_destination.%":        acctest.Ct1,
						"log_destination.logGroup": logGroupName,
						"log_destination_type":     string(awstypes.LogDestinationTypeCloudwatchLogs),
						"log_type":                 string(awstypes.LogTypeFlow),
					}),
				),
			},
			{
				Config: testAccLoggingConfigurationConfig_kinesis(streamName, rName, string(awstypes.LogDestinationTypeKinesisDataFirehose), string(awstypes.LogTypeFlow)),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLoggingConfigurationExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "logging_configuration.#", acctest.Ct1),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "logging_configuration.0.log_destination_config.*", map[string]string{
						"log_destination.%":              acctest.Ct1,
						"log_destination.deliveryStream": streamName,
						"log_destination_type":           string(awstypes.LogDestinationTypeKinesisDataFirehose),
						"log_type":                       string(awstypes.LogTypeFlow),
					}),
				),
			},
			{
				Config: testAccLoggingConfigurationConfig_s3(bucketName, rName, string(awstypes.LogDestinationTypeS3), string(awstypes.LogTypeFlow)),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLoggingConfigurationExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "logging_configuration.#", acctest.Ct1),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "logging_configuration.0.log_destination_config.*", map[string]string{
						"log_destination_type": string(awstypes.LogDestinationTypeS3),
						"log_type":             string(awstypes.LogTypeFlow),
					}),
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

func TestAccNetworkFirewallLoggingConfiguration_updateToMultipleLogDestinations(t *testing.T) {
	ctx := acctest.Context(t)
	bucketName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	streamName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_networkfirewall_logging_configuration.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.NetworkFirewallServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckLoggingConfigurationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccLoggingConfigurationConfig_s3(bucketName, rName, string(awstypes.LogDestinationTypeS3), string(awstypes.LogTypeAlert)),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLoggingConfigurationExists(ctx, resourceName),
				),
			},
			{
				Config: testAccLoggingConfigurationConfig_s3AndKinesis(bucketName, streamName, rName, string(awstypes.LogTypeAlert), string(awstypes.LogTypeFlow)),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLoggingConfigurationExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "logging_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "logging_configuration.0.log_destination_config.#", acctest.Ct2),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "logging_configuration.0.log_destination_config.*", map[string]string{
						"log_destination.%":              acctest.Ct1,
						"log_destination.deliveryStream": streamName,
						"log_destination_type":           string(awstypes.LogDestinationTypeKinesisDataFirehose),
						"log_type":                       string(awstypes.LogTypeFlow),
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "logging_configuration.0.log_destination_config.*", map[string]string{
						"log_destination.%":          acctest.Ct1,
						"log_destination.bucketName": bucketName,
						"log_destination_type":       string(awstypes.LogDestinationTypeS3),
						"log_type":                   string(awstypes.LogTypeAlert),
					}),
				),
			},
			{
				Config: testAccLoggingConfigurationConfig_s3(bucketName, rName, string(awstypes.LogDestinationTypeS3), string(awstypes.LogTypeAlert)),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLoggingConfigurationExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "logging_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "logging_configuration.0.log_destination_config.#", acctest.Ct1),
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

func TestAccNetworkFirewallLoggingConfiguration_updateToSingleAlertTypeLogDestination(t *testing.T) {
	ctx := acctest.Context(t)
	bucketName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	logGroupName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_networkfirewall_logging_configuration.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.NetworkFirewallServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckLoggingConfigurationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccLoggingConfigurationConfig_s3AndCloudWatch(bucketName, logGroupName, rName, string(awstypes.LogTypeAlert), string(awstypes.LogTypeFlow)),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLoggingConfigurationExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "logging_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "logging_configuration.0.log_destination_config.#", acctest.Ct2),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "logging_configuration.0.log_destination_config.*", map[string]string{
						"log_destination.%":        acctest.Ct1,
						"log_destination.logGroup": logGroupName,
						"log_destination_type":     string(awstypes.LogDestinationTypeCloudwatchLogs),
						"log_type":                 string(awstypes.LogTypeFlow),
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "logging_configuration.0.log_destination_config.*", map[string]string{
						"log_destination.%":          acctest.Ct1,
						"log_destination.bucketName": bucketName,
						"log_destination_type":       string(awstypes.LogDestinationTypeS3),
						"log_type":                   string(awstypes.LogTypeAlert),
					}),
				),
			},
			{
				Config: testAccLoggingConfigurationConfig_s3(bucketName, rName, string(awstypes.LogDestinationTypeS3), string(awstypes.LogTypeAlert)),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLoggingConfigurationExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "logging_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "logging_configuration.0.log_destination_config.#", acctest.Ct1),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "logging_configuration.0.log_destination_config.*", map[string]string{
						"log_destination.%":          acctest.Ct1,
						"log_destination.bucketName": bucketName,
						"log_destination_type":       string(awstypes.LogDestinationTypeS3),
						"log_type":                   string(awstypes.LogTypeAlert),
					}),
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

func TestAccNetworkFirewallLoggingConfiguration_updateToSingleFlowTypeLogDestination(t *testing.T) {
	ctx := acctest.Context(t)
	bucketName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	logGroupName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_networkfirewall_logging_configuration.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.NetworkFirewallServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckLoggingConfigurationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccLoggingConfigurationConfig_s3AndCloudWatch(bucketName, logGroupName, rName, string(awstypes.LogTypeAlert), string(awstypes.LogTypeFlow)),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLoggingConfigurationExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "logging_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "logging_configuration.0.log_destination_config.#", acctest.Ct2),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "logging_configuration.0.log_destination_config.*", map[string]string{
						"log_destination.%":        acctest.Ct1,
						"log_destination.logGroup": logGroupName,
						"log_destination_type":     string(awstypes.LogDestinationTypeCloudwatchLogs),
						"log_type":                 string(awstypes.LogTypeFlow),
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "logging_configuration.0.log_destination_config.*", map[string]string{
						"log_destination.%":          acctest.Ct1,
						"log_destination.bucketName": bucketName,
						"log_destination_type":       string(awstypes.LogDestinationTypeS3),
						"log_type":                   string(awstypes.LogTypeAlert),
					}),
				),
			},
			{
				Config: testAccLoggingConfigurationConfig_cloudWatch(logGroupName, rName, string(awstypes.LogDestinationTypeCloudwatchLogs), string(awstypes.LogTypeFlow)),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLoggingConfigurationExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "logging_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "logging_configuration.0.log_destination_config.#", acctest.Ct1),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "logging_configuration.0.log_destination_config.*", map[string]string{
						"log_destination.%":        acctest.Ct1,
						"log_destination.logGroup": logGroupName,
						"log_destination_type":     string(awstypes.LogDestinationTypeCloudwatchLogs),
						"log_type":                 string(awstypes.LogTypeFlow),
					}),
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

func TestAccNetworkFirewallLoggingConfiguration_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	bucketName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_networkfirewall_logging_configuration.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.NetworkFirewallServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckLoggingConfigurationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccLoggingConfigurationConfig_s3(bucketName, rName, string(awstypes.LogDestinationTypeS3), string(awstypes.LogTypeFlow)),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLoggingConfigurationExists(ctx, resourceName),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfnetworkfirewall.ResourceLoggingConfiguration(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckLoggingConfigurationDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_networkfirewall_logging_configuration" {
				continue
			}

			conn := acctest.Provider.Meta().(*conns.AWSClient).NetworkFirewallClient(ctx)

			_, err := tfnetworkfirewall.FindLoggingConfigurationByARN(ctx, conn, rs.Primary.ID)

			if tfresource.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("NetworkFirewall Logging Configuration %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccCheckLoggingConfigurationExists(ctx context.Context, n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).NetworkFirewallClient(ctx)

		_, err := tfnetworkfirewall.FindLoggingConfigurationByARN(ctx, conn, rs.Primary.ID)

		return err
	}
}

func testAccLoggingConfigurationConfig_base(rName string) string {
	return acctest.ConfigCompose(acctest.ConfigVPCWithSubnets(rName, 1), fmt.Sprintf(`
resource "aws_networkfirewall_firewall_policy" "test" {
  name = %[1]q

  firewall_policy {
    stateless_fragment_default_actions = ["aws:drop"]
    stateless_default_actions          = ["aws:pass"]
  }
}

resource "aws_networkfirewall_firewall" "test" {
  name                = %[1]q
  firewall_policy_arn = aws_networkfirewall_firewall_policy.test.arn
  vpc_id              = aws_vpc.test.id

  subnet_mapping {
    subnet_id = aws_subnet.test[0].id
  }
}
`, rName))
}

func testAccLoggingConfigurationConfig_baseFirewallUpdated(rName string) string {
	return acctest.ConfigCompose(acctest.ConfigVPCWithSubnets(rName, 1), fmt.Sprintf(`
resource "aws_networkfirewall_firewall_policy" "test" {
  name = %[1]q

  firewall_policy {
    stateless_fragment_default_actions = ["aws:drop"]
    stateless_default_actions          = ["aws:pass"]
  }
}

resource "aws_networkfirewall_firewall" "test" {
  name                = "%[1]s-updated"
  firewall_policy_arn = aws_networkfirewall_firewall_policy.test.arn
  vpc_id              = aws_vpc.test.id

  subnet_mapping {
    subnet_id = aws_subnet.test[0].id
  }
}
`, rName))
}

func testAccLoggingConfigurationConfig_baseS3Bucket(rName string) string {
	return fmt.Sprintf(`
resource "aws_s3_bucket" "test" {
  bucket        = %[1]q
  force_destroy = true

  lifecycle {
    create_before_destroy = true
  }
}
`, rName)
}

func testAccLoggingConfigurationConfig_baseCloudWatch(rName string) string {
	return fmt.Sprintf(`
resource "aws_cloudwatch_log_group" "test" {
  name = %[1]q

  lifecycle {
    create_before_destroy = true
  }
}
`, rName)
}

func testAccLoggingConfigurationConfig_baseFirehose(rName, streamName string) string {
	return fmt.Sprintf(`
data "aws_caller_identity" "current" {}

data "aws_partition" "current" {}

resource "aws_iam_role" "test" {
  name = %[1]q

  assume_role_policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Sid": "",
      "Effect": "Allow",
      "Principal": {
        "Service": "firehose.${data.aws_partition.current.dns_suffix}"
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

resource "aws_iam_role_policy" "test" {
  name = %[1]q
  role = aws_iam_role.test.id

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
        "${aws_s3_bucket.logs.arn}",
        "${aws_s3_bucket.logs.arn}/*"
      ]
    },
    {
      "Sid": "GlueAccess",
      "Effect": "Allow",
      "Action": [
        "glue:GetTableVersions"
      ],
      "Resource": [
        "*"
      ]
    }
  ]
}
EOF
}

resource "aws_s3_bucket" "logs" {
  bucket        = %[1]q
  force_destroy = true
}

resource "aws_kinesis_firehose_delivery_stream" "test" {
  depends_on  = [aws_iam_role_policy.test]
  name        = %[2]q
  destination = "extended_s3"

  extended_s3_configuration {
    role_arn   = aws_iam_role.test.arn
    bucket_arn = aws_s3_bucket.logs.arn
  }

  tags = {
    LogDeliveryEnabled = "placeholder"
  }

  lifecycle {
    create_before_destroy = true
    ignore_changes = [
      # Ignore changes to LogDeliveryEnabled tag as API adds this tag when broker log delivery is enabled
      tags["LogDeliveryEnabled"],
    ]
  }
}
`, rName, streamName)
}

func testAccLoggingConfigurationConfig_s3(bucketName, rName, destinationType, logType string) string {
	return acctest.ConfigCompose(
		testAccLoggingConfigurationConfig_base(rName),
		testAccLoggingConfigurationConfig_baseS3Bucket(bucketName),
		fmt.Sprintf(`
resource "aws_networkfirewall_logging_configuration" "test" {
  firewall_arn = aws_networkfirewall_firewall.test.arn

  logging_configuration {
    log_destination_config {
      log_destination = {
        bucketName = aws_s3_bucket.test.bucket
      }
      log_destination_type = %[2]q
      log_type             = %[3]q
    }
  }
}
`, rName, destinationType, logType))
}

func testAccLoggingConfigurationConfig_s3UpdateFirewallARN(bucketName, rName, destinationType, logType string) string {
	return acctest.ConfigCompose(
		testAccLoggingConfigurationConfig_baseFirewallUpdated(rName),
		testAccLoggingConfigurationConfig_baseS3Bucket(bucketName),
		fmt.Sprintf(`
resource "aws_networkfirewall_logging_configuration" "test" {
  firewall_arn = aws_networkfirewall_firewall.test.arn

  logging_configuration {
    log_destination_config {
      log_destination = {
        bucketName = aws_s3_bucket.test.bucket
      }
      log_destination_type = %[2]q
      log_type             = %[3]q
    }
  }
}
`, rName, destinationType, logType))
}

func testAccLoggingConfigurationConfig_s3UpdatePrefix(bucketName, rName, destinationType, logType string) string {
	return acctest.ConfigCompose(
		testAccLoggingConfigurationConfig_base(rName),
		testAccLoggingConfigurationConfig_baseS3Bucket(bucketName),
		fmt.Sprintf(`
resource "aws_networkfirewall_logging_configuration" "test" {
  firewall_arn = aws_networkfirewall_firewall.test.arn

  logging_configuration {
    log_destination_config {
      log_destination = {
        bucketName = aws_s3_bucket.test.bucket
        prefix     = "update-prefix"
      }
      log_destination_type = %[2]q
      log_type             = %[3]q
    }
  }
}
`, rName, destinationType, logType))
}

func testAccLoggingConfigurationConfig_kinesis(streamName, rName, destinationType, logType string) string {
	return acctest.ConfigCompose(
		testAccLoggingConfigurationConfig_base(rName),
		testAccLoggingConfigurationConfig_baseFirehose(rName, streamName),
		fmt.Sprintf(`
resource "aws_networkfirewall_logging_configuration" "test" {
  firewall_arn = aws_networkfirewall_firewall.test.arn

  logging_configuration {
    log_destination_config {
      log_destination = {
        deliveryStream = aws_kinesis_firehose_delivery_stream.test.name
      }
      log_destination_type = %[2]q
      log_type             = %[3]q
    }
  }
}
`, rName, destinationType, logType))
}

func testAccLoggingConfigurationConfig_cloudWatch(logGroupName, rName, destinationType, logType string) string {
	return acctest.ConfigCompose(
		testAccLoggingConfigurationConfig_base(rName),
		testAccLoggingConfigurationConfig_baseCloudWatch(logGroupName),
		fmt.Sprintf(`
resource "aws_networkfirewall_logging_configuration" "test" {
  firewall_arn = aws_networkfirewall_firewall.test.arn

  logging_configuration {
    log_destination_config {
      log_destination = {
        logGroup = aws_cloudwatch_log_group.test.name
      }
      log_destination_type = %[2]q
      log_type             = %[3]q
    }
  }
}
`, rName, destinationType, logType))
}

func testAccLoggingConfigurationConfig_s3AndKinesis(bucketName, streamName, rName, logTypeS3, logTypeKinesis string) string {
	return acctest.ConfigCompose(
		testAccLoggingConfigurationConfig_baseS3Bucket(bucketName),
		testAccLoggingConfigurationConfig_baseFirehose(rName, streamName),
		testAccLoggingConfigurationConfig_base(rName),
		fmt.Sprintf(`
resource "aws_networkfirewall_logging_configuration" "test" {
  firewall_arn = aws_networkfirewall_firewall.test.arn

  logging_configuration {
    log_destination_config {
      log_destination = {
        bucketName = aws_s3_bucket.test.bucket
      }
      log_destination_type = "S3"
      log_type             = %[2]q
    }

    log_destination_config {
      log_destination = {
        deliveryStream = aws_kinesis_firehose_delivery_stream.test.name
      }
      log_destination_type = "KinesisDataFirehose"
      log_type             = %[3]q
    }
  }
}
`, rName, logTypeS3, logTypeKinesis))
}

func testAccLoggingConfigurationConfig_s3AndCloudWatch(bucketName, logGroupName, rName, logTypeS3, logTypeCloudWatch string) string {
	return acctest.ConfigCompose(
		testAccLoggingConfigurationConfig_base(rName),
		testAccLoggingConfigurationConfig_baseS3Bucket(bucketName),
		testAccLoggingConfigurationConfig_baseCloudWatch(logGroupName),
		fmt.Sprintf(`
resource "aws_networkfirewall_logging_configuration" "test" {
  firewall_arn = aws_networkfirewall_firewall.test.arn

  logging_configuration {
    log_destination_config {
      log_destination = {
        bucketName = aws_s3_bucket.test.bucket
      }
      log_destination_type = "S3"
      log_type             = %[2]q
    }

    log_destination_config {
      log_destination = {
        logGroup = aws_cloudwatch_log_group.test.name
      }
      log_destination_type = "CloudWatchLogs"
      log_type             = %[3]q
    }
  }
}
`, rName, logTypeS3, logTypeCloudWatch))
}
