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

// @SDKResource("aws_guardduty_organization_configuration_feature", name="Organization Configuration Feature")
func ResourceOrganizationConfigurationFeature() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceOrganizationConfigurationFeaturePut,
		ReadWithoutTimeout:   resourceOrganizationConfigurationFeatureRead,
		UpdateWithoutTimeout: resourceOrganizationConfigurationFeaturePut,
		DeleteWithoutTimeout: schema.NoopContext,

		Schema: map[string]*schema.Schema{
			"additional_configuration": {
				Optional: true,
				ForceNew: true,
				Type:     schema.TypeList,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"auto_enable": {
							Type:         schema.TypeString,
							Required:     true,
							ValidateFunc: validation.StringInSlice(guardduty.OrgFeatureStatus_Values(), false),
						},
						names.AttrName: {
							Type:         schema.TypeString,
							Required:     true,
							ForceNew:     true,
							ValidateFunc: validation.StringInSlice(guardduty.OrgFeatureAdditionalConfiguration_Values(), false),
						},
					},
				},
			},
			"auto_enable": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringInSlice(guardduty.OrgFeatureStatus_Values(), false),
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
				ValidateFunc: validation.StringInSlice(guardduty.OrgFeature_Values(), false),
			},
		},
	}
}

func resourceOrganizationConfigurationFeaturePut(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).GuardDutyConn(ctx)

	detectorID := d.Get("detector_id").(string)

	// We have seen occasional acceptance test failures when updating multiple features on the same detector concurrently,
	// so use a mutex to ensure that multiple features being updated concurrently don't trample on each other.
	conns.GlobalMutexKV.Lock(detectorID)
	defer conns.GlobalMutexKV.Unlock(detectorID)

	output, err := FindOrganizationConfigurationByID(ctx, conn, detectorID)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading GuardDuty Organization Configuration (%s): %s", detectorID, err)
	}

	name := d.Get(names.AttrName).(string)
	feature := &guardduty.OrganizationFeatureConfiguration{
		AutoEnable: aws.String(d.Get("auto_enable").(string)),
		Name:       aws.String(name),
	}

	if v, ok := d.GetOk("additional_configuration"); ok && len(v.([]interface{})) > 0 {
		feature.AdditionalConfiguration = expandOrganizationAdditionalConfigurations(v.([]interface{}))
	}

	input := &guardduty.UpdateOrganizationConfigurationInput{
		AutoEnableOrganizationMembers: output.AutoEnableOrganizationMembers,
		DetectorId:                    aws.String(detectorID),
		Features:                      []*guardduty.OrganizationFeatureConfiguration{feature},
	}

	_, err = conn.UpdateOrganizationConfigurationWithContext(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "updating GuardDuty Organization Configuration (%s) Feature (%s): %s", detectorID, name, err)
	}

	if d.IsNewResource() {
		d.SetId(organizationConfigurationFeatureCreateResourceID(detectorID, name))
	}

	return append(diags, resourceOrganizationConfigurationFeatureRead(ctx, d, meta)...)
}

func resourceOrganizationConfigurationFeatureRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).GuardDutyConn(ctx)

	detectorID, name, err := organizationConfigurationFeatureParseResourceID(d.Id())
	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	feature, err := FindOrganizationConfigurationFeatureByTwoPartKey(ctx, conn, detectorID, name)

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] GuardDuty Organization Configuration Feature (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading GuardDuty Organization Configuration Feature (%s): %s", d.Id(), err)
	}

	if err := d.Set("additional_configuration", flattenOrganizationAdditionalConfigurationResults(feature.AdditionalConfiguration)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting additional_configuration: %s", err)
	}
	d.Set("auto_enable", feature.AutoEnable)
	d.Set("detector_id", detectorID)
	d.Set(names.AttrName, feature.Name)

	return diags
}

const organizationConfigurationFeatureResourceIDSeparator = "/"

func organizationConfigurationFeatureCreateResourceID(detectorID, name string) string {
	parts := []string{detectorID, name}
	id := strings.Join(parts, organizationConfigurationFeatureResourceIDSeparator)

	return id
}

func organizationConfigurationFeatureParseResourceID(id string) (string, string, error) {
	parts := strings.Split(id, organizationConfigurationFeatureResourceIDSeparator)

	if len(parts) == 2 && parts[0] != "" && parts[1] != "" {
		return parts[0], parts[1], nil
	}

	return "", "", fmt.Errorf("unexpected format for ID (%[1]s), expected DETECTORID%[2]sFEATURENAME", id, organizationConfigurationFeatureResourceIDSeparator)
}

func FindOrganizationConfigurationFeatureByTwoPartKey(ctx context.Context, conn *guardduty.GuardDuty, detectorID, name string) (*guardduty.OrganizationFeatureConfigurationResult, error) {
	output, err := FindOrganizationConfigurationByID(ctx, conn, detectorID)

	if err != nil {
		return nil, err
	}

	return tfresource.AssertSinglePtrResult(tfslices.Filter(output.Features, func(v *guardduty.OrganizationFeatureConfigurationResult) bool {
		return aws.StringValue(v.Name) == name
	}))
}

func expandOrganizationAdditionalConfiguration(tfMap map[string]interface{}) *guardduty.OrganizationAdditionalConfiguration {
	if tfMap == nil {
		return nil
	}

	apiObject := &guardduty.OrganizationAdditionalConfiguration{}

	if v, ok := tfMap["auto_enable"].(string); ok && v != "" {
		apiObject.AutoEnable = aws.String(v)
	}

	if v, ok := tfMap[names.AttrName].(string); ok && v != "" {
		apiObject.Name = aws.String(v)
	}

	return apiObject
}

func expandOrganizationAdditionalConfigurations(tfList []interface{}) []*guardduty.OrganizationAdditionalConfiguration {
	if len(tfList) == 0 {
		return nil
	}

	var apiObjects []*guardduty.OrganizationAdditionalConfiguration

	for _, tfMapRaw := range tfList {
		tfMap, ok := tfMapRaw.(map[string]interface{})

		if !ok {
			continue
		}

		apiObject := expandOrganizationAdditionalConfiguration(tfMap)

		if apiObject == nil {
			continue
		}

		apiObjects = append(apiObjects, apiObject)
	}

	return apiObjects
}

func flattenOrganizationAdditionalConfigurationResult(apiObject *guardduty.OrganizationAdditionalConfigurationResult) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if v := apiObject.AutoEnable; v != nil {
		tfMap["auto_enable"] = aws.StringValue(v)
	}

	if v := apiObject.Name; v != nil {
		tfMap[names.AttrName] = aws.StringValue(v)
	}

	return tfMap
}

func flattenOrganizationAdditionalConfigurationResults(apiObjects []*guardduty.OrganizationAdditionalConfigurationResult) []interface{} {
	if len(apiObjects) == 0 {
		return nil
	}

	var tfList []interface{}

	for _, apiObject := range apiObjects {
		if apiObject == nil {
			continue
		}

		tfList = append(tfList, flattenOrganizationAdditionalConfigurationResult(apiObject))
	}

	return tfList
}
