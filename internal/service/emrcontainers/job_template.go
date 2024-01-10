// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package emrcontainers

import (
	"context"
	"log"
	"time"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/emrcontainers"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/id"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
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

// @SDKResource("aws_emrcontainers_job_template", name="Job Template")
// @Tags(identifierAttribute="arn")
func ResourceJobTemplate() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceJobTemplateCreate,
		ReadWithoutTimeout:   resourceJobTemplateRead,
		// UpdateWithoutTimeout: resourceJobTemplateUpdate,
		DeleteWithoutTimeout: resourceJobTemplateDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Timeouts: &schema.ResourceTimeout{
			Delete: schema.DefaultTimeout(90 * time.Minute),
		},

		Schema: map[string]*schema.Schema{
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"job_template_data": {
				Type:     schema.TypeList,
				MaxItems: 1,
				Required: true,
				ForceNew: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"configuration_overrides": {
							Type:     schema.TypeList,
							MaxItems: 1,
							Optional: true,
							ForceNew: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"application_configuration": {
										Type:     schema.TypeList,
										MaxItems: 100,
										Optional: true,
										ForceNew: true,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"classification": {
													Type:     schema.TypeString,
													Required: true,
													ForceNew: true,
												},
												"configurations": {
													Type:     schema.TypeList,
													MaxItems: 100,
													Optional: true,
													ForceNew: true,
													Elem: &schema.Resource{
														Schema: map[string]*schema.Schema{
															"classification": {
																Type:     schema.TypeString,
																Optional: true,
																ForceNew: true,
															},
															"properties": {
																Type:     schema.TypeMap,
																Optional: true,
																Elem:     &schema.Schema{Type: schema.TypeString},
															},
														},
													},
												},
												"properties": {
													Type:     schema.TypeMap,
													Optional: true,
													Elem:     &schema.Schema{Type: schema.TypeString},
												},
											},
										},
									},
									"monitoring_configuration": {
										Type:     schema.TypeList,
										MaxItems: 1,
										Optional: true,
										ForceNew: true,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"cloud_watch_monitoring_configuration": {
													Type:     schema.TypeList,
													MaxItems: 1,
													Optional: true,
													ForceNew: true,
													Elem: &schema.Resource{
														Schema: map[string]*schema.Schema{
															"log_group_name": {
																Type:     schema.TypeString,
																Required: true,
																ForceNew: true,
															},
															"log_stream_name_prefix": {
																Type:     schema.TypeString,
																Optional: true,
																ForceNew: true,
															},
														},
													},
												},
												"persistent_app_ui": {
													Type:         schema.TypeString,
													Optional:     true,
													ForceNew:     true,
													ValidateFunc: validation.StringInSlice(emrcontainers.PersistentAppUI_Values(), false),
												},
												"s3_monitoring_configuration": {
													Type:     schema.TypeList,
													MaxItems: 1,
													Optional: true,
													ForceNew: true,
													Elem: &schema.Resource{
														Schema: map[string]*schema.Schema{
															"log_uri": {
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
								},
							},
						},
						"execution_role_arn": {
							Type:         schema.TypeString,
							Required:     true,
							ForceNew:     true,
							ValidateFunc: verify.ValidARN,
						},
						"job_driver": {
							Type:     schema.TypeList,
							MaxItems: 1,
							Required: true,
							ForceNew: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"spark_sql_job_driver": {
										Type:         schema.TypeList,
										MaxItems:     1,
										Optional:     true,
										ForceNew:     true,
										ExactlyOneOf: []string{"job_template_data.0.job_driver.0.spark_sql_job_driver", "job_template_data.0.job_driver.0.spark_submit_job_driver"},
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"entry_point": {
													Type:     schema.TypeString,
													Optional: true,
													ForceNew: true,
												},
												"spark_sql_parameters": {
													Type:     schema.TypeString,
													Optional: true,
													ForceNew: true,
												},
											},
										},
									},
									"spark_submit_job_driver": {
										Type:         schema.TypeList,
										MaxItems:     1,
										Optional:     true,
										ForceNew:     true,
										ExactlyOneOf: []string{"job_template_data.0.job_driver.0.spark_sql_job_driver", "job_template_data.0.job_driver.0.spark_submit_job_driver"},
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"entry_point": {
													Type:     schema.TypeString,
													Required: true,
													ForceNew: true,
												},
												"entry_point_arguments": {
													Type:     schema.TypeSet,
													Optional: true,
													ForceNew: true,
													Elem:     &schema.Schema{Type: schema.TypeString},
												},
												"spark_submit_parameters": {
													Type:     schema.TypeString,
													Optional: true,
													ForceNew: true,
												},
											},
										},
									},
								},
							},
						},
						"job_tags": {
							Type:     schema.TypeMap,
							Optional: true,
							ForceNew: true,
							Elem:     &schema.Schema{Type: schema.TypeString},
						},
						"release_label": {
							Type:     schema.TypeString,
							Required: true,
							ForceNew: true,
						},
					},
				},
			},
			"kms_key_arn": {
				Type:         schema.TypeString,
				Optional:     true,
				ForceNew:     true,
				ValidateFunc: verify.ValidARN,
			},
			"name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
				ValidateFunc: validation.All(
					validation.StringLenBetween(1, 64),
					validation.StringMatch(regexache.MustCompile(`[0-9A-Za-z_./#-]+`), "must contain only alphanumeric, hyphen, underscore, dot and # characters"),
				),
			},
			names.AttrTags:    tftags.TagsSchemaForceNew(),
			names.AttrTagsAll: tftags.TagsSchemaComputed(),
		},

		CustomizeDiff: verify.SetTagsDiff,
	}
}

func resourceJobTemplateCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	conn := meta.(*conns.AWSClient).EMRContainersConn(ctx)

	name := d.Get("name").(string)
	input := &emrcontainers.CreateJobTemplateInput{
		ClientToken: aws.String(id.UniqueId()),
		Name:        aws.String(name),
		Tags:        getTagsIn(ctx),
	}

	if v, ok := d.GetOk("job_template_data"); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
		input.JobTemplateData = expandJobTemplateData(v.([]interface{})[0].(map[string]interface{}))
	}

	if v, ok := d.GetOk("kms_key_arn"); ok {
		input.KmsKeyArn = aws.String(v.(string))
	}

	output, err := conn.CreateJobTemplateWithContext(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating EMR Containers Job Template (%s): %s", name, err)
	}

	d.SetId(aws.StringValue(output.Id))

	return append(diags, resourceJobTemplateRead(ctx, d, meta)...)
}

func resourceJobTemplateRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	conn := meta.(*conns.AWSClient).EMRContainersConn(ctx)

	vc, err := FindJobTemplateByID(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] EMR Containers Job Template %s not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading EMR Containers Job Template (%s): %s", d.Id(), err)
	}

	d.Set("arn", vc.Arn)
	if vc.JobTemplateData != nil {
		if err := d.Set("job_template_data", []interface{}{flattenJobTemplateData(vc.JobTemplateData)}); err != nil {
			return sdkdiag.AppendErrorf(diags, "setting job_template_data: %s", err)
		}
	} else {
		d.Set("job_template_data", nil)
	}
	d.Set("name", vc.Name)
	d.Set("kms_key_arn", vc.KmsKeyArn)

	setTagsOut(ctx, vc.Tags)

	return diags
}

func resourceJobTemplateDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	conn := meta.(*conns.AWSClient).EMRContainersConn(ctx)

	log.Printf("[INFO] Deleting EMR Containers Job Template: %s", d.Id())
	_, err := conn.DeleteJobTemplateWithContext(ctx, &emrcontainers.DeleteJobTemplateInput{
		Id: aws.String(d.Id()),
	})

	if tfawserr.ErrCodeEquals(err, emrcontainers.ErrCodeResourceNotFoundException) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting EMR Containers Job Template (%s): %s", d.Id(), err)
	}

	// if _, err = waitJobTemplateDeleted(ctx, conn, d.Id(), d.Timeout(schema.TimeoutDelete)); err != nil {
	// 	return diag.Errorf("waiting for EMR Containers Job Template (%s) delete: %s", d.Id(), err)
	// }

	return diags
}

