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
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

func ResourceEndpointConfiguration() *schema.Resource {
	return &schema.Resource{
		Create: resourceEndpointConfigurationCreate,
		Read:   resourceEndpointConfigurationRead,
		Update: resourceEndpointConfigurationUpdate,
		Delete: resourceEndpointConfigurationDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"async_inference_config": {
				Type:     schema.TypeList,
				MaxItems: 1,
				Optional: true,
				ForceNew: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"client_config": {
							Type:     schema.TypeList,
							Optional: true,
							MaxItems: 1,
							ForceNew: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"max_concurrent_invocations_per_instance": {
										Type:         schema.TypeInt,
										Optional:     true,
										ForceNew:     true,
										ValidateFunc: validation.IntBetween(1, 1000),
									},
								},
							},
						},
						"output_config": {
							Type:     schema.TypeList,
							Required: true,
							MaxItems: 1,
							ForceNew: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"kms_key_id": {
										Type:         schema.TypeString,
										Optional:     true,
										ForceNew:     true,
										ValidateFunc: verify.ValidARN,
									},
									"notification_config": {
										Type:     schema.TypeList,
										Optional: true,
										MaxItems: 1,
										ForceNew: true,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"error_topic": {
													Type:         schema.TypeString,
													Optional:     true,
													ForceNew:     true,
													ValidateFunc: verify.ValidARN,
												},
												"success_topic": {
													Type:         schema.TypeString,
													Optional:     true,
													ForceNew:     true,
													ValidateFunc: verify.ValidARN,
												},
											},
										},
									},
									"s3_output_path": {
										Type:     schema.TypeString,
										Required: true,
										ForceNew: true,
										ValidateFunc: validation.All(
											validation.StringMatch(regexp.MustCompile(`^(https|s3)://([^/])/?(.*)$`), ""),
											validation.StringLenBetween(1, 512),
										),
									},
								},
							},
						},
					},
				},
			},
			"data_capture_config": {
				Type:     schema.TypeList,
				MaxItems: 1,
				Optional: true,
				ForceNew: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"enable_capture": {
							Type:     schema.TypeBool,
							Optional: true,
							ForceNew: true,
						},

						"initial_sampling_percentage": {
							Type:         schema.TypeInt,
							Required:     true,
							ForceNew:     true,
							ValidateFunc: validation.IntBetween(0, 100),
						},

						"destination_s3_uri": {
							Type:     schema.TypeString,
							Required: true,
							ForceNew: true,
							ValidateFunc: validation.All(
								validation.StringMatch(regexp.MustCompile(`^(https|s3)://([^/])/?(.*)$`), ""),
								validation.StringLenBetween(1, 512),
							)},

						"kms_key_id": {
							Type:         schema.TypeString,
							Optional:     true,
							ForceNew:     true,
							ValidateFunc: verify.ValidARN,
						},

						"capture_options": {
							Type:     schema.TypeList,
							Required: true,
							MaxItems: 2,
							MinItems: 1,
							ForceNew: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"capture_mode": {
										Type:         schema.TypeString,
										Required:     true,
										ForceNew:     true,
										ValidateFunc: validation.StringInSlice(sagemaker.CaptureMode_Values(), false),
									},
								},
							},
						},

						"capture_content_type_header": {
							Type:     schema.TypeList,
							Optional: true,
							MaxItems: 1,
							ForceNew: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"csv_content_types": {
										Type:     schema.TypeSet,
										MinItems: 1,
										MaxItems: 10,
										Elem: &schema.Schema{
											Type: schema.TypeString,
											ValidateFunc: validation.All(
												validation.StringMatch(regexp.MustCompile(`^[a-zA-Z0-9](-*[a-zA-Z0-9])*\/[a-zA-Z0-9](-*[a-zA-Z0-9.])*`), ""),
												validation.StringLenBetween(1, 256),
											),
										},
										Optional: true,
										ForceNew: true,
									},
									"json_content_types": {
										Type:     schema.TypeSet,
										MinItems: 1,
										MaxItems: 10,
										Elem: &schema.Schema{
											Type: schema.TypeString,
											ValidateFunc: validation.All(
												validation.StringMatch(regexp.MustCompile(`^[a-zA-Z0-9](-*[a-zA-Z0-9])*\/[a-zA-Z0-9](-*[a-zA-Z0-9.])*`), ""),
												validation.StringLenBetween(1, 256),
											),
										},
										Optional: true,
										ForceNew: true,
									},
								},
							},
						},
					},
				},
			},
			"name": {
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				ForceNew:     true,
				ValidateFunc: validName,
			},
			"kms_key_arn": {
				Type:         schema.TypeString,
				Optional:     true,
				ForceNew:     true,
				ValidateFunc: verify.ValidARN,
			},
			"production_variants": {
				Type:     schema.TypeList,
				Required: true,
				MinItems: 1,
				MaxItems: 10,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"accelerator_type": {
							Type:         schema.TypeString,
							Optional:     true,
							ForceNew:     true,
							ValidateFunc: validation.StringInSlice(sagemaker.ProductionVariantAcceleratorType_Values(), false),
						},
						"initial_instance_count": {
							Type:         schema.TypeInt,
							Optional:     true,
							ForceNew:     true,
							ValidateFunc: validation.IntAtLeast(1),
						},
						"initial_variant_weight": {
							Type:         schema.TypeFloat,
							Optional:     true,
							ForceNew:     true,
							ValidateFunc: validation.FloatAtLeast(0),
							Default:      1,
						},
						"instance_type": {
							Type:         schema.TypeString,
							Optional:     true,
							ForceNew:     true,
							ValidateFunc: validation.StringInSlice(sagemaker.ProductionVariantInstanceType_Values(), false),
						},
						"model_name": {
							Type:     schema.TypeString,
							Required: true,
							ForceNew: true,
						},
						"serverless_config": {
							Type:     schema.TypeList,
							Optional: true,
							MaxItems: 1,
							ForceNew: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"max_concurrency": {
										Type:         schema.TypeInt,
										Required:     true,
										ForceNew:     true,
										ValidateFunc: validation.IntBetween(1, 200),
									},
									"memory_size_in_mb": {
										Type:         schema.TypeInt,
										Required:     true,
										ForceNew:     true,
										ValidateFunc: validation.IntInSlice([]int{1024, 2048, 3072, 4096, 5120, 6144}),
									},
								},
							},
						},
						"variant_name": {
							Type:     schema.TypeString,
							Optional: true,
							Computed: true,
							ForceNew: true,
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

func resourceEndpointConfigurationCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).SageMakerConn
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	tags := defaultTagsConfig.MergeTags(tftags.New(d.Get("tags").(map[string]interface{})))

	var name string
	if v, ok := d.GetOk("name"); ok {
		name = v.(string)
	} else {
		name = resource.UniqueId()
	}

	createOpts := &sagemaker.CreateEndpointConfigInput{
		EndpointConfigName: aws.String(name),
		ProductionVariants: expandProductionVariants(d.Get("production_variants").([]interface{})),
	}

	if v, ok := d.GetOk("kms_key_arn"); ok {
		createOpts.KmsKeyId = aws.String(v.(string))
	}

	if len(tags) > 0 {
		createOpts.Tags = Tags(tags.IgnoreAWS())
	}

	if v, ok := d.GetOk("data_capture_config"); ok {
		createOpts.DataCaptureConfig = expandDataCaptureConfig(v.([]interface{}))
	}

	if v, ok := d.GetOk("async_inference_config"); ok {
		createOpts.AsyncInferenceConfig = expandEndpointConfigAsyncInferenceConfig(v.([]interface{}))
	}

	log.Printf("[DEBUG] SageMaker Endpoint Configuration create config: %#v", *createOpts)
	_, err := conn.CreateEndpointConfig(createOpts)
	if err != nil {
		return fmt.Errorf("error creating SageMaker Endpoint Configuration: %w", err)
	}
	d.SetId(name)

	return resourceEndpointConfigurationRead(d, meta)
}

func resourceEndpointConfigurationRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).SageMakerConn
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	ignoreTagsConfig := meta.(*conns.AWSClient).IgnoreTagsConfig

	endpointConfig, err := FindEndpointConfigByName(conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] SageMaker Endpoint Configuration (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return fmt.Errorf("error reading SageMaker Endpoint Configuration (%s): %w", d.Id(), err)
	}

	d.Set("arn", endpointConfig.EndpointConfigArn)
	d.Set("name", endpointConfig.EndpointConfigName)
	d.Set("kms_key_arn", endpointConfig.KmsKeyId)

	if err := d.Set("production_variants", flattenProductionVariants(endpointConfig.ProductionVariants)); err != nil {
		return fmt.Errorf("error setting production_variants for SageMaker Endpoint Configuration (%s): %w", d.Id(), err)
	}

	if err := d.Set("data_capture_config", flattenDataCaptureConfig(endpointConfig.DataCaptureConfig)); err != nil {
		return fmt.Errorf("error setting data_capture_config for SageMaker Endpoint Configuration (%s): %w", d.Id(), err)
	}

	if err := d.Set("async_inference_config", flattenEndpointConfigAsyncInferenceConfig(endpointConfig.AsyncInferenceConfig)); err != nil {
		return fmt.Errorf("error setting async_inference_config for SageMaker Endpoint Configuration (%s): %w", d.Id(), err)
	}

	tags, err := ListTags(conn, aws.StringValue(endpointConfig.EndpointConfigArn))
	if err != nil {
		return fmt.Errorf("error listing tags for SageMaker Endpoint Configuration (%s): %w", d.Id(), err)
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

func resourceEndpointConfigurationUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).SageMakerConn

	if d.HasChange("tags_all") {
		o, n := d.GetChange("tags_all")

		if err := UpdateTags(conn, d.Get("arn").(string), o, n); err != nil {
			return fmt.Errorf("error updating SageMaker Endpoint Configuration (%s) tags: %w", d.Id(), err)
		}
	}
	return resourceEndpointConfigurationRead(d, meta)
}

func resourceEndpointConfigurationDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).SageMakerConn

	deleteOpts := &sagemaker.DeleteEndpointConfigInput{
		EndpointConfigName: aws.String(d.Id()),
	}
	log.Printf("[INFO] Deleting SageMaker Endpoint Configuration: %s", d.Id())

	_, err := conn.DeleteEndpointConfig(deleteOpts)

	if tfawserr.ErrMessageContains(err, "ValidationException", "Could not find endpoint configuration") {
		return nil
	}

	if err != nil {
		return fmt.Errorf("error deleting SageMaker Endpoint Configuration (%s): %w", d.Id(), err)
	}

	return nil
}

func expandProductionVariants(configured []interface{}) []*sagemaker.ProductionVariant {
	containers := make([]*sagemaker.ProductionVariant, 0, len(configured))

	for _, lRaw := range configured {
		data := lRaw.(map[string]interface{})

		l := &sagemaker.ProductionVariant{
			ModelName: aws.String(data["model_name"].(string)),
		}

		if v, ok := data["initial_instance_count"].(int); ok && v > 0 {
			l.InitialInstanceCount = aws.Int64(int64(v))
		}

		if v, ok := data["instance_type"].(string); ok && v != "" {
			l.InstanceType = aws.String(v)
		}

		if v, ok := data["variant_name"]; ok {
			l.VariantName = aws.String(v.(string))
		} else {
			l.VariantName = aws.String(resource.UniqueId())
		}

		if v, ok := data["initial_variant_weight"]; ok {
			l.InitialVariantWeight = aws.Float64(v.(float64))
		}

		if v, ok := data["accelerator_type"].(string); ok && v != "" {
			l.AcceleratorType = aws.String(v)
		}

		if v, ok := data["serverless_config"].([]interface{}); ok && len(v) > 0 {
			l.ServerlessConfig = expandServerlessConfig(v)
		}

		containers = append(containers, l)
	}

	return containers
}

func flattenProductionVariants(list []*sagemaker.ProductionVariant) []map[string]interface{} {
	result := make([]map[string]interface{}, 0, len(list))

	for _, i := range list {
		l := map[string]interface{}{
			"accelerator_type":       aws.StringValue(i.AcceleratorType),
			"initial_variant_weight": aws.Float64Value(i.InitialVariantWeight),
			"model_name":             aws.StringValue(i.ModelName),
			"variant_name":           aws.StringValue(i.VariantName),
		}

		if i.InitialInstanceCount != nil {
			l["initial_instance_count"] = aws.Int64Value(i.InitialInstanceCount)
		}

		if i.InstanceType != nil {
			l["instance_type"] = aws.StringValue(i.InstanceType)
		}

		if i.ServerlessConfig != nil {
			l["serverless_config"] = flattenServerlessConfig(i.ServerlessConfig)
		}

		result = append(result, l)
	}
	return result
}

func expandDataCaptureConfig(configured []interface{}) *sagemaker.DataCaptureConfig {
	if len(configured) == 0 {
		return nil
	}

	m := configured[0].(map[string]interface{})

	c := &sagemaker.DataCaptureConfig{
		InitialSamplingPercentage: aws.Int64(int64(m["initial_sampling_percentage"].(int))),
		DestinationS3Uri:          aws.String(m["destination_s3_uri"].(string)),
		CaptureOptions:            expandCaptureOptions(m["capture_options"].([]interface{})),
	}

	if v, ok := m["enable_capture"]; ok {
		c.EnableCapture = aws.Bool(v.(bool))
	}

	if v, ok := m["kms_key_id"].(string); ok && v != "" {
		c.KmsKeyId = aws.String(v)
	}

	if v, ok := m["capture_content_type_header"].([]interface{}); ok && (len(v) > 0) {
		c.CaptureContentTypeHeader = expandCaptureContentTypeHeader(v[0].(map[string]interface{}))
	}

	return c
}

