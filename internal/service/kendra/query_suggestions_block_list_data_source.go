package kendra

import (
	"context"
	"fmt"
	"regexp"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/aws/arn"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
)

func DataSourceQuerySuggestionsBlockList() *schema.Resource {
	return &schema.Resource{
		ReadWithoutTimeout: dataSourceQuerySuggestionsBlockListRead,
		Schema: map[string]*schema.Schema{
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"created_at": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"description": {
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
					regexp.MustCompile(`[a-zA-Z0-9][a-zA-Z0-9-]{35}`),
					"Starts with an alphanumeric character. Subsequently, can contain alphanumeric characters and hyphens. Fixed length of 36.",
				),
			},
			"item_count": {
				Type:     schema.TypeInt,
				Computed: true,
			},
			"name": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"query_suggestions_block_list_id": {
				Type:     schema.TypeString,
				Required: true,
				ValidateFunc: validation.StringMatch(
					regexp.MustCompile(`[a-zA-Z0-9][a-zA-Z0-9-]{35}`),
					"Starts with an alphanumeric character. Subsequently, can contain alphanumeric characters and hyphens. Fixed length of 36.",
				),
			},
			"role_arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"source_s3_path": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"bucket": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"key": {
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
			},
			"status": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"tags": tftags.TagsSchemaComputed(),
			"updated_at": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func dataSourceQuerySuggestionsBlockListRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).KendraClient()
	ignoreTagsConfig := meta.(*conns.AWSClient).IgnoreTagsConfig

	querySuggestionsBlockListID := d.Get("query_suggestions_block_list_id").(string)
	indexID := d.Get("index_id").(string)

	resp, err := FindQuerySuggestionsBlockListByID(ctx, conn, querySuggestionsBlockListID, indexID)

	if err != nil {
		return diag.Errorf("reading Kendra QuerySuggestionsBlockList (%s): %s", querySuggestionsBlockListID, err)
	}

	arn := arn.ARN{
		Partition: meta.(*conns.AWSClient).Partition,
		Service:   "kendra",
		Region:    meta.(*conns.AWSClient).Region,
		AccountID: meta.(*conns.AWSClient).AccountID,
		Resource:  fmt.Sprintf("index/%s/query-suggestions-block-list/%s", indexID, querySuggestionsBlockListID),
	}.String()
	d.Set("arn", arn)
	d.Set("created_at", aws.ToTime(resp.CreatedAt).Format(time.RFC3339))
	d.Set("description", resp.Description)
	d.Set("error_message", resp.ErrorMessage)
	d.Set("file_size_bytes", resp.FileSizeBytes)
	d.Set("index_id", resp.IndexId)
	d.Set("item_count", resp.ItemCount)
	d.Set("name", resp.Name)
	d.Set("query_suggestions_block_list_id", resp.Id)
	d.Set("role_arn", resp.RoleArn)
	d.Set("status", resp.Status)
	d.Set("updated_at", aws.ToTime(resp.UpdatedAt).Format(time.RFC3339))

	if err := d.Set("source_s3_path", flattenSourceS3Path(resp.SourceS3Path)); err != nil {
		return diag.Errorf("setting source_s3_path: %s", err)
	}

	tags, err := ListTags(ctx, conn, arn)

	if err != nil {
		return diag.Errorf("listing tags for Kendra QuerySuggestionsBlockList (%s): %s", arn, err)
	}

	tags = tags.IgnoreAWS().IgnoreConfig(ignoreTagsConfig)

	if err := d.Set("tags", tags.IgnoreAWS().IgnoreConfig(ignoreTagsConfig).Map()); err != nil {
		return diag.Errorf("setting tags: %s", err)
	}

	d.SetId(fmt.Sprintf("%s/%s", querySuggestionsBlockListID, indexID))

	return nil
}
