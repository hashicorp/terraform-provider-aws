package kendra

import (
	"context"
	"fmt"
	"log"
	"regexp"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/aws/arn"
	"github.com/aws/aws-sdk-go-v2/service/kendra/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

func ResourceDataSource() *schema.Resource {
	return &schema.Resource{
		ReadContext: resourceDataSourceRead,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(30 * time.Minute),
			Update: schema.DefaultTimeout(30 * time.Minute),
			Delete: schema.DefaultTimeout(30 * time.Minute),
		},
		CustomizeDiff: verify.SetTagsDiff,
		Schema: map[string]*schema.Schema{
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"created_at": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"data_source_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"description": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringLenBetween(1, 1000),
			},
			"error_message": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"index_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
				ValidateFunc: validation.StringMatch(
					regexp.MustCompile(`[a-zA-Z0-9][a-zA-Z0-9-]{35}`),
					"Starts with an alphanumeric character. Subsequently, can contain alphanumeric characters and hyphens. Fixed length of 36.",
				),
			},
			"language_code": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
				ValidateFunc: validation.All(
					validation.StringLenBetween(2, 10),
					validation.StringMatch(
						regexp.MustCompile(`[a-zA-Z-]*`),
						"Must have alphanumeric characters or hyphens.",
					),
				),
			},
			"name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
				ValidateFunc: validation.All(
					validation.StringLenBetween(1, 1000),
					validation.StringMatch(
						regexp.MustCompile(`[a-zA-Z0-9][a-zA-Z0-9_-]*`),
						"Starts with an alphanumeric character. Subsequently, the name must consist of alphanumerics, hyphens or underscores.",
					),
				),
			},
			"role_arn": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: verify.ValidARN,
			},
			"schedule": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"status": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"type": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringInSlice(dataSourceTypeValues(types.DataSourceType("").Values()...), false),
			},
			"updated_at": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"tags":     tftags.TagsSchema(),
			"tags_all": tftags.TagsSchemaComputed(),
		},
	}
}

func resourceDataSourceRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).KendraConn
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	ignoreTagsConfig := meta.(*conns.AWSClient).IgnoreTagsConfig

	id, indexId, err := DataSourceParseResourceID(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	resp, err := FindDataSourceByID(ctx, conn, id, indexId)

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] Kendra Data Source (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return diag.Errorf("getting Kendra Data Source (%s): %s", d.Id(), err)
	}

	arn := arn.ARN{
		Partition: meta.(*conns.AWSClient).Partition,
		Service:   "kendra",
		Region:    meta.(*conns.AWSClient).Region,
		AccountID: meta.(*conns.AWSClient).AccountID,
		Resource:  fmt.Sprintf("index/%s/data-source/%s", indexId, id),
	}.String()

	d.Set("arn", arn)
	d.Set("created_at", aws.ToTime(resp.CreatedAt).Format(time.RFC3339))
	d.Set("data_source_id", resp.Id)
	d.Set("description", resp.Description)
	d.Set("error_message", resp.ErrorMessage)
	d.Set("index_id", resp.IndexId)
	d.Set("language_code", resp.LanguageCode)
	d.Set("name", resp.Name)
	d.Set("role_arn", resp.RoleArn)
	d.Set("schedule", resp.Schedule)
	d.Set("status", resp.Status)
	d.Set("type", resp.Type)
	d.Set("updated_at", aws.ToTime(resp.UpdatedAt).Format(time.RFC3339))

	tags, err := ListTags(ctx, conn, arn)
	if err != nil {
		return diag.Errorf("listing tags for resource (%s): %s", arn, err)
	}
	tags = tags.IgnoreAWS().IgnoreConfig(ignoreTagsConfig)

	//lintignore:AWSR002
	if err := d.Set("tags", tags.RemoveDefaultConfig(defaultTagsConfig).Map()); err != nil {
		return diag.Errorf("setting tags: %s", err)
	}

	if err := d.Set("tags_all", tags.Map()); err != nil {
		return diag.Errorf("setting tags_all: %s", err)
	}

	return nil
}

// Helpers added. Could be generated or somehow use go 1.18 generics?
func dataSourceTypeValues(input ...types.DataSourceType) []string {
	var output []string

	for _, v := range input {
		output = append(output, string(v))
	}

	return output
}

func dataSourceStatusValues(input ...types.DataSourceStatus) []string {
	var output []string

	for _, v := range input {
		output = append(output, string(v))
	}

	return output
}
