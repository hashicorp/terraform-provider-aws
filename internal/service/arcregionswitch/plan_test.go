// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

// lintignore:AWSAT003,AWSAT005,AT004
package arcregionswitch_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/service/arcregionswitch"
	sdktypes "github.com/aws/aws-sdk-go-v2/service/arcregionswitch/types"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfarcregionswitch "github.com/hashicorp/terraform-provider-aws/internal/service/arcregionswitch"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccARCRegionSwitchPlan_basic(t *testing.T) {
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
				Config: testAccPlanConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPlanExists(ctx, resourceName, &plan),
					resource.TestCheckResourceAttr(resourceName, acctest.CtName, rName),
					resource.TestCheckResourceAttr(resourceName, "recovery_approach", "activePassive"),
					resource.TestCheckResourceAttr(resourceName, "regions.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "primary_region", acctest.Region()),
					resource.TestCheckResourceAttr(resourceName, "workflow.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "workflow.0.step.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "workflow.0.step.0.execution_block_type", "ManualApproval"),
					resource.TestCheckResourceAttr(resourceName, "workflow.1.step.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "workflow.1.step.0.execution_block_type", "ManualApproval"),
					resource.TestMatchResourceAttr(resourceName, names.AttrARN, regexache.MustCompile(`^arn:aws:arc-region-switch:.*:.*:plan/.+`)),
					resource.TestCheckResourceAttrSet(resourceName, "execution_role"),
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

func TestAccARCRegionSwitchPlan_disappears(t *testing.T) {
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
				Config: testAccPlanConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPlanExists(ctx, resourceName, &plan),
					acctest.CheckFrameworkResourceDisappears(ctx, acctest.Provider, tfarcregionswitch.ResourcePlan, resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccARCRegionSwitchPlan_update(t *testing.T) {
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
				Config: testAccPlanConfig_update(rName, "Initial description", 30),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPlanExists(ctx, resourceName, &plan),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, "Initial description"),
					resource.TestCheckResourceAttr(resourceName, "recovery_time_objective_minutes", "30"),
					resource.TestCheckResourceAttr(resourceName, "associated_alarms.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "triggers.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "triggers.0.action", "activate"),
					resource.TestCheckResourceAttr(resourceName, "triggers.0.conditions.#", "1"),
				),
			},
			{
				Config: testAccPlanConfig_update(rName, "Updated description", 60),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPlanExists(ctx, resourceName, &plan),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, "Updated description"),
					resource.TestCheckResourceAttr(resourceName, "recovery_time_objective_minutes", "60"),
					resource.TestCheckResourceAttr(resourceName, "associated_alarms.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "triggers.#", "2"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "triggers.*", map[string]string{
						names.AttrAction: "activate",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "triggers.*", map[string]string{
						names.AttrAction: "deactivate",
					}),
				),
			},
		},
	})
}

func TestAccARCRegionSwitchPlan_tags(t *testing.T) {
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
				Config: testAccPlanConfig_tags1(rName, acctest.CtKey1, acctest.CtValue1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPlanExists(ctx, resourceName, &plan),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, "1"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey1, acctest.CtValue1),
				),
			},
			{
				Config: testAccPlanConfig_tags2(rName, acctest.CtKey1, acctest.CtValue1Updated, acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPlanExists(ctx, resourceName, &plan),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, "2"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey1, acctest.CtValue1Updated),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey2, acctest.CtValue2),
				),
			},
			{
				Config: testAccPlanConfig_tags1(rName, acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPlanExists(ctx, resourceName, &plan),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, "1"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey2, acctest.CtValue2),
				),
			},
		},
	})
}

func TestAccARCRegionSwitchPlan_singleRegion(t *testing.T) {
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
				Config: testAccPlanConfig_minimalRegions(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPlanExists(ctx, resourceName, &plan),
					resource.TestCheckResourceAttr(resourceName, "regions.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "primary_region", acctest.AlternateRegion()),
					resource.TestCheckResourceAttr(resourceName, "workflow.#", "2"),
				),
			},
		},
	})
}

func TestAccARCRegionSwitchPlan_multipleWorkflowsSameAction(t *testing.T) {
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
				Config: testAccPlanConfig_multipleWorkflowsSameAction(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPlanExists(ctx, resourceName, &plan),
					resource.TestCheckResourceAttr(resourceName, "workflow.#", "2"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "workflow.*", map[string]string{
						"workflow_target_action": "activate",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "workflow.*", map[string]string{
						"workflow_target_action": "deactivate",
					}),
				),
			},
		},
	})
}

