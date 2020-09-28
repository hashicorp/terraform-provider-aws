package aws

import (
	"fmt"
	"log"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/kinesisanalyticsv2"
	"github.com/hashicorp/go-multierror"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/service/kinesisanalyticsv2/finder"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/tfawsresource"
)

func init() {
	resource.AddTestSweepers("aws_kinesisanalyticsv2_application", &resource.Sweeper{
		Name: "aws_kinesisanalyticsv2_application",
		F:    testSweepKinesisAnalyticsV2Application,
	})
}

func testSweepKinesisAnalyticsV2Application(region string) error {
	client, err := sharedClientForRegion(region)
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}
	conn := client.(*AWSClient).kinesisanalyticsv2conn
	input := &kinesisanalyticsv2.ListApplicationsInput{}
	var sweeperErrs *multierror.Error

	for {
		output, err := conn.ListApplications(input)
		if testSweepSkipSweepError(err) {
			log.Printf("[WARN] Skipping Kinesis Analytics v2 Application sweep for %s: %s", region, err)
			return nil
		}
		if err != nil {
			return fmt.Errorf("error retrieving Kinesis Analytics v2 Applications: %w", err)
		}

		for _, applicationSummary := range output.ApplicationSummaries {
			arn := aws.StringValue(applicationSummary.ApplicationARN)
			name := aws.StringValue(applicationSummary.ApplicationName)

			application, err := finder.ApplicationByName(conn, name)

			if err != nil {
				sweeperErr := fmt.Errorf("error reading Kinesis Analytics v2 Application (%s): %w", arn, err)
				log.Printf("[ERROR] %s", err)
				sweeperErrs = multierror.Append(sweeperErrs, sweeperErr)
				continue
			}

			log.Printf("[INFO] Deleting Kinesis Analytics v2 Application: %s", arn)
			r := resourceAwsKinesisAnalyticsV2Application()
			d := r.Data(nil)
			d.SetId(arn)
			d.Set("create_timestamp", aws.TimeValue(application.CreateTimestamp).Format(time.RFC3339))
			d.Set("name", name)
			err = r.Delete(d, client)

			if err != nil {
				log.Printf("[ERROR] %s", err)
				sweeperErrs = multierror.Append(sweeperErrs, err)
				continue
			}
		}

		if aws.StringValue(output.NextToken) == "" {
			break
		}
		input.NextToken = output.NextToken
	}

	return sweeperErrs.ErrorOrNil()
}

