// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package guardduty

import (
	"context"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/guardduty"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/customdiff"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

// @SDKResource("aws_guardduty_organization_configuration", name="Organization Configuration")
func ResourceOrganizationConfiguration() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceOrganizationConfigurationPut,
		ReadWithoutTimeout:   resourceOrganizationConfigurationRead,
		UpdateWithoutTimeout: resourceOrganizationConfigurationPut,
		DeleteWithoutTimeout: schema.NoopContext,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			"auto_enable": {
				Type:         schema.TypeBool,
				Optional:     true,
				Computed:     true,
				ExactlyOneOf: []string{"auto_enable", "auto_enable_organization_members"},
				Deprecated:   "Use auto_enable_organization_members instead",
			},
			"auto_enable_organization_members": {
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				ExactlyOneOf: []string{"auto_enable", "auto_enable_organization_members"},
				ValidateFunc: validation.StringInSlice(guardduty.AutoEnableMembers_Values(), false),
			},
			"datasources": {
				Type:     schema.TypeList,
				Optional: true,
				Computed: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
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

		CustomizeDiff: customdiff.Sequence(
			func(_ context.Context, d *schema.ResourceDiff, _ interface{}) error {
				// When creating an organization configuration with AutoEnable=true,
				// AWS will automatically set AutoEnableOrganizationMembers=NEW.
				//
				// When configuring AutoEnableOrganizationMembers=ALL or NEW,
				// AWS will automatically set AutoEnable=true.
				//
				// This diff customization keeps things consistent when configuring
				// the resource against deprecation advice from AutoEnableOrganizationMembers=ALL
				// to AutoEnable=true, and it also removes the need to use
				// AutoEnable in the resource update function.

				if attr := d.GetRawConfig().GetAttr("auto_enable_organization_members"); attr.IsKnown() && !attr.IsNull() {
					return d.SetNew("auto_enable", attr.AsString() != guardduty.AutoEnableMembersNone)
				}

				if attr := d.GetRawConfig().GetAttr("auto_enable"); attr.IsKnown() && !attr.IsNull() {
					if attr.True() {
						return d.SetNew("auto_enable_organization_members", guardduty.AutoEnableMembersNew)
					} else {
						return d.SetNew("auto_enable_organization_members", guardduty.AutoEnableMembersNone)
					}
				}

				return nil
			},
		),
	}
}

func resourceOrganizationConfigurationPut(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).GuardDutyConn(ctx)

	detectorID := d.Get("detector_id").(string)
	input := &guardduty.UpdateOrganizationConfigurationInput{
		AutoEnableOrganizationMembers: aws.String(d.Get("auto_enable_organization_members").(string)),
		DetectorId:                    aws.String(detectorID),
	}

	if v, ok := d.GetOk("datasources"); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
		input.DataSources = expandOrganizationDataSourceConfigurations(v.([]interface{})[0].(map[string]interface{}))
	}

	// We have seen occasional acceptance test failures when updating multiple features on the same detector concurrently,
	// so use a mutex to ensure that multiple features being updated concurrently don't trample on each other.
	conns.GlobalMutexKV.Lock(detectorID)
	defer conns.GlobalMutexKV.Unlock(detectorID)

	_, err := conn.UpdateOrganizationConfigurationWithContext(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "updating GuardDuty Organization Configuration (%s): %s", detectorID, err)
	}

	if d.IsNewResource() {
		d.SetId(detectorID)
	}

	return append(diags, resourceOrganizationConfigurationRead(ctx, d, meta)...)
}

func resourceOrganizationConfigurationRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).GuardDutyConn(ctx)

	output, err := FindOrganizationConfigurationByID(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
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
	d.Set("auto_enable_organization_members", output.AutoEnableOrganizationMembers)
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

	if v, ok := tfMap["kubernetes"].([]interface{}); ok && len(v) > 0 {
		apiObject.Kubernetes = expandOrganizationKubernetesConfiguration(v[0].(map[string]interface{}))
	}

	if v, ok := tfMap["malware_protection"].([]interface{}); ok && len(v) > 0 {
		apiObject.MalwareProtection = expandOrganizationMalwareProtectionConfiguration(v[0].(map[string]interface{}))
	}

	if v, ok := tfMap["s3_logs"].([]interface{}); ok && len(v) > 0 {
		apiObject.S3Logs = expandOrganizationS3LogsConfiguration(v[0].(map[string]interface{}))
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
		ScanEc2InstanceWithFindings: expandOrganizationScanEc2InstanceWithFindings(m),
	}
}

func expandOrganizationScanEc2InstanceWithFindings(tfMap map[string]interface{}) *guardduty.OrganizationScanEc2InstanceWithFindings { // nosemgrep:ci.caps3-in-func-name
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
		EbsVolumes: expandOrganizationEbsVolumes(m),
	}
}

func expandOrganizationEbsVolumes(tfMap map[string]interface{}) *guardduty.OrganizationEbsVolumes { // nosemgrep:ci.caps3-in-func-name
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
		tfMap["audit_logs"] = []interface{}{flattenOrganizationKubernetesAuditLogsConfigurationResult(v)}
	}

	return tfMap
}

func flattenOrganizationKubernetesAuditLogsConfigurationResult(apiObject *guardduty.OrganizationKubernetesAuditLogsConfigurationResult) map[string]interface{} {
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
		tfMap["scan_ec2_instance_with_findings"] = []interface{}{flattenOrganizationScanEc2InstanceWithFindingsResult(v)}
	}

	return tfMap
}

func flattenOrganizationScanEc2InstanceWithFindingsResult(apiObject *guardduty.OrganizationScanEc2InstanceWithFindingsResult) map[string]interface{} { // nosemgrep:ci.caps3-in-func-name
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if v := apiObject.EbsVolumes; v != nil {
		tfMap["ebs_volumes"] = []interface{}{flattenOrganizationEbsVolumesResult(v)}
	}

	return tfMap
}

func flattenOrganizationEbsVolumesResult(apiObject *guardduty.OrganizationEbsVolumesResult) map[string]interface{} { // nosemgrep:ci.caps3-in-func-name
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if v := apiObject.AutoEnable; v != nil {
		tfMap["auto_enable"] = aws.BoolValue(v)
	}

	return tfMap
}

func FindOrganizationConfigurationByID(ctx context.Context, conn *guardduty.GuardDuty, id string) (*guardduty.DescribeOrganizationConfigurationOutput, error) {
	input := &guardduty.DescribeOrganizationConfigurationInput{
		DetectorId: aws.String(id),
	}

	output, err := conn.DescribeOrganizationConfigurationWithContext(ctx, input)

	if tfawserr.ErrMessageContains(err, guardduty.ErrCodeBadRequestException, "The request is rejected because the input detectorId is not owned by the current account.") {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	return output, nil
}
