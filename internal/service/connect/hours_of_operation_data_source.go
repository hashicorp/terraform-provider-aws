// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package connect

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/connect"
	awstypes "github.com/aws/aws-sdk-go-v2/service/connect/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	tfslices "github.com/hashicorp/terraform-provider-aws/internal/slices"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKDataSource("aws_connect_hours_of_operation", name="Hours Of Operation")
// @Tags
func dataSourceHoursOfOperation() *schema.Resource {
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

func dataSourceHoursOfOperationRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ConnectClient(ctx)

	instanceID := d.Get(names.AttrInstanceID).(string)
	input := &connect.DescribeHoursOfOperationInput{
		InstanceId: aws.String(instanceID),
	}

	if v, ok := d.GetOk("hours_of_operation_id"); ok {
		input.HoursOfOperationId = aws.String(v.(string))
	} else if v, ok := d.GetOk(names.AttrName); ok {
		name := v.(string)
		hoursOfOperationSummary, err := findHoursOfOperationSummaryByTwoPartKey(ctx, conn, instanceID, name)

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "reading Connect Hours Of Operation (%s) summary: %s", name, err)
		}

		input.HoursOfOperationId = hoursOfOperationSummary.Id
	}

	hoursOfOperation, err := findHoursOfOperation(ctx, conn, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading Connect Hours Of Operation: %s", err)
	}

	hoursOfOperationID := aws.ToString(hoursOfOperation.HoursOfOperationId)
	id := hoursOfOperationCreateResourceID(instanceID, hoursOfOperationID)
	d.SetId(id)
	d.Set(names.AttrARN, hoursOfOperation.HoursOfOperationArn)
	if err := d.Set("config", flattenHoursOfOperationConfigs(hoursOfOperation.Config)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting config: %s", err)
	}
	d.Set(names.AttrDescription, hoursOfOperation.Description)
	d.Set("hours_of_operation_id", hoursOfOperationID)
	d.Set(names.AttrInstanceID, instanceID)
	d.Set(names.AttrName, hoursOfOperation.Name)
	d.Set("time_zone", hoursOfOperation.TimeZone)

	setTagsOut(ctx, hoursOfOperation.Tags)

	return diags
}

func findHoursOfOperationSummaryByTwoPartKey(ctx context.Context, conn *connect.Client, instanceID, name string) (*awstypes.HoursOfOperationSummary, error) {
	const maxResults = 60
	input := &connect.ListHoursOfOperationsInput{
		InstanceId: aws.String(instanceID),
		MaxResults: aws.Int32(maxResults),
	}

	return findHoursOfOperationSummary(ctx, conn, input, func(v *awstypes.HoursOfOperationSummary) bool {
		return aws.ToString(v.Name) == name
	})
}

func findHoursOfOperationSummary(ctx context.Context, conn *connect.Client, input *connect.ListHoursOfOperationsInput, filter tfslices.Predicate[*awstypes.HoursOfOperationSummary]) (*awstypes.HoursOfOperationSummary, error) {
	output, err := findHoursOfOperationSummaries(ctx, conn, input, filter)

	if err != nil {
		return nil, err
	}

	return tfresource.AssertSingleValueResult(output)
}

func findHoursOfOperationSummaries(ctx context.Context, conn *connect.Client, input *connect.ListHoursOfOperationsInput, filter tfslices.Predicate[*awstypes.HoursOfOperationSummary]) ([]awstypes.HoursOfOperationSummary, error) {
	var output []awstypes.HoursOfOperationSummary

	pages := connect.NewListHoursOfOperationsPaginator(conn, input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if errs.IsA[*awstypes.ResourceNotFoundException](err) {
			return nil, &retry.NotFoundError{
				LastError:   err,
				LastRequest: input,
			}
		}

		if err != nil {
			return nil, err
		}

		for _, v := range page.HoursOfOperationSummaryList {
			if filter(&v) {
				output = append(output, v)
			}
		}
	}

	return output, nil
}
