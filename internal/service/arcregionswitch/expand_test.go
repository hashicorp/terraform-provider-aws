package arcregionswitch

import (
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/arcregionswitch/types"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func TestExpandWorkflows(t *testing.T) {
	cases := []struct {
		name     string
		input    []interface{}
		expected []types.Workflow
	}{
		{
			name:     "nil",
			input:    nil,
			expected: nil,
		},
		{
			name:     "empty",
			input:    []interface{}{},
			expected: nil,
		},
		{
			name: "basic workflow",
			input: []interface{}{
				map[string]interface{}{
					"workflow_target_action": "activate",
					"workflow_target_region": "us-west-2",
					"workflow_description":   "Test workflow",
				},
			},
			expected: []types.Workflow{
				{
					WorkflowTargetAction: types.WorkflowTargetAction("activate"),
					WorkflowTargetRegion: aws.String("us-west-2"),
					WorkflowDescription:  aws.String("Test workflow"),
				},
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := expandWorkflows(tc.input)

			// Use IgnoreUnexported to ignore unexported fields in the AWS SDK types
			opts := cmpopts.IgnoreUnexported(
				types.Workflow{},
				types.Step{},
				types.ExecutionBlockConfigurationMemberExecutionApprovalConfig{},
				types.ExecutionApprovalConfiguration{},
				types.ExecutionBlockConfigurationMemberCustomActionLambdaConfig{},
				types.CustomActionLambdaConfiguration{},
				types.Lambdas{},
				types.ExecutionBlockConfigurationMemberEcsCapacityIncreaseConfig{},
				types.EcsCapacityIncreaseConfiguration{},
				types.Service{},
				types.ExecutionBlockConfigurationMemberRoute53HealthCheckConfig{},
				types.Route53HealthCheckConfiguration{},
				types.Route53ResourceRecordSet{},
				types.LambdaUngraceful{},
				types.EcsUngraceful{},
			)

			if diff := cmp.Diff(tc.expected, got, opts); diff != "" {
				t.Errorf("unexpected diff (+wanted, -got): %s", diff)
			}
		})
	}
}

