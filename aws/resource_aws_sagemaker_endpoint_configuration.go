package aws

import (
	"bytes"
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/sagemaker"
	"github.com/hashicorp/terraform/helper/hashcode"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/helper/schema"
	"log"
	"time"
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
				Type:     schema.TypeSet,
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
							Type:     schema.TypeInt,
							Required: true,
							ForceNew: true,
						},

						"instance_type": {
							Type:     schema.TypeString,
							Required: true,
							ForceNew: true,
						},

						"initial_variant_weight": {
							Type:     schema.TypeFloat,
							Required: true,
							ForceNew: true,
						},
					},
				},
				Set: resourceAwsSagmakerEndpointConfigEntryHash,
			},

			"kms_key_id": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
			},

			"creation_time": {
				Type:     schema.TypeString,
				Computed: true,
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

	prodVariants, err := expandProductionVariants(d.Get("production_variants").(*schema.Set).List())
	if err != nil {
		return err
	}
	createOpts.ProductionVariants = prodVariants

	if v, ok := d.GetOk("kms_key_id"); ok {
		createOpts.KmsKeyId = aws.String(v.(string))
	}

	log.Printf("[DEBUG] Sagemaker endpoint configuration create config: %#v", *createOpts)
	resp, err := conn.CreateEndpointConfig(createOpts)
	if err != nil {
		return fmt.Errorf("Error creating Sagemaker endpoint configuration: %s", err)
	}

	d.SetId(name)
	if err := d.Set("arn", resp.EndpointConfigArn); err != nil {
		return err
	}

	return resourceAwsSagemakerEndpointConfigurationUpdate(d, meta)
}

func resourceAwsSagemakerEndpointConfigurationRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).sagemakerconn

	request := &sagemaker.DescribeEndpointConfigInput{
		EndpointConfigName: aws.String(d.Id()),
	}

	endpointConfig, err := conn.DescribeEndpointConfig(request)
	if err != nil {
		if sagemakerErr, ok := err.(awserr.Error); ok && sagemakerErr.Code() == "ResourceNotFound" {
			d.SetId("")
			return nil
		}
		return fmt.Errorf("Error reading Sagemaker endpoint configuration %s: %s", d.Id(), err)
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
	if err := d.Set("creation_time", endpointConfig.CreationTime.Format(time.RFC3339)); err != nil {
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

		var name string
		if v, ok := data["variant_name"]; ok {
			name = v.(string)
		} else {
			name = resource.UniqueId()
		}

		l := &sagemaker.ProductionVariant{
			VariantName:          aws.String(name),
			InstanceType:         aws.String(data["instance_type"].(string)),
			ModelName:            aws.String(data["model_name"].(string)),
			InitialVariantWeight: aws.Float64(float64(data["initial_variant_weight"].(float64))),
			InitialInstanceCount: aws.Int64(int64(data["initial_instance_count"].(int))),
		}
		containers = append(containers, l)
	}

	return containers, nil
}

func flattenProductionVariants(list []*sagemaker.ProductionVariant) []map[string]interface{} {
	result := make([]map[string]interface{}, 0, len(list))
	for _, i := range list {
		l := map[string]interface{}{
			"variant_name":           *i.VariantName,
			"instance_type":          *i.InstanceType,
			"model_name":             *i.ModelName,
			"initial_variant_weight": *i.InitialVariantWeight,
			"initial_instance_count": *i.InitialInstanceCount,
		}
		result = append(result, l)
	}
	return result
}

func resourceAwsSagmakerEndpointConfigEntryHash(v interface{}) int {
	var buf bytes.Buffer
	m := v.(map[string]interface{})
	buf.WriteString(fmt.Sprintf("%s-", m["variant_name"].(string)))
	buf.WriteString(fmt.Sprintf("%s-", m["model_name"].(string)))
	buf.WriteString(fmt.Sprintf("%s-", m["instance_type"].(string)))
	buf.WriteString(fmt.Sprintf("%f-", m["initial_variant_weight"].(float64)))
	buf.WriteString(fmt.Sprintf("%d-", m["initial_instance_count"].(int)))

	return hashcode.String(buf.String())
}
