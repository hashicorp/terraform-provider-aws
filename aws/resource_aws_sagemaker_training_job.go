package aws

import (
	"fmt"
	"log"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/sagemaker"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/helper/schema"
	"github.com/hashicorp/terraform/helper/validation"
)

func resourceAwsSagemakerTrainingJob() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsSagemakerTrainingJobCreate,
		Read:   resourceAwsSagemakerTrainingJobRead,
		Update: resourceAwsSagemakerTrainingJobUpdate,
		Delete: resourceAwsSagemakerTrainingJobDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"name": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validateSagemakerName,
			},

			"role_arn": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},

			"algorithm_specification": {
				Type:     schema.TypeList,
				MaxItems: 1,
				Required: true,
				ForceNew: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"image": {
							Type:     schema.TypeString,
							Required: true,
							ForceNew: true,
						},
						"input_mode": {
							Type:     schema.TypeString,
							Required: true,
							ForceNew: true,
							ValidateFunc: validation.StringInSlice([]string{
								sagemaker.TrainingInputModePipe,
								sagemaker.TrainingInputModeFile,
							}, false),
						},
					},
				},
			},

			"resource_config": {
				Type:     schema.TypeList,
				MaxItems: 1,
				Required: true,
				ForceNew: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"instance_type": {
							Type:     schema.TypeString,
							Required: true,
							ForceNew: true,
						},
						"instance_count": {
							Type:     schema.TypeInt,
							Required: true,
							ForceNew: true,
						},
						"volume_size_in_gb": {
							Type:     schema.TypeInt,
							Required: true,
							ForceNew: true,
						},
					},
				},
			},

			"stopping_condition": {
				Type:     schema.TypeList,
				MaxItems: 1,
				Optional: true,
				ForceNew: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"max_runtime_in_seconds": {
							Type:     schema.TypeInt,
							Optional: true,
							Default:  24 * 60 * 60, // 24 hours
							ForceNew: true,
						},
					},
				},
			},

			"hyper_parameters": {
				Type:     schema.TypeMap,
				Optional: true,
				ForceNew: true,
			},

			"input_data_config": {
				Type:     schema.TypeList,
				MinItems: 1,
				MaxItems: 8,
				Optional: true,
				ForceNew: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"name": {
							Type:     schema.TypeString,
							Required: true,
							ForceNew: true,
						},
						"data_source": {
							Type:     schema.TypeList,
							MaxItems: 1,
							Required: true,
							ForceNew: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"s3_data_source": {
										Type:     schema.TypeList,
										MaxItems: 1,
										Optional: true,
										ForceNew: true,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"s3_data_type": {
													Type:     schema.TypeString,
													Required: true,
													ForceNew: true,
												},
												"s3_uri": {
													Type:     schema.TypeString,
													Required: true,
													ForceNew: true,
												},
												"s3_data_distribution_type": {
													Type:     schema.TypeString,
													Required: true,
													ForceNew: true,
												},
											},
										},
									},
								},
							},
						},
						"content_type": {
							Type:     schema.TypeString,
							Optional: true,
							ForceNew: true,
						},
						"compression_type": {
							Type:     schema.TypeString,
							Optional: true,
							Default:  sagemaker.CompressionTypeNone,
							ForceNew: true,
							ValidateFunc: validation.StringInSlice([]string{
								sagemaker.CompressionTypeGzip,
								sagemaker.CompressionTypeNone,
							}, false),
						},
						"record_wrapper_type": {
							Type:     schema.TypeString,
							Optional: true,
							Default:  sagemaker.RecordWrapperNone,
							ForceNew: true,
							ValidateFunc: validation.StringInSlice([]string{
								sagemaker.RecordWrapperRecordIo,
								sagemaker.RecordWrapperNone,
							}, false),
						},
					},
				},
			},

			"output_data_config": {
				Type:     schema.TypeList,
				MaxItems: 1,
				Required: true,
				ForceNew: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"s3_output_path": {
							Type:     schema.TypeString,
							Required: true,
							ForceNew: true,
						},
						"kms_key_id": {
							Type:     schema.TypeString,
							Optional: true,
							ForceNew: true,
						},
					},
				},
			},
			"tags": tagsSchema(),
		},
	}
}

func resourceAwsSagemakerTrainingJobCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).sagemakerconn

	name := d.Get("name").(string)

	createOpts := &sagemaker.CreateTrainingJobInput{
		TrainingJobName: aws.String(name),
		RoleArn:         aws.String(d.Get("role_arn").(string)),
	}

	var algorithmSpecification = new(sagemaker.AlgorithmSpecification)
	if as, ok := d.GetOk("algorithm_specification"); ok {
		asPorperties := as.([]interface{})
		properties := asPorperties[0].(map[string]interface{})

		if v, ok := properties["image"]; ok {
			algorithmSpecification.TrainingImage = aws.String(v.(string))
		}
		if v, ok := properties["input_mode"]; ok {
			algorithmSpecification.TrainingInputMode = aws.String(v.(string))
		}
	}
	createOpts.AlgorithmSpecification = algorithmSpecification

	if hp, ok := d.GetOk("hyper_parameters"); ok {
		createOpts.HyperParameters = stringMapToPointers(hp.(map[string]interface{}))
	}

	if idc, ok := d.GetOk("input_data_config"); ok {
		createOpts.InputDataConfig = expandInputDataConfig(idc)
	}
	var outputDataConfig = new(sagemaker.OutputDataConfig)
	if odc, ok := d.GetOk("output_data_config"); ok {
		odcPorperties := odc.([]interface{})
		properties := odcPorperties[0].(map[string]interface{})

		if v, ok := properties["s3_output_path"]; ok {
			outputDataConfig.S3OutputPath = aws.String(v.(string))
		}
		if v, ok := properties["kms_key_id"]; ok {
			outputDataConfig.KmsKeyId = aws.String(v.(string))
		}
	}
	createOpts.OutputDataConfig = outputDataConfig

	var resourceConfig = new(sagemaker.ResourceConfig)
	if rc, ok := d.GetOk("resource_config"); ok {
		rcPorperties := rc.([]interface{})
		properties := rcPorperties[0].(map[string]interface{})

		if v, ok := properties["instance_type"]; ok {
			resourceConfig.InstanceType = aws.String(v.(string))
		}
		if v, ok := properties["instance_count"]; ok {
			resourceConfig.InstanceCount = aws.Int64(int64(v.(int)))
		}
		if v, ok := properties["volume_size_in_gb"]; ok {
			resourceConfig.VolumeSizeInGB = aws.Int64(int64(v.(int)))
		}
	}
	createOpts.ResourceConfig = resourceConfig

	var stoppingCondition = new(sagemaker.StoppingCondition)
	if sc, ok := d.GetOk("stopping_condition"); ok {
		scPorperties := sc.([]interface{})
		properties := scPorperties[0].(map[string]interface{})

		if v, ok := properties["max_runtime_in_seconds"]; ok {
			stoppingCondition.MaxRuntimeInSeconds = aws.Int64(int64(v.(int)))
		}
	}
	createOpts.StoppingCondition = stoppingCondition

	if v, ok := d.GetOk("tags"); ok {
		tagsIn := v.(map[string]interface{})
		createOpts.Tags = tagsFromMapSagemaker(tagsIn)
	}

	log.Printf("[DEBUG] sagemaker training job create config: %#v", *createOpts)
	_, err := conn.CreateTrainingJob(createOpts)
	if err != nil {
		return fmt.Errorf("error creating sagemaker training job: %s", err)
	}

	d.SetId(name)
	log.Printf("[INFO] sagemaker training job id: %s", d.Id())
	return resourceAwsSagemakerTrainingJobRead(d, meta)
}

func resourceAwsSagemakerTrainingJobRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).sagemakerconn

	describeTrainingInput := &sagemaker.DescribeTrainingJobInput{
		TrainingJobName: aws.String(d.Id()),
	}
	trainingJob, err := conn.DescribeTrainingJob(describeTrainingInput)
	if err != nil {
		if isAWSErr(err, "", "RecordNotFound") {
			log.Printf("[LOG] unable to find sagemaker training job %q; removing from state file", d.Id())
			d.SetId("")
			return nil
		}
		return fmt.Errorf("error finding sagemaker notebook instance %q: %s", d.Id(), err)

	}

	if err := d.Set("name", trainingJob.TrainingJobName); err != nil {
		return fmt.Errorf("error setting name for sagemaker training job %q: %s", d.Id(), err)
	}
	if err := d.Set("role_arn", trainingJob.RoleArn); err != nil {
		return fmt.Errorf("error setting role_arn for sagemaker training job %q: %s", d.Id(), err)
	}
	if err := d.Set("algorithm_specification", flattenAlgorithmSpecification(trainingJob.AlgorithmSpecification)); err != nil {
		return fmt.Errorf("error setting algorithm_specification for sagemaker training job %q: %s", d.Id(), err)
	}
	if err := d.Set("resource_config", flattenResourceConfig(trainingJob.ResourceConfig)); err != nil {
		return fmt.Errorf("error setting resource_config for sagemaker training job %q: %s", d.Id(), err)
	}
	if err := d.Set("stopping_condition", flattenStoppingCondition(trainingJob.StoppingCondition)); err != nil {
		return fmt.Errorf("error setting stopping_condition for sagemaker training job %q: %s", d.Id(), err)
	}
	if err := d.Set("hyper_parameters", pointersMapToStringList(trainingJob.HyperParameters)); err != nil {
		return fmt.Errorf("error setting hyper_parameters for sagemaker training job %q: %s", d.Id(), err)
	}
	if err := d.Set("input_data_config", flattenInputDataConfig(trainingJob.InputDataConfig)); err != nil {
		return fmt.Errorf("error setting input_data_config for sagemaker training job %q: %s", d.Id(), err)
	}
	if err := d.Set("output_data_config", flattenOutputDataConfig(trainingJob.OutputDataConfig)); err != nil {
		return fmt.Errorf("error setting output_data_config for sagemaker training job %q: %s", d.Id(), err)
	}

	if err := d.Set("arn", trainingJob.TrainingJobArn); err != nil {
		return fmt.Errorf("error setting arn for sagemaker training job %q: %s", d.Id(), err)
	}

	tagsOutput, err := conn.ListTags(&sagemaker.ListTagsInput{
		ResourceArn: trainingJob.TrainingJobArn,
	})

	if err != nil {
		log.Printf("[ERR] error reading tags: %s", err)
		return err
	}

	if err := d.Set("tags", tagsToMapSagemaker(tagsOutput.Tags)); err != nil {
		return fmt.Errorf("error setting tags for sagemaker training job %q: %s", d.Id(), err)
	}
	return nil
}

func resourceAwsSagemakerTrainingJobUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).sagemakerconn

	d.Partial(true)

	if err := setSagemakerTags(conn, d); err != nil {
		return err
	}

	d.SetPartial("tags")

	d.Partial(false)

	return resourceAwsSagemakerTrainingJobRead(d, meta)
}

func resourceAwsSagemakerTrainingJobDelete(d *schema.ResourceData, meta interface{}) error {
	// Deleting a training job actually means stopping the job since training jobs cannot be deleted
	conn := meta.(*AWSClient).sagemakerconn

	_, jobStatus, err := sagemakerTrainingJobStateRefreshFunc(conn, d.Id())()
	if err != nil {
		return err
	}

	if jobStatus != "InProgress" {
		log.Printf("[INFO] Sagemaker training job is not running: %s", d.Id())
		return nil
	}

	stopTrainingJobInput := &sagemaker.StopTrainingJobInput{
		TrainingJobName: aws.String(d.Id()),
	}
	log.Printf("[INFO] Stopping Sagemaker training job: %s", d.Id())

	return resource.Retry(5*time.Minute, func() *resource.RetryError {
		_, err := conn.StopTrainingJob(stopTrainingJobInput)
		if err == nil {
			return nil
		}

		sagemakerErr, ok := err.(awserr.Error)
		if !ok {
			return resource.NonRetryableError(err)
		}

		if sagemakerErr.Code() == "ResourceNotFound" {
			return resource.RetryableError(err)
		}

		return resource.NonRetryableError(fmt.Errorf("Error stopping Sagemaker training job: %s", err))
	})
}

func sagemakerTrainingJobStateRefreshFunc(conn *sagemaker.SageMaker, name string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		describeTrainingInput := &sagemaker.DescribeTrainingJobInput{
			TrainingJobName: aws.String(name),
		}
		trainingJob, err := conn.DescribeTrainingJob(describeTrainingInput)
		if err != nil {
			if sagemakerErr, ok := err.(awserr.Error); ok && sagemakerErr.Code() == "ResourceNotFound" {
				trainingJob = nil
			} else {
				log.Printf("Error on sagemakerTrainingJobStateRefreshFunc: %s", err)
				return nil, "", err
			}
		}

		if trainingJob == nil {
			return nil, "", nil
		}

		return trainingJob, *trainingJob.TrainingJobStatus, nil
	}
}

func flattenAlgorithmSpecification(as *sagemaker.AlgorithmSpecification) []map[string]interface{} {
	attrs := map[string]interface{}{}
	result := make([]map[string]interface{}, 0, 1)

	if as.TrainingImage != nil {
		attrs["image"] = *as.TrainingImage
	}
	if as.TrainingInputMode != nil {
		attrs["input_mode"] = *as.TrainingInputMode
	}

	result = append(result, attrs)

	return result
}

func flattenResourceConfig(rc *sagemaker.ResourceConfig) []map[string]interface{} {
	attrs := map[string]interface{}{}
	result := make([]map[string]interface{}, 0, 1)

	if rc.InstanceType != nil {
		attrs["instance_type"] = *rc.InstanceType
	}
	if rc.InstanceCount != nil {
		attrs["instance_count"] = *rc.InstanceCount
	}
	if rc.VolumeSizeInGB != nil {
		attrs["volume_size_in_gb"] = *rc.VolumeSizeInGB
	}

	result = append(result, attrs)

	return result
}

