package cloudwatch_test

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/cloudwatch"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfcloudwatch "github.com/hashicorp/terraform-provider-aws/internal/service/cloudwatch"
)

func TestAccCloudWatchMetricAlarm_basic(t *testing.T) {
	var alarm cloudwatch.MetricAlarm
	resourceName := "aws_cloudwatch_metric_alarm.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, cloudwatch.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckMetricAlarmDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccMetricAlarmConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMetricAlarmExists(resourceName, &alarm),
					resource.TestCheckResourceAttr(resourceName, "metric_name", "CPUUtilization"),
					resource.TestCheckResourceAttr(resourceName, "statistic", "Average"),
					acctest.MatchResourceAttrRegionalARN(resourceName, "arn", "cloudwatch", regexp.MustCompile(`alarm:.+`)),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
					resource.TestCheckResourceAttr(resourceName, "alarm_description", "This metric monitors ec2 cpu utilization"),
					resource.TestCheckResourceAttr(resourceName, "threshold", "80"),
					resource.TestCheckResourceAttr(resourceName, "period", "120"),
					resource.TestCheckResourceAttr(resourceName, "namespace", "AWS/EC2"),
					resource.TestCheckResourceAttr(resourceName, "alarm_name", rName),
					resource.TestCheckResourceAttr(resourceName, "comparison_operator", "GreaterThanOrEqualToThreshold"),
					resource.TestCheckResourceAttr(resourceName, "datapoints_to_alarm", "0"),
					resource.TestCheckResourceAttr(resourceName, "evaluation_periods", "2"),
					resource.TestCheckResourceAttr(resourceName, "insufficient_data_actions.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "dimensions.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "dimensions.InstanceId", "i-abc123"),
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

func TestAccCloudWatchMetricAlarm_AlarmActions_ec2Automate(t *testing.T) {
	var alarm cloudwatch.MetricAlarm
	resourceName := "aws_cloudwatch_metric_alarm.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, cloudwatch.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckMetricAlarmDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccMetricAlarmConfig_actionsEC2Automate(rName, "reboot"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMetricAlarmExists(resourceName, &alarm),
					resource.TestCheckResourceAttr(resourceName, "alarm_actions.#", "1"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccMetricAlarmConfig_actionsEC2Automate(rName, "recover"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMetricAlarmExists(resourceName, &alarm),
					resource.TestCheckResourceAttr(resourceName, "alarm_actions.#", "1"),
				),
			},
			{
				Config: testAccMetricAlarmConfig_actionsEC2Automate(rName, "stop"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMetricAlarmExists(resourceName, &alarm),
					resource.TestCheckResourceAttr(resourceName, "alarm_actions.#", "1"),
				),
			},
			{
				Config: testAccMetricAlarmConfig_actionsEC2Automate(rName, "terminate"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMetricAlarmExists(resourceName, &alarm),
					resource.TestCheckResourceAttr(resourceName, "alarm_actions.#", "1"),
				),
			},
		},
	})
}

func TestAccCloudWatchMetricAlarm_AlarmActions_snsTopic(t *testing.T) {
	var alarm cloudwatch.MetricAlarm
	resourceName := "aws_cloudwatch_metric_alarm.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, cloudwatch.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckMetricAlarmDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccMetricAlarmConfig_actionsSNSTopic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMetricAlarmExists(resourceName, &alarm),
					resource.TestCheckResourceAttr(resourceName, "alarm_actions.#", "1"),
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

func TestAccCloudWatchMetricAlarm_AlarmActions_swfAction(t *testing.T) {
	var alarm cloudwatch.MetricAlarm
	resourceName := "aws_cloudwatch_metric_alarm.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, cloudwatch.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckMetricAlarmDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccMetricAlarmConfig_actionsSWFAction(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMetricAlarmExists(resourceName, &alarm),
					resource.TestCheckResourceAttr(resourceName, "alarm_actions.#", "1"),
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

func TestAccCloudWatchMetricAlarm_dataPointsToAlarm(t *testing.T) {
	var alarm cloudwatch.MetricAlarm
	resourceName := "aws_cloudwatch_metric_alarm.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, cloudwatch.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckMetricAlarmDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccMetricAlarmConfig_datapointsTo(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMetricAlarmExists(resourceName, &alarm),
					resource.TestCheckResourceAttr(resourceName, "datapoints_to_alarm", "2"),
				),
			},
		},
	})
}

