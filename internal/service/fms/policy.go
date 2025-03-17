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
// @Testing(serialize=true)
// @Testing(importIgnore="delete_all_policy_resources;policy_update_token")
func resourcePolicy() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourcePolicyCreate,
		ReadWithoutTimeout:   resourcePolicyRead,
		UpdateWithoutTimeout: resourcePolicyUpdate,
		DeleteWithoutTimeout: resourcePolicyDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		SchemaFunc: func() map[string]*schema.Schema {
			networkACLEntrySetNestedBlock := func() *schema.Schema {
				return &schema.Schema{
					Type:     schema.TypeSet,
					Optional: true,
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{
							names.AttrCIDRBlock: {
								Type:     schema.TypeString,
								Optional: true,
							},
							"egress": {
								Type:     schema.TypeBool,
								Required: true,
							},
							"icmp_type_code": {
								Type:     schema.TypeList,
								Optional: true,
								Elem: &schema.Resource{
									Schema: map[string]*schema.Schema{
										"code": {
											Type:     schema.TypeInt,
											Optional: true,
										},
										names.AttrType: {
											Type:     schema.TypeInt,
											Optional: true,
										},
									},
								},
							},
							"ipv6_cidr_block": {
								Type:     schema.TypeString,
								Optional: true,
							},
							"port_range": {
								Type:     schema.TypeList,
								Optional: true,
								Elem: &schema.Resource{
									Schema: map[string]*schema.Schema{
										"from": {
											Type:     schema.TypeInt,
											Optional: true,
										},
										"to": {
											Type:     schema.TypeInt,
											Optional: true,
										},
									},
								},
							},
							names.AttrProtocol: {
								Type:     schema.TypeString,
								Required: true,
							},
							"rule_action": {
								Type:             schema.TypeString,
								Required:         true,
								ValidateDiagFunc: enum.Validate[awstypes.NetworkAclRuleAction](),
							},
						},
					},
				}
			}

			return map[string]*schema.Schema{
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
								StateFunc: func(v any) string {
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
										"network_acl_common_policy": {
											Type:     schema.TypeList,
											Optional: true,
											MaxItems: 1,
											Elem: &schema.Resource{
												Schema: map[string]*schema.Schema{
													"network_acl_entry_set": {
														Type:     schema.TypeList,
														Optional: true,
														MaxItems: 1,
														Elem: &schema.Resource{
															Schema: map[string]*schema.Schema{
																"last_entry":  networkACLEntrySetNestedBlock(),
																"first_entry": networkACLEntrySetNestedBlock(),
																"force_remediate_for_first_entries": {
																	Type:     schema.TypeBool,
																	Required: true,
																},
																"force_remediate_for_last_entries": {
																	Type:     schema.TypeBool,
																	Required: true,
																},
															},
														},
													},
												},
											},
										},
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
			}
		},
	}
}

func resourcePolicyCreate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
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
	outputRaw, err := tfresource.RetryWhenIsA[*awstypes.InternalErrorException](ctx, timeout, func() (any, error) {
		return conn.PutPolicy(ctx, input)
	})

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating FMS Policy: %s", err)
	}

	d.SetId(aws.ToString(outputRaw.(*fms.PutPolicyOutput).Policy.PolicyId))

	return append(diags, resourcePolicyRead(ctx, d, meta)...)
}

func resourcePolicyRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
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
		diags = sdkdiag.AppendErrorf(diags, "setting exclude_map: %s", err)
	}
	d.Set("exclude_resource_tags", policy.ExcludeResourceTags)
	if err := d.Set("include_map", flattenPolicyMap(policy.IncludeMap)); err != nil {
		diags = sdkdiag.AppendErrorf(diags, "setting include_map: %s", err)
	}
	d.Set(names.AttrName, policy.PolicyName)
	d.Set("policy_update_token", policy.PolicyUpdateToken)
	d.Set("remediation_enabled", policy.RemediationEnabled)
	if err := d.Set(names.AttrResourceTags, flattenResourceTags(policy.ResourceTags)); err != nil {
		diags = sdkdiag.AppendErrorf(diags, "setting resource_tags: %s", err)
	}
	d.Set(names.AttrResourceType, policy.ResourceType)
	d.Set("resource_type_list", policy.ResourceTypeList)
	d.Set("resource_set_ids", policy.ResourceSetIds)
	securityServicePolicy := []map[string]any{{
		names.AttrType:         string(policy.SecurityServicePolicyData.Type),
		"managed_service_data": aws.ToString(policy.SecurityServicePolicyData.ManagedServiceData),
		"policy_option":        flattenPolicyOption(policy.SecurityServicePolicyData.PolicyOption),
	}}
	if err := d.Set("security_service_policy_data", securityServicePolicy); err != nil {
		diags = sdkdiag.AppendErrorf(diags, "setting security_service_policy_data: %s", err)
	}

	return diags
}

