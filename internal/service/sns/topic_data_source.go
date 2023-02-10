package sns

import (
	"context"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/arn"
	"github.com/aws/aws-sdk-go/service/sns"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
)

func DataSourceTopic() *schema.Resource {
	return &schema.Resource{
		ReadWithoutTimeout: dataSourceTopicRead,

		Schema: map[string]*schema.Schema{
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"name": {
				Type:     schema.TypeString,
				Required: true,
			},
		},
	}
}

func dataSourceTopicRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).SNSConn()

	resourceArn := ""
	name := d.Get("name").(string)

	err := conn.ListTopicsPagesWithContext(ctx, &sns.ListTopicsInput{}, func(page *sns.ListTopicsOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, topic := range page.Topics {
			topicArn := aws.StringValue(topic.TopicArn)
			arn, err := arn.Parse(topicArn)

			if err != nil {
				log.Printf("[ERROR] %s", err)

				continue
			}

			if arn.Resource == name {
				resourceArn = topicArn

				break
			}
		}

		return !lastPage
	})

	if err != nil {
		return diag.Errorf("listing SNS Topics: %s", err)
	}

	if resourceArn == "" {
		return diag.Errorf("no matching SNS Topic found")
	}

	d.SetId(resourceArn)
	d.Set("arn", resourceArn)

	return nil
}