func flattenStoppingCondition(sc *sagemaker.StoppingCondition) []map[string]interface{} {
	attrs := map[string]interface{}{}
	result := make([]map[string]interface{}, 0, 1)

	if sc.MaxRuntimeInSeconds != nil {
		attrs["max_runtime_in_seconds"] = *sc.MaxRuntimeInSeconds
	}

	result = append(result, attrs)

	return result
}

func flattenInputDataConfig(cl []*sagemaker.Channel) []map[string]interface{} {
	result := make([]map[string]interface{}, 0, len(cl))

	for _, c := range cl {
		attrs := map[string]interface{}{}
		if c.ChannelName != nil {
			attrs["name"] = *c.ChannelName
		}
		if c.DataSource != nil {
			attrs["data_source"] = flattenInputDataConfigDataSource(c.DataSource)
		}
		if c.ContentType != nil {
			attrs["content_type"] = *c.ContentType
		}
		if c.CompressionType != nil {
			attrs["compression_type"] = *c.CompressionType
		}
		if c.RecordWrapperType != nil {
			attrs["record_wrapper_type"] = *c.RecordWrapperType
		}
		result = append(result, attrs)
	}

	return result
}

func flattenInputDataConfigDataSource(ds *sagemaker.DataSource) []map[string]interface{} {
	attrs := map[string]interface{}{}
	result := make([]map[string]interface{}, 0, 1)

	if ds.S3DataSource != nil {
		attrs["s3_data_source"] = flattenInputDataConfigS3DataSource(ds.S3DataSource)
	}

	result = append(result, attrs)

	return result
}

func flattenInputDataConfigS3DataSource(s3ds *sagemaker.S3DataSource) []map[string]interface{} {
	attrs := map[string]interface{}{}
	result := make([]map[string]interface{}, 0, 1)

	if s3ds.S3DataType != nil {
		attrs["s3_data_type"] = *s3ds.S3DataType
	}
	if s3ds.S3Uri != nil {
		attrs["s3_uri"] = *s3ds.S3Uri
	}
	if s3ds.S3DataDistributionType != nil {
		attrs["s3_data_distribution_type"] = *s3ds.S3DataDistributionType
	}

	result = append(result, attrs)

	return result
}

func flattenOutputDataConfig(odc *sagemaker.OutputDataConfig) []map[string]interface{} {
	attrs := map[string]interface{}{}
	result := make([]map[string]interface{}, 0, 1)

	if odc.S3OutputPath != nil {
		attrs["s3_output_path"] = *odc.S3OutputPath
	}
	if odc.KmsKeyId != nil {
		attrs["kms_key_id"] = *odc.KmsKeyId
	}

	result = append(result, attrs)

	return result
}

func expandInputDataConfig(idc interface{}) []*sagemaker.Channel {
	idcProperties := idc.([]interface{})
	inputDataConfig := make([]*sagemaker.Channel, len(idcProperties))

	for idx, idcProps := range idcProperties {
		idcItemProperties := idcProps.(map[string]interface{})

		var channel = new(sagemaker.Channel)
		if v, ok := idcItemProperties["name"]; ok {
			channel.ChannelName = aws.String(v.(string))
		}
		if ds, ok := idcItemProperties["data_source"]; ok {
			dsProperties := ds.([]interface{})
			dsItemProperties := dsProperties[0].(map[string]interface{})

			var dataSource = new(sagemaker.DataSource)
			if s3ds, ok := dsItemProperties["s3_data_source"]; ok {
				s3dsProperties := s3ds.([]interface{})
				s3dsItemProperties := s3dsProperties[0].(map[string]interface{})

				var s3DataSource = new(sagemaker.S3DataSource)
				if v, ok := s3dsItemProperties["s3_data_type"]; ok {
					s3DataSource.S3DataType = aws.String(v.(string))
				}
				if v, ok := s3dsItemProperties["s3_uri"]; ok {
					s3DataSource.S3Uri = aws.String(v.(string))
				}
				if v, ok := s3dsItemProperties["s3_data_distribution_type"]; ok {
					s3DataSource.S3DataDistributionType = aws.String(v.(string))
				}
				dataSource.S3DataSource = s3DataSource
			}
			channel.DataSource = dataSource
		}
		if v, ok := idcItemProperties["content_type"]; ok {
			channel.ContentType = aws.String(v.(string))
		}
		if v, ok := idcItemProperties["compression_type"]; ok {
			channel.CompressionType = aws.String(v.(string))
		}
		if v, ok := idcItemProperties["record_wrapper_type"]; ok {
			channel.RecordWrapperType = aws.String(v.(string))
		}
		inputDataConfig[idx] = channel
	}
	return inputDataConfig
}
