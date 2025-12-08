// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package storagegateway

import (
	"context"
	"log"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/storagegateway"
	awstypes "github.com/aws/aws-sdk-go-v2/service/storagegateway/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_storagegateway_tape_pool", name="Tape Pool")
// @Tags(identifierAttribute="arn")
func resourceTapePool() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceTapePoolCreate,
		ReadWithoutTimeout:   resourceTapePoolRead,
		UpdateWithoutTimeout: resourceTapePoolUpdate,
		DeleteWithoutTimeout: resourceTapePoolDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			names.AttrARN: {
				Type:     schema.TypeString,
				Computed: true,
			},
			"pool_name": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringLenBetween(1, 100),
			},
			"retention_lock_time_in_days": {
				Type:         schema.TypeInt,
				Optional:     true,
				ForceNew:     true,
				Default:      0,
				ValidateFunc: validation.IntBetween(0, 36500),
			},
			"retention_lock_type": {
				Type:             schema.TypeString,
				Optional:         true,
				Default:          awstypes.RetentionLockTypeNone,
				ForceNew:         true,
				ValidateDiagFunc: enum.Validate[awstypes.RetentionLockType](),
			},
			names.AttrStorageClass: {
				Type:             schema.TypeString,
				Required:         true,
				ForceNew:         true,
				ValidateDiagFunc: enum.Validate[awstypes.TapeStorageClass](),
			},
			names.AttrTags:    tftags.TagsSchema(),
			names.AttrTagsAll: tftags.TagsSchemaComputed(),
		},
	}
}

func resourceTapePoolCreate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).StorageGatewayClient(ctx)

	name := d.Get("pool_name").(string)
	input := &storagegateway.CreateTapePoolInput{
		PoolName:                aws.String(name),
		StorageClass:            awstypes.TapeStorageClass(d.Get(names.AttrStorageClass).(string)),
		RetentionLockType:       awstypes.RetentionLockType(d.Get("retention_lock_type").(string)),
		RetentionLockTimeInDays: aws.Int32(int32(d.Get("retention_lock_time_in_days").(int))),
		Tags:                    getTagsIn(ctx),
	}

	output, err := conn.CreateTapePool(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating Storage Gateway Tape Pool (%s): %s", name, err)
	}

	d.SetId(aws.ToString(output.PoolARN))

	return append(diags, resourceTapePoolRead(ctx, d, meta)...)
}

func resourceTapePoolRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).StorageGatewayClient(ctx)

	pool, err := findTapePoolByARN(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] Storage Gateway Tape Pool (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading Storage Gateway Tape Pool (%s): %s", d.Id(), err)
	}

	d.Set(names.AttrARN, pool.PoolARN)
	d.Set("pool_name", pool.PoolName)
	d.Set("retention_lock_time_in_days", pool.RetentionLockTimeInDays)
	d.Set("retention_lock_type", pool.RetentionLockType)
	d.Set(names.AttrStorageClass, pool.StorageClass)

	return diags
}

func resourceTapePoolUpdate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics

	// Tags only.

	return append(diags, resourceTapePoolRead(ctx, d, meta)...)
}

func resourceTapePoolDelete(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).StorageGatewayClient(ctx)

	log.Printf("[DEBUG] Deleting Storage Gateway Tape Pool: %s", d.Id())
	_, err := conn.DeleteTapePool(ctx, &storagegateway.DeleteTapePoolInput{
		PoolARN: aws.String(d.Id()),
	})

	if errs.IsAErrorMessageContains[*awstypes.InvalidGatewayRequestException](err, "The specified pool was not found") {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting Storage Gateway Tape Pool (%s): %s", d.Id(), err)
	}

	return diags
}

func findTapePoolByARN(ctx context.Context, conn *storagegateway.Client, arn string) (*awstypes.PoolInfo, error) {
	input := &storagegateway.ListTapePoolsInput{
		PoolARNs: []string{arn},
	}

	return findTapePool(ctx, conn, input)
}

func findTapePool(ctx context.Context, conn *storagegateway.Client, input *storagegateway.ListTapePoolsInput) (*awstypes.PoolInfo, error) {
	output, err := findTapePools(ctx, conn, input)

	if err != nil {
		return nil, err
	}

	return tfresource.AssertSingleValueResult(output)
}

func findTapePools(ctx context.Context, conn *storagegateway.Client, input *storagegateway.ListTapePoolsInput) ([]awstypes.PoolInfo, error) {
	var output []awstypes.PoolInfo

	pages := storagegateway.NewListTapePoolsPaginator(conn, input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if err != nil {
			return nil, err
		}

		output = append(output, page.PoolInfos...)
	}

	return output, nil
}
