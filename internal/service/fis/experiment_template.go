// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package fis

import (
	"context"
	"errors"
	"regexp"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/fis"
	"github.com/aws/aws-sdk-go-v2/service/fis/types"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/id"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

const (
	ErrCodeNotFound           = 404
	ResNameExperimentTemplate = "Experiment Template"
)

// @SDKResource("aws_fis_experiment_template", name="Experiment Template")
// @Tags
func ResourceExperimentTemplate() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceExperimentTemplateCreate,
		ReadWithoutTimeout:   resourceExperimentTemplateRead,
		UpdateWithoutTimeout: resourceExperimentTemplateUpdate,
		DeleteWithoutTimeout: resourceExperimentTemplateDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(30 * time.Minute),
			Update: schema.DefaultTimeout(30 * time.Minute),
			Delete: schema.DefaultTimeout(30 * time.Minute),
		},

		CustomizeDiff: verify.SetTagsDiff,

		Schema: map[string]*schema.Schema{
			"action": {
				Type:     schema.TypeSet,
				Required: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"action_id": {
							Type:     schema.TypeString,
							Required: true,
							ValidateFunc: validation.All(
								validation.StringLenBetween(0, 128),
								validation.StringMatch(regexp.MustCompile(`^aws:[a-z0-9-]+:[a-zA-Z0-9/-]+$`), "must be in the format of aws:service-name:action-name"),
							),
						},
						"description": {
							Type:         schema.TypeString,
							Optional:     true,
							ValidateFunc: validation.StringLenBetween(0, 512),
						},
						"name": {
							Type:         schema.TypeString,
							Required:     true,
							ValidateFunc: validation.StringLenBetween(0, 64),
						},
						"parameter": {
							Type:     schema.TypeSet,
							Optional: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"key": {
										Type:         schema.TypeString,
										Required:     true,
										ValidateFunc: validation.StringLenBetween(0, 64),
									},
									"value": {
										Type:         schema.TypeString,
										Required:     true,
										ValidateFunc: validation.StringLenBetween(0, 1024),
									},
								},
							},
						},
						"start_after": {
							Type:     schema.TypeSet,
							Optional: true,
							Set:      schema.HashString,
							Elem: &schema.Schema{
								Type:         schema.TypeString,
								ValidateFunc: validation.StringLenBetween(0, 64),
							},
						},
						"target": {
							Type:     schema.TypeList,
							Optional: true,
							MaxItems: 1, //API will accept more, but return only 1
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"key": {
										Type:         schema.TypeString,
										Required:     true,
										ValidateFunc: validExperimentTemplateActionTargetKey(),
									},
									"value": {
										Type:         schema.TypeString,
										Required:     true,
										ValidateFunc: validation.StringLenBetween(0, 64),
									},
								},
							},
						},
					},
				},
			},
			"description": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringLenBetween(0, 512),
			},
			"log_configuration": {
				Type:     schema.TypeList,
				Optional: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"cloudwatch_logs_configuration": {
							Type:     schema.TypeList,
							Optional: true,
							MaxItems: 1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"log_group_arn": {
										Type:     schema.TypeString,
										Required: true,
									},
								},
							},
						},
						"log_schema_version": {
							Type:     schema.TypeInt,
							Required: true,
						},
						"s3_configuration": {
							Type:     schema.TypeList,
							Optional: true,
							MaxItems: 1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"bucket_name": {
										Type:     schema.TypeString,
										Required: true,
									},
									"prefix": {
										Type:     schema.TypeString,
										Optional: true,
									},
								},
							},
						},
					},
				},
			},
			"role_arn": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: verify.ValidARN,
			},
			"stop_condition": {
				Type:     schema.TypeSet,
				Required: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"source": {
							Type:         schema.TypeString,
							Required:     true,
							ValidateFunc: validExperimentTemplateStopConditionSource(),
						},
						"value": {
							Type:         schema.TypeString,
							Optional:     true,
							ValidateFunc: verify.ValidARN,
						},
					},
				},
			},
			"target": {
				Type:     schema.TypeSet,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"filter": {
							Type:     schema.TypeList,
							Optional: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"path": {
										Type:         schema.TypeString,
										Required:     true,
										ValidateFunc: validation.StringLenBetween(0, 256),
									},
									"values": {
										Type:     schema.TypeSet,
										Required: true,
										Set:      schema.HashString,
										Elem: &schema.Schema{
											Type:         schema.TypeString,
											ValidateFunc: validation.StringLenBetween(0, 128),
										},
									},
								},
							},
						},
						"name": {
							Type:         schema.TypeString,
							Required:     true,
							ValidateFunc: validation.StringLenBetween(0, 64),
						},
						"parameters": {
							Type:     schema.TypeMap,
							Optional: true,
							Elem:     &schema.Schema{Type: schema.TypeString},
						},
						"resource_arns": {
							Type:     schema.TypeSet,
							Optional: true,
							MaxItems: 5,
							Set:      schema.HashString,
							Elem: &schema.Schema{
								Type:         schema.TypeString,
								ValidateFunc: verify.ValidARN,
							},
						},
						"resource_tag": {
							Type:     schema.TypeSet,
							Optional: true,
							MaxItems: 50,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"key": {
										Type:         schema.TypeString,
										Required:     true,
										ValidateFunc: validation.StringLenBetween(0, 128),
									},
									"value": {
										Type:         schema.TypeString,
										Required:     true,
										ValidateFunc: validation.StringLenBetween(0, 256),
									},
								},
							},
						},
						"resource_type": {
							Type:         schema.TypeString,
							Required:     true,
							ValidateFunc: validation.StringLenBetween(0, 64),
						},
						"selection_mode": {
							Type:     schema.TypeString,
							Required: true,
							ValidateFunc: validation.All(
								validation.StringLenBetween(0, 64),
								validation.StringMatch(regexp.MustCompile(`^(ALL|COUNT\(\d+\)|PERCENT\(\d+\))$`), "must be one of ALL, COUNT(number), PERCENT(number)"),
							),
						},
					},
				},
			},
			names.AttrTags:    tftags.TagsSchemaForceNew(),
			names.AttrTagsAll: tftags.TagsSchemaComputed(),
		},
	}
}

func resourceExperimentTemplateCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).FISClient(ctx)

	input := &fis.CreateExperimentTemplateInput{
		Actions:          expandExperimentTemplateActions(d.Get("action").(*schema.Set)),
		ClientToken:      aws.String(id.UniqueId()),
		Description:      aws.String(d.Get("description").(string)),
		LogConfiguration: expandExperimentTemplateLogConfiguration(d.Get("log_configuration").([]interface{})),
		RoleArn:          aws.String(d.Get("role_arn").(string)),
		StopConditions:   expandExperimentTemplateStopConditions(d.Get("stop_condition").(*schema.Set)),
		Tags:             getTagsIn(ctx),
	}

	targets, err := expandExperimentTemplateTargets(d.Get("target").(*schema.Set))
	if err != nil {
		return create.DiagError(names.FIS, create.ErrActionCreating, ResNameExperimentTemplate, d.Get("description").(string), err)
	}
	input.Targets = targets

	output, err := conn.CreateExperimentTemplate(ctx, input)
	if err != nil {
		return create.DiagError(names.FIS, create.ErrActionCreating, ResNameExperimentTemplate, d.Get("description").(string), err)
	}

	d.SetId(aws.ToString(output.ExperimentTemplate.Id))

	return resourceExperimentTemplateRead(ctx, d, meta)
}

func resourceExperimentTemplateRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).FISClient(ctx)

	input := &fis.GetExperimentTemplateInput{Id: aws.String(d.Id())}
	out, err := conn.GetExperimentTemplate(ctx, input)

	var nf *types.ResourceNotFoundException
	if !d.IsNewResource() && errors.As(err, &nf) {
		create.LogNotFoundRemoveState(names.FIS, create.ErrActionReading, ResNameExperimentTemplate, d.Id())
		d.SetId("")
		return nil
	}

	if !d.IsNewResource() && tfawserr.ErrStatusCodeEquals(err, ErrCodeNotFound) {
		create.LogNotFoundRemoveState(names.FIS, create.ErrActionReading, ResNameExperimentTemplate, d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return create.DiagError(names.FIS, create.ErrActionReading, ResNameExperimentTemplate, d.Id(), err)
	}

	experimentTemplate := out.ExperimentTemplate
	if experimentTemplate == nil {
		return create.DiagError(names.FIS, create.ErrActionReading, ResNameExperimentTemplate, d.Id(), errors.New("empty result"))
	}

	d.SetId(aws.ToString(experimentTemplate.Id))
	d.Set("role_arn", experimentTemplate.RoleArn)
	d.Set("description", experimentTemplate.Description)

	if err := d.Set("action", flattenExperimentTemplateActions(experimentTemplate.Actions)); err != nil {
		return create.DiagSettingError(names.FIS, ResNameExperimentTemplate, d.Id(), "action", err)
	}

	if err := d.Set("log_configuration", flattenExperimentTemplateLogConfiguration(experimentTemplate.LogConfiguration)); err != nil {
		return create.DiagSettingError(names.FIS, ResNameExperimentTemplate, d.Id(), "log_configuration", err)
	}

	if err := d.Set("stop_condition", flattenExperimentTemplateStopConditions(experimentTemplate.StopConditions)); err != nil {
		return create.DiagSettingError(names.FIS, ResNameExperimentTemplate, d.Id(), "stop_condition", err)
	}

	if err := d.Set("target", flattenExperimentTemplateTargets(experimentTemplate.Targets)); err != nil {
		return create.DiagSettingError(names.FIS, ResNameExperimentTemplate, d.Id(), "target", err)
	}

	setTagsOut(ctx, experimentTemplate.Tags)

	return nil
}

func resourceExperimentTemplateUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).FISClient(ctx)

	if d.HasChangesExcept("tags", "tags_all") {
		input := &fis.UpdateExperimentTemplateInput{
			Id: aws.String(d.Id()),
		}

		if d.HasChange("action") {
			input.Actions = expandExperimentTemplateActionsForUpdate(d.Get("action").(*schema.Set))
		}

		if d.HasChange("description") {
			input.Description = aws.String(d.Get("description").(string))
		}

		if d.HasChange("log_configuration") {
			config := expandExperimentTemplateLogConfigurationForUpdate(d.Get("log_configuration").([]interface{}))
			input.LogConfiguration = config
		}

		if d.HasChange("role_arn") {
			input.RoleArn = aws.String(d.Get("role_arn").(string))
		}

		if d.HasChange("stop_condition") {
			input.StopConditions = expandExperimentTemplateStopConditionsForUpdate(d.Get("stop_condition").(*schema.Set))
		}

		if d.HasChange("target") {
			targets, err := expandExperimentTemplateTargetsForUpdate(d.Get("target").(*schema.Set))
			if err != nil {
				return create.DiagError(names.FIS, create.ErrActionUpdating, ResNameExperimentTemplate, d.Id(), err)
			}
			input.Targets = targets
		}

		_, err := conn.UpdateExperimentTemplate(ctx, input)
		if err != nil {
			return create.DiagError(names.FIS, create.ErrActionUpdating, ResNameExperimentTemplate, d.Id(), err)
		}
	}

	return resourceExperimentTemplateRead(ctx, d, meta)
}

func resourceExperimentTemplateDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).FISClient(ctx)
	_, err := conn.DeleteExperimentTemplate(ctx, &fis.DeleteExperimentTemplateInput{
		Id: aws.String(d.Id()),
	})

	var nf *types.ResourceNotFoundException
	if errors.As(err, &nf) {
		return nil
	}

	if tfawserr.ErrStatusCodeEquals(err, ErrCodeNotFound) {
		return nil
	}

	if err != nil {
		return create.DiagError(names.FIS, create.ErrActionDeleting, ResNameExperimentTemplate, d.Id(), err)
	}

	return nil
}

func expandExperimentTemplateActions(l *schema.Set) map[string]types.CreateExperimentTemplateActionInput {
	if l.Len() == 0 {
		return nil
	}

	attrs := make(map[string]types.CreateExperimentTemplateActionInput, l.Len())

	for _, m := range l.List() {
		raw := m.(map[string]interface{})
		config := types.CreateExperimentTemplateActionInput{}

		if v, ok := raw["action_id"].(string); ok && v != "" {
			config.ActionId = aws.String(v)
		}

		if v, ok := raw["description"].(string); ok && v != "" {
			config.Description = aws.String(v)
		}

		if v, ok := raw["parameter"].(*schema.Set); ok && v.Len() > 0 {
			config.Parameters = expandExperimentTemplateActionParameteres(v)
		}

		if v, ok := raw["start_after"].(*schema.Set); ok && v.Len() > 0 {
			config.StartAfter = flex.ExpandStringValueSet(v)
		}

		if v, ok := raw["target"].([]interface{}); ok && len(v) > 0 {
			config.Targets = expandExperimentTemplateActionTargets(v)
		}

		if v, ok := raw["name"].(string); ok && v != "" {
			attrs[v] = config
		}
	}

	return attrs
}

