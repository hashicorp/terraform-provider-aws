// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package fis

import (
	"context"
	"errors"
	"log"
	"time"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/fis"
	awstypes "github.com/aws/aws-sdk-go-v2/service/fis/types"
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
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_fis_experiment_template", name="Experiment Template")
// @Tags
func resourceExperimentTemplate() *schema.Resource {
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
			names.AttrAction: {
				Type:     schema.TypeSet,
				Required: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"action_id": {
							Type:     schema.TypeString,
							Required: true,
							ValidateFunc: validation.All(
								validation.StringLenBetween(0, 128),
								validation.StringMatch(regexache.MustCompile(`^aws:[0-9a-z-]+:[0-9A-Za-z/-]+$`), "must be in the format of aws:service-name:action-name"),
							),
						},
						names.AttrDescription: {
							Type:         schema.TypeString,
							Optional:     true,
							ValidateFunc: validation.StringLenBetween(0, 512),
						},
						names.AttrName: {
							Type:         schema.TypeString,
							Required:     true,
							ValidateFunc: validation.StringLenBetween(0, 64),
						},
						names.AttrParameter: {
							Type:     schema.TypeSet,
							Optional: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									names.AttrKey: {
										Type:         schema.TypeString,
										Required:     true,
										ValidateFunc: validation.StringLenBetween(0, 64),
									},
									names.AttrValue: {
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
							Elem: &schema.Schema{
								Type:         schema.TypeString,
								ValidateFunc: validation.StringLenBetween(0, 64),
							},
						},
						names.AttrTarget: {
							Type:     schema.TypeList,
							Optional: true,
							MaxItems: 1, //API will accept more, but return only 1
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									names.AttrKey: {
										Type:         schema.TypeString,
										Required:     true,
										ValidateFunc: validExperimentTemplateActionTargetKey(),
									},
									names.AttrValue: {
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
			names.AttrDescription: {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringLenBetween(0, 512),
			},
			"experiment_options": {
				Type:     schema.TypeList,
				Optional: true,
				Computed: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"account_targeting": {
							Type:             schema.TypeString,
							Optional:         true,
							ValidateDiagFunc: enum.Validate[awstypes.AccountTargeting](),
						},
						"empty_target_resolution_mode": {
							Type:             schema.TypeString,
							Optional:         true,
							ValidateDiagFunc: enum.Validate[awstypes.EmptyTargetResolutionMode](),
						},
					},
				},
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
									names.AttrBucketName: {
										Type:     schema.TypeString,
										Required: true,
									},
									names.AttrPrefix: {
										Type:     schema.TypeString,
										Optional: true,
									},
								},
							},
						},
					},
				},
			},
			names.AttrRoleARN: {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: verify.ValidARN,
			},
			"stop_condition": {
				Type:     schema.TypeSet,
				Required: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						names.AttrSource: {
							Type:         schema.TypeString,
							Required:     true,
							ValidateFunc: validExperimentTemplateStopConditionSource(),
						},
						names.AttrValue: {
							Type:         schema.TypeString,
							Optional:     true,
							ValidateFunc: verify.ValidARN,
						},
					},
				},
			},
			names.AttrTarget: {
				Type:     schema.TypeSet,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						names.AttrFilter: {
							Type:     schema.TypeList,
							Optional: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									names.AttrPath: {
										Type:         schema.TypeString,
										Required:     true,
										ValidateFunc: validation.StringLenBetween(0, 256),
									},
									names.AttrValues: {
										Type:     schema.TypeSet,
										Required: true,
										Elem: &schema.Schema{
											Type:         schema.TypeString,
											ValidateFunc: validation.StringLenBetween(0, 128),
										},
									},
								},
							},
						},
						names.AttrName: {
							Type:         schema.TypeString,
							Required:     true,
							ValidateFunc: validation.StringLenBetween(0, 64),
						},
						names.AttrParameters: {
							Type:     schema.TypeMap,
							Optional: true,
							Elem:     &schema.Schema{Type: schema.TypeString},
						},
						"resource_arns": {
							Type:     schema.TypeSet,
							Optional: true,
							MaxItems: 5,
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
									names.AttrKey: {
										Type:         schema.TypeString,
										Required:     true,
										ValidateFunc: validation.StringLenBetween(0, 128),
									},
									names.AttrValue: {
										Type:         schema.TypeString,
										Required:     true,
										ValidateFunc: validation.StringLenBetween(0, 256),
									},
								},
							},
						},
						names.AttrResourceType: {
							Type:         schema.TypeString,
							Required:     true,
							ValidateFunc: validation.StringLenBetween(0, 64),
						},
						"selection_mode": {
							Type:     schema.TypeString,
							Required: true,
							ValidateFunc: validation.All(
								validation.StringLenBetween(0, 64),
								validation.StringMatch(regexache.MustCompile(`^(ALL|COUNT\(\d+\)|PERCENT\(\d+\))$`), "must be one of ALL, COUNT(number), PERCENT(number)"),
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
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).FISClient(ctx)

	input := &fis.CreateExperimentTemplateInput{
		Actions:          expandExperimentTemplateActions(d.Get(names.AttrAction).(*schema.Set)),
		ClientToken:      aws.String(id.UniqueId()),
		Description:      aws.String(d.Get(names.AttrDescription).(string)),
		LogConfiguration: expandExperimentTemplateLogConfiguration(d.Get("log_configuration").([]interface{})),
		RoleArn:          aws.String(d.Get(names.AttrRoleARN).(string)),
		StopConditions:   expandExperimentTemplateStopConditions(d.Get("stop_condition").(*schema.Set)),
		Tags:             getTagsIn(ctx),
	}

	if v, ok := d.GetOk("experiment_options"); ok {
		input.ExperimentOptions = expandCreateExperimentTemplateExperimentOptionsInput(v.([]interface{}))
	}

	if targets, err := expandExperimentTemplateTargets(d.Get(names.AttrTarget).(*schema.Set)); err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	} else {
		input.Targets = targets
	}

	output, err := conn.CreateExperimentTemplate(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating FIS Experiment Template: %s", err)
	}

	d.SetId(aws.ToString(output.ExperimentTemplate.Id))

	return append(diags, resourceExperimentTemplateRead(ctx, d, meta)...)
}

func resourceExperimentTemplateRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).FISClient(ctx)

	experimentTemplate, err := findExperimentTemplateByID(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] FIS Experiment Template (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading FIS Experiment Template (%s): %s", d.Id(), err)
	}

	d.SetId(aws.ToString(experimentTemplate.Id))
	if err := d.Set(names.AttrAction, flattenExperimentTemplateActions(experimentTemplate.Actions)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting action: %s", err)
	}
	d.Set(names.AttrRoleARN, experimentTemplate.RoleArn)
	d.Set(names.AttrDescription, experimentTemplate.Description)
	if err := d.Set("experiment_options", flattenExperimentTemplateExperimentOptions(experimentTemplate.ExperimentOptions)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting experiment_options: %s", err)
	}
	if err := d.Set("log_configuration", flattenExperimentTemplateLogConfiguration(experimentTemplate.LogConfiguration)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting log_configuration: %s", err)
	}
	if err := d.Set("stop_condition", flattenExperimentTemplateStopConditions(experimentTemplate.StopConditions)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting stop_condition: %s", err)
	}
	if err := d.Set(names.AttrTarget, flattenExperimentTemplateTargets(experimentTemplate.Targets)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting target: %s", err)
	}

	setTagsOut(ctx, experimentTemplate.Tags)

	return diags
}

func resourceExperimentTemplateUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	conn := meta.(*conns.AWSClient).FISClient(ctx)

	if d.HasChangesExcept(names.AttrTags, names.AttrTagsAll) {
		input := &fis.UpdateExperimentTemplateInput{
			Id: aws.String(d.Id()),
		}

		if d.HasChange(names.AttrAction) {
			input.Actions = expandExperimentTemplateActionsForUpdate(d.Get(names.AttrAction).(*schema.Set))
		}

		if d.HasChange(names.AttrDescription) {
			input.Description = aws.String(d.Get(names.AttrDescription).(string))
		}

		if d.HasChange("experiment_options") {
			input.ExperimentOptions = expandUpdateExperimentTemplateExperimentOptionsInput(d.Get("experiment_options").([]interface{}))
		}

		if d.HasChange("log_configuration") {
			config := expandExperimentTemplateLogConfigurationForUpdate(d.Get("log_configuration").([]interface{}))
			input.LogConfiguration = config
		}

		if d.HasChange(names.AttrRoleARN) {
			input.RoleArn = aws.String(d.Get(names.AttrRoleARN).(string))
		}

		if d.HasChange("stop_condition") {
			input.StopConditions = expandExperimentTemplateStopConditionsForUpdate(d.Get("stop_condition").(*schema.Set))
		}

		if d.HasChange(names.AttrTarget) {
			if targets, err := expandExperimentTemplateTargetsForUpdate(d.Get(names.AttrTarget).(*schema.Set)); err != nil {
				return sdkdiag.AppendFromErr(diags, err)
			} else {
				input.Targets = targets
			}
		}

		_, err := conn.UpdateExperimentTemplate(ctx, input)

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "updating FIS Experiment Template (%s): %s", d.Id(), err)
		}
	}

	return append(diags, resourceExperimentTemplateRead(ctx, d, meta)...)
}

func resourceExperimentTemplateDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).FISClient(ctx)

	log.Printf("[DEBUG] Deleting FIS Experiment Template: %s", d.Id())
	_, err := conn.DeleteExperimentTemplate(ctx, &fis.DeleteExperimentTemplateInput{
		Id: aws.String(d.Id()),
	})

	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting FIS Experiment Template (%s): %s", d.Id(), err)
	}

	return diags
}

func findExperimentTemplateByID(ctx context.Context, conn *fis.Client, id string) (*awstypes.ExperimentTemplate, error) {
	input := &fis.GetExperimentTemplateInput{
		Id: aws.String(id),
	}

	output, err := conn.GetExperimentTemplate(ctx, input)

	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil || output.ExperimentTemplate == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	return output.ExperimentTemplate, nil
}

func expandExperimentTemplateActions(l *schema.Set) map[string]awstypes.CreateExperimentTemplateActionInput {
	if l.Len() == 0 {
		return nil
	}

	attrs := make(map[string]awstypes.CreateExperimentTemplateActionInput, l.Len())

	for _, m := range l.List() {
		raw := m.(map[string]interface{})
		config := awstypes.CreateExperimentTemplateActionInput{}

		if v, ok := raw["action_id"].(string); ok && v != "" {
			config.ActionId = aws.String(v)
		}

		if v, ok := raw[names.AttrDescription].(string); ok && v != "" {
			config.Description = aws.String(v)
		}

		if v, ok := raw[names.AttrParameter].(*schema.Set); ok && v.Len() > 0 {
			config.Parameters = expandExperimentTemplateActionParameteres(v)
		}

		if v, ok := raw["start_after"].(*schema.Set); ok && v.Len() > 0 {
			config.StartAfter = flex.ExpandStringValueSet(v)
		}

		if v, ok := raw[names.AttrTarget].([]interface{}); ok && len(v) > 0 {
			config.Targets = expandExperimentTemplateActionTargets(v)
		}

		if v, ok := raw[names.AttrName].(string); ok && v != "" {
			attrs[v] = config
		}
	}

	return attrs
}

func expandExperimentTemplateActionsForUpdate(l *schema.Set) map[string]awstypes.UpdateExperimentTemplateActionInputItem {
	if l.Len() == 0 {
		return nil
	}

	attrs := make(map[string]awstypes.UpdateExperimentTemplateActionInputItem, l.Len())

	for _, m := range l.List() {
		raw := m.(map[string]interface{})
		config := awstypes.UpdateExperimentTemplateActionInputItem{}

		if v, ok := raw["action_id"].(string); ok && v != "" {
			config.ActionId = aws.String(v)
		}

		if v, ok := raw[names.AttrDescription].(string); ok && v != "" {
			config.Description = aws.String(v)
		}

		if v, ok := raw[names.AttrParameter].(*schema.Set); ok && v.Len() > 0 {
			config.Parameters = expandExperimentTemplateActionParameteres(v)
		}

		if v, ok := raw["start_after"].(*schema.Set); ok && v.Len() > 0 {
			config.StartAfter = flex.ExpandStringValueSet(v)
		}

		if v, ok := raw[names.AttrTarget].([]interface{}); ok && len(v) > 0 {
			config.Targets = expandExperimentTemplateActionTargets(v)
		}

		if v, ok := raw[names.AttrName].(string); ok && v != "" {
			attrs[v] = config
		}
	}

	return attrs
}

