// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package cloudwatch_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/service/cloudwatch/types"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfcloudwatch "github.com/hashicorp/terraform-provider-aws/internal/service/cloudwatch"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccCloudWatchMetricAlarm_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var alarm types.MetricAlarm
	resourceName := "aws_cloudwatch_metric_alarm.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.CloudWatchServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckMetricAlarmDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccMetricAlarmConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckMetricAlarmExists(ctx, resourceName, &alarm),
					resource.TestCheckResourceAttr(resourceName, names.AttrMetricName, "CPUUtilization"),
					resource.TestCheckResourceAttr(resourceName, "statistic", "Average"),
					acctest.MatchResourceAttrRegionalARN(resourceName, names.AttrARN, "cloudwatch", regexache.MustCompile(`alarm:.+`)),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "alarm_description", "This metric monitors ec2 cpu utilization"),
					resource.TestCheckResourceAttr(resourceName, "threshold", "80"),
					resource.TestCheckResourceAttr(resourceName, "period", "120"),
					resource.TestCheckResourceAttr(resourceName, names.AttrNamespace, "AWS/EC2"),
					resource.TestCheckResourceAttr(resourceName, "alarm_name", rName),
					resource.TestCheckResourceAttr(resourceName, "comparison_operator", "GreaterThanOrEqualToThreshold"),
					resource.TestCheckResourceAttr(resourceName, "datapoints_to_alarm", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "evaluation_periods", acctest.Ct2),
					resource.TestCheckResourceAttr(resourceName, "insufficient_data_actions.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "dimensions.%", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "dimensions.InstanceId", "i-abcd1234"),
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
	ctx := acctest.Context(t)
	var alarm types.MetricAlarm
	resourceName := "aws_cloudwatch_metric_alarm.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.CloudWatchServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckMetricAlarmDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccMetricAlarmConfig_actionsEC2Automate(rName, "reboot"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckMetricAlarmExists(ctx, resourceName, &alarm),
					resource.TestCheckResourceAttr(resourceName, "alarm_actions.#", acctest.Ct1),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccMetricAlarmConfig_actionsEC2Automate(rName, "recover"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckMetricAlarmExists(ctx, resourceName, &alarm),
					resource.TestCheckResourceAttr(resourceName, "alarm_actions.#", acctest.Ct1),
				),
			},
			{
				Config: testAccMetricAlarmConfig_actionsEC2Automate(rName, "stop"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckMetricAlarmExists(ctx, resourceName, &alarm),
					resource.TestCheckResourceAttr(resourceName, "alarm_actions.#", acctest.Ct1),
				),
			},
			{
				Config: testAccMetricAlarmConfig_actionsEC2Automate(rName, "terminate"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckMetricAlarmExists(ctx, resourceName, &alarm),
					resource.TestCheckResourceAttr(resourceName, "alarm_actions.#", acctest.Ct1),
				),
			},
		},
	})
}

func TestAccCloudWatchMetricAlarm_AlarmActions_snsTopic(t *testing.T) {
	ctx := acctest.Context(t)
	var alarm types.MetricAlarm
	resourceName := "aws_cloudwatch_metric_alarm.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.CloudWatchServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckMetricAlarmDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccMetricAlarmConfig_actionsSNSTopic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckMetricAlarmExists(ctx, resourceName, &alarm),
					resource.TestCheckResourceAttr(resourceName, "alarm_actions.#", acctest.Ct1),
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
	ctx := acctest.Context(t)
	var alarm types.MetricAlarm
	resourceName := "aws_cloudwatch_metric_alarm.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.CloudWatchServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckMetricAlarmDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccMetricAlarmConfig_actionsSWFAction(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckMetricAlarmExists(ctx, resourceName, &alarm),
					resource.TestCheckResourceAttr(resourceName, "alarm_actions.#", acctest.Ct1),
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
	ctx := acctest.Context(t)
	var alarm types.MetricAlarm
	resourceName := "aws_cloudwatch_metric_alarm.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.CloudWatchServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckMetricAlarmDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccMetricAlarmConfig_datapointsTo(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckMetricAlarmExists(ctx, resourceName, &alarm),
					resource.TestCheckResourceAttr(resourceName, "datapoints_to_alarm", acctest.Ct2),
				),
			},
		},
	})
}

