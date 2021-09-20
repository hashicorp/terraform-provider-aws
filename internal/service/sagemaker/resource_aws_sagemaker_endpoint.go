package sagemaker

import (
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/sagemaker"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

func ResourceEndpoint() *schema.Resource {
	return &schema.Resource{
		Create: resourceEndpointCreate,
		Read:   resourceEndpointRead,
		Update: resourceEndpointUpdate,
		Delete: resourceEndpointDelete,
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
				ValidateFunc: validName,
			},

			"endpoint_config_name": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validName,
			},

			"tags":     tftags.TagsSchema(),
			"tags_all": tftags.TagsSchemaComputed(),
		},

		CustomizeDiff: verify.SetTagsDiff,
	}
}

func resourceEndpointCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).SageMakerConn
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	tags := defaultTagsConfig.MergeTags(tftags.New(d.Get("tags").(map[string]interface{})))

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

	if len(tags) > 0 {
		createOpts.Tags = tags.IgnoreAws().SagemakerTags()
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

	return resourceEndpointRead(d, meta)
}

func resourceEndpointRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).SageMakerConn
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	ignoreTagsConfig := meta.(*conns.AWSClient).IgnoreTagsConfig

	describeInput := &sagemaker.DescribeEndpointInput{
		EndpointName: aws.String(d.Id()),
	}

	endpoint, err := conn.DescribeEndpoint(describeInput)
	if err != nil {
		if tfawserr.ErrMessageContains(err, "ValidationException", "") {
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

	tags, err := tftags.SagemakerListTags(conn, aws.StringValue(endpoint.EndpointArn))
	if err != nil {
		return fmt.Errorf("error listing tags for Sagemaker Endpoint (%s): %s", d.Id(), err)
	}

	tags = tags.IgnoreAws().IgnoreConfig(ignoreTagsConfig)

	//lintignore:AWSR002
	if err := d.Set("tags", tags.RemoveDefaultConfig(defaultTagsConfig).Map()); err != nil {
		return fmt.Errorf("error setting tags: %w", err)
	}

	if err := d.Set("tags_all", tags.Map()); err != nil {
		return fmt.Errorf("error setting tags_all: %w", err)
	}

	return nil
}

func resourceEndpointUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).SageMakerConn

	if d.HasChange("tags_all") {
		o, n := d.GetChange("tags_all")

		if err := tftags.SagemakerUpdateTags(conn, d.Get("arn").(string), o, n); err != nil {
			return fmt.Errorf("error updating Sagemaker Endpoint (%s) tags: %s", d.Id(), err)
		}
	}

	if d.HasChange("endpoint_config_name") {
		modifyOpts := &sagemaker.UpdateEndpointInput{
			EndpointName:       aws.String(d.Id()),
			EndpointConfigName: aws.String(d.Get("endpoint_config_name").(string)),
		}

		log.Printf("[INFO] Modifying endpoint_config_name attribute for %s: %#v", d.Id(), modifyOpts)
		if _, err := conn.UpdateEndpoint(modifyOpts); err != nil {
			return fmt.Errorf("error updating SageMaker Endpoint (%s): %s", d.Id(), err)
		}

		describeInput := &sagemaker.DescribeEndpointInput{
			EndpointName: aws.String(d.Id()),
		}

		err := conn.WaitUntilEndpointInService(describeInput)
		if err != nil {
			return fmt.Errorf("error waiting for SageMaker Endpoint (%s) to be in service: %s", d.Id(), err)
		}
	}

	return resourceEndpointRead(d, meta)
}

func resourceEndpointDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).SageMakerConn

	deleteEndpointOpts := &sagemaker.DeleteEndpointInput{
		EndpointName: aws.String(d.Id()),
	}
	log.Printf("[INFO] Deleting SageMaker Endpoint: %s", d.Id())

	_, err := conn.DeleteEndpoint(deleteEndpointOpts)

	if tfawserr.ErrMessageContains(err, "ValidationException", "") {
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