func expandCreateExperimentTemplateExperimentOptionsInput(tfMap []interface{}) *awstypes.CreateExperimentTemplateExperimentOptionsInput {
	if len(tfMap) == 0 || tfMap[0] == nil {
		return nil
	}

	apiObject := &awstypes.CreateExperimentTemplateExperimentOptionsInput{}

	m := tfMap[0].(map[string]interface{})

	if v, ok := m["account_targeting"].(string); ok {
		apiObject.AccountTargeting = awstypes.AccountTargeting(v)
	}

	if v, ok := m["empty_target_resolution_mode"].(string); ok {
		apiObject.EmptyTargetResolutionMode = awstypes.EmptyTargetResolutionMode(v)
	}

	return apiObject
}

func expandUpdateExperimentTemplateExperimentOptionsInput(tfMap []interface{}) *awstypes.UpdateExperimentTemplateExperimentOptionsInput {
	if len(tfMap) == 0 || tfMap[0] == nil {
		return nil
	}

	m := tfMap[0].(map[string]interface{})

	apiObject := &awstypes.UpdateExperimentTemplateExperimentOptionsInput{}

	if v, ok := m["empty_target_resolution_mode"].(string); ok {
		apiObject.EmptyTargetResolutionMode = awstypes.EmptyTargetResolutionMode(v)
	}

	return apiObject
}

func flattenExperimentTemplateExperimentOptions(apiObject *awstypes.ExperimentTemplateExperimentOptions) []map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := make([]map[string]interface{}, 1)
	tfMap[0] = make(map[string]interface{})
	if v := apiObject.AccountTargeting; v != "" {
		tfMap[0]["account_targeting"] = v
	}

	if v := apiObject.EmptyTargetResolutionMode; v != "" {
		tfMap[0]["empty_target_resolution_mode"] = v
	}

	return tfMap
}

func expandExperimentTemplateStopConditions(l *schema.Set) []awstypes.CreateExperimentTemplateStopConditionInput {
	if l.Len() == 0 {
		return nil
	}

	items := []awstypes.CreateExperimentTemplateStopConditionInput{}

	for _, m := range l.List() {
		raw := m.(map[string]interface{})
		config := awstypes.CreateExperimentTemplateStopConditionInput{}

		if v, ok := raw[names.AttrSource].(string); ok && v != "" {
			config.Source = aws.String(v)
		}

		if v, ok := raw[names.AttrValue].(string); ok && v != "" {
			config.Value = aws.String(v)
		}

		items = append(items, config)
	}

	return items
}

