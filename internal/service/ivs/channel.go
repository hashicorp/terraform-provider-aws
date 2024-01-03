// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package ivs

import (
	"context"
	"errors"
	"log"
	"time"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ivs"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_ivs_channel", name="Channel")
// @Tags(identifierAttribute="id")
func ResourceChannel() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceChannelCreate,
		ReadWithoutTimeout:   resourceChannelRead,
		UpdateWithoutTimeout: resourceChannelUpdate,
		DeleteWithoutTimeout: resourceChannelDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(5 * time.Minute),
			Update: schema.DefaultTimeout(5 * time.Minute),
			Delete: schema.DefaultTimeout(5 * time.Minute),
		},

		Schema: map[string]*schema.Schema{
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"authorized": {
				Type:     schema.TypeBool,
				Optional: true,
				Computed: true,
			},
			"ingest_endpoint": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"latency_mode": {
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				ValidateFunc: validation.StringInSlice(ivs.ChannelLatencyMode_Values(), false),
			},
			"name": {
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				ValidateFunc: validation.StringMatch(regexache.MustCompile(`^[0-9A-Za-z_-]{0,128}$`), "must contain only alphanumeric characters, hyphen, or underscore and at most 128 characters"),
			},
			"playback_url": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"recording_configuration_arn": {
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				ValidateFunc: verify.ValidARN,
			},
			names.AttrTags:    tftags.TagsSchema(),
			names.AttrTagsAll: tftags.TagsSchemaComputed(),
			"type": {
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				ValidateFunc: validation.StringInSlice(ivs.ChannelType_Values(), false),
			},
		},

		CustomizeDiff: verify.SetTagsDiff,
	}
}

const (
	ResNameChannel = "Channel"
)

func resourceChannelCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	conn := meta.(*conns.AWSClient).IVSConn(ctx)

	in := &ivs.CreateChannelInput{
		Tags: getTagsIn(ctx),
	}

	if v, ok := d.GetOk("authorized"); ok {
		in.Authorized = aws.Bool(v.(bool))
	}

	if v, ok := d.GetOk("latency_mode"); ok {
		in.LatencyMode = aws.String(v.(string))
	}

	if v, ok := d.GetOk("name"); ok {
		in.Name = aws.String(v.(string))
	}

	if v, ok := d.GetOk("recording_configuration_arn"); ok {
		in.RecordingConfigurationArn = aws.String(v.(string))
	}

	if v, ok := d.GetOk("type"); ok {
		in.Type = aws.String(v.(string))
	}

	out, err := conn.CreateChannelWithContext(ctx, in)
	if err != nil {
		return create.AppendDiagError(diags, names.IVS, create.ErrActionCreating, ResNameChannel, d.Get("name").(string), err)
	}

	if out == nil || out.Channel == nil {
		return create.AppendDiagError(diags, names.IVS, create.ErrActionCreating, ResNameChannel, d.Get("name").(string), errors.New("empty output"))
	}

	d.SetId(aws.StringValue(out.Channel.Arn))

	if _, err := waitChannelCreated(ctx, conn, d.Id(), d.Timeout(schema.TimeoutCreate)); err != nil {
		return create.AppendDiagError(diags, names.IVS, create.ErrActionWaitingForCreation, ResNameChannel, d.Id(), err)
	}

	return append(diags, resourceChannelRead(ctx, d, meta)...)
}

func resourceChannelRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	conn := meta.(*conns.AWSClient).IVSConn(ctx)

	out, err := FindChannelByID(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] IVS Channel (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return create.AppendDiagError(diags, names.IVS, create.ErrActionReading, ResNameChannel, d.Id(), err)
	}

	d.Set("arn", out.Arn)
	d.Set("authorized", out.Authorized)
	d.Set("ingest_endpoint", out.IngestEndpoint)
	d.Set("latency_mode", out.LatencyMode)
	d.Set("name", out.Name)
	d.Set("playback_url", out.PlaybackUrl)
	d.Set("recording_configuration_arn", out.RecordingConfigurationArn)
	d.Set("type", out.Type)

	return diags
}

func resourceChannelUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	conn := meta.(*conns.AWSClient).IVSConn(ctx)

	update := false

	arn := d.Id()
	in := &ivs.UpdateChannelInput{
		Arn: aws.String(arn),
	}

	if d.HasChanges("authorized") {
		in.Authorized = aws.Bool(d.Get("authorized").(bool))
		update = true
	}

	if d.HasChanges("latency_mode") {
		in.LatencyMode = aws.String(d.Get("latency_mode").(string))
		update = true
	}

	if d.HasChanges("name") {
		in.Name = aws.String(d.Get("name").(string))
		update = true
	}

	if d.HasChanges("recording_configuration_arn") {
		in.RecordingConfigurationArn = aws.String(d.Get("recording_configuration_arn").(string))
		update = true
	}

	if d.HasChanges("type") {
		in.Type = aws.String(d.Get("type").(string))
		update = true
	}

	if !update {
		return diags
	}

	log.Printf("[DEBUG] Updating IVS Channel (%s): %#v", d.Id(), in)

	out, err := conn.UpdateChannelWithContext(ctx, in)
	if err != nil {
		return create.AppendDiagError(diags, names.IVS, create.ErrActionUpdating, ResNameChannel, d.Id(), err)
	}

	if _, err := waitChannelUpdated(ctx, conn, *out.Channel.Arn, d.Timeout(schema.TimeoutUpdate), in); err != nil {
		return create.AppendDiagError(diags, names.IVS, create.ErrActionWaitingForUpdate, ResNameChannel, d.Id(), err)
	}

	return append(diags, resourceChannelRead(ctx, d, meta)...)
}

func resourceChannelDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	conn := meta.(*conns.AWSClient).IVSConn(ctx)

	log.Printf("[INFO] Deleting IVS Channel %s", d.Id())

	_, err := conn.DeleteChannelWithContext(ctx, &ivs.DeleteChannelInput{
		Arn: aws.String(d.Id()),
	})

	if err != nil {
		if tfawserr.ErrCodeEquals(err, ivs.ErrCodeResourceNotFoundException) {
			return diags
		}

		return create.AppendDiagError(diags, names.IVS, create.ErrActionDeleting, ResNameChannel, d.Id(), err)
	}

	if _, err := waitChannelDeleted(ctx, conn, d.Id(), d.Timeout(schema.TimeoutDelete)); err != nil {
		return create.AppendDiagError(diags, names.IVS, create.ErrActionWaitingForDeletion, ResNameChannel, d.Id(), err)
	}

	return diags
}
