// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package codebuild

import (
	"context"
	"log"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/codebuild"
	"github.com/aws/aws-sdk-go-v2/service/codebuild/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_codebuild_webhook", name="Webhook")
func resourceWebhook() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceWebhookCreate,
		ReadWithoutTimeout:   resourceWebhookRead,
		DeleteWithoutTimeout: resourceWebhookDelete,
		UpdateWithoutTimeout: resourceWebhookUpdate,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			"branch_filter": {
				Type:          schema.TypeString,
				Optional:      true,
				ConflictsWith: []string{"filter_group"},
			},
			"build_type": {
				Type:             schema.TypeString,
				Optional:         true,
				ValidateDiagFunc: enum.Validate[types.WebhookBuildType](),
			},
			"filter_group": {
				Type:     schema.TypeSet,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						names.AttrFilter: {
							Type:     schema.TypeList,
							Optional: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"exclude_matched_pattern": {
										Type:     schema.TypeBool,
										Optional: true,
										Default:  false,
									},
									"pattern": {
										Type:     schema.TypeString,
										Required: true,
									},
									names.AttrType: {
										Type:             schema.TypeString,
										Required:         true,
										ValidateDiagFunc: enum.Validate[types.WebhookFilterType](),
									},
								},
							},
						},
					},
				},
				ConflictsWith: []string{"branch_filter"},
			},
			"payload_url": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"project_name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"scope_configuration": {
				Type:     schema.TypeList,
				Optional: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						names.AttrName: {
							Type:     schema.TypeString,
							Required: true,
						},
						names.AttrDomain: {
							Type:     schema.TypeString,
							Optional: true,
						},
						names.AttrScope: {
							Type:             schema.TypeString,
							Required:         true,
							ValidateDiagFunc: enum.Validate[types.WebhookScopeType](),
						},
					},
				},
			},
			"secret": {
				Type:      schema.TypeString,
				Computed:  true,
				Sensitive: true,
			},
			names.AttrURL: {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func resourceWebhookCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).CodeBuildClient(ctx)

	projectName := d.Get("project_name").(string)
	input := &codebuild.CreateWebhookInput{
		ProjectName: aws.String(projectName),
	}

	if v, ok := d.GetOk("branch_filter"); ok {
		input.BranchFilter = aws.String(v.(string))
	}

	if v, ok := d.GetOk("build_type"); ok {
		input.BuildType = types.WebhookBuildType(v.(string))
	}

	if v, ok := d.GetOk("filter_group"); ok && v.(*schema.Set).Len() > 0 {
		input.FilterGroups = expandWebhookFilterGroups(v.(*schema.Set).List())
	}

	if v, ok := d.GetOk("scope_configuration"); ok && len(v.([]interface{})) > 0 {
		input.ScopeConfiguration = expandScopeConfiguration(v.([]interface{}))
	}

	output, err := conn.CreateWebhook(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating CodeBuild Webhook (%s): %s", projectName, err)
	}

	d.SetId(projectName)
	// Secret is only returned on create.
	d.Set("secret", output.Webhook.Secret)

	return append(diags, resourceWebhookRead(ctx, d, meta)...)
}

func resourceWebhookRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).CodeBuildClient(ctx)

	webhook, err := findWebhookByProjectName(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] CodeBuild Webhook (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading CodeBuild Webhook (%s): %s", d.Id(), err)
	}

	d.Set("build_type", webhook.BuildType)
	d.Set("branch_filter", webhook.BranchFilter)
	if err := d.Set("filter_group", flattenWebhookFilterGroups(webhook.FilterGroups)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting filter_group: %s", err)
	}
	d.Set("payload_url", webhook.PayloadUrl)
	d.Set("project_name", d.Id())
	if err := d.Set("scope_configuration", flattenScopeConfiguration(webhook.ScopeConfiguration)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting scope_configuration: %s", err)
	}
	d.Set("secret", d.Get("secret").(string))
	d.Set(names.AttrURL, webhook.Url)

	return diags
}

func resourceWebhookUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).CodeBuildClient(ctx)

	input := &codebuild.UpdateWebhookInput{
		ProjectName: aws.String(d.Id()),
	}

	if v, ok := d.GetOk("build_type"); ok {
		input.BuildType = types.WebhookBuildType(v.(string))
	}

	var filterGroups [][]types.WebhookFilter
	if v, ok := d.GetOk("filter_group"); ok && v.(*schema.Set).Len() > 0 {
		filterGroups = expandWebhookFilterGroups(v.(*schema.Set).List())
	}
	if len(filterGroups) > 0 {
		input.FilterGroups = filterGroups
	} else {
		input.BranchFilter = aws.String(d.Get("branch_filter").(string))
	}

	_, err := conn.UpdateWebhook(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "updating CodeBuild Webhook (%s): %s", d.Id(), err)
	}

	return append(diags, resourceWebhookRead(ctx, d, meta)...)
}

func resourceWebhookDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).CodeBuildClient(ctx)

	log.Printf("[INFO] Deleting CodeBuild Webhook: %s", d.Id())
	_, err := conn.DeleteWebhook(ctx, &codebuild.DeleteWebhookInput{
		ProjectName: aws.String(d.Id()),
	})

	if errs.IsA[*types.ResourceNotFoundException](err) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting CodeBuild Webhook (%s): %s", d.Id(), err)
	}

	return diags
}

func findWebhookByProjectName(ctx context.Context, conn *codebuild.Client, name string) (*types.Webhook, error) {
	output, err := findProjectByNameOrARN(ctx, conn, name)

	if err != nil {
		return nil, err
	}

	if output.Webhook == nil {
		return nil, tfresource.NewEmptyResultError(name)
	}

	return output.Webhook, nil
}

func expandWebhookFilterGroups(tfList []interface{}) [][]types.WebhookFilter {
	if len(tfList) == 0 {
		return nil
	}

	var apiObjects [][]types.WebhookFilter

	for _, tfMapRaw := range tfList {
		tfMap, ok := tfMapRaw.(map[string]interface{})
		if !ok {
			continue
		}

		if v, ok := tfMap[names.AttrFilter].([]interface{}); ok && len(v) > 0 {
			apiObjects = append(apiObjects, expandWebhookFilters(v))
		}
	}

	return apiObjects
}

func expandWebhookFilters(tfList []interface{}) []types.WebhookFilter {
	if len(tfList) == 0 {
		return nil
	}

	var apiObjects []types.WebhookFilter

	for _, tfMapRaw := range tfList {
		tfMap, ok := tfMapRaw.(map[string]interface{})
		if !ok {
			continue
		}

		apiObject := expandWebhookFilter(tfMap)

		if apiObject == nil {
			continue
		}

		apiObjects = append(apiObjects, *apiObject)
	}

	return apiObjects
}

func expandWebhookFilter(tfMap map[string]interface{}) *types.WebhookFilter {
	if tfMap == nil {
		return nil
	}

	apiObject := &types.WebhookFilter{}

	if v, ok := tfMap["exclude_matched_pattern"].(bool); ok {
		apiObject.ExcludeMatchedPattern = aws.Bool(v)
	}

	if v, ok := tfMap["pattern"].(string); ok && v != "" {
		apiObject.Pattern = aws.String(v)
	}

	if v, ok := tfMap[names.AttrType].(string); ok && v != "" {
		apiObject.Type = types.WebhookFilterType(v)
	}

	return apiObject
}

func expandScopeConfiguration(tfList []interface{}) *types.ScopeConfiguration {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap := tfList[0].(map[string]interface{})

	apiObject := &types.ScopeConfiguration{
		Name:  aws.String(tfMap[names.AttrName].(string)),
		Scope: types.WebhookScopeType(tfMap[names.AttrScope].(string)),
	}

	if v, ok := tfMap[names.AttrDomain].(string); ok && v != "" {
		apiObject.Domain = aws.String(v)
	}

	return apiObject
}

func flattenWebhookFilterGroups(apiObjects [][]types.WebhookFilter) []interface{} {
	if len(apiObjects) == 0 {
		return nil
	}

	var tfList []interface{}

	for _, apiObject := range apiObjects {
		tfMap := map[string]interface{}{
			names.AttrFilter: flattenWebhookFilters(apiObject),
		}
		tfList = append(tfList, tfMap)
	}

	return tfList
}

func flattenWebhookFilters(apiObjects []types.WebhookFilter) []interface{} {
	if len(apiObjects) == 0 {
		return nil
	}

	var tfList []interface{}

	for _, apiObject := range apiObjects {
		tfList = append(tfList, flattenWebhookFilter(apiObject))
	}

	return tfList
}

func flattenWebhookFilter(apiObject types.WebhookFilter) map[string]interface{} {
	tfMap := map[string]interface{}{
		names.AttrType: apiObject.Type,
	}

	if v := apiObject.ExcludeMatchedPattern; v != nil {
		tfMap["exclude_matched_pattern"] = aws.ToBool(v)
	}

	if v := apiObject.Pattern; v != nil {
		tfMap["pattern"] = aws.ToString(v)
	}

	return tfMap
}

func flattenScopeConfiguration(apiObject *types.ScopeConfiguration) []interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{
		names.AttrName:  apiObject.Name,
		names.AttrScope: apiObject.Scope,
	}

	if apiObject.Domain != nil {
		tfMap[names.AttrDomain] = apiObject.Domain
	}

	return []interface{}{tfMap}
}
