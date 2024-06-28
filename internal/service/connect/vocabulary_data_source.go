// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package connect

import (
	"context"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/connect"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKDataSource("aws_connect_vocabulary")
func DataSourceVocabulary() *schema.Resource {
	return &schema.Resource{
		ReadWithoutTimeout: dataSourceVocabularyRead,

		Schema: map[string]*schema.Schema{
			names.AttrARN: {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrContent: {
				Type:     schema.TypeString,
				Computed: true,
			},
			"failure_reason": {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrInstanceID: {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringLenBetween(1, 100),
			},
			names.AttrLanguageCode: {
				Type:     schema.TypeString,
				Computed: true,
			},
			"last_modified_time": {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrName: {
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				ExactlyOneOf: []string{names.AttrName, "vocabulary_id"},
			},
			names.AttrState: {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrTags: tftags.TagsSchemaComputed(),
			"vocabulary_id": {
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				ExactlyOneOf: []string{"vocabulary_id", names.AttrName},
			},
		},
	}
}

func dataSourceVocabularyRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	conn := meta.(*conns.AWSClient).ConnectConn(ctx)
	ignoreTagsConfig := meta.(*conns.AWSClient).IgnoreTagsConfig

	instanceID := d.Get(names.AttrInstanceID).(string)

	input := &connect.DescribeVocabularyInput{
		InstanceId: aws.String(instanceID),
	}

	if v, ok := d.GetOk("vocabulary_id"); ok {
		input.VocabularyId = aws.String(v.(string))
	} else if v, ok := d.GetOk(names.AttrName); ok {
		name := v.(string)
		vocabularySummary, err := dataSourceGetVocabularySummaryByName(ctx, conn, instanceID, name)

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "finding Connect Vocabulary Summary by name (%s): %s", name, err)
		}

		if vocabularySummary == nil {
			return sdkdiag.AppendErrorf(diags, "finding Connect Vocabulary Summary by name (%s): not found", name)
		}

		input.VocabularyId = vocabularySummary.Id
	}

	resp, err := conn.DescribeVocabularyWithContext(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "getting Connect Vocabulary: %s", err)
	}

	if resp == nil || resp.Vocabulary == nil {
		return sdkdiag.AppendErrorf(diags, "getting Connect Vocabulary: empty response")
	}

	vocabulary := resp.Vocabulary

	d.Set(names.AttrARN, vocabulary.Arn)
	d.Set(names.AttrContent, vocabulary.Content)
	d.Set("failure_reason", vocabulary.FailureReason)
	d.Set(names.AttrInstanceID, instanceID)
	d.Set(names.AttrLanguageCode, vocabulary.LanguageCode)
	d.Set("last_modified_time", vocabulary.LastModifiedTime.Format(time.RFC3339))
	d.Set(names.AttrName, vocabulary.Name)
	d.Set(names.AttrState, vocabulary.State)
	d.Set("vocabulary_id", vocabulary.Id)

	if err := d.Set(names.AttrTags, KeyValueTags(ctx, vocabulary.Tags).IgnoreAWS().IgnoreConfig(ignoreTagsConfig).Map()); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting tags: %s", err)
	}

	d.SetId(fmt.Sprintf("%s:%s", instanceID, aws.StringValue(vocabulary.Id)))

	return diags
}

func dataSourceGetVocabularySummaryByName(ctx context.Context, conn *connect.Connect, instanceID, name string) (*connect.VocabularySummary, error) {
	var result *connect.VocabularySummary

	input := &connect.SearchVocabulariesInput{
		InstanceId:     aws.String(instanceID),
		MaxResults:     aws.Int64(SearchVocabulariesMaxResults),
		NameStartsWith: aws.String(name),
	}

	err := conn.SearchVocabulariesPagesWithContext(ctx, input, func(page *connect.SearchVocabulariesOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, qs := range page.VocabularySummaryList {
			if qs == nil {
				continue
			}

			if aws.StringValue(qs.Name) == name {
				result = qs
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
