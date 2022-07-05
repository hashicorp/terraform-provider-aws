package sqs

import (
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/sqs"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

func DataSourceQueue() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceQueueRead,
		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				Required: true,
			},
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"url": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"tags": tftags.TagsSchemaComputed(),
		},
	}
}

func dataSourceQueueRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).SQSConn
	ignoreTagsConfig := meta.(*conns.AWSClient).IgnoreTagsConfig

	name := d.Get("name").(string)

	urlOutput, err := conn.GetQueueUrl(&sqs.GetQueueUrlInput{
		QueueName: aws.String(name),
	})
	if err != nil || urlOutput.QueueUrl == nil {
		return fmt.Errorf("Error getting queue URL: %w", err)
	}

	queueURL := aws.StringValue(urlOutput.QueueUrl)

	attributesOutput, err := conn.GetQueueAttributes(&sqs.GetQueueAttributesInput{
		QueueUrl:       aws.String(queueURL),
		AttributeNames: []*string{aws.String(sqs.QueueAttributeNameQueueArn)},
	})
	if err != nil {
		return fmt.Errorf("Error getting queue attributes: %w", err)
	}

	d.Set("arn", attributesOutput.Attributes[sqs.QueueAttributeNameQueueArn])
	d.Set("url", queueURL)
	d.SetId(queueURL)

	tags, err := ListTags(conn, queueURL)

	if verify.CheckISOErrorTagsUnsupported(conn.PartitionID, err) {
		// Some partitions may not support tagging, giving error
		log.Printf("[WARN] failed listing tags for SQS Queue (%s): %s", d.Id(), err)
		return nil
	}

	if err != nil {
		return fmt.Errorf("failed listing tags for SQS Queue (%s): %w", d.Id(), err)
	}

	if err := d.Set("tags", tags.IgnoreAWS().IgnoreConfig(ignoreTagsConfig).Map()); err != nil {
		return fmt.Errorf("error setting tags: %w", err)
	}

	return nil
}