func TestAccCloudWatchMetricAlarm_treatMissingData(t *testing.T) {
	var alarm cloudwatch.MetricAlarm
	resourceName := "aws_cloudwatch_metric_alarm.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, cloudwatch.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckMetricAlarmDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccMetricAlarmConfig_treatMissingData(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMetricAlarmExists(resourceName, &alarm),
					resource.TestCheckResourceAttr(resourceName, "treat_missing_data", "missing"),
				),
			},
			{
				Config: testAccMetricAlarmConfig_treatMissingDataUpdate(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMetricAlarmExists(resourceName, &alarm),
					resource.TestCheckResourceAttr(resourceName, "treat_missing_data", "breaching"),
				),
			},
			{
				Config: testAccMetricAlarmConfig_treatMissingDataNoAttr(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMetricAlarmExists(resourceName, &alarm),
					resource.TestCheckResourceAttr(resourceName, "treat_missing_data", "missing"),
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

func TestAccCloudWatchMetricAlarm_evaluateLowSampleCountPercentiles(t *testing.T) {
	var alarm cloudwatch.MetricAlarm
	resourceName := "aws_cloudwatch_metric_alarm.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, cloudwatch.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckMetricAlarmDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccMetricAlarmConfig_treatEvaluateLowSampleCountPercentiles(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMetricAlarmExists(resourceName, &alarm),
					resource.TestCheckResourceAttr(resourceName, "evaluate_low_sample_count_percentiles", "evaluate"),
				),
			},
			{
				Config: testAccMetricAlarmConfig_treatEvaluateLowSampleCountPercentilesUpdated(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMetricAlarmExists(resourceName, &alarm),
					resource.TestCheckResourceAttr(resourceName, "evaluate_low_sample_count_percentiles", "ignore"),
				),
			},
		},
	})
}

