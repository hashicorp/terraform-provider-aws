package sagemaker

import (
	"context"
	"log"
	"regexp"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/sagemaker"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

// @SDKResource("aws_sagemaker_data_quality_job_definition")
func ResourceDataQualityJobDefinition() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceDataQualityJobDefinitionCreate,
		ReadWithoutTimeout:   resourceDataQualityJobDefinitionRead,
		UpdateWithoutTimeout: resourceDataQualityJobDefinitionUpdate,
		DeleteWithoutTimeout: resourceDataQualityJobDefinitionDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Schema: map[string]*schema.Schema{
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"data_quality_app_specification": {
				Type:     schema.TypeList,
				MaxItems: 1,
				Required: true,
				ForceNew: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"image_uri": {
							Type:         schema.TypeString,
							Required:     true,
							ForceNew:     true,
							ValidateFunc: validImage,
						},
					},
				},
			},
			"data_quality_job_input": {
				Type:     schema.TypeList,
				MaxItems: 1,
				Required: true,
				ForceNew: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"endpoint_input": {
							Type:     schema.TypeList,
							MaxItems: 1,
							Required: true,
							ForceNew: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"endpoint_name": {
										Type:         schema.TypeString,
										Required:     true,
										ForceNew:     true,
										ValidateFunc: validName,
									},
									"local_path": {
										Type:     schema.TypeString,
										Optional: true,
										Default:  "/opt/ml/processing/input",
										ForceNew: true,
										ValidateFunc: validation.All(
											validation.StringLenBetween(1, 1024),
											validation.StringMatch(regexp.MustCompile(`^\/opt\/ml\/processing\/.*`), "Must start with `/opt/ml/processing`."),
										),
									},
								},
							},
						},
					},
				},
			},
			"data_quality_job_output_config": {
				Type:     schema.TypeList,
				MaxItems: 1,
				Required: true,
				ForceNew: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"monitoring_outputs": {
							Type:     schema.TypeList,
							MinItems: 1,
							Required: true,
							ForceNew: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"s3_output": {
										Type:     schema.TypeList,
										MaxItems: 1,
										Required: true,
										ForceNew: true,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"local_path": {
													Type:     schema.TypeString,
													Optional: true,
													Default:  "/opt/ml/processing/output",
													ForceNew: true,
													ValidateFunc: validation.All(
														validation.StringLenBetween(1, 1024),
														validation.StringMatch(regexp.MustCompile(`^\/opt\/ml\/processing\/.*`), "Must start with `/opt/ml/processing`."),
													),
												},
												"s3_uri": {
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
					},
				},
			},
			"job_resources": {
				Type:     schema.TypeList,
				MaxItems: 1,
				Required: true,
				ForceNew: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"cluster_config": {
							Type:     schema.TypeList,
							MaxItems: 1,
							Required: true,
							ForceNew: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"instance_count": {
										Type:         schema.TypeInt,
										Required:     true,
										ForceNew:     true,
										ValidateFunc: validation.IntAtLeast(1),
									},
									"instance_type": {
										Type:         schema.TypeString,
										Required:     true,
										ForceNew:     true,
										ValidateFunc: validation.StringInSlice(sagemaker.ProcessingInstanceType_Values(), false),
									},
									"volume_size_in_gb": {
										Type:         schema.TypeInt,
										Required:     true,
										ForceNew:     true,
										ValidateFunc: validation.IntBetween(1, 512),
									},
								},
							},
						},
					},
				},
			},
			"role_arn": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: verify.ValidARN,
			},
			"name": {
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				ForceNew:     true,
				ValidateFunc: validName,
			},
			"tags":     tftags.TagsSchema(),
			"tags_all": tftags.TagsSchemaComputed(),
		},
		CustomizeDiff: verify.SetTagsDiff,
	}
}

func resourceDataQualityJobDefinitionCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).SageMakerConn()
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	tags := defaultTagsConfig.MergeTags(tftags.New(ctx, d.Get("tags").(map[string]interface{})))

	var name string
	if v, ok := d.GetOk("name"); ok {
		name = v.(string)
	} else {
		name = resource.UniqueId()
	}

	var roleArn string
	if v, ok := d.GetOk("role_arn"); ok {
		roleArn = v.(string)
	}

	createOpts := &sagemaker.CreateDataQualityJobDefinitionInput{
		JobDefinitionName:           aws.String(name),
		DataQualityAppSpecification: expandDataQualityAppSpecification(d.Get("data_quality_app_specification").([]interface{})),
		DataQualityJobInput:         expandDataQualityJobInput(d.Get("data_quality_job_input").([]interface{})),
		DataQualityJobOutputConfig:  expandDataQualityJobOutputConfig(d.Get("data_quality_job_output_config").([]interface{})),
		JobResources:                expandJobResources(d.Get("job_resources").([]interface{})),
		RoleArn:                     aws.String(roleArn),
	}

	if len(tags) > 0 {
		createOpts.Tags = Tags(tags.IgnoreAWS())
	}

	log.Printf("[DEBUG] SageMaker Data Quality Job Definition create config: %#v", *createOpts)
	_, err := conn.CreateDataQualityJobDefinitionWithContext(ctx, createOpts)
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating SageMaker Data Quality Job Definition: %s", err)
	}
	d.SetId(name)

	return append(diags, resourceDataQualityJobDefinitionRead(ctx, d, meta)...)
}

func resourceDataQualityJobDefinitionRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).SageMakerConn()
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	ignoreTagsConfig := meta.(*conns.AWSClient).IgnoreTagsConfig

	jobDefinition, err := FindDataQualityJobDefinitionByName(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] SageMaker Data Quality Job Definition (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading SageMaker Data Quality Job Definition (%s): %s", d.Id(), err)
	}

	d.Set("arn", jobDefinition.JobDefinitionArn)
	d.Set("name", jobDefinition.JobDefinitionName)
	d.Set("role_arn", jobDefinition.RoleArn)

	if err := d.Set("data_quality_app_specification", flattenDataQualityAppSpecification(jobDefinition.DataQualityAppSpecification)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting data_quality_app_specification for SageMaker Data Quality Job Definition (%s): %s", d.Id(), err)
	}

	if err := d.Set("data_quality_job_input", flattenDataQualityJobInput(jobDefinition.DataQualityJobInput)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting data_quality_job_input for SageMaker Data Quality Job Definition (%s): %s", d.Id(), err)
	}

	if err := d.Set("data_quality_job_output_config", flattenDataQualityJobOutputConfig(jobDefinition.DataQualityJobOutputConfig)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting data_quality_job_output_config for SageMaker Data Quality Job Definition (%s): %s", d.Id(), err)
	}

	if err := d.Set("job_resources", flattenJobResources(jobDefinition.JobResources)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting job_resources for SageMaker Data Quality Job Definition (%s): %s", d.Id(), err)
	}

	tags, err := ListTags(ctx, conn, aws.StringValue(jobDefinition.JobDefinitionArn))
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "listing tags for SageMaker Data Quality Job Definition (%s): %s", d.Id(), err)
	}

	tags = tags.IgnoreAWS().IgnoreConfig(ignoreTagsConfig)

	//lintignore:AWSR002
	if err := d.Set("tags", tags.RemoveDefaultConfig(defaultTagsConfig).Map()); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting tags: %s", err)
	}

	if err := d.Set("tags_all", tags.Map()); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting tags_all: %s", err)
	}

	return diags
}

func flattenDataQualityAppSpecification(appSpecification *sagemaker.DataQualityAppSpecification) []map[string]interface{} {
	if appSpecification == nil {
		return []map[string]interface{}{}
	}

	spec := map[string]interface{}{}

	if appSpecification.ImageUri != nil {
		spec["image_uri"] = aws.StringValue(appSpecification.ImageUri)
	}

	return []map[string]interface{}{spec}
}

func flattenDataQualityJobInput(jobInput *sagemaker.DataQualityJobInput) []map[string]interface{} {
	if jobInput == nil {
		return []map[string]interface{}{}
	}

	spec := map[string]interface{}{}

	if jobInput.EndpointInput != nil {
		spec["endpoint_input"] = flattenEndpointInput(jobInput.EndpointInput)
	}

	return []map[string]interface{}{spec}
}

