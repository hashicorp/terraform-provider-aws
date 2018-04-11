package aws

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/sqs"
	"github.com/hashicorp/errwrap"
	"github.com/hashicorp/terraform/helper/schema"
)

func dataSourceAwsSqsQueue() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceAwsSqsQueueRead,
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
		},
	}
}

func dataSourceAwsSqsQueueRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).sqsconn
	target := d.Get("name").(string)

	urlOutput, err := conn.GetQueueUrl(&sqs.GetQueueUrlInput{
		QueueName: aws.String(target),
	})
	if err != nil {
		return errwrap.Wrapf("Error getting queue URL: {{err}}", err)
	}

	queueURL := *urlOutput.QueueUrl

	attributesOutput, err := conn.GetQueueAttributes(&sqs.GetQueueAttributesInput{
		QueueUrl:       &queueURL,
		AttributeNames: []*string{aws.String(sqs.QueueAttributeNameQueueArn)},
	})
	if err != nil {
		return errwrap.Wrapf("Error getting queue attributes: {{err}}", err)
	}

	d.Set("arn", *attributesOutput.Attributes[sqs.QueueAttributeNameQueueArn])
	d.Set("url", queueURL)
	d.SetId(queueURL)

	return nil
}