func TestAccCloudWatchMetricAlarm_extendedStatistic(t *testing.T) {
	var alarm cloudwatch.MetricAlarm
	resourceName := "aws_cloudwatch_metric_alarm.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, cloudwatch.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckMetricAlarmDestroy,
		Steps: []resource.TestStep{
			{
				Config:      testAccMetricAlarmConfig_extendedStatistic(rName, "IQM(1:2)"), // IQM accepts no args
				ExpectError: regexp.MustCompile(`invalid statistic, see: https:\/\/docs\.aws\.amazon\.com\/.*`),
			},
			{
				Config:      testAccMetricAlarmConfig_extendedStatistic(rName, "iqm10"), // IQM accepts no args
				ExpectError: regexp.MustCompile(`invalid statistic, see: https:\/\/docs\.aws\.amazon\.com\/.*`),
			},
			// {  TODO: more complex regex to reject this
			// 	Config: testAccMetricAlarmConfig_extendedStatistic(rName, "PR(5%:10%)"),  // PR args must be absolute
			// 	ExpectError: regexp.MustCompile(`invalid statistic, see: https:\/\/docs\.aws\.amazon\.com\/.*`),
			// },
			// {  TODO: more complex regex to reject this
			// 	Config: testAccMetricAlarmConfig_extendedStatistic(rName, "TC(:)"),  // at least one arg must be provided
			// 	ExpectError: regexp.MustCompile(`invalid statistic, see: https:\/\/docs\.aws\.amazon\.com\/.*`),
			// },
			{
				Config:      testAccMetricAlarmConfig_extendedStatistic(rName, "WM"), // missing syntax
				ExpectError: regexp.MustCompile(`invalid statistic, see: https:\/\/docs\.aws\.amazon\.com\/.*`),
			},
			{
				Config:      testAccMetricAlarmConfig_extendedStatistic(rName, "p"), // missing arg
				ExpectError: regexp.MustCompile(`invalid statistic, see: https:\/\/docs\.aws\.amazon\.com\/.*`),
			},
			{
				Config:      testAccMetricAlarmConfig_extendedStatistic(rName, "AB(1:2)"), // unknown stat 'AB'
				ExpectError: regexp.MustCompile(`invalid statistic, see: https:\/\/docs\.aws\.amazon\.com\/.*`),
			},
			{
				Config:      testAccMetricAlarmConfig_extendedStatistic(rName, "cd42"), // unknown stat 'cd'
				ExpectError: regexp.MustCompile(`invalid statistic, see: https:\/\/docs\.aws\.amazon\.com\/.*`),
			},
			{
				Config: testAccMetricAlarmConfig_extendedStatistic(rName, "p88.0"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMetricAlarmExists(resourceName, &alarm),
					resource.TestCheckResourceAttr(resourceName, "extended_statistic", "p88.0"),
				),
			},
			{
				Config: testAccMetricAlarmConfig_extendedStatistic(rName, "p0.0"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMetricAlarmExists(resourceName, &alarm),
					resource.TestCheckResourceAttr(resourceName, "extended_statistic", "p0.0"),
				),
			},
			{
				Config: testAccMetricAlarmConfig_extendedStatistic(rName, "p100"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMetricAlarmExists(resourceName, &alarm),
					resource.TestCheckResourceAttr(resourceName, "extended_statistic", "p100"),
				),
			},
			{
				Config: testAccMetricAlarmConfig_extendedStatistic(rName, "p95"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMetricAlarmExists(resourceName, &alarm),
					resource.TestCheckResourceAttr(resourceName, "extended_statistic", "p95"),
				),
			},
			{
				Config: testAccMetricAlarmConfig_extendedStatistic(rName, "tm90"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMetricAlarmExists(resourceName, &alarm),
					resource.TestCheckResourceAttr(resourceName, "extended_statistic", "tm90"),
				),
			},
			{
				Config: testAccMetricAlarmConfig_extendedStatistic(rName, "TM(2%:98%)"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMetricAlarmExists(resourceName, &alarm),
					resource.TestCheckResourceAttr(resourceName, "extended_statistic", "TM(2%:98%)"),
				),
			},
			{
				Config: testAccMetricAlarmConfig_extendedStatistic(rName, "TM(150:1000)"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMetricAlarmExists(resourceName, &alarm),
					resource.TestCheckResourceAttr(resourceName, "extended_statistic", "TM(150:1000)"),
				),
			},
			{
				Config: testAccMetricAlarmConfig_extendedStatistic(rName, "IQM"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMetricAlarmExists(resourceName, &alarm),
					resource.TestCheckResourceAttr(resourceName, "extended_statistic", "IQM"),
				),
			},
			{
				Config: testAccMetricAlarmConfig_extendedStatistic(rName, "wm98"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMetricAlarmExists(resourceName, &alarm),
					resource.TestCheckResourceAttr(resourceName, "extended_statistic", "wm98"),
				),
			},
			{
				Config: testAccMetricAlarmConfig_extendedStatistic(rName, "PR(:300)"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMetricAlarmExists(resourceName, &alarm),
					resource.TestCheckResourceAttr(resourceName, "extended_statistic", "PR(:300)"),
				),
			},
			{
				Config: testAccMetricAlarmConfig_extendedStatistic(rName, "PR(100:2000)"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMetricAlarmExists(resourceName, &alarm),
					resource.TestCheckResourceAttr(resourceName, "extended_statistic", "PR(100:2000)"),
				),
			},
			{
				Config: testAccMetricAlarmConfig_extendedStatistic(rName, "tc90"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMetricAlarmExists(resourceName, &alarm),
					resource.TestCheckResourceAttr(resourceName, "extended_statistic", "tc90"),
				),
			},
			{
				Config: testAccMetricAlarmConfig_extendedStatistic(rName, "TC(0.005:0.030)"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMetricAlarmExists(resourceName, &alarm),
					resource.TestCheckResourceAttr(resourceName, "extended_statistic", "TC(0.005:0.030)"),
				),
			},
			{
				Config: testAccMetricAlarmConfig_extendedStatistic(rName, "TS(80%:)"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMetricAlarmExists(resourceName, &alarm),
					resource.TestCheckResourceAttr(resourceName, "extended_statistic", "TS(80%:)"),
				),
			},
			{
				Config: testAccMetricAlarmConfig_extendedStatistic(rName, "TC(:0.5)"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMetricAlarmExists(resourceName, &alarm),
					resource.TestCheckResourceAttr(resourceName, "extended_statistic", "TC(:0.5)"),
				),
			},
		},
	})
}