func flattenDataCaptureConfig(dataCaptureConfig *sagemaker.DataCaptureConfig) []map[string]interface{} {
	if dataCaptureConfig == nil {
		return []map[string]interface{}{}
	}

	cfg := map[string]interface{}{
		"initial_sampling_percentage": aws.Int64Value(dataCaptureConfig.InitialSamplingPercentage),
		"destination_s3_uri":          aws.StringValue(dataCaptureConfig.DestinationS3Uri),
		"capture_options":             flattenCaptureOptions(dataCaptureConfig.CaptureOptions),
	}

	if dataCaptureConfig.EnableCapture != nil {
		cfg["enable_capture"] = aws.BoolValue(dataCaptureConfig.EnableCapture)
	}

	if dataCaptureConfig.KmsKeyId != nil {
		cfg["kms_key_id"] = aws.StringValue(dataCaptureConfig.KmsKeyId)
	}

	if dataCaptureConfig.CaptureContentTypeHeader != nil {
		cfg["capture_content_type_header"] = flattenCaptureContentTypeHeader(dataCaptureConfig.CaptureContentTypeHeader)
	}

	return []map[string]interface{}{cfg}
}

func expandCaptureOptions(configured []interface{}) []*sagemaker.CaptureOption {
	containers := make([]*sagemaker.CaptureOption, 0, len(configured))

	for _, lRaw := range configured {
		data := lRaw.(map[string]interface{})

		l := &sagemaker.CaptureOption{
			CaptureMode: aws.String(data["capture_mode"].(string)),
		}
		containers = append(containers, l)
	}

	return containers
}

func flattenCaptureOptions(list []*sagemaker.CaptureOption) []map[string]interface{} {
	containers := make([]map[string]interface{}, 0, len(list))

	for _, lRaw := range list {
		captureOption := make(map[string]interface{})
		captureOption["capture_mode"] = aws.StringValue(lRaw.CaptureMode)
		containers = append(containers, captureOption)
	}

	return containers
}

func expandCaptureContentTypeHeader(m map[string]interface{}) *sagemaker.CaptureContentTypeHeader {
	c := &sagemaker.CaptureContentTypeHeader{}

	if v, ok := m["csv_content_types"].(*schema.Set); ok && v.Len() > 0 {
		c.CsvContentTypes = flex.ExpandStringSet(v)
	}

	if v, ok := m["json_content_types"].(*schema.Set); ok && v.Len() > 0 {
		c.JsonContentTypes = flex.ExpandStringSet(v)
	}

	return c
}

func flattenCaptureContentTypeHeader(contentTypeHeader *sagemaker.CaptureContentTypeHeader) []map[string]interface{} {
	if contentTypeHeader == nil {
		return []map[string]interface{}{}
	}

	l := make(map[string]interface{})

	if contentTypeHeader.CsvContentTypes != nil {
		l["csv_content_types"] = flex.FlattenStringSet(contentTypeHeader.CsvContentTypes)
	}

	if contentTypeHeader.JsonContentTypes != nil {
		l["json_content_types"] = flex.FlattenStringSet(contentTypeHeader.JsonContentTypes)
	}

	return []map[string]interface{}{l}
}

func expandEndpointConfigAsyncInferenceConfig(configured []interface{}) *sagemaker.AsyncInferenceConfig {
	if len(configured) == 0 {
		return nil
	}

	m := configured[0].(map[string]interface{})

	c := &sagemaker.AsyncInferenceConfig{}

	if v, ok := m["client_config"].([]interface{}); ok && len(v) > 0 {
		c.ClientConfig = expandEndpointConfigClientConfig(v)
	}

	if v, ok := m["output_config"].([]interface{}); ok && len(v) > 0 {
		c.OutputConfig = expandEndpointConfigOutputConfig(v)
	}

	return c
}

func expandEndpointConfigClientConfig(configured []interface{}) *sagemaker.AsyncInferenceClientConfig {
	if len(configured) == 0 {
		return nil
	}

	m := configured[0].(map[string]interface{})

	c := &sagemaker.AsyncInferenceClientConfig{}

	if v, ok := m["max_concurrent_invocations_per_instance"]; ok {
		c.MaxConcurrentInvocationsPerInstance = aws.Int64(int64(v.(int)))
	}

	return c
}

