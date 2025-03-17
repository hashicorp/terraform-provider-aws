// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package vpclattice

import (
	"context"
	"errors"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/vpclattice"
	"github.com/aws/aws-sdk-go-v2/service/vpclattice/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/id"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_vpclattice_listener_rule", name="Listener Rule")
// @Tags(identifierAttribute="arn")
// @Testing(tagsTest=false)
func ResourceListenerRule() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceListenerRuleCreate,
		ReadWithoutTimeout:   resourceListenerRuleRead,
		UpdateWithoutTimeout: resourceListenerRuleUpdate,
		DeleteWithoutTimeout: resourceListenerRuleDelete,

		Importer: &schema.ResourceImporter{
			StateContext: func(ctx context.Context, d *schema.ResourceData, meta any) ([]*schema.ResourceData, error) {
				idParts := strings.Split(d.Id(), "/")
				if len(idParts) != 3 || idParts[0] == "" || idParts[1] == "" || idParts[2] == "" {
					return nil, fmt.Errorf("unexpected format of ID (%q), expected SERVICE-ID/LISTENER-ID/RULE-ID", d.Id())
				}
				serviceIdentifier := idParts[0]
				listenerIdentifier := idParts[1]
				ruleId := idParts[2]
				d.Set("service_identifier", serviceIdentifier)
				d.Set("listener_identifier", listenerIdentifier)
				d.Set("rule_id", ruleId)

				return []*schema.ResourceData{d}, nil
			},
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
									"method": {
										Type:     schema.TypeString,
										Optional: true,
									},
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

const (
	ResNameListenerRule = "Listener Rule"
)

func resourceListenerRuleCreate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).VPCLatticeClient(ctx)

	name := d.Get(names.AttrName).(string)
	in := &vpclattice.CreateRuleInput{
		Action:             expandRuleAction(d.Get(names.AttrAction).([]any)[0].(map[string]any)),
		ClientToken:        aws.String(id.UniqueId()),
		ListenerIdentifier: aws.String(d.Get("listener_identifier").(string)),
		Match:              expandRuleMatch(d.Get("match").([]any)[0].(map[string]any)),
		Name:               aws.String(name),
		ServiceIdentifier:  aws.String(d.Get("service_identifier").(string)),
		Tags:               getTagsIn(ctx),
	}

	if v, ok := d.GetOk(names.AttrPriority); ok {
		in.Priority = aws.Int32(int32(v.(int)))
	}

	out, err := conn.CreateRule(ctx, in)
	if err != nil {
		return create.AppendDiagError(diags, names.VPCLattice, create.ErrActionCreating, ResNameListenerRule, name, err)
	}

	if out == nil || out.Arn == nil {
		return create.AppendDiagError(diags, names.VPCLattice, create.ErrActionCreating, ResNameListenerRule, d.Get(names.AttrName).(string), errors.New("empty output"))
	}

	d.Set("rule_id", out.Id)
	d.Set("service_identifier", in.ServiceIdentifier)
	d.Set("listener_identifier", in.ListenerIdentifier)

	parts := []string{
		d.Get("service_identifier").(string),
		d.Get("listener_identifier").(string),
		d.Get("rule_id").(string),
	}

	d.SetId(strings.Join(parts, "/"))

	return append(diags, resourceListenerRuleRead(ctx, d, meta)...)
}

func resourceListenerRuleRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).VPCLatticeClient(ctx)

	serviceId := d.Get("service_identifier").(string)
	listenerId := d.Get("listener_identifier").(string)
	ruleId := d.Get("rule_id").(string)

	out, err := FindListenerRuleByID(ctx, conn, serviceId, listenerId, ruleId)

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] VpcLattice Listener Rule (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return create.AppendDiagError(diags, names.VPCLattice, create.ErrActionReading, ResNameListenerRule, d.Id(), err)
	}

	d.Set(names.AttrARN, out.Arn)
	d.Set(names.AttrPriority, out.Priority)
	d.Set(names.AttrName, out.Name)
	d.Set("listener_identifier", listenerId)
	d.Set("service_identifier", serviceId)
	d.Set("rule_id", out.Id)

	if err := d.Set(names.AttrAction, []any{flattenRuleAction(out.Action)}); err != nil {
		return create.AppendDiagError(diags, names.VPCLattice, create.ErrActionSetting, ResNameListenerRule, d.Id(), err)
	}

	if err := d.Set("match", []any{flattenRuleMatch(out.Match)}); err != nil {
		return create.AppendDiagError(diags, names.VPCLattice, create.ErrActionSetting, ResNameListenerRule, d.Id(), err)
	}

	return diags
}

func resourceListenerRuleUpdate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).VPCLatticeClient(ctx)

	serviceId := d.Get("service_identifier").(string)
	listenerId := d.Get("listener_identifier").(string)
	ruleId := d.Get("rule_id").(string)

	if d.HasChangesExcept(names.AttrTags, names.AttrTagsAll) {
		in := &vpclattice.UpdateRuleInput{
			RuleIdentifier:     aws.String(ruleId),
			ListenerIdentifier: aws.String(listenerId),
			ServiceIdentifier:  aws.String(serviceId),
		}

		if d.HasChange(names.AttrAction) {
			if v, ok := d.GetOk(names.AttrAction); ok && len(v.([]any)) > 0 && v.([]any)[0] != nil {
				in.Action = expandRuleAction(v.([]any)[0].(map[string]any))
			}
		}

		if d.HasChange("match") {
			if v, ok := d.GetOk("match"); ok && len(v.([]any)) > 0 && v.([]any)[0] != nil {
				in.Match = expandRuleMatch(v.([]any)[0].(map[string]any))
			}
		}
		_, err := conn.UpdateRule(ctx, in)
		if err != nil {
			return create.AppendDiagError(diags, names.VPCLattice, create.ErrActionUpdating, ResNameListenerRule, d.Id(), err)
		}
	}

	return append(diags, resourceListenerRuleRead(ctx, d, meta)...)
}

