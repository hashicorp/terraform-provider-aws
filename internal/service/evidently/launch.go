// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package evidently

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/evidently"
	awstypes "github.com/aws/aws-sdk-go-v2/service/evidently/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/structure"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_evidently_launch", name="Launch")
// @Tags(identifierAttribute="arn")
func ResourceLaunch() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceLaunchCreate,
		ReadWithoutTimeout:   resourceLaunchRead,
		UpdateWithoutTimeout: resourceLaunchUpdate,
		DeleteWithoutTimeout: resourceLaunchDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(2 * time.Minute),
			Update: schema.DefaultTimeout(2 * time.Minute),
			Delete: schema.DefaultTimeout(2 * time.Minute),
		},

		Schema: map[string]*schema.Schema{
			names.AttrARN: {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrCreatedTime: {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrDescription: {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringLenBetween(0, 160),
			},
			"execution": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"ended_time": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"started_time": {
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
			},
			"groups": {
				Type:     schema.TypeList,
				Required: true,
				MinItems: 1,
				MaxItems: 5,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						names.AttrDescription: {
							Type:         schema.TypeString,
							Optional:     true,
							ValidateFunc: validation.StringLenBetween(0, 160),
						},
						"feature": {
							Type:     schema.TypeString,
							Required: true,
							ValidateFunc: validation.All(
								validation.StringLenBetween(1, 127),
								validation.StringMatch(regexache.MustCompile(`^[0-9A-Za-z_.-]*$`), "alphanumeric and can contain hyphens, underscores, and periods"),
							),
						},
						names.AttrName: {
							Type:     schema.TypeString,
							Required: true,
							ValidateFunc: validation.All(
								validation.StringLenBetween(1, 127),
								validation.StringMatch(regexache.MustCompile(`^[0-9A-Za-z_.-]*$`), "alphanumeric and can contain hyphens, underscores, and periods"),
							),
						},
						"variation": {
							Type:     schema.TypeString,
							Required: true,
							ValidateFunc: validation.All(
								validation.StringLenBetween(1, 127),
								validation.StringMatch(regexache.MustCompile(`^[0-9A-Za-z_.-]*$`), "alphanumeric and can contain hyphens, underscores, and periods"),
							),
						},
					},
				},
			},
			names.AttrLastUpdatedTime: {
				Type:     schema.TypeString,
				Computed: true,
			},
			"metric_monitors": {
				Type:     schema.TypeList,
				Optional: true,
				MinItems: 0,
				MaxItems: 3,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"metric_definition": {
							Type:     schema.TypeList,
							Required: true,
							MaxItems: 1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"entity_id_key": {
										Type:         schema.TypeString,
										Required:     true,
										ValidateFunc: validation.StringLenBetween(1, 256),
									},
									"event_pattern": {
										Type:     schema.TypeString,
										Optional: true,
										ValidateFunc: validation.All(
											validation.StringLenBetween(0, 1024),
											validation.StringIsJSON,
										),
										DiffSuppressFunc: verify.SuppressEquivalentJSONDiffs,
										StateFunc: func(v interface{}) string {
											json, _ := structure.NormalizeJsonString(v)
											return json
										},
									},
									names.AttrName: {
										Type:         schema.TypeString,
										Required:     true,
										ValidateFunc: validation.StringLenBetween(1, 255),
									},
									"unit_label": {
										Type:         schema.TypeString,
										Optional:     true,
										ValidateFunc: validation.StringLenBetween(1, 256),
									},
									"value_key": {
										Type:         schema.TypeString,
										Required:     true,
										ValidateFunc: validation.StringLenBetween(1, 256),
									},
								},
							},
						},
					},
				},
			},
			names.AttrName: {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
				ValidateFunc: validation.All(
					validation.StringLenBetween(1, 127),
					validation.StringMatch(regexache.MustCompile(`^[0-9A-Za-z_.-]*$`), "alphanumeric and can contain hyphens, underscores, and periods"),
				),
			},
			"project": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
				ValidateFunc: validation.All(
					validation.StringLenBetween(0, 2048),
					validation.StringMatch(regexache.MustCompile(`(^[0-9A-Za-z_.-]*$)|(arn:[^:]*:[^:]*:[^:]*:[^:]*:project/[0-9A-Za-z_.-]*)`), "name or arn of the project"),
				),
				DiffSuppressFunc: func(k, old, new string, d *schema.ResourceData) bool {
					// case 1: User-defined string (old) is a name and is the suffix of API-returned string (new). Check non-empty old in resoure creation scenario
					// case 2: after setting API-returned string.  User-defined string (new) is suffix of API-returned string (old)
					return (strings.HasSuffix(new, old) && old != "") || strings.HasSuffix(old, new)
				},
			},
			"randomization_salt": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringLenBetween(0, 127),
				// Default: set to the launch name if not specified
				DiffSuppressFunc: func(k, old, new string, d *schema.ResourceData) bool {
					return old == d.Get(names.AttrName).(string) && new == ""
				},
			},
			"scheduled_splits_config": {
				Type:     schema.TypeList,
				Optional: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"steps": {
							Type:     schema.TypeList,
							Required: true,
							MinItems: 1,
							MaxItems: 6,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"group_weights": {
										Type:     schema.TypeMap,
										Required: true,
										ValidateDiagFunc: validation.AllDiag(
											validation.MapKeyLenBetween(1, 127),
											validation.MapKeyMatch(regexache.MustCompile(`^[0-9A-Za-z_.-]*$`), "alphanumeric and can contain hyphens, underscores, and periods"),
										),
										Elem: &schema.Schema{
											Type:         schema.TypeInt,
											ValidateFunc: validation.IntBetween(0, 100000),
										},
									},
									"segment_overrides": {
										Type:     schema.TypeList,
										Optional: true,
										MinItems: 0,
										MaxItems: 6,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"evaluation_order": {
													Type:     schema.TypeInt,
													Required: true,
												},
												"segment": {
													Type:     schema.TypeString,
													Required: true,
													ValidateFunc: validation.All(
														validation.StringLenBetween(0, 2048),
														validation.StringMatch(regexache.MustCompile(`(^[0-9A-Za-z_.-]*$)|(arn:[^:]*:[^:]*:[^:]*:[^:]*:segment/[0-9A-Za-z._-]*)`), "name or arn of the segment"),
													),
													DiffSuppressFunc: func(k, old, new string, d *schema.ResourceData) bool {
														// case 1: User-defined string (old) is a name and is the suffix of API-returned string (new). Check non-empty old in resoure creation scenario
														// case 2: after setting API-returned string.  User-defined string (new) is suffix of API-returned string (old)
														return (strings.HasSuffix(new, old) && old != "") || strings.HasSuffix(old, new)
													},
												},
												"weights": {
													Type:     schema.TypeMap,
													Required: true,
													ValidateDiagFunc: validation.AllDiag(
														validation.MapKeyLenBetween(1, 127),
														validation.MapKeyMatch(regexache.MustCompile(`^[0-9A-Za-z_.-]*$`), "alphanumeric and can contain hyphens, underscores, and periods"),
													),
													Elem: &schema.Schema{
														Type:         schema.TypeInt,
														ValidateFunc: validation.IntBetween(0, 100000),
													},
												},
											},
										},
									},
									names.AttrStartTime: {
										Type:         schema.TypeString,
										Required:     true,
										ValidateFunc: verify.ValidUTCTimestamp,
									},
								},
							},
						},
					},
				},
			},
			names.AttrStatus: {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrStatusReason: {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrTags:    tftags.TagsSchema(),
			names.AttrTagsAll: tftags.TagsSchemaComputed(),
			names.AttrType: {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
		CustomizeDiff: verify.SetTagsDiff,
	}
}

func resourceLaunchCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	conn := meta.(*conns.AWSClient).EvidentlyClient(ctx)

	name := d.Get(names.AttrName).(string)
	project := d.Get("project").(string)
	input := &evidently.CreateLaunchInput{
		Name:    aws.String(name),
		Project: aws.String(project),
		Groups:  expandGroups(d.Get("groups").([]interface{})),
		Tags:    getTagsIn(ctx),
	}

	if v, ok := d.GetOk(names.AttrDescription); ok {
		input.Description = aws.String(v.(string))
	}

	if v, ok := d.GetOk("metric_monitors"); ok {
		input.MetricMonitors = expandMetricMonitors(v.([]interface{}))
	}

	if v, ok := d.GetOk("randomization_salt"); ok {
		input.RandomizationSalt = aws.String(v.(string))
	}

	if v, ok := d.GetOk("scheduled_splits_config"); ok {
		input.ScheduledSplitsConfig = expandScheduledSplitsConfig(v.([]interface{}))
	}

	output, err := conn.CreateLaunch(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating CloudWatch Evidently Launch (%s) for Project (%s): %s", name, project, err)
	}

	// the GetLaunch API call uses the Launch name and Project ARN
	// concat Launch name and Project Name or ARN to be used in Read for imports
	d.SetId(fmt.Sprintf("%s:%s", aws.ToString(output.Launch.Name), aws.ToString(output.Launch.Project)))

	if _, err := waitLaunchCreated(ctx, conn, d.Id(), d.Timeout(schema.TimeoutCreate)); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for CloudWatch Evidently Launch (%s) for Project (%s) creation: %s", name, project, err)
	}

	return append(diags, resourceLaunchRead(ctx, d, meta)...)
}

func resourceLaunchRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	conn := meta.(*conns.AWSClient).EvidentlyClient(ctx)

	launchName, projectNameOrARN, err := LaunchParseID(d.Id())

	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	launch, err := FindLaunchWithProjectNameorARN(ctx, conn, launchName, projectNameOrARN)

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] CloudWatch Evidently Launch (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading CloudWatch Evidently Launch (%s) for Project (%s): %s", launchName, projectNameOrARN, err)
	}

	if err := d.Set("execution", flattenExecution(launch.Execution)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting execution: %s", err)
	}

	if err := d.Set("groups", flattenGroups(launch.Groups)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting groups: %s", err)
	}

	if err := d.Set("metric_monitors", flattenMetricMonitors(launch.MetricMonitors)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting metric_monitors: %s", err)
	}

	if err := d.Set("scheduled_splits_config", flattenScheduledSplitsDefinition(launch.ScheduledSplitsDefinition)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting scheduled_splits_config: %s", err)
	}

	d.Set(names.AttrARN, launch.Arn)
	d.Set(names.AttrCreatedTime, aws.ToTime(launch.CreatedTime).Format(time.RFC3339))
	d.Set(names.AttrDescription, launch.Description)
	d.Set(names.AttrLastUpdatedTime, aws.ToTime(launch.LastUpdatedTime).Format(time.RFC3339))
	d.Set(names.AttrName, launch.Name)
	d.Set("project", launch.Project)
	d.Set("randomization_salt", launch.RandomizationSalt)
	d.Set(names.AttrStatus, launch.Status)
	d.Set(names.AttrStatusReason, launch.StatusReason)
	d.Set(names.AttrType, launch.Type)

	setTagsOut(ctx, launch.Tags)

	return diags
}

func resourceLaunchUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	conn := meta.(*conns.AWSClient).EvidentlyClient(ctx)

	if d.HasChanges(names.AttrDescription, "groups", "metric_monitors", "randomization_salt", "scheduled_splits_config") {
		name := d.Get(names.AttrName).(string)
		project := d.Get("project").(string)

		input := &evidently.UpdateLaunchInput{
			Description:           aws.String(d.Get(names.AttrDescription).(string)),
			Groups:                expandGroups(d.Get("groups").([]interface{})),
			Launch:                aws.String(name),
			Project:               aws.String(project),
			MetricMonitors:        expandMetricMonitors(d.Get("metric_monitors").([]interface{})),
			RandomizationSalt:     aws.String(d.Get("randomization_salt").(string)),
			ScheduledSplitsConfig: expandScheduledSplitsConfig(d.Get("scheduled_splits_config").([]interface{})),
		}

		_, err := conn.UpdateLaunch(ctx, input)

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "updating CloudWatch Evidently Launch (%s) for Project (%s): %s", name, project, err)
		}

		if _, err := waitLaunchUpdated(ctx, conn, d.Id(), d.Timeout(schema.TimeoutUpdate)); err != nil {
			return sdkdiag.AppendErrorf(diags, "waiting for CloudWatch Evidently Launch (%s) for Project (%s) update: %s", name, project, err)
		}
	}

	return append(diags, resourceLaunchRead(ctx, d, meta)...)
}

func resourceLaunchDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	conn := meta.(*conns.AWSClient).EvidentlyClient(ctx)

	name := d.Get(names.AttrName).(string)
	project := d.Get("project").(string)

	log.Printf("[DEBUG] Deleting CloudWatch Evidently Launch: %s", d.Id())
	_, err := conn.DeleteLaunch(ctx, &evidently.DeleteLaunchInput{
		Launch:  aws.String(name),
		Project: aws.String(project),
	})

	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting CloudWatch Evidently Launch (%s) for Project (%s): %s", name, project, err)
	}

	if _, err := waitLaunchDeleted(ctx, conn, d.Id(), d.Timeout(schema.TimeoutDelete)); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for CloudWatch Evidently Launch (%s) for Project (%s) deletion: %s", name, project, err)
	}

	return diags
}

func LaunchParseID(id string) (string, string, error) {
	launchName, projectNameOrARN, _ := strings.Cut(id, ":")

	if launchName == "" || projectNameOrARN == "" {
		return "", "", fmt.Errorf("unexpected format of ID (%s), expected launchName:projectNameOrARN", id)
	}

	return launchName, projectNameOrARN, nil
}

