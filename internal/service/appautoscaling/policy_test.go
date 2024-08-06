// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package appautoscaling_test

import (
	"context"
	"fmt"
	"reflect"
	"strings"
	"testing"
	"time"

	awstypes "github.com/aws/aws-sdk-go-v2/service/applicationautoscaling/types"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfappautoscaling "github.com/hashicorp/terraform-provider-aws/internal/service/appautoscaling"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestValidatePolicyImportInput(t *testing.T) {
	t.Parallel()

	// lintignore:AWSAT003,AWSAT005
	testCases := []struct {
		input         string
		errorExpected bool
		expected      []string
	}{
		{
			input:         "appstream/fleet/sample-fleet/appstream:fleet:DesiredCapacity/test-policy-name",
			expected:      []string{"appstream", "fleet/sample-fleet", "appstream:fleet:DesiredCapacity", "test-policy-name"},
			errorExpected: false,
		},
		{
			input:         "dynamodb/table/tableName/dynamodb:table:ReadCapacityUnits/DynamoDBReadCapacityUtilization:table/tableName",
			expected:      []string{"dynamodb", "table/tableName", "dynamodb:table:ReadCapacityUnits", "DynamoDBReadCapacityUtilization:table/tableName"},
			errorExpected: false,
		},
		{
			input:         "dynamodb/table/tableName/index/indexName/dynamodb:index:ReadCapacityUnits/DynamoDBReadCapacityUtilization:table/tableName/index/indexName",
			expected:      []string{"dynamodb", "table/tableName/index/indexName", "dynamodb:index:ReadCapacityUnits", "DynamoDBReadCapacityUtilization:table/tableName/index/indexName"},
			errorExpected: false,
		},
		{
			input:         "ec2/spot-fleet-request/sfr-d77c6508-1c1d-4e79-8789-fc019ee44c96/ec2:spot-fleet-request:TargetCapacity/test-appautoscaling-policy-ruuhd",
			expected:      []string{"ec2", "spot-fleet-request/sfr-d77c6508-1c1d-4e79-8789-fc019ee44c96", "ec2:spot-fleet-request:TargetCapacity", "test-appautoscaling-policy-ruuhd"},
			errorExpected: false,
		},
		{
			input:         "ecs/service/clusterName/serviceName/ecs:service:DesiredCount/scale-down",
			expected:      []string{"ecs", "service/clusterName/serviceName", "ecs:service:DesiredCount", "scale-down"},
			errorExpected: false,
		},
		{
			input:         "elasticmapreduce/instancegroup/j-2EEZNYKUA1NTV/ig-1791Y4E1L8YI0/elasticmapreduce:instancegroup:InstanceCount/test-appautoscaling-policy-ruuhd",
			expected:      []string{"elasticmapreduce", "instancegroup/j-2EEZNYKUA1NTV/ig-1791Y4E1L8YI0", "elasticmapreduce:instancegroup:InstanceCount", "test-appautoscaling-policy-ruuhd"},
			errorExpected: false,
		},
		{
			input:         "rds/cluster:id/rds:cluster:ReadReplicaCount/cpu-auto-scaling",
			expected:      []string{"rds", "cluster:id", "rds:cluster:ReadReplicaCount", "cpu-auto-scaling"},
			errorExpected: false,
		},
		{
			input:         "dynamodb/missing/parts",
			errorExpected: true,
		},
		{
			input:         "kafka/arn:aws:kafka:us-west-2:123456789012:cluster/example/279c0212-d057-4dba-9aa9-1c4e5a25bfc7-3/kafka:broker-storage:VolumeSize/KafkaBrokerStorageUtilization-scaling-policy:arn:aws:kafka:us-west-2:123456789012:cluster/example/279c0212-d057-4dba-9aa9-1c4e5a25bfc7-3",
			expected:      []string{"kafka", "arn:aws:kafka:us-west-2:123456789012:cluster/example/279c0212-d057-4dba-9aa9-1c4e5a25bfc7-3", "kafka:broker-storage:VolumeSize", "KafkaBrokerStorageUtilization-scaling-policy:arn:aws:kafka:us-west-2:123456789012:cluster/example/279c0212-d057-4dba-9aa9-1c4e5a25bfc7-3"},
			errorExpected: false,
		},
		{
			input:         "kafka/arn:aws:kafka:us-west-2:123456789012:cluster/example/279c0212-d057-4dba-9aa9-1c4e5a25bfc7-3/kafka:broker-storage:VolumeSize/some-autoscaler-name",
			expected:      []string{"kafka", "arn:aws:kafka:us-west-2:123456789012:cluster/example/279c0212-d057-4dba-9aa9-1c4e5a25bfc7-3", "kafka:broker-storage:VolumeSize", "some-autoscaler-name"},
			errorExpected: false,
		},
		{
			input:         "kafka/missing/parts",
			errorExpected: true,
		},
	}

	for _, tc := range testCases {
		idParts, err := tfappautoscaling.ValidPolicyImportInput(tc.input)
		if tc.errorExpected == false && err != nil {
			t.Errorf("tfappautoscaling.ValidPolicyImportInput(%q): resulted in an unexpected error: %s", tc.input, err)
		}

		if tc.errorExpected == true && err == nil {
			t.Errorf("tfappautoscaling.ValidPolicyImportInput(%q): expected an error, but returned successfully", tc.input)
		}

		if !reflect.DeepEqual(tc.expected, idParts) {
			t.Errorf("tfappautoscaling.ValidPolicyImportInput(%q): expected %q, but got %q", tc.input, strings.Join(tc.expected, "/"), strings.Join(idParts, "/"))
		}
	}
}

func TestAccAppAutoScalingPolicy_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var policy awstypes.ScalingPolicy
	appAutoscalingTargetResourceName := "aws_appautoscaling_target.test"
	resourceName := "aws_appautoscaling_policy.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.AppAutoScalingServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckPolicyDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccPolicyConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPolicyExists(ctx, resourceName, &policy),
					resource.TestCheckResourceAttr(resourceName, "alarm_arns.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, "policy_type", "StepScaling"),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrResourceID, appAutoscalingTargetResourceName, names.AttrResourceID),
					resource.TestCheckResourceAttrPair(resourceName, "scalable_dimension", appAutoscalingTargetResourceName, "scalable_dimension"),
					resource.TestCheckResourceAttrPair(resourceName, "service_namespace", appAutoscalingTargetResourceName, "service_namespace"),
					resource.TestCheckResourceAttr(resourceName, "step_scaling_policy_configuration.0.adjustment_type", "ChangeInCapacity"),
					resource.TestCheckResourceAttr(resourceName, "step_scaling_policy_configuration.0.cooldown", "60"),
					resource.TestCheckResourceAttr(resourceName, "step_scaling_policy_configuration.0.step_adjustment.#", acctest.Ct1),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "step_scaling_policy_configuration.0.step_adjustment.*", map[string]string{
						"scaling_adjustment":          acctest.Ct1,
						"metric_interval_lower_bound": acctest.Ct0,
						"metric_interval_upper_bound": "",
					}),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateIdFunc: testAccPolicyImportStateIdFunc(resourceName),
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccAppAutoScalingPolicy_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	var policy awstypes.ScalingPolicy
	resourceName := "aws_appautoscaling_policy.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.AppAutoScalingServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckPolicyDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccPolicyConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPolicyExists(ctx, resourceName, &policy),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfappautoscaling.ResourcePolicy(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccAppAutoScalingPolicy_scaleOutAndIn(t *testing.T) {
	ctx := acctest.Context(t)
	var policy awstypes.ScalingPolicy

	randClusterName := fmt.Sprintf("cluster%s", sdkacctest.RandString(10))
	randPolicyNamePrefix := fmt.Sprintf("terraform-test-foobar-%s", sdkacctest.RandString(5))

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.AppAutoScalingServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckPolicyDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccPolicyConfig_scaleOutAndIn(randClusterName, randPolicyNamePrefix),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPolicyExists(ctx, "aws_appautoscaling_policy.foobar_out", &policy),
					resource.TestCheckResourceAttr("aws_appautoscaling_policy.foobar_out", "step_scaling_policy_configuration.0.adjustment_type", "PercentChangeInCapacity"),
					resource.TestCheckResourceAttr("aws_appautoscaling_policy.foobar_out", "step_scaling_policy_configuration.0.cooldown", "60"),
					resource.TestCheckResourceAttr("aws_appautoscaling_policy.foobar_out", "step_scaling_policy_configuration.0.step_adjustment.#", acctest.Ct3),
					resource.TestCheckTypeSetElemNestedAttrs("aws_appautoscaling_policy.foobar_out", "step_scaling_policy_configuration.0.step_adjustment.*", map[string]string{
						"metric_interval_lower_bound": acctest.Ct3,
						"metric_interval_upper_bound": "",
						"scaling_adjustment":          acctest.Ct3,
					}),
					resource.TestCheckTypeSetElemNestedAttrs("aws_appautoscaling_policy.foobar_out", "step_scaling_policy_configuration.0.step_adjustment.*", map[string]string{
						"metric_interval_lower_bound": acctest.Ct1,
						"metric_interval_upper_bound": acctest.Ct3,
						"scaling_adjustment":          acctest.Ct2,
					}),
					resource.TestCheckTypeSetElemNestedAttrs("aws_appautoscaling_policy.foobar_out", "step_scaling_policy_configuration.0.step_adjustment.*", map[string]string{
						"metric_interval_lower_bound": acctest.Ct0,
						"metric_interval_upper_bound": acctest.Ct1,
						"scaling_adjustment":          acctest.Ct1,
					}),
					resource.TestCheckResourceAttr("aws_appautoscaling_policy.foobar_out", names.AttrName, fmt.Sprintf("%s-out", randPolicyNamePrefix)),
					resource.TestCheckResourceAttr("aws_appautoscaling_policy.foobar_out", "policy_type", "StepScaling"),
					resource.TestCheckResourceAttr("aws_appautoscaling_policy.foobar_out", names.AttrResourceID, fmt.Sprintf("service/%s/foobar", randClusterName)),
					resource.TestCheckResourceAttr("aws_appautoscaling_policy.foobar_out", "service_namespace", "ecs"),
					resource.TestCheckResourceAttr("aws_appautoscaling_policy.foobar_out", "scalable_dimension", "ecs:service:DesiredCount"),
					testAccCheckPolicyExists(ctx, "aws_appautoscaling_policy.foobar_in", &policy),
					resource.TestCheckResourceAttr("aws_appautoscaling_policy.foobar_in", "step_scaling_policy_configuration.0.adjustment_type", "PercentChangeInCapacity"),
					resource.TestCheckResourceAttr("aws_appautoscaling_policy.foobar_in", "step_scaling_policy_configuration.0.cooldown", "60"),
					resource.TestCheckResourceAttr("aws_appautoscaling_policy.foobar_in", "step_scaling_policy_configuration.0.step_adjustment.#", acctest.Ct3),
					resource.TestCheckTypeSetElemNestedAttrs("aws_appautoscaling_policy.foobar_in", "step_scaling_policy_configuration.0.step_adjustment.*", map[string]string{
						"metric_interval_lower_bound": "-1",
						"metric_interval_upper_bound": acctest.Ct0,
						"scaling_adjustment":          "-1",
					}),
					resource.TestCheckTypeSetElemNestedAttrs("aws_appautoscaling_policy.foobar_in", "step_scaling_policy_configuration.0.step_adjustment.*", map[string]string{
						"metric_interval_lower_bound": "-3",
						"metric_interval_upper_bound": "-1",
						"scaling_adjustment":          "-2",
					}),
					resource.TestCheckTypeSetElemNestedAttrs("aws_appautoscaling_policy.foobar_in", "step_scaling_policy_configuration.0.step_adjustment.*", map[string]string{
						"metric_interval_lower_bound": "",
						"metric_interval_upper_bound": "-3",
						"scaling_adjustment":          "-3",
					}),
					resource.TestCheckResourceAttr("aws_appautoscaling_policy.foobar_in", names.AttrName, fmt.Sprintf("%s-in", randPolicyNamePrefix)),
					resource.TestCheckResourceAttr("aws_appautoscaling_policy.foobar_in", "policy_type", "StepScaling"),
					resource.TestCheckResourceAttr("aws_appautoscaling_policy.foobar_in", names.AttrResourceID, fmt.Sprintf("service/%s/foobar", randClusterName)),
					resource.TestCheckResourceAttr("aws_appautoscaling_policy.foobar_in", "service_namespace", "ecs"),
					resource.TestCheckResourceAttr("aws_appautoscaling_policy.foobar_in", "scalable_dimension", "ecs:service:DesiredCount"),
				),
			},
			{
				ResourceName:      "aws_appautoscaling_policy.foobar_out",
				ImportState:       true,
				ImportStateIdFunc: testAccPolicyImportStateIdFunc("aws_appautoscaling_policy.foobar_out"),
				ImportStateVerify: true,
			},
			{
				ResourceName:      "aws_appautoscaling_policy.foobar_in",
				ImportState:       true,
				ImportStateIdFunc: testAccPolicyImportStateIdFunc("aws_appautoscaling_policy.foobar_in"),
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccAppAutoScalingPolicy_spotFleetRequest(t *testing.T) {
	ctx := acctest.Context(t)
	var policy awstypes.ScalingPolicy

	randPolicyName := fmt.Sprintf("test-appautoscaling-policy-%s", sdkacctest.RandString(5))
	validUntil := time.Now().UTC().Add(24 * time.Hour).Format(time.RFC3339)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.AppAutoScalingServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckPolicyDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccPolicyConfig_spotFleetRequest(randPolicyName, validUntil),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPolicyExists(ctx, "aws_appautoscaling_policy.test", &policy),
					resource.TestCheckResourceAttr("aws_appautoscaling_policy.test", names.AttrName, randPolicyName),
					resource.TestCheckResourceAttr("aws_appautoscaling_policy.test", "service_namespace", "ec2"),
					resource.TestCheckResourceAttr("aws_appautoscaling_policy.test", "scalable_dimension", "ec2:spot-fleet-request:TargetCapacity"),
				),
			},
			{
				ResourceName:      "aws_appautoscaling_policy.test",
				ImportState:       true,
				ImportStateIdFunc: testAccPolicyImportStateIdFunc("aws_appautoscaling_policy.test"),
				ImportStateVerify: true,
			},
		},
	})
}

