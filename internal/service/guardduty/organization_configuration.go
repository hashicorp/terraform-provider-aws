package guardduty

import (
	"context"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/guardduty"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
)

func ResourceOrganizationConfiguration() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceOrganizationConfigurationUpdate,
		ReadWithoutTimeout:   resourceOrganizationConfigurationRead,
		UpdateWithoutTimeout: resourceOrganizationConfigurationUpdate,
		DeleteWithoutTimeout: schema.NoopContext,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			"auto_enable": {
				Type:     schema.TypeBool,
				Required: true,
			},

			"datasources": {
				Type:     schema.TypeList,
				Optional: true,
				Computed: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"s3_logs": {
							Type:     schema.TypeList,
							Optional: true,
							Computed: true,
							MaxItems: 1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"auto_enable": {
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
															"auto_enable": {
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

			"detector_id": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.NoZeroValues,
			},
		},
	}
}

func resourceOrganizationConfigurationUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).GuardDutyConn()

	detectorID := d.Get("detector_id").(string)

	input := &guardduty.UpdateOrganizationConfigurationInput{
		AutoEnable: aws.Bool(d.Get("auto_enable").(bool)),
		DetectorId: aws.String(detectorID),
	}

	if v, ok := d.GetOk("datasources"); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
		input.DataSources = expandOrganizationDataSourceConfigurations(v.([]interface{})[0].(map[string]interface{}))
	}

	_, err := conn.UpdateOrganizationConfigurationWithContext(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "updating GuardDuty Organization Configuration (%s): %s", detectorID, err)
	}

	d.SetId(detectorID)

	return append(diags, resourceOrganizationConfigurationRead(ctx, d, meta)...)
}

func resourceOrganizationConfigurationRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).GuardDutyConn()

	input := &guardduty.DescribeOrganizationConfigurationInput{
		DetectorId: aws.String(d.Id()),
	}

	output, err := conn.DescribeOrganizationConfigurationWithContext(ctx, input)

	if tfawserr.ErrMessageContains(err, guardduty.ErrCodeBadRequestException, "The request is rejected because the input detectorId is not owned by the current account.") {
		log.Printf("[WARN] GuardDuty Organization Configuration (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading GuardDuty Organization Configuration (%s): %s", d.Id(), err)
	}

	if output == nil {
		return sdkdiag.AppendErrorf(diags, "reading GuardDuty Organization Configuration (%s): empty response", d.Id())
	}

	d.Set("auto_enable", output.AutoEnable)

	if output.DataSources != nil {
		if err := d.Set("datasources", []interface{}{flattenOrganizationDataSourceConfigurationsResult(output.DataSources)}); err != nil {
			return sdkdiag.AppendErrorf(diags, "setting datasources: %s", err)
		}
	} else {
		d.Set("datasources", nil)
	}

	d.Set("detector_id", d.Id())

	return diags
}

func expandOrganizationDataSourceConfigurations(tfMap map[string]interface{}) *guardduty.OrganizationDataSourceConfigurations {
	if tfMap == nil {
		return nil
	}

	apiObject := &guardduty.OrganizationDataSourceConfigurations{}

	if v, ok := tfMap["s3_logs"].([]interface{}); ok && len(v) > 0 {
		apiObject.S3Logs = expandOrganizationS3LogsConfiguration(v[0].(map[string]interface{}))
	}

	if v, ok := tfMap["kubernetes"].([]interface{}); ok && len(v) > 0 {
		apiObject.Kubernetes = expandOrganizationKubernetesConfiguration(v[0].(map[string]interface{}))
	}

	if v, ok := tfMap["malware_protection"].([]interface{}); ok && len(v) > 0 {
		apiObject.MalwareProtection = expandOrganizationMalwareProtectionConfiguration(v[0].(map[string]interface{}))
	}

	return apiObject
}

func expandOrganizationS3LogsConfiguration(tfMap map[string]interface{}) *guardduty.OrganizationS3LogsConfiguration {
	if tfMap == nil {
		return nil
	}

	apiObject := &guardduty.OrganizationS3LogsConfiguration{}

	if v, ok := tfMap["auto_enable"].(bool); ok {
		apiObject.AutoEnable = aws.Bool(v)
	}

	return apiObject
}

func expandOrganizationKubernetesConfiguration(tfMap map[string]interface{}) *guardduty.OrganizationKubernetesConfiguration {
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

	return &guardduty.OrganizationKubernetesConfiguration{
		AuditLogs: expandOrganizationKubernetesAuditLogsConfiguration(m),
	}
}