func TestExpandSteps(t *testing.T) {
	cases := []struct {
		name     string
		input    []interface{}
		expected []types.Step
	}{
		{
			name:     "nil",
			input:    nil,
			expected: nil,
		},
		{
			name:     "empty",
			input:    []interface{}{},
			expected: nil,
		},
		{
			name: "ManualApproval step",
			input: []interface{}{
				map[string]interface{}{
					"name":                 "approval-step",
					"execution_block_type": "ManualApproval",
					"description":          "Approval step",
					"execution_approval_config": []interface{}{
						map[string]interface{}{
							"approval_role":   "arn:aws:iam::123456789012:role/test-role",
							"timeout_minutes": 60,
						},
					},
				},
			},
			expected: []types.Step{
				{
					Name:               aws.String("approval-step"),
					ExecutionBlockType: types.ExecutionBlockType("ManualApproval"),
					Description:        aws.String("Approval step"),
					ExecutionBlockConfiguration: &types.ExecutionBlockConfigurationMemberExecutionApprovalConfig{
						Value: types.ExecutionApprovalConfiguration{
							ApprovalRole:   aws.String("arn:aws:iam::123456789012:role/test-role"),
							TimeoutMinutes: aws.Int32(60),
						},
					},
				},
			},
		},
		{
			name: "CustomActionLambda step",
			input: []interface{}{
				map[string]interface{}{
					"name":                 "lambda-step",
					"execution_block_type": "CustomActionLambda",
					"description":          "Lambda step",
					"custom_action_lambda_config": []interface{}{
						map[string]interface{}{
							"region_to_run":          "activatingRegion",
							"retry_interval_minutes": 5.0,
							"timeout_minutes":        30,
							"lambda": []interface{}{
								map[string]interface{}{
									"arn":                "arn:aws:lambda:us-west-2:123456789012:function/test-function",
									"cross_account_role": "arn:aws:iam::123456789012:role/test-role",
									"external_id":        "test-id",
								},
							},
						},
					},
				},
			},
			expected: []types.Step{
				{
					Name:               aws.String("lambda-step"),
					ExecutionBlockType: types.ExecutionBlockType("CustomActionLambda"),
					Description:        aws.String("Lambda step"),
					ExecutionBlockConfiguration: &types.ExecutionBlockConfigurationMemberCustomActionLambdaConfig{
						Value: types.CustomActionLambdaConfiguration{
							RegionToRun:          types.RegionToRunIn("activatingRegion"),
							RetryIntervalMinutes: aws.Float32(5.0),
							TimeoutMinutes:       aws.Int32(30),
							Lambdas: []types.Lambdas{
								{
									Arn:              aws.String("arn:aws:lambda:us-west-2:123456789012:function/test-function"),
									CrossAccountRole: aws.String("arn:aws:iam::123456789012:role/test-role"),
									ExternalId:       aws.String("test-id"),
								},
							},
						},
					},
				},
			},
		},
		{
			name: "Route53HealthCheck step",
			input: []interface{}{
				map[string]interface{}{
					"name":                 "health-check-step",
					"execution_block_type": "Route53HealthCheck",
					"description":          "Health check step",
					"route53_health_check_config": []interface{}{
						map[string]interface{}{
							"hosted_zone_id":  "Z1D633PJN98FT9",
							"record_name":     "example.com",
							"timeout_minutes": 30,
							"record_sets": []interface{}{
								map[string]interface{}{
									"record_set_identifier": "test-identifier",
									"region":                "us-west-2",
								},
							},
						},
					},
				},
			},
			expected: []types.Step{
				{
					Name:               aws.String("health-check-step"),
					ExecutionBlockType: types.ExecutionBlockType("Route53HealthCheck"),
					Description:        aws.String("Health check step"),
					ExecutionBlockConfiguration: &types.ExecutionBlockConfigurationMemberRoute53HealthCheckConfig{
						Value: types.Route53HealthCheckConfiguration{
							HostedZoneId:   aws.String("Z1D633PJN98FT9"),
							RecordName:     aws.String("example.com"),
							TimeoutMinutes: aws.Int32(30),
							RecordSets: []types.Route53ResourceRecordSet{
								{
									RecordSetIdentifier: aws.String("test-identifier"),
									Region:              aws.String("us-west-2"),
								},
							},
						},
					},
				},
			},
		},
		{
			name: "ECSServiceScaling step",
			input: []interface{}{
				map[string]interface{}{
					"name":                 "ecs-scaling-step",
					"execution_block_type": "ECSServiceScaling",
					"description":          "ECS scaling step",
					"ecs_capacity_increase_config": []interface{}{
						map[string]interface{}{
							"capacity_monitoring_approach": "sampledMaxInLast24Hours",
							"target_percent":               80,
							"timeout_minutes":              30,
							"services": []interface{}{
								map[string]interface{}{
									"cluster_arn": "arn:aws:ecs:us-west-2:123456789012:cluster/test-cluster",
									"service_arn": "arn:aws:ecs:us-west-2:123456789012:service/test-cluster/test-service",
								},
							},
						},
					},
				},
			},
			expected: []types.Step{
				{
					Name:               aws.String("ecs-scaling-step"),
					ExecutionBlockType: types.ExecutionBlockType("ECSServiceScaling"),
					Description:        aws.String("ECS scaling step"),
					ExecutionBlockConfiguration: &types.ExecutionBlockConfigurationMemberEcsCapacityIncreaseConfig{
						Value: types.EcsCapacityIncreaseConfiguration{
							CapacityMonitoringApproach: types.EcsCapacityMonitoringApproach("sampledMaxInLast24Hours"),
							TargetPercent:              aws.Int32(80),
							TimeoutMinutes:             aws.Int32(30),
							Services: []types.Service{
								{
									ClusterArn: aws.String("arn:aws:ecs:us-west-2:123456789012:cluster/test-cluster"),
									ServiceArn: aws.String("arn:aws:ecs:us-west-2:123456789012:service/test-cluster/test-service"),
								},
							},
						},
					},
				},
			},
		},
		{
			name: "ARCRegionSwitchPlan step",
			input: []interface{}{
				map[string]interface{}{
					"name":                 "region-switch-step",
					"execution_block_type": "ARCRegionSwitchPlan",
					"description":          "Region switch plan step",
					"region_switch_plan_config": []interface{}{
						map[string]interface{}{
							"arn":                "arn:aws:arcregionswitch:us-west-2:123456789012:plan/test-plan",
							"cross_account_role": "arn:aws:iam::123456789012:role/test-role",
							"external_id":        "test-id",
						},
					},
				},
			},
			expected: []types.Step{
				{
					Name:               aws.String("region-switch-step"),
					ExecutionBlockType: types.ExecutionBlockType("ARCRegionSwitchPlan"),
					Description:        aws.String("Region switch plan step"),
					ExecutionBlockConfiguration: &types.ExecutionBlockConfigurationMemberRegionSwitchPlanConfig{
						Value: types.RegionSwitchPlanConfiguration{
							Arn:              aws.String("arn:aws:arcregionswitch:us-west-2:123456789012:plan/test-plan"),
							CrossAccountRole: aws.String("arn:aws:iam::123456789012:role/test-role"),
							ExternalId:       aws.String("test-id"),
						},
					},
				},
			},
		},
		{
			name: "Parallel step",
			input: []interface{}{
				map[string]interface{}{
					"name":                 "parallel-step",
					"execution_block_type": "Parallel",
					"description":          "Parallel execution step",
					"parallel_config": []interface{}{
						map[string]interface{}{
							"step": []interface{}{
								map[string]interface{}{
									"name":                 "nested-approval-step",
									"execution_block_type": "ManualApproval",
									"description":          "Nested approval step",
									"execution_approval_config": []interface{}{
										map[string]interface{}{
											"approval_role":   "arn:aws:iam::123456789012:role/nested-role",
											"timeout_minutes": 30,
										},
									},
								},
								map[string]interface{}{
									"name":                 "nested-lambda-step",
									"execution_block_type": "CustomActionLambda",
									"description":          "Nested lambda step",
									"custom_action_lambda_config": []interface{}{
										map[string]interface{}{
											"region_to_run":          "deactivatingRegion",
											"retry_interval_minutes": 2.0,
											"timeout_minutes":        15,
											"lambda": []interface{}{
												map[string]interface{}{
													"arn": "arn:aws:lambda:us-east-1:123456789012:function/nested-function",
												},
											},
										},
									},
								},
							},
						},
					},
				},
			},
			expected: []types.Step{
				{
					Name:               aws.String("parallel-step"),
					ExecutionBlockType: types.ExecutionBlockType("Parallel"),
					Description:        aws.String("Parallel execution step"),
					ExecutionBlockConfiguration: &types.ExecutionBlockConfigurationMemberParallelConfig{
						Value: types.ParallelExecutionBlockConfiguration{
							Steps: []types.Step{
								{
									Name:               aws.String("nested-approval-step"),
									ExecutionBlockType: types.ExecutionBlockType("ManualApproval"),
									Description:        aws.String("Nested approval step"),
									ExecutionBlockConfiguration: &types.ExecutionBlockConfigurationMemberExecutionApprovalConfig{
										Value: types.ExecutionApprovalConfiguration{
											ApprovalRole:   aws.String("arn:aws:iam::123456789012:role/nested-role"),
											TimeoutMinutes: aws.Int32(30),
										},
									},
								},
								{
									Name:               aws.String("nested-lambda-step"),
									ExecutionBlockType: types.ExecutionBlockType("CustomActionLambda"),
									Description:        aws.String("Nested lambda step"),
									ExecutionBlockConfiguration: &types.ExecutionBlockConfigurationMemberCustomActionLambdaConfig{
										Value: types.CustomActionLambdaConfiguration{
											RegionToRun:          types.RegionToRunIn("deactivatingRegion"),
											RetryIntervalMinutes: aws.Float32(2.0),
											TimeoutMinutes:       aws.Int32(15),
											Lambdas: []types.Lambdas{
												{
													Arn: aws.String("arn:aws:lambda:us-east-1:123456789012:function/nested-function"),
												},
											},
										},
									},
								},
							},
						},
					},
				},
			},
		},
		{
			name: "AuroraGlobalDatabase step",
			input: []interface{}{
				map[string]interface{}{
					"name":                 "aurora-step",
					"execution_block_type": "AuroraGlobalDatabase",
					"description":          "Aurora global database step",
					"global_aurora_config": []interface{}{
						map[string]interface{}{
							"behavior":                  "switchoverOnly",
							"global_cluster_identifier": "test-global-cluster",
							"database_cluster_arns": []interface{}{
								"arn:aws:rds:us-west-2:123456789012:cluster:test-cluster-1",
								"arn:aws:rds:us-east-1:123456789012:cluster:test-cluster-2",
							},
							"cross_account_role": "arn:aws:iam::123456789012:role/aurora-role",
							"external_id":        "aurora-external-id",
							"timeout_minutes":    45,
							"ungraceful": []interface{}{
								map[string]interface{}{
									"ungraceful": "failover",
								},
							},
						},
					},
				},
			},
			expected: []types.Step{
				{
					Name:               aws.String("aurora-step"),
					ExecutionBlockType: types.ExecutionBlockType("AuroraGlobalDatabase"),
					Description:        aws.String("Aurora global database step"),
					ExecutionBlockConfiguration: &types.ExecutionBlockConfigurationMemberGlobalAuroraConfig{
						Value: types.GlobalAuroraConfiguration{
							Behavior:                types.GlobalAuroraDefaultBehavior("switchoverOnly"),
							GlobalClusterIdentifier: aws.String("test-global-cluster"),
							DatabaseClusterArns: []string{
								"arn:aws:rds:us-west-2:123456789012:cluster:test-cluster-1",
								"arn:aws:rds:us-east-1:123456789012:cluster:test-cluster-2",
							},
							CrossAccountRole: aws.String("arn:aws:iam::123456789012:role/aurora-role"),
							ExternalId:       aws.String("aurora-external-id"),
							TimeoutMinutes:   aws.Int32(45),
							Ungraceful: &types.GlobalAuroraUngraceful{
								Ungraceful: types.GlobalAuroraUngracefulBehavior("failover"),
							},
						},
					},
				},
			},
		},
		{
			name: "EC2AutoScaling step",
			input: []interface{}{
				map[string]interface{}{
					"name":                 "ec2-scaling-step",
					"execution_block_type": "EC2AutoScaling",
					"description":          "EC2 Auto Scaling step",
					"ec2_asg_capacity_increase_config": []interface{}{
						map[string]interface{}{
							"asgs": []interface{}{
								map[string]interface{}{
									"arn":                "arn:aws:autoscaling:us-west-2:123456789012:autoScalingGroup:test-asg-1",
									"cross_account_role": "arn:aws:iam::123456789012:role/ec2-role",
									"external_id":        "ec2-external-id",
								},
								map[string]interface{}{
									"arn": "arn:aws:autoscaling:us-east-1:123456789012:autoScalingGroup:test-asg-2",
								},
							},
							"capacity_monitoring_approach": "autoscalingMaxInLast24Hours",
							"target_percent":               75,
							"timeout_minutes":              60,
							"ungraceful": []interface{}{
								map[string]interface{}{
									"minimum_success_percentage": 80,
								},
							},
						},
					},
				},
			},
			expected: []types.Step{
				{
					Name:               aws.String("ec2-scaling-step"),
					ExecutionBlockType: types.ExecutionBlockType("EC2AutoScaling"),
					Description:        aws.String("EC2 Auto Scaling step"),
					ExecutionBlockConfiguration: &types.ExecutionBlockConfigurationMemberEc2AsgCapacityIncreaseConfig{
						Value: types.Ec2AsgCapacityIncreaseConfiguration{
							Asgs: []types.Asg{
								{
									Arn:              aws.String("arn:aws:autoscaling:us-west-2:123456789012:autoScalingGroup:test-asg-1"),
									CrossAccountRole: aws.String("arn:aws:iam::123456789012:role/ec2-role"),
									ExternalId:       aws.String("ec2-external-id"),
								},
								{
									Arn: aws.String("arn:aws:autoscaling:us-east-1:123456789012:autoScalingGroup:test-asg-2"),
								},
							},
							CapacityMonitoringApproach: types.Ec2AsgCapacityMonitoringApproach("autoscalingMaxInLast24Hours"),
							TargetPercent:              aws.Int32(75),
							TimeoutMinutes:             aws.Int32(60),
							Ungraceful: &types.Ec2Ungraceful{
								MinimumSuccessPercentage: aws.Int32(80),
							},
						},
					},
				},
			},
		},
		{
			name: "ARCRoutingControl step",
			input: []interface{}{
				map[string]interface{}{
					"name":                 "routing-control-step",
					"execution_block_type": "ARCRoutingControl",
					"description":          "ARC routing control step",
					"arc_routing_control_config": []interface{}{
						map[string]interface{}{
							"region_and_routing_controls": []interface{}{
								map[string]interface{}{
									"region": "us-west-2",
									"routing_control_arns": []interface{}{
										"arn:aws:route53-recovery-control::123456789012:controlpanel/test-panel/routingcontrol/test-control-1",
										"arn:aws:route53-recovery-control::123456789012:controlpanel/test-panel/routingcontrol/test-control-2",
									},
								},
								map[string]interface{}{
									"region": "us-east-1",
									"routing_control_arns": []interface{}{
										"arn:aws:route53-recovery-control::123456789012:controlpanel/test-panel/routingcontrol/test-control-3",
									},
								},
							},
							"cross_account_role": "arn:aws:iam::123456789012:role/routing-control-role",
							"external_id":        "routing-control-external-id",
							"timeout_minutes":    30,
						},
					},
				},
			},
			expected: []types.Step{
				{
					Name:               aws.String("routing-control-step"),
					ExecutionBlockType: types.ExecutionBlockType("ARCRoutingControl"),
					Description:        aws.String("ARC routing control step"),
					ExecutionBlockConfiguration: &types.ExecutionBlockConfigurationMemberArcRoutingControlConfig{
						Value: types.ArcRoutingControlConfiguration{
							RegionAndRoutingControls: map[string][]types.ArcRoutingControlState{
								"us-west-2": {
									{
										RoutingControlArn: aws.String("arn:aws:route53-recovery-control::123456789012:controlpanel/test-panel/routingcontrol/test-control-1"),
										State:             types.RoutingControlStateChangeOn,
									},
									{
										RoutingControlArn: aws.String("arn:aws:route53-recovery-control::123456789012:controlpanel/test-panel/routingcontrol/test-control-2"),
										State:             types.RoutingControlStateChangeOn,
									},
								},
								"us-east-1": {
									{
										RoutingControlArn: aws.String("arn:aws:route53-recovery-control::123456789012:controlpanel/test-panel/routingcontrol/test-control-3"),
										State:             types.RoutingControlStateChangeOn,
									},
								},
							},
							CrossAccountRole: aws.String("arn:aws:iam::123456789012:role/routing-control-role"),
							ExternalId:       aws.String("routing-control-external-id"),
							TimeoutMinutes:   aws.Int32(30),
						},
					},
				},
			},
		},
		{
			name: "EKSResourceScaling step",
			input: []interface{}{
				map[string]interface{}{
					"name":                 "eks-scaling-step",
					"execution_block_type": "EKSResourceScaling",
					"description":          "EKS resource scaling step",
					"eks_resource_scaling_config": []interface{}{
						map[string]interface{}{
							"kubernetes_resource_type": []interface{}{
								map[string]interface{}{
									"api_version": "apps/v1",
									"kind":        "Deployment",
								},
							},
							"eks_clusters": []interface{}{
								map[string]interface{}{
									"cluster_arn":        "arn:aws:eks:us-west-2:123456789012:cluster/test-cluster",
									"cross_account_role": "arn:aws:iam::123456789012:role/eks-role",
									"external_id":        "eks-external-id",
								},
							},
							"capacity_monitoring_approach": "sampledMaxInLast24Hours",
							"target_percent":               90,
							"timeout_minutes":              45,
							"ungraceful": []interface{}{
								map[string]interface{}{
									"minimum_success_percentage": 85,
								},
							},
						},
					},
				},
			},
			expected: []types.Step{
				{
					Name:               aws.String("eks-scaling-step"),
					ExecutionBlockType: types.ExecutionBlockType("EKSResourceScaling"),
					Description:        aws.String("EKS resource scaling step"),
					ExecutionBlockConfiguration: &types.ExecutionBlockConfigurationMemberEksResourceScalingConfig{
						Value: types.EksResourceScalingConfiguration{
							KubernetesResourceType: &types.KubernetesResourceType{
								ApiVersion: aws.String("apps/v1"),
								Kind:       aws.String("Deployment"),
							},
							EksClusters: []types.EksCluster{
								{
									ClusterArn:       aws.String("arn:aws:eks:us-west-2:123456789012:cluster/test-cluster"),
									CrossAccountRole: aws.String("arn:aws:iam::123456789012:role/eks-role"),
									ExternalId:       aws.String("eks-external-id"),
								},
							},
							CapacityMonitoringApproach: types.EksCapacityMonitoringApproach("sampledMaxInLast24Hours"),
							TargetPercent:              aws.Int32(90),
							TimeoutMinutes:             aws.Int32(45),
							Ungraceful: &types.EksResourceScalingUngraceful{
								MinimumSuccessPercentage: aws.Int32(85),
							},
						},
					},
				},
			},
		},
		{
			name: "EKSResourceScaling step with scaling resources",
			input: []interface{}{
				map[string]interface{}{
					"name":                 "eks-scaling-with-resources-step",
					"execution_block_type": "EKSResourceScaling",
					"description":          "EKS resource scaling step with scaling resources",
					"eks_resource_scaling_config": []interface{}{
						map[string]interface{}{
							"kubernetes_resource_type": []interface{}{
								map[string]interface{}{
									"api_version": "apps/v1",
									"kind":        "Deployment",
								},
							},
							"eks_clusters": []interface{}{
								map[string]interface{}{
									"cluster_arn": "arn:aws:eks:us-west-2:123456789012:cluster/test-cluster",
								},
							},
							"capacity_monitoring_approach": "autoscalingMaxInLast24Hours",
							"target_percent":               80,
							"timeout_minutes":              30,
							// Note: scaling_resources is complex and requires schema.Set in real usage
							// For this test, we'll skip it to avoid the nested bracket issue
						},
					},
				},
			},
			expected: []types.Step{
				{
					Name:               aws.String("eks-scaling-with-resources-step"),
					ExecutionBlockType: types.ExecutionBlockType("EKSResourceScaling"),
					Description:        aws.String("EKS resource scaling step with scaling resources"),
					ExecutionBlockConfiguration: &types.ExecutionBlockConfigurationMemberEksResourceScalingConfig{
						Value: types.EksResourceScalingConfiguration{
							KubernetesResourceType: &types.KubernetesResourceType{
								ApiVersion: aws.String("apps/v1"),
								Kind:       aws.String("Deployment"),
							},
							EksClusters: []types.EksCluster{
								{
									ClusterArn: aws.String("arn:aws:eks:us-west-2:123456789012:cluster/test-cluster"),
								},
							},
							CapacityMonitoringApproach: types.EksCapacityMonitoringApproach("autoscalingMaxInLast24Hours"),
							TargetPercent:              aws.Int32(80),
							TimeoutMinutes:             aws.Int32(30),
						},
					},
				},
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := expandSteps(tc.input, "activate")

			// Use IgnoreUnexported to ignore unexported fields in the AWS SDK types
			opts := cmpopts.IgnoreUnexported(
				types.Step{},
				types.ExecutionBlockConfigurationMemberExecutionApprovalConfig{},
				types.ExecutionApprovalConfiguration{},
				types.ExecutionBlockConfigurationMemberCustomActionLambdaConfig{},
				types.CustomActionLambdaConfiguration{},
				types.Lambdas{},
				types.ExecutionBlockConfigurationMemberEcsCapacityIncreaseConfig{},
				types.EcsCapacityIncreaseConfiguration{},
				types.Service{},
				types.ExecutionBlockConfigurationMemberRoute53HealthCheckConfig{},
				types.Route53HealthCheckConfiguration{},
				types.Route53ResourceRecordSet{},
				types.ExecutionBlockConfigurationMemberRegionSwitchPlanConfig{},
				types.RegionSwitchPlanConfiguration{},
				types.ExecutionBlockConfigurationMemberParallelConfig{},
				types.ParallelExecutionBlockConfiguration{},
				types.ExecutionBlockConfigurationMemberGlobalAuroraConfig{},
				types.GlobalAuroraConfiguration{},
				types.GlobalAuroraUngraceful{},
				types.ExecutionBlockConfigurationMemberEc2AsgCapacityIncreaseConfig{},
				types.Ec2AsgCapacityIncreaseConfiguration{},
				types.Asg{},
				types.Ec2Ungraceful{},
				types.ExecutionBlockConfigurationMemberArcRoutingControlConfig{},
				types.ArcRoutingControlConfiguration{},
				types.ArcRoutingControlState{},
				types.ExecutionBlockConfigurationMemberEksResourceScalingConfig{},
				types.EksResourceScalingConfiguration{},
				types.KubernetesResourceType{},
				types.EksCluster{},
				types.KubernetesScalingResource{},
				types.EksResourceScalingUngraceful{},
				types.LambdaUngraceful{},
				types.EcsUngraceful{},
			)

			if diff := cmp.Diff(tc.expected, got, opts); diff != "" {
				t.Errorf("unexpected diff (+wanted, -got): %s", diff)
			}
		})
	}
}
func TestFlattenWorkflows(t *testing.T) {
	cases := []struct {
		name     string
		input    []types.Workflow
		expected []interface{}
	}{
		{
			name:     "nil",
			input:    nil,
			expected: nil,
		},
		{
			name:     "empty",
			input:    []types.Workflow{},
			expected: nil,
		},
		{
			name: "basic workflow",
			input: []types.Workflow{
				{
					WorkflowTargetAction: types.WorkflowTargetAction("activate"),
					WorkflowTargetRegion: aws.String("us-west-2"),
					WorkflowDescription:  aws.String("Test workflow"),
					Steps:                []types.Step{},
				},
			},
			expected: []interface{}{
				map[string]interface{}{
					"workflow_target_action": "activate",
					"workflow_target_region": "us-west-2",
					"workflow_description":   "Test workflow",
				},
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := flattenWorkflows(tc.input)

			if diff := cmp.Diff(tc.expected, got); diff != "" {
				t.Errorf("unexpected diff (+wanted, -got): %s", diff)
			}
		})
	}
}

