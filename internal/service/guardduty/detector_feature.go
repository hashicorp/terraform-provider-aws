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
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_guardduty_detector_feature", name="Detector Feature")
func ResourceDetectorFeature() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceDetectorFeaturePut,
		ReadWithoutTimeout:   resourceDetectorFeatureRead,
		UpdateWithoutTimeout: resourceDetectorFeaturePut,
		DeleteWithoutTimeout: schema.NoopContext,

		Schema: map[string]*schema.Schema{
			"additional_configuration": {
				Optional: true,
				ForceNew: true,
				Type:     schema.TypeList,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						names.AttrName: {
							Type:         schema.TypeString,
							Required:     true,
							ForceNew:     true,
							ValidateFunc: validation.StringInSlice(guardduty.FeatureAdditionalConfiguration_Values(), false),
						},
						names.AttrStatus: {
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
			names.AttrName: {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringInSlice(guardduty.DetectorFeature_Values(), false),
			},
			names.AttrStatus: {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringInSlice(guardduty.FeatureStatus_Values(), false),
			},
		},
	}
}

func resourceDetectorFeaturePut(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).GuardDutyConn(ctx)

	detectorID, name := d.Get("detector_id").(string), d.Get(names.AttrName).(string)
	feature := &guardduty.DetectorFeatureConfiguration{
		Name:   aws.String(name),
		Status: aws.String(d.Get(names.AttrStatus).(string)),
	}

	if v, ok := d.GetOk("additional_configuration"); ok && len(v.([]interface{})) > 0 {
		feature.AdditionalConfiguration = expandDetectorAdditionalConfigurations(v.([]interface{}))
	}

	input := &guardduty.UpdateDetectorInput{
		DetectorId: aws.String(detectorID),
		Features:   []*guardduty.DetectorFeatureConfiguration{feature},
	}

	_, err := conn.UpdateDetectorWithContext(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "updating GuardDuty Detector (%s) Feature (%s): %s", detectorID, name, err)
	}

	if d.IsNewResource() {
		d.SetId(detectorFeatureCreateResourceID(detectorID, name))
	}

	return append(diags, resourceDetectorFeatureRead(ctx, d, meta)...)
}

func resourceDetectorFeatureRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).GuardDutyConn(ctx)

	detectorID, name, err := detectorFeatureParseResourceID(d.Id())
	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	feature, err := FindDetectorFeatureByTwoPartKey(ctx, conn, detectorID, name)

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] GuardDuty Detector Feature (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading GuardDuty Detector Feature (%s): %s", d.Id(), err)
	}

	if err := d.Set("additional_configuration", flattenDetectorAdditionalConfigurationResults(feature.AdditionalConfiguration)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting additional_configuration: %s", err)
	}
	d.Set("detector_id", detectorID)
	d.Set(names.AttrName, feature.Name)
	d.Set(names.AttrStatus, feature.Status)

	return diags
}

const detectorFeatureResourceIDSeparator = "/"

func detectorFeatureCreateResourceID(detectorID, name string) string {
	parts := []string{detectorID, name}
	id := strings.Join(parts, detectorFeatureResourceIDSeparator)

	return id
}

func detectorFeatureParseResourceID(id string) (string, string, error) {
	parts := strings.Split(id, detectorFeatureResourceIDSeparator)

	if len(parts) == 2 && parts[0] != "" && parts[1] != "" {
		return parts[0], parts[1], nil
	}

	return "", "", fmt.Errorf("unexpected format for ID (%[1]s), expected DETECTORID%[2]sFEATURENAME", id, detectorFeatureResourceIDSeparator)
}

func FindDetectorFeatureByTwoPartKey(ctx context.Context, conn *guardduty.GuardDuty, detectorID, name string) (*guardduty.DetectorFeatureConfigurationResult, error) {
	output, err := FindDetectorByID(ctx, conn, detectorID)

	if err != nil {
		return nil, err
	}

	return tfresource.AssertSinglePtrResult(tfslices.Filter(output.Features, func(v *guardduty.DetectorFeatureConfigurationResult) bool {
		return aws.StringValue(v.Name) == name
	}))
}

func expandDetectorAdditionalConfiguration(tfMap map[string]interface{}) *guardduty.DetectorAdditionalConfiguration {
	if tfMap == nil {
		return nil
	}

	apiObject := &guardduty.DetectorAdditionalConfiguration{}

	if v, ok := tfMap[names.AttrName].(string); ok && v != "" {
		apiObject.Name = aws.String(v)
	}

	if v, ok := tfMap[names.AttrStatus].(string); ok && v != "" {
		apiObject.Status = aws.String(v)
	}

	return apiObject
}

func expandDetectorAdditionalConfigurations(tfList []interface{}) []*guardduty.DetectorAdditionalConfiguration {
	if len(tfList) == 0 {
		return nil
	}

	var apiObjects []*guardduty.DetectorAdditionalConfiguration

	for _, tfMapRaw := range tfList {
		tfMap, ok := tfMapRaw.(map[string]interface{})

		if !ok {
			continue
		}

		apiObject := expandDetectorAdditionalConfiguration(tfMap)

		if apiObject == nil {
			continue
		}

		apiObjects = append(apiObjects, apiObject)
	}

	return apiObjects
}

func flattenDetectorFeatureConfigurationResult(apiObject *guardduty.DetectorFeatureConfigurationResult) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if v := apiObject.AdditionalConfiguration; v != nil {
		tfMap["additional_configuration"] = flattenDetectorAdditionalConfigurationResults(v)
	}

	if v := apiObject.Name; v != nil {
		tfMap[names.AttrName] = aws.StringValue(v)
	}

	if v := apiObject.Status; v != nil {
		tfMap[names.AttrStatus] = aws.StringValue(v)
	}

	return tfMap
}

func flattenDetectorFeatureConfigurationResults(apiObjects []*guardduty.DetectorFeatureConfigurationResult) []interface{} {
	if len(apiObjects) == 0 {
		return nil
	}

	var tfList []interface{}

	for _, apiObject := range apiObjects {
		if apiObject == nil {
			continue
		}

		tfList = append(tfList, flattenDetectorFeatureConfigurationResult(apiObject))
	}

	return tfList
}

func flattenDetectorAdditionalConfigurationResult(apiObject *guardduty.DetectorAdditionalConfigurationResult) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if v := apiObject.Name; v != nil {
		tfMap[names.AttrName] = aws.StringValue(v)
	}

	if v := apiObject.Status; v != nil {
		tfMap[names.AttrStatus] = aws.StringValue(v)
	}

	return tfMap
}

func flattenDetectorAdditionalConfigurationResults(apiObjects []*guardduty.DetectorAdditionalConfigurationResult) []interface{} {
	if len(apiObjects) == 0 {
		return nil
	}

	var tfList []interface{}

	for _, apiObject := range apiObjects {
		if apiObject == nil {
			continue
		}

		tfList = append(tfList, flattenDetectorAdditionalConfigurationResult(apiObject))
	}

	return tfList
}
