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
							Type:             schema.TypeString,
							Required:         true,
							ValidateDiagFunc: enum.Validate[awstypes.OrgFeatureStatus](),
						},
						names.AttrName: {
							Type:             schema.TypeString,
							Required:         true,
							ForceNew:         true,
							ValidateDiagFunc: enum.Validate[awstypes.OrgFeatureAdditionalConfiguration](),
						},
					},
				},
			},
			"auto_enable": {
				Type:             schema.TypeString,
				Required:         true,
				ValidateDiagFunc: enum.Validate[awstypes.OrgFeatureStatus](),
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
				ValidateDiagFunc: enum.Validate[awstypes.OrgFeature](),
			},
		},
	}
}

func resourceOrganizationConfigurationFeaturePut(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).GuardDutyClient(ctx)

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
	feature := awstypes.OrganizationFeatureConfiguration{
		AutoEnable: awstypes.OrgFeatureStatus(d.Get("auto_enable").(string)),
		Name:       awstypes.OrgFeature(name),
	}

	if v, ok := d.GetOk("additional_configuration"); ok && len(v.([]any)) > 0 {
		feature.AdditionalConfiguration = expandOrganizationAdditionalConfigurations(v.([]any))
	}

	input := &guardduty.UpdateOrganizationConfigurationInput{
		AutoEnableOrganizationMembers: output.AutoEnableOrganizationMembers,
		DetectorId:                    aws.String(detectorID),
		Features:                      []awstypes.OrganizationFeatureConfiguration{feature},
	}

	_, err = conn.UpdateOrganizationConfiguration(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "updating GuardDuty Organization Configuration (%s) Feature (%s): %s", detectorID, name, err)
	}

	if d.IsNewResource() {
		d.SetId(organizationConfigurationFeatureCreateResourceID(detectorID, name))
	}

	return append(diags, resourceOrganizationConfigurationFeatureRead(ctx, d, meta)...)
}

func resourceOrganizationConfigurationFeatureRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).GuardDutyClient(ctx)

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

func FindOrganizationConfigurationFeatureByTwoPartKey(ctx context.Context, conn *guardduty.Client, detectorID, name string) (*awstypes.OrganizationFeatureConfigurationResult, error) {
	output, err := FindOrganizationConfigurationByID(ctx, conn, detectorID)

	if err != nil {
		return nil, err
	}

	return tfresource.AssertSingleValueResult(tfslices.Filter(output.Features, func(v awstypes.OrganizationFeatureConfigurationResult) bool {
		return string(v.Name) == name
	}))
}

func expandOrganizationAdditionalConfiguration(tfMap map[string]any) awstypes.OrganizationAdditionalConfiguration {
	apiObject := awstypes.OrganizationAdditionalConfiguration{}

	if v, ok := tfMap["auto_enable"].(string); ok && v != "" {
		apiObject.AutoEnable = awstypes.OrgFeatureStatus(v)
	}

	if v, ok := tfMap[names.AttrName].(string); ok && v != "" {
		apiObject.Name = awstypes.OrgFeatureAdditionalConfiguration(v)
	}

	return apiObject
}

func expandOrganizationAdditionalConfigurations(tfList []any) []awstypes.OrganizationAdditionalConfiguration {
	if len(tfList) == 0 {
		return nil
	}

	var apiObjects []awstypes.OrganizationAdditionalConfiguration

	for _, tfMapRaw := range tfList {
		tfMap, ok := tfMapRaw.(map[string]any)

		if !ok {
			continue
		}

		apiObject := expandOrganizationAdditionalConfiguration(tfMap)

		apiObjects = append(apiObjects, apiObject)
	}

	return apiObjects
}

func flattenOrganizationAdditionalConfigurationResult(apiObject awstypes.OrganizationAdditionalConfigurationResult) map[string]any {
	tfMap := map[string]any{}

	tfMap["auto_enable"] = string(apiObject.AutoEnable)

	tfMap[names.AttrName] = string(apiObject.Name)

	return tfMap
}

func flattenOrganizationAdditionalConfigurationResults(apiObjects []awstypes.OrganizationAdditionalConfigurationResult) []any {
	if len(apiObjects) == 0 {
		return nil
	}

	var tfList []any

	for _, apiObject := range apiObjects {
		tfList = append(tfList, flattenOrganizationAdditionalConfigurationResult(apiObject))
	}

	return tfList
}