func TestAccCloudWatchMetricAlarm_expression(t *testing.T) {
	var alarm cloudwatch.MetricAlarm
	resourceName := "aws_cloudwatch_metric_alarm.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, cloudwatch.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckMetricAlarmDestroy,
		Steps: []resource.TestStep{
			{
				Config:      testAccMetricAlarmConfig_badExpression(rName),
				ExpectError: regexp.MustCompile("No metric_query may have both `expression` and a `metric` specified"),
			},
			{
				Config: testAccMetricAlarmConfig_expression(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMetricAlarmExists(resourceName, &alarm),
					resource.TestCheckResourceAttr(resourceName, "metric_query.#", "2"),
				),
			},
			{
				Config: testAccMetricAlarmConfig_crossAccount(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMetricAlarmExists(resourceName, &alarm),
					resource.TestCheckResourceAttrPair(resourceName, "metric_query.0.account_id", "data.aws_caller_identity.current", "account_id"),
				),
			},
			{
				Config: testAccMetricAlarmConfig_expressionUpdated(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMetricAlarmExists(resourceName, &alarm),
					resource.TestCheckResourceAttr(resourceName, "metric_query.#", "3"),
				),
			},
			{
				Config: testAccMetricAlarmConfig_expression(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMetricAlarmExists(resourceName, &alarm),
					resource.TestCheckResourceAttr(resourceName, "metric_query.#", "2"),
				),
			},
			{
				Config: testAccMetricAlarmConfig_anomalyDetectionExpression(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMetricAlarmExists(resourceName, &alarm),
					resource.TestCheckResourceAttr(resourceName, "metric_query.#", "2"),
				),
			},
			{
				Config: testAccMetricAlarmConfig_expressionQueryUpdated(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMetricAlarmExists(resourceName, &alarm),
					resource.TestCheckResourceAttr(resourceName, "metric_query.#", "2"),
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

func TestAccCloudWatchMetricAlarm_missingStatistic(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, cloudwatch.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckMetricAlarmDestroy,
		Steps: []resource.TestStep{
			{
				Config:      testAccMetricAlarmConfig_missingStatistic(rName),
				ExpectError: regexp.MustCompile("One of `statistic` or `extended_statistic` must be set for a cloudwatch metric alarm"),
			},
		},
	})
}

func TestAccCloudWatchMetricAlarm_tags(t *testing.T) {
	var alarm cloudwatch.MetricAlarm
	resourceName := "aws_cloudwatch_metric_alarm.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, cloudwatch.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckMetricAlarmDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccMetricAlarmConfig_tags1(rName, "key1", "value1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMetricAlarmExists(resourceName, &alarm),
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
				Config: testAccMetricAlarmConfig_tags2(rName, "key1", "value1updated", "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMetricAlarmExists(resourceName, &alarm),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1updated"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
			{
				Config: testAccMetricAlarmConfig_tags1(rName, "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMetricAlarmExists(resourceName, &alarm),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
		},
	})
}

func TestAccCloudWatchMetricAlarm_disappears(t *testing.T) {
	var alarm cloudwatch.MetricAlarm
	resourceName := "aws_cloudwatch_metric_alarm.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, cloudwatch.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckMetricAlarmDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccMetricAlarmConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMetricAlarmExists(resourceName, &alarm),
					acctest.CheckResourceDisappears(acctest.Provider, tfcloudwatch.ResourceMetricAlarm(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckMetricAlarmExists(n string, alarm *cloudwatch.MetricAlarm) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).CloudWatchConn
		resp, err := tfcloudwatch.FindMetricAlarmByName(conn, rs.Primary.ID)
		if err != nil {
			return err
		}
		if resp == nil {
			return fmt.Errorf("Alarm not found")
		}
		*alarm = *resp

		return nil
	}
}

func testAccCheckMetricAlarmDestroy(s *terraform.State) error {
	conn := acctest.Provider.Meta().(*conns.AWSClient).CloudWatchConn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_cloudwatch_metric_alarm" {
			continue
		}

		resp, err := tfcloudwatch.FindMetricAlarmByName(conn, rs.Primary.ID)
		if err == nil {
			if resp != nil && aws.StringValue(resp.AlarmName) == rs.Primary.ID {
				return fmt.Errorf("Alarm Still Exists: %s", rs.Primary.ID)
			}
		}
	}

	return nil
}

