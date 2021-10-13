package aws

import (
	"fmt"
	"log"
	"regexp"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/service/ec2/finder"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/tfresource"
)

func init() {
	resource.AddTestSweepers("aws_flow_log", &resource.Sweeper{
		Name: "aws_flow_log",
		F:    testSweepFlowLogs,
	})
}

func testSweepFlowLogs(region string) error {
	client, err := sharedClientForRegion(region)
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}
	conn := client.(*AWSClient).ec2conn
	input := &ec2.DescribeFlowLogsInput{}
	sweepResources := make([]*testSweepResource, 0)

	err = conn.DescribeFlowLogsPages(input, func(page *ec2.DescribeFlowLogsOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, flowLog := range page.FlowLogs {
			r := resourceAwsFlowLog()
			d := r.Data(nil)
			d.SetId(aws.StringValue(flowLog.FlowLogId))

			sweepResources = append(sweepResources, NewTestSweepResource(r, d, client))
		}

		return !lastPage
	})

	if testSweepSkipSweepError(err) {
		log.Printf("[WARN] Skipping Flow Log sweep for %s: %s", region, err)
		return nil
	}

	if err != nil {
		return fmt.Errorf("error listing Flow Logs (%s): %w", region, err)
	}

	err = testSweepResourceOrchestrator(sweepResources)

	if err != nil {
		return fmt.Errorf("error sweeping Flow Logs (%s): %w", region, err)
	}

	return nil
}

func TestAccAWSFlowLog_VPCID(t *testing.T) {
	var flowLog ec2.FlowLog
	cloudwatchLogGroupResourceName := "aws_cloudwatch_log_group.test"
	iamRoleResourceName := "aws_iam_role.test"
	resourceName := "aws_flow_log.test"
	vpcResourceName := "aws_vpc.test"
	rName := acctest.RandomWithPrefix("tf-acc-test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		ErrorCheck:   testAccErrorCheck(t, ec2.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckFlowLogDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccFlowLogConfig_VPCID(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFlowLogExists(resourceName, &flowLog),
					testAccMatchResourceAttrRegionalARN(resourceName, "arn", "ec2", regexp.MustCompile(`vpc-flow-log/fl-.+`)),
					resource.TestCheckResourceAttrPair(resourceName, "iam_role_arn", iamRoleResourceName, "arn"),
					resource.TestCheckResourceAttr(resourceName, "log_destination", ""),
					resource.TestCheckResourceAttr(resourceName, "log_destination_type", "cloud-watch-logs"),
					resource.TestCheckResourceAttrPair(resourceName, "log_group_name", cloudwatchLogGroupResourceName, "name"),
					resource.TestCheckResourceAttr(resourceName, "max_aggregation_interval", "600"),
					resource.TestCheckResourceAttr(resourceName, "traffic_type", "ALL"),
					resource.TestCheckResourceAttrPair(resourceName, "vpc_id", vpcResourceName, "id"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config:             testAccFlowLogConfig_LogDestinationType_CloudWatchLogs(rName),
				ExpectNonEmptyPlan: false,
			},
		},
	})
}

func TestAccAWSFlowLog_LogFormat(t *testing.T) {
	var flowLog ec2.FlowLog
	resourceName := "aws_flow_log.test"
	rName := acctest.RandomWithPrefix("tf-acc-test")
	logFormat := "${version} ${vpc-id} ${subnet-id}"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		ErrorCheck:   testAccErrorCheck(t, ec2.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckFlowLogDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccFlowLogConfig_LogFormat(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFlowLogExists(resourceName, &flowLog),
					resource.TestCheckResourceAttr(resourceName, "log_format", logFormat),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config:             testAccFlowLogConfig_LogDestinationType_CloudWatchLogs(rName),
				ExpectNonEmptyPlan: false,
			},
		},
	})
}