func TestAccCloudWatchMetricAlarm_treatMissingData(t *testing.T) {
	ctx := acctest.Context(t)
	var alarm types.MetricAlarm
	resourceName := "aws_cloudwatch_metric_alarm.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.CloudWatchServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckMetricAlarmDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccMetricAlarmConfig_treatMissingData(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckMetricAlarmExists(ctx, resourceName, &alarm),
					resource.TestCheckResourceAttr(resourceName, "treat_missing_data", "missing"),
				),
			},
			{
				Config: testAccMetricAlarmConfig_treatMissingDataUpdate(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckMetricAlarmExists(ctx, resourceName, &alarm),
					resource.TestCheckResourceAttr(resourceName, "treat_missing_data", "breaching"),
				),
			},
			{
				Config: testAccMetricAlarmConfig_treatMissingDataNoAttr(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckMetricAlarmExists(ctx, resourceName, &alarm),
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
	ctx := acctest.Context(t)
	var alarm types.MetricAlarm
	resourceName := "aws_cloudwatch_metric_alarm.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.CloudWatchServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckMetricAlarmDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccMetricAlarmConfig_treatEvaluateLowSampleCountPercentiles(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckMetricAlarmExists(ctx, resourceName, &alarm),
					resource.TestCheckResourceAttr(resourceName, "evaluate_low_sample_count_percentiles", "evaluate"),
				),
			},
			{
				Config: testAccMetricAlarmConfig_treatEvaluateLowSampleCountPercentilesUpdated(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckMetricAlarmExists(ctx, resourceName, &alarm),
					resource.TestCheckResourceAttr(resourceName, "evaluate_low_sample_count_percentiles", "ignore"),
				),
			},
		},
	})
}