func expandJobTemplateData(tfMap map[string]interface{}) *emrcontainers.JobTemplateData {
	if tfMap == nil {
		return nil
	}

	apiObject := &emrcontainers.JobTemplateData{}

	if v, ok := tfMap["configuration_overrides"].([]interface{}); ok && len(v) > 0 {
		apiObject.ConfigurationOverrides = expandConfigurationOverrides(v[0].(map[string]interface{}))
	}

	if v, ok := tfMap["execution_role_arn"].(string); ok && v != "" {
		apiObject.ExecutionRoleArn = aws.String(v)
	}

	if v, ok := tfMap["job_driver"].([]interface{}); ok && len(v) > 0 {
		apiObject.JobDriver = expandJobDriver(v[0].(map[string]interface{}))
	}

	if v, ok := tfMap["job_tags"].(map[string]interface{}); ok && len(v) > 0 {
		apiObject.JobTags = flex.ExpandStringMap(v)
	}

	if v, ok := tfMap["release_label"].(string); ok && v != "" {
		apiObject.ReleaseLabel = aws.String(v)
	}

	return apiObject
}

func expandConfigurationOverrides(tfMap map[string]interface{}) *emrcontainers.ParametricConfigurationOverrides {
	if tfMap == nil {
		return nil
	}

	apiObject := &emrcontainers.ParametricConfigurationOverrides{}

	if v, ok := tfMap["application_configuration"].([]interface{}); ok && len(v) > 0 {
		apiObject.ApplicationConfiguration = expandConfigurations(v)
	}

	if v, ok := tfMap["monitoring_configuration"].([]interface{}); ok && len(v) > 0 {
		apiObject.MonitoringConfiguration = expandMonitoringConfiguration(v[0].(map[string]interface{}))
	}

	return apiObject
}
func expandConfigurations(tfList []interface{}) []*emrcontainers.Configuration {
	if len(tfList) == 0 {
		return nil
	}

	var apiObjects []*emrcontainers.Configuration

	for _, tfMapRaw := range tfList {
		tfMap, ok := tfMapRaw.(map[string]interface{})

		if !ok {
			continue
		}

		apiObject := expandConfiguration(tfMap)

		if apiObject == nil {
			continue
		}

		apiObjects = append(apiObjects, apiObject)
	}

	return apiObjects
}

func expandConfiguration(tfMap map[string]interface{}) *emrcontainers.Configuration {
	if tfMap == nil {
		return nil
	}

	apiObject := &emrcontainers.Configuration{}

	if v, ok := tfMap["classification"].(string); ok && v != "" {
		apiObject.Classification = aws.String(v)
	}

	if v, ok := tfMap["configurations"].([]interface{}); ok && len(v) > 0 {
		apiObject.Configurations = expandConfigurations(v)
	}

	if v, ok := tfMap["properties"].(map[string]interface{}); ok && len(v) > 0 {
		apiObject.Properties = flex.ExpandStringMap(v)
	}

	return apiObject
}

func expandMonitoringConfiguration(tfMap map[string]interface{}) *emrcontainers.ParametricMonitoringConfiguration {
	if tfMap == nil {
		return nil
	}

	apiObject := &emrcontainers.ParametricMonitoringConfiguration{}

	if v, ok := tfMap["cloud_watch_monitoring_configuration"].([]interface{}); ok && len(v) > 0 {
		apiObject.CloudWatchMonitoringConfiguration = expandCloudWatchMonitoringConfiguration(v[0].(map[string]interface{}))
	}

	if v, ok := tfMap["persistent_app_ui"].(string); ok && v != "" {
		apiObject.PersistentAppUI = aws.String(v)
	}

	if v, ok := tfMap["s3_monitoring_configuration"].([]interface{}); ok && len(v) > 0 {
		apiObject.S3MonitoringConfiguration = expandS3MonitoringConfiguration(v[0].(map[string]interface{}))
	}

	return apiObject
}

