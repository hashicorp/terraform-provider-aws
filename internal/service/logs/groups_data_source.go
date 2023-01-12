package logs

import (
	"context"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/cloudwatchlogs"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
)

func init() {
	_sp.registerSDKDataSourceFactory("aws_cloudwatch_log_groups", dataSourceGroups)
}

func dataSourceGroups() *schema.Resource {
	return &schema.Resource{
		ReadWithoutTimeout: dataSourceGroupsRead,

		Schema: map[string]*schema.Schema{
			"arns": {
				Type:     schema.TypeSet,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"log_group_name_prefix": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"log_group_names": {
				Type:     schema.TypeSet,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
		},
	}
}

func dataSourceGroupsRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).LogsConn()

	input := &cloudwatchlogs.DescribeLogGroupsInput{}

	if v, ok := d.GetOk("log_group_name_prefix"); ok {
		input.LogGroupNamePrefix = aws.String(v.(string))
	}

	var output []*cloudwatchlogs.LogGroup

	err := conn.DescribeLogGroupsPagesWithContext(ctx, input, func(page *cloudwatchlogs.DescribeLogGroupsOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		output = append(output, page.LogGroups...)

		return !lastPage
	})

	if err != nil {
		return diag.Errorf("reading CloudWatch Log Groups: %s", err)
	}

	d.SetId(meta.(*conns.AWSClient).Region)

	var arns, logGroupNames []string

	for _, r := range output {
		arns = append(arns, TrimLogGroupARNWildcardSuffix(aws.StringValue(r.Arn)))
		logGroupNames = append(logGroupNames, aws.StringValue(r.LogGroupName))
	}

	d.Set("arns", arns)
	d.Set("log_group_names", logGroupNames)

	return nil
}
