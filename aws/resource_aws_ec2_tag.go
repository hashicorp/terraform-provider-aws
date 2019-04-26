package aws

import (
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/hashicorp/terraform/helper/resource"
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

func extractResourceIdFromEc2TagId(d *schema.ResourceData) string {
	i := d.Id()
	parts := strings.Split(i, "-")

	if len(parts) != 2 {
		return fmt.Errorf("Invalid resource ID; cannot look up subnet: %s", i)
	}

	return parts[0]
}

func resourceAwsEc2TagCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).ec2conn
	d.SetId(fmt.Sprintf("%s-%s", d.Get("subnet_id"), d.Get("key")))
	return resourceAwsEc2TagRead(d, meta)
}

func resourceAwsEc2TagRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).ec2conn
	id := extractResourceIdFromEc2TagId(d)

	tags, err := conn.DescribeTags(&ec2.DescribeTagsInput{
		Filters: []*ec2.Filter{
			ec2.Filter{
				Name:   "resource-id",
				Values: []*string{aws.String(id)},
			},
			ec2.Filter{
				Name:   "key",
				Values: []*string{d.Get("key").(string)},
			},
		},
	})

	if err != nil {
		return err
	}

	if len(tags) != 1 {
		return fmt.Errorf("Expected exactly 1 tag, got %d tags", len(tags))
	}

	tag := tags[0]
	d.Set("value", aws.StringValue(tag.Value))

	return nil
}

func resourceAwsEc2TagDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).ec2conn
	subnetId := extractResourceIdFromEc2TagId(d)

	_, err := conn.DeleteTags(&ec2.DeleteTagsInput{
		Resources: []*string{aws.String(d.Id())},
		Tags: []*Tags{
			Tag{
				Key:   aws.String(d.Get("tag").(string)),
				Value: aws.String(d.Get("value").(string)),
			},
		},
	})

	return nil
}
