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
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/codepipeline"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/go-cty/cty"
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
	providerGitHub = "GitHub"

	gitHubActionConfigurationOAuthToken = "OAuthToken"
)

// @SDKResource("aws_codepipeline", name="Pipeline")
// @Tags(identifierAttribute="arn")
func ResourcePipeline() *schema.Resource {
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
										Type:         schema.TypeString,
										Required:     true,
										ValidateFunc: validation.StringInSlice(codepipeline.EncryptionKeyType_Values(), false),
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
							Computed: true,
						},
						"type": {
							Type:         schema.TypeString,
							Required:     true,
							ValidateFunc: validation.StringInSlice(codepipeline.ArtifactStoreType_Values(), false),
						},
					},
				},
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
										Type:         schema.TypeString,
										Required:     true,
										ValidateFunc: validation.StringInSlice(codepipeline.ActionCategory_Values(), false),
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
										Type:         schema.TypeString,
										Required:     true,
										ValidateFunc: validation.StringInSlice(codepipeline.ActionOwner_Values(), false),
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
		},

		CustomizeDiff: verify.SetTagsDiff,
	}
}

func resourcePipelineCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	conn := meta.(*conns.AWSClient).CodePipelineConn(ctx)

	pipeline, err := expandPipelineDeclaration(d)

	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	name := d.Get("name").(string)
	input := &codepipeline.CreatePipelineInput{
		Pipeline: pipeline,
		Tags:     getTagsIn(ctx),
	}

	outputRaw, err := tfresource.RetryWhenAWSErrMessageContains(ctx, propagationTimeout, func() (interface{}, error) {
		return conn.CreatePipelineWithContext(ctx, input)
	}, codepipeline.ErrCodeInvalidStructureException, "not authorized")

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating CodePipeline (%s): %s", name, err)
	}

	d.SetId(aws.StringValue(outputRaw.(*codepipeline.CreatePipelineOutput).Pipeline.Name))

	return append(diags, resourcePipelineRead(ctx, d, meta)...)
}

func resourcePipelineRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	conn := meta.(*conns.AWSClient).CodePipelineConn(ctx)

	output, err := FindPipelineByName(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] CodePipeline %s not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading CodePipeline (%s): %s", d.Id(), err)
	}

	metadata := output.Metadata
	pipeline := output.Pipeline

	if pipeline.ArtifactStore != nil {
		if err := d.Set("artifact_store", []interface{}{flattenArtifactStore(pipeline.ArtifactStore)}); err != nil {
			return sdkdiag.AppendErrorf(diags, "setting artifact_store: %s", err)
		}
	} else if pipeline.ArtifactStores != nil {
		if err := d.Set("artifact_store", flattenArtifactStores(pipeline.ArtifactStores)); err != nil {
			return sdkdiag.AppendErrorf(diags, "setting artifact_store: %s", err)
		}
	}

	if err := d.Set("stage", flattenStageDeclarations(d, pipeline.Stages)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting stage: %s", err)
	}

	arn := aws.StringValue(metadata.PipelineArn)
	d.Set("arn", arn)
	d.Set("name", pipeline.Name)
	d.Set("role_arn", pipeline.RoleArn)

	return diags
}

func resourcePipelineUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	conn := meta.(*conns.AWSClient).CodePipelineConn(ctx)

	if d.HasChangesExcept("tags", "tags_all") {
		pipeline, err := expandPipelineDeclaration(d)

		if err != nil {
			return sdkdiag.AppendFromErr(diags, err)
		}

		_, err = conn.UpdatePipelineWithContext(ctx, &codepipeline.UpdatePipelineInput{
			Pipeline: pipeline,
		})

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "updating CodePipeline (%s): %s", d.Id(), err)
		}
	}

	return append(diags, resourcePipelineRead(ctx, d, meta)...)
}

func resourcePipelineDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	conn := meta.(*conns.AWSClient).CodePipelineConn(ctx)

	log.Printf("[INFO] Deleting CodePipeline: %s", d.Id())
	_, err := conn.DeletePipelineWithContext(ctx, &codepipeline.DeletePipelineInput{
		Name: aws.String(d.Id()),
	})

	if tfawserr.ErrCodeEquals(err, codepipeline.ErrCodePipelineNotFoundException) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting CodePipeline (%s): %s", d.Id(), err)
	}

	return diags
}

