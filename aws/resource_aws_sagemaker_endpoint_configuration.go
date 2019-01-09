package aws

import (
	"fmt"
	"log"
	"time"

	"github.com/hashicorp/terraform/helper/validation"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/sagemaker"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/helper/schema"
)

func resourceAwsSagemakerEndpointConfiguration() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsSagemakerEndpointConfigurationCreate,
		Read:   resourceAwsSagemakerEndpointConfigurationRead,
		Update: resourceAwsSagemakerEndpointConfigurationUpdate,
		Delete: resourceAwsSagemakerEndpointConfigurationDelete,
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
				Optional:     true,
				Computed:     true,
				ForceNew:     true,
				ValidateFunc: validateSagemakerName,
			},

			"production_variants": {
				Type:     schema.TypeList,
				Required: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"variant_name": {
							Type:     schema.TypeString,
							Optional: true,
							Computed: true,
							ForceNew: true,
						},

						"model_name": {
							Type:     schema.TypeString,
							Required: true,
							ForceNew: true,
						},

						"initial_instance_count": {
							Type:         schema.TypeInt,
							Required:     true,
							ForceNew:     true,
							ValidateFunc: validation.IntAtLeast(1),
						},

						"instance_type": {
							Type:     schema.TypeString,
							Required: true,
							ForceNew: true,
						},

						"initial_variant_weight": {
							Type:         schema.TypeFloat,
							Optional:     true,
							ForceNew:     true,
							ValidateFunc: FloatAtLeast(0),
						},

						"accelerator_type": {
							Type:     schema.TypeString,
							Optional: true,
							ForceNew: true,
						},
					},
				},
			},

			"kms_key_id": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
			},

			"tags": tagsSchema(),
		},
	}
}

func resourceAwsSagemakerEndpointConfigurationCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).sagemakerconn

	var name string
	if v, ok := d.GetOk("name"); ok {
		name = v.(string)
	} else {
		name = resource.UniqueId()
	}

	createOpts := &sagemaker.CreateEndpointConfigInput{
		EndpointConfigName: aws.String(name),
	}

	prodVariants, err := expandProductionVariants(d.Get("production_variants").([]interface{}))
	if err != nil {
		return err
	}
	createOpts.ProductionVariants = prodVariants

	if v, ok := d.GetOk("kms_key_id"); ok {
		createOpts.KmsKeyId = aws.String(v.(string))
	}

	if v, ok := d.GetOk("tags"); ok {
		createOpts.Tags = tagsFromMapSagemaker(v.(map[string]interface{}))
	}

	log.Printf("[DEBUG] Sagemaker endpoint configuration create config: %#v", *createOpts)
	_, err = conn.CreateEndpointConfig(createOpts)
	if err != nil {
		return fmt.Errorf("error creating Sagemaker endpoint configuration: %s", err)
	}
	d.SetId(name)

	return resourceAwsSagemakerEndpointConfigurationRead(d, meta)
}

func resourceAwsSagemakerEndpointConfigurationRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).sagemakerconn

	request := &sagemaker.DescribeEndpointConfigInput{
		EndpointConfigName: aws.String(d.Id()),
	}

	endpointConfig, err := conn.DescribeEndpointConfig(request)
	if err != nil {
		if sagemakerErr, ok := err.(awserr.Error); ok && sagemakerErr.Code() == "ValidationException" {
			log.Printf("[INFO] unable to find the sagemaker endpoint configuration resource and therefore it is removed from the state: %s", d.Id())
			d.SetId("")
			return nil
		}
		return fmt.Errorf("error reading Sagemaker endpoint configuration %s: %s", d.Id(), err)
	}

	if err := d.Set("arn", endpointConfig.EndpointConfigArn); err != nil {
		return err
	}
	if err := d.Set("name", endpointConfig.EndpointConfigName); err != nil {
		return err
	}
	if err := d.Set("production_variants", flattenProductionVariants(endpointConfig.ProductionVariants)); err != nil {
		return err
	}
	if err := d.Set("kms_key_id", endpointConfig.KmsKeyId); err != nil {
		return err
	}

	tagsOutput, err := conn.ListTags(&sagemaker.ListTagsInput{
		ResourceArn: endpointConfig.EndpointConfigArn,
	})
	if err != nil {
		return fmt.Errorf("error listing tags of Sagemaker endpoint configuration %s: %s", d.Id(), err)
	}
	if err := d.Set("tags", tagsToMapSagemaker(tagsOutput.Tags)); err != nil {
		return err
	}
	return nil
}

func resourceAwsSagemakerEndpointConfigurationUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).sagemakerconn

	d.Partial(true)

	if err := setSagemakerTags(conn, d); err != nil {
		return err
	} else {
		d.SetPartial("tags")
	}

	d.Partial(false)

	return resourceAwsSagemakerEndpointConfigurationRead(d, meta)
}

func resourceAwsSagemakerEndpointConfigurationDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).sagemakerconn

	deleteOpts := &sagemaker.DeleteEndpointConfigInput{
		EndpointConfigName: aws.String(d.Id()),
	}
	log.Printf("[INFO] Deleting Sagemaker endpoint configuration: %s", d.Id())

	return resource.Retry(5*time.Minute, func() *resource.RetryError {
		_, err := conn.DeleteEndpointConfig(deleteOpts)
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

		return resource.NonRetryableError(fmt.Errorf("Error deleting Sagemaker endpoint configuration: %s", err))
	})
}

func expandProductionVariants(configured []interface{}) ([]*sagemaker.ProductionVariant, error) {
	containers := make([]*sagemaker.ProductionVariant, 0, len(configured))

	for _, lRaw := range configured {
		data := lRaw.(map[string]interface{})

		l := &sagemaker.ProductionVariant{
			InstanceType:         aws.String(data["instance_type"].(string)),
			ModelName:            aws.String(data["model_name"].(string)),
			InitialInstanceCount: aws.Int64(int64(data["initial_instance_count"].(int))),
		}

		if v, ok := data["variant_name"]; ok {
			l.VariantName = aws.String(v.(string))
		} else {
			l.VariantName = aws.String(resource.UniqueId())
		}

		if v, ok := data["initial_variant_weight"]; ok {
			l.InitialVariantWeight = aws.Float64(v.(float64))
		} else {
			l.InitialVariantWeight = aws.Float64(1)
		}

		if v, ok := data["accelerator_type"]; ok && v.(string) != "" {
			l.AcceleratorType = aws.String(data["accelerator_type"].(string))
		}

		containers = append(containers, l)
	}

	return containers, nil
}

func flattenProductionVariants(list []*sagemaker.ProductionVariant) []map[string]interface{} {
	result := make([]map[string]interface{}, 0, len(list))

	for _, i := range list {
		l := map[string]interface{}{
			"instance_type":          *i.InstanceType,
			"model_name":             *i.ModelName,
			"initial_instance_count": *i.InitialInstanceCount,
		}
		if i.VariantName != nil {
			l["variant_name"] = *i.VariantName
		}
		if i.InitialVariantWeight != nil {
			l["initial_variant_weight"] = *i.InitialVariantWeight
		}
		if i.AcceleratorType != nil {
			l["accelerator_type"] = *i.AcceleratorType
		}

		result = append(result, l)
	}
	return result
}
