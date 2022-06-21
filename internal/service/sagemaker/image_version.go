package sagemaker

import (
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/sagemaker"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
)

func ResourceImageVersion() *schema.Resource {
	return &schema.Resource{
		Create: resourceImageVersionCreate,
		Read:   resourceImageVersionRead,
		Delete: resourceImageVersionDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"base_image": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"container_image": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"image_arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"image_name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"version": {
				Type:     schema.TypeInt,
				Computed: true,
			},
		},
	}
}

func resourceImageVersionCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).SageMakerConn

	name := d.Get("image_name").(string)
	input := &sagemaker.CreateImageVersionInput{
		ImageName: aws.String(name),
		BaseImage: aws.String(d.Get("base_image").(string)),
	}

	_, err := conn.CreateImageVersion(input)
	if err != nil {
		return fmt.Errorf("error creating SageMaker Image Version %s: %w", name, err)
	}

	d.SetId(name)

	if _, err := WaitImageVersionCreated(conn, d.Id()); err != nil {
		return fmt.Errorf("error waiting for SageMaker Image Version (%s) to be created: %w", d.Id(), err)
	}

	return resourceImageVersionRead(d, meta)
}

func resourceImageVersionRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).SageMakerConn

	image, err := FindImageVersionByName(conn, d.Id())
	if err != nil {
		if tfawserr.ErrMessageContains(err, sagemaker.ErrCodeResourceNotFound, "does not exist") {
			d.SetId("")
			log.Printf("[WARN] Unable to find SageMaker Image Version (%s); removing from state", d.Id())
			return nil
		}
		return fmt.Errorf("error reading SageMaker Image Version (%s): %w", d.Id(), err)

	}

	d.Set("arn", image.ImageVersionArn)
	d.Set("base_image", image.BaseImage)
	d.Set("image_arn", image.ImageArn)
	d.Set("container_image", image.ContainerImage)
	d.Set("version", image.Version)
	d.Set("image_name", d.Id())

	return nil
}

func resourceImageVersionDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).SageMakerConn

	input := &sagemaker.DeleteImageVersionInput{
		ImageName: aws.String(d.Id()),
		Version:   aws.Int64(int64(d.Get("version").(int))),
	}

	if _, err := conn.DeleteImageVersion(input); err != nil {
		if tfawserr.ErrMessageContains(err, sagemaker.ErrCodeResourceNotFound, "does not exist") {
			return nil
		}
		return fmt.Errorf("error deleting SageMaker Image Version (%s): %w", d.Id(), err)
	}

	if _, err := WaitImageVersionDeleted(conn, d.Id()); err != nil {
		if tfawserr.ErrMessageContains(err, sagemaker.ErrCodeResourceNotFound, "does not exist") {
			return nil
		}
		return fmt.Errorf("error waiting for SageMaker Image Version (%s) to delete: %w", d.Id(), err)
	}

	return nil
}