func TestAccCloudWatchMetricAlarm_extendedStatistic(t *testing.T) {
	ctx := acctest.Context(t)
	var alarm types.MetricAlarm
	resourceName := "aws_cloudwatch_metric_alarm.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.CloudWatchServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckMetricAlarmDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config:      testAccMetricAlarmConfig_extendedStatistic(rName, "IQM(1:2)"), // IQM accepts no args
				ExpectError: regexache.MustCompile(`invalid statistic, see: https:\/\/docs\.aws\.amazon\.com\/.*`),
			},
			{
				Config:      testAccMetricAlarmConfig_extendedStatistic(rName, "iqm10"), // IQM accepts no args
				ExpectError: regexache.MustCompile(`invalid statistic, see: https:\/\/docs\.aws\.amazon\.com\/.*`),
			},
			// {  TODO: more complex regex to reject this
			// 	Config: testAccMetricAlarmConfig_extendedStatistic(rName, "PR(5%:10%)"),  // PR args must be absolute
			// 	ExpectError: regexache.MustCompile(`invalid statistic, see: https:\/\/docs\.aws\.amazon\.com\/.*`),
			// },
			// {  TODO: more complex regex to reject this
			// 	Config: testAccMetricAlarmConfig_extendedStatistic(rName, "TC(:)"),  // at least one arg must be provided
			// 	ExpectError: regexache.MustCompile(`invalid statistic, see: https:\/\/docs\.aws\.amazon\.com\/.*`),
			// },
			{
				Config:      testAccMetricAlarmConfig_extendedStatistic(rName, "WM"), // missing syntax
				ExpectError: regexache.MustCompile(`invalid statistic, see: https:\/\/docs\.aws\.amazon\.com\/.*`),
			},
			{
				Config:      testAccMetricAlarmConfig_extendedStatistic(rName, "p"), // missing arg
				ExpectError: regexache.MustCompile(`invalid statistic, see: https:\/\/docs\.aws\.amazon\.com\/.*`),
			},
			{
				Config:      testAccMetricAlarmConfig_extendedStatistic(rName, "AB(1:2)"), // unknown stat 'AB'
				ExpectError: regexache.MustCompile(`invalid statistic, see: https:\/\/docs\.aws\.amazon\.com\/.*`),
			},
			{
				Config:      testAccMetricAlarmConfig_extendedStatistic(rName, "cd42"), // unknown stat 'cd'
				ExpectError: regexache.MustCompile(`invalid statistic, see: https:\/\/docs\.aws\.amazon\.com\/.*`),
			},
			{
				Config: testAccMetricAlarmConfig_extendedStatistic(rName, "p88.0"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckMetricAlarmExists(ctx, resourceName, &alarm),
					resource.TestCheckResourceAttr(resourceName, "extended_statistic", "p88.0"),
				),
			},
			{
				Config: testAccMetricAlarmConfig_extendedStatistic(rName, "p0.0"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckMetricAlarmExists(ctx, resourceName, &alarm),
					resource.TestCheckResourceAttr(resourceName, "extended_statistic", "p0.0"),
				),
			},
			{
				Config: testAccMetricAlarmConfig_extendedStatistic(rName, "p100"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckMetricAlarmExists(ctx, resourceName, &alarm),
					resource.TestCheckResourceAttr(resourceName, "extended_statistic", "p100"),
				),
			},
			{
				Config: testAccMetricAlarmConfig_extendedStatistic(rName, "p95"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckMetricAlarmExists(ctx, resourceName, &alarm),
					resource.TestCheckResourceAttr(resourceName, "extended_statistic", "p95"),
				),
			},
			{
				Config: testAccMetricAlarmConfig_extendedStatistic(rName, "tm90"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckMetricAlarmExists(ctx, resourceName, &alarm),
					resource.TestCheckResourceAttr(resourceName, "extended_statistic", "tm90"),
				),
			},
			{
				Config: testAccMetricAlarmConfig_extendedStatistic(rName, "TM(2%:98%)"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckMetricAlarmExists(ctx, resourceName, &alarm),
					resource.TestCheckResourceAttr(resourceName, "extended_statistic", "TM(2%:98%)"),
				),
			},
			{
				Config: testAccMetricAlarmConfig_extendedStatistic(rName, "TM(150:1000)"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckMetricAlarmExists(ctx, resourceName, &alarm),
					resource.TestCheckResourceAttr(resourceName, "extended_statistic", "TM(150:1000)"),
				),
			},
			{
				Config: testAccMetricAlarmConfig_extendedStatistic(rName, "IQM"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckMetricAlarmExists(ctx, resourceName, &alarm),
					resource.TestCheckResourceAttr(resourceName, "extended_statistic", "IQM"),
				),
			},
			{
				Config: testAccMetricAlarmConfig_extendedStatistic(rName, "wm98"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckMetricAlarmExists(ctx, resourceName, &alarm),
					resource.TestCheckResourceAttr(resourceName, "extended_statistic", "wm98"),
				),
			},
			{
				Config: testAccMetricAlarmConfig_extendedStatistic(rName, "PR(:300)"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckMetricAlarmExists(ctx, resourceName, &alarm),
					resource.TestCheckResourceAttr(resourceName, "extended_statistic", "PR(:300)"),
				),
			},
			{
				Config: testAccMetricAlarmConfig_extendedStatistic(rName, "PR(100:2000)"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckMetricAlarmExists(ctx, resourceName, &alarm),
					resource.TestCheckResourceAttr(resourceName, "extended_statistic", "PR(100:2000)"),
				),
			},
			{
				Config: testAccMetricAlarmConfig_extendedStatistic(rName, "tc90"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckMetricAlarmExists(ctx, resourceName, &alarm),
					resource.TestCheckResourceAttr(resourceName, "extended_statistic", "tc90"),
				),
			},
			{
				Config: testAccMetricAlarmConfig_extendedStatistic(rName, "TC(0.005:0.030)"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckMetricAlarmExists(ctx, resourceName, &alarm),
					resource.TestCheckResourceAttr(resourceName, "extended_statistic", "TC(0.005:0.030)"),
				),
			},
			{
				Config: testAccMetricAlarmConfig_extendedStatistic(rName, "TS(80%:)"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckMetricAlarmExists(ctx, resourceName, &alarm),
					resource.TestCheckResourceAttr(resourceName, "extended_statistic", "TS(80%:)"),
				),
			},
			{
				Config: testAccMetricAlarmConfig_extendedStatistic(rName, "TC(:0.5)"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckMetricAlarmExists(ctx, resourceName, &alarm),
					resource.TestCheckResourceAttr(resourceName, "extended_statistic", "TC(:0.5)"),
				),
			},
		},
	})
}

