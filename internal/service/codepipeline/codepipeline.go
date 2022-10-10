package codepipeline

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"log"
	"regexp"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/codepipeline"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/go-cty/cty"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

const (
	providerGitHub = "GitHub"

	gitHubActionConfigurationOAuthToken = "OAuthToken"
)

func ResourcePipeline() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourcePipelineCreate,
		ReadWithoutTimeout:   resourcePipelineRead,
		UpdateWithoutTimeout: resourcePipelineUpdate,
		DeleteWithoutTimeout: resourcePipelineDelete,

		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
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
					validation.StringMatch(regexp.MustCompile(`[A-Za-z0-9.@\-_]+`), ""),
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
										ValidateDiagFunc: verify.ValidAllDiag(
											validation.MapKeyLenBetween(1, 50),
											validation.MapKeyLenBetween(1, 1000),
										),
										Elem:             &schema.Schema{Type: schema.TypeString},
										DiffSuppressFunc: suppressStageActionConfiguration,
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
											validation.StringMatch(regexp.MustCompile(`[A-Za-z0-9.@\-_]+`), ""),
										),
									},
									"namespace": {
										Type:     schema.TypeString,
										Optional: true,
										ValidateFunc: validation.All(
											validation.StringLenBetween(1, 100),
											validation.StringMatch(regexp.MustCompile(`[A-Za-z0-9@\-_]+`), ""),
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
										ValidateDiagFunc: resourceValidateActionProvider,
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
											validation.StringMatch(regexp.MustCompile(`[0-9A-Za-z_-]+`), ""),
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
								validation.StringMatch(regexp.MustCompile(`[A-Za-z0-9.@\-_]+`), ""),
							),
						},
					},
				},
			},
			"tags":     tftags.TagsSchema(),
			"tags_all": tftags.TagsSchemaComputed(),
		},

		CustomizeDiff: verify.SetTagsDiff,
	}
}

func resourcePipelineCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).CodePipelineConn
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	tags := defaultTagsConfig.MergeTags(tftags.New(d.Get("tags").(map[string]interface{})))

	pipeline, err := expandPipelineDeclaration(d)

	if err != nil {
		return diag.FromErr(err)
	}

	name := d.Get("name").(string)
	input := &codepipeline.CreatePipelineInput{
		Pipeline: pipeline,
	}

	if len(tags) > 0 {
		input.Tags = Tags(tags.IgnoreAWS())
	}

	outputRaw, err := tfresource.RetryWhenAWSErrMessageContainsContext(ctx, propagationTimeout, func() (interface{}, error) {
		return conn.CreatePipelineWithContext(ctx, input)
	}, codepipeline.ErrCodeInvalidStructureException, "not authorized")

	if err != nil {
		return diag.Errorf("creating CodePipeline (%s): %s", name, err)
	}

	d.SetId(aws.StringValue(outputRaw.(*codepipeline.CreatePipelineOutput).Pipeline.Name))

	return resourcePipelineRead(ctx, d, meta)
}

func resourcePipelineRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).CodePipelineConn
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	ignoreTagsConfig := meta.(*conns.AWSClient).IgnoreTagsConfig

	output, err := FindPipelineByName(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] CodePipeline %s not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return diag.Errorf("reading CodePipeline (%s): %s", d.Id(), err)
	}

	metadata := output.Metadata
	pipeline := output.Pipeline

	if pipeline.ArtifactStore != nil {
		if err := d.Set("artifact_store", flattenArtifactStore(pipeline.ArtifactStore)); err != nil {
			return diag.Errorf("setting artifact_store: %s", err)
		}
	} else if pipeline.ArtifactStores != nil {
		if err := d.Set("artifact_store", flattenArtifactStores(pipeline.ArtifactStores)); err != nil {
			return diag.Errorf("setting artifact_store: %s", err)
		}
	}

	if err := d.Set("stage", flattenStages(pipeline.Stages, d)); err != nil {
		return diag.Errorf("setting stage: %s", err)
	}

	arn := aws.StringValue(metadata.PipelineArn)
	d.Set("arn", arn)
	d.Set("name", pipeline.Name)
	d.Set("role_arn", pipeline.RoleArn)

	tags, err := ListTagsWithContext(ctx, conn, arn)

	if err != nil {
		return diag.Errorf("listing tags for CodePipeline (%s): %s", arn, err)
	}

	tags = tags.IgnoreAWS().IgnoreConfig(ignoreTagsConfig)

	//lintignore:AWSR002
	if err := d.Set("tags", tags.RemoveDefaultConfig(defaultTagsConfig).Map()); err != nil {
		return diag.Errorf("setting tags: %s", err)
	}

	if err := d.Set("tags_all", tags.Map()); err != nil {
		return diag.Errorf("setting tags_all: %s", err)
	}

	return nil
}

func resourcePipelineUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).CodePipelineConn

	if d.HasChangesExcept("tags", "tags_all") {
		pipeline, err := expandPipelineDeclaration(d)

		if err != nil {
			return diag.FromErr(err)
		}

		_, err = conn.UpdatePipelineWithContext(ctx, &codepipeline.UpdatePipelineInput{
			Pipeline: pipeline,
		})

		if err != nil {
			return diag.Errorf("updating CodePipeline (%s): %s", d.Id(), err)
		}
	}

	arn := d.Get("arn").(string)
	if d.HasChange("tags_all") {
		o, n := d.GetChange("tags_all")

		if err := UpdateTagsWithContext(ctx, conn, arn, o, n); err != nil {
			return diag.Errorf("updating CodePipeline (%s) tags: %s", arn, err)
		}
	}

	return resourcePipelineRead(ctx, d, meta)
}

func resourcePipelineDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).CodePipelineConn

	log.Printf("[INFO] Deleting CodePipeline: %s", d.Id())
	_, err := conn.DeletePipelineWithContext(ctx, &codepipeline.DeletePipelineInput{
		Name: aws.String(d.Id()),
	})

	if tfawserr.ErrCodeEquals(err, codepipeline.ErrCodePipelineNotFoundException) {
		return nil
	}

	if err != nil {
		return diag.Errorf("deleting CodePipeline (%s): %s", d.Id(), err)
	}

	return nil
}

func FindPipelineByName(ctx context.Context, conn *codepipeline.CodePipeline, name string) (*codepipeline.GetPipelineOutput, error) {
	input := &codepipeline.GetPipelineInput{
		Name: aws.String(name),
	}

	output, err := conn.GetPipelineWithContext(ctx, input)

	if tfawserr.ErrCodeEquals(err, codepipeline.ErrCodePipelineNotFoundException) {
		return nil, &resource.NotFoundError{
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

func flattenArtifactStore(artifactStore *codepipeline.ArtifactStore) []interface{} {
	if artifactStore == nil {
		return []interface{}{}
	}

	values := map[string]interface{}{}
	values["type"] = aws.StringValue(artifactStore.Type)
	values["location"] = aws.StringValue(artifactStore.Location)
	if artifactStore.EncryptionKey != nil {
		as := map[string]interface{}{
			"id":   aws.StringValue(artifactStore.EncryptionKey.Id),
			"type": aws.StringValue(artifactStore.EncryptionKey.Type),
		}
		values["encryption_key"] = []interface{}{as}
	}
	return []interface{}{values}
}

func flattenArtifactStores(artifactStores map[string]*codepipeline.ArtifactStore) []interface{} {
	values := []interface{}{}
	for region, artifactStore := range artifactStores {
		store := flattenArtifactStore(artifactStore)[0].(map[string]interface{})
		store["region"] = region
		values = append(values, store)
	}
	return values
}

func flattenStages(stages []*codepipeline.StageDeclaration, d *schema.ResourceData) []interface{} {
	stagesList := []interface{}{}
	for si, stage := range stages {
		values := map[string]interface{}{}
		values["name"] = aws.StringValue(stage.Name)
		values["action"] = flattenStageActions(si, stage.Actions, d)
		stagesList = append(stagesList, values)
	}
	return stagesList
}

func flattenStageActions(si int, actions []*codepipeline.ActionDeclaration, d *schema.ResourceData) []interface{} {
	actionsList := []interface{}{}
	for ai, action := range actions {
		values := map[string]interface{}{
			"category": aws.StringValue(action.ActionTypeId.Category),
			"owner":    aws.StringValue(action.ActionTypeId.Owner),
			"provider": aws.StringValue(action.ActionTypeId.Provider),
			"version":  aws.StringValue(action.ActionTypeId.Version),
			"name":     aws.StringValue(action.Name),
		}
		if action.Configuration != nil {
			config := aws.StringValueMap(action.Configuration)

			actionProvider := aws.StringValue(action.ActionTypeId.Provider)
			if actionProvider == providerGitHub {
				if _, ok := config[gitHubActionConfigurationOAuthToken]; ok {
					// The AWS API returns "****" for the OAuthToken value. Pull the value from the configuration.
					addr := fmt.Sprintf("stage.%d.action.%d.configuration.OAuthToken", si, ai)
					config[gitHubActionConfigurationOAuthToken] = d.Get(addr).(string)
				}
			}

			values["configuration"] = config
		}

		if len(action.OutputArtifacts) > 0 {
			values["output_artifacts"] = flattenActionsOutputArtifacts(action.OutputArtifacts)
		}

		if len(action.InputArtifacts) > 0 {
			values["input_artifacts"] = flattenActionsInputArtifacts(action.InputArtifacts)
		}

		if action.RoleArn != nil {
			values["role_arn"] = aws.StringValue(action.RoleArn)
		}

		if action.RunOrder != nil {
			values["run_order"] = int(aws.Int64Value(action.RunOrder))
		}

		if action.Region != nil {
			values["region"] = aws.StringValue(action.Region)
		}

		if action.Namespace != nil {
			values["namespace"] = aws.StringValue(action.Namespace)
		}

		actionsList = append(actionsList, values)
	}
	return actionsList
}

func flattenActionsOutputArtifacts(artifacts []*codepipeline.OutputArtifact) []string {
	values := []string{}
	for _, artifact := range artifacts {
		values = append(values, aws.StringValue(artifact.Name))
	}
	return values
}

func flattenActionsInputArtifacts(artifacts []*codepipeline.InputArtifact) []string {
	values := []string{}
	for _, artifact := range artifacts {
		values = append(values, aws.StringValue(artifact.Name))
	}
	return values
}

func resourceValidateActionProvider(i interface{}, path cty.Path) diag.Diagnostics {
	v, ok := i.(string)
	if !ok {
		return diag.Errorf("expected type to be string")
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

	return nil
}

func suppressStageActionConfiguration(k, old, new string, d *schema.ResourceData) bool {
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

const gitHubTokenHashPrefix = "hash-"

func hashGitHubToken(token string) string {
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
