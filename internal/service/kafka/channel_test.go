// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package kafka_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/service/kafka"
	awstypes "github.com/aws/aws-sdk-go-v2/service/kafka/types"
	"github.com/hashicorp/terraform-plugin-testing/config"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	tfknownvalue "github.com/hashicorp/terraform-provider-aws/internal/acctest/knownvalue"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	tfkafka "github.com/hashicorp/terraform-provider-aws/internal/service/kafka"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

const (
	clusterUUIDRegexPattern = verify.UUIDRegexPattern + `-\d+`
)

// arn:<partition>:kafka:<region>:<account-id>:channel/<cluster-name>/<cluster-uuid>/<channel-name>/<channel-uuid>.
// e.g. "arn:aws:kafka:us-west-2:123456789012:channel/tf-acc-test-9184561384065770181/db5597af-dd90-447d-8095-3cc4f952da64-11/tf-acc-test-9184561384065770181/270f019d-e5eb-472f-b5fc-3d89a408a70a".
func channelARNResourceRegexPattern(clusterName, channelName string) string {
	return `channel/` + clusterName + `/` + clusterUUIDRegexPattern + `/` + channelName + `/` + verify.UUIDRegexPattern
}

func checkChannelARN(clusterName, channelName string) knownvalue.Check {
	// arn:<partition>:kafka:<region>:<account-id>:channel/<cluster-name>/<cluster-uuid>/<channel-name>/<channel-uuid>.
	return tfknownvalue.RegionalARNRegexp("kafka", regexache.MustCompile(channelARNResourceRegexPattern(clusterName, channelName)))
}

func TestAccKafkaChannel_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var v kafka.DescribeChannelOutput
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_msk_channel.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.KafkaServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckChannelDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				ConfigDirectory: config.StaticDirectory("testdata/Channel/basic/"),
				ConfigVariables: config.Variables{
					acctest.CtRName: config.StringVariable(rName),
				},
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckChannelExists(ctx, t, resourceName, &v),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrARN), checkChannelARN(rName+"-cluster", rName)),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("channel_name"), knownvalue.StringExact(rName)),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("cluster_arn"), knownvalue.NotNull()),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("destination_type"), tfknownvalue.StringExact(awstypes.ChannelDestinationTypeS3)),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrEncryptionConfiguration), knownvalue.ListSizeExact(0)),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("iceberg_destination"), knownvalue.ListSizeExact(0)),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("logging_info"), knownvalue.ListSizeExact(0)),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("s3_destination"), knownvalue.ListSizeExact(1)),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrTags), knownvalue.Null()),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("topic_configuration"), knownvalue.ListSizeExact(1)),
				},
			},
			{
				ConfigDirectory: config.StaticDirectory("testdata/Channel/basic/"),
				ConfigVariables: config.Variables{
					acctest.CtRName: config.StringVariable(rName),
				},
				ImportStateIdFunc:                    testAccChannelImportStateIDFunc(resourceName),
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: names.AttrARN,
			},
		},
	})
}

func TestAccKafkaChannel_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	var v kafka.DescribeChannelOutput
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_msk_channel.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.KafkaServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckChannelDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				ConfigDirectory: config.StaticDirectory("testdata/Channel/basic/"),
				ConfigVariables: config.Variables{
					acctest.CtRName: config.StringVariable(rName),
				},
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckChannelExists(ctx, t, resourceName, &v),
					acctest.CheckFrameworkResourceDisappears(ctx, t, tfkafka.ResourceChannel, resourceName),
				),
				ExpectNonEmptyPlan: true,
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
					},
					PostApplyPostRefresh: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
					},
				},
			},
		},
	})
}

func TestAccKafkaChannel_iceberg(t *testing.T) {
	ctx := acctest.Context(t)
	var v kafka.DescribeChannelOutput
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_msk_channel.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.KafkaServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckChannelDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				ConfigDirectory: config.StaticDirectory("testdata/Channel/iceberg_destination/"),
				ConfigVariables: config.Variables{
					acctest.CtRName: config.StringVariable(rName),
				},
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckChannelExists(ctx, t, resourceName, &v),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("destination_type"), tfknownvalue.StringExact(awstypes.ChannelDestinationTypeIceberg)),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("iceberg_destination"), knownvalue.ListSizeExact(1)),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("s3_destination"), knownvalue.ListSizeExact(0)),
				},
			},
			{
				ConfigDirectory: config.StaticDirectory("testdata/Channel/iceberg_destination/"),
				ConfigVariables: config.Variables{
					acctest.CtRName: config.StringVariable(rName),
				},
				ImportStateIdFunc:                    testAccChannelImportStateIDFunc(resourceName),
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: names.AttrARN,
			},
		},
	})
}

