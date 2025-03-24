// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package guardduty

import (
	"context"
	"log"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/guardduty"
	awstypes "github.com/aws/aws-sdk-go-v2/service/guardduty/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/customdiff"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
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
				Deprecated:   "auto_enable is deprecated. Use auto_enable_organization_members instead.",
			},
			"auto_enable_organization_members": {
				Type:             schema.TypeString,
				Optional:         true,
				Computed:         true,
				ExactlyOneOf:     []string{"auto_enable", "auto_enable_organization_members"},
				ValidateDiagFunc: enum.Validate[awstypes.AutoEnableMembers](),
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
			func(_ context.Context, d *schema.ResourceDiff, _ any) error {
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
					return d.SetNew("auto_enable", attr.AsString() != string(awstypes.AutoEnableMembersNone))
				}

				if attr := d.GetRawConfig().GetAttr("auto_enable"); attr.IsKnown() && !attr.IsNull() {
					if attr.True() {
						return d.SetNew("auto_enable_organization_members", string(awstypes.AutoEnableMembersNew))
					} else {
						return d.SetNew("auto_enable_organization_members", string(awstypes.AutoEnableMembersNone))
					}
				}

				return nil
			},
		),
	}
}

func resourceOrganizationConfigurationPut(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).GuardDutyClient(ctx)

	detectorID := d.Get("detector_id").(string)
	input := &guardduty.UpdateOrganizationConfigurationInput{
		AutoEnableOrganizationMembers: awstypes.AutoEnableMembers(d.Get("auto_enable_organization_members").(string)),
		DetectorId:                    aws.String(detectorID),
	}

	if v, ok := d.GetOk("datasources"); ok && len(v.([]any)) > 0 && v.([]any)[0] != nil {
		input.DataSources = expandOrganizationDataSourceConfigurations(v.([]any)[0].(map[string]any))
	}

	// We have seen occasional acceptance test failures when updating multiple features on the same detector concurrently,
	// so use a mutex to ensure that multiple features being updated concurrently don't trample on each other.
	conns.GlobalMutexKV.Lock(detectorID)
	defer conns.GlobalMutexKV.Unlock(detectorID)

	_, err := conn.UpdateOrganizationConfiguration(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "updating GuardDuty Organization Configuration (%s): %s", detectorID, err)
	}

	if d.IsNewResource() {
		d.SetId(detectorID)
	}

	return append(diags, resourceOrganizationConfigurationRead(ctx, d, meta)...)
}

func resourceOrganizationConfigurationRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).GuardDutyClient(ctx)

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
		if err := d.Set("datasources", []any{flattenOrganizationDataSourceConfigurationsResult(output.DataSources)}); err != nil {
			return sdkdiag.AppendErrorf(diags, "setting datasources: %s", err)
		}
	} else {
		d.Set("datasources", nil)
	}
	d.Set("detector_id", d.Id())

	return diags
}

func expandOrganizationDataSourceConfigurations(tfMap map[string]any) *awstypes.OrganizationDataSourceConfigurations {
	if tfMap == nil {
		return nil
	}

	apiObject := &awstypes.OrganizationDataSourceConfigurations{}

	if v, ok := tfMap["kubernetes"].([]any); ok && len(v) > 0 {
		apiObject.Kubernetes = expandOrganizationKubernetesConfiguration(v[0].(map[string]any))
	}

	if v, ok := tfMap["malware_protection"].([]any); ok && len(v) > 0 {
		apiObject.MalwareProtection = expandOrganizationMalwareProtectionConfiguration(v[0].(map[string]any))
	}

	if v, ok := tfMap["s3_logs"].([]any); ok && len(v) > 0 {
		apiObject.S3Logs = expandOrganizationS3LogsConfiguration(v[0].(map[string]any))
	}

	return apiObject
}

func expandOrganizationS3LogsConfiguration(tfMap map[string]any) *awstypes.OrganizationS3LogsConfiguration {
	if tfMap == nil {
		return nil
	}

	apiObject := &awstypes.OrganizationS3LogsConfiguration{}

	if v, ok := tfMap["auto_enable"].(bool); ok {
		apiObject.AutoEnable = aws.Bool(v)
	}

	return apiObject
}

func expandOrganizationKubernetesConfiguration(tfMap map[string]any) *awstypes.OrganizationKubernetesConfiguration {
	if tfMap == nil {
		return nil
	}

	l, ok := tfMap["audit_logs"].([]any)
	if !ok || len(l) == 0 {
		return nil
	}

	m, ok := l[0].(map[string]any)
	if !ok {
		return nil
	}

	return &awstypes.OrganizationKubernetesConfiguration{
		AuditLogs: expandOrganizationKubernetesAuditLogsConfiguration(m),
	}
}

func expandOrganizationMalwareProtectionConfiguration(tfMap map[string]any) *awstypes.OrganizationMalwareProtectionConfiguration {
	if tfMap == nil {
		return nil
	}

	l, ok := tfMap["scan_ec2_instance_with_findings"].([]any)
	if !ok || len(l) == 0 {
		return nil
	}

	m, ok := l[0].(map[string]any)
	if !ok {
		return nil
	}

	return &awstypes.OrganizationMalwareProtectionConfiguration{
		ScanEc2InstanceWithFindings: expandOrganizationScanEc2InstanceWithFindings(m),
	}
}

