// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package schemas

import (
	"context"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/schemas"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_schemas_discoverer", name="Discoverer")
// @Tags(identifierAttribute="arn")
func ResourceDiscoverer() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceDiscovererCreate,
		ReadWithoutTimeout:   resourceDiscovererRead,
		UpdateWithoutTimeout: resourceDiscovererUpdate,
		DeleteWithoutTimeout: resourceDiscovererDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"description": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringLenBetween(0, 256),
			},

			"source_arn": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: verify.ValidARN,
			},
			names.AttrTags:    tftags.TagsSchema(),
			names.AttrTagsAll: tftags.TagsSchemaComputed(),
		},

		CustomizeDiff: verify.SetTagsDiff,
	}
}

func resourceDiscovererCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).SchemasConn(ctx)

	sourceARN := d.Get("source_arn").(string)
	input := &schemas.CreateDiscovererInput{
		SourceArn: aws.String(sourceARN),
		Tags:      getTagsIn(ctx),
	}

	if v, ok := d.GetOk("description"); ok {
		input.Description = aws.String(v.(string))
	}

	log.Printf("[DEBUG] Creating EventBridge Schemas Discoverer: %s", input)
	output, err := conn.CreateDiscovererWithContext(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating EventBridge Schemas Discoverer (%s): %s", sourceARN, err)
	}

	d.SetId(aws.StringValue(output.DiscovererId))

	return append(diags, resourceDiscovererRead(ctx, d, meta)...)
}

func resourceDiscovererRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).SchemasConn(ctx)

	output, err := FindDiscovererByID(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] EventBridge Schemas Discoverer (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading EventBridge Schemas Discoverer (%s): %s", d.Id(), err)
	}

	d.Set("arn", output.DiscovererArn)
	d.Set("description", output.Description)
	d.Set("source_arn", output.SourceArn)

	return diags
}

func resourceDiscovererUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).SchemasConn(ctx)

	if d.HasChange("description") {
		input := &schemas.UpdateDiscovererInput{
			DiscovererId: aws.String(d.Id()),
			Description:  aws.String(d.Get("description").(string)),
		}

		log.Printf("[DEBUG] Updating EventBridge Schemas Discoverer: %s", input)
		_, err := conn.UpdateDiscovererWithContext(ctx, input)

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "updating EventBridge Schemas Discoverer (%s): %s", d.Id(), err)
		}
	}

	return append(diags, resourceDiscovererRead(ctx, d, meta)...)
}

func resourceDiscovererDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).SchemasConn(ctx)

	log.Printf("[INFO] Deleting EventBridge Schemas Discoverer (%s)", d.Id())
	_, err := conn.DeleteDiscovererWithContext(ctx, &schemas.DeleteDiscovererInput{
		DiscovererId: aws.String(d.Id()),
	})

	if tfawserr.ErrCodeEquals(err, schemas.ErrCodeNotFoundException) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting EventBridge Schemas Discoverer (%s): %s", d.Id(), err)
	}

	return diags
}
