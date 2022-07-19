package aws

import (
	"errors"
	"fmt"
	"log"
	"regexp"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/fis"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/keyvaluetags"
)

func resourceAwsFisExperimentTemplate() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsFisExperimentTemplateCreate,
		Read:   resourceAwsFisExperimentTemplateRead,
		Update: resourceAwsFisExperimentTemplateUpdate,
		Delete: resourceAwsFisExperimentTemplateDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},
		CustomizeDiff: SetTagsDiff,
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
										ValidateFunc: validateExperimentTemplateActionTargetKey(),
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
			"role_arn": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validateArn,
			},
			"stop_condition": {
				Type:     schema.TypeSet,
				Required: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"source": {
							Type:         schema.TypeString,
							Required:     true,
							ValidateFunc: validateExperimentTemplateStopConditionSource(),
						},
						"value": {
							Type:         schema.TypeString,
							Optional:     true,
							ValidateFunc: validateArn,
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
						"resource_arns": {
							Type:     schema.TypeSet,
							Optional: true,
							MaxItems: 5,
							Set:      schema.HashString,
							Elem: &schema.Schema{
								Type:         schema.TypeString,
								ValidateFunc: validateArn,
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
			"tags":     tagsSchemaForceNew(),
			"tags_all": tagsSchemaComputed(),
		},
	}
}

func resourceAwsFisExperimentTemplateCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).fisconn
	defaultTagsConfig := meta.(*AWSClient).DefaultTagsConfig
	tags := defaultTagsConfig.MergeTags(keyvaluetags.New(d.Get("tags").(map[string]interface{})))

	input := &fis.CreateExperimentTemplateInput{
		Actions:        expandAwsFisExperimentTemplateActions(d.Get("action").(*schema.Set)),
		ClientToken:    aws.String(resource.UniqueId()),
		Description:    aws.String(d.Get("description").(string)),
		RoleArn:        aws.String(d.Get("role_arn").(string)),
		StopConditions: expandAwsFisExperimentTemplateStopConditions(d.Get("stop_condition").(*schema.Set)),
		Tags:           tags.IgnoreAws().FisTags(),
	}

	targets, err := expandAwsFisExperimentTemplateTargets(d.Get("target").(*schema.Set))
	if err != nil {
		return fmt.Errorf("create Experiment Template failed: %w", err)
	}
	input.Targets = targets

	output, err := conn.CreateExperimentTemplate(input)
	if err != nil {
		return fmt.Errorf("create Experiment Template failed: %v", err)
	}

	d.SetId(aws.StringValue(output.ExperimentTemplate.Id))

	return resourceAwsFisExperimentTemplateRead(d, meta)
}

func resourceAwsFisExperimentTemplateRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).fisconn
	defaultTagsConfig := meta.(*AWSClient).DefaultTagsConfig
	ignoreTagsConfig := meta.(*AWSClient).IgnoreTagsConfig

	out, err := conn.GetExperimentTemplate(&fis.GetExperimentTemplateInput{Id: aws.String(d.Id())})
	if isAWSErrRequestFailureStatusCode(err, 404) {
		log.Printf("[WARN] Experiment Template %s not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return fmt.Errorf("describe Experiment Template failed: %w", err)
	}

	experimentTemplate := out.ExperimentTemplate
	if experimentTemplate == nil {
		log.Printf("[WARN] Experiment Template %s not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	d.SetId(aws.StringValue(experimentTemplate.Id))
	d.Set("role_arn", experimentTemplate.RoleArn)
	d.Set("description", experimentTemplate.Description)

	if err := d.Set("action", flattenAwsFisExperimentTemplateActions(experimentTemplate.Actions)); err != nil {
		return fmt.Errorf("error setting action: %w", err)
	}

	if err := d.Set("stop_condition", flattenAwsFisExperimentTemplateStopConditions(experimentTemplate.StopConditions)); err != nil {
		return fmt.Errorf("error setting stop_condition: %w", err)
	}

	if err := d.Set("target", flattenAwsFisExperimentTemplateTargets(experimentTemplate.Targets)); err != nil {
		return fmt.Errorf("error setting target: %w", err)
	}

	tags := keyvaluetags.FisKeyValueTags(experimentTemplate.Tags).IgnoreAws().IgnoreConfig(ignoreTagsConfig)

	//lintignore:AWSR002
	if err := d.Set("tags", tags.RemoveDefaultConfig(defaultTagsConfig).Map()); err != nil {
		return fmt.Errorf("error setting tags: %w", err)
	}

	if err := d.Set("tags_all", tags.Map()); err != nil {
		return fmt.Errorf("error setting tags_all: %w", err)
	}

	return nil
}

func resourceAwsFisExperimentTemplateUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).fisconn

	input := &fis.UpdateExperimentTemplateInput{
		Id: aws.String(d.Id()),
	}

	if d.HasChange("action") {
		input.Actions = expandAwsFisExperimentTemplateActionsForUpdate(d.Get("action").(*schema.Set))
	}

	if d.HasChange("description") {
		input.Description = aws.String(d.Get("description").(string))
	}

	if d.HasChange("role_arn") {
		input.RoleArn = aws.String(d.Get("role_arn").(string))
	}

	if d.HasChange("stop_condition") {
		input.StopConditions = expandAwsFisExperimentTemplateStopConditionsForUpdate(d.Get("stop_condition").(*schema.Set))
	}

	if d.HasChange("target") {
		targets, err := expandAwsFisExperimentTemplateTargetsForUpdate(d.Get("target").(*schema.Set))
		if err != nil {
			return fmt.Errorf("modify Experiment Template (%s) failed: %w", d.Id(), err)
		}
		input.Targets = targets
	}

	_, err := conn.UpdateExperimentTemplate(input)
	if err != nil {
		return fmt.Errorf("Updating Experiment Template failed: %w", err)
	}

	return resourceAwsFisExperimentTemplateRead(d, meta)
}

func resourceAwsFisExperimentTemplateDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).fisconn
	_, err := conn.DeleteExperimentTemplate(&fis.DeleteExperimentTemplateInput{
		Id: aws.String(d.Id()),
	})
	if err != nil {
		if isAWSErrRequestFailureStatusCode(err, 404) {
			log.Printf("[INFO] ExperimentTemplate %s could not be found. skipping delete.", d.Id())
			return nil
		}

		return fmt.Errorf("Deleting Experiment Template failed: %w", err)
	}

	return nil
}

func expandAwsFisExperimentTemplateActions(l *schema.Set) map[string]*fis.CreateExperimentTemplateActionInput {
	if l.Len() == 0 {
		return nil
	}

	attrs := make(map[string]*fis.CreateExperimentTemplateActionInput, l.Len())

	for _, m := range l.List() {
		raw := m.(map[string]interface{})
		config := &fis.CreateExperimentTemplateActionInput{}

		if v, ok := raw["action_id"].(string); ok && v != "" {
			config.ActionId = aws.String(v)
		}

		if v, ok := raw["description"].(string); ok && v != "" {
			config.Description = aws.String(v)
		}

		if v, ok := raw["parameter"].(*schema.Set); ok && v.Len() > 0 {
			config.Parameters = expandAwsFisExperimentTemplateActionParameteres(v)
		}

		if v, ok := raw["start_after"].(*schema.Set); ok && v.Len() > 0 {
			config.StartAfter = expandStringSet(v)
		}

		if v, ok := raw["target"].([]interface{}); ok && len(v) > 0 {
			config.Targets = expandAwsFisExperimentTemplateActionTargets(v)
		}

		if v, ok := raw["name"].(string); ok && v != "" {
			attrs[v] = config
		}
	}

	return attrs
}