func expandExperimentTemplateActionsForUpdate(l *schema.Set) map[string]types.UpdateExperimentTemplateActionInputItem {
	if l.Len() == 0 {
		return nil
	}

	attrs := make(map[string]types.UpdateExperimentTemplateActionInputItem, l.Len())

	for _, m := range l.List() {
		raw := m.(map[string]interface{})
		config := types.UpdateExperimentTemplateActionInputItem{}

		if v, ok := raw["action_id"].(string); ok && v != "" {
			config.ActionId = aws.String(v)
		}

		if v, ok := raw["description"].(string); ok && v != "" {
			config.Description = aws.String(v)
		}

		if v, ok := raw["parameter"].(*schema.Set); ok && v.Len() > 0 {
			config.Parameters = expandExperimentTemplateActionParameteres(v)
		}

		if v, ok := raw["start_after"].(*schema.Set); ok && v.Len() > 0 {
			config.StartAfter = flex.ExpandStringValueSet(v)
		}

		if v, ok := raw["target"].([]interface{}); ok && len(v) > 0 {
			config.Targets = expandExperimentTemplateActionTargets(v)
		}

		if v, ok := raw["name"].(string); ok && v != "" {
			attrs[v] = config
		}
	}

	return attrs
}

func expandExperimentTemplateStopConditions(l *schema.Set) []types.CreateExperimentTemplateStopConditionInput {
	if l.Len() == 0 {
		return nil
	}

	items := []types.CreateExperimentTemplateStopConditionInput{}

	for _, m := range l.List() {
		raw := m.(map[string]interface{})
		config := types.CreateExperimentTemplateStopConditionInput{}

		if v, ok := raw["source"].(string); ok && v != "" {
			config.Source = aws.String(v)
		}

		if v, ok := raw["value"].(string); ok && v != "" {
			config.Value = aws.String(v)
		}

		items = append(items, config)
	}

	return items
}

func expandExperimentTemplateLogConfiguration(l []interface{}) *types.CreateExperimentTemplateLogConfigurationInput {
	if len(l) == 0 {
		return nil
	}

	raw := l[0].(map[string]interface{})

	config := types.CreateExperimentTemplateLogConfigurationInput{
		LogSchemaVersion: aws.Int32(int32(raw["log_schema_version"].(int))),
	}

	if v, ok := raw["cloudwatch_logs_configuration"].([]interface{}); ok && len(v) > 0 {
		config.CloudWatchLogsConfiguration = expandExperimentTemplateCloudWatchLogsConfiguration(v)
	}

	if v, ok := raw["s3_configuration"].([]interface{}); ok && len(v) > 0 {
		config.S3Configuration = expandExperimentTemplateS3Configuration(v)
	}

	return &config
}

func expandExperimentTemplateCloudWatchLogsConfiguration(l []interface{}) *types.ExperimentTemplateCloudWatchLogsLogConfigurationInput {
	if len(l) == 0 {
		return nil
	}

	raw := l[0].(map[string]interface{})

	config := types.ExperimentTemplateCloudWatchLogsLogConfigurationInput{
		LogGroupArn: aws.String(raw["log_group_arn"].(string)),
	}
	return &config
}

func expandExperimentTemplateS3Configuration(l []interface{}) *types.ExperimentTemplateS3LogConfigurationInput {
	if len(l) == 0 {
		return nil
	}

	raw := l[0].(map[string]interface{})

	config := types.ExperimentTemplateS3LogConfigurationInput{
		BucketName: aws.String(raw["bucket_name"].(string)),
	}
	if v, ok := raw["prefix"].(string); ok && v != "" {
		config.Prefix = aws.String(v)
	}

	return &config
}

func expandExperimentTemplateStopConditionsForUpdate(l *schema.Set) []types.UpdateExperimentTemplateStopConditionInput {
	if l.Len() == 0 {
		return nil
	}

	items := []types.UpdateExperimentTemplateStopConditionInput{}

	for _, m := range l.List() {
		raw := m.(map[string]interface{})
		config := types.UpdateExperimentTemplateStopConditionInput{}

		if v, ok := raw["source"].(string); ok && v != "" {
			config.Source = aws.String(v)
		}

		if v, ok := raw["value"].(string); ok && v != "" {
			config.Value = aws.String(v)
		}

		items = append(items, config)
	}

	return items
}

