// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package sagemaker

import (
	"context"
	"log"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/sagemaker"
	awstypes "github.com/aws/aws-sdk-go-v2/service/sagemaker/types"
	"github.com/hashicorp/aws-sdk-go-base/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_sagemaker_feature_group", name="Feature Group")
// @Tags(identifierAttribute="arn")
func resourceFeatureGroup() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceFeatureGroupCreate,
		ReadWithoutTimeout:   resourceFeatureGroupRead,
		UpdateWithoutTimeout: resourceFeatureGroupUpdate,
		DeleteWithoutTimeout: resourceFeatureGroupDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			names.AttrARN: {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrDescription: {
				Type:         schema.TypeString,
				Optional:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringLenBetween(0, 128),
			},
			"event_time_feature_name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
				ValidateFunc: validation.All(
					validation.StringLenBetween(1, 64),
					validation.StringMatch(regexache.MustCompile(`^[0-9A-Za-z]([-_]*[0-9A-Za-z]){0,63}`),
						"Must start and end with an alphanumeric character and Can only contains alphanumeric characters, hyphens, underscores. Spaces are not allowed."),
				),
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
							ForceNew: true,
							ValidateFunc: validation.All(
								validation.StringLenBetween(1, 64),
								validation.StringNotInSlice([]string{"is_deleted", "write_time", "api_invocation_time"}, false),
								validation.StringMatch(regexache.MustCompile(`^[0-9A-Za-z]([-_]*[0-9A-Za-z]){0,63}`),
									"Must start and end with an alphanumeric character and Can only contains alphanumeric characters, hyphens, underscores. Spaces are not allowed."),
							),
						},
						"feature_type": {
							Type:             schema.TypeString,
							Optional:         true,
							ForceNew:         true,
							ValidateDiagFunc: enum.Validate[awstypes.FeatureType](),
						},
						"collection_config": {
							Type:     schema.TypeList,
							Optional: true,
							ForceNew: true,
							MaxItems: 1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"vector_config": {
										Type:     schema.TypeList,
										Optional: true,
										ForceNew: true,
										MaxItems: 1,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"dimension": {
													Type:         schema.TypeInt,
													Optional:     true,
													ForceNew:     true,
													ValidateFunc: validation.IntBetween(1, 8192),
												},
											},
										},
									},
								},
							},
						},
						"collection_type": {
							Type:             schema.TypeString,
							Optional:         true,
							ForceNew:         true,
							ValidateDiagFunc: enum.Validate[awstypes.CollectionType](),
						},
					},
				},
			},
			"feature_group_name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
				ValidateFunc: validation.All(
					validation.StringLenBetween(1, 64),
					validation.StringMatch(regexache.MustCompile(`^[0-9A-Za-z](-*[0-9A-Za-z]){0,63}`),
						"Must start and end with an alphanumeric character and Can only contain alphanumeric character and hyphens. Spaces are not allowed."),
				),
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
							ForceNew: true,
							MaxItems: 1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"catalog": {
										Type:     schema.TypeString,
										Optional: true,
										Computed: true,
										ForceNew: true,
									},
									names.AttrDatabase: {
										Type:     schema.TypeString,
										Optional: true,
										Computed: true,
										ForceNew: true,
									},
									names.AttrTableName: {
										Type:     schema.TypeString,
										Optional: true,
										Computed: true,
										ForceNew: true,
									},
								},
							},
						},
						"disable_glue_table_creation": {
							Type:     schema.TypeBool,
							Optional: true,
							ForceNew: true,
						},
						"s3_storage_config": {
							Type:     schema.TypeList,
							Required: true,
							ForceNew: true,
							MaxItems: 1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									names.AttrKMSKeyID: {
										Type:         schema.TypeString,
										Optional:     true,
										ForceNew:     true,
										ValidateFunc: verify.ValidARN,
									},
									"resolved_output_s3_uri": {
										Type:     schema.TypeString,
										Optional: true,
										Computed: true,
										ForceNew: true,
									},
									"s3_uri": {
										Type:     schema.TypeString,
										Required: true,
										ForceNew: true,
									},
								},
							},
						},
						"table_format": {
							Type:             schema.TypeString,
							Optional:         true,
							ForceNew:         true,
							Default:          awstypes.TableFormatGlue,
							ValidateDiagFunc: enum.Validate[awstypes.TableFormat](),
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
						"enable_online_store": {
							Type:     schema.TypeBool,
							Optional: true,
							ForceNew: true,
							Default:  false,
						},
						"security_config": {
							Type:     schema.TypeList,
							Optional: true,
							ForceNew: true,
							MaxItems: 1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									names.AttrKMSKeyID: {
										Type:         schema.TypeString,
										Optional:     true,
										ForceNew:     true,
										ValidateFunc: verify.ValidARN,
									},
								},
							},
						},
						names.AttrStorageType: {
							Type:             schema.TypeString,
							Optional:         true,
							ForceNew:         true,
							ValidateDiagFunc: enum.Validate[awstypes.StorageType](),
						},
						"ttl_duration": {
							Type:     schema.TypeList,
							Optional: true,
							MaxItems: 1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									names.AttrUnit: {
										Type:             schema.TypeString,
										Optional:         true,
										ValidateDiagFunc: enum.Validate[awstypes.TtlDurationUnit](),
									},
									names.AttrValue: {
										Type:     schema.TypeInt,
										Optional: true,
									},
								},
							},
						},
					},
				},
			},
			"record_identifier_feature_name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
				ValidateFunc: validation.All(
					validation.StringLenBetween(1, 64),
					validation.StringMatch(regexache.MustCompile(`^[0-9A-Za-z]([-_]*[0-9A-Za-z]){0,63}`),
						"Must start and end with an alphanumeric character and Can only contains alphanumeric characters, hyphens, underscores. Spaces are not allowed."),
				),
			},
			names.AttrRoleARN: {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: verify.ValidARN,
			},
			names.AttrTags:    tftags.TagsSchema(),
			names.AttrTagsAll: tftags.TagsSchemaComputed(),
			"throughput_config": {
				Type:     schema.TypeList,
				Optional: true,
				Computed: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"throughput_mode": {
							Type:             schema.TypeString,
							Optional:         true,
							Computed:         true,
							ValidateDiagFunc: enum.Validate[awstypes.ThroughputMode](),
						},
						"provisioned_read_capacity_units": {
							Type:         schema.TypeInt,
							Optional:     true,
							ValidateFunc: validation.IntBetween(0, 10000000),
						},
						"provisioned_write_capacity_units": {
							Type:         schema.TypeInt,
							Optional:     true,
							ValidateFunc: validation.IntBetween(0, 10000000),
						},
					},
				},
			},
		},
	}
}