func testAccMetricAlarmConfig_basic(rName string) string {
	return fmt.Sprintf(`
resource "aws_cloudwatch_metric_alarm" "test" {
  alarm_name                = "%s"
  comparison_operator       = "GreaterThanOrEqualToThreshold"
  evaluation_periods        = "2"
  metric_name               = "CPUUtilization"
  namespace                 = "AWS/EC2"
  period                    = "120"
  statistic                 = "Average"
  threshold                 = "80"
  alarm_description         = "This metric monitors ec2 cpu utilization"
  insufficient_data_actions = []

  dimensions = {
    InstanceId = "i-abc123"
  }
}
`, rName)
}

func testAccMetricAlarmConfig_datapointsTo(rName string) string {
	return fmt.Sprintf(`
resource "aws_cloudwatch_metric_alarm" "test" {
  alarm_name                = "%s"
  comparison_operator       = "GreaterThanOrEqualToThreshold"
  datapoints_to_alarm       = "2"
  evaluation_periods        = "2"
  metric_name               = "CPUUtilization"
  namespace                 = "AWS/EC2"
  period                    = "120"
  statistic                 = "Average"
  threshold                 = "80"
  alarm_description         = "This metric monitors ec2 cpu utilization"
  insufficient_data_actions = []

  dimensions = {
    InstanceId = "i-abc123"
  }
}
`, rName)
}

func testAccMetricAlarmConfig_treatMissingData(rName string) string {
	return fmt.Sprintf(`
resource "aws_cloudwatch_metric_alarm" "test" {
  alarm_name                = "%s"
  comparison_operator       = "GreaterThanOrEqualToThreshold"
  evaluation_periods        = "2"
  metric_name               = "CPUUtilization"
  namespace                 = "AWS/EC2"
  period                    = "120"
  statistic                 = "Average"
  threshold                 = "80"
  alarm_description         = "This metric monitors ec2 cpu utilization"
  treat_missing_data        = "missing"
  insufficient_data_actions = []

  dimensions = {
    InstanceId = "i-abc123"
  }
}
`, rName)
}

func testAccMetricAlarmConfig_treatMissingDataUpdate(rName string) string {
	return fmt.Sprintf(`
resource "aws_cloudwatch_metric_alarm" "test" {
  alarm_name                = "%s"
  comparison_operator       = "GreaterThanOrEqualToThreshold"
  evaluation_periods        = "2"
  metric_name               = "CPUUtilization"
  namespace                 = "AWS/EC2"
  period                    = "120"
  statistic                 = "Average"
  threshold                 = "80"
  alarm_description         = "This metric monitors ec2 cpu utilization"
  treat_missing_data        = "breaching"
  insufficient_data_actions = []

  dimensions = {
    InstanceId = "i-abc123"
  }
}
`, rName)
}

func testAccMetricAlarmConfig_treatMissingDataNoAttr(rName string) string {
	return fmt.Sprintf(`
resource "aws_cloudwatch_metric_alarm" "test" {
  alarm_name                = "%s"
  comparison_operator       = "GreaterThanOrEqualToThreshold"
  evaluation_periods        = "2"
  metric_name               = "CPUUtilization"
  namespace                 = "AWS/EC2"
  period                    = "120"
  statistic                 = "Average"
  threshold                 = "80"
  alarm_description         = "This metric monitors ec2 cpu utilization"
  insufficient_data_actions = []

  dimensions = {
    InstanceId = "i-abc123"
  }
}
`, rName)
}

func testAccMetricAlarmConfig_treatEvaluateLowSampleCountPercentiles(rName string) string {
	return fmt.Sprintf(`
resource "aws_cloudwatch_metric_alarm" "test" {
  alarm_name                            = "%s"
  comparison_operator                   = "GreaterThanOrEqualToThreshold"
  evaluation_periods                    = "2"
  metric_name                           = "CPUUtilization"
  namespace                             = "AWS/EC2"
  period                                = "120"
  extended_statistic                    = "p88.0"
  threshold                             = "80"
  alarm_description                     = "This metric monitors ec2 cpu utilization"
  evaluate_low_sample_count_percentiles = "evaluate"
  insufficient_data_actions             = []

  dimensions = {
    InstanceId = "i-abc123"
  }
}
`, rName)
}