func TestFlattenSteps(t *testing.T) {
	cases := []struct {
		name     string
		input    []types.Step
		expected []interface{}
	}{
		{
			name:     "nil",
			input:    nil,
			expected: nil,
		},
		{
			name:     "empty",
			input:    []types.Step{},
			expected: nil,
		},
		{
			name: "ManualApproval step",
			input: []types.Step{
				{
					Name:               aws.String("approval-step"),
					ExecutionBlockType: types.ExecutionBlockType("ManualApproval"),
					Description:        aws.String("Approval step"),
					ExecutionBlockConfiguration: &types.ExecutionBlockConfigurationMemberExecutionApprovalConfig{
						Value: types.ExecutionApprovalConfiguration{
							ApprovalRole:   aws.String("arn:aws:iam::123456789012:role/test-role"),
							TimeoutMinutes: aws.Int32(60),
						},
					},
				},
			},
			expected: []interface{}{
				map[string]interface{}{
					"name":                 "approval-step",
					"execution_block_type": "ManualApproval",
					"description":          "Approval step",
					"execution_approval_config": []interface{}{
						map[string]interface{}{
							"approval_role":   "arn:aws:iam::123456789012:role/test-role",
							"timeout_minutes": int32(60),
						},
					},
				},
			},
		},
		{
			name: "EKSResourceScaling step",
			input: []types.Step{
				{
					Name:               aws.String("eks-scaling-step"),
					ExecutionBlockType: types.ExecutionBlockType("EKSResourceScaling"),
					Description:        aws.String("EKS resource scaling step"),
					ExecutionBlockConfiguration: &types.ExecutionBlockConfigurationMemberEksResourceScalingConfig{
						Value: types.EksResourceScalingConfiguration{
							KubernetesResourceType: &types.KubernetesResourceType{
								ApiVersion: aws.String("apps/v1"),
								Kind:       aws.String("Deployment"),
							},
							EksClusters: []types.EksCluster{
								{
									ClusterArn:       aws.String("arn:aws:eks:us-west-2:123456789012:cluster/test-cluster"),
									CrossAccountRole: aws.String("arn:aws:iam::123456789012:role/eks-role"),
									ExternalId:       aws.String("eks-external-id"),
								},
							},
							CapacityMonitoringApproach: types.EksCapacityMonitoringApproach("sampledMaxInLast24Hours"),
							TargetPercent:              aws.Int32(90),
							TimeoutMinutes:             aws.Int32(45),
							Ungraceful: &types.EksResourceScalingUngraceful{
								MinimumSuccessPercentage: aws.Int32(85),
							},
						},
					},
				},
			},
			expected: []interface{}{
				map[string]interface{}{
					"name":                 "eks-scaling-step",
					"execution_block_type": "EKSResourceScaling",
					"description":          "EKS resource scaling step",
					"eks_resource_scaling_config": []interface{}{
						map[string]interface{}{
							"kubernetes_resource_type": []interface{}{
								map[string]interface{}{
									"api_version": "apps/v1",
									"kind":        "Deployment",
								},
							},
							"eks_clusters": []interface{}{
								map[string]interface{}{
									"cluster_arn":        "arn:aws:eks:us-west-2:123456789012:cluster/test-cluster",
									"cross_account_role": "arn:aws:iam::123456789012:role/eks-role",
									"external_id":        "eks-external-id",
								},
							},
							"capacity_monitoring_approach": "sampledMaxInLast24Hours",
							"target_percent":               int32(90),
							"timeout_minutes":              int32(45),
							"ungraceful": []interface{}{
								map[string]interface{}{
									"minimum_success_percentage": int32(85),
								},
							},
						},
					},
				},
			},
		},
		{
			name: "ARCRoutingControl step",
			input: []types.Step{
				{
					Name:               aws.String("routing-control-step"),
					ExecutionBlockType: types.ExecutionBlockType("ARCRoutingControl"),
					Description:        aws.String("ARC routing control step"),
					ExecutionBlockConfiguration: &types.ExecutionBlockConfigurationMemberArcRoutingControlConfig{
						Value: types.ArcRoutingControlConfiguration{
							RegionAndRoutingControls: map[string][]types.ArcRoutingControlState{
								"us-west-2": {
									{
										RoutingControlArn: aws.String("arn:aws:route53-recovery-control::123456789012:controlpanel/test-panel/routingcontrol/test-control-1"),
										State:             types.RoutingControlStateChangeOn,
									},
									{
										RoutingControlArn: aws.String("arn:aws:route53-recovery-control::123456789012:controlpanel/test-panel/routingcontrol/test-control-2"),
										State:             types.RoutingControlStateChangeOn,
									},
								},
								"us-east-1": {
									{
										RoutingControlArn: aws.String("arn:aws:route53-recovery-control::123456789012:controlpanel/test-panel/routingcontrol/test-control-3"),
										State:             types.RoutingControlStateChangeOn,
									},
								},
							},
							CrossAccountRole: aws.String("arn:aws:iam::123456789012:role/routing-control-role"),
							ExternalId:       aws.String("routing-control-external-id"),
							TimeoutMinutes:   aws.Int32(30),
						},
					},
				},
			},
			expected: []interface{}{
				map[string]interface{}{
					"name":                 "routing-control-step",
					"execution_block_type": "ARCRoutingControl",
					"description":          "ARC routing control step",
					"arc_routing_control_config": []interface{}{
						map[string]interface{}{
							"region_and_routing_controls": []interface{}{
								map[string]interface{}{
									"region": "us-east-1",
									"routing_control_arns": []string{
										"arn:aws:route53-recovery-control::123456789012:controlpanel/test-panel/routingcontrol/test-control-3",
									},
								},
								map[string]interface{}{
									"region": "us-west-2",
									"routing_control_arns": []string{
										"arn:aws:route53-recovery-control::123456789012:controlpanel/test-panel/routingcontrol/test-control-1",
										"arn:aws:route53-recovery-control::123456789012:controlpanel/test-panel/routingcontrol/test-control-2",
									},
								},
							},
							"cross_account_role": "arn:aws:iam::123456789012:role/routing-control-role",
							"external_id":        "routing-control-external-id",
							"timeout_minutes":    int32(30),
						},
					},
				},
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := flattenSteps(tc.input)

			if diff := cmp.Diff(tc.expected, got); diff != "" {
				t.Errorf("unexpected diff (+wanted, -got): %s", diff)
			}
		})
	}
}