func expandAwsFisExperimentTemplateActionsForUpdate(l *schema.Set) map[string]*fis.UpdateExperimentTemplateActionInputItem {
	if l.Len() == 0 {
		return nil
	}

	attrs := make(map[string]*fis.UpdateExperimentTemplateActionInputItem, l.Len())

	for _, m := range l.List() {
		raw := m.(map[string]interface{})
		config := &fis.UpdateExperimentTemplateActionInputItem{}

		if v, ok := raw["action_id"].(string); ok && v != "" {
			config.ActionId = aws.String(v)
		}

		if v, ok := raw["description"].(string); ok && v != "" {
			config.Description = aws.String(v)
		}

		if v, ok := raw["parameter"].(*schema.Set); ok && v.Len() > 0 {
			config.Parameters = expandAwsFisExperimentTemplateActionParameteres(v)
		}

		if v, ok := raw["start_after"].(*schema.Set); ok && v.Len() > 0 {
			config.StartAfter = expandStringSet(v)
		}

		if v, ok := raw["target"].([]interface{}); ok && len(v) > 0 {
			config.Targets = expandAwsFisExperimentTemplateActionTargets(v)
		}

		if v, ok := raw["name"].(string); ok && v != "" {
			attrs[v] = config
		}
	}

	return attrs
}

