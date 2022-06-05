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
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
)

func DataSourceHoursOfOperation() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceHoursOfOperationRead,
		Schema: map[string]*schema.Schema{
			"arn": {
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
						"start_time": {
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
					buf.WriteString(fmt.Sprintf("%+v", m["start_time"].([]interface{})))
					return create.StringHashcode(buf.String())
				},
			},
			"description": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"hours_of_operation_arn": {
				Type:       schema.TypeString,
				Computed:   true,
				Deprecated: "use 'arn' attribute instead",
			},
			"hours_of_operation_id": {
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				ExactlyOneOf: []string{"hours_of_operation_id", "name"},
			},
			"instance_id": {
				Type:     schema.TypeString,
				Required: true,
			},
			"name": {
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				ExactlyOneOf: []string{"name", "hours_of_operation_id"},
			},
			"tags": tftags.TagsSchemaComputed(),
			"time_zone": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func dataSourceHoursOfOperationRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).ConnectConn
	ignoreTagsConfig := meta.(*conns.AWSClient).IgnoreTagsConfig

	instanceID := d.Get("instance_id").(string)

	input := &connect.DescribeHoursOfOperationInput{
		InstanceId: aws.String(instanceID),
	}

	if v, ok := d.GetOk("hours_of_operation_id"); ok {
		input.HoursOfOperationId = aws.String(v.(string))
	} else if v, ok := d.GetOk("name"); ok {
		name := v.(string)
		hoursOfOperationSummary, err := dataSourceGetHoursOfOperationSummaryByName(ctx, conn, instanceID, name)

		if err != nil {
			return diag.FromErr(fmt.Errorf("error finding Connect Hours of Operation Summary by name (%s): %w", name, err))
		}

		if hoursOfOperationSummary == nil {
			return diag.FromErr(fmt.Errorf("error finding Connect Hours of Operation Summary by name (%s): not found", name))
		}

		input.HoursOfOperationId = hoursOfOperationSummary.Id
	}

	resp, err := conn.DescribeHoursOfOperation(input)

	if err != nil {
		return diag.FromErr(fmt.Errorf("error getting Connect Hours of Operation: %w", err))
	}

	if resp == nil || resp.HoursOfOperation == nil {
		return diag.FromErr(fmt.Errorf("error getting Connect Hours of Operation: empty response"))
	}

	hoursOfOperation := resp.HoursOfOperation

	d.Set("arn", hoursOfOperation.HoursOfOperationArn)
	d.Set("hours_of_operation_arn", hoursOfOperation.HoursOfOperationArn) // Deprecated
	d.Set("hours_of_operation_id", hoursOfOperation.HoursOfOperationId)
	d.Set("instance_id", instanceID)
	d.Set("description", hoursOfOperation.Description)
	d.Set("name", hoursOfOperation.Name)
	d.Set("time_zone", hoursOfOperation.TimeZone)

	if err := d.Set("config", flattenConfigs(hoursOfOperation.Config)); err != nil {
		return diag.FromErr(fmt.Errorf("error setting config: %s", err))
	}

	if err := d.Set("tags", KeyValueTags(hoursOfOperation.Tags).IgnoreAWS().IgnoreConfig(ignoreTagsConfig).Map()); err != nil {
		return diag.FromErr(fmt.Errorf("error setting tags: %s", err))
	}

	d.SetId(fmt.Sprintf("%s:%s", instanceID, aws.StringValue(hoursOfOperation.HoursOfOperationId)))

	return nil
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