func expandExperimentTemplateTargets(l *schema.Set) (map[string]types.CreateExperimentTemplateTargetInput, error) {
	if l.Len() == 0 {
		//Even though a template with no targets is valid (eg. containing just aws:fis:wait) and the API reference states that targets is not required, the key still needs to be present.
		return map[string]types.CreateExperimentTemplateTargetInput{}, nil
	}

	attrs := make(map[string]types.CreateExperimentTemplateTargetInput, l.Len())

	for _, m := range l.List() {
		raw := m.(map[string]interface{})
		config := types.CreateExperimentTemplateTargetInput{}
		var hasSeenResourceArns bool
		var hasSeenResourceTag bool

		if v, ok := raw["filter"].([]interface{}); ok && len(v) > 0 {
			config.Filters = expandExperimentTemplateTargetFilters(v)
		}

		if v, ok := raw["resource_arns"].(*schema.Set); ok && v.Len() > 0 {
			config.ResourceArns = flex.ExpandStringValueSet(v)
			hasSeenResourceArns = true
		}

		if v, ok := raw["resource_tag"].(*schema.Set); ok && v.Len() > 0 {
			//FIXME Rework this and use ConflictsWith once it supports lists
			//https://github.com/hashicorp/terraform-plugin-sdk/issues/71
			if hasSeenResourceArns {
				return nil, errors.New("Only one of resource_arns, resource_tag can be set in a target block")
			}
			config.ResourceTags = expandExperimentTemplateTargetResourceTags(v)
			hasSeenResourceTag = true
		}

		if !hasSeenResourceArns && !hasSeenResourceTag {
			return nil, errors.New("A target block requires one of resource_arns, resource_tag")
		}

		if v, ok := raw["resource_type"].(string); ok && v != "" {
			config.ResourceType = aws.String(v)
		}

		if v, ok := raw["selection_mode"].(string); ok && v != "" {
			config.SelectionMode = aws.String(v)
		}

		if v, ok := raw["parameters"].(map[string]interface{}); ok && len(v) > 0 {
			config.Parameters = flex.ExpandStringValueMap(v)
		}

		if v, ok := raw["name"].(string); ok && v != "" {
			attrs[v] = config
		}
	}

	return attrs, nil
}

func expandExperimentTemplateTargetsForUpdate(l *schema.Set) (map[string]types.UpdateExperimentTemplateTargetInput, error) {
	if l.Len() == 0 {
		return nil, nil
	}

	attrs := make(map[string]types.UpdateExperimentTemplateTargetInput, l.Len())

	for _, m := range l.List() {
		raw := m.(map[string]interface{})
		config := types.UpdateExperimentTemplateTargetInput{}
		var hasSeenResourceArns bool
		var hasSeenResourceTag bool

		if v, ok := raw["filter"].([]interface{}); ok && len(v) > 0 {
			config.Filters = expandExperimentTemplateTargetFilters(v)
		}

		if v, ok := raw["resource_arns"].(*schema.Set); ok && v.Len() > 0 {
			config.ResourceArns = flex.ExpandStringValueSet(v)
			hasSeenResourceArns = true
		}

		if v, ok := raw["resource_tag"].(*schema.Set); ok && v.Len() > 0 {
			//FIXME Rework this and use ConflictsWith once it supports lists
			//https://github.com/hashicorp/terraform-plugin-sdk/issues/71
			if hasSeenResourceArns {
				return nil, errors.New("Only one of resource_arns, resource_tag can be set in a target block")
			}
			config.ResourceTags = expandExperimentTemplateTargetResourceTags(v)
			hasSeenResourceTag = true
		}

		if !hasSeenResourceArns && !hasSeenResourceTag {
			return nil, errors.New("A target block requires one of resource_arns, resource_tag")
		}

		if v, ok := raw["resource_type"].(string); ok && v != "" {
			config.ResourceType = aws.String(v)
		}

		if v, ok := raw["selection_mode"].(string); ok && v != "" {
			config.SelectionMode = aws.String(v)
		}

		if v, ok := raw["parameters"].(map[string]interface{}); ok && len(v) > 0 {
			config.Parameters = flex.ExpandStringValueMap(v)
		}

		if v, ok := raw["name"].(string); ok && v != "" {
			attrs[v] = config
		}
	}

	return attrs, nil
}