func expandAwsFisExperimentTemplateStopConditions(l *schema.Set) []*fis.CreateExperimentTemplateStopConditionInput {
	if l.Len() == 0 {
		return nil
	}

	items := []*fis.CreateExperimentTemplateStopConditionInput{}

	for _, m := range l.List() {
		raw := m.(map[string]interface{})
		config := &fis.CreateExperimentTemplateStopConditionInput{}

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

func expandAwsFisExperimentTemplateStopConditionsForUpdate(l *schema.Set) []*fis.UpdateExperimentTemplateStopConditionInput {
	if l.Len() == 0 {
		return nil
	}

	items := []*fis.UpdateExperimentTemplateStopConditionInput{}

	for _, m := range l.List() {
		raw := m.(map[string]interface{})
		config := &fis.UpdateExperimentTemplateStopConditionInput{}

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

func expandAwsFisExperimentTemplateTargets(l *schema.Set) (map[string]*fis.CreateExperimentTemplateTargetInput, error) {
	if l.Len() == 0 {
		//Even though a template with no targets is valid (eg. containing just aws:fis:wait) and the API reference states that targets is not required, the key still needs to be present.
		return map[string]*fis.CreateExperimentTemplateTargetInput{}, nil
	}

	attrs := make(map[string]*fis.CreateExperimentTemplateTargetInput, l.Len())

	for _, m := range l.List() {
		raw := m.(map[string]interface{})
		config := &fis.CreateExperimentTemplateTargetInput{}
		var hasSeenResourceArns bool
		var hasSeenResourceTag bool

		if v, ok := raw["filter"].([]interface{}); ok && len(v) > 0 {
			config.Filters = expandAwsFisExperimentTemplateTargetFilters(v)
		}

		if v, ok := raw["resource_arns"].(*schema.Set); ok && v.Len() > 0 {
			config.ResourceArns = expandStringSet(v)
			hasSeenResourceArns = true
		}

		if v, ok := raw["resource_tag"].(*schema.Set); ok && v.Len() > 0 {
			//FIXME Rework this and use ConflictsWith once it supports lists
			//https://github.com/hashicorp/terraform-plugin-sdk/issues/71
			if hasSeenResourceArns {
				return nil, errors.New("Only one of resource_arns, resource_tag can be set in a target block")
			}
			config.ResourceTags = expandAwsFisExperimentTemplateTargetResourceTags(v)
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

		if v, ok := raw["name"].(string); ok && v != "" {
			attrs[v] = config
		}
	}

	return attrs, nil
}

func expandAwsFisExperimentTemplateTargetsForUpdate(l *schema.Set) (map[string]*fis.UpdateExperimentTemplateTargetInput, error) {
	if l.Len() == 0 {
		return nil, nil
	}

	attrs := make(map[string]*fis.UpdateExperimentTemplateTargetInput, l.Len())

	for _, m := range l.List() {
		raw := m.(map[string]interface{})
		config := &fis.UpdateExperimentTemplateTargetInput{}
		var hasSeenResourceArns bool
		var hasSeenResourceTag bool

		if v, ok := raw["filter"].([]interface{}); ok && len(v) > 0 {
			config.Filters = expandAwsFisExperimentTemplateTargetFilters(v)
		}

		if v, ok := raw["resource_arns"].(*schema.Set); ok && v.Len() > 0 {
			config.ResourceArns = expandStringSet(v)
			hasSeenResourceArns = true
		}

		if v, ok := raw["resource_tag"].(*schema.Set); ok && v.Len() > 0 {
			//FIXME Rework this and use ConflictsWith once it supports lists
			//https://github.com/hashicorp/terraform-plugin-sdk/issues/71
			if hasSeenResourceArns {
				return nil, errors.New("Only one of resource_arns, resource_tag can be set in a target block")
			}
			config.ResourceTags = expandAwsFisExperimentTemplateTargetResourceTags(v)
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

		if v, ok := raw["name"].(string); ok && v != "" {
			attrs[v] = config
		}
	}

	return attrs, nil
}

func expandAwsFisExperimentTemplateActionParameteres(l *schema.Set) map[string]*string {
	if l.Len() == 0 {
		return nil
	}

	attrs := make(map[string]*string, l.Len())

	for _, m := range l.List() {
		if len(m.(map[string]interface{})) > 0 {
			attr := expandStringMap(m.(map[string]interface{}))
			attrs[aws.StringValue(attr["key"])] = attr["value"]
		}
	}

	return attrs
}

func expandAwsFisExperimentTemplateActionTargets(l []interface{}) map[string]*string {
	if len(l) == 0 || l[0] == nil {
		return nil
	}

	attrs := make(map[string]*string, len(l))

	for _, m := range l {
		if len(m.(map[string]interface{})) > 0 {
			attr := expandStringMap(l[0].(map[string]interface{}))
			attrs[aws.StringValue(attr["key"])] = attr["value"]
		}
	}

	return attrs
}

func expandAwsFisExperimentTemplateTargetFilters(l []interface{}) []*fis.ExperimentTemplateTargetInputFilter {
	if len(l) == 0 || l[0] == nil {
		return nil
	}

	items := []*fis.ExperimentTemplateTargetInputFilter{}

	for _, m := range l {
		raw := m.(map[string]interface{})
		config := &fis.ExperimentTemplateTargetInputFilter{}

		if v, ok := raw["path"].(string); ok && v != "" {
			config.Path = aws.String(v)
		}

		if v, ok := raw["values"].(*schema.Set); ok && v.Len() > 0 {
			config.Values = expandStringSet(v)
		}

		items = append(items, config)
	}

	return items
}

func expandAwsFisExperimentTemplateTargetResourceTags(l *schema.Set) map[string]*string {
	if l.Len() == 0 {
		return nil
	}

	attrs := make(map[string]*string, l.Len())

	for _, m := range l.List() {
		if len(m.(map[string]interface{})) > 0 {
			attr := expandStringMap(m.(map[string]interface{}))
			attrs[aws.StringValue(attr["key"])] = attr["value"]
		}
	}

	return attrs
}

func flattenAwsFisExperimentTemplateActions(configured map[string]*fis.ExperimentTemplateAction) []map[string]interface{} {
	dataResources := make([]map[string]interface{}, 0, len(configured))

	for k, v := range configured {
		item := make(map[string]interface{})
		item["action_id"] = aws.StringValue(v.ActionId)
		item["description"] = aws.StringValue(v.Description)
		item["parameter"] = flattenAwsFisExperimentTemplateActionParameters(v.Parameters)
		item["start_after"] = aws.StringValueSlice(v.StartAfter)
		item["target"] = flattenAwsFisExperimentTemplateActionTargets(v.Targets)

		item["name"] = k

		dataResources = append(dataResources, item)
	}

	return dataResources
}

func flattenAwsFisExperimentTemplateStopConditions(configured []*fis.ExperimentTemplateStopCondition) []map[string]interface{} {
	dataResources := make([]map[string]interface{}, 0, len(configured))

	for _, v := range configured {
		item := make(map[string]interface{})
		item["source"] = aws.StringValue(v.Source)

		if aws.StringValue(v.Value) != "" {
			item["value"] = aws.StringValue(v.Value)
		}

		dataResources = append(dataResources, item)
	}

	return dataResources
}

func flattenAwsFisExperimentTemplateTargets(configured map[string]*fis.ExperimentTemplateTarget) []map[string]interface{} {
	dataResources := make([]map[string]interface{}, 0, len(configured))

	for k, v := range configured {
		item := make(map[string]interface{})
		item["filter"] = flattenAwsFisExperimentTemplateTargetFilters(v.Filters)
		item["resource_arns"] = aws.StringValueSlice(v.ResourceArns)
		item["resource_tag"] = flattenAwsFisExperimentTemplateTargetResourceTags(v.ResourceTags)
		item["resource_type"] = aws.StringValue(v.ResourceType)
		item["selection_mode"] = aws.StringValue(v.SelectionMode)

		item["name"] = k

		dataResources = append(dataResources, item)
	}

	return dataResources
}

func flattenAwsFisExperimentTemplateActionParameters(configured map[string]*string) []map[string]interface{} {
	dataResources := make([]map[string]interface{}, 0, len(configured))

	for k, v := range configured {
		item := make(map[string]interface{})
		item["key"] = k
		item["value"] = aws.StringValue(v)

		dataResources = append(dataResources, item)
	}

	return dataResources
}

func flattenAwsFisExperimentTemplateActionTargets(configured map[string]*string) []map[string]interface{} {
	dataResources := make([]map[string]interface{}, 0, len(configured))

	for k, v := range configured {
		item := make(map[string]interface{})
		item["key"] = k
		item["value"] = aws.StringValue(v)
		dataResources = append(dataResources, item)
	}

	return dataResources
}

func flattenAwsFisExperimentTemplateTargetFilters(configured []*fis.ExperimentTemplateTargetFilter) []map[string]interface{} {
	dataResources := make([]map[string]interface{}, 0, len(configured))

	for _, v := range configured {
		item := make(map[string]interface{})
		item["path"] = aws.StringValue(v.Path)
		item["values"] = aws.StringValueSlice(v.Values)

		dataResources = append(dataResources, item)
	}

	return dataResources
}

func flattenAwsFisExperimentTemplateTargetResourceTags(configured map[string]*string) []map[string]interface{} {
	dataResources := make([]map[string]interface{}, 0, len(configured))

	for k, v := range configured {
		item := make(map[string]interface{})
		item["key"] = k
		item["value"] = aws.StringValue(v)

		dataResources = append(dataResources, item)
	}

	return dataResources
}

func validateExperimentTemplateStopConditionSource() schema.SchemaValidateFunc {
	allowedStopConditionSources := []string{
		"aws:cloudwatch:alarm",
		"none",
	}

	return validation.All(
		validation.StringInSlice(allowedStopConditionSources, false),
	)
}

func validateExperimentTemplateActionTargetKey() schema.SchemaValidateFunc {
	allowedStopConditionSources := []string{
		"Clusters",
		"DBInstances",
		"Instances",
		"Nodegroups",
		"Roles",
	}

	return validation.All(
		validation.StringLenBetween(0, 64),
		validation.StringInSlice(allowedStopConditionSources, false),
	)
}
