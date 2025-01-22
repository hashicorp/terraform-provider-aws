// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package codepipeline

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"log"
	"strings"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/codepipeline"
	"github.com/aws/aws-sdk-go-v2/service/codepipeline/types"
	"github.com/hashicorp/go-cty/cty"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
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

const (
	providerGitHub                      = "GitHub"
	gitHubActionConfigurationOAuthToken = "OAuthToken"
)

// @SDKResource("aws_codepipeline", name="Pipeline")
// @Tags(identifierAttribute="arn")
func resourcePipeline() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourcePipelineCreate,
		ReadWithoutTimeout:   resourcePipelineRead,
		UpdateWithoutTimeout: resourcePipelineUpdate,
		DeleteWithoutTimeout: resourcePipelineDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"artifact_store": {
				Type:     schema.TypeSet,
				Required: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"encryption_key": {
							Type:     schema.TypeList,
							MaxItems: 1,
							Optional: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"id": {
										Type:     schema.TypeString,
										Required: true,
									},
									"type": {
										Type:             schema.TypeString,
										Required:         true,
										ValidateDiagFunc: enum.Validate[types.EncryptionKeyType](),
									},
								},
							},
						},
						"location": {
							Type:     schema.TypeString,
							Required: true,
						},
						"region": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"type": {
							Type:             schema.TypeString,
							Required:         true,
							ValidateDiagFunc: enum.Validate[types.ArtifactStoreType](),
						},
					},
				},
			},
			"execution_mode": {
				Type:             schema.TypeString,
				Optional:         true,
				Default:          types.ExecutionModeSuperseded,
				ValidateDiagFunc: enum.Validate[types.ExecutionMode](),
			},
			"name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
				ValidateFunc: validation.All(
					validation.StringLenBetween(1, 100),
					validation.StringMatch(regexache.MustCompile(`[0-9A-Za-z_.@-]+`), ""),
				),
			},
			"pipeline_type": {
				Type:             schema.TypeString,
				Optional:         true,
				Default:          types.PipelineTypeV1,
				ValidateDiagFunc: enum.Validate[types.PipelineType](),
			},
			"role_arn": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: verify.ValidARN,
			},
			"stage": {
				Type:     schema.TypeList,
				MinItems: 2,
				Required: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"action": {
							Type:     schema.TypeList,
							Required: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"category": {
										Type:             schema.TypeString,
										Required:         true,
										ValidateDiagFunc: enum.Validate[types.ActionCategory](),
									},
									"commands": {
										Type:     schema.TypeList,
										Optional: true,
										MinItems: 1,
										MaxItems: 50,
										Elem: &schema.Schema{
											Type:         schema.TypeString,
											ValidateFunc: validation.StringLenBetween(1, 1000),
										},
									},
									"configuration": {
										Type:     schema.TypeMap,
										Optional: true,
										ValidateDiagFunc: validation.AllDiag(
											validation.MapKeyLenBetween(1, 50),
											validation.MapKeyLenBetween(1, 1000),
										),
										Elem:             &schema.Schema{Type: schema.TypeString},
										DiffSuppressFunc: pipelineSuppressStageActionConfigurationDiff,
									},
									"input_artifacts": {
										Type:     schema.TypeList,
										Optional: true,
										Elem:     &schema.Schema{Type: schema.TypeString},
									},
									"name": {
										Type:     schema.TypeString,
										Required: true,
										ValidateFunc: validation.All(
											validation.StringLenBetween(1, 100),
											validation.StringMatch(regexache.MustCompile(`[0-9A-Za-z_.@-]+`), ""),
										),
									},
									"namespace": {
										Type:     schema.TypeString,
										Optional: true,
										ValidateFunc: validation.All(
											validation.StringLenBetween(1, 100),
											validation.StringMatch(regexache.MustCompile(`[0-9A-Za-z_@-]+`), ""),
										),
									},
									"output_artifacts": {
										Type:     schema.TypeList,
										Optional: true,
										Elem:     &schema.Schema{Type: schema.TypeString},
									},
									"owner": {
										Type:             schema.TypeString,
										Required:         true,
										ValidateDiagFunc: enum.Validate[types.ActionOwner](),
									},
									"provider": {
										Type:             schema.TypeString,
										Required:         true,
										ValidateDiagFunc: pipelineValidateActionProvider,
									},
									"region": {
										Type:     schema.TypeString,
										Optional: true,
										Computed: true,
									},
									"role_arn": {
										Type:         schema.TypeString,
										Optional:     true,
										ValidateFunc: verify.ValidARN,
									},
									"run_order": {
										Type:         schema.TypeInt,
										Optional:     true,
										Computed:     true,
										ValidateFunc: validation.IntBetween(1, 999),
									},
									"version": {
										Type:     schema.TypeString,
										Required: true,
										ValidateFunc: validation.All(
											validation.StringLenBetween(1, 9),
											validation.StringMatch(regexache.MustCompile(`[0-9A-Za-z_-]+`), ""),
										),
									},
								},
							},
						},
						"name": {
							Type:     schema.TypeString,
							Required: true,
							ValidateFunc: validation.All(
								validation.StringLenBetween(1, 100),
								validation.StringMatch(regexache.MustCompile(`[0-9A-Za-z_.@-]+`), ""),
							),
						},
					},
				},
			},
			names.AttrTags:    tftags.TagsSchema(),
			names.AttrTagsAll: tftags.TagsSchemaComputed(),
			"variable": {
				Type:     schema.TypeList,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"default_value": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"description": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"name": {
							Type:     schema.TypeString,
							Required: true,
						},
					},
				},
			},
		},

		CustomizeDiff: verify.SetTagsDiff,
	}
}

func resourcePipelineCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	conn := meta.(*conns.AWSClient).CodePipelineClient(ctx)

	pipeline, err := expandPipelineDeclaration(d)
	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	name := d.Get("name").(string)
	input := &codepipeline.CreatePipelineInput{
		Pipeline: pipeline,
		Tags:     getTagsIn(ctx),
	}

	outputRaw, err := tfresource.RetryWhenIsAErrorMessageContains[*types.InvalidStructureException](ctx, propagationTimeout, func() (interface{}, error) {
		return conn.CreatePipeline(ctx, input)
	}, "not authorized")

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating CodePipeline Pipeline (%s): %s", name, err)
	}

	d.SetId(aws.ToString(outputRaw.(*codepipeline.CreatePipelineOutput).Pipeline.Name))

	return append(diags, resourcePipelineRead(ctx, d, meta)...)
}

func resourcePipelineRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	conn := meta.(*conns.AWSClient).CodePipelineClient(ctx)

	output, err := findPipelineByName(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] CodePipeline Pipeline %s not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading CodePipeline Pipeline (%s): %s", d.Id(), err)
	}

	metadata := output.Metadata
	pipeline := output.Pipeline
	arn := aws.ToString(metadata.PipelineArn)
	d.Set("arn", arn)
	if pipeline.ArtifactStore != nil {
		if err := d.Set("artifact_store", []interface{}{flattenArtifactStore(pipeline.ArtifactStore)}); err != nil {
			return sdkdiag.AppendErrorf(diags, "setting artifact_store: %s", err)
		}
	} else if pipeline.ArtifactStores != nil {
		if err := d.Set("artifact_store", flattenArtifactStores(pipeline.ArtifactStores)); err != nil {
			return sdkdiag.AppendErrorf(diags, "setting artifact_store: %s", err)
		}
	}
	d.Set("execution_mode", pipeline.ExecutionMode)
	d.Set("name", pipeline.Name)
	d.Set("pipeline_type", pipeline.PipelineType)
	d.Set("role_arn", pipeline.RoleArn)
	if err := d.Set("stage", flattenStageDeclarations(d, pipeline.Stages)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting stage: %s", err)
	}
	if err := d.Set("variable", flattenVariableDeclarations(pipeline.Variables)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting variable: %s", err)
	}

	return diags
}

func resourcePipelineUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	conn := meta.(*conns.AWSClient).CodePipelineClient(ctx)

	if d.HasChangesExcept("tags", "tags_all") {
		pipeline, err := expandPipelineDeclaration(d)
		if err != nil {
			return sdkdiag.AppendFromErr(diags, err)
		}

		input := &codepipeline.UpdatePipelineInput{
			Pipeline: pipeline,
		}

		_, err = conn.UpdatePipeline(ctx, input)

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "updating CodePipeline Pipeline (%s): %s", d.Id(), err)
		}
	}

	return append(diags, resourcePipelineRead(ctx, d, meta)...)
}

func resourcePipelineDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	conn := meta.(*conns.AWSClient).CodePipelineClient(ctx)

	log.Printf("[INFO] Deleting CodePipeline Pipeline: %s", d.Id())
	_, err := conn.DeletePipeline(ctx, &codepipeline.DeletePipelineInput{
		Name: aws.String(d.Id()),
	})

	if errs.IsA[*types.PipelineNotFoundException](err) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting CodePipeline Pipeline (%s): %s", d.Id(), err)
	}

	return diags
}

func findPipelineByName(ctx context.Context, conn *codepipeline.Client, name string) (*codepipeline.GetPipelineOutput, error) {
	input := &codepipeline.GetPipelineInput{
		Name: aws.String(name),
	}

	output, err := conn.GetPipeline(ctx, input)

	if errs.IsA[*types.PipelineNotFoundException](err) {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil || output.Metadata == nil || output.Pipeline == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	return output, nil
}

func pipelineValidateActionProvider(i interface{}, path cty.Path) (diags diag.Diagnostics) {
	v, ok := i.(string)
	if !ok {
		return sdkdiag.AppendErrorf(diags, "expected type to be string")
	}

	if v == providerGitHub {
		return diag.Diagnostics{
			diag.Diagnostic{
				Severity: diag.Warning,
				Summary:  "The CodePipeline GitHub version 1 action provider is deprecated.",
				Detail:   "Use a GitHub version 2 action (with a CodeStar Connection `aws_codestarconnections_connection`) instead. See https://docs.aws.amazon.com/codepipeline/latest/userguide/update-github-action-connections.html",
			},
		}
	}

	return diags
}

func pipelineSuppressStageActionConfigurationDiff(k, old, new string, d *schema.ResourceData) bool {
	parts := strings.Split(k, ".")
	parts = parts[:len(parts)-2]
	providerAddr := strings.Join(append(parts, "provider"), ".")
	provider := d.Get(providerAddr).(string)

	if provider == providerGitHub && strings.HasSuffix(k, gitHubActionConfigurationOAuthToken) {
		hash := hashGitHubToken(new)
		return old == hash
	}

	return false
}

func hashGitHubToken(token string) string {
	const gitHubTokenHashPrefix = "hash-"

	// Without this check, the value was getting encoded twice
	if strings.HasPrefix(token, gitHubTokenHashPrefix) {
		return token
	}

	sum := sha256.Sum256([]byte(token))
	return gitHubTokenHashPrefix + hex.EncodeToString(sum[:])
}

func expandPipelineDeclaration(d *schema.ResourceData) (*types.PipelineDeclaration, error) {
	apiObject := &types.PipelineDeclaration{}

	if v, ok := d.GetOk("artifact_store"); ok && v.(*schema.Set).Len() > 0 {
		artifactStores := expandArtifactStores(v.(*schema.Set).List())

		switch n := len(artifactStores); n {
		case 1:
			for region, v := range artifactStores {
				if region != "" {
					return nil, errors.New("region cannot be set for a single-region CodePipeline")
				}
				v := v
				apiObject.ArtifactStore = &v
			}

		default:
			for region := range artifactStores {
				if region == "" {
					return nil, errors.New("region must be set for a cross-region CodePipeline")
				}
			}
			if n != v.(*schema.Set).Len() {
				return nil, errors.New("only one Artifact Store can be defined per region for a cross-region CodePipeline")
			}
			apiObject.ArtifactStores = artifactStores
		}
	}

	if v, ok := d.GetOk("execution_mode"); ok {
		apiObject.ExecutionMode = types.ExecutionMode(v.(string))
	}

	if v, ok := d.GetOk("name"); ok {
		apiObject.Name = aws.String(v.(string))
	}

	if v, ok := d.GetOk("pipeline_type"); ok {
		apiObject.PipelineType = types.PipelineType(v.(string))
	}

	if v, ok := d.GetOk("role_arn"); ok {
		apiObject.RoleArn = aws.String(v.(string))
	}

	if v, ok := d.GetOk("stage"); ok && len(v.([]interface{})) > 0 {
		apiObject.Stages = expandStageDeclarations(v.([]interface{}))
	}

	if v, ok := d.GetOk("variable"); ok && len(v.([]interface{})) > 0 {
		apiObject.Variables = expandVariableDeclarations(v.([]interface{}))
	}

	return apiObject, nil
}

func expandArtifactStore(tfMap map[string]interface{}) *types.ArtifactStore {
	if tfMap == nil {
		return nil
	}

	apiObject := &types.ArtifactStore{}

	if v, ok := tfMap["encryption_key"].([]interface{}); ok && len(v) > 0 && v[0] != nil {
		apiObject.EncryptionKey = expandEncryptionKey(v[0].(map[string]interface{}))
	}

	if v, ok := tfMap["location"].(string); ok && v != "" {
		apiObject.Location = aws.String(v)
	}

	if v, ok := tfMap["type"].(string); ok && v != "" {
		apiObject.Type = types.ArtifactStoreType(v)
	}

	return apiObject
}

func expandArtifactStores(tfList []interface{}) map[string]types.ArtifactStore {
	if len(tfList) == 0 {
		return nil
	}

	apiObjects := make(map[string]types.ArtifactStore, 0)

	for _, tfMapRaw := range tfList {
		tfMap, ok := tfMapRaw.(map[string]interface{})

		if !ok {
			continue
		}

		apiObject := expandArtifactStore(tfMap)

		if apiObject == nil {
			continue
		}

		var region string

		if v, ok := tfMap["region"].(string); ok && v != "" {
			region = v
		}

		apiObjects[region] = *apiObject // nosemgrep:ci.semgrep.aws.prefer-pointer-conversion-assignment
	}

	return apiObjects
}

func expandEncryptionKey(tfMap map[string]interface{}) *types.EncryptionKey {
	if tfMap == nil {
		return nil
	}

	apiObject := &types.EncryptionKey{}

	if v, ok := tfMap["id"].(string); ok && v != "" {
		apiObject.Id = aws.String(v)
	}

	if v, ok := tfMap["type"].(string); ok && v != "" {
		apiObject.Type = types.EncryptionKeyType(v)
	}

	return apiObject
}

func expandStageDeclaration(tfMap map[string]interface{}) *types.StageDeclaration {
	if tfMap == nil {
		return nil
	}

	apiObject := &types.StageDeclaration{}

	if v, ok := tfMap["action"].([]interface{}); ok && len(v) > 0 {
		apiObject.Actions = expandActionDeclarations(v)
	}

	if v, ok := tfMap["name"].(string); ok && v != "" {
		apiObject.Name = aws.String(v)
	}

	return apiObject
}

func expandStageDeclarations(tfList []interface{}) []types.StageDeclaration {
	if len(tfList) == 0 {
		return nil
	}

	var apiObjects []types.StageDeclaration

	for _, tfMapRaw := range tfList {
		tfMap, ok := tfMapRaw.(map[string]interface{})

		if !ok {
			continue
		}

		apiObject := expandStageDeclaration(tfMap)

		if apiObject == nil {
			continue
		}

		apiObjects = append(apiObjects, *apiObject)
	}

	return apiObjects
}

func expandActionDeclaration(tfMap map[string]interface{}) *types.ActionDeclaration {
	if tfMap == nil {
		return nil
	}

	apiObject := &types.ActionDeclaration{
		ActionTypeId: &types.ActionTypeId{},
	}

	if v, ok := tfMap["category"].(string); ok && v != "" {
		apiObject.ActionTypeId.Category = types.ActionCategory(v)
	}

	if v, ok := tfMap["commands"]; ok && len(v.([]interface{})) > 0 {
		apiObject.Commands = flex.ExpandStringValueList(v.([]interface{}))
	}

	if v, ok := tfMap["configuration"].(map[string]interface{}); ok && len(v) > 0 {
		apiObject.Configuration = flex.ExpandStringValueMap(v)
	}

	if v, ok := tfMap["input_artifacts"].([]interface{}); ok && len(v) > 0 {
		apiObject.InputArtifacts = expandInputArtifacts(v)
	}

	if v, ok := tfMap["name"].(string); ok && v != "" {
		apiObject.Name = aws.String(v)
	}

	if v, ok := tfMap["namespace"].(string); ok && v != "" {
		apiObject.Namespace = aws.String(v)
	}

	if v, ok := tfMap["output_artifacts"].([]interface{}); ok && len(v) > 0 {
		apiObject.OutputArtifacts = expandOutputArtifacts(v)
	}

	if v, ok := tfMap["owner"].(string); ok && v != "" {
		apiObject.ActionTypeId.Owner = types.ActionOwner(v)
	}

	if v, ok := tfMap["provider"].(string); ok && v != "" {
		apiObject.ActionTypeId.Provider = aws.String(v)
	}

	if v, ok := tfMap["region"].(string); ok && v != "" {
		apiObject.Region = aws.String(v)
	}

	if v, ok := tfMap["role_arn"].(string); ok && v != "" {
		apiObject.RoleArn = aws.String(v)
	}

	if v, ok := tfMap["run_order"].(int); ok && v != 0 {
		apiObject.RunOrder = aws.Int32(int32(v))
	}

	if v, ok := tfMap["version"].(string); ok && v != "" {
		apiObject.ActionTypeId.Version = aws.String(v)
	}

	return apiObject
}

func expandActionDeclarations(tfList []interface{}) []types.ActionDeclaration {
	if len(tfList) == 0 {
		return nil
	}

	var apiObjects []types.ActionDeclaration

	for _, tfMapRaw := range tfList {
		tfMap, ok := tfMapRaw.(map[string]interface{})

		if !ok {
			continue
		}

		apiObject := expandActionDeclaration(tfMap)

		if apiObject == nil {
			continue
		}

		apiObjects = append(apiObjects, *apiObject)
	}

	return apiObjects
}

func expandInputArtifacts(tfList []interface{}) []types.InputArtifact {
	if len(tfList) == 0 {
		return nil
	}

	var apiObjects []types.InputArtifact

	for _, v := range tfList {
		v, ok := v.(string)

		if !ok {
			continue
		}

		apiObject := types.InputArtifact{
			Name: aws.String(v),
		}

		apiObjects = append(apiObjects, apiObject)
	}

	return apiObjects
}

func expandOutputArtifacts(tfList []interface{}) []types.OutputArtifact {
	if len(tfList) == 0 {
		return nil
	}

	var apiObjects []types.OutputArtifact

	for _, v := range tfList {
		v, ok := v.(string)

		if !ok {
			continue
		}

		apiObject := types.OutputArtifact{
			Name: aws.String(v),
		}

		apiObjects = append(apiObjects, apiObject)
	}

	return apiObjects
}

func expandVariableDeclaration(tfMap map[string]interface{}) *types.PipelineVariableDeclaration {
	if tfMap == nil {
		return nil
	}

	apiObject := &types.PipelineVariableDeclaration{}

	if v, ok := tfMap["default_value"].(string); ok && v != "" {
		apiObject.DefaultValue = aws.String(v)
	}

	if v, ok := tfMap["description"].(string); ok && v != "" {
		apiObject.Description = aws.String(v)
	}

	if v, ok := tfMap["name"].(string); ok && v != "" {
		apiObject.Name = aws.String(v)
	}

	return apiObject
}

func expandVariableDeclarations(tfList []interface{}) []types.PipelineVariableDeclaration {
	if len(tfList) == 0 {
		return nil
	}

	var apiObjects []types.PipelineVariableDeclaration

	for _, tfMapRaw := range tfList {
		tfMap, ok := tfMapRaw.(map[string]interface{})

		if !ok {
			continue
		}

		apiObject := expandVariableDeclaration(tfMap)

		if apiObject == nil {
			continue
		}

		apiObjects = append(apiObjects, *apiObject)
	}

	return apiObjects
}

func flattenArtifactStore(apiObject *types.ArtifactStore) map[string]interface{} {
	tfMap := map[string]interface{}{
		"type": apiObject.Type,
	}

	if v := apiObject.EncryptionKey; v != nil {
		tfMap["encryption_key"] = []interface{}{flattenEncryptionKey(v)}
	}

	if v := apiObject.Location; v != nil {
		tfMap["location"] = aws.ToString(v)
	}

	return tfMap
}

func flattenArtifactStores(apiObjects map[string]types.ArtifactStore) []interface{} {
	if len(apiObjects) == 0 {
		return nil
	}

	var tfList []interface{}

	for region, apiObject := range apiObjects {
		tfMap := flattenArtifactStore(&apiObject)
		tfMap["region"] = region

		tfList = append(tfList, tfMap)
	}

	return tfList
}

func flattenEncryptionKey(apiObject *types.EncryptionKey) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{
		"type": apiObject.Type,
	}

	if v := apiObject.Id; v != nil {
		tfMap["id"] = aws.ToString(v)
	}

	return tfMap
}

