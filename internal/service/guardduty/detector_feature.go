// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package guardduty

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/guardduty"
	awstypes "github.com/aws/aws-sdk-go-v2/service/guardduty/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
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
							Type:             schema.TypeString,
							Required:         true,
							ForceNew:         true,
							ValidateDiagFunc: enum.Validate[awstypes.FeatureAdditionalConfiguration](),
						},
						names.AttrStatus: {
							Type:             schema.TypeString,
							Required:         true,
							ValidateDiagFunc: enum.Validate[awstypes.FeatureStatus](),
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
				Type:             schema.TypeString,
				Required:         true,
				ForceNew:         true,
				ValidateDiagFunc: enum.Validate[awstypes.DetectorFeature](),
			},
			names.AttrStatus: {
				Type:             schema.TypeString,
				Required:         true,
				ValidateDiagFunc: enum.Validate[awstypes.FeatureStatus](),
			},
		},
	}
}

func resourceDetectorFeaturePut(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).GuardDutyClient(ctx)

	detectorID, name := d.Get("detector_id").(string), d.Get(names.AttrName).(string)
	feature := awstypes.DetectorFeatureConfiguration{
		Name:   awstypes.DetectorFeature(name),
		Status: awstypes.FeatureStatus(d.Get(names.AttrStatus).(string)),
	}

	if v, ok := d.GetOk("additional_configuration"); ok && len(v.([]any)) > 0 {
		feature.AdditionalConfiguration = expandDetectorAdditionalConfigurations(v.([]any))
	}

	input := &guardduty.UpdateDetectorInput{
		DetectorId: aws.String(detectorID),
		Features:   []awstypes.DetectorFeatureConfiguration{feature},
	}

	_, err := conn.UpdateDetector(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "updating GuardDuty Detector (%s) Feature (%s): %s", detectorID, name, err)
	}

	if d.IsNewResource() {
		d.SetId(detectorFeatureCreateResourceID(detectorID, name))
	}

	return append(diags, resourceDetectorFeatureRead(ctx, d, meta)...)
}

func resourceDetectorFeatureRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).GuardDutyClient(ctx)

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

func FindDetectorFeatureByTwoPartKey(ctx context.Context, conn *guardduty.Client, detectorID, name string) (*awstypes.DetectorFeatureConfigurationResult, error) {
	output, err := FindDetectorByID(ctx, conn, detectorID)

	if err != nil {
		return nil, err
	}

	return tfresource.AssertSingleValueResult(tfslices.Filter(output.Features, func(v awstypes.DetectorFeatureConfigurationResult) bool {
		return string(v.Name) == name
	}))
}

func expandDetectorAdditionalConfiguration(tfMap map[string]any) awstypes.DetectorAdditionalConfiguration {
	apiObject := awstypes.DetectorAdditionalConfiguration{}

	if v, ok := tfMap[names.AttrName].(string); ok && v != "" {
		apiObject.Name = awstypes.FeatureAdditionalConfiguration(v)
	}

	if v, ok := tfMap[names.AttrStatus].(string); ok && v != "" {
		apiObject.Status = awstypes.FeatureStatus(v)
	}

	return apiObject
}

func expandDetectorAdditionalConfigurations(tfList []any) []awstypes.DetectorAdditionalConfiguration {
	if len(tfList) == 0 {
		return nil
	}

	var apiObjects []awstypes.DetectorAdditionalConfiguration

	for _, tfMapRaw := range tfList {
		tfMap, ok := tfMapRaw.(map[string]any)

		if !ok {
			continue
		}

		apiObject := expandDetectorAdditionalConfiguration(tfMap)

		apiObjects = append(apiObjects, apiObject)
	}

	return apiObjects
}

func flattenDetectorFeatureConfigurationResult(apiObject awstypes.DetectorFeatureConfigurationResult) map[string]any {
	tfMap := map[string]any{}

	if v := apiObject.AdditionalConfiguration; v != nil {
		tfMap["additional_configuration"] = flattenDetectorAdditionalConfigurationResults(v)
	}

	tfMap[names.AttrName] = string(apiObject.Name)
	tfMap[names.AttrStatus] = string(apiObject.Status)

	return tfMap
}

func flattenDetectorFeatureConfigurationResults(apiObjects []awstypes.DetectorFeatureConfigurationResult) []any {
	if len(apiObjects) == 0 {
		return nil
	}

	var tfList []any

	for _, apiObject := range apiObjects {
		tfList = append(tfList, flattenDetectorFeatureConfigurationResult(apiObject))
	}

	return tfList
}

func flattenDetectorAdditionalConfigurationResult(apiObject awstypes.DetectorAdditionalConfigurationResult) map[string]any {
	tfMap := map[string]any{}

	tfMap[names.AttrName] = string(apiObject.Name)
	tfMap[names.AttrStatus] = string(apiObject.Status)

	return tfMap
}

func flattenDetectorAdditionalConfigurationResults(apiObjects []awstypes.DetectorAdditionalConfigurationResult) []any {
	if len(apiObjects) == 0 {
		return nil
	}

	var tfList []any

	for _, apiObject := range apiObjects {
		tfList = append(tfList, flattenDetectorAdditionalConfigurationResult(apiObject))
	}

	return tfList
}
