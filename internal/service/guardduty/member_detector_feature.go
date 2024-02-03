// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package guardduty

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/guardduty"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	tfslices "github.com/hashicorp/terraform-provider-aws/internal/slices"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

// @SDKResource("aws_guardduty_member_detector_feature", name="Member Detector Feature")
func ResourceMemberDetectorFeature() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceMemberDetectorFeaturePut,
		ReadWithoutTimeout:   resourceMemberDetectorFeatureRead,
		UpdateWithoutTimeout: resourceMemberDetectorFeaturePut,
		DeleteWithoutTimeout: schema.NoopContext,

		Schema: map[string]*schema.Schema{
			"account_id": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: verify.ValidAccountID,
			},
			"additional_configuration": {
				Optional: true,
				ForceNew: true,
				Type:     schema.TypeList,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"name": {
							Type:         schema.TypeString,
							Required:     true,
							ForceNew:     true,
							ValidateFunc: validation.StringInSlice(guardduty.FeatureAdditionalConfiguration_Values(), false),
						},
						"status": {
							Type:         schema.TypeString,
							Required:     true,
							ValidateFunc: validation.StringInSlice(guardduty.FeatureStatus_Values(), false),
						},
					},
				},
			},
			"detector_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"name": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringInSlice(guardduty.DetectorFeature_Values(), false),
			},
			"status": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringInSlice(guardduty.FeatureStatus_Values(), false),
			},
		},
	}
}

func resourceMemberDetectorFeaturePut(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).GuardDutyConn(ctx)

	detectorID, accountID, name := d.Get("detector_id").(string), d.Get("account_id").(string), d.Get("name").(string)

	// Use a mutex to ensure that multiple features being updated concurrently don't trample on each other.
	conns.GlobalMutexKV.Lock(detectorID)
	defer conns.GlobalMutexKV.Unlock(detectorID)

	feature := &guardduty.MemberFeaturesConfiguration{
		Name:   aws.String(name),
		Status: aws.String(d.Get("status").(string)),
	}

	if v, ok := d.GetOk("additional_configuration"); ok && len(v.([]interface{})) > 0 {
		feature.AdditionalConfiguration = expandMemberDetectorAdditionalConfigurations(v.([]interface{}))
	}

	input := &guardduty.UpdateMemberDetectorsInput{
		DetectorId: aws.String(detectorID),
		AccountIds: []*string{aws.String(accountID)},
		Features:   []*guardduty.MemberFeaturesConfiguration{feature},
	}

	output, err := conn.UpdateMemberDetectorsWithContext(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "updating GuardDuty Member Detector (%s) Feature (%s): %s", detectorID, name, err)
	}

	// {"unprocessedAccounts":[{"result":"The request is rejected because the given account ID is not an associated member of account the current account.","accountId":"123456789012"}]}
	if len(output.UnprocessedAccounts) > 0 {
		return sdkdiag.AppendErrorf(diags, "updating GuardDuty Member Detector (%s) Feature (%s): %s", detectorID, name, aws.StringValue(output.UnprocessedAccounts[0].Result))
	}

	if d.IsNewResource() {
		d.SetId(memberDetectorFeatureCreateResourceID(detectorID, accountID, name))
	}

	return append(diags, resourceMemberDetectorFeatureRead(ctx, d, meta)...)
}

func resourceMemberDetectorFeatureRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).GuardDutyConn(ctx)

	detectorID, accountID, name, err := memberDetectorFeatureParseResourceID(d.Id())
	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	feature, err := FindMemberDetectorFeatureByThreePartKey(ctx, conn, detectorID, accountID, name)

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] GuardDuty Member Detector Feature (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading GuardDuty Member Detector Feature (%s): %s", d.Id(), err)
	}

	if err := d.Set("additional_configuration", flattenMemberDetectorAdditionalConfigurationResults(feature.AdditionalConfiguration)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting additional_configuration: %s", err)
	}

	d.Set("detector_id", detectorID)
	d.Set("account_id", accountID)
	d.Set("name", feature.Name)
	d.Set("status", feature.Status)

	return diags
}

