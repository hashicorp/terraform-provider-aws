package sqs

import (
	"context"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/sqs"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

func DataSourceQueue() *schema.Resource {
	return &schema.Resource{
		ReadWithoutTimeout: dataSourceQueueRead,

		Schema: map[string]*schema.Schema{
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"name": {
				Type:     schema.TypeString,
				Required: true,
			},
			"tags": tftags.TagsSchemaComputed(),
			"url": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func dataSourceQueueRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).SQSConn()
	ignoreTagsConfig := meta.(*conns.AWSClient).IgnoreTagsConfig

	name := d.Get("name").(string)

	urlOutput, err := conn.GetQueueUrlWithContext(ctx, &sqs.GetQueueUrlInput{
		QueueName: aws.String(name),
	})

	if err != nil || urlOutput.QueueUrl == nil {
		return diag.Errorf("reading SQS Queue (%s) URL: %s", name, err)
	}

	queueURL := aws.StringValue(urlOutput.QueueUrl)

	attributesOutput, err := conn.GetQueueAttributesWithContext(ctx, &sqs.GetQueueAttributesInput{
		QueueUrl:       aws.String(queueURL),
		AttributeNames: []*string{aws.String(sqs.QueueAttributeNameQueueArn)},
	})
	if err != nil {
		return diag.Errorf("reading SQS Queue (%s) attributes: %s", queueURL, err)
	}

	d.Set("arn", attributesOutput.Attributes[sqs.QueueAttributeNameQueueArn])
	d.Set("url", queueURL)
	d.SetId(queueURL)

	tags, err := ListTags(ctx, conn, queueURL)

	if verify.ErrorISOUnsupported(conn.PartitionID, err) {
		// Some partitions may not support tagging, giving error
		log.Printf("[WARN] failed listing tags for SQS Queue (%s): %s", d.Id(), err)
		return nil
	}

	if err != nil {
		return diag.Errorf("listing tags for SQS Queue (%s): %s", d.Id(), err)
	}

	if err := d.Set("tags", tags.IgnoreAWS().IgnoreConfig(ignoreTagsConfig).Map()); err != nil {
		return diag.Errorf("setting tags: %s", err)
	}

	return nil
}
