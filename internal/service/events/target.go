// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package events

import (
	"context"
	"fmt"
	"log"
	"math"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/eventbridge"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/go-cty/cty"
	multierror "github.com/hashicorp/go-multierror"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/id"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

// @SDKResource("aws_cloudwatch_event_target")
func ResourceTarget() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceTargetCreate,
		ReadWithoutTimeout:   resourceTargetRead,
		UpdateWithoutTimeout: resourceTargetUpdate,
		DeleteWithoutTimeout: resourceTargetDelete,

		Importer: &schema.ResourceImporter{
			StateContext: resourceTargetImport,
		},

		SchemaVersion: 1,
		StateUpgraders: []schema.StateUpgrader{
			{
				Type:    resourceTargetV0().CoreConfigSchema().ImpliedType(),
				Upgrade: TargetStateUpgradeV0,
				Version: 0,
			},
		},

		Schema: map[string]*schema.Schema{
			"arn": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: verify.ValidARN,
			},
			"batch_target": {
				Type:     schema.TypeList,
				Optional: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"array_size": {
							Type:         schema.TypeInt,
							Optional:     true,
							ValidateFunc: validation.IntBetween(2, 10000),
						},
						"job_attempts": {
							Type:         schema.TypeInt,
							Optional:     true,
							ValidateFunc: validation.IntBetween(1, 10),
						},
						"job_definition": {
							Type:     schema.TypeString,
							Required: true,
						},
						"job_name": {
							Type:     schema.TypeString,
							Required: true,
						},
					},
				},
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
			"ecs_target": {
				Type:     schema.TypeList,
				Optional: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"capacity_provider_strategy": {
							Type:     schema.TypeSet,
							Optional: true,
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
							ValidateFunc: validation.StringInSlice(eventbridge.LaunchType_Values(), false),
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
										Elem:     &schema.Schema{Type: schema.TypeString},
									},
									"subnets": {
										Type:     schema.TypeSet,
										Required: true,
										Elem:     &schema.Schema{Type: schema.TypeString},
									},
								},
							},
						},
						"ordered_placement_strategy": {
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
										ValidateFunc: validation.StringInSlice(eventbridge.PlacementStrategyType_Values(), false),
									},
								},
							},
						},
						"placement_constraint": {
							Type:     schema.TypeSet,
							Optional: true,
							MaxItems: 10,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"expression": {
										Type:     schema.TypeString,
										Optional: true,
									},
									"type": {
										Type:         schema.TypeString,
										Required:     true,
										ValidateFunc: validation.StringInSlice(eventbridge.PlacementConstraintType_Values(), false),
									},
								},
							},
						},
						"platform_version": {
							Type:         schema.TypeString,
							Optional:     true,
							ValidateFunc: validation.StringLenBetween(0, 1600),
						},
						"propagate_tags": {
							Type:         schema.TypeString,
							Optional:     true,
							ValidateFunc: validation.StringInSlice(eventbridge.PropagateTags_Values(), false),
						},
						"tags": tftags.TagsSchema(),
						"task_count": {
							Type:         schema.TypeInt,
							Optional:     true,
							ValidateFunc: validation.IntBetween(1, math.MaxInt32),
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
			"event_bus_name": {
				Type:         schema.TypeString,
				Optional:     true,
				ForceNew:     true,
				ValidateFunc: validBusNameOrARN,
				Default:      DefaultEventBusName,
			},
			"http_target": {
				Type:     schema.TypeList,
				Optional: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"header_parameters": {
							Type:     schema.TypeMap,
							Optional: true,
							ValidateDiagFunc: allDiagFunc(
								validation.MapKeyLenBetween(0, 512),
								validation.MapKeyMatch(regexache.MustCompile(`^[0-9A-Za-z_!#$%&'*+,.^|~-]+$`), ""), // was "," meant to be included? +-. creates a range including: +,-.
								validation.MapValueLenBetween(0, 512),
								validation.MapValueMatch(regexache.MustCompile(`^[ \t]*[\x20-\x7E]+([ \t]+[\x20-\x7E]+)*[ \t]*$`), ""),
							),
							Elem: &schema.Schema{Type: schema.TypeString},
						},
						"path_parameter_values": {
							Type:     schema.TypeList,
							Optional: true,
							Elem:     &schema.Schema{Type: schema.TypeString},
						},
						"query_string_parameters": {
							Type:     schema.TypeMap,
							Optional: true,
							ValidateDiagFunc: allDiagFunc(
								validation.MapKeyLenBetween(0, 512),
								validation.MapKeyMatch(regexache.MustCompile(`[^\x00-\x1F\x7F]+`), ""),
								validation.MapValueLenBetween(0, 512),
								validation.MapValueMatch(regexache.MustCompile(`[^\x00-\x09\x0B\x0C\x0E-\x1F\x7F]+`), ""),
							),
							Elem: &schema.Schema{Type: schema.TypeString},
						},
					},
				},
			},
			"input": {
				Type:     schema.TypeString,
				Optional: true,
				ValidateFunc: validation.All(
					validation.StringIsJSON,
					validation.StringLenBetween(0, 8192),
				),
				ConflictsWith: []string{"input_path", "input_transformer"},
				// We could be normalizing the JSON here,
				// but for built-in targets input may not be JSON
			},
			"input_path": {
				Type:          schema.TypeString,
				Optional:      true,
				ValidateFunc:  validation.StringLenBetween(0, 256),
				ConflictsWith: []string{"input", "input_transformer"},
			},
			"input_transformer": {
				Type:          schema.TypeList,
				Optional:      true,
				MaxItems:      1,
				ConflictsWith: []string{"input", "input_path"},
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"input_paths": {
							Type:     schema.TypeMap,
							Optional: true,
							Elem:     &schema.Schema{Type: schema.TypeString},
							ValidateFunc: validation.All(
								mapMaxItems(targetInputTransformerMaxInputPaths),
								mapKeysDoNotMatch(regexache.MustCompile(`^AWS.*$`), "input_path must not start with \"AWS\""),
							),
						},
						"input_template": {
							Type:         schema.TypeString,
							Required:     true,
							ValidateFunc: validation.StringLenBetween(1, 8192),
						},
					},
				},
			},
			"kinesis_target": {
				Type:     schema.TypeList,
				Optional: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"partition_key_path": {
							Type:         schema.TypeString,
							Optional:     true,
							ValidateFunc: validation.StringLenBetween(1, 256),
						},
					},
				},
			},
			"redshift_target": {
				Type:     schema.TypeList,
				Optional: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"database": {
							Type:         schema.TypeString,
							Required:     true,
							ValidateFunc: validation.StringLenBetween(1, 64),
						},
						"db_user": {
							Type:         schema.TypeString,
							Optional:     true,
							ValidateFunc: validation.StringLenBetween(1, 128),
						},
						"secrets_manager_arn": {
							Type:         schema.TypeString,
							Optional:     true,
							ValidateFunc: verify.ValidARN,
						},
						"sql": {
							Type:         schema.TypeString,
							Optional:     true,
							ValidateFunc: validation.StringLenBetween(1, 100000),
						},
						"statement_name": {
							Type:         schema.TypeString,
							Optional:     true,
							ValidateFunc: validation.StringLenBetween(1, 500),
						},
						"with_event": {
							Type:     schema.TypeBool,
							Optional: true,
						},
					},
				},
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
							ValidateFunc: validation.IntBetween(0, 86400),
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
				Optional:     true,
				ValidateFunc: verify.ValidARN,
			},
			"rule": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validateRuleName,
			},
			"run_command_targets": {
				Type:     schema.TypeList,
				Optional: true,
				MaxItems: 5,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"key": {
							Type:         schema.TypeString,
							Required:     true,
							ValidateFunc: validation.StringLenBetween(1, 128),
						},
						"values": {
							Type:     schema.TypeList,
							Required: true,
							MaxItems: 50,
							Elem: &schema.Schema{
								Type:         schema.TypeString,
								ValidateFunc: validation.StringLenBetween(1, 256),
							},
						},
					},
				},
			},
			"sagemaker_pipeline_target": {
				Type:             schema.TypeList,
				Optional:         true,
				MaxItems:         1,
				DiffSuppressFunc: verify.SuppressMissingOptionalConfigurationBlock,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"pipeline_parameter_list": {
							Type:     schema.TypeSet,
							Optional: true,
							MaxItems: 200,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"name": {
										Type:     schema.TypeString,
										Required: true,
									},
									"value": {
										Type:     schema.TypeString,
										Required: true,
									},
								},
							},
						},
					},
				},
			},
			"sqs_target": {
				Type:     schema.TypeList,
				Optional: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"message_group_id": {
							Type:     schema.TypeString,
							Optional: true,
						},
					},
				},
			},
			"target_id": {
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				ForceNew:     true,
				ValidateFunc: validateTargetID,
			},
		},
	}
}

func resourceTargetCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).EventsConn(ctx)

	rule := d.Get("rule").(string)

	var targetID string
	if v, ok := d.GetOk("target_id"); ok {
		targetID = v.(string)
	} else {
		targetID = id.UniqueId()
		d.Set("target_id", targetID)
	}
	var busName string
	if v, ok := d.GetOk("event_bus_name"); ok {
		busName = v.(string)
	}
	id := TargetCreateResourceID(busName, rule, targetID)

	input := buildPutTargetInputStruct(ctx, d)

	log.Printf("[DEBUG] Creating EventBridge Target: %s", input)
	output, err := conn.PutTargetsWithContext(ctx, input)

	if err == nil && output != nil {
		err = putTargetsError(output.FailedEntries)
	}

	if err != nil {
		return diag.Errorf("creating EventBridge Target (%s): %s", id, err)
	}

	d.SetId(id)

	return resourceTargetRead(ctx, d, meta)
}

func resourceTargetRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).EventsConn(ctx)

	busName := d.Get("event_bus_name").(string)

	t, err := FindTargetByThreePartKey(ctx, conn, busName, d.Get("rule").(string), d.Get("target_id").(string))

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] EventBridge Target (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return diag.Errorf("reading EventBridge Target (%s): %s", d.Id(), err)
	}

	d.Set("arn", t.Arn)
	d.Set("target_id", t.Id)
	d.Set("input", t.Input)
	d.Set("input_path", t.InputPath)
	d.Set("role_arn", t.RoleArn)
	d.Set("event_bus_name", busName)

	if t.RunCommandParameters != nil {
		if err := d.Set("run_command_targets", flattenTargetRunParameters(t.RunCommandParameters)); err != nil {
			return diag.Errorf("setting run_command_targets: %s", err)
		}
	}

	if t.HttpParameters != nil {
		if err := d.Set("http_target", []interface{}{flattenTargetHTTPParameters(t.HttpParameters)}); err != nil {
			return diag.Errorf("setting http_target: %s", err)
		}
	} else {
		d.Set("http_target", nil)
	}

	if t.RedshiftDataParameters != nil {
		if err := d.Set("redshift_target", flattenTargetRedshiftParameters(t.RedshiftDataParameters)); err != nil {
			return diag.Errorf("setting redshift_target: %s", err)
		}
	}

	if t.EcsParameters != nil {
		if err := d.Set("ecs_target", flattenTargetECSParameters(ctx, t.EcsParameters)); err != nil {
			return diag.Errorf("setting ecs_target: %s", err)
		}
	}

	if t.BatchParameters != nil {
		if err := d.Set("batch_target", flattenTargetBatchParameters(t.BatchParameters)); err != nil {
			return diag.Errorf("setting batch_target: %s", err)
		}
	}

	if t.KinesisParameters != nil {
		if err := d.Set("kinesis_target", flattenTargetKinesisParameters(t.KinesisParameters)); err != nil {
			return diag.Errorf("setting kinesis_target: %s", err)
		}
	}

	if t.SageMakerPipelineParameters != nil {
		if err := d.Set("sagemaker_pipeline_target", flattenTargetSageMakerPipelineParameters(t.SageMakerPipelineParameters)); err != nil {
			return diag.Errorf("setting sagemaker_pipeline_parameters: %s", err)
		}
	}

	if t.SqsParameters != nil {
		if err := d.Set("sqs_target", flattenTargetSQSParameters(t.SqsParameters)); err != nil {
			return diag.Errorf("setting sqs_target: %s", err)
		}
	}

	if t.InputTransformer != nil {
		if err := d.Set("input_transformer", flattenInputTransformer(t.InputTransformer)); err != nil {
			return diag.Errorf("setting input_transformer: %s", err)
		}
	}

	if t.RetryPolicy != nil {
		if err := d.Set("retry_policy", flattenTargetRetryPolicy(t.RetryPolicy)); err != nil {
			return diag.Errorf("setting retry_policy: %s", err)
		}
	}

	if t.DeadLetterConfig != nil {
		if err := d.Set("dead_letter_config", flattenTargetDeadLetterConfig(t.DeadLetterConfig)); err != nil {
			return diag.Errorf("setting dead_letter_config: %s", err)
		}
	}

	return nil
}

func resourceTargetUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).EventsConn(ctx)

	input := buildPutTargetInputStruct(ctx, d)

	log.Printf("[DEBUG] Updating EventBridge Target: %s", input)
	output, err := conn.PutTargetsWithContext(ctx, input)

	if err == nil && output != nil {
		err = putTargetsError(output.FailedEntries)
	}

	if err != nil {
		return diag.Errorf("updating EventBridge Target (%s): %s", d.Id(), err)
	}

	return resourceTargetRead(ctx, d, meta)
}

func resourceTargetDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).EventsConn(ctx)

	input := &eventbridge.RemoveTargetsInput{
		Ids:  []*string{aws.String(d.Get("target_id").(string))},
		Rule: aws.String(d.Get("rule").(string)),
	}

	if v, ok := d.GetOk("event_bus_name"); ok {
		input.EventBusName = aws.String(v.(string))
	}

	log.Printf("[DEBUG] Deleting EventBridge Target: %s", d.Id())
	output, err := conn.RemoveTargetsWithContext(ctx, input)

	if err == nil && output != nil {
		err = removeTargetsError(output.FailedEntries)
	}

	if tfawserr.ErrCodeEquals(err, eventbridge.ErrCodeResourceNotFoundException) {
		return nil
	}

	if err != nil {
		return diag.Errorf("deleting EventBridge Target (%s): %s", d.Id(), err)
	}

	return nil
}

func putTargetError(apiObject *eventbridge.PutTargetsResultEntry) error {
	if apiObject == nil {
		return nil
	}

	return awserr.New(aws.StringValue(apiObject.ErrorCode), aws.StringValue(apiObject.ErrorMessage), nil)
}

func putTargetsError(apiObjects []*eventbridge.PutTargetsResultEntry) error {
	var errors *multierror.Error

	for _, apiObject := range apiObjects {
		if err := putTargetError(apiObject); err != nil {
			errors = multierror.Append(errors, fmt.Errorf("%s: %w", aws.StringValue(apiObject.TargetId), err))
		}
	}

	return errors.ErrorOrNil()
}

func removeTargetError(apiObject *eventbridge.RemoveTargetsResultEntry) error {
	if apiObject == nil {
		return nil
	}

	return awserr.New(aws.StringValue(apiObject.ErrorCode), aws.StringValue(apiObject.ErrorMessage), nil)
}