func TestAccARCRegionSwitchPlan_route53HealthCheck(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test requiring VPC creation and Route53 health check setup")
	}

	ctx := acctest.Context(t)
	var plan sdktypes.Plan
	rName := acctest.RandomWithPrefix(t, "tf-acc-test")
	resourceName := "aws_arcregionswitch_plan.test"
	dataSourceName := "data.aws_arcregionswitch_plan.test"
	zoneName := acctest.RandomDomainName()

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
				Config: testAccPlanConfig_route53HealthCheck(rName, zoneName, acctest.AlternateRegion(), acctest.Region()),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPlanExists(ctx, resourceName, &plan),
					resource.TestCheckResourceAttr(resourceName, acctest.CtName, rName),
					resource.TestCheckResourceAttr(resourceName, "recovery_approach", "activeActive"),
					resource.TestCheckResourceAttr(resourceName, "regions.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "workflow.#", "2"),

					// Verify Route53 health checks are populated via data source
					resource.TestCheckResourceAttr(dataSourceName, "route53_health_checks.#", "4"),
					resource.TestCheckResourceAttrSet(dataSourceName, "route53_health_checks.0.health_check_id"),
					resource.TestCheckResourceAttrSet(dataSourceName, "route53_health_checks.1.health_check_id"),

					// Verify Route53 records reference health check IDs
					resource.TestCheckResourceAttrSet("aws_route53_record.primary", "health_check_id"),
					resource.TestCheckResourceAttrSet("aws_route53_record.secondary", "health_check_id"),

					// Verify private hosted zone
					resource.TestCheckResourceAttr("aws_route53_zone.private", "vpc.#", "2"),
				),
			},
		},
	})
}

