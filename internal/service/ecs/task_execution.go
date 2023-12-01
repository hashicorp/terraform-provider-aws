package ecs

import (
	"context"
	"log"

	"github.com/aws/aws-sdk-go/service/ecs"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
)

// @SDKResource("aws_ecs_task_execution")
func ResourceTaskExecution() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceTaskExecutionCreate,
		DeleteWithoutTimeout: resourceTaskExecutionDelete,
		ReadWithoutTimeout:   resourceTaskExecutionRead,

		Schema: map[string]*schema.Schema{
			"capacity_provider_strategy": {
				Type:     schema.TypeSet,
				Optional: true,
				ForceNew: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"base": {
							Type:         schema.TypeInt,
							Optional:     true,
							ForceNew:     true,
							ValidateFunc: validation.IntBetween(0, 100000),
						},
						"capacity_provider": {
							Type:     schema.TypeString,
							Required: true,
							ForceNew: true,
						},
						"weight": {
							Type:         schema.TypeInt,
							Optional:     true,
							ForceNew:     true,
							ValidateFunc: validation.IntBetween(0, 1000),
						},
					},
				},
			},
			"cluster": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"desired_count": {
				Type:         schema.TypeInt,
				Optional:     true,
				ForceNew:     true,
				ValidateFunc: validation.IntBetween(0, 10),
			},
			"enable_ecs_managed_tags": {
				Type:     schema.TypeBool,
				Optional: true,
				ForceNew: true,
			},
			"enable_execute_command": {
				Type:     schema.TypeBool,
				Optional: true,
				ForceNew: true,
			},
			"group": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
			},
			"launch_type": {
				Type:         schema.TypeString,
				Optional:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringInSlice(ecs.LaunchType_Values(), false),
			},
			"network_configuration": {
				Type:     schema.TypeList,
				Optional: true,
				ForceNew: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"security_groups": {
							Type:     schema.TypeSet,
							Optional: true,
							ForceNew: true,
							Elem:     &schema.Schema{Type: schema.TypeString},
							Set:      schema.HashString,
						},
						"subnets": {
							Type:     schema.TypeSet,
							Required: true,
							ForceNew: true,
							Elem:     &schema.Schema{Type: schema.TypeString},
							Set:      schema.HashString,
						},
						"assign_public_ip": {
							Type:     schema.TypeBool,
							Optional: true,
							ForceNew: true,
							Default:  false,
						},
					},
				},
			},
			"overrides": {
				Type:     schema.TypeList,
				Optional: true,
				ForceNew: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"container_overrides": {
							Type:     schema.TypeList,
							Optional: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"command": {
										Type:     schema.TypeList,
										Optional: true,
										ForceNew: true,
										Elem:     &schema.Schema{Type: schema.TypeString},
									},
									"cpu": {
										Type:     schema.TypeInt,
										Optional: true,
										ForceNew: true,
									},
									"environment": {
										Type:     schema.TypeSet,
										Optional: true,
										ForceNew: true,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"key": {
													Type:     schema.TypeString,
													Required: true,
													ForceNew: true,
												},
												"value": {
													Type:     schema.TypeString,
													Required: true,
													ForceNew: true,
												},
											},
										},
									},
									"memory": {
										Type:     schema.TypeInt,
										Optional: true,
										ForceNew: true,
									},
									"memory_reservation": {
										Type:     schema.TypeInt,
										Optional: true,
										ForceNew: true,
									},
									"name": {
										Type:     schema.TypeString,
										Required: true,
										ForceNew: true,
									},
									"resource_requirements": {
										Type:     schema.TypeSet,
										Optional: true,
										ForceNew: true,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"type": {
													Type:         schema.TypeString,
													Required:     true,
													ForceNew:     true,
													ValidateFunc: validation.StringInSlice(ecs.ResourceType_Values(), false),
												},
												"value": {
													Type:     schema.TypeString,
													Required: true,
													ForceNew: true,
												},
											},
										},
									},
								},
							},
						},
						"cpu": {
							Type:     schema.TypeString,
							Optional: true,
							ForceNew: true,
						},
						"execution_role_arn": {
							Type:     schema.TypeString,
							Optional: true,
							ForceNew: true,
						},
						"inference_accelerator_overrides": {
							Type:     schema.TypeSet,
							Optional: true,
							ForceNew: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"device_name": {
										Type:     schema.TypeString,
										Optional: true,
										ForceNew: true,
									},
									"device_type": {
										Type:     schema.TypeString,
										Optional: true,
										ForceNew: true,
									},
								},
							},
						},
						"memory": {
							Type:     schema.TypeString,
							Optional: true,
							ForceNew: true,
						},
						"task_role_arn": {
							Type:     schema.TypeString,
							Optional: true,
							ForceNew: true,
						},
					},
				},
			},
			"placement_constraints": {
				Type:     schema.TypeSet,
				Optional: true,
				ForceNew: true,
				MaxItems: 10,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"expression": {
							Type:     schema.TypeString,
							Optional: true,
							ForceNew: true,
						},
						"type": {
							Type:         schema.TypeString,
							Required:     true,
							ForceNew:     true,
							ValidateFunc: validation.StringInSlice(ecs.PlacementConstraintType_Values(), false),
						},
					},
				},
			},
			"placement_strategy": {
				Type:     schema.TypeList,
				Optional: true,
				ForceNew: true,
				MaxItems: 5,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"field": {
							Type:     schema.TypeString,
							Optional: true,
							ForceNew: true,
						},
						"type": {
							Type:     schema.TypeString,
							Required: true,
							ForceNew: true,
						},
					},
				},
			},
			"platform_version": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
			},
			"propagate_tags": {
				Type:         schema.TypeString,
				Optional:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringInSlice(ecs.PropagateTags_Values(), false),
			},
			"reference_id": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
			},
			"started_by": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
			},
			"tags": tftags.TagsSchemaForceNew(),
			"task_arns": {
				Type:     schema.TypeList,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"task_definition": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
		},
	}
}

const (
	ResNameTaskExecution = "Task Execution Resource"
)

func resourceTaskExecutionCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	return ecsRunTask(ResNameTaskExecution, ctx, d, meta)
}

func resourceTaskExecutionDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	log.Printf("[DEBUG] %s (%s) \"deleted\" by removing from state", ResNameTaskExecution, d.Id())
	return diags
}

func resourceTaskExecutionRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	return diags
}
