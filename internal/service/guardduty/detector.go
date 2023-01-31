package guardduty

import (
	"context"
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/arn"
	"github.com/aws/aws-sdk-go/service/guardduty"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

func ResourceDetector() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceDetectorCreate,
		ReadWithoutTimeout:   resourceDetectorRead,
		UpdateWithoutTimeout: resourceDetectorUpdate,
		DeleteWithoutTimeout: resourceDetectorDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			"account_id": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"datasources": {
				Type:     schema.TypeList,
				MaxItems: 1,
				Optional: true,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"s3_logs": {
							Type:     schema.TypeList,
							Optional: true,
							Computed: true,
							MaxItems: 1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"enable": {
										Type:     schema.TypeBool,
										Required: true,
									},
								},
							},
						},
						"kubernetes": {
							Type:     schema.TypeList,
							Optional: true,
							Computed: true,
							MaxItems: 1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"audit_logs": {
										Type:     schema.TypeList,
										Required: true,
										MaxItems: 1,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"enable": {
													Type:     schema.TypeBool,
													Required: true,
												},
											},
										},
									},
								},
							},
						},
						"malware_protection": {
							Type:     schema.TypeList,
							Optional: true,
							Computed: true,
							MaxItems: 1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"scan_ec2_instance_with_findings": {
										Type:     schema.TypeList,
										Required: true,
										MaxItems: 1,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"ebs_volumes": {
													Type:     schema.TypeList,
													Required: true,
													MaxItems: 1,
													Elem: &schema.Resource{
														Schema: map[string]*schema.Schema{
															"enable": {
																Type:     schema.TypeBool,
																Required: true,
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
					},
				},
			},

			"enable": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  true,
			},

			// finding_publishing_frequency is marked as Computed:true since
			// GuardDuty member accounts inherit setting from master account
			"finding_publishing_frequency": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},

			"tags":     tftags.TagsSchema(),
			"tags_all": tftags.TagsSchemaComputed(),
		},

		CustomizeDiff: verify.SetTagsDiff,
	}
}

func resourceDetectorCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).GuardDutyConn()
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	tags := defaultTagsConfig.MergeTags(tftags.New(d.Get("tags").(map[string]interface{})))

	input := guardduty.CreateDetectorInput{
		Enable: aws.Bool(d.Get("enable").(bool)),
	}

	if v, ok := d.GetOk("finding_publishing_frequency"); ok {
		input.FindingPublishingFrequency = aws.String(v.(string))
	}

	if v, ok := d.GetOk("datasources"); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
		input.DataSources = expandDataSourceConfigurations(v.([]interface{})[0].(map[string]interface{}))
	}

	if len(tags) > 0 {
		input.Tags = Tags(tags.IgnoreAWS())
	}

	log.Printf("[DEBUG] Creating GuardDuty Detector: %s", input)
	output, err := conn.CreateDetectorWithContext(ctx, &input)
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "Creating GuardDuty Detector failed: %s", err)
	}
	d.SetId(aws.StringValue(output.DetectorId))

	return append(diags, resourceDetectorRead(ctx, d, meta)...)
}

func resourceDetectorRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).GuardDutyConn()
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	ignoreTagsConfig := meta.(*conns.AWSClient).IgnoreTagsConfig

	input := guardduty.GetDetectorInput{
		DetectorId: aws.String(d.Id()),
	}

	log.Printf("[DEBUG] Reading GuardDuty Detector: %s", input)
	gdo, err := conn.GetDetectorWithContext(ctx, &input)
	if err != nil {
		if tfawserr.ErrMessageContains(err, guardduty.ErrCodeBadRequestException, "The request is rejected because the input detectorId is not owned by the current account.") {
			log.Printf("[WARN] GuardDuty detector %q not found, removing from state", d.Id())
			d.SetId("")
			return diags
		}
		return sdkdiag.AppendErrorf(diags, "Reading GuardDuty Detector '%s' failed: %s", d.Id(), err)
	}

	arn := arn.ARN{
		Partition: meta.(*conns.AWSClient).Partition,
		Region:    meta.(*conns.AWSClient).Region,
		Service:   "guardduty",
		AccountID: meta.(*conns.AWSClient).AccountID,
		Resource:  fmt.Sprintf("detector/%s", d.Id()),
	}.String()
	d.Set("arn", arn)

	d.Set("account_id", meta.(*conns.AWSClient).AccountID)

	if gdo.DataSources != nil {
		if err := d.Set("datasources", []interface{}{flattenDataSourceConfigurationsResult(gdo.DataSources)}); err != nil {
			return sdkdiag.AppendErrorf(diags, "setting datasources: %s", err)
		}
	} else {
		d.Set("datasources", nil)
	}

	d.Set("enable", aws.StringValue(gdo.Status) == guardduty.DetectorStatusEnabled)
	d.Set("finding_publishing_frequency", gdo.FindingPublishingFrequency)

	tags := KeyValueTags(gdo.Tags).IgnoreAWS().IgnoreConfig(ignoreTagsConfig)

	//lintignore:AWSR002
	if err := d.Set("tags", tags.RemoveDefaultConfig(defaultTagsConfig).Map()); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting tags: %s", err)
	}

	if err := d.Set("tags_all", tags.Map()); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting tags_all: %s", err)
	}

	return diags
}

func resourceDetectorUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).GuardDutyConn()

	if d.HasChangesExcept("tags", "tags_all") {
		input := guardduty.UpdateDetectorInput{
			DetectorId:                 aws.String(d.Id()),
			Enable:                     aws.Bool(d.Get("enable").(bool)),
			FindingPublishingFrequency: aws.String(d.Get("finding_publishing_frequency").(string)),
		}

		if d.HasChange("datasources") {
			input.DataSources = expandDataSourceConfigurations(d.Get("datasources").([]interface{})[0].(map[string]interface{}))
		}

		log.Printf("[DEBUG] Update GuardDuty Detector: %s", input)
		_, err := conn.UpdateDetectorWithContext(ctx, &input)
		if err != nil {
			return sdkdiag.AppendErrorf(diags, "Updating GuardDuty Detector '%s' failed: %s", d.Id(), err)
		}
	}

	if d.HasChange("tags_all") {
		o, n := d.GetChange("tags_all")

		if err := UpdateTags(ctx, conn, d.Get("arn").(string), o, n); err != nil {
			return sdkdiag.AppendErrorf(diags, "updating GuardDuty Detector (%s) tags: %s", d.Get("arn").(string), err)
		}
	}

	return append(diags, resourceDetectorRead(ctx, d, meta)...)
}

func resourceDetectorDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).GuardDutyConn()

	input := &guardduty.DeleteDetectorInput{
		DetectorId: aws.String(d.Id()),
	}

	err := resource.RetryContext(ctx, membershipPropagationTimeout, func() *resource.RetryError {
		_, err := conn.DeleteDetectorWithContext(ctx, input)

		if tfawserr.ErrMessageContains(err, guardduty.ErrCodeBadRequestException, "cannot delete detector while it has invited or associated members") {
			return resource.RetryableError(err)
		}

		if err != nil {
			return resource.NonRetryableError(err)
		}

		return nil
	})

	if tfresource.TimedOut(err) {
		_, err = conn.DeleteDetectorWithContext(ctx, input)
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting GuardDuty Detector (%s): %s", d.Id(), err)
	}

	return diags
}

func expandDataSourceConfigurations(tfMap map[string]interface{}) *guardduty.DataSourceConfigurations {
	if tfMap == nil {
		return nil
	}

	apiObject := &guardduty.DataSourceConfigurations{}

	if v, ok := tfMap["s3_logs"].([]interface{}); ok && len(v) > 0 {
		apiObject.S3Logs = expandS3LogsConfiguration(v[0].(map[string]interface{}))
	}
	if v, ok := tfMap["kubernetes"].([]interface{}); ok && len(v) > 0 {
		apiObject.Kubernetes = expandKubernetesConfiguration(v[0].(map[string]interface{}))
	}
	if v, ok := tfMap["malware_protection"].([]interface{}); ok && len(v) > 0 {
		apiObject.MalwareProtection = expandMalwareProtectionConfiguration(v[0].(map[string]interface{}))
	}
	return apiObject
}

func expandS3LogsConfiguration(tfMap map[string]interface{}) *guardduty.S3LogsConfiguration {
	if tfMap == nil {
		return nil
	}

	apiObject := &guardduty.S3LogsConfiguration{}

	if v, ok := tfMap["enable"].(bool); ok {
		apiObject.Enable = aws.Bool(v)
	}

	return apiObject
}

func expandKubernetesConfiguration(tfMap map[string]interface{}) *guardduty.KubernetesConfiguration {
	if tfMap == nil {
		return nil
	}

	l, ok := tfMap["audit_logs"].([]interface{})
	if !ok || len(l) == 0 {
		return nil
	}

	m, ok := l[0].(map[string]interface{})
	if !ok {
		return nil
	}

	return &guardduty.KubernetesConfiguration{
		AuditLogs: expandKubernetesAuditLogsConfiguration(m),
	}
}