func expandExperimentTemplateLogConfigurationForUpdate(l []interface{}) *types.UpdateExperimentTemplateLogConfigurationInput {
	if len(l) == 0 {
		return &types.UpdateExperimentTemplateLogConfigurationInput{}
	}

	raw := l[0].(map[string]interface{})
	config := types.UpdateExperimentTemplateLogConfigurationInput{
		LogSchemaVersion: aws.Int32(int32(raw["log_schema_version"].(int))),
	}
	if v, ok := raw["cloudwatch_logs_configuration"].([]interface{}); ok && len(v) > 0 {
		config.CloudWatchLogsConfiguration = expandExperimentTemplateCloudWatchLogsConfiguration(v)
	}

	if v, ok := raw["s3_configuration"].([]interface{}); ok && len(v) > 0 {
		config.S3Configuration = expandExperimentTemplateS3Configuration(v)
	}

	return &config
}

func expandExperimentTemplateActionParameteres(l *schema.Set) map[string]string {
	if l.Len() == 0 {
		return nil
	}

	attrs := make(map[string]string, l.Len())

	for _, m := range l.List() {
		if len(m.(map[string]interface{})) > 0 {
			attr := flex.ExpandStringValueMap(m.(map[string]interface{}))
			attrs[attr["key"]] = attr["value"]
		}
	}

	return attrs
}

func expandExperimentTemplateActionTargets(l []interface{}) map[string]string {
	if len(l) == 0 || l[0] == nil {
		return nil
	}

	attrs := make(map[string]string, len(l))

	for _, m := range l {
		if len(m.(map[string]interface{})) > 0 {
			attr := flex.ExpandStringValueMap(l[0].(map[string]interface{}))
			attrs[attr["key"]] = attr["value"]
		}
	}

	return attrs
}

func expandExperimentTemplateTargetFilters(l []interface{}) []types.ExperimentTemplateTargetInputFilter {
	if len(l) == 0 || l[0] == nil {
		return nil
	}

	items := []types.ExperimentTemplateTargetInputFilter{}

	for _, m := range l {
		raw := m.(map[string]interface{})
		config := types.ExperimentTemplateTargetInputFilter{}

		if v, ok := raw["path"].(string); ok && v != "" {
			config.Path = aws.String(v)
		}

		if v, ok := raw["values"].(*schema.Set); ok && v.Len() > 0 {
			config.Values = flex.ExpandStringValueSet(v)
		}

		items = append(items, config)
	}

	return items
}

func expandExperimentTemplateTargetResourceTags(l *schema.Set) map[string]string {
	if l.Len() == 0 {
		return nil
	}

	attrs := make(map[string]string, l.Len())

	for _, m := range l.List() {
		if len(m.(map[string]interface{})) > 0 {
			attr := flex.ExpandStringValueMap(m.(map[string]interface{}))
			attrs[attr["key"]] = attr["value"]
		}
	}

	return attrs
}

func flattenExperimentTemplateActions(configured map[string]types.ExperimentTemplateAction) []map[string]interface{} {
	dataResources := make([]map[string]interface{}, 0, len(configured))

	for k, v := range configured {
		item := make(map[string]interface{})
		item["action_id"] = aws.ToString(v.ActionId)
		item["description"] = aws.ToString(v.Description)
		item["parameter"] = flattenExperimentTemplateActionParameters(v.Parameters)
		item["start_after"] = v.StartAfter
		item["target"] = flattenExperimentTemplateActionTargets(v.Targets)

		item["name"] = k

		dataResources = append(dataResources, item)
	}

	return dataResources
}

func flattenExperimentTemplateStopConditions(configured []types.ExperimentTemplateStopCondition) []map[string]interface{} {
	dataResources := make([]map[string]interface{}, 0, len(configured))

	for _, v := range configured {
		item := make(map[string]interface{})
		item["source"] = aws.ToString(v.Source)

		if aws.ToString(v.Value) != "" {
			item["value"] = aws.ToString(v.Value)
		}

		dataResources = append(dataResources, item)
	}

	return dataResources
}