func flattenStageDeclaration(d *schema.ResourceData, i int, apiObject types.StageDeclaration) map[string]interface{} {
	tfMap := map[string]interface{}{}

	if v := apiObject.Actions; v != nil {
		tfMap["action"] = flattenActionDeclarations(d, i, v)
	}

	if v := apiObject.Name; v != nil {
		tfMap["name"] = aws.ToString(v)
	}

	return tfMap
}

func flattenStageDeclarations(d *schema.ResourceData, apiObjects []types.StageDeclaration) []interface{} {
	if len(apiObjects) == 0 {
		return nil
	}

	var tfList []interface{}

	for i, apiObject := range apiObjects {
		tfList = append(tfList, flattenStageDeclaration(d, i, apiObject))
	}

	return tfList
}

func flattenActionDeclaration(d *schema.ResourceData, i, j int, apiObject types.ActionDeclaration) map[string]interface{} {
	var actionProvider string
	tfMap := map[string]interface{}{}

	if apiObject := apiObject.ActionTypeId; apiObject != nil {
		tfMap["category"] = apiObject.Category
		tfMap["owner"] = apiObject.Owner

		if v := apiObject.Provider; v != nil {
			actionProvider = aws.ToString(v)
			tfMap["provider"] = actionProvider
		}

		if v := apiObject.Version; v != nil {
			tfMap["version"] = aws.ToString(v)
		}
	}

	if v := apiObject.Commands; len(v) > 0 {
		tfMap["commands"] = apiObject.Commands
	}

	if v := apiObject.Configuration; v != nil {
		// The AWS API returns "****" for the OAuthToken value. Copy the value from the configuration.
		if actionProvider == providerGitHub {
			if _, ok := v[gitHubActionConfigurationOAuthToken]; ok {
				key := fmt.Sprintf("stage.%d.action.%d.configuration.OAuthToken", i, j)
				v[gitHubActionConfigurationOAuthToken] = d.Get(key).(string)
			}
		}

		tfMap["configuration"] = v
	}

	if v := apiObject.InputArtifacts; len(v) > 0 {
		tfMap["input_artifacts"] = flattenInputArtifacts(v)
	}

	if v := apiObject.Name; v != nil {
		tfMap["name"] = aws.ToString(v)
	}

	if v := apiObject.Namespace; v != nil {
		tfMap["namespace"] = aws.ToString(v)
	}

	if v := apiObject.OutputArtifacts; len(v) > 0 {
		tfMap["output_artifacts"] = flattenOutputArtifacts(v)
	}

	if v := apiObject.Region; v != nil {
		tfMap["region"] = aws.ToString(v)
	}

	if v := apiObject.RoleArn; v != nil {
		tfMap["role_arn"] = aws.ToString(v)
	}

	if v := apiObject.RunOrder; v != nil {
		tfMap["run_order"] = aws.ToInt32(v)
	}

	return tfMap
}

