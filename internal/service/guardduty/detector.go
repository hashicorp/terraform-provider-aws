// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

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
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_guardduty_detector", name="Detector")
// @Tags(identifierAttribute="arn")
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
			names.AttrAccountID: {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrARN: {
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
			names.AttrTags:    tftags.TagsSchema(),
			names.AttrTagsAll: tftags.TagsSchemaComputed(),
		},

		CustomizeDiff: verify.SetTagsDiff,
	}
}

func resourceDetectorCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).GuardDutyConn(ctx)

	input := &guardduty.CreateDetectorInput{
		Enable: aws.Bool(d.Get("enable").(bool)),
		Tags:   getTagsIn(ctx),
	}

	if v, ok := d.GetOk("datasources"); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
		input.DataSources = expandDataSourceConfigurations(v.([]interface{})[0].(map[string]interface{}))
	}

	if v, ok := d.GetOk("finding_publishing_frequency"); ok {
		input.FindingPublishingFrequency = aws.String(v.(string))
	}

	output, err := conn.CreateDetectorWithContext(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating GuardDuty Detector: %s", err)
	}

	d.SetId(aws.StringValue(output.DetectorId))

	return append(diags, resourceDetectorRead(ctx, d, meta)...)
}

func resourceDetectorRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).GuardDutyConn(ctx)

	gdo, err := FindDetectorByID(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] GuardDuty Detector (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading GuardDuty Detector (%s): %s", d.Id(), err)
	}

	d.Set(names.AttrAccountID, meta.(*conns.AWSClient).AccountID)
	arn := arn.ARN{
		Partition: meta.(*conns.AWSClient).Partition,
		Region:    meta.(*conns.AWSClient).Region,
		Service:   "guardduty",
		AccountID: meta.(*conns.AWSClient).AccountID,
		Resource:  fmt.Sprintf("detector/%s", d.Id()),
	}.String()
	d.Set(names.AttrARN, arn)

	if gdo.DataSources != nil {
		if err := d.Set("datasources", []interface{}{flattenDataSourceConfigurationsResult(gdo.DataSources)}); err != nil {
			return sdkdiag.AppendErrorf(diags, "setting datasources: %s", err)
		}
	} else {
		d.Set("datasources", nil)
	}
	d.Set("enable", aws.StringValue(gdo.Status) == guardduty.DetectorStatusEnabled)
	d.Set("finding_publishing_frequency", gdo.FindingPublishingFrequency)

	setTagsOut(ctx, gdo.Tags)

	return diags
}

func resourceDetectorUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).GuardDutyConn(ctx)

	if d.HasChangesExcept(names.AttrTags, names.AttrTagsAll) {
		input := &guardduty.UpdateDetectorInput{
			DetectorId:                 aws.String(d.Id()),
			Enable:                     aws.Bool(d.Get("enable").(bool)),
			FindingPublishingFrequency: aws.String(d.Get("finding_publishing_frequency").(string)),
		}

		if d.HasChange("datasources") {
			input.DataSources = expandDataSourceConfigurations(d.Get("datasources").([]interface{})[0].(map[string]interface{}))
		}

		_, err := conn.UpdateDetectorWithContext(ctx, input)

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "updating GuardDuty Detector (%s): %s", d.Id(), err)
		}
	}

	return append(diags, resourceDetectorRead(ctx, d, meta)...)
}

func resourceDetectorDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).GuardDutyConn(ctx)

	log.Printf("[DEBUG] Deleting GuardDuty Detector: %s", d.Id())
	_, err := tfresource.RetryWhenAWSErrMessageContains(ctx, membershipPropagationTimeout, func() (interface{}, error) {
		return conn.DeleteDetectorWithContext(ctx, &guardduty.DeleteDetectorInput{
			DetectorId: aws.String(d.Id()),
		})
	}, guardduty.ErrCodeBadRequestException, "cannot delete detector while it has invited or associated members")

	if tfawserr.ErrMessageContains(err, guardduty.ErrCodeBadRequestException, "The request is rejected because the input detectorId is not owned by the current account.") {
		return diags
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

	if v, ok := tfMap["kubernetes"].([]interface{}); ok && len(v) > 0 {
		apiObject.Kubernetes = expandKubernetesConfiguration(v[0].(map[string]interface{}))
	}

	if v, ok := tfMap["malware_protection"].([]interface{}); ok && len(v) > 0 {
		apiObject.MalwareProtection = expandMalwareProtectionConfiguration(v[0].(map[string]interface{}))
	}

	if v, ok := tfMap["s3_logs"].([]interface{}); ok && len(v) > 0 {
		apiObject.S3Logs = expandS3LogsConfiguration(v[0].(map[string]interface{}))
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
		ScanEc2InstanceWithFindings: expandScanEc2InstanceWithFindings(m),
	}
}

func expandScanEc2InstanceWithFindings(tfMap map[string]interface{}) *guardduty.ScanEc2InstanceWithFindings { // nosemgrep:ci.caps3-in-func-name
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

func flattenDataSourceConfigurationsResult(apiObject *guardduty.DataSourceConfigurationsResult) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if v := apiObject.Kubernetes; v != nil {
		tfMap["kubernetes"] = []interface{}{flattenKubernetesConfiguration(v)}
	}

	if v := apiObject.MalwareProtection; v != nil {
		tfMap["malware_protection"] = []interface{}{flattenMalwareProtectionConfiguration(v)}
	}

	if v := apiObject.S3Logs; v != nil {
		tfMap["s3_logs"] = []interface{}{flattenS3LogsConfigurationResult(v)}
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
		tfMap["scan_ec2_instance_with_findings"] = []interface{}{flattenScanEc2InstanceWithFindingsResult(v)}
	}

	return tfMap
}

func flattenScanEc2InstanceWithFindingsResult(apiObject *guardduty.ScanEc2InstanceWithFindingsResult) map[string]interface{} { // nosemgrep:ci.caps3-in-func-name
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if v := apiObject.EbsVolumes; v != nil {
		tfMap["ebs_volumes"] = []interface{}{flattenEbsVolumesResult(v)}
	}

	return tfMap
}

func flattenEbsVolumesResult(apiObject *guardduty.EbsVolumesResult) map[string]interface{} { // nosemgrep:ci.caps3-in-func-name
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if v := apiObject.Status; v != nil {
		tfMap["enable"] = aws.StringValue(v) == guardduty.DataSourceStatusEnabled
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

func FindDetectorByID(ctx context.Context, conn *guardduty.GuardDuty, id string) (*guardduty.GetDetectorOutput, error) {
	input := &guardduty.GetDetectorInput{
		DetectorId: aws.String(id),
	}

	output, err := conn.GetDetectorWithContext(ctx, input)

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

// FindDetector returns the ID of the current account's active GuardDuty detector.
func FindDetector(ctx context.Context, conn *guardduty.GuardDuty) (*string, error) {
	output, err := findDetectors(ctx, conn)

	if err != nil {
		return nil, err
	}

	return tfresource.AssertSinglePtrResult(output)
}

func findDetectors(ctx context.Context, conn *guardduty.GuardDuty) ([]*string, error) {
	input := &guardduty.ListDetectorsInput{}
	var output []*string

	err := conn.ListDetectorsPagesWithContext(ctx, input, func(page *guardduty.ListDetectorsOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		output = append(output, page.DetectorIds...)

		return !lastPage
	})

	if err != nil {
		return nil, err
	}

	return output, nil
}
