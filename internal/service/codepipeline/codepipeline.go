package codepipeline

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
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
	"github.com/hashicorp/terraform-provider-aws/names"
)

const (
	providerGitHub = "GitHub"

	gitHubActionConfigurationOAuthToken = "OAuthToken"
)

func ResourceCodePipeline() *schema.Resource { // nosemgrep:codepipeline-in-func-name
	return &schema.Resource{
		Create: resourceCreate,
		Read:   resourceRead,
		Update: resourceUpdate,
		Delete: resourceDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
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
			"artifact_store": {
				Type:     schema.TypeSet,
				Required: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"location": {
							Type:     schema.TypeString,
							Required: true,
						},
						"type": {
							Type:         schema.TypeString,
							Required:     true,
							ValidateFunc: validation.StringInSlice(codepipeline.ArtifactStoreType_Values(), false),
						},
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
						"region": {
							Type:     schema.TypeString,
							Optional: true,
							Computed: true,
						},
					},
				},
			},
			"stage": {
				Type:     schema.TypeList,
				MinItems: 2,
				Required: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"name": {
							Type:     schema.TypeString,
							Required: true,
							ValidateFunc: validation.All(
								validation.StringLenBetween(1, 100),
								validation.StringMatch(regexp.MustCompile(`[A-Za-z0-9.@\-_]+`), ""),
							),
						},
						"action": {
							Type:     schema.TypeList,
							Required: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"configuration": {
										Type:     schema.TypeMap,
										Optional: true,
										ValidateDiagFunc: allDiagFunc(
											validation.MapKeyLenBetween(1, 50),
											validation.MapKeyLenBetween(1, 1000),
										),
										Elem:             &schema.Schema{Type: schema.TypeString},
										DiffSuppressFunc: suppressStageActionConfiguration,
									},
									"category": {
										Type:         schema.TypeString,
										Required:     true,
										ValidateFunc: validation.StringInSlice(codepipeline.ActionCategory_Values(), false),
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
									"version": {
										Type:     schema.TypeString,
										Required: true,
										ValidateFunc: validation.All(
											validation.StringLenBetween(1, 9),
											validation.StringMatch(regexp.MustCompile(`[0-9A-Za-z_-]+`), ""),
										),
									},
									"input_artifacts": {
										Type:     schema.TypeList,
										Optional: true,
										Elem:     &schema.Schema{Type: schema.TypeString},
									},
									"output_artifacts": {
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
									"region": {
										Type:     schema.TypeString,
										Optional: true,
										Computed: true,
									},
									"namespace": {
										Type:     schema.TypeString,
										Optional: true,
										ValidateFunc: validation.All(
											validation.StringLenBetween(1, 100),
											validation.StringMatch(regexp.MustCompile(`[A-Za-z0-9@\-_]+`), ""),
										),
									},
								},
							},
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

func resourceCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).CodePipelineConn
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	tags := defaultTagsConfig.MergeTags(tftags.New(d.Get("tags").(map[string]interface{})))

	pipeline, err := expand(d)
	if err != nil {
		return err
	}
	params := &codepipeline.CreatePipelineInput{
		Pipeline: pipeline,
		Tags:     Tags(tags.IgnoreAWS()),
	}

	var resp *codepipeline.CreatePipelineOutput
	err = resource.Retry(propagationTimeout, func() *resource.RetryError {
		var err error

		resp, err = conn.CreatePipeline(params)

		if tfawserr.ErrMessageContains(err, codepipeline.ErrCodeInvalidStructureException, "not authorized") {
			return resource.RetryableError(err)
		}

		if err != nil {
			return resource.NonRetryableError(err)
		}

		return nil
	})
	if tfresource.TimedOut(err) {
		resp, err = conn.CreatePipeline(params)
	}
	if err != nil {
		return fmt.Errorf("Error creating CodePipeline: %w", err)
	}
	if resp.Pipeline == nil {
		return fmt.Errorf("Error creating CodePipeline: invalid response from AWS")
	}

	d.SetId(aws.StringValue(resp.Pipeline.Name))

	return resourceRead(d, meta)
}

func expand(d *schema.ResourceData) (*codepipeline.PipelineDeclaration, error) {
	pipeline := codepipeline.PipelineDeclaration{
		Name:    aws.String(d.Get("name").(string)),
		RoleArn: aws.String(d.Get("role_arn").(string)),
		Stages:  expandStages(d),
	}

	pipelineArtifactStores, err := ExpandArtifactStores(d.Get("artifact_store").(*schema.Set).List())
	if err != nil {
		return nil, err
	}
	if len(pipelineArtifactStores) == 1 {
		for _, v := range pipelineArtifactStores {
			pipeline.ArtifactStore = v
		}
	} else {
		pipeline.ArtifactStores = pipelineArtifactStores
	}

	return &pipeline, nil
}

