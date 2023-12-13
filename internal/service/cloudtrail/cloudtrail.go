// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package cloudtrail

import (
	"context"
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/arn"
	"github.com/aws/aws-sdk-go/service/cloudtrail"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_cloudtrail", name="Trail")
// @Tags(identifierAttribute="arn")
func ResourceCloudTrail() *schema.Resource { // nosemgrep:ci.cloudtrail-in-func-name
	return &schema.Resource{
		CreateWithoutTimeout: resourceCloudTrailCreate,
		ReadWithoutTimeout:   resourceCloudTrailRead,
		UpdateWithoutTimeout: resourceCloudTrailUpdate,
		DeleteWithoutTimeout: resourceCloudTrailDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		SchemaVersion: 1,
		StateUpgraders: []schema.StateUpgrader{
			{
				Type:    resourceCloudTrailV0().CoreConfigSchema().ImpliedType(),
				Upgrade: cloudTrailUpgradeV0,
				Version: 0,
			},
		},

		Schema: map[string]*schema.Schema{
			"advanced_event_selector": {
				Type:          schema.TypeList,
				Optional:      true,
				ConflictsWith: []string{"event_selector"},
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"field_selector": {
							Type:     schema.TypeSet,
							Required: true,
							MinItems: 1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"ends_with": {
										Type:     schema.TypeList,
										Optional: true,
										MinItems: 1,
										Elem: &schema.Schema{
											Type:         schema.TypeString,
											ValidateFunc: validation.StringLenBetween(1, 2048),
										},
									},
									"equals": {
										Type:     schema.TypeList,
										Optional: true,
										MinItems: 1,
										Elem: &schema.Schema{
											Type:         schema.TypeString,
											ValidateFunc: validation.StringLenBetween(1, 2048),
										},
									},
									"field": {
										Type:         schema.TypeString,
										Required:     true,
										ValidateFunc: validation.StringInSlice(field_Values(), false),
									},
									"not_ends_with": {
										Type:     schema.TypeList,
										Optional: true,
										MinItems: 1,
										Elem: &schema.Schema{
											Type:         schema.TypeString,
											ValidateFunc: validation.StringLenBetween(1, 2048),
										},
									},
									"not_equals": {
										Type:     schema.TypeList,
										Optional: true,
										MinItems: 1,
										Elem: &schema.Schema{
											Type:         schema.TypeString,
											ValidateFunc: validation.StringLenBetween(1, 2048),
										},
									},
									"not_starts_with": {
										Type:     schema.TypeList,
										Optional: true,
										MinItems: 1,
										Elem: &schema.Schema{
											Type:         schema.TypeString,
											ValidateFunc: validation.StringLenBetween(1, 2048),
										},
									},
									"starts_with": {
										Type:     schema.TypeList,
										Optional: true,
										MinItems: 1,
										Elem: &schema.Schema{
											Type:         schema.TypeString,
											ValidateFunc: validation.StringLenBetween(1, 2048),
										},
									},
								},
							},
						},
						"name": {
							Type:         schema.TypeString,
							Optional:     true,
							ValidateFunc: validation.StringLenBetween(0, 1000),
						},
					},
				},
			},
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"cloud_watch_logs_group_arn": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: verify.ValidARN,
			},
			"cloud_watch_logs_role_arn": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: verify.ValidARN,
			},
			"enable_log_file_validation": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},
			"enable_logging": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  true,
			},
			"event_selector": {
				Type:          schema.TypeList,
				Optional:      true,
				MaxItems:      5,
				ConflictsWith: []string{"advanced_event_selector"},
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"data_resource": {
							Type:     schema.TypeList,
							Optional: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"type": {
										Type:         schema.TypeString,
										Required:     true,
										ValidateFunc: validation.StringInSlice(resourceType_Values(), false),
									},
									"values": {
										Type:     schema.TypeList,
										Required: true,
										MaxItems: 250,
										Elem:     &schema.Schema{Type: schema.TypeString},
									},
								},
							},
						},
						"exclude_management_event_sources": {
							Type:     schema.TypeSet,
							Optional: true,
							Elem:     &schema.Schema{Type: schema.TypeString},
						},
						"include_management_events": {
							Type:     schema.TypeBool,
							Optional: true,
							Default:  true,
						},
						"read_write_type": {
							Type:         schema.TypeString,
							Optional:     true,
							Default:      cloudtrail.ReadWriteTypeAll,
							ValidateFunc: validation.StringInSlice(cloudtrail.ReadWriteType_Values(), false),
						},
					},
				},
			},
			"home_region": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"include_global_service_events": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  true,
			},
			"insight_selector": {
				Type:     schema.TypeList,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"insight_type": {
							Type:         schema.TypeString,
							Required:     true,
							ValidateFunc: validation.StringInSlice(cloudtrail.InsightType_Values(), false),
						},
					},
				},
			},
			"is_multi_region_trail": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},
			"is_organization_trail": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},
			"kms_key_id": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: verify.ValidARN,
			},
			"name": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringLenBetween(3, 128),
			},
			"s3_bucket_name": {
				Type:     schema.TypeString,
				Required: true,
			},
			"s3_key_prefix": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringLenBetween(0, 2000),
			},
			"sns_topic_name": {
				Type:     schema.TypeString,
				Optional: true,
			},
			names.AttrTags:    tftags.TagsSchema(),
			names.AttrTagsAll: tftags.TagsSchemaComputed(),
		},

		CustomizeDiff: verify.SetTagsDiff,
	}
}

func resourceCloudTrailCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics { // nosemgrep:ci.cloudtrail-in-func-name
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).CloudTrailConn(ctx)

	name := d.Get("name").(string)
	input := &cloudtrail.CreateTrailInput{
		IncludeGlobalServiceEvents: aws.Bool(d.Get("include_global_service_events").(bool)),
		Name:                       aws.String(name),
		S3BucketName:               aws.String(d.Get("s3_bucket_name").(string)),
		TagsList:                   getTagsIn(ctx),
	}

	if v, ok := d.GetOk("cloud_watch_logs_group_arn"); ok {
		input.CloudWatchLogsLogGroupArn = aws.String(v.(string))
	}

	if v, ok := d.GetOk("cloud_watch_logs_role_arn"); ok {
		input.CloudWatchLogsRoleArn = aws.String(v.(string))
	}

	if v, ok := d.GetOk("enable_log_file_validation"); ok {
		input.EnableLogFileValidation = aws.Bool(v.(bool))
	}

	if v, ok := d.GetOk("is_multi_region_trail"); ok {
		input.IsMultiRegionTrail = aws.Bool(v.(bool))
	}

	if v, ok := d.GetOk("is_organization_trail"); ok {
		input.IsOrganizationTrail = aws.Bool(v.(bool))
	}

	if v, ok := d.GetOk("kms_key_id"); ok {
		input.KmsKeyId = aws.String(v.(string))
	}

	if v, ok := d.GetOk("s3_key_prefix"); ok {
		input.S3KeyPrefix = aws.String(v.(string))
	}

	if v, ok := d.GetOk("sns_topic_name"); ok {
		input.SnsTopicName = aws.String(v.(string))
	}

	outputRaw, err := tfresource.RetryWhen(ctx, propagationTimeout,
		func() (interface{}, error) {
			return conn.CreateTrailWithContext(ctx, input)
		},
		func(err error) (bool, error) {
			if tfawserr.ErrMessageContains(err, cloudtrail.ErrCodeInvalidCloudWatchLogsRoleArnException, "Access denied.") {
				return true, err
			}

			if tfawserr.ErrMessageContains(err, cloudtrail.ErrCodeInvalidCloudWatchLogsLogGroupArnException, "Access denied.") {
				return true, err
			}

			return false, err
		},
	)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating CloudTrail Trail (%s): %s", name, err)
	}

	d.SetId(aws.StringValue(outputRaw.(*cloudtrail.CreateTrailOutput).TrailARN))

	// AWS CloudTrail sets newly-created trails to false.
	if d.Get("enable_logging").(bool) {
		if err := setLogging(ctx, conn, d.Id(), true); err != nil {
			return sdkdiag.AppendFromErr(diags, err)
		}
	}

	if _, ok := d.GetOk("event_selector"); ok {
		if err := setEventSelectors(ctx, conn, d); err != nil {
			return sdkdiag.AppendFromErr(diags, err)
		}
	}

	if _, ok := d.GetOk("advanced_event_selector"); ok {
		if err := setAdvancedEventSelectors(ctx, conn, d); err != nil {
			return sdkdiag.AppendFromErr(diags, err)
		}
	}

	if _, ok := d.GetOk("insight_selector"); ok {
		if err := setInsightSelectors(ctx, conn, d); err != nil {
			return sdkdiag.AppendFromErr(diags, err)
		}
	}

	return append(diags, resourceCloudTrailRead(ctx, d, meta)...)
}

func resourceCloudTrailRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics { // nosemgrep:ci.cloudtrail-in-func-name
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).CloudTrailConn(ctx)

	outputRaw, err := tfresource.RetryWhenNewResourceNotFound(ctx, propagationTimeout, func() (interface{}, error) {
		return findTrailByARN(ctx, conn, d.Id())
	}, d.IsNewResource())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] CloudTrail Trail (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading CloudTrail Trail (%s): %s", d.Id(), err)
	}

	trail := outputRaw.(*cloudtrail.Trail)
	arn := aws.StringValue(trail.TrailARN)
	d.Set("arn", arn)
	d.Set("cloud_watch_logs_group_arn", trail.CloudWatchLogsLogGroupArn)
	d.Set("cloud_watch_logs_role_arn", trail.CloudWatchLogsRoleArn)
	d.Set("enable_log_file_validation", trail.LogFileValidationEnabled)
	d.Set("home_region", trail.HomeRegion)
	d.Set("include_global_service_events", trail.IncludeGlobalServiceEvents)
	d.Set("is_multi_region_trail", trail.IsMultiRegionTrail)
	d.Set("is_organization_trail", trail.IsOrganizationTrail)
	d.Set("kms_key_id", trail.KmsKeyId)
	d.Set("name", trail.Name)
	d.Set("s3_bucket_name", trail.S3BucketName)
	d.Set("s3_key_prefix", trail.S3KeyPrefix)
	d.Set("sns_topic_name", trail.SnsTopicName)

	if output, err := conn.GetTrailStatusWithContext(ctx, &cloudtrail.GetTrailStatusInput{
		Name: aws.String(d.Id()),
	}); err != nil {
		return sdkdiag.AppendErrorf(diags, "reading CloudTrail Trail (%s) status: %s", d.Id(), err)
	} else {
		d.Set("enable_logging", output.IsLogging)
	}

	if aws.BoolValue(trail.HasCustomEventSelectors) {
		input := &cloudtrail.GetEventSelectorsInput{
			TrailName: aws.String(d.Id()),
		}

		output, err := conn.GetEventSelectorsWithContext(ctx, input)

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "reading CloudTrail Trail (%s) event selectors: %s", d.Id(), err)
		}

		if err := d.Set("event_selector", flattenEventSelector(output.EventSelectors)); err != nil {
			return sdkdiag.AppendErrorf(diags, "setting event_selector")
		}

		if err := d.Set("advanced_event_selector", flattenAdvancedEventSelector(output.AdvancedEventSelectors)); err != nil {
			return sdkdiag.AppendErrorf(diags, "setting advanced_event_selector")
		}
	}

	if aws.BoolValue(trail.HasInsightSelectors) {
		input := &cloudtrail.GetInsightSelectorsInput{
			TrailName: aws.String(d.Id()),
		}

		output, err := conn.GetInsightSelectorsWithContext(ctx, input)

		if err != nil {
			if !tfawserr.ErrCodeEquals(err, cloudtrail.ErrCodeInsightNotEnabledException) {
				return sdkdiag.AppendErrorf(diags, "reading CloudTrail Trail (%s) insight selectors: %s", d.Id(), err)
			}
		} else if err := d.Set("insight_selector", flattenInsightSelector(output.InsightSelectors)); err != nil {
			return sdkdiag.AppendErrorf(diags, "setting insight_selector")
		}
	}

	return diags
}

func resourceCloudTrailUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics { // nosemgrep:ci.cloudtrail-in-func-name
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).CloudTrailConn(ctx)

	if d.HasChangesExcept("tags", "tags_all", "insight_selector", "advanced_event_selector", "event_selector", "enable_logging") {
		input := &cloudtrail.UpdateTrailInput{
			Name: aws.String(d.Id()),
		}

		if d.HasChanges("cloud_watch_logs_role_arn", "cloud_watch_logs_group_arn") {
			// Both of these need to be provided together in the update call otherwise API complains.
			input.CloudWatchLogsRoleArn = aws.String(d.Get("cloud_watch_logs_role_arn").(string))
			input.CloudWatchLogsLogGroupArn = aws.String(d.Get("cloud_watch_logs_group_arn").(string))
		}

		if d.HasChange("enable_log_file_validation") {
			input.EnableLogFileValidation = aws.Bool(d.Get("enable_log_file_validation").(bool))
		}

		if d.HasChange("include_global_service_events") {
			input.IncludeGlobalServiceEvents = aws.Bool(d.Get("include_global_service_events").(bool))
		}

		if d.HasChange("is_multi_region_trail") {
			input.IsMultiRegionTrail = aws.Bool(d.Get("is_multi_region_trail").(bool))
		}

		if d.HasChange("is_organization_trail") {
			input.IsOrganizationTrail = aws.Bool(d.Get("is_organization_trail").(bool))
		}

		if d.HasChange("kms_key_id") {
			input.KmsKeyId = aws.String(d.Get("kms_key_id").(string))
		}

		if d.HasChange("s3_bucket_name") {
			input.S3BucketName = aws.String(d.Get("s3_bucket_name").(string))
		}

		if d.HasChange("s3_key_prefix") {
			input.S3KeyPrefix = aws.String(d.Get("s3_key_prefix").(string))
		}

		if d.HasChange("sns_topic_name") {
			input.SnsTopicName = aws.String(d.Get("sns_topic_name").(string))
		}

		_, err := tfresource.RetryWhen(ctx, propagationTimeout,
			func() (interface{}, error) {
				return conn.UpdateTrailWithContext(ctx, input)
			},
			func(err error) (bool, error) {
				if tfawserr.ErrMessageContains(err, cloudtrail.ErrCodeInvalidCloudWatchLogsRoleArnException, "Access denied.") {
					return true, err
				}

				if tfawserr.ErrMessageContains(err, cloudtrail.ErrCodeInvalidCloudWatchLogsLogGroupArnException, "Access denied.") {
					return true, err
				}

				return false, err
			},
		)

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "updating CloudTrail Trail (%s): %s", d.Id(), err)
		}
	}

	if d.HasChange("enable_logging") {
		if err := setLogging(ctx, conn, d.Id(), d.Get("enable_logging").(bool)); err != nil {
			return sdkdiag.AppendFromErr(diags, err)
		}
	}

	if d.HasChange("event_selector") {
		if err := setEventSelectors(ctx, conn, d); err != nil {
			return sdkdiag.AppendFromErr(diags, err)
		}
	}

	if d.HasChange("advanced_event_selector") {
		if err := setAdvancedEventSelectors(ctx, conn, d); err != nil {
			return sdkdiag.AppendFromErr(diags, err)
		}
	}

	if d.HasChange("insight_selector") {
		if err := setInsightSelectors(ctx, conn, d); err != nil {
			return sdkdiag.AppendFromErr(diags, err)
		}
	}

	return append(diags, resourceCloudTrailRead(ctx, d, meta)...)
}

func resourceCloudTrailDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics { // nosemgrep:ci.cloudtrail-in-func-name
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).CloudTrailConn(ctx)

	log.Printf("[DEBUG] Deleting CloudTrail Trail: %s", d.Id())
	_, err := conn.DeleteTrailWithContext(ctx, &cloudtrail.DeleteTrailInput{
		Name: aws.String(d.Id()),
	})

	if tfawserr.ErrCodeEquals(err, cloudtrail.ErrCodeTrailNotFoundException) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting CloudTrail Trail (%s): %s", d.Id(), err)
	}

	return diags
}

func findTrailByARN(ctx context.Context, conn *cloudtrail.CloudTrail, arn string) (*cloudtrail.Trail, error) {
	input := &cloudtrail.DescribeTrailsInput{
		TrailNameList: aws.StringSlice([]string{arn}),
	}

	return findTrail(ctx, conn, input)
}

func findTrail(ctx context.Context, conn *cloudtrail.CloudTrail, input *cloudtrail.DescribeTrailsInput) (*cloudtrail.Trail, error) {
	output, err := findTrails(ctx, conn, input)

	if err != nil {
		return nil, err
	}

	return tfresource.AssertSinglePtrResult(output)
}

func findTrails(ctx context.Context, conn *cloudtrail.CloudTrail, input *cloudtrail.DescribeTrailsInput) ([]*cloudtrail.Trail, error) {
	output, err := conn.DescribeTrailsWithContext(ctx, input)

	if err != nil {
		return nil, err
	}

	if output == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	return output.TrailList, nil
}

func setLogging(ctx context.Context, conn *cloudtrail.CloudTrail, name string, enabled bool) error {
	if enabled {
		input := &cloudtrail.StartLoggingInput{
			Name: aws.String(name),
		}

		if _, err := conn.StartLoggingWithContext(ctx, input); err != nil {
			return fmt.Errorf("starting CloudTrail Trail (%s) logging: %w", name, err)
		}
	} else {
		input := &cloudtrail.StopLoggingInput{
			Name: aws.String(name),
		}

		if _, err := conn.StopLoggingWithContext(ctx, input); err != nil {
			return fmt.Errorf("stopping CloudTrail Trail (%s) logging: %w", name, err)
		}
	}

	return nil
}

func setEventSelectors(ctx context.Context, conn *cloudtrail.CloudTrail, d *schema.ResourceData) error {
	input := &cloudtrail.PutEventSelectorsInput{
		TrailName: aws.String(d.Id()),
	}

	eventSelectors := expandEventSelector(d.Get("event_selector").([]interface{}))
	// If no defined selectors revert to the single default selector.
	if len(eventSelectors) == 0 {
		eventSelector := &cloudtrail.EventSelector{
			IncludeManagementEvents: aws.Bool(true),
			ReadWriteType:           aws.String(cloudtrail.ReadWriteTypeAll),
			DataResources:           make([]*cloudtrail.DataResource, 0),
		}
		eventSelectors = append(eventSelectors, eventSelector)
	}
	input.EventSelectors = eventSelectors

	if _, err := conn.PutEventSelectorsWithContext(ctx, input); err != nil {
		return fmt.Errorf("setting CloudTrail Trail (%s) event selectors: %s", d.Id(), err)
	}

	return nil
}

func expandEventSelector(configured []interface{}) []*cloudtrail.EventSelector {
	eventSelectors := make([]*cloudtrail.EventSelector, 0, len(configured))

	for _, raw := range configured {
		data := raw.(map[string]interface{})
		dataResources := expandEventSelectorDataResource(data["data_resource"].([]interface{}))

		es := &cloudtrail.EventSelector{
			IncludeManagementEvents: aws.Bool(data["include_management_events"].(bool)),
			ReadWriteType:           aws.String(data["read_write_type"].(string)),
			DataResources:           dataResources,
		}

		if v, ok := data["exclude_management_event_sources"].(*schema.Set); ok && v.Len() > 0 {
			es.ExcludeManagementEventSources = flex.ExpandStringSet(v)
		}

		eventSelectors = append(eventSelectors, es)
	}

	return eventSelectors
}