// TODO: Add test for CustomizedMetricSpecification
// The field doesn't seem to be accessible for common AWS customers (yet?)
func TestAccAppAutoScalingPolicy_DynamoDB_table(t *testing.T) {
	ctx := acctest.Context(t)
	var policy awstypes.ScalingPolicy
	resourceName := "aws_appautoscaling_policy.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.AppAutoScalingServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckPolicyDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccPolicyConfig_dynamoDB(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPolicyExists(ctx, resourceName, &policy),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, fmt.Sprintf("DynamoDBWriteCapacityUtilization:table/%s", rName)),
					resource.TestCheckResourceAttr(resourceName, "policy_type", "TargetTrackingScaling"),
					resource.TestCheckResourceAttr(resourceName, "service_namespace", "dynamodb"),
					resource.TestCheckResourceAttr(resourceName, "scalable_dimension", "dynamodb:table:WriteCapacityUnits"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateIdFunc: testAccPolicyImportStateIdFunc(resourceName),
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccAppAutoScalingPolicy_DynamoDB_index(t *testing.T) {
	ctx := acctest.Context(t)
	var policy awstypes.ScalingPolicy
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	appautoscalingTargetResourceName := "aws_appautoscaling_target.test"
	resourceName := "aws_appautoscaling_policy.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.AppAutoScalingServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckPolicyDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccPolicyConfig_dynamoDBIndex(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPolicyExists(ctx, resourceName, &policy),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, fmt.Sprintf("DynamoDBWriteCapacityUtilization:table/%s/index/GameTitleIndex", rName)),
					resource.TestCheckResourceAttr(resourceName, "policy_type", "TargetTrackingScaling"),
					resource.TestCheckResourceAttrPair(resourceName, "service_namespace", appautoscalingTargetResourceName, "service_namespace"),
					resource.TestCheckResourceAttrPair(resourceName, "scalable_dimension", appautoscalingTargetResourceName, "scalable_dimension"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateIdFunc: testAccPolicyImportStateIdFunc(resourceName),
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccAppAutoScalingPolicy_multiplePoliciesSameName(t *testing.T) {
	ctx := acctest.Context(t)
	var readPolicy1 awstypes.ScalingPolicy
	var readPolicy2 awstypes.ScalingPolicy

	tableName1 := fmt.Sprintf("tf-autoscaled-table-%s", sdkacctest.RandString(5))
	tableName2 := fmt.Sprintf("tf-autoscaled-table-%s", sdkacctest.RandString(5))
	namePrefix := fmt.Sprintf("tf-appautoscaling-policy-%s", sdkacctest.RandString(5))

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.AppAutoScalingServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckPolicyDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccPolicyConfig_multiplePoliciesSameName(tableName1, tableName2, namePrefix),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPolicyExists(ctx, "aws_appautoscaling_policy.read1", &readPolicy1),
					resource.TestCheckResourceAttr("aws_appautoscaling_policy.read1", names.AttrName, namePrefix+"-read"),
					resource.TestCheckResourceAttr("aws_appautoscaling_policy.read1", names.AttrResourceID, "table/"+tableName1),
					resource.TestCheckResourceAttr("aws_appautoscaling_policy.read1", "service_namespace", "dynamodb"),
					resource.TestCheckResourceAttr("aws_appautoscaling_policy.read1", "scalable_dimension", "dynamodb:table:ReadCapacityUnits"),

					testAccCheckPolicyExists(ctx, "aws_appautoscaling_policy.read2", &readPolicy2),
					resource.TestCheckResourceAttr("aws_appautoscaling_policy.read2", names.AttrName, namePrefix+"-read"),
					resource.TestCheckResourceAttr("aws_appautoscaling_policy.read2", names.AttrResourceID, "table/"+tableName2),
					resource.TestCheckResourceAttr("aws_appautoscaling_policy.read2", "service_namespace", "dynamodb"),
					resource.TestCheckResourceAttr("aws_appautoscaling_policy.read2", "scalable_dimension", "dynamodb:table:ReadCapacityUnits"),
				),
			},
		},
	})
}