func testAccMetricAlarmConfig_treatEvaluateLowSampleCountPercentilesUpdated(rName string) string {
	return fmt.Sprintf(`
resource "aws_cloudwatch_metric_alarm" "test" {
  alarm_name                            = "%s"
  comparison_operator                   = "GreaterThanOrEqualToThreshold"
  evaluation_periods                    = "2"
  metric_name                           = "CPUUtilization"
  namespace                             = "AWS/EC2"
  period                                = "120"
  extended_statistic                    = "p88.0"
  threshold                             = "80"
  alarm_description                     = "This metric monitors ec2 cpu utilization"
  evaluate_low_sample_count_percentiles = "ignore"
  insufficient_data_actions             = []

  dimensions = {
    InstanceId = "i-abc123"
  }
}
`, rName)
}

func testAccMetricAlarmConfig_extendedStatistic(rName string, stat string) string {
	return fmt.Sprintf(`
resource "aws_cloudwatch_metric_alarm" "test" {
  alarm_name                = "%s"
  comparison_operator       = "GreaterThanOrEqualToThreshold"
  evaluation_periods        = "2"
  metric_name               = "CPUUtilization"
  namespace                 = "AWS/EC2"
  period                    = "120"
  extended_statistic        = "%s"
  threshold                 = "80"
  alarm_description         = "This metric monitors ec2 cpu utilization"
  insufficient_data_actions = []

  dimensions = {
    InstanceId = "i-abc123"
  }
}
`, rName, stat)
}

func testAccMetricAlarmConfig_missingStatistic(rName string) string {
	return fmt.Sprintf(`
resource "aws_cloudwatch_metric_alarm" "test" {
  alarm_name                = "%s"
  comparison_operator       = "GreaterThanOrEqualToThreshold"
  evaluation_periods        = "2"
  metric_name               = "CPUUtilization"
  namespace                 = "AWS/EC2"
  period                    = "120"
  threshold                 = "80"
  alarm_description         = "This metric monitors ec2 cpu utilization"
  insufficient_data_actions = []

  dimensions = {
    InstanceId = "i-abc123"
  }
}
`, rName)
}

func testAccMetricAlarmConfig_expression(rName string) string {
	return fmt.Sprintf(`
resource "aws_cloudwatch_metric_alarm" "test" {
  alarm_name                = "%s"
  comparison_operator       = "GreaterThanOrEqualToThreshold"
  evaluation_periods        = "2"
  threshold                 = "80"
  alarm_description         = "This metric monitors ec2 cpu utilization"
  insufficient_data_actions = []

  metric_query {
    id          = "e1"
    expression  = "m1"
    label       = "cat"
    return_data = "true"
  }

  metric_query {
    id = "m1"

    metric {
      metric_name = "CPUUtilization"
      namespace   = "AWS/EC2"
      period      = "120"
      stat        = "Average"
      unit        = "Count"

      dimensions = {
        InstanceId = "i-abc123"
      }
    }
  }
}
`, rName)
}

func testAccMetricAlarmConfig_crossAccount(rName string) string {
	return fmt.Sprintf(`
data "aws_caller_identity" "current" {}

resource "aws_cloudwatch_metric_alarm" "test" {
  alarm_name                = "%s"
  comparison_operator       = "GreaterThanOrEqualToThreshold"
  evaluation_periods        = "2"
  threshold                 = "80"
  alarm_description         = "This metric monitors ec2 cpu utilization"
  insufficient_data_actions = []

  metric_query {
    id          = "m1"
    account_id  = data.aws_caller_identity.current.account_id
    return_data = "true"

    metric {
      metric_name = "CPUUtilization"
      namespace   = "AWS/EC2"
      period      = "120"
      stat        = "Average"
      unit        = "Count"

      dimensions = {
        InstanceId = "i-abc123"
      }
    }
  }
}
`, rName)
}