func TestAccCloudWatchMetricAlarm_metricQuery(t *testing.T) {
	ctx := acctest.Context(t)
	var alarm types.MetricAlarm
	resourceName := "aws_cloudwatch_metric_alarm.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.CloudWatchServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckMetricAlarmDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config:      testAccMetricAlarmConfig_badMetricQuery(rName),
				ExpectError: regexache.MustCompile("No metric_query may have both `expression` and a `metric` specified"),
			},
			{
				Config: testAccMetricAlarmConfig_metricQueryExpressionQuery(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckMetricAlarmExists(ctx, resourceName, &alarm),
					resource.TestCheckResourceAttr(resourceName, "metric_query.#", acctest.Ct1),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "metric_query.*", map[string]string{
						names.AttrID:         "m1",
						names.AttrExpression: "SELECT MAX(MillisBehindLatest) FROM SCHEMA(\"foo\", Operation, ShardId) WHERE Operation = 'ProcessTask'",
						"period":             "60",
						"label":              "cat",
						"return_data":        acctest.CtTrue,
					}),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"metric_query"},
			},
			{
				Config: testAccMetricAlarmConfig_metricQueryExpressionReference(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckMetricAlarmExists(ctx, resourceName, &alarm),
					resource.TestCheckResourceAttr(resourceName, "metric_query.#", acctest.Ct2),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "metric_query.*", map[string]string{
						names.AttrID:         "e1",
						names.AttrExpression: "m1",
						"label":              "cat",
						"return_data":        acctest.CtTrue,
						"period":             "",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "metric_query.*", map[string]string{
						names.AttrID:                     "m1",
						"metric.#":                       acctest.Ct1,
						"metric.0.metric_name":           "CPUUtilization",
						"metric.0.namespace":             "AWS/EC2",
						"metric.0.period":                "120",
						"metric.0.stat":                  "Average",
						"metric.0.unit":                  "Count",
						"metric.0.dimensions.%":          acctest.Ct1,
						"metric.0.dimensions.InstanceId": "i-abcd1234",
						"period":                         "",
					}),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"metric_query"},
			},
			{
				Config: testAccMetricAlarmConfig_metricQueryCrossAccount(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckMetricAlarmExists(ctx, resourceName, &alarm),
					resource.TestCheckResourceAttr(resourceName, "metric_query.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "metric_query.0.id", "m1"),
					resource.TestCheckResourceAttrPair(resourceName, "metric_query.0.account_id", "data.aws_caller_identity.current", names.AttrAccountID),
					resource.TestCheckNoResourceAttr(resourceName, "metric_query.0.period"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"metric_query"},
			},
			{
				Config: testAccMetricAlarmConfig_metricQueryExpressionReferenceUpdated(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckMetricAlarmExists(ctx, resourceName, &alarm),
					resource.TestCheckResourceAttr(resourceName, "metric_query.#", acctest.Ct3),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "metric_query.*", map[string]string{
						names.AttrID:         "e1",
						names.AttrExpression: "m1",
						"label":              "cat",
						"return_data":        "",
						"period":             "",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "metric_query.*", map[string]string{
						names.AttrID:         "e2",
						names.AttrExpression: "e1",
						"label":              "bug",
						"return_data":        acctest.CtTrue,
						"period":             "",
					}),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"metric_query"},
			},
			{
				Config: testAccMetricAlarmConfig_metricQueryExpressionReference(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckMetricAlarmExists(ctx, resourceName, &alarm),
					resource.TestCheckResourceAttr(resourceName, "metric_query.#", acctest.Ct2),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"metric_query"},
			},
			{
				Config: testAccMetricAlarmConfig_anomalyDetectionExpression(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckMetricAlarmExists(ctx, resourceName, &alarm),
					resource.TestCheckResourceAttr(resourceName, "metric_query.#", acctest.Ct2),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "metric_query.*", map[string]string{
						names.AttrID:         "e1",
						names.AttrExpression: "ANOMALY_DETECTION_BAND(m1)",
						"label":              "CPUUtilization (Expected)",
						"return_data":        acctest.CtTrue,
						"period":             "",
					}),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"metric_query"},
			},
		},
	})
}