func flattenExperimentTemplateTargets(configured map[string]types.ExperimentTemplateTarget) []map[string]interface{} {
	dataResources := make([]map[string]interface{}, 0, len(configured))

	for k, v := range configured {
		item := make(map[string]interface{})
		item["filter"] = flattenExperimentTemplateTargetFilters(v.Filters)
		item["resource_arns"] = v.ResourceArns
		item["resource_tag"] = flattenExperimentTemplateTargetResourceTags(v.ResourceTags)
		item["resource_type"] = aws.ToString(v.ResourceType)
		item["selection_mode"] = aws.ToString(v.SelectionMode)
		item["parameters"] = v.Parameters

		item["name"] = k

		dataResources = append(dataResources, item)
	}

	return dataResources
}

func flattenExperimentTemplateLogConfiguration(configured *types.ExperimentTemplateLogConfiguration) []map[string]interface{} {
	if configured == nil {
		return make([]map[string]interface{}, 0)
	}

	dataResources := make([]map[string]interface{}, 1)
	dataResources[0] = make(map[string]interface{})
	dataResources[0]["log_schema_version"] = configured.LogSchemaVersion
	dataResources[0]["cloudwatch_logs_configuration"] = flattenCloudWatchLogsConfiguration(configured.CloudWatchLogsConfiguration)
	dataResources[0]["s3_configuration"] = flattenS3Configuration(configured.S3Configuration)

	return dataResources
}

func flattenCloudWatchLogsConfiguration(configured *types.ExperimentTemplateCloudWatchLogsLogConfiguration) []map[string]interface{} {
	if configured == nil {
		return make([]map[string]interface{}, 0)
	}

	dataResources := make([]map[string]interface{}, 1)
	dataResources[0] = make(map[string]interface{})
	dataResources[0]["log_group_arn"] = configured.LogGroupArn

	return dataResources
}

func flattenS3Configuration(configured *types.ExperimentTemplateS3LogConfiguration) []map[string]interface{} {
	if configured == nil {
		return make([]map[string]interface{}, 0)
	}

	dataResources := make([]map[string]interface{}, 1)
	dataResources[0] = make(map[string]interface{})
	dataResources[0]["bucket_name"] = configured.BucketName
	if aws.ToString(configured.Prefix) != "" {
		dataResources[0]["prefix"] = configured.Prefix
	}

	return dataResources
}

func flattenExperimentTemplateActionParameters(configured map[string]string) []map[string]interface{} {
	dataResources := make([]map[string]interface{}, 0, len(configured))

	for k, v := range configured {
		item := make(map[string]interface{})
		item["key"] = k
		item["value"] = v

		dataResources = append(dataResources, item)
	}

	return dataResources
}

func flattenExperimentTemplateActionTargets(configured map[string]string) []map[string]interface{} {
	dataResources := make([]map[string]interface{}, 0, len(configured))

	for k, v := range configured {
		item := make(map[string]interface{})
		item["key"] = k
		item["value"] = v
		dataResources = append(dataResources, item)
	}

	return dataResources
}

func flattenExperimentTemplateTargetFilters(configured []types.ExperimentTemplateTargetFilter) []map[string]interface{} {
	dataResources := make([]map[string]interface{}, 0, len(configured))

	for _, v := range configured {
		item := make(map[string]interface{})
		item["path"] = aws.ToString(v.Path)
		item["values"] = v.Values

		dataResources = append(dataResources, item)
	}

	return dataResources
}

func flattenExperimentTemplateTargetResourceTags(configured map[string]string) []map[string]interface{} {
	dataResources := make([]map[string]interface{}, 0, len(configured))

	for k, v := range configured {
		item := make(map[string]interface{})
		item["key"] = k
		item["value"] = v

		dataResources = append(dataResources, item)
	}

	return dataResources
}

func validExperimentTemplateStopConditionSource() schema.SchemaValidateFunc {
	allowedStopConditionSources := []string{
		"aws:cloudwatch:alarm",
		"none",
	}

	return validation.All(
		validation.StringInSlice(allowedStopConditionSources, false),
	)
}

func validExperimentTemplateActionTargetKey() schema.SchemaValidateFunc {
	// See https://docs.aws.amazon.com/fis/latest/userguide/actions.html#action-targets
	allowedStopConditionSources := []string{
		"Cluster",
		"Clusters",
		"DBInstances",
		"Instances",
		"Nodegroups",
		"Pods",
		"Roles",
		"SpotInstances",
		"Subnets",
		"Tasks",
		"Volumes",
	}

	return validation.All(
		validation.StringLenBetween(0, 64),
		validation.StringInSlice(allowedStopConditionSources, false),
	)
}
