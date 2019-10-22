package aws

import (
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/sagemaker"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
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
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validateSagemakerName,
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

	if v, ok := d.GetOk("tags"); ok {
		createOpts.Tags = tagsFromMapSagemaker(v.(map[string]interface{}))
	}

	log.Printf("[DEBUG] SageMaker Endpoint create config: %#v", *createOpts)
	_, err := conn.CreateEndpoint(createOpts)
	if err != nil {
		return fmt.Errorf("error creating SageMaker Endpoint: %s", err)
	}

	d.SetId(name)

	describeInput := &sagemaker.DescribeEndpointInput{
		EndpointName: aws.String(name),
	}

	if err := conn.WaitUntilEndpointInService(describeInput); err != nil {
		return fmt.Errorf("error waiting for SageMaker Endpoint (%s) to be in service: %s", name, err)
	}

	return resourceAwsSagemakerEndpointRead(d, meta)
}

func resourceAwsSagemakerEndpointRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).sagemakerconn

	describeInput := &sagemaker.DescribeEndpointInput{
		EndpointName: aws.String(d.Id()),
	}

	endpoint, err := conn.DescribeEndpoint(describeInput)
	if err != nil {
		if isAWSErr(err, "ValidationException", "") {
			log.Printf("[INFO] unable to find the SageMaker Endpoint resource and therefore it is removed from the state: %s", d.Id())
			d.SetId("")
			return nil
		}
		return err
	}
	if aws.StringValue(endpoint.EndpointStatus) == sagemaker.EndpointStatusDeleting {
		log.Printf("[WARN] SageMaker Endpoint (%s) is deleting, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err := d.Set("name", endpoint.EndpointName); err != nil {
		return err
	}
	if err := d.Set("endpoint_config_name", endpoint.EndpointConfigName); err != nil {
		return err
	}

	if err := d.Set("arn", endpoint.EndpointArn); err != nil {
		return err
	}

	tagsOutput, err := conn.ListTags(&sagemaker.ListTagsInput{
		ResourceArn: endpoint.EndpointArn,
	})
	if err != nil {
		return fmt.Errorf("error listing tags for SageMaker Endpoint (%s): %s", d.Id(), err)
	}

	if err := d.Set("tags", tagsToMapSagemaker(tagsOutput.Tags)); err != nil {
		return err
	}

	return nil
}

func resourceAwsSagemakerEndpointUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).sagemakerconn

	d.Partial(true)

	if err := setSagemakerTags(conn, d); err != nil {
		return err
	}
	d.SetPartial("tags")

	if d.HasChange("endpoint_config_name") {
		modifyOpts := &sagemaker.UpdateEndpointInput{
			EndpointName:       aws.String(d.Id()),
			EndpointConfigName: aws.String(d.Get("endpoint_config_name").(string)),
		}

		log.Printf("[INFO] Modifying endpoint_config_name attribute for %s: %#v", d.Id(), modifyOpts)
		if _, err := conn.UpdateEndpoint(modifyOpts); err != nil {
			return fmt.Errorf("error updating SageMaker Endpoint (%s): %s", d.Id(), err)
		}
		d.SetPartial("endpoint_config_name")

		describeInput := &sagemaker.DescribeEndpointInput{
			EndpointName: aws.String(d.Id()),
		}

		err := conn.WaitUntilEndpointInService(describeInput)
		if err != nil {
			return fmt.Errorf("error waiting for SageMaker Endpoint (%s) to be in service: %s", d.Id(), err)
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
	log.Printf("[INFO] Deleting SageMaker Endpoint: %s", d.Id())

	_, err := conn.DeleteEndpoint(deleteEndpointOpts)

	if isAWSErr(err, "ValidationException", "") {
		return nil
	}

	if err != nil {
		return fmt.Errorf("error deleting SageMaker Endpoint (%s): %s", d.Id(), err)
	}

	describeInput := &sagemaker.DescribeEndpointInput{
		EndpointName: aws.String(d.Id()),
	}

	if err := conn.WaitUntilEndpointDeleted(describeInput); err != nil {
		return fmt.Errorf("error waiting for SageMaker Endpoint (%s) to be deleted: %s", d.Id(), err)
	}

	return nil
}