func flattenActionDeclarations(d *schema.ResourceData, i int, apiObjects []types.ActionDeclaration) []interface{} {
	if len(apiObjects) == 0 {
		return nil
	}

	var tfList []interface{}

	for j, apiObject := range apiObjects {
		tfList = append(tfList, flattenActionDeclaration(d, i, j, apiObject))
	}

	return tfList
}

func flattenInputArtifacts(apiObjects []types.InputArtifact) []string {
	if len(apiObjects) == 0 {
		return nil
	}

	var tfList []*string

	for _, apiObject := range apiObjects {
		tfList = append(tfList, apiObject.Name)
	}

	return aws.ToStringSlice(tfList)
}

func flattenOutputArtifacts(apiObjects []types.OutputArtifact) []string {
	if len(apiObjects) == 0 {
		return nil
	}

	var tfList []*string

	for _, apiObject := range apiObjects {
		tfList = append(tfList, apiObject.Name)
	}

	return aws.ToStringSlice(tfList)
}

func flattenVariableDeclaration(apiObject types.PipelineVariableDeclaration) map[string]interface{} {
	tfMap := map[string]interface{}{}

	if v := apiObject.DefaultValue; v != nil {
		tfMap["default_value"] = aws.ToString(v)
	}

	if v := apiObject.Description; v != nil {
		tfMap["description"] = aws.ToString(v)
	}

	if v := apiObject.Name; v != nil {
		tfMap["name"] = aws.ToString(v)
	}

	return tfMap
}

func flattenVariableDeclarations(apiObjects []types.PipelineVariableDeclaration) []interface{} {
	if len(apiObjects) == 0 {
		return nil
	}

	var tfList []interface{}

	for _, apiObject := range apiObjects {
		tfList = append(tfList, flattenVariableDeclaration(apiObject))
	}

	return tfList
}
