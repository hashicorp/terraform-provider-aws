// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package fms

import (
	"context"
	"log"
	"regexp"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/fms"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

const (
	ResNamePolicy = "Policy"
)

// @SDKResource("aws_fms_policy", name="Policy")
// @Tags(identifierAttribute="arn")
func ResourcePolicy() *schema.Resource {
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
				ValidateFunc:  validation.StringMatch(regexp.MustCompile(`^([\p{L}\p{Z}\p{N}_.:/=+\-@]*)$`), "must match a supported resource type, such as AWS::EC2::VPC, see also: https://docs.aws.amazon.com/fms/2018-01-01/APIReference/API_Policy.html"),
				ConflictsWith: []string{"resource_type_list"},
			},
			"resource_type_list": {
				Type:     schema.TypeSet,
				Optional: true,
				Computed: true,
				Elem: &schema.Schema{
					Type:         schema.TypeString,
					ValidateFunc: validation.StringMatch(regexp.MustCompile(`^([\p{L}\p{Z}\p{N}_.:/=+\-@]*)$`), "must match a supported resource type, such as AWS::EC2::VPC, see also: https://docs.aws.amazon.com/fms/2018-01-01/APIReference/API_Policy.html"),
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
							Type:             schema.TypeString,
							Optional:         true,
							DiffSuppressFunc: verify.SuppressEquivalentJSONDiffs,
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
													Type:         schema.TypeString,
													Optional:     true,
													ValidateFunc: validation.StringInSlice(fms.FirewallDeploymentModel_Values(), false),
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
													Type:         schema.TypeString,
													Optional:     true,
													ValidateFunc: validation.StringInSlice(fms.FirewallDeploymentModel_Values(), false),
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
	conn := meta.(*conns.AWSClient).FMSConn(ctx)

	input := &fms.PutPolicyInput{
		Policy:  resourcePolicyExpandPolicy(d),
		TagList: getTagsIn(ctx),
	}

	output, err := conn.PutPolicyWithContext(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating FMS Policy: %s", err)
	}

	d.SetId(aws.StringValue(output.Policy.PolicyId))

	return append(diags, resourcePolicyRead(ctx, d, meta)...)
}

func resourcePolicyRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).FMSConn(ctx)

	output, err := FindPolicyByID(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] FMS Policy %s not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading FMS Policy (%s): %s", d.Id(), err)
	}

	arn := aws.StringValue(output.PolicyArn)
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
		"type":                 aws.StringValue(policy.SecurityServicePolicyData.Type),
		"managed_service_data": aws.StringValue(policy.SecurityServicePolicyData.ManagedServiceData),
		"policy_option":        flattenPolicyOption(policy.SecurityServicePolicyData.PolicyOption),
	}}
	if err := d.Set("security_service_policy_data", securityServicePolicy); err != nil {
		sdkdiag.AppendErrorf(diags, "setting security_service_policy_data: %s", err)
	}

	return diags
}

func resourcePolicyUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).FMSConn(ctx)

	if d.HasChangesExcept("tags", "tags_all") {
		input := &fms.PutPolicyInput{
			Policy: resourcePolicyExpandPolicy(d),
		}

		_, err := conn.PutPolicyWithContext(ctx, input)

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "updating FMS Policy (%s): %s", d.Id(), err)
		}
	}

	return append(diags, resourcePolicyRead(ctx, d, meta)...)
}

func resourcePolicyDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).FMSConn(ctx)

	log.Printf("[DEBUG] Deleting FMS Policy: %s", d.Id())
	_, err := conn.DeletePolicyWithContext(ctx, &fms.DeletePolicyInput{
		PolicyId:                 aws.String(d.Id()),
		DeleteAllPolicyResources: aws.Bool(d.Get("delete_all_policy_resources").(bool)),
	})

	if tfawserr.ErrCodeEquals(err, fms.ErrCodeResourceNotFoundException) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting FMS Policy (%s): %s", d.Id(), err)
	}

	return diags
}