func TestAccAWSKinesisAnalyticsV2Application_basic(t *testing.T) {
	var v kinesisanalyticsv2.ApplicationDetail
	resourceName := "aws_kinesisanalyticsv2_application.test"
	iamRoleResourceName := "aws_iam_role.test.0"
	rName := acctest.RandomWithPrefix("tf-acc-test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSKinesisAnalyticsV2(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckKinesisAnalyticsV2ApplicationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccKinesisAnalyticsV2ApplicationConfigBasic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckKinesisAnalyticsV2ApplicationExists(resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.#", "0"),
					testAccCheckResourceAttrRegionalARN(resourceName, "arn", "kinesisanalytics", fmt.Sprintf("application/%s", rName)),
					resource.TestCheckResourceAttr(resourceName, "cloudwatch_logging_options.#", "0"),
					resource.TestCheckResourceAttrSet(resourceName, "create_timestamp"),
					resource.TestCheckResourceAttr(resourceName, "description", ""),
					resource.TestCheckResourceAttrSet(resourceName, "last_update_timestamp"),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "runtime_environment", "SQL-1_0"),
					resource.TestCheckResourceAttrPair(resourceName, "service_execution_role", iamRoleResourceName, "arn"),
					resource.TestCheckResourceAttr(resourceName, "status", "READY"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
					resource.TestCheckResourceAttr(resourceName, "version_id", "1"),
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

func TestAccAWSKinesisAnalyticsV2Application_disappears(t *testing.T) {
	var v kinesisanalyticsv2.ApplicationDetail
	resourceName := "aws_kinesisanalyticsv2_application.test"
	rName := acctest.RandomWithPrefix("tf-acc-test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSKinesisAnalyticsV2(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckKinesisAnalyticsV2ApplicationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccKinesisAnalyticsV2ApplicationConfigBasic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckKinesisAnalyticsV2ApplicationExists(resourceName, &v),
					testAccCheckResourceDisappears(testAccProvider, resourceAwsKinesisAnalyticsV2Application(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccAWSKinesisAnalyticsV2Application_UpdateServiceExecutionRole(t *testing.T) {
	var v kinesisanalyticsv2.ApplicationDetail
	resourceName := "aws_kinesisanalyticsv2_application.test"
	iamRoleResourceName1 := "aws_iam_role.test.0"
	iamRoleResourceName2 := "aws_iam_role.test.1"
	rName := acctest.RandomWithPrefix("tf-acc-test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSKinesisAnalyticsV2(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckKinesisAnalyticsV2ApplicationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccKinesisAnalyticsV2ApplicationConfigBasicPlusDescription(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckKinesisAnalyticsV2ApplicationExists(resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.#", "0"),
					testAccCheckResourceAttrRegionalARN(resourceName, "arn", "kinesisanalytics", fmt.Sprintf("application/%s", rName)),
					resource.TestCheckResourceAttr(resourceName, "cloudwatch_logging_options.#", "0"),
					resource.TestCheckResourceAttrSet(resourceName, "create_timestamp"),
					resource.TestCheckResourceAttr(resourceName, "description", "Testing"),
					resource.TestCheckResourceAttrSet(resourceName, "last_update_timestamp"),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "runtime_environment", "SQL-1_0"),
					resource.TestCheckResourceAttrPair(resourceName, "service_execution_role", iamRoleResourceName1, "arn"),
					resource.TestCheckResourceAttr(resourceName, "status", "READY"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
					resource.TestCheckResourceAttr(resourceName, "version_id", "1"),
				),
			},
			{
				Config: testAccKinesisAnalyticsV2ApplicationConfigBasicServiceExecutionRoleUpdated(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckKinesisAnalyticsV2ApplicationExists(resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.#", "0"),
					testAccCheckResourceAttrRegionalARN(resourceName, "arn", "kinesisanalytics", fmt.Sprintf("application/%s", rName)),
					resource.TestCheckResourceAttr(resourceName, "cloudwatch_logging_options.#", "0"),
					resource.TestCheckResourceAttrSet(resourceName, "create_timestamp"),
					resource.TestCheckResourceAttr(resourceName, "description", "Testing"),
					resource.TestCheckResourceAttrSet(resourceName, "last_update_timestamp"),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "runtime_environment", "SQL-1_0"),
					resource.TestCheckResourceAttrPair(resourceName, "service_execution_role", iamRoleResourceName2, "arn"),
					resource.TestCheckResourceAttr(resourceName, "status", "READY"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
					resource.TestCheckResourceAttr(resourceName, "version_id", "2"),
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

func TestAccAWSKinesisAnalyticsV2Application_simpleSql(t *testing.T) {
	var v kinesisanalyticsv2.ApplicationDetail
	resourceName := "aws_kinesisanalyticsv2_application.test"
	rName := acctest.RandomWithPrefix("tf-acc-test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSKinesisAnalyticsV2(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckKinesisAnalyticsV2ApplicationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccKinesisAnalyticsV2ApplicationConfigSimpleSql(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckKinesisAnalyticsV2ApplicationExists(resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "version_id", "1"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.application_code_configuration.0.code_content.0.text_content", "testCode\n"),
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

func TestAccAWSKinesisAnalyticsV2Application_UpdateApplicationCodeConfiguration(t *testing.T) {
	var v kinesisanalyticsv2.ApplicationDetail
	resourceName := "aws_kinesisanalyticsv2_application.test"
	rName := acctest.RandomWithPrefix("tf-acc-test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSKinesisAnalyticsV2(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckKinesisAnalyticsV2ApplicationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccKinesisAnalyticsV2ApplicationConfigSimpleSql(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckKinesisAnalyticsV2ApplicationExists(resourceName, &v),
				),
			},
			{
				Config: testAccKinesisAnalyticsV2ApplicationConfigApplicationCodeConfigurationUpdated(rName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "version_id", "2"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.application_code_configuration.0.code_content.0.text_content", "testCode2\n"),
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

func TestAccAWSKinesisAnalyticsV2Application_AddCloudWatchLoggingOptions(t *testing.T) {
	var v kinesisanalyticsv2.ApplicationDetail
	resourceName := "aws_kinesisanalyticsv2_application.test"
	rName := acctest.RandomWithPrefix("tf-acc-test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSKinesisAnalyticsV2(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckKinesisAnalyticsV2ApplicationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccKinesisAnalyticsV2ApplicationConfigSimpleSql(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckKinesisAnalyticsV2ApplicationExists(resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "version_id", "1"),
				),
			},
			{
				Config: testAccKinesisAnalyticsV2ApplicationConfigCloudWatchLoggingOptions(rName, 0),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckKinesisAnalyticsV2ApplicationExists(resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "version_id", "2"),
					resource.TestCheckResourceAttr(resourceName, "cloudwatch_logging_options.#", "1"),
					resource.TestCheckResourceAttrPair(resourceName, "cloudwatch_logging_options.0.log_stream_arn", "aws_cloudwatch_log_stream.test.0", "arn"),
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

func TestAccAWSKinesisAnalyticsV2Application_UpdateCloudWatchLoggingOptions(t *testing.T) {
	var v kinesisanalyticsv2.ApplicationDetail
	resourceName := "aws_kinesisanalyticsv2_application.test"
	rName := acctest.RandomWithPrefix("tf-acc-test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSKinesisAnalyticsV2(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckKinesisAnalyticsV2ApplicationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccKinesisAnalyticsV2ApplicationConfigCloudWatchLoggingOptions(rName, 0),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckKinesisAnalyticsV2ApplicationExists(resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "version_id", "1"),
					resource.TestCheckResourceAttr(resourceName, "cloudwatch_logging_options.#", "1"),
					resource.TestCheckResourceAttrPair(resourceName, "cloudwatch_logging_options.0.log_stream_arn", "aws_cloudwatch_log_stream.test.0", "arn"),
				),
			},
			{
				Config: testAccKinesisAnalyticsV2ApplicationConfigCloudWatchLoggingOptions(rName, 1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckKinesisAnalyticsV2ApplicationExists(resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "version_id", "2"),
					resource.TestCheckResourceAttr(resourceName, "cloudwatch_logging_options.#", "1"),
					resource.TestCheckResourceAttrPair(resourceName, "cloudwatch_logging_options.0.log_stream_arn", "aws_cloudwatch_log_stream.test.1", "arn"),
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

func TestAccAWSKinesisAnalyticsV2Application_KinesisFirehoseInput(t *testing.T) {
	var v kinesisanalyticsv2.ApplicationDetail
	resourceName := "aws_kinesisanalyticsv2_application.test"
	rName := acctest.RandomWithPrefix("tf-acc-test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSKinesisAnalyticsV2(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckKinesisAnalyticsV2ApplicationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccKinesisAnalyticsV2ApplicationConfigKinesisFirehoseInput(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckKinesisAnalyticsV2ApplicationExists(resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.input.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.input.0.kinesis_firehose.#", "1"),
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

func TestAccAWSKinesisAnalyticsV2Application_FlinkApplication(t *testing.T) {
	var v kinesisanalyticsv2.ApplicationDetail
	resourceName := "aws_kinesisanalyticsv2_application.test"
	rName := acctest.RandomWithPrefix("tf-acc-test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSKinesisAnalyticsV2(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckKinesisAnalyticsV2ApplicationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccKinesisAnalyticsV2ApplicationConfigFlinkApplication(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckKinesisAnalyticsV2ApplicationExists(resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "version_id", "1"),
					resource.TestCheckResourceAttr(resourceName, "cloudwatch_logging_options.#", "1"),
					resource.TestCheckResourceAttrPair(resourceName, "cloudwatch_logging_options.0.log_stream_arn", "aws_cloudwatch_log_stream.test", "arn"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.environment_properties.0.property_group.#", "1"),
					tfawsresource.TestCheckTypeSetElemNestedAttrs(resourceName, "application_configuration.0.environment_properties.0.property_group.*", map[string]string{
						"property_group_id": "abcdef",
						"property_map.key1": "val1",
						"property_map.key2": "val2",
					}),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.flink_application_configuration.0.checkpoint_configuration.#", "1"),
					tfawsresource.TestCheckTypeSetElemNestedAttrs(resourceName, "application_configuration.0.flink_application_configuration.0.checkpoint_configuration.*", map[string]string{
						"checkpointing_enabled":         "true",
						"checkpoint_interval":           "30000",
						"configuration_type":            "CUSTOM",
						"min_pause_between_checkpoints": "10000",
					}),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.flink_application_configuration.0.monitoring_configuration.#", "1"),
					tfawsresource.TestCheckTypeSetElemNestedAttrs(resourceName, "application_configuration.0.flink_application_configuration.0.monitoring_configuration.*", map[string]string{
						"configuration_type": "CUSTOM",
						"log_level":          "WARN",
						"metrics_level":      "APPLICATION",
					}),
					tfawsresource.TestCheckTypeSetElemNestedAttrs(resourceName, "application_configuration.0.flink_application_configuration.0.parallelism_configuration.*", map[string]string{
						"configuration_type":  "CUSTOM",
						"autoscaling_enabled": "true",
						"parallelism":         "1",
						"parallelism_per_kpu": "1",
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

func TestAccAWSKinesisAnalyticsV2Application_UpdateFlinkApplication(t *testing.T) {
	var v kinesisanalyticsv2.ApplicationDetail
	resourceName := "aws_kinesisanalyticsv2_application.test"
	rName := acctest.RandomWithPrefix("tf-acc-test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSKinesisAnalyticsV2(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckKinesisAnalyticsV2ApplicationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccKinesisAnalyticsV2ApplicationConfigFlinkApplication(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckKinesisAnalyticsV2ApplicationExists(resourceName, &v),
				),
			},
			{
				Config: testAccKinesisAnalyticsV2ApplicationConfigFlinkApplicationUpdated(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckKinesisAnalyticsV2ApplicationExists(resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "version_id", "2"),
					tfawsresource.TestCheckTypeSetElemNestedAttrs(resourceName, "application_configuration.0.flink_application_configuration.0.parallelism_configuration.*", map[string]string{
						"autoscaling_enabled": "false",
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

func TestAccAWSKinesisAnalyticsV2Application_KinesisStreamInput(t *testing.T) {
	var v kinesisanalyticsv2.ApplicationDetail
	resourceName := "aws_kinesisanalyticsv2_application.test"
	rName := acctest.RandomWithPrefix("tf-acc-test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSKinesisAnalytics(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckKinesisAnalyticsApplicationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccKinesisAnalyticsV2ApplicationConfigKinesisStreamInput(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckKinesisAnalyticsV2ApplicationExists(resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "version_id", "1"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.input.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.input.0.name_prefix", "test_prefix"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.input.0.kinesis_stream.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.input.0.parallelism.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.input.0.schema.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.input.0.schema.0.record_column.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.input.0.schema.0.record_format.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.input.0.schema.0.record_format.0.mapping_parameters.0.json.#", "1"),
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

func TestAccAWSKineissAnalyticsV2Application_UpdateKinesisStreamInput(t *testing.T) {
	var v kinesisanalyticsv2.ApplicationDetail
	resourceName := "aws_kinesisanalyticsv2_application.test"
	rName := acctest.RandomWithPrefix("tf-acc-test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSKinesisAnalytics(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckKinesisAnalyticsApplicationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccKinesisAnalyticsV2ApplicationConfigKinesisStreamInput(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckKinesisAnalyticsV2ApplicationExists(resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "version_id", "1"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.input.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.input.0.name_prefix", "test_prefix"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.input.0.kinesis_stream.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.input.0.schema.0.record_format.0.mapping_parameters.0.json.#", "1"),
				),
			},
			{
				Config: testAccKinesisAnalyticsV2ApplicationConfigKinesisStreamInputUpdated(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckKinesisAnalyticsV2ApplicationExists(resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "version_id", "2"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.input.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.input.0.name_prefix", "test_prefix2"),
					resource.TestCheckResourceAttrPair(resourceName, "application_configuration.0.sql_application_configuration.0.input.0.kinesis_stream.0.resource_arn", "aws_kinesis_stream.test", "arn"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.input.0.parallelism.0.count", "2"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.input.0.schema.0.record_columns.0.name", "test2"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.input.0.schema.0.record_format.0.mapping_parameters.0.csv.#", "1"),
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

func TestAccAWSKinesisAnalyticsV2Application_AddKinesisStreamInput(t *testing.T) {
	var v kinesisanalyticsv2.ApplicationDetail
	resourceName := "aws_kinesisanalyticsv2_application.test"
	rName := acctest.RandomWithPrefix("tf-acc-test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSKinesisAnalyticsV2(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckKinesisAnalyticsApplicationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccKinesisAnalyticsV2ApplicationConfigSimpleSql(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckKinesisAnalyticsV2ApplicationExists(resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "version_id", "1"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.input.#", "0"),
				),
			},
			{
				Config: testAccKinesisAnalyticsV2ApplicationConfigKinesisStreamInput(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckKinesisAnalyticsV2ApplicationExists(resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "version_id", "2"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.input.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.input.0.name_prefix", "test_prefix"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.input.0.kinesis_stream.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.input.0.parallelism.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.input.0.schema.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.input.0.schema.0.record_column.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.input.0.schema.0.record_format.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.input.0.schema.0.record_format.0.mapping_parameters.0.json.#", "1"),
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

func TestAccAWSKinesisAnalyticsV2Application_KinesisStreamOutput(t *testing.T) {
	var v kinesisanalyticsv2.ApplicationDetail
	resourceName := "aws_kinesisanalyticsv2_application.test"
	rName := acctest.RandomWithPrefix("tf-acc-test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSKinesisAnalyticsV2(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckKinesisAnalyticsV2ApplicationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccKinesisAnalyticsV2ApplicationConfigKinesisStreamOutput(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckKinesisAnalyticsV2ApplicationExists(resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "version_id", "1"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.output.#", "1"),
					tfawsresource.TestCheckTypeSetElemNestedAttrs(resourceName, "application_configuration.0.sql_application_configuration.0.output.*", map[string]string{
						"name":                        "test_name",
						"kinesis_stream.#":            "1",
						"schema.#":                    "1",
						"schema.0.record_format_type": "JSON",
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

func TestAccAWSKinesisAnalyticsV2Application_UpdateKinesisStreamOutput(t *testing.T) {
	var v kinesisanalyticsv2.ApplicationDetail
	resourceName := "aws_kinesisanalyticsv2_application.test"
	rName := acctest.RandomWithPrefix("tf-acc-test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSKinesisAnalyticsV2(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckKinesisAnalyticsApplicationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccKinesisAnalyticsV2ApplicationConfigKinesisStreamOutput(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckKinesisAnalyticsV2ApplicationExists(resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "version_id", "1"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.output.#", "1"),
					tfawsresource.TestCheckTypeSetElemNestedAttrs(resourceName, "application_configuration.0.sql_application_configuration.0.output.*", map[string]string{
						"name":                        "test_name",
						"kinesis_stream.#":            "1",
						"schema.#":                    "1",
						"schema.0.record_format_type": "JSON",
					}),
				),
			},
			{
				Config: testAccKinesisAnalyticsV2ApplicationConfigKinesisStreamOutputUpdated(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckKinesisAnalyticsV2ApplicationExists(resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "version_id", "2"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.output.#", "1"),
					tfawsresource.TestCheckTypeSetElemNestedAttrs(resourceName, "application_configuration.0.sql_application_configuration.0.output.*", map[string]string{
						"name":                        "test_name2",
						"kinesis_stream.#":            "1",
						"schema.#":                    "1",
						"schema.0.record_format_type": "CSV",
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

func TestAccAWSKinesisAnalyticsV2Application_MultipleKinesisStreamOutputs(t *testing.T) {
	var v kinesisanalyticsv2.ApplicationDetail
	resourceName := "aws_kinesisanalyticsv2_application.test"
	rName := acctest.RandomWithPrefix("tf-acc-test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSKinesisAnalyticsV2(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckKinesisAnalyticsApplicationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccKinesisAnalyticsV2ApplicationConfigMultipleOutputs(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckKinesisAnalyticsV2ApplicationExists(resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.output.#", "2"),
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

func TestAccAWSKinesisAnalyticsV2Application_AddKinesisStreamOutput(t *testing.T) {
	var v kinesisanalyticsv2.ApplicationDetail
	resourceName := "aws_kinesisanalyticsv2_application.test"
	rName := acctest.RandomWithPrefix("tf-acc-test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSKinesisAnalyticsV2(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckKinesisAnalyticsV2ApplicationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccKinesisAnalyticsV2ApplicationConfigSimpleSql(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckKinesisAnalyticsV2ApplicationExists(resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "version_id", "1"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.output.#", "0"),
				),
			},
			{
				Config: testAccKinesisAnalyticsV2ApplicationConfigKinesisStreamOutput(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckKinesisAnalyticsV2ApplicationExists(resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "version_id", "2"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.output.#", "1"),
					tfawsresource.TestCheckTypeSetElemNestedAttrs(resourceName, "application_configuration.0.sql_application_configuration.0.output.*", map[string]string{
						"name":             "test_name",
						"kinesis_stream.#": "1",
						"schema.#":         "1",
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

func TestAccAWSKinesisAnalyticsV2Application_LambdaOutput(t *testing.T) {
	var v kinesisanalyticsv2.ApplicationDetail
	resourceName := "aws_kinesisanalyticsv2_application.test"
	rName := acctest.RandomWithPrefix("tf-acc-test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSKinesisAnalyticsV2(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckKinesisAnalyticsV2ApplicationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccKinesisAnalyticsV2ApplicationConfigLambdaOutput(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckKinesisAnalyticsV2ApplicationExists(resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "version_id", "1"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.output.#", "1"),
					tfawsresource.TestCheckTypeSetElemNestedAttrs(resourceName, "application_configuration.0.sql_application_configuration.0.output.*", map[string]string{
						"lambda.#": "1",
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

func TestAccAWSKinesisAnalyticsV2Application_AddLambdaOutput(t *testing.T) {
	var v kinesisanalyticsv2.ApplicationDetail
	resourceName := "aws_kinesisanalyticsv2_application.test"
	rName := acctest.RandomWithPrefix("tf-acc-test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSKinesisAnalyticsV2(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckKinesisAnalyticsV2ApplicationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccKinesisAnalyticsV2ApplicationConfigSimpleSql(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckKinesisAnalyticsV2ApplicationExists(resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "version_id", "1"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.output.#", "0"),
				),
			},
			{
				Config: testAccKinesisAnalyticsV2ApplicationConfigLambdaOutput(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckKinesisAnalyticsV2ApplicationExists(resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "version_id", "2"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.output.#", "1"),
					tfawsresource.TestCheckTypeSetElemNestedAttrs(resourceName, "application_configuration.0.sql_application_configuration.0.output.*", map[string]string{
						"lambda.#": "1",
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

func TestAccAWSKinesisAnalyticsV2Application_ReferenceDataSource(t *testing.T) {
	var v kinesisanalyticsv2.ApplicationDetail
	resourceName := "aws_kinesisanalyticsv2_application.test"
	rName := acctest.RandomWithPrefix("tf-acc-test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSKinesisAnalyticsV2(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckKinesisAnalyticsApplicationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccKinesisAnalyticsV2ApplicationConfigReferenceDataSource(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckKinesisAnalyticsV2ApplicationExists(resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "version_id", "1"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.reference_data_sources.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.reference_data_sources.0.schema.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.reference_data_sources.0.schema.0.record_column.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.reference_data_sources.0.schema.0.record_format.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.reference_data_sources.0.schema.0.record_format.0.mapping_parameters.0.json.#", "1"),
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

func TestAccAWSKinesisAnalyticsV2Application_UpdateReferenceDataSource(t *testing.T) {
	var v kinesisanalyticsv2.ApplicationDetail
	resourceName := "aws_kinesisanalyticsv2_application.test"
	rName := acctest.RandomWithPrefix("tf-acc-test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSKinesisAnalyticsV2(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckKinesisAnalyticsV2ApplicationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccKinesisAnalyticsV2ApplicationConfigReferenceDataSource(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckKinesisAnalyticsV2ApplicationExists(resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "version_id", "1"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.reference_data_sources.#", "1"),
				),
			},
			{
				Config: testAccKinesisAnalyticsV2ApplicationConfigReferenceDataSourceUpdated(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckKinesisAnalyticsV2ApplicationExists(resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "version_id", "2"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.reference_data_sources.#", "1"),
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

func TestAccAWSKinesisAnalyticsV2Application_Tags(t *testing.T) {
	var v kinesisanalyticsv2.ApplicationDetail
	resourceName := "aws_kinesisanalyticsv2_application.test"
	rName := acctest.RandomWithPrefix("tf-acc-test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSKinesisAnalyticsV2(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckKinesisAnalyticsV2ApplicationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccKinesisAnalyticsV2ApplicationConfigTags1(rName, "key1", "value1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckKinesisAnalyticsV2ApplicationExists(resourceName, &v),
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
				Config: testAccKinesisAnalyticsV2ApplicationConfigTags2(rName, "key1", "value1updated", "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckKinesisAnalyticsV2ApplicationExists(resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1updated"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
			{
				Config: testAccKinesisAnalyticsV2ApplicationConfigTags1(rName, "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckKinesisAnalyticsV2ApplicationExists(resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
		},
	})
}

func testAccCheckKinesisAnalyticsV2ApplicationDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*AWSClient).kinesisanalyticsv2conn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_kinesisanalyticsv2_application" {
			continue
		}

		_, err := finder.ApplicationByName(conn, rs.Primary.Attributes["name"])
		if isAWSErr(err, kinesisanalyticsv2.ErrCodeResourceNotFoundException, "") {
			continue
		}
		if err != nil {
			return err
		}

		return fmt.Errorf("Kinesis Analytics v2 Application %s still exists", rs.Primary.ID)
	}
	return nil
}

func testAccCheckKinesisAnalyticsV2ApplicationExists(n string, v *kinesisanalyticsv2.ApplicationDetail) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No Kinesis Analytics v2 Application ID is set")
		}

		conn := testAccProvider.Meta().(*AWSClient).kinesisanalyticsv2conn

		application, err := finder.ApplicationByName(conn, rs.Primary.Attributes["name"])
		if err != nil {
			return err
		}

		*v = *application

		return nil
	}
}

func testAccPreCheckAWSKinesisAnalyticsV2(t *testing.T) {
	conn := testAccProvider.Meta().(*AWSClient).kinesisanalyticsv2conn

	input := &kinesisanalyticsv2.ListApplicationsInput{}

	_, err := conn.ListApplications(input)

	if testAccPreCheckSkipError(err) {
		t.Skipf("skipping acceptance testing: %s", err)
	}

	if err != nil {
		t.Fatalf("unexpected PreCheck error: %s", err)
	}
}

func testAccKinesisAnalyticsV2ApplicationConfigBaseServiceExecutionIamRole(rName string, nRoles int) string {
	return fmt.Sprintf(`
resource "aws_iam_role" "test" {
  count = %[2]d

  name               = "%[1]s.${count.index}"
  assume_role_policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Action": ["sts:AssumeRole"],
      "Principal": {"Service": "firehose.amazonaws.com"}
    },
    {
      "Effect": "Allow",
      "Action": ["sts:AssumeRole"],
      "Principal": {"Service": "kinesisanalytics.amazonaws.com"}
    },
    {
      "Effect": "Allow",
      "Action": ["sts:AssumeRole"],
      "Principal": {"Service": "lambda.amazonaws.com"}
    }
  ]
}
EOF
}

resource "aws_iam_policy" "test" {
  name   = %[1]q
  policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Action": ["firehose:*"],
      "Resource": ["*"]
    },
    {
      "Effect": "Allow",
      "Action": ["lambda:*"],
      "Resource": ["*"]
    },
    {
      "Effect": "Allow",
      "Action": ["s3:*"],
      "Resource": ["*"]
    }
  ]
}
EOF
}

resource "aws_iam_role_policy_attachment" "test" {
  count = %[2]d

  role       = aws_iam_role.test[count.index].name
  policy_arn = aws_iam_policy.test.arn
}
`, rName, nRoles)
}

func testAccKinesisAnalyticsV2ApplicationConfigBaseFlinkApplication(rName string) string {
	return fmt.Sprintf(`
resource "aws_s3_bucket" "test" {
  bucket = %[1]q
}

resource "aws_s3_bucket_object" "test" {
  bucket = aws_s3_bucket.test.bucket
  key    = %[1]q
  source = "test-fixtures/flink-app.jar"
}

resource "aws_cloudwatch_log_group" "test" {
  name = %[1]q
}

resource "aws_cloudwatch_log_stream" "test" {
  name           = %[1]q
  log_group_name = aws_cloudwatch_log_group.test.name
}
`, rName)
}

func testAccKinesisAnalyticsV2ApplicationConfigBasic(rName string) string {
	return composeConfig(
		testAccKinesisAnalyticsV2ApplicationConfigBaseServiceExecutionIamRole(rName, 1),
		fmt.Sprintf(`
resource "aws_kinesisanalyticsv2_application" "test" {
  name                   = %[1]q
  runtime_environment    = "SQL-1_0"
  service_execution_role = aws_iam_role.test.0.arn
}
`, rName))
}

func testAccKinesisAnalyticsV2ApplicationConfigBasicPlusDescription(rName string) string {
	return composeConfig(
		testAccKinesisAnalyticsV2ApplicationConfigBaseServiceExecutionIamRole(rName, 2),
		fmt.Sprintf(`
resource "aws_kinesisanalyticsv2_application" "test" {
  name                   = %[1]q
  runtime_environment    = "SQL-1_0"
  service_execution_role = aws_iam_role.test.0.arn
  description            = "Testing"
}
`, rName))
}

func testAccKinesisAnalyticsV2ApplicationConfigBasicServiceExecutionRoleUpdated(rName string) string {
	return composeConfig(
		testAccKinesisAnalyticsV2ApplicationConfigBaseServiceExecutionIamRole(rName, 2),
		fmt.Sprintf(`
resource "aws_kinesisanalyticsv2_application" "test" {
  name                   = %[1]q
  runtime_environment    = "SQL-1_0"
  service_execution_role = aws_iam_role.test.1.arn
  description            = "Testing"
}
`, rName))
}

func testAccKinesisAnalyticsV2ApplicationConfigSimpleSql(rName string) string {
	return composeConfig(
		testAccKinesisAnalyticsV2ApplicationConfigBaseServiceExecutionIamRole(rName, 1),
		fmt.Sprintf(`
resource "aws_kinesisanalyticsv2_application" "test" {
  name                   = %[1]q
  runtime_environment    = "SQL-1_0"
  service_execution_role = aws_iam_role.test.0.arn

  application_configuration {
    application_code_configuration {
      code_content {
        text_content = "testCode\n"
      }

      code_content_type = "PLAINTEXT"
    }

    sql_application_configuration {}
  }
}
`, rName))
}

func testAccKinesisAnalyticsV2ApplicationConfigApplicationCodeConfigurationUpdated(rName string) string {
	return composeConfig(
		testAccKinesisAnalyticsV2ApplicationConfigBaseServiceExecutionIamRole(rName, 1),
		fmt.Sprintf(`
resource "aws_kinesisanalyticsv2_application" "test" {
  name                   = %[1]q
  runtime_environment    = "SQL-1_0"
  service_execution_role = aws_iam_role.test.0.arn

  application_configuration {
    application_code_configuration {
      code_content {
        text_content = "testCode2\n"
      }

      code_content_type = "PLAINTEXT"
    }

    sql_application_configuration {}
  }
}
`, rName))
}

func testAccKinesisAnalyticsV2ApplicationConfigCloudWatchLoggingOptions(rName string, streamIndex int) string {
	return composeConfig(
		testAccKinesisAnalyticsV2ApplicationConfigBaseServiceExecutionIamRole(rName, 1),
		fmt.Sprintf(`
resource "aws_cloudwatch_log_group" "test" {
  name = %[1]q
}

resource "aws_cloudwatch_log_stream" "test" {
  count = 2

  name           = "%[1]s.${count.index}"
  log_group_name = aws_cloudwatch_log_group.test.name
}

resource "aws_kinesisanalyticsv2_application" "test" {
  name                   = %[1]q
  runtime_environment    = "SQL-1_0"
  service_execution_role = aws_iam_role.test.0.arn

  application_configuration {
    application_code_configuration {
      code_content {
        text_content = "testCode\n"
      }

      code_content_type = "PLAINTEXT"
    }

    sql_application_configuration {}
  }

  cloudwatch_logging_options {
    log_stream_arn = aws_cloudwatch_log_stream.test.%[2]d.arn
  }
}
`, rName, streamIndex))
}

func testAccKinesisAnalyticsV2ApplicationConfigKinesisFirehoseInput(rName string) string {
	return composeConfig(
		testAccKinesisAnalyticsV2ApplicationConfigBaseServiceExecutionIamRole(rName, 1),
		fmt.Sprintf(`
resource "aws_s3_bucket" "test" {
  bucket = %[1]q
}

resource "aws_lambda_function" "test" {
  filename      = "test-fixtures/lambdatest.zip"
  function_name = %[1]q
  handler       = "exports.example"
  role          = aws_iam_role.test.0.arn
  runtime       = "nodejs12.x"
}

resource "aws_kinesis_firehose_delivery_stream" "test" {
  name        = %[1]q
  destination = "extended_s3"

  extended_s3_configuration {
    bucket_arn = aws_s3_bucket.test.arn
    role_arn   = aws_iam_role.test.0.arn
  }
}

resource "aws_kinesisanalyticsv2_application" "test" {
  name                   = %[1]q
  runtime_environment    = "SQL-1_0"
  service_execution_role = aws_iam_role.test.0.arn

  application_configuration {
    application_code_configuration {
      code_content {
        text_content = "testCode\n"
      }

      code_content_type = "PLAINTEXT"
    }

    sql_application_configuration {
      input {
        name_prefix = "test_prefix"

        kinesis_firehose {
          resource_arn = aws_kinesis_firehose_delivery_stream.test.arn
        }

        parallelism {
          count = 1
        }

        schema {
          record_column {
            mapping  = "$.test"
            name     = "test"
            sql_type = "VARCHAR(8)"
          }

          record_encoding = "UTF-8"

          record_format {
            mapping_parameters {
              csv {
                record_column_delimiter = ","
                record_row_delimiter    = "\n"
              }
            }
          }
        }
      }
    }
  }
}
`, rName))
}

func testAccKinesisAnalyticsV2ApplicationConfigFlinkApplication(rName string) string {
	return composeConfig(
		testAccKinesisAnalyticsV2ApplicationConfigBaseServiceExecutionIamRole(rName, 1),
		testAccKinesisAnalyticsV2ApplicationConfigBaseFlinkApplication(rName),
		fmt.Sprintf(`
resource "aws_kinesisanalyticsv2_application" "test" {
  name                   = %[1]q
  runtime_environment    = "FLINK-1_8"
  service_execution_role = aws_iam_role.test.0.arn

  application_configuration {
    application_snapshot_configuration {
      snapshots_enabled = true
    }

    application_code_configuration {
      code_content {
        s3_content_location {
          bucket_arn = aws_s3_bucket.test.arn
          file_key   = aws_s3_bucket_object.test.key
        }
      }

      code_content_type = "ZIPFILE"
    }

    flink_application_configuration {
      checkpoint_configuration {
        checkpointing_enabled         = true
        checkpoint_interval           = 30000
        configuration_type            = "CUSTOM"
        min_pause_between_checkpoints = 10000
      }

      parallelism_configuration {
        configuration_type  = "CUSTOM"
        autoscaling_enabled = true
        parallelism         = 1
        parallelism_per_kpu = 1
      }

      monitoring_configuration {
        configuration_type = "CUSTOM"
        log_level          = "WARN"
        metrics_level      = "APPLICATION"
      }
    }

    environment_properties {
      property_group {
        property_group_id  = "abcdef"

        property_map = {
          key1 = "val1"
          key2 = "val2"
        }
      }
    }
  }

  cloudwatch_logging_options {
    log_stream_arn = aws_cloudwatch_log_stream.test.arn
  }
}
`, rName))
}

func testAccKinesisAnalyticsV2ApplicationConfigFlinkApplicationUpdated(rName string) string {
	return composeConfig(
		testAccKinesisAnalyticsV2ApplicationConfigBaseServiceExecutionIamRole(rName, 1),
		testAccKinesisAnalyticsV2ApplicationConfigBaseFlinkApplication(rName),
		fmt.Sprintf(`
resource "aws_kinesisanalyticsv2_application" "test" {
  name                   = %[1]q
  runtime_environment    = "FLINK-1_8"
  service_execution_role = aws_iam_role.test.0.arn

  application_configuration {
    application_snapshot_configuration {
      snapshots_enabled = true
    }

    application_code_configuration {
      code_content {
        s3_content_location {
          bucket_arn = aws_s3_bucket.test.arn
          file_key   = aws_s3_bucket_object.test.key
        }
      }

      code_content_type = "ZIPFILE"
    }

    flink_application_configuration {
      checkpoint_configuration {
        checkpointing_enabled         = true
        checkpoint_interval           = 30000
        configuration_type            = "CUSTOM"
        min_pause_between_checkpoints = 10000
      }

      parallelism_configuration {
        configuration_type  = "CUSTOM"
        autoscaling_enabled = false
        parallelism         = 1
        parallelism_per_kpu = 1
      }

      monitoring_configuration {
        configuration_type = "CUSTOM"
        log_level          = "WARN"
        metrics_level      = "APPLICATION"
      }
    }

    environment_properties {
      property_group {
        property_group_id  = "abcdef"

        property_map = {
          key1 = "val1"
          key2 = "val2"
        }
      }
    }
  }

  cloudwatch_logging_options {
    log_stream_arn = aws_cloudwatch_log_stream.test.arn
  }
}
`, rName))
}

func testAccKinesisAnalyticsV2ApplicationConfigKinesisStreamInput(rName string) string {
	return composeConfig(
		testAccKinesisAnalyticsV2ApplicationConfigBaseServiceExecutionIamRole(rName, 1),
		fmt.Sprintf(`
resource "aws_kinesis_stream" "test" {
  count = 2

  name        = "%[1]s.${count.index}"
  shard_count = 1
}

resource "aws_kinesisanalyticsv2_application" "test" {
  name                   = %[1]q
  runtime_environment    = "SQL-1_0"
  service_execution_role = aws_iam_role.test.0.arn

  application_configuration {
    application_code_configuration {
      code_content {
        text_content = "testCode\n"
      }

      code_content_type = "PLAINTEXT"
    }

    sql_application_configuration {
      input {
        name_prefix = "test_prefix"

        kinesis_stream {
          resource_arn = aws_kinesis_stream.test.0.arn
        }

        parallelism {
          count = 1
        }

        schema {
          record_column {
            mapping  = "$.test"
            name     = "test"
            sql_type = "VARCHAR(8)"
          }

          record_encoding = "UTF-8"

          record_format {
            mapping_parameters {
              json {
                record_row_path = "$"
              }
            }
          }
        }
      }
    }
  }
}
`, rName))
}

func testAccKinesisAnalyticsV2ApplicationConfigKinesisStreamInputUpdated(rName string) string {
	return composeConfig(
		testAccKinesisAnalyticsV2ApplicationConfigBaseServiceExecutionIamRole(rName, 1),
		fmt.Sprintf(`
resource "aws_kinesis_stream" "test" {
  count = 2

  name        = "%[1]s.${count.index}"
  shard_count = 1
}

resource "aws_kinesisanalyticsv2_application" "test" {
  name                   = %[1]q
  runtime_environment    = "SQL-1_0"
  service_execution_role = aws_iam_role.test.0.arn

  application_configuration {
    application_code_configuration {
      code_content {
        text_content = "testCode\n"
      }

      code_content_type = "PLAINTEXT"
    }

    sql_application_configuration {
      input {
        name_prefix = "test_prefix2"

        kinesis_stream {
          resource_arn = aws_kinesis_stream.test.1.arn
        }

        parallelism {
          count = 2
        }

        schema {
          record_column {
            mapping  = "$.test2"
            name     = "test2"
            sql_type = "VARCHAR(8)"
          }

          record_encoding = "UTF-8"

          record_format {
            mapping_parameters {
              csv {
                record_column_delimiter = ","
                record_row_delimiter    = "\n"
              }
            }
          }
        }
      }
    }
  }
}
`, rName))
}

func testAccKinesisAnalyticsV2ApplicationConfigKinesisStreamOutput(rName string) string {
	return composeConfig(
		testAccKinesisAnalyticsV2ApplicationConfigBaseServiceExecutionIamRole(rName, 1),
		fmt.Sprintf(`
resource "aws_kinesis_stream" "test" {
  count = 2

  name        = "%[1]s.${count.index}"
  shard_count = 1
}

resource "aws_kinesisanalyticsv2_application" "test" {
  name                   = %[1]q
  runtime_environment    = "SQL-1_0"
  service_execution_role = aws_iam_role.test.0.arn

  application_configuration {
    application_code_configuration {
      code_content {
        text_content = "testCode\n"
      }

      code_content_type = "PLAINTEXT"
    }

    sql_application_configuration {
      output {
        name = "test_name"

        kinesis_stream {
          resource_arn = aws_kinesis_stream.test.0.arn
        }

        schema {
          record_format_type = "JSON"
        }
      }
    }
  }
}
`, rName))
}

func testAccKinesisAnalyticsV2ApplicationConfigKinesisStreamOutputUpdated(rName string) string {
	return composeConfig(
		testAccKinesisAnalyticsV2ApplicationConfigBaseServiceExecutionIamRole(rName, 1),
		fmt.Sprintf(`
resource "aws_kinesis_stream" "test" {
  count = 2

  name        = "%[1]s.${count.index}"
  shard_count = 1
}

resource "aws_kinesisanalyticsv2_application" "test" {
  name                   = %[1]q
  runtime_environment    = "SQL-1_0"
  service_execution_role = aws_iam_role.test.0.arn

  application_configuration {
    application_code_configuration {
      code_content {
        text_content = "testCode\n"
      }

      code_content_type = "PLAINTEXT"
    }

    sql_application_configuration {
      output {
        name = "test_name2"

        kinesis_stream {
          resource_arn = aws_kinesis_stream.test.1.arn
        }

        schema {
          record_format_type = "CSV"
        }
      }
    }
  }
}
`, rName))
}

func testAccKinesisAnalyticsV2ApplicationConfigMultipleOutputs(rName string) string {
	return composeConfig(
		testAccKinesisAnalyticsV2ApplicationConfigBaseServiceExecutionIamRole(rName, 1),
		fmt.Sprintf(`
resource "aws_kinesis_stream" "test" {
  count = 2

  name        = "%[1]s.${count.index}"
  shard_count = 1
}

resource "aws_kinesisanalyticsv2_application" "test" {
  name                   = %[1]q
  runtime_environment    = "SQL-1_0"
  service_execution_role = aws_iam_role.test.0.arn

  application_configuration {
    application_code_configuration {
      code_content {
        text_content = "testCode\n"
      }

      code_content_type = "PLAINTEXT"
    }

    sql_application_configuration {
      output {
        name = "test_name1"

        kinesis_stream {
          resource_arn = aws_kinesis_stream.test.0.arn
        }

        schema {
          record_format_type = "JSON"
        }
      }

      output {
        name = "test_name2"

        kinesis_stream {
          resource_arn = aws_kinesis_stream.test.1.arn
        }

        schema {
          record_format_type = "JSON"
        }
      }
    }
  }
}
`, rName))
}

func testAccKinesisAnalyticsV2ApplicationConfigLambdaOutput(rName string) string {
	return composeConfig(
		testAccKinesisAnalyticsV2ApplicationConfigBaseServiceExecutionIamRole(rName, 1),
		fmt.Sprintf(`
resource "aws_lambda_function" "test" {
  filename      = "test-fixtures/lambdatest.zip"
  function_name = %[1]q
  handler       = "exports.example"
  role          = aws_iam_role.test.0.arn
  runtime       = "nodejs12.x"
}

resource "aws_kinesisanalyticsv2_application" "test" {
  name                   = %[1]q
  runtime_environment    = "SQL-1_0"
  service_execution_role = aws_iam_role.test.0.arn

  application_configuration {
    application_code_configuration {
      code_content {
        text_content = "testCode\n"
      }

      code_content_type = "PLAINTEXT"
    }

    sql_application_configuration {
      output {
        name = "test_name"

        lambda {
          resource_arn = aws_lambda_function.test.arn
        }

        schema {
          record_format_type = "JSON"
        }
      }
    }
  }
}
`, rName))
}

func testAccKinesisAnalyticsV2ApplicationConfigReferenceDataSource(rName string) string {
	return composeConfig(
		testAccKinesisAnalyticsV2ApplicationConfigBaseServiceExecutionIamRole(rName, 1),
		fmt.Sprintf(`
resource "aws_s3_bucket" "test" {
  count = 2

  bucket = "%[1]s.${count.index}"
}

resource "aws_kinesisanalyticsv2_application" "test" {
  name                   = %[1]q
  runtime_environment    = "SQL-1_0"
  service_execution_role = aws_iam_role.test.0.arn

  application_configuration {
    application_code_configuration {
      code_content {
        text_content = "testCode\n"
      }

      code_content_type = "PLAINTEXT"
    }

    sql_application_configuration {
      reference_data_sources {
        table_name = "test_table"

        s3 {
          bucket_arn = aws_s3_bucket.test.0.arn
          file_key   = %[1]q
        }

        schema {
          record_column {
            mapping  = "$.test"
            name     = "test"
            sql_type = "VARCHAR(8)"
          }

          record_encoding = "UTF-8"

          record_format {
            mapping_parameters {
              json {
                record_row_path = "$"
              }
            }
          }
        }
      }
    }
  }
}
`, rName))
}

func testAccKinesisAnalyticsV2ApplicationConfigReferenceDataSourceUpdated(rName string) string {
	return composeConfig(
		testAccKinesisAnalyticsV2ApplicationConfigBaseServiceExecutionIamRole(rName, 1),
		fmt.Sprintf(`
resource "aws_s3_bucket" "test" {
  count = 2

  bucket = "%[1]s.${count.index}"
}

resource "aws_kinesisanalyticsv2_application" "test" {
  name                   = %[1]q
  runtime_environment    = "SQL-1_0"
  service_execution_role = aws_iam_role.test.0.arn

  application_configuration {
    application_code_configuration {
      code_content {
        text_content = "testCode\n"
      }

      code_content_type = "PLAINTEXT"
    }

    sql_application_configuration {
      reference_data_sources {
        table_name = "test_table2"

        s3 {
          bucket_arn =  aws_s3_bucket.test.1.arn
          file_key   = %[1]q
        }

        schema {
          record_column {
            mapping  = "$.test2"
            name     = "test2"
            sql_type = "VARCHAR(8)"
          }

          record_encoding = "UTF-8"

          record_format {
            mapping_parameters {
              csv {
                record_column_delimiter = ","
                record_row_delimiter    = "\n"
              }
            }
          }
        }
      }
    }
  }
}
`, rName))
}

func testAccKinesisAnalyticsV2ApplicationConfigTags1(rName, tagKey1, tagValue1 string) string {
	return composeConfig(
		testAccKinesisAnalyticsV2ApplicationConfigBaseServiceExecutionIamRole(rName, 1),
		fmt.Sprintf(`
resource "aws_kinesisanalyticsv2_application" "test" {
  name                   = %[1]q
  runtime_environment    = "SQL-1_0"
  service_execution_role = aws_iam_role.test.0.arn

  tags = {
    %[2]q = %[3]q
  }
}
`, rName, tagKey1, tagValue1))
}

func testAccKinesisAnalyticsV2ApplicationConfigTags2(rName, tagKey1, tagValue1, tagKey2, tagValue2 string) string {
	return composeConfig(
		testAccKinesisAnalyticsV2ApplicationConfigBaseServiceExecutionIamRole(rName, 1),
		fmt.Sprintf(`
resource "aws_kinesisanalyticsv2_application" "test" {
  name                   = %[1]q
  runtime_environment    = "SQL-1_0"
  service_execution_role = aws_iam_role.test.0.arn

  tags = {
    %[2]q = %[3]q
    %[4]q = %[5]q
  }
}
`, rName, tagKey1, tagValue1, tagKey2, tagValue2))
}
