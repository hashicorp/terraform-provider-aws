// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package vpclattice

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/vpclattice"
	"github.com/aws/aws-sdk-go-v2/service/vpclattice/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	sdkid "github.com/hashicorp/terraform-plugin-sdk/v2/helper/id"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_vpclattice_listener_rule", name="Listener Rule")
// @Tags(identifierAttribute="arn")
// @Testing(tagsTest=false)
func resourceListenerRule() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceListenerRuleCreate,
		ReadWithoutTimeout:   resourceListenerRuleRead,
		UpdateWithoutTimeout: resourceListenerRuleUpdate,
		DeleteWithoutTimeout: resourceListenerRuleDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(30 * time.Minute),
			Update: schema.DefaultTimeout(30 * time.Minute),
			Delete: schema.DefaultTimeout(30 * time.Minute),
		},

		Schema: map[string]*schema.Schema{
			names.AttrAction: {
				Type:     schema.TypeList,
				MaxItems: 1,
				Required: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"fixed_response": {
							Type:     schema.TypeList,
							MaxItems: 1,
							Optional: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									names.AttrStatusCode: {
										Type:     schema.TypeInt,
										Required: true,
									},
								},
							},
						},
						"forward": {
							Type:     schema.TypeList,
							Optional: true,
							MaxItems: 1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"target_groups": {
										Type:     schema.TypeList,
										Required: true,
										MinItems: 1,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"target_group_identifier": {
													Type:     schema.TypeString,
													Required: true,
												},
												names.AttrWeight: {
													Type:         schema.TypeInt,
													ValidateFunc: validation.IntBetween(0, 999),
													Default:      100,
													Optional:     true,
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
			names.AttrARN: {
				Type:     schema.TypeString,
				Computed: true,
			},
			"listener_identifier": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"match": {
				Type:             schema.TypeList,
				Required:         true,
				MaxItems:         1,
				DiffSuppressFunc: verify.SuppressMissingOptionalConfigurationBlock,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"http_match": {
							Type:             schema.TypeList,
							Optional:         true,
							DiffSuppressFunc: verify.SuppressMissingOptionalConfigurationBlock,
							MaxItems:         1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"header_matches": {
										Type:             schema.TypeList,
										Optional:         true,
										DiffSuppressFunc: verify.SuppressMissingOptionalConfigurationBlock,
										MinItems:         1,
										MaxItems:         5,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"case_sensitive": {
													Type:     schema.TypeBool,
													Optional: true,
												},
												"match": {
													Type:             schema.TypeList,
													Required:         true,
													MaxItems:         1,
													DiffSuppressFunc: verify.SuppressMissingOptionalConfigurationBlock,
													Elem: &schema.Resource{
														Schema: map[string]*schema.Schema{
															"contains": {
																Type:     schema.TypeString,
																Optional: true,
															},
															"exact": {
																Type:     schema.TypeString,
																Optional: true,
															},
															names.AttrPrefix: {
																Type:     schema.TypeString,
																Optional: true,
															},
														},
													},
												},
												names.AttrName: {
													Type:     schema.TypeString,
													Required: true,
												},
											},
										},
									},
									"method": {
										Type:     schema.TypeString,
										Optional: true,
									},
									"path_match": {
										Type:     schema.TypeList,
										Optional: true,
										MaxItems: 1,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"case_sensitive": {
													Type:     schema.TypeBool,
													Optional: true,
												},
												"match": {
													Type:     schema.TypeList,
													Required: true,
													MaxItems: 1,
													Elem: &schema.Resource{
														Schema: map[string]*schema.Schema{
															"exact": {
																Type:     schema.TypeString,
																Optional: true,
															},
															names.AttrPrefix: {
																Type:     schema.TypeString,
																Optional: true,
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
					},
				},
			},
			names.AttrName: {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringLenBetween(3, 63),
			},
			names.AttrPriority: {
				Type:         schema.TypeInt,
				Required:     true,
				ValidateFunc: validation.IntBetween(1, 100),
			},
			"rule_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"service_identifier": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			names.AttrTags:    tftags.TagsSchema(),
			names.AttrTagsAll: tftags.TagsSchemaComputed(),
		},
	}
}

func resourceListenerRuleCreate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).VPCLatticeClient(ctx)

	name := d.Get(names.AttrName).(string)
	serviceID, listenerID := d.Get("service_identifier").(string), d.Get("listener_identifier").(string)
	input := vpclattice.CreateRuleInput{
		Action:             expandRuleAction(d.Get(names.AttrAction).([]any)[0].(map[string]any)),
		ClientToken:        aws.String(sdkid.UniqueId()),
		ListenerIdentifier: aws.String(listenerID),
		Match:              expandRuleMatch(d.Get("match").([]any)[0].(map[string]any)),
		Name:               aws.String(name),
		ServiceIdentifier:  aws.String(serviceID),
		Tags:               getTagsIn(ctx),
	}

	if v, ok := d.GetOk(names.AttrPriority); ok {
		input.Priority = aws.Int32(int32(v.(int)))
	}

	output, err := conn.CreateRule(ctx, &input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating VPCLattice Listener Rule (%s): %s", name, err)
	}

	d.SetId(listenerRuleCreateResourceID(serviceID, listenerID, aws.ToString(output.Id)))

	return append(diags, resourceListenerRuleRead(ctx, d, meta)...)
}

func resourceListenerRuleRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).VPCLatticeClient(ctx)

	serviceID, listenerID, ruleID, err := listenerRuleParseResourceID(d.Id())
	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	output, err := findListenerRuleByThreePartKey(ctx, conn, serviceID, listenerID, ruleID)

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] VPCLattice Listener Rule (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading VPCLattice Listener Rule (%s): %s", d.Id(), err)
	}

	if err := d.Set(names.AttrAction, []any{flattenRuleAction(output.Action)}); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting action: %s", err)
	}
	d.Set(names.AttrARN, output.Arn)
	d.Set("listener_identifier", listenerID)
	if err := d.Set("match", []any{flattenRuleMatch(output.Match)}); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting match: %s", err)
	}
	d.Set(names.AttrName, output.Name)
	d.Set(names.AttrPriority, output.Priority)
	d.Set("rule_id", output.Id)
	d.Set("service_identifier", serviceID)

	return diags
}

func resourceListenerRuleUpdate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).VPCLatticeClient(ctx)

	serviceID, listenerID, ruleID, err := listenerRuleParseResourceID(d.Id())
	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	if d.HasChangesExcept(names.AttrTags, names.AttrTagsAll) {
		input := vpclattice.UpdateRuleInput{
			ListenerIdentifier: aws.String(listenerID),
			RuleIdentifier:     aws.String(ruleID),
			ServiceIdentifier:  aws.String(serviceID),
		}

		if d.HasChange(names.AttrAction) {
			if v, ok := d.GetOk(names.AttrAction); ok && len(v.([]any)) > 0 && v.([]any)[0] != nil {
				input.Action = expandRuleAction(v.([]any)[0].(map[string]any))
			}
		}

		if d.HasChange("match") {
			if v, ok := d.GetOk("match"); ok && len(v.([]any)) > 0 && v.([]any)[0] != nil {
				input.Match = expandRuleMatch(v.([]any)[0].(map[string]any))
			}
		}
		_, err := conn.UpdateRule(ctx, &input)

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "updating VPCLattice Listener Rule (%s): %s", d.Id(), err)
		}
	}

	return append(diags, resourceListenerRuleRead(ctx, d, meta)...)
}

func resourceListenerRuleDelete(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).VPCLatticeClient(ctx)

	serviceID, listenerID, ruleID, err := listenerRuleParseResourceID(d.Id())
	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	log.Printf("[INFO] Deleting VPCLattice Listener Rule: %s", d.Id())
	input := vpclattice.DeleteRuleInput{
		ListenerIdentifier: aws.String(listenerID),
		RuleIdentifier:     aws.String(ruleID),
		ServiceIdentifier:  aws.String(serviceID),
	}
	_, err = conn.DeleteRule(ctx, &input)

	if errs.IsA[*types.ResourceNotFoundException](err) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting VPCLattice Listener Rule (%s): %s", d.Id(), err)
	}

	return diags
}

const listenerRuleResourceIDSeparator = "/"

func listenerRuleCreateResourceID(serviceID, listenerID, ruleID string) string {
	parts := []string{serviceID, listenerID, ruleID}
	id := strings.Join(parts, listenerRuleResourceIDSeparator)

	return id
}

func listenerRuleParseResourceID(id string) (string, string, string, error) {
	parts := strings.SplitN(id, listenerRuleResourceIDSeparator, 3)

	if len(parts) != 3 || parts[0] == "" || parts[1] == "" || parts[2] == "" {
		return "", "", "", fmt.Errorf("unexpected format for ID (%[1]s), expected SERVICE-ID%[2]sLISTENER-ID%[2]sRULE-ID", id, listenerRuleResourceIDSeparator)
	}

	return parts[0], parts[1], parts[2], nil
}