func TestAccAppAutoScalingPolicy_multiplePoliciesSameResource(t *testing.T) {
	ctx := acctest.Context(t)
	var readPolicy awstypes.ScalingPolicy
	var writePolicy awstypes.ScalingPolicy

	tableName := fmt.Sprintf("tf-autoscaled-table-%s", sdkacctest.RandString(5))
	namePrefix := fmt.Sprintf("tf-appautoscaling-policy-%s", sdkacctest.RandString(5))

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.AppAutoScalingServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckPolicyDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccPolicyConfig_multiplePoliciesSameResource(tableName, namePrefix),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPolicyExists(ctx, "aws_appautoscaling_policy.read", &readPolicy),
					resource.TestCheckResourceAttr("aws_appautoscaling_policy.read", names.AttrName, namePrefix+"-read"),
					resource.TestCheckResourceAttr("aws_appautoscaling_policy.read", names.AttrResourceID, "table/"+tableName),
					resource.TestCheckResourceAttr("aws_appautoscaling_policy.read", "service_namespace", "dynamodb"),
					resource.TestCheckResourceAttr("aws_appautoscaling_policy.read", "scalable_dimension", "dynamodb:table:ReadCapacityUnits"),

					testAccCheckPolicyExists(ctx, "aws_appautoscaling_policy.write", &writePolicy),
					resource.TestCheckResourceAttr("aws_appautoscaling_policy.write", names.AttrName, namePrefix+"-write"),
					resource.TestCheckResourceAttr("aws_appautoscaling_policy.write", names.AttrResourceID, "table/"+tableName),
					resource.TestCheckResourceAttr("aws_appautoscaling_policy.write", "service_namespace", "dynamodb"),
					resource.TestCheckResourceAttr("aws_appautoscaling_policy.write", "scalable_dimension", "dynamodb:table:WriteCapacityUnits"),
				),
			},
			{
				ResourceName:      "aws_appautoscaling_policy.read",
				ImportState:       true,
				ImportStateIdFunc: testAccPolicyImportStateIdFunc("aws_appautoscaling_policy.read"),
				ImportStateVerify: true,
			},
			{
				ResourceName:      "aws_appautoscaling_policy.write",
				ImportState:       true,
				ImportStateIdFunc: testAccPolicyImportStateIdFunc("aws_appautoscaling_policy.write"),
				ImportStateVerify: true,
			},
		},
	})
}