func resourceFeatureGroupCreate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).SageMakerClient(ctx)

	name := d.Get("feature_group_name").(string)
	input := &sagemaker.CreateFeatureGroupInput{
		FeatureGroupName:            aws.String(name),
		EventTimeFeatureName:        aws.String(d.Get("event_time_feature_name").(string)),
		RecordIdentifierFeatureName: aws.String(d.Get("record_identifier_feature_name").(string)),
		RoleArn:                     aws.String(d.Get(names.AttrRoleARN).(string)),
		FeatureDefinitions:          expandFeatureGroupFeatureDefinition(d.Get("feature_definition").([]any)),
		Tags:                        getTagsIn(ctx),
	}

	if v, ok := d.GetOk(names.AttrDescription); ok {
		input.Description = aws.String(v.(string))
	}

	if v, ok := d.GetOk("offline_store_config"); ok {
		input.OfflineStoreConfig = expandFeatureGroupOfflineStoreConfig(v.([]any))
	}

	if v, ok := d.GetOk("online_store_config"); ok {
		input.OnlineStoreConfig = expandFeatureGroupOnlineStoreConfig(v.([]any))
	}

	if v, ok := d.GetOk("throughput_config"); ok {
		input.ThroughputConfig = expandThroughputConfig(v.([]any))
	}

	log.Printf("[DEBUG] SageMaker AI Feature Group create config: %#v", *input)
	err := retry.RetryContext(ctx, propagationTimeout, func() *retry.RetryError {
		_, err := conn.CreateFeatureGroup(ctx, input)
		if err != nil {
			if tfawserr.ErrMessageContains(err, "ValidationException", "The execution role ARN is invalid.") {
				return retry.RetryableError(err)
			}
			if tfawserr.ErrMessageContains(err, "ValidationException", "Invalid S3Uri provided") {
				return retry.RetryableError(err)
			}
			return retry.NonRetryableError(err)
		}

		return nil
	})
	if tfresource.TimedOut(err) {
		_, err = conn.CreateFeatureGroup(ctx, input)
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating SageMaker AI Feature Group: %s", err)
	}

	d.SetId(name)

	if _, err := waitFeatureGroupCreated(ctx, conn, d.Id()); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for SageMaker AI Feature Group (%s) to create: %s", d.Id(), err)
	}

	return append(diags, resourceFeatureGroupRead(ctx, d, meta)...)
}

func resourceFeatureGroupRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).SageMakerClient(ctx)

	output, err := findFeatureGroupByName(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] SageMaker AI Feature Group (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading SageMaker AI Feature Group (%s): %s", d.Id(), err)
	}

	d.Set("feature_group_name", output.FeatureGroupName)
	d.Set("event_time_feature_name", output.EventTimeFeatureName)
	d.Set(names.AttrDescription, output.Description)
	d.Set("record_identifier_feature_name", output.RecordIdentifierFeatureName)
	d.Set(names.AttrRoleARN, output.RoleArn)
	d.Set(names.AttrARN, output.FeatureGroupArn)

	if err := d.Set("feature_definition", flattenFeatureGroupFeatureDefinition(output.FeatureDefinitions)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting feature_definition for SageMaker AI Feature Group (%s): %s", d.Id(), err)
	}

	if err := d.Set("online_store_config", flattenFeatureGroupOnlineStoreConfig(output.OnlineStoreConfig)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting online_store_config for SageMaker AI Feature Group (%s): %s", d.Id(), err)
	}

	if err := d.Set("offline_store_config", flattenFeatureGroupOfflineStoreConfig(output.OfflineStoreConfig)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting offline_store_config for SageMaker AI Feature Group (%s): %s", d.Id(), err)
	}

	if err := d.Set("throughput_config", flattenThroughputConfig(output.ThroughputConfig)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting throughput_config for SageMaker AI Feature Group (%s): %s", d.Id(), err)
	}

	return diags
}

func resourceFeatureGroupUpdate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).SageMakerClient(ctx)

	if d.HasChangesExcept(names.AttrTags, names.AttrTagsAll) {
		input := &sagemaker.UpdateFeatureGroupInput{
			FeatureGroupName: aws.String(d.Id()),
		}

		if d.HasChange("online_store_config") {
			input.OnlineStoreConfig = expandFeatureGroupOnlineStoreConfigUpdate(d.Get("online_store_config").([]any))
		}

		if d.HasChange("throughput_config") {
			input.ThroughputConfig = expandThroughputConfigUpdate(d.Get("throughput_config").([]any))
		}

		_, err := conn.UpdateFeatureGroup(ctx, input)

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "updating SageMaker AI Feature Group (%s): %s", d.Id(), err)
		}

		if _, err := waitFeatureGroupUpdated(ctx, conn, d.Id()); err != nil {
			return sdkdiag.AppendErrorf(diags, "waiting for SageMaker AI Feature Group (%s) to update: %s", d.Id(), err)
		}
	}

	return append(diags, resourceFeatureGroupRead(ctx, d, meta)...)
}

func resourceFeatureGroupDelete(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).SageMakerClient(ctx)

	input := &sagemaker.DeleteFeatureGroupInput{
		FeatureGroupName: aws.String(d.Id()),
	}

	if _, err := conn.DeleteFeatureGroup(ctx, input); err != nil {
		if errs.IsA[*awstypes.ResourceNotFound](err) {
			return diags
		}
		return sdkdiag.AppendErrorf(diags, "deleting SageMaker AI Feature Group (%s): %s", d.Id(), err)
	}

	if _, err := waitFeatureGroupDeleted(ctx, conn, d.Id()); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for SageMaker AI Feature Group (%s) to delete: %s", d.Id(), err)
	}

	return diags
}