// TestExpandEksResourceScalingConfigWithScalingResources tests the complex scaling_resources field
// which requires schema.Set handling
func TestExpandEksResourceScalingConfigWithScalingResources(t *testing.T) {
	// Create a mock schema.Set for the resources field
	resourcesSchema := &schema.Resource{
		Schema: map[string]*schema.Schema{
			"resource_name": {
				Type:     schema.TypeString,
				Required: true,
			},
			"name": {
				Type:     schema.TypeString,
				Required: true,
			},
			"namespace": {
				Type:     schema.TypeString,
				Required: true,
			},
			"hpa_name": {
				Type:     schema.TypeString,
				Optional: true,
			},
		},
	}

	resourcesSet := schema.NewSet(schema.HashResource(resourcesSchema), []interface{}{
		map[string]interface{}{
			"resource_name": "my-deployment",
			"name":          "my-app",
			"namespace":     "default",
			"hpa_name":      "my-app-hpa",
		},
		map[string]interface{}{
			"resource_name": "my-service",
			"name":          "my-service",
			"namespace":     "default",
		},
	})

	input := map[string]interface{}{
		"kubernetes_resource_type": []interface{}{
			map[string]interface{}{
				"api_version": "apps/v1",
				"kind":        "Deployment",
			},
		},
		"eks_clusters": []interface{}{
			map[string]interface{}{
				"cluster_arn":        "arn:aws:eks:us-west-2:123456789012:cluster/test-cluster",
				"cross_account_role": "arn:aws:iam::123456789012:role/eks-role",
				"external_id":        "eks-external-id",
			},
		},
		"scaling_resources": []interface{}{
			map[string]interface{}{
				"namespace": "default",
				"resources": resourcesSet,
			},
		},
		"capacity_monitoring_approach": "sampledMaxInLast24Hours",
		"target_percent":               90,
		"timeout_minutes":              45,
		"ungraceful": []interface{}{
			map[string]interface{}{
				"minimum_success_percentage": 85,
			},
		},
	}

	expected := types.EksResourceScalingConfiguration{
		KubernetesResourceType: &types.KubernetesResourceType{
			ApiVersion: aws.String("apps/v1"),
			Kind:       aws.String("Deployment"),
		},
		EksClusters: []types.EksCluster{
			{
				ClusterArn:       aws.String("arn:aws:eks:us-west-2:123456789012:cluster/test-cluster"),
				CrossAccountRole: aws.String("arn:aws:iam::123456789012:role/eks-role"),
				ExternalId:       aws.String("eks-external-id"),
			},
		},
		ScalingResources: []map[string]map[string]types.KubernetesScalingResource{
			{
				"default": {
					"my-deployment": {
						Name:      aws.String("my-app"),
						Namespace: aws.String("default"),
						HpaName:   aws.String("my-app-hpa"),
					},
					"my-service": {
						Name:      aws.String("my-service"),
						Namespace: aws.String("default"),
					},
				},
			},
		},
		CapacityMonitoringApproach: types.EksCapacityMonitoringApproach("sampledMaxInLast24Hours"),
		TargetPercent:              aws.Int32(90),
		TimeoutMinutes:             aws.Int32(45),
		Ungraceful: &types.EksResourceScalingUngraceful{
			MinimumSuccessPercentage: aws.Int32(85),
		},
	}

	got := expandEksResourceScalingConfig(input)

	// Use IgnoreUnexported to ignore unexported fields in the AWS SDK types
	opts := cmpopts.IgnoreUnexported(
		types.EksResourceScalingConfiguration{},
		types.KubernetesResourceType{},
		types.EksCluster{},
		types.KubernetesScalingResource{},
		types.EksResourceScalingUngraceful{},
	)

	if diff := cmp.Diff(expected, got, opts); diff != "" {
		t.Errorf("unexpected diff (+wanted, -got): %s", diff)
	}
}