func TestAccCloudWatchMetricAlarm_missingStatistic(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.CloudWatchServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckMetricAlarmDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config:      testAccMetricAlarmConfig_missingStatistic(rName),
				ExpectError: regexache.MustCompile("One of `statistic` or `extended_statistic` must be set for a cloudwatch metric alarm"),
			},
		},
	})
}

func TestAccCloudWatchMetricAlarm_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	var alarm types.MetricAlarm
	resourceName := "aws_cloudwatch_metric_alarm.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.CloudWatchServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckMetricAlarmDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccMetricAlarmConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckMetricAlarmExists(ctx, resourceName, &alarm),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfcloudwatch.ResourceMetricAlarm(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckMetricAlarmExists(ctx context.Context, n string, v *types.MetricAlarm) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).CloudWatchClient(ctx)

		output, err := tfcloudwatch.FindMetricAlarmByName(ctx, conn, rs.Primary.ID)

		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

func testAccCheckMetricAlarmDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).CloudWatchClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_cloudwatch_metric_alarm" {
				continue
			}

			_, err := tfcloudwatch.FindMetricAlarmByName(ctx, conn, rs.Primary.ID)

			if tfresource.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("CloudWatch Metric Alarm %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccMetricAlarmConfig_basic(rName string) string {
	return fmt.Sprintf(`
resource "aws_cloudwatch_metric_alarm" "test" {
  alarm_name                = %[1]q
  comparison_operator       = "GreaterThanOrEqualToThreshold"
  evaluation_periods        = 2
  metric_name               = "CPUUtilization"
  namespace                 = "AWS/EC2"
  period                    = 120
  statistic                 = "Average"
  threshold                 = 80
  alarm_description         = "This metric monitors ec2 cpu utilization"
  insufficient_data_actions = []

  dimensions = {
    InstanceId = "i-abcd1234"
  }
}
`, rName)
}

func testAccMetricAlarmConfig_datapointsTo(rName string) string {
	return fmt.Sprintf(`
resource "aws_cloudwatch_metric_alarm" "test" {
  alarm_name                = %[1]q
  comparison_operator       = "GreaterThanOrEqualToThreshold"
  datapoints_to_alarm       = 2
  evaluation_periods        = 2
  metric_name               = "CPUUtilization"
  namespace                 = "AWS/EC2"
  period                    = 120
  statistic                 = "Average"
  threshold                 = 80
  alarm_description         = "This metric monitors ec2 cpu utilization"
  insufficient_data_actions = []

  dimensions = {
    InstanceId = "i-abcd1234"
  }
}
`, rName)
}

func testAccMetricAlarmConfig_treatMissingData(rName string) string {
	return fmt.Sprintf(`
resource "aws_cloudwatch_metric_alarm" "test" {
  alarm_name                = %[1]q
  comparison_operator       = "GreaterThanOrEqualToThreshold"
  evaluation_periods        = 2
  metric_name               = "CPUUtilization"
  namespace                 = "AWS/EC2"
  period                    = 120
  statistic                 = "Average"
  threshold                 = 80
  alarm_description         = "This metric monitors ec2 cpu utilization"
  treat_missing_data        = "missing"
  insufficient_data_actions = []

  dimensions = {
    InstanceId = "i-abcd1234"
  }
}
`, rName)
}

