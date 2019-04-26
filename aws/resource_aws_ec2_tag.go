package aws

import (
	"fmt"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/hashicorp/terraform/helper/schema"
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

func extractResourceIdFromEc2TagId(d *schema.ResourceData) (string, error) {
	i := d.Id()
	parts := strings.Split(i, ":")

	if len(parts) != 2 {
		return "", fmt.Errorf("Invalid resource ID; cannot look up resource: %s", i)
	}

	return parts[0], nil
}

func resourceAwsEc2TagCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).ec2conn

	resourceID := d.Get("resource_id").(string)

	_, err := conn.CreateTags(&ec2.CreateTagsInput{
		Resources: []*string{aws.String(resourceID)},
		Tags: []*ec2.Tag{
			{
				Key:   aws.String(d.Get("key").(string)),
				Value: aws.String(d.Get("value").(string)),
			},
		},
	})

	if err != nil {
		return err
	}

	d.SetId(fmt.Sprintf("%s:%s", resourceID, d.Get("key").(string)))
	return resourceAwsEc2TagRead(d, meta)
}

func resourceAwsEc2TagRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).ec2conn
	id, err := extractResourceIdFromEc2TagId(d)

	if err != nil {
		return err
	}

	key := d.Get("key").(string)

	tags, err := conn.DescribeTags(&ec2.DescribeTagsInput{
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
		return err
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
	id, err := extractResourceIdFromEc2TagId(d)

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
		return err
	}

	return nil
}
