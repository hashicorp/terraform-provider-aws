package scheduler

import (
	"context"
	"errors"
	"fmt"
	"log"
	"regexp"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/scheduler"
	"github.com/aws/aws-sdk-go-v2/service/scheduler/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func ResourceSchedule() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceScheduleCreate,
		ReadWithoutTimeout:   resourceScheduleRead,
		UpdateWithoutTimeout: resourceScheduleUpdate,
		DeleteWithoutTimeout: resourceScheduleDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			"name": {
				Type:          schema.TypeString,
				Optional:      true,
				Computed:      true,
				ForceNew:      true,
				ConflictsWith: []string{"name_prefix"},
				ValidateDiagFunc: validation.ToDiagFunc(validation.All(
					validation.StringLenBetween(1, 64),
					validation.StringMatch(regexp.MustCompile(`^[0-9a-zA-Z-_.]+$`), `The name must consist of alphanumerics, hyphens, underscores and dot(.).`),
				)),
			},
			"name_prefix": {
				Type:          schema.TypeString,
				Optional:      true,
				Computed:      true,
				ForceNew:      true,
				ConflictsWith: []string{"name"},
				ValidateDiagFunc: validation.ToDiagFunc(validation.All(
					validation.StringLenBetween(1, 64-resource.UniqueIDSuffixLength),
					validation.StringMatch(regexp.MustCompile(`^[0-9a-zA-Z-_.]+$`), `The name must consist of alphanumerics, hyphens, and underscores.`),
				)),
			},
			"client_token": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
				ValidateDiagFunc: validation.ToDiagFunc(validation.All(
					validation.StringLenBetween(1, 64),
					validation.StringMatch(regexp.MustCompile(`^[0-9a-zA-Z-_]+$`), `must consist of alphanumerics, hyphens and underscores.`),
				)),
			},
			"description": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringLenBetween(0, 512),
			},
			"end_date": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.IsRFC3339Time,
			},
			"flexible_time_window": {
				Type:     schema.TypeList,
				Required: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"maximum_window_in_minutes": {
							Type:         schema.TypeInt,
							Optional:     true,
							ValidateFunc: validation.IntBetween(1, 1440),
						},
						"mode": {
							Type:         schema.TypeString,
							Required:     true,
							ValidateFunc: validation.StringInSlice(toStringSlice(types.FlexibleTimeWindowMode("").Values()), false),
						},
					},
				},
			},
			"group_name": {
				Type:          schema.TypeString,
				Optional:      true,
				ConflictsWith: []string{"name_prefix"},
				ValidateDiagFunc: validation.ToDiagFunc(validation.All(
					validation.StringLenBetween(1, 64),
					validation.StringMatch(regexp.MustCompile(`^[0-9a-zA-Z-_.]+$`), `The name must consist of alphanumerics, hyphens, underscores and dot(.).`),
				)),
			},
			"kms_key_arn": {
				Type:     schema.TypeString,
				Optional: true,
				ValidateDiagFunc: validation.ToDiagFunc(validation.All(
					validation.StringLenBetween(1, 2048),
					validation.StringMatch(regexp.MustCompile(`^arn:aws(-[a-z]+)?:kms:[a-z0-9\-]+:\d{12}:(key|alias)\/[0-9a-zA-Z-_]*$`), `must be arn of KMS key or alias`),
				)),
			},
			"schedule_expression": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringLenBetween(1, 256),
			},
			"schedule_expression_timezone": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringLenBetween(1, 50),
			},
			"start_date": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.IsRFC3339Time,
			},
			"state": {
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				ValidateFunc: validation.StringInSlice(toStringSlice(types.ScheduleState("").Values()), false),
			},
			"target": {
				Type:     schema.TypeList,
				Required: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"arn": {
							Type:         schema.TypeString,
							Required:     true,
							ValidateFunc: verify.ValidARN,
						},
						"dead_letter_config": {
							Type:     schema.TypeList,
							Optional: true,
							MaxItems: 1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"arn": {
										Type:         schema.TypeString,
										Optional:     true,
										ValidateFunc: verify.ValidARN,
									},
								},
							},
						},
						"input": {
							Type:         schema.TypeString,
							Optional:     true,
							ValidateFunc: validation.StringIsNotEmpty,
						},
						"retry_policy": {
							Type:     schema.TypeList,
							Optional: true,
							MaxItems: 1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"maximum_event_age_in_seconds": {
										Type:         schema.TypeInt,
										Optional:     true,
										ValidateFunc: validation.IntBetween(60, 86400),
									},
									"maximum_retry_attempts": {
										Type:         schema.TypeInt,
										Optional:     true,
										ValidateFunc: validation.IntBetween(0, 185),
									},
								},
							},
						},
						"role_arn": {
							Type:         schema.TypeString,
							Required:     true,
							ValidateFunc: verify.ValidARN,
						},

						"ecs_parameters": {
							Type:     schema.TypeList,
							Optional: true,
							MaxItems: 1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"capacity_provider_strategy": {
										Type:     schema.TypeList,
										Optional: true,
										MaxItems: 1,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"base": {
													Type:         schema.TypeInt,
													Optional:     true,
													ValidateFunc: validation.IntBetween(0, 100000),
												},
												"capacity_provider": {
													Type:     schema.TypeString,
													Required: true,
												},
												"weight": {
													Type:         schema.TypeInt,
													Optional:     true,
													ValidateFunc: validation.IntBetween(0, 1000),
												},
											},
										},
									},
									"enable_ecs_managed_tags": {
										Type:     schema.TypeBool,
										Optional: true,
										Default:  false,
									},
									"enable_execute_command": {
										Type:     schema.TypeBool,
										Optional: true,
										Default:  false,
									},
									"group": {
										Type:         schema.TypeString,
										Optional:     true,
										ValidateFunc: validation.StringLenBetween(1, 255),
									},
									"launch_type": {
										Type:         schema.TypeString,
										Optional:     true,
										ValidateFunc: validation.StringInSlice(toStringSlice(types.LaunchType("").Values()), false),
									},
									"network_configuration": {
										Type:     schema.TypeList,
										Optional: true,
										MaxItems: 1,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"assign_public_ip": {
													Type:     schema.TypeBool,
													Optional: true,
													Default:  false,
												},
												"security_groups": {
													Type:     schema.TypeSet,
													Optional: true,
													MaxItems: 5,
													Elem:     &schema.Schema{Type: schema.TypeString},
												},
												"subnets": {
													Type:     schema.TypeSet,
													Required: true,
													MaxItems: 16,
													Elem:     &schema.Schema{Type: schema.TypeString},
												},
											},
										},
									},
									"placement_constraint": {
										Type:     schema.TypeList,
										Optional: true,
										MaxItems: 10,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"expression": {
													Type:         schema.TypeString,
													Optional:     true,
													ValidateFunc: validation.StringLenBetween(0, 2000),
												},
												"type": {
													Type:         schema.TypeString,
													Required:     true,
													ValidateFunc: validation.StringInSlice(toStringSlice(types.PlacementConstraintType("").Values()), false),
												},
											},
										},
									},
									"placement_strategy": {
										Type:     schema.TypeList,
										Optional: true,
										MaxItems: 5,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"field": {
													Type:         schema.TypeString,
													Optional:     true,
													ValidateFunc: validation.StringLenBetween(0, 255),
												},
												"type": {
													Type:         schema.TypeString,
													Required:     true,
													ValidateFunc: validation.StringInSlice(toStringSlice(types.PlacementStrategyType("").Values()), false),
												},
											},
										},
									},
									"platform_version": {
										Type:         schema.TypeString,
										Optional:     true,
										ValidateFunc: validation.StringLenBetween(1, 64),
									},
									"propagate_tags": {
										Type:         schema.TypeString,
										Optional:     true,
										ValidateFunc: validation.StringInSlice(toStringSlice(types.PropagateTags("").Values()), false),
									},
									"reference_id": {
										Type:         schema.TypeString,
										Optional:     true,
										ValidateFunc: validation.StringLenBetween(0, 1024),
									},
									"tags": tftags.TagsSchema(),
									"task_count": {
										Type:         schema.TypeInt,
										Optional:     true,
										ValidateFunc: validation.IntBetween(1, 10),
										Default:      1,
									},
									"task_definition_arn": {
										Type:         schema.TypeString,
										Required:     true,
										ValidateFunc: verify.ValidARN,
									},
								},
							},
						},
						"event_bridge_parameters": {
							Type:     schema.TypeList,
							Optional: true,
							MaxItems: 1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"detail_type": {
										Type:         schema.TypeString,
										Required:     true,
										ValidateFunc: validation.StringLenBetween(1, 128),
									},
									"source": {
										Type:         schema.TypeString,
										Required:     true,
										ValidateFunc: validation.StringLenBetween(1, 256),
									},
								},
							},
						},
						"kinesis_parameters": {
							Type:     schema.TypeList,
							Optional: true,
							MaxItems: 1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"partition_key": {
										Type:         schema.TypeString,
										Required:     true,
										ValidateFunc: validation.StringLenBetween(1, 256),
									},
								},
							},
						},
						"sage_maker_pipeline_parameters": {
							Type:     schema.TypeList,
							Optional: true,
							MaxItems: 1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"sage_maker_pipeline_parameters": {
										Type:     schema.TypeList,
										Optional: true,
										MaxItems: 200,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"name": {
													Type:         schema.TypeString,
													Required:     true,
													ValidateFunc: validation.StringLenBetween(1, 256),
													ValidateDiagFunc: validation.ToDiagFunc(validation.All(
														validation.StringLenBetween(1, 256),
														validation.StringMatch(regexp.MustCompile(`^[0-9a-zA-Z-_]+$`), `must consist of alphanumerics, hyphens and underscores.`),
													)),
												},
												"value": {
													Type:         schema.TypeString,
													Required:     true,
													ValidateFunc: validation.StringLenBetween(1, 1024),
												},
											},
										},
									},
								},
							},
						},
						"sqs_parameters": {
							Type:     schema.TypeList,
							Optional: true,
							MaxItems: 1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"message_group_id": {
										Type:         schema.TypeString,
										Required:     true,
										ValidateFunc: validation.StringLenBetween(1, 128),
									},
								},
							},
						},
					},
				},
			},

			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"creation_date": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"last_modification_date": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func toStringSlice[T any](origin []T) []string {
	out := make([]string, 0, len(origin))

	for _, v := range origin {
		out = append(out, fmt.Sprintf("%v", v))
	}

	return out
}

const (
	ResNameSchedule = "Schedule"
)

func resourceScheduleCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	return resourceScheduleRead(ctx, d, meta)
}

func resourceScheduleRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).SchedulerClient

	_, err := findSchedule(ctx, conn, d.Id(), d.Get("group_name").(string))

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] EventBridge Scheduler Schedule (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return create.DiagError(names.Scheduler, create.ErrActionReading, ResNameSchedule, d.Id(), err)
	}
	return nil
}

func resourceScheduleUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	return resourceScheduleRead(ctx, d, meta)
}

func resourceScheduleDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).SchedulerClient

	log.Printf("[INFO] Deleting EventBridge Scheduler Schedule %s", d.Id())

	_, err := conn.DeleteSchedule(ctx, &scheduler.DeleteScheduleInput{
		Name: aws.String(d.Id()),
	})
	if err != nil {
		if errors.As(err, &types.ResourceNotFoundException{}) {
			return nil
		}
		return create.DiagError(names.Scheduler, create.ErrActionDeleting, ResNameSchedule, d.Id(), err)
	}

	return nil
}

func expandflexible_time_window(tfMap map[string]interface{}) *types.FlexibleTimeWindow {
	return nil
}