func flattenEndpointInput(endpointInput *sagemaker.EndpointInput) []map[string]interface{} {
	if endpointInput == nil {
		return []map[string]interface{}{}
	}

	spec := map[string]interface{}{}

	if endpointInput.EndpointName != nil {
		spec["endpoint_name"] = aws.StringValue(endpointInput.EndpointName)
	}

	if endpointInput.LocalPath != nil {
		spec["local_path"] = aws.StringValue(endpointInput.LocalPath)
	}

	return []map[string]interface{}{spec}
}

func flattenDataQualityJobOutputConfig(outputConfig *sagemaker.MonitoringOutputConfig) []map[string]interface{} {
	if outputConfig == nil {
		return []map[string]interface{}{}
	}

	spec := map[string]interface{}{}

	if outputConfig.MonitoringOutputs != nil {
		spec["monitoring_outputs"] = flattenMonitoringOutputs(outputConfig.MonitoringOutputs)
	}

	return []map[string]interface{}{spec}
}

func flattenMonitoringOutputs(list []*sagemaker.MonitoringOutput) []map[string]interface{} {
	containers := make([]map[string]interface{}, 0, len(list))

	for _, lRaw := range list {
		monitoringOutput := make(map[string]interface{})
		monitoringOutput["s3_output"] = flattenS3Output(lRaw.S3Output)
		containers = append(containers, monitoringOutput)
	}

	return containers
}

func flattenS3Output(s3Output *sagemaker.MonitoringS3Output) []map[string]interface{} {
	if s3Output == nil {
		return []map[string]interface{}{}
	}

	spec := map[string]interface{}{}

	if s3Output.LocalPath != nil {
		spec["local_path"] = aws.StringValue(s3Output.LocalPath)
	}

	if s3Output.S3Uri != nil {
		spec["s3_uri"] = aws.StringValue(s3Output.S3Uri)
	}

	return []map[string]interface{}{spec}
}

func flattenJobResources(jobResources *sagemaker.MonitoringResources) []map[string]interface{} {
	if jobResources == nil {
		return []map[string]interface{}{}
	}

	spec := map[string]interface{}{}

	if jobResources.ClusterConfig != nil {
		spec["cluster_config"] = flattenClusterConfig(jobResources.ClusterConfig)
	}

	return []map[string]interface{}{spec}
}

func flattenClusterConfig(clusterConfig *sagemaker.MonitoringClusterConfig) []map[string]interface{} {
	if clusterConfig == nil {
		return []map[string]interface{}{}
	}

	spec := map[string]interface{}{}

	if clusterConfig.InstanceCount != nil {
		spec["instance_count"] = aws.Int64Value(clusterConfig.InstanceCount)
	}

	if clusterConfig.InstanceType != nil {
		spec["instance_type"] = aws.StringValue(clusterConfig.InstanceType)
	}

	if clusterConfig.VolumeSizeInGB != nil {
		spec["volume_size_in_gb"] = aws.Int64Value(clusterConfig.VolumeSizeInGB)
	}

	return []map[string]interface{}{spec}
}

func resourceDataQualityJobDefinitionUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).SageMakerConn()

	if d.HasChange("tags_all") {
		o, n := d.GetChange("tags_all")

		if err := UpdateTags(ctx, conn, d.Get("arn").(string), o, n); err != nil {
			return sdkdiag.AppendErrorf(diags, "updating SageMaker Data Quality Job Definition (%s) tags: %s", d.Id(), err)
		}
	}
	return append(diags, resourceEndpointConfigurationRead(ctx, d, meta)...)
}

func resourceDataQualityJobDefinitionDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).SageMakerConn()

	deleteOpts := &sagemaker.DeleteDataQualityJobDefinitionInput{
		JobDefinitionName: aws.String(d.Id()),
	}
	log.Printf("[INFO] Deleting SageMaker Data Quality Job Definition : %s", d.Id())

	_, err := conn.DeleteDataQualityJobDefinitionWithContext(ctx, deleteOpts)

	if tfawserr.ErrMessageContains(err, "ValidationException", "Could not find data quality job definition") {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting SageMaker Data Quality Job Definition (%s): %s", d.Id(), err)
	}

	return diags
}