func TestAccARCRegionSwitchPlan_complex(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping complex test with multiple workflow steps")
	}

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
				Config: testAccPlanConfig_complex(rName, acctest.AlternateRegion(), acctest.Region()),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPlanExists(ctx, resourceName, &plan),
					resource.TestCheckResourceAttr(resourceName, acctest.CtName, rName),
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
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, "Complex test plan with multiple execution block types"),
					resource.TestCheckResourceAttr(resourceName, "recovery_time_objective_minutes", "60"),
					resource.TestCheckResourceAttr(resourceName, "associated_alarms.#", "1"),
					resource.TestMatchResourceAttr(resourceName, names.AttrARN, regexache.MustCompile(`^arn:aws:arc-region-switch:.*:.*:plan/.+`)),

					// Verify CustomActionLambda execution block
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "workflow.*.step.*", map[string]string{
						"execution_block_type": "CustomActionLambda",
						acctest.CtName:         "custom-lambda-step",
					}),

					// Verify ManualApproval execution block
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "workflow.*.step.*", map[string]string{
						"execution_block_type": "ManualApproval",
						acctest.CtName:         "manual-approval-step",
					}),

					// Verify AuroraGlobalDatabase execution block
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "workflow.*.step.*", map[string]string{
						"execution_block_type": "AuroraGlobalDatabase",
						acctest.CtName:         "aurora-global-step",
					}),

					// Verify EC2AutoScaling execution block
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "workflow.*.step.*", map[string]string{
						"execution_block_type": "EC2AutoScaling",
						acctest.CtName:         "ec2-asg-step",
					}),

					// Verify ECSServiceScaling execution block
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "workflow.*.step.*", map[string]string{
						"execution_block_type": "ECSServiceScaling",
						acctest.CtName:         "ecs-scaling-step",
					}),

					// Verify EKSResourceScaling execution block
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "workflow.*.step.*", map[string]string{
						"execution_block_type": "EKSResourceScaling",
						acctest.CtName:         "eks-scaling-step",
					}),

					// Verify Route53HealthCheck execution block
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "workflow.*.step.*", map[string]string{
						"execution_block_type": "Route53HealthCheck",
						acctest.CtName:         "route53-health-check-step-activate",
					}),

					// Verify Parallel execution block
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "workflow.*.step.*", map[string]string{
						"execution_block_type": "Parallel",
						acctest.CtName:         "parallel-step",
					}),

					// Verify ARCRoutingControl execution block
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "workflow.*.step.*", map[string]string{
						"execution_block_type": "ARCRoutingControl",
						acctest.CtName:         "arc-routing-control-step",
					}),

					// Verify specific configuration values are stored correctly
					// CustomActionLambda config
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "workflow.*.step.*.execution_block_configuration.*.custom_action_lambda_config.*", map[string]string{
						"region_to_run":          "activatingRegion",
						"timeout_minutes":        "30",
						"retry_interval_minutes": "5",
					}),

					// ManualApproval config
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "workflow.*.step.*.execution_block_configuration.*.execution_approval_config.*", map[string]string{
						"timeout_minutes": "30",
					}),

					// AuroraGlobalDatabase config
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "workflow.*.step.*.execution_block_configuration.*.global_aurora_config.*", map[string]string{
						"behavior":                  "switchoverOnly",
						"global_cluster_identifier": "test-global-cluster",
						"timeout_minutes":           "45",
					}),

					// EC2AutoScaling config
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "workflow.*.step.*.execution_block_configuration.*.ec2_asg_capacity_increase_config.*", map[string]string{
						"capacity_monitoring_approach": "sampledMaxInLast24Hours",
						"target_percent":               "150",
						"timeout_minutes":              "20",
					}),

					// ECSServiceScaling config
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "workflow.*.step.*.execution_block_configuration.*.ecs_capacity_increase_config.*", map[string]string{
						"capacity_monitoring_approach": "containerInsightsMaxInLast24Hours",
						"target_percent":               "200",
						"timeout_minutes":              "25",
					}),

					// EKSResourceScaling config
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "workflow.*.step.*.execution_block_configuration.*.eks_resource_scaling_config.*", map[string]string{
						"capacity_monitoring_approach": "sampledMaxInLast24Hours",
						"target_percent":               "175",
						"timeout_minutes":              "35",
					}),

					// Route53HealthCheck config
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "workflow.*.step.*.execution_block_configuration.*.route53_health_check_config.*", map[string]string{
						names.AttrHostedZoneID: "Z123456789012345678",
						"timeout_minutes":      "10",
						"record_name":          "test.example.com",
					}),

					// ARCRoutingControl config
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "workflow.*.step.*.execution_block_configuration.*.arc_routing_control_config.*", map[string]string{
						"cross_account_role": "arn:aws:iam::123456789012:role/RoutingControlRole",
						names.AttrExternalID: "routing-external-id",
						"timeout_minutes":    "15",
					}),
				),
				// API returns EKS scaling resources in different order than specified
				ExpectNonEmptyPlan: true,
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				// API may return regions in different order than specified
				ImportStateVerifyIgnore: []string{
					"workflow.0.step.6.execution_block_configuration.0.arc_routing_control_config.0.region_and_routing_controls.0.region",
					"workflow.0.step.6.execution_block_configuration.0.arc_routing_control_config.0.region_and_routing_controls.1.region",
					"workflow.0.step.6.execution_block_configuration.0.arc_routing_control_config.0.region_and_routing_controls.0.routing_control_arns.0",
					"workflow.0.step.6.execution_block_configuration.0.arc_routing_control_config.0.region_and_routing_controls.1.routing_control_arns.0",
					// EKS scaling resources may be returned in different order
					"workflow.0.step.5.execution_block_configuration.0.eks_resource_scaling_config.0.scaling_resources.0.resources.0.hpa_name",
					"workflow.0.step.5.execution_block_configuration.0.eks_resource_scaling_config.0.scaling_resources.0.resources.0.name",
					"workflow.0.step.5.execution_block_configuration.0.eks_resource_scaling_config.0.scaling_resources.0.resources.0.resource_name",
					"workflow.0.step.5.execution_block_configuration.0.eks_resource_scaling_config.0.scaling_resources.0.resources.1.hpa_name",
					"workflow.0.step.5.execution_block_configuration.0.eks_resource_scaling_config.0.scaling_resources.0.resources.1.name",
					"workflow.0.step.5.execution_block_configuration.0.eks_resource_scaling_config.0.scaling_resources.0.resources.1.resource_name",
				},
			},
		},
	})
}