func resourceListenerRuleDelete(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).VPCLatticeClient(ctx)

	serviceId := d.Get("service_identifier").(string)
	listenerId := d.Get("listener_identifier").(string)
	ruleId := d.Get("rule_id").(string)

	log.Printf("[INFO] Deleting VpcLattice Listening Rule: %s", d.Id())
	input := vpclattice.DeleteRuleInput{
		ListenerIdentifier: aws.String(listenerId),
		RuleIdentifier:     aws.String(ruleId),
		ServiceIdentifier:  aws.String(serviceId),
	}
	_, err := conn.DeleteRule(ctx, &input)

	if err != nil {
		var nfe *types.ResourceNotFoundException
		if errors.As(err, &nfe) {
			return diags
		}

		return create.AppendDiagError(diags, names.VPCLattice, create.ErrActionDeleting, ResNameListenerRule, d.Id(), err)
	}

	return diags
}

func FindListenerRuleByID(ctx context.Context, conn *vpclattice.Client, serviceIdentifier string, listenerIdentifier string, ruleId string) (*vpclattice.GetRuleOutput, error) {
	in := &vpclattice.GetRuleInput{
		ListenerIdentifier: aws.String(listenerIdentifier),
		RuleIdentifier:     aws.String(ruleId),
		ServiceIdentifier:  aws.String(serviceIdentifier),
	}
	out, err := conn.GetRule(ctx, in)
	if err != nil {
		var nfe *types.ResourceNotFoundException
		if errors.As(err, &nfe) {
			return nil, &retry.NotFoundError{
				LastError:   err,
				LastRequest: in,
			}
		}

		return nil, err
	}
	if out == nil || out.Id == nil {
		return nil, tfresource.NewEmptyResultError(in)
	}

	return out, nil
}

func flattenRuleAction(apiObject types.RuleAction) map[string]any {
	if apiObject == nil {
		return nil
	}

	tfMap := make(map[string]any)

	if v, ok := apiObject.(*types.RuleActionMemberFixedResponse); ok {
		tfMap["fixed_response"] = []any{flattenRuleActionMemberFixedResponse(v)}
	}
	if v, ok := apiObject.(*types.RuleActionMemberForward); ok {
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

	if v := apiObject.Method; v != nil {
		tfMap["method"] = aws.ToString(v)
	}

	if v := apiObject.HeaderMatches; v != nil {
		tfMap["header_matches"] = flattenHeaderMatches(v)
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

	if v := apiObject.Name; v != nil {
		tfMap[names.AttrName] = aws.ToString(v)
	}

	if v := apiObject.Match; v != nil {
		tfMap["match"] = []any{flattenHeaderMatchType(v)}
	}

	return tfMap
}
func flattenHeaderMatchType(apiObject types.HeaderMatchType) map[string]any {
	if apiObject == nil {
		return nil
	}

	tfMap := make(map[string]any)

	if v, ok := apiObject.(*types.HeaderMatchTypeMemberContains); ok {
		return flattenHeaderMatchTypeMemberContains(v)
	} else if v, ok := apiObject.(*types.HeaderMatchTypeMemberExact); ok {
		return flattenHeaderMatchTypeMemberExact(v)
	} else if v, ok := apiObject.(*types.HeaderMatchTypeMemberPrefix); ok {
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

	if v, ok := apiObject.(*types.PathMatchTypeMemberExact); ok {
		return flattenPathMatchTypeMemberExact(v)
	} else if v, ok := apiObject.(*types.PathMatchTypeMemberPrefix); ok {
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

		apiObject := expandWeightedTargetGroup(tfMap)

		apiObjects = append(apiObjects, apiObject)
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

		apiObject := expandHeaderMatch(tfMap)

		apiObjects = append(apiObjects, apiObject)
	}

	return apiObjects
}

func expandHeaderMatch(tfMap map[string]any) types.HeaderMatch {
	apiObject := types.HeaderMatch{}

	if v, ok := tfMap["case_sensitive"].(bool); ok {
		apiObject.CaseSensitive = aws.Bool(v)
	}

	if v, ok := tfMap[names.AttrName].(string); ok {
		apiObject.Name = aws.String(v)
	}

	if v, ok := tfMap["match"].([]any); ok && len(v) > 0 {
		matchObj := v[0].(map[string]any)
		if matchV, ok := matchObj["exact"].(string); ok && matchV != "" {
			apiObject.Match = expandHeaderMatchTypeMemberExact(matchObj)
		}
		if matchV, ok := matchObj[names.AttrPrefix].(string); ok && matchV != "" {
			apiObject.Match = expandHeaderMatchTypeMemberPrefix(matchObj)
		}
		if matchV, ok := matchObj["contains"].(string); ok && matchV != "" {
			apiObject.Match = expandHeaderMatchTypeMemberContains(matchObj)
		}
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
		matchObj := v[0].(map[string]any)
		if matchV, ok := matchObj["exact"].(string); ok && matchV != "" {
			apiObject.Match = expandPathMatchTypeMemberExact(matchObj)
		}
		if matchV, ok := matchObj[names.AttrPrefix].(string); ok && matchV != "" {
			apiObject.Match = expandPathMatchTypeMemberPrefix(matchObj)
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