func expandDataQualityAppSpecification(configured []interface{}) *sagemaker.DataQualityAppSpecification {
	if len(configured) == 0 {
		return nil
	}

	m := configured[0].(map[string]interface{})

	c := &sagemaker.DataQualityAppSpecification{}

	if v, ok := m["image_uri"].(string); ok && v != "" {
		c.ImageUri = aws.String(v)
	}

	return c
}

func expandDataQualityJobInput(configured []interface{}) *sagemaker.DataQualityJobInput {
	if len(configured) == 0 {
		return nil
	}

	m := configured[0].(map[string]interface{})

	c := &sagemaker.DataQualityJobInput{}

	if v, ok := m["endpoint_input"].([]interface{}); ok && len(v) > 0 {
		c.EndpointInput = expandEndpointInput(v)
	}

	return c
}

func expandEndpointInput(configured []interface{}) *sagemaker.EndpointInput {
	if len(configured) == 0 {
		return nil
	}

	m := configured[0].(map[string]interface{})

	c := &sagemaker.EndpointInput{}

	if v, ok := m["endpoint_name"].(string); ok && v != "" {
		c.EndpointName = aws.String(v)
	}

	if v, ok := m["local_path"].(string); ok && v != "" {
		c.LocalPath = aws.String(v)
	}

	return c
}

func expandDataQualityJobOutputConfig(configured []interface{}) *sagemaker.MonitoringOutputConfig {
	if len(configured) == 0 {
		return nil
	}

	m := configured[0].(map[string]interface{})

	c := &sagemaker.MonitoringOutputConfig{}

	if v, ok := m["monitoring_outputs"].([]interface{}); ok && len(v) > 0 {
		c.MonitoringOutputs = expandMonitoringOutputs(v)
	}

	return c
}

func expandMonitoringOutputs(configured []interface{}) []*sagemaker.MonitoringOutput {
	containers := make([]*sagemaker.MonitoringOutput, 0, len(configured))

	for _, lRaw := range configured {
		data := lRaw.(map[string]interface{})

		l := &sagemaker.MonitoringOutput{
			S3Output: expandS3Output(data["s3_output"].([]interface{})),
		}
		containers = append(containers, l)
	}

	return containers
}

func expandS3Output(configured []interface{}) *sagemaker.MonitoringS3Output {
	if len(configured) == 0 {
		return nil
	}

	m := configured[0].(map[string]interface{})

	c := &sagemaker.MonitoringS3Output{}

	if v, ok := m["local_path"].(string); ok && v != "" {
		c.LocalPath = aws.String(v)
	}

	if v, ok := m["s3_uri"].(string); ok && v != "" {
		c.S3Uri = aws.String(v)
	}

	return c
}

func expandJobResources(configured []interface{}) *sagemaker.MonitoringResources {
	if len(configured) == 0 {
		return nil
	}

	m := configured[0].(map[string]interface{})

	c := &sagemaker.MonitoringResources{}

	if v, ok := m["cluster_config"].([]interface{}); ok && len(v) > 0 {
		c.ClusterConfig = expandClusterConfig(v)
	}

	return c
}

func expandClusterConfig(configured []interface{}) *sagemaker.MonitoringClusterConfig {
	if len(configured) == 0 {
		return nil
	}

	m := configured[0].(map[string]interface{})

	c := &sagemaker.MonitoringClusterConfig{}

	if v, ok := m["instance_count"].(int); ok && v > 0 {
		c.InstanceCount = aws.Int64(int64(v))
	}
	if v, ok := m["instance_type"].(string); ok && v != "" {
		c.InstanceType = aws.String(v)
	}

	if v, ok := m["volume_size_in_gb"].(int); ok && v > 0 {
		c.VolumeSizeInGB = aws.Int64(int64(v))
	}

	return c
}