func resourcePolicyUpdate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
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
		_, err := tfresource.RetryWhenIsA[*awstypes.InternalErrorException](ctx, timeout, func() (any, error) {
			return conn.PutPolicy(ctx, input)
		})

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "updating FMS Policy (%s): %s", d.Id(), err)
		}
	}

	return append(diags, resourcePolicyRead(ctx, d, meta)...)
}

func resourcePolicyDelete(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
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
		ExcludeMap:                     expandPolicyMap(d.Get("exclude_map").([]any)),
		ExcludeResourceTags:            d.Get("exclude_resource_tags").(bool),
		IncludeMap:                     expandPolicyMap(d.Get("include_map").([]any)),
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

	if v, ok := d.GetOk(names.AttrResourceTags); ok && len(v.(map[string]any)) > 0 {
		for k, v := range flex.ExpandStringValueMap(v.(map[string]any)) {
			apiObject.ResourceTags = append(apiObject.ResourceTags, awstypes.ResourceTag{
				Key:   aws.String(k),
				Value: aws.String(v),
			})
		}
	}

	tfMap := d.Get("security_service_policy_data").([]any)[0].(map[string]any)
	apiObject.SecurityServicePolicyData = &awstypes.SecurityServicePolicyData{
		ManagedServiceData: aws.String(tfMap["managed_service_data"].(string)),
		Type:               awstypes.SecurityServiceType(tfMap[names.AttrType].(string)),
	}

	if v, ok := tfMap["policy_option"].([]any); ok && len(v) > 0 && v[0] != nil {
		apiObject.SecurityServicePolicyData.PolicyOption = expandPolicyOption(v[0].(map[string]any))
	}

	return apiObject
}

func expandPolicyOption(tfMap map[string]any) *awstypes.PolicyOption {
	if tfMap == nil {
		return nil
	}

	apiObject := &awstypes.PolicyOption{}

	if v, ok := tfMap["network_acl_common_policy"].([]any); ok && len(v) > 0 && v[0] != nil {
		apiObject.NetworkAclCommonPolicy = expandNetworkACLCommonPolicy(v[0].(map[string]any))
	}

	if v, ok := tfMap["network_firewall_policy"].([]any); ok && len(v) > 0 && v[0] != nil {
		apiObject.NetworkFirewallPolicy = expandNetworkFirewallPolicy(v[0].(map[string]any))
	}

	if v, ok := tfMap["third_party_firewall_policy"].([]any); ok && len(v) > 0 && v[0] != nil {
		apiObject.ThirdPartyFirewallPolicy = expandThirdPartyFirewallPolicy(v[0].(map[string]any))
	}

	return apiObject
}

func expandNetworkACLCommonPolicy(tfMap map[string]any) *awstypes.NetworkAclCommonPolicy {
	if tfMap == nil {
		return nil
	}

	apiObject := &awstypes.NetworkAclCommonPolicy{}

	if v, ok := tfMap["network_acl_entry_set"].([]any); ok && len(v) > 0 && v[0] != nil {
		apiObject.NetworkAclEntrySet = expandNetworkACLEntrySet(v[0].(map[string]any))
	}

	return apiObject
}

func expandNetworkACLEntrySet(tfMap map[string]any) *awstypes.NetworkAclEntrySet {
	if tfMap == nil {
		return nil
	}

	apiObject := &awstypes.NetworkAclEntrySet{}

	if v, ok := tfMap["first_entry"].(*schema.Set); ok && v.Len() > 0 {
		apiObject.FirstEntries = expandNetworkACLEntries(v.List())
	}

	if v, ok := tfMap["force_remediate_for_first_entries"].(bool); ok {
		apiObject.ForceRemediateForFirstEntries = aws.Bool(v)
	}

	if v, ok := tfMap["force_remediate_for_last_entries"].(bool); ok {
		apiObject.ForceRemediateForLastEntries = aws.Bool(v)
	}

	if v, ok := tfMap["last_entry"].(*schema.Set); ok && v.Len() > 0 {
		apiObject.LastEntries = expandNetworkACLEntries(v.List())
	}

	return apiObject
}