func removeTargetsError(apiObjects []*eventbridge.RemoveTargetsResultEntry) error {
	var errors *multierror.Error

	for _, apiObject := range apiObjects {
		if err := removeTargetError(apiObject); err != nil {
			errors = multierror.Append(errors, fmt.Errorf("%s: %w", aws.StringValue(apiObject.TargetId), err))
		}
	}

	return errors.ErrorOrNil()
}

func buildPutTargetInputStruct(ctx context.Context, d *schema.ResourceData) *eventbridge.PutTargetsInput {
	e := &eventbridge.Target{
		Arn: aws.String(d.Get("arn").(string)),
		Id:  aws.String(d.Get("target_id").(string)),
	}

	if v, ok := d.GetOk("input"); ok {
		e.Input = aws.String(v.(string))
	}
	if v, ok := d.GetOk("input_path"); ok {
		e.InputPath = aws.String(v.(string))
	}

	if v, ok := d.GetOk("role_arn"); ok {
		e.RoleArn = aws.String(v.(string))
	}

	if v, ok := d.GetOk("run_command_targets"); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
		e.RunCommandParameters = expandTargetRunParameters(v.([]interface{}))
	}

	if v, ok := d.GetOk("ecs_target"); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
		e.EcsParameters = expandTargetECSParameters(ctx, v.([]interface{}))
	}

	if v, ok := d.GetOk("redshift_target"); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
		e.RedshiftDataParameters = expandTargetRedshiftParameters(v.([]interface{}))
	}

	if v, ok := d.GetOk("http_target"); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
		e.HttpParameters = expandTargetHTTPParameters(v.([]interface{})[0].(map[string]interface{}))
	}

	if v, ok := d.GetOk("batch_target"); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
		e.BatchParameters = expandTargetBatchParameters(v.([]interface{}))
	}

	if v, ok := d.GetOk("kinesis_target"); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
		e.KinesisParameters = expandTargetKinesisParameters(v.([]interface{}))
	}

	if v, ok := d.GetOk("sqs_target"); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
		e.SqsParameters = expandTargetSQSParameters(v.([]interface{}))
	}

	if v, ok := d.GetOk("sagemaker_pipeline_target"); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
		e.SageMakerPipelineParameters = expandTargetSageMakerPipelineParameters(v.([]interface{}))
	}

	if v, ok := d.GetOk("input_transformer"); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
		e.InputTransformer = expandTransformerParameters(v.([]interface{}))
	}

	if v, ok := d.GetOk("retry_policy"); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
		e.RetryPolicy = expandRetryPolicyParameters(v.([]interface{}))
	}

	if v, ok := d.GetOk("dead_letter_config"); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
		e.DeadLetterConfig = expandDeadLetterParametersConfig(v.([]interface{}))
	}

	input := eventbridge.PutTargetsInput{
		Rule:    aws.String(d.Get("rule").(string)),
		Targets: []*eventbridge.Target{e},
	}
	if v, ok := d.GetOk("event_bus_name"); ok {
		input.EventBusName = aws.String(v.(string))
	}

	return &input
}

func expandTargetRunParameters(config []interface{}) *eventbridge.RunCommandParameters {
	commands := make([]*eventbridge.RunCommandTarget, 0)
	for _, c := range config {
		param := c.(map[string]interface{})
		command := &eventbridge.RunCommandTarget{
			Key:    aws.String(param["key"].(string)),
			Values: flex.ExpandStringList(param["values"].([]interface{})),
		}
		commands = append(commands, command)
	}

	command := &eventbridge.RunCommandParameters{
		RunCommandTargets: commands,
	}

	return command
}

func expandTargetRedshiftParameters(config []interface{}) *eventbridge.RedshiftDataParameters {
	redshiftParameters := &eventbridge.RedshiftDataParameters{}
	for _, c := range config {
		param := c.(map[string]interface{})

		redshiftParameters.Database = aws.String(param["database"].(string))
		redshiftParameters.Sql = aws.String(param["sql"].(string))

		if val, ok := param["with_event"].(bool); ok {
			redshiftParameters.WithEvent = aws.Bool(val)
		}

		if val, ok := param["statement_name"].(string); ok && val != "" {
			redshiftParameters.StatementName = aws.String(val)
		}

		if val, ok := param["secrets_manager_arn"].(string); ok && val != "" {
			redshiftParameters.SecretManagerArn = aws.String(val)
		}

		if val, ok := param["db_user"].(string); ok && val != "" {
			redshiftParameters.DbUser = aws.String(val)
		}
	}

	return redshiftParameters
}

