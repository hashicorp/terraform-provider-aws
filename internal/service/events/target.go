// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package events

import (
	"context"
	"errors"
	"fmt"
	"log"
	"math"
	"strings"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/eventbridge"
	"github.com/aws/aws-sdk-go-v2/service/eventbridge/types"
	"github.com/hashicorp/aws-sdk-go-base/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/id"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	tfslices "github.com/hashicorp/terraform-provider-aws/internal/slices"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_cloudwatch_event_target", name="Target")
func resourceTarget() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceTargetCreate,
		ReadWithoutTimeout:   resourceTargetRead,
		UpdateWithoutTimeout: resourceTargetUpdate,
		DeleteWithoutTimeout: resourceTargetDelete,

		Importer: &schema.ResourceImporter{
			StateContext: func(ctx context.Context, d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
				busName, ruleName, targetID, err := targetParseImportID(d.Id())
				if err != nil {
					return []*schema.ResourceData{}, err
				}

				id := targetCreateResourceID(busName, ruleName, targetID)
				d.SetId(id)
				d.Set("target_id", targetID)
				d.Set(names.AttrRule, ruleName)
				d.Set("event_bus_name", busName)

				return []*schema.ResourceData{d}, nil
			},
		},

		SchemaVersion: 1,
		StateUpgraders: []schema.StateUpgrader{
			{
				Type:    resourceTargetV0().CoreConfigSchema().ImpliedType(),
				Upgrade: targetStateUpgradeV0,
				Version: 0,
			},
		},

		Schema: map[string]*schema.Schema{
			names.AttrARN: {
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
						names.AttrARN: {
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
						names.AttrCapacityProviderStrategy: {
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
									names.AttrWeight: {
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
							Type:             schema.TypeString,
							Optional:         true,
							ValidateDiagFunc: enum.Validate[types.LaunchType](),
						},
						names.AttrNetworkConfiguration: {
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
									names.AttrSecurityGroups: {
										Type:     schema.TypeSet,
										Optional: true,
										Elem:     &schema.Schema{Type: schema.TypeString},
									},
									names.AttrSubnets: {
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
									names.AttrField: {
										Type:         schema.TypeString,
										Optional:     true,
										ValidateFunc: validation.StringLenBetween(0, 255),
									},
									names.AttrType: {
										Type:             schema.TypeString,
										Required:         true,
										ValidateDiagFunc: enum.Validate[types.PlacementStrategyType](),
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
									names.AttrExpression: {
										Type:     schema.TypeString,
										Optional: true,
									},
									names.AttrType: {
										Type:             schema.TypeString,
										Required:         true,
										ValidateDiagFunc: enum.Validate[types.PlacementConstraintType](),
									},
								},
							},
						},
						"platform_version": {
							Type:         schema.TypeString,
							Optional:     true,
							ValidateFunc: validation.StringLenBetween(0, 1600),
						},
						names.AttrPropagateTags: {
							Type:             schema.TypeString,
							Optional:         true,
							ValidateDiagFunc: enum.Validate[types.PropagateTags](),
						},
						names.AttrTags: tftags.TagsSchema(),
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
			names.AttrForceDestroy: {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
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
							ValidateDiagFunc: validation.AllDiag(
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
							ValidateDiagFunc: validation.AllDiag(
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
							ValidateDiagFunc: validation.AllDiag(
								verify.MapSizeAtMost(targetInputTransformerMaxInputPaths),
								verify.MapKeyNoMatch(regexache.MustCompile(`^AWS.*$`), `must not start with "AWS"`),
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
						names.AttrDatabase: {
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
			names.AttrRoleARN: {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: verify.ValidARN,
			},
			names.AttrRule: {
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
						names.AttrKey: {
							Type:         schema.TypeString,
							Required:     true,
							ValidateFunc: validation.StringLenBetween(1, 128),
						},
						names.AttrValues: {
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
									names.AttrName: {
										Type:     schema.TypeString,
										Required: true,
									},
									names.AttrValue: {
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
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).EventsClient(ctx)

	ruleName := d.Get(names.AttrRule).(string)
	var targetID string
	if v, ok := d.GetOk("target_id"); ok {
		targetID = v.(string)
	} else {
		targetID = id.UniqueId()
		d.Set("target_id", targetID)
	}
	var eventBusName string
	if v, ok := d.GetOk("event_bus_name"); ok {
		eventBusName = v.(string)
	}
	id := targetCreateResourceID(eventBusName, ruleName, targetID)

	input := expandPutTargetsInput(ctx, d)

	output, err := conn.PutTargets(ctx, input)

	if err == nil && output != nil {
		err = putTargetsError(output.FailedEntries)
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating EventBridge Target (%s): %s", id, err)
	}

	d.SetId(id)

	return append(diags, resourceTargetRead(ctx, d, meta)...)
}

func resourceTargetRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).EventsClient(ctx)

	eventBusName := d.Get("event_bus_name").(string)
	target, err := findTargetByThreePartKey(ctx, conn, eventBusName, d.Get(names.AttrRule).(string), d.Get("target_id").(string))

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] EventBridge Target (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading EventBridge Target (%s): %s", d.Id(), err)
	}

	d.Set(names.AttrARN, target.Arn)
	d.Set("event_bus_name", eventBusName)
	d.Set(names.AttrForceDestroy, d.Get(names.AttrForceDestroy).(bool))
	d.Set("input", target.Input)
	d.Set("input_path", target.InputPath)
	d.Set(names.AttrRoleARN, target.RoleArn)
	d.Set("target_id", target.Id)

	if target.RunCommandParameters != nil {
		if err := d.Set("run_command_targets", flattenTargetRunParameters(target.RunCommandParameters)); err != nil {
			return sdkdiag.AppendErrorf(diags, "setting run_command_targets: %s", err)
		}
	}

	if target.HttpParameters != nil {
		if err := d.Set("http_target", []interface{}{flattenTargetHTTPParameters(target.HttpParameters)}); err != nil {
			return sdkdiag.AppendErrorf(diags, "setting http_target: %s", err)
		}
	} else {
		d.Set("http_target", nil)
	}

	if target.RedshiftDataParameters != nil {
		if err := d.Set("redshift_target", flattenTargetRedshiftParameters(target.RedshiftDataParameters)); err != nil {
			return sdkdiag.AppendErrorf(diags, "setting redshift_target: %s", err)
		}
	}

	if target.EcsParameters != nil {
		if err := d.Set("ecs_target", flattenTargetECSParameters(ctx, target.EcsParameters)); err != nil {
			return sdkdiag.AppendErrorf(diags, "setting ecs_target: %s", err)
		}
	}

	if target.BatchParameters != nil {
		if err := d.Set("batch_target", flattenTargetBatchParameters(target.BatchParameters)); err != nil {
			return sdkdiag.AppendErrorf(diags, "setting batch_target: %s", err)
		}
	}

	if target.KinesisParameters != nil {
		if err := d.Set("kinesis_target", flattenTargetKinesisParameters(target.KinesisParameters)); err != nil {
			return sdkdiag.AppendErrorf(diags, "setting kinesis_target: %s", err)
		}
	}

	if target.SageMakerPipelineParameters != nil {
		if err := d.Set("sagemaker_pipeline_target", flattenTargetSageMakerPipelineParameters(target.SageMakerPipelineParameters)); err != nil {
			return sdkdiag.AppendErrorf(diags, "setting sagemaker_pipeline_parameters: %s", err)
		}
	}

	if target.SqsParameters != nil {
		if err := d.Set("sqs_target", flattenTargetSQSParameters(target.SqsParameters)); err != nil {
			return sdkdiag.AppendErrorf(diags, "setting sqs_target: %s", err)
		}
	}

	if target.InputTransformer != nil {
		if err := d.Set("input_transformer", flattenInputTransformer(target.InputTransformer)); err != nil {
			return sdkdiag.AppendErrorf(diags, "setting input_transformer: %s", err)
		}
	}

	if target.RetryPolicy != nil {
		if err := d.Set("retry_policy", flattenTargetRetryPolicy(target.RetryPolicy)); err != nil {
			return sdkdiag.AppendErrorf(diags, "setting retry_policy: %s", err)
		}
	}

	if target.DeadLetterConfig != nil {
		if err := d.Set("dead_letter_config", flattenTargetDeadLetterConfig(target.DeadLetterConfig)); err != nil {
			return sdkdiag.AppendErrorf(diags, "setting dead_letter_config: %s", err)
		}
	}

	return diags
}

func resourceTargetUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).EventsClient(ctx)

	if d.HasChangesExcept(names.AttrForceDestroy) {
		input := expandPutTargetsInput(ctx, d)

		output, err := conn.PutTargets(ctx, input)

		if err == nil && output != nil {
			err = putTargetsError(output.FailedEntries)
		}

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "updating EventBridge Target (%s): %s", d.Id(), err)
		}
	}

	return append(diags, resourceTargetRead(ctx, d, meta)...)
}

func resourceTargetDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).EventsClient(ctx)

	input := &eventbridge.RemoveTargetsInput{
		Ids:  []string{d.Get("target_id").(string)},
		Rule: aws.String(d.Get(names.AttrRule).(string)),
	}

	if v, ok := d.GetOk("event_bus_name"); ok {
		input.EventBusName = aws.String(v.(string))
	}

	if v, ok := d.GetOk(names.AttrForceDestroy); ok {
		input.Force = v.(bool)
	}

	log.Printf("[DEBUG] Deleting EventBridge Target: %s", d.Id())
	output, err := conn.RemoveTargets(ctx, input)

	if err == nil && output != nil {
		err = removeTargetsError(output.FailedEntries)
	}

	if errs.IsA[*types.ResourceNotFoundException](err) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting EventBridge Target (%s): %s", d.Id(), err)
	}

	return diags
}

func findTargetByThreePartKey(ctx context.Context, conn *eventbridge.Client, busName, ruleName, targetID string) (*types.Target, error) {
	input := &eventbridge.ListTargetsByRuleInput{
		Rule:  aws.String(ruleName),
		Limit: aws.Int32(100), // Set limit to allowed maximum to prevent API throttling
	}
	if busName != "" {
		input.EventBusName = aws.String(busName)
	}

	return findTarget(ctx, conn, input, func(v *types.Target) bool {
		return targetID == aws.ToString(v.Id)
	})
}

func findTarget(ctx context.Context, conn *eventbridge.Client, input *eventbridge.ListTargetsByRuleInput, filter tfslices.Predicate[*types.Target]) (*types.Target, error) {
	output, err := findTargets(ctx, conn, input, filter)

	if err != nil {
		return nil, err
	}

	return tfresource.AssertSingleValueResult(output)
}

func findTargets(ctx context.Context, conn *eventbridge.Client, input *eventbridge.ListTargetsByRuleInput, filter tfslices.Predicate[*types.Target]) ([]types.Target, error) {
	var output []types.Target

	err := listTargetsByRulePages(ctx, conn, input, func(page *eventbridge.ListTargetsByRuleOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, v := range page.Targets {
			if filter(&v) {
				output = append(output, v)
			}
		}

		return !lastPage
	})

	if tfawserr.ErrCodeEquals(err, errCodeValidationException) || errs.IsA[*types.ResourceNotFoundException](err) || (err != nil && regexache.MustCompile(" not found$").MatchString(err.Error())) {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	return output, nil
}

// Terraform resource IDs for Targets are not parseable as the separator used ("-") is also a valid character in both the rule name and the target ID.
const (
	targetResourceIDSeparator = "-"
	targetImportIDSeparator   = "/"
)

func targetCreateResourceID(eventBusName, ruleName, targetID string) string {
	var parts []string

	if eventBusName == "" || eventBusName == DefaultEventBusName {
		parts = []string{ruleName, targetID}
	} else {
		parts = []string{eventBusName, ruleName, targetID}
	}

	id := strings.Join(parts, targetResourceIDSeparator)

	return id
}

func targetParseImportID(id string) (string, string, string, error) {
	parts := strings.Split(id, targetImportIDSeparator)

	if len(parts) == 2 && parts[0] != "" && parts[1] != "" {
		return DefaultEventBusName, parts[0], parts[1], nil
	}
	if len(parts) == 3 && parts[0] != "" && parts[1] != "" && parts[2] != "" {
		return parts[0], parts[1], parts[2], nil
	}
	if len(parts) > 3 {
		iTarget := strings.LastIndex(id, targetImportIDSeparator)
		targetID := id[iTarget+1:]
		iRule := strings.LastIndex(id[:iTarget], targetImportIDSeparator)
		eventBusName := id[:iRule]
		ruleName := id[iRule+1 : iTarget]
		if eventBusARNPattern.MatchString(eventBusName) && ruleName != "" && targetID != "" {
			return eventBusName, ruleName, targetID, nil
		}
		if partnerEventBusPattern.MatchString(eventBusName) && ruleName != "" && targetID != "" {
			return eventBusName, ruleName, targetID, nil
		}
	}

	return "", "", "", fmt.Errorf("unexpected format for ID (%[1]s), expected EVENTBUSNAME%[2]sRULENAME%[2]sTARGETID or RULENAME%[2]sTARGETID", id, targetImportIDSeparator)
}

func putTargetError(apiObject types.PutTargetsResultEntry) error {
	return errs.APIError(aws.ToString(apiObject.ErrorCode), aws.ToString(apiObject.ErrorMessage))
}

func putTargetsError(apiObjects []types.PutTargetsResultEntry) error {
	var errs []error

	for _, apiObject := range apiObjects {
		errs = append(errs, fmt.Errorf("%s: %w", aws.ToString(apiObject.TargetId), putTargetError(apiObject)))
	}

	return errors.Join(errs...)
}

func removeTargetError(apiObject types.RemoveTargetsResultEntry) error {
	return errs.APIError(aws.ToString(apiObject.ErrorCode), aws.ToString(apiObject.ErrorMessage))
}

func removeTargetsError(apiObjects []types.RemoveTargetsResultEntry) error {
	var errs []error

	for _, apiObject := range apiObjects {
		errs = append(errs, fmt.Errorf("%s: %w", aws.ToString(apiObject.TargetId), removeTargetError(apiObject)))
	}

	return errors.Join(errs...)
}

func expandPutTargetsInput(ctx context.Context, d *schema.ResourceData) *eventbridge.PutTargetsInput {
	target := types.Target{
		Arn: aws.String(d.Get(names.AttrARN).(string)),
		Id:  aws.String(d.Get("target_id").(string)),
	}

	if v, ok := d.GetOk("input"); ok {
		target.Input = aws.String(v.(string))
	}

	if v, ok := d.GetOk("input_path"); ok {
		target.InputPath = aws.String(v.(string))
	}

	if v, ok := d.GetOk(names.AttrRoleARN); ok {
		target.RoleArn = aws.String(v.(string))
	}

	if v, ok := d.GetOk("run_command_targets"); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
		target.RunCommandParameters = expandTargetRunParameters(v.([]interface{}))
	}

	if v, ok := d.GetOk("ecs_target"); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
		target.EcsParameters = expandTargetECSParameters(ctx, v.([]interface{}))
	}

	if v, ok := d.GetOk("redshift_target"); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
		target.RedshiftDataParameters = expandTargetRedshiftParameters(v.([]interface{}))
	}

	if v, ok := d.GetOk("http_target"); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
		target.HttpParameters = expandTargetHTTPParameters(v.([]interface{})[0].(map[string]interface{}))
	}

	if v, ok := d.GetOk("batch_target"); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
		target.BatchParameters = expandTargetBatchParameters(v.([]interface{}))
	}

	if v, ok := d.GetOk("kinesis_target"); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
		target.KinesisParameters = expandTargetKinesisParameters(v.([]interface{}))
	}

	if v, ok := d.GetOk("sqs_target"); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
		target.SqsParameters = expandTargetSQSParameters(v.([]interface{}))
	}

	if v, ok := d.GetOk("sagemaker_pipeline_target"); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
		target.SageMakerPipelineParameters = expandTargetSageMakerPipelineParameters(v.([]interface{}))
	}

	if v, ok := d.GetOk("input_transformer"); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
		target.InputTransformer = expandTransformerParameters(v.([]interface{}))
	}

	if v, ok := d.GetOk("retry_policy"); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
		target.RetryPolicy = expandRetryPolicyParameters(v.([]interface{}))
	}

	if v, ok := d.GetOk("dead_letter_config"); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
		target.DeadLetterConfig = expandDeadLetterParametersConfig(v.([]interface{}))
	}

	input := &eventbridge.PutTargetsInput{
		Rule:    aws.String(d.Get(names.AttrRule).(string)),
		Targets: []types.Target{target},
	}

	if v, ok := d.GetOk("event_bus_name"); ok {
		input.EventBusName = aws.String(v.(string))
	}

	return input
}