func findListenerRuleByThreePartKey(ctx context.Context, conn *vpclattice.Client, serviceID, listenerID, ruleID string) (*vpclattice.GetRuleOutput, error) {
	input := vpclattice.GetRuleInput{
		ListenerIdentifier: aws.String(listenerID),
		RuleIdentifier:     aws.String(ruleID),
		ServiceIdentifier:  aws.String(serviceID),
	}

	output, err := findListenerRule(ctx, conn, &input)

	if err != nil {
		return nil, err
	}

	if output.Id == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	return output, nil
}

func findListenerRule(ctx context.Context, conn *vpclattice.Client, input *vpclattice.GetRuleInput) (*vpclattice.GetRuleOutput, error) {
	output, err := conn.GetRule(ctx, input)

	if errs.IsA[*types.ResourceNotFoundException](err) {
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

func flattenRuleAction(apiObject types.RuleAction) map[string]any {
	if apiObject == nil {
		return nil
	}

	tfMap := make(map[string]any)

	switch v := apiObject.(type) {
	case *types.RuleActionMemberFixedResponse:
		tfMap["fixed_response"] = []any{flattenRuleActionMemberFixedResponse(v)}
	case *types.RuleActionMemberForward:
		tfMap["forward"] = []any{flattenForwardAction(v)}
	}

	return tfMap
}

func flattenRuleActionMemberFixedResponse(apiObject *types.RuleActionMemberFixedResponse) map[string]any {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]any{}

	if v := apiObject.Value.StatusCode; v != nil {
		tfMap[names.AttrStatusCode] = aws.ToInt32(v)
	}

	return tfMap
}

func flattenForwardAction(apiObject *types.RuleActionMemberForward) map[string]any {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]any{}

	if v := apiObject.Value.TargetGroups; v != nil {
		tfMap["target_groups"] = flattenWeightedTargetGroups(v)
	}

	return tfMap
}

func flattenWeightedTargetGroups(apiObjects []types.WeightedTargetGroup) []any {
	if len(apiObjects) == 0 {
		return nil
	}

	var tfList []any

	for _, apiObject := range apiObjects {
		tfList = append(tfList, flattenWeightedTargetGroup(&apiObject))
	}

	return tfList
}

func flattenWeightedTargetGroup(apiObject *types.WeightedTargetGroup) map[string]any {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]any{}

	if v := apiObject.TargetGroupIdentifier; v != nil {
		tfMap["target_group_identifier"] = aws.ToString(v)
	}

	if v := apiObject.Weight; v != nil {
		tfMap[names.AttrWeight] = aws.ToInt32(v)
	}

	return tfMap
}

func flattenRuleMatch(apiObject types.RuleMatch) map[string]any {
	if apiObject == nil {
		return nil
	}

	tfMap := make(map[string]any)

	if v, ok := apiObject.(*types.RuleMatchMemberHttpMatch); ok {
		tfMap["http_match"] = []any{flattenHTTPMatch(&v.Value)}
	}

	return tfMap
}

func flattenHTTPMatch(apiObject *types.HttpMatch) map[string]any {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]any{}

	if v := apiObject.HeaderMatches; v != nil {
		tfMap["header_matches"] = flattenHeaderMatches(v)
	}

	if v := apiObject.Method; v != nil {
		tfMap["method"] = aws.ToString(v)
	}

	if v := apiObject.PathMatch; v != nil {
		tfMap["path_match"] = flattenPathMatch(v)
	}

	return tfMap
}

func flattenHeaderMatches(apiObjects []types.HeaderMatch) []any {
	if len(apiObjects) == 0 {
		return nil
	}

	var tfList []any

	for _, apiObject := range apiObjects {
		tfList = append(tfList, flattenHeaderMatch(&apiObject))
	}

	return tfList
}

func flattenHeaderMatch(apiObject *types.HeaderMatch) map[string]any {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]any{}

	if v := apiObject.CaseSensitive; v != nil {
		tfMap["case_sensitive"] = aws.ToBool(v)
	}

	if v := apiObject.Match; v != nil {
		tfMap["match"] = []any{flattenHeaderMatchType(v)}
	}

	if v := apiObject.Name; v != nil {
		tfMap[names.AttrName] = aws.ToString(v)
	}

	return tfMap
}
func flattenHeaderMatchType(apiObject types.HeaderMatchType) map[string]any {
	if apiObject == nil {
		return nil
	}

	tfMap := make(map[string]any)

	switch v := apiObject.(type) {
	case *types.HeaderMatchTypeMemberContains:
		return flattenHeaderMatchTypeMemberContains(v)
	case *types.HeaderMatchTypeMemberExact:
		return flattenHeaderMatchTypeMemberExact(v)
	case *types.HeaderMatchTypeMemberPrefix:
		return flattenHeaderMatchTypeMemberPrefix(v)
	}

	return tfMap
}

