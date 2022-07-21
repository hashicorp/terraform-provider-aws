package comprehend

import (
	"context"
	"errors"
	"log"
	"regexp"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/aws/arn"
	"github.com/aws/aws-sdk-go-v2/service/comprehend"
	"github.com/aws/aws-sdk-go-v2/service/comprehend/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

func ResourceEntityRecognizer() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceEntityRecognizerCreate,
		ReadWithoutTimeout:   resourceEntityRecognizerRead,
		UpdateWithoutTimeout: resourceEntityRecognizerUpdate,
		DeleteWithoutTimeout: resourceEntityRecognizerDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(20 * time.Minute),
			Update: schema.DefaultTimeout(20 * time.Minute),
			Delete: schema.DefaultTimeout(20 * time.Minute),
		},

		Schema: map[string]*schema.Schema{
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"data_access_role_arn": {
				Type:     schema.TypeString,
				Required: true,
			},
			"input_data_config": {
				Type:     schema.TypeList,
				Required: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"annotations": {
							Type:     schema.TypeList,
							Optional: true,
							MaxItems: 1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"s3_uri": {
										Type:     schema.TypeString,
										Required: true,
									},
									"test_s3_uri": {
										Type:     schema.TypeString,
										Optional: true,
									},
								},
							},
						},
						"augmented_manifests": {
							Type:     schema.TypeList,
							Optional: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"annotation_data_s3_uri": {
										Type:     schema.TypeString,
										Optional: true,
									},
									"attribute_names": {
										Type:     schema.TypeList,
										Required: true,
										Elem:     &schema.Schema{Type: schema.TypeString},
									},
									"document_type": {
										Type:             schema.TypeString,
										Optional:         true,
										ValidateDiagFunc: enum.Validate[types.AugmentedManifestsDocumentTypeFormat](),
										Default:          types.AugmentedManifestsDocumentTypeFormatPlainTextDocument,
									},
									"s3_uri": {
										Type:     schema.TypeString,
										Required: true,
									},
									"source_documents_s3_uri": {
										Type:     schema.TypeString,
										Optional: true,
									},
									"split": {
										Type:             schema.TypeString,
										Optional:         true,
										ValidateDiagFunc: enum.Validate[types.Split](),
										Default:          types.SplitTrain,
									},
								},
							},
						},
						"data_format": {
							Type:             schema.TypeString,
							Optional:         true,
							ValidateDiagFunc: enum.Validate[types.EntityRecognizerDataFormat](),
							Default:          types.EntityRecognizerDataFormatComprehendCsv,
						},
						"documents": {
							Type:     schema.TypeList,
							Optional: true,
							MaxItems: 1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"input_format": {
										Type:             schema.TypeString,
										Optional:         true,
										ValidateDiagFunc: enum.Validate[types.InputFormat](),
										Default:          types.InputFormatOneDocPerLine,
									},
									"s3_uri": {
										Type:     schema.TypeString,
										Required: true,
									},
									"test_s3_uri": {
										Type:     schema.TypeString,
										Optional: true,
									},
								},
							},
						},
						"entity_list": {
							Type:     schema.TypeList,
							Optional: true,
							MaxItems: 1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"s3_uri": {
										Type:     schema.TypeString,
										Required: true,
									},
								},
							},
						},
						"entity_types": {
							Type:     schema.TypeSet,
							Required: true,
							MaxItems: 25,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"type": {
										Type:     schema.TypeString,
										Required: true,
										ValidateFunc: validation.All(
											validation.StringIsNotEmpty,
											validation.StringDoesNotContainAny("\n\r\t,"),
										),
									},
								},
							},
						},
					},
				},
			},
			"language_code": {
				Type:             schema.TypeString,
				Required:         true,
				ValidateDiagFunc: enum.Validate[types.SyntaxLanguageCode](),
			},
			"model_kms_key_id": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"name": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validModelName,
			},
			"tags":     tftags.TagsSchema(),
			"tags_all": tftags.TagsSchemaComputed(),
			"version_name": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validModelVersionName,
			},
			"volume_kms_key_id": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"vpc_config": {
				Type:     schema.TypeList,
				Optional: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"security_group_ids": {
							Type:     schema.TypeList,
							Required: true,
							Elem:     &schema.Schema{Type: schema.TypeString},
						},
						"subnets": {
							Type:     schema.TypeList,
							Required: true,
							Elem:     &schema.Schema{Type: schema.TypeString},
						},
					},
				},
			},
		},

		CustomizeDiff: verify.SetTagsDiff,
	}
}

func resourceEntityRecognizerCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).ComprehendConn

	in := &comprehend.CreateEntityRecognizerInput{
		DataAccessRoleArn:  aws.String(d.Get("data_access_role_arn").(string)),
		InputDataConfig:    expandInputDataConfig(d.Get("input_data_config").([]interface{})),
		LanguageCode:       types.LanguageCode(d.Get("language_code").(string)),
		RecognizerName:     aws.String(d.Get("name").(string)),
		VpcConfig:          expandVPCConfig(d.Get("vpc_config").([]interface{})),
		ClientRequestToken: aws.String(resource.UniqueId()),
	}

	if v, ok := d.Get("model_kms_key_id").(string); ok && v != "" {
		in.ModelKmsKeyId = aws.String(v)
	}

	if v, ok := d.Get("version_name").(string); ok && v != "" {
		in.VersionName = aws.String(v)
	}

	if v, ok := d.Get("volume_kms_key_id").(string); ok && v != "" {
		in.VolumeKmsKeyId = aws.String(v)
	}

	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	tags := defaultTagsConfig.MergeTags(tftags.New(d.Get("tags").(map[string]interface{})))

	if len(tags) > 0 {
		in.Tags = Tags(tags.IgnoreAWS())
	}

	// Because the IAM credentials aren't evaluated until later, we need to ensure we wait for the IAM propagation delay
	time.Sleep(iamPropagationTimeout)

	out, err := conn.CreateEntityRecognizer(ctx, in)
	if err != nil {
		return diag.Errorf("creating Amazon Comprehend Entity Recognizer (%s): %s", d.Get("name").(string), err)
	}

	if out == nil || out.EntityRecognizerArn == nil {
		return diag.Errorf("creating Amazon Comprehend Entity Recognizer (%s): empty output", d.Get("name").(string))
	}

	d.SetId(aws.ToString(out.EntityRecognizerArn))

	if _, err := waitEntityRecognizerCreated(ctx, conn, d.Id(), d.Timeout(schema.TimeoutCreate)); err != nil {
		return diag.Errorf("waiting for Amazon Comprehend Entity Recognizer (%s) to be created: %s", d.Id(), err)
	}

	return resourceEntityRecognizerRead(ctx, d, meta)
}

func resourceEntityRecognizerRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).ComprehendConn

	out, err := FindEntityRecognizerByID(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] Comprehend Entity Recognizer (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return diag.Errorf("reading Comprehend Entity Recognizer (%s): %s", d.Id(), err)
	}

	d.Set("arn", out.EntityRecognizerArn)
	d.Set("data_access_role_arn", out.DataAccessRoleArn)
	d.Set("language_code", out.LanguageCode)
	d.Set("model_kms_key_id", out.ModelKmsKeyId)
	d.Set("version_name", out.VersionName)
	d.Set("volume_kms_key_id", out.VolumeKmsKeyId)

	// DescribeEntityRecognizer() doesn't return the model name
	arn, err := arn.Parse(aws.ToString(out.EntityRecognizerArn))
	if err != nil {
		return diag.Errorf("reading Comprehend Entity Recognizer (%s): %s", d.Id(), err)
	}
	re := regexp.MustCompile(`entity-recognizer/(.*)`)
	name := re.FindStringSubmatch(arn.Resource)[1]
	d.Set("name", name)

	if err := d.Set("input_data_config", flattenInputDataConfig(out.InputDataConfig)); err != nil {
		return diag.Errorf("setting input_data_config: %s", err)
	}

	if err := d.Set("vpc_config", flattenVPCConfig(out.VpcConfig)); err != nil {
		return diag.Errorf("setting vpc_config: %s", err)
	}

	tags, err := ListTags(ctx, conn, d.Id())
	if err != nil {
		return diag.Errorf("listing tags for Comprehend Entity Recognizer (%s): %s", d.Id(), err)
	}

	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	ignoreTagsConfig := meta.(*conns.AWSClient).IgnoreTagsConfig
	tags = tags.IgnoreAWS().IgnoreConfig(ignoreTagsConfig)

	if err := d.Set("tags", tags.RemoveDefaultConfig(defaultTagsConfig).Map()); err != nil {
		return diag.Errorf("setting tags: %s", err)
	}

	if err := d.Set("tags_all", tags.Map()); err != nil {
		return diag.Errorf("setting tags_all: %s", err)
	}

	return nil
}

func resourceEntityRecognizerUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).ComprehendConn

	if d.HasChange("tags_all") {
		o, n := d.GetChange("tags_all")

		if err := UpdateTags(ctx, conn, d.Id(), o, n); err != nil {
			return diag.Errorf("updating tags for Comprehend Entity Recognizer (%s): %s", d.Id(), err)
		}
	}

	return resourceEntityRecognizerRead(ctx, d, meta)
}

func resourceEntityRecognizerDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).ComprehendConn

	log.Printf("[INFO] Stopping Comprehend Entity Recognizer (%s)", d.Id())

	_, err := conn.StopTrainingEntityRecognizer(ctx, &comprehend.StopTrainingEntityRecognizerInput{
		EntityRecognizerArn: aws.String(d.Id()),
	})
	if err != nil {
		var nfe *types.ResourceNotFoundException
		if errors.As(err, &nfe) {
			return nil
		}

		return diag.Errorf("stopping Comprehend Entity Recognizer (%s): %s", d.Id(), err)
	}

	if _, err := waitEntityRecognizerStopped(ctx, conn, d.Id(), d.Timeout(schema.TimeoutDelete)); err != nil {
		var nfe *types.ResourceNotFoundException
		if errors.As(err, &nfe) {
			return nil
		}

		return diag.Errorf("waiting for Comprehend Entity Recognizer (%s) to be deleted: %s", d.Id(), err)
	}

	log.Printf("[INFO] Deleting Comprehend Entity Recognizer (%s)", d.Id())

	_, err = conn.DeleteEntityRecognizer(ctx, &comprehend.DeleteEntityRecognizerInput{
		EntityRecognizerArn: aws.String(d.Id()),
	})
	if err != nil {
		var nfe *types.ResourceNotFoundException
		if errors.As(err, &nfe) {
			return nil
		}

		return diag.Errorf("deleting Comprehend Entity Recognizer (%s): %s", d.Id(), err)
	}

	if _, err := waitEntityRecognizerDeleted(ctx, conn, d.Id(), d.Timeout(schema.TimeoutDelete)); err != nil {
		return diag.Errorf("waiting for Comprehend Entity Recognizer (%s) to be deleted: %s", d.Id(), err)
	}

	return nil
}

func waitEntityRecognizerCreated(ctx context.Context, conn *comprehend.Client, id string, timeout time.Duration) (*types.EntityRecognizerProperties, error) {
	stateConf := &resource.StateChangeConf{
		Pending: enum.Slice(types.ModelStatusSubmitted, types.ModelStatusTraining),
		Target:  enum.Slice(types.ModelStatusTrained),
		Refresh: statusEntityRecognizer(ctx, conn, id),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)
	if out, ok := outputRaw.(*types.EntityRecognizerProperties); ok {
		return out, err
	}

	return nil, err
}

func waitEntityRecognizerStopped(ctx context.Context, conn *comprehend.Client, id string, timeout time.Duration) (*types.EntityRecognizerProperties, error) {
	stateConf := &resource.StateChangeConf{
		Pending: enum.Slice(types.ModelStatusSubmitted, types.ModelStatusTraining, types.ModelStatusDeleting, types.ModelStatusStopRequested),
		Target:  enum.Slice(types.ModelStatusTrained, types.ModelStatusStopped, types.ModelStatusInError),
		Refresh: statusEntityRecognizer(ctx, conn, id),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)
	if out, ok := outputRaw.(*types.EntityRecognizerProperties); ok {
		return out, err
	}

	return nil, err
}

func waitEntityRecognizerDeleted(ctx context.Context, conn *comprehend.Client, id string, timeout time.Duration) (*types.EntityRecognizerProperties, error) {
	stateConf := &resource.StateChangeConf{
		Pending: enum.Slice(types.ModelStatusSubmitted, types.ModelStatusTraining, types.ModelStatusDeleting, types.ModelStatusInError, types.ModelStatusStopRequested),
		Target:  []string{},
		Refresh: statusEntityRecognizer(ctx, conn, id),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)
	if out, ok := outputRaw.(*types.EntityRecognizerProperties); ok {
		return out, err
	}

	return nil, err
}

func statusEntityRecognizer(ctx context.Context, conn *comprehend.Client, id string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		out, err := FindEntityRecognizerByID(ctx, conn, id)
		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return out, string(out.Status), nil
	}
}