func ExpandArtifactStores(configs []interface{}) (map[string]*codepipeline.ArtifactStore, error) {
	if len(configs) == 0 {
		return nil, nil
	}

	regions := make([]string, 0, len(configs))
	pipelineArtifactStores := make(map[string]*codepipeline.ArtifactStore)
	for _, config := range configs {
		region, store := expandArtifactStoreData(config.(map[string]interface{}))
		regions = append(regions, region)
		pipelineArtifactStores[region] = store
	}

	if len(regions) == 1 {
		if regions[0] != "" {
			return nil, errors.New("region cannot be set for a single-region CodePipeline")
		}
	} else {
		for _, v := range regions {
			if v == "" {
				return nil, errors.New("region must be set for a cross-region CodePipeline")
			}
		}
		if len(configs) != len(pipelineArtifactStores) {
			return nil, errors.New("only one Artifact Store can be defined per region for a cross-region CodePipeline")
		}
	}

	return pipelineArtifactStores, nil
}

func expandArtifactStoreData(data map[string]interface{}) (string, *codepipeline.ArtifactStore) {
	pipelineArtifactStore := codepipeline.ArtifactStore{
		Location: aws.String(data["location"].(string)),
		Type:     aws.String(data["type"].(string)),
	}
	tek := data["encryption_key"].([]interface{})
	if len(tek) > 0 {
		vk := tek[0].(map[string]interface{})
		ek := codepipeline.EncryptionKey{
			Type: aws.String(vk["type"].(string)),
			Id:   aws.String(vk["id"].(string)),
		}
		pipelineArtifactStore.EncryptionKey = &ek
	}

	return data["region"].(string), &pipelineArtifactStore
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

func expandStages(d *schema.ResourceData) []*codepipeline.StageDeclaration {
	stages := d.Get("stage").([]interface{})
	pipelineStages := []*codepipeline.StageDeclaration{}

	for _, stage := range stages {
		data := stage.(map[string]interface{})
		a := data["action"].([]interface{})
		actions := expandActions(a)
		pipelineStages = append(pipelineStages, &codepipeline.StageDeclaration{
			Name:    aws.String(data["name"].(string)),
			Actions: actions,
		})
	}
	return pipelineStages
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

func expandActions(a []interface{}) []*codepipeline.ActionDeclaration {
	actions := []*codepipeline.ActionDeclaration{}
	for _, config := range a {
		data := config.(map[string]interface{})

		conf := flex.ExpandStringMap(data["configuration"].(map[string]interface{}))

		action := codepipeline.ActionDeclaration{
			ActionTypeId: &codepipeline.ActionTypeId{
				Category: aws.String(data["category"].(string)),
				Owner:    aws.String(data["owner"].(string)),

				Provider: aws.String(data["provider"].(string)),
				Version:  aws.String(data["version"].(string)),
			},
			Name:          aws.String(data["name"].(string)),
			Configuration: conf,
		}

		oa := data["output_artifacts"].([]interface{})
		if len(oa) > 0 {
			outputArtifacts := expandActionsOutputArtifacts(oa)
			action.OutputArtifacts = outputArtifacts

		}
		ia := data["input_artifacts"].([]interface{})
		if len(ia) > 0 {
			inputArtifacts := expandActionsInputArtifacts(ia)
			action.InputArtifacts = inputArtifacts

		}
		ra := data["role_arn"].(string)
		if ra != "" {
			action.RoleArn = aws.String(ra)
		}
		ro := data["run_order"].(int)
		if ro > 0 {
			action.RunOrder = aws.Int64(int64(ro))
		}
		r := data["region"].(string)
		if r != "" {
			action.Region = aws.String(r)
		}
		ns := data["namespace"].(string)
		if len(ns) > 0 {
			action.Namespace = aws.String(ns)
		}
		actions = append(actions, &action)
	}
	return actions
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

func expandActionsOutputArtifacts(s []interface{}) []*codepipeline.OutputArtifact {
	outputArtifacts := []*codepipeline.OutputArtifact{}
	for _, artifact := range s {
		if artifact == nil {
			continue
		}
		outputArtifacts = append(outputArtifacts, &codepipeline.OutputArtifact{
			Name: aws.String(artifact.(string)),
		})
	}
	return outputArtifacts
}

func flattenActionsOutputArtifacts(artifacts []*codepipeline.OutputArtifact) []string {
	values := []string{}
	for _, artifact := range artifacts {
		values = append(values, aws.StringValue(artifact.Name))
	}
	return values
}

func expandActionsInputArtifacts(s []interface{}) []*codepipeline.InputArtifact {
	outputArtifacts := []*codepipeline.InputArtifact{}
	for _, artifact := range s {
		if artifact == nil {
			continue
		}
		outputArtifacts = append(outputArtifacts, &codepipeline.InputArtifact{
			Name: aws.String(artifact.(string)),
		})
	}
	return outputArtifacts
}

func flattenActionsInputArtifacts(artifacts []*codepipeline.InputArtifact) []string {
	values := []string{}
	for _, artifact := range artifacts {
		values = append(values, aws.StringValue(artifact.Name))
	}
	return values
}

func resourceRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).CodePipelineConn
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	ignoreTagsConfig := meta.(*conns.AWSClient).IgnoreTagsConfig

	resp, err := conn.GetPipeline(&codepipeline.GetPipelineInput{
		Name: aws.String(d.Id()),
	})

	if !d.IsNewResource() && tfawserr.ErrCodeEquals(err, codepipeline.ErrCodePipelineNotFoundException) {
		names.LogNotFoundRemoveState(names.CodePipeline, names.ErrActionReading, resPipeline, d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return names.Error(names.CodePipeline, names.ErrActionReading, resPipeline, d.Id(), err)
	}

	metadata := resp.Metadata
	pipeline := resp.Pipeline

	if pipeline.ArtifactStore != nil {
		if err := d.Set("artifact_store", flattenArtifactStore(pipeline.ArtifactStore)); err != nil {
			return err
		}
	} else if pipeline.ArtifactStores != nil {
		if err := d.Set("artifact_store", flattenArtifactStores(pipeline.ArtifactStores)); err != nil {
			return err
		}
	}

	if err := d.Set("stage", flattenStages(pipeline.Stages, d)); err != nil {
		return err
	}

	arn := aws.StringValue(metadata.PipelineArn)
	d.Set("arn", arn)
	d.Set("name", pipeline.Name)
	d.Set("role_arn", pipeline.RoleArn)

	tags, err := ListTags(conn, arn)

	if err != nil {
		return fmt.Errorf("error listing tags for CodePipeline (%s): %w", arn, err)
	}

	tags = tags.IgnoreAWS().IgnoreConfig(ignoreTagsConfig)

	//lintignore:AWSR002
	if err := d.Set("tags", tags.RemoveDefaultConfig(defaultTagsConfig).Map()); err != nil {
		return fmt.Errorf("error setting tags: %w", err)
	}

	if err := d.Set("tags_all", tags.Map()); err != nil {
		return fmt.Errorf("error setting tags_all: %w", err)
	}

	return nil
}

func resourceUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).CodePipelineConn

	if d.HasChangesExcept("tags", "tags_all") {
		pipeline, err := expand(d)
		if err != nil {
			return err
		}
		params := &codepipeline.UpdatePipelineInput{
			Pipeline: pipeline,
		}
		_, err = conn.UpdatePipeline(params)

		if err != nil {
			return fmt.Errorf("[ERROR] Error updating CodePipeline (%s): %w", d.Id(), err)
		}
	}

	arn := d.Get("arn").(string)
	if d.HasChange("tags_all") {
		o, n := d.GetChange("tags_all")

		if err := UpdateTags(conn, arn, o, n); err != nil {
			return fmt.Errorf("error updating CodePipeline (%s) tags: %w", arn, err)
		}
	}

	return resourceRead(d, meta)
}

func resourceDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).CodePipelineConn

	_, err := conn.DeletePipeline(&codepipeline.DeletePipelineInput{
		Name: aws.String(d.Id()),
	})

	if tfawserr.ErrCodeEquals(err, codepipeline.ErrCodePipelineNotFoundException) {
		return nil
	}

	if err != nil {
		return fmt.Errorf("error deleting CodePipeline (%s): %w", d.Id(), err)
	}

	return err
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

// https://github.com/hashicorp/terraform-plugin-sdk/issues/780.
func allDiagFunc(validators ...schema.SchemaValidateDiagFunc) schema.SchemaValidateDiagFunc {
	return func(i interface{}, k cty.Path) diag.Diagnostics {
		var diags diag.Diagnostics
		for _, validator := range validators {
			diags = append(diags, validator(i, k)...)
		}
		return diags
	}
}