func expandKubernetesAuditLogsConfiguration(tfMap map[string]interface{}) *guardduty.KubernetesAuditLogsConfiguration {
	if tfMap == nil {
		return nil
	}

	apiObject := &guardduty.KubernetesAuditLogsConfiguration{}

	if v, ok := tfMap["enable"].(bool); ok {
		apiObject.Enable = aws.Bool(v)
	}

	return apiObject
}

func expandMalwareProtectionConfiguration(tfMap map[string]interface{}) *guardduty.MalwareProtectionConfiguration {
	if tfMap == nil {
		return nil
	}

	l, ok := tfMap["scan_ec2_instance_with_findings"].([]interface{})
	if !ok || len(l) == 0 {
		return nil
	}

	m, ok := l[0].(map[string]interface{})
	if !ok {
		return nil
	}

	return &guardduty.MalwareProtectionConfiguration{
		ScanEc2InstanceWithFindings: expandMalwareProtectionScanEC2InstanceWithFindingsConfiguration(m),
	}
}

func expandMalwareProtectionScanEC2InstanceWithFindingsConfiguration(tfMap map[string]interface{}) *guardduty.ScanEc2InstanceWithFindings {
	if tfMap == nil {
		return nil
	}

	l, ok := tfMap["ebs_volumes"].([]interface{})
	if !ok || len(l) == 0 {
		return nil
	}

	m, ok := l[0].(map[string]interface{})
	if !ok {
		return nil
	}

	apiObject := &guardduty.ScanEc2InstanceWithFindings{
		EbsVolumes: expandMalwareProtectionEBSVolumesConfiguration(m),
	}
	return apiObject
}

func expandMalwareProtectionEBSVolumesConfiguration(tfMap map[string]interface{}) *bool {
	if tfMap == nil {
		return nil
	}

	var apiObject *bool

	if v, ok := tfMap["enable"].(bool); ok {
		apiObject = aws.Bool(v)
	}

	return apiObject
}

func flattenDataSourceConfigurationsResult(apiObject *guardduty.DataSourceConfigurationsResult) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if v := apiObject.S3Logs; v != nil {
		tfMap["s3_logs"] = []interface{}{flattenS3LogsConfigurationResult(v)}
	}

	if v := apiObject.Kubernetes; v != nil {
		tfMap["kubernetes"] = []interface{}{flattenKubernetesConfiguration(v)}
	}

	if v := apiObject.MalwareProtection; v != nil {
		tfMap["malware_protection"] = []interface{}{flattenMalwareProtectionConfiguration(v)}
	}

	return tfMap
}

func flattenS3LogsConfigurationResult(apiObject *guardduty.S3LogsConfigurationResult) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if v := apiObject.Status; v != nil {
		tfMap["enable"] = aws.StringValue(v) == guardduty.DataSourceStatusEnabled
	}

	return tfMap
}

func flattenKubernetesConfiguration(apiObject *guardduty.KubernetesConfigurationResult) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if v := apiObject.AuditLogs; v != nil {
		tfMap["audit_logs"] = []interface{}{flattenKubernetesAuditLogsConfiguration(v)}
	}

	return tfMap
}

func flattenKubernetesAuditLogsConfiguration(apiObject *guardduty.KubernetesAuditLogsConfigurationResult) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if v := apiObject.Status; v != nil {
		tfMap["enable"] = aws.StringValue(v) == guardduty.DataSourceStatusEnabled
	}

	return tfMap
}

func flattenMalwareProtectionConfiguration(apiObject *guardduty.MalwareProtectionConfigurationResult) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if v := apiObject.ScanEc2InstanceWithFindings; v != nil {
		tfMap["scan_ec2_instance_with_findings"] = []interface{}{flattenMalwareProtectionScanEC2InstanceWithFindingsConfigurationResult(v)}
	}

	return tfMap
}

func flattenMalwareProtectionScanEC2InstanceWithFindingsConfigurationResult(apiObject *guardduty.ScanEc2InstanceWithFindingsResult) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if v := apiObject.EbsVolumes; v != nil {
		tfMap["ebs_volumes"] = []interface{}{flattenMalwareProtectionEBSVolumesConfigurationResult(v)}
	}

	return tfMap
}

func flattenMalwareProtectionEBSVolumesConfigurationResult(apiObject *guardduty.EbsVolumesResult) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if v := apiObject.Status; v != nil {
		tfMap["enable"] = aws.StringValue(v) == guardduty.DataSourceStatusEnabled
	}

	return tfMap
}
