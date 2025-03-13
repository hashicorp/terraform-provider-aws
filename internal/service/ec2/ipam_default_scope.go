// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package ec2

import (
	"context"
	"log"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_vpc_ipam_default_scope", name="IPAM Default Scope")
// @Tags(identifierAttribute="id")
// @Testing(tagsTest=false)
func resourceIPAMDefaultScope() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceIPAMDefaultScopeCreate,
		ReadWithoutTimeout:   resourceIPAMScopeRead,
		UpdateWithoutTimeout: resourceIPAMScopeUpdate,
		DeleteWithoutTimeout: resourceIPAMDefaultScopeDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(3 * time.Minute),
			Update: schema.DefaultTimeout(3 * time.Minute),
			Delete: schema.DefaultTimeout(3 * time.Minute),
		},

		Schema: map[string]*schema.Schema{
			"default_scope_id": {
				Type:     schema.TypeString,
				Required: true,
			},
			names.AttrARN: {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrDescription: {
				Type:     schema.TypeString,
				Computed: true,
			},
			"ipam_arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"ipam_id": {
				Type:     schema.TypeString,
				Computed: true,
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
	}
}

func resourceIPAMDefaultScopeCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).EC2Client(ctx)

	scope, err := findIPAMScopeByID(ctx, conn, d.Get("default_scope_id").(string))

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading IPAM Scope (%s): %s", d.Id(), err)
	}

	d.SetId(aws.ToString(scope.IpamScopeId))

	// Configure tags.
	ignoreTagsConfig := meta.(*conns.AWSClient).IgnoreTagsConfig(ctx)
	newTags := keyValueTags(ctx, getTagsIn(ctx))
	oldTags := keyValueTags(ctx, scope.Tags).IgnoreSystem(names.EC2).IgnoreConfig(ignoreTagsConfig)

	if !oldTags.Equal(newTags) {
		if err := updateTags(ctx, conn, d.Id(), oldTags, newTags); err != nil {
			return sdkdiag.AppendErrorf(diags, "updating EC2 Default IPAM Scope (%s) tags: %s", d.Id(), err)
		}
	}

	return append(diags, resourceIPAMScopeRead(ctx, d, meta)...)
}

func resourceIPAMDefaultScopeDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	log.Printf("[WARN] Default IPAM Scope (%s) not deleted, removing from state", d.Id())
	d.SetId("")

	return diags
}
