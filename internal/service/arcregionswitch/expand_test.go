// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package arcregionswitch

import (
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	awstypes "github.com/aws/aws-sdk-go-v2/service/arcregionswitch/types"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/assert"
)

func TestExpandAssociatedAlarmsFromFramework(t *testing.T) {
	tests := []struct {
		name     string
		input    []associatedAlarmModel
		expected map[string]awstypes.AssociatedAlarm
	}{
		{
			name:     "empty input",
			input:    []associatedAlarmModel{},
			expected: nil,
		},
		{
			name: "single alarm with all fields",
			input: []associatedAlarmModel{
				{
					Name:               types.StringValue("test-alarm"),
					AlarmType:          types.StringValue("applicationHealth"),
					ResourceIdentifier: types.StringValue("test-resource"),
					CrossAccountRole:   types.StringValue("test-role"),
					ExternalId:         types.StringValue("test-external-id"),
				},
			},
			expected: map[string]awstypes.AssociatedAlarm{
				"test-alarm": {
					AlarmType:          awstypes.AlarmTypeApplicationHealth,
					ResourceIdentifier: aws.String("test-resource"),
					CrossAccountRole:   aws.String("test-role"),
					ExternalId:         aws.String("test-external-id"),
				},
			},
		},
		{
			name: "single alarm with required fields only",
			input: []associatedAlarmModel{
				{
					Name:               types.StringValue("test-alarm"),
					AlarmType:          types.StringValue("applicationHealth"),
					ResourceIdentifier: types.StringValue("test-resource"),
					CrossAccountRole:   types.StringNull(),
					ExternalId:         types.StringNull(),
				},
			},
			expected: map[string]awstypes.AssociatedAlarm{
				"test-alarm": {
					AlarmType:          awstypes.AlarmTypeApplicationHealth,
					ResourceIdentifier: aws.String("test-resource"),
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := expandAssociatedAlarmsFromFramework(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestExpandExecutionApprovalConfig(t *testing.T) {
	tests := []struct {
		name     string
		input    stepModel
		expected *awstypes.ExecutionBlockConfigurationMemberExecutionApprovalConfig
	}{
		{
			name: "null config",
			input: stepModel{
				ExecutionApprovalConfig: types.ListNull(getExecutionApprovalConfigObjectType()),
			},
			expected: nil,
		},
		{
			name: "config with timeout",
			input: stepModel{
				ExecutionApprovalConfig: types.ListValueMust(getExecutionApprovalConfigObjectType(), []attr.Value{
					types.ObjectValueMust(getExecutionApprovalConfigObjectType().AttrTypes, map[string]attr.Value{
						"approval_role":   types.StringValue("test-role"),
						"timeout_minutes": types.Int64Value(30),
					}),
				}),
			},
			expected: &awstypes.ExecutionBlockConfigurationMemberExecutionApprovalConfig{
				Value: awstypes.ExecutionApprovalConfiguration{
					ApprovalRole:   aws.String("test-role"),
					TimeoutMinutes: aws.Int32(30),
				},
			},
		},
		{
			name: "config without timeout",
			input: stepModel{
				ExecutionApprovalConfig: types.ListValueMust(getExecutionApprovalConfigObjectType(), []attr.Value{
					types.ObjectValueMust(getExecutionApprovalConfigObjectType().AttrTypes, map[string]attr.Value{
						"approval_role":   types.StringValue("test-role"),
						"timeout_minutes": types.Int64Null(),
					}),
				}),
			},
			expected: &awstypes.ExecutionBlockConfigurationMemberExecutionApprovalConfig{
				Value: awstypes.ExecutionApprovalConfiguration{
					ApprovalRole: aws.String("test-role"),
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := expandExecutionApprovalConfig(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestExpandRoute53HealthCheckConfig(t *testing.T) {
	tests := []struct {
		name     string
		input    stepModel
		expected *awstypes.ExecutionBlockConfigurationMemberRoute53HealthCheckConfig
	}{
		{
			name: "null config",
			input: stepModel{
				Route53HealthCheckConfig: types.ListNull(getRoute53HealthCheckConfigObjectType()),
			},
			expected: nil,
		},
		{
			name: "config with all fields",
			input: stepModel{
				Route53HealthCheckConfig: types.ListValueMust(getRoute53HealthCheckConfigObjectType(), []attr.Value{
					types.ObjectValueMust(getRoute53HealthCheckConfigObjectType().AttrTypes, map[string]attr.Value{
						"hosted_zone_id":     types.StringValue("Z123456789"),
						"record_name":        types.StringValue("test.example.com"),
						"cross_account_role": types.StringValue("test-role"),
						"external_id":        types.StringValue("test-external-id"),
						"timeout_minutes":    types.Int64Value(30),
						"record_sets": types.ListValueMust(getRecordSetObjectType(), []attr.Value{
							types.ObjectValueMust(getRecordSetObjectType().AttrTypes, map[string]attr.Value{
								"record_set_identifier": types.StringValue("primary"),
								"region":                types.StringValue("us-east-1"),
							}),
						}),
					}),
				}),
			},
			expected: &awstypes.ExecutionBlockConfigurationMemberRoute53HealthCheckConfig{
				Value: awstypes.Route53HealthCheckConfiguration{
					HostedZoneId:     aws.String("Z123456789"),
					RecordName:       aws.String("test.example.com"),
					CrossAccountRole: aws.String("test-role"),
					ExternalId:       aws.String("test-external-id"),
					TimeoutMinutes:   aws.Int32(30),
					RecordSets: []awstypes.Route53ResourceRecordSet{
						{
							RecordSetIdentifier: aws.String("primary"),
							Region:              aws.String("us-east-1"),
						},
					},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := expandRoute53HealthCheckConfig(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestExpandCustomActionLambdaConfig(t *testing.T) {
	tests := []struct {
		name     string
		input    stepModel
		expected *awstypes.ExecutionBlockConfigurationMemberCustomActionLambdaConfig
	}{
		{
			name: "null config",
			input: stepModel{
				CustomActionLambdaConfig: types.ListNull(getCustomActionLambdaConfigObjectType()),
			},
			expected: nil,
		},
		{
			name: "config with ungraceful behavior",
			input: stepModel{
				CustomActionLambdaConfig: types.ListValueMust(getCustomActionLambdaConfigObjectType(), []attr.Value{
					types.ObjectValueMust(getCustomActionLambdaConfigObjectType().AttrTypes, map[string]attr.Value{
						"region_to_run":          types.StringValue("activatingRegion"),
						"retry_interval_minutes": types.Float64Value(5.0),
						"timeout_minutes":        types.Int64Value(30),
						"lambda": types.ListValueMust(getLambdaObjectType(), []attr.Value{
							types.ObjectValueMust(getLambdaObjectType().AttrTypes, map[string]attr.Value{
								"arn":                types.StringValue("arn:aws:lambda:us-east-1:123456789012:function:test"),
								"cross_account_role": types.StringValue("test-role"),
								"external_id":        types.StringValue("test-external-id"),
							}),
						}),
						"ungraceful": types.ListValueMust(getUngracefulObjectType(), []attr.Value{
							types.ObjectValueMust(getUngracefulObjectType().AttrTypes, map[string]attr.Value{
								"behavior": types.StringValue("skip"),
							}),
						}),
					}),
				}),
			},
			expected: &awstypes.ExecutionBlockConfigurationMemberCustomActionLambdaConfig{
				Value: awstypes.CustomActionLambdaConfiguration{
					RegionToRun:          awstypes.RegionToRunInActivatingRegion,
					RetryIntervalMinutes: aws.Float32(5.0),
					TimeoutMinutes:       aws.Int32(30),
					Lambdas: []awstypes.Lambdas{
						{
							Arn:              aws.String("arn:aws:lambda:us-east-1:123456789012:function:test"),
							CrossAccountRole: aws.String("test-role"),
							ExternalId:       aws.String("test-external-id"),
						},
					},
					Ungraceful: &awstypes.LambdaUngraceful{
						Behavior: awstypes.LambdaUngracefulBehaviorSkip,
					},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := expandCustomActionLambdaConfig(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestExpandGlobalAuroraConfig(t *testing.T) {
	tests := []struct {
		name     string
		input    stepModel
		expected *awstypes.ExecutionBlockConfigurationMemberGlobalAuroraConfig
	}{
		{
			name: "null config",
			input: stepModel{
				GlobalAuroraConfig: types.ListNull(getGlobalAuroraConfigObjectType()),
			},
			expected: nil,
		},
		{
			name: "config with all fields",
			input: stepModel{
				GlobalAuroraConfig: types.ListValueMust(getGlobalAuroraConfigObjectType(), []attr.Value{
					types.ObjectValueMust(getGlobalAuroraConfigObjectType().AttrTypes, map[string]attr.Value{
						"behavior":                  types.StringValue("switchoverOnly"),
						"global_cluster_identifier": types.StringValue("test-global-cluster"),
						"database_cluster_arns":     types.ListValueMust(types.StringType, []attr.Value{types.StringValue("arn:aws:rds:us-east-1:123456789012:cluster:test-cluster")}),
						"cross_account_role":        types.StringValue("test-role"),
						"external_id":               types.StringValue("test-external-id"),
						"timeout_minutes":           types.Int64Value(30),
						"ungraceful": types.ListValueMust(getGlobalAuroraUngracefulObjectType(), []attr.Value{
							types.ObjectValueMust(getGlobalAuroraUngracefulObjectType().AttrTypes, map[string]attr.Value{
								"ungraceful": types.StringValue("failover"),
							}),
						}),
					}),
				}),
			},
			expected: &awstypes.ExecutionBlockConfigurationMemberGlobalAuroraConfig{
				Value: awstypes.GlobalAuroraConfiguration{
					Behavior:                awstypes.GlobalAuroraDefaultBehaviorSwitchoverOnly,
					GlobalClusterIdentifier: aws.String("test-global-cluster"),
					DatabaseClusterArns:     []string{"arn:aws:rds:us-east-1:123456789012:cluster:test-cluster"},
					CrossAccountRole:        aws.String("test-role"),
					ExternalId:              aws.String("test-external-id"),
					TimeoutMinutes:          aws.Int32(30),
					Ungraceful: &awstypes.GlobalAuroraUngraceful{
						Ungraceful: awstypes.GlobalAuroraUngracefulBehaviorFailover,
					},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := expandGlobalAuroraConfig(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestExpandEc2AsgCapacityIncreaseConfig(t *testing.T) {
	tests := []struct {
		name     string
		input    stepModel
		expected *awstypes.ExecutionBlockConfigurationMemberEc2AsgCapacityIncreaseConfig
	}{
		{
			name: "null config",
			input: stepModel{
				Ec2AsgCapacityIncreaseConfig: types.ListNull(getEc2AsgCapacityIncreaseConfigObjectType()),
			},
			expected: nil,
		},
		{
			name: "config with ungraceful",
			input: stepModel{
				Ec2AsgCapacityIncreaseConfig: types.ListValueMust(getEc2AsgCapacityIncreaseConfigObjectType(), []attr.Value{
					types.ObjectValueMust(getEc2AsgCapacityIncreaseConfigObjectType().AttrTypes, map[string]attr.Value{
						"capacity_monitoring_approach": types.StringValue("sampledMaxInLast24Hours"),
						"target_percent":               types.Int64Value(150),
						"timeout_minutes":              types.Int64Value(30),
						"asgs": types.ListValueMust(getAsgObjectType(), []attr.Value{
							types.ObjectValueMust(getAsgObjectType().AttrTypes, map[string]attr.Value{
								"arn":                types.StringValue("arn:aws:autoscaling:us-east-1:123456789012:autoScalingGroup:test-asg"),
								"cross_account_role": types.StringValue("test-role"),
								"external_id":        types.StringValue("test-external-id"),
							}),
						}),
						"ungraceful": types.ListValueMust(getEc2UngracefulObjectType(), []attr.Value{
							types.ObjectValueMust(getEc2UngracefulObjectType().AttrTypes, map[string]attr.Value{
								"minimum_success_percentage": types.Int64Value(80),
							}),
						}),
					}),
				}),
			},
			expected: &awstypes.ExecutionBlockConfigurationMemberEc2AsgCapacityIncreaseConfig{
				Value: awstypes.Ec2AsgCapacityIncreaseConfiguration{
					CapacityMonitoringApproach: awstypes.Ec2AsgCapacityMonitoringApproachSampledMaxInLast24Hours,
					TargetPercent:              aws.Int32(150),
					TimeoutMinutes:             aws.Int32(30),
					Asgs: []awstypes.Asg{
						{
							Arn:              aws.String("arn:aws:autoscaling:us-east-1:123456789012:autoScalingGroup:test-asg"),
							CrossAccountRole: aws.String("test-role"),
							ExternalId:       aws.String("test-external-id"),
						},
					},
					Ungraceful: &awstypes.Ec2Ungraceful{
						MinimumSuccessPercentage: aws.Int32(80),
					},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := expandEc2AsgCapacityIncreaseConfig(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestExpandEcsCapacityIncreaseConfig(t *testing.T) {
	tests := []struct {
		name     string
		input    stepModel
		expected *awstypes.ExecutionBlockConfigurationMemberEcsCapacityIncreaseConfig
	}{
		{
			name: "null config",
			input: stepModel{
				EcsCapacityIncreaseConfig: types.ListNull(getEcsCapacityIncreaseConfigObjectType()),
			},
			expected: nil,
		},
		{
			name: "config with services",
			input: stepModel{
				EcsCapacityIncreaseConfig: types.ListValueMust(getEcsCapacityIncreaseConfigObjectType(), []attr.Value{
					types.ObjectValueMust(getEcsCapacityIncreaseConfigObjectType().AttrTypes, map[string]attr.Value{
						"capacity_monitoring_approach": types.StringValue("sampledMaxInLast24Hours"),
						"target_percent":               types.Int64Value(200),
						"timeout_minutes":              types.Int64Value(45),
						"services": types.ListValueMust(getServiceObjectType(), []attr.Value{
							types.ObjectValueMust(getServiceObjectType().AttrTypes, map[string]attr.Value{
								"cluster_arn":        types.StringValue("arn:aws:ecs:us-east-1:123456789012:cluster/test-cluster"),
								"service_arn":        types.StringValue("arn:aws:ecs:us-east-1:123456789012:service/test-service"),
								"cross_account_role": types.StringValue("test-role"),
								"external_id":        types.StringValue("test-external-id"),
							}),
						}),
						"ungraceful": types.ListValueMust(getEcsUngracefulObjectType(), []attr.Value{
							types.ObjectValueMust(getEcsUngracefulObjectType().AttrTypes, map[string]attr.Value{
								"minimum_success_percentage": types.Int64Value(90),
							}),
						}),
					}),
				}),
			},
			expected: &awstypes.ExecutionBlockConfigurationMemberEcsCapacityIncreaseConfig{
				Value: awstypes.EcsCapacityIncreaseConfiguration{
					CapacityMonitoringApproach: awstypes.EcsCapacityMonitoringApproachSampledMaxInLast24Hours,
					TargetPercent:              aws.Int32(200),
					TimeoutMinutes:             aws.Int32(45),
					Services: []awstypes.Service{
						{
							ClusterArn:       aws.String("arn:aws:ecs:us-east-1:123456789012:cluster/test-cluster"),
							ServiceArn:       aws.String("arn:aws:ecs:us-east-1:123456789012:service/test-service"),
							CrossAccountRole: aws.String("test-role"),
							ExternalId:       aws.String("test-external-id"),
						},
					},
					Ungraceful: &awstypes.EcsUngraceful{
						MinimumSuccessPercentage: aws.Int32(90),
					},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := expandEcsCapacityIncreaseConfig(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestExpandArcRoutingControlConfig(t *testing.T) {
	tests := []struct {
		name     string
		input    stepModel
		expected *awstypes.ExecutionBlockConfigurationMemberArcRoutingControlConfig
	}{
		{
			name: "null config",
			input: stepModel{
				ArcRoutingControlConfig: types.ListNull(getArcRoutingControlConfigObjectType()),
			},
			expected: nil,
		},
		{
			name: "config with routing controls",
			input: stepModel{
				ArcRoutingControlConfig: types.ListValueMust(getArcRoutingControlConfigObjectType(), []attr.Value{
					types.ObjectValueMust(getArcRoutingControlConfigObjectType().AttrTypes, map[string]attr.Value{
						"cross_account_role": types.StringValue("test-role"),
						"external_id":        types.StringValue("test-external-id"),
						"timeout_minutes":    types.Int64Value(30),
						"region_and_routing_controls": types.ListValueMust(getRegionAndRoutingControlsObjectType(), []attr.Value{
							types.ObjectValueMust(getRegionAndRoutingControlsObjectType().AttrTypes, map[string]attr.Value{
								"region": types.StringValue("us-east-1"),
								"routing_control_arns": types.ListValueMust(types.StringType, []attr.Value{
									types.StringValue("arn:aws:route53-recovery-control::123456789012:controlpanel/test/routingcontrol/test1"),
								}),
							}),
						}),
					}),
				}),
			},
			expected: &awstypes.ExecutionBlockConfigurationMemberArcRoutingControlConfig{
				Value: awstypes.ArcRoutingControlConfiguration{
					CrossAccountRole: aws.String("test-role"),
					ExternalId:       aws.String("test-external-id"),
					TimeoutMinutes:   aws.Int32(30),
					RegionAndRoutingControls: map[string][]awstypes.ArcRoutingControlState{
						"us-east-1": {
							{
								RoutingControlArn: aws.String("arn:aws:route53-recovery-control::123456789012:controlpanel/test/routingcontrol/test1"),
								State:             awstypes.RoutingControlStateChangeOn,
							},
						},
					},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := expandArcRoutingControlConfig(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}