func expandTargetECSParameters(ctx context.Context, tfList []interface{}) *eventbridge.EcsParameters {
	ecsParameters := &eventbridge.EcsParameters{}
	for _, c := range tfList {
		tfMap := c.(map[string]interface{})
		tags := tftags.New(ctx, tfMap["tags"].(map[string]interface{}))

		if v, ok := tfMap["capacity_provider_strategy"].(*schema.Set); ok && v.Len() > 0 {
			ecsParameters.CapacityProviderStrategy = expandTargetCapacityProviderStrategy(v.List())
		}

		if v, ok := tfMap["group"].(string); ok && v != "" {
			ecsParameters.Group = aws.String(v)
		}

		if v, ok := tfMap["launch_type"].(string); ok && v != "" {
			ecsParameters.LaunchType = aws.String(v)
		}

		if v, ok := tfMap["network_configuration"]; ok {
			ecsParameters.NetworkConfiguration = expandTargetECSParametersNetworkConfiguration(v.([]interface{}))
		}

		if v, ok := tfMap["platform_version"].(string); ok && v != "" {
			ecsParameters.PlatformVersion = aws.String(v)
		}

		if v, ok := tfMap["placement_constraint"].(*schema.Set); ok && v.Len() > 0 {
			ecsParameters.PlacementConstraints = expandTargetPlacementConstraints(v.List())
		}

		if v, ok := tfMap["ordered_placement_strategy"]; ok {
			ecsParameters.PlacementStrategy = expandTargetPlacementStrategies(v.([]interface{}))
		}

		if v, ok := tfMap["propagate_tags"].(string); ok && v != "" {
			ecsParameters.PropagateTags = aws.String(v)
		}

		if len(tags) > 0 {
			ecsParameters.Tags = Tags(tags.IgnoreAWS())
		}

		ecsParameters.EnableExecuteCommand = aws.Bool(tfMap["enable_execute_command"].(bool))
		ecsParameters.EnableECSManagedTags = aws.Bool(tfMap["enable_ecs_managed_tags"].(bool))
		ecsParameters.TaskCount = aws.Int64(int64(tfMap["task_count"].(int)))
		ecsParameters.TaskDefinitionArn = aws.String(tfMap["task_definition_arn"].(string))
	}

	return ecsParameters
}

func expandRetryPolicyParameters(rp []interface{}) *eventbridge.RetryPolicy {
	retryPolicy := &eventbridge.RetryPolicy{}

	for _, v := range rp {
		params := v.(map[string]interface{})

		if val, ok := params["maximum_event_age_in_seconds"].(int); ok {
			retryPolicy.MaximumEventAgeInSeconds = aws.Int64(int64(val))
		}

		if val, ok := params["maximum_retry_attempts"].(int); ok {
			retryPolicy.MaximumRetryAttempts = aws.Int64(int64(val))
		}
	}

	return retryPolicy
}

func expandDeadLetterParametersConfig(dlp []interface{}) *eventbridge.DeadLetterConfig {
	deadLetterConfig := &eventbridge.DeadLetterConfig{}

	for _, v := range dlp {
		params := v.(map[string]interface{})

		if val, ok := params["arn"].(string); ok && val != "" {
			deadLetterConfig.Arn = aws.String(val)
		}
	}

	return deadLetterConfig
}

func expandTargetECSParametersNetworkConfiguration(nc []interface{}) *eventbridge.NetworkConfiguration {
	if len(nc) == 0 {
		return nil
	}
	awsVpcConfig := &eventbridge.AwsVpcConfiguration{}
	raw := nc[0].(map[string]interface{})
	if val, ok := raw["security_groups"]; ok {
		awsVpcConfig.SecurityGroups = flex.ExpandStringSet(val.(*schema.Set))
	}
	awsVpcConfig.Subnets = flex.ExpandStringSet(raw["subnets"].(*schema.Set))
	if val, ok := raw["assign_public_ip"].(bool); ok {
		awsVpcConfig.AssignPublicIp = aws.String(eventbridge.AssignPublicIpDisabled)
		if val {
			awsVpcConfig.AssignPublicIp = aws.String(eventbridge.AssignPublicIpEnabled)
		}
	}

	return &eventbridge.NetworkConfiguration{AwsvpcConfiguration: awsVpcConfig}
}

func expandTargetBatchParameters(config []interface{}) *eventbridge.BatchParameters {
	batchParameters := &eventbridge.BatchParameters{}
	for _, c := range config {
		param := c.(map[string]interface{})
		batchParameters.JobDefinition = aws.String(param["job_definition"].(string))
		batchParameters.JobName = aws.String(param["job_name"].(string))
		if v, ok := param["array_size"].(int); ok && v > 1 && v <= 10000 {
			arrayProperties := &eventbridge.BatchArrayProperties{}
			arrayProperties.Size = aws.Int64(int64(v))
			batchParameters.ArrayProperties = arrayProperties
		}
		if v, ok := param["job_attempts"].(int); ok && v > 0 && v <= 10 {
			retryStrategy := &eventbridge.BatchRetryStrategy{}
			retryStrategy.Attempts = aws.Int64(int64(v))
			batchParameters.RetryStrategy = retryStrategy
		}
	}

	return batchParameters
}

func expandTargetKinesisParameters(config []interface{}) *eventbridge.KinesisParameters {
	kinesisParameters := &eventbridge.KinesisParameters{}
	for _, c := range config {
		param := c.(map[string]interface{})
		if v, ok := param["partition_key_path"].(string); ok && v != "" {
			kinesisParameters.PartitionKeyPath = aws.String(v)
		}
	}

	return kinesisParameters
}

