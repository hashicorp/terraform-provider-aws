// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package sns

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/aws/arn"
	"github.com/aws/aws-sdk-go-v2/service/sns"
	"github.com/aws/aws-sdk-go-v2/service/sns/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	tfslices "github.com/hashicorp/terraform-provider-aws/internal/slices"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKDataSource("aws_sns_topic")
func dataSourceTopic() *schema.Resource {
	return &schema.Resource{
		ReadWithoutTimeout: dataSourceTopicRead,

		Schema: map[string]*schema.Schema{
			names.AttrARN: {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrName: {
				Type:     schema.TypeString,
				Required: true,
			},
		},
	}
}

func dataSourceTopicRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).SNSClient(ctx)

	name := d.Get(names.AttrName).(string)
	topic, err := findTopicByName(ctx, conn, name)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading SNS Topic (%s): %s", name, err)
	}

	topicARN := aws.ToString(topic.TopicArn)
	d.SetId(topicARN)
	d.Set(names.AttrARN, topicARN)

	return diags
}

func findTopicByName(ctx context.Context, conn *sns.Client, name string) (*types.Topic, error) {
	input := &sns.ListTopicsInput{}

	return findTopic(ctx, conn, input, func(v types.Topic) bool {
		arn, err := arn.Parse(aws.ToString(v.TopicArn))
		return err == nil && arn.Resource == name
	})
}

func findTopic(ctx context.Context, conn *sns.Client, input *sns.ListTopicsInput, filter tfslices.Predicate[types.Topic]) (*types.Topic, error) {
	output, err := findTopics(ctx, conn, input, filter)

	if err != nil {
		return nil, err
	}

	return tfresource.AssertSingleValueResult(output)
}

func findTopics(ctx context.Context, conn *sns.Client, input *sns.ListTopicsInput, filter tfslices.Predicate[types.Topic]) ([]types.Topic, error) {
	var output []types.Topic

	pages := sns.NewListTopicsPaginator(conn, input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if err != nil {
			return nil, err
		}

		for _, v := range page.Topics {
			if filter(v) {
				output = append(output, v)
			}
		}
	}

	return output, nil
}
