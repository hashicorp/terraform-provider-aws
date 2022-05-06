package sagemaker

import (
	"fmt"
	"log"
	"regexp"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/sagemaker"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

func ResourceFeatureGroup() *schema.Resource {
	return &schema.Resource{
		Create: resourceFeatureGroupCreate,
		Read:   resourceFeatureGroupRead,
		Update: resourceFeatureGroupUpdate,
		Delete: resourceFeatureGroupDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"feature_group_name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
				ValidateFunc: validation.All(
					validation.StringLenBetween(1, 64),
					validation.StringMatch(regexp.MustCompile(`^[a-zA-Z0-9](-*[a-zA-Z0-9]){0,63}`),
						"Must start and end with an alphanumeric character and Can only contain alphanumeric character and hyphens. Spaces are not allowed."),
				),
			},
			"record_identifier_feature_name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
				ValidateFunc: validation.All(
					validation.StringLenBetween(1, 64),
					validation.StringMatch(regexp.MustCompile(`^[a-zA-Z0-9]([-_]*[a-zA-Z0-9]){0,63}`),
						"Must start and end with an alphanumeric character and Can only contains alphanumeric characters, hyphens, underscores. Spaces are not allowed."),
				),
			},
			"event_time_feature_name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
				ValidateFunc: validation.All(
					validation.StringLenBetween(1, 64),
					validation.StringMatch(regexp.MustCompile(`^[a-zA-Z0-9]([-_]*[a-zA-Z0-9]){0,63}`),
						"Must start and end with an alphanumeric character and Can only contains alphanumeric characters, hyphens, underscores. Spaces are not allowed."),
				),
			},
			"description": {
				Type:         schema.TypeString,
				Optional:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringLenBetween(0, 128),
			},
			"role_arn": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: verify.ValidARN,
			},
			"feature_definition": {
				Type:     schema.TypeList,
				Required: true,
				ForceNew: true,
				MinItems: 1,
				MaxItems: 2500,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"feature_name": {
							Type:     schema.TypeString,
							Optional: true,
							ValidateFunc: validation.All(
								validation.StringLenBetween(1, 64),
								validation.StringNotInSlice([]string{"is_deleted", "write_time", "api_invocation_time"}, false),
								validation.StringMatch(regexp.MustCompile(`^[a-zA-Z0-9]([-_]*[a-zA-Z0-9]){0,63}`),
									"Must start and end with an alphanumeric character and Can only contains alphanumeric characters, hyphens, underscores. Spaces are not allowed."),
							),
						},
						"feature_type": {
							Type:         schema.TypeString,
							Optional:     true,
							ValidateFunc: validation.StringInSlice(sagemaker.FeatureType_Values(), false),
						},
					},
				},
			},
			"offline_store_config": {
				Type:         schema.TypeList,
				Optional:     true,
				ForceNew:     true,
				MaxItems:     1,
				AtLeastOneOf: []string{"offline_store_config", "online_store_config"},
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"data_catalog_config": {
							Type:     schema.TypeList,
							Optional: true,
							Computed: true,
							MaxItems: 1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"catalog": {
										Type:     schema.TypeString,
										Optional: true,
										Computed: true,
									},
									"database": {
										Type:     schema.TypeString,
										Optional: true,
										Computed: true,
									},
									"table_name": {
										Type:     schema.TypeString,
										Optional: true,
										Computed: true,
									},
								},
							},
						},
						"s3_storage_config": {
							Type:     schema.TypeList,
							Required: true,
							MaxItems: 1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"kms_key_id": {
										Type:         schema.TypeString,
										Optional:     true,
										ValidateFunc: verify.ValidARN,
									},
									"s3_uri": {
										Type:     schema.TypeString,
										Required: true,
									},
								},
							},
						},
						"disable_glue_table_creation": {
							Type:     schema.TypeBool,
							Optional: true,
						},
					},
				},
			},
			"online_store_config": {
				Type:         schema.TypeList,
				Optional:     true,
				ForceNew:     true,
				MaxItems:     1,
				AtLeastOneOf: []string{"offline_store_config", "online_store_config"},
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"security_config": {
							Type:     schema.TypeList,
							Optional: true,
							MaxItems: 1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"kms_key_id": {
										Type:         schema.TypeString,
										Optional:     true,
										ValidateFunc: verify.ValidARN,
									},
								},
							},
						},
						"enable_online_store": {
							Type:     schema.TypeBool,
							Optional: true,
							Default:  false,
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

func resourceFeatureGroupCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).SageMakerConn
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	tags := defaultTagsConfig.MergeTags(tftags.New(d.Get("tags").(map[string]interface{})))

	name := d.Get("feature_group_name").(string)

	input := &sagemaker.CreateFeatureGroupInput{
		FeatureGroupName:            aws.String(name),
		EventTimeFeatureName:        aws.String(d.Get("event_time_feature_name").(string)),
		RecordIdentifierFeatureName: aws.String(d.Get("record_identifier_feature_name").(string)),
		RoleArn:                     aws.String(d.Get("role_arn").(string)),
		FeatureDefinitions:          expandFeatureGroupFeatureDefinition(d.Get("feature_definition").([]interface{})),
	}

	if v, ok := d.GetOk("description"); ok {
		input.Description = aws.String(v.(string))
	}

	if len(tags) > 0 {
		input.Tags = Tags(tags.IgnoreAWS())
	}

	if v, ok := d.GetOk("offline_store_config"); ok {
		input.OfflineStoreConfig = expandFeatureGroupOfflineStoreConfig(v.([]interface{}))
	}

	if v, ok := d.GetOk("online_store_config"); ok {
		input.OnlineStoreConfig = expandFeatureGroupOnlineStoreConfig(v.([]interface{}))
	}

	log.Printf("[DEBUG] SageMaker Feature Group create config: %#v", *input)
	err := resource.Retry(propagationTimeout, func() *resource.RetryError {
		_, err := conn.CreateFeatureGroup(input)
		if err != nil {
			if tfawserr.ErrMessageContains(err, "ValidationException", "The execution role ARN is invalid.") {
				return resource.RetryableError(err)
			}
			if tfawserr.ErrMessageContains(err, "ValidationException", "Invalid S3Uri provided") {
				return resource.RetryableError(err)
			}
			return resource.NonRetryableError(err)
		}

		return nil
	})
	if tfresource.TimedOut(err) {
		_, err = conn.CreateFeatureGroup(input)
	}

	if err != nil {
		return fmt.Errorf("Error creating SageMaker Feature Group: %w", err)
	}

	d.SetId(name)

	if _, err := WaitFeatureGroupCreated(conn, d.Id()); err != nil {
		return fmt.Errorf("error waiting for SageMaker Feature Group (%s) to create: %w", d.Id(), err)
	}

	return resourceFeatureGroupRead(d, meta)
}

func resourceFeatureGroupRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).SageMakerConn
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	ignoreTagsConfig := meta.(*conns.AWSClient).IgnoreTagsConfig

	output, err := FindFeatureGroupByName(conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] SageMaker Feature Group (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return fmt.Errorf("error reading SageMaker Feature Group (%s): %w", d.Id(), err)
	}

	arn := aws.StringValue(output.FeatureGroupArn)
	d.Set("feature_group_name", output.FeatureGroupName)
	d.Set("event_time_feature_name", output.EventTimeFeatureName)
	d.Set("description", output.Description)
	d.Set("record_identifier_feature_name", output.RecordIdentifierFeatureName)
	d.Set("role_arn", output.RoleArn)
	d.Set("arn", arn)

	if err := d.Set("feature_definition", flattenFeatureGroupFeatureDefinition(output.FeatureDefinitions)); err != nil {
		return fmt.Errorf("error setting feature_definition for SageMaker Feature Group (%s): %w", d.Id(), err)
	}

	if err := d.Set("online_store_config", flattenFeatureGroupOnlineStoreConfig(output.OnlineStoreConfig)); err != nil {
		return fmt.Errorf("error setting online_store_config for SageMaker Feature Group (%s): %w", d.Id(), err)
	}

	if err := d.Set("offline_store_config", flattenFeatureGroupOfflineStoreConfig(output.OfflineStoreConfig)); err != nil {
		return fmt.Errorf("error setting offline_store_config for SageMaker Feature Group (%s): %w", d.Id(), err)
	}

	tags, err := ListTags(conn, arn)
	if err != nil {
		return fmt.Errorf("error listing tags for SageMaker Feature Group (%s): %w", d.Id(), err)
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

func resourceFeatureGroupUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).SageMakerConn

	if d.HasChange("tags_all") {
		o, n := d.GetChange("tags_all")

		if err := UpdateTags(conn, d.Get("arn").(string), o, n); err != nil {
			return fmt.Errorf("error updating SageMaker Feature Group (%s) tags: %w", d.Id(), err)
		}
	}

	return resourceFeatureGroupRead(d, meta)
}

func resourceFeatureGroupDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).SageMakerConn

	input := &sagemaker.DeleteFeatureGroupInput{
		FeatureGroupName: aws.String(d.Id()),
	}

	if _, err := conn.DeleteFeatureGroup(input); err != nil {
		if tfawserr.ErrCodeEquals(err, sagemaker.ErrCodeResourceNotFound) {
			return nil
		}
		return fmt.Errorf("error deleting SageMaker Feature Group (%s): %w", d.Id(), err)
	}

	if _, err := WaitFeatureGroupDeleted(conn, d.Id()); err != nil {
		if tfawserr.ErrCodeEquals(err, sagemaker.ErrCodeResourceNotFound) {
			return nil
		}
		return fmt.Errorf("error waiting for SageMaker Feature Group (%s) to delete: %w", d.Id(), err)
	}

	return nil
}

func expandFeatureGroupFeatureDefinition(l []interface{}) []*sagemaker.FeatureDefinition {
	featureDefs := make([]*sagemaker.FeatureDefinition, 0, len(l))

	for _, lRaw := range l {
		data := lRaw.(map[string]interface{})

		featureDef := &sagemaker.FeatureDefinition{
			FeatureName: aws.String(data["feature_name"].(string)),
			FeatureType: aws.String(data["feature_type"].(string)),
		}

		featureDefs = append(featureDefs, featureDef)
	}

	return featureDefs
}

func flattenFeatureGroupFeatureDefinition(config []*sagemaker.FeatureDefinition) []map[string]interface{} {
	features := make([]map[string]interface{}, 0, len(config))

	for _, i := range config {
		feature := map[string]interface{}{
			"feature_name": aws.StringValue(i.FeatureName),
			"feature_type": aws.StringValue(i.FeatureType),
		}

		features = append(features, feature)
	}
	return features
}

func expandFeatureGroupOnlineStoreConfig(l []interface{}) *sagemaker.OnlineStoreConfig {
	if len(l) == 0 || l[0] == nil {
		return nil
	}

	m := l[0].(map[string]interface{})

	config := &sagemaker.OnlineStoreConfig{
		EnableOnlineStore: aws.Bool(m["enable_online_store"].(bool)),
	}

	if v, ok := m["security_config"].([]interface{}); ok && len(v) > 0 {
		config.SecurityConfig = expandFeatureGroupOnlineStoreConfigSecurityConfig(v)
	}

	return config
}

func flattenFeatureGroupOnlineStoreConfig(config *sagemaker.OnlineStoreConfig) []map[string]interface{} {
	if config == nil {
		return []map[string]interface{}{}
	}

	m := map[string]interface{}{
		"enable_online_store": aws.BoolValue(config.EnableOnlineStore),
	}

	if config.SecurityConfig != nil {
		m["security_config"] = flattenFeatureGroupOnlineStoreConfigSecurityConfig(config.SecurityConfig)
	}

	return []map[string]interface{}{m}
}

func expandFeatureGroupOnlineStoreConfigSecurityConfig(l []interface{}) *sagemaker.OnlineStoreSecurityConfig {
	if len(l) == 0 || l[0] == nil {
		return nil
	}

	m := l[0].(map[string]interface{})

	config := &sagemaker.OnlineStoreSecurityConfig{
		KmsKeyId: aws.String(m["kms_key_id"].(string)),
	}

	return config
}

