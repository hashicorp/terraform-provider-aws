// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package ec2

import (
	"context"
	"log"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	awstypes "github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/hashicorp/aws-sdk-go-base/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/id"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_vpc_ipam_scope", name="IPAM Scope")
// @Tags(identifierAttribute="id")
// @Testing(tagsTest=false)
func resourceIPAMScope() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceIPAMScopeCreate,
		ReadWithoutTimeout:   resourceIPAMScopeRead,
		UpdateWithoutTimeout: resourceIPAMScopeUpdate,
		DeleteWithoutTimeout: resourceIPAMScopeDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(3 * time.Minute),
			Update: schema.DefaultTimeout(3 * time.Minute),
			Delete: schema.DefaultTimeout(3 * time.Minute),
		},

		Schema: map[string]*schema.Schema{
			names.AttrARN: {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrDescription: {
				Type:     schema.TypeString,
				Optional: true,
			},
			"ipam_arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"ipam_id": {
				Type:     schema.TypeString,
				Required: true,
			},
			"ipam_scope_type": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"is_default": {
				Type:     schema.TypeBool,
				Computed: true,
			},
			"pool_count": {
				Type:     schema.TypeInt,
				Computed: true,
			},
			names.AttrTags:    tftags.TagsSchema(),
			names.AttrTagsAll: tftags.TagsSchemaComputed(),
		},

		CustomizeDiff: verify.SetTagsDiff,
	}
}

func resourceIPAMScopeCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).EC2Client(ctx)

	input := &ec2.CreateIpamScopeInput{
		ClientToken:       aws.String(id.UniqueId()),
		IpamId:            aws.String(d.Get("ipam_id").(string)),
		TagSpecifications: getTagSpecificationsIn(ctx, awstypes.ResourceTypeIpamScope),
	}

	if v, ok := d.GetOk(names.AttrDescription); ok {
		input.Description = aws.String(v.(string))
	}

	output, err := conn.CreateIpamScope(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating IPAM Scope: %s", err)
	}

	d.SetId(aws.ToString(output.IpamScope.IpamScopeId))

	if _, err := waitIPAMScopeCreated(ctx, conn, d.Id(), d.Timeout(schema.TimeoutCreate)); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for IPAM Scope (%s) create: %s", d.Id(), err)
	}

	return append(diags, resourceIPAMScopeRead(ctx, d, meta)...)
}

func resourceIPAMScopeRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).EC2Client(ctx)

	scope, err := findIPAMScopeByID(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] IPAM Scope (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading IPAM Scope (%s): %s", d.Id(), err)
	}

	ipamID := strings.Split(aws.ToString(scope.IpamArn), "/")[1]
	d.Set(names.AttrARN, scope.IpamScopeArn)
	d.Set(names.AttrDescription, scope.Description)
	d.Set("ipam_arn", scope.IpamArn)
	d.Set("ipam_id", ipamID)
	d.Set("ipam_scope_type", scope.IpamScopeType)
	d.Set("is_default", scope.IsDefault)
	d.Set("pool_count", scope.PoolCount)

	setTagsOut(ctx, scope.Tags)

	return diags
}

func resourceIPAMScopeUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).EC2Client(ctx)

	if d.HasChange(names.AttrDescription) {
		input := &ec2.ModifyIpamScopeInput{
			IpamScopeId: aws.String(d.Id()),
		}

		if v, ok := d.GetOk(names.AttrDescription); ok {
			input.Description = aws.String(v.(string))
		}

		_, err := conn.ModifyIpamScope(ctx, input)

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "updating IPAM Scope (%s): %s", d.Id(), err)
		}

		if _, err := waitIPAMScopeUpdated(ctx, conn, d.Id(), d.Timeout(schema.TimeoutUpdate)); err != nil {
			return sdkdiag.AppendErrorf(diags, "waiting for IPAM Scope (%s) update: %s", d.Id(), err)
		}
	}

	return append(diags, resourceIPAMScopeRead(ctx, d, meta)...)
}

func resourceIPAMScopeDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).EC2Client(ctx)

	log.Printf("[DEBUG] Deleting IPAM Scope: %s", d.Id())
	_, err := conn.DeleteIpamScope(ctx, &ec2.DeleteIpamScopeInput{
		IpamScopeId: aws.String(d.Id()),
	})

	if tfawserr.ErrCodeEquals(err, errCodeInvalidIPAMScopeIdNotFound) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting IPAM Scope: (%s): %s", d.Id(), err)
	}

	if _, err := waitIPAMScopeDeleted(ctx, conn, d.Id(), d.Timeout(schema.TimeoutDelete)); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for IPAM Scope (%s) delete: %s", d.Id(), err)
	}

	return diags
}