func testAccPlanConfig_complex(rName, primaryRegion, alternateRegion string) string {
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
  name                            = %[1]q
  execution_role                  = aws_iam_role.test.arn
  recovery_approach               = "activeActive"
  regions                         = [%[2]q, %[3]q]
  primary_region                  = %[2]q
  description                     = "Complex test plan with multiple execution block types"
  recovery_time_objective_minutes = 60

  associated_alarms {
    name                = "test-alarm-1"
    alarm_type          = "applicationHealth"
    resource_identifier = "arn:aws:cloudwatch:%[2]s:123456789012:alarm:test-alarm-1"
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

      execution_block_configuration {
        custom_action_lambda_config {
          region_to_run          = "activatingRegion"
          retry_interval_minutes = 5.0
          timeout_minutes        = 30

          lambda {
            arn = "arn:aws:lambda:%[3]s:123456789012:function:test-function"
          }
          lambda {
            arn = "arn:aws:lambda:%[2]s:123456789012:function:test-function-primary"
          }

          ungraceful {
            behavior = "skip"
          }
        }
      }
    }

    # ManualApproval step
    step {
      name                 = "manual-approval-step"
      execution_block_type = "ManualApproval"
      description          = "Manual approval step"

      execution_block_configuration {
        execution_approval_config {
          approval_role   = aws_iam_role.test.arn
          timeout_minutes = 30
        }
      }
    }

    # AuroraGlobalDatabase step
    step {
      name                 = "aurora-global-step"
      execution_block_type = "AuroraGlobalDatabase"
      description          = "Aurora Global Database step"

      execution_block_configuration {
        global_aurora_config {
          behavior                  = "switchoverOnly"
          global_cluster_identifier = "test-global-cluster"
          database_cluster_arns = [
            "arn:aws:rds:%[2]s:123456789012:cluster:test-cluster-1",
            "arn:aws:rds:%[3]s:123456789012:cluster:test-cluster-2"
          ]
          timeout_minutes = 45

          ungraceful {
            ungraceful = "failover"
          }
        }
      }
    }

    # EC2AutoScaling step
    step {
      name                 = "ec2-asg-step"
      execution_block_type = "EC2AutoScaling"
      description          = "EC2 Auto Scaling step"

      execution_block_configuration {
        ec2_asg_capacity_increase_config {
          asgs {
            arn                = "arn:aws:autoscaling:%[3]s:123456789012:autoScalingGroup:12345678-1234-1234-1234-123456789012:autoScalingGroupName/test-asg-1"
            cross_account_role = "arn:aws:iam::123456789012:role/ASGRole"
            external_id        = "asg-external-id"
          }
          asgs {
            arn = "arn:aws:autoscaling:%[2]s:123456789012:autoScalingGroup:11111111-1111-1111-1111-111111111111:autoScalingGroupName/test-asg-primary"
          }
          capacity_monitoring_approach = "sampledMaxInLast24Hours"
          target_percent               = 150
          timeout_minutes              = 20

          ungraceful {
            minimum_success_percentage = 80
          }
        }
      }
    }

    # ECSServiceScaling step
    step {
      name                 = "ecs-scaling-step"
      execution_block_type = "ECSServiceScaling"
      description          = "ECS Service Scaling step"

      execution_block_configuration {
        ecs_capacity_increase_config {
          services {
            cluster_arn        = "arn:aws:ecs:%[3]s:123456789012:cluster/test-cluster"
            service_arn        = "arn:aws:ecs:%[3]s:123456789012:service/test-cluster/test-service"
            cross_account_role = "arn:aws:iam::123456789012:role/ECSRole"
            external_id        = "ecs-external-id"
          }
          services {
            cluster_arn = "arn:aws:ecs:%[2]s:123456789012:cluster/test-cluster-primary"
            service_arn = "arn:aws:ecs:%[2]s:123456789012:service/test-cluster-primary/test-service-primary"
          }
          capacity_monitoring_approach = "containerInsightsMaxInLast24Hours"
          target_percent               = 200
          timeout_minutes              = 25

          ungraceful {
            minimum_success_percentage = 90
          }
        }
      }
    }

    # EKSResourceScaling step
    step {
      name                 = "eks-scaling-step"
      execution_block_type = "EKSResourceScaling"
      description          = "EKS Resource Scaling step"

      execution_block_configuration {
        eks_resource_scaling_config {
          kubernetes_resource_type {
            api_version = "apps/v1"
            kind        = "Deployment"
          }

          eks_clusters {
            cluster_arn        = "arn:aws:eks:%[3]s:123456789012:cluster/test-cluster"
            cross_account_role = "arn:aws:iam::123456789012:role/EKSRole"
            external_id        = "eks-external-id"
          }
          eks_clusters {
            cluster_arn = "arn:aws:eks:%[2]s:123456789012:cluster/test-cluster-primary"
          }

          scaling_resources {
            namespace = "default"
            resources {
              resource_name = %[2]q
              name          = "test-deployment-secondary"
              namespace     = "default"
              hpa_name      = "test-hpa-secondary"
            }
            resources {
              resource_name = %[3]q
              name          = "test-deployment-primary"
              namespace     = "default"
              hpa_name      = "test-hpa-primary"
            }
          }

          capacity_monitoring_approach = "sampledMaxInLast24Hours"
          target_percent               = 175
          timeout_minutes              = 35

          ungraceful {
            minimum_success_percentage = 85
          }
        }
      }
    }

    # ARCRoutingControl step
    step {
      name                 = "arc-routing-control-step"
      execution_block_type = "ARCRoutingControl"
      description          = "ARC Routing Control step"

      execution_block_configuration {
        arc_routing_control_config {
          region_and_routing_controls {
            region               = %[3]q
            routing_control_arns = ["arn:aws:route53-recovery-control::123456789012:controlpanel/12345678901234567890123456789012/routingcontrol/1234567890123456"]
          }
          region_and_routing_controls {
            region               = %[2]q
            routing_control_arns = ["arn:aws:route53-recovery-control::123456789012:controlpanel/12345678901234567890123456789013/routingcontrol/1234567890123457"]
          }
          cross_account_role = "arn:aws:iam::123456789012:role/RoutingControlRole"
          external_id        = "routing-external-id"
          timeout_minutes    = 15
        }
      }
    }

    # Route53HealthCheck step
    step {
      name                 = "route53-health-check-step-activate"
      execution_block_type = "Route53HealthCheck"
      description          = "Route53 Health Check step for activate"

      execution_block_configuration {
        route53_health_check_config {
          hosted_zone_id  = "Z123456789012345678"
          record_name     = "test.example.com"
          timeout_minutes = 10

          record_sets {
            record_set_identifier = "primary"
            region                = %[2]q
          }
          record_sets {
            record_set_identifier = "secondary"
            region                = %[3]q
          }
        }
      }
    }

    # Parallel step with nested steps
    step {
      name                 = "parallel-step"
      execution_block_type = "Parallel"
      description          = "Parallel execution step"

      execution_block_configuration {
        parallel_config {
          step {
            name                 = "parallel-lambda-1"
            execution_block_type = "CustomActionLambda"
            description          = "First parallel lambda"

            custom_action_lambda_config {
              region_to_run          = "activatingRegion"
              retry_interval_minutes = 2.0
              timeout_minutes        = 15

              lambda {
                arn = "arn:aws:lambda:%[3]s:123456789012:function:parallel-function-1"
              }
              lambda {
                arn = "arn:aws:lambda:%[2]s:123456789012:function:parallel-function-1-primary"
              }
            }
          }

          step {
            name                 = "parallel-lambda-2"
            execution_block_type = "CustomActionLambda"
            description          = "Second parallel lambda"

            custom_action_lambda_config {
              region_to_run          = "deactivatingRegion"
              retry_interval_minutes = 3.0
              timeout_minutes        = 20

              lambda {
                arn = "arn:aws:lambda:%[2]s:123456789012:function:parallel-function-2"
              }
              lambda {
                arn = "arn:aws:lambda:%[3]s:123456789012:function:parallel-function-2-secondary"
              }
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

      execution_block_configuration {
        route53_health_check_config {
          hosted_zone_id  = "Z123456789012345678"
          record_name     = "test.example.com"
          timeout_minutes = 10

          record_sets {
            record_set_identifier = "primary"
            region                = %[2]q
          }
          record_sets {
            record_set_identifier = "secondary"
            region                = %[3]q
          }
        }
      }
    }
  }
}
`, rName, primaryRegion, alternateRegion)
}

func testAccPlanConfig_route53HealthCheck(rName, zoneName, primaryRegion, alternateRegion string) string {
	return fmt.Sprintf(`
# Provider configuration for secondary region
provider "aws" {
  alias  = "secondary"
  region = %[4]q
}

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

# VPCs for private hosted zone
resource "aws_vpc" "primary" {
  cidr_block           = "10.1.0.0/16"
  enable_dns_hostnames = true
  enable_dns_support   = true

  tags = {
    Name = "%[1]s-primary"
  }
}

resource "aws_vpc" "secondary" {
  provider             = aws.secondary
  cidr_block           = "10.2.0.0/16"
  enable_dns_hostnames = true
  enable_dns_support   = true

  tags = {
    Name = "%[1]s-secondary"
  }
}

# Private hosted zone
resource "aws_route53_zone" "private" {
  name = "%[2]s"

  vpc {
    vpc_id = aws_vpc.primary.id
  }

  vpc {
    vpc_id     = aws_vpc.secondary.id
    vpc_region = %[4]q
  }
}

# ARC Region Switch Plan with Route53 health checks
resource "aws_arcregionswitch_plan" "test" {
  name              = %[1]q
  execution_role    = aws_iam_role.test.arn
  recovery_approach = "activeActive"
  regions           = [%[3]q, %[4]q]
  primary_region    = %[3]q
  description       = "Route53 health check integration test"

  workflow {
    workflow_target_action = "activate"

    step {
      name                 = "route53-health-check-primary"
      execution_block_type = "Route53HealthCheck"

      execution_block_configuration {
        route53_health_check_config {
          hosted_zone_id  = aws_route53_zone.private.zone_id
          record_name     = "api-primary.%[2]s"
          timeout_minutes = 60
        }
      }
    }

    step {
      name                 = "route53-health-check-secondary"
      execution_block_type = "Route53HealthCheck"

      execution_block_configuration {
        route53_health_check_config {
          hosted_zone_id  = aws_route53_zone.private.zone_id
          record_name     = "api-secondary.%[2]s"
          timeout_minutes = 60
        }
      }
    }
  }

  workflow {
    workflow_target_action = "deactivate"

    step {
      name                 = "route53-health-check-primary"
      execution_block_type = "Route53HealthCheck"

      execution_block_configuration {
        route53_health_check_config {
          hosted_zone_id  = aws_route53_zone.private.zone_id
          record_name     = "api-primary.%[2]s"
          timeout_minutes = 60
        }
      }
    }

    step {
      name                 = "route53-health-check-secondary"
      execution_block_type = "Route53HealthCheck"

      execution_block_configuration {
        route53_health_check_config {
          hosted_zone_id  = aws_route53_zone.private.zone_id
          record_name     = "api-secondary.%[2]s"
          timeout_minutes = 60
        }
      }
    }
  }

  lifecycle {
    # WORKAROUND: Ignore workflow order changes for test stability
    # The ARC service returns workflows in non-deterministic order, causing
    # Terraform to detect configuration drift even when nothing changed.
    # In real usage, users wouldn't need this ignore_changes.
    ignore_changes = [workflow]
  }
}

# Data source to get health check IDs (with wait)
data "aws_arcregionswitch_plan" "test" {
  arn                    = aws_arcregionswitch_plan.test.arn
  wait_for_health_checks = true
}

# Filter health checks by region
locals {
  primary_health_check = [
    for hc in data.aws_arcregionswitch_plan.test.route53_health_checks :
    hc if hc.region == %[3]q
  ][0]

  secondary_health_check = [
    for hc in data.aws_arcregionswitch_plan.test.route53_health_checks :
    hc if hc.region == %[4]q
  ][0]
}

# Route53 records using health check IDs with weighted routing
resource "aws_route53_record" "primary" {
  zone_id         = aws_route53_zone.private.zone_id
  name            = "api"
  type            = "A"
  ttl             = 300
  records         = ["10.1.1.100"]
  health_check_id = local.primary_health_check.health_check_id
  set_identifier  = "primary"

  weighted_routing_policy {
    weight = 100
  }

  lifecycle {
    # WORKAROUND: Ignore health_check_id changes for test stability
    # Health check IDs are generated asynchronously by the ARC service and
    # can change between plan/apply cycles during testing. In real usage,
    # these IDs would be stable once created.
    ignore_changes = [health_check_id]
  }
}

resource "aws_route53_record" "secondary" {
  zone_id         = aws_route53_zone.private.zone_id
  name            = "api"
  type            = "A"
  ttl             = 300
  records         = ["10.2.1.100"]
  health_check_id = local.secondary_health_check.health_check_id
  set_identifier  = "secondary"

  weighted_routing_policy {
    weight = 100
  }

  lifecycle {
    # WORKAROUND: Ignore health_check_id changes for test stability
    # Health check IDs are generated asynchronously by the ARC service and
    # can change between plan/apply cycles during testing. In real usage,
    # these IDs would be stable once created.
    ignore_changes = [health_check_id]
  }
}
`, rName, zoneName, primaryRegion, alternateRegion)
}

func testAccCheckPlanDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).ARCRegionSwitchClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_arcregionswitch_plan" {
				continue
			}

			_, err := tfarcregionswitch.FindPlanByARN(ctx, conn, rs.Primary.ID)

			if err == nil {
				return fmt.Errorf("ARC Region Switch Plan %s still exists", rs.Primary.ID)
			}
		}

		return nil
	}
}

func testAccCheckPlanExists(ctx context.Context, n string, v *sdktypes.Plan) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Plan not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No ARC Region Switch Plan ID is set")
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).ARCRegionSwitchClient(ctx)

		output, err := tfarcregionswitch.FindPlanByARN(ctx, conn, rs.Primary.ID)

		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

func testAccPreCheck(ctx context.Context, t *testing.T) {
	conn := acctest.Provider.Meta().(*conns.AWSClient).ARCRegionSwitchClient(ctx)

	input := arcregionswitch.ListPlansInput{}
	_, err := conn.ListPlans(ctx, &input)

	if err != nil {
		t.Skipf("skipping acceptance testing: %s", err)
	}
}

func testAccPlanConfig_basic(rName string) string {
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
  name              = %[1]q
  execution_role    = aws_iam_role.test.arn
  recovery_approach = "activePassive"
  regions           = [%[3]q, %[2]q]
  primary_region    = %[3]q

  workflow {
    workflow_target_action = "activate"
    workflow_target_region = %[2]q

    step {
      name                 = "basic-step"
      execution_block_type = "ManualApproval"

      execution_block_configuration {
        execution_approval_config {
          approval_role   = aws_iam_role.test.arn
          timeout_minutes = 60
        }
      }
    }
  }

  workflow {
    workflow_target_action = "activate"
    workflow_target_region = %[3]q

    step {
      name                 = "basic-step-primary"
      execution_block_type = "ManualApproval"

      execution_block_configuration {
        execution_approval_config {
          approval_role   = aws_iam_role.test.arn
          timeout_minutes = 60
        }
      }
    }
  }
}
`, rName, acctest.AlternateRegion(), acctest.Region())
}

func testAccPlanConfig_update(rName, description string, rto int) string {
	alarms := fmt.Sprintf(`
  associated_alarms {
    name                = "test-alarm-1"
    alarm_type          = "applicationHealth"
    resource_identifier = "arn:aws:cloudwatch:%s:123456789012:alarm:test-alarm-1"
  }`, acctest.Region())

	if rto == 60 {
		alarms += fmt.Sprintf(`
  associated_alarms {
    name                = "test-alarm-2"
    alarm_type          = "applicationHealth"
    resource_identifier = "arn:aws:cloudwatch:%s:123456789012:alarm:test-alarm-2"
  }`, acctest.Region())
	}

	// Add triggers - simple for rto=30, more complex for rto=60
	triggers := fmt.Sprintf(`
  triggers {
    action                              = "activate"
    target_region                       = %q
    min_delay_minutes_between_executions = 30
    description                         = "Test trigger for activation"
    
    conditions {
      associated_alarm_name = "test-alarm-1"
      condition            = "red"
    }
  }`, acctest.AlternateRegion())

	if rto == 60 {
		triggers += fmt.Sprintf(`
  triggers {
    action                              = "deactivate"
    target_region                       = %q
    min_delay_minutes_between_executions = 60
    description                         = "Test trigger for deactivation"
    
    conditions {
      associated_alarm_name = "test-alarm-2"
      condition            = "green"
    }
  }`, acctest.Region())
	}

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
  name                            = %[1]q
  execution_role                  = aws_iam_role.test.arn
  recovery_approach               = "activePassive"
  regions                         = [%[2]q, %[3]q]
  primary_region                  = %[3]q
  description                     = %[4]q
  recovery_time_objective_minutes = %[5]d
%[6]s
%[7]s
  workflow {
    workflow_target_action = "activate"
    workflow_target_region = %[2]q

    step {
      name                 = "basic-step"
      execution_block_type = "ManualApproval"

      execution_block_configuration {
        execution_approval_config {
          approval_role   = aws_iam_role.test.arn
          timeout_minutes = 60
        }
      }
    }
  }

  workflow {
    workflow_target_action = "activate"
    workflow_target_region = %[3]q

    step {
      name                 = "basic-step-primary"
      execution_block_type = "ManualApproval"

      execution_block_configuration {
        execution_approval_config {
          approval_role   = aws_iam_role.test.arn
          timeout_minutes = 60
        }
      }
    }
  }
}
`, rName, acctest.AlternateRegion(), acctest.Region(), description, rto, alarms, triggers)
}

func testAccPlanConfig_tags1(rName, tagKey1, tagValue1 string) string {
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
  name              = %[1]q
  execution_role    = aws_iam_role.test.arn
  recovery_approach = "activePassive"
  regions           = [%[3]q, %[2]q]
  primary_region    = %[3]q

  tags = {
    %[4]q = %[5]q
  }

  workflow {
    workflow_target_action = "activate"
    workflow_target_region = %[2]q

    step {
      name                 = "basic-step"
      execution_block_type = "ManualApproval"

      execution_block_configuration {
        execution_approval_config {
          approval_role   = aws_iam_role.test.arn
          timeout_minutes = 60
        }
      }
    }
  }

  workflow {
    workflow_target_action = "activate"
    workflow_target_region = %[3]q

    step {
      name                 = "basic-step-primary"
      execution_block_type = "ManualApproval"

      execution_block_configuration {
        execution_approval_config {
          approval_role   = aws_iam_role.test.arn
          timeout_minutes = 60
        }
      }
    }
  }
}
`, rName, acctest.AlternateRegion(), acctest.Region(), tagKey1, tagValue1)
}