func FindEntityRecognizerByID(ctx context.Context, conn *comprehend.Client, id string) (*types.EntityRecognizerProperties, error) {
	in := &comprehend.DescribeEntityRecognizerInput{
		EntityRecognizerArn: aws.String(id),
	}

	out, err := conn.DescribeEntityRecognizer(ctx, in)
	if err != nil {
		var nfe *types.ResourceNotFoundException
		if errors.As(err, &nfe) {
			return nil, &resource.NotFoundError{
				LastError:   err,
				LastRequest: in,
			}
		}

		return nil, err
	}

	if out == nil || out.EntityRecognizerProperties == nil {
		return nil, tfresource.NewEmptyResultError(in)
	}

	return out.EntityRecognizerProperties, nil
}

func flattenInputDataConfig(apiObject *types.EntityRecognizerInputDataConfig) []interface{} {
	if apiObject == nil {
		return nil
	}

	m := map[string]interface{}{
		"entity_types":        flattenEntityTypes(apiObject.EntityTypes),
		"annotations":         flattenAnnotations(apiObject.Annotations),
		"augmented_manifests": flattenAugmentedManifests(apiObject.AugmentedManifests),
		"data_format":         apiObject.DataFormat,
		"documents":           flattenDocuments(apiObject.Documents),
		"entity_list":         flattenEntityList(apiObject.EntityList),
	}

	return []interface{}{m}
}

func flattenEntityTypes(apiObjects []types.EntityTypesListItem) []interface{} {
	if len(apiObjects) == 0 {
		return nil
	}

	var l []interface{}

	for _, apiObject := range apiObjects {
		l = append(l, flattenEntityTypesListItem(&apiObject))
	}

	return l
}

func flattenEntityTypesListItem(apiObject *types.EntityTypesListItem) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	m := map[string]interface{}{
		"type": aws.ToString(apiObject.Type),
	}

	return m
}

func flattenAnnotations(apiObject *types.EntityRecognizerAnnotations) []interface{} {
	if apiObject == nil {
		return nil
	}

	m := map[string]interface{}{
		"s3_uri": aws.ToString(apiObject.S3Uri),
	}

	if v := apiObject.TestS3Uri; v != nil {
		m["test_s3_uri"] = aws.ToString(v)
	}

	return []interface{}{m}
}

func flattenAugmentedManifests(apiObjects []types.AugmentedManifestsListItem) []interface{} {
	if len(apiObjects) == 0 {
		return nil
	}

	var l []interface{}

	for _, apiObject := range apiObjects {
		l = append(l, flattenAugmentedManifestsListItem(&apiObject))
	}

	return l
}

func flattenAugmentedManifestsListItem(apiObject *types.AugmentedManifestsListItem) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	m := map[string]interface{}{
		"attribute_names": FlattenStringList(apiObject.AttributeNames),
		"s3_uri":          aws.ToString(apiObject.S3Uri),
		"document_type":   apiObject.DocumentType,
		"split":           apiObject.Split,
	}

	if v := apiObject.AnnotationDataS3Uri; v != nil {
		m["annotation_data_s3_uri"] = aws.ToString(v)
	}

	if v := apiObject.SourceDocumentsS3Uri; v != nil {
		m["source_documents_s3_uri"] = aws.ToString(v)
	}

	return m
}

func flattenDocuments(apiObject *types.EntityRecognizerDocuments) []interface{} {
	if apiObject == nil {
		return nil
	}

	m := map[string]interface{}{
		"s3_uri":       aws.ToString(apiObject.S3Uri),
		"input_format": apiObject.InputFormat,
	}

	if v := apiObject.TestS3Uri; v != nil {
		m["test_s3_uri"] = aws.ToString(v)
	}

	return []interface{}{m}
}

func flattenEntityList(apiObject *types.EntityRecognizerEntityList) []interface{} {
	if apiObject == nil {
		return nil
	}

	m := map[string]interface{}{
		"s3_uri": aws.ToString(apiObject.S3Uri),
	}

	return []interface{}{m}
}

func flattenVPCConfig(apiObject *types.VpcConfig) []interface{} {
	if apiObject == nil {
		return nil
	}

	m := map[string]interface{}{
		"security_group_ids": apiObject.SecurityGroupIds,
		"subnets":            apiObject.Subnets,
	}

	return []interface{}{m}
}