func expandEndpointConfigOutputConfig(configured []interface{}) *sagemaker.AsyncInferenceOutputConfig {
	if len(configured) == 0 {
		return nil
	}

	m := configured[0].(map[string]interface{})

	c := &sagemaker.AsyncInferenceOutputConfig{
		S3OutputPath: aws.String(m["s3_output_path"].(string)),
	}

	if v, ok := m["kms_key_id"].(string); ok && v != "" {
		c.KmsKeyId = aws.String(v)
	}

	if v, ok := m["notification_config"].([]interface{}); ok && len(v) > 0 {
		c.NotificationConfig = expandEndpointConfigNotificationConfig(v)
	}

	return c
}

func expandEndpointConfigNotificationConfig(configured []interface{}) *sagemaker.AsyncInferenceNotificationConfig {
	if len(configured) == 0 {
		return nil
	}

	m := configured[0].(map[string]interface{})

	c := &sagemaker.AsyncInferenceNotificationConfig{}

	if v, ok := m["error_topic"].(string); ok && v != "" {
		c.ErrorTopic = aws.String(v)
	}

	if v, ok := m["success_topic"].(string); ok && v != "" {
		c.SuccessTopic = aws.String(v)
	}

	return c
}

func expandServerlessConfig(configured []interface{}) *sagemaker.ProductionVariantServerlessConfig {
	if len(configured) == 0 {
		return nil
	}

	m := configured[0].(map[string]interface{})

	c := &sagemaker.ProductionVariantServerlessConfig{}

	if v, ok := m["max_concurrency"].(int); ok {
		c.MaxConcurrency = aws.Int64(int64(v))
	}

	if v, ok := m["memory_size_in_mb"].(int); ok {
		c.MemorySizeInMB = aws.Int64(int64(v))
	}

	return c
}

func flattenEndpointConfigAsyncInferenceConfig(config *sagemaker.AsyncInferenceConfig) []map[string]interface{} {
	if config == nil {
		return []map[string]interface{}{}
	}

	cfg := map[string]interface{}{}

	if config.ClientConfig != nil {
		cfg["client_config"] = flattenEndpointConfigClientConfig(config.ClientConfig)
	}

	if config.OutputConfig != nil {
		cfg["output_config"] = flattenEndpointConfigOutputConfig(config.OutputConfig)
	}

	return []map[string]interface{}{cfg}
}

func flattenEndpointConfigClientConfig(config *sagemaker.AsyncInferenceClientConfig) []map[string]interface{} {
	if config == nil {
		return []map[string]interface{}{}
	}

	cfg := map[string]interface{}{}

	if config.MaxConcurrentInvocationsPerInstance != nil {
		cfg["max_concurrent_invocations_per_instance"] = aws.Int64Value(config.MaxConcurrentInvocationsPerInstance)
	}

	return []map[string]interface{}{cfg}
}

func flattenEndpointConfigOutputConfig(config *sagemaker.AsyncInferenceOutputConfig) []map[string]interface{} {
	if config == nil {
		return []map[string]interface{}{}
	}

	cfg := map[string]interface{}{
		"s3_output_path": aws.StringValue(config.S3OutputPath),
	}

	if config.KmsKeyId != nil {
		cfg["kms_key_id"] = aws.StringValue(config.KmsKeyId)
	}

	if config.NotificationConfig != nil {
		cfg["notification_config"] = flattenEndpointConfigNotificationConfig(config.NotificationConfig)
	}

	return []map[string]interface{}{cfg}
}

func flattenEndpointConfigNotificationConfig(config *sagemaker.AsyncInferenceNotificationConfig) []map[string]interface{} {
	if config == nil {
		return []map[string]interface{}{}
	}

	cfg := map[string]interface{}{}

	if config.ErrorTopic != nil {
		cfg["error_topic"] = aws.StringValue(config.ErrorTopic)
	}

	if config.SuccessTopic != nil {
		cfg["success_topic"] = aws.StringValue(config.SuccessTopic)
	}

	return []map[string]interface{}{cfg}
}

func flattenServerlessConfig(config *sagemaker.ProductionVariantServerlessConfig) []map[string]interface{} {
	if config == nil {
		return []map[string]interface{}{}
	}

	cfg := map[string]interface{}{}

	if config.MaxConcurrency != nil {
		cfg["max_concurrency"] = aws.Int64Value(config.MaxConcurrency)
	}

	if config.MemorySizeInMB != nil {
		cfg["memory_size_in_mb"] = aws.Int64Value(config.MemorySizeInMB)
	}

	return []map[string]interface{}{cfg}
}
