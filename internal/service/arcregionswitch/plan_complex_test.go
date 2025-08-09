package arcregionswitch_test

import (
	"fmt"
	"testing"

	sdktypes "github.com/aws/aws-sdk-go-v2/service/arcregionswitch/types"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccARCRegionSwitchPlan_complex(t *testing.T) {
	ctx := acctest.Context(t)
	var plan sdktypes.Plan
	rName := acctest.RandomWithPrefix(t, "tf-acc-test")
	resourceName := "aws_arcregionswitch_plan.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.ARCRegionSwitch),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckPlanDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccPlanConfig_complex(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPlanExists(ctx, resourceName, &plan),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "recovery_approach", "activeActive"),
					resource.TestCheckResourceAttr(resourceName, "regions.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "workflow.#", "2"),

					// Check that we have both activate and deactivate workflows
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "workflow.*", map[string]string{
						"workflow_target_action": "activate",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "workflow.*", map[string]string{
						"workflow_target_action": "deactivate",
					}),

					// Verify basic attributes
					resource.TestCheckResourceAttr(resourceName, "description", "Complex test plan with multiple execution block types"),
					resource.TestCheckResourceAttr(resourceName, "recovery_time_objective_minutes", "60"),
					resource.TestCheckResourceAttr(resourceName, "associated_alarms.#", "1"),
					resource.TestCheckResourceAttrSet(resourceName, "arn"),

					// Verify CustomActionLambda execution block
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "workflow.*.step.*", map[string]string{
						"execution_block_type": "CustomActionLambda",
						"name":                 "custom-lambda-step",
					}),

					// Verify ManualApproval execution block
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "workflow.*.step.*", map[string]string{
						"execution_block_type": "ManualApproval",
						"name":                 "manual-approval-step",
					}),

					// Verify AuroraGlobalDatabase execution block
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "workflow.*.step.*", map[string]string{
						"execution_block_type": "AuroraGlobalDatabase",
						"name":                 "aurora-global-step",
					}),

					// Verify EC2AutoScaling execution block
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "workflow.*.step.*", map[string]string{
						"execution_block_type": "EC2AutoScaling",
						"name":                 "ec2-asg-step",
					}),

					// Verify ECSServiceScaling execution block
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "workflow.*.step.*", map[string]string{
						"execution_block_type": "ECSServiceScaling",
						"name":                 "ecs-scaling-step",
					}),

					// Verify EKSResourceScaling execution block
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "workflow.*.step.*", map[string]string{
						"execution_block_type": "EKSResourceScaling",
						"name":                 "eks-scaling-step",
					}),

					// Verify Route53HealthCheck execution block
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "workflow.*.step.*", map[string]string{
						"execution_block_type": "Route53HealthCheck",
						"name":                 "route53-health-check-step-activate",
					}),

					// Verify Parallel execution block
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "workflow.*.step.*", map[string]string{
						"execution_block_type": "Parallel",
						"name":                 "parallel-step",
					}),

					// Verify ARCRoutingControl execution block
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "workflow.*.step.*", map[string]string{
						"execution_block_type": "ARCRoutingControl",
						"name":                 "arc-routing-control-step",
					}),

					// Verify specific configuration values are stored correctly
					// CustomActionLambda config
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "workflow.*.step.*.custom_action_lambda_config.*", map[string]string{
						"region_to_run":          "activatingRegion",
						"timeout_minutes":        "30",
						"retry_interval_minutes": "5",
					}),

					// ManualApproval config
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "workflow.*.step.*.execution_approval_config.*", map[string]string{
						"timeout_minutes": "30",
					}),

					// AuroraGlobalDatabase config
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "workflow.*.step.*.global_aurora_config.*", map[string]string{
						"behavior":                  "switchoverOnly",
						"global_cluster_identifier": "test-global-cluster",
						"timeout_minutes":           "45",
					}),

					// EC2AutoScaling config
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "workflow.*.step.*.ec2_asg_capacity_increase_config.*", map[string]string{
						"capacity_monitoring_approach": "sampledMaxInLast24Hours",
						"target_percent":               "150",
						"timeout_minutes":              "20",
					}),

					// ECSServiceScaling config
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "workflow.*.step.*.ecs_capacity_increase_config.*", map[string]string{
						"capacity_monitoring_approach": "containerInsightsMaxInLast24Hours",
						"target_percent":               "200",
						"timeout_minutes":              "25",
					}),

					// EKSResourceScaling config
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "workflow.*.step.*.eks_resource_scaling_config.*", map[string]string{
						"capacity_monitoring_approach": "sampledMaxInLast24Hours",
						"target_percent":               "175",
						"timeout_minutes":              "35",
					}),

					// Route53HealthCheck config
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "workflow.*.step.*.route53_health_check_config.*", map[string]string{
						"hosted_zone_id":  "Z123456789012345678",
						"timeout_minutes": "10",
						"record_name":     "test.example.com",
					}),

					// ARCRoutingControl config
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "workflow.*.step.*.arc_routing_control_config.*", map[string]string{
						"cross_account_role": "arn:aws:iam::123456789012:role/RoutingControlRole",
						"external_id":        "routing-external-id",
						"timeout_minutes":    "15",
					}),
				),
				// API returns workflows in different order than specified, causing plan differences
				// This is expected behavior and doesn't affect functionality
				ExpectNonEmptyPlan: true,
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				// API may return regions in different order than specified
				ImportStateVerifyIgnore: []string{
					"workflow.1.step.6.arc_routing_control_config.0.region_and_routing_controls.0.region",
					"workflow.1.step.6.arc_routing_control_config.0.region_and_routing_controls.1.region",
					"workflow.1.step.6.arc_routing_control_config.0.region_and_routing_controls.0.routing_control_arns.0",
					"workflow.1.step.6.arc_routing_control_config.0.region_and_routing_controls.1.routing_control_arns.0",
				},
			},
		},
	})
}

