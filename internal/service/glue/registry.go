// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package glue

import (
	"context"
	"log"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/glue"
	awstypes "github.com/aws/aws-sdk-go-v2/service/glue/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_glue_registry", name="Registry")
// @Tags(identifierAttribute="arn")
func ResourceRegistry() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceRegistryCreate,
		ReadWithoutTimeout:   resourceRegistryRead,
		UpdateWithoutTimeout: resourceRegistryUpdate,
		DeleteWithoutTimeout: resourceRegistryDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		CustomizeDiff: verify.SetTagsDiff,

		Schema: map[string]*schema.Schema{
			names.AttrARN: {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrDescription: {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringLenBetween(0, 2048),
			},
			"registry_name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
				ValidateFunc: validation.All(
					validation.StringLenBetween(1, 255),
					validation.StringMatch(regexache.MustCompile(`[0-9A-Za-z_$#-]+$`), ""),
				),
			},
			names.AttrTags:    tftags.TagsSchema(),
			names.AttrTagsAll: tftags.TagsSchemaComputed(),
		},
	}
}

func resourceRegistryCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).GlueClient(ctx)

	input := &glue.CreateRegistryInput{
		RegistryName: aws.String(d.Get("registry_name").(string)),
		Tags:         getTagsIn(ctx),
	}

	if v, ok := d.GetOk(names.AttrDescription); ok {
		input.Description = aws.String(v.(string))
	}

	log.Printf("[DEBUG] Creating Glue Registry: %+v", input)
	output, err := conn.CreateRegistry(ctx, input)
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating Glue Registry: %s", err)
	}
	d.SetId(aws.ToString(output.RegistryArn))

	return append(diags, resourceRegistryRead(ctx, d, meta)...)
}

func resourceRegistryRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).GlueClient(ctx)

	output, err := FindRegistryByID(ctx, conn, d.Id())
	if err != nil {
		if errs.IsA[*awstypes.EntityNotFoundException](err) {
			log.Printf("[WARN] Glue Registry (%s) not found, removing from state", d.Id())
			d.SetId("")
			return diags
		}
		return sdkdiag.AppendErrorf(diags, "reading Glue Registry (%s): %s", d.Id(), err)
	}

	if output == nil {
		log.Printf("[WARN] Glue Registry (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	arn := aws.ToString(output.RegistryArn)
	d.Set(names.AttrARN, arn)
	d.Set(names.AttrDescription, output.Description)
	d.Set("registry_name", output.RegistryName)

	return diags
}

func resourceRegistryUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).GlueClient(ctx)

	if d.HasChanges(names.AttrDescription) {
		input := &glue.UpdateRegistryInput{
			RegistryId: createRegistryID(d.Id()),
		}

		if v, ok := d.GetOk(names.AttrDescription); ok {
			input.Description = aws.String(v.(string))
		}

		log.Printf("[DEBUG] Updating Glue Registry: %#v", input)
		_, err := conn.UpdateRegistry(ctx, input)
		if err != nil {
			return sdkdiag.AppendErrorf(diags, "updating Glue Registry (%s): %s", d.Id(), err)
		}
	}

	return append(diags, resourceRegistryRead(ctx, d, meta)...)
}

func resourceRegistryDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).GlueClient(ctx)

	log.Printf("[DEBUG] Deleting Glue Registry: %s", d.Id())
	input := &glue.DeleteRegistryInput{
		RegistryId: createRegistryID(d.Id()),
	}

	_, err := conn.DeleteRegistry(ctx, input)
	if err != nil {
		if errs.IsA[*awstypes.EntityNotFoundException](err) {
			return diags
		}
		return sdkdiag.AppendErrorf(diags, "deleting Glue Registry (%s): %s", d.Id(), err)
	}

	_, err = waitRegistryDeleted(ctx, conn, d.Id())
	if err != nil {
		if errs.IsA[*awstypes.EntityNotFoundException](err) {
			return diags
		}
		return sdkdiag.AppendErrorf(diags, "waiting for Glue Registry (%s) to be deleted: %s", d.Id(), err)
	}

	return diags
}
