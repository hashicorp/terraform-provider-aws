// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package devicefarm

import (
	"context"
	"log"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/devicefarm"
	awstypes "github.com/aws/aws-sdk-go-v2/service/devicefarm/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_devicefarm_network_profile", name="Network Profile")
// @Tags(identifierAttribute="arn")
func resourceNetworkProfile() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceNetworkProfileCreate,
		ReadWithoutTimeout:   resourceNetworkProfileRead,
		UpdateWithoutTimeout: resourceNetworkProfileUpdate,
		DeleteWithoutTimeout: resourceNetworkProfileDelete,

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
				ValidateFunc: validation.StringLenBetween(0, 16384),
			},
			"downlink_bandwidth_bits": {
				Type:         schema.TypeInt,
				Optional:     true,
				Default:      104857600,
				ValidateFunc: validation.IntBetween(0, 104857600),
			},
			"downlink_delay_ms": {
				Type:         schema.TypeInt,
				Optional:     true,
				ValidateFunc: validation.IntBetween(0, 2000),
			},
			"downlink_jitter_ms": {
				Type:         schema.TypeInt,
				Optional:     true,
				ValidateFunc: validation.IntBetween(0, 2000),
			},
			"downlink_loss_percent": {
				Type:         schema.TypeInt,
				Optional:     true,
				ValidateFunc: validation.IntBetween(0, 100),
			},
			names.AttrName: {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringLenBetween(0, 256),
			},
			"project_arn": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: verify.ValidARN,
			},
			"uplink_bandwidth_bits": {
				Type:         schema.TypeInt,
				Optional:     true,
				Default:      104857600,
				ValidateFunc: validation.IntBetween(0, 104857600),
			},
			"uplink_delay_ms": {
				Type:         schema.TypeInt,
				Optional:     true,
				ValidateFunc: validation.IntBetween(0, 2000),
			},
			"uplink_jitter_ms": {
				Type:         schema.TypeInt,
				Optional:     true,
				ValidateFunc: validation.IntBetween(0, 2000),
			},
			"uplink_loss_percent": {
				Type:         schema.TypeInt,
				Optional:     true,
				ValidateFunc: validation.IntBetween(0, 100),
			},
			names.AttrType: {
				Type:             schema.TypeString,
				Optional:         true,
				Default:          awstypes.NetworkProfileTypePrivate,
				ValidateDiagFunc: enum.Validate[awstypes.NetworkProfileType](),
			},
			names.AttrTags:    tftags.TagsSchema(),
			names.AttrTagsAll: tftags.TagsSchemaComputed(),
		},
		CustomizeDiff: verify.SetTagsDiff,
	}
}

func resourceNetworkProfileCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).DeviceFarmClient(ctx)

	name := d.Get(names.AttrName).(string)
	input := &devicefarm.CreateNetworkProfileInput{
		Name:       aws.String(name),
		ProjectArn: aws.String(d.Get("project_arn").(string)),
	}

	if v, ok := d.GetOk(names.AttrDescription); ok {
		input.Description = aws.String(v.(string))
	}

	if v, ok := d.GetOk(names.AttrType); ok {
		input.Type = awstypes.NetworkProfileType(v.(string))
	}

	if v, ok := d.GetOk("downlink_bandwidth_bits"); ok {
		input.DownlinkBandwidthBits = aws.Int64(int64(v.(int)))
	}

	if v, ok := d.GetOk("downlink_delay_ms"); ok {
		input.DownlinkDelayMs = aws.Int64(int64(v.(int)))
	}

	if v, ok := d.GetOk("downlink_jitter_ms"); ok {
		input.DownlinkJitterMs = aws.Int64(int64(v.(int)))
	}

	if v, ok := d.GetOk("downlink_loss_percent"); ok {
		input.DownlinkLossPercent = int32(v.(int))
	}

	if v, ok := d.GetOk("uplink_bandwidth_bits"); ok {
		input.UplinkBandwidthBits = aws.Int64(int64(v.(int)))
	}

	if v, ok := d.GetOk("uplink_delay_ms"); ok {
		input.UplinkDelayMs = aws.Int64(int64(v.(int)))
	}

	if v, ok := d.GetOk("uplink_jitter_ms"); ok {
		input.UplinkJitterMs = aws.Int64(int64(v.(int)))
	}

	if v, ok := d.GetOk("uplink_loss_percent"); ok {
		input.UplinkLossPercent = int32(v.(int))
	}

	output, err := conn.CreateNetworkProfile(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating DeviceFarm Network Profile (%s): %s", name, err)
	}

	d.SetId(aws.ToString(output.NetworkProfile.Arn))

	if err := createTags(ctx, conn, d.Id(), getTagsIn(ctx)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting DeviceFarm Network Profile (%s) tags: %s", d.Id(), err)
	}

	return append(diags, resourceNetworkProfileRead(ctx, d, meta)...)
}

func resourceNetworkProfileRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).DeviceFarmClient(ctx)

	project, err := findNetworkProfileByARN(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] DeviceFarm Network Profile (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading DeviceFarm Network Profile (%s): %s", d.Id(), err)
	}

	arn := aws.ToString(project.Arn)
	d.Set(names.AttrARN, arn)
	d.Set(names.AttrName, project.Name)
	d.Set(names.AttrDescription, project.Description)
	d.Set("downlink_bandwidth_bits", project.DownlinkBandwidthBits)
	d.Set("downlink_delay_ms", project.DownlinkDelayMs)
	d.Set("downlink_jitter_ms", project.DownlinkJitterMs)
	d.Set("downlink_loss_percent", project.DownlinkLossPercent)
	d.Set("uplink_bandwidth_bits", project.UplinkBandwidthBits)
	d.Set("uplink_delay_ms", project.UplinkDelayMs)
	d.Set("uplink_jitter_ms", project.UplinkJitterMs)
	d.Set("uplink_loss_percent", project.UplinkLossPercent)
	d.Set(names.AttrType, project.Type)

	projectArn, err := decodeProjectARN(arn, "networkprofile", meta)
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "decoding project_arn (%s): %s", arn, err)
	}

	d.Set("project_arn", projectArn)

	return diags
}

func resourceNetworkProfileUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).DeviceFarmClient(ctx)

	if d.HasChangesExcept(names.AttrTags, names.AttrTagsAll) {
		input := &devicefarm.UpdateNetworkProfileInput{
			Arn: aws.String(d.Id()),
		}

		if d.HasChange(names.AttrName) {
			input.Name = aws.String(d.Get(names.AttrName).(string))
		}

		if d.HasChange(names.AttrDescription) {
			input.Description = aws.String(d.Get(names.AttrDescription).(string))
		}

		if d.HasChange(names.AttrType) {
			input.Type = awstypes.NetworkProfileType(d.Get(names.AttrType).(string))
		}

		if d.HasChange("downlink_bandwidth_bits") {
			input.DownlinkBandwidthBits = aws.Int64(int64(d.Get("downlink_bandwidth_bits").(int)))
		}

		if d.HasChange("downlink_delay_ms") {
			input.DownlinkDelayMs = aws.Int64(int64(d.Get("downlink_delay_ms").(int)))
		}

		if d.HasChange("downlink_jitter_ms") {
			input.DownlinkJitterMs = aws.Int64(int64(d.Get("downlink_jitter_ms").(int)))
		}

		if d.HasChange("downlink_loss_percent") {
			input.DownlinkLossPercent = int32(d.Get("downlink_loss_percent").(int))
		}

		if d.HasChange("uplink_bandwidth_bits") {
			input.UplinkBandwidthBits = aws.Int64(int64(d.Get("uplink_bandwidth_bits").(int)))
		}

		if d.HasChange("uplink_delay_ms") {
			input.UplinkDelayMs = aws.Int64(int64(d.Get("uplink_delay_ms").(int)))
		}

		if d.HasChange("uplink_jitter_ms") {
			input.UplinkJitterMs = aws.Int64(int64(d.Get("uplink_jitter_ms").(int)))
		}

		if d.HasChange("uplink_loss_percent") {
			input.UplinkLossPercent = int32(d.Get("uplink_loss_percent").(int))
		}

		_, err := conn.UpdateNetworkProfile(ctx, input)

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "updating DeviceFarm Network Profile (%s): %s", d.Id(), err)
		}
	}

	return append(diags, resourceNetworkProfileRead(ctx, d, meta)...)
}

func resourceNetworkProfileDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).DeviceFarmClient(ctx)

	log.Printf("[DEBUG] Deleting DeviceFarm Network Profile: %s", d.Id())
	_, err := conn.DeleteNetworkProfile(ctx, &devicefarm.DeleteNetworkProfileInput{
		Arn: aws.String(d.Id()),
	})

	if errs.IsA[*awstypes.NotFoundException](err) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting DeviceFarm Network Profile (%s): %s", d.Id(), err)
	}

	return diags
}

func findNetworkProfileByARN(ctx context.Context, conn *devicefarm.Client, arn string) (*awstypes.NetworkProfile, error) {
	input := &devicefarm.GetNetworkProfileInput{
		Arn: aws.String(arn),
	}
	output, err := conn.GetNetworkProfile(ctx, input)

	if errs.IsA[*awstypes.NotFoundException](err) {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil || output.NetworkProfile == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	return output.NetworkProfile, nil
}