func expandOrganizationScanEc2InstanceWithFindings(tfMap map[string]any) *awstypes.OrganizationScanEc2InstanceWithFindings { // nosemgrep:ci.caps3-in-func-name
	if tfMap == nil {
		return nil
	}

	l, ok := tfMap["ebs_volumes"].([]any)
	if !ok || len(l) == 0 {
		return nil
	}

	m, ok := l[0].(map[string]any)
	if !ok {
		return nil
	}

	return &awstypes.OrganizationScanEc2InstanceWithFindings{
		EbsVolumes: expandOrganizationEbsVolumes(m),
	}
}

func expandOrganizationEbsVolumes(tfMap map[string]any) *awstypes.OrganizationEbsVolumes { // nosemgrep:ci.caps3-in-func-name
	if tfMap == nil {
		return nil
	}

	apiObject := &awstypes.OrganizationEbsVolumes{}

	if v, ok := tfMap["auto_enable"].(bool); ok {
		apiObject.AutoEnable = aws.Bool(v)
	}

	return apiObject
}

func expandOrganizationKubernetesAuditLogsConfiguration(tfMap map[string]any) *awstypes.OrganizationKubernetesAuditLogsConfiguration {
	if tfMap == nil {
		return nil
	}

	apiObject := &awstypes.OrganizationKubernetesAuditLogsConfiguration{}

	if v, ok := tfMap["enable"].(bool); ok {
		apiObject.AutoEnable = aws.Bool(v)
	}

	return apiObject
}

func flattenOrganizationDataSourceConfigurationsResult(apiObject *awstypes.OrganizationDataSourceConfigurationsResult) map[string]any {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]any{}

	if v := apiObject.S3Logs; v != nil {
		tfMap["s3_logs"] = []any{flattenOrganizationS3LogsConfigurationResult(v)}
	}
	if v := apiObject.Kubernetes; v != nil {
		tfMap["kubernetes"] = []any{flattenOrganizationKubernetesConfigurationResult(v)}
	}
	if v := apiObject.MalwareProtection; v != nil {
		tfMap["malware_protection"] = []any{flattenOrganizationMalwareProtectionConfigurationResult(v)}
	}
	return tfMap
}

func flattenOrganizationS3LogsConfigurationResult(apiObject *awstypes.OrganizationS3LogsConfigurationResult) map[string]any {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]any{}

	if v := apiObject.AutoEnable; v != nil {
		tfMap["auto_enable"] = aws.ToBool(v)
	}

	return tfMap
}

func flattenOrganizationKubernetesConfigurationResult(apiObject *awstypes.OrganizationKubernetesConfigurationResult) map[string]any {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]any{}

	if v := apiObject.AuditLogs; v != nil {
		tfMap["audit_logs"] = []any{flattenOrganizationKubernetesAuditLogsConfigurationResult(v)}
	}

	return tfMap
}

func flattenOrganizationKubernetesAuditLogsConfigurationResult(apiObject *awstypes.OrganizationKubernetesAuditLogsConfigurationResult) map[string]any {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]any{}

	if v := apiObject.AutoEnable; v != nil {
		tfMap["enable"] = aws.ToBool(v)
	}

	return tfMap
}

func flattenOrganizationMalwareProtectionConfigurationResult(apiObject *awstypes.OrganizationMalwareProtectionConfigurationResult) map[string]any {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]any{}

	if v := apiObject.ScanEc2InstanceWithFindings; v != nil {
		tfMap["scan_ec2_instance_with_findings"] = []any{flattenOrganizationScanEc2InstanceWithFindingsResult(v)}
	}

	return tfMap
}

func flattenOrganizationScanEc2InstanceWithFindingsResult(apiObject *awstypes.OrganizationScanEc2InstanceWithFindingsResult) map[string]any { // nosemgrep:ci.caps3-in-func-name
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]any{}

	if v := apiObject.EbsVolumes; v != nil {
		tfMap["ebs_volumes"] = []any{flattenOrganizationEbsVolumesResult(v)}
	}

	return tfMap
}

func flattenOrganizationEbsVolumesResult(apiObject *awstypes.OrganizationEbsVolumesResult) map[string]any { // nosemgrep:ci.caps3-in-func-name
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]any{}

	if v := apiObject.AutoEnable; v != nil {
		tfMap["auto_enable"] = aws.ToBool(v)
	}

	return tfMap
}

func FindOrganizationConfigurationByID(ctx context.Context, conn *guardduty.Client, id string) (*guardduty.DescribeOrganizationConfigurationOutput, error) {
	input := &guardduty.DescribeOrganizationConfigurationInput{
		DetectorId: aws.String(id),
	}

	output, err := conn.DescribeOrganizationConfiguration(ctx, input)

	if errs.IsAErrorMessageContains[*awstypes.BadRequestException](err, "The request is rejected because the input detectorId is not owned by the current account.") {
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