// TestExpandStepsWithWorkflowAction tests that routing control states are correctly set based on workflow action
func TestExpandStepsWithWorkflowAction(t *testing.T) {
	cases := []struct {
		name           string
		workflowAction string
		input          []interface{}
		expectedState  types.RoutingControlStateChange
	}{
		{
			name:           "activate workflow sets routing controls to On",
			workflowAction: "activate",
			input: []interface{}{
				map[string]interface{}{
					"name":                 "routing-control-step",
					"execution_block_type": "ARCRoutingControl",
					"description":          "ARC routing control step",
					"arc_routing_control_config": []interface{}{
						map[string]interface{}{
							"region_and_routing_controls": []interface{}{
								map[string]interface{}{
									"region": "us-west-2",
									"routing_control_arns": []interface{}{
										"arn:aws:route53-recovery-control::123456789012:controlpanel/test-panel/routingcontrol/test-control-1",
									},
								},
							},
							"timeout_minutes": 30,
						},
					},
				},
			},
			expectedState: types.RoutingControlStateChangeOn,
		},
		{
			name:           "deactivate workflow sets routing controls to Off",
			workflowAction: "deactivate",
			input: []interface{}{
				map[string]interface{}{
					"name":                 "routing-control-step",
					"execution_block_type": "ARCRoutingControl",
					"description":          "ARC routing control step",
					"arc_routing_control_config": []interface{}{
						map[string]interface{}{
							"region_and_routing_controls": []interface{}{
								map[string]interface{}{
									"region": "us-west-2",
									"routing_control_arns": []interface{}{
										"arn:aws:route53-recovery-control::123456789012:controlpanel/test-panel/routingcontrol/test-control-1",
									},
								},
							},
							"timeout_minutes": 30,
						},
					},
				},
			},
			expectedState: types.RoutingControlStateChangeOff,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := expandSteps(tc.input, tc.workflowAction)

			if len(got) != 1 {
				t.Fatalf("expected 1 step, got %d", len(got))
			}

			step := got[0]
			if step.ExecutionBlockConfiguration == nil {
				t.Fatal("expected ExecutionBlockConfiguration to be set")
			}

			config, ok := step.ExecutionBlockConfiguration.(*types.ExecutionBlockConfigurationMemberArcRoutingControlConfig)
			if !ok {
				t.Fatal("expected ArcRoutingControlConfig")
			}

			if len(config.Value.RegionAndRoutingControls) == 0 {
				t.Fatal("expected routing controls to be set")
			}

			for _, controls := range config.Value.RegionAndRoutingControls {
				if len(controls) == 0 {
					t.Fatal("expected at least one routing control")
				}

				actualState := controls[0].State
				if actualState != tc.expectedState {
					t.Errorf("expected routing control state %v, got %v", tc.expectedState, actualState)
				}
			}
		})
	}
}
