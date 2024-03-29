// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package fms

import (
	"context"
	"log"

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
			"arn": {
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
			"description": {
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
			"name": {
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
			"resource_tags": tftags.TagsSchema(),
			"resource_type": {
				Type:          schema.TypeString,
				Optional:      true,
				Computed:      true,
				ValidateFunc:  validation.StringMatch(regexache.MustCompile(`^([\p{L}\p{Z}\p{N}_.:/=+\-@]*)$`), "must match a supported resource type, such as AWS::EC2::VPC, see also: https://docs.aws.amazon.com/fms/2018-01-01/APIReference/API_Policy.html"),
				ConflictsWith: []string{"resource_type_list"},
			},
			"resource_type_list": {
				Type:     schema.TypeSet,
				Optional: true,
				Computed: true,
				Elem: &schema.Schema{
					Type:         schema.TypeString,
					ValidateFunc: validation.StringMatch(regexache.MustCompile(`^([\p{L}\p{Z}\p{N}_.:/=+\-@]*)$`), "must match a supported resource type, such as AWS::EC2::VPC, see also: https://docs.aws.amazon.com/fms/2018-01-01/APIReference/API_Policy.html"),
				},
				ConflictsWith: []string{"resource_type"},
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
						"type": {
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
		Policy:  resourcePolicyExpandPolicy(d),
		TagList: getTagsIn(ctx),
	}

	output, err := conn.PutPolicy(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating FMS Policy: %s", err)
	}

	d.SetId(aws.ToString(output.Policy.PolicyId))

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
	d.Set("arn", arn)
	policy := output.Policy
	d.Set("delete_unused_fm_managed_resources", policy.DeleteUnusedFMManagedResources)
	d.Set("description", policy.PolicyDescription)
	if err := d.Set("exclude_map", flattenPolicyMap(policy.ExcludeMap)); err != nil {
		sdkdiag.AppendErrorf(diags, "setting exclude_map: %s", err)
	}
	d.Set("exclude_resource_tags", policy.ExcludeResourceTags)
	if err := d.Set("include_map", flattenPolicyMap(policy.IncludeMap)); err != nil {
		sdkdiag.AppendErrorf(diags, "setting include_map: %s", err)
	}
	d.Set("name", policy.PolicyName)
	d.Set("policy_update_token", policy.PolicyUpdateToken)
	d.Set("remediation_enabled", policy.RemediationEnabled)
	if err := d.Set("resource_tags", flattenResourceTags(policy.ResourceTags)); err != nil {
		sdkdiag.AppendErrorf(diags, "setting resource_tags: %s", err)
	}
	d.Set("resource_type", policy.ResourceType)
	if err := d.Set("resource_type_list", policy.ResourceTypeList); err != nil {
		sdkdiag.AppendErrorf(diags, "setting resource_type_list: %s", err)
	}
	securityServicePolicy := []map[string]interface{}{{
		"type":                 string(policy.SecurityServicePolicyData.Type),
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

	if d.HasChangesExcept("tags", "tags_all") {
		input := &fms.PutPolicyInput{
			Policy: resourcePolicyExpandPolicy(d),
		}

		_, err := conn.PutPolicy(ctx, input)

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
		DeleteAllPolicyResources: aws.ToBool(d.Get("delete_all_policy_resources").(*bool)),
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

func resourcePolicyExpandPolicy(d *schema.ResourceData) *awstypes.Policy {
	resourceType := aws.String("ResourceTypeList")
	resourceTypeList := flex.ExpandStringSet(d.Get("resource_type_list").(*schema.Set))
	if t, ok := d.GetOk("resource_type"); ok {
		resourceType = aws.String(t.(string))
	}

	fmsPolicy := &awstypes.Policy{
		DeleteUnusedFMManagedResources: aws.ToBool(d.Get("delete_unused_fm_managed_resources").(*bool)),
		ExcludeResourceTags:            aws.ToBool(d.Get("exclude_resource_tags").(*bool)),
		PolicyDescription:              aws.String(d.Get("description").(string)),
		PolicyName:                     aws.String(d.Get("name").(string)),
		RemediationEnabled:             aws.ToBool(d.Get("remediation_enabled").(*bool)),
		ResourceType:                   resourceType,
		ResourceTypeList:               aws.ToStringSlice(resourceTypeList),
	}

	if d.Id() != "" {
		fmsPolicy.PolicyId = aws.String(d.Id())
		fmsPolicy.PolicyUpdateToken = aws.String(d.Get("policy_update_token").(string))
	}

	fmsPolicy.ExcludeMap = expandPolicyMap(d.Get("exclude_map").([]interface{}))

	fmsPolicy.IncludeMap = expandPolicyMap(d.Get("include_map").([]interface{}))

	fmsPolicy.ResourceTags = constructResourceTags(d.Get("resource_tags"))

	securityServicePolicy := d.Get("security_service_policy_data").([]interface{})[0].(map[string]interface{})
	fmsPolicy.SecurityServicePolicyData = &awstypes.SecurityServicePolicyData{
		ManagedServiceData: aws.String(securityServicePolicy["managed_service_data"].(string)),
		Type:               awstypes.SecurityServiceType(securityServicePolicy["type"].(string)),
	}

	if v, ok := securityServicePolicy["policy_option"].([]interface{}); ok && len(v) > 0 && v[0] != nil {
		fmsPolicy.SecurityServicePolicyData.PolicyOption = expandPolicyOption(v[0].(map[string]interface{}))
	}

	return fmsPolicy
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

func expandPolicyMap(set []interface{}) map[string][]string {
	fmsPolicyMap := map[string][]string{}
	if len(set) > 0 {
		if _, ok := set[0].(map[string]interface{}); !ok {
			return fmsPolicyMap
		}
		for key, listValue := range set[0].(map[string]interface{}) {
			var flatKey string
			switch key {
			case "account":
				flatKey = "ACCOUNT"
			case "orgunit":
				flatKey = "ORG_UNIT"
			}

			for _, value := range listValue.(*schema.Set).List() {
				fmsPolicyMap[flatKey] = append(fmsPolicyMap[flatKey], aws.ToString(value.(*string)))
			}
		}
	}
	return fmsPolicyMap
}

func flattenPolicyMap(fmsPolicyMap map[string][]string) []interface{} {
	flatPolicyMap := map[string]interface{}{}

	for key, value := range fmsPolicyMap {
		switch key {
		case "ACCOUNT":
			flatPolicyMap["account"] = value
		case "ORG_UNIT":
			flatPolicyMap["orgunit"] = value
		default:
			log.Printf("[WARNING] Unexpected key (%q) found in FMS policy", key)
		}
	}

	return []interface{}{flatPolicyMap}
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

func flattenResourceTags(resourceTags []awstypes.ResourceTag) map[string]interface{} {
	resTags := map[string]interface{}{}

	for _, v := range resourceTags {
		resTags[*v.Key] = v.Value
	}
	return resTags
}

func constructResourceTags(rTags interface{}) []awstypes.ResourceTag {
	var rTagList []awstypes.ResourceTag

	tags := rTags.(map[string]interface{})
	for k, v := range tags {
		rTagList = append(rTagList, awstypes.ResourceTag{Key: aws.String(k), Value: aws.String(v.(string))})
	}

	return rTagList
}
