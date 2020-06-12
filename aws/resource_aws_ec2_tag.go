package aws

import (
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
)

func resourceAwsEc2Tag() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsEc2TagCreate,
		Read:   resourceAwsEc2TagRead,
		Delete: resourceAwsEc2TagDelete,

		Schema: map[string]*schema.Schema{
			"resource_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"key": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"value": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
		},
	}
}

func extractResourceIDAndKeyFromEc2TagID(id string) (string, string, error) {
	parts := strings.Split(id, ":")

	if len(parts) != 2 {
		return "", "", fmt.Errorf("Invalid resource ID; cannot look up resource: %s", id)
	}

	return parts[0], parts[1], nil
}

func resourceAwsEc2TagCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).ec2conn

	resourceID := d.Get("resource_id").(string)
	key := d.Get("key").(string)
	value := d.Get("value").(string)

	_, err := conn.CreateTags(&ec2.CreateTagsInput{
		Resources: []*string{aws.String(resourceID)},
		Tags: []*ec2.Tag{
			{
				Key:   aws.String(key),
				Value: aws.String(value),
			},
		},
	})

	if err != nil {
		return fmt.Errorf("error creating EC2 Tag (%s) for resource (%s): %w", key, resourceID, err)
	}

	// Handle EC2 eventual consistency on creation
	log.Printf("[DEBUG] Waiting for tag %s on resource %s to become available", key, resourceID)
	retryError := resource.Retry(5*time.Minute, func() *resource.RetryError {
		var tags *ec2.DescribeTagsOutput
		tags, err = conn.DescribeTags(&ec2.DescribeTagsInput{
			Filters: []*ec2.Filter{
				{
					Name:   aws.String("resource-id"),
					Values: []*string{aws.String(resourceID)},
				},
				{
					Name:   aws.String("key"),
					Values: []*string{aws.String(key)},
				},
			},
		})

		if err != nil {
			return resource.NonRetryableError(err)
		}

		// tag not found _yet_
		if len(tags.Tags) == 0 {
			return resource.RetryableError(&resource.NotFoundError{})
		}

		return nil
	})

	if retryError != nil {
		if isResourceNotFoundError(err) {
			return fmt.Errorf("error creating EC2 Tag (%s) on resource (%s): %w", key, resourceID, err)
		}
	}

	d.SetId(fmt.Sprintf("%s:%s", resourceID, key))
	return resourceAwsEc2TagRead(d, meta)
}

func resourceAwsEc2TagRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).ec2conn
	id, _, err := extractResourceIDAndKeyFromEc2TagID(d.Id())

	if err != nil {
		return err
	}

	key := d.Get("key").(string)
	var tags *ec2.DescribeTagsOutput

	tags, err = conn.DescribeTags(&ec2.DescribeTagsInput{
		Filters: []*ec2.Filter{
			{
				Name:   aws.String("resource-id"),
				Values: []*string{aws.String(id)},
			},
			{
				Name:   aws.String("key"),
				Values: []*string{aws.String(key)},
			},
		},
	})

	if err != nil {
		return fmt.Errorf("error reading EC2 Tag (%s) on resource (%s): %w", key, id, err)
	}

	if len(tags.Tags) == 0 {
		// The API call did not fail but the tag does not exists on resource
		// Did not find the tag, as per contract with TF report:https://www.terraform.io/docs/extend/writing-custom-providers.html
		log.Printf("[WARN]There are no tags on resource %s", id)
		d.SetId("")
		return nil
	}

	if len(tags.Tags) != 1 {
		return fmt.Errorf("Expected exactly 1 tag, got %d tags for key %s", len(tags.Tags), key)
	}

	tag := tags.Tags[0]
	d.Set("value", aws.StringValue(tag.Value))

	return nil
}

func resourceAwsEc2TagDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).ec2conn
	id, _, err := extractResourceIDAndKeyFromEc2TagID(d.Id())

	if err != nil {
		return err
	}

	_, err = conn.DeleteTags(&ec2.DeleteTagsInput{
		Resources: []*string{aws.String(id)},
		Tags: []*ec2.Tag{
			{
				Key:   aws.String(d.Get("key").(string)),
				Value: aws.String(d.Get("value").(string)),
			},
		},
	})

	if err != nil {
		return fmt.Errorf("error deleting EC2 Tag (%s) on resource (%s): %w", d.Get("key").(string), id, err)
	}

	return nil
}
