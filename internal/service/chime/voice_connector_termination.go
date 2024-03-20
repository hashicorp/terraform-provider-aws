// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package chime

import (
	"context"
	"log"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/chimesdkvoice"
	awstypes "github.com/aws/aws-sdk-go-v2/service/chimesdkvoice/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

// @SDKResource("aws_chime_voice_connector_termination")
func ResourceVoiceConnectorTermination() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceVoiceConnectorTerminationCreate,
		ReadWithoutTimeout:   resourceVoiceConnectorTerminationRead,
		UpdateWithoutTimeout: resourceVoiceConnectorTerminationUpdate,
		DeleteWithoutTimeout: resourceVoiceConnectorTerminationDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			"calling_regions": {
				Type:     schema.TypeSet,
				Required: true,
				MinItems: 1,
				Elem: &schema.Schema{
					Type:         schema.TypeString,
					ValidateFunc: validation.StringLenBetween(2, 2),
				},
			},
			"cidr_allow_list": {
				Type:     schema.TypeSet,
				Required: true,
				MinItems: 1,
				Elem: &schema.Schema{
					Type:         schema.TypeString,
					ValidateFunc: validation.IsCIDRNetwork(27, 32),
				},
			},
			"cps_limit": {
				Type:         schema.TypeInt,
				Optional:     true,
				Default:      1,
				ValidateFunc: validation.IntAtLeast(1),
			},
			"default_phone_number": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringMatch(regexache.MustCompile(`^\+?[1-9]\d{1,14}$`), "must match ^\\+?[1-9]\\d{1,14}$"),
			},
			"disabled": {
				Type:     schema.TypeBool,
				Optional: true,
			},
			"voice_connector_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
		},
	}
}

func resourceVoiceConnectorTerminationCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	conn := meta.(*conns.AWSClient).ChimeSDKVoiceClient(ctx)

	vcId := d.Get("voice_connector_id").(string)

	input := &chimesdkvoice.PutVoiceConnectorTerminationInput{
		VoiceConnectorId: aws.String(vcId),
	}

	termination := &awstypes.Termination{
		CidrAllowedList: flex.ExpandStringValueSet(d.Get("cidr_allow_list").(*schema.Set)),
		CallingRegions:  flex.ExpandStringValueSet(d.Get("calling_regions").(*schema.Set)),
	}

	if v, ok := d.GetOk("disabled"); ok {
		termination.Disabled = aws.Bool(v.(bool))
	}

	if v, ok := d.GetOk("cps_limit"); ok {
		termination.CpsLimit = aws.Int32(int32(v.(int)))
	}

	if v, ok := d.GetOk("default_phone_number"); ok {
		termination.DefaultPhoneNumber = aws.String(v.(string))
	}

	input.Termination = termination

	if _, err := conn.PutVoiceConnectorTermination(ctx, input); err != nil {
		return sdkdiag.AppendErrorf(diags, "creating Chime Voice Connector (%s) termination: %s", vcId, err)
	}

	d.SetId(vcId)

	return append(diags, resourceVoiceConnectorTerminationRead(ctx, d, meta)...)
}

func resourceVoiceConnectorTerminationRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	conn := meta.(*conns.AWSClient).ChimeSDKVoiceClient(ctx)

	resp, err := FindVoiceConnectorResourceWithRetry(ctx, d.IsNewResource(), func() (*awstypes.Termination, error) {
		return findVoiceConnectorTerminationByID(ctx, conn, d.Id())
	})

	if tfresource.TimedOut(err) {
		resp, err = findVoiceConnectorTerminationByID(ctx, conn, d.Id())
	}

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] Chime Voice Connector (%s) termination not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "getting Chime Voice Connector (%s) termination: %s", d.Id(), err)
	}

	d.Set("cps_limit", resp.CpsLimit)
	d.Set("disabled", resp.Disabled)
	d.Set("default_phone_number", resp.DefaultPhoneNumber)

	if err := d.Set("calling_regions", flex.FlattenStringValueList(resp.CallingRegions)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting termination calling regions (%s): %s", d.Id(), err)
	}
	if err := d.Set("cidr_allow_list", flex.FlattenStringValueList(resp.CidrAllowedList)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting termination cidr allow list (%s): %s", d.Id(), err)
	}

	d.Set("voice_connector_id", d.Id())

	return diags
}

func resourceVoiceConnectorTerminationUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	conn := meta.(*conns.AWSClient).ChimeSDKVoiceClient(ctx)

	if d.HasChanges("calling_regions", "cidr_allow_list", "disabled", "cps_limit", "default_phone_number") {
		termination := &awstypes.Termination{
			CallingRegions:  flex.ExpandStringValueSet(d.Get("calling_regions").(*schema.Set)),
			CidrAllowedList: flex.ExpandStringValueSet(d.Get("cidr_allow_list").(*schema.Set)),
			CpsLimit:        aws.Int32(int32(d.Get("cps_limit").(int))),
		}

		if v, ok := d.GetOk("default_phone_number"); ok {
			termination.DefaultPhoneNumber = aws.String(v.(string))
		}

		if v, ok := d.GetOk("disabled"); ok {
			termination.Disabled = aws.Bool(v.(bool))
		}

		input := &chimesdkvoice.PutVoiceConnectorTerminationInput{
			VoiceConnectorId: aws.String(d.Id()),
			Termination:      termination,
		}

		_, err := conn.PutVoiceConnectorTermination(ctx, input)

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "updating Chime Voice Connector (%s) termination: %s", d.Id(), err)
		}
	}

	return append(diags, resourceVoiceConnectorTerminationRead(ctx, d, meta)...)
}

func resourceVoiceConnectorTerminationDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	conn := meta.(*conns.AWSClient).ChimeSDKVoiceClient(ctx)

	input := &chimesdkvoice.DeleteVoiceConnectorTerminationInput{
		VoiceConnectorId: aws.String(d.Id()),
	}

	_, err := conn.DeleteVoiceConnectorTermination(ctx, input)

	if errs.IsA[*awstypes.NotFoundException](err) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting Chime Voice Connector termination (%s): %s", d.Id(), err)
	}

	return diags
}

func findVoiceConnectorTerminationByID(ctx context.Context, conn *chimesdkvoice.Client, id string) (*awstypes.Termination, error) {
	in := &chimesdkvoice.GetVoiceConnectorInput{
		VoiceConnectorId: aws.String(id),
	}

	input := &chimesdkvoice.GetVoiceConnectorTerminationInput{
		VoiceConnectorId: aws.String(id),
	}

	resp, err := conn.GetVoiceConnectorTermination(ctx, input)
	if errs.IsA[*awstypes.NotFoundException](err) {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: in,
		}
	}

	if resp == nil || resp.Termination == nil {
		return nil, tfresource.NewEmptyResultError(in)
	}

	if err != nil {
		return nil, err
	}

	return resp.Termination, nil
}