func FindPipelineByName(ctx context.Context, conn *codepipeline.CodePipeline, name string) (*codepipeline.GetPipelineOutput, error) {
	input := &codepipeline.GetPipelineInput{
		Name: aws.String(name),
	}

	output, err := conn.GetPipelineWithContext(ctx, input)

	if tfawserr.ErrCodeEquals(err, codepipeline.ErrCodePipelineNotFoundException) {
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

func expandPipelineDeclaration(d *schema.ResourceData) (*codepipeline.PipelineDeclaration, error) {
	apiObject := &codepipeline.PipelineDeclaration{}

	if v, ok := d.GetOk("artifact_store"); ok && v.(*schema.Set).Len() > 0 {
		artifactStores := expandArtifactStores(v.(*schema.Set).List())

		switch n := len(artifactStores); n {
		case 1:
			for region, v := range artifactStores {
				if region != "" {
					return nil, errors.New("region cannot be set for a single-region CodePipeline")
				}
				apiObject.ArtifactStore = v
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

	if v, ok := d.GetOk("name"); ok {
		apiObject.Name = aws.String(v.(string))
	}

	if v, ok := d.GetOk("role_arn"); ok {
		apiObject.RoleArn = aws.String(v.(string))
	}

	if v, ok := d.GetOk("stage"); ok && len(v.([]interface{})) > 0 {
		apiObject.Stages = expandStageDeclarations(v.([]interface{}))
	}

	return apiObject, nil
}

func expandArtifactStore(tfMap map[string]interface{}) *codepipeline.ArtifactStore {
	if tfMap == nil {
		return nil
	}

	apiObject := &codepipeline.ArtifactStore{}

	if v, ok := tfMap["encryption_key"].([]interface{}); ok && len(v) > 0 && v[0] != nil {
		apiObject.EncryptionKey = expandEncryptionKey(v[0].(map[string]interface{}))
	}

	if v, ok := tfMap["location"].(string); ok && v != "" {
		apiObject.Location = aws.String(v)
	}

	if v, ok := tfMap["type"].(string); ok && v != "" {
		apiObject.Type = aws.String(v)
	}

	return apiObject
}

func expandArtifactStores(tfList []interface{}) map[string]*codepipeline.ArtifactStore {
	if len(tfList) == 0 {
		return nil
	}

	apiObjects := make(map[string]*codepipeline.ArtifactStore, 0)

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

		apiObjects[region] = apiObject
	}

	return apiObjects
}

func expandEncryptionKey(tfMap map[string]interface{}) *codepipeline.EncryptionKey {
	if tfMap == nil {
		return nil
	}

	apiObject := &codepipeline.EncryptionKey{}

	if v, ok := tfMap["id"].(string); ok && v != "" {
		apiObject.Id = aws.String(v)
	}

	if v, ok := tfMap["type"].(string); ok && v != "" {
		apiObject.Type = aws.String(v)
	}

	return apiObject
}

func expandStageDeclaration(tfMap map[string]interface{}) *codepipeline.StageDeclaration {
	if tfMap == nil {
		return nil
	}

	apiObject := &codepipeline.StageDeclaration{}

	if v, ok := tfMap["action"].([]interface{}); ok && len(v) > 0 {
		apiObject.Actions = expandActionDeclarations(v)
	}

	if v, ok := tfMap["name"].(string); ok && v != "" {
		apiObject.Name = aws.String(v)
	}

	return apiObject
}

func expandStageDeclarations(tfList []interface{}) []*codepipeline.StageDeclaration {
	if len(tfList) == 0 {
		return nil
	}

	var apiObjects []*codepipeline.StageDeclaration

	for _, tfMapRaw := range tfList {
		tfMap, ok := tfMapRaw.(map[string]interface{})

		if !ok {
			continue
		}

		apiObject := expandStageDeclaration(tfMap)

		if apiObject == nil {
			continue
		}

		apiObjects = append(apiObjects, apiObject)
	}

	return apiObjects
}

func expandActionDeclaration(tfMap map[string]interface{}) *codepipeline.ActionDeclaration {
	if tfMap == nil {
		return nil
	}

	apiObject := &codepipeline.ActionDeclaration{
		ActionTypeId: &codepipeline.ActionTypeId{},
	}

	if v, ok := tfMap["category"].(string); ok && v != "" {
		apiObject.ActionTypeId.Category = aws.String(v)
	}

	if v, ok := tfMap["configuration"].(map[string]interface{}); ok && len(v) > 0 {
		apiObject.Configuration = flex.ExpandStringMap(v)
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
		apiObject.ActionTypeId.Owner = aws.String(v)
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
		apiObject.RunOrder = aws.Int64(int64(v))
	}

	if v, ok := tfMap["version"].(string); ok && v != "" {
		apiObject.ActionTypeId.Version = aws.String(v)
	}

	return apiObject
}

func expandActionDeclarations(tfList []interface{}) []*codepipeline.ActionDeclaration {
	if len(tfList) == 0 {
		return nil
	}

	var apiObjects []*codepipeline.ActionDeclaration

	for _, tfMapRaw := range tfList {
		tfMap, ok := tfMapRaw.(map[string]interface{})

		if !ok {
			continue
		}

		apiObject := expandActionDeclaration(tfMap)

		if apiObject == nil {
			continue
		}

		apiObjects = append(apiObjects, apiObject)
	}

	return apiObjects
}

func expandInputArtifacts(tfList []interface{}) []*codepipeline.InputArtifact {
	if len(tfList) == 0 {
		return nil
	}

	var apiObjects []*codepipeline.InputArtifact

	for _, v := range tfList {
		v, ok := v.(string)

		if !ok {
			continue
		}

		apiObject := &codepipeline.InputArtifact{
			Name: aws.String(v),
		}

		apiObjects = append(apiObjects, apiObject)
	}

	return apiObjects
}

func expandOutputArtifacts(tfList []interface{}) []*codepipeline.OutputArtifact {
	if len(tfList) == 0 {
		return nil
	}

	var apiObjects []*codepipeline.OutputArtifact

	for _, v := range tfList {
		v, ok := v.(string)

		if !ok {
			continue
		}

		apiObject := &codepipeline.OutputArtifact{
			Name: aws.String(v),
		}

		apiObjects = append(apiObjects, apiObject)
	}

	return apiObjects
}

func flattenArtifactStore(apiObject *codepipeline.ArtifactStore) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if v := apiObject.EncryptionKey; v != nil {
		tfMap["encryption_key"] = []interface{}{flattenEncryptionKey(v)}
	}

	if v := apiObject.Location; v != nil {
		tfMap["location"] = aws.StringValue(v)
	}

	if v := apiObject.Type; v != nil {
		tfMap["type"] = aws.StringValue(v)
	}

	return tfMap
}

func flattenArtifactStores(apiObjects map[string]*codepipeline.ArtifactStore) []interface{} {
	if len(apiObjects) == 0 {
		return nil
	}

	var tfList []interface{}

	for region, apiObject := range apiObjects {
		if apiObject == nil {
			continue
		}

		tfMap := flattenArtifactStore(apiObject)
		tfMap["region"] = region

		tfList = append(tfList, tfMap)
	}

	return tfList
}

func flattenEncryptionKey(apiObject *codepipeline.EncryptionKey) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if v := apiObject.Id; v != nil {
		tfMap["id"] = aws.StringValue(v)
	}

	if v := apiObject.Type; v != nil {
		tfMap["type"] = aws.StringValue(v)
	}

	return tfMap
}

func flattenStageDeclaration(d *schema.ResourceData, i int, apiObject *codepipeline.StageDeclaration) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if v := apiObject.Actions; v != nil {
		tfMap["action"] = flattenActionDeclarations(d, i, v)
	}

	if v := apiObject.Name; v != nil {
		tfMap["name"] = aws.StringValue(v)
	}

	return tfMap
}

func flattenStageDeclarations(d *schema.ResourceData, apiObjects []*codepipeline.StageDeclaration) []interface{} {
	if len(apiObjects) == 0 {
		return nil
	}

	var tfList []interface{}

	for i, apiObject := range apiObjects {
		if apiObject == nil {
			continue
		}

		tfList = append(tfList, flattenStageDeclaration(d, i, apiObject))
	}

	return tfList
}

func flattenActionDeclaration(d *schema.ResourceData, i, j int, apiObject *codepipeline.ActionDeclaration) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	var actionProvider string
	tfMap := map[string]interface{}{}

	if apiObject := apiObject.ActionTypeId; apiObject != nil {
		if v := apiObject.Category; v != nil {
			tfMap["category"] = aws.StringValue(v)
		}

		if v := apiObject.Owner; v != nil {
			tfMap["owner"] = aws.StringValue(v)
		}

		if v := apiObject.Provider; v != nil {
			actionProvider = aws.StringValue(v)
			tfMap["provider"] = actionProvider
		}

		if v := apiObject.Version; v != nil {
			tfMap["version"] = aws.StringValue(v)
		}
	}

	if v := apiObject.Configuration; v != nil {
		v := aws.StringValueMap(v)

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
		tfMap["name"] = aws.StringValue(v)
	}

	if v := apiObject.Namespace; v != nil {
		tfMap["namespace"] = aws.StringValue(v)
	}

	if v := apiObject.OutputArtifacts; len(v) > 0 {
		tfMap["output_artifacts"] = flattenOutputArtifacts(v)
	}

	if v := apiObject.Region; v != nil {
		tfMap["region"] = aws.StringValue(v)
	}

	if v := apiObject.RoleArn; v != nil {
		tfMap["role_arn"] = aws.StringValue(v)
	}

	if v := apiObject.RunOrder; v != nil {
		tfMap["run_order"] = aws.Int64Value(v)
	}

	return tfMap
}

func flattenActionDeclarations(d *schema.ResourceData, i int, apiObjects []*codepipeline.ActionDeclaration) []interface{} {
	if len(apiObjects) == 0 {
		return nil
	}

	var tfList []interface{}

	for j, apiObject := range apiObjects {
		if apiObject == nil {
			continue
		}

		tfList = append(tfList, flattenActionDeclaration(d, i, j, apiObject))
	}

	return tfList
}

func flattenInputArtifacts(apiObjects []*codepipeline.InputArtifact) []string {
	if len(apiObjects) == 0 {
		return nil
	}

	var tfList []*string

	for _, apiObject := range apiObjects {
		if apiObject == nil {
			continue
		}

		tfList = append(tfList, apiObject.Name)
	}

	return aws.StringValueSlice(tfList)
}

func flattenOutputArtifacts(apiObjects []*codepipeline.OutputArtifact) []string {
	if len(apiObjects) == 0 {
		return nil
	}

	var tfList []*string

	for _, apiObject := range apiObjects {
		if apiObject == nil {
			continue
		}

		tfList = append(tfList, apiObject.Name)
	}

	return aws.StringValueSlice(tfList)
}