// Reference: https://github.com/hashicorp/terraform-provider-aws/issues/7963
func TestAccAppAutoScalingPolicy_ResourceID_forceNew(t *testing.T) {
	ctx := acctest.Context(t)
	var policy awstypes.ScalingPolicy
	appAutoscalingTargetResourceName := "aws_appautoscaling_target.test"
	resourceName := "aws_appautoscaling_policy.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.AppAutoScalingServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckPolicyDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccPolicyConfig_resourceIDForceNew1(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPolicyExists(ctx, resourceName, &policy),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrResourceID, appAutoscalingTargetResourceName, names.AttrResourceID),
					resource.TestCheckResourceAttrPair(resourceName, "scalable_dimension", appAutoscalingTargetResourceName, "scalable_dimension"),
					resource.TestCheckResourceAttrPair(resourceName, "service_namespace", appAutoscalingTargetResourceName, "service_namespace"),
				),
			},
			{
				Config: testAccPolicyConfig_resourceIDForceNew2(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPolicyExists(ctx, resourceName, &policy),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrResourceID, appAutoscalingTargetResourceName, names.AttrResourceID),
					resource.TestCheckResourceAttrPair(resourceName, "scalable_dimension", appAutoscalingTargetResourceName, "scalable_dimension"),
					resource.TestCheckResourceAttrPair(resourceName, "service_namespace", appAutoscalingTargetResourceName, "service_namespace"),
				),
			},
		},
	})
}

