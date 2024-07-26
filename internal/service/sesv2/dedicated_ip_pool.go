// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package sesv2

import (
	"context"
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/aws/arn"
	"github.com/aws/aws-sdk-go-v2/service/sesv2"
	"github.com/aws/aws-sdk-go-v2/service/sesv2/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_sesv2_dedicated_ip_pool", name="Dedicated IP Pool")
// @Tags(identifierAttribute="arn")
func ResourceDedicatedIPPool() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceDedicatedIPPoolCreate,
		ReadWithoutTimeout:   resourceDedicatedIPPoolRead,
		UpdateWithoutTimeout: resourceDedicatedIPPoolUpdate,
		DeleteWithoutTimeout: resourceDedicatedIPPoolDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(30 * time.Minute),
			Update: schema.DefaultTimeout(30 * time.Minute),
			Delete: schema.DefaultTimeout(30 * time.Minute),
		},

		Schema: map[string]*schema.Schema{
			names.AttrARN: {
				Type:     schema.TypeString,
				Computed: true,
			},
			"pool_name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"scaling_mode": {
				Type:             schema.TypeString,
				Optional:         true,
				Computed:         true,
				ForceNew:         true,
				ValidateDiagFunc: enum.Validate[types.ScalingMode](),
			},
			names.AttrTags:    tftags.TagsSchema(),
			names.AttrTagsAll: tftags.TagsSchemaComputed(),
		},

		CustomizeDiff: verify.SetTagsDiff,
	}
}

const (
	ResNameDedicatedIPPool = "Dedicated IP Pool"
)

func resourceDedicatedIPPoolCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).SESV2Client(ctx)

	in := &sesv2.CreateDedicatedIpPoolInput{
		PoolName: aws.String(d.Get("pool_name").(string)),
		Tags:     getTagsIn(ctx),
	}
	if v, ok := d.GetOk("scaling_mode"); ok {
		in.ScalingMode = types.ScalingMode(v.(string))
	}

	out, err := conn.CreateDedicatedIpPool(ctx, in)
	if err != nil {
		return create.AppendDiagError(diags, names.SESV2, create.ErrActionCreating, ResNameDedicatedIPPool, d.Get("pool_name").(string), err)
	}
	if out == nil {
		return create.AppendDiagError(diags, names.SESV2, create.ErrActionCreating, ResNameDedicatedIPPool, d.Get("pool_name").(string), errors.New("empty output"))
	}

	d.SetId(d.Get("pool_name").(string))
	return append(diags, resourceDedicatedIPPoolRead(ctx, d, meta)...)
}

func resourceDedicatedIPPoolRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).SESV2Client(ctx)

	out, err := FindDedicatedIPPoolByID(ctx, conn, d.Id())
	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] SESV2 DedicatedIPPool (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}
	if err != nil {
		return create.AppendDiagError(diags, names.SESV2, create.ErrActionReading, ResNameDedicatedIPPool, d.Id(), err)
	}
	poolName := aws.ToString(out.DedicatedIpPool.PoolName)
	d.Set("pool_name", poolName)
	d.Set("scaling_mode", string(out.DedicatedIpPool.ScalingMode))
	d.Set(names.AttrARN, poolNameToARN(meta, poolName))

	return diags
}

func resourceDedicatedIPPoolUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	// Tags only.
	return resourceDedicatedIPPoolRead(ctx, d, meta)
}

func resourceDedicatedIPPoolDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).SESV2Client(ctx)

	log.Printf("[INFO] Deleting SESV2 DedicatedIPPool %s", d.Id())
	_, err := conn.DeleteDedicatedIpPool(ctx, &sesv2.DeleteDedicatedIpPoolInput{
		PoolName: aws.String(d.Id()),
	})

	if err != nil {
		var nfe *types.NotFoundException
		if errors.As(err, &nfe) {
			return diags
		}
		return create.AppendDiagError(diags, names.SESV2, create.ErrActionDeleting, ResNameDedicatedIPPool, d.Id(), err)
	}

	return diags
}

func FindDedicatedIPPoolByID(ctx context.Context, conn *sesv2.Client, id string) (*sesv2.GetDedicatedIpPoolOutput, error) {
	in := &sesv2.GetDedicatedIpPoolInput{
		PoolName: aws.String(id),
	}
	out, err := conn.GetDedicatedIpPool(ctx, in)
	if err != nil {
		var nfe *types.NotFoundException
		if errors.As(err, &nfe) {
			return nil, &retry.NotFoundError{
				LastError:   err,
				LastRequest: in,
			}
		}

		return nil, err
	}

	if out == nil || out.DedicatedIpPool == nil {
		return nil, tfresource.NewEmptyResultError(in)
	}

	return out, nil
}

func poolNameToARN(meta interface{}, poolName string) string {
	return arn.ARN{
		Partition: meta.(*conns.AWSClient).Partition,
		Service:   "ses",
		Region:    meta.(*conns.AWSClient).Region,
		AccountID: meta.(*conns.AWSClient).AccountID,
		Resource:  fmt.Sprintf("dedicated-ip-pool/%s", poolName),
	}.String()
}