func expandGroups(tfMaps []interface{}) []awstypes.LaunchGroupConfig {
	apiObjects := make([]awstypes.LaunchGroupConfig, 0, len(tfMaps))

	for _, tfMap := range tfMaps {
		apiObjects = append(apiObjects, expandGroup(tfMap.(map[string]interface{})))
	}

	return apiObjects
}

func expandGroup(tfMap map[string]interface{}) awstypes.LaunchGroupConfig {
	apiObject := awstypes.LaunchGroupConfig{
		Feature:   aws.String(tfMap["feature"].(string)),
		Name:      aws.String(tfMap[names.AttrName].(string)),
		Variation: aws.String(tfMap["variation"].(string)),
	}

	if v, ok := tfMap[names.AttrDescription]; ok {
		apiObject.Description = aws.String(v.(string))
	}

	return apiObject
}

func expandMetricMonitors(tfMaps []interface{}) []awstypes.MetricMonitorConfig {
	apiObjects := make([]awstypes.MetricMonitorConfig, 0, len(tfMaps))

	for _, tfMap := range tfMaps {
		apiObjects = append(apiObjects, expandMetricMonitor(tfMap.(map[string]interface{})))
	}

	return apiObjects
}

func expandMetricMonitor(tfMap map[string]interface{}) awstypes.MetricMonitorConfig {
	apiObject := awstypes.MetricMonitorConfig{
		MetricDefinition: expandMetricDefinition(tfMap["metric_definition"].([]interface{})),
	}

	return apiObject
}

func expandMetricDefinition(tfList []interface{}) *awstypes.MetricDefinitionConfig {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap := tfList[0].(map[string]interface{})

	apiObject := &awstypes.MetricDefinitionConfig{
		EntityIdKey: aws.String(tfMap["entity_id_key"].(string)),
		Name:        aws.String(tfMap[names.AttrName].(string)),
		ValueKey:    aws.String(tfMap["value_key"].(string)),
	}

	if v, ok := tfMap["event_pattern"]; ok && v != "" {
		apiObject.EventPattern = aws.String(v.(string))
	}

	if v, ok := tfMap["unit_label"]; ok && v != "" {
		apiObject.UnitLabel = aws.String(v.(string))
	}

	return apiObject
}

func expandScheduledSplitsConfig(tfList []interface{}) *awstypes.ScheduledSplitsLaunchConfig {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap := tfList[0].(map[string]interface{})

	apiObject := &awstypes.ScheduledSplitsLaunchConfig{
		Steps: expandSteps(tfMap["steps"].([]interface{})),
	}

	return apiObject
}

func expandSteps(tfMaps []interface{}) []awstypes.ScheduledSplitConfig {
	apiObjects := make([]awstypes.ScheduledSplitConfig, 0, len(tfMaps))

	for _, tfMap := range tfMaps {
		apiObjects = append(apiObjects, expandStep(tfMap.(map[string]interface{})))
	}

	return apiObjects
}

func expandStep(tfMap map[string]interface{}) awstypes.ScheduledSplitConfig {
	t, _ := time.Parse(time.RFC3339, tfMap[names.AttrStartTime].(string))
	startTime := aws.Time(t)

	apiObject := awstypes.ScheduledSplitConfig{
		GroupWeights:     flex.ExpandInt64ValueMap(tfMap["group_weights"].(map[string]interface{})),
		SegmentOverrides: expandSegmentOverrides(tfMap["segment_overrides"].([]interface{})),
		StartTime:        startTime,
	}

	return apiObject
}

func expandSegmentOverrides(tfMaps []interface{}) []awstypes.SegmentOverride {
	apiObjects := make([]awstypes.SegmentOverride, 0, len(tfMaps))

	for _, tfMap := range tfMaps {
		apiObjects = append(apiObjects, expandSegmentOverride(tfMap.(map[string]interface{})))
	}

	return apiObjects
}

func expandSegmentOverride(tfMap map[string]interface{}) awstypes.SegmentOverride {
	apiObject := awstypes.SegmentOverride{
		EvaluationOrder: aws.Int64(int64(tfMap["evaluation_order"].(int))),
		Segment:         aws.String(tfMap["segment"].(string)),
		Weights:         flex.ExpandInt64ValueMap(tfMap["weights"].(map[string]interface{})),
	}

	return apiObject
}

