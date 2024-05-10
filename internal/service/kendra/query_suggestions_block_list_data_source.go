// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package kendra

import (
	"context"
	"fmt"
	"time"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/aws/arn"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKDataSource("aws_kendra_query_suggestions_block_list")
func DataSourceQuerySuggestionsBlockList() *schema.Resource {
	return &schema.Resource{
		ReadWithoutTimeout: dataSourceQuerySuggestionsBlockListRead,
		Schema: map[string]*schema.Schema{
			names.AttrARN: {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrCreatedAt: {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrDescription: {
				Type:     schema.TypeString,
				Computed: true,
			},
			"error_message": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"file_size_bytes": {
				Type:     schema.TypeInt,
				Computed: true,
			},
			"index_id": {
				Type:     schema.TypeString,
				Required: true,
				ValidateFunc: validation.StringMatch(
					regexache.MustCompile(`[0-9A-Za-z][0-9A-Za-z-]{35}`),
					"Starts with an alphanumeric character. Subsequently, can contain alphanumeric characters and hyphens. Fixed length of 36.",
				),
			},
			"item_count": {
				Type:     schema.TypeInt,
				Computed: true,
			},
			names.AttrName: {
				Type:     schema.TypeString,
				Computed: true,
			},
			"query_suggestions_block_list_id": {
				Type:     schema.TypeString,
				Required: true,
				ValidateFunc: validation.StringMatch(
					regexache.MustCompile(`[0-9A-Za-z][0-9A-Za-z-]{35}`),
					"Starts with an alphanumeric character. Subsequently, can contain alphanumeric characters and hyphens. Fixed length of 36.",
				),
			},
			names.AttrRoleARN: {
				Type:     schema.TypeString,
				Computed: true,
			},
			"source_s3_path": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						names.AttrBucket: {
							Type:     schema.TypeString,
							Computed: true,
						},
						names.AttrKey: {
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
			},
			names.AttrStatus: {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrTags: tftags.TagsSchemaComputed(),
			"updated_at": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func dataSourceQuerySuggestionsBlockListRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	conn := meta.(*conns.AWSClient).KendraClient(ctx)
	ignoreTagsConfig := meta.(*conns.AWSClient).IgnoreTagsConfig

	querySuggestionsBlockListID := d.Get("query_suggestions_block_list_id").(string)
	indexID := d.Get("index_id").(string)

	resp, err := FindQuerySuggestionsBlockListByID(ctx, conn, querySuggestionsBlockListID, indexID)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading Kendra QuerySuggestionsBlockList (%s): %s", querySuggestionsBlockListID, err)
	}

	arn := arn.ARN{
		Partition: meta.(*conns.AWSClient).Partition,
		Service:   "kendra",
		Region:    meta.(*conns.AWSClient).Region,
		AccountID: meta.(*conns.AWSClient).AccountID,
		Resource:  fmt.Sprintf("index/%s/query-suggestions-block-list/%s", indexID, querySuggestionsBlockListID),
	}.String()
	d.Set(names.AttrARN, arn)
	d.Set(names.AttrCreatedAt, aws.ToTime(resp.CreatedAt).Format(time.RFC3339))
	d.Set(names.AttrDescription, resp.Description)
	d.Set("error_message", resp.ErrorMessage)
	d.Set("file_size_bytes", resp.FileSizeBytes)
	d.Set("index_id", resp.IndexId)
	d.Set("item_count", resp.ItemCount)
	d.Set(names.AttrName, resp.Name)
	d.Set("query_suggestions_block_list_id", resp.Id)
	d.Set(names.AttrRoleARN, resp.RoleArn)
	d.Set(names.AttrStatus, resp.Status)
	d.Set("updated_at", aws.ToTime(resp.UpdatedAt).Format(time.RFC3339))

	if err := d.Set("source_s3_path", flattenSourceS3Path(resp.SourceS3Path)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting source_s3_path: %s", err)
	}

	tags, err := listTags(ctx, conn, arn)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "listing tags for Kendra QuerySuggestionsBlockList (%s): %s", arn, err)
	}

	tags = tags.IgnoreAWS().IgnoreConfig(ignoreTagsConfig)

	if err := d.Set(names.AttrTags, tags.IgnoreAWS().IgnoreConfig(ignoreTagsConfig).Map()); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting tags: %s", err)
	}

	d.SetId(fmt.Sprintf("%s/%s", querySuggestionsBlockListID, indexID))

	return diags
}
