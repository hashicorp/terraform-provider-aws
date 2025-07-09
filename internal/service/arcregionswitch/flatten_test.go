// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package arcregionswitch

import (
	"context"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	awstypes "github.com/aws/aws-sdk-go-v2/service/arcregionswitch/types"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/assert"
)

func TestFlattenExecutionApprovalConfig(t *testing.T) {
	tests := []struct {
		name     string
		input    *awstypes.ExecutionBlockConfigurationMemberExecutionApprovalConfig
		expected types.List
	}{
		{
			name: "config with timeout",
			input: &awstypes.ExecutionBlockConfigurationMemberExecutionApprovalConfig{
				Value: awstypes.ExecutionApprovalConfiguration{
					ApprovalRole:   aws.String("test-role"),
					TimeoutMinutes: aws.Int32(30),
				},
			},
			expected: types.ListValueMust(getExecutionApprovalConfigObjectType(), []attr.Value{
				types.ObjectValueMust(getExecutionApprovalConfigObjectType().AttrTypes, map[string]attr.Value{
					"approval_role":   types.StringValue("test-role"),
					"timeout_minutes": types.Int64Value(30),
				}),
			}),
		},
		{
			name: "config without timeout",
			input: &awstypes.ExecutionBlockConfigurationMemberExecutionApprovalConfig{
				Value: awstypes.ExecutionApprovalConfiguration{
					ApprovalRole: aws.String("test-role"),
				},
			},
			expected: types.ListValueMust(getExecutionApprovalConfigObjectType(), []attr.Value{
				types.ObjectValueMust(getExecutionApprovalConfigObjectType().AttrTypes, map[string]attr.Value{
					"approval_role":   types.StringValue("test-role"),
					"timeout_minutes": types.Int64Null(),
				}),
			}),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var diags diag.Diagnostics
			result := flattenExecutionApprovalConfig(tt.input, &diags)
			assert.False(t, diags.HasError())
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestFlattenRoute53HealthCheckConfig(t *testing.T) {
	tests := []struct {
		name     string
		input    *awstypes.ExecutionBlockConfigurationMemberRoute53HealthCheckConfig
		expected types.List
	}{
		{
			name: "config with all fields",
			input: &awstypes.ExecutionBlockConfigurationMemberRoute53HealthCheckConfig{
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
			expected: types.ListValueMust(getRoute53HealthCheckConfigObjectType(), []attr.Value{
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
		{
			name: "config with required fields only",
			input: &awstypes.ExecutionBlockConfigurationMemberRoute53HealthCheckConfig{
				Value: awstypes.Route53HealthCheckConfiguration{
					HostedZoneId: aws.String("Z123456789"),
					RecordName:   aws.String("test.example.com"),
				},
			},
			expected: types.ListValueMust(getRoute53HealthCheckConfigObjectType(), []attr.Value{
				types.ObjectValueMust(getRoute53HealthCheckConfigObjectType().AttrTypes, map[string]attr.Value{
					"hosted_zone_id":     types.StringValue("Z123456789"),
					"record_name":        types.StringValue("test.example.com"),
					"cross_account_role": types.StringNull(),
					"external_id":        types.StringNull(),
					"timeout_minutes":    types.Int64Null(),
					"record_sets":        types.ListNull(getRecordSetObjectType()),
				}),
			}),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var diags diag.Diagnostics
			result := flattenRoute53HealthCheckConfig(tt.input, &diags)
			assert.False(t, diags.HasError())
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestFlattenCustomActionLambdaConfig(t *testing.T) {
	tests := []struct {
		name     string
		input    *awstypes.ExecutionBlockConfigurationMemberCustomActionLambdaConfig
		expected types.List
	}{
		{
			name: "config with ungraceful behavior",
			input: &awstypes.ExecutionBlockConfigurationMemberCustomActionLambdaConfig{
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
			expected: types.ListValueMust(getCustomActionLambdaConfigObjectType(), []attr.Value{
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
		{
			name: "config without ungraceful",
			input: &awstypes.ExecutionBlockConfigurationMemberCustomActionLambdaConfig{
				Value: awstypes.CustomActionLambdaConfiguration{
					RegionToRun:          awstypes.RegionToRunInActivatingRegion,
					RetryIntervalMinutes: aws.Float32(5.0),
					Lambdas: []awstypes.Lambdas{
						{
							Arn: aws.String("arn:aws:lambda:us-east-1:123456789012:function:test"),
						},
					},
				},
			},
			expected: types.ListValueMust(getCustomActionLambdaConfigObjectType(), []attr.Value{
				types.ObjectValueMust(getCustomActionLambdaConfigObjectType().AttrTypes, map[string]attr.Value{
					"region_to_run":          types.StringValue("activatingRegion"),
					"retry_interval_minutes": types.Float64Value(5.0),
					"timeout_minutes":        types.Int64Null(),
					"lambda": types.ListValueMust(getLambdaObjectType(), []attr.Value{
						types.ObjectValueMust(getLambdaObjectType().AttrTypes, map[string]attr.Value{
							"arn":                types.StringValue("arn:aws:lambda:us-east-1:123456789012:function:test"),
							"cross_account_role": types.StringNull(),
							"external_id":        types.StringNull(),
						}),
					}),
					"ungraceful": types.ListNull(getUngracefulObjectType()),
				}),
			}),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var diags diag.Diagnostics
			result := flattenCustomActionLambdaConfig(tt.input, &diags)
			assert.False(t, diags.HasError())
			assert.Equal(t, tt.expected, result)
		})
	}
}
func TestFlattenGlobalAuroraConfig(t *testing.T) {
	tests := []struct {
		name     string
		input    *awstypes.ExecutionBlockConfigurationMemberGlobalAuroraConfig
		expected types.List
	}{
		{
			name: "config with all fields",
			input: &awstypes.ExecutionBlockConfigurationMemberGlobalAuroraConfig{
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
			expected: types.ListValueMust(getGlobalAuroraConfigObjectType(), []attr.Value{
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
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var diags diag.Diagnostics
			ctx := context.Background()
			result := flattenGlobalAuroraConfig(ctx, tt.input, &diags)
			assert.False(t, diags.HasError())
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestFlattenEc2AsgCapacityIncreaseConfig(t *testing.T) {
	tests := []struct {
		name     string
		input    *awstypes.ExecutionBlockConfigurationMemberEc2AsgCapacityIncreaseConfig
		expected types.List
	}{
		{
			name: "config with ungraceful",
			input: &awstypes.ExecutionBlockConfigurationMemberEc2AsgCapacityIncreaseConfig{
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
			expected: types.ListValueMust(getEc2AsgCapacityIncreaseConfigObjectType(), []attr.Value{
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
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var diags diag.Diagnostics
			result := flattenEc2AsgCapacityIncreaseConfig(tt.input, &diags)
			assert.False(t, diags.HasError())
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestFlattenArcRoutingControlConfig(t *testing.T) {
	tests := []struct {
		name     string
		input    *awstypes.ExecutionBlockConfigurationMemberArcRoutingControlConfig
		expected types.List
	}{
		{
			name: "config with routing controls",
			input: &awstypes.ExecutionBlockConfigurationMemberArcRoutingControlConfig{
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
			expected: types.ListValueMust(getArcRoutingControlConfigObjectType(), []attr.Value{
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
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var diags diag.Diagnostics
			ctx := context.Background()
			result := flattenArcRoutingControlConfig(ctx, tt.input, &diags)
			assert.False(t, diags.HasError())
			assert.Equal(t, tt.expected, result)
		})
	}
}