func expandExperimentTemplateLogConfiguration(l []interface{}) *awstypes.CreateExperimentTemplateLogConfigurationInput {
	if len(l) == 0 {
		return nil
	}

	raw := l[0].(map[string]interface{})

	config := awstypes.CreateExperimentTemplateLogConfigurationInput{
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

func expandExperimentTemplateCloudWatchLogsConfiguration(l []interface{}) *awstypes.ExperimentTemplateCloudWatchLogsLogConfigurationInput {
	if len(l) == 0 {
		return nil
	}

	raw := l[0].(map[string]interface{})

	config := awstypes.ExperimentTemplateCloudWatchLogsLogConfigurationInput{
		LogGroupArn: aws.String(raw["log_group_arn"].(string)),
	}
	return &config
}

func expandExperimentTemplateS3Configuration(l []interface{}) *awstypes.ExperimentTemplateS3LogConfigurationInput {
	if len(l) == 0 {
		return nil
	}

	raw := l[0].(map[string]interface{})

	config := awstypes.ExperimentTemplateS3LogConfigurationInput{
		BucketName: aws.String(raw[names.AttrBucketName].(string)),
	}
	if v, ok := raw[names.AttrPrefix].(string); ok && v != "" {
		config.Prefix = aws.String(v)
	}

	return &config
}

func expandExperimentTemplateStopConditionsForUpdate(l *schema.Set) []awstypes.UpdateExperimentTemplateStopConditionInput {
	if l.Len() == 0 {
		return nil
	}

	items := []awstypes.UpdateExperimentTemplateStopConditionInput{}

	for _, m := range l.List() {
		raw := m.(map[string]interface{})
		config := awstypes.UpdateExperimentTemplateStopConditionInput{}

		if v, ok := raw[names.AttrSource].(string); ok && v != "" {
			config.Source = aws.String(v)
		}

		if v, ok := raw[names.AttrValue].(string); ok && v != "" {
			config.Value = aws.String(v)
		}

		items = append(items, config)
	}

	return items
}

func expandExperimentTemplateTargets(l *schema.Set) (map[string]awstypes.CreateExperimentTemplateTargetInput, error) {
	if l.Len() == 0 {
		//Even though a template with no targets is valid (eg. containing just aws:fis:wait) and the API reference states that targets is not required, the key still needs to be present.
		return map[string]awstypes.CreateExperimentTemplateTargetInput{}, nil
	}

	attrs := make(map[string]awstypes.CreateExperimentTemplateTargetInput, l.Len())

	for _, m := range l.List() {
		raw := m.(map[string]interface{})
		config := awstypes.CreateExperimentTemplateTargetInput{}
		var hasSeenResourceArns bool

		if v, ok := raw[names.AttrFilter].([]interface{}); ok && len(v) > 0 {
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
		}

		if v, ok := raw[names.AttrResourceType].(string); ok && v != "" {
			config.ResourceType = aws.String(v)
		}

		if v, ok := raw["selection_mode"].(string); ok && v != "" {
			config.SelectionMode = aws.String(v)
		}

		if v, ok := raw[names.AttrParameters].(map[string]interface{}); ok && len(v) > 0 {
			config.Parameters = flex.ExpandStringValueMap(v)
		}

		if v, ok := raw[names.AttrName].(string); ok && v != "" {
			attrs[v] = config
		}
	}

	return attrs, nil
}

func expandExperimentTemplateTargetsForUpdate(l *schema.Set) (map[string]awstypes.UpdateExperimentTemplateTargetInput, error) {
	if l.Len() == 0 {
		return nil, nil
	}

	attrs := make(map[string]awstypes.UpdateExperimentTemplateTargetInput, l.Len())

	for _, m := range l.List() {
		raw := m.(map[string]interface{})
		config := awstypes.UpdateExperimentTemplateTargetInput{}
		var hasSeenResourceArns bool

		if v, ok := raw[names.AttrFilter].([]interface{}); ok && len(v) > 0 {
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
		}

		if v, ok := raw[names.AttrResourceType].(string); ok && v != "" {
			config.ResourceType = aws.String(v)
		}

		if v, ok := raw["selection_mode"].(string); ok && v != "" {
			config.SelectionMode = aws.String(v)
		}

		if v, ok := raw[names.AttrParameters].(map[string]interface{}); ok && len(v) > 0 {
			config.Parameters = flex.ExpandStringValueMap(v)
		}

		if v, ok := raw[names.AttrName].(string); ok && v != "" {
			attrs[v] = config
		}
	}

	return attrs, nil
}

func expandExperimentTemplateLogConfigurationForUpdate(l []interface{}) *awstypes.UpdateExperimentTemplateLogConfigurationInput {
	if len(l) == 0 {
		return &awstypes.UpdateExperimentTemplateLogConfigurationInput{}
	}

	raw := l[0].(map[string]interface{})
	config := awstypes.UpdateExperimentTemplateLogConfigurationInput{
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
			attrs[attr[names.AttrKey]] = attr[names.AttrValue]
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
			attrs[attr[names.AttrKey]] = attr[names.AttrValue]
		}
	}

	return attrs
}

func expandExperimentTemplateTargetFilters(l []interface{}) []awstypes.ExperimentTemplateTargetInputFilter {
	if len(l) == 0 || l[0] == nil {
		return nil
	}

	items := []awstypes.ExperimentTemplateTargetInputFilter{}

	for _, m := range l {
		raw := m.(map[string]interface{})
		config := awstypes.ExperimentTemplateTargetInputFilter{}

		if v, ok := raw[names.AttrPath].(string); ok && v != "" {
			config.Path = aws.String(v)
		}

		if v, ok := raw[names.AttrValues].(*schema.Set); ok && v.Len() > 0 {
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
			attrs[attr[names.AttrKey]] = attr[names.AttrValue]
		}
	}

	return attrs
}

func flattenExperimentTemplateActions(configured map[string]awstypes.ExperimentTemplateAction) []map[string]interface{} {
	dataResources := make([]map[string]interface{}, 0, len(configured))

	for k, v := range configured {
		item := make(map[string]interface{})
		item["action_id"] = aws.ToString(v.ActionId)
		item[names.AttrDescription] = aws.ToString(v.Description)
		item[names.AttrParameter] = flattenExperimentTemplateActionParameters(v.Parameters)
		item["start_after"] = v.StartAfter
		item[names.AttrTarget] = flattenExperimentTemplateActionTargets(v.Targets)

		item[names.AttrName] = k

		dataResources = append(dataResources, item)
	}

	return dataResources
}

func flattenExperimentTemplateStopConditions(configured []awstypes.ExperimentTemplateStopCondition) []map[string]interface{} {
	dataResources := make([]map[string]interface{}, 0, len(configured))

	for _, v := range configured {
		item := make(map[string]interface{})
		item[names.AttrSource] = aws.ToString(v.Source)

		if aws.ToString(v.Value) != "" {
			item[names.AttrValue] = aws.ToString(v.Value)
		}

		dataResources = append(dataResources, item)
	}

	return dataResources
}

func flattenExperimentTemplateTargets(configured map[string]awstypes.ExperimentTemplateTarget) []map[string]interface{} {
	dataResources := make([]map[string]interface{}, 0, len(configured))

	for k, v := range configured {
		item := make(map[string]interface{})
		item[names.AttrFilter] = flattenExperimentTemplateTargetFilters(v.Filters)
		item["resource_arns"] = v.ResourceArns
		item["resource_tag"] = flattenExperimentTemplateTargetResourceTags(v.ResourceTags)
		item[names.AttrResourceType] = aws.ToString(v.ResourceType)
		item["selection_mode"] = aws.ToString(v.SelectionMode)
		item[names.AttrParameters] = v.Parameters

		item[names.AttrName] = k

		dataResources = append(dataResources, item)
	}

	return dataResources
}

func flattenExperimentTemplateLogConfiguration(configured *awstypes.ExperimentTemplateLogConfiguration) []map[string]interface{} {
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

func flattenCloudWatchLogsConfiguration(configured *awstypes.ExperimentTemplateCloudWatchLogsLogConfiguration) []map[string]interface{} {
	if configured == nil {
		return make([]map[string]interface{}, 0)
	}

	dataResources := make([]map[string]interface{}, 1)
	dataResources[0] = make(map[string]interface{})
	dataResources[0]["log_group_arn"] = configured.LogGroupArn

	return dataResources
}

func flattenS3Configuration(configured *awstypes.ExperimentTemplateS3LogConfiguration) []map[string]interface{} {
	if configured == nil {
		return make([]map[string]interface{}, 0)
	}

	dataResources := make([]map[string]interface{}, 1)
	dataResources[0] = make(map[string]interface{})
	dataResources[0][names.AttrBucketName] = configured.BucketName
	if aws.ToString(configured.Prefix) != "" {
		dataResources[0][names.AttrPrefix] = configured.Prefix
	}

	return dataResources
}

func flattenExperimentTemplateActionParameters(configured map[string]string) []map[string]interface{} {
	dataResources := make([]map[string]interface{}, 0, len(configured))

	for k, v := range configured {
		item := make(map[string]interface{})
		item[names.AttrKey] = k
		item[names.AttrValue] = v

		dataResources = append(dataResources, item)
	}

	return dataResources
}

func flattenExperimentTemplateActionTargets(configured map[string]string) []map[string]interface{} {
	dataResources := make([]map[string]interface{}, 0, len(configured))

	for k, v := range configured {
		item := make(map[string]interface{})
		item[names.AttrKey] = k
		item[names.AttrValue] = v
		dataResources = append(dataResources, item)
	}

	return dataResources
}

func flattenExperimentTemplateTargetFilters(configured []awstypes.ExperimentTemplateTargetFilter) []map[string]interface{} {
	dataResources := make([]map[string]interface{}, 0, len(configured))

	for _, v := range configured {
		item := make(map[string]interface{})
		item[names.AttrPath] = aws.ToString(v.Path)
		item[names.AttrValues] = v.Values

		dataResources = append(dataResources, item)
	}

	return dataResources
}

func flattenExperimentTemplateTargetResourceTags(configured map[string]string) []map[string]interface{} {
	dataResources := make([]map[string]interface{}, 0, len(configured))

	for k, v := range configured {
		item := make(map[string]interface{})
		item[names.AttrKey] = k
		item[names.AttrValue] = v

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
	allowedActionTargets := []string{
		"AutoScalingGroups",
		"Buckets",
		"Cluster",
		"Clusters",
		"DBInstances",
		"Instances",
		"Nodegroups",
		"Pods",
		"ReplicationGroups",
		"Roles",
		"SpotInstances",
		"Subnets",
		"Tables",
		"Tasks",
		"TransitGateways",
		"Volumes",
	}

	return validation.All(
		validation.StringLenBetween(0, 64),
		validation.StringInSlice(allowedActionTargets, false),
	)
}
