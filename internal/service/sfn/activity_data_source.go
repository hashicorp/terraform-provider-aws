// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package sfn

import (
	"context"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/sfn"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
)

// @SDKDataSource("aws_sfn_activity")
func DataSourceActivity() *schema.Resource {
	return &schema.Resource{
		ReadWithoutTimeout: dataSourceActivityRead,

		Schema: map[string]*schema.Schema{
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
				Optional: true,
				ExactlyOneOf: []string{
					"arn",
					"name",
				},
			},
			"creation_date": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"name": {
				Type:     schema.TypeString,
				Computed: true,
				Optional: true,
				ExactlyOneOf: []string{
					"arn",
					"name",
				},
			},
		},
	}
}

func dataSourceActivityRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).SFNConn(ctx)

	if v, ok := d.GetOk("name"); ok {
		name := v.(string)
		var activities []*sfn.ActivityListItem

		err := conn.ListActivitiesPagesWithContext(ctx, &sfn.ListActivitiesInput{}, func(page *sfn.ListActivitiesOutput, lastPage bool) bool {
			if page == nil {
				return !lastPage
			}

			for _, v := range page.Activities {
				if name == aws.StringValue(v.Name) {
					activities = append(activities, v)
				}
			}

			return !lastPage
		})

		if err != nil {
			return diag.Errorf("listing Step Functions Activities: %s", err)
		}

		if n := len(activities); n == 0 {
			return diag.Errorf("no Step Functions Activities matched")
		} else if n > 1 {
			return diag.Errorf("%d Step Functions Activities matched; use additional constraints to reduce matches to a single Activity", n)
		}

		activity := activities[0]

		arn := aws.StringValue(activity.ActivityArn)
		d.SetId(arn)
		d.Set("arn", arn)
		d.Set("creation_date", activity.CreationDate.Format(time.RFC3339))
		d.Set("name", activity.Name)
	} else if v, ok := d.GetOk("arn"); ok {
		arn := v.(string)
		activity, err := FindActivityByARN(ctx, conn, arn)

		if err != nil {
			return diag.Errorf("reading Step Functions Activity (%s): %s", arn, err)
		}

		arn = aws.StringValue(activity.ActivityArn)
		d.SetId(arn)
		d.Set("arn", arn)
		d.Set("creation_date", activity.CreationDate.Format(time.RFC3339))
		d.Set("name", activity.Name)
	}

	return nil
}