func FindPolicyByID(ctx context.Context, conn *fms.FMS, id string) (*fms.GetPolicyOutput, error) {
	input := &fms.GetPolicyInput{
		PolicyId: aws.String(id),
	}

	output, err := conn.GetPolicyWithContext(ctx, input)

	if tfawserr.ErrCodeEquals(err, fms.ErrCodeResourceNotFoundException) {
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

func resourcePolicyExpandPolicy(d *schema.ResourceData) *fms.Policy {
	resourceType := aws.String("ResourceTypeList")
	resourceTypeList := flex.ExpandStringSet(d.Get("resource_type_list").(*schema.Set))
	if t, ok := d.GetOk("resource_type"); ok {
		resourceType = aws.String(t.(string))
	}

	fmsPolicy := &fms.Policy{
		DeleteUnusedFMManagedResources: aws.Bool(d.Get("delete_unused_fm_managed_resources").(bool)),
		ExcludeResourceTags:            aws.Bool(d.Get("exclude_resource_tags").(bool)),
		PolicyDescription:              aws.String(d.Get("description").(string)),
		PolicyName:                     aws.String(d.Get("name").(string)),
		RemediationEnabled:             aws.Bool(d.Get("remediation_enabled").(bool)),
		ResourceType:                   resourceType,
		ResourceTypeList:               resourceTypeList,
	}

	if d.Id() != "" {
		fmsPolicy.PolicyId = aws.String(d.Id())
		fmsPolicy.PolicyUpdateToken = aws.String(d.Get("policy_update_token").(string))
	}

	fmsPolicy.ExcludeMap = expandPolicyMap(d.Get("exclude_map").([]interface{}))

	fmsPolicy.IncludeMap = expandPolicyMap(d.Get("include_map").([]interface{}))

	fmsPolicy.ResourceTags = constructResourceTags(d.Get("resource_tags"))

	securityServicePolicy := d.Get("security_service_policy_data").([]interface{})[0].(map[string]interface{})
	fmsPolicy.SecurityServicePolicyData = &fms.SecurityServicePolicyData{
		ManagedServiceData: aws.String(securityServicePolicy["managed_service_data"].(string)),
		Type:               aws.String(securityServicePolicy["type"].(string)),
	}

	if v, ok := securityServicePolicy["policy_option"].([]interface{}); ok && len(v) > 0 && v[0] != nil {
		fmsPolicy.SecurityServicePolicyData.PolicyOption = expandPolicyOption(v[0].(map[string]interface{}))
	}

	return fmsPolicy
}

func expandPolicyOption(tfMap map[string]interface{}) *fms.PolicyOption {
	if tfMap == nil {
		return nil
	}

	apiObject := &fms.PolicyOption{}

	if v, ok := tfMap["network_firewall_policy"].([]interface{}); ok && len(v) > 0 && v[0] != nil {
		apiObject.NetworkFirewallPolicy = expandPolicyOptionNetworkFirewall(v[0].(map[string]interface{}))
	}

	if v, ok := tfMap["third_party_firewall_policy"].([]interface{}); ok && len(v) > 0 && v[0] != nil {
		apiObject.ThirdPartyFirewallPolicy = expandPolicyOptionThirdPartyFirewall(v[0].(map[string]interface{}))
	}

	return apiObject
}

func expandPolicyOptionNetworkFirewall(tfMap map[string]interface{}) *fms.NetworkFirewallPolicy {
	if tfMap == nil {
		return nil
	}

	apiObject := &fms.NetworkFirewallPolicy{}

	if v, ok := tfMap["firewall_deployment_model"].(string); ok {
		apiObject.FirewallDeploymentModel = aws.String(v)
	}

	return apiObject
}

func expandPolicyOptionThirdPartyFirewall(tfMap map[string]interface{}) *fms.ThirdPartyFirewallPolicy {
	if tfMap == nil {
		return nil
	}

	apiObject := &fms.ThirdPartyFirewallPolicy{}

	if v, ok := tfMap["firewall_deployment_model"].(string); ok {
		apiObject.FirewallDeploymentModel = aws.String(v)
	}

	return apiObject
}

func expandPolicyMap(set []interface{}) map[string][]*string {
	fmsPolicyMap := map[string][]*string{}
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
				fmsPolicyMap[flatKey] = append(fmsPolicyMap[flatKey], aws.String(value.(string)))
			}
		}
	}
	return fmsPolicyMap
}

func flattenPolicyMap(fmsPolicyMap map[string][]*string) []interface{} {
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

func flattenPolicyOption(fmsPolicyOption *fms.PolicyOption) []interface{} {
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

func flattenPolicyOptionNetworkFirewall(fmsNetworkFirewallPolicy *fms.NetworkFirewallPolicy) []interface{} {
	if fmsNetworkFirewallPolicy == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if v := fmsNetworkFirewallPolicy.FirewallDeploymentModel; v != nil {
		tfMap["firewall_deployment_model"] = aws.StringValue(v)
	}

	return []interface{}{tfMap}
}

func flattenPolicyOptionThirdPartyFirewall(fmsThirdPartyFirewallPolicy *fms.ThirdPartyFirewallPolicy) []interface{} {
	if fmsThirdPartyFirewallPolicy == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if v := fmsThirdPartyFirewallPolicy.FirewallDeploymentModel; v != nil {
		tfMap["firewall_deployment_model"] = aws.StringValue(v)
	}

	return []interface{}{tfMap}
}

func flattenResourceTags(resourceTags []*fms.ResourceTag) map[string]interface{} {
	resTags := map[string]interface{}{}

	for _, v := range resourceTags {
		resTags[*v.Key] = v.Value
	}
	return resTags
}

func constructResourceTags(rTags interface{}) []*fms.ResourceTag {
	var rTagList []*fms.ResourceTag

	tags := rTags.(map[string]interface{})
	for k, v := range tags {
		rTagList = append(rTagList, &fms.ResourceTag{Key: aws.String(k), Value: aws.String(v.(string))})
	}

	return rTagList
}