func expandTargetSQSParameters(config []interface{}) *eventbridge.SqsParameters {
	sqsParameters := &eventbridge.SqsParameters{}
	for _, c := range config {
		param := c.(map[string]interface{})
		if v, ok := param["message_group_id"].(string); ok && v != "" {
			sqsParameters.MessageGroupId = aws.String(v)
		}
	}

	return sqsParameters
}

func expandTargetSageMakerPipelineParameterList(tfList []interface{}) []*eventbridge.SageMakerPipelineParameter {
	if len(tfList) == 0 {
		return nil
	}

	var result []*eventbridge.SageMakerPipelineParameter

	for _, tfMapRaw := range tfList {
		if tfMapRaw == nil {
			continue
		}

		tfMap := tfMapRaw.(map[string]interface{})

		apiObject := &eventbridge.SageMakerPipelineParameter{}

		if v, ok := tfMap["name"].(string); ok && v != "" {
			apiObject.Name = aws.String(v)
		}

		if v, ok := tfMap["value"].(string); ok && v != "" {
			apiObject.Value = aws.String(v)
		}

		result = append(result, apiObject)
	}

	return result
}

func expandTargetSageMakerPipelineParameters(config []interface{}) *eventbridge.SageMakerPipelineParameters {
	sageMakerPipelineParameters := &eventbridge.SageMakerPipelineParameters{}
	for _, c := range config {
		param := c.(map[string]interface{})
		if v, ok := param["pipeline_parameter_list"].(*schema.Set); ok && v.Len() > 0 {
			sageMakerPipelineParameters.PipelineParameterList = expandTargetSageMakerPipelineParameterList(v.List())
		}
	}

	return sageMakerPipelineParameters
}

func expandTargetHTTPParameters(tfMap map[string]interface{}) *eventbridge.HttpParameters {
	if tfMap == nil {
		return nil
	}

	apiObject := &eventbridge.HttpParameters{}

	if v, ok := tfMap["header_parameters"].(map[string]interface{}); ok && len(v) > 0 {
		apiObject.HeaderParameters = flex.ExpandStringMap(v)
	}

	if v, ok := tfMap["path_parameter_values"].([]interface{}); ok && len(v) > 0 {
		apiObject.PathParameterValues = flex.ExpandStringList(v)
	}

	if v, ok := tfMap["query_string_parameters"].(map[string]interface{}); ok && len(v) > 0 {
		apiObject.QueryStringParameters = flex.ExpandStringMap(v)
	}

	return apiObject
}

func expandTransformerParameters(config []interface{}) *eventbridge.InputTransformer {
	transformerParameters := &eventbridge.InputTransformer{}

	inputPathsMaps := map[string]*string{}

	for _, c := range config {
		param := c.(map[string]interface{})
		inputPaths := param["input_paths"].(map[string]interface{})

		for k, v := range inputPaths {
			inputPathsMaps[k] = aws.String(v.(string))
		}
		transformerParameters.InputTemplate = aws.String(param["input_template"].(string))
	}
	transformerParameters.InputPathsMap = inputPathsMaps

	return transformerParameters
}

func flattenTargetRunParameters(runCommand *eventbridge.RunCommandParameters) []map[string]interface{} {
	result := make([]map[string]interface{}, 0)

	for _, x := range runCommand.RunCommandTargets {
		config := make(map[string]interface{})

		config["key"] = aws.StringValue(x.Key)
		config["values"] = flex.FlattenStringList(x.Values)

		result = append(result, config)
	}

	return result
}

func flattenTargetECSParameters(ctx context.Context, ecsParameters *eventbridge.EcsParameters) []map[string]interface{} {
	config := make(map[string]interface{})
	if ecsParameters.Group != nil {
		config["group"] = aws.StringValue(ecsParameters.Group)
	}

	if ecsParameters.LaunchType != nil {
		config["launch_type"] = aws.StringValue(ecsParameters.LaunchType)
	}

	config["network_configuration"] = flattenTargetECSParametersNetworkConfiguration(ecsParameters.NetworkConfiguration)
	if ecsParameters.PlatformVersion != nil {
		config["platform_version"] = aws.StringValue(ecsParameters.PlatformVersion)
	}

	if ecsParameters.PropagateTags != nil {
		config["propagate_tags"] = aws.StringValue(ecsParameters.PropagateTags)
	}

	if ecsParameters.PlacementConstraints != nil {
		config["placement_constraint"] = flattenTargetPlacementConstraints(ecsParameters.PlacementConstraints)
	}

	if ecsParameters.PlacementStrategy != nil {
		config["ordered_placement_strategy"] = flattenTargetPlacementStrategies(ecsParameters.PlacementStrategy)
	}

	if ecsParameters.CapacityProviderStrategy != nil {
		config["capacity_provider_strategy"] = flattenTargetCapacityProviderStrategy(ecsParameters.CapacityProviderStrategy)
	}

	config["tags"] = KeyValueTags(ctx, ecsParameters.Tags).IgnoreAWS().Map()
	config["enable_execute_command"] = aws.BoolValue(ecsParameters.EnableExecuteCommand)
	config["enable_ecs_managed_tags"] = aws.BoolValue(ecsParameters.EnableECSManagedTags)
	config["task_count"] = aws.Int64Value(ecsParameters.TaskCount)
	config["task_definition_arn"] = aws.StringValue(ecsParameters.TaskDefinitionArn)
	result := []map[string]interface{}{config}
	return result
}