func flattenFeatureGroupOnlineStoreConfigSecurityConfig(config *sagemaker.OnlineStoreSecurityConfig) []map[string]interface{} {
	if config == nil {
		return []map[string]interface{}{}
	}

	m := map[string]interface{}{
		"kms_key_id": aws.StringValue(config.KmsKeyId),
	}

	return []map[string]interface{}{m}
}

func expandFeatureGroupOfflineStoreConfig(l []interface{}) *sagemaker.OfflineStoreConfig {
	if len(l) == 0 || l[0] == nil {
		return nil
	}

	m := l[0].(map[string]interface{})

	config := &sagemaker.OfflineStoreConfig{}

	if v, ok := m["s3_storage_config"].([]interface{}); ok && len(v) > 0 {
		config.S3StorageConfig = expandFeatureGroupOfflineStoreConfigS3StorageConfig(v)
	}

	if v, ok := m["data_catalog_config"].([]interface{}); ok && len(v) > 0 {
		config.DataCatalogConfig = expandFeatureGroupOfflineStoreConfigDataCatalogConfig(v)
	}

	if v, ok := m["disable_glue_table_creation"].(bool); ok {
		config.DisableGlueTableCreation = aws.Bool(v)
	}

	return config
}

func flattenFeatureGroupOfflineStoreConfig(config *sagemaker.OfflineStoreConfig) []map[string]interface{} {
	if config == nil {
		return []map[string]interface{}{}
	}

	m := map[string]interface{}{
		"disable_glue_table_creation": aws.BoolValue(config.DisableGlueTableCreation),
	}

	if config.DataCatalogConfig != nil {
		m["data_catalog_config"] = flattenFeatureGroupOfflineStoreConfigDataCatalogConfig(config.DataCatalogConfig)
	}

	if config.S3StorageConfig != nil {
		m["s3_storage_config"] = flattenFeatureGroupOfflineStoreConfigS3StorageConfig(config.S3StorageConfig)
	}

	return []map[string]interface{}{m}
}

func expandFeatureGroupOfflineStoreConfigS3StorageConfig(l []interface{}) *sagemaker.S3StorageConfig {
	if len(l) == 0 || l[0] == nil {
		return nil
	}

	m := l[0].(map[string]interface{})

	config := &sagemaker.S3StorageConfig{
		S3Uri: aws.String(m["s3_uri"].(string)),
	}

	if v, ok := m["kms_key_id"].(string); ok && v != "" {
		config.KmsKeyId = aws.String(m["kms_key_id"].(string))
	}

	return config
}

func flattenFeatureGroupOfflineStoreConfigS3StorageConfig(config *sagemaker.S3StorageConfig) []map[string]interface{} {
	if config == nil {
		return []map[string]interface{}{}
	}

	m := map[string]interface{}{
		"s3_uri": aws.StringValue(config.S3Uri),
	}

	if config.KmsKeyId != nil {
		m["kms_key_id"] = aws.StringValue(config.KmsKeyId)
	}

	return []map[string]interface{}{m}
}

func expandFeatureGroupOfflineStoreConfigDataCatalogConfig(l []interface{}) *sagemaker.DataCatalogConfig {
	if len(l) == 0 || l[0] == nil {
		return nil
	}

	m := l[0].(map[string]interface{})

	config := &sagemaker.DataCatalogConfig{
		Catalog:   aws.String(m["catalog"].(string)),
		Database:  aws.String(m["database"].(string)),
		TableName: aws.String(m["table_name"].(string)),
	}

	return config
}

func flattenFeatureGroupOfflineStoreConfigDataCatalogConfig(config *sagemaker.DataCatalogConfig) []map[string]interface{} {
	if config == nil {
		return []map[string]interface{}{}
	}

	m := map[string]interface{}{
		"catalog":    aws.StringValue(config.Catalog),
		"database":   aws.StringValue(config.Database),
		"table_name": aws.StringValue(config.TableName),
	}

	return []map[string]interface{}{m}
}