const memberDetectorFeatureResourceIDSeparator = "/"

func memberDetectorFeatureCreateResourceID(detectorID, accountID, name string) string {
	parts := []string{detectorID, accountID, name}
	id := strings.Join(parts, memberDetectorFeatureResourceIDSeparator)

	return id
}

func memberDetectorFeatureParseResourceID(id string) (string, string, string, error) {
	parts := strings.Split(id, memberDetectorFeatureResourceIDSeparator)

	if len(parts) == 3 && parts[0] != "" && parts[1] != "" && parts[2] != "" {
		return parts[0], parts[1], parts[2], nil
	}

	return "", "", "", fmt.Errorf("unexpected format for ID (%[1]s), expected DETECTORID%[2]sACCOUNTID%[2]sFEATURENAME", id, memberDetectorFeatureResourceIDSeparator)
}

func FindMemberDetectorFeatureByThreePartKey(ctx context.Context, conn *guardduty.GuardDuty, detectorID, accountID, name string) (*guardduty.MemberFeaturesConfigurationResult, error) {
	output, err := findMemberConfigurationByDetectorAndAccountID(ctx, conn, detectorID, accountID)

	if err != nil {
		return nil, err
	}

	return tfresource.AssertSinglePtrResult(tfslices.Filter(output.Features, func(v *guardduty.MemberFeaturesConfigurationResult) bool {
		return aws.StringValue(v.Name) == name
	}))
}

func expandMemberDetectorAdditionalConfiguration(tfMap map[string]interface{}) *guardduty.MemberAdditionalConfiguration {
	if tfMap == nil {
		return nil
	}

	apiObject := &guardduty.MemberAdditionalConfiguration{}

	if v, ok := tfMap["name"].(string); ok && v != "" {
		apiObject.Name = aws.String(v)
	}

	if v, ok := tfMap["status"].(string); ok && v != "" {
		apiObject.Status = aws.String(v)
	}

	return apiObject
}

func expandMemberDetectorAdditionalConfigurations(tfList []interface{}) []*guardduty.MemberAdditionalConfiguration {
	if len(tfList) == 0 {
		return nil
	}

	var apiObjects []*guardduty.MemberAdditionalConfiguration

	for _, tfMapRaw := range tfList {
		tfMap, ok := tfMapRaw.(map[string]interface{})

		if !ok {
			continue
		}

		apiObject := expandMemberDetectorAdditionalConfiguration(tfMap)

		if apiObject == nil {
			continue
		}

		apiObjects = append(apiObjects, apiObject)
	}

	return apiObjects
}

func flattenMemberDetectorAdditionalConfigurationResult(apiObject *guardduty.MemberAdditionalConfigurationResult) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if v := apiObject.Name; v != nil {
		tfMap["name"] = aws.StringValue(v)
	}

	if v := apiObject.Status; v != nil {
		tfMap["status"] = aws.StringValue(v)
	}

	return tfMap
}

func flattenMemberDetectorAdditionalConfigurationResults(apiObjects []*guardduty.MemberAdditionalConfigurationResult) []interface{} {
	if len(apiObjects) == 0 {
		return nil
	}

	var tfList []interface{}

	for _, apiObject := range apiObjects {
		if apiObject == nil {
			continue
		}

		tfList = append(tfList, flattenMemberDetectorAdditionalConfigurationResult(apiObject))
	}

	return tfList
}

func findMemberConfigurationByDetectorAndAccountID(ctx context.Context, conn *guardduty.GuardDuty, detectorID string, accountID string) (*guardduty.MemberDataSourceConfiguration, error) {
	input := &guardduty.GetMemberDetectorsInput{
		DetectorId: aws.String(detectorID),
		AccountIds: []*string{aws.String(accountID)},
	}

	output, err := conn.GetMemberDetectorsWithContext(ctx, input)

	if err != nil {
		return nil, err
	}

	if output == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	if output.MemberDataSourceConfigurations == nil || len(output.MemberDataSourceConfigurations) == 0 {
		return nil, tfresource.NewEmptyResultError(input)
	}

	return output.MemberDataSourceConfigurations[0], nil
}