func flattenTargetRedshiftParameters(redshiftParameters *eventbridge.RedshiftDataParameters) []map[string]interface{} {
	config := make(map[string]interface{})

	if redshiftParameters == nil {
		return []map[string]interface{}{config}
	}

	config["database"] = aws.StringValue(redshiftParameters.Database)
	config["db_user"] = aws.StringValue(redshiftParameters.DbUser)
	config["secrets_manager_arn"] = aws.StringValue(redshiftParameters.SecretManagerArn)
	config["sql"] = aws.StringValue(redshiftParameters.Sql)
	config["statement_name"] = aws.StringValue(redshiftParameters.StatementName)
	config["with_event"] = aws.BoolValue(redshiftParameters.WithEvent)

	result := []map[string]interface{}{config}
	return result
}

func flattenTargetECSParametersNetworkConfiguration(nc *eventbridge.NetworkConfiguration) []interface{} {
	if nc == nil {
		return nil
	}

	result := make(map[string]interface{})
	result["security_groups"] = flex.FlattenStringSet(nc.AwsvpcConfiguration.SecurityGroups)
	result["subnets"] = flex.FlattenStringSet(nc.AwsvpcConfiguration.Subnets)

	if nc.AwsvpcConfiguration.AssignPublicIp != nil {
		result["assign_public_ip"] = aws.StringValue(nc.AwsvpcConfiguration.AssignPublicIp) == eventbridge.AssignPublicIpEnabled
	}

	return []interface{}{result}
}

func flattenTargetBatchParameters(batchParameters *eventbridge.BatchParameters) []map[string]interface{} {
	config := make(map[string]interface{})
	config["job_definition"] = aws.StringValue(batchParameters.JobDefinition)
	config["job_name"] = aws.StringValue(batchParameters.JobName)
	if batchParameters.ArrayProperties != nil {
		config["array_size"] = int(aws.Int64Value(batchParameters.ArrayProperties.Size))
	}
	if batchParameters.RetryStrategy != nil {
		config["job_attempts"] = int(aws.Int64Value(batchParameters.RetryStrategy.Attempts))
	}
	result := []map[string]interface{}{config}
	return result
}

func flattenTargetKinesisParameters(kinesisParameters *eventbridge.KinesisParameters) []map[string]interface{} {
	config := make(map[string]interface{})
	config["partition_key_path"] = aws.StringValue(kinesisParameters.PartitionKeyPath)
	result := []map[string]interface{}{config}
	return result
}

func flattenTargetSageMakerPipelineParameters(sageMakerParameters *eventbridge.SageMakerPipelineParameters) []map[string]interface{} {
	config := make(map[string]interface{})
	config["pipeline_parameter_list"] = flattenTargetSageMakerPipelineParameter(sageMakerParameters.PipelineParameterList)
	result := []map[string]interface{}{config}
	return result
}

func flattenTargetSageMakerPipelineParameter(pcs []*eventbridge.SageMakerPipelineParameter) []map[string]interface{} {
	if len(pcs) == 0 {
		return nil
	}
	results := make([]map[string]interface{}, 0)
	for _, pc := range pcs {
		c := make(map[string]interface{})
		c["name"] = aws.StringValue(pc.Name)
		c["value"] = aws.StringValue(pc.Value)

		results = append(results, c)
	}
	return results
}

func flattenTargetSQSParameters(sqsParameters *eventbridge.SqsParameters) []map[string]interface{} {
	config := make(map[string]interface{})
	config["message_group_id"] = aws.StringValue(sqsParameters.MessageGroupId)
	result := []map[string]interface{}{config}
	return result
}

func flattenTargetHTTPParameters(apiObject *eventbridge.HttpParameters) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if v := apiObject.HeaderParameters; v != nil {
		tfMap["header_parameters"] = aws.StringValueMap(v)
	}

	if v := apiObject.PathParameterValues; v != nil {
		tfMap["path_parameter_values"] = aws.StringValueSlice(v)
	}

	if v := apiObject.QueryStringParameters; v != nil {
		tfMap["query_string_parameters"] = aws.StringValueMap(v)
	}

	return tfMap
}

func flattenInputTransformer(inputTransformer *eventbridge.InputTransformer) []map[string]interface{} {
	config := make(map[string]interface{})
	inputPathsMap := make(map[string]string)
	for k, v := range inputTransformer.InputPathsMap {
		inputPathsMap[k] = aws.StringValue(v)
	}
	config["input_template"] = aws.StringValue(inputTransformer.InputTemplate)
	config["input_paths"] = inputPathsMap

	result := []map[string]interface{}{config}
	return result
}

func flattenTargetRetryPolicy(rp *eventbridge.RetryPolicy) []map[string]interface{} {
	config := make(map[string]interface{})

	config["maximum_event_age_in_seconds"] = aws.Int64Value(rp.MaximumEventAgeInSeconds)
	config["maximum_retry_attempts"] = aws.Int64Value(rp.MaximumRetryAttempts)

	result := []map[string]interface{}{config}
	return result
}