func TestAccKafkaChannel_s3DataFreshness(t *testing.T) {
	ctx := acctest.Context(t)
	var v kafka.DescribeChannelOutput
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_msk_channel.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.KafkaServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckChannelDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				ConfigDirectory: config.StaticDirectory("testdata/Channel/s3_destination.data_freshness_in_seconds/"),
				ConfigVariables: config.Variables{
					acctest.CtRName:             config.StringVariable(rName),
					"data_freshness_in_seconds": config.IntegerVariable(300),
				},
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckChannelExists(ctx, t, resourceName, &v),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("destination_type"), tfknownvalue.StringExact(awstypes.ChannelDestinationTypeS3)),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("iceberg_destination"), knownvalue.ListSizeExact(0)),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("s3_destination"), knownvalue.ListExact([]knownvalue.Check{knownvalue.ObjectPartial(map[string]knownvalue.Check{
						"data_freshness_in_seconds": knownvalue.Int32Exact(300),
					})})),
				},
			},
			{
				ConfigDirectory: config.StaticDirectory("testdata/Channel/s3_destination.data_freshness_in_seconds/"),
				ConfigVariables: config.Variables{
					acctest.CtRName: config.StringVariable(rName),
				},
				ImportStateIdFunc:                    testAccChannelImportStateIDFunc(resourceName),
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: names.AttrARN,
			},
			{
				ConfigDirectory: config.StaticDirectory("testdata/Channel/s3_destination.data_freshness_in_seconds/"),
				ConfigVariables: config.Variables{
					acctest.CtRName:             config.StringVariable(rName),
					"data_freshness_in_seconds": config.IntegerVariable(600),
				},
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckChannelExists(ctx, t, resourceName, &v),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionUpdate),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("destination_type"), tfknownvalue.StringExact(awstypes.ChannelDestinationTypeS3)),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("iceberg_destination"), knownvalue.ListSizeExact(0)),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("s3_destination"), knownvalue.ListExact([]knownvalue.Check{knownvalue.ObjectPartial(map[string]knownvalue.Check{
						"data_freshness_in_seconds": knownvalue.Int32Exact(600),
					})})),
				},
			},
			{
				ConfigDirectory: config.StaticDirectory("testdata/Channel/s3_destination.data_freshness_in_seconds/"),
				ConfigVariables: config.Variables{
					acctest.CtRName: config.StringVariable(rName),
				},
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckChannelExists(ctx, t, resourceName, &v),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionNoop),
					},
				},
			},
		},
	})
}

func TestAccKafkaChannel_encryptionConfiguration(t *testing.T) {
	ctx := acctest.Context(t)
	var v kafka.DescribeChannelOutput
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_msk_channel.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.KafkaServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckChannelDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				ConfigDirectory: config.StaticDirectory("testdata/Channel/encryption_configuration/"),
				ConfigVariables: config.Variables{
					acctest.CtRName: config.StringVariable(rName),
				},
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckChannelExists(ctx, t, resourceName, &v),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("destination_type"), tfknownvalue.StringExact(awstypes.ChannelDestinationTypeS3)),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrEncryptionConfiguration), knownvalue.ListExact([]knownvalue.Check{knownvalue.ObjectExact(map[string]knownvalue.Check{
						names.AttrKMSKeyARN: knownvalue.NotNull(),
					})})),
				},
			},
			{
				ConfigDirectory: config.StaticDirectory("testdata/Channel/encryption_configuration/"),
				ConfigVariables: config.Variables{
					acctest.CtRName: config.StringVariable(rName),
				},
				ImportStateIdFunc:                    testAccChannelImportStateIDFunc(resourceName),
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: names.AttrARN,
			},
		},
	})
}

func TestAccKafkaChannel_loggingInfo(t *testing.T) {
	ctx := acctest.Context(t)
	var v kafka.DescribeChannelOutput
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_msk_channel.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.KafkaServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckChannelDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				ConfigDirectory: config.StaticDirectory("testdata/Channel/logging_info/"),
				ConfigVariables: config.Variables{
					acctest.CtRName: config.StringVariable(rName),
				},
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckChannelExists(ctx, t, resourceName, &v),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("destination_type"), tfknownvalue.StringExact(awstypes.ChannelDestinationTypeS3)),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("logging_info"), knownvalue.ListExact([]knownvalue.Check{knownvalue.ObjectExact(map[string]knownvalue.Check{
						names.AttrCloudWatchLogs: knownvalue.ListSizeExact(1),
						"firehose":               knownvalue.ListSizeExact(0),
						"s3":                     knownvalue.ListSizeExact(0),
					})})),
				},
			},
			{
				ConfigDirectory: config.StaticDirectory("testdata/Channel/logging_info/"),
				ConfigVariables: config.Variables{
					acctest.CtRName: config.StringVariable(rName),
				},
				ImportStateIdFunc:                    testAccChannelImportStateIDFunc(resourceName),
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: names.AttrARN,
			},
		},
	})
}

func testAccChannelImportStateIDFunc(resourceName string) resource.ImportStateIdFunc {
	return acctest.AttrsImportStateIdFunc(resourceName, ",", names.AttrARN, "cluster_arn")
}

func testAccCheckChannelDestroy(ctx context.Context, t *testing.T) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.ProviderMeta(ctx, t).KafkaClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_msk_channel" {
				continue
			}

			_, err := tfkafka.FindChannelByTwoPartKey(ctx, conn, rs.Primary.Attributes[names.AttrARN], rs.Primary.Attributes["cluster_arn"])

			if retry.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("MSK Channel %s still exists", rs.Primary.Attributes[names.AttrARN])
		}

		return nil
	}
}

func testAccCheckChannelExists(ctx context.Context, t *testing.T, n string, v *kafka.DescribeChannelOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.ProviderMeta(ctx, t).KafkaClient(ctx)

		output, err := tfkafka.FindChannelByTwoPartKey(ctx, conn, rs.Primary.Attributes[names.AttrARN], rs.Primary.Attributes["cluster_arn"])
		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}