func TestAccAWSFlowLog_SubnetID(t *testing.T) {
	var flowLog ec2.FlowLog
	cloudwatchLogGroupResourceName := "aws_cloudwatch_log_group.test"
	iamRoleResourceName := "aws_iam_role.test"
	resourceName := "aws_flow_log.test"
	subnetResourceName := "aws_subnet.test"
	rName := acctest.RandomWithPrefix("tf-acc-test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		ErrorCheck:   testAccErrorCheck(t, ec2.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckFlowLogDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccFlowLogConfig_SubnetID(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFlowLogExists(resourceName, &flowLog),
					resource.TestCheckResourceAttrPair(resourceName, "iam_role_arn", iamRoleResourceName, "arn"),
					resource.TestCheckResourceAttr(resourceName, "log_destination", ""),
					resource.TestCheckResourceAttr(resourceName, "log_destination_type", "cloud-watch-logs"),
					resource.TestCheckResourceAttrPair(resourceName, "log_group_name", cloudwatchLogGroupResourceName, "name"),
					resource.TestCheckResourceAttr(resourceName, "max_aggregation_interval", "600"),
					resource.TestCheckResourceAttrPair(resourceName, "subnet_id", subnetResourceName, "id"),
					resource.TestCheckResourceAttr(resourceName, "traffic_type", "ALL"),
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

func TestAccAWSFlowLog_LogDestinationType_CloudWatchLogs(t *testing.T) {
	var flowLog ec2.FlowLog
	cloudwatchLogGroupResourceName := "aws_cloudwatch_log_group.test"
	resourceName := "aws_flow_log.test"
	rName := acctest.RandomWithPrefix("tf-acc-test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		ErrorCheck:   testAccErrorCheck(t, ec2.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckFlowLogDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccFlowLogConfig_LogDestinationType_CloudWatchLogs(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFlowLogExists(resourceName, &flowLog),
					// We automatically trim :* from ARNs if present
					testAccCheckResourceAttrRegionalARN(resourceName, "log_destination", "logs", fmt.Sprintf("log-group:%s", rName)),
					resource.TestCheckResourceAttr(resourceName, "log_destination_type", "cloud-watch-logs"),
					resource.TestCheckResourceAttrPair(resourceName, "log_group_name", cloudwatchLogGroupResourceName, "name"),
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

func TestAccAWSFlowLog_LogDestinationType_S3(t *testing.T) {
	var flowLog ec2.FlowLog
	s3ResourceName := "aws_s3_bucket.test"
	resourceName := "aws_flow_log.test"
	rName := acctest.RandomWithPrefix("tf-acc-test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		ErrorCheck:   testAccErrorCheck(t, ec2.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckFlowLogDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccFlowLogConfig_LogDestinationType_S3(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFlowLogExists(resourceName, &flowLog),
					resource.TestCheckResourceAttrPair(resourceName, "log_destination", s3ResourceName, "arn"),
					resource.TestCheckResourceAttr(resourceName, "log_destination_type", "s3"),
					resource.TestCheckResourceAttr(resourceName, "log_group_name", ""),
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

func TestAccAWSFlowLog_LogDestinationType_S3_Invalid(t *testing.T) {
	rName := acctest.RandomWithPrefix("tf-acc-test-flow-log-s3-invalid")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		ErrorCheck:   testAccErrorCheck(t, ec2.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckFlowLogDestroy,
		Steps: []resource.TestStep{
			{
				Config:      testAccFlowLogConfig_LogDestinationType_S3_Invalid(rName),
				ExpectError: regexp.MustCompile(`(Access Denied for LogDestination|does not exist)`),
			},
		},
	})
}

func TestAccAWSFlowLog_LogDestinationType_S3_DO_PlainText(t *testing.T) {
	var flowLog ec2.FlowLog
	s3ResourceName := "aws_s3_bucket.test"
	resourceName := "aws_flow_log.test"
	rName := acctest.RandomWithPrefix("tf-acc-test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		ErrorCheck:   testAccErrorCheck(t, ec2.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckFlowLogDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccFlowLogConfig_LogDestinationType_S3_DO_PlainText(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFlowLogExists(resourceName, &flowLog),
					resource.TestCheckResourceAttrPair(resourceName, "log_destination", s3ResourceName, "arn"),
					resource.TestCheckResourceAttr(resourceName, "log_destination_type", "s3"),
					resource.TestCheckResourceAttr(resourceName, "log_group_name", ""),
					resource.TestCheckResourceAttr(resourceName, "destination_options.0.file_format", "plain-text"),
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

func TestAccAWSFlowLog_LogDestinationType_S3_DO_PlainText_HiveCompatible(t *testing.T) {
	var flowLog ec2.FlowLog
	s3ResourceName := "aws_s3_bucket.test"
	resourceName := "aws_flow_log.test"
	rName := acctest.RandomWithPrefix("tf-acc-test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		ErrorCheck:   testAccErrorCheck(t, ec2.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckFlowLogDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccFlowLogConfig_LogDestinationType_S3_DO_PlainText_HiveCompatible_PerHour(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFlowLogExists(resourceName, &flowLog),
					resource.TestCheckResourceAttrPair(resourceName, "log_destination", s3ResourceName, "arn"),
					resource.TestCheckResourceAttr(resourceName, "log_destination_type", "s3"),
					resource.TestCheckResourceAttr(resourceName, "log_group_name", ""),
					resource.TestCheckResourceAttr(resourceName, "destination_options.0.file_format", "plain-text"),
					resource.TestCheckResourceAttr(resourceName, "destination_options.0.hive_compatible_partitions", "true"),
					resource.TestCheckResourceAttr(resourceName, "destination_options.0.per_hour_partition", "true"),
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

func TestAccAWSFlowLog_LogDestinationType_S3_DO_Parquet(t *testing.T) {
	var flowLog ec2.FlowLog
	s3ResourceName := "aws_s3_bucket.test"
	resourceName := "aws_flow_log.test"
	rName := acctest.RandomWithPrefix("tf-acc-test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		ErrorCheck:   testAccErrorCheck(t, ec2.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckFlowLogDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccFlowLogConfig_LogDestinationType_S3_DO_Parquet(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFlowLogExists(resourceName, &flowLog),
					resource.TestCheckResourceAttrPair(resourceName, "log_destination", s3ResourceName, "arn"),
					resource.TestCheckResourceAttr(resourceName, "log_destination_type", "s3"),
					resource.TestCheckResourceAttr(resourceName, "log_group_name", ""),
					resource.TestCheckResourceAttr(resourceName, "destination_options.0.file_format", "parquet"),
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

func TestAccAWSFlowLog_LogDestinationType_S3_DO_Parquet_HiveCompatible(t *testing.T) {
	var flowLog ec2.FlowLog
	s3ResourceName := "aws_s3_bucket.test"
	resourceName := "aws_flow_log.test"
	rName := acctest.RandomWithPrefix("tf-acc-test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		ErrorCheck:   testAccErrorCheck(t, ec2.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckFlowLogDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccFlowLogConfig_LogDestinationType_S3_DO_Parquet_HiveCompatible(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFlowLogExists(resourceName, &flowLog),
					resource.TestCheckResourceAttrPair(resourceName, "log_destination", s3ResourceName, "arn"),
					resource.TestCheckResourceAttr(resourceName, "log_destination_type", "s3"),
					resource.TestCheckResourceAttr(resourceName, "log_group_name", ""),
					resource.TestCheckResourceAttr(resourceName, "destination_options.0.file_format", "parquet"),
					resource.TestCheckResourceAttr(resourceName, "destination_options.0.hive_compatible_partitions", "true"),
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

func TestAccAWSFlowLog_LogDestinationType_S3_DO_Parquet_HiveCompatible_PerHour(t *testing.T) {
	var flowLog ec2.FlowLog
	s3ResourceName := "aws_s3_bucket.test"
	resourceName := "aws_flow_log.test"
	rName := acctest.RandomWithPrefix("tf-acc-test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		ErrorCheck:   testAccErrorCheck(t, ec2.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckFlowLogDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccFlowLogConfig_LogDestinationType_S3_DO_Parquet_HiveCompatible_PerHour(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFlowLogExists(resourceName, &flowLog),
					resource.TestCheckResourceAttrPair(resourceName, "log_destination", s3ResourceName, "arn"),
					resource.TestCheckResourceAttr(resourceName, "log_destination_type", "s3"),
					resource.TestCheckResourceAttr(resourceName, "log_group_name", ""),
					resource.TestCheckResourceAttr(resourceName, "destination_options.0.file_format", "parquet"),
					resource.TestCheckResourceAttr(resourceName, "destination_options.0.hive_compatible_partitions", "true"),
					resource.TestCheckResourceAttr(resourceName, "destination_options.0.per_hour_partition", "true"),
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

func TestAccAWSFlowLog_LogDestinationType_MaxAggregationInterval(t *testing.T) {
	var flowLog ec2.FlowLog
	resourceName := "aws_flow_log.test"
	rName := acctest.RandomWithPrefix("tf-acc-test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		ErrorCheck:   testAccErrorCheck(t, ec2.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckFlowLogDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccFlowLogConfig_MaxAggregationInterval(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFlowLogExists(resourceName, &flowLog),
					resource.TestCheckResourceAttr(resourceName, "max_aggregation_interval", "60"),
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

func TestAccAWSFlowLog_tags(t *testing.T) {
	var flowLog ec2.FlowLog
	resourceName := "aws_flow_log.test"
	rName := acctest.RandomWithPrefix("tf-acc-test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		ErrorCheck:   testAccErrorCheck(t, ec2.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckFlowLogDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccFlowLogConfigTags1(rName, "key1", "value1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFlowLogExists(resourceName, &flowLog),
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
				Config: testAccFlowLogConfigTags2(rName, "key1", "value1updated", "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFlowLogExists(resourceName, &flowLog),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1updated"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
			{
				Config: testAccFlowLogConfigTags1(rName, "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFlowLogExists(resourceName, &flowLog),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
		},
	})
}

func TestAccAWSFlowLog_disappears(t *testing.T) {
	var flowLog ec2.FlowLog
	resourceName := "aws_flow_log.test"
	rName := acctest.RandomWithPrefix("tf-acc-test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		ErrorCheck:   testAccErrorCheck(t, ec2.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckFlowLogDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccFlowLogConfig_VPCID(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFlowLogExists(resourceName, &flowLog),
					testAccCheckResourceDisappears(testAccProvider, resourceAwsFlowLog(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckFlowLogExists(n string, v *ec2.FlowLog) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No Flow Log ID is set")
		}

		conn := testAccProvider.Meta().(*AWSClient).ec2conn

		output, err := finder.FlowLogByID(conn, rs.Primary.ID)

		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

func testAccCheckFlowLogDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*AWSClient).ec2conn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_flow_log" {
			continue
		}

		_, err := finder.FlowLogByID(conn, rs.Primary.ID)

		if tfresource.NotFound(err) {
			continue
		}

		if err != nil {
			return err
		}

		return fmt.Errorf("Flow Log %s still exists", rs.Primary.ID)
	}

	return nil
}

func testAccFlowLogConfigBase(rName string) string {
	return fmt.Sprintf(`
resource "aws_vpc" "test" {
  cidr_block = "10.0.0.0/16"

  tags = {
    Name = %[1]q
  }
}
`, rName)
}

func testAccFlowLogConfig_LogDestinationType_CloudWatchLogs(rName string) string {
	return testAccFlowLogConfigBase(rName) + fmt.Sprintf(`
data "aws_partition" "current" {}

resource "aws_iam_role" "test" {
  name = %[1]q

  assume_role_policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Principal": {
        "Service": [
          "ec2.${data.aws_partition.current.dns_suffix}"
        ]
      },
      "Action": [
        "sts:AssumeRole"
      ]
    }
  ]
}
EOF
}

resource "aws_cloudwatch_log_group" "test" {
  name = %[1]q
}

resource "aws_flow_log" "test" {
  iam_role_arn         = aws_iam_role.test.arn
  log_destination      = aws_cloudwatch_log_group.test.arn
  log_destination_type = "cloud-watch-logs"
  traffic_type         = "ALL"
  vpc_id               = aws_vpc.test.id
}
`, rName)
}

func testAccFlowLogConfig_LogDestinationType_S3(rName string) string {
	return testAccFlowLogConfigBase(rName) + fmt.Sprintf(`
resource "aws_s3_bucket" "test" {
  bucket        = %[1]q
  force_destroy = true
}

resource "aws_flow_log" "test" {
  log_destination      = aws_s3_bucket.test.arn
  log_destination_type = "s3"
  traffic_type         = "ALL"
  vpc_id               = aws_vpc.test.id
}
`, rName)
}

func testAccFlowLogConfig_LogDestinationType_S3_Invalid(rName string) string {
	return testAccFlowLogConfigBase(rName) + `
data "aws_partition" "current" {}

resource "aws_flow_log" "test" {
  log_destination      = "arn:${data.aws_partition.current.partition}:s3:::does-not-exist"
  log_destination_type = "s3"
  traffic_type         = "ALL"
  vpc_id               = aws_vpc.test.id
}
`
}

func testAccFlowLogConfig_LogDestinationType_S3_DO_PlainText(rName string) string {
	return testAccFlowLogConfigBase(rName) + fmt.Sprintf(`
resource "aws_s3_bucket" "test" {
  bucket        = %[1]q
  force_destroy = true
}

resource "aws_flow_log" "test" {
  log_destination      = aws_s3_bucket.test.arn
  log_destination_type = "s3"
  traffic_type         = "ALL"
  vpc_id               = aws_vpc.test.id
  destination_options {
    file_format = "plain-text"
  }
}
`, rName)
}

func testAccFlowLogConfig_LogDestinationType_S3_DO_PlainText_HiveCompatible_PerHour(rName string) string {
	return testAccFlowLogConfigBase(rName) + fmt.Sprintf(`
resource "aws_s3_bucket" "test" {
  bucket        = %[1]q
  force_destroy = true
}

resource "aws_flow_log" "test" {
  log_destination      = aws_s3_bucket.test.arn
  log_destination_type = "s3"
  traffic_type         = "ALL"
  vpc_id               = aws_vpc.test.id
  destination_options {
    file_format                = "plain-text"
    hive_compatible_partitions = true
    per_hour_partition         = true
  }
}
`, rName)
}

func testAccFlowLogConfig_LogDestinationType_S3_DO_Parquet(rName string) string {
	return testAccFlowLogConfigBase(rName) + fmt.Sprintf(`
resource "aws_s3_bucket" "test" {
  bucket        = %[1]q
  force_destroy = true
}

resource "aws_flow_log" "test" {
  log_destination      = aws_s3_bucket.test.arn
  log_destination_type = "s3"
  traffic_type         = "ALL"
  vpc_id               = aws_vpc.test.id
  destination_options {
    file_format = "parquet"
  }
}
`, rName)
}

func testAccFlowLogConfig_LogDestinationType_S3_DO_Parquet_HiveCompatible(rName string) string {
	return testAccFlowLogConfigBase(rName) + fmt.Sprintf(`
resource "aws_s3_bucket" "test" {
  bucket        = %[1]q
  force_destroy = true
}

resource "aws_flow_log" "test" {
  log_destination      = aws_s3_bucket.test.arn
  log_destination_type = "s3"
  traffic_type         = "ALL"
  vpc_id               = aws_vpc.test.id
  destination_options {
    file_format                = "parquet"
    hive_compatible_partitions = true
  }
}
`, rName)
}

func testAccFlowLogConfig_LogDestinationType_S3_DO_Parquet_HiveCompatible_PerHour(rName string) string {
	return testAccFlowLogConfigBase(rName) + fmt.Sprintf(`
resource "aws_s3_bucket" "test" {
  bucket        = %[1]q
  force_destroy = true
}

resource "aws_flow_log" "test" {
  log_destination      = aws_s3_bucket.test.arn
  log_destination_type = "s3"
  traffic_type         = "ALL"
  vpc_id               = aws_vpc.test.id
  destination_options {
    file_format                = "parquet"
    hive_compatible_partitions = true
    per_hour_partition         = true
  }
}
`, rName)
}

func testAccFlowLogConfig_SubnetID(rName string) string {
	return testAccFlowLogConfigBase(rName) + fmt.Sprintf(`
resource "aws_subnet" "test" {
  cidr_block = "10.0.1.0/24"
  vpc_id     = aws_vpc.test.id

  tags = {
    Name = %[1]q
  }
}

data "aws_partition" "current" {}

resource "aws_iam_role" "test" {
  name = %[1]q

  assume_role_policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Principal": {
        "Service": [
          "ec2.${data.aws_partition.current.dns_suffix}"
        ]
      },
      "Action": [
        "sts:AssumeRole"
      ]
    }
  ]
}
EOF
}

resource "aws_cloudwatch_log_group" "test" {
  name = %[1]q
}

resource "aws_flow_log" "test" {
  iam_role_arn   = aws_iam_role.test.arn
  log_group_name = aws_cloudwatch_log_group.test.name
  subnet_id      = aws_subnet.test.id
  traffic_type   = "ALL"
}
`, rName)
}

func testAccFlowLogConfig_VPCID(rName string) string {
	return testAccFlowLogConfigBase(rName) + fmt.Sprintf(`
data "aws_partition" "current" {}

resource "aws_iam_role" "test" {
  name = %[1]q

  assume_role_policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Principal": {
        "Service": [
          "ec2.${data.aws_partition.current.dns_suffix}"
        ]
      },
      "Action": [
        "sts:AssumeRole"
      ]
    }
  ]
}
EOF
}

resource "aws_cloudwatch_log_group" "test" {
  name = %[1]q
}

resource "aws_flow_log" "test" {
  iam_role_arn   = aws_iam_role.test.arn
  log_group_name = aws_cloudwatch_log_group.test.name
  traffic_type   = "ALL"
  vpc_id         = aws_vpc.test.id
}
`, rName)
}

func testAccFlowLogConfig_LogFormat(rName string) string {
	return testAccFlowLogConfigBase(rName) + fmt.Sprintf(`
data "aws_partition" "current" {}

resource "aws_iam_role" "test" {
  name = %[1]q

  assume_role_policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Principal": {
        "Service": [
          "ec2.${data.aws_partition.current.dns_suffix}"
        ]
      },
      "Action": [
        "sts:AssumeRole"
      ]
    }
  ]
}
EOF
}

resource "aws_cloudwatch_log_group" "test" {
  name = %[1]q
}

resource "aws_s3_bucket" "test" {
  bucket        = %[1]q
  force_destroy = true
}

resource "aws_flow_log" "test" {
  log_destination      = aws_s3_bucket.test.arn
  log_destination_type = "s3"
  traffic_type         = "ALL"
  vpc_id               = aws_vpc.test.id
  log_format           = "$${version} $${vpc-id} $${subnet-id}"
}
`, rName)
}

func testAccFlowLogConfigTags1(rName, tagKey1, tagValue1 string) string {
	return testAccFlowLogConfigBase(rName) + fmt.Sprintf(`
data "aws_partition" "current" {}

resource "aws_iam_role" "test" {
  name = %[1]q

  assume_role_policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Principal": {
        "Service": [
          "ec2.${data.aws_partition.current.dns_suffix}"
        ]
      },
      "Action": [
        "sts:AssumeRole"
      ]
    }
  ]
}
EOF
}

resource "aws_cloudwatch_log_group" "test" {
  name = %[1]q
}

resource "aws_flow_log" "test" {
  iam_role_arn   = aws_iam_role.test.arn
  log_group_name = aws_cloudwatch_log_group.test.name
  traffic_type   = "ALL"
  vpc_id         = aws_vpc.test.id

  tags = {
    %[2]q = %[3]q
  }
}
`, rName, tagKey1, tagValue1)
}

func testAccFlowLogConfigTags2(rName, tagKey1, tagValue1, tagKey2, tagValue2 string) string {
	return testAccFlowLogConfigBase(rName) + fmt.Sprintf(`
data "aws_partition" "current" {}

resource "aws_iam_role" "test" {
  name = %[1]q

  assume_role_policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Principal": {
        "Service": [
          "ec2.${data.aws_partition.current.dns_suffix}"
        ]
      },
      "Action": [
        "sts:AssumeRole"
      ]
    }
  ]
}
EOF
}

resource "aws_cloudwatch_log_group" "test" {
  name = %[1]q
}

resource "aws_flow_log" "test" {
  iam_role_arn   = aws_iam_role.test.arn
  log_group_name = aws_cloudwatch_log_group.test.name
  traffic_type   = "ALL"
  vpc_id         = aws_vpc.test.id

  tags = {
    %[2]q = %[3]q
    %[4]q = %[5]q
  }
}
`, rName, tagKey1, tagValue1, tagKey2, tagValue2)
}

func testAccFlowLogConfig_MaxAggregationInterval(rName string) string {
	return testAccFlowLogConfigBase(rName) + fmt.Sprintf(`
data "aws_partition" "current" {}

resource "aws_iam_role" "test" {
  name = %[1]q

  assume_role_policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Principal": {
        "Service": [
          "ec2.${data.aws_partition.current.dns_suffix}"
        ]
      },
      "Action": [
        "sts:AssumeRole"
      ]
    }
  ]
}
EOF
}

resource "aws_cloudwatch_log_group" "test" {
  name = %[1]q
}

resource "aws_flow_log" "test" {
  iam_role_arn   = aws_iam_role.test.arn
  log_group_name = aws_cloudwatch_log_group.test.name
  traffic_type   = "ALL"
  vpc_id         = aws_vpc.test.id

  max_aggregation_interval = 60
}
`, rName)
}