func TestAccAppAutoScalingPolicy_TargetTrack_metricMath(t *testing.T) {
	ctx := acctest.Context(t)
	var policy awstypes.ScalingPolicy
	appAutoscalingTargetResourceName := "aws_appautoscaling_target.test"
	resourceName := "aws_appautoscaling_policy.metric_math_test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.AppAutoScalingServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckPolicyDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccPolicyConfig_targetTrackingMetricMath(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPolicyExists(ctx, resourceName, &policy),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrResourceID, appAutoscalingTargetResourceName, names.AttrResourceID),
					resource.TestCheckResourceAttrPair(resourceName, "scalable_dimension", appAutoscalingTargetResourceName, "scalable_dimension"),
					resource.TestCheckResourceAttrPair(resourceName, "service_namespace", appAutoscalingTargetResourceName, "service_namespace"),
					resource.TestCheckResourceAttr(resourceName, "target_tracking_scaling_policy_configuration.0.target_value", "12.3"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateIdFunc: testAccPolicyImportStateIdFunc(resourceName),
				ImportStateVerify: true,
			},
		},
	})
}

func testAccPolicyConfig_targetTrackingMetricMath(rName string) string {
	return acctest.ConfigCompose(testAccPolicyConfig_basic(rName), fmt.Sprintf(`
resource "aws_appautoscaling_policy" "metric_math_test" {
  name               = "%[1]s-tracking"
  policy_type        = "TargetTrackingScaling"
  resource_id        = aws_appautoscaling_target.test.resource_id
  scalable_dimension = aws_appautoscaling_target.test.scalable_dimension
  service_namespace  = aws_appautoscaling_target.test.service_namespace

  target_tracking_scaling_policy_configuration {
    customized_metric_specification {
      metrics {
        id          = "m1"
        expression  = "TIME_SERIES(20)"
        return_data = false
      }
      metrics {
        id = "m2"
        metric_stat {
          metric {
            namespace   = "foo"
            metric_name = "bar"
          }
          unit = "Percent"
          stat = "Sum"
        }
        return_data = false
      }
      metrics {
        id = "m3"
        metric_stat {
          metric {
            namespace   = "foo"
            metric_name = "bar"
            dimensions {
              name  = "x"
              value = "y"
            }
            dimensions {
              name  = "y"
              value = "x"
            }
          }
          unit = "Percent"
          stat = "Sum"
        }
        return_data = false
      }
      metrics {
        id          = "e1"
        expression  = "m1 + m2 + m3"
        return_data = true
      }
    }
    target_value = 12.3
  }
}
`, rName))
}

func testAccCheckPolicyExists(ctx context.Context, n string, v *awstypes.ScalingPolicy) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).AppAutoScalingClient(ctx)

		output, err := tfappautoscaling.FindScalingPolicyByFourPartKey(ctx, conn, rs.Primary.Attributes[names.AttrName], rs.Primary.Attributes["service_namespace"], rs.Primary.Attributes[names.AttrResourceID], rs.Primary.Attributes["scalable_dimension"])

		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

func testAccCheckPolicyDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).AppAutoScalingClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_appautoscaling_policy" {
				continue
			}

			_, err := tfappautoscaling.FindScalingPolicyByFourPartKey(ctx, conn, rs.Primary.Attributes[names.AttrName], rs.Primary.Attributes["service_namespace"], rs.Primary.Attributes[names.AttrResourceID], rs.Primary.Attributes["scalable_dimension"])

			if tfresource.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("Application Auto Scaling Scaling Policy %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccPolicyConfig_basic(rName string) string {
	return fmt.Sprintf(`
resource "aws_ecs_cluster" "test" {
  name = %[1]q
}

resource "aws_ecs_task_definition" "test" {
  family = %[1]q

  container_definitions = <<EOF
[
  {
    "name": "busybox",
    "image": "busybox:latest",
    "cpu": 10,
    "memory": 128,
    "essential": true
  }
]
EOF
}

resource "aws_ecs_service" "test" {
  cluster                            = aws_ecs_cluster.test.id
  deployment_maximum_percent         = 200
  deployment_minimum_healthy_percent = 50
  desired_count                      = 0
  name                               = %[1]q
  task_definition                    = aws_ecs_task_definition.test.arn
}

resource "aws_appautoscaling_target" "test" {
  max_capacity       = 4
  min_capacity       = 0
  resource_id        = "service/${aws_ecs_cluster.test.name}/${aws_ecs_service.test.name}"
  scalable_dimension = "ecs:service:DesiredCount"
  service_namespace  = "ecs"
}

resource "aws_appautoscaling_policy" "test" {
  name               = %[1]q
  resource_id        = aws_appautoscaling_target.test.resource_id
  scalable_dimension = aws_appautoscaling_target.test.scalable_dimension
  service_namespace  = aws_appautoscaling_target.test.service_namespace

  step_scaling_policy_configuration {
    adjustment_type         = "ChangeInCapacity"
    cooldown                = 60
    metric_aggregation_type = "Average"

    step_adjustment {
      metric_interval_lower_bound = 0
      scaling_adjustment          = 1
    }
  }
}
`, rName)
}

func testAccPolicyConfig_spotFleetRequest(rName, validUntil string) string {
	return acctest.ConfigCompose(acctest.ConfigLatestAmazonLinux2HVMEBSX8664AMI(), fmt.Sprintf(`
data "aws_partition" "current" {}

resource "aws_iam_role" "test" {
  assume_role_policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Principal": {
        "Service": [
          "spotfleet.${data.aws_partition.current.dns_suffix}",
          "ec2.${data.aws_partition.current.dns_suffix}"
        ]
      },
      "Action": "sts:AssumeRole"
    }
  ]
}
EOF
}

resource "aws_iam_role_policy_attachment" "test" {
  role       = aws_iam_role.test.name
  policy_arn = "arn:${data.aws_partition.current.partition}:iam::aws:policy/service-role/AmazonEC2SpotFleetTaggingRole"
}

resource "aws_spot_fleet_request" "test" {
  iam_fleet_role                      = aws_iam_role.test.arn
  spot_price                          = "0.005"
  target_capacity                     = 2
  valid_until                         = %[2]q
  terminate_instances_with_expiration = true

  launch_specification {
    instance_type = "m3.medium"
    ami           = data.aws_ami.amzn2-ami-minimal-hvm-ebs-x86_64.id
  }
}

resource "aws_appautoscaling_target" "test" {
  service_namespace  = "ec2"
  resource_id        = "spot-fleet-request/${aws_spot_fleet_request.test.id}"
  scalable_dimension = "ec2:spot-fleet-request:TargetCapacity"
  min_capacity       = 1
  max_capacity       = 3
}

resource "aws_appautoscaling_policy" "test" {
  name               = %[1]q
  resource_id        = aws_appautoscaling_target.test.resource_id
  scalable_dimension = aws_appautoscaling_target.test.scalable_dimension
  service_namespace  = aws_appautoscaling_target.test.service_namespace

  step_scaling_policy_configuration {
    adjustment_type         = "ChangeInCapacity"
    cooldown                = 60
    metric_aggregation_type = "Average"

    step_adjustment {
      metric_interval_lower_bound = 0
      scaling_adjustment          = 1
    }
  }
}
`, rName, validUntil))
}

