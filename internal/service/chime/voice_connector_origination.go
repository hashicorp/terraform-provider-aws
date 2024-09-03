// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package chime

import (
	"context"
	"log"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/chimesdkvoice"
	awstypes "github.com/aws/aws-sdk-go-v2/service/chimesdkvoice/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_chime_voice_connector_origination")
func ResourceVoiceConnectorOrigination() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceVoiceConnectorOriginationCreate,
		ReadWithoutTimeout:   resourceVoiceConnectorOriginationRead,
		UpdateWithoutTimeout: resourceVoiceConnectorOriginationUpdate,
		DeleteWithoutTimeout: resourceVoiceConnectorOriginationDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			"disabled": {
				Type:     schema.TypeBool,
				Optional: true,
			},
			"route": {
				Type:     schema.TypeSet,
				Required: true,
				MinItems: 1,
				MaxItems: 20,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"host": {
							Type:         schema.TypeString,
							Required:     true,
							ValidateFunc: validation.IsIPAddress,
						},
						names.AttrPort: {
							Type:         schema.TypeInt,
							Optional:     true,
							Default:      5060,
							ValidateFunc: validation.IsPortNumber,
						},
						names.AttrPriority: {
							Type:         schema.TypeInt,
							Required:     true,
							ValidateFunc: validation.IntBetween(1, 99),
						},
						names.AttrProtocol: {
							Type:             schema.TypeString,
							Required:         true,
							ValidateDiagFunc: enum.Validate[awstypes.OriginationRouteProtocol](),
						},
						names.AttrWeight: {
							Type:         schema.TypeInt,
							Required:     true,
							ValidateFunc: validation.IntBetween(1, 99),
						},
					},
				},
			},
			"voice_connector_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
		},
	}
}

func resourceVoiceConnectorOriginationCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	conn := meta.(*conns.AWSClient).ChimeSDKVoiceClient(ctx)

	vcId := d.Get("voice_connector_id").(string)

	input := &chimesdkvoice.PutVoiceConnectorOriginationInput{
		VoiceConnectorId: aws.String(vcId),
		Origination: &awstypes.Origination{
			Routes: expandOriginationRoutes(d.Get("route").(*schema.Set).List()),
		},
	}

	if v, ok := d.GetOk("disabled"); ok {
		input.Origination.Disabled = aws.Bool(v.(bool))
	}

	if _, err := conn.PutVoiceConnectorOrigination(ctx, input); err != nil {
		return sdkdiag.AppendErrorf(diags, "creating Chime Voice Connector (%s) origination: %s", vcId, err)
	}

	d.SetId(vcId)

	return append(diags, resourceVoiceConnectorOriginationRead(ctx, d, meta)...)
}

func resourceVoiceConnectorOriginationRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	conn := meta.(*conns.AWSClient).ChimeSDKVoiceClient(ctx)

	resp, err := FindVoiceConnectorResourceWithRetry(ctx, d.IsNewResource(), func() (*awstypes.Origination, error) {
		return findVoiceConnectorOriginationByID(ctx, conn, d.Id())
	})

	if tfresource.TimedOut(err) {
		resp, err = findVoiceConnectorOriginationByID(ctx, conn, d.Id())
	}

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] Chime Voice Connector (%s) origination not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "getting Chime Voice Connector (%s) origination: %s", d.Id(), err)
	}

	d.Set("disabled", resp.Disabled)
	d.Set("voice_connector_id", d.Id())

	if err := d.Set("route", flattenOriginationRoutes(resp.Routes)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting Chime Voice Connector (%s) origination routes: %s", d.Id(), err)
	}

	return diags
}

func resourceVoiceConnectorOriginationUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	conn := meta.(*conns.AWSClient).ChimeSDKVoiceClient(ctx)

	if d.HasChanges("route", "disabled") {
		input := &chimesdkvoice.PutVoiceConnectorOriginationInput{
			VoiceConnectorId: aws.String(d.Id()),
			Origination: &awstypes.Origination{
				Routes: expandOriginationRoutes(d.Get("route").(*schema.Set).List()),
			},
		}

		if v, ok := d.GetOk("disabled"); ok {
			input.Origination.Disabled = aws.Bool(v.(bool))
		}

		_, err := conn.PutVoiceConnectorOrigination(ctx, input)

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "updating Chime Voice Connector (%s) origination: %s", d.Id(), err)
		}
	}

	return append(diags, resourceVoiceConnectorOriginationRead(ctx, d, meta)...)
}

func resourceVoiceConnectorOriginationDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	conn := meta.(*conns.AWSClient).ChimeSDKVoiceClient(ctx)

	input := &chimesdkvoice.DeleteVoiceConnectorOriginationInput{
		VoiceConnectorId: aws.String(d.Id()),
	}

	_, err := conn.DeleteVoiceConnectorOrigination(ctx, input)

	if errs.IsA[*awstypes.NotFoundException](err) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting Chime Voice Connector (%s) origination: %s", d.Id(), err)
	}

	return diags
}

func expandOriginationRoutes(data []interface{}) []awstypes.OriginationRoute {
	var originationRoutes []awstypes.OriginationRoute

	for _, rItem := range data {
		item := rItem.(map[string]interface{})
		originationRoutes = append(originationRoutes, awstypes.OriginationRoute{
			Host:     aws.String(item["host"].(string)),
			Port:     aws.Int32(int32(item[names.AttrPort].(int))),
			Priority: aws.Int32(int32(item[names.AttrPriority].(int))),
			Protocol: awstypes.OriginationRouteProtocol(item[names.AttrProtocol].(string)),
			Weight:   aws.Int32(int32(item[names.AttrWeight].(int))),
		})
	}

	return originationRoutes
}

func flattenOriginationRoutes(routes []awstypes.OriginationRoute) []interface{} {
	var rawRoutes []interface{}

	for _, route := range routes {
		r := map[string]interface{}{
			"host":             aws.ToString(route.Host),
			names.AttrPort:     aws.ToInt32(route.Port),
			names.AttrPriority: aws.ToInt32(route.Priority),
			names.AttrProtocol: string(route.Protocol),
			names.AttrWeight:   aws.ToInt32(route.Weight),
		}

		rawRoutes = append(rawRoutes, r)
	}

	return rawRoutes
}

func findVoiceConnectorOriginationByID(ctx context.Context, conn *chimesdkvoice.Client, id string) (*awstypes.Origination, error) {
	in := &chimesdkvoice.GetVoiceConnectorOriginationInput{
		VoiceConnectorId: aws.String(id),
	}

	resp, err := conn.GetVoiceConnectorOrigination(ctx, in)

	if errs.IsA[*awstypes.NotFoundException](err) {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: in,
		}
	}

	if resp == nil || resp.Origination == nil {
		return nil, tfresource.NewEmptyResultError(in)
	}

	if err != nil {
		return nil, err
	}

	return resp.Origination, nil
}
