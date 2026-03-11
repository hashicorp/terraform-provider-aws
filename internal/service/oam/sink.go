// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

// DONOTCOPY: Copying old resources spreads bad habits. Use skaff instead.

package oam

import (
	"context"
	"log"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/oam"
	awstypes "github.com/aws/aws-sdk-go-v2/service/oam/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_oam_sink", name="Sink")
// @Tags(identifierAttribute="arn")
func resourceSink() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceSinkCreate,
		ReadWithoutTimeout:   resourceSinkRead,
		UpdateWithoutTimeout: resourceSinkUpdate,
		DeleteWithoutTimeout: resourceSinkDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Timeouts: &schema.ResourceTimeout{
			// These aren't used but are retained for backwards compatibility.
			Create: schema.DefaultTimeout(1 * time.Minute),
			Update: schema.DefaultTimeout(1 * time.Minute),
			Delete: schema.DefaultTimeout(1 * time.Minute),
		},

		Schema: map[string]*schema.Schema{
			names.AttrARN: {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrName: {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"sink_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrTags:    tftags.TagsSchema(),
			names.AttrTagsAll: tftags.TagsSchemaComputed(),
		},
	}
}

func resourceSinkCreate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ObservabilityAccessManagerClient(ctx)

	name := d.Get(names.AttrName).(string)
	in := oam.CreateSinkInput{
		Name: aws.String(name),
		Tags: getTagsIn(ctx),
	}

	out, err := conn.CreateSink(ctx, &in)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating ObservabilityAccessManager Sink (%s): %s", name, err)
	}

	d.SetId(aws.ToString(out.Arn))

	return append(diags, resourceSinkRead(ctx, d, meta)...)
}

func resourceSinkRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ObservabilityAccessManagerClient(ctx)

	out, err := findSinkByID(ctx, conn, d.Id())

	if !d.IsNewResource() && retry.NotFound(err) {
		log.Printf("[WARN] ObservabilityAccessManager Sink (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading ObservabilityAccessManager Sink (%s): %s", d.Id(), err)
	}

	d.Set(names.AttrARN, out.Arn)
	d.Set(names.AttrName, out.Name)
	d.Set("sink_id", out.Id)

	return diags
}

func resourceSinkUpdate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	// Tags only.
	return resourceSinkRead(ctx, d, meta)
}

func resourceSinkDelete(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ObservabilityAccessManagerClient(ctx)

	log.Printf("[INFO] Deleting ObservabilityAccessManager Sink: %s", d.Id())
	in := oam.DeleteSinkInput{
		Identifier: aws.String(d.Id()),
	}
	_, err := conn.DeleteSink(ctx, &in)

	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting ObservabilityAccessManager Sink (%s): %s", d.Id(), err)
	}

	return diags
}

func findSinkByID(ctx context.Context, conn *oam.Client, id string) (*oam.GetSinkOutput, error) {
	in := oam.GetSinkInput{
		Identifier: aws.String(id),
	}

	return findSink(ctx, conn, &in)
}

func findSink(ctx context.Context, conn *oam.Client, input *oam.GetSinkInput) (*oam.GetSinkOutput, error) {
	output, err := conn.GetSink(ctx, input)

	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return nil, &retry.NotFoundError{
			LastError: err,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil || output.Arn == nil {
		return nil, tfresource.NewEmptyResultError()
	}

	return output, nil
}
