// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package ssm

import (
	"context"
	"log"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ssm"
	awstypes "github.com/aws/aws-sdk-go-v2/service/ssm/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	tfslices "github.com/hashicorp/terraform-provider-aws/internal/slices"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_ssm_resource_data_sync", name="Resource Data Sync")
func resourceResourceDataSync() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceResourceDataSyncCreate,
		ReadWithoutTimeout:   resourceResourceDataSyncRead,
		DeleteWithoutTimeout: resourceResourceDataSyncDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			names.AttrName: {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"s3_destination": {
				Type:     schema.TypeList,
				Required: true,
				ForceNew: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						names.AttrBucketName: {
							Type:     schema.TypeString,
							Required: true,
							ForceNew: true,
						},
						names.AttrKMSKeyARN: {
							Type:         schema.TypeString,
							Optional:     true,
							ForceNew:     true,
							ValidateFunc: verify.ValidARN,
						},
						names.AttrPrefix: {
							Type:     schema.TypeString,
							Optional: true,
							ForceNew: true,
						},
						names.AttrRegion: {
							Type:         schema.TypeString,
							Required:     true,
							ForceNew:     true,
							ValidateFunc: verify.ValidRegionName,
						},
						"sync_format": {
							Type:             schema.TypeString,
							Optional:         true,
							ForceNew:         true,
							Default:          awstypes.ResourceDataSyncS3FormatJsonSerde,
							ValidateDiagFunc: enum.Validate[awstypes.ResourceDataSyncS3Format](),
						},
					},
				},
			},
		},
	}
}

func resourceResourceDataSyncCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).SSMClient(ctx)

	name := d.Get(names.AttrName).(string)
	input := &ssm.CreateResourceDataSyncInput{
		S3Destination: expandResourceDataSyncS3Destination(d),
		SyncName:      aws.String(name),
	}

	const (
		timeout = 1 * time.Minute
	)
	_, err := tfresource.RetryWhenIsAErrorMessageContains[*awstypes.ResourceDataSyncInvalidConfigurationException](ctx, timeout, func() (interface{}, error) {
		return conn.CreateResourceDataSync(ctx, input)
	}, "S3 write failed for bucket")

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating SSM Resource Data Sync (%s): %s", name, err)
	}

	d.SetId(name)

	return append(diags, resourceResourceDataSyncRead(ctx, d, meta)...)
}

func resourceResourceDataSyncRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).SSMClient(ctx)

	syncItem, err := findResourceDataSyncByName(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] SSM Resource Data Sync (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading SSM Resource Data Sync (%s): %s", d.Id(), err)
	}

	d.Set(names.AttrName, syncItem.SyncName)
	if err := d.Set("s3_destination", flattenResourceDataSyncS3Destination(syncItem.S3Destination)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting s3_destination: %s", err)
	}

	return diags
}

func resourceResourceDataSyncDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).SSMClient(ctx)

	log.Printf("[DEBUG] Deleting SSM Resource Data Sync: %s", d.Id())
	_, err := conn.DeleteResourceDataSync(ctx, &ssm.DeleteResourceDataSyncInput{
		SyncName: aws.String(d.Id()),
	})

	if errs.IsA[*awstypes.ResourceDataSyncNotFoundException](err) {
		return diags
	}
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting SSM Resource Data Sync (%s): %s", d.Id(), err)
	}

	return diags
}

func findResourceDataSyncByName(ctx context.Context, conn *ssm.Client, name string) (*awstypes.ResourceDataSyncItem, error) {
	input := &ssm.ListResourceDataSyncInput{}

	return findResourceDataSync(ctx, conn, input, func(v *awstypes.ResourceDataSyncItem) bool {
		return aws.ToString(v.SyncName) == name
	})
}

func findResourceDataSync(ctx context.Context, conn *ssm.Client, input *ssm.ListResourceDataSyncInput, filter tfslices.Predicate[*awstypes.ResourceDataSyncItem]) (*awstypes.ResourceDataSyncItem, error) {
	output, err := findResourceDataSyncs(ctx, conn, input, filter)

	if err != nil {
		return nil, err
	}

	return tfresource.AssertSingleValueResult(output)
}

func findResourceDataSyncs(ctx context.Context, conn *ssm.Client, input *ssm.ListResourceDataSyncInput, filter tfslices.Predicate[*awstypes.ResourceDataSyncItem]) ([]awstypes.ResourceDataSyncItem, error) {
	var output []awstypes.ResourceDataSyncItem

	pages := ssm.NewListResourceDataSyncPaginator(conn, input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if err != nil {
			return nil, err
		}

		for _, v := range page.ResourceDataSyncItems {
			if filter(&v) {
				output = append(output, v)
			}
		}
	}

	return output, nil
}

func flattenResourceDataSyncS3Destination(apiObject *awstypes.ResourceDataSyncS3Destination) []interface{} {
	tfMap := make(map[string]interface{})

	tfMap[names.AttrBucketName] = aws.ToString(apiObject.BucketName)
	tfMap[names.AttrRegion] = aws.ToString(apiObject.Region)
	tfMap["sync_format"] = apiObject.SyncFormat
	if apiObject.AWSKMSKeyARN != nil {
		tfMap[names.AttrKMSKeyARN] = aws.ToString(apiObject.AWSKMSKeyARN)
	}
	if apiObject.Prefix != nil {
		tfMap[names.AttrPrefix] = aws.ToString(apiObject.Prefix)
	}

	return []interface{}{tfMap}
}

func expandResourceDataSyncS3Destination(d *schema.ResourceData) *awstypes.ResourceDataSyncS3Destination {
	tfMap := d.Get("s3_destination").([]interface{})[0].(map[string]interface{})
	apiObject := &awstypes.ResourceDataSyncS3Destination{
		BucketName: aws.String(tfMap[names.AttrBucketName].(string)),
		Region:     aws.String(tfMap[names.AttrRegion].(string)),
		SyncFormat: awstypes.ResourceDataSyncS3Format(tfMap["sync_format"].(string)),
	}

	if v, ok := tfMap[names.AttrKMSKeyARN].(string); ok && v != "" {
		apiObject.AWSKMSKeyARN = aws.String(v)
	}

	if v, ok := tfMap[names.AttrPrefix].(string); ok && v != "" {
		apiObject.Prefix = aws.String(v)
	}

	return apiObject
}