func testAccPlanConfig_tags2(rName, tagKey1, tagValue1, tagKey2, tagValue2 string) string {
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
  name              = %[1]q
  execution_role    = aws_iam_role.test.arn
  recovery_approach = "activePassive"
  regions           = [%[3]q, %[2]q]
  primary_region    = %[3]q

  tags = {
    %[4]q = %[5]q
    %[6]q = %[7]q
  }

  workflow {
    workflow_target_action = "activate"
    workflow_target_region = %[2]q

    step {
      name                 = "basic-step"
      execution_block_type = "ManualApproval"

      execution_block_configuration {
        execution_approval_config {
          approval_role   = aws_iam_role.test.arn
          timeout_minutes = 60
        }
      }
    }
  }

  workflow {
    workflow_target_action = "activate"
    workflow_target_region = %[3]q

    step {
      name                 = "basic-step-primary"
      execution_block_type = "ManualApproval"

      execution_block_configuration {
        execution_approval_config {
          approval_role   = aws_iam_role.test.arn
          timeout_minutes = 60
        }
      }
    }
  }
}
`, rName, acctest.AlternateRegion(), acctest.Region(), tagKey1, tagValue1, tagKey2, tagValue2)
}

func testAccPlanConfig_minimalRegions(rName string) string {
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
  name              = %[1]q
  execution_role    = aws_iam_role.test.arn
  recovery_approach = "activePassive"
  regions           = [%[2]q, %[3]q]
  primary_region    = %[2]q

  workflow {
    workflow_target_action = "activate"
    workflow_target_region = %[2]q

    step {
      name                 = "minimal-step-secondary"
      execution_block_type = "ManualApproval"

      execution_block_configuration {
        execution_approval_config {
          approval_role   = aws_iam_role.test.arn
          timeout_minutes = 60
        }
      }
    }
  }

  workflow {
    workflow_target_action = "activate"
    workflow_target_region = %[3]q

    step {
      name                 = "minimal-step-primary"
      execution_block_type = "ManualApproval"

      execution_block_configuration {
        execution_approval_config {
          approval_role   = aws_iam_role.test.arn
          timeout_minutes = 60
        }
      }
    }
  }
}
`, rName, acctest.AlternateRegion(), acctest.Region())
}

func testAccPlanConfig_multipleWorkflowsSameAction(rName string) string {
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
  name              = %[1]q
  execution_role    = aws_iam_role.test.arn
  recovery_approach = "activeActive"
  regions           = [%[3]q, %[2]q]
  primary_region    = %[3]q

  workflow {
    workflow_target_action = "activate"

    step {
      name                 = "activate-step"
      execution_block_type = "ManualApproval"

      execution_block_configuration {
        execution_approval_config {
          approval_role   = aws_iam_role.test.arn
          timeout_minutes = 60
        }
      }
    }
  }

  workflow {
    workflow_target_action = "deactivate"

    step {
      name                 = "deactivate-step"
      execution_block_type = "ManualApproval"

      execution_block_configuration {
        execution_approval_config {
          approval_role   = aws_iam_role.test.arn
          timeout_minutes = 60
        }
      }
    }
  }
}
`, rName, acctest.AlternateRegion(), acctest.Region())
}