func testAccPolicyConfig_dynamoDB(rName string) string {
	return fmt.Sprintf(`
resource "aws_dynamodb_table" "test" {
  name           = %[1]q
  read_capacity  = 5
  write_capacity = 5
  hash_key       = "FooKey"

  attribute {
    name = "FooKey"
    type = "S"
  }
}

resource "aws_appautoscaling_target" "test" {
  service_namespace  = "dynamodb"
  resource_id        = "table/${aws_dynamodb_table.test.name}"
  scalable_dimension = "dynamodb:table:WriteCapacityUnits"
  min_capacity       = 1
  max_capacity       = 10
}

resource "aws_appautoscaling_policy" "test" {
  name               = "DynamoDBWriteCapacityUtilization:${aws_appautoscaling_target.test.resource_id}"
  policy_type        = "TargetTrackingScaling"
  service_namespace  = "dynamodb"
  resource_id        = "table/${aws_dynamodb_table.test.name}"
  scalable_dimension = "dynamodb:table:WriteCapacityUnits"

  target_tracking_scaling_policy_configuration {
    predefined_metric_specification {
      predefined_metric_type = "DynamoDBWriteCapacityUtilization"
    }

    scale_in_cooldown  = 10
    scale_out_cooldown = 10
    target_value       = 70
  }

  depends_on = [aws_appautoscaling_target.test]
}
`, rName)
}

func testAccPolicyConfig_dynamoDBIndex(rName string) string {
	return fmt.Sprintf(`
resource "aws_dynamodb_table" "test" {
  name           = %[1]q
  read_capacity  = 1
  write_capacity = 1
  hash_key       = "UserId"
  range_key      = "GameTitle"

  attribute {
    name = "UserId"
    type = "S"
  }

  attribute {
    name = "GameTitle"
    type = "S"
  }

  attribute {
    name = "TopScore"
    type = "N"
  }

  global_secondary_index {
    name               = "GameTitleIndex"
    hash_key           = "GameTitle"
    range_key          = "TopScore"
    write_capacity     = 1
    read_capacity      = 1
    projection_type    = "INCLUDE"
    non_key_attributes = ["UserId"]
  }
}

resource "aws_appautoscaling_target" "test" {
  service_namespace  = "dynamodb"
  resource_id        = "table/${aws_dynamodb_table.test.name}/index/GameTitleIndex"
  scalable_dimension = "dynamodb:index:WriteCapacityUnits"
  min_capacity       = 1
  max_capacity       = 10
}

resource "aws_appautoscaling_policy" "test" {
  name               = "DynamoDBWriteCapacityUtilization:${aws_appautoscaling_target.test.resource_id}"
  policy_type        = "TargetTrackingScaling"
  service_namespace  = aws_appautoscaling_target.test.service_namespace
  resource_id        = aws_appautoscaling_target.test.resource_id
  scalable_dimension = aws_appautoscaling_target.test.scalable_dimension

  target_tracking_scaling_policy_configuration {
    predefined_metric_specification {
      predefined_metric_type = "DynamoDBWriteCapacityUtilization"
    }

    scale_in_cooldown  = 10
    scale_out_cooldown = 10
    target_value       = 70
  }
}
`, rName)
}

func testAccPolicyConfig_multiplePoliciesSameName(tableName1, tableName2, namePrefix string) string {
	return fmt.Sprintf(`
resource "aws_dynamodb_table" "dynamodb_table_test1" {
  name           = %[1]q
  read_capacity  = 1
  write_capacity = 1
  hash_key       = "FooKey"

  attribute {
    name = "FooKey"
    type = "S"
  }
}

resource "aws_dynamodb_table" "dynamodb_table_test2" {
  name           = %[2]q
  read_capacity  = 1
  write_capacity = 1
  hash_key       = "FooKey"

  attribute {
    name = "FooKey"
    type = "S"
  }
}

resource "aws_appautoscaling_target" "read1" {
  service_namespace  = "dynamodb"
  resource_id        = "table/${aws_dynamodb_table.dynamodb_table_test1.name}"
  scalable_dimension = "dynamodb:table:ReadCapacityUnits"
  min_capacity       = 1
  max_capacity       = 10
}

resource "aws_appautoscaling_policy" "read1" {
  name               = "%[3]s-read"
  policy_type        = "TargetTrackingScaling"
  service_namespace  = "dynamodb"
  resource_id        = aws_appautoscaling_target.read1.resource_id
  scalable_dimension = aws_appautoscaling_target.read1.scalable_dimension

  target_tracking_scaling_policy_configuration {
    predefined_metric_specification {
      predefined_metric_type = "DynamoDBReadCapacityUtilization"
    }

    scale_in_cooldown  = 10
    scale_out_cooldown = 10
    target_value       = 70
  }
}

resource "aws_appautoscaling_target" "read2" {
  service_namespace  = "dynamodb"
  resource_id        = "table/${aws_dynamodb_table.dynamodb_table_test2.name}"
  scalable_dimension = "dynamodb:table:ReadCapacityUnits"
  min_capacity       = 1
  max_capacity       = 10
}

resource "aws_appautoscaling_policy" "read2" {
  name               = "%[3]s-read"
  policy_type        = "TargetTrackingScaling"
  service_namespace  = "dynamodb"
  resource_id        = "table/${aws_dynamodb_table.dynamodb_table_test2.name}"
  scalable_dimension = aws_appautoscaling_target.read2.scalable_dimension

  target_tracking_scaling_policy_configuration {
    predefined_metric_specification {
      predefined_metric_type = "DynamoDBReadCapacityUtilization"
    }

    scale_in_cooldown  = 10
    scale_out_cooldown = 10
    target_value       = 70
  }
}
`, tableName1, tableName2, namePrefix)
}