func expandEventSelectorDataResource(configured []interface{}) []*cloudtrail.DataResource {
	dataResources := make([]*cloudtrail.DataResource, 0, len(configured))

	for _, raw := range configured {
		data := raw.(map[string]interface{})

		dataResource := &cloudtrail.DataResource{
			Type:   aws.String(data["type"].(string)),
			Values: flex.ExpandStringList(data["values"].([]interface{})),
		}

		dataResources = append(dataResources, dataResource)
	}

	return dataResources
}

func flattenEventSelector(configured []*cloudtrail.EventSelector) []map[string]interface{} {
	eventSelectors := make([]map[string]interface{}, 0, len(configured))

	// Prevent default configurations shows differences
	if len(configured) == 1 && len(configured[0].DataResources) == 0 && aws.StringValue(configured[0].ReadWriteType) == cloudtrail.ReadWriteTypeAll && len(configured[0].ExcludeManagementEventSources) == 0 {
		return eventSelectors
	}

	for _, raw := range configured {
		item := make(map[string]interface{})
		item["read_write_type"] = aws.StringValue(raw.ReadWriteType)
		item["exclude_management_event_sources"] = flex.FlattenStringSet(raw.ExcludeManagementEventSources)
		item["include_management_events"] = aws.BoolValue(raw.IncludeManagementEvents)
		item["data_resource"] = flattenEventSelectorDataResource(raw.DataResources)

		eventSelectors = append(eventSelectors, item)
	}

	return eventSelectors
}

func flattenEventSelectorDataResource(configured []*cloudtrail.DataResource) []map[string]interface{} {
	dataResources := make([]map[string]interface{}, 0, len(configured))

	for _, raw := range configured {
		item := make(map[string]interface{})
		item["type"] = aws.StringValue(raw.Type)
		item["values"] = flex.FlattenStringList(raw.Values)

		dataResources = append(dataResources, item)
	}

	return dataResources
}

func setAdvancedEventSelectors(ctx context.Context, conn *cloudtrail.CloudTrail, d *schema.ResourceData) error {
	input := &cloudtrail.PutEventSelectorsInput{
		AdvancedEventSelectors: expandAdvancedEventSelector(d.Get("advanced_event_selector").([]interface{})),
		TrailName:              aws.String(d.Id()),
	}

	if _, err := conn.PutEventSelectorsWithContext(ctx, input); err != nil {
		return fmt.Errorf("setting CloudTrail Trail (%s) advanced selectors: %w", d.Id(), err)
	}

	return nil
}

func expandAdvancedEventSelector(configured []interface{}) []*cloudtrail.AdvancedEventSelector {
	advancedEventSelectors := make([]*cloudtrail.AdvancedEventSelector, 0, len(configured))

	for _, raw := range configured {
		data := raw.(map[string]interface{})
		fieldSelectors := expandAdvancedEventSelectorFieldSelector(data["field_selector"].(*schema.Set))

		aes := &cloudtrail.AdvancedEventSelector{
			Name:           aws.String(data["name"].(string)),
			FieldSelectors: fieldSelectors,
		}

		advancedEventSelectors = append(advancedEventSelectors, aes)
	}

	return advancedEventSelectors
}

