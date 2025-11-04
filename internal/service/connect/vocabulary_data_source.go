// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package connect

import (
	"context"
	"time"

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

// @SDKDataSource("aws_connect_vocabulary", name="Vocabulary")
// @Tags
func dataSourceVocabulary() *schema.Resource {
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

func dataSourceVocabularyRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ConnectClient(ctx)

	instanceID := d.Get(names.AttrInstanceID).(string)

	input := &connect.DescribeVocabularyInput{
		InstanceId: aws.String(instanceID),
	}

	if v, ok := d.GetOk("vocabulary_id"); ok {
		input.VocabularyId = aws.String(v.(string))
	} else if v, ok := d.GetOk(names.AttrName); ok {
		name := v.(string)
		vocabularySummary, err := findVocabularySummaryByTwoPartKey(ctx, conn, instanceID, name)

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "reading Connect Vocabulary (%s) summary: %s", name, err)
		}

		input.VocabularyId = vocabularySummary.Id
	}

	vocabulary, err := findVocabulary(ctx, conn, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading Connect Vocabulary: %s", err)
	}

	vocabularyID := aws.ToString(vocabulary.Id)
	id := vocabularyCreateResourceID(instanceID, vocabularyID)
	d.SetId(id)
	d.Set(names.AttrARN, vocabulary.Arn)
	d.Set(names.AttrContent, vocabulary.Content)
	d.Set("failure_reason", vocabulary.FailureReason)
	d.Set(names.AttrInstanceID, instanceID)
	d.Set(names.AttrLanguageCode, vocabulary.LanguageCode)
	d.Set("last_modified_time", vocabulary.LastModifiedTime.Format(time.RFC3339))
	d.Set(names.AttrName, vocabulary.Name)
	d.Set(names.AttrState, vocabulary.State)
	d.Set("vocabulary_id", vocabularyID)

	setTagsOut(ctx, vocabulary.Tags)

	return diags
}

func findVocabularySummaryByTwoPartKey(ctx context.Context, conn *connect.Client, instanceID, name string) (*awstypes.VocabularySummary, error) {
	const maxResults = 60
	input := &connect.SearchVocabulariesInput{
		InstanceId:     aws.String(instanceID),
		MaxResults:     aws.Int32(maxResults),
		NameStartsWith: aws.String(name),
	}

	return findVocabularySummary(ctx, conn, input, func(v *awstypes.VocabularySummary) bool {
		return aws.ToString(v.Name) == name
	})
}

func findVocabularySummary(ctx context.Context, conn *connect.Client, input *connect.SearchVocabulariesInput, filter tfslices.Predicate[*awstypes.VocabularySummary]) (*awstypes.VocabularySummary, error) {
	output, err := findVocabularySummaries(ctx, conn, input, filter)

	if err != nil {
		return nil, err
	}

	return tfresource.AssertSingleValueResult(output)
}

func findVocabularySummaries(ctx context.Context, conn *connect.Client, input *connect.SearchVocabulariesInput, filter tfslices.Predicate[*awstypes.VocabularySummary]) ([]awstypes.VocabularySummary, error) {
	var output []awstypes.VocabularySummary

	pages := connect.NewSearchVocabulariesPaginator(conn, input)
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

		for _, v := range page.VocabularySummaryList {
			if filter(&v) {
				output = append(output, v)
			}
		}
	}

	return output, nil
}
