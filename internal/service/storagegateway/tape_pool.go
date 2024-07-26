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
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
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
			names.AttrStorageClass: {
				Type:             schema.TypeString,
				Required:         true,
				ForceNew:         true,
				ValidateDiagFunc: enum.Validate[awstypes.TapeStorageClass](),
			},
			"retention_lock_type": {
				Type:             schema.TypeString,
				Optional:         true,
				Default:          awstypes.RetentionLockTypeNone,
				ForceNew:         true,
				ValidateDiagFunc: enum.Validate[awstypes.RetentionLockType](),
			},
			"retention_lock_time_in_days": {
				Type:         schema.TypeInt,
				Optional:     true,
				ForceNew:     true,
				Default:      0,
				ValidateFunc: validation.IntBetween(0, 36500),
			},
			names.AttrTags:    tftags.TagsSchema(),
			names.AttrTagsAll: tftags.TagsSchemaComputed(),
		},

		CustomizeDiff: verify.SetTagsDiff,
	}
}

func resourceTapePoolCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).StorageGatewayClient(ctx)

	input := &storagegateway.CreateTapePoolInput{
		PoolName:                aws.String(d.Get("pool_name").(string)),
		StorageClass:            awstypes.TapeStorageClass(d.Get(names.AttrStorageClass).(string)),
		RetentionLockType:       awstypes.RetentionLockType(d.Get("retention_lock_type").(string)),
		RetentionLockTimeInDays: aws.Int32(int32(d.Get("retention_lock_time_in_days").(int))),
		Tags:                    getTagsIn(ctx),
	}

	log.Printf("[DEBUG] Creating Storage Gateway Tape Pool: %#v", input)
	output, err := conn.CreateTapePool(ctx, input)
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating Storage Gateway Tape Pool: %s", err)
	}

	d.SetId(aws.ToString(output.PoolARN))

	return append(diags, resourceTapePoolRead(ctx, d, meta)...)
}

func resourceTapePoolRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).StorageGatewayClient(ctx)

	input := &storagegateway.ListTapePoolsInput{
		PoolARNs: []string{d.Id()},
	}

	log.Printf("[DEBUG] Reading Storage Gateway Tape Pool: %#v", input)
	output, err := conn.ListTapePools(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "listing Storage Gateway Tape Pools: %s", err)
	}

	if output == nil || len(output.PoolInfos) == 0 || aws.ToString(output.PoolInfos[0].PoolARN) != d.Id() {
		log.Printf("[WARN] Storage Gateway Tape Pool %q not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	pool := output.PoolInfos[0]

	poolArn := aws.ToString(pool.PoolARN)
	d.Set(names.AttrARN, poolArn)
	d.Set("pool_name", pool.PoolName)
	d.Set("retention_lock_time_in_days", pool.RetentionLockTimeInDays)
	d.Set("retention_lock_type", pool.RetentionLockType)
	d.Set(names.AttrStorageClass, pool.StorageClass)

	return diags
}

func resourceTapePoolUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	// Tags only.

	return append(diags, resourceTapePoolRead(ctx, d, meta)...)
}

func resourceTapePoolDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).StorageGatewayClient(ctx)

	input := &storagegateway.DeleteTapePoolInput{
		PoolARN: aws.String(d.Id()),
	}

	log.Printf("[DEBUG] Deleting Storage Gateway Tape Pool: %#v", input)
	_, err := conn.DeleteTapePool(ctx, input)

	if errs.IsAErrorMessageContains[*awstypes.InvalidGatewayRequestException](err, "The specified pool was not found") {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting Storage Gateway Tape Pool %q: %s", d.Id(), err)
	}

	return diags
}