func testAccPolicyConfig_multiplePoliciesSameResource(tableName, namePrefix string) string {
	return fmt.Sprintf(`
resource "aws_dynamodb_table" "dynamodb_table_test" {
  name           = %[1]q
  read_capacity  = 5
  write_capacity = 5
  hash_key       = "FooKey"

  attribute {
    name = "FooKey"
    type = "S"
  }
}

resource "aws_appautoscaling_target" "write" {
  service_namespace  = "dynamodb"
  resource_id        = "table/${aws_dynamodb_table.dynamodb_table_test.name}"
  scalable_dimension = "dynamodb:table:WriteCapacityUnits"
  min_capacity       = 1
  max_capacity       = 10
}

resource "aws_appautoscaling_policy" "write" {
  name               = "%s-write"
  policy_type        = "TargetTrackingScaling"
  service_namespace  = "dynamodb"
  resource_id        = "table/${aws_dynamodb_table.dynamodb_table_test.name}"
  scalable_dimension = "dynamodb:table:WriteCapacityUnits"

  target_tracking_scaling_policy_configuration {
    predefined_metric_specification {
      predefined_metric_type = "DynamoDBWriteCapacityUtilization"
    }

    scale_in_cooldown  = 10
    scale_out_cooldown = 10
    target_value       = 70
  }

  depends_on = [aws_appautoscaling_target.write]
}

resource "aws_appautoscaling_target" "read" {
  service_namespace  = "dynamodb"
  resource_id        = "table/${aws_dynamodb_table.dynamodb_table_test.name}"
  scalable_dimension = "dynamodb:table:ReadCapacityUnits"
  min_capacity       = 1
  max_capacity       = 10
}

resource "aws_appautoscaling_policy" "read" {
  name               = "%s-read"
  policy_type        = "TargetTrackingScaling"
  service_namespace  = "dynamodb"
  resource_id        = "table/${aws_dynamodb_table.dynamodb_table_test.name}"
  scalable_dimension = "dynamodb:table:ReadCapacityUnits"

  target_tracking_scaling_policy_configuration {
    predefined_metric_specification {
      predefined_metric_type = "DynamoDBReadCapacityUtilization"
    }

    scale_in_cooldown  = 10
    scale_out_cooldown = 10
    target_value       = 70
  }

  depends_on = [aws_appautoscaling_target.read]
}
`, tableName, namePrefix, namePrefix)
}

func testAccPolicyConfig_scaleOutAndIn(
	randClusterName string,
	randPolicyNamePrefix string) string {
	return fmt.Sprintf(`
resource "aws_ecs_cluster" "foo" {
  name = "%s"
}

resource "aws_ecs_task_definition" "task" {
  family = "foobar"

  container_definitions = <<EOF
[
	{
		"name": "busybox",
		"image": "busybox:latest",
		"cpu": 10,
		"memory": 128,
		"essential": true
	}
]
EOF
}

resource "aws_ecs_service" "service" {
  name                               = "foobar"
  cluster                            = aws_ecs_cluster.foo.id
  task_definition                    = aws_ecs_task_definition.task.arn
  desired_count                      = 1
  deployment_maximum_percent         = 200
  deployment_minimum_healthy_percent = 50
}

resource "aws_appautoscaling_target" "tgt" {
  service_namespace  = "ecs"
  resource_id        = "service/${aws_ecs_cluster.foo.name}/${aws_ecs_service.service.name}"
  scalable_dimension = "ecs:service:DesiredCount"
  min_capacity       = 1
  max_capacity       = 4
}

resource "aws_appautoscaling_policy" "foobar_out" {
  name               = "%s-out"
  service_namespace  = "ecs"
  resource_id        = "service/${aws_ecs_cluster.foo.name}/${aws_ecs_service.service.name}"
  scalable_dimension = "ecs:service:DesiredCount"

  step_scaling_policy_configuration {
    adjustment_type         = "PercentChangeInCapacity"
    cooldown                = 60
    metric_aggregation_type = "Average"

    step_adjustment {
      metric_interval_lower_bound = 3
      scaling_adjustment          = 3
    }

    step_adjustment {
      metric_interval_upper_bound = 3
      metric_interval_lower_bound = 1
      scaling_adjustment          = 2
    }

    step_adjustment {
      metric_interval_upper_bound = 1
      metric_interval_lower_bound = 0
      scaling_adjustment          = 1
    }
  }

  depends_on = [aws_appautoscaling_target.tgt]
}

resource "aws_appautoscaling_policy" "foobar_in" {
  name               = "%s-in"
  service_namespace  = "ecs"
  resource_id        = "service/${aws_ecs_cluster.foo.name}/${aws_ecs_service.service.name}"
  scalable_dimension = "ecs:service:DesiredCount"

  step_scaling_policy_configuration {
    adjustment_type         = "PercentChangeInCapacity"
    cooldown                = 60
    metric_aggregation_type = "Average"

    step_adjustment {
      metric_interval_upper_bound = 0
      metric_interval_lower_bound = -1
      scaling_adjustment          = -1
    }

    step_adjustment {
      metric_interval_upper_bound = -1
      metric_interval_lower_bound = -3
      scaling_adjustment          = -2
    }

    step_adjustment {
      metric_interval_upper_bound = -3
      scaling_adjustment          = -3
    }
  }

  depends_on = [aws_appautoscaling_target.tgt]
}
`, randClusterName, randPolicyNamePrefix, randPolicyNamePrefix)
}

