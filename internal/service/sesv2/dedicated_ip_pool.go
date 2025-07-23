// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package sesv2

import (
	"context"
	"errors"
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/sesv2"
	"github.com/aws/aws-sdk-go-v2/service/sesv2/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_sesv2_dedicated_ip_pool", name="Dedicated IP Pool")
// @Tags(identifierAttribute="arn")
func resourceDedicatedIPPool() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceDedicatedIPPoolCreate,
		ReadWithoutTimeout:   resourceDedicatedIPPoolRead,
		UpdateWithoutTimeout: resourceDedicatedIPPoolUpdate,
		DeleteWithoutTimeout: resourceDedicatedIPPoolDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
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
	}
}

const (
	resNameDedicatedIPPool = "Dedicated IP Pool"
)

func resourceDedicatedIPPoolCreate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
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
		return create.AppendDiagError(diags, names.SESV2, create.ErrActionCreating, resNameDedicatedIPPool, d.Get("pool_name").(string), err)
	}
	if out == nil {
		return create.AppendDiagError(diags, names.SESV2, create.ErrActionCreating, resNameDedicatedIPPool, d.Get("pool_name").(string), errors.New("empty output"))
	}

	d.SetId(d.Get("pool_name").(string))
	return append(diags, resourceDedicatedIPPoolRead(ctx, d, meta)...)
}

func resourceDedicatedIPPoolRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).SESV2Client(ctx)

	out, err := findDedicatedIPPoolByName(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] SESV2 DedicatedIPPool (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return create.AppendDiagError(diags, names.SESV2, create.ErrActionReading, resNameDedicatedIPPool, d.Id(), err)
	}

	poolName := aws.ToString(out.PoolName)
	d.Set(names.AttrARN, dedicatedIPPoolARN(ctx, meta.(*conns.AWSClient), poolName))
	d.Set("pool_name", poolName)
	d.Set("scaling_mode", out.ScalingMode)

	return diags
}

func resourceDedicatedIPPoolUpdate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	// Tags only.
	return resourceDedicatedIPPoolRead(ctx, d, meta)
}

func resourceDedicatedIPPoolDelete(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).SESV2Client(ctx)

	log.Printf("[INFO] Deleting SESV2 DedicatedIPPool %s", d.Id())
	_, err := conn.DeleteDedicatedIpPool(ctx, &sesv2.DeleteDedicatedIpPoolInput{
		PoolName: aws.String(d.Id()),
	})

	if errs.IsA[*types.NotFoundException](err) {
		return diags
	}

	if err != nil {
		return create.AppendDiagError(diags, names.SESV2, create.ErrActionDeleting, resNameDedicatedIPPool, d.Id(), err)
	}

	return diags
}

func findDedicatedIPPoolByName(ctx context.Context, conn *sesv2.Client, name string) (*types.DedicatedIpPool, error) {
	input := &sesv2.GetDedicatedIpPoolInput{
		PoolName: aws.String(name),
	}

	return findDedicatedIPPool(ctx, conn, input)
}

func findDedicatedIPPool(ctx context.Context, conn *sesv2.Client, input *sesv2.GetDedicatedIpPoolInput) (*types.DedicatedIpPool, error) {
	output, err := conn.GetDedicatedIpPool(ctx, input)

	if errs.IsA[*types.NotFoundException](err) {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil || output.DedicatedIpPool == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	return output.DedicatedIpPool, nil
}

func dedicatedIPPoolARN(ctx context.Context, c *conns.AWSClient, poolName string) string {
	return c.RegionalARN(ctx, "ses", fmt.Sprintf("dedicated-ip-pool/%s", poolName))
}
