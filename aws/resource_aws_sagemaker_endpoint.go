package aws

import (
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/sagemaker"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/helper/schema"
	"log"
	"time"
)

func resourceAwsSagemakerEndpoint() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsSagemakerEndpointCreate,
		Read:   resourceAwsSagemakerEndpointRead,
		Update: resourceAwsSagemakerEndpointUpdate,
		Delete: resourceAwsSagemakerEndpointDelete,
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

			"endpoint_config_name": {
				Type:     schema.TypeString,
				Required: true,
			},

			"creation_time": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"last_modified_time": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"tags": tagsSchema(),
		},
	}
}

func resourceAwsSagemakerEndpointCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).sagemakerconn

	var name string
	if v, ok := d.GetOk("name"); ok {
		name = v.(string)
	} else {
		name = resource.UniqueId()
	}

	createOpts := &sagemaker.CreateEndpointInput{
		EndpointName:       aws.String(name),
		EndpointConfigName: aws.String(d.Get("endpoint_config_name").(string)),
	}

	log.Printf("[DEBUG] SageMaker endpoint create config: %#v", *createOpts)
	endpoint, err := conn.CreateEndpoint(createOpts)
	if err != nil {
		return fmt.Errorf("error creating SageMaker endpoint: %s", err)
	}

	d.SetId(name)
	if err := d.Set("arn", endpoint.EndpointArn); err != nil {
		return err
	}
	log.Printf("[INFO] SageMaker endpoint ID: %s", d.Id())

	log.Printf("[DEBUG] Waiting for SageMaker endpoint (%s) to become available", d.Id())
	stateConf := &resource.StateChangeConf{
		Pending: []string{"Creating"},
		Target:  []string{"InService"},
		Refresh: SagemakerEndpointStateRefreshFunc(conn, d.Id()),
		Timeout: 15 * time.Minute,
	}
	if _, err := stateConf.WaitForState(); err != nil {
		return fmt.Errorf(
			"error while waiting for SageMaker endpoint (%s) to become available: %s",
			d.Id(), err)
	}

	return resourceAwsSagemakerEndpointUpdate(d, meta)
}

func resourceAwsSagemakerEndpointRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).sagemakerconn

	endpointRaw, _, err := SagemakerEndpointStateRefreshFunc(conn, d.Id())()
	if err != nil {
		if sagemakerErr, ok := err.(awserr.Error); ok && sagemakerErr.Code() == "ValidationException" {
			d.SetId("")
			return nil
		}
		return err
	}
	if endpointRaw == nil {
		d.SetId("")
		return nil
	}

	endpoint := endpointRaw.(*sagemaker.DescribeEndpointOutput)

	if err := d.Set("name", endpoint.EndpointName); err != nil {
		return err
	}
	if err := d.Set("endpoint_config_name", endpoint.EndpointConfigName); err != nil {
		return err
	}
	if err := d.Set("creation_time", endpoint.CreationTime.Format(time.RFC3339)); err != nil {
		return err
	}
	if err := d.Set("last_modified_time", endpoint.LastModifiedTime.Format(time.RFC3339)); err != nil {
		return err
	}
	if err := d.Set("arn", endpoint.EndpointArn); err != nil {
		return err
	}

	tagsOutput, err := conn.ListTags(&sagemaker.ListTagsInput{
		ResourceArn: endpoint.EndpointArn,
	})

	d.Set("tags", tagsToMapSagemaker(tagsOutput.Tags))

	return nil
}

func resourceAwsSagemakerEndpointUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).sagemakerconn

	d.Partial(true)

	if err := setSagemakerTags(conn, d); err != nil {
		return err
	}
	d.SetPartial("tags")

	if d.HasChange("endpoint_config_name") && !d.IsNewResource() {
		modifyOpts := &sagemaker.UpdateEndpointInput{
			EndpointName:       aws.String(d.Id()),
			EndpointConfigName: aws.String(d.Get("endpoint_config_name").(string)),
		}
		log.Printf(
			"[INFO] Modifying endpoint_config_name attribute for %s: %#v",
			d.Id(), modifyOpts)
		if _, err := conn.UpdateEndpoint(modifyOpts); err != nil {
			return err
		}
		d.SetPartial("endpoint_config_name")

		log.Printf("[DEBUG] Waiting for SageMaker endpoint (%s) to be updated", d.Id())
		stateConf := &resource.StateChangeConf{
			Pending: []string{"Updating"},
			Target:  []string{"InService"},
			Refresh: SagemakerEndpointStateRefreshFunc(conn, d.Id()),
			Timeout: 15 * time.Minute,
		}
		if _, err := stateConf.WaitForState(); err != nil {
			return fmt.Errorf(
				"error updating SageMaker endpoint (%s): %s",
				d.Id(), err)
		}
	}

	d.Partial(false)

	return resourceAwsSagemakerEndpointRead(d, meta)
}

func resourceAwsSagemakerEndpointDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).sagemakerconn

	deleteEndpointOpts := &sagemaker.DeleteEndpointInput{
		EndpointName: aws.String(d.Id()),
	}
	log.Printf("[INFO] Deleting Sagemaker endpoint: %s", d.Id())

	return resource.Retry(5*time.Minute, func() *resource.RetryError {
		_, err := conn.DeleteEndpoint(deleteEndpointOpts)
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

		return resource.NonRetryableError(fmt.Errorf("Error deleting Sagemaker endpoint: %s", err))
	})
}

func SagemakerEndpointStateRefreshFunc(conn *sagemaker.SageMaker, name string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		describeEndpointOpts := &sagemaker.DescribeEndpointInput{
			EndpointName: aws.String(name),
		}
		endpoint, err := conn.DescribeEndpoint(describeEndpointOpts)
		if err != nil {
			if sagemakerErr, ok := err.(awserr.Error); ok && sagemakerErr.Code() == "ResourceNotFound" {
				endpoint = nil
			} else {
				log.Printf("Error on SagemakerEndpointStateRefresh: %s", err)
				return nil, "", err
			}
		}

		if endpoint == nil {
			return nil, "", nil
		}

		return endpoint, *endpoint.EndpointStatus, nil
	}
}