func expandCloudWatchMonitoringConfiguration(tfMap map[string]interface{}) *emrcontainers.ParametricCloudWatchMonitoringConfiguration {
	if tfMap == nil {
		return nil
	}

	apiObject := &emrcontainers.ParametricCloudWatchMonitoringConfiguration{}

	if v, ok := tfMap["log_group_mame"].(string); ok && v != "" {
		apiObject.LogGroupName = aws.String(v)
	}

	if v, ok := tfMap["log_stream_name_prefix"].(string); ok && v != "" {
		apiObject.LogStreamNamePrefix = aws.String(v)
	}

	return apiObject
}

func expandS3MonitoringConfiguration(tfMap map[string]interface{}) *emrcontainers.ParametricS3MonitoringConfiguration {
	if tfMap == nil {
		return nil
	}

	apiObject := &emrcontainers.ParametricS3MonitoringConfiguration{}

	if v, ok := tfMap["log_uri"].(string); ok && v != "" {
		apiObject.LogUri = aws.String(v)
	}

	return apiObject
}

func expandJobDriver(tfMap map[string]interface{}) *emrcontainers.JobDriver {
	if tfMap == nil {
		return nil
	}

	apiObject := &emrcontainers.JobDriver{}

	if v, ok := tfMap["spark_sql_job_driver"].([]interface{}); ok && len(v) > 0 {
		apiObject.SparkSqlJobDriver = expandSparkSQLJobDriver(v[0].(map[string]interface{}))
	}

	if v, ok := tfMap["spark_submit_job_driver"].([]interface{}); ok && len(v) > 0 {
		apiObject.SparkSubmitJobDriver = expandSparkSubmitJobDriver(v[0].(map[string]interface{}))
	}

	return apiObject
}

func expandSparkSQLJobDriver(tfMap map[string]interface{}) *emrcontainers.SparkSqlJobDriver {
	if tfMap == nil {
		return nil
	}

	apiObject := &emrcontainers.SparkSqlJobDriver{}

	if v, ok := tfMap["entry_point"].(string); ok && v != "" {
		apiObject.EntryPoint = aws.String(v)
	}

	if v, ok := tfMap["spark_sql_parameters"].(string); ok && v != "" {
		apiObject.SparkSqlParameters = aws.String(v)
	}

	return apiObject
}

func expandSparkSubmitJobDriver(tfMap map[string]interface{}) *emrcontainers.SparkSubmitJobDriver {
	if tfMap == nil {
		return nil
	}

	apiObject := &emrcontainers.SparkSubmitJobDriver{}

	if v, ok := tfMap["entry_point"].(string); ok && v != "" {
		apiObject.EntryPoint = aws.String(v)
	}

	if v, ok := tfMap["entry_point_arguments"].(*schema.Set); ok && v.Len() > 0 {
		apiObject.EntryPointArguments = flex.ExpandStringSet(v)
	}

	if v, ok := tfMap["spark_submit_parameters"].(string); ok && v != "" {
		apiObject.SparkSubmitParameters = aws.String(v)
	}

	return apiObject
}

func flattenJobTemplateData(apiObject *emrcontainers.JobTemplateData) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if v := apiObject.ConfigurationOverrides; v != nil {
		tfMap["configuration_overrides"] = []interface{}{flattenConfigurationOverrides(v)}
	}

	if v := apiObject.ExecutionRoleArn; v != nil {
		tfMap["execution_role_arn"] = aws.StringValue(v)
	}

	if v := apiObject.JobDriver; v != nil {
		tfMap["job_driver"] = []interface{}{flattenJobDriver(v)}
	}

	if v := apiObject.JobTags; v != nil {
		tfMap["job_tags"] = aws.StringValueMap(v)
	}

	if v := apiObject.ReleaseLabel; v != nil {
		tfMap["release_label"] = aws.StringValue(v)
	}

	return tfMap
}

func flattenConfigurationOverrides(apiObject *emrcontainers.ParametricConfigurationOverrides) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if v := apiObject.ApplicationConfiguration; v != nil {
		tfMap["application_configuration"] = []interface{}{flattenConfigurations(v)}
	}

	if v := apiObject.MonitoringConfiguration; v != nil {
		tfMap["monitoring_configuration"] = []interface{}{flattenMonitoringConfiguration(v)}
	}

	return tfMap
}