func testAccMetricAlarmConfig_anomalyDetectionExpression(rName string) string {
	return fmt.Sprintf(`
resource "aws_cloudwatch_metric_alarm" "test" {
  alarm_name                = "%s"
  comparison_operator       = "GreaterThanUpperThreshold"
  evaluation_periods        = "2"
  threshold_metric_id       = "e1"
  alarm_description         = "This metric monitors ec2 cpu utilization"
  insufficient_data_actions = []

  metric_query {
    id          = "e1"
    expression  = "ANOMALY_DETECTION_BAND(m1)"
    label       = "CPUUtilization (Expected)"
    return_data = "true"
  }

  metric_query {
    id          = "m1"
    return_data = "true"

    metric {
      metric_name = "CPUUtilization"
      namespace   = "AWS/EC2"
      period      = "120"
      stat        = "Average"
      unit        = "Count"

      dimensions = {
        InstanceId = "i-abc123"
      }
    }
  }
}
`, rName)
}

func testAccMetricAlarmConfig_expressionUpdated(rName string) string {
	return fmt.Sprintf(`
resource "aws_cloudwatch_metric_alarm" "test" {
  alarm_name                = "%s"
  comparison_operator       = "GreaterThanOrEqualToThreshold"
  evaluation_periods        = "2"
  threshold                 = "80"
  alarm_description         = "This metric monitors ec2 cpu utilization"
  insufficient_data_actions = []

  metric_query {
    id         = "e1"
    expression = "m1"
    label      = "cat"
  }

  metric_query {
    id          = "e2"
    expression  = "e1"
    label       = "bug"
    return_data = "true"
  }

  metric_query {
    id = "m1"

    metric {
      metric_name = "CPUUtilization"
      namespace   = "AWS/EC2"
      period      = "120"
      stat        = "p95.45"
      unit        = "Count"

      dimensions = {
        InstanceId = "i-abc123"
      }
    }
  }
}
`, rName)
}

func testAccMetricAlarmConfig_expressionQueryUpdated(rName string) string {
	return fmt.Sprintf(`
resource "aws_cloudwatch_metric_alarm" "test" {
  alarm_name                = "%s"
  comparison_operator       = "GreaterThanOrEqualToThreshold"
  evaluation_periods        = "2"
  threshold                 = "80"
  alarm_description         = "This metric monitors ec2 cpu utilization"
  insufficient_data_actions = []

  metric_query {
    id          = "e1"
    expression  = "m1"
    label       = "cat"
    return_data = "true"
  }

  metric_query {
    id = "m1"

    metric {
      metric_name = "CPUUtilization"
      namespace   = "AWS/EC2"
      period      = "120"
      stat        = "Maximum"
      unit        = "Count"

      dimensions = {
        InstanceId = "i-abc123"
      }
    }
  }
}
`, rName)
}

func testAccMetricAlarmConfig_badExpression(rName string) string {
	return fmt.Sprintf(`
resource "aws_cloudwatch_metric_alarm" "test" {
  alarm_name                = "%s"
  comparison_operator       = "GreaterThanOrEqualToThreshold"
  evaluation_periods        = "2"
  threshold                 = "80"
  alarm_description         = "This metric monitors ec2 cpu utilization"
  insufficient_data_actions = []

  metric_query {
    id         = "e1"
    expression = "m1"
    label      = "cat"

    metric {
      metric_name = "CPUUtilization"
      namespace   = "AWS/EC2"
      period      = "120"
      stat        = "Average"
      unit        = "Count"

      dimensions = {
        InstanceId = "i-abc123"
      }
    }
  }
}
`, rName)
}

// EC2 Automate requires a valid EC2 instance
// ValidationError: Invalid use of EC2 'Recover' action. i-abc123 is not a valid EC2 instance.
func testAccMetricAlarmConfig_actionsEC2Automate(rName, action string) string {
	return acctest.ConfigCompose(
		acctest.ConfigLatestAmazonLinuxHVMEBSAMI(),
		acctest.AvailableEC2InstanceTypeForAvailabilityZone("aws_subnet.test.availability_zone", "t3.micro", "t2.micro"),
		fmt.Sprintf(`
data "aws_partition" "current" {}

data "aws_region" "current" {}

resource "aws_vpc" "test" {
  cidr_block = "172.16.0.0/16"

  tags = {
    Name = %[1]q
  }
}

resource "aws_subnet" "test" {
  cidr_block = "172.16.0.0/24"
  vpc_id     = aws_vpc.test.id

  tags = {
    Name = %[1]q
  }
}

resource "aws_instance" "test" {
  ami           = data.aws_ami.amzn-ami-minimal-hvm-ebs.id
  instance_type = data.aws_ec2_instance_type_offering.available.instance_type
  subnet_id     = aws_subnet.test.id

  tags = {
    Name = %[1]q
  }
}

resource "aws_cloudwatch_metric_alarm" "test" {
  alarm_actions       = ["arn:${data.aws_partition.current.partition}:automate:${data.aws_region.current.name}:ec2:%[2]s"]
  alarm_description   = "Status checks have failed for system"
  alarm_name          = %[1]q
  comparison_operator = "GreaterThanThreshold"
  evaluation_periods  = "2"
  metric_name         = "StatusCheckFailed_System"
  namespace           = "AWS/EC2"
  period              = "60"
  statistic           = "Minimum"
  threshold           = "0"
  unit                = "Count"

  dimensions = {
    InstanceId = aws_instance.test.id
  }
}
`, rName, action))
}

