// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package schemas

import (
	"context"
	"log"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/schemas"
	awstypes "github.com/aws/aws-sdk-go-v2/service/schemas/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_schemas_discoverer", name="Discoverer")
// @Tags(identifierAttribute="arn")
func resourceDiscoverer() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceDiscovererCreate,
		ReadWithoutTimeout:   resourceDiscovererRead,
		UpdateWithoutTimeout: resourceDiscovererUpdate,
		DeleteWithoutTimeout: resourceDiscovererDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			names.AttrARN: {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrDescription: {
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
	conn := meta.(*conns.AWSClient).SchemasClient(ctx)

	sourceARN := d.Get("source_arn").(string)
	input := &schemas.CreateDiscovererInput{
		SourceArn: aws.String(sourceARN),
		Tags:      getTagsIn(ctx),
	}

	if v, ok := d.GetOk(names.AttrDescription); ok {
		input.Description = aws.String(v.(string))
	}

	output, err := conn.CreateDiscoverer(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating EventBridge Schemas Discoverer (%s): %s", sourceARN, err)
	}

	d.SetId(aws.ToString(output.DiscovererId))

	return append(diags, resourceDiscovererRead(ctx, d, meta)...)
}

func resourceDiscovererRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).SchemasClient(ctx)

	output, err := findDiscovererByID(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] EventBridge Schemas Discoverer (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading EventBridge Schemas Discoverer (%s): %s", d.Id(), err)
	}

	d.Set(names.AttrARN, output.DiscovererArn)
	d.Set(names.AttrDescription, output.Description)
	d.Set("source_arn", output.SourceArn)

	return diags
}

func resourceDiscovererUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).SchemasClient(ctx)

	if d.HasChange(names.AttrDescription) {
		input := &schemas.UpdateDiscovererInput{
			DiscovererId: aws.String(d.Id()),
			Description:  aws.String(d.Get(names.AttrDescription).(string)),
		}

		_, err := conn.UpdateDiscoverer(ctx, input)

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "updating EventBridge Schemas Discoverer (%s): %s", d.Id(), err)
		}
	}

	return append(diags, resourceDiscovererRead(ctx, d, meta)...)
}

func resourceDiscovererDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).SchemasClient(ctx)

	log.Printf("[INFO] Deleting EventBridge Schemas Discoverer (%s)", d.Id())
	_, err := conn.DeleteDiscoverer(ctx, &schemas.DeleteDiscovererInput{
		DiscovererId: aws.String(d.Id()),
	})

	if errs.IsA[*awstypes.NotFoundException](err) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting EventBridge Schemas Discoverer (%s): %s", d.Id(), err)
	}

	return diags
}

func findDiscovererByID(ctx context.Context, conn *schemas.Client, id string) (*schemas.DescribeDiscovererOutput, error) {
	input := &schemas.DescribeDiscovererInput{
		DiscovererId: aws.String(id),
	}

	output, err := conn.DescribeDiscoverer(ctx, input)

	if errs.IsA[*awstypes.NotFoundException](err) {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	return output, nil
}