func flattenConfigurations(apiObjects []*emrcontainers.Configuration) []interface{} {
	if len(apiObjects) == 0 {
		return nil
	}

	var tfList []interface{}

	for _, apiObject := range apiObjects {
		if apiObject == nil {
			continue
		}

		tfList = append(tfList, flattenConfiguration(apiObject))
	}

	return tfList
}

func flattenConfiguration(apiObject *emrcontainers.Configuration) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if v := apiObject.Classification; v != nil {
		tfMap["classification"] = aws.StringValue(v)
	}

	if v := apiObject.Properties; v != nil {
		tfMap["properties"] = aws.StringValueMap(v)
	}

	return tfMap
}

func flattenMonitoringConfiguration(apiObject *emrcontainers.ParametricMonitoringConfiguration) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if v := apiObject.CloudWatchMonitoringConfiguration; v != nil {
		tfMap["cloud_watch_monitoring_configuration"] = []interface{}{flattenCloudWatchMonitoringConfiguration(v)}
	}

	if v := apiObject.PersistentAppUI; v != nil {
		tfMap["persistent_app_ui"] = aws.StringValue(v)
	}

	if v := apiObject.S3MonitoringConfiguration; v != nil {
		tfMap["s3_monitoring_configuration"] = []interface{}{flattenS3MonitoringConfiguration(v)}
	}

	return tfMap
}

func flattenCloudWatchMonitoringConfiguration(apiObject *emrcontainers.ParametricCloudWatchMonitoringConfiguration) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if v := apiObject.LogGroupName; v != nil {
		tfMap["log_group_name"] = aws.StringValue(v)
	}

	if v := apiObject.LogStreamNamePrefix; v != nil {
		tfMap["log_stream_name_prefix"] = aws.StringValue(v)
	}

	return tfMap
}

func flattenS3MonitoringConfiguration(apiObject *emrcontainers.ParametricS3MonitoringConfiguration) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if v := apiObject.LogUri; v != nil {
		tfMap["log_uri"] = aws.StringValue(v)
	}

	return tfMap
}

func flattenJobDriver(apiObject *emrcontainers.JobDriver) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if v := apiObject.SparkSqlJobDriver; v != nil {
		tfMap["spark_sql_job_driver"] = []interface{}{flattenSparkSQLJobDriver(v)}
	}

	if v := apiObject.SparkSubmitJobDriver; v != nil {
		tfMap["spark_submit_job_driver"] = []interface{}{flattenSparkSubmitJobDriver(v)}
	}

	return tfMap
}

func flattenSparkSQLJobDriver(apiObject *emrcontainers.SparkSqlJobDriver) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if v := apiObject.EntryPoint; v != nil {
		tfMap["entry_point"] = aws.StringValue(v)
	}

	if v := apiObject.SparkSqlParameters; v != nil {
		tfMap["spark_sql_parameters"] = aws.StringValue(v)
	}

	return tfMap
}

func flattenSparkSubmitJobDriver(apiObject *emrcontainers.SparkSubmitJobDriver) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if v := apiObject.EntryPoint; v != nil {
		tfMap["entry_point"] = aws.StringValue(v)
	}

	if v := apiObject.EntryPointArguments; v != nil {
		tfMap["entry_point_arguments"] = flex.FlattenStringSet(v)
	}

	if v := apiObject.SparkSubmitParameters; v != nil {
		tfMap["spark_submit_parameters"] = aws.StringValue(v)
	}

	return tfMap
}

func findJobTemplate(ctx context.Context, conn *emrcontainers.EMRContainers, input *emrcontainers.DescribeJobTemplateInput) (*emrcontainers.JobTemplate, error) {
	output, err := conn.DescribeJobTemplateWithContext(ctx, input)

	if tfawserr.ErrCodeEquals(err, emrcontainers.ErrCodeResourceNotFoundException) {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil || output.JobTemplate == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	return output.JobTemplate, nil
}

func FindJobTemplateByID(ctx context.Context, conn *emrcontainers.EMRContainers, id string) (*emrcontainers.JobTemplate, error) {
	input := &emrcontainers.DescribeJobTemplateInput{
		Id: aws.String(id),
	}

	output, err := findJobTemplate(ctx, conn, input)

	if err != nil {
		return nil, err
	}

	return output, nil
}