func testAccMetricAlarmConfig_actionsSNSTopic(rName string) string {
	return fmt.Sprintf(`
resource "aws_sns_topic" "test" {
  name = %q
}

resource "aws_cloudwatch_metric_alarm" "test" {
  alarm_actions       = [aws_sns_topic.test.arn]
  alarm_description   = "Status checks have failed for system"
  alarm_name          = %q
  comparison_operator = "GreaterThanThreshold"
  evaluation_periods  = "2"
  metric_name         = "StatusCheckFailed_System"
  namespace           = "AWS/EC2"
  period              = "60"
  statistic           = "Minimum"
  threshold           = "0"
  unit                = "Count"

  dimensions = {
    InstanceId = "i-abc123"
  }
}
`, rName, rName)
}

func testAccMetricAlarmConfig_actionsSWFAction(rName string) string {
	return fmt.Sprintf(`
data "aws_caller_identity" "current" {
}

data "aws_partition" "current" {
}

data "aws_region" "current" {
}

resource "aws_cloudwatch_metric_alarm" "test" {
  alarm_actions       = ["arn:${data.aws_partition.current.partition}:swf:${data.aws_region.current.name}:${data.aws_caller_identity.current.account_id}:action/actions/AWS_EC2.InstanceId.Reboot/1.0"]
  alarm_description   = "Status checks have failed, rebooting system."
  alarm_name          = %q
  comparison_operator = "GreaterThanThreshold"
  evaluation_periods  = "5"
  metric_name         = "StatusCheckFailed_Instance"
  namespace           = "AWS/EC2"
  period              = "60"
  statistic           = "Minimum"
  threshold           = "0"
  unit                = "Count"

  dimensions = {
    InstanceId = "i-abc123"
  }
}
`, rName)
}

func testAccMetricAlarmConfig_tags1(rName, tagKey1, tagValue1 string) string {
	return fmt.Sprintf(`
resource "aws_cloudwatch_metric_alarm" "test" {
  alarm_name                = %[1]q
  comparison_operator       = "GreaterThanOrEqualToThreshold"
  evaluation_periods        = "2"
  metric_name               = "CPUUtilization"
  namespace                 = "AWS/EC2"
  period                    = "120"
  statistic                 = "Average"
  threshold                 = "80"
  alarm_description         = "This metric monitors ec2 cpu utilization"
  insufficient_data_actions = []

  dimensions = {
    InstanceId = "i-abc123"
  }

  tags = {
    %[2]q = %[3]q
  }
}
`, rName, tagKey1, tagValue1)
}

func testAccMetricAlarmConfig_tags2(rName, tagKey1, tagValue1, tagKey2, tagValue2 string) string {
	return fmt.Sprintf(`
resource "aws_cloudwatch_metric_alarm" "test" {
  alarm_name                = %[1]q
  comparison_operator       = "GreaterThanOrEqualToThreshold"
  evaluation_periods        = "2"
  metric_name               = "CPUUtilization"
  namespace                 = "AWS/EC2"
  period                    = "120"
  statistic                 = "Average"
  threshold                 = "80"
  alarm_description         = "This metric monitors ec2 cpu utilization"
  insufficient_data_actions = []

  dimensions = {
    InstanceId = "i-abc123"
  }

  tags = {
    %[2]q = %[3]q
    %[4]q = %[5]q
  }
}
`, rName, tagKey1, tagValue1, tagKey2, tagValue2)
}