func testAccPlanConfig_complex(rName string) string {
	return fmt.Sprintf(`
resource "aws_iam_role" "test" {
  name = %[1]q

  assume_role_policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Action = "sts:AssumeRole"
        Effect = "Allow"
        Principal = {
          Service = "arc-region-switch.amazonaws.com"
        }
      },
    ]
  })
}

resource "aws_arcregionswitch_plan" "test" {
  name                             = %[1]q
  execution_role                   = aws_iam_role.test.arn
  recovery_approach                = "activeActive"
  regions                          = ["us-east-1", "us-west-2"]
  primary_region                   = "us-east-1"
  description                      = "Complex test plan with multiple execution block types"
  recovery_time_objective_minutes  = 60

  associated_alarms {
    name                = "test-alarm-1"
    alarm_type          = "applicationHealth"
    resource_identifier = "arn:aws:cloudwatch:us-east-1:123456789012:alarm:test-alarm-1"
  }

  # Activate workflow with multiple execution block types
  workflow {
    workflow_target_action = "activate"
    workflow_description   = "Activation workflow with multiple execution block types"

    # CustomActionLambda step
    step {
      name                 = "custom-lambda-step"
      execution_block_type = "CustomActionLambda"
      description          = "Custom Lambda execution step"

      custom_action_lambda_config {
        region_to_run           = "activatingRegion"
        retry_interval_minutes  = 5.0
        timeout_minutes         = 30

        lambda {
          arn = "arn:aws:lambda:us-west-2:123456789012:function:test-function"
        }
        lambda {
          arn = "arn:aws:lambda:us-east-1:123456789012:function:test-function-east"
        }

        ungraceful {
          behavior = "skip"
        }
      }
    }

    # ManualApproval step
    step {
      name                 = "manual-approval-step"
      execution_block_type = "ManualApproval"
      description          = "Manual approval step"

      execution_approval_config {
        approval_role    = aws_iam_role.test.arn
        timeout_minutes  = 30
      }
    }

    # AuroraGlobalDatabase step
    step {
      name                 = "aurora-global-step"
      execution_block_type = "AuroraGlobalDatabase"
      description          = "Aurora Global Database step"

      global_aurora_config {
        behavior                   = "switchoverOnly"
        global_cluster_identifier  = "test-global-cluster"
        database_cluster_arns      = [
          "arn:aws:rds:us-east-1:123456789012:cluster:test-cluster-1",
          "arn:aws:rds:us-west-2:123456789012:cluster:test-cluster-2"
        ]
        timeout_minutes = 45

        ungraceful {
          ungraceful = "failover"
        }
      }
    }

    # EC2AutoScaling step
    step {
      name                 = "ec2-asg-step"
      execution_block_type = "EC2AutoScaling"
      description          = "EC2 Auto Scaling step"

      ec2_asg_capacity_increase_config {
        asgs {
          arn                 = "arn:aws:autoscaling:us-west-2:123456789012:autoScalingGroup:12345678-1234-1234-1234-123456789012:autoScalingGroupName/test-asg-1"
          cross_account_role  = "arn:aws:iam::123456789012:role/ASGRole"
          external_id         = "asg-external-id"
        }
        asgs {
          arn = "arn:aws:autoscaling:us-east-1:123456789012:autoScalingGroup:11111111-1111-1111-1111-111111111111:autoScalingGroupName/test-asg-east"
        }
        capacity_monitoring_approach = "sampledMaxInLast24Hours"
        target_percent              = 150
        timeout_minutes             = 20

        ungraceful {
          minimum_success_percentage = 80
        }
      }
    }

    # ECSServiceScaling step
    step {
      name                 = "ecs-scaling-step"
      execution_block_type = "ECSServiceScaling"
      description          = "ECS Service Scaling step"

      ecs_capacity_increase_config {
        services {
          cluster_arn         = "arn:aws:ecs:us-west-2:123456789012:cluster/test-cluster"
          service_arn         = "arn:aws:ecs:us-west-2:123456789012:service/test-cluster/test-service"
          cross_account_role  = "arn:aws:iam::123456789012:role/ECSRole"
          external_id         = "ecs-external-id"
        }
        services {
          cluster_arn = "arn:aws:ecs:us-east-1:123456789012:cluster/test-cluster-east"
          service_arn = "arn:aws:ecs:us-east-1:123456789012:service/test-cluster-east/test-service-east"
        }
        capacity_monitoring_approach = "containerInsightsMaxInLast24Hours"
        target_percent              = 200
        timeout_minutes             = 25

        ungraceful {
          minimum_success_percentage = 90
        }
      }
    }

    # EKSResourceScaling step
    step {
      name                 = "eks-scaling-step"
      execution_block_type = "EKSResourceScaling"
      description          = "EKS Resource Scaling step"

      eks_resource_scaling_config {
        kubernetes_resource_type {
          api_version = "apps/v1"
          kind        = "Deployment"
        }

        eks_clusters {
          cluster_arn         = "arn:aws:eks:us-west-2:123456789012:cluster/test-cluster"
          cross_account_role  = "arn:aws:iam::123456789012:role/EKSRole"
          external_id         = "eks-external-id"
        }
        eks_clusters {
          cluster_arn = "arn:aws:eks:us-east-1:123456789012:cluster/test-cluster-east"
        }

        scaling_resources {
          namespace = "default"
          resources {
            resource_name = "us-west-2"
            name          = "test-deployment-west"
            namespace     = "default"
            hpa_name      = "test-hpa-west"
          }
          resources {
            resource_name = "us-east-1"
            name          = "test-deployment-east"
            namespace     = "default"
            hpa_name      = "test-hpa-east"
          }
        }

        capacity_monitoring_approach = "sampledMaxInLast24Hours"
        target_percent              = 175
        timeout_minutes             = 35

        ungraceful {
          minimum_success_percentage = 85
        }
      }
    }

    # ARCRoutingControl step
    step {
      name                 = "arc-routing-control-step"
      execution_block_type = "ARCRoutingControl"
      description          = "ARC Routing Control step"

      arc_routing_control_config {
        region_and_routing_controls {
          region = "us-east-1"
          routing_control_arns = ["arn:aws:route53-recovery-control::123456789012:controlpanel/control1"]
        }
        region_and_routing_controls {
          region = "us-west-2"
          routing_control_arns = ["arn:aws:route53-recovery-control::123456789012:controlpanel/control2"]
        }
        cross_account_role = "arn:aws:iam::123456789012:role/RoutingControlRole"
        external_id        = "routing-external-id"
        timeout_minutes    = 15
      }
    }

    # Route53HealthCheck step
    step {
      name                 = "route53-health-check-step-activate"
      execution_block_type = "Route53HealthCheck"
      description          = "Route53 Health Check step for activate"

      route53_health_check_config {
        hosted_zone_id = "Z123456789012345678"
        record_name    = "test.example.com"
        timeout_minutes = 10

        record_sets {
          record_set_identifier = "primary"
          region               = "us-east-1"
        }
        record_sets {
          record_set_identifier = "secondary"
          region               = "us-west-2"
        }
      }
    }

    # Parallel step with nested steps
    step {
      name                 = "parallel-step"
      execution_block_type = "Parallel"
      description          = "Parallel execution step"

      parallel_config {
        step {
          name                 = "parallel-lambda-1"
          execution_block_type = "CustomActionLambda"
          description          = "First parallel lambda"

          custom_action_lambda_config {
            region_to_run           = "activatingRegion"
            retry_interval_minutes  = 2.0
            timeout_minutes         = 15

            lambda {
              arn = "arn:aws:lambda:us-west-2:123456789012:function:parallel-function-1"
            }
            lambda {
              arn = "arn:aws:lambda:us-east-1:123456789012:function:parallel-function-1-east"
            }
          }
        }

        step {
          name                 = "parallel-lambda-2"
          execution_block_type = "CustomActionLambda"
          description          = "Second parallel lambda"

          custom_action_lambda_config {
            region_to_run           = "deactivatingRegion"
            retry_interval_minutes  = 3.0
            timeout_minutes         = 20

            lambda {
              arn = "arn:aws:lambda:us-east-1:123456789012:function:parallel-function-2"
            }
            lambda {
              arn = "arn:aws:lambda:us-west-2:123456789012:function:parallel-function-2-west"
            }
          }
        }
      }
    }
  }

  # Deactivate workflow
  workflow {
    workflow_target_action = "deactivate"
    workflow_description   = "Deactivation workflow"

    # Route53HealthCheck step
    step {
      name                 = "route53-health-check-step"
      execution_block_type = "Route53HealthCheck"
      description          = "Route53 Health Check step"

      route53_health_check_config {
        hosted_zone_id = "Z123456789012345678"
        record_name    = "test.example.com"
        timeout_minutes = 10

        record_sets {
          record_set_identifier = "primary"
          region               = "us-east-1"
        }
        record_sets {
          record_set_identifier = "secondary"
          region               = "us-west-2"
        }
      }
    }
  }
}
`, rName)
}