func expandNetworkACLEntries(tfList []any) []awstypes.NetworkAclEntry {
	if len(tfList) == 0 {
		return nil
	}

	apiObjects := []awstypes.NetworkAclEntry{}

	for _, tfMapRaw := range tfList {
		tfMap, ok := tfMapRaw.(map[string]any)
		if !ok {
			continue
		}

		apiObjects = append(apiObjects, expandNetworkACLEntry(tfMap))
	}

	return apiObjects
}

func expandNetworkACLEntry(tfMap map[string]any) awstypes.NetworkAclEntry {
	apiObject := awstypes.NetworkAclEntry{}

	if v, ok := tfMap[names.AttrCIDRBlock].(string); ok && v != "" {
		apiObject.CidrBlock = aws.String(v)
	}

	if v, ok := tfMap["egress"].(bool); ok {
		apiObject.Egress = aws.Bool(v)
	}

	if v, ok := tfMap["icmp_type_code"].([]any); ok && len(v) > 0 {
		apiObject.IcmpTypeCode = expandNetworkACLIcmpTypeCode(v[0].(map[string]any))
	}

	if v, ok := tfMap["ipv6_cidr_block"].(string); ok && v != "" {
		apiObject.Ipv6CidrBlock = aws.String(v)
	}

	if v, ok := tfMap["port_range"].([]any); ok && len(v) > 0 {
		apiObject.PortRange = expandNetworkACLPortRange(v[0].(map[string]any))
	}

	if v, ok := tfMap[names.AttrProtocol].(string); ok && v != "" {
		apiObject.Protocol = aws.String(v)
	}

	if v, ok := tfMap["rule_action"].(string); ok && v != "" {
		apiObject.RuleAction = awstypes.NetworkAclRuleAction(v)
	}

	return apiObject
}

func expandNetworkACLIcmpTypeCode(tfMap map[string]any) *awstypes.NetworkAclIcmpTypeCode {
	if tfMap == nil {
		return nil
	}

	apiObject := &awstypes.NetworkAclIcmpTypeCode{}

	apiObject.Code = aws.Int32(int32(tfMap["code"].(int)))
	apiObject.Type = aws.Int32(int32(tfMap[names.AttrType].(int)))

	return apiObject
}

func expandNetworkACLPortRange(tfMap map[string]any) *awstypes.NetworkAclPortRange {
	if tfMap == nil {
		return nil
	}

	apiObject := &awstypes.NetworkAclPortRange{}

	apiObject.From = aws.Int32(int32(tfMap["from"].(int)))
	apiObject.To = aws.Int32(int32(tfMap["to"].(int)))

	return apiObject
}

func expandNetworkFirewallPolicy(tfMap map[string]any) *awstypes.NetworkFirewallPolicy {
	if tfMap == nil {
		return nil
	}

	apiObject := &awstypes.NetworkFirewallPolicy{}

	if v, ok := tfMap["firewall_deployment_model"].(string); ok {
		apiObject.FirewallDeploymentModel = awstypes.FirewallDeploymentModel(v)
	}

	return apiObject
}

func expandThirdPartyFirewallPolicy(tfMap map[string]any) *awstypes.ThirdPartyFirewallPolicy {
	if tfMap == nil {
		return nil
	}

	apiObject := &awstypes.ThirdPartyFirewallPolicy{}

	if v, ok := tfMap["firewall_deployment_model"].(string); ok {
		apiObject.FirewallDeploymentModel = awstypes.FirewallDeploymentModel(v)
	}

	return apiObject
}

