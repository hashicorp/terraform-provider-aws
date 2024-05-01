// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package ivs

import (
	"context"
	"errors"
	"log"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ivs"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_ivs_playback_key_pair", name="Playback Key Pair")
// @Tags(identifierAttribute="id")
func ResourcePlaybackKeyPair() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourcePlaybackKeyPairCreate,
		ReadWithoutTimeout:   resourcePlaybackKeyPairRead,
		DeleteWithoutTimeout: resourcePlaybackKeyPairDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(5 * time.Minute),
			Delete: schema.DefaultTimeout(5 * time.Minute),
		},

		Schema: map[string]*schema.Schema{
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"fingerprint": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"name": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
				ForceNew: true,
			},
			"public_key": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			names.AttrTags:    tftags.TagsSchemaForceNew(),
			names.AttrTagsAll: tftags.TagsSchemaComputed(),
		},

		CustomizeDiff: verify.SetTagsDiff,
	}
}

const (
	ResNamePlaybackKeyPair = "Playback Key Pair"
)

func resourcePlaybackKeyPairCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	conn := meta.(*conns.AWSClient).IVSConn(ctx)

	in := &ivs.ImportPlaybackKeyPairInput{
		PublicKeyMaterial: aws.String(d.Get("public_key").(string)),
		Tags:              getTagsIn(ctx),
	}

	if v, ok := d.GetOk("name"); ok {
		in.Name = aws.String(v.(string))
	}

	out, err := conn.ImportPlaybackKeyPairWithContext(ctx, in)
	if err != nil {
		return create.AppendDiagError(diags, names.IVS, create.ErrActionCreating, ResNamePlaybackKeyPair, d.Get("name").(string), err)
	}

	if out == nil || out.KeyPair == nil {
		return create.AppendDiagError(diags, names.IVS, create.ErrActionCreating, ResNamePlaybackKeyPair, d.Get("name").(string), errors.New("empty output"))
	}

	d.SetId(aws.StringValue(out.KeyPair.Arn))

	if _, err := waitPlaybackKeyPairCreated(ctx, conn, d.Id(), d.Timeout(schema.TimeoutCreate)); err != nil {
		return create.AppendDiagError(diags, names.IVS, create.ErrActionWaitingForCreation, ResNamePlaybackKeyPair, d.Id(), err)
	}

	return append(diags, resourcePlaybackKeyPairRead(ctx, d, meta)...)
}

func resourcePlaybackKeyPairRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	conn := meta.(*conns.AWSClient).IVSConn(ctx)

	out, err := FindPlaybackKeyPairByID(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] IVS PlaybackKeyPair (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return create.AppendDiagError(diags, names.IVS, create.ErrActionReading, ResNamePlaybackKeyPair, d.Id(), err)
	}

	d.Set("arn", out.Arn)
	d.Set("name", out.Name)
	d.Set("fingerprint", out.Fingerprint)

	return diags
}

func resourcePlaybackKeyPairDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	conn := meta.(*conns.AWSClient).IVSConn(ctx)

	log.Printf("[INFO] Deleting IVS PlaybackKeyPair %s", d.Id())

	_, err := conn.DeletePlaybackKeyPairWithContext(ctx, &ivs.DeletePlaybackKeyPairInput{
		Arn: aws.String(d.Id()),
	})

	if tfawserr.ErrCodeEquals(err, ivs.ErrCodeResourceNotFoundException) {
		return diags
	}

	if err != nil {
		return create.AppendDiagError(diags, names.IVS, create.ErrActionDeleting, ResNamePlaybackKeyPair, d.Id(), err)
	}

	if _, err := waitPlaybackKeyPairDeleted(ctx, conn, d.Id(), d.Timeout(schema.TimeoutDelete)); err != nil {
		return create.AppendDiagError(diags, names.IVS, create.ErrActionWaitingForDeletion, ResNamePlaybackKeyPair, d.Id(), err)
	}

	return diags
}
