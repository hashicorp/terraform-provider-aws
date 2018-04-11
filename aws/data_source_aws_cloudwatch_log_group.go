package aws

import (
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/cloudwatchlogs"
	"github.com/hashicorp/terraform/helper/schema"
)

func dataSourceAwsCloudwatchLogGroup() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceAwsCloudwatchLogGroupRead,

		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"creation_time": {
				Type:     schema.TypeInt,
				Computed: true,
			},
		},
	}
}

func dataSourceAwsCloudwatchLogGroupRead(d *schema.ResourceData, meta interface{}) error {
	name := aws.String(d.Get("name").(string))
	conn := meta.(*AWSClient).cloudwatchlogsconn

	input := &cloudwatchlogs.DescribeLogGroupsInput{
		LogGroupNamePrefix: name,
	}

	resp, err := conn.DescribeLogGroups(input)
	if err != nil {
		return err
	}

	var logGroup *cloudwatchlogs.LogGroup

	for _, lg := range resp.LogGroups {
		if *lg.LogGroupName == *name {
			logGroup = lg
			break
		}
	}

	if logGroup == nil {
		return fmt.Errorf("No log group named %s found\n", *name)
	}

	d.SetId(*logGroup.LogGroupName)
	d.Set("arn", logGroup.Arn)
	d.Set("creation_time", logGroup.CreationTime)

	return nil
}