func flattenExecution(apiObjects *awstypes.LaunchExecution) []interface{} {
	if apiObjects == nil {
		return nil
	}

	values := map[string]interface{}{}

	if apiObjects.EndedTime != nil {
		values["ended_time"] = aws.ToTime(apiObjects.EndedTime).Format(time.RFC3339)
	}

	if apiObjects.StartedTime != nil {
		values["started_time"] = aws.ToTime(apiObjects.StartedTime).Format(time.RFC3339)
	}

	return []interface{}{values}
}

func flattenGroups(apiObjects []awstypes.LaunchGroup) []interface{} {
	if len(apiObjects) == 0 {
		return nil
	}

	var tfList []interface{}

	for _, apiObject := range apiObjects {
		if apiObject.Name == nil {
			continue
		}

		tfList = append(tfList, flattenGroup(apiObject))
	}

	return tfList
}

func flattenGroup(apiObject awstypes.LaunchGroup) map[string]interface{} {
	if apiObject.Name == nil {
		return nil
	}

	tfMap := map[string]interface{}{
		names.AttrName: aws.ToString(apiObject.Name),
	}

	for feature, variation := range apiObject.FeatureVariations {
		tfMap["feature"] = feature
		tfMap["variation"] = variation
	}

	if v := apiObject.Description; v != nil {
		tfMap[names.AttrDescription] = aws.ToString(v)
	}

	return tfMap
}

func flattenMetricMonitors(apiObjects []awstypes.MetricMonitor) []interface{} {
	if len(apiObjects) == 0 {
		return nil
	}

	var tfList []interface{}

	for _, apiObject := range apiObjects {
		if apiObject == (awstypes.MetricMonitor{}) {
			continue
		}

		tfList = append(tfList, flattenMetricMonitor(apiObject))
	}

	return tfList
}

func flattenMetricMonitor(apiObject awstypes.MetricMonitor) map[string]interface{} {
	if apiObject == (awstypes.MetricMonitor{}) {
		return nil
	}

	tfMap := map[string]interface{}{
		"metric_definition": flattenMetricMonitorDefinition(apiObject.MetricDefinition),
	}

	return tfMap
}

func flattenMetricMonitorDefinition(apiObject *awstypes.MetricDefinition) []interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{
		"entity_id_key": aws.ToString(apiObject.EntityIdKey),
		names.AttrName:  aws.ToString(apiObject.Name),
		"value_key":     aws.ToString(apiObject.ValueKey),
	}

	if v := apiObject.EventPattern; v != nil {
		tfMap["event_pattern"] = aws.ToString(v)
	}

	if v := apiObject.UnitLabel; v != nil {
		tfMap["unit_label"] = aws.ToString(v)
	}

	return []interface{}{tfMap}
}

func flattenScheduledSplitsDefinition(apiObject *awstypes.ScheduledSplitsLaunchDefinition) []interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{
		"steps": flattenSteps(apiObject.Steps),
	}

	return []interface{}{tfMap}
}

func flattenSteps(apiObjects []awstypes.ScheduledSplit) []interface{} {
	if len(apiObjects) == 0 {
		return nil
	}

	var tfList []interface{}

	for _, apiObject := range apiObjects {
		if apiObject.StartTime == nil {
			continue
		}

		tfList = append(tfList, flattenStep(apiObject))
	}

	return tfList
}

func flattenStep(apiObject awstypes.ScheduledSplit) map[string]interface{} {
	if apiObject.StartTime == nil {
		return nil
	}

	tfMap := map[string]interface{}{
		"group_weights":     apiObject.GroupWeights,
		names.AttrStartTime: aws.ToTime(apiObject.StartTime).Format(time.RFC3339),
	}

	if v := apiObject.SegmentOverrides; v != nil {
		tfMap["segment_overrides"] = flattenSegmentOverrides(v)
	}

	return tfMap
}

func flattenSegmentOverrides(apiObjects []awstypes.SegmentOverride) []interface{} {
	if len(apiObjects) == 0 {
		return nil
	}

	var tfList []interface{}

	for _, apiObject := range apiObjects {
		if apiObject.EvaluationOrder == nil {
			continue
		}

		tfList = append(tfList, flattenSegmentOverride(apiObject))
	}

	return tfList
}

func flattenSegmentOverride(apiObject awstypes.SegmentOverride) map[string]interface{} {
	if apiObject.EvaluationOrder == nil {
		return nil
	}

	tfMap := map[string]interface{}{
		"evaluation_order": aws.ToInt64(apiObject.EvaluationOrder),
		"segment":          aws.ToString(apiObject.Segment),
		"weights":          apiObject.Weights,
	}

	return tfMap
}