func testAccMetricAlarmConfig_treatMissingDataUpdate(rName string) string {
	return fmt.Sprintf(`
resource "aws_cloudwatch_metric_alarm" "test" {
  alarm_name                = %[1]q
  comparison_operator       = "GreaterThanOrEqualToThreshold"
  evaluation_periods        = 2
  metric_name               = "CPUUtilization"
  namespace                 = "AWS/EC2"
  period                    = 120
  statistic                 = "Average"
  threshold                 = 80
  alarm_description         = "This metric monitors ec2 cpu utilization"
  treat_missing_data        = "breaching"
  insufficient_data_actions = []

  dimensions = {
    InstanceId = "i-abcd1234"
  }
}
`, rName)
}

func testAccMetricAlarmConfig_treatMissingDataNoAttr(rName string) string {
	return fmt.Sprintf(`
resource "aws_cloudwatch_metric_alarm" "test" {
  alarm_name                = %[1]q
  comparison_operator       = "GreaterThanOrEqualToThreshold"
  evaluation_periods        = 2
  metric_name               = "CPUUtilization"
  namespace                 = "AWS/EC2"
  period                    = 120
  statistic                 = "Average"
  threshold                 = 80
  alarm_description         = "This metric monitors ec2 cpu utilization"
  insufficient_data_actions = []

  dimensions = {
    InstanceId = "i-abcd1234"
  }
}
`, rName)
}

func testAccMetricAlarmConfig_treatEvaluateLowSampleCountPercentiles(rName string) string {
	return fmt.Sprintf(`
resource "aws_cloudwatch_metric_alarm" "test" {
  alarm_name                            = %[1]q
  comparison_operator                   = "GreaterThanOrEqualToThreshold"
  evaluation_periods                    = 2
  metric_name                           = "CPUUtilization"
  namespace                             = "AWS/EC2"
  period                                = 120
  extended_statistic                    = "p88.0"
  threshold                             = 80
  alarm_description                     = "This metric monitors ec2 cpu utilization"
  evaluate_low_sample_count_percentiles = "evaluate"
  insufficient_data_actions             = []

  dimensions = {
    InstanceId = "i-abcd1234"
  }
}
`, rName)
}

func testAccMetricAlarmConfig_treatEvaluateLowSampleCountPercentilesUpdated(rName string) string {
	return fmt.Sprintf(`
resource "aws_cloudwatch_metric_alarm" "test" {
  alarm_name                            = %[1]q
  comparison_operator                   = "GreaterThanOrEqualToThreshold"
  evaluation_periods                    = 2
  metric_name                           = "CPUUtilization"
  namespace                             = "AWS/EC2"
  period                                = 120
  extended_statistic                    = "p88.0"
  threshold                             = 80
  alarm_description                     = "This metric monitors ec2 cpu utilization"
  evaluate_low_sample_count_percentiles = "ignore"
  insufficient_data_actions             = []

  dimensions = {
    InstanceId = "i-abcd1234"
  }
}
`, rName)
}

func testAccMetricAlarmConfig_extendedStatistic(rName, stat string) string {
	return fmt.Sprintf(`
resource "aws_cloudwatch_metric_alarm" "test" {
  alarm_name                = %[1]q
  comparison_operator       = "GreaterThanOrEqualToThreshold"
  evaluation_periods        = 2
  metric_name               = "CPUUtilization"
  namespace                 = "AWS/EC2"
  period                    = 120
  extended_statistic        = %[2]q
  threshold                 = 80
  alarm_description         = "This metric monitors ec2 cpu utilization"
  insufficient_data_actions = []

  dimensions = {
    InstanceId = "i-abcd1234"
  }
}
`, rName, stat)
}

func testAccMetricAlarmConfig_missingStatistic(rName string) string {
	return fmt.Sprintf(`
resource "aws_cloudwatch_metric_alarm" "test" {
  alarm_name                = %[1]q
  comparison_operator       = "GreaterThanOrEqualToThreshold"
  evaluation_periods        = 2
  metric_name               = "CPUUtilization"
  namespace                 = "AWS/EC2"
  period                    = 120
  threshold                 = 80
  alarm_description         = "This metric monitors ec2 cpu utilization"
  insufficient_data_actions = []

  dimensions = {
    InstanceId = "i-abcd1234"
  }
}
`, rName)
}