func flattenHeaderMatchTypeMemberContains(apiObject *types.HeaderMatchTypeMemberContains) map[string]any {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]any{
		"contains": apiObject.Value,
	}

	return tfMap
}

func flattenHeaderMatchTypeMemberExact(apiObject *types.HeaderMatchTypeMemberExact) map[string]any {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]any{
		"exact": apiObject.Value,
	}

	return tfMap
}

func flattenHeaderMatchTypeMemberPrefix(apiObject *types.HeaderMatchTypeMemberPrefix) map[string]any {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]any{
		names.AttrPrefix: apiObject.Value,
	}

	return tfMap
}

func flattenPathMatch(apiObject *types.PathMatch) []any {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]any{}

	if v := apiObject.CaseSensitive; v != nil {
		tfMap["case_sensitive"] = aws.ToBool(v)
	}

	if v := apiObject.Match; v != nil {
		tfMap["match"] = []any{flattenPathMatchType(v)}
	}

	return []any{tfMap}
}

func flattenPathMatchType(apiObject types.PathMatchType) map[string]any {
	if apiObject == nil {
		return nil
	}

	tfMap := make(map[string]any)

	switch v := apiObject.(type) {
	case *types.PathMatchTypeMemberExact:
		return flattenPathMatchTypeMemberExact(v)
	case *types.PathMatchTypeMemberPrefix:
		return flattenPathMatchTypeMemberPrefix(v)
	}

	return tfMap
}

func flattenPathMatchTypeMemberExact(apiObject *types.PathMatchTypeMemberExact) map[string]any {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]any{
		"exact": apiObject.Value,
	}

	return tfMap
}

func flattenPathMatchTypeMemberPrefix(apiObject *types.PathMatchTypeMemberPrefix) map[string]any {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]any{
		names.AttrPrefix: apiObject.Value,
	}

	return tfMap
}

func expandRuleAction(tfMap map[string]any) types.RuleAction {
	var apiObject types.RuleAction

	if v, ok := tfMap["fixed_response"].([]any); ok && len(v) > 0 && v[0] != nil {
		apiObject = expandFixedResponseAction(v[0].(map[string]any))
	} else if v, ok := tfMap["forward"].([]any); ok && len(v) > 0 && v[0] != nil {
		apiObject = expandForwardAction(v[0].(map[string]any))
	}

	return apiObject
}

func expandFixedResponseAction(tfMap map[string]any) *types.RuleActionMemberFixedResponse {
	apiObject := &types.RuleActionMemberFixedResponse{}

	if v, ok := tfMap[names.AttrStatusCode].(int); ok && v != 0 {
		apiObject.Value.StatusCode = aws.Int32(int32(v))
	}

	return apiObject
}

func expandForwardAction(tfMap map[string]any) *types.RuleActionMemberForward {
	apiObject := &types.RuleActionMemberForward{}

	if v, ok := tfMap["target_groups"].([]any); ok && len(v) > 0 && v != nil {
		apiObject.Value.TargetGroups = expandWeightedTargetGroups(v)
	}

	return apiObject
}

func expandWeightedTargetGroups(tfList []any) []types.WeightedTargetGroup {
	if len(tfList) == 0 {
		return nil
	}

	var apiObjects []types.WeightedTargetGroup

	for _, tfMapRaw := range tfList {
		tfMap, ok := tfMapRaw.(map[string]any)
		if !ok {
			continue
		}

		apiObjects = append(apiObjects, expandWeightedTargetGroup(tfMap))
	}

	return apiObjects
}

func expandWeightedTargetGroup(tfMap map[string]any) types.WeightedTargetGroup {
	apiObject := types.WeightedTargetGroup{}

	if v, ok := tfMap["target_group_identifier"].(string); ok && v != "" {
		apiObject.TargetGroupIdentifier = aws.String(v)
	}

	if v, ok := tfMap[names.AttrWeight].(int); ok && v != 0 {
		apiObject.Weight = aws.Int32(int32(v))
	}

	return apiObject
}