func expandPolicyMap(tfList []any) map[string][]string {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap := tfList[0].(map[string]any)

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

func flattenPolicyMap(apiObject map[string][]string) []any {
	tfMap := map[string]any{}

	for k, v := range apiObject {
		switch k {
		case "ACCOUNT":
			tfMap["account"] = v
		case "ORG_UNIT":
			tfMap["orgunit"] = v
		}
	}

	return []any{tfMap}
}

func flattenPolicyOption(apiObject *awstypes.PolicyOption) []any {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]any{}

	if v := apiObject.NetworkAclCommonPolicy; v != nil {
		tfMap["network_acl_common_policy"] = flattenNetworkACLCommonPolicy(apiObject.NetworkAclCommonPolicy)
	}

	if v := apiObject.NetworkFirewallPolicy; v != nil {
		tfMap["network_firewall_policy"] = flattenNetworkFirewallPolicy(apiObject.NetworkFirewallPolicy)
	}

	if v := apiObject.ThirdPartyFirewallPolicy; v != nil {
		tfMap["third_party_firewall_policy"] = flattenThirdPartyFirewallPolicy(apiObject.ThirdPartyFirewallPolicy)
	}

	return []any{tfMap}
}

func flattenNetworkACLCommonPolicy(apiObject *awstypes.NetworkAclCommonPolicy) []any {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]any{}

	if v := apiObject.NetworkAclEntrySet; v != nil {
		tfMap["network_acl_entry_set"] = flattenNetworkACLEntrySet(v)
	}

	return []any{tfMap}
}

func flattenNetworkACLEntrySet(apiObject *awstypes.NetworkAclEntrySet) []any {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]any{}

	if v := apiObject.FirstEntries; v != nil {
		tfMap["first_entry"] = flattenNetworkACLEntries(v)
	}

	if v := apiObject.ForceRemediateForFirstEntries; v != nil {
		tfMap["force_remediate_for_first_entries"] = aws.ToBool(v)
	}

	if v := apiObject.ForceRemediateForLastEntries; v != nil {
		tfMap["force_remediate_for_last_entries"] = aws.ToBool(v)
	}

	if v := apiObject.LastEntries; v != nil {
		tfMap["last_entry"] = flattenNetworkACLEntries(v)
	}

	return []any{tfMap}
}

func flattenNetworkACLEntries(apiObjects []awstypes.NetworkAclEntry) []any {
	if len(apiObjects) == 0 {
		return nil
	}

	var tfList []any

	for _, networkAclEntry := range apiObjects {
		tfList = append(tfList, flattenNetworkACLEntry(networkAclEntry))
	}

	return tfList
}

func flattenNetworkACLEntry(apiObject awstypes.NetworkAclEntry) any {
	tfMap := map[string]any{}

	if v := apiObject.CidrBlock; v != nil {
		tfMap[names.AttrCIDRBlock] = aws.ToString(v)
	}

	if v := apiObject.Egress; v != nil {
		tfMap["egress"] = aws.ToBool(v)
	}

	if v := apiObject.IcmpTypeCode; v != nil {
		tfMap["icmp_type_code"] = []any{map[string]any{
			"code":         aws.ToInt32(v.Code),
			names.AttrType: aws.ToInt32(v.Type),
		}}
	}

	if v := apiObject.Ipv6CidrBlock; v != nil {
		tfMap["ipv6_cidr_block"] = v
	}

	if v := apiObject.PortRange; v != nil {
		tfMap["port_range"] = []any{map[string]any{
			"from": aws.ToInt32(v.From),
			"to":   aws.ToInt32(v.To),
		}}
	}

	if v := apiObject.Protocol; v != nil {
		tfMap[names.AttrProtocol] = aws.ToString(v)
	}

	tfMap["rule_action"] = apiObject.RuleAction

	return tfMap
}

func flattenNetworkFirewallPolicy(apiObject *awstypes.NetworkFirewallPolicy) []any {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]any{}

	tfMap["firewall_deployment_model"] = apiObject.FirewallDeploymentModel

	return []any{tfMap}
}

func flattenThirdPartyFirewallPolicy(apiObject *awstypes.ThirdPartyFirewallPolicy) []any {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]any{}

	tfMap["firewall_deployment_model"] = apiObject.FirewallDeploymentModel

	return []any{tfMap}
}

func flattenResourceTags(apiObjects []awstypes.ResourceTag) map[string]any {
	tfMap := map[string]any{}

	for _, v := range apiObjects {
		tfMap[aws.ToString(v.Key)] = aws.ToString(v.Value)
	}

	return tfMap
}
