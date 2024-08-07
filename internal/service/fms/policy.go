// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package fms

import (
	"context"
	"log"
	"time"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/fms"
	awstypes "github.com/aws/aws-sdk-go-v2/service/fms/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/structure"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_fms_policy", name="Policy")
// @Tags(identifierAttribute="arn")
func resourcePolicy() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourcePolicyCreate,
		ReadWithoutTimeout:   resourcePolicyRead,
		UpdateWithoutTimeout: resourcePolicyUpdate,
		DeleteWithoutTimeout: resourcePolicyDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		CustomizeDiff: verify.SetTagsDiff,

		Schema: map[string]*schema.Schema{
			names.AttrARN: {
				Type:     schema.TypeString,
				Computed: true,
			},
			"delete_all_policy_resources": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  true,
			},
			"delete_unused_fm_managed_resources": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},
			names.AttrDescription: {
				Type:     schema.TypeString,
				Optional: true,
			},
			"exclude_resource_tags": {
				Type:     schema.TypeBool,
				Required: true,
			},
			"exclude_map": {
				Type:             schema.TypeList,
				MaxItems:         1,
				Optional:         true,
				DiffSuppressFunc: verify.SuppressMissingOptionalConfigurationBlock,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"account": {
							Type:     schema.TypeSet,
							Optional: true,
							Elem: &schema.Schema{
								Type: schema.TypeString,
							},
						},
						"orgunit": {
							Type:     schema.TypeSet,
							Optional: true,
							Elem: &schema.Schema{
								Type: schema.TypeString,
							},
						},
					},
				},
			},
			"include_map": {
				Type:             schema.TypeList,
				MaxItems:         1,
				Optional:         true,
				DiffSuppressFunc: verify.SuppressMissingOptionalConfigurationBlock,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"account": {
							Type:     schema.TypeSet,
							Optional: true,
							Elem: &schema.Schema{
								Type: schema.TypeString,
							},
						},
						"orgunit": {
							Type:     schema.TypeSet,
							Optional: true,
							Elem: &schema.Schema{
								Type: schema.TypeString,
							},
						},
					},
				},
			},
			names.AttrName: {
				Type:     schema.TypeString,
				Required: true,
			},
			"policy_update_token": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"remediation_enabled": {
				Type:     schema.TypeBool,
				Optional: true,
			},
			names.AttrResourceTags: tftags.TagsSchema(),
			names.AttrResourceType: {
				Type:          schema.TypeString,
				Optional:      true,
				Computed:      true,
				ValidateFunc:  validation.StringMatch(regexache.MustCompile(`^([\p{L}\p{Z}\p{N}_.:/=+\-@]*)$`), "must match a supported resource type, such as AWS::EC2::VPC, see also: https://docs.aws.amazon.com/fms/2018-01-01/APIReference/API_Policy.html"),
				ConflictsWith: []string{"resource_type_list"},
			},
			"resource_set_ids": {
				Type:     schema.TypeSet,
				Optional: true,
				Computed: true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
			"resource_type_list": {
				Type:     schema.TypeSet,
				Optional: true,
				Computed: true,
				Elem: &schema.Schema{
					Type:         schema.TypeString,
					ValidateFunc: validation.StringMatch(regexache.MustCompile(`^([\p{L}\p{Z}\p{N}_.:/=+\-@]*)$`), "must match a supported resource type, such as AWS::EC2::VPC, see also: https://docs.aws.amazon.com/fms/2018-01-01/APIReference/API_Policy.html"),
				},
				ConflictsWith: []string{names.AttrResourceType},
			},
			"security_service_policy_data": {
				Type:     schema.TypeList,
				Required: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"managed_service_data": {
							Type:                  schema.TypeString,
							Optional:              true,
							ValidateFunc:          validation.StringIsJSON,
							DiffSuppressFunc:      suppressEquivalentManagedServiceDataJSON,
							DiffSuppressOnRefresh: true,
							StateFunc: func(v interface{}) string {
								json, _ := structure.NormalizeJsonString(v)
								return json
							},
						},
						"policy_option": {
							Type:     schema.TypeList,
							Optional: true,
							MaxItems: 1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"network_firewall_policy": {
										Type:     schema.TypeList,
										Optional: true,
										MaxItems: 1,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"firewall_deployment_model": {
													Type:             schema.TypeString,
													Optional:         true,
													ValidateDiagFunc: enum.Validate[awstypes.FirewallDeploymentModel](),
												},
											},
										},
									},
									"third_party_firewall_policy": {
										Type:     schema.TypeList,
										Optional: true,
										MaxItems: 1,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"firewall_deployment_model": {
													Type:             schema.TypeString,
													Optional:         true,
													ValidateDiagFunc: enum.Validate[awstypes.FirewallDeploymentModel](),
												},
											},
										},
									},
								},
							},
						},
						names.AttrType: {
							Type:     schema.TypeString,
							Required: true,
						},
					},
				},
			},
			names.AttrTags:    tftags.TagsSchema(),
			names.AttrTagsAll: tftags.TagsSchemaComputed(),
		},
	}
}

func resourcePolicyCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).FMSClient(ctx)

	input := &fms.PutPolicyInput{
		Policy:  expandPolicy(d),
		TagList: getTagsIn(ctx),
	}

	// System problems can arise during FMS policy updates (maybe also creation),
	// so we set the following operation as retryable.
	// Reference: https://github.com/hashicorp/terraform-provider-aws/issues/23946.
	const (
		timeout = 1 * time.Minute
	)
	outputRaw, err := tfresource.RetryWhenIsA[*awstypes.InternalErrorException](ctx, timeout, func() (interface{}, error) {
		return conn.PutPolicy(ctx, input)
	})

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating FMS Policy: %s", err)
	}

	d.SetId(aws.ToString(outputRaw.(*fms.PutPolicyOutput).Policy.PolicyId))

	return append(diags, resourcePolicyRead(ctx, d, meta)...)
}

func resourcePolicyRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).FMSClient(ctx)

	output, err := findPolicyByID(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] FMS Policy %s not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading FMS Policy (%s): %s", d.Id(), err)
	}

	arn := aws.ToString(output.PolicyArn)
	d.Set(names.AttrARN, arn)
	policy := output.Policy
	d.Set("delete_unused_fm_managed_resources", policy.DeleteUnusedFMManagedResources)
	d.Set(names.AttrDescription, policy.PolicyDescription)
	if err := d.Set("exclude_map", flattenPolicyMap(policy.ExcludeMap)); err != nil {
		sdkdiag.AppendErrorf(diags, "setting exclude_map: %s", err)
	}
	d.Set("exclude_resource_tags", policy.ExcludeResourceTags)
	if err := d.Set("include_map", flattenPolicyMap(policy.IncludeMap)); err != nil {
		sdkdiag.AppendErrorf(diags, "setting include_map: %s", err)
	}
	d.Set(names.AttrName, policy.PolicyName)
	d.Set("policy_update_token", policy.PolicyUpdateToken)
	d.Set("remediation_enabled", policy.RemediationEnabled)
	if err := d.Set(names.AttrResourceTags, flattenResourceTags(policy.ResourceTags)); err != nil {
		sdkdiag.AppendErrorf(diags, "setting resource_tags: %s", err)
	}
	d.Set(names.AttrResourceType, policy.ResourceType)
	d.Set("resource_type_list", policy.ResourceTypeList)
	d.Set("resource_set_ids", policy.ResourceSetIds)
	securityServicePolicy := []map[string]interface{}{{
		names.AttrType:         string(policy.SecurityServicePolicyData.Type),
		"managed_service_data": aws.ToString(policy.SecurityServicePolicyData.ManagedServiceData),
		"policy_option":        flattenPolicyOption(policy.SecurityServicePolicyData.PolicyOption),
	}}
	if err := d.Set("security_service_policy_data", securityServicePolicy); err != nil {
		sdkdiag.AppendErrorf(diags, "setting security_service_policy_data: %s", err)
	}

	return diags
}

func resourcePolicyUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).FMSClient(ctx)

	if d.HasChangesExcept(names.AttrTags, names.AttrTagsAll) {
		input := &fms.PutPolicyInput{
			Policy: expandPolicy(d),
		}

		// System problems can arise during FMS policy updates (maybe also creation),
		// so we set the following operation as retryable.
		// Reference: https://github.com/hashicorp/terraform-provider-aws/issues/23946.
		const (
			timeout = 1 * time.Minute
		)
		_, err := tfresource.RetryWhenIsA[*awstypes.InternalErrorException](ctx, timeout, func() (interface{}, error) {
			return conn.PutPolicy(ctx, input)
		})

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "updating FMS Policy (%s): %s", d.Id(), err)
		}
	}

	return append(diags, resourcePolicyRead(ctx, d, meta)...)
}

func resourcePolicyDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).FMSClient(ctx)

	log.Printf("[DEBUG] Deleting FMS Policy: %s", d.Id())
	_, err := conn.DeletePolicy(ctx, &fms.DeletePolicyInput{
		PolicyId:                 aws.String(d.Id()),
		DeleteAllPolicyResources: d.Get("delete_all_policy_resources").(bool),
	})

	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting FMS Policy (%s): %s", d.Id(), err)
	}

	return diags
}

func findPolicyByID(ctx context.Context, conn *fms.Client, id string) (*fms.GetPolicyOutput, error) {
	input := &fms.GetPolicyInput{
		PolicyId: aws.String(id),
	}

	output, err := conn.GetPolicy(ctx, input)

	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil || output.Policy == nil || output.Policy.SecurityServicePolicyData == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	return output, nil
}

func expandPolicy(d *schema.ResourceData) *awstypes.Policy {
	resourceType := aws.String("ResourceTypeList")
	if v, ok := d.GetOk(names.AttrResourceType); ok {
		resourceType = aws.String(v.(string))
	}

	apiObject := &awstypes.Policy{
		DeleteUnusedFMManagedResources: d.Get("delete_unused_fm_managed_resources").(bool),
		ExcludeMap:                     expandPolicyMap(d.Get("exclude_map").([]interface{})),
		ExcludeResourceTags:            d.Get("exclude_resource_tags").(bool),
		IncludeMap:                     expandPolicyMap(d.Get("include_map").([]interface{})),
		PolicyDescription:              aws.String(d.Get(names.AttrDescription).(string)),
		PolicyName:                     aws.String(d.Get(names.AttrName).(string)),
		RemediationEnabled:             d.Get("remediation_enabled").(bool),
		ResourceType:                   resourceType,
		ResourceTypeList:               flex.ExpandStringValueSet(d.Get("resource_type_list").(*schema.Set)),
		ResourceSetIds:                 flex.ExpandStringValueSet(d.Get("resource_set_ids").(*schema.Set)),
	}

	if d.Id() != "" {
		apiObject.PolicyId = aws.String(d.Id())
		apiObject.PolicyUpdateToken = aws.String(d.Get("policy_update_token").(string))
	}

	if v, ok := d.GetOk(names.AttrResourceTags); ok && len(v.(map[string]interface{})) > 0 {
		for k, v := range flex.ExpandStringValueMap(v.(map[string]interface{})) {
			apiObject.ResourceTags = append(apiObject.ResourceTags, awstypes.ResourceTag{
				Key:   aws.String(k),
				Value: aws.String(v),
			})
		}
	}

	tfMap := d.Get("security_service_policy_data").([]interface{})[0].(map[string]interface{})
	apiObject.SecurityServicePolicyData = &awstypes.SecurityServicePolicyData{
		ManagedServiceData: aws.String(tfMap["managed_service_data"].(string)),
		Type:               awstypes.SecurityServiceType(tfMap[names.AttrType].(string)),
	}

	if v, ok := tfMap["policy_option"].([]interface{}); ok && len(v) > 0 && v[0] != nil {
		apiObject.SecurityServicePolicyData.PolicyOption = expandPolicyOption(v[0].(map[string]interface{}))
	}

	return apiObject
}