func findFeatureGroupByName(ctx context.Context, conn *sagemaker.Client, name string) (*sagemaker.DescribeFeatureGroupOutput, error) {
	input := &sagemaker.DescribeFeatureGroupInput{
		FeatureGroupName: aws.String(name),
	}

	output, err := conn.DescribeFeatureGroup(ctx, input)

	if errs.IsA[*awstypes.ResourceNotFound](err) {
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

func expandFeatureGroupFeatureDefinition(l []any) []awstypes.FeatureDefinition {
	featureDefs := make([]awstypes.FeatureDefinition, 0, len(l))

	for _, lRaw := range l {
		data := lRaw.(map[string]any)

		featureDef := awstypes.FeatureDefinition{
			FeatureName: aws.String(data["feature_name"].(string)),
			FeatureType: awstypes.FeatureType(data["feature_type"].(string)),
		}

		if v, ok := data["collection_config"].([]any); ok && len(v) > 0 {
			featureDef.CollectionConfig = expandCollectionConfig(v)
		}

		if v, ok := data["collection_type"].(string); ok && v != "" {
			featureDef.CollectionType = awstypes.CollectionType(v)
		}

		featureDefs = append(featureDefs, featureDef)
	}

	return featureDefs
}

func expandCollectionConfig(l []any) *awstypes.CollectionConfigMemberVectorConfig {
	if len(l) == 0 || l[0] == nil {
		return nil
	}

	m := l[0].(map[string]any)

	vectorConfig := expandVectorConfig(m["vector_config"].([]any))

	config := &awstypes.CollectionConfigMemberVectorConfig{
		Value: *vectorConfig,
	}

	return config
}

func expandVectorConfig(l []any) *awstypes.VectorConfig {
	if len(l) == 0 || l[0] == nil {
		return nil
	}

	m := l[0].(map[string]any)

	config := &awstypes.VectorConfig{
		Dimension: aws.Int32(int32(m["dimension"].(int))),
	}

	return config
}

func flattenFeatureGroupFeatureDefinition(config []awstypes.FeatureDefinition) []map[string]any {
	features := make([]map[string]any, 0, len(config))

	for _, i := range config {
		feature := map[string]any{
			"feature_name": aws.ToString(i.FeatureName),
			"feature_type": i.FeatureType,
		}

		if i.CollectionConfig != nil {
			feature["collection_config"] = flattenCollectionConfig(i.CollectionConfig.(*awstypes.CollectionConfigMemberVectorConfig))
		}

		if i.CollectionType != "" {
			feature["collection_type"] = i.CollectionType
		}

		features = append(features, feature)
	}
	return features
}

func flattenCollectionConfig(config *awstypes.CollectionConfigMemberVectorConfig) []map[string]any {
	if config == nil {
		return []map[string]any{}
	}

	m := map[string]any{
		"vector_config": flattenVectorConfig(&config.Value),
	}

	return []map[string]any{m}
}

func flattenVectorConfig(config *awstypes.VectorConfig) []map[string]any {
	if config == nil {
		return []map[string]any{}
	}

	m := map[string]any{
		"dimension": aws.ToInt32(config.Dimension),
	}

	return []map[string]any{m}
}

func expandFeatureGroupOnlineStoreConfig(l []any) *awstypes.OnlineStoreConfig {
	if len(l) == 0 || l[0] == nil {
		return nil
	}

	m := l[0].(map[string]any)

	config := &awstypes.OnlineStoreConfig{
		EnableOnlineStore: aws.Bool(m["enable_online_store"].(bool)),
	}

	if v, ok := m["security_config"].([]any); ok && len(v) > 0 {
		config.SecurityConfig = expandFeatureGroupOnlineStoreConfigSecurityConfig(v)
	}

	if v, ok := m[names.AttrStorageType].(string); ok && v != "" {
		config.StorageType = awstypes.StorageType(v)
	}

	if v, ok := m["ttl_duration"].([]any); ok && len(v) > 0 {
		config.TtlDuration = expandFeatureGroupOnlineStoreConfigTTLDuration(v)
	}

	return config
}

func flattenFeatureGroupOnlineStoreConfig(config *awstypes.OnlineStoreConfig) []map[string]any {
	if config == nil {
		return []map[string]any{}
	}

	m := map[string]any{
		names.AttrStorageType: config.StorageType,
		"enable_online_store": aws.ToBool(config.EnableOnlineStore),
	}

	if config.SecurityConfig != nil {
		m["security_config"] = flattenFeatureGroupOnlineStoreConfigSecurityConfig(config.SecurityConfig)
	}

	if config.TtlDuration != nil {
		m["ttl_duration"] = flattenFeatureGroupOnlineStoreConfigTTLDuration(config.TtlDuration)
	}

	return []map[string]any{m}
}

func expandFeatureGroupOnlineStoreConfigSecurityConfig(l []any) *awstypes.OnlineStoreSecurityConfig {
	if len(l) == 0 || l[0] == nil {
		return nil
	}

	m := l[0].(map[string]any)

	config := &awstypes.OnlineStoreSecurityConfig{
		KmsKeyId: aws.String(m[names.AttrKMSKeyID].(string)),
	}

	return config
}

func flattenFeatureGroupOnlineStoreConfigSecurityConfig(config *awstypes.OnlineStoreSecurityConfig) []map[string]any {
	if config == nil {
		return []map[string]any{}
	}

	m := map[string]any{
		names.AttrKMSKeyID: aws.ToString(config.KmsKeyId),
	}

	return []map[string]any{m}
}

func expandFeatureGroupOnlineStoreConfigTTLDuration(l []any) *awstypes.TtlDuration {
	if len(l) == 0 || l[0] == nil {
		return nil
	}

	m := l[0].(map[string]any)

	config := &awstypes.TtlDuration{
		Unit:  awstypes.TtlDurationUnit(m[names.AttrUnit].(string)),
		Value: aws.Int32(int32(m[names.AttrValue].(int))),
	}

	return config
}

func flattenFeatureGroupOnlineStoreConfigTTLDuration(config *awstypes.TtlDuration) []map[string]any {
	if config == nil {
		return []map[string]any{}
	}

	m := map[string]any{
		names.AttrUnit:  config.Unit,
		names.AttrValue: aws.ToInt32(config.Value),
	}

	return []map[string]any{m}
}

func expandFeatureGroupOfflineStoreConfig(l []any) *awstypes.OfflineStoreConfig {
	if len(l) == 0 || l[0] == nil {
		return nil
	}

	m := l[0].(map[string]any)

	config := &awstypes.OfflineStoreConfig{}

	if v, ok := m["s3_storage_config"].([]any); ok && len(v) > 0 {
		config.S3StorageConfig = expandFeatureGroupOfflineStoreConfigS3StorageConfig(v)
	}

	if v, ok := m["data_catalog_config"].([]any); ok && len(v) > 0 {
		config.DataCatalogConfig = expandFeatureGroupOfflineStoreConfigDataCatalogConfig(v)
	}

	if v, ok := m["disable_glue_table_creation"].(bool); ok {
		config.DisableGlueTableCreation = aws.Bool(v)
	}

	if v, ok := m["table_format"].(string); ok {
		config.TableFormat = awstypes.TableFormat(v)
	}

	return config
}

func flattenFeatureGroupOfflineStoreConfig(config *awstypes.OfflineStoreConfig) []map[string]any {
	if config == nil {
		return []map[string]any{}
	}

	m := map[string]any{
		"disable_glue_table_creation": aws.ToBool(config.DisableGlueTableCreation),
		"table_format":                config.TableFormat,
	}

	if config.DataCatalogConfig != nil {
		m["data_catalog_config"] = flattenFeatureGroupOfflineStoreConfigDataCatalogConfig(config.DataCatalogConfig)
	}

	if config.S3StorageConfig != nil {
		m["s3_storage_config"] = flattenFeatureGroupOfflineStoreConfigS3StorageConfig(config.S3StorageConfig)
	}

	return []map[string]any{m}
}

func expandFeatureGroupOfflineStoreConfigS3StorageConfig(l []any) *awstypes.S3StorageConfig {
	if len(l) == 0 || l[0] == nil {
		return nil
	}

	m := l[0].(map[string]any)

	config := &awstypes.S3StorageConfig{
		S3Uri: aws.String(m["s3_uri"].(string)),
	}

	if v, ok := m[names.AttrKMSKeyID].(string); ok && v != "" {
		config.KmsKeyId = aws.String(m[names.AttrKMSKeyID].(string))
	}

	if v, ok := m["resolved_output_s3_uri"].(string); ok && v != "" {
		config.ResolvedOutputS3Uri = aws.String(m["resolved_output_s3_uri"].(string))
	}

	return config
}

func flattenFeatureGroupOfflineStoreConfigS3StorageConfig(config *awstypes.S3StorageConfig) []map[string]any {
	if config == nil {
		return []map[string]any{}
	}

	m := map[string]any{
		"s3_uri": aws.ToString(config.S3Uri),
	}

	if config.KmsKeyId != nil {
		m[names.AttrKMSKeyID] = aws.ToString(config.KmsKeyId)
	}

	if config.ResolvedOutputS3Uri != nil {
		m["resolved_output_s3_uri"] = aws.ToString(config.ResolvedOutputS3Uri)
	}

	return []map[string]any{m}
}

func expandFeatureGroupOfflineStoreConfigDataCatalogConfig(l []any) *awstypes.DataCatalogConfig {
	if len(l) == 0 || l[0] == nil {
		return nil
	}

	m := l[0].(map[string]any)

	config := &awstypes.DataCatalogConfig{
		Catalog:   aws.String(m["catalog"].(string)),
		Database:  aws.String(m[names.AttrDatabase].(string)),
		TableName: aws.String(m[names.AttrTableName].(string)),
	}

	return config
}

func flattenFeatureGroupOfflineStoreConfigDataCatalogConfig(config *awstypes.DataCatalogConfig) []map[string]any {
	if config == nil {
		return []map[string]any{}
	}

	m := map[string]any{
		"catalog":           aws.ToString(config.Catalog),
		names.AttrDatabase:  aws.ToString(config.Database),
		names.AttrTableName: aws.ToString(config.TableName),
	}

	return []map[string]any{m}
}

func expandFeatureGroupOnlineStoreConfigUpdate(l []any) *awstypes.OnlineStoreConfigUpdate {
	if len(l) == 0 || l[0] == nil {
		return nil
	}

	m := l[0].(map[string]any)

	config := &awstypes.OnlineStoreConfigUpdate{}

	if v, ok := m["ttl_duration"].([]any); ok && len(v) > 0 {
		config.TtlDuration = expandFeatureGroupOnlineStoreConfigTTLDuration(v)
	}

	return config
}

func expandThroughputConfig(l []any) *awstypes.ThroughputConfig {
	if len(l) == 0 || l[0] == nil {
		return nil
	}

	m := l[0].(map[string]any)

	config := &awstypes.ThroughputConfig{
		ThroughputMode: awstypes.ThroughputMode(m["throughput_mode"].(string)),
	}

	if config.ThroughputMode == awstypes.ThroughputModeProvisioned {
		if v, ok := m["provisioned_read_capacity_units"].(int); ok {
			config.ProvisionedReadCapacityUnits = aws.Int32(int32(v))
		}

		if v, ok := m["provisioned_write_capacity_units"].(int); ok {
			config.ProvisionedWriteCapacityUnits = aws.Int32(int32(v))
		}
	}

	return config
}

func expandThroughputConfigUpdate(l []any) *awstypes.ThroughputConfigUpdate {
	if len(l) == 0 || l[0] == nil {
		return nil
	}

	m := l[0].(map[string]any)

	config := &awstypes.ThroughputConfigUpdate{
		ThroughputMode: awstypes.ThroughputMode(m["throughput_mode"].(string)),
	}

	if v, ok := m["provisioned_read_capacity_units"].(int); ok {
		config.ProvisionedReadCapacityUnits = aws.Int32(int32(v))
	}

	if v, ok := m["provisioned_write_capacity_units"].(int); ok {
		config.ProvisionedWriteCapacityUnits = aws.Int32(int32(v))
	}

	return config
}

func flattenThroughputConfig(config *awstypes.ThroughputConfigDescription) []map[string]any {
	if config == nil {
		return []map[string]any{}
	}

	m := map[string]any{
		"throughput_mode": config.ThroughputMode,
	}

	if config.ProvisionedReadCapacityUnits != nil {
		m["provisioned_read_capacity_units"] = aws.ToInt32(config.ProvisionedReadCapacityUnits)
	}

	if config.ProvisionedWriteCapacityUnits != nil {
		m["provisioned_write_capacity_units"] = aws.ToInt32(config.ProvisionedWriteCapacityUnits)
	}

	return []map[string]any{m}
}