func expandAdvancedEventSelectorFieldSelector(configured *schema.Set) []*cloudtrail.AdvancedFieldSelector {
	fieldSelectors := make([]*cloudtrail.AdvancedFieldSelector, 0, configured.Len())

	for _, raw := range configured.List() {
		data := raw.(map[string]interface{})
		fieldSelector := &cloudtrail.AdvancedFieldSelector{
			Field: aws.String(data["field"].(string)),
		}

		if v, ok := data["equals"].([]interface{}); ok && len(v) > 0 {
			fieldSelector.Equals = flex.ExpandStringList(v)
		}

		if v, ok := data["not_equals"].([]interface{}); ok && len(v) > 0 {
			fieldSelector.NotEquals = flex.ExpandStringList(v)
		}

		if v, ok := data["starts_with"].([]interface{}); ok && len(v) > 0 {
			fieldSelector.StartsWith = flex.ExpandStringList(v)
		}

		if v, ok := data["not_starts_with"].([]interface{}); ok && len(v) > 0 {
			fieldSelector.NotStartsWith = flex.ExpandStringList(v)
		}

		if v, ok := data["ends_with"].([]interface{}); ok && len(v) > 0 {
			fieldSelector.EndsWith = flex.ExpandStringList(v)
		}

		if v, ok := data["not_ends_with"].([]interface{}); ok && len(v) > 0 {
			fieldSelector.NotEndsWith = flex.ExpandStringList(v)
		}

		fieldSelectors = append(fieldSelectors, fieldSelector)
	}

	return fieldSelectors
}

func flattenAdvancedEventSelector(configured []*cloudtrail.AdvancedEventSelector) []map[string]interface{} {
	advancedEventSelectors := make([]map[string]interface{}, 0, len(configured))

	for _, raw := range configured {
		item := make(map[string]interface{})
		item["name"] = aws.StringValue(raw.Name)
		item["field_selector"] = flattenAdvancedEventSelectorFieldSelector(raw.FieldSelectors)

		advancedEventSelectors = append(advancedEventSelectors, item)
	}

	return advancedEventSelectors
}

func flattenAdvancedEventSelectorFieldSelector(configured []*cloudtrail.AdvancedFieldSelector) []map[string]interface{} {
	fieldSelectors := make([]map[string]interface{}, 0, len(configured))

	for _, raw := range configured {
		item := make(map[string]interface{})
		item["field"] = aws.StringValue(raw.Field)
		if raw.Equals != nil {
			item["equals"] = flex.FlattenStringList(raw.Equals)
		}
		if raw.NotEquals != nil {
			item["not_equals"] = flex.FlattenStringList(raw.NotEquals)
		}
		if raw.StartsWith != nil {
			item["starts_with"] = flex.FlattenStringList(raw.StartsWith)
		}
		if raw.NotStartsWith != nil {
			item["not_starts_with"] = flex.FlattenStringList(raw.NotStartsWith)
		}
		if raw.EndsWith != nil {
			item["ends_with"] = flex.FlattenStringList(raw.EndsWith)
		}
		if raw.NotEndsWith != nil {
			item["not_ends_with"] = flex.FlattenStringList(raw.NotEndsWith)
		}

		fieldSelectors = append(fieldSelectors, item)
	}

	return fieldSelectors
}

func setInsightSelectors(ctx context.Context, conn *cloudtrail.CloudTrail, d *schema.ResourceData) error {
	input := &cloudtrail.PutInsightSelectorsInput{
		InsightSelectors: expandInsightSelector(d.Get("insight_selector").([]interface{})),
		TrailName:        aws.String(d.Id()),
	}

	if _, err := conn.PutInsightSelectorsWithContext(ctx, input); err != nil {
		return fmt.Errorf("setting CloudTrail Trail (%s) insight selectors: %w", d.Id(), err)
	}

	return nil
}

func expandInsightSelector(configured []interface{}) []*cloudtrail.InsightSelector {
	insightSelectors := make([]*cloudtrail.InsightSelector, 0, len(configured))

	for _, raw := range configured {
		data := raw.(map[string]interface{})

		is := &cloudtrail.InsightSelector{
			InsightType: aws.String(data["insight_type"].(string)),
		}
		insightSelectors = append(insightSelectors, is)
	}

	return insightSelectors
}

func flattenInsightSelector(configured []*cloudtrail.InsightSelector) []map[string]interface{} {
	insightSelectors := make([]map[string]interface{}, 0, len(configured))

	for _, raw := range configured {
		item := make(map[string]interface{})
		item["insight_type"] = aws.StringValue(raw.InsightType)

		insightSelectors = append(insightSelectors, item)
	}

	return insightSelectors
}