func testAccMetricAlarmConfig_metricQueryExpressionReference(rName string) string {
	return fmt.Sprintf(`
resource "aws_cloudwatch_metric_alarm" "test" {
  alarm_name                = %[1]q
  comparison_operator       = "GreaterThanOrEqualToThreshold"
  evaluation_periods        = 2
  threshold                 = 80
  alarm_description         = "This metric monitors ec2 cpu utilization"
  insufficient_data_actions = []

  metric_query {
    id          = "e1"
    expression  = "m1"
    label       = "cat"
    return_data = true
  }

  metric_query {
    id = "m1"

    metric {
      metric_name = "CPUUtilization"
      namespace   = "AWS/EC2"
      period      = 120
      stat        = "Average"
      unit        = "Count"

      dimensions = {
        InstanceId = "i-abcd1234"
      }
    }
  }
}
`, rName)
}

func testAccMetricAlarmConfig_metricQueryExpressionQuery(rName string) string {
	return fmt.Sprintf(`
resource "aws_cloudwatch_metric_alarm" "test" {
  alarm_name          = %[1]q
  comparison_operator = "GreaterThanOrEqualToThreshold"
  evaluation_periods  = 3
  datapoints_to_alarm = 3
  threshold           = 30000
  treat_missing_data  = "breaching"

  metric_query {
    id          = "m1"
    expression  = "SELECT MAX(MillisBehindLatest) FROM SCHEMA(\"foo\", Operation, ShardId) WHERE Operation = 'ProcessTask'"
    period      = 60
    label       = "cat"
    return_data = true
  }
}
`, rName)
}

func testAccMetricAlarmConfig_metricQueryCrossAccount(rName string) string {
	return fmt.Sprintf(`
data "aws_caller_identity" "current" {}

resource "aws_cloudwatch_metric_alarm" "test" {
  alarm_name                = %[1]q
  comparison_operator       = "GreaterThanOrEqualToThreshold"
  evaluation_periods        = 2
  threshold                 = 80
  alarm_description         = "This metric monitors ec2 cpu utilization"
  insufficient_data_actions = []

  metric_query {
    id          = "m1"
    account_id  = data.aws_caller_identity.current.account_id
    return_data = true

    metric {
      metric_name = "CPUUtilization"
      namespace   = "AWS/EC2"
      period      = 120
      stat        = "Average"
      unit        = "Count"

      dimensions = {
        InstanceId = "i-abcd1234"
      }
    }
  }
}
`, rName)
}

func testAccMetricAlarmConfig_anomalyDetectionExpression(rName string) string {
	return fmt.Sprintf(`
resource "aws_cloudwatch_metric_alarm" "test" {
  alarm_name                = %[1]q
  comparison_operator       = "GreaterThanUpperThreshold"
  evaluation_periods        = 2
  threshold_metric_id       = "e1"
  alarm_description         = "This metric monitors ec2 cpu utilization"
  insufficient_data_actions = []

  metric_query {
    id          = "e1"
    expression  = "ANOMALY_DETECTION_BAND(m1)"
    label       = "CPUUtilization (Expected)"
    return_data = true
  }

  metric_query {
    id          = "m1"
    return_data = true

    metric {
      metric_name = "CPUUtilization"
      namespace   = "AWS/EC2"
      period      = 120
      stat        = "Average"
      unit        = "Count"

      dimensions = {
        InstanceId = "i-abcd1234"
      }
    }
  }
}
`, rName)
}

func testAccMetricAlarmConfig_metricQueryExpressionReferenceUpdated(rName string) string {
	return fmt.Sprintf(`
resource "aws_cloudwatch_metric_alarm" "test" {
  alarm_name                = %[1]q
  comparison_operator       = "GreaterThanOrEqualToThreshold"
  evaluation_periods        = 2
  threshold                 = 80
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
    return_data = true
  }

  metric_query {
    id = "m1"

    metric {
      metric_name = "CPUUtilization"
      namespace   = "AWS/EC2"
      period      = 120
      stat        = "p95.45"
      unit        = "Count"

      dimensions = {
        InstanceId = "i-abcd1234"
      }
    }
  }
}
`, rName)
}