func expandOrganizationMalwareProtectionConfiguration(tfMap map[string]interface{}) *guardduty.OrganizationMalwareProtectionConfiguration {
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

	return &guardduty.OrganizationMalwareProtectionConfiguration{
		ScanEc2InstanceWithFindings: expandOrganizationScanEC2InstanceWithFindingsConfiguration(m),
	}
}

func expandOrganizationScanEC2InstanceWithFindingsConfiguration(tfMap map[string]interface{}) *guardduty.OrganizationScanEc2InstanceWithFindings {
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

	return &guardduty.OrganizationScanEc2InstanceWithFindings{
		EbsVolumes: expandOrganizationEBSVolumesConfiguration(m),
	}
}

func expandOrganizationEBSVolumesConfiguration(tfMap map[string]interface{}) *guardduty.OrganizationEbsVolumes {
	if tfMap == nil {
		return nil
	}

	apiObject := &guardduty.OrganizationEbsVolumes{}

	if v, ok := tfMap["auto_enable"].(bool); ok {
		apiObject.AutoEnable = aws.Bool(v)
	}

	return apiObject
}

func expandOrganizationKubernetesAuditLogsConfiguration(tfMap map[string]interface{}) *guardduty.OrganizationKubernetesAuditLogsConfiguration {
	if tfMap == nil {
		return nil
	}

	apiObject := &guardduty.OrganizationKubernetesAuditLogsConfiguration{}

	if v, ok := tfMap["enable"].(bool); ok {
		apiObject.AutoEnable = aws.Bool(v)
	}

	return apiObject
}

func flattenOrganizationDataSourceConfigurationsResult(apiObject *guardduty.OrganizationDataSourceConfigurationsResult) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if v := apiObject.S3Logs; v != nil {
		tfMap["s3_logs"] = []interface{}{flattenOrganizationS3LogsConfigurationResult(v)}
	}
	if v := apiObject.Kubernetes; v != nil {
		tfMap["kubernetes"] = []interface{}{flattenOrganizationKubernetesConfigurationResult(v)}
	}
	if v := apiObject.MalwareProtection; v != nil {
		tfMap["malware_protection"] = []interface{}{flattenOrganizationMalwareProtectionConfigurationResult(v)}
	}
	return tfMap
}

func flattenOrganizationS3LogsConfigurationResult(apiObject *guardduty.OrganizationS3LogsConfigurationResult) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if v := apiObject.AutoEnable; v != nil {
		tfMap["auto_enable"] = aws.BoolValue(v)
	}

	return tfMap
}

func flattenOrganizationKubernetesConfigurationResult(apiObject *guardduty.OrganizationKubernetesConfigurationResult) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if v := apiObject.AuditLogs; v != nil {
		tfMap["audit_logs"] = []interface{}{flattenOrganizationKubernetesAuditLogsConfiguration(v)}
	}

	return tfMap
}

func flattenOrganizationKubernetesAuditLogsConfiguration(apiObject *guardduty.OrganizationKubernetesAuditLogsConfigurationResult) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if v := apiObject.AutoEnable; v != nil {
		tfMap["enable"] = aws.BoolValue(v)
	}

	return tfMap
}

func flattenOrganizationMalwareProtectionConfigurationResult(apiObject *guardduty.OrganizationMalwareProtectionConfigurationResult) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if v := apiObject.ScanEc2InstanceWithFindings; v != nil {
		tfMap["scan_ec2_instance_with_findings"] = []interface{}{flattenOrganizationMalwareProtectionScanEC2InstanceWithFindingsResult(v)}
	}

	return tfMap
}

func flattenOrganizationMalwareProtectionScanEC2InstanceWithFindingsResult(apiObject *guardduty.OrganizationScanEc2InstanceWithFindingsResult) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if v := apiObject.EbsVolumes; v != nil {
		tfMap["ebs_volumes"] = []interface{}{flattenOrganizationMalwareProtectionEBSVolumesResult(v)}
	}

	return tfMap
}

func flattenOrganizationMalwareProtectionEBSVolumesResult(apiObject *guardduty.OrganizationEbsVolumesResult) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if v := apiObject.AutoEnable; v != nil {
		tfMap["auto_enable"] = aws.BoolValue(v)
	}

	return tfMap
}