func expandTargetRunParameters(config []interface{}) *types.RunCommandParameters {
	commands := make([]types.RunCommandTarget, 0)
	for _, c := range config {
		param := c.(map[string]interface{})
		command := types.RunCommandTarget{
			Key:    aws.String(param[names.AttrKey].(string)),
			Values: flex.ExpandStringValueList(param[names.AttrValues].([]interface{})),
		}
		commands = append(commands, command)
	}

	command := &types.RunCommandParameters{
		RunCommandTargets: commands,
	}

	return command
}

func expandTargetRedshiftParameters(config []interface{}) *types.RedshiftDataParameters {
	redshiftParameters := &types.RedshiftDataParameters{}
	for _, c := range config {
		param := c.(map[string]interface{})

		redshiftParameters.Database = aws.String(param[names.AttrDatabase].(string))
		redshiftParameters.Sql = aws.String(param["sql"].(string))

		if val, ok := param["with_event"].(bool); ok {
			redshiftParameters.WithEvent = val
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

func expandTargetECSParameters(ctx context.Context, tfList []interface{}) *types.EcsParameters {
	ecsParameters := &types.EcsParameters{}
	for _, c := range tfList {
		tfMap := c.(map[string]interface{})
		tags := tftags.New(ctx, tfMap[names.AttrTags].(map[string]interface{}))

		if v, ok := tfMap[names.AttrCapacityProviderStrategy].(*schema.Set); ok && v.Len() > 0 {
			ecsParameters.CapacityProviderStrategy = expandTargetCapacityProviderStrategy(v.List())
		}

		if v, ok := tfMap["group"].(string); ok && v != "" {
			ecsParameters.Group = aws.String(v)
		}

		if v, ok := tfMap["launch_type"].(string); ok && v != "" {
			ecsParameters.LaunchType = types.LaunchType(v)
		}

		if v, ok := tfMap[names.AttrNetworkConfiguration]; ok {
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

		if v, ok := tfMap[names.AttrPropagateTags].(string); ok && v != "" {
			ecsParameters.PropagateTags = types.PropagateTags(v)
		}

		if len(tags) > 0 {
			ecsParameters.Tags = Tags(tags.IgnoreAWS())
		}

		ecsParameters.EnableExecuteCommand = tfMap["enable_execute_command"].(bool)
		ecsParameters.EnableECSManagedTags = tfMap["enable_ecs_managed_tags"].(bool)
		ecsParameters.TaskCount = aws.Int32(int32(tfMap["task_count"].(int)))
		ecsParameters.TaskDefinitionArn = aws.String(tfMap["task_definition_arn"].(string))
	}

	return ecsParameters
}

func expandRetryPolicyParameters(rp []interface{}) *types.RetryPolicy {
	retryPolicy := &types.RetryPolicy{}

	for _, v := range rp {
		params := v.(map[string]interface{})

		if val, ok := params["maximum_event_age_in_seconds"].(int); ok {
			retryPolicy.MaximumEventAgeInSeconds = aws.Int32(int32(val))
		}

		if val, ok := params["maximum_retry_attempts"].(int); ok {
			retryPolicy.MaximumRetryAttempts = aws.Int32(int32(val))
		}
	}

	return retryPolicy
}

func expandDeadLetterParametersConfig(dlp []interface{}) *types.DeadLetterConfig {
	deadLetterConfig := &types.DeadLetterConfig{}

	for _, v := range dlp {
		params := v.(map[string]interface{})

		if val, ok := params[names.AttrARN].(string); ok && val != "" {
			deadLetterConfig.Arn = aws.String(val)
		}
	}

	return deadLetterConfig
}

func expandTargetECSParametersNetworkConfiguration(nc []interface{}) *types.NetworkConfiguration {
	if len(nc) == 0 {
		return nil
	}
	awsVpcConfig := &types.AwsVpcConfiguration{}
	raw := nc[0].(map[string]interface{})
	if val, ok := raw[names.AttrSecurityGroups]; ok {
		awsVpcConfig.SecurityGroups = flex.ExpandStringValueSet(val.(*schema.Set))
	}
	awsVpcConfig.Subnets = flex.ExpandStringValueSet(raw[names.AttrSubnets].(*schema.Set))
	if val, ok := raw["assign_public_ip"].(bool); ok {
		awsVpcConfig.AssignPublicIp = types.AssignPublicIpDisabled
		if val {
			awsVpcConfig.AssignPublicIp = types.AssignPublicIpEnabled
		}
	}

	return &types.NetworkConfiguration{AwsvpcConfiguration: awsVpcConfig}
}

func expandTargetBatchParameters(config []interface{}) *types.BatchParameters {
	batchParameters := &types.BatchParameters{}
	for _, c := range config {
		param := c.(map[string]interface{})
		batchParameters.JobDefinition = aws.String(param["job_definition"].(string))
		batchParameters.JobName = aws.String(param["job_name"].(string))
		if v, ok := param["array_size"].(int); ok && v > 1 && v <= 10000 {
			arrayProperties := &types.BatchArrayProperties{}
			arrayProperties.Size = int32(v)
			batchParameters.ArrayProperties = arrayProperties
		}
		if v, ok := param["job_attempts"].(int); ok && v > 0 && v <= 10 {
			retryStrategy := &types.BatchRetryStrategy{}
			retryStrategy.Attempts = int32(v)
			batchParameters.RetryStrategy = retryStrategy
		}
	}

	return batchParameters
}

func expandTargetKinesisParameters(config []interface{}) *types.KinesisParameters {
	kinesisParameters := &types.KinesisParameters{}
	for _, c := range config {
		param := c.(map[string]interface{})
		if v, ok := param["partition_key_path"].(string); ok && v != "" {
			kinesisParameters.PartitionKeyPath = aws.String(v)
		}
	}

	return kinesisParameters
}

func expandTargetSQSParameters(config []interface{}) *types.SqsParameters {
	sqsParameters := &types.SqsParameters{}
	for _, c := range config {
		param := c.(map[string]interface{})
		if v, ok := param["message_group_id"].(string); ok && v != "" {
			sqsParameters.MessageGroupId = aws.String(v)
		}
	}

	return sqsParameters
}

func expandTargetSageMakerPipelineParameterList(tfList []interface{}) []types.SageMakerPipelineParameter {
	if len(tfList) == 0 {
		return nil
	}

	var result []types.SageMakerPipelineParameter

	for _, tfMapRaw := range tfList {
		if tfMapRaw == nil {
			continue
		}

		tfMap := tfMapRaw.(map[string]interface{})

		apiObject := types.SageMakerPipelineParameter{}

		if v, ok := tfMap[names.AttrName].(string); ok && v != "" {
			apiObject.Name = aws.String(v)
		}

		if v, ok := tfMap[names.AttrValue].(string); ok && v != "" {
			apiObject.Value = aws.String(v)
		}

		result = append(result, apiObject)
	}

	return result
}

func expandTargetSageMakerPipelineParameters(config []interface{}) *types.SageMakerPipelineParameters {
	sageMakerPipelineParameters := &types.SageMakerPipelineParameters{}
	for _, c := range config {
		param := c.(map[string]interface{})
		if v, ok := param["pipeline_parameter_list"].(*schema.Set); ok && v.Len() > 0 {
			sageMakerPipelineParameters.PipelineParameterList = expandTargetSageMakerPipelineParameterList(v.List())
		}
	}

	return sageMakerPipelineParameters
}

func expandTargetHTTPParameters(tfMap map[string]interface{}) *types.HttpParameters {
	if tfMap == nil {
		return nil
	}

	apiObject := &types.HttpParameters{}

	if v, ok := tfMap["header_parameters"].(map[string]interface{}); ok && len(v) > 0 {
		apiObject.HeaderParameters = flex.ExpandStringValueMap(v)
	}

	if v, ok := tfMap["path_parameter_values"].([]interface{}); ok && len(v) > 0 {
		apiObject.PathParameterValues = flex.ExpandStringValueList(v)
	}

	if v, ok := tfMap["query_string_parameters"].(map[string]interface{}); ok && len(v) > 0 {
		apiObject.QueryStringParameters = flex.ExpandStringValueMap(v)
	}

	return apiObject
}

func expandTransformerParameters(config []interface{}) *types.InputTransformer {
	transformerParameters := &types.InputTransformer{}

	inputPathsMaps := map[string]string{}

	for _, c := range config {
		param := c.(map[string]interface{})
		inputPaths := param["input_paths"].(map[string]interface{})

		for k, v := range inputPaths {
			inputPathsMaps[k] = v.(string)
		}
		transformerParameters.InputTemplate = aws.String(param["input_template"].(string))
	}
	transformerParameters.InputPathsMap = inputPathsMaps

	return transformerParameters
}

func flattenTargetRunParameters(runCommand *types.RunCommandParameters) []map[string]interface{} {
	result := make([]map[string]interface{}, 0)

	for _, x := range runCommand.RunCommandTargets {
		config := make(map[string]interface{})

		config[names.AttrKey] = aws.ToString(x.Key)
		config[names.AttrValues] = x.Values

		result = append(result, config)
	}

	return result
}

func flattenTargetECSParameters(ctx context.Context, ecsParameters *types.EcsParameters) []map[string]interface{} {
	config := make(map[string]interface{})
	if ecsParameters.Group != nil {
		config["group"] = aws.ToString(ecsParameters.Group)
	}

	config["launch_type"] = ecsParameters.LaunchType

	config[names.AttrNetworkConfiguration] = flattenTargetECSParametersNetworkConfiguration(ecsParameters.NetworkConfiguration)
	if ecsParameters.PlatformVersion != nil {
		config["platform_version"] = aws.ToString(ecsParameters.PlatformVersion)
	}

	config[names.AttrPropagateTags] = ecsParameters.PropagateTags

	if ecsParameters.PlacementConstraints != nil {
		config["placement_constraint"] = flattenTargetPlacementConstraints(ecsParameters.PlacementConstraints)
	}

	if ecsParameters.PlacementStrategy != nil {
		config["ordered_placement_strategy"] = flattenTargetPlacementStrategies(ecsParameters.PlacementStrategy)
	}

	if ecsParameters.CapacityProviderStrategy != nil {
		config[names.AttrCapacityProviderStrategy] = flattenTargetCapacityProviderStrategy(ecsParameters.CapacityProviderStrategy)
	}

	config[names.AttrTags] = KeyValueTags(ctx, ecsParameters.Tags).IgnoreAWS().Map()
	config["enable_execute_command"] = ecsParameters.EnableExecuteCommand
	config["enable_ecs_managed_tags"] = ecsParameters.EnableECSManagedTags
	config["task_count"] = aws.ToInt32(ecsParameters.TaskCount)
	config["task_definition_arn"] = aws.ToString(ecsParameters.TaskDefinitionArn)
	result := []map[string]interface{}{config}
	return result
}

func flattenTargetRedshiftParameters(redshiftParameters *types.RedshiftDataParameters) []map[string]interface{} {
	config := make(map[string]interface{})

	if redshiftParameters == nil {
		return []map[string]interface{}{config}
	}

	config[names.AttrDatabase] = aws.ToString(redshiftParameters.Database)
	config["db_user"] = aws.ToString(redshiftParameters.DbUser)
	config["secrets_manager_arn"] = aws.ToString(redshiftParameters.SecretManagerArn)
	config["sql"] = aws.ToString(redshiftParameters.Sql)
	config["statement_name"] = aws.ToString(redshiftParameters.StatementName)
	config["with_event"] = redshiftParameters.WithEvent

	result := []map[string]interface{}{config}
	return result
}

func flattenTargetECSParametersNetworkConfiguration(nc *types.NetworkConfiguration) []interface{} {
	if nc == nil {
		return nil
	}

	result := make(map[string]interface{})
	result[names.AttrSecurityGroups] = nc.AwsvpcConfiguration.SecurityGroups
	result[names.AttrSubnets] = nc.AwsvpcConfiguration.Subnets
	result["assign_public_ip"] = nc.AwsvpcConfiguration.AssignPublicIp == types.AssignPublicIpEnabled

	return []interface{}{result}
}

func flattenTargetBatchParameters(batchParameters *types.BatchParameters) []map[string]interface{} {
	config := make(map[string]interface{})
	config["job_definition"] = aws.ToString(batchParameters.JobDefinition)
	config["job_name"] = aws.ToString(batchParameters.JobName)
	if batchParameters.ArrayProperties != nil {
		config["array_size"] = batchParameters.ArrayProperties.Size
	}
	if batchParameters.RetryStrategy != nil {
		config["job_attempts"] = batchParameters.RetryStrategy.Attempts
	}
	result := []map[string]interface{}{config}
	return result
}

func flattenTargetKinesisParameters(kinesisParameters *types.KinesisParameters) []map[string]interface{} {
	config := make(map[string]interface{})
	config["partition_key_path"] = aws.ToString(kinesisParameters.PartitionKeyPath)
	result := []map[string]interface{}{config}
	return result
}

func flattenTargetSageMakerPipelineParameters(sageMakerParameters *types.SageMakerPipelineParameters) []map[string]interface{} {
	config := make(map[string]interface{})
	config["pipeline_parameter_list"] = flattenTargetSageMakerPipelineParameter(sageMakerParameters.PipelineParameterList)
	result := []map[string]interface{}{config}
	return result
}

func flattenTargetSageMakerPipelineParameter(pcs []types.SageMakerPipelineParameter) []map[string]interface{} {
	if len(pcs) == 0 {
		return nil
	}
	results := make([]map[string]interface{}, 0)
	for _, pc := range pcs {
		c := make(map[string]interface{})
		c[names.AttrName] = aws.ToString(pc.Name)
		c[names.AttrValue] = aws.ToString(pc.Value)

		results = append(results, c)
	}
	return results
}

func flattenTargetSQSParameters(sqsParameters *types.SqsParameters) []map[string]interface{} {
	config := make(map[string]interface{})
	config["message_group_id"] = aws.ToString(sqsParameters.MessageGroupId)
	result := []map[string]interface{}{config}
	return result
}

func flattenTargetHTTPParameters(apiObject *types.HttpParameters) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if v := apiObject.HeaderParameters; v != nil {
		tfMap["header_parameters"] = v
	}

	if v := apiObject.PathParameterValues; v != nil {
		tfMap["path_parameter_values"] = v
	}

	if v := apiObject.QueryStringParameters; v != nil {
		tfMap["query_string_parameters"] = v
	}

	return tfMap
}

func flattenInputTransformer(inputTransformer *types.InputTransformer) []map[string]interface{} {
	config := make(map[string]interface{})
	config["input_template"] = aws.ToString(inputTransformer.InputTemplate)
	config["input_paths"] = inputTransformer.InputPathsMap

	result := []map[string]interface{}{config}
	return result
}

func flattenTargetRetryPolicy(rp *types.RetryPolicy) []map[string]interface{} {
	config := make(map[string]interface{})

	config["maximum_event_age_in_seconds"] = aws.ToInt32(rp.MaximumEventAgeInSeconds)
	config["maximum_retry_attempts"] = aws.ToInt32(rp.MaximumRetryAttempts)

	result := []map[string]interface{}{config}
	return result
}

func flattenTargetDeadLetterConfig(dlc *types.DeadLetterConfig) []map[string]interface{} {
	config := make(map[string]interface{})

	config[names.AttrARN] = aws.ToString(dlc.Arn)

	result := []map[string]interface{}{config}
	return result
}

func expandTargetPlacementConstraints(tfList []interface{}) []types.PlacementConstraint {
	if len(tfList) == 0 {
		return nil
	}

	var result []types.PlacementConstraint

	for _, tfMapRaw := range tfList {
		if tfMapRaw == nil {
			continue
		}

		tfMap := tfMapRaw.(map[string]interface{})

		apiObject := types.PlacementConstraint{}

		if v, ok := tfMap[names.AttrExpression].(string); ok && v != "" {
			apiObject.Expression = aws.String(v)
		}

		if v, ok := tfMap[names.AttrType].(string); ok && v != "" {
			apiObject.Type = types.PlacementConstraintType(v)
		}

		result = append(result, apiObject)
	}

	return result
}

func expandTargetPlacementStrategies(tfList []interface{}) []types.PlacementStrategy {
	if len(tfList) == 0 {
		return nil
	}

	var result []types.PlacementStrategy

	for _, tfMapRaw := range tfList {
		if tfMapRaw == nil {
			continue
		}

		tfMap := tfMapRaw.(map[string]interface{})

		apiObject := types.PlacementStrategy{}

		if v, ok := tfMap[names.AttrField].(string); ok && v != "" {
			apiObject.Field = aws.String(v)
		}

		if v, ok := tfMap[names.AttrType].(string); ok && v != "" {
			apiObject.Type = types.PlacementStrategyType(v)
		}

		result = append(result, apiObject)
	}

	return result
}

func expandTargetCapacityProviderStrategy(tfList []interface{}) []types.CapacityProviderStrategyItem {
	if len(tfList) == 0 {
		return nil
	}

	var result []types.CapacityProviderStrategyItem

	for _, tfMapRaw := range tfList {
		if tfMapRaw == nil {
			continue
		}

		cp := tfMapRaw.(map[string]interface{})

		apiObject := types.CapacityProviderStrategyItem{}

		if val, ok := cp["base"]; ok {
			apiObject.Base = int32(val.(int))
		}

		if val, ok := cp[names.AttrWeight]; ok {
			apiObject.Weight = int32(val.(int))
		}

		if val, ok := cp["capacity_provider"]; ok {
			apiObject.CapacityProvider = aws.String(val.(string))
		}

		result = append(result, apiObject)
	}

	return result
}

func flattenTargetPlacementConstraints(pcs []types.PlacementConstraint) []map[string]interface{} {
	if len(pcs) == 0 {
		return nil
	}
	results := make([]map[string]interface{}, 0)
	for _, pc := range pcs {
		c := make(map[string]interface{})
		c[names.AttrType] = pc.Type
		if pc.Expression != nil {
			c[names.AttrExpression] = aws.ToString(pc.Expression)
		}

		results = append(results, c)
	}
	return results
}

func flattenTargetPlacementStrategies(pcs []types.PlacementStrategy) []map[string]interface{} {
	if len(pcs) == 0 {
		return nil
	}
	results := make([]map[string]interface{}, 0)
	for _, pc := range pcs {
		c := make(map[string]interface{})
		c[names.AttrType] = pc.Type
		if pc.Field != nil {
			c[names.AttrField] = aws.ToString(pc.Field)
		}

		results = append(results, c)
	}
	return results
}

func flattenTargetCapacityProviderStrategy(cps []types.CapacityProviderStrategyItem) []map[string]interface{} {
	if cps == nil {
		return nil
	}
	results := make([]map[string]interface{}, 0)
	for _, cp := range cps {
		s := make(map[string]interface{})
		s["capacity_provider"] = aws.ToString(cp.CapacityProvider)
		s[names.AttrWeight] = cp.Weight
		s["base"] = cp.Base
		results = append(results, s)
	}
	return results
}