func expandRuleMatch(tfMap map[string]any) types.RuleMatch {
	apiObject := &types.RuleMatchMemberHttpMatch{}

	if v, ok := tfMap["http_match"].([]any); ok && len(v) > 0 && v[0] != nil {
		apiObject.Value = expandHTTPMatch(v[0].(map[string]any))
	}

	return apiObject
}

func expandHTTPMatch(tfMap map[string]any) types.HttpMatch {
	apiObject := types.HttpMatch{}

	if v, ok := tfMap["header_matches"].([]any); ok && len(v) > 0 && v != nil {
		apiObject.HeaderMatches = expandHeaderMatches(v)
	}

	if v, ok := tfMap["method"].(string); ok {
		apiObject.Method = aws.String(v)
	}

	if v, ok := tfMap["path_match"].([]any); ok && len(v) > 0 && v != nil {
		apiObject.PathMatch = expandPathMatch(v[0].(map[string]any))
	}

	return apiObject
}

func expandHeaderMatches(tfList []any) []types.HeaderMatch {
	if len(tfList) == 0 {
		return nil
	}

	var apiObjects []types.HeaderMatch

	for _, tfMapRaw := range tfList {
		tfMap, ok := tfMapRaw.(map[string]any)
		if !ok {
			continue
		}

		apiObjects = append(apiObjects, expandHeaderMatch(tfMap))
	}

	return apiObjects
}

func expandHeaderMatch(tfMap map[string]any) types.HeaderMatch {
	apiObject := types.HeaderMatch{}

	if v, ok := tfMap["case_sensitive"].(bool); ok {
		apiObject.CaseSensitive = aws.Bool(v)
	}

	if v, ok := tfMap["match"].([]any); ok && len(v) > 0 {
		tfMap := v[0].(map[string]any)

		if v, ok := tfMap["exact"].(string); ok && v != "" {
			apiObject.Match = expandHeaderMatchTypeMemberExact(tfMap)
		} else if v, ok := tfMap[names.AttrPrefix].(string); ok && v != "" {
			apiObject.Match = expandHeaderMatchTypeMemberPrefix(tfMap)
		} else if v, ok := tfMap["contains"].(string); ok && v != "" {
			apiObject.Match = expandHeaderMatchTypeMemberContains(tfMap)
		}
	}

	if v, ok := tfMap[names.AttrName].(string); ok {
		apiObject.Name = aws.String(v)
	}

	return apiObject
}

func expandHeaderMatchTypeMemberContains(tfMap map[string]any) types.HeaderMatchType {
	apiObject := &types.HeaderMatchTypeMemberContains{}

	if v, ok := tfMap["contains"].(string); ok && v != "" {
		apiObject.Value = v
	}

	return apiObject
}

func expandHeaderMatchTypeMemberPrefix(tfMap map[string]any) types.HeaderMatchType {
	apiObject := &types.HeaderMatchTypeMemberPrefix{}

	if v, ok := tfMap[names.AttrPrefix].(string); ok && v != "" {
		apiObject.Value = v
	}

	return apiObject
}

func expandHeaderMatchTypeMemberExact(tfMap map[string]any) types.HeaderMatchType {
	apiObject := &types.HeaderMatchTypeMemberExact{}

	if v, ok := tfMap["exact"].(string); ok && v != "" {
		apiObject.Value = v
	}

	return apiObject
}

func expandPathMatch(tfMap map[string]any) *types.PathMatch {
	apiObject := &types.PathMatch{}

	if v, ok := tfMap["case_sensitive"].(bool); ok {
		apiObject.CaseSensitive = aws.Bool(v)
	}

	if v, ok := tfMap["match"].([]any); ok && len(v) > 0 {
		tfMap := v[0].(map[string]any)

		if v, ok := tfMap["exact"].(string); ok && v != "" {
			apiObject.Match = expandPathMatchTypeMemberExact(tfMap)
		} else if v, ok := tfMap[names.AttrPrefix].(string); ok && v != "" {
			apiObject.Match = expandPathMatchTypeMemberPrefix(tfMap)
		}
	}

	return apiObject
}

func expandPathMatchTypeMemberExact(tfMap map[string]any) types.PathMatchType {
	apiObject := &types.PathMatchTypeMemberExact{}

	if v, ok := tfMap["exact"].(string); ok && v != "" {
		apiObject.Value = v
	}

	return apiObject
}

func expandPathMatchTypeMemberPrefix(tfMap map[string]any) types.PathMatchType {
	apiObject := &types.PathMatchTypeMemberPrefix{}

	if v, ok := tfMap[names.AttrPrefix].(string); ok && v != "" {
		apiObject.Value = v
	}

	return apiObject
}