func flattenTargetDeadLetterConfig(dlc *eventbridge.DeadLetterConfig) []map[string]interface{} {
	config := make(map[string]interface{})

	config["arn"] = aws.StringValue(dlc.Arn)

	result := []map[string]interface{}{config}
	return result
}

func expandTargetPlacementConstraints(tfList []interface{}) []*eventbridge.PlacementConstraint {
	if len(tfList) == 0 {
		return nil
	}

	var result []*eventbridge.PlacementConstraint

	for _, tfMapRaw := range tfList {
		if tfMapRaw == nil {
			continue
		}

		tfMap := tfMapRaw.(map[string]interface{})

		apiObject := &eventbridge.PlacementConstraint{}

		if v, ok := tfMap["expression"].(string); ok && v != "" {
			apiObject.Expression = aws.String(v)
		}

		if v, ok := tfMap["type"].(string); ok && v != "" {
			apiObject.Type = aws.String(v)
		}

		result = append(result, apiObject)
	}

	return result
}

func expandTargetPlacementStrategies(tfList []interface{}) []*eventbridge.PlacementStrategy {
	if len(tfList) == 0 {
		return nil
	}

	var result []*eventbridge.PlacementStrategy

	for _, tfMapRaw := range tfList {
		if tfMapRaw == nil {
			continue
		}

		tfMap := tfMapRaw.(map[string]interface{})

		apiObject := &eventbridge.PlacementStrategy{}

		if v, ok := tfMap["field"].(string); ok && v != "" {
			apiObject.Field = aws.String(v)
		}

		if v, ok := tfMap["type"].(string); ok && v != "" {
			apiObject.Type = aws.String(v)
		}

		result = append(result, apiObject)
	}

	return result
}

func expandTargetCapacityProviderStrategy(tfList []interface{}) []*eventbridge.CapacityProviderStrategyItem {
	if len(tfList) == 0 {
		return nil
	}

	var result []*eventbridge.CapacityProviderStrategyItem

	for _, tfMapRaw := range tfList {
		if tfMapRaw == nil {
			continue
		}

		cp := tfMapRaw.(map[string]interface{})

		apiObject := &eventbridge.CapacityProviderStrategyItem{}

		if val, ok := cp["base"]; ok {
			apiObject.Base = aws.Int64(int64(val.(int)))
		}

		if val, ok := cp["weight"]; ok {
			apiObject.Weight = aws.Int64(int64(val.(int)))
		}

		if val, ok := cp["capacity_provider"]; ok {
			apiObject.CapacityProvider = aws.String(val.(string))
		}

		result = append(result, apiObject)
	}

	return result
}

func flattenTargetPlacementConstraints(pcs []*eventbridge.PlacementConstraint) []map[string]interface{} {
	if len(pcs) == 0 {
		return nil
	}
	results := make([]map[string]interface{}, 0)
	for _, pc := range pcs {
		c := make(map[string]interface{})
		c["type"] = aws.StringValue(pc.Type)
		if pc.Expression != nil {
			c["expression"] = aws.StringValue(pc.Expression)
		}

		results = append(results, c)
	}
	return results
}

func flattenTargetPlacementStrategies(pcs []*eventbridge.PlacementStrategy) []map[string]interface{} {
	if len(pcs) == 0 {
		return nil
	}
	results := make([]map[string]interface{}, 0)
	for _, pc := range pcs {
		c := make(map[string]interface{})
		c["type"] = aws.StringValue(pc.Type)
		if pc.Field != nil {
			c["field"] = aws.StringValue(pc.Field)
		}

		results = append(results, c)
	}
	return results
}

func flattenTargetCapacityProviderStrategy(cps []*eventbridge.CapacityProviderStrategyItem) []map[string]interface{} {
	if cps == nil {
		return nil
	}
	results := make([]map[string]interface{}, 0)
	for _, cp := range cps {
		s := make(map[string]interface{})
		s["capacity_provider"] = aws.StringValue(cp.CapacityProvider)
		if cp.Weight != nil {
			s["weight"] = aws.Int64Value(cp.Weight)
		}
		if cp.Base != nil {
			s["base"] = aws.Int64Value(cp.Base)
		}
		results = append(results, s)
	}
	return results
}

func resourceTargetImport(ctx context.Context, d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	busName, ruleName, targetID, err := TargetParseImportID(d.Id())
	if err != nil {
		return []*schema.ResourceData{}, err
	}

	id := TargetCreateResourceID(busName, ruleName, targetID)
	d.SetId(id)
	d.Set("target_id", targetID)
	d.Set("rule", ruleName)
	d.Set("event_bus_name", busName)

	return []*schema.ResourceData{d}, nil
}

func allDiagFunc(validators ...schema.SchemaValidateDiagFunc) schema.SchemaValidateDiagFunc {
	return func(i interface{}, k cty.Path) diag.Diagnostics {
		var diags diag.Diagnostics
		for _, validator := range validators {
			diags = append(diags, validator(i, k)...)
		}
		return diags
	}
}
