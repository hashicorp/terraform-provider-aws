// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package connect

import (
	"bytes"
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/connect"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKDataSource("aws_connect_hours_of_operation")
func DataSourceHoursOfOperation() *schema.Resource {
	return &schema.Resource{
		ReadWithoutTimeout: dataSourceHoursOfOperationRead,
		Schema: map[string]*schema.Schema{
			names.AttrARN: {
				Type:     schema.TypeString,
				Computed: true,
			},
			"config": {
				Type:     schema.TypeSet,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"day": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"end_time": {
							Type:     schema.TypeList,
							Computed: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"hours": {
										Type:     schema.TypeInt,
										Computed: true,
									},
									"minutes": {
										Type:     schema.TypeInt,
										Computed: true,
									},
								},
							},
						},
						names.AttrStartTime: {
							Type:     schema.TypeList,
							Computed: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"hours": {
										Type:     schema.TypeInt,
										Computed: true,
									},
									"minutes": {
										Type:     schema.TypeInt,
										Computed: true,
									},
								},
							},
						},
					},
				},
				Set: func(v interface{}) int {
					var buf bytes.Buffer
					m := v.(map[string]interface{})
					buf.WriteString(m["day"].(string))
					buf.WriteString(fmt.Sprintf("%+v", m["end_time"].([]interface{})))
					buf.WriteString(fmt.Sprintf("%+v", m[names.AttrStartTime].([]interface{})))
					return create.StringHashcode(buf.String())
				},
			},
			names.AttrDescription: {
				Type:     schema.TypeString,
				Computed: true,
			},
			"hours_of_operation_id": {
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				ExactlyOneOf: []string{"hours_of_operation_id", names.AttrName},
			},
			names.AttrInstanceID: {
				Type:     schema.TypeString,
				Required: true,
			},
			names.AttrName: {
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				ExactlyOneOf: []string{names.AttrName, "hours_of_operation_id"},
			},
			names.AttrTags: tftags.TagsSchemaComputed(),
			"time_zone": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func dataSourceHoursOfOperationRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	conn := meta.(*conns.AWSClient).ConnectConn(ctx)
	ignoreTagsConfig := meta.(*conns.AWSClient).IgnoreTagsConfig

	instanceID := d.Get(names.AttrInstanceID).(string)

	input := &connect.DescribeHoursOfOperationInput{
		InstanceId: aws.String(instanceID),
	}

	if v, ok := d.GetOk("hours_of_operation_id"); ok {
		input.HoursOfOperationId = aws.String(v.(string))
	} else if v, ok := d.GetOk(names.AttrName); ok {
		name := v.(string)
		hoursOfOperationSummary, err := dataSourceGetHoursOfOperationSummaryByName(ctx, conn, instanceID, name)

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "finding Connect Hours of Operation Summary by name (%s): %s", name, err)
		}

		if hoursOfOperationSummary == nil {
			return sdkdiag.AppendErrorf(diags, "finding Connect Hours of Operation Summary by name (%s): not found", name)
		}

		input.HoursOfOperationId = hoursOfOperationSummary.Id
	}

	resp, err := conn.DescribeHoursOfOperationWithContext(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "getting Connect Hours of Operation: %s", err)
	}

	if resp == nil || resp.HoursOfOperation == nil {
		return sdkdiag.AppendErrorf(diags, "getting Connect Hours of Operation: empty response")
	}

	hoursOfOperation := resp.HoursOfOperation

	d.Set(names.AttrARN, hoursOfOperation.HoursOfOperationArn)
	d.Set("hours_of_operation_id", hoursOfOperation.HoursOfOperationId)
	d.Set(names.AttrInstanceID, instanceID)
	d.Set(names.AttrDescription, hoursOfOperation.Description)
	d.Set(names.AttrName, hoursOfOperation.Name)
	d.Set("time_zone", hoursOfOperation.TimeZone)

	if err := d.Set("config", flattenConfigs(hoursOfOperation.Config)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting config: %s", err)
	}

	if err := d.Set(names.AttrTags, KeyValueTags(ctx, hoursOfOperation.Tags).IgnoreAWS().IgnoreConfig(ignoreTagsConfig).Map()); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting tags: %s", err)
	}

	d.SetId(fmt.Sprintf("%s:%s", instanceID, aws.StringValue(hoursOfOperation.HoursOfOperationId)))

	return diags
}

func dataSourceGetHoursOfOperationSummaryByName(ctx context.Context, conn *connect.Connect, instanceID, name string) (*connect.HoursOfOperationSummary, error) {
	var result *connect.HoursOfOperationSummary

	input := &connect.ListHoursOfOperationsInput{
		InstanceId: aws.String(instanceID),
		MaxResults: aws.Int64(ListHoursOfOperationsMaxResults),
	}

	err := conn.ListHoursOfOperationsPagesWithContext(ctx, input, func(page *connect.ListHoursOfOperationsOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, cf := range page.HoursOfOperationSummaryList {
			if cf == nil {
				continue
			}

			if aws.StringValue(cf.Name) == name {
				result = cf
				return false
			}
		}

		return !lastPage
	})

	if err != nil {
		return nil, err
	}

	return result, nil
}