func expandPolicyOption(tfMap map[string]interface{}) *awstypes.PolicyOption {
	if tfMap == nil {
		return nil
	}

	apiObject := &awstypes.PolicyOption{}

	if v, ok := tfMap["network_firewall_policy"].([]interface{}); ok && len(v) > 0 && v[0] != nil {
		apiObject.NetworkFirewallPolicy = expandPolicyOptionNetworkFirewall(v[0].(map[string]interface{}))
	}

	if v, ok := tfMap["third_party_firewall_policy"].([]interface{}); ok && len(v) > 0 && v[0] != nil {
		apiObject.ThirdPartyFirewallPolicy = expandPolicyOptionThirdPartyFirewall(v[0].(map[string]interface{}))
	}

	return apiObject
}

func expandPolicyOptionNetworkFirewall(tfMap map[string]interface{}) *awstypes.NetworkFirewallPolicy {
	if tfMap == nil {
		return nil
	}

	apiObject := &awstypes.NetworkFirewallPolicy{}

	if v, ok := tfMap["firewall_deployment_model"].(string); ok {
		apiObject.FirewallDeploymentModel = awstypes.FirewallDeploymentModel(v)
	}

	return apiObject
}

func expandPolicyOptionThirdPartyFirewall(tfMap map[string]interface{}) *awstypes.ThirdPartyFirewallPolicy {
	if tfMap == nil {
		return nil
	}

	apiObject := &awstypes.ThirdPartyFirewallPolicy{}

	if v, ok := tfMap["firewall_deployment_model"].(string); ok {
		apiObject.FirewallDeploymentModel = awstypes.FirewallDeploymentModel(v)
	}

	return apiObject
}

func expandPolicyMap(tfList []interface{}) map[string][]string {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap := tfList[0].(map[string]interface{})

	if tfMap == nil {
		return nil
	}

	apiObject := map[string][]string{}

	for k, v := range tfMap {
		switch v := flex.ExpandStringValueSet(v.(*schema.Set)); k {
		case "account":
			apiObject["ACCOUNT"] = v
		case "orgunit":
			apiObject["ORG_UNIT"] = v
		}
	}

	return apiObject
}

func flattenPolicyMap(apiObject map[string][]string) []interface{} {
	tfMap := map[string]interface{}{}

	for k, v := range apiObject {
		switch k {
		case "ACCOUNT":
			tfMap["account"] = v
		case "ORG_UNIT":
			tfMap["orgunit"] = v
		}
	}

	return []interface{}{tfMap}
}

func flattenPolicyOption(fmsPolicyOption *awstypes.PolicyOption) []interface{} {
	if fmsPolicyOption == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if v := fmsPolicyOption.NetworkFirewallPolicy; v != nil {
		tfMap["network_firewall_policy"] = flattenPolicyOptionNetworkFirewall(fmsPolicyOption.NetworkFirewallPolicy)
	}

	if v := fmsPolicyOption.ThirdPartyFirewallPolicy; v != nil {
		tfMap["third_party_firewall_policy"] = flattenPolicyOptionThirdPartyFirewall(fmsPolicyOption.ThirdPartyFirewallPolicy)
	}

	return []interface{}{tfMap}
}

func flattenPolicyOptionNetworkFirewall(fmsNetworkFirewallPolicy *awstypes.NetworkFirewallPolicy) []interface{} {
	if fmsNetworkFirewallPolicy == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if v := fmsNetworkFirewallPolicy.FirewallDeploymentModel; v != "" {
		tfMap["firewall_deployment_model"] = string(v)
	}

	return []interface{}{tfMap}
}

func flattenPolicyOptionThirdPartyFirewall(fmsThirdPartyFirewallPolicy *awstypes.ThirdPartyFirewallPolicy) []interface{} {
	if fmsThirdPartyFirewallPolicy == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if v := fmsThirdPartyFirewallPolicy.FirewallDeploymentModel; v != "" {
		tfMap["firewall_deployment_model"] = string(v)
	}

	return []interface{}{tfMap}
}

func flattenResourceTags(apiObjects []awstypes.ResourceTag) map[string]interface{} {
	tfMap := map[string]interface{}{}

	for _, v := range apiObjects {
		tfMap[aws.ToString(v.Key)] = aws.ToString(v.Value)
	}

	return tfMap
}
