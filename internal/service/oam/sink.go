// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package oam

import (
	"context"
	"errors"
	"log"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/oam"
	"github.com/aws/aws-sdk-go-v2/service/oam/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_oam_sink", name="Sink")
// @Tags(identifierAttribute="id")
func ResourceSink() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceSinkCreate,
		ReadWithoutTimeout:   resourceSinkRead,
		UpdateWithoutTimeout: resourceSinkUpdate,
		DeleteWithoutTimeout: resourceSinkDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Timeouts: &schema.ResourceTimeout{
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

		CustomizeDiff: verify.SetTagsDiff,
	}
}

const (
	ResNameSink = "Sink"
)

func resourceSinkCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ObservabilityAccessManagerClient(ctx)

	in := &oam.CreateSinkInput{
		Name: aws.String(d.Get(names.AttrName).(string)),
		Tags: getTagsIn(ctx),
	}

	out, err := conn.CreateSink(ctx, in)
	if err != nil {
		return create.AppendDiagError(diags, names.ObservabilityAccessManager, create.ErrActionCreating, ResNameSink, d.Get(names.AttrName).(string), err)
	}

	if out == nil {
		return create.AppendDiagError(diags, names.ObservabilityAccessManager, create.ErrActionCreating, ResNameSink, d.Get(names.AttrName).(string), errors.New("empty output"))
	}

	d.SetId(aws.ToString(out.Arn))

	return append(diags, resourceSinkRead(ctx, d, meta)...)
}

func resourceSinkRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ObservabilityAccessManagerClient(ctx)

	out, err := findSinkByID(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] ObservabilityAccessManager Sink (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return create.AppendDiagError(diags, names.ObservabilityAccessManager, create.ErrActionReading, ResNameSink, d.Id(), err)
	}

	d.Set(names.AttrARN, out.Arn)
	d.Set(names.AttrName, out.Name)
	d.Set("sink_id", out.Id)

	return nil
}

func resourceSinkUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	// Tags only.
	return resourceSinkRead(ctx, d, meta)
}

func resourceSinkDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ObservabilityAccessManagerClient(ctx)

	log.Printf("[INFO] Deleting ObservabilityAccessManager Sink %s", d.Id())

	_, err := conn.DeleteSink(ctx, &oam.DeleteSinkInput{
		Identifier: aws.String(d.Id()),
	})

	if err != nil {
		var nfe *types.ResourceNotFoundException
		if errors.As(err, &nfe) {
			return nil
		}

		return create.AppendDiagError(diags, names.ObservabilityAccessManager, create.ErrActionDeleting, ResNameSink, d.Id(), err)
	}

	return nil
}

func findSinkByID(ctx context.Context, conn *oam.Client, id string) (*oam.GetSinkOutput, error) {
	in := &oam.GetSinkInput{
		Identifier: aws.String(id),
	}
	out, err := conn.GetSink(ctx, in)
	if err != nil {
		var nfe *types.ResourceNotFoundException
		if errors.As(err, &nfe) {
			return nil, &retry.NotFoundError{
				LastError:   err,
				LastRequest: in,
			}
		}

		return nil, err
	}

	if out == nil {
		return nil, tfresource.NewEmptyResultError(in)
	}

	return out, nil
}
