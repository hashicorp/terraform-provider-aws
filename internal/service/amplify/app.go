// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package amplify

import (
	"context"
	"log"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/amplify"
	"github.com/aws/aws-sdk-go-v2/service/amplify/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/customdiff"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
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

// @SDKResource("aws_amplify_app", name="App")
// @Tags(identifierAttribute="arn")
// @Testing(existsType="github.com/aws/aws-sdk-go-v2/service/amplify/types;types.App", serialize=true, serializeDelay=true)
func resourceApp() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceAppCreate,
		ReadWithoutTimeout:   resourceAppRead,
		UpdateWithoutTimeout: resourceAppUpdate,
		DeleteWithoutTimeout: resourceAppDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		CustomizeDiff: customdiff.Sequence(
			verify.SetTagsDiff,
			customdiff.ForceNewIfChange(names.AttrDescription, func(_ context.Context, old, new, meta interface{}) bool {
				// Any existing value cannot be cleared.
				return new.(string) == ""
			}),
			customdiff.ForceNewIfChange("iam_service_role_arn", func(_ context.Context, old, new, meta interface{}) bool {
				// Any existing value cannot be cleared.
				return new.(string) == ""
			}),
		),

		Schema: map[string]*schema.Schema{
			"access_token": {
				Type:         schema.TypeString,
				Optional:     true,
				Sensitive:    true,
				ValidateFunc: validation.StringLenBetween(1, 255),
			},
			names.AttrARN: {
				Type:     schema.TypeString,
				Computed: true,
			},
			"auto_branch_creation_config": {
				Type:     schema.TypeList,
				Optional: true,
				Computed: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"basic_auth_credentials": {
							Type:         schema.TypeString,
							Optional:     true,
							Sensitive:    true,
							ValidateFunc: validation.StringLenBetween(1, 2000),
							DiffSuppressFunc: func(k, old, new string, d *schema.ResourceData) bool {
								// These credentials are ignored if basic auth is not enabled.
								if d.Get("auto_branch_creation_config.0.enable_basic_auth").(bool) {
									return old == new
								}

								return true
							},
						},
						"build_spec": {
							Type:         schema.TypeString,
							Optional:     true,
							ValidateFunc: validation.StringLenBetween(1, 25000),
						},
						"enable_auto_build": {
							Type:     schema.TypeBool,
							Optional: true,
						},
						"enable_basic_auth": {
							Type:     schema.TypeBool,
							Optional: true,
						},
						"enable_performance_mode": {
							Type:     schema.TypeBool,
							Optional: true,
							ForceNew: true,
						},
						"enable_pull_request_preview": {
							Type:     schema.TypeBool,
							Optional: true,
						},
						"environment_variables": {
							Type:     schema.TypeMap,
							Optional: true,
							Elem:     &schema.Schema{Type: schema.TypeString},
						},
						"framework": {
							Type:         schema.TypeString,
							Optional:     true,
							ValidateFunc: validation.StringLenBetween(1, 255),
						},
						"pull_request_environment_name": {
							Type:         schema.TypeString,
							Optional:     true,
							ValidateFunc: validation.StringLenBetween(1, 255),
						},
						names.AttrStage: {
							Type:             schema.TypeString,
							Optional:         true,
							ValidateDiagFunc: enum.Validate[types.Stage](),
							DiffSuppressFunc: func(k, old, new string, d *schema.ResourceData) bool {
								// API returns "NONE" by default.
								if old == stageNone && new == "" {
									return true
								}

								return old == new
							},
						},
					},
				},
			},
			"auto_branch_creation_patterns": {
				Type:     schema.TypeSet,
				Optional: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
				DiffSuppressFunc: func(k, old, new string, d *schema.ResourceData) bool {
					// These patterns are ignored if branch auto-creation is not enabled.
					if d.Get("enable_auto_branch_creation").(bool) {
						return old == new
					}

					return true
				},
			},
			"basic_auth_credentials": {
				Type:         schema.TypeString,
				Optional:     true,
				Sensitive:    true,
				ValidateFunc: validation.StringLenBetween(1, 2000),
				DiffSuppressFunc: func(k, old, new string, d *schema.ResourceData) bool {
					// These credentials are ignored if basic auth is not enabled.
					if d.Get("enable_basic_auth").(bool) {
						return old == new
					}

					return true
				},
			},
			"build_spec": {
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				ValidateFunc: validation.StringLenBetween(1, 25000),
			},
			"custom_headers": {
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				ValidateFunc: validation.StringLenBetween(1, 25000),
			},
			"custom_rule": {
				Type:     schema.TypeList,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						names.AttrCondition: {
							Type:         schema.TypeString,
							Optional:     true,
							ValidateFunc: validation.StringLenBetween(1, 2048),
						},
						names.AttrSource: {
							Type:         schema.TypeString,
							Required:     true,
							ValidateFunc: validation.StringLenBetween(1, 2048),
						},
						names.AttrStatus: {
							Type:     schema.TypeString,
							Optional: true,
							ValidateFunc: validation.StringInSlice([]string{
								"200",
								"301",
								"302",
								"404",
								"404-200",
							}, false),
						},
						names.AttrTarget: {
							Type:         schema.TypeString,
							Required:     true,
							ValidateFunc: validation.StringLenBetween(1, 2048),
						},
					},
				},
			},
			"default_domain": {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrDescription: {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringLenBetween(1, 1000),
			},
			"enable_auto_branch_creation": {
				Type:     schema.TypeBool,
				Optional: true,
			},
			"enable_basic_auth": {
				Type:     schema.TypeBool,
				Optional: true,
			},
			"enable_branch_auto_build": {
				Type:     schema.TypeBool,
				Optional: true,
			},
			"enable_branch_auto_deletion": {
				Type:     schema.TypeBool,
				Optional: true,
			},
			"environment_variables": {
				Type:     schema.TypeMap,
				Optional: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"iam_service_role_arn": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: verify.ValidARN,
			},
			names.AttrName: {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringLenBetween(1, 255),
			},
			"oauth_token": {
				Type:         schema.TypeString,
				Optional:     true,
				Sensitive:    true,
				ValidateFunc: validation.StringLenBetween(1, 1000),
			},
			"platform": {
				Type:             schema.TypeString,
				Optional:         true,
				Default:          types.PlatformWeb,
				ValidateDiagFunc: enum.Validate[types.Platform](),
			},
			"production_branch": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"branch_name": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"last_deploy_time": {
							Type:     schema.TypeString,
							Computed: true,
						},
						names.AttrStatus: {
							Type:     schema.TypeString,
							Computed: true,
						},
						"thumbnail_url": {
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
			},
			"repository": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringLenBetween(1, 1000),
			},
			names.AttrTags:    tftags.TagsSchema(),
			names.AttrTagsAll: tftags.TagsSchemaComputed(),
		},
	}
}

func resourceAppCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).AmplifyClient(ctx)

	name := d.Get(names.AttrName).(string)
	input := &amplify.CreateAppInput{
		Name: aws.String(name),
		Tags: getTagsIn(ctx),
	}

	if v, ok := d.GetOk("access_token"); ok {
		input.AccessToken = aws.String(v.(string))
	}

	if v, ok := d.GetOk("auto_branch_creation_config"); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
		input.AutoBranchCreationConfig = expandAutoBranchCreationConfig(v.([]interface{})[0].(map[string]interface{}))
	}

	if v, ok := d.GetOk("auto_branch_creation_patterns"); ok && v.(*schema.Set).Len() > 0 {
		input.AutoBranchCreationPatterns = flex.ExpandStringValueSet(v.(*schema.Set))
	}

	if v, ok := d.GetOk("basic_auth_credentials"); ok {
		input.BasicAuthCredentials = aws.String(v.(string))
	}

	if v, ok := d.GetOk("build_spec"); ok {
		input.BuildSpec = aws.String(v.(string))
	}

	if v, ok := d.GetOk("custom_headers"); ok {
		input.CustomHeaders = aws.String(v.(string))
	}

	if v, ok := d.GetOk("custom_rule"); ok && len(v.([]interface{})) > 0 {
		input.CustomRules = expandCustomRules(v.([]interface{}))
	}

	if v, ok := d.GetOk(names.AttrDescription); ok {
		input.Description = aws.String(v.(string))
	}

	if v, ok := d.GetOk("enable_auto_branch_creation"); ok {
		input.EnableAutoBranchCreation = aws.Bool(v.(bool))
	}

	if v, ok := d.GetOk("enable_basic_auth"); ok {
		input.EnableBasicAuth = aws.Bool(v.(bool))
	}

	if v, ok := d.GetOk("enable_branch_auto_build"); ok {
		input.EnableBranchAutoBuild = aws.Bool(v.(bool))
	}

	if v, ok := d.GetOk("enable_branch_auto_deletion"); ok {
		input.EnableBranchAutoDeletion = aws.Bool(v.(bool))
	}

	if v, ok := d.GetOk("environment_variables"); ok && len(v.(map[string]interface{})) > 0 {
		input.EnvironmentVariables = flex.ExpandStringValueMap(v.(map[string]interface{}))
	}

	if v, ok := d.GetOk("iam_service_role_arn"); ok {
		input.IamServiceRoleArn = aws.String(v.(string))
	}

	if v, ok := d.GetOk("oauth_token"); ok {
		input.OauthToken = aws.String(v.(string))
	}

	if v, ok := d.GetOk("platform"); ok {
		input.Platform = types.Platform(v.(string))
	}

	if v, ok := d.GetOk("repository"); ok {
		input.Repository = aws.String(v.(string))
	}

	output, err := conn.CreateApp(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating Amplify App (%s): %s", name, err)
	}

	d.SetId(aws.ToString(output.App.AppId))

	return append(diags, resourceAppRead(ctx, d, meta)...)
}

func resourceAppRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).AmplifyClient(ctx)

	app, err := findAppByID(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] Amplify App (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading Amplify App (%s): %s", d.Id(), err)
	}

	d.Set(names.AttrARN, app.AppArn)
	if app.AutoBranchCreationConfig != nil {
		if err := d.Set("auto_branch_creation_config", []interface{}{flattenAutoBranchCreationConfig(app.AutoBranchCreationConfig)}); err != nil {
			return sdkdiag.AppendErrorf(diags, "setting auto_branch_creation_config: %s", err)
		}
	} else {
		d.Set("auto_branch_creation_config", nil)
	}
	d.Set("auto_branch_creation_patterns", aws.StringSlice(app.AutoBranchCreationPatterns))
	d.Set("basic_auth_credentials", app.BasicAuthCredentials)
	d.Set("build_spec", app.BuildSpec)
	d.Set("custom_headers", app.CustomHeaders)
	if err := d.Set("custom_rule", flattenCustomRules(app.CustomRules)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting custom_rule: %s", err)
	}
	d.Set("default_domain", app.DefaultDomain)
	d.Set(names.AttrDescription, app.Description)
	d.Set("enable_auto_branch_creation", app.EnableAutoBranchCreation)
	d.Set("enable_basic_auth", app.EnableBasicAuth)
	d.Set("enable_branch_auto_build", app.EnableBranchAutoBuild)
	d.Set("enable_branch_auto_deletion", app.EnableBranchAutoDeletion)
	d.Set("environment_variables", aws.StringMap(app.EnvironmentVariables))
	d.Set("iam_service_role_arn", app.IamServiceRoleArn)
	d.Set(names.AttrName, app.Name)
	d.Set("platform", app.Platform)
	if app.ProductionBranch != nil {
		if err := d.Set("production_branch", []interface{}{flattenProductionBranch(app.ProductionBranch)}); err != nil {
			return sdkdiag.AppendErrorf(diags, "setting production_branch: %s", err)
		}
	} else {
		d.Set("production_branch", nil)
	}
	d.Set("repository", app.Repository)

	setTagsOut(ctx, app.Tags)

	return diags
}

func resourceAppUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).AmplifyClient(ctx)

	if d.HasChangesExcept(names.AttrTags, names.AttrTagsAll) {
		input := &amplify.UpdateAppInput{
			AppId: aws.String(d.Id()),
		}

		if d.HasChange("access_token") {
			input.AccessToken = aws.String(d.Get("access_token").(string))
		}

		if d.HasChange("auto_branch_creation_config") {
			input.AutoBranchCreationConfig = expandAutoBranchCreationConfig(d.Get("auto_branch_creation_config").([]interface{})[0].(map[string]interface{}))

			if d.HasChange("auto_branch_creation_config.0.environment_variables") {
				if v := d.Get("auto_branch_creation_config.0.environment_variables").(map[string]interface{}); len(v) == 0 {
					input.AutoBranchCreationConfig.EnvironmentVariables = map[string]string{"": ""}
				}
			}
		}

		if d.HasChange("auto_branch_creation_patterns") {
			input.AutoBranchCreationPatterns = flex.ExpandStringValueSet(d.Get("auto_branch_creation_patterns").(*schema.Set))
		}

		if d.HasChange("basic_auth_credentials") {
			input.BasicAuthCredentials = aws.String(d.Get("basic_auth_credentials").(string))
		}

		if d.HasChange("build_spec") {
			input.BuildSpec = aws.String(d.Get("build_spec").(string))
		}

		if d.HasChange("custom_headers") {
			input.CustomHeaders = aws.String(d.Get("custom_headers").(string))
		}

		if d.HasChange("custom_rule") {
			if v := d.Get("custom_rule").([]interface{}); len(v) > 0 {
				input.CustomRules = expandCustomRules(v)
			} else {
				input.CustomRules = []types.CustomRule{}
			}
		}

		if d.HasChange(names.AttrDescription) {
			input.Description = aws.String(d.Get(names.AttrDescription).(string))
		}

		if d.HasChange("enable_auto_branch_creation") {
			input.EnableAutoBranchCreation = aws.Bool(d.Get("enable_auto_branch_creation").(bool))
		}

		if d.HasChange("enable_basic_auth") {
			input.EnableBasicAuth = aws.Bool(d.Get("enable_basic_auth").(bool))
		}

		if d.HasChange("enable_branch_auto_build") {
			input.EnableBranchAutoBuild = aws.Bool(d.Get("enable_branch_auto_build").(bool))
		}

		if d.HasChange("enable_branch_auto_deletion") {
			input.EnableBranchAutoDeletion = aws.Bool(d.Get("enable_branch_auto_deletion").(bool))
		}

		if d.HasChange("environment_variables") {
			if v := d.Get("environment_variables").(map[string]interface{}); len(v) > 0 {
				input.EnvironmentVariables = flex.ExpandStringValueMap(v)
			} else {
				input.EnvironmentVariables = map[string]string{"": ""}
			}
		}

		if d.HasChange("iam_service_role_arn") {
			input.IamServiceRoleArn = aws.String(d.Get("iam_service_role_arn").(string))
		}

		if d.HasChange(names.AttrName) {
			input.Name = aws.String(d.Get(names.AttrName).(string))
		}

		if d.HasChange("oauth_token") {
			input.OauthToken = aws.String(d.Get("oauth_token").(string))
		}

		if d.HasChange("platform") {
			input.Platform = types.Platform(d.Get("platform").(string))
		}

		if d.HasChange("repository") {
			input.Repository = aws.String(d.Get("repository").(string))
		}

		_, err := conn.UpdateApp(ctx, input)

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "updating Amplify App (%s): %s", d.Id(), err)
		}
	}

	return append(diags, resourceAppRead(ctx, d, meta)...)
}

func resourceAppDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).AmplifyClient(ctx)

	log.Printf("[DEBUG] Deleting Amplify App: %s", d.Id())
	_, err := conn.DeleteApp(ctx, &amplify.DeleteAppInput{
		AppId: aws.String(d.Id()),
	})

	if errs.IsA[*types.NotFoundException](err) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting Amplify App (%s): %s", d.Id(), err)
	}

	return diags
}

func findAppByID(ctx context.Context, conn *amplify.Client, id string) (*types.App, error) {
	input := &amplify.GetAppInput{
		AppId: aws.String(id),
	}

	output, err := conn.GetApp(ctx, input)

	if errs.IsA[*types.NotFoundException](err) {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil || output.App == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	return output.App, nil
}

func expandAutoBranchCreationConfig(tfMap map[string]interface{}) *types.AutoBranchCreationConfig {
	if tfMap == nil {
		return nil
	}

	apiObject := &types.AutoBranchCreationConfig{}

	if v, ok := tfMap["basic_auth_credentials"].(string); ok && v != "" {
		apiObject.BasicAuthCredentials = aws.String(v)
	}

	if v, ok := tfMap["build_spec"].(string); ok && v != "" {
		apiObject.BuildSpec = aws.String(v)
	}

	if v, ok := tfMap["enable_auto_build"].(bool); ok {
		apiObject.EnableAutoBuild = aws.Bool(v)
	}

	if v, ok := tfMap["enable_basic_auth"].(bool); ok {
		apiObject.EnableBasicAuth = aws.Bool(v)
	}

	if v, ok := tfMap["enable_performance_mode"].(bool); ok {
		apiObject.EnablePerformanceMode = aws.Bool(v)
	}

	if v, ok := tfMap["enable_pull_request_preview"].(bool); ok {
		apiObject.EnablePullRequestPreview = aws.Bool(v)
	}

	if v, ok := tfMap["environment_variables"].(map[string]interface{}); ok && len(v) > 0 {
		apiObject.EnvironmentVariables = flex.ExpandStringValueMap(v)
	}

	if v, ok := tfMap["framework"].(string); ok && v != "" {
		apiObject.Framework = aws.String(v)
	}

	if v, ok := tfMap["pull_request_environment_name"].(string); ok && v != "" {
		apiObject.PullRequestEnvironmentName = aws.String(v)
	}

	if v, ok := tfMap[names.AttrStage].(string); ok && v != "" && v != stageNone {
		apiObject.Stage = types.Stage(v)
	}

	return apiObject
}

func flattenAutoBranchCreationConfig(apiObject *types.AutoBranchCreationConfig) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if v := apiObject.BasicAuthCredentials; v != nil {
		tfMap["basic_auth_credentials"] = aws.ToString(v)
	}

	if v := apiObject.BuildSpec; v != nil {
		tfMap["build_spec"] = aws.ToString(v)
	}

	if v := apiObject.EnableAutoBuild; v != nil {
		tfMap["enable_auto_build"] = aws.ToBool(v)
	}

	if v := apiObject.EnableBasicAuth; v != nil {
		tfMap["enable_basic_auth"] = aws.ToBool(v)
	}

	if v := apiObject.EnablePerformanceMode; v != nil {
		tfMap["enable_performance_mode"] = aws.ToBool(v)
	}

	if v := apiObject.EnablePullRequestPreview; v != nil {
		tfMap["enable_pull_request_preview"] = aws.ToBool(v)
	}

	if v := apiObject.EnvironmentVariables; v != nil {
		tfMap["environment_variables"] = v
	}

	if v := apiObject.Framework; v != nil {
		tfMap["framework"] = aws.ToString(v)
	}

	if v := apiObject.PullRequestEnvironmentName; v != nil {
		tfMap["pull_request_environment_name"] = aws.ToString(v)
	}

	tfMap[names.AttrStage] = apiObject.Stage

	return tfMap
}