func expandInputDataConfig(tfList []interface{}) *types.EntityRecognizerInputDataConfig {
	if len(tfList) == 0 {
		return nil
	}

	tfMap := tfList[0].(map[string]interface{})

	a := &types.EntityRecognizerInputDataConfig{
		EntityTypes:        expandEntityTypes(tfMap["entity_types"].(*schema.Set)),
		Annotations:        expandAnnotations(tfMap["annotations"].([]interface{})),
		AugmentedManifests: expandAugmentedManifests(tfMap["augmented_manifests"].([]interface{})),
		DataFormat:         types.EntityRecognizerDataFormat(tfMap["data_format"].(string)),
		Documents:          expandDocuments(tfMap["documents"].([]interface{})),
		EntityList:         expandEntityList(tfMap["entity_list"].([]interface{})),
	}

	return a
}

func expandEntityTypes(tfSet *schema.Set) []types.EntityTypesListItem {
	if tfSet.Len() == 0 {
		return nil
	}

	var s []types.EntityTypesListItem

	for _, r := range tfSet.List() {
		m, ok := r.(map[string]interface{})
		if !ok {
			continue
		}

		a := expandEntityTypesListItem(m)
		if a == nil {
			continue
		}

		s = append(s, *a)
	}

	return s
}

func expandEntityTypesListItem(tfMap map[string]interface{}) *types.EntityTypesListItem {
	if tfMap == nil {
		return nil
	}

	a := &types.EntityTypesListItem{}

	if v, ok := tfMap["type"].(string); ok && v != "" {
		a.Type = aws.String(v)
	}

	return a
}

func expandAnnotations(tfList []interface{}) *types.EntityRecognizerAnnotations {
	if len(tfList) == 0 {
		return nil
	}

	tfMap := tfList[0].(map[string]interface{})

	a := &types.EntityRecognizerAnnotations{
		S3Uri: aws.String(tfMap["s3_uri"].(string)),
	}

	if v, ok := tfMap["test_s3_uri"].(string); ok && v != "" {
		a.TestS3Uri = aws.String(v)
	}

	return a
}

func expandAugmentedManifests(tfList []interface{}) []types.AugmentedManifestsListItem {
	if len(tfList) == 0 {
		return nil
	}

	var s []types.AugmentedManifestsListItem

	for _, r := range tfList {
		m, ok := r.(map[string]interface{})

		if !ok {
			continue
		}

		a := expandAugmentedManifestsListItem(m)

		if a == nil {
			continue
		}

		s = append(s, *a)
	}

	return s
}

func expandAugmentedManifestsListItem(tfMap map[string]interface{}) *types.AugmentedManifestsListItem {
	if tfMap == nil {
		return nil
	}

	a := &types.AugmentedManifestsListItem{
		AttributeNames: ExpandStringList(tfMap["attribute_names"].([]interface{})),
		S3Uri:          aws.String(tfMap["s3_uri"].(string)),
		DocumentType:   types.AugmentedManifestsDocumentTypeFormat(tfMap["document_type"].(string)),
		Split:          types.Split(tfMap["split"].(string)),
	}

	if v, ok := tfMap["annotation_data_s3_uri"].(string); ok && v != "" {
		a.AnnotationDataS3Uri = aws.String(v)
	}

	if v, ok := tfMap["source_documents_s3_uri"].(string); ok && v != "" {
		a.SourceDocumentsS3Uri = aws.String(v)
	}

	return a
}

func expandDocuments(tfList []interface{}) *types.EntityRecognizerDocuments {
	if len(tfList) == 0 {
		return nil
	}

	tfMap := tfList[0].(map[string]interface{})

	a := &types.EntityRecognizerDocuments{
		S3Uri:       aws.String(tfMap["s3_uri"].(string)),
		InputFormat: types.InputFormat(tfMap["input_format"].(string)),
	}

	if v, ok := tfMap["test_s3_uri"].(string); ok && v != "" {
		a.TestS3Uri = aws.String(v)
	}

	return a
}

func expandEntityList(tfList []interface{}) *types.EntityRecognizerEntityList {
	if len(tfList) == 0 {
		return nil
	}

	tfMap := tfList[0].(map[string]interface{})

	a := &types.EntityRecognizerEntityList{
		S3Uri: aws.String(tfMap["s3_uri"].(string)),
	}

	return a
}

func expandVPCConfig(tfList []interface{}) *types.VpcConfig {
	if len(tfList) == 0 {
		return nil
	}

	tfMap := tfList[0].(map[string]interface{})

	a := &types.VpcConfig{
		SecurityGroupIds: ExpandStringList(tfMap["security_group_ids"].([]interface{})),
		Subnets:          ExpandStringList(tfMap["subnets"].([]interface{})),
	}

	return a
}