func testAccMetricAlarmConfig_badMetricQuery(rName string) string {
	return fmt.Sprintf(`
resource "aws_cloudwatch_metric_alarm" "test" {
  alarm_name                = %[1]q
  comparison_operator       = "GreaterThanOrEqualToThreshold"
  evaluation_periods        = 2
  threshold                 = 80
  alarm_description         = "This metric monitors ec2 cpu utilization"
  insufficient_data_actions = []

  metric_query {
    id         = "e1"
    expression = "m1"
    label      = "cat"

    metric {
      metric_name = "CPUUtilization"
      namespace   = "AWS/EC2"
      period      = 120
      stat        = "Average"
      unit        = "Count"

      dimensions = {
        InstanceId = "i-abcd1234"
      }
    }
  }
}
`, rName)
}

// EC2 Automate requires a valid EC2 instance
// ValidationError: Invalid use of EC2 'Recover' action. i-abcd1234 is not a valid EC2 instance.
func testAccMetricAlarmConfig_actionsEC2Automate(rName, action string) string {
	return acctest.ConfigCompose(
		acctest.ConfigLatestAmazonLinux2HVMEBSX8664AMI(),
		acctest.ConfigVPCWithSubnets(rName, 1),
		acctest.AvailableEC2InstanceTypeForAvailabilityZone("aws_subnet.test[0].availability_zone", "t3.micro", "t2.micro"),
		fmt.Sprintf(`
data "aws_partition" "current" {}
data "aws_region" "current" {}

resource "aws_instance" "test" {
  ami           = data.aws_ami.amzn2-ami-minimal-hvm-ebs-x86_64.id
  instance_type = data.aws_ec2_instance_type_offering.available.instance_type
  subnet_id     = aws_subnet.test[0].id

  tags = {
    Name = %[1]q
  }
}

resource "aws_cloudwatch_metric_alarm" "test" {
  alarm_actions       = ["arn:${data.aws_partition.current.partition}:automate:${data.aws_region.current.name}:ec2:%[2]s"]
  alarm_description   = "Status checks have failed for system"
  alarm_name          = %[1]q
  comparison_operator = "GreaterThanThreshold"
  evaluation_periods  = 2
  metric_name         = "StatusCheckFailed_System"
  namespace           = "AWS/EC2"
  period              = 60
  statistic           = "Minimum"
  threshold           = 0
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
  name = %[1]q
}

resource "aws_cloudwatch_metric_alarm" "test" {
  alarm_actions       = [aws_sns_topic.test.arn]
  alarm_description   = "Status checks have failed for system"
  alarm_name          = %[1]q
  comparison_operator = "GreaterThanThreshold"
  evaluation_periods  = 2
  metric_name         = "StatusCheckFailed_System"
  namespace           = "AWS/EC2"
  period              = 60
  statistic           = "Minimum"
  threshold           = 0
  unit                = "Count"

  dimensions = {
    InstanceId = "i-abcd1234"
  }
}
`, rName)
}

func testAccMetricAlarmConfig_actionsSWFAction(rName string) string {
	return fmt.Sprintf(`
data "aws_caller_identity" "current" {}
data "aws_partition" "current" {}
data "aws_region" "current" {}

resource "aws_cloudwatch_metric_alarm" "test" {
  alarm_actions       = ["arn:${data.aws_partition.current.partition}:swf:${data.aws_region.current.name}:${data.aws_caller_identity.current.account_id}:action/actions/AWS_EC2.InstanceId.Reboot/1.0"]
  alarm_description   = "Status checks have failed, rebooting system."
  alarm_name          = %[1]q
  comparison_operator = "GreaterThanThreshold"
  evaluation_periods  = 5
  metric_name         = "StatusCheckFailed_Instance"
  namespace           = "AWS/EC2"
  period              = 60
  statistic           = "Minimum"
  threshold           = 0
  unit                = "Count"

  dimensions = {
    InstanceId = "i-abcd1234"
  }
}
`, rName)
}