func testAccPolicyConfig_resourceIDForceNewBase(rName string) string {
	return fmt.Sprintf(`
resource "aws_ecs_cluster" "test" {
  name = %[1]q
}

resource "aws_ecs_task_definition" "test" {
  family = %[1]q

  container_definitions = <<EOF
[
  {
    "name": "busybox",
    "image": "busybox:latest",
    "cpu": 10,
    "memory": 128,
    "essential": true
  }
]
EOF
}

resource "aws_ecs_service" "test1" {
  cluster                            = aws_ecs_cluster.test.id
  deployment_maximum_percent         = 200
  deployment_minimum_healthy_percent = 50
  desired_count                      = 0
  name                               = "%[1]s-1"
  task_definition                    = aws_ecs_task_definition.test.arn
}

resource "aws_ecs_service" "test2" {
  cluster                            = aws_ecs_cluster.test.id
  deployment_maximum_percent         = 200
  deployment_minimum_healthy_percent = 50
  desired_count                      = 0
  name                               = "%[1]s-2"
  task_definition                    = aws_ecs_task_definition.test.arn
}
`, rName)
}

func testAccPolicyConfig_resourceIDForceNew1(rName string) string {
	return testAccPolicyConfig_resourceIDForceNewBase(rName) + fmt.Sprintf(`
resource "aws_appautoscaling_target" "test" {
  max_capacity       = 4
  min_capacity       = 0
  resource_id        = "service/${aws_ecs_cluster.test.name}/${aws_ecs_service.test1.name}"
  scalable_dimension = "ecs:service:DesiredCount"
  service_namespace  = "ecs"
}

resource "aws_appautoscaling_policy" "test" {
  # The usage of depends_on here is intentional as this used to be a documented example
  depends_on = [aws_appautoscaling_target.test]

  name               = %[1]q
  resource_id        = "service/${aws_ecs_cluster.test.name}/${aws_ecs_service.test1.name}"
  scalable_dimension = "ecs:service:DesiredCount"
  service_namespace  = "ecs"

  step_scaling_policy_configuration {
    adjustment_type         = "ChangeInCapacity"
    cooldown                = 60
    metric_aggregation_type = "Average"

    step_adjustment {
      metric_interval_lower_bound = 0
      scaling_adjustment          = 1
    }
  }
}

resource "aws_cloudwatch_metric_alarm" "test" {
  alarm_actions       = [aws_appautoscaling_policy.test.arn]
  alarm_name          = %[1]q
  comparison_operator = "LessThanOrEqualToThreshold"
  evaluation_periods  = "5"
  metric_name         = "CPUReservation"
  namespace           = "AWS/ECS"
  period              = "60"
  statistic           = "Average"
  threshold           = "0"

  dimensions = {
    ClusterName = aws_ecs_cluster.test.name
  }
}
`, rName)
}

func testAccPolicyConfig_resourceIDForceNew2(rName string) string {
	return testAccPolicyConfig_resourceIDForceNewBase(rName) + fmt.Sprintf(`
resource "aws_appautoscaling_target" "test" {
  max_capacity       = 4
  min_capacity       = 0
  resource_id        = "service/${aws_ecs_cluster.test.name}/${aws_ecs_service.test2.name}"
  scalable_dimension = "ecs:service:DesiredCount"
  service_namespace  = "ecs"
}

resource "aws_appautoscaling_policy" "test" {
  # The usage of depends_on here is intentional as this used to be a documented example
  depends_on = [aws_appautoscaling_target.test]

  name               = %[1]q
  resource_id        = "service/${aws_ecs_cluster.test.name}/${aws_ecs_service.test2.name}"
  scalable_dimension = "ecs:service:DesiredCount"
  service_namespace  = "ecs"

  step_scaling_policy_configuration {
    adjustment_type         = "ChangeInCapacity"
    cooldown                = 60
    metric_aggregation_type = "Average"

    step_adjustment {
      metric_interval_lower_bound = 0
      scaling_adjustment          = 1
    }
  }
}

resource "aws_cloudwatch_metric_alarm" "test" {
  alarm_actions       = [aws_appautoscaling_policy.test.arn]
  alarm_name          = %[1]q
  comparison_operator = "LessThanOrEqualToThreshold"
  evaluation_periods  = "5"
  metric_name         = "CPUReservation"
  namespace           = "AWS/ECS"
  period              = "60"
  statistic           = "Average"
  threshold           = "0"

  dimensions = {
    ClusterName = aws_ecs_cluster.test.name
  }
}
`, rName)
}

func testAccPolicyImportStateIdFunc(resourceName string) resource.ImportStateIdFunc {
	return func(s *terraform.State) (string, error) {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return "", fmt.Errorf("Not found: %s", resourceName)
		}

		id := fmt.Sprintf("%s/%s/%s/%s",
			rs.Primary.Attributes["service_namespace"],
			rs.Primary.Attributes[names.AttrResourceID],
			rs.Primary.Attributes["scalable_dimension"],
			rs.Primary.Attributes[names.AttrName])

		return id, nil
	}
}