// aws_cloudtrail's Schema @v5.24.0 minus validators.
func resourceCloudTrailV0() *schema.Resource { // nosemgrep:ci.cloudtrail-in-func-name
	return &schema.Resource{
		Schema: map[string]*schema.Schema{
			"advanced_event_selector": {
				Type:          schema.TypeList,
				Optional:      true,
				ConflictsWith: []string{"event_selector"},
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"field_selector": {
							Type:     schema.TypeSet,
							Required: true,
							MinItems: 1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"ends_with": {
										Type:     schema.TypeList,
										Optional: true,
										MinItems: 1,
										Elem: &schema.Schema{
											Type: schema.TypeString,
										},
									},
									"equals": {
										Type:     schema.TypeList,
										Optional: true,
										MinItems: 1,
										Elem: &schema.Schema{
											Type: schema.TypeString,
										},
									},
									"field": {
										Type:     schema.TypeString,
										Required: true,
									},
									"not_ends_with": {
										Type:     schema.TypeList,
										Optional: true,
										MinItems: 1,
										Elem: &schema.Schema{
											Type: schema.TypeString,
										},
									},
									"not_equals": {
										Type:     schema.TypeList,
										Optional: true,
										MinItems: 1,
										Elem: &schema.Schema{
											Type: schema.TypeString,
										},
									},
									"not_starts_with": {
										Type:     schema.TypeList,
										Optional: true,
										MinItems: 1,
										Elem: &schema.Schema{
											Type: schema.TypeString,
										},
									},
									"starts_with": {
										Type:     schema.TypeList,
										Optional: true,
										MinItems: 1,
										Elem: &schema.Schema{
											Type: schema.TypeString,
										},
									},
								},
							},
						},
						"name": {
							Type:     schema.TypeString,
							Optional: true,
						},
					},
				},
			},
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"cloud_watch_logs_group_arn": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"cloud_watch_logs_role_arn": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"enable_log_file_validation": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},
			"enable_logging": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  true,
			},
			"event_selector": {
				Type:          schema.TypeList,
				Optional:      true,
				MaxItems:      5,
				ConflictsWith: []string{"advanced_event_selector"},
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"data_resource": {
							Type:     schema.TypeList,
							Optional: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"type": {
										Type:     schema.TypeString,
										Required: true,
									},
									"values": {
										Type:     schema.TypeList,
										Required: true,
										MaxItems: 250,
										Elem:     &schema.Schema{Type: schema.TypeString},
									},
								},
							},
						},
						"exclude_management_event_sources": {
							Type:     schema.TypeSet,
							Optional: true,
							Elem:     &schema.Schema{Type: schema.TypeString},
						},
						"include_management_events": {
							Type:     schema.TypeBool,
							Optional: true,
							Default:  true,
						},
						"read_write_type": {
							Type:     schema.TypeString,
							Optional: true,
							Default:  cloudtrail.ReadWriteTypeAll,
						},
					},
				},
			},
			"home_region": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"include_global_service_events": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  true,
			},
			"insight_selector": {
				Type:     schema.TypeList,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"insight_type": {
							Type:     schema.TypeString,
							Required: true,
						},
					},
				},
			},
			"is_multi_region_trail": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},
			"is_organization_trail": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},
			"kms_key_id": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"s3_bucket_name": {
				Type:     schema.TypeString,
				Required: true,
			},
			"s3_key_prefix": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"sns_topic_name": {
				Type:     schema.TypeString,
				Optional: true,
			},
			names.AttrTags:    tftags.TagsSchema(),
			names.AttrTagsAll: tftags.TagsSchemaComputed(),
		},
	}
}

func cloudTrailUpgradeV0(_ context.Context, rawState map[string]interface{}, meta interface{}) (map[string]interface{}, error) { // nosemgrep:ci.cloudtrail-in-func-name
	if rawState == nil {
		rawState = map[string]interface{}{}
	}

	if !arn.IsARN(rawState["id"].(string)) {
		rawState["id"] = rawState["arn"]
	}

	return rawState, nil
}
