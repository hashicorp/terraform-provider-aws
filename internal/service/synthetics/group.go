// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package synthetics

import (
	"context"
	"log"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/synthetics"
	awstypes "github.com/aws/aws-sdk-go-v2/service/synthetics/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_synthetics_group", name="Group")
// @Tags(identifierAttribute="arn")
func ResourceGroup() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceGroupCreate,
		ReadWithoutTimeout:   resourceGroupRead,
		UpdateWithoutTimeout: resourceGroupUpdate,
		DeleteWithoutTimeout: resourceGroupDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			names.AttrARN: {
				Type:     schema.TypeString,
				Computed: true,
			},
			"group_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrName: {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			names.AttrTags:    tftags.TagsSchema(),
			names.AttrTagsAll: tftags.TagsSchemaComputed(),
		},

		CustomizeDiff: verify.SetTagsDiff,
	}
}

func resourceGroupCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).SyntheticsClient(ctx)

	name := d.Get(names.AttrName).(string)
	in := &synthetics.CreateGroupInput{
		Name: aws.String(name),
		Tags: getTagsIn(ctx),
	}

	out, err := conn.CreateGroup(ctx, in)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating Synthetics Group (%s): %s", name, err)
	}

	if out == nil || out.Group == nil {
		return sdkdiag.AppendErrorf(diags, "creating Synthetics Group (%s): Empty output", name)
	}

	d.SetId(aws.ToString(out.Group.Name))

	return append(diags, resourceGroupRead(ctx, d, meta)...)
}

func resourceGroupRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).SyntheticsClient(ctx)

	group, err := FindGroupByName(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] Synthetics Group (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading Synthetics Group (%s): %s", d.Id(), err)
	}

	d.Set(names.AttrARN, group.Arn)
	d.Set("group_id", group.Id)
	d.Set(names.AttrName, group.Name)

	setTagsOut(ctx, group.Tags)

	return diags
}

func resourceGroupUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	// Tags only.
	return resourceGroupRead(ctx, d, meta)
}

func resourceGroupDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).SyntheticsClient(ctx)

	log.Printf("[DEBUG] Deleting Synthetics Group %s", d.Id())

	_, err := conn.DeleteGroup(ctx, &synthetics.DeleteGroupInput{
		GroupIdentifier: aws.String(d.Id()),
	})

	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting Synthetics Group (%s): %s", d.Id(), err)
	}

	return diags
}