func expandCustomRule(tfMap map[string]interface{}) *types.CustomRule {
	if tfMap == nil {
		return nil
	}

	apiObject := &types.CustomRule{}

	if v, ok := tfMap[names.AttrCondition].(string); ok && v != "" {
		apiObject.Condition = aws.String(v)
	}

	if v, ok := tfMap[names.AttrSource].(string); ok && v != "" {
		apiObject.Source = aws.String(v)
	}

	if v, ok := tfMap[names.AttrStatus].(string); ok && v != "" {
		apiObject.Status = aws.String(v)
	}

	if v, ok := tfMap[names.AttrTarget].(string); ok && v != "" {
		apiObject.Target = aws.String(v)
	}

	return apiObject
}

func expandCustomRules(tfList []interface{}) []types.CustomRule {
	if len(tfList) == 0 {
		return nil
	}

	var apiObjects []types.CustomRule

	for _, tfMapRaw := range tfList {
		tfMap, ok := tfMapRaw.(map[string]interface{})

		if !ok {
			continue
		}

		apiObject := expandCustomRule(tfMap)

		if apiObject == nil {
			continue
		}

		apiObjects = append(apiObjects, *apiObject)
	}

	return apiObjects
}

func flattenCustomRule(apiObject types.CustomRule) map[string]interface{} {
	tfMap := map[string]interface{}{}

	if v := apiObject.Condition; v != nil {
		tfMap[names.AttrCondition] = aws.ToString(v)
	}

	if v := apiObject.Source; v != nil {
		tfMap[names.AttrSource] = aws.ToString(v)
	}

	if v := apiObject.Status; v != nil {
		tfMap[names.AttrStatus] = aws.ToString(v)
	}

	if v := apiObject.Target; v != nil {
		tfMap[names.AttrTarget] = aws.ToString(v)
	}

	return tfMap
}

func flattenCustomRules(apiObjects []types.CustomRule) []interface{} {
	if len(apiObjects) == 0 {
		return nil
	}

	var tfList []interface{}

	for _, apiObject := range apiObjects {
		tfList = append(tfList, flattenCustomRule(apiObject))
	}

	return tfList
}

func flattenProductionBranch(apiObject *types.ProductionBranch) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if v := apiObject.BranchName; v != nil {
		tfMap["branch_name"] = aws.ToString(v)
	}

	if v := apiObject.LastDeployTime; v != nil {
		tfMap["last_deploy_time"] = aws.ToTime(v).Format(time.RFC3339)
	}

	if v := apiObject.Status; v != nil {
		tfMap[names.AttrStatus] = aws.ToString(v)
	}

	if v := apiObject.ThumbnailUrl; v != nil {
		tfMap["thumbnail_url"] = aws.ToString(v)
	}

	return tfMap
}
