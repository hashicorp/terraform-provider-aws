package logs

import (
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/cloudwatchlogs"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
)

func DataSourceGroups() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceGroupsRead,

		Schema: map[string]*schema.Schema{
			"arns": {
				Type:     schema.TypeSet,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"log_group_name_prefix": {
				Type:     schema.TypeString,
				Required: true,
			},
			"log_group_names": {
				Type:     schema.TypeSet,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
		},
	}
}

func dataSourceGroupsRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).LogsConn

	input := &cloudwatchlogs.DescribeLogGroupsInput{
		LogGroupNamePrefix: aws.String(d.Get("log_group_name_prefix").(string)),
	}

	var results []*cloudwatchlogs.LogGroup

	err := conn.DescribeLogGroupsPages(input, func(page *cloudwatchlogs.DescribeLogGroupsOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		results = append(results, page.LogGroups...)

		return !lastPage
	})

	if err != nil {
		return fmt.Errorf("error reading CloudWatch Log Groups: %w", err)
	}

	d.SetId(meta.(*conns.AWSClient).Region)

	var arns, logGroupNames []string

	for _, r := range results {
		arns = append(arns, TrimLogGroupARNWildcardSuffix(aws.StringValue(r.Arn)))
		logGroupNames = append(logGroupNames, aws.StringValue(r.LogGroupName))
	}

	d.Set("arns", arns)
	d.Set("log_group_names", logGroupNames)

	return nil
}
