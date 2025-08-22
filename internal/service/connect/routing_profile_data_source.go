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
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	tfslices "github.com/hashicorp/terraform-provider-aws/internal/slices"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKDataSource("aws_connect_routing_profile", name="Routing Profile")
// @Tags
func dataSourceRoutingProfile() *schema.Resource {
	return &schema.Resource{
		ReadWithoutTimeout: dataSourceRoutingProfileRead,

		Schema: map[string]*schema.Schema{
			names.AttrARN: {
				Type:     schema.TypeString,
				Computed: true,
			},
			"default_outbound_queue_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrDescription: {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrInstanceID: {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringLenBetween(1, 100),
			},
			"media_concurrencies": {
				Type:     schema.TypeSet,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"channel": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"concurrency": {
							Type:     schema.TypeInt,
							Computed: true,
						},
					},
				},
			},
			names.AttrName: {
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				ExactlyOneOf: []string{names.AttrName, "routing_profile_id"},
			},
			"queue_configs": {
				Type:     schema.TypeSet,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"channel": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"delay": {
							Type:     schema.TypeInt,
							Computed: true,
						},
						names.AttrPriority: {
							Type:     schema.TypeInt,
							Computed: true,
						},
						"queue_arn": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"queue_id": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"queue_name": {
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
			},
			"routing_profile_id": {
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				ExactlyOneOf: []string{"routing_profile_id", names.AttrName},
			},
			names.AttrTags: tftags.TagsSchemaComputed(),
		},
	}
}

func dataSourceRoutingProfileRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ConnectClient(ctx)

	instanceID := d.Get(names.AttrInstanceID).(string)
	input := &connect.DescribeRoutingProfileInput{
		InstanceId: aws.String(instanceID),
	}

	if v, ok := d.GetOk("routing_profile_id"); ok {
		input.RoutingProfileId = aws.String(v.(string))
	} else if v, ok := d.GetOk(names.AttrName); ok {
		name := v.(string)
		routingProfileSummary, err := findRoutingProfileSummaryByTwoPartKey(ctx, conn, instanceID, name)

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "reading Connect Routing Profile (%s) summary: %s", name, err)
		}

		input.RoutingProfileId = routingProfileSummary.Id
	}

	routingProfile, err := findRoutingProfile(ctx, conn, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading Connect Routing Profile: %s", err)
	}

	routingProfileID := aws.ToString(routingProfile.RoutingProfileId)
	id := routingProfileCreateResourceID(instanceID, routingProfileID)
	d.SetId(id)
	d.Set(names.AttrARN, routingProfile.RoutingProfileArn)
	d.Set("default_outbound_queue_id", routingProfile.DefaultOutboundQueueId)
	d.Set(names.AttrDescription, routingProfile.Description)
	d.Set(names.AttrInstanceID, instanceID)
	if err := d.Set("media_concurrencies", flattenMediaConcurrencies(routingProfile.MediaConcurrencies)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting media_concurrencies: %s", err)
	}
	d.Set(names.AttrName, routingProfile.Name)
	d.Set("routing_profile_id", routingProfileID)

	queueConfigs, err := findRoutingConfigQueueConfigSummariesByTwoPartKey(ctx, conn, instanceID, routingProfileID)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading Connect Routing Profile (%s) Queue Config summaries: %s", d.Id(), err)
	}

	if err := d.Set("queue_configs", flattenRoutingConfigQueueConfigSummaries(queueConfigs)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting queue_configs: %s", err)
	}

	setTagsOut(ctx, routingProfile.Tags)

	return diags
}

func findRoutingProfileSummaryByTwoPartKey(ctx context.Context, conn *connect.Client, instanceID, name string) (*awstypes.RoutingProfileSummary, error) {
	const maxResults = 60
	input := &connect.ListRoutingProfilesInput{
		InstanceId: aws.String(instanceID),
		MaxResults: aws.Int32(maxResults),
	}

	return findRoutingProfileSummary(ctx, conn, input, func(v *awstypes.RoutingProfileSummary) bool {
		return aws.ToString(v.Name) == name
	})
}

func findRoutingProfileSummary(ctx context.Context, conn *connect.Client, input *connect.ListRoutingProfilesInput, filter tfslices.Predicate[*awstypes.RoutingProfileSummary]) (*awstypes.RoutingProfileSummary, error) {
	output, err := findRoutingProfileSummaries(ctx, conn, input, filter)

	if err != nil {
		return nil, err
	}

	return tfresource.AssertSingleValueResult(output)
}

func findRoutingProfileSummaries(ctx context.Context, conn *connect.Client, input *connect.ListRoutingProfilesInput, filter tfslices.Predicate[*awstypes.RoutingProfileSummary]) ([]awstypes.RoutingProfileSummary, error) {
	var output []awstypes.RoutingProfileSummary

	pages := connect.NewListRoutingProfilesPaginator(conn, input)
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

		for _, v := range page.RoutingProfileSummaryList {
			if filter(&v) {
				output = append(output, v)
			}
		}
	}

	return output, nil
}
